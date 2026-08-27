package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"law-oa-go/internal/common"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/security"
	"law-oa-go/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DemoAggregateHandler provides stable, database-backed aggregate payloads for
// the demo workbenches while the underlying domain services continue to evolve.
type DemoAggregateHandler struct {
	db              *gorm.DB
	conflictService services.ConflictDetectionService
	contactService  *services.ClientContactService
	authz           *services.AuthorizationService
}

var allowedAssistantDraftFields = map[string]struct{}{
	"title":       {},
	"case_type":   {},
	"priority":    {},
	"description": {},
}

func NewDemoAggregateHandler(db *gorm.DB) *DemoAggregateHandler {
	return &DemoAggregateHandler{db: db}
}

// SetConflictDetectionService wires the real conflict engine into the intake
// workbench. The legacy endpoint used to create a queued row without ever
// executing a search, which made the UI report a result that did not exist.
func (h *DemoAggregateHandler) SetConflictDetectionService(service services.ConflictDetectionService) {
	h.conflictService = service
}

func (h *DemoAggregateHandler) SetAuthorizationService(authz *services.AuthorizationService) {
	h.authz = authz
}

func (h *DemoAggregateHandler) SetClientContactService(service *services.ClientContactService) {
	h.contactService = service
}

// authorizeIntakeClient keeps a client selected in the intake workbench
// behind the same object boundary as the client API. The production router
// always wires authz; a missing service is therefore an explicit fail-closed
// configuration error rather than an invitation to trust a caller-supplied ID.
func (h *DemoAggregateHandler) authorizeIntakeClient(c *gin.Context, clientID uint) bool {
	if clientID == 0 {
		common.APIBadRequest(c, "接案客户无效", "必须选择有效的客户档案")
		return false
	}
	if h.authz == nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "CLIENT_AUTHZ_UNAVAILABLE", "客户权限服务未初始化，已阻止使用案件客户上下文")
		return false
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return false
	}
	allowed, err := h.authz.CanReadClient(c.Request.Context(), actor, clientID)
	if err != nil {
		common.APIInternalServerError(c, "客户权限校验失败", err.Error())
		return false
	}
	if !allowed {
		forbidObjectAccess(c)
		return false
	}
	return true
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
	pendingApprovalWhere := "deleted_at IS NULL AND status IN ?"
	pendingApprovalArgs := []interface{}{pendingApprovalStatuses}
	role := c.GetString("role")
	inboxWhere := h.inboxFilter("is_completed = ?")
	inboxArgs := []interface{}{false}
	if userID, ok := middleware.GetCurrentUserID(c); ok && userID > 0 {
		wallWhere := `(
			ethical_wall_enabled = ?
			OR EXISTS (
				SELECT 1 FROM case_ethical_wall_whitelist wall_access
				WHERE wall_access.case_id = cases.id AND wall_access.user_id = ?
			)
		)`
		activeCasesWhere += " AND " + wallWhere
		activeCasesArgs = append(activeCasesArgs, false, userID)
		caseStageWhere += " AND " + wallWhere
		caseStageArgs = append(caseStageArgs, false, userID)
	}
	if lawyerID, ok := currentLawyerScope(c); ok {
		activeCasesWhere += " AND lawyer_id = ?"
		activeCasesArgs = append(activeCasesArgs, lawyerID)
		caseStageWhere += " AND lawyer_id = ?"
		caseStageArgs = append(caseStageArgs, lawyerID)
		pendingApprovalWhere += " AND (applicant_id = ? OR current_approver_id = ?)"
		pendingApprovalArgs = append(pendingApprovalArgs, fmt.Sprint(lawyerID), fmt.Sprint(lawyerID))
		inboxWhere += " AND user_id = ?"
		inboxArgs = append(inboxArgs, lawyerID)
	} else if userID, ok := middleware.GetCurrentUserID(c); ok && services.IsConflictReviewRole(role) {
		// Dedicated conflict reviewers are not general matter managers, but they
		// still need approvals explicitly assigned to them. Object-level checks
		// below continue to enforce the ethical-wall boundary.
		pendingApprovalWhere += " AND (applicant_id = ? OR current_approver_id = ?)"
		pendingApprovalArgs = append(pendingApprovalArgs, fmt.Sprint(userID), fmt.Sprint(userID))
	}
	if canViewAllMatterData(c) {
		// inbox_items only stores an assignee and a source ID. The legacy
		// aggregate query cannot prove that source belongs to a visible case.
		inboxWhere = "1 = 0"
		inboxArgs = nil
	}
	if !canViewAllMatterData(c) {
		// Conflict officers and technical administrators are not general matter
		// managers. They may use their dedicated queues, but must not infer firm
		// workload, approval volume, inbox volume, or stage distribution from the
		// command-center aggregates.
		if _, lawyerScoped := currentLawyerScope(c); !lawyerScoped {
			activeCasesWhere = "1 = 0"
			activeCasesArgs = nil
			caseStageWhere = "1 = 0"
			caseStageArgs = nil
			if !services.IsConflictReviewRole(role) {
				pendingApprovalWhere = "1 = 0"
				pendingApprovalArgs = nil
			}
			inboxWhere = "1 = 0"
			inboxArgs = nil
		}
	}
	pendingApprovalRows := h.visibleApprovalRows(c, pendingApprovalWhere, pendingApprovalArgs...)
	pendingApprovalCount := int64(len(pendingApprovalRows))
	conflictWhere := "check_status IN ? OR risk_level IN ?"
	conflictArgs := []interface{}{openConflictStatuses, visibleConflictRiskLevels}
	if lawyerID, ok := currentLawyerScope(c); ok {
		conflictWhere = "(" + conflictWhere + ") AND user_id = ?"
		conflictArgs = append(conflictArgs, lawyerID)
	}
	if !canReadConflictQueue(c) {
		conflictWhere = "1 = 0"
		conflictArgs = nil
	}
	includeAllConflicts := c.Query("conflict_scope") == "all"
	riskQueueLimit := 8
	if includeAllConflicts {
		riskQueueLimit = 100
	}
	riskQueue := h.riskRows(c, riskQueueLimit, includeAllConflicts)
	clientCount := h.visibleClientCount(c)
	if lawyerID, ok := currentLawyerScope(c); ok {
		// Keep the command-center aggregate consistent with the client list:
		// a lawyer must not learn the size of another lawyer's client portfolio.
		clientCount = h.accessibleClientCount(lawyerID)
	}
	intakeCount := h.visibleIntakeCount(c)
	if isAssistantRole(c) {
		// Assistants work only with their own collaboration drafts. Global client
		// and intake counts reveal information that is unrelated to that draft.
		clientCount = 0
		intakeCount = 0
	}
	if !canViewAllMatterData(c) {
		if _, lawyerScoped := currentLawyerScope(c); !lawyerScoped {
			clientCount = 0
			intakeCount = 0
		}
	}

	common.APISuccess(c, gin.H{
		"summary": gin.H{
			"active_cases":        h.count("cases", activeCasesWhere, activeCasesArgs...),
			"clients":             clientCount,
			"pending_approvals":   pendingApprovalCount,
			"open_conflict_tasks": h.visibleConflictRecordCount(c, conflictWhere, conflictArgs...),
			"unread_inbox":        h.countAny([]string{"inbox_items"}, inboxWhere, inboxArgs...),
		},
		"workflow": gin.H{
			"intake":     intakeCount,
			"conflict":   h.visibleConflictRecordCount(c, conflictWhere, conflictArgs...),
			"approval":   pendingApprovalCount,
			"activation": h.count("cases", activeCasesWhere, activeCasesArgs...),
		},
		"todo_items":              h.todoRows(c, 10),
		"risk_queue":              riskQueue,
		"approval_queue":          h.approvalRows(c, 8),
		"case_rows":               h.caseRows(c, 20),
		"case_stage_distribution": h.groupedCounts("cases", "status", caseStageWhere, 10, caseStageArgs...),
		"risk_distribution":       h.visibleConflictRiskCounts(c, conflictWhere, 10, conflictArgs...),
		"overdue_tasks":           h.overdueRows(c, 5),
		"recent_activities":       h.activityRows(c, 10),
		"generated_at":            time.Now(),
	})
}

func (h *DemoAggregateHandler) ClientMasterProfile(c *gin.Context) {
	id := c.Param("id")
	if !h.canAccessClientProfile(c, id) {
		return
	}
	var client map[string]interface{}
	if h.first("clients", id, &client) != nil {
		common.APINotFound(c, "客户不存在", "指定客户不存在或已删除")
		return
	}
	safeClient := sanitizeClientAggregateRow(client)
	clientIDValue, _ := strconv.ParseUint(strings.TrimSpace(fmt.Sprint(client["id"])), 10, 64)
	var primaryContact *services.ClientContactResponse
	if h.contactService != nil {
		contact, contactErr := h.contactService.GetPrimaryContact(c.Request.Context(), uint(clientIDValue))
		if contactErr != nil {
			common.APIInternalServerError(c, "读取主联系人失败", contactErr.Error())
			return
		}
		primaryContact = contact
	}
	if primaryContact == nil && strings.TrimSpace(valueString(client, "contact_person")) != "" {
		primaryContact = &services.ClientContactResponse{
			ClientID:  uint(clientIDValue),
			Name:      valueString(client, "contact_person"),
			Phone:     valueString(safeClient, "contact_phone"),
			IsPrimary: true,
			Legacy:    true,
		}
	}

	actorID, _ := middleware.GetCurrentUserID(c)
	matterWhere := `client_id = ? AND deleted_at IS NULL AND (
		ethical_wall_enabled = ?
		OR EXISTS (
			SELECT 1 FROM case_ethical_wall_whitelist wall_access
			WHERE wall_access.case_id = cases.id AND wall_access.user_id = ?
		)
	)`
	matterArgs := []interface{}{id, false, actorID}
	if !canViewAllMatterData(c) {
		actorIDText, _ := currentUserIDString(c)
		matterWhere = `client_id = ? AND deleted_at IS NULL AND (lawyer_id = ? OR created_by = ?) AND (
			ethical_wall_enabled = ?
			OR EXISTS (
				SELECT 1 FROM case_ethical_wall_whitelist wall_access
				WHERE wall_access.case_id = cases.id AND wall_access.user_id = ?
			)
		)`
		matterArgs = []interface{}{id, actorIDText, actorIDText, false, actorID}
	}
	conflictWhere := "client_id = ?"
	conflictArgs := []interface{}{fmt.Sprint(client["id"])}
	if !canViewAllMatterData(c) {
		actorIDText, _ := currentUserIDString(c)
		conflictWhere += " AND user_id = ?"
		conflictArgs = append(conflictArgs, actorIDText)
	}

	common.APISuccess(c, gin.H{
		"client":           safeClient,
		"primary_contact":  primaryContact,
		"completeness":     h.clientCompleteness(safeClient),
		"related_parties":  h.clientRelatedParties(c, id, fmt.Sprint(client["name"])),
		"matter_history":   h.recentRows("cases", matterWhere, 5, matterArgs...),
		"conflict_history": h.clientConflictHistory(c, conflictWhere, conflictArgs...),
	})
}

func (h *DemoAggregateHandler) IntakeWorkbench(c *gin.Context) {
	role, _ := middleware.GetCurrentRole(c)
	if !services.CanReadCaseIntake(role) {
		common.APIForbidden(c, "无权读取接案记录", "当前账号没有接案工作台查看权限")
		return
	}
	id := c.Param("id")
	var intake map[string]interface{}
	if h.first("case_intakes", id, &intake) != nil {
		common.APINotFound(c, "接案记录不存在", "指定接案记录不存在")
		return
	}
	if !canViewAllMatterData(c) {
		actorID, ok := currentUserIDString(c)
		if !ok {
			return
		}
		if strings.TrimSpace(fmt.Sprint(intake["created_by"])) != actorID {
			forbidObjectAccess(c)
			return
		}
	}
	if clientID := intakeClientID(intake["client_id"]); clientID > 0 && !isAssistantRole(c) {
		// A management role may open the intake row by workflow permission, but
		// the selected client still has to pass the same ethical-wall boundary as
		// the client and conflict APIs. This prevents a guessed intake ID from
		// becoming a back door into a walled client context.
		if !h.authorizeIntakeClient(c, clientID) {
			return
		}
	}
	parties := h.recentRows("case_intake_parties", "intake_id = ?", 20, intake["id"])
	materials := h.recentRows("case_materials", "intake_id = ?", 20, intake["id"])
	client := sanitizeClientAggregateRow(h.firstByID("clients", intake["client_id"]))

	common.APISuccess(c, gin.H{
		"intake":    sanitizeAggregateRow("case_intakes", intake),
		"client":    client,
		"parties":   parties,
		"materials": materials,
	})
}

func (h *DemoAggregateHandler) ApprovalsWorkbench(c *gin.Context) {
	where := "deleted_at IS NULL"
	args := []interface{}{}
	waiverWhere := "status IN ?"
	waiverArgs := []interface{}{[]string{"SUBMITTED", "UNDER_REVIEW"}}
	if userID, ok := middleware.GetCurrentUserID(c); ok && !canViewAllMatterData(c) {
		where += " AND (applicant_id = ? OR current_approver_id = ?)"
		args = append(args, fmt.Sprint(userID), fmt.Sprint(userID))
		waiverWhere += " AND (created_by = ? OR assigned_reviewer = ?)"
		waiverArgs = append(waiverArgs, fmt.Sprint(userID), fmt.Sprint(userID))
	}
	rows := h.visibleApprovalRows(c, where, args...)
	waiverRows := h.visibleWaiverRows(c, waiverWhere, waiverArgs...)
	pendingCount := countApprovalAggregateRows(rows, "status", "submitted", "under_review", "resubmitted")
	revisionCount := countApprovalAggregateRows(rows, "status", "needs_revision")
	conflictCount := countApprovalAggregateRows(rows, "type", "conflict_approval")
	financeCount := countApprovalAggregateRows(rows, "type", "finance")
	common.APISuccess(c, gin.H{
		"stats": gin.H{
			"pending":        pendingCount,
			"needs_revision": revisionCount,
			"waiver_review":  int64(len(waiverRows)),
		},
		"items": approvalAggregateSummaryRows(rows, 12),
		"queues": []gin.H{
			{"key": "conflict", "label": "冲突审批", "count": conflictCount},
			{"key": "waiver", "label": "豁免评估", "count": int64(len(waiverRows))},
			{"key": "finance", "label": "财务审批", "count": financeCount},
		},
	})
}

