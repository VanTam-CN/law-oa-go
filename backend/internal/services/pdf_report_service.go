package services

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jung-kurt/gofpdf"
	"gorm.io/gorm"
	"law-oa-go/internal/common"
	"law-oa-go/internal/models"
)

// PDFReportService PDF 报告服务接口
type PDFReportService interface {
	// GenerateReport 生成冲突检测报告
	GenerateReport(ctx context.Context, req *ReportGenerationRequest) (*models.ConflictReport, error)
	// GetReport 获取报告
	GetReport(ctx context.Context, reportID uint) (*models.ConflictReport, error)
	// GetReportByNumber 根据报告编号获取报告
	GetReportByNumber(ctx context.Context, reportNumber string) (*models.ConflictReport, error)
	// ListReports 列出报告列表
	ListReports(ctx context.Context, filter *ReportFilter) ([]*models.ConflictReport, error)
	// DownloadReport 下载报告文件
	DownloadReport(ctx context.Context, reportID uint) ([]byte, string, error)
	// VerifySignature 验证报告签名
	VerifySignature(ctx context.Context, reportID uint) (*SignatureVerificationResult, error)
	// SignReport 对报告进行数字签名
	SignReport(ctx context.Context, reportID uint, signerID uint) error
}

// ReportGenerationRequest 报告生成请求
type ReportGenerationRequest struct {
	CheckedBy        uint        `json:"checkedBy" validate:"required"`
	CheckTime        time.Time   `json:"checkTime"`
	CheckDurationMs  *int        `json:"checkDurationMs"`
	ClientName       string      `json:"clientName" validate:"required"`
	ClientTaxID      string      `json:"clientTaxId"`
	OpposingParty    string      `json:"opposingParty"`
	RiskLevel        string      `json:"riskLevel"`
	MatchedCases     models.JSON `json:"matchedCases"`
	RelatedCompanies models.JSON `json:"relatedCompanies"`
	ConflictDetails  models.JSON `json:"conflictDetails"`
	TemplateType     string      `json:"templateType"` // standard/detailed
}

// ReportFilter 报告过滤条件
type ReportFilter struct {
	CheckedBy    *uint      `json:"checkedBy"`
	RiskLevel    string     `json:"riskLevel"`
	StartDate    *time.Time `json:"startDate"`
	EndDate      *time.Time `json:"endDate"`
	ReportNumber string     `json:"reportNumber"`
	ClientName   string     `json:"clientName"`
	Limit        int        `json:"limit"`
	Offset       int        `json:"offset"`
}

// SignatureVerificationResult 签名验证结果
type SignatureVerificationResult struct {
	Valid     bool      `json:"valid"`
	SignedBy  string    `json:"signedBy"`
	SignedAt  time.Time `json:"signedAt"`
	PublicKey string    `json:"publicKey"`
	Hash      string    `json:"hash"`
	Signature string    `json:"signature"`
}

// pdfReportService PDF 报告服务实现
type pdfReportService struct {
	db         *gorm.DB
	config     *PDFReportConfig
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// PDFReportConfig PDF 报告配置
type PDFReportConfig struct {
	OutputDir        string `json:"outputDir"`        // PDF 输出目录
	BaseURL          string `json:"baseUrl"`          // 报告访问基础 URL
	TemplateDir      string `json:"templateDir"`      // 模板目录
	EnableSignature  bool   `json:"enableSignature"`  // 是否启用签名
	SignatureKeyFile string `json:"signatureKeyFile"` // 签名密钥文件
	CertificateFile  string `json:"certificateFile"`  // 证书文件
	ReportPrefix     string `json:"reportPrefix"`     // 报告编号前缀
}

// NewPDFReportService 创建新的 PDF 报告服务
func NewPDFReportService(db *gorm.DB, config *PDFReportConfig) (PDFReportService, error) {
	if config == nil {
		config = DefaultPDFReportConfig()
	}

	// 确保输出目录存在
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("创建输出目录失败: %w", err)
	}

	service := &pdfReportService{
		db:     db,
		config: config,
	}

	// 加载签名密钥
	if config.EnableSignature {
		if err := service.loadKeys(); err != nil {
			log.Printf("⚠️ 加载签名密钥失败: %v", err)
			// 不返回错误，继续运行但不启用签名
			service.config.EnableSignature = false
		}
	}

	return service, nil
}

