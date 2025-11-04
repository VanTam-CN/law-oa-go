package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"law-oa-go/internal/models"
)

// 统计仓库实现
func (r *enhancedConflictRepository) GetConflictCheckStatistics(ctx context.Context, query *StatisticsQuery) (*models.ProfessionalConflictCheckStats, error) {
	var stats models.ProfessionalConflictCheckStats

	// 设置统计期间
	if !query.StartDate.IsZero() {
		stats.PeriodStart = query.StartDate
	} else {
		stats.PeriodStart = time.Now().AddDate(0, -1, 0) // 默认最近一个月
	}

	if !query.EndDate.IsZero() {
		stats.PeriodEnd = query.EndDate
	} else {
		stats.PeriodEnd = time.Now()
	}

	// 获取请求数量统计
	if err := r.db.WithContext(ctx).
		Model(&models.ProfessionalConflictCheckRequest{}).
		Where("requested_date BETWEEN ? AND ?", stats.PeriodStart, stats.PeriodEnd).
		Count(&stats.TotalRequests).Error; err != nil {
		return nil, fmt.Errorf("获取总请求数统计失败: %w", err)
	}

	// 获取待处理请求数
	if err := r.db.WithContext(ctx).
		Model(&models.ProfessionalConflictCheckRequest{}).
		Where("status IN ('PENDING', 'IN_PROGRESS')").
		Count(&stats.PendingRequests).Error; err != nil {
		return nil, fmt.Errorf("获取待处理请求统计失败: %w", err)
	}

	// 获取已完成请求数
	if err := r.db.WithContext(ctx).
		Model(&models.ProfessionalConflictCheckRequest{}).
		Where("status = 'COMPLETED' AND requested_date BETWEEN ? AND ?",
			stats.PeriodStart, stats.PeriodEnd).
		Count(&stats.CompletedRequests).Error; err != nil {
		return nil, fmt.Errorf("获取已完成请求统计失败: %w", err)
	}

	// 获取冲突检测统计
	if err := r.db.WithContext(ctx).
		Model(&models.MultidimensionalConflictResult{}).
		Joins("JOIN professional_conflict_check_requests ON multidimensional_conflict_results.check_request_id = professional_conflict_check_requests.id").
		Where("professional_conflict_check_requests.requested_date BETWEEN ? AND ?",
			stats.PeriodStart, stats.PeriodEnd).
		Count(&stats.TotalConflicts).Error; err != nil {
		return nil, fmt.Errorf("获取冲突总数统计失败: %w", err)
	}

	// 获取各严重级别冲突数
	if err := r.db.WithContext(ctx).
		Model(&models.MultidimensionalConflictResult{}).
		Joins("JOIN professional_conflict_check_requests ON multidimensional_conflict_results.check_request_id = professional_conflict_check_requests.id").
		Where("professional_conflict_check_requests.requested_date BETWEEN ? AND ? AND severity_level = 'CRITICAL'",
			stats.PeriodStart, stats.PeriodEnd).
		Count(&stats.CriticalConflicts).Error; err != nil {
		return nil, fmt.Errorf("获取严重冲突统计失败: %w", err)
	}

	if err := r.db.WithContext(ctx).
		Model(&models.MultidimensionalConflictResult{}).
		Joins("JOIN professional_conflict_check_requests ON multidimensional_conflict_results.check_request_id = professional_conflict_check_requests.id").
		Where("professional_conflict_check_requests.requested_date BETWEEN ? AND ? AND severity_level = 'HIGH'",
			stats.PeriodStart, stats.PeriodEnd).
		Count(&stats.HighConflicts).Error; err != nil {
		return nil, fmt.Errorf("获取高风险冲突统计失败: %w", err)
	}

	if err := r.db.WithContext(ctx).
		Model(&models.MultidimensionalConflictResult{}).
		Joins("JOIN professional_conflict_check_requests ON multidimensional_conflict_results.check_request_id = professional_conflict_check_requests.id").
		Where("professional_conflict_check_requests.requested_date BETWEEN ? AND ? AND severity_level = 'MEDIUM'",
			stats.PeriodStart, stats.PeriodEnd).
		Count(&stats.MediumConflicts).Error; err != nil {
		return nil, fmt.Errorf("获取中风险冲突统计失败: %w", err)
	}

	if err := r.db.WithContext(ctx).
		Model(&models.MultidimensionalConflictResult{}).
		Joins("JOIN professional_conflict_check_requests ON multidimensional_conflict_results.check_request_id = professional_conflict_check_requests.id").
		Where("professional_conflict_check_requests.requested_date BETWEEN ? AND ? AND severity_level = 'LOW'",
			stats.PeriodStart, stats.PeriodEnd).
		Count(&stats.LowConflicts).Error; err != nil {
		return nil, fmt.Errorf("获取低风险冲突统计失败: %w", err)
	}

	// 获取处理效率统计
	var avgProcessingTime sql.NullInt64
	if err := r.db.WithContext(ctx).
		Model(&models.ProfessionalConflictCheckRequest{}).
		Where("status = 'COMPLETED' AND requested_date BETWEEN ? AND ?",
			stats.PeriodStart, stats.PeriodEnd).
		Select("AVG(TIMESTAMPDIFF(MINUTE, started_at, completed_at))").
		Scan(&avgProcessingTime).Error; err != nil {
		return nil, fmt.Errorf("获取平均处理时间统计失败: %w", err)
	}

	if avgProcessingTime.Valid {
		stats.AvgProcessingTime = int(avgProcessingTime.Int64)
	}

	// 计算SLA合规率
	var onTimeCount int64
	var totalCount int64

	r.db.WithContext(ctx).
		Model(&models.ProfessionalConflictCheckRequest{}).
		Where("status = 'COMPLETED' AND requested_date BETWEEN ? AND ?",
			stats.PeriodStart, stats.PeriodEnd).
		Count(&totalCount)

	if totalCount > 0 {
		r.db.WithContext(ctx).
			Model(&models.ProfessionalConflictCheckRequest{}).
			Where("status = 'COMPLETED' AND completed_at <= required_by_date AND requested_date BETWEEN ? AND ?",
				stats.PeriodStart, stats.PeriodEnd).
			Count(&onTimeCount)

		stats.SlaComplianceRate = float64(onTimeCount) / float64(totalCount) * 100
	}

	// 获取豁免统计
	r.db.WithContext(ctx).
		Model(&models.WaiverApplication{}).
		Where("submission_date BETWEEN ? AND ?", stats.PeriodStart, stats.PeriodEnd).
		Count(&stats.WaiverRequests)

	r.db.WithContext(ctx).
		Model(&models.WaiverApplication{}).
		Where("status = 'APPROVED' AND submission_date BETWEEN ? AND ?",
			stats.PeriodStart, stats.PeriodEnd).
		Count(&stats.WaiversApproved)

	r.db.WithContext(ctx).
		Model(&models.WaiverApplication{}).
		Where("status = 'REJECTED' AND submission_date BETWEEN ? AND ?",
			stats.PeriodStart, stats.PeriodEnd).
		Count(&stats.WaiversRejected)

	return &stats, nil
}

