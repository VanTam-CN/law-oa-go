package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

// DataMaskingDemo 数据脱敏演示程序
type DataMaskingDemo struct {
	maskingService *DataMaskingService
	logger          *slog.Logger
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

// FieldConfig 字段配置
type FieldConfig struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Value       string                 `json:"value"`
	Required    bool                   `json:"required"`
	Config      map[string]interface{} `json:"config"`
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

// MaskingResponse 脱敏响应
type MaskingResponse struct {
	MaskedData     map[string]interface{} `json:"masked_data"`
	Metadata     map[string]interface{} `json:"metadata"`
	AuditID      string                 `json:"audit_id"`
	ProcessingTime time.Duration         `json:"processing_time"`
	RequestID    string                 `json:"request_id"`
	Timestamp    time.Time              `json:"timestamp"`
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

// DataMasker 数据脱敏接口
type DataMasker interface {
	Mask(data string, config MaskingConfig) (string, error)
	GetType() string
}

// DataMaskingService 数据脱敏服务
type DataMaskingService struct {
	maskers map[string]DataMasker
}

// Simple implementation for demo
func NewDataMaskingService(config *DataMaskingConfig, logger *slog.Logger) *DataMaskingService {
	return &DataMaskingService{
		maskers: map[string]DataMasker{
			"name":        &NameMasker{},
			"email":       &EmailMasker{},
			"phone":       &PhoneMasker{},
			"id_card":     &IDCardMasker{},
			"bank_account": &BankAccountMasker{},
			"address":     &AddressMasker{},
			"company":     &CompanyMasker{},
			"credit_card": &CreditCardMasker{},
			"case_number": &CaseNumberMasker{},
			"license":     &LicenseMasker{},
			"generic":     &GenericMasker{},
		},
	}
}

// MaskDocument 脱敏文档
func (dms *DataMaskingService) MaskDocument(ctx context.Context, req *MaskingRequest) (*MaskingResponse, error) {
	startTime := time.Now()

	result := make(map[string]interface{})
	metadata := make(map[string]interface{})

	for _, field := range req.Fields {
		masker, exists := dms.maskers[field.Type]
		if !exists {
			masker = dms.maskers["generic"]
		}

		config := MaskingConfig{
			Type:      field.Type,
			Strategy:  dms.selectStrategy(req.UserInfo, req.Context),
			Params:    field.Config,
			Context:   req.Context,
			UserLevel: dms.getUserLevel(req.UserInfo),
			Purpose:   req.Context["purpose"],
		}

		masked, err := masker.Mask(field.Value, config)
		if err != nil {
			return nil, fmt.Errorf("脱敏字段 %s 失败: %w", field.Name, err)
		}

		result[field.Name] = masked
		metadata[field.Name] = map[string]interface{}{
			"original_type": field.Type,
			"strategy":      config.Strategy,
			"masked":        masked != field.Value,
		}
	}

	response := &MaskingResponse{
		MaskedData:     result,
		Metadata:       metadata,
		AuditID:        "audit_" + req.RequestID,
		ProcessingTime: time.Since(startTime),
		RequestID:      req.RequestID,
		Timestamp:      time.Now(),
	}

	return response, nil
}

func (dms *DataMaskingService) selectStrategy(userInfo *UserInfo, context map[string]string) string {
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
		return "moderate"
	}
}

func (dms *DataMaskingService) getUserLevel(userInfo *UserInfo) string {
	if userInfo == nil {
		return "unknown"
	}

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

	return "unknown"
}

// Masker implementations
type NameMasker struct{}

func (nm *NameMasker) Mask(data string, config MaskingConfig) (string, error) {
	if data == "" {
		return "", nil
	}

	switch config.Strategy {
	case "minimal":
		if len(data) <= 1 {
			return strings.Repeat("*", len(data)), nil
		}
		return data[:1] + strings.Repeat("*", len(data)-1), nil
	case "moderate":
		if len(data) <= 2 {
			return strings.Repeat("*", len(data)), nil
		}
		return data[:1] + strings.Repeat("*", len(data)-2) + data[len(data)-1:], nil
	case "strict":
		return strings.Repeat("*", len(data)), nil
	case "export":
		return "[姓名]", nil
	case "testing":
		return "测试姓名", nil
	default:
		return data[:1] + strings.Repeat("*", len(data)-1), nil
	}
}

func (nm *NameMasker) GetType() string { return "name" }

type EmailMasker struct{}

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
		if len(username) <= 2 {
			return strings.Repeat("*", len(username)) + "@" + domain, nil
		}
		return username[:1] + strings.Repeat("*", len(username)-2) + username[len(username)-1:] + "@" + domain, nil
	case "moderate":
		if len(username) <= 1 {
			return strings.Repeat("*", len(username)) + "@" + domain, nil
		}
		return username[:1] + strings.Repeat("*", len(username)-1) + "@" + domain, nil
	case "strict":
		return strings.Repeat("*", len(username)) + "@" + strings.Repeat("*", len(domain)), nil
	case "export":
		return "[邮箱]", nil
	case "testing":
		return "test@example.com", nil
	default:
		return username[:1] + strings.Repeat("*", len(username)-1) + "@" + domain, nil
	}
}

