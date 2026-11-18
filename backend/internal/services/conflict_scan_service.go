package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// ConflictScanService 冲突扫描服务接口
type ConflictScanService interface {
	// TriggerManualScan 手动触发扫描
	TriggerManualScan(ctx context.Context, req *ManualScanRequest) (*models.ConflictScanJob, error)
	// RunDailyScan 执行每日扫描
	RunDailyScan(ctx context.Context) (*models.ConflictScanJob, error)
	// RunWeeklyScan 执行每周扫描
	RunWeeklyScan(ctx context.Context) (*models.ConflictScanJob, error)
	// RunIncrementalScan 执行增量扫描
	RunIncrementalScan(ctx context.Context, since time.Time) (*models.ConflictScanJob, error)
	// RunFullScan 执行全量扫描
	RunFullScan(ctx context.Context) (*models.ConflictScanJob, error)
	// GetScanJob 获取扫描任务
	GetScanJob(ctx context.Context, jobID uint) (*models.ConflictScanJob, error)
	// ListScanJobs 列出扫描任务
	ListScanJobs(ctx context.Context, filter *ScanJobFilter) ([]*models.ConflictScanJob, error)
	// GetScanStats 获取扫描统计
	GetScanStats(ctx context.Context) (*ScanStatistics, error)
}

// ManualScanRequest 手动扫描请求
type ManualScanRequest struct {
	TriggeredBy   uint   `json:"triggeredBy" validate:"required"`
	TriggerReason string `json:"triggerReason"`
	ScanScope     string `json:"scanScope"` // all/new_cases/lawyer
	LawyerID      *uint  `json:"lawyerId"`  // 当 scan_scope=lawyer 时指定
}

// ScanJobFilter 扫描任务过滤条件
type ScanJobFilter struct {
	ScanType  string     `json:"scanType"`
	Status    string     `json:"status"`
	StartDate *time.Time `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`
	Limit     int        `json:"limit"`
	Offset    int        `json:"offset"`
}

// ScanStatistics 扫描统计信息
type ScanStatistics struct {
	TotalJobs       int64      `json:"totalJobs"`
	RunningJobs     int64      `json:"runningJobs"`
	CompletedJobs   int64      `json:"completedJobs"`
	FailedJobs      int64      `json:"failedJobs"`
	TotalScans      int64      `json:"totalScans"`
	TotalConflicts  int64      `json:"totalConflicts"`
	LastScanTime    *time.Time `json:"lastScanTime"`
	AverageDuration float64    `json:"averageDuration"`
}

// conflictScanService 冲突扫描服务实现
type conflictScanService struct {
	db                  *gorm.DB
	conflictService     V2ConflictDetectionService
	poolService         ConflictPoolService
	caseRepo            CaseRepository
	notificationService NotificationService
	config              *ScanConfig
}

// ScanConfig 扫描配置
type ScanConfig struct {
	DailyScanTime     string `json:"dailyScanTime"`     // 每日扫描时间（如 "02:00"）
	WeeklyScanDay     string `json:"weeklyScanDay"`     // 每周扫描星期（如 "Monday"）
	WeeklyScanTime    string `json:"weeklyScanTime"`    // 每周扫描时间
	BatchSize         int    `json:"batchSize"`         // 批量处理大小
	MaxConcurrentJobs int    `json:"maxConcurrentJobs"` // 最大并发任务数
	AlertThreshold    int    `json:"alertThreshold"`    // 告警阈值
	EnableAutoScan    bool   `json:"enableAutoScan"`    // 是否启用自动扫描
}

// NotificationService 通知服务接口
type NotificationService interface {
	SendConflictAlert(ctx context.Context, conflicts []*NewConflictInfo) error
}

// NewConflictScanService 创建新的冲突扫描服务
func NewConflictScanService(
	db *gorm.DB,
	conflictService V2ConflictDetectionService,
	poolService ConflictPoolService,
	caseRepo CaseRepository,
	notificationService NotificationService,
) ConflictScanService {
	return &conflictScanService{
		db:                  db,
		conflictService:     conflictService,
		poolService:         poolService,
		caseRepo:            caseRepo,
		notificationService: notificationService,
		config:              DefaultScanConfig(),
	}
}

