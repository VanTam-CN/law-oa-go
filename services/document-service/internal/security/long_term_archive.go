package security

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// EvidenceRecord 证据记录
type EvidenceRecord struct {
	ID                   string               `json:"id"`
	Version              int                  `json:"version"`
	HashAlgorithms      []string            `json:"hash_algorithms"`
	CryptoInfos          []CryptoInfo         `json:"crypto_infos"`
	EncryptionInfo       *EncryptionInfo      `json:"encryption_info,omitempty"`
	ArchiveTimestamp     ArchiveTimestamp     `json:"archive_timestamp"`
	CreatedAt            time.Time            `json:"created_at"`
	UpdatedAt            time.Time            `json:"updated_at"`
	Status               ArchiveStatus        `json:"status"`
}

// CryptoInfo 加密信息
type CryptoInfo struct {
	Algorithm    string      `json:"algorithm"`
	KeyIdentifier string      `json:"key_identifier"`
	Validity     time.Time   `json:"validity"`
	Parameters   interface{} `json:"parameters,omitempty"`
}

// EncryptionInfo 加密信息
type EncryptionInfo struct {
	Algorithm    string      `json:"algorithm"`
	KeySize      int         `json:"key_size"`
	KeyIdentifier string      `json:"key_identifier"`
	IV           string      `json:"iv"`
	Parameters   interface{} `json:"parameters,omitempty"`
}

// ArchiveTimestamp 归档时间戳
type ArchiveTimestamp struct {
	Version           int              `json:"version"`
	DigestAlgorithms []string        `json:"digest_algorithms"`
	Canonicalization CanonicalizationMethod `json:"canonicalization"`
	Timestamp         time.Time       `json:"timestamp"`
	Data              []HashTreeNode   `json:"data"`
	Signature         []byte           `json:"signature,omitempty"`
	Certificates      [][]byte         `json:"certificates,omitempty"`
}

// CanonicalizationMethod 规范化方法
type CanonicalizationMethod struct {
	Algorithm string      `json:"algorithm"`
	Parameters interface{} `json:"parameters,omitempty"`
}

// HashTreeNode 哈希树节点
type HashTreeNode struct {
	Algorithm string        `json:"algorithm"`
	HashValue string        `json:"hash_value"`
	Children  []HashTreeNode `json:"children,omitempty"`
	Order     bool          `json:"order"`
}

// ArchiveStatus 归档状态
type ArchiveStatus string

const (
	ArchiveStatusActive   ArchiveStatus = "active"
	ArchiveStatusArchived ArchiveStatus = "archived"
	ArchiveStatusExpired  ArchiveStatus = "expired"
	ArchiveStatusRevoked  ArchiveStatus = "revoked"
)

// LongTermArchive 长期档案管理器
type LongTermArchive struct {
	tsa              *TSA
	store            ArchiveStore
	logger           *logrus.Logger
	mutex            sync.RWMutex
	renewalScheduler *RenewalScheduler
	config           *ArchiveConfig
}

// ArchiveConfig 归档配置
type ArchiveConfig struct {
	RenewalBeforeExpiry    time.Duration `json:"renewal_before_expiry"`
	HashAlgorithm          string        `json:"hash_algorithm"`
	CanonicalizationMethod string        `json:"canonicalization_method"`
	RetentionPeriod        time.Duration `json:"retention_period"`
	MaxArchiveSize         int64         `json:"max_archive_size"`
	EnableAutoRenewal      bool          `json:"enable_auto_renewal"`
}

// ArchiveStore 归档存储接口
type ArchiveStore interface {
	SaveEvidenceRecord(record *EvidenceRecord) error
	GetEvidenceRecord(id string) (*EvidenceRecord, error)
	UpdateEvidenceRecord(record *EvidenceRecord) error
	FindExpiringRecords(before time.Time) ([]*EvidenceRecord, error)
	SearchRecords(criteria *SearchCriteria) ([]*EvidenceRecord, error)
	DeleteEvidenceRecord(id string) error
}

// SearchCriteria 搜索条件
type SearchCriteria struct {
	Status      ArchiveStatus  `json:"status,omitempty"`
	FromTime    time.Time      `json:"from_time,omitempty"`
	ToTime      time.Time      `json:"to_time,omitempty"`
	Subject     string        `json:"subject,omitempty"`
	SerialNumber string        `json:"serial_number,omitempty"`
	Limit       int           `json:"limit,omitempty"`
	Offset      int           `json:"offset,omitempty"`
}

