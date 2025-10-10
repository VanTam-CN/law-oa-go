# 律所管理系统PRD（Go+Vue微服务版）｜Law Firm Management System PRD (Go+Vue Microservices)

### TL;DR（简介）

**中文：** 本项目旨在打造一套覆盖律所核心业务、功能全面、安全可扩展的现代化管理系统，以 Go 作为后端、Vue 作为前端，采用微服务架构和前后端分离。该系统适用于各类型律所及法律服务企业，支持案件管理、客户管理、文档归档、日程提醒、财务、人事以及多元化权限与合规要求。同时，系统预留与OpenAI、企查查等第三方API的无缝集成接口，具备利益冲突自动化管理等先进模块，方便跨地域多团队协同。产品定位为让中西方开发团队及业务人员、实习生均能易于理解和自定义扩展。

**English:** This project aims to build a comprehensive, secure, and extensible modern management system for law firms, using Go for backend and Vue for frontend. The solution follows a microservices architecture with strict front-end/back-end separation, targeting law firms and legal service providers of all types. The system covers case management, client records, document archiving, scheduling/reminders, financials, HR, multi-level permissions, and compliance. It features extension-ready APIs for easy integration with third-party platforms (OpenAI, Qichacha, etc.) and an advanced conflict-of-interest module—all designed for ease-of-use and customizability by cross-cultural teams and interns alike.

---

## 目标｜Goals

### 业务目标｜Business Goals

* 中文：实现律所案件、客户、业务流全流程数字化与可视化管理。

* 中文：提升案件处理效率，减少人力与管理成本10%以上。

* 中文：加强合规、权限和利益冲突审查能力，杜绝风险点。

* 中文：为律所智能化升级和未来AI+法律服务提供系统底座。

* English: Achieve full digitalization and visualization of law firm cases, clients, and business processes.

* English: Boost case processing efficiency, reducing labor and management costs by at least 10%.

* English: Strengthen compliance, permissions, and conflict-of-interest oversight to eliminate risk hazards.

* English: Build a robust platform enabling law firm smart upgrades and future AI/legal intelligence service integration.

### 用户目标｜User Goals

* 中文：让律师和助理快速上手平台，提升协作与办案透明度。

* 中文：客户可自助查询进度、资料与账务，体验更佳。

* 中文：一键获取历史数据与智能推荐，降低信息壁垒。

* 中文：全程权限可控，信息更安全。

* English: Enable lawyers and assistants to onboard quickly and enhance cooperation and case transparency.

* English: Allow clients to self-check case progress, documents, and invoices for a better experience.

* English: One-click access to historical data and smart recommendations, lowering information barriers.

* English: Full access control throughout, ensuring greater information security.

### 非目标范围｜Non-Goals

* 中文：不涉及线下案件实地调查与资产评估工具集成。

* 中文：不包含金融性业务（如融资、独立结算模块）。

* 中文：非特定国家/地区特殊法律标准专用定制。

* English: Excludes offline case investigation or property assessment tool integration.

* English: Does not cover financial business such as financing or independent settlement modules.

* English: Not a purpose-built system for specific national/regional legal standards only.

---

## 用户故事（详细业务场景）｜User Stories (Detailed Scenarios)

### 管理合伙人 | Managing Partner

* 中文：作为管理合伙人，我需要随时在仪表盘查看所有案件来源、项目进度、团队利用率及账款，以精准预判律所经营趋势和配置资源。

* 英文：As a Managing Partner, I want real-time dashboard visibility of case pipelines, team utilization, and receivables, so I can predict business trends and manage firm resources effectively.

### 执业律师 | Attorney

* 中文：作为律师，每当接收新客户时，希望自动弹出利益冲突核查流程，系统能基于历史客户/对方/合作记录模糊检索并分级提示风险，避免误伤合规。

* 英文：As an Attorney, when onboarding new clients, I want an automated conflict check based on historical and related-party data, with clear risk-level alerts, to avoid compliance breaches.

* 中文：希望案件时间线自动关联邮件、会议、合同版本，团队检索跟进零死角。

* 英文：I want all case timeline events—including emails, meetings, and document versions—linked and available for complete team review at any stage.

### 法务助理 | Paralegal

* 中文：能够批量扫描、上传证据材料并自动按案号、类别归档，且系统识别缺失材料并智能提醒。

* 英文：I want to batch scan/upload evidence and have the system auto-classify by case/type, alerting missing required docs.

* 中文：可一键生成任务清单，自动流转到相关责任人，保障高效沟通和跟进。

* 英文：Auto-generate and route task checklists to assigned teammates, with clear in-system notifications.

### 财务人员 | Billing Clerk/Finance

* 中文：批量开票前审核凭条，异常工时及时与律师对账，账单准确率高，提高回款率。

* 英文：Prior to bulk invoicing, review fee entries and reconcile exceptions with attorneys for accuracy and faster collection.

* 中文：系统自动区分税率和汇率，支持多币种和国际客户。

* 英文：System should apply correct tax/currency codes for local/overseas matters, supporting global billing needs.

### IT/管理员 | IT/Admin

* 中文：分级设置部门与案件权限，并确保所有权限变更留痕，支持后续审计与合规。

* 英文：Configure granular access rules by department/matter, with permanent logs of all permission changes for security and audit.

### 客户 | Client Contact

