package security

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
)

// DigitalSignatureService 数字签名服务接口
type DigitalSignatureService interface {
	// 基本签名操作
	SignDocument(ctx context.Context, request *SignRequest) (*SignResult, error)
	VerifySignature(ctx context.Context, request *VerifyRequest) (*VerifyResult, error)

	// 多签名操作
	CosignDocument(ctx context.Context, request *CosignRequest) (*CosignResult, error)
	VerifyCosignatures(ctx context.Context, request *VerifyCosignRequest) (*VerifyCosignResult, error)

	// 批量操作
	BatchSign(ctx context.Context, requests []*SignRequest) ([]*SignResult, error)
	BatchVerify(ctx context.Context, requests []*VerifyRequest) ([]*VerifyResult, error)

	// 证书管理
	GenerateCertificate(ctx context.Context, request *CertRequest) (*CertificateResult, error)
	IssueCertificate(ctx context.Context, request *IssueRequest) (*CertificateResult, error)
	RevokeCertificate(ctx context.Context, request *RevokeRequest) (*RevokeResult, error)

	// 验证链管理
	VerifyCertificateChain(ctx context.Context, cert *x509.Certificate) (*ChainVerifyResult, error)
	GetTrustStore(ctx context.Context) (*TrustStore, error)
	UpdateTrustStore(ctx context.Context, certs []*x509.Certificate) error
}

// SignRequest 签名请求
type SignRequest struct {
	DocumentID      string            `json:"document_id"`
	DocumentContent  []byte            `json:"document_content"`
	Format          string            `json:"format"`          // p7b, pkcs7, pdf, json
	Algorithm       string            `json:"algorithm"`       // sha256, sha384, sha512
	KeyID           string            `json:"key_id"`
	Password        string            `json:"password,omitempty"`
	Timestamp       bool              `json:"timestamp"`
	TimestampURL    string            `json:"timestamp_url,omitempty"`
	Reason          string            `json:"reason"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	UserID          string            `json:"user_id"`
	ClientIP        string            `json:"client_ip"`
	UserAgent        string            `json:"user_agent"`
}

// SignResult 签名结果
type SignResult struct {
	Success       bool                `json:"success"`
	Signature     *DigitalSignature   `json:"signature,omitempty"`
	SignedDocument *SignedDocument     `json:"signed_document,omitempty"`
	Error         string              `json:"error,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Timestamp     time.Time           `json:"timestamp"`
}

// VerifyRequest 验证请求
type VerifyRequest struct {
	DocumentID      string            `json:"document_id"`
	Signature     *DigitalSignature   `json:"signature"`
	DocumentContent  []byte            `json:"document_content,omitempty"`
	CheckTimestamp bool              `json:"check_timestamp"`
	TrustedOnly    bool              `json:"trusted_only"`
	UserID         string            `json:"user_id"`
	ClientIP       string            `json:"client_ip"`
}

// VerifyResult 验证结果
type VerifyResult struct {
	Success         bool                `json:"success"`
	Valid           bool                `json:"valid"`
	SignerInfo      *SignerInfo         `json:"signer_info,omitempty"`
	TimestampStatus TimestampStatus     `json:"timestamp_status"`
	ChainValid      bool                `json:"chain_valid"`
	Error           string              `json:"error,omitempty"`
	Warnings        []string            `json:"warnings,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	VerifiedAt      time.Time           `json:"verified_at"`
}

// DigitalSignature 数字签名
type DigitalSignature struct {
	ID              string            `json:"id"`
	Algorithm       string            `json:"algorithm"`
	SignatureValue  []byte            `json:"signature_value"`
	CertificateInfo *CertificateInfo  `json:"certificate_info"`
	Timestamp       *Timestamp        `json:"timestamp,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	ExpiresAt       time.Time         `json:"expires_at"`
}

