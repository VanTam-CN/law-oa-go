package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"law-oa-go/internal/common"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/services"
)

type DashboardHandler struct {
	db *gorm.DB
}

func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

type DashboardStats struct {
	TotalCases     int     `json:"total_cases"`
	ActiveCases    int     `json:"active_cases"`
	CompletedCases int     `json:"completed_cases"`
	TotalClients   int     `json:"total_clients"`
	NewClients     int     `json:"new_clients"`
	TotalLawyers   int     `json:"total_lawyers"`
	MonthlyRevenue float64 `json:"monthly_revenue"`
}

type TodoItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	DueDate     string `json:"due_date"`
	Status      string `json:"status"`
}

type ActivityItem struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	User        string `json:"user"`
}

// GetDashboardStatistics 获取仪表盘统计数据
func (h *DashboardHandler) GetDashboardStatistics(c *gin.Context) {
	caseScope, caseArgs := h.scopedDashboardScope(c, "cases")
	clientScope, clientArgs := h.scopedDashboardScope(c, "clients")
	stats := DashboardStats{
		TotalCases:     h.countScoped("cases", "deleted_at IS NULL", nil, caseScope, caseArgs...),
		ActiveCases:    h.countScoped("cases", "deleted_at IS NULL AND status IN ?", []interface{}{[]string{"pending", "in_progress", "active"}}, caseScope, caseArgs...),
		CompletedCases: h.countScoped("cases", "deleted_at IS NULL AND status IN ?", []interface{}{[]string{"completed", "closed"}}, caseScope, caseArgs...),
		TotalClients:   h.countScoped("clients", "deleted_at IS NULL", nil, clientScope, clientArgs...),
		NewClients:     h.countScoped("clients", "deleted_at IS NULL AND created_at >= ?", []interface{}{time.Now().AddDate(0, 0, -30)}, clientScope, clientArgs...),
		TotalLawyers:   h.dashboardLawyerCount(c),
		MonthlyRevenue: h.monthlyRevenue(c),
	}

	common.APISuccess(c, stats)
}

// GetDashboardTodos 获取待办事项
func (h *DashboardHandler) GetDashboardTodos(c *gin.Context) {
	todos := h.todoItems(10, c)

	common.APISuccess(c, gin.H{
		"todos": todos,
		"total": len(todos),
	})
}

// GetDashboardActivities 获取活动记录
func (h *DashboardHandler) GetDashboardActivities(c *gin.Context) {
	activities := h.activityItems(10, c)

	common.APISuccess(c, gin.H{
		"activities": activities,
		"total":      len(activities),
	})
}

func (h *DashboardHandler) tableExists(table string) bool {
	return h.db != nil && h.db.Migrator().HasTable(table)
}

func (h *DashboardHandler) hasColumn(table, column string) bool {
	return h.db != nil && h.db.Migrator().HasColumn(table, column)
}

func (h *DashboardHandler) count(table, where string, args ...interface{}) int {
	if !h.tableExists(table) {
		return 0
	}

	var total int64
	query := h.db.Table(table)
	if where != "" {
		query = query.Where(where, args...)
	}
	if err := query.Count(&total).Error; err != nil {
		return 0
	}
	return int(total)
}

func (h *DashboardHandler) countScoped(table, baseWhere string, baseArgs []interface{}, scope string, scopeArgs ...interface{}) int {
	where := baseWhere
	if scope != "" {
		where += " AND (" + scope + ")"
	}
	args := append([]interface{}{}, baseArgs...)
	args = append(args, scopeArgs...)
	return h.count(table, where, args...)
}

func (h *DashboardHandler) dashboardLawyerCount(c *gin.Context) int {
	if services.IsBusinessMatterManagementRole(c.GetString("role")) {
		return h.count("users", "deleted_at IS NULL AND role = ? AND status = ?", "lawyer", "active")
	}
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		return 0
	}
	return h.count("users", "deleted_at IS NULL AND id = ?", userID)
}

