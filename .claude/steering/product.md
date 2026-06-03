# Product Steering Document

## 产品概述

Law OA Go 是一个基于 Go 1.23+ 构建的现代化律师事务所办公自动化系统，采用单体架构设计，为中小型律师事务所提供完整的数字化解决方案。

**当前版本**: v2.4.0
**最后更新**: 2026-04-08
**目标客户**: 10-100 人规模中小型律师事务所

## 核心价值主张

### 目标用户
- 中小型律师事务所（10-100人规模）
- 法律从业者（律师、助理、管理人员）
- 需要数字化案件管理的法律机构

### 核心价值
- **高性能**: Go 原生并发 + Redis 缓存 + GORM PrepareStmt
- **安全可靠**: JWT + RBAC + bcrypt + 令牌撤销 + 速率限制 + PII脱敏 + 安全头部
- **现代化前端**: React 18 + TypeScript 5.0 + Ant Design 5 + Vite 5
- **数据库灵活**: PostgreSQL / MySQL / SQLite 三库自适应
- **易维护**: 265 Go 文件(101k LOC) + 204 TS/TSX(73k LOC)，清晰分层
- **生产就绪**: Docker + 健康检查 + Prometheus + Grafana + Jaeger
- **企业级搜索**: Elasticsearch 8.9 全文检索
- **隔离墙机制**: 按案件维度的利益冲突隔离保护
- **在线编辑**: OnlyOffice 集成，浏览器内编辑 Docx

## 核心功能模块

### 已完成功能（完成度 >= 80%）

| 模块 | 完成度 | 关键能力 |
|------|--------|----------|
| 认证授权 | 100% | JWT + RBAC + 令牌撤销 + 设备管理 |
| 用户管理 | 95% | CRUD + 头像 + 权限 + PII脱敏 |
| 客户管理 | 95% | 档案 + 分类 + 搜索 + 案件关联 |
| 案件管理 | 90% | 7阶段状态机 + 增强版 + 团队协作 |
| 审批流程 | 95% | 多级 + 条件分支 + 代理审批 + 超时处理 + 业务Hook |
| 财务管理 | 92% | 合同/发票/支付/佣金/坏账/费率模板/费用测算/分成规则 |
| 冲突检测 | 92% | Entity系统 + v2增强 + 3层穿透 + 审核 + 事件Hook |
| 隔离墙 | 92% | 启用/禁用 + 白名单 + 访问日志 + 中间件 |
| 搜索功能 | 90% | ES全文检索 + 法条搜索 + 智能排序 |
| 通知系统 | 85% | 通知队列 + 模板 + 审批流程 + 批量操作 |
| 信任账户 | 85% | 账户管理 + 交易记录 + 余额监控 |
| 内容过滤 | 90% | 敏感词管理 + 内容检测 + 过滤日志 |
| 统计报表 | 85% | ECharts + Recharts + 导出 |
| 待办事项 | 82% | 收件箱 + 定时提醒 + 超时升级 |
| 文档管理 | 72% | 上传/下载/预览 + OnlyOffice编辑 + 版本控制 + 回收站 + 卷宗模板 |
| 工具模块 | 80% | 诉讼费/利息/截止日期计算器 |

### 部分完成功能

| 模块 | 完成度 | 缺失部分 |
|------|--------|----------|
| 离职流程 | 60% | 权限交接、文档移交、审批流未完善 |
| 文档管理 | 72% | 细粒度权限管理、高级全文搜索未完成 |

### 未开发功能

| 模块 | 计划 | Sprint |
|------|------|--------|
| 企业微信移动端 | 自建应用 + H5 审批/冲突快查 | Sprint 2 |
| 计时收费 | 网页/移动端计时器 + 工时审核 | Sprint 2 |
| 回款登记 | 银行流水认领 + 三表 | Sprint 2 |
| 天眼查集成 | 工商信息自动填充 | Sprint 3 |
| PII 加密 | AES-GCM 加密存储 | Sprint 3 |
| 电子签章 | e签宝/法大大 API | Sprint 3 |
| 知识库 | 结案归档 + 模板库 | Sprint 3 |

## Sprint 1 完成清单 (v2.4.0)

### 后端新增/修改文件

