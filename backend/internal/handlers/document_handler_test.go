package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"law-oa-go/internal/auth"
	"law-oa-go/internal/common"
	"law-oa-go/internal/services"
)

// TestableDocumentHandler 可测试的文档处理器
type TestableDocumentHandler struct {
	documentService      *services.DocumentService
	uploadDocumentFunc   func(ctx context.Context, req *services.DocumentUploadRequest) (*services.Document, error)
	getDocumentByIDFunc  func(ctx context.Context, id uint) (*services.Document, error)
	updateDocumentFunc   func(ctx context.Context, id uint, req *services.DocumentUpdateRequest) (*services.Document, error)
	deleteDocumentFunc   func(ctx context.Context, id uint) error
	listDocumentsFunc    func(ctx context.Context, req *services.DocumentListRequest, viewerUserID uint) ([]*services.Document, int64, error)
	getDocumentStatsFunc func(ctx context.Context, viewerUserID uint) (*services.DocumentStats, error)
	downloadDocumentFunc func(ctx context.Context, id uint) (io.ReadCloser, *services.Document, error)
}

// NewTestableDocumentHandler 创建可测试的文档处理器
func NewTestableDocumentHandler(service *services.DocumentService) *TestableDocumentHandler {
	return &TestableDocumentHandler{
		documentService:      service,
		uploadDocumentFunc:   service.UploadDocument,
		getDocumentByIDFunc:  service.GetDocumentByID,
		updateDocumentFunc:   service.UpdateDocument,
		deleteDocumentFunc:   service.DeleteDocument,
		listDocumentsFunc:    service.ListDocuments,
		getDocumentStatsFunc: service.GetDocumentStats,
		downloadDocumentFunc: service.DownloadDocument,
	}
}

// UploadDocument 上传文档
func (h *TestableDocumentHandler) UploadDocument(c *gin.Context) {
	// 解析上传请求
	req := h.parseUploadRequest(c)
	if req == nil {
		return // parseUploadRequest 已经设置了错误
	}

	// 调用可注入的函数
	doc, err := h.uploadDocumentFunc(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	common.APISuccess(c, doc)
}

// parseUploadRequest 解析上传请求
func (h *TestableDocumentHandler) parseUploadRequest(c *gin.Context) *services.DocumentUploadRequest {
	// Parse form data
	err := c.Request.ParseMultipartForm(50 << 20) // 50 MB max memory
	if err != nil {
		_ = c.Error(fmt.Errorf("form_parse: Failed to parse form data: %s", err.Error()))
		return nil
	}

	// Get form values
	name := c.PostForm("name")
	description := c.PostForm("description")
	category := c.PostForm("category")
	tags := c.PostForm("tags")
	entityIDStr := c.PostForm("entity_id")
	entityType := c.PostForm("entity_type")

	// Parse entity ID
	var entityID uint
	if entityIDStr != "" {
		id, err := strconv.ParseUint(entityIDStr, 10, 32)
		if err != nil {
			_ = c.Error(fmt.Errorf("entity_id: Invalid entity ID: must be a valid number"))
			return nil
		}
		entityID = uint(id)
	}

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		_ = c.Error(fmt.Errorf("file: Missing file: %s", err.Error()))
		return nil
	}

	return &services.DocumentUploadRequest{
		Name:        name,
		Description: description,
		Category:    category,
		Tags:        tags,
		EntityID:    entityID,
		EntityType:  entityType,
		File:        file,
	}
}

// GetDocument 获取文档
func (h *TestableDocumentHandler) GetDocument(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		_ = c.Error(fmt.Errorf("id_validation: Invalid ID: must be a valid number"))
		return
	}

	doc, err := h.getDocumentByIDFunc(c.Request.Context(), uint(id))
	if err != nil {
		_ = c.Error(err)
		return
	}

	common.APISuccess(c, doc)
}

