package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"law-oa-go/internal/models"
)

// CompanyAPIProvider 工商 API 提供商类型
type CompanyAPIProvider string

const (
	ProviderQichacha   CompanyAPIProvider = "qichacha"   // 企查查
	ProviderTianyancha CompanyAPIProvider = "tianyancha" // 天眼查
	ProviderMock       CompanyAPIProvider = "mock"       // 模拟数据（开发测试用）
)

// CompanyAPIService 工商 API 服务接口
type CompanyAPIService interface {
	// SearchCompany 搜索公司信息
	SearchCompany(ctx context.Context, keyword string, provider CompanyAPIProvider) (*CompanySearchResult, error)
	// GetCompanyDetail 获取公司详细信息
	GetCompanyDetail(ctx context.Context, companyName string, taxID string, provider CompanyAPIProvider) (*CompanyDetail, error)
	// GetShareholding 获取股权穿透信息
	GetShareholding(ctx context.Context, taxID string, provider CompanyAPIProvider) (*ShareholdingInfo, error)
	// GetRelatedCompanies 获取关联公司信息
	GetRelatedCompanies(ctx context.Context, taxID string, provider CompanyAPIProvider) ([]*RelatedCompany, error)
}

// CompanySearchResult 公司搜索结果
type CompanySearchResult struct {
	Success    bool                `json:"success"`
	Message    string              `json:"message"`
	TotalCount int                 `json:"totalCount"`
	Companies  []*CompanyBasicInfo `json:"companies"`
	Provider   CompanyAPIProvider  `json:"provider"`
	QueryTime  time.Time           `json:"queryTime"`
	DurationMs int                 `json:"durationMs"`
}

// CompanyBasicInfo 公司基本信息
type CompanyBasicInfo struct {
	Name              string `json:"name"`
	TaxID             string `json:"taxId"`             // 统一社会信用代码/税号
	Status            string `json:"status"`            // 状态：在业、注销等
	LegalPerson       string `json:"legalPerson"`       // 法定代表人
	RegisteredCapital string `json:"registeredCapital"` // 注册资本
	EstablishDate     string `json:"establishDate"`     // 成立日期
	Province          string `json:"province"`          // 省份
	City              string `json:"city"`              // 城市
	Industry          string `json:"industry"`          // 行业
}

// CompanyDetail 公司详细信息
type CompanyDetail struct {
	BasicInfo         *CompanyBasicInfo `json:"basicInfo"`
	BusinessScope     string            `json:"businessScope"`     // 经营范围
	RegisteredAddress string            `json:"registeredAddress"` // 注册地址
	Shareholders      []*Shareholder    `json:"shareholders"`      // 股东信息
	Executives        []*Executive      `json:"executives"`        // 高管信息
	Changes           []*CompanyChange  `json:"changes"`           // 变更记录
}

// Shareholder 股东信息
type Shareholder struct {
	Name          string  `json:"name"`
	Type          string  `json:"type"`          // 股东类型：个人/企业
	HolderRatio   float64 `json:"holderRatio"`   // 持股比例
	CapitalAmount string  `json:"capitalAmount"` // 认缴出资额
}

// Executive 高管信息
type Executive struct {
	Name     string `json:"name"`
	Position string `json:"position"` // 职位
}

// CompanyChange 公司变更记录
type CompanyChange struct {
	ChangeDate    string `json:"changeDate"`
	ChangeItem    string `json:"changeItem"`
	BeforeContent string `json:"beforeContent"`
	AfterContent  string `json:"afterContent"`
}

// ShareholdingInfo 股权穿透信息
type ShareholdingInfo struct {
	TaxID              string             `json:"taxId"`
	CompanyName        string             `json:"companyName"`
	DirectShareholders []*Shareholder     `json:"directShareholders"` // 直接股东
	BeneficialOwners   []*BeneficialOwner `json:"beneficialOwners"`   // 最终受益人
	UltimateController string             `json:"ultimateController"` // 最终控制人
	PenetrationDepth   int                `json:"penetrationDepth"`   // 穿透层级
	RelatedCompanies   []*RelatedCompany  `json:"relatedCompanies"`   // 关联公司
}

