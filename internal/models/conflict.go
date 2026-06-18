package models

import (
	"strings"
	"time"
)

// ConflictCheckRequest 冲突检测请求
type ConflictCheckRequest struct {
	CheckID                   string              `json:"checkId,omitempty"`
	ClientID                  string              `json:"clientId" validate:"required"`
	ClientName                string              `json:"clientName" validate:"required"`
	ClientType                string              `json:"clientType" validate:"oneof=PERSON COMPANY ANY"`
	OtherParties              []string            `json:"otherParties"`
	Parties                   []ConflictPartyInfo `json:"parties,omitempty"`
	CaseName                  string              `json:"caseName" validate:"required"`
	CaseType                  string              `json:"caseType" validate:"required"`
	SearchYears               int                 `json:"searchYears"`
	IncludeCorporateRelations bool                `json:"includeCorporateRelations"`
	SearchDepth               string              `json:"searchDepth" validate:"oneof=BASIC STANDARD DEEP"`
	UserID                    string              `json:"userId"` // 改为字符串类型，前端会发送字符串
	RequestTime               time.Time           `json:"requestTime"`
}

// ConflictPartyInfo describes the legal role of a party being checked.
type ConflictPartyInfo struct {
	Name         string `json:"name"`
	Role         string `json:"role"`
	EntityType   string `json:"entityType,omitempty"`
	RelationNote string `json:"relationNote,omitempty"`
}

// ConflictCheckResponse 冲突检测响应
type ConflictCheckResponse struct {
	CheckID         string           `json:"checkId"`
	HasConflict     bool             `json:"hasConflict"`
	ConflictCases   []*ConflictCase  `json:"conflictCases"`
	CheckStatistics *CheckStatistics `json:"checkStatistics"`
	RiskAssessment  *RiskAssessment  `json:"riskAssessment"`
	Recommendations []string         `json:"recommendations"`
	CheckTime       time.Time        `json:"checkTime"`
	Duration        int64            `json:"duration"`
	MCPStandards    *MCPStandards    `json:"mcpStandards"`
}

// ConflictCase 冲突案例
type ConflictCase struct {
	ID              string          `json:"id" gorm:"primarykey"`
	CheckID         string          `json:"checkId" gorm:"index"`
	CaseID          string          `json:"caseId"`
	CaseName        string          `json:"caseName"`
	CaseNo          string          `json:"caseNo"`
	CaseType        string          `json:"caseType" gorm:"column:case_type"`
	ConflictType    string          `json:"conflictType"`
	RiskLevel       string          `json:"riskLevel"`
	Description     string          `json:"description"`
	CaseStatus      string          `json:"caseStatus"`
	ClientID        string          `json:"clientId"`
	OpposingParties JSONStringArray `json:"opposingParties" gorm:"type:json"`
	ConflictDetails string          `json:"conflictDetails"`
	CreatedAt       time.Time       `json:"createdAt"`
}

// RuleMatch 规则匹配结果
type RuleMatch struct {
	RuleID     string  `json:"ruleId"`
	RuleName   string  `json:"ruleName"`
	Matched    bool    `json:"matched"`
	RiskScore  float64 `json:"riskScore"`
	Reason     string  `json:"reason"`
	Confidence float64 `json:"confidence"`
}

