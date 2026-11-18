package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

func conflictApprovalWorkbenchContext(path string, userID uint) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, path, bytes.NewBuffer(nil))
	context.Set("user_id", userID)
	context.Set("role", "conflict_officer")
	context.Set("username", "独立冲突核查人")
	return context, recorder
}

func TestConflictOfficerSeesOnlyAssignedConflictApproval(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`CREATE TABLE conflict_check_records (
			check_id TEXT PRIMARY KEY, user_id INTEGER, case_name TEXT, client_id TEXT,
			client_name TEXT, case_type TEXT, check_status TEXT, risk_level TEXT,
			has_conflict BOOLEAN, check_result TEXT, search_parameters TEXT,
			created_at DATETIME, updated_at DATETIME
		)`,
		`CREATE TABLE approval_requests (
			id TEXT PRIMARY KEY, request_number TEXT, title TEXT, applicant_id TEXT,
			current_approver_id TEXT, current_approver_name TEXT, status TEXT,
			conflict_check_id TEXT, metadata TEXT, type TEXT, deleted_at DATETIME,
			created_at DATETIME
		)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.AutoMigrate(&models.Client{}, &models.Case{}, &models.CaseEthicalWallWhitelist{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.Client{ID: 1, Name: "新案客户", Type: "企业", Status: "active", Version: 1}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.Case{
		ID: 910, CaseNumber: "CASE-910", Title: "待核查案件", ClientID: 1,
		LawyerID: 34, CaseType: "commercial", Status: "active",
	}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO conflict_check_records
		(check_id, user_id, case_name, client_id, client_name, case_type, check_status,
		 risk_level, has_conflict, check_result, search_parameters)
		VALUES ('CHECK-ASSIGNED-910', 34, '待核查案件', '1', '新案客户', 'commercial',
		 'COMPLETED', 'MEDIUM', 1, '{}', '{"subjectCaseId":"910"}')`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO approval_requests
		(id, request_number, title, applicant_id, current_approver_id,
		 current_approver_name, status, conflict_check_id, metadata, type)
		VALUES ('APP-ASSIGNED-910', 'APR-ASSIGNED-910', '待处理冲突审批', '34', '35',
		 '独立冲突核查人', 'submitted', 'CHECK-ASSIGNED-910', '{}', 'conflict_approval')`).Error; err != nil {
		t.Fatal(err)
	}

	handler := NewDemoAggregateHandler(db)
	handler.SetAuthorizationService(services.NewAuthorizationService(
		repositories.NewCaseRepository(db), nil, nil, nil,
	))

	assignedContext, assignedRecorder := conflictApprovalWorkbenchContext(
		"/api/v1/dashboard/command-center?conflict_scope=all", 35,
	)
	handler.CommandCenter(assignedContext)
	if assignedRecorder.Code != http.StatusOK {
		t.Fatalf("assigned reviewer response: %d %s", assignedRecorder.Code, assignedRecorder.Body.String())
	}
	var response struct {
		Data struct {
			Summary struct {
				ActiveCases      int64 `json:"active_cases"`
				Clients          int64 `json:"clients"`
				PendingApprovals int64 `json:"pending_approvals"`
			} `json:"summary"`
			RiskQueue []map[string]interface{} `json:"risk_queue"`
		} `json:"data"`
	}
	if err := json.Unmarshal(assignedRecorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Data.Summary.ActiveCases != 0 || response.Data.Summary.Clients != 0 {
		t.Fatalf("reviewer received general matter aggregates: %+v", response.Data.Summary)
	}
	if response.Data.Summary.PendingApprovals != 1 {
		t.Fatalf("expected one assigned approval: %+v", response.Data.Summary)
	}
	if len(response.Data.RiskQueue) != 1 || response.Data.RiskQueue[0]["approval_id"] != "APP-ASSIGNED-910" {
		t.Fatalf("assigned approval link missing: %#v", response.Data.RiskQueue)
	}

	unassignedContext, _ := conflictApprovalWorkbenchContext(
		"/api/v1/dashboard/command-center?conflict_scope=all", 36,
	)
	unassignedRows := handler.riskRows(unassignedContext, 8, true)
	if len(unassignedRows) != 1 {
		t.Fatalf("review queue unexpectedly hidden: %#v", unassignedRows)
	}
	if _, leaked := unassignedRows[0]["approval_id"]; leaked {
		t.Fatalf("another reviewer's approval link leaked: %#v", unassignedRows[0])
	}
}

func TestConflictApprovalSnapshotPreservesCaseType(t *testing.T) {
	fields := conflictApprovalSnapshotFields(
		map[string]interface{}{
			"client_name": "虚构客户",
			"case_name":   "虚构设备采购争议",
			"case_type":   "civil",
		},
		map[string]interface{}{},
		map[string]interface{}{},
		nil,
	)
	config, ok := fields["case_creation_config"].(gin.H)
	if !ok {
		t.Fatalf("missing case creation config: %#v", fields)
	}
	if config["case_type"] != "civil" || config["case_name"] != "虚构设备采购争议" {
		t.Fatalf("case facts changed in approval snapshot: %#v", config)
	}
}
