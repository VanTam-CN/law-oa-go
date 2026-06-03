package models

import (
	"time"

	"gorm.io/gorm"
)

// EntityType 实体类型
type EntityType string

const (
	EntityTypeIndividual   EntityType = "INDIVIDUAL"   // 自然人
	EntityTypeLegalPerson  EntityType = "LEGAL_PERSON" // 法人
	EntityTypeOrganization EntityType = "ORGANIZATION" // 组织机构
)

// EntityStatus 实体状态
type EntityStatus string

const (
	EntityStatusActive    EntityStatus = "ACTIVE"    // 活跃
	EntityStatusInactive  EntityStatus = "INACTIVE"  // 非活跃
	EntityStatusBlacklist EntityStatus = "BLACKLIST" // 黑名单
	EntityStatusMerged    EntityStatus = "MERGED"    // 已合并
)

// IdentityType 证件类型
type IdentityType string

const (
	IdentityTypeIDCard        IdentityType = "ID_CARD"         // 身份证
	IdentityTypePassport      IdentityType = "PASSPORT"        // 护照
	IdentityTypeBusinessLicense IdentityType = "BUSINESS_LICENSE" // 营业执照
	IdentityTypeOrgCode       IdentityType = "ORGANIZATION_CODE" // 组织机构代码
	IdentityTypeSocialCredit  IdentityType = "SOCIAL_CREDIT_CODE" // 统一社会信用代码
	IdentityTypeOther         IdentityType = "OTHER"           // 其他
)

// Gender 性别
type Gender string

const (
	GenderMale   Gender = "MALE"   // 男
	GenderFemale Gender = "FEMALE" // 女
	GenderOther  Gender = "OTHER"  // 其他
)

// Entity 法律实体模型
type Entity struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at" gorm:"index"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 基本信息
	EntityType EntityType `json:"entity_type" gorm:"column:entity_type;type:varchar(20);not null;index:idx_entity_type"`
	Name       string     `json:"name" gorm:"column:name;size:200;not null;index:idx_entity_name"`
	Alias      string     `json:"alias" gorm:"column:alias;size:500"`

	// 身份信息
	IdentityType   IdentityType `json:"identity_type" gorm:"column:identity_type;type:varchar(30);index:idx_identity"`
	IdentityNumber string       `json:"identity_number" gorm:"column:identity_number;size:100;index:idx_identity"`

	// 状态
	Status EntityStatus `json:"status" gorm:"column:status;type:varchar(20);not null;default:'ACTIVE';index:idx_status"`

	// 自然人特有字段
	Gender      Gender     `json:"gender,omitempty" gorm:"column:gender;type:varchar(10)"`
	Nationality string     `json:"nationality,omitempty" gorm:"column:nationality;size:50"`
	BirthDate   *time.Time `json:"birth_date,omitempty" gorm:"column:birth_date"`

	// 法人特有字段
	LegalRepresentative string  `json:"legal_representative,omitempty" gorm:"column:legal_representative;size:100"`
	RegisteredCapital   float64 `json:"registered_capital,omitempty" gorm:"column:registered_capital"`
	EstablishDate       *time.Time `json:"establish_date,omitempty" gorm:"column:establish_date"`
	BusinessScope       string  `json:"business_scope,omitempty" gorm:"column:business_scope;type:text"`

	// 联系信息
	Address       string `json:"address,omitempty" gorm:"column:address;type:text"`
	Phone        string `json:"phone,omitempty" gorm:"column:phone;size:20"`
	Email        string `json:"email,omitempty" gorm:"column:email;size:100"`
	ContactPerson string `json:"contact_person,omitempty" gorm:"column:contact_person;size:100"`

	// 备注信息
	Notes string `json:"notes,omitempty" gorm:"column:notes;type:text"`

	// 关联数据
	Relations       []EntityRelation       `json:"relations,omitempty" gorm:"foreignKey:SourceEntityID"`
	NameHistory     []EntityNameHistory    `json:"name_history,omitempty" gorm:"foreignKey:EntityID"`
	CaseParties     []CaseParty            `json:"case_parties,omitempty" gorm:"foreignKey:EntityID"`
}

