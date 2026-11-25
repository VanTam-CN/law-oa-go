//go:build np_m1_pg

package m1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"law-oa-go/internal/handlers"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/security"
	"law-oa-go/internal/services"
)

// Both completions are paused inside the production detector before a native edit
// and a source-linked strict save. All four protected relations are compared in PG.
func TestM3ConflictCompletionAfterSourceSave(t *testing.T) {
	for _, mode := range []string{"sync", "async"} {
		t.Run(mode, func(t *testing.T) {
			db := isolatedDB(t)
			f := seedConflictPGFixture(t, db, true, "")
			t.Setenv("NP_ASSISTANT_ACTORS", fmt.Sprint(f.userID))
			router, token := nativePartyRouter(t, db, f)
			registerIntakeTestSession(t, db, f.userID, token)
			exec := func(sql string, args ...any) {
				t.Helper()
				if err := db.Exec(sql, args...).Error; err != nil {
					t.Fatal(err)
				}
			}
			cipher, digest, err := security.ProtectIdentityNumber("SYNTHETIC-M3-" + f.intakeID)
			if err != nil {
				t.Fatal(err)
			}
			exec("UPDATE clients SET identity_type='SOCIAL_CREDIT_CODE',identity_number_ciphertext=?,identity_number_digest=? WHERE id=?", cipher, digest, f.clientID)
			exec("INSERT INTO case_intake_parties(intake_id,party_role,entity_type,entity_name,identity_type,identity_number_ciphertext,identity_number_digest) VALUES (?,'opposing_party','LEGAL_PERSON','Synthetic M3 opponent','SOCIAL_CREDIT_CODE',?,?)", f.intakeID, cipher, digest)
			exec("UPDATE case_intakes SET status='draft',metadata=jsonb_set(metadata,'{subject_case_id}',to_jsonb(?::text)) WHERE id=?", f.caseID, f.intakeID)
			path := "/api/v1/case-intakes/" + f.intakeID
			request := func(method, path string, input any) map[string]any {
				t.Helper()
				body, err := json.Marshal(input)
				if err != nil {
					t.Fatal(err)
				}
				req := httptest.NewRequest(method, path, bytes.NewReader(body))
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Idempotency-Key", uuid.NewString())
				req.Header.Set("X-Intake-Revision", f.revision)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				var out struct {
					Data  map[string]any
					Error struct{ Code string }
				}
				if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil || w.Code != 200 {
					t.Fatalf("%s %s status=%d code=%s err=%v", method, path, w.Code, out.Error.Code, err)
				}
				return out.Data
			}
			confirmed := request("POST", path+"/facts-confirmation", map[string]any{})
			f.revision = confirmed["revision"].(string)
			repo := repositories.NewConflictRepository(db, nil)
			gate := gatedRiskAssessor{RiskAssessor: services.NewRiskAssessor(nil, nil), started: make(chan struct{}), resume: make(chan struct{})}
			engine := services.NewConflictDetectionService(repo, gate, repositories.NewUserRepository(db), repositories.NewClientRepository(db), repositories.NewCaseRepository(db))
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			async := services.NewAsyncConflictCheckService(repo, engine)
			var taskID string
			result := make(chan *httptest.ResponseRecorder, 1)
			if mode == "sync" {
				h := handlers.NewDemoAggregateHandler(db)
				h.SetConflictDetectionService(engine)
				h.SetAuthorizationService(services.NewAuthorizationService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db), repositories.NewDocumentRepository(db)))
				g := gin.New()
				g.POST("/intakes/:id/check", middleware.AuthMiddleware(), h.StartIntakeConflictCheck)
				req := httptest.NewRequest("POST", "/intakes/"+f.intakeID+"/check", nil).WithContext(ctx)
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("X-Intake-Revision", f.revision)
				go func() { w := httptest.NewRecorder(); g.ServeHTTP(w, req); result <- w }()
			} else {
				task, err := async.CreateTask(ctx, &models.ConflictCheckRequest{IntakeID: f.intakeID, IntakeRevision: f.revision, SubjectCaseID: f.caseID, ClientID: f.clientID, ClientType: "COMPANY", ClientName: "Synthetic client", CaseName: "Synthetic M3 intake", CaseType: "civil", OtherParties: []string{"Synthetic M3 opponent"}, UserID: fmt.Sprint(f.userID)})
				if err != nil {
					t.Fatal(err)
				}
				taskID = task.CheckID
			}
			released := false
			defer func() {
				if !released {
					close(gate.resume)
				}
			}()
			select {
			case <-gate.started:
			case w := <-result:
				t.Fatalf("sync check exited before edit: %d %s", w.Code, w.Body.String())
			case <-ctx.Done():
				t.Fatal("formal detector did not start")
			}
			rev, state, _, _ := conflictPGIntakeState(t, db, f.intakeID)
			if rev != f.revision || state != "lawyer_facts_confirmed" {
				t.Fatal("start changed input")
			}
			edit := request("PUT", path, map[string]any{"expected_revision": rev, "title": "New synthetic facts"})
			nativeRev := edit["revision"].(string)
			if nativeRev == rev || edit["status"] != "draft" {
				t.Fatal("native edit did not invalidate confirmation")
			}
			work := request("POST", "/api/v1/case-intake-assist/contexts", map[string]any{"intake_id": f.intakeID, "expected_revision": nativeRev})
			turn := request("POST", "/api/v1/case-intake-assist/contexts/"+work["id"].(string)+"/turns", map[string]any{"expected_context_version": 1, "request_id": uuid.NewString(), "text": "Synthetic M3 updated facts."})
			candidate := turn["candidate"].(map[string]any)
			selection := map[string]any{"context_id": work["id"], "expected_context_version": turn["context_version"], "fields": map[string]any{"description": map[string]any{"candidate_id": candidate["id"], "requested_mode": "candidate_sources", "reviewed": true}}}
			saved := request("PUT", path, map[string]any{"expected_revision": nativeRev, "changed_fields": []string{"description"}, "description": candidate["description"], "source_selection": selection})
			if saved["revision"] == nativeRev || saved["source_revision"] != saved["revision"] || saved["state"] != "committed" {
				t.Fatal("missing source-linked receipt")
			}
			snapshot := func() string {
				t.Helper()
				var row string
				err := db.Raw(`SELECT jsonb_build_array(to_jsonb(i),to_jsonb(c),to_jsonb(s),to_jsonb(w))::text
                    FROM case_intakes i JOIN cases c ON c.id=?
                    JOIN case_intake_source_snapshots s ON s.intake_id=i.id AND s.saved_revision=i.metadata->>'np_description_source_revision'
                    JOIN case_intake_write_requests w ON w.intake_id=i.id AND w.source_revision=s.saved_revision
                    WHERE i.id=? AND i.status='draft' AND i.metadata->>'draft_revision'=? AND w.request_id=?`, f.caseID, f.intakeID, saved["revision"], saved["request_id"]).Scan(&row).Error
				if err != nil || row == "" {
					t.Fatalf("missing linked rows: %v", err)
				}
				return row
			}
			before := snapshot()
			close(gate.resume)
			released = true
			if mode == "sync" {
				select {
				case w := <-result:
					if w.Code != 409 || !strings.Contains(w.Body.String(), "INTAKE_DRAFT_REVISION_CONFLICT") {
						t.Fatalf("unexpected sync completion %d", w.Code)
					}
				case <-ctx.Done():
					t.Fatal("no sync receipt")
				}
			} else {
				for {
					rec, err := repo.GetCheckRecord(ctx, taskID)
					if err != nil || rec == nil {
						t.Fatal("missing async record", err)
					}
					if rec.CheckStatus == services.ConflictTaskStatusFailed && strings.Contains(fmt.Sprint(rec.CheckResult["error"]), repositories.ErrIntakeRevisionConflict.Error()) {
						break
					}
					if rec.CheckStatus != services.ConflictTaskStatusRunning {
						t.Fatalf("unexpected async terminal status %s", rec.CheckStatus)
					}
					select {
					case <-ctx.Done():
						t.Fatal(ctx.Err())
					case <-time.After(10 * time.Millisecond):
					}
				}
			}
			if snapshot() != before {
				t.Fatal("late check changed the intake, linked case, source or receipt")
			}
		})
	}
}
