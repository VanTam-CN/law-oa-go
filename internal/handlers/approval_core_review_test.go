package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"path/filepath"
	"reflect"
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
	"law-oa-go/internal/services"
)

// Every test owns its database; these fixtures do not use the shared PG harness.
func coreReviewDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "review.db")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&models.ApprovalRequest{}, &models.ApprovalRecord{}, &models.ApprovalNotification{}, &models.ApprovalSnapshot{}, &models.ApprovalWorkflow{}, &models.User{}, &models.ConflictCheckRecord{}); err != nil {
		t.Fatal(err)
	}
	for _, user := range []models.User{
		{ID: 7, Username: "review-applicant", Name: "虚构申请人", Email: "applicant@example.invalid", Password: "test", Role: "lawyer", Status: "active"},
		{ID: 9, Username: "review-approver", Name: "虚构审批人", Email: "approver@example.invalid", Password: "test", Role: "admin", Status: "active"},
	} {
		if err := db.Create(&user).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func coreReviewApproval(t *testing.T, db *gorm.DB, status, metadata string) *models.ApprovalRequest {
	t.Helper()
	row := &models.ApprovalRequest{ID: uuid.NewString(), RequestNumber: uuid.NewString(), Title: "虚构审批", Type: "ordinary", Content: "申请", ApplicantID: "7", ApplicantName: "申请人", CurrentApproverID: "9", WorkflowType: "STANDARD_APPROVAL", Metadata: metadata, Status: status, CreatedBy: "7"}
	if err := db.Create(row).Error; err != nil {
		t.Fatal(err)
	}
	return row
}

func coreReviewHTTP(t *testing.T, handler gin.HandlerFunc, method, target, id, body, role string, actor uint) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Params = gin.Params{{Key: "id", Value: id}}
	c.Set("user_id", actor)
	c.Set("role", role)
	c.Request = httptest.NewRequest(method, target, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	handler(c)
	return recorder
}

func coreReviewData(t *testing.T, response *httptest.ResponseRecorder, status int) map[string]interface{} {
	t.Helper()
	if response.Code != status {
		t.Fatalf("HTTP %d, want %d: %s", response.Code, status, response.Body.String())
	}
	var envelope map[string]interface{}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	data, _ := envelope["data"].(map[string]interface{})
	return data
}

const coreReviewSecrets = `{"parties":[{"name":"当前主体","identity_number_masked":"31***77","identifiers":{"id_card":"hmac-sha256:PARTY-SECRET"}}],"search_parameters":"{\"candidate\":\"hmac-sha256:SEARCH-SECRET\",\"identity_number_digest\":\"DIGEST-SECRET\",\"name\":\"当前主体\"}","conflict_cases":[{"case_no":"CASE-HIDDEN","case_name":"历史秘密"}],"identity_number_ciphertext":"CIPHER-SECRET"}`

func coreReviewNoSecrets(t *testing.T, raw string, role string) {
	t.Helper()
	for _, secret := range []string{"hmac-sha256", "PARTY-SECRET", "SEARCH-SECRET", "DIGEST-SECRET", "CIPHER-SECRET"} {
		if strings.Contains(raw, secret) {
			t.Fatalf("protected value %q leaked: %s", secret, raw)
		}
	}
	if role == "lawyer" && (strings.Contains(raw, "CASE-HIDDEN") || strings.Contains(raw, "历史秘密")) {
		t.Fatalf("historical evidence leaked: %s", raw)
	}
	if !strings.Contains(raw, "当前主体") || !strings.Contains(raw, "31***77") {
		t.Fatalf("current masked subject lost: %s", raw)
	}
}

func TestCoreReviewProjectionProtectsAllRolesAndCopies(t *testing.T) {
	for _, role := range []string{"lawyer", "conflict_officer"} {
		t.Run(role, func(t *testing.T) {
			approval := &models.ApprovalRequest{Metadata: coreReviewSecrets, ActionVersion: "round-2"}
			payload := map[string]interface{}{"approval": approval, "snapshot": json.RawMessage(coreReviewSecrets), "large_id": uint64(9007199254740993)}
			raw, err := json.Marshal(projectApprovalResponseForViewer(payload, role))
			if err != nil {
				t.Fatal(err)
			}
			coreReviewNoSecrets(t, string(raw), role)
			if approval.Metadata != coreReviewSecrets {
				t.Fatal("projection mutated source")
			}
			if !strings.Contains(string(raw), "9007199254740993") || !strings.Contains(string(raw), "round-2") {
				t.Fatal("projection changed version/ID")
			}
			if role == "conflict_officer" && !strings.Contains(string(raw), "CASE-HIDDEN") {
				t.Fatal("reviewer lost permitted evidence")
			}
			projected := projectApprovalForViewer(approval, role)
			if projected == nil || !json.Valid([]byte(projected.Metadata)) {
				t.Fatal("metadata string contract lost")
			}
		})
	}
}

func TestCoreReviewCreateBlocksControlledMetadata(t *testing.T) {
	db := coreReviewDB(t)
	h := NewApprovalHandler(db)
	metadata := []string{
		`{"intake_id":"i"}`, `{"conflict_check_id":"c"}`, `{"conflict_task_id":"c"}`,
		`{"conflict_result":{"checkId":"c"}}`, `{"conflict_result":{"check_id":"c"}}`, `{"conflict_result":{"id":"c"}}`,
		`{"conflict_check":{"conflict_check_id":"c"}}`, `{"conflict_check":{"check_id":"c"}}`, `{"conflict_check":{"checkId":"c"}}`,
		`{"approval_snapshot":{"metadata":{"intake_id":"i"}}}`, `{"active_submission_snapshot_id":"s"}`,
		`{"case_creation_config":{}}`, `{"integration_type":"both"}`,
	}
	for _, endpoint := range []struct {
		name    string
		handler gin.HandlerFunc
	}{{"create", h.CreateApproval}, {"template", h.CreateFromTemplate}} {
		for i, input := range metadata {
			t.Run(fmt.Sprintf("%s/%d", endpoint.name, i), func(t *testing.T) {
				body := `{"template_name":"ordinary","title":"正常名称","type":"ordinary","content":"申请","metadata":` + input + `}`
				response := coreReviewHTTP(t, endpoint.handler, "POST", "/approvals", "", body, "lawyer", 7)
				coreReviewData(t, response, 409)
				if !strings.Contains(response.Body.String(), "APPROVAL_CONTROLLED_WORKFLOW_REQUIRED") {
					t.Fatal(response.Body.String())
				}
			})
		}
	}
	var count int64
	if err := db.Model(&models.ApprovalRequest{}).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("rejected create wrote %d rows: %v", count, err)
	}
}

