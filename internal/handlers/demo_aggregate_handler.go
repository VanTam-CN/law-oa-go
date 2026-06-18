package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"law-oa-go/internal/common"
	"law-oa-go/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DemoAggregateHandler provides stable, database-backed aggregate payloads for
// the demo workbenches while the underlying domain services continue to evolve.
type DemoAggregateHandler struct {
	db *gorm.DB
}

func NewDemoAggregateHandler(db *gorm.DB) *DemoAggregateHandler {
	return &DemoAggregateHandler{db: db}
}

func (h *DemoAggregateHandler) CommandCenter(c *gin.Context) {
	activeStatuses := []string{"pending", "in_progress", "active"}
	openConflictStatuses := []string{"QUEUED", "RUNNING", "PROCESSING"}
	visibleConflictRiskLevels := []string{"HIGH", "CRITICAL", "MEDIUM"}
	pendingApprovalStatuses := []string{"submitted", "under_review", "resubmitted"}
	activeCasesWhere := "deleted_at IS NULL AND status IN ?"
	activeCasesArgs := []interface{}{activeStatuses}
	caseStageWhere := "deleted_at IS NULL"
	caseStageArgs := []interface{}{}
	if lawyerID, ok := currentLawyerScope(c); ok {
		activeCasesWhere += " AND lawyer_id = ?"
		activeCasesArgs = append(activeCasesArgs, lawyerID)
		caseStageWhere += " AND lawyer_id = ?"
		caseStageArgs = append(caseStageArgs, lawyerID)
	}
	conflictWhere := "check_status IN ? OR risk_level IN ?"
	conflictArgs := []interface{}{openConflictStatuses, visibleConflictRiskLevels}
	if lawyerID, ok := currentLawyerScope(c); ok {
		conflictWhere = "(" + conflictWhere + ") AND user_id = ?"
		conflictArgs = append(conflictArgs, lawyerID)
	}
	riskQueue := h.riskRows(c, 8)

	common.APISuccess(c, gin.H{
		"summary": gin.H{
			"active_cases":        h.count("cases", activeCasesWhere, activeCasesArgs...),
			"clients":             h.count("clients", "deleted_at IS NULL"),
			"pending_approvals":   h.count("approval_requests", "deleted_at IS NULL AND status IN ?", pendingApprovalStatuses),
			"open_conflict_tasks": h.count("conflict_check_records", conflictWhere, conflictArgs...),
			"unread_inbox":        h.countAny([]string{"inbox_items"}, "deleted_at IS NULL AND is_completed = ?", false),
		},
		"workflow": gin.H{
			"intake":     h.countAny([]string{"case_intakes"}, "status IN ?", []string{"draft", "materials_pending", "conflict_ready", "conflict_checking"}),
			"conflict":   h.count("conflict_check_records", conflictWhere, conflictArgs...),
			"approval":   h.count("approval_requests", "deleted_at IS NULL AND status IN ?", pendingApprovalStatuses),
			"activation": h.count("cases", activeCasesWhere, activeCasesArgs...),
		},
		"todo_items":              h.todoRows(10),
		"risk_queue":              riskQueue,
		"approval_queue":          h.approvalRows(8),
		"case_rows":               h.caseRows(c, 20),
		"case_stage_distribution": h.groupedCounts("cases", "status", caseStageWhere, 10, caseStageArgs...),
		"risk_distribution":       h.groupedCounts("conflict_check_records", "risk_level", conflictWhere, 10, conflictArgs...),
		"overdue_tasks":           h.overdueRows(5),
		"recent_activities":       h.activityRows(10),
		"generated_at":            time.Now(),
	})
}

func (h *DemoAggregateHandler) ClientMasterProfile(c *gin.Context) {
	id := c.Param("id")
	var client map[string]interface{}
	if h.first("clients", id, &client) != nil {
		common.APINotFound(c, "客户不存在", "指定客户不存在或已删除")
		return
	}

	common.APISuccess(c, gin.H{
		"client":           client,
		"completeness":     h.clientCompleteness(client),
		"related_parties":  h.clientRelatedParties(id, fmt.Sprint(client["name"])),
		"matter_history":   h.recentRows("cases", "client_id = ? AND deleted_at IS NULL", 5, id),
		"conflict_history": h.recentRowsAny([]string{"conflict_check_records"}, "client_id = ?", 5, fmt.Sprint(client["id"])),
	})
}

