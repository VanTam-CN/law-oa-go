package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
)

// newIPv4TestServer avoids relying on the host's IPv6 loopback permissions.
// The application behavior under test only needs a local HTTP endpoint.
func newIPv4TestServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("创建 IPv4 测试监听器失败: %v", err)
	}
	ts := &httptest.Server{
		Listener: listener,
		Config:   &http.Server{Handler: handler},
	}
	ts.Start()
	return ts
}

func setupOnlyOfficeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	if err := db.AutoMigrate(&models.Document{}, &models.User{}, &models.DocumentLock{}); err != nil {
		t.Fatalf("迁移失败: %v", err)
	}
	return db
}

func setupOnlyOfficeTestHandler(t *testing.T, db *gorm.DB) *OnlyOfficeHandler {
	t.Helper()
	tmpDir := t.TempDir()
	lockService := services.NewDocumentLockService(db, nil)
	versionService := services.NewDocumentVersionService(nil, tmpDir)
	return NewOnlyOfficeHandler(db, versionService, lockService, "", "", "", tmpDir)
}

func signOnlyOfficeCallback(t *testing.T, h *OnlyOfficeHandler, req CallbackRequest) []byte {
	t.Helper()
	req.Token = ""
	unsigned, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal unsigned callback: %v", err)
	}
	signingPayload, err := canonicalCallbackSigningPayload(unsigned)
	if err != nil {
		t.Fatalf("canonical callback payload: %v", err)
	}
	req.Token = h.GenerateCallbackToken(string(signingPayload))
	signed, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal signed callback: %v", err)
	}
	return signed
}

// TestGetDocumentType 测试文档类型判断
func TestGetDocumentType(t *testing.T) {
	h := &OnlyOfficeHandler{}

	tests := []struct {
		ext  string
		want string
	}{
		{".docx", DocumentTypeWord},
		{".doc", DocumentTypeWord},
		{".pdf", DocumentTypeWord},
		{".txt", DocumentTypeWord},
		{".xlsx", DocumentTypeCell},
		{".xls", DocumentTypeCell},
		{".csv", DocumentTypeCell},
		{".pptx", DocumentTypeSlide},
		{".ppt", DocumentTypeSlide},
		{".unknown", DocumentTypeWord},
	}

	for _, tt := range tests {
		got := h.getDocumentType(tt.ext)
		if got != tt.want {
			t.Errorf("getDocumentType(%q) = %q, want %q", tt.ext, got, tt.want)
		}
	}
}

// TestGenerateAndParseDocKey 测试文档 key 的生成和解析
func TestGenerateAndParseDocKey(t *testing.T) {
	h := &OnlyOfficeHandler{}
	now := time.Now()
	key := h.generateDocKey(42, now)

	parsed, err := h.parseDocKey(key)
	if err != nil {
		t.Fatalf("parseDocKey 失败: %v", err)
	}
	if parsed != 42 {
		t.Errorf("解析后 DocumentID = %d, 期望 42", parsed)
	}
}

// TestParseDocKey_Invalid 测试无效 key
func TestParseDocKey_Invalid(t *testing.T) {
	h := &OnlyOfficeHandler{}

	_, err := h.parseDocKey("invalid")
	if err == nil {
		t.Fatal("期望解析错误，但没有返回错误")
	}
}

// TestIsValidOutputType 测试输出类型验证
func TestIsValidOutputType(t *testing.T) {
	h := &OnlyOfficeHandler{}

	validTypes := []string{"pdf", "docx", "xlsx", "pptx", "odt", "txt", "csv"}
	for _, vt := range validTypes {
		if !h.isValidOutputType(vt) {
			t.Errorf("isValidOutputType(%q) = false, 期望 true", vt)
		}
	}

	if h.isValidOutputType("exe") {
		t.Error("isValidOutputType(\"exe\") = true, 期望 false")
	}
	if h.isValidOutputType("") {
		t.Error("isValidOutputType(\"\") = true, 期望 false")
	}
}