func (h *DemoAggregateHandler) LawyersResourceCenter(c *gin.Context) {
	common.APISuccess(c, gin.H{
		"summary": gin.H{
			"lawyers":      h.count("users", "deleted_at IS NULL AND role IN ?", []string{"lawyer", "admin"}),
			"departments":  h.distinctCount("users", "department", "deleted_at IS NULL"),
			"active_cases": h.visibleCaseCount(c, "deleted_at IS NULL AND status IN ?", []interface{}{[]string{"pending", "in_progress"}}),
			// The legacy task table has no verified subject-case join. Keep the
			// resource center honest until it can use a wall-aware task query.
			"pending_tasks": int64(0),
		},
		"lawyers":     h.recentRows("users", "deleted_at IS NULL AND role IN ?", 20, []string{"lawyer", "admin"}),
		"capacity":    h.groupedCounts("users", "department", "deleted_at IS NULL AND role IN ?", 20, []string{"lawyer", "admin"}),
		"assignments": h.caseRows(c, 8),
		"tasks":       h.todoRows(c, 8),
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
			// The legacy payment aggregate has no verified case subject join.
			// Finance must provide a separate wall-aware report.
			"revenue_ytd":       float64(0),
			"active_cases":      h.visibleCaseCount(c, "deleted_at IS NULL AND status IN ?", []interface{}{[]string{"pending", "in_progress"}}),
			"new_clients_30d":   h.visibleClientCountWithWhere(c, "clients.created_at >= ?", []interface{}{time.Now().AddDate(0, 0, -30)}),
			"high_risk_matters": h.visibleConflictRecordCount(c, "risk_level IN ?", []string{"HIGH", "CRITICAL"}),
		},
		// Keep dashboard trends on the same wall-aware, minimum-field projection
		// as the case list. Returning raw case rows here would bypass the
		// response boundary used by the rest of the workbench.
		"trends": h.caseRows(c, 12),
	})
}

func (h *DemoAggregateHandler) todoRows(c *gin.Context, limit int) []gin.H {
	where := h.inboxFilter("is_completed = ?")
	args := []interface{}{false}
	if c != nil {
		if userID, ok := middleware.GetCurrentUserID(c); ok && !canViewAllMatterData(c) {
			where += " AND user_id = ?"
			args = append(args, userID)
		} else if canViewAllMatterData(c) {
			return []gin.H{}
		}
	}
	rows := h.recentRowsAny([]string{"inbox_items"}, where, limit, args...)
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

func (h *DemoAggregateHandler) riskRows(c *gin.Context, limit int, includeAll bool) []gin.H {
	if !canReadConflictQueue(c) {
		return []gin.H{}
	}
	where := "check_status IN ? OR risk_level IN ?"
	args := []interface{}{
		[]string{"QUEUED", "RUNNING", "PROCESSING"},
		[]string{"HIGH", "CRITICAL", "MEDIUM"},
	}
	if includeAll {
		where = "check_status IN ?"
		args = []interface{}{[]string{"QUEUED", "RUNNING", "PROCESSING", "COMPLETED"}}
	}
	if lawyerID, ok := currentLawyerScope(c); ok {
		where = "(" + where + ") AND user_id = ?"
		args = append(args, lawyerID)
	}
	rows := h.visibleConflictRecordRows(c, where, args...)
	if limit > 0 && len(rows) > limit {
		rows = rows[:limit]
	}
	approvalLinks := h.visibleConflictApprovalLinks(c)
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		checkID := fmt.Sprint(row["check_id"])
		searchParameters := objectValue(row["search_parameters"])
		conflictCases := []map[string]interface{}{}
		if checkID != "" && h.tableExists("conflict_cases") {
			_ = h.db.Table("conflict_cases").
				Where("check_id = ?", checkID).
				Order("CASE risk_level WHEN 'CRITICAL' THEN 4 WHEN 'HIGH' THEN 3 WHEN 'MEDIUM' THEN 2 WHEN 'LOW' THEN 1 ELSE 0 END DESC, created_at DESC").
				Find(&conflictCases).Error
		}
		primaryConflict := primaryConflictCase(conflictCases)
		checkResult := objectValue(row["check_result"])
		primaryEvidence := primaryConflictEvidence(primaryConflict, checkResult)
		matchedSubject := firstNonEmpty(
			valueString(primaryEvidence, "requestedParty"),
			valueString(primaryEvidence, "requested_party"),
			valueString(primaryEvidence, "matchedEntity"),
			valueString(primaryEvidence, "matched_entity"),
			primaryConflictSubject(primaryConflict, row),
		)
		matchedEntity := firstNonEmpty(valueString(primaryEvidence, "matchedEntity"), valueString(primaryEvidence, "matched_entity"))
		matchType := firstNonEmpty(valueString(primaryEvidence, "matchType"), valueString(primaryEvidence, "match_type"), primaryConflictValue(primaryConflict, "conflict_type"))
		matchAlgorithm := firstNonEmpty(
			valueString(primaryEvidence, "matchAlgorithm"),
			valueString(primaryEvidence, "match_algorithm"),
			valueString(objectValue(objectValue(checkResult["riskAssessment"])["matchEvidence"]), "algorithm"),
		)
		partyRole := firstNonEmpty(valueString(primaryEvidence, "partyRole"), valueString(primaryEvidence, "party_role"))
		ruleCode := firstNonEmpty(valueString(primaryEvidence, "ruleCode"), valueString(primaryEvidence, "rule_code"))
		visibleRiskLevel := row["risk_level"]
		visibleHasConflict := row["has_conflict"]
		visibleConflictCases := conflictCases
		visibleCheckResult := checkResult
		visibleSearchParameters := row["search_parameters"]
		sourceCase := primaryEvidenceSourceCase(primaryEvidence)
		evidenceSummary := primaryConflictValue(primaryConflict, "description")
		if !canReviewConflict(c) {
			visibleSearchParameters = redactConflictQueueSearchParameters(row["search_parameters"])
			if isNoConflictQueueRecord(row, checkResult, conflictCases) {
				visibleCheckResult = redactNoConflictQueueCheckResult(checkResult)
			} else if isCoverageLimitedNoEvidenceQueueRecord(row, checkResult, conflictCases) {
				visibleCheckResult = redactCoverageLimitedNoEvidenceCheckResult(checkResult)
				sourceCase = ""
				matchedSubject = ""
				matchedEntity = ""
				matchType = ""
				matchAlgorithm = ""
				partyRole = ""
				ruleCode = ""
				visibleRiskLevel = "REVIEW_REQUIRED"
				visibleHasConflict = false
				evidenceSummary = "未发现匹配记录，但检索范围受限，需独立人工复核。"
			} else {
				visibleConflictCases = redactConflictQueueCases(conflictCases)
				visibleCheckResult = redactConflictQueueCheckResult(checkResult)
				sourceCase = "受限"
				matchedSubject = ""
				matchedEntity = ""
				matchType = ""
				matchAlgorithm = ""
				partyRole = ""
				ruleCode = ""
				visibleRiskLevel = "REVIEW_REQUIRED"
				evidenceSummary = "存在受隔离记录，请联系独立冲突核查人。"
			}
		}
		item := gin.H{
			"id":                row["check_id"],
			"case_id":           firstNonEmpty(valueString(searchParameters, "subjectCaseId"), valueString(searchParameters, "subject_case_id")),
			"case_number":       firstNonEmpty(valueString(searchParameters, "subjectCaseNumber"), valueString(searchParameters, "subject_case_number")),
			"intake_id":         firstNonEmpty(valueString(searchParameters, "intakeId"), valueString(searchParameters, "intake_id")),
			"title":             row["case_name"],
			"case_type":         row["case_type"],
			"client_id":         row["client_id"],
			"client_name":       row["client_name"],
			"matched_subject":   matchedSubject,
			"matched_entity":    matchedEntity,
			"matched_type":      matchType,
			"match_type":        matchType,
			"match_algorithm":   matchAlgorithm,
			"party_role":        partyRole,
			"rule_code":         ruleCode,
			"source_case":       sourceCase,
			"evidence_summary":  evidenceSummary,
			"status":            row["check_status"],
			"risk_level":        visibleRiskLevel,
			"has_conflict":      visibleHasConflict,
			"owner":             row["user_id"],
			"duration":          row["duration"],
			"check_time":        row["check_time"],
			"created_at":        row["created_at"],
			"updated_at":        row["updated_at"],
			"search_parameters": visibleSearchParameters,
			"check_result":      jsonStringValue(visibleCheckResult),
			"conflict_cases":    visibleConflictCases,
		}
		if approval, ok := approvalLinks[checkID]; ok {
			item["approval_id"] = approval["id"]
			item["approval_status"] = approval["status"]
			item["approval_request_number"] = approval["request_number"]
			item["approval_current_approver_id"] = approval["current_approver_id"]
		}
		items = append(items, item)
	}
	return items
}

// visibleConflictApprovalLinks returns only active conflict approvals the
// current actor may open. Non-management actors are constrained to approvals
// they submitted or are currently assigned to process.
func (h *DemoAggregateHandler) visibleConflictApprovalLinks(c *gin.Context) map[string]map[string]interface{} {
	links := make(map[string]map[string]interface{})
	if c == nil || !h.tableExists("approval_requests") {
		return links
	}
	where := "deleted_at IS NULL AND type = ? AND status IN ?"
	args := []interface{}{"conflict_approval", []string{"submitted", "under_review", "resubmitted"}}
	if !canViewAllMatterData(c) {
		userID, ok := middleware.GetCurrentUserID(c)
		if !ok || userID == 0 {
			return links
		}
		where += " AND (applicant_id = ? OR current_approver_id = ?)"
		args = append(args, fmt.Sprint(userID), fmt.Sprint(userID))
	}
	for _, row := range h.visibleApprovalRows(c, where, args...) {
		checkID := firstNonEmpty(strings.TrimSpace(fmt.Sprint(row["conflict_check_id"])))
		if checkID == "" {
			continue
		}
		if _, exists := links[checkID]; !exists {
			links[checkID] = row
		}
	}
	return links
}

// isNoConflictQueueRecord identifies the only conflict result that may be
// shown to an ordinary lawyer without the restricted-hit projection. It must
// be an explicit no-match result with complete archive coverage and no hidden
// evidence; missing coverage fails closed into the review-required branch.
func isNoConflictQueueRecord(row map[string]interface{}, checkResult map[string]interface{}, conflictCases []map[string]interface{}) bool {
	if interfaceBool(row["has_conflict"]) || len(conflictCases) > 0 {
		return false
	}
	decision := objectValue(checkResult["decision"])
	decisionStatus := strings.ToUpper(firstNonEmpty(
		valueString(decision, "status"),
		valueString(checkResult, "decisionStatus"),
	))
	if decisionStatus != "NO_MATCH_FOUND" && decisionStatus != "NO_CONFLICT" && decisionStatus != "CLEAR" {
		return false
	}
	coverageStatus := strings.ToUpper(firstNonEmpty(
		valueString(decision, "coverageStatus"),
		valueString(decision, "coverage_status"),
		valueString(checkResult, "coverageStatus"),
		valueString(checkResult, "coverage_status"),
	))
	if coverageStatus != "COMPLETE" {
		return false
	}
	if interfaceBool(objectValue(checkResult["riskAssessment"])["requiresApproval"]) {
		return false
	}
	if interfaceInt(decision["evidenceCount"]) > 0 || len(mapSliceValue(checkResult["evidence"])) > 0 || len(mapSliceValue(checkResult["matchEvidence"])) > 0 {
		return false
	}
	risk := strings.ToUpper(firstNonEmpty(
		valueString(objectValue(checkResult["riskAssessment"]), "overallRisk"),
		fmt.Sprint(row["risk_level"]),
	))
	return risk == "MINIMAL" || risk == "LOW"
}

// isCoverageLimitedNoEvidenceQueueRecord distinguishes an incomplete search
// from a real restricted hit. Both remain blocked for independent review, but
// the ordinary lawyer must not be told that a hidden matter exists when the
// engine only knows that archive coverage is incomplete.
func isCoverageLimitedNoEvidenceQueueRecord(row map[string]interface{}, checkResult map[string]interface{}, conflictCases []map[string]interface{}) bool {
	if len(conflictCases) > 0 {
		return false
	}
	decision := objectValue(checkResult["decision"])
	coverageStatus := strings.ToUpper(firstNonEmpty(
		valueString(decision, "coverageStatus"),
		valueString(decision, "coverage_status"),
		valueString(checkResult, "coverageStatus"),
		valueString(checkResult, "coverage_status"),
	))
	if coverageStatus != "COVERAGE_LIMITED" {
		return false
	}
	if interfaceInt(decision["restrictedCount"]) > 0 ||
		interfaceInt(decision["evidenceCount"]) > 0 ||
		len(mapSliceValue(checkResult["evidence"])) > 0 ||
		len(mapSliceValue(checkResult["matchEvidence"])) > 0 ||
		len(mapSliceValue(checkResult["conflictCases"])) > 0 {
		return false
	}
	assessment := objectValue(checkResult["riskAssessment"])
	return len(mapSliceValue(assessment["evidence"])) == 0 &&
		len(mapSliceValue(assessment["matchEvidence"])) == 0
}

// visibleConflictRecordRows is the aggregate-layer counterpart of the
// conflict-check repository's ethical-wall predicate. The legacy record table
// stores the subject case in JSON, so filtering after loading is deliberate:
// it avoids dialect-specific JSON SQL while keeping pagination and aggregate
// counts on the same authorized set.
func (h *DemoAggregateHandler) visibleConflictRecordRows(c *gin.Context, where string, args ...interface{}) []map[string]interface{} {
	if !h.tableExists("conflict_check_records") {
		return []map[string]interface{}{}
	}
	var rows []map[string]interface{}
	query := h.db.Table("conflict_check_records")
	if where != "" {
		query = query.Where(where, args...)
	}
	if h.hasColumn("conflict_check_records", "created_at") {
		query = query.Order("created_at DESC")
	}
	if err := query.Find(&rows).Error; err != nil {
		return []map[string]interface{}{}
	}
	visible := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		if h.conflictRecordVisibleToActor(c, row) {
			visible = append(visible, sanitizeAggregateRow("conflict_check_records", row))
		}
	}
	return visible
}

func (h *DemoAggregateHandler) visibleConflictRecordCount(c *gin.Context, where string, args ...interface{}) int64 {
	return int64(len(h.visibleConflictRecordRows(c, where, args...)))
}

