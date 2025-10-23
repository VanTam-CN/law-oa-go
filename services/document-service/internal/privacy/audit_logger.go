package privacy

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"
)

// AuditLogger 审计日志器
type AuditLogger struct {
	logger *slog.Logger
	config *AuditConfig
	buffer *AuditBuffer
	mu     sync.RWMutex
}

// AuditConfig 审计配置
type AuditConfig struct {
	EnableFileLogging bool          `json:"enable_file_logging"`
	LogFilePath       string        `json:"log_file_path"`
	BufferSize        int           `json:"buffer_size"`
	FlushInterval     time.Duration `json:"flush_interval"`
	RotationSize      int64         `json:"rotation_size"`
	RotationInterval  time.Duration `json:"rotation_interval"`
	EnableEncryption  bool          `json:"enable_encryption"`
	EncryptionKey     string        `json:"encryption_key"`
}

// AuditBuffer 审计缓冲区
type AuditBuffer struct {
	entries   []AuditEntry
	maxSize   int
	mutex     sync.Mutex
	flushFunc func([]AuditEntry)
}

// AuditEntry 审计条目
type AuditEntry struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	EventType   string                 `json:"event_type"`
	UserID      string                 `json:"user_id"`
	Username    string                 `json:"username"`
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource"`
	RequestID   string                 `json:"request_id"`
	Data        map[string]interface{} `json:"data"`
	Result      string                 `json:"result"`
	Error       string                 `json:"error,omitempty"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	ProcessingTime time.Duration      `json:"processing_time"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// AuditStatistics 审计统计
type AuditStatistics struct {
	TotalEvents    int64            `json:"total_events"`
	MaskingRequests int64           `json:"masking_requests"`
	MaskingSuccess int64            `json:"masking_success"`
	MaskingErrors  int64            `json:"masking_errors"`
	PrivacyViolations int64         `json:"privacy_violations"`
	AccessAttempts   int64           `json:"access_attempts"`
	EventsByType     map[string]int64 `json:"events_by_type"`
	EventsByUser     map[string]int64 `json:"events_by_user"`
	TopUsers         []UserStat       `json:"top_users"`
	TopResources     []ResourceStat   `json:"top_resources"`
	RecentViolations []AuditEntry   `json:"recent_violations"`
	mu               sync.RWMutex
}

// UserStat 用户统计
type UserStat struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Count     int64  `json:"count"`
	LastSeen  time.Time `json:"last_seen"`
}

// ResourceStat 资源统计
type ResourceStat struct {
	Resource string `json:"resource"`
	Count    int64  `json:"count"`
	LastUsed time.Time `json:"last_used"`
}

// ComplianceEvent 合规事件
type ComplianceEvent struct {
	ID            string                 `json:"id"`
	Timestamp     time.Time              `json:"timestamp"`
	Type          string                 `json:"type"`
	Severity      string                 `json:"severity"`
	Description   string                 `json:"description"`
	AffectedData  string                 `json:"affected_data"`
	UserContext   map[string]interface{} `json:"user_context"`
	Action        string                 `json:"action"`
	Status        string                 `json:"status"`
	Resolved      bool                   `json:"resolved"`
	ResolvedAt    *time.Time            `json:"resolved_at,omitempty"`
}

// PrivacyAlert 隐私告警
type PrivacyAlert struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	AlertType   string                 `json:"alert_type"`
	Severity    string                 `json:"severity"`
	Message     string                 `json:"message"`
	Source      string                 `json:"source"`
	Details     map[string]interface{} `json:"details"`
	Processed   bool                   `json:"processed"`
	ProcessedAt *time.Time            `json:"processed_at,omitempty"`
}

// NewAuditLogger 创建审计日志器
func NewAuditLogger(logger *slog.Logger) *AuditLogger {
	config := &AuditConfig{
		EnableFileLogging: true,
		LogFilePath:       "/tmp/privacy_audit.log",
		BufferSize:        1000,
		FlushInterval:     5 * time.Second,
		RotationSize:      100 * 1024 * 1024, // 100MB
		RotationInterval:  24 * time.Hour,
		EnableEncryption:  false,
	}

	return NewAuditLoggerWithConfig(logger, config)
}

