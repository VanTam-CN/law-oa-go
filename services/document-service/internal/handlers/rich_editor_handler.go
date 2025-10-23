package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"law-oa-go/internal/editing"
	"law-oa-go/internal/models"
	"law-oa-go/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RichEditorHandler 富文本编辑器处理器
type RichEditorHandler struct {
	richEditor *editing.RichTextEditor
	logger     *logrus.Logger
}

// NewRichEditorHandler 创建富文本编辑器处理器
func NewRichEditorHandler(
	richEditor *editing.RichTextEditor,
	logger *logrus.Logger,
) *RichEditorHandler {
	return &RichEditorHandler{
		richEditor: richEditor,
		logger:     logger,
	}
}

// InitializeEditor 初始化富文本编辑器
// @Summary 初始化富文本编辑器
// @Description 初始化指定会话的富文本编辑器配置
// @Tags 富文本编辑器
// @Accept json
// @Produce json
// @Param sessionId path string true "会话ID"
// @Success 200 {object} editing.RichTextConfig
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/rich-text/{sessionId}/initialize [post]
func (h *RichEditorHandler) InitializeEditor(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_session_id",
			Message: "缺少会话ID",
		})
		return
	}

	config, err := h.richEditor.InitializeEditor(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("初始化富文本编辑器失败")

		if err.Error() == "获取编辑会话失败" || err.Error() == "获取编辑器配置失败" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "resource_not_found",
				Message: "会话或配置不存在",
				Details: err.Error(),
			})
			return
		}

		if err.Error() == "没有编辑权限" {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "permission_denied",
				Message: "没有编辑权限",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "initialize_editor_failed",
			Message: "初始化富文本编辑器失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data:    config,
		Message: "初始化富文本编辑器成功",
	})
}

// LoadContent 加载文档内容
// @Summary 加载文档内容
// @Description 加载指定会话的文档内容
// @Tags 富文本编辑器
// @Accept json
// @Produce json
// @Param sessionId path string true "会话ID"
// @Success 200 {object} editing.RichTextContent
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/rich-text/{sessionId}/content [get]
func (h *RichEditorHandler) LoadContent(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_session_id",
			Message: "缺少会话ID",
		})
		return
	}

	content, err := h.richEditor.LoadContent(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("加载文档内容失败")

		if err.Error() == "获取编辑会话失败" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "session_not_found",
				Message: "编辑会话不存在",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "load_content_failed",
			Message: "加载文档内容失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data:    content,
		Message: "加载文档内容成功",
	})
}

// SaveContent 保存文档内容
// @Summary 保存文档内容
// @Description 保存指定会话的文档内容
// @Tags 富文本编辑器
// @Accept json
// @Produce json
// @Param sessionId path string true "会话ID"
// @Param request body editing.RichTextContent true "文档内容"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/rich-text/{sessionId}/content [post]
func (h *RichEditorHandler) SaveContent(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_session_id",
			Message: "缺少会话ID",
		})
		return
	}

	var content editing.RichTextContent
	if err := c.ShouldBindJSON(&content); err != nil {
		h.logger.WithError(err).Error("绑定保存内容请求失败")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	// 验证内容
	if err := h.validateRichTextContent(&content); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "内容验证失败",
			Details: err.Error(),
		})
		return
	}

	err := h.richEditor.SaveContent(c.Request.Context(), sessionID, &content)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("保存文档内容失败")

		if err.Error() == "获取编辑会话失败" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "session_not_found",
				Message: "编辑会话不存在",
				Details: err.Error(),
			})
			return
		}

		if err.Error() == "没有编辑权限" {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "permission_denied",
				Message: "没有编辑权限",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "save_content_failed",
			Message: "保存文档内容失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "保存文档内容成功",
	})
}

