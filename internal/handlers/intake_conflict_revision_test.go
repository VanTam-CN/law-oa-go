package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/security"
	"law-oa-go/internal/services"
)

type intakeRevisionStubService struct {
	beforeReturn func()
	mu           sync.Mutex
	inserts      int
	db           *gorm.DB
}

func (s *intakeRevisionStubService) PerformConflictCheck(ctx context.Context, request *models.ConflictCheckRequest) (*models.ConflictCheckResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("请求为空")
	}
	s.mu.Lock()
	s.inserts++
	checkID := fmt.Sprintf("CCT-I15-%d", s.inserts)
	parameters, err := json.Marshal(map[string]string{
		"clientId":       request.ClientID,
		"intakeId":       request.IntakeID,
		"intakeRevision": request.IntakeRevision,
	})
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}
	if err := s.db.WithContext(ctx).Exec(
		"INSERT INTO conflict_check_records (check_id, client_id, search_parameters, check_status) VALUES (?, ?, ?, ?)",
		checkID, request.ClientID, string(parameters), "COMPLETED",
	).Error; err != nil {
		s.mu.Unlock()
		return nil, err
	}
	s.mu.Unlock()
	if s.beforeReturn != nil {
		s.beforeReturn()
	}
	return &models.ConflictCheckResponse{
		CheckID:     checkID,
		HasConflict: false,
		CheckTime:   time.Now(),
		Decision: &models.ConflictDecisionSummary{
			Status:         "CLEAR",
			CoverageStatus: "COMPLETE",
		},
	}, nil
}

func (s *intakeRevisionStubService) GetCheckHistory(context.Context, string, int) ([]*models.ConflictCheckRecord, error) {
	return nil, nil
}

func (s *intakeRevisionStubService) GetConflictStats(context.Context, string) (*repositories.ConflictStats, error) {
	return &repositories.ConflictStats{}, nil
}

func newIntakeRevisionWorld(t *testing.T) (*DemoAggregateHandler, *gin.Engine, *gorm.DB, *intakeRevisionStubService) {
	t.Helper()
	t.Setenv("SUBJECT_DATA_KEY", strings.Repeat("i", 32))
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&models.Client{}); err != nil {
		t.Fatalf("migrate client: %v", err)
	}
	ddl := []string{
		"CREATE TABLE conflict_check_records (check_id TEXT PRIMARY KEY, client_id TEXT, search_parameters TEXT, check_status TEXT)",
		"CREATE TABLE case_intakes (id TEXT PRIMARY KEY, intake_code TEXT, title TEXT, case_type TEXT, description TEXT, priority TEXT, client_id INTEGER, created_by TEXT, status TEXT, metadata TEXT, created_at DATETIME, updated_at DATETIME)",
		"CREATE TABLE case_intake_parties (id INTEGER PRIMARY KEY AUTOINCREMENT, intake_id TEXT, entity_name TEXT, entity_type TEXT, party_role TEXT, relation_depth INTEGER, identity_type TEXT, identity_number_ciphertext TEXT, identity_number_digest TEXT, aliases TEXT, metadata TEXT, created_at DATETIME)",
		"CREATE TABLE case_intake_source_snapshots (intake_id TEXT, saved_revision TEXT)",
		"CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, role TEXT, status TEXT, department TEXT, deleted_at DATETIME)",
		"CREATE TABLE cases (id INTEGER PRIMARY KEY AUTOINCREMENT, client_id INTEGER, lawyer_id TEXT, created_by TEXT, deleted_at DATETIME, ethical_wall_enabled BOOLEAN DEFAULT FALSE, conflict_check_id TEXT, conflict_coverage_status TEXT)",
		"CREATE TABLE case_ethical_wall_whitelist (id INTEGER PRIMARY KEY AUTOINCREMENT, case_id INTEGER, user_id TEXT)",
	}
	for _, stmt := range ddl {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("create table: %v", err)
		}
	}
	ciphertext, digest, err := security.ProtectIdentityNumber(security.NormalizeIdentityNumber(string(models.IdentityTypeSocialCredit), "91310000MA1FL0000X"))
	if err != nil {
		t.Fatalf("protect identity: %v", err)
	}
	if err := db.Create(&models.Client{
		Name:                     "星河智联科技有限公司",
		Type:                     "企业",
		Email:                    "client-33@example.com",
		IdentityType:             models.IdentityTypeSocialCredit,
		IdentityNumberCiphertext: ciphertext,
		IdentityNumberDigest:     digest,
		CreatedBy:                7,
	}).Error; err != nil {
		t.Fatalf("seed client: %v", err)
	}
	if err := db.Exec("UPDATE clients SET id = 33 WHERE id = (SELECT MAX(id) FROM clients)").Error; err != nil {
		t.Fatalf("fix client id: %v", err)
	}
	seeds := []string{
		"INSERT INTO users (id, name, role, status, department) VALUES (7, '律师甲', 'lawyer', 'active', '争议解决')",
		"INSERT INTO case_intakes (id, intake_code, title, case_type, client_id, created_by, status, metadata) VALUES ('intake-revision', 'INT-I15', '版本守护草稿', 'civil', 33, '7', 'draft', '{}')",
		"INSERT INTO case_intake_parties (intake_id, entity_name, entity_type, party_role, identity_type, identity_number_ciphertext, identity_number_digest) VALUES ('intake-revision', '星河智联科技有限公司', 'LEGAL_PERSON', 'client', 'SOCIAL_CREDIT_CODE', ?, ?)",
		"INSERT INTO case_intake_parties (intake_id, entity_name, entity_type, party_role, identity_type, identity_number_ciphertext, identity_number_digest) VALUES ('intake-revision', '对手公司', 'LEGAL_PERSON', 'opposing_party', 'SOCIAL_CREDIT_CODE', ?, ?)",
	}
	for index, stmt := range seeds {
		var err error
		if index < 2 {
			err = db.Exec(stmt).Error
		} else {
			err = db.Exec(stmt, ciphertext, digest).Error
		}
		if err != nil {
			t.Fatalf("seed %d: %v", index, err)
		}
	}
	stub := &intakeRevisionStubService{db: db}
	handler := NewDemoAggregateHandler(db)
	handler.SetConflictDetectionService(stub)
	handler.SetAuthorizationService(services.NewAuthorizationService(
		repositories.NewCaseRepository(db),
		repositories.NewClientRepository(db),
		repositories.NewUserRepository(db),
		repositories.NewDocumentRepository(db),
	))
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", c.GetHeader("X-Test-Role"))
		c.Set("user_id", c.GetHeader("X-Test-User-Id"))
		c.Next()
	})
	router.POST("/api/v1/case-intakes/:id/facts-confirmation", handler.ConfirmIntakeFacts)
	router.PUT("/api/v1/case-intakes/:id", handler.UpdateCaseIntake)
	router.POST("/api/v1/case-intakes/:id/conflict-check", handler.StartIntakeConflictCheck)
	return handler, router, db, stub
}

