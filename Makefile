.PHONY: build run test clean docker-build docker-run

# 构建
build:
	go build -o bin/server ./cmd/server

# 运行
run:
	go run ./cmd/server/main.go

# 测试
test:
	go test ./...

# 清理
clean:
	rm -rf bin/
	rm -f main

# Docker构建
docker-build:
	docker build -t eino-chatbot:latest .

# Docker运行
docker-run:
	docker run -d \
		-p 8080:8080 \
		-v $(PWD)/configs:/app/configs \
		--name eino-chatbot \
		eino-chatbot:latest

# 格式化代码
fmt:
	go fmt ./...

# 代码检查
lint:
	golangci-lint run ./...

# 安装依赖
deps:
	go mod download
	go mod tidy

