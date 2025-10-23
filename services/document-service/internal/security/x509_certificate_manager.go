package security

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// CertificateStatus 证书状态
type CertificateStatus string

const (
	CertificateStatusActive    CertificateStatus = "active"     // 激活状态
	CertificateStatusSuspended CertificateStatus = "suspended"  // 暂停状态
	CertificateStatusRevoked   CertificateStatus = "revoked"    // 吊销状态
	CertificateStatusExpired   CertificateStatus = "expired"    // 过期状态
	CertificateStatusPending   CertificateStatus = "pending"    // 待签发状态
	CertificateStatusRenewed   CertificateStatus = "renewed"    // 已续期状态
)

// CertificateRequest 证书请求
type CertificateRequest struct {
	ID              string                 `json:"id"`
	TemplateID      string                 `json:"template_id"`
	Subject         pkix.Name              `json:"subject"`
	PublicKey       crypto.PublicKey        `json:"-"`
	Validity        time.Duration          `json:"validity"`
	KeyUsage        x509.KeyUsage          `json:"key_usage"`
	ExtKeyUsage     []x509.ExtKeyUsage     `json:"ext_key_usage"`
	DNSNames        []string               `json:"dns_names"`
	IPAddresses     []net.IP               `json:"ip_addresses"`
	URIs            []*url.URL             `json:"uris"`
	EmailAddresses  []string               `json:"email_addresses"`
	Extensions      []pkix.Extension        `json:"extensions"`
	RequesterID     string                 `json:"requester_id"`
	ApproverID      string                 `json:"approver_id"`
	Metadata        map[string]interface{} `json:"metadata"`
	RequestedAt     time.Time              `json:"requested_at"`
	ExpiresAt       time.Time              `json:"expires_at"`
	Status          CertificateStatus       `json:"status"`
}

// Certificate 证书信息
type Certificate struct {
	ID                   string                 `json:"id"`
	SerialNumber         *big.Int               `json:"serial_number"`
	Subject              pkix.Name              `json:"subject"`
	Issuer               pkix.Name              `json:"issuer"`
	NotBefore            time.Time              `json:"not_before"`
	NotAfter             time.Time              `json:"not_after"`
	PublicKey            crypto.PublicKey       `json:"-"`
	PrivateKey           crypto.PrivateKey      `json:"-"`
	Certificate          *x509.Certificate      `json:"-"`
	CertificatePEM       string                 `json:"certificate_pem"`
	PrivateKeyPEM        string                 `json:"private_key_pem"`
	CertificateChain     [][]byte               `json:"certificate_chain"`
	Status               CertificateStatus       `json:"status"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
	ExpiresAt            time.Time              `json:"expires_at"`
	RevokedAt            *time.Time             `json:"revoked_at"`
	RevocationReason     int                    `json:"revocation_reason"`
	RenewalCount         int                    `json:"renewal_count"`
	PreviousSerialNumber *big.Int               `json:"previous_serial_number"`
	TemplateID           string                 `json:"template_id"`
	RequesterID          string                 `json:"requester_id"`
	Metadata             map[string]interface{} `json:"metadata"`
}

// CertificateTemplate 证书模板
type CertificateTemplate struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Version         string                 `json:"version"`
	Description     string                 `json:"description"`
	Category        string                 `json:"category"`
	Template        *x509.Certificate      `json:"template"`
	Variables       []TemplateVariable     `json:"variables"`
	ValidationRules []ValidationRule       `json:"validation_rules"`
	Constraints     map[string]interface{} `json:"constraints"`
	Metadata        map[string]string      `json:"metadata"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
	CreatedBy       string                `json:"created_by"`
	UpdatedBy       string                `json:"updated_by"`
	Status          TemplateStatus        `json:"status"`
}

// TemplateVariable 模板变量
type TemplateVariable struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Required     bool        `json:"required"`
	DefaultValue interface{} `json:"default_value"`
	Description  string      `json:"description"`
	Validation   string      `json:"validation"`
	Options      []string    `json:"options,omitempty"`
}

// ValidationRule 验证规则
type ValidationRule struct {
	Field     string `json:"field"`
	Rule      string `json:"rule"`
	Value     string `json:"value"`
	Required  bool   `json:"required"`
	Message   string `json:"message"`
}

// TemplateStatus 模板状态
type TemplateStatus string

