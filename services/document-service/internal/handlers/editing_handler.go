package handlers

import (
	"net/http"
	"strconv"

	"law-oa-go/internal/services"
	"law-oa-go/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// EditingHandler 编辑处理器
type EditingHandler struct {
	editService       services.EditingService
	collabServer      *websocket.CollaborationServer
	logger            *logrus.Logger
}

// NewEditingHandler 创建编辑处理器
func NewEditingHandler(
	editService services.EditingService,
	collabServer *websocket.CollaborationServer,
	logger *logrus.Logger,
) *EditingHandler {
	return &EditingHandler{
		editService:  editService,
		collabServer: collabServer,
		logger:       logger,
	}
}

// CreateSession 创建编辑会话
// @Summary 创建编辑会话
// @Description 为用户创建一个新的编辑会话
// @Tags 编辑
// @Accept json
// @Produce json
// @Param request body services.CreateSessionRequest true "创建会话请求"
// @Success 200 {object} services.SessionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/sessions [post]
func (h *EditingHandler) CreateSession(c *gin.Context) {
	var req services.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("绑定创建会话请求失败")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	// 验证请求
	if err := h.validateCreateSessionRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "请求验证失败",
			Details: err.Error(),
		})
		return
	}

	response, err := h.editService.CreateSession(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("创建编辑会话失败")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "create_session_failed",
			Message: "创建编辑会话失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data:    response,
		Message: "创建编辑会话成功",
	})
}

// GetSession 获取编辑会话
// @Summary 获取编辑会话
// @Description 根据会话ID获取编辑会话信息
// @Tags 编辑
// @Accept json
// @Produce json
// @Param sessionId path string true "会话ID"
// @Success 200 {object} services.SessionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/sessions/{sessionId} [get]
func (h *EditingHandler) GetSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_session_id",
			Message: "缺少会话ID",
		})
		return
	}

	response, err := h.editService.GetSession(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("获取编辑会话失败")

		if err.Error() == "编辑会话已过期" {
			c.JSON(http.StatusGone, ErrorResponse{
				Error:   "session_expired",
				Message: "编辑会话已过期",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "session_not_found",
			Message: "编辑会话不存在",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data:    response,
		Message: "获取编辑会话成功",
	})
}

// UpdateSessionStatus 更新会话状态
// @Summary 更新会话状态
// @Description 更新编辑会话的状态信息
// @Tags 编辑
// @Accept json
// @Produce json
// @Param sessionId path string true "会话ID"
// @Param request body services.SessionStatus true "会话状态"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/sessions/{sessionId}/status [put]
func (h *EditingHandler) UpdateSessionStatus(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_session_id",
			Message: "缺少会话ID",
		})
		return
	}

	var status services.SessionStatus
	if err := c.ShouldBindJSON(&status); err != nil {
		h.logger.WithError(err).Error("绑定会话状态请求失败")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	err := h.editService.UpdateSessionStatus(c.Request.Context(), sessionID, &status)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("更新会话状态失败")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "update_session_failed",
			Message: "更新会话状态失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "更新会话状态成功",
	})
}

// CloseSession 关闭编辑会话
// @Summary 关闭编辑会话
// @Description 关闭指定的编辑会话
// @Tags 编辑
// @Accept json
// @Produce json
// @Param sessionId path string true "会话ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/sessions/{sessionId} [delete]
func (h *EditingHandler) CloseSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_session_id",
			Message: "缺少会话ID",
		})
		return
	}

	err := h.editService.CloseSession(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("关闭编辑会话失败")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "close_session_failed",
			Message: "关闭编辑会话失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "关闭编辑会话成功",
	})
}

// HandleEditOperation 处理编辑操作
// @Summary 处理编辑操作
// @Description 处理用户的编辑操作
// @Tags 编辑
// @Accept json
// @Produce json
// @Param request body services.EditOperationRequest true "编辑操作请求"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/operations [post]
func (h *EditingHandler) HandleEditOperation(c *gin.Context) {
	var req services.EditOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("绑定编辑操作请求失败")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	err := h.editService.HandleEditOperation(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("处理编辑操作失败")

		if err.Error() == "没有编辑权限" {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "permission_denied",
				Message: "没有编辑权限",
				Details: err.Error(),
			})
			return
		}

		if err.Error() == "编辑会话已过期" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "session_expired",
				Message: "编辑会话已过期",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "handle_operation_failed",
			Message: "处理编辑操作失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "处理编辑操作成功",
	})
}

// GetDocumentOperations 获取文档操作列表
// @Summary 获取文档操作列表
// @Description 获取指定文档的编辑操作列表
// @Tags 编辑
// @Accept json
// @Produce json
// @Param documentId path string true "文档ID"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} services.OperationListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/documents/{documentId}/operations [get]
func (h *EditingHandler) GetDocumentOperations(c *gin.Context) {
	documentID := c.Param("documentId")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_document_id",
			Message: "缺少文档ID",
		})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	response, err := h.editService.GetDocumentOperations(c.Request.Context(), documentID, page, limit)
	if err != nil {
		h.logger.WithError(err).WithField("document_id", documentID).Error("获取文档操作失败")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "get_operations_failed",
			Message: "获取文档操作失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data:    response,
		Message: "获取文档操作成功",
	})
}

