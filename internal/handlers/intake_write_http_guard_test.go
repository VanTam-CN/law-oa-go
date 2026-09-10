package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// intakeWriteGuardRouter builds a minimal engine that mimics the auth
// middleware: role and user id land in the Gin context before the handler.
func intakeWriteGuardRouter(db *gorm.DB) (*gin.Engine, *DemoAggregateHandler) {
	gin.SetMode(gin.TestMode)
	h := NewDemoAggregateHandler(db)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", c.GetHeader("X-Test-Role"))
		c.Set("user_id", c.GetHeader("X-Test-User-Id"))
		c.Next()
	})
	r.PUT("/api/v1/case-intakes/:id", h.UpdateCaseIntake)
	return r, h
}

func newIntakeWriteGuardDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`CREATE TABLE case_intakes (id TEXT PRIMARY KEY, title TEXT, case_type TEXT, client_id INTEGER, created_by TEXT, status TEXT, metadata TEXT, updated_at DATETIME)`).Error; err != nil {
		t.Fatalf("create table: %v", err)
	}
	if err := db.Exec(`INSERT INTO case_intakes (id, title, case_type, client_id, created_by, status, metadata) VALUES ('intake-owned', '初始标题', 'civil', 3, '7', 'draft', '{}')`).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
	return db
}

func intakeWriteSnapshot(t *testing.T, db *gorm.DB) (int64, string) {
	t.Helper()
	var total int64
	if err := db.Table("case_intakes").Count(&total).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	var title string
	if err := db.Table("case_intakes").Select("title").Where("id = ?", "intake-owned").Scan(&title).Error; err != nil {
		t.Fatalf("title: %v", err)
	}
	return total, title
}

func performIntakeWrite(t *testing.T, r *gin.Engine, role, userID string) (int, string) {
	t.Helper()
	body := `{"title":"被告方合同纠纷二审案","case_type":"civil"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/case-intakes/intake-owned", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Role", role)
	req.Header.Set("X-Test-User-Id", userID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &payload)
	return w.Code, payload.Error.Code
}

func TestIntakeWriteDeniesPlainUserAndConflictOfficerBeforeObjectScope(t *testing.T) {
	db := newIntakeWriteGuardDB(t)
	r, _ := intakeWriteGuardRouter(db)
	beforeCount, beforeTitle := intakeWriteSnapshot(t, db)
	for _, tc := range []struct{ name, role string }{{"plain user", "user"}, {"conflict officer", "conflict_officer"}} {
		code, errCode := performIntakeWrite(t, r, tc.role, "9")
		if code != http.StatusForbidden || errCode != "INTAKE_WRITE_ROLE_FORBIDDEN" {
			t.Fatalf("%s: status=%d code=%q want 403 INTAKE_WRITE_ROLE_FORBIDDEN", tc.name, code, errCode)
		}
	}
	afterCount, afterTitle := intakeWriteSnapshot(t, db)
	if afterCount != beforeCount || afterTitle != beforeTitle {
		t.Fatalf("denied writes mutated intake: count %d->%d title %q->%q", beforeCount, afterCount, beforeTitle, afterTitle)
	}
}

func TestIntakeWriteDeniesOtherLawyerViaObjectScope(t *testing.T) {
	db := newIntakeWriteGuardDB(t)
	r, _ := intakeWriteGuardRouter(db)
	beforeCount, beforeTitle := intakeWriteSnapshot(t, db)
	code, _ := performIntakeWrite(t, r, "lawyer", "8")
	if code != http.StatusForbidden {
		t.Fatalf("other lawyer: status=%d want 403; body may reveal error code", code)
	}
	afterCount, afterTitle := intakeWriteSnapshot(t, db)
	if afterCount != beforeCount || afterTitle != beforeTitle {
		t.Fatalf("denied write mutated intake: count %d->%d title %q->%q", beforeCount, afterCount, beforeTitle, afterTitle)
	}
}
