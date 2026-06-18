package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var allowedCaseIntakeCreateFields = map[string]struct{}{
	"client_id":   {},
	"title":       {},
	"case_type":   {},
	"status":      {},
	"priority":    {},
	"description": {},
	"metadata":    {},
	"created_by":  {},
}

var allowedCaseIntakePartyFields = map[string]struct{}{
	"case_id":        {},
	"intake_id":      {},
	"entity_name":    {},
	"entity_type":    {},
	"party_role":     {},
	"relation_depth": {},
	"metadata":       {},
	"created_at":     {},
}

var allowedCaseMaterialFields = map[string]struct{}{
	"case_id":       {},
	"intake_id":     {},
	"name":          {},
	"material_type": {},
	"status":        {},
	"required":      {},
	"storage_url":   {},
	"metadata":      {},
	"created_at":    {},
	"updated_at":    {},
}

var allowedCaseIntakeUpdateFields = map[string]struct{}{
	"client_id":   {},
	"title":       {},
	"case_type":   {},
	"status":      {},
	"priority":    {},
	"description": {},
	"metadata":    {},
}

var allowedDemoSumColumns = map[string]map[string]struct{}{
	"payments": {
		"amount": {},
	},
}

func (h *DemoAggregateHandler) first(table, id string, dest *map[string]interface{}) error {
	if !h.tableExists(table) {
		return gorm.ErrRecordNotFound
	}
	q := h.db.Table(table).Where("id = ?", id)
	if h.hasColumn(table, "deleted_at") {
		q = q.Where("deleted_at IS NULL")
	}
	return q.Take(dest).Error
}

func (h *DemoAggregateHandler) recentRows(table, where string, limit int, args ...interface{}) []map[string]interface{} {
	if !h.tableExists(table) {
		return []map[string]interface{}{}
	}
	var rows []map[string]interface{}
	q := h.db.Table(table)
	if where != "" {
		if len(args) == 1 && args[0] == nil {
			q = q.Where(where)
		} else {
			q = q.Where(where, args...)
		}
	}
	if h.hasColumn(table, "created_at") {
		q = q.Order("created_at DESC")
	}
	_ = q.Limit(limit).Find(&rows).Error
	return rows
}

func (h *DemoAggregateHandler) recentRowsAny(tables []string, where string, limit int, args ...interface{}) []map[string]interface{} {
	for _, table := range tables {
		if h.tableExists(table) {
			return h.recentRows(table, where, limit, args...)
		}
	}
	return []map[string]interface{}{}
}

func (h *DemoAggregateHandler) count(table, where string, args ...interface{}) int64 {
	if !h.tableExists(table) {
		return 0
	}
	var total int64
	q := h.db.Table(table)
	if where != "" {
		if len(args) == 1 && args[0] == nil {
			q = q.Where(where)
		} else {
			q = q.Where(where, args...)
		}
	}
	_ = q.Count(&total).Error
	return total
}

func (h *DemoAggregateHandler) countAny(tables []string, where string, args ...interface{}) int64 {
	for _, table := range tables {
		if h.tableExists(table) {
			return h.count(table, where, args...)
		}
	}
	return 0
}

func (h *DemoAggregateHandler) distinctCount(table, column, where string) int64 {
	if !h.tableExists(table) || !h.hasColumn(table, column) {
		return 0
	}
	var total int64
	q := h.db.Table(table).Distinct(column)
	if where != "" {
		q = q.Where(where)
	}
	_ = q.Count(&total).Error
	return total
}

func (h *DemoAggregateHandler) sumAny(tables []string, column, where string, args ...interface{}) float64 {
	for _, table := range tables {
		if !h.isAllowedSumColumn(table, column) || !h.tableExists(table) || !h.hasColumn(table, column) {
			continue
		}
		var total float64
		q := h.db.Table(table).Select(clause.Expr{
			SQL:                "COALESCE(SUM(?), 0)",
			Vars:               []interface{}{clause.Column{Name: column}},
			WithoutParentheses: true,
		})
		if where != "" {
			if len(args) == 1 && args[0] == nil {
				q = q.Where(where)
			} else {
				q = q.Where(where, args...)
			}
		}
		_ = q.Scan(&total).Error
		return total
	}
	return 0
}

