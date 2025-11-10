package models

import (
	"time"

	"gorm.io/gorm"
)

// ApprovalRequest 审批申请表
type ApprovalRequest struct {
	ID                   string            `json:"id" gorm:"column:id;type:varchar(36);primaryKey;default:(uuid_generate_v4())"`
	RequestNumber        string            `json:"request_number" gorm:"column:request_number;type:varchar(50);uniqueIndex;not null"`
	Title                string            `json:"title" gorm:"column:title;type:varchar(255);not null"`
	Type                 string            `json:"type" gorm:"column:type;type:varchar(50);not null"`
	Category             string            `json:"category" gorm:"column:category;type:varchar(100)"`
	Content              string            `json:"content" gorm:"column:content;type:text;not null"`

	// 申请人信息
	ApplicantID          string            `json:"applicant_id" gorm:"column:applicant_id;type:varchar(36);not null"`
	ApplicantName        string            `json:"applicant_name" gorm:"column:applicant_name;type:varchar(255);not null"`
	ApplicantTitle       string            `json:"applicant_title" gorm:"column:applicant_title;type:varchar(100)"`
	DepartmentID         string            `json:"department_id" gorm:"column:department_id;type:varchar(36)"`
	DepartmentName       string            `json:"department_name" gorm:"column:department_name;type:varchar(255)"`

	// 紧急程度
	Urgency              string            `json:"urgency" gorm:"column:urgency;type:varchar(20);default:'normal';not null"`
	Priority             string            `json:"priority" gorm:"column:priority;type:varchar(20);default:'medium';not null"`

	// 时间相关
	ExpectedEffectiveDate *time.Time       `json:"expected_effective_date" gorm:"column:expected_effective_date"`
	ExpectedExpiryDate   *time.Time       `json:"expected_expiry_date" gorm:"column:expected_expiry_date"`
	DurationDays         int               `json:"duration_days" gorm:"column:duration_days"`

	// 申请状态
	Status               string            `json:"status" gorm:"column:status;type:varchar(20);default:'draft';not null;index"`
	SubmissionDate       *time.Time        `json:"submission_date" gorm:"column:submission_date"`

	// 当前审批信息
	CurrentStage         string            `json:"current_stage" gorm:"column:current_stage;type:varchar(100)"`
	CurrentApproverID    string            `json:"current_approver_id" gorm:"column:current_approver_id;type:varchar(36)"`
	CurrentApproverName  string            `json:"current_approver_name" gorm:"column:current_approver_name;type:varchar(255)"`

	// 审批流程配置
	WorkflowType         string            `json:"workflow_type" gorm:"column:workflow_type;type:varchar(100);not null"`
	WorkflowConfig       string            `json:"workflow_config" gorm:"column:workflow_config;type:json"`

	// 附加信息
	Attachments          string            `json:"attachments" gorm:"column:attachments;type:json"`
	Metadata             string            `json:"metadata" gorm:"column:metadata;type:json"`

	// 审计信息
	CreatedBy            string            `json:"created_by" gorm:"column:created_by;type:varchar(36);not null"`
	UpdatedBy            string            `json:"updated_by" gorm:"column:updated_by;type:varchar(36)"`
	CreatedAt            time.Time         `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt            time.Time         `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt            gorm.DeletedAt    `json:"-" gorm:"column:deleted_at;index"`

	// 关联数据
	Records              []ApprovalRecord  `json:"records,omitempty" gorm:"foreignKey:ApprovalRequestID"`
	Notifications        []ApprovalNotification `json:"notifications,omitempty" gorm:"foreignKey:ApprovalRequestID"`
}

