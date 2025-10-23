package compliance

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// KYCVerificationType KYC验证类型
type KYCVerificationType string

const (
	KYCTypeBasic            KYCVerificationType = "BASIC"            // 基础验证
	KYCTypeStandard         KYCVerificationType = "STANDARD"         // 标准验证
	KYCTypeEnhanced         KYCVerificationType = "ENHANCED"         // 增强验证
	KYCTypeUltimate         KYCVerificationType = "ULTIMATE"         // 终极验证
	KYCTypePEP              KYCVerificationType = "PEP"              // PEP验证
	KYCTypeSanctions        KYCVerificationType = "SANCTIONS"        // 制裁名单验证
	KYCTypeDocument         KYCVerificationType = "DOCUMENT"         // 文档验证
	KYCTypeBiometric        KYCVerificationType = "BIOMETRIC"        // 生物识别验证
	KYCTypeAddress          KYCVerificationType = "ADDRESS"          // 地址验证
	KYCTypeSourceOfFunds    KYCVerificationType = "SOURCE_OF_FUNDS"  // 资金来源验证
)

// PersonalInfo 个人信息
type PersonalInfo struct {
	Title           string                 `json:"title,omitempty"`
	FirstName       string                 `json:"first_name"`
	MiddleName      string                 `json:"middle_name,omitempty"`
	LastName        string                 `json:"last_name"`
	Suffix          string                 `json:"suffix,omitempty"`
	FullName        string                 `json:"full_name"`
	LocalName       string                 `json:"local_name,omitempty"`
	DateOfBirth     *time.Time             `json:"date_of_birth"`
	PlaceOfBirth    string                 `json:"place_of_birth,omitempty"`
	CountryOfBirth  string                 `json:"country_of_birth,omitempty"`
	Nationality     []string               `json:"nationality"`
	Citizenship     []string               `json:"citizenship"`
	Gender          string                 `json:"gender,omitempty"`
	MaritalStatus   string                 `json:"marital_status,omitempty"`
	Occupation      string                 `json:"occupation,omitempty"`
	Employer        string                 `json:"employer,omitempty"`
	Industry        string                 `json:"industry,omitempty"`
	AnnualIncome    float64                `json:"annual_income,omitempty"`
	TaxIDNumber    string                 `json:"tax_id_number,omitempty"`
	SocialSecurity string                 `json:"social_security,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// Document 文档信息
type Document struct {
	DocumentID       string                 `json:"document_id"`
	DocumentType     string                 `json:"document_type"`
	DocumentNumber   string                 `json:"document_number"`
	IssuingCountry   string                 `json:"issuing_country"`
	IssuingAuthority string                 `json:"issuing_authority,omitempty"`
	IssueDate        *time.Time             `json:"issue_date,omitempty"`
	ExpiryDate       *time.Time             `json:"expiry_date,omitempty"`
	CountryOfBirth    string                 `json:"country_of_birth,omitempty"`
	DocumentImage    DocumentImage         `json:"document_image"`
	VerificationData VerificationData     `json:"verification_data"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// DocumentImage 文档图像
type DocumentImage struct {
	FrontImage    string    `json:"front_image"`    // Base64编码的正面图像
	BackImage     string    `json:"back_image"`     // Base64编码的背面图像
	SelfieImage   string    `json:"selfie_image"`   // Base64编码的自拍图像
	ImageHash     string    `json:"image_hash"`     // 图像哈希
	QualityScore   float64   `json:"quality_score"`  // 图像质量评分
	UploadTime     time.Time `json:"upload_time"`
	ExifData       string    `json:"exif_data"`      // EXIF数据
}

// VerificationData 验证数据
type VerificationData struct {
	VerificationStatus string                 `json:"verification_status"`
	VerificationMethod string                 `json:"verification_method"`
	VerificationScore  float64                `json:"verification_score"`
	VerifiedBy         string                 `json:"verified_by,omitempty"`
	VerifiedAt         *time.Time             `json:"verified_at,omitempty"`
	ExpiryDate         *time.Time             `json:"expiry_date,omitempty"`
	ExtractedData      map[string]interface{} `json:"extracted_data"`
	ConfidenceMetrics  map[string]float64     `json:"confidence_metrics"`
	VerificationLog    []VerificationStep     `json:"verification_log"`
	FraudIndicators    []FraudIndicator       `json:"fraud_indicators"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// VerificationStep 验证步骤
type VerificationStep struct {
	StepID       string                 `json:"step_id"`
	StepName     string                 `json:"step_name"`
	StepType     string                 `json:"step_type"`
	Status       string                 `json:"status"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      *time.Time             `json:"end_time,omitempty"`
	Duration     time.Duration          `json:"duration"`
	Result       string                 `json:"result"`
	Confidence   float64                `json:"confidence"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
}

// FraudIndicator 欺诈指标
type FraudIndicator struct {
	IndicatorID   string                 `json:"indicator_id"`
	IndicatorType string                 `json:"indicator_type"`
	Description   string                 `json:"description"`
	Severity      string                 `json:"severity"`
	Confidence    float64                `json:"confidence"`
	DetectedAt    time.Time              `json:"detected_at"`
	Evidence      map[string]interface{} `json:"evidence,omitempty"`
}

// KYCVerificationRequest KYC验证请求
type KYCVerificationRequest struct {
	RequestID         string                 `json:"request_id"`
	ClientID          string                 `json:"client_id"`
	VerificationType  KYCVerificationType    `json:"verification_type"`
	PersonalInfo      PersonalInfo           `json:"personal_info"`
	Documents         []Document             `json:"documents"`
	AddressInfo       AddressInfo            `json:"address_info,omitempty"`
	ContactInfo       ContactInfo            `json:"contact_info,omitempty"`
	BusinessInfo      *BusinessInfo          `json:"business_info,omitempty"`
	SourceOfFunds     SourceOfFundsInfo      `json:"source_of_funds,omitempty"`
	PurposeOfAccount  string                 `json:"purpose_of_account,omitempty"`
	RequestedBy       string                 `json:"requested_by"`
	Priority          int                    `json:"priority"`
	ScheduledTime     *time.Time             `json:"scheduled_time,omitempty"`
	Context           map[string]interface{} `json:"context,omitempty"`
}

// AddressInfo 地址信息
type AddressInfo struct {
	CurrentAddress      Address              `json:"current_address"`
	PreviousAddresses   []Address            `json:"previous_addresses,omitempty"`
	MailingAddress      *Address             `json:"mailing_address,omitempty"`
	BusinessAddress     *Address             `json:"business_address,omitempty"`
	AddressVerification AddressVerification  `json:"address_verification"`
}

// AddressVerification 地址验证
type AddressVerification struct {
	VerificationMethod string                 `json:"verification_method"`
	VerificationStatus string                 `json:"verification_status"`
	VerifiedBy         string                 `json:"verified_by,omitempty"`
	VerifiedAt         *time.Time             `json:"verified_at,omitempty"`
	ExpiryDate         *time.Time             `json:"expiry_date,omitempty"`
	ProofDocuments     []string               `json:"proof_documents,omitempty"`
	Evidence           map[string]interface{} `json:"evidence,omitempty"`
}

// BusinessInfo 企业信息（用于企业客户）
type BusinessInfo struct {
	CompanyName        string                 `json:"company_name"`
	CompanyType        string                 `json:"company_type"`
	RegistrationNumber string                 `json:"registration_number"`
	RegistrationDate   *time.Time             `json:"registration_date,omitempty"`
	Jurisdiction       string                 `json:"jurisdiction"`
	BusinessAddress    Address               `json:"business_address"`
	BusinessNature     string                 `json:"business_nature"`
	Industry           string                 `json:"industry"`
	AnnualRevenue      float64                `json:"annual_revenue,omitempty"`
	OwnershipStructure []OwnershipInfo       `json:"ownership_structure,omitempty"`
	Directors          []DirectorInfo         `json:"directors,omitempty"`
	BeneficialOwners   []BeneficialOwnerInfo  `json:"beneficial_owners,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// OwnershipInfo 所有权信息
type OwnershipInfo struct {
	OwnerID          string    `json:"owner_id"`
	OwnerName        string    `json:"owner_name"`
	OwnerType        string    `json:"owner_type"`        // INDIVIDUAL/COMPANY
	OwnershipPercent float64   `json:"ownership_percent"`
	ControlPercent   float64   `json:"control_percent"`
	DateOfAcquisition *time.Time `json:"date_of_acquisition,omitempty"`
}

// DirectorInfo 董事信息
type DirectorInfo struct {
	DirectorID   string                 `json:"director_id"`
	Name         string                 `json:"name"`
	Position     string                 `json:"position"`
	AppointedDate *time.Time            `json:"appointed_date,omitempty"`
	IsPEP        bool                   `json:"is_pep"`
	Nationality  string                 `json:"nationality"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// BeneficialOwnerInfo 受益所有人信息
type BeneficialOwnerInfo struct {
	OwnerID         string                 `json:"owner_id"`
	Name            string                 `json:"name"`
	OwnershipPercent float64               `json:"ownership_percent"`
	ControlPercent  float64                `json:"control_percent"`
	DateOfBirth     *time.Time             `json:"date_of_birth,omitempty"`
	Nationality     string                 `json:"nationality"`
	Address         Address               `json:"address"`
	IsPEP           bool                   `json:"is_pep"`
	Documents       []Document             `json:"documents,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// SourceOfFundsInfo 资金来源信息
type SourceOfFundsInfo struct {
	SourceTypes      []string               `json:"source_types"`
	EstimatedAmount  float64                `json:"estimated_amount"`
	Currency         string                 `json:"currency"`
	SourceDescription string                 `json:"source_description"`
	Documentation    []SourceDocument      `json:"documentation"`
	Verification     SourceVerification     `json:"verification"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// SourceDocument 资金来源文档
type SourceDocument struct {
	DocumentID   string                 `json:"document_id"`
	DocumentType string                 `json:"document_type"`
	Description  string                 `json:"description"`
	FileData     string                 `json:"file_data"`     // Base64编码
	FileName     string                 `json:"file_name"`
	FileSize     int64                  `json:"file_size"`
	UploadDate   time.Time              `json:"upload_date"`
	Verified     bool                   `json:"verified"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// SourceVerification 资金来源验证
type SourceVerification struct {
	VerificationMethod string                 `json:"verification_method"`
	VerificationStatus string                 `json:"verification_status"`
	VerifiedBy         string                 `json:"verified_by,omitempty"`
	VerifiedAt         *time.Time             `json:"verified_at,omitempty"`
	ExpiryDate         *time.Time             `json:"expiry_date,omitempty"`
	Evidence           map[string]interface{} `json:"evidence,omitempty"`
}

// KYCVerificationResult KYC验证结果
type KYCVerificationResult struct {
	RequestID           string                 `json:"request_id"`
	ClientID            string                 `json:"client_id"`
	VerificationType    KYCVerificationType    `json:"verification_type"`
	VerificationStatus  VerificationStatus     `json:"verification_status"`
	OverallScore        float64                `json:"overall_score"`
	RiskLevel           RiskLevel              `json:"risk_level"`
	VerifiedFields      []string               `json:"verified_fields"`
	FailedFields        []string               `json:"failed_fields"`
	PendingFields       []string               `json:"pending_fields"`
	RiskFactors         []KYCRiskFactor        `json:"risk_factors"`
	VerificationSummary VerificationSummary    `json:"verification_summary"`
	Recommendations     []KYCRecommendation    `json:"recommendations"`
	RequiredActions     []KYCRequiredAction    `json:"required_actions"`
	VerificationExpiry  time.Time              `json:"verification_expiry"`
	NextReviewDate      time.Time              `json:"next_review_date"`
	CheckedBy           string                 `json:"checked_by"`
	ProcessingTime      time.Duration          `json:"processing_time"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// VerificationStatus 验证状态
type VerificationStatus string

const (
	StatusVerified          VerificationStatus = "VERIFIED"          // 已验证
	StatusNotVerified      VerificationStatus = "NOT_VERIFIED"      // 未验证
	StatusPending          VerificationStatus = "PENDING"          // 待验证
StatusFailed           VerificationStatus = "FAILED"           // 验证失败
StatusExpired          VerificationStatus = "EXPIRED"          // 已过期
StatusAdditionalInfo   VerificationStatus = "ADDITIONAL_INFO"  // 需要额外信息
	StatusRejected        VerificationStatus = "REJECTED"        // 已拒绝
	StatusManualReview    VerificationStatus = "MANUAL_REVIEW"    // 需要人工审核
)

// VerificationSummary 验证摘要
type VerificationSummary struct {
	TotalFields         int                    `json:"total_fields"`
	VerifiedFields      int                    `json:"verified_fields"`
	FailedFields        int                    `json:"failed_fields"`
	PendingFields       int                    `json:"pending_fields"`
	DocumentVerification DocumentVerificationSummary `json:"document_verification"`
	AddressVerification  AddressVerificationSummary   `json:"address_verification"`
	IdentityMatch       IdentityMatchSummary      `json:"identity_match"`
	RiskAssessment      RiskAssessmentSummary     `json:"risk_assessment"`
}

// DocumentVerificationSummary 文档验证摘要
type DocumentVerificationSummary struct {
	TotalDocuments    int                    `json:"total_documents"`
	VerifiedDocuments int                    `json:"verified_documents"`
	FailedDocuments   int                    `json:"failed_documents"`
	PendingDocuments  int                    `json:"pending_documents"`
	DocumentDetails   []DocumentDetail       `json:"document_details"`
	AverageScore      float64                `json:"average_score"`
	FraudIndicators   int                    `json:"fraud_indicators"`
}

// DocumentDetail 文档详情
type DocumentDetail struct {
	DocumentType       string  `json:"document_type"`
	DocumentNumber     string  `json:"document_number"`
	VerificationStatus string  `json:"verification_status"`
	VerificationScore  float64 `json:"verification_score"`
	ExpiryDate         *time.Time `json:"expiry_date,omitempty"`
	HasFraudIndicators bool    `json:"has_fraud_indicators"`
}

// AddressVerificationSummary 地址验证摘要
type AddressVerificationSummary struct {
	AddressType        string  `json:"address_type"`
	VerificationMethod string  `json:"verification_method"`
	VerificationStatus string  `json:"verification_status"`
	VerificationScore  float64 `json:"verification_score"`
	MatchedDocuments   int     `json:"matched_documents"`
	TotalDocuments     int     `json:"total_documents"`
}

// IdentityMatchSummary 身份匹配摘要
type IdentityMatchSummary struct {
	NameMatchScore     float64 `json:"name_match_score"`
	DOBMatchScore      float64 `json:"dob_match_score"`
	DocumentMatchScore float64 `json:"document_match_score"`
	BiometricMatchScore float64 `json:"biometric_match_score,omitempty"`
	OverallMatchScore  float64 `json:"overall_match_score"`
	MatchMethod        string  `json:"match_method"`
}

// RiskAssessmentSummary 风险评估摘要
type RiskAssessmentSummary struct {
	RiskScore     float64          `json:"risk_score"`
	RiskLevel     RiskLevel        `json:"risk_level"`
	RiskFactors   []string         `json:"risk_factors"`
	RiskCategory  string           `json:"risk_category"`
	LastAssessed  time.Time        `json:"last_assessed"`
	NextReview    time.Time        `json:"next_review"`
	MitigationMeasures []string    `json:"mitigation_measures"`
}

// KYCRiskFactor KYC风险因子
type KYCRiskFactor struct {
	FactorID       string                 `json:"factor_id"`
	Category       string                 `json:"category"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	RiskLevel      RiskLevel              `json:"risk_level"`
	Weight         float64                `json:"weight"`
	Score          float64                `json:"score"`
	Status         string                 `json:"status"`
	DetectedDate   time.Time              `json:"detected_date"`
	Evidence       map[string]interface{} `json:"evidence,omitempty"`
}

// KYCRecommendation KYC建议
type KYCRecommendation struct {
	RecommendationID string                 `json:"recommendation_id"`
	Priority         int                    `json:"priority"`
	Category         string                 `json:"category"`
	Title            string                 `json:"title"`
	Description      string                 `json:"description"`
	ActionItems      []string               `json:"action_items"`
	Deadline         *time.Time             `json:"deadline,omitempty"`
	AssignedTo       string                 `json:"assigned_to,omitempty"`
	Status           VerificationStatus     `json:"status"`
	Evidence         map[string]interface{} `json:"evidence,omitempty"`
}

// KYCRequiredAction KYC必要行动
type KYCRequiredAction struct {
	ActionID         string                 `json:"action_id"`
	Title            string                 `json:"title"`
	Description      string                 `json:"description"`
	Priority         int                    `json:"priority"`
	Category         string                 `json:"category"`
	Deadline         time.Time              `json:"deadline"`
	AssignedTo       string                 `json:"assigned_to"`
	Status           VerificationStatus     `json:"status"`
	Dependencies     []string               `json:"dependencies,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// KYCComplianceService KYC合规服务
type KYCComplianceService struct {
	complianceEngine   ComplianceEngine
	documentVerifier   DocumentVerifier
	addressValidator   AddressValidator
	identityMatcher    IdentityMatcher
	riskAssessor       KYCRiskAssessor
	logger             *slog.Logger
}

// NewKYCComplianceService 创建KYC合规服务
func NewKYCComplianceService(engine ComplianceEngine, docVerifier DocumentVerifier, addrValidator AddressValidator, identityMatcher IdentityMatcher, riskAssessor KYCRiskAssessor, logger *slog.Logger) *KYCComplianceService {
	return &KYCComplianceService{
		complianceEngine: engine,
		documentVerifier: docVerifier,
		addressValidator: addrValidator,
		identityMatcher:  identityMatcher,
		riskAssessor:     riskAssessor,
		logger:           logger,
	}
}

// PerformKYCVerification 执行KYC验证
func (s *KYCComplianceService) PerformKYCVerification(ctx context.Context, req *KYCVerificationRequest) (*KYCVerificationResult, error) {
	startTime := time.Now()

	s.logger.Info("开始KYC验证",
		"client_id", req.ClientID,
		"verification_type", req.VerificationType,
		"request_id", req.RequestID)

	// 1. 数据验证
	if err := s.validateRequest(req); err != nil {
		return nil, fmt.Errorf("请求验证失败: %w", err)
	}

	// 2. 文档验证
	documentSummary, err := s.verifyDocuments(ctx, req.Documents, req.VerificationType)
	if err != nil {
		s.logger.Error("文档验证失败", "error", err)
		return nil, fmt.Errorf("文档验证失败: %w", err)
	}

	// 3. 地址验证
	addressSummary, err := s.verifyAddress(ctx, req.AddressInfo, req.VerificationType)
	if err != nil {
		s.logger.Error("地址验证失败", "error", err)
		return nil, fmt.Errorf("地址验证失败: %w", err)
	}

	// 4. 身份匹配
	identitySummary, err := s.performIdentityMatching(ctx, req.PersonalInfo, req.Documents)
	if err != nil {
		s.logger.Error("身份匹配失败", "error", err)
		return nil, fmt.Errorf("身份匹配失败: %w", err)
	}

	// 5. 风险评估
	riskSummary, err := s.performRiskAssessment(ctx, req, documentSummary, addressSummary, identitySummary)
	if err != nil {
		s.logger.Error("风险评估失败", "error", err)
		return nil, fmt.Errorf("风险评估失败: %w", err)
	}

	// 6. 构建验证摘要
	verificationSummary := VerificationSummary{
		TotalFields:         s.calculateTotalFields(req),
		VerifiedFields:      documentSummary.VerifiedDocuments + addressSummary.MatchedDocuments,
		FailedFields:        documentSummary.FailedDocuments,
		PendingFields:       documentSummary.PendingDocuments,
		DocumentVerification: documentSummary,
		AddressVerification:  addressSummary,
		IdentityMatch:       identitySummary,
		RiskAssessment:      riskSummary,
	}

	// 7. 确定验证状态和评分
	verificationStatus, overallScore := s.determineVerificationStatus(verificationSummary)

	// 8. 生成建议和必要行动
	recommendations := s.generateKYCRecommendations(req, verificationSummary, riskSummary)
	requiredActions := s.generateKYCRequiredActions(req, verificationSummary, riskSummary)

	// 9. 计算到期时间和下次审查日期
	expiryDate := s.calculateVerificationExpiry(req.VerificationType, verificationStatus, riskSummary.RiskLevel)
	nextReviewDate := s.calculateNextReviewDate(req.VerificationType, riskSummary.RiskLevel)

	// 10. 构建结果
	result := &KYCVerificationResult{
		RequestID:           req.RequestID,
		ClientID:            req.ClientID,
		VerificationType:    req.VerificationType,
		VerificationStatus:  verificationStatus,
		OverallScore:        overallScore,
		RiskLevel:           riskSummary.RiskLevel,
		VerifiedFields:      s.getVerifiedFields(verificationSummary),
		FailedFields:        s.getFailedFields(verificationSummary),
		PendingFields:       s.getPendingFields(verificationSummary),
		RiskFactors:         riskSummary.RiskFactors,
		VerificationSummary: verificationSummary,
		Recommendations:     recommendations,
		RequiredActions:     requiredActions,
		VerificationExpiry:  expiryDate,
		NextReviewDate:      nextReviewDate,
		CheckedBy:           "kyc_compliance_service",
		ProcessingTime:      time.Since(startTime),
		Metadata: map[string]interface{}{
			"verification_level": string(req.VerificationType),
			"document_count":    len(req.Documents),
		},
	}

	s.logger.Info("KYC验证完成",
		"client_id", result.ClientID,
		"verification_type", result.VerificationType,
		"status", result.VerificationStatus,
		"risk_level", result.RiskLevel,
		"processing_time", result.ProcessingTime)

	return result, nil
}

// validateRequest 验证请求
func (s *KYCComplianceService) validateRequest(req *KYCVerificationRequest) error {
	if req.RequestID == "" {
		return fmt.Errorf("请求ID不能为空")
	}

	if req.ClientID == "" {
		return fmt.Errorf("客户ID不能为空")
	}

	if req.VerificationType == "" {
		return fmt.Errorf("验证类型不能为空")
	}

	if req.PersonalInfo.FirstName == "" || req.PersonalInfo.LastName == "" {
		return fmt.Errorf("客户姓名不能为空")
	}

	if len(req.Documents) == 0 {
		return fmt.Errorf("至少需要提供一份文档")
	}

	return nil
}

// verifyDocuments 验证文档
func (s *KYCComplianceService) verifyDocuments(ctx context.Context, documents []Document, verificationType KYCVerificationType) (DocumentVerificationSummary, error) {
	summary := DocumentVerificationSummary{
		TotalDocuments: len(documents),
		DocumentDetails: make([]DocumentDetail, 0, len(documents)),
	}

	totalScore := 0.0
	fraudIndicators := 0

	for _, doc := range documents {
		// 执行文档验证
		verificationResult, err := s.documentVerifier.VerifyDocument(ctx, doc, verificationType)
		if err != nil {
			s.logger.Warn("文档验证失败", "document_id", doc.DocumentID, "error", err)
			summary.FailedDocuments++
			continue
		}

		// 统计验证结果
		switch verificationResult.Status {
		case "VERIFIED":
			summary.VerifiedDocuments++
			totalScore += verificationResult.Score
		case "FAILED":
			summary.FailedDocuments++
		case "PENDING":
			summary.PendingDocuments++
		}

		// 检查欺诈指标
		if len(verificationResult.FraudIndicators) > 0 {
			fraudIndicators += len(verificationResult.FraudIndicators)
		}

		// 添加文档详情
		detail := DocumentDetail{
			DocumentType:       doc.DocumentType,
			DocumentNumber:     doc.DocumentNumber,
			VerificationStatus: verificationResult.Status,
			VerificationScore:  verificationResult.Score,
			ExpiryDate:         doc.ExpiryDate,
			HasFraudIndicators: len(verificationResult.FraudIndicators) > 0,
		}
		summary.DocumentDetails = append(summary.DocumentDetails, detail)
	}

	// 计算平均分
	if summary.VerifiedDocuments > 0 {
		summary.AverageScore = totalScore / float64(summary.VerifiedDocuments)
	}
	summary.FraudIndicators = fraudIndicators

	return summary, nil
}

// verifyAddress 验证地址
func (s *KYCComplianceService) verifyAddress(ctx context.Context, addressInfo AddressInfo, verificationType KYCVerificationType) (AddressVerificationSummary, error) {
	summary := AddressVerificationSummary{
		AddressType: addressInfo.CurrentAddress.Type,
	}

	// 执行地址验证
	verificationResult, err := s.addressValidator.VerifyAddress(ctx, addressInfo.CurrentAddress, verificationType)
	if err != nil {
		s.logger.Warn("地址验证失败", "error", err)
		summary.VerificationStatus = "FAILED"
		summary.VerificationScore = 0.0
		return summary, nil
	}

	summary.VerificationMethod = verificationResult.Method
	summary.VerificationStatus = verificationResult.Status
	summary.VerificationScore = verificationResult.Score
	summary.MatchedDocuments = verificationResult.MatchedDocuments
	summary.TotalDocuments = verificationResult.TotalDocuments

	return summary, nil
}

// performIdentityMatching 执行身份匹配
func (s *KYCComplianceService) performIdentityMatching(ctx context.Context, personalInfo PersonalInfo, documents []Document) (IdentityMatchSummary, error) {
	summary := IdentityMatchSummary{
		MatchMethod: "DOCUMENT_BASED",
	}

	// 执行身份匹配
	matchResult, err := s.identityMatcher.MatchIdentity(ctx, personalInfo, documents)
	if err != nil {
		s.logger.Warn("身份匹配失败", "error", err)
		summary.OverallMatchScore = 0.0
		return summary, nil
	}

	summary.NameMatchScore = matchResult.NameScore
	summary.DOBMatchScore = matchResult.DOBScore
	summary.DocumentMatchScore = matchResult.DocumentScore
	summary.OverallMatchScore = matchResult.OverallScore
	summary.MatchMethod = matchResult.Method

	// 如果有生物识别验证
	if matchResult.BiometricScore > 0 {
		summary.BiometricMatchScore = matchResult.BiometricScore
	}

	return summary, nil
}

// performRiskAssessment 执行风险评估
func (s *KYCComplianceService) performRiskAssessment(ctx context.Context, req *KYCVerificationRequest, docSummary DocumentVerificationSummary, addrSummary AddressVerificationSummary, identitySummary IdentityMatchSummary) (RiskAssessmentSummary, error) {
	// 构建风险评估请求
	assessmentReq := &KYCRiskAssessmentRequest{
		ClientID:          req.ClientID,
		PersonalInfo:      req.PersonalInfo,
		BusinessInfo:      req.BusinessInfo,
		VerificationType:  req.VerificationType,
		DocumentSummary:   docSummary,
		AddressSummary:    addrSummary,
		IdentitySummary:   identitySummary,
		SourceOfFunds:     req.SourceOfFunds,
		PurposeOfAccount: req.PurposeOfAccount,
	}

	// 执行风险评估
	assessmentResult, err := s.riskAssessor.AssessRisk(ctx, assessmentReq)
	if err != nil {
		s.logger.Warn("风险评估失败", "error", err)
		// 返回默认值
		return RiskAssessmentSummary{
			RiskScore:    50.0,
			RiskLevel:    RiskLevelMedium,
			RiskFactors:  []string{"风险评估失败"},
			LastAssessed: time.Now(),
			NextReview:   time.Now().AddDate(0, 6, 0), // 6个月后
		}, nil
	}

	return RiskAssessmentSummary{
		RiskScore:          assessmentResult.RiskScore,
		RiskLevel:          assessmentResult.RiskLevel,
		RiskFactors:        assessmentResult.RiskFactors,
		RiskCategory:       assessmentResult.Category,
		LastAssessed:       time.Now(),
		NextReview:         assessmentResult.NextReviewDate,
		MitigationMeasures: assessmentResult.MitigationMeasures,
	}, nil
}

// calculateTotalFields 计算总字段数
func (s *KYCComplianceService) calculateTotalFields(req *KYCVerificationRequest) int {
	// 简化实现：基于文档数量和验证类型估算
	baseFields := len(req.Documents) * 10 // 每个文档大约10个字段

	// 根据验证类型增加字段
	switch req.VerificationType {
	case KYCTypeEnhanced, KYCTypeUltimate:
		baseFields += 20
	case KYCTypeStandard:
		baseFields += 10
	}

	// 如果有企业信息
	if req.BusinessInfo != nil {
		baseFields += 15
	}

	// 如果有资金来源信息
	if req.SourceOfFunds.SourceTypes != nil {
		baseFields += 10
	}

	return baseFields
}

// determineVerificationStatus 确定验证状态
func (s *KYCComplianceService) determineVerificationStatus(summary VerificationSummary) (VerificationStatus, float64) {
	// 计算总体评分
	totalFields := summary.TotalFields
	verifiedFields := summary.VerifiedFields
	failedFields := summary.FailedFields

	if totalFields == 0 {
		return StatusFailed, 0.0
	}

	overallScore := float64(verifiedFields) / float64(totalFields) * 100

	// 检查是否有欺诈指标
	if summary.DocumentVerification.FraudIndicators > 0 {
		return StatusRejected, overallScore
	}

	// 检查是否有失败的文档
	if failedFields > 0 {
		if verifiedFields == 0 {
			return StatusFailed, overallScore
		}
		return StatusAdditionalInfo, overallScore
	}

	// 检查验证完整性
	if summary.PendingFields > 0 {
		return StatusPending, overallScore
	}

	// 根据评分确定状态
	if overallScore >= 90 {
		return StatusVerified, overallScore
	} else if overallScore >= 70 {
		return StatusManualReview, overallScore
	} else {
		return StatusAdditionalInfo, overallScore
	}
}

// generateKYCRecommendations 生成KYC建议
func (s *KYCComplianceService) generateKYCRecommendations(req *KYCVerificationRequest, summary VerificationSummary, riskSummary RiskAssessmentSummary) []KYCRecommendation {
	var recommendations []KYCRecommendation

	// 基于验证结果生成建议
	if summary.DocumentVerification.FailedDocuments > 0 {
		recommendations = append(recommendations, KYCRecommendation{
			RecommendationID: s.generateID("kyc_rec"),
			Priority:         1,
			Category:         "document_verification",
			Title:            "重新提交验证失败的文档",
			Description:      "部分文档验证失败，请重新提交清晰的文档照片",
			ActionItems:      []string{"重新拍摄文档", "确保文档清晰完整", "检查文档是否在有效期内"},
			Status:           StatusPending,
		})
	}

	if summary.AddressVerification.VerificationStatus != "VERIFIED" {
		recommendations = append(recommendations, KYCRecommendation{
			RecommendationID: s.generateID("kyc_rec"),
			Priority:         2,
			Category:         "address_verification",
			Title:            "完成地址验证",
			Description:      "地址验证需要额外信息或重新验证",
			ActionItems:      []string{"提供地址证明文件", "确认地址信息准确性"},
			Status:           StatusPending,
		})
	}

	// 基于风险等级生成建议
	switch riskSummary.RiskLevel {
	case RiskLevelHigh, RiskLevelCritical:
		recommendations = append(recommendations, KYCRecommendation{
			RecommendationID: s.generateID("kyc_rec"),
			Priority:         1,
			Category:         "risk_mitigation",
			Title:            "高风险客户 - 加强监控",
			Description:      "客户风险等级较高，需要加强监控和审查",
			ActionItems:      []string{"加强交易监控", "定期风险评估", "限制某些业务类型"},
			Status:           StatusPending,
		})
	}

	// 基于欺诈指标生成建议
	if summary.DocumentVerification.FraudIndicators > 0 {
		recommendations = append(recommendations, KYCRecommendation{
			RecommendationID: s.generateID("kyc_rec"),
			Priority:         1,
			Category:         "fraud_prevention",
			Title:            "发现欺诈指标 - 需要人工审核",
			Description:      "文档中发现了欺诈指标，需要进行人工审核",
			ActionItems:      []string{"人工审核文档", "联系客户确认信息", "考虑拒绝申请"},
			Status:           StatusPending,
		})
	}

	return recommendations
}

// generateKYCRequiredActions 生成KYC必要行动
func (s *KYCComplianceService) generateKYCRequiredActions(req *KYCVerificationRequest, summary VerificationSummary, riskSummary RiskAssessmentSummary) []KYCRequiredAction {
	var actions []KYCRequiredAction

	// 文档验证失败行动
	if summary.DocumentVerification.FailedDocuments > 0 {
		actions = append(actions, KYCRequiredAction{
			ActionID:    s.generateID("kyc_action"),
			Title:       "重新提交验证失败的文档",
			Description: fmt.Sprintf("有%d份文档验证失败，需要重新提交", summary.DocumentVerification.FailedDocuments),
			Priority:    1,
			Category:    "document_verification",
			Deadline:    time.Now().AddDate(0, 0, 7), // 7天内
			AssignedTo:  "client",
			Status:      StatusPending,
		})
	}

	// 地址验证行动
	if summary.AddressVerification.VerificationStatus != "VERIFIED" {
		actions = append(actions, KYCRequiredAction{
			ActionID:    s.generateID("kyc_action"),
			Title:       "完成地址验证",
			Description: "地址验证未通过，需要提供额外证明文件",
			Priority:    2,
			Category:    "address_verification",
			Deadline:    time.Now().AddDate(0, 0, 14), // 14天内
			AssignedTo:  "client",
			Status:      StatusPending,
		})
	}

	// 高风险客户行动
	if riskSummary.RiskLevel == RiskLevelHigh || riskSummary.RiskLevel == RiskLevelCritical {
		actions = append(actions, KYCRequiredAction{
			ActionID:    s.generateID("kyc_action"),
			Title:       "制定风险缓解计划",
			Description: "为高风险客户制定和实施风险缓解计划",
			Priority:    1,
			Category:    "risk_management",
			Deadline:    time.Now().AddDate(0, 0, 30), // 30天内
			AssignedTo:  "risk_manager",
			Status:      StatusPending,
		})
	}

	// 资金来源验证行动（如果适用）
	if req.SourceOfFunds.SourceTypes != nil && len(req.SourceOfFunds.SourceTypes) > 0 {
		actions = append(actions, KYCRequiredAction{
			ActionID:    s.generateID("kyc_action"),
			Title:       "验证资金来源",
			Description: "需要验证客户声明的资金来源",
			Priority:    2,
			Category:    "source_of_funds",
			Deadline:    time.Now().AddDate(0, 1, 0), // 1个月内
			AssignedTo:  "compliance_officer",
			Status:      StatusPending,
		})
	}

	return actions
}

// calculateVerificationExpiry 计算验证到期时间
func (s *KYCComplianceService) calculateVerificationExpiry(verificationType KYCVerificationType, status VerificationStatus, riskLevel RiskLevel) time.Time {
	// 基础有效期
	var baseExpiry time.Duration

	switch verificationType {
	case KYCTypeBasic:
		baseExpiry = time.Hour * 24 * 365   // 1年
	case KYCTypeStandard:
		baseExpiry = time.Hour * 24 * 730   // 2年
	case KYCTypeEnhanced:
		baseExpiry = time.Hour * 24 * 1095  // 3年
	case KYCTypeUltimate:
		baseExpiry = time.Hour * 24 * 1825  // 5年
	default:
		baseExpiry = time.Hour * 24 * 365   // 默认1年
	}

	// 根据状态调整
	if status == StatusVerified {
		// 已验证使用完整有效期
	} else if status == StatusManualReview || status == StatusAdditionalInfo {
		baseExpiry = time.Hour * 24 * 180 // 6个月
	} else {
		baseExpiry = time.Hour * 24 * 90   // 3个月
	}

	// 根据风险等级调整
	var multiplier float64
	switch riskLevel {
	case RiskLevelCritical:
		multiplier = 0.25
	case RiskLevelHigh:
		multiplier = 0.5
	case RiskLevelMedium:
		multiplier = 0.75
	case RiskLevelLow:
		multiplier = 1.0
	}

	adjustedExpiry := time.Duration(float64(baseExpiry) * multiplier)

	return time.Now().Add(adjustedExpiry)
}

// calculateNextReviewDate 计算下次审查日期
func (s *KYCComplianceService) calculateNextReviewDate(verificationType KYCVerificationType, riskLevel RiskLevel) time.Time {
	// 基础审查间隔
	var baseInterval time.Duration

	switch verificationType {
	case KYCTypeBasic:
		baseInterval = time.Hour * 24 * 180   // 6个月
	case KYCTypeStandard:
		baseInterval = time.Hour * 24 * 365   // 1年
	case KYCTypeEnhanced:
		baseInterval = time.Hour * 24 * 730   // 2年
	case KYCTypeUltimate:
		baseInterval = time.Hour * 24 * 1095  // 3年
	default:
		baseInterval = time.Hour * 24 * 365   // 默认1年
	}

	// 根据风险等级调整
	var multiplier float64
	switch riskLevel {
	case RiskLevelCritical:
		multiplier = 0.25
	case RiskLevelHigh:
		multiplier = 0.5
	case RiskLevelMedium:
		multiplier = 0.75
	case RiskLevelLow:
		multiplier = 1.0
	}

	adjustedInterval := time.Duration(float64(baseInterval) * multiplier)

	return time.Now().Add(adjustedInterval)
}

// 以下是辅助方法

// getVerifiedFields 获取已验证字段
func (s *KYCComplianceService) getVerifiedFields(summary VerificationSummary) []string {
	var fields []string

	// 文档验证字段
	for _, doc := range summary.DocumentVerification.DocumentDetails {
		if doc.VerificationStatus == "VERIFIED" {
			fields = append(fields, doc.DocumentType)
		}
	}

	// 地址验证字段
	if summary.AddressVerification.VerificationStatus == "VERIFIED" {
		fields = append(fields, "ADDRESS")
	}

	return fields
}

// getFailedFields 获取失败字段
func (s *KYCComplianceService) getFailedFields(summary VerificationSummary) []string {
	var fields []string

	// 文档验证失败字段
	for _, doc := range summary.DocumentVerification.DocumentDetails {
		if doc.VerificationStatus == "FAILED" {
			fields = append(fields, doc.DocumentType)
		}
	}

	// 地址验证失败字段
	if summary.AddressVerification.VerificationStatus == "FAILED" {
		fields = append(fields, "ADDRESS")
	}

	return fields
}

// getPendingFields 获取待处理字段
func (s *KYCComplianceService) getPendingFields(summary VerificationSummary) []string {
	var fields []string

	// 文档验证待处理字段
	for _, doc := range summary.DocumentVerification.DocumentDetails {
		if doc.VerificationStatus == "PENDING" {
			fields = append(fields, doc.DocumentType)
		}
	}

	// 地址验证待处理字段
	if summary.AddressVerification.VerificationStatus == "PENDING" {
		fields = append(fields, "ADDRESS")
	}

	return fields
}

// generateID 生成ID
func (s *KYCComplianceService) generateID(prefix string) string {
	data := fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:16]
}

// 以下是服务接口定义

// DocumentVerifier 文档验证器接口
type DocumentVerifier interface {
	VerifyDocument(ctx context.Context, document Document, verificationType KYCVerificationType) (*DocumentVerificationResult, error)
}

// DocumentVerificationResult 文档验证结果
type DocumentVerificationResult struct {
	Status           string               `json:"status"`
	Score            float64              `json:"score"`
	Method           string               `json:"method"`
	VerifiedBy       string               `json:"verified_by,omitempty"`
	VerifiedAt       *time.Time           `json:"verified_at,omitempty"`
	ExpiryDate       *time.Time           `json:"expiry_date,omitempty"`
	ExtractedData    map[string]interface{} `json:"extracted_data"`
	FraudIndicators []FraudIndicator      `json:"fraud_indicators"`
}

// AddressValidator 地址验证器接口
type AddressValidator interface {
	VerifyAddress(ctx context.Context, address Address, verificationType KYCVerificationType) (*AddressVerificationResult, error)
}

// AddressVerificationResult 地址验证结果
type AddressVerificationResult struct {
	Method          string `json:"method"`
	Status          string `json:"status"`
	Score           float64 `json:"score"`
	MatchedDocuments int    `json:"matched_documents"`
	TotalDocuments  int    `json:"total_documents"`
}

// IdentityMatcher 身份匹配器接口
type IdentityMatcher interface {
	MatchIdentity(ctx context.Context, personalInfo PersonalInfo, documents []Document) (*IdentityMatchResult, error)
}

// IdentityMatchResult 身份匹配结果
type IdentityMatchResult struct {
	NameScore        float64 `json:"name_score"`
	DOBScore         float64 `json:"dob_score"`
	DocumentScore    float64 `json:"document_score"`
	BiometricScore   float64 `json:"biometric_score,omitempty"`
	OverallScore     float64 `json:"overall_score"`
	Method           string  `json:"method"`
}

// KYCRiskAssessor KYC风险评估器接口
type KYCRiskAssessor interface {
	AssessRisk(ctx context.Context, req *KYCRiskAssessmentRequest) (*KYCRiskAssessmentResult, error)
}

// KYCRiskAssessmentRequest KYC风险评估请求
type KYCRiskAssessmentRequest struct {
	ClientID          string                     `json:"client_id"`
	PersonalInfo      PersonalInfo               `json:"personal_info"`
	BusinessInfo      *BusinessInfo              `json:"business_info,omitempty"`
	VerificationType  KYCVerificationType        `json:"verification_type"`
	DocumentSummary   DocumentVerificationSummary `json:"document_summary"`
	AddressSummary    AddressVerificationSummary  `json:"address_summary"`
	IdentitySummary   IdentityMatchSummary       `json:"identity_summary"`
	SourceOfFunds     SourceOfFundsInfo          `json:"source_of_funds,omitempty"`
	PurposeOfAccount  string                     `json:"purpose_of_account,omitempty"`
}

// KYCRiskAssessmentResult KYC风险评估结果
type KYCRiskAssessmentResult struct {
	RiskScore          float64       `json:"risk_score"`
	RiskLevel          RiskLevel     `json:"risk_level"`
	RiskFactors        []string      `json:"risk_factors"`
	Category           string        `json:"category"`
	NextReviewDate     time.Time     `json:"next_review_date"`
	MitigationMeasures []string      `json:"mitigation_measures"`
}