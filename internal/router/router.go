package router

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"law-oa-go/internal/handlers"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/services"
)

func Init(app *gin.Engine, db *gorm.DB, redisClient *redis.Client, esClient *elasticsearch.Client) {
	// 初始化服务
	userService := services.NewUserService(db)
	clientService := services.NewClientService(db)
	caseService := services.NewCaseService(db)

	// 初始化处理器
	authHandler := handlers.NewAuthHandler(userService)
	userHandler := handlers.NewUserHandler(userService)
	clientHandler := handlers.NewClientHandler(clientService)
	caseHandler := handlers.NewCaseHandler(caseService)

	// API路由组
	api := app.Group("/api/v1")

	// 健康检查
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Service is healthy",
		})
	})

	// 认证相关路由（无需认证）
	auth := api.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/register", authHandler.Register)
		auth.POST("/refresh", authHandler.RefreshToken)
	}

	// 需要认证的路由
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		// 用户相关路由
		user := protected.Group("/users")
		{
			user.GET("/profile", authHandler.GetProfile)
			user.PUT("/profile", authHandler.UpdateProfile)
			user.POST("/change-password", authHandler.ChangePassword)
			user.POST("/logout", authHandler.Logout)
		}

		// 管理员用户管理路由
		adminUsers := protected.Group("/admin/users")
		adminUsers.Use(middleware.RoleMiddleware("admin"))
		{
			adminUsers.GET("", userHandler.ListUsers)
			adminUsers.GET("/:id", userHandler.GetUser)
			adminUsers.POST("", userHandler.CreateUser)
			adminUsers.PUT("/:id", userHandler.UpdateUser)
			adminUsers.DELETE("/:id", userHandler.DeleteUser)
		}

		// 客户相关路由
		clients := protected.Group("/clients")
		{
			clients.GET("", clientHandler.ListClients)
			clients.GET("/:id", clientHandler.GetClient)
			clients.POST("", clientHandler.CreateClient)
			clients.PUT("/:id", clientHandler.UpdateClient)
			clients.DELETE("/:id", clientHandler.DeleteClient)
			clients.GET("/stats", clientHandler.GetClientStats)
		}

		// 案件相关路由
		cases := protected.Group("/cases")
		{
			cases.GET("", caseHandler.ListCases)
			cases.GET("/:id", caseHandler.GetCase)
			cases.POST("", caseHandler.CreateCase)
			cases.PUT("/:id", caseHandler.UpdateCase)
			cases.DELETE("/:id", caseHandler.DeleteCase)
			cases.GET("/stats", caseHandler.GetCaseStats)
			cases.POST("/:id/assign", caseHandler.AssignLawyer)
			cases.POST("/:id/status", caseHandler.UpdateCaseStatus)
		}

		// 文档相关路由（占位符）
		documents := protected.Group("/documents")
		{
			documents.GET("", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "Get all documents",
					"data":    []string{},
				})
			})

			documents.GET("/:id", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "Get document by ID",
					"data":    map[string]interface{}{},
				})
			})

			documents.POST("", func(c *gin.Context) {
				c.JSON(201, gin.H{
					"message": "Create document",
					"data":    map[string]interface{}{},
				})
			})

			documents.PUT("/:id", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "Update document",
					"data":    map[string]interface{}{},
				})
			})

			documents.DELETE("/:id", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "Delete document",
					"data":    map[string]interface{}{},
				})
			})
		}

		// 性能测试路由
		performance := protected.Group("/performance")
		performance.Use(middleware.RoleMiddleware("admin"))
		{
			performance.GET("/cache", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "Cache performance test",
					"data":    map[string]interface{}{},
				})
			})

			performance.GET("/database", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "Database performance test",
					"data":    map[string]interface{}{},
				})
			})
		}
	}

	// 404处理
	app.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"error": "Route not found",
		})
	})
}