// ApprovalWorkflow 审批工作流定义表
type ApprovalWorkflow struct {
	ID                     string            `json:"id" gorm:"column:id;type:varchar(36);primaryKey;default:(uuid_generate_v4())"`
	WorkflowCode           string            `json:"workflow_code" gorm:"column:workflow_code;type:varchar(100);uniqueIndex;not null"`
	WorkflowName           string            `json:"workflow_name" gorm:"column:workflow_name;type:varchar(255);not null"`
	WorkflowType           string            `json:"workflow_type" gorm:"column:workflow_type;type:varchar(100);not null"`

	// 适用范围
	ApplicableTypes        string            `json:"applicable_types" gorm:"column:applicable_types;type:json"`
	ApplicableDepartments  string            `json:"applicable_departments" gorm:"column:applicable_departments;type:json"`
	ApplicableRoles        string            `json:"applicable_roles" gorm:"column:applicable_roles;type:json"`

	// 工作流配置
	Stages                 string            `json:"stages" gorm:"column:stages;type:json;not null"`
	Conditions             string            `json:"conditions" gorm:"column:conditions;type:json"`
	Timeouts               string            `json:"timeouts" gorm:"column:timeouts;type:json"`

	// 权限配置
	Permissions            string            `json:"permissions" gorm:"column:permissions;type:json"`
	Notifications          string            `json:"notifications" gorm:"column:notifications;type:json"`

	// 状态和版本
	Status                 string            `json:"status" gorm:"column:status;type:varchar(20);default:'active';not null;index"`
	Version                int               `json:"version" gorm:"column:version;default:1"`
	EffectiveDate          time.Time         `json:"effective_date" gorm:"column:effective_date;type:date;not null"`
	ExpiryDate             *time.Time        `json:"expiry_date" gorm:"column:expiry_date;type:date"`

	// 审计信息
	CreatedBy              string            `json:"created_by" gorm:"column:created_by;type:varchar(36);not null"`
	UpdatedBy              string            `json:"updated_by" gorm:"column:updated_by;type:varchar(36)"`
	CreatedAt              time.Time         `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt              time.Time         `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// ApprovalRecord 审批记录表
type ApprovalRecord struct {
	ID                  string            `json:"id" gorm:"column:id;type:varchar(36);primaryKey;default:(uuid_generate_v4())"`
	ApprovalRequestID   string            `json:"approval_request_id" gorm:"column:approval_request_id;type:varchar(36);not null;index"`

	// 审批基本信息
	Stage               string            `json:"stage" gorm:"column:stage;type:varchar(100);not null"`
	StageOrder          int               `json:"stage_order" gorm:"column:stage_order;not null"`
	ApproverID          string            `json:"approver_id" gorm:"column:approver_id;type:varchar(36);not null"`
	ApproverName        string            `json:"approver_name" gorm:"column:approver_name;type:varchar(255);not null"`
	ApproverTitle       string            `json:"approver_title" gorm:"column:approver_title;type:varchar(100)"`
	ApproverRole        string            `json:"approver_role" gorm:"column:approver_role;type:varchar(100)"`

	// 审批决定
	Decision            string            `json:"decision" gorm:"column:decision;type:varchar(20);not null"`
	DecisionReason      string            `json:"decision_reason" gorm:"column:decision_reason;type:text;not null"`
	DecisionComments    string            `json:"decision_comments" gorm:"column:decision_comments;type:text"`

	// 审批条件
	ApprovedConditions  string            `json:"approved_conditions" gorm:"column:approved_conditions;type:json"`
	ImposedRequirements string            `json:"imposed_requirements" gorm:"column:imposed_requirements;type:json"`
	FollowUpActions     string            `json:"follow_up_actions" gorm:"column:follow_up_actions;type:json"`

	// 时间信息
	ApprovalDate        time.Time         `json:"approval_date" gorm:"column:approval_date;default:CURRENT_TIMESTAMP"`
	EffectiveDate       *time.Time        `json:"effective_date" gorm:"column:effective_date"`
	NextReviewDate      *time.Time        `json:"next_review_date" gorm:"column:next_review_date"`

	// 附件和证据
	SupportingDocuments string            `json:"supporting_documents" gorm:"column:supporting_documents;type:json"`
	EvidenceReferences  string            `json:"evidence_references" gorm:"column:evidence_references;type:json"`

	// 状态
	Status              string            `json:"status" gorm:"column:status;type:varchar(20);default:'active'"`

	// 审计信息
	CreatedAt           time.Time         `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt           time.Time         `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`

	// 关联数据
	ApprovalRequest     ApprovalRequest   `json:"approval_request,omitempty" gorm:"foreignKey:ApprovalRequestID"`
}

// ApprovalTemplate 审批模板表
type ApprovalTemplate struct {
	ID                  string            `json:"id" gorm:"column:id;type:varchar(36);primaryKey;default:(uuid_generate_v4())"`
	TemplateCode        string            `json:"template_code" gorm:"column:template_code;type:varchar(100);uniqueIndex;not null"`
	TemplateName        string            `json:"template_name" gorm:"column:template_name;type:varchar(255);not null"`
	TemplateType        string            `json:"template_type" gorm:"column:template_type;type:varchar(100);not null"`
	Category            string            `json:"category" gorm:"column:category;type:varchar(100)"`
	WorkflowType        string            `json:"workflow_type" gorm:"column:workflow_type;type:varchar(100);not null"`

	// 模板内容
	TemplateContent     string            `json:"template_content" gorm:"column:template_content;type:json;not null"`
	FormSchema          string            `json:"form_schema" gorm:"column:form_schema;type:json"`
	ValidationRules     string            `json:"validation_rules" gorm:"column:validation_rules;type:json"`

	// 默认配置
	DefaultValues       string            `json:"default_values" gorm:"column:default_values;type:json"`
	RequiredFields      string            `json:"required_fields" gorm:"column:required_fields;type:json"`
	OptionalFields      string            `json:"optional_fields" gorm:"column:optional_fields;type:json"`

	// 适用条件
	ApplicableScenarios string            `json:"applicable_scenarios" gorm:"column:applicable_scenarios;type:json"`
	ApplicableRoles     string            `json:"applicable_roles" gorm:"column:applicable_roles;type:json"`
	ApplicableDepartments string          `json:"applicable_departments" gorm:"column:applicable_departments;type:json"`

	// 状态和版本
	Status              string            `json:"status" gorm:"column:status;type:varchar(20);default:'active';not null;index"`
	Version             int               `json:"version" gorm:"column:version;default:1"`
	EffectiveDate       time.Time         `json:"effective_date" gorm:"column:effective_date;type:date;not null"`
	ExpiryDate          *time.Time        `json:"expiry_date" gorm:"column:expiry_date;type:date"`

	// 使用统计
	UsageCount          int               `json:"usage_count" gorm:"column:usage_count;default:0"`
	LastUsedDate        *time.Time        `json:"last_used_date" gorm:"column:last_used_date"`

	// 审计信息
	CreatedBy           string            `json:"created_by" gorm:"column:created_by;type:varchar(36);not null"`
	UpdatedBy           string            `json:"updated_by" gorm:"column:updated_by;type:varchar(36)"`
	ApprovedBy          string            `json:"approved_by" gorm:"column:approved_by;type:varchar(36)"`
	CreatedAt           time.Time         `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt           time.Time         `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// ApprovalNotification 审批通知记录表
type ApprovalNotification struct {
	ID                  string            `json:"id" gorm:"column:id;type:varchar(36);primaryKey;default:(uuid_generate_v4())"`
	ApprovalRequestID   string            `json:"approval_request_id" gorm:"column:approval_request_id;type:varchar(36);not null;index"`
	ApprovalRecordID    string            `json:"approval_record_id" gorm:"column:approval_record_id;type:varchar(36)"`

	// 通知信息
	NotificationType    string            `json:"notification_type" gorm:"column:notification_type;type:varchar(20);not null"`
	RecipientID         string            `json:"recipient_id" gorm:"column:recipient_id;type:varchar(36);not null"`
	RecipientName       string            `json:"recipient_name" gorm:"column:recipient_name;type:varchar(255);not null"`
	RecipientEmail      string            `json:"recipient_email" gorm:"column:recipient_email;type:varchar(255)"`

	// 通知内容
	Subject             string            `json:"subject" gorm:"column:subject;type:varchar(500);not null"`
	Content             string            `json:"content" gorm:"column:content;type:text;not null"`

	// 发送信息
	SendMethod          string            `json:"send_method" gorm:"column:send_method;type:varchar(20);not null"`
	SendStatus          string            `json:"send_status" gorm:"column:send_status;type:varchar(20);default:'pending'"`
	SendAttempts        int               `json:"send_attempts" gorm:"column:send_attempts;default:0"`

	// 时间信息
	ScheduledAt         *time.Time        `json:"scheduled_at" gorm:"column:scheduled_at"`
	SentAt              *time.Time        `json:"sent_at" gorm:"column:sent_at"`

	// 响应信息
	ReadAt              *time.Time        `json:"read_at" gorm:"column:read_at"`
	ResponseAction      string            `json:"response_action" gorm:"column:response_action;type:varchar(100)"`

	// 错误信息
	ErrorMessage        string            `json:"error_message" gorm:"column:error_message;type:text"`

	// 审计信息
	CreatedAt           time.Time         `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt           time.Time         `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`

	// 关联数据
	ApprovalRequest     ApprovalRequest   `json:"approval_request,omitempty" gorm:"foreignKey:ApprovalRequestID"`
}

// ApprovalStats 审批统计信息
type ApprovalStats struct {
	TotalRequests     int `json:"total_requests"`
	PendingRequests   int `json:"pending_requests"`
	MyPendingRequests int `json:"my_pending_requests"`
	ApprovedRequests  int `json:"approved_requests"`
	RejectedRequests  int `json:"rejected_requests"`
}

// ApprovalListRequest 审批列表查询请求
type ApprovalListRequest struct {
	Page        int               `json:"page" form:"page" binding:"min=1"`
	PageSize    int               `json:"page_size" form:"page_size" binding:"min=1,max=100"`
	Status      string            `json:"status" form:"status"`
	Type        string            `json:"type" form:"type"`
	ApplicantID string            `json:"applicant_id" form:"applicant_id"`
	Keyword     string            `json:"keyword" form:"keyword"`
	StartDate   string            `json:"start_date" form:"start_date"`
	EndDate     string            `json:"end_date" form:"end_date"`
	SortBy      string            `json:"sort_by" form:"sort_by"`
	SortOrder   string            `json:"sort_order" form:"sort_order"`
}

// ApprovalListResponse 审批列表响应
type ApprovalListResponse struct {
	List       []ApprovalRequest `json:"list"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// CreateApprovalRequest 创建审批申请请求
type CreateApprovalRequest struct {
	Title             string                 `json:"title" binding:"required"`
	Type              string                 `json:"type" binding:"required"`
	Category          string                 `json:"category"`
	Content           string                 `json:"content" binding:"required"`
	WorkflowType      string                 `json:"workflow_type"`
	Urgency           string                 `json:"urgency"`
	Priority          string                 `json:"priority"`
	ExpectedEffectiveDate string              `json:"expected_effective_date"`
	ExpectedExpiryDate   string              `json:"expected_expiry_date"`
	DurationDays      int                    `json:"duration_days"`
	Metadata          map[string]interface{} `json:"metadata"`
	Attachments       []map[string]interface{} `json:"attachments"`
}

// UpdateApprovalRequest 更新审批申请请求
type UpdateApprovalRequest struct {
	Title             string                 `json:"title"`
	Content           string                 `json:"content"`
	Metadata          map[string]interface{} `json:"metadata"`
	Attachments       []map[string]interface{} `json:"attachments"`
}

// ApprovalDecisionRequest 审批决定请求
type ApprovalDecisionRequest struct {
	Decision           string                 `json:"decision" binding:"required,oneof=approve reject request_changes defer escalate reassign"`
	DecisionReason     string                 `json:"decision_reason" binding:"required"`
	DecisionComments   string                 `json:"decision_comments"`
	ApprovedConditions map[string]interface{} `json:"approved_conditions"`
	ImposedRequirements map[string]interface{} `json:"imposed_requirements"`
	FollowUpActions    []map[string]interface{} `json:"follow_up_actions"`
	SupportingDocuments []map[string]interface{} `json:"supporting_documents"`
	EvidenceReferences  []map[string]interface{} `json:"evidence_references"`
	NextApproverID     string                 `json:"next_approver_id"` // 用于转派
}

// TableName 设置表名
func (ApprovalRequest) TableName() string {
	return "approval_requests"
}

func (ApprovalWorkflow) TableName() string {
	return "approval_workflows"
}

func (ApprovalRecord) TableName() string {
	return "approval_records"
}

func (ApprovalTemplate) TableName() string {
	return "approval_templates"
}

func (ApprovalNotification) TableName() string {
	return "approval_notifications"
}

// 审批状态常量
const (
	ApprovalStatusDraft        = "draft"
	ApprovalStatusSubmitted    = "submitted"
	ApprovalStatusUnderReview  = "under_review"
	ApprovalStatusApproved     = "approved"
	ApprovalStatusRejected     = "rejected"
	ApprovalStatusCancelled    = "cancelled"
	ApprovalStatusExpired      = "expired"
)

// 审批决定常量
const (
	ApprovalDecisionApprove        = "approve"
	ApprovalDecisionReject         = "reject"
	ApprovalDecisionRequestChanges = "request_changes"
	ApprovalDecisionDefer          = "defer"
	ApprovalDecisionEscalate       = "escalate"
	ApprovalDecisionReassign       = "reassign"
)

// 紧急程度常量
const (
	ApprovalUrgencyNormal    = "normal"
	ApprovalUrgencyUrgent    = "urgent"
	ApprovalUrgencyVeryUrgent = "very_urgent"
)

// 优先级常量
const (
	ApprovalPriorityLow      = "low"
	ApprovalPriorityMedium   = "medium"
	ApprovalPriorityHigh     = "high"
	ApprovalPriorityCritical = "critical"
)

// 工作流状态常量
const (
	WorkflowStatusActive     = "active"
	WorkflowStatusInactive   = "inactive"
	WorkflowStatusDeprecated = "deprecated"
)

// 模板状态常量
const (
	TemplateStatusActive     = "active"
	TemplateStatusInactive   = "inactive"
	TemplateStatusUnderReview = "under_review"
)

// 通知类型常量
const (
	NotificationTypeSubmission   = "submission"
	NotificationTypeApproval     = "approval"
	NotificationTypeRejection    = "rejection"
	NotificationTypeReminder     = "reminder"
	NotificationTypeEscalation   = "escalation"
	NotificationTypeCompletion   = "completion"
)

// 发送方式常量
const (
	SendMethodEmail  = "email"
	SendMethodSMS    = "sms"
	SendMethodSystem = "system"
	SendMethodWechat = "wechat"
)

// 发送状态常量
const (
	SendStatusPending  = "pending"
	SendStatusSent     = "sent"
	SendStatusFailed   = "failed"
	SendStatusCancelled = "cancelled"
)