package models

import (
	"time"
)

// ============================================================================
// Law OA Go v2.2.0 Models
// ============================================================================
// 基于需求规格说明书 v1.2 (CODE FREEZE) 的完整数据模型
//
// 包含模块:
//   - 待办事项系统 (Inbox System)
//   - 利益冲突检测 (Conflict Detection)
//   - 文档管理 (Document Management)
//   - 财务管理 (Financial Management)
//   - 通知系统 (Notification System)
//   - 客户门户 (Client Portal)
//   - 风控核心 (Risk Control)
// ============================================================================

// ============================================================================
// 1. 待办事项系统 (Inbox System)
// ============================================================================

// InboxItem 统一收件箱待办事项
type InboxItem struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联信息
	UserID      uint   `json:"user_id" gorm:"not null;index:idx_user_unread,priority:1;index:idx_user_priority,priority:1"`
	SourceType  string `json:"source_type" gorm:"size:50;not null;comment:来源类型: deadline/approval/task"`
	SourceID    uint   `json:"source_id" gorm:"not null;comment:来源记录ID"`
	Title       string `json:"title" gorm:"size:255;not null"`
	Content     string `json:"content" gorm:"type:text"`

	// 优先级与时间
	Priority    string     `json:"priority" gorm:"size:20;not null;default:'medium';index:idx_user_priority,priority:2;comment:优先级: critical/high/medium/low"`
	DueDate     *time.Time `json:"due_date" gorm:"index:idx_user_unread,priority:3"`
	DueDateType string     `json:"due_date_type" gorm:"size:50;comment:日期类型: hearing/appeal/evidence等"`

	// 状态
	IsRead      bool       `json:"is_read" gorm:"default:false;index:idx_user_unread,priority:2"`
	ReadAt      *time.Time `json:"read_at"`
	IsCompleted bool       `json:"is_completed" gorm:"default:false;index:idx_user_priority,priority:3"`
	CompletedAt *time.Time `json:"completed_at"`

	// 提醒
	ReminderSent  bool       `json:"reminder_sent" gorm:"default:false"`
	ReminderCount int        `json:"reminder_count" gorm:"default:0"`
	Escalated     bool       `json:"escalated" gorm:"default:false;comment:是否已升级通知上级"`
	EscalatedAt   *time.Time `json:"escalated_at"`

	// 延后
	SnoozedUntil *time.Time `json:"snoozed_until" gorm:"comment:延后到何时提醒"`
	SnoozedCount int        `json:"snoozed_count" gorm:"default:0"`
}

func (InboxItem) TableName() string {
	return "inbox_items"
}

// InboxReminderRule 待办提醒规则
type InboxReminderRule struct {
	ID             uint     `json:"id" gorm:"primarykey"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	RuleName       string   `json:"rule_name" gorm:"size:100;not null"`
	DueDateType    string   `json:"due_date_type" gorm:"size:50;not null;uniqueIndex:uk_rule_type,priority:1;comment:日期类型"`
	Priority       string   `json:"priority" gorm:"size:20;not null;uniqueIndex:uk_rule_type,priority:2;comment:优先级"`
	ReminderOffsets JSONArray `json:"reminder_offsets" gorm:"type:json;not null;comment:提醒偏移量列表"`

	IsActive       bool     `json:"is_active" gorm:"default:true"`
}

func (InboxReminderRule) TableName() string {
	return "inbox_reminder_rules"
}

// ============================================================================
// 2. 利益冲突检测 (Conflict Detection)
// ============================================================================

// LawyerConflictPool 律师冲突检测预计算池
type LawyerConflictPool struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	LawyerID    uint   `json:"lawyer_id" gorm:"not null;index:idx_lawyer_entity,priority:1;comment:律师ID"`

	// 实体信息
	EntityType           string  `json:"entity_type" gorm:"size:20;not null;comment:实体类型: company/individual"`
	EntityName           string  `json:"entity_name" gorm:"size:200;not null"`
	EntityNameStandard   string  `json:"entity_name_standard" gorm:"size:200;not null;index:idx_lawyer_entity,priority:2;index:idx_entity_search,priority:1;comment:工商标准名称"`
	EntityTaxID          string  `json:"entity_tax_id" gorm:"size:50;index:idx_entity_search,priority:2"`
	EntityAliases        JSON    `json:"entity_aliases" gorm:"type:json;comment:别名列表"`

	// 关系信息
	RelationshipType     string  `json:"relationship_type" gorm:"size:50;not null;comment:关系: client/opposing/witness"`
	CaseID               uint    `json:"case_id" gorm:"not null;index:idx_case;comment:关联案件ID"`
	CaseTitle            string  `json:"case_title" gorm:"size:200"`

	// 股权穿透数据
	ShareholdingInfo     JSON    `json:"shareholding_info" gorm:"type:json;comment:股权穿透信息"`
	RelatedCompanies     JSON    `json:"related_companies" gorm:"type:json;comment:关联公司列表"`

	// 数据来源
	DataSource           string  `json:"data_source" gorm:"size:50;default:'manual';comment:数据来源: manual/api/import"`
	APIProvider          string  `json:"api_provider" gorm:"size:50;comment:API提供商: qichacha/tianyancha"`
	LastVerifiedAt       *time.Time `json:"last_verified_at"`
}

func (LawyerConflictPool) TableName() string {
	return "lawyer_conflict_pool"
}

// ConflictReport 冲突检测报告
type ConflictReport struct {
	ID           uint      `json:"id" gorm:"primarykey"`
	CreatedAt    time.Time `json:"created_at"`

	ReportNumber string    `json:"report_number" gorm:"size:50;uniqueIndex;not null;comment:报告编号"`

	// 检测信息
	CheckedBy    uint      `json:"checked_by" gorm:"not null;index:idx_checked_by;comment:检测人ID"`
	CheckTime    time.Time `json:"check_time" gorm:"not null;comment:检测时间"`
	CheckDurationMs *int   `json:"check_duration_ms" gorm:"comment:检测耗时(毫秒)"`

	// 检测对象
	ClientName   string    `json:"client_name" gorm:"size:200;not null;index:idx_client;comment:客户标准名称"`
	ClientTaxID  string    `json:"client_tax_id" gorm:"size:50;comment:客户税号"`
	OpposingParty string   `json:"opposing_party" gorm:"size:200;comment:对方当事人"`

	// 检测结果
	RiskLevel    string    `json:"risk_level" gorm:"size:20;not null;index:idx_risk_level;comment:风险等级: CRITICAL/HIGH/MEDIUM/LOW/PASS"`
	MatchedCases JSON      `json:"matched_cases" gorm:"type:json;comment:匹配的案件列表"`
	RelatedCompanies JSON  `json:"related_companies" gorm:"type:json;comment:关联公司列表"`
	ConflictDetails JSON   `json:"conflict_details" gorm:"type:json;comment:详细冲突信息"`

	// 报告文件
	ReportURL    string    `json:"report_url" gorm:"size:500;comment:PDF报告地址"`
	ReportGeneratedAt *time.Time `json:"report_generated_at"`

	// 审批信息
	ReviewedBy   *uint     `json:"reviewed_by" gorm:"comment:复核人ID"`
	ReviewedAt   *time.Time `json:"reviewed_at"`
	ReviewNotes  string    `json:"review_notes" gorm:"type:text;comment:复核意见"`
}

func (ConflictReport) TableName() string {
	return "conflict_reports"
}

// CompanyAPICall 工商API调用记录
type CompanyAPICall struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"created_at"`

	// API信息
	APIProvider string    `json:"api_provider" gorm:"size:50;not null;index:idx_provider,priority:2;comment:API提供商"`
	APIEndpoint string    `json:"api_endpoint" gorm:"size:200;not null"`
	RequestParams JSON   `json:"request_params" gorm:"type:json;comment:请求参数"`

	// 搜索信息
	SearchKeyword string  `json:"search_keyword" gorm:"size:200;not null;index:idx_search;comment:搜索关键词"`
	MatchedCompanyName string `json:"matched_company_name" gorm:"size:200;comment:匹配到的公司名称"`
	MatchedCompanyTaxID string `json:"matched_company_tax_id" gorm:"size:50;index:idx_company;comment:匹配到的税号"`

	// 响应信息
	ResponseStatus string `json:"response_status" gorm:"size:20;comment:响应状态: success/failed/partial"`
	ResponseData   JSON   `json:"response_data" gorm:"type:json;comment:响应数据"`
	ErrorMessage   string `json:"error_message" gorm:"type:text;comment:错误信息"`

	// 调用信息
	CalledBy       uint    `json:"called_by" gorm:"not null;comment:调用者ID"`
	CallDurationMs *int    `json:"call_duration_ms" gorm:"comment:调用耗时"`
}

