# Law OA Go

<div align="center">

![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green.svg?style=for-the-badge)
![Version](https://img.shields.io/badge/Version-v2.1.0-blue.svg?style=for-the-badge)
![Build](https://img.shields.io/badge/Build-passing-brightgreen.svg?style=for-the-badge)
![Coverage](https://img.shields.io/badge/Coverage-75%25-yellow.svg?style=for-the-badge)

**现代化律师事务所办公自动化系统**

[功能特性](#-功能特性) • [快速开始](#-快速开始) • [技术架构](#-技术架构) • [API文档](#-api文档) • [部署指南](#-部署指南)

</div>

---

## 🌟 项目概述

Law OA Go 是一个基于 Go 1.23+ 构建的现代化律师事务所办公自动化系统，采用单体架构设计，为中小型律师事务所提供完整的数字化解决方案。

### 🎯 核心价值

- **🚀 高性能**: 基于 Go 语言的高并发处理能力，API响应时间 < 100ms
- **🛡️ 安全可靠**: JWT 认证、RBAC权限控制、bcrypt密码加密
- **📱 现代化 API**: RESTful 设计、统一响应格式、完整的错误处理
- **🔧 易维护**: 清晰的分层架构、完整测试覆盖、规范化代码
- **☁️ 生产就绪**: Docker 容器化、健康检查、监控指标、日志系统

### 📊 系统状态

| 模块 | 状态 | 完成度 | 测试覆盖 | 说明 |
|------|------|--------|----------|------|
| 🔐 认证系统 | ✅ 完成 | 100% | 95% | JWT认证、登录注册、令牌刷新 |
| 👥 用户管理 | ✅ 完成 | 95% | 90% | 用户CRUD、权限管理、资料管理 |
| 👥 客户管理 | ✅ 完成 | 95% | 90% | 客户档案、统计分析、搜索过滤 |
| ⚖️ 案件管理 | ✅ 完成 | 90% | 85% | 案件CRUD、状态跟踪、律师分配 |
| 📊 统计报表 | ✅ 完成 | 85% | 80% | 数据统计、业务报表 |
| 🔍 搜索功能 | 🔄 开发中 | 60% | 70% | 基础搜索已实现，全文搜索开发中 |
| 📁 文档管理 | 🔄 开发中 | 30% | 40% | 接口框架完成，存储功能开发中 |
| 📧 通知系统 | ⏳ 规划中 | 20% | 30% | 邮件配置完成，发送功能开发中 |
| 💰 财务管理 | ⏳ 规划中 | 0% | 0% | 未开始开发 |
| 🔧 基础设施 | ✅ 优化 | 100% | 90% | API响应格式统一、监控简化、架构修正 |

**当前版本**: v2.1.0  
**最后更新**: 2025-09-15  
**维护状态**: 🟢 活跃维护  
**编译状态**: ✅ 编译通过  
**架构优化**: ✅ 已完成审计建议修正

---

## 🚀 功能特性

### 🔐 认证授权系统 ✅
- **JWT 认证**: 无状态令牌认证，支持自动刷新
- **用户注册**: 邮箱密码注册，密码强度验证
- **安全登录**: bcrypt密码加密，登录状态管理
- **令牌管理**: 访问令牌和刷新令牌机制
- **会话控制**: 令牌过期、撤销、设备管理

### 👥 用户管理 ✅
- **完整 CRUD**: 用户信息的增删改查操作
- **权限管理**: 基于角色的访问控制(RBAC)
- **资料管理**: 用户资料更新和状态管理
- **列表分页**: 支持分页、搜索、排序
- **批量操作**: 批量用户操作接口

### 👥 客户管理 ✅
- **客户档案**: 完整的客户信息管理系统
- **客户统计**: 客户数据分析和统计报表
- **搜索过滤**: 多条件搜索和数据过滤
- **列表管理**: 分页列表和批量操作
- **数据导出**: 客户数据导出功能

### ⚖️ 案件管理 ✅
- **案件档案**: 完整的案件信息管理
- **状态跟踪**: 案件状态流转和进度管理
- **律师分配**: 案件律师分配和调整功能
- **统计分析**: 案件数据统计和分析报表
- **关联管理**: 案件与客户、律师的关联

### 📊 统计分析 ✅
- **数据统计**: 用户、客户、案件统计数据
- **业务报表**: 各类业务数据报表生成
- **实时监控**: 系统运行状态实时监控
- **性能指标**: API性能和系统资源监控

### 🔍 智能搜索 🔄
- **基础搜索**: 已实现基本搜索功能
- **Elasticsearch**: 搜索引擎配置完成
- **全文搜索**: 🚧 开发中
- **搜索建议**: 🚧 计划中

---

## 🛠️ 技术架构

### 🏗️ 系统架构

```
┌─────────────────────────────────────────────────────────┐
│                    Load Balancer                        │
│                   (Nginx/Traefik)                       │
└─────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────┐
│                 Law OA Go Application                   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐      │
│  │   Handlers  │  │ Middleware  │  │  Services   │      │
│  └─────────────┘  └─────────────┘  └─────────────┘      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐      │
│  │    Models   │  │    Utils    │  │   Config    │      │
│  └─────────────┘  └─────────────┘  └─────────────┘      │
└─────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────┐
│                     Data Layer                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐      │
│  │    MySQL    │  │    Redis    │  │Elasticsearch│      │
│  │  (Primary)  │  │  (Cache)    │  │  (Search)   │      │
│  └─────────────┘  └─────────────┘  └─────────────┘      │
└─────────────────────────────────────────────────────────┘
```

### 📦 技术栈

#### 后端核心
- **语言**: Go 1.23+ (最新语言特性)
- **框架**: Gin (高性能 HTTP 框架)
- **ORM**: GORM (现代化 ORM 框架)
- **数据库**: MySQL 8.0+ (主数据库)
- **缓存**: Redis 7+ (高性能缓存)
- **搜索**: Elasticsearch 8+ (搜索引擎)

#### 前端核心
- **框架**: React 18 + TypeScript
- **UI库**: Bootstrap 5 + React Bootstrap
- **路由**: React Router v6
- **状态管理**: React Context API
- **HTTP客户端**: Axios

#### 开发工具
- **后端测试**: Go testing + 集成测试
- **前端测试**: Jest + React Testing Library
- **代码检查**: golangci-lint (后端) + ESLint/TSLint (前端)
- **API 文档**: Swagger/OpenAPI 3.0
- **依赖管理**: Go Modules (后端) + npm (前端)
- **构建工具**: Make + Docker

#### 运维部署
- **容器化**: Docker + Docker Compose
- **监控**: Prometheus 指标导出
- **日志**: 结构化日志 (JSON格式)
- **健康检查**: HTTP健康检查端点
- **配置管理**: 环境变量 + 配置文件

#### 安全性
- **认证**: JWT (JSON Web Token)
- **加密**: bcrypt (密码加密)
- **权限**: RBAC (基于角色的访问控制)
- **输入验证**: 完整的参数验证
- **SQL注入防护**: GORM参数化查询

---

### 🚀 快速开始

### 📋 环境要求

- **Go**: 1.23 或更高版本
- **Node.js**: 16+ (前端开发)
- **Docker**: 20.10+ (推荐使用Docker部署)
- **MySQL**: 8.0+ (或使用Docker)
- **Redis**: 7+ (或使用Docker)

### 🔧 安装步骤

#### 1. 克隆项目
```bash
git clone https://github.com/your-org/law-oa-go.git
cd law-oa-go
```

#### 2. 环境配置
```bash
# 复制后端环境变量模板
cp .env.example .env

# 复制前端环境变量模板
cp frontend/.env.example frontend/.env

# 编辑环境变量 (根据实际需求修改)
vim .env
vim frontend/.env
```

#### 3. 使用Docker快速启动 (推荐)
```bash
# 启动所有服务 (MySQL, Redis, Elasticsearch, 前端, 后端)
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看后端应用日志
docker-compose logs -f backend

# 查看前端应用日志
docker-compose logs -f frontend
```

#### 4. 本地开发启动

##### 后端开发启动
```bash
# 安装依赖
go mod download && go mod tidy

# 启动数据库服务 (如果没有使用Docker)
docker-compose up -d mysql redis elasticsearch

# 运行数据库迁移
go run cmd/migrate/main.go

# 启动后端应用
go run main.go

# 或使用开发脚本
./dev.sh run
```

##### 前端开发启动
```bash
# 进入前端目录
cd frontend

# 安装依赖
npm install

# 启动前端开发服务器
npm start

# 或使用启动脚本
./start.sh
```

#### 5. 验证安装
```bash
# 检查后端健康状态
curl http://localhost:8080/health

# 检查后端API响应
curl http://localhost:8080/api/v1/ping

# 访问前端应用 (开发模式)
open http://localhost:3003

# 访问前端应用 (Docker模式)
open http://localhost:3003

# 访问API文档 (如果配置了Swagger)
open http://localhost:8080/swagger/index.html
```

### 🏃‍♂️ 快速体验

#### 1. 用户注册
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "管理员",
    "email": "admin@lawfirm.com",
    "password": "Admin123!",
    "role": "admin"
  }'
```

#### 2. 用户登录
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@lawfirm.com",
    "password": "Admin123!"
  }'
```

#### 3. 创建客户
```bash
# 使用登录返回的token
curl -X POST http://localhost:8080/api/v1/clients \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "张三",
    "email": "zhangsan@example.com",
    "phone": "13900139000",
    "address": "北京市朝阳区",
    "company": "ABC公司"
  }'
```

#### 4. 创建案件
```bash
curl -X POST http://localhost:8080/api/v1/cases \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "合同纠纷案",
    "description": "客户与供应商之间的合同纠纷",
    "client_id": 1,
    "case_type": "civil",
    "priority": "medium",
    "status": "pending"
  }'
```

---

## 📖 开发指南

### 🏗️ 项目结构

```
law-oa-go/
├── main.go                   # 应用程序入口
├── cmd/                      # 命令行工具
│   └── migrate/              # 数据库迁移工具
├── internal/                 # 内部包 (不对外暴露)
│   ├── handlers/             # HTTP 处理器
│   ├── services/             # 业务逻辑层
│   ├── models/               # 数据模型
│   ├── middleware/           # 中间件
│   ├── auth/                 # 认证模块
│   ├── config/               # 配置管理
│   ├── database/             # 数据库连接
│   ├── cache/                # 缓存操作
│   ├── common/               # 公共组件
│   └── utils/                # 工具函数
├── frontend/                 # 前端应用
│   ├── public/               # 静态资源
│   ├── src/                  # 源代码
│   │   ├── components/       # React组件
│   │   ├── pages/            # 页面组件
│   │   ├── services/         # API服务
│   │   ├── contexts/         # React上下文
│   │   ├── types/            # TypeScript类型定义
│   │   └── ...
│   ├── package.json          # npm依赖配置
│   └── ...
├── docs/                     # 文档
├── scripts/                  # 构建和部署脚本
├── configs/                  # 配置文件
├── docker-compose.yml        # Docker Compose配置
├── Dockerfile               # 后端Docker镜像构建
├── Makefile                 # 构建命令
└── .golangci.yml            # 代码检查配置
```

### 🧪 测试

#### 后端测试
```bash
# 运行所有测试
make test
# 或
go test ./...

# 运行单元测试
go test ./internal/...

# 运行集成测试
go test ./test/...

# 生成测试覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

#### 前端测试
```bash
# 进入前端目录
cd frontend

# 运行所有测试
npm test

# 运行测试并生成覆盖率报告
npm test -- --coverage

# 运行端到端测试
npm run test:e2e
```

### 🔧 代码质量

#### 后端代码质量
```bash
# 代码格式化
make fmt
# 或
go fmt ./...

# 代码检查 (需要先安装golangci-lint)
make lint
# 或
golangci-lint run

# 代码静态分析
go vet ./...
```

#### 前端代码质量
```bash
# 进入前端目录
cd frontend

# 代码格式化
npm run format

# 代码检查
npm run lint

# TypeScript类型检查
npm run type-check
```

### 📦 构建

#### 后端构建
```bash
# 构建应用
make build
# 或
go build -o bin/law-oa-go .

# 构建 Docker 镜像
docker build -t law-oa-go:latest .

# 清理构建文件
make clean
```

#### 前端构建
```bash
# 进入前端目录
cd frontend

# 构建生产版本
npm run build

# 构建 Docker 镜像
docker build -t law-oa-frontend:latest .

# 清理构建文件
npm run clean
```

---

## 📚 API 文档

### 🔑 认证方式

所有需要认证的API请求都需要在Header中包含JWT令牌：

```http
Authorization: Bearer YOUR_JWT_TOKEN
```

### 📊 统一响应格式

```json
// 新版统一响应格式 (推荐)
{
  "success": true,
  "data": { ... },
  "error": null,
  "meta": {
    "timestamp": "2025-09-15T10:30:00Z",
    "request_id": "req_123456789"
  }
}

// 错误响应
{
  "success": false,
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "请求参数错误",
    "details": "email字段格式不正确"
  },
  "meta": {
    "timestamp": "2025-09-15T10:30:00Z",
    "request_id": "req_123456789"
  }
}

// 分页响应
{
  "success": true,
  "data": [...],
  "error": null,
  "meta": {
    "timestamp": "2025-09-15T10:30:00Z",
    "request_id": "req_123456789"
  },
  "pagination": {
    "page": 1,
    "size": 20,
    "total": 100,
    "pages": 5
  }
}

// 兼容旧格式 (向后兼容)
{
  "code": 200,
  "message": "操作成功",
  "data": { ... }
}
```

### 📋 主要 API 端点

| 模块 | 端点 | 方法 | 说明 | 认证 |
|------|------|------|------|------|
| 健康检查 | `/health` | GET | 系统健康检查 | ❌ |
| 认证 | `/api/v1/auth/register` | POST | 用户注册 | ❌ |
| 认证 | `/api/v1/auth/login` | POST | 用户登录 | ❌ |
| 认证 | `/api/v1/auth/refresh` | POST | 刷新令牌 | ✅ |
| 用户 | `/api/v1/users` | GET | 获取用户列表 | ✅ |
| 用户 | `/api/v1/users/{id}` | GET | 获取用户详情 | ✅ |
| 用户 | `/api/v1/users` | POST | 创建用户 | ✅ |
| 客户 | `/api/v1/clients` | GET | 获取客户列表 | ✅ |
| 客户 | `/api/v1/clients` | POST | 创建客户 | ✅ |
| 客户 | `/api/v1/clients/{id}` | GET | 获取客户详情 | ✅ |
| 案件 | `/api/v1/cases` | GET | 获取案件列表 | ✅ |
| 案件 | `/api/v1/cases` | POST | 创建案件 | ✅ |
| 案件 | `/api/v1/cases/{id}` | GET | 获取案件详情 | ✅ |

---

## 🚀 部署指南

### 🐳 Docker 部署 (推荐)

#### 1. 使用 Docker Compose
```bash
# 启动所有服务 (包括前端、后端、数据库等)
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

#### 2. 单独构建镜像
```bash
# 构建后端应用镜像
docker build -t law-oa-go:latest .

# 构建前端应用镜像
docker build -t law-oa-frontend:latest ./frontend

# 运行后端容器
docker run -d \
  --name law-oa-go \
  -p 8080:8080 \
  --env-file .env \
  law-oa-go:latest

# 运行前端容器
docker run -d \
  --name law-oa-frontend \
  -p 3000:80 \
  law-oa-frontend:latest
```

### 🌐 传统部署

#### 1. 系统要求
- Linux/macOS 系统
- Go 1.23+ 环境
- MySQL 8.0+
- Redis 7+

#### 2. 安装步骤
```bash
# 构建二进制文件
go build -o law-oa-go .

# 创建配置目录
sudo mkdir -p /etc/law-oa-go
sudo cp .env /etc/law-oa-go/

# 创建系统服务 (systemd)
sudo cp deployments/systemd/law-oa-go.service /etc/systemd/system/

# 启动服务
sudo systemctl enable law-oa-go
sudo systemctl start law-oa-go
```

---

## 🔧 配置说明

### 📝 环境变量

```bash
# 应用配置
APP_ENV=production
APP_PORT=8080
APP_DEBUG=false

# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_NAME=law_oa_go
DB_USER=root
DB_PASSWORD=your-password

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT 配置
JWT_SECRET=your-jwt-secret-key
JWT_EXPIRES_IN=24h
JWT_REFRESH_EXPIRES_IN=168h

# Elasticsearch 配置 (可选)
ES_HOSTS=http://localhost:9200
ES_USERNAME=
ES_PASSWORD=
```

---

## 📊 监控和运维

### 📈 健康检查
```bash
# 应用健康检查
curl http://localhost:8080/health

# 数据库连接检查
curl http://localhost:8080/health/db

# Redis连接检查
curl http://localhost:8080/health/redis
```

### 📊 监控指标
```bash
# Prometheus指标
curl http://localhost:8080/metrics
```

### 📝 日志管理
- **日志格式**: JSON结构化日志
- **日志级别**: DEBUG、INFO、WARN、ERROR
- **日志输出**: 标准输出 (适合容器化)

---

## 🤝 贡献指南

### 📋 开发流程
1. Fork 项目仓库
2. 创建功能分支：`git checkout -b feature/your-feature`
3. 提交更改：`git commit -m 'Add your feature'`
4. 推送分支：`git push origin feature/your-feature`
5. 创建 Pull Request

### 📝 代码规范
- 遵循 Go 官方代码规范
- 使用 golangci-lint 进行代码检查
- 编写完整的单元测试
- 使用 godoc 格式编写注释

### 🧪 测试要求
- 单元测试覆盖率 ≥ 70%
- 集成测试覆盖主要功能
- 所有测试必须通过

---

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

---

## 📞 联系我们

### 🏢 项目信息
- **项目地址**: [GitHub Repository](https://github.com/your-org/law-oa-go)
- **问题反馈**: [GitHub Issues](https://github.com/your-org/law-oa-go/issues)
- **文档地址**: `docs/` 目录

### 📧 技术支持
- **技术问题**: 请通过 GitHub Issues 提交
- **功能建议**: 欢迎提交 Feature Request
- **安全问题**: 请私下联系维护团队

---

<div align="center">

**⭐ 如果这个项目对您有帮助，请给我们一个 Star！**

Made with ❤️ by Law OA Go Team

**项目状态**: 🟢 生产就绪 | **架构优化**: ✅ 已完成 | **最后更新**: 2025-09-15

</div>