func TestCoreReviewMutationResponsesAndUpdateErrors(t *testing.T) {
	db := coreReviewDB(t)
	h := NewApprovalHandler(db)
	response := coreReviewHTTP(t, h.CreateApproval, "POST", "/approvals", "", `{"title":"申请","type":"ordinary","content":"申请","metadata":`+coreReviewSecrets+`}`, "lawyer", 7)
	data := coreReviewData(t, response, 200)
	coreReviewNoSecrets(t, response.Body.String(), "lawyer")
	id := data["id"].(string)
	response = coreReviewHTTP(t, h.UpdateApproval, "PUT", "/approvals/"+id, id, `{"title":"已修改"}`, "lawyer", 7)
	coreReviewData(t, response, 200)
	coreReviewNoSecrets(t, response.Body.String(), "lawyer")
	var saved models.ApprovalRequest
	if err := db.First(&saved, "id = ?", id).Error; err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(saved.Metadata, "PARTY-SECRET") {
		t.Fatal("response projection changed stored metadata")
	}
	response = coreReviewHTTP(t, h.UpdateApproval, "PUT", "/approvals/"+id, id, `{"title":"非申请人修改"}`, "conflict_officer", 9)
	coreReviewData(t, response, 403)
	response = coreReviewHTTP(t, h.SubmitApproval, "POST", "/approvals/"+id, id, `{}`, "lawyer", 7)
	coreReviewData(t, response, 200)
	coreReviewNoSecrets(t, response.Body.String(), "lawyer")
	// Ordinary business returns keep their original non-structured requirements.
	response = coreReviewHTTP(t, h.ProcessApprovalDecision, "POST", "/decision", id, `{"decision":"request_changes","decision_reason":"补说明","imposed_requirements":{"business_note":"原业务协议"}}`, "conflict_officer", 9)
	coreReviewData(t, response, 200)
	coreReviewNoSecrets(t, response.Body.String(), "conflict_officer")
	var record models.ApprovalRecord
	if err := db.Where("approval_request_id = ?", id).First(&record).Error; err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(record.ImposedRequirements, "business_note") {
		t.Fatal("ordinary business protocol was changed")
	}
	response = coreReviewHTTP(t, h.ResubmitApproval, "POST", "/resubmit", id, `{"revision_note":"已补说明","ordinary_extra":true}`, "lawyer", 7)
	coreReviewData(t, response, 200)
	coreReviewNoSecrets(t, response.Body.String(), "lawyer")
	response = coreReviewHTTP(t, h.UpdateApproval, "PUT", "/approvals/"+id, id, `{"title":"不应改"}`, "lawyer", 7)
	coreReviewData(t, response, 400)
	linked := coreReviewApproval(t, db, models.ApprovalStatusNeedsRevision, `{"intake_id":"intake"}`)
	response = coreReviewHTTP(t, h.UpdateApproval, "PUT", "/approvals/"+linked.ID, linked.ID, `{"metadata":{}}`, "lawyer", 7)
	coreReviewData(t, response, 409)
	if !strings.Contains(response.Body.String(), "INTAKE_REVISION_ROUTE_REQUIRED") {
		t.Fatal(response.Body.String())
	}
	var after models.ApprovalRequest
	if err := db.First(&after, "id = ?", linked.ID).Error; err != nil || after.Metadata != linked.Metadata {
		t.Fatalf("rejected update changed metadata: %v", err)
	}
}

