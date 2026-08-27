package models

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTestingModels(t *testing.T) {
	// 使用内存SQLite数据库进行测试
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 自动迁移所有模型
	err = db.AutoMigrate(
		&TestSuite{},
		&TestExecution{},
		&TestResult{},
		&UserEvent{},
		&UserSession{},
		&AnalyticsReport{},
		&SystemMetric{},
		&Alert{},
		&AlertNotification{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	t.Run("TestSuite CRUD", func(t *testing.T) {
		// Create test suite
		suite := &TestSuite{
			Name:        "API Health Check",
			Type:        TestTypeAPI,
			Description: "Basic API health check tests",
			Environment: "test",
			IsActive:    true,
		}

		err := db.Create(suite).Error
		if err != nil {
			t.Fatalf("Failed to create test suite: %v", err)
		}

		// Read test suite
		var retrievedSuite TestSuite
		err = db.First(&retrievedSuite, "id = ?", suite.ID).Error
		if err != nil {
			t.Fatalf("Failed to retrieve test suite: %v", err)
		}

		if retrievedSuite.Name != suite.Name {
			t.Errorf("Expected name %s, got %s", suite.Name, retrievedSuite.Name)
		}

		// Update test suite
		retrievedSuite.Description = "Updated description"
		err = db.Save(&retrievedSuite).Error
		if err != nil {
			t.Fatalf("Failed to update test suite: %v", err)
		}

		// Delete test suite
		err = db.Delete(&retrievedSuite).Error
		if err != nil {
			t.Fatalf("Failed to delete test suite: %v", err)
		}
	})

	t.Run("TestExecution with relationships", func(t *testing.T) {
		// Create test suite first
		suite := &TestSuite{
			Name:        "Integration Test Suite",
			Type:        TestTypeIntegration,
			Description: "Integration test suite",
			Environment: "test",
			IsActive:    true,
		}
		err := db.Create(suite).Error
		if err != nil {
			t.Fatalf("Failed to create test suite: %v", err)
		}

		// Create test execution
		execution := &TestExecution{
			SuiteID:     suite.ID,
			Status:      TestStatusRunning,
			Environment: "test",
			TriggerType: TriggerManual,
			StartedAt:   &[]time.Time{time.Now()}[0],
		}
		err = db.Create(execution).Error
		if err != nil {
			t.Fatalf("Failed to create test execution: %v", err)
		}

		// Create test results
		results := []*TestResult{
			{
				ExecutionID: execution.ID,
				TestName:    "Test API Health",
				TestType:    "api",
				Status:      TestResultPassed,
				DurationMs:  150,
			},
			{
				ExecutionID: execution.ID,
				TestName:    "Test Database Connection",
				TestType:    "database",
				Status:      TestResultPassed,
				DurationMs:  200,
			},
		}

		err = db.CreateInBatches(results, 100).Error
		if err != nil {
			t.Fatalf("Failed to create test results: %v", err)
		}

		// Test relationships
		var executionWithResults TestExecution
		err = db.Preload("TestResults").First(&executionWithResults, "id = ?", execution.ID).Error
		if err != nil {
			t.Fatalf("Failed to retrieve execution with results: %v", err)
		}

		if len(executionWithResults.TestResults) != 2 {
			t.Errorf("Expected 2 test results, got %d", len(executionWithResults.TestResults))
		}
	})

	t.Run("UserEvent and UserSession", func(t *testing.T) {
		// Create user session
		session := &UserSession{
			UserID:     "user123",
			StartedAt:  time.Now(),
			Browser:    "Chrome",
			DeviceType: "desktop",
			IPAddress:  "192.168.1.1",
		}
		err := db.Create(session).Error
		if err != nil {
			t.Fatalf("Failed to create user session: %v", err)
		}

		// Create user events
		events := []*UserEvent{
			{
				UserID:    "user123",
				SessionID: session.ID,
				EventType: "page_view",
				PageURL:   "/dashboard",
				Timestamp: time.Now(),
			},
			{
				UserID:    "user123",
				SessionID: session.ID,
				EventType: "click",
				Element:   "#submit-button",
				PageURL:   "/dashboard",
				Timestamp: time.Now().Add(1 * time.Second),
			},
		}

		err = db.CreateInBatches(events, 200).Error
		if err != nil {
			t.Fatalf("Failed to create user events: %v", err)
		}

		// Verify events are associated with session
		var eventCount int64
		err = db.Model(&UserEvent{}).Where("session_id = ?", session.ID).Count(&eventCount).Error
		if err != nil {
			t.Fatalf("Failed to count events: %v", err)
		}

		if eventCount != 2 {
			t.Errorf("Expected 2 events, got %d", eventCount)
		}
	})

	t.Run("SystemMetric with labels", func(t *testing.T) {
		metric := &SystemMetric{
			MetricName: "http_requests_total",
			MetricType: MetricTypeCounter,
			Value:      100.0,
			Labels: map[string]string{
				"method": "GET",
				"status": "200",
			},
			Timestamp: time.Now(),
		}

		err := db.Create(metric).Error
		if err != nil {
			t.Fatalf("Failed to create system metric: %v", err)
		}

		// Query with label filter
		var retrievedMetric SystemMetric
		err = db.First(&retrievedMetric, "metric_name = ?", "http_requests_total").Error
		if err != nil {
			t.Fatalf("Failed to retrieve system metric: %v", err)
		}

		if retrievedMetric.Value != 100.0 {
			t.Errorf("Expected value 100.0, got %f", retrievedMetric.Value)
		}

		if retrievedMetric.Labels["method"] != "GET" {
			t.Errorf("Expected method GET, got %s", retrievedMetric.Labels["method"])
		}
	})

	t.Run("Alert with notifications", func(t *testing.T) {
		alert := &Alert{
			Title:       "High Error Rate",
			Description: "Error rate exceeded threshold",
			Severity:    AlertSeverityWarning,
			Source:      "monitoring",
			Status:      AlertStatusActive,
			TriggeredAt: time.Now(),
		}

		err := db.Create(alert).Error
		if err != nil {
			t.Fatalf("Failed to create alert: %v", err)
		}

		// Create notifications
		notifications := []*AlertNotification{
			{
				AlertID:     alert.ID,
				ChannelType: NotificationTypeEmail,
				Recipient:   "admin@example.com",
				Status:      NotificationStatusPending,
			},
			{
				AlertID:     alert.ID,
				ChannelType: NotificationTypeWebhook,
				Recipient:   "https://hooks.slack.com/webhook",
				Status:      NotificationStatusPending,
			},
		}

		err = db.CreateInBatches(notifications, 100).Error
		if err != nil {
			t.Fatalf("Failed to create alert notifications: %v", err)
		}

		// Test relationship
		var alertWithNotifications Alert
		err = db.Preload("Notifications").First(&alertWithNotifications, "id = ?", alert.ID).Error
		if err != nil {
			t.Fatalf("Failed to retrieve alert with notifications: %v", err)
		}

		if len(alertWithNotifications.Notifications) != 2 {
			t.Errorf("Expected 2 notifications, got %d", len(alertWithNotifications.Notifications))
		}
	})

	t.Run("AnalyticsReport with JSON data", func(t *testing.T) {
		reportData := map[string]interface{}{
			"summary": map[string]interface{}{
				"total_tests":  10,
				"passed_tests": 8,
				"failed_tests": 2,
				"success_rate": 0.8,
				"avg_duration": 150.5,
			},
			"details": []map[string]interface{}{
				{
					"test_name": "API Health Check",
					"status":    "passed",
					"duration":  120,
				},
				{
					"test_name": "Database Test",
					"status":    "failed",
					"duration":  200,
				},
			},
		}

		report := &AnalyticsReport{
			ReportType:  ReportTypeTestSummary,
			Title:       "Daily Test Report",
			Description: "Automated daily test execution report",
			Data:        reportData,
			Status:      ReportStatusCompleted,
			GeneratedBy: "system",
			GeneratedAt: &[]time.Time{time.Now()}[0],
		}

		err := db.Create(report).Error
		if err != nil {
			t.Fatalf("Failed to create analytics report: %v", err)
		}

		// Verify JSON data is stored correctly
		var retrievedReport AnalyticsReport
		err = db.First(&retrievedReport, "id = ?", report.ID).Error
		if err != nil {
			t.Fatalf("Failed to retrieve analytics report: %v", err)
		}

		if retrievedReport.Title != report.Title {
			t.Errorf("Expected title %s, got %s", report.Title, retrievedReport.Title)
		}

		// The Data field should be properly serialized/deserialized
		if retrievedReport.Data == nil {
			t.Error("Expected report data to be non-nil")
		}
	})

	t.Run("Test Config JSON serialization", func(t *testing.T) {
		config := &TestConfig{
			Timeout: 30,
			Variables: map[string]interface{}{
				"BASE_URL": "http://localhost:8080",
				"API_KEY":  "test-key-12345",
			},
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer token123",
			},
			Parallel: true,
			Retries:  3,
			Setup: []TestStep{
				{
					Name:   "Setup database",
					Type:   "sql",
					Action: "CREATE TABLE test_table (id INT)",
				},
			},
		}

		suite := &TestSuite{
			Name:        "Suite with Config",
			Type:        TestTypeAPI,
			Description: "Test suite with complex configuration",
			Config:      config,
			Environment: "test",
			IsActive:    true,
		}

		err := db.Create(suite).Error
		if err != nil {
			t.Fatalf("Failed to create test suite with config: %v", err)
		}

		// Retrieve and verify config
		var retrievedSuite TestSuite
		err = db.First(&retrievedSuite, "id = ?", suite.ID).Error
		if err != nil {
			t.Fatalf("Failed to retrieve test suite with config: %v", err)
		}

		if retrievedSuite.Config.Timeout != 30 {
			t.Errorf("Expected timeout 30, got %d", retrievedSuite.Config.Timeout)
		}

		if retrievedSuite.Config.Variables["BASE_URL"] != "http://localhost:8080" {
			t.Errorf("Expected BASE_URL http://localhost:8080, got %s", retrievedSuite.Config.Variables["BASE_URL"])
		}

		if len(retrievedSuite.Config.Setup) != 1 {
			t.Errorf("Expected 1 setup step, got %d", len(retrievedSuite.Config.Setup))
		}
	})

	t.Log("All model tests passed successfully!")
}
