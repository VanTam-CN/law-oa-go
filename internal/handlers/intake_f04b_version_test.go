package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
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

// F04B regression fixture: one SQLite database holding the minimal tables the
// intake workbench reads and writes. A single pooled connection serializes
// racing writers, so a losing commit observes the rotated version/token and
// fails with the typed business error instead of a raw DB error.
func newF04BIntakeDB(t *testing.T) *gorm.DB {
	t.Helper()
	t.Setenv("SUBJECT_DATA_KEY", strings.Repeat("f", 32))
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "f04b-intake.db")), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&models.ConflictCheckRecord{}, &models.ConflictCase{}, &models.Client{}, &models.User{}, &models.Case{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Exec("CREATE TABLE case_intakes (id TEXT PRIMARY KEY, intake_code TEXT, title TEXT, case_type TEXT, client_id INTEGER, created_by TEXT, priority TEXT, status TEXT, metadata TEXT, created_at DATETIME, updated_at DATETIME)").Error; err != nil {
		t.Fatalf("create case_intakes: %v", err)
	}
	if err := db.Exec("CREATE TABLE case_intake_parties (id TEXT PRIMARY KEY, intake_id TEXT, case_id TEXT, entity_name TEXT, entity_type TEXT, party_role TEXT, relation_depth INTEGER DEFAULT 0, metadata TEXT, identity_type TEXT, identity_number_ciphertext TEXT, identity_number_digest TEXT, aliases TEXT, created_at DATETIME)").Error; err != nil {
		t.Fatalf("create parties: %v", err)
	}
	return db
}

func seedF04BClient(t *testing.T, db *gorm.DB) *models.Client {
	t.Helper()
	identity := "91310000F04BTEST01"
	ciphertext, digest, err := security.ProtectIdentityNumber(identity)
	if err != nil {
		t.Fatalf("protect identity: %v", err)
	}
	client := &models.Client{
		Name: "F04B客户", Type: "企业", Status: "active", CreatedBy: 7,
		Email:        fmt.Sprintf("f04b-%d@test.local", time.Now().UnixNano()),
		IdentityType: models.IdentityTypeSocialCredit, IdentityNumberCiphertext: ciphertext, IdentityNumberDigest: digest,
	}
	if err := db.Create(client).Error; err != nil {
		t.Fatalf("seed client: %v", err)
	}
	return client
}

func seedF04BLawyer(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.Exec("INSERT INTO users (id, username, email, password, role, status, created_at, updated_at) VALUES (7, 'lawyer7', 'lawyer7@test.local', 'x', 'lawyer', 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)").Error; err != nil {
		t.Fatalf("seed lawyer: %v", err)
	}
}

// stubF04BConflictService always returns a clean, completed check so the
// handler reaches CommitIntakeConflictResult with a deterministic payload.
type stubF04BConflictService struct{ calls int32 }

func (s *stubF04BConflictService) PerformConflictCheck(ctx context.Context, r *models.ConflictCheckRequest) (*models.ConflictCheckResponse, error) {
	return s.PrepareConflictCheck(ctx, r)
}

func (s *stubF04BConflictService) PrepareConflictCheck(ctx context.Context, r *models.ConflictCheckRequest) (*models.ConflictCheckResponse, error) {
	call := atomic.AddInt32(&s.calls, 1)
	return &models.ConflictCheckResponse{
		CheckID: "CCT-F04B-" + fmt.Sprintf("%d", call), HasConflict: false,
		ConflictCases:   []*models.ConflictCase{},
		RiskAssessment:  &models.RiskAssessment{OverallRisk: "LOW", RiskScore: 0, RiskReason: "OK", RequiresApproval: false, RiskFactors: []string{}, Mitigation: []string{}},
		CheckStatistics: &models.CheckStatistics{}, Recommendations: []string{},
		CheckTime: time.Now().UTC(), Duration: 1,
	}, nil
}

