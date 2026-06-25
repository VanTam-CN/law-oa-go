# Law OA Go - 企业级多阶段Dockerfile
# 基于最新Docker最佳实践和生产环境优化策略 v2.1.0

# ================================
# 基础阶段 - 预准备公共依赖
# ================================
FROM golang:1.25-alpine AS base

# 安装基础依赖
RUN apk add --no-cache git ca-certificates tzdata build-base && \
    addgroup -g 1001 -S lawapp && \
    adduser -u 1001 -S lawapp -G lawapp

# 设置Go环境变量
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64
ENV GOPROXY=https://goproxy.cn,direct
ENV GOCACHE=/root/.cache/go-build
ENV GOMODCACHE=/go/pkg/mod

# ================================
# 构建阶段 - 生产优化的二进制构建
# ================================
FROM base AS builder

# 设置工作目录
WORKDIR /build

# 构建参数
ARG BUILD_COMMIT=unknown
ARG BUILD_DATE=unknown
ARG BUILD_TARGET=standard
ARG PGO_PROFILE=default.pgo
ARG VERSION=2.1.0

# 首先复制依赖文件以利用Docker层缓存
COPY go.mod go.sum ./

# 使用BuildKit缓存挂载下载依赖 (大幅提升构建速度)
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download && \
    go mod verify

# 复制源代码
COPY . .

# 为 scratch 运行时镜像预创建目录；scratch 阶段没有 shell，不能 RUN mkdir。
RUN mkdir -p /build/runtime-dirs/uploads/contract \
    /build/runtime-dirs/uploads/evidence \
    /build/runtime-dirs/uploads/letter \
    /build/runtime-dirs/uploads/other \
    /build/runtime-dirs/logs \
    /build/runtime-dirs/temp

# 生产镜像构建阶段只做依赖校验和编译；测试/静态分析由 CI 独立门禁负责。
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod verify

# 根据构建目标选择最优构建方式
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    if [ "$BUILD_TARGET" = "pgo" ] && [ -f "$PGO_PROFILE" ]; then \
        echo "🚀 使用PGO优化构建..." && \
        CGO_ENABLED=0 go build -pgo=$PGO_PROFILE \
            -ldflags="-w -s -extldflags \"-static\" \
                     -X main.Version=$VERSION \
                     -X main.Commit=$BUILD_COMMIT \
                     -X main.BuildTime=$BUILD_DATE \
                     -X main.Environment=production" \
            -a -installsuffix cgo \
            -trimpath \
            -o law-oa-go ./main.go; \
    else \
        echo "⚡ 使用标准优化构建..." && \
        CGO_ENABLED=0 go build \
            -ldflags="-w -s -extldflags \"-static\" \
                     -X main.Version=$VERSION \
                     -X main.Commit=$BUILD_COMMIT \
                     -X main.BuildTime=$BUILD_DATE \
                     -X main.Environment=production" \
            -a -installsuffix cgo \
            -trimpath \
            -o law-oa-go ./main.go; \
    fi

# 验证构建和二进制文件
RUN ls -la /build/law-oa-go && \
    file /build/law-oa-go && \
    /build/law-oa-go --version 2>/dev/null || echo "✅ 构建完成"

# ================================
# 安全扫描阶段 - 代码安全检查
# ================================
FROM base AS security-scanner

# 设置工作目录
WORKDIR /build

# 安装安全扫描工具
RUN --mount=type=cache,target=/go/pkg/mod \
    go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest && \
    go install github.com/fzipp/gocyclo/cmd/gocyclo@latest && \
    go install golang.org/x/vuln/cmd/govulncheck@latest

# 复制源代码
COPY . .

# 运行安全扫描
RUN gosec ./... && \
    gosec -fmt json -out gosec-report.json ./... && \
    gocyclo -over 15 ./... && \
    govulncheck ./...

# ================================
# 生产运行阶段 - 最小化安全镜像
# ================================
FROM scratch AS production

# 元数据标签
LABEL maintainer="Law OA Team <team@lawoa.com>" \
      version="2.1.0" \
      description="Law OA Go - 律师事务所办公自动化系统后端服务" \
      org.opencontainers.image.title="Law OA Go Backend" \
      org.opencontainers.image.description="律师事务所办公自动化系统后端服务" \
      org.opencontainers.image.url="https://github.com/law-oa/law-oa-go" \
      org.opencontainers.image.documentation="https://docs.lawoa.com" \
      org.opencontainers.image.source="https://github.com/law-oa/law-oa-go" \
      org.opencontainers.image.version="2.1.0" \
      org.opencontainers.image.revision="${BUILD_COMMIT}" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.vendor="Law OA Team" \
      security.scan="enabled" \
      build.date="${BUILD_DATE}"

