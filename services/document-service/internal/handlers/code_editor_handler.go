package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"law-oa-go/internal/editing"

	"github.com/gin-gonic/gin"
)

// CodeEditorHandler 代码编辑器处理器
type CodeEditorHandler struct {
	codeEditor editing.CodeEditor
	logger     *logrus.Logger
}

// NewCodeEditorHandler 创建代码编辑器处理器
func NewCodeEditorHandler(
	codeEditor editing.CodeEditor,
	logger *logrus.Logger,
) *CodeEditorHandler {
	return &CodeEditorHandler{
		codeEditor: codeEditor,
		logger:     logger,
	}
}

// InitializeCodeEditor 初始化代码编辑器
// @Summary 初始化代码编辑器
// @Description 初始化指定会话的代码编辑器配置
// @Tags 代码编辑器
// @Accept json
// @Produce json
// @Param sessionId path string true "会话ID"
// @Success 200 {object} editing.CodeEditorConfig
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/code-editor/{sessionId}/initialize [post]
func (h *CodeEditorHandler) InitializeCodeEditor(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_session_id",
			Message: "缺少会话ID",
		})
		return
	}

	config, err := h.codeEditor.InitializeEditor(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("初始化代码编辑器失败")

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
			Message: "初始化代码编辑器失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data:    config,
		Message: "初始化代码编辑器成功",
	})
}

// LoadCode 加载代码内容
// @Summary 加载代码内容
// @Description 加载指定会话的代码内容
// @Tags 代码编辑器
// @Accept json
// @Produce json
// @Param sessionId path string true "会话ID"
// @Success 200 {object} editing.CodeContent
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/code-editor/{sessionId}/code [get]
func (h *CodeEditorHandler) LoadCode(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_session_id",
			Message: "缺少会话ID",
		})
		return
	}

	content, err := h.codeEditor.LoadCode(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("加载代码内容失败")

		if err.Error() == "会话不存在" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "session_not_found",
				Message: "编辑会话不存在",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "load_code_failed",
			Message: "加载代码内容失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data:    content,
		Message: "加载代码内容成功",
	})
}

// SaveCode 保存代码内容
// @Summary 保存代码内容
// @Description 保存指定会话的代码内容
// @Tags 代码编辑器
// @Accept json
// @Produce json
// @Param sessionId path string true "会话ID"
// @Param request body editing.CodeContent true "代码内容"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/code-editor/{sessionId}/code [post]
func (h *CodeEditorHandler) SaveCode(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_session_id",
			Message: "缺少会话ID",
		})
		return
	}

	var content editing.CodeContent
	if err := c.ShouldBindJSON(&content); err != nil {
		h.logger.WithError(err).Error("绑定保存代码请求失败")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	// 验证内容
	if err := h.validateCodeContent(&content); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "内容验证失败",
			Details: err.Error(),
		})
		return
	}

	err := h.codeEditor.SaveCode(c.Request.Context(), sessionID, &content)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("保存代码内容失败")

		if err.Error() == "会话不存在" {
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
			Error:   "save_code_failed",
			Message: "保存代码内容失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "保存代码内容成功",
	})
}

// HandleOperation 处理编辑操作
// @Summary 处理编辑操作
// @Description 处理代码编辑器的编辑操作
// @Tags 代码编辑器
// @Accept json
// @Produce json
// @Param sessionId path string true "会话ID"
// @Param request body editing.CodeOperation true "编辑操作"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/code-editor/{sessionId}/operations [post]
func (h *CodeEditorHandler) HandleOperation(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_session_id",
			Message: "缺少会话ID",
		})
		return
	}

	var operation editing.CodeOperation
	if err := c.ShouldBindJSON(&operation); err != nil {
		h.logger.WithError(err).Error("绑定编辑操作请求失败")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	// 设置会话ID
	operation.SessionID = sessionID

	// 验证操作
	if err := h.validateCodeOperation(&operation); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "操作验证失败",
			Details: err.Error(),
		})
		return
	}

	err := h.codeEditor.HandleOperation(c.Request.Context(), sessionID, &operation)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"session_id":    sessionID,
			"operation_type": operation.Type,
			"user_id":       operation.UserID,
		}).Error("处理编辑操作失败")

		if err.Error() == "会话不存在" {
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

// GetDiagnostics 获取诊断信息
// @Summary 获取诊断信息
// @Description 获取代码编辑器的诊断信息
// @Tags 代码编辑器
// @Accept json
// @Produce json
// @Param sessionId path string true "会话ID"
// @Success 200 {object} []editing.DiagnosticInfo
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/code-editor/{sessionId}/diagnostics [get]
func (h *CodeEditorHandler) GetDiagnostics(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_session_id",
			Message: "缺少会话ID",
		})
		return
	}

	diagnostics, err := h.codeEditor.GetDiagnostics(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("获取诊断信息失败")

		if err.Error() == "会话不存在" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "session_not_found",
				Message: "编辑会话不存在",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "get_diagnostics_failed",
			Message: "获取诊断信息失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data:    diagnostics,
		Message: "获取诊断信息成功",
	})
}