* 中文：通过加密门户实时跟进案件、上传资料、在线沟通律师、查询账单和支付历史。

* 英文：Access secure portal for case updates, file uploads, live chat with attorney, and invoice/payment review.

### 进阶与特色场景 | Advanced, Realistic Scenarios

* 中文：合伙人自动获知是否有对方与本所历史客户重合并触发伦理隔离机制。

* 英文：Partners are auto-alerted if opponent overlaps with historic firm clients, triggering ethical wall if needed.

* 中文：系统根据政策和地区监管要求自动弹窗业务流程指引，防范因新人误操作导致合规问题。

* 英文：System proactively guides users (esp. interns/newcomers) with real-time policy/regulatory workflow tips, reducing compliance risks.

---

## 功能需求（含微服务模块拆分）｜Functional Requirements (with Microservices Split)

以下功能均基于前后端解耦模式，后端Go服务按领域（Domain Driven Design）独立拆分，前端Vue以模块组件化实现。所有API均提供标准REST接口，关键服务预留人工智能/数据分析/第三方API集成能力。

All features are designed for frontend-backend separation: Go-based backend microservices split by core domain, Vue-based frontend in modular components. APIs use standard REST conventions, every core service is ready for AI, analytics, or 3rd-party API integrations.

### 1\. 案件管理服务 Case Management Service

* 中文：案件录入、流程分解、进度追踪、标签与类型管理、案件协办、历史版本及日志。

* 英文：New case onboarding, workflow breakdown, progress tracking, matter tags/types, collaboration, full change/version logs.

一、核心业务逻辑 Core Business Logic

* 中文：

  1. 用户可新建案件，填写案件类型、关联客户、主办律师、标签、案号等信息。

  2. 案件支持阶段拆解（如：立案、调查、开庭、判决、执行等），每个阶段录入时间、负责人、进展说明和文件。

  3. 所有案件均有完整的历史变更记录及日志，便于溯源。

  4. 支持多角色协同编辑、跨部门流转。

  5. 案件可以和日程/工单/文档/客户等模块双向关联。

* English:

  1. Users can create new cases by filling in type, linked client, lead attorney, tags, case number, and more.

  2. Each case can be split into stages (e.g. filing, investigation, court, judgment, enforcement), each with assigned owner, timestamps, progress notes, and files.

  3. All actions logged for full version/audit trail.

  4. Multi-role, cross-department co-editing and transfer is supported.

  5. Cases are bi-directionally linked with scheduling, tasks, docs, client profiles, etc.

二、主要数据结构 Main Data Structures

* 中文：

  * 案件（Case）：id, 名称, 案号, 类型, 客户id, 主办律师id, 阶段\[\], 标签\[\], 创建时间, 更新时间, 状态, 进展描述

  * 案件阶段（CaseStage）：id, 案件id, 类型, 负责人id, 开始时间, 结束时间, 阶段描述, 附件\[\]

  * 日志（CaseLog）：id, 案件id, 操作类型, 操作人id, 时间, 备注

* English:

  * Case: id, name, case_number, type, client_id, lead_attorney_id, stages\[\], tags\[\], created_at, updated_at, status, progress_desc

  * CaseStage: id, case_id, stage_type, owner_id, start_time, end_time, desc, attachments\[\]

  * CaseLog: id, case_id, action, operator_id, timestamp, note

（简化ER图描述 Text-based ER Overview）

Cases 1---n CaseStage 1---n CaseLog

Cases *--* Client

Cases *--* User (Attorney)

三、接口示例 API Examples

* 中文：

  * 新建案件 POST /api/cases

    * 请求体体例：{ name, case_number, type, client_id, lead_attorney_id, tags }

    * 响应：{ id, created_at }

  * 查询案件 GET /api/cases?filter=xxx

  * 编辑案件 PATCH /api/cases/:id

  * 流程推进/添加阶段 POST /api/cases/:id/stages

  * 查看案件日志 GET /api/cases/:id/logs

* English:

  * Create Case: POST /api/cases

    * Payload: { name, case_number, type, client_id, lead_attorney_id, tags }

    * Response: { id, created_at }

  * Query Cases: GET /api/cases?filter=xxx

  * Edit Case: PATCH /api/cases/:id

  * Add Stage: POST /api/cases/:id/stages

  * View Logs: GET /api/cases/:id/logs

如需每个字段详细含义及前端页面交互流程，可继续补充！

### 2\. 客户/联系人管理服务 Client & Contact Service

* 中文：客户/团队数据管理、联系人关联、快速检索、KYC与风险评级、利益冲突关联校验。

* 英文：Entity & contact records, linking, quick search, KYC/risk ratings, conflict-of-interest backend checks.

一、核心业务逻辑 Core Business Logic

* 中文：

  1. 支持企业客户（单位、机构）和个人客户的录入、编辑、分组和标签。

  2. 每个客户可添加多个联系人并绑定对应职位、联系方式、与案件/合同/历史事项的多级关联。

  3. 提供批量导入、智能去重、自动识别历史交互背景。

  4. 集成KYC（合规尽调）、黑名单校验、自动同步工商/法律公示与企查查API拉取结果。

  5. 支持多标签、多级权限管理，客户动态信息与案件、合同、文档双向同步。

  6. 敏感客户、黑名单自动预警至合伙人和合规员。

