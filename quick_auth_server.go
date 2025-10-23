package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// User 用户结构
type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Password string `json:"-"` // 不在JSON中返回
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// 模拟用户数据（用于测试）
var users = []User{
	{
		ID:       1,
		Username: "admin@lawoa.com",
		Name:     "系统管理员",
		Email:    "admin@lawoa.com",
		Role:     "admin",
		Password: "admin123",
	},
	{
		ID:       2,
		Username: "lawyer1@lawoa.com",
		Name:     "张律师",
		Email:    "lawyer1@lawoa.com",
		Role:     "lawyer",
		Password: "123456",
	},
	{
		ID:       3,
		Username: "lawyer2@lawoa.com",
		Name:     "李律师",
		Email:    "lawyer2@lawoa.com",
		Role:     "lawyer",
		Password: "123456",
	},
}

func main() {
	fmt.Println("🚀 启动快速认证服务...")
	fmt.Println("=================================================")

	// 创建Gin应用
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()
	app.Use(gin.Logger())
	app.Use(gin.Recovery())
	app.Use(corsMiddleware())

	// 添加路由
	setupRoutes(app)

	// 启动服务器
	go func() {
		fmt.Printf("🌐 服务器启动在: http://localhost:8080\n")
		fmt.Println("📋 可用端点:")
		fmt.Println("  POST /api/v1/auth/login - 用户登录")
		fmt.Println("  GET  /api/v1/auth/profile - 获取用户信息")
		fmt.Println("  GET  / - 服务信息")
		fmt.Println("=" + "="*49)

		if err := app.Run(":8080"); err != nil {
			fmt.Printf("❌ 服务器启动失败: %v\n", err)
		}
	}()

	// 等待信号
	select {}
}

// setupRoutes 设置路由
func setupRoutes(app *gin.Engine) {
	// 根路径
	app.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":     "快速认证服务",
			"version":     "1.0.0",
			"status":      "running",
			"description": "用于测试前端登录的简化认证服务",
			"endpoints": []string{
				"POST /api/v1/auth/login - 用户登录",
				"GET /api/v1/auth/profile - 获取用户信息",
			},
		})
	})

	// API路由组
	api := app.Group("/api/v1")
	{
		// 认证路由
		auth := api.Group("/auth")
		{
			auth.POST("/login", loginHandler)
			auth.GET("/profile", profileHandler)
		}

		// 模拟冲突检测路由（用于测试前端集成）
		conflict := api.Group("/conflict")
		{
			conflict.POST("/check", conflictCheckHandler)
			conflict.GET("/health", conflictHealthHandler)
		}

		// 模拟用户管理路由
		users := api.Group("/users")
		{
			users.GET("", listUsersHandler)
		}

		// 模拟案件管理路由
		cases := api.Group("/cases")
		{
			cases.GET("", listCasesHandler)
		}
	}
}

// corsMiddleware CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// loginHandler 登录处理器
func loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"data":    nil,
		})
		return
	}

	// 验证用户名和密码
	for _, user := range users {
		if user.Username == req.Username && user.Password == req.Password {
			// 生成简单的token（实际项目中应使用JWT）
			token := fmt.Sprintf("token_%d_%d", user.ID, time.Now().Unix())

			// 清除密码字段
			userResponse := User{
				ID:       user.ID,
				Username: user.Username,
				Name:     user.Name,
				Email:    user.Email,
				Role:     user.Role,
			}

			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "登录成功",
				"data": LoginResponse{
					Token: token,
					User:  userResponse,
				},
			})
			return
		}
	}

	// 用户名或密码错误
	c.JSON(http.StatusUnauthorized, gin.H{
		"code":    401,
		"message": "用户名或密码错误",
		"data":    nil,
	})
}

// profileHandler 用户信息处理器
func profileHandler(c *gin.Context) {
	// 模拟从token中获取用户信息
	// 实际项目中应从token中解析
	user := User{
		ID:       1,
		Username: "admin@lawoa.com",
		Name:     "系统管理员",
		Email:    "admin@lawoa.com",
		Role:     "admin",
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    user,
	})
}

// conflictCheckHandler 冲突检测处理器（模拟）
func conflictCheckHandler(c *gin.Context) {
	var request map[string]interface{}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"data":    nil,
		})
		return
	}

	// 模拟冲突检测结果
	response := map[string]interface{}{
		"checkId":     "CC_" + fmt.Sprintf("%d", time.Now().Unix()),
		"hasConflict":  true,
		"conflictCases": []map[string]interface{}{
			{
				"id":           "conflict_1",
				"caseName":     "腾讯诉字节跳动案",
				"conflictType":  "行业竞争冲突",
				"riskLevel":    "HIGH",
				"description":  "检测到同行业竞争关系",
			},
		},
		"riskAssessment": map[string]interface{}{
			"overallRisk":     "HIGH",
			"riskScore":       75.5,
			"requiresApproval": true,
			"recommendations": []string{
				"建议详细审查案件情况",
				"考虑是否需要回避",
			},
		},
		"checkTime": time.Now(),
		"duration":  "150ms",
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    response,
	})
}

// conflictHealthHandler 冲突检测健康检查
func conflictHealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": map[string]interface{}{
			"status":    "healthy",
			"service":   "conflict-check",
			"timestamp": time.Now(),
			"version":   "v1.0.0",
		},
	})
}

// listUsersHandler 用户列表处理器
func listUsersHandler(c *gin.Context) {
	// 返回用户列表（不包含密码）
	var userList []User
	for _, user := range users {
		userList = append(userList, User{
			ID:       user.ID,
			Username: user.Username,
			Name:     user.Name,
			Email:    user.Email,
			Role:     user.Role,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    userList,
	})
}

// listCasesHandler 案件列表处理器
func listCasesHandler(c *gin.Context) {
	// 模拟案件列表
	cases := []map[string]interface{}{
		{
			"id":          1,
			"title":       "腾讯诉字节跳动案",
			"description": "短视频平台版权纠纷",
			"clientName":  "腾讯科技",
			"lawyerName":  "张律师",
			"status":      "active",
			"caseType":    "知识产权",
		},
		{
			"id":          2,
			"title":       "阿里巴巴商业秘密案",
			"description": "电商平台商业秘密保护",
			"clientName":  "阿里巴巴",
			"lawyerName":  "张律师",
			"status":      "active",
			"caseType":    "商业纠纷",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    cases,
	})
}