func coreReviewLegacy(t *testing.T, db *gorm.DB, clientIDs ...uint) (*models.ApprovalRequest, models.ApprovalSnapshot) {
	t.Helper()
	facts := map[string]interface{}{"intake_id": "intake-legacy", "title": "冻结事实"}
	if len(clientIDs) > 0 {
		facts["client"] = map[string]interface{}{"id": clientIDs[0]}
	}
	snapshot := map[string]interface{}{"snapshot_id": "SNAP_original", "metadata": facts}
	metadata := map[string]interface{}{"intake_id": "intake-legacy", "title": "冻结事实", "approval_snapshot": snapshot}
	if len(clientIDs) > 0 {
		metadata["client"] = facts["client"]
	}
	approval := coreReviewApproval(t, db, models.ApprovalStatusNeedsRevision, jsonStringValue(metadata))
	row := models.ApprovalSnapshot{ID: uuid.NewString(), ApprovalRequestID: approval.ID, SnapshotType: "conflict_approval", SourceVersion: 1, SnapshotData: jsonStringValue(snapshot)}
	if err := db.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
	return approval, row
}

func coreReviewReturn(t *testing.T, db *gorm.DB, approval *models.ApprovalRequest, reason, imposed string, at time.Time) models.ApprovalRecord {
	t.Helper()
	row := models.ApprovalRecord{ID: uuid.NewString(), ApprovalRequestID: approval.ID, Decision: models.ApprovalDecisionRequestChanges, DecisionReason: reason, ImposedRequirements: imposed, CreatedAt: at, ApproverID: "9", SubmissionSnapshotID: approval.ActiveSubmissionSnapshotID}
	if err := db.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
	return row
}