// GetCompletions 获取代码补全
// @Summary 获取代码补全
// @Description 获取指定位置的代码补全建议
// @Tags 代码编辑器
// @Accept json
// @Produce json
// @Param sessionId path string true "会话ID"
// @Param request body editing.Position true "位置信息"
// @Success 200 {object} []editing.CompletionItem
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/code-editor/{sessionId}/completions [post]
func (h *CodeEditorHandler) GetCompletions(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_session_id",
			Message: "缺少会话ID",
		})
		return
	}

	var position editing.Position
	if err := c.ShouldBindJSON(&position); err != nil {
		h.logger.WithError(err).Error("绑定位置请求失败")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	// 验证位置
	if err := h.validatePosition(&position); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "位置验证失败",
			Details: err.Error(),
		})
		return
	}

	completions, err := h.codeEditor.GetCompletions(c.Request.Context(), sessionID, &position)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("获取代码补全失败")

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "get_completions_failed",
			Message: "获取代码补全失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data:    completions,
		Message: "获取代码补全成功",
	})
}

// GetHoverInfo 获取悬停信息
// @Summary 获取悬停信息
// @Description 获取指定位置的悬停信息
// @Tags 代码编辑器
// @Accept json
// @Produce json
// @Param sessionId path string true "会话ID"
// @Param request body editing.Position true "位置信息"
// @Success 200 {object} editing.HoverInfo
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/code-editor/{sessionId}/hover [post]
func (h *CodeEditorHandler) GetHoverInfo(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_session_id",
			Message: "缺少会话ID",
		})
		return
	}

	var position editing.Position
	if err := c.ShouldBindJSON(&position); err != nil {
		h.logger.WithError(err).Error("绑定位置请求失败")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	// 验证位置
	if err := h.validatePosition(&position); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "位置验证失败",
			Details: err.Error(),
		})
		return
	}

	hoverInfo, err := h.codeEditor.GetHoverInfo(c.Request.Context(), sessionID, &position)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("获取悬停信息失败")

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "get_hover_info_failed",
			Message: "获取悬停信息失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data:    hoverInfo,
		Message: "获取悬停信息成功",
	})
}

// FormatCode 格式化代码
// @Summary 格式化代码
// @Description 格式化指定范围的代码
// @Tags 代码编辑器
// @Accept json
// @Produce json
// @Param sessionId path string true "会话ID"
// @Param request body editing.Range true "格式化范围"
// @Success 200 {object} editing.TextEdit
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/code-editor/{sessionId}/format [post]
func (h *CodeEditorHandler) FormatCode(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_session_id",
			Message: "缺少会话ID",
		})
		return
	}

	var range_ editing.Range
	if err := c.ShouldBindJSON(&range_); err != nil {
		h.logger.WithError(err).Error("绑定范围请求失败")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	// 验证范围
	if err := h.validateRange(&range_); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "范围验证失败",
			Details: err.Error(),
		})
		return
	}

	textEdit, err := h.codeEditor.FormatCode(c.Request.Context(), sessionID, &range_)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("格式化代码失败")

		if err.Error() == "会话不存在" {
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
			Error:   "format_code_failed",
			Message: "格式化代码失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data:    textEdit,
		Message: "格式化代码成功",
	})
}

// GetLanguageServiceConfig 获取语言服务配置
// @Summary 获取语言服务配置
// @Description 获取代码编辑器的语言服务配置
// @Tags 代码编辑器
// @Accept json
// @Produce json
// @Param sessionId path string true "会话ID"
// @Success 200 {object} editing.LanguageServiceConfig
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/code-editor/{sessionId}/language-config [get]
func (h *CodeEditorHandler) GetLanguageServiceConfig(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_session_id",
			Message: "缺少会话ID",
		})
		return
	}

	config, err := h.codeEditor.GetLanguageServiceConfig(sessionID)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("获取语言服务配置失败")

		if err.Error() == "会话不存在" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "session_not_found",
				Message: "编辑会话不存在",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "get_language_config_failed",
			Message: "获取语言服务配置失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data:    config,
		Message: "获取语言服务配置成功",
	})
}