// BeneficialOwner 最终受益人
type BeneficialOwner struct {
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Ownership float64 `json:"ownership"` // 最终受益比例
}

// RelatedCompany 关联公司
type RelatedCompany struct {
	Name         string  `json:"name"`
	TaxID        string  `json:"taxId"`
	RelationType string  `json:"relationType"` // 关系类型：投资/被投资/同一控制等
	RelationDesc string  `json:"relationDesc"` // 关系描述
	Ratio        float64 `json:"ratio"`        // 关联比例
}

// companyAPIService 工商 API 服务实现
type companyAPIService struct {
	httpClient   *http.Client
	config       *CompanyAPIConfig
	callRecorder APICallRecorder
}

// CompanyAPIConfig API 配置
type CompanyAPIConfig struct {
	QichachaToken     string        `json:"qichachaToken"`
	TianyanchaToken   string        `json:"tianyanchaToken"`
	BaseURLQichacha   string        `json:"baseUrlQichacha"`
	BaseURLTianyancha string        `json:"baseUrlTianyancha"`
	Timeout           time.Duration `json:"timeout"`
	EnableCache       bool          `json:"enableCache"`
	CacheTTL          time.Duration `json:"cacheTTL"`
	MaxRetries        int           `json:"maxRetries"`
}

// APICallRecorder API 调用记录器接口
type APICallRecorder interface {
	RecordCall(ctx context.Context, record *models.CompanyAPICall) error
}

// NewCompanyAPIService 创建新的工商 API 服务
func NewCompanyAPIService(config *CompanyAPIConfig, recorder APICallRecorder) CompanyAPIService {
	if config == nil {
		config = DefaultCompanyAPIConfig()
	}

	// 创建 HTTP 客户端
	httpClient := &http.Client{
		Timeout: config.Timeout,
		Transport: &loggingTransport{
			transport: http.DefaultTransport,
		},
	}

	return &companyAPIService{
		httpClient:   httpClient,
		config:       config,
		callRecorder: recorder,
	}
}

// DefaultCompanyAPIConfig 默认 API 配置
func DefaultCompanyAPIConfig() *CompanyAPIConfig {
	return &CompanyAPIConfig{
		BaseURLQichacha:   "https://api.qichacha.com",
		BaseURLTianyancha: "https://open.api.tianyancha.com",
		Timeout:           30 * time.Second,
		EnableCache:       true,
		CacheTTL:          24 * time.Hour,
		MaxRetries:        3,
	}
}

// loggingTransport 日志记录传输层
type loggingTransport struct {
	transport http.RoundTripper
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := t.transport.RoundTrip(req)
	duration := time.Since(start)

	log.Printf("[API] %s %s - Status: %d, Duration: %v",
		req.Method,
		req.URL.String(),
		func() int {
			if resp != nil {
				return resp.StatusCode
			}
			return 0
		}(),
		duration,
	)

	return resp, err
}

// SearchCompany 搜索公司信息
func (s *companyAPIService) SearchCompany(ctx context.Context, keyword string, provider CompanyAPIProvider) (*CompanySearchResult, error) {
	startTime := time.Now()

	// 验证输入
	if strings.TrimSpace(keyword) == "" {
		return nil, fmt.Errorf("搜索关键词不能为空")
	}

	log.Printf("🔍 开始搜索公司: keyword=%s, provider=%s", keyword, provider)

	// 根据提供商选择不同的实现
	var result *CompanySearchResult
	var err error

	switch provider {
	case ProviderQichacha:
		result, err = s.searchQichacha(ctx, keyword)
	case ProviderTianyancha:
		result, err = s.searchTianyancha(ctx, keyword)
	case ProviderMock:
		result, err = s.searchMock(ctx, keyword)
	default:
		// 默认使用模拟数据
		result, err = s.searchMock(ctx, keyword)
	}

	duration := time.Since(startTime)
	if result != nil {
		result.DurationMs = int(duration.Milliseconds())
		result.QueryTime = startTime
	}

	// 记录 API 调用
	s.recordAPICall(ctx, provider, "search", keyword, result, err, duration)

	if err != nil {
		log.Printf("❌ 搜索公司失败: %v", err)
		return nil, fmt.Errorf("搜索公司失败: %w", err)
	}

	log.Printf("✅ 搜索公司成功: 找到 %d 条结果", result.TotalCount)
	return result, nil
}

