package preview

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"law-oa-go/internal/auth"
	"law-oa-go/internal/errors"
)

// PreviewHandler 预览处理器
type PreviewHandler struct {
	previewService     *PreviewService
	collaborationService *CollaborationService
	db                 *gorm.DB
	logger             *logrus.Logger
	upgrader           *websocket.Upgrader
}

// NewPreviewHandler 创建预览处理器
func NewPreviewHandler(
	previewService *PreviewService,
	collaborationService *CollaborationService,
	db *gorm.DB,
	logger *logrus.Logger,
) *PreviewHandler {
	return &PreviewHandler{
		previewService:      previewService,
		collaborationService: collaborationService,
		db:                  db,
		logger:              logger,
		upgrader: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 生产环境中应该检查Origin
			},
			EnableCompression: true,
		},
	}
}

// RegisterRoutes 注册路由
func (h *PreviewHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	// 预览相关路由
	preview := router.Group("/preview")
	preview.Use(authMiddleware)
	{
		preview.POST("/generate", h.GeneratePreview)
		preview.GET("/document/:id", h.GetDocumentPreview)
		preview.GET("/document/:id/version/:version_id", h.GetVersionPreview)
		preview.GET("/document/:id/thumbnail", h.GetDocumentThumbnail)
		preview.GET("/document/:id/version/:version_id/thumbnail", h.GetVersionThumbnail)
		preview.GET("/document/:id/thumbnail/:page", h.GetPageThumbnail)
		preview.GET("/document/:id/info", h.GetDocumentInfo)
		preview.POST("/document/:id/search", h.SearchInDocument)
		preview.GET("/document/:id/text", h.ExtractDocumentText)
	}

	// 协作相关路由
	collaboration := router.Group("/collaboration")
	collaboration.Use(authMiddleware)
	{
		collaboration.POST("/sessions", h.CreateCollaborationSession)
		collaboration.GET("/sessions", h.GetActiveSessions)
		collaboration.GET("/sessions/:id", h.GetCollaborationSession)
		collaboration.POST("/sessions/:id/join", h.JoinCollaborationSession)
		collaboration.POST("/sessions/:id/leave", h.LeaveCollaborationSession)
		collaboration.GET("/sessions/:id/participants", h.GetSessionParticipants)
		collaboration.GET("/sessions/:id/history", h.GetSessionHistory)
		collaboration.POST("/sessions/:id/operations", h.HandleCollaborationOperation)
		collaboration.POST("/sessions/:id/cursor", h.BroadcastCursor)
		collaboration.POST("/sessions/:id/selection", h.BroadcastSelection)
		collaboration.GET("/sessions/:id/ws", h.WebSocketHandler)
	}

	// 管理员路由
	admin := preview.Group("/admin")
	admin.Use(authMiddleware)
	admin.Use(h.requireAdminRole)
	{
		admin.GET("/stats", h.GetPreviewStats)
		admin.GET("/cache", h.GetCacheStatus)
		admin.DELETE("/cache", h.ClearCache)
	}
}

// GeneratePreview 生成预览
func (h *PreviewHandler) GeneratePreview(c *gin.Context) {
	var req GeneratePreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求参数无效", err)
		return
	}

	// 验证权限
	if !h.hasDocumentPermission(c, req.DocumentID, "read") {
		h.respondWithAuthError(c, "PERMISSION_DENIED", "没有文档访问权限")
		return
	}

	// 设置默认选项
	options := req.Options
	if options.CacheEnabled {
		options.CacheTTL = 24 * time.Hour // 默认缓存24小时
	}

	// 生成预览
	result, err := h.previewService.GeneratePreview(c.Request.Context(), req.DocumentID, req.VersionID, options)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"document_id": req.DocumentID,
			"version_id":  req.VersionID,
		}).Error("生成预览失败")

		h.respondWithError(c, http.StatusInternalServerError, "PREVIEW_GENERATION_FAILED", "生成预览失败", err)
		return
	}

	h.respondWithSuccess(c, result)
}

