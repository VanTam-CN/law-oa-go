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
	"encoding/pem"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// TSA 时间戳授权机构
type TSA struct {
	ID              string
	SigningKey      *rsa.PrivateKey
	SigningCert     *x509.Certificate
	PolicyOID        asn1.ObjectIdentifier
	TSAName          string
	Accuracy         Accuracy
	NextSerial       *big.Int
	Logger           *logrus.Logger
	mutex            sync.RWMutex
}

// Accuracy 时间精度
type Accuracy struct {
	Seconds int `json:"seconds,omitempty"`
	Millis  int `json:"millis,omitempty"`
	Micros  int `json:"micros,omitempty"`
}

// MessageImprint 消息摘要
type MessageImprint struct {
	HashAlgorithm   AlgorithmIdentifier `asn1:"optional"`
	HashedMessage   []byte               `asn1:"optional"`
}

// AlgorithmIdentifier 算法标识符
type AlgorithmIdentifier struct {
	Algorithm asn1.ObjectIdentifier
	Parameters asn1.RawValue `asn1:"optional"`
}

// GeneralName 通用名称
type GeneralName struct {
	DirectoryName string `asn1:"optional,tag:4"`
	UniformResourceIdentifier string `asn1:"optional,tag:6"`
	RFC822Name string `asn1:"optional,tag:1"`
}

// TSTInfo 时间戳信息
type TSTInfo struct {
	Version         int                `asn1:"default:1"`
	Policy          asn1.ObjectIdentifier `asn1:"optional"`
	MessageImprint  MessageImprint
	SerialNumber    *big.Int
	GenTime         time.Time
	Accuracy        Accuracy `asn1:"optional"`
	Ordering        bool `asn1:"default:false"`
	Nonce           *big.Int `asn1:"optional"`
	TSA             GeneralName `asn1:"optional"`
	Extensions      []pkix.Extension `asn1:"optional,tag:3"`
}

// TimeStampToken 时间戳令牌
type TimeStampToken struct {
	Version        int
	TSTInfo        TSTInfo
	Signature      []byte
	Certificates   []asn1.RawValue `asn1:"tag:0"`
}

// TimestampRequest 时间戳请求
type TimestampRequest struct {
	ID             string
	Data           []byte
	HashAlgorithm  string
	PolicyOID      string
	Nonce          *big.Int
	CertReq        bool
	RequesterID    string
	CreatedAt      time.Time
}

// TimestampResponse 时间戳响应
type TimestampResponse struct {
	ID             string
	Status         TimestampStatus
	Token          *TimeStampToken
	Error          error
	ProcessingTime time.Duration
	CreatedAt      time.Time
}

// TimestampStatus 时间戳状态
type TimestampStatus struct {
	Status        int    `json:"status"`
	StatusString  string `json:"status_string"`
	FailureInfo   string `json:"failure_info,omitempty"`
}

// 时间戳状态常量
const (
	StatusGranted            = 0
	StatusGrantedWithMods    = 1
	StatusRejection          = 2
	StatusWaiting            = 3
	StatusRevocationWarning   = 4
	StatusRevocationNotification = 5
)

// 算法OID常量
var (
	oidSHA256 = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 1}
	oidSHA384 = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 2}
	oidSHA512 = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 3}
)

// 时间戳策略OID
var (
	oidTSAPolicy1 = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 27164, 1, 1}
	oidTSAPolicy2 = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 27164, 1, 2}
)

// NewTSA 创建时间戳授权机构
func NewTSA(id, tsaName string, signingKey *rsa.PrivateKey, signingCert *x509.Certificate, logger *logrus.Logger) (*TSA, error) {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	// 生成初始序列号
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("生成TSA序列号失败: %w", err)
	}

	tsa := &TSA{
		ID:          id,
		SigningKey:  signingKey,
		SigningCert: signingCert,
		PolicyOID:   oidTSAPolicy1, // 默认使用第一个策略
		TSAName:     tsaName,
		Accuracy:    Accuracy{Seconds: 1},
		NextSerial:  new(big.Int).Add(serialNumber, big.NewInt(1)),
		Logger:      logger,
	}

	logger.WithFields(logrus.Fields{
		"tsa_id":   id,
		"tsa_name": tsaName,
		"policy":   tsa.PolicyOID.String(),
	}).Info("时间戳授权机构创建成功")

	return tsa, nil
}