// DestroyEditor 销毁编辑器
// @Summary 销毁编辑器
// @Description 销毁指定会话的代码编辑器
// @Tags 代码编辑器
// @Accept json
// @Produce json
// @Param sessionId path string true "会话ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/code-editor/{sessionId}/destroy [post]
func (h *CodeEditorHandler) DestroyEditor(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_session_id",
			Message: "缺少会话ID",
		})
		return
	}

	err := h.codeEditor.DestroyEditor(sessionID)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("销毁代码编辑器失败")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "destroy_editor_failed",
			Message: "销毁代码编辑器失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "销毁代码编辑器成功",
	})
}

// GetEditorStats 获取编辑器统计信息
// @Summary 获取编辑器统计信息
// @Description 获取代码编辑器的统计信息
// @Tags 代码编辑器
// @Accept json
// @Produce json
// @Param sessionId path string true "会话ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/editing/code-editor/{sessionId}/stats [get]
func (h *CodeEditorHandler) GetEditorStats(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_session_id",
			Message: "缺少会话ID",
		})
		return
	}

	// 这里应该从编辑器获取统计信息
	// 暂时返回基本信息
	stats := map[string]interface{}{
		"session_id":   sessionID,
		"editor_type":  "code",
		"features": []string{
			"syntax_highlighting",
			"code_completion",
			"error_diagnostics",
			"code_formatting",
			"intellisense",
		},
		"supported_languages": []string{
			"go",
			"typescript",
			"javascript",
			"json",
			"css",
			"html",
		},
		"performance": map[string]interface{}{
			"diagnostics_enabled": true,
			"completion_enabled":  true,
			"formatting_enabled":  true,
		},
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data:    stats,
		Message: "获取编辑器统计成功",
	})
}

// 验证方法

// validateCodeContent 验证代码内容
func (h *CodeEditorHandler) validateCodeContent(content *editing.CodeContent) error {
	if content == nil {
		return fmt.Errorf("内容不能为空")
	}

	if content.URI == "" {
		return fmt.Errorf("文档URI不能为空")
	}

	if content.Language == "" {
		return fmt.Errorf("语言不能为空")
	}

	if content.Version < 0 {
		return fmt.Errorf("版本号不能为负数")
	}

	return nil
}

// validateCodeOperation 验证代码操作
func (h *CodeEditorHandler) validateCodeOperation(operation *editing.CodeOperation) error {
	if operation == nil {
		return fmt.Errorf("操作不能为空")
	}

	if operation.Type == "" {
		return fmt.Errorf("操作类型不能为空")
	}

	validTypes := []string{"insert", "delete", "replace", "format", "complete", "hover", "diagnostic"}
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

	if operation.Version < 0 {
		return fmt.Errorf("版本号不能为负数")
	}

	// 根据操作类型验证其他字段
	switch operation.Type {
	case "insert", "delete", "replace":
		if operation.Range == nil {
			return fmt.Errorf("范围操作需要指定范围")
		}
	case "complete", "hover":
		if operation.Position == nil {
			return fmt.Errorf("补全和悬停操作需要指定位置")
		}
	}

	return nil
}

// validatePosition 验证位置
func (h *CodeEditorHandler) validatePosition(position *editing.Position) error {
	if position == nil {
		return fmt.Errorf("位置不能为空")
	}

	if position.Line < 0 {
		return fmt.Errorf("行号不能为负数")
	}

	if position.Character < 0 {
		return fmt.Errorf("字符位置不能为负数")
	}

	return nil
}

// validateRange 验证范围
func (h *CodeEditorHandler) validateRange(range_ *editing.Range) error {
	if range_ == nil {
		return fmt.Errorf("范围不能为空")
	}

	if err := h.validatePosition(range_.Start); err != nil {
		return fmt.Errorf("起始位置无效: %w", err)
	}

	if err := h.validatePosition(range_.End); err != nil {
		return fmt.Errorf("结束位置无效: %w", err)
	}

	// 验证范围的合理性
	if range_.Start.Line > range_.End.Line ||
		(range_.Start.Line == range_.End.Line && range_.Start.Character > range_.End.Character) {
		return fmt.Errorf("起始位置不能大于结束位置")
	}

	return nil
}