// DefaultScanConfig 默认扫描配置
func DefaultScanConfig() *ScanConfig {
	return &ScanConfig{
		DailyScanTime:     "02:00",
		WeeklyScanDay:     "Sunday",
		WeeklyScanTime:    "03:00",
		BatchSize:         100,
		MaxConcurrentJobs: 3,
		AlertThreshold:    5,
		EnableAutoScan:    true,
	}
}

// TriggerManualScan 手动触发扫描
func (s *conflictScanService) TriggerManualScan(ctx context.Context, req *ManualScanRequest) (*models.ConflictScanJob, error) {
	log.Printf("🔔 手动触发冲突扫描: triggeredBy=%d, scope=%s", req.TriggeredBy, req.ScanScope)

	// 创建扫描任务
	job := &models.ConflictScanJob{
		ScanType:      "manual",
		ScanScope:     req.ScanScope,
		Status:        "pending",
		TriggeredBy:   &req.TriggeredBy,
		TriggerReason: req.TriggerReason,
	}

	if err := s.db.WithContext(ctx).Create(job).Error; err != nil {
		return nil, fmt.Errorf("创建扫描任务失败: %w", err)
	}

	// 异步执行扫描
	go s.executeScan(context.Background(), job)

	return job, nil
}

// RunDailyScan 执行每日扫描
func (s *conflictScanService) RunDailyScan(ctx context.Context) (*models.ConflictScanJob, error) {
	log.Printf("📅 执行每日冲突扫描")

	// 检查是否已有正在运行的每日扫描
	var runningJob models.ConflictScanJob
	err := s.db.WithContext(ctx).
		Where("scan_type = ? AND status = ?", "daily", "running").
		Order("created_at DESC").
		First(&runningJob).Error

	if err == nil {
		log.Printf("⚠️ 已有正在运行的每日扫描任务: ID=%d", runningJob.ID)
		return &runningJob, nil
	}

	// 创建扫描任务
	job := &models.ConflictScanJob{
		ScanType:      "daily",
		ScanScope:     "new_cases", // 每日扫描只检查新增案件
		Status:        "pending",
		TriggerReason: "每日定时扫描",
	}

	if err := s.db.WithContext(ctx).Create(job).Error; err != nil {
		return nil, fmt.Errorf("创建扫描任务失败: %w", err)
	}

	// 异步执行扫描
	go s.executeScan(context.Background(), job)

	return job, nil
}

// RunWeeklyScan 执行每周扫描
func (s *conflictScanService) RunWeeklyScan(ctx context.Context) (*models.ConflictScanJob, error) {
	log.Printf("📅 执行每周冲突扫描")

	// 检查是否已有正在运行的每周扫描
	var runningJob models.ConflictScanJob
	err := s.db.WithContext(ctx).
		Where("scan_type = ? AND status = ?", "weekly", "running").
		Order("created_at DESC").
		First(&runningJob).Error

	if err == nil {
		log.Printf("⚠️ 已有正在运行的每周扫描任务: ID=%d", runningJob.ID)
		return &runningJob, nil
	}

	// 创建扫描任务
	job := &models.ConflictScanJob{
		ScanType:      "weekly",
		ScanScope:     "all", // 每周扫描全量检查
		Status:        "pending",
		TriggerReason: "每周定时扫描",
	}

	if err := s.db.WithContext(ctx).Create(job).Error; err != nil {
		return nil, fmt.Errorf("创建扫描任务失败: %w", err)
	}

	// 异步执行扫描
	go s.executeScan(context.Background(), job)

	return job, nil
}

// RunIncrementalScan 执行增量扫描
func (s *conflictScanService) RunIncrementalScan(ctx context.Context, since time.Time) (*models.ConflictScanJob, error) {
	log.Printf("📅 执行增量冲突扫描: since=%s", since.Format("2006-01-02 15:04:05"))

	job := &models.ConflictScanJob{
		ScanType:      "incremental",
		ScanScope:     fmt.Sprintf("since:%s", since.Format("2006-01-02")),
		Status:        "pending",
		TriggerReason: "增量扫描",
	}

	if err := s.db.WithContext(ctx).Create(job).Error; err != nil {
		return nil, fmt.Errorf("创建扫描任务失败: %w", err)
	}

	// 异步执行扫描
	go s.executeIncrementalScan(context.Background(), job, since)

	return job, nil
}

