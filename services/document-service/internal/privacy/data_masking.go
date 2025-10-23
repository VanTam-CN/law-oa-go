package privacy

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"time"
)

// DataMaskingService 数据脱敏服务
type DataMaskingService struct {
	logger      *slog.Logger
	config      *DataMaskingConfig
	maskers     map[string]DataMasker
	auditLogger *AuditLogger
	cache       *MaskingCache
	mu          sync.RWMutex
}

// DataMaskingConfig 数据脱敏配置
type DataMaskingConfig struct {
	EnableCache      bool          `json:"enable_cache"`
	CacheTTL         time.Duration `json:"cache_ttl"`
	EnableAudit      bool          `json:"enable_audit"`
	DefaultStrategy  string        `json:"default_strategy"`
	EncryptionKey    string        `json:"encryption_key"`
	HashSalt         string        `json:"hash_salt"`
	PerformanceMode  bool          `json:"performance_mode"`
	MaxConcurrency   int           `json:"max_concurrency"`
}

// DataMasker 数据脱敏接口
type DataMasker interface {
	Mask(data string, config MaskingConfig) (string, error)
	GetType() string
}

// MaskingConfig 脱敏配置
type MaskingConfig struct {
	Type      string                 `json:"type"`
	Strategy  string                 `json:"strategy"`
	Params    map[string]interface{} `json:"params"`
	Context   map[string]string     `json:"context"`
	UserLevel string                 `json:"user_level"`
	Purpose   string                 `json:"purpose"`
}

// MaskingRequest 脱敏请求
type MaskingRequest struct {
	Data      string                 `json:"data"`
	Fields    []FieldConfig          `json:"fields"`
	Context   map[string]string     `json:"context"`
	UserInfo  *UserInfo              `json:"user_info"`
	Metadata  map[string]interface{} `json:"metadata"`
	RequestID string                 `json:"request_id"`
	Timestamp time.Time              `json:"timestamp"`
}

// FieldConfig 字段配置
type FieldConfig struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Value       string                 `json:"value"`
	Required    bool                   `json:"required"`
	Config      map[string]interface{} `json:"config"`
}

// UserInfo 用户信息
type UserInfo struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
	Department  string   `json:"department"`
	Clearance   string   `json:"clearance"`
	Permissions []string `json:"permissions"`
}

// MaskingResponse 脱敏响应
type MaskingResponse struct {
	MaskedData   map[string]interface{} `json:"masked_data"`
	Metadata     map[string]interface{} `json:"metadata"`
	AuditID      string                 `json:"audit_id"`
	ProcessingTime time.Duration         `json:"processing_time"`
	RequestID    string                 `json:"request_id"`
	Timestamp    time.Time              `json:"timestamp"`
}

// NewDataMaskingService 创建数据脱敏服务
func NewDataMaskingService(config *DataMaskingConfig, logger *slog.Logger) *DataMaskingService {
	service := &DataMaskingService{
		logger:  logger,
		config:  config,
		maskers: make(map[string]DataMasker),
	}

	// 初始化脱敏器
	service.initializeMaskers()

	// 初始化缓存
	if config.EnableCache {
		service.cache = NewMaskingCache(config.CacheTTL)
	}

	// 初始化审计日志
	if config.EnableAudit {
		service.auditLogger = NewAuditLogger(logger)
	}

	return service
}

// initializeMaskers 初始化脱敏器
func (dms *DataMaskingService) initializeMaskers() {
	dms.maskers["name"] = NewNameMasker()
	dms.maskers["email"] = NewEmailMasker()
	dms.maskers["phone"] = NewPhoneMasker()
	dms.maskers["id_card"] = NewIDCardMasker()
	dms.maskers["bank_account"] = NewBankAccountMasker()
	dms.maskers["address"] = NewAddressMasker()
	dms.maskers["company"] = NewCompanyMasker()
	dms.maskers["credit_card"] = NewCreditCardMasker()
	dms.maskers["case_number"] = NewCaseNumberMasker()
	dms.maskers["license"] = NewLicenseMasker()
	dms.maskers["generic"] = NewGenericMasker()
}

