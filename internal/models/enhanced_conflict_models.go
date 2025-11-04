package models

import (
	"time"

	"gorm.io/gorm"
)

// 基础审计模型
type BaseModel struct {
	ID        string         `gorm:"type:varchar(36);primaryKey;comment:主键ID" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;comment:软删除时间" json:"deleted_at,omitempty"`
}

// JSON字段类型已在json_types.go中定义，直接使用

// 1. 增强客户管理模型

// 客户档案
type ClientProfile struct {
	BaseModel

	// 基本信息
	ClientNumber         string    `gorm:"type:varchar(50);uniqueIndex;comment:客户编号" json:"client_number"`
	ClientType           string    `gorm:"type:enum('INDIVIDUAL','CORPORATE','GOVERNMENT','NON_PROFIT','TRUST');comment:客户类型" json:"client_type"`
	ClientCategory       string    `gorm:"type:varchar(100);comment:客户分类" json:"client_category"`
	ClientStatus         string    `gorm:"type:enum('PROSPECTIVE','ACTIVE','DORMANT','FORMER','BLACKLISTED');default:'PROSPECTIVE';comment:客户状态" json:"client_status"`

	// 个人客户信息
	IndividualInfo       *IndividualClientInfo `gorm:"type:json;comment:个人客户信息" json:"individual_info,omitempty"`

	// 企业客户信息
	CorporateInfo        *CorporateClientInfo  `gorm:"type:json;comment:企业客户信息" json:"corporate_info,omitempty"`

	// 联系信息
	PrimaryContact       *ContactInfo          `gorm:"type:json;comment:主要联系人信息" json:"primary_contact"`
	BillingContact       *ContactInfo          `gorm:"type:json;comment:账单联系人信息" json:"billing_contact,omitempty"`

	// 客户关系
	RelatedClients       []ClientRelationship  `gorm:"foreignKey:ClientID" json:"related_clients,omitempty"`

	// 名称变体和别名
	NameVariants         []ClientNameVariant   `gorm:"foreignKey:ClientID" json:"name_variants,omitempty"`

	// 行业分类
	IndustryClassifications []IndustryClassification `gorm:"foreignKey:ClientID" json:"industry_classifications,omitempty"`

	// 内部信息
	AssignedPartner      string    `gorm:"type:varchar(36);comment:主管合伙人" json:"assigned_partner"`
	AssignedAttorney     string    `gorm:"type:varchar(36);comment:主管律师" json:"assigned_attorney"`
	BillingDepartment    string    `gorm:"type:varchar(100);comment:计费部门" json:"billing_department"`
	OfficeLocation       string    `gorm:"type:varchar(100);comment:办公地点" json:"office_location"`

	// 元数据
	DataQualityScore     float64   `gorm:"type:decimal(5,2);default:0.00;comment:数据质量评分" json:"data_quality_score"`
	LastVerifiedDate     *time.Time `gorm:"comment:最后验证日期" json:"last_verified_date,omitempty"`
	Source               string    `gorm:"type:varchar(100);comment:数据来源" json:"source"`

	// 风险和合规
	RiskProfile          *ClientRiskProfile     `gorm:"foreignKey:ClientID" json:"risk_profile,omitempty"`
	ComplianceStatus     string                 `gorm:"type:enum('COMPLIANT','UNDER_REVIEW','FLAGGED','SANCTIONED');default:'COMPLIANT';comment:合规状态" json:"compliance_status"`
	DueDiligenceRequired  bool                   `gorm:"default:false;comment:需要尽职调查" json:"due_diligence_required"`
}

// 个人客户信息
type IndividualClientInfo struct {
	Title               string    `json:"title"`
	FirstName           string    `json:"first_name"`
	LastName            string    `json:"last_name"`
	MiddleName          *string   `json:"middle_name,omitempty"`
	MaidenName          *string   `json:"maiden_name,omitempty"`
	Gender              string    `json:"gender"`
	DateOfBirth         string    `json:"date_of_birth"`
	PlaceOfBirth        string    `json:"place_of_birth"`
	Nationality         string    `json:"nationality"`
	IDDocuments         []IDDocument `json:"id_documents"`
	Occupation          string    `json:"occupation"`
	Education           string    `json:"education"`
	MaritalStatus       string    `json:"marital_status"`
	SpouseName          *string   `json:"spouse_name,omitempty"`
	FamilyInfo          *FamilyInfo `json:"family_info,omitempty"`
}

// 企业客户信息
type CorporateClientInfo struct {
	LegalName           string    `json:"legal_name"`
	TradeName           *string   `json:"trade_name,omitempty"`
	RegistrationNumber  string    `json:"registration_number"`
	TaxNumber           string    `json:"tax_number"`
	Jurisdiction        string    `json:"jurisdiction"`
	IncorporationDate   string    `json:"incorporation_date"`
	CompanyType         string    `json:"company_type"`
	ShareStructure      *ShareStructure `json:"share_structure,omitempty"`
	Directors           []DirectorInfo `json:"directors"`
	Officers            []OfficerInfo `json:"officers"`
	RegisteredAddress   *AddressInfo `json:"registered_address,omitempty"`
	PrincipalAddress    *AddressInfo `json:"principal_address,omitempty"`
	BusinessActivities  []string  `json:"business_activities"`
	ParentCompany       *ParentCompanyInfo `json:"parent_company,omitempty"`
	Subsidiaries        []SubsidiaryInfo `json:"subsidiaries,omitempty"`
}

// 身份证明文件
type IDDocument struct {
	DocumentType        string    `json:"document_type"`
	DocumentNumber      string    `json:"document_number"`
	IssuingCountry      string    `json:"issuing_country"`
	IssueDate           string    `json:"issue_date"`
	ExpiryDate          *string   `json:"expiry_date,omitempty"`
	IsPrimary           bool      `json:"is_primary"`
}

// 家庭信息
type FamilyInfo struct {
	Dependents          int       `json:"dependents"`
	MinorChildren       int       `json:"minor_children"`
	FamilyOffice        *string   `json:"family_office,omitempty"`
	WealthSource        string    `json:"wealth_source"`
	NetWorth            *string   `json:"net_worth,omitempty"`
}

// 股权结构
type ShareStructure struct {
	ShareCapital        string    `json:"share_capital"`
	ShareClasses        []ShareClass `json:"share_classes"`
	BeneficialOwners    []BeneficialOwner `json:"beneficial_owners"`
}

// 股份类别
type ShareClass struct {
	ClassType           string    `json:"class_type"`
	Percentage          float64   `json:"percentage"`
	Rights              []string  `json:"rights"`
}

// 实益所有人
type BeneficialOwner struct {
	Name                string    `json:"name"`
	Percentage          float64   `json:"percentage"`
	Citizenship         string    `json:"citizenship"`
	IDNumber            string    `json:"id_number"`
}

// 董事信息
type DirectorInfo struct {
	Name                string    `json:"name"`
	Title               string    `json:"title"`
	AppointmentDate     string    `json:"appointment_date"`
	Nationality         string    `json:"nationality"`
	IsIndependent       bool      `json:"is_independent"`
}

