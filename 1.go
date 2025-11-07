package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/schema"
)

func main() {
	ctx := context.Background()

	// 1. 初始化模型
	model, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: "http://localhost:11434",
		Model:   "deepseek-r1:8b", // 或其他可用模型
	})
	if err != nil {
		log.Fatal(err)
	}

	// 2. 构建消息
	messages := []*schema.Message{
		schema.SystemMessage("你是一个友好的助手"),
		schema.UserMessage("你好，介绍一下你自己"),
	}

	// 3. 生成回复
	response, err := model.Generate(ctx, messages)
	if err != nil {
		log.Fatal(err)
	}

	// 4. 输出结果
	fmt.Println(response.Content)
}
