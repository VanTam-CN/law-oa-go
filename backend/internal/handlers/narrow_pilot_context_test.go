package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"law-oa-go/internal/config"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/models"
)

func npContextFixture(t *testing.T) (*NarrowPilotHandler, *gin.Engine, string, *gorm.DB) {
	t.Helper()
	h, db := newDraftRevisionDB(t)
	if err := db.Exec("ALTER TABLE users ADD COLUMN deleted_at DATETIME").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.AuthTokenSession{}); err != nil {
		t.Fatal(err)
	}
	middleware.InitJWT(&config.Config{JWT: config.JWTConfig{Secret: strings.Repeat("n", 40)}})
	token, _, err := middleware.GenerateToken(7, "np-lawyer", "lawyer")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := middleware.ValidateToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.AuthTokenSession{ID: uuid.NewString(), UserID: 7, AccessTokenUUID: claims.UUID, RefreshTokenUUID: uuid.NewString(), AccessTokenExpires: time.Now().Add(15 * time.Minute), RefreshTokenExpires: time.Now().Add(time.Hour)}).Error; err != nil {
		t.Fatal(err)
	}
	t.Setenv("NP_ASSISTANT_ACTORS", "7")
	n := NewNarrowPilotHandler(h)
	r := gin.New()
	r.Use(middleware.AuthMiddleware())
	r.POST("/contexts", n.CreateContext)
	r.GET("/contexts/:id", n.GetContext)
	r.DELETE("/contexts/:id", n.DeleteContext)
	r.POST("/contexts/:id/turns", n.Turn)
	return n, r, token, db
}
func npCall(t *testing.T, r http.Handler, token, method, path string, body interface{}) (int, map[string]interface{}) {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(raw))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var out map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode status %d: %s", w.Code, w.Body.String())
	}
	return w.Code, out
}
func npData(t *testing.T, code int, out map[string]interface{}) map[string]interface{} {
	t.Helper()
	if code != 200 {
		t.Fatalf("status %d: %v", code, out)
	}
	return out["data"].(map[string]interface{})
}

func TestNPMultiTurnInvalidationAndSessionExpiry(t *testing.T) {
	n, r, token, db := npContextFixture(t)
	code, out := npCall(t, r, token, "POST", "/contexts", map[string]interface{}{"intake_id": "intake-rev", "expected_revision": "legacy-intake-rev"})
	ctx := npData(t, code, out)
	id := ctx["id"].(string)
	expiry, _ := time.Parse(time.RFC3339Nano, ctx["expires_at"].(string))
	if time.Until(expiry) > 16*time.Minute {
		t.Fatal("expiry exceeded actual access session")
	}
	turn := map[string]interface{}{"request_id": uuid.NewString(), "expected_context_version": 1, "text": "虚构：交付日期尚待核对。"}
	code, out = npCall(t, r, token, "POST", "/contexts/"+id+"/turns", turn)
	first := npData(t, code, out)
	oldID := first["candidate"].(map[string]interface{})["id"].(string)
	code, out = npCall(t, r, token, "POST", "/contexts/"+id+"/turns", turn)
	if npData(t, code, out)["context_version"] != float64(2) {
		t.Fatal("duplicate accepted twice")
	}
	turn["text"] = "异参"
	code, _ = npCall(t, r, token, "POST", "/contexts/"+id+"/turns", turn)
	if code != 409 {
		t.Fatal("reused turn key not rejected")
	}
	code, out = npCall(t, r, token, "POST", "/contexts/"+id+"/turns", map[string]interface{}{"request_id": uuid.NewString(), "expected_context_version": 2, "text": "补充：尚未确认具体日期。"})
	latest := npData(t, code, out)["candidate"].(map[string]interface{})
	if latest["id"] == oldID || len(latest["sources"].([]interface{})) != 2 {
		t.Fatal("candidate not replaced or accepted sources lost")
	}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("PUT", "/", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)
	c.Set("user_id", uint(7))
	selection := &npSourceSelection{ContextID: id, ExpectedContextVersion: 2, Fields: map[string]npFieldSelection{"description": {CandidateID: oldID, RequestedMode: "candidate_sources", Reviewed: true}}}
	if _, err := n.prepareSource(c, map[string]interface{}{"id": "intake-rev", "metadata": "{}"}, selection, first["candidate"].(map[string]interface{})["description"].(string)); err == nil {
		t.Fatal("old candidate retained eligibility")
	}
	if err := db.Table("auth_token_sessions").Where("user_id = ?", 7).Update("access_revoked_at", time.Now()).Error; err != nil {
		t.Fatal(err)
	}
	code, _ = npCall(t, r, token, "GET", "/contexts/"+id, nil)
	if code != 401 {
		t.Fatalf("revoked access read status %d", code)
	}
}

func TestNPContextExpiryAndChangedDraft(t *testing.T) {
	n, r, token, db := npContextFixture(t)
	code, out := npCall(t, r, token, "POST", "/contexts", map[string]interface{}{"intake_id": "intake-rev", "expected_revision": "legacy-intake-rev"})
	id := npData(t, code, out)["id"].(string)
	n.mu.Lock()
	n.contexts[id].Touched = time.Now().Add(-31 * time.Minute)
	n.mu.Unlock()
	code, _ = npCall(t, r, token, "GET", "/contexts/"+id, nil)
	if code != 409 {
		t.Fatal("idle context read succeeded")
	}
	code, out = npCall(t, r, token, "POST", "/contexts", map[string]interface{}{"intake_id": "intake-rev", "expected_revision": "legacy-intake-rev"})
	id = npData(t, code, out)["id"].(string)
	stampDraftRevision(t, db, "new-revision")
	code, _ = npCall(t, r, token, "GET", "/contexts/"+id, nil)
	if code != 409 {
		t.Fatal("changed object still delivered applicable candidate")
	}
}

func TestNPContextEndReleasesExpiryTimer(t *testing.T) {
	n, r, token, _ := npContextFixture(t)
	code, out := npCall(t, r, token, "POST", "/contexts", map[string]interface{}{"intake_id": "intake-rev", "expected_revision": "legacy-intake-rev"})
	id := npData(t, code, out)["id"].(string)
	n.mu.Lock()
	prior := n.contexts[id]
	n.mu.Unlock()
	code, out = npCall(t, r, token, "DELETE", "/contexts/"+id, map[string]interface{}{"expected_context_version": 1})
	npData(t, code, out)
	n.mu.Lock()
	_, present := n.contexts[id]
	n.mu.Unlock()
	if present {
		t.Fatal("ended context retained")
	}
	wasActive := prior.timer.Reset(time.Hour)
	prior.timer.Stop()
	if wasActive {
		t.Fatal("ended context retained its expiry timer")
	}
}
