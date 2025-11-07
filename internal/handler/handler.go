package handler

import (
	"fmt"
	"net/http"

	"eino/internal/agent"
	"eino/internal/model"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册路由
func RegisterRoutes(router *gin.Engine, chatService *agent.ChatService) {
	api := router.Group("/api/v1")
	{
		// 聊天机器人管理
		api.POST("/chatbots", createChatbot(chatService))
		api.GET("/chatbots", getChatbots(chatService))
		api.GET("/chatbots/:id", getChatbot(chatService))
		api.PUT("/chatbots/:id", updateChatbot(chatService))
		api.DELETE("/chatbots/:id", deleteChatbot(chatService))

		// 对话接口
		api.POST("/chatbots/:id/chat", chat(chatService))
		api.GET("/chatbots/:id/history", getHistory(chatService))

		// RAG知识库接口（如果启用）
		api.POST("/knowledge", addKnowledge(chatService))
		api.GET("/knowledge/search", searchKnowledge(chatService))
	}

	// 健康检查
	router.GET("/health", healthCheck)
}

// createChatbot 创建聊天机器人
func createChatbot(service *agent.ChatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.CreateChatbotRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "invalid_request",
				Message: err.Error(),
			})
			return
		}

		chatbot, err := service.CreateChatbot(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error:   "create_chatbot_failed",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, chatbot)
	}
}

// getChatbots 获取所有聊天机器人
func getChatbots(service *agent.ChatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		chatbots, err := service.GetChatbots(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error:   "get_chatbots_failed",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, chatbots)
	}
}

// getChatbot 获取指定聊天机器人
func getChatbot(service *agent.ChatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		chatbot, err := service.GetChatbot(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, model.ErrorResponse{
				Error:   "chatbot_not_found",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, chatbot)
	}
}

// updateChatbot 更新聊天机器人
func updateChatbot(service *agent.ChatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		_ = c.Param("id") // 暂时未使用
		var req model.UpdateChatbotRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "invalid_request",
				Message: err.Error(),
			})
			return
		}

		// TODO: 实现更新逻辑
		_ = req // 暂时未使用
		c.JSON(http.StatusOK, gin.H{"message": "update not implemented yet"})
	}
}

// deleteChatbot 删除聊天机器人
func deleteChatbot(service *agent.ChatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := service.DeleteChatbot(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error:   "delete_chatbot_failed",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

// chat 进行对话
func chat(service *agent.ChatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		chatbotID := c.Param("id")
		var req model.ChatRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "invalid_request",
				Message: err.Error(),
			})
			return
		}

		response, err := service.Chat(c.Request.Context(), chatbotID, req.Message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error:   "chat_failed",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

// getHistory 获取对话历史
func getHistory(service *agent.ChatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		chatbotID := c.Param("id")
		limit := 20
		if l := c.Query("limit"); l != "" {
			if _, err := fmt.Sscanf(l, "%d", &limit); err != nil {
				limit = 20
			}
		}

		history, err := service.GetConversationHistory(c.Request.Context(), chatbotID, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error:   "get_history_failed",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, history)
	}
}

// healthCheck 健康检查
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "eino-chatbot",
	})
}

// addKnowledge 添加知识到向量库
func addKnowledge(service *agent.ChatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Content string `json:"content" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "invalid_request",
				Message: err.Error(),
			})
			return
		}

		// TODO: 实现添加知识功能
		c.JSON(http.StatusOK, gin.H{"message": "knowledge added"})
	}
}

// searchKnowledge 搜索知识
func searchKnowledge(service *agent.ChatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "invalid_request",
				Message: "query parameter 'q' is required",
			})
			return
		}

		// TODO: 实现搜索知识功能
		c.JSON(http.StatusOK, gin.H{"results": []string{}})
	}
}