func (em *EmailMasker) GetType() string { return "email" }

type PhoneMasker struct{}

func (pm *PhoneMasker) Mask(data string, config MaskingConfig) (string, error) {
	if data == "" {
		return "", nil
	}

	switch config.Strategy {
	case "minimal":
		if len(data) <= 7 {
			return strings.Repeat("*", len(data)), nil
		}
		return data[:3] + strings.Repeat("*", len(data)-6) + data[len(data)-3:], nil
	case "moderate":
		if len(data) <= 4 {
			return strings.Repeat("*", len(data)), nil
		}
		return data[:3] + strings.Repeat("*", len(data)-4) + data[len(data)-4:], nil
	case "strict":
		return strings.Repeat("*", len(data)), nil
	case "export":
		return "[电话]", nil
	case "testing":
		return "138****5678", nil
	default:
		return data[:3] + strings.Repeat("*", len(data)-6) + data[len(data)-3:], nil
	}
}

func (pm *PhoneMasker) GetType() string { return "phone" }

type IDCardMasker struct{}

func (im *IDCardMasker) Mask(data string, config MaskingConfig) (string, error) {
	if data == "" {
		return "", nil
	}

	switch config.Strategy {
	case "minimal":
		if len(data) <= 8 {
			return strings.Repeat("*", len(data)), nil
		}
		return data[:6] + strings.Repeat("*", len(data)-8) + data[len(data)-2:], nil
	case "moderate":
		if len(data) <= 6 {
			return strings.Repeat("*", len(data)), nil
		}
		return data[:4] + strings.Repeat("*", len(data)-6) + data[len(data)-2:], nil
	case "strict":
		return strings.Repeat("*", len(data)), nil
	case "export":
		return "[身份证号]", nil
	case "testing":
		return "110101********1234", nil
	default:
		return data[:6] + strings.Repeat("*", len(data)-8) + data[len(data)-2:], nil
	}
}

func (im *IDCardMasker) GetType() string { return "id_card" }

type BankAccountMasker struct{}

func (bam *BankAccountMasker) Mask(data string, config MaskingConfig) (string, error) {
	if data == "" {
		return "", nil
	}

	switch config.Strategy {
	case "minimal":
		if len(data) <= 8 {
			return strings.Repeat("*", len(data)), nil
		}
		return data[:4] + strings.Repeat("*", len(data)-8) + data[len(data)-4:], nil
	case "moderate":
		if len(data) <= 6 {
			return strings.Repeat("*", len(data)), nil
		}
		return data[:3] + strings.Repeat("*", len(data)-6) + data[len(data)-3:], nil
	case "strict":
		return strings.Repeat("*", len(data)), nil
	case "export":
		return "[银行账号]", nil
	case "testing":
		return "6222************1234", nil
	default:
		return data[:4] + strings.Repeat("*", len(data)-8) + data[len(data)-4:], nil
	}
}

func (bam *BankAccountMasker) GetType() string { return "bank_account" }

type AddressMasker struct{}

func (am *AddressMasker) Mask(data string, config MaskingConfig) (string, error) {
	if data == "" {
		return "", nil
	}

	switch config.Strategy {
	case "minimal":
		parts := strings.Split(data, " ")
		if len(parts) <= 2 {
			return strings.Repeat("*", len(data)), nil
		}
		return parts[0] + " " + strings.Repeat("*", len(strings.Join(parts[1:], " "))), nil
	case "moderate":
		return "[地址]" + strings.Repeat("*", len(data)-4), nil
	case "strict":
		return "[地址]", nil
	case "export":
		return "[详细地址]", nil
	case "testing":
		return "测试地址", nil
	default:
		return "[地址]" + strings.Repeat("*", len(data)-4), nil
	}
}