func (h *DemoAggregateHandler) IntakeWorkbench(c *gin.Context) {
	id := c.Param("id")
	var intake map[string]interface{}
	if h.first("case_intakes", id, &intake) != nil {
		common.APINotFound(c, "接案记录不存在", "指定接案记录不存在")
		return
	}
	parties := h.recentRows("case_intake_parties", "intake_id = ?", 20, intake["id"])
	materials := h.recentRows("case_materials", "intake_id = ?", 20, intake["id"])

	common.APISuccess(c, gin.H{
		"intake":    intake,
		"client":    h.firstByID("clients", intake["client_id"]),
		"parties":   parties,
		"materials": materials,
		"team":      h.recentRows("users", "deleted_at IS NULL AND role IN ?", 10, []string{"lawyer", "admin"}),
	})
}

func (h *DemoAggregateHandler) ApprovalsWorkbench(c *gin.Context) {
	common.APISuccess(c, gin.H{
		"stats": gin.H{
			"pending":        h.count("approval_requests", "deleted_at IS NULL AND status IN ?", []string{"submitted", "under_review", "resubmitted"}),
			"needs_revision": h.count("approval_requests", "deleted_at IS NULL AND status = ?", "needs_revision"),
			"waiver_review":  h.countAny([]string{"waiver_requests"}, "status IN ?", []string{"submitted", "under_review"}),
		},
		"items": h.recentRows("approval_requests", "deleted_at IS NULL", 12),
		"queues": []gin.H{
			{"key": "conflict", "label": "冲突审批", "count": h.count("approval_requests", "deleted_at IS NULL AND type = ?", "conflict_approval")},
			{"key": "waiver", "label": "豁免评估", "count": h.countAny([]string{"waiver_requests"}, "status IN ?", []string{"submitted", "under_review"})},
			{"key": "finance", "label": "财务审批", "count": h.count("approval_requests", "deleted_at IS NULL AND type = ?", "finance")},
		},
	})
}

func (h *DemoAggregateHandler) LawyersResourceCenter(c *gin.Context) {
	common.APISuccess(c, gin.H{
		"summary": gin.H{
			"lawyers":       h.count("users", "deleted_at IS NULL AND role IN ?", []string{"lawyer", "admin"}),
			"departments":   h.distinctCount("users", "department", "deleted_at IS NULL"),
			"active_cases":  h.count("cases", "deleted_at IS NULL AND status IN ?", []string{"pending", "in_progress"}),
			"pending_tasks": h.countAny([]string{"inbox_items"}, "deleted_at IS NULL AND is_completed = ?", false),
		},
		"lawyers":     h.recentRows("users", "deleted_at IS NULL AND role IN ?", 20, []string{"lawyer", "admin"}),
		"capacity":    h.groupedCounts("users", "department", "deleted_at IS NULL AND role IN ?", 20, []string{"lawyer", "admin"}),
		"assignments": h.caseRows(c, 8),
		"tasks":       h.todoRows(8),
	})
}

func (h *DemoAggregateHandler) AdminAccessCenter(c *gin.Context) {
	common.APISuccess(c, gin.H{
		"summary": gin.H{
			"users":           h.count("users", "deleted_at IS NULL"),
			"active_users":    h.count("users", "deleted_at IS NULL AND status = ?", "active"),
			"disabled_users":  h.count("users", "deleted_at IS NULL AND status <> ?", "active"),
			"roles":           h.countAny([]string{"roles"}, "deleted_at IS NULL"),
			"permissions":     h.countAny([]string{"permissions"}, ""),
			"pending_changes": h.count("approval_requests", "deleted_at IS NULL AND type IN ? AND status IN ?", []string{"permission_change", "access_request"}, []string{"submitted", "under_review"}),
		},
		"users":              h.recentRows("users", "deleted_at IS NULL", 20),
		"roles":              h.rbacRoleRows(20),
		"permission_changes": h.recentRows("approval_requests", "deleted_at IS NULL AND type IN ?", 8, []string{"permission_change", "access_request"}),
		"audit_events":       h.recentRowsAny([]string{"risk_audit_events", "audit_logs"}, "", 10),
	})
}