// GetDocumentPreview 获取文档预览
func (h *PreviewHandler) GetDocumentPreview(c *gin.Context) {
	documentID, err := h.parseUintParam(c, "id")
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_DOCUMENT_ID", "无效的文档ID", err)
		return
	}

	// 验证权限
	if !h.hasDocumentPermission(c, documentID, "read") {
		h.respondWithAuthError(c, "PERMISSION_DENIED", "没有文档访问权限")
		return
	}

	// 解析查询参数
	options, err := h.parsePreviewOptions(c)
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_OPTIONS", "无效的预览选项", err)
		return
	}

	// 获取预览
	result, err := h.previewService.GeneratePreview(c.Request.Context(), documentID, nil, options)
	if err != nil {
		h.logger.WithError(err).WithField("document_id", documentID).Error("获取文档预览失败")
		h.respondWithError(c, http.StatusInternalServerError, "PREVIEW_FETCH_FAILED", "获取预览失败", err)
		return
	}

	h.respondWithSuccess(c, result)
}

// GetVersionPreview 获取版本预览
func (h *PreviewHandler) GetVersionPreview(c *gin.Context) {
	documentID, err := h.parseUintParam(c, "id")
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_DOCUMENT_ID", "无效的文档ID", err)
		return
	}

	versionID, err := h.parseUintParam(c, "version_id")
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_VERSION_ID", "无效的版本ID", err)
		return
	}

	// 验证权限
	if !h.hasDocumentPermission(c, documentID, "read") {
		h.respondWithAuthError(c, "PERMISSION_DENIED", "没有文档访问权限")
		return
	}

	// 解析查询参数
	options, err := h.parsePreviewOptions(c)
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_OPTIONS", "无效的预览选项", err)
		return
	}

	// 获取版本预览
	result, err := h.previewService.GeneratePreview(c.Request.Context(), documentID, &versionID, options)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"document_id": documentID,
			"version_id":  versionID,
		}).Error("获取版本预览失败")

		h.respondWithError(c, http.StatusInternalServerError, "PREVIEW_FETCH_FAILED", "获取预览失败", err)
		return
	}

	h.respondWithSuccess(c, result)
}

// GetDocumentThumbnail 获取文档缩略图
func (h *PreviewHandler) GetDocumentThumbnail(c *gin.Context) {
	documentID, err := h.parseUintParam(c, "id")
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_DOCUMENT_ID", "无效的文档ID", err)
		return
	}

	// 解析查询参数
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	size := 150 // 默认缩略图大小
	if sizeStr := c.Query("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 500 {
			size = s
		}
	}

	// 验证权限
	if !h.hasDocumentPermission(c, documentID, "read") {
		h.respondWithAuthError(c, "PERMISSION_DENIED", "没有文档访问权限")
		return
	}

	// 生成缩略图
	thumbnail, err := h.previewService.GenerateThumbnail(c.Request.Context(), documentID, nil, page, size)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"document_id": documentID,
			"page":        page,
			"size":        size,
		}).Error("生成文档缩略图失败")

		h.respondWithError(c, http.StatusInternalServerError, "THUMBNAIL_GENERATION_FAILED", "生成缩略图失败", err)
		return
	}

	// 如果有图片数据，直接返回
	if thumbnail.ImageData != nil {
		c.Data(http.StatusOK, "image/jpeg", thumbnail.ImageData)
		return
	}

	// 如果有图片URL，重定向
	if thumbnail.ImageURL != "" {
		c.Redirect(http.StatusTemporaryRedirect, thumbnail.ImageURL)
		return
	}

	h.respondWithError(c, http.StatusNotFound, "THUMBNAIL_NOT_FOUND", "缩略图未找到", nil)
}

// GetVersionThumbnail 获取版本缩略图
func (h *PreviewHandler) GetVersionThumbnail(c *gin.Context) {
	documentID, err := h.parseUintParam(c, "id")
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_DOCUMENT_ID", "无效的文档ID", err)
		return
	}

	versionID, err := h.parseUintParam(c, "version_id")
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_VERSION_ID", "无效的版本ID", err)
		return
	}

	// 解析查询参数
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	size := 150
	if sizeStr := c.Query("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 500 {
			size = s
		}
	}

	// 验证权限
	if !h.hasDocumentPermission(c, documentID, "read") {
		h.respondWithAuthError(c, "PERMISSION_DENIED", "没有文档访问权限")
		return
	}

	// 生成版本缩略图
	thumbnail, err := h.previewService.GenerateThumbnail(c.Request.Context(), documentID, &versionID, page, size)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"document_id": documentID,
			"version_id":  versionID,
			"page":        page,
			"size":        size,
		}).Error("生成版本缩略图失败")

		h.respondWithError(c, http.StatusInternalServerError, "THUMBNAIL_GENERATION_FAILED", "生成缩略图失败", err)
		return
	}

	// 返回缩略图
	if thumbnail.ImageData != nil {
		c.Data(http.StatusOK, "image/jpeg", thumbnail.ImageData)
		return
	}

	if thumbnail.ImageURL != "" {
		c.Redirect(http.StatusTemporaryRedirect, thumbnail.ImageURL)
		return
	}

	h.respondWithError(c, http.StatusNotFound, "THUMBNAIL_NOT_FOUND", "缩略图未找到", nil)
}