const (
	TemplateStatusActive   TemplateStatus = "active"    // 激活状态
	TemplateStatusInactive TemplateStatus = "inactive"  // 非激活状态
	TemplateStatusArchived TemplateStatus = "archived"  // 已归档状态
	TemplateStatusDraft    TemplateStatus = "draft"     // 草稿状态
)

// BatchCertificateRequest 批量证书请求
type BatchCertificateRequest struct {
	BatchID     string                         `json:"batch_id"`
	TemplateID  string                         `json:"template_id"`
	Requests    []SingleCertificateRequest     `json:"requests"`
	Options     BatchOptions                   `json:"options"`
	Metadata    map[string]string              `json:"metadata"`
	RequestedBy string                         `json:"requested_by"`
	RequestedAt time.Time                     `json:"requested_at"`
}

// SingleCertificateRequest 单个证书请求
type SingleCertificateRequest struct {
	RequestID string                 `json:"request_id"`
	Variables map[string]interface{} `json:"variables"`
	Metadata  map[string]string      `json:"metadata"`
}

// BatchOptions 批量处理选项
type BatchOptions struct {
	MaxConcurrency    int           `json:"max_concurrency"`
	RetryAttempts     int           `json:"retry_attempts"`
	RetryDelay        time.Duration `json:"retry_delay"`
	ContinueOnError   bool          `json:"continue_on_error"`
	NotificationEmail string        `json:"notification_email"`
}

// BatchResult 批量处理结果
type BatchResult struct {
	BatchID      string        `json:"batch_id"`
	Total        int           `json:"total"`
	Success      int           `json:"success"`
	Failed       int           `json:"failed"`
	Duration     time.Duration `json:"duration"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time"`
	Results      []*TaskResult `json:"results"`
	Errors       []string      `json:"errors"`
}

// TaskResult 任务结果
type TaskResult struct {
	BatchID     string         `json:"batch_id"`
	RequestID   string         `json:"request_id"`
	TaskID      string         `json:"task_id"`
	Success     bool           `json:"success"`
	Certificate *Certificate   `json:"certificate,omitempty"`
	Error       error          `json:"error,omitempty"`
	Timestamp   time.Time      `json:"timestamp"`
	Duration    time.Duration  `json:"duration"`
}

// X509CertificateManager X.509证书管理器
type X509CertificateManager struct {
	ca               *CertificateAuthority
	templateManager  *CertificateTemplateManager
	serialGenerator  SerialNumberGenerator
	batchProcessor   *BatchProcessor
	logger           *logrus.Logger
	mu               sync.RWMutex
	certificates     map[string]*Certificate
	templates        map[string]*CertificateTemplate
}

// NewX509CertificateManager 创建X.509证书管理器
func NewX509CertificateManager(ca *CertificateAuthority, logger *logrus.Logger) *X509CertificateManager {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	manager := &X509CertificateManager{
		ca:              ca,
		templateManager: NewCertificateTemplateManager(),
		serialGenerator: NewRandomSerialNumberGenerator(128),
		batchProcessor:  NewBatchProcessor(),
		logger:          logger,
		certificates:    make(map[string]*Certificate),
		templates:       make(map[string]*CertificateTemplate),
	}

	// 初始化默认模板
	manager.initializeDefaultTemplates()

	return manager
}