func (h *DemoAggregateHandler) SettingsOverview(c *gin.Context) {
	common.APISuccess(c, gin.H{
		"summary": gin.H{
			"settings": h.count("system_settings", ""),
			"modules":  h.distinctCount("system_settings", "category", ""),
		},
		"modules":  h.groupedCounts("system_settings", "category", "", 20),
		"settings": h.recentRowsAny([]string{"system_settings"}, "", 50),
	})
}

func (h *DemoAggregateHandler) ExecutiveDashboard(c *gin.Context) {
	common.APISuccess(c, gin.H{
		"kpis": gin.H{
			"revenue_ytd":       h.sumAny([]string{"payments"}, "amount", "status IN ?", []string{"confirmed", "paid"}),
			"active_cases":      h.count("cases", "deleted_at IS NULL AND status IN ?", []string{"pending", "in_progress"}),
			"new_clients_30d":   h.count("clients", "deleted_at IS NULL AND created_at >= ?", time.Now().AddDate(0, 0, -30)),
			"high_risk_matters": h.count("conflict_check_records", "risk_level IN ?", []string{"HIGH", "CRITICAL"}),
		},
		"trends": h.recentRows("cases", "deleted_at IS NULL", 12),
	})
}

func (h *DemoAggregateHandler) todoRows(limit int) []gin.H {
	rows := h.recentRowsAny([]string{"inbox_items"}, "deleted_at IS NULL AND is_completed = ?", limit, false)
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, gin.H{
			"id":          row["id"],
			"type":        row["source_type"],
			"title":       row["title"],
			"content":     row["content"],
			"priority":    row["priority"],
			"due_at":      row["due_date"],
			"is_read":     row["is_read"],
			"source_id":   row["source_id"],
			"source_type": row["source_type"],
		})
	}
	return items
}

func (h *DemoAggregateHandler) riskRows(c *gin.Context, limit int) []gin.H {
	where := "check_status IN ? OR risk_level IN ?"
	args := []interface{}{
		[]string{"QUEUED", "RUNNING", "PROCESSING"},
		[]string{"HIGH", "CRITICAL", "MEDIUM"},
	}
	if lawyerID, ok := currentLawyerScope(c); ok {
		where = "(" + where + ") AND user_id = ?"
		args = append(args, lawyerID)
	}
	rows := h.recentRows("conflict_check_records", where, limit, args...)
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		checkID := fmt.Sprint(row["check_id"])
		conflictCases := []map[string]interface{}{}
		if checkID != "" && h.tableExists("conflict_cases") {
			_ = h.db.Table("conflict_cases").
				Where("check_id = ?", checkID).
				Order("CASE risk_level WHEN 'CRITICAL' THEN 4 WHEN 'HIGH' THEN 3 WHEN 'MEDIUM' THEN 2 WHEN 'LOW' THEN 1 ELSE 0 END DESC, created_at DESC").
				Find(&conflictCases).Error
		}
		primaryConflict := primaryConflictCase(conflictCases)
		items = append(items, gin.H{
			"id":                row["check_id"],
			"title":             row["case_name"],
			"case_type":         row["case_type"],
			"client_id":         row["client_id"],
			"client_name":       row["client_name"],
			"matched_subject":   primaryConflictSubject(primaryConflict, row),
			"matched_type":      primaryConflictValue(primaryConflict, "conflict_type"),
			"evidence_summary":  primaryConflictValue(primaryConflict, "description"),
			"status":            row["check_status"],
			"risk_level":        row["risk_level"],
			"has_conflict":      row["has_conflict"],
			"owner":             row["user_id"],
			"duration":          row["duration"],
			"check_time":        row["check_time"],
			"created_at":        row["created_at"],
			"updated_at":        row["updated_at"],
			"search_parameters": row["search_parameters"],
			"check_result":      row["check_result"],
			"conflict_cases":    conflictCases,
		})
	}
	return items
}

func primaryConflictCase(conflictCases []map[string]interface{}) map[string]interface{} {
	if len(conflictCases) == 0 {
		return nil
	}
	return conflictCases[0]
}

func primaryConflictValue(conflictCase map[string]interface{}, key string) string {
	if conflictCase == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(conflictCase[key]))
}