* English:

  1. Manage both corporate/institutional and individual clients: create, edit, group, tag.

  2. Each client can have multiple contacts attached—with positions, contact info, direct links to cases/contracts/past matters.

  3. Provide batch import, smart deduplication, and auto-recognition of previous interactions.

  4. Integrated KYC, blacklist checking, automatic sync with business/legal public info (via Qichacha API or equivalents).

  5. Multi-tag and layered permission controls; client activity is bi-directionally synced across matters, contracts, and documents.

  6. Sensitive/blacklisted clients auto-flagged to partners/compliance officers.

二、主要数据结构 Main Data Structures

* 中文：

  * 客户 Client: id, 名称, 类型, 企业信息/个人信息(统一社会信用代码/身份证号/国籍/等), 标签\[\], 管理人id, 风险评级, KYC状态, 创建时间, 客户状态, 联系人\[\], 关联案件\[\], 关联合同\[\]

  * 联系人 Contact: id, 客户id, 姓名, 职位, 电话, 邮箱, 历史交互\[\], 备注, 黑名单标记

* English:

  * Client: id, name, type, entity_info/personal_info (registration_no/national_id/country/etc.), tags\[\], manager_id, risk_score, kyc_status, created_at, client_status, contacts\[\], related_cases\[\], related_contracts\[\]

  * Contact: id, client_id, name, position, phone, email, past_interactions\[\], note, blacklist_flag

（简化ER图描述 Text-based ER Overview）

Client 1---n Contact

Client *--* Case/Contract/Doc

三、接口示例 API Examples

* 中文：

  * 新增/批量导入客户 POST /api/clients

    * 请求体：{ name, type, entity_info/personal_info, tags, manager_id }

  * 编辑客户 PATCH /api/clients/:id

  * 客户检索 GET /api/clients?filter=xxx

  * 添加联系人 POST /api/contacts

  * 客户风险校验 POST /api/clients/:id/check-risk

  * 客户企业拉取企查查 API: GET /api/clients/:id/qichacha

* English:

  * Create/import client: POST /api/clients

  * Edit client: PATCH /api/clients/:id

  * Search/list clients: GET /api/clients?filter=xxx

  * Add contact: POST /api/contacts

  * Check risk/KYC: POST /api/clients/:id/check-risk

  * Fetch entity public record (Qichacha etc.): GET /api/clients/:id/qichacha

如需列出典型客户动态、前端操作流程或API数据字典示例，可进一步细化！

### 3\. 文档与证据归档服务 Document & Evidence Archiving

* 中文：全类型文件上传、OCR及AI文本提取、分类与标签、权限层级、全文及条件检索、证据链管理。

* 英文：Universal file upload, OCR/AI text extraction, document tagging/classification, tiered permissions, full/conditional search, evidence traceability.

一、核心业务逻辑 Core Business Logic

* 中文：

  1. 用户通过案件、客户等模块上传证据、合同、鉴定文书、财务票据等全类型文件。

  2. 系统自动识别文件类型，支持OCR文本抽取与AI标签、智能关联至案件、阶段、客户/对方等。

  3. 文件按案件、证据类型、标签、上传人、日期、权限等维度多级归档检索。

  4. 支持按角色和案件设置细粒度文件、文件夹访问权限。违规访问自动记录到审计日志。

  5. 针对涉密或合规类文件，系统提示合规风险并强制敏感操作二次验权。

  6. 与案件、日程、任务、利益冲突、AI分析等模块数据打通，实现双向同步与调用。

* English:

  1. Users upload all types of files (evidence, contracts, expert opinions, invoices, etc.) via cases, clients, or other entry points.

  2. System auto-detects file types and uses OCR/AIlabelling, then links docs to matter, stage, client/counterparty as relevant.

  3. Multi-level archiving and search: by case, evidence type, tags, uploader, date, permission, etc.

  4. Fine-grained access control per file/folder by role and case; unauthorized access triggers audit log entry.

  5. For confidential/compliance documents, the system highlights risk and enforces 2FA for sensitive operations.

  6. Integrates with cases, calendar, tasks, conflicts, AI; enables fully bi-directional linking and retrieval.

二、主要数据结构 Main Data Structures

* 中文：

  * 文档 Document: id, 名称, 路径, 案件id, 客户id, 上传人id, 类型, 标签\[\], 权限\[\], 文件大小, 上传时间, 状态, OCR内容, AI标签, 安全等级

  * 文件夹 Folder: id, 父级id, 名称, 权限\[\], 创建人id, 创建时间, 状态

  * 文件访问日志 DocAccessLog: id, 文档id, 用户id, 操作类型, 时间, 备注

* English:

  * Document: id, name, path, case_id, client_id, uploader_id, type, tags\[\], permissions\[\], size, uploaded_at, status, ocr_text, ai_labels, security_level

  * Folder: id, parent_id, name, permissions\[\], creator_id, created_at, status

  * DocAccessLog: id, document_id, user_id, action, timestamp, note

（简化ER图描述 Text-based ER Overview）

Folder 1---n Document

Document *--* Case/Client

Document 1--n DocAccessLog

三、接口示例 API Examples

* 中文：

  * 上传文档 POST /api/documents

    * 请求体：{ name, file, case_id, client_id, type, tags }

    * 响应：{ id, uploaded_at }

  * 文档批量检索 GET /api/documents?filter=xxx

  * 权限设置 PATCH /api/documents/:id/permissions

  * 下载/预览 GET /api/documents/:id/download

  * 文件操作日志查询 GET /api/documents/:id/logs

