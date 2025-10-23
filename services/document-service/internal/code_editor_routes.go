package editing

import (
	"law-oa-go/internal/handlers"
	"law-oa-go/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RegisterCodeEditorRoutes 注册代码编辑器路由
func RegisterCodeEditorRoutes(
	router *gin.RouterGroup,
	editService services.EditingService,
	storageService services.StorageService,
	notifyService services.NotificationService,
	logger *logrus.Logger,
) {
	// 创建代码编辑器
	codeEditor := NewCodeEditor(
		editService,
		storageService,
		notifyService,
	)

	// 创建处理器
	codeEditorHandler := handlers.NewCodeEditorHandler(codeEditor, logger)

	// 注册路由
	codeEditorGroup := router.Group("/code-editor")
	{
		// 编辑器初始化和管理
		codeEditorGroup.POST("/:sessionId/initialize", codeEditorHandler.InitializeCodeEditor)
		codeEditorGroup.POST("/:sessionId/destroy", codeEditorHandler.DestroyEditor)

		// 代码内容操作
		codeEditorGroup.GET("/:sessionId/code", codeEditorHandler.LoadCode)
		codeEditorGroup.POST("/:sessionId/code", codeEditorHandler.SaveCode)

		// 编辑操作
		codeEditorGroup.POST("/:sessionId/operations", codeEditorHandler.HandleOperation)

		// 语言服务功能
		codeEditorGroup.GET("/:sessionId/diagnostics", codeEditorHandler.GetDiagnostics)
		codeEditorGroup.POST("/:sessionId/completions", codeEditorHandler.GetCompletions)
		codeEditorGroup.POST("/:sessionId/hover", codeEditorHandler.GetHoverInfo)
		codeEditorGroup.POST("/:sessionId/format", codeEditorHandler.FormatCode)
		codeEditorGroup.GET("/:sessionId/language-config", codeEditorHandler.GetLanguageServiceConfig)

		// 统计信息
		codeEditorGroup.GET("/:sessionId/stats", codeEditorHandler.GetEditorStats)

		// WebSocket支持（如果有的话）
		// codeEditorGroup.GET("/:sessionId/ws", codeEditorHandler.HandleWebSocket)
	}

	logger.Info("代码编辑器路由注册完成")
}