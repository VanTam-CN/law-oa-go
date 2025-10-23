package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// JSONB 自定义JSONB类型
type JSONB json.RawMessage

// Value 实现driver.Valuer接口
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return string(j), nil
}

// Scan 实现sql.Scanner接口
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = v
	case string:
		*j = JSONB(v)
	default:
		return nil
	}
	return nil
}

// MarshalJSON 实现json.Marshaler接口
func (j JSONB) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return []byte(j), nil
}

// UnmarshalJSON 实现json.Unmarshaler接口
func (j *JSONB) UnmarshalJSON(data []byte) error {
	if data == nil {
		*j = nil
		return nil
	}
	*j = JSONB(data)
	return nil
}

// ===================================
// 1. 用户相关模型（完整版）
// ===================================

// User 完整用户模型
type UserComplete struct {
	ID           uint           `json:"id" gorm:"primarykey"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// 基本信息
	Username     string  `json:"username" gorm:"size:50;not null;uniqueIndex"`
	Name         string  `json:"name" gorm:"size:100"`
	RealName     string  `json:"real_name" gorm:"size:50"`
	Email        string  `json:"email" gorm:"size:100;not null;uniqueIndex"`
	Password     string  `json:"-" gorm:"size:255;not null"`
	Phone        string  `json:"phone" gorm:"size:20"`
	Avatar       string  `json:"avatar" gorm:"size:255"`

	// 角色权限
	Role         string  `json:"role" gorm:"size:50;not null;default:'user'"`
	RoleID       *uint   `json:"role_id" gorm:"column:role_id;index"`
	DepartmentID *uint   `json:"department_id" gorm:"column:department_id;index"`

	// 状态信息
	Status       string `json:"status" gorm:"size:20;default:'active'"`

	// 登录信息
	LastLoginAt  *time.Time `json:"last_login_at"`
	LastLoginIP  string      `json:"last_login_ip" gorm:"size:45"`

	// 其他信息
	Remark       string `json:"remark" gorm:"type:text"`

	// 关联
	Department   *Department    `json:"department,omitempty" gorm:"foreignKey:DepartmentID"`
	UserRole     []UserRole      `json:"user_roles,omitempty" gorm:"foreignKey:UserID"`
}

// Department 部门模型
type Department struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	Name        string  `json:"name" gorm:"size:100;not null"`
	Code        string  `json:"code" gorm:"size:50;not null;uniqueIndex"`
	ParentID    uint    `json:"parent_id" gorm:"column:parent_id;default:0"`
	LeaderID    *uint   `json:"leader_id" gorm:"column:leader_id;index"`
	Description string  `json:"description" gorm:"type:text"`
	SortOrder   int     `json:"sort_order" gorm:"column:sort_order;default:0"`
	Status      int     `json:"status" gorm:"default:1"`

	// 关联
	Leader      *User      `json:"leader,omitempty" gorm:"foreignKey:LeaderID"`
	Users       []User     `json:"users,omitempty" gorm:"foreignKey:DepartmentID"`
}

func (Department) TableName() string {
	return "departments"
}

func (UserComplete) TableName() string {
	return "users"
}

// ===================================
// 2. 律师相关模型
// ===================================

// Lawyer 律师模型
type Lawyer struct {
	ID           uint           `json:"id" gorm:"primarykey"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	LawyerName   string  `json:"lawyer_name" gorm:"column:lawyer_name;size:50;not null"`
	Phone        string  `json:"phone" gorm:"size:20"`
	Email        string  `json:"email" gorm:"size:100"`
	LicenseNo    string  `json:"license_no" gorm:"column:license_no;size:50;uniqueIndex"`
	Position     string  `json:"position" gorm:"size:50"`
	Department   string  `json:"department" gorm:"size:100"`
	Specialty    string  `json:"specialty" gorm:"type:text"`
	Status       string  `json:"status" gorm:"size:20;default:'active'"`
	Remark       string  `json:"remark" gorm:"type:text"`

	// 关联
	CasesAsMain  []Case      `json:"cases_as_main,omitempty" gorm:"foreignKey:LawyerID"`
	CasesAsAssist []Case      `json:"cases_as_assist,omitempty" gorm:"foreignKey:AssistingLawyerID"`
}

func (Lawyer) TableName() string {
	return "lawyers"
}

// ===================================
// 3. 客户相关模型（完整版）
// ===================================

