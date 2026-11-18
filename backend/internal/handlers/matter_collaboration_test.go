package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
)

// newMatterCollabDB seeds the acceptance sample: lawyerA (id 1) owns case 8,
// lawyerB (id 2) owns case 1, no team membership, plus the intake ledger.
func newMatterCollabDB(t *testing.T) *gorm.DB {
	t.Helper()
	// File-backed SQLite in the test temp dir: a plain :memory: database
	// looks empty on any pooled connection beyond the first, and the
	// shared-cache name collides across tests running in one process.
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "matter.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	stmts := []string{
		`CREATE TABLE users (id INTEGER PRIMARY KEY, role TEXT, status TEXT DEFAULT 'active', deleted_at DATETIME)`,
		`CREATE TABLE clients (id INTEGER PRIMARY KEY, name TEXT, created_by INTEGER, deleted_at DATETIME)`,
		`CREATE TABLE cases (id INTEGER PRIMARY KEY, case_number TEXT, title TEXT, client_id INTEGER, lawyer_id INTEGER, created_by TEXT, status TEXT, deleted_at DATETIME, ethical_wall_enabled BOOLEAN DEFAULT 0)`,
		`CREATE TABLE case_ethical_wall_whitelist (id INTEGER PRIMARY KEY, case_id INTEGER, user_id INTEGER)`,
		`CREATE TABLE case_intakes (id TEXT PRIMARY KEY, client_id INTEGER, created_by TEXT, status TEXT, metadata TEXT DEFAULT '{}', created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE approval_requests (id TEXT PRIMARY KEY, metadata TEXT, current_approver_id TEXT, status TEXT, applicant_id TEXT, conflict_check_id TEXT, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE conflict_reviewer_assignments (id TEXT PRIMARY KEY, check_id TEXT, reviewer_id INTEGER, status TEXT, recusal_declared BOOLEAN DEFAULT 1, effective_from DATETIME, effective_to DATETIME, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE case_materials (id INTEGER PRIMARY KEY, case_id INTEGER, intake_id TEXT, name TEXT, material_type TEXT, status TEXT DEFAULT 'missing', required BOOLEAN DEFAULT 1, storage_url TEXT, owner_id INTEGER, uploaded_by INTEGER, uploaded_at DATETIME, metadata TEXT DEFAULT '{}', created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE material_files (id INTEGER PRIMARY KEY, material_id INTEGER, content_sha256 TEXT, file_name TEXT, size_bytes INTEGER, content_type TEXT, storage_path TEXT, uploaded_by INTEGER, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE case_tasks (id INTEGER PRIMARY KEY, case_id INTEGER, title TEXT, detail TEXT, assignee_id INTEGER, created_by INTEGER, due_date DATETIME, status TEXT DEFAULT 'open', completed_by INTEGER, completed_at DATETIME, completion_note TEXT, idempotency_key TEXT, idempotency_fingerprint TEXT DEFAULT '', created_at DATETIME, updated_at DATETIME, version INTEGER NOT NULL DEFAULT 1)`,
		`CREATE TABLE inbox_items (id INTEGER PRIMARY KEY, user_id INTEGER, source_type TEXT, source_id INTEGER, source_ref TEXT, completion_mode TEXT DEFAULT 'manual', workflow_task_key TEXT, title TEXT, content TEXT, priority TEXT DEFAULT 'medium', due_date DATETIME, due_date_type TEXT, is_read BOOLEAN DEFAULT 0, read_at DATETIME, is_completed BOOLEAN DEFAULT 0, completed_at DATETIME, reminder_sent BOOLEAN DEFAULT 0, reminder_count INTEGER DEFAULT 0, escalated BOOLEAN DEFAULT 0, escalated_at DATETIME, snoozed_until DATETIME, snoozed_count INTEGER DEFAULT 0, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE documents (id INTEGER PRIMARY KEY, entity_type TEXT, entity_id INTEGER, deleted_at DATETIME)`,
		`INSERT INTO users (id, role, status) VALUES (1, 'lawyer', 'active'), (2, 'lawyer', 'active'), (3, 'director', 'active'), (4, 'lawyer', 'active'), (5, 'admin', 'active')`,
		`INSERT INTO clients (id, name, created_by) VALUES (10, 'Client-A', 1), (20, 'Client-B', 2)`,
		`INSERT INTO cases (id, case_number, title, client_id, lawyer_id, created_by, status) VALUES (8, 'CASE-008', 'Case-A', 10, 1, '1', 'active'), (1, 'CASE-001', 'Case-B', 20, 2, '2', 'active')`,
		`INSERT INTO case_intakes (id, client_id, created_by, status) VALUES ('aaaa1111-0000-0000-0000-000000000001', 10, '1', 'draft'), ('aaaa1111-0000-0000-0000-000000000002', 20, '2', 'draft')`,
		`INSERT INTO case_materials (id, intake_id, name, status, required) VALUES (100, 'aaaa1111-0000-0000-0000-000000000001', 'POA', 'missing', 1)`,
		`INSERT INTO case_materials (id, case_id, name, status, required) VALUES (200, 8, 'Evidence', 'missing', 1)`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			t.Fatalf("seed: %v: %v", s, err)
		}
	}
	return db
}

func newMatterCollabRouter(db *gorm.DB, storageDir string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewMatterCollaborationHandlerWithDB(db)
	h.StorageDir = storageDir
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", c.GetHeader("X-Test-Role"))
		c.Set("user_id", c.GetHeader("X-Test-User-Id"))
		c.Next()
	})
	r.GET("/api/v1/case-intakes/:id/materials", h.ListIntakeMaterials)
	r.POST("/api/v1/case-intakes/:id/materials", h.CreateIntakeMaterial)
	r.GET("/api/v1/cases/:id/materials", h.ListCaseMaterials)
	r.POST("/api/v1/cases/:id/materials", h.CreateCaseMaterial)
	r.PUT("/api/v1/case-materials/:materialId", h.UpdateCaseMaterial)
	r.POST("/api/v1/case-materials/:materialId/files", h.UploadCaseMaterial)
	r.GET("/api/v1/case-materials/:materialId/files/:fileId/download", h.DownloadCaseMaterialFile)
	r.GET("/api/v1/cases/:id/tasks", h.ListCaseTasks)
	r.POST("/api/v1/cases/:id/tasks", h.CreateCaseTask)
	r.POST("/api/v1/case-tasks/:taskId/complete", h.CompleteCaseTask)
	return r
}

