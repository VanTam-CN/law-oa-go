package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"law-oa-go/internal/config"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
)

type npPGFixture struct {
	db                  *gorm.DB
	handler             *NarrowPilotHandler
	server              *httptest.Server
	token, id, revision string
	owner               uint
}

func npPostgresFixture(t *testing.T) *npPGFixture {
	t.Helper()
	dsn := os.Getenv("NP_TEST_DSN")
	if dsn == "" {
		t.Skip("NP_TEST_DSN is required for isolated PostgreSQL evidence")
	}
	if !strings.Contains(dsn, "dbname=np_first_slice") || !strings.Contains(dsn, "host=127.0.0.1") || !strings.Contains(dsn, "port=18790") {
		t.Fatal("refusing non-isolated NP_TEST_DSN")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	var name string
	if err := db.Raw("select current_database()").Scan(&name).Error; err != nil || name != "np_first_slice_dev" {
		t.Fatal("wrong database", name, err)
	}
	tag := uuid.NewString()
	user := models.User{Username: "np-" + tag, Name: "虚构试验律师", Email: tag + "@np.invalid", Password: "unused-test-only", Role: "lawyer", Status: "active"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	client := models.Client{Name: "虚构客户-" + tag, Type: "企业", Email: tag + "-client@np.invalid", CreatedBy: user.ID}
	if err := db.Create(&client).Error; err != nil {
		t.Fatal(err)
	}
	id := uuid.NewString()
	revision := uuid.NewString()
	if err := db.Table("case_intakes").Create(map[string]interface{}{"id": id, "intake_code": "NP-" + tag, "title": "虚构接案·甲", "client_id": client.ID, "created_by": user.ID, "description": "原摘要标记", "status": "draft", "metadata": fmt.Sprintf(`{"draft_revision":%q,"np_unrelated":{"keep":true,"big":9007199254740993,"decimal":0.12345678901234567890123456789},"lawyer_id":%d}`, revision, user.ID)}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Table("case_intake_parties").Create(map[string]interface{}{"intake_id": id, "entity_name": "虚构对方", "entity_type": "LEGAL_PERSON", "party_role": "opposing_party", "identity_type": "SOCIAL_CREDIT_CODE", "identity_number_ciphertext": "test-only-cipher", "identity_number_digest": "test-only-digest", "metadata": `{"relationship":"keep"}`}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Table("case_materials").Create(map[string]interface{}{"intake_id": id, "name": "虚构材料", "material_type": "other", "metadata": `{"preserve":true}`}).Error; err != nil {
		t.Fatal(err)
	}
	middleware.InitJWT(&config.Config{JWT: config.JWTConfig{Secret: strings.Repeat("n", 40)}})
	token, _, err := middleware.GenerateToken(user.ID, user.Username, "lawyer")
	if err != nil {
		t.Fatal(err)
	}
	claims, _ := middleware.ValidateToken(token)
	if err := db.Create(&models.AuthTokenSession{ID: uuid.NewString(), UserID: user.ID, AccessTokenUUID: claims.UUID, RefreshTokenUUID: uuid.NewString(), AccessTokenExpires: time.Now().Add(time.Hour), RefreshTokenExpires: time.Now().Add(2 * time.Hour)}).Error; err != nil {
		t.Fatal(err)
	}
	h := NewDemoAggregateHandler(db)
	h.SetAuthorizationService(services.NewAuthorizationServiceWithDB(db))
	n := NewNarrowPilotHandler(h)
	n.allowed[fmt.Sprint(user.ID)] = true
	r := gin.New()
	r.Use(middleware.AuthMiddleware())
	materialHandler := NewMatterCollaborationHandler(services.NewAuthorizationServiceWithDB(db), db)
	materialHandler.StorageDir = t.TempDir()
	r.POST("/api/v1/case-materials/:materialId/files", materialHandler.UploadCaseMaterial)
	r.GET("/api/v1/case-materials/:materialId/files", materialHandler.ListCaseMaterialFiles)
	r.GET("/api/v1/case-materials/:materialId/files/:fileId/download", materialHandler.DownloadCaseMaterialFile)
	r.GET("/api/v1/case-intakes/:id", h.IntakeWorkbench)
	r.PUT("/api/v1/case-intakes/:id", h.UpdateCaseIntake)
	r.GET("/api/v1/case-intakes/:id/source-snapshots/:revision", n.SourceSnapshot)
	r.GET("/api/v1/case-intakes/write-results/:request_id", n.WriteResult)
	r.POST("/api/v1/case-intakes/write-results/:request_id/abandon", n.Abandon)
	r.POST("/api/v1/case-intake-assist/contexts", n.CreateContext)
	r.POST("/api/v1/case-intake-assist/contexts/:id/turns", n.Turn)
	r.GET("/api/v1/case-intake-assist/contexts/:id", n.GetContext)
	server := httptest.NewServer(r)
	t.Cleanup(server.Close)
	return &npPGFixture{db, n, server, token, id, revision, user.ID}
}
func (f *npPGFixture) request(t *testing.T, method, path, key string, body interface{}) (int, map[string]interface{}) {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(method, f.server.URL+path, bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+f.token)
	req.Header.Set("Content-Type", "application/json")
	if key != "" {
		req.Header.Set("Idempotency-Key", key)
	}
	res, err := (&http.Client{Timeout: 8 * time.Second}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var out map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return res.StatusCode, out
}
func (f *npPGFixture) candidate(t *testing.T) (string, map[string]interface{}) {
	t.Helper()
	code, out := f.request(t, "POST", "/api/v1/case-intake-assist/contexts", "", map[string]interface{}{"intake_id": f.id, "expected_revision": f.revision})
	ctx := npData(t, code, out)
	id := ctx["id"].(string)
	for i, text := range []string{"这是虚构资料。因对方未按时交货，我方要求确认交付时间。", "补充：交付日期尚未核对，不能写成已确认。"} {
		code, out = f.request(t, "POST", "/api/v1/case-intake-assist/contexts/"+id+"/turns", "", map[string]interface{}{"request_id": uuid.NewString(), "expected_context_version": i + 1, "text": text})
		ctx = npData(t, code, out)
	}
	candidate := ctx["candidate"].(map[string]interface{})
	return candidate["description"].(string), map[string]interface{}{"context_id": id, "expected_context_version": 3, "fields": map[string]interface{}{"description": map[string]interface{}{"candidate_id": candidate["id"], "requested_mode": "candidate_sources", "reviewed": true}}}
}
func (f *npPGFixture) body(text string, source interface{}) map[string]interface{} {
	body := map[string]interface{}{"expected_revision": f.revision, "changed_fields": []string{"description"}, "description": text}
	if source != nil {
		body["source_selection"] = source
	}
	return body
}
func (f *npPGFixture) preserved(t *testing.T) string {
	t.Helper()
	var result string
	err := f.db.Raw(`SELECT jsonb_build_object('title',title,'client',client_id,'status',status,'unknown',metadata->'np_unrelated','parties',(SELECT jsonb_agg(to_jsonb(p) ORDER BY id) FROM case_intake_parties p WHERE intake_id=i.id),'files',(SELECT jsonb_agg(to_jsonb(f) ORDER BY f.id) FROM material_files f JOIN case_materials m ON f.material_id=m.id WHERE m.intake_id=i.id),'materials',(SELECT jsonb_agg(to_jsonb(m) ORDER BY id) FROM case_materials m WHERE intake_id=i.id))::text FROM case_intakes i WHERE id=?`, f.id).Scan(&result).Error
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestNPPostgresSaveReopenConcurrencyAndReceipts(t *testing.T) {
	f := npPostgresFixture(t)
	f.materialVersions(t, true)
	original := f.preserved(t)
	text, selection := f.candidate(t)
	body := f.body(text, selection)
	key := uuid.NewString()
	path := "/api/v1/case-intakes/" + f.id
	code, out := f.request(t, "PUT", path, key, body)
	saved := npData(t, code, out)
	newRev := saved["revision"].(string)
	if newRev == f.revision || saved["id"] != f.id {
		t.Fatal("not actual saved revision", saved)
	}
	if f.preserved(t) != original {
		t.Fatal("unselected fields or child rows changed")
	}
	code, out = f.request(t, "GET", path, "", nil)
	restored := npData(t, code, out)
	if restored["revision"] != newRev || restored["intake"].(map[string]interface{})["description"] != text {
		t.Fatal("native reopen differs", restored)
	}
	f.materialVersions(t, false)
	code, out = f.request(t, "GET", path+"/source-snapshots/"+newRev, "", nil)
	source := npData(t, code, out)
	encoded, _ := json.Marshal(source)
	if !bytes.Contains(encoded, []byte("交付日期尚未核对")) {
		t.Fatal("saved source missing", source)
	}
	code, out = f.request(t, "PUT", path, uuid.NewString(), f.body("B标签纯手工摘要", nil))
	if code != 409 {
		t.Fatalf("stale manual overwritten %d %v", code, out)
	}
	code, out = f.request(t, "PUT", path, key, body)
	if npData(t, code, out)["revision"] != newRev {
		t.Fatal("same-key replay did not return original")
	}
	var receipts, snapshots int64
	if err := f.db.Model(&npReceiptRow{}).Where("actor_id=? AND request_id=?", fmt.Sprint(f.owner), key).Count(&receipts).Error; err != nil {
		t.Fatal(err)
	}
	if err := f.db.Model(&npSnapshotRow{}).Where("intake_id=?", f.id).Count(&snapshots).Error; err != nil {
		t.Fatal(err)
	}
	if receipts != 1 || snapshots != 1 {
		t.Fatalf("replay/stale save created duplicates: receipts=%d snapshots=%d", receipts, snapshots)
	}
	changed := f.body("同键不同参数", nil)
	code, _ = f.request(t, "PUT", path, key, changed)
	if code != 409 {
		t.Fatal("same-key different body allowed")
	}
	code, _ = f.request(t, "PUT", path, "", map[string]interface{}{"expected_revision": newRev, "title": "全量绕行", "parties": []interface{}{}})
	if code != 409 {
		t.Fatal("legacy bypass allowed")
	}
	code, out = f.request(t, "POST", "/api/v1/case-intakes/write-results/"+key+"/abandon", "", map[string]interface{}{"action": "update", "intake_id": f.id})
	if npData(t, code, out)["state"] != "committed" {
		t.Fatal("abandon hid saved result")
	}
	tombstone := uuid.NewString()
	code, out = f.request(t, "GET", "/api/v1/case-intakes/write-results/"+tombstone, "", nil)
	if npData(t, code, out)["state"] != "unknown" {
		t.Fatal("read created result")
	}
	code, out = f.request(t, "POST", "/api/v1/case-intakes/write-results/"+tombstone+"/abandon", "", map[string]interface{}{"action": "update", "intake_id": f.id})
	if npData(t, code, out)["state"] != "abandoned" {
		t.Fatal("missing tombstone")
	}
	f.revision = newRev
	code, _ = f.request(t, "PUT", path, tombstone, f.body("迟到原请求不得写入", nil))
	if code != 409 {
		t.Fatal("late request resurrected")
	}
	if f.preserved(t) != original {
		t.Fatal("rejected requests changed protected fields")
	}
	// The same text with a new reviewed source is a new association, not no_change.
	textAgain, newSelection := f.candidate(t)
	if textAgain != text {
		t.Fatal("synthetic same-text source fixture changed")
	}
	code, out = f.request(t, "PUT", path, uuid.NewString(), f.body(textAgain, newSelection))
	rebound := npData(t, code, out)
	if rebound["state"] != "committed" || rebound["revision"] == newRev {
		t.Fatal("source-only rebinding did not advance version", rebound)
	}
	f.revision = rebound["revision"].(string)
	code, out = f.request(t, "PUT", path, uuid.NewString(), f.body(text, nil))
	if npData(t, code, out)["state"] != "no_change" {
		t.Fatal("identical manual text should be no change")
	}
	manualKey := uuid.NewString()
	code, out = f.request(t, "PUT", path, manualKey, f.body("律师手工修改摘要", nil))
	manual := npData(t, code, out)
	code, out = f.request(t, "GET", path+"/source-snapshots/"+manual["revision"].(string), "", nil)
	manualSource := npData(t, code, out)
	if manualSource["source_revision"] != "" && manualSource["source_revision"] != nil {
		t.Fatal("manual edits retained automatic linkage", manualSource)
	}
	t.Log("PG: multi-turn source, native reopen, stale manual, exact replay, different digest, read-only unknown, tombstone late arrival, no-change and manual unlink verified")
}

func (f *npPGFixture) materialVersions(t *testing.T, upload bool) {
	t.Helper()
	var materialID uint
	if err := f.db.Table("case_materials").Select("id").Where("intake_id=?", f.id).Scan(&materialID).Error; err != nil {
		t.Fatal(err)
	}
	path := fmt.Sprintf("/api/v1/case-materials/%d/files", materialID)
	contents := []string{"fictional original material bytes", "fictional corrected material bytes"}
	if upload {
		for version, text := range contents {
			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			part, err := writer.CreateFormFile("file", fmt.Sprintf("np-v%d.txt", version+1))
			if err != nil {
				t.Fatal(err)
			}
			if _, err = part.Write([]byte(text)); err != nil {
				t.Fatal(err)
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			req, err := http.NewRequest("POST", f.server.URL+path, &body)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Authorization", "Bearer "+f.token)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			response, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			raw, _ := io.ReadAll(response.Body)
			response.Body.Close()
			if response.StatusCode != 200 {
				t.Fatalf("material version upload %d: %d %s", version, response.StatusCode, raw)
			}
		}
	}
	code, out := f.request(t, "GET", path, "", nil)
	files := npData(t, code, out)["files"].([]interface{})
	if len(files) != 2 {
		t.Fatal("material versions missing", files)
	}
	for i, file := range files {
		row := file.(map[string]interface{})
		req, err := http.NewRequest("GET", f.server.URL+row["download_url"].(string), nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+f.token)
		response, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := io.ReadAll(response.Body)
		response.Body.Close()
		if response.StatusCode != 200 || string(raw) != contents[i] {
			t.Fatalf("original bytes mismatch for version %d: status=%d", i, response.StatusCode)
		}
	}
}

// The proxy receives the committed backend response but withholds it from the client.
// Closing the real client socket proves a lost response, rather than a mocked UI error.
func TestNPPostgresActualDisconnectAfterCommit(t *testing.T) {
	f := npPostgresFixture(t)
	text, selection := f.candidate(t)
	body := f.body(text, selection)
	key := uuid.NewString()
	upstream, _ := url.Parse(f.server.URL)
	proxy := httputil.NewSingleHostReverseProxy(upstream)
	committed := make(chan struct{})
	release := make(chan struct{})
	proxy.ModifyResponse = func(r *http.Response) error {
		if r.StatusCode == 200 {
			close(committed)
			<-release
		}
		return nil
	}
	proxy.ErrorLog = nil
	server := httptest.NewServer(proxy)
	defer server.Close()
	u, _ := url.Parse(server.URL)
	conn, err := net.Dial("tcp", u.Host)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(body)
	_, err = fmt.Fprintf(conn, "PUT /api/v1/case-intakes/%s HTTP/1.1\r\nHost: %s\r\nAuthorization: Bearer %s\r\nIdempotency-Key: %s\r\nContent-Type: application/json\r\nContent-Length: %d\r\n\r\n%s", f.id, u.Host, f.token, key, len(raw), raw)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-committed:
	case <-time.After(8 * time.Second):
		close(release)
		conn.Close()
		t.Fatal("backend did not commit")
	}
	var stored string
	if err := f.db.Table("case_intakes").Select("description").Where("id=?", f.id).Scan(&stored).Error; err != nil || stored != text {
		close(release)
		conn.Close()
		t.Fatal("no committed DB fact before disconnect", err)
	}
	_ = conn.Close()
	close(release)
	code, out := f.request(t, "GET", "/api/v1/case-intakes/write-results/"+key, "", nil)
	result := npData(t, code, out)
	if result["state"] != "committed" {
		t.Fatal("lost result not recoverable", result)
	}
	code, out = f.request(t, "PUT", "/api/v1/case-intakes/"+f.id, key, body)
	if npData(t, code, out)["revision"] != result["revision"] {
		t.Fatal("disconnected retry advanced twice")
	}
	var count int64
	if err := f.db.Model(&npSnapshotRow{}).Where("intake_id=?", f.id).Count(&count).Error; err != nil || count != 1 {
		t.Fatal("disconnect produced extra snapshots", count, err)
	}
	t.Log("PG commit observed before real downstream TCP close; GET and same-key replay recovered the original revision")
}

func TestNPPostgresActualDisconnectBeforeCommit(t *testing.T) {
	f := npPostgresFixture(t)
	key := uuid.NewString()
	block := f.db.Begin()
	if block.Error != nil {
		t.Fatal(block.Error)
	}
	defer block.Rollback()
	var locked string
	if err := block.Raw("SELECT id::text FROM case_intakes WHERE id=? FOR UPDATE", f.id).Scan(&locked).Error; err != nil {
		t.Fatal(err)
	}
	u, _ := url.Parse(f.server.URL)
	conn, err := net.Dial("tcp", u.Host)
	if err != nil {
		t.Fatal(err)
	}
	body := f.body("断线前未提交的虚构摘要", nil)
	raw, _ := json.Marshal(body)
	if _, err := fmt.Fprintf(conn, "PUT /api/v1/case-intakes/%s HTTP/1.1\r\nHost: %s\r\nAuthorization: Bearer %s\r\nIdempotency-Key: %s\r\nContent-Type: application/json\r\nContent-Length: %d\r\n\r\n%s", f.id, u.Host, f.token, key, len(raw), raw); err != nil {
		conn.Close()
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	waiting := false
	for time.Now().Before(deadline) {
		var count int64
		if err := f.db.Raw("SELECT count(*) FROM pg_stat_activity WHERE datname=current_database() AND pid<>pg_backend_pid() AND wait_event_type='Lock' AND query LIKE '%case_intakes%'").Scan(&count).Error; err != nil {
			conn.Close()
			t.Fatal(err)
		}
		if count > 0 {
			waiting = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !waiting {
		conn.Close()
		t.Fatal("did not observe actual PostgreSQL row-lock wait")
	}
	_ = conn.Close()
	// Wait until PostgreSQL observes cancellation before releasing the lock.
	deadline = time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		var count int64
		_ = f.db.Raw("SELECT count(*) FROM pg_stat_activity WHERE datname=current_database() AND pid<>pg_backend_pid() AND wait_event_type='Lock' AND query LIKE '%case_intakes%'").Scan(&count).Error
		if count == 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err := block.Commit().Error; err != nil {
		t.Fatal(err)
	}
	code, out := f.request(t, "GET", "/api/v1/case-intakes/write-results/"+key, "", nil)
	if npData(t, code, out)["state"] != "unknown" {
		t.Fatal("cancelled uncommitted save gained receipt", out)
	}
	code, out = f.request(t, "POST", "/api/v1/case-intakes/write-results/"+key+"/abandon", "", map[string]interface{}{"action": "update", "intake_id": f.id})
	if npData(t, code, out)["state"] != "abandoned" {
		t.Fatal("cancelled save not explicitly closed")
	}
	code, _ = f.request(t, "PUT", "/api/v1/case-intakes/"+f.id, key, body)
	if code != 409 {
		t.Fatal("cancelled then abandoned request resurrected")
	}
	var text string
	if err := f.db.Table("case_intakes").Select("description").Where("id=?", f.id).Scan(&text).Error; err != nil || text != "原摘要标记" {
		t.Fatal("precommit disconnect changed draft", text, err)
	}
	t.Log("Observed real PG row-lock wait; client TCP close cancelled transaction; read-only unknown, explicit tombstone and late retry verified")
}

func TestNPPostgresParallelWritersAndBusyAbandon(t *testing.T) {
	f := npPostgresFixture(t)
	path := "/api/v1/case-intakes/" + f.id
	start := make(chan struct{})
	statuses := make(chan int, 2)
	for _, text := range []string{"并发手工摘要A", "并发手工摘要B"} {
		go func(text string) {
			<-start
			raw, _ := json.Marshal(f.body(text, nil))
			req, _ := http.NewRequest("PUT", f.server.URL+path, bytes.NewReader(raw))
			req.Header.Set("Authorization", "Bearer "+f.token)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Idempotency-Key", uuid.NewString())
			response, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
			if err != nil {
				statuses <- 0
				return
			}
			response.Body.Close()
			statuses <- response.StatusCode
		}(text)
	}
	close(start)
	first, second := <-statuses, <-statuses
	if !(first == 200 && second == 409 || first == 409 && second == 200) {
		t.Fatalf("concurrent results %d,%d", first, second)
	}
	key := uuid.NewString()
	block := f.db.Begin()
	defer block.Rollback()
	if err := block.Exec("SELECT pg_advisory_xact_lock(?)", npAdvisoryKey(fmt.Sprint(f.owner), key)).Error; err != nil {
		t.Fatal(err)
	}
	code, _ := f.request(t, "POST", "/api/v1/case-intakes/write-results/"+key+"/abandon", "", map[string]interface{}{"action": "update", "intake_id": f.id})
	if code != 409 {
		t.Fatal("busy abandon claimed success", code)
	}
	code, out := f.request(t, "GET", "/api/v1/case-intakes/write-results/"+key, "", nil)
	if npData(t, code, out)["state"] != "unknown" {
		t.Fatal("busy abandon persisted tombstone")
	}
	if err := block.Commit().Error; err != nil {
		t.Fatal(err)
	}
	code, out = f.request(t, "POST", "/api/v1/case-intakes/write-results/"+key+"/abandon", "", map[string]interface{}{"action": "update", "intake_id": f.id})
	if npData(t, code, out)["state"] != "abandoned" {
		t.Fatal("unlocked abandon failed")
	}
}

func TestNPPostgresSourcePermissionsAndStrictBoundary(t *testing.T) {
	f := npPostgresFixture(t)
	path := "/api/v1/case-intakes/" + f.id
	for _, extra := range []map[string]interface{}{{"materials": nil}, {"material_changes": []interface{}{}}, {"parties": []interface{}{}}, {"metadata": map[string]interface{}{}}, {"title": "越界标题"}, {"description": nil}, {"description": ""}, {"clear_fields": []string{"description", "description"}}} {
		body := f.body("合法虚构摘要", nil)
		for k, v := range extra {
			body[k] = v
		}
		code, _ := f.request(t, "PUT", path, uuid.NewString(), body)
		if code != 400 {
			t.Fatal("ambiguous update not rejected", extra, code)
		}
	}
	text, selection := f.candidate(t)
	code, out := f.request(t, "PUT", path, uuid.NewString(), f.body(text, selection))
	revision := npData(t, code, out)["revision"].(string)
	if err := f.db.Model(&models.User{}).Where("id=?", f.owner).Update("role", "admin").Error; err != nil {
		t.Fatal(err)
	}
	code, _ = f.request(t, "GET", path+"/source-snapshots/"+revision, "", nil)
	if code != 403 {
		t.Fatal("admin role expanded source access", code)
	}
	code, _ = f.request(t, "PUT", path, "", map[string]interface{}{"title": "旧协议绕行", "expected_revision": revision})
	if code != 403 {
		t.Fatal("admin bypassed protected legacy guard", code)
	}
	if err := f.db.Model(&models.User{}).Where("id=?", f.owner).Update("role", "lawyer").Error; err != nil {
		t.Fatal(err)
	}
	f.handler.allowed = map[string]bool{}
	f.revision = revision
	code, _ = f.request(t, "PUT", path, uuid.NewString(), f.body("关闭助手仍可手工保存", nil))
	if code != 200 {
		t.Fatal("assistant switch blocked native description save", code)
	}
}

func TestNPPostgresLegacySaveAfterStrictManualStillChecksRevision(t *testing.T) {
	f := npPostgresFixture(t)
	path := "/api/v1/case-intakes/" + f.id
	code, out := f.request(t, "PUT", path, uuid.NewString(), f.body("严格手工保存须保留", nil))
	npData(t, code, out)
	code, out = f.request(t, "PUT", path, "", map[string]interface{}{"expected_revision": f.revision, "title": "旧表单标题", "description": "旧全量摘要"})
	if code != 409 {
		t.Fatalf("legacy stale after strict manual: %d %v", code, out)
	}
	var text string
	if err := f.db.Table("case_intakes").Select("description").Where("id=?", f.id).Scan(&text).Error; err != nil || text != "严格手工保存须保留" {
		t.Fatal("legacy bypassed revision", text, err)
	}
}

func TestNPPostgresDescriptionWhitespaceIsExact(t *testing.T) {
	f := npPostgresFixture(t)
	path := "/api/v1/case-intakes/" + f.id
	if err := f.db.Table("case_intakes").Where("id=?", f.id).Update("description", "  字段原文  ").Error; err != nil {
		t.Fatal(err)
	}
	code, out := f.request(t, "PUT", path, uuid.NewString(), f.body("  字段原文  ", nil))
	same := npData(t, code, out)
	if same["state"] != "no_change" || same["revision"] != f.revision {
		t.Fatal("identical whitespace was treated as a change", same)
	}
	code, out = f.request(t, "PUT", path, uuid.NewString(), f.body("字段原文", nil))
	changed := npData(t, code, out)
	if changed["state"] != "committed" || changed["revision"] == f.revision {
		t.Fatal("explicit whitespace removal was lost", changed)
	}
	var stored string
	if err := f.db.Table("case_intakes").Select("description").Where("id=?", f.id).Scan(&stored).Error; err != nil || stored != "字段原文" {
		t.Fatal("not exact stored text", stored, err)
	}
}
