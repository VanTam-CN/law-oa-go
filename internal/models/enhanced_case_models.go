package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// EnhancedCase 增强的案例模型
// 扩展原有Case模型以支持多客户、冲突检测、豁免管理等高级功能
type EnhancedCase struct {
	gorm.Model
	// 基础字段（继承自原有Case模型）
	Title       string  `json:"title" gorm:"size:200;not null;comment:案件标题"`
	Description string  `json:"description" gorm:"type:text;comment:案件描述"`
	ClientID    *uint   `json:"client_id,omitempty" gorm:"comment:原有客户ID，保持向后兼容"`
	CaseType    string  `json:"case_type" gorm:"size:50;not null;comment:案件类型"`
	Priority    string  `json:"priority" gorm:"size:20;default:'medium';comment:优先级"`
	Status      string  `json:"status" gorm:"size:20;default:'pending';comment:案件状态"`
	StartDate   *time.Time `json:"start_date,omitempty" gorm:"comment:开始日期"`
	EndDate     *time.Time `json:"end_date,omitempty" gorm:"comment:结束日期"`

	// 增强字段 - 客户管理
	ClientProfileIDs ClientProfileIDs `json:"client_profile_ids" gorm:"type:jsonb;default:'[]'::jsonb;comment:客户档案ID列表"`

	// 增强字段 - 冲突检测
	ConflictCheckRequestID   *string              `json:"conflict_check_request_id,omitempty" gorm:"size:36;comment:冲突检测请求ID"`
	ConflictDetectionStatus  ConflictDetectionStatus `json:"conflict_detection_status" gorm:"size:20;default:'PENDING';comment:冲突检测状态"`
	ConflictDetectionResult  *ConflictDetectionResult `json:"conflict_detection_result,omitempty" gorm:"type:jsonb;comment:冲突检测结果"`
	RiskLevel                RiskLevel            `json:"risk_level,omitempty" gorm:"size:20;comment:整体风险等级"`

	// 增强字段 - 豁免管理
	WaiverApplicationID *string       `json:"waiver_application_id,omitempty" gorm:"size:36;comment:豁免申请ID"`
	WaiverStatus         WaiverStatus  `json:"waiver_status,omitempty" gorm:"size:20;default:'NONE';comment:豁免状态"`

	// 增强字段 - 信息屏障
	EthicalScreenEstablished bool               `json:"ethical_screen_established" gorm:"default:false;comment:是否建立信息屏障"`

	// 增强字段 - 业务管理
	AssignedBy         *string              `json:"assigned_by,omitempty" gorm:"size:36;comment:分配者ID"`
	PracticeArea       string               `json:"practice_area" gorm:"size:50;comment:业务领域"`
	EstimatedDuration  string               `json:"estimated_duration" gorm:"size:50;comment:预估持续时间"`
	BillingMethod      string               `json:"billing_method" gorm:"size:50;comment:计费方式"`
	TeamAssignment     TeamAssignment       `json:"team_assignment" gorm:"type:jsonb;default:'{}'::jsonb;comment:团队分配信息"`
	ConflictMetadata   ConflictMetadata     `json:"conflict_metadata" gorm:"type:jsonb;default:'{}'::jsonb;comment:冲突检测元数据"`
	CreatedViaConflictCheck bool             `json:"created_via_conflict_check" gorm:"default:false;comment:是否通过冲突检测流程创建"`

	// 关联字段
	ClientProfiles          []ClientProfile           `json:"client_profiles,omitempty" gorm:"many2many:case_client_profiles;"`
	ConflictCheckRequest    *ProfessionalConflictCheckRequest `json:"conflict_check_request,omitempty" gorm:"foreignKey:ConflictCheckRequestID"`
	WaiverApplication       *WaiverApplication         `json:"waiver_application,omitempty" gorm:"foreignKey:WaiverApplicationID"`
	ConflictRecords         []CaseConflictRecord      `json:"conflict_records,omitempty" gorm:"foreignKey:CaseID"`
	WaiverAssociations      []CaseWaiverAssociation   `json:"waiver_associations,omitempty" gorm:"foreignKey:CaseID"`
	EthicalScreens          []CaseEthicalScreen       `json:"ethical_screens,omitempty" gorm:"foreignKey:CaseID"`
}

