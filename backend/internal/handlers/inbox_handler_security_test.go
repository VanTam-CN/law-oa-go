package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

const (
	inboxOwnerID  uint = 101
	inboxOtherID  uint = 202
	inboxOwnerUID uint = 11
	inboxOtherUID uint = 22
)

func TestInboxItemHandlersRejectCrossUserAccess(t *testing.T) {
	router, db := newInboxOwnershipTestRouter(t, inboxOwnerUID)

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "get", method: http.MethodGet, path: "/inbox/202"},
		{name: "update", method: http.MethodPut, path: "/inbox/202", body: `{"title":"unauthorized update"}`},
		{name: "mark read", method: http.MethodPut, path: "/inbox/202/read"},
		{name: "complete", method: http.MethodPut, path: "/inbox/202/complete"},
		{name: "snooze", method: http.MethodPut, path: "/inbox/202/snooze", body: `{"until":"2030-01-02T09:00:00Z"}`},
		{name: "delete", method: http.MethodDelete, path: "/inbox/202"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			recorder := performInboxRequest(router, tc.method, tc.path, tc.body)
			if recorder.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusNotFound, recorder.Body.String())
			}
			assertInboxItemUnchanged(t, db, inboxOtherID)
		})
	}
}

func TestInboxItemHandlersAllowOwnerAccess(t *testing.T) {
	router, db := newInboxOwnershipTestRouter(t, inboxOwnerUID)

	if recorder := performInboxRequest(router, http.MethodGet, "/inbox/101", ""); recorder.Code != http.StatusOK {
		t.Fatalf("owner get status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if recorder := performInboxRequest(router, http.MethodPut, "/inbox/101", `{"title":"owner updated title"}`); recorder.Code != http.StatusOK {
		t.Fatalf("owner update status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if recorder := performInboxRequest(router, http.MethodPut, "/inbox/101", `{"title":"owner updated title"}`); recorder.Code != http.StatusOK {
		t.Fatalf("owner unchanged update status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if recorder := performInboxRequest(router, http.MethodPut, "/inbox/101/read", ""); recorder.Code != http.StatusOK {
		t.Fatalf("owner mark read status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if recorder := performInboxRequest(router, http.MethodPut, "/inbox/101/snooze", `{"until":"2030-01-02T09:00:00Z"}`); recorder.Code != http.StatusOK {
		t.Fatalf("owner snooze status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if recorder := performInboxRequest(router, http.MethodPut, "/inbox/101/complete", ""); recorder.Code != http.StatusOK {
		t.Fatalf("owner complete status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var item models.InboxItem
	if err := db.First(&item, inboxOwnerID).Error; err != nil {
		t.Fatalf("load owner inbox item: %v", err)
	}
	if item.Title != "owner updated title" || !item.IsRead || item.ReadAt == nil || !item.IsCompleted || item.CompletedAt == nil || item.SnoozedUntil == nil {
		t.Fatalf("owner changes were not persisted: %+v", item)
	}

	if recorder := performInboxRequest(router, http.MethodDelete, "/inbox/101", ""); recorder.Code != http.StatusConflict {
		t.Fatalf("owner delete status = %d, want %d; body = %s", recorder.Code, http.StatusConflict, recorder.Body.String())
	}
	var retained models.InboxItem
	if err := db.First(&retained, inboxOwnerID).Error; err != nil {
		t.Fatalf("load retained owner inbox item: %v", err)
	}
	if retained.Title != "owner updated title" || !retained.IsCompleted {
		t.Fatalf("owner inbox item was changed by rejected delete: %+v", retained)
	}
}

func TestInboxItemHandlersRequireAuthenticatedUser(t *testing.T) {
	router, _ := newInboxOwnershipTestRouter(t, 0)
	recorder := performInboxRequest(router, http.MethodGet, "/inbox/101", "")
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusUnauthorized, recorder.Body.String())
	}
}

func newInboxOwnershipTestRouter(t *testing.T, userID uint) (*gin.Engine, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/inbox.db"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.InboxItem{}); err != nil {
		t.Fatalf("migrate inbox item: %v", err)
	}
	seedInboxOwnershipItems(t, db)

	handler := NewInboxHandler(services.NewInboxService(repositories.NewInboxRepository(db), nil))
	router := gin.New()
	router.Use(func(c *gin.Context) {
		if userID != 0 {
			c.Set("user_id", userID)
		}
		c.Next()
	})
	router.GET("/inbox/:id", handler.GetInboxItem)
	router.PUT("/inbox/:id", handler.UpdateInboxItem)
	router.DELETE("/inbox/:id", handler.DeleteInboxItem)
	router.PUT("/inbox/:id/read", handler.MarkAsRead)
	router.PUT("/inbox/:id/complete", handler.MarkAsCompleted)
	router.PUT("/inbox/:id/snooze", handler.SnoozeInboxItem)

	return router, db
}

func seedInboxOwnershipItems(t *testing.T, db *gorm.DB) {
	t.Helper()
	items := []models.InboxItem{
		{
			ID:         inboxOwnerID,
			UserID:     inboxOwnerUID,
			SourceType: "task",
			SourceID:   1,
			Title:      "owner inbox item",
			Priority:   "medium",
		},
		{
			ID:         inboxOtherID,
			UserID:     inboxOtherUID,
			SourceType: "task",
			SourceID:   2,
			Title:      "other user inbox item",
			Priority:   "high",
		},
	}
	if err := db.Create(&items).Error; err != nil {
		t.Fatalf("seed inbox items: %v", err)
	}
}

func performInboxRequest(router http.Handler, method, path, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func assertInboxItemUnchanged(t *testing.T, db *gorm.DB, id uint) {
	t.Helper()

	var item models.InboxItem
	if err := db.First(&item, id).Error; err != nil {
		t.Fatalf("load target inbox item: %v", err)
	}
	if item.Title != "other user inbox item" || item.IsRead || item.ReadAt != nil || item.IsCompleted || item.CompletedAt != nil || item.SnoozedUntil != nil || item.SnoozedCount != 0 {
		t.Fatalf("cross-user request changed inbox item: %+v", item)
	}
}