func (r *enhancedConflictRepository) GetConflictTrends(ctx context.Context, period string) ([]*ConflictTrend, error) {
	var trends []*ConflictTrend

	var startDate time.Time
	var groupBy string

	switch period {
	case "daily":
		startDate = time.Now().AddDate(0, 0, -30)
		groupBy = "DATE(requested_date)"
	case "weekly":
		startDate = time.Now().AddDate(0, 0, -12)
		groupBy = "YEARWEEK(requested_date)"
	case "monthly":
		startDate = time.Now().AddDate(0, -12, 0)
		groupBy = "DATE_FORMAT(requested_date, '%Y-%m')"
	default:
		return nil, fmt.Errorf("不支持的时间周期: %s", period)
	}

	query := `
		SELECT
			` + groupBy + ` as period,
			COUNT(*) as conflicts,
			SUM(CASE WHEN status = 'RESOLVED' THEN 1 ELSE 0 END) as resolutions,
			SUM(CASE WHEN status IN ('DETECTED', 'UNDER_REVIEW') THEN 1 ELSE 0 END) as pending
		FROM multidimensional_conflict_results
		WHERE created_at >= ?
		GROUP BY ` + groupBy + `
		ORDER BY period
	`

	if err := r.db.WithContext(ctx).
		Raw(query, startDate).
		Scan(&trends).Error; err != nil {
		return nil, fmt.Errorf("获取冲突趋势数据失败: %w", err)
	}

	return trends, nil
}

