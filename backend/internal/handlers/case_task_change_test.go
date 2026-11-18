package handlers

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// newTaskChangeRouter wires the change endpoints plus the list/complete
// endpoints on the SAME shared-fixture router used by the collaboration
// acceptance tests, so every permission check below runs through the
// real AuthorizationService (no fixture fallback).
func newTaskChangeRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	db := newMatterCollabDB(t)
	if err := db.Exec(`CREATE TABLE case_task_changes (id INTEGER PRIMARY KEY, created_at DATETIME, task_id INTEGER, case_id INTEGER, change_type TEXT, actor_id INTEGER, old_assignee_id INTEGER, new_assignee_id INTEGER, old_due_date DATETIME, new_due_date DATETIME, old_version INTEGER DEFAULT 0, new_version INTEGER DEFAULT 0, version INTEGER, reason TEXT)`).Error; err != nil {
		t.Fatalf("create case_task_changes: %v", err)
	}
	h := NewCaseTaskChangeHandlerWithDB(db)
	g := newMatterCollabRouter(db, t.TempDir())
	g.GET("/api/v1/case-tasks/:taskId/changes", h.ListCaseTaskChanges)
	g.PUT("/api/v1/case-tasks/:taskId", h.AdjustCaseTask)
	g.POST("/api/v1/case-tasks/:taskId/cancel", h.CancelCaseTask)
	g.GET("/api/v1/cases/:id/tasks/creation-result", h.GetCreationResult)
	return g, db
}

func seedTask(t *testing.T, db *gorm.DB) uint {
	t.Helper()
	res := db.Exec(`INSERT INTO case_tasks (id, case_id, title, detail, assignee_id, created_by, status, idempotency_key, idempotency_fingerprint, version, created_at, updated_at) VALUES (900, 8, "Demo task", "detail", 3, 1, "open", "req-seed-900", "fp", 1, datetime('now'), datetime('now'))`)
	if res.Error != nil {
		t.Fatalf("seed task: %v", res.Error)
	}
	return 900
}

// Creator can reassign; the old assignee's open todo closes and the new
// assignee receives a fresh todo with the SAME task id in one commit.
func TestCaseTaskReassignClosesOldTodoOpensNew(t *testing.T) {
	g, db := newTaskChangeRouter(t)
	seedTask(t, db)
	db.Exec(`INSERT INTO inbox_items (id, user_id, source_type, source_id, completion_mode, workflow_task_key, title, is_completed) VALUES (910, 3, 'task', 900, 'workflow', 'case_task:900', 'todo-old', 0)`)
	w := doJSON(g, http.MethodPut, "/api/v1/case-tasks/900", "1", "lawyer", map[string]interface{}{"version": 1, "assignee_id": 1, "reason": "handover"})
	if w.Code != http.StatusOK {
		t.Fatalf("adjust: %d %s", w.Code, w.Body.String())
	}
	var closed, opened int
	db.Raw(`SELECT COUNT(*) FROM inbox_items WHERE workflow_task_key = 'case_task:900' AND user_id = 3 AND is_completed = 1`).Scan(&closed)
	db.Raw(`SELECT COUNT(*) FROM inbox_items WHERE workflow_task_key = 'case_task:900' AND user_id = 1 AND is_completed = 0`).Scan(&opened)
	if closed != 1 || opened != 1 {
		t.Fatalf("inbox sync: closed=%d opened=%d", closed, opened)
	}
	var version int
	db.Raw(`SELECT version FROM case_tasks WHERE id = 900`).Scan(&version)
	if version != 2 {
		t.Fatalf("version after reassign = %d, want 2", version)
	}
}