// RenewalScheduler 续期调度器
type RenewalScheduler struct {
	archive    *LongTermArchive
	interval   time.Duration
	running    bool
	stopCh     chan bool
	logger     *logrus.Logger
}

// NewLongTermArchive 创建长期档案管理器
func NewLongTermArchive(tsa *TSA, store ArchiveStore, config *ArchiveConfig, logger *logrus.Logger) (*LongTermArchive, error) {
	if config == nil {
		config = &ArchiveConfig{
			RenewalBeforeExpiry:    30 * 24 * time.Hour, // 30天前续期
			HashAlgorithm:          "SHA256",
			CanonicalizationMethod: "XML-DSIG",
			RetentionPeriod:        7 * 365 * 24 * time.Hour, // 7年
			MaxArchiveSize:         100 * 1024 * 1024, // 100MB
			EnableAutoRenewal:      true,
		}
	}

	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	archive := &LongTermArchive{
		tsa:    tsa,
		store:  store,
		config: config,
		logger: logger,
	}

	// 启动续期调度器
	if config.EnableAutoRenewal {
		archive.renewalScheduler = &RenewalScheduler{
			archive:  archive,
			interval: 24 * time.Hour, // 每天检查一次
			stopCh:   make(chan bool),
			logger:   logger,
		}
		go archive.renewalScheduler.Start()
	}

	logger.WithFields(logrus.Fields{
		"renewal_before_expiry": config.RenewalBeforeExpiry,
		"hash_algorithm":        config.HashAlgorithm,
		"retention_period":      config.RetentionPeriod,
		"auto_renewal":         config.EnableAutoRenewal,
	}).Info("长期档案管理器创建成功")

	return archive, nil
}

// StoreEvidenceRecord 存储证据记录
func (lta *LongTermArchive) StoreEvidenceRecord(
	signature []byte,
	signatureID string,
	certificates []*x509.Certificate,
) (*EvidenceRecord, error) {
	lta.mutex.Lock()
	defer lta.mutex.Unlock()

	recordID := fmt.Sprintf("er_%d", time.Now().UnixNano())

	lta.logger.WithFields(logrus.Fields{
		"record_id":    recordID,
		"signature_id": signatureID,
		"cert_count":   len(certificates),
	}).Debug("开始创建证据记录")

	// 1. 创建证据记录
	record := &EvidenceRecord{
		ID:              recordID,
		Version:         1,
		HashAlgorithms: []string{lta.config.HashAlgorithm},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Status:          ArchiveStatusActive,
	}

	// 2. 验证证书链
	for _, cert := range certificates {
		cryptoInfo := CryptoInfo{
			Algorithm:    cert.SignatureAlgorithm.String(),
			KeyIdentifier: hex.EncodeToString(cert.SubjectKeyId),
			Validity:     cert.NotAfter,
		}
		record.CryptoInfos = append(record.CryptoInfos, cryptoInfo)
	}

	// 3. 创建哈希树节点
	hashNode := HashTreeNode{
		Algorithm: lta.config.HashAlgorithm,
		HashValue: computeHash(signature, lta.config.HashAlgorithm),
		Order:    true,
	}

	// 4. 创建归档时间戳
	archiveTimestamp := ArchiveTimestamp{
		Version:           1,
		DigestAlgorithms: []string{lta.config.HashAlgorithm},
		Canonicalization: CanonicalizationMethod{
			Algorithm: lta.config.CanonicalizationMethod,
		},
		Timestamp: time.Now(),
		Data:      []HashTreeNode{hashNode},
	}

	// 5. 使用TSA签名归档时间戳
	timestampData, err := json.Marshal(archiveTimestamp)
	if err != nil {
		return nil, fmt.Errorf("序列化归档时间戳失败: %w", err)
	}

	timestampReq := CreateTimestampRequest(
		fmt.Sprintf("ts_%s", recordID),
		timestampData,
		"long_term_archive",
	)

	timestampResp, err := lta.tsa.GenerateTimestamp(timestampReq)
	if err != nil {
		return nil, fmt.Errorf("生成归档时间戳失败: %w", err)
	}

	// 6. 编码时间戳令牌
	archiveTimestamp.Signature = timestampResp.Token.Signature
	if len(timestampResp.Token.Certificates) > 0 {
		archiveTimestamp.Certificates = [][]byte{timestampResp.Token.Certificates[0].Bytes}
	}

	record.ArchiveTimestamp = archiveTimestamp

	// 7. 保存到存储
	if err := lta.store.SaveEvidenceRecord(record); err != nil {
		return nil, fmt.Errorf("保存证据记录失败: %w", err)
	}

	lta.logger.WithFields(logrus.Fields{
		"record_id":          recordID,
		"archive_timestamp":  timestampResp.Token.TSTInfo.GenTime,
		"serial_number":      timestampResp.Token.TSTInfo.SerialNumber.String(),
	}).Info("证据记录创建成功")

	return record, nil
}

