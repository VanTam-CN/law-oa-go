package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

func TestInboxPublicAPIDeniesForgedWorkflowFields(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "api.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.InboxItem{}, &models.User{}))
	require.NoError(t, db.Create(&models.User{ID: 1, Username: "personal", Email: "personal@test.invalid", Password: "x", Role: "lawyer", Status: "active"}).Error)
	h := NewInboxHandler(services.NewInboxService(repositories.NewInboxRepository(db), repositories.NewUserRepository(db)))
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("user_id", uint(1)); c.Next() })
	r.POST("/inbox", h.CreateInboxItem)
	r.PUT("/inbox/:id", h.UpdateInboxItem)
	r.GET("/inbox", h.ListInboxItems)
	r.GET("/inbox/stats", h.GetInboxStats)
	base := map[string]interface{}{"title": "个人提醒", "priority": "medium"}
	for _, field := range []string{"completion_mode", "source_ref", "workflow_task_key", "workflow_group", "workflow_purpose", "workflow_version", "assigned_to", "workflow_action", "is_completed", "closed_reason", "archived_at"} {
		t.Run(field, func(t *testing.T) {
			body := map[string]interface{}{}
			for k, v := range base {
				body[k] = v
			}
			body[field] = nil
			raw, _ := json.Marshal(body)
			rec := performInboxRequest(r, http.MethodPost, "/inbox", string(raw))
			require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
		})
	}
	for _, suffix := range []string{`,"source_type":"approval","source_id":9`, `,"source_type":"handoff","source_id":9`, `,"user_id":2`, `,"source_type":"task","source_id":9`} {
		rec := performInboxRequest(r, http.MethodPost, "/inbox", `{"title":"伪造","priority":"medium"`+suffix+`}`)
		require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
	}
	var count int64
	require.NoError(t, db.Model(&models.InboxItem{}).Count(&count).Error)
	require.Zero(t, count)
	rec := performInboxRequest(r, http.MethodPost, "/inbox", `{"title":"正常个人提醒","priority":"medium"}`)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var item models.InboxItem
	require.NoError(t, db.First(&item).Error)
	require.Equal(t, "manual", item.CompletionMode)
	require.Zero(t, item.SourceID)
	require.Empty(t, item.SourceRef)
	for _, body := range []string{`{"completion_mode":"manual"}`, `{"source_ref":"approval-uuid"}`, `{"source_type":"task"}`, `{"source_id":0}`, `{"user_id":2}`, `{"is_completed":true}`, `{"archived_at":null}`} {
		rec = performInboxRequest(r, http.MethodPut, fmt.Sprintf("/inbox/%d", item.ID), body)
		require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
	}
	rec = performInboxRequest(r, http.MethodGet, "/inbox", "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), `"needs_action":1`)
	rec = performInboxRequest(r, http.MethodGet, "/inbox/stats", "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), `"needs_action":1`)
}

func TestInboxLegacyHTTPArchiveContract(t *testing.T) {
	r, db := newInboxOwnershipTestRouter(t, inboxOwnerUID)
	h := NewInboxHandler(services.NewInboxService(repositories.NewInboxRepository(db), nil))
	r.PUT("/inbox/:id/archive", h.ArchiveInboxItem)
	legacy := models.InboxItem{UserID: inboxOwnerUID, SourceType: "approval", SourceID: 777, Title: "不可泄漏身份", Content: "不可泄漏意见"}
	require.NoError(t, db.Create(&legacy).Error)
	path := fmt.Sprintf("/inbox/%d", legacy.ID)
	rec := performInboxRequest(r, http.MethodPut, path+"/complete", "")
	require.Equal(t, 409, rec.Code)
	require.Contains(t, rec.Body.String(), "INBOX_ARCHIVE_REQUIRED")
	rec = performInboxRequest(r, http.MethodGet, path, "")
	require.Equal(t, 200, rec.Code)
	require.NotContains(t, rec.Body.String(), "不可泄漏")
	require.Contains(t, rec.Body.String(), "legacy_unresolved")
	rec = performInboxRequest(r, http.MethodPut, path+"/archive", "")
	require.Equal(t, 200, rec.Code)
	require.NoError(t, db.First(&legacy, legacy.ID).Error)
	require.False(t, legacy.IsCompleted)
	require.NotNil(t, legacy.ArchivedAt)
	require.Equal(t, "不可泄漏身份", legacy.Title)
	rec = performInboxRequest(r, http.MethodPut, "/inbox/202/archive", "")
	require.Equal(t, 404, rec.Code)
}

func TestInboxWorkflowHTTPMutationAndSnoozeContract(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "workflow-api.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.InboxItem{}, &models.User{}, &models.ApprovalRequest{}, &models.ApprovalRecord{}))
	for _, id := range []uint{1, 2} {
		require.NoError(t, db.Create(&models.User{ID: id, Username: fmt.Sprint("lawyer", id), Email: fmt.Sprintf("lawyer%d@test.invalid", id), Password: "x", Role: "lawyer", Status: "active"}).Error)
	}
	due := time.Now().Add(24 * time.Hour)
	a := &models.ApprovalRequest{ID: "api-approval", RequestNumber: "APR-API", ApplicantID: "2", CurrentApproverID: "1", CurrentStage: "审批", Status: models.ApprovalStatusSubmitted, Type: "contract", TimeoutAt: &due, Title: "私有审批标题"}
	require.NoError(t, db.Create(a).Error)
	require.NoError(t, db.Transaction(func(tx *gorm.DB) error { return services.ReconcileApprovalInboxInTx(tx, a) }))
	var item models.InboxItem
	require.NoError(t, db.Where("user_id = ?", 1).First(&item).Error)
	h := NewInboxHandler(services.NewInboxService(repositories.NewInboxRepository(db), repositories.NewUserRepository(db)))
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("user_id", uint(1)); c.Next() })
	r.PUT("/inbox/:id/complete", h.MarkAsCompleted)
	r.PUT("/inbox/:id/archive", h.ArchiveInboxItem)
	r.PUT("/inbox/:id/snooze", h.SnoozeInboxItem)
	r.PUT("/inbox/:id", h.UpdateInboxItem)
	r.DELETE("/inbox/:id", h.DeleteInboxItem)
	r.GET("/inbox", h.ListInboxItems)
	r.GET("/inbox/:id", h.GetInboxItem)
	path := fmt.Sprintf("/inbox/%d", item.ID)
	for _, suffix := range []string{"/complete", "/archive"} {
		rec := performInboxRequest(r, http.MethodPut, path+suffix, "")
		require.Equal(t, 409, rec.Code, rec.Body.String())
		require.Contains(t, rec.Body.String(), "INBOX_WORKFLOW_ACTION_REQUIRED")
	}
	rec := performInboxRequest(r, http.MethodPut, path, `{"due_date":"2030-01-01"}`)
	require.Equal(t, 409, rec.Code, rec.Body.String())
	rec = performInboxRequest(r, http.MethodDelete, path, "")
	require.Equal(t, 409, rec.Code)
	rec = performInboxRequest(r, http.MethodPut, path+"/snooze", `{"until":"2030-01-02T09:00:00Z"}`)
	require.Equal(t, 200, rec.Code, rec.Body.String())
	rec = performInboxRequest(r, http.MethodGet, "/inbox?workflow_bucket=needs_action", "")
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), `"needs_action":1`)
	require.NoError(t, db.First(&item, item.ID).Error)
	require.False(t, item.IsCompleted)
	require.Equal(t, due.Unix(), item.DueDate.Unix())
	require.NotNil(t, item.SnoozedUntil)
	require.NoError(t, db.First(a, "id = ?", a.ID).Error)
	require.Equal(t, due.Unix(), a.TimeoutAt.Unix())
	require.Equal(t, models.ApprovalStatusSubmitted, a.Status)
	require.NoError(t, db.Model(&models.User{}).Where("id = 1").Update("status", "inactive").Error)
	rec = performInboxRequest(r, http.MethodGet, path, "")
	require.Equal(t, 404, rec.Code)
	require.NotContains(t, rec.Body.String(), "私有审批标题")
	rec = performInboxRequest(r, http.MethodGet, "/inbox", "")
	require.Equal(t, 200, rec.Code)
	require.Contains(t, rec.Body.String(), `"total":0`)
}
