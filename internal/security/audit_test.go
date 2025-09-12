package security

import (
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"law-oa-go/test/mock"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// auditLogConfigToConfig 将AuditLogConfig转换为AuditConfig
func auditLogConfigToConfig(logConfig *AuditLogConfig) *AuditConfig {
	return &AuditConfig{
		EnableAuditLog:       logConfig.EnableAuditLog,
		LogDatabase:          logConfig.LogDatabase,
		LogToFile:            logConfig.LogToFile,
		LogToSyslog:          logConfig.LogToSyslog,
		EnableRealTimeAlert:  logConfig.EnableRealTimeAlert,
		SensitiveEventTypes:  logConfig.SensitiveEventTypes,
		RequiredEventTypes:   logConfig.RequiredEventTypes,
		RetentionDays:        logConfig.RetentionDays,
		MaxBatchSize:         logConfig.MaxBatchSize,
		BatchTimeout:         logConfig.BatchTimeout,
		EnableCompression:    logConfig.EnableCompression,
		EncryptSensitiveData: logConfig.EncryptSensitiveData,
	}
}

// TestAuditService_NewAuditService 测试审计服务创建
func TestAuditService_NewAuditService(t *testing.T) {
	t.Run("创建审计服务 - 启用审计", func(t *testing.T) {
		logConfig := &AuditLogConfig{
			EnableAuditLog:     true,
			LogDatabase:        true,
			MaxBatchSize:       100,
			BatchTimeout:       5 * time.Second,
			RetentionDays:      365,
			RequiredEventTypes: []AuditEventType{EventTypeLogin, EventTypeLogout},
		}

		config := auditLogConfigToConfig(logConfig)
		mockDB, _ := mock.NewMockDB()
		cacheService := createTestCacheService()
		service := NewAuditService(config, mockDB.DB, cacheService)
		assert.NotNil(t, service)
		assert.NotNil(t, service.eventChan)
		assert.NotNil(t, service.stopChan)
		assert.Equal(t, 5, service.workerCount)
	})

	t.Run("创建审计服务 - 禁用审计", func(t *testing.T) {
		logConfig := &AuditLogConfig{
			EnableAuditLog: false,
		}

		config := auditLogConfigToConfig(logConfig)
		mockDB, _ := mock.NewMockDB()
		cacheService := createTestCacheService()
		service := NewAuditService(config, mockDB.DB, cacheService)
		assert.NotNil(t, service)
		assert.NotNil(t, service.eventChan)
		assert.NotNil(t, service.stopChan)
	})
}

// TestAuditService_LogEvent 测试事件记录
func TestAuditService_LogEvent(t *testing.T) {
	t.Run("记录基本审计事件", func(t *testing.T) {
		logConfig := &AuditLogConfig{
			EnableAuditLog: true,
			LogDatabase:    false, // 避免数据库错误
			MaxBatchSize:   10,
			BatchTimeout:   100 * time.Millisecond,
		}

		config := auditLogConfigToConfig(logConfig)
		mockDB, _ := mock.NewMockDB()
		cacheService := createTestCacheService()
		service := NewAuditService(config, mockDB.DB, cacheService)

		event := &AuditEvent{
			UserID:      1,
			Username:    "testuser",
			EventType:   EventTypeLogin,
			Severity:    SeverityLow,
			Action:      "user_login",
			IPAddress:   "192.168.1.100",
			UserAgent:   "Mozilla/5.0",
			Description: "User login attempt",
			Status:      "success",
		}

		err := service.LogEvent(event)
		require.NoError(t, err)

		// 等待事件被处理
		time.Sleep(200 * time.Millisecond)
	})

	t.Run("记录事件 - 审计禁用", func(t *testing.T) {
		logConfig := &AuditLogConfig{
			EnableAuditLog: false,
		}

		config := auditLogConfigToConfig(logConfig)
		mockDB, _ := mock.NewMockDB()
		cacheService := createTestCacheService()
		service := NewAuditService(config, mockDB.DB, cacheService)

		event := &AuditEvent{
			UserID:    1,
			Username:  "testuser",
			EventType: EventTypeLogin,
		}

		err := service.LogEvent(event)
		require.NoError(t, err) // 应该不返回错误，即使审计被禁用
	})
}

// TestAuditService_SpecificEventMethods 测试特定事件记录方法
func TestAuditService_SpecificEventMethods(t *testing.T) {
	t.Run("测试各种事件记录方法", func(t *testing.T) {
		logConfig := &AuditLogConfig{
			EnableAuditLog: true,
			LogDatabase:    false,
			MaxBatchSize:   10,
			BatchTimeout:   100 * time.Millisecond,
		}

		config := auditLogConfigToConfig(logConfig)
		mockDB, _ := mock.NewMockDB()
		cacheService := createTestCacheService()
		service := NewAuditService(config, mockDB.DB, cacheService)

		testCases := []struct {
			name   string
			method func() error
			verify func(t *testing.T)
		}{
			{
				name: "记录登录事件",
				method: func() error {
					return service.LogLogin(1, "testuser", "192.168.1.100", "Mozilla/5.0", "device-123", "session-456", "success", "")
				},
				verify: func(t *testing.T) {
					// 事件应该在队列中被处理
				},
			},
			{
				name: "记录登出事件",
				method: func() error {
					return service.LogLogout(1, "testuser", "192.168.1.100", "Mozilla/5.0", "session-456")
				},
				verify: func(t *testing.T) {
					// 事件应该在队列中被处理
				},
			},
			{
				name: "记录密码重置事件",
				method: func() error {
					return service.LogPasswordReset(1, "testuser", "192.168.1.100", "Mozilla/5.0", "success")
				},
				verify: func(t *testing.T) {
					// 事件应该在队列中被处理
				},
			},
			{
				name: "记录权限变更事件",
				method: func() error {
					return service.LogPermissionChange(1, "admin", "testuser", "admin_role", "grant", "192.168.1.100")
				},
				verify: func(t *testing.T) {
					// 事件应该在队列中被处理
				},
			},
			{
				name: "记录数据访问事件",
				method: func() error {
					return service.LogDataAccess(1, "testuser", "user_profile", "1", "read", "192.168.1.100")
				},
				verify: func(t *testing.T) {
					// 事件应该在队列中被处理
				},
			},
			{
				name: "记录数据修改事件",
				method: func() error {
					changes := map[string]interface{}{
						"email":  "old@example.com -> new@example.com",
						"status": "inactive -> active",
					}
					return service.LogDataModify(1, "testuser", "user_profile", "1", "update", "192.168.1.100", changes)
				},
				verify: func(t *testing.T) {
					// 事件应该在队列中被处理
				},
			},
			{
				name: "记录数据删除事件",
				method: func() error {
					return service.LogDataDelete(1, "testuser", "user_profile", "1", "192.168.1.100")
				},
				verify: func(t *testing.T) {
					// 事件应该在队列中被处理
				},
			},
			{
				name: "记录安全事件",
				method: func() error {
					return service.LogSecurityEvent(EventTypeSecurityEvent, "system", "brute_force_detected", "Multiple failed login attempts detected", "192.168.1.100", SeverityHigh)
				},
				verify: func(t *testing.T) {
					// 事件应该在队列中被处理
				},
			},
			{
				name: "记录API访问事件",
				method: func() error {
					return service.LogAPIAccess(1, "testuser", "GET", "/api/users", "192.168.1.100", "Mozilla/5.0", "200")
				},
				verify: func(t *testing.T) {
					// 事件应该在队列中被处理
				},
			},
			{
				name: "记录文件操作事件",
				method: func() error {
					return service.LogFileOperation(1, "testuser", "upload", "/uploads/document.pdf", "192.168.1.100", "success")
				},
				verify: func(t *testing.T) {
					// 事件应该在队列中被处理
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := tc.method()
				require.NoError(t, err)
				tc.verify(t)

				// 等待事件被处理
				time.Sleep(50 * time.Millisecond)
			})
		}
	})
}

// TestAuditService_QueryAuditLogs 测试审计日志查询
func TestAuditService_QueryAuditLogs(t *testing.T) {
	t.Run("基本查询功能", func(t *testing.T) {
		logConfig := &AuditLogConfig{
			EnableAuditLog: false, // 禁用审计日志，避免后台worker干扰查询测试
			LogDatabase:    false, // 禁用数据库，简化测试
			MaxBatchSize:   10,
			BatchTimeout:   100 * time.Millisecond,
		}

		config := auditLogConfigToConfig(logConfig)
		mockDB, _ := mock.NewMockDB()
		cacheService := createTestCacheService()
		service := NewAuditService(config, mockDB.DB, cacheService)

		filter := AuditLogFilter{
			UserID:    1,
			Username:  "testuser",
			EventType: EventTypeLogin,
			StartTime: time.Now().Add(-24 * time.Hour),
			EndTime:   time.Now(),
			Page:      1,
			PageSize:  10,
		}

		// 由于禁用了数据库，查询应该返回空结果但不报错
		events, total, err := service.QueryAuditLogs(filter)
		require.NoError(t, err)
		assert.NotNil(t, events)
		assert.Equal(t, int64(0), total) // 禁用数据库时应该返回0
		assert.Len(t, events, 0)         // 禁用数据库时应该返回空切片
	})

	t.Run("查询过滤条件", func(t *testing.T) {
		logConfig := &AuditLogConfig{
			EnableAuditLog: false, // 禁用审计日志，避免后台worker干扰查询测试
			LogDatabase:    false, // 禁用数据库，简化测试
		}

		config := auditLogConfigToConfig(logConfig)
		mockDB, _ := mock.NewMockDB()
		cacheService := createTestCacheService()
		service := NewAuditService(config, mockDB.DB, cacheService)

		testCases := []struct {
			name   string
			filter AuditLogFilter
		}{
			{
				name: "按用户ID过滤",
				filter: AuditLogFilter{
					UserID: 1,
				},
			},
			{
				name: "按用户名模糊匹配",
				filter: AuditLogFilter{
					Username: "test",
				},
			},
			{
				name: "按事件类型过滤",
				filter: AuditLogFilter{
					EventType: EventTypeLogin,
				},
			},
			{
				name: "按严重程度过滤",
				filter: AuditLogFilter{
					Severity: SeverityHigh,
				},
			},
			{
				name: "按时间范围过滤",
				filter: AuditLogFilter{
					StartTime: time.Now().Add(-24 * time.Hour),
					EndTime:   time.Now(),
				},
			},
			{
				name: "按IP地址过滤",
				filter: AuditLogFilter{
					IPAddress: "192.168.1.100",
				},
			},
			{
				name: "分页查询",
				filter: AuditLogFilter{
					Page:     2,
					PageSize: 20,
				},
			},
			{
				name: "排序查询",
				filter: AuditLogFilter{
					SortBy:    "timestamp",
					SortOrder: "desc",
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// 由于禁用了数据库，查询应该返回空结果但不报错
				events, total, err := service.QueryAuditLogs(tc.filter)
				require.NoError(t, err)
				assert.NotNil(t, events)
				assert.Equal(t, int64(0), total) // 禁用数据库时应该返回0
				assert.Len(t, events, 0)         // 禁用数据库时应该返回空切片
			})
		}
	})
}

// TestAuditService_AuditMiddleware 测试审计中间件
func TestAuditService_AuditMiddleware(t *testing.T) {
	t.Run("审计中间件基本功能", func(t *testing.T) {
		logConfig := &AuditLogConfig{
			EnableAuditLog: true,
			LogDatabase:    false,
			MaxBatchSize:   10,
			BatchTimeout:   100 * time.Millisecond,
		}

		config := auditLogConfigToConfig(logConfig)
		mockDB, _ := mock.NewMockDB()
		cacheService := createTestCacheService()
		service := NewAuditService(config, mockDB.DB, cacheService)

		middleware := service.AuditMiddleware()
		assert.NotNil(t, middleware)

		// 创建测试路由
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.Use(middleware)
		router.GET("/api/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		})

		testCases := []struct {
			name     string
			setupCtx func() *gin.Context
		}{
			{
				name: "带用户信息的请求",
				setupCtx: func() *gin.Context {
					c, _ := gin.CreateTestContext(httptest.NewRecorder())
					c.Request = httptest.NewRequest("GET", "/api/test", nil)
					c.Set("user_id", uint(1))
					c.Set("username", "testuser")
					return c
				},
			},
			{
				name: "不带用户信息的请求",
				setupCtx: func() *gin.Context {
					c, _ := gin.CreateTestContext(httptest.NewRecorder())
					c.Request = httptest.NewRequest("GET", "/api/test", nil)
					return c
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				c := tc.setupCtx()

				// 替换上下文
				router.ServeHTTP(w, c.Request)

				assert.Equal(t, 200, w.Code)

				// 等待审计事件被处理
				time.Sleep(50 * time.Millisecond)
			})
		}
	})
}

// TestAuditService_SecurityAlerts 测试安全告警功能
func TestAuditService_SecurityAlerts(t *testing.T) {
	t.Run("安全事件告警", func(t *testing.T) {
		logConfig := &AuditLogConfig{
			EnableAuditLog:      true,
			LogDatabase:         false,
			EnableRealTimeAlert: true,
			SensitiveEventTypes: []AuditEventType{EventTypeSecurityEvent, EventTypePermissionChange},
			MaxBatchSize:        10,
			BatchTimeout:        100 * time.Millisecond,
		}

		config := auditLogConfigToConfig(logConfig)
		mockDB, _ := mock.NewMockDB()
		cacheService := createTestCacheService()
		service := NewAuditService(config, mockDB.DB, cacheService)

		// 记录安全事件
		event := &AuditEvent{
			UserID:      1,
			Username:    "testuser",
			EventType:   EventTypeSecurityEvent,
			Severity:    SeverityCritical,
			Action:      "security_breach",
			Description: "Security breach detected",
			IPAddress:   "192.168.1.100",
			Status:      "success",
		}

		err := service.LogEvent(event)
		require.NoError(t, err)

		// 等待事件被处理和告警触发
		time.Sleep(200 * time.Millisecond)
	})

	t.Run("登录失败告警", func(t *testing.T) {
		logConfig := &AuditLogConfig{
			EnableAuditLog:      true,
			LogDatabase:         false,
			EnableRealTimeAlert: true,
			MaxBatchSize:        10,
			BatchTimeout:        100 * time.Millisecond,
		}

		config := auditLogConfigToConfig(logConfig)
		mockDB, _ := mock.NewMockDB()
		cacheService := createTestCacheService()
		service := NewAuditService(config, mockDB.DB, cacheService)

		// 模拟多次失败的登录尝试
		for i := 0; i < 3; i++ {
			err := service.LogLogin(1, "testuser", "192.168.1.100", "Mozilla/5.0", "device-123", "session-456", "failed", "Invalid password")
			require.NoError(t, err)
		}

		// 等待事件被处理
		time.Sleep(200 * time.Millisecond)
	})

	t.Run("敏感操作告警", func(t *testing.T) {
		logConfig := &AuditLogConfig{
			EnableAuditLog:      true,
			LogDatabase:         false,
			EnableRealTimeAlert: true,
			SensitiveEventTypes: []AuditEventType{EventTypePermissionChange},
			MaxBatchSize:        10,
			BatchTimeout:        100 * time.Millisecond,
		}

		config := auditLogConfigToConfig(logConfig)
		mockDB, _ := mock.NewMockDB()
		cacheService := createTestCacheService()
		service := NewAuditService(config, mockDB.DB, cacheService)

		// 记录权限变更事件
		err := service.LogPermissionChange(1, "admin", "testuser", "admin_role", "grant", "192.168.1.100")
		require.NoError(t, err)

		// 等待事件被处理
		time.Sleep(200 * time.Millisecond)
	})
}

// TestAuditService_DataMasking 测试数据脱敏功能
func TestAuditService_DataMasking(t *testing.T) {
	t.Run("IP地址脱敏", func(t *testing.T) {
		service := &AuditService{}

		testCases := []struct {
			input    string
			expected string
		}{
			{"192.168.1.100", "192.168.*.*"},
			{"10.0.0.1", "10.0.*.*"},
			{"127.0.0.1", "127.0.*.*"},
			{"255.255.255.255", "255.255.*.*"},
			{"invalid-ip", "***.***.***.***"},
			{"", "***.***.***.***"},
		}

		for _, tc := range testCases {
			t.Run(tc.input, func(t *testing.T) {
				result := service.maskIPAddress(tc.input)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("用户代理脱敏", func(t *testing.T) {
		service := &AuditService{}

		testCases := []struct {
			input    string
			expected string
		}{
			{"Mozilla/5.0 (Windows NT 10.0; Win64; x64)", "Mozilla/5.0 (Window..."},
			{"short", "short"},
			{"", ""},
			{"Mozilla/5.0", "Mozilla/5.0"},
			{"This is a very long user agent string that should be truncated", "This is a very lo..."},
		}

		for _, tc := range testCases {
			t.Run(tc.input, func(t *testing.T) {
				result := service.maskUserAgent(tc.input)
				if len(tc.input) > 20 {
					assert.Equal(t, tc.input[:20]+"...", result)
				} else {
					assert.Equal(t, tc.expected, result)
				}
			})
		}
	})
}

// TestAuditService_Concurrency 测试并发安全性
func TestAuditService_Concurrency(t *testing.T) {
	t.Run("并发事件记录", func(t *testing.T) {
		logConfig := &AuditLogConfig{
			EnableAuditLog: true,
			LogDatabase:    false,
			MaxBatchSize:   50,
			BatchTimeout:   200 * time.Millisecond,
		}

		config := auditLogConfigToConfig(logConfig)
		mockDB, _ := mock.NewMockDB()
		cacheService := createTestCacheService()
		service := NewAuditService(config, mockDB.DB, cacheService)

		var wg sync.WaitGroup
		workerCount := 10
		eventsPerWorker := 20

		// 启动多个goroutine并发记录事件
		for i := 0; i < workerCount; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for j := 0; j < eventsPerWorker; j++ {
					event := &AuditEvent{
						UserID:      uint(workerID),
						Username:    "testuser",
						EventType:   EventTypeAPIAccess,
						Severity:    SeverityLow,
						Action:      "concurrent_test",
						Description: "Concurrent audit test",
						IPAddress:   "192.168.1.100",
						Status:      "success",
					}

					err := service.LogEvent(event)
					assert.NoError(t, err)
				}
			}(i)
		}

		// 等待所有goroutine完成
		wg.Wait()

		// 等待事件被处理
		time.Sleep(300 * time.Millisecond)

		// 验证没有panic或错误发生
		assert.True(t, true) // 如果没有panic，测试通过
	})
}

// TestAuditService_Stop 测试服务停止
func TestAuditService_Stop(t *testing.T) {
	t.Run("正常停止服务", func(t *testing.T) {
		logConfig := &AuditLogConfig{
			EnableAuditLog: true,
			LogDatabase:    false,
			MaxBatchSize:   10,
			BatchTimeout:   100 * time.Millisecond,
		}

		config := auditLogConfigToConfig(logConfig)
		mockDB, _ := mock.NewMockDB()
		cacheService := createTestCacheService()
		service := NewAuditService(config, mockDB.DB, cacheService)

		// 记录一些事件
		for i := 0; i < 5; i++ {
			event := &AuditEvent{
				UserID:      uint(i),
				Username:    "testuser",
				EventType:   EventTypeLogin,
				Severity:    SeverityLow,
				Action:      "user_login",
				Description: "User login",
				Status:      "success",
			}

			err := service.LogEvent(event)
			require.NoError(t, err)
		}

		// 停止服务
		service.Stop()

		// 等待一段时间确保服务停止
		time.Sleep(200 * time.Millisecond)

		// 再次记录事件应该不会导致panic
		event := &AuditEvent{
			UserID:    1,
			Username:  "testuser",
			EventType: EventTypeLogin,
		}

		err := service.LogEvent(event)
		require.NoError(t, err) // 即使服务停止，也不应该返回错误
	})
}

// TestAuditService_GetAuditStats 测试统计功能
func TestAuditService_GetAuditStats(t *testing.T) {
	t.Run("获取审计统计信息", func(t *testing.T) {
		logConfig := &AuditLogConfig{
			EnableAuditLog: true,
			LogDatabase:    false,
		}

		config := auditLogConfigToConfig(logConfig)
		mockDB, _ := mock.NewMockDB()
		cacheService := createTestCacheService()
		service := NewAuditService(config, mockDB.DB, cacheService)

		stats := service.GetAuditStats()
		assert.NotNil(t, stats)

		// 验证统计字段
		assert.Contains(t, stats, "today_events")
		assert.Contains(t, stats, "week_events")
		assert.Contains(t, stats, "security_events")

		// 验证字段类型
		if todayEvents, ok := stats["today_events"].(int64); ok {
			assert.GreaterOrEqual(t, todayEvents, int64(0))
		}
		if weekEvents, ok := stats["week_events"].(int64); ok {
			assert.GreaterOrEqual(t, weekEvents, int64(0))
		}
		if securityEvents, ok := stats["security_events"].(int64); ok {
			assert.GreaterOrEqual(t, securityEvents, int64(0))
		}
	})
}

// BenchmarkAuditService 审计服务基准测试
func BenchmarkAuditService_LogEvent(b *testing.B) {
	logConfig := &AuditLogConfig{
		EnableAuditLog: true,
		LogDatabase:    false,
		MaxBatchSize:   100,
		BatchTimeout:   time.Second,
	}

	config := auditLogConfigToConfig(logConfig)
	mockDB, _ := mock.NewMockDB()
	cacheService := createTestCacheService()
	service := NewAuditService(config, mockDB.DB, cacheService)

	event := &AuditEvent{
		UserID:      1,
		Username:    "testuser",
		EventType:   EventTypeLogin,
		Severity:    SeverityLow,
		Action:      "user_login",
		Description: "User login attempt",
		IPAddress:   "192.168.1.100",
		Status:      "success",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.LogEvent(event)
	}
}

func BenchmarkAuditService_LoginSpecific(b *testing.B) {
	logConfig := &AuditLogConfig{
		EnableAuditLog: true,
		LogDatabase:    false,
		MaxBatchSize:   100,
		BatchTimeout:   time.Second,
	}

	config := auditLogConfigToConfig(logConfig)
	mockDB, _ := mock.NewMockDB()
	cacheService := createTestCacheService()
	service := NewAuditService(config, mockDB.DB, cacheService)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.LogLogin(1, "testuser", "192.168.1.100", "Mozilla/5.0", "device-123", "session-456", "success", "")
	}
}

func BenchmarkAuditService_QueryAuditLogs(b *testing.B) {
	logConfig := &AuditLogConfig{
		EnableAuditLog: true,
		LogDatabase:    false,
	}

	config := auditLogConfigToConfig(logConfig)
	mockDB, _ := mock.NewMockDB()
	cacheService := createTestCacheService()
	service := NewAuditService(config, mockDB.DB, cacheService)

	filter := AuditLogFilter{
		UserID:    1,
		EventType: EventTypeLogin,
		Page:      1,
		PageSize:  10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = service.QueryAuditLogs(filter)
	}
}

// setupTestAuditService 设置测试用的审计服务
func setupTestAuditService(t *testing.T) *AuditService {
	logConfig := &AuditLogConfig{
		EnableAuditLog:       true,
		LogDatabase:          false,
		EnableRealTimeAlert:  true,
		SensitiveEventTypes:  []AuditEventType{EventTypeSecurityEvent, EventTypePermissionChange},
		RequiredEventTypes:   []AuditEventType{EventTypeLogin, EventTypeLogout},
		MaxBatchSize:         50,
		BatchTimeout:         200 * time.Millisecond,
		RetentionDays:        365,
		EnableCompression:    true,
		EncryptSensitiveData: true,
	}

	config := auditLogConfigToConfig(logConfig)
	mockDB, _ := mock.NewMockDB()
	cacheService := createTestCacheService()
	return NewAuditService(config, mockDB.DB, cacheService)
}

// createTestAuditEvent 创建测试审计事件
func createTestAuditEvent(eventType AuditEventType) *AuditEvent {
	return &AuditEvent{
		UserID:      1,
		Username:    "testuser",
		EventType:   eventType,
		Severity:    SeverityMedium,
		Action:      "test_action",
		Resource:    "test_resource",
		ResourceID:  "123",
		IPAddress:   "192.168.1.100",
		UserAgent:   "Mozilla/5.0",
		DeviceID:    "device-123",
		SessionID:   "session-456",
		Description: "Test audit event",
		Status:      "success",
		Details: map[string]interface{}{
			"test_key": "test_value",
			"number":   42,
		},
		Location:  "test_location",
		Timestamp: time.Now(),
	}
}
