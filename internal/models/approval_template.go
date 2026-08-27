package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// 预置模板常量
const (
	TemplateSealApproval     = "seal_approval"     // 用印审批
	TemplateCaseFiling       = "case_filing"       // 立案审批
	TemplateLeaveApproval    = "leave_approval"    // 请假审批
	TemplatePurchaseApproval = "purchase_approval" // 采购审批
)

// ApprovalTemplateConfig 审批模板完整配置（用于JSON存储）
type ApprovalTemplateConfig struct {
	// 审批节点列表
	Stages []ApprovalStage `json:"stages"`
	// 条件规则
	Conditions []ApprovalConditionRule `json:"conditions,omitempty"`
	// 超时配置（小时）
	Timeouts map[string]int `json:"timeouts,omitempty"`
	// 通知配置
	Notifications NotificationConfig `json:"notifications,omitempty"`
	// 表单配置
	FormSchema *FormSchema `json:"form_schema,omitempty"`
}

// Scan 实现 sql.Scanner 接口
func (c *ApprovalTemplateConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal ApprovalTemplateConfig value")
	}
	return json.Unmarshal(bytes, c)
}

// Value 实现 driver.Valuer 接口
func (c ApprovalTemplateConfig) Value() (driver.Value, error) {
	if len(c.Stages) == 0 {
		return "{}", nil
	}
	return json.Marshal(c)
}

// ApprovalStage 审批节点
type ApprovalStage struct {
	// 节点唯一标识
	StageKey string `json:"stage_key"`
	// 节点名称
	StageName string `json:"stage_name"`
	// 节点顺序
	Order int `json:"order"`
	// 审批类型: serial(串行), parallel(并行), conditional(条件)
	StageType string `json:"stage_type"`
	// 审批人配置
	Approvers []ApproverConfig `json:"approvers"`
	// 审批模式: and(会签-全部通过), or(或签-任一通过)
	ApprovalMode string `json:"approval_mode"`
	// 是否允许转签
	AllowReassign bool `json:"allow_reassign"`
	// 是否允许加签
	AllowAddApprovers bool `json:"allow_add_approvers"`
	// 是否允许退回
	AllowReturn bool `json:"allow_return"`
	// 自动通过规则（如金额低于阈值）
	AutoPassCondition string `json:"auto_pass_condition,omitempty"`
}

// ApproverConfig 审批人配置
type ApproverConfig struct {
	// 审批人类型: user(指定用户), role(角色), department_head(部门负责人), initiator_leader(申请人上级)
	ApproverType string `json:"approver_type"`
	// 审批人ID（当类型为user时）
	UserID string `json:"user_id,omitempty"`
	// 角色代码（当类型为role时）
	RoleCode string `json:"role_code,omitempty"`
	// 部门ID（当类型为department_head时）
	DepartmentID string `json:"department_id,omitempty"`
	// 审批人姓名（用于显示）
	UserName string `json:"user_name,omitempty"`
	// 是否必须审批
	Required bool `json:"required"`
}

// ApprovalConditionRule 审批条件规则
type ApprovalConditionRule struct {
	// 条件唯一标识
	ConditionKey string `json:"condition_key"`
	// 条件名称
	ConditionName string `json:"condition_name"`
	// 条件表达式（支持简单比较）
	Expression string `json:"expression"`
	// 满足条件时的流程分支
	ThenStageKey string `json:"then_stage_key"`
	// 不满足条件时的流程分支
	ElseStageKey string `json:"else_stage_key,omitempty"`
}

// NotificationConfig 通知配置
type NotificationConfig struct {
	// 提交时通知
	OnSubmit bool `json:"on_submit"`
	// 审批通过通知
	OnApprove bool `json:"on_approve"`
	// 审批拒绝通知
	OnReject bool `json:"on_reject"`
	// 超时提醒
	OnTimeout bool `json:"on_timeout"`
	// 超时时间（小时）
	TimeoutHours int `json:"timeout_hours"`
}

// FormSchema 表单配置
type FormSchema struct {
	// 表单字段
	Fields []FormField `json:"fields"`
	// 验证规则
	ValidationRules map[string]string `json:"validation_rules,omitempty"`
}

