package handlers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"law-oa-go/internal/middleware"
	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
)

// OnlyOfficeHandler OnlyOffice 在线编辑处理器
type OnlyOfficeHandler struct {
	db                *gorm.DB
	versionService    *services.DocumentVersionService
	lockService       *services.DocumentLockService
	onlyOfficeURL     string
	onlyOfficeSecret  string
	backendURL        string
	httpClient        *http.Client
	storageDir        string
	conversionTasks   map[string]*ConversionTask
	conversionMu      sync.RWMutex
}

// ConversionTask 转换任务
type ConversionTask struct {
	TaskID      string    `json:"task_id"`
	DocumentID  uint      `json:"document_id"`
	Status      string    `json:"status"` // pending, processing, completed, error
	OutputType  string    `json:"output_type"`
	OutputURL   string    `json:"output_url,omitempty"`
	OutputPath  string    `json:"output_path,omitempty"`
	Error       string    `json:"error,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
}

// NewOnlyOfficeHandler 创建 OnlyOffice 处理器
func NewOnlyOfficeHandler(
	db *gorm.DB,
	versionService *services.DocumentVersionService,
	lockService *services.DocumentLockService,
	onlyOfficeURL string,
	onlyOfficeSecret string,
	backendURL string,
	storageDir string,
) *OnlyOfficeHandler {
	return &OnlyOfficeHandler{
		db:               db,
		versionService:   versionService,
		lockService:      lockService,
		onlyOfficeURL:    strings.TrimRight(onlyOfficeURL, "/"),
		onlyOfficeSecret: onlyOfficeSecret,
		backendURL:       strings.TrimRight(backendURL, "/"),
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		storageDir:      storageDir,
		conversionTasks: make(map[string]*ConversionTask),
	}
}

// EditorConfig OnlyOffice 编辑器配置
type EditorConfig struct {
	Document     DocumentConfig     `json:"document"`
	DocumentType string             `json:"documentType"`
	EditorConfig EditorConfigInner `json:"editorConfig"`
}

// DocumentConfig 文档配置
type DocumentConfig struct {
	Key       string       `json:"key"`
	URL       string       `json:"url"`
	Title     string       `json:"title"`
	FileType  string       `json:"fileType"`
	Info      DocumentInfo `json:"info"`
	Permissions Permissions `json:"permissions"`
}

// DocumentInfo 文档信息
type DocumentInfo struct {
	Author     string `json:"author"`
	Created    string `json:"created"`
	Owner      string `json:"owner"`
	Uploaded   string `json:"uploaded"`
}

// Permissions 权限配置
type Permissions struct {
	Comment      bool `json:"comment"`
	Download     bool `json:"download"`
	Edit         bool `json:"edit"`
	FillForms    bool `json:"fillForms"`
	ModifyFilter bool `json:"modifyFilter"`
	ModifyContentControl bool `json:"modifyContentControl"`
	Review       bool `json:"review"`
}

// EditorConfigInner 编辑器配置
type EditorConfigInner struct {
	Mode             string             `json:"mode"`
	CallbackURL      string             `json:"callbackUrl"`
	Customization    Customization      `json:"customization"`
	User             User               `json:"user"`
	Embedded         Embedded           `json:"embedded"`
}

// Customization 自定义配置
type Customization struct {
	About             bool              `json:"about"`
	Comments          bool              `json:"comments"`
	Customer          map[string]string `json:"customer"`
	Feedback          bool              `json:"feedback"`
	Forcesave         bool              `json:"forcesave"`
	Help              bool              `json:"help"`
	Macro             bool              `json:"macro"`
	Metrics           bool              `json:"metrics"`
	Plugins           bool              `json:"plugins"`
	CompactHeader     bool              `json:"compactHeader"`
	CompactToolbar    bool              `json:"compactToolbar"`
	Customization     bool              `json:"customization"`
	Zoom              int               `json:"zoom"`
	ToolbarNoTabs     bool              `json:"toolbarNoTabs"`
	ToolbarHideFileName bool `json:"toolbarHideFileName"`
}

// User 用户信息
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Group string `json:"group"`
}

// Embedded 嵌入配置
type Embedded struct {
	SaveURL       string `json:"saveUrl"`
	ShareURL      string `json:"shareUrl"`
	EmbedURL      string `json:"embedUrl"`
	ToolbarDocked string `json:"toolbarDocked"`
}

// OpenEditorRequest 打开编辑器请求。
// 注意：UserID 不再从请求体读取，强制使用 middleware.GetCurrentUserID(c)。
// 客户端若仍提交 user_id 字段，将被忽略以避免越权（例如假装是其他用户）。
type OpenEditorRequest struct {
	DocumentID uint   `json:"document_id" binding:"required"`
	UserID     uint   `json:"user_id"` // 已废弃，保留字段以向后兼容客户端；handler 会覆盖为上下文用户
	Mode       string `json:"mode"`    // edit 或 view
}

// CallbackRequest OnlyOffice 回调请求
type CallbackRequest struct {
	Actions   []CallbackAction `json:"actions"`
	Key       string           `json:"key"`
	Status    int              `json:"status"`
	URL       string           `json:"url"`
	Users     []string         `json:"users"`
	Token     string           `json:"token"`
}

// CallbackAction 回调操作
type CallbackAction struct {
	Type   string `json:"type"`
	UserID string `json:"userid"`
}

// ConversionRequest OnlyOffice 转换 API 请求
type ConversionRequest struct {
	Async    bool   `json:"async"`
	Filetype string `json:"filetype"`
	Key      string `json:"key"`
	OutputType string `json:"outputType"`
	Title    string `json:"title"`
	URL      string `json:"url"`
}

// ConversionResponse OnlyOffice 转换 API 响应
type ConversionResponse struct {
	EndKey    string `json:"endKey"`
	FileType  string `json:"fileType"`
	FileSize  int64  `json:"fileSize"`
	Key       string `json:"key"`
	Percent   int    `json:"percent"`
	Status    int    `json:"status"` // 0-unknown, 1-queued, 2-processing, 3-converted, 4-converting error, 5-error, 6-async
	URL       string `json:"url"`
}

// 常量
const (
	// 文档状态
	StatusEditing       = 1 // 正在编辑
	StatusMustSave      = 2 // 必须保存
	StatusCorrupted     = 3 // 文档损坏
	StatusClosed        = 6 // 已关闭
	StatusMustForceSave = 7 // 强制保存

	// 文档类型
	DocumentTypeWord  = "word"
	DocumentTypeCell  = "cell"
	DocumentTypeSlide = "slide"

	// 编辑模式
	ModeEdit = "edit"
	ModeView = "view"

	// 转换状态
	ConversionStatusPending    = "pending"
	ConversionStatusProcessing = "processing"
	ConversionStatusCompleted  = "completed"
	ConversionStatusError      = "error"
)

// OpenEditor 打开文档编辑器
func (h *OnlyOfficeHandler) OpenEditor(c *gin.Context) {
	var req OpenEditorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// 安全：以认证上下文中的用户 ID 为准，忽略请求体里的 user_id
	viewerUserID, exists := middleware.GetCurrentUserID(c)
	if !exists || viewerUserID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}
	req.UserID = viewerUserID

	// 获取文档
	var doc models.Document
	if err := h.db.WithContext(ctx).First(&doc, req.DocumentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	// 获取用户信息
	var user models.User
	if err := h.db.WithContext(ctx).First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 确定编辑模式
	mode := req.Mode
	if mode == "" {
		mode = ModeEdit
	}

	// 检查是否需要获取锁
	if mode == ModeEdit {
		lockReq := &services.AcquireLockRequest{
			DocumentID: req.DocumentID,
			UserID:     req.UserID,
			UserName:   user.Name,
			IPAddress:  c.ClientIP(),
			UserAgent:  c.GetHeader("User-Agent"),
			IsCheckout: false,
		}

		lockStatus, err := h.lockService.AcquireLock(ctx, lockReq)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("Failed to acquire lock: %v", err)})
			return
		}

		if !lockStatus.CanEdit {
			c.JSON(http.StatusLocked, gin.H{
				"error":     "Document is locked by another user",
				"locked_by": lockStatus.LockedByName,
				"locked_at": lockStatus.LockedAt,
			})
			return
		}
	}

	// 生成文档 URL
	docURL := fmt.Sprintf("%s/api/documents/%d/download", h.backendURL, req.DocumentID)

	// 生成回调 URL
	callbackURL := fmt.Sprintf("%s/api/documents/onlyoffice/callback", h.backendURL)

	// 生成文档 key (用于唯一标识文档编辑会话)
	key := h.generateDocKey(req.DocumentID, doc.UpdatedAt)

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(doc.Filename))

	// 确定文档类型
	docType := h.getDocumentType(ext)

	// 获取创建者信息
	creatorName := "Unknown"
	if doc.EntityID != 0 {
		var creator models.User
		if err := h.db.WithContext(ctx).Select("name").First(&creator, doc.EntityID).Error; err == nil {
			creatorName = creator.Name
		}
	}

	// 构建编辑器配置
	config := EditorConfig{
		Document: DocumentConfig{
			Key:      key,
			URL:      docURL,
			Title:    doc.Name,
			FileType: strings.TrimPrefix(ext, "."),
			Info: DocumentInfo{
				Author:   creatorName,
				Created:  doc.CreatedAt.Format("2006-01-02 15:04:05"),
				Owner:    creatorName,
				Uploaded: doc.CreatedAt.Format("2006-01-02 15:04:05"),
			},
			Permissions: Permissions{
				Comment:               mode == ModeEdit,
				Download:              true,
				Edit:                  mode == ModeEdit,
				FillForms:             mode == ModeEdit,
				ModifyFilter:          true,
				ModifyContentControl:  true,
				Review:                mode == ModeEdit,
			},
		},
		DocumentType: docType,
		EditorConfig: EditorConfigInner{
			Mode:        mode,
			CallbackURL: callbackURL,
			Customization: Customization{
				About:              false,
				Comments:           true,
				Customer:           map[string]string{"info": "Law OA Go"},
				Feedback:           false,
				Forcesave:          true,
				Help:               false,
				Macro:              false,
				Metrics:            false,
				Plugins:            false,
				CompactHeader:      false,
				CompactToolbar:     false,
				Customization:      false,
				Zoom:               100,
				ToolbarNoTabs:      false,
				ToolbarHideFileName: false,
			},
			User: User{
				ID:    strconv.Itoa(int(req.UserID)),
				Name:  user.Name,
				Group: h.getUserGroup(&user),
			},
			Embedded: Embedded{
				// SaveURL 留空：OnlyOffice 使用 callback URL (POST) 触发保存，
				// 而非 Embedded.SaveURL（该字段要求独立 POST 端点且不携带编辑 token）
				SaveURL:       "",
				ToolbarDocked: "top",
			},
		},
	}

	c.JSON(http.StatusOK, config)
}

// HandleCallback 处理 OnlyOffice 回调
func (h *OnlyOfficeHandler) HandleCallback(c *gin.Context) {
	// 先读取 body 用于 HMAC 验证
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	var req CallbackRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证 token (如果配置了密钥，使用 HMAC 验证)
	if h.onlyOfficeSecret != "" {
		if !h.validateCallbackPayload(bodyBytes, req.Token) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
	}

	// 解析 key 获取文档 ID
	documentID, err := h.parseDocKey(req.Key)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document key"})
		return
	}

	ctx := c.Request.Context()

	switch req.Status {
	case StatusMustSave, StatusMustForceSave:
		if err := h.saveDocument(ctx, documentID, req.URL); err != nil {
			log.Printf("[OnlyOffice] 保存文档失败 (DocID: %d): %v", documentID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"error": 0})

	case StatusEditing:
		c.JSON(http.StatusOK, gin.H{"error": 0})

	case StatusClosed:
		if len(req.Users) > 0 {
			userID, _ := strconv.Atoi(req.Users[0])
			if userID > 0 {
				if err := h.lockService.ReleaseLock(ctx, &services.ReleaseLockRequest{
					DocumentID: documentID,
					UserID:     uint(userID),
					Force:      false,
				}); err != nil {
					log.Printf("[OnlyOffice] 警告: 释放文档锁失败 (DocID: %d, UserID: %d): %v", documentID, userID, err)
				}
			}
		}
		c.JSON(http.StatusOK, gin.H{"error": 0})

	default:
		c.JSON(http.StatusOK, gin.H{"error": 0})
	}
}

// saveDocument 从 OnlyOffice 下载保存的文档并更新存储
func (h *OnlyOfficeHandler) saveDocument(ctx context.Context, documentID uint, downloadURL string) error {
	if downloadURL == "" {
		return fmt.Errorf("download URL is empty")
	}

	// 获取文档记录
	var doc models.Document
	if err := h.db.WithContext(ctx).First(&doc, documentID).Error; err != nil {
		return fmt.Errorf("document not found: %w", err)
	}

	// 从 OnlyOffice 下载编辑后的文档
	resp, err := h.httpClient.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download from OnlyOffice: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("OnlyOffice returned status %d: %s", resp.StatusCode, string(body))
	}

	// 创建版本备份目录
	versionDir := filepath.Join(h.storageDir, "versions", fmt.Sprintf("doc_%d", documentID))
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		return fmt.Errorf("failed to create version directory: %w", err)
	}

	// 备份当前文件为版本
	if doc.Filepath != "" {
		if _, err := os.Stat(doc.Filepath); err == nil {
			timestamp := time.Now().Unix()
			backupFilename := fmt.Sprintf("v%d_%d_%s", timestamp, timestamp, filepath.Base(doc.Filepath))
			backupPath := filepath.Join(versionDir, backupFilename)
			if err := copyFile(doc.Filepath, backupPath); err != nil {
				log.Printf("[OnlyOffice] 警告: 版本备份失败 (DocID: %d): %v", documentID, err)
			}
		}
	}

	// 保存新文件（使用随机后缀避免并发覆盖）
	outFile, err := os.CreateTemp(filepath.Dir(doc.Filepath), filepath.Base(doc.Filepath)+".*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	written, err := io.Copy(outFile, resp.Body)
	outFile.Close()
	tmpPath := outFile.Name()

	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write document: %w", err)
	}

	// 原子替换
	if err := os.Rename(outFile.Name(), doc.Filepath); err != nil {
		os.Remove(outFile.Name())
		return fmt.Errorf("failed to replace document file: %w", err)
	}

	// 更新数据库记录
	now := time.Now()
	updates := map[string]interface{}{
		"filesize":   written,
		"updated_at": now,
	}
	if err := h.db.WithContext(ctx).Model(&models.Document{}).Where("id = ?", documentID).Updates(updates).Error; err != nil {
		log.Printf("[OnlyOffice] 警告: 更新文档记录失败 (DocID: %d): %v", documentID, err)
	}

	log.Printf("[OnlyOffice] 文档已保存 (DocID: %d, Size: %d bytes)", documentID, written)
	return nil
}

// ConvertDocument 转换文档格式
func (h *OnlyOfficeHandler) ConvertDocument(c *gin.Context) {
	var req struct {
		DocumentID uint   `json:"document_id" binding:"required"`
		OutputType string `json:"output_type" binding:"required"` // pdf, docx, xlsx, etc.
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// 获取文档
	var doc models.Document
	if err := h.db.WithContext(ctx).First(&doc, req.DocumentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	// 验证输出类型
	outputType := strings.ToLower(strings.TrimPrefix(req.OutputType, "."))
	if !h.isValidOutputType(outputType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Unsupported output type: %s", outputType)})
		return
	}

	// 生成文档下载 URL
	docURL := fmt.Sprintf("%s/api/documents/%d/download", h.backendURL, req.DocumentID)

	// 生成唯一 key
	key := fmt.Sprintf("convert_%d_%d", req.DocumentID, time.Now().Unix())

	// 构建转换请求
	convertReq := ConversionRequest{
		Async:      true,
		Filetype:   strings.TrimPrefix(strings.ToLower(filepath.Ext(doc.Filename)), "."),
		Key:        key,
		OutputType: outputType,
		Title:      doc.Name,
		URL:        docURL,
	}

	// 生成 token
	token := ""
	if h.onlyOfficeSecret != "" {
		payload, _ := json.Marshal(convertReq)
		token = h.GenerateCallbackToken(string(payload))
	}

	// 调用 OnlyOffice Conversion API
	convertURL := fmt.Sprintf("%s/CoAuthoring/CommandService.ashx", h.onlyOfficeURL)
	reqBody, _ := json.Marshal(convertReq)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", convertURL, strings.NewReader(string(reqBody)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create request: %v", err)})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Conversion request failed: %v", err)})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read conversion response"})
		return
	}

	var convertResp ConversionResponse
	if err := json.Unmarshal(body, &convertResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to parse conversion response: %v", err)})
		return
	}

	// 创建转换任务
	taskID := key
	task := &ConversionTask{
		TaskID:     taskID,
		DocumentID: req.DocumentID,
		Status:     ConversionStatusPending,
		OutputType: outputType,
		CreatedAt:  time.Now(),
	}

	h.conversionMu.Lock()
	h.conversionTasks[taskID] = task
	h.conversionMu.Unlock()

	// 如果 OnlyOffice 直接返回结果（同步）
	if convertResp.Status == 3 && convertResp.URL != "" {
		go h.processConversionResult(taskID, &convertResp)
	} else if convertResp.Status == 6 {
		// 异步模式，需要轮询
		go h.pollConversionStatus(taskID, key)
	} else if convertResp.Status >= 4 {
		task.Status = ConversionStatusError
		task.Error = fmt.Sprintf("Conversion failed with status %d", convertResp.Status)
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id":     taskID,
		"document_id": req.DocumentID,
		"status":      task.Status,
		"output_type": outputType,
	})
}

// CheckConversionStatus 检查转换状态
func (h *OnlyOfficeHandler) CheckConversionStatus(c *gin.Context) {
	taskID := c.Query("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task_id is required"})
		return
	}

	h.conversionMu.RLock()
	task, exists := h.conversionTasks[taskID]
	h.conversionMu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversion task not found"})
		return
	}

	response := gin.H{
		"task_id":     task.TaskID,
		"document_id": task.DocumentID,
		"status":      task.Status,
		"output_type": task.OutputType,
	}

	if task.Status == ConversionStatusCompleted {
		response["url"] = fmt.Sprintf("%s/api/documents/%d/download/converted/%s",
			h.backendURL, task.DocumentID, task.OutputType)
	}
	if task.Error != "" {
		response["error"] = task.Error
	}

	c.JSON(http.StatusOK, response)
}

// DownloadConverted 下载转换后的文档
func (h *OnlyOfficeHandler) DownloadConverted(c *gin.Context) {
	documentID, err := strconv.ParseUint(c.Param("document_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	outputType := c.Param("output_type")

	// 查找完成的转换任务
	h.conversionMu.RLock()
	var task *ConversionTask
	for _, t := range h.conversionTasks {
		if t.DocumentID == uint(documentID) && t.OutputType == outputType && t.Status == ConversionStatusCompleted {
			task = t
			break
		}
	}
	h.conversionMu.RUnlock()

	if task == nil || task.OutputPath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Converted file not found"})
		return
	}

	c.FileAttachment(task.OutputPath, fmt.Sprintf("document.%s", outputType))
}

// processConversionResult 处理转换结果（下载转换后的文件）
func (h *OnlyOfficeHandler) processConversionResult(taskID string, resp *ConversionResponse) {
	h.conversionMu.Lock()
	task, exists := h.conversionTasks[taskID]
	if !exists {
		h.conversionMu.Unlock()
		return
	}
	task.Status = ConversionStatusProcessing
	h.conversionMu.Unlock()

	// 下载转换后的文件
	httpResp, err := h.httpClient.Get(resp.URL)
	if err != nil {
		h.conversionMu.Lock()
		task.Status = ConversionStatusError
		task.Error = fmt.Sprintf("Failed to download converted file: %v", err)
		h.conversionMu.Unlock()
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		h.conversionMu.Lock()
		task.Status = ConversionStatusError
		task.Error = fmt.Sprintf("Download failed with status %d", httpResp.StatusCode)
		h.conversionMu.Unlock()
		return
	}

	// 保存转换后的文件
	convertDir := filepath.Join(h.storageDir, "converted", fmt.Sprintf("doc_%d", task.DocumentID))
	if err := os.MkdirAll(convertDir, 0755); err != nil {
		h.conversionMu.Lock()
		task.Status = ConversionStatusError
		task.Error = fmt.Sprintf("Failed to create directory: %v", err)
		h.conversionMu.Unlock()
		return
	}

	outputPath := filepath.Join(convertDir, fmt.Sprintf("document.%s", task.OutputType))
	outFile, err := os.Create(outputPath)
	if err != nil {
		h.conversionMu.Lock()
		task.Status = ConversionStatusError
		task.Error = fmt.Sprintf("Failed to create output file: %v", err)
		h.conversionMu.Unlock()
		return
	}

	_, err = io.Copy(outFile, httpResp.Body)
	outFile.Close()

	if err != nil {
		os.Remove(outputPath)
		h.conversionMu.Lock()
		task.Status = ConversionStatusError
		task.Error = fmt.Sprintf("Failed to save converted file: %v", err)
		h.conversionMu.Unlock()
		return
	}

	h.conversionMu.Lock()
	task.Status = ConversionStatusCompleted
	task.OutputPath = outputPath
	task.OutputURL = resp.URL
	task.CompletedAt = time.Now()
	h.conversionMu.Unlock()

	log.Printf("[OnlyOffice] 转换完成 (TaskID: %s, DocID: %d, Type: %s)", taskID, task.DocumentID, task.OutputType)
}

// pollConversionStatus 轮询异步转换状态
func (h *OnlyOfficeHandler) pollConversionStatus(taskID string, key string) {
	maxRetries := 30
	interval := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		time.Sleep(interval)

		checkURL := fmt.Sprintf("%s/CoAuthoring/CommandService.ashx?key=%s", h.onlyOfficeURL, key)
		resp, err := h.httpClient.Get(checkURL)
		if err != nil {
			continue
		}

		body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
		resp.Body.Close()
		if err != nil {
			continue
		}

		var result ConversionResponse
		if err := json.Unmarshal(body, &result); err != nil {
			continue
		}

		if result.Status == 3 && result.URL != "" {
			h.processConversionResult(taskID, &result)
			return
		}

		if result.Status >= 4 {
			h.conversionMu.Lock()
			if task, exists := h.conversionTasks[taskID]; exists {
				task.Status = ConversionStatusError
				task.Error = fmt.Sprintf("Conversion failed with status %d", result.Status)
			}
			h.conversionMu.Unlock()
			return
		}
	}

	// 超时
	h.conversionMu.Lock()
	if task, exists := h.conversionTasks[taskID]; exists {
		task.Status = ConversionStatusError
		task.Error = "Conversion timed out"
	}
	h.conversionMu.Unlock()
}

// isValidOutputType 验证输出类型是否支持
func (h *OnlyOfficeHandler) isValidOutputType(outputType string) bool {
	supported := map[string]bool{
		"pdf":  true,
		"docx": true,
		"doc":  true,
		"xlsx": true,
		"xls":  true,
		"pptx": true,
		"ppt":  true,
		"odt":  true,
		"ods":  true,
		"odp":  true,
		"txt":  true,
		"csv":  true,
		"rtf":  true,
	}
	return supported[outputType]
}

// generateDocKey 生成文档 key
func (h *OnlyOfficeHandler) generateDocKey(documentID uint, updatedAt time.Time) string {
	return fmt.Sprintf("%d_%d", documentID, updatedAt.Unix())
}

// parseDocKey 解析文档 key
func (h *OnlyOfficeHandler) parseDocKey(key string) (uint, error) {
	parts := strings.Split(key, "_")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid key format")
	}
	documentID, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	return uint(documentID), nil
}

// getDocumentType 获取文档类型
func (h *OnlyOfficeHandler) getDocumentType(ext string) string {
	switch ext {
	case ".doc", ".docx", ".odt", ".rtf", ".txt", ".html", ".htm", ".mht", ".pdf":
		return DocumentTypeWord
	case ".xls", ".xlsx", ".ods", ".csv":
		return DocumentTypeCell
	case ".ppt", ".pptx", ".odp":
		return DocumentTypeSlide
	default:
		return DocumentTypeWord
	}
}

// getUserGroup 获取用户组
func (h *OnlyOfficeHandler) getUserGroup(user *models.User) string {
	switch user.Role {
	case "admin":
		return "admin"
	case "lawyer", "partner":
		return "lawyer"
	case "assistant":
		return "assistant"
	default:
		return "user"
	}
}

// validateCallbackPayload 使用 HMAC 验证回调请求的完整性
func (h *OnlyOfficeHandler) validateCallbackPayload(payload []byte, token string) bool {
	if token == "" {
		return false
	}

	expectedMAC := h.GenerateCallbackToken(string(payload))
	return hmac.Equal([]byte(token), []byte(expectedMAC))
}

// validateCallbackToken 验证回调 token（兼容 header 传递方式）
func (h *OnlyOfficeHandler) validateCallbackToken(c *gin.Context) bool {
	token := c.GetHeader("Authorization")
	if token == "" {
		token = c.GetHeader("X-Security-Token")
	}

	if token == "" {
		return false
	}

	// Header 传递模式下，token 本身是 secret 的 HMAC
	tokenMAC := h.GenerateCallbackToken(token)
	return hmac.Equal([]byte(token), []byte(tokenMAC))
}

// GenerateCallbackToken 生成回调 token
func (h *OnlyOfficeHandler) GenerateCallbackToken(payload string) string {
	if h.onlyOfficeSecret == "" {
		return ""
	}

	mac := hmac.New(sha256.New, []byte(h.onlyOfficeSecret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