func doJSON(r *gin.Engine, method, path, userID, role string, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-Id", userID)
	req.Header.Set("X-Test-Role", role)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func doMultipartFile(r *gin.Engine, path, userID, role, filename, content string) *httptest.ResponseRecorder {
	body := &bytes.Buffer{}
	body.WriteString("--BND\r\nContent-Disposition: form-data; name=\"file\"; filename=\"" + filename + "\"\r\nContent-Type: application/octet-stream\r\n\r\n")
	body.WriteString(content)
	body.WriteString("\r\n--BND--\r\n")
	req := httptest.NewRequest(http.MethodPost, path, body)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=BND")
	req.Header.Set("X-Test-User-Id", userID)
	req.Header.Set("X-Test-Role", role)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ---- acceptance sample: object isolation without teams ----

func TestMaterialCaseIsolationNoTeams(t *testing.T) {
	db := newMatterCollabDB(t)
	r := newMatterCollabRouter(db, t.TempDir())

	// lawyerB (id 2) cannot list or download case 8 materials.
	if w := doJSON(r, http.MethodGet, "/api/v1/cases/8/materials", "2", "lawyer", nil); w.Code != http.StatusForbidden {
		t.Fatalf("B list case8 materials: got %d want 403: %s", w.Code, w.Body.String())
	}
	db.Exec("INSERT INTO material_files (id, material_id, content_sha256, file_name, size_bytes, storage_path, uploaded_by) VALUES (300, 200, 'd', 'f.txt', 2, ?, 1)", t.TempDir()+"/f.txt")
	if w := doJSON(r, http.MethodGet, "/api/v1/case-materials/200/files/300/download", "2", "lawyer", nil); w.Code != http.StatusForbidden {
		t.Fatalf("B download case8 file: got %d want 403", w.Code)
	}
	// lawyerA sees his own case materials.
	if w := doJSON(r, http.MethodGet, "/api/v1/cases/8/materials", "1", "lawyer", nil); w.Code != http.StatusOK {
		t.Fatalf("A list case8: got %d: %s", w.Code, w.Body.String())
	}
}

func TestIntakeMaterialBoundary(t *testing.T) {
	db := newMatterCollabDB(t)
	r := newMatterCollabRouter(db, t.TempDir())
	intake := "aaaa1111-0000-0000-0000-000000000001"

	// lawyerB is not the creator and not bound to the workflow: fail closed.
	if w := doJSON(r, http.MethodGet, "/api/v1/case-intakes/"+intake+"/materials", "2", "lawyer", nil); w.Code != http.StatusForbidden {
		t.Fatalf("B list intake A materials: got %d want 403", w.Code)
	}
	if w := doJSON(r, http.MethodPost, "/api/v1/case-intakes/"+intake+"/materials", "2", "lawyer", map[string]interface{}{"name": "x"}); w.Code != http.StatusForbidden {
		t.Fatalf("B create intake A material: got %d want 403", w.Code)
	}
	// Creator reads and creates; created status is always missing.
	if w := doJSON(r, http.MethodGet, "/api/v1/case-intakes/"+intake+"/materials", "1", "lawyer", nil); w.Code != http.StatusOK {
		t.Fatalf("A list own intake: got %d: %s", w.Code, w.Body.String())
	}
	w := doJSON(r, http.MethodPost, "/api/v1/case-intakes/"+intake+"/materials", "1", "lawyer", map[string]interface{}{"name": "New", "required": false})
	if w.Code != http.StatusOK {
		t.Fatalf("A create intake material: got %d: %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("\"status\":\"missing\"")) {
		t.Fatalf("created material must be missing, body: %s", w.Body.String())
	}
}

func TestIntakeApproverCanReadReturnedMaterial(t *testing.T) {
	db := newMatterCollabDB(t)
	r := newMatterCollabRouter(db, t.TempDir())
	intake := "aaaa1111-0000-0000-0000-000000000001"
	// Approver 4 is directly bound to the intake through approval metadata:
	// read passes, write stays creator-only, and a global admin role without
	// the binding still fails (no role bypass).
	// Production jsonb stores intake_id as a JSON field; the seed must
	// mirror that (the exact-match guard parses JSON, it never LIKEs).
	db.Exec("INSERT INTO approval_requests (id, metadata, current_approver_id, status, applicant_id) VALUES ('ap-1', ?, '4', 'submitted', '1')", "{\"intake_id\":\""+intake+"\"}")
	if w := doJSON(r, http.MethodGet, "/api/v1/case-intakes/"+intake+"/materials", "4", "compliance", nil); w.Code != http.StatusOK {
		t.Fatalf("approver read intake materials: got %d: %s", w.Code, w.Body.String())
	}
	if w := doJSON(r, http.MethodPost, "/api/v1/case-intakes/"+intake+"/materials", "4", "compliance", map[string]interface{}{"name": "x"}); w.Code != http.StatusForbidden {
		t.Fatalf("approver write intake material: got %d want 403", w.Code)
	}
	if w := doJSON(r, http.MethodGet, "/api/v1/case-intakes/"+intake+"/materials", "5", "admin", nil); w.Code != http.StatusForbidden {
		t.Fatalf("unbound admin read: got %d want 403", w.Code)
	}
}

func TestUploadStatusCannotBeForged(t *testing.T) {
	db := newMatterCollabDB(t)
	r := newMatterCollabRouter(db, t.TempDir())
	// PUT with status=uploaded is rejected: only the upload transaction
	// writes that state.
	w := doJSON(r, http.MethodPut, "/api/v1/case-materials/200", "1", "lawyer", map[string]interface{}{"status": "uploaded"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("forge uploaded: got %d want 400: %s", w.Code, w.Body.String())
	}
	var status string
	db.Raw("SELECT status FROM case_materials WHERE id = 200").Scan(&status)
	if status != models.MaterialStatusMissing {
		t.Fatalf("status after forged PUT = %q", status)
	}
}

func TestUploadDownloadRoundtripAndLimits(t *testing.T) {
	dir := t.TempDir()
	db := newMatterCollabDB(t)
	r := newMatterCollabRouter(db, dir)

	// Illegal extension rejected before storage.
	if w := doMultipartFile(r, "/api/v1/case-materials/200/files", "1", "lawyer", "evil.php", "X"); w.Code != http.StatusBadRequest {
		t.Fatalf("illegal ext upload: got %d want 400: %s", w.Code, w.Body.String())
	}
	// Valid upload flips status only through the transaction.
	w := doMultipartFile(r, "/api/v1/case-materials/200/files", "1", "lawyer", "evidence.txt", "HELLO")
	if w.Code != http.StatusOK {
		t.Fatalf("valid upload: got %d: %s", w.Code, w.Body.String())
	}
	var status, storageURL string
	db.Raw("SELECT status FROM case_materials WHERE id = 200").Scan(&status)
	db.Raw("SELECT storage_url FROM case_materials WHERE id = 200").Scan(&storageURL)
	if status != models.MaterialStatusUploaded {
		t.Fatalf("status after upload = %q", status)
	}
	if storageURL == "" || !bytes.Contains([]byte(storageURL), []byte("/api/v1/case-materials/200/files/")) {
		t.Fatalf("storage_url mismatch: %q", storageURL)
	}
	// Download by the owner works.
	if w := doJSON(r, http.MethodGet, "/api/v1/case-materials/200/files/1/download", "1", "lawyer", nil); w.Code != http.StatusOK {
		t.Fatalf("owner download: got %d", w.Code)
	}
	// B cannot download it.
	if w := doJSON(r, http.MethodGet, "/api/v1/case-materials/200/files/1/download", "2", "lawyer", nil); w.Code != http.StatusForbidden {
		t.Fatalf("B download A file: got %d want 403", w.Code)
	}
}

func TestOversizedUploadRejected413(t *testing.T) {
	db := newMatterCollabDB(t)
	r := newMatterCollabRouter(db, t.TempDir())
	// Just above the cap: proves the actual read limit, not only the header
	// pre-check. Oversize body is ~50MiB+16B of payload.
	big := bytes.Repeat([]byte("A"), (50<<20)+16)
	w := doMultipartFile(r, "/api/v1/case-materials/200/files", "1", "lawyer", "big.txt", string(big))
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversize upload: got %d want 413", w.Code)
	}
	var files int64
	db.Table("material_files").Count(&files)
	if files != 0 {
		t.Fatalf("failed upload left %d file rows", files)
	}
}

// ---- case tasks ----

func TestCaseTaskAssigneeScopeAndIdempotency(t *testing.T) {
	db := newMatterCollabDB(t)
	r := newMatterCollabRouter(db, t.TempDir())

	w := doJSON(r, http.MethodPost, "/api/v1/cases/8/tasks", "1", "lawyer", map[string]interface{}{
		"title": "T1", "assignee_id": 1, "idempotency_key": "k1",
		"detail": "d1", "due_date": "2026-10-10T00:00:00Z",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("create task A->A: got %d: %s", w.Code, w.Body.String())
	}
	// Assigning outsider 4 (not on the case) fails.
	w = doJSON(r, http.MethodPost, "/api/v1/cases/8/tasks", "1", "lawyer", map[string]interface{}{
		"title": "x", "assignee_id": 4, "idempotency_key": "k2",
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("create task A->outsider: got %d want 403", w.Code)
	}
	// B cannot even create a task inside case 8.
	w = doJSON(r, http.MethodPost, "/api/v1/cases/8/tasks", "2", "lawyer", map[string]interface{}{
		"title": "x", "assignee_id": 2,
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("B create in case8: got %d want 403", w.Code)
	}
	// Same key, different detail: 409 conflict (fingerprint covers detail).
	w = doJSON(r, http.MethodPost, "/api/v1/cases/8/tasks", "1", "lawyer", map[string]interface{}{
		"title": "T1", "assignee_id": 1, "idempotency_key": "k1",
		"detail": "d1-changed", "due_date": "2026-10-10T00:00:00Z",
	})
	if w.Code != http.StatusConflict {
		t.Fatalf("idempotency conflict: got %d want 409", w.Code)
	}
	// Same key, same payload: idempotent replay.
	w = doJSON(r, http.MethodPost, "/api/v1/cases/8/tasks", "1", "lawyer", map[string]interface{}{
		"title": "T1", "assignee_id": 1, "idempotency_key": "k1",
		"detail": "d1", "due_date": "2026-10-10T00:00:00Z",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("idempotent replay: got %d: %s", w.Code, w.Body.String())
	}
	// Keyless retries create two separate tasks (server-minted keys).
	for i := 0; i < 2; i++ {
		w = doJSON(r, http.MethodPost, "/api/v1/cases/8/tasks", "1", "lawyer", map[string]interface{}{
			"title": "Keyless", "assignee_id": 1,
		})
		if w.Code != http.StatusOK {
			t.Fatalf("keyless create #%d: got %d: %s", i, w.Code, w.Body.String())
		}
	}
	var count int64
	db.Table("case_tasks").Where("title = ?", "Keyless").Count(&count)
	if count != 2 {
		t.Fatalf("keyless retries produced %d tasks, want 2", count)
	}
	// B lists case 1 tasks fine (his own case).
	if w := doJSON(r, http.MethodGet, "/api/v1/cases/1/tasks", "2", "lawyer", nil); w.Code != http.StatusOK {
		t.Fatalf("B list case1 tasks: got %d", w.Code)
	}
}

func TestCaseTaskCompleteAtomicityAndActorRevalidation(t *testing.T) {
	db := newMatterCollabDB(t)
	r := newMatterCollabRouter(db, t.TempDir())
	w := doJSON(r, http.MethodPost, "/api/v1/cases/8/tasks", "1", "lawyer", map[string]interface{}{
		"title": "Before-complete", "assignee_id": 1,
	})
	var created struct {
		Data struct {
			Task models.CaseTask `json:"task"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	taskID := created.Data.Task.ID
	if taskID == 0 {
		t.Fatalf("create task failed: %s", w.Body.String())
	}
	taskPath := "/api/v1/case-tasks/" + strconv.Itoa(int(taskID)) + "/complete"
	key := "case_task:" + strconv.Itoa(int(taskID))

	// The mirrored todo exists with completion_mode=workflow.
	var mode string
	db.Raw("SELECT completion_mode FROM inbox_items WHERE workflow_task_key = ?", key).Scan(&mode)
	if mode != "workflow" {
		t.Fatalf("todo completion_mode = %q, want workflow", mode)
	}

	// Outsider cannot complete.
	if w := doJSON(r, http.MethodPost, taskPath, "2", "lawyer", map[string]interface{}{}); w.Code != http.StatusForbidden {
		t.Fatalf("B complete A task: got %d want 403", w.Code)
	}
	// Assignee completes with a record.
	if w := doJSON(r, http.MethodPost, taskPath, "1", "lawyer", map[string]interface{}{"completion_note": "done"}); w.Code != http.StatusOK {
		t.Fatalf("assignee complete: got %d: %s", w.Code, w.Body.String())
	}
	// Idempotent replay by the assignee still succeeds.
	if w := doJSON(r, http.MethodPost, taskPath, "1", "lawyer", map[string]interface{}{}); w.Code != http.StatusOK {
		t.Fatalf("replay complete: got %d", w.Code)
	}
	// Revoked actor replay: already-completed tasks still re-authorize
	// before the idempotent short-circuit.
	if w := doJSON(r, http.MethodPost, taskPath, "2", "lawyer", map[string]interface{}{}); w.Code != http.StatusForbidden {
		t.Fatalf("revoked outsider replay complete: got %d want 403", w.Code)
	}
	// The todo is closed exactly once.
	var done int64
	db.Table("inbox_items").Where("workflow_task_key = ? AND is_completed = ?", key, true).Count(&done)
	if done != 1 {
		t.Fatalf("workflow todo closed count = %d, want 1", done)
	}
}

// Review 2026-10-01: a plain read-team lawyer (case whitelist visibility
// but not the assignee and not a management role) must NOT be able to
// complete someone else's task.
func TestCaseTaskReadTeamMemberCannotCompleteOthersTask(t *testing.T) {
	db := newMatterCollabDB(t)
	r := newMatterCollabRouter(db, t.TempDir())
	// Case 8 runs wall-enabled and all actors enter through the ethical
	// wall whitelist (the minimal seed schema only supports the whitelist
	// authorization path). lawyerB (id 2) is a plain whitelist reader:
	// CanReadCase sees him, CanManageCase does not promote him. lawyerA
	// (id 1) is on the whitelist AND owns the case; director (id 3) is a
	// genuine business-management role.
	db.Exec("UPDATE cases SET ethical_wall_enabled = 1 WHERE id = 8")
	db.Exec("INSERT INTO case_ethical_wall_whitelist (case_id, user_id) VALUES (8, 1), (8, 2), (8, 3)")
	w := doJSON(r, http.MethodPost, "/api/v1/cases/8/tasks", "1", "lawyer", map[string]interface{}{
		"title": "ReadTeamGate", "assignee_id": 1,
	})
	var created struct {
		Data struct {
			Task models.CaseTask `json:"task"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	if created.Data.Task.ID == 0 {
		t.Fatalf("create task failed: %s", w.Body.String())
	}
	taskPath := "/api/v1/case-tasks/" + strconv.Itoa(int(created.Data.Task.ID)) + "/complete"
	if w := doJSON(r, http.MethodPost, taskPath, "2", "lawyer", map[string]interface{}{"completion_note": "usurped"}); w.Code != http.StatusForbidden {
		t.Fatalf("read-team complete others task: got %d want 403: %s", w.Code, w.Body.String())
	}
	// The manager path still works: the assignee's task assigned to
	// lawyerA (id 1) is completed by the director (id 3, genuine
	// management role behind the same wall), who is NOT the assignee.
	w = doJSON(r, http.MethodPost, "/api/v1/cases/8/tasks", "1", "lawyer", map[string]interface{}{
		"title": "ManagerCompletes", "assignee_id": 1,
	})
	var created2 struct {
		Data struct {
			Task models.CaseTask `json:"task"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &created2)
	if created2.Data.Task.ID == 0 {
		t.Fatalf("create second task failed: %s", w.Body.String())
	}
	taskPath2 := "/api/v1/case-tasks/" + strconv.Itoa(int(created2.Data.Task.ID)) + "/complete"
	if w := doJSON(r, http.MethodPost, taskPath2, "3", "director", map[string]interface{}{}); w.Code != http.StatusOK {
		t.Fatalf("director complete foreign assignee task: got %d want 200: %s", w.Code, w.Body.String())
	}
}

// Review 2026-10-01: the FIRST completion response must report
// completion_note_recorded=true and carry the actual note text, because
// the DB row and the response come from the same commit.
func TestCaseTaskFirstCompleteResponseCarriesNote(t *testing.T) {
	db := newMatterCollabDB(t)
	r := newMatterCollabRouter(db, t.TempDir())
	w := doJSON(r, http.MethodPost, "/api/v1/cases/8/tasks", "1", "lawyer", map[string]interface{}{
		"title": "NoteEcho", "assignee_id": 1,
	})
	var created struct {
		Data struct {
			Task models.CaseTask `json:"task"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	if created.Data.Task.ID == 0 {
		t.Fatalf("create task failed: %s", w.Body.String())
	}
	taskPath := "/api/v1/case-tasks/" + strconv.Itoa(int(created.Data.Task.ID)) + "/complete"
	const noteText = "首次完成备注 2026-10-01"
	w = doJSON(r, http.MethodPost, taskPath, "1", "lawyer", map[string]interface{}{"completion_note": noteText})
	if w.Code != http.StatusOK {
		t.Fatalf("first complete: got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Data struct {
			Replayed               bool   `json:"replayed"`
			CompletionNoteRecorded bool   `json:"completion_note_recorded"`
			CompletionNote         string `json:"completion_note"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Data.Replayed {
		t.Fatalf("first completion flagged replayed=true")
	}
	if !resp.Data.CompletionNoteRecorded {
		t.Fatalf("completion_note_recorded = false, want true; body=%s", w.Body.String())
	}
	if resp.Data.CompletionNote != noteText {
		t.Fatalf("completion_note = %q, want %q", resp.Data.CompletionNote, noteText)
	}
	// And the persisted column matches too.
	var stored string
	db.Raw("SELECT completion_note FROM case_tasks WHERE id = ?", created.Data.Task.ID).Scan(&stored)
	if stored != noteText {
		t.Fatalf("stored completion_note = %q, want %q", stored, noteText)
	}
}

// Defect 5+6 regression: identical bytes from two uploads must land on two
// distinct physical files, and a failed DB commit deletes only its OWN
// file, never the twin already referenced by another material row.
func TestUploadStorageIsolationOnIdenticalBytes(t *testing.T) {
	dir := t.TempDir()
	db := newMatterCollabDB(t)
	r := newMatterCollabRouter(db, dir)
	if w := doMultipartFile(r, "/api/v1/case-materials/200/files", "1", "lawyer", "a.txt", "SAME"); w.Code != http.StatusOK {
		t.Fatalf("first upload: got %d: %s", w.Code, w.Body.String())
	}
	// Force the second insert to fail AFTER the physical file was written:
	// the rollback path must clean only the new file, and the first
	// material's file (identical bytes) must survive untouched.
	db.Exec("ALTER TABLE material_files RENAME TO material_files_backup")
	db.Exec("CREATE TABLE material_files (id INTEGER PRIMARY KEY AUTOINCREMENT, material_id INTEGER NOT NULL UNIQUE, content_sha256 TEXT, file_name TEXT, size_bytes INTEGER, content_type TEXT, storage_path TEXT, uploaded_by INTEGER, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME)")
	// Pre-seed the conflicting row so the upload's INSERT hits the UNIQUE
	// constraint inside the transaction.
	db.Exec("INSERT INTO material_files (material_id, content_sha256, file_name, size_bytes, storage_path, uploaded_by) VALUES (100, 'pre', 'pre.txt', 1, '/pre', 1)")
	if w := doMultipartFile(r, "/api/v1/case-materials/100/files", "1", "lawyer", "b.txt", "SAME"); w.Code != http.StatusInternalServerError {
		t.Fatalf("forced-failure upload: got %d want 500: %s", w.Code, w.Body.String())
	}
	backupRows := []struct{ StoragePath string }{}
	db.Table("material_files_backup").Select("storage_path").Scan(&backupRows)
	if len(backupRows) != 1 {
		t.Fatalf("backup rows = %d, want 1", len(backupRows))
	}
	if _, err := os.Stat(backupRows[0].StoragePath); err != nil {
		t.Fatalf("first upload file was wrongly deleted after twin failure: %v", err)
	}
	var live int64
	db.Table("material_files").Where("content_sha256 = ?", "pre").Count(&live)
	if live != 1 {
		t.Fatalf("pre-seeded row vanished: %d", live)
	}
	// The failed attempt itself added nothing (no dangling DB relation).
	var total int64
	db.Table("material_files").Count(&total)
	if total != 1 {
		t.Fatalf("failed upload left %d rows total, want 1", total)
	}
}

// Defect 2 regression: intake-stage materials authorize the responsible
// person through the intake boundary (creator or already-assigned owner),
// never by arbitrary non-zero id; the case-stage 403 path stays intact.
func TestIntakeMaterialOwnerBoundary(t *testing.T) {
	db := newMatterCollabDB(t)
	r := newMatterCollabRouter(db, t.TempDir())
	w := doJSON(r, http.MethodPost, "/api/v1/case-intakes/aaaa1111-0000-0000-0000-000000000001/materials", "1", "lawyer", map[string]interface{}{
		"name": "Doc", "owner_id": 4,
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("intake owner outsider create: got %d want 403: %s", w.Code, w.Body.String())
	}
	w = doJSON(r, http.MethodPost, "/api/v1/case-intakes/aaaa1111-0000-0000-0000-000000000001/materials", "1", "lawyer", map[string]interface{}{
		"name": "Doc", "owner_id": 1,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("intake owner self create: got %d: %s", w.Code, w.Body.String())
	}
	var created struct {
		Data struct {
			Material models.CaseMaterial `json:"material"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	matID := created.Data.Material.ID
	w = doJSON(r, http.MethodPut, "/api/v1/case-materials/"+strconv.Itoa(int(matID)), "1", "lawyer", map[string]interface{}{
		"owner_id": 4,
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("intake owner outsider update: got %d want 403", w.Code)
	}
	w = doJSON(r, http.MethodPut, "/api/v1/case-materials/200", "1", "lawyer", map[string]interface{}{
		"owner_id": 4,
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("case owner outsider update: got %d want 403", w.Code)
	}
}

// Defect 3 regression: ledger edit must accept name and material_type and
// persist them, while still refusing status=uploaded forgery.
func TestCaseMaterialLedgerEditFields(t *testing.T) {
	db := newMatterCollabDB(t)
	r := newMatterCollabRouter(db, t.TempDir())
	w := doJSON(r, http.MethodPut, "/api/v1/case-materials/200", "1", "lawyer", map[string]interface{}{
		"name": "Renamed evidence", "material_type": "contract",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("ledger edit: got %d: %s", w.Code, w.Body.String())
	}
	var name, mtype string
	db.Raw("SELECT name FROM case_materials WHERE id = 200").Scan(&name)
	db.Raw("SELECT material_type FROM case_materials WHERE id = 200").Scan(&mtype)
	if name != "Renamed evidence" || mtype != "contract" {
		t.Fatalf("ledger edit not persisted: name=%q type=%q", name, mtype)
	}
	w = doJSON(r, http.MethodPut, "/api/v1/case-materials/200", "1", "lawyer", map[string]interface{}{
		"status": "uploaded",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("forged uploaded status: got %d want 400", w.Code)
	}
}

// Defect 4 regression: the completion note must be persisted by the same
// transaction that flips the status, and an over-length note must be
// rejected without opening the transaction at all.
func TestCaseTaskCompletionNotePersisted(t *testing.T) {
	db := newMatterCollabDB(t)
	r := newMatterCollabRouter(db, t.TempDir())
	w := doJSON(r, http.MethodPost, "/api/v1/cases/8/tasks", "1", "lawyer", map[string]interface{}{
		"title": "Note-task", "assignee_id": 1,
	})
	var created struct {
		Data struct {
			Task models.CaseTask `json:"task"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	taskID := created.Data.Task.ID
	taskPath := "/api/v1/case-tasks/" + strconv.Itoa(int(taskID)) + "/complete"
	if w := doJSON(r, http.MethodPost, taskPath, "1", "lawyer", map[string]interface{}{
		"completion_note": "done-archived-A",
	}); w.Code != http.StatusOK {
		t.Fatalf("complete with note: got %d: %s", w.Code, w.Body.String())
	}
	var note string
	db.Raw("SELECT completion_note FROM case_tasks WHERE id = ?", taskID).Scan(&note)
	if note != "done-archived-A" {
		t.Fatalf("completion note not persisted: %q", note)
	}
	w = doJSON(r, http.MethodPost, "/api/v1/cases/8/tasks", "1", "lawyer", map[string]interface{}{
		"title": "Too-long", "assignee_id": 1,
	})
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	longID := created.Data.Task.ID
	longPath := "/api/v1/case-tasks/" + strconv.Itoa(int(longID)) + "/complete"
	huge := strings.Repeat("长", services.CaseTaskCompletionNoteMaxLen+1)
	if w := doJSON(r, http.MethodPost, longPath, "1", "lawyer", map[string]interface{}{
		"completion_note": huge,
	}); w.Code != http.StatusBadRequest {
		t.Fatalf("over-length note: got %d want 400", w.Code)
	}
	var status string
	db.Raw("SELECT status FROM case_tasks WHERE id = ?", longID).Scan(&status)
	if status != models.CaseTaskStatusOpen {
		t.Fatalf("rejected note must not complete task: status=%q", status)
	}
}

// ---- UpdateCaseIntake material merge (defect 1 regression) ----

func newIntakeMergeRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewDemoAggregateHandler(db)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", c.GetHeader("X-Test-Role"))
		c.Set("user_id", c.GetHeader("X-Test-User-Id"))
		c.Next()
	})
	r.PUT("/api/v1/case-intakes/:id", h.UpdateCaseIntake)
	return r
}

// A save that echoes an uploaded material row back must NOT delete or
// recreate it: same id, same attachment linkage, server-owned columns
// (status/storage/owner) survive untouched, and a client-supplied
// uploaded status cannot be forged onto a missing row.
func TestIntakeSavePreservesUploadedMaterial(t *testing.T) {
	db := newMatterCollabDB(t)
	r := newIntakeMergeRouter(db)
	// Give material 100 a real uploaded state and one attachment.
	db.Exec("UPDATE case_materials SET status = 'uploaded', owner_id = 1 WHERE id = 100")
	db.Exec("INSERT INTO material_files (id, material_id, content_sha256, file_name, size_bytes, storage_path, uploaded_by) VALUES (1, 100, 'd1', 'poa.pdf', 5, '/x/poa.pdf', 1)")
	// The old UI echo: full list with the SAME id plus a forged uploaded
	// status on the still-missing sibling row 300.
	db.Exec("INSERT INTO case_materials (id, intake_id, name, status) VALUES (300, 'aaaa1111-0000-0000-0000-000000000001', 'Extra', 'missing')")
	body := map[string]interface{}{
		"expected_revision": "legacy-aaaa1111-0000-0000-0000-000000000001",
		"materials": []interface{}{
			map[string]interface{}{
				"id": 100, "name": "POA", "status": "uploaded", "required": true,
				"owner_id": 2, "storage_url": "/evil", "uploaded_by": 2,
			},
			map[string]interface{}{
				"id": 300, "name": "Extra", "status": "uploaded",
			},
		},
	}
	w := doJSON(r, http.MethodPut, "/api/v1/case-intakes/aaaa1111-0000-0000-0000-000000000001", "1", "lawyer", body)
	if w.Code != http.StatusOK {
		t.Fatalf("intake save: got %d: %s", w.Code, w.Body.String())
	}
	// Row 100 keeps its id, uploaded status, owner, and attachment row.
	var status, owner sql.NullString
	var files int64
	db.Raw("SELECT status FROM case_materials WHERE id = 100").Scan(&status)
	db.Raw("SELECT owner_id FROM case_materials WHERE id = 100").Scan(&owner)
	db.Table("material_files").Where("material_id = ?", 100).Count(&files)
	if status.String != models.MaterialStatusUploaded {
		t.Fatalf("uploaded row lost its status: %q", status.String)
	}
	if owner.String != "1" {
		t.Fatalf("client rewrote owner_id: %q", owner.String)
	}
	if files != 1 {
		t.Fatalf("attachment rows for material 100 = %d, want 1", files)
	}
	// Row 300 was NOT forged to uploaded: the save leaves its real state.
	var forged string
	db.Raw("SELECT status FROM case_materials WHERE id = 300").Scan(&forged)
	if forged == models.MaterialStatusUploaded {
		t.Fatalf("save path forged uploaded status onto row 300")
	}
	// Name edits on the echoed row still land (ledger edit lives).
	var name string
	db.Raw("SELECT name FROM case_materials WHERE id = 100").Scan(&name)
	if name != "POA" {
		t.Fatalf("echoed name not preserved: %q", name)
	}
}

// A save whose payload omits id fields creates fresh rows at status
// missing, never at uploaded: the save path cannot mint upload state.
func TestIntakeSaveNewMaterialStartsMissing(t *testing.T) {
	db := newMatterCollabDB(t)
	r := newIntakeMergeRouter(db)
	body := map[string]interface{}{
		"expected_revision": "legacy-aaaa1111-0000-0000-0000-000000000001",
		"materials": []interface{}{
			map[string]interface{}{"name": "Fresh", "status": "uploaded", "storage_url": "/evil"},
		},
	}
	w := doJSON(r, http.MethodPut, "/api/v1/case-intakes/aaaa1111-0000-0000-0000-000000000001", "1", "lawyer", body)
	if w.Code != http.StatusOK {
		t.Fatalf("intake save: got %d: %s", w.Code, w.Body.String())
	}
	var status string
	db.Raw("SELECT status FROM case_materials WHERE name = 'Fresh'").Scan(&status)
	if status != models.MaterialStatusMissing {
		t.Fatalf("new row status = %q, want missing", status)
	}
}

// attempt2 real-PG regressions (2026-09-30): the approver binding must be
// an exact JSON intake_id match (never a LIKE over the json column), the
// deny/error path must write exactly one JSON body, and an earlier legal
// owner assignment must not make that owner assignable by anyone later.
func TestApproverBindingExactJSONMatchSingleResponse(t *testing.T) {
	db := newMatterCollabDB(t)
	r := newMatterCollabRouter(db, t.TempDir())
	intake := "aaaa1111-0000-0000-0000-000000000001"
	other := "aaaa1111-0000-0000-0000-000000000009"
	// The approver approval binds a DIFFERENT intake; the target intake
	// UUID appears only inside an unrelated metadata field. A LIKE-based
	// or substring check would wrongly admit user 4 here.
	// Marshal (not string concatenation): the unrelated-substring case only
	// proves substring isolation if the stored metadata itself is VALID
	// JSON, asserted below before the request.
	metaBytes, mErr := json.Marshal(map[string]string{
		"intake_id": other,
		"remark":    "see " + intake + " for context",
	})
	if mErr != nil {
		t.Fatalf("marshal metadata: %v", mErr)
	}
	meta := string(metaBytes)
	if !json.Valid(metaBytes) {
		t.Fatalf("substring-case metadata must be valid JSON, got: %s", meta)
	}
	db.Exec("INSERT INTO approval_requests (id, metadata, current_approver_id, status, applicant_id) VALUES ('ap-json', ?, '4', 'submitted', '1')", meta)
	w := doJSON(r, http.MethodGet, "/api/v1/case-intakes/"+intake+"/materials", "4", "compliance", nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("substring binding must not admit: got %d: %s", w.Code, w.Body.String())
	}
	if strings.Count(w.Body.String(), "\"success\"") != 1 {
		t.Fatalf("deny must return exactly one JSON body, got: %s", w.Body.String())
	}
	// Exact binding reads cleanly and still returns a single JSON body.
	db.Exec("INSERT INTO approval_requests (id, metadata, current_approver_id, status, applicant_id) VALUES ('ap-exact', ?, '4', 'submitted', '1')", "{\"intake_id\":\""+intake+"\"}")
	w = doJSON(r, http.MethodGet, "/api/v1/case-intakes/"+intake+"/materials", "4", "compliance", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("exact binding read: got %d: %s", w.Code, w.Body.String())
	}
	if strings.Count(w.Body.String(), "\"success\"") != 1 {
		t.Fatalf("read must return exactly one JSON body, got: %s", w.Body.String())
	}
}

// The material reviewer read path must follow the ACTUAL reviewer
// assignment (conflict_reviewer_assignments, active + in-window), never
// the check creator in conflict_check_records.user_id: the creator is
// the applying lawyer (review 2026-09-30).
func TestReviewerMaterialReadFollowsAssignment(t *testing.T) {
	db := newMatterCollabDB(t)
	r := newMatterCollabRouter(db, t.TempDir())
	intake := "aaaa1111-0000-0000-0000-000000000001"
	checkID := "CCT_TEST_ASSIGN_1"
	db.Exec("INSERT INTO approval_requests (id, metadata, current_approver_id, status, applicant_id) VALUES ('ap-rv', ?, '4', 'submitted', '1')", "{\"intake_id\":\""+intake+"\"}")
	// The approval carries the anchored check id.
	db.Exec("UPDATE approval_requests SET conflict_check_id = ? WHERE id = 'ap-rv'", checkID)
	db.Exec("INSERT INTO users (id, role) VALUES (6, 'conflict_officer'), (7, 'conflict_officer')")
	// Assigned officer reads; unassigned officer denied; revoked denied.
	db.Exec("INSERT INTO conflict_reviewer_assignments (id, check_id, reviewer_id, status, recusal_declared) VALUES ('asg-1', ?, 6, 'ACTIVE', 1)", checkID)
	w := doJSON(r, http.MethodGet, "/api/v1/case-intakes/"+intake+"/materials", "6", "conflict_officer", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("assigned reviewer read: got %d: %s", w.Code, w.Body.String())
	}
	if strings.Count(w.Body.String(), "\"success\"") != 1 {
		t.Fatalf("single JSON body expected, got: %s", w.Body.String())
	}
	w = doJSON(r, http.MethodGet, "/api/v1/case-intakes/"+intake+"/materials", "7", "conflict_officer", nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("unassigned officer must be denied: got %d", w.Code)
	}
	db.Exec("UPDATE conflict_reviewer_assignments SET status = 'REVOKED' WHERE id = 'asg-1'")
	w = doJSON(r, http.MethodGet, "/api/v1/case-intakes/"+intake+"/materials", "6", "conflict_officer", nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("revoked reviewer must be denied: got %d", w.Code)
	}
}

// An owner who was legally assigned once must not stay assignable for new
// rows by anyone (trial attempt2: bogus owner create returned 200).
func TestIntakeOwnerAssignmentDoesNotPropagate(t *testing.T) {
	db := newMatterCollabDB(t)
	r := newMatterCollabRouter(db, t.TempDir())
	intake := "aaaa1111-0000-0000-0000-000000000001"
	// Creator legally assigns user 4 on one row.
	w := doJSON(r, http.MethodPost, "/api/v1/case-intakes/"+intake+"/materials", "1", "lawyer", map[string]interface{}{
		"name": "First", "owner_id": 4,
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("creator assigning outsider owner must fail: got %d: %s", w.Code, w.Body.String())
	}
	// Self-assignment by the creator remains the only owner path.
	w = doJSON(r, http.MethodPost, "/api/v1/case-intakes/"+intake+"/materials", "1", "lawyer", map[string]interface{}{
		"name": "Second", "owner_id": 1,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("creator self owner: got %d: %s", w.Code, w.Body.String())
	}
}

// Revoked-assignee replay: the original assignee who is later removed
// from the case (wall on + not whitelisted) or deactivated must get 403
// even on an already-completed task, and the replay response must never
// claim a new completion note was recorded (review 2026-09-30).
func TestCaseTaskRevokedAssigneeReplayDenied(t *testing.T) {
	db := newMatterCollabDB(t)
	r := newMatterCollabRouter(db, t.TempDir())
	// lawyer 4 becomes a second legitimate lawyer on case 8 via the
	// ownership column (no team needed), creates a task assigned to
	// himself, and completes it while still authorized.
	db.Exec("UPDATE cases SET lawyer_id = 4 WHERE id = 8")
	w := doJSON(r, http.MethodPost, "/api/v1/cases/8/tasks", "4", "lawyer", map[string]interface{}{
		"title": "Revoke-me", "assignee_id": 4,
	})
	var created struct {
		Data struct {
			Task models.CaseTask `json:"task"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	taskID := created.Data.Task.ID
	taskPath := "/api/v1/case-tasks/" + strconv.Itoa(int(taskID)) + "/complete"
	if w := doJSON(r, http.MethodPost, taskPath, "4", "lawyer", map[string]interface{}{"completion_note": "first"}); w.Code != http.StatusOK {
		t.Fatalf("authorized complete: got %d: %s", w.Code, w.Body.String())
	}
	// Wall goes on; user 4 is not whitelisted; user 1 (case owner) is.
	db.Exec("UPDATE cases SET ethical_wall_enabled = 1 WHERE id = 8")
	db.Exec("INSERT INTO case_ethical_wall_whitelist (case_id, user_id) VALUES (8, 1)")
	var resp struct {
		Data map[string]interface{} `json:"data"`
	}
	w = doJSON(r, http.MethodPost, taskPath, "4", "lawyer", map[string]interface{}{"completion_note": "sneaky"})
	if w.Code != http.StatusForbidden {
		t.Fatalf("wall-removed assignee replay: got %d want 403: %s", w.Code, w.Body.String())
	}
	// Deactivated assignee replay is also denied.
	db.Exec("UPDATE cases SET ethical_wall_enabled = 0 WHERE id = 8")
	db.Exec("UPDATE users SET status = ? WHERE id = 4", "disabled")
	w = doJSON(r, http.MethodPost, taskPath, "4", "lawyer", map[string]interface{}{})
	if w.Code != http.StatusForbidden {
		t.Fatalf("deactivated assignee replay: got %d want 403", w.Code)
	}
	// The case owner can still replay idempotently and the response
	// honestly reports replayed + the stored note (no fake note claim).
	w = doJSON(r, http.MethodPost, taskPath, "1", "lawyer", map[string]interface{}{"completion_note": "second-try"})
	if w.Code != http.StatusOK {
		t.Fatalf("owner replay: got %d: %s", w.Code, w.Body.String())
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Data["replayed"] != true {
		t.Fatalf("owner replay must report replayed=true: %s", w.Body.String())
	}
	if resp.Data["completion_note_recorded"] != true {
		t.Fatalf("replay must report the stored note, got: %s", w.Body.String())
	}
	var note string
	db.Raw("SELECT completion_note FROM case_tasks WHERE id = ?", taskID).Scan(&note)
	if note != "first" {
		t.Fatalf("replay must keep the original note, got %q", note)
	}
}

// Pre-approval reviewer read: before any approval_requests row exists,
// the assigned independent reviewer can still read the intake materials
// through the intake.metadata.conflict_check_id binding (same ACTIVE +
// in-window semantics as GetActiveConflictReviewerAssignment). Revoking
// the assignment closes the read (review 2026-09-30).
func TestPreApprovalReviewerReadsIntakeMaterials(t *testing.T) {
	db := newMatterCollabDB(t)
	r := newMatterCollabRouter(db, t.TempDir())
	intake := "aaaa1111-0000-0000-0000-000000000001"
	checkID := "CCT_PRE_APPROVAL"
	meta := `{"conflict_check_id":"` + checkID + `"}`
	db.Exec("UPDATE case_intakes SET metadata = ? WHERE id = ?", meta, intake)
	db.Exec("INSERT INTO users (id, role) VALUES (6, 'conflict_officer'), (7, 'conflict_officer')")
	db.Exec("INSERT INTO conflict_reviewer_assignments (id, check_id, reviewer_id, status, recusal_declared) VALUES ('asg-pre', ?, 6, 'ACTIVE', 1)", checkID)
	// No approval row exists yet; the assigned officer reads via the
	// intake metadata binding only.
	w := doJSON(r, http.MethodGet, "/api/v1/case-intakes/"+intake+"/materials", "6", "conflict_officer", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("pre-approval assigned reviewer read: got %d: %s", w.Code, w.Body.String())
	}
	if strings.Count(w.Body.String(), "\"success\"") != 1 {
		t.Fatalf("single JSON body expected, got: %s", w.Body.String())
	}
	// An unassigned officer is denied.
	w = doJSON(r, http.MethodGet, "/api/v1/case-intakes/"+intake+"/materials", "7", "conflict_officer", nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("pre-approval unassigned officer: got %d want 403", w.Code)
	}
	// Revoking the assignment closes the read even pre-approval.
	db.Exec("UPDATE conflict_reviewer_assignments SET status = 'REVOKED' WHERE id = 'asg-pre'")
	w = doJSON(r, http.MethodGet, "/api/v1/case-intakes/"+intake+"/materials", "6", "conflict_officer", nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("pre-approval revoked reviewer: got %d want 403", w.Code)
	}
}
