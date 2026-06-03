package handlers

import (
	"fmt"
	"law-oa-go/internal/common"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	stats := DashboardStats{
		TotalCases:     h.count("cases", "deleted_at IS NULL"),
		ActiveCases:    h.count("cases", "deleted_at IS NULL AND status IN ?", []string{"pending", "in_progress", "active"}),
		CompletedCases: h.count("cases", "deleted_at IS NULL AND status IN ?", []string{"completed", "closed"}),
		TotalClients:   h.count("clients", "deleted_at IS NULL"),
		NewClients:     h.count("clients", "deleted_at IS NULL AND created_at >= ?", time.Now().AddDate(0, 0, -30)),
		TotalLawyers:   h.count("users", "deleted_at IS NULL AND role = ? AND status = ?", "lawyer", "active"),
		MonthlyRevenue: h.monthlyRevenue(),
	}

	common.APISuccess(c, stats)
}

// GetDashboardTodos 获取待办事项
func (h *DashboardHandler) GetDashboardTodos(c *gin.Context) {
	todos := h.todoItems(10)

	common.APISuccess(c, gin.H{
		"todos": todos,
		"total": len(todos),
	})
}

// GetDashboardActivities 获取活动记录
func (h *DashboardHandler) GetDashboardActivities(c *gin.Context) {
	activities := h.activityItems(10)

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

func (h *DashboardHandler) monthlyRevenue() float64 {
	if !h.tableExists("payments") || !h.hasColumn("payments", "amount") {
		return 0
	}

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	var total float64
	query := h.db.Table("payments").
		Select("COALESCE(SUM(amount), 0)").
		Where("created_at >= ?", monthStart)

	if h.hasColumn("payments", "status") {
		query = query.Where("status IN ?", []string{"confirmed", "paid"})
	}

	if err := query.Scan(&total).Error; err != nil {
		return 0
	}
	return total
}

func (h *DashboardHandler) todoItems(limit int) []TodoItem {
	if !h.tableExists("inbox_items") {
		return []TodoItem{}
	}

	rows := make([]map[string]interface{}, 0)
	query := h.db.Table("inbox_items").
		Where("deleted_at IS NULL").
		Order("is_completed ASC, due_date ASC, created_at DESC").
		Limit(limit)

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

func (h *DashboardHandler) activityItems(limit int) []ActivityItem {
	if h.tableExists("risk_audit_events") {
		activities := h.activitiesFromTable("risk_audit_events", "risk", "summary", "event_type", "actor_id", limit)
		if len(activities) > 0 {
			return activities
		}
	}

	if h.tableExists("approval_requests") {
		activities := h.activitiesFromTable("approval_requests", "approval", "title", "status", "applicant_name", limit)
		if len(activities) > 0 {
			return activities
		}
	}

	if h.tableExists("cases") {
		return h.activitiesFromTable("cases", "case", "title", "status", "lawyer_id", limit)
	}

	return []ActivityItem{}
}

func (h *DashboardHandler) activitiesFromTable(table, itemType, titleColumn, descriptionColumn, userColumn string, limit int) []ActivityItem {
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