// NewAuditLoggerWithConfig 使用配置创建审计日志器
func NewAuditLoggerWithConfig(logger *slog.Logger, config *AuditConfig) *AuditLogger {
	auditLogger := &AuditLogger{
		logger: logger,
		config: config,
	}

	// 创建缓冲区
	if config.BufferSize > 0 {
		auditLogger.buffer = NewAuditBuffer(config.BufferSize, auditLogger.flushBuffer)
	}

	return auditLogger
}

// LogMaskingRequest 记录脱敏请求
func (al *AuditLogger) LogMaskingRequest(req *MaskingRequest) string {
	entry := &AuditEntry{
		ID:        generateID(),
		Timestamp: time.Now(),
		EventType: "masking_request",
		RequestID: req.RequestID,
		Data: map[string]interface{}{
			"data_length":   len(req.Data),
			"field_count":   len(req.Fields),
			"purpose":       req.Context["purpose"],
			"user_level":    al.getUserLevel(req.UserInfo),
			"department":    req.UserInfo.Department,
			"roles":         req.UserInfo.Roles,
		},
		Result: "pending",
	}

	if req.UserInfo != nil {
		entry.UserID = req.UserInfo.UserID
		entry.Username = req.UserInfo.Username
	}

	al.logEntry(entry)
	return entry.ID
}

// LogMaskingSuccess 记录脱敏成功
func (al *AuditLogger) LogMaskingSuccess(auditID string, req *MaskingRequest, resp *MaskingResponse) {
	entry := &AuditEntry{
		ID:             auditID,
		Timestamp:       time.Now(),
		EventType:       "masking_success",
		RequestID:       req.RequestID,
		Data: map[string]interface{}{
			"original_length": len(req.Data),
			"masked_length":   len(fmt.Sprintf("%v", resp.MaskedData)),
			"processing_time": resp.ProcessingTime.Milliseconds(),
			"strategy":        req.Context["strategy"],
		},
		Result:         "success",
		ProcessingTime: resp.ProcessingTime,
	}

	if req.UserInfo != nil {
		entry.UserID = req.UserInfo.UserID
		entry.Username = req.UserInfo.Username
	}

	al.logEntry(entry)
}

// LogMaskingError 记录脱敏错误
func (al *AuditLogger) LogMaskingError(auditID, fieldName string, err error) {
	entry := &AuditEntry{
		ID:        auditID,
		Timestamp: time.Now(),
		EventType: "masking_error",
		RequestID: auditID,
		Data: map[string]interface{}{
			"field_name": fieldName,
			"error_type": fmt.Sprintf("%T", err),
		},
		Result: "error",
		Error:  err.Error(),
	}

	al.logEntry(entry)
}

// LogAccessAttempt 记录访问尝试
func (al *AuditLogger) LogAccessAttempt(userID, resource, action string, allowed bool, reason string) {
	entry := &AuditEntry{
		ID:        generateID(),
		Timestamp: time.Now(),
		EventType: "access_attempt",
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		Data: map[string]interface{}{
			"allowed": allowed,
		},
		Result: "access_attempt",
	}

	if !allowed {
		entry.Error = reason
	}

	al.logEntry(entry)
}

// LogPrivacyViolation 记录隐私违规
func (al *AuditLogger) LogPrivacyViolation(userID, violationType, description string, severity string) {
	entry := &AuditEntry{
		ID:        generateID(),
		Timestamp: time.Now(),
		EventType: "privacy_violation",
		UserID:    userID,
		Data: map[string]interface{}{
			"violation_type": violationType,
			"description":    description,
			"severity":       severity,
		},
		Result: "violation",
	}

	al.logEntry(entry)
}

// LogComplianceEvent 记录合规事件
func (al *AuditLogger) LogComplianceEvent(event *ComplianceEvent) {
	entry := &AuditEntry{
		ID:        event.ID,
		Timestamp: event.Timestamp,
		EventType: "compliance_event",
		UserID:    getStringFromMap(event.UserContext, "user_id"),
		Data: map[string]interface{}{
			"event_type":    event.Type,
			"severity":      event.Severity,
			"description":   event.Description,
			"affected_data": event.AffectedData,
			"action":        event.Action,
			"status":        event.Status,
			"resolved":      event.Resolved,
		},
		Result: "compliance_event",
	}

	if event.Resolved && event.ResolvedAt != nil {
		entry.Data["resolved_at"] = event.ResolvedAt.Format(time.RFC3339)
	}

	al.logEntry(entry)
}

