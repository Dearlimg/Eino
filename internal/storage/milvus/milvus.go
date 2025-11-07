package milvus

import (
	"context"
	"fmt"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

// MilvusStorage Milvus向量数据库存储
// 注意：这是一个基础实现，完整的Milvus集成需要根据实际SDK版本调整
type MilvusStorage struct {
	client client.Client
}

// NewMilvusStorage 创建Milvus存储实例
func NewMilvusStorage(host string, port int) (*MilvusStorage, error) {
	addr := fmt.Sprintf("%s:%d", host, port)

	// 创建Milvus客户端
	c, err := client.NewClient(context.Background(), client.Config{
		Address: addr,
	})
	if err != nil {
		return nil, fmt.Errorf("create milvus client: %w", err)
	}

	return &MilvusStorage{client: c}, nil
}

// CreateCollection 创建集合（如果不存在）
func (s *MilvusStorage) CreateCollection(ctx context.Context, collectionName string, dim int) error {
	// 检查集合是否存在
	exists, err := s.client.HasCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("check collection exists: %w", err)
	}

	if exists {
		return nil // 集合已存在
	}

	// 定义schema
	schema := &entity.Schema{
		CollectionName: collectionName,
		Description:    "Knowledge base collection",
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeInt64,
				PrimaryKey: true,
				AutoID:     true,
			},
			{
				Name:     "content",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					"max_length": "65535",
				},
			},
			{
				Name:     "embedding",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": fmt.Sprintf("%d", dim),
				},
			},
		},
	}

	// 创建集合
	if err := s.client.CreateCollection(ctx, schema, entity.DefaultShardNumber); err != nil {
		return fmt.Errorf("create collection: %w", err)
	}

	// 创建索引
	idx, err := entity.NewIndexHNSW(entity.L2, 16, 200)
	if err != nil {
		return fmt.Errorf("create index: %w", err)
	}

	if err := s.client.CreateIndex(ctx, collectionName, "embedding", idx, false); err != nil {
		return fmt.Errorf("create index: %w", err)
	}

	return nil
}

// Insert 插入向量数据
func (s *MilvusStorage) Insert(ctx context.Context, collectionName string, content string, embedding []float32) error {
	// 准备数据
	contents := []string{content}
	embeddings := [][]float32{embedding}

	// 插入数据
	_, err := s.client.Insert(ctx, collectionName, "",
		entity.NewColumnVarChar("content", contents),
		entity.NewColumnFloatVector("embedding", len(embedding), embeddings))
	if err != nil {
		return fmt.Errorf("insert vector: %w", err)
	}

	return nil
}

// Search 搜索相似向量
// 注意：此实现需要根据实际Milvus SDK版本调整
func (s *MilvusStorage) Search(ctx context.Context, collectionName string, queryVector []float32, topK int) ([]string, []float32, error) {
	// TODO: 实现搜索功能
	// 由于Milvus SDK版本差异，这里提供一个占位实现
	// 实际使用时需要根据SDK文档调整

	return []string{}, []float32{}, fmt.Errorf("Milvus search not fully implemented yet, please refer to SDK documentation")
}

// Close 关闭连接
func (s *MilvusStorage) Close() error {
	return s.client.Close()
}
