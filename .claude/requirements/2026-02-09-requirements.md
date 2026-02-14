# Law OA Go 需求规格说明书

**日期**: 2026-02-09
**版本**: 1.2 (CODE FREEZE)
**状态**: 待签字确认，冻结后开始开发

---

## 文档修订历史

| 版本 | 日期 | 修订人 | 修订内容 |
|------|------|--------|----------|
| 1.0 | 2026-02-09 | AI技术合伙人 | 初始版本 |
| 1.1 | 2026-02-09 | AI技术合伙人 | 补充边缘场景（补充协议、全文检索、文件夹模板、微信绑定、定期扫描、数据迁移） |
| 1.2 | 2026-02-09 | 张律师 + AI技术合伙人 | **CODE FREEZE 版本**，补充风控核心场景（隔离墙、代管款、离职交接） |

---

## 一、项目概述

### 1.1 项目目标
构建一个现代化的律师事务所办公自动化系统，解决当前律所面临的**文档版本混乱、利益冲突检测不准确、客户沟通效率低、财务管理不透明**等核心痛点。

### 1.2 设计原则
- **数据主权**: 所有数据自建服务器存储，不使用第三方云端服务
- **人工确认**: 所有对外通知必须经过律师人工确认
- **痕迹管理**: 所有关键操作保留可追溯记录
- **白名单模式**: 客户只能看到律师明确授权的信息

---

## 二、核心功能需求

### 2.1 待办事项系统（Inbox + 事件触发）

#### 2.1.1 数据模型

**统一收件箱表 (inbox_items)**

```sql
CREATE TABLE inbox_items (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    source_type VARCHAR(50) NOT NULL,  -- deadline/approval/task
    source_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    priority VARCHAR(20) NOT NULL,     -- critical/high/medium/low
    due_date DATETIME,
    due_date_type VARCHAR(50),         -- hearing/appeal/evidence等
    is_read BOOLEAN DEFAULT FALSE,
    read_at DATETIME,
    is_completed BOOLEAN DEFAULT FALSE,
    completed_at DATETIME,
    reminder_sent BOOLEAN DEFAULT FALSE,
    reminder_count INT DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_unread (user_id, is_read, due_date),
    INDEX idx_user_priority (user_id, priority, is_completed)
);
```

#### 2.1.2 关键日期分类

| 等级 | 类型 | 示例 | 预警策略 |
|------|------|------|----------|
| 🔴 Critical | 上诉期限 | 民事判决收到后15天 | T-30, T-15, T-7, T-3, T-1, T-0 |
| 🔴 Critical | 举证期限 | 法院指定的证据提交截止日 | T-14, T-7, T-3, T-1, T-0 |
| 🔴 Critical | 诉讼时效 | 三年诉讼时效届满日 | T-90, T-30, T-7, T-1, T-0 |
| 🔴 Critical | 执行申请期限 | 判决生效后2年 | T-180, T-90, T-30, T-7, T-1 |
| 🟠 Important | 开庭日期 | 庭审日期 | T-7, T-3, T-1 |
| 🟠 Important | 庭前会议 | 证据交换 | T-3, T-1 |
| 🟠 Important | 调查取证 | 查档截止日 | T-3, T-1 |
| 🔵 Normal | 缴费期限 | 提醒客户交费 | T-3, T-1 |
| 🔵 Normal | 结案归档 | 内勤催归档 | T-7, T-3, T-1 |

#### 2.1.3 升级机制（Escalation Policy）

```
Critical 级别待办 + T-3 状态 + 未完成
  → 自动通知该用户的 supervisor_id（上级/合伙人）
```

**User 表扩展**：
```sql
ALTER TABLE users ADD COLUMN supervisor_id BIGINT;
ALTER TABLE users ADD INDEX idx_supervisor (supervisor_id);
```

#### 2.1.4 时效计算器

**案件详情页新增功能**：
- 输入："收到判决日期：2026-02-10"
- 自动计算："上诉截止日：2026-02-25"
- 自动写入 inbox_items 表

---

### 2.2 利益冲突检测（Conflict Detection）

#### 2.2.1 工商数据 API 集成

**技术选型**: 企查查/天眼查 API

**业务流程**：
```
1. 律师输入："腾讯"
2. 系统调用 API → 弹出下拉列表
   - 深圳市腾讯计算机系统有限公司（存续）
   - 腾讯科技（深圳）有限公司（存续）
3. 律师选择标准工商名称
4. 系统获取股权穿透图
5. 写入 conflict_pool 表（预计算）
6. 实时冲突检测（毫秒级）
```

#### 2.2.2 关联关系检测规则

| 关系类型 | 定义 | 冲突等级 |
|----------|------|----------|
| **严格禁止** | 母公司、全资子公司、控股（>50%） | CRITICAL |
| **风险提示** | 参股（<20%）、共同高管 | HIGH |
| **行业竞争** | 同行业直接竞争 | MEDIUM |

