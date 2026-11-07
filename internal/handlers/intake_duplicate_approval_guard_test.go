package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/security"
)

func newDuplicateApprovalGuardDB(t *testing.T) *gorm.DB {
	t.Helper()
	t.Setenv("SUBJECT_DATA_KEY", strings.Repeat("g", 32))
	dbPath := filepath.Join(t.TempDir(), fmt.Sprintf("dup_guard_%d.db", time.Now().UnixNano()))
	// busy_timeout + WAL (mattn DSN params): concurrent requests in the
	// guard tests hit the
	// same SQLite file; without a busy timeout a read racing a write fails
	// immediately instead of waiting out the write lock.
	db, err := gorm.Open(sqlite.Open("file:"+dbPath+"?_busy_timeout=5000&_journal_mode=WAL&_fk=0"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(1)
		t.Cleanup(func() { _ = sqlDB.Close() })
	}
	if err := db.AutoMigrate(
		&models.ConflictCheckRecord{},
		&models.ApprovalRequest{},
		&models.ApprovalWorkflow{},
		&models.Case{},
		&models.Client{},
		&models.User{},
		&models.Entity{},
		&models.CaseParty{},
		&models.ConflictReview{},
		&models.ConflictReviewerAssignment{},
		&models.ApprovalRecord{},
		&models.ApprovalNotification{},
		&models.ApprovalSnapshot{},
		&models.InboxItem{},
		&repositories.CaseCreationTracking{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	if err := db.Exec(`CREATE TABLE case_intakes (
		id TEXT PRIMARY KEY,
		intake_code TEXT,
		title TEXT,
		case_type TEXT,
		client_id INTEGER,
		created_by TEXT,
		priority TEXT,
		description TEXT,
		status TEXT,
		metadata TEXT,
		created_at DATETIME,
		updated_at DATETIME,
		approval_claim_token TEXT)`).Error; err != nil {
		t.Fatalf("create case_intakes: %v", err)
	}
	if err := db.Exec(`CREATE TABLE case_intake_parties (
		id INTEGER PRIMARY KEY AUTOINCREMENT, intake_id TEXT, entity_name TEXT,
		entity_type TEXT, party_role TEXT, identity_type TEXT,
		identity_number_digest TEXT, identity_number_ciphertext TEXT, metadata TEXT)`).Error; err != nil {
		t.Fatalf("create case_intake_parties: %v", err)
	}
	if err := db.Exec(`CREATE TABLE case_materials (id INTEGER PRIMARY KEY AUTOINCREMENT, intake_id TEXT, metadata TEXT)`).Error; err != nil {
		t.Fatalf("create case_materials: %v", err)
	}
	for _, user := range []models.User{
		{ID: 1, Username: "admin-guard", Name: "审批人", Email: "dup-admin@example.invalid", Password: "x", Role: "admin", Status: "active"},
		{ID: 7, Username: "lawyer-guard", Name: "虚构申请律师", Email: "dup-lawyer@example.invalid", Password: "x", Role: "lawyer", Status: "active"},
		{ID: 9, Username: "reviewer-guard", Name: "虚构独立复核人", Email: "dup-reviewer@example.invalid", Password: "x", Role: "conflict_officer", Status: "active"},
	} {
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("seed account: %v", err)
		}
	}
	client := &models.Client{ID: 1, Name: "虚构守卫客户", Type: "个人", CreatedBy: 7, IdentityType: models.IdentityTypeIDCard, IdentityNumber: "310110199001010077"}
	if err := repositories.NewClientRepository(db).Create(context.Background(), client); err != nil {
		t.Fatalf("seed protected client: %v", err)
	}
	workflow := &models.ApprovalWorkflow{
		WorkflowCode: "CASE_APPROVAL", WorkflowName: "守卫审批", WorkflowType: "CASE_APPROVAL", Status: "active", CreatedBy: "1", EffectiveDate: time.Now(),
		ApplicableTypes: "[]", ApplicableDepartments: "[]", ApplicableRoles: "[]", Conditions: "{}", Timeouts: "{}", Permissions: "{}", Notifications: "{}",
		Stages: `[{"stage_name":"partner_review","stage_order":1,"approver_id":"1","required":true}]`,
	}
	if err := db.Create(workflow).Error; err != nil {
		t.Fatalf("seed workflow: %v", err)
	}
	seedDuplicateApprovalGuardIntake(t, db, "intake-dup", "CHK-DUP-1", "重复提交守卫样本")
	return db
}

// Seed the complete, independently reviewed input protocol. Approval creation,
// freezing, inbox writes and duplicate guards still run through the real handler.
// Each clone gets its own intake-bound formal check and reviewer assignment.
func seedDuplicateApprovalGuardIntake(t *testing.T, db *gorm.DB, intakeID, checkID, title string) {
	t.Helper()
	var client models.Client
	if err := db.First(&client, 1).Error; err != nil {
		t.Fatalf("read fixture client: %v", err)
	}
	partyDigest, err := security.IdentityDigest("310110199001010088")
	if err != nil {
		t.Fatalf("digest fixture party: %v", err)
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	metadata, err := json.Marshal(map[string]interface{}{
		"draft_revision": "draft-" + intakeID, "lawyer_id": "7", "lawyer_facts_version": "facts-" + intakeID,
		"lawyer_facts_confirmed_by": "7", "lawyer_facts_confirmed_at": now.Add(-time.Minute).Format(time.RFC3339),
		"conflict_check_id": checkID, "billing_method": "hourly",
	})
	if err != nil {
		t.Fatalf("marshal intake metadata: %v", err)
	}
	if err := db.Exec(`INSERT INTO case_intakes (id, intake_code, title, case_type, client_id, created_by, priority, description, status, metadata, created_at, updated_at)
		VALUES (?, ?, ?, 'civil', 1, '7', 'medium', '虚构守卫测试', 'conflict_ready', ?, ?, ?)`, intakeID, intakeID, title, string(metadata), now, now).Error; err != nil {
		t.Fatalf("seed intake %s: %v", intakeID, err)
	}
	if err := db.Exec(`INSERT INTO case_intake_parties (intake_id, entity_name, entity_type, party_role, identity_type, identity_number_digest, metadata)
		VALUES (?, '虚构对方', 'INDIVIDUAL', 'OPPOSING_PARTY', 'ID_CARD', ?, '{"localId":"opposing-1"}')`, intakeID, partyDigest).Error; err != nil {
		t.Fatalf("seed party %s: %v", intakeID, err)
	}
	check := &models.ConflictCheckRecord{
		CheckID: checkID, ClientID: "1", ClientName: client.Name, CaseName: title, CaseType: "civil",
		CheckStatus: "COMPLETED", UserID: 7, RiskLevel: "LOW", CheckTime: now,
		SearchParameters: models.JSON{"intakeId": intakeID, "clientType": "PERSON", "clientIdentifiers": map[string]interface{}{"id_card": "hmac-sha256:" + client.IdentityNumberDigest},
			"parties": []interface{}{map[string]interface{}{"name": "虚构对方", "role": "OPPOSING_PARTY", "entityType": "INDIVIDUAL", "identifiers": map[string]interface{}{"id_card": "hmac-sha256:" + partyDigest}}}},
		CheckResult: models.JSON{"checkId": checkID, "hasConflict": false, "decision": map[string]interface{}{"coverageStatus": "COMPLETE"}},
	}
	if err := db.Create(check).Error; err != nil {
		t.Fatalf("seed formal check %s: %v", checkID, err)
	}
	if err := db.Create(&models.ConflictReviewerAssignment{ID: uuid.NewString(), CheckID: checkID, ReviewerID: 9, AssignedBy: 1,
		Status: models.ConflictReviewerAssignmentActive, RecusalDeclared: true, IndependenceReason: "虚构独立复核声明", CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("seed independent assignment %s: %v", checkID, err)
	}
	// Complete no-conflict checks have an empty evidence list; the protocol
	// hashes its canonical JSON [] with SHA-256 (conflictEvidenceHash).
	evidenceHash := fmt.Sprintf("%x", sha256.Sum256([]byte("[]")))
	if err := db.Create(&models.ConflictReview{CheckID: checkID, Decision: "no_conflict", Notes: "虚构完整覆盖复核", ReviewerID: 9,
		ReviewerName: "虚构独立复核人", EvidenceHash: evidenceHash, CreatedAt: now.Add(time.Second)}).Error; err != nil {
		t.Fatalf("seed independent review %s: %v", checkID, err)
	}
}

func newDuplicateApprovalGuardRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := newIntegrationHandlerForTest(db)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", c.GetHeader("X-Test-Role"))
		c.Set("user_id", c.GetHeader("X-Test-User-Id"))
		c.Next()
	})
	r.POST("/api/v1/integration/approvals/with-conflict", h.CreateApprovalWithConflict)
	return r
}

func postWithConflict(t *testing.T, r *gin.Engine, body string) (int, map[string]interface{}) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/integration/approvals/with-conflict", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Role", "lawyer")
	req.Header.Set("X-Test-User-Id", "7")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	return w.Code, resp
}

