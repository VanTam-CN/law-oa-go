package security

import (
	"log/slog"
	"time"

	"gorm.io/gorm"
)

// NewAccessAuditLogger 创建访问审计日志器
func NewAccessAuditLogger(logger *slog.Logger, db *gorm.DB) *AccessAuditLogger {
	auditLogger := &AccessAuditLogger{
		logger: logger,
		db:     db,
	}

	// 自动迁移表
	if err := auditLogger.autoMigrate(); err != nil {
		logger.With("error", err).Error("访问审计表迁移失败")
	}

	return auditLogger
}

// Log 记录访问日志
func (aal *AccessAuditLogger) Log(log *AccessLog) error {
	// 确保时间戳不为零
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now().UTC()
	}

	// 保存到数据库
	if err := aal.db.Create(log).Error; err != nil {
		aal.logger.With("error", err, "request_id", log.RequestID).Error("保存访问日志失败")
		return err
	}

	// 记录结构化日志
	level := slog.LevelInfo
	if !log.Allowed {
		level = slog.LevelWarn
	}

	aal.logger.With("level", level).Log(context.Background(), "访问检查",
		"user_id", log.UserID,
		"resource", log.Resource,
		"action", log.Action,
		"allowed", log.Allowed,
		"reason", log.Reason,
		"ip_address", log.IPAddress,
		"request_id", log.RequestID,
		"duration_ms", log.Duration,
		"timestamp", log.CreatedAt.Format(time.RFC3339),
	)

	return nil
}

// GetAccessLogs 获取访问日志
func (aal *AccessAuditLogger) GetAccessLogs(filter map[string]interface{}, limit, offset int) ([]*AccessLog, int64, error) {
	var logs []*AccessLog
	var total int64

	query := aal.db.Model(&AccessLog{})

	// 应用过滤条件
	if userID, ok := filter["user_id"]; ok {
		query = query.Where("user_id = ?", userID)
	}
	if resource, ok := filter["resource"]; ok {
		query = query.Where("resource = ?", resource)
	}
	if action, ok := filter["action"]; ok {
		query = query.Where("action = ?", action)
	}
	if allowed, ok := filter["allowed"]; ok {
		query = query.Where("allowed = ?", allowed)
	}
	if startTime, ok := filter["start_time"]; ok {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime, ok := filter["end_time"]; ok {
		query = query.Where("created_at <= ?", endTime)
	}
	if ipAddress, ok := filter["ip_address"]; ok {
		query = query.Where("ip_address = ?", ipAddress)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询访问日志总数失败: %w", err)
	}

	// 获取分页数据
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("查询访问日志失败: %w", err)
	}

	return logs, total, nil
}

// GetAccessStats 获取访问统计信息
func (aal *AccessAuditLogger) GetAccessStats(period string) (map[string]interface{}, error) {
	var startTime time.Time
	switch period {
	case "hour":
		startTime = time.Now().Add(-time.Hour)
	case "day":
		startTime = time.Now().AddDate(0, 0, -1)
	case "week":
		startTime = time.Now().AddDate(0, 0, -7)
	case "month":
		startTime = time.Now().AddDate(0, -1, 0)
	default:
		startTime = time.Now().AddDate(0, 0, -1) // 默认一天
	}

	stats := make(map[string]interface{})

	// 总访问次数
	var totalRequests int64
	if err := aal.db.Model(&AccessLog{}).Where("created_at >= ?", startTime).Count(&totalRequests).Error; err != nil {
		return nil, fmt.Errorf("查询总访问次数失败: %w", err)
	}
	stats["total_requests"] = totalRequests

	// 允许的访问次数
	var allowedRequests int64
	if err := aal.db.Model(&AccessLog{}).Where("created_at >= ? AND allowed = ?", startTime, true).Count(&allowedRequests).Error; err != nil {
		return nil, fmt.Errorf("查询允许访问次数失败: %w", err)
	}
	stats["allowed_requests"] = allowedRequests

	// 拒绝的访问次数
	var deniedRequests int64
	if err := aal.db.Model(&AccessLog{}).Where("created_at >= ? AND allowed = ?", startTime, false).Count(&deniedRequests).Error; err != nil {
		return nil, fmt.Errorf("查询拒绝访问次数失败: %w", err)
	}
	stats["denied_requests"] = deniedRequests

	// 计算允许率
	if totalRequests > 0 {
		stats["allowance_rate"] = float64(allowedRequests) / float64(totalRequests)
	} else {
		stats["allowance_rate"] = 0.0
	}

	// 平均响应时间
	var avgDuration sql.NullFloat64
	if err := aal.db.Model(&AccessLog{}).Where("created_at >= ?", startTime).
		Select("AVG(duration)").Scan(&avgDuration).Error; err == nil && avgDuration.Valid {
		stats["avg_duration_ms"] = avgDuration.Float64
	}

	// 按用户统计访问次数
	var userStats []struct {
		UserID string `json:"user_id"`
		Count  int64  `json:"count"`
	}
	if err := aal.db.Model(&AccessLog{}).
		Select("user_id, COUNT(*) as count").
		Where("created_at >= ?", startTime).
		Group("user_id").
		Order("count DESC").
		Limit(10).
		Scan(&userStats).Error; err == nil {
		stats["top_users"] = userStats
	}

	// 按资源统计访问次数
	var resourceStats []struct {
		Resource string `json:"resource"`
		Count    int64  `json:"count"`
	}
	if err := aal.db.Model(&AccessLog{}).
		Select("resource, COUNT(*) as count").
		Where("created_at >= ?", startTime).
		Group("resource").
		Order("count DESC").
		Limit(10).
		Scan(&resourceStats).Error; err == nil {
		stats["top_resources"] = resourceStats
	}

	// 按操作统计访问次数
	var actionStats []struct {
		Action string `json:"action"`
		Count  int64  `json:"count"`
	}
	if err := aal.db.Model(&AccessLog{}).
		Select("action, COUNT(*) as count").
		Where("created_at >= ?", startTime).
		Group("action").
		Order("count DESC").
		Scan(&actionStats).Error; err == nil {
		stats["action_distribution"] = actionStats
	}

	// 按拒绝原因统计
	var reasonStats []struct {
		Reason string `json:"reason"`
		Count  int64  `json:"count"`
	}
	if err := aal.db.Model(&AccessLog{}).
		Select("reason, COUNT(*) as count").
		Where("created_at >= ? AND allowed = ?", startTime, false).
		Group("reason").
		Order("count DESC").
		Limit(10).
		Scan(&reasonStats).Error; err == nil {
		stats["denial_reasons"] = reasonStats
	}

	stats["period"] = period
	stats["start_time"] = startTime.Format(time.RFC3339)

	return stats, nil
}