func intakeRevisionRequest(t *testing.T, router *gin.Engine, method, path, body, revisionHeader string) (int, map[string]interface{}) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Role", "lawyer")
	req.Header.Set("X-Test-User-Id", "7")
	if revisionHeader != "" {
		req.Header.Set("X-Intake-Revision", revisionHeader)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	var payload map[string]interface{}
	_ = json.Unmarshal(recorder.Body.Bytes(), &payload)
	return recorder.Code, payload
}

func readIntakeRevisionState(t *testing.T, db *gorm.DB) (string, string, string) {
	t.Helper()
	row := map[string]interface{}{}
	if err := db.Table("case_intakes").Where("id = ?", "intake-revision").Take(&row).Error; err != nil {
		t.Fatalf("reload intake: %v", err)
	}
	rawMetadata, _ := row["metadata"].(string)
	if rawMetadata == "" {
		rawMetadata = fmt.Sprint(row["metadata"])
	}
	var metadata map[string]interface{}
	_ = json.Unmarshal([]byte(rawMetadata), &metadata)
	revision, _ := metadata["draft_revision"].(string)
	if revision == "" {
		revision = "legacy-intake-revision"
	}
	checkID, _ := metadata["conflict_check_id"].(string)
	return fmt.Sprint(row["status"]), revision, checkID
}

func countIntakeRevisionBoundCases(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var count int64
	if err := db.Table("cases").Where("conflict_check_id IS NOT NULL AND conflict_check_id <> ''").Count(&count).Error; err != nil {
		t.Fatalf("count bound cases: %v", err)
	}
	return count
}

func expectIntakeRevisionConflict(t *testing.T, code int, payload map[string]interface{}) {
	t.Helper()
	if code != http.StatusConflict {
		t.Fatalf("status=%d want 409 body=%v", code, payload)
	}
	errorObject, _ := payload["error"].(map[string]interface{})
	if errorObject == nil || errorObject["code"] != "INTAKE_DRAFT_REVISION_CONFLICT" {
		t.Fatalf("want INTAKE_DRAFT_REVISION_CONFLICT, got %v", payload)
	}
}

