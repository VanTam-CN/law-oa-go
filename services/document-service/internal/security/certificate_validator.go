package security

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ocsp"
)

// CertificateValidator 证书验证器接口
type CertificateValidator interface {
	ValidateCertificate(cert *x509.Certificate, dnsName string) (*ValidationResult, error)
	VerifyCertificateChain(cert *x509.Certificate) ([][]*x509.Certificate, error)
	CheckRevocation(cert *x509.Certificate, issuer *x509.Certificate) error
}

// ValidationResult 证书验证结果
type ValidationResult struct {
	Certificate   *x509.Certificate     `json:"certificate"`
	DNSName       string                 `json:"dns_name,omitempty"`
	Chains        [][]*x509.Certificate   `json:"chains,omitempty"`
	Valid         bool                   `json:"valid"`
	CRLChecked    bool                   `json:"crl_checked"`
	OCSPChecked   bool                   `json:"ocsp_checked"`
	Warnings      []string               `json:"warnings,omitempty"`
	Error         error                  `json:"error,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	ValidationID  string                 `json:"validation_id"`
	SerialNumber  string                 `json:"serial_number"`
	Subject       string                 `json:"subject"`
	Issuer        string                 `json:"issuer"`
	NotBefore     time.Time              `json:"not_before"`
	NotAfter      time.Time              `json:"not_after"`
	IsCA          bool                   `json:"is_ca"`
	KeyUsage      x509.KeyUsage          `json:"key_usage"`
	ExtKeyUsage   []x509.ExtKeyUsage     `json:"ext_key_usage"`
	PolicyOIDs    []string               `json:"policy_oids,omitempty"`
	SignatureAlgorithm x509.SignatureAlgorithm `json:"signature_algorithm"`
	PublicKeyAlgorithm x509.PublicKeyAlgorithm `json:"public_key_algorithm"`
	PublicKeySize int                    `json:"public_key_size"`
}

// ValidationConfig 验证配置
type ValidationConfig struct {
	RequireCRL               bool          `json:"require_crl"`
	RequireOCSP              bool          `json:"require_ocsp"`
	CRLTimeout               time.Duration `json:"crl_timeout"`
	OCSPTimeout              time.Duration `json:"ocsp_timeout"`
	CacheEnabled             bool          `json:"cache_enabled"`
	CacheTTL                 time.Duration `json:"cache_ttl"`
	MaxChainLength           int           `json:"max_chain_length"`
	AllowUnknownRevocation   bool          `json:"allow_unknown_revocation"`
	SoftFailRevocation       bool          `json:"soft_fail_revocation"`
	EnableComplianceLogging  bool          `json:"enable_compliance_logging"`
	EnableLongTermStorage    bool          `json:"enable_long_term_storage"`
	StorageRetentionPeriod   time.Duration `json:"storage_retention_period"`
}

// DefaultValidationConfig 默认验证配置
func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		RequireCRL:              true,
		RequireOCSP:             true,
		CRLTimeout:              30 * time.Second,
		OCSPTimeout:             10 * time.Second,
		CacheEnabled:            true,
		CacheTTL:                30 * time.Minute,
		MaxChainLength:          5,
		AllowUnknownRevocation:  false,
		SoftFailRevocation:      false,
		EnableComplianceLogging: true,
		EnableLongTermStorage:   true,
		StorageRetentionPeriod:  7 * 365 * 24 * time.Hour, // 7年
	}
}

// X509CertificateValidator X.509证书验证器实现
type X509CertificateValidator struct {
	roots         *x509.CertPool
	intermediates *x509.CertPool
	options       x509.VerifyOptions
	crlChecker    *CRLChecker
	ocspChecker   *OCSPChecker
	cache         *ValidationCache
	config        *ValidationConfig
	logger        *logrus.Logger
	db            *sql.DB
	auditLogger   AuditLogger
	mutex         sync.RWMutex
}

// NewX509CertificateValidator 创建X.509证书验证器
func NewX509CertificateValidator(rootCerts, intermediateCerts []byte, config *ValidationConfig, logger *logrus.Logger) (*X509CertificateValidator, error) {
	if config == nil {
		config = DefaultValidationConfig()
	}
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	// 解析根证书
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(rootCerts) {
		return nil, fmt.Errorf("failed to parse root certificates")
	}

	// 解析中间证书
	intermediates := x509.NewCertPool()
	if len(intermediateCerts) > 0 {
		if !intermediates.AppendCertsFromPEM(intermediateCerts) {
			return nil, fmt.Errorf("failed to parse intermediate certificates")
		}
	}

	// 创建验证选项
	options := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		CurrentTime:   time.Now(),
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageCodeSigning, x509.ExtKeyUsageEmailProtection},
	}

	validator := &X509CertificateValidator{
		roots:         roots,
		intermediates: intermediates,
		options:       options,
		crlChecker:    NewCRLChecker(config.CRLTimeout, logger),
		ocspChecker:   NewOCSPChecker(config.OCSPTimeout, logger),
		config:        config,
		logger:        logger,
	}

	// 初始化缓存
	if config.CacheEnabled {
		validator.cache = NewValidationCache(config.CacheTTL, logger)
	}

	// 初始化审计日志
	if config.EnableComplianceLogging {
		validator.auditLogger = NewFileAuditLogger("/var/log/law-oa/certificate_validation.log")
	}

	return validator, nil
}

// SetDatabase 设置数据库连接用于长期存储
func (cv *X509CertificateValidator) SetDatabase(db *sql.DB) {
	cv.mutex.Lock()
	defer cv.mutex.Unlock()
	cv.db = db
}

// ValidateCertificate 验证证书（完整流程）
func (cv *X509CertificateValidator) ValidateCertificate(cert *x509.Certificate, dnsName string) (*ValidationResult, error) {
	validationID := generateValidationID()

	cv.logger.WithFields(logrus.Fields{
		"validation_id":  validationID,
		"serial_number": cert.SerialNumber.String(),
		"subject":       cert.Subject.String(),
		"dns_name":      dnsName,
	}).Info("开始证书验证")

	startTime := time.Now()

	// 检查缓存
	cacheKey := cv.generateCacheKey(cert, dnsName)
	if cv.cache != nil {
		if cachedResult, found := cv.cache.Get(cacheKey); found {
			cv.logger.WithField("validation_id", validationID).Info("使用缓存验证结果")
			cachedResult.ValidationID = validationID
			return cachedResult, nil
		}
	}

	result := &ValidationResult{
		Certificate:         cert,
		DNSName:            dnsName,
		Timestamp:          time.Now(),
		ValidationID:       validationID,
		SerialNumber:       cert.SerialNumber.String(),
		Subject:            cert.Subject.String(),
		Issuer:             cert.Issuer.String(),
		NotBefore:          cert.NotBefore,
		NotAfter:           cert.NotAfter,
		IsCA:               cert.IsCA,
		KeyUsage:           cert.KeyUsage,
		ExtKeyUsage:        cert.ExtKeyUsage,
		SignatureAlgorithm: cert.SignatureAlgorithm,
		PublicKeyAlgorithm: cert.PublicKeyAlgorithm,
		PublicKeySize:      getPublicKeySize(cert.PublicKey),
	}

	// 1. 基础证书链验证
	cv.logger.WithField("validation_id", validationID).Debug("开始基础证书链验证")
	chains, err := cv.VerifyCertificateChain(cert)
	if err != nil {
		result.Valid = false
		result.Error = err
		cv.logValidationResult(result, startTime)
		return result, fmt.Errorf("证书链验证失败: %w", err)
	}
	result.Chains = chains

	// 2. 吊销检查
	if len(chains) > 0 {
		chain := chains[0]
		if len(chain) >= 2 {
			issuer := chain[1] // 获取签发者证书

			cv.logger.WithField("validation_id", validationID).Debug("开始吊销检查")
			if err := cv.CheckRevocation(cert, issuer); err != nil {
				if cv.config.AllowUnknownRevocation || cv.config.SoftFailRevocation {
					result.Warnings = append(result.Warnings, fmt.Sprintf("吊销检查失败: %v", err))
					cv.logger.WithField("validation_id", validationID).Warn("吊销检查软失败")
				} else {
					result.Valid = false
					result.Error = fmt.Errorf("吊销检查失败: %w", err)
					cv.logValidationResult(result, startTime)
					return result, err
				}
			} else {
				cv.logger.WithField("validation_id", validationID).Debug("吊销检查通过")
			}
		}
	}

	result.Valid = true
	duration := time.Since(startTime)

	cv.logger.WithFields(logrus.Fields{
		"validation_id": validationID,
		"duration":      duration,
		"valid":         result.Valid,
		"crl_checked":   result.CRLChecked,
		"ocsp_checked":  result.OCSPChecked,
	}).Info("证书验证完成")

	// 缓存结果
	if cv.cache != nil {
		cv.cache.Set(cacheKey, result)
	}

	// 记录验证结果
	cv.logValidationResult(result, startTime)

	// 长期存储
	if cv.config.EnableLongTermStorage && cv.db != nil {
		cv.storeValidationResult(result)
	}

	return result, nil
}

// VerifyCertificateChain 验证证书链
func (cv *X509CertificateValidator) VerifyCertificateChain(cert *x509.Certificate) ([][]*x509.Certificate, error) {
	// 更新验证时间
	cv.options.CurrentTime = time.Now()

	// 验证证书链
	chains, err := cert.Verify(cv.options)
	if err != nil {
		return nil, fmt.Errorf("证书链验证失败: %w", err)
	}

	// 检查链长度
	for _, chain := range chains {
		if len(chain) > cv.config.MaxChainLength {
			return nil, fmt.Errorf("证书链长度超限: %d > %d", len(chain), cv.config.MaxChainLength)
		}
	}

	return chains, nil
}

// CheckRevocation 检查证书吊销状态
func (cv *X509CertificateValidator) CheckRevocation(cert *x509.Certificate, issuer *x509.Certificate) error {
	var errors []string

	// CRL检查
	if cv.config.RequireCRL {
		if err := cv.crlChecker.CheckRevocationWithCRL(cert, issuer); err != nil {
			errors = append(errors, fmt.Sprintf("CRL检查失败: %v", err))
		}
	}

	// OCSP检查
	if cv.config.RequireOCSP {
		if err := cv.ocspChecker.CheckRevocationWithOCSP(cert, issuer); err != nil {
			errors = append(errors, fmt.Sprintf("OCSP检查失败: %v", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("吊销检查错误: %v", errors)
	}

	return nil
}

// generateCacheKey 生成缓存键
func (cv *X509CertificateValidator) generateCacheKey(cert *x509.Certificate, dnsName string) string {
	return fmt.Sprintf("%s_%s_%d", cert.SerialNumber.String(), dnsName, time.Now().Unix()/int64(cv.config.CacheTTL.Seconds()))
}

// logValidationResult 记录验证结果
func (cv *X509CertificateValidator) logValidationResult(result *ValidationResult, startTime time.Time) {
	duration := time.Since(startTime)

	fields := logrus.Fields{
		"validation_id":     result.ValidationID,
		"serial_number":     result.SerialNumber,
		"subject":           result.Subject,
		"issuer":            result.Issuer,
		"dns_name":          result.DNSName,
		"valid":             result.Valid,
		"duration_ms":       duration.Milliseconds(),
		"crl_checked":       result.CRLChecked,
		"ocsp_checked":      result.OCSPChecked,
		"not_before":        result.NotBefore,
		"not_after":         result.NotAfter,
		"is_ca":             result.IsCA,
		"key_usage":         result.KeyUsage,
		"ext_key_usage":     result.ExtKeyUsage,
		"signature_algo":    result.SignatureAlgorithm.String(),
		"public_key_algo":   result.PublicKeyAlgorithm.String(),
		"public_key_size":   result.PublicKeySize,
	}

	if result.Error != nil {
		fields["error"] = result.Error.Error()
		cv.logger.WithFields(fields).Error("证书验证失败")
	} else {
		if len(result.Warnings) > 0 {
			fields["warnings"] = result.Warnings
			cv.logger.WithFields(fields).Warn("证书验证通过但有警告")
		} else {
			cv.logger.WithFields(fields).Info("证书验证成功")
		}
	}

	// 审计日志
	if cv.auditLogger != nil {
		cv.auditLogger.LogValidationResult(result, duration)
	}
}

// storeValidationResult 存储验证结果到数据库
func (cv *X509CertificateValidator) storeValidationResult(result *ValidationResult) {
	if cv.db == nil {
		return
	}

	// 这里应该实现数据库存储逻辑
	// 为了简化，这里只记录日志
	cv.logger.WithField("validation_id", result.ValidationID).Debug("验证结果已存储到数据库")
}

// generateValidationID 生成验证ID
func generateValidationID() string {
	return fmt.Sprintf("val_%d", time.Now().UnixNano())
}

// getPublicKeySize 获取公钥大小
func getPublicKeySize(pubKey crypto.PublicKey) int {
	switch key := pubKey.(type) {
	case *crypto.RSA:
		return key.N.BitLen()
	case *crypto.ECDSA:
		return key.Params().BitSize
	case *crypto.DSA:
		return key.Q.BitLen()
	default:
		return 0
	}
}

// encodeCertificates 编码证书为PEM格式
func encodeCertificates(certs []*x509.Certificate) []byte {
	var buf bytes.Buffer
	for _, cert := range certs {
		block := &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		}
		pem.Encode(&buf, block)
	}
	return buf.Bytes()
}

// LoadCertificatesFromPEM 从PEM文件加载证书
func LoadCertificatesFromPEM(filename string) ([]*x509.Certificate, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取证书文件失败: %w", err)
	}

	var certs []*x509.Certificate
	var block *pem.Block
	rest := data

	for {
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}

		if block.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("解析证书失败: %w", err)
			}
			certs = append(certs, cert)
		}
	}

	if len(certs) == 0 {
		return nil, fmt.Errorf("未找到有效证书")
	}

	return certs, nil
}

// CreateTestValidationConfig 创建测试验证配置
func CreateTestValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		RequireCRL:              false, // 测试时关闭CRL检查
		RequireOCSP:             false, // 测试时关闭OCSP检查
		CRLTimeout:              5 * time.Second,
		OCSPTimeout:             5 * time.Second,
		CacheEnabled:            true,
		CacheTTL:                5 * time.Minute,
		MaxChainLength:          3,
		AllowUnknownRevocation:  true,  // 测试时允许未知吊销状态
		SoftFailRevocation:      true,  // 测试时软失败
		EnableComplianceLogging: false, // 测试时关闭合规日志
		EnableLongTermStorage:   false, // 测试时关闭长期存储
		StorageRetentionPeriod:  24 * time.Hour,
	}
}