func TestCoreReviewDetailSupplementContextReadOnly(t *testing.T) {
	for _, structured := range []bool{false, true} {
		t.Run(fmt.Sprint(structured), func(t *testing.T) {
			db := coreReviewDB(t)
			approval, snapshot := coreReviewLegacy(t, db)
			now := time.Now().UTC().Truncate(time.Second)
			coreReviewReturn(t, db, approval, "上一轮意见", `{}`, now.Add(-time.Hour))
			imposed := ` {"business_note":"历史字节保留"} `
			if structured {
				var err error
				imposed, err = services.NormalizeSupplementRequirements(map[string]interface{}{"schema_version": 1, "supplement_requests": []map[string]interface{}{{"category": "matter_explanation", "field_key": "description", "description": "补充说明", "required": true}}}, "最新退回意见")
				if err != nil {
					t.Fatal(err)
				}
			}
			record := coreReviewReturn(t, db, approval, "最新退回意见", imposed, now)
			wanted, err := services.ReadSupplementRequests(&record)
			if err != nil {
				t.Fatal(err)
			}
			h := NewApprovalHandler(db)
			for i := 0; i < 2; i++ {
				response := coreReviewHTTP(t, h.GetApproval, "GET", "/approvals/"+approval.ID, approval.ID, "", "lawyer", 7)
				data := coreReviewData(t, response, 200)
				var got []models.SupplementRequest
				if err := json.Unmarshal([]byte(jsonStringValue(data["supplement_requests"])), &got); err != nil {
					t.Fatal(err)
				}
				if data["return_record_id"] != record.ID || data["return_reason"] != record.DecisionReason || !reflect.DeepEqual(got, wanted) {
					t.Fatalf("wrong top-level return context: %s", response.Body.String())
				}
			}
			var persisted models.ApprovalRecord
			if err := db.First(&persisted, "id = ?", record.ID).Error; err != nil || persisted.ImposedRequirements != imposed {
				t.Fatalf("legacy return rewritten: %v", err)
			}
			var frozen models.ApprovalSnapshot
			if err := db.First(&frozen, "id = ?", snapshot.ID).Error; err != nil || frozen.SnapshotData != snapshot.SnapshotData {
				t.Fatalf("snapshot rewritten: %v", err)
			}
			var subject models.ApprovalRequest
			if err := db.First(&subject, "id = ?", approval.ID).Error; err != nil || subject.Metadata != approval.Metadata || subject.ActiveSubmissionSnapshotID != nil {
				t.Fatalf("approval rewritten: %v", err)
			}
		})
	}
}

func TestCoreReviewDetailAuthorizationAndAmbiguousReturn(t *testing.T) {
	db := coreReviewDB(t)
	approval, _ := coreReviewLegacy(t, db)
	now := time.Now().UTC().Truncate(time.Second)
	coreReviewReturn(t, db, approval, "秘密退回意见", `{}`, now)
	coreReviewReturn(t, db, approval, "另一条意见", `{}`, now)
	h := NewApprovalHandler(db)
	response := coreReviewHTTP(t, h.GetApproval, "GET", "/approvals/"+approval.ID, approval.ID, "", "lawyer", 8)
	coreReviewData(t, response, 403)
	if strings.Contains(response.Body.String(), "意见") {
		t.Fatal("unauthorized return leaked")
	}
	response = coreReviewHTTP(t, h.GetApproval, "GET", "/approvals/"+approval.ID, approval.ID, "", "lawyer", 7)
	coreReviewData(t, response, 409)
	if !strings.Contains(response.Body.String(), "APPROVAL_ROUND_STALE") {
		t.Fatal(response.Body.String())
	}
}