func intakeRevisionData(t *testing.T, payload map[string]interface{}) map[string]interface{} {
	t.Helper()
	data, _ := payload["data"].(map[string]interface{})
	if data == nil {
		t.Fatalf("missing data object: %v", payload)
	}
	return data
}

func TestStartIntakeConflictCheckFreshRevisionLinksOnce(t *testing.T) {
	_, router, db, stub := newIntakeRevisionWorld(t)
	code, payload := intakeRevisionRequest(t, router, http.MethodPost, "/api/v1/case-intakes/intake-revision/facts-confirmation", "{}", "legacy-intake-revision")
	if code != http.StatusOK {
		t.Fatalf("confirm status=%d body=%v", code, payload)
	}
	confirmed := intakeRevisionData(t, payload)
	stale := fmt.Sprint(confirmed["revision"])
	code, payload = intakeRevisionRequest(t, router, http.MethodPost, "/api/v1/case-intakes/intake-revision/conflict-check", "{}", stale)
	if code != http.StatusOK {
		t.Fatalf("start status=%d body=%v", code, payload)
	}
	data := intakeRevisionData(t, payload)
	status, storedRevision, boundCheckID := readIntakeRevisionState(t, db)
	completionRevision := fmt.Sprint(data["revision"])
	if completionRevision == "" || completionRevision != storedRevision {
		t.Fatalf("completion revision mismatch: response=%q stored=%q body=%v", completionRevision, storedRevision, payload)
	}
	if status != "conflict_ready" || boundCheckID != fmt.Sprint(data["taskId"]) {
		t.Fatalf("intake not linked: status=%q checkID=%q body=%v", status, boundCheckID, payload)
	}
	if stub.insertCount() != 1 {
		t.Fatalf("stub inserts=%d want 1", stub.insertCount())
	}
	if boundCases := countIntakeRevisionBoundCases(t, db); boundCases != 0 {
		t.Fatalf("bound cases=%d want 0", boundCases)
	}
}

func TestStartIntakeConflictCheckRejectsStaleRevisionAfterNativeEdit(t *testing.T) {
	_, router, db, stub := newIntakeRevisionWorld(t)
	code, payload := intakeRevisionRequest(t, router, http.MethodPost, "/api/v1/case-intakes/intake-revision/facts-confirmation", "{}", "legacy-intake-revision")
	if code != http.StatusOK {
		t.Fatalf("confirm status=%d body=%v", code, payload)
	}
	stale := fmt.Sprint(intakeRevisionData(t, payload)["revision"])
	code, payload = intakeRevisionRequest(t, router, http.MethodPut, "/api/v1/case-intakes/intake-revision", "{\"title\":\"事实已修改\",\"expected_revision\":\""+stale+"\"}", "")
	if code != http.StatusOK {
		t.Fatalf("edit status=%d body=%v", code, payload)
	}
	fresh := fmt.Sprint(intakeRevisionData(t, payload)["revision"])
	if fresh == "" || fresh == stale {
		t.Fatalf("edit must rotate revision: stale=%q fresh=%q body=%v", stale, fresh, payload)
	}
	code, payload = intakeRevisionRequest(t, router, http.MethodPost, "/api/v1/case-intakes/intake-revision/conflict-check", "{}", stale)
	expectIntakeRevisionConflict(t, code, payload)
	status, storedRevision, boundCheckID := readIntakeRevisionState(t, db)
	if status != "draft" || storedRevision != fresh || boundCheckID != "" {
		t.Fatalf("stale start mutated intake: status=%q revision=%q checkID=%q", status, storedRevision, boundCheckID)
	}
	if stub.insertCount() != 0 {
		t.Fatalf("stale start must not reach stub, inserts=%d", stub.insertCount())
	}
	if boundCases := countIntakeRevisionBoundCases(t, db); boundCases != 0 {
		t.Fatalf("stale start bound cases=%d want 0", boundCases)
	}
	code, payload = intakeRevisionRequest(t, router, http.MethodPost, "/api/v1/case-intakes/intake-revision/facts-confirmation", "{}", fresh)
	if code != http.StatusOK {
		t.Fatalf("reconfirm status=%d body=%v", code, payload)
	}
	reconfirmed := fmt.Sprint(intakeRevisionData(t, payload)["revision"])
	if reconfirmed == fresh {
		t.Fatalf("reconfirm must rotate revision: %q", reconfirmed)
	}
	code, payload = intakeRevisionRequest(t, router, http.MethodPost, "/api/v1/case-intakes/intake-revision/conflict-check", "{}", reconfirmed)
	if code != http.StatusOK {
		t.Fatalf("fresh retry status=%d body=%v", code, payload)
	}
	status, storedRevision, boundCheckID = readIntakeRevisionState(t, db)
	if status != "conflict_ready" || boundCheckID != fmt.Sprint(intakeRevisionData(t, payload)["taskId"]) {
		t.Fatalf("fresh retry did not link: status=%q checkID=%q body=%v", status, boundCheckID, payload)
	}
	if storedRevision != fmt.Sprint(intakeRevisionData(t, payload)["revision"]) {
		t.Fatalf("fresh retry revision mismatch: stored=%q body=%v", storedRevision, payload)
	}
	if stub.insertCount() != 1 {
		t.Fatalf("retry inserts=%d want 1", stub.insertCount())
	}
}