// 高管信息
type OfficerInfo struct {
	Name                string    `json:"name"`
	Title               string    `json:"title"`
	Responsibilities    []string  `json:"responsibilities"`
	AppointmentDate     string    `json:"appointment_date"`
}

// 母公司信息
type ParentCompanyInfo struct {
	Name                string    `json:"name"`
	Jurisdiction        string    `json:"jurisdiction"`
	OwnershipPercentage float64   `json:"ownership_percentage"`
	ControlType         string    `json:"control_type"`
}

// 子公司信息
type SubsidiaryInfo struct {
	Name                string    `json:"name"`
	Jurisdiction        string    `json:"jurisdiction"`
	OwnershipPercentage float64   `json:"ownership_percentage"`
	BusinessActivities  []string  `json:"business_activities"`
}

// 联系信息
type ContactInfo struct {
	Name                string    `json:"name"`
	Title               *string   `json:"title,omitempty"`
	Email               string    `json:"email"`
	Phone               *string   `json:"phone,omitempty"`
	Mobile              *string   `json:"mobile,omitempty"`
	Address             *AddressInfo `json:"address,omitempty"`
	IsPrimary           bool      `json:"is_primary"`
}

// 地址信息
type AddressInfo struct {
	AddressLine1        string    `json:"address_line1"`
	AddressLine2        *string   `json:"address_line2,omitempty"`
	City                string    `json:"city"`
	State               *string   `json:"state,omitempty"`
	PostalCode          string    `json:"postal_code"`
	Country             string    `json:"country"`
	AddressType         string    `json:"address_type"`
}

// 客户关系
type ClientRelationship struct {
	BaseModel

	ClientID            string    `gorm:"type:varchar(36);index;comment:客户ID" json:"client_id"`
	RelatedClientID     string    `gorm:"type:varchar(36);index;comment:关联客户ID" json:"related_client_id"`
	RelationshipType    string    `gorm:"type:enum('PARENT_SUBSIDIARY','SISTER_COMPANIES','JOINT_VENTURE','SUPPLIER_CUSTOMER','COMPETITOR','AFFILIATED_PERSON','FAMILY_MEMBER','BUSINESS_PARTNER');comment:关系类型" json:"relationship_type"`
	RelationshipLevel   string    `gorm:"type:enum('DIRECT','INDIRECT','BENEFICIAL','CONTROL','SIGNIFICANT_INFLUENCE');comment:关系层级" json:"relationship_level"`
	Description         *string   `gorm:"type:text;comment:关系描述" json:"description,omitempty"`
	StartDate           *time.Time `gorm:"comment:开始日期" json:"start_date,omitempty"`
	EndDate             *time.Time `gorm:"comment:结束日期" json:"end_date,omitempty"`
	IsActive            bool      `gorm:"default:true;comment:是否活跃" json:"is_active"`
	FinancialThreshold  *float64  `gorm:"type:decimal(15,2);comment:财务门槛" json:"financial_threshold,omitempty"`
	LastUpdated         time.Time `gorm:"comment:最后更新时间" json:"last_updated"`

	Client              *ClientProfile `gorm:"foreignKey:ClientID" json:"client,omitempty"`
	RelatedClient       *ClientProfile `gorm:"foreignKey:RelatedClientID" json:"related_client,omitempty"`
}

// 客户名称变体
type ClientNameVariant struct {
	BaseModel

	ClientID            string    `gorm:"type:varchar(36);index;comment:客户ID" json:"client_id"`
	VariantName         string    `gorm:"type:varchar(500);comment:变体名称" json:"variant_name"`
	VariantType         string    `gorm:"type:enum('FORMAL_NAME','TRADE_NAME','ABBREVIATION','TRANSLITERATION','FORMER_NAME','DBA','FICTITIOUS_NAME','ALIAS');comment:变体类型" json:"variant_type"`
	Language            string    `gorm:"type:varchar(10);comment:语言" json:"language"`
	Jurisdiction        *string   `gorm:"type:varchar(100);comment:司法管辖区" json:"jurisdiction,omitempty"`
	IsPrimary           bool      `gorm:"default:false;comment:是否主要名称" json:"is_primary"`
	UsageContext        *string   `gorm:"type:text;comment:使用场景" json:"usage_context,omitempty"`
	Source              *string   `gorm:"type:varchar(100);comment:来源" json:"source,omitempty"`
	VerificationStatus  string    `gorm:"type:enum('VERIFIED','UNVERIFIED','DISPUTED');default:'UNVERIFIED';comment:验证状态" json:"verification_status"`

	Client              *ClientProfile `gorm:"foreignKey:ClientID" json:"client,omitempty"`
}

// 客户行业分类
type ClientIndustryClassification struct {
	BaseModel

	ClientID            string    `gorm:"type:varchar(36);index;comment:客户ID" json:"client_id"`
	ClassificationSystem string    `gorm:"type:enum('SIC','NAICS','ISIC','GICS','NACE','KSIC','CUSTOM');comment:分类系统" json:"classification_system"`
	Code                string    `gorm:"type:varchar(20);comment:分类代码" json:"code"`
	Description         string    `gorm:"type:varchar(500);comment:分类描述" json:"description"`
	Level               int       `gorm:"comment:分类级别" json:"level"`
	IsPrimary           bool      `gorm:"default:false;comment:是否主要分类" json:"is_primary"`
	Confidence          float64   `gorm:"type:decimal(3,2);default:1.00;comment:置信度" json:"confidence"`
	AssignedBy          string    `gorm:"type:varchar(100);comment:分配人" json:"assigned_by"`
	AssignedDate        time.Time `gorm:"comment:分配日期" json:"assigned_date"`

	Client              *ClientProfile `gorm:"foreignKey:ClientID" json:"client,omitempty"`
}