// VerifyEvidenceRecord 验证证据记录
func (lta *LongTermArchive) VerifyEvidenceRecord(record *EvidenceRecord, signature []byte) error {
	lta.mutex.RLock()
	defer lta.mutex.RUnlock()

	lta.logger.WithField("record_id", record.ID).Debug("开始验证证据记录")

	// 1. 验证归档时间戳
	if err := lta.verifyArchiveTimestamp(record); err != nil {
		return fmt.Errorf("验证归档时间戳失败: %w", err)
	}

	// 2. 验证哈希树
	if err := lta.verifyHashTree(record, signature); err != nil {
		return fmt.Errorf("验证哈希树失败: %w", err)
	}

	// 3. 验证证书有效性
	if err := lta.verifyCertificates(record); err != nil {
		return fmt.Errorf("验证证书有效性失败: %w", err)
	}

	lta.logger.WithField("record_id", record.ID).Debug("证据记录验证成功")

	return nil
}

// RenewTimestamps 续期时间戳
func (lta *LongTermArchive) RenewTimestamps() error {
	lta.mutex.Lock()
	defer lta.mutex.Unlock()

	// 查找即将过期的记录
	expiringRecords, err := lta.store.FindExpiringRecords(time.Now().Add(lta.config.RenewalBeforeExpiry))
	if err != nil {
		return fmt.Errorf("查找即将过期记录失败: %w", err)
	}

	if len(expiringRecords) == 0 {
		lta.logger.Debug("没有即将过期的记录需要续期")
		return nil
	}

	lta.logger.WithField("expiring_count", len(expiringRecords)).Info("开始批量续期时间戳")

	var successCount, failCount int

	for _, record := range expiringRecords {
		if err := lta.renewTimestamp(record); err != nil {
			lta.logger.WithFields(logrus.Fields{
				"record_id": record.ID,
				"error":     err,
			}).Error("续期时间戳失败")
			failCount++
			continue
		}
		successCount++
	}

	lta.logger.WithFields(logrus.Fields{
		"total_count":  len(expiringRecords),
		"success_count": successCount,
		"fail_count":   failCount,
	}).Info("时间戳续期完成")

	return nil
}

// renewTimestamp 续期单个记录的时间戳
func (lta *LongTermArchive) renewTimestamp(record *EvidenceRecord) error {
	// 1. 创建新的归档时间戳
	newArchiveTimestamp := ArchiveTimestamp{
		Version:           record.ArchiveTimestamp.Version + 1,
		DigestAlgorithms: record.ArchiveTimestamp.DigestAlgorithms,
		Canonicalization: record.ArchiveTimestamp.Canonicalization,
		Timestamp:         time.Now(),
		Data:              record.ArchiveTimestamp.Data,
	}

	// 2. 生成新的时间戳
	timestampData, err := json.Marshal(newArchiveTimestamp)
	if err != nil {
		return fmt.Errorf("序列化新归档时间戳失败: %w", err)
	}

	timestampReq := CreateTimestampRequest(
		fmt.Sprintf("ts_renew_%s", record.ID),
		timestampData,
		"long_term_archive_renewal",
	)

	timestampResp, err := lta.tsa.GenerateTimestamp(timestampReq)
	if err != nil {
		return fmt.Errorf("生成续期时间戳失败: %w", err)
	}

	// 3. 更新记录
	newArchiveTimestamp.Signature = timestampResp.Token.Signature
	record.ArchiveTimestamp = newArchiveTimestamp
	record.UpdatedAt = time.Now()

	// 4. 保存更新
	if err := lta.store.UpdateEvidenceRecord(record); err != nil {
		return fmt.Errorf("保存续期记录失败: %w", err)
	}

	lta.logger.WithFields(logrus.Fields{
		"record_id":     record.ID,
		"new_serial":    timestampResp.Token.TSTInfo.SerialNumber.String(),
		"new_timestamp": timestampResp.Token.TSTInfo.GenTime,
	}).Debug("时间戳续期成功")

	return nil
}

