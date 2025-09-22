# 多阶段构建
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的工具
RUN apk add --no-cache git ca-certificates tzdata

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建参数
ARG BUILD_COMMIT=unknown
ARG BUILD_DATE=unknown
ARG BUILD_TARGET=standard
ARG PGO_PROFILE=default.pgo

# 根据构建目标选择构建方式
RUN if [ "$BUILD_TARGET" = "pgo" ] && [ -f "$PGO_PROFILE" ]; then \
        echo "使用PGO优化构建..." && \
        go build -pgo=$PGO_PROFILE -ldflags="-s -w -X main.Version=${BUILD_COMMIT} -X main.BuildTime=${BUILD_DATE}" -o law-oa .; \
    else \
        echo "使用标准构建..." && \
        go build -ldflags="-s -w -X main.Version=${BUILD_COMMIT} -X main.BuildTime=${BUILD_DATE}" -o law-oa .; \
    fi

# 验证构建
RUN ./law-oa --version

# 运行阶段
FROM alpine:latest

# 安装 ca-certificates 和 timezone 数据
RUN apk --no-cache add ca-certificates tzdata curl

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN addgroup -g 1000 appgroup && adduser -u 1000 -G appgroup -s /bin/sh -D appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/law-oa .

# 复制配置文件和目录
COPY --from=builder /app/config ./config
COPY --from=builder /app/.env.example .env

# 创建必要目录
RUN mkdir -p uploads/contract uploads/evidence uploads/letter uploads/other logs && \
    chown -R appuser:appgroup /app

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# 启动应用
CMD ["./law-oa"]