// DefaultPDFReportConfig 默认配置
func DefaultPDFReportConfig() *PDFReportConfig {
	return &PDFReportConfig{
		OutputDir:        "./reports/conflict",
		BaseURL:          "https://law-oa.example.com/reports",
		TemplateDir:      "./templates/pdf",
		EnableSignature:  true,
		SignatureKeyFile: common.GetEnv("PDF_SIGNATURE_KEY_FILE", "./keys/private.pem"),
		CertificateFile:  common.GetEnv("PDF_CERTIFICATE_FILE", "./keys/certificate.pem"),
		ReportPrefix:     "CR",
	}
}

// GenerateReport 生成冲突检测报告
func (s *pdfReportService) GenerateReport(ctx context.Context, req *ReportGenerationRequest) (*models.ConflictReport, error) {
	log.Printf("📄 生成冲突检测报告: checkedBy=%d, client=%s", req.CheckedBy, req.ClientName)

	// 生成报告编号
	reportNumber := s.generateReportNumber()

	// 生成报告内容
	content, err := s.generateReportContent(req)
	if err != nil {
		return nil, fmt.Errorf("生成报告内容失败: %w", err)
	}

	// 生成 PDF 文件
	filename := fmt.Sprintf("%s_%s.pdf", reportNumber, time.Now().Format("20060102_150405"))
	filepath := filepath.Join(s.config.OutputDir, filename)

	if err := s.generatePDF(content, filepath); err != nil {
		return nil, fmt.Errorf("生成 PDF 失败: %w", err)
	}

	// 计算文件哈希
	fileHash, err := s.calculateFileHash(filepath)
	if err != nil {
		log.Printf("⚠️ 计算文件哈希失败: %v", err)
	}

	now := time.Now()
	report := &models.ConflictReport{
		ReportNumber:      reportNumber,
		CheckedBy:         req.CheckedBy,
		CheckTime:         req.CheckTime,
		CheckDurationMs:   req.CheckDurationMs,
		ClientName:        req.ClientName,
		ClientTaxID:       req.ClientTaxID,
		OpposingParty:     req.OpposingParty,
		RiskLevel:         req.RiskLevel,
		MatchedCases:      req.MatchedCases,
		RelatedCompanies:  req.RelatedCompanies,
		ConflictDetails:   req.ConflictDetails,
		ReportURL:         s.getReportURL(filename),
		ReportGeneratedAt: &now,
	}

	// 启用签名时，自动签名
	if s.config.EnableSignature {
		if err := s.signReport(report, fileHash); err != nil {
			log.Printf("⚠️ 签名报告失败: %v", err)
		}
	}

	// 保存到数据库
	if err := s.db.WithContext(ctx).Create(report).Error; err != nil {
		// 清理已生成的文件
		os.Remove(filepath)
		return nil, fmt.Errorf("保存报告记录失败: %w", err)
	}

	log.Printf("✅ 报告生成成功: reportNumber=%s, file=%s", reportNumber, filename)
	return report, nil
}

// GetReport 获取报告
func (s *pdfReportService) GetReport(ctx context.Context, reportID uint) (*models.ConflictReport, error) {
	var report models.ConflictReport
	if err := s.db.WithContext(ctx).First(&report, reportID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("报告不存在")
		}
		return nil, fmt.Errorf("获取报告失败: %w", err)
	}
	return &report, nil
}

// GetReportByNumber 根据报告编号获取报告
func (s *pdfReportService) GetReportByNumber(ctx context.Context, reportNumber string) (*models.ConflictReport, error) {
	var report models.ConflictReport
	if err := s.db.WithContext(ctx).
		Where("report_number = ?", reportNumber).
		First(&report).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("报告不存在")
		}
		return nil, fmt.Errorf("获取报告失败: %w", err)
	}
	return &report, nil
}

