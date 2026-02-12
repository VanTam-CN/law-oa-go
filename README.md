# Law OA Go

<div align="center">

![Go](https://img.shields.io/badge/Go-1.23-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-ISC-green.svg?style=for-the-badge)
![Version](https://img.shields.io/badge/Version-v2.2.0-blue.svg?style=for-the-badge)
![Database](https://img.shields.io/badge/Database-PostgreSQL%2BMySQL%2BSQLite-blue.svg?style=for-the-badge)
![React](https://img.shields.io/badge/React-18.2.0-61DAFB?style=for-the-badge&logo=react&logoColor=white)
![TypeScript](https://img.shields.io/badge/TypeScript-5.0.2-3178C6?style=for-the-badge&logo=typescript&logoColor=white)
![Vite](https://img.shields.io/badge/Vite-5.1.0-646CFF?style=for-the-badge&logo=vite&logoColor=white)

**现代化律师事务所办公自动化系统**

[功能特性](#-功能特性)  [技术架构](#-技术架构)  [快速开始](#-快速开始)  [部署指南](#-部署指南)  [API文档](#api文档)

</div>

---

## 项目概述

Law OA Go 是一个基于 Go 1.23 构建的现代化律师事务所办公自动化系统，采用单体架构设计，为中小型律师事务所提供数字化解决方案。系统支持 PostgreSQL、MySQL 和 SQLite 多种数据库环境部署。

### 核心价值

- **高性能**: 基于 Go 语言的高并发处理能力，优化的API响应
- **安全可靠**: JWT 认证、RBAC权限控制、bcrypt密码加密、令牌撤销机制
- **现代化前端**: React 18 + TypeScript 5.0 + Ant Design 5 + Vite 5
- **数据库灵活**: 支持 PostgreSQL、MySQL 和 SQLite，环境自适应
- **易维护**: 清晰的分层架构、测试体系、规范化代码
- **生产就绪**: Docker 容器化、健康检查、监控指标、结构化日志
- **企业级搜索**: 基于 Elasticsearch 8.9 的全文检索功能
- **可观测性**: Prometheus + Grafana + Jaeger 完整监控体系
- **隔离墙机制**: 律师事务所利益冲突隔离保护

### 系统状态

| 模块 | 状态 | 完成度 | 数据库支持 | 说明 |
|------|------|--------|------------|------|
| 认证系统 | 完成 | 100% | PostgreSQL + MySQL + SQLite | JWT认证、用户管理、RBAC、令牌撤销 |
| 用户管理 | 完成 | 95% | PostgreSQL + MySQL + SQLite | 用户信息、权限管理 |
| 客户管理 | 完成 | 95% | PostgreSQL + MySQL + SQLite | 分类管理、统计分析 |
| 案件管理 | 完成 | 90% | PostgreSQL + MySQL + SQLite | 状态跟踪、律师分配、增强版 |
| 统计报表 | 完成 | 85% | PostgreSQL + MySQL + SQLite | ECharts图表展示 |
| 搜索功能 | 完成 | 90% | Elasticsearch 8.9 | 全文检索、智能排序 |
| 文档管理 | 部分完成 | 65% | PostgreSQL + MySQL + SQLite | 基础上传下载、统计报告 |
| 通知系统 | 完成 | 85% | PostgreSQL + MySQL + SQLite | 通知队列、模板管理 |
| 财务管理 | 完成 | 80% | PostgreSQL + MySQL + SQLite | 合同、发票、支付、佣金报告 |
| 冲突检测 | 完成 | 90% | PostgreSQL + MySQL + SQLite | 多维度检测、风险评估、v2增强版 |
| 审批流程 | 完成 | 95% | PostgreSQL + MySQL + SQLite | 多级审批、状态机、冲突集成 |
| 信任账户 | 完成 | 85% | PostgreSQL + MySQL + SQLite | 账户管理、交易记录 |
| 内容过滤 | 完成 | 90% | PostgreSQL + MySQL + SQLite | 敏感词检测、内容过滤 |
| 隔离墙 | 完成 | 90% | PostgreSQL + MySQL + SQLite | 利益冲突隔离、白名单 |
| 待办事项 | 完成 | 80% | PostgreSQL + MySQL + SQLite | 收件箱管理 |
| 离职流程 | 部分完成 | 60% | PostgreSQL + MySQL + SQLite | 基础流程框架 |
| 协作聊天 | 未开发 | 0% | - | 计划中 |

**当前版本**: v2.2.0
**最后更新**: 2026-02-12
**维护状态**: 活跃维护
**编译状态**: 编译通过
**数据库状态**: PostgreSQL + MySQL + SQLite 多环境支持
**代码规模**: ~600 Go文件, ~15k TypeScript文件
**整体完成度**: 约 80%

---

## 功能特性

### 认证与授权
- JWT 令牌认证机制 (golang-jwt/jwt/v5)
- 用户注册、登录、登出
- 令牌自动刷新
- 密码安全加密存储 (bcrypt)
- 角色权限管理（RBAC）
- 中间件权限控制
- **令牌撤销服务**: 设备管理、用户令牌撤销、撤销历史

### 用户管理
- 用户信息管理
- 头像上传功能
- 密码修改
- 用户状态管理
- 权限分配
- 用户统计报表

### 客户管理
- 客户档案管理
- 客户分类（个人/企业）
- 联系信息管理
- 客户统计分析
- 高级搜索过滤
- 客户案件关联

### 案件管理
- 案件信息管理（基础版 + 增强版）
- 案件状态跟踪
- 律师分配
- 案件文档管理
- 案件统计分析
- 案件优先级管理
- 案件费用管理
- 案件客户关联

### 搜索功能
- 多词智能搜索
- 实时搜索建议
- 搜索结果高亮
- 相关性排序
- 分类搜索过滤
- Elasticsearch 8.9 集成
- 全文检索优化
- 搜索历史记录
- 法条搜索

### 统计报表
- 实时数据统计
- 业务图表展示 (ECharts 5.6.0 + Recharts 3.1.2)
- 导出报表功能
- 性能指标监控
- 自定义报表

### 冲突检测系统
- 多维度冲突检测（律师利益、客户关系、行业竞争）
- 智能风险评估（CRITICAL/HIGH/MEDIUM/LOW）
- 检测报告生成
- 冲突历史记录
- 实时冲突检查
- 检测统计分析
- 冲突分类服务
- **v2增强版**: 更精准的检测算法

### 审批流程管理
- 多级审批流程
- 审批状态机（待提交→审批中→通过/驳回/取消）
- 审批历史记录
- 审批权限控制
- 审批统计分析
- 审批分配器
- **冲突检测集成**: 审批与冲突检测联动

### 财务管理系统
- **合同管理**: 合同列表、创建、编辑、状态跟踪
- **发票管理**: 发票开具、跟踪、统计
- **支付管理**: 支付记录、状态跟踪
- **佣金报告**: 律师佣金计算、报表生成
- **坏账服务**: 坏账管理

### 信任账户管理
- 信任账户创建与管理
- 交易记录跟踪
- 账户余额监控
- 交易统计报表

### 文档管理
- 文档上传/下载
- 文档预览
- 搜索和过滤
- 统计报告
- 文档统计（存储使用、用户活动、合规报告）
> 注意：版本控制、权限管理、回收站等高级功能开发中

### 通知系统
- 通知队列管理
- 通知模板系统（创建、编辑、启用/禁用）
- 通知审批流程（通过/拒绝）
- 批量操作（确认、取消）
- 通知统计分析

### 内容过滤系统
- 敏感词管理（CRUD、批量导入、批量操作）
- 内容检测与过滤
- 过滤日志记录
- 敏感词统计分析
- 缓存管理

### 隔离墙机制
- 律师事务所利益冲突隔离
- 白名单管理
- 中间件保护
- 案件访问控制

### 待办事项（收件箱）
- 待办事项创建与管理
- 统计信息
- 状态跟踪

### 离职流程
- 员工离职管理基础框架

### 工具模块
- 诉讼费计算器
- 利息计算器
- 截止日期计算器
- 法律搜索

### 系统管理
- 系统监控面板
- 操作日志记录
- 性能监控
- 缓存管理
- 配置管理
- 安全审计功能
- 健康检查端点

---

## 技术架构

### 后端技术栈
- **语言**: Go 1.23 (toolchain go1.23.6)
- **框架**: Gin Web Framework v1.10.1
- **数据库**: PostgreSQL / MySQL 8.0 / SQLite
- **ORM**: GORM v1.30.0
- **缓存**: Redis go-redis v9.0.5
- **搜索**: Elasticsearch 8.9 (go-elasticsearch v8.9.0)
- **认证**: JWT (golang-jwt/jwt/v5 v5.0.0)
- **日志**: Zap v1.24.0 + Lumberjack v2.0.0
- **监控**: Prometheus v1.16.0 + OpenTelemetry v1.38.0
- **追踪**: Jaeger v1.17.0
- **验证**: go-playground/validator v10.26.0
- **配置管理**: Viper v1.16.0
- **CORS**: gin-contrib/cors v1.7.6
- **UUID**: google/uuid v1.6.0
- **数据库迁移**: golang-migrate/migrate v4.17.1
- **API文档**: Swagger (swaggo/gin-swagger v1.5.3)

### 前端技术栈
- **框架**: React 18.2.0 + TypeScript 5.0.2
- **构建工具**: Vite 5.1.0
- **UI 组件**: Ant Design 5.16.1
- **状态管理**:
  - Redux Toolkit 2.9.1
  - TanStack React Query 5.90.5
  - Zustand 5.0.8
- **路由**: React Router 7.9.4
- **HTTP 客户端**: Axios 1.12.2
- **图表**:
  - ECharts 5.6.0
  - Recharts 3.1.2
  - echarts-for-react 3.0.2
- **工具库**:
  - Lodash 4.17.21
  - Dayjs 1.11.10
- **测试**:
  - Jest 30.2.0
  - Vitest 3.2.4
  - Testing Library
- **代码质量**:
  - ESLint 8.57.1
  - Prettier 3.6.2
  - Husky 9.1.7
- **文档处理**: Puppeteer 24.22.3
- **代理**: http-proxy-middleware 3.0.5

### 基础设施
- **容器化**: Docker & Docker Compose
- **反向代理**: Nginx
- **监控**: Prometheus + Grafana
- **追踪**: Jaeger Distributed Tracing
- **日志**: 结构化日志系统 (Zap + Lumberjack)
- **CI/CD**: GitHub Actions + Husky + Commitlint
- **文档存储**: 本地存储
- **缓存**: Redis 多级缓存
- **搜索**: Elasticsearch 8.9 集群 + Kibana 8.9

### 架构设计
```
┌─────────────────────────────────────────────────────────────┐
│                        Frontend Layer                       │
│              React 18 + TypeScript 5.0 + Ant Design 5       │
│                 Vite 5 + Redux + React Query                │
├─────────────────────────────────────────────────────────────┤
│                      API Gateway                            │
│                  Nginx + Gin Middleware                     │
│           Authentication + Rate Limiting + Ethical Wall     │
├─────────────────────────────────────────────────────────────┤
│                    Business Logic                           │
│  ┌──────────────┬──────────────┬──────────────┬────────────┐ │
│  │ User/Auth    │ Case Mgmt    │ Document     │ Conflict   │ │
│  │ Services     │ Services     │ Services     │ Detection  │ │
│  └──────────────┴──────────────┴──────────────┴────────────┘ │
│  ┌──────────────┬──────────────┬──────────────┬────────────┐ │
│  │ Notification │ Approval     │ Search       │ Dashboard  │ │
│  │ Services     │ Services     │ Services     │ Services   │ │
│  └──────────────┴──────────────┴──────────────┴────────────┘ │
│  ┌──────────────┬──────────────┬──────────────┬────────────┐ │
│  │ Finance      │ Trust Acct   │ Content      │ Ethical    │ │
│  │ Services     │ Services     │ Filter       │ Wall       │ │
│  └──────────────┴──────────────┴──────────────┴────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                      Data Layer                             │
│  ┌──────────────┬──────────────┬──────────────┬────────────┐ │
│  │ PostgreSQL   │ MySQL        │ Redis Cache  │ File Store │ │
│  │ Primary      │ Support      │ Multi-level  │ Local      │ │
│  └──────────────┴──────────────┴──────────────┴────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                   Search & Observability                     │
│  ┌──────────────┬──────────────┬──────────────┬────────────┐ │
│  │ Elasticsearch│ Prometheus   │ Grafana      │ Jaeger     │ │
│  │ 8.9          │ Metrics      │ Dashboards   │ Tracing    │ │
│  └──────────────┴──────────────┴──────────────┴────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

---

## 快速开始

### 环境要求
- Go 1.23+
- Node.js 18+
- PostgreSQL 15+ 或 MySQL 8.0+ 或 SQLite 3
- Redis 7+
- Docker & Docker Compose (推荐)

### 数据库配置

#### PostgreSQL（推荐）
```bash
# 启动 PostgreSQL 服务
docker compose -f docker-compose.yml up -d postgresql redis

# 配置环境变量
cp .env.example .env
# 编辑 .env 文件，设置 PostgreSQL 配置
# DB_DRIVER=postgres
# DB_HOST=localhost
# DB_PORT=5432
# DB_USERNAME=law_oa_user
# DB_PASSWORD=your_password
# DB_DATABASE=law_oa_db
```

#### MySQL
```bash
# 启动 MySQL 服务
docker compose -f docker-compose.yml up -d mysql redis

# 使用默认配置
cp .env.example .env
# DB_DRIVER=mysql
```

#### SQLite（开发环境）
```bash
# 使用默认 SQLite 配置
cp .env.example .env
# 设置 DB_DRIVER=sqlite
# DB_DATABASE=./law_oa.db
```

### 快速启动

1. **克隆项目**
```bash
git clone https://github.com/VanTam-CN/law-oa-go.git
cd law-oa-go
```

2. **配置环境**
```bash
# 复制环境配置文件
cp .env.example .env

# 编辑配置（设置数据库连接、JWT密钥等）
vim .env
```

3. **启动服务**
```bash
# 方式一：使用 Docker Compose（推荐）
docker compose up -d

# 方式二：本地开发
# 启动后端
go run cmd/server/main.go

# 启动前端
cd frontend
npm install
npm run dev
```

4. **访问应用**
- 前端: http://localhost:3003
- 后端 API: http://localhost:8080
- API 文档: http://localhost:8080/swagger/index.html
- 健康检查: http://localhost:8080/api/health
- Grafana: http://localhost:3000
- Prometheus: http://localhost:9090
- Jaeger: http://localhost:16686
- Kibana: http://localhost:5601

### 验证安装

访问 http://localhost:3003，使用默认管理员账号：
- 用户名: admin
- 密码: admin123

---

## API 文档

### 主要 API 端点

#### 认证相关
- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/logout` - 用户登出
- `POST /api/v1/auth/revoke/user` - 撤销用户令牌
- `POST /api/v1/auth/revoke/device` - 撤销设备令牌
- `GET /api/v1/auth/devices/:user_id` - 获取活跃设备

#### 客户管理
- `GET /api/v1/clients` - 获取客户列表
- `POST /api/v1/clients` - 创建客户
- `GET /api/v1/clients/:id` - 获取客户详情
- `PUT /api/v1/clients/:id` - 更新客户信息
- `DELETE /api/v1/clients/:id` - 删除客户
- `GET /api/v1/clients/stats` - 客户统计

#### 案件管理
- `GET /api/v1/cases` - 获取案件列表
- `POST /api/v1/cases` - 创建案件
- `GET /api/v1/cases/:id` - 获取案件详情
- `PUT /api/v1/cases/:id` - 更新案件信息
- `DELETE /api/v1/cases/:id` - 删除案件

#### 增强案件管理
- `GET /api/v1/enhanced-cases` - 获取增强案件列表
- `POST /api/v1/enhanced-cases` - 创建增强案件
- `POST /api/v1/enhanced-cases/:id/conflict-check` - 执行冲突检测
- `POST /api/v1/enhanced-cases/:id/clients` - 添加客户到案件

#### 统计分析
- `GET /api/v1/dashboard/statistics` - 获取统计数据
- `GET /api/v1/dashboard/todos` - 获取待办事项
- `GET /api/v1/dashboard/activities` - 获取活动记录

#### 冲突检测
- `POST /api/v1/conflict/check` - 执行冲突检测
- `GET /api/v1/conflict/history` - 获取检测历史
- `GET /api/v1/conflict/stats` - 获取冲突统计

#### 审批流程
- `GET /api/v1/approvals` - 获取审批列表
- `POST /api/v1/approvals` - 创建审批
- `POST /api/v1/approvals/:id/submit` - 提交审批
- `POST /api/v1/approvals/:id/decision` - 处理审批决定
- `POST /api/v1/approvals/:id/resubmit` - 重新提交
- `POST /api/v1/approvals/:id/cancel` - 取消审批
- `GET /api/v1/approvals/stats` - 审批统计

#### 集成审批
- `POST /api/v1/integration/approvals/with-conflict` - 创建带冲突检测的审批
- `GET /api/v1/integration/approvals/:id/status` - 获取集成状态
- `POST /api/v1/integration/approvals/:id/case` - 执行案件创建
- `GET /api/v1/integration/statistics` - 获取集成统计

#### 通知系统
- `GET /api/v1/notifications` - 获取通知队列列表
- `POST /api/v1/notifications` - 创建通知
- `POST /api/v1/notifications/:id/approve` - 审批通过
- `POST /api/v1/notifications/:id/reject` - 审批拒绝
- `POST /api/v1/notifications/:id/send` - 发送通知
- `GET /api/v1/notifications/stats` - 通知统计

#### 通知模板
- `GET /api/v1/notification-templates` - 获取模板列表
- `POST /api/v1/notification-templates` - 创建模板
- `PUT /api/v1/notification-templates/:id/toggle` - 切换启用状态

#### 内容过滤
- `POST /api/v1/content-filter/words` - 创建敏感词
- `GET /api/v1/content-filter/words` - 获取敏感词列表
- `POST /api/v1/content-filter/check` - 检查内容
- `POST /api/v1/content-filter/filter` - 过滤内容
- `GET /api/v1/content-filter/logs` - 获取过滤日志

#### 财务管理
- `GET /api/v1/finance/contracts` - 获取合同列表
- `GET /api/v1/finance/invoices` - 获取发票列表
- `GET /api/v1/finance/payments` - 获取支付列表
- `GET /api/v1/finance/commissions` - 获取佣金报告

#### 信任账户
- `GET /api/v1/trust/accounts` - 获取信任账户列表
- `GET /api/v1/trust/transactions` - 获取交易记录
- `GET /api/v1/trust/stats` - 获取统计

#### 文档管理
- `GET /api/v1/documents` - 获取文档列表
- `POST /api/v1/documents` - 上传文档
- `GET /api/v1/documents/:id` - 获取文档详情
- `PUT /api/v1/documents/:id` - 更新文档
- `DELETE /api/v1/documents/:id` - 删除文档
- `GET /api/v1/documents/:id/download` - 下载文档
- `GET /api/v1/documents/:id/preview` - 预览文档
- `GET /api/v1/documents/stats/overview` - 文档统计

#### 待办事项
- `GET /api/v1/inbox` - 获取待办列表
- `POST /api/v1/inbox` - 创建待办
- `PUT /api/v1/inbox/:id` - 更新待办
- `DELETE /api/v1/inbox/:id` - 删除待办
- `GET /api/v1/inbox/stats` - 获取统计

### API 认证
所有 API（除登录注册外）都需要在请求头中包含 JWT 令牌：

```http
Authorization: Bearer <your-jwt-token>
```

### 响应格式
所有 API 都采用统一的响应格式：

```json
{
  "data": { ... },
  "error": null,
  "message": "success",
  "timestamp": "2026-02-12T22:30:00Z"
}
```

---

## 部署指南

### Docker 部署（推荐）

1. **准备环境**
```bash
# 安装 Docker 和 Docker Compose
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# 验证安装
docker --version
docker compose --version
```

2. **配置环境**
```bash
# 复制并编辑配置文件
cp .env.example .env
vim .env
```

3. **启动服务**
```bash
# 启动所有服务
docker compose up -d

# 查看服务状态
docker compose ps

# 查看日志
docker compose logs -f
```

4. **访问服务**
- 前端应用: http://localhost:3003
- 后端 API: http://localhost:8080
- Grafana: http://localhost:3000 (admin/admin2024)
- Prometheus: http://localhost:9090
- Jaeger: http://localhost:16686
- Kibana: http://localhost:5601

### 手动部署

1. **后端部署**
```bash
# 编译后端
go build -o law-oa-server ./cmd/server

# 启动服务
./law-oa-server
```

2. **前端部署**
```bash
# 构建前端
cd frontend
npm run build

# 使用 nginx 或其他 web 服务器托管 dist 目录
```

### 环境变量配置

#### 数据库配置
```bash
# PostgreSQL
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=law_oa_user
DB_PASSWORD=your_password
DB_DATABASE=law_oa_db
DB_SSLMODE=disable

# MySQL
DB_DRIVER=mysql
DB_HOST=localhost
DB_PORT=3306
DB_USERNAME=root
DB_PASSWORD=your_password
DB_DATABASE=law_oa

# SQLite
DB_DRIVER=sqlite
DB_DATABASE=./law_oa.db
```

#### Redis 配置
```bash
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_password
REDIS_DB=0
```

#### Elasticsearch 配置
```bash
ES_HOST=http://localhost
ES_PORT=9200
ES_USERNAME=elastic
ES_PASSWORD=your_password
```

#### JWT 配置
```bash
JWT_SECRET=your-very-secure-jwt-secret-key-minimum-32-characters
JWT_EXPIRES_IN=3600
JWT_REFRESH_IN=7200
```

---

## 数据库管理

### 数据库支持

项目现已完全支持多数据库，包含以下特性：

#### 数据库支持
- **PostgreSQL**: 生产环境推荐，完整功能支持
- **MySQL**: 传统环境支持，完整功能兼容
- **SQLite**: 开发环境支持，轻量级部署

#### 数据库切换

项目支持运行时数据库切换，通过环境变量配置：

```bash
# 使用 PostgreSQL
DB_DRIVER=postgres

# 使用 MySQL
DB_DRIVER=mysql

# 使用 SQLite
DB_DRIVER=sqlite
```

#### 迁移指南

1. **数据备份**
```bash
# 备份 MySQL 数据
mysqldump -u root -p law_oa > backup.sql

# 备份 PostgreSQL 数据
pg_dump -U law_oa_user law_oa_db > backup.sql
```

2. **创建数据库**
```sql
-- PostgreSQL
CREATE DATABASE law_oa_db;
CREATE USER law_oa_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE law_oa_db TO law_oa_user;

-- MySQL
CREATE DATABASE law_oa CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'law_oa_user'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON law_oa.* TO 'law_oa_user'@'localhost';
```

3. **运行迁移**
```bash
# 使用 GORM 自动迁移
go run cmd/server/main.go

# 或使用迁移工具
make migrate-up
```

### 搜索优化

#### 全文搜索
- **多词搜索**: 支持空格分隔的搜索词
- **相关性排序**: 智能结果排序
- **中文支持**: 优化的中文文本处理
- **性能优化**: 索引优化和缓存策略

#### Elasticsearch 集成
- **自动索引**: 实时数据同步
- **高级搜索**: 复杂查询支持
- **性能监控**: 搜索性能统计
- **集群管理**: 健康检查和故障转移

---

## 开发指南

### 项目结构
```
law-oa-go/
├── cmd/                      # 应用入口点
│   ├── server/              # 主服务器启动
│   ├── migrate/             # 数据库迁移工具
│   └── gen_password/        # 密码生成工具
├── internal/                 # 核心业务逻辑
│   ├── handlers/            # HTTP 处理器
│   ├── services/            # 业务服务层
│   ├── repositories/        # 数据访问层
│   ├── models/              # 数据模型
│   ├── middleware/          # 中间件
│   ├── database/            # 数据库相关
│   ├── cache/               # 缓存模块
│   ├── elasticsearch/       # 搜索引擎
│   ├── config/              # 配置管理
│   ├── security/            # 安全模块
│   ├── auth/                # 认证模块
│   ├── rbac/                # 权限控制
│   ├── logger/              # 日志配置
│   ├── metrics/             # 监控指标
│   ├── tracing/             # 分布式追踪
│   └── utils/               # 工具函数
├── frontend/                 # 前端代码
│   ├── src/                # React 组件源码
│   │   ├── pages/          # 页面组件
│   │   ├── components/     # 通用组件
│   │   ├── services/       # API 服务
│   │   ├── stores/         # 状态管理
│   │   ├── hooks/          # 自定义 Hooks
│   │   ├── utils/          # 工具函数
│   │   ├── types/          # TypeScript 类型
│   │   └── assets/styles/  # 样式文件
│   ├── tests/              # 测试文件
│   └── package.json        # 前端依赖配置
├── scripts/                  # 脚本工具
├── configs/                  # 配置文件
├── docs/                     # 项目文档
├── tests/                    # 测试目录
├── migrations/               # 数据库迁移文件
├── docker-compose.yml        # Docker Compose 配置
├── Dockerfile               # 后端 Docker 镜像
├── Makefile                 # 构建命令配置
├── go.mod / go.sum          # Go 模块依赖
├── .env.example             # 环境变量模板
├── .air.toml                # 热重载配置
├── .golangci.yml            # 代码检查配置
└── README.md                # 项目说明文档
```

### 开发环境设置

1. **后端开发**
```bash
# 安装依赖
go mod download

# 运行测试
go test ./...

# 代码格式化
gofmt -s -w ./...

# 静态检查
golangci-lint run

# 热重载开发
air
```

2. **前端开发**
```bash
# 安装依赖
npm install

# 开发模式
npm run dev

# 构建生产版本
npm run build

# 运行测试
npm run test
npm run test:coverage

# 类型检查
npm run type-check

# 代码检查
npm run lint
npm run lint:fix

# 代码格式化
npm run format
```

### 代码规范

#### Go 代码规范
- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 使用 `golangci-lint` 进行静态检查
- API 接口必须有文档注释

#### 前端代码规范
- 使用 TypeScript 严格模式
- 遵循 ESLint 和 Prettier 配置
- 组件使用函数式组件
- 使用 React Hooks 管理状态
- 遵循 Ant Design 设计规范

### 测试策略

#### 后端测试
```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./internal/services

# 生成测试覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### 前端测试
```bash
# 运行单元测试
npm run test

# 运行集成测试
npm run test:integration

# 生成测试覆盖率报告
npm run test:coverage
```

---

## 监控与日志

### 应用监控

#### 性能指标
- **响应时间**: API 响应时间统计
- **吞吐量**: 每秒请求数
- **错误率**: 错误请求比例
- **资源使用**: CPU、内存、磁盘使用率

#### 健康检查
```bash
# 应用健康状态
GET /api/health

# 数据库连接状态
GET /api/health/database

# Redis 连接状态
GET /api/health/redis

# Elasticsearch 状态
GET /api/health/elasticsearch
```

### 日志管理

#### 日志级别
- `debug`: 调试信息
- `info`: 一般信息
- `warn`: 警告信息
- `error`: 错误信息
- `fatal`: 致命错误

#### 日志格式
```json
{
  "level": "info",
  "timestamp": "2026-02-12T22:30:00Z",
  "message": "User login successful",
  "module": "auth",
  "user_id": "123",
  "ip": "192.168.1.100",
  "request_id": "req-123456",
  "duration": 45
}
```

### 可观测性工具

#### Prometheus + Grafana
- **指标收集**: Prometheus 1.16.0
- **可视化**: Grafana (版本由 docker 镜像决定)
- **告警规则**: 自定义告警配置
- **仪表板**: 预配置的业务仪表板

#### Jaeger 分布式追踪
- **追踪系统**: Jaeger 1.17.0
- **OpenTelemetry**: 集成 OTEL SDK
- **性能分析**: 请求链路追踪
- **瓶颈定位**: 自动化性能分析

#### 日志聚合
- **日志收集**: Lumberjack 日志轮转
- **结构化日志**: Zap 高性能日志
- **日志查询**: Kibana 日志查询
- **日志分析**: Elasticsearch 日志索引

---

## 故障排除

### 常见问题

#### 数据库连接失败
```bash
# 检查数据库服务状态
docker ps | grep postgresql

# 检查连接配置
cat .env | grep DB_

# 测试数据库连接
docker exec law-oa-postgresql psql -U law_oa_user -d law_oa_db -c "SELECT 1;"
```

#### Redis 连接失败
```bash
# 检查 Redis 服务
docker ps | grep redis

# 测试 Redis 连接
docker exec law-oa-redis redis-cli ping
```

#### 端口占用
```bash
# 查看端口占用
lsof -i :8080

# 更改后端端口
export PORT=8081
go run cmd/server/main.go
```

#### 前端构建失败
```bash
# 清理缓存
rm -rf node_modules package-lock.json
npm install

# 检查 Node.js 版本
node --version  # 需要 >= 18.0
```

### 性能优化

#### 数据库优化
- 定期执行 `VACUUM` 和 `ANALYZE` (PostgreSQL)
- 合理设置连接池大小
- 添加适当的索引
- 监控慢查询

#### 缓存优化
- 合理设置 Redis 过期时间
- 使用缓存预热策略
- 监控缓存命中率

#### 搜索优化
- 定期重建搜索索引
- 优化查询语句
- 监控搜索性能

---

## 路线图

### 待完成功能

#### 高优先级
- [ ] 协作聊天功能（WebSocket 实时通信）
- [ ] 文档管理高级功能（版本控制、权限管理、回收站）
- [ ] 离职流程完善

#### 中优先级
- [ ] 移动端适配
- [ ] 高级报表功能

#### 低优先级
- [ ] 国际化支持
- [ ] 第三方集成（邮件、短信等）

---

## 更新日志

### v2.2.0 (2026-02-12)
- **财务管理**: 完整实现合同、发票、支付、佣金报告模块
- **信任账户**: 信任账户管理与交易记录
- **通知系统**: 通知队列与模板管理系统
- **内容过滤**: 敏感词管理与内容检测过滤
- **隔离墙**: 律师事务所利益冲突隔离机制
- **令牌撤销**: JWT令牌撤销服务（设备管理、用户令牌撤销）
- **冲突检测v2**: 增强版冲突检测算法
- **待办事项**: 收件箱管理系统
- **离职流程**: 基础流程框架
- **审批集成**: 审批流程与冲突检测深度集成

### v2.1.0 (2026-02-09)
- **重大更新**: 冲突检测系统全面完成
  - 多维度冲突检测算法
  - 智能风险评估系统
  - 检测报告生成和历史记录
  - 实时冲突检查功能
  - 冲突分类服务
- **审批流程系统**: 全新实现
  - 多级审批流程管理
  - 审批状态机
  - 审批权限控制
  - 审批历史追踪
  - 审批分配器
- **文档管理**: 完整实现
  - 文件上传/下载
  - 版本控制
  - 权限管理
  - 文档预览
  - 搜索和过滤
  - 回收站功能
- **搜索优化**: Elasticsearch 集成完成
  - 全文检索功能
  - 高级查询支持
  - 搜索性能优化
  - 搜索历史记录
- **技术升级**:
  - PostgreSQL 数据库适配完成
  - 多级缓存架构优化
  - 前端组件库更新
  - API 性能优化
  - 监控体系完善
- **项目优化**:
  - 代码结构规范化
  - 测试覆盖率提升
  - 文档体系完善
  - CI/CD 流程优化
  - 清理冗余文件，释放约550MB空间
  - 前端代码质量改进和优化
  - API响应格式标准化

### v2.0.0 (2025-09-28)
- **架构重构**: 全面重构项目架构
- **UI 升级**: 前端框架现代化升级
- **安全增强**: 权限系统全面升级
- **监控完善**: 监控系统简化优化
- **文档完善**: 开发文档和 API 文档完善

---

## 贡献指南

### 贡献流程

1. Fork 项目仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'feat: Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 开发规范

1. 遵循项目的代码规范
2. 添加适当的测试用例
3. 更新相关文档
4. 确保所有测试通过
5. 提交前运行 `go mod tidy` 和 `npm run build`

### 提交规范

```
feat: 新功能
fix: 修复问题
docs: 文档更新
style: 代码格式
refactor: 代码重构
test: 测试相关
chore: 构建过程或辅助工具的变动
perf: 性能优化
ci: CI/CD相关
build: 构建系统相关
revert: 回滚提交
```

---

## 许可证

本项目采用 ISC 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

---

## 联系方式

- **项目维护者**: VanTam-CN
- **项目地址**: [GitHub Repository](https://github.com/VanTam-CN/law-oa-go)
- **问题反馈**: [Issues Page](https://github.com/VanTam-CN/law-oa-go/issues)

---

<div align="center">

**感谢使用 Law OA Go！**

如果这个项目对您有帮助，请给我们一个

</div>
