package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"eino/internal/agent"
	"eino/internal/config"
	"eino/internal/handler"
	"eino/internal/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化存储
	storageService, err := storage.NewStorage(cfg.Storage)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storageService.Close()

	// 初始化聊天机器人服务
	chatService, err := agent.NewChatService(cfg, storageService)
	if err != nil {
		log.Fatalf("Failed to initialize chat service: %v", err)
	}

	// 初始化HTTP处理器
	router := gin.Default()
	handler.RegisterRoutes(router, chatService)

	// 启动HTTP服务器
	srv := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: router,
	}

	// 优雅关闭
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server started on %s", cfg.Server.Address)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