func TestCoreReviewResubmitUsesDurableIntakeAssociation(t *testing.T) {
	db := coreReviewDB(t)
	approval := coreReviewApproval(t, db, models.ApprovalStatusNeedsRevision, `{}`)
	row := models.ApprovalSnapshot{ID: uuid.NewString(), ApprovalRequestID: approval.ID, SnapshotType: "conflict_approval", SourceVersion: 1, SnapshotData: `{"metadata":{"intake_id":"intake-original"}}`}
	if err := db.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
	h := NewApprovalHandler(db)
	// Unknown fields would be ignored by the old note-only branch. Durable
	// history must select the strict DTO before any transition can occur.
	response := coreReviewHTTP(t, h.ResubmitApproval, "POST", "/resubmit", approval.ID, `{"revision_note":"旧页","metadata":{"intake_id":""}}`, "lawyer", 7)
	coreReviewData(t, response, 400)
	response = coreReviewHTTP(t, h.ResubmitApproval, "POST", "/resubmit", approval.ID, `{"revision_note":"旧页"}`, "lawyer", 7)
	coreReviewData(t, response, 400)
	if !strings.Contains(response.Body.String(), "INTAKE_RESUBMIT_VERSION_REQUIRED") {
		t.Fatal(response.Body.String())
	}
	var saved models.ApprovalRequest
	if err := db.First(&saved, "id = ?", approval.ID).Error; err != nil || saved.Status != approval.Status || saved.Metadata != approval.Metadata {
		t.Fatalf("legacy route bypassed durable history: %v", err)
	}
}

func TestCoreReviewSnapshotHistoryAndMissing(t *testing.T) {
	db := coreReviewDB(t)
	approval, legacy := coreReviewLegacy(t, db, 7)
	activeID := uuid.NewString()
	active := models.ApprovalSnapshot{ID: activeID, ApprovalRequestID: approval.ID, SnapshotType: models.IntakeSubmissionSnapshotType, SourceVersion: 2, SnapshotData: jsonStringValue(map[string]interface{}{"snapshot_id": activeID, "approval_request_id": approval.ID, "source_version": 2, "intake_id": "intake-legacy", "conflict_check_id": "check-new", "metadata": map[string]interface{}{"intake_id": "intake-legacy", "conflict_check_id": "check-new", "title": "当前轮次"}})}
	if err := db.Create(&active).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(approval).Update("active_submission_snapshot_id", activeID).Error; err != nil {
		t.Fatal(err)
	}
	h := coreReviewSnapshotHandler(t, db)
	for _, version := range []struct{ id, title string }{{"", "当前轮次"}, {activeID, "当前轮次"}, {legacy.ID, "冻结事实"}} {
		t.Run("version/"+version.id, func(t *testing.T) {
			response := coreReviewHTTP(t, h.GetApprovalSnapshot, "GET", "/snapshot?snapshot_id="+version.id, approval.ID, "", "lawyer", 7)
			data := coreReviewData(t, response, 200)
			viewed := version.id
			if viewed == "" {
				viewed = activeID
			}
			if data["active_submission_snapshot_id"] != activeID || data["viewed_snapshot_id"] != viewed || !strings.Contains(jsonStringValue(data["snapshot"]), version.title) {
				t.Fatalf("wrong history selection: %s", response.Body.String())
			}
			if version.id == legacy.ID && !strings.Contains(jsonStringValue(data["snapshot"]), "SNAP_original") {
				t.Fatal("legacy JSON identifier overwritten")
			}
		})
	}
	other := coreReviewApproval(t, db, models.ApprovalStatusDraft, `{}`)
	foreign := models.ApprovalSnapshot{ID: uuid.NewString(), ApprovalRequestID: other.ID, SnapshotType: "conflict_approval", SourceVersion: 1, SnapshotData: `{"title":"other private snapshot"}`}
	if err := db.Create(&foreign).Error; err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{foreign.ID, uuid.NewString(), "invalid"} {
		response := coreReviewHTTP(t, h.GetApprovalSnapshot, "GET", "/snapshot?snapshot_id="+id, approval.ID, "", "lawyer", 7)
		coreReviewData(t, response, 404)
	}
	for _, metadata := range []string{`{}`, `{"intake_id":"no-snapshot"}`} {
		missing := coreReviewApproval(t, db, models.ApprovalStatusDraft, metadata)
		response := coreReviewHTTP(t, h.GetApprovalSnapshot, "GET", "/snapshot", missing.ID, "", "lawyer", 7)
		coreReviewData(t, response, 404)
	}
	var persisted models.ApprovalSnapshot
	if err := db.First(&persisted, "id = ?", legacy.ID).Error; err != nil || persisted.SnapshotData != legacy.SnapshotData {
		t.Fatalf("history mutated: %v", err)
	}
	var saved models.ApprovalRequest
	if err := db.First(&saved, "id = ?", approval.ID).Error; err != nil || saved.ActiveSubmissionSnapshotID == nil || *saved.ActiveSubmissionSnapshotID != activeID {
		t.Fatalf("active pointer changed: %v", err)
	}
	response := coreReviewHTTP(t, h.GetApprovalSnapshot, "GET", "/snapshot?snapshot_id="+legacy.ID, approval.ID, "", "lawyer", 8)
	coreReviewData(t, response, 403)
}

