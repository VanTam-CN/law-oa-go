# 示例律师事务所OA

<div align="center">

![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-ISC-green.svg?style=for-the-badge)
![Version](https://img.shields.io/badge/Version-v2.4.0-blue.svg?style=for-the-badge)
![Database](https://img.shields.io/badge/Database-PostgreSQL%20bootstrap-blue.svg?style=for-the-badge)
![React](https://img.shields.io/badge/React-18.2.0-61DAFB?style=for-the-badge&logo=react&logoColor=white)
![TypeScript](https://img.shields.io/badge/TypeScript-5.0.2-3178C6?style=for-the-badge&logo=typescript&logoColor=white)
![Vite](https://img.shields.io/badge/Vite-5.1.0-646CFF?style=for-the-badge&logo=vite&logoColor=white)

**示例律师事务所OA：现代化律师事务所办公自动化系统**

[功能特性](#-功能特性) · [技术架构](#-技术架构) · [快速开始](#-快速开始) · [部署指南](#-部署指南) · [迭代计划](#迭代计划)

</div>

---

## 项目概述

示例律师事务所OA 是一个基于 Go 1.25 构建的现代化律师事务所办公自动化系统，采用单体架构设计，为中小型律师事务所（10-100人）提供完整的数字化解决方案。生产安装入口使用 PostgreSQL schema bootstrap；代码仍保留部分 MySQL/SQLite 兼容路径，但不代表这些方言已经通过当前 P0 生产验收。前端基于 React 18 + TypeScript + Ant Design 5。

### 核心价值

- **高性能**: Go 原生并发 + Redis 多级缓存 + GORM PrepareStmt
- **安全可靠**: JWT + RBAC + bcrypt + 令牌撤销 + 速率限制 + CORS 白名单 + 安全头部 + 客户数据脱敏
- **现代化前端**: React 18 + TypeScript 5.0 + Ant Design 5 + Vite 5 + Redux + React Query
- **生产数据库入口**: PostgreSQL 16 schema bootstrap；MySQL/SQLite 兼容路径需另行完成方言迁移验收
- **可审计部署骨架**: Docker 容器化 + 健康检查 + Prometheus + Grafana + Jaeger；生产放行仍需完成律所政策确认和数据覆盖验收
- **企业级搜索**: Elasticsearch 8.9 全文检索
- **隔离墙机制**: 按案件维度的利益冲突隔离保护

---

## 系统状态

**当前版本**: v2.4.0
**最后更新**: 2026-07-19

**生产放行状态**：当前代码已通过 QA 后端全量测试、前端类型/Lint/构建、
ARM64 生产镜像构建和 P0 冲突隔离回归；正式接案前仍必须完成律所档案覆盖、
PD-01 至 PD-07 决策物、AT-01 至 AT-12 验收和真实 PostgreSQL TLS/部署演练。
生产 Kubernetes 入口为 `k8s/` canonical manifests，仓库中的 Helm Chart 已弃用。
**编译状态**: `go build ./...` 通过
**代码规模**: 265 Go 文件 (101k LOC) + 204 TS/TSX 文件 (73k LOC)
**数据库迁移**: 60 组版本化 migration 文件 + `001_schema_v2.2.0.sql`
**发布口径**: 律师端受控试用 MVP；不等同于未配置即上线的生产版本

### 当前 MVP 口径 (2026-07-16)

当前主线已经收敛到律师端可受控试用 MVP，重点覆盖从接案到冲突复核、审批成案、案件回看和客户档案的核心闭环。冲突检查现在按全所历史档案口径执行，并在档案覆盖未被律所确认、主体身份信息不足或主体变更未完成独立复核时阻断后续动作。演示数据已统一匿名化，公开仓库不包含真实律所名称、真实客户域名或个人账号。

> **重要**：本仓库当前不是“开箱即用的生产合规系统”。正式上线前必须完成 `docs/利益冲突/conflict-p0-law-firm-trial-spec.md` 中 PD-01 至 PD-07 的律所书面决策、权威档案导入/覆盖确认、冲突核查人指定、敏感主体密钥托管和真实账号隔离验收。没有完整覆盖配置时，系统会保持 `COVERAGE_LIMITED`，不允许以“无冲突”或“已通过”放行案件。

| 场景 | 状态 | 说明 |
|------|------|------|
| 登录与工作台 | ✅ | 支持匿名演示账号登录，工作台展示待办、冲突复核、审批和案件入口 |
| 新建立案 | ✅ | 基本信息、客户、对方当事人、负责律师、费用字段和保存/提交流程可用 |
| 利益冲突检查 | ⚠️ 受门禁控制 | 默认进入检测清单；按全所历史档案检索；覆盖不完整时明确阻断，不显示“无冲突”放行 |
| 冲突审核与审批 | ⚠️ 受门禁控制 | 需要独立冲突核查人复核；直接冲突不可用豁免流程绕过 |
| 审批通过成案 | ⚠️ 受门禁控制 | 审批通过后自动创建并回填正式案件 ID；关联案件和对外动作仍受主体版本门禁约束 |
| 案件管理 | ✅ 受控可用 | 列表、详情、状态阶段、下一步操作和关联冲突复核入口可用；主体变更必须重新复核 |
| 客户档案 | ✅ | 客户主档案、关联方、历史委托、联系人、附件入口和快捷操作可用 |
| 权限边界 | ✅ | 律师访问财务/用户管理等非授权模块时显示权限或 MVP 范围提示 |
| 收件箱/财务深度流程 | 🚧 | 不是当前律师端 MVP 主路径，保留路由和权限提示 |

### 模块完成度

| 模块 | 完成度 | 后端 | 前端 | 说明 |
|------|--------|------|------|------|
| 认证授权 | 100% | ✅ | ✅ | JWT + RBAC + 令牌撤销 + 设备管理 |
| 用户管理 | 95% | ✅ | ✅ | CRUD + 头像 + 权限 + 资料 |
| 客户管理 | 95% | ✅ | ✅ | 档案 + 分类 + 搜索 + 统计 + PII脱敏 |
| 案件管理 | 90% | ✅ | ✅ | 基础版 + 增强版 + 状态跟踪 + 律师分配 |
| 审批流程 | 95% | ✅ | ✅ | 多级审批 + 状态机 + 条件分支 + 冲突集成 + 代理审批 + 超时处理 |
| 财务管理 | 92% | ✅ | ✅ | 合同 + 发票 + 支付 + 佣金 + 坏账 + 费率模板 + 费用测算 + 分成规则 |
| 冲突检测 | 92% | ✅ | ✅ | Entity系统 + v2增强 + 3层递归穿透 + 风险评估 + 审核流程 + 事件Hook |
| 隔离墙 | 92% | ✅ | ✅ | 启用/禁用 + 白名单CRUD + 访问日志 + 中间件保护 |
| 搜索功能 | 90% | ✅ | ✅ | ES全文检索 + 法条搜索 + 智能排序 |
| 文档管理 | 72% | ✅ | ✅ | 上传/下载/预览 + OnlyOffice在线编辑 + 版本控制 + 回收站 + 卷宗目录模板 |
| 通知系统 | 85% | ✅ | ✅ | 通知队列 + 模板管理 + 审批流程 + 批量操作 |
| 信任账户 | 85% | ✅ | ✅ | 账户管理 + 交易记录 + 余额监控 |
| 内容过滤 | 90% | ✅ | ✅ | 敏感词管理 + 内容检测 + 过滤日志 |
| 统计报表 | 85% | ✅ | ✅ | ECharts + Recharts 图表 + 导出 |
| 待办事项 | 82% | ✅ | ✅ | 收件箱管理 + 定时提醒 + 超时升级 |
| 工具模块 | 80% | ✅ | ✅ | 诉讼费/利息/截止日期计算器 |
| 离职流程 | 60% | ✅ | ✅ | 基础框架，待完善 |
| 协作聊天 | 0% | ❌ | ❌ | 计划中，优先用企业微信代替 |
| 移动端 | 0% | ❌ | ❌ | Sprint 2: 企业微信自建应用 |

### Sprint 1 已完成功能清单 (v2.4.0)

| 任务 | 状态 | 关键文件 |
|------|------|----------|
| 冲突关联穿透(递归3层) | ✅ | `entity_repository.go`, `conflict_hook_service.go` |
| 案件全生命周期(7阶段状态机) | ✅ | `case_state_machine.go`, `case_status.go` |
| 诉讼时效引擎(T-7/T-3/T-0预警) | ✅ | `deadline_calculator.go`, `deadline_handler.go` |
| 审批评论与附件 | ✅ | `approval_service.go` |
| 代理审批配置 | ✅ | `approval_delegation.go/.go` (model/service/repo/handler 全链路) |
| 审批超时处理 | ✅ | `scheduler_service.go` (48h超时/升级/预警) |
| ONLYOFFICE集成 | ✅ | `onlyoffice_handler.go` (回调HMAC + 转换API) |
| 文档版本控制 | ✅ | `document_version_service.go`, `document_lock_service.go` |
| 卷宗目录模板 | ✅ | `folder_template_service.go`, `case_folder.go` (递归创建+树构建) |
| 审批条件分支(规则引擎) | ✅ | `approval_template_service.go`, `approval_assigner.go` |
| 审批业务联动Hook | ✅ | `approval_business_hooks.go`, `approval_conflict_integration_service.go` |
| 冲突审核流程 | ✅ | `conflict_review_handler.go`, `conflict_hook_service.go` |
| 分成规则引擎 | ✅ | `commission_rule_repository.go` (4种计费方式) |
| 安全加固(4项) | ✅ | `models.go`(PII脱敏) + `types.go`(速率限制/CORS/安全头部) |
| Gemini代码审查修复(36项) | ✅ | CRITICAL/HIGH/MEDIUM 全部修复 |

### 生产放行状态

| 检查项 | 状态 | 说明 |
|--------|------|------|
| Go 编译 | ✅ | `go build ./...` 零错误 |
| Docker 构建 | ⚠️ 需目标构建机复核 | 多阶段构建 + 健康检查；生产镜像不再携带 `.env` 模板；当前环境尚未完成基础镜像拉取后的实际构建 |
| CORS 白名单 | ✅ | 生产启动拒绝本地/通配来源，使用 `CORS_ALLOWED_ORIGINS` |
| 速率限制 | ✅ | 滑动窗口 100 req/min per IP |
| 安全头部 | ✅ | CSP/HSTS/COOP/COEP |
| JWT / 应用密钥 | ✅ | 生产启动要求非默认强密钥 |
| 主体身份加密 | ✅ | 生产启动要求独立的 `SUBJECT_DATA_KEY`，不得与 JWT 复用 |
| P0 档案覆盖 | ⛔ 必须配置 | 四类 `conflict_search_scopes` 必须 ACTIVE + COMPLETE，并绑定零缺口的 `conflict_index_build_runs`；否则 readiness 失败 |
| 主体变更重检 | ✅ 代码门禁 | 主体版本、独立复核、覆盖状态和并发版本检查由服务端强制 |
| 文档对外输出 | ✅ 代码门禁 | 案件文档下载/预览/转换与主体有效版本绑定 |
| 权威档案与政策 | ⛔ 待律所确认 | PD-01 至 PD-07 不是代码可以替代的决策，未完成不得宣称生产放行 |
| IDOR / 隔离墙 | ⚠️ 必须实测 | 需要使用 A/B/冲突核查人真实权限数据完成反向越权验收 |

---

## 功能特性

### 认证与授权
- JWT 令牌认证 (golang-jwt/v5)
- 用户注册、登录、登出
- 令牌撤销：设备管理、用户撤销、全量撤销
- RBAC 角色权限控制 + AdminAuthMiddleware
- 活跃设备查询 + 撤销历史

### 案件管理
- 基础版 + 增强版双模式
- **7阶段状态机**: 线索 → 风控 → 签约 → 办案 → 庭审 → 结案 → 归档
- 案件状态流转跟踪
- 律师分配 + 团队协作
- 案件文档关联 + 费用管理

### 冲突检测系统
- **Entity实体管理**: 个人/企业法律实体，关联关系，名称变更历史
- **3层递归穿透**: PostgreSQL Recursive CTE 实现母→子→孙公司穿透
- **事件驱动Hook**: 案件新增/变更当事人自动触发二次检查
- 多维度检测（律师利益/客户关系/行业竞争）
- 智能风险评估（CRITICAL/HIGH/MEDIUM/LOW）
- **v2增强版**: Entity驱动 + 审核流程 + PDF报告
- 检测统计分析

### 审批流程管理
- 多级审批 + 状态机（待提交→审批中→通过/驳回/取消/过期）
- **条件分支规则引擎**: 基于JSON配置（标的额/案件类型/风险等级路由）
- **代理审批**: 循环检测 + 递归代理链（最大5层）
- **审批超时处理**: 48h自动升级 + 80%预警 + 过期标记
- **业务联动Hook**: 审批通过后自动触发后续操作
- 审批评论与附件支持
- 冲突检测集成（审批自动触发冲突检查）

### 财务管理系统
- **合同管理**: 列表/创建/编辑/详情/状态跟踪
- **发票管理**: 开具/跟踪/统计
- **支付管理**: 支付记录/状态跟踪
- **佣金报告**: 4种计费方式（按时/固定/混合/顾问）+ 绩效奖金 + 成本扣除
- **分成规则引擎**: 规则 CRUD + 数据一致性校验
- **坏账核销**: 管理 + 记录
- **费率模板**: 按案件类型（诉讼/非诉讼/咨询）配置
- **费用测算**: 基于模板实时计算预估费用

### 文档管理
- 上传/下载/预览/搜索/过滤
- **OnlyOffice 在线编辑**: Docker 部署 + HMAC 回调 + 转换 API + 锁机制
- **版本控制**: 覆盖上传自动版本 + 版本锁定 + 回滚
- **回收站**: 软删除 + 恢复 + 永久删除
- **卷宗目录模板**: 预设诉讼/非诉标准目录，递归创建树形结构
- 文档统计（存储使用、用户活动）
- > 待完成：权限管理（细粒度）、高级搜索（全文）

### 诉讼时效引擎
- 判决书日期 → 自动计算上诉期(15天)/执行时效(2年)
- **T-7/T-3/T-0 三级预警**
- 截止日期计算器工具

### 隔离墙机制
- 按案件维度启用/禁用
- 白名单 CRUD（添加/移除/查询）
- 访问日志记录
- API 中间件自动拦截

### 信任账户管理
- 账户创建/管理
- 交易记录跟踪
- 余额监控 + 最低余额告警中间件

### 通知系统
- 通知队列管理
- 模板系统（创建/编辑/启用/禁用）
- 通知审批流程
- 批量操作

### 内容过滤系统
- 敏感词管理（CRUD + 批量导入）
- 内容检测与过滤
- 过滤日志 + 统计分析
- 缓存管理

### 定时调度服务
- 提醒检查（每小时）
- 待办升级检查（每6小时）
- 审批超时检查（每30分钟）

### 工具模块
- 诉讼费计算器
- 利息计算器
- 截止日期计算器
- 法律搜索

---

## 技术架构

### 后端技术栈

| 组件 | 版本 | 用途 |
|------|------|------|
| Go | 1.25 | 主语言 |
| Gin | v1.10.1 | Web 框架 |
| GORM | v1.30.0 | ORM |
| PostgreSQL | 15+ | 当前生产 bootstrap 数据库 |
| MySQL | 8.0 | 代码兼容路径，当前不作为生产安装入口 |
| SQLite | 3 | 开发数据库 |
| Redis | v9.0.5 | 缓存 + 会话 |
| Elasticsearch | 8.9 | 全文搜索 |
| JWT | v5.0.0 | 认证 |
| Zap | v1.24.0 | 结构化日志 |
| Prometheus | v1.16.0 | 指标监控 |
| Jaeger | v1.17.0 | 分布式追踪 |
| Swagger | v1.5.3 | API 文档 |

### 前端技术栈

| 组件 | 版本 | 用途 |
|------|------|------|
| React | 18.2.0 | UI 框架 |
| TypeScript | 5.0.2 | 类型安全 |
| Vite | 5.1.0 | 构建工具 |
| Ant Design | 5.16.1 | UI 组件库 |
| Redux Toolkit | 2.9.1 | 状态管理 |
| React Query | 5.90.5 | 服务端状态 |
| Zustand | 5.0.8 | 轻量状态 |
| React Router | 7.9.4 | 路由 |
| Axios | 1.12.2 | HTTP 客户端 |
| ECharts | 5.6.0 | 图表 |
| Jest + Vitest | 30/3.2 | 测试 |

### 架构设计

```
┌─────────────────────────────────────────────────────────────────────┐
│                     Frontend (React 18 + TS 5)                      │
│           Ant Design 5 · Redux · React Query · Vite 5               │
├─────────────────────────────────────────────────────────────────────┤
│                        API Gateway (Nginx)                       │
│          JWT Auth · Rate Limit · CORS · Security Headers · Ethical  │
├─────────────────────────────────────────────────────────────────────┤
│                       Business Logic (Go)                            │
│  ┌──────────┬──────────┬──────────┬──────────┬──────────────────┐  │
│  │Auth/User │  Case    │Approval │Conflict │ Finance           │  │
│  │          │ Mgmt     │  Flow   │ Detect  │ (Contract/Invoice │  │
│  │          │          │  Engine  │ Engine  │  /Payment/Comm)   │  │
│  ├──────────┼──────────┼──────────┼──────────┼──────────────────┤  │
│  │Document │ Trust    │ Notif   │Content  │ Ethical Wall      │  │
│  │(OO Edit) │ Account  │  Queue  │ Filter  │ + Inbox/Scheduler │  │
│  ├──────────┼──────────┼──────────┼──────────┼──────────────────┤  │
│  │Entity   │Deadline │Folder   │ Team    │ Analytics          │  │
│  │System   │ Engine  │ Template│ Perm    │ + Search/Stats    │  │
│  └──────────┴──────────┴──────────┴──────────┴──────────────────┘  │
├─────────────────────────────────────────────────────────────────────┤
│                        Data Layer                                │
│  ┌──────────┬──────────┬──────────┬──────────┬──────────────┐     │
│  │PostgreSQL│ MySQL   │ Redis   │Elasticsearch│ Local Files│     │
│  │ Primary  │ Support │ Cache   │ 8.9 Search│ /Uploads   │     │
│  └──────────┴──────────┴──────────┴──────────┴──────────────┘     │
├─────────────────────────────────────────────────────────────────────┤
│                     Observability                                 │
│  ┌──────────┬──────────┬──────────┬──────────────────────────┐    │
│  │Prometheus│ Grafana  │ Jaeger  │ Zap Structured Logs    │    │
│  │ Metrics  │Dashboard│ Tracing │ + Lumberjack Rotation  │    │
│  └──────────┴──────────┴──────────┴──────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────┘
```

### 代码分层

```
internal/
├── handlers/         # 32 个 HTTP 处理器（API 入口）
├── services/         # 75 个业务服务（核心逻辑）
├── repositories/     # 42 个数据仓储（数据访问）
├── models/           # 22 个数据模型（结构定义）
├── middleware/       # 18 个中间件（认证/日志/安全/隔离墙）
├── security/         # 安全模块（速率限制/CORS/安全头部）
├── auth/             # 认证模块（JWT/令牌撤销/上下文）
├── router/           # 路由注册（259 个端点）
├── config/           # 配置管理（Viper + 环境变量）
├── cache/            # 缓存模块
├── metrics/          # Prometheus 指标
└── tracing/          # OpenTelemetry 追踪
```

---

## 快速开始

### 环境要求
- Go 1.25+ / Node.js 18+ / Docker & Docker Compose
- Docker Compose 当前编排 PostgreSQL 16、Redis、Elasticsearch 等服务；生产 schema 通过一次性 `bootstrap` 初始化，历史混合 SQL 迁移目录不作为生产入口
- Redis 7+ / Elasticsearch 8.9（可选）

### 一键启动

```bash
git clone https://github.com/VanTam-CN/law-oa-go.git
cd law-oa-go

# 配置环境变量
cp .env.example .env
mkdir -p ./data/{postgres,redis,elasticsearch,prometheus,grafana,uploads,logs}

# Docker 启动（PostgreSQL + Redis + ES；首次启动会执行一次 schema bootstrap）
docker compose up -d

# 或本地开发
go run .                        # 后端 :8080
cd frontend && npm run dev      # 前端 :3003
```

开发环境可以使用本地 `.env`。生产环境必须显式提供 `ENVIRONMENT=production`、`JWT_SECRET`、`APP_SECRET`、`SUBJECT_DATA_KEY`、数据库密钥和真实前端 CORS 来源；主体密钥必须是独立的 32 字节密钥。生产后端会在缺少这些条件时拒绝启动，健康检查还会在档案覆盖未完成时返回未就绪。

生产上线前的最小顺序（代码已提供可运行的 PostgreSQL bootstrap，但仍不等于已完成律所生产放行）：

1. 完成并签署 `docs/利益冲突/conflict-p0-law-firm-trial-spec.md` 的 PD-01 至 PD-07 决策物。
2. 准备 PostgreSQL 15+ 空库或备份副本，配置 `DB_DRIVER=postgres`、`DB_SSLMODE=require` 及生产密钥；执行 `docker compose up -d --build`，确认 `migrate` 日志显示 bootstrap 成功，再检查 `GET /health/ready`。生产不执行历史混合 SQL 的 `up/down`。
3. 先运行 `go run ./cmd/backfill-sensitive-identities` 盘点历史明文，完成数据库备份和审批后再运行 `go run ./cmd/backfill-sensitive-identities --apply`，确认复核结果为零。
4. 先运行只读 `go run ./cmd/backfill-conflict-index` 盘点四类主体索引；完成档案核对和备份后，再以冲突核查岗身份运行 `go run ./cmd/backfill-conflict-index --apply --actor-id <用户ID> --evidence-reference <凭证引用>`，保留四个 `run_id`、数量和哈希。
5. 由冲突核查岗通过 `GET/PUT /api/v1/conflict-v2/search-scopes` 登记案件、客户、主体、关系四类权威档案来源，并在 COMPLETE 请求中绑定对应 `index_run_id`；版本号和凭证必须与运行记录一致。
6. 删除或停用所有 `@example.test` / `demo_*` 演示账号；生产 readiness 会主动阻断仍启用的演示账号。
7. 确认 `GET /health/ready` 返回 `ready: true`，并使用律师 A、律师 B、冲突核查人三个独立账号完成反向越权验收。
8. 只使用脱敏或虚构种子数据完成验收；不要将真实客户、真实身份证号或真实统一社会信用代码放入公开仓库或演示数据库。

### 访问地址

| 服务 | 地址 |
|------|------|
| 前端应用 | http://localhost:3003 |
| 后端 API | http://localhost:8080 |
| API 文档 | http://localhost:8080/swagger/index.html |
| 健康检查 | http://localhost:8080/api/v1/health |
| Grafana | http://localhost:3000 |
| Prometheus | http://localhost:9090 |
| Jaeger | http://localhost:16686 |
| OnlyOffice | 按 `ONLYOFFICE_URL` 配置，默认回调集成使用 http://localhost:9090 |

### 匿名演示账号（仅本地/验收环境）

仓库中的历史试用 SQL 会写入匿名演示账号和演示业务数据，但该目录包含多个历史数据库方言，已被生产迁移前置检查明确禁止作为当前安装入口。不要在生产数据库执行；需要律师验收时，应使用已审核的非生产 PostgreSQL 快照或单独的脱敏 fixture 流程：

```bash
# 当前生产入口只初始化结构，不写入演示数据
go run ./cmd/migrate -command bootstrap
```

| 角色 | 邮箱 | 密码 |
|------|------|------|
| 管理员 | `demo.admin@example.test` | `Demo@2026` |
| 律师 | `demo.lawyer@example.test` | `Demo@2026` |
| 助理 | `demo.assistant@example.test` | `Demo@2026` |
| 财务 | `demo.finance@example.test` | `Demo@2026` |

如果使用历史试用数据，必须在隔离的测试数据库中按对应历史版本单独验证方言和数据脚本，并在验收后销毁或重置；**这些账号不能用于生产**，生产健康检查会拒绝仍启用的演示账号。

### 当前推荐验证命令

```bash
go build ./...
cd frontend && npm run type-check
cd frontend && npm run test:e2e -- case-create-full-workflow.spec.ts
```

---

## API 文档

### 主要端点（259个）

#### 公开路由
- `GET /api/v1/health` - 健康检查
- `POST /api/v1/auth/login` - 登录
- `POST /api/v1/auth/register` - 注册
- `GET /api/v1/legal/statutes/search` - 法条搜索

#### 认证路由
- `POST /api/v1/auth/logout` - 登出
- `POST /api/v1/auth/revoke/user` - 撤销用户令牌
- `POST /api/v1/auth/revoke/device` - 撤销设备令牌
- `GET /api/v1/auth/devices/:user_id` - 活跃设备

#### 审批流程（含代理 + 超时）
- `GET/POST /api/v1/approvals` - 审批列表/创建
- `POST /api/v1/approvals/:id/submit` - 提交
- `POST /api/v1/approvals/:id/decision` - 审批决定
- `GET/POST/PUT /api/v1/approval-templates` - 审批模板 CRUD
- `GET/POST/DELETE /api/v1/approvals/delegations` - 代理审批配置
- `GET /api/v1/approvals/delegations/my` - 我的代理配置

#### 冲突检测（P0 规范路径）
- `POST /api/v1/conflict-v2/entities` - 创建法律实体
- `POST /api/v2/entities/:id/relations` - 添加关联关系
- `POST /api/v1/conflict/check` - 按案件主体执行全量冲突检测（生产唯一检测入口）
- `GET /api/v1/conflict-v2/search-scopes` - 查看权威档案覆盖登记（仅冲突核查岗/管理角色）
- `PUT /api/v1/conflict-v2/search-scopes/:id` - 登记或更新档案来源覆盖与核对凭证（仅冲突核查岗/管理角色）

> `/api/v1/conflict-v2/checks` 及其详情、报告、统计接口属于旧版兼容服务。为避免形成第二套冲突结论路径，生产环境统一返回 `MVP_MODULE_UNAVAILABLE`；本地开发环境可用于兼容性调试，但不得作为上线验收依据。

#### 财务管理
- `GET/POST /api/v1/finance/contracts` - 合同
- `GET/POST /api/v1/finance/invoices` - 发票
- `GET /api/v1/finance/payments` - 支付
- `GET /api/v1/finance/commissions` - 佣金报告
- `GET/POST /api/v1/finance/fee-templates` - 费率模板
- `POST /api/v1/finance/fee-templates/calculate` - 费用测算
- `GET/POST /api/v1/finance/commission-rules` - 分成规则

#### 文档管理（OnlyOffice + 版本 + 卷宗）
- `GET/POST /api/v1/documents` - 文档列表/上传
- `GET/PUT/DELETE /api/v1/documents/:id` - 文档 CRUD
- `GET /api/v1/documents/:id/download` - 下载
- `GET /api/v1/documents/:id/preview` - 预览
- `GET /api/v1/documents/:id/versions` - 版本历史
- `POST /api/v1/documents/:id/versions/:versionId/restore` - 版本恢复
- `GET /api/v1/documents/recycle` - 回收站
- `GET/POST /api/v1/folder-templates` - 卷宗模板
- `POST /api/v1/folder-templates/:id/apply` - 应用模板

#### 隔离墙
- `POST /api/v1/cases/:id/ethical-wall/enable|disable` - 启用/禁用
- `GET/POST/DELETE /api/v1/cases/:id/ethical-wall/whitelist` - 白名单
- `GET /api/v1/cases/:id/ethical-wall/logs` - 访问日志

#### 诉讼时效
- `POST /api/v1/deadline/calculate` - 时效计算
- `GET /api/v1/deadline/cases` - 预警案件列表

---

## 迭代计划

### Sprint 1: 基础能力增强 ✅ 已完成 (v2.4.0)

- [x] 冲突关联穿透(递归3层)
- [x] 案件全生命周期(7阶段状态机)
- [x] 诉讼时效引擎(T-7/T-3/T-0预警)
- [x] 审批评论/附件/代理/条件分支/超时处理
- [x] ONLYOFFICE 在线编辑集成
- [x] 文档版本控制 + 回收站
- [x] 卷宗目录模板
- [x] 分成规则引擎
- [x] 安全加固(PII脱敏/速率限制/CORS/安全头部)
- [x] 代码质量修复(36项Gemini审查)

### Sprint 2: 核心业务闭环 (Week 5-8) 🔜 待开始

- [ ] **企业微信自建应用**: AgentId/Secret 配置 + OAuth2 免登
- [ ] **消息推送**: 审批待办/开庭提醒推送到企微
- [ ] **移动端 H5**: React + antd-mobile 微应用（审批/冲突快查/搜索）
- [ ] **计时收费**: 网页/移动端计时器 + 工时单录入
- [ ] **回款登记**: 银行流水认领 + 关联合同/发票
- [ ] **核心三表**: 创收贡献表、应收账龄表、律师提成明细表
- [ ] **案件可视化时间轴**: 关键法律节点 + 颜色编码

### Sprint 3: 差异化与商业化 (Week 9-12) 📋 规划中

- [ ] **天眼查集成**: 公司名自动填充工商信息 + 冲突数据补充
- [ ] **PII 数据加密**: AES-GCM 加密存储身份证/手机号
- [ ] **文档模板库**: 全所共享合同/诉状模板 + 标签分类
- [ ] **知识库自动归档**: 结案时文书自动同步
- [ ] **动态水印**: 文档预览平铺用户名+IP+时间
- [ ] **电子签章**: 集成 e签宝/法大大 API
- [ ] **简单 OCR**: 证据图片文字提取
- [ ] **可视化驾驶舱**: 律所经营概览
- [ ] **Docker 一键部署优化**: 简化安装到 30 分钟

### YAGNI 原则：坚决不做

| 功能 | 理由 |
|------|------|
| 复杂 LLM 私有化部署 | 1-2人团队撑不起 GPU，仅做 API 调用 |
| 完善的 HRM 系统 | 律所人事复杂，先用 Excel |
| 财务全科目凭证 | OA 不是 ERP，不抢金蝶/用友 |
| 社交化协作(类钉钉群聊) | 直接用微信/企微 |
| 原生 App(iOS/Android) | 企业微信 H5 + PWA 足够 |
| 国际化(多语言) | 中小律所无此需求 |

详细迭代计划见 [docs/iteration-plan.md](docs/iteration-plan.md)。

---

## 更新日志

### v2.4.0 (2026-04-08)

**Sprint 1 完成 - 基础能力增强**

后端 (15 项新功能):
- **冲突关联穿透**: PostgreSQL Recursive CTE 实现 3 层 Entity 递归穿透
- **案件状态机**: 7 阶段完整生命周期（线索→风控→签约→办案→庭审→结案→归档）
- **诉讼时效引擎**: T-7/T-3/T-0 三级预警，截止日期自动计算
- **代理审批**: 循环检测 + 递归代理链(5层) + IDOR 防护
- **审批超时处理**: 48h 自动升级 + 80% 预警 + 过期标记
- **ONLYOFFICE 集成**: HMAC 回调验证 + 文档转换 API + 锁机制
- **文档版本控制**: 覆盖上传版本 + 版本锁定/解锁 + 回滚
- **卷宗目录模板**: 递归创建树形文件夹 + 应用模板
- **分成规则引擎**: 按案件类型/阶段配置，4种计费方式
- **审批条件分支**: JSON 规则引擎 + 自动分配器
- **审批业务联动**: Hook 机制（冲突检测/状态变更联动）
- **冲突审核流程**: 提交审核 + 通过/拒绝 + 记录

安全加固 (4 项):
- **PII 脱敏**: 身份证/手机号 `json:"-"` 不返回明文
- **速率限制**: 滑动窗口 100 req/min per IP
- **CORS 白名单**: 环境变量配置
- **安全头部**: CSP/HSTS/COOP/COEP 中间件

代码质量 (36 项 Gemini 审查修复):
- CRITICAL: 自代理 CHECK 约束、并发创建唯一约束
- HIGH: IDOR 权限(列表/撤销)、parseWorkflowTimeouts 实现、strconv 替换 Sscanf、ReleaseLock 错误日志、默认模板原子性
- MEDIUM: GetFolderPath N+1→递归 CTE、DocumentCount LEFT JOIN、硬编码阈值→常量、注释路由清理

### v2.3.0 (2026-04-04)
- Entity 实体系统 + 冲突检测增强 + 财务闭环 + 隔离墙增强 + 审批冲突集成

### v2.2.0 (2026-02-12)
- 财务管理 + 信任账户 + 通知系统 + 内容过滤 + 隔离墙 + 令牌撤销 + 待办事项

### v2.1.0 (2026-02-09)
- 冲突检测系统 + 审批流程 + 文档管理 + 搜索优化

### v2.0.0 (2025-09-28)
- 架构重构 + UI 现代化 + 安全增强

---

## 贡献指南

### 提交规范

```
feat: 新功能
fix: 修复问题
docs: 文档更新
refactor: 代码重构
test: 测试相关
chore: 构建/辅助工具
perf: 性能优化
ci: CI/CD 相关
```

### 开发规范

- Go: `gofmt` + `golangci-lint` + Swagger 注释
- 前端: TypeScript 严格模式 + ESLint + Prettier + Ant Design 规范
- 提交前: `go build ./...` + `npm run build`

---

## 许可证

ISC License - [LICENSE](LICENSE)

## 联系方式

- **维护者**: VanTam-CN
- **仓库**: [GitHub](https://github.com/VanTam-CN/law-oa-go)
- **Issues**: [GitHub Issues](https://github.com/VanTam-CN/law-oa-go/issues)