func primaryConflictSubject(conflictCase map[string]interface{}, fallback map[string]interface{}) string {
	description := primaryConflictValue(conflictCase, "description")
	if subject := extractQuotedSubject(description, "当前对方当事人"); subject != "" {
		return subject
	}
	if subject := firstNonEmpty(
		primaryConflictValue(conflictCase, "case_name"),
		primaryConflictValue(conflictCase, "client_name"),
	); subject != "" {
		return subject
	}
	return strings.TrimSpace(fmt.Sprint(fallback["client_name"]))
}

func extractQuotedSubject(text string, label string) string {
	prefix := label + " '"
	start := strings.Index(text, prefix)
	if start < 0 {
		return ""
	}
	remaining := text[start+len(prefix):]
	end := strings.Index(remaining, "'")
	if end < 0 {
		return ""
	}
	return strings.TrimSpace(remaining[:end])
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" && trimmed != "<nil>" {
			return trimmed
		}
	}
	return ""
}

func (h *DemoAggregateHandler) approvalRows(limit int) []gin.H {
	rows := h.recentRows("approval_requests", "deleted_at IS NULL AND status IN ?", limit, []string{"submitted", "under_review", "resubmitted"})
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, gin.H{
			"id":                    row["id"],
			"request_number":        row["request_number"],
			"title":                 row["title"],
			"type":                  row["type"],
			"status":                row["status"],
			"priority":              row["priority"],
			"current_stage":         row["current_stage"],
			"current_approver_name": row["current_approver_name"],
			"created_at":            row["created_at"],
			"timeout_at":            row["timeout_at"],
		})
	}
	return items
}

func currentLawyerScope(c *gin.Context) (uint, bool) {
	if role, _ := middleware.GetCurrentRole(c); role != "lawyer" {
		return 0, false
	}
	return middleware.GetCurrentUserID(c)
}

func (h *DemoAggregateHandler) caseRows(c *gin.Context, limit int) []gin.H {
	if !h.tableExists("cases") {
		return []gin.H{}
	}
	var rows []map[string]interface{}
	q := h.db.Table("cases AS c").
		Select("c.id, c.case_number, c.title, c.client_id, c.lawyer_id, c.case_type, c.status, c.priority, c.updated_at, cl.name AS client_name, u.name AS lawyer_name").
		Joins("LEFT JOIN clients cl ON c.client_id = cl.id").
		Joins("LEFT JOIN users u ON c.lawyer_id = u.id").
		Where("c.deleted_at IS NULL").
		Order("c.updated_at DESC").
		Limit(limit)
	if lawyerID, ok := currentLawyerScope(c); ok {
		q = q.Where("c.lawyer_id = ?", lawyerID)
	}
	_ = q.Scan(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, gin.H{
			"id":          row["id"],
			"case_number": row["case_number"],
			"title":       row["title"],
			"client_id":   row["client_id"],
			"lawyer_id":   row["lawyer_id"],
			"client_name": row["client_name"],
			"case_type":   row["case_type"],
			"status":      row["status"],
			"priority":    row["priority"],
			"lawyer_name": row["lawyer_name"],
			"updated_at":  row["updated_at"],
		})
	}
	return items
}

func (h *DemoAggregateHandler) overdueRows(limit int) []gin.H {
	rows := h.recentRowsAny([]string{"inbox_items"}, "deleted_at IS NULL AND is_completed = ? AND due_date < ?", limit, false, time.Now())
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, gin.H{
			"id":       row["id"],
			"title":    row["title"],
			"priority": row["priority"],
			"due_at":   row["due_date"],
		})
	}
	return items
}

func (h *DemoAggregateHandler) activityRows(limit int) []gin.H {
	rows := h.recentRowsAny([]string{"risk_audit_events"}, "", limit)
	if len(rows) == 0 {
		rows = h.recentRows("approval_requests", "deleted_at IS NULL", limit)
	}
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		title := row["summary"]
		if title == nil {
			title = row["title"]
		}
		items = append(items, gin.H{
			"id":         row["id"],
			"title":      title,
			"type":       row["event_type"],
			"created_at": row["created_at"],
			"actor_id":   row["actor_id"],
		})
	}
	return items
}

