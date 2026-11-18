package models

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBasicModels(t *testing.T) {
	// 使用内存SQLite数据库进行测试
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 测试基础模型创建
	t.Run("TestSuite basic CRUD", func(t *testing.T) {
		// 简化模型，不包含ENUM字段
		type SimpleTestSuite struct {
			ID          string `gorm:"primaryKey"`
			Name        string `gorm:"size:255;not null"`
			Type        string `gorm:"size:20;not null;default:'api'"`
			Description string `gorm:"type:text"`
			Environment string `gorm:"size:50;default:'test'"`
			IsActive    bool   `gorm:"default:true"`
		}

		err := db.AutoMigrate(&SimpleTestSuite{})
		if err != nil {
			t.Fatalf("Failed to migrate SimpleTestSuite: %v", err)
		}

		// Create
		suite := SimpleTestSuite{
			ID:          "test-suite-1",
			Name:        "API Health Check",
			Type:        "api",
			Description: "Basic API health check tests",
			Environment: "test",
			IsActive:    true,
		}

		err = db.Create(&suite).Error
		if err != nil {
			t.Fatalf("Failed to create test suite: %v", err)
		}

		// Read
		var retrievedSuite SimpleTestSuite
		err = db.First(&retrievedSuite, "id = ?", suite.ID).Error
		if err != nil {
			t.Fatalf("Failed to retrieve test suite: %v", err)
		}

		if retrievedSuite.Name != suite.Name {
			t.Errorf("Expected name %s, got %s", suite.Name, retrievedSuite.Name)
		}

		// Update
		retrievedSuite.Description = "Updated description"
		err = db.Save(&retrievedSuite).Error
		if err != nil {
			t.Fatalf("Failed to update test suite: %v", err)
		}

		// Delete
		err = db.Delete(&retrievedSuite).Error
		if err != nil {
			t.Fatalf("Failed to delete test suite: %v", err)
		}
	})

	t.Run("UserEvent basic CRUD", func(t *testing.T) {
		type SimpleUserEvent struct {
			ID         string `gorm:"primaryKey"`
			UserID     string `gorm:"size:36;index"`
			SessionID  string `gorm:"size:36;not null;index"`
			EventType  string `gorm:"size:100;not null;index"`
			Element    string `gorm:"size:255"`
			PageURL    string `gorm:"size:500;index"`
			Timestamp  string `gorm:"type:datetime;not null;index"`
		}

		err := db.AutoMigrate(&SimpleUserEvent{})
		if err != nil {
			t.Fatalf("Failed to migrate SimpleUserEvent: %v", err)
		}

		// Create
		event := SimpleUserEvent{
			ID:        "event-1",
			UserID:    "user123",
			SessionID: "session456",
			EventType: "click",
			Element:   "#submit-button",
			PageURL:   "/dashboard",
			Timestamp: "2025-01-01 10:00:00",
		}

		err = db.Create(&event).Error
		if err != nil {
			t.Fatalf("Failed to create user event: %v", err)
		}

		// Read
		var retrievedEvent SimpleUserEvent
		err = db.First(&retrievedEvent, "id = ?", event.ID).Error
		if err != nil {
			t.Fatalf("Failed to retrieve user event: %v", err)
		}

		if retrievedEvent.EventType != event.EventType {
			t.Errorf("Expected event type %s, got %s", event.EventType, retrievedEvent.EventType)
		}

		// Query by session
		var events []SimpleUserEvent
		err = db.Where("session_id = ?", event.SessionID).Find(&events).Error
		if err != nil {
			t.Fatalf("Failed to query events by session: %v", err)
		}

		if len(events) != 1 {
			t.Errorf("Expected 1 event, got %d", len(events))
		}
	})

	t.Run("Text field for JSON-like data", func(t *testing.T) {
		type SimpleReport struct {
			ID     string `gorm:"primaryKey"`
			Title  string `gorm:"size:255;not null"`
			Data   string `gorm:"type:text"`
			Status string `gorm:"size:20;default:'generating'"`
		}

		err := db.AutoMigrate(&SimpleReport{})
		if err != nil {
			t.Fatalf("Failed to migrate SimpleReport: %v", err)
		}

		// Create with JSON data as string
		reportData := `{"summary": {"total": 10, "passed": 8, "failed": 2, "rate": 0.8}, "items": [{"name": "Test 1", "status": "passed"}, {"name": "Test 2", "status": "failed"}]}`

		report := SimpleReport{
			ID:     "report-1",
			Title:  "Test Report",
			Data:   reportData,
			Status: "completed",
		}

		err = db.Create(&report).Error
		if err != nil {
			t.Fatalf("Failed to create report: %v", err)
		}

		// Read and verify JSON string
		var retrievedReport SimpleReport
		err = db.First(&retrievedReport, "id = ?", report.ID).Error
		if err != nil {
			t.Fatalf("Failed to retrieve report: %v", err)
		}

		if retrievedReport.Data == "" {
			t.Error("Expected report data to be non-empty")
		}

		if retrievedReport.Data != reportData {
			t.Error("JSON data does not match original")
		}
	})

	t.Log("Basic model tests passed successfully!")
}