type coreReviewIntegratedDecision struct {
	services.ApprovalConflictIntegrationService
	result *models.ApprovalRequest
	called bool
}

func (s *coreReviewIntegratedDecision) ProcessApprovalWithConflict(context.Context, string, string, string, *models.ApprovalDecisionRequest) (*models.ApprovalRequest, error) {
	s.called = true
	return s.result, nil
}

func TestCoreReviewIntegratedDecisionResponseProjection(t *testing.T) {
	db := coreReviewDB(t)
	approval := coreReviewApproval(t, db, models.ApprovalStatusSubmitted, `{"case_creation_config":{}}`)
	for _, role := range []string{"lawyer", "conflict_officer"} {
		t.Run(role, func(t *testing.T) {
			result := *approval
			result.Metadata = coreReviewSecrets
			stub := &coreReviewIntegratedDecision{result: &result}
			h := NewApprovalHandler(db)
			h.SetIntegratedDecisionService(stub)
			response := coreReviewHTTP(t, h.ProcessApprovalDecision, "POST", "/decision", approval.ID, `{"decision":"approve","decision_reason":"同意"}`, role, 9)
			coreReviewData(t, response, 200)
			if !stub.called {
				t.Fatal("did not exercise integrated decision success")
			}
			coreReviewNoSecrets(t, response.Body.String(), role)
			if stub.result.Metadata != coreReviewSecrets {
				t.Fatal("projection mutated service result")
			}
		})
	}
}

func TestCoreReviewCurrentReturnMustMatchSubmission(t *testing.T) {
	db := coreReviewDB(t)
	approval := coreReviewApproval(t, db, models.ApprovalStatusNeedsRevision, `{"intake_id":"current-intake"}`)
	snapshotID := uuid.NewString()
	snapshot := models.ApprovalSnapshot{ID: snapshotID, ApprovalRequestID: approval.ID, SnapshotType: models.IntakeSubmissionSnapshotType, SourceVersion: 1, SnapshotData: jsonStringValue(map[string]interface{}{
		"snapshot_id": snapshotID, "approval_request_id": approval.ID, "source_version": 1, "intake_id": "current-intake", "conflict_check_id": "current-check",
		"metadata": map[string]interface{}{"intake_id": "current-intake", "conflict_check_id": "current-check"},
	})}
	if err := db.Create(&snapshot).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(approval).Update("active_submission_snapshot_id", snapshotID).Error; err != nil {
		t.Fatal(err)
	}
	check := models.ConflictCheckRecord{CheckID: "current-check", SearchParameters: models.JSON{"metadata": map[string]interface{}{"intake_id": "current-intake"}}}
	if err := db.Create(&check).Error; err != nil {
		t.Fatal(err)
	}
	h := NewApprovalHandler(db)
	response := coreReviewHTTP(t, h.GetApproval, "GET", "/details", approval.ID, "", "lawyer", 7)
	coreReviewData(t, response, 409) // No verifiable return record.
	old := *approval
	old.ActiveSubmissionSnapshotID = nil
	now := time.Now().UTC().Truncate(time.Second)
	coreReviewReturn(t, db, &old, "过期轮次意见", `{}`, now.Add(-time.Hour))
	response = coreReviewHTTP(t, h.GetApproval, "GET", "/details", approval.ID, "", "lawyer", 7)
	coreReviewData(t, response, 409)
	if !strings.Contains(response.Body.String(), "APPROVAL_ROUND_STALE") {
		t.Fatal(response.Body.String())
	}
	approval.ActiveSubmissionSnapshotID = &snapshotID
	record := coreReviewReturn(t, db, approval, "本轮退回意见", `{}`, now)
	response = coreReviewHTTP(t, h.GetApproval, "GET", "/details", approval.ID, "", "lawyer", 7)
	data := coreReviewData(t, response, 200)
	if data["return_record_id"] != record.ID {
		t.Fatalf("wrong current return: %s", response.Body.String())
	}
}