func (CompanyAPICall) TableName() string {
	return "company_api_calls"
}

// ConflictScanJob 利益冲突定期扫描任务
type ConflictScanJob struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"created_at"`

	// 扫描配置
	ScanType    string    `json:"scan_type" gorm:"size:20;not null;index:idx_type_time,priority:1;comment:扫描类型: daily/weekly/manual"`
	ScanScope   string    `json:"scan_scope" gorm:"size:50;default:'all';comment:扫描范围: all/new_cases/lawyer"`

	// 扫描状态
	Status      string    `json:"status" gorm:"size:20;default:'pending';index:idx_status;comment:状态: pending/running/completed/failed"`

	// 扫描结果
	StartedAt       *time.Time `json:"started_at" gorm:"comment:开始时间"`
	CompletedAt     *time.Time `json:"completed_at" gorm:"comment:完成时间"`
	ScannedCases    int        `json:"scanned_cases" gorm:"default:0;comment:扫描的案件数"`
	ScannedLawyers  int        `json:"scanned_lawyers" gorm:"default:0;comment:扫描的律师数"`
	FoundConflicts  int        `json:"found_conflicts" gorm:"default:0;comment:发现的冲突数"`
	NewConflicts    JSON       `json:"new_conflicts" gorm:"type:json;comment:新发现的冲突列表"`

	// 告警状态
	TriggeredAlerts bool       `json:"triggered_alerts" gorm:"default:false;comment:是否已触发告警"`
	AlertSentAt     *time.Time `json:"alert_sent_at" gorm:"comment:告警发送时间"`

	// 触发信息
	TriggeredBy     *uint      `json:"triggered_by" gorm:"comment:触发者ID (manual时)"`
	TriggerReason   string     `json:"trigger_reason" gorm:"size:200;comment:触发原因"`

	ErrorMessage    string     `json:"error_message" gorm:"type:text;comment:错误信息"`
}

func (ConflictScanJob) TableName() string {
	return "conflict_scan_jobs"
}

// ============================================================================
// 3. 文档管理 (Document Management)
// ============================================================================

// DocumentLock 文档锁
type DocumentLock struct {
	DocumentID    uint      `json:"document_id" gorm:"primarykey;comment:文档ID"`
	LockedBy      uint      `json:"locked_by" gorm:"not null;index:idx_locked_by;comment:锁定人ID"`
	LockedAt      time.Time `json:"locked_at" gorm:"not null;comment:锁定时间"`
	ExpiresAt     time.Time `json:"expires_at" gorm:"not null;index:idx_expires;comment:过期时间"`
	ForceUnlock   bool      `json:"force_unlock" gorm:"default:false;comment:管理员强制解锁标记"`

	// 签出/签入模式
	IsCheckedOut  bool       `json:"is_checked_out" gorm:"default:false;comment:是否被签出(离线编辑)"`
	CheckedOutAt  *time.Time `json:"checked_out_at" gorm:"comment:签出时间"`
	CheckoutIP    string     `json:"checkout_ip" gorm:"size:45;comment:签出IP"`

	LastActivity  *time.Time `json:"last_activity" gorm:"comment:最后活动时间"`
}

func (DocumentLock) TableName() string {
	return "document_locks"
}

// DocumentVersionNew 文档版本历史 (新版本，避免与现有DocumentVersion冲突)
type DocumentVersionNew struct {
	ID               uint      `json:"id" gorm:"primarykey"`
	CreatedAt        time.Time `json:"created_at"`

	DocumentID       uint      `json:"document_id" gorm:"not null;uniqueIndex:uk_doc_version,priority:1;index:idx_doc_current,priority:1;comment:文档ID"`
	Version          int       `json:"version" gorm:"not null;uniqueIndex:uk_doc_version,priority:2;comment:版本号"`

	// 文件信息
	Filename         string    `json:"filename" gorm:"size:255;not null"`
	Filepath         string    `json:"filepath" gorm:"size:500;not null"`
	FileHash         string    `json:"file_hash" gorm:"size:64;not null;comment:文件SHA256哈希"`
	FileSize         int64     `json:"file_size" gorm:"not null;comment:文件大小(字节)"`
	MimeType         string    `json:"mime_type" gorm:"size:100"`

	// 版本信息
	CreatedBy        uint      `json:"created_by" gorm:"not null;index:idx_created_by;comment:创建者ID"`
	ChangeDescription string   `json:"change_description" gorm:"type:text;comment:变更说明"`
	ChangeType       string    `json:"change_type" gorm:"size:50;default:'manual';comment:变更类型: manual/checkout/auto"`

	// 状态
	IsCurrent        bool      `json:"is_current" gorm:"default:false;index:idx_doc_current,priority:2;comment:是否当前版本"`
}

func (DocumentVersionNew) TableName() string {
	return "document_versions"
}

// DocumentIndexQueue 文档全文索引队列
type DocumentIndexQueue struct {
	ID           uint      `json:"id" gorm:"primarykey"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	DocumentID   uint      `json:"document_id" gorm:"not null;index:idx_document;comment:文档ID"`
	FilePath     string    `json:"file_path" gorm:"size:500;not null"`
	MimeType     string    `json:"mime_type" gorm:"size:100"`

	// 索引状态
	Status       string    `json:"status" gorm:"size:20;default:'pending';index:idx_status;comment:状态: pending/processing/completed/failed"`
	IndexedAt    *time.Time `json:"indexed_at" gorm:"comment:索引完成时间"`
	ErrorMessage string    `json:"error_message" gorm:"type:text;comment:错误信息"`
	RetryCount   int       `json:"retry_count" gorm:"default:0;comment:重试次数"`

	// 索引内容摘要
	ContentPreview string   `json:"content_preview" gorm:"type:text;comment:内容预览(前500字符)"`
	WordCount      int      `json:"word_count" gorm:"comment:字数统计"`
}

func (DocumentIndexQueue) TableName() string {
	return "document_index_queue"
}