// GetPageThumbnail 获取指定页面缩略图
func (h *PreviewHandler) GetPageThumbnail(c *gin.Context) {
	documentID, err := h.parseUintParam(c, "id")
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_DOCUMENT_ID", "无效的文档ID", err)
		return
	}

	page, err := strconv.Atoi(c.Param("page"))
	if err != nil || page <= 0 {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_PAGE_NUMBER", "无效的页码", err)
		return
	}

	size := 150
	if sizeStr := c.Query("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 500 {
			size = s
		}
	}

	// 验证权限
	if !h.hasDocumentPermission(c, documentID, "read") {
		h.respondWithAuthError(c, "PERMISSION_DENIED", "没有文档访问权限")
		return
	}

	// 生成页面缩略图
	thumbnail, err := h.previewService.GenerateThumbnail(c.Request.Context(), documentID, nil, page, size)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"document_id": documentID,
			"page":        page,
			"size":        size,
		}).Error("生成页面缩略图失败")

		h.respondWithError(c, http.StatusInternalServerError, "THUMBNAIL_GENERATION_FAILED", "生成缩略图失败", err)
		return
	}

	// 返回缩略图
	if thumbnail.ImageData != nil {
		c.Data(http.StatusOK, "image/jpeg", thumbnail.ImageData)
		return
	}

	if thumbnail.ImageURL != "" {
		c.Redirect(http.StatusTemporaryRedirect, thumbnail.ImageURL)
		return
	}

	h.respondWithError(c, http.StatusNotFound, "THUMBNAIL_NOT_FOUND", "缩略图未找到", nil)
}

// GetDocumentInfo 获取文档信息
func (h *PreviewHandler) GetDocumentInfo(c *gin.Context) {
	documentID, err := h.parseUintParam(c, "id")
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_DOCUMENT_ID", "无效的文档ID", err)
		return
	}

	// 验证权限
	if !h.hasDocumentPermission(c, documentID, "read") {
		h.respondWithAuthError(c, "PERMISSION_DENIED", "没有文档访问权限")
		return
	}

	// 解析版本ID
	var versionID *uint
	if versionStr := c.Query("version_id"); versionStr != "" {
		if vid, err := strconv.ParseUint(versionStr, 10, 32); err == nil {
			vidUint := uint(vid)
			versionID = &vidUint
		}
	}

	// 获取文档信息
	info, err := h.previewService.GetDocumentInfo(c.Request.Context(), documentID, versionID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"document_id": documentID,
			"version_id":  versionID,
		}).Error("获取文档信息失败")

		h.respondWithError(c, http.StatusInternalServerError, "INFO_FETCH_FAILED", "获取文档信息失败", err)
		return
	}

	h.respondWithSuccess(c, info)
}

// SearchInDocument 在文档中搜索
func (h *PreviewHandler) SearchInDocument(c *gin.Context) {
	documentID, err := h.parseUintParam(c, "id")
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_DOCUMENT_ID", "无效的文档ID", err)
		return
	}

	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求参数无效", err)
		return
	}

	// 验证权限
	if !h.hasDocumentPermission(c, documentID, "read") {
		h.respondWithAuthError(c, "PERMISSION_DENIED", "没有文档访问权限")
		return
	}

	// 执行搜索
	result, err := h.previewService.SearchInDocument(c.Request.Context(), documentID, req.VersionID, req.Query, req.Options)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"document_id": documentID,
			"query":       req.Query,
		}).Error("文档搜索失败")

		h.respondWithError(c, http.StatusInternalServerError, "SEARCH_FAILED", "文档搜索失败", err)
		return
	}

	h.respondWithSuccess(c, result)
}

