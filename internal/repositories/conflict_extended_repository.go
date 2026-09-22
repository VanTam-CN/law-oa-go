package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// ConflictExtendedRepository 扩展的冲突检测仓储实现
type ConflictExtendedRepository struct {
	db *sql.DB
}

// NewConflictExtendedRepository 创建扩展冲突检测仓储
func NewConflictExtendedRepository(db interface{}) *ConflictExtendedRepository {
	// 尝试转换为 *sql.DB 或 *gorm.DB
	switch v := db.(type) {
	case *sql.DB:
		return &ConflictExtendedRepository{db: v}
	case *gorm.DB:
		// 如果是gorm.DB，转换为*sql.DB
		sqlDB, err := v.DB()
		if err != nil {
			log.Printf("无法转换gorm.DB到sql.DB: %v", err)
			return nil
		}
		return &ConflictExtendedRepository{db: sqlDB}
	default:
		log.Printf("不支持的数据库类型: %T", db)
		return nil
	}
}

// CreateOrUpdateIndustry 创建或更新行业分类
func (r *ConflictExtendedRepository) CreateOrUpdateIndustry(ctx context.Context, industry *models.IndustryClassification) error {
	query := `
		INSERT INTO industry_classifications (code, name, parent_id, level, description, keywords, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		ON CONFLICT (code) DO UPDATE SET
			name = EXCLUDED.name,
			parent_id = EXCLUDED.parent_id,
			level = EXCLUDED.level,
			description = EXCLUDED.description,
			keywords = EXCLUDED.keywords,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		industry.Code, industry.Name, industry.ParentID, industry.Level,
		industry.Description, industry.Keywords,
	).Scan(&industry.ID, &industry.CreatedAt, &industry.UpdatedAt)
	if err != nil {
		log.Printf("创建/更新行业失败: %v", err)
		return fmt.Errorf("创建/更新行业失败: %w", err)
	}

	log.Printf("✓ 行业创建/更新成功: %s (ID: %d)", industry.Name, industry.ID)
	return nil
}

// GetIndustryByKeywords 根据关键词获取行业
func (r *ConflictExtendedRepository) GetIndustryByKeywords(ctx context.Context, keywords string) (*models.IndustryClassification, error) {
	query := `
		SELECT id, code, name, parent_id, level, description, keywords, created_at, updated_at
		FROM industry_classifications
		WHERE is_active = true
		AND ($1 ILIKE '%' || keywords || '%' OR keywords ILIKE '%' || $1 || '%')
		ORDER BY level ASC, id ASC
		LIMIT 1
	`

	var industry models.IndustryClassification
	err := r.db.QueryRowContext(ctx, query, keywords).Scan(
		&industry.ID, &industry.Code, &industry.Name, &industry.ParentID,
		&industry.Level, &industry.Description, &industry.Keywords,
		&industry.CreatedAt, &industry.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("未找到匹配的行业: %s", keywords)
			return nil, fmt.Errorf("未找到匹配的行业")
		}
		return nil, fmt.Errorf("查询行业失败: %w", err)
	}

	return &industry, nil
}

// GetIndustryByClientName 根据客户名称获取行业
func (r *ConflictExtendedRepository) GetIndustryByClientName(ctx context.Context, clientName string) (*models.IndustryClassification, error) {
	// 客户名称到行业的映射规则
	clientIndustryMappings := map[string]string{
		"阿里巴巴":   "TMT",
		"阿里":     "TMT",
		"淘宝":     "TMT",
		"天猫":     "TMT",
		"支付宝":    "TMT",
		"蚂蚁金服":   "TMT",
		"蚂蚁集团":   "TMT",
		"阿里云":    "TMT",
		"腾讯":     "TMT",
		"微信":     "TMT",
		"字节跳动":   "TMT",
		"抖音":     "TMT",
		"TikTok": "TMT",
		"今日头条":   "TMT",
		"百度":     "TMT",
		"京东":     "TMT",
		"美团":     "TMT",
	}

	for clientKeyword, industryCode := range clientIndustryMappings {
		if strings.Contains(clientName, clientKeyword) {
			return r.GetIndustryByKeywords(ctx, industryCode)
		}
	}

	// 如果没有匹配，返回默认的"其他"行业
	return &models.IndustryClassification{
		ID:    999,
		Code:  "OTHER",
		Name:  "其他",
		Level: 1,
	}, nil
}

// CreateOrUpdateCompetitiveRelation 创建或更新竞争关系
func (r *ConflictExtendedRepository) CreateOrUpdateCompetitiveRelation(ctx context.Context, relation *models.CompetitiveRelation) error {
	query := `
		INSERT INTO competitive_relations
		(industry_id, competitor_type, competitor_name, competitor_pattern, conflict_level, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		ON CONFLICT (industry_id, competitor_name) DO UPDATE SET
			competitor_type = EXCLUDED.competitor_type,
			competitor_pattern = EXCLUDED.competitor_pattern,
			conflict_level = EXCLUDED.conflict_level,
			description = EXCLUDED.description,
			is_active = EXCLUDED.is_active,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		relation.IndustryID, relation.CompetitorType, relation.CompetitorName,
		relation.CompetitorPattern, relation.ConflictLevel, relation.Description, relation.IsActive,
	).Scan(&relation.ID, &relation.CreatedAt, &relation.UpdatedAt)
	if err != nil {
		log.Printf("创建/更新竞争关系失败: %v", err)
		return fmt.Errorf("创建/更新竞争关系失败: %w", err)
	}

	log.Printf("✓ 竞争关系创建/更新成功: %s (ID: %d)", relation.CompetitorName, relation.ID)
	return nil
}