// 客户风险档案
type ClientRiskProfile struct {
	BaseModel

	ClientID            string    `gorm:"type:varchar(36);uniqueIndex;comment:客户ID" json:"client_id"`
	OverallRisk         string    `gorm:"type:enum('LOW','MEDIUM','HIGH','CRITICAL');default:'LOW';comment:整体风险等级" json:"overall_risk"`
	RiskScore           float64   `gorm:"type:decimal(5,2);default:0.00;comment:风险评分" json:"risk_score"`

	// 风险因素
	ReputationRisk      float64   `gorm:"type:decimal(3,2);default:0.00;comment:声誉风险" json:"reputation_risk"`
	FinancialRisk       float64   `gorm:"type:decimal(3,2);default:0.00;comment:财务风险" json:"financial_risk"`
	RegulatoryRisk      float64   `gorm:"type:decimal(3,2);default:0.00;comment:监管风险" json:"regulatory_risk"`
	OperationalRisk     float64   `gorm:"type:decimal(3,2);default:0.00;comment:运营风险" json:"operational_risk"`

	// 监管状态
	RegulatoryFlags     JSON      `gorm:"type:json;comment:监管标记" json:"regulatory_flags"`
	SanctionsList       []string  `gorm:"type:json;comment:制裁名单" json:"sanctions_list"`
	PepStatus           string    `gorm:"type:enum('NOT_PEP','FORMER_PEP','CURRENT_PEP','RELATED_TO_PEP');default:'NOT_PEP';comment:政治人物状态" json:"pep_status"`

	// 历史记录
	PastIncidents       []PastIncident `gorm:"type:json;comment:历史事件" json:"past_incidents"`
	DueDiligenceLevel   string    `gorm:"type:enum('BASIC','STANDARD','ENHANCED','INTENSIVE');default:'STANDARD';comment:尽职调查级别" json:"due_diligence_level"`
	LastAssessmentDate  time.Time `gorm:"comment:最后评估日期" json:"last_assessment_date"`
	NextReviewDate      time.Time `gorm:"comment:下次审查日期" json:"next_review_date"`

	// 监控要求
	MonitoringRequired  bool      `gorm:"default:false;comment:需要监控" json:"monitoring_required"`
	MonitoringFrequency string    `gorm:"type:enum('WEEKLY','MONTHLY','QUARTERLY','ANNUALLY','AD_HOC');comment:监控频率" json:"monitoring_frequency"`

	Client              *ClientProfile `gorm:"foreignKey:ClientID" json:"client,omitempty"`
}

// 历史事件
type PastIncident struct {
	IncidentType        string    `json:"incident_type"`
	IncidentDate        string    `json:"incident_date"`
	Description         string    `json:"description"`
	Resolution          *string   `json:"resolution,omitempty"`
	Impact              string    `json:"impact"`
}

// 2. 冲突分类标准模型

// 冲突类型
type ConflictType struct {
	BaseModel

	Code                string    `gorm:"type:varchar(50);uniqueIndex;comment:类型代码" json:"code"`
	Name                string    `gorm:"type:varchar(200);comment:类型名称" json:"name"`
	Category            string    `gorm:"type:varchar(100);comment:所属分类" json:"category"`
	SubCategory         *string   `gorm:"type:varchar(100);comment:子分类" json:"sub_category,omitempty"`

	// 分类标准
	IBAGuideline        *string   `gorm:"type:text;comment:IBA指导原则" json:"iba_guideline,omitempty"`
	ABARule             *string   `gorm:"type:text;comment:ABA规则" json:"aba_rule,omitempty"`
	ChineseBarRule      *string   `gorm:"type:text;comment:中国律师协会规则" json:"chinese_bar_rule,omitempty"`

	// 风险评估
	DefaultRiskLevel    string    `gorm:"type:enum('LOW','MEDIUM','HIGH','CRITICAL');default:'MEDIUM';comment:默认风险等级" json:"default_risk_level"`
	SeverityScore       float64   `gorm:"type:decimal(3,2);default:0.00;comment:严重性评分" json:"severity_score"`

	// 豁免条件
	WaiverPossible      bool      `gorm:"default:true;comment:是否可以豁免" json:"waiver_possible"`
	WaiverConditions    JSON      `gorm:"type:json;comment:豁免条件" json:"waiver_conditions"`
	ApprovalRequired    JSON      `gorm:"type:json;comment:审批要求" json:"approval_required"`

	// 描述和示例
	Description         string    `gorm:"type:text;comment:描述" json:"description"`
	Examples            []string  `gorm:"type:json;comment:示例" json:"examples"`
	Scenarios           []string  `gorm:"type:json;comment:场景" json:"scenarios"`

	// 处理指导
	ResolutionSteps     JSON      `gorm:"type:json;comment:解决步骤" json:"resolution_steps"`
	PreventiveMeasures  JSON      `gorm:"type:json;comment:预防措施" json:"preventive_measures"`

	// 状态和版本
	Status              string    `gorm:"type:enum('ACTIVE','INACTIVE','UNDER_REVIEW');default:'ACTIVE';comment:状态" json:"status"`
	Version             int       `gorm:"default:1;comment:版本" json:"version"`
	EffectiveDate       time.Time `gorm:"comment:生效日期" json:"effective_date"`
	ReviewDate          *time.Time `gorm:"comment:审查日期" json:"review_date,omitempty"`

	// 分类标准引用
	ClassificationRefs  []ClassificationReference `gorm:"foreignKey:ConflictTypeID" json:"classification_refs,omitempty"`
}

// 分类标准引用
type ClassificationReference struct {
	BaseModel

	ConflictTypeID      string    `gorm:"type:varchar(36);index;comment:冲突类型ID" json:"conflict_type_id"`
	StandardType        string    `gorm:"type:enum('IBA','ABA','CHINESE_BAR','LOCAL_REGULATION','FIRM_POLICY');comment:标准类型" json:"standard_type"`
	StandardName        string    `gorm:"type:varchar(200);comment:标准名称" json:"standard_name"`
	ReferenceSection    string    `gorm:"type:varchar(500);comment:引用章节" json:"reference_section"`
	KeyRequirements     string    `gorm:"type:text;comment:关键要求" json:"key_requirements"`
	ComplianceLevel     string    `gorm:"type:enum('MANDATORY','RECOMMENDED','BEST_PRACTICE');comment:合规级别" json:"compliance_level"`
	Jurisdiction        *string   `gorm:"type:varchar(100);comment:司法管辖区" json:"jurisdiction,omitempty"`
	LastUpdated         time.Time `gorm:"comment:最后更新" json:"last_updated"`

	ConflictType        *ConflictType `gorm:"foreignKey:ConflictTypeID" json:"conflict_type,omitempty"`
}

// 风险等级定义
type RiskLevelDefinition struct {
	BaseModel

	Level               string    `gorm:"type:varchar(20);uniqueIndex;comment:风险等级" json:"level"`
	Name                string    `gorm:"type:varchar(100);comment:等级名称" json:"name"`
	Description         string    `gorm:"type:text;comment:描述" json:"description"`
	ScoreRange          string    `gorm:"type:varchar(20);comment:评分范围" json:"score_range"`
	ColorCode           string    `gorm:"type:varchar(7);comment:颜色代码" json:"color_code"`

	// 处理要求
	ReviewRequired      bool      `gorm:"default:false;comment:需要审查" json:"review_required"`
	ApprovalRequired    bool      `gorm:"default:false;comment:需要审批" json:"approval_required"`
	EscalationTrigger   bool      `gorm:"default:false;comment:触发上报" json:"escalation_trigger"`

	// 时间要求
	ResponseTime        int       `gorm:"comment:响应时间（小时）" json:"response_time"`
	ResolutionTime      int       `gorm:"comment:解决时间（天）" json:"resolution_time"`

	// 预防措施
	PreventiveActions   JSON      `gorm:"type:json;comment:预防措施" json:"preventive_actions"`
	MonitoringPlan      JSON      `gorm:"type:json;comment:监控计划" json:"monitoring_plan"`
}

// 3. 豁免管理模型