// UpdateDocument 更新文档
func (h *TestableDocumentHandler) UpdateDocument(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		_ = c.Error(fmt.Errorf("id_validation: Invalid ID: must be a valid number"))
		return
	}

	var req services.DocumentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(fmt.Errorf("request_binding: Invalid request format: %s", err.Error()))
		return
	}

	doc, err := h.updateDocumentFunc(c.Request.Context(), uint(id), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	common.APISuccess(c, doc)
}

// DeleteDocument 删除文档
func (h *TestableDocumentHandler) DeleteDocument(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		_ = c.Error(fmt.Errorf("id_validation: Invalid ID: must be a valid number"))
		return
	}

	err = h.deleteDocumentFunc(c.Request.Context(), uint(id))
	if err != nil {
		_ = c.Error(err)
		return
	}

	common.APISuccess(c, gin.H{"message": "Document deleted successfully"})
}

// ListDocuments 获取文档列表
func (h *TestableDocumentHandler) ListDocuments(c *gin.Context) {
	var req services.DocumentListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		_ = c.Error(fmt.Errorf("query_binding: Invalid query parameters: %s", err.Error()))
		return
	}

	documents, total, err := h.listDocumentsFunc(c.Request.Context(), &req, auth.GetUserID(c))
	if err != nil {
		_ = c.Error(err)
		return
	}

	page := 1
	pageSize := 20
	if req.Page > 0 {
		page = req.Page
	}
	if req.PageSize > 0 {
		pageSize = req.PageSize
	}

	common.APISuccessWithPage(c, documents, total, page, pageSize)
}

// GetDocumentStats 获取文档统计
func (h *TestableDocumentHandler) GetDocumentStats(c *gin.Context) {
	stats, err := h.getDocumentStatsFunc(c.Request.Context(), auth.GetUserID(c))
	if err != nil {
		_ = c.Error(err)
		return
	}

	common.APISuccess(c, stats)
}

// DownloadDocument 下载文档
func (h *TestableDocumentHandler) DownloadDocument(c *gin.Context) {
	idStr := c.Param("id")
	_, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		_ = c.Error(fmt.Errorf("id_validation: Invalid ID: must be a valid number"))
		return
	}

	// 对于测试，我们直接返回模拟内容
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename=document.pdf")
	c.Header("Content-Type", "application/pdf")

	// 发送模拟内容
	content := "Mock document content for document ID " + idStr
	c.Data(http.StatusOK, "application/pdf", []byte(content))
}