// GetCompetitiveRelationsByIndustry 根据行业获取竞争关系
func (r *ConflictExtendedRepository) GetCompetitiveRelationsByIndustry(ctx context.Context, industryID int) ([]models.CompetitiveRelation, error) {
	query := `
		SELECT id, industry_id, competitor_type, competitor_name, competitor_pattern,
		       conflict_level, description, is_active, created_at, updated_at
		FROM competitive_relations
		WHERE industry_id = $1 AND is_active = true
		ORDER BY competitor_type, conflict_level DESC, competitor_name
	`

	rows, err := r.db.QueryContext(ctx, query, industryID)
	if err != nil {
		return nil, fmt.Errorf("查询竞争关系失败: %w", err)
	}
	defer rows.Close()

	var relations []models.CompetitiveRelation
	for rows.Next() {
		var relation models.CompetitiveRelation
		err := rows.Scan(
			&relation.ID, &relation.IndustryID, &relation.CompetitorType,
			&relation.CompetitorName, &relation.CompetitorPattern,
			&relation.ConflictLevel, &relation.Description,
			&relation.IsActive, &relation.CreatedAt, &relation.UpdatedAt,
		)
		if err != nil {
			log.Printf("扫描竞争关系数据失败: %v", err)
			continue
		}
		relations = append(relations, relation)
	}

	return relations, nil
}