// CaseFolderTemplate 案件文件夹模板
type CaseFolderTemplate struct {
	ID             uint      `json:"id" gorm:"primarykey"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	Name           string    `json:"name" gorm:"size:100;not null"`
	Description    string    `json:"description" gorm:"type:text"`
	FolderStructure JSON     `json:"folder_structure" gorm:"type:json;not null;comment:文件夹结构定义"`

	// 分类
	CaseType       string    `json:"case_type" gorm:"size:50;index:idx_case_type;comment:适用案件类型"`
	IsDefault      bool      `json:"is_default" gorm:"default:false"`
	IsActive       bool      `json:"is_active" gorm:"default:true;index:idx_active"`

	// 模板文件
	TemplateFiles  JSON      `json:"template_files" gorm:"type:json;comment:模板文件列表"`

	CreatedBy      uint      `json:"created_by" gorm:"not null;comment:创建者ID"`
}

func (CaseFolderTemplate) TableName() string {
	return "case_folder_templates"
}

// ============================================================================
// 4. 财务管理 (Financial Management)
// ============================================================================

// Contract 合同
type Contract struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	ContractCode string   `json:"contract_code" gorm:"size:50;uniqueIndex;not null;comment:合同编号"`

	// 关联信息
	CaseID      *uint     `json:"case_id" gorm:"index:idx_case;comment:关联案件ID"`
	ClientID    uint      `json:"client_id" gorm:"not null;index:idx_client;comment:客户ID"`

	// 金额信息
	ContractAmount float64 `json:"contract_amount" gorm:"type:decimal(15,2);not null;comment:合同金额"`
	Currency    string    `json:"currency" gorm:"size:10;default:'CNY';comment:币种"`

	// 合同条款
	BillingCycle string   `json:"billing_cycle" gorm:"size:50;comment:计费周期:一次性/分期/按小时"`
	PaymentTerms string   `json:"payment_terms" gorm:"size:100;comment:付款条件"`
	StartDate   *time.Time `json:"start_date" gorm:"comment:开始日期"`
	EndDate     *time.Time `json:"end_date" gorm:"comment:结束日期"`

	// 合同状态
	Status      string    `json:"status" gorm:"size:20;default:'draft';index:idx_status;comment:状态: draft/active/suspended/completed/cancelled"`

	// 补充协议
	ParentContractID *uint `json:"parent_contract_id" gorm:"index:idx_parent;comment:主合同ID(补充协议时使用)"`
	ContractType     string `json:"contract_type" gorm:"size:20;default:'original';comment:合同类型: original/supplementary"`

	// 签署信息
	SignedAt    *time.Time `json:"signed_at" gorm:"comment:签署日期"`
	DocumentID  *uint      `json:"document_id" gorm:"comment:合同文档ID"`

	// GORM 关联（不存数据库）
	Case   *Case   `json:"case,omitempty" gorm:"foreignKey:CaseID"`
	Client *Client `json:"client,omitempty" gorm:"foreignKey:ClientID"`
}

func (Contract) TableName() string {
	return "contracts"
}

// PaymentMilestone 付款计划
type PaymentMilestone struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	ContractID  uint      `json:"contract_id" gorm:"not null;index:idx_contract;comment:合同ID"`
	Name        string    `json:"name" gorm:"size:200;not null;comment:里程碑名称"`
	Sequence    int       `json:"sequence" gorm:"not null;comment:顺序号"`
	Amount      float64   `json:"amount" gorm:"type:decimal(15,2);not null;comment:金额"`
	Percentage  float64   `json:"percentage" gorm:"type:decimal(5,2);comment:占比(%)"`
	DueDate     *time.Time `json:"due_date" gorm:"comment:到期日期"`
	Condition   string    `json:"condition" gorm:"type:text;comment:付款条件"`

	Status      string    `json:"status" gorm:"size:20;default:'pending';index:idx_status;comment:状态: pending/billed/paid/overdue"`
	InvoiceID   *uint     `json:"invoice_id" gorm:"index:idx_invoice;comment:关联发票ID"`
	PaidAmount  float64   `json:"paid_amount" gorm:"type:decimal(15,2);default:0;comment:已付金额"`

	// GORM 关联
	Contract *Contract `json:"contract,omitempty" gorm:"foreignKey:ContractID"`
}

func (PaymentMilestone) TableName() string {
	return "payment_milestones"
}

// Invoice 发票
type Invoice struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	InvoiceCode string    `json:"invoice_code" gorm:"size:50;uniqueIndex;not null;comment:发票编号"`

	// 关联信息
	ContractID  *uint     `json:"contract_id" gorm:"index:idx_contract;comment:合同ID"`
	MilestoneID *uint     `json:"milestone_id" gorm:"comment:付款计划ID"`
	ClientID    uint      `json:"client_id" gorm:"not null;index:idx_client;comment:客户ID"`

	// 金额信息
	Amount      float64   `json:"amount" gorm:"type:decimal(15,2);not null;comment:发票金额(不含税)"`
	TaxRate     float64   `json:"tax_rate" gorm:"type:decimal(5,2);default:0;comment:税率(%)"`
	TaxAmount   float64   `json:"tax_amount" gorm:"type:decimal(15,2);comment:税额"`
	TotalAmount float64   `json:"total_amount" gorm:"type:decimal(15,2);comment:价税合计"`

	// 客户开票信息快照
	ClientName        string  `json:"client_name" gorm:"size:200;not null;comment:客户名称"`
	ClientTaxID       string  `json:"client_tax_id" gorm:"size:50;comment:纳税人识别号"`
	ClientAddress     string  `json:"client_address" gorm:"type:text;comment:地址"`
	ClientBankName    string  `json:"client_bank_name" gorm:"size:100;comment:开户行"`
	ClientBankAccount string  `json:"client_bank_account" gorm:"size:50;comment:银行账号"`

	// 发票类型
	InvoiceType       string `json:"invoice_type" gorm:"size:20;default:'normal';comment:发票类型: normal/credit(红字)"`
	OriginalInvoiceID *uint  `json:"original_invoice_id" gorm:"index:idx_original;comment:原发票ID(红冲时)"`
	RefundReason      string `json:"refund_reason" gorm:"type:text;comment:退费原因"`
	WriteOffAmount    float64 `json:"write_off_amount" gorm:"type:decimal(15,2);comment:核销金额"`

	// 开票流程
	Status            string    `json:"status" gorm:"size:20;default:'draft';index:idx_status;comment:状态: draft/submitted/approved/issued/received/cancelled"`
	SubmittedAt       *time.Time `json:"submitted_at" gorm:"comment:提交时间"`
	ApprovedByFinanceAt *time.Time `json:"approved_by_finance_at" gorm:"comment:财务复核时间"`
	IssuedAt          *time.Time `json:"issued_at" gorm:"comment:开票时间"`
	ReceivedAt        *time.Time `json:"received_at" gorm:"comment:客户签收时间"`

	// 电子发票
	ElectronicInvoiceURL  string `json:"electronic_invoice_url" gorm:"size:500;comment:电子发票URL"`
	ElectronicInvoiceCode string `json:"electronic_invoice_code" gorm:"size:50;comment:发票代码"`
	ElectronicInvoiceNumber string `json:"electronic_invoice_number" gorm:"size:50;comment:发票号码"`

	// 审批信息
	CreatedBy   uint      `json:"created_by" gorm:"not null;comment:创建者ID"`
	SubmittedBy *uint     `json:"submitted_by" gorm:"comment:提交者ID"`
	ApprovedBy  *uint     `json:"approved_by" gorm:"comment:审批人ID"`

	// GORM 关联
	Client   *Client   `json:"client,omitempty" gorm:"foreignKey:ClientID"`
	Contract *Contract `json:"contract,omitempty" gorm:"foreignKey:ContractID"`
}

func (Invoice) TableName() string {
	return "invoices"
}

// Payment 回款记录
type Payment struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"created_at"`

	PaymentCode string    `json:"payment_code" gorm:"size:50;uniqueIndex;not null;comment:回款编号"`

	InvoiceID   uint      `json:"invoice_id" gorm:"not null;index:idx_invoice;comment:发票ID"`
	Amount      float64   `json:"amount" gorm:"type:decimal(15,2);not null;comment:回款金额"`
	PaymentDate time.Time `json:"payment_date" gorm:"not null;index:idx_date;comment:付款日期"`

	// 付款方式
	PaymentMethod string  `json:"payment_method" gorm:"size:50;comment:付款方式: bank_transfer/cash/other"`
	ReferenceNo   string  `json:"reference_no" gorm:"size:100;comment:银行流水号"`
	PayerName     string  `json:"payer_name" gorm:"size:200;comment:付款人"`
	PayerAccount  string  `json:"payer_account" gorm:"size:100;comment:付款账号"`

	// 凭证
	AttachmentID *uint    `json:"attachment_id" gorm:"comment:回款凭证ID"`

	// 确认信息
	ConfirmedBy  uint     `json:"confirmed_by" gorm:"not null;comment:确认人ID"`
	ConfirmedAt  *time.Time `json:"confirmed_at" gorm:"comment:确认时间"`
	Status       string   `json:"status" gorm:"size:20;default:'confirmed';index:idx_status;comment:状态: pending/confirmed/rejected"`

	Remark       string   `json:"remark" gorm:"type:text;comment:备注"`

	// GORM 关联
	Invoice *Invoice `json:"invoice,omitempty" gorm:"foreignKey:InvoiceID"`
}