* English:

  * Upload Document: POST /api/documents

    * Body: { name, file, case_id, client_id, type, tags }

    * Response: { id, uploaded_at }

  * Batch Search: GET /api/documents?filter=xxx

  * Set Permissions: PATCH /api/documents/:id/permissions

  * Download/Preview: GET /api/documents/:id/download

  * Access Log: GET /api/documents/:id/logs

如需AI标签/OCR示例、文件加密与追踪详细机制，可继续细化！

### 4\. 日程与任务工单服务 Scheduling & Task Service

* 中文：日历同步、会议/诉讼/截止节点、工单流转、自动提醒、自定义视图、进度甘特图。

* 英文：Calendar sync, meetings/litigation/deadlines, workflow tasks, smart reminders, customizable views, Gantt progress tracking.

### 5\. 财务与计费服务 Finance & Billing Service

* 中文：工时录入、费用申请与审批、自动计费与发票出具、多币种/多税率、回款跟踪、集成第三方支付。

* 英文：Work-time entry, fee requests/approvals, automated billing/invoicing, multi-currency/tax, receivables tracking, integrated 3rd-party payments.

### 6\. 权限与合规管理服务 Permissions & Compliance Service

* 中文：多级角色、细粒度功能和数据权限、敏感操作审计、合规校验/策略自动提示。

* 英文：Multi-level roles, fine-grained feature/data access, operation audit trails, compliance check rules and proactive hints.

### 7\. 消息与沟通中心 Notification & Communication Center

* 中文：系统通知、内部消息、外部邮件对接、文件流转提醒、跨平台推送/短信整合。

* 英文：System announcements, in-system chat, external email sync, file circulation alerts, cross-device push/SMS.

### 8\. 利益冲突与伦理墙 Conflict of Interest & Ethical Wall

* 中文：自动化冲突检测流程、分级风险判定、利益墙隔离、全量事件可追溯、合规申报与审批。

* 英文：Automated conflict checks, risk-tiering, ethical wall enforcement, full traceability, e-filing and approval for disclosures.

利益冲突与伦理墙模块｜Conflict of Interest & Ethical Wall Module

一、核心业务逻辑 Core Business Logic

* 中文：

  1. 新增案件或新增客户流程“自动触发”利益冲突检索，系统基于客户/对方/关联自然人/合作历史多维模糊匹配，支持自定义风险等级阈值。

  2. 检索结果如有疑似冲突，系统实时弹窗告知涉及的历史事项与潜在冲突关系，在相关用户流程中强制“说明”或“报备”操作。

  3. 合伙人/合规专员有权进入冲突处置审批流程，可设“利益墙”——即部分律师对敏感案件的信息隔离，不可见案件全部或部分资料。

  4. 所有冲突查询与处置过程均自动记录并纳入全系统操作日志。

  5. 符合合规特殊需求时（例如跨区、跨国/跨所协作），可引入多重审批和隔离规则.

* English:

  1. Whenever a new case or client is added, an automated conflict check is triggered: the system fuzzy-matches on client, counterparty, associated individuals, and collaboration history. Risk thresholds are customizable.

  2. If a potential conflict is detected, the system pops up an immediate alert with involved prior matters and parties. Action is required: user must explain or submit for review.

  3. Partners/compliance officers launch resolution workflows. "Ethical wall" can be applied—blocking some attorneys from viewing (part of) a matter’s details.

  4. All conflict checks and resolutions are auto-logged for full auditability.

  5. Multi-step approval/isolation policies are supported for complex (cross-border, cross-firm) matters.

二、主要数据结构 Main Data Structures

* 中文：

  * 利益冲突查询 ConflictCheck：id, 触发类型(case/client), 触发对象id, 检索结果集\[\], 风险等级, 查询用户id, 查询时间, 处置状态(status), 审批链\[\], 操作日志\[\]

  * 利益墙配置 EthicalWall: id, 案件id, 隔离用户id\[\], 可见字段\[\], 可访问时段\[\], 状态

  * 审批记录 ApprovalRecord: id, 冲突查询id, 审批人id, 意见, 时间, 结果

* English:

  * ConflictCheck: id, trigger_type (case/client), trigger_id, matched_items\[\], risk_level, user_id, checked_at, status, approval_chain\[\], op_logs\[\]

  * EthicalWall: id, case_id, restricted_user_ids\[\], visible_fields\[\], allowed_times\[\], status

  * ApprovalRecord: id, conflict_check_id, approver_id, comment, timestamp, result

（简化ER图描述 Text-based ER Overview）

ConflictCheck n---1 Case/Client

ConflictCheck 1---n ApprovalRecord

EthicalWall 1---n User (Attorneys)

三、接口示例 API Examples

* 中文：

  * 发起冲突检索 POST /api/conflict-check

    * 请求体：{ type, object_id, operator_id }

    * 响应：{ id, risk_level, matched_items, next_step }

  * 查看冲突审批 GET /api/conflict-check/:id/approvals

  * 配置/查询利益墙 PATCH/GET /api/ethical-wall/:case_id

