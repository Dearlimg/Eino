# 部署文档

## 本地开发

### 1. 环境准备

```bash
# 安装Go 1.21+
go version

# 安装Ollama（可选，用于本地模型）
# macOS
brew install ollama

# 启动Ollama服务
ollama serve

# 下载模型
ollama pull deepseek-r1:8b
```

### 2. 配置

编辑 `configs/config.yaml`，根据实际情况修改配置：

```yaml
model:
  provider: "ollama"
  base_url: "http://localhost:11434"
  model: "deepseek-r1:8b"
```

### 3. 运行

```bash
# 方式1: 直接运行
go run cmd/server/main.go

# 方式2: 使用Makefile
make run

# 方式3: 构建后运行
make build
./bin/server
```

## Docker部署

### 1. 构建镜像

```bash
docker build -t eino-chatbot:latest .
```

### 2. 运行容器

```bash
docker run -d \
  --name eino-chatbot \
  -p 8080:8080 \
  -v $(pwd)/configs:/app/configs \
  eino-chatbot:latest
```

### 3. 查看日志

```bash
docker logs -f eino-chatbot
```

## 服务器部署（阿里云ECS）

### 1. 准备服务器环境

```bash
# 连接到服务器
ssh root@47.118.19.28

# 安装Go
wget https://go.dev/dl/go1.21.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.21.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# 或使用Docker（推荐）
```

### 2. 使用Docker部署

```bash
# 在服务器上
cd /opt
git clone <your-repo>
cd eino

# 构建镜像
docker build -t eino-chatbot:latest .

# 运行容器
docker run -d \
  --name eino-chatbot \
  --restart=always \
  -p 8080:8080 \
  -v /opt/eino/configs:/app/configs \
  eino-chatbot:latest
```

### 3. 配置Nginx反向代理（可选）

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

### 4. 配置防火墙

```bash
# 开放8080端口
firewall-cmd --permanent --add-port=8080/tcp
firewall-cmd --reload
```

## 使用MySQL存储

### 1. 创建数据库

```sql
CREATE DATABASE eino_chatbot CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE eino_chatbot;

CREATE TABLE chatbots (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    personality TEXT,
    background TEXT,
    system_prompt TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE conversations (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    chatbot_id VARCHAR(36) NOT NULL,
    user_message TEXT NOT NULL,
    bot_message TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_chatbot_id (chatbot_id),
    INDEX idx_created_at (created_at)
);
```

### 2. 更新配置

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

### 3. 实现MySQL存储层

需要实现 `internal/storage/mysql/mysql.go`（TODO）

## 使用Redis存储

### 1. 更新配置

```yaml
storage:
  type: "redis"
  redis:
    host: "47.118.19.28"
    port: 6379
    password: ""
    db: 0
```

### 2. 实现Redis存储层

需要实现 `internal/storage/redis/redis.go`（TODO）

## 使用Milvus向量数据库（RAG支持）

### 1. 配置Milvus

Milvus已经在服务器上运行（端口19530），可以直接使用。

### 2. 更新配置

```yaml
storage:
  milvus:
    host: "47.118.19.28"
    port: 19530
```

### 3. 实现Milvus集成

需要实现向量存储和检索功能（TODO）

## 监控和日志

### 1. 查看应用日志

```bash
# Docker容器日志
docker logs -f eino-chatbot

# 系统日志
journalctl -u eino-chatbot -f
```

### 2. 健康检查

```bash
curl http://localhost:8080/health
```

## 性能优化

1. **连接池**：配置数据库连接池大小
2. **缓存**：使用Redis缓存常用数据
3. **限流**：实现API限流保护
4. **监控**：集成Prometheus和Grafana

## 故障排查

### 常见问题

1. **模型连接失败**
   - 检查Ollama服务是否运行
   - 检查模型名称是否正确
   - 检查网络连接

2. **存储错误**
   - 检查数据库连接配置
   - 检查表结构是否正确

3. **端口占用**
   - 修改配置文件中的端口
   - 或停止占用端口的服务

## 备份和恢复

### 备份

```bash
# 备份数据库
mysqldump -u root -p eino_chatbot > backup.sql

# 备份配置文件
tar -czf configs-backup.tar.gz configs/
```

### 恢复

```bash
# 恢复数据库
mysql -u root -p eino_chatbot < backup.sql
```