// CertificateInfo 证书信息
type CertificateInfo struct {
	ID           string            `json:"id"`
	SerialNumber string            `json:"serial_number"`
	Subject      *CertificateSubject `json:"subject"`
	Issuer       *CertificateSubject `json:"issuer"`
	NotBefore    time.Time         `json:"not_before"`
	NotAfter     time.Time         `json:"not_after"`
	KeyUsage     []x509.KeyUsage   `json:"key_usage"`
	ExtKeyUsage  []x509.ExtKeyUsage `json:"ext_key_usage"`
	IsCA         bool              `json:"is_ca"`
	Fingerprint  string            `json:"fingerprint"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// CertificateSubject 证书主体信息
type CertificateSubject struct {
	CommonName         string            `json:"common_name"`
	Country            []string          `json:"country"`
	Organization       []string          `json:"organization"`
	OrganizationalUnit []string          `json:"organizational_unit"`
	Email              []string          `json:"email"`
}

// SignedDocument 已签名文档
type SignedDocument struct {
	ID           string            `json:"id"`
	DocumentID   string            `json:"document_id"`
	ContentType  string            `json:"content_type"`
	Content      []byte            `json:"content"`
	Signatures   []*DigitalSignature `json:"signatures"`
	Format       string            `json:"format"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// Timestamp 时间戳
type Timestamp struct {
	ID        string    `json:"id"`
	Hash      []byte    `json:"hash"`
	Time      time.Time `json:"time"`
	URL       string    `json:"url"`
	TSAInfo   string    `json:"tsa_info"`
	Signature []byte    `json:"signature"`
	CreatedAt time.Time `json:"created_at"`
}

// TimestampStatus 时间戳状态
type TimestampStatus string

const (
	TimestampStatusValid       TimestampStatus = "valid"
	TimestampStatusInvalid     TimestampStatus = "invalid"
	TimestampStatusExpired    TimestampStatus = "expired"
	TimestampStatusUnavailable TimestampStatus = "unavailable"
)

// SignerInfo 签名者信息
type SignerInfo struct {
	CertificateID string         `json:"certificate_id"`
	SubjectName    string         `json:"subject_name"`
	Email          string         `json:"email"`
	Organization   string         `json:"organization"`
	IsValid        bool           `json:"is_valid"`
	IsTrusted     bool           `json:"is_trusted"`
	VerifiedAt     time.Time      `json:"verified_at"`
}

// CertRequest 证书请求
type CertRequest struct {
	Subject         *CertificateSubject `json:"subject"`
	KeyAlgorithm    string            `json:"key_algorithm"`
	KeySize         int               `json:"key_size"`
	ValidityPeriod   int               `json:"validity_period"` // 天数
	IsCA            bool              `json:"is_ca"`
	DNSNames        []string          `json:"dns_names"`
	IPAddress      []string          `json:"ip_addresses"`
	KeyUsage        []x509.KeyUsage   `json:"key_usage"`
	ExtKeyUsage     []x509.ExtKeyUsage `json:"ext_key_usage"`
	UserID          string            `json:"user_id"`
	Reason          string            `json:"reason"`
}

// CertificateResult 证书结果
type CertificateResult struct {
	Success      bool                `json:"success"`
	Certificate  *CertificateInfo      `json:"certificate,omitempty"`
	PrivateKey   string              `json:"private_key,omitempty"`
	Error        string              `json:"error,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time           `json:"created_at"`
}

// IssueRequest 签发请求
type IssueRequest struct {
	CSR            string `json:"csr"`
	Certificate    *CertificateInfo `json:"certificate"`
	ValidityPeriod int    `json:"validity_period"`
	UserID          string `json:"user_id"`
	Reason          string `json:"reason"`
}

// RevokeRequest 吊销请求
type RevokeRequest struct {
	CertificateID string    `json:"certificate_id"`
	Reason       string    `json:"reason"`
	RevokedAt    time.Time `json:"revoked_at"`
	UserID       string    `json:"user_id"`
}

// RevokeResult 吊销结果
type RevokeResult struct {
	Success      bool      `json:"success"`
	CertificateID string    `json:"certificate_id"`
	Error        string    `json:"error,omitempty"`
	RevokedAt    time.Time `json:"revoked_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// CosignRequest 联合签名请求
type CosignRequest struct {
	DocumentID       string            `json:"document_id"`
	ExistingSignature *DigitalSignature `json:"existing_signature"`
	AdditionalSigners []string          `json:"additional_signers"`
	Reason           string            `json:"reason"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	UserID           string            `json:"user_id"`
}

// CosignResult 联合签名结果
type CosignResult struct {
	Success       bool                `json:"success"`
	Signatures    []*DigitalSignature `json:"signatures,omitempty"`
	Document      *SignedDocument     `json:"document,omitempty"`
	Error         string              `json:"error,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Timestamp     time.Time           `json:"timestamp"`
}

// VerifyCosignRequest 验证联合签名请求
type VerifyCosignRequest struct {
	DocumentID      string            `json:"document_id"`
	Document       *SignedDocument     `json:"document"`
	RequiredSigners []string          `json:"required_signers"`
	CheckTimestamp bool              `json:"check_timestamp"`
	TrustedOnly     bool              `json:"trusted_only"`
	UserID          string            `json:"user_id"`
}

// VerifyCosignResult 验证联合签名结果
type VerifyCosignResult struct {
	Success         bool                `json:"success"`
	ValidSignatures  []*SignerInfo       `json:"valid_signatures"`
	InvalidSignatures []*SignerInfo       `json:"invalid_signatures"`
	MissingSigners  []string            `json:"missing_signers"`
	TimestampStatus  TimestampStatus     `json:"timestamp_status"`
	ChainValid      bool                `json:"chain_valid"`
	Error           string              `json:"error,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	VerifiedAt      time.Time           `json:"verified_at"`
}

// ChainVerifyResult 证书链验证结果
type ChainVerifyResult struct {
	Success           bool                `json:"success"`
	ChainLength       int                `json:"chain_length"`
	RootTrusted       bool                `json:"root_trusted"`
	AllCertificatesValid bool                `json:"all_certificates_valid"`
	RevokedCertificates  []string            `json:"revoked_certificates"`
	ExpiredCertificates  []string            `json:"expired_certificates"`
	ValidationErrors   []string            `json:"validation_errors"`
	Warnings         []string            `json:"warnings"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	VerifiedAt         time.Time           `json:"verified_at"`
}

// TrustStore 信任存储
type TrustStore struct {
	RootCertificates []*x509.Certificate `json:"root_certificates"`
	IntermediateCerts   []*x509.Certificate `json:"intermediate_certs"`
	CRLs              []byte            `json:"crls"`
	OCSPServers       []string          `json:"ocsp_servers"`
	LastUpdated       time.Time         `json:"last_updated"`
}

// KeyPair 密钥对
type KeyPair struct {
	ID           string            `json:"id"`
	Algorithm    string            `json:"algorithm"`
	PrivateKey   []byte            `json:"private_key"`
	PublicKey    []byte            `json:"public_key"`
	Certificate  *CertificateInfo  `json:"certificate"`
	CreatedAt    time.Time         `json:"created_at"`
	ExpiresAt    time.Time         `json:"expires_at"`
	IsActive     bool              `json:"is_active"`
	UserID       string            `json:"user_id"`
}

// CertificateStore 证书存储接口
type CertificateStore interface {
	StoreCertificate(ctx context.Context, cert *CertificateInfo) error
	GetCertificate(ctx context.Context, certID string) (*CertificateInfo, error)
	ListCertificates(ctx context.Context, userID string) ([]*CertificateInfo, error)
	DeleteCertificate(ctx context.Context, certID string) error
	StoreKeyPair(ctx context.Context, keyPair *KeyPair) error
	GetKeyPair(ctx context.Context, keyID string) (*KeyPair, error)
	ListKeyPairs(ctx context.Context, userID string) ([]*KeyPair, error)
	RevokeCertificate(ctx context.Context, certID string, reason string) error
	GetRevokedCertificates(ctx context.Context) ([]*CertificateInfo, error)
}

// TimestampService 时间戳服务接口
type TimestampService interface {
	GenerateTimestamp(ctx context.Context, data []byte, timestampURL string) (*Timestamp, error)
	VerifyTimestamp(ctx context.Context, timestamp *Timestamp, data []byte) (*TimestampVerifyResult, error)
}

// TimestampVerifyResult 时间戳验证结果
type TimestampVerifyResult struct {
	Valid     bool      `json:"valid"`
	VerifiedAt time.Time `json:"verified_at"`
	Error     string    `json:"error,omitempty"`
	TSAInfo   string    `json:"tsa_info,omitempty"`
}

// AuditLogger 审计日志接口
type AuditLogger interface {
	LogSignature(ctx context.Context, event *AuditEvent) error
	LogCertificate(ctx context.Context, event *AuditEvent) error
	LogTimestamp(ctx context.Context, event *AuditEvent) error
	GetAuditLog(ctx context.Context, filters *AuditFilters) ([]*AuditEvent, error)
}

// AuditEvent 审计事件
type AuditEvent struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	ResourceID string    `json:"resource_id"`
	IPAddress string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	Details   map[string]interface{} `json:"details"`
	Success   bool      `json:"success"`
}

// AuditFilters 审计过滤器
type AuditFilters struct {
	UserID     string     `json:"user_id,omitempty"`
	Action     string     `json:"action,omitempty"`
	Resource   string     `json:"resource,omitempty"`
	StartTime time.Time  `json:"start_time,omitempty"`
	EndTime   time.Time  `json:"end_time,omitempty"`
	IPAddress string     `json:"ip_address,omitempty"`
	Limit      int        `json:"limit,omitempty"`
	Offset     int        `json:"offset,omitempty"`
}

// DigitalSignatureManager 数字签名管理器实现
type DigitalSignatureManager struct {
	signatureStore    CertificateStore
	timestampService TimestampService
	auditLogger      AuditLogger
	trustStore       *TrustStore
	keyStore         CertificateStore
	logger           *logrus.Logger
	config           *SignatureConfig
}

// SignatureConfig 签名配置
type SignatureConfig struct {
	DefaultAlgorithm    string            `json:"default_algorithm"`
	TimestampEnabled    bool              `json:"timestamp_enabled"`
	DefaultTSAURL      string            `json:"default_tsa_url"`
	MaxSignatureSize    int               `json:"max_signature_size"`
	CacheEnabled        bool              `json:"cache_enabled"`
	CacheTTL           int               `json:"cache_ttl"`
	ValidationEnabled    bool              `json:"validation_enabled"`
	ComplianceLevel     string            `json:"compliance_level"`
}

// DefaultSignatureConfig 默认签名配置
func DefaultSignatureConfig() *SignatureConfig {
	return &SignatureConfig{
		DefaultAlgorithm:  "sha256",
		TimestampEnabled:   true,
		DefaultTSAURL:     "http://timestamp.digicert.com",
		MaxSignatureSize:  1024 * 1024, // 1MB
		CacheEnabled:       true,
		CacheTTL:           3600,     // 1小时
		ValidationEnabled:   true,
		ComplianceLevel:     "legal",
	}
}

// NewDigitalSignatureManager 创建数字签名管理器
func NewDigitalSignatureManager(
	certStore CertificateStore,
	keyStore CertificateStore,
	timestampService TimestampService,
	auditLogger AuditLogger,
	trustStore *TrustStore,
	logger *logrus.Logger,
	config *SignatureConfig,
) DigitalSignatureService {
	if config == nil {
		config = DefaultSignatureConfig()
	}

	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	return &DigitalSignatureManager{
		signatureStore:    certStore,
		timestampService: timestampService,
		auditLogger:      auditLogger,
		trustStore:       trustStore,
		keyStore:         keyStore,
		logger:           logger,
		config:           config,
	}
}

// SignDocument 签名文档
func (dsm *DigitalSignatureManager) SignDocument(ctx context.Context, request *SignRequest) (*SignResult, error) {
	startTime := time.Now()

	defer func() {
		dsm.logger.WithFields(logrus.Fields{
			"document_id": request.DocumentID,
			"algorithm":  request.Algorithm,
			"duration":  time.Since(startTime),
			"user_id":     request.UserID,
		}).Info("文档签名完成")
	}()

	// 验证请求
	if err := dsm.validateSignRequest(ctx, request); err != nil {
		return &SignResult{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// 获取签名密钥
	keyPair, err := dsm.getSigningKey(ctx, request.KeyID, request.Password)
	if err != nil {
		return &SignResult{
			Success: false,
			Error:   fmt.Sprintf("获取签名密钥失败: %v", err),
		}, err
	}

	// 计算文档哈希
	hash, err := dsm.calculateHash(request.DocumentContent, request.Algorithm)
	if err != nil {
		return &SignResult{
			Success: false,
			Error:   fmt.Sprintf("计算文档哈希失败: %v", err),
		}, err
	}

	// 生成签名
	signatureValue, err := dsm.signHash(hash, keyPair.PrivateKey, keyPair.Algorithm)
	if err != nil {
		return &SignResult{
			Success: false,
			Error:   fmt.Sprintf("生成签名失败: %v", err),
		}, err
	}

	// 获取证书信息
	certInfo, err := dsm.getCertificateInfo(ctx, keyPair.ID)
	if err != nil {
		return &SignResult{
			Success: false,
			Error:   fmt.Sprintf("获取证书信息失败: %v", err),
		}, err
	}

	// 生成时间戳
	var timestamp *Timestamp
	if request.Timestamp {
		timestamp, err = dsm.generateTimestamp(ctx, hash, request.TimestampURL)
		if err != nil {
			dsm.logger.WithError(err).Warn("生成时间戳失败，继续签名流程")
		}
	}

	// 创建数字签名对象
	signature := &DigitalSignature{
		ID:              dsm.generateID(),
		Algorithm:       request.Algorithm,
		SignatureValue:  signatureValue,
		CertificateInfo: certInfo,
		Timestamp:       timestamp,
		Metadata:        request.Metadata,
		CreatedAt:       time.Now(),
		ExpiresAt:       time.Now().AddDate(0, 0, 30), // 30天
	}

	// 创建已签名文档
	signedDoc := &SignedDocument{
		ID:          dsm.generateID(),
		DocumentID:  request.DocumentID,
	ContentType: dsm.detectContentType(request.DocumentContent),
	Content:     request.DocumentContent,
	Signatures: []*DigitalSignature{signature},
	Format:      request.Format,
		Metadata:    request.Metadata,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 记录审计日志
	err = dsm.auditLogger.LogSignature(ctx, &AuditEvent{
		ID:         signature.ID,
		Timestamp:  signature.CreatedAt,
		UserID:     request.UserID,
		Action:     "document_signature",
		Resource:   "document",
		ResourceID: request.DocumentID,
		IPAddress: request.ClientIP,
		UserAgent:   request.UserAgent,
		Details: map[string]interface{}{
			"algorithm":      request.Algorithm,
			"certificate_id": certInfo.ID,
			"document_size":   len(request.DocumentContent),
			"format":         request.Format,
			"has_timestamp": timestamp != nil,
		},
		Success: true,
	})
	if err != nil {
		dsm.logger.WithError(err).Error("记录签名审计日志失败")
	}

	return &SignResult{
		Success:       true,
		Signature:     signature,
		SignedDocument: signedDoc,
		Metadata: map[string]interface{}{
			"algorithm":      request.Algorithm,
			"certificate_id": certInfo.ID,
			"has_timestamp": timestamp != nil,
			"signing_time":   time.Since(startTime),
		},
		Timestamp: time.Now(),
	}, nil
}

// VerifySignature 验证签名
func (dsm *DigitalSignatureManager) VerifySignature(ctx context.Context, request *VerifyRequest) (*VerifyResult, error) {
	startTime := time.Now()

	defer func() {
		dsm.logger.WithFields(logrus.Fields{
			"document_id": request.DocumentID,
			"check_timestamp": request.CheckTimestamp,
			"duration": time.Since(startTime),
			"user_id":     request.UserID,
		}).Info("签名验证完成")
	}()

	// 验证请求
	if err := dsm.validateVerifyRequest(ctx, request); err != nil {
		return &VerifyResult{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// 获取签名信息
	signature := request.Signature
	if signature == nil {
		return &VerifyResult{
			Success: false,
			Error:   "签名信息为空",
		}, fmt.Errorf("签名信息为空")
	}

	// 验证证书链
	certValid, certInfo, chainResult := dsm.verifyCertificateChain(ctx, signature.CertificateInfo)
	if !certValid {
		return &VerifyResult{
			Success:    false,
			Valid:      false,
			ChainValid: false,
			Error:      "证书验证失败",
			Metadata: map[string]interface{}{
				"chain_errors": chainResult.ValidationErrors,
			},
			VerifiedAt: time.Now(),
		}, nil
	}

	// 验证签名
	signatureValid, err := dsm.verifySignatureValue(
		signature.SignatureValue,
		signature.Algorithm,
		certInfo.PublicKey,
		request.DocumentContent,
	)
	if err != nil {
		return &VerifyResult{
			Success:    false,
			Valid:      false,
			Error:      fmt.Sprintf("签名验证失败: %v", err),
			ChainValid:  chainResult.AllCertificatesValid,
			SignerInfo: dsm.createSignerInfo(certInfo),
			VerifiedAt: time.Now(),
		}, err
	}

	// 验证时间戳
	timestampStatus := TimestampStatusUnavailable
	if request.CheckTimestamp && signature.Timestamp != nil {
		timestampStatus, _ = dsm.verifyTimestamp(ctx, signature.Timestamp, request.DocumentContent)
	}

	// 记录审计日志
	err = dsm.auditLogger.LogSignature(ctx, &AuditEvent{
		ID:         signature.ID,
		Timestamp:  time.Now(),
		UserID:     request.UserID,
		Action:     "signature_verification",
		Resource:   "document",
		ResourceID: request.DocumentID,
		IPAddress: request.ClientIP,
		UserAgent:   request.UserAgent,
		Details: map[string]interface{}{
			"certificate_id":    certInfo.ID,
			"signature_valid":    signatureValid,
			"chain_valid":       chainResult.AllCertificatesValid,
			"timestamp_status":  timestampStatus,
			"trusted_only":      request.TrustedOnly,
		},
		Success: signatureValid && chainResult.AllCertificatesValid && timestampStatus == TimestampStatusValid,
	})
	if err != nil {
		dsm.logger.WithError(err).Error("记录验证审计日志失败")
	}

	// 创建签名者信息
	signerInfo := dsm.createSignerInfo(certInfo)
	signerInfo.IsValid = signatureValid
	signerInfo.IsTrusted = dsm.isCertificateTrusted(certInfo)
	signerInfo.VerifiedAt = time.Now()

	return &VerifyResult{
		Success:         signatureValid && chainResult.AllCertificatesValid,
		Valid:           signatureValid,
		SignerInfo:       signerInfo,
		TimestampStatus:  timestampStatus,
		ChainValid:       chainResult.AllCertificatesValid,
		Warnings:        chainResult.Warnings,
		Metadata: map[string]interface{}{
			"certificate_id": certInfo.ID,
			"algorithm":      signature.Algorithm,
			"verification_time": time.Since(startTime),
		},
		VerifiedAt: time.Now(),
	}, nil
}

// CosignDocument 联合签名
func (dsm *DigitalSignatureManager) CosignDocument(ctx context.Context, request *CosignRequest) (*CosignResult, error) {
	startTime := time.Now()

	defer func() {
		dsm.logger.WithFields(logrus.Fields{
			"document_id": request.DocumentID,
			"additional_signers": len(request.AdditionalSigners),
			"duration": time.Since(startTime),
		"user_id": request.UserID,
	}).Info("联合签名完成")
	}()

	// 获取现有签名文档
	existingSignature := request.ExistingSignature
	if existingSignature == nil {
		return &CosignResult{
			Success: false,
			Error:   "现有签名为空",
		}, fmt.Errorf("现有签名为空")
	}

	// 验证现有签名
	verifyReq := &VerifyRequest{
		DocumentID: request.DocumentID,
		Signature: existingSignature,
		UserID:    request.UserID,
	}

	// 如果提供了文档内容，验证现有签名
	if len(request.DocumentID) > 0 {
		verifyReq.DocumentContent = dsm.getDocumentContent(ctx, request.DocumentID)
	}

	verifyResult, err := dsm.VerifySignature(ctx, verifyReq)
	if err != nil || !verifyResult.Success {
		return &CosignResult{
			Success: false,
			Error:   fmt.Sprintf("现有签名验证失败: %v", err),
		}, err
	}

	// 创建新的签名副本
	newSignatures := make([]*DigitalSignature, 0)
	newSignatures = append(newSignatures, existingSignature)

	// 为每个额外签名者添加签名
	for _, signerID := range request.AdditionalSigners {
		// 创建签名请求
		signReq := &SignRequest{
			DocumentID:     request.DocumentID,
			DocumentContent: dsm.getDocumentContent(ctx, request.DocumentID),
			Algorithm:      dsm.config.DefaultAlgorithm,
			KeyID:          signerID,
			Reason:         request.Reason,
			Timestamp:       dsm.config.TimestampEnabled,
			UserID:          request.UserID,
		}

		// 签名
		signResult, err := dsm.SignDocument(ctx, signReq)
		if err != nil {
			dsm.logger.WithError(err).WithField("signer_id", signerID).Error("联合签名失败")
			continue
		}

		if signResult.Success {
			newSignatures = append(newSignatures, signResult.Signature)
		}
	}

	// 更新已签名文档
	signedDoc := &SignedDocument{
		ID:          dsm.generateID(),
		DocumentID:  request.DocumentID,
		ContentType: dsm.detectContentType(dsm.getDocumentContent(ctx, request.DocumentID)),
		Content:     dsm.getDocumentContent(ctx, request.DocumentID),
		Signatures: newSignatures,
		Format:      "p7b", // 联合签名使用PKCS#7格式
		Metadata:    request.Metadata,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 记录审计日志
	err = dsm.auditLogger.LogSignature(ctx, &AuditEvent{
		ID:         dsm.generateID(),
		Timestamp:  time.Now(),
		UserID:     request.UserID,
		Action:     "cosign_document",
		Resource:   "document",
		ResourceID: request.DocumentID,
		IPAddress: "",
		UserAgent:   "",
		Details: map[string]interface{}{
			"existing_signatures": len(existingSignature.Signature),
			"additional_signers":  len(request.AdditionalSigners),
			"final_signatures":    len(newSignatures),
			"reason":             request.Reason,
		},
		Success: len(newSignatures) > 1,
	})
	if err != nil {
		dsm.logger.WithError(err).Error("记录联合签名审计日志失败")
	}

	return &CosignResult{
		Success:    len(newSignatures) > 1,
		Signatures: newSignatures,
		Document:  signedDoc,
		Metadata: map[string]interface{}{
			"cosign_time": time.Since(startTime),
			"signer_count": len(newSignatures),
		},
		Timestamp: time.Now(),
	}, nil
}

// VerifyCosignatures 验证联合签名
func (dsm *DigitalSignatureManager) VerifyCosignatures(ctx context.Context, request *VerifyCosignRequest) (*VerifyCosignResult, error) {
	startTime := time.Now()

	defer func() {
		dsm.logger.WithFields(logrus.Fields{
			"document_id": request.DocumentID,
			"required_signers": len(request.RequiredSigners),
			"duration": time.Since(startTime),
			"user_id":     request.UserID,
		}).Info("联合签名验证完成")
	}()

	// 验证文档格式
	signedDoc := request.Document
	if signedDoc == nil {
		return &VerifyCosignResult{
			Success: false,
			Error:   "已签名文档为空",
		}, fmt.Errorf("已签名文档为空")
	}

	// 验证所有签名
	validSignatures := make([]*SignerInfo, 0)
	invalidSignatures := make([]*SignerInfo, 0)
	missingSigners := make([]string, 0)

	// 检查所需的签名者
	requiredSigners := make(map[string]bool)
	for _, signer := range request.RequiredSigners {
		requiredSigners[signer] = true
	}

	// 验证每个签名
	for _, signature := range signedDoc.Signatures {
		verifyReq := &VerifyRequest{
			DocumentID:      request.DocumentID,
			Signature:      signature,
			DocumentContent: signedDoc.Content,
			CheckTimestamp:  request.CheckTimestamp,
			TrustedOnly:     request.TrustedOnly,
			UserID:          request.UserID,
		}

		result, err := dsm.VerifySignature(ctx, verifyReq)
		if err != nil || !result.Success {
			invalidSignatures = append(invalidSignatures, result.SignerInfo)
			continue
		}

		// 检查是否是所需的签名者
		signerID := dsm.getSignerID(signature.CertificateInfo)
		if _, required := requiredSigners[signerID]; required {
			validSignatures = append(validSignatures, result.SignerInfo)
			delete(requiredSigners, signerID)
		} else {
			invalidSignatures = append(invalidSignatures, result.SignerInfo)
		}
	}

	// 记录缺失的签名者
	for signerID := range requiredSigners {
		missingSigners = append(missingSigners, signerID)
	}

	// 验证时间戳
	timestampStatus := TimestampStatusValid
	if request.CheckTimestamp {
		// 检查所有签名的时间戳状态
		for _, signature := range signedDoc.Signatures {
			if signature.Timestamp != nil {
				status, _ := dsm.verifyTimestamp(ctx, signature.Timestamp, signedDoc.Content)
				if status != TimestampStatusValid {
					timestampStatus = status
					break
				}
			}
		}
	}

	// 验证证书链
	chainValid := true
	for _, signature := range signedDoc.Signatures {
		if signature.CertificateInfo != nil {
			_, _, chainResult := dsm.verifyCertificateChain(ctx, signature.CertificateInfo)
			if !chainResult.AllCertificatesValid {
				chainValid = false
				break
			}
		}
	}

	// 记录审计日志
	err := dsm.auditLogger.LogSignature(ctx, &AuditEvent{
		ID:         dsm.generateID(),
		Timestamp:  time.Now(),
		UserID:     request.UserID,
		Action:     "cosign_verification",
		Resource:   "document",
		ResourceID: request.DocumentID,
		IPAddress: "",
		UserAgent:   "",
		Details: map[string]interface{}{
			"total_signatures": len(signedDoc.Signatures),
			"valid_signatures": len(validSignatures),
			"invalid_signatures": len(invalidSignatures),
			"missing_signers": len(missingSigners),
			"timestamp_status": timestampStatus,
			"chain_valid": chainValid,
		},
		Success: len(validSignatures) > 0 && chainValid && len(missingSigners) == 0,
	})
	if err != nil {
		dsm.logger.WithError(err).Error("记录联合签名验证审计日志失败")
	}

	return &VerifyCosignResult{
		Success:         len(validSignatures) > 0 && chainValid && len(missingSigners) == 0,
		ValidSignatures:  validSignatures,
		InvalidSignatures: invalidSignatures,
		MissingSigners:  missingSigners,
		TimestampStatus:  timestampStatus,
		ChainValid:      chainValid,
		Error:           "",
		Metadata: map[string]interface{}{
			"verification_time": time.Since(startTime),
			"signer_count":    len(signedDoc.Signatures),
		},
		VerifiedAt:        time.Now(),
	}, nil
}

// BatchSign 批量签名
func (dsm *DigitalSignatureManager) BatchSign(ctx context.Context, requests []*SignRequest) ([]*SignResult, error) {
	results := make([]*SignResult, len(requests))

	for i, req := range requests {
		result, err := dsm.SignDocument(ctx, req)
		if err != nil {
			results[i] = &SignResult{
				Success: false,
				Error:   err.Error(),
			}
		} else {
			results[i] = result
		}
	}

	return results, nil
}

// BatchVerify 批量验证
func (dsm *DigitalSignatureManager) BatchVerify(ctx context.Context, requests []*VerifyRequest) ([]*VerifyResult, error) {
	results := make([]*VerifyResult, len(requests))

	for i, req := range requests {
		result, err := dsm.VerifySignature(ctx, req)
		if err != nil {
			results[i] = &VerifyResult{
				Success: false,
				Error:   err.Error(),
			}
		} else {
			results[i] = result
		}
	}

	return results, nil
}

// GenerateCertificate 生成证书
func (dsm *DigitalSignatureManager) GenerateCertificate(ctx context.Context, request *CertRequest) (*CertificateResult, error) {
	startTime := time.Now()

	defer func() {
		dsm.logger.WithFields(logrus.Fields{
			"user_id": request.UserID,
			"algorithm": request.KeyAlgorithm,
			"duration": time.Since(startTime),
		}).Info("证书生成完成")
	}()

	// 生成私钥
	var privateKey crypto.PrivateKey
	var err error

	switch request.KeyAlgorithm {
	case "RSA":
		privateKey, err = rsa.GenerateKey(rand.Reader, request.KeySize)
	case "ECDSA":
		privateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case "Ed25519":
		privateKey, err = ed25519.GenerateKey(rand.Reader)
	default:
		privateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	}

	if err != nil {
		return &CertificateResult{
			Success: false,
			Error:   fmt.Sprintf("生成私钥失败: %v", err),
		}, err
	}

	// 生成序列号
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return &CertificateResult{
			Success: false,
			Error:   fmt.Sprintf("生成序列号失败: %v", err),
		}, err
	}

	// 计算有效期
	notBefore := time.Now()
	notAfter := notBefore.AddDate(0, 0, request.ValidityPeriod)

	// 创建证书模板
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:         request.Subject.CommonName,
			Country:           request.Subject.Country,
			Organization:       request.Subject.Organization,
			OrganizationalUnit: request.Subject.OrganizationalUnit,
			Email:             request.Subject.Email,
		},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		KeyUsage:    request.KeyUsage,
		ExtKeyUsage: request.ExtKeyUsage,
		BasicConstraintsValid: true,
		IsCA:        request.IsCA,
		DNSNames:    request.DNSNames,
		IPAddresses: request.IPAddresses,
	}

	// 签发自签名证书
	if request.IsCA {
		template.IsCA = true
		template.MaxPathLen = 2
	}

	// 生成证书
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return &CertificateResult{
			Success: false,
			Error:   fmt.Sprintf("生成证书失败: %v", err),
		}, err
	}

	// 解析证书
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return &CertificateResult{
			Success: false,
			Error:   fmt.Sprintf("解析证书失败: %v", err),
		}, err
	}

	// 创建证书信息
	certInfo := dsm.createCertificateInfo(cert)

	// 编码私钥
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return &CertificateResult{
			Success: false,
			Error:   fmt.Sprintf("编码私钥失败: %v", err),
		}, err
	}

	privateKeyPEM := pem.EncodeToMemory(privateKeyBytes, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// 编码证书
	certPEM := string(certPEM)

	// 创建密钥对
	keyPair := &KeyPair{
		ID:           dsm.generateID(),
		Algorithm:    request.KeyAlgorithm,
		PrivateKey:   privateKeyPEM,
		PublicKey:    certPEM,
		Certificate:  certInfo,
		CreatedAt:    time.Now(),
		ExpiresAt:    notAfter,
		IsActive:     true,
		UserID:       request.UserID,
	}

	// 存储密钥对
	err = dsm.keyStore.StoreKeyPair(ctx, keyPair)
	if err != nil {
		return &CertificateResult{
			Success: false,
			Error:   fmt.Sprintf("存储密钥对失败: %v", err),
		}, err
	}

	// 存储证书
	err = dsm.signatureStore.StoreCertificate(ctx, certInfo)
	if err != nil {
		return &CertificateResult{
			Success: false,
			Error:   fmt.Sprintf("存储证书失败: %v", err),
		}, err
	}

	// 记录审计日志
	err = dsm.auditLogger.LogCertificate(ctx, &AuditEvent{
		ID:         dsm.generateID(),
		Timestamp:  time.Now(),
		UserID:     request.UserID,
		Action:     "certificate_generation",
		Resource:   "certificate",
		ResourceID: certInfo.ID,
		IPAddress: "",
		UserAgent:   "",
		Details: map[string]interface{}{
			"key_algorithm":  request.KeyAlgorithm,
			"key_size":      request.KeySize,
			"validity_period": request.ValidityPeriod,
			"is_ca":          request.IsCA,
			"serial_number":  certInfo.SerialNumber,
			"fingerprint":   certInfo.Fingerprint,
		},
		Success: true,
	})
	if err != nil {
		dsm.logger.WithError(err).Error("记录证书生成审计日志失败")
	}

	return &CertificateResult{
		Success:     true,
		Certificate: certInfo,
		PrivateKey:  privateKeyPEM,
		Metadata: map[string]interface{}{
			"generation_time": time.Since(startTime),
			"key_size":       request.KeySize,
			"validity_days":   request.ValidityPeriod,
		},
		CreatedAt:    time.Now(),
	}, nil
}

// IssueCertificate 签发证书
func (dsm *DigitalSignatureManager) IssueCertificate(ctx context.Context, request *IssueRequest) (*CertificateResult, error) {
	startTime := time.Now()

	defer func() {
		dsm.logger.WithFields(logrus.Fields{
			"user_id": request.UserID,
			"validity_period": request.ValidityPeriod,
			"duration": time.Since(startTime),
		}).Info("证书签发完成")
	}()

	// 解析CSR
	csr, err := x509.ParseCertificateRequest([]byte(request.CSR))
	if err != nil {
		return &CertificateResult{
			Success: false,
			Error:   fmt.Sprintf("解析CSR失败: %v", err),
		}, err
	}

	// 获取CA证书
	caKeyPair, err := dsm.getCertificateKeyPair(ctx, "ca")
	if err != nil {
		return &CertificateResult{
			Success: false,
			Error:   fmt.Sprintf("获取CA密钥对失败: %v", err),
		}, err
	}

	caCert, err := x509.ParseCertificate([]byte(caKeyPair.Certificate))
	if err != nil {
		return &CertificateRequest{
			Success: false,
			Error:   fmt.Sprintf("解析CA证书失败: %v", err),
		}, err
	}

	// 设置证书有效期
	notAfter := time.Now().AddDate(0, 0, request.ValidityPeriod)

	// 创建证书模板
	template := csr.Subject
	template.NotAfter = notAfter
	template.Subject.CommonName = csr.Subject.CommonName
	template.Subject.Country = csr.Subject.Country
	template.Subject.Organization = csr.Subject.Organization
	template.Subject.OrganizationalUnit = csr.Subject.OrganizationalUnit
	template.Subject.Email = csr.Subject.Email

	// 使用CA签名生成证书
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, csr.PublicKey, caKeyPair.PrivateKey)
	if err != nil {
		return &CertificateResult{
			Success: false,
			Error:   fmt.Sprintf("签发证书失败: %v", err),
		}, err
	}

	// 解析证书
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return &CertificateResult{
			Success: false,
			Error:   fmt.Sprintf("解析签发证书失败: %v", err),
		}, err
	}

	// 创建证书信息
	certInfo := dsm.createCertificateInfo(cert)

	// 存储证书
	err = dsm.signatureStore.StoreCertificate(ctx, certInfo)
	if err != nil {
		return &CertificateResult{
			Success: false,
			Error:   fmt.Sprintf("存储签发证书失败: %v", err),
		}, err
	}

	// 记录审计日志
	err = dsm.auditLogger.LogCertificate(ctx, &AuditEvent{
		ID:         dsm.generateID(),
		Timestamp:  time.Now(),
		UserID:     request.UserID,
		Action:     "certificate_issuance",
		Resource:   "certificate",
		ResourceID: certInfo.ID,
		IPAddress: "",
		UserAgent:   "",
		Details: map[string]interface{}{
			"ca_subject":     caCert.Subject.CommonName,
			"serial_number": certInfo.SerialNumber,
			"validity_period": request.ValidityPeriod,
		},
		Success: true,
	})
	if err != nil {
		dsm.logger.WithError(err).Error("记录证书签发审计日志失败")
	}

	return &CertificateResult{
		Success:     true,
		Certificate: certInfo,
		Metadata: map[string]interface{}{
			"issuance_time": time.Since(startTime),
			"validity_period":  request.ValidityPeriod,
		},
		CreatedAt:    time.Now(),
	}, nil
}