func (r *enhancedConflictRepository) GetRiskDistribution(ctx context.Context, filters *RiskDistributionFilters) (*RiskDistribution, error) {
	var distribution RiskDistribution

	query := r.db.WithContext(ctx).Model(&models.ClientRiskProfile{})

	if !filters.StartDate.IsZero() {
		query = query.Where("last_assessment_date >= ?", filters.StartDate)
	}

	if !filters.EndDate.IsZero() {
		query = query.Where("last_assessment_date <= ?", filters.EndDate)
	}

	if len(filters.ClientIDs) > 0 {
		query = query.Where("client_id IN ?", filters.ClientIDs)
	}

	// 统计各风险等级客户数
	query.Where("overall_risk = 'LOW'").Count(&distribution.Low)
	query.Where("overall_risk = 'MEDIUM'").Count(&distribution.Medium)
	query.Where("overall_risk = 'HIGH'").Count(&distribution.High)
	query.Where("overall_risk = 'CRITICAL'").Count(&distribution.Critical)

	return &distribution, nil
}

func (r *enhancedConflictRepository) GetWaiverStatistics(ctx context.Context, query *StatisticsQuery) (*WaiverStatistics, error) {
	var stats WaiverStatistics

	startDate := query.StartDate
	if startDate.IsZero() {
		startDate = time.Now().AddDate(0, -1, 0)
	}

	endDate := query.EndDate
	if endDate.IsZero() {
		endDate = time.Now()
	}

	// 获取总申请数
	r.db.WithContext(ctx).
		Model(&models.WaiverApplication{}).
		Where("submission_date BETWEEN ? AND ?", startDate, endDate).
		Count(&stats.TotalRequests)

	// 获取各状态申请数
	r.db.WithContext(ctx).
		Model(&models.WaiverApplication{}).
		Where("status = 'APPROVED' AND submission_date BETWEEN ? AND ?", startDate, endDate).
		Count(&stats.ApprovedCount)

	r.db.WithContext(ctx).
		Model(&models.WaiverApplication{}).
		Where("status = 'REJECTED' AND submission_date BETWEEN ? AND ?", startDate, endDate).
		Count(&stats.RejectedCount)

	r.db.WithContext(ctx).
		Model(&models.WaiverApplication{}).
		Where("status IN ('SUBMITTED', 'UNDER_REVIEW', 'REVIEW_COMPLETED') AND submission_date BETWEEN ? AND ?", startDate, endDate).
		Count(&stats.PendingCount)

	// 计算批准率
	if stats.TotalRequests > 0 {
		stats.ApprovalRate = float64(stats.ApprovedCount) / float64(stats.TotalRequests) * 100
	}

	// 计算平均处理天数
	var avgDays sql.NullFloat64
	r.db.WithContext(ctx).
		Model(&models.WaiverApplication{}).
		Where("status IN ('APPROVED', 'REJECTED') AND submission_date BETWEEN ? AND ?", startDate, endDate).
		Select("AVG(DATEDIFF(COALESCE(updated_at, submission_date), submission_date))").
		Scan(&avgDays)

	if avgDays.Valid {
		stats.AvgProcessingDays = int64(avgDays.Float64)
	}

	return &stats, nil
}