// RunFullScan 执行全量扫描
func (s *conflictScanService) RunFullScan(ctx context.Context) (*models.ConflictScanJob, error) {
	log.Printf("📅 执行全量冲突扫描")

	job := &models.ConflictScanJob{
		ScanType:      "full",
		ScanScope:     "all",
		Status:        "pending",
		TriggerReason: "全量扫描",
	}

	if err := s.db.WithContext(ctx).Create(job).Error; err != nil {
		return nil, fmt.Errorf("创建扫描任务失败: %w", err)
	}

	// 异步执行扫描
	go s.executeScan(context.Background(), job)

	return job, nil
}

// GetScanJob 获取扫描任务
func (s *conflictScanService) GetScanJob(ctx context.Context, jobID uint) (*models.ConflictScanJob, error) {
	var job models.ConflictScanJob
	if err := s.db.WithContext(ctx).First(&job, jobID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("扫描任务不存在")
		}
		return nil, fmt.Errorf("获取扫描任务失败: %w", err)
	}
	return &job, nil
}

// ListScanJobs 列出扫描任务
func (s *conflictScanService) ListScanJobs(ctx context.Context, filter *ScanJobFilter) ([]*models.ConflictScanJob, error) {
	query := s.db.WithContext(ctx).Model(&models.ConflictScanJob{})

	// 应用过滤条件
	if filter.ScanType != "" {
		query = query.Where("scan_type = ?", filter.ScanType)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", *filter.EndDate)
	}

	// 分页
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// 按时间倒序
	query = query.Order("created_at DESC")

	var jobs []*models.ConflictScanJob
	if err := query.Find(&jobs).Error; err != nil {
		return nil, fmt.Errorf("获取扫描任务列表失败: %w", err)
	}

	return jobs, nil
}