func (Payment) TableName() string {
	return "payments"
}

// BadDebtRecord 坏账核销记录
type BadDebtRecord struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联信息
	ContractID  uint      `json:"contract_id" gorm:"not null;index:idx_contract;comment:合同ID"`
	InvoiceID   *uint     `json:"invoice_id" gorm:"comment:发票ID"`

	// 金额信息
	OriginalAmount float64 `json:"original_amount" gorm:"type:decimal(15,2);not null;comment:原始应收金额"`
	WriteOffAmount float64 `json:"write_off_amount" gorm:"type:decimal(15,2);not null;comment:核销金额"`
	RemainingAmount float64 `json:"remaining_amount" gorm:"type:decimal(15,2);comment:剩余金额"`

	// 核销原因
	Reason      string    `json:"reason" gorm:"type:text;not null;comment:核销原因"`
	ReasonType  string    `json:"reason_type" gorm:"size:50;comment:原因类型: bankruptcy/dispute/uncollectible/other"`

	// 审批流程
	Status      string    `json:"status" gorm:"size:20;default:'pending';index:idx_status;comment:状态: pending/approved/rejected"`
	ApprovedBy  *uint     `json:"approved_by" gorm:"comment:审批人ID"`
	ApprovedAt  *time.Time `json:"approved_at" gorm:"comment:审批时间"`
	ApprovalNotes string  `json:"approval_notes" gorm:"type:text;comment:审批意见"`

	// 附件
	AttachmentIDs JSON     `json:"attachment_ids" gorm:"type:json;comment:证明材料ID列表"`

	// GORM 关联
	Contract *Contract `json:"contract,omitempty" gorm:"foreignKey:ContractID"`
}

func (BadDebtRecord) TableName() string {
	return "bad_debt_records"
}

// CommissionRecord 提成记录
type CommissionRecord struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	CommissionCode string  `json:"commission_code" gorm:"size:50;uniqueIndex;not null;comment:提成编号"`

	// 关联信息
	ContractID  uint      `json:"contract_id" gorm:"not null;index:idx_contract;comment:合同ID"`
	PaymentID   uint      `json:"payment_id" gorm:"not null;index:idx_payment;comment:回款ID"`
	CaseID      *uint     `json:"case_id" gorm:"comment:案件ID"`

	// 受益人信息
	BeneficiaryID   uint    `json:"beneficiary_id" gorm:"not null;index:idx_beneficiary,priority:1;comment:受益人ID"`
	BeneficiaryRole string  `json:"beneficiary_role" gorm:"size:50;not null;index:idx_beneficiary,priority:2;comment:角色: source/lawyer/assistant"`

	// 金额计算
	PaymentAmount   float64 `json:"payment_amount" gorm:"type:decimal(15,2);not null;comment:回款金额"`
	CostDeduction   float64 `json:"cost_deduction" gorm:"type:decimal(15,2);default:0;comment:成本扣除"`
	CommissionBase  float64 `json:"commission_base" gorm:"type:decimal(15,2);not null;comment:提成基数"`
	CommissionRate  float64 `json:"commission_rate" gorm:"type:decimal(5,2);not null;comment:提成比例(%)"`
	CommissionAmount float64 `json:"commission_amount" gorm:"type:decimal(15,2);not null;comment:提成金额"`

	// 计算/支付状态
	CalculatedAt    *time.Time `json:"calculated_at" gorm:"comment:计算时间"`
	Status          string    `json:"status" gorm:"size:20;default:'pending';index:idx_beneficiary,priority:3;comment:状态: pending/calculated/paid/cancelled"`

	// 支付信息
	PaidDate        *time.Time `json:"paid_date" gorm:"comment:支付日期"`
	PaymentVoucher  string    `json:"payment_voucher" gorm:"size:100;comment:支付凭证号"`

	// GORM 关联
	Contract *Contract `json:"contract,omitempty" gorm:"foreignKey:ContractID"`
	Payment  *Payment  `json:"payment,omitempty" gorm:"foreignKey:PaymentID"`
	Case     *Case     `json:"case,omitempty" gorm:"foreignKey:CaseID"`
}

func (CommissionRecord) TableName() string {
	return "commission_records"
}