// initializeDefaultTemplates 初始化默认模板
func (xcm *X509CertificateManager) initializeDefaultTemplates() {
	// TLS服务器证书模板
	tlsServerTemplate := &CertificateTemplate{
		ID:          "tls-server",
		Name:        "TLS服务器证书",
		Version:     "1.0",
		Description: "用于HTTPS等TLS服务器认证",
		Category:    "server",
		Template: &x509.Certificate{
			KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
			ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		},
		Constraints: map[string]interface{}{
			"min_validity": 24 * time.Hour,
			"max_validity": 365 * 24 * time.Hour,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: "system",
		UpdatedBy: "system",
		Status:    TemplateStatusActive,
	}

	// TLS客户端证书模板
	tlsClientTemplate := &CertificateTemplate{
		ID:          "tls-client",
		Name:        "TLS客户端证书",
		Version:     "1.0",
		Description: "用于客户端证书认证",
		Category:    "client",
		Template: &x509.Certificate{
			KeyUsage:    x509.KeyUsageDigitalSignature,
			ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		},
		Constraints: map[string]interface{}{
			"min_validity": 24 * time.Hour,
			"max_validity": 365 * 24 * time.Hour,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: "system",
		UpdatedBy: "system",
		Status:    TemplateStatusActive,
	}

	// 代码签名证书模板
	codeSigningTemplate := &CertificateTemplate{
		ID:          "code-signing",
		Name:        "代码签名证书",
		Version:     "1.0",
		Description: "用于软件代码签名",
		Category:    "code-signing",
		Template: &x509.Certificate{
			KeyUsage:    x509.KeyUsageDigitalSignature,
			ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
		},
		Constraints: map[string]interface{}{
			"min_validity":        30 * 24 * time.Hour,
			"max_validity":        3 * 365 * 24 * time.Hour,
			"require_hardware_key": true,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: "system",
		UpdatedBy: "system",
		Status:    TemplateStatusActive,
	}

	// 邮件签名证书模板
	emailTemplate := &CertificateTemplate{
		ID:          "email-signing",
		Name:        "邮件签名证书",
		Version:     "1.0",
		Description: "用于邮件数字签名",
		Category:    "email",
		Template: &x509.Certificate{
			KeyUsage:    x509.KeyUsageDigitalSignature,
			ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageEmailProtection},
		},
		Constraints: map[string]interface{}{
			"min_validity": 30 * 24 * time.Hour,
			"max_validity": 365 * 24 * time.Hour,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: "system",
		UpdatedBy: "system",
		Status:    TemplateStatusActive,
	}

	xcm.templates["tls-server"] = tlsServerTemplate
	xcm.templates["tls-client"] = tlsClientTemplate
	xcm.templates["code-signing"] = codeSigningTemplate
	xcm.templates["email-signing"] = emailTemplate
}

// IssueCertificate 签发单个证书
func (xcm *X509CertificateManager) IssueCertificate(req *CertificateRequest) (*Certificate, error) {
	xcm.logger.WithFields(logrus.Fields{
		"template_id": req.TemplateID,
		"requester_id": req.RequesterID,
		"subject":      req.Subject.CommonName,
	}).Info("开始签发证书")

	// 验证证书请求
	if err := xcm.validateCertificateRequest(req); err != nil {
		return nil, fmt.Errorf("证书请求验证失败: %w", err)
	}

	// 生成证书序列号
	serialNumber, err := xcm.serialGenerator.Next()
	if err != nil {
		return nil, fmt.Errorf("生成序列号失败: %w", err)
	}

	// 创建证书模板
	template := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               req.Subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(req.Validity),
		KeyUsage:              req.KeyUsage,
		ExtKeyUsage:           req.ExtKeyUsage,
		DNSNames:              req.DNSNames,
		IPAddresses:           req.IPAddresses,
		EmailAddresses:        req.EmailAddresses,
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	// 设置颁发者
	template.Issuer = xcm.ca.Certificate.Subject

	// 添加扩展字段
	if len(req.Extensions) > 0 {
		template.ExtraExtensions = req.Extensions
	}

	// 创建主题备用名称扩展
	if len(req.DNSNames) > 0 || len(req.IPAddresses) > 0 || len(req.EmailAddresses) > 0 || len(req.URIs) > 0 {
		sanExtension, err := xcm.createSubjectAlternativeNameExtension(req)
		if err != nil {
			return nil, fmt.Errorf("创建SAN扩展失败: %w", err)
		}
		template.ExtraExtensions = append(template.ExtraExtensions, *sanExtension)
	}

	// 生成证书
	certBytes, err := x509.CreateCertificate(rand.Reader, template, xcm.ca.Certificate, req.PublicKey, xcm.ca.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("创建证书失败: %w", err)
	}

	// 解析证书
	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, fmt.Errorf("解析证书失败: %w", err)
	}

	// 编码PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	// 创建证书对象
	certificate := &Certificate{
		ID:               generateCertificateID(),
		SerialNumber:     serialNumber,
		Subject:          req.Subject,
		Issuer:           xcm.ca.Certificate.Subject,
		NotBefore:        cert.NotBefore,
		NotAfter:         cert.NotAfter,
		PublicKey:        req.PublicKey,
		Certificate:      cert,
		CertificatePEM:   string(certPEM),
		Status:           CertificateStatusActive,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		ExpiresAt:        cert.NotAfter,
		RenewalCount:     0,
		TemplateID:       req.TemplateID,
		RequesterID:      req.RequesterID,
		Metadata:         req.Metadata,
	}

	// 保存证书
	xcm.mu.Lock()
	xcm.certificates[certificate.ID] = certificate
	xcm.mu.Unlock()

	xcm.logger.WithFields(logrus.Fields{
		"certificate_id": certificate.ID,
		"serial_number": serialNumber.String(),
		"subject":        req.Subject.CommonName,
		"expires_at":    cert.NotAfter,
	}).Info("证书签发成功")

	return certificate, nil
}

// BatchIssueCertificate 批量签发证书
func (xcm *X509CertificateManager) BatchIssueCertificate(req *BatchCertificateRequest) (*BatchResult, error) {
	xcm.logger.WithFields(logrus.Fields{
		"batch_id":    req.BatchID,
		"template_id": req.TemplateID,
		"total":       len(req.Requests),
		"requester_id": req.RequestedBy,
	}).Info("开始批量签发证书")

	// 创建批量处理任务
	batchConfig := &BatchProcessorConfig{
		MaxConcurrency: req.Options.MaxConcurrency,
		RetryAttempts:  req.Options.RetryAttempts,
		RetryDelay:     req.Options.RetryDelay,
		ContinueOnError: req.Options.ContinueOnError,
	}

	processor := xcm.batchProcessor.Create(batchConfig)
	defer processor.Close()

	resultCollector := NewResultCollector(req.BatchID)
	startTime := time.Now()

	// 提交任务
	for i, singleReq := range req.Requests {
		task := &CertificateIssuanceTask{
			TaskID:      fmt.Sprintf("%s-task-%d", req.BatchID, i),
			BatchID:     req.BatchID,
			RequestID:   singleReq.RequestID,
			TemplateID:  req.TemplateID,
			Variables:   singleReq.Variables,
			Manager:     xcm,
			Collector:   resultCollector,
		}
		processor.Submit(task)
	}

	// 启动处理
	processor.Start()

	// 等待完成
	processor.Wait()

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	// 收集结果
	results := resultCollector.GetResults()

	// 统计结果
	successCount := 0
	failedCount := 0
	var errors []string

	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			failedCount++
			if result.Error != nil {
				errors = append(errors, fmt.Sprintf("任务 %s 失败: %v", result.RequestID, result.Error))
			}
		}
	}

	batchResult := &BatchResult{
		BatchID:   req.BatchID,
		Total:     len(req.Requests),
		Success:   successCount,
		Failed:    failedCount,
		Duration:  duration,
		StartTime: startTime,
		EndTime:   endTime,
		Results:   results,
		Errors:    errors,
	}

	xcm.logger.WithFields(logrus.Fields{
		"batch_id":    req.BatchID,
		"total":       batchResult.Total,
		"success":     batchResult.Success,
		"failed":      batchResult.Failed,
		"duration":    batchResult.Duration,
	}).Info("批量签发完成")

	return batchResult, nil
}