func (r *enhancedConflictRepository) GetWaiverApprovalTrends(ctx context.Context, period string) ([]*WaiverApprovalTrend, error) {
	var trends []*WaiverApprovalTrend

	var startDate time.Time
	var groupBy string

	switch period {
	case "daily":
		startDate = time.Now().AddDate(0, 0, -30)
		groupBy = "DATE(submission_date)"
	case "weekly":
		startDate = time.Now().AddDate(0, 0, -12)
		groupBy = "YEARWEEK(submission_date)"
	case "monthly":
		startDate = time.Now().AddDate(0, -12, 0)
		groupBy = "DATE_FORMAT(submission_date, '%Y-%m')"
	default:
		return nil, fmt.Errorf("不支持的时间周期: %s", period)
	}

	query := `
		SELECT
			` + groupBy + ` as period,
			COUNT(*) as requests,
			SUM(CASE WHEN status = 'APPROVED' THEN 1 ELSE 0 END) as approvals,
			SUM(CASE WHEN status = 'REJECTED' THEN 1 ELSE 0 END) as rejections
		FROM waiver_applications
		WHERE submission_date >= ?
		GROUP BY ` + groupBy + `
		ORDER BY period
	`

	if err := r.db.WithContext(ctx).
		Raw(query, startDate).
		Scan(&trends).Error; err != nil {
		return nil, fmt.Errorf("获取豁免审批趋势数据失败: %w", err)
	}

	return trends, nil
}

func (r *enhancedConflictRepository) GetClientRiskStatistics(ctx context.Context, query *StatisticsQuery) (*ClientRiskStatistics, error) {
	var stats ClientRiskStatistics

	dbQuery := r.db.WithContext(ctx).Model(&models.ClientRiskProfile{})

	if !query.StartDate.IsZero() {
		dbQuery = dbQuery.Where("last_assessment_date >= ?", query.StartDate)
	}

	if !query.EndDate.IsZero() {
		dbQuery = dbQuery.Where("last_assessment_date <= ?", query.EndDate)
	}

	// 获取总客户数
	dbQuery.Count(&stats.TotalClients)

	// 获取各风险等级客户数
	dbQuery.Where("overall_risk = 'LOW'").Count(&stats.LowRiskClients)
	dbQuery.Where("overall_risk = 'MEDIUM'").Count(&stats.MediumRiskClients)
	dbQuery.Where("overall_risk = 'HIGH'").Count(&stats.HighRiskClients)
	dbQuery.Where("overall_risk = 'CRITICAL'").Count(&stats.CriticalRiskClients)

	// 获取需要监控的客户数
	dbQuery.Where("monitoring_required = ?", true).Count(&stats.MonitoringRequiredClients)

	return &stats, nil
}

func (r *enhancedConflictRepository) GetHighRiskClients(ctx context.Context, limit int) ([]*HighRiskClient, error) {
	var clients []*HighRiskClient

	if err := r.db.WithContext(ctx).
		Table("client_risk_profiles").
		Select(`
			client_profiles.id as client_id,
			client_profiles.client_number as client_name,
			client_risk_profiles.risk_score,
			client_risk_profiles.overall_risk as risk_level,
			client_risk_profiles.last_assessment_date as last_update
		`).
		Joins("JOIN client_profiles ON client_profiles.id = client_risk_profiles.client_id").
		Where("client_risk_profiles.overall_risk IN ('HIGH', 'CRITICAL')").
		Where("client_profiles.client_status IN ('ACTIVE', 'DORMANT')").
		Order("client_risk_profiles.risk_score DESC, client_risk_profiles.last_assessment_date DESC").
		Limit(limit).
		Scan(&clients).Error; err != nil {
		return nil, fmt.Errorf("获取高风险客户失败: %w", err)
	}

	return clients, nil
}

