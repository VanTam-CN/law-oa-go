package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"law-oa-go/internal/middleware"
)

func TestNPPostgresCreateReplayCarriesSourceHistoryFact(t *testing.T) {
	f := npPostgresFixture(t)
	key := uuid.NewString()
	// The sealed v14+86 bootstrap lacks the optional 042 column. Model the upgraded
	// installation on one connection without changing the baseline schema.
	err := f.db.Connection(func(conn *gorm.DB) error {
		conn = conn.Session(&gorm.Session{NewDB: true})
		exec := func(sql string, args ...any) {
			t.Helper()
			if err := conn.Exec(sql, args...).Error; err != nil {
				t.Fatal(err)
			}
		}
		exec("CREATE TEMP TABLE case_intakes AS SELECT * FROM public.case_intakes WHERE id=?", f.id)
		exec("CREATE TEMP TABLE case_intake_source_snapshots (intake_id uuid)")
		exec("SET search_path TO pg_temp,public")
		defer conn.Exec("RESET search_path")
		exec("ALTER TABLE pg_temp.case_intakes ADD COLUMN idempotency_key varchar(120)")
		exec("UPDATE pg_temp.case_intakes SET idempotency_key=?", key)
		h := f.handler.h.intakeWriteTransaction(conn)
		router := gin.New()
		router.Use(middleware.AuthMiddleware())
		router.POST("/api/v1/case-intakes", h.CreateCaseIntake)
		replayed := *f
		replayed.server = httptest.NewServer(router)
		defer replayed.server.Close()
		replay := func(revision string) map[string]interface{} {
			t.Helper()
			code, out := replayed.request(t, "POST", "/api/v1/case-intakes", key, map[string]any{"title": "old payload"})
			data := npData(t, code, out)
			if data["id"] != f.id || data["revision"] != revision || data["idempotent_replay"] != true {
				t.Fatal("wrong replay identity or version")
			}
			return data
		}
		if replay(f.revision)["has_source_history"] != false {
			t.Fatal("expected verified absence")
		}
		text, selection := f.candidate(t)
		code, out := f.request(t, "PUT", "/api/v1/case-intakes/"+f.id, uuid.NewString(), f.body(text, selection))
		saved := npData(t, code, out)
		rev := saved["revision"].(string)
		exec("UPDATE pg_temp.case_intakes SET metadata=(SELECT metadata FROM public.case_intakes WHERE id=?)", f.id)
		if replay(rev)["has_source_history"] != true {
			t.Fatal("source pointer missing on replay")
		}
		exec("INSERT INTO pg_temp.case_intake_source_snapshots SELECT intake_id FROM public.case_intake_source_snapshots WHERE intake_id=?", f.id)
		exec("UPDATE pg_temp.case_intakes SET metadata=metadata-'np_description_source_revision'")
		if replay(rev)["has_source_history"] != true {
			t.Fatal("detached history missing on replay")
		}
		exec("ALTER TABLE pg_temp.case_intake_source_snapshots RENAME COLUMN intake_id TO unavailable")
		if _, ok := replay(rev)["has_source_history"]; ok {
			t.Fatal("failed lookup must remain unknown")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