// CommissionRule 分成规则
type CommissionRule struct {
	ID              uint    `json:"id" gorm:"primaryKey"`
	Name            string  `json:"name" gorm:"size:100;not null;comment:规则名称"`
	Role            string  `json:"role" gorm:"size:50;not null;comment:适用角色(source/lawyer/assistant)"`
	MinAmount       float64 `json:"min_amount" gorm:"type:decimal(15,2);default:0;comment:最小金额"`
	MaxAmount       float64 `json:"max_amount" gorm:"type:decimal(15,2);default:0;comment:最大金额(0=不限)"`
	BaseRate        float64 `json:"base_rate" gorm:"type:decimal(5,2);not null;comment:基础提成比例(%)"`
	PerformanceRate float64 `json:"performance_rate" gorm:"type:decimal(5,2);default:0;comment:绩效提成比例(%)"`
	Priority        int     `json:"priority" gorm:"default:0;comment:优先级(越大越优先)"`
	Active          bool    `json:"active" gorm:"default:true;comment:是否启用"`
	EffectiveDate   *string    `json:"effective_date" gorm:"comment:生效日期"`
	ExpiryDate      *string    `json:"expiry_date" gorm:"comment:失效日期"`
	CreatedAt       time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (CommissionRule) TableName() string {
	return "commission_rules"
}

// ============================================================================
// 5. 通知系统 (Notification System)
// ============================================================================

// NotificationQueue 通知预览队列
type NotificationQueue struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"created_at"`

	// 触发信息
	TriggerType string    `json:"trigger_type" gorm:"size:50;not null;index:idx_trigger,priority:1;comment:触发类型"`
	TriggerID   uint      `json:"trigger_id" gorm:"not null;index:idx_trigger,priority:2;comment:触发记录ID"`
	CaseID      *uint     `json:"case_id" gorm:"comment:关联案件ID"`

	// 接收人信息
	RecipientType string   `json:"recipient_type" gorm:"size:20;not null;index:idx_recipient,priority:1;comment:接收人类型: client/lawyer/admin"`
	RecipientID   uint     `json:"recipient_id" gorm:"not null;index:idx_recipient,priority:2;comment:接收人ID"`
	RecipientName string   `json:"recipient_name" gorm:"size:100;not null;comment:接收人姓名"`
	RecipientContact string `json:"recipient_contact" gorm:"size:200;comment:联系方式(邮箱/手机/OpenID)"`

	// 通知内容
	Channel     string    `json:"channel" gorm:"size:20;not null;comment:通知渠道: email/sms/wechat"`
	Subject     string    `json:"subject" gorm:"size:200;comment:标题(邮件等)"`
	Content     string    `json:"content" gorm:"type:text;not null;comment:通知内容"`
	TemplateID  string    `json:"template_id" gorm:"size:50;comment:模板ID"`

	// 状态
	Status      string    `json:"status" gorm:"size:20;default:'pending';index:idx_status;comment:状态: pending/approved/sent/cancelled/failed"`
	Priority    string    `json:"priority" gorm:"size:20;default:'normal';comment:优先级: urgent/normal/low"`

	// 审核信息
	CreatedBy   uint      `json:"created_by" gorm:"not null;comment:创建者ID"`
	ApprovedBy  *uint     `json:"approved_by" gorm:"comment:审批人ID"`
	ApprovedAt  *time.Time `json:"approved_at" gorm:"comment:审批时间"`

	// 发送信息
	SentAt              *time.Time `json:"sent_at" gorm:"comment:发送时间"`
	SentRetryCount      int        `json:"sent_retry_count" gorm:"default:0;comment:重试次数"`
	ErrorMessage        string     `json:"error_message" gorm:"type:text;comment:错误信息"`

	// 敏感信息标记
	ContainsSensitiveInfo bool      `json:"contains_sensitive_info" gorm:"default:false;comment:包含敏感信息"`
	AutoSend              bool      `json:"auto_send" gorm:"default:false;comment:是否自动发送"`

	// 外部消息ID
	ExternalMessageID string `json:"external_message_id" gorm:"size:100;comment:外部消息ID(如微信msg_id)"`

	// 关联
	CreatedByUser  *User `json:"created_by_user,omitempty" gorm:"foreignKey:CreatedBy"`
	ApprovedByUser *User `json:"approved_by_user,omitempty" gorm:"foreignKey:ApprovedBy"`
	Case           *Case `json:"case,omitempty" gorm:"foreignKey:CaseID"`
}

func (NotificationQueue) TableName() string {
	return "notification_queue"
}

// NotificationTemplate 通知模板
type NotificationTemplate struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	TemplateCode string   `json:"template_code" gorm:"size:50;uniqueIndex;not null;comment:模板代码"`
	TemplateName string   `json:"template_name" gorm:"size:100;not null"`

	// 适用信息
	Channel      string   `json:"channel" gorm:"size:20;not null;index:idx_channel_event,priority:1;comment:适用渠道: email/sms/wechat"`
	RecipientType string  `json:"recipient_type" gorm:"size:20;not null;comment:接收人类型: client/lawyer/admin"`
	TriggerEvent string   `json:"trigger_event" gorm:"size:100;not null;index:idx_channel_event,priority:2;comment:触发事件"`

	// 模板内容
	SubjectTemplate string `json:"subject_template" gorm:"size:200;comment:标题模板"`
	ContentTemplate string `json:"content_template" gorm:"type:text;not null;comment:内容模板(支持变量替换)"`

	// 变量定义
	Variables    JSON     `json:"variables" gorm:"type:json;comment:可用变量列表"`

	// 自动发送规则
	AutoSend     bool     `json:"auto_send" gorm:"default:false;comment:是否自动发送"`
	RequiresApproval bool `json:"requires_approval" gorm:"default:true;comment:是否需要审批"`

	IsActive     bool     `json:"is_active" gorm:"default:true"`
}

func (NotificationTemplate) TableName() string {
	return "notification_templates"
}

// SensitiveWord 敏感词库
type SensitiveWord struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Word        string    `json:"word" gorm:"size:200;not null;uniqueIndex;comment:敏感词"`
	WordType    string    `json:"word_type" gorm:"size:50;not null;index:idx_type;comment:词类型: political/porn/violence/other"`
	Category    string    `json:"category" gorm:"size:50;index:idx_category;comment:分类"`
	Severity    string    `json:"severity" gorm:"size:20;default:'medium';comment:严重程度: low/medium/high/critical"`
	Replacement string    `json:"replacement" gorm:"size:200;comment:替换词"`
	IsActive    bool      `json:"is_active" gorm:"default:true;index:idx_active;comment:是否启用"`
	Description string    `json:"description" gorm:"type:text;comment:说明"`

	// 统计信息
	HitCount    int       `json:"hit_count" gorm:"default:0;comment:命中次数"`
	LastHitAt   *time.Time `json:"last_hit_at" gorm:"comment:最后命中时间"`

	// 操作信息
	CreatedBy   uint      `json:"created_by" gorm:"not null;comment:创建者ID"`
	UpdatedBy   uint      `json:"updated_by" gorm:"not null;comment:更新者ID"`
}

func (SensitiveWord) TableName() string {
	return "sensitive_words"
}

// ContentFilterLog 内容过滤日志
type ContentFilterLog struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"created_at"`

	// 来源信息
	ContentType string    `json:"content_type" gorm:"size:50;not null;index:idx_content_type;comment:内容类型: notification/email/document"`
	ContentID   uint      `json:"content_id" gorm:"not null;index:idx_content;comment:内容ID"`
	OriginalContent string `json:"original_content" gorm:"type:text;comment:原始内容"`

	// 过滤结果
	FilteredContent string `json:"filtered_content" gorm:"type:text;comment:过滤后内容"`
	Hits            JSON    `json:"hits" gorm:"type:json;comment:命中详情"`
	IsBlocked       bool    `json:"is_blocked" gorm:"default:false;comment:是否被拦截"`
	RequiresApproval bool  `json:"requires_approval" gorm:"default:false;comment:是否需要审批"`

	// 操作信息
	ProcessedBy uint     `json:"processed_by" gorm:"comment:处理者ID"`
	ActionTaken string   `json:"action_taken" gorm:"size:50;comment:采取的行动: blocked/replaced/approved"`
}

func (ContentFilterLog) TableName() string {
	return "content_filter_logs"
}

// ============================================================================
// 6. 客户门户 (Client Portal)
// ============================================================================

// ClientAccount 客户门户账户
type ClientAccount struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	ClientID    uint      `json:"client_id" gorm:"not null;uniqueIndex:idx_client;comment:客户ID"`

	// 账户信息
	Username    string    `json:"username" gorm:"size:50;uniqueIndex;not null;comment:用户名(手机号)"`
	PasswordHash string   `json:"password_hash" gorm:"size:255;not null"`
	Phone       string    `json:"phone" gorm:"size:20;uniqueIndex"`

	// 状态
	Status      string    `json:"status" gorm:"size:20;default:'active';comment:状态: active/disabled"`
	LastLoginAt *time.Time `json:"last_login_at" gorm:"comment:最后登录时间"`
	LastLoginIP string    `json:"last_login_ip" gorm:"size:45;comment:最后登录IP"`

	// 微信绑定
	WechatOpenID string  `json:"wechat_openid" gorm:"size:100;uniqueIndex:idx_wechat;comment:微信OpenID"`
	WechatUnionID string `json:"wechat_unionid" gorm:"size:100;comment:微信UnionID"`
	WechatNickname string `json:"wechat_nickname" gorm:"size:100;comment:微信昵称"`
	WechatBoundAt *time.Time `json:"wechat_bound_at" gorm:"comment:绑定时间"`

	// 授权控制(白名单模式)
	AuthorizedCases JSON   `json:"authorized_cases" gorm:"type:json;comment:授权可见的案件ID列表"`

	// 安全设置
	PasswordChangedAt *time.Time `json:"password_changed_at" gorm:"comment:密码修改时间"`
	FailedLoginCount  int       `json:"failed_login_count" gorm:"default:0;comment:失败登录次数"`
	LockedUntil       *time.Time `json:"locked_until" gorm:"comment:锁定至"`
}

