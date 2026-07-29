package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func intakeClientID(value interface{}) uint {
	switch typed := value.(type) {
	case uint:
		return typed
	case uint64:
		return uint(typed)
	case int:
		if typed > 0 {
			return uint(typed)
		}
	case int64:
		if typed > 0 {
			return uint(typed)
		}
	case float64:
		if typed > 0 && typed == float64(uint(typed)) {
			return uint(typed)
		}
	case json.Number:
		if parsed, err := strconv.ParseUint(string(typed), 10, 32); err == nil {
			return uint(parsed)
		}
	case string:
		if parsed, err := strconv.ParseUint(strings.TrimSpace(typed), 10, 32); err == nil {
			return uint(parsed)
		}
	}
	return 0
}

func objectValue(value interface{}) map[string]interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		return typed
	case []byte:
		var result map[string]interface{}
		if json.Unmarshal(typed, &result) == nil {
			return result
		}
	case string:
		var result map[string]interface{}
		if json.Unmarshal([]byte(typed), &result) == nil {
			return result
		}
	}
	return map[string]interface{}{}
}

func valueString(values map[string]interface{}, key string) string {
	if values == nil {
		return ""
	}
	value := strings.TrimSpace(fmt.Sprint(values[key]))
	if value == "<nil>" {
		return ""
	}
	return value
}

func subjectCaseIDFromAggregateSearchParameters(value interface{}) string {
	parameters := objectValue(value)
	for _, key := range []string{"subjectCaseId", "subject_case_id", "caseId", "case_id"} {
		if candidate := valueString(parameters, key); candidate != "" && candidate != "<nil>" {
			return candidate
		}
	}
	return ""
}

func subjectIntakeIDFromAggregateSearchParameters(value interface{}) string {
	parameters := objectValue(value)
	for _, key := range []string{"intakeId", "intake_id"} {
		if candidate := valueString(parameters, key); candidate != "" && candidate != "<nil>" {
			return candidate
		}
	}
	return ""
}

func stringListValue(value interface{}) []string {
	var result []string
	switch typed := value.(type) {
	case []interface{}:
		for _, item := range typed {
			if value := strings.TrimSpace(fmt.Sprint(item)); value != "" && value != "<nil>" {
				result = append(result, value)
			}
		}
	case []string:
		for _, item := range typed {
			if value := strings.TrimSpace(item); value != "" {
				result = append(result, value)
			}
		}
	case string:
		for _, item := range strings.Split(typed, ",") {
			if value := strings.TrimSpace(item); value != "" {
				result = append(result, value)
			}
		}
	}
	return result
}

