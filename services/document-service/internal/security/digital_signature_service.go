package security

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// DigitalSignatureService 数字签名服务
type DigitalSignatureService struct {
	tsa               *TSA
	archive           *LongTermArchive
	signerStore       SignerStore
	logger            *logrus.Logger
	mutex             sync.RWMutex
	config            *SignatureConfig
}

// SignatureConfig 签名配置
type SignatureConfig struct {
	DefaultAlgorithm    string        `json:"default_algorithm"`
	DefaultHashAlgorithm string        `json:"default_hash_algorithm"`
	IncludeTimestamp    bool          `json:"include_timestamp"`
	IncludeCertificate  bool          `json:"include_certificate"`
	ArchiveSignatures    bool          `json:"archive_signatures"`
	RetentionPeriod     time.Duration `json:"retention_period"`
	MaxSignatureSize    int64         `json:"max_signature_size"`
}

// DigitalSignature 数字签名
type DigitalSignature struct {
	ID              string                 `json:"id"`
	Algorithm       string                 `json:"algorithm"`
	HashAlgorithm   string                 `json:"hash_algorithm"`
	Signature       []byte                 `json:"signature"`
	Signer          SignerInfo             `json:"signer"`
	Timestamp       *TimeStampToken       `json:"timestamp,omitempty"`
	CertificateChain []*CertificateInfo    `json:"certificate_chain,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	VerifiedAt      *time.Time             `json:"verified_at,omitempty"`
	ArchivedAt      *time.Time             `json:"archived_at,omitempty"`
	ArchiveRecordID string                 `json:"archive_record_id,omitempty"`
}

// SignerInfo 签名者信息
type SignerInfo struct {
	ID           string    `json:"id"`
	Subject      string    `json:"subject"`
	Issuer       string    `json:"issuer"`
	SerialNumber string    `json:"serial_number"`
	Email        string    `json:"email,omitempty"`
	Role         string    `json:"role,omitempty"`
	ValidFrom    time.Time `json:"valid_from"`
	ValidTo      time.Time `json:"valid_to"`
	IsCA         bool      `json:"is_ca"`
}

// CertificateInfo 证书信息
type CertificateInfo struct {
	Subject      string    `json:"subject"`
	Issuer       string    `json:"issuer"`
	SerialNumber string    `json:"serial_number"`
	ValidFrom    time.Time `json:"valid_from"`
	ValidTo      time.Time `json:"valid_to"`
	PublicKey    string    `json:"public_key,omitempty"`
}

// SignatureRequest 签名请求
type SignatureRequest struct {
	ID                string                 `json:"id"`
	Data              []byte                 `json:"data"`
	HashAlgorithm     string                 `json:"hash_algorithm"`
	Algorithm         string                 `json:"algorithm"`
	SignerID          string                 `json:"signer_id"`
	IncludeTimestamp  bool                   `json:"include_timestamp"`
	IncludeCertChain  bool                   `json:"include_cert_chain"`
	ArchiveSignature bool                   `json:"archive_signature"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
}