func (h *DemoAggregateHandler) visibleConflictRiskCounts(c *gin.Context, where string, limit int, args ...interface{}) []gin.H {
	rows := h.visibleConflictRecordRows(c, where, args...)
	counts := make(map[string]int64)
	for _, row := range rows {
		key := strings.TrimSpace(fmt.Sprint(row["risk_level"]))
		if key == "" || key == "<nil>" {
			key = "UNKNOWN"
		}
		counts[key]++
	}
	result := make([]gin.H, 0, len(counts))
	for key, count := range counts {
		result = append(result, gin.H{"key": key, "count": count})
	}
	sort.SliceStable(result, func(i, j int) bool {
		left := result[i]["count"].(int64)
		right := result[j]["count"].(int64)
		if left == right {
			return fmt.Sprint(result[i]["key"]) < fmt.Sprint(result[j]["key"])
		}
		return left > right
	})
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result
}

// conflictRecordVisibleToActor intentionally fails closed for a reviewer when
// the record has no provable subject-case context. Ownership is sufficient for
// a normal lawyer's own legacy task, but a management/reviewer queue must never
// expose a global record whose ethical-wall status cannot be checked.
func (h *DemoAggregateHandler) conflictRecordVisibleToActor(c *gin.Context, row map[string]interface{}) bool {
	if c == nil || row == nil || !canReadConflictQueue(c) {
		return false
	}
	actorID, ok := middleware.GetCurrentUserID(c)
	if !ok || actorID == 0 {
		return false
	}
	role, _ := middleware.GetCurrentRole(c)
	ownerID, ownerErr := strconv.ParseUint(strings.TrimSpace(fmt.Sprint(row["user_id"])), 10, 32)
	ownerOK := ownerErr == nil && ownerID > 0
	checkID := strings.TrimSpace(fmt.Sprint(row["check_id"]))
	caseIDText := h.conflictRecordSubjectCaseID(checkID, row["search_parameters"])
	if caseIDText == "" || caseIDText == "0" {
		intakeID := subjectIntakeIDFromAggregateSearchParameters(row["search_parameters"])
		if intakeID != "" && h.authz != nil {
			actor, actorOK := currentAuthActor(c)
			if !actorOK {
				return false
			}
			allowed, authErr := h.authz.CanReadConflictIntakeContext(c.Request.Context(), actor, intakeID, uint(ownerID), fmt.Sprint(row["client_id"]))
			if authErr != nil || !allowed {
				return false
			}
			if strings.EqualFold(role, "lawyer") || services.IsIntakeAssistantRole(role) {
				return ownerOK && uint(ownerID) == actorID
			}
			return services.IsConflictReviewRole(role)
		}
		return ownerOK && uint(ownerID) == actorID && !services.IsConflictReviewRole(role)
	}
	caseID, err := strconv.ParseUint(caseIDText, 10, 32)
	if err != nil || caseID == 0 || h.authz == nil {
		return false
	}
	actor := services.AuthActor{UserID: actorID, Role: role}
	allowed, authErr := h.authz.CanReadConflictContext(c.Request.Context(), actor, uint(caseID))
	if authErr != nil || !allowed {
		return false
	}
	if strings.EqualFold(role, "lawyer") || services.IsIntakeAssistantRole(role) {
		return ownerOK && uint(ownerID) == actorID
	}
	return services.IsConflictReviewRole(role)
}

func (h *DemoAggregateHandler) conflictRecordSubjectCaseID(checkID string, searchParameters interface{}) string {
	if caseID := subjectCaseIDFromAggregateSearchParameters(searchParameters); caseID != "" {
		return caseID
	}
	if checkID == "" || !h.tableExists("cases") || !h.hasColumn("cases", "conflict_check_id") {
		return ""
	}
	var caseIDs []uint
	if err := h.db.Table("cases").Where("conflict_check_id = ? AND deleted_at IS NULL", checkID).Limit(2).Pluck("id", &caseIDs).Error; err != nil {
		return ""
	}
	if len(caseIDs) != 1 {
		if len(caseIDs) > 1 {
			return "ambiguous"
		}
		return ""
	}
	return strconv.FormatUint(uint64(caseIDs[0]), 10)
}

func redactConflictQueueSearchParameters(value interface{}) interface{} {
	parameters := objectValue(value)
	if len(parameters) == 0 {
		return map[string]interface{}{}
	}
	serialized, err := json.Marshal(parameters)
	if err != nil {
		return map[string]interface{}{}
	}
	var projected map[string]interface{}
	if err := json.Unmarshal(serialized, &projected); err != nil {
		return map[string]interface{}{}
	}
	for _, key := range []string{"clientIdentifiers", "client_identifiers", "identifiers"} {
		if projected[key] != nil {
			projected[key] = redactedIdentifierNotice()
		}
	}
	for _, key := range []string{"parties", "otherParties", "other_parties"} {
		items := mapSliceValue(projected[key])
		if len(items) == 0 {
			continue
		}
		for _, item := range items {
			redactIdentifierFields(item)
		}
		projected[key] = items
	}
	return projected
}

// redactConflictQueueCases applies the same disclosure boundary as the task
// result endpoint. The dashboard is a separate API surface and must never
// rely on the browser to hide historical matter details.
func redactConflictQueueCases(conflictCases []map[string]interface{}) []map[string]interface{} {
	if len(conflictCases) == 0 {
		return []map[string]interface{}{}
	}
	result := make([]map[string]interface{}, 0, len(conflictCases))
	for _, conflictCase := range conflictCases {
		result = append(result, redactConflictQueueCase(conflictCase))
	}
	return result
}

func redactConflictQueueCheckResult(checkResult map[string]interface{}) map[string]interface{} {
	if checkResult == nil {
		return map[string]interface{}{}
	}
	serialized, err := json.Marshal(checkResult)
	if err != nil {
		return map[string]interface{}{}
	}
	var projected map[string]interface{}
	if err := json.Unmarshal(serialized, &projected); err != nil {
		return map[string]interface{}{}
	}
	for _, key := range []string{"conflictCases", "conflict_cases"} {
		cases := mapSliceValue(projected[key])
		if len(cases) == 0 {
			continue
		}
		redacted := make([]interface{}, 0, len(cases))
		for _, conflictCase := range cases {
			redacted = append(redacted, redactConflictQueueCase(conflictCase))
		}
		projected[key] = redacted
	}
	projected["review"] = nil
	for _, key := range []string{"evidence", "matchEvidence"} {
		if evidence := mapSliceValue(projected[key]); len(evidence) > 0 {
			redacted := make([]interface{}, 0, len(evidence))
			for _, item := range evidence {
				redacted = append(redacted, redactConflictQueueEvidence(item))
			}
			projected[key] = redacted
		}
	}
	for _, key := range []string{"clientIdentifiers", "client_identifiers", "identifiers"} {
		if projected[key] != nil {
			projected[key] = redactedIdentifierNotice()
		}
	}
	for _, key := range []string{"normalizedSubjects", "normalized_subjects"} {
		subjects := mapSliceValue(projected[key])
		if len(subjects) == 0 {
			continue
		}
		for _, subject := range subjects {
			redactIdentifierFields(subject)
		}
		projected[key] = subjects
	}
	// Keep legacy dashboard records conservative until an independent review or
	// an approved waiver is present. The UI derives the final display from this
	// status plus the review/waiver objects.
	decision := objectValue(projected["decision"])
	decision["status"] = "REVIEW_REQUIRED"
	decision["recommendation"] = "存在受隔离记录，请联系独立冲突核查人；确认前不得视为已确认冲突或无冲突。"
	decision["requiresManualReview"] = true
	decision["ruleCodes"] = []interface{}{}
	decision["evidenceCount"] = 0
	decision["restrictedCount"] = len(mapSliceValue(projected["conflictCases"]))
	projected["decision"] = decision
	projected["riskAssessment"] = map[string]interface{}{
		"overallRisk":      "REVIEW_REQUIRED",
		"riskScore":        0,
		"riskReason":       "存在受隔离记录，需独立冲突核查人处理",
		"requiresApproval": true,
		"riskFactors":      []interface{}{"存在受隔离记录"},
		"mitigation":       []interface{}{"联系独立冲突核查人"},
	}
	projected["recommendations"] = []interface{}{"存在受隔离记录，请联系独立冲突核查人。"}
	return projected
}

func redactNoConflictQueueCheckResult(checkResult map[string]interface{}) map[string]interface{} {
	if checkResult == nil {
		return map[string]interface{}{}
	}
	serialized, err := json.Marshal(checkResult)
	if err != nil {
		return map[string]interface{}{}
	}
	var projected map[string]interface{}
	if err := json.Unmarshal(serialized, &projected); err != nil {
		return map[string]interface{}{}
	}
	for _, key := range []string{"clientIdentifiers", "client_identifiers", "identifiers"} {
		if projected[key] != nil {
			projected[key] = redactedIdentifierNotice()
		}
	}
	for _, key := range []string{"normalizedSubjects", "normalized_subjects"} {
		subjects := mapSliceValue(projected[key])
		for _, subject := range subjects {
			redactIdentifierFields(subject)
		}
		if len(subjects) > 0 {
			projected[key] = subjects
		}
	}
	return projected
}

func redactCoverageLimitedNoEvidenceCheckResult(checkResult map[string]interface{}) map[string]interface{} {
	projected := redactNoConflictQueueCheckResult(checkResult)
	decision := objectValue(projected["decision"])
	decision["status"] = "REVIEW_REQUIRED"
	decision["requiresManualReview"] = true
	decision["evidenceCount"] = 0
	decision["recommendation"] = "未发现匹配记录，但检索范围受限；请由独立冲突核查人补充核查。"
	projected["decision"] = decision
	projected["riskAssessment"] = map[string]interface{}{
		"overallRisk":      "REVIEW_REQUIRED",
		"riskScore":        nil,
		"riskReason":       "未发现匹配记录，但权威档案覆盖完整性尚未确认",
		"requiresApproval": true,
		"riskFactors":      []interface{}{"检索范围受限"},
		"mitigation":       []interface{}{"由独立冲突核查人补充核查"},
	}
	projected["recommendations"] = []interface{}{"未发现匹配记录，但检索范围受限，不能据此确认无冲突。"}
	return projected
}

func redactedIdentifierNotice() map[string]interface{} {
	return map[string]interface{}{"notice": "完整身份标识仅向获授权冲突核查人显示"}
}

func redactIdentifierFields(value map[string]interface{}) {
	for _, key := range []string{"identifiers", "clientIdentifiers", "client_identifiers"} {
		if value[key] != nil {
			value[key] = redactedIdentifierNotice()
		}
	}
}

func redactConflictQueueCase(conflictCase map[string]interface{}) map[string]interface{} {
	if conflictCase == nil {
		return map[string]interface{}{}
	}
	serialized, err := json.Marshal(conflictCase)
	if err != nil {
		return map[string]interface{}{}
	}
	var projected map[string]interface{}
	if err := json.Unmarshal(serialized, &projected); err != nil {
		return map[string]interface{}{}
	}
	for _, key := range []string{"caseId", "case_id"} {
		projected[key] = ""
	}
	for _, key := range []string{"caseNo", "case_no"} {
		projected[key] = "受限"
	}
	for _, key := range []string{"caseName", "case_name"} {
		projected[key] = "受限历史事项"
	}
	for _, key := range []string{"caseType", "case_type"} {
		projected[key] = ""
	}
	for _, key := range []string{"id", "checkId", "check_id", "clientId", "client_id", "caseStatus", "case_status", "createdAt", "created_at", "riskLevel", "risk_level", "matchType", "match_type", "ruleCode", "rule_code"} {
		projected[key] = ""
	}
	projected["conflictType"] = "受限记录"
	projected["conflict_type"] = "受限记录"
	projected["opposingParties"] = []interface{}{}
	projected["opposing_parties"] = []interface{}{}
	projected["requiresManualReview"] = true
	projected["restricted"] = true
	for _, key := range []string{"description", "conflictDetails", "conflict_details"} {
		projected[key] = "存在受隔离记录，请联系独立冲突核查人。"
	}
	for _, key := range []string{"evidence", "matchEvidence"} {
		if evidence := mapSliceValue(projected[key]); len(evidence) > 0 {
			redacted := make([]interface{}, 0, len(evidence))
			for _, item := range evidence {
				redacted = append(redacted, redactConflictQueueEvidence(item))
			}
			projected[key] = redacted
		}
	}
	return projected
}