// LogPrivacyAlert 记录隐私告警
func (al *AuditLogger) LogPrivacyAlert(alert *PrivacyAlert) {
	entry := &AuditEntry{
		ID:        alert.ID,
		Timestamp: alert.Timestamp,
		EventType: "privacy_alert",
		Data: map[string]interface{}{
			"alert_type": alert.AlertType,
			"severity":   alert.Severity,
			"message":    alert.Message,
			"source":     alert.Source,
			"details":    alert.Details,
			"processed":  alert.Processed,
		},
		Result: "alert",
	}

	if alert.Processed && alert.ProcessedAt != nil {
		entry.Data["processed_at"] = alert.ProcessedAt.Format(time.RFC3339)
	}

	al.logEntry(entry)
}

// logEntry 记录审计条目
func (al *AuditLogger) logEntry(entry *AuditEntry) {
	al.mu.Lock()
	defer al.mu.Unlock()

	// 更新统计
	al.updateStatistics(entry)

	// 输出到日志
	al.outputToLogger(entry)

	// 如果启用了文件日志，也输出到文件
	if al.config.EnableFileLogging {
		if al.buffer != nil {
			al.buffer.Add(*entry)
		} else {
			al.writeToFile(entry)
		}
	}
}

// outputToLogger 输出到日志
func (al *AuditLogger) outputToLogger(entry *AuditEntry) {
	attrs := []slog.Attr{
		slog.String("audit_id", entry.ID),
		slog.Time("timestamp", entry.Timestamp),
		slog.String("event_type", entry.EventType),
		slog.String("result", entry.Result),
	}

	if entry.UserID != "" {
		attrs = append(attrs, slog.String("user_id", entry.UserID))
	}
	if entry.Username != "" {
		attrs = append(attrs, slog.String("username", entry.Username))
	}
	if entry.Action != "" {
		attrs = append(attrs, slog.String("action", entry.Action))
	}
	if entry.Resource != "" {
		attrs = append(attrs, slog.String("resource", entry.Resource))
	}
	if entry.RequestID != "" {
		attrs = append(attrs, slog.String("request_id", entry.RequestID))
	}
	if entry.Error != "" {
		attrs = append(attrs, slog.String("error", entry.Error))
	}
	if entry.ProcessingTime > 0 {
		attrs = append(attrs, slog.Duration("processing_time", entry.ProcessingTime))
	}

	for key, value := range entry.Data {
		attrs = append(attrs, slog.Any(key, value))
	}

	// 根据事件类型选择日志级别
	level := slog.LevelInfo
	switch entry.EventType {
	case "masking_error", "privacy_violation", "compliance_violation":
		level = slog.LevelError
	case "access_attempt":
		if entry.Error != "" {
			level = slog.LevelWarn
		}
	case "privacy_alert":
		if entry.Data["severity"] == "high" {
			level = slog.LevelError
		} else if entry.Data["severity"] == "medium" {
			level = slog.LevelWarn
		}
	}

	al.logger.Log(context.Background(), level, "审计事件", attrs...)
}

// writeToFile 写入文件
func (al *AuditLogger) writeToFile(entry *AuditEntry) {
	// 这里可以实现文件写入逻辑
	// 由于复杂性，这里只是简单示例
	data, err := json.Marshal(entry)
	if err != nil {
		al.logger.Error("序列化审计条目失败", "error", err)
		return
	}

	// 简单写入文件（实际实现需要考虑文件轮转、加密等）
	_ = data
}

// flushBuffer 刷新缓冲区
func (al *AuditLogger) flushBuffer(entries []AuditEntry) {
	for _, entry := range entries {
		al.writeToFile(&entry)
	}
}

// getUserLevel 获取用户级别
func (al *AuditLogger) getUserLevel(userInfo *UserInfo) string {
	if userInfo == nil {
		return "unknown"
	}

	for _, role := range userInfo.Roles {
		switch role {
		case "super_admin", "admin":
			return "admin"
		case "partner", "senior_partner":
			return "lawyer"
		case "lawyer", "associate":
			return "lawyer"
		case "assistant", "paralegal":
			return "assistant"
		case "client":
			return "client"
		}
	}

	return "unknown"
}