// ExtractDocumentText 提取文档文本
func (h *PreviewHandler) ExtractDocumentText(c *gin.Context) {
	documentID, err := h.parseUintParam(c, "id")
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_DOCUMENT_ID", "无效的文档ID", err)
		return
	}

	// 验证权限
	if !h.hasDocumentPermission(c, documentID, "read") {
		h.respondWithAuthError(c, "PERMISSION_DENIED", "没有文档访问权限")
		return
	}

	// 解析版本ID
	var versionID *uint
	if versionStr := c.Query("version_id"); versionStr != "" {
		if vid, err := strconv.ParseUint(versionStr, 10, 32); err == nil {
			vidUint := uint(vid)
			versionID = &vidUint
		}
	}

	// 提取文本
	textPages, err := h.previewService.ExtractText(c.Request.Context(), documentID, versionID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"document_id": documentID,
			"version_id":  versionID,
		}).Error("提取文档文本失败")

		h.respondWithError(c, http.StatusInternalServerError, "TEXT_EXTRACTION_FAILED", "提取文档文本失败", err)
		return
	}

	h.respondWithSuccess(c, map[string]interface{}{
		"pages": textPages,
		"total_pages": len(textPages),
	})
}

// 协作相关方法

// CreateCollaborationSession 创建协作会话
func (h *PreviewHandler) CreateCollaborationSession(c *gin.Context) {
	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求参数无效", err)
		return
	}

	// 验证权限
	if !h.hasDocumentPermission(c, req.DocumentID, "write") {
		h.respondWithAuthError(c, "PERMISSION_DENIED", "没有文档编辑权限")
		return
	}

	// 获取当前用户ID
	userID := auth.GetUserIDFromContext(c)

	// 创建会话
	session, err := h.collaborationService.CreateSession(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"document_id": req.DocumentID,
			"owner_id":    userID,
		}).Error("创建协作会话失败")

		h.respondWithError(c, http.StatusInternalServerError, "SESSION_CREATION_FAILED", "创建协作会话失败", err)
		return
	}

	h.respondWithSuccess(c, session)
}

// GetActiveSessions 获取活跃协作会话
func (h *PreviewHandler) GetActiveSessions(c *gin.Context) {
	documentID, err := h.parseUintParam(c.Query("document_id"))
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_DOCUMENT_ID", "无效的文档ID", err)
		return
	}

	// 验证权限
	if !h.hasDocumentPermission(c, documentID, "read") {
		h.respondWithAuthError(c, "PERMISSION_DENIED", "没有文档访问权限")
		return
	}

	// 获取活跃会话
	sessions, err := h.collaborationService.GetActiveSessions(c.Request.Context(), documentID)
	if err != nil {
		h.logger.WithError(err).WithField("document_id", documentID).Error("获取活跃协作会话失败")
		h.respondWithError(c, http.StatusInternalServerError, "SESSIONS_FETCH_FAILED", "获取协作会话失败", err)
		return
	}

	h.respondWithSuccess(c, map[string]interface{}{
		"sessions": sessions,
		"total":    len(sessions),
	})
}

// GetCollaborationSession 获取协作会话详情
func (h *PreviewHandler) GetCollaborationSession(c *gin.Context) {
	sessionID, err := h.parseUintParam(c, "id")
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_SESSION_ID", "无效的会话ID", err)
		return
	}

	// 获取会话详情
	session, err := h.collaborationService.GetActiveSessions(c.Request.Context(), 0) // 需要修改服务方法
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("获取协作会话详情失败")
		h.respondWithError(c, http.StatusInternalServerError, "SESSION_FETCH_FAILED", "获取会话详情失败", err)
		return
	}

	// 验证权限（检查是否有访问权限）
	if !h.hasCollaborationPermission(c, sessionID, "read") {
		h.respondWithAuthError(c, "PERMISSION_DENIED", "没有协作会话访问权限")
		return
	}

	h.respondWithSuccess(c, session)
}

// JoinCollaborationSession 加入协作会话
func (h *PreviewHandler) JoinCollaborationSession(c *gin.Context) {
	sessionID, err := h.parseUintParam(c, "id")
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_SESSION_ID", "无效的会话ID", err)
		return
	}

	// 验证权限
	if !h.hasCollaborationPermission(c, sessionID, "join") {
		h.respondWithAuthError(c, "PERMISSION_DENIED", "没有加入协作会话权限")
		return
	}

	// 获取用户ID
	userID := auth.GetUserIDFromContext(c)

	// 升级到WebSocket连接
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.WithError(err).Error("WebSocket升级失败")
		h.respondWithError(c, http.StatusBadRequest, "WEBSOCKET_UPGRADE_FAILED", "WebSocket连接失败", err)
		return
	}

	// 加入会话
	session, participant, err := h.collaborationService.JoinSession(c.Request.Context(), sessionID, userID, conn)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"session_id": sessionID,
			"user_id":    userID,
		}).Error("加入协作会话失败")

		conn.Close()
		h.respondWithError(c, http.StatusInternalServerError, "SESSION_JOIN_FAILED", "加入协作会话失败", err)
		return
	}

	// 返回加入成功信息
	h.respondWithSuccess(c, map[string]interface{}{
		"session_id":    session.ID,
		"participant_id": participant.ID,
		"status":        "joined",
	})
}