// CaseClientProfile 案例-客户档案关联
type CaseClientProfile struct {
	ID                      string    `json:"id" gorm:"type:varchar(36);primaryKey;comment:主键ID"`
	CaseID                  uint      `json:"case_id" gorm:"not null;comment:案例ID"`
	ClientProfileID         string    `json:"client_profile_id" gorm:"size:36;not null;comment:客户档案ID"`
	ClientRole              ClientRole `json:"client_role" gorm:"size:50;not null;default:'PRIMARY';comment:客户角色"`
	RelationshipDescription *string   `json:"relationship_description,omitempty" gorm:"type:text;comment:关系描述"`
	CreatedAt               time.Time `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt               time.Time `json:"updated_at" gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt               gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index;comment:软删除时间"`

	// 关联字段
	Case          *EnhancedCase   `json:"case,omitempty" gorm:"foreignKey:CaseID"`
	ClientProfile *ClientProfile  `json:"client_profile,omitempty" gorm:"foreignKey:ClientProfileID"`
}

// CaseConflictRecord 案例冲突检测记录
type CaseConflictRecord struct {
	ID                      string                `json:"id" gorm:"type:varchar(36);primaryKey;comment:主键ID"`
	CaseID                  uint                  `json:"case_id" gorm:"not null;comment:案例ID"`
	ConflictCheckRequestID  string                `json:"conflict_check_request_id" gorm:"size:36;not null;comment:冲突检测请求ID"`
	DetectionDate           time.Time             `json:"detection_date" gorm:"comment:检测时间"`
	ConflictTypesDetected   ConflictTypes         `json:"conflict_types_detected" gorm:"type:jsonb;default:'[]'::jsonb;comment:检测到的冲突类型"`
	RiskAssessment          EnhancedRiskAssessment `json:"risk_assessment" gorm:"type:jsonb;not null;comment:风险评估"`
	AffectedParties         AffectedParties       `json:"affected_parties" gorm:"type:jsonb;default:'[]'::jsonb;comment:受影响的相关方"`
	RecommendedActions      RecommendedActions    `json:"recommended_actions" gorm:"type:jsonb;default:'[]'::jsonb;comment:建议措施"`
	DetectionRulesApplied   DetectionRulesApplied `json:"detection_rules_applied" gorm:"type:jsonb;default:'[]'::jsonb;comment:应用的检测规则"`
	Status                  string                `json:"status" gorm:"size:20;default:'COMPLETED';comment:检测状态"`
	CreatedAt               time.Time             `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt               time.Time             `json:"updated_at" gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt               gorm.DeletedAt        `json:"deleted_at,omitempty" gorm:"index;comment:软删除时间"`

	// 关联字段
	Case                 *EnhancedCase                  `json:"case,omitempty" gorm:"foreignKey:CaseID"`
	ConflictCheckRequest *ProfessionalConflictCheckRequest `json:"conflict_check_request,omitempty" gorm:"foreignKey:ConflictCheckRequestID"`
}