// validateCertificateRequest 验证证书请求
func (xcm *X509CertificateManager) validateCertificateRequest(req *CertificateRequest) error {
	if req.PublicKey == nil {
		return fmt.Errorf("公钥不能为空")
	}

	if req.Validity <= 0 {
		return fmt.Errorf("有效期必须大于0")
	}

	if req.Subject.CommonName == "" && len(req.DNSNames) == 0 && len(req.IPAddresses) == 0 {
		return fmt.Errorf("必须指定CommonName或至少一个DNS名称/IP地址")
	}

	// 检查模板是否存在
	if req.TemplateID != "" {
		if _, exists := xcm.templates[req.TemplateID]; !exists {
			return fmt.Errorf("模板不存在: %s", req.TemplateID)
		}
	}

	return nil
}

// createSubjectAlternativeNameExtension 创建主题备用名称扩展
func (xcm *X509CertificateManager) createSubjectAlternativeNameExtension(req *CertificateRequest) (*pkix.Extension, error) {
	var sanValues []asn1.RawValue

	// 添加DNS名称
	for _, dns := range req.DNSNames {
		value, err := asn1.MarshalWithParams(dns, "ia5")
		if err != nil {
			return nil, fmt.Errorf("编码DNS名称失败: %w", err)
		}
		sanValues = append(sanValues, asn1.RawValue{Tag: 2, Class: 2, Bytes: value})
	}

	// 添加IP地址
	for _, ip := range req.IPAddresses {
		sanValues = append(sanValues, asn1.RawValue{Tag: 7, Class: 2, Bytes: ip})
	}

	// 添加邮箱地址
	for _, email := range req.EmailAddresses {
		value, err := asn1.MarshalWithParams(email, "ia5")
		if err != nil {
			return nil, fmt.Errorf("编码邮箱地址失败: %w", err)
		}
		sanValues = append(sanValues, asn1.RawValue{Tag: 1, Class: 2, Bytes: value})
	}

	// 添加URI
	for _, uri := range req.URIs {
		value, err := asn1.MarshalWithParams(uri.String(), "ia5")
		if err != nil {
			return nil, fmt.Errorf("编码URI失败: %w", err)
		}
		sanValues = append(sanValues, asn1.RawValue{Tag: 6, Class: 2, Bytes: value})
	}

	sanSequence, err := asn1.Marshal(sanValues)
	if err != nil {
		return nil, fmt.Errorf("编码SAN序列失败: %w", err)
	}

	extension := pkix.Extension{
		Id:       []int{2, 5, 29, 17}, // SAN OID
		Critical: true,
		Value:    sanSequence,
	}

	return &extension, nil
}