// JoinCollaboration 加入协作
// @Summary 加入协作
// @Description 用户加入文档协作编辑
// @Tags 协作
// @Accept json
// @Produce json
// @Param request body services.JoinCollaborationRequest true "加入协作请求"
// @Success 200 {object} services.CollaborationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/collaboration/join [post]
func (h *EditingHandler) JoinCollaboration(c *gin.Context) {
	var req services.JoinCollaborationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("绑定加入协作请求失败")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	response, err := h.editService.JoinCollaboration(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("加入协作失败")

		if err.Error() == "没有协作权限" {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "permission_denied",
				Message: "没有协作权限",
				Details: err.Error(),
			})
			return
		}

		if err.Error() == "协作用户数量已达上限" {
			c.JSON(http.StatusTooManyRequests, ErrorResponse{
				Error:   "room_full",
				Message: "协作房间人数已满",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "join_collaboration_failed",
			Message: "加入协作失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data:    response,
		Message: "加入协作成功",
	})
}

// LeaveCollaboration 离开协作
// @Summary 离开协作
// @Description 用户离开文档协作编辑
// @Tags 协作
// @Accept json
// @Produce json
// @Param documentId path string true "文档ID"
// @Param userId path string true "用户ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/collaboration/documents/{documentId}/users/{userId} [delete]
func (h *EditingHandler) LeaveCollaboration(c *gin.Context) {
	documentID := c.Param("documentId")
	userID := c.Param("userId")

	if documentID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_parameters",
			Message: "缺少文档ID或用户ID",
		})
		return
	}

	err := h.editService.LeaveCollaboration(c.Request.Context(), documentID, userID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"document_id": documentID,
			"user_id":     userID,
		}).Error("离开协作失败")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "leave_collaboration_failed",
			Message: "离开协作失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "离开协作成功",
	})
}

// GetCollaborators 获取协作者列表
// @Summary 获取协作者列表
// @Description 获取文档的协作者列表
// @Tags 协作
// @Accept json
// @Produce json
// @Param documentId path string true "文档ID"
// @Success 200 {object} services.CollaboratorListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/collaboration/documents/{documentId}/collaborators [get]
func (h *EditingHandler) GetCollaborators(c *gin.Context) {
	documentID := c.Param("documentId")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_document_id",
			Message: "缺少文档ID",
		})
		return
	}

	response, err := h.editService.GetCollaborators(c.Request.Context(), documentID)
	if err != nil {
		h.logger.WithError(err).WithField("document_id", documentID).Error("获取协作者列表失败")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "get_collaborators_failed",
			Message: "获取协作者列表失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data:    response,
		Message: "获取协作者列表成功",
	})
}

// CreateVersion 创建文档版本
// @Summary 创建文档版本
// @Description 为文档创建新版本
// @Tags 版本
// @Accept json
// @Produce json
// @Param request body services.CreateVersionRequest true "创建版本请求"
// @Success 200 {object} services.VersionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/documents/versions [post]
func (h *EditingHandler) CreateVersion(c *gin.Context) {
	var req services.CreateVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("绑定创建版本请求失败")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	response, err := h.editService.CreateVersion(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("创建文档版本失败")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "create_version_failed",
			Message: "创建文档版本失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data:    response,
		Message: "创建文档版本成功",
	})
}

// GetVersions 获取版本列表
// @Summary 获取版本列表
// @Description 获取文档的版本列表
// @Tags 版本
// @Accept json
// @Produce json
// @Param documentId path string true "文档ID"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} services.VersionListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/documents/{documentId}/versions [get]
func (h *EditingHandler) GetVersions(c *gin.Context) {
	documentID := c.Param("documentId")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_document_id",
			Message: "缺少文档ID",
		})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	response, err := h.editService.GetVersions(c.Request.Context(), documentID, page, limit)
	if err != nil {
		h.logger.WithError(err).WithField("document_id", documentID).Error("获取版本列表失败")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "get_versions_failed",
			Message: "获取版本列表失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data:    response,
		Message: "获取版本列表成功",
	})
}

// CompareVersions 比较版本
// @Summary 比较版本
// @Description 比较两个文档版本的差异
// @Tags 版本
// @Accept json
// @Produce json
// @Param documentId path string true "文档ID"
// @Param fromVersion query string true "源版本ID"
// @Param toVersion query string true "目标版本ID"
// @Success 200 {object} services.VersionComparisonResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/documents/{documentId}/versions/compare [get]
func (h *EditingHandler) CompareVersions(c *gin.Context) {
	documentID := c.Param("documentId")
	fromVersion := c.Query("fromVersion")
	toVersion := c.Query("toVersion")

	if documentID == "" || fromVersion == "" || toVersion == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_parameters",
			Message: "缺少文档ID、源版本ID或目标版本ID",
		})
		return
	}

	response, err := h.editService.CompareVersions(c.Request.Context(), documentID, fromVersion, toVersion)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"document_id":   documentID,
			"from_version":  fromVersion,
			"to_version":    toVersion,
		}).Error("比较版本失败")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "compare_versions_failed",
			Message: "比较版本失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data:    response,
		Message: "比较版本成功",
	})
}

