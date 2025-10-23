package editing

import (
	"law-oa-go/internal/handlers"
	"law-oa-go/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RegisterRichEditorRoutes 注册富文本编辑器路由
func RegisterRichEditorRoutes(
	router *gin.RouterGroup,
	editService services.EditingService,
	storageService services.StorageService,
	notifyService services.NotificationService,
	logger *logrus.Logger,
) {
	// 创建富文本编辑器
	yjsProvider := NewYjsProvider()
	cursorManager := NewCursorManager()
	richEditor := NewRichTextEditor(
		editService,
		storageService,
		notifyService,
		yjsProvider,
		cursorManager,
	)

	// 创建处理器
	richEditorHandler := handlers.NewRichEditorHandler(richEditor, logger)

	// 注册路由
	richEditorGroup := router.Group("/rich-text")
	{
		// 编辑器初始化和管理
		richEditorGroup.POST("/:sessionId/initialize", richEditorHandler.InitializeEditor)
		richEditorGroup.POST("/:sessionId/destroy", richEditorHandler.DestroyEditor)

	// 内容操作
		richEditorGroup.GET("/:sessionId/content", richEditorHandler.LoadContent)
		richEditorGroup.POST("/:sessionId/content", richEditorHandler.SaveContent)
		richEditorGroup.POST("/:sessionId/convert/html", richEditorHandler.ConvertToHTML)
		richEditorGroup.POST("/:sessionId/convert/plain", richEditorHandler.ConvertToPlainText)

	// 编辑操作
		richEditorGroup.POST("/:sessionId/operations", richEditorHandler.HandleOperation)

	// 光标管理
		richEditorGroup.PUT("/:sessionId/cursor", richEditorHandler.UpdateCursor)
		richEditorGroup.GET("/:sessionId/cursors", richEditorHandler.GetCursors)

	// 统计信息
		richEditorGroup.GET("/:sessionId/stats", richEditorHandler.GetEditorStats)

		// WebSocket支持（如果有的话）
		// richEditorGroup.GET("/:sessionId/ws", richEditorHandler.HandleWebSocket)
	}

	logger.Info("富文本编辑器路由注册完成")
}

// YjsProviderImpl Yjs提供者实现
type YjsProviderImpl struct {
	documents map[string]YjsDocument
	// 其他需要的字段
}

// NewYjsProvider 创建Yjs提供者
func NewYjsProvider() YjsProvider {
	return &YjsProviderImpl{
		documents: make(map[string]YjsDocument),
	}
}

// CursorManagerImpl 光标管理器实现
type CursorManagerImpl struct {
	cursors map[string]*CursorInfo
	// 其他需要的字段
}

// NewCursorManager 创建光标管理器
func NewCursorManager() CursorManager {
	return &CursorManagerImpl{
		cursors: make(map[string]*CursorInfo),
	}
}