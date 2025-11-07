package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Model   ModelConfig   `yaml:"model"`
	Agent   AgentConfig   `yaml:"agent"`
	Storage StorageConfig `yaml:"storage"`
	RAG     RAGConfig     `yaml:"rag"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Address string `yaml:"address"`
	Mode    string `yaml:"mode"` // debug, release
}

// ModelConfig 模型配置
type ModelConfig struct {
	Provider string `yaml:"provider"` // ollama, openai, anthropic
	BaseURL  string `yaml:"base_url"`
	APIKey   string `yaml:"api_key"`
	Model    string `yaml:"model"`
	Timeout  int    `yaml:"timeout"` // 秒
}

// AgentConfig Agent配置
type AgentConfig struct {
	MaxRetries   int     `yaml:"max_retries"`
	Timeout      int     `yaml:"timeout"` // 秒
	MaxTokens    int     `yaml:"max_tokens"`
	Temperature  float64 `yaml:"temperature"`
	MaxHistory   int     `yaml:"max_history"` // 最大对话历史条数
	EnableStream bool    `yaml:"enable_stream"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type   string `yaml:"type"` // memory, mysql, redis
	MySQL  MySQLConfig
	Redis  RedisConfig
	Milvus MilvusConfig
}

// MySQLConfig MySQL配置
type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// MilvusConfig Milvus配置
type MilvusConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// RAGConfig RAG配置
type RAGConfig struct {
	Enabled        bool   `yaml:"enabled"`
	OllamaURL      string `yaml:"ollama_url"`
	EmbeddingModel string `yaml:"embedding_model"`
}

// Load 加载配置文件
func Load(path string) (*Config, error) {
	// 如果路径是相对路径，尝试从多个位置查找
	if !filepath.IsAbs(path) {
		// 获取当前工作目录
		wd, err := os.Getwd()
		if err == nil {
			// 从当前工作目录向上查找配置文件
			currentDir := wd
			maxDepth := 5 // 最多向上查找5层

			for i := 0; i < maxDepth; i++ {
				testPath := filepath.Join(currentDir, path)
				if _, err := os.Stat(testPath); err == nil {
					path = testPath
					break
				}
				// 向上查找
				parent := filepath.Dir(currentDir)
				if parent == currentDir {
					break // 已到达根目录
				}
				currentDir = parent
			}
		}

		// 如果还没找到，尝试可执行文件目录
		execPath, err := os.Executable()
		if err == nil {
			execDir := filepath.Dir(execPath)
			testPath := filepath.Join(execDir, path)
			if _, err := os.Stat(testPath); err == nil {
				path = testPath
			} else {
				// 尝试可执行文件目录的上级
				testPath = filepath.Join(execDir, "..", path)
				if _, err := os.Stat(testPath); err == nil {
					path = testPath
				}
			}
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	// 设置默认值
	if cfg.Server.Address == "" {
		cfg.Server.Address = ":8080"
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "release"
	}
	if cfg.Agent.MaxRetries == 0 {
		cfg.Agent.MaxRetries = 3
	}
	if cfg.Agent.Timeout == 0 {
		cfg.Agent.Timeout = 30
	}
	if cfg.Agent.MaxHistory == 0 {
		cfg.Agent.MaxHistory = 20
	}
	if cfg.Model.Timeout == 0 {
		cfg.Model.Timeout = 60
	}

	return &cfg, nil
}

// GetModelTimeout 获取模型超时时间
func (c *Config) GetModelTimeout() time.Duration {
	return time.Duration(c.Model.Timeout) * time.Second
}

// GetAgentTimeout 获取Agent超时时间
func (c *Config) GetAgentTimeout() time.Duration {
	return time.Duration(c.Agent.Timeout) * time.Second
}
