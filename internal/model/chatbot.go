package model

import "time"

// Chatbot 聊天机器人模型
type Chatbot struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Personality  string    `json:"personality"`   // 性格设定
	Background   string    `json:"background"`    // 背景设定
	SystemPrompt string    `json:"system_prompt"` // 系统提示词（自动生成）
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateChatbotRequest 创建聊天机器人请求
type CreateChatbotRequest struct {
	Name        string `json:"name" binding:"required"`
	Personality string `json:"personality" binding:"required"`
	Background  string `json:"background" binding:"required"`
}

// UpdateChatbotRequest 更新聊天机器人请求
type UpdateChatbotRequest struct {
	Name        string `json:"name"`
	Personality string `json:"personality"`
	Background  string `json:"background"`
}

// Conversation 对话记录
type Conversation struct {
	ID          int64     `json:"id"`
	ChatbotID   string    `json:"chatbot_id"`
	UserMessage string    `json:"user_message"`
	BotMessage  string    `json:"bot_message"`
	CreatedAt   time.Time `json:"created_at"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Message string `json:"message" binding:"required"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Message   string    `json:"message"`
	Duration  int64     `json:"duration"` // 毫秒
	Timestamp time.Time `json:"timestamp"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