// CaseWaiverAssociation 案例豁免关联
type CaseWaiverAssociation struct {
	ID                     string            `json:"id" gorm:"type:varchar(36);primaryKey;comment:主键ID"`
	CaseID                 uint              `json:"case_id" gorm:"not null;comment:案例ID"`
	WaiverApplicationID    string            `json:"waiver_application_id" gorm:"size:36;not null;comment:豁免申请ID"`
	AssociationType        WaiverAssociationType `json:"association_type" gorm:"size:50;not null;default:'REQUIRED';comment:关联类型"`
	ConflictSummary        string            `json:"conflict_summary" gorm:"type:text;not null;comment:冲突摘要"`
	WaiverConditions       WaiverConditions  `json:"waiver_conditions" gorm:"type:jsonb;default:'{}'::jsonb;comment:豁免条件"`
	MonitoringRequirements MonitoringRequirements `json:"monitoring_requirements" gorm:"type:jsonb;default:'{}'::jsonb;comment:监控要求"`
	Status                 string            `json:"status" gorm:"size:20;default:'PENDING';comment:关联状态"`
	CreatedAt              time.Time         `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt              time.Time         `json:"updated_at" gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt              gorm.DeletedAt    `json:"deleted_at,omitempty" gorm:"index;comment:软删除时间"`

	// 关联字段
	Case              *EnhancedCase     `json:"case,omitempty" gorm:"foreignKey:CaseID"`
	WaiverApplication *WaiverApplication `json:"waiver_application,omitempty" gorm:"foreignKey:WaiverApplicationID"`
}

// CaseEthicalScreen 案例信息屏障
type CaseEthicalScreen struct {
	ID                   string              `json:"id" gorm:"type:varchar(36);primaryKey;comment:主键ID"`
	CaseID               uint                `json:"case_id" gorm:"not null;comment:案例ID"`
	ScreenType           EthicalScreenType   `json:"screen_type" gorm:"size:50;not null;comment:信息屏障类型"`
	ScreenedLawyers      ScreenedLawyers     `json:"screened_lawyers" gorm:"type:jsonb;not null;default:'[]'::jsonb;comment:被屏障的律师"`
	ScreenedTeams        ScreenedTeams       `json:"screened_teams" gorm:"type:jsonb;not null;default:'[]'::jsonb;comment:被屏障的团队"`
	RestrictedInformation RestrictedInformation `json:"restricted_information" gorm:"type:jsonb;not null;default:'{}'::jsonb;comment:受限信息"`
	AccessPermissions    AccessPermissions    `json:"access_permissions" gorm:"type:jsonb;not null;default:'{}'::jsonb;comment:访问权限"`
	MonitoringPlan       MonitoringPlan      `json:"monitoring_plan" gorm:"type:jsonb;not null;default:'{}'::jsonb;comment:监控计划"`
	EffectiveDate        time.Time           `json:"effective_date" gorm:"comment:生效时间"`
	ExpiryDate           *time.Time          `json:"expiry_date,omitempty" gorm:"comment:到期时间"`
	Status               string              `json:"status" gorm:"size:20;default:'ACTIVE';comment:状态"`
	EstablishedBy        string              `json:"established_by" gorm:"size:36;not null;comment:建立人ID"`
	CreatedAt            time.Time           `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt            time.Time           `json:"updated_at" gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt            gorm.DeletedAt      `json:"deleted_at,omitempty" gorm:"index;comment:软删除时间"`

	// 关联字段
	Case         *EnhancedCase `json:"case,omitempty" gorm:"foreignKey:CaseID"`
	EstablishedByUser *User    `json:"established_by_user,omitempty" gorm:"foreignKey:EstablishedBy"`
}

// 自定义类型定义

// ClientProfileIDs 客户档案ID列表类型
type ClientProfileIDs []string

func (c ClientProfileIDs) Value() (driver.Value, error) {
	if len(c) == 0 {
		return "[]", nil
	}
	return json.Marshal(c)
}

func (c *ClientProfileIDs) Scan(value interface{}) error {
	if value == nil {
		*c = ClientProfileIDs{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, c)
	case string:
		return json.Unmarshal([]byte(v), c)
	default:
		return fmt.Errorf("cannot scan %T into ClientProfileIDs", value)
	}
}

// ConflictDetectionStatus 冲突检测状态
type ConflictDetectionStatus string

const (
	ConflictDetectionStatusPending     ConflictDetectionStatus = "PENDING"
	ConflictDetectionStatusInProgress ConflictDetectionStatus = "IN_PROGRESS"
	ConflictDetectionStatusCompleted  ConflictDetectionStatus = "COMPLETED"
	ConflictDetectionStatusFailed     ConflictDetectionStatus = "FAILED"
	ConflictDetectionStatusNotRequired ConflictDetectionStatus = "NOT_REQUIRED"
)