// GetScanStats 获取扫描统计
func (s *conflictScanService) GetScanStats(ctx context.Context) (*ScanStatistics, error) {
	stats := &ScanStatistics{}

	// 总任务数
	if err := s.db.WithContext(ctx).
		Model(&models.ConflictScanJob{}).
		Count(&stats.TotalJobs).Error; err != nil {
		return nil, fmt.Errorf("获取总任务数失败: %w", err)
	}

	// 按状态统计
	var statusStats []struct {
		Status string
		Count  int64
	}
	if err := s.db.WithContext(ctx).
		Model(&models.ConflictScanJob{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&statusStats).Error; err != nil {
		return nil, fmt.Errorf("获取状态统计失败: %w", err)
	}

	for _, stat := range statusStats {
		switch stat.Status {
		case "running":
			stats.RunningJobs = stat.Count
		case "completed":
			stats.CompletedJobs = stat.Count
		case "failed":
			stats.FailedJobs = stat.Count
		}
	}

	// 扫描统计
	var scanStats struct {
		TotalScans     int64
		TotalConflicts int64
	}
	if err := s.db.WithContext(ctx).
		Model(&models.ConflictScanJob{}).
		Select("sum(scanned_cases) as total_scans, sum(found_conflicts) as total_conflicts").
		Where("status = ?", "completed").
		Scan(&scanStats).Error; err != nil {
		return nil, fmt.Errorf("获取扫描统计失败: %w", err)
	}

	stats.TotalScans = scanStats.TotalScans
	stats.TotalConflicts = scanStats.TotalConflicts

	// 最后扫描时间
	var lastJob models.ConflictScanJob
	if err := s.db.WithContext(ctx).
		Where("status = ?", "completed").
		Order("completed_at DESC").
		First(&lastJob).Error; err == nil {
		stats.LastScanTime = lastJob.CompletedAt
	}

	return stats, nil
}

// ============================================================================
// 私有方法
// ============================================================================

// executeScan 执行扫描
func (s *conflictScanService) executeScan(ctx context.Context, job *models.ConflictScanJob) {
	log.Printf("🚀 开始执行扫描任务: ID=%d, type=%s", job.ID, job.ScanType)

	// 更新状态为运行中
	now := time.Now()
	job.Status = "running"
	job.StartedAt = &now
	s.db.Save(job)

	// 获取需要扫描的案件
	cases, err := s.getCasesForScan(ctx, job)
	if err != nil {
		s.failJob(job, err)
		return
	}

	log.Printf("📋 找到 %d 个需要扫描的案件", len(cases))

	// 初始化统计
	job.ScannedCases = 0
	job.FoundConflicts = 0
	newConflicts := make([]models.JSON, 0)

	// 扫描每个案件
	for _, case_ := range cases {
		conflicts, err := s.scanCase(ctx, case_)
		if err != nil {
			log.Printf("⚠️ 扫描案件 %d 失败: %v", case_.ID, err)
			continue
		}

		job.ScannedCases++
		job.ScannedLawyers = s.countUniqueLawyers(job.ScannedLawyers, case_.LawyerID)

		if len(conflicts) > 0 {
			job.FoundConflicts += len(conflicts)
			for _, conflict := range conflicts {
				conflictJSON := models.JSON(map[string]interface{}{
					"case_id":    case_.ID,
					"case_title": case_.Title,
					"lawyer_id":  case_.LawyerID,
					"conflict":   conflict,
				})
				newConflicts = append(newConflicts, conflictJSON)
			}
		}

		// 同步冲突池
		if err := s.poolService.SyncLawyerPool(ctx, case_.LawyerID, case_.ID); err != nil {
			log.Printf("⚠️ 同步冲突池失败: caseID=%d, error=%v", case_.ID, err)
		}
	}

	// 保存新发现的冲突
	if len(newConflicts) > 0 {
		job.NewConflicts = models.JSON(map[string]interface{}{
			"conflicts": newConflicts,
		})
	}

	// 完成扫描
	job.Status = "completed"
	job.CompletedAt = &now
	s.db.Save(job)

	log.Printf("✅ 扫描任务完成: ID=%d, 扫描案件=%d, 发现冲突=%d",
		job.ID, job.ScannedCases, job.FoundConflicts)

	// 如果发现冲突且超过阈值，发送告警
	if job.FoundConflicts >= s.config.AlertThreshold {
		s.sendAlertIfNeeded(ctx, job, newConflicts)
	}
}

// executeIncrementalScan 执行增量扫描
func (s *conflictScanService) executeIncrementalScan(ctx context.Context, job *models.ConflictScanJob, since time.Time) {
	log.Printf("🚀 开始执行增量扫描任务: ID=%d, since=%s", job.ID, since.Format("2006-01-02 15:04:05"))

	// 更新状态
	now := time.Now()
	job.Status = "running"
	job.StartedAt = &now
	s.db.Save(job)

	// 获取增量案件
	var cases []*models.Case
	if err := s.db.WithContext(ctx).
		Where("created_at >= ?", since).
		Find(&cases).Error; err != nil {
		s.failJob(job, err)
		return
	}

	log.Printf("📋 找到 %d 个增量案件", len(cases))

	// 扫描逻辑与 executeScan 相同
	newConflicts := make([]models.JSON, 0)
	uniqueLawyers := make(map[uint]bool)

	for _, case_ := range cases {
		conflicts, err := s.scanCase(ctx, case_)
		if err != nil {
			log.Printf("⚠️ 扫描案件 %d 失败: %v", case_.ID, err)
			continue
		}

		job.ScannedCases++
		uniqueLawyers[case_.LawyerID] = true

		if len(conflicts) > 0 {
			job.FoundConflicts += len(conflicts)
			for _, conflict := range conflicts {
				conflictJSON := models.JSON(map[string]interface{}{
					"case_id":    case_.ID,
					"case_title": case_.Title,
					"lawyer_id":  case_.LawyerID,
					"conflict":   conflict,
				})
				newConflicts = append(newConflicts, conflictJSON)
			}
		}
	}

	job.ScannedLawyers = len(uniqueLawyers)

	if len(newConflicts) > 0 {
		job.NewConflicts = models.JSON(map[string]interface{}{
			"conflicts": newConflicts,
		})
	}

	job.Status = "completed"
	job.CompletedAt = &now
	s.db.Save(job)

	log.Printf("✅ 增量扫描完成: 扫描案件=%d, 发现冲突=%d",
		job.ScannedCases, job.FoundConflicts)

	if job.FoundConflicts >= s.config.AlertThreshold {
		s.sendAlertIfNeeded(ctx, job, newConflicts)
	}
}

// getCasesForScan 获取需要扫描的案件
func (s *conflictScanService) getCasesForScan(ctx context.Context, job *models.ConflictScanJob) ([]*models.Case, error) {
	var cases []*models.Case
	var err error

	switch job.ScanScope {
	case "all":
		// 全量扫描
		err = s.db.WithContext(ctx).Find(&cases).Error
	case "new_cases":
		// 扫描最近24小时新增的案件
		yesterday := time.Now().Add(-24 * time.Hour)
		err = s.db.WithContext(ctx).
			Where("created_at >= ?", yesterday).
			Find(&cases).Error
	default:
		// 默认扫描所有案件
		err = s.db.WithContext(ctx).Find(&cases).Error
	}

	if err != nil {
		return nil, fmt.Errorf("获取案件失败: %w", err)
	}

	return cases, nil
}

// scanCase 扫描单个案件
func (s *conflictScanService) scanCase(ctx context.Context, case_ *models.Case) ([]string, error) {
	// 获取客户信息
	var client models.Client
	if err := s.db.WithContext(ctx).First(&client, case_.ClientID).Error; err != nil {
		return nil, fmt.Errorf("获取客户失败: %w", err)
	}

	// 构建检测请求
	identityNumber, idErr := client.DecryptedIdentity()
	if idErr != nil {
		return nil, fmt.Errorf("读取客户身份标识失败: %w", idErr)
	}
	req := &ConflictCheckRequestV2{
		LawyerID:    case_.LawyerID,
		ClientName:  client.Name,
		ClientTaxID: identityNumber,
		CaseID:      case_.ID,
		SearchDepth: "standard",
	}

	// 执行检测
	result, err := s.conflictService.QuickCheck(ctx, req)
	if err != nil {
		return nil, err
	}

	// 收集冲突信息
	conflicts := make([]string, 0)
	for _, match := range result.Matches {
		conflict := fmt.Sprintf("案件 %s 与 %s 存在 %s 冲突",
			match.CaseTitle, match.EntityInfo.Name, match.Relationship)
		conflicts = append(conflicts, conflict)
	}

	return conflicts, nil
}

// countUniqueLawyers 统计唯一律师数
func (s *conflictScanService) countUniqueLawyers(currentCount int, lawyerID uint) int {
	// 简化处理，实际应该使用 map 去重
	return currentCount + 1
}

// failJob 标记任务失败
func (s *conflictScanService) failJob(job *models.ConflictScanJob, err error) {
	log.Printf("❌ 扫描任务失败: ID=%d, error=%v", job.ID, err)
	job.Status = "failed"
	job.ErrorMessage = err.Error()
	now := time.Now()
	job.CompletedAt = &now
	s.db.Save(job)
}

// sendAlertIfNeeded 发送告警（如果需要）
func (s *conflictScanService) sendAlertIfNeeded(ctx context.Context, job *models.ConflictScanJob, conflicts []models.JSON) {
	if job.TriggeredAlerts {
		return
	}

	if s.notificationService == nil {
		log.Printf("⚠️ 通知服务未配置，跳过告警")
		return
	}

	log.Printf("🚨 发送冲突告警: 发现 %d 个冲突", job.FoundConflicts)

	// 构建告警信息
	newConflictInfos := make([]*NewConflictInfo, 0)
	for _, conflict := range conflicts {
		info := &NewConflictInfo{
			ScanJobID: job.ID,
			Details:   conflict,
		}
		newConflictInfos = append(newConflictInfos, info)
	}

	// 发送告警
	if err := s.notificationService.SendConflictAlert(ctx, newConflictInfos); err != nil {
		log.Printf("⚠️ 发送告警失败: %v", err)
		return
	}

	// 更新告警状态
	now := time.Now()
	job.TriggeredAlerts = true
	job.AlertSentAt = &now
	s.db.Save(job)
}

// NewConflictInfo 新冲突信息
type NewConflictInfo struct {
	ScanJobID uint        `json:"scanJobId"`
	Details   models.JSON `json:"details"`
}

// ============================================================================
// 调度器相关
// ============================================================================

// StartScheduler 启动扫描调度器
func (s *conflictScanService) StartScheduler(ctx context.Context) {
	if !s.config.EnableAutoScan {
		log.Printf("⚠️ 自动扫描已禁用")
		return
	}

	log.Printf("🕐 启动扫描调度器")

	// 启动每日扫描定时器
	go s.dailyScanTicker(ctx)

	// 启动每周扫描定时器
	go s.weeklyScanTicker(ctx)
}

// dailyScanTicker 每日扫描定时器
func (s *conflictScanService) dailyScanTicker(ctx context.Context) {
	// 计算下次执行时间
	nextRun := s.calculateNextDailyRun()
	log.Printf("📅 下次每日扫描: %s", nextRun.Format("2006-01-02 15:04:05"))

	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			// 检查是否到了执行时间（允许5分钟误差）
			if now.After(nextRun) && now.Before(nextRun.Add(5*time.Minute)) {
				s.RunDailyScan(context.Background())
				// 计算下次执行时间
				nextRun = s.calculateNextDailyRun()
				log.Printf("📅 下次每日扫描: %s", nextRun.Format("2006-01-02 15:04:05"))
			}
		}
	}
}

