package optimizations

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// QueryOptimizer 查询优化器
type QueryOptimizer struct {
	db *gorm.DB
}

// NewQueryOptimizer 创建查询优化器
func NewQueryOptimizer(db *gorm.DB) *QueryOptimizer {
	return &QueryOptimizer{db: db}
}

// OptimizedCaseListQuery 优化的案件列表查询
type OptimizedCaseListQuery struct {
	Status   string
	CaseType string
	Priority string
	ClientID uint
	LawyerID uint
	Search   string
	Page     int
	PageSize int
	OrderBy  string
}

// ExecuteOptimizedCaseList 执行优化的案件列表查询
func (qo *QueryOptimizer) ExecuteOptimizedCaseList(ctx context.Context, query *OptimizedCaseListQuery) ([]map[string]interface{}, int64, error) {
	// 优化分页参数
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	// 构建基础查询
	dbQuery := qo.db.WithContext(ctx).Table("cases")

	// 添加筛选条件
	whereConditions := make([]string, 0)
	args := make([]interface{}, 0)

	if query.Status != "" {
		whereConditions = append(whereConditions, "status = ?")
		args = append(args, query.Status)
	}
	if query.CaseType != "" {
		whereConditions = append(whereConditions, "case_type = ?")
		args = append(args, query.CaseType)
	}
	if query.Priority != "" {
		whereConditions = append(whereConditions, "priority = ?")
		args = append(args, query.Priority)
	}
	if query.ClientID > 0 {
		whereConditions = append(whereConditions, "client_id = ?")
		args = append(args, query.ClientID)
	}
	if query.LawyerID > 0 {
		whereConditions = append(whereConditions, "lawyer_id = ?")
		args = append(args, query.LawyerID)
	}

	// 优化搜索查询
	if query.Search != "" {
		searchTerm := "%" + strings.ToLower(query.Search) + "%"
		// 使用更高效的搜索条件
		if isNumeric(query.Search) {
			whereConditions = append(whereConditions, "(LOWER(title) LIKE ? OR LOWER(description) LIKE ? OR id = ?)")
			args = append(args, searchTerm, searchTerm, query.Search)
		} else {
			whereConditions = append(whereConditions, "(LOWER(title) LIKE ? OR LOWER(description) LIKE ?)")
			args = append(args, searchTerm, searchTerm)
		}
	}

	// 构建WHERE子句
	if len(whereConditions) > 0 {
		dbQuery = dbQuery.Where(strings.Join(whereConditions, " AND "), args...)
	}

	// 获取总数（使用优化的计数查询）
	var total int64
	countQuery := qo.db.WithContext(ctx).Table("cases")
	if len(whereConditions) > 0 {
		countQuery = countQuery.Where(strings.Join(whereConditions, " AND "), args...)
	}

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count cases: %w", err)
	}

	// 优化排序
	orderBy := "created_at DESC"
	if query.OrderBy != "" {
		// 验证排序字段
		allowedOrders := map[string]bool{
			"created_at": true,
			"updated_at": true,
			"title":      true,
			"status":     true,
			"priority":   true,
			"case_type":  true,
		}

		if allowedOrders[query.OrderBy] {
			orderBy = query.OrderBy + " DESC"
		}
	}

	// 优化分页查询 - 使用索引字段
	offset := (query.Page - 1) * query.PageSize
	var results []map[string]interface{}

	querySQL := `
		SELECT
			c.id,
			c.title,
			c.case_type,
			c.status,
			c.priority,
			c.created_at,
			c.updated_at,
			c.client_id,
			c.lawyer_id,
			cl.name as client_name,
			cl.company as client_company,
			u.name as lawyer_name,
			(CASE
				WHEN c.status = 'pending' THEN 1
				WHEN c.status = 'active' THEN 2
				WHEN c.status = 'closed' THEN 3
				WHEN c.status = 'suspended' THEN 4
				ELSE 5
			END) as status_order,
			(CASE
				WHEN c.priority = 'urgent' THEN 1
				WHEN c.priority = 'high' THEN 2
				WHEN c.priority = 'medium' THEN 3
				WHEN c.priority = 'low' THEN 4
				ELSE 5
			END) as priority_order
		FROM cases c
		LEFT JOIN clients cl ON c.client_id = cl.id
		LEFT JOIN users u ON c.lawyer_id = u.id
	`

	// 添加WHERE条件
	if len(whereConditions) > 0 {
		querySQL += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	// 添加排序和分页
	querySQL += " ORDER BY " + orderBy
	querySQL += " LIMIT ? OFFSET ?"
	finalArgs := append(args, query.PageSize, offset)

	if err := qo.db.WithContext(ctx).Raw(querySQL, finalArgs...).Scan(&results).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list cases: %w", err)
	}

	return results, total, nil
}