func (h *DemoAggregateHandler) clientCompleteness(client map[string]interface{}) gin.H {
	required := []struct {
		Key   string
		Label string
	}{
		{"name", "客户名称"},
		{"type", "客户类型"},
		{"email", "电子邮箱"},
		{"phone", "联系电话"},
		{"address", "地址"},
		{"industry", "所属行业"},
		{"contact_person", "联系人"},
	}
	missing := make([]string, 0)
	checks := make([]gin.H, 0, len(required))
	for _, item := range required {
		value := fmt.Sprint(client[item.Key])
		complete := value != "" && value != "<nil>"
		if !complete {
			missing = append(missing, item.Key)
		}
		status := "complete"
		if !complete {
			status = "missing"
		}
		checks = append(checks, gin.H{"key": item.Key, "label": item.Label, "status": status})
	}
	score := 0
	if len(required) > 0 {
		score = (len(required) - len(missing)) * 100 / len(required)
	}
	return gin.H{
		"score":                    score,
		"missing_fields":           missing,
		"ready_for_conflict_check": len(missing) == 0,
		"checks":                   checks,
	}
}

func (h *DemoAggregateHandler) clientRelatedParties(clientID, clientName string) []gin.H {
	if !h.tableExists("case_intake_parties") || !h.tableExists("case_intakes") {
		return []gin.H{}
	}
	var rows []map[string]interface{}
	_ = h.db.Table("case_intake_parties AS p").
		Select("p.entity_name AS name, p.party_role AS relationship_type, p.relation_depth AS depth, p.metadata").
		Joins("JOIN case_intakes ci ON ci.id = p.intake_id").
		Where("ci.client_id = ? AND p.entity_name <> ?", clientID, clientName).
		Order("p.created_at DESC").
		Limit(20).
		Scan(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, gin.H{
			"name":              row["name"],
			"relationship_type": row["relationship_type"],
			"depth":             row["depth"],
			"metadata":          row["metadata"],
		})
	}
	return items
}

func (h *DemoAggregateHandler) firstByID(table string, id interface{}) map[string]interface{} {
	if id == nil {
		return nil
	}
	var row map[string]interface{}
	if h.first(table, fmt.Sprint(id), &row) == nil {
		return row
	}
	return nil
}

func (h *DemoAggregateHandler) CreateCaseIntake(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}
	parties := mapSliceValue(payload["parties"])
	materials := mapSliceValue(payload["materials"])
	intakePayload := filterMap(payload, allowedCaseIntakeCreateFields)
	now := time.Now()
	intakeID := uuid.NewString()
	intakePayload["id"] = intakeID
	intakePayload["intake_code"] = fmt.Sprintf("INT-%s", now.Format("20060102150405"))
	intakePayload["status"] = stringValue(intakePayload["status"], "draft")
	intakePayload["priority"] = stringValue(intakePayload["priority"], "medium")
	intakePayload["metadata"] = jsonStringValue(intakePayload["metadata"])
	intakePayload["created_at"] = now
	intakePayload["updated_at"] = now

	if h.tableExists("case_intakes") {
		err := h.db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Table("case_intakes").Create(intakePayload).Error; err != nil {
				return err
			}
			if h.tableExists("case_intake_parties") {
				for _, party := range parties {
					row := filterMap(party, allowedCaseIntakePartyFields)
					row["intake_id"] = intakeID
					row["entity_name"] = stringValue(row["entity_name"], stringValue(party["name"], "未命名主体"))
					row["entity_type"] = stringValue(row["entity_type"], "company")
					row["party_role"] = stringValue(row["party_role"], stringValue(party["role"], "related_party"))
					if _, ok := row["relation_depth"]; !ok {
						row["relation_depth"] = 0
					}
					row["metadata"] = jsonStringValue(row["metadata"])
					row["created_at"] = now
					if err := tx.Table("case_intake_parties").Create(row).Error; err != nil {
						return err
					}
				}
			}
			if h.tableExists("case_materials") {
				for _, material := range materials {
					row := filterMap(material, allowedCaseMaterialFields)
					row["intake_id"] = intakeID
					row["name"] = stringValue(row["name"], "未命名材料")
					row["material_type"] = stringValue(row["material_type"], "document")
					row["status"] = stringValue(row["status"], "missing")
					if _, ok := row["required"]; !ok {
						row["required"] = true
					}
					row["metadata"] = jsonStringValue(row["metadata"])
					row["created_at"] = now
					row["updated_at"] = now
					if err := tx.Table("case_materials").Create(row).Error; err != nil {
						return err
					}
				}
			}
			return nil
		})
		if err != nil {
			common.APIInternalServerError(c, "创建接案失败", err.Error())
			return
		}
	}
	intakePayload["parties"] = parties
	intakePayload["materials"] = materials
	c.JSON(http.StatusCreated, common.APIResponse{Success: true, Data: intakePayload, Meta: common.ResponseMeta{Timestamp: now, Version: "v1", Server: "law-oa-go", Environment: "development"}})
}