// TableName 指定表名
func (Entity) TableName() string {
	return "entities"
}
type RelationType string
// RelationType 关系类型

const (
	RelationTypeParentCompany      RelationType = "PARENT_COMPANY"       // 母公司
	RelationTypeSubsidiary         RelationType = "SUBSIDIARY"          // 子公司
	RelationTypeActualController   RelationType = "ACTUAL_CONTROLLER"   // 实际控制人
	RelationTypeMajorShareholder   RelationType = "MAJOR_SHAREHOLDER"   // 大股东
	RelationTypeNomineeShareholder RelationType = "NOMINEE_SHAREHOLDER" // 代持股东
	RelationTypeBranch             RelationType = "BRANCH"              // 分支机构
	RelationTypeJointVenture       RelationType = "JOINT_VENTURE"       // 合资企业
	RelationTypeRelatedParty       RelationType = "RELATED_PARTY"       // 关联方
	RelationTypeFamilyMember       RelationType = "FAMILY_MEMBER"       // 家庭成员
	RelationTypeSpouse             RelationType = "SPOUSE"              // 配偶
)

// EntityRelation 实体关联关系模型
type EntityRelation struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at" gorm:"index"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联实体
	SourceEntityID uint   `json:"source_entity_id" gorm:"column:source_entity_id;not null;index:idx_source_entity"`
	SourceEntity   Entity `json:"source_entity,omitempty" gorm:"foreignKey:SourceEntityID"`

	TargetEntityID uint   `json:"target_entity_id" gorm:"column:target_entity_id;not null;index:idx_target_entity"`
	TargetEntity   Entity `json:"target_entity,omitempty" gorm:"foreignKey:TargetEntityID"`

	// 关系信息
	RelationType      RelationType `json:"relation_type" gorm:"column:relation_type;type:varchar(30);not null;index:idx_relation_type"`
	ShareholdingRatio float64      `json:"shareholding_ratio,omitempty" gorm:"column:shareholding_ratio"`
	Description       string       `json:"description,omitempty" gorm:"column:description;type:text"`

	// 时间信息
	StartDate *time.Time `json:"start_date,omitempty" gorm:"column:start_date"`
	EndDate   *time.Time `json:"end_date,omitempty" gorm:"column:end_date"`
	IsActive  bool       `json:"is_active" gorm:"column:is_active;default:true"`

	// 来源信息
	DataSource string `json:"data_source,omitempty" gorm:"column:data_source;size:100"`
}

// TableName 指定表名
func (EntityRelation) TableName() string {
	return "entity_relations"
}

// EntityNameHistory 实体名称变更记录
type EntityNameHistory struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联实体
	EntityID uint   `json:"entity_id" gorm:"column:entity_id;not null;index:idx_entity"`
	Entity   Entity `json:"entity,omitempty" gorm:"foreignKey:EntityID"`

	// 名称变更信息
	OldName     string    `json:"old_name" gorm:"column:old_name;size:200;not null"`
	NewName     string    `json:"new_name" gorm:"column:new_name;size:200;not null"`
	ChangeDate  time.Time `json:"change_date" gorm:"column:change_date;not null"`
	ChangeReason string   `json:"change_reason,omitempty" gorm:"column:change_reason;size:500"`
}

// TableName 指定表名
func (EntityNameHistory) TableName() string {
	return "entity_name_history"
}

// PartyRole 当事人角色
type PartyRole string