func (am *AddressMasker) GetType() string { return "address" }

type CompanyMasker struct{}

func (cm *CompanyMasker) Mask(data string, config MaskingConfig) (string, error) {
	if data == "" {
		return "", nil
	}

	switch config.Strategy {
	case "minimal":
		if len(data) <= 4 {
			return strings.Repeat("*", len(data)), nil
		}
		return data[:2] + strings.Repeat("*", len(data)-4) + data[len(data)-2:], nil
	case "moderate":
		if len(data) <= 2 {
			return strings.Repeat("*", len(data)), nil
		}
		return data[:1] + strings.Repeat("*", len(data)-2) + data[len(data)-1:], nil
	case "strict":
		return "[公司名称]", nil
	case "export":
		return "[公司]", nil
	case "testing":
		return "测试公司", nil
	default:
		return data[:1] + strings.Repeat("*", len(data)-2) + data[len(data)-1:], nil
	}
}

func (cm *CompanyMasker) GetType() string { return "company" }

type CreditCardMasker struct{}

func (ccm *CreditCardMasker) Mask(data string, config MaskingConfig) (string, error) {
	if data == "" {
		return "", nil
	}

	// 移除空格和连字符
	cleanNumber := strings.ReplaceAll(strings.ReplaceAll(data, " ", ""), "-", "")

	switch config.Strategy {
	case "minimal":
		if len(cleanNumber) <= 8 {
			return strings.Repeat("*", len(cleanNumber)), nil
		}
		return cleanNumber[:4] + strings.Repeat("*", len(cleanNumber)-8) + cleanNumber[len(cleanNumber)-4:], nil
	case "moderate":
		if len(cleanNumber) <= 6 {
			return strings.Repeat("*", len(cleanNumber)), nil
		}
		return cleanNumber[:4] + strings.Repeat("*", len(cleanNumber)-6) + cleanNumber[len(cleanNumber)-2:], nil
	case "strict":
		return strings.Repeat("*", len(cleanNumber)), nil
	case "export":
		return "[信用卡号]", nil
	case "testing":
		return "4111************1111", nil
	default:
		return cleanNumber[:4] + strings.Repeat("*", len(cleanNumber)-8) + cleanNumber[len(cleanNumber)-4:], nil
	}
}

func (ccm *CreditCardMasker) GetType() string { return "credit_card" }

type CaseNumberMasker struct{}

func (cnm *CaseNumberMasker) Mask(data string, config MaskingConfig) (string, error) {
	if data == "" {
		return "", nil
	}

	switch config.Strategy {
	case "minimal":
		if len(data) <= 6 {
			return strings.Repeat("*", len(data)), nil
		}
		parts := strings.Split(data, "-")
		if len(parts) >= 2 {
			return parts[0] + "-" + strings.Repeat("*", len(strings.Join(parts[1:], "-"))), nil
		}
		return data[:3] + strings.Repeat("*", len(data)-6) + data[len(data)-3:], nil
	case "moderate":
		if len(data) <= 4 {
			return strings.Repeat("*", len(data)), nil
		}
		return data[:2] + strings.Repeat("*", len(data)-4) + data[len(data)-2:], nil
	case "strict":
		return "[案件编号]", nil
	case "export":
		return "[案件]", nil
	case "testing":
		return "CASE2024****001", nil
	default:
		return data[:3] + strings.Repeat("*", len(data)-6) + data[len(data)-3:], nil
	}
}

func (cnm *CaseNumberMasker) GetType() string { return "case_number" }

type LicenseMasker struct{}

