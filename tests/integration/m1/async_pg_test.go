//go:build np_m1_pg

package m1_test

import (
	"context"
	"encoding/json"
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

type pausedDetector struct {
	services.ConflictDetectionService
	started chan struct{}
	resume  chan struct{}
}

func (d pausedDetector) PerformConflictCheck(ctx context.Context, r *models.ConflictCheckRequest) (*models.ConflictCheckResponse, error) {
	close(d.started)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-d.resume:
		return &models.ConflictCheckResponse{CheckID: r.CheckID, CheckTime: time.Now()}, nil
	}
}

func TestAsyncPGCompletionAfterNativeEdit(t *testing.T) {
	db := isolatedDB(t)
	f := seedConflictPGFixture(t, db, true, "")
	repo := repositories.NewConflictRepository(db, nil)
	engine := pausedDetector{started: make(chan struct{}), resume: make(chan struct{})}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	service := services.NewAsyncConflictCheckService(repo, engine)
	req := &models.ConflictCheckRequest{IntakeID: f.intakeID, IntakeRevision: f.revision, SubjectCaseID: f.caseID, ClientID: f.clientID, ClientType: "COMPANY", ClientName: "Synthetic client", CaseName: "Synthetic intake", CaseType: "civil", OtherParties: []string{"Synthetic opponent"}, UserID: fmt.Sprint(f.userID)}
	task, err := service.CreateTask(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	resumed := false
	defer func() {
		if !resumed {
			close(engine.resume)
		}
	}()
	select {
	case <-engine.started:
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
	var snapshot string
	read := func() string {
		t.Helper()
		var current string
		if err := db.Raw("SELECT to_jsonb(i)::text FROM case_intakes i WHERE id=?", f.intakeID).Scan(&current).Error; err != nil {
			t.Fatal(err)
		}
		return current
	}
	snapshot = read()
	close(engine.resume)
	resumed = true
	for {
		record, err := repo.GetCheckRecord(ctx, task.CheckID)
		if err != nil {
			t.Fatal(err)
		}
		if record.CheckStatus == services.ConflictTaskStatusFailed {
			if !strings.Contains(fmt.Sprint(record.CheckResult["error"]), repositories.ErrIntakeRevisionConflict.Error()) {
				t.Fatal("unexpected failure", record.CheckResult["error"])
			}
			break
		}
		if record.CheckStatus == services.ConflictTaskStatusCompleted {
			t.Fatal("late result marked completed")
		}
		select {
		case <-ctx.Done():
			t.Fatal("async completion timed out")
		case <-time.After(10 * time.Millisecond):
		}
	}
	if snapshot != read() || conflictPGCaseCheckID(t, db, f.caseID) != "" {
		t.Fatal("late completion altered edited intake or case pointer")
	}
}