// RevokeCertificate 吊销证书
func (dsm *DigitalSignatureManager) RevokeCertificate(ctx context.Context, request *RevokeRequest) (*RevokeResult, error) {
	startTime := time.Now()

	defer func() {
		dsm.logger.WithFields(logrus.Fields{
			"certificate_id": request.CertificateID,
			"reason":      request.Reason,
			"user_id":     request.UserID,
			"duration":  time.Since(startTime),
	}).Info("证书吊销完成")
	}()

	// 获取证书信息
	certInfo, err := dsm.signatureStore.GetCertificate(ctx, request.CertificateID)
	if err != nil {
		return &RevokeResult{
			Success:      false,
			CertificateID: request.CertificateID,
			Error:        fmt.Sprintf("获取证书失败: %v", err),
		}, err
	}

	// 更新证书状态为已吊销
	certInfo.IsActive = false
	err = dsm.signatureStore.StoreCertificate(ctx, certInfo)
	if err != nil {
		return &RevokeResult{
			Success:      false,
			CertificateID: request.CertificateID,
			Error:        fmt.Sprintf("更新证书状态失败: %v", err),
		}, err
	}

	// 创建CRL条目
	crlEntry := pkix.CertificateList{
		SerialNumber: certInfo.SerialNumber,
	RevocationTime: request.RevokedAt,
	}

	// 记录审计日志
	err = dsm.auditLogger.LogCertificate(ctx, &AuditEvent{
		ID:         dsm.generateID(),
		Timestamp:  time.Now(),
		UserID:     request.UserID,
		Action:     "certificate_revocation",
		Resource:   "certificate",
		ResourceID: request.CertificateID,
		IPAddress: "",
		UserAgent:   "",
		Details: map[string]interface{}{
			"reason":      request.Reason,
			"serial_number": certInfo.SerialNumber,
		},
		Success: true,
	})
	if err != nil {
		dsm.logger.WithError(err).Error("记录证书吊销审计日志失败")
	}

	return &RevokeResult{
		Success:      true,
		CertificateID: request.CertificateID,
		RevokedAt:    request.RevokedAt,
		Metadata: map[string]interface{}{
			"revocation_time": time.Since(startTime),
		},
	}, nil
}