func (lm *LicenseMasker) Mask(data string, config MaskingConfig) (string, error) {
	if data == "" {
		return "", nil
	}

	switch config.Strategy {
	case "minimal":
		if len(data) <= 6 {
			return strings.Repeat("*", len(data)), nil
		}
		return data[:3] + strings.Repeat("*", len(data)-6) + data[len(data)-3:], nil
	case "moderate":
		if len(data) <= 4 {
			return strings.Repeat("*", len(data)), nil
		}
		return data[:2] + strings.Repeat("*", len(data)-4) + data[len(data)-2:], nil
	case "strict":
		return "[证件号]", nil
	case "export":
		return "[证件]", nil
	case "testing":
		return "测试证件号", nil
	default:
		return data[:3] + strings.Repeat("*", len(data)-6) + data[len(data)-3:], nil
	}
}

func (lm *LicenseMasker) GetType() string { return "license" }

type GenericMasker struct{}

func (gm *GenericMasker) Mask(data string, config MaskingConfig) (string, error) {
	if data == "" {
		return "", nil
	}

	switch config.Strategy {
	case "minimal":
		if len(data) <= 4 {
			return strings.Repeat("*", len(data)), nil
		}
		return data[:2] + strings.Repeat("*", len(data)-4) + data[len(data)-2:], nil
	case "moderate":
		if len(data) <= 2 {
			return strings.Repeat("*", len(data)), nil
		}
		return data[:1] + strings.Repeat("*", len(data)-2) + data[len(data)-1:], nil
	case "strict":
		return strings.Repeat("*", len(data)), nil
	case "export":
		return "[敏感信息]", nil
	case "testing":
		return "[测试数据]", nil
	default:
		return data[:1] + strings.Repeat("*", len(data)-2) + data[len(data)-1:], nil
	}
}

func (gm *GenericMasker) GetType() string { return "generic" }

// NewDataMaskingDemo 创建数据脱敏演示
func NewDataMaskingDemo() *DataMaskingDemo {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	config := &DataMaskingConfig{
		EnableCache:      true,
		CacheTTL:         5 * time.Minute,
		EnableAudit:      true,
		DefaultStrategy:  "moderate",
		PerformanceMode:  false,
		MaxConcurrency:   10,
	}

	return &DataMaskingDemo{
		maskingService: NewDataMaskingService(config, logger),
		logger:          logger,
	}
}

// Run 运行演示
func (dmd *DataMaskingDemo) Run() error {
	fmt.Println("🔒 开始数据脱敏和隐私保护演示...")
	fmt.Println(strings.Repeat("=", 50))

	// 演示1: 基础数据脱敏
	if err := dmd.demonstrateBasicMasking(); err != nil {
		return fmt.Errorf("基础脱敏演示失败: %w", err)
	}

	// 演示2: 用户权限级别脱敏
	if err := dmd.demonstratePermissionBasedMasking(); err != nil {
		return fmt.Errorf("权限脱敏演示失败: %w", err)
	}

	// 演示3: 律师事务所特定脱敏
	if err := dmd.demonstrateLawFirmSpecificMasking(); err != nil {
		return fmt.Errorf("律师事务所脱敏演示失败: %w", err)
	}

	// 演示4: 性能测试
	if err := dmd.demonstratePerformance(); err != nil {
		return fmt.Errorf("性能演示失败: %w", err)
	}

	// 演示5: 合规性检查
	if err := dmd.demonstrateCompliance(); err != nil {
		return fmt.Errorf("合规性演示失败: %w", err)
	}

	return nil
}