func (h *DemoAggregateHandler) UpdateCaseIntake(c *gin.Context) {
	id := c.Param("id")
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}
	payload = filterMap(payload, allowedCaseIntakeUpdateFields)
	if len(payload) == 0 {
		common.APIBadRequest(c, "请求参数错误", "没有可更新的字段")
		return
	}
	if _, ok := payload["metadata"]; ok {
		payload["metadata"] = jsonStringValue(payload["metadata"])
	}
	payload["updated_at"] = time.Now()
	if h.tableExists("case_intakes") {
		if err := h.db.Table("case_intakes").Where("id = ?", id).Updates(payload).Error; err != nil {
			common.APIInternalServerError(c, "更新接案失败", err.Error())
			return
		}
	}
	payload["id"] = id
	common.APISuccess(c, payload)
}

func (h *DemoAggregateHandler) StartIntakeConflictCheck(c *gin.Context) {
	intakeID := c.Param("id")
	var intake map[string]interface{}
	if h.first("case_intakes", intakeID, &intake) != nil {
		common.APINotFound(c, "接案记录不存在", "指定接案记录不存在")
		return
	}
	taskID := "CCT_" + uuid.NewString()
	now := time.Now()
	if h.tableExists("conflict_check_records") {
		record := gin.H{
			"check_id":          taskID,
			"client_id":         fmt.Sprint(intake["client_id"]),
			"client_name":       fmt.Sprint(intake["client_id"]),
			"case_name":         fmt.Sprint(intake["title"]),
			"case_type":         fmt.Sprint(intake["case_type"]),
			"check_status":      "QUEUED",
			"has_conflict":      false,
			"risk_level":        "LOW",
			"search_parameters": "{}",
			"check_result":      "{}",
			"check_time":        now,
			"created_at":        now,
			"updated_at":        now,
		}
		if err := h.db.Table("conflict_check_records").Create(record).Error; err != nil {
			common.APIInternalServerError(c, "创建冲突任务失败", err.Error())
			return
		}
	}
	common.APISuccess(c, gin.H{
		"taskId":                     taskID,
		"intake_id":                  intakeID,
		"status":                     "QUEUED",
		"recommendedPollingInterval": 2,
		"createdAt":                  now,
	})
}