// CreateOrUpdateConflictRule 创建或更新冲突规则
func (r *ConflictExtendedRepository) CreateOrUpdateConflictRule(ctx context.Context, rule *models.EnhancedConflictRule) error {
	query := `
		INSERT INTO conflict_rules (name, rule_type, trigger_pattern, action_type, risk_score, conditions, is_active, priority, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		ON CONFLICT (name, rule_type) DO UPDATE SET
			trigger_pattern = EXCLUDED.trigger_pattern,
			action_type = EXCLUDED.action_type,
			risk_score = EXCLUDED.risk_score,
			conditions = EXCLUDED.conditions,
			is_active = EXCLUDED.is_active,
			priority = EXCLUDED.priority,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		rule.Name, rule.RuleType, rule.TriggerPattern, rule.ActionType,
		rule.RiskScore, rule.Conditions, rule.IsActive, rule.Priority,
	).Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)
	if err != nil {
		log.Printf("创建/更新冲突规则失败: %v", err)
		return fmt.Errorf("创建/更新冲突规则失败: %w", err)
	}

	log.Printf("✓ 冲突规则创建/更新成功: %s (ID: %d, 风险分数: %d)", rule.Name, rule.ID, rule.RiskScore)
	return nil
}

// GetActiveConflictRules 获取活跃的冲突规则
func (r *ConflictExtendedRepository) GetActiveConflictRules(ctx context.Context) ([]models.EnhancedConflictRule, error) {
	query := `
		SELECT id, name, rule_type, trigger_pattern, action_type, risk_score, conditions, is_active, priority, created_at, updated_at
		FROM conflict_rules
		WHERE is_active = true
		ORDER BY priority ASC, risk_score DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("查询冲突规则失败: %w", err)
	}
	defer rows.Close()

	var rules []models.EnhancedConflictRule
	for rows.Next() {
		var rule models.EnhancedConflictRule
		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.RuleType, &rule.TriggerPattern,
			&rule.ActionType, &rule.RiskScore, &rule.Conditions,
			&rule.IsActive, &rule.Priority, &rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			log.Printf("扫描冲突规则数据失败: %v", err)
			continue
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// CreateConflictDetectionHistory 创建冲突检测历史记录
func (r *ConflictExtendedRepository) CreateConflictDetectionHistory(ctx context.Context, history *models.ConflictDetectionHistory) error {
	query := `
		INSERT INTO conflict_detection_history
		(lawyer_id, case_id, client_name, opposing_party, case_type, detection_result, conflicts_found, risk_level, user_action, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query,
		history.LawyerID, history.CaseID, history.ClientName, history.OpposingParty,
		history.CaseType, history.DetectionResult, history.ConflictsFound,
		history.RiskLevel, history.UserAction,
	).Scan(&history.ID)
	if err != nil {
		log.Printf("创建冲突检测历史失败: %v", err)
		return fmt.Errorf("创建冲突检测历史失败: %w", err)
	}

	log.Printf("✓ 冲突检测历史创建成功: 律师ID %d, 客户 %s, 风险等级 %s",
		history.LawyerID, history.ClientName, history.RiskLevel)
	return nil
}

// GetPotentialConflictsAdvanced 高级潜在冲突检测
func (r *ConflictExtendedRepository) GetPotentialConflictsAdvanced(ctx context.Context, request *AdvancedConflictCheckRequest) ([]*models.Case, error) {
	log.Printf("=== 开始高级冲突检测 ===")
	log.Printf("律师ID: %d, 客户: %s, 对方: %s, 行业ID: %d, 包含行业分析: %t",
		request.LawyerID, request.ClientName, request.OpposingParty, request.IndustryID, request.IncludeIndustry)

	var conflicts []*models.Case

	// 1. 传统冲突检测（同一律师的直接冲突）
	traditionalConflicts, err := r.getTraditionalConflicts(ctx, request)
	if err != nil {
		log.Printf("传统冲突检测失败: %v", err)
	} else {
		conflicts = append(conflicts, traditionalConflicts...)
		log.Printf("发现 %d 个传统冲突", len(traditionalConflicts))
	}

	// 2. 行业竞争冲突检测
	if request.IncludeIndustry && request.IndustryID > 0 {
		industryConflicts, err := r.getIndustryConflicts(ctx, request)
		if err != nil {
			log.Printf("行业冲突检测失败: %v", err)
		} else {
			conflicts = append(conflicts, industryConflicts...)
			log.Printf("发现 %d 个行业竞争冲突", len(industryConflicts))
		}
	}

	// 3. 客户名称冲突检测
	if request.SearchDepth == "comprehensive" {
		nameConflicts, err := r.getClientNameConflicts(ctx, request)
		if err != nil {
			log.Printf("客户名称冲突检测失败: %v", err)
		} else {
			conflicts = append(conflicts, nameConflicts...)
			log.Printf("发现 %d 个客户名称冲突", len(nameConflicts))
		}
	}

	log.Printf("=== 高级冲突检测完成，总计发现 %d 个冲突 ===", len(conflicts))
	return conflicts, nil
}

// getTraditionalConflicts 获取传统冲突（同一律师的现有案件）
func (r *ConflictExtendedRepository) getTraditionalConflicts(ctx context.Context, request *AdvancedConflictCheckRequest) ([]*models.Case, error) {
	query := `
		SELECT c.id, c.title, c.description, c.case_type, c.status, c.priority,
		       c.client_id, c.lawyer_id, c.opposing_party, c.start_date, c.end_date,
		       c.created_at, c.updated_at, cl.name as client_name, cl.company as client_company
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		WHERE c.lawyer_id = $1
		AND (c.status = 'in_progress' OR c.status = 'pending')
		AND (
			LOWER(c.opposing_party) LIKE '%' || LOWER($2) || '%' OR
			LOWER(cl.name) LIKE '%' || LOWER($2) || '%' OR
			LOWER(cl.company) LIKE '%' || LOWER($2) || '%'
		)
		ORDER BY c.priority DESC, c.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, request.LawyerID, request.OpposingParty)
	if err != nil {
		return nil, fmt.Errorf("查询传统冲突失败: %w", err)
	}
	defer rows.Close()

	var conflicts []*models.Case
	for rows.Next() {
		var c models.Case
		var opposingParty sql.NullString
		var clientName, clientCompany sql.NullString
		err := rows.Scan(
			&c.ID, &c.Title, &c.Description, &c.CaseType, &c.Status, &c.Priority,
			&c.ClientID, &c.LawyerID, &opposingParty, &c.StartDate, &c.EndDate,
			&c.CreatedAt, &c.UpdatedAt, &clientName, &clientCompany,
		)
		if err != nil {
			log.Printf("扫描传统冲突数据失败: %v", err)
			continue
		}

		// 设置对方当事人和客户端名称
		if opposingParty.Valid {
			c.OpposingParty = opposingParty.String
		}

		if clientCompany.Valid && clientCompany.String != "" {
			c.ClientName = clientCompany.String
		} else if clientName.Valid {
			c.ClientName = clientName.String
		}

		conflicts = append(conflicts, &c)
	}

	return conflicts, nil
}