// RiskLevel 风险等级
type RiskLevel string

const (
	RiskLevelLow       RiskLevel = "LOW"
	RiskLevelMedium    RiskLevel = "MEDIUM"
	RiskLevelHigh      RiskLevel = "HIGH"
	RiskLevelCritical  RiskLevel = "CRITICAL"
	RiskLevelNotAssessed RiskLevel = "NOT_ASSESSED"
)

// WaiverStatus 豁免状态
type WaiverStatus string

const (
	WaiverStatusNone     WaiverStatus = "NONE"
	WaiverStatusPending  WaiverStatus = "PENDING"
	WaiverStatusApproved WaiverStatus = "APPROVED"
	WaiverStatusRejected WaiverStatus = "REJECTED"
	WaiverStatusExpired  WaiverStatus = "EXPIRED"
)

// ClientRole 客户角色
type ClientRole string

const (
	ClientRolePrimary     ClientRole = "PRIMARY"
	ClientRoleSecondary   ClientRole = "SECONDARY"
	ClientRoleOpposing    ClientRole = "OPPOSING"
	ClientRoleThirdParty  ClientRole = "THIRD_PARTY"
)

// WaiverAssociationType 豁免关联类型
type WaiverAssociationType string

const (
	WaiverAssociationTypeRequired  WaiverAssociationType = "REQUIRED"
	WaiverAssociationTypeElective WaiverAssociationType = "ELECTIVE"
	WaiverAssociationTypeProactive WaiverAssociationType = "PROACTIVE"
)

// EthicalScreenType 信息屏障类型
type EthicalScreenType string

const (
	EthicalScreenTypeInformationBarrier EthicalScreenType = "INFORMATION_BARRIER"
	EthicalScreenTypeChineseWall        EthicalScreenType = "CHINESE_WALL"
	EthicalScreenTypeEthicalWall        EthicalScreenType = "ETHICAL_WALL"
)

// JSON类型定义

// ConflictDetectionResult 冲突检测结果
type ConflictDetectionResult struct {
	TotalConflicts      int                    `json:"total_conflicts"`
	HighRiskConflicts   int                    `json:"high_risk_conflicts"`
	MediumRiskConflicts int                    `json:"medium_risk_conflicts"`
	LowRiskConflicts    int                    `json:"low_risk_conflicts"`
	DetectionSummary    string                 `json:"detection_summary"`
	WaiverRequired      bool                   `json:"waiver_required"`
	WaiverPossible      bool                   `json:"waiver_possible"`
	DetectedConflicts   []ConflictItem         `json:"detected_conflicts"`
	Recommendations     []string               `json:"recommendations"`
	NextSteps          []string               `json:"next_steps"`
}

// ConflictItem 冲突项
type ConflictItem struct {
	Type           string   `json:"type"`
	Severity       string   `json:"severity"`
	Description    string   `json:"description"`
	AffectedParties []string `json:"affected_parties"`
	WaiverPossible bool     `json:"waiver_possible"`
	Rules          []string `json:"rules"`
}

// TeamAssignment 团队分配信息
type TeamAssignment struct {
	PrimaryLawyer      string                 `json:"primary_lawyer"`
	AssistingLawyers   []string               `json:"assisting_lawyers"`
	TeamMembers        []TeamMember           `json:"team_members"`
	AllocationRules    map[string]interface{} `json:"allocation_rules"`
	WorkloadDistribution map[string]int       `json:"workload_distribution"`
}

// TeamMember 团队成员
type TeamMember struct {
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	Capacity  int    `json:"capacity"`
	Assigned  bool   `json:"assigned"`
}