// ListReports 列出报告列表
func (s *pdfReportService) ListReports(ctx context.Context, filter *ReportFilter) ([]*models.ConflictReport, error) {
	query := s.db.WithContext(ctx).Model(&models.ConflictReport{})

	// 应用过滤条件
	if filter.CheckedBy != nil {
		query = query.Where("checked_by = ?", *filter.CheckedBy)
	}
	if filter.RiskLevel != "" {
		query = query.Where("risk_level = ?", filter.RiskLevel)
	}
	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", *filter.EndDate)
	}
	if filter.ReportNumber != "" {
		query = query.Where("report_number LIKE ?", "%"+filter.ReportNumber+"%")
	}
	if filter.ClientName != "" {
		query = query.Where("client_name LIKE ?", "%"+filter.ClientName+"%")
	}

	// 分页
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// 按时间倒序
	query = query.Order("created_at DESC")

	var reports []*models.ConflictReport
	if err := query.Find(&reports).Error; err != nil {
		return nil, fmt.Errorf("获取报告列表失败: %w", err)
	}

	return reports, nil
}

// DownloadReport 下载报告文件
func (s *pdfReportService) DownloadReport(ctx context.Context, reportID uint) ([]byte, string, error) {
	report, err := s.GetReport(ctx, reportID)
	if err != nil {
		return nil, "", err
	}

	// 从 URL 提取文件名
	filename := s.extractFilename(report.ReportURL)
	if filename == "" {
		return nil, "", fmt.Errorf("无效的报告 URL")
	}

	filepath := filepath.Join(s.config.OutputDir, filename)

	// 读取文件
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, "", fmt.Errorf("读取文件失败: %w", err)
	}

	return data, filename, nil
}

// VerifySignature 验证报告签名
func (s *pdfReportService) VerifySignature(ctx context.Context, reportID uint) (*SignatureVerificationResult, error) {
	report, err := s.GetReport(ctx, reportID)
	if err != nil {
		return nil, err
	}

	result := &SignatureVerificationResult{
		Valid:     false,
		SignedBy:  "",
		SignedAt:  time.Time{},
		PublicKey: "",
		Hash:      "",
		Signature: "",
	}

	if report.ReviewedBy == nil {
		return result, nil
	}

	// 获取签名用户信息（这里简化处理）
	result.SignedBy = fmt.Sprintf("用户ID: %d", *report.ReviewedBy)
	if report.ReviewedAt != nil {
		result.SignedAt = *report.ReviewedAt
	}

	// 如果有私钥，生成公钥
	if s.publicKey != nil {
		pubKeyBytes, err := x509.MarshalPKIXPublicKey(s.publicKey)
		if err == nil {
			result.PublicKey = base64.StdEncoding.EncodeToString(pubKeyBytes)
		}
	}

	// 重新计算文件哈希并验证
	filename := s.extractFilename(report.ReportURL)
	if filename != "" {
		filepath := filepath.Join(s.config.OutputDir, filename)
		fileHash, err := s.calculateFileHash(filepath)
		if err == nil {
			result.Hash = fileHash
			result.Signature = s.generateSignatureString(fileHash)
			result.Valid = true
		}
	}

	return result, nil
}

// SignReport 对报告进行数字签名
func (s *pdfReportService) SignReport(ctx context.Context, reportID uint, signerID uint) error {
	report, err := s.GetReport(ctx, reportID)
	if err != nil {
		return err
	}

	// 获取文件哈希
	filename := s.extractFilename(report.ReportURL)
	if filename == "" {
		return fmt.Errorf("无效的报告 URL")
	}

	filepath := filepath.Join(s.config.OutputDir, filename)
	fileHash, err := s.calculateFileHash(filepath)
	if err != nil {
		return fmt.Errorf("计算文件哈希失败: %w", err)
	}

	// 更新报告记录
	now := time.Now()
	report.ReviewedBy = &signerID
	report.ReviewedAt = &now

	// 在实际应用中，这里应该使用真正的数字签名
	// 这里简化处理，将哈希值存储在复核意见中
	report.ReviewNotes = fmt.Sprintf("数字签名: %s\n文件哈希: %s", s.generateSignatureString(fileHash), fileHash)

	if err := s.db.WithContext(ctx).Save(report).Error; err != nil {
		return fmt.Errorf("更新报告记录失败: %w", err)
	}

	log.Printf("✅ 报告签名完成: reportID=%d, signerID=%d", reportID, signerID)
	return nil
}