// GetSuspiciousActivities 获取可疑活动
func (aal *AccessAuditLogger) GetSuspiciousActivities(hours int) ([]*AccessLog, error) {
	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)

	var logs []*AccessLog

	// 查询可疑活动
	// 1. 短时间内多次失败尝试
	subQuery := `
		SELECT user_id, COUNT(*) as failure_count
		FROM access_logs
		WHERE created_at >= ? AND allowed = ?
		GROUP BY user_id
		HAVING failure_count >= 5
	`

	query := aal.db.Where(`
		user_id IN (?) AND created_at >= ? AND allowed = ?
	`, aal.db.Raw(subQuery, startTime, false), startTime, false)

	// 2. 访问敏感资源但被拒绝
	query = query.Or("resource LIKE ? AND allowed = ?", "%sensitive%", false)

	// 3. 异常IP地址访问
	query = query.Or("ip_address != ? AND allowed = ?", "office_network", false)

	if err := query.Order("created_at DESC").Limit(100).Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("查询可疑活动失败: %w", err)
	}

	return logs, nil
}

// GetUserAccessHistory 获取用户访问历史
func (aal *AccessAuditLogger) GetUserAccessHistory(userID string, limit int) ([]*AccessLog, error) {
	var logs []*AccessLog
	if err := aal.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("查询用户访问历史失败: %w", err)
	}

	return logs, nil
}

// GetResourceAccessHistory 获取资源访问历史
func (aal *AccessAuditLogger) GetResourceAccessHistory(resourceID string, limit int) ([]*AccessLog, error) {
	var logs []*AccessLog
	if err := aal.db.Where("resource = ?", resourceID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("查询资源访问历史失败: %w", err)
	}

	return logs, nil
}

// CleanupOldLogs 清理旧日志
func (aal *AccessAuditLogger) CleanupOldLogs(retentionDays int) error {
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	result := aal.db.Where("created_at < ?", cutoffTime).Delete(&AccessLog{})
	if result.Error != nil {
		return fmt.Errorf("清理旧日志失败: %w", result.Error)
	}

	aal.logger.Info("旧日志清理完成",
		"deleted_count", result.RowsAffected,
		"cutoff_time", cutoffTime.Format(time.RFC3339),
		"retention_days", retentionDays,
	)

	return nil
}

// autoMigrate 自动迁移数据库表
func (aal *AccessAuditLogger) autoMigrate() error {
	if err := aal.db.AutoMigrate(&AccessLog{}); err != nil {
		return fmt.Errorf("迁移访问日志表失败: %w", err)
	}

	// 创建索引
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_access_logs_user_id ON access_logs(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_access_logs_resource ON access_logs(resource)",
		"CREATE INDEX IF NOT EXISTS idx_access_logs_action ON access_logs(action)",
		"CREATE INDEX IF NOT EXISTS idx_access_logs_allowed ON access_logs(allowed)",
		"CREATE INDEX IF NOT EXISTS idx_access_logs_created_at ON access_logs(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_access_logs_user_resource ON access_logs(user_id, resource)",
		"CREATE INDEX IF NOT EXISTS idx_access_logs_request_id ON access_logs(request_id)",
	}

	for _, indexSQL := range indexes {
		if err := aal.db.Exec(indexSQL).Error; err != nil {
			aal.logger.With("error", err).Warn("创建索引失败", "index", indexSQL)
		}
	}

	aal.logger.Info("访问审计表迁移完成")
	return nil
}