// GetCompanyDetail 获取公司详细信息
func (s *companyAPIService) GetCompanyDetail(ctx context.Context, companyName string, taxID string, provider CompanyAPIProvider) (*CompanyDetail, error) {
	startTime := time.Now()

	log.Printf("🔍 获取公司详情: name=%s, taxID=%s, provider=%s", companyName, taxID, provider)

	var detail *CompanyDetail
	var err error

	switch provider {
	case ProviderQichacha:
		detail, err = s.getDetailQichacha(ctx, companyName, taxID)
	case ProviderTianyancha:
		detail, err = s.getDetailTianyancha(ctx, companyName, taxID)
	case ProviderMock:
		detail, err = s.getDetailMock(ctx, companyName, taxID)
	default:
		detail, err = s.getDetailMock(ctx, companyName, taxID)
	}

	duration := time.Since(startTime)

	// 记录 API 调用
	keyword := companyName
	if taxID != "" {
		keyword = taxID
	}
	s.recordAPICall(ctx, provider, "detail", keyword, detail, err, duration)

	if err != nil {
		log.Printf("❌ 获取公司详情失败: %v", err)
		return nil, fmt.Errorf("获取公司详情失败: %w", err)
	}

	log.Printf("✅ 获取公司详情成功")
	return detail, nil
}

// GetShareholding 获取股权穿透信息
func (s *companyAPIService) GetShareholding(ctx context.Context, taxID string, provider CompanyAPIProvider) (*ShareholdingInfo, error) {
	startTime := time.Now()

	log.Printf("🔍 获取股权穿透信息: taxID=%s, provider=%s", taxID, provider)

	var info *ShareholdingInfo
	var err error

	switch provider {
	case ProviderQichacha:
		info, err = s.getShareholdingQichacha(ctx, taxID)
	case ProviderTianyancha:
		info, err = s.getShareholdingTianyancha(ctx, taxID)
	case ProviderMock:
		info, err = s.getShareholdingMock(ctx, taxID)
	default:
		info, err = s.getShareholdingMock(ctx, taxID)
	}

	duration := time.Since(startTime)
	s.recordAPICall(ctx, provider, "shareholding", taxID, info, err, duration)

	if err != nil {
		log.Printf("❌ 获取股权穿透信息失败: %v", err)
		return nil, fmt.Errorf("获取股权穿透信息失败: %w", err)
	}

	log.Printf("✅ 获取股权穿透信息成功: 穿透层级=%d", info.PenetrationDepth)
	return info, nil
}

// GetRelatedCompanies 获取关联公司信息
func (s *companyAPIService) GetRelatedCompanies(ctx context.Context, taxID string, provider CompanyAPIProvider) ([]*RelatedCompany, error) {
	startTime := time.Now()

	log.Printf("🔍 获取关联公司信息: taxID=%s, provider=%s", taxID, provider)

	var companies []*RelatedCompany
	var err error

	switch provider {
	case ProviderQichacha:
		companies, err = s.getRelatedQichacha(ctx, taxID)
	case ProviderTianyancha:
		companies, err = s.getRelatedTianyancha(ctx, taxID)
	case ProviderMock:
		companies, err = s.getRelatedMock(ctx, taxID)
	default:
		companies, err = s.getRelatedMock(ctx, taxID)
	}

	duration := time.Since(startTime)
	s.recordAPICall(ctx, provider, "related", taxID, companies, err, duration)

	if err != nil {
		log.Printf("❌ 获取关联公司信息失败: %v", err)
		return nil, fmt.Errorf("获取关联公司信息失败: %w", err)
	}

	log.Printf("✅ 获取关联公司信息成功: 找到 %d 家关联公司", len(companies))
	return companies, nil
}

// ============================================================================
// 企查查 API 实现
// ============================================================================