// VerifyCertificateChain 验证证书链
func (dsm *DigitalSignatureManager) VerifyCertificateChain(ctx context.Context, cert *x509.Certificate) (bool, *CertificateInfo, *ChainVerifyResult) {
	if cert == nil {
		return false, nil, &ChainVerifyResult{
			Success: false,
			Error:   "证书为空",
		}
	}

	certInfo := dsm.createCertificateInfo(cert)

	// 获取信任存储
	trustStore := dsm.getTrustStore(ctx)
	if trustStore == nil {
		return false, certInfo, &ChainVerifyResult{
			Success:     false,
			Error:   "信任存储未初始化",
		}
	}

	// 验证证书链
	opts := x509.VerifyOptions{
		Roots: trustStore.GetRootPool(),
		KeyUsages: []x509.KeyUsage{x509.KeyUsageDigitalSignature, x509.KeyUsageCertSign},
	}

	// 验证证书
	_, err := cert.Verify(opts)
	if err != nil {
		dsm.logger.WithError(err).Warn("证书验证失败")

		// 检查具体的错误类型
		switch err.(type) {
		case x509.CertificateInvalidError:
			return false, certInfo, &ChainVerifyResult{
				Success:     false,
				Error:       fmt.Sprintf("证书无效: %v", err),
				ValidationErrors: []string{err.Error()},
			}
		case x509.UnknownAuthorityError:
			return false, certInfo, &ChainVerifyResult{
				Success:     false,
				Error:       fmt.Sprintf("未知证书颁发机构: %v", err),
				ValidationErrors: []string{err.Error()},
			}
		default:
			return false, certInfo, &ChainVerifyResult{
				Success:     false,
				Error:       fmt.Sprintf("证书链验证失败: %v", err),
				ValidationErrors: []string{err.Error()},
			}
		}
	}

	// 检查证书是否在信任根中
	isTrusted := dsm.isCertificateTrusted(certInfo)

	// 构建验证结果
	warnings := []string{}
	if cert.NotAfter.Before(time.Now().Add(-24 * time.Hour)) {
		warnings = append(warnings, "证书即将过期")
	}

	chainResult = &ChainVerifyResult{
		Success:              true,
		ChainLength:           1,
		RootTrusted:           isTrusted,
		AllCertificatesValid:    true,
		RevokedCertificates:   []string{},
		ExpiredCertificates:   []string{},
		ValidationErrors:       []string{},
		Warnings:             warnings,
		Metadata: map[string]interface{}{
			"verified_at": time.Now(),
		},
	}

	return true, certInfo, chainResult
}

