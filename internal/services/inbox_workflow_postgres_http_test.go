package services_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"law-oa-go/internal/handlers"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

func TestInboxB2PostgresHTTPListStatsGetAndHome(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, approvalID := services.InboxB2PostgresHTTPFixture(t)
	inbox := handlers.NewInboxHandler(services.NewInboxService(repositories.NewInboxRepository(db), repositories.NewUserRepository(db)))
	home := handlers.NewDemoAggregateHandler(db)
	home.SetAuthorizationService(services.NewAuthorizationService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db), nil))
	for _, actor := range []struct {
		id                   uint
		role, bucket, action string
	}{
		{22, "partner", "needs_action", "decide_approval"}, {11, "lawyer", "waiting", "view_only"},
	} {
		t.Run(actor.role, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) { c.Set("user_id", actor.id); c.Set("role", actor.role); c.Next() })
			router.GET("/inbox", inbox.ListInboxItems)
			router.GET("/inbox/stats", inbox.GetInboxStats)
			router.GET("/inbox/:id", inbox.GetInboxItem)
			router.GET("/home", home.CommandCenter)
			get := func(path string, target interface{}) {
				t.Helper()
				w := httptest.NewRecorder()
				router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
				require.Equal(t, http.StatusOK, w.Code, w.Body.String())
				var envelope struct {
					Success bool            `json:"success"`
					Data    json.RawMessage `json:"data"`
				}
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &envelope))
				require.True(t, envelope.Success, w.Body.String())
				require.NoError(t, json.Unmarshal(envelope.Data, target))
			}
			var page services.ListInboxItemsResponse
			get("/inbox?page=1&page_size=20&workflow_bucket="+actor.bucket, &page)
			require.Len(t, page.Items, 1)
			require.EqualValues(t, 1, page.Pagination.Total)
			require.Equal(t, approvalID, page.Items[0].WorkflowAction.ResourceID)
			require.Equal(t, actor.action, page.Items[0].WorkflowAction.ActionCode)
			var stats services.InboxStatsResponse
			get("/inbox/stats", &stats)
			require.Equal(t, page.Stats, &stats)
			var detail services.InboxItemResponse
			get(fmt.Sprintf("/inbox/%d", page.Items[0].ID), &detail)
			require.Equal(t, page.Items[0].WorkflowAction, detail.WorkflowAction)
			var dashboard struct {
				Todos []services.InboxItemResponse `json:"todo_items"`
				Stats services.InboxStatsResponse  `json:"inbox_stats"`
			}
			get("/home", &dashboard)
			require.Equal(t, stats, dashboard.Stats)
			if actor.bucket == "needs_action" {
				require.Len(t, dashboard.Todos, 1)
				require.Equal(t, detail.ID, dashboard.Todos[0].ID)
				require.Equal(t, detail.WorkflowAction, dashboard.Todos[0].WorkflowAction)
				require.EqualValues(t, 1, stats.NeedsAction)
			} else {
				require.Empty(t, dashboard.Todos, "waiting tasks are not actionable home tasks")
				require.EqualValues(t, 1, stats.Waiting)
			}
			t.Logf("B2 real HTTP actor=%s list=1 get_id=%d stats_needs_action=%d stats_waiting=%d home_todos=%d home_stats_equal=true", actor.role, detail.ID, stats.NeedsAction, stats.Waiting, len(dashboard.Todos))
		})
	}
}