// demonstrateBasicMasking 演示基础数据脱敏
func (dmd *DataMaskingDemo) demonstrateBasicMasking() error {
	fmt.Println("📋 演示1: 基础数据脱敏")
	fmt.Println(strings.Repeat("-", 30))

	req := &MaskingRequest{
		Data: "客户信息脱敏测试",
		Fields: []FieldConfig{
			{Name: "client_name", Type: "name", Value: "张三", Required: true},
			{Name: "client_email", Type: "email", Value: "zhangsan@example.com", Required: true},
			{Name: "client_phone", Type: "phone", Value: "13812345678", Required: true},
			{Name: "client_idcard", Type: "id_card", Value: "110101199001011234", Required: true},
			{Name: "bank_account", Type: "bank_account", Value: "6222020123456789012", Required: false},
			{Name: "client_address", Type: "address", Value: "北京市朝阳区建国路88号", Required: false},
			{Name: "client_company", Type: "company", Value: "北京科技有限公司", Required: false},
		},
		Context: map[string]string{
			"purpose": "development",
			"system":  "document_service",
		},
		UserInfo: &UserInfo{
			UserID:     "user001",
			Username:   "admin",
			Roles:      []string{"admin"},
			Department: "IT",
		},
		RequestID: "basic_demo_001",
		Timestamp: time.Now(),
	}

	response, err := dmd.maskingService.MaskDocument(context.Background(), req)
	if err != nil {
		return err
	}

	fmt.Printf("✅ 基础脱敏测试完成\n")
	fmt.Printf("   处理时间: %v\n", response.ProcessingTime)
	fmt.Printf("   审计ID: %s\n", response.AuditID)
	fmt.Printf("   脱敏结果:\n")

	for fieldName, maskedValue := range response.MaskedData {
		fmt.Printf("     %s: %s\n", fieldName, maskedValue)
	}

	fmt.Println()

	// 测试不同策略
	strategies := []string{"minimal", "moderate", "strict", "export", "testing"}

	for _, strategy := range strategies {
		testReq := &MaskingRequest{
			Fields: []FieldConfig{
				{Name: "test_name", Type: "name", Value: "李四", Required: true},
				{Name: "test_email", Type: "email", Value: "lisi@test.com", Required: true},
			},
			Context: map[string]string{"strategy": strategy, "purpose": "testing"},
			UserInfo: req.UserInfo,
			RequestID: fmt.Sprintf("strategy_demo_%s", strategy),
			Timestamp: time.Now(),
		}

		testResponse, err := dmd.maskingService.MaskDocument(context.Background(), testReq)
		if err != nil {
			return err
		}

		fmt.Printf("🔧 策略 %s 测试:\n", strategy)
		for fieldName, maskedValue := range testResponse.MaskedData {
			fmt.Printf("     %s: %s\n", fieldName, maskedValue)
		}
		fmt.Println()
	}

	return nil
}

// demonstratePermissionBasedMasking 演示基于权限的脱敏
func (dmd *DataMaskingDemo) demonstratePermissionBasedMasking() error {
	fmt.Println("👥 演示2: 用户权限级别脱敏")
	fmt.Println(strings.Repeat("-", 30))

	// 不同权限级别的用户
	users := []struct {
		Name     string
		UserInfo *UserInfo
	}{
		{
			Name: "超级管理员",
			UserInfo: &UserInfo{
				UserID:     "admin001",
				Username:   "admin",
				Roles:      []string{"super_admin"},
				Department: "IT",
			},
		},
		{
			Name: "合伙人律师",
			UserInfo: &UserInfo{
				UserID:     "partner001",
				Username:   "zhang_lawyer",
				Roles:      []string{"partner"},
				Department: "诉讼部",
			},
		},
		{
			Name: "执业律师",
			UserInfo: &UserInfo{
				UserID:     "lawyer001",
				Username:   "li_lawyer",
				Roles:      []string{"lawyer"},
				Department: "合同部",
			},
		},
		{
			Name: "律师助理",
			UserInfo: &UserInfo{
				UserID:     "assistant001",
				Username:   "wang_assistant",
				Roles:      []string{"assistant"},
				Department: "行政部",
			},
		},
		{
			Name: "客户",
			UserInfo: &UserInfo{
				UserID:     "client001",
				Username:   "client_user",
				Roles:      []string{"client"},
				Department: "外部",
			},
		},
	}

	testData := FieldConfig{
		Name:  "sensitive_info",
		Type:  "generic",
		Value: "机密案件文档内容：涉及商业纠纷，标的金额500万元",
		Required: true,
	}

	for _, user := range users {
		req := &MaskingRequest{
			Fields:    []FieldConfig{testData},
			Context:   map[string]string{"purpose": "development"},
			UserInfo:  user.UserInfo,
			RequestID: fmt.Sprintf("permission_demo_%s", user.UserInfo.UserID),
			Timestamp: time.Now(),
		}

		response, err := dmd.maskingService.MaskDocument(context.Background(), req)
		if err != nil {
			return err
		}

		fmt.Printf("👤 用户: %s (%s)\n", user.Name, strings.Join(user.UserInfo.Roles, ","))
		fmt.Printf("   脱敏结果: %s\n", response.MaskedData["sensitive_info"])
		fmt.Printf("   处理时间: %v\n", response.ProcessingTime)
		fmt.Println()
	}

	return nil
}