// GenerateTimestamp 生成时间戳
func (tsa *TSA) GenerateTimestamp(req *TimestampRequest) (*TimestampResponse, error) {
	startTime := time.Now()
	tsa.mutex.Lock()
	defer tsa.mutex.Unlock()

	tsa.Logger.WithFields(logrus.Fields{
		"request_id":   req.ID,
		"requester_id": req.RequesterID,
		"data_size":    len(req.Data),
	}).Debug("开始生成时间戳")

	// 1. 计算数据摘要
	hash, err := tsa.calculateHash(req.Data, req.HashAlgorithm)
	if err != nil {
		return &TimestampResponse{
			ID: req.ID,
			Status: TimestampStatus{
				Status:       StatusRejection,
				StatusString: "rejection",
				FailureInfo:  fmt.Sprintf("计算数据摘要失败: %v", err),
			},
			Error:          err,
			ProcessingTime: time.Since(startTime),
			CreatedAt:      time.Now(),
		}, nil
	}

	// 2. 构造MessageImprint
	var policyOID asn1.ObjectIdentifier
	if req.PolicyOID != "" {
		policyOID, _ = parseOID(req.PolicyOID)
	} else {
		policyOID = tsa.PolicyOID
	}

	var hashAlgOID asn1.ObjectIdentifier
	switch req.HashAlgorithm {
	case "SHA256":
		hashAlgOID = oidSHA256
	case "SHA384":
		hashAlgOID = oidSHA384
	case "SHA512":
		hashAlgOID = oidSHA512
	default:
		hashAlgOID = oidSHA256
	}

	messageImprint := MessageImprint{
		HashAlgorithm: AlgorithmIdentifier{
			Algorithm:  hashAlgOID,
			Parameters: asn1.RawValue{},
		},
		HashedMessage: hash,
	}

	// 3. 构造TSTInfo
	tstInfo := TSTInfo{
		Version:        1,
		Policy:         policyOID,
		MessageImprint: messageImprint,
		SerialNumber:   tsa.NextSerial,
		GenTime:        time.Now().UTC(),
		Accuracy:       tsa.Accuracy,
		Ordering:       true,
		Nonce:          req.Nonce,
		TSA: GeneralName{
			DirectoryName: tsa.TSAName,
		},
	}

	// 4. 签名TSTInfo
	signature, err := tsa.signTSTInfo(tstInfo)
	if err != nil {
		return &TimestampResponse{
			ID: req.ID,
			Status: TimestampStatus{
				Status:       StatusRejection,
				StatusString: "rejection",
				FailureInfo:  fmt.Sprintf("签名失败: %v", err),
			},
			Error:          err,
			ProcessingTime: time.Since(startTime),
			CreatedAt:      time.Now(),
		}, nil
	}

	// 5. 构造时间戳令牌
	timestampToken := &TimeStampToken{
		Version: 1,
		TSTInfo:  tstInfo,
		Signature: signature,
	}

	// 6. 添加证书（如果请求）
	if req.CertReq {
		certBytes, err := x509.MarshalPKIXCertificate(tsa.SigningCert)
		if err == nil {
			timestampToken.Certificates = []asn1.RawValue{
				{
					FullBytes: certBytes,
					Bytes:     certBytes,
				},
			}
		}
	}

	// 7. 更新序列号
	tsa.NextSerial = new(big.Int).Add(tsa.NextSerial, big.NewInt(1))

	processingTime := time.Since(startTime)

	tsa.Logger.WithFields(logrus.Fields{
		"request_id":     req.ID,
		"serial_number":  tstInfo.SerialNumber.String(),
		"gen_time":       tstInfo.GenTime,
		"processing_time": processingTime,
	}).Info("时间戳生成成功")

	return &TimestampResponse{
		ID: req.ID,
		Status: TimestampStatus{
			Status:       StatusGranted,
			StatusString: "granted",
		},
		Token:          timestampToken,
		ProcessingTime: processingTime,
		CreatedAt:      time.Now(),
	}, nil
}