func (ClientAccount) TableName() string {
	return "client_accounts"
}

// WechatInvitation 微信邀请记录
type WechatInvitation struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"created_at"`

	InviteToken string    `json:"invite_token" gorm:"size:100;uniqueIndex:idx_token;not null;comment:邀请token"`
	ClientID    uint      `json:"client_id" gorm:"not null;index:idx_client;comment:客户ID"`
	InvitedBy   uint      `json:"invited_by" gorm:"not null;comment:邀请人ID(律师)"`

	// 授权范围
	AuthorizedCases JSON   `json:"authorized_cases" gorm:"type:json;comment:授权的案件列表"`

	// 状态
	Status      string    `json:"status" gorm:"size:20;default:'pending';index:idx_status;comment:状态: pending/accepted/expired/cancelled"`

	// 时间
	ExpiresAt   time.Time `json:"expires_at" gorm:"not null;comment:过期时间"`
	AcceptedAt  *time.Time `json:"accepted_at" gorm:"comment:接受时间"`
	WechatOpenID string   `json:"wechat_openid" gorm:"size:100;comment:绑定的OpenID"`
}

func (WechatInvitation) TableName() string {
	return "wechat_invitations"
}

// ============================================================================
// 7. 风控核心 (Risk Control Critical Scenarios)
// ============================================================================

// CaseEthicalWallWhitelist 隔离墙白名单
type CaseEthicalWallWhitelist struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"created_at"`

	CaseID    uint      `json:"case_id" gorm:"not null;uniqueIndex:uk_case_user,priority:1;index:idx_case;comment:案件ID"`
	UserID    uint      `json:"user_id" gorm:"not null;uniqueIndex:uk_case_user,priority:2;index:idx_user;comment:用户ID"`
	GrantedBy uint      `json:"granted_by" gorm:"not null;comment:授权人ID"`
	GrantedAt time.Time `json:"granted_at" gorm:"default:CURRENT_TIMESTAMP;comment:授权时间"`
	Reason     string    `json:"reason" gorm:"type:text;comment:授权理由"`

	// 关联
	User         *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	GrantedByUser *User `json:"granted_by_user,omitempty" gorm:"foreignKey:GrantedBy"`
	Case         *Case `json:"case,omitempty" gorm:"foreignKey:CaseID"`
}

func (CaseEthicalWallWhitelist) TableName() string {
	return "case_ethical_wall_whitelist"
}

// EthicalWallAccessLog 隔离墙访问日志
type EthicalWallAccessLog struct {
	ID           uint      `json:"id" gorm:"primarykey"`
	CaseID       uint      `json:"case_id" gorm:"not null;index:idx_case;comment:案件ID"`
	UserID       uint      `json:"user_id" gorm:"not null;index:idx_user;comment:用户ID"`
	AccessType   string    `json:"access_type" gorm:"size:20;not null;comment:访问类型: view/search/export"`
	AccessResult string    `json:"access_result" gorm:"size:20;not null;index:idx_result;comment:访问结果: allowed/denied"`
	IPAddress    string    `json:"ip_address" gorm:"size:45;comment:IP地址"`
	UserAgent    string    `json:"user_agent" gorm:"type:text;comment:User-Agent"`
	AttemptedAt  time.Time `json:"attempted_at" gorm:"default:CURRENT_TIMESTAMP;comment:尝试时间"`

	// 关联
	User         *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Case         *Case `json:"case,omitempty" gorm:"foreignKey:CaseID"`
}

func (EthicalWallAccessLog) TableName() string {
	return "ethical_wall_access_logs"
}

// ClientTrustAccount 客户代管账户
type ClientTrustAccount struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	ClientID    uint      `json:"client_id" gorm:"not null;index:idx_client;comment:客户ID"`
	AccountCode string    `json:"account_code" gorm:"size:50;uniqueIndex;not null;comment:代管账户编号"`

	// 余额信息
	Balance     float64   `json:"balance" gorm:"type:decimal(15,2);default:0;comment:当前余额"`
	Currency    string    `json:"currency" gorm:"size:10;default:'CNY';comment:币种"`
	FrozenAmount float64  `json:"frozen_amount" gorm:"type:decimal(15,2);default:0;comment:冻结金额"`

	// 资金用途限制
	PurposeRestriction string `json:"purpose_restriction" gorm:"size:200;comment:资金用途说明"`
	AuthorizedUses     JSON   `json:"authorized_uses" gorm:"type:json;comment:授权用途列表"`

	// 状态
	Status      string    `json:"status" gorm:"size:20;default:'active';index:idx_status;comment:状态: active/frozen/closed"`
	OpenedAt    *time.Time `json:"opened_at" gorm:"comment:开户时间"`
	ClosedAt    *time.Time `json:"closed_at" gorm:"comment:销户时间"`
}

func (ClientTrustAccount) TableName() string {
	return "client_trust_accounts"
}

// ClientTrustTransaction 代管款交易记录
type ClientTrustTransaction struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	AccountID   uint      `json:"account_id" gorm:"not null;index:idx_account;comment:账户ID"`
	TransactionCode string `json:"transaction_code" gorm:"size:50;uniqueIndex;not null;comment:交易编号"`

	// 交易信息
	TransactionType string `json:"transaction_type" gorm:"size:20;not null;index:idx_type;comment:交易类型: deposit/deposit_refund/withdraw/transfer"`
	Amount      float64   `json:"amount" gorm:"type:decimal(15,2);not null;comment:金额"`
	Description string    `json:"description" gorm:"type:text;not null;comment:交易说明"`

	// 用途关联
	CaseID      *uint     `json:"case_id" gorm:"index:idx_case;comment:关联案件ID"`
	PurposeCode string    `json:"purpose_code" gorm:"size:50;comment:用途代码: court_fee/evidence_fee/investigation_fee等"`

	// 支出信息
	RecipientName     string `json:"recipient_name" gorm:"size:200;comment:收款方名称"`
	RecipientBankAccount string `json:"recipient_bank_account" gorm:"size:50;comment:收款账号"`
	RecipientBankName string `json:"recipient_bank_name" gorm:"size:100;comment:收款银行"`

	// 状态
	Status      string    `json:"status" gorm:"size:20;default:'pending';index:idx_status;comment:状态: pending/completed/cancelled"`
	CompletedAt *time.Time `json:"completed_at" gorm:"comment:完成时间"`
	AttachmentID *uint    `json:"attachment_id" gorm:"comment:凭证附件ID"`

	// 审计信息
	CreatedBy   uint      `json:"created_by" gorm:"not null;comment:创建者ID"`
	ApprovedBy  *uint     `json:"approved_by" gorm:"comment:审批人ID"`
	ApprovedAt  *time.Time `json:"approved_at" gorm:"comment:审批时间"`
}

