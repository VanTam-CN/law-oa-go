# Law OA Go 操作手册

**版本**: v2.4.0
**更新日期**: 2026-04-07
**适用环境**: Docker Compose 部署

---

## 目录

1. [系统概述](#1-系统概述)
2. [环境准备](#2-环境准备)
3. [快速启动](#3-快速启动)
4. [服务管理](#4-服务管理)
5. [环境变量配置](#5-环境变量配置)
6. [数据库管理](#6-数据库管理)
7. [监控与可观测性](#7-监控与可观测性)
8. [API 使用指南](#8-api-使用指南)
9. [安全配置](#9-安全配置)
10. [故障排除](#10-故障排除)
11. [备份与恢复](#11-备份与恢复)
12. [开发环境](#12-开发环境)

---

## 1. 系统概述

### 1.1 架构概览

Law OA Go 采用单体架构，Go 后端 + React 前端，通过 Docker Compose 编排 8 个服务：

```
┌─────────────┐    ┌──────────────┐    ┌────────────────┐
│  Frontend   │───▶│   Backend    │───▶│  MySQL 8.0     │
│  React 18   │    │  Go 1.23+    │    │  (主数据库)     │
│  :3003      │    │  :8080       │    │  :33060        │
└─────────────┘    └──────┬───────┘    └────────────────┘
                          │
                   ┌──────┼───────┐
                   ▼      ▼       ▼
              ┌────────┐ ┌─────┐ ┌──────────────┐
              │ Redis  │ │ ES  │ │ OnlyOffice   │
              │ :6379  │ │8.11 │ │ (文档编辑)    │
              └────────┘ │9200 │ └──────────────┘
                         └─────┘
                   ┌──────┼───────┐
                   ▼      ▼       ▼
              ┌────────┐ ┌─────┐ ┌────────┐
              │Prometheus│ │Jaeger│ │Grafana │
              │ :9090   │ │16686│ │ :3000  │
              └────────┘ └─────┘ └────────┘
```

### 1.2 服务清单

| 服务 | 容器名 | 端口 | 说明 |
|------|--------|------|------|
| backend | law-oa-backend | 8080 | Go 后端 API 服务 |
| frontend | law-oa-frontend | 3003 | React 前端应用 |
| mysql | law-oa-mysql | 33060 | MySQL 8.0 数据库 |
| redis | law-oa-redis | 6379 | Redis 7 缓存 |
| elasticsearch | law-oa-elasticsearch | 9200/9300 | ES 8.11 全文检索 |
| kibana | law-oa-kibana | 5601 | 日志分析平台 |
| jaeger | law-oa-jaeger | 16686 | 分布式追踪 |
| prometheus | law-oa-prometheus | 9090 | 指标监控 |
| grafana | law-oa-grafana | 3000 | 可视化仪表盘 |

### 1.3 资源需求

| 服务 | CPU 限制 | 内存限制 | CPU 预留 | 内存预留 |
|------|----------|----------|----------|----------|
| backend | 2.0 | 1G | 0.5 | 256M |
| frontend | 1.0 | 512M | 0.25 | 128M |
| mysql | 2.0 | 2G | 0.5 | 512M |
| redis | 1.0 | 512M | 0.25 | 128M |
| elasticsearch | 2.0 | 1G | 0.5 | 512M |
| 其他 | 1.0 | 512M~1G | 0.25 | 128M~256M |

**最低服务器配置**: 4核 CPU / 8G 内存 / 50G SSD

---

## 2. 环境准备

### 2.1 依赖安装

```bash
# Docker (>= 24.0)
docker --version

# Docker Compose V2 (>= 2.20)
docker compose version

# Git
git --version
```

### 2.2 克隆项目

```bash
git clone <repository-url> law-oa-go
cd law-oa-go
```

### 2.3 创建数据目录

```bash
mkdir -p ./data/{mysql,redis,elasticsearch,prometheus,grafana,uploads,logs}
```

---

## 3. 快速启动

### 3.1 配置环境变量

```bash
# 复制环境变量模板
cp .env.example .env.local

# 编辑关键配置（至少修改以下项）
vi .env.local
```

**必须修改的配置项**:

```bash
# JWT 密钥（至少32字符，生产环境必须更换）
JWT_SECRET=your-production-jwt-secret-at-least-32-chars

# CORS 允许的域名（替换为实际域名）
CORS_ALLOWED_ORIGINS=https://your-domain.com

# 数据库密码
DB_PASSWORD=your-strong-db-password
MYSQL_ROOT_PASSWORD=your-strong-root-password

# Redis 密码（建议设置）
REDIS_PASSWORD=your-redis-password
```

### 3.2 启动所有服务

```bash
# 加载环境变量并启动
set -a; source .env.local; set +a
docker compose up -d

# 查看启动状态
docker compose ps
```

### 3.3 等待服务就绪

所有服务都配置了健康检查，可通过以下命令确认：

```bash
# 查看健康状态
docker compose ps --format "table {{.Name}}\t{{.Status}}"

# 预期输出：所有服务应显示 "healthy" 或 "running"
```

### 3.4 执行数据库迁移

```bash
# 进入后端容器执行迁移
docker compose exec backend ./law-oa-go migrate

# 或使用 make 命令（需本地安装 migrate 工具）
make migrate-up
```

### 3.5 创建初始管理员

```bash
# 调用注册接口创建管理员
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "your-admin-password",
    "email": "admin@lawoa.com",
    "role": "admin"
  }'
```

### 3.6 验证部署

```bash
# 健康检查
curl http://localhost:8080/api/v1/health

# 预期响应
# {"service":"law-oa","status":"ok"}
```

---

## 4. 服务管理

### 4.1 启动/停止/重启

```bash
# 启动所有服务（后台运行）
docker compose up -d

# 停止所有服务
docker compose down

# 重启所有服务
docker compose restart

# 重启单个服务
docker compose restart backend

# 停止并清除数据卷（危险操作，会删除数据）
docker compose down -v
```

### 4.2 查看状态

```bash
# 查看所有服务状态
docker compose ps

# 查看资源使用
docker stats --no-stream

# 查看后端日志
docker compose logs -f backend --tail=100

# 查看所有服务日志
docker compose logs -f --tail=50
```

### 4.3 更新部署

```bash
# 拉取最新代码
git pull origin main

# 重新构建并启动
docker compose up -d --build

# 仅重建后端
docker compose up -d --build backend
```

### 4.4 版本回退

```bash
# 查看历史版本
git tag -l

# 回退到指定版本
git checkout v2.3.0
docker compose up -d --build
```

---

## 5. 环境变量配置

### 5.1 核心配置

| 变量 | 默认值 | 说明 | 生产建议 |
|------|--------|------|----------|
| `ENVIRONMENT` | development | 运行环境 | `production` |
| `DEBUG` | true | 调试模式 | `false` |
| `PORT` | 8080 | 后端端口 | 保持 8080 |
| `GIN_MODE` | release | Gin 模式 | `release` |

### 5.2 数据库配置

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `DB_HOST` | mysql | 数据库主机（容器内） |
| `DB_PORT` | 3306 | 数据库端口 |
| `DB_USER` | lawuser | 数据库用户 |
| `DB_PASSWORD` | lawpass | 数据库密码 |
| `DB_NAME` | law_oa | 数据库名 |
| `DB_MAX_OPEN_CONNS` | 100 | 最大连接数 |
| `DB_MAX_IDLE_CONNS` | 20 | 最大空闲连接 |
| `DB_CONN_MAX_LIFETIME` | 1h | 连接最大存活时间 |

### 5.3 Redis 配置

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `REDIS_HOST` | redis | Redis 主机 |
| `REDIS_PORT` | 6379 | Redis 端口 |
| `REDIS_PASSWORD` | (空) | Redis 密码 |
| `REDIS_POOL_SIZE` | 20 | 连接池大小 |

### 5.4 JWT 认证配置

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `JWT_SECRET` | (需设置) | JWT 签名密钥，**至少32字符** |
| `JWT_EXPIRE_HOURS` | 24 | Access Token 有效期（小时） |
| `JWT_REFRESH_EXPIRE_HOURS` | 168 | Refresh Token 有效期（小时） |

### 5.5 安全配置

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `CORS_ALLOWED_ORIGINS` | localhost | 允许的跨域域名，逗号分隔 |
| `RATE_LIMIT_ENABLED` | true | 是否启用限流 |
| `RATE_LIMIT_REQUESTS_PER_MINUTE` | 100 | 每分钟请求限制 |
| `RATE_LIMIT_BURST` | 20 | 突发请求上限 |

### 5.6 监控配置

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `PROMETHEUS_ENABLED` | true | 启用 Prometheus 指标 |
| `PROMETHEUS_PATH` | /metrics | 指标采集路径 |
| `TRACING_ENABLED` | true | 启用分布式追踪 |
| `JAEGER_ENDPOINT` | http://jaeger:14268/api/traces | Jaeger 采集端点 |

### 5.7 OnlyOffice 配置

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `ONLYOFFICE_URL` | http://localhost:9090 | OnlyOffice 服务地址 |
| `ONLYOFFICE_SECRET` | (空) | OnlyOffice JWT 密钥 |
| `BACKEND_URL` | http://localhost:8080 | 后端回调地址 |

### 5.8 文件上传配置

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `UPLOAD_MAX_FILE_SIZE` | 50MB | 单文件最大大小 |
| `UPLOAD_ALLOWED_TYPES` | jpeg,png,pdf,doc,docx | 允许的文件类型 |
| `UPLOAD_STORAGE_PATH` | /app/uploads | 上传存储路径 |

---

## 6. 数据库管理

### 6.1 数据库迁移

系统使用 25 个迁移文件管理数据库结构。

```bash
# 执行迁移（升级）
make migrate-up

# 回滚迁移
make migrate-down

# 创建新迁移
make migrate-create name=add_new_table

# 查看迁移状态
docker compose exec mysql mysql -u lawuser -plawpass law_oa -e "SELECT * FROM schema_migrations;"
```

### 6.2 数据库连接

```bash
# 从宿主机连接
mysql -h 127.0.0.1 -P 33060 -u lawuser -plawpass law_oa

# 从容器内连接
docker compose exec mysql mysql -u lawuser -plawpass law_oa

# 查看数据库列表
docker compose exec mysql mysql -u lawuser -plawpass -e "SHOW DATABASES;"
```

### 6.3 数据库表概览

系统包含 43+ 张表，核心表包括：

| 表名 | 说明 |
|------|------|
| `users` | 用户表 |
| `clients` | 客户表 |
| `cases` | 案件表 |
| `approval_requests` | 审批请求表 |
| `approval_records` | 审批记录表 |
| `approval_delegations` | 代理审批配置表 |
| `documents` | 文档表 |
| `case_folders` | 案件文件夹表 |
| `folder_templates` | 卷宗模板表 |
| `contracts` | 合同表 |
| `invoices` | 发票表 |
| `payments` | 回款表 |
| `commissions` | 提成表 |
| `commission_rules` | 分成规则表 |
| `fee_templates` | 费率模板表 |
| `bad_debts` | 坏账表 |
| `trust_accounts` | 代管款账户表 |
| `trust_transactions` | 代管款交易表 |
| `entities` | 冲突主体表 |
| `entity_relations` | 主体关联表 |
| `conflict_checks` | 冲突检测记录表 |
| `ethical_walls` | 隔离墙配置表 |
| `inbox_items` | 待办事项表 |
| `notifications` | 通知表 |
| `notification_templates` | 通知模板表 |
| `sensitive_words` | 敏感词表 |

### 6.4 数据库性能调优

MySQL 容器已配置以下优化参数：

```
innodb_buffer_pool_size = 256M
max_connections = 200
query_cache_size = 32M
slow_query_log = ON
long_query_time = 2s
```

**查看慢查询日志**:

```bash
docker compose exec mysql cat /var/log/mysql/slow.log
```

### 6.5 常用查询

```sql
-- 查看案件统计
SELECT status, COUNT(*) FROM cases GROUP BY status;

-- 查看待处理审批
SELECT id, title, status, applicant_name FROM approval_requests
WHERE status IN ('submitted', 'under_review') ORDER BY created_at DESC;

-- 查看活跃代理配置
SELECT * FROM approval_delegations WHERE is_active = true AND (valid_until IS NULL OR valid_until > NOW());

-- 查看隔离墙状态
SELECT ew.*, c.title as case_title FROM ethical_walls ew
JOIN cases c ON ew.case_id = c.id WHERE ew.is_active = true;
```

---

## 7. 监控与可观测性

### 7.1 Prometheus 监控

**访问地址**: `http://localhost:9090`

```bash
# 检查 Prometheus 健康
curl http://localhost:9090/-/healthy

# 常用查询
# API 请求总量
http_requests_total

# API 请求延迟 (P99)
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))

# 活跃连接数
go_goroutines
```

### 7.2 Grafana 仪表盘

**访问地址**: `http://localhost:3000`
**默认账号**: admin / admin2024

已预配置的仪表盘包括：
- 系统概览
- API 性能
- 数据库监控
- Redis 缓存

### 7.3 Jaeger 分布式追踪

**访问地址**: `http://localhost:16686`

用于追踪 API 请求的完整调用链路，排查性能瓶颈。

### 7.4 Kibana 日志分析

**访问地址**: `http://localhost:5601`

配合 Elasticsearch 进行日志搜索和分析。

### 7.5 健康检查端点

| 服务 | 端点 | 说明 |
|------|------|------|
| 后端 | `GET /api/v1/health` | 应用健康检查 |
| Prometheus | `GET /metrics` | Prometheus 指标 |

```bash
# 后端健康检查
curl http://localhost:8080/api/v1/health

# Prometheus 指标
curl http://localhost:8080/metrics
```

---

## 8. API 使用指南

### 8.1 认证

**登录获取 Token**:

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "your-password"}'
```

响应中包含 `access_token`，后续请求在 Header 中携带：

```
Authorization: Bearer <access_token>
```

**登出撤销 Token**:

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer <token>"
```

### 8.2 审批流程

**创建审批**:

```bash
curl -X POST http://localhost:8080/api/v1/approvals \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "合同审批",
    "type": "contract",
    "description": "新案件合同审批",
    "priority": "high"
  }'
```

**提交审批**:

```bash
curl -X POST http://localhost:8080/api/v1/approvals/{id}/submit \
  -H "Authorization: Bearer <token>"
```

**处理审批决定**:

```bash
curl -X POST http://localhost:8080/api/v1/approvals/{id}/decision \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"decision": "approved", "comment": "同意"}'
```

### 8.3 代理审批

**创建代理配置**:

```bash
curl -X POST http://localhost:8080/api/v1/approvals/delegations \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "delegator_id": "user-uuid-1",
    "delegate_id": "user-uuid-2",
    "valid_from": "2026-04-01T00:00:00Z",
    "valid_until": "2026-04-30T23:59:59Z",
    "reason": "出差期间代理审批"
  }'
```

**查看我的代理配置**:

```bash
curl http://localhost:8080/api/v1/approvals/delegations/my \
  -H "Authorization: Bearer <token>"
```

**撤销代理**:

```bash
curl -X DELETE http://localhost:8080/api/v1/approvals/delegations/{id} \
  -H "Authorization: Bearer <token>"
```

### 8.4 冲突检测

**创建冲突检测**:

```bash
curl -X POST http://localhost:8080/api/v1/conflict-v2/checks \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "entity_ids": [1, 2, 3],
    "case_id": 1
  }'
```

### 8.5 隔离墙

**启用隔离墙**:

```bash
curl -X POST http://localhost:8080/api/v1/cases/{case_id}/ethical-wall \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"reason": "利益冲突保护"}'
```

**添加白名单**:

```bash
curl -X POST http://localhost:8080/api/v1/cases/{case_id}/ethical-wall/whitelist \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"user_id": 5, "reason": "需要查阅历史文件"}'
```

### 8.6 卷宗模板

**应用模板创建文件夹**:

```bash
curl -X POST http://localhost:8080/api/v1/folder-templates/apply \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": 1,
    "case_id": 10
  }'
```

### 8.7 文档在线编辑

**打开 OnlyOffice 编辑器**:

```bash
curl -X POST http://localhost:8080/api/v1/documents/onlyoffice/open \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"document_id": 1}'
```

### 8.8 时效计算

**计算诉讼时效**:

```bash
curl -X POST http://localhost:8080/api/v1/deadlines/calculate \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "case_type": "civil",
    "event_date": "2026-04-01",
    "deadline_type": "statute_of_limitations"
  }'
```

### 8.9 API 路由总览

系统提供 259 个 API 端点，主要路由组：

| 路由组 | 路径前缀 | 认证 | 说明 |
|--------|----------|------|------|
| 公开API | `/api/v1/auth`, `/api/v1/legal` | 无 | 登录注册、法条搜索 |
| 用户管理 | `/api/v1/admin/users`, `/api/v1/users` | JWT | 用户 CRUD |
| 案件管理 | `/api/v1/cases` | JWT + 隔离墙 | 案件 CRUD |
| 审批管理 | `/api/v1/approvals` | JWT | 审批全流程 |
| 代理审批 | `/api/v1/approvals/delegations` | JWT | 代理配置 |
| 财务管理 | `/api/v1/finance` | JWT | 合同/发票/回款/提成 |
| 代管款 | `/api/v1/trust` | JWT | 代管款账户/交易 |
| 冲突检测 | `/api/v1/conflict-v2` | JWT | Entity 冲突检测 |
| 隔离墙 | `/api/v1/cases/:id/ethical-wall` | JWT + 隔离墙 | 隔离墙管理 |
| 文档管理 | `/api/v1/documents` | JWT + 隔离墙 | 文档上传/预览/编辑 |
| OnlyOffice | `/api/v1/documents/onlyoffice` | JWT | 在线编辑/格式转换 |
| 通知管理 | `/api/v1/notifications` | JWT | 通知队列/模板 |
| 待办事项 | `/api/v1/inbox` | JWT | 待办/提醒 |
| 时效管理 | `/api/v1/deadlines` | JWT | 时效计算 |
| 卷宗模板 | `/api/v1/folder-templates` | JWT | 卷宗模板管理 |
| 内容过滤 | `/api/v1/content-filter` | JWT | 敏感词/内容检测 |
| 团队管理 | `/api/v1/teams` | JWT | 团队分配/权限 |
| 仪表盘 | `/api/v1/dashboard` | JWT | 统计概览 |

---

## 9. 安全配置

### 9.1 生产环境安全清单

- [ ] 更换 `JWT_SECRET` 为强随机密钥（>= 32 字符）
- [ ] 设置 `CORS_ALLOWED_ORIGINS` 为实际域名（不使用通配符）
- [ ] 设置 `REDIS_PASSWORD`
- [ ] 设置 `DB_PASSWORD` 和 `MYSQL_ROOT_PASSWORD` 为强密码
- [ ] 设置 `ENVIRONMENT=production`，关闭 `DEBUG`
- [ ] 设置 `GIN_MODE=release`
- [ ] 启用 HTTPS（配置 SSL 证书或使用反向代理）
- [ ] 配置 `RATE_LIMIT_REQUESTS_PER_MINUTE` 合理值
- [ ] 限制数据库端口不对外暴露

### 9.2 限流配置

系统使用滑动窗口限流器（sync.Map + 时间窗口）：

```bash
# 启用限流
RATE_LIMIT_ENABLED=true

# 每分钟每 IP 最大请求数
RATE_LIMIT_REQUESTS_PER_MINUTE=100

# 突发请求上限
RATE_LIMIT_BURST=20
```

### 9.3 PII 数据脱敏

以下字段在 API 响应中自动脱敏，不返回明文：
- 身份证号 (`IDCard`)
- 手机号 (`Phone`)

### 9.4 安全头部

系统自动添加以下安全头部：
- `Content-Security-Policy`
- `Strict-Transport-Security`
- `Cross-Origin-Opener-Policy`
- `Cross-Origin-Embedder-Policy`

### 9.5 令牌管理

```bash
# 撤销某用户所有令牌
curl -X POST http://localhost:8080/api/v1/auth/revoke/user \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{"user_id": 123}'

# 撤销某设备令牌
curl -X POST http://localhost:8080/api/v1/auth/revoke/device \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{"user_id": 123, "device_id": "device-uuid"}'

# 查看活跃设备
curl http://localhost:8080/api/v1/auth/devices/123 \
  -H "Authorization: Bearer <admin-token>"
```

---

## 10. 故障排除

### 10.1 常见问题

#### 后端无法启动

```bash
# 查看后端日志
docker compose logs backend --tail=200

# 常见原因：
# 1. 数据库连接失败 → 检查 DB_HOST/DB_PORT/DB_PASSWORD
# 2. Redis 连接失败 → 检查 REDIS_HOST/REDIS_PORT
# 3. 端口被占用 → 检查 8080 端口
# 4. 迁移未执行 → 运行 make migrate-up
```

#### 健康检查失败

```bash
# 检查服务状态
docker compose ps

# 查看具体服务的健康检查日志
docker inspect --format='{{json .State.Health}}' law-oa-backend | python3 -m json.tool
```

#### 数据库连接数耗尽

```bash
# 查看当前连接数
docker compose exec mysql mysql -u lawuser -plawpass -e "SHOW STATUS LIKE 'Threads_connected';"

# 查看连接详情
docker compose exec mysql mysql -u lawuser -plawpass -e "SHOW PROCESSLIST;"

# 解决方案：调大 DB_MAX_OPEN_CONNS 或检查连接泄漏
```

#### Elasticsearch 启动失败

```bash
# 常见原因：vm.max_map_count 不够
# 临时解决
docker compose exec elasticsearch sysctl -w vm.max_map_count=262144

# 永久解决（宿主机）
sudo sysctl -w vm.max_map_count=262144
echo "vm.max_map_count=262144" | sudo tee -a /etc/sysctl.conf
```

#### 内存不足

```bash
# 查看各服务内存使用
docker stats --no-stream

# 减少内存限制：编辑 docker-compose.yml 中对应服务的 deploy.resources.limits.memory
```

### 10.2 日志查看

```bash
# 后端实时日志
docker compose logs -f backend

# 后端最近100行日志
docker compose logs --tail=100 backend

# MySQL 慢查询日志
docker compose exec mysql tail -f /var/log/mysql/slow.log

# 所有服务错误日志
docker compose logs -f 2>&1 | grep -i error
```

### 10.3 容器调试

```bash
# 进入后端容器
docker compose exec backend sh

# 进入 MySQL 容器
docker compose exec mysql bash

# 检查容器内网络连通性
docker compose exec backend ping -c 3 mysql
docker compose exec backend ping -c 3 redis
```

---

## 11. 备份与恢复

### 11.1 数据库备份

```bash
# 手动备份
docker compose exec mysql mysqldump -u root -p${MYSQL_ROOT_PASSWORD} --single-transaction --routines --triggers law_oa > backup_$(date +%Y%m%d_%H%M%S).sql

# 压缩备份
docker compose exec mysql mysqldump -u root -p${MYSQL_ROOT_PASSWORD} --single-transaction law_oa | gzip > backup_$(date +%Y%m%d).sql.gz
```

### 11.2 数据恢复

```bash
# 从备份恢复
docker compose exec -T mysql mysql -u root -p${MYSQL_ROOT_PASSWORD} law_oa < backup_20260407.sql

# 从压缩备份恢复
gunzip < backup_20260407.sql.gz | docker compose exec -T mysql mysql -u root -p${MYSQL_ROOT_PASSWORD} law_oa
```

### 11.3 上传文件备份

上传文件存储在 `./data/uploads/` 目录（Docker 卷映射）。

```bash
# 备份上传文件
tar -czf uploads_backup_$(date +%Y%m%d).tar.gz ./data/uploads/

# 恢复上传文件
tar -xzf uploads_backup_20260407.tar.gz -C ./
```

### 11.4 定时备份（crontab）

```bash
# 编辑 crontab
crontab -e

# 每天凌晨2点备份数据库
0 2 * * * cd /path/to/law-oa-go && docker compose exec -T mysql mysqldump -u root -plawroot2024 --single-transaction law_oa | gzip > /backup/law_oa_$(date +\%Y\%m\%d).sql.gz

# 每周日凌晨3点备份上传文件
0 3 * * 0 tar -czf /backup/uploads_$(date +\%Y\%m\%d).tar.gz /path/to/law-oa-go/data/uploads/

# 保留最近30天备份
0 4 * * * find /backup -name "law_oa_*.sql.gz" -mtime +30 -delete
```

---

## 12. 开发环境

### 12.1 本地开发

```bash
# 安装 Go 依赖
make deps

# 代码格式化
make fmt

# 代码检查
make lint

# 运行测试
make test

# 运行安全检查
make security

# 开发模式（热重载）
make dev
```

### 12.2 前端开发

```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm start

# 构建生产版本
npm run build

# 运行前端测试
npm test

# 运行 E2E 测试
npx playwright test
```

### 12.3 Make 命令速查

| 命令 | 说明 |
|------|------|
| `make build` | 构建应用 |
| `make build-linux` | 构建 Linux 版本 |
| `make docker-build` | 构建 Docker 镜像 |
| `make test` | 运行所有测试 |
| `make test-unit` | 单元测试 |
| `make test-coverage` | 生成覆盖率报告 |
| `make lint` | 代码检查 |
| `make security` | 安全检查（gosec） |
| `make quality` | 完整质量检查 |
| `make ci` | CI 流水线 |
| `make release` | 发布准备 |
| `make migrate-up` | 执行数据库迁移 |
| `make migrate-down` | 回滚数据库迁移 |
| `make migrate-create name=xxx` | 创建迁移文件 |
| `make help` | 查看所有命令 |

### 12.4 定时任务

系统内置 3 个定时任务（goroutine），随后端服务自动启动：

| 任务 | 间隔 | 说明 |
|------|------|------|
| 提醒检查 | 1 小时 | 检查到期待办并发送提醒 |
| 升级检查 | 6 小时 | 检查超时 critical 待办并升级 |
| 审批超时检查 | 30 分钟 | 检查审批超时（48h）并发送预警 |

审批超时相关常量：

| 常量 | 值 | 说明 |
|------|------|------|
| 审批超时阈值 | 48 小时 | 默认超时时间 |
| 超时预警比例 | 80% | 超时前发送提醒 |
| 待办升级天数 | 3 天 | critical 待办升级阈值 |
| 每次处理数量 | 50 条 | 审批超时批处理大小 |