// VerifyTimestamp 验证时间戳
func (tsa *TSA) VerifyTimestamp(token *TimeStampToken, data []byte) error {
	tsa.mutex.RLock()
	defer tsa.mutex.RUnlock()

	tsa.Logger.WithFields(logrus.Fields{
		"serial_number": token.TSTInfo.SerialNumber.String(),
		"data_size":     len(data),
	}).Debug("开始验证时间戳")

	// 1. 验证签名
	if err := tsa.verifySignature(token); err != nil {
		return fmt.Errorf("签名验证失败: %w", err)
	}

	// 2. 验证数据完整性
	if err := tsa.verifyMessageImprint(token, data); err != nil {
		return fmt.Errorf("数据完整性验证失败: %w", err)
	}

	// 3. 验证时间有效性
	if err := tsa.validateTime(token); err != nil {
		return fmt.Errorf("时间有效性验证失败: %w", err)
	}

	// 4. 验证序列号
	if token.TSTInfo.SerialNumber.Cmp(tsa.NextSerial) >= 0 {
		return fmt.Errorf("序列号无效：未来序列号")
	}

	tsa.Logger.WithFields(logrus.Fields{
		"serial_number": token.TSTInfo.SerialNumber.String(),
		"gen_time":      token.TSTInfo.GenTime,
	}).Debug("时间戳验证成功")

	return nil
}

// calculateHash 计算数据摘要
func (tsa *TSA) calculateHash(data []byte, algorithm string) ([]byte, error) {
	switch algorithm {
	case "SHA256":
		hash := sha256.Sum256(data)
		return hash[:], nil
	case "SHA384":
		hash := crypto.SHA384.New().Sum(data)
		return hash, nil
	case "SHA512":
		hash := crypto.SHA512.New().Sum(data)
		return hash, nil
	default:
		hash := sha256.Sum256(data)
		return hash[:], nil
	}
}

// signTSTInfo 签名TSTInfo
func (tsa *TSA) signTSTInfo(tstInfo TSTInfo) ([]byte, error) {
	// 序列化TSTInfo
	tstInfoBytes, err := asn1.Marshal(tstInfo)
	if err != nil {
		return nil, fmt.Errorf("序列化TSTInfo失败: %w", err)
	}

	// 计算TSTInfo的摘要
	hashedInfo := sha256.Sum256(tstInfoBytes)

	// 创建签名
	signature, err := rsa.SignPKCS1v15(rand.Reader, tsa.SigningKey, crypto.SHA256, hashedInfo[:])
	if err != nil {
		return nil, fmt.Errorf("RSA签名失败: %w", err)
	}

	return signature, nil
}

// verifySignature 验证签名
func (tsa *TSA) verifySignature(token *TimeStampToken) error {
	// 序列化TSTInfo
	tstInfoBytes, err := asn1.Marshal(token.TSTInfo)
	if err != nil {
		return fmt.Errorf("序列化TSTInfo失败: %w", err)
	}

	// 计算TSTInfo的摘要
	hashedInfo := sha256.Sum256(tstInfoBytes)

	// 验证签名
	err = rsa.VerifyPKCS1v15(&tsa.SigningKey.PublicKey, crypto.SHA256, hashedInfo[:], token.Signature)
	if err != nil {
		return fmt.Errorf("RSA签名验证失败: %w", err)
	}

	return nil
}

// verifyMessageImprint 验证消息摘要
func (tsa *TSA) verifyMessageImprint(token *TimeStampToken, data []byte) error {
	// 计算当前数据的摘要
	var algorithm string
	switch token.TSTInfo.MessageImprint.HashAlgorithm.Algorithm {
	case oidSHA256:
		algorithm = "SHA256"
	case oidSHA384:
		algorithm = "SHA384"
	case oidSHA512:
		algorithm = "SHA512"
	default:
		algorithm = "SHA256"
	}

	hash, err := tsa.calculateHash(data, algorithm)
	if err != nil {
		return fmt.Errorf("计算数据摘要失败: %w", err)
	}

	// 比较摘要
	if !equalBytes(hash, token.TSTInfo.MessageImprint.HashedMessage) {
		return fmt.Errorf("数据摘要不匹配")
	}

	return nil
}

// validateTime 验证时间有效性
func (tsa *TSA) validateTime(token *TimeStampToken) error {
	// 检查生成时间是否合理
	now := time.Now()
	genTime := token.TSTInfo.GenTime

	// 时间不能是未来时间（允许5秒的时钟偏差）
	if genTime.After(now.Add(5 * time.Second)) {
		return fmt.Errorf("生成时间不能是未来时间: %s", genTime.Format(time.RFC3339))
	}

	// 时间不能太早（TSA创建时间之前）
	if genTime.Before(tsa.SigningCert.NotBefore) {
		return fmt.Errorf("生成时间早于TSA证书生效时间: %s", genTime.Format(time.RFC3339))
	}

	return nil
}