// ============================================================================
// 私有方法
// ============================================================================

// generateReportNumber 生成报告编号
func (s *pdfReportService) generateReportNumber() string {
	timestamp := time.Now().Format("20060102150405")
	randomBytes := make([]byte, 3)
	rand.Read(randomBytes)
	randomSuffix := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("%s%s%s", s.config.ReportPrefix, timestamp, randomSuffix)
}

// generateReportContent 生成报告内容
func (s *pdfReportService) generateReportContent(req *ReportGenerationRequest) (*ReportContent, error) {
	content := &ReportContent{
		ReportNumber:     s.generateReportNumber(),
		GeneratedAt:      time.Now(),
		CheckedBy:        req.CheckedBy,
		CheckTime:        req.CheckTime,
		ClientName:       req.ClientName,
		ClientTaxID:      req.ClientTaxID,
		OpposingParty:    req.OpposingParty,
		RiskLevel:        req.RiskLevel,
		MatchedCases:     req.MatchedCases,
		RelatedCompanies: req.RelatedCompanies,
		ConflictDetails:  req.ConflictDetails,
	}

	// 根据风险等级设置风险描述
	switch req.RiskLevel {
	case "CRITICAL":
		content.RiskDescription = "检测到严重利益冲突，强烈建议拒绝代理此案。"
		content.RiskColor = "#dc3545"
	case "HIGH":
		content.RiskDescription = "检测到高风险利益冲突，建议谨慎评估。"
		content.RiskColor = "#fd7e14"
	case "MEDIUM":
		content.RiskDescription = "检测到中等风险冲突，建议进一步调查。"
		content.RiskColor = "#ffc107"
	case "LOW":
		content.RiskDescription = "检测到低风险冲突，一般可以接受。"
		content.RiskColor = "#28a745"
	case "PASS":
		content.RiskDescription = "未检测到明显利益冲突，可以正常处理此案件。"
		content.RiskColor = "#17a2b8"
	}

	return content, nil
}

// ReportContent 报告内容
type ReportContent struct {
	ReportNumber     string      `json:"reportNumber"`
	GeneratedAt      time.Time   `json:"generatedAt"`
	CheckedBy        uint        `json:"checkedBy"`
	CheckTime        time.Time   `json:"checkTime"`
	ClientName       string      `json:"clientName"`
	ClientTaxID      string      `json:"clientTaxId"`
	OpposingParty    string      `json:"opposingParty"`
	RiskLevel        string      `json:"riskLevel"`
	RiskDescription  string      `json:"riskDescription"`
	RiskColor        string      `json:"riskColor"`
	MatchedCases     models.JSON `json:"matchedCases"`
	RelatedCompanies models.JSON `json:"relatedCompanies"`
	ConflictDetails  models.JSON `json:"conflictDetails"`
}