// ConflictMetadata 冲突检测元数据
type ConflictMetadata struct {
	LastCheckTime     time.Time              `json:"last_check_time"`
	CheckTrigger      string                 `json:"check_trigger"`
	CheckedBy         string                 `json:"checked_by"`
	CheckScope        []string               `json:"check_scope"`
	AppliedRules      []string               `json:"applied_rules"`
	ProcessingTime    int                    `json:"processing_time_ms"`
	SystemVersion     string                 `json:"system_version"`
}

// ConflictTypes 冲突类型列表
type ConflictTypes []string

func (c ConflictTypes) Value() (driver.Value, error) {
	if len(c) == 0 {
		return "[]", nil
	}
	return json.Marshal(c)
}

func (c *ConflictTypes) Scan(value interface{}) error {
	if value == nil {
		*c = ConflictTypes{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, c)
	case string:
		return json.Unmarshal([]byte(v), c)
	default:
		return fmt.Errorf("cannot scan %T into ConflictTypes", value)
	}
}

// EnhancedRiskAssessment 增强的风险评估
type EnhancedRiskAssessment struct {
	OverallScore    float64                `json:"overall_score"`
	RiskLevel       string                 `json:"risk_level"`
	RiskFactors     []RiskFactor           `json:"risk_factors"`
	MitigationPlan  []string               `json:"mitigation_plan"`
	ConfidenceLevel float64                `json:"confidence_level"`
}

// RiskFactor 风险因子
type RiskFactor struct {
	Name        string  `json:"name"`
	Score       float64 `json:"score"`
	Weight      float64 `json:"weight"`
	Description string  `json:"description"`
}

// AffectedParties 受影响的相关方
type AffectedParties []AffectedParty

func (a AffectedParties) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "[]", nil
	}
	return json.Marshal(a)
}

func (a *AffectedParties) Scan(value interface{}) error {
	if value == nil {
		*a = AffectedParties{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, a)
	case string:
		return json.Unmarshal([]byte(v), a)
	default:
		return fmt.Errorf("cannot scan %T into AffectedParties", value)
	}
}

// AffectedParty 受影响方
type AffectedParty struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Role        string   `json:"role"`
	ImpactLevel string   `json:"impact_level"`
	Concerns    []string `json:"concerns"`
}

// RecommendedActions 建议措施
type RecommendedActions []string

func (r RecommendedActions) Value() (driver.Value, error) {
	if len(r) == 0 {
		return "[]", nil
	}
	return json.Marshal(r)
}

func (r *RecommendedActions) Scan(value interface{}) error {
	if value == nil {
		*r = RecommendedActions{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, r)
	case string:
		return json.Unmarshal([]byte(v), r)
	default:
		return fmt.Errorf("cannot scan %T into RecommendedActions", value)
	}
}

// DetectionRulesApplied 应用的检测规则
type DetectionRulesApplied []DetectionRule

func (d DetectionRulesApplied) Value() (driver.Value, error) {
	if len(d) == 0 {
		return "[]", nil
	}
	return json.Marshal(d)
}

func (d *DetectionRulesApplied) Scan(value interface{}) error {
	if value == nil {
		*d = DetectionRulesApplied{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, d)
	case string:
		return json.Unmarshal([]byte(v), d)
	default:
		return fmt.Errorf("cannot scan %T into DetectionRulesApplied", value)
	}
}

// DetectionRule 检测规则
type DetectionRule struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Priority    int     `json:"priority"`
	Description string  `json:"description"`
	Triggered   bool    `json:"triggered"`
	Score       float64 `json:"score"`
}

// WaiverConditions 豁免条件
type WaiverConditions struct {
	ConsentRequired      bool     `json:"consent_required"`
	DisclosureLevel      string   `json:"disclosure_level"`
	MonitoringRequired  bool     `json:"monitoring_required"`
	Restrictions         []string `json:"restrictions"`
	ApprovalRequirements []string `json:"approval_requirements"`
}

// MonitoringRequirements 监控要求
type MonitoringRequirements struct {
	Frequency       string   `json:"frequency"`
	MonitoringTypes []string `json:"monitoring_types"`
	ReportSchedule  string   `json:"report_schedule"`
	EscalationRules []string `json:"escalation_rules"`
}