func (r *enhancedConflictRepository) GetProcessingEfficiencyStats(ctx context.Context, query *StatisticsQuery) (*ProcessingEfficiencyStats, error) {
	var stats ProcessingEfficiencyStats

	dbQuery := r.db.WithContext(ctx).Model(&models.ProfessionalConflictCheckRequest{})

	if !query.StartDate.IsZero() {
		dbQuery = dbQuery.Where("requested_date >= ?", query.StartDate)
	}

	if !query.EndDate.IsZero() {
		dbQuery = dbQuery.Where("requested_date <= ?", query.EndDate)
	}

	// 获取平均处理时间
	var avgTime sql.NullInt64
	dbQuery.Where("status = 'COMPLETED'").
		Select("AVG(TIMESTAMPDIFF(MINUTE, started_at, completed_at))").
		Scan(&avgTime)

	if avgTime.Valid {
		stats.AvgProcessingTime = avgTime.Int64
	}

	// 获取快速处理率（30分钟内完成）
	var fastCount, totalCount int64
	dbQuery.Where("status = 'COMPLETED'").Count(&totalCount)

	if totalCount > 0 {
		dbQuery.Where("status = 'COMPLETED' AND TIMESTAMPDIFF(MINUTE, started_at, completed_at) <= 30").
			Count(&fastCount)
		stats.FastProcessingRate = float64(fastCount) / float64(totalCount) * 100
	}

	// 获取按时完成率
	var onTimeCount int64
	if totalCount > 0 {
		dbQuery.Where("status = 'COMPLETED' AND completed_at <= required_by_date").
			Count(&onTimeCount)
		stats.OnTimeCompletionRate = float64(onTimeCount) / float64(totalCount) * 100
	}

	// 获取当前队列长度
	r.db.WithContext(ctx).
		Model(&models.ProfessionalConflictCheckRequest{}).
		Where("status IN ('PENDING', 'IN_PROGRESS')").
		Count(&stats.QueueLength)

	return &stats, nil
}

func (r *enhancedConflictRepository) GetSlaComplianceStats(ctx context.Context, query *StatisticsQuery) (*SlaComplianceStats, error) {
	var stats SlaComplianceStats

	dbQuery := r.db.WithContext(ctx).Model(&models.ProfessionalConflictCheckRequest{})

	if !query.StartDate.IsZero() {
		dbQuery = dbQuery.Where("requested_date >= ?", query.StartDate)
	}

	if !query.EndDate.IsZero() {
		dbQuery = dbQuery.Where("requested_date <= ?", query.EndDate)
	}

	// 获取总体SLA合规率
	var totalCount, breachedCount int64
	dbQuery.Where("status = 'COMPLETED'").Count(&totalCount)

	if totalCount > 0 {
		dbQuery.Where("status = 'COMPLETED' AND completed_at > required_by_date").
			Count(&breachedCount)
		stats.OverallComplianceRate = float64(totalCount-breachedCount) / float64(totalCount) * 100
		stats.BreachedRequests = breachedCount
		stats.TotalRequests = totalCount
	}

	// 获取各优先级的SLA合规率
	priorities := []string{"LOW", "MEDIUM", "HIGH", "URGENT", "CRITICAL"}
	stats.PriorityCompliance = make(map[string]float64)

	for _, priority := range priorities {
		var priorityTotal, priorityBreached int64
		dbQuery.Where("status = 'COMPLETED' AND priority = ?", priority).Count(&priorityTotal)

		if priorityTotal > 0 {
			dbQuery.Where("status = 'COMPLETED' AND priority = ? AND completed_at > required_by_date", priority).
				Count(&priorityBreached)
			stats.PriorityCompliance[priority] = float64(priorityTotal-priorityBreached) / float64(priorityTotal) * 100
		}
	}

	return &stats, nil
}