#### 2.2.3 冲突检测报告

**每次检测必须生成 PDF 报告**，包含：
- 检索关键词
- 检索时间
- 匹配到的历史案件
- 匹配到的关联公司
- 系统风险评级（CRITICAL/HIGH/MEDIUM/LOW/PASS）
- 检测人签名

**报告存储**：
```sql
CREATE TABLE conflict_reports (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    report_number VARCHAR(50) UNIQUE NOT NULL,
    checked_by BIGINT NOT NULL,           -- 检测人
    client_name VARCHAR(200) NOT NULL,     -- 标准工商名称
    opposing_party VARCHAR(200),          -- 对方当事人
    risk_level VARCHAR(20) NOT NULL,       -- CRITICAL/HIGH/MEDIUM/LOW/PASS
    matched_cases JSON,                   -- 匹配的案件列表
    related_companies JSON,               -- 关联公司列表
    report_url VARCHAR(500),               -- PDF 报告地址
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

---

### 2.3 文档管理（Document Management）

#### 2.3.1 OnlyOffice 集成

**部署方式**: Docker 私有化部署

**文件锁机制**：
```sql
CREATE TABLE document_locks (
    document_id BIGINT PRIMARY KEY,
    locked_by BIGINT NOT NULL,
    locked_at DATETIME NOT NULL,
    expires_at DATETIME NOT NULL,          -- 30分钟无活动自动释放
    force_unlock BOOLEAN DEFAULT FALSE,    -- 管理员强制解锁
    last_activity DATETIME,
    INDEX idx_expires (expires_at)
);
```

**编辑模式**：
- **在线编辑**: OnlyOffice，实时自动保存
- **离线编辑**: Check-out / Check-in 机制

#### 2.3.2 Check-out / Check-in 流程

```
1. 律师点击"签出"
   → 系统锁定文档
   → 下载当前版本到本地
   → 状态变更为"离线编辑中"

2. 律师在离线编辑
   → 其他人只能查看（只读模式）
   → 显示"张律正在离线编辑"

3. 律师连网，点击"签入"
   → 上传修改后的文件
   → 系统自动创建新版本
   → 解锁文档
   → 版本号递增
```

#### 2.3.3 版本控制

**版本字段**：
```sql
ALTER TABLE documents ADD COLUMN version INT DEFAULT 1;
ALTER TABLE documents ADD COLUMN current_version BOOLEAN DEFAULT TRUE;

CREATE TABLE document_versions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    document_id BIGINT NOT NULL,
    version INT NOT NULL,
    filename VARCHAR(255) NOT NULL,
    filepath VARCHAR(500) NOT NULL,
    created_by BIGINT NOT NULL,
    change_description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_document_version (document_id, version)
);
```

#### 2.3.4 文档可见性控制

```sql
ALTER TABLE documents ADD COLUMN visibility VARCHAR(20) DEFAULT 'internal';
-- internal: 仅内部可见
-- client_visible: 客户可下载
-- public: 公开（罕见）

-- 只有律师能标记为 client_visible
UPDATE documents SET visibility = 'client_visible'
WHERE id = ? AND created_by = ?;
```

---

### 2.4 财务管理（Billing & Invoicing）

#### 2.4.1 数据模型

**合同表 (contracts)**：
```sql
CREATE TABLE contracts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    contract_code VARCHAR(50) UNIQUE NOT NULL,
    case_id BIGINT,
    client_id BIGINT NOT NULL,
    contract_amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(10) DEFAULT 'CNY',
    billing_cycle VARCHAR(50),              -- 一次性/分期/按小时
    payment_terms VARCHAR(100),
    start_date DATE,
    end_date DATE,
    status VARCHAR(20) DEFAULT 'draft',
    signed_at DATE,
    document_id BIGINT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**付款计划 (payment_milestones)**：