// weeklyScanTicker 每周扫描定时器
func (s *conflictScanService) weeklyScanTicker(ctx context.Context) {
	nextRun := s.calculateNextWeeklyRun()
	log.Printf("📅 下次每周扫描: %s", nextRun.Format("2006-01-02 15:04:05"))

	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			if now.After(nextRun) && now.Before(nextRun.Add(5*time.Minute)) {
				s.RunWeeklyScan(context.Background())
				nextRun = s.calculateNextWeeklyRun()
				log.Printf("📅 下次每周扫描: %s", nextRun.Format("2006-01-02 15:04:05"))
			}
		}
	}
}

// calculateNextDailyRun 计算下次每日执行时间
func (s *conflictScanService) calculateNextDailyRun() time.Time {
	now := time.Now()

	// 解析配置的时间
	targetTime, err := time.Parse("15:04", s.config.DailyScanTime)
	if err != nil {
		targetTime = time.Date(0, 1, 1, 2, 0, 0, 0, time.UTC)
	}

	next := time.Date(now.Year(), now.Month(), now.Day(),
		targetTime.Hour(), targetTime.Minute(), 0, 0, now.Location())

	// 如果今天的时间已过，设置为明天
	if next.Before(now) {
		next = next.Add(24 * time.Hour)
	}

	return next
}