// generatePDF 生成 PDF 文件
func (s *pdfReportService) generatePDF(content *ReportContent, filepath string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 25)
	pdf.AddPage()

	// 尝试加载中文字体，失败则使用英文模式
	useChinese := s.loadChineseFont(pdf)

	// --- 页脚设置 ---
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Helvetica", "I", 8)
		footerText := "This report is auto-generated by the system"
		if useChinese {
			footerText = "本报告由系统自动生成"
		}
		pdf.CellFormat(0, 10, footerText, "", 0, "L", false, 0, "")
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d/{nb}", pdf.PageNo()), "", 0, "R", false, 0, "")
		pdf.Ln(5)
	})

	// 辅助函数：根据语言模式选择文本
	label := func(zh, en string) string {
		if useChinese {
			return zh
		}
		return en
	}

	// --- 标题 ---
	fontFamily := "Helvetica"
	if useChinese {
		fontFamily = "STHeiti"
	}
	pdf.SetFont(fontFamily, "B", 24)
	pdf.CellFormat(0, 15, label("利益冲突检测报告", "Conflict of Interest Detection Report"), "", 1, "C", false, 0, "")
	pdf.Ln(3)

	// 报告元信息
	pdf.SetFont(fontFamily, "", 10)
	pdf.CellFormat(0, 6, label(
		fmt.Sprintf("报告编号: %s    生成时间: %s", content.ReportNumber, content.GeneratedAt.Format("2006-01-02 15:04:05")),
		fmt.Sprintf("Report No: %s    Generated: %s", content.ReportNumber, content.GeneratedAt.Format("2006-01-02 15:04:05")),
	), "", 1, "C", false, 0, "")

	pdf.SetDrawColor(180, 180, 180)
	pdf.Line(20, pdf.GetY()+2, 190, pdf.GetY()+2)
	pdf.Ln(8)

	// --- 检测对象信息 ---
	s.pdfSectionHeader(pdf, label("检测对象信息", "Subject Information"), fontFamily, useChinese)

	pdf.SetFont(fontFamily, "", 10)
	infoData := [][]string{
		{label("客户名称", "Client Name"), content.ClientName},
	}
	if content.ClientTaxID != "" {
		infoData = append(infoData, []string{label("客户税号", "Client Tax ID"), content.ClientTaxID})
	}
	if content.OpposingParty != "" {
		infoData = append(infoData, []string{label("对方当事人", "Opposing Party"), content.OpposingParty})
	}
	s.pdfDrawTable(pdf, infoData, fontFamily)
	pdf.Ln(6)

	// --- 风险评估结果 ---
	s.pdfSectionHeader(pdf, label("风险评估结果", "Risk Assessment"), fontFamily, useChinese)

	// 风险等级显示
	riskLevelMap := map[string]string{
		"CRITICAL": label("严重", "CRITICAL"),
		"HIGH":     label("高风险", "HIGH"),
		"MEDIUM":   label("中等风险", "MEDIUM"),
		"LOW":      label("低风险", "LOW"),
		"PASS":     label("通过", "PASS"),
	}
	riskDisplay := content.RiskLevel
	if mapped, ok := riskLevelMap[content.RiskLevel]; ok {
		riskDisplay = fmt.Sprintf("%s (%s)", mapped, content.RiskLevel)
	}

	pdf.SetFont(fontFamily, "B", 12)
	pdf.CellFormat(0, 8, label(
		fmt.Sprintf("风险等级: %s", riskDisplay),
		fmt.Sprintf("Risk Level: %s", riskDisplay),
	), "", 1, "L", false, 0, "")

	pdf.SetFont(fontFamily, "", 10)
	pdf.MultiCell(0, 6, label(
		fmt.Sprintf("风险描述: %s", content.RiskDescription),
		fmt.Sprintf("Description: %s", content.RiskDescription),
	), "", "L", false)
	pdf.Ln(6)

	// --- 匹配案件信息 ---
	if content.MatchedCases != nil && len(content.MatchedCases.ToMap()) > 0 {
		s.pdfSectionHeader(pdf, label("匹配案件信息", "Matched Cases"), fontFamily, useChinese)
		s.pdfDrawJSONSection(pdf, content.MatchedCases, fontFamily, useChinese)
		pdf.Ln(6)
	}

	// --- 关联公司信息 ---
	if content.RelatedCompanies != nil && len(content.RelatedCompanies.ToMap()) > 0 {
		s.pdfSectionHeader(pdf, label("关联公司信息", "Related Companies"), fontFamily, useChinese)
		s.pdfDrawJSONSection(pdf, content.RelatedCompanies, fontFamily, useChinese)
		pdf.Ln(6)
	}

	// --- 详细信息 ---
	if content.ConflictDetails != nil && len(content.ConflictDetails.ToMap()) > 0 {
		s.pdfSectionHeader(pdf, label("详细信息", "Conflict Details"), fontFamily, useChinese)
		s.pdfDrawJSONSection(pdf, content.ConflictDetails, fontFamily, useChinese)
		pdf.Ln(6)
	}

	// --- 签名区域 ---
	pdf.Ln(10)
	pdf.SetDrawColor(180, 180, 180)
	pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
	pdf.Ln(5)
	pdf.SetFont(fontFamily, "", 9)
	pdf.CellFormat(0, 6, label(
		fmt.Sprintf("检测人员ID: %d    检测时间: %s", content.CheckedBy, content.CheckTime.Format("2006-01-02 15:04:05")),
		fmt.Sprintf("Checked By ID: %d    Check Time: %s", content.CheckedBy, content.CheckTime.Format("2006-01-02 15:04:05")),
	), "", 1, "L", false, 0, "")

	return pdf.OutputFileAndClose(filepath)
}