// 豁免申请
type WaiverApplication struct {
	BaseModel

	ApplicationNumber   string    `gorm:"type:varchar(50);uniqueIndex;comment:申请编号" json:"application_number"`

	// 关联信息
	ConflictCheckID     string    `gorm:"type:varchar(36);index;comment:关联的冲突检查ID" json:"conflict_check_id"`
	CaseID              *string   `gorm:"type:varchar(36);index;comment:关联的案件ID" json:"case_id,omitempty"`
	ClientID            string    `gorm:"type:varchar(36);index;comment:客户ID" json:"client_id"`
	LawyerID            string    `gorm:"type:varchar(36);index;comment:律师ID" json:"lawyer_id"`

	// 申请类型
	WaiverType          string    `gorm:"type:enum('INFORMED_CONSENT','ETHICAL_BARRIER','INFORMATION_SCREEN','STRUCTURAL_BARRIER');comment:豁免类型" json:"waiver_type"`
	WaiverCategory      string    `gorm:"type:enum('CLIENT_CONSENT','BARRIER_IMPLEMENTATION','MONITORING_ARRANGEMENT','SPECIAL_CIRCUMSTANCES');comment:豁免类别" json:"waiver_category"`

	// 冲突详情
	ConflictSummary     string    `gorm:"type:text;comment:冲突情况摘要" json:"conflict_summary"`
	Conflicts           JSON      `gorm:"type:json;comment:冲突详情列表" json:"conflicts"`
	RiskAssessment      JSON      `gorm:"type:json;comment:风险评估结果" json:"risk_assessment"`

	// 豁免条件
	ProposedConditions  JSON      `gorm:"type:json;comment:建议的豁免条件" json:"proposed_conditions"`
	Limitations         JSON      `gorm:"type:json;comment:限制条件" json:"limitations"`
	MonitoringRequirements JSON   `gorm:"type:json;comment:监控要求" json:"monitoring_requirements"`
	ReportingRequirements JSON   `gorm:"type:json;comment:报告要求" json:"reporting_requirements"`

	// 时间管理
	RequestedEffectiveDate time.Time `gorm:"comment:申请生效日期" json:"requested_effective_date"`
	RequestedExpiryDate   *time.Time `gorm:"comment:申请到期日期" json:"requested_expiry_date,omitempty"`
	DurationDays         *int      `gorm:"comment:豁免期限（天）" json:"duration_days,omitempty"`

	// 申请理由
	Rationale            string    `gorm:"type:text;comment:申请理由" json:"rationale"`
	SupportingEvidence   JSON      `gorm:"type:json;comment:支持证据" json:"supporting_evidence"`
	AlternativesConsidered JSON    `gorm:"type:json;comment:考虑的其他方案" json:"alternatives_considered"`

	// 客户信息
	ClientRepresentativeName string `gorm:"type:varchar(255);comment:客户代表姓名" json:"client_representative_name"`
	ClientRepresentativeTitle *string `gorm:"type:varchar(100);comment:客户代表职位" json:"client_representative_title,omitempty"`
	ClientRepresentativeContact *string `gorm:"type:varchar(255);comment:客户代表联系方式" json:"client_representative_contact,omitempty"`

	// 律师信息
	RequestingLawyerName string    `gorm:"type:varchar(255);comment:申请律师姓名" json:"requesting_lawyer_name"`
	RequestingLawyerTitle *string `gorm:"type:varchar(100);comment:申请律师职位" json:"requesting_lawyer_title,omitempty"`
	SupervisingLawyerName *string `gorm:"type:varchar(255);comment:监督律师姓名" json:"supervising_lawyer_name,omitempty"`

	// 状态管理
	Status              string    `gorm:"type:enum('DRAFT','SUBMITTED','UNDER_REVIEW','REVIEW_COMPLETED','APPROVED','REJECTED','EXPIRED','REVOKED');default:'DRAFT';comment:状态" json:"status"`
	SubmissionDate      *time.Time `gorm:"comment:提交日期" json:"submission_date,omitempty"`
	ReviewPriority      string    `gorm:"type:enum('LOW','MEDIUM','HIGH','URGENT');default:'MEDIUM';comment:审核优先级" json:"review_priority"`

	// 审批流程
	CurrentStage        *string   `gorm:"type:enum('INITIAL_REVIEW','DEPARTMENT_REVIEW','COMPLIANCE_REVIEW','MANAGEMENT_APPROVAL','FINAL_APPROVAL');comment:当前阶段" json:"current_stage,omitempty"`
	AssignedReviewer    *string   `gorm:"type:varchar(36);comment:分配的审核人" json:"assigned_reviewer,omitempty"`
	ReviewDeadline      *time.Time `gorm:"comment:审核截止日期" json:"review_deadline,omitempty"`

	// 审计信息
	CreatedBy           string    `gorm:"type:varchar(36);comment:创建人" json:"created_by"`
	UpdatedBy           *string   `gorm:"type:varchar(36);comment:更新人" json:"updated_by,omitempty"`

	// 关联数据
	ApprovalRecords     []WaiverApprovalRecord `gorm:"foreignKey:WaiverApplicationID" json:"approval_records,omitempty"`
	Signatures          []WaiverSignature      `gorm:"foreignKey:WaiverApplicationID" json:"signatures,omitempty"`
	MonitoringRecords   []WaiverMonitoringRecord `gorm:"foreignKey:WaiverApplicationID" json:"monitoring_records,omitempty"`
}

// 豁免审批记录
type WaiverApprovalRecord struct {
	BaseModel

	WaiverApplicationID string    `gorm:"type:varchar(36);index;comment:豁免申请ID" json:"waiver_application_id"`

	// 审批基本信息
	ApprovalStage       string    `gorm:"type:varchar(100);comment:审批阶段" json:"approval_stage"`
	ApproverID          string    `gorm:"type:varchar(36);comment:审批人ID" json:"approver_id"`
	ApproverName        string    `gorm:"type:varchar(255);comment:审批人姓名" json:"approver_name"`
	ApproverTitle       *string   `gorm:"type:varchar(100);comment:审批人职位" json:"approver_title,omitempty"`
	ApproverRole        string    `gorm:"type:enum('LAWYER','DEPARTMENT_HEAD','COMPLIANCE_OFFICER','MANAGING_PARTNER','ETHICS_COMMITTEE');comment:审批人角色" json:"approver_role"`

	// 审批决定
	Decision            string    `gorm:"type:enum('APPROVE','REJECT','REQUEST_CHANGES','DEFER','ESCALATE');comment:审批决定" json:"decision"`
	DecisionReason      string    `gorm:"type:text;comment:审批理由" json:"decision_reason"`
	DecisionComments    *string   `gorm:"type:text;comment:审批意见" json:"decision_comments,omitempty"`

	// 条件和限制
	ApprovedConditions  JSON      `gorm:"type:json;comment:批准的条件" json:"approved_conditions"`
	ImposedLimitations  JSON      `gorm:"type:json;comment:施加的限制" json:"imposed_limitations"`
	MonitoringRequirements JSON   `gorm:"type:json;comment:监控要求" json:"monitoring_requirements"`
	ReportingRequirements JSON   `gorm:"type:json;comment:报告要求" json:"reporting_requirements"`

	// 风险评估
	RiskAssessment      JSON      `gorm:"type:json;comment:风险评估结果" json:"risk_assessment"`
	RiskMitigationPlan  JSON      `gorm:"type:json;comment:风险缓解计划" json:"risk_mitigation_plan"`
	FollowUpActions     JSON      `gorm:"type:json;comment:后续行动" json:"follow_up_actions"`

	// 时间信息
	ApprovalDate        time.Time `gorm:"comment:审批日期" json:"approval_date"`
	EffectiveDate       *time.Time `gorm:"comment:生效日期" json:"effective_date,omitempty"`
	ExpiryDate          *time.Time `gorm:"comment:到期日期" json:"expiry_date,omitempty"`
	NextReviewDate      *time.Time `gorm:"comment:下次审查日期" json:"next_review_date,omitempty"`

	// 附件和证据
	SupportingDocuments JSON      `gorm:"type:json;comment:支持文件" json:"supporting_documents"`
	EvidenceReferences  JSON      `gorm:"type:json;comment:证据引用" json:"evidence_references"`

	// 状态
	Status              string    `gorm:"type:enum('ACTIVE','SUPERSEDED','EXPIRED','REVOKED');default:'ACTIVE';comment:状态" json:"status"`

	WaiverApplication   *WaiverApplication `gorm:"foreignKey:WaiverApplicationID" json:"waiver_application,omitempty"`
}