// MaskDocument 脱敏文档
func (dms *DataMaskingService) MaskDocument(ctx context.Context, req *MaskingRequest) (*MaskingResponse, error) {
	startTime := time.Now()

	// 记录审计日志
	var auditID string
	if dms.auditLogger != nil {
		auditID = dms.auditLogger.LogMaskingRequest(req)
	}

	// 选择脱敏策略
	strategy := dms.selectMaskingStrategy(req.UserInfo, req.Context)

	// 处理脱敏请求
	result := make(map[string]interface{})
	metadata := make(map[string]interface{})

	for _, field := range req.Fields {
		maskedValue, err := dms.maskField(field, strategy, req)
		if err != nil {
			if dms.auditLogger != nil {
				dms.auditLogger.LogMaskingError(auditID, field.Name, err)
			}
			return nil, fmt.Errorf("脱敏字段 %s 失败: %w", field.Name, err)
		}

		result[field.Name] = maskedValue
		metadata[field.Name] = map[string]interface{}{
			"original_type": field.Type,
			"strategy":      strategy,
			"masked":        maskedValue != field.Value,
		}
	}

	// 构建响应
	response := &MaskingResponse{
		MaskedData:     result,
		Metadata:       metadata,
		AuditID:        auditID,
		ProcessingTime: time.Since(startTime),
		RequestID:      req.RequestID,
		Timestamp:      time.Now(),
	}

	// 记录成功日志
	if dms.auditLogger != nil {
		dms.auditLogger.LogMaskingSuccess(auditID, req, response)
	}

	return response, nil
}

// maskField 脱敏单个字段
func (dms *DataMaskingService) maskField(field FieldConfig, strategy string, req *MaskingRequest) (string, error) {
	// 检查缓存
	if dms.cache != nil {
		if cached, found := dms.cache.Get(field.Value, strategy, field.Type); found {
			return cached, nil
		}
	}

	// 获取脱敏器
	masker, exists := dms.maskers[field.Type]
	if !exists {
		masker = dms.maskers["generic"]
	}

	// 构建脱敏配置
	config := MaskingConfig{
		Type:      field.Type,
		Strategy:  strategy,
		Params:    field.Config,
		Context:   req.Context,
		UserLevel: dms.getUserLevel(req.UserInfo),
		Purpose:   req.Context["purpose"],
	}

	// 执行脱敏
	masked, err := masker.Mask(field.Value, config)
	if err != nil {
		return "", err
	}

	// 更新缓存
	if dms.cache != nil {
		dms.cache.Set(field.Value, masked, strategy, field.Type)
	}

	return masked, nil
}

// selectMaskingStrategy 选择脱敏策略
func (dms *DataMaskingService) selectMaskingStrategy(userInfo *UserInfo, context map[string]string) string {
	// 根据用户级别选择策略
	userLevel := dms.getUserLevel(userInfo)
	purpose := context["purpose"]

	switch {
	case userLevel == "admin" && purpose == "development":
		return "minimal"
	case userLevel == "lawyer" && (purpose == "development" || purpose == "analysis"):
		return "moderate"
	case userLevel == "assistant" || purpose == "sharing":
		return "strict"
	case purpose == "export":
		return "export"
	case purpose == "testing":
		return "testing"
	default:
		return dms.config.DefaultStrategy
	}
}

// getUserLevel 获取用户级别
func (dms *DataMaskingService) getUserLevel(userInfo *UserInfo) string {
	if userInfo == nil {
		return "unknown"
	}

	// 检查角色
	for _, role := range userInfo.Roles {
		switch role {
		case "super_admin", "admin":
			return "admin"
		case "partner", "senior_partner":
			return "lawyer"
		case "lawyer", "associate":
			return "lawyer"
		case "assistant", "paralegal":
			return "assistant"
		case "client":
			return "client"
		}
	}

	// 检查权限
	for _, permission := range userInfo.Permissions {
		if strings.Contains(permission, "admin") {
			return "admin"
		}
		if strings.Contains(permission, "lawyer") {
			return "lawyer"
		}
	}

	return "unknown"
}