func (h *DashboardHandler) monthlyRevenue(c *gin.Context) float64 {
	// The legacy dashboard has no reliable payment -> contract -> case
	// subject context, so a firm-wide payment sum could reveal information
	// about an ethically walled matter. Finance must expose only a separately
	// authorized, wall-aware report; this compatibility field is unavailable.
	return 0
}

func (h *DashboardHandler) todoItems(limit int, c *gin.Context) []TodoItem {
	if !h.tableExists("inbox_items") {
		return []TodoItem{}
	}

	rows := make([]map[string]interface{}, 0)
	query := h.db.Table("inbox_items").
		Where("deleted_at IS NULL").
		Order("is_completed ASC, due_date ASC, created_at DESC").
		Limit(limit)
	if scope, args := h.scopedDashboardScope(c, "inbox_items"); scope != "" {
		query = query.Where(scope, args...)
	}

	if err := query.Find(&rows).Error; err != nil {
		return []TodoItem{}
	}

	items := make([]TodoItem, 0, len(rows))
	for _, row := range rows {
		status := "pending"
		if boolValue(row["is_completed"]) {
			status = "completed"
		}
		items = append(items, TodoItem{
			ID:          stringValue(row["id"], ""),
			Title:       stringValue(row["title"], ""),
			Description: stringValue(row["content"], ""),
			Priority:    stringValue(row["priority"], "medium"),
			DueDate:     timeString(row["due_date"]),
			Status:      status,
		})
	}
	return items
}

func (h *DashboardHandler) activityItems(limit int, c *gin.Context) []ActivityItem {
	if h.tableExists("risk_audit_events") {
		activities := h.activitiesFromTable("risk_audit_events", "risk", "summary", "event_type", "actor_id", limit, c)
		if len(activities) > 0 {
			return activities
		}
	}

	if h.tableExists("approval_requests") {
		activities := h.activitiesFromTable("approval_requests", "approval", "title", "status", "applicant_name", limit, c)
		if len(activities) > 0 {
			return activities
		}
	}

	if h.tableExists("cases") {
		return h.activitiesFromTable("cases", "case", "title", "status", "lawyer_id", limit, c)
	}

	return []ActivityItem{}
}

func (h *DashboardHandler) activitiesFromTable(table, itemType, titleColumn, descriptionColumn, userColumn string, limit int, c *gin.Context) []ActivityItem {
	if !h.tableExists(table) {
		return []ActivityItem{}
	}

	rows := make([]map[string]interface{}, 0)
	query := h.db.Table(table)
	if h.hasColumn(table, "deleted_at") {
		query = query.Where("deleted_at IS NULL")
	}
	if h.hasColumn(table, "created_at") {
		query = query.Order("created_at DESC")
	}
	if scope, args := h.scopedDashboardScope(c, table); scope != "" {
		query = query.Where(scope, args...)
	}

	if err := query.Limit(limit).Find(&rows).Error; err != nil {
		return []ActivityItem{}
	}

	items := make([]ActivityItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, ActivityItem{
			ID:          stringValue(row["id"], ""),
			Type:        itemType,
			Title:       stringValue(row[titleColumn], itemType),
			Description: fmt.Sprintf("%s: %s", descriptionColumn, stringValue(row[descriptionColumn], "")),
			CreatedAt:   timeString(row["created_at"]),
			User:        stringValue(row[userColumn], ""),
		})
	}
	return items
}