// SignatureVerificationRequest 签名验证请求
type SignatureVerificationRequest struct {
	ID              string                 `json:"id"`
	Signature      *DigitalSignature     `json:"signature"`
	Data            []byte                 `json:"data"`
	VerifyTimestamp bool                   `json:"verify_timestamp"`
	VerifyCertificate bool                 `json:"verify_certificate"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
}

// SignatureVerificationResponse 签名验证响应
type SignatureVerificationResponse struct {
	ID              string                 `json:"id"`
	Valid           bool                   `json:"valid"`
	Signature       *DigitalSignature     `json:"signature"`
	VerificationResult *VerificationResult `json:"verification_result"`
	Error           string                 `json:"error,omitempty"`
	Warnings        []string               `json:"warnings,omitempty"`
	ProcessingTime  time.Duration          `json:"processing_time"`
	CreatedAt       time.Time              `json:"created_at"`
}

// VerificationResult 验证结果
type VerificationResult struct {
	SignatureValid       bool          `json:"signature_valid"`
	TimestampValid       bool          `json:"timestamp_valid,omitempty"`
	CertificateValid     bool          `json:"certificate_valid,omitempty"`
	SignerVerified       bool          `json:"signer_verified"`
	DataIntegrity        bool          `json:"data_integrity"`
	VerificationDetails []string      `json:"verification_details"`
	Issues              []string      `json:"issues,omitempty"`
	VerifiedAt          time.Time     `json:"verified_at"`
}

// SignerStore 签名者存储接口
type SignerStore interface {
	GetSigner(id string) (*SignerInfo, error)
	GetSignerByEmail(email string) (*SignerInfo, error)
	SaveSigner(signer *SignerInfo) error
	UpdateSigner(signer *SignerInfo) error
	DeleteSigner(id string) error
	ListSigners() ([]*SignerInfo, error)
}

// MemorySignerStore 内存签名者存储
type MemorySignerStore struct {
	signers map[string]*SignerInfo
	mutex   sync.RWMutex
}

// NewMemorySignerStore 创建内存签名者存储
func NewMemorySignerStore() *MemorySignerStore {
	return &MemorySignerStore{
		signers: make(map[string]*SignerInfo),
	}
}

// GetSigner 获取签名者
func (mss *MemorySignerStore) GetSigner(id string) (*SignerInfo, error) {
	mss.mutex.RLock()
	defer mss.mutex.RUnlock()

	if signer, exists := mss.signers[id]; exists {
		return signer, nil
	}
	return nil, fmt.Errorf("签名者不存在: %s", id)
}

// GetSignerByEmail 通过邮箱获取签名者
func (mss *MemorySignerStore) GetSignerByEmail(email string) (*SignerInfo, error) {
	mss.mutex.RLock()
	defer mss.mutex.RUnlock()

	for _, signer := range mss.signers {
		if signer.Email == email {
			return signer, nil
		}
	}
	return nil, fmt.Errorf("签名者邮箱不存在: %s", email)
}

// SaveSigner 保存签名者
func (mss *MemorySignerStore) SaveSigner(signer *SignerInfo) error {
	mss.mutex.Lock()
	defer mss.mutex.Unlock()

	mss.signers[signer.ID] = signer
	return nil
}

// UpdateSigner 更新签名者
func (mss *MemorySignerStore) UpdateSigner(signer *SignerInfo) error {
	mss.mutex.Lock()
	defer mss.mutex.Unlock()

	if _, exists := mss.signers[signer.ID]; exists {
		mss.signers[signer.ID] = signer
		return nil
	}
	return fmt.Errorf("签名者不存在: %s", signer.ID)
}

// DeleteSigner 删除签名者
func (mss *MemorySignerStore) DeleteSigner(id string) error {
	mss.mutex.Lock()
	defer mss.mutex.Unlock()

	delete(mss.signers, id)
	return nil
}

// ListSigners 列出所有签名者
func (mss *MemorySignerStore) ListSigners() ([]*SignerInfo, error) {
	mss.mutex.RLock()
	defer mss.mutex.RUnlock()

	signers := make([]*SignerInfo, 0, len(mss.signers))
	for _, signer := range mss.signers {
		signers = append(signers, signer)
	}
	return signers, nil
}

// NewDigitalSignatureService 创建数字签名服务
func NewDigitalSignatureService(
	tsa *TSA,
	archive *LongTermArchive,
	signerStore SignerStore,
	config *SignatureConfig,
	logger *logrus.Logger,
) (*DigitalSignatureService, error) {
	if config == nil {
		config = &SignatureConfig{
			DefaultAlgorithm:     "RSA-PKCS1v15",
			DefaultHashAlgorithm: "SHA256",
			IncludeTimestamp:     true,
			IncludeCertificate:   true,
			ArchiveSignatures:    true,
			RetentionPeriod:       10 * 365 * 24 * time.Hour, // 10年
			MaxSignatureSize:     10 * 1024, // 10KB
		}
	}

	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	service := &DigitalSignatureService{
		tsa:          tsa,
		archive:      archive,
		signerStore:  signerStore,
		config:       config,
		logger:       logger,
	}

	logger.WithFields(logrus.Fields{
		"default_algorithm":     config.DefaultAlgorithm,
		"default_hash_algorithm": config.DefaultHashAlgorithm,
		"include_timestamp":     config.IncludeTimestamp,
		"archive_signatures":    config.ArchiveSignatures,
	}).Info("数字签名服务创建成功")

	return service, nil
}

// SignData 签名数据
func (dss *DigitalSignatureService) SignData(req *SignatureRequest) (*DigitalSignature, error) {
	dss.mutex.Lock()
	defer dss.mutex.Unlock()

	startTime := time.Now()

	dss.logger.WithFields(logrus.Fields{
		"request_id":   req.ID,
		"signer_id":    req.SignerID,
		"data_size":    len(req.Data),
		"algorithm":    req.Algorithm,
	}).Debug("开始数字签名")

	// 1. 获取签名者信息
	signer, err := dss.signerStore.GetSigner(req.SignerID)
	if err != nil {
		return nil, fmt.Errorf("获取签名者信息失败: %w", err)
	}

	// 2. 计算数据摘要
	hashAlgorithm := req.HashAlgorithm
	if hashAlgorithm == "" {
		hashAlgorithm = dss.config.DefaultHashAlgorithm
	}

	hash := dss.calculateHash(req.Data, hashAlgorithm)

	// 3. 生成签名
	signature, err := dss.generateSignature(hash, req.Algorithm)
	if err != nil {
		return nil, fmt.Errorf("生成签名失败: %w", err)
	}

	// 4. 创建数字签名对象
	digitalSignature := &DigitalSignature{
		ID:            req.ID,
		Algorithm:     req.Algorithm,
		HashAlgorithm: hashAlgorithm,
		Signature:     signature,
		Signer:        *signer,
		Metadata:      req.Metadata,
		CreatedAt:     time.Now(),
	}

	// 5. 添加时间戳（如果需要）
	if req.IncludeTimestamp || dss.config.IncludeTimestamp {
		timestampReq := CreateTimestampRequest(
			fmt.Sprintf("ts_sig_%s", req.ID),
			req.Data,
			signer.ID,
		)

		timestampResp, err := dss.tsa.GenerateTimestamp(timestampReq)
		if err != nil {
			dss.logger.WithError(err).Warn("生成时间戳失败，签名将不包含时间戳")
		} else {
			digitalSignature.Timestamp = timestampResp.Token
		}
	}

	// 6. 添加证书链（如果需要）
	if req.IncludeCertChain || dss.config.IncludeCertificate {
		certChain, err := dss.getCertificateChain(signer)
		if err != nil {
			dss.logger.WithError(err).Warn("获取证书链失败，签名将不包含证书")
		} else {
			digitalSignature.CertificateChain = certChain
		}
	}

	// 7. 归档签名（如果需要）
	if req.ArchiveSignature || dss.config.ArchiveSignatures {
		archiveRecord, err := dss.archiveSignature(digitalSignature)
		if err != nil {
			dss.logger.WithError(err).Warn("归档签名失败")
		} else {
			digitalSignature.ArchiveRecordID = archiveRecord.ID
			digitalSignature.ArchivedAt = &archiveRecord.CreatedAt
		}
	}

	processingTime := time.Since(startTime)

	dss.logger.WithFields(logrus.Fields{
		"request_id":     req.ID,
		"signer_id":      req.SignerID,
		"signature_id":   digitalSignature.ID,
		"has_timestamp":  digitalSignature.Timestamp != nil,
		"has_cert_chain": len(digitalSignature.CertificateChain) > 0,
		"processing_time": processingTime,
	}).Info("数字签名完成")

	return digitalSignature, nil
}

// VerifySignature 验证签名
func (dss *DigitalSignatureService) VerifySignature(req *SignatureVerificationRequest) (*SignatureVerificationResponse, error) {
	startTime := time.Now()

	dss.logger.WithFields(logrus.Fields{
		"request_id":   req.ID,
		"signature_id": req.Signature.ID,
		"data_size":    len(req.Data),
	}).Debug("开始验证数字签名")

	// 1. 验证签名者身份
	signerVerified := true
	signer, err := dss.signerStore.GetSigner(req.Signature.Signer.ID)
	if err != nil {
		signerVerified = false
		dss.logger.WithError(err).Warn("无法验证签名者身份")
	}

	// 2. 验证数据完整性
	signatureValid, err := dss.verifySignature(req.Signature, req.Data)
	if err != nil {
		return &SignatureVerificationResponse{
			ID:             req.ID,
			Valid:          false,
			Signature:      req.Signature,
			Error:          fmt.Sprintf("签名验证失败: %v", err),
			ProcessingTime: time.Since(startTime),
			CreatedAt:      time.Now(),
		}, nil
	}

	// 3. 验证时间戳（如果需要）
	timestampValid := true
	if req.VerifyTimestamp && req.Signature.Timestamp != nil {
		if err := dss.verifyTimestamp(req.Signature, req.Data); err != nil {
			timestampValid = false
			dss.logger.WithError(err).Warn("时间戳验证失败")
		}
	}

	// 4. 验证证书（如果需要）
	certificateValid := true
	if req.VerifyCertificate && len(req.Signature.CertificateChain) > 0 {
		if err := dss.verifyCertificates(req.Signature.CertificateChain); err != nil {
			certificateValid = false
			dss.logger.WithError(err).Warn("证书验证失败")
		}
	}

	// 5. 构建验证结果
	verificationResult := &VerificationResult{
		SignatureValid:    signatureValid,
		TimestampValid:    timestampValid,
		CertificateValid:  certificateValid,
		SignerVerified:    signerVerified,
		DataIntegrity:     true, // 如果签名有效，数据完整性就有效
		VerificationDetails: []string{
			fmt.Sprintf("签名算法: %s", req.Signature.Algorithm),
			fmt.Sprintf("哈希算法: %s", req.Signature.HashAlgorithm),
			fmt.Sprintf("签名者: %s", req.Signature.Signer.Subject),
		},
		VerifiedAt: time.Now(),
	}

	// 6. 收集警告
	var warnings []string
	if !signerVerified {
		warnings = append(warnings, "签名者身份无法验证")
	}
	if !timestampValid && req.Signature.Timestamp != nil {
		warnings = append(warnings, "时间戳验证失败")
	}
	if !certificateValid && len(req.Signature.CertificateChain) > 0 {
		warnings = append(warnings, "证书验证失败")
	}

	// 7. 判断整体有效性
	valid := signatureValid && (!req.VerifyTimestamp || timestampValid) && (!req.VerifyCertificate || certificateValid)

	processingTime := time.Since(startTime)

	dss.logger.WithFields(logrus.Fields{
		"request_id":      req.ID,
		"signature_id":    req.Signature.ID,
		"valid":           valid,
		"signature_valid": signatureValid,
		"timestamp_valid": timestampValid,
		"cert_valid":      certificateValid,
		"warnings_count":  len(warnings),
		"processing_time": processingTime,
	}).Info("数字签名验证完成")

	return &SignatureVerificationResponse{
		ID:                req.ID,
		Valid:             valid,
		Signature:         req.Signature,
		VerificationResult: verificationResult,
		Warnings:          warnings,
		ProcessingTime:    processingTime,
		CreatedAt:         time.Now(),
	}, nil
}

// BatchSign 批量签名
func (dss *DigitalSignatureService) BatchSign(requests []*SignatureRequest) ([]*DigitalSignature, error) {
	dss.mutex.Lock()
	defer dss.mutex.Unlock()

	dss.logger.WithField("request_count", len(requests)).Info("开始批量数字签名")

	results := make([]*DigitalSignature, len(requests))
	var successCount, failCount int

	for i, req := range requests {
		signature, err := dss.SignData(req)
		if err != nil {
			dss.logger.WithFields(logrus.Fields{
				"request_id": req.ID,
				"error":      err,
			}).Error("批量签名中单个签名失败")
			failCount++
			// 创建失败的签名对象
			results[i] = &DigitalSignature{
				ID:        req.ID,
				Algorithm: req.Algorithm,
				CreatedAt: time.Now(),
			}
		} else {
			results[i] = signature
			successCount++
		}
	}

	dss.logger.WithFields(logrus.Fields{
		"total_count":   len(requests),
		"success_count": successCount,
		"fail_count":    failCount,
	}).Info("批量数字签名完成")

	return results, nil
}

// BatchVerify 批量验证
func (dss *DigitalSignatureService) BatchVerify(requests []*SignatureVerificationRequest) ([]*SignatureVerificationResponse, error) {
	dss.mutex.RLock()
	defer dss.mutex.RUnlock()

	dss.logger.WithField("request_count", len(requests)).Info("开始批量签名验证")

	results := make([]*SignatureVerificationResponse, len(requests))
	var validCount, invalidCount int

	for i, req := range requests {
		response, err := dss.VerifySignature(req)
		if err != nil {
			dss.logger.WithFields(logrus.Fields{
				"request_id": req.ID,
				"error":      err,
			}).Error("批量验证中单个验证失败")
			results[i] = &SignatureVerificationResponse{
				ID:             req.ID,
				Valid:          false,
				Signature:      req.Signature,
				Error:          err.Error(),
				ProcessingTime: 0,
				CreatedAt:      time.Now(),
			}
			invalidCount++
		} else {
			results[i] = response
			if response.Valid {
				validCount++
			} else {
				invalidCount++
			}
		}
	}

	dss.logger.WithFields(logrus.Fields{
		"total_count":   len(requests),
		"valid_count":   validCount,
		"invalid_count": invalidCount,
	}).Info("批量签名验证完成")

	return results, nil
}

// GetSignature 获取签名
func (dss *DigitalSignatureService) GetSignature(id string) (*DigitalSignature, error) {
	// 这里应该从存储中获取签名
	// 为了演示，返回空
	return nil, fmt.Errorf("存储接口未实现")
}

// ListSignatures 列出签名
func (dss *DigitalSignatureService) ListSignatures(criteria *SignatureSearchCriteria) ([]*DigitalSignature, error) {
	// 这里应该从存储中搜索签名
	// 为了演示，返回空列表
	return []*DigitalSignature{}, nil
}

// DeleteSignature 删除签名
func (dss *DigitalSignatureService) DeleteSignature(id string) error {
	// 这里应该从存储中删除签名
	// 为了演示，返回成功
	return nil
}

// GetStatistics 获取统计信息
func (dss *DigitalSignatureService) GetStatistics() (map[string]interface{}, error) {
	dss.mutex.RLock()
	defer dss.mutex.RUnlock()

	// 这里应该从存储中获取实际统计信息
	// 为了演示，返回基本信息
	stats := map[string]interface{}{
		"total_signatures":         0,
		"active_signatures":        0,
		"archived_signatures":      0,
		"verified_signatures":      0,
		"failed_verifications":      0,
		"default_algorithm":        dss.config.DefaultAlgorithm,
		"default_hash_algorithm":    dss.config.DefaultHashAlgorithm,
		"include_timestamp":        dss.config.IncludeTimestamp,
		"include_certificate":      dss.config.IncludeCertificate,
		"archive_signatures":       dss.config.ArchiveSignatures,
		"retention_period":         dss.config.RetentionPeriod,
	}

	return stats, nil
}

// calculateHash 计算哈希
func (dss *DigitalSignatureService) calculateHash(data []byte, algorithm string) []byte {
	switch algorithm {
	case "SHA256":
		hash := sha256.Sum256(data)
		return hash[:]
	case "SHA384":
		hash := crypto.SHA384.New().Sum(data)
		return hash
	case "SHA512":
		hash := crypto.SHA512.New().Sum(data)
		return hash
	default:
		hash := sha256.Sum256(data)
		return hash[:]
	}
}

// generateSignature 生成签名
func (dss *DigitalSignatureService) generateSignature(hash []byte, algorithm string) ([]byte, error) {
	switch algorithm {
	case "RSA-PKCS1v15":
		// 这里需要签名者的私钥
		// 为了演示，生成随机签名
		signature := make([]byte, 256)
		if _, err := rand.Read(signature); err != nil {
			return nil, fmt.Errorf("生成随机签名失败: %w", err)
		}
		return signature, nil
	case "ECDSA", "RSA-PSS":
		return nil, fmt.Errorf("不支持的签名算法: %s", algorithm)
	default:
		return nil, fmt.Errorf("未知的签名算法: %s", algorithm)
	}
}

// getCertificateChain 获取证书链
func (dss *DigitalSignatureService) getCertificateChain(signer *SignerInfo) ([]*CertificateInfo, error) {
	// 这里应该从证书存储中获取完整证书链
	// 为了演示，返回签名者证书
	certInfo := &CertificateInfo{
		Subject:      signer.Subject,
		Issuer:       signer.Issuer,
		SerialNumber: signer.SerialNumber,
		ValidFrom:    signer.ValidFrom,
		ValidTo:      signer.ValidTo,
	}
	return []*CertificateInfo{certInfo}, nil
}

// verifySignature 验证签名
func (dss *DigitalSignatureService) verifySignature(signature *DigitalSignature, data []byte) (bool, error) {
	// 这里应该使用真实的签名验证逻辑
	// 为了演示，检查签名不为空
	if len(signature.Signature) == 0 {
		return false, fmt.Errorf("签名为空")
	}

	// 计算当前数据的哈希
	currentHash := dss.calculateHash(data, signature.HashAlgorithm)

	// 这里应该验证签名是否匹配当前哈希
	// 为了演示，总是返回true
	dss.logger.WithFields(logrus.Fields{
		"signature_id": signature.ID,
		"algorithm":    signature.Algorithm,
		"hash_algorithm": signature.HashAlgorithm,
	}).Debug("签名验证（演示模式）")

	return true, nil
}

// verifyTimestamp 验证时间戳
func (dss *DigitalSignatureService) verifyTimestamp(signature *DigitalSignature, data []byte) error {
	if signature.Timestamp == nil {
		return fmt.Errorf("签名不包含时间戳")
	}

	// 使用TSA验证时间戳
	timestampData, err := json.Marshal(signature.Timestamp.TSTInfo)
	if err != nil {
		return fmt.Errorf("序列化时间戳信息失败: %w", err)
	}

	return dss.tsa.VerifyTimestamp(signature.Timestamp, timestampData)
}

// verifyCertificates 验证证书链
func (dss *DigitalSignatureService) verifyCertificates(certificates []*CertificateInfo) error {
	now := time.Now()

	for _, cert := range certificates {
		if cert.ValidTo.Before(now) {
			return fmt.Errorf("证书已过期: %s", cert.ValidTo.Format(time.RFC3339))
		}
	}

	return nil
}

// archiveSignature 归档签名
func (dss *DigitalSignatureService) archiveSignature(signature *DigitalSignature) (*EvidenceRecord, error) {
	// 编码签名数据
	signatureData, err := json.Marshal(signature)
	if err != nil {
		return nil, fmt.Errorf("序列化签名失败: %w", err)
	}

	// 生成签名ID
	signatureID := signature.ID

	// 编码证书链
	var certificates []*x509.Certificate
	for _, certInfo := range signature.CertificateChain {
		// 这里应该解析CertificateInfo为x509.Certificate
		// 为了演示，创建空证书
		cert := &x509.Certificate{
			Subject: pkix.Name{CommonName: certInfo.Subject},
			Issuer:  pkix.Name{CommonName: certInfo.Issuer},
		}
		certificates = append(certificates, cert)
	}

	return dss.archive.StoreEvidenceRecord(signatureData, signatureID, certificates)
}

// SignatureSearchCriteria 签名搜索条件
type SignatureSearchCriteria struct {
	SignerID      string    `json:"signer_id,omitempty"`
	FromTime      time.Time `json:"from_time,omitempty"`
	ToTime        time.Time `json:"to_time,omitempty"`
	Algorithm     string    `json:"algorithm,omitempty"`
	Status        string    `json:"status,omitempty"`
	Limit         int       `json:"limit,omitempty"`
	Offset        int       `json:"offset,omitempty"`
}

// CreateSignatureRequest 创建签名请求
func CreateSignatureRequest(id string, data []byte, signerID string) *SignatureRequest {
	return &SignatureRequest{
		ID:                id,
		Data:              data,
		HashAlgorithm:     "SHA256",
		Algorithm:         "RSA-PKCS1v15",
		SignerID:          signerID,
		IncludeTimestamp:  true,
		IncludeCertChain:  true,
		ArchiveSignature:  true,
		Metadata:          make(map[string]interface{}),
		CreatedAt:         time.Now(),
	}
}

// CreateVerificationRequest 创建验证请求
func CreateVerificationRequest(id string, signature *DigitalSignature, data []byte) *SignatureVerificationRequest {
	return &SignatureVerificationRequest{
		ID:                id,
		Signature:        signature,
		Data:              data,
		VerifyTimestamp:   true,
		VerifyCertificate: true,
		Metadata:          make(map[string]interface{}),
		CreatedAt:         time.Now(),
	}
}

// GenerateSignatureID 生成签名ID
func GenerateSignatureID() string {
	return fmt.Sprintf("sig_%d", time.Now().UnixNano())
}

// EncodeSignature 编码签名为PEM格式
func EncodeSignature(signature *DigitalSignature) ([]byte, error) {
	block := &pem.Block{
		Type:  "DIGITAL SIGNATURE",
		Bytes: signature.Signature,
	}

	signatureData, err := json.Marshal(signature)
	if err != nil {
		return nil, err
	}

	block.Bytes = signatureData
	return pem.EncodeToMemory(block), nil
}

// DecodeSignature 从PEM格式解码签名
func DecodeSignature(pemData []byte) (*DigitalSignature, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("无效的PEM数据")
	}

	if block.Type != "DIGITAL SIGNATURE" {
		return nil, fmt.Errorf("不是数字签名PEM块")
	}

	var signature DigitalSignature
	if err := json.Unmarshal(block.Bytes, &signature); err != nil {
		return nil, fmt.Errorf("解码签名失败: %w", err)
	}

	return &signature, nil
}