func redactConflictQueueEvidence(evidence map[string]interface{}) map[string]interface{} {
	if evidence == nil {
		return map[string]interface{}{}
	}
	serialized, err := json.Marshal(evidence)
	if err != nil {
		return map[string]interface{}{}
	}
	var projected map[string]interface{}
	if err := json.Unmarshal(serialized, &projected); err != nil {
		return map[string]interface{}{}
	}
	for _, key := range []string{"sourceCaseId", "source_case_id", "sourceCaseNumber", "source_case_number", "sourceCaseName", "source_case_name", "lawyerName", "lawyer_name"} {
		projected[key] = ""
	}
	for _, key := range []string{"evidenceId", "evidence_id", "ruleCode", "rule_code", "matchType", "match_type", "sourceType", "source_type", "requestedParty", "requested_party", "matchedEntity", "matched_entity", "partyRole", "party_role", "historicalRole", "historical_role", "sourceUpdatedAt", "source_updated_at", "lawyerName", "lawyer_name"} {
		projected[key] = ""
	}
	projected["summary"] = "存在受隔离记录，请联系独立冲突核查人。"
	projected["restricted"] = true
	return projected
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

func primaryConflictEvidence(conflictCase map[string]interface{}, checkResult map[string]interface{}) map[string]interface{} {
	structuredCases := mapSliceValue(checkResult["conflictCases"])
	primaryID := firstNonEmpty(primaryConflictValue(conflictCase, "id"), primaryConflictValue(conflictCase, "case_id"))
	for _, candidate := range structuredCases {
		candidateID := firstNonEmpty(valueString(candidate, "id"), valueString(candidate, "caseId"), valueString(candidate, "case_id"))
		if primaryID != "" && candidateID == primaryID {
			if evidence := firstEvidence(candidate["evidence"]); evidence != nil {
				return evidence
			}
		}
	}
	for _, candidate := range structuredCases {
		if evidence := firstEvidence(candidate["evidence"]); evidence != nil {
			return evidence
		}
	}
	return firstEvidence(checkResult["evidence"])
}

func firstEvidence(value interface{}) map[string]interface{} {
	items, ok := value.([]interface{})
	if !ok || len(items) == 0 {
		return nil
	}
	evidence, _ := items[0].(map[string]interface{})
	return evidence
}

func primaryEvidenceSourceCase(evidence map[string]interface{}) string {
	return firstNonEmpty(
		valueString(evidence, "sourceCaseName"), valueString(evidence, "source_case_name"),
		valueString(evidence, "sourceCaseNumber"), valueString(evidence, "source_case_number"),
		valueString(evidence, "sourceCaseId"), valueString(evidence, "source_case_id"),
	)
}

func primaryConflictSubject(conflictCase map[string]interface{}, fallback map[string]interface{}) string {
	description := primaryConflictValue(conflictCase, "description")
	if subject := extractQuotedSubject(description, "当前对方当事人"); subject != "" {
		return subject
	}
	if subject := firstNonEmpty(primaryConflictValue(conflictCase, "client_name")); subject != "" {
		return subject
	}
	return strings.TrimSpace(fmt.Sprint(fallback["client_name"]))
}

func extractQuotedSubject(text string, label string) string {
	for _, quote := range [][2]string{{"'", "'"}, {"\"", "\""}, {"“", "”"}, {"‘", "’"}} {
		prefix := label + " " + quote[0]
		start := strings.Index(text, prefix)
		if start < 0 {
			prefix = label + quote[0]
			start = strings.Index(text, prefix)
		}
		if start < 0 {
			continue
		}
		remaining := text[start+len(prefix):]
		if end := strings.Index(remaining, quote[1]); end >= 0 {
			return strings.TrimSpace(remaining[:end])
		}
	}
	return ""
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

// visibleApprovalRows applies the same object-level conflict context check to
// approval aggregates that the dedicated approval detail endpoint applies.
// Loading the rows before sanitizing is intentional: conflict_check_id and
// metadata are authorization inputs, never response fields.
func (h *DemoAggregateHandler) visibleApprovalRows(c *gin.Context, where string, args ...interface{}) []map[string]interface{} {
	if !h.tableExists("approval_requests") {
		return []map[string]interface{}{}
	}
	var rows []map[string]interface{}
	query := h.db.Table("approval_requests")
	if where != "" {
		query = query.Where(where, args...)
	}
	if h.hasColumn("approval_requests", "created_at") {
		query = query.Order("created_at DESC")
	}
	if err := query.Find(&rows).Error; err != nil {
		return []map[string]interface{}{}
	}
	visible := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		if h.approvalRowVisibleToActor(c, row) {
			visible = append(visible, sanitizeAggregateRow("approval_requests", row))
		}
	}
	return visible
}

func (h *DemoAggregateHandler) approvalRowVisibleToActor(c *gin.Context, row map[string]interface{}) bool {
	if c == nil || row == nil {
		return false
	}
	actorID, ok := middleware.GetCurrentUserID(c)
	if !ok || actorID == 0 {
		return false
	}
	checkID := firstNonEmpty(
		strings.TrimSpace(fmt.Sprint(row["conflict_check_id"])),
	)
	metadata := integrationJSONMap(row["metadata"])
	if checkID == "" {
		checkID = firstNonEmpty(
			integrationMetadataString(metadata, "conflict_check_id"),
			integrationMetadataString(metadata, "conflict_task_id"),
		)
	}
	isConflictApproval := strings.EqualFold(strings.TrimSpace(fmt.Sprint(row["type"])), "conflict_approval") || checkID != ""
	if !isConflictApproval {
		// The base query already scopes ordinary approvals to the viewer where
		// required. Only conflict-bound records need the stricter case check.
		return true
	}
	// Applicant/current-approver identity is not an ethical-wall grant. The
	// underlying case context must still be resolved and authorized below.
	return h.conflictContextVisibleToActor(c, checkID)
}

func (h *DemoAggregateHandler) conflictContextVisibleToActor(c *gin.Context, checkID string) bool {
	if c == nil || strings.TrimSpace(checkID) == "" || h.authz == nil || !h.tableExists("conflict_check_records") {
		return false
	}
	var record map[string]interface{}
	if err := h.db.Table("conflict_check_records").Select("search_parameters, user_id, client_id").Where("check_id = ?", checkID).Take(&record).Error; err != nil {
		return false
	}
	actorID, ok := middleware.GetCurrentUserID(c)
	if !ok || actorID == 0 {
		return false
	}
	role, _ := middleware.GetCurrentRole(c)
	actor := services.AuthActor{UserID: actorID, Role: role}
	caseIDText := h.conflictRecordSubjectCaseID(checkID, record["search_parameters"])
	if strings.EqualFold(caseIDText, "ambiguous") {
		return false
	}
	var allowed bool
	var authErr error
	if caseID, err := strconv.ParseUint(caseIDText, 10, 32); err == nil && caseID > 0 {
		allowed, authErr = h.authz.CanReadConflictContext(c.Request.Context(), actor, uint(caseID))
	} else {
		intakeID := subjectIntakeIDFromAggregateSearchParameters(record["search_parameters"])
		if intakeID == "" {
			return false
		}
		ownerID, _ := strconv.ParseUint(strings.TrimSpace(fmt.Sprint(record["user_id"])), 10, 32)
		allowed, authErr = h.authz.CanReadConflictIntakeContext(c.Request.Context(), actor, intakeID, uint(ownerID), fmt.Sprint(record["client_id"]))
	}
	return authErr == nil && allowed
}

func (h *DemoAggregateHandler) visibleWaiverRows(c *gin.Context, where string, args ...interface{}) []map[string]interface{} {
	if !h.tableExists("waiver_applications") {
		return []map[string]interface{}{}
	}
	var rows []map[string]interface{}
	query := h.db.Table("waiver_applications")
	if where != "" {
		query = query.Where(where, args...)
	}
	if h.hasColumn("waiver_applications", "created_at") {
		query = query.Order("created_at DESC")
	}
	if err := query.Find(&rows).Error; err != nil {
		return []map[string]interface{}{}
	}
	visible := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		if h.waiverRowVisibleToActor(c, row) {
			visible = append(visible, sanitizeAggregateRow("waiver_applications", row))
		}
	}
	return visible
}

func (h *DemoAggregateHandler) waiverRowVisibleToActor(c *gin.Context, row map[string]interface{}) bool {
	if c == nil || row == nil {
		return false
	}
	actorID, ok := middleware.GetCurrentUserID(c)
	if !ok || actorID == 0 {
		return false
	}
	actorIDText := strconv.FormatUint(uint64(actorID), 10)
	if actorIDText == strings.TrimSpace(fmt.Sprint(row["created_by"])) || actorIDText == strings.TrimSpace(fmt.Sprint(row["assigned_reviewer"])) || actorIDText == strings.TrimSpace(fmt.Sprint(row["lawyer_id"])) {
		return true
	}
	return h.conflictContextVisibleToActor(c, strings.TrimSpace(fmt.Sprint(row["conflict_check_id"])))
}

func countApprovalAggregateRows(rows []map[string]interface{}, key string, values ...string) int64 {
	wanted := make(map[string]struct{}, len(values))
	for _, value := range values {
		wanted[strings.ToLower(strings.TrimSpace(value))] = struct{}{}
	}
	var count int64
	for _, row := range rows {
		if _, ok := wanted[strings.ToLower(strings.TrimSpace(fmt.Sprint(row[key])))]; ok {
			count++
		}
	}
	return count
}