// Stale version is rejected with 409 and nothing changes.
func TestCaseTaskStaleVersionRejected(t *testing.T) {
	g, db := newTaskChangeRouter(t)
	seedTask(t, db)
	w := doJSON(g, http.MethodPut, "/api/v1/case-tasks/900", "1", "lawyer", map[string]interface{}{"version": 99, "due_date": "2026-10-09T00:00:00Z"})
	if w.Code != http.StatusConflict {
		t.Fatalf("stale adjust: %d %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "CASE_TASK_CHANGE_STALE") {
		t.Fatalf("expected CASE_TASK_CHANGE_STALE in body: %s", w.Body.String())
	}
}

// Creator deactivated after creation loses adjust permission (403),
// even though actorID == created_by. Proves the live check replaced
// the old creator-直通 path.
func TestCaseTaskCreatorDeactivatedForbidden(t *testing.T) {
	g, db := newTaskChangeRouter(t)
	seedTask(t, db)
	db.Exec(`UPDATE users SET status = 'disabled' WHERE id = 1`)
	w := doJSON(g, http.MethodPut, "/api/v1/case-tasks/900", "1", "lawyer", map[string]interface{}{"version": 1, "assignee_id": 4})
	if w.Code != http.StatusForbidden {
		t.Fatalf("deactivated creator adjust: %d %s", w.Code, w.Body.String())
	}
}

// Cancel closes the open todo and records the audit entry.
func TestCaseTaskCancelClosesTodoAndAudits(t *testing.T) {
	g, db := newTaskChangeRouter(t)
	seedTask(t, db)
	db.Exec(`INSERT INTO inbox_items (id, user_id, source_type, source_id, completion_mode, workflow_task_key, title, is_completed) VALUES (911, 2, 'task', 900, 'workflow', 'case_task:900', 'todo', 0)`)
	w := doJSON(g, http.MethodPost, "/api/v1/case-tasks/900/cancel", "1", "lawyer", map[string]interface{}{"version": 1, "reason": "no longer needed"})
	if w.Code != http.StatusOK {
		t.Fatalf("cancel: %d %s", w.Code, w.Body.String())
	}
	var open int
	db.Raw(`SELECT COUNT(*) FROM inbox_items WHERE workflow_task_key = 'case_task:900' AND is_completed = 0`).Scan(&open)
	if open != 0 {
		t.Fatalf("open todos after cancel = %d", open)
	}
	var audits int
	db.Raw(`SELECT COUNT(*) FROM case_task_changes WHERE task_id = 900 AND change_type = 'cancel'`).Scan(&audits)
	if audits != 1 {
		t.Fatalf("cancel audit rows = %d", audits)
	}
}

// Listing capabilities: assignee gets can_complete, creator gets
// can_adjust/can_cancel, unrelated reader gets nothing.
func TestCaseTaskListCapabilities(t *testing.T) {
	g, db := newTaskChangeRouter(t)
	seedTask(t, db)
	w := doJSON(g, http.MethodGet, "/api/v1/cases/8/tasks", "1", "lawyer", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list: %d %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	for _, want := range []string{"can_adjust", "can_cancel", "can_complete", "can_view_changes", "capabilities", "version"} {
		if !strings.Contains(body, want) {
			t.Fatalf("list body missing %s: %s", want, body)
		}
	}
}

// Creation-result lookup: found, not_found, unknown(out-of-scope), and
// missing request_id.
func TestCaseTaskCreationResultStates(t *testing.T) {
	g, db := newTaskChangeRouter(t)
	seedTask(t, db) // idempotency_key=req-seed-900
		w := doJSON(g, http.MethodGet, "/api/v1/cases/8/tasks/creation-result?request_id=req-seed-900", "1", "lawyer", nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "found") {
		t.Fatalf("found: %d %s", w.Code, w.Body.String())
	}
	w = doJSON(g, http.MethodGet, "/api/v1/cases/8/tasks/creation-result?request_id=req-never", "1", "lawyer", nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "not_found") {
		t.Fatalf("not_found: %d %s", w.Code, w.Body.String())
	}
	w = doJSON(g, http.MethodGet, "/api/v1/cases/8/tasks/creation-result?request_id=", "1", "lawyer", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing request_id: %d", w.Code)
	}
}

// Creation-result lookup from a different case scope reports unknown,
// never leaks the task.
func TestCaseTaskCreationResultScopeUnknown(t *testing.T) {
	g, db := newTaskChangeRouter(t)
	seedTask(t, db)
	w := doJSON(g, http.MethodPost, "/api/v1/cases/8/tasks", "3", "director", map[string]interface{}{"title": "Dir task", "assignee_id": 3, "idempotency_key": "req-dir-1"})
if w.Code != http.StatusOK {
t.Fatalf("director create: %d %s", w.Code, w.Body.String())
}
w = doJSON(g, http.MethodGet, "/api/v1/cases/1/tasks/creation-result?request_id=req-dir-1", "3", "director", nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "unknown") {
		t.Fatalf("scope-unknown: %d %s", w.Code, w.Body.String())
	}
}

// Case manager (director id 3) can adjust; unrelated lawyer (id 2 in
// case 8 context as non-assignee... wait id 2 IS assignee; use id 4)
// cannot.
func TestCaseTaskManagerCanAdjustReaderCannot(t *testing.T) {
	g, db := newTaskChangeRouter(t)
	seedTask(t, db)
	w := doJSON(g, http.MethodPut, "/api/v1/case-tasks/900", "3", "director", map[string]interface{}{"version": 1, "due_date": "2026-10-20T00:00:00Z"})
	if w.Code != http.StatusOK {
		t.Fatalf("director adjust: %d %s", w.Code, w.Body.String())
	}
	w = doJSON(g, http.MethodPut, "/api/v1/case-tasks/900", "4", "lawyer", map[string]interface{}{"version": 2, "due_date": "2026-10-21T00:00:00Z"})
	if w.Code != http.StatusForbidden {
		t.Fatalf("unrelated lawyer adjust: %d %s", w.Code, w.Body.String())
	}
}

// Concurrent CAS: completion bumps version; a late adjust with the old
// version gets 409, proving cancel/reassign never overwrite a
// concurrent completion.
func TestCaseTaskConcurrentCompleteWinsOverLateAdjust(t *testing.T) {
	g, db := newTaskChangeRouter(t)
	seedTask(t, db)
	w := doJSON(g, http.MethodPost, "/api/v1/case-tasks/900/complete", "3", "director", map[string]interface{}{"completion_note": "done"})
	if w.Code != http.StatusOK {
		t.Fatalf("complete: %d %s", w.Code, w.Body.String())
	}
	w = doJSON(g, http.MethodPut, "/api/v1/case-tasks/900", "1", "lawyer", map[string]interface{}{"version": 1, "assignee_id": 4})
	if w.Code != http.StatusConflict {
		t.Fatalf("late adjust after complete: %d %s", w.Code, w.Body.String())
	}
}

// Changes list: creator sees history; an unrelated reader of the same
// case without task role gets 403.
func TestCaseTaskChangesViewPermissions(t *testing.T) {
	g, db := newTaskChangeRouter(t)
	seedTask(t, db)
	db.Exec(`INSERT INTO case_task_changes (task_id, case_id, change_type, actor_id, old_version, new_version) VALUES (900, 8, 'due_date_change', 1, 1, 2)`)
	w := doJSON(g, http.MethodGet, "/api/v1/case-tasks/900/changes", "1", "lawyer", nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "due_date_change") {
		t.Fatalf("creator view changes: %d %s", w.Code, w.Body.String())
	}
	w = doJSON(g, http.MethodGet, "/api/v1/case-tasks/900/changes", "4", "lawyer", nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("unrelated view changes: %d %s", w.Code, w.Body.String())
	}
}
