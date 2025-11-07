# 存储扩展实现计划

## 📋 实现优先级

### Phase 1: MySQL 存储（高优先级）
**预计时间**：2-3天
**价值**：数据持久化，生产环境必需

### Phase 2: Redis 存储（中优先级）
**预计时间**：1-2天
**价值**：性能优化，高并发支持

### Phase 3: Milvus 集成（中优先级）
**预计时间**：3-5天
**价值**：RAG能力，智能检索

---

## 🗄️ Phase 1: MySQL 存储实现

### 1.1 数据库设计

```sql
-- 聊天机器人表
CREATE TABLE chatbots (
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
CREATE TABLE conversations (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    chatbot_id VARCHAR(36) NOT NULL,
    user_message TEXT NOT NULL,
    bot_message TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_chatbot_id (chatbot_id),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (chatbot_id) REFERENCES chatbots(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 1.2 实现文件结构

```
internal/storage/mysql/
├── mysql.go          # MySQL存储实现
├── models.go         # 数据库模型
└── migrations/        # 数据库迁移脚本
    └── 001_init.sql
```

### 1.3 核心功能

- [x] 连接池管理
- [x] CRUD操作
- [x] 事务支持
- [x] 连接重试
- [x] 查询优化

---

## 🔴 Phase 2: Redis 存储实现

### 2.1 数据结构设计

```
# 聊天机器人缓存
chatbot:{id} -> JSON

# 对话历史（有序集合，按时间排序）
conversations:{chatbot_id} -> ZSET (score=timestamp, member=conversation_id)

# 会话状态
session:{chatbot_id} -> HASH (context, last_message_time)

# 限流计数
rate_limit:{user_id} -> STRING (计数)
```

### 2.2 实现文件结构

```
internal/storage/redis/
├── redis.go          # Redis存储实现
├── cache.go          # 缓存操作
└── session.go        # 会话管理
```

### 2.3 核心功能

- [x] 连接池管理
- [x] 缓存操作
- [x] 会话管理
- [x] 限流功能
- [x] 过期策略

---

## 🔍 Phase 3: Milvus 集成

### 3.1 向量数据库设计

```go
// 知识库集合结构
Collection: "knowledge_base"
Fields:
  - id: INT64 (主键)
  - content: VARCHAR (原始文本)
  - embedding: FLOAT_VECTOR (768维向量)
  - metadata: JSON (元数据：来源、类型等)
```

### 3.2 实现文件结构

```
internal/storage/milvus/
├── milvus.go         # Milvus客户端
├── embedding.go      # 向量化服务
└── search.go         # 搜索服务

internal/service/
├── rag.go            # RAG服务
└── knowledge.go      # 知识库管理
```

### 3.3 核心功能

- [x] 连接管理
- [x] 集合创建和管理
- [x] 向量插入和搜索
- [x] 嵌入模型集成
- [x] RAG流程实现

### 3.4 RAG流程

```go
// 1. 添加知识
func (s *KnowledgeService) AddKnowledge(ctx context.Context, content string) error {
    // 生成向量
    embedding := s.embeddingModel.Embed(ctx, content)
    
    // 插入Milvus
    return s.milvus.Insert(ctx, embedding, content)
}

// 2. 检索相关知识
func (s *KnowledgeService) Search(ctx context.Context, query string, topK int) ([]string, error) {
    // 查询向量化
    queryEmbedding := s.embeddingModel.Embed(ctx, query)
    
    // Milvus搜索
    return s.milvus.Search(ctx, queryEmbedding, topK)
}

// 3. RAG增强对话
func (s *ChatService) ChatWithRAG(ctx context.Context, chatbotID string, userMessage string) (*model.ChatResponse, error) {
    // 检索相关知识
    knowledge := s.knowledgeService.Search(ctx, userMessage, 3)
    
    // 构建增强的上下文
    messages := s.buildRAGMessages(chatbot, knowledge, history, userMessage)
    
    // 生成回答
    return s.model.Generate(ctx, messages)
}
```

---

## 🛠️ 实现步骤

### Step 1: 添加依赖

```bash
# MySQL驱动
go get github.com/go-sql-driver/mysql

# Redis客户端
go get github.com/redis/go-redis/v9

# Milvus客户端
go get github.com/milvus-io/milvus-sdk-go/v2

# 嵌入模型（用于向量化）
go get github.com/cloudwego/eino-ext/components/model/ollama
```

### Step 2: 实现MySQL存储

1. 创建 `internal/storage/mysql/mysql.go`
2. 实现 `Storage` 接口
3. 添加数据库连接池
4. 实现CRUD操作
5. 添加迁移脚本

### Step 3: 实现Redis存储

1. 创建 `internal/storage/redis/redis.go`
2. 实现缓存操作
3. 实现会话管理
4. 添加限流功能

### Step 4: 集成Milvus

1. 创建 `internal/storage/milvus/milvus.go`
2. 集成嵌入模型
3. 实现向量搜索
4. 实现RAG服务

### Step 5: 更新配置

```yaml
storage:
  type: "mysql"  # 或 "redis", "milvus"
  mysql:
    host: "47.118.19.28"
    port: 3307
    # ...
  redis:
    host: "47.118.19.28"
    port: 6379
    # ...
  milvus:
    host: "47.118.19.28"
    port: 19530
```

---

## 📊 测试计划

### MySQL测试
- [ ] 连接测试
- [ ] CRUD操作测试
- [ ] 并发测试
- [ ] 性能测试

### Redis测试
- [ ] 连接测试
- [ ] 缓存测试
- [ ] 会话测试
- [ ] 限流测试

### Milvus测试
- [ ] 连接测试
- [ ] 向量插入测试
- [ ] 搜索测试
- [ ] RAG流程测试

---

## 🎯 预期效果

### 功能增强
- ✅ 数据持久化（MySQL）
- ✅ 高性能缓存（Redis）
- ✅ 智能检索（Milvus）
- ✅ RAG能力（知识库问答）

### 性能提升
- ✅ 响应时间：减少50-80%
- ✅ 并发能力：提升10倍+
- ✅ 数据容量：支持TB级数据

### 用户体验
- ✅ 更快的响应速度
- ✅ 更准确的回答（RAG）
- ✅ 更好的个性化体验