func approvalAggregateSummaryRows(rows []map[string]interface{}, limit int) []gin.H {
	if limit > 0 && len(rows) > limit {
		rows = rows[:limit]
	}
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

func (h *DemoAggregateHandler) approvalRows(c *gin.Context, limit int) []gin.H {
	where := "deleted_at IS NULL AND status IN ?"
	args := []interface{}{[]string{"submitted", "under_review", "resubmitted"}}
	if c != nil {
		if userID, ok := middleware.GetCurrentUserID(c); ok && !canViewAllMatterData(c) {
			where += " AND (applicant_id = ? OR current_approver_id = ?)"
			args = append(args, fmt.Sprint(userID), fmt.Sprint(userID))
		}
	}
	return approvalAggregateSummaryRows(h.visibleApprovalRows(c, where, args...), limit)
}

func currentLawyerScope(c *gin.Context) (uint, bool) {
	role, _ := middleware.GetCurrentRole(c)
	if role != "lawyer" && !services.IsIntakeAssistantRole(role) {
		return 0, false
	}
	return middleware.GetCurrentUserID(c)
}

func (h *DemoAggregateHandler) accessibleClientCount(lawyerID uint) int64 {
	if !h.tableExists("cases") {
		return 0
	}
	var total int64
	_ = h.db.Table("cases").
		Where(`deleted_at IS NULL AND client_id IS NOT NULL AND (lawyer_id = ? OR created_by = ?)
			AND (
				ethical_wall_enabled = ?
				OR EXISTS (
					SELECT 1 FROM case_ethical_wall_whitelist wall_access
					WHERE wall_access.case_id = cases.id AND wall_access.user_id = ?
				)
			)`, lawyerID, fmt.Sprint(lawyerID), false, lawyerID).
		Distinct("client_id").
		Count(&total).Error
	return total
}

// visibleClientCount is the management-side client aggregate for the command
// center. Management roles can operate firm-wide, but an ethical wall still
// removes protected clients unless the current actor is explicitly whitelisted
// on every protected matter that would make the client visible.
func (h *DemoAggregateHandler) visibleClientCount(c *gin.Context) int64 {
	return h.visibleClientCountWithWhere(c, "", nil)
}

func (h *DemoAggregateHandler) visibleClientCountWithWhere(c *gin.Context, where string, args []interface{}) int64 {
	if !h.tableExists("clients") || !h.tableExists("cases") || !h.tableExists("case_ethical_wall_whitelist") ||
		!h.hasColumn("cases", "client_id") || !h.hasColumn("cases", "ethical_wall_enabled") {
		return 0
	}
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok || userID == 0 {
		return 0
	}
	query := h.db.Table("clients AS clients").Where("clients.deleted_at IS NULL")
	if where != "" {
		query = query.Where(where, args...)
	}
	query = query.Where(`NOT EXISTS (
		SELECT 1 FROM cases protected_cases
		WHERE protected_cases.client_id = clients.id
		  AND protected_cases.deleted_at IS NULL
		  AND protected_cases.ethical_wall_enabled = ?
		  AND NOT EXISTS (
			SELECT 1 FROM case_ethical_wall_whitelist wall_access
			WHERE wall_access.case_id = protected_cases.id AND wall_access.user_id = ?
		  )
	)`, true, userID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0
	}
	return total
}

func (h *DemoAggregateHandler) visibleCaseCount(c *gin.Context, where string, args []interface{}) int64 {
	if !h.tableExists("cases") || !h.tableExists("case_ethical_wall_whitelist") || !h.hasColumn("cases", "ethical_wall_enabled") {
		return 0
	}
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok || userID == 0 {
		return 0
	}
	query := h.db.Table("cases").Where(where, args...)
	if lawyerID, scoped := currentLawyerScope(c); scoped {
		query = query.Where("(lawyer_id = ? OR created_by = ?)", lawyerID, fmt.Sprint(lawyerID))
	}
	query = query.Where(`(
		ethical_wall_enabled = ?
		OR EXISTS (
			SELECT 1 FROM case_ethical_wall_whitelist wall_access
			WHERE wall_access.case_id = cases.id AND wall_access.user_id = ?
		)
	)`, false, userID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0
	}
	return total
}

// visibleIntakeCount keeps the command-center intake badge consistent with
// the object-level intake boundary. A lawyer/assistant only counts their own
// drafts; management counts unprotected or explicitly whitelisted contexts.
func (h *DemoAggregateHandler) visibleIntakeCount(c *gin.Context) int64 {
	if !h.tableExists("case_intakes") || !h.hasColumn("case_intakes", "status") ||
		!h.tableExists("cases") || !h.tableExists("case_ethical_wall_whitelist") ||
		!h.hasColumn("cases", "client_id") || !h.hasColumn("cases", "ethical_wall_enabled") {
		return 0
	}
	actorID, ok := middleware.GetCurrentUserID(c)
	if !ok || actorID == 0 {
		return 0
	}
	query := h.db.Table("case_intakes").Where("status IN ?", []string{"draft", "materials_pending", "conflict_ready", "conflict_checking"})
	if lawyerID, scoped := currentLawyerScope(c); scoped && h.hasColumn("case_intakes", "created_by") {
		query = query.Where("created_by = ?", fmt.Sprint(lawyerID))
	} else if !canViewAllMatterData(c) {
		return 0
	}
	if h.hasColumn("case_intakes", "client_id") {
		query = query.Where(`(
			case_intakes.client_id IS NULL
			OR case_intakes.client_id = 0
			OR NOT EXISTS (
				SELECT 1 FROM cases protected_cases
				WHERE protected_cases.client_id = case_intakes.client_id
				  AND protected_cases.deleted_at IS NULL
				  AND protected_cases.ethical_wall_enabled = ?
				  AND NOT EXISTS (
					SELECT 1 FROM case_ethical_wall_whitelist wall_access
					WHERE wall_access.case_id = protected_cases.id AND wall_access.user_id = ?
				  )
			)
		)`, true, actorID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0
	}
	return total
}

func (h *DemoAggregateHandler) caseRows(c *gin.Context, limit int) []gin.H {
	if !h.tableExists("cases") {
		return []gin.H{}
	}
	role, _ := middleware.GetCurrentRole(c)
	if services.IsTechnicalAdminRole(role) || (services.IsConflictReviewRole(role) && !services.IsBusinessMatterManagementRole(role)) {
		// Conflict officers use the conflict queue. A general case list would
		// disclose unrelated matter names and client identities. Technical
		// administrators have the same boundary unless separately assigned a
		// business role.
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
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok || userID == 0 {
		return []gin.H{}
	}
	q = q.Where(`(
		c.ethical_wall_enabled = ?
		OR EXISTS (
			SELECT 1 FROM case_ethical_wall_whitelist wall_access
			WHERE wall_access.case_id = c.id AND wall_access.user_id = ?
		)
	)`, false, userID)
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

func (h *DemoAggregateHandler) overdueRows(c *gin.Context, limit int) []gin.H {
	if canViewAllMatterData(c) {
		return []gin.H{}
	}
	where := h.inboxFilter("is_completed = ? AND due_date < ?")
	args := []interface{}{false, time.Now()}
	if userID, ok := middleware.GetCurrentUserID(c); ok && !canViewAllMatterData(c) {
		where += " AND user_id = ?"
		args = append(args, userID)
	}
	rows := h.recentRowsAny([]string{"inbox_items"}, where, limit, args...)
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

func (h *DemoAggregateHandler) activityRows(c *gin.Context, limit int) []gin.H {
	if !canViewAllMatterData(c) {
		return []gin.H{}
	}
	rows := h.recentRowsAny([]string{"risk_audit_events"}, "", limit)
	if len(rows) == 0 {
		// Approval rows may contain conflict titles and applicant identities.
		// The fallback must use the same conflict-context filter as the approval
		// workbench instead of reading the table directly.
		rows = h.visibleApprovalRows(c, "deleted_at IS NULL")
		if limit > 0 && len(rows) > limit {
			rows = rows[:limit]
		}
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

func canReadConflictQueue(c *gin.Context) bool {
	role, _ := middleware.GetCurrentRole(c)
	return strings.EqualFold(role, "lawyer") || services.IsConflictReviewRole(role)
}

func isAssistantRole(c *gin.Context) bool {
	role, _ := middleware.GetCurrentRole(c)
	return services.IsIntakeAssistantRole(role)
}

func (h *DemoAggregateHandler) canAccessClientProfile(c *gin.Context, clientID string) bool {
	if h.authz != nil {
		parsedClientID, err := strconv.ParseUint(strings.TrimSpace(clientID), 10, 32)
		if err != nil || parsedClientID == 0 {
			common.APIBadRequest(c, "客户编号无效", "客户编号必须为有效数字")
			return false
		}
		actor, ok := currentAuthActor(c)
		if !ok {
			return false
		}
		allowed, err := h.authz.CanReadClient(c.Request.Context(), actor, uint(parsedClientID))
		if err != nil {
			common.APIInternalServerError(c, "权限校验失败", err.Error())
			return false
		}
		if !allowed {
			forbidObjectAccess(c)
			return false
		}
		return true
	}

	// Compatibility path for isolated aggregate unit tests. Production wiring
	// always supplies the centralized authorization service above.
	actorID, ok := currentUserIDString(c)
	if !ok {
		return false
	}
	actorNumericID, _ := middleware.GetCurrentUserID(c)
	query := h.db.Table("cases").Where(`client_id = ? AND deleted_at IS NULL AND (
		ethical_wall_enabled = ?
		OR EXISTS (
			SELECT 1 FROM case_ethical_wall_whitelist wall_access
			WHERE wall_access.case_id = cases.id AND wall_access.user_id = ?
		)
	)`, clientID, false, actorNumericID)
	if !canViewAllMatterData(c) {
		query = query.Where("(lawyer_id = ? OR created_by = ?)", actorID, actorID)
	}
	var count int64
	err := query.Count(&count).Error
	if err != nil {
		common.APIInternalServerError(c, "权限校验失败", err.Error())
		return false
	}
	if count == 0 {
		forbidObjectAccess(c)
		return false
	}
	return true
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
	identityCiphertext := firstNonEmpty(valueString(client, "identity_number_ciphertext"), valueString(client, "id_card_ciphertext"))
	identityDigest := firstNonEmpty(valueString(client, "identity_number_digest"), valueString(client, "id_card_digest"))
	identityComplete := identityCiphertext != "" && identityDigest != ""
	identityLabel := "身份证件号码"
	if strings.EqualFold(strings.TrimSpace(valueString(client, "type")), "企业") {
		identityLabel = "统一社会信用代码"
	}
	identityStatus := "complete"
	if !identityComplete {
		missing = append(missing, "protected_identity")
		identityStatus = "missing"
	}
	checks = append(checks, gin.H{"key": "protected_identity", "label": identityLabel, "status": identityStatus})
	score := 0
	totalRequired := len(required) + 1
	if totalRequired > 0 {
		score = (totalRequired - len(missing)) * 100 / totalRequired
	}
	return gin.H{
		"score":                    score,
		"missing_fields":           missing,
		"ready_for_conflict_check": len(missing) == 0,
		"checks":                   checks,
	}
}

func (h *DemoAggregateHandler) clientRelatedParties(c *gin.Context, clientID, clientName string) []gin.H {
	if !h.tableExists("case_intake_parties") || !h.tableExists("case_intakes") {
		return []gin.H{}
	}
	var rows []map[string]interface{}
	query := h.db.Table("case_intake_parties AS p").
		Select("p.entity_name AS name, p.party_role AS relationship_type, p.relation_depth AS depth, p.metadata").
		Joins("JOIN case_intakes ci ON ci.id = p.intake_id").
		Where("ci.client_id = ? AND p.entity_name <> ?", clientID, clientName)
	if actorID, ok := currentUserIDString(c); ok {
		actorNumericID, _ := middleware.GetCurrentUserID(c)
		query = query.Where(`(
			(p.case_id IS NULL AND ci.created_by = ?)
			OR (p.case_id IS NOT NULL AND EXISTS (
				SELECT 1 FROM cases party_cases
				WHERE party_cases.id = p.case_id
				  AND party_cases.deleted_at IS NULL
				  AND (
					  party_cases.ethical_wall_enabled = ?
					  OR EXISTS (
						  SELECT 1 FROM case_ethical_wall_whitelist wall_access
						  WHERE wall_access.case_id = party_cases.id AND wall_access.user_id = ?
					  )
				  )
				  AND (? OR party_cases.lawyer_id = ? OR party_cases.created_by = ?)
			))
		)`, actorID, false, actorNumericID, canViewAllMatterData(c), actorNumericID, actorID)
	}
	_ = query.Order("p.created_at DESC").Limit(20).Scan(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, gin.H{
			"name":              row["name"],
			"relationship_type": row["relationship_type"],
			"depth":             row["depth"],
			"metadata":          redactAggregateSensitiveValue(row["metadata"]),
		})
	}
	return items
}

func (h *DemoAggregateHandler) clientConflictHistory(c *gin.Context, where string, args ...interface{}) []map[string]interface{} {
	rows := h.visibleConflictRecordRows(c, where, args...)
	if len(rows) > 5 {
		rows = rows[:5]
	}
	if canReviewConflict(c) {
		return rows
	}
	for _, row := range rows {
		if result := objectValue(row["check_result"]); len(result) > 0 {
			row["check_result"] = redactConflictQueueCheckResult(result)
		}
		row["search_parameters"] = redactConflictQueueSearchParameters(row["search_parameters"])
		delete(row, "conflict_cases")
	}
	return rows
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
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	if !services.CanCreateCaseIntake(actor.Role) {
		common.APIForbidden(c, "无权创建接案记录", "当前账号没有接案创建权限")
		return
	}
	createdBy := strconv.FormatUint(uint64(actor.UserID), 10)
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}
	idempotencyKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	if len(idempotencyKey) > 120 {
		common.APIBadRequest(c, "幂等键无效", "Idempotency-Key 长度不能超过 120 个字符")
		return
	}
	if idempotencyKey != "" && h.hasColumn("case_intakes", "idempotency_key") {
		var existing map[string]interface{}
		if err := h.db.Table("case_intakes").
			Where("idempotency_key = ? AND created_by = ?", idempotencyKey, createdBy).
			Take(&existing).Error; err == nil {
			common.APISuccess(c, existing)
			return
		}
	}
	parties := mapSliceValue(payload["parties"])
	materials := mapSliceValue(payload["materials"])
	if containsPlaintextIdentityField(payload["metadata"]) {
		common.APIBadRequest(c, "身份信息登记方式不受支持", "身份证号、统一社会信用代码等身份标识必须登记到受保护的客户或主体档案，不能写入接案备注")
		return
	}
	for _, party := range parties {
		if containsPlaintextIdentityField(party["metadata"]) {
			common.APIBadRequest(c, "身份信息登记方式不受支持", "对方或关联方身份标识必须登记到受保护的主体档案，不能写入接案备注")
			return
		}
	}
	for _, material := range materials {
		if containsPlaintextIdentityField(material["metadata"]) {
			common.APIBadRequest(c, "身份信息登记方式不受支持", "材料备注不能包含未保护的身份标识")
			return
		}
	}
	intakePayload := filterMap(payload, allowedCaseIntakeCreateFields)
	if isAssistantRole(c) {
		if payload["client_id"] != nil || payload["parties"] != nil {
			common.APIForbidden(c, "助理协作草稿不能录入当事人身份信息", "客户、对方和关联方必须由执业律师确认后录入")
			return
		}
		intakePayload = filterMap(payload, allowedAssistantDraftFields)
		intakePayload["status"] = "assistant_draft"
		intakePayload["metadata"] = jsonStringValue(map[string]interface{}{
			"assistant_draft": true,
			"created_by_role": "assistant",
		})
	} else if value, present := payload["client_id"]; present && value != nil && strings.TrimSpace(fmt.Sprint(value)) != "" {
		if !h.authorizeIntakeClient(c, intakeClientID(value)) {
			return
		}
	}
	if !isAssistantRole(c) && len(parties) > 0 && !h.intakePartyIdentitySchemaReady() {
		common.NewAPIError(c, http.StatusServiceUnavailable, "INTAKE_PARTY_IDENTITY_SCHEMA_REQUIRED", "接案当事人身份保护字段尚未迁移，已阻止保存；请执行数据库迁移 000070")
		return
	}
	now := time.Now()
	intakeID := uuid.NewString()
	intakePayload["id"] = intakeID
	intakePayload["intake_code"] = fmt.Sprintf("INT-%s", now.Format("20060102150405"))
	intakePayload["status"] = stringValue(intakePayload["status"], "draft")
	intakePayload["priority"] = stringValue(intakePayload["priority"], "medium")
	intakePayload["metadata"] = jsonStringValue(intakePayload["metadata"])
	intakePayload["created_by"] = createdBy
	if idempotencyKey != "" && h.hasColumn("case_intakes", "idempotency_key") {
		intakePayload["idempotency_key"] = idempotencyKey
	}
	intakePayload["created_at"] = now
	intakePayload["updated_at"] = now

	if h.tableExists("case_intakes") {
		err := h.db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Table("case_intakes").Create(intakePayload).Error; err != nil {
				return err
			}
			if h.tableExists("case_intake_parties") {
				for _, party := range parties {
					row, err := prepareCaseIntakePartyRow(party, intakeID, now)
					if err != nil {
						return err
					}
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
			if idempotencyKey != "" && h.hasColumn("case_intakes", "idempotency_key") {
				var existing map[string]interface{}
				if findErr := h.db.Table("case_intakes").
					Where("idempotency_key = ? AND created_by = ?", idempotencyKey, createdBy).
					Take(&existing).Error; findErr == nil {
					common.APISuccess(c, existing)
					return
				}
			}
			if isSubjectWorkflowError(err) {
				writeSubjectWorkflowError(c, err)
				return
			}
			common.APIInternalServerError(c, "创建接案失败", err.Error())
			return
		}
	}
	safeParties := make([]map[string]interface{}, 0, len(parties))
	for _, party := range parties {
		safeParties = append(safeParties, sanitizeAggregateRow("case_intake_parties", party))
	}
	intakePayload["parties"] = safeParties
	intakePayload["materials"] = materials
	c.JSON(http.StatusCreated, common.APIResponse{Success: true, Data: intakePayload, Meta: common.ResponseMeta{Timestamp: now, Version: "v1", Server: "law-oa-go", Environment: "development"}})
}

func (h *DemoAggregateHandler) UpdateCaseIntake(c *gin.Context) {
	id := c.Param("id")
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	if !services.CanCreateCaseIntake(actor.Role) {
		common.APIForbidden(c, "无权更新接案记录", "当前账号没有接案工作台权限")
		return
	}
	actorID := strconv.FormatUint(uint64(actor.UserID), 10)
	var rawPayload map[string]interface{}
	if err := c.ShouldBindJSON(&rawPayload); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}
	parties := mapSliceValue(rawPayload["parties"])
	materials := mapSliceValue(rawPayload["materials"])
	if containsPlaintextIdentityField(rawPayload["metadata"]) {
		common.APIBadRequest(c, "身份信息登记方式不受支持", "身份证号、统一社会信用代码等身份标识必须登记到受保护的客户或主体档案，不能写入接案备注")
		return
	}
	for _, party := range parties {
		if containsPlaintextIdentityField(party["metadata"]) {
			common.APIBadRequest(c, "身份信息登记方式不受支持", "对方或关联方身份标识必须登记到受保护的主体档案，不能写入接案备注")
			return
		}
	}
	for _, material := range materials {
		if containsPlaintextIdentityField(material["metadata"]) {
			common.APIBadRequest(c, "身份信息登记方式不受支持", "材料备注不能包含未保护的身份标识")
			return
		}
	}
	payload := filterMap(rawPayload, allowedCaseIntakeUpdateFields)
	if len(payload) == 0 && len(parties) == 0 && len(materials) == 0 {
		common.APIBadRequest(c, "请求参数错误", "没有可更新的字段")
		return
	}
	if _, ok := payload["metadata"]; ok {
		payload["metadata"] = jsonStringValue(payload["metadata"])
	}
	if !isAssistantRole(c) && len(parties) > 0 && !h.intakePartyIdentitySchemaReady() {
		common.NewAPIError(c, http.StatusServiceUnavailable, "INTAKE_PARTY_IDENTITY_SCHEMA_REQUIRED", "接案当事人身份保护字段尚未迁移，已阻止更新；请执行数据库迁移 000070")
		return
	}
	payload["updated_at"] = time.Now()
	if !h.tableExists("case_intakes") {
		common.APINotFound(c, "接案记录不存在", "接案记录表不可用")
		return
	}
	var existing map[string]interface{}
	if err := h.db.Table("case_intakes").Where("id = ?", id).Take(&existing).Error; err != nil {
		common.APINotFound(c, "接案记录不存在", "指定接案记录不存在")
		return
	}
	role, _ := middleware.GetCurrentRole(c)
	isLawyerTakingOverAssistantDraft := strings.EqualFold(strings.TrimSpace(role), "lawyer") && strings.TrimSpace(fmt.Sprint(existing["status"])) == "assistant_draft"
	if !canViewAllMatterData(c) && strings.TrimSpace(fmt.Sprint(existing["created_by"])) != actorID && !isLawyerTakingOverAssistantDraft {
		forbidObjectAccess(c)
		return
	}
	if isAssistantRole(c) {
		if rawPayload["client_id"] != nil || rawPayload["parties"] != nil {
			common.APIForbidden(c, "助理协作草稿不能录入当事人身份信息", "客户、对方和关联方必须由执业律师确认后录入")
			return
		}
		if strings.TrimSpace(fmt.Sprint(existing["status"])) != "assistant_draft" {
			common.APIForbidden(c, "该接案草稿已进入律师处理阶段", "助理不能修改律师确认后的接案信息")
			return
		}
		payload = filterMap(rawPayload, allowedAssistantDraftFields)
		payload["status"] = "assistant_draft"
		payload["metadata"] = jsonStringValue(map[string]interface{}{
			"assistant_draft": true,
			"created_by_role": "assistant",
		})
	}
	if clientIDValue, present := rawPayload["client_id"]; present && clientIDValue != nil && !isAssistantRole(c) {
		if !h.authorizeIntakeClient(c, intakeClientID(clientIDValue)) {
			return
		}
	}
	if isLawyerTakingOverAssistantDraft {
		// A lawyer taking over an assistant draft becomes accountable for the
		// later confirmation and conflict-check trail.
		payload["created_by"] = actorID
	}
	// Any change to facts that can affect the conflict result invalidates the
	// prior conclusion at the server boundary. The UI may show a stale badge,
	// but the backend must enforce the same rule for API clients and scripts.
	conflictInputChanged := rawPayload["client_id"] != nil || rawPayload["title"] != nil ||
		rawPayload["case_type"] != nil || rawPayload["metadata"] != nil || rawPayload["parties"] != nil
	if !isAssistantRole(c) && conflictInputChanged {
		currentStatus := strings.TrimSpace(fmt.Sprint(existing["status"]))
		if currentStatus == "lawyer_facts_confirmed" || currentStatus == "conflict_ready" {
			metadata := objectValue(existing["metadata"])
			for key, value := range objectValue(rawPayload["metadata"]) {
				metadata[key] = value
			}
			metadata["conflict_invalidated_at"] = time.Now().UTC().Format(time.RFC3339)
			metadata["conflict_invalidated_by"] = actorID
			metadata["conflict_invalidation_reason"] = "接案事实发生变化，必须重新确认并检测"
			payload["status"] = "draft"
			payload["metadata"] = jsonStringValue(metadata)
		}
	}

	now := time.Now()
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if len(payload) > 0 {
			if err := tx.Table("case_intakes").Where("id = ?", id).Updates(payload).Error; err != nil {
				return err
			}
		}
		if h.tableExists("case_intake_parties") && rawPayload["parties"] != nil {
			if err := tx.Table("case_intake_parties").Where("intake_id = ?", id).Delete(map[string]interface{}{}).Error; err != nil {
				return err
			}
			for _, party := range parties {
				row, err := prepareCaseIntakePartyRow(party, id, now)
				if err != nil {
					return err
				}
				if err := tx.Table("case_intake_parties").Create(row).Error; err != nil {
					return err
				}
			}
		}
		if h.tableExists("case_materials") && rawPayload["materials"] != nil {
			if err := tx.Table("case_materials").Where("intake_id = ?", id).Delete(map[string]interface{}{}).Error; err != nil {
				return err
			}
			for _, material := range materials {
				row := filterMap(material, allowedCaseMaterialFields)
				row["intake_id"] = id
				row["name"] = stringValue(row["name"], "未命名材料")
				row["material_type"] = stringValue(row["material_type"], "document")
				row["status"] = stringValue(row["status"], "missing")
				if _, exists := row["required"]; !exists {
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
		if isSubjectWorkflowError(err) {
			writeSubjectWorkflowError(c, err)
			return
		}
		common.APIInternalServerError(c, "更新接案失败", err.Error())
		return
	}
	var updated map[string]interface{}
	if err := h.db.Table("case_intakes").Where("id = ?", id).Take(&updated).Error; err != nil {
		common.APIInternalServerError(c, "读取更新后的接案失败", err.Error())
		return
	}
	common.APISuccess(c, updated)
}

func prepareCaseIntakePartyRow(party map[string]interface{}, intakeID interface{}, now time.Time) (map[string]interface{}, error) {
	row := filterMap(party, allowedCaseIntakePartyFields)
	row["intake_id"] = intakeID
	row["entity_name"] = strings.TrimSpace(stringValue(row["entity_name"], stringValue(party["name"], "")))
	row["entity_type"] = normalizeIntakeEntityType(stringValue(row["entity_type"], stringValue(party["entityType"], "")))
	row["party_role"] = strings.ToLower(strings.TrimSpace(stringValue(row["party_role"], stringValue(party["role"], "related_party"))))
	if row["entity_name"] == "" {
		return nil, services.NewSubjectWorkflowError("INTAKE_PARTY_NAME_REQUIRED", "接案当事人名称不能为空")
	}
	if _, ok := row["relation_depth"]; !ok {
		row["relation_depth"] = 0
	}
	row["metadata"] = jsonStringValue(row["metadata"])
	row["created_at"] = now
	if row["party_role"] == "client" {
		return row, nil
	}
	identityType := strings.ToUpper(strings.TrimSpace(stringValue(party["identity_type"], stringValue(party["identityType"], ""))))
	identityNumber := security.NormalizeIdentityNumber(identityType, stringValue(party["identity_number"], stringValue(party["identityNumber"], "")))
	if !validIntakeIdentityType(row["entity_type"], identityType) || len([]rune(identityNumber)) < 4 {
		return nil, services.NewSubjectWorkflowError("INTAKE_PARTY_IDENTITY_REQUIRED", fmt.Sprintf("当事人“%s”必须提供与主体类型匹配的可核验身份标识", row["entity_name"]))
	}
	ciphertext, digest, err := security.ProtectIdentityNumber(identityNumber)
	if err != nil {
		return nil, services.NewSubjectWorkflowError("SUBJECT_DATA_KEY_REQUIRED", "主体身份保护密钥不可用，已阻止保存接案当事人")
	}
	row["identity_type"] = identityType
	row["identity_number_ciphertext"] = ciphertext
	row["identity_number_digest"] = digest
	row["aliases"] = strings.Join(stringListValue(firstNonEmptyValue(party["aliases"], party["alias"])), ",")
	return row, nil
}

func normalizeIntakeEntityType(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "PERSON", "INDIVIDUAL", "个人", "自然人":
		return "INDIVIDUAL"
	case "ORGANIZATION", "组织", "其他组织":
		return "ORGANIZATION"
	default:
		return "LEGAL_PERSON"
	}
}

func validIntakeIdentityType(entityType interface{}, identityType string) bool {
	if fmt.Sprint(entityType) == "INDIVIDUAL" {
		return identityType == "ID_CARD" || identityType == "PASSPORT" || identityType == "OTHER"
	}
	return identityType == "SOCIAL_CREDIT_CODE" || identityType == "BUSINESS_LICENSE" || identityType == "ORGANIZATION_CODE" || identityType == "OTHER"
}

func (h *DemoAggregateHandler) intakePartyIdentitySchemaReady() bool {
	for _, column := range []string{"identity_type", "identity_number_ciphertext", "identity_number_digest", "aliases"} {
		if !h.hasColumn("case_intake_parties", column) {
			return false
		}
	}
	return true
}

// ConfirmIntakeFacts records the lawyer's accountability for the factual
// intake before a conflict query can be started. Status is deliberately not a
// client-controlled field: only this transition may enter the checked state.
func (h *DemoAggregateHandler) ConfirmIntakeFacts(c *gin.Context) {
	role, _ := middleware.GetCurrentRole(c)
	if !strings.EqualFold(strings.TrimSpace(role), "lawyer") {
		common.NewAPIError(c, http.StatusForbidden, "INTAKE_FACT_CONFIRMATION_FORBIDDEN", "仅执业律师可以确认当事人事实")
		return
	}
	intakeID := c.Param("id")
	actorID, ok := currentUserIDString(c)
	if !ok {
		return
	}
	var intake map[string]interface{}
	if h.first("case_intakes", intakeID, &intake) != nil {
		common.APINotFound(c, "接案记录不存在", "指定接案记录不存在")
		return
	}
	if !canViewAllMatterData(c) && strings.TrimSpace(fmt.Sprint(intake["created_by"])) != actorID && strings.TrimSpace(fmt.Sprint(intake["status"])) != "assistant_draft" {
		forbidObjectAccess(c)
		return
	}
	if strings.TrimSpace(fmt.Sprint(intake["client_id"])) == "" || strings.TrimSpace(fmt.Sprint(intake["title"])) == "" || strings.TrimSpace(fmt.Sprint(intake["case_type"])) == "" {
		common.APIBadRequest(c, "冲突检查前置资料不完整", "请先填写客户、案件名称和案件类型")
		return
	}
	if clientID := intakeClientID(intake["client_id"]); clientID == 0 || !h.authorizeIntakeClient(c, clientID) {
		return
	}
	if h.tableExists("case_intake_parties") {
		var count int64
		if err := h.db.Table("case_intake_parties").Where("intake_id = ?", intakeID).Count(&count).Error; err != nil || count == 0 {
			common.APIBadRequest(c, "冲突检查前置资料不完整", "至少需要一名对方当事人或相关方")
			return
		}
	}
	metadata := objectValue(intake["metadata"])
	metadata["lawyer_facts_confirmed_at"] = time.Now().UTC().Format(time.RFC3339)
	metadata["lawyer_facts_confirmed_by"] = actorID
	if err := h.db.Table("case_intakes").Where("id = ?", intakeID).Updates(map[string]interface{}{
		"created_by": actorID,
		"status":     "lawyer_facts_confirmed",
		"metadata":   jsonStringValue(metadata),
		"updated_at": time.Now(),
	}).Error; err != nil {
		common.APIInternalServerError(c, "确认接案事实失败", err.Error())
		return
	}
	common.APISuccess(c, gin.H{"id": intakeID, "status": "lawyer_facts_confirmed"})
}

func (h *DemoAggregateHandler) StartIntakeConflictCheck(c *gin.Context) {
	if !canRunConflictCheck(c) {
		common.NewAPIError(c, http.StatusForbidden, "CONFLICT_CHECK_FORBIDDEN", "仅执业律师或独立冲突核查人可以运行利益冲突检查")
		return
	}
	intakeID := c.Param("id")
	var intake map[string]interface{}
	if h.first("case_intakes", intakeID, &intake) != nil {
		common.APINotFound(c, "接案记录不存在", "指定接案记录不存在")
		return
	}
	actorID, ok := currentUserIDString(c)
	if !ok {
		return
	}
	if !canViewAllMatterData(c) {
		if strings.TrimSpace(fmt.Sprint(intake["created_by"])) != actorID {
			forbidObjectAccess(c)
			return
		}
	}
	if strings.TrimSpace(fmt.Sprint(intake["status"])) != "lawyer_facts_confirmed" {
		common.NewAPIError(c, http.StatusConflict, "INTAKE_FACTS_NOT_CONFIRMED", "请由负责律师先确认当事人事实，再运行利益冲突检查")
		return
	}
	if clientID := intakeClientID(intake["client_id"]); clientID == 0 || !h.authorizeIntakeClient(c, clientID) {
		return
	}
	if h.conflictService == nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "CONFLICT_SERVICE_UNAVAILABLE", "正式冲突检测服务未初始化，已阻止创建空任务")
		return
	}

	var client map[string]interface{}
	if err := h.db.Table("clients").Where("id = ?", intake["client_id"]).Take(&client).Error; err != nil {
		common.APIBadRequest(c, "冲突检查前置资料不完整", "接案草稿关联的客户不存在")
		return
	}
	metadata := objectValue(intake["metadata"])
	if containsPlaintextIdentityField(metadata) {
		common.APIBadRequest(c, "冲突检查前置资料不完整", "接案记录包含未保护的身份标识，请先登记到受保护的客户或主体档案")
		return
	}
	clientName := firstNonEmpty(fmt.Sprint(client["name"]), fmt.Sprint(intake["client_name"]))
	clientType := strings.ToUpper(strings.TrimSpace(fmt.Sprint(client["type"])))
	if clientType == "企业" || clientType == "公司" || clientType == "COMPANY" {
		clientType = "COMPANY"
	} else if clientType == "个人" || clientType == "PERSON" {
		clientType = "PERSON"
	} else {
		clientType = "ANY"
	}
	clientIdentifiers := stringMapValue(firstNonEmptyValue(
		client["identifiers"],
	))
	clientIdentityCiphertext := firstNonEmpty(valueString(client, "identity_number_ciphertext"), valueString(client, "id_card_ciphertext"))
	clientIdentityDigest := firstNonEmpty(valueString(client, "identity_number_digest"), valueString(client, "id_card_digest"))
	if clientIdentityCiphertext == "" || clientIdentityDigest == "" {
		common.NewAPIError(c, http.StatusConflict, "INTAKE_CLIENT_IDENTITY_REQUIRED", "客户主档案缺少受保护的可核验身份标识，不能运行冲突检测")
		return
	}
	clientIdentity, decryptErr := security.DecryptIdentityNumber(clientIdentityCiphertext)
	if decryptErr != nil || strings.TrimSpace(clientIdentity) == "" {
		common.NewAPIError(c, http.StatusConflict, "CLIENT_IDENTITY_UNREADABLE", "客户主档案的身份标识无法安全读取，不能运行冲突检测")
		return
	}
	if clientIdentifiers == nil {
		clientIdentifiers = map[string]string{}
	}
	clientIdentityType := strings.ToLower(strings.TrimSpace(valueString(client, "identity_type")))
	switch strings.ToUpper(clientIdentityType) {
	case "SOCIAL_CREDIT_CODE":
		clientIdentityType = "unified_social_credit_code"
	case "BUSINESS_LICENSE":
		clientIdentityType = "business_license"
	case "ORGANIZATION_CODE":
		clientIdentityType = "organization_code"
	case "PASSPORT":
		clientIdentityType = "passport"
	case "ID_CARD":
		clientIdentityType = "id_card"
	default:
		if clientType == "COMPANY" {
			clientIdentityType = "unified_social_credit_code"
		} else {
			clientIdentityType = "id_card"
		}
	}
	clientIdentifiers[clientIdentityType] = security.NormalizeIdentityNumber(clientIdentityType, clientIdentity)
	clientAliases := stringListValue(firstNonEmptyValue(metadata["client_aliases"], metadata["clientAliases"]))
	lawyerID := actorID
	if canViewAllMatterData(c) {
		lawyerID = firstNonEmpty(valueString(metadata, "lawyer_id"), actorID)
	}
	lawyerIDUint, lawyerErr := strconv.ParseUint(strings.TrimSpace(lawyerID), 10, 32)
	if lawyerErr != nil || lawyerIDUint == 0 {
		common.NewAPIError(c, http.StatusConflict, "INTAKE_LAWYER_REQUIRED", "冲突检测必须绑定有效的承办律师")
		return
	}
	var lawyer map[string]interface{}
	if err := h.db.Table("users").Select("id, role, status").Where("id = ? AND deleted_at IS NULL", uint(lawyerIDUint)).Take(&lawyer).Error; err != nil {
		common.NewAPIError(c, http.StatusConflict, "INTAKE_LAWYER_INVALID", "接案记录绑定的承办律师不存在")
		return
	}
	if !strings.EqualFold(strings.TrimSpace(fmt.Sprint(lawyer["status"])), "active") || !strings.EqualFold(strings.TrimSpace(fmt.Sprint(lawyer["role"])), "lawyer") {
		common.NewAPIError(c, http.StatusConflict, "INTAKE_LAWYER_INVALID", "接案记录绑定的承办律师未启用或账号角色无效")
		return
	}
	parties := make([]models.ConflictPartyInfo, 0)
	otherParties := make([]string, 0)
	if h.tableExists("case_intake_parties") {
		var partyRows []map[string]interface{}
		if err := h.db.Table("case_intake_parties").Where("intake_id = ?", intakeID).Find(&partyRows).Error; err != nil {
			common.APIInternalServerError(c, "读取接案当事人失败", err.Error())
			return
		}
		seenParties := map[string]struct{}{}
		for _, row := range partyRows {
			name := firstNonEmpty(valueString(row, "entity_name"), valueString(row, "name"))
			if name == "" || strings.EqualFold(name, clientName) {
				continue
			}
			role := strings.ToUpper(firstNonEmpty(valueString(row, "party_role"), valueString(row, "role"), "RELATED_PARTY"))
			entityType := strings.ToUpper(firstNonEmpty(valueString(row, "entity_type"), "ANY"))
			partyMetadata := objectValue(row["metadata"])
			if containsPlaintextIdentityField(partyMetadata) {
				common.APIBadRequest(c, "冲突检查前置资料不完整", "对方或关联方身份标识未登记在受保护的主体档案中")
				return
			}
			identityType := strings.ToUpper(strings.TrimSpace(valueString(row, "identity_type")))
			ciphertext := strings.TrimSpace(valueString(row, "identity_number_ciphertext"))
			if identityType == "" || ciphertext == "" || strings.TrimSpace(valueString(row, "identity_number_digest")) == "" {
				common.NewAPIError(c, http.StatusConflict, "INTAKE_PARTY_IDENTITY_REQUIRED", fmt.Sprintf("当事人“%s”缺少受保护的可核验身份标识，不能运行冲突检测", name))
				return
			}
			identityNumber, decryptErr := security.DecryptIdentityNumber(ciphertext)
			if decryptErr != nil {
				common.NewAPIError(c, http.StatusConflict, "SUBJECT_IDENTITY_UNREADABLE", fmt.Sprintf("当事人“%s”的身份标识无法安全读取，不能运行冲突检测", name))
				return
			}
			identifiers := map[string]string{strings.ToLower(identityType): security.NormalizeIdentityNumber(identityType, identityNumber)}
			aliases := stringListValue(firstNonEmptyValue(row["aliases"], partyMetadata["aliases"], partyMetadata["client_aliases"], partyMetadata["clientAliases"]))
			parties = append(parties, models.ConflictPartyInfo{Name: name, Role: role, EntityType: entityType, Identifiers: identifiers, Aliases: aliases})
			if _, exists := seenParties[name]; !exists {
				seenParties[name] = struct{}{}
				otherParties = append(otherParties, name)
			}
		}
	}
	if len(otherParties) == 0 {
		common.APIBadRequest(c, "冲突检查前置资料不完整", "至少需要一名对方当事人或相关方")
		return
	}
	actorIDUint, _ := strconv.ParseUint(actorID, 10, 32)
	actorRole, _ := middleware.GetCurrentRole(c)
	request := &models.ConflictCheckRequest{
		CheckID:                   "CCT_" + uuid.NewString(),
		SubjectCaseID:             firstNonEmpty(valueString(metadata, "subject_case_id"), valueString(metadata, "subjectCaseId")),
		SubjectCaseNumber:         firstNonEmpty(valueString(metadata, "subject_case_number"), valueString(metadata, "subjectCaseNumber")),
		IntakeID:                  intakeID,
		ClientID:                  fmt.Sprint(intake["client_id"]),
		ClientName:                clientName,
		ClientType:                clientType,
		ClientIdentifiers:         clientIdentifiers,
		ClientAliases:             clientAliases,
		OtherParties:              otherParties,
		Parties:                   parties,
		CaseName:                  strings.TrimSpace(fmt.Sprint(intake["title"])),
		CaseType:                  strings.TrimSpace(fmt.Sprint(intake["case_type"])),
		SearchYears:               0,
		IncludeCorporateRelations: true,
		SearchDepth:               "STANDARD",
		UserID:                    lawyerID,
		ActorUserID:               uint(actorIDUint),
		ActorRole:                 actorRole,
		RequestTime:               time.Now(),
	}
	if request.CaseName == "" || request.CaseType == "" {
		common.APIBadRequest(c, "冲突检查前置资料不完整", "案件名称和案件类型不能为空")
		return
	}
	result, err := h.conflictService.PerformConflictCheck(c.Request.Context(), request)
	if err != nil {
		common.APIInternalServerError(c, "执行冲突检查失败", err.Error())
		return
	}
	if result == nil || strings.TrimSpace(result.CheckID) == "" {
		common.NewAPIError(c, http.StatusServiceUnavailable, "CONFLICT_RESULT_UNAVAILABLE", "冲突检测未返回可追溯的检测记录，已阻止继续办理")
		return
	}
	// This is the primary intake workflow, so it must apply the same
	// least-disclosure projection as the standalone conflict endpoint before
	// serializing the result to the requesting lawyer. The persisted evidence
	// remains available to the independently authorized reviewer workflow.
	projectConflictResponseForViewer(c, result)
	coverageStatus := ""
	if result.Decision != nil {
		coverageStatus = result.Decision.CoverageStatus
	}
	if err := repositories.LinkConflictCheckToCase(c.Request.Context(), h.db, repositories.ConflictSubjectAssociation{
		CheckID:           result.CheckID,
		SubjectCaseID:     firstNonEmpty(valueString(metadata, "subject_case_id"), valueString(metadata, "subjectCaseId")),
		SubjectCaseNumber: firstNonEmpty(valueString(metadata, "subject_case_number"), valueString(metadata, "subjectCaseNumber")),
		IntakeID:          intakeID,
		ClientID:          fmt.Sprint(intake["client_id"]),
		CoverageStatus:    coverageStatus,
		CheckedAt:         result.CheckTime,
	}); err != nil {
		common.APIInternalServerError(c, "保存冲突检测结果失败", err.Error())
		return
	}
	metadata["conflict_check_id"] = result.CheckID
	if result.Decision != nil {
		metadata["conflict_coverage_status"] = result.Decision.CoverageStatus
	}
	metadata["conflict_checked_at"] = result.CheckTime
	if err := h.db.Table("case_intakes").Where("id = ?", intakeID).Updates(map[string]interface{}{
		"status":     "conflict_ready",
		"metadata":   jsonStringValue(metadata),
		"updated_at": time.Now(),
	}).Error; err != nil {
		common.APIInternalServerError(c, "保存冲突检测结果失败", "检测结果未能写回接案记录，已阻止继续办理")
		return
	}

	common.APISuccess(c, gin.H{
		"taskId":                     result.CheckID,
		"checkId":                    result.CheckID,
		"intake_id":                  intakeID,
		"status":                     "COMPLETED",
		"recommendedPollingInterval": 2,
		"createdAt":                  result.CheckTime,
		"result":                     result,
	})
}

func (h *DemoAggregateHandler) CreateConflictApproval(c *gin.Context) {
	taskID := c.Param("task_id")
	if !h.tableExists("approval_requests") {
		common.APIInternalServerError(c, "审批表不存在", "approval_requests table is required")
		return
	}
	applicantID, ok := currentUserIDString(c)
	if !ok {
		return
	}
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}
	response, err := h.createConflictApproval(c, taskID, applicantID, payload)
	if err != nil {
		if businessErr, ok := err.(conflictApprovalError); ok {
			common.NewAPIError(c, businessErr.status, businessErr.code, businessErr.message, businessErr.details)
			return
		}
		common.APIInternalServerError(c, "创建冲突审批失败", err.Error())
		return
	}
	common.APISuccess(c, response)
}

type conflictApprovalError struct {
	status  int
	code    string
	message string
	details string
}

func (e conflictApprovalError) Error() string { return e.message }

func (h *DemoAggregateHandler) createConflictApproval(c *gin.Context, taskID, applicantID string, payload map[string]interface{}) (gin.H, error) {
	var response gin.H
	hasConflictApprovalID := h.hasColumn("approval_requests", "conflict_check_id")
	hasWaiverApplications := h.tableExists("waiver_applications")
	hasConflictCases := h.tableExists("conflict_cases")
	hasApprovalSnapshots := h.tableExists("approval_snapshots")
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var row map[string]interface{}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Table("conflict_check_records").Where("check_id = ?", taskID).Take(&row).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return conflictApprovalError{status: http.StatusNotFound, code: "CONFLICT_TASK_NOT_FOUND", message: "冲突检测记录不存在"}
			}
			return err
		}
		ownerID, parseErr := strconv.ParseUint(strings.TrimSpace(fmt.Sprint(row["user_id"])), 10, 32)
		task := &services.ConflictCheckTaskResponse{
			OwnerID:       uint(ownerID),
			SubjectCaseID: subjectCaseIDFromAggregateSearchParameters(row["search_parameters"]),
		}
		if parseErr != nil || !canAccessConflictTaskForActor(c, h.authz, task) {
			return conflictApprovalError{status: http.StatusForbidden, code: "FORBIDDEN", message: "无权为其他律师的冲突任务创建审批"}
		}
		if strings.ToUpper(strings.TrimSpace(fmt.Sprint(row["check_status"]))) != "COMPLETED" {
			return conflictApprovalError{status: http.StatusConflict, code: "CONFLICT_CHECK_NOT_COMPLETED", message: "冲突检测尚未完成，不能创建冲突审批"}
		}

		checkResult := objectValue(row["check_result"])
		decision := objectValue(checkResult["decision"])
		decisionStatus := strings.ToUpper(valueString(decision, "status"))
		waiverStatus := strings.ToUpper(valueString(objectValue(checkResult["waiver"]), "status"))
		if decisionStatus == "WAIVED" || waiverStatus == "APPROVED" {
			return conflictApprovalError{status: http.StatusConflict, code: "CONFLICT_ALREADY_WAIVED", message: "冲突记录已获批准豁免，不能再创建冲突审批"}
		}
		if decisionStatus == "WAIVER_PENDING" || decisionStatus == "UNDER_REVIEW" || waiverStatus == "UNDER_REVIEW" || waiverStatus == "SUBMITTED" {
			return conflictApprovalError{status: http.StatusConflict, code: "CONFLICT_WAIVER_IN_PROGRESS", message: "冲突豁免正在复核，不能并行创建冲突审批"}
		}
		if hasWaiverApplications {
			var waiver map[string]interface{}
			waiverErr := tx.Table("waiver_applications").Select("id, status").
				Where("conflict_check_id = ? AND deleted_at IS NULL", taskID).Order("created_at DESC").Take(&waiver).Error
			if waiverErr == nil {
				switch strings.ToUpper(strings.TrimSpace(fmt.Sprint(waiver["status"]))) {
				case "APPROVED":
					return conflictApprovalError{status: http.StatusConflict, code: "CONFLICT_ALREADY_WAIVED", message: "冲突记录已获批准豁免，不能再创建冲突审批"}
				case "SUBMITTED", "UNDER_REVIEW":
					return conflictApprovalError{status: http.StatusConflict, code: "CONFLICT_WAIVER_IN_PROGRESS", message: "冲突豁免正在复核，不能并行创建冲突审批"}
				}
			} else if waiverErr != gorm.ErrRecordNotFound {
				return waiverErr
			}
		}

		if hasConflictApprovalID {
			var existing map[string]interface{}
			existingErr := tx.Table("approval_requests").
				Where("conflict_check_id = ? AND type = ? AND deleted_at IS NULL", taskID, "conflict_approval").
				Order("created_at DESC").Take(&existing).Error
			if existingErr == nil {
				status := strings.ToLower(strings.TrimSpace(fmt.Sprint(existing["status"])))
				if status == "submitted" || status == "under_review" || status == "resubmitted" {
					if strings.TrimSpace(fmt.Sprint(existing["applicant_id"])) != applicantID {
						return conflictApprovalError{status: http.StatusConflict, code: "CONFLICT_APPROVAL_OWNER_MISMATCH", message: "现有冲突审批的申请人与检测任务所有者不一致，请联系管理员修复数据后重试"}
					}
					response = conflictApprovalResponse(existing, taskID, true)
					return nil
				}
				return conflictApprovalError{status: http.StatusConflict, code: "CONFLICT_APPROVAL_FINAL", message: "该冲突任务已有终态审批，不能重复创建同类审批"}
			}
			if existingErr != gorm.ErrRecordNotFound {
				return existingErr
			}
		}
		if terminalConflictReview(checkResult) {
			return conflictApprovalError{status: http.StatusConflict, code: "CONFLICT_REVIEW_FINAL", message: "冲突记录已有终态人工复核结论，不能重复创建冲突审批"}
		}

		parameters := objectValue(row["search_parameters"])
		expectedSubjectCaseID := strings.TrimSpace(fmt.Sprint(payload["expected_subject_case_id"]))
		if expectedSubjectCaseID != "" && expectedSubjectCaseID != "<nil>" {
			actualSubjectCaseID := firstNonEmpty(valueString(parameters, "subjectCaseId"), valueString(parameters, "subject_case_id"))
			if actualSubjectCaseID == "" || actualSubjectCaseID != expectedSubjectCaseID {
				return conflictApprovalError{status: http.StatusBadRequest, code: "CONFLICT_CASE_MISMATCH", message: "检测记录与当前案件不一致，已阻止创建冲突审批"}
			}
		}

		conflictCases := []map[string]interface{}{}
		if hasConflictCases {
			if err := tx.Table("conflict_cases").Where("check_id = ?", taskID).
				Order("CASE risk_level WHEN 'CRITICAL' THEN 4 WHEN 'HIGH' THEN 3 WHEN 'MEDIUM' THEN 2 WHEN 'LOW' THEN 1 ELSE 0 END DESC, created_at DESC").
				Find(&conflictCases).Error; err != nil {
				return err
			}
		}
		if conflictApprovalNotRequired(row, checkResult, decision, len(conflictCases)) {
			return conflictApprovalError{status: http.StatusConflict, code: "CONFLICT_APPROVAL_NOT_REQUIRED", message: "当前冲突检查结果为 CLEAR/LOW 且无有效命中，不需要独立冲突审批"}
		}

		var applicantRow map[string]interface{}
		applicantName := c.GetString("username")
		if err := tx.Table("users").Select("id, name, username, department").Where("id = ? AND deleted_at IS NULL", applicantID).Take(&applicantRow).Error; err == nil {
			applicantName = firstNonEmpty(fmt.Sprint(applicantRow["name"]), fmt.Sprint(applicantRow["username"]), applicantName)
		}
		applicantName = firstNonEmpty(applicantName, "当前用户")
		approver, err := h.findConflictApproverDB(tx, applicantID)
		if err != nil {
			return conflictApprovalError{status: http.StatusBadRequest, code: "CONFLICT_APPROVER_REQUIRED", message: "未找到可用的独立审批人", details: err.Error()}
		}

		approvalID := uuid.NewString()
		now := time.Now()
		requestNumber := fmt.Sprintf("APR-%s-%s", now.Format("20060102150405"), strings.ToUpper(approvalID[:8]))
		conflictRecord := conflictApprovalRecord(row)
		conflictResult := conflictApprovalResult(taskID, row, conflictRecord, checkResult, conflictCases)
		snapshotFields := conflictApprovalSnapshotFields(row, parameters, checkResult, conflictCases)
		metadata := gin.H{
			"conflict_task_id": taskID, "source": "real_api", "conflict_record": conflictRecord,
			"conflict_result": conflictResult, "conflict_cases": conflictCases,
			"client_name": snapshotFields["client_name"], "opposing_parties": snapshotFields["opposing_parties"],
			"subjects": snapshotFields["subjects"], "normalizedSubjects": snapshotFields["normalizedSubjects"],
			"decision": snapshotFields["decision"], "evidence": snapshotFields["evidence"],
		}
		approval := gin.H{
			"id": approvalID, "request_number": requestNumber,
			"title": stringValue(payload["title"], fmt.Sprintf("冲突审查审批 - %s", taskID)),
			"type":  "conflict_approval", "category": "conflict_review",
			"content":      stringValue(payload["content"], "利益冲突检查结果进入审批复核。"),
			"applicant_id": applicantID, "applicant_name": applicantName,
			"applicant_title": stringValue(payload["applicant_title"], "律师"),
			"department_id":   stringValue(payload["department_id"], "risk"),
			"department_name": firstNonEmpty(fmt.Sprint(applicantRow["department"]), "业务部门"),
			"urgency":         stringValue(payload["urgency"], "urgent"), "priority": stringValue(payload["priority"], "high"),
			"status": "submitted", "submission_date": now, "current_stage": "合规复核",
			"current_approver_id": approver["id"], "current_approver_name": approver["name"],
			"workflow_type": "CONFLICT_APPROVAL", "workflow_config": "{}", "attachments": "[]",
			"metadata": jsonStringValue(metadata), "created_by": applicantID, "created_at": now, "updated_at": now,
		}
		if hasConflictApprovalID {
			approval["conflict_check_id"] = taskID
			approval["conflict_risk_level"] = row["risk_level"]
			approval["conflict_check_time"] = row["created_at"]
			approval["conflict_result"] = jsonStringValue(conflictResult)
		}
		if err := tx.Table("approval_requests").Create(cleanInsertMap(approval)).Error; err != nil {
			return err
		}
		if hasApprovalSnapshots {
			snapshotFields["conflict_task_id"] = taskID
			snapshotFields["conflict_record"] = conflictRecord
			snapshotFields["conflict_result"] = conflictResult
			snapshotFields["conflict_cases"] = conflictCases
			snapshotFields["approval"] = approval
			snapshotFields["metadata"] = metadata
			snapshot := gin.H{
				"id": uuid.NewString(), "approval_request_id": approvalID, "snapshot_type": "conflict_approval",
				"snapshot_data": jsonStringValue(snapshotFields), "source_version": 1, "created_at": now,
			}
			if err := tx.Table("approval_snapshots").Create(cleanInsertMap(snapshot)).Error; err != nil {
				return err
			}
		}
		response = conflictApprovalResponse(approval, taskID, false)
		response["snapshot_url"] = fmt.Sprintf("/api/v1/approvals/%s/snapshot", approvalID)
		response["submitted_at"] = now
		return nil
	})
	return response, err
}

