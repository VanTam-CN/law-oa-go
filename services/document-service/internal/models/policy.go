package models

import (
	"time"

	"gorm.io/gorm"
)

// Policy 策略模型
type Policy struct {
	ID          uint                   `gorm:"primaryKey" json:"id"`
	Name        string                 `gorm:"size:100;not null;index" json:"name"`
	Description string                 `gorm:"size:500" json:"description"`
	Version     int                    `gorm:"not null;default:1" json:"version"`
	Enabled     bool                   `gorm:"not null;default:true" json:"enabled"`
	Priority    int                    `gorm:"not null;default:100" json:"priority"`
	Effect      string                 `gorm:"size:10;not null;default:'allow'" json:"effect"`
	Subject     string                 `gorm:"type:jsonb" json:"subject"`
	Resource    string                 `gorm:"type:jsonb" json:"resource"`
	Action      string                 `gorm:"type:jsonb" json:"action"`
	Environment string                 `gorm:"type:jsonb" json:"environment"`
	Conditions  string                 `gorm:"type:jsonb" json:"conditions"`
	Tags        []string               `gorm:"type:text[]" json:"tags"`
	TenantID    string                 `gorm:"size:50;not null;index" json:"tenant_id"`
	CreatedBy   uint                   `gorm:"not null;index" json:"created_by"`
	CreatedAt   time.Time              `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time              `gorm:"not null" json:"updated_at"`
	DeletedAt   gorm.DeletedAt         `gorm:"index" json:"-"`

	// 关联
	Creator User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// TableName 表名
func (Policy) TableName() string {
	return "policies"
}

// PolicyVersion 策略版本模型
type PolicyVersion struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	PolicyID    uint      `gorm:"not null;index" json:"policy_id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	Version     int       `gorm:"not null" json:"version"`
	Enabled     bool      `gorm:"not null;default:true" json:"enabled"`
	Priority    int       `gorm:"not null;default:100" json:"priority"`
	Effect      string    `gorm:"size:10;not null;default:'allow'" json:"effect"`
	Subject     string    `gorm:"type:jsonb" json:"subject"`
	Resource    string    `gorm:"type:jsonb" json:"resource"`
	Action      string    `gorm:"type:jsonb" json:"action"`
	Environment string    `gorm:"type:jsonb" json:"environment"`
	Conditions  string    `gorm:"type:jsonb" json:"conditions"`
	Tags        []string  `gorm:"type:text[]" json:"tags"`
	TenantID    string    `gorm:"size:50;not null;index" json:"tenant_id"`
	CreatedBy   uint      `gorm:"not null" json:"created_by"`
	CreatedAt   time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null" json:"updated_at"`

	// 关联
	Creator User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// TableName 表名
func (PolicyVersion) TableName() string {
	return "policy_versions"
}

// PolicyExecution 策略执行记录
type PolicyExecution struct {
	ID           uint                   `gorm:"primaryKey" json:"id"`
	PolicyID     uint                   `gorm:"not null;index" json:"policy_id"`
	PolicyName   string                 `gorm:"size:100;not null" json:"policy_name"`
	PolicyVersion int                   `gorm:"not null" json:"policy_version"`
	RequestID    string                 `gorm:"size:100;not null;index" json:"request_id"`
	SubjectID    uint                   `gorm:"not null;index" json:"subject_id"`
	ResourceID   string                 `gorm:"size:100;not null" json:"resource_id"`
	ResourceType string                 `gorm:"size:50;not null" json:"resource_type"`
	Action       string                 `gorm:"size:50;not null" json:"action"`
	Decision     string                 `gorm:"size:10;not null" json:"decision"`
	Effect       string                 `gorm:"size:10;not null" json:"effect"`
	Reason       string                 `gorm:"size:500" json:"reason"`
	Duration     int                    `gorm:"not null" json:"duration"` // 毫秒
	TTL          int                    `json:"ttl"` // 缓存TTL秒数
	TenantID     string                 `gorm:"size:50;not null;index" json:"tenant_id"`
	IPAddress    string                 `gorm:"size:45" json:"ip_address"`
	UserAgent    string                 `gorm:"size:500" json:"user_agent"`
	CreatedAt    time.Time              `gorm:"not null;index" json:"created_at"`

	// 执行上下文
	ExecutionContext string `gorm:"type:jsonb" json:"execution_context"`
}

// TableName 表名
func (PolicyExecution) TableName() string {
	return "policy_executions"
}

// PolicyObligation 策略义务
type PolicyObligation struct {
	ID           uint                   `gorm:"primaryKey" json:"id"`
	PolicyID     uint                   `gorm:"not null;index" json:"policy_id"`
	Type         string                 `gorm:"size:50;not null" json:"type"`
	Description  string                 `gorm:"size:500" json:"description"`
	Attributes   string                 `gorm:"type:jsonb" json:"attributes"`
	Enabled      bool                   `gorm:"not null;default:true" json:"enabled"`
	Priority     int                    `gorm:"not null;default:100" json:"priority"`
	TenantID     string                 `gorm:"size:50;not null;index" json:"tenant_id"`
	CreatedBy    uint                   `gorm:"not null" json:"created_by"`
	CreatedAt    time.Time              `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time              `gorm:"not null" json:"updated_at"`
	DeletedAt    gorm.DeletedAt         `gorm:"index" json:"-"`

	// 关联
	Creator User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// TableName 表名
func (PolicyObligation) TableName() string {
	return "policy_obligations"
}

// PolicyTemplate 策略模板
type PolicyTemplate struct {
	ID          uint                   `gorm:"primaryKey" json:"id"`
	Name        string                 `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Description string                 `gorm:"size:500" json:"description"`
	Category    string                 `gorm:"size:50;not null" json:"category"`
	Subject     string                 `gorm:"type:jsonb" json:"subject"`
	Resource    string                 `gorm:"type:jsonb" json:"resource"`
	Action      string                 `gorm:"type:jsonb" json:"action"`
	Environment string                 `gorm:"type:jsonb" json:"environment"`
	Conditions  string                 `gorm:"type:jsonb" json:"conditions"`
	Tags        []string               `gorm:"type:text[]" json:"tags"`
	Enabled     bool                   `gorm:"not null;default:true" json:"enabled"`
	Public      bool                   `gorm:"not null;default:false" json:"public"`
	CreatedBy   uint                   `gorm:"not null" json:"created_by"`
	CreatedAt   time.Time              `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time              `gorm:"not null" json:"updated_at"`
	DeletedAt   gorm.DeletedAt         `gorm:"index" json:"-"`

	// 关联
	Creator User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// TableName 表名
func (PolicyTemplate) TableName() string {
	return "policy_templates"
}

// PolicyEvaluation 策略评估结果
type PolicyEvaluation struct {
	ID           uint                   `gorm:"primaryKey" json:"id"`
	RequestID    string                 `gorm:"size:100;not null;uniqueIndex" json:"request_id"`
	SubjectID    uint                   `gorm:"not null;index" json:"subject_id"`
	ResourceID   string                 `gorm:"size:100;not null" json:"resource_id"`
	ResourceType string                 `gorm:"size:50;not null" json:"resource_type"`
	Action       string                 `gorm:"size:50;not null" json:"action"`
	Allowed      bool                   `gorm:"not null" json:"allowed"`
	Effect       string                 `gorm:"size:10;not null" json:"effect"`
	Reason       string                 `gorm:"size:500" json:"reason"`
	Duration     int                    `gorm:"not null" json:"duration"`
	TTL          int                    `json:"ttl"`
	TenantID     string                 `gorm:"size:50;not null;index" json:"tenant_id"`
	EvaluatedAt  time.Time              `gorm:"not null" json:"evaluated_at"`
	ExpiresAt    time.Time              `gorm:"index" json:"expires_at"`

	// 匹配的策略
	MatchedPolicies string `gorm:"type:jsonb" json:"matched_policies"`
	AppliedPolicy  uint   `json:"applied_policy"`

	// 义务
	Obligations string `gorm:"type:jsonb" json:"obligations"`

	// 评估上下文
	EvaluationContext string `gorm:"type:jsonb" json:"evaluation_context"`
}

// TableName 表名
func (PolicyEvaluation) TableName() string {
	return "policy_evaluations"
}

// PolicyConflict 策略冲突
type PolicyConflict struct {
	ID           uint                   `gorm:"primaryKey" json:"id"`
	Policy1ID    uint                   `gorm:"not null;index" json:"policy1_id"`
	Policy2ID    uint                   `gorm:"not null;index" json:"policy2_id"`
	ConflictType string                 `gorm:"size:50;not null" json:"conflict_type"`
	Description  string                 `gorm:"size:500" json:"description"`
	Severity     string                 `gorm:"size:20;not null;default:'medium'" json:"severity"`
	Status       string                 `gorm:"size:20;not null;default:'open'" json:"status"`
	TenantID     string                 `gorm:"size:50;not null;index" json:"tenant_id"`
	ResolvedBy   *uint                  `json:"resolved_by"`
	ResolvedAt   *time.Time             `json:"resolved_at"`
	CreatedAt    time.Time              `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time              `gorm:"not null" json:"updated_at"`

	// 关联
	Policy1    Policy `gorm:"foreignKey:Policy1ID" json:"policy1,omitempty"`
	Policy2    Policy `gorm:"foreignKey:Policy2ID" json:"policy2,omitempty"`
	Resolver   *User  `gorm:"foreignKey:ResolvedBy" json:"resolver,omitempty"`
}

// TableName 表名
func (PolicyConflict) TableName() string {
	return "policy_conflicts"
}

// PolicyRecommendation 策略推荐
type PolicyRecommendation struct {
	ID            uint                   `gorm:"primaryKey" json:"id"`
	Name          string                 `gorm:"size:100;not null" json:"name"`
	Description   string                 `gorm:"size:500" json:"description"`
	Category      string                 `gorm:"size:50;not null" json:"category"`
	Subject       string                 `gorm:"type:jsonb" json:"subject"`
	Resource      string                 `gorm:"type:jsonb" json:"resource"`
	Action        string                 `gorm:"type:jsonb" json:"action"`
	Environment   string                 `gorm:"type:jsonb" json:"environment"`
	Conditions    string                 `gorm:"type:jsonb" json:"conditions"`
	Priority      int                    `gorm:"not null;default:100" json:"priority"`
	Effect        string                 `gorm:"size:10;not null;default:'allow'" json:"effect"`
	Justification string                 `gorm:"size:1000" json:"justification"`
	Tags          []string               `gorm:"type:text[]" json:"tags"`
	Enabled       bool                   `gorm:"not null;default:true" json:"enabled"`
	TenantID      string                 `gorm:"size:50;not null;index" json:"tenant_id"`
	CreatedBy     uint                   `gorm:"not null" json:"created_by"`
	CreatedAt     time.Time              `gorm:"not null" json:"created_at"`
	UpdatedAt     time.Time              `gorm:"not null" json:"updated_at"`

	// 关联
	Creator User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// TableName 表名
func (PolicyRecommendation) TableName() string {
	return "policy_recommendations"
}