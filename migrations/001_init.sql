-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS eino_chatbot CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE eino_chatbot;

-- 聊天机器人表
CREATE TABLE IF NOT EXISTS chatbots (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    personality TEXT,
    background TEXT,
    system_prompt TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 对话记录表
CREATE TABLE IF NOT EXISTS conversations (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    chatbot_id VARCHAR(36) NOT NULL,
    user_message TEXT NOT NULL,
    bot_message TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_chatbot_id (chatbot_id),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (chatbot_id) REFERENCES chatbots(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 知识库表（可选，用于管理知识库元数据）
CREATE TABLE IF NOT EXISTS knowledge_base (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    content TEXT NOT NULL,
    source VARCHAR(255),
    category VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_category (category),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

