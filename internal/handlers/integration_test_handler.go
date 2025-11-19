package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// IntegrationTestHandler 集成测试处理器 - 提供无需认证的测试端点
type IntegrationTestHandler struct {
	db *gorm.DB
}

// NewIntegrationTestHandler 创建集成测试处理器
func NewIntegrationTestHandler(db *gorm.DB) *IntegrationTestHandler {
	return &IntegrationTestHandler{
		db: db,
	}
}

// TestIntegrationHealth 测试集成功能健康状态
func (h *IntegrationTestHandler) TestIntegrationHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "集成功能测试端点正常",
		"database": gin.H{
			"connected": h.db != nil,
		},
		"endpoints": []string{
			"GET  /test/integration/health - 健康检查",
			"POST /test/integration/demo - 演示集成功能",
			"POST /test/integration/conflict-demo - 演示冲突检测",
			"GET  /test/integration/db-connectivity - 数据库连接测试",
		},
	})
}

// TestIntegratedApprovalDemo 测试集成审批演示
func (h *IntegrationTestHandler) TestIntegratedApprovalDemo(c *gin.Context) {
	// 模拟请求数据
	mockRequest := gin.H{
		"type":            "案件代理",
		"title":           "集成测试演示申请",
		"content":         "这是一个演示集成工作流的测试申请",
		"applicant_name":  "测试律师",
		"department_name": "测试部门",
		"workflow_type":   "standard",
		"urgency":         "medium",
		"priority":        "normal",
		"conflict_check_config": gin.H{
			"user_id":               "test_lawyer_001",
			"client_ids":            []string{"test_client_001", "test_client_002"},
			"search_scope":          "all",
			"check_type":            "basic",
			"include_potential":     true,
			"mitigation_required":   false,
		},
		"metadata": gin.H{
			"test_mode":        true,
			"demo_mode":        true,
			"integration_test": true,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "集成审批演示请求模拟成功",
		"data": gin.H{
			"approval_id":     "demo_approval_" + time.Now().Format("20060102150405"),
			"status":          "created",
			"created_at":      time.Now().Format(time.RFC3339),
			"conflict_check": gin.H{
				"check_id":   "demo_check_" + time.Now().Format("20060102150405"),
				"status":     "completed",
				"has_conflict": false,
				"risk_level":  "LOW",
			},
		},
		"mock_request": mockRequest,
	})
}

// TestConflictCheckDemo 测试冲突检测演示
func (h *IntegrationTestHandler) TestConflictCheckDemo(c *gin.Context) {
	// 模拟请求数据
	mockRequest := gin.H{
		"user_id":        "test_lawyer_001",
		"client_ids":     []string{"demo_client_001", "demo_client_002"},
		"search_scope":   "all",
		"check_type":     "basic",
		"include_potential": true,
		"mitigation_required": false,
	}

	// 模拟响应数据
	mockResponse := gin.H{
		"check_id":     "demo_check_" + time.Now().Format("20060102150405"),
		"status":       "completed",
		"has_conflict": false,
		"conflict_count": 0,
		"risk_level":   "LOW",
		"risk_score":   15.5,
		"check_time":   time.Now().Format(time.RFC3339),
		"duration":     1250,
		"conflict_cases": []gin.H{},
		"recommendations": []string{
			"未发现明显冲突",
			"建议进行深度搜索以确保全面性",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "冲突检测演示成功",
		"data":    mockResponse,
		"mock_request": mockRequest,
	})
}

// TestDatabaseConnectivity 测试数据库连接
func (h *IntegrationTestHandler) TestDatabaseConnectivity(c *gin.Context) {
	// 执行简单查询测试连接
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
			"message": "获取底层数据库连接失败",
		})
		return
	}

	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
			"message": "数据库ping失败",
		})
		return
	}

	// 测试集成表是否存在
	var count int64
	if err := h.db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name IN ('conflict_check_associations', 'case_creation_associations', 'approval_integration_metadata')").Scan(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
			"message": "查询集成表失败",
		})
		return
	}

	// 测试基本表存在情况
	var basicTablesCount int64
	if err := h.db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name IN ('users', 'clients', 'cases', 'approval_requests')").Scan(&basicTablesCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
			"message": "查询基础表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "数据库连接正常",
		"database": gin.H{
			"connected": true,
			"driver":   "postgres",
			"status":   "healthy",
			"integration_tables": count,
			"basic_tables":     basicTablesCount,
			"total_tables":      count + basicTablesCount,
		},
	})
}

// TestIntegrationStatistics 测试集成统计演示
func (h *IntegrationTestHandler) TestIntegrationStatistics(c *gin.Context) {
	// 模拟统计数据
	mockStats := gin.H{
		"total_integrations":      42,
		"completed_integrations":  38,
		"pending_integrations":    4,
		"conflict_detection_rate": 95.5,
		"auto_case_creation_rate": 78.2,
		"average_processing_time": 1800, // 毫秒
		"success_rate":            90.5,
		"daily_stats": gin.H{
			"today": gin.H{
				"created":    8,
				"completed":  6,
				"failed":     2,
			},
		},
		"weekly_stats": gin.H{
			"created":    45,
			"completed":  42,
			"failed":     3,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "集成统计获取成功（演示数据）",
		"data":    mockStats,
	})
}

// RegisterTestRoutes 注册测试路由
func RegisterTestRoutes(app *gin.Engine, db *gorm.DB) {
	// 创建测试处理器
	testHandler := NewIntegrationTestHandler(db)

	// 注册测试路由组
	testGroup := app.Group("/test/integration")
	{
		testGroup.GET("/health", testHandler.TestIntegrationHealth)           // 健康检查
		testGroup.POST("/demo", testHandler.TestIntegratedApprovalDemo)      // 演示集成功能
		testGroup.POST("/conflict-demo", testHandler.TestConflictCheckDemo)  // 演示冲突检测
		testGroup.GET("/db-connectivity", testHandler.TestDatabaseConnectivity) // 数据库连接测试
		testGroup.GET("/statistics", testHandler.TestIntegrationStatistics)    // 集成统计
	}
}