// ScreenedLawyers 被屏障的律师列表
type ScreenedLawyers []ScreenedLawyer

func (s ScreenedLawyers) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "[]", nil
	}
	return json.Marshal(s)
}

func (s *ScreenedLawyers) Scan(value interface{}) error {
	if value == nil {
		*s = ScreenedLawyers{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return fmt.Errorf("cannot scan %T into ScreenedLawyers", value)
	}
}

// ScreenedLawyer 被屏障的律师
type ScreenedLawyer struct {
	UserID      string   `json:"user_id"`
	Name        string   `json:"name"`
	Department  string   `json:"department"`
	ScreenType  string   `json:"screen_type"`
	StartDate   time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Reason      string   `json:"reason"`
}

// ScreenedTeams 被屏障的团队列表
type ScreenedTeams []ScreenedTeam

func (s ScreenedTeams) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "[]", nil
	}
	return json.Marshal(s)
}

func (s *ScreenedTeams) Scan(value interface{}) error {
	if value == nil {
		*s = ScreenedTeams{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return fmt.Errorf("cannot scan %T into ScreenedTeams", value)
	}
}

// ScreenedTeam 被屏障的团队
type ScreenedTeam struct {
	TeamID      string   `json:"team_id"`
	TeamName    string   `json:"team_name"`
	ScreenType  string   `json:"screen_type"`
	Members     []string `json:"members"`
	StartDate   time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Reason      string   `json:"reason"`
}

// RestrictedInformation 受限信息
type RestrictedInformation struct {
	Categories      []string          `json:"categories"`
	DocumentTypes   []string          `json:"document_types"`
	Communication   []string          `json:"communication"`
	AccessLevel     string            `json:"access_level"`
	ExpiryRules     map[string]string `json:"expiry_rules"`
	ApprovalRequired bool             `json:"approval_required"`
}

// AccessPermissions 访问权限
type AccessPermissions struct {
	ReadAccess     []string `json:"read_access"`
	WriteAccess    []string `json:"write_access"`
	DeleteAccess   []string `json:"delete_access"`
	ShareAccess    []string `json:"share_access"`
	ApprovalMatrix map[string][]string `json:"approval_matrix"`
}

// MonitoringPlan 监控计划
type MonitoringPlan struct {
	Frequency      string            `json:"frequency"`
	Methods        []string          `json:"methods"`
	Triggers       []string          `json:"triggers"`
	ReportTemplate string            `json:"report_template"`
	NotifyPeople   []string          `json:"notify_people"`
	EscalationRules map[string]string `json:"escalation_rules"`
}

// TableName 设置表名
func (EnhancedCase) TableName() string {
	return "cases"
}

func (CaseClientProfile) TableName() string {
	return "case_client_profiles"
}

func (CaseConflictRecord) TableName() string {
	return "case_conflict_records"
}

func (CaseWaiverAssociation) TableName() string {
	return "case_waiver_associations"
}

func (CaseEthicalScreen) TableName() string {
	return "case_ethical_screens"
}

// BeforeCreate GORM钩子
func (e *EnhancedCase) BeforeCreate(tx *gorm.DB) error {
	// 如果有客户档案ID，设置冲突检测状态
	if len(e.ClientProfileIDs) > 0 && e.ConflictDetectionStatus == "" {
		e.ConflictDetectionStatus = ConflictDetectionStatusPending
	}

	// 设置默认风险等级
	if e.RiskLevel == "" {
		e.RiskLevel = RiskLevelNotAssessed
	}

	return nil
}

// BeforeUpdate GORM钩子
func (e *EnhancedCase) BeforeUpdate(tx *gorm.DB) error {
	// 当客户档案发生变化时，重新标记冲突检测状态
	if tx.Statement.Changed("client_profile_ids") {
		e.ConflictDetectionStatus = ConflictDetectionStatusPending
	}

	return nil
}