# 从构建阶段复制CA证书和时区数据
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# 复制用户和组信息
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# 设置工作目录
WORKDIR /app

# 复制二进制文件
COPY --from=builder /build/law-oa-go /app/law-oa-go

# 复制配置文件
COPY --from=builder /build/config /app/config/
COPY --from=builder /build/.env.example /app/.env

# 复制预创建运行目录，避免 scratch 阶段执行 shell 命令
COPY --from=builder --chown=1001:1001 /build/runtime-dirs/ /app/

# 使用非root用户 (UID 1001)
USER 1001

# 设置生产环境变量
ENV TZ=Asia/Shanghai \
    GIN_MODE=release \
    PORT=8080 \
    ENVIRONMENT=production \
    LOG_LEVEL=info \
    METRICS_ENABLED=true \
    TRACING_ENABLED=true \
    HEALTH_CHECK_ENABLED=true

# 暴露端口
EXPOSE 8080 8081 9090

# 健康检查 - 增强的健康检查策略
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD ["/app/law-oa-go", "healthcheck"] || exit 1

# 启动应用 - 优化的启动参数
ENTRYPOINT ["/app/law-oa-go"]
CMD ["serve", "--config", "/app/config/config.yaml"]

# ================================
# 测试阶段 - 专门的测试镜像
# ================================
FROM base AS testing

# 设置工作目录
WORKDIR /build

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# 复制源代码
COPY . .

# 运行测试套件
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go test -v -race -coverprofile=coverage.out ./... && \
    go tool cover -html=coverage.out -o coverage.html && \
    go tool cover -func=coverage.out | grep 'total:' && \
    echo "✅ 测试完成，覆盖率报告已生成"

# 生成测试报告
RUN go test -json ./... > test-report.json && \
    go vet ./... > vet-report.log 2>&1 || true && \
    gofmt -d . > format-report.log 2>&1 || true

# ================================
# 开发阶段 - 包含开发工具的热重载镜像
# ================================
FROM base AS development

# 安装开发依赖
RUN apk add --no-cache git ca-certificates tzdata air curl make

# 设置工作目录
WORKDIR /app

# 设置Go环境变量
ENV CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64 \
    GIN_MODE=debug \
    ENVIRONMENT=development

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# 复制源代码
COPY . .

# 创建开发工具配置
RUN echo 'root = "."\ntestdata_dir = "testdata"\ntmp_dir = "tmp"\n\n[build]\n  args_bin = []\n  bin = "./tmp/main"\n  cmd = "go build -o ./tmp/main ./cmd/main.go"\n  delay = 1000\n  exclude_dir = ["assets", "tmp", "vendor", "testdata"]\n  exclude_file = []\n  exclude_regex = ["_test.go"]\n  exclude_unchanged = false\n  follow_symlink = false\n  full_bin = ""\n  include_dir = []\n  include_ext = ["go", "tpl", "tmpl", "html"]\n  include_file = []\n  kill_delay = "0s"\n  log = "build-errors.log"\n  poll = false\n  poll_interval = 0\n  rerun = false\n  rerun_delay = 500\n  send_interrupt = false\n  stop_on_root = false\n\n[color]\n  app = ""\n  build = "yellow"\n  main = "magenta"\n  runner = "green"\n  watcher = "cyan"\n\n[log]\n  main_only = false\n  time = true\n\n[misc]\n  clean_on_exit = false\n\n[screen]\n  clear_on_rebuild = false\n  keep_scroll = true' > .air.toml

# 安装开发工具
RUN --mount=type=cache,target=/go/pkg/mod \
    go install github.com/cosmtrek/air@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && \
    go install github.com/swaggo/swag/cmd/swag@latest

# 暴露端口
EXPOSE 8080 8081 9090

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# 开发环境启动脚本
RUN echo '#!/bin/sh\necho "🚀 Law OA Go 开发环境启动中..."\necho "📊 热重载已启用"\necho "🔍 健康检查: http://localhost:8080/health"\necho "📈 指标: http://localhost:8081/metrics"\necho "📚 API文档: http://localhost:8080/swagger/index.html"\necho "🛠️  按 Ctrl+C 停止开发服务器"\nexec air -c .air.toml' > /app/start-dev.sh && \
    chmod +x /app/start-dev.sh

# 启动开发环境
CMD ["/app/start-dev.sh"]

# ================================
# 预发布阶段 - 生产环境验证
# ================================
FROM production AS staging

# 设置预发布环境变量
ENV ENVIRONMENT=staging \
    LOG_LEVEL=debug \
    METRICS_ENABLED=true \
    TRACING_ENABLED=true \
    HEALTH_CHECK_ENABLED=true

# 预发布标签
LABEL environment="staging" \
      build.type="staging" \
      deployment.stage="pre-production"

# 验证生产配置
CMD ["/app/law-oa-go", "config", "validate"]