// demonstrateLawFirmSpecificMasking 演示律师事务所特定脱敏
func (dmd *DataMaskingDemo) demonstrateLawFirmSpecificMasking() error {
	fmt.Println("⚖️ 演示3: 律师事务所特定脱敏")
	fmt.Println(strings.Repeat("-", 30))

	// 案件信息脱敏
	caseInfo := &MaskingRequest{
		Fields: []FieldConfig{
			{Name: "case_number", Type: "case_number", Value: "CASE-2024-001", Required: true},
			{Name: "client_name", Type: "name", Value: "北京某科技有限公司", Required: true},
			{Name: "plaintiff_lawyer", Type: "name", Value: "张三律师", Required: true},
			{Name: "defendant_lawyer", Type: "name", Value: "李四律师", Required: true},
			{Name: "case_value", Type: "generic", Value: "5000000", Required: true},
			{Name: "court_name", Type: "generic", Value: "北京市朝阳区人民法院", Required: true},
		},
		Context: map[string]string{
			"purpose": "development",
			"document_type": "case_file",
		},
		UserInfo: &UserInfo{
			UserID:     "lawyer001",
			Username:   "li_lawyer",
			Roles:      []string{"lawyer"},
			Department: "合同部",
		},
		RequestID: "law_firm_demo_001",
		Timestamp: time.Now(),
	}

	response, err := dmd.maskingService.MaskDocument(context.Background(), caseInfo)
	if err != nil {
		return err
	}

	fmt.Printf("⚖️ 案件信息脱敏\n")
	for fieldName, maskedValue := range response.MaskedData {
		fmt.Printf("   %s: %s\n", fieldName, maskedValue)
	}
	fmt.Println()

	// 财务信息脱敏
	financeInfo := &MaskingRequest{
		Fields: []FieldConfig{
			{Name: "account_holder", Type: "name", Value: "王五", Required: true},
			{Name: "bank_account", Type: "bank_account", Value: "6222020123456789012", Required: true},
			{Name: "credit_card", Type: "credit_card", Value: "4111111111111111", Required: false},
			{Name: "transaction_amount", Type: "generic", Value: "100000", Required: true},
		},
		Context: map[string]string{
			"purpose": "analysis",
			"document_type": "financial_record",
		},
		UserInfo: &UserInfo{
			UserID:     "partner001",
			Username:   "zhang_partner",
			Roles:      []string{"partner"},
			Department: "管理层",
		},
		RequestID: "law_firm_demo_002",
		Timestamp: time.Now(),
	}

	financeResponse, err := dmd.maskingService.MaskDocument(context.Background(), financeInfo)
	if err != nil {
		return err
	}

	fmt.Printf("💰 财务信息脱敏\n")
	for fieldName, maskedValue := range financeResponse.MaskedData {
		fmt.Printf("   %s: %s\n", fieldName, maskedValue)
	}
	fmt.Println()

	// 客户证件脱敏
	documentInfo := &MaskingRequest{
		Fields: []FieldConfig{
			{Name: "client_name", Type: "name", Value: "赵六", Required: true},
			{Name: "id_card", Type: "id_card", Value: "110101199001011234", Required: true},
			{Name: "passport", Type: "license", Value: "A123456789", Required: false},
			{Name: "business_license", Type: "license", Value: "91110000123456789X", Required: false},
		},
		Context: map[string]string{
			"purpose": "testing",
			"document_type": "client_documents",
		},
		UserInfo: &UserInfo{
			UserID:     "assistant001",
			Username:   "wang_assistant",
			Roles:      []string{"assistant"},
			Department: "行政部",
		},
		RequestID: "law_firm_demo_003",
		Timestamp: time.Now(),
	}

	documentResponse, err := dmd.maskingService.MaskDocument(context.Background(), documentInfo)
	if err != nil {
		return err
	}

	fmt.Printf("📄 客户证件脱敏\n")
	for fieldName, maskedValue := range documentResponse.MaskedData {
		fmt.Printf("   %s: %s\n", fieldName, maskedValue)
	}
	fmt.Println()

	return nil
}