// OptimizedCaseDetailQuery 优化的案件详情查询
func (qo *QueryOptimizer) OptimizedCaseDetailQuery(ctx context.Context, caseID uint) (map[string]interface{}, error) {
	var result map[string]interface{}

	// 使用原生SQL进行优化查询
	query := `
		SELECT
			c.*,
			cl.name as client_name,
			cl.company as client_company,
			cl.email as client_email,
			cl.phone as client_phone,
			u.name as lawyer_name,
			u.email as lawyer_email,
			u.phone as lawyer_phone,
			(
				SELECT COUNT(*)
				FROM documents d
				WHERE d.case_id = c.id AND d.deleted_at IS NULL
			) as document_count,
			(
				SELECT MAX(created_at)
				FROM case_activities ca
				WHERE ca.case_id = c.id
			) as last_activity_at
		FROM cases c
		LEFT JOIN clients cl ON c.client_id = cl.id
		LEFT JOIN users u ON c.lawyer_id = u.id
		WHERE c.id = ?
	`

	if err := qo.db.WithContext(ctx).Raw(query, caseID).Scan(&result).Error; err != nil {
		return nil, fmt.Errorf("failed to get case details: %w", err)
	}

	return result, nil
}

// OptimizedSearchQuery 优化的搜索查询
func (qo *QueryOptimizer) OptimizedSearchQuery(ctx context.Context, searchTerm string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 使用全文搜索索引（如果可用）
	var results []map[string]interface{}
	var total int64

	if isNumeric(searchTerm) {
		// 数字ID精确匹配
		caseQuery := `
			SELECT
				id, title, case_type, status, priority, created_at,
				'case' as search_type,
				1 as relevance_score
			FROM cases
			WHERE id = ?
			ORDER BY created_at DESC
		`

		if err := qo.db.WithContext(ctx).Raw(caseQuery, searchTerm).Scan(&results).Error; err != nil {
			return nil, 0, fmt.Errorf("failed to search by id: %w", err)
		}

		total = int64(len(results))
	} else {
		// 文本搜索
		searchPattern := "%" + strings.ToLower(searchTerm) + "%"

		searchQuery := `
			SELECT
				id, title, case_type, status, priority, created_at,
				'case' as search_type,
				CASE
					WHEN LOWER(title) LIKE ? THEN 3
					WHEN LOWER(description) LIKE ? THEN 2
					ELSE 1
				END as relevance_score
			FROM cases
			WHERE LOWER(title) LIKE ? OR LOWER(description) LIKE ?
			ORDER BY relevance_score DESC, created_at DESC
		`

		if err := qo.db.WithContext(ctx).Raw(searchQuery, searchPattern, searchPattern, searchPattern, searchPattern).Scan(&results).Error; err != nil {
			return nil, 0, fmt.Errorf("failed to search cases: %w", err)
		}

		total = int64(len(results))
	}

	// 客户搜索
	clientQuery := `
		SELECT
			id, name, company as display_name, 'client' as search_type,
			2 as relevance_score
		FROM clients
		WHERE LOWER(name) LIKE ? OR LOWER(company) LIKE ?
		ORDER BY relevance_score DESC
	`

	var clientResults []map[string]interface{}
	searchPattern := "%" + strings.ToLower(searchTerm) + "%"

	if err := qo.db.WithContext(ctx).Raw(clientQuery, searchPattern, searchPattern).Scan(&clientResults).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search clients: %w", err)
	}

	total += int64(len(clientResults))
	results = append(results, clientResults...)

	// 律师搜索
	lawyerQuery := `
		SELECT
			id, name, 'lawyer' as search_type,
			2 as relevance_score
		FROM users
		WHERE role = 'lawyer' AND LOWER(name) LIKE ?
		ORDER BY relevance_score DESC
	`

	var lawyerResults []map[string]interface{}
	if err := qo.db.WithContext(ctx).Raw(lawyerQuery, searchPattern).Scan(&lawyerResults).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search lawyers: %w", err)
	}

	total += int64(len(lawyerResults))
	results = append(results, lawyerResults...)

	// 应用分页
	offset := (page - 1) * pageSize
	if offset < len(results) {
		end := offset + pageSize
		if end > len(results) {
			end = len(results)
		}
		results = results[offset:end]
	} else {
		results = []map[string]interface{}{}
	}

	return results, total, nil
}

