package storage

import (
	"context"
	"eino/internal/config"
	"eino/internal/model"
	"eino/internal/storage/memory"
	"eino/internal/storage/mysql"
	"eino/internal/storage/redis"
	"fmt"
)

// Storage 存储接口
type Storage interface {
	// Chatbot相关
	SaveChatbot(ctx context.Context, chatbot *model.Chatbot) error
	GetChatbot(ctx context.Context, id string) (*model.Chatbot, error)
	GetChatbots(ctx context.Context) ([]*model.Chatbot, error)
	DeleteChatbot(ctx context.Context, id string) error

	// Conversation相关
	SaveConversation(ctx context.Context, conv *model.Conversation) error
	GetConversationHistory(ctx context.Context, chatbotID string, limit int) ([]*model.Conversation, error)

	// 关闭连接
	Close() error
}

// NewStorage 创建存储实例
func NewStorage(cfg config.StorageConfig) (Storage, error) {
	switch cfg.Type {
	case "memory":
		return memory.NewMemoryStorage(), nil
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.MySQL.User,
			cfg.MySQL.Password,
			cfg.MySQL.Host,
			cfg.MySQL.Port,
			cfg.MySQL.Database,
		)
		return mysql.NewMySQLStorage(dsn)
	case "redis":
		addr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
		return redis.NewRedisStorage(addr, cfg.Redis.Password, cfg.Redis.DB)
	default:
		return memory.NewMemoryStorage(), nil
	}
}
