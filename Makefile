# Resend邮箱验证码审核页系统构建脚本

.PHONY: build run test clean setup docker-build docker-run

# 默认目标
all: setup build

# 安装依赖
setup:
	go mod tidy
	go mod download

# 构建项目
build:
	go build -o bin/server ./cmd/server

# 运行项目
run:
	go run ./cmd/server

# 测试
test:
	go test ./... -v

# 清理
clean:
	rm -rf bin/ data/ *.db

# 创建数据目录
init-data:
	mkdir -p data

# 创建配置文件
init-config:
	cp .env.example .env
	@echo "请编辑 .env 文件配置相关参数"

# Docker构建
docker-build:
	docker build -t resend-auth-system:latest .

# Docker运行
docker-run:
	docker run -p 8080:8080 --env-file .env resend-auth-system:latest

# 开发模式运行
dev:
	APP_ENV=development go run ./cmd/server

# 生产模式运行
prod:
	APP_ENV=production go run ./cmd/server

# 生成Swagger文档
swagger:
	swag init -g cmd/server/main.go -o docs

# 代码格式化
fmt:
	go fmt ./...

# 代码检查
lint:
	golangci-lint run ./...

# 数据库迁移
migrate:
	go run migrations/main.go

# 帮助信息
help:
	@echo "可用命令:"
	@echo "  make setup      - 安装依赖"
	@echo "  make build      - 构建项目"
	@echo "  make run        - 运行项目"
	@echo "  make test       - 运行测试"
	@echo "  make clean      - 清理构建文件"
	@echo "  make dev        - 开发模式运行"
	@echo "  make prod       - 生产模式运行"
	@echo "  make fmt        - 格式化代码"
	@echo "  make lint       - 代码检查"
	@echo "  make docker-build - Docker构建"
	@echo "  make docker-run - Docker运行"