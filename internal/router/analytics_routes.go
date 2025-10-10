package router

import (
	"github.com/gin-gonic/gin"
	"law-oa-go/internal/handlers"
)

// SetupAnalyticsRoutes 设置分析相关路由
func SetupAnalyticsRoutes(r *gin.Engine, analyticsHandler *handlers.AnalyticsHandler, authMiddleware gin.HandlerFunc) {
	// 分析路由组
	analyticsGroup := r.Group("/api/v1/analytics")
	analyticsGroup.Use(authMiddleware) // 需要认证
	// analyticsGroup.Use(middleware.RequestID()) // 添加请求ID (暂时禁用)
	{
		// 会话管理
		sessions := analyticsGroup.Group("/sessions")
		{
			sessions.POST("", analyticsHandler.CreateSession)           // 创建会话
			sessions.GET("/:id", analyticsHandler.GetSession)          // 获取会话详情
			sessions.PUT("/:id", analyticsHandler.UpdateSession)       // 更新会话
		}

		// 页面浏览追踪
		pageViews := analyticsGroup.Group("/page-views")
		{
			pageViews.POST("", analyticsHandler.TrackPageView) // 追踪页面浏览
		}

		// 事件追踪
		events := analyticsGroup.Group("/events")
		{
			events.POST("", analyticsHandler.TrackEvent)        // 追踪单个事件
			events.POST("/batch", analyticsHandler.BatchTrackEvents) // 批量追踪事件
		}

		// 用户旅程管理
		journeys := analyticsGroup.Group("/journeys")
		{
			journeys.POST("", analyticsHandler.CreateJourney) // 创建用户旅程
		}

		// 行为分析
		analysis := analyticsGroup.Group("/analysis")
		{
			analysis.GET("/users/:user_id/behavior", analyticsHandler.GetUserBehaviorAnalysis) // 获取用户行为分析
			analysis.POST("/users/:user_id/patterns", analyticsHandler.DetectBehaviorPatterns) // 检测行为模式
		}

		// 统计数据
		stats := analyticsGroup.Group("/stats")
		{
			stats.GET("/realtime/dashboard", analyticsHandler.GetRealTimeDashboard) // 获取实时仪表板
			stats.POST("/realtime/update", analyticsHandler.UpdateRealTimeStats)    // 更新实时统计
			stats.GET("/users/daily-active", analyticsHandler.GetDailyActiveUsers)  // 获取日活跃用户
			stats.GET("/page-views", analyticsHandler.GetPageViewStats)             // 获取页面浏览统计
			stats.GET("/events", analyticsHandler.GetEventStats)                    // 获取事件统计
		}
	}

	// 公开的分析路由（不需要认证，用于数据收集等）
	publicAnalyticsGroup := r.Group("/api/v1/analytics/public")
	{
		publicAnalyticsGroup.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "healthy",
				"service": "analytics-engine",
				"version": "1.0.0",
			})
		})
	}
}

// RegisterAnalyticsSwaggerDocs 注册分析相关的Swagger文档
func RegisterAnalyticsSwaggerDocs() {
	// 这里可以添加Swagger文档注册逻辑
	// 例如使用swaggo库注册API文档
}