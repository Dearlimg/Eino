package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"eino/internal/storage/milvus"
	"github.com/cloudwego/eino/schema"
)

// RAGService RAG（检索增强生成）服务
type RAGService struct {
	ollamaURL      string
	embeddingModel string
	milvusStorage  *milvus.MilvusStorage
	collectionName string
	embeddingDim   int
}

// NewRAGService 创建RAG服务
func NewRAGService(milvusHost string, milvusPort int, ollamaURL string) (*RAGService, error) {

	// 初始化Milvus存储
	milvusStorage, err := milvus.NewMilvusStorage(milvusHost, milvusPort)
	if err != nil {
		return nil, fmt.Errorf("create milvus storage: %w", err)
	}

	service := &RAGService{
		ollamaURL:      ollamaURL,
		embeddingModel: "nomic-embed-text", // 使用支持嵌入的模型
		milvusStorage:  milvusStorage,
		collectionName: "knowledge_base",
		embeddingDim:   768, // nomic-embed-text的维度
	}

	// 创建集合（如果不存在）
	if err := service.milvusStorage.CreateCollection(context.Background(), service.collectionName, service.embeddingDim); err != nil {
		return nil, fmt.Errorf("create collection: %w", err)
	}

	return service, nil
}

// generateEmbedding 使用Ollama API生成嵌入向量
func (s *RAGService) generateEmbedding(ctx context.Context, text string) ([]float32, error) {
	url := fmt.Sprintf("%s/api/embeddings", s.ollamaURL)

	reqBody := map[string]interface{}{
		"model":  s.embeddingModel,
		"prompt": text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error: %s", string(body))
	}

	var result struct {
		Embedding []float64 `json:"embedding"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// 转换为float32
	embedding := make([]float32, len(result.Embedding))
	for i, v := range result.Embedding {
		embedding[i] = float32(v)
	}

	return embedding, nil
}

// AddKnowledge 添加知识到向量库
func (s *RAGService) AddKnowledge(ctx context.Context, content string) error {
	// 生成嵌入向量
	embedding, err := s.generateEmbedding(ctx, content)
	if err != nil {
		return fmt.Errorf("generate embedding: %w", err)
	}

	// 插入Milvus
	return s.milvusStorage.Insert(ctx, s.collectionName, content, embedding)
}

// SearchKnowledge 搜索相关知识
func (s *RAGService) SearchKnowledge(ctx context.Context, query string, topK int) ([]string, error) {
	// 生成查询向量
	embedding, err := s.generateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("generate query embedding: %w", err)
	}

	// 搜索相似向量
	contents, _, err := s.milvusStorage.Search(ctx, s.collectionName, embedding, topK)
	if err != nil {
		return nil, fmt.Errorf("search vectors: %w", err)
	}

	return contents, nil
}

// EnhanceMessages 增强消息列表（添加相关知识）
func (s *RAGService) EnhanceMessages(ctx context.Context, userMessage string, originalMessages []*schema.Message) ([]*schema.Message, error) {
	// 搜索相关知识
	knowledge, err := s.SearchKnowledge(ctx, userMessage, 3)
	if err != nil {
		// 如果搜索失败，返回原始消息
		return originalMessages, nil
	}

	if len(knowledge) == 0 {
		return originalMessages, nil
	}

	// 构建增强的消息列表
	enhancedMessages := make([]*schema.Message, 0)

	// 添加系统提示（包含相关知识）
	knowledgeText := strings.Join(knowledge, "\n\n")
	systemPrompt := fmt.Sprintf("以下是从知识库中检索到的相关信息，请基于这些信息回答问题：\n\n%s", knowledgeText)

	// 查找并更新系统消息，或添加新的系统消息
	foundSystem := false
	for _, msg := range originalMessages {
		if msg.Role == "system" {
			// 更新现有系统消息
			enhancedMessages = append(enhancedMessages, schema.SystemMessage(msg.Content+"\n\n"+knowledgeText))
			foundSystem = true
		} else {
			enhancedMessages = append(enhancedMessages, msg)
		}
	}

	if !foundSystem {
		// 如果没有系统消息，在开头添加
		enhancedMessages = append([]*schema.Message{schema.SystemMessage(systemPrompt)}, enhancedMessages...)
	}

	return enhancedMessages, nil
}

// Close 关闭服务
func (s *RAGService) Close() error {
	return s.milvusStorage.Close()
}