**新建文件 (12个)**:
- `internal/models/approval_delegation.go` - 代理审批模型
- `internal/models/case_folder.go` - 案件文件夹模型
- `internal/repositories/approval_delegation_repository.go` - 代理审批仓储
- `internal/repositories/folder_template_repository.go` - 卷宗模板仓储
- `internal/services/approval_delegation_service.go` - 代理审批服务(循环检测+递归链)
- `internal/services/conflict_hook_service.go` - 冲突事件Hook
- `internal/services/folder_template_service.go` - 卷宗模板服务(递归创建)
- `internal/services/deadline_calculator.go` - 诉讼时效引擎
- `internal/handlers/approval_delegation_handler.go` - 代理审批API
- `internal/handlers/folder_template_handler.go` - 卷宗模板API
- `internal/handlers/deadline_handler.go` - 时效计算API
- `migrations/000024_approval_delegation.up.sql` - 代理审批表
- `migrations/000025_case_folders.up.sql` - 卷宗文件夹表

**重要修改文件**:
- `internal/services/scheduler_service.go` - 审批超时检查 + 常量化
- `internal/models/approval_models.go` - 超时/升级字段
- `internal/router/router.go` - 新路由 + 注释路由清理
- `internal/models/models.go` - PII 脱敏(json:"-")
- `internal/security/types.go` - 速率限制/CORS/安全头部激活

### 安全加固

| 项 | 措施 | 文件 |
|----|------|------|
| PII 脱敏 | 身份证/手机号 API 不返回明文 | `models.go` |
| 速率限制 | 滑动窗口 100 req/min per IP | `cmd/server/main.go` |
| CORS | 环境变量白名单 | `types.go` |
| 安全头部 | CSP/HSTS/COOP/COEP | `types.go` |
| IDOR | 代理列表/撤销权限校验 | `approval_delegation_handler.go` |
| 自代理约束 | CHECK(delegator != delegate) | `000024.up.sql` |
| 递归限制 | 冲突穿透3层 + 代理链5层 | `conflict_hook_service.go` |
| 超时常量 | 48h/3天/80% 提取为常量 | `scheduler_service.go` |

## 业务规则

### 权限管理
- RBAC 角色控制（admin/user）
- AdminAuthMiddleware 管理员专属路由
- 代理审批：管理员/委托人/创建者三重校验

### 审批规则
- 状态机：draft → submitted → under_review → approved/rejected/cancelled/expired
- 超时：48h（可按工作流配置）自动升级
- 预警：80% 时间时发送提醒
- 代理：递归查找有效代理，最大5层，防循环

### 冲突检测规则
- 递归穿透：3层 Entity 关联关系
- 事件驱动：案件新增/变更当事人自动触发
- 风险等级：CRITICAL/HIGH/MEDIUM/LOW 四级
- 审核流程：检测→提交→审核→通过/拒绝

### 数据库
- 25 个 migration 文件
- 43+ 张表（GORM AutoMigrate 16 张 + migration 手动创建）
- 支持 PostgreSQL 递归 CTE

## 技术特色

### 架构
- 单体应用，分层架构：handlers(32) → services(75) → repositories(42) → models(22)
- 中间件链：Auth → RateLimit → CORS → SecurityHeaders → EthicalWall → Logging → Metrics
- 定时调度：3 个 goroutine（提醒/升级/超时检查）

### 性能
- GORM PrepareStmt 预编译缓存
- SkipDefaultTransaction 跳过默认事务
- 滑动窗口速率限制器（sync.Map + 时间窗口）
- GetFolderPath 递归 CTE 替代 N+1 查询
- DocumentCount LEFT JOIN 统计

### 安全
- JWT 无状态认证 + 令牌撤销
- bcrypt 密码加密
- PII 数据 API 层脱敏
- 输入验证（validator）
- 审计日志记录

## 用户价值

### 效率提升
- 自动化审批流 + 条件分支 + 代理审批
- 诉讼时效自动计算 + 三级预警
- 文档在线编辑 + 版本控制
- 冲突检测自动触发（事件驱动）

### 合规保护
- 3层关联穿透 + 事件驱动二次检查
- 隔离墙中间件 + 访问日志
- PII 脱敏 + 安全头部

### 技术优势
- 101k Go + 73k TS 代码量
- 259 个 API 端点
- 完整可观测性（Prometheus + Grafana + Jaeger）
- Docker 一键部署