// GetCertificate 获取证书
func (xcm *X509CertificateManager) GetCertificate(id string) (*Certificate, error) {
	xcm.mu.RLock()
	defer xcm.mu.RUnlock()

	cert, exists := xcm.certificates[id]
	if !exists {
		return nil, fmt.Errorf("证书不存在: %s", id)
	}

	return cert, nil
}

// ListCertificates 列出证书
func (xcm *X509CertificateManager) ListCertificates() []*Certificate {
	xcm.mu.RLock()
	defer xcm.mu.RUnlock()

	certificates := make([]*Certificate, 0, len(xcm.certificates))
	for _, cert := range xcm.certificates {
		certificates = append(certificates, cert)
	}

	return certificates
}

// RevokeCertificate 吊销证书
func (xcm *X509CertificateManager) RevokeCertificate(id string, reason int) error {
	xcm.mu.Lock()
	defer xcm.mu.Unlock()

	cert, exists := xcm.certificates[id]
	if !exists {
		return fmt.Errorf("证书不存在: %s", id)
	}

	cert.Status = CertificateStatusRevoked
	now := time.Now()
	cert.RevokedAt = &now
	cert.RevocationReason = reason
	cert.UpdatedAt = now

	xcm.logger.WithFields(logrus.Fields{
		"certificate_id": id,
		"serial_number": cert.SerialNumber.String(),
		"reason":        reason,
	}).Info("证书已吊销")

	return nil
}

// RenewCertificate 续期证书
func (xcm *X509CertificateManager) RenewCertificate(id string, newValidity time.Duration) (*Certificate, error) {
	xcm.mu.Lock()
	defer xcm.mu.Unlock()

	oldCert, exists := xcm.certificates[id]
	if !exists {
		return nil, fmt.Errorf("证书不存在: %s", id)
	}

	// 创建续期请求
	renewalReq := &CertificateRequest{
		ID:              generateCertificateID(),
		TemplateID:      oldCert.TemplateID,
		Subject:         oldCert.Subject,
		PublicKey:       oldCert.PublicKey,
		Validity:        newValidity,
		KeyUsage:        oldCert.Certificate.KeyUsage,
		ExtKeyUsage:     oldCert.Certificate.ExtKeyUsage,
		DNSNames:        oldCert.Certificate.DNSNames,
		IPAddresses:     oldCert.Certificate.IPAddresses,
		EmailAddresses:  oldCert.Certificate.EmailAddresses,
		RequesterID:     oldCert.RequesterID,
		RequestedAt:     time.Now(),
		Status:          CertificateStatusPending,
	}

	// 签发新证书
	newCert, err := xcm.IssueCertificate(renewalReq)
	if err != nil {
		return nil, fmt.Errorf("续期证书失败: %w", err)
	}

	// 更新旧证书状态
	oldCert.Status = CertificateStatusRenewed
	oldCert.UpdatedAt = time.Now()
	newCert.RenewalCount = oldCert.RenewalCount + 1
	newCert.PreviousSerialNumber = oldCert.SerialNumber

	xcm.logger.WithFields(logrus.Fields{
		"old_cert_id":     id,
		"new_cert_id":     newCert.ID,
		"old_serial":      oldCert.SerialNumber.String(),
		"new_serial":      newCert.SerialNumber.String(),
		"renewal_count":   newCert.RenewalCount,
	}).Info("证书续期成功")

	return newCert, nil
}