// 电子签名记录
type WaiverSignature struct {
	BaseModel

	WaiverApplicationID string    `gorm:"type:varchar(36);index;comment:豁免申请ID" json:"waiver_application_id"`

	// 签名方信息
	SignerType          string    `gorm:"type:enum('CLIENT','LAWYER','WITNESS','NOTARY','APPROVER');comment:签名方类型" json:"signer_type"`
	SignerName          string    `gorm:"type:varchar(255);comment:签名方姓名" json:"signer_name"`
	SignerTitle         *string   `gorm:"type:varchar(100);comment:签名方职位" json:"signer_title,omitempty"`
	SignerOrganization  *string   `gorm:"type:varchar(255);comment:签名方组织" json:"signer_organization,omitempty"`
	SignerContact       *string   `gorm:"type:varchar(255);comment:签名方联系方式" json:"signer_contact,omitempty"`

	// 签名内容
	SignatureContent    string    `gorm:"type:text;comment:签名内容" json:"signature_content"`
	SignedStatement     string    `gorm:"type:text;comment:签署的声明" json:"signed_statement"`
	TermsAccepted       JSON      `gorm:"type:json;comment:接受的条款" json:"terms_accepted"`

	// 签名技术信息
	SignatureMethod     string    `gorm:"type:enum('ELECTRONIC_SIGNATURE','DIGITAL_SIGNATURE','WET_SIGNATURE','VERBAL_CONSENT');comment:签名方式" json:"signature_method"`
	SignatureAlgorithm  *string   `gorm:"type:varchar(100);comment:签名算法" json:"signature_algorithm,omitempty"`
	DigitalSignatureHash *string  `gorm:"type:varchar(500);comment:数字签名哈希" json:"digital_signature_hash,omitempty"`
	SignatureTimestamp  time.Time `gorm:"comment:签名时间戳" json:"signature_timestamp"`

	// 验证信息
	VerificationStatus   string    `gorm:"type:enum('VERIFIED','PENDING','FAILED','EXPIRED');default:'PENDING';comment:验证状态" json:"verification_status"`
	VerificationMethod   *string   `gorm:"type:varchar(100);comment:验证方式" json:"verification_method,omitempty"`
	VerificationResult   JSON      `gorm:"type:json;comment:验证结果" json:"verification_result"`
	VerifiedAt          *time.Time `gorm:"comment:验证时间" json:"verified_at,omitempty"`

	// IP地址和设备信息
	SignerIPAddress     *string   `gorm:"type:varchar(45);comment:签名方IP地址" json:"signer_ip_address,omitempty"`
	UserAgent           *string   `gorm:"type:text;comment:用户代理信息" json:"user_agent,omitempty"`
	DeviceFingerprint   *string   `gorm:"type:varchar(255);comment:设备指纹" json:"device_fingerprint,omitempty"`
	LocationInfo        JSON      `gorm:"type:json;comment:位置信息" json:"location_info"`

	// 附加文件
	SignedDocumentURL   *string   `gorm:"type:varchar(500);comment:签名文件URL" json:"signed_document_url,omitempty"`
	BackupDocumentURL   *string   `gorm:"type:varchar(500);comment:备份文件URL" json:"backup_document_url,omitempty"`
	CertificateURL      *string   `gorm:"type:varchar(500);comment:证书URL" json:"certificate_url,omitempty"`

	// 状态
	Status              string    `gorm:"type:enum('ACTIVE','REVOKED','EXPIRED','INVALID');default:'ACTIVE';comment:状态" json:"status"`
	RevocationReason    *string   `gorm:"type:text;comment:撤销原因" json:"revocation_reason,omitempty"`
	RevocationDate      *time.Time `gorm:"comment:撤销日期" json:"revocation_date,omitempty"`

	WaiverApplication   *WaiverApplication `gorm:"foreignKey:WaiverApplicationID" json:"waiver_application,omitempty"`
}