func TestIntakeDuplicateApprovalGuard(t *testing.T) {
	orig := newIntegrationHandlerForTest
	newIntegrationHandlerForTest = func(db *gorm.DB) *IntegrationHandler {
		return orig(db)
	}
	t.Cleanup(func() { newIntegrationHandlerForTest = orig })
	db := newDuplicateApprovalGuardDB(t)
	r := newDuplicateApprovalGuardRouter(db)
	body := fmt.Sprintf(`{
		"type": "case_creation",
		"title": "重复提交守卫样本",
		"content": "守卫样本",
		"applicant_name": "qa_lawyer_a",
		"department_name": "公司业务部",
		"workflow_type": "CASE_APPROVAL",
		"metadata": {
			"intake_id": "intake-dup",
			"conflict_check_id": "CHK-DUP-1"
		}
	}`)
	code, resp := postWithConflict(t, r, body)
	if code != http.StatusOK {
		t.Fatalf("first submission must pass the guard, got %d: %v", code, resp)
	}
	code2, resp2 := postWithConflict(t, r, body)
	if code2 != http.StatusConflict {
		t.Fatalf("duplicate submission must be rejected 409, got %d: %v", code2, resp2)
	}
	errCode, _ := resp2["code"].(string)
	if errCode == "" {
		if errObj, ok := resp2["error"].(map[string]interface{}); ok {
			errCode, _ = errObj["code"].(string)
		}
	}
	if errCode != "INTAKE_APPROVAL_ALREADY_IN_PROGRESS" {
		t.Fatalf("expected INTAKE_APPROVAL_ALREADY_IN_PROGRESS, got %q (payload: %v)", errCode, resp2)
	}
	var count int64
	db.Table("approval_requests").Count(&count)
	if count != 1 {
		t.Fatalf("exactly one approval must exist after duplicate rejection, got %d", count)
	}
}
