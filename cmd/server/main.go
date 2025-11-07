package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"eino/internal/agent"
	"eino/internal/config"
	"eino/internal/handler"
	"eino/internal/service"
	"eino/internal/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	// 支持从环境变量或命令行参数获取配置路径
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	cfg, err := config.Load(configPath)
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

	// 初始化RAG服务（如果启用）
	if cfg.RAG.Enabled {
		ragService, err := service.NewRAGService(
			cfg.Storage.Milvus.Host,
			cfg.Storage.Milvus.Port,
			cfg.RAG.OllamaURL,
		)
		if err != nil {
			log.Printf("Warning: Failed to initialize RAG service: %v", err)
		} else {
			chatService.SetRAGService(ragService)
			log.Println("RAG service enabled")
		}
	}

	// 初始化HTTP处理器
	router := gin.Default()

	// 静态文件服务（前端页面）
	// 查找web目录，支持多种运行方式
	workDir, _ := os.Getwd()
	var webDir string

	// 尝试多个可能的路径
	possiblePaths := []string{
		filepath.Join(workDir, "web"),
		filepath.Join(workDir, "..", "web"),
		filepath.Join(workDir, "..", "..", "web"),
		filepath.Join(workDir, "..", "..", "..", "web"),
	}

	for _, path := range possiblePaths {
		absPath, _ := filepath.Abs(path)
		indexPath := filepath.Join(absPath, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			webDir = absPath
			log.Printf("Found web directory at: %s", webDir)
			break
		}
	}

	// 注册API路由（先注册，避免被静态文件路由拦截）
	handler.RegisterRoutes(router, chatService)

	if webDir == "" {
		log.Printf("Warning: web directory not found, static files will not be served")
		log.Printf("Searched in: %v", possiblePaths)
	} else {
		// 注册静态文件路由
		router.Static("/web", webDir)
		// 根路径返回index.html
		router.GET("/", func(c *gin.Context) {
			c.File(filepath.Join(webDir, "index.html"))
		})
		log.Printf("Static files served from: %s", webDir)
	}

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