* English:

  * Trigger Conflict Check: POST /api/conflict-check

    * Body: { type, object_id, operator_id }

    * Response: { id, risk_level, matched_items, next_step }

  * Get Conflict Approvals: GET /api/conflict-check/:id/approvals

  * Manage/Query Ethical Wall: PATCH/GET /api/ethical-wall/:case_id

---

### 9\. 智能与第三方API接入 AI & External API Connector

* 中文：集成OpenAI（法律摘要/自助问答/合同智能比对）、企查查（尽职调查/工商信息拉取）、API网关对接模式、插件扩展。

* 英文：Integrations for OpenAI (smart legal Q&A, doc summaries, contract review), Qichacha (due diligence, company data), API gateway model, plugin ready.

### 10\. 统计与数据分析服务 Analytics & Reporting Service

* 中文：案件/客户/财务/绩效统计，业务预警，数据可视化仪表盘，多维度导出、API拉取。

* 英文：Case, client, financial, and performance dashboards, business alerts, data visualization, flexible reporting and API extraction.

如需每个功能组下“具体业务逻辑/接口示例/数据结构”分开详细说明，或需要按优先级排序，请进一步指示。

### 详细功能与优先级｜Detailed Feature Breakdown

Features grouped by microservice (Priority: P0 critical, P1 important, P2 later)

* Auth & Identity (Priority: P0)

  \-- SSO/OIDC + JWT: Unified authentication with MFA and session management.

  \-- RBAC/ABAC: Roles and attribute-based access; matter-level permissions.

  \-- Audit Logging: Immutable logs for sign-in, privilege changes, data access.

* Firm Admin & Settings (Priority: P0)

  \-- Firm/Office/Department: Manage org units, billing rates, holidays.

  \-- User/Role: Provisioning, license seats, approval flows.

  \-- Billing Settings: Tax, currencies, invoice numbering, templates.

* Client & Matter Management (Priority: P0)

  \-- Client Intake: Forms, required fields, KYC/AML placeholders, duplicate detection.

  \-- Matter Lifecycle: Open/assign, phases, custom fields, status tracking.

  \-- Parties & Relationships: Client, counterparty, related entities, counsel.

  \-- Engagement: Engagement letters, scope, fee arrangements.

* Conflict of Interest (Priority: P0)

  \-- Conflict Search: Fuzzy match across clients, parties, attorneys, prior matters.

  \-- Review Workflow: Flag severity, capture notes, clearance/waiver process.

  \-- Ethical Walls: Access restrictions, acknowledgment tracking, monitoring.

* Time & Expense (Priority: P0)

  \-- Timers & Quick Entry: Multi-timers, calendar/email suggestion to time.

  \-- Rate Cards: Standard/negotiated rates, discounts, caps, AFAs.

  \-- Expenses: Hard/soft costs, receipts, approvals.

* Billing & Payments (Priority: P1)

  \-- Pre-bill Review: Adjustments, narrations, write-downs, approvals.

  \-- Invoicing: PDF/HTML, schedule, batch runs, multi-currency/tax.

  \-- Payments: Online payment link, partials, trust account application.

* Trust & Accounting Bridge (Priority: P1)

  \-- Trust Ledgers: Deposits, disbursements, three-way reconciliation basics.

  \-- Export/Sync: Bridge to external GL/ERP systems (e.g., summary journals).

* Document Management (Priority: P1)

  \-- Storage & Versioning: Folders, tags, version control, check-in/out.

  \-- Preview & OCR: Common formats, text extraction for search.

  \-- Share & Links: Secure links, expiry, watermark.

* Task & Workflow (Priority: P1)

  \-- Tasks & Checklists: Templates per matter type, due dates, reminders.

  \-- Calendar & Deadlines: Court dates, limitation periods, dependencies.

  \-- Comments & Mentions: Threaded comments, @mentions, notifications.

* Reporting & Analytics (Priority: P1)

  \-- Dashboards: Intake funnel, WIP, utilization, AR aging.

  \-- Exports: CSV/Excel/PDF; scheduled emails.

  \-- Custom Reports: Saved filters, widgets.

* Notifications & Communications (Priority: P1)

  \-- Email/SMS/Push: Events, approvals, reminders.

  \-- Templates: Merge fields, localization, preview.

* Integration Gateway (Priority: P1)

  \-- Connectors: OpenAI, Qichacha (企查查), payment gateways, storage, email.

  \-- Webhooks & Events: Subscriptions, retry, signing secrets.

  \-- LLM Extensions: Prompt templates, guardrails, token/account quotas.

* Client Portal (Priority: P2)

  \-- Secure Access: View matters, invoices, pay online, upload docs.

  \-- Messaging: Controlled two-way communication.

---

## User Experience

A detailed journey from first contact to billing, emphasizing clarity and low-friction.

**Entry Point & First-Time User Experience**

* Access via SSO link or invite email; initial password-less magic link optional.

* Onboarding wizard (5 minutes): firm settings, offices, billing preferences, roles, initial data import (CSV).

* Guided tours: client intake, conflict checks, timekeepers setup, invoice template.

**Core Experience**

* Step 1: Create Client & Matter

  * Minimal required fields; inline duplicate detection; KYC checklist.

  * Validate required fields and conflict check prerequisites.

  * If valid, proceed to conflict screening; else highlight errors with tips.

* Step 2: Conflict Check

  * Auto-search across historical clients, parties, attorneys, opposing counsel.

  * Display matches with confidence scores and risk tags (direct/related/name-similarity).

  * Actions: Request waiver, add notes, escalate to Conflicts Committee, or clear.

