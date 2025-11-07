# 存储扩展使用指南

## 📋 快速配置

### MySQL存储配置

1. **创建数据库**
```bash
mysql -h 47.118.19.28 -P 3307 -u root -p < migrations/001_init.sql
```

2. **更新配置**
```yaml
storage:
  type: "mysql"
  mysql:
    host: "47.118.19.28"
    port: 3307
    user: "root"
    password: "your_password"
    database: "eino_chatbot"
```

### Redis存储配置

1. **更新配置**
```yaml
storage:
  type: "redis"
  redis:
    host: "47.118.19.28"
    port: 6379
    password: ""
    db: 0
```

### Milvus + RAG配置

1. **确保Milvus运行**
```bash
# 检查Milvus状态
curl http://47.118.19.28:9091/healthz
```

2. **下载嵌入模型**
```bash
ollama pull nomic-embed-text
```

3. **更新配置**
```yaml
storage:
  milvus:
    host: "47.118.19.28"
    port: 19530

rag:
  enabled: true
  ollama_url: "http://localhost:11434"
  embedding_model: "nomic-embed-text"
```

## 🚀 使用示例

### 添加知识到向量库

```bash
curl -X POST http://localhost:8080/api/v1/knowledge \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Go语言是Google开发的编程语言，具有并发、简洁、快速编译的特点。"
  }'
```

### 搜索知识

```bash
curl "http://localhost:8080/api/v1/knowledge/search?q=Go语言特点"
```

### 使用RAG增强的对话

启用RAG后，对话会自动检索相关知识并增强回答：

```bash
curl -X POST http://localhost:8080/api/v1/chatbots/{id}/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Go语言有什么特点？"
  }'
```

系统会自动：
1. 从向量库检索相关知识
2. 将知识添加到上下文
3. LLM基于知识生成准确回答

## 📊 存储选择建议

- **开发/测试**：使用 `memory`（快速启动）
- **生产环境（小规模）**：使用 `mysql`（数据持久化）
- **生产环境（高并发）**：使用 `mysql` + `redis`（缓存加速）
- **需要知识库问答**：启用 `rag` + `milvus`

