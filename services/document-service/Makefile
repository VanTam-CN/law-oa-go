.PHONY: help build run test clean dev docker-build docker-up docker-down migrate lint format

# 默认目标
help:
	@echo "Available commands:"
	@echo "  make build        - Build the application"
	@echo "  make run          - Run the application"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make dev          - Run in development mode"
	@echo "  make migrate      - Run database migrations"
	@echo "  make lint         - Run linter"
	@echo "  make format       - Format code"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-up    - Start Docker Compose services"
	@echo "  make docker-down  - Stop Docker Compose services"
	@echo "  make swagger      - Generate Swagger documentation"

# 构建
build:
	go build -o bin/document-service cmd/server/main.go

# 运行
run:
	./bin/document-service

# 开发模式
dev:
	go run cmd/server/main.go

# 测试
test:
	go test -v ./...

# 测试覆盖率
test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# 清理
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -cache

# 代码格式化
format:
	go fmt ./...
	goimports -w .

# 代码检查
lint:
	golangci-lint run ./...

# 数据库迁移
migrate:
	@echo "Running database migrations..."
	go run cmd/migrate/main.go

# Docker 构建
docker-build:
	docker build -t document-service:latest .

# Docker 启动
docker-up:
	docker-compose up -d

# Docker 停止
docker-down:
	docker-compose down

# Docker 重启
docker-restart:
	docker-compose restart

# Docker 日志
docker-logs:
	docker-compose logs -f document-service

# 数据库连接测试
db-test:
	mysql -h localhost -P 3306 -u root -ppassword -e "SELECT 1"

# Redis 连接测试
redis-test:
	redis-cli -h localhost -p 6379 ping

# Elasticsearch 健康检查
es-health:
	curl -X GET "localhost:9200/_cluster/health?pretty"

# MinIO 健康检查
minio-health:
	curl -X GET "http://localhost:9000/minio/health/live"

# 生成 Swagger 文档
swagger:
	swag init -g cmd/server/main.go -o api/openapi

# 生成依赖
deps:
	go mod tidy
	go mod vendor

# 安装开发工具
install-tools:
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

# 安全扫描
security:
	gosec ./...

# 性能分析
pprof:
	go tool pprof http://localhost:8080/debug/pprof/profile

# 负载测试
load-test:
	hey -n 1000 -c 10 http://localhost:8080/health

# 创建环境配置文件
setup-env:
	cp configs/config.example.yaml configs/config.yaml
	@echo "Please edit configs/config.yaml with your environment settings"

# 数据库种子数据
seed:
	go run cmd/seed/main.go

# 备份数据库
backup:
	mysqldump -h localhost -P 3306 -u root -ppassword document_service > backup_$(shell date +%Y%m%d_%H%M%S).sql

# 恢复数据库
restore:
	@read -p "Enter backup file: " backup_file; \
	mysqldump -h localhost -P 3306 -u root -ppassword document_service < $$backup_file

# 监控指标
metrics:
	curl http://localhost:8080/metrics

# 生产部署检查
prod-check:
	@echo "Checking production readiness..."
	@echo "1. Environment variables:"
	@env | grep DOC_SERVICE || echo "No DOC_SERVICE variables found"
	@echo "2. Service health:"
	curl -s http://localhost:8080/health || echo "Service not responding"
	@echo "3. Database connection:"
	mysqladmin ping -h localhost -P 3306 -u root -ppassword || echo "Database not responding"
	@echo "4. Redis connection:"
	redis-cli -h localhost -p 6379 ping || echo "Redis not responding"
	@echo "5. Elasticsearch health:"
	curl -s http://localhost:9200/_cluster/health || echo "Elasticsearch not responding"
	@echo "6. MinIO health:"
	curl -s http://localhost:9000/minio/health/live || echo "MinIO not responding"

# 版本信息
version:
	@echo "Document Service Version:"
	@git describe --tags --always 2>/dev/null || echo "Unknown (no git)"

# 发布前检查
pre-release: lint test security prod-check
	@echo "Pre-release checks completed!"

# 完整的开发环境搭建
setup-dev: install-tools setup-env docker-up
	@echo "Development environment setup completed!"
	@echo "Please wait for services to start, then run 'make migrate'"