// scopedDashboardScope adds the ethical-wall constraint that the legacy
// dashboard endpoints previously lacked. The free dashboardScope helper
// below is retained for compatibility with older unit tests and callers; all
// HTTP handlers must use this database-backed method.
func (h *DashboardHandler) scopedDashboardScope(c *gin.Context, table string) (string, []interface{}) {
	baseScope, baseArgs := dashboardScope(c, table)
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok || userID == 0 {
		return "1 = 0", nil
	}

	// These legacy rows carry free-form titles or summaries and do not expose
	// a verified subject-case relationship. Returning no rows is safer than
	// allowing a wall-protected matter to leak through an activity feed.
	if table == "approval_requests" || table == "risk_audit_events" {
		return "1 = 0", nil
	}
	if services.IsBusinessMatterManagementRole(c.GetString("role")) && table == "inbox_items" {
		return "1 = 0", nil
	}

	// A missing wall column means the database is not on the production
	// schema. Do not silently fall back to an unscoped dashboard in that case.
	if (table == "cases" || table == "clients") &&
		(!h.tableExists("cases") || !h.hasColumn("cases", "ethical_wall_enabled") || !h.tableExists("case_ethical_wall_whitelist")) {
		return "1 = 0", nil
	}

	wallScope, wallArgs := dashboardEthicalWallScope(table, userID)
	if baseScope == "" {
		return wallScope, wallArgs
	}
	return "(" + baseScope + ") AND (" + wallScope + ")", append(baseArgs, wallArgs...)
}

func dashboardEthicalWallScope(table string, userID uint) (string, []interface{}) {
	switch table {
	case "cases":
		return `(
			cases.ethical_wall_enabled = ?
			OR EXISTS (
				SELECT 1 FROM case_ethical_wall_whitelist wall_access
				WHERE wall_access.case_id = cases.id AND wall_access.user_id = ?
			)
		)`, []interface{}{false, userID}
	case "clients":
		// A client with no cases is visible. Once a client has a protected
		// case, it is hidden until the viewer is explicitly whitelisted for
		// that case.
		return `NOT EXISTS (
			SELECT 1 FROM cases protected_case
			WHERE protected_case.client_id = clients.id
			  AND protected_case.deleted_at IS NULL
			  AND protected_case.ethical_wall_enabled = ?
			  AND NOT EXISTS (
				SELECT 1 FROM case_ethical_wall_whitelist wall_access
				WHERE wall_access.case_id = protected_case.id AND wall_access.user_id = ?
			  )
		)`, []interface{}{true, userID}
	default:
		return "1 = 0", nil
	}
}

// dashboardScope returns the server-side visibility predicate for legacy
// dashboard queries. Management roles may see firm-wide aggregates; every
// other authenticated role receives only its own matter/task/approval rows.
// Unknown or missing identities fail closed to an empty result.
func dashboardScope(c *gin.Context, table string) (string, []interface{}) {
	if c == nil || services.IsBusinessMatterManagementRole(c.GetString("role")) {
		return "", nil
	}
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok || userID == 0 {
		return "1 = 0", nil
	}
	userIDString := strconv.FormatUint(uint64(userID), 10)
	switch table {
	case "cases":
		return "lawyer_id = ? OR created_by = ?", []interface{}{userID, userIDString}
	case "clients":
		return `EXISTS (
			SELECT 1 FROM cases case_scope
			WHERE case_scope.client_id = clients.id
			  AND case_scope.deleted_at IS NULL
			  AND (case_scope.lawyer_id = ? OR case_scope.created_by = ?
			       OR EXISTS (
					SELECT 1 FROM case_ethical_wall_whitelist wall_access
					WHERE wall_access.case_id = case_scope.id AND wall_access.user_id = ?
				))
		)`, []interface{}{userID, userIDString, userID}
	case "inbox_items":
		return "user_id = ?", []interface{}{userID}
	case "risk_audit_events":
		return "actor_id = ?", []interface{}{userIDString}
	case "approval_requests":
		return "applicant_id = ? OR current_approver_id = ?", []interface{}{userIDString, userIDString}
	default:
		return "1 = 0", nil
	}
}

func boolValue(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case int:
		return v != 0
	case int64:
		return v != 0
	case uint:
		return v != 0
	case string:
		return v == "true" || v == "1"
	default:
		return false
	}
}

func timeString(value interface{}) string {
	switch v := value.(type) {
	case time.Time:
		return v.Format(time.RFC3339)
	case *time.Time:
		if v == nil {
			return ""
		}
		return v.Format(time.RFC3339)
	case []byte:
		return string(v)
	default:
		return stringValue(value, "")
	}
}