// FormField 表单字段
type FormField struct {
	// 字段名
	Name string `json:"name"`
	// 显示名称
	Label string `json:"label"`
	// 字段类型: text, number, date, select, textarea, file
	Type string `json:"type"`
	// 是否必填
	Required bool `json:"required"`
	// 选项（用于select类型）
	Options []string `json:"options,omitempty"`
	// 默认值
	Default interface{} `json:"default,omitempty"`
	// 占位符
	Placeholder string `json:"placeholder,omitempty"`
}

// SealApprovalMetadata 用印审批元数据
type SealApprovalMetadata struct {
	SealType       string   `json:"seal_type"`       // 印章类型：公章、合同章、法人章等
	SealCount      int      `json:"seal_count"`      // 用印次数
	SealImportance string   `json:"seal_importance"` // 重要程度：low, medium, high
	DocumentTitle  string   `json:"document_title"`  // 文件标题
	DocumentType   string   `json:"document_type"`   // 文件类型
	ContractValue  float64  `json:"contract_value"`  // 合同金额（如有）
	DocumentFiles  []string `json:"document_files"`  // 附件文件列表
	UsePurpose     string   `json:"use_purpose"`     // 用印目的
	ExpectedDate   string   `json:"expected_date"`   // 预计用印日期
}

// CaseCreationMetadata 立案审批元数据
type CaseCreationMetadata struct {
	ClientID              uint    `json:"client_id"`               // 客户ID
	ClientName            string  `json:"client_name"`             // 客户名称
	CaseTitle             string  `json:"case_title"`              // 案件标题
	CaseType              string  `json:"case_type"`               // 案件类型
	CaseValue             float64 `json:"case_value"`              // 标的额
	OpposingParty         string  `json:"opposing_party"`          // 对方当事人
	LawyerID              uint    `json:"lawyer_id"`               // 承办律师ID
	LawyerName            string  `json:"lawyer_name"`             // 承办律师姓名
	Urgency               string  `json:"urgency"`                 // 紧急程度
	ExpectedPeriod        string  `json:"expected_period"`         // 预计期限
	RiskLevel             string  `json:"risk_level"`              // 风险等级
	ConflictCheckRequired bool    `json:"conflict_check_required"` // 是否需要冲突核查
	CaseDescription       string  `json:"case_description"`        // 案件描述
	LegalBasis            string  `json:"legal_basis"`             // 法律依据
}

// GetSealApprovalTemplate 获取用印审批模板配置
func GetSealApprovalTemplate() ApprovalTemplateConfig {
	return ApprovalTemplateConfig{
		Stages: []ApprovalStage{
			{
				StageKey:     "department_review",
				StageName:    "部门审核",
				Order:        1,
				StageType:    "serial",
				ApprovalMode: "or",
				Approvers: []ApproverConfig{
					{
						ApproverType: "role",
						RoleCode:     "department_head",
						UserName:     "部门负责人",
						Required:     true,
					},
				},
				AllowReassign:     true,
				AllowAddApprovers: false,
				AllowReturn:       false,
			},
			{
				StageKey:     "partner_review",
				StageName:    "合伙人审核",
				Order:        2,
				StageType:    "conditional",
				ApprovalMode: "or",
				Approvers: []ApproverConfig{
					{
						ApproverType: "role",
						RoleCode:     "partner",
						UserName:     "合伙人",
						Required:     true,
					},
				},
				AllowReassign:     true,
				AllowAddApprovers: true,
				AllowReturn:       true,
			},
			{
				StageKey:     "director_review",
				StageName:    "主任审核",
				Order:        3,
				StageType:    "conditional",
				ApprovalMode: "or",
				Approvers: []ApproverConfig{
					{
						ApproverType: "role",
						RoleCode:     "director",
						UserName:     "主任",
						Required:     true,
					},
				},
				AllowReassign:     false,
				AllowAddApprovers: false,
				AllowReturn:       true,
			},
		},
		Conditions: []ApprovalConditionRule{
			{
				ConditionKey:  "high_value_seal",
				ConditionName: "重要用印需主任审批",
				Expression:    "seal_importance == 'high' || seal_count >= 3",
				ThenStageKey:  "director_review",
				ElseStageKey:  "end",
			},
			{
				ConditionKey:  "medium_value_seal",
				ConditionName: "一般用印需合伙人审批",
				Expression:    "seal_importance == 'medium' || seal_count >= 1",
				ThenStageKey:  "partner_review",
				ElseStageKey:  "end",
			},
		},
		Timeouts: map[string]int{
			"department_review": 24,
			"partner_review":    48,
			"director_review":   72,
		},
		Notifications: NotificationConfig{
			OnSubmit:     true,
			OnApprove:    true,
			OnReject:     true,
			OnTimeout:    true,
			TimeoutHours: 24,
		},
		FormSchema: &FormSchema{
			Fields: []FormField{
				{Name: "document_title", Label: "文件标题", Type: "text", Required: true},
				{Name: "document_type", Label: "文件类型", Type: "select", Required: true, Options: []string{"合同", "函件", "证明", "其他"}},
				{Name: "seal_type", Label: "印章类型", Type: "select", Required: true, Options: []string{"公章", "合同章", "法人章", "财务章"}},
				{Name: "seal_count", Label: "用印次数", Type: "number", Required: true, Default: 1},
				{Name: "seal_importance", Label: "重要程度", Type: "select", Required: true, Options: []string{"low", "medium", "high"}},
				{Name: "contract_value", Label: "合同金额", Type: "number", Required: false},
				{Name: "use_purpose", Label: "用印目的", Type: "textarea", Required: true},
				{Name: "document_files", Label: "附件文件", Type: "file", Required: false},
			},
		},
	}
}