func (s *companyAPIService) searchQichacha(ctx context.Context, keyword string) (*CompanySearchResult, error) {
	if s.config.QichachaToken == "" {
		return nil, fmt.Errorf("企查查 API Token 未配置")
	}

	// 构建请求 URL
	apiURL := fmt.Sprintf("%s/ECIV4/Search", s.config.BaseURLQichacha)
	params := url.Values{}
	params.Set("keyword", keyword)
	params.Set("pageSize", "20")

	fullURL := apiURL + "?" + params.Encode()

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+s.config.QichachaToken)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API 返回错误: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var response struct {
		ErrorCode string `json:"error_code"`
		Reason    string `json:"reason"`
		Data      struct {
			TotalCount int                      `json:"total"`
			List       []map[string]interface{} `json:"list"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if response.ErrorCode != "0" {
		return nil, fmt.Errorf("API 返回错误: %s", response.Reason)
	}

	// 转换结果
	result := &CompanySearchResult{
		Success:    true,
		Message:    "查询成功",
		TotalCount: response.Data.TotalCount,
		Companies:  make([]*CompanyBasicInfo, 0),
		Provider:   ProviderQichacha,
	}

	for _, item := range response.Data.List {
		company := &CompanyBasicInfo{
			Name:              getStringValue(item, "name"),
			TaxID:             getStringValue(item, "creditCode"),
			Status:            getStringValue(item, "status"),
			LegalPerson:       getStringValue(item, "operName"),
			RegisteredCapital: getStringValue(item, "regCapital"),
			EstablishDate:     getStringValue(item, "estiblishTime"),
			Province:          getStringValue(item, "province"),
			City:              getStringValue(item, "city"),
			Industry:          getStringValue(item, "industry"),
		}
		result.Companies = append(result.Companies, company)
	}

	return result, nil
}

func (s *companyAPIService) getDetailQichacha(ctx context.Context, companyName, taxID string) (*CompanyDetail, error) {
	// 企查查详情 API 实现
	// 实际生产环境需要调用真实的企查查 API
	return nil, fmt.Errorf("企查查详情 API 待实现")
}

func (s *companyAPIService) getShareholdingQichacha(ctx context.Context, taxID string) (*ShareholdingInfo, error) {
	// 企查查股权穿透 API 实现
	return nil, fmt.Errorf("企查查股权穿透 API 待实现")
}

func (s *companyAPIService) getRelatedQichacha(ctx context.Context, taxID string) ([]*RelatedCompany, error) {
	// 企查查关联公司 API 实现
	return nil, fmt.Errorf("企查查关联公司 API 待实现")
}

// ============================================================================
// 天眼查 API 实现
// ============================================================================

func (s *companyAPIService) searchTianyancha(ctx context.Context, keyword string) (*CompanySearchResult, error) {
	if s.config.TianyanchaToken == "" {
		return nil, fmt.Errorf("天眼查 API Token 未配置")
	}

	// 天眼查搜索 API 实现
	// 实际生产环境需要调用真实的天眼查 API
	return nil, fmt.Errorf("天眼查 API 待实现")
}

func (s *companyAPIService) getDetailTianyancha(ctx context.Context, companyName, taxID string) (*CompanyDetail, error) {
	return nil, fmt.Errorf("天眼查详情 API 待实现")
}

func (s *companyAPIService) getShareholdingTianyancha(ctx context.Context, taxID string) (*ShareholdingInfo, error) {
	return nil, fmt.Errorf("天眼查股权穿透 API 待实现")
}

func (s *companyAPIService) getRelatedTianyancha(ctx context.Context, taxID string) ([]*RelatedCompany, error) {
	return nil, fmt.Errorf("天眼查关联公司 API 待实现")
}

// ============================================================================
// 模拟数据实现（开发测试用）
// ============================================================================

func (s *companyAPIService) searchMock(ctx context.Context, keyword string) (*CompanySearchResult, error) {
	// 模拟数据
	mockCompanies := []*CompanyBasicInfo{
		{
			Name:              "腾讯科技（深圳）有限公司",
			TaxID:             "91440300123456789X",
			Status:            "在业",
			LegalPerson:       "马化腾",
			RegisteredCapital: "5000万元人民币",
			EstablishDate:     "2000-02-24",
			Province:          "广东",
			City:              "深圳",
			Industry:          "软件和信息技术服务业",
		},
		{
			Name:              "阿里巴巴（中国）有限公司",
			TaxID:             "91330100567890123Y",
			Status:            "在业",
			LegalPerson:       "张勇",
			RegisteredCapital: "10000万美元",
			EstablishDate:     "2007-03-26",
			Province:          "浙江",
			City:              "杭州",
			Industry:          "软件和信息技术服务业",
		},
		{
			Name:              "北京百度网讯科技有限公司",
			TaxID:             "91110100789012345Z",
			Status:            "在业",
			LegalPerson:       "梁志祥",
			RegisteredCapital: "100万元人民币",
			EstablishDate:     "2001-06-05",
			Province:          "北京",
			City:              "北京",
			Industry:          "软件和信息技术服务业",
		},
	}

	// 过滤匹配的结果
	var filtered []*CompanyBasicInfo
	for _, company := range mockCompanies {
		if strings.Contains(company.Name, keyword) || strings.Contains(company.TaxID, keyword) {
			filtered = append(filtered, company)
		}
	}

	return &CompanySearchResult{
		Success:    true,
		Message:    "查询成功",
		TotalCount: len(filtered),
		Companies:  filtered,
		Provider:   ProviderMock,
	}, nil
}

func (s *companyAPIService) getDetailMock(ctx context.Context, companyName, taxID string) (*CompanyDetail, error) {
	// 模拟公司详情
	return &CompanyDetail{
		BasicInfo: &CompanyBasicInfo{
			Name:              companyName,
			TaxID:             taxID,
			Status:            "在业",
			LegalPerson:       "模拟法人",
			RegisteredCapital: "1000万元人民币",
			EstablishDate:     "2010-01-01",
			Province:          "北京",
			City:              "北京",
			Industry:          "软件和信息技术服务业",
		},
		BusinessScope:     "技术开发、技术咨询、技术服务；销售计算机、软件及辅助设备。",
		RegisteredAddress: "北京市海淀区模拟路123号",
		Shareholders: []*Shareholder{
			{
				Name:          "股东A",
				Type:          "企业",
				HolderRatio:   60.0,
				CapitalAmount: "600万元",
			},
			{
				Name:          "股东B",
				Type:          "个人",
				HolderRatio:   40.0,
				CapitalAmount: "400万元",
			},
		},
		Executives: []*Executive{
			{
				Name:     "张三",
				Position: "执行董事",
			},
			{
				Name:     "李四",
				Position: "经理",
			},
		},
	}, nil
}

func (s *companyAPIService) getShareholdingMock(ctx context.Context, taxID string) (*ShareholdingInfo, error) {
	// 模拟股权穿透数据
	return &ShareholdingInfo{
		TaxID:       taxID,
		CompanyName: "模拟公司有限公司",
		DirectShareholders: []*Shareholder{
			{
				Name:          "控股股东A",
				Type:          "企业",
				HolderRatio:   51.0,
				CapitalAmount: "510万元",
			},
		},
		BeneficialOwners: []*BeneficialOwner{
			{
				Name:      "最终受益人A",
				Type:      "个人",
				Ownership: 51.0,
			},
		},
		UltimateController: "最终受益人A",
		PenetrationDepth:   2,
		RelatedCompanies: []*RelatedCompany{
			{
				Name:         "关联公司A",
				TaxID:        "91110000XXXX",
				RelationType: "投资",
				RelationDesc: "控股子公司",
				Ratio:        100.0,
			},
		},
	}, nil
}

func (s *companyAPIService) getRelatedMock(ctx context.Context, taxID string) ([]*RelatedCompany, error) {
	// 模拟关联公司数据
	return []*RelatedCompany{
		{
			Name:         "关联公司A",
			TaxID:        "91110000XXXX001",
			RelationType: "投资",
			RelationDesc: "控股子公司",
			Ratio:        100.0,
		},
		{
			Name:         "关联公司B",
			TaxID:        "91110000XXXX002",
			RelationType: "被投资",
			RelationDesc: "参股公司",
			Ratio:        30.0,
		},
	}, nil
}

// ============================================================================
// 辅助方法
// ============================================================================

// recordAPICall 记录 API 调用
func (s *companyAPIService) recordAPICall(ctx context.Context, provider CompanyAPIProvider, endpoint, keyword string, result interface{}, err error, duration time.Duration) {
	if s.callRecorder == nil {
		return
	}

	var matchedCompanyName string
	var matchedCompanyTaxID string
	var responseData models.JSON
	var responseStatus string
	var errorMessage string

	if err != nil {
		responseStatus = "failed"
		errorMessage = err.Error()
	} else {
		responseStatus = "success"
		// 将结果序列化为 JSON
		if data, err := json.Marshal(result); err == nil {
			responseData = models.JSON(map[string]interface{}{
				"data": string(data),
			})
		}
	}

	// 尝试提取匹配的公司名称和税号
	if keyword != "" {
		if len(keyword) == 18 { // 统一社会信用代码长度
			matchedCompanyTaxID = keyword
		} else {
			matchedCompanyName = keyword
		}
	}

	record := &models.CompanyAPICall{
		APIProvider:         string(provider),
		APIEndpoint:         endpoint,
		SearchKeyword:       keyword,
		MatchedCompanyName:  matchedCompanyName,
		MatchedCompanyTaxID: matchedCompanyTaxID,
		ResponseStatus:      responseStatus,
		ResponseData:        responseData,
		ErrorMessage:        errorMessage,
		CallDurationMs:      func() *int { i := int(duration.Milliseconds()); return &i }(),
	}

	// 异步记录，不影响主流程
	go func() {
		if recordErr := s.callRecorder.RecordCall(context.Background(), record); recordErr != nil {
			log.Printf("⚠️ 记录 API 调用失败: %v", recordErr)
		}
	}()
}

// getStringValue 从 map 中获取字符串值
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// ============================================================================
// API 调用记录器实现
// ============================================================================

// DatabaseAPICallRecorder 数据库 API 调用记录器
type DatabaseAPICallRecorder struct {
	db APICallRecordDB
}

// APICallRecordDB 数据库接口
type APICallRecordDB interface {
	Create(ctx context.Context, call *models.CompanyAPICall) error
}

// NewDatabaseAPICallRecorder 创建数据库记录器
func NewDatabaseAPICallRecorder(db APICallRecordDB) APICallRecorder {
	return &DatabaseAPICallRecorder{db: db}
}

// RecordCall 记录 API 调用
func (r *DatabaseAPICallRecorder) RecordCall(ctx context.Context, record *models.CompanyAPICall) error {
	return r.db.Create(ctx, record)
}

// ============================================================================
// 工具函数
// ============================================================================

// NormalizeCompanyName 标准化公司名称
func NormalizeCompanyName(name string) string {
	// 移除常见的公司后缀
	suffixes := []string{
		"有限公司", "股份有限公司", "集团有限公司",
		"有限责任公司", "控股有限公司", "科技有限公司",
	}

	result := name
	for _, suffix := range suffixes {
		result = strings.TrimSuffix(result, suffix)
	}

	return strings.TrimSpace(result)
}

// ParseTaxID 解析税号/统一社会信用代码
func ParseTaxID(taxID string) (bool, error) {
	// 统一社会信用代码为18位
	if len(taxID) != 18 {
		return false, fmt.Errorf("税号长度不正确")
	}

	// 简单验证：第1位为数字或大写字母
	firstChar := taxID[0]
	if !((firstChar >= '0' && firstChar <= '9') || (firstChar >= 'A' && firstChar <= 'Z')) {
		return false, fmt.Errorf("税号格式不正确")
	}

	return true, nil
}

// CalculatePenetrationDepth 计算股权穿透深度
func CalculatePenetrationDepth(shareholders []*Shareholder) int {
	maxDepth := 0
	for _, shareholder := range shareholders {
		if shareholder.Type == "企业" {
			// 企业股东需要进一步穿透
			maxDepth = max(maxDepth, 1)
		}
	}
	return maxDepth
}

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GetRiskLevelByShareholding 根据持股比例获取风险等级
func GetRiskLevelByShareholding(ratio float64) string {
	switch {
	case ratio >= 50:
		return "CRITICAL" // 控股股东
	case ratio >= 25:
		return "HIGH" // 重要股东
	case ratio >= 10:
		return "MEDIUM" // 一般股东
	default:
		return "LOW" // 小股东
	}
}

// FormatRatio 格式化比例显示
func FormatRatio(ratio float64) string {
	return strconv.FormatFloat(ratio, 'f', 2, 64) + "%"
}