// ClientComplete 完整客户模型
type ClientComplete struct {
	ID           uint           `json:"id" gorm:"primarykey"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// 基本信息
	ClientName   string  `json:"client_name" gorm:"column:client_name;size:100"`
	Name         string  `json:"name" gorm:"column:name;size:100;not null"`
	Type         string  `json:"type" gorm:"column:type;size:20;not null;default:'个人'"`
	Email        string  `json:"email" gorm:"size:100;index"`
	Phone        string  `json:"phone" gorm:"size:20;index"`
	Address      string  `json:"address" gorm:"type:text"`

	// 企业信息
	Company      string  `json:"company" gorm:"size:100"`
	IDCard       string  `json:"id_card" gorm:"column:id_card;size:18"`

	// 行业信息
	Industry     string  `json:"industry" gorm:"column:industry;size:50"`
	ContactPerson string `json:"contact_person" gorm:"column:contact_person;size:50"`
	ContactPhone string `json:"contact_phone" gorm:"column:contact_phone;size:20"`

	// 客户管理
	Source       string `json:"source" gorm:"column:source;size:50"`
	LawyerID     *uint  `json:"lawyer_id" gorm:"column:lawyer_id;index"`
	Notes        string `json:"notes" gorm:"column:notes;type:text"`
	Remark       string `json:"remark" gorm:"type:text"`
	Status       string `json:"status" gorm:"size:20;default:'active'"`

	// 关联
	Lawyer       *Lawyer     `json:"lawyer,omitempty" gorm:"foreignKey:LawyerID"`
	Cases        []Case       `json:"cases,omitempty" gorm:"foreignKey:ClientID"`
}

func (ClientComplete) TableName() string {
	return "clients"
}

// ===================================
// 4. 案件相关模型（完整版）
// ===================================

// CaseComplete 完整案件模型
type CaseComplete struct {
	ID                    uint           `json:"id" gorm:"primarykey"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `json:"-" gorm:"index"`

	// 基本信息
	CaseNo                string  `json:"case_no" gorm:"column:case_no;size:50;uniqueIndex"`
	CaseName              string  `json:"case_name" gorm:"column:case_name;size:200"`
	Title                 string  `json:"title" gorm:"size:200;not null"`
	Description           string  `json:"description" gorm:"type:text"`

	// 关联信息
	ClientID              uint    `json:"client_id" gorm:"not null;index"`
	ClientName             string  `json:"client_name" gorm:"-"` // 临时字段，不映射到数据库
	LawyerID              uint    `json:"lawyer_id" gorm:"not null;index"`
	AssistingLawyerID     *uint   `json:"assisting_lawyer_id" gorm:"column:assisting_lawyer_id;index"`

	// 案件信息
	CaseType              string  `json:"case_type" gorm:"size:50;not null"`
	Priority              string  `json:"priority" gorm:"size:20;default:'medium'"`
	Status                string  `json:"status" gorm:"size:20;default:'pending'"`
	ProjectCode           string  `json:"project_code" gorm:"column:project_code;size:50"`
	ProjectType           string  `json:"project_type" gorm:"column:project_type;size:50"`
	TeamMembers           string  `json:"team_members" gorm:"type:text"`

	// 金额信息
	ContractAmount        *float64 `json:"contract_amount" gorm:"column:contract_amount;type:decimal(12,2)"`

	// 时间信息
	StartDate             *time.Time `json:"start_date"`
	EndDate               *time.Time `json:"end_date"`

	// 当事人信息
	PrincipalInfo         string  `json:"principal_info" gorm:"type:text"`
	OpponentInfo         string  `json:"opponent_info" gorm:"type:text"`
	CauseOfAction        string  `json:"cause_of_action" gorm:"column:cause_of_action;type:text"`

	// 费用信息
	BillingMethod         string  `json:"billing_method" gorm:"column:billing_method;size:50"`

	// 风险管理
	ConflictCheckStatus   string  `json:"conflict_check_status" gorm:"column:conflict_check_status;size:20;default:'pending'"`
	IsMajorRisk          bool    `json:"is_major_risk" gorm:"column:is_major_risk;default:false"`
	IsMassCase           bool    `json:"is_mass_case" gorm:"column:is_mass_case;default:false"`
	IsSensitiveCase      bool    `json:"is_sensitive_case" gorm:"column:is_sensitive_case;default:false"`

	// 文档信息
	ContractDocument     string  `json:"contract_document" gorm:"column:contract_document;size:500"`
	LegalLetterDocument  string  `json:"legal_letter_document" gorm:"column:legal_letter_document;size:500"`
	OtherDocuments      string  `json:"other_documents" gorm:"type:text"`

	// 其他信息
	Remark               string  `json:"remark" gorm:"type:text"`

	// 关联
	Client               *ClientComplete `json:"client,omitempty" gorm:"foreignKey:ClientID"`
	Lawyer               *UserComplete   `json:"lawyer,omitempty" gorm:"foreignKey:LawyerID"`
	AssistingLawyer      *UserComplete   `json:"assisting_lawyer,omitempty" gorm:"foreignKey:AssistingLawyerID"`
	CaseProgress         []CaseProgress  `json:"case_progress,omitempty" gorm:"foreignKey:CaseID"`
	CaseDocuments        []CaseDocument  `json:"case_documents,omitempty" gorm:"foreignKey:CaseID"`
	FinancialRecords     []FinancialRecord `json:"financial_records,omitempty" gorm:"foreignKey:CaseID"`
	Schedules            []Schedule      `json:"schedules,omitempty" gorm:"foreignKey:CaseID"`
}