* Step 3: Engagement & Team Assignment

  * Generate engagement letter from template; e-sign placeholder; store version.

  * Assign responsible attorney, team, rate card; set matter phase and budget.

* Step 4: Work Execution

  * Tasks/checklists generated from matter type; due dates; reminders.

  * Document workspace with upload, versioning, previews, quick tag.

  * Time capture via timers, meeting parsing, suggested entries.

* Step 5: Pre-billing Review & Invoicing

  * Review WIP; edit narratives; apply discounts/caps; approval workflow.

  * Generate invoice; send via email/portal; payment link.

  * Apply payments and trust transfers with audit trail.

* Step 6: Reporting & Close

  * Dashboards for utilization, AR aging, realization; export snapshots.

  * Close matter; archive documents; retain according to policy.

**Advanced Features & Edge Cases**

* Ethical wall breach attempt: block access, alert admin, log incident.

* Duplicate client/matter: merge with conflict-safe rules.

* Offline time capture: local cache, background sync with conflict resolution.

* Multi-currency billing: FX at invoice date; revaluation notes.

* Data residency: route storage by region; enforce PIPL/GDPR restrictions.

**UI/UX Highlights**

* Accessible color contrast, keyboard-first navigation, screen-reader labels.

* Responsive layout for desktop/tablet/mobile; density switch for data tables.

* Global search with typeahead; saved filters and views.

* Inline guidance and empty-state examples; undo for destructive actions.

* Full i18n (Chinese/English UI, date/time/number locale, time zone awareness).

---

## Narrative

Li, a managing partner at a growing firm, struggles with fragmented tools: spreadsheets for intake, email chains for conflict checks, and manual time entries that arrive late and incomplete. Invoices are delayed, realization suffers, and partners worry about ethical exposure from rushed onboarding.

With our Go+Vue microservices platform, Li’s team opens a new matter in minutes. As intake completes, the system runs a conflict search across clients, matters, and related parties, surfaces potential issues, and shepherds the clearance or waiver workflow. An ethical wall is created automatically, limiting access to authorized attorneys. The paralegal launches a checklist and builds the matter workspace with structured folders and templates. Attorneys capture time via timers and suggested entries derived from calendar and email, while documents are versioned and searchable. The billing clerk uses pre-bill review to adjust narratives and discounts before sending invoices with payment links.

Within weeks, the firm sees faster onboarding, fewer write-offs, and higher compliance confidence. Dashboards reveal utilization and AR trends, guiding decisions on staffing and pricing. Clients enjoy clarity through timely invoices and secure sharing. The firm scales operations without sacrificing governance, turning operational discipline into competitive advantage.

---

## Success Metrics

* 40% reduction in intake-to-matter-open time (median days).

* 10% increase in billable utilization (hours captured vs. capacity).

* 15% reduction in write-offs/discounts (realization rate improvement).

* DSO (days sales outstanding) reduced by 20%.

* Conflict clearance turnaround under 24 hours (P90).

* System availability ≥ 99.9% and error rate < 0.1% of requests.

### User-Centric Metrics

* Weekly active users / total eligible users ≥ 85%.

* Time captured per user per week vs. baseline +10–15%.

* CSAT ≥ 4.5/5 on onboarding and monthly NPS ≥ 40.

### Business Metrics

* Revenue per lawyer +8–12%; AR aging (90+ days) reduced by 25%.

* Admin hours per invoice -30%; operating margin +5 pts.

### Technical Metrics

* P95 API latency < 300 ms for core endpoints; search P95 < 800 ms.

* Uptime ≥ 99.9%; background job success rate ≥ 99.5%.

### Tracking Plan

* Events: client_created, conflict_check_started/completed, matter_opened, time_entry_created, prebill_approved, invoice_sent, payment_received, document_uploaded, task_completed, ethical_wall_created, access_denied_wall.

* Funnels: intake → conflict → engagement → first time entry → first invoice.

* Cohorts: attorney adoption, practice group utilization, client payment behavior.

* Error/Performance: API 4xx/5xx rates, job queue delays, search timeouts.

---

## Technical Considerations

### Technical Needs

* Microservices: Auth, Admin, Clients/Matters, Conflicts, Time/Expense, Billing, Trust, Documents, Tasks/Workflow, Reporting, Notifications, Integration Gateway.

* API Gateway: REST v1 externally; internal service-to-service RPC; rate limiting, auth, metrics.

* Front-end: Vue SPA with router, state store, i18n, component library.

* Back-end: Go services with clean architecture; background jobs; event bus for decoupling.

* Search: Dedicated index for conflict and document full-text/fuzzy queries.

* Storage: Relational DB for core data; object storage for documents; cache for sessions/queues.

* Observability: Centralized logs, traces, metrics, alerting.

### Data Structure Design

Primary entities and relationships (key fields only; simplified):

* Users (1..*) Roles (M:N); Users (1..*) TimeEntries; Users (1..\*) Matters (as team).

* Firm (1..*) Office; Client (1..*) Matter; Matter (1..\*) Party (roles: client, adverse, related).

* Matter (1..\*) Documents, Tasks, TimeEntries, Expenses, Invoices.

* ConflictCheck (1..\*) ConflictResult (per matched entity); ConflictCheck (1)→(1) Matter.

