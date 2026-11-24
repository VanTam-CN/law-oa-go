//go:build np_m1_pg

package m1_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"law-oa-go/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"law-oa-go/internal/config"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/models"
	"law-oa-go/internal/router"
)

// This target is run explicitly against the dedicated, bootstrapped M1 database.
func isolatedDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(postgres.Open("host=127.0.0.1 port=18790 user=np_m1 dbname=np_first_slice_dev sslmode=disable"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	var identity string
	if err = db.Raw("SELECT current_database() || ':' || current_user").Scan(&identity).Error; err != nil || identity != "np_first_slice_dev:np_m1" {
		t.Fatalf("wrong test DB: %v", err)
	}
	return db
}

func registerIntakeTestSession(t *testing.T, db *gorm.DB, userID uint, token string) {
	t.Helper()
	claims, err := middleware.ValidateToken(token)
	if err != nil {
		t.Fatal(err)
	}
	session := models.AuthTokenSession{ID: uuid.NewString(), UserID: userID, AccessTokenUUID: claims.UUID, RefreshTokenUUID: uuid.NewString(), AccessTokenExpires: time.Now().Add(time.Hour), RefreshTokenExpires: time.Now().Add(2 * time.Hour)}
	if err := db.Create(&session).Error; err != nil {
		t.Fatal(err)
	}
}

func TestNativeConcurrentDraftAndHistory(t *testing.T) {
	db := isolatedDB(t)
	var err error
	name := uuid.NewString()
	user := models.User{Username: "m1-" + name, Name: "M1 synthetic lawyer", Email: name + "@np.invalid", Password: "disabled-synthetic", Role: "lawyer", Status: "active"}
	if err = db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	client := models.Client{Name: "M1 synthetic client", Email: name + "-client@np.invalid", CreatedBy: user.ID, Type: "企业"}
	if err = db.Create(&client).Error; err != nil {
		t.Fatal(err)
	}
	run := func(q string, args ...interface{}) {
		t.Helper()
		if err := db.Exec(q, args...).Error; err != nil {
			t.Fatal(err)
		}
	}
	run("INSERT INTO case_intakes(id,intake_code,title,created_by,client_id,status,metadata) VALUES(?,?,?,?,?,'draft','{\"precision\":9007199254740993}')", name, name, "synthetic", user.ID, client.ID)
	t.Setenv("JWT_SECRET", "synthetic-only-for-the-isolated-m1-test")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	middleware.InitJWT(cfg)
	r := gin.New()
	if err := router.Init(r, db, nil); err != nil {
		t.Fatal(err)
	}
	tokenFor := func(id uint) string {
		token, _, err := middleware.GenerateToken(id, "synthetic", "lawyer")
		if err != nil {
			t.Fatal(err)
		}
		registerIntakeTestSession(t, db, id, token)
		return token
	}
	owner := tokenFor(user.ID)
	foreignUser := models.User{Username: "other-" + name, Name: "other", Email: "other-" + name + "@np.invalid", Password: "disabled", Role: "lawyer", Status: "active"}
	if err := db.Create(&foreignUser).Error; err != nil {
		t.Fatal(err)
	}
	foreign := tokenFor(foreignUser.ID)
	for _, role := range []string{"client", "opposing_party"} {
		run("INSERT INTO case_intake_parties(intake_id,entity_name,party_role,identity_type,identity_number_ciphertext,identity_number_digest,metadata) VALUES(?,?,?,'ID_CARD',?,?,?::jsonb)", name, role, role, "private-cipher-"+role, "private-digest-"+role, `{"nested":{"identity_number":"private-plain-`+role+`"}}`)
	}
	request := func(method, path, token string, body []byte) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, "/api/v1"+path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}
	type result struct {
		code  int
		value string
		body  []byte
	}
	put := func(desc, rev string) result {
		body, _ := json.Marshal(map[string]string{"description": desc, "expected_revision": rev})
		w := request(http.MethodPut, "/case-intakes/"+name, owner, body)
		return result{w.Code, desc, w.Body.Bytes()}
	}
	start := make(chan struct{})
	out := make(chan result, 2)
	for _, v := range []string{"A", "B"} {
		go func(v string) { <-start; out <- put(v, "legacy-"+name) }(v)
	}
	close(start)
	a, b := <-out, <-out
	if a.code != 200 {
		a, b = b, a
	}
	if a.code != 200 || b.code != 409 {
		t.Fatalf("concurrent HTTP codes %d %d, %s %s", a.code, b.code, a.body, b.body)
	}
	var row struct{ Description, Rev, PreciseValue string }
	q := "SELECT description, metadata->>'draft_revision' rev,metadata->>'precision' AS precise_value FROM case_intakes WHERE id=?"
	if err = db.Raw(q, name).Scan(&row).Error; err != nil || row.Description != a.value || row.Rev == "" || row.PreciseValue != "9007199254740993" {
		t.Fatalf("saved revision/content/precision mismatch: %v", err)
	}
	for _, path := range []string{"/case-intakes/" + name, "/cases/intake-workbench/" + name, "/case-intakes"} {
		w := request(http.MethodGet, path, owner, nil)
		if w.Code != 200 || !strings.Contains(w.Body.String(), row.Rev) || strings.Contains(w.Body.String(), "private-") {
			t.Fatalf("owner read %s: code %d, version or redaction mismatch", path, w.Code)
		}
		if path != "/case-intakes" {
			var resp struct {
				Data struct {
					Intake      struct{ Description string }
					Description string
				}
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil || (resp.Data.Intake.Description != a.value && resp.Data.Description != a.value) {
				t.Fatalf("owner reopen %s: saved summary missing", path)
			}
		}
		w = request(http.MethodGet, path, foreign, nil)
		if path == "/case-intakes" {
			if w.Code != 200 || strings.Contains(w.Body.String(), name) {
				t.Fatalf("foreign list exposed intake or failed: %d", w.Code)
			}
		} else if w.Code != 403 {
			t.Fatalf("foreign detail %s: %d", path, w.Code)
		}
	}
	for token, expected := range map[string]int{owner: 1, foreign: 0} {
		w := request(http.MethodGet, "/dashboard/command-center", token, nil)
		var response struct {
			Data struct{ Workflow struct{ Intake *int } }
		}
		if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &response) != nil || response.Data.Workflow.Intake == nil || *response.Data.Workflow.Intake != expected || strings.Contains(w.Body.String(), "private-") {
			t.Fatalf("dashboard intake count or redaction mismatch: %d", w.Code)
		}
	}
	w := request(http.MethodGet, "/case-intakes/"+uuid.NewString(), owner, nil)
	if w.Code != 404 || !json.Valid(w.Body.Bytes()) {
		t.Fatal("missing intake did not reach business handler")
	}
	// Hold the same row lock a future strict writer uses, then establish history.
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	defer tx.Rollback()
	if err = tx.Exec("SELECT 1 FROM case_intakes WHERE id=? FOR UPDATE", name).Error; err != nil {
		t.Fatal(err)
	}
	pending := make(chan result, 1)
	go func() { pending <- put("late full save", row.Rev) }()
	var blocker int
	if err = tx.Raw("SELECT pg_backend_pid()").Scan(&blocker).Error; err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		var waiting bool
		if err := db.Raw("SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE ? = ANY(pg_blocking_pids(pid)))", blocker).Scan(&waiting).Error; err != nil {
			t.Fatal(err)
		}
		if waiting {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("native write did not wait on the intake row")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err = tx.Exec("INSERT INTO case_intake_source_snapshots(intake_id,saved_revision,serverpack) VALUES(?,?,'{}')", name, row.Rev).Error; err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit().Error; err != nil {
		t.Fatal(err)
	}
	got := <-pending
	if got.code != 409 || !bytes.Contains(got.body, []byte("UPDATE_PROTOCOL_REQUIRED")) {
		t.Fatalf("historical draft accepted full write: %d", got.code)
	}
	previous := row
	if err = db.Raw(q, name).Scan(&row).Error; err != nil || row != previous {
		t.Fatalf("protected draft changed: %v", err)
	}
}

func TestSchemaRequiresFrozenKeysAndTypes(t *testing.T) {
	db := isolatedDB(t)
	if err := database.RequireNarrowPilotSchema(db); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, ddl string }{
		{"wrong-index-columns", "ALTER TABLE case_intake_source_snapshots DROP CONSTRAINT np_source_intake_revision_unique; CREATE UNIQUE INDEX np_source_intake_revision_unique ON case_intake_source_snapshots(id)"},
		{"missing-fk", "ALTER TABLE case_intake_write_requests DROP CONSTRAINT case_intake_write_requests_intake_id_fkey"},
		// Unbounded varchar preserves existing UUID values while differing from the required text type.
		{"wrong-column-type", "ALTER TABLE case_intake_write_requests ALTER COLUMN source_revision TYPE varchar"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rollback := errors.New("rollback temporary schema variant")
			err := db.Transaction(func(tx *gorm.DB) error {
				for _, stmt := range strings.Split(tc.ddl, ";") {
					if err := tx.Exec(stmt).Error; err != nil {
						return err
					}
				}
				r := gin.New()
				if err := router.Init(r, tx, nil); err == nil || len(r.Routes()) != 0 {
					t.Error("business router mounted with invalid schema")
				}
				if err := database.RequireNarrowPilotSchema(tx); err == nil {
					t.Error("guard accepted schema variant")
				}
				return rollback
			})
			if !errors.Is(err, rollback) {
				t.Fatal(err)
			}
		})
	}
}
