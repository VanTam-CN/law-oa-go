# Law OA Go 企业级部署指南

> **生产入口说明（2026-07-19）**：当前生产数据库入口是 PostgreSQL
> schema bootstrap。Kubernetes 请只使用 `k8s/` 下按目录拆分的 canonical
> manifests，并先阅读 [`k8s/README.md`](../k8s/README.md)。仓库根目录的
> `k8s/deployment.yaml` 仅保留为弃用标记，不再创建旧版应用 Deployment；
> 不要依据本文件早期的 MySQL、`law-oa-app` 或 `kubectl apply -f k8s/`
> 示例部署生产环境。

<div align="center">

![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![Kubernetes](https://img.shields.io/badge/Kubernetes-Ready-326CE5?style=for-the-badge&logo=kubernetes&logoColor=white)
![Kubernetes manifests](https://img.shields.io/badge/Kubernetes-manifests-326CE5?style=for-the-badge&logo=kubernetes&logoColor=white)
![Production](https://img.shields.io/badge/Production-Grade-green.svg?style=for-the-badge)

**生产环境部署最佳实践**

 [环境准备](#-环境准备) • [Docker部署](#-docker部署) • [Kubernetes部署](#-kubernetes部署) • [监控配置](#-监控配置) • [安全配置](#-安全配置)

</div>

---

## 📋 目录

- [1. 环境准备](#1-环境准备)
- [2. Docker部署](#2-docker部署)
- [3. Kubernetes部署](#3-kubernetes部署)
- [4. Helm部署](#4-helm部署)
- [5. 监控配置](#5-监控配置)
- [6. 安全配置](#6-安全配置)
- [7. 性能优化](#7-性能优化)
- [8. 备份策略](#8-备份策略)
- [9. 故障排除](#9-故障排除)

---

## 🏗️ 部署架构

### 企业级系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                     负载均衡层                               │
│              ┌─────────────┐                                │
│              │   Ingress   │                                │
│              │  Controller │                                │
│              └─────────────┘                                │
├─────────────────────────────────────────────────────────────┤
│                     应用层                                   │
│              ┌─────────────┐                                │
│              │   Law OA    │     ┌─────────────┐            │
│              │      Go     │     │  Frontend   │            │
│              │    Backend  │     │   React     │            │
│              └─────────────┘     └─────────────┘            │
├─────────────────────────────────────────────────────────────┤
│                     数据层                                   │
│  ┌─────────────┐ ┌─────────────┐  ┌─────────────────┐      │
│  │   MySQL     │ │   Redis     │  │   Optional      │      │
│  │   Primary   │ │   Cache     │  │  Search Service  │      │
│  └─────────────┘ └─────────────┘  └─────────────────┘      │
├─────────────────────────────────────────────────────────────┤
│                    监控和日志层                               │
│  ┌─────────────┐ ┌─────────────┐  ┌─────────────────┐      │
│  │ Prometheus  │ │   Grafana   │  │     Jaeger      │      │
│  │  Metrics    │ │  Dashboard  │  │    Tracing      │      │
│  └─────────────┘ └─────────────┘  └─────────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

### 部署策略

1. **容器化部署**: 基于Docker的多阶段构建优化
2. **编排部署**: Kubernetes集群管理和自动扩缩容
3. **包管理**: Helm Charts版本控制和配置管理
4. **CI/CD流水线**: GitHub Actions自动化构建和部署
5. **监控告警**: Prometheus + Grafana 全链路监控，默认通过 `observability` profile 启用
6. **日志收集**: 结构化日志和集中式日志管理
7. **安全加固**: 网络策略、RBAC、Pod安全策略

---

## 1. 环境准备

### 1.1 系统要求

#### 最低配置
- **CPU**: 2核心
- **内存**: 4GB RAM
- **存储**: 50GB SSD
- **网络**: 100Mbps

#### 推荐配置
- **CPU**: 4核心以上
- **内存**: 8GB RAM以上
- **存储**: 200GB SSD以上
- **网络**: 1Gbps以上

#### 生产环境配置
- **CPU**: 8核心以上
- **内存**: 16GB RAM以上
- **存储**: 500GB NVMe SSD以上
- **网络**: 10Gbps以上

### 1.2 依赖软件

#### 必需软件
```bash
# Docker 20.10+
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Docker Compose v2.0+
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Git 2.30+
sudo apt-get update
sudo apt-get install git -y
```

#### Kubernetes环境（可选）
```bash
# kubectl 1.29+
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Helm 3.14+
curl https://get.helm.sh/helm-v3.14.0-linux-amd64.tar.gz | tar xz
sudo mv linux-amd64/helm /usr/local/bin/
```

### 1.3 网络和防火墙

#### 端口要求
```bash
# 应用端口
8080/tcp  # Go API服务
3003/tcp  # React前端

# 数据库端口
3306/tcp  # MySQL
5432/tcp  # PostgreSQL
6379/tcp  # Redis
# 监控端口
9090/tcp  # Prometheus（observability profile）
3000/tcp  # Grafana（observability profile）
14268/tcp # Jaeger（observability profile）
```

#### 防火墙配置
```bash
# Ubuntu/Debian
sudo ufw allow 8080/tcp
sudo ufw allow 3003/tcp
sudo ufw allow from 10.0.0.0/8 to any port 3306
sudo ufw allow from 10.0.0.0/8 to any port 5432
sudo ufw allow from 10.0.0.0/8 to any port 6379

# CentOS/RHEL
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --permanent --add-port=3003/tcp
sudo firewall-cmd --reload
```

---

## 2. Docker部署

### 2.1 快速部署

#### 使用Docker Compose
```bash
# 1. 克隆项目
git clone <repository-url>
cd law-oa-go

# 2. 配置环境变量
cp .env.example .env
vim .env

# 3. 启动所有服务
docker compose up -d

# 4. 验证部署
docker compose ps
curl http://localhost:8080/health
```

#### 企业级Docker Compose配置
```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
      target: production
    ports:
      - "8080:8080"
    environment:
      - GIN_MODE=release
      - DB_HOST=mysql
      - REDIS_HOST=redis
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: ${DB_ROOT_PASSWORD}
      MYSQL_DATABASE: ${DB_NAME}
      MYSQL_USER: ${DB_USER}
      MYSQL_PASSWORD: ${DB_PASSWORD}
    volumes:
      - mysql_data:/var/lib/mysql
    ports:
      - "3306:3306"
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  prometheus:
    image: prom/prometheus:latest
    profiles: [observability]
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    profiles: [observability]
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana_data:/var/lib/grafana
    restart: unless-stopped

volumes:
  mysql_data:
  redis_data:
  prometheus_data:
  grafana_data:
```

### 2.2 多阶段构建优化

#### 生产环境Dockerfile特性
- **多阶段构建**: 减少镜像大小和提高安全性
- **非root用户**: 增强容器安全性
- **健康检查**: 内置应用健康状态监控
- **安全扫描**: 集成漏洞检测阶段

#### 构建命令
```bash
# 构建生产镜像
docker build -t law-oa-go:2.1.0 --target production .

# 构建开发镜像
docker build -t law-oa-go:dev --target development .

# 安全扫描
docker build -t law-oa-go:scan --target security-scanner .
```

### 2.3 容器管理

#### 常用命令
```bash
# 查看容器状态
docker compose ps

# 查看日志
docker compose logs -f app

# 进入容器
docker compose exec app sh

# 重启服务
docker compose restart app

# 更新服务
docker compose pull
docker compose up -d

# 清理资源
docker system prune -f
docker volume prune -f
```

#### 性能监控
```bash
# 查看资源使用
docker stats

# 查看容器详情
docker inspect law-oa-go_app_1

# 监控健康状态
curl http://localhost:8080/health
```

---

## 3. Kubernetes部署

### 3.1 准备工作

#### 创建命名空间
```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: law-oa
  labels:
    name: law-oa
    environment: production
```

#### 配置ConfigMap
```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: law-oa-config
  namespace: law-oa
data:
  config.yaml: |
    server:
      port: 8080
      mode: release
    database:
      host: mysql
      port: 3306
      name: law_oa
    redis:
      host: redis
      port: 6379
    logging:
      level: info
      format: json
```

#### 配置Secret
```yaml
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: law-oa-secret
  namespace: law-oa
type: Opaque
data:
  db-password: <base64-encoded-password>
  jwt-secret: <base64-encoded-jwt-secret>
  redis-password: <base64-encoded-redis-password>
```

### 3.2 应用部署

#### 企业级Deployment配置

OnlyOffice 默认关闭。只有在业务确实需要在线编辑时，才在 Deployment 或 Compose 中显式注入 `ONLYOFFICE_ENABLED=true`、`ONLYOFFICE_URL`、`ONLYOFFICE_SECRET` 和 `BACKEND_URL`；其中两个 URL 必须是纯 origin，不能带路径、查询串、片段或凭据。

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: law-oa-app
  namespace: law-oa
  labels:
    app: law-oa-app
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 25%
      maxSurge: 25%
  selector:
    matchLabels:
      app: law-oa-app
  template:
    metadata:
      labels:
        app: law-oa-app
    spec:
      containers:
      - name: law-oa-app
        image: law-registry.com/law-oa-go:2.1.0
        ports:
        - containerPort: 8080
        env:
        - name: GIN_MODE
          value: "release"
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: law-oa-secret
              key: db-password
        resources:
          requests:
            cpu: 500m
            memory: 1Gi
          limits:
            cpu: 2000m
            memory: 2Gi
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
        securityContext:
          runAsNonRoot: true
          runAsUser: 65534
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
```

#### Service配置
```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: law-oa-service
  namespace: law-oa
  labels:
    app: law-oa-service
spec:
  selector:
    app: law-oa-app
  ports:
  - name: http
    port: 8080
    targetPort: 8080
    protocol: TCP
  type: ClusterIP
```

#### Ingress配置
```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: law-oa-ingress
  namespace: law-oa
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/rate-limit: "100"
spec:
  tls:
  - hosts:
    - api.lawoa.com
    secretName: law-oa-tls
  rules:
  - host: api.lawoa.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: law-oa-service
            port:
              number: 8080
```

### 3.3 自动扩缩容

#### HPA配置
```yaml
# hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: law-oa-hpa
  namespace: law-oa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: law-oa-app
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
```

### 3.4 部署命令

```bash
# 按顺序应用生产清单；不要使用 kubectl apply -f k8s/ 触发历史清单
kubectl apply -f k8s/namespaces/law-oa.yaml
kubectl apply -f k8s/configmaps/law-oa-config.yaml
kubectl apply -f k8s/persistentvolumeclaims/uploads.yaml
kubectl apply -f k8s/services/backend.yaml
kubectl apply -f k8s/deployments/backend.yaml

# 查看部署状态
kubectl get pods -n law-oa
kubectl get services -n law-oa
kubectl get ingress -n law-oa

# 查看日志
kubectl logs -f deployment/law-oa-backend -n law-oa

# 扩容应用
kubectl scale deployment law-oa-backend --replicas=5 -n law-oa

# 更新应用
kubectl set image deployment/law-oa-backend backend=YOUR_REGISTRY/law-oa/backend:2.1.1 -n law-oa

# 回滚应用
kubectl rollout undo deployment/law-oa-backend -n law-oa
```

---

## 4. Helm部署

当前生产发布不使用 Helm。仓库中的 `helm/law-oa-go/` 仅保留为弃用兼容物，
其旧版 MySQL、默认凭据和依赖组合不代表当前生产支持矩阵。请使用
`k8s/README.md` 描述的 PostgreSQL canonical manifests；Secret 必须由
Secret Manager 或 External Secrets 创建，不能通过 Helm values 或仓库模板
写入真实值。

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
go build -o /opt/law-oa-go/bin/law-oa-go .

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

---

## 📞 技术支持

### 联系方式
- **技术支持邮箱**: support@lawoa.com
- **紧急故障热线**: +86-xxx-xxxx-xxxx
- **技术文档**: https://docs.lawoa.com
- **问题反馈**: https://github.com/law-oa-go/issues

### 服务级别协议（SLA）
- **P1-紧急故障**: 15分钟响应，4小时解决
- **P2-重要问题**: 30分钟响应，8小时解决
- **P3-一般问题**: 2小时响应，24小时解决
- **P4-功能请求**: 1个工作日响应，按优先级排期

### 扩展阅读
- **📖 企业级运维指南**: [DEPLOYMENT_ENTERPRISE.md](./DEPLOYMENT_ENTERPRISE.md)
  - 详细的监控配置（Prometheus + Grafana）
  - 完整的安全配置（网络策略、RBAC、Pod安全）
  - 性能优化策略（数据库、缓存、应用调优）
  - 备份和灾难恢复方案
  - CI/CD流水线配置
  - 故障排除和恢复流程

---

## 🛠️ 技术栈

### 核心技术
- **语言**: Go 1.23+
- **Web框架**: Gin
- **数据库**: MySQL 8.0+ / PostgreSQL 15+
- **缓存**: Redis 7.0+
- **搜索**: 数据库搜索回退（默认），Elasticsearch 仅作为可选扩展
- **ORM**: GORM v1.30

### 部署和运维
- **容器化**: Docker & Docker Compose
- **编排**: Kubernetes 1.29+
- **包管理**: Helm 3.14+
- **监控**: Prometheus + Grafana
- **日志**: 结构化日志 + 可选集中式日志
- **追踪**: Jaeger

### 安全
- **认证**: JWT (golang-jwt/v5)
- **授权**: RBAC
- **加密**: TLS 1.3
- **扫描**: Trivy 安全扫描

### CI/CD
- **版本控制**: Git + GitHub
- **流水线**: GitHub Actions
- **镜像仓库**: Harbor / Docker Hub
- **部署策略**: 蓝绿部署、金丝雀发布

---

<div align="center">

**Law OA Go 部署指南**

🚀 企业级律师事务所办公自动化系统

版本: v2.1.0 | 最后更新: 2025-10-17

</div>