* Invoice (1..*) Payment; TrustAccount (1..*) TrustTransaction.

Example tables:

* users: id, firm_id, email, name, role, status, mfa_enabled, locale, timezone, created_at, updated_at

* roles: id, code, name, description; role_permissions: role_id, permission_code

* clients: id, firm_id, name, type (individual/company), reg_no, country, risk_level, created_at

* matters: id, firm_id, client_id, title, type, status, lead_attorney_id, phase, open_date, close_date, budget_amount, rate_card_id

* parties: id, firm_id, matter_id, name, type (adverse/related/counsel), reg_no, country, contact_info

* conflict_checks: id, firm_id, matter_id, initiated_by, status (pending/cleared/blocked/waiver), score_max, created_at, decided_at

* conflict_results: id, conflict_check_id, entity_type (client/matter/party/attorney), entity_id, match_score, risk_tag, note, disposition

* time_entries: id, firm_id, matter_id, user_id, date, hours, activity_code, rate, amount, narrative, status (draft/posted)

* expenses: id, firm_id, matter_id, date, type, amount, tax, receipt_url, status

* invoices: id, firm_id, client_id, matter_id, number, currency, issue_date, due_date, subtotal, tax, total, status

* payments: id, firm_id, invoice_id, amount, method, received_at, reference, status

* trust_accounts: id, firm_id, name, bank, currency, balance; trust_transactions: id, account_id, type, amount, matter_id, reference, created_at

* documents: id, firm_id, matter_id, folder_id, title, version, storage_key, size, checksum, tags, created_by, created_at

* tasks: id, firm_id, matter_id, title, description, assignee_id, due_date, status, priority, checklist_json

* audit_logs: id, firm_id, user_id, action, target_type, target_id, payload_json, ip, user_agent, created_at

Text ER diagram (simplified): Firm → Offices; Firm → Users; Firm → Clients → Matters

Matters → Parties; Matters → Documents; Matters → Tasks

Matters → TimeEntries, Expenses, Invoices → Payments

Matters → ConflictChecks → ConflictResults

Firm → TrustAccounts → TrustTransactions

### API Specification

General

* Base URL: /api/v1; Auth: Bearer JWT (OIDC), optional API keys for integrations.

* Pagination: page, page_size; Sorting: sort=field:asc|desc; Filtering: q, fields.

* Idempotency-Key header for POST/PUT with retries.

* Errors: { code, message, details, trace_id }.

* Webhooks: HMAC-signed, retry with exponential backoff.

Core endpoints (selected):

* Auth: POST /auth/login, POST /auth/refresh, POST /auth/logout, GET /me

* Users/Roles: GET/POST /users, GET/PUT /users/{id}, GET/POST /roles

* Clients: GET/POST /clients, GET/PUT /clients/{id}

* Matters: GET/POST /matters, GET/PUT /matters/{id}, POST /matters/{id}/assign

* Parties: GET/POST /matters/{id}/parties

* Conflict:

  * POST /matters/{id}/conflict-checks (start)

  * GET /conflict-checks/{id} (status/results)

  * POST /conflict-checks/{id}/decide (clear/block/waiver, notes)

  * POST /conflict-checks/{id}/waiver-request (attach docs, recipients)

* Time/Expense:

  * POST /time-entries, GET /time-entries?matter_id=

  * POST /expenses, GET /expenses?matter_id=

* Billing/Payments:

  * POST /invoices (from WIP), GET /invoices/{id}, POST /invoices/{id}/send

  * POST /payments (apply to invoice), GET /payments/{id}

  * POST /trust-transactions (deposit/disbursement/transfer)

* Documents:

  * POST /matters/{id}/documents (multipart), GET /documents/{id}, GET /matters/{id}/documents

  * POST /documents/{id}/share-links

* Tasks/Workflow:

  * POST /matters/{id}/tasks, GET /tasks?assignee_id=, POST /tasks/{id}/complete

* Reporting: GET /reports/wip, GET /reports/ar-aging, GET /reports/utilization

* Integrations:

  * GET /connectors, POST /connectors (register)

  * POST /connectors/{id}/secrets (store/update)

  * POST /llm/prompts (run with template_id, inputs)

  * GET /external/qichacha/search?keyword=

* Webhooks: POST /webhooks (register), DELETE /webhooks/{id}

LLM/Third-party extension notes

* LLM requests use prompt templates with variables, grounding via matter docs (RAG), PII redaction option, audit logs of prompts/completions.

* Rate limits and cost caps per user/firm; offline queue fallback.

* Third-party connectors follow OAuth2/API-key patterns; secrets stored encrypted; events emitted on sync.

### “利益冲突”管理模块｜Conflict of Interest Management

Scope & scenarios

* Conflicts at new client/matter intake, party updates, team changes, or merging entities.

* Types: direct adverse, positional, former client, related entity, name-similarity/alias, lateral screening.

Logic

* Index entities: clients, matters, parties, attorneys, counsel, company registries.

* Matching: exact, phonetic/romanization for Chinese names, n-gram fuzzy, alias dictionaries.

* Scoring: weight by entity type, role, recency, matter type, jurisdiction.

* Disposition workflow: pending → review → cleared / blocked / waiver requested → approved/denied.

* Ethical wall: when cleared-with-wall or blocked-to-others, enforce ABAC; require acknowledgments.

Data structures

