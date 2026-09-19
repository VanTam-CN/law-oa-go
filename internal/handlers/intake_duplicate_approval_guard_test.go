package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

func newDuplicateApprovalGuardDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), fmt.Sprintf("dup_guard_%d.db", time.Now().UnixNano()))
	db, err := gorm.Open(sqlite.Open("file:"+dbPath), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
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
		status TEXT,
		metadata TEXT,
		created_at DATETIME,
		updated_at DATETIME)`).Error; err != nil {
		t.Fatalf("create case_intakes: %v", err)
	}
	const checkResult = `{"coverage_status":"COMPLETE","parties":[]}`
	if err := db.Create(&models.User{Username: "admin-guard", Name: "审批人", Email: "dup-admin@example.com", Password: "x", Role: "admin", Status: "active"}).Error; err != nil {
		t.Fatalf("seed approver: %v", err)
	}
	if err := db.Exec(`INSERT INTO case_intakes (id, title, case_type, client_id, created_by, status, metadata) VALUES
		('intake-dup', '重复提交守卫样本', 'civil', 1, '7', 'conflict_ready', ?)`,
		`{"conflict_check_id":"CHK-DUP-1"}`).Error; err != nil {
		t.Fatalf("seed intake: %v", err)
	}
	if err := db.Exec(`INSERT INTO conflict_check_records (check_id, check_status, check_result) VALUES
		('CHK-DUP-1', 'COMPLETED', ?)`, checkResult).Error; err != nil {
		t.Fatalf("seed check: %v", err)
	}
	return db
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
