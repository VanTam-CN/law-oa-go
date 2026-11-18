package services

import (
	"testing"

	"github.com/stretchr/testify/require"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// The batched task projection must apply the same grant model as
// AuthorizationService.CanReadCase and CaseTaskService.caseTaskAssigneeAllowed:
// an explicit ethical-wall whitelist row is a valid read grant for a
// collaborator who is neither the case lawyer nor the creator. Revoked,
// missing, and deleted cases stay fail-closed.
func TestCaseTaskInboxProjection_AllowsWhitelistedCollaborator(t *testing.T) {
	t.Run("whitelisted_collaborator_sees_open_action", func(t *testing.T) {
		db := conflictReviewInboxDB(t)
		require.NoError(t, db.Create(&models.User{ID: 30, Username: "collab", Name: "协作律师",
			Email: "collab@example.test", Password: "x", Role: "lawyer", Status: "active"}).Error)
		require.NoError(t, db.Model(&models.Case{}).Where("id = ?", 100).
			Updates(map[string]interface{}{"client_id": 1, "ethical_wall_enabled": true}).Error)
		require.NoError(t, db.Exec("INSERT INTO case_ethical_wall_whitelist (case_id, user_id, granted_by) VALUES (100, 30, 10)").Error)
		svc := NewInboxService(repositories.NewInboxRepository(db), repositories.NewUserRepository(db))
		task := caseTaskActionRow{ID: 7, CaseID: 100, Title: "协作任务（虚构）", Status: "open", OwnerID: 30}
		item := &models.InboxItem{ID: 901, UserID: 30, SourceRef: "7", WorkflowTaskKey: "case_task:7", CompletionMode: "workflow"}
		visible := svc.prefetchCaseAccess(30, db, map[string]caseTaskActionRow{"7": task}, []*models.InboxItem{item})
		require.True(t, visible["7"], "whitelist grant must keep the task reachable")
		action := &WorkflowAction{}
		fillCaseTaskAction(action, item, task, visible["7"])
		require.Equal(t, workflowActionCodeOpenCaseTask, action.ActionCode)
	})
	t.Run("revoked_whitelist_fails_closed", func(t *testing.T) {
		db := conflictReviewInboxDB(t)
		require.NoError(t, db.Create(&models.User{ID: 30, Username: "collab", Name: "协作律师",
			Email: "collab@example.test", Password: "x", Role: "lawyer", Status: "active"}).Error)
		require.NoError(t, db.Model(&models.Case{}).Where("id = ?", 100).
			Updates(map[string]interface{}{"client_id": 1, "ethical_wall_enabled": true}).Error)
		svc := NewInboxService(repositories.NewInboxRepository(db), repositories.NewUserRepository(db))
		task := caseTaskActionRow{ID: 7, CaseID: 100, Title: "协作任务（虚构）", Status: "open", OwnerID: 30}
		item := &models.InboxItem{ID: 901, UserID: 30, SourceRef: "7", WorkflowTaskKey: "case_task:7", CompletionMode: "workflow"}
		visible := svc.prefetchCaseAccess(30, db, map[string]caseTaskActionRow{"7": task}, []*models.InboxItem{item})
		require.False(t, visible["7"], "no grant means no access")
		action := &WorkflowAction{}
		fillCaseTaskAction(action, item, task, visible["7"])
		require.NotEqual(t, workflowActionCodeOpenCaseTask, action.ActionCode)
	})
	t.Run("missing_and_deleted_case_fail_closed", func(t *testing.T) {
		db := conflictReviewInboxDB(t)
		require.NoError(t, db.Create(&models.User{ID: 30, Username: "collab", Name: "协作律师",
			Email: "collab@example.test", Password: "x", Role: "lawyer", Status: "active"}).Error)
		svc := NewInboxService(repositories.NewInboxRepository(db), repositories.NewUserRepository(db))
		task := caseTaskActionRow{ID: 8, CaseID: 555, Title: "缺失案件任务（虚构）", Status: "open", OwnerID: 30}
		item := &models.InboxItem{ID: 902, UserID: 30, SourceRef: "8", WorkflowTaskKey: "case_task:8", CompletionMode: "workflow"}
		visible := svc.prefetchCaseAccess(30, db, map[string]caseTaskActionRow{"8": task}, []*models.InboxItem{item})
		require.False(t, visible["8"], "missing case stays invisible")

		require.NoError(t, db.Exec("UPDATE cases SET deleted_at = CURRENT_TIMESTAMP WHERE id = 100").Error)
		deletedTask := caseTaskActionRow{ID: 9, CaseID: 100, Title: "已删案件任务（虚构）", Status: "open", OwnerID: 30}
		deletedItem := &models.InboxItem{ID: 903, UserID: 30, SourceRef: "9", WorkflowTaskKey: "case_task:9", CompletionMode: "workflow"}
		visible = svc.prefetchCaseAccess(30, db, map[string]caseTaskActionRow{"9": deletedTask}, []*models.InboxItem{deletedItem})
		require.False(t, visible["9"], "deleted case stays invisible")
	})
}