* conflict_checks: id, matter_id, initiated_by, status, score_max, created_at, decided_at

* conflict_results: id, conflict_check_id, entity_type, entity_id, match_score, risk_tag (direct/related/positional/name), note, disposition

* ethical_walls: id, matter_id, scope (docs/time/notes), created_by, created_at

* wall_members: wall_id, user_id, role (allowed/denied), acknowledged_at

APIs

* POST /matters/{id}/conflict-checks → {check_id, status=pending}

* GET /conflict-checks/{id} → results\[\], score_max, suggested_disposition

* POST /conflict-checks/{id}/decide {disposition, notes, wall_scope, approver_id}

* POST /ethical-walls {matter_id, scope, members\[\]}

* GET /ethical-walls/{id}, POST /ethical-walls/{id}/acknowledge

Process flow

* Trigger: intake submit → pre-commit conflict check.

* Search & score → present UI with highlights and evidence.

* Reviewer assigns disposition; if waiver needed, generate letter and route for signatures.

* If wall required, create wall, set ABAC rules, notify team; log all in audit.

* On updates (new parties, new team members), auto re-check with delta.

Controls

* Must-block for direct conflicts unless waiver approved.

* Automated access denials logged; PII minimization for reviewers when possible.

* SLA targets and escalation rules; dashboard for aged pending checks.

### 第三方平台集成能力｜3rd Party Platform Integration

OpenAI (example)

* Use cases: summarize documents, draft engagement clauses, convert emails to time entries, generate invoice narratives.

* Flow: user action → prompt template + context (matter metadata, selected docs) → moderation/guardrails → completion → store LLMResponse with trace.

* Controls: token quotas, model selection policy, PII redaction toggle, human-in-the-loop approvals for sensitive tasks.

Qichacha 企查查 (example)

* Use cases: company verification during intake; fetch UBOs, legal reps, risk signals.

* Flow: intake form → search API → select entity → hydrate client/party fields → store source and timestamp → periodic refresh.

* Reliability: caching with TTL, rate limit handling, fallback manual entry.

Extension model

* IntegrationConnector registry with: id, type (LLM/CRM/KYC/Payments), auth scheme, scopes, secrets.

* Event-driven sync via webhooks; signed callbacks; replay support.

* Mappings & transforms for field-level normalization; sandbox vs production keys.

### Data Storage & Privacy

* Data flow: SPA → API Gateway → microservices → DB/Search/Object store → async jobs.

* Multi-tenant isolation by firm_id; row-level security for matter isolation; encryption at rest and in transit.

* Compliance: PIPL (China), GDPR (EU), SOC2-aligned controls; data residency by region; retention policies per matter.

* Access: least privilege, periodic role reviews, admin impersonation with explicit consent and logs.

### Scalability & Performance

* Target 500–2,000 users per firm; 50 concurrent attorneys; search index up to 5M documents/parties.

* Horizontal scaling for stateless services; background workers for indexing/OCR/billing runs.

* Caching for frequent reads; async conflict searches with progressive results.

### Potential Challenges

* Name matching accuracy for Chinese/English entities; mitigate with hybrid scoring and review.

* Ethical wall enforcement across search, documents, and logs; require ABAC and consistent checks.

* Billing/Trust compliance nuances across jurisdictions; configurable rules engine.

* LLM data leakage risks; implement redaction, allowlists, and audit.

---

## Milestones & Sequencing

Lean roadmap optimized for speed and learning; total 6 weeks target.

### Project Estimate

Large: 4–8 weeks (target 6 weeks for MVP with critical workflows)

### Team Size & Composition

Small Team: 1–2 total people

* Full-Stack Engineer (Go/Vue/DevOps): builds services, front-end, CI/CD, integrations.

* Product Designer/PM: requirements, UX/UI, testing, stakeholder demos, docs.

### Suggested Phases

**Phase 0: Inception & Foundations (3 days)**

* Key Deliverables: PM—requirements freeze (MVP scope), UX wireframes; Engineer—repo, CI/CD, auth skeleton, data model baseline.

* Dependencies: Access to domain experts; environment credentials.

**Phase 1: Intake + Conflicts + Matters (2 weeks)**

* Key Deliverables: Engineer—clients/matters service, conflict engine MVP (search, score, decide), ethical walls, basic DDL; PM—UX for intake and conflict review.

* Dependencies: Search index standing up; seed test data.

**Phase 2: Time & Expense + Documents (2 weeks)**

* Key Deliverables: Engineer—time/expense capture, document storage/versioning, global search; PM—task/checklist templates, docs UX polish.

* Dependencies: Object storage and OCR worker.

**Phase 3: Billing & Trust + Reporting (1 week)**

* Key Deliverables: Engineer—pre-bill review, invoicing, payments/trust ledger; PM—billing templates, dashboards.

* Dependencies: Payment sandbox, invoice template assets.

**Phase 4: Integrations + Hardening & Beta (1 week)**

* Key Deliverables: Engineer—OpenAI + Qichacha connectors, webhooks, role audits, observability; PM—UAT, training materials, release notes.

* Dependencies: API keys, data privacy review.

**Go-Live & Feedback Loop (ongoing, 1 week post-launch)**

* Key Deliverables: PM—collect KPIs and user feedback; Engineer—hotfixes, backlog grooming.

* Dependencies: Analytics dashboards and alerting in place.