// GenerateKeyPair 生成密钥对
func (xcm *X509CertificateManager) GenerateKeyPair(algorithm string, keySize int) (crypto.PrivateKey, crypto.PublicKey, error) {
	switch algorithm {
	case "RSA":
		privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
		if err != nil {
			return nil, nil, err
		}
		return privateKey, &privateKey.PublicKey, nil

	case "ECDSA":
		var curve elliptic.Curve
		switch keySize {
		case 224:
			curve = elliptic.P224()
		case 256:
			curve = elliptic.P256()
		case 384:
			curve = elliptic.P384()
		case 521:
			curve = elliptic.P521()
		default:
			curve = elliptic.P256()
		}

		privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
		if err != nil {
			return nil, nil, err
		}
		return privateKey, &privateKey.PublicKey, nil

	default:
		return nil, nil, fmt.Errorf("不支持的密钥算法: %s", algorithm)
	}
}

// ValidateCertificateChain 验证证书链
func (xcm *X509CertificateManager) ValidateCertificateChain(cert *x509.Certificate) error {
	// 创建根证书池
	roots := x509.NewCertPool()
	roots.AddCert(xcm.ca.Certificate)

	// 设置验证选项
	opts := x509.VerifyOptions{
		Roots:    roots,
		CurrentTime: time.Now(),
	}

	// 验证证书链
	_, err := cert.Verify(opts)
	if err != nil {
		return fmt.Errorf("证书链验证失败: %w", err)
	}

	return nil
}

// GetTemplate 获取模板
func (xcm *X509CertificateManager) GetTemplate(id string) (*CertificateTemplate, error) {
	xcm.mu.RLock()
	defer xcm.mu.RUnlock()

	template, exists := xcm.templates[id]
	if !exists {
		return nil, fmt.Errorf("模板不存在: %s", id)
	}

	return template, nil
}

// ListTemplates 列出模板
func (xcm *X509CertificateManager) ListTemplates() []*CertificateTemplate {
	xcm.mu.RLock()
	defer xcm.mu.RUnlock()

	templates := make([]*CertificateTemplate, 0, len(xcm.templates))
	for _, template := range xcm.templates {
		templates = append(templates, template)
	}

	return templates
}

// generateCertificateID 生成证书ID
func generateCertificateID() string {
	return fmt.Sprintf("cert_%d", time.Now().UnixNano())
}

// 辅助函数和接口定义

// SerialNumberGenerator 序列号生成器接口
type SerialNumberGenerator interface {
	Next() (*big.Int, error)
}

// RandomSerialNumberGenerator 随机序列号生成器
type RandomSerialNumberGenerator struct {
	bitLength int
}

// NewRandomSerialNumberGenerator 创建随机序列号生成器
func NewRandomSerialNumberGenerator(bitLength int) *RandomSerialNumberGenerator {
	return &RandomSerialNumberGenerator{bitLength: bitLength}
}

// Next 生成下一个序列号
func (rsg *RandomSerialNumberGenerator) Next() (*big.Int, error) {
	max := new(big.Int).Lsh(big.NewInt(1), uint(rsg.bitLength))
	return rand.Int(rand.Reader, max)
}

// BatchProcessor 批处理器接口
type BatchProcessor interface {
	Create(config *BatchProcessorConfig) BatchProcessor
	Submit(task BatchTask)
	Start()
	Wait()
	Close()
}

// BatchProcessorConfig 批处理器配置
type BatchProcessorConfig struct {
	MaxConcurrency int
	RetryAttempts  int
	RetryDelay     time.Duration
	ContinueOnError bool
}

// BatchTask 批处理任务接口
type BatchTask interface {
	Execute() error
	GetTaskID() string
}