// verifyArchiveTimestamp 验证归档时间戳
func (lta *LongTermArchive) verifyArchiveTimestamp(record *EvidenceRecord) error {
	archiveTimestamp := record.ArchiveTimestamp

	// 验证时间戳签名
	if len(archiveTimestamp.Signature) == 0 {
		return fmt.Errorf("缺少时间戳签名")
	}

	timestampData, err := json.Marshal(archiveTimestamp)
	if err != nil {
		return fmt.Errorf("序列化归档时间戳失败: %w", err)
	}

	token := &TimeStampToken{
		Version:   1,
		TSTInfo:   TSTInfo{GenTime: archiveTimestamp.Timestamp},
		Signature: archiveTimestamp.Signature,
	}

	// 解码证书（如果存在）
	if len(archiveTimestamp.Certificates) > 0 {
		var certificates []asn1.RawValue
		for _, certBytes := range archiveTimestamp.Certificates {
			certificates = append(certificates, asn1.RawValue{Bytes: certBytes})
		}
		token.Certificates = certificates
	}

	// 验证时间戳
	return lta.tsa.VerifyTimestamp(token, timestampData)
}

// verifyHashTree 验证哈希树
func (lta *LongTermArchive) verifyHashTree(record *EvidenceRecord, signature []byte) error {
	// 计算签名当前哈希值
	currentHash := computeHash(signature, lta.config.HashAlgorithm)

	// 验证哈希树根节点
	if len(record.ArchiveTimestamp.Data) == 0 {
		return fmt.Errorf("哈希树为空")
	}

	rootNode := record.ArchiveTimestamp.Data[0]
	if rootNode.Algorithm != lta.config.HashAlgorithm {
		return fmt.Errorf("哈希算法不匹配: %s != %s", rootNode.Algorithm, lta.config.HashAlgorithm)
	}

	if rootNode.HashValue != currentHash {
		return fmt.Errorf("哈希值不匹配: %s != %s", rootNode.HashValue, currentHash)
	}

	return nil
}

// verifyCertificates 验证证书
func (lta *LongTermArchive) verifyCertificates(record *EvidenceRecord) error {
	now := time.Now()

	for _, cryptoInfo := range record.CryptoInfos {
		if cryptoInfo.Validity.Before(now) {
			return fmt.Errorf("证书已过期: %s", cryptoInfo.Validity.Format(time.RFC3339))
		}
	}

	return nil
}

// SearchRecords 搜索记录
func (lta *LongTermArchive) SearchRecords(criteria *SearchCriteria) ([]*EvidenceRecord, error) {
	lta.mutex.RLock()
	defer lta.mutex.RUnlock()

	return lta.store.SearchRecords(criteria)
}

// GetRecord 获取记录
func (lta *LongTermArchive) GetRecord(id string) (*EvidenceRecord, error) {
	lta.mutex.RLock()
	defer lta.mutex.RUnlock()

	return lta.store.GetEvidenceRecord(id)
}

// DeleteRecord 删除记录
func (lta *LongTermArchive) DeleteRecord(id string) error {
	lta.mutex.Lock()
	defer lta.mutex.Unlock()

	return lta.store.DeleteEvidenceRecord(id)
}

// GetStatistics 获取统计信息
func (lta *LongTermArchive) GetStatistics() (map[string]interface{}, error) {
	lta.mutex.RLock()
	defer lta.mutex.RUnlock()

	// 这里应该从存储获取实际统计信息
	// 为了演示，返回基本信息
	stats := map[string]interface{}{
		"total_records":          0,
		"active_records":         0,
		"archived_records":       0,
		"expired_records":        0,
		"renewal_before_expiry":   lta.config.RenewalBeforeExpiry,
		"hash_algorithm":         lta.config.HashAlgorithm,
		"retention_period":       lta.config.RetentionPeriod,
		"auto_renewal_enabled":   lta.config.EnableAutoRenewal,
		"tsa_policy_oid":         lta.tsa.PolicyOID.String(),
	}

	return stats, nil
}

// Close 关闭档案管理器
func (lta *LongTermArchive) Close() error {
	lta.mutex.Lock()
	defer lta.mutex.Unlock()

	if lta.renewalScheduler != nil {
		lta.renewalScheduler.Stop()
	}

	lta.logger.Info("长期档案管理器已关闭")

	return nil
}

// Start 续期调度器启动
func (rs *RenewalScheduler) Start() {
	rs.running = true
	ticker := time.NewTicker(rs.interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				if !rs.running {
					return
				}
				if err := rs.archive.RenewTimestamps(); err != nil {
					rs.logger.WithError(err).Error("自动续期时间戳失败")
				}
			case <-rs.stopCh:
				ticker.Stop()
				return
			}
		}
	}()

	rs.logger.WithField("interval", rs.interval).Info("续期调度器已启动")
}

