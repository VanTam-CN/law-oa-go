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

# 运行所有测试
.PHONY: test
test:
	@echo "运行所有测试..."
	@go test -v -race -coverprofile=coverage.out ./...

# 只运行单元测试
.PHONY: test-unit
test-unit:
	@echo "运行单元测试..."
	@go test -v -race -coverprofile=unit_coverage.out ./internal/handlers/... ./internal/services/... ./internal/repositories/...

# 只运行集成测试
.PHONY: test-integration
test-integration:
	@echo "运行集成测试..."
	@go test -v -race -coverprofile=integration_coverage.out ./tests/integration/...

# 只运行E2E测试
.PHONY: test-e2e
test-e2e:
	@echo "运行端到端测试..."
	@go test -v -race -coverprofile=e2e_coverage.out ./tests/e2e/...

# 只运行性能测试
.PHONY: test-performance
test-performance:
	@echo "运行性能测试..."
	@go test -v -bench=. -benchmem -coverprofile=performance_coverage.out ./tests/performance/...

# 运行认证测试
.PHONY: test-auth
test-auth:
	@echo "运行认证相关测试..."
	@go test -v -race ./internal/handlers/auth_handler_test.go ./internal/services/user_service_test.go ./tests/integration/auth_integration_test.go

# 运行用户管理测试
.PHONY: test-users
test-users:
	@echo "运行用户管理测试..."
	@go test -v -race ./internal/handlers/user_handler_test.go ./internal/services/user_service_test.go

# 运行客户管理测试
.PHONY: test-clients
test-clients:
	@echo "运行客户管理测试..."
	@go test -v -race ./internal/handlers/client_handler_test.go ./internal/services/client_service_test.go

# 运行数据库测试
.PHONY: test-database
test-database:
	@echo "运行数据库相关测试..."
	@go test -v -race ./tests/integration/database_integration_test.go ./tests/performance/database_benchmark_test.go

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

# API基准测试
.PHONY: bench-api
bench-api:
	@echo "运行API基准测试..."
	@go test -bench=Benchmark -benchmem ./tests/performance/api_benchmark_test.go

# 数据库基准测试
.PHONY: bench-database
bench-database:
	@echo "运行数据库基准测试..."
	@go test -bench=Benchmark -benchmem ./tests/performance/database_benchmark_test.go

# 负载测试
.PHONY: test-load
test-load:
	@echo "运行负载测试..."
	@go test -v ./tests/performance/load_test.go -timeout=5m

# 并发用户测试
.PHONY: test-concurrent
test-concurrent:
	@echo "运行并发用户测试..."
	@go test -v -run=TestConcurrentUsers ./tests/performance/load_test.go -timeout=2m

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

# 数据库初始化
.PHONY: migrate-bootstrap migrate-up
migrate-bootstrap:
	@echo "执行 PostgreSQL 生产 schema bootstrap..."
	@go run ./cmd/migrate -command bootstrap

migrate-up: migrate-bootstrap

.PHONY: migrate-down
migrate-down:
	@echo "生产 bootstrap 不支持 down；请使用经过验证的数据库备份恢复。"
	@exit 1

.PHONY: qa-seed-conflict-p0 qa-verify-conflict-p0
qa-seed-conflict-p0:
	@echo "写入非生产 PostgreSQL 三角色冲突验收夹具；不会写入生产 schema bootstrap。"
	@go run ./cmd/qa-fixture -mode seed

qa-verify-conflict-p0:
	@echo "核验非生产 PostgreSQL 三角色冲突验收夹具。"
	@go run ./cmd/qa-fixture -mode verify

.PHONY: migrate-create
migrate-create:
	@echo "当前生产入口使用显式 PostgreSQL schema bootstrap，不在混合历史目录创建新迁移。"
	@exit 1

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
	@mkdir -p profiles/
	@go test -cpuprofile=profiles/cpu.prof -memprofile=profiles/mem.prof -bench=. ./...
	@echo "使用 'go tool pprof profiles/cpu.prof' 或 'go tool pprof profiles/mem.prof' 查看结果"

# Fuzzing配置
FUZZ_TIME ?= 30s
FUZZERS ?= $(shell nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)
FUZZ_OUTPUT_DIR ?= fuzzing-results

# PGO构建配置
BINARY_NAME := law-oa-server
PROFILE_DIR := profiles
DEFAULT_PGO_FILE := default.pgo

# PGO相关目标
.PHONY: pgo-build
pgo-build:
	@echo "PGO优化构建..."
	@mkdir -p bin/