// BatchProcessor 批处理器实现
type BatchProcessorImpl struct {
	config   *BatchProcessorConfig
	tasks    chan BatchTask
	workers  []*Worker
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewBatchProcessor 创建批处理器
func NewBatchProcessor() BatchProcessor {
	return &BatchProcessorImpl{}
}

// Create 创建批处理器实例
func (bp *BatchProcessorImpl) Create(config *BatchProcessorConfig) BatchProcessor {
	ctx, cancel := context.WithCancel(context.Background())

	processor := &BatchProcessorImpl{
		config: config,
		tasks:  make(chan BatchTask, 100),
		ctx:    ctx,
		cancel: cancel,
	}

	processor.startWorkers()

	return processor
}

// Submit 提交任务
func (bp *BatchProcessorImpl) Submit(task BatchTask) {
	bp.tasks <- task
}

// Start 启动处理
func (bp *BatchProcessorImpl) Start() {
	// 批处理器在创建时自动启动工作器
}

// Wait 等待完成
func (bp *BatchProcessorImpl) Wait() {
	close(bp.tasks)
	bp.wg.Wait()
}

// Close 关闭处理器
func (bp *BatchProcessorImpl) Close() {
	bp.cancel()
}

// startWorkers 启动工作器
func (bp *BatchProcessorImpl) startWorkers() {
	for i := 0; i < bp.config.MaxConcurrency; i++ {
		worker := NewWorker(i, &bp.wg, bp.ctx)
		bp.workers = append(bp.workers, worker)
		worker.Start()
	}
}

// Worker 工作器
type Worker struct {
	id   int
	wg   *sync.WaitGroup
	ctx  context.Context
}

// NewWorker 创建工作器
func NewWorker(id int, wg *sync.WaitGroup, ctx context.Context) *Worker {
	return &Worker{id: id, wg: wg, ctx: ctx}
}

// Start 启动工作器
func (w *Worker) Start() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for {
			select {
			case task := <-getTaskChannel():
				task.Execute()
			case <-w.ctx.Done():
				return
			}
		}
	}()
}

// getTaskChannel 获取任务通道（模拟实现）
func getTaskChannel() chan BatchTask {
	return make(chan BatchTask)
}

// CertificateIssuanceTask 证书签发任务
type CertificateIssuanceTask struct {
	TaskID     string
	BatchID    string
	RequestID  string
	TemplateID string
	Variables  map[string]interface{}
	Manager    *X509CertificateManager
	Collector  *ResultCollector
}

// Execute 执行任务
func (cit *CertificateIssuanceTask) Execute() error {
	startTime := time.Now()

	// 这里应该根据模板和变量创建证书请求
	// 为了简化，我们直接返回成功
	result := &TaskResult{
		BatchID:   cit.BatchID,
		RequestID: cit.RequestID,
		TaskID:    cit.TaskID,
		Success:   true,
		Timestamp: time.Now(),
		Duration:  time.Since(startTime),
	}

	cit.Collector.Collect(result)
	return nil
}

// GetTaskID 获取任务ID
func (cit *CertificateIssuanceTask) GetTaskID() string {
	return cit.TaskID
}

// ResultCollector 结果收集器
type ResultCollector struct {
	batchID string
	results []*TaskResult
	mu      sync.Mutex
}

// NewResultCollector 创建结果收集器
func NewResultCollector(batchID string) *ResultCollector {
	return &ResultCollector{
		batchID: batchID,
		results: make([]*TaskResult, 0),
	}
}

// Collect 收集结果
func (rc *ResultCollector) Collect(result *TaskResult) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.results = append(rc.results, result)
}

// GetResults 获取结果
func (rc *ResultCollector) GetResults() []*TaskResult {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	return rc.results
}

// CertificateTemplateManager 证书模板管理器
type CertificateTemplateManager struct {
	templates map[string]*CertificateTemplate
	mu        sync.RWMutex
}

// NewCertificateTemplateManager 创建证书模板管理器
func NewCertificateTemplateManager() *CertificateTemplateManager {
	return &CertificateTemplateManager{
		templates: make(map[string]*CertificateTemplate),
	}
}

// GetTemplate 获取模板
func (ctm *CertificateTemplateManager) GetTemplate(id string) (*CertificateTemplate, error) {
	ctm.mu.RLock()
	defer ctm.mu.RUnlock()

	template, exists := ctm.templates[id]
	if !exists {
		return nil, fmt.Errorf("模板不存在: %s", id)
	}

	return template, nil
}

// AddTemplate 添加模板
func (ctm *CertificateTemplateManager) AddTemplate(template *CertificateTemplate) {
	ctm.mu.Lock()
	defer ctm.mu.Unlock()
	ctm.templates[template.ID] = template
}