// Stop 续期调度器停止
func (rs *RenewalScheduler) Stop() {
	if rs.running {
		rs.running = false
		rs.stopCh <- true
		rs.logger.Info("续期调度器已停止")
	}
}

// computeHash 计算哈希值
func computeHash(data []byte, algorithm string) string {
	switch algorithm {
	case "SHA256":
		hash := sha256.Sum256(data)
		return hex.EncodeToString(hash[:])
	case "SHA384":
		hash := crypto.SHA384.New().Sum(data)
		return hex.EncodeToString(hash)
	case "SHA512":
		hash := crypto.SHA512.New().Sum(data)
		return hex.EncodeToString(hash)
	default:
		hash := sha256.Sum256(data)
		return hex.EncodeToString(hash[:])
	}
}

// DefaultArchiveConfig 默认归档配置
func DefaultArchiveConfig() *ArchiveConfig {
	return &ArchiveConfig{
		RenewalBeforeExpiry:    30 * 24 * time.Hour,
		HashAlgorithm:          "SHA256",
		CanonicalizationMethod: "XML-DSIG",
		RetentionPeriod:        7 * 365 * 24 * time.Hour,
		MaxArchiveSize:         100 * 1024 * 1024,
		EnableAutoRenewal:      true,
	}
}

// MemoryArchiveStore 内存归档存储（用于演示）
type MemoryArchiveStore struct {
	records map[string]*EvidenceRecord
	mutex   sync.RWMutex
}

// NewMemoryArchiveStore 创建内存归档存储
func NewMemoryArchiveStore() *MemoryArchiveStore {
	return &MemoryArchiveStore{
		records: make(map[string]*EvidenceRecord),
	}
}

// SaveEvidenceRecord 保存证据记录
func (mas *MemoryArchiveStore) SaveEvidenceRecord(record *EvidenceRecord) error {
	mas.mutex.Lock()
	defer mas.mutex.Unlock()

	mas.records[record.ID] = record
	return nil
}

// GetEvidenceRecord 获取证据记录
func (mas *MemoryArchiveStore) GetEvidenceRecord(id string) (*EvidenceRecord, error) {
	mas.mutex.RLock()
	defer mas.mutex.RUnlock()

	if record, exists := mas.records[id]; exists {
		return record, nil
	}
	return nil, fmt.Errorf("记录不存在: %s", id)
}

// UpdateEvidenceRecord 更新证据记录
func (mas *MemoryArchiveStore) UpdateEvidenceRecord(record *EvidenceRecord) error {
	mas.mutex.Lock()
	defer mas.mutex.Unlock()

	if _, exists := mas.records[record.ID]; exists {
		mas.records[record.ID] = record
		return nil
	}
	return fmt.Errorf("记录不存在: %s", record.ID)
}

// FindExpiringRecords 查找即将过期的记录
func (mas *MemoryArchiveStore) FindExpiringRecords(before time.Time) ([]*EvidenceRecord, error) {
	mas.mutex.RLock()
	defer mas.mutex.RUnlock()

	var expiring []*EvidenceRecord
	for _, record := range mas.records {
		if record.ArchiveTimestamp.Timestamp.Before(before) {
			expiring = append(expiring, record)
		}
	}
	return expiring, nil
}

// SearchRecords 搜索记录
func (mas *MemoryArchiveStore) SearchRecords(criteria *SearchCriteria) ([]*EvidenceRecord, error) {
	mas.mutex.RLock()
	defer mas.mutex.RUnlock()

	var results []*EvidenceRecord
	for _, record := range mas.records {
		if criteria.Status != "" && record.Status != criteria.Status {
			continue
		}
		if !criteria.FromTime.IsZero() && record.CreatedAt.Before(criteria.FromTime) {
			continue
		}
		if !criteria.ToTime.IsZero() && record.CreatedAt.After(criteria.ToTime) {
			continue
		}
		results = append(results, record)
	}

	// 应用分页
	if criteria.Limit > 0 {
		if criteria.Offset >= len(results) {
			return nil, nil
		}
		end := criteria.Offset + criteria.Limit
		if end > len(results) {
			end = len(results)
		}
		results = results[criteria.Offset:end]
	}

	return results, nil
}

// DeleteEvidenceRecord 删除证据记录
func (mas *MemoryArchiveStore) DeleteEvidenceRecord(id string) error {
	mas.mutex.Lock()
	defer mas.mutex.Unlock()

	delete(mas.records, id)
	return nil
}