ifeq ($(wildcard $(DEFAULT_PGO_FILE)),)
	@echo "警告: 未找到PGO配置文件 $(DEFAULT_PGO_FILE)"
	@echo "尝试使用剖析数据进行构建..."
	@go build \
		-tags="netgo osusergo" \
		-ldflags="$(LDFLAGS) -s -w" \
		-pgo="$(PROFILE_DIR)/cpu.prof" \
		-o bin/$(BINARY_NAME) \
		./main.go
else
	@echo "使用PGO配置文件: $(DEFAULT_PGO_FILE)"
	@go build \
		-tags="netgo osusergo" \
		-ldflags="$(LDFLAGS) -s -w" \
		-pgo="$(DEFAULT_PGO_FILE)" \
		-o bin/$(BINARY_NAME) \
		./main.go
endif
	@echo "PGO构建完成: bin/$(BINARY_NAME)"

.PHONY: pgo-full
pgo-full: profile pgo-build
	@echo "完整PGO构建流程完成"

.PHONY: pgo-test
pgo-test:
	@echo "使用测试数据进行PGO构建..."
	@mkdir -p $(PROFILE_DIR)
	@go test -cpuprofile=$(PROFILE_DIR)/test_cpu.prof ./...
	@go build \
		-tags="netgo osusergo" \
		-ldflags="$(LDFLAGS) -s -w" \
		-pgo="$(PROFILE_DIR)/test_cpu.prof" \
		-o bin/$(BINARY_NAME) \
		./main.go
	@echo "测试PGO构建完成: bin/$(BINARY_NAME)"

.PHONY: pgo-bench
pgo-bench:
	@echo "使用基准测试数据进行PGO构建..."
	@mkdir -p $(PROFILE_DIR)
	@go test -bench=. -benchmem -cpuprofile=$(PROFILE_DIR)/bench_cpu.prof ./...
	@go build \
		-tags="netgo osusergo" \
		-ldflags="$(LDFLAGS) -s -w" \
		-pgo="$(PROFILE_DIR)/bench_cpu.prof" \
		-o bin/$(BINARY_NAME) \
		./main.go
	@echo "基准测试PGO构建完成: bin/$(BINARY_NAME)"

# 运行工作负载生成器
.PHONY: workload
workload:
	@echo "运行工作负载生成器..."
	@mkdir -p $(PROFILE_DIR)
	@go run scripts/comprehensive_pgo_workload.go

# 运行HTTP工作负载
.PHONY: http-workload
http-workload:
	@echo "运行HTTP工作负载..."
	@mkdir -p $(PROFILE_DIR)
	@go run scripts/pgo_workload.go

# 快速PGO构建
.PHONY: quick-pgo
quick-pgo:
	@echo "快速PGO构建..."
	@chmod +x scripts/quick_pgo.sh
	@./scripts/quick_pgo.sh -p

# 检查PGO支持
.PHONY: check-pgo
check-pgo:
	@echo "检查PGO支持..."
	@echo "Go版本: $(GO_VERSION)"
	@if go build --help | grep -q "pgo"; then \
		echo "✓ PGO支持已启用"; \
	else \
		echo "✗ PGO支持未启用"; \
		exit 1; \
	fi

