package handlers

import (
	"net/http"
	"sync"
	"testing"
)

func TestIndependentTaskCreatorLosesCaseAccessCannotCancel(t *testing.T) {
	r, db := newTaskChangeRouter(t)
	seedTask(t, db)
	if err := db.Exec("UPDATE cases SET lawyer_id = 2, created_by = '2' WHERE id = 8").Error; err != nil {
		t.Fatal(err)
	}
	w := doJSON(r, http.MethodPost, "/api/v1/case-tasks/900/cancel", "1", "lawyer", map[string]interface{}{"version": 1})
	if w.Code != http.StatusForbidden {
		t.Fatalf("revoked creator must be denied: %d %s", w.Code, w.Body.String())
	}
	var status string
	if err := db.Raw("SELECT status FROM case_tasks WHERE id = 900").Scan(&status).Error; err != nil {
		t.Fatal(err)
	}
	if status != "open" {
		t.Fatalf("denied cancel changed task: %s", status)
	}
}

func TestIndependentCaseReaderCannotAdjustOthersTask(t *testing.T) {
	r, db := newTaskChangeRouter(t)
	seedTask(t, db)
	if err := db.Exec("INSERT INTO case_ethical_wall_whitelist (case_id, user_id) VALUES (8, 4)").Error; err != nil {
		t.Fatal(err)
	}
	read := doJSON(r, http.MethodGet, "/api/v1/cases/8/tasks", "4", "lawyer", nil)
	if read.Code != http.StatusOK {
		t.Fatalf("fixture reader must genuinely read case: %d %s", read.Code, read.Body.String())
	}
	w := doJSON(r, http.MethodPut, "/api/v1/case-tasks/900", "4", "lawyer", map[string]interface{}{"version": 1, "assignee_id": 4})
	if w.Code != http.StatusForbidden {
		t.Fatalf("reader is not manager/creator: %d %s", w.Code, w.Body.String())
	}
}

// True concurrent requests, not a sequential stale-version example. A
// storage-level loser may return a fail-closed database error on SQLite;
// the terminal task, audit and all workflow todos must still agree.
func TestIndependentConcurrentTaskCancelAndCompleteStayAtomic(t *testing.T) {
	r, db := newTaskChangeRouter(t)
	seedTask(t, db)
	if err := db.Exec("PRAGMA busy_timeout = 1500").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO inbox_items (user_id, source_type, source_id, completion_mode, workflow_task_key, title, is_completed) VALUES (3, 'task', 900, 'workflow', 'case_task:900', '并发虚构待办', 0)").Error; err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	results := make(chan int, 2)
	var wg sync.WaitGroup
	for _, op := range []struct {
		path, user, role string
		body             interface{}
	}{
		{"/api/v1/case-tasks/900/cancel", "1", "lawyer", map[string]interface{}{"version": 1, "reason": "虚构事项取消"}},
		{"/api/v1/case-tasks/900/complete", "3", "director", map[string]interface{}{"completion_note": "虚构事项办妥"}},
	} {
		wg.Add(1)
		go func(op struct {
			path, user, role string
			body             interface{}
		}) { defer wg.Done(); <-start; results <- doJSON(r, http.MethodPost, op.path, op.user, op.role, op.body).Code }(op)
	}
	close(start)
	wg.Wait()
	close(results)
	ok := 0
	for code := range results {
		if code == http.StatusOK {
			ok++
		}
	}
	if ok != 1 {
		t.Fatalf("exactly one terminal operation must succeed, got %d", ok)
	}
	var task struct {
		Status         string
		Version        uint
		CompletionNote string
	}
	if err := db.Table("case_tasks").Where("id = 900").Take(&task).Error; err != nil {
		t.Fatal(err)
	}
	var openTodos, cancelAudits int64
	if err := db.Table("inbox_items").Where("workflow_task_key = ? AND is_completed = ?", "case_task:900", false).Count(&openTodos).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Table("case_task_changes").Where("task_id = ? AND change_type = ?", 900, "cancel").Count(&cancelAudits).Error; err != nil {
		t.Fatal(err)
	}
	if task.Version != 2 || openTodos != 0 {
		t.Fatalf("non-atomic terminal state: task=%+v open_todos=%d", task, openTodos)
	}
	if task.Status == "completed" {
		if cancelAudits != 0 || task.CompletionNote != "虚构事项办妥" {
			t.Fatalf("completion winner mixed with cancel: task=%+v cancel_audits=%d", task, cancelAudits)
		}
	} else if task.Status == "cancelled" {
		if cancelAudits != 1 || task.CompletionNote != "" {
			t.Fatalf("cancel winner mixed with completion: task=%+v cancel_audits=%d", task, cancelAudits)
		}
	} else {
		t.Fatalf("no valid terminal state: %+v", task)
	}
}