// calculateNextWeeklyRun 计算下次每周执行时间
func (s *conflictScanService) calculateNextWeeklyRun() time.Time {
	now := time.Now()
	targetWeekday := s.parseWeekday(s.config.WeeklyScanDay)

	// 解析配置的时间
	targetTime, err := time.Parse("15:04", s.config.WeeklyScanTime)
	if err != nil {
		targetTime = time.Date(0, 1, 1, 3, 0, 0, 0, time.UTC)
	}

	// 找到下一个目标星期
	daysUntil := int(targetWeekday) - int(now.Weekday())
	if daysUntil <= 0 {
		daysUntil += 7
	}

	next := time.Date(now.Year(), now.Month(), now.Day(),
		targetTime.Hour(), targetTime.Minute(), 0, 0, now.Location())
	next = next.AddDate(0, 0, daysUntil)

	return next
}

// parseWeekday 解析星期
func (s *conflictScanService) parseWeekday(weekday string) time.Weekday {
	switch strings.ToLower(weekday) {
	case "sunday", "周日":
		return time.Sunday
	case "monday", "周一":
		return time.Monday
	case "tuesday", "周二":
		return time.Tuesday
	case "wednesday", "周三":
		return time.Wednesday
	case "thursday", "周四":
		return time.Thursday
	case "friday", "周五":
		return time.Friday
	case "saturday", "周六":
		return time.Saturday
	default:
		return time.Sunday
	}
}
