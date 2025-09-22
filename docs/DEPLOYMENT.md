# 法律事务所自动化系统 - 部署指南

## 概述

Law OA Go 是基于 Go 语言开发的法律事务所自动化系统，采用单体架构设计。本指南提供基础的部署流程和配置说明。

## 部署架构

### 系统架构图

```
┌─────────────────────────────────────────────────────────────┐
│                    Web服务器层                              │
│              ┌─────────────┐                                │
│              │   Nginx     │                                │
│              │  Web Server │                                │
│              └─────────────┘                                │
├─────────────────────────────────────────────────────────────┤
│                    应用层                                    │
│              ┌─────────────┐                                │
│              │   Law OA    │                                │
│              │      Go     │                                │
│              └─────────────┘                                │
├─────────────────────────────────────────────────────────────┤
│                    数据层                                    │
│  ┌─────────────┐ ┌─────────────┐                          │
│  │   MySQL     │ │   Redis     │                          │
│  │    Database │ │   Cache     │                          │
│  └─────────────┘ └─────────────┘                          │
└─────────────────────────────────────────────────────────────┘
```

### 部署策略

1. **单机部署**: 简单直接的单体应用部署
2. **手动部署**: 基础的部署脚本和配置管理
3. **基础监控**: 应用状态监控和日志记录
4. **数据备份**: 基本的数据库备份策略

## 快速开始

### 1. 环境要求

#### 系统要求
- **操作系统**: Linux (Ubuntu 20.04+ / CentOS 8+)
- **CPU**: 最少 2 核心
- **内存**: 最少 2GB
- **磁盘**: 最少 20GB

#### 软件依赖
- **Go**: 1.23+ (开发环境)
- **Docker**: 20.10+ (可选)
- **MySQL**: 8.0+ (生产环境)
- **Redis**: 6.0+ (可选)
- **Nginx**: 1.20+ (可选)

#### 网络要求
- **端口开放**: 8080 (应用端口)
- **防火墙配置**: 允许必要端口通信

### 2. 克隆项目

```bash
git clone <repository-url>
cd law-oa-go
```

### 3. 配置环境变量

```bash
# 复制环境变量模板（如果存在）
cp .env.example .env 2>/dev/null || echo "Creating .env file..."

# 编辑配置文件
vim .env
```

### 4. 初始化数据库

```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE law_oa CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 运行数据库迁移（如果有迁移文件）
go run cmd/migrate/main.go
```

### 5. 启动应用

```bash
# 编译应用
go build -o law-oa-go cmd/server/main.go

# 运行应用
./law-oa-go
```

### 6. 验证部署

```bash
# 检查应用健康状态
curl http://localhost:8080/health
```

### 7. 访问服务

- **API服务**: http://localhost:8080
- **健康检查**: http://localhost:8080/health

## 生产环境部署

### 1. 部署前准备

#### 系统初始化

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装必要工具
sudo apt install -y curl wget git vim

# 创建应用用户
sudo useradd -m -s /bin/bash appuser
sudo usermod -aG sudo appuser

# 创建部署目录
sudo mkdir -p /opt/law-oa-go/{bin,config,logs,backups}
sudo chown -R appuser:appuser /opt/law-oa-go
```

#### 数据库准备

```bash
# 安装MySQL
sudo apt install -y mysql-server

# 创建数据库和用户
sudo mysql -e "CREATE DATABASE law_oa CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
sudo mysql -e "CREATE USER 'law_oa_user'@'localhost' IDENTIFIED BY 'secure_password';"
sudo mysql -e "GRANT ALL PRIVILEGES ON law_oa.* TO 'law_oa_user'@'localhost';"
sudo mysql -e "FLUSH PRIVILEGES;"

# 重启MySQL
sudo systemctl restart mysql
```

#### Redis准备（可选）

```bash
# 安装Redis
sudo apt install -y redis-server

# 重启Redis
sudo systemctl restart redis
```

### 2. 手动部署

```bash
# 克隆代码
git clone <repository-url> /opt/law-oa-go/source
cd /opt/law-oa-go/source

# 配置环境
cp .env.production .env
vim .env

# 编译应用
go build -o /opt/law-oa-go/bin/law-oa-go cmd/server/main.go

# 创建服务配置
sudo tee /etc/systemd/system/law-oa-go.service > /dev/null <<EOF
[Unit]
Description=Law OA Go Service
After=network.target mysql.service

