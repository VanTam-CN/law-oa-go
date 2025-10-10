package security

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gorm.io/gorm"
	"law-oa-go/internal/cache"
)

var (
	auditLogDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "audit_log_duration_seconds",
		Help:    "Duration of audit log operations",
		Buckets: []float64{0.001, 0.01, 0.1, 1, 10},
	}, []string{"operation"})

	auditLogErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "audit_log_errors_total",
		Help: "Total number of audit log errors",
	}, []string{"operation", "type"})

	auditLogOperations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "audit_log_operations_total",
		Help: "Total number of audit log operations",
	}, []string{"operation", "category"})
)

// AuditEventType 审计事件类型
type AuditEventType string

const (
	EventTypeLogin            AuditEventType = "login"
	EventTypeLogout           AuditEventType = "logout"
	EventTypePasswordReset    AuditEventType = "password_reset"
	EventTypePermissionChange AuditEventType = "permission_change"
	EventTypeDataAccess       AuditEventType = "data_access"
	EventTypeDataModify       AuditEventType = "data_modify"
	EventTypeDataDelete       AuditEventType = "data_delete"
	EventTypeSystemConfig     AuditEventType = "system_config"
	EventTypeSecurityEvent    AuditEventType = "security_event"
	EventTypeAPIAccess        AuditEventType = "api_access"
	EventTypeFileOperation    AuditEventType = "file_operation"
)

// AuditEventSeverity 审计事件严重程度
type AuditEventSeverity string

const (
	SeverityLow      AuditEventSeverity = "low"
	SeverityMedium   AuditEventSeverity = "medium"
	SeverityHigh     AuditEventSeverity = "high"
	SeverityCritical AuditEventSeverity = "critical"
)