func terminalConflictReview(checkResult map[string]interface{}) bool {
	reviewDecision := strings.ToLower(valueString(objectValue(checkResult["review"]), "decision"))
	switch reviewDecision {
	case "no_conflict", "confirmed_conflict", "false_positive":
		return true
	default:
		return false
	}
}

func conflictApprovalResponse(approval map[string]interface{}, taskID string, reused bool) gin.H {
	return gin.H{
		"approval_id": approval["id"], "request_number": approval["request_number"], "conflict_task": taskID,
		"status": approval["status"], "current_approver_id": approval["current_approver_id"],
		"current_approver_name": approval["current_approver_name"], "reused": reused,
	}
}

func conflictApprovalRecord(row map[string]interface{}) gin.H {
	return gin.H{
		"check_id": row["check_id"], "user_id": row["user_id"], "case_name": row["case_name"],
		"client_id": row["client_id"], "client_name": row["client_name"], "case_type": row["case_type"],
		"status": row["check_status"], "risk_level": row["risk_level"], "has_conflict": row["has_conflict"],
		"check_result": row["check_result"], "search_parameters": row["search_parameters"],
		"created_at": row["created_at"], "updated_at": row["updated_at"],
	}
}

func conflictApprovalNotRequired(row, checkResult, decision map[string]interface{}, conflictCaseCount int) bool {
	coverageStatus := strings.ToUpper(firstNonEmpty(
		valueString(decision, "coverageStatus"),
		valueString(decision, "coverage_status"),
	))
	// A machine CLEAR/LOW result is never sufficient when the authoritative
	// archive coverage is incomplete or absent. The caller must create an
	// independently reviewable approval instead of treating a limited search
	// as a clean result.
	if coverageStatus != "COMPLETE" {
		return false
	}
	status := strings.ToUpper(valueString(decision, "status"))
	if status == "CLEAR" {
		return true
	}
	risk := strings.ToUpper(firstNonEmpty(valueString(objectValue(checkResult["riskAssessment"]), "overallRisk"), fmt.Sprint(row["risk_level"])))
	requiresApproval := interfaceBool(objectValue(checkResult["riskAssessment"])["requiresApproval"]) || interfaceBool(decision["requiresManualReview"])
	evidenceCount := interfaceInt(decision["evidenceCount"])
	hasConflict := interfaceBool(row["has_conflict"])
	return risk == "LOW" && !requiresApproval && !hasConflict && evidenceCount == 0 && conflictCaseCount == 0
}

