# Law OA Go Project Makefile

# 变量定义
APP_NAME := law-oa-go
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date +%Y-%m-%d_%H:%M:%S)
GO_VERSION := $(shell go version | awk '{print $$3}')

# 构建标志
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GoVersion=$(GO_VERSION)"

# 默认目标
.PHONY: all
all: clean test build

# 清理
.PHONY: clean
clean:
	@echo "清理构建文件..."
	@rm -rf bin/
	@rm -rf tmp/
	@go clean -cache
	@go clean -testcache

# 安装依赖
.PHONY: deps
deps:
	@echo "安装依赖..."
	@go mod download
	@go mod tidy

# 代码格式化
.PHONY: fmt
fmt:
	@echo "格式化代码..."
	@go fmt ./...
	@goimports -w .

# 代码检查
.PHONY: lint
lint:
	@echo "代码检查..."
	@go vet ./...
	@/Users/mac/go/bin/golangci-lint run

# 运行测试
.PHONY: test
test:
	@echo "运行测试..."
	@go test -v -race -coverprofile=coverage.out ./...

# 测试覆盖率
.PHONY: test-coverage
test-coverage: test
	@echo "生成测试覆盖率报告..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

# 基准测试
.PHONY: bench
bench:
	@echo "运行基准测试..."
	@go test -bench=. -benchmem ./...

# 构建
.PHONY: build
build:
	@echo "构建应用..."
	@mkdir -p bin/
	@go build $(LDFLAGS) -o bin/$(APP_NAME) .

# 构建Linux版本
.PHONY: build-linux
build-linux:
	@echo "构建Linux版本..."
	@mkdir -p bin/
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-linux .

# 构建Docker镜像
.PHONY: docker-build
docker-build:
	@echo "构建Docker镜像..."
	@docker build -t $(APP_NAME):$(VERSION) .
	@docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest

# 运行应用
.PHONY: run
run:
	@echo "运行应用..."
	@go run . --config=config/config.yaml

# 开发模式运行
.PHONY: dev
dev:
	@echo "开发模式运行..."
	@air -c .air.toml

# 数据库迁移
.PHONY: migrate-up
migrate-up:
	@echo "执行数据库迁移..."
	@migrate -path migrations -database "mysql://root:@tcp(localhost:3306)/law_oa" up

.PHONY: migrate-down
migrate-down:
	@echo "回滚数据库迁移..."
	@migrate -path migrations -database "mysql://root:@tcp(localhost:3306)/law_oa" down

.PHONY: migrate-create
migrate-create:
	@echo "创建迁移文件: $(name)"
	@migrate create -ext sql -dir migrations $(name)

# 生成API文档
.PHONY: docs
docs:
	@echo "生成API文档..."
	@swag init -g main.go -o docs/

# 安全检查
.PHONY: security
security:
	@echo "安全检查..."
	@gosec ./...

# 性能分析
.PHONY: profile
profile:
	@echo "性能分析..."
	@go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./...
	@echo "使用 'go tool pprof cpu.prof' 或 'go tool pprof mem.prof' 查看结果"

# 代码质量检查
.PHONY: quality
quality: fmt lint test security
	@echo "代码质量检查完成"

# 完整构建流程
.PHONY: ci
ci: deps quality build test-coverage
	@echo "CI流程完成"

# 部署准备
.PHONY: release
release: clean ci build-linux docker-build
	@echo "发布准备完成"

# 帮助信息
.PHONY: help
help:
	@echo "可用的命令:"
	@echo "  make deps          - 安装依赖"
	@echo "  make fmt           - 格式化代码"
	@echo "  make lint          - 代码检查"
	@echo "  make test          - 运行测试"
	@echo "  make test-coverage - 生成测试覆盖率报告"
	@echo "  make bench         - 运行基准测试"
	@echo "  make build         - 构建应用"
	@echo "  make build-linux   - 构建Linux版本"
	@echo "  make docker-build  - 构建Docker镜像"
	@echo "  make run           - 运行应用"
	@echo "  make dev           - 开发模式运行"
	@echo "  make migrate-up    - 执行数据库迁移"
	@echo "  make migrate-down  - 回滚数据库迁移"
	@echo "  make docs          - 生成API文档"
	@echo "  make security      - 安全检查"
	@echo "  make profile       - 性能分析"
	@echo "  make quality       - 代码质量检查"
	@echo "  make ci            - CI流程"
	@echo "  make release       - 发布准备"
	@echo "  make clean         - 清理构建文件"
	@echo "  make help          - 显示帮助信息"