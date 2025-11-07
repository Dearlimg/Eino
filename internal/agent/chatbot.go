package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"eino/internal/config"
	"eino/internal/model"
	"eino/internal/storage"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

// ChatService 聊天服务
type ChatService struct {
	model   *ollama.ChatModel
	config  *config.Config
	storage storage.Storage
}

// NewChatService 创建聊天服务
func NewChatService(cfg *config.Config, storage storage.Storage) (*ChatService, error) {
	ctx := context.Background()

	var chatModel *ollama.ChatModel
	var err error

	switch cfg.Model.Provider {
	case "ollama":
		chatModel, err = ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
			BaseURL: cfg.Model.BaseURL,
			Model:   cfg.Model.Model,
		})
		if err != nil {
			return nil, fmt.Errorf("create ollama model: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported model provider: %s", cfg.Model.Provider)
	}

	return &ChatService{
		model:   chatModel,
		config:  cfg,
		storage: storage,
	}, nil
}

// CreateChatbot 创建聊天机器人实例
func (s *ChatService) CreateChatbot(ctx context.Context, req *model.CreateChatbotRequest) (*model.Chatbot, error) {
	chatbot := &model.Chatbot{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Personality: req.Personality,
		Background:  req.Background,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 构建系统提示词
	systemPrompt := s.buildSystemPrompt(req.Personality, req.Background)
	chatbot.SystemPrompt = systemPrompt

	// 保存到存储
	if err := s.storage.SaveChatbot(ctx, chatbot); err != nil {
		return nil, fmt.Errorf("save chatbot: %w", err)
	}

	return chatbot, nil
}

// Chat 进行对话
func (s *ChatService) Chat(ctx context.Context, chatbotID string, userMessage string) (*model.ChatResponse, error) {
	// 获取聊天机器人配置
	chatbot, err := s.storage.GetChatbot(ctx, chatbotID)
	if err != nil {
		return nil, fmt.Errorf("get chatbot: %w", err)
	}

	// 获取对话历史
	history, err := s.storage.GetConversationHistory(ctx, chatbotID, s.config.Agent.MaxHistory)
	if err != nil {
		return nil, fmt.Errorf("get conversation history: %w", err)
	}

	// 构建消息列表
	messages := s.buildMessages(chatbot.SystemPrompt, history, userMessage)

	// 设置超时上下文
	modelCtx, cancel := context.WithTimeout(ctx, s.config.GetModelTimeout())
	defer cancel()

	// 生成回复
	startTime := time.Now()
	response, err := s.model.Generate(modelCtx, messages)
	if err != nil {
		return nil, fmt.Errorf("generate response: %w", err)
	}
	duration := time.Since(startTime)

	// 保存对话记录
	conversation := &model.Conversation{
		ChatbotID:   chatbotID,
		UserMessage: userMessage,
		BotMessage:  response.Content,
		CreatedAt:   time.Now(),
	}

	if err := s.storage.SaveConversation(ctx, conversation); err != nil {
		return nil, fmt.Errorf("save conversation: %w", err)
	}

	return &model.ChatResponse{
		Message:   response.Content,
		Duration:  duration.Milliseconds(),
		Timestamp: time.Now(),
	}, nil
}

// StreamChat 流式对话
func (s *ChatService) StreamChat(ctx context.Context, chatbotID string, userMessage string, callback func(string)) error {
	// 获取聊天机器人配置
	chatbot, err := s.storage.GetChatbot(ctx, chatbotID)
	if err != nil {
		return fmt.Errorf("get chatbot: %w", err)
	}

	// 获取对话历史
	history, err := s.storage.GetConversationHistory(ctx, chatbotID, s.config.Agent.MaxHistory)
	if err != nil {
		return fmt.Errorf("get conversation history: %w", err)
	}

	// 构建消息列表
	messages := s.buildMessages(chatbot.SystemPrompt, history, userMessage)

	// 设置超时上下文
	modelCtx, cancel := context.WithTimeout(ctx, s.config.GetModelTimeout())
	defer cancel()

	// 流式生成
	stream, err := s.model.Stream(modelCtx, messages)
	if err != nil {
		return fmt.Errorf("stream generate: %w", err)
	}

	var fullResponse strings.Builder
	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("stream recv: %w", err)
		}

		content := chunk.Content
		fullResponse.WriteString(content)
		if callback != nil {
			callback(content)
		}
	}

	// 保存完整对话记录
	conversation := &model.Conversation{
		ChatbotID:   chatbotID,
		UserMessage: userMessage,
		BotMessage:  fullResponse.String(),
		CreatedAt:   time.Now(),
	}

	if err := s.storage.SaveConversation(ctx, conversation); err != nil {
		return fmt.Errorf("save conversation: %w", err)
	}

	return nil
}

// GetChatbots 获取所有聊天机器人
func (s *ChatService) GetChatbots(ctx context.Context) ([]*model.Chatbot, error) {
	return s.storage.GetChatbots(ctx)
}

// GetChatbot 获取指定聊天机器人
func (s *ChatService) GetChatbot(ctx context.Context, chatbotID string) (*model.Chatbot, error) {
	return s.storage.GetChatbot(ctx, chatbotID)
}

// DeleteChatbot 删除聊天机器人
func (s *ChatService) DeleteChatbot(ctx context.Context, chatbotID string) error {
	return s.storage.DeleteChatbot(ctx, chatbotID)
}

// GetConversationHistory 获取对话历史
func (s *ChatService) GetConversationHistory(ctx context.Context, chatbotID string, limit int) ([]*model.Conversation, error) {
	return s.storage.GetConversationHistory(ctx, chatbotID, limit)
}

// buildSystemPrompt 构建系统提示词
func (s *ChatService) buildSystemPrompt(personality, background string) string {
	var parts []string

	if personality != "" {
		parts = append(parts, fmt.Sprintf("性格设定：%s", personality))
	}

	if background != "" {
		parts = append(parts, fmt.Sprintf("背景设定：%s", background))
	}

	if len(parts) > 0 {
		return strings.Join(parts, "\n\n") + "\n\n请严格按照以上设定进行对话，保持角色的一致性。"
	}

	return "你是一个友好的AI助手。"
}

// buildMessages 构建消息列表
func (s *ChatService) buildMessages(systemPrompt string, history []*model.Conversation, userMessage string) []*schema.Message {
	messages := make([]*schema.Message, 0)

	// 添加系统消息
	if systemPrompt != "" {
		messages = append(messages, schema.SystemMessage(systemPrompt))
	}

	// 添加历史对话
	for _, conv := range history {
		messages = append(messages, schema.UserMessage(conv.UserMessage))
		messages = append(messages, schema.AssistantMessage(conv.BotMessage, nil))
	}

	// 添加当前用户消息
	messages = append(messages, schema.UserMessage(userMessage))

	return messages
}