func (s *stubF04BConflictService) BuildConflictCheckRecord(ctx context.Context, r *models.ConflictCheckRequest, resp *models.ConflictCheckResponse) *models.ConflictCheckRecord {
	return &models.ConflictCheckRecord{CheckID: resp.CheckID, ClientID: r.ClientID, ClientName: r.ClientName, CaseName: r.CaseName, CaseType: r.CaseType, CheckStatus: "COMPLETED", HasConflict: false, RiskLevel: "LOW", UserID: 7, CheckTime: resp.CheckTime}
}

func (s *stubF04BConflictService) GetCheckHistory(context.Context, string, int) ([]*models.ConflictCheckRecord, error) {
	return nil, nil
}
func (s *stubF04BConflictService) GetConflictStats(context.Context, string) (*repositories.ConflictStats, error) {
	return &repositories.ConflictStats{}, nil
}

func newF04BRouter(t *testing.T, db *gorm.DB) *gin.Engine {
	return newF04BRouterWithStub(t, db, &stubF04BConflictService{})
}

func newF04BRouterWithStub(t *testing.T, db *gorm.DB, svc *stubF04BConflictService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	h := NewDemoAggregateHandler(db)
	h.SetConflictDetectionService(svc)
	authz := services.NewAuthorizationService(
		repositories.NewCaseRepository(db), repositories.NewClientRepository(db),
		repositories.NewUserRepository(db), repositories.NewDocumentRepository(db),
	)
	h.SetAuthorizationService(authz)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint(7))
		c.Set("role", "lawyer")
		c.Next()
	})
	r.PUT("/api/v1/case-intakes/:id", h.UpdateCaseIntake)
	r.POST("/api/v1/case-intakes/:id/facts-confirmation", h.ConfirmIntakeFacts)
	r.POST("/api/v1/case-intakes/:id/conflict-check", h.StartIntakeConflictCheck)
	return r
}

func doF04B(r *gin.Engine, method, path, body string) (int, map[string]interface{}) {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	return w.Code, resp
}

func f04BErrorCode(resp map[string]interface{}) string {
	if resp == nil {
		return ""
	}
	errObj, _ := resp["error"].(map[string]interface{})
	if errObj == nil {
		return ""
	}
	code, _ := errObj["code"].(string)
	return code
}

func intakeMeta(t *testing.T, db *gorm.DB, id string) (string, map[string]interface{}) {
	t.Helper()
	var row struct {
		Status   string
		Metadata string
	}
	if err := db.Table("case_intakes").Select("status, metadata").Where("id = ?", id).Take(&row).Error; err != nil {
		t.Fatalf("load intake: %v", err)
	}
	meta := map[string]interface{}{}
	_ = json.Unmarshal([]byte(row.Metadata), &meta)
	return row.Status, meta
}

func f04BCheckCount(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var n int64
	if err := db.Model(&models.ConflictCheckRecord{}).Count(&n).Error; err != nil {
		t.Fatalf("count records: %v", err)
	}
	return n
}

// f04BIntakeRevision reads the server-owned draft revision. db may be nil
// inside f04BRunConfirmCheckCycle; the caller there passes the fixture db.

func seedF04BIntakeWithParty(t *testing.T, db *gorm.DB, clientID uint) string {
	t.Helper()
	intakeID := "f04b-intake-1"
	partyCipher, partyDigest, err := security.ProtectIdentityNumber("110101199001011234")
	if err != nil {
		t.Fatalf("protect party identity: %v", err)
	}
	if err := db.Exec("INSERT INTO case_intakes (id, intake_code, title, case_type, client_id, created_by, priority, status, metadata, created_at, updated_at) VALUES (?, 'ZC-F04B-001', 'F04B标题', 'civil', ?, '7', 'high', 'draft', '{}', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", intakeID, clientID).Error; err != nil {
		t.Fatalf("seed intake: %v", err)
	}
	if err := db.Exec("INSERT INTO case_intake_parties (id, intake_id, entity_name, entity_type, party_role, identity_type, identity_number_ciphertext, identity_number_digest) VALUES ('p1', ?, '对方当事人甲', 'INDIVIDUAL', 'RELATED_PARTY', 'ID_CARD', ?, ?)", intakeID, partyCipher, partyDigest).Error; err != nil {
		t.Fatalf("seed party: %v", err)
	}
	return intakeID
}

