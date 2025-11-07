# 快速启动指南

## 5分钟快速开始

### 1. 确保Ollama运行

```bash
# 检查Ollama是否运行
curl http://localhost:11434/api/tags

# 如果没有运行，启动Ollama
ollama serve

# 下载模型（新终端）
ollama pull deepseek-r1:8b
```

### 2. 启动服务

```bash
# 方式1: 直接运行
go run cmd/server/main.go

# 方式2: 使用Makefile
make run
```

### 3. 测试API

```bash
# 健康检查
curl http://localhost:8080/health

# 创建聊天机器人
curl -X POST http://localhost:8080/api/v1/chatbots \
  -H "Content-Type: application/json" \
  -d '{
    "name": "技术助手",
    "personality": "你是一个专业、严谨、乐于助人的技术顾问。",
    "background": "你是一位拥有10年经验的Go语言开发工程师。"
  }'

# 保存返回的chatbot_id，然后进行对话
curl -X POST http://localhost:8080/api/v1/chatbots/{chatbot_id}/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "你好，介绍一下你自己"
  }'
```

### 4. 使用示例脚本

```bash
# 确保安装了jq
brew install jq  # macOS
# 或 apt-get install jq  # Linux

# 运行示例脚本
./scripts/example.sh
```

## 常见问题

### Q: 连接Ollama失败？

A: 检查：
1. Ollama服务是否运行：`curl http://localhost:11434/api/tags`
2. 配置文件中的`base_url`是否正确
3. 模型名称是否正确（使用`ollama list`查看可用模型）

### Q: 端口被占用？

A: 修改`configs/config.yaml`中的端口：
```yaml
server:
  address: ":8081"  # 改为其他端口
```

### Q: 如何查看日志？

A: 服务日志会直接输出到控制台。如果使用Docker：
```bash
docker logs -f eino-chatbot
```

## 下一步

- 查看[完整API文档](docs/API.md)
- 查看[部署文档](docs/DEPLOY.md)
- 阅读[README](README.md)了解更多功能