func (ClientTrustTransaction) TableName() string {
	return "client_trust_transactions"
}

// OffboardingRecord 离职交接记录
type OffboardingRecord struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"created_at"`

	UserID      uint      `json:"user_id" gorm:"not null;index:idx_user;comment:离职用户ID"`
	InitiatedBy uint      `json:"initiated_by" gorm:"not null;comment:发起人ID"`
	InitiatedAt time.Time `json:"initiated_at" gorm:"default:CURRENT_TIMESTAMP;comment:发起时间"`

	// 案件移交
	OriginalCases       JSON       `json:"original_cases" gorm:"type:json;not null;comment:原主办案件列表"`
	NewLawyerID         *uint      `json:"new_lawyer_id" gorm:"comment:接收律师ID"`
	CaseTransferCompletedAt *time.Time `json:"case_transfer_completed_at" gorm:"comment:案件移交完成时间"`

	// 待办事项移交
	OriginalInboxItems  JSON       `json:"original_inbox_items" gorm:"type:json;not null;comment:原待办事项列表"`
	InboxTransferCompletedAt *time.Time `json:"inbox_transfer_completed_at" gorm:"comment:待办移交完成时间"`

	// 文档处理
	DocumentDisposalMethod string   `json:"document_disposal_method" gorm:"size:50;comment:文档处理方式: delete/transfer/revoke_access"`
	DocumentDisposalCompletedAt *time.Time `json:"document_disposal_completed_at" gorm:"comment:文档处理完成时间"`

	// 财务结算
	SettlementCalculated bool      `json:"settlement_calculated" gorm:"default:false;comment:是否计算提成"`
	SettlementAmount    float64   `json:"settlement_amount" gorm:"type:decimal(15,2);comment:结算金额"`
	SettlementPaid      bool      `json:"settlement_paid" gorm:"default:false;comment:是否已支付"`

	// 状态
	Status      string    `json:"status" gorm:"size:20;default:'pending';index:idx_status;comment:状态: pending/in_progress/completed/cancelled"`
	Notes       string    `json:"notes" gorm:"type:text;comment:备注"`

	CompletedAt *time.Time `json:"completed_at" gorm:"comment:完成时间"`
}

func (OffboardingRecord) TableName() string {
	return "offboarding_records"
}

// OffboardingTransferDetail 离职交接详情
type OffboardingTransferDetail struct {
	ID           uint      `json:"id" gorm:"primarykey"`
	OffboardingID uint      `json:"offboarding_id" gorm:"not null;index:idx_offboarding;comment:交接记录ID"`
	TransferType string    `json:"transfer_type" gorm:"size:50;not null;comment:移交类型: case/inbox/document"`

	// 原所有者
	OriginalOwnerID uint     `json:"original_owner_id" gorm:"not null;comment:原所有者ID"`

	// 新所有者
	NewOwnerID   *uint     `json:"new_owner_id" gorm:"comment:新所有者ID"`

	// 移交内容
	ItemID       uint      `json:"item_id" gorm:"not null;comment:项目ID(案件ID/待办ID等)"`
	ItemName     string    `json:"item_name" gorm:"size:200;not null;comment:项目名称"`

	// 状态
	TransferStatus string   `json:"transfer_status" gorm:"size:20;default:'pending';index:idx_status;comment:状态: pending/completed/failed"`
	TransferredAt   *time.Time `json:"transferred_at" gorm:"comment:移交时间"`
	ErrorMessage    string    `json:"error_message" gorm:"type:text;comment:错误信息"`
}

func (OffboardingTransferDetail) TableName() string {
	return "offboarding_transfer_details"
}

// TokenRevocationLog 令牌撤销记录
type TokenRevocationLog struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	UserID      uint      `json:"user_id" gorm:"not null;index:idx_user;comment:用户ID"`
	RevocationType string `json:"revocation_type" gorm:"size:50;not null;index:idx_type;comment:撤销类型: offboarding/password_reset/manual"`
	RevokedBy   *uint     `json:"revoked_by" gorm:"comment:撤销操作人ID"`

	// 撤销范围
	RevokeAll   bool      `json:"revoke_all" gorm:"default:true;comment:是否撤销所有令牌"`
	RevokedTokens JSON    `json:"revoked_tokens" gorm:"type:json;comment:撤销的令牌列表"`

	// 令牌信息
	TokenType   string    `json:"token_type" gorm:"size:50;comment:令牌类型: access_token/refresh_token/api_key"`
	RevokedAt   time.Time `json:"revoked_at" gorm:"default:CURRENT_TIMESTAMP;comment:撤销时间"`
	IPAddress   string    `json:"ip_address" gorm:"size:45;comment:操作IP"`
}

func (TokenRevocationLog) TableName() string {
	return "token_revocation_logs"
}

// ============================================================================
// 8. 数据迁移支持 (Migration Support)
// ============================================================================

// DataImportTask 数据导入任务
type DataImportTask struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"created_at"`

	TaskCode    string    `json:"task_code" gorm:"size:50;uniqueIndex;not null;comment:任务编号"`

	// 导入信息
	ImportType  string    `json:"import_type" gorm:"size:50;not null;index:idx_type;comment:导入类型: client/case/contract/document"`
	FilePath    string    `json:"file_path" gorm:"size:500;comment:导入文件路径"`
	TotalRows   int       `json:"total_rows" gorm:"default:0;comment:总行数"`

	// 进度
	ProcessedRows int      `json:"processed_rows" gorm:"default:0;comment:已处理行数"`
	SuccessRows   int      `json:"success_rows" gorm:"default:0;comment:成功行数"`
	FailedRows    int      `json:"failed_rows" gorm:"default:0;comment:失败行数"`

	// 状态
	Status      string    `json:"status" gorm:"size:20;default:'pending';index:idx_status;comment:状态: pending/processing/completed/failed"`

	// 结果
	ErrorSummary JSON     `json:"error_summary" gorm:"type:json;comment:错误汇总"`
	ResultSummary JSON    `json:"result_summary" gorm:"type:json;comment:结果汇总"`

	// 操作信息
	CreatedBy   uint      `json:"created_by" gorm:"not null;comment:创建者ID"`
	StartedAt   *time.Time `json:"started_at" gorm:"comment:开始时间"`
	CompletedAt *time.Time `json:"completed_at" gorm:"comment:完成时间"`
}

func (DataImportTask) TableName() string {
	return "data_import_tasks"
}

// DataImportError 数据导入错误明细
type DataImportError struct {
	ID           uint      `json:"id" gorm:"primarykey"`
	CreatedAt    time.Time `json:"created_at"`

	ImportTaskID uint      `json:"import_task_id" gorm:"not null;index:idx_task;comment:导入任务ID"`
	RowNumber    int       `json:"row_number" gorm:"not null;comment:行号"`
	RowData      JSON      `json:"row_data" gorm:"type:json;comment:行数据"`
	ErrorMessage string    `json:"error_message" gorm:"type:text;not null;comment:错误信息"`
	ErrorType    string    `json:"error_type" gorm:"size:50;comment:错误类型"`
}

func (DataImportError) TableName() string {
	return "data_import_errors"
}

// ============================================================================
// 9. User 表扩展字段 (v2.2.0)
// ============================================================================