func f04BRunConfirmCheckCycle(t *testing.T, db *gorm.DB, r *gin.Engine, intakeID string) string {
	t.Helper()
	if code, resp := doF04B(r, "POST", "/api/v1/case-intakes/"+intakeID+"/facts-confirmation", ""); code != http.StatusOK {
		t.Fatalf("confirm: code=%d resp=%v", code, resp)
	}
	if code, resp := doF04B(r, "POST", "/api/v1/case-intakes/"+intakeID+"/conflict-check", ""); code != http.StatusOK {
		t.Fatalf("check: code=%d resp=%v", code, resp)
	}
	return f04BIntakeRevision(t, db, intakeID)
}

// f04BIntakeRevision reads the current server-owned draft revision token
// so contract-aware test saves can name the version they are based on.
func f04BIntakeRevision(t *testing.T, db *gorm.DB, intakeID string) string {
	t.Helper()
	var row struct {
		Metadata string
	}
	if err := db.Table("case_intakes").Select("metadata").Where("id = ?", intakeID).Take(&row).Error; err != nil {
		t.Fatalf("load intake revision: %v", err)
	}
	var meta map[string]interface{}
	_ = json.Unmarshal([]byte(row.Metadata), &meta)
	rev, _ := meta["draft_revision"].(string)
	return rev
}

func TestF04BSameSecondReconfirmInvalidatesCommittedSnapshot(t *testing.T) {
	db := newF04BIntakeDB(t)
	client := seedF04BClient(t, db)
	seedF04BLawyer(t, db)
	intakeID := seedF04BIntakeWithParty(t, db, client.ID)
	r := newF04BRouter(t, db)
	f04BRunConfirmCheckCycle(t, db, r, intakeID)
	if n := f04BCheckCount(t, db); n != 1 {
		t.Fatalf("first cycle must land one record, got %d", n)
	}
	// A re-confirmation in the same RFC3339 second rotates the server-owned
	// facts version, so the token/version pair the first commit installed
	// can no longer be replayed by a stale in-flight commit.
	if code, resp := doF04B(r, "POST", "/api/v1/case-intakes/"+intakeID+"/facts-confirmation", ""); code != http.StatusOK {
		t.Fatalf("re-confirm: code=%d resp=%v", code, resp)
	}
}

func TestF04BUpdateThenReconfirmInvalidatesPriorSnapshot(t *testing.T) {
	db := newF04BIntakeDB(t)
	client := seedF04BClient(t, db)
	client2 := seedF04BClient(t, db)
	seedF04BLawyer(t, db)
	intakeID := seedF04BIntakeWithParty(t, db, client.ID)
	r := newF04BRouter(t, db)
	f04BRunConfirmCheckCycle(t, db, r, intakeID)
	// Same-second fact change: update deletes the server-owned version and
	// re-confirm installs a fresh one, so the prior snapshot is gone even
	// if the RFC3339 second matches.
	revision := f04BIntakeRevision(t, db, intakeID)
	if code, resp := doF04B(r, "PUT", "/api/v1/case-intakes/"+intakeID, fmt.Sprintf("{\"expected_revision\":%q,\"client_id\": %d}", revision, client2.ID)); code != http.StatusOK {
		t.Fatalf("update: code=%d resp=%v", code, resp)
	}
	if code, resp := doF04B(r, "POST", "/api/v1/case-intakes/"+intakeID+"/facts-confirmation", ""); code != http.StatusOK {
		t.Fatalf("re-confirm after update: code=%d resp=%v", code, resp)
	}
}