func conflictApprovalResult(taskID string, row map[string]interface{}, record gin.H, checkResult map[string]interface{}, conflictCases []map[string]interface{}) gin.H {
	result := gin.H{}
	for key, value := range checkResult {
		result[key] = value
	}
	result["checkId"] = taskID
	result["record"] = record
	if _, ok := result["conflictCases"]; !ok {
		result["conflictCases"] = conflictCases
	}
	if _, ok := result["riskAssessment"]; !ok {
		result["riskAssessment"] = gin.H{"overallRisk": stringValue(row["risk_level"], "LOW"), "riskScore": 0}
	}
	return result
}

func conflictApprovalSnapshotFields(row, parameters, checkResult map[string]interface{}, conflictCases []map[string]interface{}) gin.H {
	normalizedSubjects := firstNonNil(checkResult["normalizedSubjects"], checkResult["normalized_subjects"], []interface{}{})
	subjects := firstNonNil(parameters["subjects"], parameters["parties"], normalizedSubjects, []interface{}{})
	opposingParties := firstNonNil(parameters["opposingParties"], parameters["opposing_parties"], parameters["otherParties"], []interface{}{})
	return gin.H{
		"client_name": strings.TrimSpace(fmt.Sprint(row["client_name"])),
		"case_creation_config": gin.H{
			"case_type": strings.TrimSpace(fmt.Sprint(row["case_type"])),
			"case_name": strings.TrimSpace(fmt.Sprint(row["case_name"])),
		},
		"opposing_parties":   opposingParties,
		"subjects":           subjects,
		"normalizedSubjects": normalizedSubjects,
		"decision":           firstNonNil(checkResult["decision"], map[string]interface{}{}),
		"evidence":           conflictApprovalEvidence(checkResult, conflictCases),
	}
}