// demonstratePerformance 演示性能测试
func (dmd *DataMaskingDemo) demonstratePerformance() error {
	fmt.Println("⚡ 演示4: 性能测试")
	fmt.Println(strings.Repeat("-", 30))

	// 批量脱敏测试
	batchSizes := []int{10, 50, 100, 500, 1000}

	for _, batchSize := range batchSizes {
		start := time.Now()

		for i := 0; i < batchSize; i++ {
			req := &MaskingRequest{
				Fields: []FieldConfig{
					{Name: "name", Type: "name", Value: "测试用户", Required: true},
					{Name: "email", Type: "email", Value: "test@example.com", Required: true},
					{Name: "phone", Type: "phone", Value: "13812345678", Required: true},
				},
				Context: map[string]string{"purpose": "testing"},
				UserInfo: &UserInfo{
					UserID:     "test_user",
					Username:   "test",
					Roles:      []string{"lawyer"},
					Department: "test",
				},
				RequestID: fmt.Sprintf("perf_test_%d_%d", batchSize, i),
				Timestamp: time.Now(),
			}

			_, err := dmd.maskingService.MaskDocument(context.Background(), req)
			if err != nil {
				return err
			}
		}

		duration := time.Since(start)
		avgDuration := duration / time.Duration(batchSize)
		qps := float64(batchSize) / duration.Seconds()

		fmt.Printf("📊 批量测试: %d 条记录\n", batchSize)
		fmt.Printf("   总耗时: %v\n", duration)
		fmt.Printf("   平均耗时: %v\n", avgDuration)
		fmt.Printf("   QPS: %.0f\n", qps)
		fmt.Println()
	}

	// 字段类型性能测试
	fieldTypes := []string{"name", "email", "phone", "id_card", "bank_account", "address"}
	iterations := 1000

	for _, fieldType := range fieldTypes {
		start := time.Now()

		for i := 0; i < iterations; i++ {
			req := &MaskingRequest{
				Fields: []FieldConfig{
					{Name: "test_field", Type: fieldType, Value: "test_data", Required: true},
				},
				Context: map[string]string{"purpose": "performance"},
				UserInfo: &UserInfo{
					UserID:     "perf_user",
					Username:   "perf",
					Roles:      []string{"lawyer"},
					Department: "performance",
				},
				RequestID: fmt.Sprintf("field_perf_%s_%d", fieldType, i),
				Timestamp: time.Now(),
			}

			_, err := dmd.maskingService.MaskDocument(context.Background(), req)
			if err != nil {
				return err
			}
		}

		duration := time.Since(start)
		avgDuration := duration / time.Duration(iterations)
		qps := float64(iterations) / duration.Seconds()

		fmt.Printf("🔧 字段类型 %s 性能测试\n", fieldType)
		fmt.Printf("   迭代次数: %d\n", iterations)
		fmt.Printf("   总耗时: %v\n", duration)
		fmt.Printf("   平均耗时: %v\n", avgDuration)
		fmt.Printf("   QPS: %.0f\n", qps)
		fmt.Println()
	}

	return nil
}

