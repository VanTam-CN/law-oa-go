package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

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
}

// NewOnlyOfficeHandler 创建 OnlyOffice 处理器
func NewOnlyOfficeHandler(
	db *gorm.DB,
	versionService *services.DocumentVersionService,
	lockService *services.DocumentLockService,
	onlyOfficeURL string,
	onlyOfficeSecret string,
	backendURL string,
) *OnlyOfficeHandler {
	return &OnlyOfficeHandler{
		db:               db,
		versionService:   versionService,
		lockService:      lockService,
		onlyOfficeURL:    onlyOfficeURL,
		onlyOfficeSecret: onlyOfficeSecret,
		backendURL:       backendURL,
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
	About             bool `json:"about"`
	Comments          bool `json:"comments"`
	Customer          map[string]string `json:"customer"`
	Feedback          bool `json:"feedback"`
	Forcesave         bool `json:"forcesave"`
	Help              bool `json:"help"`
	Macro             bool `json:"macro"`
	Metrics           bool `json:"metrics"`
	Plugins           bool `json:"plugins"`
	CompactHeader     bool `json:"compactHeader"`
	CompactToolbar    bool `json:"compactToolbar"`
	Customization     bool `json:"customization"`
	Zoom              int  `json:"zoom"`
	ToolbarNoTabs     bool `json:"toolbarNoTabs"`
	ToolbarHideFileName bool `json:"toolbarHideFileName"`
}

// User 用户信息
type User struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Group       string `json:"group"`
}

// Embedded 嵌入配置
type Embedded struct {
	SaveURL     string `json:"saveUrl"`
	ShareURL    string `json:"shareUrl"`
	EmbedURL    string `json:"embedUrl"`
	ToolbarDocked string `json:"toolbarDocked"`
}

// OpenEditorRequest 打开编辑器请求
type OpenEditorRequest struct {
	DocumentID uint   `json:"document_id" binding:"required"`
	UserID     uint   `json:"user_id" binding:"required"`
	Mode       string `json:"mode"` // edit 或 view
}

// CallbackRequest OnlyOffice 回调请求
type CallbackRequest struct {
	Actions   []CallbackAction `json:"actions"`
	Key       int              `json:"key"`
	Status    int              `json:"status"`
	URL       string           `json:"url"`
	Users     []string         `json:"users"`
	Token     string           `json:"token"`
}

// CallbackAction 回调操作
type CallbackAction struct {
	Type    string `json:"type"`
	UserID  string `json:"userid"`
}

// 常量
const (
	// 文档状态
	StatusEditing     = 1  // 正在编辑
	StatusMustSave    = 2  // 必须保存
	StatusCorrupted   = 3  // 文档损坏
	StatusClosed      = 6  // 已关闭
	StatusMustForceSave = 7 // 强制保存

	// 文档类型
	DocumentTypeWord     = "word"
	DocumentTypeCell     = "cell"
	DocumentTypeSlide    = "slide"

	// 编辑模式
	ModeEdit = "edit"
	ModeView = "view"
)

// OpenEditor 打开文档编辑器
func (h *OnlyOfficeHandler) OpenEditor(c *gin.Context) {
	var req OpenEditorRequest
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
		// 尝试获取锁
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
				"error": "Document is locked by another user",
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
	var creator models.User
	creatorName := "Unknown"
	if doc.EntityID != 0 {
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
				Comment:  mode == ModeEdit,
				Download: true,
				Edit:     mode == ModeEdit,
				FillForms: mode == ModeEdit,
				ModifyFilter: true,
				ModifyContentControl: true,
				Review:   mode == ModeEdit,
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
				SaveURL:     docURL,
				ShareURL:    "",
				EmbedURL:    "",
				ToolbarDocked: "top",
			},
		},
	}

	c.JSON(http.StatusOK, config)
}

// HandleCallback 处理 OnlyOffice 回调
func (h *OnlyOfficeHandler) HandleCallback(c *gin.Context) {
	var req CallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证 token (如果配置了密钥)
	if h.onlyOfficeSecret != "" {
		if !h.validateCallbackToken(c) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
	}

	// 解析 key 获取文档 ID
	documentID, err := h.parseDocKey(strconv.Itoa(req.Key))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document key"})
		return
	}

	ctx := c.Request.Context()

	switch req.Status {
	case StatusMustSave, StatusMustForceSave:
		// 保存文档
		if err := h.saveDocument(ctx, documentID, req.URL); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 返回成功
		c.JSON(http.StatusOK, gin.H{"error": 0})

	case StatusEditing:
		// 文档正在编辑中，返回 0 表示正在保存
		c.JSON(http.StatusOK, gin.H{"error": 0})

	case StatusClosed:
		// 文档已关闭，释放锁
		// 从用户列表中获取最后一个用户并释放锁
		if len(req.Users) > 0 {
			userID, _ := strconv.Atoi(req.Users[0])
			if userID > 0 {
				_ = h.lockService.ReleaseLock(ctx, &services.ReleaseLockRequest{
					DocumentID: documentID,
					UserID:     uint(userID),
					Force:      false,
				})
			}
		}
		c.JSON(http.StatusOK, gin.H{"error": 0})

	default:
		c.JSON(http.StatusOK, gin.H{"error": 0})
	}
}

// saveDocument 保存文档
func (h *OnlyOfficeHandler) saveDocument(ctx context.Context, documentID uint, downloadURL string) error {
	// 下载文档
	// TODO: 实现从 OnlyOffice 下载文档并保存到存储

	// 创建新版本
	// TODO: 自动创建版本记录

	return nil
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

// validateCallbackToken 验证回调 token
func (h *OnlyOfficeHandler) validateCallbackToken(c *gin.Context) bool {
	// OnlyOffice 在 header 中传递 token
	token := c.GetHeader("Authorization")
	if token == "" {
		token = c.GetHeader("X-Security-Token")
	}

	// 如果没有 token，检查 body 中的 token
	if token == "" {
		var body map[string]interface{}
		if err := json.NewDecoder(c.Request.Body).Decode(&body); err == nil {
			if t, ok := body["token"].(string); ok {
				token = t
			}
		}
	}

	if token == "" {
		return false
	}

	// 验证 token
	// 这里简化处理，实际应该使用 JWT 验证
	return token == h.onlyOfficeSecret
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

// ConvertDocument 转换文档格式
func (h *OnlyOfficeHandler) ConvertDocument(c *gin.Context) {
	var req struct {
		DocumentID uint   `json:"document_id" binding:"required"`
		OutputType string `json:"output_type" binding:"required"` // pdf, docx, etc.
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: 实现文档转换功能
	// 使用 OnlyOffice Conversion API

	c.JSON(http.StatusNotImplemented, gin.H{"error": "Document conversion not implemented yet"})
}

// CheckConversionStatus 检查转换状态
func (h *OnlyOfficeHandler) CheckConversionStatus(c *gin.Context) {
	taskID := c.Query("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task_id is required"})
		return
	}

	// TODO: 查询转换状态

	c.JSON(http.StatusNotImplemented, gin.H{"error": "Conversion status check not implemented yet"})
}