// HandleOperation 处理编辑操作
// @Summary 处理编辑操作
// @Description 处理富文本编辑器的编辑操作
// @Tags 富文本编辑器
// @Accept json
// @Produce json
// @Param sessionId path string true "会话ID"
// @Param request body editing.RichTextOperation true "编辑操作"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/rich-text/{sessionId}/operations [post]
func (h *RichEditorHandler) HandleOperation(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_session_id",
			Message: "缺少会话ID",
		})
		return
	}

	var operation editing.RichTextOperation
	if err := c.ShouldBindJSON(&operation); err != nil {
		h.logger.WithError(err).Error("绑定编辑操作请求失败")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	// 验证操作
	if err := h.validateRichTextOperation(&operation); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "操作验证失败",
			Details: err.Error(),
		})
		return
	}

	err := h.richEditor.HandleOperation(c.Request.Context(), sessionID, &operation)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"session_id":    sessionID,
			"operation_type": operation.Type,
			"user_id":       operation.UserID,
		}).Error("处理编辑操作失败")

		if err.Error() == "获取编辑会话失败" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "session_not_found",
				Message: "编辑会话不存在",
				Details: err.Error(),
			})
			return
		}

		if err.Error() == "没有编辑权限" {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "permission_denied",
				Message: "没有编辑权限",
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

// UpdateCursor 更新光标位置
// @Summary 更新光标位置
// @Description 更新指定用户在文档中的光标位置
// @Tags 富文本编辑器
// @Accept json
// @Produce json
// @Param sessionId path string true "会话ID"
// @Param request body editing.CursorInfo true "光标信息"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/rich-text/{sessionId}/cursor [put]
func (h *RichEditorHandler) UpdateCursor(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_session_id",
			Message: "缺少会话ID",
		})
		return
	}

	var cursor editing.CursorInfo
	if err := c.ShouldBindJSON(&cursor); err != nil {
		h.logger.WithError(err).Error("绑定光标更新请求失败")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	err := h.richEditor.UpdateCursor(c.Request.Context(), sessionID, &cursor)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"session_id": sessionID,
			"user_id":    cursor.UserID,
		}).Error("更新光标位置失败")

		if err.Error() == "获取编辑会话失败" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "session_not_found",
				Message: "编辑会话不存在",
				Details: err.Error(),
			})
			return
		}

		if err.Error() == "没有查看权限" {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "permission_denied",
				Message: "没有查看权限",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "update_cursor_failed",
			Message: "更新光标位置失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "更新光标位置成功",
	})
}

// GetCursors 获取所有光标
// @Summary 获取所有光标
// @Description 获取文档中所有协作者的光标位置
// @Tags 富文本编辑器
// @Accept json
// @Produce json
// @Param sessionId path string true "会话ID"
// @Success 200 {object} []editing.CursorInfo
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/rich-text/{sessionId}/cursors [get]
func (h *RichEditorHandler) GetCursors(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_session_id",
			Message: "缺少会话ID",
		})
		return
	}

	cursors, err := h.richEditor.GetCursors(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("获取光标列表失败")

		if err.Error() == "获取编辑会话失败" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "session_not_found",
				Message: "编辑会话不存在",
				Details: err.Error(),
			})
			return
		}

		if err.Error() == "没有查看权限" {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "permission_denied",
				Message: "没有查看权限",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "get_cursors_failed",
			Message: "获取光标列表失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data:    cursors,
		Message: "获取光标列表成功",
	})
}

// ConvertToHTML 转换为HTML
// @Summary 转换为HTML
// @Description 将Delta格式转换为HTML格式
// @Tags 富文本编辑器
// @Accept json
// @Produce json
// @Param sessionId path string true "会话ID"
// @Param request body map[string]interface{} true "Delta数据"
// @Success 200 {object} ConvertResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/rich-text/{sessionId}/convert/html [post]
func (h *RichEditorHandler) ConvertToHTML(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_session_id",
			Message: "缺少会话ID",
		})
		return
	}

	var request map[string]interface{}
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.WithError(err).Error("绑定转换请求失败")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	deltaData, ok := request["delta"]
	if !ok {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_delta",
			Message: "缺少Delta数据",
		})
		return
	}

	// 这里需要将interface{}转换为models.DeltaData
	// 简化实现，实际中需要完整的类型转换
	delta := &models.DeltaData{}
	if deltaDataMap, ok := deltaData.(map[string]interface{}); ok {
		if ops, ok := deltaDataMap["ops"].([]interface{}); ok {
			delta.Ops = ops
		}
	}

	html, err := h.richEditor.ConvertToHTML(sessionID, delta)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("转换为HTML失败")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "convert_to_html_failed",
			Message: "转换为HTML失败",
			Details: err.Error(),
		})
		return
	}

	response := ConvertResponse{
		Format:  "html",
		Content: html,
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data:    response,
		Message: "转换为HTML成功",
	})
}