func (h *DemoAggregateHandler) CreateConflictApproval(c *gin.Context) {
	taskID := c.Param("task_id")
	if !h.tableExists("approval_requests") {
		common.APIInternalServerError(c, "审批表不存在", "approval_requests table is required")
		return
	}

	var payload map[string]interface{}
	_ = c.ShouldBindJSON(&payload)

	approvalID := uuid.NewString()
	now := time.Now()
	requestNumber := fmt.Sprintf("APR-%s", now.Format("20060102150405"))
	title := stringValue(payload["title"], fmt.Sprintf("冲突审查审批 - %s", taskID))
	content := stringValue(payload["content"], "利益冲突检查结果进入审批复核。")
	applicantName := stringValue(payload["applicant_name"], "当前用户")
	departmentName := stringValue(payload["department_name"], "合规风控部")
	priority := stringValue(payload["priority"], "high")
	conflictRecord := gin.H{}
	conflictCases := []map[string]interface{}{}
	if h.tableExists("conflict_check_records") {
		var row map[string]interface{}
		if err := h.db.Table("conflict_check_records").Where("check_id = ?", taskID).Take(&row).Error; err == nil {
			conflictRecord = gin.H{
				"check_id":     row["check_id"],
				"case_name":    row["case_name"],
				"client_id":    row["client_id"],
				"client_name":  row["client_name"],
				"case_type":    row["case_type"],
				"status":       row["check_status"],
				"risk_level":   row["risk_level"],
				"has_conflict": row["has_conflict"],
				"check_result": row["check_result"],
				"created_at":   row["created_at"],
				"updated_at":   row["updated_at"],
			}
		}
	}
	if h.tableExists("conflict_cases") {
		_ = h.db.Table("conflict_cases").
			Where("check_id = ?", taskID).
			Order("risk_level DESC, created_at DESC").
			Find(&conflictCases).Error
	}
	conflictResult := gin.H{
		"checkId": taskID,
		"riskAssessment": gin.H{
			"overallRisk": stringValue(conflictRecord["risk_level"], "LOW"),
			"riskScore":   0,
		},
		"record":        conflictRecord,
		"conflictCases": conflictCases,
	}
	metadata := gin.H{
		"conflict_task_id": taskID,
		"source":           "real_api",
		"conflict_record":  conflictRecord,
		"conflict_result":  conflictResult,
		"conflict_cases":   conflictCases,
	}

	approval := gin.H{
		"id":                    approvalID,
		"request_number":        requestNumber,
		"title":                 title,
		"type":                  "conflict_approval",
		"category":              "conflict_review",
		"content":               content,
		"applicant_id":          stringValue(payload["applicant_id"], "1"),
		"applicant_name":        applicantName,
		"applicant_title":       stringValue(payload["applicant_title"], "律师"),
		"department_id":         stringValue(payload["department_id"], "risk"),
		"department_name":       departmentName,
		"urgency":               stringValue(payload["urgency"], "urgent"),
		"priority":              priority,
		"status":                "submitted",
		"submission_date":       now,
		"current_stage":         "合规复核",
		"current_approver_id":   stringValue(payload["current_approver_id"], "1"),
		"current_approver_name": stringValue(payload["current_approver_name"], "合规负责人"),
		"workflow_type":         "CONFLICT_APPROVAL",
		"workflow_config":       "{}",
		"attachments":           "[]",
		"metadata":              jsonStringValue(metadata),
		"created_by":            stringValue(payload["created_by"], "1"),
		"created_at":            now,
		"updated_at":            now,
	}
	approvalRow := cleanInsertMap(approval)
	if err := h.db.Table("approval_requests").Create(&approvalRow).Error; err != nil {
		common.APIInternalServerError(c, "创建冲突审批失败", err.Error())
		return
	}

	if h.tableExists("approval_snapshots") {
		snapshot := gin.H{
			"id":                  uuid.NewString(),
			"approval_request_id": approvalID,
			"snapshot_type":       "conflict_approval",
			"snapshot_data": jsonStringValue(gin.H{
				"conflict_task_id": taskID,
				"conflict_record":  conflictRecord,
				"conflict_result":  conflictResult,
				"conflict_cases":   conflictCases,
				"approval":         approval,
				"metadata":         approval["metadata"],
			}),
			"source_version": 1,
			"created_at":     now,
		}
		snapshotRow := cleanInsertMap(snapshot)
		_ = h.db.Table("approval_snapshots").Create(&snapshotRow).Error
	}

	common.APISuccess(c, gin.H{
		"approval_id":    approvalID,
		"request_number": requestNumber,
		"conflict_task":  taskID,
		"status":         "submitted",
		"snapshot_url":   fmt.Sprintf("/api/v1/approvals/%s/snapshot", approvalID),
		"submitted_at":   now,
	})
}

func (h *DemoAggregateHandler) GetApprovalSnapshot(c *gin.Context) {
	id := c.Param("id")
	if h.tableExists("approval_snapshots") {
		var row map[string]interface{}
		if err := h.db.Table("approval_snapshots").Where("approval_request_id = ?", id).Order("created_at DESC").Take(&row).Error; err == nil {
			common.APISuccess(c, gin.H{
				"approval_id": id,
				"snapshot":    row["snapshot_data"],
				"immutable":   true,
			})
			return
		}
	}
	if h.tableExists("approval_requests") {
		var approval map[string]interface{}
		if err := h.db.Table("approval_requests").Where("id = ?", id).Take(&approval).Error; err == nil {
			metadataText := fmt.Sprint(approval["metadata"])
			var metadata map[string]interface{}
			if err := json.Unmarshal([]byte(metadataText), &metadata); err == nil {
				if snapshot, ok := metadata["approval_snapshot"]; ok {
					common.APISuccess(c, gin.H{
						"approval_id": id,
						"snapshot":    snapshot,
						"immutable":   true,
					})
					return
				}
			}
		}
	}
	common.APINotFound(c, "审批快照不存在", "指定审批没有可用的不可变快照")
}