// GetTrustStore 获取信任存储
func (dsm *DigitalSignatureManager) GetTrustStore(ctx context.Context) (*TrustStore, error) {
	if dsm.trustStore == nil {
		// 创建默认信任存储
		dsm.trustStore = &TrustStore{
			RootCertificates: []*x509.Certificate{},
			LastUpdated:     time.Now(),
		}
	}

	return dsm.trustStore, nil
}

// UpdateTrustStore 更新信任存储
func (dsm *DigitalSignatureManager) UpdateTrustStore(ctx context.Context, certs []*x509.Certificate) error {
	if dsm.trustStore == nil {
		return fmt.Errorf("信任存储未初始化")
	}

	// 验证所有根证书
	validRoots := make([]*x509.Certificate, 0)
	for _, cert := range certs {
		if cert.IsCA && cert.CheckSignatureFrom(cert.PublicKey) == nil {
			validRoots = append(validRoots, cert)
		}
	}

	dsm.trustStore.RootCertificates = validRoots
	dsm.trustStore.LastUpdated = time.Now()

	dsm.logger.WithFields(logrus.Fields{
		"root_certs_count": len(validRoots),
		"updated_at": dsm.trustStore.LastUpdated,
	}).Info("信任存储更新完成")

	return nil
}

// 辅助方法

