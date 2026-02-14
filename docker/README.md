# OnlyOffice Document Server 部署说明

## 概述

本配置用于部署 OnlyOffice Document Server，为 Law OA Go 系统提供在线文档编辑功能。

## 快速开始

### 1. 启动服务

```bash
# 创建数据目录
mkdir -p data/onlyoffice/{db,redis,rabbitmq,log,data,fonts,cache}

# 启动 OnlyOffice 服务
docker-compose -f docker/docker-compose.document.yml up -d

# 查看服务状态
docker-compose -f docker/docker-compose.document.yml ps

# 查看日志
docker-compose -f docker/docker-compose.document.yml logs -f onlyoffice-documentserver
```

### 2. 验证服务

服务启动后，访问健康检查端点：

```bash
curl http://localhost:9090/healthcheck
```

返回 `{"status":"ok"}` 表示服务正常。

## 配置说明

### 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `ONLYOFFICE_PORT` | 9090 | OnlyOffice 服务端口 |
| `ONLYOFFICE_DB_PORT` | 55432 | PostgreSQL 数据库端口 |
| `ONLYOFFICE_REDIS_PORT` | 6380 | Redis 端口 |
| `ONLYOFFICE_JWT_SECRET` | your-onlyoffice-jwt-secret-change-me | JWT 密钥（生产环境必须修改） |

### 服务端口

| 服务 | 容器端口 | 主机端口 | 说明 |
|------|----------|----------|------|
| OnlyOffice Document Server | 80 | 9090 | 文档编辑服务 |
| PostgreSQL | 5432 | 55432 | OnlyOffice 数据库 |
| Redis | 6379 | 6380 | 缓存服务 |
| RabbitMQ | 5672/15672 | 15772/16772 | 消息队列/管理界面 |

## 数据持久化

所有数据存储在 `./data/onlyoffice/` 目录下：

```
data/onlyoffice/
├── db/          # PostgreSQL 数据
├── redis/       # Redis 数据
├── rabbitmq/    # RabbitMQ 数据
├── log/         # 日志文件
├── data/        # 文档数据
├── fonts/       # 自定义字体
└── cache/       # 缓存文件
```

## 生产环境配置

### 1. JWT 密钥配置

在生产环境中，必须配置强密钥：

```bash
export ONLYOFFICE_JWT_SECRET=$(openssl rand -base64 32)
docker-compose -f docker/docker-compose.document.yml up -d
```

### 2. HTTPS 配置

生产环境建议使用 Nginx 反向代理配置 HTTPS：

```nginx
server {
    listen 443 ssl;
    server_name office.yourdomain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:9090;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket 支持
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

### 3. 资源限制

根据实际使用情况调整 `docker-compose.document.yml` 中的资源限制：

- **小型团队**（<50人）：CPU 1.0，内存 1G
- **中型团队**（50-200人）：CPU 2.0，内存 2G
- **大型团队**（>200人）：考虑部署多实例负载均衡

## 维护操作

### 查看日志

```bash
# OnlyOffice 服务日志
docker-compose -f docker/docker-compose.document.yml logs -f onlyoffice-documentserver

# 数据库日志
docker-compose -f docker/docker-compose.document.yml logs -f onlyoffice-db

# Redis 日志
docker-compose -f docker/docker-compose.document.yml logs -f onlyoffice-redis
```

### 备份数据

```bash
# 备份 PostgreSQL
docker exec law-oa-onlyoffice-db pg_dump -U onlyoffice onlyoffice > backup_$(date +%Y%m%d).sql

# 备份整个数据目录
tar -czf onlyoffice_backup_$(date +%Y%m%d).tar.gz data/onlyoffice/
```

### 清理缓存

```bash
# 进入容器
docker exec -it law-oa-onlyoffice bash

# 清理字体缓存
rm -rf /var/lib/onlyoffice/documentserver/cache/*

# 重启服务
docker-compose -f docker/docker-compose.document.yml restart onlyoffice-documentserver
```

### 更新服务

```bash
# 拉取最新镜像
docker-compose -f docker/docker-compose.document.yml pull

# 重新创建容器
docker-compose -f docker/docker-compose.document.yml up -d --force-recreate
```

## 故障排查

### 服务启动失败

1. 检查端口是否被占用：
```bash
lsof -i :9090
lsof -i :55432
lsof -i :6380
```

2. 检查数据目录权限：
```bash
chmod -R 755 data/onlyoffice/
```

3. 查看详细日志：
```bash
docker-compose -f docker/docker-compose.document.yml logs --tail=100
```

### 文档无法打开

1. 验证 JWT 密钥配置
2. 检查后端回调 URL 是否可访问
3. 确认文档格式受支持

### 性能问题

1. 增加容器资源限制
2. 清理缓存
3. 考虑使用独立的 Redis 实例

## 支持的文档格式

- **文本文档**：DOC, DOCX, ODT, RTF, TXT, PDF, HTML, EPUB
- **电子表格**：XLS, XLSX, ODS, CSV
- **演示文稿**：PPT, PPTX, ODP

## 参考链接

- [OnlyOffice Document Server 官方文档](https://api.onlyoffice.com/docserver/)
- [Docker 部署指南](https://helpcenter.onlyoffice.com/installation/docker-document-server-install.aspx)
- [API 文档](https://api.onlyoffice.com/editors/basic/)