// CaseProgress 案件进度模型
type CaseProgress struct {
	ID            uint           `json:"id" gorm:"primarykey"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`

	CaseID        uint           `json:"case_id" gorm:"not null;index"`
	Stage         string         `json:"stage" gorm:"size:50;not null"`
	Title         string         `json:"title" gorm:"size:200;not null"`
	Description   string         `json:"description" gorm:"type:text"`
	Status        string         `json:"status" gorm:"size:20;default:'pending'"`
	DueDate       *time.Time     `json:"due_date"`
	CompletedAt   *time.Time     `json:"completed_at"`
	CreatedBy     uint           `json:"created_by" gorm:"not null"`

	// 关联
	Case          *CaseComplete   `json:"case,omitempty" gorm:"foreignKey:CaseID"`
	Creator       *UserComplete    `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
}

// CaseDocument 案件文档模型
type CaseDocument struct {
	ID           uint           `json:"id" gorm:"primarykey"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	CaseID       uint    `json:"case_id" gorm:"not null;index"`
	Name         string  `json:"name" gorm:"size:200;not null"`
	Type         string  `json:"type" gorm:"size:50;not null"`
	FilePath     string  `json:"file_path" gorm:"size:500;not null"`
	FileSize     int64   `json:"file_size"`
	MimeType     string  `json:"mime_type" gorm:"size:100"`
	Description  string  `json:"description" gorm:"type:text"`
	UploadedBy   uint    `json:"uploaded_by" gorm:"not null;index"`

	// 关联
	Case         *CaseComplete   `json:"case,omitempty" gorm:"foreignKey:CaseID"`
	Uploader     *UserComplete   `json:"uploader,omitempty" gorm:"foreignKey:UploadedBy"`
}

func (CaseComplete) TableName() string {
	return "cases"
}

func (CaseProgress) TableName() string {
	return "case_progress"
}

func (CaseDocument) TableName() string {
	return "case_documents"
}

// ===================================
// 5. 利益冲突检测系统模型（完整版）
// ===================================

// LawEntity 法律实体模型
type LawEntity struct {
	ID            uint           `json:"id" gorm:"primarykey"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	EntityName    string  `json:"entity_name" gorm:"column:entity_name;size:200;not null"`
	EntityType    string  `json:"entity_type" gorm:"column:entity_type;size:50"`
	EntitySubtype string  `json:"entity_subtype" gorm:"column:entity_subtype;size:50"`
	IDCard        string  `json:"id_card" gorm:"size:20"`
	LicenseNo     string  `json:"license_no" gorm:"column:license_no;size:50"`
	Address       string  `json:"address" gorm:"type:text"`
	ContactInfo   string  `json:"contact_info" gorm:"type:text"`
	RiskLevel     string  `json:"risk_level" gorm:"column:risk_level;size:20;default:'low'"`
	Status        string  `json:"status" gorm:"size:20;default:'active'"`
	Remark        string  `json:"remark" gorm:"type:text"`

	// 关联
	Aliases       []LawEntityAlias   `json:"aliases,omitempty" gorm:"foreignKey:EntityID"`
	Relations     []LawEntityRelation `json:"relations,omitempty" gorm:"foreignKey:SourceEntityID"`
}

// LawEntityAlias 法律实体别名模型
type LawEntityAlias struct {
	ID            uint           `json:"id" gorm:"primarykey"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	EntityID      uint    `json:"entity_id" gorm:"not null;index"`
	AliasName     string  `json:"alias_name" gorm:"column:alias_name;size:200;not null"`
	AliasType     string  `json:"alias_type" gorm:"column:alias_type;size:50"`
	Status        string  `json:"status" gorm:"size:20;default:'active'"`
	Remark        string  `json:"remark" gorm:"type:text"`

	// 关联
	Entity        *LawEntity `json:"entity,omitempty" gorm:"foreignKey:EntityID"`
}