// validateSignRequest 验证签名请求
func (dsm *DigitalSignatureManager) validateSignRequest(ctx context.Context, request *SignRequest) error {
	if request.DocumentID == "" {
		return fmt.Errorf("文档ID不能为空")
	}

	if len(request.DocumentContent) == 0 {
		return fmt.Errorf("文档内容不能为空")
	}

	if request.Algorithm == "" {
		request.Algorithm = dsm.config.DefaultAlgorithm
	}

	if request.KeyID == "" {
		return fmt.Errorf("密钥ID不能为空")
	}

	if request.Format == "" {
		request.Format = "p7b" // 默认使用PKCS#7格式
	}

	// 验证文档大小
	if len(request.DocumentContent) > dsm.config.MaxSignatureSize {
		return fmt.Errorf("文档大小超过最大限制: %d bytes", dsm.config.MaxSignatureSize)
	}

	return nil
}

// validateVerifyRequest 验证验证请求
func (dsm *DigitalSignatureManager) validateVerifyRequest(ctx context.Context, request *VerifyRequest) error {
	if request.DocumentID == "" {
		return fmt.Errorf("文档ID不能为空")
	}

	if request.Signature == nil {
		return fmt.Errorf("签名信息不能为空")
	}

	// 如果提供了文档内容，验证签名与文档的一致性
	if len(request.DocumentContent) > 0 {
		calculatedHash, err := dsm.calculateHash(request.DocumentContent, request.Signature.Algorithm)
		if err != nil {
			return fmt.Errorf("计算文档哈希失败: %v", err)
		}

		storedHash, err := dsm.calculateHash(request.Signature.SignatureValue, request.Signature.Algorithm)
		if err != nil {
			return fmt.Errorf("获取签名哈希失败: %v", err)
		}

		if !bytes.Equal(calculatedHash, storedHash) {
			return fmt.Errorf("文档内容与签名不匹配")
		}
	}

	return nil
}