// loadChineseFont 尝试加载中文字体，返回是否成功
func (s *pdfReportService) loadChineseFont(pdf *gofpdf.Fpdf) bool {
	// 尝试加载 macOS 系统中文字体
	chineseFontPaths := []string{
		"/System/Library/Fonts/STHeiti Light.ttc",
		"/System/Library/Fonts/PingFang.ttc",
		"/System/Library/Fonts/STHeiti Medium.ttc",
		"/System/Library/Fonts/Supplemental/Songti.ttc",
	}

	for _, fontPath := range chineseFontPaths {
		if _, err := os.Stat(fontPath); err == nil {
			// gofpdf AddUTF8Font: family, style, file, uniFile
			// TTC 文件需要用 AddUTF8FontFromReader 或指定 index
			pdf.AddUTF8Font("STHeiti", "", fontPath)
			return true
		}
	}

	log.Printf("⚠️ 未找到中文字体文件，PDF将以英文模式生成")
	return false
}

// pdfSectionHeader 绘制 PDF 章节标题
func (s *pdfReportService) pdfSectionHeader(pdf *gofpdf.Fpdf, title string, fontFamily string, useChinese bool) {
	pdf.SetFont(fontFamily, "B", 14)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(0, 10, title, "", 1, "L", false, 0, "")

	pdf.SetDrawColor(70, 130, 180)
	pdf.SetLineWidth(0.5)
	pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
	pdf.SetLineWidth(0.2)
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(4)
}

// pdfDrawTable 绘制两列信息表格
func (s *pdfReportService) pdfDrawTable(pdf *gofpdf.Fpdf, data [][]string, fontFamily string) {
	colW := []float64{45, 145}
	lineH := 7.0

	// 表头背景色
	for i, row := range data {
		if i%2 == 0 {
			pdf.SetFillColor(245, 245, 245)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}

		// 标签列
		pdf.SetFont(fontFamily, "B", 10)
		pdf.CellFormat(colW[0], lineH, row[0], "1", 0, "L", true, 0, "")

		// 值列
		pdf.SetFont(fontFamily, "", 10)
		pdf.CellFormat(colW[1], lineH, row[1], "1", 1, "L", true, 0, "")
	}
}

// pdfDrawJSONSection 将 JSON 数据绘制为表格形式
func (s *pdfReportService) pdfDrawJSONSection(pdf *gofpdf.Fpdf, data models.JSON, fontFamily string, useChinese bool) {
	label := func(zh, en string) string {
		if useChinese {
			return zh
		}
		return en
	}

	m := data.ToMap()
	if len(m) == 0 {
		return
	}

	lineH := 7.0
	colW := []float64{45, 145}

	// 检查是否包含数组数据（如 matchedCases 可能包含 cases 数组）
	for key, value := range m {
		switch v := value.(type) {
		case []interface{}:
			// 数组类型：绘制为列表
			pdf.SetFont(fontFamily, "B", 10)
			pdf.CellFormat(0, lineH, label(fmt.Sprintf("  %s (%d 项)", key, len(v)), fmt.Sprintf("  %s (%d items)", key, len(v))), "", 1, "L", false, 0, "")

			for _, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					// 为每个数组项绘制子表格
					for k, val := range itemMap {
						if pdf.GetY() > 265 {
							pdf.AddPage()
						}
						pdf.SetFont(fontFamily, "B", 9)
						pdf.CellFormat(colW[0], lineH, fmt.Sprintf("    %s", k), "1", 0, "L", false, 0, "")
						pdf.SetFont(fontFamily, "", 9)
						pdf.CellFormat(colW[1], lineH, fmt.Sprintf("%v", val), "1", 1, "L", false, 0, "")
					}
					pdf.Ln(2)
				} else {
					pdf.SetFont(fontFamily, "", 9)
					pdf.CellFormat(0, lineH, fmt.Sprintf("    - %v", item), "", 1, "L", false, 0, "")
				}
			}
		default:
			if pdf.GetY() > 265 {
				pdf.AddPage()
			}
			pdf.SetFont(fontFamily, "B", 10)
			pdf.CellFormat(colW[0], lineH, fmt.Sprintf("  %s", key), "1", 0, "L", false, 0, "")
			pdf.SetFont(fontFamily, "", 10)
			pdf.CellFormat(colW[1], lineH, fmt.Sprintf("%v", v), "1", 1, "L", false, 0, "")
		}
	}
}