func conflictApprovalEvidence(checkResult map[string]interface{}, conflictCases []map[string]interface{}) []interface{} {
	evidence := make([]interface{}, 0)
	if direct, ok := checkResult["evidence"].([]interface{}); ok {
		evidence = append(evidence, direct...)
	}
	for _, conflictCase := range mapSliceValue(checkResult["conflictCases"]) {
		if items, ok := conflictCase["evidence"].([]interface{}); ok {
			evidence = append(evidence, items...)
		}
	}
	if len(evidence) == 0 {
		for _, conflictCase := range conflictCases {
			evidence = append(evidence, conflictCase)
		}
	}
	return evidence
}

func firstNonNil(values ...interface{}) interface{} {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func interfaceBool(value interface{}) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case float64:
		return typed != 0
	case int64:
		return typed != 0
	case []byte:
		return strings.EqualFold(strings.TrimSpace(string(typed)), "true") || string(typed) == "1"
	case string:
		return strings.EqualFold(strings.TrimSpace(typed), "true") || strings.TrimSpace(typed) == "1"
	default:
		return false
	}
}

func interfaceInt(value interface{}) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	default:
		return 0
	}
}

func (h *DemoAggregateHandler) findConflictApprover(applicantID string) (map[string]interface{}, error) {
	if !h.tableExists("users") {
		return nil, fmt.Errorf("用户表不存在")
	}
	return h.findConflictApproverDB(h.db, applicantID)
}

func (h *DemoAggregateHandler) findConflictApproverDB(db *gorm.DB, applicantID string) (map[string]interface{}, error) {
	var approver map[string]interface{}
	err := db.Table("users").
		Select("id, name, username, role").
		Where("deleted_at IS NULL AND status = ? AND id <> ?", "active", applicantID).
		Where("role IN ?", []string{"director", "partner", "compliance", "risk", "risk_control", "management", "conflict_officer"}).
		Order("CASE role WHEN 'conflict_officer' THEN 1 WHEN 'compliance' THEN 2 WHEN 'risk' THEN 3 WHEN 'risk_control' THEN 4 WHEN 'director' THEN 5 WHEN 'partner' THEN 6 WHEN 'management' THEN 7 ELSE 8 END").
		Order("id ASC").
		Take(&approver).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("请先配置合规、风控、主任、合伙人或冲突核查人账号")
		}
		return nil, err
	}
	approver["id"] = strings.TrimSpace(fmt.Sprint(approver["id"]))
	approver["name"] = firstNonEmpty(fmt.Sprint(approver["name"]), fmt.Sprint(approver["username"]), "合规负责人")
	if approver["id"] == applicantID {
		return nil, fmt.Errorf("申请人不能审批自己的冲突申请")
	}
	return approver, nil
}

func (h *DemoAggregateHandler) GetApprovalSnapshot(c *gin.Context) {
	id := c.Param("id")
	if !h.tableExists("approval_requests") {
		common.APINotFound(c, "审批不存在", "审批记录表不可用")
		return
	}
	var approval map[string]interface{}
	if err := h.db.Table("approval_requests").Where("id = ? AND deleted_at IS NULL", id).Take(&approval).Error; err != nil {
		common.APINotFound(c, "审批不存在", "指定审批不存在或已删除")
		return
	}
	if !h.canAccessApprovalSnapshot(c, approval) {
		return
	}
	if h.tableExists("approval_snapshots") {
		var row map[string]interface{}
		if err := h.db.Table("approval_snapshots").Where("approval_request_id = ?", id).Order("created_at DESC").Take(&row).Error; err == nil {
			common.APISuccess(c, gin.H{
				"approval_id": id,
				"snapshot":    projectApprovalSnapshotForViewer(row["snapshot_data"], c.GetString("role")),
				"immutable":   true,
			})
			return
		}
	}
	metadataText := fmt.Sprint(approval["metadata"])
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(metadataText), &metadata); err == nil {
		if snapshot, ok := metadata["approval_snapshot"]; ok {
			common.APISuccess(c, gin.H{
				"approval_id": id,
				"snapshot":    projectApprovalSnapshotForViewer(snapshot, c.GetString("role")),
				"immutable":   true,
			})
			return
		}
	}
	common.APINotFound(c, "审批快照不存在", "指定审批没有可用的不可变快照")
}

func (h *DemoAggregateHandler) canAccessApprovalSnapshot(c *gin.Context, approval map[string]interface{}) bool {
	actor, ok := currentAuthActor(c)
	if !ok {
		return false
	}
	applicantID := strings.TrimSpace(fmt.Sprint(approval["applicant_id"]))
	approverID := strings.TrimSpace(fmt.Sprint(approval["current_approver_id"]))
	if strconv.FormatUint(uint64(actor.UserID), 10) == applicantID || strconv.FormatUint(uint64(actor.UserID), 10) == approverID {
		return true
	}
	if h.authz == nil || (!services.IsConflictReviewRole(actor.Role) && !services.IsBusinessMatterManagementRole(actor.Role)) {
		forbidObjectAccess(c)
		return false
	}

	checkID := firstNonEmpty(
		strings.TrimSpace(fmt.Sprint(approval["conflict_check_id"])),
	)
	metadata := integrationJSONMap(approval["metadata"])
	if checkID == "" {
		checkID = firstNonEmpty(
			integrationMetadataString(metadata, "conflict_check_id"),
			integrationMetadataString(metadata, "conflict_task_id"),
		)
	}
	if checkID == "" || !h.tableExists("conflict_check_records") {
		forbidObjectAccess(c)
		return false
	}
	var record struct {
		SearchParameters []byte
	}
	if err := h.db.Table("conflict_check_records").Select("search_parameters").Where("check_id = ?", checkID).Take(&record).Error; err != nil {
		forbidObjectAccess(c)
		return false
	}
	caseIDText := subjectCaseIDFromAggregateSearchParameters(record.SearchParameters)
	caseID, err := strconv.ParseUint(caseIDText, 10, 32)
	if err != nil || caseID == 0 {
		forbidObjectAccess(c)
		return false
	}
	allowed, authErr := h.authz.CanReadConflictContext(c.Request.Context(), actor, uint(caseID))
	if authErr != nil {
		common.APIInternalServerError(c, "审批快照权限校验失败", authErr.Error())
		return false
	}
	if !allowed {
		forbidObjectAccess(c)
		return false
	}
	return true
}