const (
	PartyRolePlaintiff       PartyRole = "PLAINTIFF"        // 原告
	PartyRoleDefendant       PartyRole = "DEFENDANT"        // 被告
	PartyRoleThirdParty      PartyRole = "THIRD_PARTY"      // 第三人
	PartyRoleInterestedParty PartyRole = "INTERESTED_PARTY" // 利害关系人
	PartyRoleWitness         PartyRole = "WITNESS"          // 证人
	PartyRoleExpert          PartyRole = "EXPERT"           // 鉴定人
)

// PartyType 当事人类型
type PartyType string

const (
	PartyTypeClient       PartyType = "CLIENT"        // 委托人
	PartyTypeOpposing     PartyType = "OPPOSING"      // 对手方
	PartyTypeCoDefendant  PartyType = "CO_DEFENDANT"  // 共同被告
	PartyTypeCoPlaintiff  PartyType = "CO_PLAINTIFF"  // 共同原告
)

// CaseParty 案件当事人关系模型
type CaseParty struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at" gorm:"index"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联案件和实体
	CaseID  uint   `json:"case_id" gorm:"column:case_id;not null;index:idx_case_party"`
	Case    Case   `json:"case,omitempty" gorm:"foreignKey:CaseID"`
	EntityID uint   `json:"entity_id" gorm:"column:entity_id;not null;index:idx_entity_party"`
	Entity   Entity `json:"entity,omitempty" gorm:"foreignKey:EntityID"`

	// 角色信息
	Role      PartyRole `json:"role" gorm:"column:role;type:varchar(30);not null;index:idx_role"`
	PartyType PartyType `json:"party_type" gorm:"column:party_type;type:varchar(20);not null"`

	// 详情信息
	Description string `json:"description,omitempty" gorm:"column:description;type:text"`

	// 顺序
	DisplayOrder int `json:"display_order" gorm:"column:display_order;default:0"`
}

// TableName 指定表名
func (CaseParty) TableName() string {
	return "case_parties"
}


// CheckStatus 冲突审查状态
type CheckStatus string

const (
	CheckStatusPending              CheckStatus = "PENDING"                // 待审查
	CheckStatusProcessing           CheckStatus = "PROCESSING"             // 审查中
	CheckStatusInProgress           CheckStatus = "IN_PROGRESS"            // 审查中（别名）
	CheckStatusCompleted            CheckStatus = "COMPLETED"              // 已完成（无冲突）
	CheckStatusCompletedWithConflict CheckStatus = "COMPLETED_WITH_CONFLICT" // 已完成（有冲突）
	CheckStatusApproved             CheckStatus = "APPROVED"               // 已批准
	CheckStatusRejected             CheckStatus = "REJECTED"               // 已拒绝
	CheckStatusFailed               CheckStatus = "FAILED"                 // 失败
)

// CheckResult 冲突审查结果
type CheckResult struct {
	HasConflict    bool      `json:"has_conflict"`
	TotalConflicts int       `json:"total_conflicts"`
	CompletedAt    time.Time `json:"completed_at"`
}

// ConflictCheck 冲突检查记录
type ConflictCheck struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联案件
	CaseID uint `json:"case_id" gorm:"column:case_id;not null;index:idx_case_check"`
	Case   Case `json:"case,omitempty" gorm:"foreignKey:CaseID"`

	// 检查信息
	Status       string     `json:"status" gorm:"column:status;type:varchar(50);not null;default:'PENDING';index:idx_check_status"`
	RequestedBy  uint       `json:"requested_by" gorm:"column:requested_by;not null;index:idx_requested_by"`
	RequestedAt  time.Time  `json:"requested_at" gorm:"column:requested_at;not null"`
	CheckedBy    *uint      `json:"checked_by,omitempty" gorm:"column:checked_by"`
	CheckedAt    *time.Time `json:"checked_at,omitempty" gorm:"column:checked_at"`

	// 结果摘要
	Result        *CheckResult `json:"result,omitempty" gorm:"column:result;serializer:json"`
	ResultSummary string       `json:"result_summary,omitempty" gorm:"column:result_summary;type:text"`
	TotalConflicts int          `json:"total_conflicts" gorm:"column:total_conflicts;default:0"`
	CriticalCount  int          `json:"critical_count" gorm:"column:critical_count;default:0"`
	HighCount      int          `json:"high_count" gorm:"column:high_count;default:0"`
	MediumCount    int          `json:"medium_count" gorm:"column:medium_count;default:0"`
	LowCount       int          `json:"low_count" gorm:"column:low_count;default:0"`

	// 检查参数
	CheckParams JSON `json:"check_params,omitempty" gorm:"column:check_params;type:json"`

	// 报告信息
	ReportPath        string     `json:"report_path,omitempty" gorm:"column:report_path;size:500"`
	ReportGeneratedAt *time.Time `json:"report_generated_at,omitempty" gorm:"column:report_generated_at"`

	// 关联详情
	ConflictDetails []ConflictDetail `json:"conflict_details,omitempty" gorm:"foreignKey:ConflictCheckID"`
}