func (h *DemoAggregateHandler) groupedCounts(table, column, where string, limit int, args ...interface{}) []gin.H {
	if !h.tableExists(table) || !h.hasColumn(table, column) {
		return []gin.H{}
	}
	var rows []struct {
		Key   string `gorm:"column:metric_key"`
		Count int64  `gorm:"column:count"`
	}
	q := h.db.Table(table).Select(fmt.Sprintf("%s AS metric_key, COUNT(*) AS count", column))
	if where != "" {
		if len(args) == 1 && args[0] == nil {
			q = q.Where(where)
		} else {
			q = q.Where(where, args...)
		}
	}
	_ = q.Group(column).Order("count DESC").Limit(limit).Scan(&rows).Error
	result := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		result = append(result, gin.H{"key": row.Key, "count": row.Count})
	}
	return result
}

func (h *DemoAggregateHandler) rbacRoleRows(limit int) []gin.H {
	if !h.tableExists("roles") || !h.hasColumn("roles", "code") {
		return h.groupedCounts("users", "role", "deleted_at IS NULL", limit)
	}

	var rows []struct {
		Key   string `gorm:"column:metric_key"`
		Count int64  `gorm:"column:count"`
	}

	query := h.db.Table("roles").Where("roles.deleted_at IS NULL")
	if h.hasColumn("roles", "status") {
		query = query.Where("roles.status = ?", "active")
	}

	if h.tableExists("user_roles") {
		query = query.
			Select("roles.code AS metric_key, COUNT(user_roles.user_id) AS count").
			Joins("LEFT JOIN user_roles ON user_roles.role_id = roles.id").
			Group("roles.id, roles.code, roles.sort_order").
			Order("roles.sort_order ASC, roles.id ASC")
	} else {
		query = query.
			Select("roles.code AS metric_key, 0 AS count").
			Order("roles.sort_order ASC, roles.id ASC")
	}

	_ = query.Limit(limit).Scan(&rows).Error
	result := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		result = append(result, gin.H{"key": row.Key, "count": row.Count})
	}
	return result
}

func (h *DemoAggregateHandler) tableExists(table string) bool {
	return h.db != nil && h.db.Migrator().HasTable(table)
}

func (h *DemoAggregateHandler) hasColumn(table, column string) bool {
	return h.db != nil && h.db.Migrator().HasColumn(table, column)
}

func (h *DemoAggregateHandler) isAllowedSumColumn(table, column string) bool {
	columns, ok := allowedDemoSumColumns[table]
	if !ok {
		return false
	}
	_, ok = columns[column]
	return ok
}

func filterMap(input map[string]interface{}, allowed map[string]struct{}) map[string]interface{} {
	filtered := make(map[string]interface{}, len(input))
	for key, value := range input {
		if _, ok := allowed[key]; ok {
			filtered[key] = value
		}
	}
	return filtered
}

func cleanInsertMap(input gin.H) map[string]interface{} {
	cleaned := make(map[string]interface{}, len(input))
	for key, value := range input {
		if value != nil {
			cleaned[key] = value
		}
	}
	return cleaned
}

func stringValue(value interface{}, fallback string) string {
	if s, ok := value.(string); ok && s != "" {
		return s
	}
	return fallback
}

func jsonStringValue(value interface{}) string {
	if value == nil {
		return "{}"
	}
	if s, ok := value.(string); ok && s != "" {
		return s
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func mapSliceValue(value interface{}) []map[string]interface{} {
	items, ok := value.([]interface{})
	if !ok {
		return nil
	}
	rows := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		row, ok := item.(map[string]interface{})
		if ok {
			rows = append(rows, row)
		}
	}
	return rows
}