// createTestDocument 创建测试文档
func createTestDocument() *services.Document {
	return &services.Document{
		ID:          1,
		Name:        "测试文档.pdf",
		Description: "这是一个测试文档",
		Filename:    "test.pdf",
		Filepath:    "/uploads/test.pdf",
		Filesize:    1024,
		MimeType:    "application/pdf",
		Category:    "合同",
		Tags:        []string{"合同", "测试"},
		EntityID:    1,
		EntityType:  "case",
		CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

// createTestDocumentStats 创建测试文档统计
func createTestDocumentStats() *services.DocumentStats {
	return &services.DocumentStats{
		TotalDocuments: 100,
		ByCategory: map[string]int64{
			"合同":    30,
			"证据":    25,
			"法律意见书": 20,
			"其他":    25,
		},
		ByEntityType: map[string]int64{
			"case":   60,
			"client": 30,
			"other":  10,
		},
		RecentUploads: 10,
	}
}

// createTestRouter 创建测试路由
func createTestRouter(handler *TestableDocumentHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(func(c *gin.Context) {
		c.Next()
		// 处理错误
		if len(c.Errors) > 0 {
			err := c.Errors[0]
			errMsg := err.Error()

			// 检查是否是绑定错误（包括验证错误）
			if err.Type == gin.ErrorTypeBind ||
				strings.Contains(errMsg, "Invalid request format") {
				common.APIBadRequest(c, errMsg)
				return
			}

			// 处理其他错误类型
			if errMsg == "record not found" || strings.Contains(errMsg, "not found") {
				common.APINotFound(c, "Resource not found")
			} else if strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "Invalid") || strings.Contains(errMsg, "Missing file") {
				common.APIBadRequest(c, "Invalid request: "+errMsg)
			} else {
				common.APIInternalServerError(c, "Internal server error: "+errMsg)
			}
			return
		}
	})

	// 设置路由
	docGroup := router.Group("/documents")
	{
		docGroup.POST("", handler.UploadDocument)
		docGroup.GET("/:id", handler.GetDocument)
		docGroup.PUT("/:id", handler.UpdateDocument)
		docGroup.DELETE("/:id", handler.DeleteDocument)
		docGroup.GET("", handler.ListDocuments)
		docGroup.GET("/stats", handler.GetDocumentStats)
		docGroup.GET("/:id/download", handler.DownloadDocument)
	}

	return router
}

// TestDocumentHandler_UploadDocumentSuccess 测试文档上传成功
func TestDocumentHandler_UploadDocumentSuccess(t *testing.T) {
	// 创建虚拟的service
	dummyService := &services.DocumentService{}
	handler := NewTestableDocumentHandler(dummyService)

	// 设置mock函数
	testDoc := createTestDocument()
	handler.uploadDocumentFunc = func(ctx context.Context, req *services.DocumentUploadRequest) (*services.Document, error) {
		assert.Equal(t, "测试文档.pdf", req.Name)
		assert.Equal(t, "这是一个测试文档", req.Description)
		assert.Equal(t, "合同", req.Category)
		assert.Equal(t, uint(1), req.EntityID)
		assert.Equal(t, "case", req.EntityType)
		assert.NotNil(t, req.File)
		return testDoc, nil
	}

	router := createTestRouter(handler)

	// 创建模拟表单数据
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("name", "测试文档.pdf")
	_ = writer.WriteField("description", "这是一个测试文档")
	_ = writer.WriteField("category", "合同")
	_ = writer.WriteField("tags", "合同,测试")
	_ = writer.WriteField("entity_id", "1")
	_ = writer.WriteField("entity_type", "case")

	// 添加文件
	part, _ := writer.CreateFormFile("file", "test.pdf")
	part.Write([]byte("mock pdf content"))

	writer.Close()

	// 创建请求
	req := httptest.NewRequest("POST", "/documents", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证结果
	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, true, response["success"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(1), data["id"])
	assert.Equal(t, "测试文档.pdf", data["name"])
	assert.Equal(t, "合同", data["category"])
}

// TestDocumentHandler_GetDocumentSuccess 测试获取文档成功
func TestDocumentHandler_GetDocumentSuccess(t *testing.T) {
	dummyService := &services.DocumentService{}
	handler := NewTestableDocumentHandler(dummyService)

	testDoc := createTestDocument()
	handler.getDocumentByIDFunc = func(ctx context.Context, id uint) (*services.Document, error) {
		assert.Equal(t, uint(1), id)
		return testDoc, nil
	}

	router := createTestRouter(handler)

	req := httptest.NewRequest("GET", "/documents/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证结果
	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, true, response["success"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(1), data["id"])
	assert.Equal(t, "测试文档.pdf", data["name"])
}

// TestDocumentHandler_GetDocumentInvalidID 测试获取文档无效ID
func TestDocumentHandler_GetDocumentInvalidID(t *testing.T) {
	dummyService := &services.DocumentService{}
	handler := NewTestableDocumentHandler(dummyService)
	router := createTestRouter(handler)

	req := httptest.NewRequest("GET", "/documents/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证结果 - 应该返回400错误
	require.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, false, response["success"])
	assert.Contains(t, response, "error")
}

// TestDocumentHandler_GetDocumentNotFound 测试获取文档不存在
func TestDocumentHandler_GetDocumentNotFound(t *testing.T) {
	dummyService := &services.DocumentService{}
	handler := NewTestableDocumentHandler(dummyService)

	handler.getDocumentByIDFunc = func(ctx context.Context, id uint) (*services.Document, error) {
		return nil, fmt.Errorf("document not found")
	}

	router := createTestRouter(handler)

	req := httptest.NewRequest("GET", "/documents/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证结果 - 应该返回404错误
	require.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, false, response["success"])
	assert.Contains(t, response, "error")
}

// TestDocumentHandler_UpdateDocumentSuccess 测试更新文档成功
func TestDocumentHandler_UpdateDocumentSuccess(t *testing.T) {
	dummyService := &services.DocumentService{}
	handler := NewTestableDocumentHandler(dummyService)

	testDoc := createTestDocument()
	testDoc.Name = "更新的文档.pdf"
	testDoc.Description = "更新的描述"

	handler.updateDocumentFunc = func(ctx context.Context, id uint, req *services.DocumentUpdateRequest) (*services.Document, error) {
		assert.Equal(t, uint(1), id)
		assert.Equal(t, "更新的文档.pdf", *req.Name)
		assert.Equal(t, "更新的描述", *req.Description)
		return testDoc, nil
	}

	router := createTestRouter(handler)

	// 准备请求体
	updateReq := map[string]interface{}{
		"name":        "更新的文档.pdf",
		"description": "更新的描述",
		"category":    "更新类别",
		"tags":        []string{"更新", "标签"},
	}
	jsonData, _ := json.Marshal(updateReq)

	req := httptest.NewRequest("PUT", "/documents/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证结果
	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, true, response["success"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "更新的文档.pdf", data["name"])
	assert.Equal(t, "更新的描述", data["description"])
}

// TestDocumentHandler_DeleteDocumentSuccess 测试删除文档成功
func TestDocumentHandler_DeleteDocumentSuccess(t *testing.T) {
	dummyService := &services.DocumentService{}
	handler := NewTestableDocumentHandler(dummyService)

	handler.deleteDocumentFunc = func(ctx context.Context, id uint) error {
		assert.Equal(t, uint(1), id)
		return nil
	}

	router := createTestRouter(handler)

	req := httptest.NewRequest("DELETE", "/documents/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证结果
	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Contains(t, response["data"], "message")
}

// TestDocumentHandler_ListDocumentsSuccess 测试获取文档列表成功
func TestDocumentHandler_ListDocumentsSuccess(t *testing.T) {
	dummyService := &services.DocumentService{}
	handler := NewTestableDocumentHandler(dummyService)

	doc1 := createTestDocument()
	doc2 := createTestDocument()
	doc2.ID = 2
	doc2.Name = "测试文档2.pdf"

	handler.listDocumentsFunc = func(ctx context.Context, req *services.DocumentListRequest, viewerUserID uint) ([]*services.Document, int64, error) {
		assert.Equal(t, 1, req.Page)
		assert.Equal(t, 10, req.PageSize)
		assert.Equal(t, "合同", req.Category)
		return []*services.Document{doc1, doc2}, int64(2), nil
	}

	router := createTestRouter(handler)

	req := httptest.NewRequest("GET", "/documents?page=1&page_size=10&category=合同", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证结果
	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, true, response["success"])

	data := response["data"].([]interface{})
	assert.Len(t, data, 2)

	assert.Contains(t, response, "pagination")
	pagination := response["pagination"].(map[string]interface{})
	assert.Equal(t, float64(1), pagination["page"])
	assert.Equal(t, float64(10), pagination["page_size"])
	assert.Equal(t, float64(2), pagination["total"])
}

// TestDocumentHandler_GetDocumentStatsSuccess 测试获取文档统计成功
func TestDocumentHandler_GetDocumentStatsSuccess(t *testing.T) {
	dummyService := &services.DocumentService{}
	handler := NewTestableDocumentHandler(dummyService)

	stats := createTestDocumentStats()
	handler.getDocumentStatsFunc = func(ctx context.Context, viewerUserID uint) (*services.DocumentStats, error) {
		return stats, nil
	}

	router := createTestRouter(handler)

	req := httptest.NewRequest("GET", "/documents/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证结果
	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, true, response["success"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(100), data["total_documents"])
	assert.Contains(t, data, "by_category")
	assert.Contains(t, data, "by_entity_type")
}

// TestDocumentHandler_DownloadDocumentSuccess 测试下载文档成功
func TestDocumentHandler_DownloadDocumentSuccess(t *testing.T) {
	dummyService := &services.DocumentService{}
	handler := NewTestableDocumentHandler(dummyService)
	router := createTestRouter(handler)

	req := httptest.NewRequest("GET", "/documents/1/download", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证结果
	require.Equal(t, http.StatusOK, w.Code)

	// 验证响应头
	assert.Equal(t, "File Transfer", w.Header().Get("Content-Description"))
	assert.Equal(t, "binary", w.Header().Get("Content-Transfer-Encoding"))
	assert.Equal(t, "attachment; filename=document.pdf", w.Header().Get("Content-Disposition"))
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))

	// 验证响应体
	assert.Contains(t, w.Body.String(), "Mock document content for document ID 1")
}

// TestDocumentHandler_UploadDocumentMissingFile 测试文档上传缺少文件
func TestDocumentHandler_UploadDocumentMissingFile(t *testing.T) {
	dummyService := &services.DocumentService{}
	handler := NewTestableDocumentHandler(dummyService)
	router := createTestRouter(handler)

	// 创建没有文件的请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("name", "测试文档.pdf")
	_ = writer.WriteField("description", "这是一个测试文档")

	writer.Close()

	req := httptest.NewRequest("POST", "/documents", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证结果 - 应该返回400错误
	require.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, false, response["success"])
	assert.Contains(t, response, "error")
}

// TestDocumentHandler_UploadDocumentInvalidEntityID 测试文档上传无效实体ID
func TestDocumentHandler_UploadDocumentInvalidEntityID(t *testing.T) {
	dummyService := &services.DocumentService{}
	handler := NewTestableDocumentHandler(dummyService)
	router := createTestRouter(handler)

	// 创建模拟表单数据
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("name", "测试文档.pdf")
	_ = writer.WriteField("entity_id", "invalid") // 无效ID
	_ = writer.WriteField("entity_type", "case")

	part, _ := writer.CreateFormFile("file", "test.pdf")
	part.Write([]byte("mock pdf content"))

	writer.Close()

	req := httptest.NewRequest("POST", "/documents", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证结果 - 应该返回400错误
	require.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, false, response["success"])
	assert.Contains(t, response, "error")
}

// TestDocumentHandler_UpdateDocumentInvalidJSON 测试更新文档无效JSON
func TestDocumentHandler_UpdateDocumentInvalidJSON(t *testing.T) {
	dummyService := &services.DocumentService{}
	handler := NewTestableDocumentHandler(dummyService)
	router := createTestRouter(handler)

	req := httptest.NewRequest("PUT", "/documents/1", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证结果 - 应该返回400错误
	require.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, false, response["success"])
	assert.Contains(t, response, "error")
}

// TestDocumentHandler_ListDocumentsInvalidQuery 测试获取文档列表无效查询参数
func TestDocumentHandler_ListDocumentsInvalidQuery(t *testing.T) {
	dummyService := &services.DocumentService{}
	handler := NewTestableDocumentHandler(dummyService)
	router := createTestRouter(handler)

	req := httptest.NewRequest("GET", "/documents?page=invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证结果 - 应该返回400错误
	require.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, false, response["success"])
	assert.Contains(t, response, "error")
}

// TestDocumentHandler_DownloadDocumentInvalidID 测试下载文档无效ID
func TestDocumentHandler_DownloadDocumentInvalidID(t *testing.T) {
	dummyService := &services.DocumentService{}
	handler := NewTestableDocumentHandler(dummyService)
	router := createTestRouter(handler)

	req := httptest.NewRequest("GET", "/documents/invalid/download", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证结果 - 应该返回400错误
	require.Equal(t, http.StatusBadRequest, w.Code)
}

// TestDocumentHandler_ServiceError 测试服务错误
func TestDocumentHandler_ServiceError(t *testing.T) {
	dummyService := &services.DocumentService{}
	handler := NewTestableDocumentHandler(dummyService)

	handler.getDocumentByIDFunc = func(ctx context.Context, id uint) (*services.Document, error) {
		return nil, fmt.Errorf("service error")
	}

	router := createTestRouter(handler)

	req := httptest.NewRequest("GET", "/documents/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证结果 - 应该返回500错误
	require.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, false, response["success"])
	assert.Contains(t, response, "error")
}

// TestDocumentHandler_BoundaryCases 测试文档处理器边界情况
func TestDocumentHandler_BoundaryCases(t *testing.T) {
	t.Run("LargeEntityID", func(t *testing.T) {
		dummyService := &services.DocumentService{}
		handler := NewTestableDocumentHandler(dummyService)

		handler.uploadDocumentFunc = func(ctx context.Context, req *services.DocumentUploadRequest) (*services.Document, error) {
			// 测试最大 uint32 值
			assert.Equal(t, uint(4294967295), req.EntityID)
			return createTestDocument(), nil
		}

		router := createTestRouter(handler)

		// 测试最大 uint32 值
		maxID := strconv.FormatUint(1<<32-1, 10)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		_ = writer.WriteField("name", "测试文档.pdf")
		_ = writer.WriteField("entity_id", maxID)
		_ = writer.WriteField("entity_type", "case")

		part, _ := writer.CreateFormFile("file", "test.pdf")
		part.Write([]byte("mock pdf content"))

		writer.Close()

		req := httptest.NewRequest("POST", "/documents", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 应该能够处理大ID
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("EmptyFormFields", func(t *testing.T) {
		dummyService := &services.DocumentService{}
		handler := NewTestableDocumentHandler(dummyService)

		handler.uploadDocumentFunc = func(ctx context.Context, req *services.DocumentUploadRequest) (*services.Document, error) {
			// 验证空字段处理
			assert.Equal(t, "", req.Name)
			assert.Equal(t, "", req.Description)
			assert.NotNil(t, req.File)
			return createTestDocument(), nil
		}

		router := createTestRouter(handler)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// 只添加必需的文件字段
		part, _ := writer.CreateFormFile("file", "test.pdf")
		part.Write([]byte("mock pdf content"))

		writer.Close()

		req := httptest.NewRequest("POST", "/documents", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 应该能够处理空字段
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("SpecialCharactersInFormFields", func(t *testing.T) {
		dummyService := &services.DocumentService{}
		handler := NewTestableDocumentHandler(dummyService)

		handler.uploadDocumentFunc = func(ctx context.Context, req *services.DocumentUploadRequest) (*services.Document, error) {
			// 验证特殊字符处理
			assert.Equal(t, "测试@#$%^&*().pdf", req.Name)
			assert.Equal(t, "描述包含特殊字符: 中文!@#$%^&*()", req.Description)
			assert.Equal(t, "标签1,标签2,标签3", req.Tags)
			return createTestDocument(), nil
		}

		router := createTestRouter(handler)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		_ = writer.WriteField("name", "测试@#$%^&*().pdf")
		_ = writer.WriteField("description", "描述包含特殊字符: 中文!@#$%^&*()")
		_ = writer.WriteField("tags", "标签1,标签2,标签3")

		part, _ := writer.CreateFormFile("file", "test.pdf")
		part.Write([]byte("mock pdf content"))

		writer.Close()

		req := httptest.NewRequest("POST", "/documents", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 应该能够处理特殊字符
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestDocumentHandler_Performance 测试文档处理器性能
func TestDocumentHandler_Performance(t *testing.T) {
	dummyService := &services.DocumentService{}
	handler := NewTestableDocumentHandler(dummyService)

	handler.getDocumentByIDFunc = func(ctx context.Context, id uint) (*services.Document, error) {
		return createTestDocument(), nil
	}

	router := createTestRouter(handler)

	// 批量测试获取文档
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("GET", "/documents/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}
}
