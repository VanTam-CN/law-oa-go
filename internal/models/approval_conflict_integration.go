package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// StringSlice 用于GORM处理字符串切片的自定义类型
type StringSlice []string

// Value 实现driver.Valuer接口
func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan 实现sql.Scanner接口
func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return nil
	}
}

// StringMap 用于GORM处理字符串映射的自定义类型
type StringMap map[string]interface{}

// Value 实现driver.Valuer接口
func (m StringMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan 实现sql.Scanner接口
func (m *StringMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, m)
	case string:
		return json.Unmarshal([]byte(v), m)
	default:
		return nil
	}
}

// ApprovalConflictIntegration 审批与冲突检测集成模型
// 扩展现有的审批申请模型，添加与冲突检测和案件创建的关联

// ConflictCheckAssociation 冲突检测关联信息
type ConflictCheckAssociation struct {
	gorm.Model
	// 冲突检测基本信息
	CheckID     string    `json:"check_id" gorm:"type:varchar(255);index"` // 冲突检测ID
	CheckTime   time.Time `json:"check_time"`                              // 检测时间
	RiskLevel   string    `json:"risk_level" gorm:"type:varchar(50)"`      // 风险等级
	RiskScore   float64   `json:"risk_score"`                              // 风险评分
	HasConflict bool      `json:"has_conflict"`                            // 是否存在冲突

	// 检测结果摘要
	ConflictCount int         `json:"conflict_count"`                  // 冲突案例数量
	ConflictTypes StringSlice `json:"conflict_types" gorm:"type:json"` // 冲突类型列表

	// 客户和案件信息
	ClientID   string `json:"client_id" gorm:"type:varchar(255);index"` // 客户ID
	ClientName string `json:"client_name" gorm:"type:varchar(255)"`     // 客户名称
	CaseName   string `json:"case_name" gorm:"type:varchar(255)"`       // 案件名称
	CaseType   string `json:"case_type" gorm:"type:varchar(100)"`       // 案件类型

	// 检测参数和范围
	SearchParameters StringMap `json:"search_parameters" gorm:"type:json"`    // 搜索参数
	SearchScope      string    `json:"search_scope" gorm:"type:varchar(100)"` // 搜索范围

	// 建议和处理
	Recommendations StringSlice `json:"recommendations" gorm:"type:json"` // 建议
	Mitigation      StringSlice `json:"mitigation" gorm:"type:json"`      // 缓解措施

	// 审批指导信息
	RequiresApproval bool        `json:"requires_approval"`                       // 是否需要审批
	ApprovalLevel    string      `json:"approval_level" gorm:"type:varchar(100)"` // 审批级别
	RiskFactors      StringSlice `json:"risk_factors" gorm:"type:json"`           // 风险因素
}

// CaseCreationAssociation 案件创建关联信息
type CaseCreationAssociation struct {
	gorm.Model
	// 案件创建状态
	Created      bool      `json:"created"`                                // 是否已创建案件
	CaseID       string    `json:"case_id" gorm:"type:varchar(255);index"` // 案件ID
	CaseNumber   string    `json:"case_number" gorm:"type:varchar(255)"`   // 案件编号
	CreationTime time.Time `json:"creation_time"`                          // 案件创建时间

	// 数据映射信息
	DataMapping      StringMap   `json:"data_mapping" gorm:"type:json"`      // 数据映射详情
	MappedFields     StringSlice `json:"mapped_fields" gorm:"type:json"`     // 已映射字段列表
	ValidationErrors StringSlice `json:"validation_errors" gorm:"type:json"` // 验证错误列表

	// 条件和限制
	AppliedConditions   StringSlice `json:"applied_conditions" gorm:"type:json"`   // 应用的条件
	ImposedRequirements StringSlice `json:"imposed_requirements" gorm:"type:json"` // 施加的要求

	// 状态和跟踪
	Status        string `json:"status" gorm:"type:varchar(100)"` // 创建状态
	StatusMessage string `json:"status_message" gorm:"type:text"` // 状态消息
	RetryAttempts int    `json:"retry_attempts"`                  // 重试次数
	LastError     string `json:"last_error" gorm:"type:text"`     // 最后错误信息

	// 工作流信息
	WorkflowStep string `json:"workflow_step" gorm:"type:varchar(100)"` // 当前工作流步骤
	NextAction   string `json:"next_action" gorm:"type:varchar(100)"`   // 下一步操作
}