// LawEntityRelation 法律实体关系模型
type LawEntityRelation struct {
	ID            uint           `json:"id" gorm:"primarykey"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	SourceEntityID uint    `json:"source_entity_id" gorm:"not null;index"`
	TargetEntityID uint    `json:"target_entity_id" gorm:"not null;index"`
	RelationType  string  `json:"relation_type" gorm:"column:relation_type;size:50;not null"`
	RelationDesc  string  `json:"relation_desc" gorm:"column:relation_desc;type:text"`
	StartDate     *time.Time `json:"start_date"`
	EndDate       *time.Time `json:"end_date"`
	Status        string  `json:"status" gorm:"size:20;default:'active'"`
	Remark        string  `json:"remark" gorm:"type:text"`

	// 关联
	SourceEntity  *LawEntity `json:"source_entity,omitempty" gorm:"foreignKey:SourceEntityID"`
	TargetEntity  *LawEntity `json:"target_entity,omitempty" gorm:"foreignKey:TargetEntityID"`
}

func (LawEntity) TableName() string {
	return "law_entities"
}

func (LawEntityAlias) TableName() string {
	return "law_entity_aliases"
}

func (LawEntityRelation) TableName() string {
	return "law_entity_relations"
}

// ConflictCheckRecord 冲突检查记录模型（完整版）
type ConflictCheckRecordComplete struct {
	ID              string    `json:"check_id" gorm:"primarykey;column:check_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	ClientID        string    `json:"client_id" gorm:"not null;index"`
	ClientName      string    `json:"client_name" gorm:"column:client_name;size:255;not null"`
	CaseName        string    `json:"case_name" gorm:"column:case_name;size:255;not null"`
	CaseType        string    `json:"case_type" gorm:"column:case_type;size:100;not null"`
	CheckStatus     string    `json:"check_status" gorm:"column:check_status;size:20;default:'PROCESSING'"`
	HasConflict     bool      `json:"has_conflict" gorm:"default:false"`
	RiskLevel       string    `json:"risk_level" gorm:"column:risk_level;size:20;default:'LOW'"`

	// 原有字段
	SearchParameters JSONB     `json:"search_parameters" gorm:"type:jsonb"`
	CheckResult     JSONB     `json:"check_result" gorm:"type:jsonb"`
	UserID          *uint     `json:"user_id" gorm:"column:user_id;index"`
	CheckTime       time.Time `json:"check_time" gorm:"column:check_time"`
	Duration        *int64    `json:"duration"`

	// MySQL原字段
	CaseID          *uint     `json:"case_id" gorm:"index"`
	TargetID        *uint     `json:"target_id" gorm:"column:target_id;index"`
	TargetName      string    `json:"target_name" gorm:"column:target_name;size:200"`
	TargetType      string    `json:"target_type" gorm:"column:target_type;size:50"`
	ConflictDesc    string    `json:"conflict_desc" gorm:"column:conflict_desc;type:text;not null"`
	RelatedCaseID   *uint     `json:"related_case_id" gorm:"column:related_case_id;index"`
	Recommendation   string    `json:"recommendation" gorm:"type:text"`
	CheckedBy       string    `json:"checked_by" gorm:"column:checked_by;size:50"`
	CheckedAt       *time.Time `json:"checked_at" gorm:"column:checked_at"`
	ResolvedBy      string    `json:"resolved_by" gorm:"column:resolved_by;size:50"`
	ResolvedAt      *time.Time `json:"resolved_at" gorm:"column:resolved_at"`
	Resolution      string    `json:"resolution" gorm:"type:text"`
	Remark          string    `json:"remark" gorm:"type:text"`

	// 关联
	User            *UserComplete `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Case            *CaseComplete `json:"case,omitempty" gorm:"foreignKey:CaseID"`
	RelatedCase     *CaseComplete `json:"related_case,omitempty" gorm:"foreignKey:RelatedCaseID"`
}

func (ConflictCheckRecordComplete) TableName() string {
	return "conflict_check_records"
}

// ===================================
// 6. 文档管理系统模型（完整版）
// ===================================

// CaseDocument 案件文档模型

// DocumentVersion 文档版本模型
type DocumentVersion struct {
	ID           uint           `json:"id" gorm:"primarykey"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	DocumentID   uint      `json:"document_id" gorm:"not null;index"`
	VersionNo    int       `json:"version_no" gorm:"column:version_no;not null"`
	FilePath     string    `json:"file_path" gorm:"column:file_path;size:500;not null"`
	FileHash     string    `json:"file_hash" gorm:"column:file_hash;size:64;not null"`
	FileSize     int64     `json:"file_size" gorm:"column:file_size;not null"`
	UploaderID   uint      `json:"uploader_id" gorm:"not null"`
	ChangeLog    string    `json:"change_log" gorm:"type:text"`
	UploadTime   time.Time `json:"upload_time" gorm:"column:upload_time;not null"`
	IsCurrent    bool      `json:"is_current" gorm:"default:false"`
	Remark       string    `json:"remark" gorm:"type:text"`

	// 关联
	Document     *Document     `json:"document,omitempty" gorm:"foreignKey:DocumentID"`
	Uploader     *UserComplete `json:"uploader,omitempty" gorm:"foreignKey:UploaderID"`
}