func TestCoreReviewHistoryDoesNotRelaxAuthority(t *testing.T) {
	db := coreReviewDB(t)
	approval, legacy := coreReviewLegacy(t, db, 7)
	conflicting := models.ApprovalSnapshot{ID: uuid.NewString(), ApprovalRequestID: approval.ID, SnapshotType: "old_business_type", SourceVersion: 1, SnapshotData: `{"snapshot_id":"SNAP_other","metadata":{"intake_id":"intake-legacy","title":"不同旧事实"}}`}
	if err := db.Create(&conflicting).Error; err != nil {
		t.Fatal(err)
	}
	h := coreReviewSnapshotHandler(t, db)
	response := coreReviewHTTP(t, h.GetApprovalSnapshot, "GET", "/snapshot", approval.ID, "", "lawyer", 7)
	coreReviewData(t, response, 409)
	if !strings.Contains(response.Body.String(), "APPROVAL_SNAPSHOT_UNRESOLVED") {
		t.Fatal(response.Body.String())
	}
	response = coreReviewHTTP(t, h.GetApprovalSnapshot, "GET", "/snapshot?snapshot_id="+legacy.ID, approval.ID, "", "lawyer", 7)
	coreReviewData(t, response, 200) // An exact historical read makes no active-authority claim.
	for _, raw := range []string{`{broken`, `{"approval_request_id":"other-approval"}`, `{"approval":{"id":"other-approval"}}`} {
		row := models.ApprovalSnapshot{ID: uuid.NewString(), ApprovalRequestID: approval.ID, SnapshotType: "conflict_approval", SourceVersion: 1, SnapshotData: raw}
		if err := db.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
		response = coreReviewHTTP(t, h.GetApprovalSnapshot, "GET", "/snapshot?snapshot_id="+row.ID, approval.ID, "", "lawyer", 7)
		coreReviewData(t, response, 409)
	}
	missingActive := uuid.NewString()
	if err := db.Model(approval).Update("active_submission_snapshot_id", missingActive).Error; err != nil {
		t.Fatal(err)
	}
	response = coreReviewHTTP(t, h.GetApprovalSnapshot, "GET", "/snapshot", approval.ID, "", "lawyer", 7)
	coreReviewData(t, response, 409) // A corrupt active pointer is not an empty history.
}

// Snapshot readers now enforce real client/case object permissions. Supply
// those objects in history fixtures without changing any selection assertions.
func coreReviewSnapshotHandler(t *testing.T, db *gorm.DB) *DemoAggregateHandler {
	t.Helper()
	if err := db.AutoMigrate(&models.Client{}, &models.Case{}, &models.CaseEthicalWallWhitelist{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.Client{ID: 7, Name: "虚构历史客户", Type: "个人", CreatedBy: 7}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("CREATE TABLE case_intakes (id TEXT PRIMARY KEY, client_id INTEGER, created_by TEXT)").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO case_intakes VALUES ('intake-legacy', 7, '7')").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.ConflictCheckRecord{CheckID: "check-new", ClientID: "7", UserID: 7, SearchParameters: models.JSON{"intakeId": "intake-legacy"}}).Error; err != nil {
		t.Fatal(err)
	}
	h := NewDemoAggregateHandler(db)
	h.SetAuthorizationService(services.NewAuthorizationService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db), nil))
	return h
}
