package models

import (
	"time"
)

// IndustryClassification 行业分类定义
type IndustryClassification struct {
	ID          int    `json:"id" gorm:"primaryKey"`
	Code        string `json:"code" gorm:"unique;not null"`        // 行业代码，如 "TMT"、"FINANCE"
	Name        string `json:"name" gorm:"not null"`             // 行业名称，如 "科技、媒体和通信"
	ParentID    *int   `json:"parent_id"`                        // 父级行业ID，支持层级分类
	Level       int    `json:"level" gorm:"default:1"`           // 层级，1=一级，2=二级，3=三级
	Description string `json:"description"`                      // 行业描述
	Keywords    string `json:"keywords" gorm:"type:text"`        // 关键词，逗号分隔
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CompetitiveRelation 竞争关系定义
type CompetitiveRelation struct {
	ID               int    `json:"id" gorm:"primaryKey"`
	IndustryID       int    `json:"industry_id"`                          // 行业ID
	CompetitorType   string `json:"competitor_type" gorm:"size:50"`      // 竞争者类型：direct, indirect, substitute
	CompetitorName   string `json:"competitor_name" gorm:"not null"`     // 竞争者名称
	CompetitorPattern string `json:"competitor_pattern" gorm:"type:text"` // 竞争者匹配模式（正则或关键词）
	ConflictLevel    string `json:"conflict_level" gorm:"size:20"`       // 冲突等级：HIGH, MEDIUM, LOW
	Description      string `json:"description"`                         // 关系描述
	IsActive         bool   `json:"is_active" gorm:"default:true"`       // 是否启用
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// EnhancedConflictRule 增强的冲突检测规则
type EnhancedConflictRule struct {
	ID             int    `json:"id" gorm:"primaryKey"`
	Name           string `json:"name" gorm:"not null"`              // 规则名称
	RuleType       string `json:"rule_type" gorm:"size:50"`          // 规则类型：industry, client, case_type, time_range
	TriggerPattern string `json:"trigger_pattern" gorm:"type:text"`  // 触发模式
	ActionType     string `json:"action_type" gorm:"size:50"`        // 动作类型：block, warn, notify
	RiskScore      int    `json:"risk_score" gorm:"default:50"`      // 风险分数
	Conditions     string `json:"conditions" gorm:"type:text"`       // 条件JSON
	IsActive       bool   `json:"is_active" gorm:"default:true"`     // 是否启用
	Priority       int    `json:"priority" gorm:"default:100"`       // 优先级
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ConflictDetectionHistory 冲突检测历史
type ConflictDetectionHistory struct {
	ID               int    `json:"id" gorm:"primaryKey"`
	LawyerID         int    `json:"lawyer_id"`
	CaseID           *int   `json:"case_id"`
	ClientName       string `json:"client_name"`
	OpposingParty    string `json:"opposing_party"`
	CaseType         string `json:"case_type"`
	DetectionResult  string `json:"detection_result" gorm:"type:text"`  // 检测结果JSON
	ConflictsFound   int    `json:"conflicts_found"`                     // 发现的冲突数
	RiskLevel        string `json:"risk_level" gorm:"size:20"`           // 风险等级
	UserAction       string `json:"user_action" gorm:"size:50"`          // 用户操作
	CreatedAt        time.Time `json:"created_at"`
}

// ConflictCaseConflict 冲突案件关联表（多对多）
type ConflictCaseConflict struct {
	ID            int `json:"id" gorm:"primaryKey"`
	CaseID        int `json:"case_id"`
	ConflictCaseID int `json:"conflict_case_id"`
	ConflictType  string `json:"conflict_type" gorm:"size:50"`  // 冲突类型
	RiskScore     int    `json:"risk_score"`                    // 此冲突的风险分数
	CreatedAt     time.Time `json:"created_at"`
}