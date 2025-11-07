# API 文档

## 基础信息

- Base URL: `http://localhost:8080/api/v1`
- Content-Type: `application/json`

## 错误响应格式

```json
{
  "error": "error_code",
  "message": "详细错误信息"
}
```

## 端点列表

### 1. 创建聊天机器人

**POST** `/chatbots`

请求体：
```json
{
  "name": "机器人名称",
  "personality": "性格设定描述",
  "background": "背景设定描述"
}
```

响应：`201 Created`
```json
{
  "id": "uuid",
  "name": "机器人名称",
  "personality": "性格设定描述",
  "background": "背景设定描述",
  "system_prompt": "自动生成的系统提示词",
  "created_at": "2025-01-XX...",
  "updated_at": "2025-01-XX..."
}
```

### 2. 获取所有聊天机器人

**GET** `/chatbots`

响应：`200 OK`
```json
[
  {
    "id": "uuid",
    "name": "机器人名称",
    ...
  }
]
```

### 3. 获取指定聊天机器人

**GET** `/chatbots/{id}`

响应：`200 OK`
```json
{
  "id": "uuid",
  "name": "机器人名称",
  ...
}
```

### 4. 更新聊天机器人

**PUT** `/chatbots/{id}`

请求体：
```json
{
  "name": "新名称",
  "personality": "新性格",
  "background": "新背景"
}
```

### 5. 删除聊天机器人

**DELETE** `/chatbots/{id}`

响应：`200 OK`
```json
{
  "message": "deleted"
}
```

### 6. 进行对话

**POST** `/chatbots/{id}/chat`

请求体：
```json
{
  "message": "用户消息"
}
```

响应：`200 OK`
```json
{
  "message": "AI回复",
  "duration": 1234,
  "timestamp": "2025-01-XX..."
}
```

### 7. 获取对话历史

**GET** `/chatbots/{id}/history?limit=20`

查询参数：
- `limit`: 返回的记录数（默认20）

响应：`200 OK`
```json
[
  {
    "id": 1,
    "chatbot_id": "uuid",
    "user_message": "用户消息",
    "bot_message": "AI回复",
    "created_at": "2025-01-XX..."
  }
]
```

### 8. 健康检查

**GET** `/health`

响应：`200 OK`
```json
{
  "status": "ok",
  "service": "eino-chatbot"
}
```

## 使用示例

### cURL示例

```bash
# 创建聊天机器人
curl -X POST http://localhost:8080/api/v1/chatbots \
  -H "Content-Type: application/json" \
  -d '{
    "name": "技术助手",
    "personality": "专业、严谨、乐于助人",
    "background": "资深Go语言开发工程师"
  }'

# 进行对话
curl -X POST http://localhost:8080/api/v1/chatbots/{chatbot_id}/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Go语言有什么特点？"
  }'
```

### JavaScript示例

```javascript
// 创建聊天机器人
const chatbot = await fetch('http://localhost:8080/api/v1/chatbots', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    name: '技术助手',
    personality: '专业、严谨、乐于助人',
    background: '资深Go语言开发工程师'
  })
}).then(r => r.json());

// 进行对话
const response = await fetch(`http://localhost:8080/api/v1/chatbots/${chatbot.id}/chat`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    message: 'Go语言有什么特点？'
  })
}).then(r => r.json());

console.log(response.message);
```