// ApprovalIntegrationMetadata 审批集成元数据
// 扩展审批申请的metadata字段以支持集成功能
type ApprovalIntegrationMetadata struct {
	gorm.Model
	// 关联信息标识
	IntegrationType string    `json:"integration_type" gorm:"type:varchar(50);index"`      // 集成类型: conflict, case, both
	IntegrationID   string    `json:"integration_id" gorm:"type:varchar(255);uniqueIndex"` // 集成ID
	IntegrationTime time.Time `json:"integration_time"`                                    // 集成时间

	// 冲突检测关联 - 外键关联
	ConflictCheckID uint                     `json:"conflict_check_id,omitempty" gorm:"index"` // 冲突检测关联ID
	ConflictCheck   ConflictCheckAssociation `json:"conflict_check,omitempty" gorm:"foreignKey:ConflictCheckID"`

	// 案件创建关联 - 外键关联
	CaseCreationID uint                    `json:"case_creation_id,omitempty" gorm:"index"` // 案件创建关联ID
	CaseCreation   CaseCreationAssociation `json:"case_creation,omitempty" gorm:"foreignKey:CaseCreationID"`

	// 流程控制
	AutoSubmitted    bool      `json:"auto_submitted" gorm:"default:false"`                     // 是否自动提交
	TriggerSource    string    `json:"trigger_source" gorm:"type:varchar(50);default:'manual'"` // 触发源: manual, auto
	WorkflowOverride StringMap `json:"workflow_override" gorm:"type:json"`                      // 工作流覆盖配置

	// 审计信息
	CreatedBy string `json:"created_by" gorm:"type:varchar(255)"` // 创建人
	UpdatedBy string `json:"updated_by" gorm:"type:varchar(255)"` // 更新人
	Version   int    `json:"version" gorm:"default:1"`            // 元数据版本
}

// GetConflictIntegrationType 获取冲突集成类型
func GetConflictIntegrationType() string {
	return "conflict_approval_integration"
}

// GetCaseIntegrationType 获取案件集成类型
func GetCaseIntegrationType() string {
	return "approval_case_integration"
}

// ValidateConflictAssociation 验证冲突关联信息
func (c *ConflictCheckAssociation) Validate() error {
	if c.CheckID == "" {
		return &ConflictError{Code: "INTEGRATION_001", Message: "冲突检测ID不能为空"}
	}
	if c.ClientID == "" {
		return &ConflictError{Code: "INTEGRATION_002", Message: "客户ID不能为空"}
	}
	if c.CaseName == "" {
		return &ConflictError{Code: "INTEGRATION_003", Message: "案件名称不能为空"}
	}
	return nil
}

// ValidateCaseAssociation 验证案件关联信息
func (c *CaseCreationAssociation) Validate() error {
	if c.Created && c.CaseID == "" {
		return &ConflictError{Code: "INTEGRATION_004", Message: "案件已创建但案件ID为空"}
	}
	return nil
}

// ValidateIntegrationMetadata 验证集成元数据
func (m *ApprovalIntegrationMetadata) Validate() error {
	if m.IntegrationID == "" {
		return &ConflictError{Code: "INTEGRATION_005", Message: "集成ID不能为空"}
	}
	if m.IntegrationType == "" {
		return &ConflictError{Code: "INTEGRATION_006", Message: "集成类型不能为空"}
	}

	// 验证集成类型
	validTypes := map[string]bool{"conflict": true, "case": true, "both": true}
	if !validTypes[m.IntegrationType] {
		return &ConflictError{Code: "INTEGRATION_007", Message: "无效的集成类型"}
	}

	// 如果包含冲突关联，验证冲突信息
	if m.IntegrationType == "conflict" || m.IntegrationType == "both" {
		if m.ConflictCheckID == 0 {
			return &ConflictError{Code: "INTEGRATION_008", Message: "缺少冲突关联信息"}
		}
		if err := m.ConflictCheck.Validate(); err != nil {
			return err
		}
	}

	// 如果包含案件关联，验证案件信息
	if m.IntegrationType == "case" || m.IntegrationType == "both" {
		if m.CaseCreationID == 0 {
			return &ConflictError{Code: "INTEGRATION_009", Message: "缺少案件关联信息"}
		}
		if err := m.CaseCreation.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// IsIntegrationComplete 检查集成是否完成
func (m *ApprovalIntegrationMetadata) IsIntegrationComplete() bool {
	switch m.IntegrationType {
	case "conflict":
		return m.ConflictCheckID != 0 && m.ConflictCheck.CheckID != ""
	case "case":
		return m.CaseCreationID != 0 && m.CaseCreation.Created
	case "both":
		return (m.ConflictCheckID != 0 && m.ConflictCheck.CheckID != "") &&
			(m.CaseCreationID != 0 && m.CaseCreation.Created)
	default:
		return false
	}
}

// GetIntegrationStatus 获取集成状态
func (m *ApprovalIntegrationMetadata) GetIntegrationStatus() string {
	if m.IsIntegrationComplete() {
		return "completed"
	}
	if m.ConflictCheckID != 0 && m.ConflictCheck.CheckID != "" {
		if m.CaseCreationID == 0 {
			return "pending_case_creation"
		}
		return "partial"
	}
	return "pending"
}