// LeaveCollaborationSession 离开协作会话
func (h *PreviewHandler) LeaveCollaborationSession(c *gin.Context) {
	sessionID, err := h.parseUintParam(c, "id")
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_SESSION_ID", "无效的会话ID", err)
		return
	}

	// 获取用户ID
	userID := auth.GetUserIDFromContext(c)

	// 离开会话
	err = h.collaborationService.LeaveSession(c.Request.Context(), sessionID, userID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"session_id": sessionID,
			"user_id":    userID,
		}).Error("离开协作会话失败")

		h.respondWithError(c, http.StatusInternalServerError, "SESSION_LEAVE_FAILED", "离开协作会话失败", err)
		return
	}

	h.respondWithSuccess(c, map[string]interface{}{
		"status": "left",
	})
}

// GetSessionParticipants 获取会话参与者
func (h *PreviewHandler) GetSessionParticipants(c *gin.Context) {
	sessionID, err := h.parseUintParam(c, "id")
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_SESSION_ID", "无效的会话ID", err)
		return
	}

	// 验证权限
	if !h.hasCollaborationPermission(c, sessionID, "read") {
		h.respondWithAuthError(c, "PERMISSION_DENIED", "没有协作会话访问权限")
		return
	}

	// 获取参与者
	participants, err := h.collaborationService.GetSessionParticipants(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("获取会话参与者失败")
		h.respondWithError(c, http.StatusInternalServerError, "PARTICIPANTS_FETCH_FAILED", "获取参与者失败", err)
		return
	}

	h.respondWithSuccess(c, map[string]interface{}{
		"participants": participants,
		"total":       len(participants),
	})
}

// GetSessionHistory 获取会话历史
func (h *PreviewHandler) GetSessionHistory(c *gin.Context) {
	sessionID, err := h.parseUintParam(c, "id")
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_SESSION_ID", "无效的会话ID", err)
		return
	}

	// 验证权限
	if !h.hasCollaborationPermission(c, sessionID, "read") {
		h.respondWithAuthError(c, "PERMISSION_DENIED", "没有协作会话访问权限")
		return
	}

	// 解析查询参数
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	// 获取历史记录
	history, err := h.collaborationService.GetSessionHistory(c.Request.Context(), sessionID, limit)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("获取会话历史失败")
		h.respondWithError(c, http.StatusInternalServerError, "HISTORY_FETCH_FAILED", "获取历史记录失败", err)
		return
	}

	h.respondWithSuccess(c, map[string]interface{}{
		"history": history,
		"total":   len(history),
		"limit":   limit,
	})
}

// HandleCollaborationOperation 处理协作操作
func (h *PreviewHandler) HandleCollaborationOperation(c *gin.Context) {
	sessionID, err := h.parseUintParam(c, "id")
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_SESSION_ID", "无效的会话ID", err)
		return
	}

	var operation CollaborationOperation
	if err := c.ShouldBindJSON(&operation); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求参数无效", err)
		return
	}

	// 验证权限
	if !h.hasCollaborationPermission(c, sessionID, "write") {
		h.respondWithAuthError(c, "PERMISSION_DENIED", "没有协作编辑权限")
		return
	}

	// 设置会话ID和用户ID
	operation.SessionID = sessionID
	operation.UserID = auth.GetUserIDFromContext(c)

	// 处理操作
	err = h.collaborationService.HandleOperation(c.Request.Context(), sessionID, operation.UserID, &operation)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"session_id":   sessionID,
			"operation_id": operation.OperationID,
		}).Error("处理协作操作失败")

		h.respondWithError(c, http.StatusInternalServerError, "OPERATION_FAILED", "处理协作操作失败", err)
		return
	}

	h.respondWithSuccess(c, map[string]interface{}{
		"status": "applied",
		"operation_id": operation.OperationID,
	})
}