// TableName 指定表名
func (ConflictCheck) TableName() string {
	return "conflict_checks"
}

// ConflictDetail 冲突详情模型
// 用于记录利益冲突审查中检测到的具体冲突点
type ConflictDetail struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at" gorm:"index"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联审查记录
	ConflictCheckID uint         `json:"conflict_check_id" gorm:"column:conflict_check_id;not null;index:idx_conflict_check"` // 冲突审查ID
	ConflictCheck   ConflictCheck `json:"conflict_check,omitempty" gorm:"foreignKey:ConflictCheckID"`

	// 匹配的实体和案件
	MatchedEntityID uint   `json:"matched_entity_id" gorm:"column:matched_entity_id;not null;index:idx_matched_entity"` // 匹配的实体ID
	MatchedEntity   Entity `json:"matched_entity,omitempty" gorm:"foreignKey:MatchedEntityID"`
	MatchedCaseID   *uint  `json:"matched_case_id,omitempty" gorm:"column:matched_case_id;index:idx_matched_case"`   // 匹配的案件ID
	MatchedCase     Case   `json:"matched_case,omitempty" gorm:"foreignKey:MatchedCaseID"`

	// 冲突信息 - 使用字符串类型避免与其他文件的类型定义冲突
	ConflictType string `json:"conflict_type" gorm:"column:conflict_type;type:varchar(50);not null;index:idx_conflict_type"` // 冲突类型: IDENTITY_MATCH, NAME_SIMILAR, RELATIONSHIP, CASE_ASSOCIATION
	RiskLevel    string `json:"risk_level" gorm:"column:risk_level;type:varchar(20);not null;index:idx_risk_level"`         // 风险等级: LOW, MEDIUM, HIGH, CRITICAL

	// 冲突描述
	Description string `json:"description" gorm:"column:description;type:text;not null"` // 冲突描述
	Evidence     string `json:"evidence,omitempty" gorm:"column:evidence;type:text"`       // 证据说明

	// 处理建议
	Recommendation string `json:"recommendation,omitempty" gorm:"column:recommendation;type:text"` // 处理建议

	// 豁免信息
	IsWaived    bool       `json:"is_waived" gorm:"column:is_waived;default:false"`                 // 是否已豁免
	WaivedBy    *uint      `json:"waived_by,omitempty" gorm:"column:waived_by"`                      // 豁免批准人ID
	WaivedAt    *time.Time `json:"waived_at,omitempty" gorm:"column:waived_at"`                      // 豁免时间
	WaiveReason string     `json:"waive_reason,omitempty" gorm:"column:waive_reason;size:500"`        // 豁免原因

	// 匹配原因（供服务层使用）
	MatchReason string `json:"match_reason,omitempty" gorm:"column:match_reason;size:500"` // 匹配原因说明
}

// TableName 指定表名
func (ConflictDetail) TableName() string {
	return "conflict_details"
}