func TestStartIntakeConflictCheckRejectsSecondSubmissionWithOldRevision(t *testing.T) {
	_, router, db, stub := newIntakeRevisionWorld(t)
	code, payload := intakeRevisionRequest(t, router, http.MethodPost, "/api/v1/case-intakes/intake-revision/facts-confirmation", "{}", "legacy-intake-revision")
	if code != http.StatusOK {
		t.Fatalf("confirm status=%d body=%v", code, payload)
	}
	used := fmt.Sprint(intakeRevisionData(t, payload)["revision"])
	code, payload = intakeRevisionRequest(t, router, http.MethodPost, "/api/v1/case-intakes/intake-revision/conflict-check", "{}", used)
	if code != http.StatusOK {
		t.Fatalf("first start status=%d body=%v", code, payload)
	}
	first := intakeRevisionData(t, payload)
	code, payload = intakeRevisionRequest(t, router, http.MethodPost, "/api/v1/case-intakes/intake-revision/conflict-check", "{}", used)
	expectIntakeRevisionConflict(t, code, payload)
	status, storedRevision, boundCheckID := readIntakeRevisionState(t, db)
	if status != "conflict_ready" || boundCheckID != fmt.Sprint(first["taskId"]) {
		t.Fatalf("replay changed intake: status=%q checkID=%q body=%v", status, boundCheckID, payload)
	}
	if storedRevision != fmt.Sprint(first["revision"]) {
		t.Fatalf("replay changed revision: stored=%q first=%q", storedRevision, first["revision"])
	}
	if boundCases := countIntakeRevisionBoundCases(t, db); boundCases != 0 {
		t.Fatalf("replay bound cases=%d want 0", boundCases)
	}
	if stub.insertCount() != 1 {
		t.Fatalf("replay inserts=%d want 1", stub.insertCount())
	}
}

func (s *intakeRevisionStubService) insertCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.inserts
}

func TestStartIntakeConflictCheckLateResultAfterNativeEdit(t *testing.T) {
	_, r, db, stub := newIntakeRevisionWorld(t)
	path := "/api/v1/case-intakes/intake-revision"
	code, body := intakeRevisionRequest(t, r, http.MethodPost, path+"/facts-confirmation", "{}", "legacy-intake-revision")
	if code != 200 {
		t.Fatalf("confirm: %d %v", code, body)
	}
	original := fmt.Sprint(intakeRevisionData(t, body)["revision"])
	var after, meta string
	stub.beforeReturn = func() {
		code, body := intakeRevisionRequest(t, r, http.MethodPut, path, `{"title":"later facts","expected_revision":"`+original+`"}`, "")
		if code != 200 {
			t.Fatalf("native edit during detection: %d %v", code, body)
		}
		after = fmt.Sprint(intakeRevisionData(t, body)["revision"])
		if err := db.Table("case_intakes").Select("metadata").Where("id = ?", "intake-revision").Scan(&meta).Error; err != nil {
			t.Fatal(err)
		}
	}
	code, body = intakeRevisionRequest(t, r, http.MethodPost, path+"/conflict-check", "{}", original)
	expectIntakeRevisionConflict(t, code, body)
	status, rev, check := readIntakeRevisionState(t, db)
	var final string
	if err := db.Table("case_intakes").Select("metadata").Where("id = ?", "intake-revision").Scan(&final).Error; err != nil {
		t.Fatal(err)
	}
	if after == "" || status != "draft" || rev != after || check != "" || meta != final || countIntakeRevisionBoundCases(t, db) != 0 || stub.insertCount() != 1 {
		t.Fatal("late completion changed edited facts or case binding")
	}
}
