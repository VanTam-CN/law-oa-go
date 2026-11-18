package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newIntakeDraftListRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewDemoAggregateHandler(db)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", c.GetHeader("X-Test-Role"))
		c.Set("user_id", c.GetHeader("X-Test-User-Id"))
		c.Next()
	})
	r.GET("/api/v1/case-intakes", h.ListIntakeDrafts)
	return r
}

func newIntakeDraftListDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`CREATE TABLE case_intakes (
		id TEXT PRIMARY KEY, intake_code TEXT, title TEXT, case_type TEXT,
		client_id INTEGER, created_by TEXT, status TEXT, metadata TEXT,
		created_at DATETIME, updated_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create table: %v", err)
	}
	seed := []struct{ id, code, title, createdBy, status, updatedAt string }{
		{"i-1", "INT-001", "张三买卖合同纠纷", "7", "draft", "2026-09-10 08:00:00"},
		{"i-2", "INT-002", "李四劳动争议", "7", "draft", "2026-09-09 08:00:00"},
		{"i-3", "INT-003", "王五侵权案", "8", "draft", "2026-09-10 09:00:00"},
		{"i-4", "INT-004", "赵六咨询", "7", "lawyer_facts_confirmed", "2026-09-10 07:00:00"},
	}
	for _, s := range seed {
		if err := db.Exec(
			`INSERT INTO case_intakes (id, intake_code, title, case_type, client_id, created_by, status, metadata, created_at, updated_at)
			 VALUES (?, ?, ?, 'civil', 3, ?, ?, '{}', ?, ?)`,
			s.id, s.code, s.title, s.createdBy, s.status, s.updatedAt, s.updatedAt,
		).Error; err != nil {
			t.Fatalf("seed %s: %v", s.id, err)
		}
	}
	return db
}

// C-item regression: an intake whose active approval has already created a
// formal case must leave the recovery list, while an approval that reached
// approved without completed case creation (or is still live) must stay
// recoverable. The predicate stays intake-scope filtered (metadata.intake_id)
// so unrelated approvals can never hide a lawyer's own in-progress draft.
func TestListIntakeDraftsExcludesIntakesWithCompletedCaseCreation(t *testing.T) {
	db := newIntakeDraftListDB(t)
	statements := []string{
		"CREATE TABLE approval_requests (id TEXT PRIMARY KEY, type TEXT, status TEXT, metadata TEXT, deleted_at DATETIME)",
		"CREATE TABLE approval_case_creation_tracking (id TEXT PRIMARY KEY, approval_request_id TEXT, intake_id TEXT, creation_status TEXT)",
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("create approval table: %v", err)
		}
	}
	// The shared fixture seeds i-3 under lawyer 8; this scenario must run as
	// the owning lawyer 7, otherwise the live-approval expectation is never
	// exercised by the owner-scoped query.
	if err := db.Exec("UPDATE case_intakes SET created_by = '7' WHERE id = 'i-3'").Error; err != nil {
		t.Fatalf("assign i-3 to lawyer 7: %v", err)
	}
	type seed struct {
		name     string
		approval string
		tracking string
		intakeID string
		hidden   bool
	}
	seeds := []seed{
		{"completed case creation hides the draft", "INSERT INTO approval_requests (id, type, status, metadata) VALUES ('ap-done', 'case_creation', 'approved', '{\"intake_id\":\"i-1\"}')", "INSERT INTO approval_case_creation_tracking (id, approval_request_id, intake_id, creation_status) VALUES ('tr-done', 'ap-done', 'i-1', 'completed')", "i-1", true},
		{"failed case creation keeps the draft recoverable", "INSERT INTO approval_requests (id, type, status, metadata) VALUES ('ap-fail', 'case_creation', 'approved', '{\"intake_id\":\"i-2\"}')", "INSERT INTO approval_case_creation_tracking (id, approval_request_id, intake_id, creation_status) VALUES ('tr-fail', 'ap-fail', 'i-2', 'failed')", "i-2", false},
		{"live approval keeps the draft recoverable", "INSERT INTO approval_requests (id, type, status, metadata) VALUES ('ap-live', 'case_creation', 'under_review', '{\"intake_id\":\"i-3\"}')", "", "i-3", false},
	}
	for _, item := range seeds {
		if err := db.Exec(item.approval).Error; err != nil {
			t.Fatalf("seed %s approval: %v", item.name, err)
		}
		if item.tracking != "" {
			if err := db.Exec(item.tracking).Error; err != nil {
				t.Fatalf("seed %s tracking: %v", item.name, err)
			}
		}
	}
	// i-4 (lawyer_facts_confirmed, no approval at all) must stay listed and
	// guards against an accidental cross-intake leak of the exclusion.

	r := newIntakeDraftListRouter(db)
	code, data := listIntakeDrafts(t, r, "lawyer", "7", "")
	if code != http.StatusOK {
		t.Fatalf("owner list: status=%d want 200", code)
	}
	ids := map[string]bool{}
	for _, raw := range data["items"].([]interface{}) {
		row, _ := raw.(map[string]interface{})
		ids[fmt.Sprint(row["id"])] = true
	}
	for _, item := range seeds {
		if item.hidden == ids[item.intakeID] {
			t.Fatalf("%s: intake %s listed=%v want listed=%v (ids=%v)", item.name, item.intakeID, ids[item.intakeID], !item.hidden, ids)
		}
	}
	if !ids["i-4"] {
		t.Fatalf("unlinked in-progress intake i-4 missing from recovery drafts (ids=%v)", ids)
	}
}

// The A3 real PostgreSQL probe caught the shared recovery predicate raising
// SQLSTATE 42883 because case_intakes.id is uuid while the tracking intake_id
// column is text. These regressions keep the dialect boundary observable:
// the generated SQL must cast only the uuid-side intake id on PostgreSQL, and
// uuid-shaped identifiers must keep their real list/count semantics.
func TestIntakeIDMatchExprUsesPortableComparisonPerDialect(t *testing.T) {
	sqliteDB := newIntakeDraftListDB(t)
	sqliteHandler := NewDemoAggregateHandler(sqliteDB)
	if got, want := sqliteHandler.intakeIDMatchExpr("case_intakes.id"), "case_intakes.id"; got != want {
		t.Fatalf("sqlite predicate must keep the plain text comparison: got %q want %q", got, want)
	}
	predicate := sqliteHandler.completedCaseCreationPredicate()
	if !strings.Contains(predicate, "completed_creation.intake_id = case_intakes.id") {
		t.Fatalf("non-postgres predicate lost the direct id equality: %s", predicate)
	}
	postgresHandler := &DemoAggregateHandler{db: &gorm.DB{Config: &gorm.Config{Dialector: uuidShapeDialector{}}}}
	if got, want := postgresHandler.intakeIDMatchExpr("case_intakes.id"), "CAST(case_intakes.id AS TEXT)"; got != want {
		t.Fatalf("postgres predicate must cast the uuid-side intake id to text: got %q want %q", got, want)
	}
	postgresPredicate := postgresHandler.completedCaseCreationPredicate()
	if !strings.Contains(postgresPredicate, "completed_creation.intake_id = CAST(case_intakes.id AS TEXT)") {
		t.Fatalf("postgres predicate lost the CAST comparison: %s", postgresPredicate)
	}
	if strings.Contains(postgresPredicate, "CAST(completed_creation.intake_id") || strings.Contains(postgresPredicate, "::uuid") {
		t.Fatalf("postgres predicate must not wrap or convert the text tracking side: %s", postgresPredicate)
	}
}

func listIntakeDrafts(t *testing.T, r *gin.Engine, role, userID, query string) (int, gin.H) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/case-intakes"+query, nil)
	req.Header.Set("X-Test-Role", role)
	req.Header.Set("X-Test-User-Id", userID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var payload struct {
		Data gin.H `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &payload)
	return w.Code, payload.Data
}