[Service]
Type=simple
User=appuser
WorkingDirectory=/opt/law-oa-go
ExecStart=/opt/law-oa-go/bin/law-oa-go
Restart=always
RestartSec=5
Environment=ENVIRONMENT=production

[Install]
WantedBy=multi-user.target
EOF

# 启动服务
sudo systemctl daemon-reload
sudo systemctl enable law-oa-go
sudo systemctl start law-oa-go
```

### 3. Docker部署方式（可选）

#### Docker Compose部署

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    image: law-oa-go:latest
    ports:
      - "8080:8080"
    environment:
      - ENVIRONMENT=production
      - DB_HOST=mysql
      - DB_PASSWORD=${DB_PASSWORD}
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      - mysql
    volumes:
      - ./config:/app/config
      - ./logs:/app/logs
    restart: unless-stopped

  mysql:
    image: mysql:8.0
    environment:
      - MYSQL_DATABASE=law_oa
      - MYSQL_ROOT_PASSWORD=${DB_PASSWORD}
    volumes:
      - mysql_data:/var/lib/mysql
    ports:
      - "3306:3306"
    restart: unless-stopped

  redis:
    image: redis:6-alpine
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    restart: unless-stopped

volumes:
  mysql_data:
  redis_data:
```

### 4. Nginx配置（可选）

#### 简单反向代理配置

```nginx
# /etc/nginx/sites-available/law-oa-go
server {
    listen 80;
    server_name api.lawfirm.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /health {
        proxy_pass http://127.0.0.1:8080/health;
        access_log off;
    }
}
```

## 配置说明

### 环境变量配置

```bash
# 服务器配置
ENVIRONMENT=production
PORT=8080
LOG_LEVEL=info

# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_secure_password
DB_NAME=law_oa

# JWT 配置
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRE=3600

# Redis 配置（可选）
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

### 配置文件配置

项目支持 YAML 配置文件 `config.yaml`：

```yaml
# 生产环境配置
environment: production
port: "8080"
log_level: info

# 数据库配置
database:
  host: localhost
  port: "3306"
  username: root
  password: your_secure_password
  database: law_oa
  charset: utf8mb4
  parse_time: true
  loc: Local

# JWT 配置
jwt:
  secret: your-secret-key-change-in-production
  expire: 3600
```

## API 文档

### 认证接口

#### 登录
- **URL**: `POST /api/v1/auth/login`
- **请求体**:
  ```json
  {
    "username": "admin",
    "password": "password"
  }
  ```
- **响应**:
  ```json
  {
    "success": true,
    "data": {
      "token": "jwt-token",
      "user": {
        "id": 1,
        "username": "admin",
        "email": "admin@lawfirm.com"
      }
    },
    "meta": {
      "timestamp": "2024-01-01T00:00:00Z",
      "version": "v1",
      "server": "law-oa-go"
    }
  }
  ```

### 案件管理接口

#### 获取案件列表
- **URL**: `GET /api/v1/cases`
- **查询参数**:
  - `page`: 页码 (默认: 1)
  - `size`: 每页大小 (默认: 10)
- **响应**:
  ```json
  {
    "success": true,
    "data": {
      "cases": [...],
      "total": 100,
      "page": 1,
      "size": 10
    },
    "meta": {
      "timestamp": "2024-01-01T00:00:00Z",
      "version": "v1",
      "server": "law-oa-go"
    }
  }
  ```

#### 创建案件
- **URL**: `POST /api/v1/cases`
- **请求体**:
  ```json
  {
    "case_no": "CASE2024001",
    "case_name": "案件名称",
    "case_type": "CIVIL",
    "project_type": "CASE",
    "principal_info": "委托人信息",
    "opponent_info": "对方当事人信息",
    "cause_of_action": "案由",
    "description": "案件描述",
    "contract_amount": 50000.00,
    "billing_method": "FIXED"
  }
  ```

## 监控和日志

### 应用监控

#### 健康检查端点

```bash
# 基础健康检查
curl http://localhost:8080/health
```

### 日志管理

#### 简单日志配置

```bash
# 应用日志
/var/log/law-oa-go/app.log