// TestSaveDocument_InvalidURL 测试空下载 URL
func TestSaveDocument_InvalidURL(t *testing.T) {
	db := setupOnlyOfficeTestDB(t)
	h := setupOnlyOfficeTestHandler(t, db)

	err := h.saveDocument(context.Background(), 999, "")
	if err == nil {
		t.Fatal("期望错误，但没有返回")
	}
}

// TestSaveDocument_DocumentNotFound 测试文档不存在
func TestSaveDocument_DocumentNotFound(t *testing.T) {
	db := setupOnlyOfficeTestDB(t)
	h := setupOnlyOfficeTestHandler(t, db)

	err := h.saveDocument(context.Background(), 999, "http://example.com/file.docx")
	if err == nil {
		t.Fatal("期望错误，但没有返回")
	}
}

// TestSaveDocument_Success 测试文档保存（使用本地 HTTP 服务器模拟）
func TestSaveDocument_Success(t *testing.T) {
	db := setupOnlyOfficeTestDB(t)

	// 创建文档记录
	doc := &models.Document{
		Name:     "测试文档",
		Filename: "test.docx",
		Filepath: filepath.Join(t.TempDir(), "test.docx"),
		Filesize: 100,
		MimeType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		Status:   "active",
	}
	db.Create(doc)

	// 创建初始文件
	if err := os.WriteFile(doc.Filepath, []byte("original content"), 0o644); err != nil {
		t.Fatalf("创建初始文件失败: %v", err)
	}

	// 创建模拟的 OnlyOffice 服务器
	content := "edited content from onlyoffice"
	ts := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write([]byte(content))
	}))
	defer ts.Close()

	// SSRF 防护要求：onlyOfficeURL 必须与下载 URL 的 host:port 一致
	h := setupOnlyOfficeHandlerWithServer(t, db, ts.URL)
	h.onlyOfficeSecret = "test-onlyoffice-secret-at-least-32-characters"
	h.httpClient = ts.Client()

	err := h.saveDocument(context.Background(), doc.ID, ts.URL+"/file.docx")
	if err != nil {
		t.Fatalf("saveDocument 失败: %v", err)
	}

	// 验证文件已更新
	data, err := os.ReadFile(doc.Filepath)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}
	if string(data) != content {
		t.Errorf("文件内容 = %q, 期望 %q", string(data), content)
	}

	// 验证数据库已更新
	var updated models.Document
	db.First(&updated, doc.ID)
	if updated.Filesize != int64(len(content)) {
		t.Errorf("数据库 filesize = %d, 期望 %d", updated.Filesize, len(content))
	}
}

// TestSaveDocument_CreatesVersionBackup 测试版本备份创建
func TestSaveDocument_CreatesVersionBackup(t *testing.T) {
	db := setupOnlyOfficeTestDB(t)
	tmpDir := t.TempDir()

	doc := &models.Document{
		Name:     "备份测试",
		Filename: "backup.docx",
		Filepath: filepath.Join(tmpDir, "test.docx"),
		Filesize: 10,
		Status:   "active",
	}
	db.Create(doc)
	os.WriteFile(doc.Filepath, []byte("original"), 0o644)

	ts := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("new version"))
	}))
	defer ts.Close()

	h := setupOnlyOfficeHandlerWithServer(t, db, ts.URL)
	h.onlyOfficeSecret = "test-onlyoffice-secret-at-least-32-characters"
	h.httpClient = ts.Client()

	err := h.saveDocument(context.Background(), doc.ID, ts.URL+"/file.docx")
	if err != nil {
		t.Fatalf("saveDocument 失败: %v", err)
	}

	// 验证版本备份目录存在
	entries, err := os.ReadDir(filepath.Join(h.storageDir, "versions"))
	if err != nil {
		t.Fatalf("读取版本目录失败: %v", err)
	}
	if len(entries) == 0 {
		t.Error("期望创建版本备份，但目录为空")
	}
}