// SetPolicy 设置时间戳策略
func (tsa *TSA) SetPolicy(policyOID string) error {
	tsa.mutex.Lock()
	defer tsa.mutex.Unlock()

	oid, err := parseOID(policyOID)
	if err != nil {
		return fmt.Errorf("无效的策略OID: %w", err)
	}

	tsa.PolicyOID = oid
	tsa.Logger.WithField("policy_oid", policyOID).Info("时间戳策略已更新")

	return nil
}

// SetAccuracy 设置时间精度
func (tsa *TSA) SetAccuracy(seconds, millis, micros int) {
	tsa.mutex.Lock()
	defer tsa.mutex.Unlock()

	tsa.Accuracy = Accuracy{
		Seconds: seconds,
		Millis:  millis,
		Micros:  micros,
	}

	tsa.Logger.WithFields(logrus.Fields{
		"seconds": seconds,
		"millis":  millis,
		"micros":  micros,
	}).Info("时间精度已更新")
}

// GetStatus 获取TSA状态
func (tsa *TSA) GetStatus() map[string]interface{} {
	tsa.mutex.RLock()
	defer tsa.mutex.RUnlock()

	status := map[string]interface{}{
		"id":                  tsa.ID,
		"tsa_name":           tsa.TSAName,
		"policy_oid":         tsa.PolicyOID.String(),
		"next_serial":        tsa.NextSerial.String(),
		"certificate": map[string]interface{}{
			"subject":        tsa.SigningCert.Subject.String(),
			"issuer":         tsa.SigningCert.Issuer.String(),
			"not_before":     tsa.SigningCert.NotBefore,
			"not_after":      tsa.SigningCert.NotAfter,
			"serial_number":  tsa.SigningCert.SerialNumber.String(),
		},
		"accuracy": map[string]interface{}{
			"seconds": tsa.Accuracy.Seconds,
			"millis":  tsa.Accuracy.Millis,
			"micros":  tsa.Accuracy.Micros,
		},
	}

	return status
}

// equalBytes 比较字节数组
func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// parseOID 解析OID字符串
func parseOID(oidStr string) (asn1.ObjectIdentifier, error) {
	var oid asn1.ObjectIdentifier
	if _, err := fmt.Sscanf(oidStr, "%v", &oid); err != nil {
		return nil, fmt.Errorf("无效的OID格式: %w", err)
	}
	return oid, nil
}

// CreateTimestampRequest 创建时间戳请求
func CreateTimestampRequest(id string, data []byte, requesterID string) *TimestampRequest {
	return &TimestampRequest{
		ID:            id,
		Data:          data,
		HashAlgorithm: "SHA256",
		PolicyOID:     "", // 使用TSA默认策略
		Nonce:         generateNonce(),
		CertReq:       true,
		RequesterID:   requesterID,
		CreatedAt:     time.Now(),
	}
}

// generateNonce 生成随机数
func generateNonce() *big.Int {
	nonce, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 64))
	if err != nil {
		return big.NewInt(0)
	}
	return nonce
}

// GenerateRequestID 生成请求ID
func GenerateRequestID() string {
	return fmt.Sprintf("ts_req_%d", time.Now().UnixNano())
}

// EncodeTimestampToken 编码时间戳令牌为DER
func EncodeTimestampToken(token *TimeStampToken) ([]byte, error) {
	return asn1.Marshal(token)
}

// DecodeTimestampToken 从DER解码时间戳令牌
func DecodeTimestampToken(data []byte) (*TimeStampToken, error) {
	var token TimeStampToken
	_, err := asn1.Unmarshal(data, &token)
	if err != nil {
		return nil, fmt.Errorf("解码时间戳令牌失败: %w", err)
	}
	return &token, nil
}

// EncodeTimestampTokenAsPEM 编码时间戳令牌为PEM格式
func EncodeTimestampTokenAsPEM(token *TimeStampToken) ([]byte, error) {
	derBytes, err := EncodeTimestampToken(token)
	if err != nil {
		return nil, err
	}

	block := &pem.Block{
		Type:  "TIMESTAMP TOKEN",
		Bytes: derBytes,
	}

	var buf []byte
	return pem.EncodeToMemory(block), nil
}

// DecodeTimestampTokenFromPEM 从PEM格式解码时间戳令牌
func DecodeTimestampTokenFromPEM(pemData []byte) (*TimeStampToken, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("无效的PEM数据")
	}

	if block.Type != "TIMESTAMP TOKEN" {
		return nil, fmt.Errorf("不是时间戳令牌PEM块")
	}

	return DecodeTimestampToken(block.Bytes)
}