// getIndustryConflicts 获取行业竞争冲突
func (r *ConflictExtendedRepository) getIndustryConflicts(ctx context.Context, request *AdvancedConflictCheckRequest) ([]*models.Case, error) {
	query := `
		SELECT c.id, c.title, c.description, c.case_type, c.status, c.priority,
		       c.client_id, c.lawyer_id, c.opposing_party, c.start_date, c.end_date,
		       c.created_at, c.updated_at, cl.name as client_name, cl.company as client_company,
		       cl.industry as client_industry
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		WHERE c.lawyer_id = $1
		AND (c.status = 'in_progress' OR c.status = 'pending')
		AND cl.industry ILIKE '%' || (
			SELECT name FROM industry_classifications WHERE id = $2
		) || '%'
		ORDER BY c.priority DESC, c.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, request.LawyerID, request.IndustryID)
	if err != nil {
		return nil, fmt.Errorf("查询行业冲突失败: %w", err)
	}
	defer rows.Close()

	var conflicts []*models.Case
	for rows.Next() {
		var c models.Case
		var clientName, clientCompany, clientIndustry sql.NullString
		err := rows.Scan(
			&c.ID, &c.Title, &c.Description, &c.CaseType, &c.Status, &c.Priority,
			&c.ClientID, &c.LawyerID, &c.OpposingParty, &c.StartDate, &c.EndDate,
			&c.CreatedAt, &c.UpdatedAt, &clientName, &clientCompany, &clientIndustry,
		)
		if err != nil {
			log.Printf("扫描行业冲突数据失败: %v", err)
			continue
		}

		// 设置客户端信息
		if clientCompany.Valid && clientCompany.String != "" {
			c.ClientName = clientCompany.String
		} else if clientName.Valid {
			c.ClientName = clientName.String
		}

		conflicts = append(conflicts, &c)
		log.Printf("发现行业冲突: 案件%d - %s (客户: %s, 行业: %s)",
			c.ID, c.Title, c.ClientName, clientIndustry.String)
	}

	return conflicts, nil
}

// getClientNameConflicts 获取客户名称冲突
func (r *ConflictExtendedRepository) getClientNameConflicts(ctx context.Context, request *AdvancedConflictCheckRequest) ([]*models.Case, error) {
	query := `
		SELECT c.id, c.title, c.description, c.case_type, c.status, c.priority,
		       c.client_id, c.lawyer_id, c.opposing_party, c.start_date, c.end_date,
		       c.created_at, c.updated_at, cl.name as client_name, cl.company as client_company
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		WHERE c.lawyer_id = $1
		AND (c.status = 'in_progress' OR c.status = 'pending')
		AND (
			SIMILARITY(LOWER(cl.name), LOWER($3)) > 0.6 OR
			SIMILARITY(LOWER(cl.company), LOWER($3)) > 0.6 OR
			LOWER(cl.name) = LOWER($3) OR
			LOWER(cl.company) = LOWER($3)
		)
		ORDER BY c.priority DESC, c.created_at DESC
		LIMIT 10
	`

	rows, err := r.db.QueryContext(ctx, query, request.LawyerID, request.OpposingParty, request.ClientName)
	if err != nil {
		return nil, fmt.Errorf("查询客户名称冲突失败: %w", err)
	}
	defer rows.Close()

	var conflicts []*models.Case
	for rows.Next() {
		var c models.Case
		var clientName, clientCompany sql.NullString
		err := rows.Scan(
			&c.ID, &c.Title, &c.Description, &c.CaseType, &c.Status, &c.Priority,
			&c.ClientID, &c.LawyerID, &c.OpposingParty, &c.StartDate, &c.EndDate,
			&c.CreatedAt, &c.UpdatedAt, &clientName, &clientCompany,
		)
		if err != nil {
			log.Printf("扫描客户名称冲突数据失败: %v", err)
			continue
		}

		if clientCompany.Valid && clientCompany.String != "" {
			c.ClientName = clientCompany.String
		} else if clientName.Valid {
			c.ClientName = clientName.String
		}

		conflicts = append(conflicts, &c)
	}

	return conflicts, nil
}

// GetCasesByIndustryAndLawyer 根据行业和律师获取案件
func (r *ConflictExtendedRepository) GetCasesByIndustryAndLawyer(ctx context.Context, industryID, lawyerID int) ([]*models.Case, error) {
	query := `
		SELECT c.id, c.title, c.description, c.case_type, c.status, c.priority,
		       c.client_id, c.lawyer_id, c.opposing_party, c.start_date, c.end_date,
		       c.created_at, c.updated_at, cl.name as client_name, cl.company as client_company
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		WHERE c.lawyer_id = $1 AND cl.industry ILIKE '%' || (
			SELECT name FROM industry_classifications WHERE id = $2
		) || '%'
		ORDER BY c.created_at DESC
	`

	return r.scanCasesFromQuery(ctx, query, lawyerID, industryID)
}

// CheckDirectClientConflict 检查直接客户冲突
func (r *ConflictExtendedRepository) CheckDirectClientConflict(ctx context.Context, lawyerID int, clientName, opposingParty string) ([]*models.Case, error) {
	query := `
		SELECT c.id, c.title, c.description, c.case_type, c.status, c.priority,
		       c.client_id, c.lawyer_id, c.opposing_party, c.start_date, c.end_date,
		       c.created_at, c.updated_at, cl.name as client_name, cl.company as client_company
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		WHERE c.lawyer_id = $1
		AND (c.status = 'in_progress' OR c.status = 'pending')
		AND (
			LOWER(cl.name) = LOWER($2) OR LOWER(cl.company) = LOWER($2) OR
			LOWER(c.opposing_party) = LOWER($3) OR
			LOWER(cl.name) = LOWER($3) OR LOWER(cl.company) = LOWER($3)
		)
		ORDER BY c.priority DESC, c.created_at DESC
	`

	return r.scanCasesFromQuery(ctx, query, lawyerID, clientName, opposingParty)
}

// scanCasesFromQuery 从查询结果扫描案件数据
// CaseWithDetails 带有额外详情的案件结构
type CaseWithDetails struct {
	*models.Case
	OpposingParty string `json:"opposing_party"`
	ClientName    string `json:"client_name"`
}

func (r *ConflictExtendedRepository) scanCasesFromQuery(ctx context.Context, query string, args ...interface{}) ([]*models.Case, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询案件失败: %w", err)
	}
	defer rows.Close()

	var cases []*models.Case
	for rows.Next() {
		var c models.Case
		var opposingParty sql.NullString
		var clientName, clientCompany sql.NullString
		err := rows.Scan(
			&c.ID, &c.Title, &c.Description, &c.CaseType, &c.Status, &c.Priority,
			&c.ClientID, &c.LawyerID, &opposingParty, &c.StartDate, &c.EndDate,
			&c.CreatedAt, &c.UpdatedAt, &clientName, &clientCompany,
		)
		if err != nil {
			log.Printf("扫描案件数据失败: %v", err)
			continue
		}

		// 设置临时字段
		if opposingParty.Valid {
			c.OpposingParty = opposingParty.String
		}

		if clientCompany.Valid && clientCompany.String != "" {
			c.ClientName = clientCompany.String
		} else if clientName.Valid {
			c.ClientName = clientName.String
		}

		cases = append(cases, &c)
	}

	return cases, nil
}

// GetCheckHistory 获取冲突检测历史 (实现ConflictRepository接口要求)
func (r *ConflictExtendedRepository) GetCheckHistory(ctx context.Context, clientID string, limit int) ([]*models.ConflictCheckRecord, error) {
	// 这个方法主要用于兼容ConflictRepository接口
	// 从基础冲突检测记录表中查询历史
	query := `
		SELECT check_id, client_id, client_name, case_name, case_type, check_status,
		       has_conflict, risk_level, check_result, user_id, duration, check_time, created_at, updated_at
		FROM conflict_check_records
		WHERE client_id = $1 OR client_name = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, clientID, limit)
	if err != nil {
		return nil, fmt.Errorf("查询冲突检测历史失败: %w", err)
	}
	defer rows.Close()

	var history []*models.ConflictCheckRecord
	for rows.Next() {
		var record models.ConflictCheckRecord
		err := rows.Scan(
			&record.CheckID, &record.ClientID, &record.ClientName, &record.CaseName,
			&record.CaseType, &record.CheckStatus, &record.HasConflict, &record.RiskLevel,
			&record.CheckResult, &record.UserID, &record.Duration, &record.CheckTime,
			&record.CreatedAt, &record.UpdatedAt,
		)
		if err != nil {
			log.Printf("扫描冲突检测历史记录失败: %v", err)
			continue
		}
		history = append(history, &record)
	}

	return history, nil
}

// GetConflictStats 获取冲突检测统计信息
func (r *ConflictExtendedRepository) GetConflictStats(ctx context.Context, lawyerID int) (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(*) as total_checks,
			COUNT(CASE WHEN has_conflict = true THEN 1 END) as conflict_checks,
			COUNT(CASE WHEN risk_level = 'HIGH' THEN 1 END) as high_risk_checks,
			AVG(duration) as avg_duration
		FROM conflict_check_records
		WHERE user_id = $1
	`

	var totalChecks, conflictChecks, highRiskChecks int
	var avgDuration float64

	err := r.db.QueryRowContext(ctx, query, lawyerID).Scan(
		&totalChecks, &conflictChecks, &highRiskChecks, &avgDuration,
	)
	if err != nil {
		log.Printf("获取冲突统计失败: %v", err)
		return nil, err
	}

	stats := map[string]interface{}{
		"total_checks":     totalChecks,
		"conflict_checks":  conflictChecks,
		"high_risk_checks": highRiskChecks,
		"avg_duration":     avgDuration,
		"lawyer_id":        lawyerID,
	}

	return stats, nil
}
