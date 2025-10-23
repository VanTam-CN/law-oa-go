package compliance

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"
)

// AMLCheckType AML检查类型
type AMLCheckType string

const (
	AMLCheckTypeOnboarding    AMLCheckType = "ONBOARDING"    // 新客户准入检查
	AMLCheckTypeTransaction   AMLCheckType = "TRANSACTION"   // 交易监控检查
	AMLCheckTypePeriodic      AMLCheckType = "PERIODIC"      // 定期审查检查
	AMLCheckTypeAdverse       AMLCheckType = "ADVERSE"       // 负面信息检查
	AMLCheckTypeEnhanced      AMLCheckType = "ENHANCED"      // 强化尽调检查
)

// ClientInformation 客户信息
type ClientInformation struct {
	ClientID           string                 `json:"client_id"`
	ClientType         string                 `json:"client_type"`           // 个人/机构
	Name               string                 `json:"name"`
	BirthDate          *time.Time             `json:"birth_date,omitempty"`
	Citizenship        []string               `json:"citizenship"`
	Residency          string                 `json:"residency"`
	Occupation         string                 `json:"occupation,omitempty"`
	Industry           string                 `json:"industry,omitempty"`
	AnnualIncome       float64                `json:"annual_income,omitempty"`
	NetWorth          float64                `json:"net_worth,omitempty"`
	PoliticallyExposed bool                   `json:"politically_exposed"`
	Sanctioned         bool                   `json:"sanctioned"`
	WatchList          []string               `json:"watch_list"`
	BeneficialOwners   []BeneficialOwner      `json:"beneficial_owners,omitempty"`
	Addresses          []Address              `json:"addresses"`
	ContactInfo        ContactInformation     `json:"contact_info"`
	RiskProfile        ClientRiskProfile      `json:"risk_profile"`
	Documents          []ClientDocument       `json:"documents"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// BeneficialOwner 受益所有人
type BeneficialOwner struct {
	OwnerID            string    `json:"owner_id"`
	Name               string    `json:"name"`
	OwnershipPercent   float64   `json:"ownership_percent"`
	Citizenship         string    `json:"citizenship"`
	PoliticallyExposed bool      `json:"politically_exposed"`
	BirthDate          *time.Time `json:"birth_date,omitempty"`
}

// Address 地址信息
type Address struct {
	Type        string    `json:"type"`        // 注册/通讯/经营
	Country     string    `json:"country"`
	Province    string    `json:"province"`
	City        string    `json:"city"`
	AddressLine string    `json:"address_line"`
	PostalCode  string    `json:"postal_code"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}

// ContactInformation 联系信息
type ContactInformation struct {
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Website  string `json:"website,omitempty"`
	Fax      string `json:"fax,omitempty"`
}

// ClientRiskProfile 客户风险档案
type ClientRiskProfile struct {
	RiskScore       float64              `json:"risk_score"`
	RiskLevel       RiskLevel            `json:"risk_level"`
	RiskFactors     []RiskFactor         `json:"risk_factors"`
	LastAssessed    time.Time            `json:"last_assessed"`
	NextReviewDate  time.Time            `json:"next_review_date"`
	ReviewHistory   []RiskAssessment     `json:"review_history"`
	MitigationMeasures []MitigationMeasure `json:"mitigation_measures"`
}