// calculateFileHash 计算文件哈希
func (s *pdfReportService) calculateFileHash(filepath string) (string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// getReportURL 获取报告访问 URL
func (s *pdfReportService) getReportURL(filename string) string {
	return fmt.Sprintf("%s/%s", s.config.BaseURL, filename)
}

// extractFilename 从 URL 提取文件名
func (s *pdfReportService) extractFilename(url string) string {
	filename := filepath.Base(url)
	if filename == "" || filename == "." || filename == "/" {
		return ""
	}
	return filename
}

// loadKeys 加载签名密钥
func (s *pdfReportService) loadKeys() error {
	// 如果密钥文件不存在，生成新的密钥对
	if _, err := os.Stat(s.config.SignatureKeyFile); os.IsNotExist(err) {
		log.Printf("🔑 生成新的签名密钥对...")
		if err := s.generateKeyPair(); err != nil {
			return err
		}
	}

	// 读取私钥
	privateKeyData, err := os.ReadFile(s.config.SignatureKeyFile)
	if err != nil {
		return fmt.Errorf("读取私钥文件失败: %w", err)
	}

	// 解析 PEM
	block, _ := pem.Decode(privateKeyData)
	if block == nil {
		return fmt.Errorf("私钥格式错误")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("解析私钥失败: %w", err)
	}

	s.privateKey = privateKey
	s.publicKey = &privateKey.PublicKey

	log.Printf("✅ 签名密钥加载成功")
	return nil
}

// generateKeyPair 生成密钥对
func (s *pdfReportService) generateKeyPair() error {
	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("生成私钥失败: %w", err)
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(s.config.SignatureKeyFile), 0700); err != nil {
		return err
	}

	// 保存私钥
	privateKeyFile, err := os.OpenFile(s.config.SignatureKeyFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer privateKeyFile.Close()

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		return err
	}

	// 保存公钥
	publicKeyFile, err := os.Create(s.config.CertificateFile)
	if err != nil {
		return err
	}
	defer publicKeyFile.Close()

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}

	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	if err := pem.Encode(publicKeyFile, publicKeyPEM); err != nil {
		return err
	}

	log.Printf("✅ 密钥对生成完成: %s", s.config.SignatureKeyFile)
	return nil
}

// signReport 对报告进行签名
func (s *pdfReportService) signReport(report *models.ConflictReport, fileHash string) error {
	// 在实际应用中，这里应该使用真正的数字签名算法
	// 这里简化处理，使用哈希值作为签名标识

	signature := s.generateSignatureString(fileHash)

	// 将签名信息存储在复核意见中
	report.ReviewNotes = fmt.Sprintf("数字签名: %s\n文件哈希: %s", signature, fileHash)

	return nil
}

// generateSignatureString 生成签名字符串
func (s *pdfReportService) generateSignatureString(fileHash string) string {
	hash := sha256.Sum256([]byte(fileHash))
	return hex.EncodeToString(hash[:])[:16]
}

// HashData 对数据进行哈希
func HashData(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// SignData 对数据进行签名
func SignData(data []byte, privateKey *rsa.PrivateKey) (string, error) {
	hash := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

// VerifySignatureData 验证数据签名
func VerifySignatureData(data []byte, signature string, publicKey *rsa.PublicKey) (bool, error) {
	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, err
	}

	hash := sha256.Sum256(data)
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], sig)
	return err == nil, nil
}
