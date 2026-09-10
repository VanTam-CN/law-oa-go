package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newIntakeCreateDB builds an in-memory case_intakes table mirroring the
// production unique constraint on intake_code so duplicate codes really fail
// the insert instead of passing silently.
func newIntakeCreateDB(t *testing.T, name string) *gorm.DB {
	t.Helper()
	// File-backed SQLite: a :memory: database is scoped per connection, and
	// concurrent httptest requests need multiple connections. The unique
	// index still enforces intake_code across all of them.
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), name+".db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec("CREATE TABLE case_intakes (" +
		"id TEXT PRIMARY KEY," +
		"intake_code TEXT NOT NULL UNIQUE," +
		"title TEXT," +
		"case_type TEXT," +
		"client_id INTEGER," +
		"created_by TEXT," +
		"priority TEXT," +
		"status TEXT," +
		"metadata TEXT," +
		"created_at DATETIME," +
		"updated_at DATETIME)").Error; err != nil {
		t.Fatalf("create table: %v", err)
	}
	return db
}

func newIntakeCreateRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewDemoAggregateHandler(db)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", c.GetHeader("X-Test-Role"))
		c.Set("user_id", c.GetHeader("X-Test-User-Id"))
		c.Next()
	})
	r.POST("/api/v1/case-intakes", h.CreateCaseIntake)
	return r
}

func postIntake(t *testing.T, r *gin.Engine) (int, map[string]interface{}) {
	t.Helper()
	const payload = `{"title":"并发接案回归","case_type":"civil"}`
	body := strings.NewReader(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/case-intakes", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Role", "lawyer")
	req.Header.Set("X-Test-User-Id", "7")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	return w.Code, resp
}

func TestCreateCaseIntakeSameSecondUniqueCodes(t *testing.T) {
	origNow := timeNow
	t.Cleanup(func() { timeNow = origNow })
	pinned := time.Date(2026, 9, 10, 13, 33, 4, 0, time.Local)
	timeNow = func() time.Time { return pinned }

	db := newIntakeCreateDB(t, "same-second")
	r := newIntakeCreateRouter(db)

	status1, resp1 := postIntake(t, r)
	status2, resp2 := postIntake(t, r)
	if status1 != http.StatusCreated || status2 != http.StatusCreated {
		t.Fatalf("status1=%d status2=%d, want both 201", status1, status2)
	}
	d1, _ := resp1["data"].(map[string]interface{})
	d2, _ := resp2["data"].(map[string]interface{})
	if d1 == nil || d2 == nil {
		t.Fatalf("missing data payloads: %v %v", resp1, resp2)
	}
	code1, _ := d1["intake_code"].(string)
	code2, _ := d2["intake_code"].(string)
	if !strings.HasPrefix(code1, "INT-20260910133304-") || !strings.HasPrefix(code2, "INT-20260910133304-") {
		t.Fatalf("codes lost readable timestamp prefix: %q %q", code1, code2)
	}
	if code1 == code2 {
		t.Fatalf("same-second intakes produced identical codes: %q", code1)
	}
	var count int64
	db.Table("case_intakes").Count(&count)
	if count != 2 {
		t.Fatalf("case_intakes rows = %d, want 2", count)
	}
}

func TestCreateCaseIntakeCodeConflictRetriesWithNewCode(t *testing.T) {
	origGen := generateIntakeCode
	t.Cleanup(func() { generateIntakeCode = origGen })

	call := 0
	generateIntakeCode = func() string {
		call++
		if call == 1 {
			return "INT-20260910133304-DUP01"
		}
		return origGen()
	}
	db := newIntakeCreateDB(t, "retry")
	r := newIntakeCreateRouter(db)
	if err := db.Exec("INSERT INTO case_intakes (id, intake_code, title, case_type, created_by, status, metadata, created_at, updated_at) " +
		"VALUES ('seed-1', 'INT-20260910133304-DUP01', 'seed', 'civil', '9', 'draft', '{}', '2026-09-10 13:33:04', '2026-09-10 13:33:04')").Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	status, resp := postIntake(t, r)
	if status != http.StatusCreated {
		t.Fatalf("status = %d body=%v, want 201 after retry", status, resp)
	}
	d, _ := resp["data"].(map[string]interface{})
	code, _ := d["intake_code"].(string)
	if code == "" || code == "INT-20260910133304-DUP01" {
		t.Fatalf("code not regenerated after conflict: %q", code)
	}
	var count int64
	db.Table("case_intakes").Count(&count)
	if count != 2 {
		t.Fatalf("case_intakes rows = %d, want 2", count)
	}
}

func TestCreateCaseIntakeConcurrentSameSecond(t *testing.T) {
	origNow := timeNow
	t.Cleanup(func() { timeNow = origNow })
	pinned := time.Date(2026, 9, 10, 13, 33, 4, 0, time.Local)
	timeNow = func() time.Time { return pinned }

	db := newIntakeCreateDB(t, "concurrent")
	r := newIntakeCreateRouter(db)

	const workers = 8
	statuses := make([]int, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			statuses[i], _ = postIntake(t, r)
		}(i)
	}
	wg.Wait()

	for i, s := range statuses {
		if s != http.StatusCreated {
			t.Fatalf("worker %d status = %d, want 201", i, s)
		}
	}
	var codes []string
	if err := db.Raw("SELECT intake_code FROM case_intakes").Scan(&codes).Error; err != nil {
		t.Fatalf("scan codes: %v", err)
	}
	if len(codes) != workers {
		t.Fatalf("rows = %d, want %d", len(codes), workers)
	}
	seen := map[string]bool{}
	for _, code := range codes {
		if seen[code] {
			t.Fatalf("duplicate code among concurrent intakes: %q", code)
		}
		seen[code] = true
		if !strings.HasPrefix(code, "INT-20260910133304-") {
			t.Fatalf("code missing readable prefix: %q", code)
		}
	}
}
