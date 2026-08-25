# Law OA Go v3

![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)
![React](https://img.shields.io/badge/React-18-61DAFB?logo=react&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?logo=postgresql&logoColor=white)
![License](https://img.shields.io/badge/License-Apache--2.0-blue.svg)

Law OA Go 是一个面向小型律师事务所的案件协作系统。v3 聚焦一个清晰闭环：律师登记客户与案件线索，系统校验主体身份并执行利益冲突检查，核查人员独立复核，审批通过后生成正式案件，律师再在案件、客户和待办视图中持续协作。

系统默认保持简单：一个 Go 服务、一个 React 前端和 PostgreSQL。缓存、全文搜索增强、在线文档编辑和观测组件都是可选扩展，不是启动依赖。

## 功能

### 律师工作台

- 汇总待办事项、冲突复核、审批任务和案件动态。
- 提供新建立案、案件检索和客户档案入口。
- 根据律师、核查人和管理角色显示不同操作范围。

### 接案与案件

- 结构化登记客户、对方和相关方信息。
- 案件线索通过审批后生成正式案件编号。
- 支持案件列表、详情、阶段状态、关联客户、冲突复核和后续主体变更。
- 客户档案包含主档案、关联方、历史委托、联系人和附件入口。

### 身份与利益冲突控制

- 自然人和企业分别登记可核验身份信息。
- 敏感身份原文加密存储，不会出现在浏览器草稿中。
- 冲突检查覆盖既有案件、客户、主体和关联关系。
- 检查结果需要独立核查人复核，未完成前不会放行成案。
- 支持新主体登记、合并/驳回、重新检查和版本留痕。

### 协作与权限

- JWT 登录、角色权限和服务端会话撤销。
- 浏览器退出会撤销当前 token，旧 token 不能继续访问受保护接口。
- 待办中心支持筛选、查看、标记完成和延后提醒。
- 律师访问未授权模块时收到明确的权限或范围提示。

### 文档与文件

- 案件文件上传、下载和基础管理。
- 可选接入 OnlyOffice 在线编辑。

## 技术栈

| 层     | 技术                                       |
| ------ | ------------------------------------------ |
| 前端   | React 18、TypeScript 5、Ant Design 5、Vite |
| 后端   | Go 1.25、Gin、GORM                         |
| 数据库 | PostgreSQL 16                              |
| 认证   | JWT、RBAC、PostgreSQL 会话撤销             |
| 部署   | Docker Compose、Kubernetes manifests       |

```text
React SPA
   │ HTTP / JSON
   ▼
Go monolith
   ├── domain services
   ├── repositories
   └── PostgreSQL
```

核心业务状态、认证会话、审批记录和审计数据都保存在 PostgreSQL。应用不依赖外部缓存即可运行。

## 快速启动

### 环境要求

- Go 1.25+
- Node.js 18+
- Docker 和 Docker Compose

### 使用 Docker Compose

```bash
git clone https://github.com/VanTam-CN/law-oa-go.git
cd law-oa-go
cp .env.example .env

# 补齐 ENVIRONMENT、JWT_SECRET、APP_SECRET、SUBJECT_DATA_KEY 等必填配置
$EDITOR .env

docker compose up -d --build
```

默认服务：

| 服务       | 地址                               |
| ---------- | ---------------------------------- |
| 前端       | http://localhost:3003              |
| API        | http://localhost:8080              |
| 健康检查   | http://localhost:8080/health/ready |
| PostgreSQL | localhost:5432                     |

Compose 会在应用启动前执行 PostgreSQL schema bootstrap。

### 本地开发

```bash
# 安装依赖
go mod download
cd frontend && npm install && cd ..

# 准备 PostgreSQL，并在 shell 中设置 DB_DRIVER、DB_HOST、DB_PORT、
# DB_USERNAME、DB_PASSWORD、DB_DATABASE 和 DB_SSLMODE

# 初始化 schema
go run ./cmd/migrate -command bootstrap

# 启动 API
go run .

# 另开终端启动前端
cd frontend
VITE_API_PROXY_TARGET=http://localhost:8080 npm run dev
```

开发环境配置说明见 [docs/CONFIGURATION.md](docs/CONFIGURATION.md)。

## 测试与验证

```bash
# 后端
go test ./... -count=1
go vet ./...
go build ./...

# 前端
cd frontend
npm run type-check
npm run lint
npm run test
npm run build
```

当前基线的关键路径验证包括律师工作台、接案、冲突复核、审批成案、案件/客户回看、律师数据隔离，以及真实浏览器中的服务端登出撤销。

## 可选扩展

| 扩展                          | 启用方式                                       | 用途                 |
| ----------------------------- | ---------------------------------------------- | -------------------- |
| Redis                         | Compose `cache` profile                        | 响应与热点数据缓存   |
| Prometheus / Grafana / Jaeger | Compose `observability` profile                | 指标、面板和链路追踪 |
| OnlyOffice                    | `ONLYOFFICE_ENABLED=true` 并配置服务地址与密钥 | 文档在线编辑         |
| Elasticsearch                 | 单独扩展                                       | 法条与全文搜索增强   |

默认安装不包含这些服务。

## 部署

- 容器编排入口：[docker-compose.yml](docker-compose.yml)
- Kubernetes 入口：[k8s/README.md](k8s/README.md)
- 配置说明：[docs/CONFIGURATION.md](docs/CONFIGURATION.md)
- API 说明：[docs/API.md](docs/API.md)
- 开发约定：[docs/DEVELOPER_GUIDE.md](docs/DEVELOPER_GUIDE.md)

生产部署必须使用独立密钥、真实 CORS 来源、TLS、数据库备份与恢复演练，并完成目标机构的档案覆盖和权限验收。

## 目录结构

```text
cmd/                 命令行入口（迁移、QA 工具、辅助任务）
internal/            Go 业务、接口、仓储、中间件与配置
frontend/src/        React 页面、组件、服务与状态
k8s/                 Kubernetes manifests
docker/              辅助容器编排
docs/                配置、API、开发与领域文档
migrations/          SQL 迁移与 schema 材料
```

## License

This project is licensed under the Apache License 2.0. See [LICENSE](LICENSE).

When redistributing the project or derivative works, keep the applicable copyright, patent, trademark, and attribution notices; if a `NOTICE` file is included, retain its attribution notices. The license also provides a patent grant and terminates that grant if you initiate specified patent litigation. The work is provided on an "AS IS" basis, without warranties or conditions.