// TestHandleCallback_MustSave 测试回调处理 - 必须保存
func TestHandleCallback_MustSave(t *testing.T) {
	db := setupOnlyOfficeTestDB(t)
	tmpDir := t.TempDir()

	doc := &models.Document{
		Name:     "回调测试",
		Filename: "callback.docx",
		Filepath: filepath.Join(tmpDir, "test.docx"),
		Filesize: 10,
		Status:   "active",
	}
	db.Create(doc)
	os.WriteFile(doc.Filepath, []byte("before"), 0o644)

	ts := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("after save"))
	}))
	defer ts.Close()

	h := setupOnlyOfficeHandlerWithServer(t, db, ts.URL)
	h.onlyOfficeSecret = "test-onlyoffice-secret-at-least-32-characters"
	h.httpClient = ts.Client()

	// 生成与 parseDocKey 兼容的 key
	docKey := h.generateDocKey(doc.ID, doc.UpdatedAt)

	body := CallbackRequest{
		Key:    docKey,
		Status: StatusMustSave,
		URL:    ts.URL + "/saved.docx",
	}

	bodyBytes := signOnlyOfficeCallback(t, h, body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/documents/onlyoffice/callback", bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	h.HandleCallback(c)

	if w.Code != http.StatusOK {
		t.Errorf("状态码 = %d, 期望 %d", w.Code, http.StatusOK)
	}
}

func TestHandleCallback_RejectsMissingSecret(t *testing.T) {
	h := &OnlyOfficeHandler{onlyOfficeSecret: ""}
	body := CallbackRequest{
		Key:    "1_123",
		Status: StatusEditing,
	}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/documents/onlyoffice/callback", bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	h.HandleCallback(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("未配置密钥时回调必须拒绝: got %d body=%s", w.Code, w.Body.String())
	}
}

func TestHandleCallback_RejectsInvalidToken(t *testing.T) {
	h := &OnlyOfficeHandler{onlyOfficeSecret: "test-onlyoffice-secret-at-least-32-characters"}
	body := CallbackRequest{
		Key:    "1_123",
		Status: StatusEditing,
		Token:  "invalid",
	}
	bodyBytes, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/documents/onlyoffice/callback", bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	h.HandleCallback(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("无效 token 必须拒绝: got %d body=%s", w.Code, w.Body.String())
	}
}