// 豁免监控记录
type WaiverMonitoringRecord struct {
	BaseModel

	WaiverApplicationID string    `gorm:"type:varchar(36);index;comment:豁免申请ID" json:"waiver_application_id"`

	// 监控基本信息
	MonitoringType      string    `gorm:"type:enum('COMPLIANCE_CHECK','RISK_ASSESSMENT','BARRIER_EFFECTIVENESS','PERFORMANCE_REVIEW','CLIENT_FEEDBACK');comment:监控类型" json:"monitoring_type"`
	MonitoringDate      time.Time `gorm:"comment:监控日期" json:"monitoring_date"`
	MonitoringPeriodStart *time.Time `gorm:"comment:监控期间开始" json:"monitoring_period_start,omitempty"`
	MonitoringPeriodEnd   *time.Time `gorm:"comment:监控期间结束" json:"monitoring_period_end,omitempty"`

	// 监控内容
	MonitoringItems     JSON      `gorm:"type:json;comment:监控项目" json:"monitoring_items"`
	CheckResults        JSON      `gorm:"type:json;comment:检查结果" json:"check_results"`
	Findings            JSON      `gorm:"type:json;comment:发现的问题" json:"findings"`
	Observations        *string   `gorm:"type:text;comment:观察记录" json:"observations,omitempty"`

	// 风险评估
	CurrentRiskLevel    *string   `gorm:"type:enum('LOW','MEDIUM','HIGH','CRITICAL');comment:当前风险等级" json:"current_risk_level,omitempty"`
	RiskTrend           *string   `gorm:"type:enum('IMPROVING','STABLE','DETERIORATING');comment:风险趋势" json:"risk_trend,omitempty"`
	RiskFactors         JSON      `gorm:"type:json;comment:风险因素" json:"risk_factors"`
	RiskMitigationStatus JSON     `gorm:"type:json;comment:风险缓解状态" json:"risk_mitigation_status"`

	// 合规状况
	ComplianceStatus    string    `gorm:"type:enum('COMPLIANT','PARTIALLY_COMPLIANT','NON_COMPLIANT','UNDER_REVIEW');default:'UNDER_REVIEW';comment:合规状况" json:"compliance_status"`
	ComplianceIssues    JSON      `gorm:"type:json;comment:合规问题" json:"compliance_issues"`
	CorrectiveActions   JSON      `gorm:"type:json;comment:纠正措施" json:"corrective_actions"`

	// 豁免条件执行情况
	ConditionsCompliance JSON     `gorm:"type:json;comment:条件执行情况" json:"conditions_compliance"`
	LimitationAdherence  JSON     `gorm:"type:json;comment:限制遵守情况" json:"limitation_adherence"`
	BarrierEffectiveness JSON     `gorm:"type:json;comment:屏障有效性" json:"barrier_effectiveness"`

	// 后续行动
	RecommendedActions  JSON      `gorm:"type:json;comment:建议行动" json:"recommended_actions"`
	RequiredFollowUp    JSON      `gorm:"type:json;comment:需要跟进事项" json:"required_follow_up"`
	NextMonitoringDate  *time.Time `gorm:"comment:下次监控日期" json:"next_monitoring_date,omitempty"`

	// 报告信息
	ReportGenerated     bool      `gorm:"default:false;comment:是否已生成报告" json:"report_generated"`
	ReportURL           *string   `gorm:"type:varchar(500);comment:报告URL" json:"report_url,omitempty"`
	ReportRecipients    JSON      `gorm:"type:json;comment:报告接收人" json:"report_recipients"`

	// 状态
	Status              string    `gorm:"type:enum('SCHEDULED','IN_PROGRESS','COMPLETED','OVERDUE','CANCELLED');default:'SCHEDULED';comment:状态" json:"status"`

	// 审计信息
	MonitoredBy         string    `gorm:"type:varchar(36);comment:监控人" json:"monitored_by"`
	ReviewedBy          *string   `gorm:"type:varchar(36);comment:审核人" json:"reviewed_by,omitempty"`

	WaiverApplication   *WaiverApplication `gorm:"foreignKey:WaiverApplicationID" json:"waiver_application,omitempty"`
}

// 豁免模板
type WaiverTemplate struct {
	BaseModel

	TemplateName        string    `gorm:"type:varchar(255);comment:模板名称" json:"template_name"`
	TemplateCode        string    `gorm:"type:varchar(100);uniqueIndex;comment:模板代码" json:"template_code"`

	// 模板分类
	TemplateType        string    `gorm:"type:enum('INFORMED_CONSENT','ETHICAL_BARRIER','INFORMATION_SCREEN','MONITORING_PLAN');comment:模板类型" json:"template_type"`
	TemplateCategory    *string   `gorm:"type:varchar(100);comment:模板分类" json:"template_category,omitempty"`
	PracticeArea        *string   `gorm:"type:varchar(100);comment:执业领域" json:"practice_area,omitempty"`

	// 模板内容
	TemplateContent     JSON      `gorm:"type:json;comment:模板内容" json:"template_content"`
	RequiredClauses     JSON      `gorm:"type:json;comment:必需条款" json:"required_clauses"`
	OptionalClauses     JSON      `gorm:"type:json;comment:可选条款" json:"optional_clauses"`
	Placeholders        JSON      `gorm:"type:json;comment:占位符说明" json:"placeholders"`

	// 适用条件
	ApplicableScenarios JSON      `gorm:"type:json;comment:适用场景" json:"applicable_scenarios"`
	ConflictTypes       JSON      `gorm:"type:json;comment:适用的冲突类型" json:"conflict_types"`
	RiskLevels          JSON      `gorm:"type:json;comment:适用的风险等级" json:"risk_levels"`
	JurisdictionRules   JSON      `gorm:"type:json;comment:司法管辖区规则" json:"jurisdiction_rules"`

	// 审批要求
	ApprovalRequirements JSON     `gorm:"type:json;comment:审批要求" json:"approval_requirements"`
	RequiredApprovers   JSON      `gorm:"type:json;comment:必需审批人" json:"required_approvers"`
	ApprovalWorkflow    JSON      `gorm:"type:json;comment:审批流程" json:"approval_workflow"`

	// 状态和版本
	Status              string    `gorm:"type:enum('ACTIVE','INACTIVE','UNDER_REVIEW','DEPRECATED');default:'ACTIVE';comment:状态" json:"status"`
	Version             int       `gorm:"default:1;comment:版本号" json:"version"`
	EffectiveDate       time.Time `gorm:"comment:生效日期" json:"effective_date"`
	ExpiryDate          *time.Time `gorm:"comment:到期日期" json:"expiry_date,omitempty"`

	// 使用统计
	UsageCount          int       `gorm:"default:0;comment:使用次数" json:"usage_count"`
	LastUsedDate        *time.Time `gorm:"comment:最后使用日期" json:"last_used_date,omitempty"`

	// 审计信息
	CreatedBy           string    `gorm:"type:varchar(36);comment:创建人" json:"created_by"`
	UpdatedBy           *string   `gorm:"type:varchar(36);comment:更新人" json:"updated_by,omitempty"`
	ApprovedBy          *string   `gorm:"type:varchar(36);comment:审批人" json:"approved_by,omitempty"`
}

// 4. 增强冲突检测模型