// RiskFactor 风险因子
type RiskFactor struct {
	FactorID      string                 `json:"factor_id"`
	Category      string                 `json:"category"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Weight        float64                `json:"weight"`
	Value         interface{}            `json:"value"`
	Impact        string                 `json:"impact"`      // HIGH/MEDIUM/LOW
	Status        string                 `json:"status"`      // ACTIVE/MITIGATED/RESOLVED
	DetectedDate  time.Time              `json:"detected_date"`
	LastUpdated   time.Time              `json:"last_updated"`
	Evidence      map[string]interface{} `json:"evidence,omitempty"`
}

// RiskAssessment 风险评估
type RiskAssessment struct {
	AssessmentID      string                 `json:"assessment_id"`
	PreviousScore     float64                `json:"previous_score"`
	CurrentScore      float64                `json:"current_score"`
	ScoreChange       float64                `json:"score_change"`
	AssessmentDate    time.Time              `json:"assessment_date"`
	Assessor          string                 `json:"assessor"`
	Findings          []string               `json:"findings"`
	Recommendations   []string               `json:"recommendations"`
	NextReviewDate    time.Time              `json:"next_review_date"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// MitigationMeasure 缓解措施
type MitigationMeasure struct {
	MeasureID       string                 `json:"measure_id"`
	Type            string                 `json:"type"`
	Description     string                 `json:"description"`
	Status          string                 `json:"status"`          // ACTIVE/INACTIVE/EXPIRED
	ImplementedDate time.Time              `json:"implemented_date"`
	ExpiryDate      *time.Time             `json:"expiry_date,omitempty"`
	Effectiveness   float64                `json:"effectiveness"`   // 0-1
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ClientDocument 客户文档
type ClientDocument struct {
	DocumentID      string                 `json:"document_id"`
	DocumentType    string                 `json:"document_type"`
	DocumentNumber  string                 `json:"document_number"`
	IssuingCountry  string                 `json:"issuing_country"`
	IssuingDate     *time.Time             `json:"issuing_date,omitempty"`
	ExpiryDate      *time.Time             `json:"expiry_date,omitempty"`
	VerificationStatus string               `json:"verification_status"`
	VerifiedDate    *time.Time             `json:"verified_date,omitempty"`
	VerifiedBy      string                 `json:"verified_by,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// TransactionData 交易数据
type TransactionData struct {
	TransactionID    string                 `json:"transaction_id"`
	AccountNumber    string                 `json:"account_number"`
	Amount           float64                `json:"amount"`
	Currency         string                 `json:"currency"`
	TransactionType  string                 `json:"transaction_type"`
	TransactionDate  time.Time              `json:"transaction_date"`
	Description      string                 `json:"description"`
	Counterparty     CounterpartyInfo       `json:"counterparty"`
	PaymentMethod    string                 `json:"payment_method"`
	PaymentReference string                 `json:"payment_reference,omitempty"`
	Location         string                 `json:"location,omitempty"`
	DeviceInfo       string                 `json:"device_info,omitempty"`
	IPAddress        string                 `json:"ip_address,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// CounterpartyInfo 交易对手信息
type CounterpartyInfo struct {
	CounterpartyID   string    `json:"counterparty_id"`
	Name             string    `json:"name"`
	Country          string    `json:"country"`
	AccountNumber    string    `json:"account_number,omitempty"`
	BankName         string    `json:"bank_name,omitempty"`
	Relationship     string    `json:"relationship,omitempty"`
	IsHighRisk       bool      `json:"is_high_risk"`
	Sanctioned       bool      `json:"sanctioned"`
	PoliticallyExposed bool    `json:"politically_exposed"`
}

// AMLCheckResult AML检查结果
type AMLCheckResult struct {
	CheckID            string                 `json:"check_id"`
	ClientID           string                 `json:"client_id"`
	CheckType          AMLCheckType           `json:"check_type"`
	RiskScore          float64                `json:"risk_score"`
	RiskLevel          RiskLevel              `json:"risk_level"`
	MatchedRules       []string               `json:"matched_rules"`
	SuspiciousPatterns []SuspiciousPattern    `json:"suspicious_patterns"`
	FlaggedTransactions []string             `json:"flagged_transactions"`
	WatchListHits      []WatchListHit        `json:"watch_list_hits"`
	SanctionHits       []SanctionHit         `json:"sanction_hits"`
	PEPIndicators      []PEPIndicator         `json:"pep_indicators"`
	Recommendations    []AMLRecommendation    `json:"recommendations"`
	RequiredActions    []AMLRequiredAction    `json:"required_actions"`
	CheckTimestamp     time.Time              `json:"check_timestamp"`
	NextReviewDate     time.Time              `json:"next_review_date"`
	CheckedBy          string                 `json:"checked_by"`
	ProcessingTime     time.Duration          `json:"processing_time"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// SuspiciousPattern 可疑模式
type SuspiciousPattern struct {
	PatternID     string                 `json:"pattern_id"`
	PatternType   string                 `json:"pattern_type"`
	Description   string                 `json:"description"`
	Confidence    float64                `json:"confidence"`
	Transactions  []string               `json:"transactions"`
	DetectedDate  time.Time              `json:"detected_date"`
	Evidence      map[string]interface{} `json:"evidence,omitempty"`
}

// WatchListHit 监控名单命中
type WatchListHit struct {
	HitID           string                 `json:"hit_id"`
	WatchListName   string                 `json:"watch_list_name"`
	MatchedName     string                 `json:"matched_name"`
	MatchScore      float64                `json:"match_score"`
	MatchType       string                 `json:"match_type"`      // EXACT/PARTIAL/FUZZY
	SourceInfo      string                 `json:"source_info"`
	ListType        string                 `json:"list_type"`       // SANCTION/PEP/ADVERSE_MEDIA
	ReferenceID     string                 `json:"reference_id,omitempty"`
	Confidence      float64                `json:"confidence"`
	LastUpdated     time.Time              `json:"last_updated"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// SanctionHit 制裁命中
type SanctionHit struct {
	HitID           string                 `json:"hit_id"`
	SanctionList    string                 `json:"sanction_list"`
	MatchedName     string                 `json:"matched_name"`
	MatchScore      float64                `json:"match_score"`
	SanctionType    string                 `json:"sanction_type"`
	EffectiveDate   *time.Time             `json:"effective_date,omitempty"`
	ExpiryDate      *time.Time             `json:"expiry_date,omitempty"`
	Authority       string                 `json:"authority"`
	ReferenceID     string                 `json:"reference_id,omitempty"`
	Confidence      float64                `json:"confidence"`
	LastUpdated     time.Time              `json:"last_updated"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// PEPIndicator PEP指标
type PEPIndicator struct {
	IndicatorID     string                 `json:"indicator_id"`
	PEPType         string                 `json:"pep_type"`           // DOMESTIC/FOREIGN/INTERNATIONAL
	Position        string                 `json:"position"`
	Organization    string                 `json:"organization"`
	Jurisdiction    string                 `json:"jurisdiction"`
	StartDate       *time.Time             `json:"start_date,omitempty"`
	EndDate         *time.Time             `json:"end_date,omitempty"`
	CurrentStatus   string                 `json:"current_status"`    // ACTIVE/FORMER
	Relationship    string                 `json:"relationship,omitempty"`
	Confidence      float64                `json:"confidence"`
	LastUpdated     time.Time              `json:"last_updated"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// AMLRecommendation AML建议
type AMLRecommendation struct {
	RecommendationID string                 `json:"recommendation_id"`
	Priority         int                    `json:"priority"`
	Category         string                 `json:"category"`
	Title            string                 `json:"title"`
	Description      string                 `json:"description"`
	ActionItems      []string               `json:"action_items"`
	Deadline         *time.Time             `json:"deadline,omitempty"`
	AssignedTo       string                 `json:"assigned_to,omitempty"`
	Status           ComplianceStatus       `json:"status"`
	Evidence         map[string]interface{} `json:"evidence,omitempty"`
}

// AMLRequiredAction AML必要行动
type AMLRequiredAction struct {
	ActionID         string                 `json:"action_id"`
	Title            string                 `json:"title"`
	Description      string                 `json:"description"`
	Priority         int                    `json:"priority"`
	Category         string                 `json:"category"`
	Deadline         time.Time              `json:"deadline"`
	AssignedTo       string                 `json:"assigned_to"`
	Status           ComplianceStatus       `json:"status"`
	Dependencies     []string               `json:"dependencies,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// AMLComplianceService AML合规服务
type AMLComplianceService struct {
	complianceEngine ComplianceEngine
	watchListService WatchListService
	riskCalculator   AMLRiskCalculator
	logger           *slog.Logger
}

// NewAMLComplianceService 创建AML合规服务
func NewAMLComplianceService(engine ComplianceEngine, watchListService WatchListService, riskCalculator AMLRiskCalculator, logger *slog.Logger) *AMLComplianceService {
	return &AMLComplianceService{
		complianceEngine: engine,
		watchListService: watchListService,
		riskCalculator:   riskCalculator,
		logger:           logger,
	}
}

// PerformAMLCheck 执行AML检查
func (s *AMLComplianceService) PerformAMLCheck(ctx context.Context, req *AMLCheckRequest) (*AMLCheckResult, error) {
	startTime := time.Now()

	s.logger.Info("开始AML检查",
		"client_id", req.ClientID,
		"check_type", req.CheckType,
		"request_id", req.RequestID)

	// 1. 数据验证
	if err := s.validateRequest(req); err != nil {
		return nil, fmt.Errorf("请求验证失败: %w", err)
	}

	// 2. 准备检查数据
	checkData := s.prepareCheckData(req)

	// 3. 执行合规引擎检查
	complianceReq := &ComplianceCheckRequest{
		RequestID:   req.RequestID,
		CheckType:   "aml_" + string(req.CheckType),
		SubjectID:   req.ClientID,
		SubjectType: "client",
		Data:        checkData,
		Priority:    s.getCheckPriority(req.CheckType),
		RequestedBy: req.RequestedBy,
	}

	complianceResult, err := s.complianceEngine.PerformCheck(ctx, complianceReq)
	if err != nil {
		return nil, fmt.Errorf("执行合规检查失败: %w", err)
	}

	// 4. 执行AML特定检查
	amlSpecificResult, err := s.performAMLSpecificChecks(ctx, req)
	if err != nil {
		s.logger.Error("AML特定检查失败",
			"client_id", req.ClientID,
			"error", err)
		// 不返回错误，继续处理
		amlSpecificResult = &AMLSpecificResult{}
	}

	// 5. 计算AML风险评分
	riskScore, riskLevel := s.calculateAMLRisk(ctx, req, complianceResult, amlSpecificResult)

	// 6. 生成AML特定建议和行动
	amlRecommendations := s.generateAMLRecommendations(req, complianceResult, amlSpecificResult)
	amlRequiredActions := s.generateAMLRequiredActions(req, complianceResult, amlSpecificResult, riskLevel)

	// 7. 计算下次审查日期
	nextReviewDate := s.calculateNextReviewDate(req.CheckType, riskLevel)

	// 8. 构建结果
	result := &AMLCheckResult{
		CheckID:            req.RequestID,
		ClientID:           req.ClientID,
		CheckType:          req.CheckType,
		RiskScore:          riskScore,
		RiskLevel:          riskLevel,
		MatchedRules:       s.extractMatchedRules(complianceResult),
		SuspiciousPatterns: amlSpecificResult.SuspiciousPatterns,
		FlaggedTransactions: amlSpecificResult.FlaggedTransactions,
		WatchListHits:      amlSpecificResult.WatchListHits,
		SanctionHits:       amlSpecificResult.SanctionHits,
		PEPIndicators:      amlSpecificResult.PEPIndicators,
		Recommendations:    amlRecommendations,
		RequiredActions:    amlRequiredActions,
		CheckTimestamp:     time.Now(),
		NextReviewDate:     nextReviewDate,
		CheckedBy:          "aml_compliance_service",
		ProcessingTime:     time.Since(startTime),
		Metadata: map[string]interface{}{
			"compliance_request_id": complianceResult.RequestID,
			"rules_evaluated":       complianceResult.Metadata["rules_executed"],
		},
	}

	s.logger.Info("AML检查完成",
		"client_id", result.ClientID,
		"check_type", result.CheckType,
		"risk_score", result.RiskScore,
		"risk_level", result.RiskLevel,
		"processing_time", result.ProcessingTime)

	return result, nil
}

// AMLCheckRequest AML检查请求
type AMLCheckRequest struct {
	RequestID       string            `json:"request_id"`
	ClientID        string            `json:"client_id"`
	CheckType       AMLCheckType      `json:"check_type"`
	ClientInfo      ClientInformation `json:"client_info"`
	TransactionData []TransactionData `json:"transaction_data,omitempty"`
	HistoricalData  []string          `json:"historical_data,omitempty"` // 历史检查ID
	RequestedBy     string            `json:"requested_by"`
	ScheduledTime   *time.Time        `json:"scheduled_time,omitempty"`
}

// AMLSpecificResult AML特定检查结果
type AMLSpecificResult struct {
	SuspiciousPatterns  []SuspiciousPattern `json:"suspicious_patterns"`
	FlaggedTransactions []string           `json:"flagged_transactions"`
	WatchListHits       []WatchListHit     `json:"watch_list_hits"`
	SanctionHits        []SanctionHit      `json:"sanction_hits"`
	PEPIndicators       []PEPIndicator     `json:"pep_indicators"`
}

// validateRequest 验证请求
func (s *AMLComplianceService) validateRequest(req *AMLCheckRequest) error {
	if req.RequestID == "" {
		return fmt.Errorf("请求ID不能为空")
	}

	if req.ClientID == "" {
		return fmt.Errorf("客户ID不能为空")
	}

	if req.CheckType == "" {
		return fmt.Errorf("检查类型不能为空")
	}

	if req.ClientInfo.Name == "" {
		return fmt.Errorf("客户姓名不能为空")
	}

	return nil
}

// prepareCheckData 准备检查数据
func (s *AMLComplianceService) prepareCheckData(req *AMLCheckRequest) map[string]interface{} {
	data := make(map[string]interface{})

	// 客户基本信息
	data["client_id"] = req.ClientInfo.ClientID
	data["client_type"] = req.ClientInfo.ClientType
	data["name"] = req.ClientInfo.Name
	data["citizenship"] = req.ClientInfo.Citizenship
	data["residency"] = req.ClientInfo.Residency
	data["politically_exposed"] = req.ClientInfo.PoliticallyExposed
	data["sanctioned"] = req.ClientInfo.Sanctioned
	data["watch_list"] = req.ClientInfo.WatchList

	// 财务信息
	if req.ClientInfo.AnnualIncome > 0 {
		data["annual_income"] = req.ClientInfo.AnnualIncome
	}
	if req.ClientInfo.NetWorth > 0 {
		data["net_worth"] = req.ClientInfo.NetWorth
	}

	// 职业和行业
	if req.ClientInfo.Occupation != "" {
		data["occupation"] = req.ClientInfo.Occupation
	}
	if req.ClientInfo.Industry != "" {
		data["industry"] = req.ClientInfo.Industry
	}

	// 交易数据
	if len(req.TransactionData) > 0 {
		data["transaction_count"] = len(req.TransactionData)
		data["total_amount"] = s.calculateTotalAmount(req.TransactionData)
		data["high_value_transactions"] = s.countHighValueTransactions(req.TransactionData)
		data["international_transactions"] = s.countInternationalTransactions(req.TransactionData)
	}

	// 检查类型特定数据
	switch req.CheckType {
	case AMLCheckTypeTransaction:
		data["recent_transactions"] = s.getRecentTransactions(req.TransactionData, 30) // 最近30天
	case AMLCheckTypePeriodic:
		data["last_check_date"] = s.getLastCheckDate(req.HistoricalData)
	case AMLCheckTypeAdverse:
		data["adverse_media_search"] = true
	}

	return data
}

// getCheckPriority 获取检查优先级
func (s *AMLComplianceService) getCheckPriority(checkType AMLCheckType) int {
	switch checkType {
	case AMLCheckTypeOnboarding:
		return 1 // 最高优先级
	case AMLCheckTypeAdverse:
		return 2
	case AMLCheckTypeTransaction:
		return 3
	case AMLCheckTypeEnhanced:
		return 4
	case AMLCheckTypePeriodic:
		return 5
	default:
		return 10
	}
}

// performAMLSpecificChecks 执行AML特定检查
func (s *AMLComplianceService) performAMLSpecificChecks(ctx context.Context, req *AMLCheckRequest) (*AMLSpecificResult, error) {
	result := &AMLSpecificResult{}

	// 1. 监控名单检查
	if watchListHits, err := s.performWatchListCheck(ctx, req); err != nil {
		s.logger.Warn("监控名单检查失败", "error", err)
	} else {
		result.WatchListHits = watchListHits
	}

	// 2. 制裁名单检查
	if sanctionHits, err := s.performSanctionCheck(ctx, req); err != nil {
		s.logger.Warn("制裁名单检查失败", "error", err)
	} else {
		result.SanctionHits = sanctionHits
	}

	// 3. PEP检查
	if pepIndicators, err := s.performPEPCheck(ctx, req); err != nil {
		s.logger.Warn("PEP检查失败", "error", err)
	} else {
		result.PEPIndicators = pepIndicators
	}

	// 4. 可疑模式检查（针对交易检查）
	if req.CheckType == AMLCheckTypeTransaction {
		if suspiciousPatterns, err := s.performSuspiciousPatternCheck(ctx, req); err != nil {
			s.logger.Warn("可疑模式检查失败", "error", err)
		} else {
			result.SuspiciousPatterns = suspiciousPatterns
		}

		// 标记可疑交易
		flaggedTransactions := s.identifyFlaggedTransactions(req.TransactionData, result.SuspiciousPatterns)
		result.FlaggedTransactions = flaggedTransactions
	}

	return result, nil
}

// performWatchListCheck 执行监控名单检查
func (s *AMLComplianceService) performWatchListCheck(ctx context.Context, req *AMLCheckRequest) ([]WatchListHit, error) {
	if s.watchListService == nil {
		return nil, fmt.Errorf("监控名单服务未配置")
	}

	// 构建搜索实体
	searchEntity := &WatchListSearchEntity{
		Name:         req.ClientInfo.Name,
		DateOfBirth:  req.ClientInfo.BirthDate,
		Citizenship:  req.ClientInfo.Citizenship,
		Residency:    req.ClientInfo.Residency,
		DocumentInfo: s.extractDocumentInfo(req.ClientInfo.Documents),
	}

	// 执行搜索
	hits, err := s.watchListService.Search(ctx, searchEntity)
	if err != nil {
		return nil, fmt.Errorf("监控名单搜索失败: %w", err)
	}

	// 过滤和评分
	filteredHits := s.filterAndScoreWatchListHits(hits, req.ClientInfo)

	return filteredHits, nil
}

// performSanctionCheck 执行制裁检查
func (s *AMLComplianceService) performSanctionCheck(ctx context.Context, req *AMLCheckRequest) ([]SanctionHit, error) {
	if s.watchListService == nil {
		return nil, fmt.Errorf("制裁名单服务未配置")
	}

	// 使用专门的制裁名单检查
	searchEntity := &WatchListSearchEntity{
		Name:         req.ClientInfo.Name,
		DateOfBirth:  req.ClientInfo.BirthDate,
		Citizenship:  req.ClientInfo.Citizenship,
		Residency:    req.ClientInfo.Residency,
		DocumentInfo: s.extractDocumentInfo(req.ClientInfo.Documents),
	}

	sanctionHits, err := s.watchListService.SearchSanctions(ctx, searchEntity)
	if err != nil {
		return nil, fmt.Errorf("制裁名单搜索失败: %w", err)
	}

	return sanctionHits, nil
}

// performPEPCheck 执行PEP检查
func (s *AMLComplianceService) performPEPCheck(ctx context.Context, req *AMLCheckRequest) ([]PEPIndicator, error) {
	if s.watchListService == nil {
		return nil, fmt.Errorf("PEP检查服务未配置")
	}

	// 如果已经标记为PEP，直接返回
	if req.ClientInfo.PoliticallyExposed {
		return []PEPIndicator{
			{
				IndicatorID:   fmt.Sprintf("pep_%s", req.ClientID),
				PEPType:       "KNOWN",
				Position:      req.ClientInfo.Occupation,
				CurrentStatus: "ACTIVE",
				Confidence:    1.0,
				LastUpdated:   time.Now(),
			},
		}, nil
	}

	// 搜索PEP名单
	searchEntity := &WatchListSearchEntity{
		Name:         req.ClientInfo.Name,
		DateOfBirth:  req.ClientInfo.BirthDate,
		Citizenship:  req.ClientInfo.Citizenship,
		Occupation:   req.ClientInfo.Occupation,
	}

	pepIndicators, err := s.watchListService.SearchPEP(ctx, searchEntity)
	if err != nil {
		return nil, fmt.Errorf("PEP搜索失败: %w", err)
	}

	return pepIndicators, nil
}

// performSuspiciousPatternCheck 执行可疑模式检查
func (s *AMLComplianceService) performSuspiciousPatternCheck(ctx context.Context, req *AMLCheckRequest) ([]SuspiciousPattern, error) {
	var patterns []SuspiciousPattern

	// 1. 检查大额交易模式
	if largeValuePattern := s.detectLargeValuePattern(req.TransactionData); largeValuePattern != nil {
		patterns = append(patterns, *largeValuePattern)
	}

	// 2. 检查结构化交易模式
	if structuringPattern := s.detectStructuringPattern(req.TransactionData); structuringPattern != nil {
		patterns = append(patterns, *structuringPattern)
	}

	// 3. 检查异常时间模式
	if timingPattern := s.detectUnusualTimingPattern(req.TransactionData); timingPattern != nil {
		patterns = append(patterns, *timingPattern)
	}

	// 4. 检查高风险地区交易模式
	if highRiskPattern := s.detectHighRiskGeographicPattern(req.TransactionData); highRiskPattern != nil {
		patterns = append(patterns, *highRiskPattern)
	}

	// 5. 检查频率异常模式
	if frequencyPattern := s.detectUnusualFrequencyPattern(req.TransactionData); frequencyPattern != nil {
		patterns = append(patterns, *frequencyPattern)
	}

	return patterns, nil
}

// calculateAMLRisk 计算AML风险评分
func (s *AMLComplianceService) calculateAMLRisk(ctx context.Context, req *AMLCheckRequest, complianceResult *ComplianceCheckResult, amlSpecificResult *AMLSpecificResult) (float64, RiskLevel) {
	baseScore := complianceResult.OverallScore

	// 根据AML特定发现调整评分
	amlAdjustment := 0.0

	// 监控名单命中调整
	if len(amlSpecificResult.WatchListHits) > 0 {
		amlAdjustment += float64(len(amlSpecificResult.WatchListHits)) * 15
	}

	// 制裁命中调整
	if len(amlSpecificResult.SanctionHits) > 0 {
		amlAdjustment += float64(len(amlSpecificResult.SanctionHits)) * 25
	}

	// PEP指示器调整
	if len(amlSpecificResult.PEPIndicators) > 0 {
		amlAdjustment += float64(len(amlSpecificResult.PEPIndicators)) * 20
	}

	// 可疑模式调整
	if len(amlSpecificResult.SuspiciousPatterns) > 0 {
		amlAdjustment += float64(len(amlSpecificResult.SuspiciousPatterns)) * 10
	}

	// 计算最终评分
	finalScore := baseScore + amlAdjustment

	// 确保评分在合理范围内
	if finalScore < 0 {
		finalScore = 0
	} else if finalScore > 100 {
		finalScore = 100
	}

	// 确定风险等级
	riskLevel := s.determineRiskLevelFromScore(finalScore, len(amlSpecificResult.SanctionHits) > 0, len(amlSpecificResult.PEPIndicators) > 0)

	return finalScore, riskLevel
}

// determineRiskLevelFromScore 根据评分确定风险等级
func (s *AMLComplianceService) determineRiskLevelFromScore(score float64, hasSanctions bool, hasPEP bool) RiskLevel {
	// 制裁命中或PEP直接设为高风险
	if hasSanctions {
		return RiskLevelCritical
	}

	if hasPEP {
		return RiskLevelHigh
	}

	// 根据评分确定等级
	if score >= 80 {
		return RiskLevelHigh
	} else if score >= 60 {
		return RiskLevelMedium
	} else if score >= 40 {
		return RiskLevelLow
	} else {
		return RiskLevelLow
	}
}

// generateAMLRecommendations 生成AML建议
func (s *AMLComplianceService) generateAMLRecommendations(req *AMLCheckRequest, complianceResult *ComplianceCheckResult, amlSpecificResult *AMLSpecificResult) []AMLRecommendation {
	var recommendations []AMLRecommendation

	// 基于合规引擎结果生成建议
	for _, rec := range complianceResult.Recommendations {
		amlRec := AMLRecommendation{
			RecommendationID: rec.RecommendationID,
			Priority:         rec.Priority,
			Category:         "compliance_engine",
			Title:            rec.Title,
			Description:      rec.Description,
			ActionItems:      rec.ActionItems,
			AssignedTo:       rec.AssignedTo,
			Status:           rec.Status,
		}

		if rec.Deadline != nil {
			amlRec.Deadline = rec.Deadline
		}

		recommendations = append(recommendations, amlRec)
	}

	// AML特定建议
	if len(amlSpecificResult.SanctionHits) > 0 {
		recommendations = append(recommendations, AMLRecommendation{
			RecommendationID: s.generateID("aml_rec"),
			Priority:         1,
			Category:         "sanction_hit",
			Title:            "制裁名单命中 - 立即采取行动",
			Description:      "客户出现在制裁名单上，需要立即采取相应措施",
			ActionItems:      []string{"冻结账户", "报告监管机构", "进行内部调查"},
			Deadline:         &time.Time{},
			Status:           StatusPending,
		})
		*recommendations[len(recommendations)-1].Deadline = time.Now().AddDate(0, 0, 1) // 1天内
	}

	if len(amlSpecificResult.PEPIndicators) > 0 {
		recommendations = append(recommendations, AMLRecommendation{
			RecommendationID: s.generateID("aml_rec"),
			Priority:         2,
			Category:         "pep_identified",
			Title:            "政治公众人物识别 - 强化尽职调查",
			Description:      "客户被识别为政治公众人物，需要进行强化尽职调查",
			ActionItems:      []string{"获取高级管理层批准", "加强持续监控", "定期审查"},
			Deadline:         &time.Time{},
			Status:           StatusPending,
		})
		*recommendations[len(recommendations)-1].Deadline = time.Now().AddDate(0, 0, 7) // 7天内
	}

	if len(amlSpecificResult.SuspiciousPattern) > 0 {
		recommendations = append(recommendations, AMLRecommendation{
			RecommendationID: s.generateID("aml_rec"),
			Priority:         3,
			Category:         "suspicious_activity",
			Title:            "可疑交易模式识别 - 进一步调查",
			Description:      "检测到可疑交易模式，需要进行进一步调查",
			ActionItems:      []string{"分析交易背景", "评估业务合理性", "考虑报告可疑交易"},
			Status:           StatusPending,
		})
	}

	return recommendations
}

// generateAMLRequiredActions 生成AML必要行动
func (s *AMLComplianceService) generateAMLRequiredActions(req *AMLCheckRequest, complianceResult *ComplianceCheckResult, amlSpecificResult *AMLSpecificResult, riskLevel RiskLevel) []AMLRequiredAction {
	var actions []AMLRequiredAction

	// 制裁命中行动
	if len(amlSpecificResult.SanctionHits) > 0 {
		actions = append(actions, AMLRequiredAction{
			ActionID:    s.generateID("aml_action"),
			Title:       "立即冻结账户",
			Description: "根据制裁名单命中结果，立即冻结客户账户",
			Priority:    1,
			Category:    "sanction_compliance",
			Deadline:    time.Now().AddDate(0, 0, 1), // 1天内
			AssignedTo:  "compliance_officer",
			Status:      StatusPending,
		})
	}

	// PEP相关行动
	if len(amlSpecificResult.PEPIndicators) > 0 {
		actions = append(actions, AMLRequiredAction{
			ActionID:    s.generateID("aml_action"),
			Title:       "获取高级管理层批准",
			Description: "PEP客户需要获得高级管理层的批准才能建立业务关系",
			Priority:    2,
			Category:    "pep_compliance",
			Deadline:    time.Now().AddDate(0, 0, 7), // 7天内
			AssignedTo:  "senior_management",
			Status:      StatusPending,
		})
	}

	// 风险评估行动
	if riskLevel == RiskLevelHigh || riskLevel == RiskLevelCritical {
		actions = append(actions, AMLRequiredAction{
			ActionID:    s.generateID("aml_action"),
			Title:       "制定风险缓解措施",
			Description: "为高风险客户制定并实施适当的风险缓解措施",
			Priority:    3,
			Category:    "risk_mitigation",
			Deadline:    time.Now().AddDate(0, 0, 14), // 14天内
			AssignedTo:  "risk_manager",
			Status:      StatusPending,
		})
	}

	// 持续监控行动
	if req.CheckType == AMLCheckTypeOnboarding {
		actions = append(actions, AMLRequiredAction{
			ActionID:    s.generateID("aml_action"),
			Title:       "建立持续监控机制",
			Description: "为客户建立适当的持续监控机制",
			Priority:    4,
			Category:    "ongoing_monitoring",
			Deadline:    time.Now().AddDate(0, 1, 0), // 1个月内
			AssignedTo:  "monitoring_team",
			Status:      StatusPending,
		})
	}

	return actions
}

// calculateNextReviewDate 计算下次审查日期
func (s *AMLComplianceService) calculateNextReviewDate(checkType AMLCheckType, riskLevel RiskLevel) time.Time {
	var baseInterval time.Duration

	// 根据检查类型确定基础间隔
	switch checkType {
	case AMLCheckTypeOnboarding:
		baseInterval = time.Hour * 24 * 30 // 30天
	case AMLCheckTypeTransaction:
		baseInterval = time.Hour * 24 * 7 // 7天
	case AMLCheckTypePeriodic:
		baseInterval = time.Hour * 24 * 365 // 1年
	case AMLCheckTypeAdverse:
		baseInterval = time.Hour * 24 * 14 // 14天
	case AMLCheckTypeEnhanced:
		baseInterval = time.Hour * 24 * 90 // 90天
	default:
		baseInterval = time.Hour * 24 * 180 // 180天
	}

	// 根据风险等级调整间隔
	var multiplier float64
	switch riskLevel {
	case RiskLevelCritical:
		multiplier = 0.25
	case RiskLevelHigh:
		multiplier = 0.5
	case RiskLevelMedium:
		multiplier = 1.0
	case RiskLevelLow:
		multiplier = 2.0
	}

	adjustedInterval := time.Duration(float64(baseInterval) * multiplier)

	return time.Now().Add(adjustedInterval)
}

// 以下是辅助方法

// calculateTotalAmount 计算交易总金额
func (s *AMLComplianceService) calculateTotalAmount(transactions []TransactionData) float64 {
	var total float64
	for _, tx := range transactions {
		total += math.Abs(tx.Amount)
	}
	return total
}

// countHighValueTransactions 统计大额交易
func (s *AMLComplianceService) countHighValueTransactions(transactions []TransactionData) int {
	count := 0
	for _, tx := range transactions {
		if math.Abs(tx.Amount) >= 10000 { // 大额交易阈值
			count++
		}
	}
	return count
}

// countInternationalTransactions 统计跨境交易
func (s *AMLComplianceService) countInternationalTransactions(transactions []TransactionData) int {
	count := 0
	for _, tx := range transactions {
		if tx.Counterparty.Country != "CN" { // 假设中国为境内
			count++
		}
	}
	return count
}

// getRecentTransactions 获取最近交易
func (s *AMLComplianceService) getRecentTransactions(transactions []TransactionData, days int) []TransactionData {
	cutoff := time.Now().AddDate(0, 0, -days)
	var recent []TransactionData

	for _, tx := range transactions {
		if tx.TransactionDate.After(cutoff) {
			recent = append(recent, tx)
		}
	}

	return recent
}

// getLastCheckDate 获取上次检查日期
func (s *AMLComplianceService) getLastCheckDate(historicalData []string) *time.Time {
	if len(historicalData) == 0 {
		return nil
	}

	// 简化实现：返回当前时间减去30天
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	return &thirtyDaysAgo
}

// extractDocumentInfo 提取文档信息
func (s *AMLComplianceService) extractDocumentInfo(documents []ClientDocument) map[string]interface{} {
	info := make(map[string]interface{})

	for _, doc := range documents {
		info[doc.DocumentType] = map[string]interface{}{
			"number":      doc.DocumentNumber,
			"country":     doc.IssuingCountry,
			"expiry_date": doc.ExpiryDate,
		}
	}

	return info
}

// filterAndScoreWatchListHits 过滤和评分监控名单命中
func (s *AMLComplianceService) filterAndScoreWatchListHits(hits []WatchListHit, clientInfo ClientInformation) []WatchListHit {
	var filtered []WatchListHit

	for _, hit := range hits {
		// 只保留高置信度的命中
		if hit.Confidence >= 0.7 {
			filtered = append(filtered, hit)
		}
	}

	return filtered
}

// 以下是可疑模式检测方法

// detectLargeValuePattern 检测大额交易模式
func (s *AMLComplianceService) detectLargeValuePattern(transactions []TransactionData) *SuspiciousPattern {
	threshold := 50000.0 // 大额交易阈值
	count := 0
	var transactionIDs []string

	for _, tx := range transactions {
		if math.Abs(tx.Amount) >= threshold {
			count++
			transactionIDs = append(transactionIDs, tx.TransactionID)
		}
	}

	if count >= 3 { // 3笔以上大额交易
		return &SuspiciousPattern{
			PatternID:    "large_value",
			PatternType:  "LARGE_VALUE_TRANSACTIONS",
			Description:  "检测到多笔大额交易",
			Confidence:   0.8,
			Transactions: transactionIDs,
			DetectedDate: time.Now(),
			Evidence: map[string]interface{}{
				"count":     count,
				"threshold": threshold,
			},
		}
	}

	return nil
}

// detectStructuringPattern 检测结构化交易模式
func (s *AMLComplianceService) detectStructuringPattern(transactions []TransactionData) *SuspiciousPattern {
	// 简化实现：检查是否有多笔接近报告阈值的交易
	reportingThreshold := 10000.0
	thresholdRange := 1000.0 // 阈值范围

	var structuredTxs []string
	for _, tx := range transactions {
		amount := math.Abs(tx.Amount)
		if amount >= reportingThreshold-thresholdRange && amount <= reportingThreshold+thresholdRange {
			structuredTxs = append(structuredTxs, tx.TransactionID)
		}
	}

	if len(structuredTxs) >= 5 { // 5笔以上接近阈值的交易
		return &SuspiciousPattern{
			PatternID:    "structuring",
			PatternType:  "TRANSACTION_STRUCTURING",
			Description:  "检测到疑似结构化交易模式，多笔交易接近报告阈值",
			Confidence:   0.7,
			Transactions: structuredTxs,
			DetectedDate: time.Now(),
			Evidence: map[string]interface{}{
				"count":           len(structuredTxs),
				"threshold":       reportingThreshold,
				"threshold_range": thresholdRange,
			},
		}
	}

	return nil
}

// detectUnusualTimingPattern 检测异常时间模式
func (s *AMLComplianceService) detectUnusualTimingPattern(transactions []TransactionData) *SuspiciousPattern {
	// 简化实现：检查非工作时间交易
	var unusualTxs []string

	for _, tx := range transactions {
		hour := tx.TransactionDate.Hour()
		if hour < 9 || hour > 17 { // 非工作时间
			unusualTxs = append(unusualTxs, tx.TransactionID)
		}
	}

	if len(unusualTxs) >= 3 { // 3笔以上非工作时间交易
		return &SuspiciousPattern{
			PatternID:    "unusual_timing",
			PatternType:  "UNUSUAL_TIMING",
			Description:  "检测到异常时间交易模式，多笔交易发生在非工作时间",
			Confidence:   0.6,
			Transactions: unusualTxs,
			DetectedDate: time.Now(),
			Evidence: map[string]interface{}{
				"count": len(unusualTxs),
			},
		}
	}

	return nil
}

// detectHighRiskGeographicPattern 检测高风险地区交易模式
func (s *AMLComplianceService) detectHighRiskGeographicPattern(transactions []TransactionData) *SuspiciousPattern {
	// 高风险地区列表
	highRiskCountries := []string{"AF", "IR", "KP", "SY", "MM"}

	var highRiskTxs []string
	for _, tx := range transactions {
		for _, country := range highRiskCountries {
			if tx.Counterparty.Country == country {
				highRiskTxs = append(highRiskTxs, tx.TransactionID)
				break
			}
		}
	}

	if len(highRiskTxs) > 0 {
		return &SuspiciousPattern{
			PatternID:    "high_risk_geographic",
			PatternType:  "HIGH_RISK_GEOGRAPHIC",
			Description:  "检测到与高风险地区的交易",
			Confidence:   0.8,
			Transactions: highRiskTxs,
			DetectedDate: time.Now(),
			Evidence: map[string]interface{}{
				"count":     len(highRiskTxs),
				"countries": highRiskCountries,
			},
		}
	}

	return nil
}

// detectUnusualFrequencyPattern 检测异常频率模式
func (s *AMLComplianceService) detectUnusualFrequencyPattern(transactions []TransactionData) *SuspiciousPattern {
	// 简化实现：检查24小时内超过阈值的交易数量
	hourlyThreshold := 50
	transactionMap := make(map[int]int) // 小时 -> 交易数量

	for _, tx := range transactions {
		hour := tx.TransactionDate.Hour()
		transactionMap[hour]++
	}

	var frequentHours []int
	for hour, count := range transactionMap {
		if count > hourlyThreshold {
			frequentHours = append(frequentHours, hour)
		}
	}

	if len(frequentHours) > 0 {
		var transactionIDs []string
		for _, tx := range transactions {
			for _, hour := range frequentHours {
				if tx.TransactionDate.Hour() == hour {
					transactionIDs = append(transactionIDs, tx.TransactionID)
					break
				}
			}
		}

		return &SuspiciousPattern{
			PatternID:    "unusual_frequency",
			PatternType:  "UNUSUAL_FREQUENCY",
			Description:  "检测到异常交易频率模式",
			Confidence:   0.7,
			Transactions: transactionIDs,
			DetectedDate: time.Now(),
			Evidence: map[string]interface{}{
				"threshold":      hourlyThreshold,
				"affected_hours": frequentHours,
			},
		}
	}

	return nil
}

// identifyFlaggedTransactions 标识可疑交易
func (s *AMLComplianceService) identifyFlaggedTransactions(transactions []TransactionData, patterns []SuspiciousPattern) []string {
	flaggedSet := make(map[string]bool)

	// 将模式中的交易添加到标记集合
	for _, pattern := range patterns {
		for _, txID := range pattern.Transactions {
			flaggedSet[txID] = true
		}
	}

	// 转换为切片
	var flagged []string
	for txID := range flaggedSet {
		flagged = append(flagged, txID)
	}

	return flagged
}

// extractMatchedRules 提取匹配的规则
func (s *AMLComplianceService) extractMatchedRules(result *ComplianceCheckResult) []string {
	var matchedRules []string

	// 简化实现：从元数据中提取匹配的规则
	if result.Metadata != nil {
		if rules, exists := result.Metadata["matched_rules"]; exists {
			if ruleList, ok := rules.([]string); ok {
				matchedRules = ruleList
			}
		}
	}

	return matchedRules
}

// generateID 生成ID
func (s *AMLComplianceService) generateID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

// WatchListService 监控名单服务接口
type WatchListService interface {
	Search(ctx context.Context, entity *WatchListSearchEntity) ([]WatchListHit, error)
	SearchSanctions(ctx context.Context, entity *WatchListSearchEntity) ([]SanctionHit, error)
	SearchPEP(ctx context.Context, entity *WatchListSearchEntity) ([]PEPIndicator, error)
}

// WatchListSearchEntity 监控名单搜索实体
type WatchListSearchEntity struct {
	Name         string                 `json:"name"`
	DateOfBirth  *time.Time             `json:"date_of_birth,omitempty"`
	Citizenship  []string               `json:"citizenship,omitempty"`
	Residency    string                 `json:"residency,omitempty"`
	Occupation   string                 `json:"occupation,omitempty"`
	DocumentInfo map[string]interface{} `json:"document_info,omitempty"`
}

// AMLRiskCalculator AML风险计算器接口
type AMLRiskCalculator interface {
	CalculateRisk(ctx context.Context, client *ClientInformation, transactions []TransactionData) (float64, RiskLevel, error)
}