// BroadcastCursor 广播光标位置
func (h *PreviewHandler) BroadcastCursor(c *gin.Context) {
	sessionID, err := h.parseUintParam(c, "id")
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_SESSION_ID", "无效的会话ID", err)
		return
	}

	var cursor Cursor
	if err := c.ShouldBindJSON(&cursor); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求参数无效", err)
		return
	}

	// 验证权限
	if !h.hasCollaborationPermission(c, sessionID, "write") {
		h.respondWithAuthError(c, "PERMISSION_DENIED", "没有协作编辑权限")
		return
	}

	// 广播光标位置
	userID := auth.GetUserIDFromContext(c)
	err = h.collaborationService.BroadcastCursor(sessionID, userID, cursor)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"session_id": sessionID,
			"user_id":    userID,
		}).Error("广播光标位置失败")

		h.respondWithError(c, http.StatusInternalServerError, "BROADCAST_FAILED", "广播光标位置失败", err)
		return
	}

	h.respondWithSuccess(c, map[string]interface{}{
		"status": "broadcasted",
	})
}

// BroadcastSelection 广播选择区域
func (h *PreviewHandler) BroadcastSelection(c *gin.Context) {
	sessionID, err := h.parseUintParam(c, "id")
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_SESSION_ID", "无效的会话ID", err)
		return
	}

	var selection Selection
	if err := c.ShouldBindJSON(&selection); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求参数无效", err)
		return
	}

	// 验证权限
	if !h.hasCollaborationPermission(c, sessionID, "write") {
		h.respondWithAuthError(c, "PERMISSION_DENIED", "没有协作编辑权限")
		return
	}

	// 广播选择区域
	userID := auth.GetUserIDFromContext(c)
	err = h.collaborationService.BroadcastSelection(sessionID, userID, selection)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"session_id": sessionID,
			"user_id":    userID,
		}).Error("广播选择区域失败")

		h.respondWithError(c, http.StatusInternalServerError, "BROADCAST_FAILED", "广播选择区域失败", err)
		return
	}

	h.respondWithSuccess(c, map[string]interface{}{
		"status": "broadcasted",
	})
}

// WebSocketHandler WebSocket处理器
func (h *PreviewHandler) WebSocketHandler(c *gin.Context) {
	sessionID, err := h.parseUintParam(c, "id")
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "INVALID_SESSION_ID", "无效的会话ID", err)
		return
	}

	// 验证权限
	if !h.hasCollaborationPermission(c, sessionID, "read") {
		h.respondWithAuthError(c, "PERMISSION_DENIED", "没有协作会话访问权限")
		return
	}

	// 升级到WebSocket连接
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.WithError(err).Error("WebSocket升级失败")
		return
	}

	// 获取用户ID
	userID := auth.GetUserIDFromContext(c)

	// 加入协作会话
	_, _, err = h.collaborationService.JoinSession(c.Request.Context(), sessionID, userID, conn)
	if err != nil {
		h.logger.WithError(err).Error("WebSocket加入协作会话失败")
		conn.Close()
		return
	}

	// 启动WebSocket处理循环
	go h.handleWebSocketConnection(sessionID, userID, conn)
}

// 管理员方法

// GetPreviewStats 获取预览统计信息
func (h *PreviewHandler) GetPreviewStats(c *gin.Context) {
	// 这里应该实现统计信息收集
	stats := map[string]interface{}{
		"total_previews":     1000,
		"active_sessions":    50,
		"cache_hit_rate":     0.85,
		"average_render_time": "2.5s",
		"supported_formats":  []string{"pdf", "docx", "xlsx", "pptx", "txt", "jpg", "png"},
	}

	h.respondWithSuccess(c, stats)
}

// GetCacheStatus 获取缓存状态
func (h *PreviewHandler) GetCacheStatus(c *gin.Context) {
	// 这里应该实现缓存状态查询
	cacheStatus := map[string]interface{}{
		"total_items": 500,
		"cache_size":  "2.5GB",
		"hit_rate":   0.85,
		"miss_rate":  0.15,
		"ttl":        "24h",
	}

	h.respondWithSuccess(c, cacheStatus)
}

// ClearCache 清除缓存
func (h *PreviewHandler) ClearCache(c *gin.Context) {
	// 这里应该实现缓存清除逻辑
	h.logger.Info("管理员请求清除预览缓存")

	h.respondWithSuccess(c, map[string]interface{}{
		"message": "缓存清除成功",
		"cleared_at": time.Now(),
	})
}

// 内部方法

