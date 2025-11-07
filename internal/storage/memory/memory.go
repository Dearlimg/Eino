package memory

import (
	"context"
	"eino/internal/model"
	"sync"
	"time"
)

// MemoryStorage 内存存储实现
type MemoryStorage struct {
	chatbots      map[string]*model.Chatbot
	conversations map[string][]*model.Conversation
	mu            sync.RWMutex
	convID        int64
}

// NewMemoryStorage 创建内存存储实例
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		chatbots:      make(map[string]*model.Chatbot),
		conversations: make(map[string][]*model.Conversation),
		convID:        1,
	}
}

// SaveChatbot 保存聊天机器人
func (s *MemoryStorage) SaveChatbot(ctx context.Context, chatbot *model.Chatbot) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	chatbot.UpdatedAt = time.Now()
	s.chatbots[chatbot.ID] = chatbot
	return nil
}

// GetChatbot 获取聊天机器人
func (s *MemoryStorage) GetChatbot(ctx context.Context, id string) (*model.Chatbot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	chatbot, ok := s.chatbots[id]
	if !ok {
		return nil, ErrChatbotNotFound
	}

	// 返回副本，避免并发修改
	result := *chatbot
	return &result, nil
}

// GetChatbots 获取所有聊天机器人
func (s *MemoryStorage) GetChatbots(ctx context.Context) ([]*model.Chatbot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	chatbots := make([]*model.Chatbot, 0, len(s.chatbots))
	for _, chatbot := range s.chatbots {
		// 返回副本
		cb := *chatbot
		chatbots = append(chatbots, &cb)
	}

	return chatbots, nil
}

// DeleteChatbot 删除聊天机器人
func (s *MemoryStorage) DeleteChatbot(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.chatbots[id]; !ok {
		return ErrChatbotNotFound
	}

	delete(s.chatbots, id)
	delete(s.conversations, id)
	return nil
}

// SaveConversation 保存对话记录
func (s *MemoryStorage) SaveConversation(ctx context.Context, conv *model.Conversation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	conv.ID = s.convID
	s.convID++
	conv.CreatedAt = time.Now()

	if s.conversations[conv.ChatbotID] == nil {
		s.conversations[conv.ChatbotID] = make([]*model.Conversation, 0)
	}

	// 返回副本
	convCopy := *conv
	s.conversations[conv.ChatbotID] = append(s.conversations[conv.ChatbotID], &convCopy)
	return nil
}

// GetConversationHistory 获取对话历史
func (s *MemoryStorage) GetConversationHistory(ctx context.Context, chatbotID string, limit int) ([]*model.Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	convs, ok := s.conversations[chatbotID]
	if !ok {
		return []*model.Conversation{}, nil
	}

	// 获取最近的limit条记录
	start := 0
	if len(convs) > limit {
		start = len(convs) - limit
	}

	result := make([]*model.Conversation, 0, limit)
	for i := start; i < len(convs); i++ {
		// 返回副本
		conv := *convs[i]
		result = append(result, &conv)
	}

	return result, nil
}

// Close 关闭存储
func (s *MemoryStorage) Close() error {
	return nil
}