// GetCaseCreationApprovalTemplate 获取立案审批模板配置
func GetCaseCreationApprovalTemplate() ApprovalTemplateConfig {
	return ApprovalTemplateConfig{
		Stages: []ApprovalStage{
			{
				StageKey:     "conflict_check",
				StageName:    "利益冲突核查",
				Order:        1,
				StageType:    "serial",
				ApprovalMode: "or",
				Approvers: []ApproverConfig{
					{
						ApproverType: "role",
						RoleCode:     "compliance",
						UserName:     "合规专员",
						Required:     true,
					},
				},
				AllowReassign:     false,
				AllowAddApprovers: false,
				AllowReturn:       false,
			},
			{
				StageKey:     "department_review",
				StageName:    "部门审核",
				Order:        2,
				StageType:    "serial",
				ApprovalMode: "or",
				Approvers: []ApproverConfig{
					{
						ApproverType: "role",
						RoleCode:     "department_head",
						UserName:     "部门负责人",
						Required:     true,
					},
				},
				AllowReassign:     true,
				AllowAddApprovers: false,
				AllowReturn:       false,
			},
			{
				StageKey:     "partner_review",
				StageName:    "合伙人审批",
				Order:        3,
				StageType:    "conditional",
				ApprovalMode: "or",
				Approvers: []ApproverConfig{
					{
						ApproverType: "role",
						RoleCode:     "partner",
						UserName:     "合伙人",
						Required:     true,
					},
				},
				AllowReassign:     true,
				AllowAddApprovers: true,
				AllowReturn:       true,
			},
			{
				StageKey:     "director_review",
				StageName:    "主任审批",
				Order:        4,
				StageType:    "conditional",
				ApprovalMode: "or",
				Approvers: []ApproverConfig{
					{
						ApproverType: "role",
						RoleCode:     "director",
						UserName:     "主任",
						Required:     true,
					},
				},
				AllowReassign:     false,
				AllowAddApprovers: false,
				AllowReturn:       true,
			},
		},
		Conditions: []ApprovalConditionRule{
			{
				ConditionKey:  "high_value_case",
				ConditionName: "高标的额案件需主任审批",
				Expression:    "case_value >= 1000000",
				ThenStageKey:  "director_review",
				ElseStageKey:  "end",
			},
			{
				ConditionKey:  "medium_value_case",
				ConditionName: "中等标的额案件需合伙人审批",
				Expression:    "case_value >= 100000",
				ThenStageKey:  "partner_review",
				ElseStageKey:  "end",
			},
		},
		Timeouts: map[string]int{
			"conflict_check":    8,
			"department_review": 24,
			"partner_review":    48,
			"director_review":   72,
		},
		Notifications: NotificationConfig{
			OnSubmit:     true,
			OnApprove:    true,
			OnReject:     true,
			OnTimeout:    true,
			TimeoutHours: 24,
		},
		FormSchema: &FormSchema{
			Fields: []FormField{
				{Name: "client_name", Label: "客户名称", Type: "text", Required: true},
				{Name: "case_title", Label: "案件标题", Type: "text", Required: true},
				{Name: "case_type", Label: "案件类型", Type: "select", Required: true, Options: []string{"民事诉讼", "刑事诉讼", "行政诉讼", "仲裁", "非诉讼"}},
				{Name: "case_value", Label: "标的额", Type: "number", Required: true},
				{Name: "opposing_party", Label: "对方当事人", Type: "text", Required: true},
				{Name: "lawyer_name", Label: "承办律师", Type: "text", Required: true},
				{Name: "urgency", Label: "紧急程度", Type: "select", Required: true, Options: []string{"一般", "紧急", "非常紧急"}},
				{Name: "expected_period", Label: "预计期限", Type: "text", Required: false},
				{Name: "risk_level", Label: "风险等级", Type: "select", Required: true, Options: []string{"低", "中", "高"}},
				{Name: "case_description", Label: "案件描述", Type: "textarea", Required: true},
				{Name: "legal_basis", Label: "法律依据", Type: "textarea", Required: false},
			},
		},
	}
}