// RestoreVersion 恢复版本
// @Summary 恢复版本
// @Description 将文档恢复到指定版本
// @Tags 版本
// @Accept json
// @Produce json
// @Param documentId path string true "文档ID"
// @Param versionId path string true "版本ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/documents/{documentId}/versions/{versionId}/restore [post]
func (h *EditingHandler) RestoreVersion(c *gin.Context) {
	documentID := c.Param("documentId")
	versionID := c.Param("versionId")

	if documentID == "" || versionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_parameters",
			Message: "缺少文档ID或版本ID",
		})
		return
	}

	err := h.editService.RestoreVersion(c.Request.Context(), documentID, versionID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"document_id": documentID,
			"version_id":  versionID,
		}).Error("恢复版本失败")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "restore_version_failed",
			Message: "恢复版本失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "恢复版本成功",
	})
}

// GetEditorConfig 获取编辑器配置
// @Summary 获取编辑器配置
// @Description 获取指定类型编辑器的配置
// @Tags 配置
// @Accept json
// @Produce json
// @Param editorType path string true "编辑器类型"
// @Success 200 {object} services.EditorConfigResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/config/{editorType} [get]
func (h *EditingHandler) GetEditorConfig(c *gin.Context) {
	editorType := c.Param("editorType")
	if editorType == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_editor_type",
			Message: "缺少编辑器类型",
		})
		return
	}

	response, err := h.editService.GetEditorConfig(c.Request.Context(), editorType)
	if err != nil {
		h.logger.WithError(err).WithField("editor_type", editorType).Error("获取编辑器配置失败")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "get_config_failed",
			Message: "获取编辑器配置失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data:    response,
		Message: "获取编辑器配置成功",
	})
}

// UpdateEditorConfig 更新编辑器配置
// @Summary 更新编辑器配置
// @Description 更新指定类型编辑器的配置
// @Tags 配置
// @Accept json
// @Produce json
// @Param editorType path string true "编辑器类型"
// @Param request body services.UpdateEditorConfigRequest true "更新配置请求"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/config/{editorType} [put]
func (h *EditingHandler) UpdateEditorConfig(c *gin.Context) {
	editorType := c.Param("editorType")
	if editorType == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_editor_type",
			Message: "缺少编辑器类型",
		})
		return
	}

	var req services.UpdateEditorConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("绑定更新配置请求失败")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	req.EditorType = editorType

	err := h.editService.UpdateEditorConfig(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).WithField("editor_type", editorType).Error("更新编辑器配置失败")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "update_config_failed",
			Message: "更新编辑器配置失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "更新编辑器配置成功",
	})
}

// GetCollaborationStats 获取协作统计信息
// @Summary 获取协作统计信息
// @Description 获取当前协作会话的统计信息
// @Tags 协作
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/collaboration/stats [get]
func (h *EditingHandler) GetCollaborationStats(c *gin.Context) {
	stats := h.collabServer.GetRoomStats()

	c.JSON(http.StatusOK, SuccessResponse{
		Data:    stats,
		Message: "获取协作统计信息成功",
	})
}

// HandleWebSocket 处理WebSocket连接
// @Summary 建立WebSocket连接
// @Description 为协作编辑建立WebSocket连接
// @Tags WebSocket
// @Param documentId path string true "文档ID"
// @Param session_token query string true "会话令牌"
// @Param socket_id query string true "Socket ID"
// @Success 101 {string} string "WebSocket升级成功"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/ws/collaboration/{documentId} [get]
func (h *EditingHandler) HandleWebSocket(c *gin.Context) {
	h.collabServer.HandleWebSocket(c)
}

// validateCreateSessionRequest 验证创建会话请求
func (h *EditingHandler) validateCreateSessionRequest(req *services.CreateSessionRequest) error {
	if req.DocumentID == "" {
		return fmt.Errorf("文档ID不能为空")
	}
	if req.UserID == "" {
		return fmt.Errorf("用户ID不能为空")
	}
	if req.EditorType == "" {
		return fmt.Errorf("编辑器类型不能为空")
	}

	validEditorTypes := []string{"rich-text", "code", "markdown"}
	isValidEditorType := false
	for _, validType := range validEditorTypes {
		if req.EditorType == validType {
			isValidEditorType = true
			break
		}
	}
	if !isValidEditorType {
		return fmt.Errorf("无效的编辑器类型: %s", req.EditorType)
	}

	return nil
}