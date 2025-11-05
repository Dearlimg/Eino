package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/schema"
)

func main() {
	// 创建上下文
	ctx := context.Background()

	// 初始化 Ollama ChatModel
	chatModel, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: "http://localhost:11434", // Ollama 服务地址
		Model:   "deepseek-r1:8b",         // 模型名称，请根据你的实际情况修改
	})
	if err != nil {
		log.Fatalf("初始化 ChatModel 失败: %v", err)
	}

	// 构建对话消息
	messages := []*schema.Message{
		schema.UserMessage(""),
	}

	// 生成回复
	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		log.Fatalf("生成回复失败: %v", err)
	}

	// 输出回复内容
	fmt.Println("=== Ollama 回复 ===")
	fmt.Println(response.Content)
}
