#!/bin/bash

# Eino聊天机器人示例脚本

BASE_URL="http://localhost:8080/api/v1"

echo "=== 创建聊天机器人 ==="
CHATBOT_RESPONSE=$(curl -s -X POST "${BASE_URL}/chatbots" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "技术助手",
    "personality": "你是一个专业、严谨、乐于助人的技术顾问。说话简洁明了，喜欢用代码示例说明问题。",
    "background": "你是一位拥有10年经验的Go语言开发工程师，精通分布式系统、微服务架构和云原生技术。"
  }')

echo "$CHATBOT_RESPONSE" | jq '.'

CHATBOT_ID=$(echo "$CHATBOT_RESPONSE" | jq -r '.id')
echo -e "\n聊天机器人ID: $CHATBOT_ID\n"

echo "=== 进行对话 ==="
curl -s -X POST "${BASE_URL}/chatbots/${CHATBOT_ID}/chat" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "你好，介绍一下你自己"
  }' | jq '.'

echo -e "\n=== 继续对话 ==="
curl -s -X POST "${BASE_URL}/chatbots/${CHATBOT_ID}/chat" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Go语言有什么优势？"
  }' | jq '.'

echo -e "\n=== 获取对话历史 ==="
curl -s -X GET "${BASE_URL}/chatbots/${CHATBOT_ID}/history?limit=10" | jq '.'