func TestListIntakeDraftsScopesAndFields(t *testing.T) {
	r := newIntakeDraftListRouter(newIntakeDraftListDB(t))

	code, data := listIntakeDrafts(t, r, "lawyer", "7", "")
	if code != http.StatusOK {
		t.Fatalf("owner list: status=%d want 200", code)
	}
	items, _ := data["items"].([]interface{})
	// Recovery contract (defect-3 fix): in-progress intakes such as
	// lawyer_facts_confirmed must stay discoverable after the browser
	// localStorage snapshot is gone. Only terminal states stay out.
	if len(items) != 3 {
		t.Fatalf("owner list: items=%d want 3 (own drafts + own in-progress confirmed intake)", len(items))
	}
	statuses := map[string]bool{}
	for _, raw := range items {
		row, _ := raw.(map[string]interface{})
		status, _ := row["status"].(string)
		statuses[status] = true
		for _, forbidden := range []string{"metadata", "parties", "materials"} {
			if _, exists := row[forbidden]; exists {
				t.Fatalf("owner list: forbidden field %q leaked", forbidden)
			}
		}
	}
	if !statuses["lawyer_facts_confirmed"] {
		t.Fatalf("owner list: in-progress lawyer_facts_confirmed intake missing from recovery list")
	}

	code, data = listIntakeDrafts(t, r, "director", "9", "")
	if code != http.StatusOK {
		t.Fatalf("management list: status=%d want 200", code)
	}
	items, _ = data["items"].([]interface{})
	if len(items) != 4 {
		t.Fatalf("management list: items=%d want 4", len(items))
	}
	code, _ = listIntakeDrafts(t, r, "director", "9", "?page=2&page_size=2")
	if code != http.StatusOK {
		t.Fatalf("pagination: status=%d want 200", code)
	}
	code, _ = listIntakeDrafts(t, r, "director", "9", "?page_size=999")
	if code != http.StatusOK {
		t.Fatalf("oversized page_size: status=%d want 200", code)
	}
}