// ConvertToPlainText 转换为纯文本
// @Summary 转换为纯文本
// @Description 将Delta格式转换为纯文本格式
// @Tags 富文本编辑器
// @Accept json
// @Produce json
// @Param sessionId path string true "会话ID"
// @Param request body map[string]interface{} true "Delta数据"
// @Success 200 {object} ConvertResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/rich-text/{sessionId}/convert/plain [post]
func (h *RichEditorHandler) ConvertToPlainText(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_session_id",
			Message: "缺少会话ID",
		})
		return
	}

	var request map[string]interface{}
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.WithError(err).Error("绑定转换请求失败")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	deltaData, ok := request["delta"]
	if !ok {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_delta",
			Message: "缺少Delta数据",
		})
		return
	}

	// 这里需要将interface{}转换为models.DeltaData
	// 简化实现，实际中需要完整的类型转换
	delta := &models.DeltaData{}
	if deltaDataMap, ok := deltaData.(map[string]interface{}); ok {
		if ops, ok := deltaDataMap["ops"].([]interface{}); ok {
			delta.Ops = ops
		}
	}

	text, err := h.richEditor.ConvertToPlainText(sessionID, delta)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("转换为纯文本失败")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "convert_to_plain_text_failed",
			Message: "转换为纯文本失败",
			Details: err.Error(),
		})
		return
	}

	response := ConvertResponse{
		Format:  "plain_text",
		Content: text,
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data:    response,
		Message: "转换为纯文本成功",
	})
}

// GetEditorStats 获取编辑器统计信息
// @Summary 获取编辑器统计信息
// @Description 获取富文本编辑器的统计信息
// @Tags 富文本编辑器
// @Accept json
// @Produce json
// @Param sessionId path string true "会话ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/rich-text/{sessionId}/stats [get]
func (h *RichEditorHandler) GetEditorStats(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_session_id",
			Message: "缺少会话ID",
		})
		return
	}

	stats, err := h.richEditor.GetEditorStats(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("获取编辑器统计失败")

		if err.Error() == "获取编辑会话失败" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "session_not_found",
				Message: "编辑会话不存在",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "get_stats_failed",
			Message: "获取编辑器统计失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data:    stats,
		Message: "获取编辑器统计成功",
	})
}

// DestroyEditor 销毁编辑器
// @Summary 销毁编辑器
// @Description 销毁指定会话的富文本编辑器
// @Tags 富文本编辑器
// @Accept json
// @Produce json
// @Param sessionId path string true "会话ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/rich-text/{sessionId}/destroy [post]
func (h *RichEditorHandler) DestroyEditor(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_session_id",
			Message: "缺少会话ID",
		})
		return
	}

	err := h.richEditor.DestroyEditor(sessionID)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("销毁富文本编辑器失败")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "destroy_editor_failed",
			Message: "销毁富文本编辑器失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "销毁富文本编辑器成功",
	})
}

// ConvertResponse 转换响应
type ConvertResponse struct {
	Format  string `json:"format"`
	Content string `json:"content"`
}

// 验证方法

// validateRichTextContent 验证富文本内容
func (h *RichEditorHandler) validateRichTextContent(content *editing.RichTextContent) error {
	if content == nil {
		return fmt.Errorf("内容不能为空")
	}

	if content.Delta == nil {
		return fmt.Errorf("Delta数据不能为空")
	}

	if content.Length < 0 {
		return fmt.Errorf("内容长度不能为负数")
	}

	return nil
}

// validateRichTextOperation 验证富文本操作
func (h *RichEditorHandler) validateRichTextOperation(operation *editing.RichTextOperation) error {
	if operation == nil {
		return fmt.Errorf("操作不能为空")
	}

	if operation.Type == "" {
		return fmt.Errorf("操作类型不能为空")
	}

	validTypes := []string{"insert", "delete", "retain", "format", "cursor", "selection"}
	isValidType := false
	for _, validType := range validTypes {
		if operation.Type == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return fmt.Errorf("无效的操作类型: %s", operation.Type)
	}

	if operation.UserID == "" {
		return fmt.Errorf("用户ID不能为空")
	}

	if operation.SessionID == "" {
		return fmt.Errorf("会话ID不能为空")
	}

	return nil
}