// OptimizedStatsQuery 优化的统计查询
func (qo *QueryOptimizer) OptimizedStatsQuery(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 并行执行多个统计查询
	type result struct {
		TotalCases   int64 `json:"total_cases"`
		ActiveCases  int64 `json:"active_cases"`
		PendingCases int64 `json:"pending_cases"`
		ClosedCases  int64 `json:"closed_cases"`
		TotalClients int64 `json:"total_clients"`
		TotalLawyers int64 `json:"total_lawyers"`
	}

	var r result

	// 使用原生SQL进行高效的统计查询
	statsQuery := `
		SELECT
			(SELECT COUNT(*) FROM cases WHERE deleted_at IS NULL) as total_cases,
			(SELECT COUNT(*) FROM cases WHERE status IN ('active', 'in_progress') AND deleted_at IS NULL) as active_cases,
			(SELECT COUNT(*) FROM cases WHERE status = 'pending' AND deleted_at IS NULL) as pending_cases,
			(SELECT COUNT(*) FROM cases WHERE status IN ('closed', 'completed', 'suspended') AND deleted_at IS NULL) as closed_cases,
			(SELECT COUNT(*) FROM clients WHERE deleted_at IS NULL) as total_clients,
			(SELECT COUNT(*) FROM users WHERE role = 'lawyer' AND deleted_at IS NULL) as total_lawyers
	`

	if err := qo.db.WithContext(ctx).Raw(statsQuery).Scan(&r).Error; err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	stats["total_cases"] = r.TotalCases
	stats["active_cases"] = r.ActiveCases
	stats["pending_cases"] = r.PendingCases
	stats["closed_cases"] = r.ClosedCases
	stats["total_clients"] = r.TotalClients
	stats["total_lawyers"] = r.TotalLawyers

	// 按类型统计
	caseTypeStats := make(map[string]int64)
	var typeResults []struct {
		CaseType string `json:"case_type"`
		Count    int64  `json:"count"`
	}

	typeQuery := `SELECT case_type, COUNT(*) as count FROM cases WHERE deleted_at IS NULL GROUP BY case_type`
	if err := qo.db.WithContext(ctx).Raw(typeQuery).Scan(&typeResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get type stats: %w", err)
	}

	for _, tr := range typeResults {
		caseTypeStats[tr.CaseType] = tr.Count
	}
	stats["cases_by_type"] = caseTypeStats

	// 按优先级统计
	priorityStats := make(map[string]int64)
	var priorityResults []struct {
		Priority string `json:"priority"`
		Count    int64  `json:"count"`
	}

	priorityQuery := `SELECT priority, COUNT(*) as count FROM cases WHERE deleted_at IS NULL GROUP BY priority`
	if err := qo.db.WithContext(ctx).Raw(priorityQuery).Scan(&priorityResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get priority stats: %w", err)
	}

	for _, pr := range priorityResults {
		priorityStats[pr.Priority] = pr.Count
	}
	stats["cases_by_priority"] = priorityStats

	// 近期案件统计
	var recentCases []struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}

	recentQuery := `
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM cases
		WHERE created_at >= NOW() - INTERVAL '30 days' AND deleted_at IS NULL
		GROUP BY DATE(created_at)
		ORDER BY date DESC
		LIMIT 30
	`

	if err := qo.db.WithContext(ctx).Raw(recentQuery).Scan(&recentCases).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent stats: %w", err)
	}

	stats["recent_cases"] = recentCases

	return stats, nil
}