// getSigningKey 获取签名密钥
func (dsm *DigitalSignatureManager) getSigningKey(ctx context.Context, keyID, password string) (*KeyPair, error) {
	if password != "" {
		// 使用密码解密私钥
		keyPair, err := dsm.keyStore.GetKeyPair(ctx, keyID)
		if err != nil {
			return nil, err
		}

		// 这里应该实现密码解密逻辑
		// 简化处理：假设已经解密
		return keyPair, nil
	}

	// 直接获取密钥对
	return dsm.keyStore.GetKeyPair(ctx, keyID)
}

// getCertificateInfo 从证书创建证书信息
func (dsm *DigitalSignatureManager) createCertificateInfo(cert *x509.Certificate) *CertificateInfo {
	subject := cert.Subject
	issuer := cert.Issuer

	return &CertificateInfo{
		ID:           dsm.generateID(),
		SerialNumber: cert.SerialNumber.String(),
		Subject: &CertificateSubject{
			CommonName:         subject.CommonName,
			Country:           subject.Country,
			Organization:       subject.Organization,
			OrganizationalUnit: subject.OrganizationalUnit,
			Email:             subject.Email,
		},
		Issuer: &CertificateSubject{
			CommonName:         issuer.CommonName,
			Country:           issuer.Country,
			Organization:       issuer.Organization,
			OrganizationalUnit: issuer.OrganizationalUnit,
			Email:             issuer.Email,
		},
		NotBefore:    cert.NotBefore,
		NotAfter:     cert.NotAfter,
		KeyUsage:     cert.KeyUsage,
		ExtKeyUsage:  cert.ExtKeyUsage,
		IsCA:         cert.IsCA,
		Fingerprint:  fmt.Sprintf("%X", cert.Fingerprint),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// getSignerID 获取签名者ID
func (dsm *DigitalSignatureManager) getSignerID(certInfo *CertificateInfo) string {
	// 使用证书指纹作为签名者ID
	return certInfo.Fingerprint
}

// isCertificateTrusted 检查证书是否受信任
func (dsm *DigitalSignatureManager) isCertificateInfo(trusted *CertificateInfo) bool {
	if dsm.trustStore == nil {
		return false
	}

	// 检查是否在信任存储中
	for _, rootCert := range dsm.trustStore.RootCertificates {
		if rootCert.Subject.CommonName == trusted.Issuer.CommonName &&
			bytes.Equal(rootCert.Subject.RawSubject, trusted.Issuer.RawSubject) {
			return true
		}
	}

	// 检查证书链是否通过验证
	if dsm.trustStore.LastUpdated.IsZero() {
		return false
	}

	// 可以添加更多的信任检查逻辑
	return false
}

// calculateDocumentHash 计算文档哈希
func (dsm *DigitalSignatureManager) calculateDocumentHash(content []byte, algorithm string) ([]byte, error) {
	switch algorithm {
	case "sha256":
		hash := sha256.Sum256(content)
		return hash[:], nil
	case "sha384":
		hash := sha256.Sum384(content)
		return hash[:], nil
	case "sha512":
		hash := sha256.Sum512(content)
		return hash[:], nil
	default:
		return nil, fmt.Errorf("不支持的哈希算法: %s", algorithm)
	}
}

// calculateHash 计算签名哈希
func (dsm *DigitalSignatureManager) calculateHash(signatureValue []byte, algorithm string) ([]byte, error) {
	switch algorithm {
	case "sha256":
		hash := sha256.Sum256(signatureValue)
		return hash[:], nil
	case "sha384":
		hash := hash.Sum384(signatureValue)
		return hash[:], nil
	case "sha512":
		hash := hash.Sum512(signatureValue)
		return hash[:], nil
	default:
		return nil, fmt.Errorf("不支持的哈希算法: %s", algorithm)
	}
}

// signHash 对哈希进行签名
func (dsm *DigitalSignatureManager) signHash(hash []byte, privateKey crypto.PrivateKey, algorithm string) ([]byte, error) {
	var err error
	var signature []byte

	switch algorithm {
	case "sha256":
		var h sha256.Hash
		h.Write(hash)
		signature, err = rsa.SignPKCS1v15(rand.Reader, privateKey, h.Sum(), nil)
	case "sha384":
		var h sha384.Hash
		h.Write(hash)
		signature, err = rsa.SignPKCS1v15(rand.Reader, privateKey, h.Sum(), nil)
	case "sha512":
		var h sha512.Hash
		h.Write(hash)
		signature, err = rsa.SignPKCS1v15(rand.Reader, privateKey, h.Sum(), nil)
	case "ecdsa":
		signature, err = ecdsa.SignASN1(rand.Reader, privateKey, hash)
	case "ed25519":
		signature, err = ed25519.Sign(rand.Reader, privateKey, hash)
	default:
		return nil, fmt.Errorf("不支持的签名算法: %s", algorithm)
	}

	return signature, err
}

// detectContentType 检测文档内容类型
func (dsm *DigitalSignatureManager) detectContentType(content []byte) string {
	contentType := "application/octet-stream"

	// 简单的内容类型检测
	if len(content) > 4 {
		// PDF文件
		if bytes.HasPrefix(content, []byte("%PDF-")) {
			contentType = "application/pdf"
		}
		// XML文件
		if bytes.HasPrefix(content, []byte("<?xml")) {
			contentType = "application/xml"
		}
		// JSON文件
		if bytes.HasPrefix(content, []byte("{")) || bytes.HasSuffix(content, []byte("}")) {
			contentType = "application/json"
		}
		// PEM文件
		if bytes.HasPrefix(content, []byte("-----BEGIN")) {
			contentType = "application/x-pem"
		}
	}

	return contentType
}

// getDocumentContent 获取文档内容
func (dsm *DigitalSignatureManager) getDocumentContent(ctx context.Context, documentID string) []byte {
	// 这里应该从文档存储服务获取文档内容
	// 简化处理，返回空内容
	return []byte{}
}

// generateID 生成唯一ID
func (dsm *DigitalSignatureManager) generateID() string {
	return fmt.Sprintf("sig_%d", time.Now().UnixNano())
}

// createSignerInfo 创建签名者信息
func (dsm *DigitalSignatureManager) createSignerInfo(certInfo *CertificateInfo) *SignerInfo {
	return &SignerInfo{
		CertificateID: certInfo.ID,
		SubjectName:    dsm.getSubjectName(certInfo.Subject),
	Email:          dsm.getEmail(certInfo.Subject),
		Organization:   dsm.getOrganization(certInfo.Subject),
		IsValid:        true,
		IsTrusted:      dsm.isCertificateTrusted(certInfo),
		VerifiedAt:      time.Now(),
	}
}

// getSubjectName 获取主体名称
func (dsm *DigitalSignatureManager) getSubjectName(subject pkix.Name) string {
	if len(subject.CommonName) > 0 {
		return subject.CommonName
	}
	if len(subject.Email) > 0 {
		return subject.Email[0]
	}
	return ""
}

// getEmail 获取邮箱
func (dsm *DigitalSignatureManager) getEmail(subject pkix.Name) string {
	if len(subject.Email) > 0 {
		return subject.Email[0]
	}
	return ""
}

// getOrganization 获取组织
func (dsm *DigitalSignatureManager) getOrganization(subject pkix.Name) string {
	if len(subject.Organization) > 0 {
		return subject.Organization[0]
	}
	return ""
}

// generateTimestamp 生成时间戳
func (dsm *DigitalSignatureManager) generateTimestamp(ctx context.Context, data []byte, timestampURL string) (*Timestamp, error) {
	// 计算数据哈希
	hash := sha256.Sum256(data)

	// 创建时间戳对象
	timestamp := &Timestamp{
		ID:        dsm.generateID(),
		Hash:      hash,
		Time:      time.Now(),
		URL:       timestampURL,
		CreatedAt: time.Now(),
	}

	// 如果提供了时间戳服务URL，调用时间戳服务
	if timestampURL != "" {
		// 这里应该调用外部时间戳服务
		// 简化处理，返回基本的时间戳
		timestamp.TSAInfo = fmt.Sprintf("TSA: %s", timestampURL)
	}

	return timestamp, nil
}

// verifyTimestamp 验证时间戳
func (dsm *DigitalSignatureManager) verifyTimestamp(ctx context.Context, timestamp *Timestamp, originalData []byte) (TimestampStatus, error) {
	if timestamp == nil {
		return TimestampStatusUnavailable, fmt.Errorf("时间戳为空")
	}

	// 验证时间戳
	originalHash := sha256.Sum256(originalData)
	if !bytes.Equal(timestamp.Hash, originalHash) {
		return TimestampStatusInvalid, nil
	}

	// 检查时间戳有效性
	// 这里应该检查时间戳是否过期（比如24小时内有效）
	maxAge := 24 * time.Hour
	if time.Since(timestamp.Time) > maxAge {
		return TimestampStatusExpired, nil
	}

	return TimestampStatusValid, nil
}

// GetRootPool 获取根证书池
func (ts *TrustStore) GetRootPool() *x509.CertPool {
	pool := x509.NewCertPool()

	for _, cert := range ts.RootCertificates {
		pool.AddCert(cert)
	}

	return pool
}