// Regression for the main-review probe TestIndependentA3LegacyTracking...:
// a legacy tracking schema without intake_id must degrade to a no-op so the
// recovery endpoint returns 200 instead of a 500 SQL error.
func TestListIntakeDraftsToleratesLegacyTrackingWithoutIntakeColumn(t *testing.T) {
	db := newIntakeDraftListDB(t)
	statements := []string{
		"CREATE TABLE approval_requests (id TEXT PRIMARY KEY, type TEXT, status TEXT, metadata TEXT, deleted_at DATETIME)",
		"CREATE TABLE approval_case_creation_tracking (id TEXT PRIMARY KEY, approval_request_id TEXT, creation_status TEXT)",
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("create approval table: %v", err)
		}
	}
	r := newIntakeDraftListRouter(db)
	code, data := listIntakeDrafts(t, r, "lawyer", "7", "")
	if code != http.StatusOK {
		t.Fatalf("legacy tracking schema: status=%d want 200 (recovery must not 500)", code)
	}
	items, _ := data["items"].([]interface{})
	if len(items) != 3 {
		t.Fatalf("legacy tracking schema: items=%d want 3", len(items))
	}
}

// Regression for the main-review probe TestIndependentA3CompletedConflictCase...:
// a conflict_approval that already completed its formal-case binding must
// leave the recovery list, mirroring RequiresCaseCreationApproval's first two
// branches. Metadata-config approvals cannot be matched portably in SQL and
// stay listed; the recovery list errs toward over-listing rather than hiding
// a still-recoverable draft.
func TestListIntakeDraftsExcludesCompletedConflictCaseApproval(t *testing.T) {
	db := newIntakeDraftListDB(t)
	statements := []string{
		"CREATE TABLE approval_requests (id TEXT PRIMARY KEY, type TEXT, status TEXT, metadata TEXT, deleted_at DATETIME)",
		"CREATE TABLE approval_case_creation_tracking (id TEXT PRIMARY KEY, approval_request_id TEXT, intake_id TEXT, case_id TEXT, creation_status TEXT)",
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("create approval table: %v", err)
		}
	}
	if err := db.Exec(`INSERT INTO approval_requests (id, type, status, metadata) VALUES ('ap', 'conflict_approval', 'approved', '{"intake_id":"i-1"}')`).Error; err != nil {
		t.Fatalf("seed conflict approval: %v", err)
	}
	if err := db.Exec("INSERT INTO approval_case_creation_tracking (id, approval_request_id, intake_id, case_id, creation_status) VALUES ('tr', 'ap', 'i-1', '101', 'completed')").Error; err != nil {
		t.Fatalf("seed tracking: %v", err)
	}
	r := newIntakeDraftListRouter(db)
	code, data := listIntakeDrafts(t, r, "lawyer", "7", "")
	if code != http.StatusOK {
		t.Fatalf("conflict-approval list: status=%d want 200", code)
	}
	ids := map[string]bool{}
	for _, raw := range data["items"].([]interface{}) {
		row, _ := raw.(map[string]interface{})
		ids[fmt.Sprint(row["id"])] = true
	}
	if ids["i-1"] {
		t.Fatalf("completed conflict_approval binding still advertised as draft: ids=%v", ids)
	}
	if !ids["i-2"] || !ids["i-4"] {
		t.Fatalf("unrelated drafts must stay recoverable: ids=%v", ids)
	}
}