// 辅助函数
func isNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// CreateIndexes 创建性能优化所需的数据库索引
func (qo *QueryOptimizer) CreateIndexes() error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_cases_status ON cases(status)",
		"CREATE INDEX IF NOT EXISTS idx_cases_case_type ON cases(case_type)",
		"CREATE INDEX IF NOT EXISTS idx_cases_priority ON cases(priority)",
		"CREATE INDEX IF NOT EXISTS idx_cases_client_id ON cases(client_id)",
		"CREATE INDEX IF NOT EXISTS idx_cases_lawyer_id ON cases(lawyer_id)",
		"CREATE INDEX IF NOT EXISTS idx_cases_created_at ON cases(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_cases_updated_at ON cases(updated_at)",
		"CREATE INDEX IF NOT EXISTS idx_cases_title ON cases(title)",
		"CREATE INDEX IF NOT EXISTS idx_cases_description ON cases(description)",
		"CREATE INDEX IF NOT EXISTS idx_clients_name ON clients(name)",
		"CREATE INDEX IF NOT EXISTS idx_clients_company ON clients(company)",
		"CREATE INDEX IF NOT EXISTS idx_users_name ON users(name)",
		"CREATE INDEX IF NOT EXISTS idx_users_role ON users(role)",
		"CREATE INDEX IF NOT EXISTS idx_documents_case_id ON documents(case_id)",
		"CREATE INDEX IF NOT EXISTS idx_case_activities_case_id ON case_activities(case_id)",
	}

	for _, indexSQL := range indexes {
		if err := qo.db.Exec(indexSQL).Error; err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// AnalyzeQueryPerformance 分析查询性能
func (qo *QueryOptimizer) AnalyzeQueryPerformance(ctx context.Context, query string) (map[string]interface{}, error) {
	// 执行EXPLAIN查询
	explainQuery := "EXPLAIN " + query
	var explainResults []map[string]interface{}

	if err := qo.db.WithContext(ctx).Raw(explainQuery).Scan(&explainResults).Error; err != nil {
		return nil, fmt.Errorf("failed to analyze query: %w", err)
	}

	// 分析结果
	analysis := map[string]interface{}{
		"query":         query,
		"explain":       explainResults,
		"analysis_time": time.Now(),
	}

	return analysis, nil
}

// OptimizeDatabase 优化数据库配置
func (qo *QueryOptimizer) OptimizeDatabase() error {
	// 检测数据库类型并应用相应的优化
	driverName := qo.db.Dialector.Name()

	if driverName == "mysql" {
		// MySQL优化配置
		optimizations := []string{
			"SET GLOBAL innodb_buffer_pool_size = 1073741824", // 1GB
			"SET GLOBAL innodb_log_file_size = 268435456",     // 256MB
			"SET GLOBAL innodb_flush_log_at_trx_commit = 2",
			"SET GLOBAL sync_binlog = 0",
			"SET GLOBAL innodb_flush_method = O_DIRECT",
			"SET GLOBAL innodb_file_per_table = 1",
		}

		for _, opt := range optimizations {
			if err := qo.db.Exec(opt).Error; err != nil {
				return fmt.Errorf("failed to apply MySQL optimization: %w", err)
			}
		}
	} else if driverName == "postgres" {
		// PostgreSQL优化配置
		optimizations := []string{
			"ALTER SYSTEM SET shared_buffers = '256MB'",
			"ALTER SYSTEM SET effective_cache_size = '1GB'",
			"ALTER SYSTEM SET maintenance_work_mem = '64MB'",
			"ALTER SYSTEM SET checkpoint_completion_target = 0.9",
			"ALTER SYSTEM SET wal_buffers = '16MB'",
			"ALTER SYSTEM SET default_statistics_target = 100",
			"ALTER SYSTEM SET random_page_cost = 1.1",
			"ALTER SYSTEM SET effective_io_concurrency = 200",
		}

		for _, opt := range optimizations {
			if err := qo.db.Exec(opt).Error; err != nil {
				// 某些配置可能需要超级用户权限，记录但不报错
				fmt.Printf("Warning: Failed to apply PostgreSQL optimization '%s': %v\n", opt, err)
			}
		}

		// 重新加载配置
		if err := qo.db.Exec("SELECT pg_reload_conf()").Error; err != nil {
			fmt.Printf("Warning: Failed to reload PostgreSQL configuration: %v\n", err)
		}
	}

	return nil
}
