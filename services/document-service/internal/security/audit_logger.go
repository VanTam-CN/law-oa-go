package security

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// AuditLogger 审计日志接口
type AuditLogger interface {
	LogValidationResult(result *ValidationResult, duration time.Duration)
	LogValidationFailure(serialNumber, policy, reason string)
	LogTrustStoreUpdate(storeName, action string, rootsCount, intermediatesCount int)
	LogRevocationCheck(certID, method, result string, duration time.Duration)
	LogCertificateIssuance(certID, serialNumber, subject, templateID string, requesterID string)
	LogCertificateRevocation(certID, serialNumber, reason int)
	LogSecurityEvent(eventType, severity, message string, metadata map[string]interface{})
	Close() error
}

// AuditEvent 审计事件
type AuditEvent struct {
	Timestamp    time.Time              `json:"timestamp"`
	EventType    string                 `json:"event_type"`
	CertID       string                 `json:"cert_id,omitempty"`
	SerialNumber string                 `json:"serial_number,omitempty"`
	Subject      string                 `json:"subject,omitempty"`
	Issuer       string                 `json:"issuer,omitempty"`
	Policy       string                 `json:"policy,omitempty"`
	Result       string                 `json:"result,omitempty"`
	Reason       string                 `json:"reason,omitempty"`
	Duration     int64                  `json:"duration_ms,omitempty"`
	Method       string                 `json:"method,omitempty"`
	Severity     string                 `json:"severity,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// FileAuditLogger 文件审计日志实现
type FileAuditLogger struct {
	file      *os.File
	encoder   *json.Encoder
	mutex     sync.Mutex
	logger    *logrus.Logger
	filePath  string
	maxSize   int64
	current   int64
}

// NewFileAuditLogger 创建文件审计日志器
func NewFileAuditLogger(filePath string) (*FileAuditLogger, error) {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建审计日志目录失败: %w", err)
	}

	// 打开文件
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("打开审计日志文件失败: %w", err)
	}

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &FileAuditLogger{
		file:     file,
		encoder:  json.NewEncoder(file),
		logger:   logger,
		filePath: filePath,
		maxSize:  100 * 1024 * 1024, // 默认100MB
	}, nil
}

// SetMaxSize 设置文件最大大小
func (fal *FileAuditLogger) SetMaxSize(size int64) {
	fal.mutex.Lock()
	defer fal.mutex.Unlock()
	fal.maxSize = size
}

// LogValidationResult 记录验证结果
func (fal *FileAuditLogger) LogValidationResult(result *ValidationResult, duration time.Duration) {
	event := &AuditEvent{
		Timestamp:    result.Timestamp,
		EventType:    "validation_result",
		CertID:       result.ValidationID,
		SerialNumber: result.SerialNumber,
		Subject:      result.Subject,
		Issuer:       result.Issuer,
		Result:       boolToString(result.Valid),
		Duration:     duration.Milliseconds(),
		Metadata: map[string]interface{}{
			"dns_name":              result.DNSName,
			"crl_checked":           result.CRLChecked,
			"ocsp_checked":          result.OCSPChecked,
			"warnings":              result.Warnings,
			"not_before":            result.NotBefore,
			"not_after":             result.NotAfter,
			"is_ca":                 result.IsCA,
			"key_usage":             result.KeyUsage,
			"ext_key_usage":         result.ExtKeyUsage,
			"signature_algorithm":   result.SignatureAlgorithm.String(),
			"public_key_algorithm":  result.PublicKeyAlgorithm.String(),
			"public_key_size":       result.PublicKeySize,
		},
	}

	if result.Error != nil {
		event.Reason = result.Error.Error()
	}

	fal.writeEvent(event)
}

// LogValidationFailure 记录验证失败
func (fal *FileAuditLogger) LogValidationFailure(serialNumber, policy, reason string) {
	event := &AuditEvent{
		Timestamp:    time.Now(),
		EventType:    "validation_failure",
		SerialNumber: serialNumber,
		Policy:       policy,
		Reason:       reason,
		Severity:     "ERROR",
	}

	fal.writeEvent(event)
}

// LogTrustStoreUpdate 记录信任存储更新
func (fal *FileAuditLogger) LogTrustStoreUpdate(storeName, action string, rootsCount, intermediatesCount int) {
	event := &AuditEvent{
		Timestamp: time.Now(),
		EventType: "trust_store_update",
		Severity:  "INFO",
		Metadata: map[string]interface{}{
			"store_name":        storeName,
			"action":            action,
			"roots_count":       rootsCount,
			"intermediates_count": intermediatesCount,
		},
	}

	fal.writeEvent(event)
}

// LogRevocationCheck 记录吊销检查
func (fal *FileAuditLogger) LogRevocationCheck(certID, method, result string, duration time.Duration) {
	event := &AuditEvent{
		Timestamp: time.Now(),
		EventType: "revocation_check",
		CertID:    certID,
		Method:    method,
		Result:    result,
		Duration:  duration.Milliseconds(),
		Severity:  "INFO",
	}

	fal.writeEvent(event)
}

// LogCertificateIssuance 记录证书签发
func (fal *FileAuditLogger) LogCertificateIssuance(certID, serialNumber, subject, templateID string, requesterID string) {
	event := &AuditEvent{
		Timestamp:    time.Now(),
		EventType:    "certificate_issuance",
		CertID:       certID,
		SerialNumber: serialNumber,
		Subject:      subject,
		Severity:     "INFO",
		Metadata: map[string]interface{}{
			"template_id": templateID,
			"requester_id": requesterID,
		},
	}

	fal.writeEvent(event)
}

// LogCertificateRevocation 记录证书吊销
func (fal *FileAuditLogger) LogCertificateRevocation(certID, serialNumber string, reason int) {
	event := &AuditEvent{
		Timestamp:    time.Now(),
		EventType:    "certificate_revocation",
		CertID:       certID,
		SerialNumber: serialNumber,
		Reason:       fmt.Sprintf("Revocation reason: %d", reason),
		Severity:     "WARNING",
	}

	fal.writeEvent(event)
}

// LogSecurityEvent 记录安全事件
func (fal *FileAuditLogger) LogSecurityEvent(eventType, severity, message string, metadata map[string]interface{}) {
	event := &AuditEvent{
		Timestamp: time.Now(),
		EventType: eventType,
		Reason:    message,
		Severity:  severity,
		Metadata:  metadata,
	}

	fal.writeEvent(event)
}

// writeEvent 写入事件到文件
func (fal *FileAuditLogger) writeEvent(event *AuditEvent) {
	fal.mutex.Lock()
	defer fal.mutex.Unlock()

	// 检查文件大小，如果超过限制则轮转
	if err := fal.rotateIfNeeded(); err != nil {
		fal.logger.WithError(err).Error("审计日志文件轮转失败")
		return
	}

	// 写入事件
	if err := fal.encoder.Encode(event); err != nil {
		fal.logger.WithError(err).Error("写入审计事件失败")
		return
	}

	fal.logger.WithFields(logrus.Fields{
		"event_type": event.EventType,
		"cert_id":    event.CertID,
	}).Debug("审计事件已记录")
}

// rotateIfNeeded 需要时轮转文件
func (fal *FileAuditLogger) rotateIfNeeded() error {
	if fal.maxSize <= 0 {
		return nil // 不限制大小
	}

	stat, err := fal.file.Stat()
	if err != nil {
		return err
	}

	if stat.Size() >= fal.maxSize {
		// 关闭当前文件
		fal.file.Close()

		// 重命名当前文件
		timestamp := time.Now().Format("20060102_150405")
		oldPath := fal.filePath + "." + timestamp
		if err := os.Rename(fal.filePath, oldPath); err != nil {
			return err
		}

		fal.logger.WithField("old_path", oldPath).Info("审计日志文件已轮转")

		// 创建新文件
		file, err := os.OpenFile(fal.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}

		fal.file = file
		fal.encoder = json.NewEncoder(file)
	}

	return nil
}

// Close 关闭审计日志器
func (fal *FileAuditLogger) Close() error {
	fal.mutex.Lock()
	defer fal.mutex.Unlock()

	if fal.file != nil {
		return fal.file.Close()
	}
	return nil
}

// DatabaseAuditLogger 数据库审计日志实现
type DatabaseAuditLogger struct {
	db     *sql.DB
	logger *logrus.Logger
	mutex  sync.Mutex
}

// NewDatabaseAuditLogger 创建数据库审计日志器
func NewDatabaseAuditLogger(db *sql.DB, logger *logrus.Logger) (*DatabaseAuditLogger, error) {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	// 创建审计日志表
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS audit_events (
			id BIGSERIAL PRIMARY KEY,
			timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
			event_type VARCHAR(100) NOT NULL,
			cert_id VARCHAR(200),
			serial_number VARCHAR(100),
			subject TEXT,
			issuer TEXT,
			policy VARCHAR(100),
			result VARCHAR(20),
			reason TEXT,
			duration_ms BIGINT,
			method VARCHAR(50),
			severity VARCHAR(20) DEFAULT 'INFO',
			metadata JSONB,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_audit_events_timestamp ON audit_events(timestamp);
		CREATE INDEX IF NOT EXISTS idx_audit_events_cert_id ON audit_events(cert_id);
		CREATE INDEX IF NOT EXISTS idx_audit_events_serial_number ON audit_events(serial_number);
		CREATE INDEX IF NOT EXISTS idx_audit_events_event_type ON audit_events(event_type);
		CREATE INDEX IF NOT EXISTS idx_audit_events_severity ON audit_events(severity);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		return nil, fmt.Errorf("创建审计日志表失败: %w", err)
	}

	return &DatabaseAuditLogger{
		db:     db,
		logger: logger,
	}, nil
}

// LogValidationResult 记录验证结果到数据库
func (dal *DatabaseAuditLogger) LogValidationResult(result *ValidationResult, duration time.Duration) {
	metadata, _ := json.Marshal(map[string]interface{}{
		"dns_name":              result.DNSName,
		"crl_checked":           result.CRLChecked,
		"ocsp_checked":          result.OCSPChecked,
		"warnings":              result.Warnings,
		"not_before":            result.NotBefore,
		"not_after":             result.NotAfter,
		"is_ca":                 result.IsCA,
		"key_usage":             result.KeyUsage,
		"ext_key_usage":         result.ExtKeyUsage,
		"signature_algorithm":   result.SignatureAlgorithm.String(),
		"public_key_algorithm":  result.PublicKeyAlgorithm.String(),
		"public_key_size":       result.PublicKeySize,
	})

	var reason sql.NullString
	if result.Error != nil {
		reason = sql.NullString{String: result.Error.Error(), Valid: true}
	}

	_, err := dal.db.Exec(`
		INSERT INTO audit_events (timestamp, event_type, cert_id, serial_number, subject, issuer,
			result, reason, duration_ms, severity, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`,
		result.Timestamp,
		"validation_result",
		result.ValidationID,
		result.SerialNumber,
		result.Subject,
		result.Issuer,
		boolToString(result.Valid),
		reason,
		duration.Milliseconds(),
		"INFO",
		string(metadata),
	)

	if err != nil {
		dal.logger.WithError(err).Error("写入验证结果到数据库失败")
	}
}

// 实现其他审计日志方法...
func (dal *DatabaseAuditLogger) LogValidationFailure(serialNumber, policy, reason string) {
	dal.logger.Info("数据库审计日志：验证失败")
}

func (dal *DatabaseAuditLogger) LogTrustStoreUpdate(storeName, action string, rootsCount, intermediatesCount int) {
	dal.logger.Info("数据库审计日志：信任存储更新")
}

func (dal *DatabaseAuditLogger) LogRevocationCheck(certID, method, result string, duration time.Duration) {
	dal.logger.Info("数据库审计日志：吊销检查")
}

func (dal *DatabaseAuditLogger) LogCertificateIssuance(certID, serialNumber, subject, templateID string, requesterID string) {
	dal.logger.Info("数据库审计日志：证书签发")
}

func (dal *DatabaseAuditLogger) LogCertificateRevocation(certID, serialNumber string, reason int) {
	dal.logger.Info("数据库审计日志：证书吊销")
}

func (dal *DatabaseAuditLogger) LogSecurityEvent(eventType, severity, message string, metadata map[string]interface{}) {
	dal.logger.Info("数据库审计日志：安全事件")
}

func (dal *DatabaseAuditLogger) Close() error {
	return nil
}

// boolToString 布尔值转字符串
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// MultiAuditLogger 多通道审计日志器
type MultiAuditLogger struct {
	loggers []AuditLogger
	logger  *logrus.Logger
	mutex   sync.RWMutex
}

// NewMultiAuditLogger 创建多通道审计日志器
func NewMultiAuditLogger(logger *logrus.Logger) *MultiAuditLogger {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	return &MultiAuditLogger{
		loggers: make([]AuditLogger, 0),
		logger:  logger,
	}
}

// AddLogger 添加审计日志器
func (mal *MultiAuditLogger) AddLogger(logger AuditLogger) {
	mal.mutex.Lock()
	defer mal.mutex.Unlock()

	mal.loggers = append(mal.loggers, logger)
	mal.logger.Info("审计日志器已添加")
}

// LogValidationResult 记录验证结果到所有日志器
func (mal *MultiAuditLogger) LogValidationResult(result *ValidationResult, duration time.Duration) {
	mal.mutex.RLock()
	loggers := make([]AuditLogger, len(mal.loggers))
	copy(loggers, mal.loggers)
	mal.mutex.RUnlock()

	for _, logger := range loggers {
		logger.LogValidationResult(result, duration)
	}
}

// 实现其他审计日志方法...
func (mal *MultiAuditLogger) LogValidationFailure(serialNumber, policy, reason string) {
	mal.mutex.RLock()
	loggers := make([]AuditLogger, len(mal.loggers))
	copy(loggers, mal.loggers)
	mal.mutex.RUnlock()

	for _, logger := range loggers {
		logger.LogValidationFailure(serialNumber, policy, reason)
	}
}

func (mal *MultiAuditLogger) LogTrustStoreUpdate(storeName, action string, rootsCount, intermediatesCount int) {
	mal.mutex.RLock()
	loggers := make([]AuditLogger, len(mal.loggers))
	copy(loggers, mal.loggers)
	mal.mutex.RUnlock()

	for _, logger := range loggers {
		logger.LogTrustStoreUpdate(storeName, action, rootsCount, intermediatesCount)
	}
}

func (mal *MultiAuditLogger) LogRevocationCheck(certID, method, result string, duration time.Duration) {
	mal.mutex.RLock()
	loggers := make([]AuditLogger, len(mal.loggers))
	copy(loggers, mal.loggers)
	mal.mutex.RUnlock()

	for _, logger := range loggers {
		logger.LogRevocationCheck(certID, method, result, duration)
	}
}

func (mal *MultiAuditLogger) LogCertificateIssuance(certID, serialNumber, subject, templateID string, requesterID string) {
	mal.mutex.RLock()
	loggers := make([]AuditLogger, len(mal.loggers))
	copy(loggers, mal.loggers)
	mal.mutex.RUnlock()

	for _, logger := range loggers {
		logger.LogCertificateIssuance(certID, serialNumber, subject, templateID, requesterID)
	}
}

func (mal *MultiAuditLogger) LogCertificateRevocation(certID, serialNumber string, reason int) {
	mal.mutex.RLock()
	loggers := make([]AuditLogger, len(mal.loggers))
	copy(loggers, mal.loggers)
	mal.mutex.RUnlock()

	for _, logger := range loggers {
		logger.LogCertificateRevocation(certID, serialNumber, reason)
	}
}

func (mal *MultiAuditLogger) LogSecurityEvent(eventType, severity, message string, metadata map[string]interface{}) {
	mal.mutex.RLock()
	loggers := make([]AuditLogger, len(mal.loggers))
	copy(loggers, mal.loggers)
	mal.mutex.RUnlock()

	for _, logger := range loggers {
		logger.LogSecurityEvent(eventType, severity, message, metadata)
	}
}

func (mal *MultiAuditLogger) Close() error {
	mal.mutex.RLock()
	loggers := make([]AuditLogger, len(mal.loggers))
	copy(loggers, mal.loggers)
	mal.mutex.RUnlock()

	for _, logger := range loggers {
		if err := logger.Close(); err != nil {
			mal.logger.WithError(err).Error("关闭审计日志器失败")
		}
	}

	return nil
}