// ConflictRule 冲突检测规则
type ConflictRule struct {
	ID          string          `json:"id" gorm:"primarykey"`
	Name        string          `json:"name" validate:"required"`
	Description string          `json:"description"`
	Type        string          `json:"type" validate:"required"`
	Category    string          `json:"category" validate:"required"`
	Conditions  JSON            `json:"conditions" gorm:"type:json"`
	Actions     JSONStringArray `json:"actions" gorm:"type:json"`
	Priority    int             `json:"priority"`
	Active      bool            `json:"active" gorm:"default:true"`
	Version     int             `json:"version"`
	MCPSource   string          `json:"mcpSource"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

// RiskAssessment 风险评估
type RiskAssessment struct {
	OverallRisk      string   `json:"overallRisk"`
	RiskScore        float64  `json:"riskScore"`
	RiskReason       string   `json:"riskReason"`
	RequiresApproval bool     `json:"requiresApproval"`
	ApprovalLevel    string   `json:"approvalLevel,omitempty"`
	RiskFactors      []string `json:"riskFactors"`
	Mitigation       []string `json:"mitigation"`
}

// MCPStandards MCP标准
type MCPStandards struct {
	Version        string          `json:"version"`
	LastUpdated    time.Time       `json:"lastUpdated"`
	Standards      JSON            `json:"standards" gorm:"type:json"`
	BestPractices  JSONStringArray `json:"bestPractices" gorm:"type:json"`
	Compliance     JSONStringArray `json:"compliance" gorm:"type:json"`
	RiskThresholds JSON            `json:"riskThresholds" gorm:"type:json"`
	Active         bool            `json:"active" gorm:"default:true"`
}

// CheckStatistics 检查统计
type CheckStatistics struct {
	TotalCasesChecked         int64     `json:"totalCasesChecked"`
	ClientHistoryCases        int64     `json:"clientHistoryCases"`
	RelatedPartiesChecked     int64     `json:"relatedPartiesChecked"`
	CorporateRelationsChecked int64     `json:"corporateRelationsChecked"`
	TimeRange                 string    `json:"timeRange"`
	SearchScope               string    `json:"searchScope"`
	StartTime                 time.Time `json:"startTime"`
	EndTime                   time.Time `json:"endTime"`
}

// ConflictCheckRecord 冲突检测记录
type ConflictCheckRecord struct {
	CheckID          string    `json:"checkId" gorm:"primaryKey;column:check_id;type:varchar"`
	ClientID         string    `json:"clientId" gorm:"column:client_id;index;type:varchar"`
	ClientName       string    `json:"clientName" gorm:"column:client_name"`
	CaseName         string    `json:"caseName" gorm:"column:case_name"`
	CaseType         string    `json:"caseType" gorm:"column:case_type"`
	CheckStatus      string    `json:"checkStatus" gorm:"column:check_status;default:PROCESSING"`
	HasConflict      bool      `json:"hasConflict" gorm:"column:has_conflict"`
	RiskLevel        string    `json:"riskLevel" gorm:"column:risk_level"`
	SearchParameters JSON      `json:"searchParameters" gorm:"column:search_parameters;type:json"`
	CheckResult      JSON      `json:"checkResult" gorm:"column:check_result;type:json"`
	UserID           uint      `json:"userId" gorm:"column:user_id"`
	Duration         int64     `json:"duration" gorm:"column:duration"`
	CheckTime        time.Time `json:"checkTime" gorm:"column:check_time"`
	CreatedAt        time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt        time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

// ClientRelation 客户关系
type ClientRelation struct {
	ID              string    `json:"id" gorm:"primarykey"`
	ClientID        string    `json:"clientId" gorm:"index"`
	RelatedClientID string    `json:"relatedClientId" gorm:"index"`
	RelationType    string    `json:"relationType"`
	RelationDetail  string    `json:"relationDetail"`
	Active          bool      `json:"active" gorm:"default:true"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// Validation functions

// ValidateConflictCheckRequest 验证冲突检测请求
func (r *ConflictCheckRequest) Validate() error {
	r.ClientID = strings.TrimSpace(r.ClientID)
	r.ClientName = strings.TrimSpace(r.ClientName)
	r.ClientType = strings.ToUpper(strings.TrimSpace(r.ClientType))
	r.CaseName = strings.TrimSpace(r.CaseName)
	r.CaseType = strings.TrimSpace(r.CaseType)
	r.SearchDepth = strings.ToUpper(strings.TrimSpace(r.SearchDepth))
	r.UserID = strings.TrimSpace(r.UserID)

	if r.ClientID == "" {
		return &ConflictError{Code: "VALIDATION_001", Message: "客户ID不能为空"}
	}
	if r.ClientName == "" {
		return &ConflictError{Code: "VALIDATION_002", Message: "客户名称不能为空"}
	}
	// 🔧 修复：允许ANY类型，与前端validate标签保持一致
	if r.ClientType != "PERSON" && r.ClientType != "COMPANY" && r.ClientType != "ANY" {
		return &ConflictError{Code: "VALIDATION_003", Message: "客户类型必须是PERSON、COMPANY或ANY"}
	}
	if r.CaseName == "" {
		return &ConflictError{Code: "VALIDATION_004", Message: "案件名称不能为空"}
	}
	if r.CaseType == "" {
		return &ConflictError{Code: "VALIDATION_005", Message: "案件类型不能为空"}
	}
	// 🔧 修复：验证前端发送的英文案件类型值
	validCaseTypes := map[string]bool{
		"civil": true, "commercial": true, "criminal": true,
		"administrative": true, "labor": true, "intellectual": true,
		"financial": true, "arbitration": true, "consultation": true,
		"other": true, "知识产权": true, "民事": true, "商事": true,
		"刑事": true, "行政": true, "劳动": true, "金融": true,
		"仲裁": true, "咨询": true, "其他": true,
	}
	if !validCaseTypes[r.CaseType] {
		return &ConflictError{Code: "VALIDATION_005", Message: "案件类型无效，支持的类型: civil, commercial, criminal, administrative, labor, intellectual, financial, arbitration, consultation, other"}
	}
	if r.SearchDepth != "" && r.SearchDepth != "BASIC" && r.SearchDepth != "STANDARD" && r.SearchDepth != "DEEP" {
		return &ConflictError{Code: "VALIDATION_006", Message: "搜索深度必须是BASIC、STANDARD或DEEP"}
	}
	if r.SearchYears < 0 || r.SearchYears > 20 {
		return &ConflictError{Code: "VALIDATION_011", Message: "搜索年限必须为0或1-20年之间；0表示使用默认值"}
	}
	if r.UserID == "" {
		return &ConflictError{Code: "VALIDATION_012", Message: "承办律师ID不能为空"}
	}

	cleanParties := make([]string, 0, len(r.OtherParties))
	seen := make(map[string]struct{}, len(r.OtherParties))
	for _, party := range r.OtherParties {
		party = strings.TrimSpace(party)
		if party == "" {
			continue
		}
		if _, ok := seen[party]; ok {
			continue
		}
		seen[party] = struct{}{}
		cleanParties = append(cleanParties, party)
	}
	r.OtherParties = cleanParties

	cleanPartyInfo := make([]ConflictPartyInfo, 0, len(r.Parties))
	for _, party := range r.Parties {
		party.Name = strings.TrimSpace(party.Name)
		party.Role = strings.ToUpper(strings.TrimSpace(party.Role))
		party.EntityType = strings.ToUpper(strings.TrimSpace(party.EntityType))
		party.RelationNote = strings.TrimSpace(party.RelationNote)
		if party.Name == "" {
			continue
		}
		cleanPartyInfo = append(cleanPartyInfo, party)
		if party.Role == "OPPOSING_PARTY" || party.Role == "OPPOSING" || party.Role == "ADVERSE" {
			if _, ok := seen[party.Name]; !ok {
				seen[party.Name] = struct{}{}
				r.OtherParties = append(r.OtherParties, party.Name)
			}
		}
	}
	r.Parties = cleanPartyInfo
	return nil
}

// ValidateConflictRule 验证冲突检测规则
func (r *ConflictRule) Validate() error {
	if r.Name == "" {
		return &ConflictError{Code: "VALIDATION_007", Message: "规则名称不能为空"}
	}
	if r.Type == "" {
		return &ConflictError{Code: "VALIDATION_008", Message: "规则类型不能为空"}
	}
	if r.Category == "" {
		return &ConflictError{Code: "VALIDATION_009", Message: "规则分类不能为空"}
	}
	if len(r.Conditions.ToMap()) == 0 {
		return &ConflictError{Code: "VALIDATION_010", Message: "规则条件不能为空"}
	}
	return nil
}

// Helper functions

// GetRiskScoreLevel 获取风险等级
func GetRiskScoreLevel(score float64) string {
	switch {
	case score >= 0.8:
		return "CRITICAL"
	case score >= 0.6:
		return "HIGH"
	case score >= 0.4:
		return "MEDIUM"
	case score >= 0.2:
		return "LOW"
	default:
		return "MINIMAL"
	}
}

// GetRiskLevelColor 获取风险等级对应的颜色
func GetRiskLevelColor(level string) string {
	switch level {
	case "CRITICAL":
		return "#dc3545" // red
	case "HIGH":
		return "#fd7e14" // orange
	case "MEDIUM":
		return "#ffc107" // yellow
	case "LOW":
		return "#28a745" // green
	case "MINIMAL":
		return "#17a2b8" // blue
	default:
		return "#6c757d" // gray
	}
}

// ConflictError 冲突检测错误
type ConflictError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *ConflictError) Error() string {
	return e.Message
}

// Predefined errors
var (
	ErrMCPServiceUnavailable   = &ConflictError{Code: "MCP_001", Message: "MCP服务不可用"}
	ErrRuleExecutionFailed     = &ConflictError{Code: "RULE_001", Message: "规则执行失败"}
	ErrDataValidationFailed    = &ConflictError{Code: "DATA_001", Message: "数据验证失败"}
	ErrPermissionDenied        = &ConflictError{Code: "AUTH_001", Message: "权限不足"}
	ErrDatabaseOperationFailed = &ConflictError{Code: "DB_001", Message: "数据库操作失败"}
	ErrCacheOperationFailed    = &ConflictError{Code: "CACHE_001", Message: "缓存操作失败"}
	ErrTimeoutExceeded         = &ConflictError{Code: "TIMEOUT_001", Message: "操作超时"}
	ErrRateLimitExceeded       = &ConflictError{Code: "RATE_001", Message: "请求频率超限"}
)