// 专业冲突检查请求
type ProfessionalConflictCheckRequest struct {
	BaseModel

	CheckNumber         string    `gorm:"type:varchar(50);uniqueIndex;comment:检查编号" json:"check_number"`

	// 检查基本信息
	CheckType           string    `gorm:"type:enum('NEW_MATTER_ENGAGEMENT','LAWYER_HIRING','CLIENT_ONBOARDING','ONGOING_MONITORING','PERIODIC_REVIEW','SPECIAL_REQUEST');comment:检查类型" json:"check_type"`
	Priority            string    `gorm:"type:enum('LOW','MEDIUM','HIGH','URGENT','CRITICAL');default:'MEDIUM';comment:优先级" json:"priority"`
	RequestedBy         string    `gorm:"type:varchar(36);comment:申请人" json:"requested_by"`
	Department          string    `gorm:"type:varchar(100);comment:申请部门" json:"department"`

	// 关联实体
	CaseID              *string   `gorm:"type:varchar(36);index;comment:案件ID" json:"case_id,omitempty"`
	ClientIDs           JSON      `gorm:"type:json;comment:客户ID列表" json:"client_ids"`
	LawyerIDs           JSON      `gorm:"type:json;comment:律师ID列表" json:"lawyer_ids"`
	MatterDetails       JSON      `gorm:"type:json;comment:案件详情" json:"matter_details"`

	// 检查范围
	CheckScope          JSON      `gorm:"type:json;comment:检查范围" json:"check_scope"`
	EntitiesToCheck     JSON      `gorm:"type:json;comment:待检查实体" json:"entities_to_check"`
	CheckParameters     JSON      `gorm:"type:json;comment:检查参数" json:"check_parameters"`

	// 时间要求
	RequestedDate       time.Time `gorm:"comment:申请日期" json:"requested_date"`
	RequiredByDate      *time.Time `gorm:"comment:要求完成日期" json:"required_by_date,omitempty"`
	EstimatedDuration   *int      `gorm:"comment:预计用时（分钟）" json:"estimated_duration,omitempty"`

	// 状态管理
	Status              string    `gorm:"type:enum('PENDING','IN_PROGRESS','COMPLETED','REVIEW_REQUIRED','APPROVED','REJECTED','CANCELLED');default:'PENDING';comment:状态" json:"status"`
	CurrentStage        *string   `gorm:"type:varchar(100);comment:当前阶段" json:"current_stage,omitempty"`
	AssignedTo          *string   `gorm:"type:varchar(36);comment:分配给" json:"assigned_to,omitempty"`

	// 检查结果
	OverallResult       *string   `gorm:"type:enum('NO_CONFLICT','POTENTIAL_CONFLICT','ACTUAL_CONFLICT','UNABLE_TO_DETERMINE');comment:整体结果" json:"overall_result,omitempty"`
	ConflictsFound      int       `gorm:"default:0;comment:发现冲突数量" json:"conflicts_found"`
	RiskLevel           *string   `gorm:"type:enum('LOW','MEDIUM','HIGH','CRITICAL');comment:风险等级" json:"risk_level,omitempty"`

	// 处理要求
	ReviewRequired      bool      `gorm:"default:false;comment:需要审查" json:"review_required"`
	ApprovalRequired    bool      `gorm:"default:false;comment:需要审批" json:"approval_required"`
	WaiverRequired      bool      `gorm:"default:false;comment:需要豁免" json:"waiver_required"`

	// 审计信息
	StartedAt           *time.Time `gorm:"comment:开始时间" json:"started_at,omitempty"`
	CompletedAt         *time.Time `gorm:"comment:完成时间" json:"completed_at,omitempty"`
	ReviewedBy          *string   `gorm:"type:varchar(36);comment:审查人" json:"reviewed_by,omitempty"`
	ReviewedAt          *time.Time `gorm:"comment:审查时间" json:"reviewed_at,omitempty"`
	ApprovedBy          *string   `gorm:"type:varchar(36);comment:审批人" json:"approved_by,omitempty"`
	ApprovedAt          *time.Time `gorm:"comment:审批时间" json:"approved_at,omitempty"`

	// 关联数据
	ConflictResults     []MultidimensionalConflictResult `gorm:"foreignKey:CheckRequestID" json:"conflict_results,omitempty"`
	RuleExecutions      []ConflictRuleExecution `gorm:"foreignKey:CheckRequestID" json:"rule_executions,omitempty"`
}

// 多维度冲突结果
type MultidimensionalConflictResult struct {
	BaseModel

	CheckRequestID      string    `gorm:"type:varchar(36);index;comment:检查请求ID" json:"check_request_id"`

	// 冲突基本信息
	ConflictID          string    `gorm:"type:varchar(50);uniqueIndex;comment:冲突ID" json:"conflict_id"`
	ConflictType        string    `gorm:"type:varchar(100);comment:冲突类型" json:"conflict_type"`
	ConflictCategory    string    `gorm:"type:varchar(100);comment:冲突类别" json:"conflict_category"`
	SeverityLevel       string    `gorm:"type:enum('LOW','MEDIUM','HIGH','CRITICAL');comment:严重程度" json:"severity_level"`

	// 涉及实体
	PrimaryEntity       JSON      `gorm:"type:json;comment:主要实体" json:"primary_entity"`
	SecondaryEntity     JSON      `gorm:"type:json;comment:次要实体" json:"secondary_entity"`
	RelatedEntities     JSON      `gorm:"type:json;comment:相关实体" json:"related_entities"`

	// 冲突详情
	Description         string    `gorm:"type:text;comment:冲突描述" json:"description"`
	Scenario            string    `gorm:"type:text;comment:具体场景" json:"scenario"`
	Implications        JSON      `gorm:"type:json;comment:影响分析" json:"implications"`
	RiskFactors         JSON      `gorm:"type:json;comment:风险因素" json:"risk_factors"`

	// 检测方法
	DetectionMethod     string    `gorm:"type:varchar(100);comment:检测方法" json:"detection_method"`
	Confidence          float64   `gorm:"type:decimal(3,2);default:0.00;comment:置信度" json:"confidence"`
	EvidenceStrength    string    `gorm:"type:enum('WEAK','MODERATE','STRONG','CONCLUSIVE');comment:证据强度" json:"evidence_strength"`

	// 法律依据
	LegalBasis          JSON      `gorm:"type:json;comment:法律依据" json:"legal_basis"`
	RegulatoryReferences JSON     `gorm:"type:json;comment:监管引用" json:"regulatory_references"`
	CasePrecedents      JSON      `gorm:"type:json;comment:案例先例" json:"case_precedents"`

	// 处理建议
	RecommendedAction   string    `gorm:"type:enum('PROCEED','CAUTION','REJECT','SEEK_WAIVER','IMPLEMENT_BARRIER','REFER_TO_ETHICS_COMMITTEE');comment:建议行动" json:"recommended_action"`
	ResolutionOptions   JSON      `gorm:"type:json;comment:解决方案选项" json:"resolution_options"`
	PreventiveMeasures  JSON      `gorm:"type:json;comment:预防措施" json:"preventive_measures"`

	// 豁免可能性
	WaiverPossible      bool      `gorm:"default:false;comment:是否可豁免" json:"waiver_possible"`
	WaiverConditions    JSON      `gorm:"type:json;comment:豁免条件" json:"waiver_conditions"`
	WaiverProcess       JSON      `gorm:"type:json;comment:豁免流程" json:"waiver_process"`

	// 状态跟踪
	Status              string    `gorm:"type:enum('DETECTED','UNDER_REVIEW','RESOLVED','WAVED','MONITORED');default:'DETECTED';comment:状态" json:"status"`
	AssignedTo          *string   `gorm:"type:varchar(36);comment:分配给" json:"assigned_to,omitempty"`
	ResolvedBy          *string   `gorm:"type:varchar(36);comment:解决人" json:"resolved_by,omitempty"`
	ResolvedAt          *time.Time `gorm:"comment:解决时间" json:"resolved_at,omitempty"`
	Resolution          JSON      `gorm:"type:json;comment:解决方案" json:"resolution"`

	// 后续监控
	MonitoringRequired  bool      `gorm:"default:false;comment:需要监控" json:"monitoring_required"`
	MonitoringPlan      JSON      `gorm:"type:json;comment:监控计划" json:"monitoring_plan"`
	NextReviewDate      *time.Time `gorm:"comment:下次审查日期" json:"next_review_date,omitempty"`

	CheckRequest        *ProfessionalConflictCheckRequest `gorm:"foreignKey:CheckRequestID" json:"check_request,omitempty"`
}