# 系统日志
journalctl -u law-oa-go -f
```

#### 日志轮转配置

```bash
# /etc/logrotate.d/law-oa-go
/var/log/law-oa-go/*.log {
    daily
    missingok
    rotate 7
    compress
    notifempty
    create 0644 appuser appuser
}
```

## 性能优化

### 缓存策略（可选）

- 使用 Redis 缓存热点数据
- 优化数据库查询

### 数据库优化

- 使用 GORM 进行数据库操作
- 支持数据库连接池
- 自动数据库迁移

## 安全考虑

### 认证和授权

- JWT 令牌认证
- 角色权限控制
- 令牌过期管理

### 数据安全

- 密码加密存储
- SQL 注入防护
- XSS 防护

### 网络安全

- HTTPS 加密传输
- CORS 跨域配置
- 请求频率限制

## 备份和恢复

### 数据库备份

#### 简单备份脚本

```bash
#!/bin/bash
# /scripts/backup-database.sh

DB_HOST="localhost"
DB_PORT="3306"
DB_NAME="law_oa"
DB_USER="root"
BACKUP_DIR="/backup/database"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/law_oa_${DATE}.sql"
RETENTION_DAYS=7

# 创建备份目录
mkdir -p ${BACKUP_DIR}

# 创建数据库备份
echo "Creating database backup..."
mysqldump -h ${DB_HOST} -P ${DB_PORT} -u ${DB_USER} -p ${DB_NAME} > ${BACKUP_FILE}

# 清理旧备份
find ${BACKUP_DIR} -name "law_oa_*.sql" -mtime +${RETENTION_DAYS} -delete

echo "Backup completed: ${BACKUP_FILE}"
```

### 恢复流程

#### 数据库恢复

```bash
#!/bin/bash
# /scripts/restore-database.sh

BACKUP_FILE=$1
DB_HOST="localhost"
DB_PORT="3306"
DB_NAME="law_oa"
DB_USER="root"

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup_file>"
    exit 1
fi

# 停止应用服务
systemctl stop law-oa-go

# 恢复数据库
echo "Restoring database..."
mysql -h ${DB_HOST} -P ${DB_PORT} -u ${DB_USER} -p ${DB_NAME} < ${BACKUP_FILE}

# 启动应用服务
systemctl start law-oa-go

echo "Database restore completed"
```

## 安全考虑

### 系统安全

#### 防火墙配置

```bash
# 安装ufw
sudo apt install ufw

# 配置防火墙规则
sudo ufw default deny incoming
sudo ufw default allow outgoing

# 允许必要端口
sudo ufw allow ssh
sudo ufw allow 8080

# 启用防火墙
sudo ufw enable
```

#### 系统加固

```bash
# 系统更新
sudo apt update && sudo apt upgrade -y

# 配置SSH安全
sudo vim /etc/ssh/sshd_config
# 修改: PermitRootLogin no
# 修改: PasswordAuthentication no

# 重启SSH服务
sudo systemctl restart sshd
```

### 应用安全

#### 环境变量安全

```bash
# 使用环境变量文件
export DB_PASSWORD="your_secure_password"
export JWT_SECRET="your_very_secure_jwt_secret"
```

## 故障排除

### 常见问题诊断

#### 1. 应用启动失败

```bash
# 检查应用状态
systemctl status law-oa-go

# 查看应用日志
journalctl -u law-oa-go -f

# 检查依赖服务
systemctl status mysql
```

#### 2. 数据库连接问题

```bash
# 测试数据库连接
mysql -h localhost -u root -p -e "SELECT 1;"

# 检查数据库状态
sudo systemctl status mysql
```

#### 3. 性能问题

```bash
# 检查系统资源
top
free -h
df -h
```

### 日志分析

```bash
# 查看应用日志
journalctl -u law-oa-go -f

# 查看错误日志
journalctl -u law-oa-go --since "1 hour ago" | grep ERROR
```

## 技术栈

### 核心技术

- **语言**: Go 1.23+
- **Web框架**: Gin
- **数据库**: MySQL 8.0+
- **缓存**: Redis 6.0+ (可选)
- **ORM**: GORM

### 部署和运维

- **容器化**: Docker (可选)
- **Web服务器**: Nginx (可选)

### 安全

- **认证**: JWT
- **授权**: RBAC
- **加密**: TLS 1.3

---

**注意**: 本文档为简化版部署指南，适用于单体应用架构。