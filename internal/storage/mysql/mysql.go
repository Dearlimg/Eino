package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"eino/internal/model"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLStorage MySQL存储实现
type MySQLStorage struct {
	db *sql.DB
}

// NewMySQLStorage 创建MySQL存储实例
func NewMySQLStorage(dsn string) (*MySQLStorage, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open mysql connection: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	return &MySQLStorage{db: db}, nil
}

// SaveChatbot 保存聊天机器人
func (s *MySQLStorage) SaveChatbot(ctx context.Context, chatbot *model.Chatbot) error {
	query := `
		INSERT INTO chatbots (id, name, personality, background, system_prompt, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			name = VALUES(name),
			personality = VALUES(personality),
			background = VALUES(background),
			system_prompt = VALUES(system_prompt),
			updated_at = VALUES(updated_at)
	`

	_, err := s.db.ExecContext(ctx, query,
		chatbot.ID,
		chatbot.Name,
		chatbot.Personality,
		chatbot.Background,
		chatbot.SystemPrompt,
		chatbot.CreatedAt,
		chatbot.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("save chatbot: %w", err)
	}

	return nil
}

// GetChatbot 获取聊天机器人
func (s *MySQLStorage) GetChatbot(ctx context.Context, id string) (*model.Chatbot, error) {
	query := `
		SELECT id, name, personality, background, system_prompt, created_at, updated_at
		FROM chatbots
		WHERE id = ?
	`

	var chatbot model.Chatbot
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&chatbot.ID,
		&chatbot.Name,
		&chatbot.Personality,
		&chatbot.Background,
		&chatbot.SystemPrompt,
		&chatbot.CreatedAt,
		&chatbot.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrChatbotNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get chatbot: %w", err)
	}

	return &chatbot, nil
}

// GetChatbots 获取所有聊天机器人
func (s *MySQLStorage) GetChatbots(ctx context.Context) ([]*model.Chatbot, error) {
	query := `
		SELECT id, name, personality, background, system_prompt, created_at, updated_at
		FROM chatbots
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get chatbots: %w", err)
	}
	defer rows.Close()

	var chatbots []*model.Chatbot
	for rows.Next() {
		var chatbot model.Chatbot
		if err := rows.Scan(
			&chatbot.ID,
			&chatbot.Name,
			&chatbot.Personality,
			&chatbot.Background,
			&chatbot.SystemPrompt,
			&chatbot.CreatedAt,
			&chatbot.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan chatbot: %w", err)
		}
		chatbots = append(chatbots, &chatbot)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return chatbots, nil
}

// DeleteChatbot 删除聊天机器人
func (s *MySQLStorage) DeleteChatbot(ctx context.Context, id string) error {
	query := `DELETE FROM chatbots WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete chatbot: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrChatbotNotFound
	}

	return nil
}

// SaveConversation 保存对话记录
func (s *MySQLStorage) SaveConversation(ctx context.Context, conv *model.Conversation) error {
	query := `
		INSERT INTO conversations (chatbot_id, user_message, bot_message, created_at)
		VALUES (?, ?, ?, ?)
	`

	result, err := s.db.ExecContext(ctx, query,
		conv.ChatbotID,
		conv.UserMessage,
		conv.BotMessage,
		conv.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("save conversation: %w", err)
	}

	// 获取插入的ID
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}

	conv.ID = id
	return nil
}

// GetConversationHistory 获取对话历史
func (s *MySQLStorage) GetConversationHistory(ctx context.Context, chatbotID string, limit int) ([]*model.Conversation, error) {
	query := `
		SELECT id, chatbot_id, user_message, bot_message, created_at
		FROM conversations
		WHERE chatbot_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, query, chatbotID, limit)
	if err != nil {
		return nil, fmt.Errorf("get conversation history: %w", err)
	}
	defer rows.Close()

	var conversations []*model.Conversation
	for rows.Next() {
		var conv model.Conversation
		if err := rows.Scan(
			&conv.ID,
			&conv.ChatbotID,
			&conv.UserMessage,
			&conv.BotMessage,
			&conv.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan conversation: %w", err)
		}
		conversations = append(conversations, &conv)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	// 反转顺序，使最早的对话在前
	for i, j := 0, len(conversations)-1; i < j; i, j = i+1, j-1 {
		conversations[i], conversations[j] = conversations[j], conversations[i]
	}

	return conversations, nil
}

// Close 关闭数据库连接
func (s *MySQLStorage) Close() error {
	return s.db.Close()
}
