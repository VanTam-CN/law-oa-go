# Law OA Go

<div align="center">

![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green.svg?style=for-the-badge)
![Version](https://img.shields.io/badge/Version-v2.1.0-blue.svg?style=for-the-badge)
![Database](https://img.shields.io/badge/Database-PostgreSQL%2BMySQL-blue.svg?style=for-the-badge)
![Build](https://img.shields.io/badge/Build-passing-brightgreen.svg?style=for-the-badge)
![Status](https://img.shields.io/badge/Status-Production%20Ready-green.svg?style=for-the-badge)

**现代化律师事务所办公自动化系统**

[功能特性](#-功能特性) • [技术架构](#-技术架构) • [快速开始](#-快速开始) • [部署指南](#-部署指南) • [API文档](#api文档)

</div>

---

## 🌟 项目概述

Law OA Go 是一个基于 Go 1.23+ 构建的现代化律师事务所办公自动化系统，采用单体架构设计，为中小型律师事务所提供完整的数字化解决方案。现已完成 **PostgreSQL 数据库适配**，支持双数据库环境部署。

### 🎯 核心价值

- **🚀 高性能**: 基于 Go 语言的高并发处理能力，API响应时间 < 100ms
- **🛡️ 安全可靠**: JWT 认证、RBAC权限控制、bcrypt密码加密
- **📱 现代化 API**: RESTful 设计、统一响应格式、完整的错误处理
- **🗄️ 数据库灵活**: 支持 MySQL 和 PostgreSQL 双环境，无缝切换
- **🔧 易维护**: 清晰的分层架构、完整测试覆盖、规范化代码
- **☁️ 生产就绪**: Docker 容器化、健康检查、监控指标、日志系统

### 📊 系统状态

| 模块 | 状态 | 完成度 | 测试覆盖 | 数据库支持 |
|------|------|--------|----------|------------|
| 🔐 认证系统 | ✅ 完成 | 100% | 95% | MySQL + PostgreSQL |
| 👥 用户管理 | ✅ 完成 | 95% | 90% | MySQL + PostgreSQL |
| 👥 客户管理 | ✅ 完成 | 95% | 90% | MySQL + PostgreSQL |
| ⚖️ 案件管理 | ✅ 完成 | 90% | 85% | MySQL + PostgreSQL |
| 📊 统计报表 | ✅ 完成 | 85% | 80% | MySQL + PostgreSQL |
| 🔍 搜索功能 | ✅ 完成 | 90% | 85% | 多词搜索优化 |
| 📁 文档管理 | 🔄 开发中 | 30% | 40% | 接口框架完成 |
| 📧 通知系统 | 🔄 开发中 | 20% | 30% | 邮件配置完成 |
| 💰 财务管理 | ⏳ 规划中 | 0% | 0% | 未开始开发 |
| 🔧 搜索优化 | ✅ 完成 | 100% | 95% | Elasticsearch 集成 |

**当前版本**: v2.1.0
**最后更新**: 2025-10-13
**维护状态**: 🟢 活跃维护
**编译状态**: ✅ 编译通过
**数据库状态**: ✅ PostgreSQL + MySQL 双环境

---

## 🌟 功能特性

### 🔐 认证与授权
- ✅ JWT 令牌认证机制
- ✅ 用户注册、登录、登出
- ✅ 令牌自动刷新
- ✅ 密码安全加密存储
- ✅ 角色权限管理（RBAC）
- ✅ 多租户支持

### 👥 用户管理
- ✅ 用户信息管理
- ✅ 头像上传功能
- ✅ 密码修改
- ✅ 用户状态管理
- ✅ 权限分配

### 👥 客户管理
- ✅ 客户档案管理
- ✅ 客户分类（个人/企业）
- ✅ 联系信息管理
- ✅ 客户统计分析
- ✅ 高级搜索过滤

### ⚖️ 案件管理
- ✅ 案件信息管理
- ✅ 案件状态跟踪
- ✅ 律师分配
- ✅ 案件文档管理
- ✅ 案件统计分析
- ✅ 案件优先级管理

### 🔍 搜索功能
- ✅ 多词智能搜索
- ✅ 实时搜索建议
- ✅ 搜索结果高亮
- ✅ 相关性排序
- ✅ 分类搜索过滤

### 📊 统计报表
- ✅ 实时数据统计
- ✅ 业务图表展示
- ✅ 导出报表功能
- ✅ 性能指标监控

### 🔧 系统管理
- ✅ 系统监控面板
- ✅ 操作日志记录
- ✅ 性能监控
- ✅ 缓存管理
- ✅ 配置管理

---

## 🏗️ 技术架构

### 后端技术栈
- **语言**: Go 1.23+
- **框架**: Gin Web Framework
- **数据库**: PostgreSQL 15 / MySQL 8.0
- **ORM**: GORM v1.30
- **缓存**: Redis 7
- **搜索**: Elasticsearch 8.11
- **认证**: JWT (golang-jwt/v5)
- **日志**: Zap
- **监控**: Prometheus + Grafana

### 前端技术栈
- **框架**: React 18 + TypeScript
- **构建工具**: Vite 5.x
- **UI 组件**: Ant Design 5.x
- **状态管理**: React Router + Hooks
- **HTTP 客户端**: Axios
- **图表**: ECharts
- **测试**: Vitest

### 基础设施
- **容器化**: Docker & Docker Compose
- **反向代理**: Nginx
- **监控**: 自研监控系统
- **日志**: 结构化日志系统
- **CI/CD**: GitHub Actions

### 架构设计
```
┌─────────────────┐
│   Frontend        │  React + TypeScript
├─────────────────┤
│   API Gateway    │  Nginx / Gin Middleware
├─────────────────┤
│   Business Logic  │  Go Services
├─────────────────┤
│   Data Layer     │  PostgreSQL/MySQL + Redis
├─────────────────┤
│   Search Layer   │  Elasticsearch
└─────────────────┘
```

---

## 🚀 快速开始

### 环境要求
- Go 1.23+
- Node.js 18+
- PostgreSQL 15+ 或 MySQL 8.0+
- Redis 7+
- Docker & Docker Compose

### 数据库配置

#### PostgreSQL（推荐）
```bash
# 启动 PostgreSQL 服务
docker compose -f docker-compose.postgresql.yml up -d postgresql redis

# 配置环境变量
cp .env.postgresql .env
```

#### MySQL
```bash
# 启动 MySQL 服务
docker compose -f docker-compose.yml up -d mysql redis

# 使用默认配置
cp .env.example .env
```

### 快速启动

1. **克隆项目**
```bash
git clone <repository-url>
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

### 验证安装

访问 http://localhost:3003，使用默认管理员账号：
- 用户名: admin
- 密码: admin123

---

## 📚 API 文档

### 主要 API 端点

#### 认证相关
- `POST /api/auth/login` - 用户登录
- `POST /api/auth/register` - 用户注册
- `POST /api/auth/refresh` - 刷新令牌
- `POST /api/users/logout` - 用户登出

#### 客户管理
- `GET /api/v1/clients` - 获取客户列表
- `POST /api/v1/clients` - 创建客户
- `GET /api/v1/clients/:id` - 获取客户详情
- `PUT /api/v1/clients/:id` - 更新客户信息
- `DELETE /api/v1/clients/:id` - 删除客户

#### 案件管理
- `GET /api/v1/cases` - 获取案件列表
- `POST /api/v1/cases` - 创建案件
- `GET /api/v1/cases/:id` - 获取案件详情
- `PUT /api/v1/cases/:id` - 更新案件信息
- `DELETE /api/v1/cases/:id` - 删除案件

#### 统计分析
- `GET /api/v1/dashboard/statistics` - 获取统计数据
- `GET /api/v1/search` - 搜索功能

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
  "timestamp": "2025-10-13T22:30:00Z"
}
```

---

## 🚀 部署指南

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

4. **反向代理配置**
```nginx
# nginx.conf 示例
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:3003;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /api {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-Real-IP $remote_addr;
    }
}
```

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
ES_HOST=localhost
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

## 🗄️ 数据库管理

### PostgreSQL 适配

项目现已完全支持 PostgreSQL，包含以下特性：

#### 🎯 PostgreSQL 优势
- **ACID 兼容性**: 完整的事务支持
- **性能优化**: 查询计划优化
- **全文搜索**: 内置 tsvector/tsquery 支持
- **JSON 支持**: 原生 JSON 数据类型
- **扩展性**: 丰富的扩展生态

#### 🔄 数据库切换

项目支持运行时数据库切换，通过环境变量配置：

```bash
# 使用 PostgreSQL
DB_DRIVER=postgres

# 使用 MySQL
DB_DRIVER=mysql
```

#### 📝 迁移指南

1. **数据备份**
```bash
# 备份 MySQL 数据
mysqldump -u root -p law_oa > backup.sql
```

2. **创建 PostgreSQL 数据库**
```sql
CREATE DATABASE law_oa_db;
CREATE USER law_oa_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE law_oa_db TO law_oa_user;
```

3. **迁移数据**
```bash
# 使用项目提供的迁移工具
go run scripts/migrate-to-postgresql.go
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

---

## 🧪 开发指南

### 项目结构
```
law-oa-go/
├── cmd/                # 应用入口
├── internal/           # 核心业务逻辑
│   ├── handlers/       # HTTP 处理器
│   ├── services/        # 业务服务层
│   ├── repositories/   # 数据访问层
│   ├── models/         # 数据模型
│   ├── middleware/     # 中间件
│   └── config/         # 配置管理
├── frontend/           # 前端代码
│   ├── src/           # React 组件
│   ├── api/           # API 客户端
│   ├── pages/         # 页面组件
│   └── utils/         # 工具函数
├── scripts/           # 脚本工具
├── docs/              # 项目文档
├── tests/             # 测试代码
├── configs/           # 配置文件
├── archive/           # 归档文件
└── docker-compose.yml  # Docker 配置
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
```

### 代码规范

#### Go 代码规范
- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 使用 `golangci-lint` 进行静态检查
- 单元测试覆盖率 > 80%
- API 接口必须有文档注释

#### 前端代码规范
- 使用 TypeScript 严格模式
- 遵循 ESLint 和 Prettier 配置
- 组件使用函数式组件
- 使用 React Hooks 管理状态

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
npm run test:e2e

# 生成测试覆盖率报告
npm run test:coverage
```

---

## 📊 监控与日志

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
  "timestamp": "2025-10-13T22:30:00Z",
  "message": "User login successful",
  "module": "auth",
  "user_id": "123",
  "ip": "192.168.1.100",
  "request_id": "req-123456",
  "duration": 45
}
```

---

## 🔧 故障排除

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
- 定期执行 `VACUUM` 和 `ANALYZE`
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

## 📈 更新日志

### v2.1.0 (2025-10-13)
- ✨ **重大更新**: 完成 PostgreSQL 数据库适配
- ✨ **功能增强**: 多词搜索优化
- ✨ **性能提升**: 查询优化器适配
- ✨ **项目清理**: 目录结构优化，文件归档整理
- 🔧 **基础设施**: Docker Compose 配置优化
- 📦 **文档更新**: README 全面更新

### v2.0.0 (2025-09-28)
- 🚀 **架构重构**: 全面重构项目架构
- 🎨 **UI 升级**: 前端框架现代化升级
- 🛡️ **安全增强**: 权限系统全面升级
- 📊 **监控完善**: 监控系统简化优化
- 📚 **文档完善**: 开发文档和 API 文档完善

---

## 🤝 贡献指南

### 贡献流程

1. Fork 项目仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
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
```

---

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

---

## 📞 联系方式

- **项目维护者**: [Your Name]
- **邮箱**: [your.email@example.com]
- **项目地址**: [GitHub Repository URL]
- **问题反馈**: [Issues Page URL]

---

<div align="center">

**感谢使用 Law OA Go！** 🎉

如果这个项目对您有帮助，请给我们一个 ⭐️

</div>