// AuditEvent 审计事件
type AuditEvent struct {
	ID           uint                   `json:"id" gorm:"primaryKey"`
	UserID       uint                   `json:"user_id" gorm:"index"`
	Username     string                 `json:"username" gorm:"index"`
	EventType    AuditEventType         `json:"event_type" gorm:"index"`
	Severity     AuditEventSeverity     `json:"severity" gorm:"index"`
	Action       string                 `json:"action"`
	Resource     string                 `json:"resource"`
	ResourceID   string                 `json:"resource_id"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	DeviceID     string                 `json:"device_id"`
	SessionID    string                 `json:"session_id"`
	Description  string                 `json:"description"`
	Details      map[string]interface{} `json:"details" gorm:"type:json"`
	Status       string                 `json:"status"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Location     string                 `json:"location"`
	Timestamp    time.Time              `json:"timestamp" gorm:"index"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// AuditConfig 审计配置
type AuditLogConfig struct {
	EnableAuditLog       bool
	LogDatabase          bool
	LogToFile            bool
	LogToSyslog          bool
	EnableRealTimeAlert  bool
	SensitiveEventTypes  []AuditEventType
	RequiredEventTypes   []AuditEventType
	RetentionDays        int
	MaxBatchSize         int
	BatchTimeout         time.Duration
	EnableCompression    bool
	EncryptSensitiveData bool
}

// AuditService 审计服务
type AuditService struct {
	config       *AuditConfig
	db           *gorm.DB
	cacheService *cache.CacheService
	eventChan    chan *AuditEvent
	workerCount  int
	stopChan     chan struct{}
}

// NewAuditService 创建审计服务
func NewAuditService(config *AuditConfig, db *gorm.DB, cacheService *cache.CacheService) *AuditService {
	service := &AuditService{
		config:       config,
		db:           db,
		cacheService: cacheService,
		eventChan:    make(chan *AuditEvent, config.MaxBatchSize*2),
		workerCount:  5,
		stopChan:     make(chan struct{}),
	}

	// 启动工作协程
	if config.EnableAuditLog {
		service.startWorkers()
	}

	return service
}

// startWorkers 启动工作协程
func (s *AuditService) startWorkers() {
	for i := 0; i < s.workerCount; i++ {
		go s.worker(i)
	}
}

// worker 工作协程
func (s *AuditService) worker(id int) {
	log.Printf("Audit worker %d started", id)

	batch := make([]*AuditEvent, 0, s.config.MaxBatchSize)
	timer := time.NewTimer(s.config.BatchTimeout)
	defer timer.Stop()

	for {
		select {
		case event := <-s.eventChan:
			batch = append(batch, event)

			if len(batch) >= s.config.MaxBatchSize {
				s.processBatch(batch)
				batch = batch[:0]
				timer.Reset(s.config.BatchTimeout)
			}

		case <-timer.C:
			if len(batch) > 0 {
				s.processBatch(batch)
				batch = batch[:0]
			}
			timer.Reset(s.config.BatchTimeout)

		case <-s.stopChan:
			// 处理剩余事件
			if len(batch) > 0 {
				s.processBatch(batch)
			}
			log.Printf("Audit worker %d stopped", id)
			return
		}
	}
}

// processBatch 处理批次事件
func (s *AuditService) processBatch(events []*AuditEvent) {
	start := time.Now()
	defer func() {
		auditLogDuration.WithLabelValues("process_batch").Observe(time.Since(start).Seconds())
	}()

	if s.config.LogDatabase {
		if err := s.saveToDatabase(events); err != nil {
			log.Printf("Failed to save audit events to database: %v", err)
			auditLogErrors.WithLabelValues("save_database", "database_error").Inc()
		}
	}

	if s.config.LogToFile {
		if err := s.saveToFile(events); err != nil {
			log.Printf("Failed to save audit events to file: %v", err)
			auditLogErrors.WithLabelValues("save_file", "file_error").Inc()
		}
	}

	if s.config.EnableRealTimeAlert {
		s.checkForAlerts(events)
	}

	auditLogOperations.WithLabelValues("process_batch", "batch_processing").Inc()
}

// saveToDatabase 保存到数据库
func (s *AuditService) saveToDatabase(events []*AuditEvent) error {
	start := time.Now()
	defer func() {
		auditLogDuration.WithLabelValues("save_database").Observe(time.Since(start).Seconds())
	}()

	if len(events) == 0 {
		return nil
	}

	// 批量插入
	if err := s.db.CreateInBatches(events, 100).Error; err != nil {
		return fmt.Errorf("failed to insert audit events: %w", err)
	}

	return nil
}

// saveToFile 保存到文件
func (s *AuditService) saveToFile(events []*AuditEvent) error {
	start := time.Now()
	defer func() {
		auditLogDuration.WithLabelValues("save_file").Observe(time.Since(start).Seconds())
	}()

	// TODO: 实现文件日志记录
	// 这里可以添加文件日志记录逻辑
	return nil
}

// checkForAlerts 检查告警
func (s *AuditService) checkForAlerts(events []*AuditEvent) {
	for _, event := range events {
		// 检查安全事件
		if event.EventType == EventTypeSecurityEvent {
			s.triggerSecurityAlert(event)
		}

		// 检查失败登录
		if event.EventType == EventTypeLogin && event.Status == "failed" {
			s.checkLoginFailures(event)
		}

		// 检查敏感操作
		if s.isSensitiveEvent(event) {
			s.triggerSensitiveOperationAlert(event)
		}
	}
}

// triggerSecurityAlert 触发安全告警
func (s *AuditService) triggerSecurityAlert(event *AuditEvent) {
	log.Printf("SECURITY ALERT: %s - %s - %s", event.Username, event.Action, event.Description)
	// TODO: 发送告警通知
}

// checkLoginFailures 检查登录失败
func (s *AuditService) checkLoginFailures(event *AuditEvent) {
	key := fmt.Sprintf("login_failures:%s:%s", event.IPAddress, event.Username)

	var count int
	if s.cacheService != nil {
		if err := s.cacheService.Get(key, &count); err == nil {
			count++
		} else {
			count = 1
		}
		s.cacheService.Set(key, count, time.Hour)
	} else {
		// 没有缓存服务，直接记录日志
		log.Printf("LOGIN FAILURE ALERT: User %s failed login from %s (no cache available)", event.Username, event.IPAddress)
		return
	}

	// 如果5分钟内失败超过5次，触发告警
	if count >= 5 {
		log.Printf("BRUTE FORCE ALERT: Multiple failed login attempts from %s for user %s", event.IPAddress, event.Username)
	}
}

// isSensitiveEvent 检查是否为敏感事件
func (s *AuditService) isSensitiveEvent(event *AuditEvent) bool {
	for _, eventType := range s.config.SensitiveEventTypes {
		if event.EventType == eventType {
			return true
		}
	}
	return false
}

// triggerSensitiveOperationAlert 触发敏感操作告警
func (s *AuditService) triggerSensitiveOperationAlert(event *AuditEvent) {
	log.Printf("SENSITIVE OPERATION ALERT: %s performed %s on %s", event.Username, event.Action, event.Resource)
	// TODO: 发送敏感操作告警
}

// LogEvent 记录审计事件
func (s *AuditService) LogEvent(event *AuditEvent) error {
	start := time.Now()
	defer func() {
		auditLogDuration.WithLabelValues("log_event").Observe(time.Since(start).Seconds())
	}()

	if !s.config.EnableAuditLog {
		return nil
	}

	// 设置默认值
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 加密敏感数据
	if s.config.EncryptSensitiveData {
		event = s.encryptSensitiveData(event)
	}

	// 发送到处理队列
	select {
	case s.eventChan <- event:
		auditLogOperations.WithLabelValues("log_event", string(event.EventType)).Inc()
		return nil
	default:
		// 队列已满，直接处理
		s.processBatch([]*AuditEvent{event})
		return nil
	}
}

// encryptSensitiveData 加密敏感数据
func (s *AuditService) encryptSensitiveData(event *AuditEvent) *AuditEvent {
	// 加密敏感信息
	if event.IPAddress != "" {
		event.IPAddress = s.maskIPAddress(event.IPAddress)
	}

	if event.UserAgent != "" {
		event.UserAgent = s.maskUserAgent(event.UserAgent)
	}

	return event
}

// maskIPAddress 掩码IP地址
func (s *AuditService) maskIPAddress(ip string) string {
	parts := strings.Split(ip, ".")
	if len(parts) == 4 {
		return fmt.Sprintf("%s.%s.*.*", parts[0], parts[1])
	}
	return "***.***.***.***"
}

// maskUserAgent 掩码用户代理
func (s *AuditService) maskUserAgent(ua string) string {
	// 简单的UA掩码，可以根据需要更复杂
	if len(ua) > 20 {
		return ua[:20] + "..."
	}
	return ua
}

// LogLogin 记录登录事件
func (s *AuditService) LogLogin(userID uint, username, ipAddress, userAgent, deviceID, sessionID, status, errorMessage string) error {
	event := &AuditEvent{
		UserID:       userID,
		Username:     username,
		EventType:    EventTypeLogin,
		Severity:     s.getLoginSeverity(status),
		Action:       "user_login",
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		DeviceID:     deviceID,
		SessionID:    sessionID,
		Description:  fmt.Sprintf("User %s login attempt", username),
		Status:       status,
		ErrorMessage: errorMessage,
		Details: map[string]interface{}{
			"login_method": "password",
		},
	}

	return s.LogEvent(event)
}

// LogLogout 记录登出事件
func (s *AuditService) LogLogout(userID uint, username, ipAddress, userAgent, sessionID string) error {
	event := &AuditEvent{
		UserID:      userID,
		Username:    username,
		EventType:   EventTypeLogout,
		Severity:    SeverityLow,
		Action:      "user_logout",
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		SessionID:   sessionID,
		Description: fmt.Sprintf("User %s logged out", username),
		Status:      "success",
	}

	return s.LogEvent(event)
}

// LogPasswordReset 记录密码重置事件
func (s *AuditService) LogPasswordReset(userID uint, username, ipAddress, userAgent, status string) error {
	event := &AuditEvent{
		UserID:      userID,
		Username:    username,
		EventType:   EventTypePasswordReset,
		Severity:    SeverityMedium,
		Action:      "password_reset",
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Description: fmt.Sprintf("User %s password reset", username),
		Status:      status,
	}

	return s.LogEvent(event)
}

// LogPermissionChange 记录权限变更事件
func (s *AuditService) LogPermissionChange(userID uint, username, targetUser, permission, action, ipAddress string) error {
	event := &AuditEvent{
		UserID:      userID,
		Username:    username,
		EventType:   EventTypePermissionChange,
		Severity:    SeverityHigh,
		Action:      "permission_change",
		Resource:    "user_permissions",
		ResourceID:  targetUser,
		IPAddress:   ipAddress,
		Description: fmt.Sprintf("User %s %s permission %s for user %s", username, action, permission, targetUser),
		Status:      "success",
		Details: map[string]interface{}{
			"target_user": targetUser,
			"permission":  permission,
			"action":      action,
		},
	}

	return s.LogEvent(event)
}

// LogDataAccess 记录数据访问事件
func (s *AuditService) LogDataAccess(userID uint, username, resource, resourceID, action, ipAddress string) error {
	event := &AuditEvent{
		UserID:      userID,
		Username:    username,
		EventType:   EventTypeDataAccess,
		Severity:    SeverityMedium,
		Action:      action,
		Resource:    resource,
		ResourceID:  resourceID,
		IPAddress:   ipAddress,
		Description: fmt.Sprintf("User %s accessed %s %s", username, resource, resourceID),
		Status:      "success",
	}

	return s.LogEvent(event)
}

// LogDataModify 记录数据修改事件
func (s *AuditService) LogDataModify(userID uint, username, resource, resourceID, action, ipAddress string, changes map[string]interface{}) error {
	event := &AuditEvent{
		UserID:      userID,
		Username:    username,
		EventType:   EventTypeDataModify,
		Severity:    SeverityMedium,
		Action:      action,
		Resource:    resource,
		ResourceID:  resourceID,
		IPAddress:   ipAddress,
		Description: fmt.Sprintf("User %s modified %s %s", username, resource, resourceID),
		Status:      "success",
		Details:     changes,
	}

	return s.LogEvent(event)
}

// LogDataDelete 记录数据删除事件
func (s *AuditService) LogDataDelete(userID uint, username, resource, resourceID, ipAddress string) error {
	event := &AuditEvent{
		UserID:      userID,
		Username:    username,
		EventType:   EventTypeDataDelete,
		Severity:    SeverityHigh,
		Action:      "delete",
		Resource:    resource,
		ResourceID:  resourceID,
		IPAddress:   ipAddress,
		Description: fmt.Sprintf("User %s deleted %s %s", username, resource, resourceID),
		Status:      "success",
	}

	return s.LogEvent(event)
}

// LogSecurityEvent 记录安全事件
func (s *AuditService) LogSecurityEvent(eventType AuditEventType, username, action, description, ipAddress string, severity AuditEventSeverity) error {
	event := &AuditEvent{
		Username:    username,
		EventType:   eventType,
		Severity:    severity,
		Action:      action,
		IPAddress:   ipAddress,
		Description: description,
		Status:      "success",
	}

	return s.LogEvent(event)
}

// LogAPIAccess 记录API访问事件
func (s *AuditService) LogAPIAccess(userID uint, username, method, endpoint, ipAddress, userAgent, statusCode string) error {
	event := &AuditEvent{
		UserID:      userID,
		Username:    username,
		EventType:   EventTypeAPIAccess,
		Severity:    s.getAPISeverity(statusCode),
		Action:      method,
		Resource:    endpoint,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Description: fmt.Sprintf("API %s %s - %s", method, endpoint, statusCode),
		Status:      statusCode,
		Details: map[string]interface{}{
			"method":      method,
			"endpoint":    endpoint,
			"status_code": statusCode,
		},
	}

	return s.LogEvent(event)
}

// LogFileOperation 记录文件操作事件
func (s *AuditService) LogFileOperation(userID uint, username, action, filePath, ipAddress string, status string) error {
	event := &AuditEvent{
		UserID:      userID,
		Username:    username,
		EventType:   EventTypeFileOperation,
		Severity:    s.getFileOperationSeverity(action),
		Action:      action,
		Resource:    "file",
		ResourceID:  filePath,
		IPAddress:   ipAddress,
		Description: fmt.Sprintf("User %s %s file %s", username, action, filePath),
		Status:      status,
	}

	return s.LogEvent(event)
}

// getLoginSeverity 获取登录事件严重程度
func (s *AuditService) getLoginSeverity(status string) AuditEventSeverity {
	if status == "failed" {
		return SeverityMedium
	}
	return SeverityLow
}

// getAPISeverity 获取API事件严重程度
func (s *AuditService) getAPISeverity(statusCode string) AuditEventSeverity {
	if strings.HasPrefix(statusCode, "4") {
		return SeverityMedium
	}
	if strings.HasPrefix(statusCode, "5") {
		return SeverityHigh
	}
	return SeverityLow
}

// getFileOperationSeverity 获取文件操作严重程度
func (s *AuditService) getFileOperationSeverity(action string) AuditEventSeverity {
	switch action {
	case "delete", "upload":
		return SeverityMedium
	case "download", "share":
		return SeverityLow
	default:
		return SeverityLow
	}
}

// QueryAuditLogs 查询审计日志
func (s *AuditService) QueryAuditLogs(filter AuditLogFilter) ([]*AuditEvent, int64, error) {
	start := time.Now()
	defer func() {
		auditLogDuration.WithLabelValues("query_logs").Observe(time.Since(start).Seconds())
	}()

	// 如果数据库日志记录被禁用，返回空结果
	if !s.config.LogDatabase {
		return []*AuditEvent{}, int64(0), nil
	}

	var events []*AuditEvent
	var total int64

	query := s.db.Model(&AuditEvent{})

	// 应用过滤条件
	if filter.UserID != 0 {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.Username != "" {
		query = query.Where("username LIKE ?", "%"+filter.Username+"%")
	}
	if filter.EventType != "" {
		query = query.Where("event_type = ?", filter.EventType)
	}
	if filter.Severity != "" {
		query = query.Where("severity = ?", filter.Severity)
	}
	if !filter.StartTime.IsZero() {
		query = query.Where("timestamp >= ?", filter.StartTime)
	}
	if !filter.EndTime.IsZero() {
		query = query.Where("timestamp <= ?", filter.EndTime)
	}
	if filter.IPAddress != "" {
		query = query.Where("ip_address = ?", filter.IPAddress)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count audit events: %w", err)
	}

	// 应用排序和分页
	if filter.SortBy != "" {
		sortOrder := "DESC"
		if filter.SortOrder == "asc" {
			sortOrder = "ASC"
		}
		query = query.Order(fmt.Sprintf("%s %s", filter.SortBy, sortOrder))
	} else {
		query = query.Order("timestamp DESC")
	}

	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	// 执行查询
	if err := query.Find(&events).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to query audit events: %w", err)
	}

	// 解密敏感数据
	if s.config.EncryptSensitiveData {
		for _, event := range events {
			s.decryptEventData(event)
		}
	}

	auditLogOperations.WithLabelValues("query_logs", "query").Inc()
	return events, total, nil
}

// decryptEventData 解密事件数据
func (s *AuditService) decryptEventData(event *AuditEvent) {
	// TODO: 实现数据解密逻辑
}

// AuditLogFilter 审计日志过滤器
type AuditLogFilter struct {
	UserID    uint               `json:"user_id"`
	Username  string             `json:"username"`
	EventType AuditEventType     `json:"event_type"`
	Severity  AuditEventSeverity `json:"severity"`
	StartTime time.Time          `json:"start_time"`
	EndTime   time.Time          `json:"end_time"`
	IPAddress string             `json:"ip_address"`
	Page      int                `json:"page"`
	PageSize  int                `json:"page_size"`
	SortBy    string             `json:"sort_by"`
	SortOrder string             `json:"sort_order"`
}

// AuditMiddleware 审计中间件
func (s *AuditService) AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 获取用户信息
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")

		// 处理请求
		c.Next()

		// 记录API访问事件
		if s.config.EnableAuditLog {
			var uid uint
			if userID != nil {
				uid = userID.(uint)
			}

			var uname string
			if username != nil {
				uname = username.(string)
			}

			s.LogAPIAccess(
				uid,
				uname,
				c.Request.Method,
				c.Request.URL.Path,
				c.ClientIP(),
				c.Request.UserAgent(),
				fmt.Sprintf("%d", c.Writer.Status()),
			)
		}

		auditLogDuration.WithLabelValues("middleware").Observe(time.Since(start).Seconds())
	}
}

// Stop 停止审计服务
func (s *AuditService) Stop() {
	if s.config.EnableAuditLog {
		close(s.stopChan)
	}
}

// GetAuditStats 获取审计统计信息
func (s *AuditService) GetAuditStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// 获取今日事件数量
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	var todayCount int64
	s.db.Model(&AuditEvent{}).Where("timestamp >= ? AND timestamp < ?", today, tomorrow).Count(&todayCount)
	stats["today_events"] = todayCount

	// 获取本周事件数量
	weekStart := today.AddDate(0, 0, -int(today.Weekday()))
	var weekCount int64
	s.db.Model(&AuditEvent{}).Where("timestamp >= ?", weekStart).Count(&weekCount)
	stats["week_events"] = weekCount

	// 获取安全事件数量
	var securityCount int64
	s.db.Model(&AuditEvent{}).Where("event_type = ?", EventTypeSecurityEvent).Count(&securityCount)
	stats["security_events"] = securityCount

	return stats
}
