package repositories

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"law-oa-go/internal/models"
)

// BasicConflictRepository 基础冲突检测数据仓库接口
type BasicConflictRepository interface {
	// 保存冲突检测记录
	SaveCheckRecord(ctx context.Context, record *models.ConflictCheckRecord) error
	// 获取单条冲突检测记录
	GetCheckRecord(ctx context.Context, checkID string) (*models.ConflictCheckRecord, error)
	// 更新冲突检测记录
	UpdateCheckRecord(ctx context.Context, record *models.ConflictCheckRecord) error
	// 获取冲突检测历史
	GetCheckHistory(ctx context.Context, clientID string, limit int) ([]*models.ConflictCheckRecord, error)
	// 获取冲突案例
	GetConflictCases(ctx context.Context, params *ConflictSearchParams) ([]*models.ConflictCase, error)
	// 获取潜在冲突案例（从主案件表）
	GetPotentialConflicts(ctx context.Context, clientID string, lawyerID uint, otherParties []string, since time.Time) ([]*models.ConflictCase, error)
	// 获取客户关系
	GetClientRelations(ctx context.Context, clientID string) ([]*models.ClientRelation, error)
	// 保存冲突案例
	SaveConflictCases(ctx context.Context, cases []*models.ConflictCase) error
	// 获取冲突规则
	GetConflictRules(ctx context.Context, activeOnly bool) ([]*models.ConflictRule, error)
	// 保存冲突规则
	SaveConflictRule(ctx context.Context, rule *models.ConflictRule) error
	// 更新冲突规则
	UpdateConflictRule(ctx context.Context, rule *models.ConflictRule) error
	// 获取MCP标准
	GetMCPStandards(ctx context.Context, activeOnly bool) (*models.MCPStandards, error)
	// 保存MCP标准
	SaveMCPStandards(ctx context.Context, standards *models.MCPStandards) error
	// 获取统计信息
	GetConflictStats(ctx context.Context, clientID string) (*ConflictStats, error)
}