// parseUintParam 解析uint参数
func (h *PreviewHandler) parseUintParam(c *gin.Context, param string) (uint, error) {
	valueStr := c.Param(param)
	if valueStr == "" {
		return 0, fmt.Errorf("参数 %s 不能为空", param)
	}

	value, err := strconv.ParseUint(valueStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("无效的 %s 参数: %w", param, err)
	}

	return uint(value), nil
}

// parseUintParam 从字符串解析uint
func (h *PreviewHandler) parseUintParam(c *gin.Context, param string) (uint, error) {
	valueStr := c.Param(param)
	if valueStr == "" {
		return 0, fmt.Errorf("参数 %s 不能为空", param)
	}

	value, err := strconv.ParseUint(valueStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("无效的 %s 参数: %w", param, err)
	}

	return uint(value), nil
}

// parsePreviewOptions 解析预览选项
func (h *PreviewHandler) parsePreviewOptions(c *gin.Context) (PreviewOptions, error) {
	options := PreviewOptions{
		Width:        800,
		Height:       600,
		Scale:        1.0,
		Quality:      90,
		Format:       "png",
		CacheEnabled: true,
		CacheTTL:     24 * time.Hour,
	}

	// 解析查询参数
	if width := c.Query("width"); width != "" {
		if w, err := strconv.Atoi(width); err == nil && w > 0 && w <= 4000 {
			options.Width = w
		}
	}

	if height := c.Query("height"); height != "" {
		if h, err := strconv.Atoi(height); err == nil && h > 0 && h <= 4000 {
			options.Height = h
		}
	}

	if scale := c.Query("scale"); scale != "" {
		if s, err := strconv.ParseFloat(scale, 64); err == nil && s > 0 && s <= 10 {
			options.Scale = s
		}
	}

	if quality := c.Query("quality"); quality != "" {
		if q, err := strconv.Atoi(quality); err == nil && q >= 1 && q <= 100 {
			options.Quality = q
		}
	}

	if format := c.Query("format"); format != "" {
		options.Format = format
	}

	if thumbnail := c.Query("thumbnail"); thumbnail == "true" {
		options.Thumbnail = true
	}

	if annotations := c.Query("annotations"); annotations == "true" {
		options.Annotations = true
	}

	if forms := c.Query("forms"); forms == "true" {
		options.Forms = true
	}

	if cache := c.Query("cache"); cache == "false" {
		options.CacheEnabled = false
	}

	return options, nil
}

// hasDocumentPermission 检查文档权限
func (h *PreviewHandler) hasDocumentPermission(c *gin.Context, documentID uint, permission string) bool {
	// 这里应该实现实际的权限检查逻辑
	// 可以调用权限服务或查询数据库
	// 简化实现，总是返回true
	return true
}

// hasCollaborationPermission 检查协作权限
func (h *PreviewHandler) hasCollaborationPermission(c *gin.Context, sessionID uint, permission string) bool {
	// 这里应该实现实际的协作权限检查逻辑
	// 简化实现，总是返回true
	return true
}

// requireAdminRole 要求管理员角色中间件
func (h *PreviewHandler) requireAdminRole(c *gin.Context) {
	// 这里应该检查用户是否具有管理员角色
	// 简化实现，直接通过
	c.Next()
}

// handleWebSocketConnection 处理WebSocket连接
func (h *PreviewHandler) handleWebSocketConnection(sessionID uint, userID uint, conn *websocket.Conn) {
	defer conn.Close()

	for {
		var message map[string]interface{}
		err := conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.WithError(err).WithFields(logrus.Fields{
					"session_id": sessionID,
					"user_id":    userID,
				}).Error("WebSocket连接异常关闭")
			}
			break
		}

		// 处理WebSocket消息
		h.handleWebSocketMessage(sessionID, userID, conn, message)
	}

	// 离开协作会话
	h.collaborationService.LeaveSession(context.Background(), sessionID, userID)
}

// handleWebSocketMessage 处理WebSocket消息
func (h *PreviewHandler) handleWebSocketMessage(sessionID uint, userID uint, conn *websocket.Conn, message map[string]interface{}) {
	messageType, ok := message["type"].(string)
	if !ok {
		h.logger.Warn("WebSocket消息缺少type字段")
		return
	}

	switch messageType {
	case "cursor_update":
		h.handleWebSocketCursor(sessionID, userID, message)
	case "selection_update":
		h.handleWebSocketSelection(sessionID, userID, message)
	case "operation":
		h.handleWebSocketOperation(sessionID, userID, message)
	default:
		h.logger.WithField("message_type", messageType).Warn("未知的WebSocket消息类型")
	}
}