// TestConvertDocument_InvalidOutputType 测试无效输出类型
func TestConvertDocument_InvalidOutputType(t *testing.T) {
	db := setupOnlyOfficeTestDB(t)
	h := setupOnlyOfficeTestHandler(t, db)

	doc := &models.Document{Name: "测试", Filename: "test.docx", Status: "active"}
	db.Create(doc)

	body := `{"document_id": 1, "output_type": "exe"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/documents/onlyoffice/convert", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.ConvertDocument(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("状态码 = %d, 期望 %d", w.Code, http.StatusBadRequest)
	}
}

// TestCheckConversionStatus_NotFound 测试查询不存在的转换任务
func TestCheckConversionStatus_NotFound(t *testing.T) {
	db := setupOnlyOfficeTestDB(t)
	h := setupOnlyOfficeTestHandler(t, db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/documents/onlyoffice/convert/status?task_id=nonexistent", nil)

	h.CheckConversionStatus(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("状态码 = %d, 期望 %d", w.Code, http.StatusNotFound)
	}
}

// TestGenerateCallbackToken 测试 token 生成
func TestGenerateCallbackToken(t *testing.T) {
	h := &OnlyOfficeHandler{onlyOfficeSecret: "test-secret"}

	token1 := h.GenerateCallbackToken("payload1")
	token2 := h.GenerateCallbackToken("payload2")
	token3 := h.GenerateCallbackToken("payload1")

	if token1 == token2 {
		t.Error("不同 payload 不应生成相同 token")
	}
	if token1 != token3 {
		t.Error("相同 payload 应生成相同 token")
	}
	if token1 == "" {
		t.Error("token 不应为空")
	}
}

// TestGenerateCallbackToken_EmptySecret 测试空密钥
func TestGenerateCallbackToken_EmptySecret(t *testing.T) {
	h := &OnlyOfficeHandler{onlyOfficeSecret: ""}

	token := h.GenerateCallbackToken("payload")
	if token != "" {
		t.Error("空密钥应返回空 token")
	}
}

// =====================================================================
// 计划 Task 4: SSRF 防护 + 大小限制 + 失败保护
// =====================================================================

// helper：构造带 onlyOfficeURL 的 handler，方便 SSRF 场景测试
func setupOnlyOfficeHandlerWithServer(t *testing.T, db *gorm.DB, onlyOfficeURL string) *OnlyOfficeHandler {
	t.Helper()
	tmpDir := t.TempDir()
	lockService := services.NewDocumentLockService(db, nil)
	versionService := services.NewDocumentVersionService(nil, tmpDir)
	return NewOnlyOfficeHandler(db, versionService, lockService, onlyOfficeURL, "", "", tmpDir)
}

// TestSaveDocument_SSRF_RejectsDifferentHost 当 onlyOfficeURL 配置为 A 时，下载 URL 必须在 A
func TestSaveDocument_SSRF_RejectsDifferentHost(t *testing.T) {
	db := setupOnlyOfficeTestDB(t)
	doc := &models.Document{
		Name: "ssrf", Filename: "f.docx", Filepath: filepath.Join(t.TempDir(), "f.docx"),
		Filesize: 10, Status: "active",
	}
	db.Create(doc)
	os.WriteFile(doc.Filepath, []byte("orig"), 0o644)

	// 配置 onlyOfficeURL 指向 onlyoffice.example, 但下载 URL 指向 evil.example
	h := setupOnlyOfficeHandlerWithServer(t, db, "http://onlyoffice.example:9090")

	err := h.saveDocument(context.Background(), doc.ID, "http://evil.example/file.docx")
	if err == nil {
		t.Fatal("不同 host 必须被拒绝")
	}
	// 原文件不得被改动
	data, _ := os.ReadFile(doc.Filepath)
	if string(data) != "orig" {
		t.Errorf("原文件被修改: %q", string(data))
	}
}

// TestSaveDocument_SSRF_RejectsDifferentPort 同 host 不同端口
func TestSaveDocument_SSRF_RejectsDifferentPort(t *testing.T) {
	db := setupOnlyOfficeTestDB(t)
	doc := &models.Document{
		Name: "ssrf-port", Filename: "f.docx", Filepath: filepath.Join(t.TempDir(), "f.docx"),
		Filesize: 10, Status: "active",
	}
	db.Create(doc)
	os.WriteFile(doc.Filepath, []byte("orig"), 0o644)

	h := setupOnlyOfficeHandlerWithServer(t, db, "http://onlyoffice.example:9090")

	err := h.saveDocument(context.Background(), doc.ID, "http://onlyoffice.example:8080/file.docx")
	if err == nil {
		t.Fatal("不同端口必须被拒绝")
	}
}

// TestSaveDocument_SSRF_RejectsLoopback 只允许 host:port = onlyOfficeURL
func TestSaveDocument_SSRF_RejectsLoopback(t *testing.T) {
	db := setupOnlyOfficeTestDB(t)
	doc := &models.Document{
		Name: "ssrf-loopback", Filename: "f.docx", Filepath: filepath.Join(t.TempDir(), "f.docx"),
		Filesize: 10, Status: "active",
	}
	db.Create(doc)
	os.WriteFile(doc.Filepath, []byte("orig"), 0o644)

	// 配置 onlyOfficeURL 为公网，下载 URL 指向 127.0.0.1
	h := setupOnlyOfficeHandlerWithServer(t, db, "http://onlyoffice.example:9090")

	err := h.saveDocument(context.Background(), doc.ID, "http://127.0.0.1:9090/file.docx")
	if err == nil {
		t.Fatal("loopback 必须被拒绝")
	}
}

// TestSaveDocument_SSRF_RejectsUnsupportedScheme 非 http/https 必须 reject
func TestSaveDocument_SSRF_RejectsUnsupportedScheme(t *testing.T) {
	db := setupOnlyOfficeTestDB(t)
	doc := &models.Document{
		Name: "ssrf-scheme", Filename: "f.docx", Filepath: filepath.Join(t.TempDir(), "f.docx"),
		Filesize: 10, Status: "active",
	}
	db.Create(doc)
	os.WriteFile(doc.Filepath, []byte("orig"), 0o644)

	h := setupOnlyOfficeHandlerWithServer(t, db, "http://onlyoffice.example:9090")

	for _, raw := range []string{
		"file:///etc/passwd",
		"ftp://onlyoffice.example:9090/file.docx",
		"gopher://onlyoffice.example:9090/file.docx",
	} {
		if err := h.saveDocument(context.Background(), doc.ID, raw); err == nil {
			t.Errorf("scheme %s 必须被拒绝", raw)
		}
	}
}

// TestSaveDocument_SSRF_RejectsUserInfo 含 userinfo 的 URL 必须 reject（避免 https://allowed@evil）
func TestSaveDocument_SSRF_RejectsUserInfo(t *testing.T) {
	db := setupOnlyOfficeTestDB(t)
	doc := &models.Document{
		Name: "ssrf-userinfo", Filename: "f.docx", Filepath: filepath.Join(t.TempDir(), "f.docx"),
		Filesize: 10, Status: "active",
	}
	db.Create(doc)
	os.WriteFile(doc.Filepath, []byte("orig"), 0o644)

	h := setupOnlyOfficeHandlerWithServer(t, db, "http://onlyoffice.example:9090")

	raw := "http://evil@onlyoffice.example:9090/file.docx"
	if err := h.saveDocument(context.Background(), doc.ID, raw); err == nil {
		t.Fatal("含 userinfo 的 URL 必须被拒绝")
	}
}

// TestSaveDocument_SizeLimit_50MiB 超过 50 MiB 的响应必须 reject，原文件不动
func TestSaveDocument_SizeLimit_50MiB(t *testing.T) {
	db := setupOnlyOfficeTestDB(t)
	doc := &models.Document{
		Name: "big", Filename: "big.docx", Filepath: filepath.Join(t.TempDir(), "big.docx"),
		Filesize: 10, Status: "active",
	}
	db.Create(doc)
	os.WriteFile(doc.Filepath, []byte("orig"), 0o644)

	// 模拟 OnlyOffice 服务器返回超过 50 MiB 的内容
	ts := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		// 写入 50 MiB + 1 字节，超过 limit
		w.Write(make([]byte, MaxOnlyOfficeDownloadBytes+1))
	}))
	defer ts.Close()

	h := setupOnlyOfficeHandlerWithServer(t, db, ts.URL)
	h.httpClient = ts.Client()

	err := h.saveDocument(context.Background(), doc.ID, ts.URL+"/file.docx")
	if err == nil {
		t.Fatal("超过 50 MiB 必须返回错误")
	}

	data, _ := os.ReadFile(doc.Filepath)
	if string(data) != "orig" {
		t.Errorf("原文件不得被覆盖: got %q", string(data))
	}
}

// TestSaveDocument_DownloadFailure_PreservesOriginal 5xx 状态不得覆盖原文件
func TestSaveDocument_DownloadFailure_PreservesOriginal(t *testing.T) {
	db := setupOnlyOfficeTestDB(t)
	doc := &models.Document{
		Name: "fail", Filename: "f.docx", Filepath: filepath.Join(t.TempDir(), "f.docx"),
		Filesize: 10, Status: "active",
	}
	db.Create(doc)
	os.WriteFile(doc.Filepath, []byte("orig"), 0o644)

	ts := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	h := setupOnlyOfficeHandlerWithServer(t, db, ts.URL)
	h.httpClient = ts.Client()

	err := h.saveDocument(context.Background(), doc.ID, ts.URL+"/file.docx")
	if err == nil {
		t.Fatal("5xx 必须返回错误")
	}

	data, _ := os.ReadFile(doc.Filepath)
	if string(data) != "orig" {
		t.Errorf("5xx 时原文件不得被覆盖: got %q", string(data))
	}
}

// TestSaveDocument_RedirectRejected 302 重定向必须被拒绝（不允许 follow 到任意 host）
func TestSaveDocument_RedirectRejected(t *testing.T) {
	db := setupOnlyOfficeTestDB(t)
	doc := &models.Document{
		Name: "redirect", Filename: "f.docx", Filepath: filepath.Join(t.TempDir(), "f.docx"),
		Filesize: 10, Status: "active",
	}
	db.Create(doc)
	os.WriteFile(doc.Filepath, []byte("orig"), 0o644)

	// 目标 host：allowed.com:9090；重定向到 evil.com
	ts := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://evil.example/file.docx", http.StatusFound)
	}))
	defer ts.Close()

	h := setupOnlyOfficeHandlerWithServer(t, db, ts.URL)
	h.httpClient = ts.Client()

	err := h.saveDocument(context.Background(), doc.ID, ts.URL+"/file.docx")
	if err == nil {
		t.Fatal("重定向到非允许 host 必须被拒绝")
	}

	data, _ := os.ReadFile(doc.Filepath)
	if string(data) != "orig" {
		t.Errorf("重定向场景下原文件不得被覆盖: got %q", string(data))
	}
}

// TestSaveDocument_SSRF_AllowedHostSucceeds 配置匹配的 URL 必须正常保存
func TestSaveDocument_SSRF_AllowedHostSucceeds(t *testing.T) {
	db := setupOnlyOfficeTestDB(t)
	doc := &models.Document{
		Name: "ok", Filename: "f.docx", Filepath: filepath.Join(t.TempDir(), "f.docx"),
		Filesize: 10, Status: "active",
	}
	db.Create(doc)
	os.WriteFile(doc.Filepath, []byte("orig"), 0o644)

	ts := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write([]byte("new content"))
	}))
	defer ts.Close()

	h := setupOnlyOfficeHandlerWithServer(t, db, ts.URL)
	h.httpClient = ts.Client()

	if err := h.saveDocument(context.Background(), doc.ID, ts.URL+"/file.docx"); err != nil {
		t.Fatalf("允许 host 必须成功: %v", err)
	}

	data, _ := os.ReadFile(doc.Filepath)
	if string(data) != "new content" {
		t.Errorf("文件内容 = %q, 期望 %q", string(data), "new content")
	}
}

func TestProcessConversionResult_RejectsNonOnlyOfficeURL(t *testing.T) {
	db := setupOnlyOfficeTestDB(t)
	tsHit := false
	ts := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tsHit = true
		w.Write([]byte("converted"))
	}))
	defer ts.Close()

	h := setupOnlyOfficeHandlerWithServer(t, db, "http://onlyoffice.example:9090")
	h.httpClient = ts.Client()
	h.conversionTasks["task-1"] = &ConversionTask{
		TaskID:     "task-1",
		DocumentID: 1,
		Status:     ConversionStatusPending,
		OutputType: "pdf",
		CreatedAt:  time.Now(),
	}

	h.processConversionResult("task-1", &ConversionResponse{Status: 3, URL: ts.URL + "/converted.pdf"})

	task := h.conversionTasks["task-1"]
	if task.Status != ConversionStatusError {
		t.Fatalf("非 OnlyOffice host 的转换下载必须失败: status=%s error=%s", task.Status, task.Error)
	}
	if tsHit {
		t.Fatal("SSRF 校验失败时不应发起 HTTP 请求")
	}
}

// TestGetUserGroup 测试用户组判断
func TestGetUserGroup(t *testing.T) {
	h := &OnlyOfficeHandler{}

	tests := []struct {
		role string
		want string
	}{
		{"admin", "admin"},
		{"lawyer", "lawyer"},
		{"partner", "lawyer"},
		{"assistant", "assistant"},
		{"unknown", "user"},
	}

	for _, tt := range tests {
		user := &models.User{Role: tt.role}
		got := h.getUserGroup(user)
		if got != tt.want {
			t.Errorf("getUserGroup(%q) = %q, want %q", tt.role, got, tt.want)
		}
	}
}