// DocumentPermission 文档权限模型
type DocumentPermission struct {
	ID           uint           `json:"id" gorm:"primarykey"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`

	DocumentID   uint        `json:"document_id" gorm:"not null;index"`
	UserID       uint        `json:"user_id" gorm:"not null;index"`
	PermissionType string     `json:"permission_type" gorm:"column:permission_type;size:50;not null"`
	CreatedBy    uint        `json:"created_by" gorm:"not null"`
	ExpiresAt    *time.Time  `json:"expires_at"`
	Status       string      `json:"status" gorm:"size:20;default:'active'"`
	Remark       string      `json:"remark" gorm:"type:text"`

	// 关联
	Document     *Document     `json:"document,omitempty" gorm:"foreignKey:DocumentID"`
	User         *UserComplete `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Creator      *UserComplete `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
}

// DocumentCategory 文档分类模型
type DocumentCategory struct {
	ID             uint           `json:"id" gorm:"primarykey"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	CategoryName   string  `json:"category_name" gorm:"column:category_name;size:100;uniqueIndex;not null"`
	CategoryKey    string  `json:"category_key" gorm:"column:category_key;size:100;uniqueIndex;not null"`
	ParentID       *uint   `json:"parent_id" gorm:"index"`
	Description    string  `json:"description" gorm:"type:text"`
	Sort           int     `json:"sort" gorm:"default:0"`
	Status         string  `json:"status" gorm:"size:20;default:'active'"`
	Icon           string  `json:"icon" gorm:"size:100"`
	Color          string  `json:"color" gorm:"size:20"`
	DocumentCount  int     `json:"document_count" gorm:"column:document_count;default:0"`
	Remark         string  `json:"remark" gorm:"type:text"`

	// 关联
	Parent         *DocumentCategory   `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children       []DocumentCategory  `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}

func (DocumentCategory) TableName() string {
	return "document_categories"
}

func (DocumentVersion) TableName() string {
	return "document_versions"
}

func (DocumentPermission) TableName() string {
	return "document_permissions"
}

// ===================================
// 7. 系统管理模型
// ===================================

// SystemConfig 系统配置模型
type SystemConfig struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	ConfigKey   string  `json:"config_key" gorm:"column:config_key;size:100;uniqueIndex;not null"`
	ConfigValue string  `json:"config_value" gorm:"type:text"`
	ConfigType  string  `json:"config_type" gorm:"column:config_type;size:20;default:'string'"`
	Description string  `json:"description" gorm:"type:text"`
	IsSystem    bool    `json:"is_system" gorm:"column:is_system;default:false"`
	Sort        int     `json:"sort" gorm:"default:0"`
	Status      string  `json:"status" gorm:"size:20;default:'active'"`
}

// OperationLog 操作日志模型
type OperationLog struct {
	ID            uint      `json:"id" gorm:"primarykey"`
	CreatedAt     time.Time `json:"created_at"`

	UserID        *uint     `json:"user_id" gorm:"index"`
	Username      string    `json:"username" gorm:"size:50"`
	Operation     string    `json:"operation" gorm:"size:100;not null"`
	Method        string    `json:"method" gorm:"size:20;not null"`
	Path          string    `json:"path" gorm:"size:255;not null"`
	Params        string    `json:"params" gorm:"type:text"`
	IP            string    `json:"ip" gorm:"size:45"`
	UserAgent     string    `json:"user_agent" gorm:"type:text"`
	Status        int       `json:"status" gorm:"default:200"`
	ErrorMessage  string    `json:"error_message" gorm:"column:error_message;type:text"`
	ExecutionTime *int64    `json:"execution_time"`

	// 关联
	User          *UserComplete `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (SystemConfig) TableName() string {
	return "system_configs"
}

func (OperationLog) TableName() string {
	return "operation_logs"
}

// ===================================
// 8. 财务管理模型
// ===================================

// FinancialRecord 财务记录模型
type FinancialRecord struct {
	ID               uint           `json:"id" gorm:"primarykey"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`

	CaseID           *uint          `json:"case_id" gorm:"index"`
	ClientID         *uint          `json:"client_id" gorm:"index"`
	Type             string         `json:"type" gorm:"size:50;not null"` // income, expense
	Category         string         `json:"category" gorm:"size:50;not null"`
	Amount           float64        `json:"amount" gorm:"type:decimal(15,2);not null"`
	Currency         string         `json:"currency" gorm:"size:10;default:'CNY'"`
	Description      string         `json:"description" gorm:"type:text"`
	TransactionDate  time.Time       `json:"transaction_date" gorm:"column:transaction_date;not null"`
	Status           string         `json:"status" gorm:"size:20;default:'pending'"`
	PaymentMethod    string         `json:"payment_method" gorm:"column:payment_method;size:50"`
	InvoiceNumber    string         `json:"invoice_number" gorm:"column:invoice_number;size:100"`
	CreatedBy        uint           `json:"created_by" gorm:"not null"`

	// 关联
	Case             *CaseComplete  `json:"case,omitempty" gorm:"foreignKey:CaseID"`
	Client           *ClientComplete `json:"client,omitempty" gorm:"foreignKey:ClientID"`
	Creator          *UserComplete   `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
}

func (FinancialRecord) TableName() string {
	return "financial_records"
}

// ===================================
// 9. 消息通知模型
// ===================================

// Notification 消息通知模型
type Notification struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time      `json:"created_at"`

	UserID      uint      `json:"user_id" gorm:"not null;index"`
	Type        string    `json:"type" gorm:"size:50;not null"` // system, case, client, finance
	Title       string    `json:"title" gorm:"size:200;not null"`
	Content     string    `json:"content" gorm:"type:text;not null"`
	RelatedID   *uint     `json:"related_id"`
	RelatedType string    `json:"related_type" gorm:"column:related_type;size:50"`
	IsRead      bool      `json:"is_read" gorm:"default:false"`
	Priority    string    `json:"priority" gorm:"size:20;default:'normal'"`
	ExpiresAt   *time.Time `json:"expires_at"`

	// 关联
	User        *UserComplete `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (Notification) TableName() string {
	return "notifications"
}

// ===================================
// 10. 日程安排模型
// ===================================

// Schedule 日程安排模型
type Schedule struct {
	ID           uint           `json:"id" gorm:"primarykey"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`

	UserID       uint        `json:"user_id" gorm:"not null;index"`
	Title        string      `json:"title" gorm:"size:200;not null"`
	Description  string      `json:"description" gorm:"type:text"`
	StartTime    time.Time   `json:"start_time" gorm:"column:start_time;not null"`
	EndTime      time.Time   `json:"end_time" gorm:"column:end_time;not null"`
	Type         string      `json:"type" gorm:"size:50;not null"` // meeting, hearing, deadline, task
	RelatedID    *uint       `json:"related_id"`
	RelatedType  string      `json:"related_type" gorm:"column:related_type;size:50"`
	Location     string      `json:"location" gorm:"size:200"`
	Participants JSONB       `json:"participants" gorm:"type:jsonb"`
	ReminderTime *time.Time  `json:"reminder_time" gorm:"column:reminder_time"`
	IsAllDay     bool        `json:"is_all_day" gorm:"column:is_all_day;default:false"`
	Status       string      `json:"status" gorm:"size:20;default:'scheduled'"`

	// 关联
	User         *UserComplete `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Case         *CaseComplete  `json:"case,omitempty" gorm:"foreignKey:RelatedID;where:related_type='case'"`
}

func (Schedule) TableName() string {
	return "schedules"
}

// ===================================
// 11. 用户行为分析模型（简化版）
// ===================================

// UserSession 用户会话模型