// updateStatistics 更新统计信息
func (al *AuditLogger) updateStatistics(entry *AuditEntry) {
	// 这里可以实现统计更新逻辑
	// 由于复杂性，这里只是占位符
}

// GetStatistics 获取审计统计
func (al *AuditLogger) GetStatistics() map[string]interface{} {
	// 这里可以实现统计获取逻辑
	return map[string]interface{}{
		"total_events": 0,
		"enabled_features": map[string]bool{
			"file_logging": al.config.EnableFileLogging,
			"encryption":   al.config.EnableEncryption,
		},
	}
}

// SearchEvents 搜索事件
func (al *AuditLogger) SearchEvents(query *EventQuery) ([]AuditEntry, error) {
	// 这里可以实现事件搜索逻辑
	return []AuditEntry{}, nil
}

// EventQuery 事件查询
type EventQuery struct {
	StartTime   *time.Time     `json:"start_time"`
	EndTime     *time.Time     `json:"end_time"`
	UserID      string         `json:"user_id"`
	EventType   string         `json:"event_type"`
	Resource    string         `json:"resource"`
	Action      string         `json:"action"`
	Result      string         `json:"result"`
	Limit       int            `json:"limit"`
	Offset      int            `json:"offset"`
	OrderBy     string         `json:"order_by"`
	SortOrder   string         `json:"sort_order"`
}

// NewAuditBuffer 创建审计缓冲区
func NewAuditBuffer(maxSize int, flushFunc func([]AuditEntry)) *AuditBuffer {
	buffer := &AuditBuffer{
		entries:   make([]AuditEntry, 0, maxSize),
		maxSize:   maxSize,
		flushFunc: flushFunc,
	}

	// 启动定时刷新
	go buffer.startPeriodicFlush()

	return buffer
}

// Add 添加条目到缓冲区
func (ab *AuditBuffer) Add(entry AuditEntry) {
	ab.mutex.Lock()
	defer ab.mutex.Unlock()

	ab.entries = append(ab.entries, entry)

	// 如果达到最大大小，立即刷新
	if len(ab.entries) >= ab.maxSize {
		ab.flush()
	}
}

// flush 刷新缓冲区
func (ab *AuditBuffer) flush() {
	if len(ab.entries) == 0 {
		return
	}

	entries := make([]AuditEntry, len(ab.entries))
	copy(entries, ab.entries)
	ab.entries = ab.entries[:0]

	// 异步执行刷新
	go ab.flushFunc(entries)
}

// startPeriodicFlush 启动定期刷新
func (ab *AuditBuffer) startPeriodicFlush() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ab.flush()
	}
}

// Close 关闭缓冲区
func (ab *AuditBuffer) Close() {
	ab.flush()
}

// generateID 生成ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// getStringFromMap 从map获取字符串值
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, exists := m[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// AuditReport 审计报告
type AuditReport struct {
	Period          string                 `json:"period"`
	StartTime       time.Time              `json:"start_time"`
	EndTime         time.Time              `json:"end_time"`
	Statistics      *AuditStatistics       `json:"statistics"`
	Summary         map[string]interface{} `json:"summary"`
	TopUsers        []UserStat             `json:"top_users"`
	TopResources    []ResourceStat         `json:"top_resources"`
	Violations      []AuditEntry           `json:"violations"`
	ComplianceScore float64               `json:"compliance_score"`
	GeneratedAt     time.Time              `json:"generated_at"`
}

// GenerateReport 生成审计报告
func (al *AuditLogger) GenerateReport(period string, startTime, endTime time.Time) (*AuditReport, error) {
	// 这里可以实现报告生成逻辑
	return &AuditReport{
		Period:          period,
		StartTime:       startTime,
		EndTime:         endTime,
		Statistics:      &AuditStatistics{},
		Summary:         make(map[string]interface{}),
		TopUsers:        []UserStat{},
		TopResources:    []ResourceStat{},
		Violations:      []AuditEntry{},
		ComplianceScore: 0.0,
		GeneratedAt:     time.Now(),
	}, nil
}