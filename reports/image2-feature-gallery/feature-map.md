# Law OA 功能地图与出图分批

## 依据

功能梳理基于以下现有事实收敛：

- `README.md` 中的模块完成度、功能特性与主要端点
- `docs/PRD.md` 中的功能需求和用户场景
- `docs/2026-02-09-requirements.md` 中的 P0/P1 风控与协同需求
- `docs/利益冲突/law-firm-coi-prd.md` 中的冲突管理专项设计
- `frontend/src/pages/` 中已存在的前端页面
- `internal/router/router.go` 中已开放的业务路由

## 功能地图

### P0 核心闭环

| 模块 | 当前能力 | 参考实现 |
|---|---|---|
| 工作台 | 统计、待办、活动 | `dashboard/Dashboard.tsx`, `/dashboard/*` |
| 客户管理 | 档案、分类、搜索、统计、PII 脱敏 | `client/ClientManagement.tsx`, `/clients` |
| 案件管理 | 列表、详情、创建、状态流转、增强版 | `case/*`, `/cases`, `/enhanced-cases` |
| 利益冲突检测 | 检查、历史、统计、增强版 v2、实体穿透 | `conflict/*`, `/conflict/*` |
| 审批工作流 | 创建、提交、决策、重提、取消、统计 | `approval/*`, `/approvals/*` |

### P1 企业级支撑

| 模块 | 当前能力 | 参考实现 |
|---|---|---|
| 冲突复核与豁免 | 冲突审核流程、审批联动、历史记录 | 冲突与审批集成、专项 PRD |
| 伦理墙 | 启用/禁用、白名单、访问日志 | `/cases/:id/ethical-wall/*` |
| 文档管理 | 上传、列表、下载、预览、回收站、统计 | `file/FileManagement.tsx`, `/documents/*` |
| OnlyOffice | 在线编辑、转换、回调、锁 | `document/OnlineEditor.tsx`, `/documents/onlyoffice/*` |
| 卷宗目录 | 模板、应用、案件目录树 | `/folder-templates/*`, `/cases/:case_id/folders` |
| 待办收件箱 | 列表、统计、完成、延后 | `inbox/InboxList.tsx`, `/inbox/*` |
| 通知中心 | 队列、模板、审批、发送 | `notification/NotificationQueue.tsx`, `/notifications/*` |

### P2 经营与专业工具

| 模块 | 当前能力 | 参考实现 |
|---|---|---|
| 财务管理 | 合同、付款计划、发票、支付、分成、坏账 | `finance/*`, `/finance/*` |
| 信托账户 | 账户、交易、余额监控 | `trust/*` |
| 法律检索 | 法条搜索、分类、标签、建议 | `legal/LegalCaseSearch.tsx`, `/legal/*` |
| 工具模块 | 截止日期、诉讼费、利息、法律搜索 | `tools/*`, `/deadlines/*` |
| 统计分析 | 仪表板、分析报表、行为统计 | `analytics/AnalyticsDashboard.tsx`, `/analytics/*` |

## 分批策略

### Batch 01 核心接案闭环

1. 工作台指挥中心
2. 客户主档案
3. 接案工作台
4. 利益冲突检查结果
5. 冲突联动审批

### Batch 02 风控与隔离闭环

6. 冲突复核看板
7. 伦理墙控制台
8. 披露与豁免流程
9. 风险审计留痕

### Batch 03 文档协同闭环

10. 文档库
11. OnlyOffice 在线编辑
12. 卷宗目录模板
13. 待办中心
14. 通知队列

### Batch 04 经营与分析闭环

15. 合同与财务总览
16. 发票与回款跟踪
17. 分成规则中心
18. 信托账户总览
19. 管理层分析仪表板

## 统一视觉原则

- 全新品牌化重设计，但不偏离律所后台产品语义
- 中文界面，高密度专业工作台，不做营销官网
- 色彩以暖白、深蓝、青绿、琥珀、暗红为主
- 16:9 桌面端页面样例，适合汇报和产品展示
- 强调真实字段、状态标签、表格密度、侧栏详情和操作区