func stringMapValue(value interface{}) map[string]string {
	result := map[string]string{}
	if values, ok := value.(map[string]interface{}); ok {
		for key, item := range values {
			value := strings.TrimSpace(fmt.Sprint(item))
			if value != "" && value != "<nil>" {
				result[strings.ToLower(strings.TrimSpace(key))] = value
			}
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// containsPlaintextIdentityField rejects identity numbers embedded in free-form
// intake metadata. Identity values must enter through the protected client or
// entity registry so they can be encrypted, hashed for lookup, and audited.
func containsPlaintextIdentityField(value interface{}) bool {
	sensitiveKeys := map[string]struct{}{
		"idcard": {}, "clientidcard": {}, "identitynumber": {}, "passport": {},
		"unifiedsocialcreditcode": {}, "socialcreditcode": {}, "taxid": {},
		"entitytaxid": {}, "clienttaxid": {}, "organizationcode": {}, "orgcode": {},
	}
	var visit func(interface{}) bool
	visit = func(current interface{}) bool {
		switch typed := current.(type) {
		case map[string]interface{}:
			for key, item := range typed {
				normalized := strings.ToLower(strings.NewReplacer("_", "", "-", "", " ", "").Replace(strings.TrimSpace(key)))
				if _, sensitive := sensitiveKeys[normalized]; sensitive {
					return true
				}
				if visit(item) {
					return true
				}
			}
		case []interface{}:
			for _, item := range typed {
				if visit(item) {
					return true
				}
			}
		case []map[string]interface{}:
			for _, item := range typed {
				if visit(item) {
					return true
				}
			}
		}
		return false
	}
	return visit(value)
}

func firstNonEmptyValue(values ...interface{}) interface{} {
	for _, value := range values {
		switch typed := value.(type) {
		case nil:
			continue
		case string:
			if strings.TrimSpace(typed) != "" {
				return value
			}
		case map[string]interface{}:
			if len(typed) > 0 {
				return value
			}
		case []interface{}:
			if len(typed) > 0 {
				return value
			}
		default:
			if rendered := strings.TrimSpace(fmt.Sprint(value)); rendered != "" && rendered != "<nil>" {
				return value
			}
		}
	}
	return nil
}

var allowedCaseIntakeCreateFields = map[string]struct{}{
	"client_id":   {},
	"title":       {},
	"case_type":   {},
	"priority":    {},
	"description": {},
	"metadata":    {},
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

func (h *DemoAggregateHandler) inboxFilter(where string) string {
	if h.hasColumn("inbox_items", "deleted_at") {
		return "deleted_at IS NULL AND " + where
	}
	return where
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
	for index, row := range rows {
		rows[index] = sanitizeAggregateRow(table, row)
	}
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

// sanitizeAggregateRow is the last response boundary for legacy workbench
// aggregates. These handlers intentionally use map-backed queries for
// compatibility with several schema generations, so returning the map as-is
// would also return password hashes, identity ciphertext/digests, or raw
// workflow metadata that the screen never needs.
func sanitizeAggregateRow(table string, row map[string]interface{}) map[string]interface{} {
	if row == nil {
		return map[string]interface{}{}
	}
	result := make(map[string]interface{}, len(row))
	for key, value := range row {
		result[key] = value
	}

	switch strings.ToLower(strings.TrimSpace(table)) {
	case "users":
		for key := range result {
			if isAggregateSecretKey(key) {
				delete(result, key)
			}
		}
	case "clients":
		result = sanitizeClientAggregateRow(result)
	case "case_intake_parties", "case_materials", "case_intakes":
		if value, ok := result["metadata"]; ok {
			result["metadata"] = redactAggregateSensitiveValue(value)
		}
		for _, key := range []string{"identifiers", "client_identifiers", "clientIdentifiers", "id_card", "id_card_ciphertext", "id_card_digest", "unified_social_credit_code"} {
			delete(result, key)
		}
	case "approval_requests":
		// The approval workbench only needs summary columns. Full snapshots,
		// attachments and arbitrary metadata are served by their dedicated,
		// authorization-aware endpoints.
		for _, key := range []string{"metadata", "workflow_config", "attachments", "conflict_result", "approval_snapshot"} {
			delete(result, key)
		}
	case "system_settings":
		for key := range result {
			if isAggregateSecretKey(key) || strings.EqualFold(strings.TrimSpace(key), "value") || strings.EqualFold(strings.TrimSpace(key), "setting_value") {
				result[key] = "已配置（敏感值不在聚合接口返回）"
			}
		}
	case "risk_audit_events", "audit_logs":
		for _, key := range []string{"payload", "details", "metadata", "request_body", "response_body", "context"} {
			delete(result, key)
		}
	}
	return result
}

func sanitizeClientAggregateRow(row map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(row))
	for key, value := range row {
		result[key] = value
	}
	identityPresent := false
	for _, key := range []string{"id_card", "id_card_ciphertext", "id_card_digest", "unified_social_credit_code", "unifiedSocialCreditCode", "identifiers"} {
		if value := strings.TrimSpace(fmt.Sprint(result[key])); value != "" && value != "<nil>" {
			identityPresent = true
		}
		delete(result, key)
	}
	if identityPresent {
		result["identity_status"] = "已登记（受保护）"
	}
	for _, key := range []string{"phone", "contact_phone"} {
		if value := strings.TrimSpace(fmt.Sprint(result[key])); value != "" && value != "<nil>" {
			result[key] = maskAggregatePhone(value)
		}
	}
	return result
}

func isAggregateSecretKey(key string) bool {
	normalized := strings.ToLower(strings.NewReplacer("_", "", "-", "", " ", "").Replace(strings.TrimSpace(key)))
	for _, marker := range []string{"password", "passwd", "secret", "token", "apikey", "privatekey", "ciphertext", "digest", "identifier", "identitynumber", "idcard", "socialcreditcode"} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func redactAggregateSensitiveValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(typed))
		for key, item := range typed {
			if isAggregateSecretKey(key) {
				continue
			}
			result[key] = redactAggregateSensitiveValue(item)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(typed))
		for index, item := range typed {
			result[index] = redactAggregateSensitiveValue(item)
		}
		return result
	default:
		return value
	}
}

func maskAggregatePhone(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 7 {
		return "***"
	}
	return value[:3] + "****" + value[len(value)-4:]
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
