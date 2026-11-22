package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"law-oa-go/internal/common"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// intakeDraftRevisionKey is the server-owned draft revision token stored in
// the intake metadata. Reading paths only consume the token; the locked
// update/confirm paths own rotation.
const intakeDraftRevisionKey = "draft_revision"

// intakeDraftRevision reads the server-owned draft revision token from a raw
// metadata value (map, JSON string, or JSON bytes). Missing or malformed
// metadata yields an empty token.
func intakeDraftRevision(raw interface{}) string {
	value, _ := objectValue(raw)[intakeDraftRevisionKey].(string)
	return strings.TrimSpace(value)
}

// intakeDraftRevisionFor returns the server-owned revision token for an
// intake row, deriving a stable legacy token from the row identity when the
// stored metadata predates the revision guard. The derivation must stay
// server-side and identical across every read path (list, workbench detail,
// conflict hints, and the locked update/confirm re-check) so pre-guard drafts
// can be loaded and saved exactly once instead of being permanently locked
// behind a missing token.
func intakeDraftRevisionFor(intakeID string, raw interface{}) string {
	if revision := intakeDraftRevision(raw); revision != "" {
		return revision
	}
	return "legacy-" + strings.TrimSpace(intakeID)
}

// ListIntakeDrafts is the minimal server-side recovery list for the intake
// page. A fresh browser session has no localStorage entry, so "saved and
// exited" drafts must be discoverable from the normal menu without a deep
// link. The list is server-authoritative, authorization-filtered in SQL,
// paginated, and masked at the same boundary as the workbench aggregate: no
// plaintext identity values and no raw workflow metadata.
func (h *DemoAggregateHandler) ListIntakeDrafts(c *gin.Context) {
	if !h.tableExists("case_intakes") {
		common.APISuccess(c, gin.H{"items": []gin.H{}, "page": 1, "page_size": 10, "total": 0})
		return
	}
	query := h.visibleDraftQuery(c)
	if query == nil {
		common.APIForbidden(c, "无法读取接案草稿", "当前角色或读取条件不满足")
		return
	}
	page := 1
	if v, err := strconv.Atoi(c.Query("page")); err == nil && v > 0 {
		page = v
	}
	pageSize := 10
	if v, err := strconv.Atoi(c.Query("page_size")); err == nil && v > 0 && v <= 50 {
		pageSize = v
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		common.APIInternalServerError(c, "读取接案草稿列表失败", err.Error())
		return
	}
	rows := []map[string]interface{}{}
	if err := query.
		Order("updated_at DESC, id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&rows).Error; err != nil {
		common.APIInternalServerError(c, "读取接案草稿列表失败", err.Error())
		return
	}
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, gin.H{
			"id":          row["id"],
			"intake_code": row["intake_code"],
			"title":       row["title"],
			"case_type":   row["case_type"],
			"status":      row["status"],
			"client_id":   row["client_id"],
			"created_by":  row["created_by"],
			"updated_at":  row["updated_at"],
			"revision":    intakeDraftRevisionFor(fmt.Sprint(row["id"]), row["metadata"]),
		})
	}
	common.APISuccess(c, gin.H{
		"items":     items,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	})
}

// excludeIntakesWithCompletedCaseCreation removes intakes whose formal-case
// approval has already been completed as a real case from the recovery-draft
// projection. The intake row itself stays conflict_ready (frozen submission
// snapshots are audit history), so the only durable signal that a draft no
// longer needs recovery is the tracking row written inside the case-creation
// transaction. When the tracking table is absent or predates the intake_id
// column (older schemas) the predicate degrades to a no-op so the recovery
// list never breaks.
func (h *DemoAggregateHandler) excludeIntakesWithCompletedCaseCreation(query *gorm.DB) *gorm.DB {
	if !h.tableExists("approval_requests") || !h.tableExists("approval_case_creation_tracking") ||
		!h.hasColumn("approval_case_creation_tracking", "intake_id") {
		return query
	}
	return query.Where(h.completedCaseCreationPredicate())
}

// completedCaseCreationPredicate returns the shared NOT EXISTS filter that
// keeps already-created intakes off the recovery projection. It is dialect
// aware only through intakeIDMatchExpr; every other comparison stays raw so
// the same text renders on SQLite, MySQL, and PostgreSQL.
func (h *DemoAggregateHandler) completedCaseCreationPredicate() string {
	return `NOT EXISTS (
		SELECT 1 FROM approval_requests approved_creation
		JOIN approval_case_creation_tracking completed_creation
		  ON completed_creation.approval_request_id = approved_creation.id
		WHERE approved_creation.type IN ('case_creation', 'conflict_approval')
		  AND approved_creation.status = 'approved'
		  AND approved_creation.deleted_at IS NULL
		  AND completed_creation.intake_id = ` + h.intakeIDMatchExpr("case_intakes.id") + `
		  AND completed_creation.creation_status = 'completed'
	)`
}

// intakeIDMatchExpr compares the tracking intake_id column (text in every
// shipped schema) against the enclosing case_intakes id column. PostgreSQL
// keeps case_intakes.id in the uuid type, so the raw equality raised
// SQLSTATE 42883 (text = uuid) and turned both the recovery list and the
// workbench count into 500/0; casting the uuid side to text restores the
// portable comparison without converting arbitrary tracking text into uuid.
func (h *DemoAggregateHandler) intakeIDMatchExpr(intakeIDColumn string) string {
	if h.db != nil && h.db.Config != nil && h.db.Config.Dialector != nil && h.db.Config.Dialector.Name() == "postgres" {
		return "CAST(" + intakeIDColumn + " AS TEXT)"
	}
	return intakeIDColumn
}
