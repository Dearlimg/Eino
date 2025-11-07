package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"eino/internal/model"

	"github.com/redis/go-redis/v9"
)

// RedisStorage Redis存储实现
type RedisStorage struct {
	client *redis.Client
}

// NewRedisStorage 创建Redis存储实例
func NewRedisStorage(addr, password string, db int) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
		PoolSize: 10,
	})

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &RedisStorage{client: client}, nil
}

// SaveChatbot 保存聊天机器人（缓存）
func (s *RedisStorage) SaveChatbot(ctx context.Context, chatbot *model.Chatbot) error {
	key := fmt.Sprintf("chatbot:%s", chatbot.ID)

	data, err := json.Marshal(chatbot)
	if err != nil {
		return fmt.Errorf("marshal chatbot: %w", err)
	}

	// 缓存24小时
	return s.client.Set(ctx, key, data, 24*time.Hour).Err()
}

// GetChatbot 获取聊天机器人（从缓存）
func (s *RedisStorage) GetChatbot(ctx context.Context, id string) (*model.Chatbot, error) {
	key := fmt.Sprintf("chatbot:%s", id)

	data, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, ErrChatbotNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get chatbot: %w", err)
	}

	var chatbot model.Chatbot
	if err := json.Unmarshal(data, &chatbot); err != nil {
		return nil, fmt.Errorf("unmarshal chatbot: %w", err)
	}

	return &chatbot, nil
}

// GetChatbots 获取所有聊天机器人（Redis不适合，需要从主存储获取）
func (s *RedisStorage) GetChatbots(ctx context.Context) ([]*model.Chatbot, error) {
	// Redis不适合存储列表，应该从MySQL获取
	// 这里返回空，实际应该从主存储（MySQL）获取
	return []*model.Chatbot{}, nil
}

// DeleteChatbot 删除聊天机器人（从缓存）
func (s *RedisStorage) DeleteChatbot(ctx context.Context, id string) error {
	key := fmt.Sprintf("chatbot:%s", id)
	return s.client.Del(ctx, key).Err()
}

// SaveConversation 保存对话记录（添加到有序集合）
func (s *RedisStorage) SaveConversation(ctx context.Context, conv *model.Conversation) error {
	key := fmt.Sprintf("conversations:%s", conv.ChatbotID)

	data, err := json.Marshal(conv)
	if err != nil {
		return fmt.Errorf("marshal conversation: %w", err)
	}

	// 使用时间戳作为分数，对话ID作为成员
	score := float64(conv.CreatedAt.Unix())
	member := fmt.Sprintf("%d", conv.ID)

	// 添加到有序集合
	if err := s.client.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: member,
	}).Err(); err != nil {
		return fmt.Errorf("zadd conversation: %w", err)
	}

	// 保存对话内容
	contentKey := fmt.Sprintf("conversation:%s:%d", conv.ChatbotID, conv.ID)
	if err := s.client.Set(ctx, contentKey, data, 7*24*time.Hour).Err(); err != nil {
		return fmt.Errorf("set conversation content: %w", err)
	}

	// 限制有序集合大小（保留最近1000条）
	s.client.ZRemRangeByRank(ctx, key, 0, -1001)

	return nil
}

// GetConversationHistory 获取对话历史
func (s *RedisStorage) GetConversationHistory(ctx context.Context, chatbotID string, limit int) ([]*model.Conversation, error) {
	key := fmt.Sprintf("conversations:%s", chatbotID)

	// 从有序集合获取最近的对话ID
	members, err := s.client.ZRevRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("zrevrange: %w", err)
	}

	var conversations []*model.Conversation
	for _, member := range members {
		contentKey := fmt.Sprintf("conversation:%s:%s", chatbotID, member)
		data, err := s.client.Get(ctx, contentKey).Bytes()
		if err != nil {
			continue // 跳过已过期的对话
		}

		var conv model.Conversation
		if err := json.Unmarshal(data, &conv); err != nil {
			continue
		}

		conversations = append(conversations, &conv)
	}

	// 反转顺序，使最早的对话在前
	for i, j := 0, len(conversations)-1; i < j; i, j = i+1, j-1 {
		conversations[i], conversations[j] = conversations[j], conversations[i]
	}

	return conversations, nil
}

// SetSession 设置会话状态
func (s *RedisStorage) SetSession(ctx context.Context, chatbotID string, data map[string]interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("session:%s", chatbotID)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	return s.client.Set(ctx, key, jsonData, ttl).Err()
}

// GetSession 获取会话状态
func (s *RedisStorage) GetSession(ctx context.Context, chatbotID string) (map[string]interface{}, error) {
	key := fmt.Sprintf("session:%s", chatbotID)

	data, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return make(map[string]interface{}), nil
	}
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}

	var session map[string]interface{}
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}

	return session, nil
}

// RateLimit 限流检查
func (s *RedisStorage) RateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	count, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("incr rate limit: %w", err)
	}

	if count == 1 {
		// 设置过期时间
		s.client.Expire(ctx, key, window)
	}

	return count <= int64(limit), nil
}

// Close 关闭Redis连接
func (s *RedisStorage) Close() error {
	return s.client.Close()
}