// ConflictSearchParams 冲突案例搜索参数
type ConflictSearchParams struct {
	ClientID  string    `json:"clientId"`
	CaseType  string    `json:"caseType"`
	RiskLevel string    `json:"riskLevel"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
	Page      int       `json:"page"`
	PageSize  int       `json:"pageSize"`
}

// ConflictStats 冲突检测统计
type ConflictStats struct {
	TotalChecks     int64     `json:"totalChecks"`
	ConflictChecks  int64     `json:"conflictChecks"`
	HighRiskChecks  int64     `json:"highRiskChecks"`
	AverageDuration float64   `json:"averageDuration"`
	LastCheckTime   time.Time `json:"lastCheckTime"`
}

// conflictRepository 冲突检测数据仓库实现
type conflictRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewConflictRepository 创建新的冲突检测数据仓库
func NewConflictRepository(db *gorm.DB, redis *redis.Client) BasicConflictRepository {
	return &conflictRepository{
		db:    db,
		redis: redis,
	}
}

// SaveCheckRecord 保存冲突检测记录
func (r *conflictRepository) SaveCheckRecord(ctx context.Context, record *models.ConflictCheckRecord) error {
	record.UpdatedAt = time.Now()
	if record.CreatedAt.IsZero() {
		record.CreatedAt = record.UpdatedAt
	}

	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "check_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"client_id",
			"client_name",
			"case_name",
			"case_type",
			"check_status",
			"has_conflict",
			"risk_level",
			"search_parameters",
			"check_result",
			"user_id",
			"duration",
			"check_time",
			"updated_at",
		}),
	}).Create(record).Error
	if err != nil {
		return fmt.Errorf("保存冲突检测记录失败: %w", err)
	}

	// 缓存最近的结果
	if r.redis != nil {
		cacheKey := fmt.Sprintf("conflict:last_check:%s", record.ClientID)
		r.redis.Set(ctx, cacheKey, record, 24*time.Hour)
	}

	return nil
}

// GetCheckRecord 获取单条冲突检测记录
func (r *conflictRepository) GetCheckRecord(ctx context.Context, checkID string) (*models.ConflictCheckRecord, error) {
	var record models.ConflictCheckRecord
	err := r.db.WithContext(ctx).Where("check_id = ?", checkID).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("获取冲突检测记录失败: %w", err)
	}
	return &record, nil
}

// UpdateCheckRecord 更新冲突检测记录
func (r *conflictRepository) UpdateCheckRecord(ctx context.Context, record *models.ConflictCheckRecord) error {
	if err := r.db.WithContext(ctx).Save(record).Error; err != nil {
		return fmt.Errorf("更新冲突检测记录失败: %w", err)
	}
	return nil
}

// GetCheckHistory 获取冲突检测历史
func (r *conflictRepository) GetCheckHistory(ctx context.Context, clientID string, limit int) ([]*models.ConflictCheckRecord, error) {
	var records []*models.ConflictCheckRecord

	query := r.db.WithContext(ctx).Where("client_id = ?", clientID).
		Order("check_time DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("获取冲突检测历史失败: %w", err)
	}

	return records, nil
}

// GetConflictCases 获取冲突案例
func (r *conflictRepository) GetConflictCases(ctx context.Context, params *ConflictSearchParams) ([]*models.ConflictCase, error) {
	var cases []*models.ConflictCase

	query := r.db.WithContext(ctx).Model(&models.ConflictCase{})

	if params.ClientID != "" {
		query = query.Where("client_id = ?", params.ClientID)
	}
	if params.CaseType != "" {
		query = query.Where("case_type = ?", params.CaseType)
	}
	if params.RiskLevel != "" {
		query = query.Where("risk_level = ?", params.RiskLevel)
	}
	if !params.StartDate.IsZero() {
		query = query.Where("created_at >= ?", params.StartDate)
	}
	if !params.EndDate.IsZero() {
		query = query.Where("created_at <= ?", params.EndDate)
	}

	// 分页
	if params.Page > 0 && params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		query = query.Offset(offset).Limit(params.PageSize)
	}

	// 按时间倒序
	query = query.Order("created_at DESC")

	if err := query.Find(&cases).Error; err != nil {
		return nil, fmt.Errorf("获取冲突案例失败: %w", err)
	}

	return cases, nil
}

// GetPotentialConflicts 获取潜在冲突案例（从主案件表）
func (r *conflictRepository) GetPotentialConflicts(ctx context.Context, clientID string, lawyerID uint, otherParties []string, since time.Time) ([]*models.ConflictCase, error) {
	var conflictCases []*models.ConflictCase

	log.Printf("🔍 查询潜在冲突: clientID=%s, lawyerID=%d", clientID, lawyerID)

	// 🔧 修复：需要将字符串clientID转换为uint，同时保持原逻辑
	var clientIDUint uint
	if _, err := fmt.Sscanf(clientID, "%d", &clientIDUint); err != nil {
		// 如果转换失败，说明clientID格式不对，记录错误并返回空结果
		log.Printf("⚠️ 客户ID格式错误: %s, 无法转换为uint", clientID)
		return conflictCases, nil
	}

	sinceFilter := ""
	args := []interface{}{lawyerID, clientIDUint}
	if !since.IsZero() {
		sinceFilter = "AND c.created_at >= ?"
		args = append(args, since)
	}

	// 查询主案件表，查找同一律师代理的其他案件
	query := fmt.Sprintf(`
		SELECT
			c.id as case_id,
			c.case_number,
			c.title as case_name,
			c.case_type,
			c.description,
			c.client_id,
			cl.name as client_name,
			cl.type as client_type,
			u.name as lawyer_name,
			c.created_at,
			c.lawyer_id
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		JOIN users u ON c.lawyer_id = u.id
		WHERE c.lawyer_id = ? AND c.client_id != ?
		AND c.deleted_at IS NULL
		%s
		ORDER BY c.created_at DESC
		LIMIT 50
	`, sinceFilter)

	rows, err := r.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, fmt.Errorf("查询潜在冲突案例失败: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var caseModel models.Case
		var clientName, clientType, lawyerName string
		var foundLawyerID uint

		err := rows.Scan(
			&caseModel.ID,
			&caseModel.CaseNumber,
			&caseModel.Title,
			&caseModel.CaseType,
			&caseModel.Description,
			&caseModel.ClientID,
			&clientName,
			&clientType,
			&lawyerName,
			&caseModel.CreatedAt,
			&foundLawyerID,
		)
		if err != nil {
			log.Printf("⚠️ 扫描案件数据失败: %v", err)
			continue
		}

		log.Printf("📋 发现潜在冲突案件: ID=%d, 标题=%s, 客户=%s, 律师=%s",
			caseModel.ID, caseModel.Title, clientName, lawyerName)

		// 创建冲突案例对象
		conflictCase := &models.ConflictCase{
			ID:           fmt.Sprintf("case_%d", caseModel.ID),
			CaseID:       fmt.Sprintf("%d", caseModel.ID),
			CaseName:     caseModel.Title,
			CaseNo:       caseModel.CaseNumber,
			CaseType:     caseModel.CaseType,
			Description:  fmt.Sprintf("律师 %s 同时代理了案件 '%s'，存在潜在利益冲突", lawyerName, caseModel.Title),
			ClientID:     fmt.Sprintf("%d", caseModel.ClientID),
			RiskLevel:    "MEDIUM", // 默认中等风险
			ConflictType: "代理冲突",
			CaseStatus:   "active",
			CreatedAt:    caseModel.CreatedAt,
		}

		conflictCases = append(conflictCases, conflictCase)
		log.Printf("✅ 创建冲突案例: %s", conflictCase.ID)
	}

	// 如果提供了其他当事人信息，也查询相关的案件
	if len(otherParties) > 0 {
		for _, party := range otherParties {
			// 查询包含对方当事人名称的案件描述
			partySinceFilter := ""
			partyArgs := []interface{}{"%" + party + "%", "%" + party + "%"}
			if !since.IsZero() {
				partySinceFilter = "AND c.created_at >= ?"
				partyArgs = append([]interface{}{since}, partyArgs...)
			}

			partyQuery := fmt.Sprintf(`
				SELECT
					c.id as case_id,
					c.case_number,
					c.title as case_name,
					c.case_type,
					c.description,
					c.client_id,
					cl.name as client_name,
					cl.type as client_type,
					u.name as lawyer_name,
					c.created_at,
					c.lawyer_id
				FROM cases c
				JOIN clients cl ON c.client_id = cl.id
				JOIN users u ON c.lawyer_id = u.id
				WHERE c.deleted_at IS NULL
				%s
				AND (c.title ILIKE ? OR c.description ILIKE ?)
				ORDER BY c.created_at DESC
				LIMIT 20
			`, partySinceFilter)

			partyRows, err := r.db.WithContext(ctx).Raw(partyQuery, partyArgs...).Rows()
			if err != nil {
				continue
			}

			for partyRows.Next() {
				var caseModel models.Case
				var clientName, clientType, lawyerName string
				var foundLawyerID uint

				err := partyRows.Scan(
					&caseModel.ID,
					&caseModel.CaseNumber,
					&caseModel.Title,
					&caseModel.CaseType,
					&caseModel.Description,
					&caseModel.ClientID,
					&clientName,
					&clientType,
					&lawyerName,
					&caseModel.CreatedAt,
					&foundLawyerID,
				)
				if err != nil {
					continue
				}

				// 如果是同一个律师的案件，跳过（已经在上面查过了）
				if foundLawyerID == lawyerID {
					continue
				}

				// 创建冲突案例对象
				conflictCase := &models.ConflictCase{
					ID:           fmt.Sprintf("case_%d", caseModel.ID),
					CaseID:       fmt.Sprintf("%d", caseModel.ID),
					CaseName:     caseModel.Title,
					CaseNo:       caseModel.CaseNumber,
					CaseType:     caseModel.CaseType,
					Description:  caseModel.Description,
					ClientID:     fmt.Sprintf("%d", caseModel.ClientID),
					RiskLevel:    "HIGH", // 对方当事人冲突，高风险
					ConflictType: "对方当事人冲突",
					CaseStatus:   "active",
					CreatedAt:    caseModel.CreatedAt,
				}

				conflictCases = append(conflictCases, conflictCase)
			}
			partyRows.Close()
		}
	}

	log.Printf("🎯 冲突检测完成: 找到 %d 个潜在冲突案例", len(conflictCases))
	return conflictCases, nil
}

// GetClientRelations 获取客户关系
func (r *conflictRepository) GetClientRelations(ctx context.Context, clientID string) ([]*models.ClientRelation, error) {
	var relations []*models.ClientRelation

	// 缓存检查
	if r.redis != nil {
		cacheKey := fmt.Sprintf("conflict:client_relations:%s", clientID)
		if cached, err := r.redis.Get(ctx, cacheKey).Result(); err == nil {
			// 这里应该从缓存反序列化，简化处理
			_ = cached
		}
	}

	if err := r.db.WithContext(ctx).
		Where("client_id = ? AND active = ?", clientID, true).
		Find(&relations).Error; err != nil {
		return nil, fmt.Errorf("获取客户关系失败: %w", err)
	}

	// 缓存结果
	if r.redis != nil && len(relations) > 0 {
		cacheKey := fmt.Sprintf("conflict:client_relations:%s", clientID)
		r.redis.Set(ctx, cacheKey, relations, 2*time.Hour)
	}

	return relations, nil
}

// SaveConflictCases 保存冲突案例
func (r *conflictRepository) SaveConflictCases(ctx context.Context, cases []*models.ConflictCase) error {
	if len(cases) == 0 {
		return nil
	}

	// 批量插入
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).CreateInBatches(cases, 100).Error; err != nil {
		return fmt.Errorf("批量保存冲突案例失败: %w", err)
	}

	return nil
}

// GetConflictRules 获取冲突规则
func (r *conflictRepository) GetConflictRules(ctx context.Context, activeOnly bool) ([]*models.ConflictRule, error) {
	var rules []*models.ConflictRule

	query := r.db.WithContext(ctx).Model(&models.ConflictRule{})
	if activeOnly {
		query = query.Where("active = ?", true)
	}

	// 按优先级排序
	query = query.Order("priority DESC")

	if err := query.Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("获取冲突规则失败: %w", err)
	}

	return rules, nil
}

// SaveConflictRule 保存冲突规则
func (r *conflictRepository) SaveConflictRule(ctx context.Context, rule *models.ConflictRule) error {
	// 验证规则
	if err := rule.Validate(); err != nil {
		return fmt.Errorf("规则验证失败: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(rule).Error; err != nil {
		return fmt.Errorf("保存冲突规则失败: %w", err)
	}

	// 清除缓存
	if r.redis != nil {
		r.redis.Del(ctx, "conflict:rules:active")
	}

	return nil
}

// UpdateConflictRule 更新冲突规则
func (r *conflictRepository) UpdateConflictRule(ctx context.Context, rule *models.ConflictRule) error {
	// 验证规则
	if err := rule.Validate(); err != nil {
		return fmt.Errorf("规则验证失败: %w", err)
	}

	if err := r.db.WithContext(ctx).Save(rule).Error; err != nil {
		return fmt.Errorf("更新冲突规则失败: %w", err)
	}

	// 清除缓存
	if r.redis != nil {
		r.redis.Del(ctx, "conflict:rules:active")
	}

	return nil
}

// GetMCPStandards 获取MCP标准
func (r *conflictRepository) GetMCPStandards(ctx context.Context, activeOnly bool) (*models.MCPStandards, error) {
	var standards models.MCPStandards

	query := r.db.WithContext(ctx).Model(&models.MCPStandards{})
	if activeOnly {
		query = query.Where("active = ?", true)
	}

	if err := query.Order("last_updated DESC").First(&standards).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMCPStandardsNotFound
		}
		return nil, fmt.Errorf("获取MCP标准失败: %w", err)
	}

	return &standards, nil
}

// SaveMCPStandards 保存MCP标准
func (r *conflictRepository) SaveMCPStandards(ctx context.Context, standards *models.MCPStandards) error {
	if err := r.db.WithContext(ctx).Save(standards).Error; err != nil {
		return fmt.Errorf("保存MCP标准失败: %w", err)
	}

	// 清除缓存
	if r.redis != nil {
		r.redis.Del(ctx, "conflict:mcp_standards:active")
	}

	return nil
}

// GetConflictStats 获取统计信息
func (r *conflictRepository) GetConflictStats(ctx context.Context, clientID string) (*ConflictStats, error) {
	stats := &ConflictStats{}

	// 基础查询
	query := r.db.WithContext(ctx).Model(&models.ConflictCheckRecord{})
	if clientID != "" {
		query = query.Where("client_id = ?", clientID)
	}

	// 总检查次数
	if err := query.Count(&stats.TotalChecks).Error; err != nil {
		return nil, fmt.Errorf("获取总检查次数失败: %w", err)
	}

	// 有冲突的检查次数
	if err := query.Where("has_conflict = ?", true).Count(&stats.ConflictChecks).Error; err != nil {
		return nil, fmt.Errorf("获取冲突检查次数失败: %w", err)
	}

	// 高风险检查次数
	if err := query.Where("risk_level IN ?", []string{"HIGH", "CRITICAL"}).Count(&stats.HighRiskChecks).Error; err != nil {
		return nil, fmt.Errorf("获取高风险检查次数失败: %w", err)
	}

	// 平均持续时间
	var avgDuration float64
	if err := query.Select("AVG(duration)").Scan(&avgDuration).Error; err != nil {
		return nil, fmt.Errorf("获取平均持续时间失败: %w", err)
	}
	stats.AverageDuration = avgDuration

	// 最后检查时间
	var lastCheck time.Time
	if err := query.Select("MAX(check_time)").Scan(&lastCheck).Error; err != nil {
		return nil, fmt.Errorf("获取最后检查时间失败: %w", err)
	}
	stats.LastCheckTime = lastCheck

	return stats, nil
}

// 自定义错误
var (
	ErrMCPStandardsNotFound = errors.New("MCP标准未找到")
)