# PGO构建报告
.PHONY: pgo-report
pgo-report:
	@echo "生成PGO构建报告..."
	@mkdir -p bin/
	@echo "=== PGO构建报告 ===" > bin/pgo_report.txt
	@echo "项目: $(APP_NAME)" >> bin/pgo_report.txt
	@echo "版本: $(VERSION)" >> bin/pgo_report.txt
	@echo "构建时间: $(BUILD_TIME)" >> bin/pgo_report.txt
	@echo "Go版本: $(GO_VERSION)" >> bin/pgo_report.txt
	@echo "" >> bin/pgo_report.txt
	@echo "=== PGO配置状态 ===" >> bin/pgo_report.txt
	@if [ -f $(DEFAULT_PGO_FILE) ]; then \
		echo "PGO配置文件: 存在" >> bin/pgo_report.txt; \
	else \
		echo "PGO配置文件: 不存在" >> bin/pgo_report.txt; \
	fi
	@if [ -d $(PROFILE_DIR) ] && [ -n "$$(ls -A $(PROFILE_DIR)/*.prof 2>/dev/null)" ]; then \
		echo "剖析数据: 存在" >> bin/pgo_report.txt; \
		ls -la $(PROFILE_DIR)/*.prof >> bin/pgo_report.txt; \
	else \
		echo "剖析数据: 不存在" >> bin/pgo_report.txt; \
	fi
	@echo "PGO构建报告已生成: bin/pgo_report.txt"

# Fuzzing测试目标
.PHONY: fuzz-all
fuzz-all:
	@echo "运行所有Fuzzing测试..."
	@chmod +x scripts/fuzzing_test.sh
	@./scripts/fuzzing_test.sh -a -t $(FUZZ_TIME) -f $(FUZZERS) -o $(FUZZ_OUTPUT_DIR)

.PHONY: fuzz-security
fuzz-security:
	@echo "运行安全组件Fuzzing测试..."
	@chmod +x scripts/fuzzing_test.sh
	@./scripts/fuzzing_test.sh -s -t $(FUZZ_TIME) -f $(FUZZERS) -o $(FUZZ_OUTPUT_DIR)

.PHONY: fuzz-validators
fuzz-validators:
	@echo "运行验证器Fuzzing测试..."
	@chmod +x scripts/fuzzing_test.sh
	@./scripts/fuzzing_test.sh -v -t $(FUZZ_TIME) -f $(FUZZERS) -o $(FUZZ_OUTPUT_DIR)

.PHONY: fuzz-db
fuzz-db:
	@echo "运行数据库Fuzzing测试..."
	@chmod +x scripts/fuzzing_test.sh
	@./scripts/fuzzing_test.sh -r -t $(FUZZ_TIME) -f $(FUZZERS) -o $(FUZZ_OUTPUT_DIR)

.PHONY: fuzz-cache
fuzz-cache:
	@echo "运行缓存Fuzzing测试..."
	@chmod +x scripts/fuzzing_test.sh
	@./scripts/fuzzing_test.sh -c -t $(FUZZ_TIME) -f $(FUZZERS) -o $(FUZZ_OUTPUT_DIR)

.PHONY: fuzz-concurrent
fuzz-concurrent:
	@echo "运行并发Fuzzing测试..."
	@chmod +x scripts/fuzzing_test.sh
	@./scripts/fuzzing_test.sh -n -t $(FUZZ_TIME) -f $(FUZZERS) -o $(FUZZ_OUTPUT_DIR)

.PHONY: fuzz-quick
fuzz-quick:
	@echo "运行快速Fuzzing测试..."
	@chmod +x scripts/fuzzing_test.sh
	@./scripts/fuzzing_test.sh -a -t 10s -f $(FUZZERS) -o $(FUZZ_OUTPUT_DIR)

.PHONY: fuzz-long
fuzz-long:
	@echo "运行长时间Fuzzing测试..."
	@chmod +x scripts/fuzzing_test.sh
	@./scripts/fuzzing_test.sh -a -t 10m -f $(FUZZERS) -o $(FUZZ_OUTPUT_DIR)

.PHONY: fuzz-clean
fuzz-clean:
	@echo "清理Fuzzing缓存和结果..."
	@go clean -fuzzcache
	@rm -rf $(FUZZ_OUTPUT_DIR)/
	@find . -name "*fuzz*" -type d -exec rm -rf {} + 2>/dev/null || true
	@echo "Fuzzing清理完成"

# 手动Fuzzing测试（用于调试）
.PHONY: fuzz-manual-jwt
fuzz-manual-jwt:
	@echo "手动运行JWT Fuzzing测试..."
	@go test -fuzz=Fuzz_JWTKeyManager_ValidateToken -fuzztime=$(FUZZ_TIME) -v ./internal/security/

.PHONY: fuzz-manual-validators
fuzz-manual-validators:
	@echo "手动运行验证器Fuzzing测试..."
	@go test -fuzz=Fuzz_CaseValidator_Validate -fuzztime=$(FUZZ_TIME) -v ./internal/validators/
	@go test -fuzz=Fuzz_ClientValidator_Validate -fuzztime=$(FUZZ_TIME) -v ./internal/validators/
	@go test -fuzz=Fuzz_LawyerValidator_Validate -fuzztime=$(FUZZ_TIME) -v ./internal/validators/

.PHONY: fuzz-manual-db
fuzz-manual-db:
	@echo "手动运行数据库Fuzzing测试..."
	@go test -fuzz=Fuzz_QueryBuilder_Where -fuzztime=$(FUZZ_TIME) -v ./internal/repositories/
	@go test -fuzz=Fuzz_QueryBuilder_Like -fuzztime=$(FUZZ_TIME) -v ./internal/repositories/
	@go test -fuzz=Fuzz_QueryBuilder_Order -fuzztime=$(FUZZ_TIME) -v ./internal/repositories/

.PHONY: fuzz-manual-cache
fuzz-manual-cache:
	@echo "手动运行缓存Fuzzing测试..."
	@go test -fuzz=Fuzz_CacheService_SetAndGet -fuzztime=$(FUZZ_TIME) -v ./internal/cache/
	@go test -fuzz=Fuzz_CacheService_SetWithExpiration -fuzztime=$(FUZZ_TIME) -v ./internal/cache/
	@go test -fuzz=Fuzz_CacheService_ConcurrentAccess -fuzztime=$(FUZZ_TIME) -v ./internal/cache/
	@go test -fuzz=Fuzz_LayeredCache_Get -fuzztime=$(FUZZ_TIME) -v ./internal/cache/

.PHONY: fuzz-manual-concurrent
fuzz-manual-concurrent:
	@echo "手动运行并发Fuzzing测试..."
	@go test -fuzz=Fuzz_WorkerPool_SubmitTask -fuzztime=$(FUZZ_TIME) -v ./internal/concurrency/
	@go test -fuzz=Fuzz_CircuitBreaker_Execute -fuzztime=$(FUZZ_TIME) -v ./internal/concurrency/
	@go test -fuzz=Fuzz_RateLimiter_Allow -fuzztime=$(FUZZ_TIME) -v ./internal/concurrency/
	@go test -fuzz=Fuzz_ConcurrentService_SubmitTask -fuzztime=$(FUZZ_TIME) -v ./internal/concurrency/

# 检查Fuzzing支持
.PHONY: check-fuzz
check-fuzz:
	@echo "检查Fuzzing支持..."
	@echo "Go版本: $(GO_VERSION)"
	@if go help | grep -q "fuzz"; then \
		echo "✓ Fuzzing支持已启用"; \
	else \
		echo "✗ Fuzzing支持未启用，请使用Go 1.23+"; \
		exit 1; \
	fi

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
	@echo "  make test          - 运行所有测试"
	@echo "  make test-unit     - 只运行单元测试"
	@echo "  make test-integration - 只运行集成测试"
	@echo "  make test-e2e      - 只运行端到端测试"
	@echo "  make test-performance - 只运行性能测试"
	@echo "  make test-auth     - 运行认证相关测试"
	@echo "  make test-users    - 运行用户管理测试"
	@echo "  make test-clients  - 运行客户管理测试"
	@echo "  make test-database - 运行数据库相关测试"
	@echo "  make test-coverage - 生成测试覆盖率报告"
	@echo "  make bench         - 运行基准测试"
	@echo "  make bench-api     - 运行API基准测试"
	@echo "  make bench-database - 运行数据库基准测试"
	@echo "  make test-load     - 运行负载测试"
	@echo "  make test-concurrent - 运行并发用户测试"
	@echo "  make build         - 构建应用"
	@echo "  make build-linux   - 构建Linux版本"
	@echo "  make docker-build  - 构建Docker镜像"
	@echo "  make run           - 运行应用"
	@echo "  make dev           - 开发模式运行"
	@echo "  make migrate-bootstrap - 初始化 PostgreSQL 生产 schema"
	@echo "  make migrate-up    - 执行 schema bootstrap（同 migrate-bootstrap）"
	@echo "  make migrate-down  - 拒绝生产回滚，改用备份恢复"
	@echo "  make qa-seed-conflict-p0 - 写入非生产三角色冲突验收夹具"
	@echo "  make qa-verify-conflict-p0 - 核验非生产三角色冲突验收夹具"
	@echo "  make docs          - 生成API文档"
	@echo "  make security      - 安全检查"
	@echo "  make profile       - 性能分析"
	@echo "  make quality       - 代码质量检查"
	@echo "  make ci            - CI流程"
	@echo "  make release       - 发布准备"
	@echo "  make clean         - 清理构建文件"
	@echo ""
	@echo "PGO相关命令:"
	@echo "  make pgo-build     - PGO优化构建"
	@echo "  make pgo-full      - 完整PGO构建流程（生成剖析+构建）"
	@echo "  make pgo-test      - 使用测试数据进行PGO构建"
	@echo "  make pgo-bench     - 使用基准测试数据进行PGO构建"
	@echo "  make workload       - 运行工作负载生成器"
	@echo "  make http-workload - 运行HTTP工作负载"
	@echo "  make quick-pgo      - 快速PGO构建"
	@echo "  make check-pgo      - 检查PGO支持"
	@echo "  make pgo-report     - 生成PGO构建报告"
	@echo ""
	@echo "Fuzzing测试命令:"
	@echo "  make fuzz-all       - 运行所有Fuzzing测试"
	@echo "  make fuzz-security  - 运行安全组件Fuzzing测试"
	@echo "  make fuzz-validators - 运行验证器Fuzzing测试"
	@echo "  make fuzz-db        - 运行数据库Fuzzing测试"
	@echo "  make fuzz-cache     - 运行缓存Fuzzing测试"
	@echo "  make fuzz-concurrent - 运行并发Fuzzing测试"
	@echo "  make fuzz-quick     - 快速Fuzzing测试"
	@echo "  make fuzz-long      - 长时间Fuzzing测试"
	@echo "  make fuzz-clean     - 清理Fuzzing缓存和结果"
	@echo ""
	@echo "  make help          - 显示帮助信息"