// GetTemplateConfigByType 根据类型获取模板配置
func GetTemplateConfigByType(templateType string) (ApprovalTemplateConfig, error) {
	switch templateType {
	case TemplateSealApproval:
		return GetSealApprovalTemplate(), nil
	case TemplateCaseFiling:
		return GetCaseCreationApprovalTemplate(), nil
	default:
		return ApprovalTemplateConfig{}, errors.New("unknown template type: " + templateType)
	}
}

// GetPresetTemplates 获取所有预置模板
func GetPresetTemplates() map[string]ApprovalTemplateConfig {
	return map[string]ApprovalTemplateConfig{
		TemplateSealApproval: GetSealApprovalTemplate(),
		TemplateCaseFiling:   GetCaseCreationApprovalTemplate(),
	}
}

// ApprovalTemplate 审批模板定义
// 用于定义不同类型审批的流程模板，如用印审批、立案审批等
type ApprovalTemplate struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	Name        string         `json:"name" gorm:"size:50;not null;uniqueIndex"` // 模板标识名: seal_approval, case_filing
	DisplayName string         `json:"display_name" gorm:"size:100;not null"`    // 显示名称: 用印审批, 立案审批
	Description string         `json:"description" gorm:"type:text"`             // 模板描述
	Steps       string         `json:"steps" gorm:"type:jsonb;not null"`         // 审批步骤配置(JSON)
	Conditions  string         `json:"conditions" gorm:"type:jsonb"`             // 条件分支配置(JSON)
	IsActive    bool           `json:"is_active" gorm:"default:true;not null"`   // 是否启用
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// ApprovalStep 审批步骤配置
type ApprovalStep struct {
	Order        int    `json:"order"`         // 步骤顺序
	Name         string `json:"name"`          // 步骤名称
	ApproverType string `json:"approver_type"` // ROLE, SPECIFIC_USER, DEPARTMENT_HEAD
	ApproverRole string `json:"approver_role"` // 角色名(当type=ROLE时)
	ApproverID   uint   `json:"approver_id"`   // 具体审批人ID(当type=SPECIFIC_USER时)
	IsRequired   bool   `json:"is_required"`   // 是否必须
	SignType     string `json:"sign_type"`     // 签章类型: SINGLE, COUNTERSIGN(会签), OR_SIGN(或签)
	AutoPass     bool   `json:"auto_pass"`     // 是否自动通过(用于条件分支)
}

// ApprovalCondition 审批条件配置
type ApprovalCondition struct {
	Field    string      `json:"field"`     // 判断字段
	Operator string      `json:"operator"`  // 操作符: gt, lt, eq, contains
	Value    interface{} `json:"value"`     // 比较值
	ThenStep int         `json:"then_step"` // 满足条件跳转到的步骤
	ElseStep int         `json:"else_step"` // 不满足条件跳转到的步骤
}

// TableName 设置表名
func (ApprovalTemplate) TableName() string {
	return "approval_templates"
}

// 审批人类型常量
const (
	ApproverTypeRole           = "ROLE"            // 角色审批
	ApproverTypeSpecificUser   = "SPECIFIC_USER"   // 指定用户
	ApproverTypeDepartmentHead = "DEPARTMENT_HEAD" // 部门主任
	ApproverTypeSuperior       = "SUPERIOR"        // 上级
)

// 签章类型常量
const (
	SignTypeSingle      = "SINGLE"      // 单签
	SignTypeCountersign = "COUNTERSIGN" // 会签(所有审批人都需要审批)
	SignTypeOrSign      = "OR_SIGN"     // 或签(任意一人审批即可)
)