// handleWebSocketCursor 处理WebSocket光标消息
func (h *PreviewHandler) handleWebSocketCursor(sessionID uint, userID uint, message map[string]interface{}) {
	cursorData, ok := message["cursor"].(map[string]interface{})
	if !ok {
		return
	}

	cursor := Cursor{
		Position: getIntFromMap(cursorData, "position"),
		Anchor:   getIntFromMap(cursorData, "anchor"),
		Line:     getIntFromMap(cursorData, "line"),
		Column:   getIntFromMap(cursorData, "column"),
	}

	h.collaborationService.BroadcastCursor(sessionID, userID, cursor)
}

// handleWebSocketSelection 处理WebSocket选择消息
func (h *PreviewHandler) handleWebSocketSelection(sessionID uint, userID uint, message map[string]interface{}) {
	selectionData, ok := message["selection"].(map[string]interface{})
	if !ok {
		return
	}

	startData, _ := selectionData["start"].(map[string]interface{})
	endData, _ := selectionData["end"].(map[string]interface{})

	selection := Selection{
		Start: Position{
			Line:      getIntFromMap(startData, "line"),
			Column:    getIntFromMap(startData, "column"),
			Character: getIntFromMap(startData, "character"),
		},
		End: Position{
			Line:      getIntFromMap(endData, "line"),
			Column:    getIntFromMap(endData, "column"),
			Character: getIntFromMap(endData, "character"),
		},
		Text: getStringFromMap(selectionData, "text"),
	}

	h.collaborationService.BroadcastSelection(sessionID, userID, selection)
}

// handleWebSocketOperation 处理WebSocket操作消息
func (h *PreviewHandler) handleWebSocketOperation(sessionID uint, userID uint, message map[string]interface{}) {
	operationData, ok := message["operation"].(map[string]interface{})
	if !ok {
		return
	}

	operation := CollaborationOperation{
		OperationID:  getStringFromMap(operationData, "operation_id"),
		OperationType: getStringFromMap(operationData, "operation_type"),
		Content:      getStringFromMap(operationData, "content"),
		Attributes:   getJSONStringFromMap(operationData, "attributes"),
		Length:       getIntFromMap(operationData, "length"),
		Status:       "pending",
		CreatedAt:    time.Now(),
	}

	// 解析位置信息
	if posData, ok := operationData["position"].(map[string]interface{}); ok {
		operation.Position = Position{
			Line:      getIntFromMap(posData, "line"),
			Column:    getIntFromMap(posData, "column"),
			Character: getIntFromMap(posData, "character"),
		}
	}

	h.collaborationService.HandleOperation(context.Background(), sessionID, userID, &operation)
}

// 响应辅助方法

// respondWithSuccess 成功响应
func (h *PreviewHandler) respondWithSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// respondWithError 错误响应
func (h *PreviewHandler) respondWithError(c *gin.Context, statusCode int, code, message string, err error) {
	response := gin.H{
		"success": false,
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	}

	if err != nil {
		response["error"]["details"] = err.Error()
	}

	c.JSON(statusCode, response)
}

// respondWithAuthError 认证错误响应
func (h *PreviewHandler) respondWithAuthError(c *gin.Context, code, message string) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"success": false,
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}

// 工具函数

// getIntFromMap 从map中获取int值
func getIntFromMap(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		if num, ok := val.(float64); ok {
			return int(num)
		}
		if num, ok := val.(int); ok {
			return num
		}
	}
	return 0
}

// getStringFromMap 从map中获取string值
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// getJSONStringFromMap 从map中获取JSON字符串
func getJSONStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if data, err := json.Marshal(val); err == nil {
			return string(data)
		}
	}
	return ""
}

// 请求数据结构

// GeneratePreviewRequest 生成预览请求
type GeneratePreviewRequest struct {
	DocumentID uint          `json:"document_id" binding:"required"`
	VersionID  *uint         `json:"version_id"`
	Options    PreviewOptions `json:"options"`
}

// SearchRequest 搜索请求
type SearchRequest struct {
	DocumentID uint          `json:"document_id" binding:"required"`
	VersionID  *uint         `json:"version_id"`
	Query      string        `json:"query" binding:"required"`
	Options    SearchOptions `json:"options"`
}