```sql
CREATE TABLE payment_milestones (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    contract_id BIGINT NOT NULL,
    name VARCHAR(200) NOT NULL,
    sequence INT NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    percentage DECIMAL(5,2),
    due_date DATE,
    condition TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    invoice_id BIGINT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**发票表 (invoices)**：
```sql
CREATE TABLE invoices (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    invoice_code VARCHAR(50) UNIQUE NOT NULL,
    contract_id BIGINT,
    milestone_id BIGINT,
    client_id BIGINT NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    tax_amount DECIMAL(15,2),
    total_amount DECIMAL(15,2),
    -- 客户开票信息快照
    client_name VARCHAR(200) NOT NULL,
    client_tax_id VARCHAR(50),
    client_address TEXT,
    client_bank_name VARCHAR(100),
    client_bank_account VARCHAR(50),
    -- 开票流程
    status VARCHAR(20) DEFAULT 'draft',
    submitted_at DATETIME,                  -- 行政提交
    approved_by_finance_at DATETIME,        -- 财务复核
    issued_at DATETIME,                     -- 开票
    received_at DATETIME,                   -- 客户签收
    electronic_invoice_url VARCHAR(500),
    created_by BIGINT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**回款记录 (payments)**：
```sql
CREATE TABLE payments (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    invoice_id BIGINT NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    payment_date DATE NOT NULL,
    payment_method VARCHAR(50),
    reference_no VARCHAR(100),              -- 银行流水号
    attachment_id BIGINT,                   -- 回款凭证（必须上传）
    confirmed_by BIGINT NOT NULL,           -- 财务确认
    confirmed_at DATETIME,
    remark TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### 2.4.2 业务规则

**开票金额校验**：
```
申请开票金额 > 合同金额 + 所有补充协议金额 - 已开票金额
→ 拒绝，提示超额
```

**补充协议支持**：
```sql
ALTER TABLE contracts ADD COLUMN parent_contract_id BIGINT;
ALTER TABLE contracts ADD COLUMN contract_type VARCHAR(20) DEFAULT 'original';
-- original: 原始合同
-- supplementary: 补充协议
ALTER TABLE contracts ADD INDEX idx_parent (parent_contract_id);
```

**退费与红冲**：
```sql
ALTER TABLE invoices ADD COLUMN invoice_type VARCHAR(20) DEFAULT 'normal';
-- normal: 正常发票（蓝字）
-- credit: 红字发票（红冲/退费）
ALTER TABLE invoices ADD COLUMN original_invoice_id BIGINT;  -- 原发票ID（红冲时关联）
ALTER TABLE invoices ADD COLUMN refund_reason TEXT;
ALTER TABLE invoices ADD COLUMN write_off_amount DECIMAL(15,2);  -- 核销金额
```

**坏账核销流程**：
```sql
CREATE TABLE bad_debt_records (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    contract_id BIGINT NOT NULL,
    original_amount DECIMAL(15,2) NOT NULL,     -- 原始应收金额
    write_off_amount DECIMAL(15,2) NOT NULL,    -- 核销金额
    reason TEXT NOT NULL,                      -- 核销原因
    status VARCHAR(20) DEFAULT 'pending',       -- pending/approved/rejected
    approved_by BIGINT,                         -- 合伙人审批
    approved_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**开票金额校验**：
```
申请开票金额 > 合同金额 - 已开票金额
→ 拒绝，提示超额
```

**开票审批流程**：
```
律师提交 → 行政初审（抬头、税号）→ 财务复核（金额） → 开票
```

**提成计算规则**：
```
提成基数 = 回款金额 - 成本扣除（差旅费、诉讼费等）

案源人：20%-30%
主办律师：40%-50%
协办律师：10%-15%
助理：5%-10%
```

**User 表扩展**：
```sql
ALTER TABLE users ADD COLUMN commission_rate DECIMAL(5,2);
ALTER TABLE users ADD COLUMN role_type VARCHAR(50);  -- source/lawyer/assistant
```

---

### 2.5 客户门户与通知（Client Portal & Notification）

#### 2.5.1 通知预览队列

**核心安全机制**：
```sql
CREATE TABLE notification_queue (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    trigger_type VARCHAR(50) NOT NULL,
    trigger_id BIGINT NOT NULL,
    case_id BIGINT,
    recipient_type VARCHAR(20) NOT NULL,    -- client/lawyer/admin
    recipient_id BIGINT NOT NULL,
    recipient_name VARCHAR(100) NOT NULL,
    recipient_contact VARCHAR(200),
    channel VARCHAR(20) NOT NULL,            -- email/sms/wechat
    subject VARCHAR(200),
    content TEXT NOT NULL,
    template_id VARCHAR(50),
    status VARCHAR(20) DEFAULT 'pending',    -- pending/approved/sent/cancelled
    priority VARCHAR(20) DEFAULT 'normal',
    -- 审核信息
    created_by BIGINT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    approved_by BIGINT,
    approved_at DATETIME,
    sent_at DATETIME,
    -- 敏感信息标记
    contains_sensitive_info BOOLEAN DEFAULT FALSE,
    auto_send BOOLEAN DEFAULT FALSE,
    INDEX idx_status (status),
    INDEX idx_recipient (recipient_id, status)
);
```

#### 2.5.2 通知渠道配置

| 渠道 | 用途 | 服务商 |
|------|------|--------|
| **微信服务号** | 主要通知渠道 | 微信模板消息 API |
| **短信** | 紧急提醒 | 阿里云/腾讯云短信 |
| **邮件** | 正式文件/长报告 | SMTP |

#### 2.5.3 自动发送规则

**可以自动发送**（无需律师确认）：
- 系统维护通知
- 财务收据/电子发票（财务确认入账后）

**绝对禁止自动发送**（必须律师确认）：
- 所有案件实体进展（立案、开庭、判决、取证）

#### 2.5.4 客户门户功能

**可以做的**：
- ✅ 查看案件进度条
- ✅ 查看日历（开庭时间）
- ✅ 上传文件（证据材料）
- ✅ 下载标记为"Client Visible"的文档

**不能做的**：
- ❌ 留言/聊天功能
- ❌ 下载内部文档
- ❌ 查看未授权的案件

#### 2.5.5 敏感信息过滤

```go
// 发送给客户之前，过滤敏感信息
func FilterSensitiveInfo(content string) string {
    patterns := []struct {
        regex string
        replacement string
    }{
        {`风险评估：[高|中|低]`, "风险评估：已评估"},
        {`胜诉概率：\d+%`, "胜诉概率：已评估"},
        {`我们的策略是.*?风险`, "我们的策略已制定"},
        {`对方弱点：.*`, "对方情况已分析"},
        {`内部评估：.*`, "内部评估已完成"},
    }

    filtered := content
    for _, p := range patterns {
        re := regexp.MustCompile(p.regex)
        filtered = re.ReplaceAllString(filtered, p.replacement)
    }
    return filtered
}
```

#### 2.5.6 客户账户模型

```sql
CREATE TABLE client_accounts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    client_id BIGINT NOT NULL UNIQUE,
    username VARCHAR(50) UNIQUE NOT NULL,  -- 手机号
    password_hash VARCHAR(255) NOT NULL,
    phone VARCHAR(20) UNIQUE,
    status VARCHAR(20) DEFAULT 'active',
    last_login_at DATETIME,
    -- 微信绑定
    wechat_openid VARCHAR(100),               -- 微信OpenID
    wechat_nickname VARCHAR(100),             -- 微信昵称
    wechat_bound_at DATETIME,                -- 绑定时间
    -- 白名单模式：只有明确授权的案件可见
    authorized_cases JSON,                  -- [case_id, case_id, ...]
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### 2.5.7 微信绑定流程

**邀请/绑定机制**：
```
1. 律师在后台生成专属邀请二维码
   → 系统生成带 token 的二维码
   → 二维码包含：client_id, invite_token, expiry

2. 客户扫码关注服务号
   → 微信推送 OpenID 到系统
   → 系统验证 token 有效性
   → 弹出绑定确认页面

3. 客户确认绑定
   → 输入手机号验证码
   → 系统绑定 client_id 与 OpenID
   → 绑定成功通知律师

4. 之后可以发送模板消息
   → 系统使用 OpenID 推送消息
```

**二维码生成接口**：
```
POST /api/v1/clients/:id/wechat/qrcode
  Response: { qrcode_url: string, invite_token: string, expires_at: datetime }
```

**微信回调处理**：
```
POST /api/v1/wechat/callback
  接收微信推送的事件（关注、扫码等）
```

---

---

### 2.6 边缘场景与高级功能

#### 2.6.1 文档全文检索

**需求背景**：律师通常记不住文件名，但记得文档内容中的关键词（如"锂电池爆炸"）。

**技术实现**：复用 Elasticsearch，扩展文档内容索引
```sql
CREATE TABLE document_index_queue (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    document_id BIGINT NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    mime_type VARCHAR(100),
    status VARCHAR(20) DEFAULT 'pending',   -- pending/processing/completed/failed
    error_message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_status (status)
);
```

**支持的文件类型**：
- Word (.docx): 提取纯文本内容
- PDF: 提取文本内容（需 PDFium 或类似工具）
- Txt: 直接索引
- Excel: 提取单元格内容

**权限过滤**：搜索结果必须过滤掉用户无权查看的案件，特别是设有"防火墙"的利益冲突案件。

#### 2.6.2 案件文件夹模板

**需求背景**：每个律师建文件夹的习惯不一样，需要标准化。

**模板定义**：
```sql
CREATE TABLE case_folder_templates (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,           -- 模板名称
    description TEXT,
    folder_structure JSON NOT NULL,     -- 文件夹结构定义
    is_default BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**标准目录结构示例**：
```json
{
  "folders": [
    {
      "name": "01_客户证据",
      "description": "客户提供的证据材料"
    },
    {
      "name": "02_法律文书",
      "description": "起诉状、答辩状、代理词等",
      "subfolders": [
        {"name": "起诉状", "template": "template_indictment.docx"},
        {"name": "答辩状", "template": "template_answer.docx"}
      ]
    },
    {
      "name": "03_法院传票与通知",
      "description": "法院送达的各种文书"
    },
    {
      "name": "04_研究报告与备忘录",
      "description": "内部分析材料"
    },
    {
      "name": "05_结案材料",
      "description": "判决书、裁定书、结案报告"
    }
  ]
}
```

**自动创建触发点**：
- 案件状态变更为"已立案"
- 律师手动点击"创建案件文件夹"

#### 2.6.3 利益冲突定期扫描

**需求背景**：利益冲突是动态的，新入职律师或新案件可能产生新的冲突。

**定时扫描任务**：
```sql
CREATE TABLE conflict_scan_jobs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    scan_type VARCHAR(20) NOT NULL,          -- daily/weekly/manual
    status VARCHAR(20) DEFAULT 'pending',    -- pending/running/completed/failed
    started_at DATETIME,
    completed_at DATETIME,
    scanned_cases INT DEFAULT 0,            -- 扫描的案件数
    found_conflicts INT DEFAULT 0,          -- 发现的冲突数
    new_conflicts JSON,                     -- 新发现的冲突列表
    triggered_alerts BOOLEAN DEFAULT FALSE,  -- 是否已触发告警
    error_message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**扫描策略**：
- **每日扫描**: 检查新增案件与现有案件的冲突
- **每周扫描**: 全量扫描所有案件
- **手动触发**: 律师/管理员手动发起
- **事件触发**: 新律师入职、新案件创建

**告警机制**：
```
Critical 级别新冲突 → 立即通知管理合伙人
High 级别新冲突 → 汇总后每日通知
Medium/Low 级别 → 仅记录，不主动告警
```

---

### 2.7 风控核心场景（Risk Control Critical Scenarios）

> **重要声明**：以下三个场景属于"职业生涯终结"级别的风控需求，必须在系统上线前实现。

#### 2.7.1 隔离墙机制（Ethical Wall / Chinese Wall）

**场景背景**：
冲突检测发现利益冲突，但客户同意豁免。代理 A 公司起诉 B 公司的团队，绝对不能看到代理 B 公司业务的任何信息。

**数据模型扩展**：
```sql
ALTER TABLE cases ADD COLUMN ethical_wall_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE cases ADD COLUMN ethical_wall_description TEXT;

-- 隔离墙白名单（只有被明确授权的人可以访问）
CREATE TABLE case_ethical_wall_whitelist (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    case_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    granted_by BIGINT NOT NULL,              -- 授权人
    granted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    reason TEXT,                             -- 授权理由
    INDEX idx_case (case_id),
    INDEX idx_user (user_id)
);
```

**强制隔离逻辑**：
```
IF case.ethical_wall_enabled == TRUE:
    用户访问该案件时：
    1. 检查 case_ethical_wall_whitelist 表
    2. 如果 user_id 不在白名单中：
       → 拒绝访问（返回 403 Forbidden）
       → 记录访问日志（谁尝试访问、何时、从哪个IP）
    3. 搜索时：该案件不出现在任何搜索结果中
    4. 统计时：该案件数据不出现在任何报表中

    重要例外：
    → 系统管理员（admin）同样受隔离墙限制
    → 高级合伙人同样受隔离墙限制（除非被明确加入白名单）
```

**UI 提示**：
```
┌─────────────────────────────────────────────────────────────────┐
│  ⚠️  隔离墙启用                                                 │
│                                                                   │
│  该案件启用了利益冲突隔离墙。只有被明确授权的人员才能访问。    │
│  当前您无权限查看此案件。如需申请访问，请联系：张合伙人         │
│                                                                   │
│  [返回案件列表]                                                     │
└─────────────────────────────────────────────────────────────────┘
```

**开启隔离墙的条件**：
- 必须有客户书面豁免同意书（上传到文档模块）
- 必须经过管理合伙人审批
- 审批记录永久保存

#### 2.7.2 客户代管款管理（Client Trust Funds / Retainers）

**场景背景**：
客户的钱打到律所账户代交诉讼费，这笔钱不是律所收入，必须单独管理。

**数据模型**：
```sql
-- 客户代管账户
CREATE TABLE client_trust_accounts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    client_id BIGINT NOT NULL,
    account_code VARCHAR(50) UNIQUE NOT NULL,  -- 代管账户编号
    balance DECIMAL(15,2) DEFAULT 0,          -- 当前余额
    currency VARCHAR(10) DEFAULT 'CNY',

    -- 资金用途限制
    purpose_restriction VARCHAR(200),       -- 资金用途说明
    authorized_uses JSON,                  -- 授权用途列表

    status VARCHAR(20) DEFAULT 'active',      -- active/frozen/closed
    opened_at DATETIME,
    closed_at DATETIME,

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_client (client_id),
    INDEX idx_code (account_code)
);

-- 代管款交易记录
CREATE TABLE client_trust_transactions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    account_id BIGINT NOT NULL,

    -- 交易信息
    transaction_type VARCHAR(20) NOT NULL,  -- deposit/deposit_refund/withdraw
    amount DECIMAL(15,2) NOT NULL,
    description TEXT NOT NULL,               -- "代收诉讼费"、"代缴法院费用"

    -- 用途关联
    case_id BIGINT,                           -- 关联案件
    purpose_code VARCHAR(50),                -- 诉讼费/执行费/调查费

    -- 支出信息（如果是支出）
    recipient_name VARCHAR(200),             -- 收款方（如"XX人民法院"）
    recipient_bank_account VARCHAR(50),

    -- 状态
    status VARCHAR(20) DEFAULT 'pending',    -- pending/completed/cancelled
    completed_at DATETIME,
    attachment_id BIGINT,                   -- 付款凭证

    -- 审计信息
    created_by BIGINT NOT NULL,
    approved_by BIGINT,                       -- 财务审批
    approved_at DATETIME,

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_account (account_id),
    INDEX idx_case (case_id),
    INDEX idx_status (status)
);
```

**业务规则**：
```
1. 资金入账时，财务必须分类：
   - 律师费收入 → 律所收入
   - 代管款 → client_trust_accounts

2. 代管款支出限制：
   - 只能用于指定客户的指定用途
   - 不能用于冲抵律师费（除非客户书面授权）
   - 不能用于支付律所运营成本

3. 报表分离：
   - 《利润表》只包含律所收入
   - 《代管款余额表》单独列示
```

**余额不足拦截**：
```
律师提交"使用代管款支付诉讼费"申请
→ 系统检查对应客户托管账户余额
→ 余额不足 → 拒绝并通知客户充值
→ 余额充足 → 提示财务审批
```

#### 2.7.3 离职交接机制（Kill Switch & Offboarding）

**场景背景**：
律师突然离职，必须立即切断系统访问，并快速移交案件。

**数据模型**：
```sql
ALTER TABLE users ADD COLUMN offboarding_status VARCHAR(20) DEFAULT 'active';
-- active: 正常
-- offboarding: 离职交接中
-- deactivated: 已停用

-- 离职交接记录
CREATE TABLE offboarding_records (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    initiated_by BIGINT NOT NULL,           -- 发起人
    initiated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    -- 案件移交
    original_cases JSON NOT NULL,         -- 原主办案件列表 [case_id, ...]
    new_lawyer_id BIGINT,                  -- 接收律师
    case_transfer_completed_at DATETIME,

    -- 待办事项移交
    original_inbox_items JSON NOT NULL,    -- 原待办事项列表
    inbox_transfer_completed_at DATETIME,

    -- 文档处理
    document_disposal_method VARCHAR(50),  -- delete/transfer/revoke_access
    document_disposal_completed_at DATETIME,

    -- 财务结算
    settlement_calculated BOOLEAN DEFAULT FALSE,
    settlement_amount DECIMAL(15,2),
    settlement_paid BOOLEAN DEFAULT FALSE,

    -- 状态
    status VARCHAR(20) DEFAULT 'pending',    -- pending/in_progress/completed
    notes TEXT,

    completed_at DATETIME,
    INDEX idx_user (user_id),
    INDEX idx_status (status)
);
```

**一键离职流程（原子化操作）**：
```go
type OffboardingRequest struct {
    UserID          uint
    InitiatorID     uint
    NewLawyerID     uint   // 接收主办案件的律师（必填）
    NewAssistantID   uint   // 接收待办事项的助理（必填）
    DocumentAction  string // delete/transfer/revoke_access
}

func (s *UserService) InitiateOffboarding(req *OffboardingRequest) error {
    // 1. 立即强制登出所有设备
    s.revokeAllTokens(req.UserID)

    // 2. 冻结账号
    s.updateUserStatus(req.UserID, "offboarding")

    // 3. 批量移交案件
    s.transferCases(req.UserID, req.NewLawyerID)

    // 4. 批量移交待办事项
    s.transferInboxItems(req.UserID, req.NewAssistantID)

    // 5. 文档权限处理
    s.handleDocuments(req.UserID, req.DocumentAction)

    // 6. 通知相关人员
    s.notifyStakeholders(req)

    return nil
}
```

**强制登出机制**：
```
管理员点击"启动离职程序"后：
1. 立即清除该用户所有 RefreshToken
2. WebSocket 断开所有连接
3. 手机App退出登录状态
4. 账号状态变更为 "offboarding"
5. 即使密码正确也无法登录
```

**案件移交界面**：
```
┌─────────────────────────────────────────────────────────────────┐
│  离职交接：张三律师                                            │
│                                                                   │
│  案件移交 (3 个)                                                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ ☑ (2026)深南法民初字第001号  → 接收人：李四律师      │   │
│  │ ☑ (2026)深南法民初字第005号  → 接收人：王五律师      │   │
│  │ ☐ (2026)深南法民初字第008号  → 请选择接收人: [▼]    │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                   │
│  待办移交 (5 个)                                                    │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ ☑ "上诉期限倒计时3天"     → 接收人：李四助理        │   │
│  │ ☑ "法院开庭通知"           → 接收人：李四助理        │   │
│  │ ☑ "客户合同续签提醒"       → 接收人：[▼]           │   │
│  │ ☐ "费用报销申请"           → 请选择接收人: [▼]        │   │
│  │ ☐ "结案报告提交"           → 请选择接收人: [▼]        │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                   │
│  文档处理                                                          │
│  ○ 删除本地缓存的文档                                         │
│  ○ 撤销所有文档编辑权限                                         │
│  ○ 转移所有权到新律师                                           │
│                                                                   │
│  [取消]  [确认交接]                                                │
└─────────────────────────────────────────────────────────────────┘
```

---

## 三、非功能需求

### 3.1 性能要求
- 冲突检测响应时间：< 100ms
- Dashboard 加载时间：< 1s
- 在线文档编辑：无延迟

### 3.2 安全要求
- 所有数据自建服务器存储
- 不使用第三方云端服务（除 OnlyOffice 本地部署）
- 客户数据白名单访问
- 审计日志保留至少3年

### 3.3 可用性要求
- 系统可用性：99.9%
- 数据备份：每日自动备份
- 灾难恢复：RPO < 1小时，RTO < 4小时

---

## 四、接口设计

### 4.1 待办事项接口

```
GET /api/v1/inbox/todos
GET /api/v1/inbox/todos/today
GET /api/v1/inbox/todos/overdue
POST /api/v1/inbox/todos/:id/complete
POST /api/v1/inbox/todos/:id/snooze   -- 延后提醒
```

### 4.2 冲突检测接口

```
POST /api/v1/conflict/check
  Request: { client_name: string, opposing_party: string }
  Response: { risk_level: string, matched_cases: [], report_url: string }

GET /api/v1/conflict/company/search
  Request: { keyword: string }
  Response: { companies: [{name, tax_id, status}] }

GET /api/v1/conflict/reports/:id/download
  Response: PDF file
```

### 4.3 文档管理接口

```
POST /api/v1/documents/:id/checkout
POST /api/v1/documents/:id/checkin
GET /api/v1/documents/:id/lock/status
GET /api/v1/documents/:id/versions
POST /api/v1/documents/:id/versions/:id/restore
POST /api/v1/documents/:id/visibility    -- 设置可见性
```

### 4.4 财务管理接口

```
POST /api/v1/invoices/validate           -- 校验开票金额
POST /api/v1/invoices/:id/submit          -- 提交审批
POST /api/v1/payments/confirm             -- 确认回款
GET /api/v1/commission/calculations       -- 提成计算
```

### 4.5 通知接口

```
GET /api/v1/notifications/queue         -- 待确认通知
POST /api/v1/notifications/queue/:id/approve  -- 确认发送
POST /api/v1/notifications/queue/:id/edit      -- 编辑后发送
POST /api/v1/notifications/queue/:id/cancel    -- 取消发送
```

---

## 五、开发优先级

### P0（必须有）- 第一阶段
1. 待办事项系统（Inbox + 事件触发）
2. 利益冲突检测（工商API + 预计算）
3. 文档版本控制（OnlyOffice 集成）
4. 通知预览队列
5. 财务闭环（合同-发票-回款）

### P1（重要）- 第二阶段
1. 客户门户（只读）
2. 微信通知集成
3. 提成自动计算
4. 文档 Check-out/Check-in

### P2（需要）- 第三阶段
1. 短信通知集成
2. 移动端适配
3. 高级报表功能

---

## 七、数据迁移策略

### 7.1 数据迁移工具需求

**背景**：律所过去10年的案件数据都在老系统或 Excel 中，需要导入到新系统。

### 7.1.1 Excel 批量导入

**支持的数据类型**：
- 客户基本信息
- 案件基本信息
- 合同信息

**导入流程**：
```
1. 下载标准导入模板
2. 填写数据
3. 上传 Excel 文件
4. 系统校验数据格式
5. 预览导入结果
6. 确认导入
```

**数据校验规则**：
- 客户名称不能为空
- 案件标题不能为空
- 日期格式必须正确
- 金额必须是数字
- 必填字段检查

### 7.1.2 历史文档批量上传

**挂载规则**：
- 根据文件名中的案号自动关联到对应案件
- 根据文件名关键词自动分类到对应文件夹
- 无法自动识别的文件放入"待整理"文件夹

**文件命名规范**：
```
推荐格式：案号_文档类型_版本.扩展名
示例：(2026)深南法民初字第001号_起诉状_v1.docx

备选格式：客户名_案件名_文档类型.扩展名
示例：腾讯科技_不正当竞争_起诉状.docx
```

**上传接口**：
```
POST /api/v1/migration/upload/batch
  Request: { files: [], case_id: number }
  Response: { success: int, failed: int, errors: [] }

POST /api/v1/migration/auto-classify
  Request: { case_id: number, folder_path: string }
  自动识别文件类型并分类到标准文件夹
```

### 7.1.3 冲突检测池初始化

**目的**：从历史案件数据构建冲突检测池，确保系统上线时冲突检测有效。

**初始化流程**：
```
1. 导入历史案件数据
2. 提取所有当事人信息
3. 清理和标准化公司名称
4. 构建律师-当事人关系表
5. 预计算冲突关系
```

**数据清理规则**：
- 去除无效字符（空格、特殊符号等）
- 统一公司名称后缀（"有限公司" → "有限"）
- 识别并关联公司别名（"腾讯" → "深圳市腾讯计算机系统有限公司"）

### 7.1.4 迁移数据验证

**验证检查项**：
- [ ] 案件数量与老系统一致
- [ ] 客户数量与老系统一致
- [ ] 合同金额与老系统一致
- [ ] 冲突检测池数据完整
- [ ] 文档关联正确率 > 95%
- [ ] 用户账户权限正确

### 7.1.5 回滚计划

**如果迁移失败**：
- 保留原系统数据至少6个月
- 记录所有迁移操作日志
- 提供数据回滚功能

---

## 九、附录

### 9.1 术语表

| 术语 | 定义 |
|------|------|
| **Inbox** | 统一待办事项收件箱 |
| **Escalation** | 升级机制（未完成待办自动通知上级） |
| **Check-out** | 签出（下载文档到本地编辑，系统锁定） |
| **Check-in** | 签入（上传修改后的文档，解锁） |
| **White List** | 白名单模式（客户只能看到明确授权的信息） |
| **Red Invoice** | 红字发票（退费/红冲） |
| **Write-off** | 坏账核销 |
| **Supplemental Contract** | 补充协议 |
| **Ethical Wall** | 隔离墙（利益冲突强制隔离） |
| **Trust Funds** | 客户代管款（第三方资金） |
| **Offboarding** | 离职交接（律师离职流程） |

### 9.2 变更记录

**v1.1 新增内容**：
1. 财务模块非理想路径（补充协议、退费、坏账）
2. 文档全文检索
3. 案件文件夹模板
4. 微信绑定流程
5. 利益冲突定期扫描
6. 数据迁移策略

**v1.2 新增内容（CODE FREEZE 版本）**：
1. **隔离墙机制**（Ethical Wall）- 职业道德强制隔离
2. **客户代管款管理**（Trust Funds）- 资金合规分离
3. **离职交接机制**（Offboarding）- 一键离职交接

### 9.3 合规性声明

本需求规格说明书符合以下标准和规范：
- ISO 27001 信息安全管理体系
- 律师事务所管理办法
- 律师执业行为规范
- 电子发票管理办法
- 律协业务操作指引

### 9.4 参考文档
- 律师事务所管理办法
- 律师执业行为规范
- 电子发票管理办法

---

## 十、开发计划

### 10.1 版本规划

| 阶段 | 版本 | 功能 | 预计周期 |
|------|------|------|----------|
| **Phase 1** | v2.2.0 | P0 功能 + 风控核心场景 | 8周 |
| **Phase 2** | v2.3.0 | P1 功能 + 客户门户 | 6周 |
| **Phase 3** | v2.4.0 | P2 功能 + 优化 | 4周 |

### 10.2 第一阶段（v2.2.0）详细任务

| 优先级 | 功能模块 | 任务描述 |
|--------|----------|----------|
| P0 | 隔离墙机制 | Ethical Wall 数据模型、权限控制、UI |
| P0 | 客户代管款 | Trust Accounts 表、交易记录、余额校验 |
| P0 | 离职交接 | Offboarding 流程、案件移交、强制登出 |
| P0 | 待办事项系统 | Inbox 表、事件触发、升级机制 |
| P0 | 冲突检测 | 工商API集成、预计算池 |
| P0 | 文档版本控制 | OnlyOffice 集成、文件锁、版本管理 |
| P0 | 通知预览队列 | Notification Queue、律师确认 |
| P0 | 财务闭环 | 合同-发票-回款-提成 |

---

**🎯 文档状态：CODE FREEZE**

**签字确认**：
- [ ] 合伙人签字：_____________  日期：_____________
- [ ] 风控合伙人签字：_____________ 日期：_____________

签字后，需求规格说明书不得修改，所有变更必须通过变更请求（Change Request）流程。

---

*文档结束*

### 6.2 参考文档
- 律师事务所管理办法
- 律师执业行为规范
- 电子发票管理办法

---

*文档结束*
