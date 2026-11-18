# React 前端 Dockerfile - Law OA Go
# 基于多阶段构建的现代化前端部署方案

# ================================
# 构建阶段 - 构建React应用
# ================================
FROM node:18-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache libc6-compat

# 复制package文件
COPY package*.json ./

# 设置npm镜像源（加速构建）
RUN npm config set registry https://registry.npmmirror.com

# 安装依赖
RUN npm ci --silent

# 复制源代码
COPY . .

# 设置构建环境变量
ARG REACT_APP_API_URL=/api
ARG REACT_APP_ENVIRONMENT=production
ARG VITE_ONLYOFFICE_ENABLED=false
ARG VITE_ONLYOFFICE_URL=
ARG GENERATE_SOURCEMAP=false

ENV REACT_APP_API_URL=${REACT_APP_API_URL}
ENV REACT_APP_ENVIRONMENT=${REACT_APP_ENVIRONMENT}
ENV VITE_ONLYOFFICE_ENABLED=${VITE_ONLYOFFICE_ENABLED}
ENV VITE_ONLYOFFICE_URL=${VITE_ONLYOFFICE_URL}
ENV GENERATE_SOURCEMAP=${GENERATE_SOURCEMAP}
ENV NODE_ENV=production
ENV NODE_OPTIONS=--max-old-space-size=4096

# 构建应用
RUN npm run build

# 验证构建结果
RUN ls -la /app/dist && \
    test -f /app/dist/index.html && \
    echo "✅ React应用构建成功"

# ================================
# 生产阶段 - Nginx服务器
# ================================
FROM nginx:1.24-alpine AS production

ARG BACKEND_UPSTREAM=backend:8080

# 元数据标签
LABEL maintainer="Law OA Team <team@lawoa.com>" \
      version="2.1.0" \
      description="Law OA Go - 律师事务所办公自动化系统前端服务" \
      org.opencontainers.image.title="Law OA Go Frontend" \
      org.opencontainers.image.description="律师事务所办公自动化系统前端服务" \
      org.opencontainers.image.url="https://github.com/law-oa/law-oa-go" \
      org.opencontainers.image.version="2.1.0" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.vendor="Law OA Team"

# 安装必要的工具。官方 Nginx 镜像已经提供非 root 的 nginx 用户。
RUN apk add --no-cache curl && rm -f /etc/nginx/conf.d/default.conf

# 创建自定义Nginx配置
RUN cat > /etc/nginx/nginx.conf << 'EOF'
user nginx;
worker_processes auto;
error_log /var/log/nginx/error.log warn;
pid /var/run/nginx.pid;

events {
    worker_connections 1024;
    use epoll;
    multi_accept on;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    # 日志格式
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for"';

    access_log /var/log/nginx/access.log main;

    # 性能优化
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    client_max_body_size 20M;

    # Gzip压缩
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/json
        application/javascript
        application/xml+rss
        application/atom+xml
        image/svg+xml;

    # 缓存配置
    map $sent_http_content_type $expires {
        default                    off;
        text/html                  epoch;
        text/css                   max;
        application/javascript     max;
        ~image/                    max;
        ~font/                     max;
    }

    server {
        listen 80;
        server_name localhost;
        root /usr/share/nginx/html;
        index index.html index.htm;

        # 安全头部
        add_header X-Frame-Options "SAMEORIGIN" always;
        add_header X-XSS-Protection "1; mode=block" always;
        add_header X-Content-Type-Options "nosniff" always;
        add_header Referrer-Policy "no-referrer-when-downgrade" always;
        add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;

        # 缓存配置
        expires $expires;

        # 静态资源处理
        location /static/ {
            expires 1y;
            add_header Cache-Control "public, immutable";
        }

        # API代理（可选）
        location /api/ {
            proxy_pass http://backend:8080;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_connect_timeout 30s;
            proxy_send_timeout 30s;
            proxy_read_timeout 30s;
        }

        # React Router支持
        location / {
            try_files $uri $uri/ /index.html;
            add_header Cache-Control "no-cache, no-store, must-revalidate";
            add_header Pragma "no-cache";
            add_header Expires "0";
        }

        # 健康检查端点
        location /health {
            access_log off;
            return 200 "healthy\n";
            add_header Content-Type text/plain;
        }

        # 错误页面
        error_page 404 /index.html;
        error_page 500 502 503 504 /50x.html;
        location = /50x.html {
            root /usr/share/nginx/html;
        }
    }
}
EOF

# Compose uses the backend service name; Kubernetes injects its canonical
# Service name as a build argument. The frontend remains same-origin for the
# browser while Nginx forwards /api requests internally.
RUN sed -i "s#proxy_pass http://backend:8080;#proxy_pass http://${BACKEND_UPSTREAM};#" /etc/nginx/nginx.conf

# 复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 设置正确的权限
RUN chown -R nginx:nginx /usr/share/nginx/html && \
    chown -R nginx:nginx /var/cache/nginx && \
    chown -R nginx:nginx /var/log/nginx && \
    chown -R nginx:nginx /etc/nginx/conf.d && \
    touch /var/run/nginx.pid && \
    chown -R nginx:nginx /var/run/nginx.pid

# 切换到非root用户
USER nginx

# 暴露端口
EXPOSE 80

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:80 || exit 1

# 设置环境变量
ENV TZ=Asia/Shanghai \
    NGINX_WORKER_PROCESSES=auto \
    NGINX_WORKER_CONNECTIONS=1024

# 启动nginx
CMD ["nginx", "-g", "daemon off;"]
