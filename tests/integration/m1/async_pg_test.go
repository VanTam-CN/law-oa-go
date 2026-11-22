//go:build np_m1_pg

package m1_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/handlers"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

// Pause inside the normal risk assessment. The detector, queue, evidence persistence
// and version-guarded association all remain wired to their production implementations.
type gatedRiskAssessor struct {
	services.RiskAssessor
	started chan struct{}
	resume  chan struct{}
}

func (a gatedRiskAssessor) GenerateRecommendations(ctx context.Context, risk string, cases []*models.ConflictCase) ([]string, error) {
	close(a.started)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-a.resume:
		return a.RiskAssessor.GenerateRecommendations(ctx, risk, cases)
	}
}

func TestAsyncPGCompletionAfterNativeEdit(t *testing.T)   { checkFormalAsyncPG(t, true, false) }
func TestAsyncPGFormalDetectorCompletesOnce(t *testing.T) { checkFormalAsyncPG(t, false, false) }

func TestAsyncPGTerminalWriteFailureRollsBack(t *testing.T) { checkFormalAsyncPG(t, false, true) }

func checkFormalAsyncPG(t *testing.T, editDuringCheck, rejectTerminal bool) {
	t.Helper()
	db := isolatedDB(t)
	f := seedConflictPGFixture(t, db, true, "")
	repo := repositories.NewConflictRepository(db, nil)
	gate := gatedRiskAssessor{RiskAssessor: services.NewRiskAssessor(nil, nil), started: make(chan struct{}), resume: make(chan struct{})}
	engine := services.NewConflictDetectionService(repo, gate, repositories.NewUserRepository(db), repositories.NewClientRepository(db), repositories.NewCaseRepository(db))
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	service := services.NewAsyncConflictCheckService(repo, engine)
	req := &models.ConflictCheckRequest{IntakeID: f.intakeID, IntakeRevision: f.revision, SubjectCaseID: f.caseID, ClientID: f.clientID, ClientType: "COMPANY", ClientName: "Synthetic client", CaseName: "Synthetic intake", CaseType: "civil", OtherParties: []string{"Synthetic opponent"}, UserID: fmt.Sprint(f.userID)}
	const syntheticIdentifier = "M1-ASYNC-SYNTHETIC-ID"
	t.Setenv("SUBJECT_DATA_KEY", strings.Repeat("p", 32))
	req.ClientIdentifiers = map[string]string{"registration": syntheticIdentifier}
	task, err := service.CreateTask(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	resumed := false
	defer func() {
		if !resumed {
			close(gate.resume)
		}
	}()
	select {
	case <-gate.started:
	case <-ctx.Done():
		t.Fatal("async detector did not start")
	}
	var captured string
	if err := db.Raw("SELECT search_parameters->>'intakeRevision' FROM conflict_check_records WHERE check_id=?", task.CheckID).Scan(&captured).Error; err != nil || captured != f.revision {
		t.Fatal("input version missing from queued record", err)
	}
	rev, status, _, _ := conflictPGIntakeState(t, db, f.intakeID)
	if rev != f.revision || status != "lawyer_facts_confirmed" {
		t.Fatal("start changed confirmed input")
	}
	if editDuringCheck {
		h := handlers.NewDemoAggregateHandler(db)
		h.SetAuthorizationService(services.NewAuthorizationService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db), repositories.NewDocumentRepository(db)))
		r := gin.New()
		r.PUT("/case-intakes/:id", func(c *gin.Context) { c.Set("role", "lawyer"); c.Set("user_id", f.userID); h.UpdateCaseIntake(c) })
		body, _ := json.Marshal(map[string]string{"expected_revision": rev, "title": "Edited while check was running"})
		put := httptest.NewRequest("PUT", "/case-intakes/"+f.intakeID, strings.NewReader(string(body)))
		put.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, put)
		if w.Code != 200 {
			t.Fatalf("native edit returned %d", w.Code)
		}
	}
	read := func() string {
		t.Helper()
		var current string
		if err := db.Raw("SELECT to_jsonb(i)::text FROM case_intakes i WHERE id=?", f.intakeID).Scan(&current).Error; err != nil {
			t.Fatal(err)
		}
		return current
	}
	snapshot := read()
	if rejectTerminal {
		var quoted string
		if err := db.Raw("SELECT quote_literal(?)", task.CheckID).Scan(&quoted).Error; err != nil {
			t.Fatal(err)
		}
		const constraint = "m1_fixture_terminal_failure"
		if err := db.Exec("ALTER TABLE conflict_check_records ADD CONSTRAINT " + constraint + " CHECK (check_id <> " + quoted + " OR check_status <> 'COMPLETED')").Error; err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := db.Exec("ALTER TABLE conflict_check_records DROP CONSTRAINT " + constraint).Error; err != nil {
				t.Error(err)
			}
		})
	}
	close(gate.resume)
	resumed = true
	for {
		record, err := repo.GetCheckRecord(ctx, task.CheckID)
		if err != nil {
			t.Fatal(err)
		}
		if record.CheckStatus == services.ConflictTaskStatusFailed {
			if rejectTerminal && record.CheckResult["error"] == nil {
				t.Fatal("terminal failure missing reason")
			}
			if !rejectTerminal && (!editDuringCheck || !strings.Contains(fmt.Sprint(record.CheckResult["error"]), repositories.ErrIntakeRevisionConflict.Error())) {
				t.Fatal("unexpected failure", record.CheckResult["error"])
			}
			break
		}
		if record.CheckStatus == services.ConflictTaskStatusCompleted {
			if editDuringCheck || rejectTerminal {
				t.Fatal("late result marked completed")
			}
			break
		}
		if rejectTerminal && conflictPGCaseCheckID(t, db, f.caseID) != "" {
			t.Fatal("business link committed although terminal write rejected")
		}
		select {
		case <-ctx.Done():
			t.Fatal("async completion timed out")
		case <-time.After(10 * time.Millisecond):
		}
	}
	var count int64
	if err := db.Table("conflict_check_records").Where("check_id=?", task.CheckID).Count(&count).Error; err != nil || count != 1 {
		t.Fatal("incorrect audit record count", count, err)
	}
	if editDuringCheck || rejectTerminal {
		if snapshot != read() || conflictPGCaseCheckID(t, db, f.caseID) != "" {
			t.Fatal("late completion altered edited intake or case pointer")
		}
		return
	}
	savedRev, state, linked, precision := conflictPGIntakeState(t, db, f.intakeID)
	if savedRev == "" || savedRev == rev || state != "conflict_ready" || linked != task.CheckID || precision != conflictPGBigPrecision || conflictPGCaseCheckID(t, db, f.caseID) != task.CheckID {
		t.Fatal("fresh formal check did not atomically link both objects")
	}
	var audit struct{ Input, Result string }
	if err := db.Raw("SELECT search_parameters::text AS input,check_result::text AS result FROM conflict_check_records WHERE check_id=?", task.CheckID).Scan(&audit).Error; err != nil {
		t.Fatal(err)
	}
	for _, document := range []string{audit.Input, audit.Result} {
		if strings.Contains(document, syntheticIdentifier) || !strings.Contains(document, "hmac-sha256:") {
			t.Fatal("identifier missing or unprotected in persisted audit")
		}
	}
	var subjects int64
	if err := db.Model(&models.ConflictSubjectVersion{}).Where("source_id=?", task.CheckID).Count(&subjects).Error; err != nil || subjects == 0 {
		t.Fatal("missing formal detector evidence", err)
	}
	// Re-delivery of the original completion cannot rotate the version a second time.
	replayed := repositories.ConflictSubjectAssociation{CheckID: task.CheckID, ClientID: f.clientID, SubjectCaseID: f.caseID, IntakeID: f.intakeID, IntakeRevision: rev}
	if err := repositories.LinkConflictCheckToCase(ctx, db, replayed); !errors.Is(err, repositories.ErrIntakeRevisionConflict) {
		t.Fatal("unexpected completion replay outcome", err)
	}
	endRev, _, _, _ := conflictPGIntakeState(t, db, f.intakeID)
	if endRev != savedRev {
		t.Fatal("completion replay advanced draft revision")
	}
}