// GetMaskingStatistics 获取脱敏统计信息
func (dms *DataMaskingService) GetMaskingStatistics(ctx context.Context) (map[string]interface{}, error) {
	dms.mu.RLock()
	defer dms.mu.RUnlock()

	stats := map[string]interface{}{
		"total_maskers":      len(dms.maskers),
		"cache_enabled":      dms.config.EnableCache,
		"audit_enabled":      dms.config.EnableAudit,
		"default_strategy":   dms.config.DefaultStrategy,
		"performance_mode":   dms.config.PerformanceMode,
		"max_concurrency":    dms.config.MaxConcurrency,
		"available_strategies": []string{"minimal", "moderate", "strict", "export", "testing"},
	}

	// 添加缓存统计
	if dms.cache != nil {
		stats["cache_stats"] = dms.cache.GetStatistics()
	}

	// 添加审计统计
	if dms.auditLogger != nil {
		stats["audit_stats"] = dms.auditLogger.GetStatistics()
	}

	return stats, nil
}

// ValidateMaskingConfig 验证脱敏配置
func (dms *DataMaskingService) ValidateMaskingConfig(config *MaskingConfig) error {
	if config == nil {
		return fmt.Errorf("脱敏配置不能为空")
	}

	if config.Type == "" {
		return fmt.Errorf("脱敏类型不能为空")
	}

	if config.Strategy == "" {
		config.Strategy = dms.config.DefaultStrategy
	}

	// 验证策略
	validStrategies := []string{"minimal", "moderate", "strict", "export", "testing"}
	valid := false
	for _, strategy := range validStrategies {
		if config.Strategy == strategy {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("无效的脱敏策略: %s", config.Strategy)
	}

	return nil
}

// NameMasker 姓名脱敏器
type NameMasker struct {
	patterns map[string]*regexp.Regexp
}

func NewNameMasker() *NameMasker {
	return &NameMasker{
		patterns: map[string]*regexp.Regexp{
			"chinese":      regexp.MustCompile(`[\p{Han}]{2,4}`),
			"western":      regexp.MustCompile(`[A-Z][a-z]+\s+[A-Z][a-z]+`),
			"single":       regexp.MustCompile(`[\p{L}]{2,}`),
		},
	}
}

func (nm *NameMasker) Mask(data string, config MaskingConfig) (string, error) {
	if data == "" {
		return "", nil
	}

	switch config.Strategy {
	case "minimal":
		return nm.maskMinimal(data), nil
	case "moderate":
		return nm.maskModerate(data), nil
	case "strict":
		return nm.maskStrict(data), nil
	case "export":
		return nm.maskExport(data), nil
	case "testing":
		return nm.maskTesting(data), nil
	default:
		return nm.maskModerate(data), nil
	}
}

func (nm *NameMasker) maskMinimal(name string) string {
	if len(name) <= 1 {
		return strings.Repeat("*", len(name))
	}
	return name[:1] + strings.Repeat("*", len(name)-1)
}

func (nm *NameMasker) maskModerate(name string) string {
	if len(name) <= 2 {
		return strings.Repeat("*", len(name))
	}
	return name[:1] + strings.Repeat("*", len(name)-2) + name[len(name)-1:]
}

func (nm *NameMasker) maskStrict(name string) string {
	return strings.Repeat("*", len(name))
}

func (nm *NameMasker) maskExport(name string) string {
	return "[姓名]"
}

func (nm *NameMasker) maskTesting(name string) string {
	return "测试姓名"
}

func (nm *NameMasker) GetType() string {
	return "name"
}

// EmailMasker 邮箱脱敏器
type EmailMasker struct{}

func NewEmailMasker() *EmailMasker {
	return &EmailMasker{}
}

func (em *EmailMasker) Mask(data string, config MaskingConfig) (string, error) {
	if data == "" {
		return "", nil
	}

	parts := strings.Split(data, "@")
	if len(parts) != 2 {
		return "****@****.com", nil
	}

	username := parts[0]
	domain := parts[1]

	switch config.Strategy {
	case "minimal":
		return em.maskMinimal(username, domain), nil
	case "moderate":
		return em.maskModerate(username, domain), nil
	case "strict":
		return em.maskStrict(username, domain), nil
	case "export":
		return em.maskExport(username, domain), nil
	case "testing":
		return em.maskTesting(username, domain), nil
	default:
		return em.maskModerate(username, domain), nil
	}
}

func (em *EmailMasker) maskMinimal(username, domain string) string {
	if len(username) <= 2 {
		return strings.Repeat("*", len(username)) + "@" + domain
	}
	return username[:1] + strings.Repeat("*", len(username)-2) + username[len(username)-1:] + "@" + domain
}

func (em *EmailMasker) maskModerate(username, domain string) string {
	if len(username) <= 1 {
		return strings.Repeat("*", len(username)) + "@" + domain
	}
	return username[:1] + strings.Repeat("*", len(username)-1) + "@" + domain
}

func (em *EmailMasker) maskStrict(username, domain string) string {
	return strings.Repeat("*", len(username)) + "@" + strings.Repeat("*", len(domain))
}

func (em *EmailMasker) maskExport(username, domain string) string {
	return "[邮箱]"
}

func (em *EmailMasker) maskTesting(username, domain string) string {
	return "test@example.com"
}

func (em *EmailMasker) GetType() string {
	return "email"
}

// PhoneMasker 电话脱敏器
type PhoneMasker struct {
	patterns map[string]*regexp.Regexp
}

func NewPhoneMasker() *PhoneMasker {
	return &PhoneMasker{
		patterns: map[string]*regexp.Regexp{
			"mobile":     regexp.MustCompile(`^1[3-9]\d{9}$`),
			"landline":   regexp.MustCompile(`^0\d{2,3}\d{7,8}$`),
			"international": regexp.MustCompile(`^\+\d{1,3}\d{4,14}$`),
		},
	}
}

func (pm *PhoneMasker) Mask(data string, config MaskingConfig) (string, error) {
	if data == "" {
		return "", nil
	}

	switch config.Strategy {
	case "minimal":
		return pm.maskMinimal(data), nil
	case "moderate":
		return pm.maskModerate(data), nil
	case "strict":
		return pm.maskStrict(data), nil
	case "export":
		return pm.maskExport(data), nil
	case "testing":
		return pm.maskTesting(data), nil
	default:
		return pm.maskModerate(data), nil
	}
}

func (pm *PhoneMasker) maskMinimal(phone string) string {
	if len(phone) <= 7 {
		return strings.Repeat("*", len(phone))
	}
	return phone[:3] + strings.Repeat("*", len(phone)-6) + phone[len(phone)-3:]
}

func (pm *PhoneMasker) maskModerate(phone string) string {
	if len(phone) <= 4 {
		return strings.Repeat("*", len(phone))
	}
	return phone[:3] + strings.Repeat("*", len(phone)-4) + phone[len(phone)-4:]
}

func (pm *PhoneMasker) maskStrict(phone string) string {
	return strings.Repeat("*", len(phone))
}

func (pm *PhoneMasker) maskExport(phone string) string {
	return "[电话]"
}

func (pm *PhoneMasker) maskTesting(phone string) string {
	return "138****5678"
}

func (pm *PhoneMasker) GetType() string {
	return "phone"
}

// IDCardMasker 身份证脱敏器
type IDCardMasker struct {
	pattern *regexp.Regexp
}

func NewIDCardMasker() *IDCardMasker {
	return &IDCardMasker{
		pattern: regexp.MustCompile(`^\d{17}[\dXx]$|^\d{15}$`),
	}
}

func (im *IDCardMasker) Mask(data string, config MaskingConfig) (string, error) {
	if data == "" {
		return "", nil
	}

	switch config.Strategy {
	case "minimal":
		return im.maskMinimal(data), nil
	case "moderate":
		return im.maskModerate(data), nil
	case "strict":
		return im.maskStrict(data), nil
	case "export":
		return im.maskExport(data), nil
	case "testing":
		return im.maskTesting(data), nil
	default:
		return im.maskModerate(data), nil
	}
}

func (im *IDCardMasker) maskMinimal(id string) string {
	if len(id) <= 8 {
		return strings.Repeat("*", len(id))
	}
	return id[:6] + strings.Repeat("*", len(id)-8) + id[len(id)-2:]
}

func (im *IDCardMasker) maskModerate(id string) string {
	if len(id) <= 6 {
		return strings.Repeat("*", len(id))
	}
	return id[:4] + strings.Repeat("*", len(id)-6) + id[len(id)-2:]
}

func (im *IDCardMasker) maskStrict(id string) string {
	return strings.Repeat("*", len(id))
}

func (im *IDCardMasker) maskExport(id string) string {
	return "[身份证号]"
}

func (im *IDCardMasker) maskTesting(id string) string {
	return "110101********1234"
}

func (im *IDCardMasker) GetType() string {
	return "id_card"
}

// BankAccountMasker 银行账号脱敏器
type BankAccountMasker struct{}

func NewBankAccountMasker() *BankAccountMasker {
	return &BankAccountMasker{}
}

func (bam *BankAccountMasker) Mask(data string, config MaskingConfig) (string, error) {
	if data == "" {
		return "", nil
	}

	switch config.Strategy {
	case "minimal":
		return bam.maskMinimal(data), nil
	case "moderate":
		return bam.maskModerate(data), nil
	case "strict":
		return bam.maskStrict(data), nil
	case "export":
		return bam.maskExport(data), nil
	case "testing":
		return bam.maskTesting(data), nil
	default:
		return bam.maskModerate(data), nil
	}
}

func (bam *BankAccountMasker) maskMinimal(account string) string {
	if len(account) <= 8 {
		return strings.Repeat("*", len(account))
	}
	return account[:4] + strings.Repeat("*", len(account)-8) + account[len(account)-4:]
}

func (bam *BankAccountMasker) maskModerate(account string) string {
	if len(account) <= 6 {
		return strings.Repeat("*", len(account))
	}
	return account[:3] + strings.Repeat("*", len(account)-6) + account[len(account)-3:]
}

func (bam *BankAccountMasker) maskStrict(account string) string {
	return strings.Repeat("*", len(account))
}

func (bam *BankAccountMasker) maskExport(account string) string {
	return "[银行账号]"
}

func (bam *BankAccountMasker) maskTesting(account string) string {
	return "6222************1234"
}

func (bam *BankAccountMasker) GetType() string {
	return "bank_account"
}

// GenericMasker 通用脱敏器
type GenericMasker struct{}

func NewGenericMasker() *GenericMasker {
	return &GenericMasker{}
}

func (gm *GenericMasker) Mask(data string, config MaskingConfig) (string, error) {
	if data == "" {
		return "", nil
	}

	switch config.Strategy {
	case "minimal":
		return gm.maskMinimal(data), nil
	case "moderate":
		return gm.maskModerate(data), nil
	case "strict":
		return gm.maskStrict(data), nil
	case "export":
		return gm.maskExport(data), nil
	case "testing":
		return gm.maskTesting(data), nil
	default:
		return gm.maskModerate(data), nil
	}
}

func (gm *GenericMasker) maskMinimal(data string) string {
	if len(data) <= 4 {
		return strings.Repeat("*", len(data))
	}
	return data[:2] + strings.Repeat("*", len(data)-4) + data[len(data)-2:]
}

func (gm *GenericMasker) maskModerate(data string) string {
	if len(data) <= 2 {
		return strings.Repeat("*", len(data))
	}
	return data[:1] + strings.Repeat("*", len(data)-2) + data[len(data)-1:]
}

func (gm *GenericMasker) maskStrict(data string) string {
	return strings.Repeat("*", len(data))
}

func (gm *GenericMasker) maskExport(data string) string {
	return "[敏感信息]"
}

func (gm *GenericMasker) maskTesting(data string) string {
	return "[测试数据]"
}

func (gm *GenericMasker) GetType() string {
	return "generic"
}

// 更多脱敏器实现...
// AddressMasker, CompanyMasker, CreditCardMasker, CaseNumberMasker, LicenseMasker

// AddressMasker 地址脱敏器
type AddressMasker struct{}

func NewAddressMasker() *AddressMasker {
	return &AddressMasker{}
}

func (am *AddressMasker) Mask(data string, config MaskingConfig) (string, error) {
	if data == "" {
		return "", nil
	}

	switch config.Strategy {
	case "minimal":
		return am.maskMinimal(data), nil
	case "moderate":
		return am.maskModerate(data), nil
	case "strict":
		return am.maskStrict(data), nil
	case "export":
		return am.maskExport(data), nil
	case "testing":
		return am.maskTesting(data), nil
	default:
		return am.maskModerate(data), nil
	}
}

func (am *AddressMasker) maskMinimal(address string) string {
	parts := strings.Split(address, " ")
	if len(parts) <= 2 {
		return strings.Repeat("*", len(address))
	}
	return parts[0] + " " + strings.Repeat("*", len(strings.Join(parts[1:], " ")))
}

func (am *AddressMasker) maskModerate(address string) string {
	return "[地址]" + strings.Repeat("*", len(address)-4)
}

func (am *AddressMasker) maskStrict(address string) string {
	return "[地址]"
}

func (am *AddressMasker) maskExport(address string) string {
	return "[详细地址]"
}

func (am *AddressMasker) maskTesting(address string) string {
	return "测试地址"
}

func (am *AddressMasker) GetType() string {
	return "address"
}

// CompanyMasker 公司脱敏器
type CompanyMasker struct{}

func NewCompanyMasker() *CompanyMasker {
	return &CompanyMasker{}
}

func (cm *CompanyMasker) Mask(data string, config MaskingConfig) (string, error) {
	if data == "" {
		return "", nil
	}

	switch config.Strategy {
	case "minimal":
		return cm.maskMinimal(data), nil
	case "moderate":
		return cm.maskModerate(data), nil
	case "strict":
		return cm.maskStrict(data), nil
	case "export":
		return cm.maskExport(data), nil
	case "testing":
		return cm.maskTesting(data), nil
	default:
		return cm.maskModerate(data), nil
	}
}

func (cm *CompanyMasker) maskMinimal(company string) string {
	if len(company) <= 4 {
		return strings.Repeat("*", len(company))
	}
	return company[:2] + strings.Repeat("*", len(company)-4) + company[len(company)-2:]
}

func (cm *CompanyMasker) maskModerate(company string) string {
	if len(company) <= 2 {
		return strings.Repeat("*", len(company))
	}
	return company[:1] + strings.Repeat("*", len(company)-2) + company[len(company)-1:]
}

func (cm *CompanyMasker) maskStrict(company string) string {
	return "[公司名称]"
}

func (cm *CompanyMasker) maskExport(company string) string {
	return "[公司]"
}

func (cm *CompanyMasker) maskTesting(company string) string {
	return "测试公司"
}

func (cm *CompanyMasker) GetType() string {
	return "company"
}

// CreditCardMasker 信用卡脱敏器
type CreditCardMasker struct {
	pattern *regexp.Regexp
}

func NewCreditCardMasker() *CreditCardMasker {
	return &CreditCardMasker{
		pattern: regexp.MustCompile(`^\d{13,19}$`),
	}
}

func (ccm *CreditCardMasker) Mask(data string, config MaskingConfig) (string, error) {
	if data == "" {
		return "", nil
	}

	// 移除空格和连字符
	cleanNumber := strings.ReplaceAll(strings.ReplaceAll(data, " ", ""), "-", "")

	switch config.Strategy {
	case "minimal":
		return ccm.maskMinimal(cleanNumber), nil
	case "moderate":
		return ccm.maskModerate(cleanNumber), nil
	case "strict":
		return ccm.maskStrict(cleanNumber), nil
	case "export":
		return ccm.maskExport(cleanNumber), nil
	case "testing":
		return ccm.maskTesting(cleanNumber), nil
	default:
		return ccm.maskModerate(cleanNumber), nil
	}
}

func (ccm *CreditCardMasker) maskMinimal(card string) string {
	if len(card) <= 8 {
		return strings.Repeat("*", len(card))
	}
	return card[:4] + strings.Repeat("*", len(card)-8) + card[len(card)-4:]
}

func (ccm *CreditCardMasker) maskModerate(card string) string {
	if len(card) <= 6 {
		return strings.Repeat("*", len(card))
	}
	return card[:4] + strings.Repeat("*", len(card)-6) + card[len(card)-2:]
}

func (ccm *CreditCardMasker) maskStrict(card string) string {
	return strings.Repeat("*", len(card))
}

func (ccm *CreditCardMasker) maskExport(card string) string {
	return "[信用卡号]"
}

func (ccm *CreditCardMasker) maskTesting(card string) string {
	return "4111************1111"
}

func (ccm *CreditCardMasker) GetType() string {
	return "credit_card"
}

// CaseNumberMasker 案件编号脱敏器
type CaseNumberMasker struct{}

func NewCaseNumberMasker() *CaseNumberMasker {
	return &CaseNumberMasker{}
}

func (cnm *CaseNumberMasker) Mask(data string, config MaskingConfig) (string, error) {
	if data == "" {
		return "", nil
	}

	switch config.Strategy {
	case "minimal":
		return cnm.maskMinimal(data), nil
	case "moderate":
		return cnm.maskModerate(data), nil
	case "strict":
		return cnm.maskStrict(data), nil
	case "export":
		return cnm.maskExport(data), nil
	case "testing":
		return cnm.maskTesting(data), nil
	default:
		return cnm.maskModerate(data), nil
	}
}

func (cnm *CaseNumberMasker) maskMinimal(caseNumber string) string {
	if len(caseNumber) <= 6 {
		return strings.Repeat("*", len(caseNumber))
	}
	parts := strings.Split(caseNumber, "-")
	if len(parts) >= 2 {
		return parts[0] + "-" + strings.Repeat("*", len(strings.Join(parts[1:], "-")))
	}
	return caseNumber[:3] + strings.Repeat("*", len(caseNumber)-6) + caseNumber[len(caseNumber)-3:]
}

func (cnm *CaseNumberMasker) maskModerate(caseNumber string) string {
	if len(caseNumber) <= 4 {
		return strings.Repeat("*", len(caseNumber))
	}
	return caseNumber[:2] + strings.Repeat("*", len(caseNumber)-4) + caseNumber[len(caseNumber)-2:]
}

func (cnm *CaseNumberMasker) maskStrict(caseNumber string) string {
	return "[案件编号]"
}

func (cnm *CaseNumberMasker) maskExport(caseNumber string) string {
	return "[案件]"
}

func (cnm *CaseNumberMasker) maskTesting(caseNumber string) string {
	return "CASE2024****001"
}

func (cnm *CaseNumberMasker) GetType() string {
	return "case_number"
}

// LicenseMasker 证件脱敏器
type LicenseMasker struct{}

func NewLicenseMasker() *LicenseMasker {
	return &LicenseMasker{}
}

func (lm *LicenseMasker) Mask(data string, config MaskingConfig) (string, error) {
	if data == "" {
		return "", nil
	}

	switch config.Strategy {
	case "minimal":
		return lm.maskMinimal(data), nil
	case "moderate":
		return lm.maskModerate(data), nil
	case "strict":
		return lm.maskStrict(data), nil
	case "export":
		return lm.maskExport(data), nil
	case "testing":
		return lm.maskTesting(data), nil
	default:
		return lm.maskModerate(data), nil
	}
}

func (lm *LicenseMasker) maskMinimal(license string) string {
	if len(license) <= 6 {
		return strings.Repeat("*", len(license))
	}
	return license[:3] + strings.Repeat("*", len(license)-6) + license[len(license)-3:]
}

func (lm *LicenseMasker) maskModerate(license string) string {
	if len(license) <= 4 {
		return strings.Repeat("*", len(license))
	}
	return license[:2] + strings.Repeat("*", len(license)-4) + license[len(license)-2:]
}

func (lm *LicenseMasker) maskStrict(license string) string {
	return "[证件号]"
}

func (lm *LicenseMasker) maskExport(license string) string {
	return "[证件]"
}

func (lm *LicenseMasker) maskTesting(license string) string {
	return "测试证件号"
}

func (lm *LicenseMasker) GetType() string {
	return "license"
}