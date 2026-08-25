# Law OA Go — Independent Baseline

![Version](https://img.shields.io/badge/version-v3.0.0--independent-blue)
![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)
![React](https://img.shields.io/badge/React-18-61DAFB?logo=react&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?logo=postgresql&logoColor=white)
![License](https://img.shields.io/badge/License-ISC-green.svg)

Law OA Go 是一个小型律师事务所受控试用 MVP。它提供一个从 **接案登记 → 身份与冲突门禁 → 独立复核 → 审批成案 → 案件与客户回看** 的闭环，而不是完整的商业化律所管理系统。

> **版本口径**：这是从历史仓库瘦身后的独立基线。默认运行路径只依赖 PostgreSQL 和应用本体；Redis、Elasticsearch、OnlyOffice 和观测栈都是显式启用的可选能力。
>
> **生产口径**：当前仍为 **NO-GO**。本仓库只包含受控试用和本地/CI 验证证据，不代表真实律所档案覆盖、生产部署或运维验收已经完成。

## 1. 当前能力

| 能力           | 当前状态        | 说明                                                                                |
| -------------- | --------------- | ----------------------------------------------------------------------------------- |
| 登录与会话     | 可用            | JWT 登录、RBAC、PostgreSQL 会话撤销；UI 退出会调用服务端撤销，旧 token 重放返回 401 |
| 律师工作台     | 可用            | 汇总待办、冲突复核、审批和案件入口                                                  |
| 接案登记       | 受门禁控制      | 客户、对方和相关方必须登记可核验身份；敏感身份原文加密，不进入浏览器草稿            |
| 利益冲突检查   | 受门禁控制      | 按全所历史档案口径检索；覆盖不完整时保持阻断，不得输出“无冲突”放行                  |
| 独立复核与审批 | 受门禁控制      | 支持冲突核查人独立复核、审批通过成案和案件回溯                                      |
| 案件管理       | 受控可用        | 案件列表、详情、状态阶段、冲突复核入口和主体变更重检                                |
| 客户档案       | 受控可用        | 客户主档案、关联方、历史委托、联系人和附件入口                                      |
| 待办中心       | 可用            | 待办筛选、查看、标记完成和延后提醒                                                  |
| 权限边界       | 可用            | 律师访问未授权模块时返回权限或 MVP 范围提示                                         |
| 财务深度流程   | 不在 MVP 主路径 | 保留路由和权限提示，不作为当前试用承诺                                              |
| 文档在线编辑   | 可选            | OnlyOffice 默认关闭，启用时必须配置 URL 与强密钥                                    |
| 缓存与搜索增强 | 可选            | Redis 仅用于显式 cache profile；默认搜索走数据库，Elasticsearch 不是运行依赖        |

## 2. 架构

```text
React 18 + TypeScript 5 + Ant Design 5 + Vite
        │  HTTP / JSON
        ▼
Go 1.25 单体后端（Gin + GORM + domain services）
        │
        ├── PostgreSQL 16：业务、认证会话、审计与冲突权威数据
        ├── 本地文件卷：上传与文档存储
        ├── Redis（可选 cache profile）
        ├── Elasticsearch（可选扩展，默认不部署）
        └── OnlyOffice（可选文档编辑）
```

关键设计：

- PostgreSQL 是唯一生产 bootstrap 数据库；MySQL/SQLite 只代表历史兼容代码路径，不作为当前安装入口。
- 认证会话、刷新令牌轮换、设备撤销和文档锁使用 PostgreSQL 权威存储；Redis 失效或禁用时核心认证与业务路径可继续运行。
- 冲突检查依赖四类权威档案覆盖：案件档案、客户档案、主体名册、关联关系档案。覆盖未确认时，系统保持 `COVERAGE_LIMITED` 并阻断放行。
- Docker 默认栈只包含 migrate、PostgreSQL、backend 和 frontend；`cache` / `observability` 必须显式启用。
- Kubernetes 入口使用 [k8s/README.md](k8s/README.md) 对应的 canonical manifests；仓库内 Helm Chart 属于历史路径，不再作为发布入口。

## 3. 快速开始

### 3.1 环境要求

- Go 1.25+
- Node.js 18+
- Docker 和 Docker Compose（仅容器部署需要）
- PostgreSQL 16（本地容器默认提供；生产建议 15+ 受管数据库）

### 3.2 Docker 默认栈

```bash
git clone https://github.com/VanTam-CN/law-oa-go.git
cd law-oa-go
cp .env.example .env

# Compose 会强制检查安全变量；先按目标环境补齐这些值
# ENVIRONMENT、JWT_SECRET、APP_SECRET、SUBJECT_DATA_KEY
vim .env

docker compose up -d --build
```

默认启动的服务：

| 服务       | 默认地址                           | 说明                      |
| ---------- | ---------------------------------- | ------------------------- |
| frontend   | http://localhost:3003              | Nginx 托管 React 构建产物 |
| backend    | http://localhost:8080              | Go API                    |
| PostgreSQL | localhost:5432                     | 业务与认证会话权威存储    |
| health     | http://localhost:8080/health/ready | 数据库、存储和门禁状态    |

可选 profile：

```bash
# Redis 缓存增强；必须设置强密码并显式接线
REDIS_HOST=redis REDIS_PASSWORD=<strong-password> \
  docker compose --profile cache up -d

# Prometheus / Grafana / Jaeger
docker compose --profile observability up -d
```

Elasticsearch / Kibana 不在默认栈中，也不是当前 MVP 的运行依赖。

### 3.3 本地开发

```bash
# 1. 安装依赖
go mod download
cd frontend && npm install && cd ..

# 2. 准备独立 PostgreSQL，并设置环境变量
#    DB_DRIVER=postgres
#    DB_HOST=127.0.0.1
#    DB_PORT=5432
#    DB_USERNAME=<user>
#    DB_PASSWORD=<password>
#    DB_DATABASE=<database>
#    DB_SSLMODE=disable   # 本地开发；生产必须 require

# 3. 初始化 PostgreSQL schema；不会写入演示业务数据
go run ./cmd/migrate -command bootstrap

# 4. 启动后端
go run .

# 5. 另开终端启动前端，并代理到本地后端
cd frontend
VITE_API_PROXY_TARGET=http://localhost:8080 npm run dev
```

默认后端端口为 8080；Vite 会输出实际前端端口。更完整的配置项见 [docs/CONFIGURATION.md](docs/CONFIGURATION.md)。

### 3.4 非生产 QA 夹具

QA 夹具只允许用于隔离的 PostgreSQL 测试库，不用于生产：

```bash
# 必须显式确认，并设置 QA_FIXTURE_CONFIRM 与 QA_PASSWORD
go run ./cmd/qa-fixture -mode seed
go run ./cmd/qa-fixture -mode verify
```

常用命令也可通过 `make qa-seed-conflict-p0` 和 `make qa-verify-conflict-p0` 调用。

## 4. 验证基线

独立基线最近一次验证日期为 **2026-08-25**：

```bash
# 后端
go test ./internal/auth ./internal/handlers ./internal/router -count=1
go build ./...

# 前端
npm run test -- --runTestsByPath \
  src/services/__tests__/auth.test.ts \
  src/components/layout/__tests__/Header.test.tsx \
  --coverage=false
npm run type-check
npm run lint
npm run build
```

以上全部通过。

另完成一次真实浏览器受控验收：

- 运行形态：真实 Chromium + 独立 PostgreSQL 16 + `REDIS_HOST` 为空。
- 真实 UI 登录律师 A 后点击“退出登录”。
- 浏览器捕获 `POST /api/v1/auth/logout`，请求体为 `{ "token": "<current-access-token>" }`。
- 服务端返回 200，本地 token 与用户信息被清理，页面回到 `/login`。
- 使用退出前保存的旧 token 重放 `GET /api/v1/users/me`，返回 401“令牌已失效”。

证据边界：这是本地受控环境、虚构 QA 数据和真实浏览器操作，不是生产放行证据。

## 5. 生产 NO-GO 边界

即使应用能启动，也必须完成以下事项后才能讨论生产接案：

1. 完成 [conflict-p0-law-firm-trial-spec.md](docs/利益冲突/conflict-p0-law-firm-trial-spec.md) 中的 PD-01 至 PD-07 律所决策物。
2. 导入并核对案件、客户、主体名册、关联关系四类权威档案，登记来源、时效、责任人和零缺口对账证据。
3. 任命独立冲突核查人，并完成律师 A / 律师 B / 核查人的反向越权验收。
4. 使用独立的 `JWT_SECRET`、`APP_SECRET` 和 32 字节 `SUBJECT_DATA_KEY`，配置生产 CORS、TLS、备份和恢复演练。
5. 停用全部演示账号，确认 `/health/ready` 返回 ready。
6. 使用 [production-release-checklist.md](docs/利益冲突/production-release-checklist.md) 留存逐项责任人和证据。

在这些条件完成前，不得将本系统表述为“已上线生产版本”“已通过生产验收”或“可替代律师专业判断”。

## 6. 当前不包含的内容

- 默认 Redis / Elasticsearch / OnlyOffice / Prometheus / Grafana / Jaeger。
- 移动端、协作聊天和企业微信自建应用。
- 财务深度商业化流程。
- 自动导入真实律所档案或自动判断“无冲突”。
- 历史混合 SQL 的生产 up/down 迁移入口；生产只使用 PostgreSQL bootstrap 与备份恢复。

## 7. 常用命令

| 目的                     | 命令                                |
| ------------------------ | ----------------------------------- |
| 构建 Go 应用             | `make build`                        |
| 运行 Go 测试             | `go test ./...`                     |
| Go 静态检查              | `go vet ./...`                      |
| 初始化 PostgreSQL schema | `make migrate-bootstrap`            |
| 写入 QA 冲突夹具         | `make qa-seed-conflict-p0`          |
| 校验 QA 冲突夹具         | `make qa-verify-conflict-p0`        |
| 前端类型检查             | `cd frontend && npm run type-check` |
| 前端 lint                | `cd frontend && npm run lint`       |
| 前端构建                 | `cd frontend && npm run build`      |

## 8. 文档索引

- [docs/CONFIGURATION.md](docs/CONFIGURATION.md)：环境变量和运行配置
- [docs/DEVELOPER_GUIDE.md](docs/DEVELOPER_GUIDE.md)：开发约定
- [docs/API.md](docs/API.md)：API 说明
- [docker-compose.yml](docker-compose.yml)：默认栈与可选 profile
- [k8s/README.md](k8s/README.md)：Kubernetes manifests 入口
- [docs/利益冲突/production-release-checklist.md](docs/利益冲突/production-release-checklist.md)：生产放行检查

## 9. License

ISC License。