// 冲突检测规则
type ConflictDetectionRule struct {
	BaseModel

	RuleCode            string    `gorm:"type:varchar(100);uniqueIndex;comment:规则代码" json:"rule_code"`
	RuleName            string    `gorm:"type:varchar(200);comment:规则名称" json:"rule_name"`
	RuleType            string    `gorm:"type:enum('ENTITY_MATCH','RELATIONSHIP_DETECTION','TEMPORAL_PROXIMITY','FINANCIAL_THRESHOLD','BUSINESS_COMPETITION','REGULATORY_RESTRICTION','FIRM_POLICY');comment:规则类型" json:"rule_type"`
	RuleCategory        string    `gorm:"type:varchar(100);comment:规则分类" json:"rule_category"`

	// 规则条件
	Conditions          JSON      `gorm:"type:json;comment:触发条件" json:"conditions"`
	Parameters          JSON      `gorm:"type:json;comment:规则参数" json:"parameters"`
	Thresholds          JSON      `gorm:"type:json;comment:阈值设置" json:"thresholds"`

	// 执行逻辑
	ExecutionLogic      string    `gorm:"type:text;comment:执行逻辑" json:"execution_logic"`
	Algorithms          JSON      `gorm:"type:json;comment:算法配置" json:"algorithms"`
	DataSources         JSON      `gorm:"type:json;comment:数据源" json:"data_sources"`

	// 优先级和权重
	Priority            int       `gorm:"default:50;comment:优先级" json:"priority"`
	Weight              float64   `gorm:"type:decimal(3,2);default:1.00;comment:权重" json:"weight"`

	// 输出配置
	OutputFormat        JSON      `gorm:"type:json;comment:输出格式" json:"output_format"`
	ResultMapping       JSON      `gorm:"type:json;comment:结果映射" json:"result_mapping"`

	// 状态管理
	Status              string    `gorm:"type:enum('ACTIVE','INACTIVE','TESTING','DEPRECATED');default:'ACTIVE';comment:状态" json:"status"`
	Version             int       `gorm:"default:1;comment:版本" json:"version"`

	// 性能统计
	ExecutionCount      int       `gorm:"default:0;comment:执行次数" json:"execution_count"`
	SuccessRate         float64   `gorm:"type:decimal(5,2);default:0.00;comment:成功率" json:"success_rate"`
	AvgExecutionTime    int       `gorm:"default:0;comment:平均执行时间（毫秒）" json:"avg_execution_time"`
	LastExecuted        *time.Time `gorm:"comment:最后执行时间" json:"last_executed,omitempty"`

	// 规则执行记录
	RuleExecutions      []ConflictRuleExecution `gorm:"foreignKey:RuleID" json:"rule_executions,omitempty"`
}

// 冲突规则执行记录
type ConflictRuleExecution struct {
	BaseModel

	CheckRequestID      string    `gorm:"type:varchar(36);index;comment:检查请求ID" json:"check_request_id"`
	RuleID              string    `gorm:"type:varchar(36);index;comment:规则ID" json:"rule_id"`

	// 执行信息
	ExecutionSequence   int       `gorm:"comment:执行序号" json:"execution_sequence"`
	TriggerConditions   JSON      `gorm:"type:json;comment:触发条件" json:"trigger_conditions"`
	InputData           JSON      `gorm:"type:json;comment:输入数据" json:"input_data"`

	// 执行结果
	ExecutionResult     string    `gorm:"type:enum('SUCCESS','FAILURE','ERROR','SKIPPED','TIMEOUT');comment:执行结果" json:"execution_result"`
	ResultData          JSON      `gorm:"type:json;comment:结果数据" json:"result_data"`
	ConflictsDetected   int       `gorm:"default:0;comment:检测到的冲突数" json:"conflicts_detected"`

	// 性能指标
	ExecutionTime       int       `gorm:"comment:执行时间（毫秒）" json:"execution_time"`
	MemoryUsed          int       `gorm:"comment:内存使用（字节）" json:"memory_used"`
	RecordsProcessed    int       `gorm:"default:0;comment:处理记录数" json:"records_processed"`

	// 错误信息
	ErrorType           *string   `gorm:"type:varchar(100);comment:错误类型" json:"error_type,omitempty"`
	ErrorMessage        *string   `gorm:"type:text;comment:错误信息" json:"error_message,omitempty"`
	StackTrace          *string   `gorm:"type:text;comment:错误堆栈" json:"stack_trace,omitempty"`

	// 审计信息
	ExecutedBy          string    `gorm:"type:varchar(36);comment:执行人" json:"executed_by"`
	ExecutedAt          time.Time `gorm:"comment:执行时间" json:"executed_at"`

	CheckRequest        *ProfessionalConflictCheckRequest `gorm:"foreignKey:CheckRequestID" json:"check_request,omitempty"`
	Rule                *ConflictDetectionRule `gorm:"foreignKey:RuleID" json:"rule,omitempty"`
}

// 专业冲突检查统计视图
type ProfessionalConflictCheckStats struct {
	PeriodStart         time.Time `gorm:"comment:统计期间开始" json:"period_start"`
	PeriodEnd           time.Time `gorm:"comment:统计期间结束" json:"period_end"`

	// 请求数量统计
	TotalRequests       int64       `gorm:"comment:总请求数" json:"total_requests"`
	PendingRequests     int64       `gorm:"comment:待处理请求" json:"pending_requests"`
	CompletedRequests   int64       `gorm:"comment:已完成请求" json:"completed_requests"`

	// 冲突检测统计
	TotalConflicts      int64       `gorm:"comment:发现冲突总数" json:"total_conflicts"`
	CriticalConflicts   int64       `gorm:"comment:严重冲突数" json:"critical_conflicts"`
	HighConflicts       int64       `gorm:"comment:高风险冲突数" json:"high_conflicts"`
	MediumConflicts     int64       `gorm:"comment:中风险冲突数" json:"medium_conflicts"`
	LowConflicts        int64       `gorm:"comment:低风险冲突数" json:"low_conflicts"`

	// 处理效率统计
	AvgProcessingTime   int       `gorm:"comment:平均处理时间（分钟）" json:"avg_processing_time"`
	SlaComplianceRate   float64   `gorm:"type:decimal(5,2);comment:SLA合规率" json:"sla_compliance_rate"`

	// 豁免统计
	WaiverRequests      int64       `gorm:"comment:豁免申请数" json:"waiver_requests"`
	WaiversApproved     int64       `gorm:"comment:豁免批准数" json:"waivers_approved"`
	WaiversRejected     int64       `gorm:"comment:豁免拒绝数" json:"waivers_rejected"`
}