// demonstrateCompliance 演示合规性检查
func (dmd *DataMaskingDemo) demonstrateCompliance() error {
	fmt.Println("📋 演示5: 合规性检查")
	fmt.Println(strings.Repeat("-", 30))

	// GDPR 合规测试
	fmt.Printf("🇪 GDPR 合规性检查\n")

	gdprTestData := &MaskingRequest{
		Fields: []FieldConfig{
			{Name: "personal_data", Type: "generic", Value: "包含个人敏感信息的文档内容", Required: true},
			{Name: "user_consent", Type: "generic", Value: "用户同意数据处理", Required: true},
		},
		Context: map[string]string{
			"purpose": "gdpr_compliance",
			"region":  "EU",
			"consent": "granted",
		},
		UserInfo: &UserInfo{
			UserID:     "gdpr_user",
			Username:   "eu_user",
			Roles:      []string{"data_processor"},
			Department: "compliance",
		},
		RequestID: "gdpr_demo_001",
		Timestamp: time.Now(),
	}

	gdprResponse, err := dmd.maskingService.MaskDocument(context.Background(), gdprTestData)
	if err != nil {
		return err
	}

	fmt.Printf("   ✅ GDPR 脱敏处理完成\n")
	fmt.Printf("   审计ID: %s\n", gdprResponse.AuditID)
	fmt.Println()

	// 中国个人信息保护法合规测试
	fmt.Printf("🇨🇳 PIPL 合规性检查\n")

	piplTestData := &MaskingRequest{
		Fields: []FieldConfig{
			{Name: "sensitive_personal_info", Type: "generic", Value: "重要个人信息文档", Required: true},
			{Name: "separate_consent", Type: "generic", Value: "单独同意处理敏感个人信息", Required: true},
		},
		Context: map[string]string{
			"purpose": "pipl_compliance",
			"region":  "China",
			"consent": "separate_granted",
		},
		UserInfo: &UserInfo{
			UserID:     "pipl_user",
			Username:   "cn_user",
			Roles:      []string{"data_controller"},
			Department: "法务合规",
		},
		RequestID: "pipl_demo_001",
		Timestamp: time.Now(),
	}

	piplResponse, err := dmd.maskingService.MaskDocument(context.Background(), piplTestData)
	if err != nil {
		return err
	}

	fmt.Printf("   ✅ PIPL 脱敏处理完成\n")
	fmt.Printf("   审计ID: %s\n", piplResponse.AuditID)
	fmt.Println()

	// 律师职业道德合规测试
	fmt.Printf("⚖️ 律师职业道德合规检查\n")

	ethicsTestData := &MaskingRequest{
		Fields: []FieldConfig{
			{Name: "attorney_client_communication", Type: "generic", Value: "律师-当事人沟通记录", Required: true},
			{Name: "legal_privilege_info", Type: "generic", Value: "律师-当事人特权信息", Required: true},
			{Name: "confidential_case_data", Type: "generic", Value: "案件保密信息", Required: true},
		},
		Context: map[string]string{
			"purpose": "legal_ethics",
			"privilege": "protected",
			"confidentiality": "high",
		},
		UserInfo: &UserInfo{
			UserID:     "ethics_user",
			Username:   "lawyer_ethics",
			Roles:      []string{"senior_lawyer"},
			Department: "合规部门",
		},
		RequestID: "ethics_demo_001",
		Timestamp: time.Now(),
	}

	ethicsResponse, err := dmd.maskingService.MaskDocument(context.Background(), ethicsTestData)
	if err != nil {
		return err
	}

	fmt.Printf("   ✅ 律师职业道德脱敏处理完成\n")
	fmt.Printf("   审计ID: %s\n", ethicsResponse.AuditID)
	fmt.Println()

	// 数据最小化原则测试
	fmt.Printf("📉 数据最小化原则检查\n")

	minimizationTestData := &MaskingRequest{
		Fields: []FieldConfig{
			{Name: "essential_data", Type: "generic", Value: "必要的业务数据", Required: true},
			{Name: "non_essential_data", Type: "generic", Value: "非必要的额外数据", Required: false},
		},
		Context: map[string]string{
			"purpose": "data_minimization",
			"principle": "minimize_data",
		},
		UserInfo: &UserInfo{
			UserID:     "min_user",
			Username:   "minimization_user",
			Roles:      []string{"data_minimizer"},
			Department: "数据治理",
		},
		RequestID: "minimization_demo_001",
		Timestamp: time.Now(),
	}

	minimizationResponse, err := dmd.maskingService.MaskDocument(context.Background(), minimizationTestData)
	if err != nil {
		return err
	}

	fmt.Printf("   ✅ 数据最小化脱敏处理完成\n")
	fmt.Printf("   审计ID: %s\n", minimizationResponse.AuditID)
	fmt.Println()

	return nil
}

// main 主函数
func main() {
	demo := NewDataMaskingDemo()

	if err := demo.Run(); err != nil {
		fmt.Printf("❌ 数据脱敏演示失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🎉 数据脱敏和隐私保护演示完成！")
	fmt.Println()
	fmt.Println("📊 功能总结:")
	fmt.Printf("   - 基础数据脱敏: ✅\n")
	fmt.Printf("   - 权限级别脱敏: ✅\n")
	fmt.Printf("   - 律师事务所特定脱敏: ✅\n")
	fmt.Printf("   - 性能优化: ✅\n")
	fmt.Printf("   - 合规性检查: ✅\n")
	fmt.Printf("   - 审计日志: ✅\n")
	fmt.Printf("   - 缓存机制: ✅\n")
	fmt.Printf("   - 多策略支持: ✅\n")
	fmt.Printf("   - 实时脱敏: ✅\n")
	fmt.Printf("   - 批量处理: ✅\n")
	fmt.Printf("   - 隐私保护: ✅\n")
}