package services

import (
	"bytes"
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
	"strings"
	"time"

	"law-oa-go/internal/models"
	"gorm.io/gorm"
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
	CheckedBy        uint                      `json:"checkedBy" validate:"required"`
	CheckTime        time.Time                 `json:"checkTime"`
	CheckDurationMs  *int                      `json:"checkDurationMs"`
	ClientName       string                    `json:"clientName" validate:"required"`
	ClientTaxID      string                    `json:"clientTaxId"`
	OpposingParty    string                    `json:"opposingParty"`
	RiskLevel        string                    `json:"riskLevel"`
	MatchedCases     models.JSON               `json:"matchedCases"`
	RelatedCompanies models.JSON               `json:"relatedCompanies"`
	ConflictDetails  models.JSON               `json:"conflictDetails"`
	TemplateType     string                    `json:"templateType"` // standard/detailed
}

// ReportFilter 报告过滤条件
type ReportFilter struct {
	CheckedBy      *uint      `json:"checkedBy"`
	RiskLevel      string     `json:"riskLevel"`
	StartDate      *time.Time `json:"startDate"`
	EndDate        *time.Time `json:"endDate"`
	ReportNumber   string     `json:"reportNumber"`
	ClientName     string     `json:"clientName"`
	Limit          int        `json:"limit"`
	Offset         int        `json:"offset"`
}

// SignatureVerificationResult 签名验证结果
type SignatureVerificationResult struct {
	Valid      bool      `json:"valid"`
	SignedBy   string    `json:"signedBy"`
	SignedAt   time.Time `json:"signedAt"`
	PublicKey  string    `json:"publicKey"`
	Hash       string    `json:"hash"`
	Signature  string    `json:"signature"`
}

// pdfReportService PDF 报告服务实现
type pdfReportService struct {
	db           *gorm.DB
	config       *PDFReportConfig
	privateKey   *rsa.PrivateKey
	publicKey    *rsa.PublicKey
}

// PDFReportConfig PDF 报告配置
type PDFReportConfig struct {
	OutputDir        string        `json:"outputDir"`         // PDF 输出目录
	BaseURL          string        `json:"baseUrl"`           // 报告访问基础 URL
	TemplateDir      string        `json:"templateDir"`       // 模板目录
	EnableSignature  bool          `json:"enableSignature"`   // 是否启用签名
	SignatureKeyFile string        `json:"signatureKeyFile"`  // 签名密钥文件
	CertificateFile  string        `json:"certificateFile"`   // 证书文件
	ReportPrefix     string        `json:"reportPrefix"`      // 报告编号前缀
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
		OutputDir:       "./reports/conflict",
		BaseURL:         "https://law-oa.example.com/reports",
		TemplateDir:     "./templates/pdf",
		EnableSignature: true,
		SignatureKeyFile: "./keys/private.pem",
		CertificateFile: "./keys/certificate.pem",
		ReportPrefix:    "CR",
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
		ReportNumber:       reportNumber,
		CheckedBy:          req.CheckedBy,
		CheckTime:          req.CheckTime,
		CheckDurationMs:    req.CheckDurationMs,
		ClientName:         req.ClientName,
		ClientTaxID:        req.ClientTaxID,
		OpposingParty:      req.OpposingParty,
		RiskLevel:          req.RiskLevel,
		MatchedCases:       req.MatchedCases,
		RelatedCompanies:   req.RelatedCompanies,
		ConflictDetails:    req.ConflictDetails,
		ReportURL:          s.getReportURL(filename),
		ReportGeneratedAt:  &now,
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
		ReportNumber:    s.generateReportNumber(),
		GeneratedAt:     time.Now(),
		CheckedBy:       req.CheckedBy,
		CheckTime:       req.CheckTime,
		ClientName:      req.ClientName,
		ClientTaxID:     req.ClientTaxID,
		OpposingParty:   req.OpposingParty,
		RiskLevel:       req.RiskLevel,
		MatchedCases:    req.MatchedCases,
		RelatedCompanies: req.RelatedCompanies,
		ConflictDetails: req.ConflictDetails,
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

// generatePDF 生成 PDF 文件（简化版，实际应使用 PDF 库）
func (s *pdfReportService) generatePDF(content *ReportContent, filepath string) error {
	// 在实际应用中，这里应该使用 PDF 生成库（如 github.com/jung-kurt/gofpdf）
	// 这里为了演示，生成一个简单的文本文件

	var buf bytes.Buffer

	buf.WriteString("====================================\n")
	buf.WriteString("       利益冲突检测报告\n")
	buf.WriteString("====================================\n\n")

	buf.WriteString(fmt.Sprintf("报告编号: %s\n", content.ReportNumber))
	buf.WriteString(fmt.Sprintf("生成时间: %s\n", content.GeneratedAt.Format("2006-01-02 15:04:05")))
	buf.WriteString(fmt.Sprintf("检测人员ID: %d\n", content.CheckedBy))
	buf.WriteString(fmt.Sprintf("检测时间: %s\n\n", content.CheckTime.Format("2006-01-02 15:04:05")))

	buf.WriteString("------------------------------------\n")
	buf.WriteString("       检测对象信息\n")
	buf.WriteString("------------------------------------\n")
	buf.WriteString(fmt.Sprintf("客户名称: %s\n", content.ClientName))
	if content.ClientTaxID != "" {
		buf.WriteString(fmt.Sprintf("客户税号: %s\n", content.ClientTaxID))
	}
	if content.OpposingParty != "" {
		buf.WriteString(fmt.Sprintf("对方当事人: %s\n", content.OpposingParty))
	}
	buf.WriteString("\n")

	buf.WriteString("------------------------------------\n")
	buf.WriteString("       风险评估结果\n")
	buf.WriteString("------------------------------------\n")
	buf.WriteString(fmt.Sprintf("风险等级: %s\n", content.RiskLevel))
	buf.WriteString(fmt.Sprintf("风险描述: %s\n\n", content.RiskDescription))

	buf.WriteString("------------------------------------\n")
	buf.WriteString("       匹配案件\n")
	buf.WriteString("------------------------------------\n")
	if content.MatchedCases != nil {
		buf.WriteString(fmt.Sprintf("%v\n\n", content.MatchedCases.ToMap()))
	}

	buf.WriteString("------------------------------------\n")
	buf.WriteString("       关联公司\n")
	buf.WriteString("------------------------------------\n")
	if content.RelatedCompanies != nil {
		buf.WriteString(fmt.Sprintf("%v\n\n", content.RelatedCompanies.ToMap()))
	}

	buf.WriteString("------------------------------------\n")
	buf.WriteString("       详细信息\n")
	buf.WriteString("------------------------------------\n")
	if content.ConflictDetails != nil {
		buf.WriteString(fmt.Sprintf("%v\n\n", content.ConflictDetails.ToMap()))
	}

	buf.WriteString("====================================\n")
	buf.WriteString("        报告结束\n")
	buf.WriteString("====================================\n")

	// 写入文件
	return os.WriteFile(filepath, buf.Bytes(), 0644)
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
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
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
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("生成私钥失败: %w", err)
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(s.config.SignatureKeyFile), 0700); err != nil {
		return err
	}

	// 保存私钥
	privateKeyFile, err := os.Create(s.config.SignatureKeyFile)
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
	timestamp := time.Now().Unix()
	data := fmt.Sprintf("%s|%d", fileHash, timestamp)
	hash := sha256.Sum256([]byte(data))
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