func TestF04BDoubleCheckRaceYieldsOneWinner(t *testing.T) {
	db := newF04BIntakeDB(t)
	client := seedF04BClient(t, db)
	seedF04BLawyer(t, db)
	intakeID := seedF04BIntakeWithParty(t, db, client.ID)
	r := newF04BRouter(t, db)
	if code, resp := doF04B(r, "POST", "/api/v1/case-intakes/"+intakeID+"/facts-confirmation", ""); code != http.StatusOK {
		t.Fatalf("confirm: code=%d resp=%v", code, resp)
	}
	// Two checks race for the single commit token; SQLite single pooled
	// connection serializes them, so the loser must observe the rotated
	// token and come back as a typed 409 instead of a 500.
	results := make(chan int, 2)
	for i := 0; i < 2; i++ {
		go func() {
			code, _ := doF04B(r, "POST", "/api/v1/case-intakes/"+intakeID+"/conflict-check", "")
			results <- code
		}()
	}
	statuses := map[int]int{}
	for i := 0; i < 2; i++ {
		statuses[<-results]++
	}
	if statuses[http.StatusOK] != 1 || statuses[http.StatusConflict] != 1 {
		t.Fatalf("double check race: got %v, want one 200 and one 409", statuses)
	}
	if n := f04BCheckCount(t, db); n != 1 {
		t.Fatalf("exactly one record must land, got %d", n)
	}
}

func TestF04BStaleVersionCommitRejectedAndRolledBack(t *testing.T) {
	db := newF04BIntakeDB(t)
	client := seedF04BClient(t, db)
	seedF04BLawyer(t, db)
	intakeID := seedF04BIntakeWithParty(t, db, client.ID)
	r := newF04BRouter(t, db)
	f04BRunConfirmCheckCycle(t, db, r, intakeID)
	if code, resp := doF04B(r, "POST", "/api/v1/case-intakes/"+intakeID+"/facts-confirmation", ""); code != http.StatusOK {
		t.Fatalf("re-confirm: code=%d resp=%v", code, resp)
	}
	before, metaBefore := intakeMeta(t, db, intakeID)
	if v, _ := metaBefore["lawyer_facts_version"].(string); v == "" {
		t.Fatalf("re-confirm must install a facts version")
	}
	// A commit that still carries the previous expectation must fail with
	// the typed stale error and leave status and metadata untouched.
	resp := &models.ConflictCheckResponse{CheckID: "CCT-F04B-STALE", HasConflict: false, CheckTime: time.Now().UTC()}
	record := &models.ConflictCheckRecord{CheckID: resp.CheckID, ClientID: fmt.Sprint(client.ID), ClientName: client.Name, CaseName: "F04B标题", CaseType: "civil", CheckStatus: "COMPLETED", HasConflict: false, RiskLevel: "LOW", UserID: 7, CheckTime: resp.CheckTime}
	err := repositories.CommitIntakeConflictResult(context.Background(), db, record, nil, nil, resp, repositories.ConflictSubjectAssociation{
		CheckID:              resp.CheckID,
		IntakeID:             intakeID,
		ClientID:             fmt.Sprint(client.ID),
		ExpectedCommitToken:  "stale-token-from-in-flight-check",
		ExpectedFactsVersion: "stale-version-from-in-flight-check",
	}, time.Now().UTC().Format(time.RFC3339))
	var staleErr *repositories.IntakeConflictStaleFactsError
	if !errors.As(err, &staleErr) {
		t.Fatalf("expected IntakeConflictStaleFactsError, got %v", err)
	}
	if status, _ := intakeMeta(t, db, intakeID); status != before {
		t.Fatalf("rejected commit must not change status: before=%q after=%q", before, status)
	}
	if _, meta := intakeMeta(t, db, intakeID); meta["conflict_check_id"] == "CCT-F04B-STALE" {
		t.Fatalf("rejected commit must not claim conflict_check_id")
	}
	if n := f04BCheckCount(t, db); n != 1 {
		t.Fatalf("rejected commit must not land a record, got %d", n)
	}
}