// UserExtension 用户扩展字段 (通过 embedding 使用)
// 这些字段已通过迁移添加到 users 表
type UserExtension struct {
	SupervisorID     *uint    `json:"supervisor_id" gorm:"column:supervisor_id;index:idx_supervisor;comment:上级/合伙人ID"`
	CommissionRate   float64  `json:"commission_rate" gorm:"column:commission_rate;type:decimal(5,2);comment:提成比例(%)"`
	RoleType         string   `json:"role_type" gorm:"column:role_type;size:50;comment:角色类型: source/lawyer/assistant"`
	OffboardingStatus string  `json:"offboarding_status" gorm:"column:offboarding_status;size:20;default:'active';comment:离职状态: active/offboarding/deactivated"`
}

// ============================================================================
// 10. Case 表扩展字段 (v2.2.0)
// ============================================================================

// CaseExtension 案件扩展字段 (隔离墙相关)
// 这些字段已通过迁移添加到 cases 表
type CaseExtension struct {
	EthicalWallEnabled    bool       `json:"ethical_wall_enabled" gorm:"column:ethical_wall_enabled;default:false;index:idx_ethical_wall;comment:是否启用隔离墙"`
	EthicalWallDescription string   `json:"ethical_wall_description" gorm:"column:ethical_wall_description;type:text;comment:隔离墙说明"`
	EthicalWallEnabledBy  *uint     `json:"ethical_wall_enabled_by" gorm:"column:ethical_wall_enabled_by;comment:启用人ID"`
	EthicalWallEnabledAt  *time.Time `json:"ethical_wall_enabled_at" gorm:"column:ethical_wall_enabled_at;comment:启用时间"`
}

// ============================================================================
// 初始化数据函数
// ============================================================================

// DefaultInboxReminderRules 返回默认的待办提醒规则
// 提醒偏移量说明: 负数表示提前天数，正数表示延后天数，0 表示当天
// 存储格式为 JSON 数组字符串，如 "[-30, -15, -7, -3, -1, 0]"
func DefaultInboxReminderRules() []InboxReminderRule {
	return []InboxReminderRule{
		{
			RuleName:        "上诉期限提醒",
			DueDateType:     "appeal_deadline",
			Priority:        "critical",
			ReminderOffsets: FromIntSlice([]int{-30, -15, -7, -3, -1, 0}),
			IsActive:        true,
		},
		{
			RuleName:        "举证期限提醒",
			DueDateType:     "evidence_deadline",
			Priority:        "critical",
			ReminderOffsets: FromIntSlice([]int{-14, -7, -3, -1, 0}),
			IsActive:        true,
		},
		{
			RuleName:        "诉讼时效提醒",
			DueDateType:     "statute_of_limitations",
			Priority:        "critical",
			ReminderOffsets: FromIntSlice([]int{-90, -30, -7, -1, 0}),
			IsActive:        true,
		},
		{
			RuleName:        "执行申请期限提醒",
			DueDateType:     "execution_deadline",
			Priority:        "critical",
			ReminderOffsets: FromIntSlice([]int{-180, -90, -30, -7, -1}),
			IsActive:        true,
		},
		{
			RuleName:        "开庭日期提醒",
			DueDateType:     "hearing",
			Priority:        "important",
			ReminderOffsets: FromIntSlice([]int{-7, -3, -1}),
			IsActive:        true,
		},
		{
			RuleName:        "庭前会议提醒",
			DueDateType:     "pretrial_conference",
			Priority:        "important",
			ReminderOffsets: FromIntSlice([]int{-3, -1}),
			IsActive:        true,
		},
		{
			RuleName:        "调查取证提醒",
			DueDateType:     "investigation",
			Priority:        "important",
			ReminderOffsets: FromIntSlice([]int{-3, -1}),
			IsActive:        true,
		},
		{
			RuleName:        "缴费期限提醒",
			DueDateType:     "payment",
			Priority:        "normal",
			ReminderOffsets: FromIntSlice([]int{-3, -1}),
			IsActive:        true,
		},
		{
			RuleName:        "结案归档提醒",
			DueDateType:     "case_closing",
			Priority:        "normal",
			ReminderOffsets: FromIntSlice([]int{-7, -3, -1}),
			IsActive:        true,
		},
	}
}

// DefaultNotificationTemplates 返回默认的通知模板
func DefaultNotificationTemplates() []NotificationTemplate {
	return []NotificationTemplate{
		{
			TemplateCode:     "SYSTEM_MAINTENANCE",
			TemplateName:     "系统维护通知",
			Channel:          "email",
			RecipientType:    "lawyer",
			TriggerEvent:     "system_maintenance",
			SubjectTemplate:  "系统维护通知",
			ContentTemplate:  "尊敬的{name}，系统将于{start_time}至{end_time}进行维护，届时将暂停服务。",
			Variables:        JSON(map[string]interface{}{"variables": []string{"name", "start_time", "end_time"}}),
			AutoSend:         true,
			RequiresApproval: false,
			IsActive:         true,
		},
		{
			TemplateCode:     "PAYMENT_RECEIVED",
			TemplateName:     "收款确认通知",
			Channel:          "wechat",
			RecipientType:    "client",
			TriggerEvent:     "payment_received",
			SubjectTemplate:  "收款确认",
			ContentTemplate:  "尊敬的客户，我们已收到您的付款{amount}元，感谢您的配合。",
			Variables:        JSON(map[string]interface{}{"variables": []string{"amount", "payment_date"}}),
			AutoSend:         true,
			RequiresApproval: false,
			IsActive:         true,
		},
		{
			TemplateCode:     "CASE_HEARING",
			TemplateName:     "开庭提醒",
			Channel:          "wechat",
			RecipientType:    "client",
			TriggerEvent:     "hearing_reminder",
			SubjectTemplate:  "开庭提醒",
			ContentTemplate:  "尊敬的客户，您的案件\"{case_title}\"将于{hearing_date}在{court}开庭。",
			Variables:        JSON(map[string]interface{}{"variables": []string{"case_title", "hearing_date", "court"}}),
			AutoSend:         false,
			RequiresApproval: true,
			IsActive:         true,
		},
		{
			TemplateCode:     "CASE_PROGRESS",
			TemplateName:     "案件进展通知",
			Channel:          "wechat",
			RecipientType:    "client",
			TriggerEvent:     "case_progress",
			SubjectTemplate:  "案件进展",
			ContentTemplate:  "尊敬的客户，您的案件\"{case_title}\"有新进展：{progress}。",
			Variables:        JSON(map[string]interface{}{"variables": []string{"case_title", "progress"}}),
			AutoSend:         false,
			RequiresApproval: true,
			IsActive:         true,
		},
	}
}

// DefaultCaseFolderTemplate 返回默认的案件文件夹模板
func DefaultCaseFolderTemplate(createdBy uint) CaseFolderTemplate {
	return CaseFolderTemplate{
		Name:             "标准民商事案件模板",
		Description:      "适用于一般民商事案件",
		FolderStructure:  JSON(map[string]interface{}{
			"folders": []map[string]interface{}{
				{"name": "01_客户证据", "description": "客户提供的证据材料"},
				{
					"name":        "02_法律文书",
					"description": "起诉状、答辩状、代理词等",
					"subfolders": []map[string]interface{}{
						{"name": "起诉状", "template": "template_indictment.docx"},
						{"name": "答辩状", "template": "template_answer.docx"},
						{"name": "代理词", "template": "template_opinion.docx"},
					},
				},
				{"name": "03_法院传票与通知", "description": "法院送达的各种文书"},
				{"name": "04_研究报告与备忘录", "description": "内部分析材料"},
				{"name": "05_结案材料", "description": "判决书、裁定书、结案报告"},
			},
		}),
		CaseType:   "civil",
		IsDefault:  true,
		IsActive:   true,
		CreatedBy:  createdBy,
	}
}
