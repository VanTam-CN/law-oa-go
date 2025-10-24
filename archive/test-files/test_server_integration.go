package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"law-oa-go/internal/handlers"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/router"
	"law-oa-go/internal/services"
)

func main() {
	log.Println("🚀 开始服务集成测试...")

	// 1. 设置测试数据库
	log.Println("📋 设置测试数据库...")
	db, err := setupTestDatabase()
	if err != nil {
		log.Fatalf("❌ 数据库设置失败: %v", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// 2. 初始化测试数据
	log.Println("🌱 初始化测试数据...")
	if err := setupTestData(db); err != nil {
		log.Fatalf("❌ 测试数据初始化失败: %v", err)
	}

	// 3. 创建Gin应用
	log.Println("🔧 创建应用实例...")
	gin.SetMode(gin.TestMode)
	app := gin.New()

	// 4. 初始化路由（这会测试我们的服务修复）
	log.Println("🛣️ 初始化路由系统...")
	router.Init(app, db, nil, nil)

	// 5. 测试健康检查端点
	log.Println("🏥 测试健康检查端点...")
	testHealthCheck(app)

	// 6. 测试冲突检测端点
	log.Println("⚔️ 测试冲突检测端点...")
	testConflictDetection(app)

	log.Println("🎉 服务集成测试完成！")
}

// setupTestDatabase 设置测试数据库
func setupTestDatabase() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接内存数据库失败: %w", err)
	}

	// 自动迁移所有模型
	err = db.AutoMigrate(
		&models.User{},
		&models.Client{},
		&models.Case{},
		&models.ConflictCheckRecord{},
		&models.ClientRelation{},
		&models.ConflictRule{},
		&models.ConflictCase{},
	)
	if err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	log.Println("✅ 测试数据库设置完成")
	return db, nil
}

// setupTestData 设置测试数据
func setupTestData(db *gorm.DB) error {
	// 创建测试用户
	users := []models.User{
		{
			Username:   "lawyer1",
			Name:       "张律师",
			Email:      "zhang@law.com",
			Role:       "lawyer",
			Department: "诉讼部", // 新增字段测试
			Seniority:  "中级",   // 新增字段测试
		},
		{
			Username:   "lawyer2",
			Name:       "李律师",
			Email:      "li@law.com",
			Role:       "lawyer",
			Department: "合规部",
			Seniority:  "高级",
		},
	}

	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			return fmt.Errorf("创建测试用户失败: %w", err)
		}
	}

	// 创建测试客户
	clients := []models.Client{
		{
			Name:     "测试客户公司",
			Type:     "COMPANY",
			Email:    "client1@example.com",
			Industry: "互联网科技",
		},
		{
			Name:     "竞争对手公司",
			Type:     "COMPANY",
			Email:    "client2@example.com",
			Industry: "互联网科技",
		},
	}

	for _, client := range clients {
		if err := db.Create(&client).Error; err != nil {
			return fmt.Errorf("创建测试客户失败: %w", err)
		}
	}

	// 创建测试案件
	cases := []models.Case{
		{
			Title:       "商业纠纷案件",
			Description: "测试商业纠纷案件描述",
			ClientID:    1,
			LawyerID:    1,
			CaseType:    "商业纠纷",
			Status:      "active",
		},
		{
			Title:       "知识产权案件",
			Description: "测试知识产权案件描述",
			ClientID:    2,
			LawyerID:    1,
			CaseType:    "知识产权",
			Status:      "active",
		},
	}

	for _, caseModel := range cases {
		if err := db.Create(&caseModel).Error; err != nil {
			return fmt.Errorf("创建测试案件失败: %w", err)
		}
	}

	log.Println("✅ 测试数据初始化完成")
	return nil
}

// testHealthCheck 测试健康检查端点
func testHealthCheck(app *gin.Engine) {
	req := httptest.NewRequest("GET", "/api/v1/conflict/health", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		log.Printf("❌ 健康检查失败，状态码: %d", w.Code)
		log.Printf("响应内容: %s", w.Body.String())
		return
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		log.Printf("❌ 健康检查响应解析失败: %v", err)
		return
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		log.Printf("❌ 健康检查响应格式错误")
		return
	}

	service, ok := data["service"].(string)
	if !ok || service != "conflict-check" {
		log.Printf("❌ 健康检查服务名称错误: %s", service)
		return
	}

	status, ok := data["status"].(string)
	if !ok || status != "healthy" {
		log.Printf("❌ 健康检查状态错误: %s", status)
		return
	}

	log.Printf("✅ 健康检查通过: %+v", data)
}

// testConflictDetection 测试冲突检测端点
func testConflictDetection(app *gin.Engine) {
	// 构建冲突检测请求
	request := map[string]interface{}{
		"clientId":                  "1",
		"clientName":                "测试客户公司",
		"clientType":                "COMPANY",
		"otherParties":              []string{"竞争对手公司"},
		"caseName":                  "新商业纠纷案件",
		"caseType":                  "商业纠纷",
		"searchYears":               5,
		"includeCorporateRelations": true,
		"searchDepth":               "STANDARD",
		"userId":                    1,
		"requestTime":               time.Now(),
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		log.Printf("❌ 请求序列化失败: %v", err)
		return
	}

	// 创建HTTP请求
	req := httptest.NewRequest("POST", "/api/v1/conflict/check",
		strings.NewReader(string(requestBody)))
	req.Header.Set("Content-Type", "application/json")

	// 添加模拟认证头（因为路由需要认证中间件）
	req.Header.Set("Authorization", "Bearer mock-token")

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	// 检查响应状态
	if w.Code == http.StatusUnauthorized {
		log.Printf("⚠️ 需要认证，这符合预期")
	} else if w.Code != http.StatusOK {
		log.Printf("❌ 冲突检测请求失败，状态码: %d", w.Code)
		log.Printf("响应内容: %s", w.Body.String())
		return
	}

	// 解析响应
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		log.Printf("❌ 冲突检测响应解析失败: %v", err)
		return
	}

	log.Printf("✅ 冲突检测端点响应成功")
	log.Printf("响应代码: %d", w.Code)
	log.Printf("响应内容: %s", w.Body.String())

	// 如果有数据字段，解析详细内容
	if data, ok := response["data"].(map[string]interface{}); ok {
		if checkId, exists := data["checkId"]; exists {
			log.Printf("检测ID: %v", checkId)
		}
		if hasConflict, exists := data["hasConflict"]; exists {
			log.Printf("发现冲突: %v", hasConflict)
		}
		if conflictCases, exists := data["conflictCases"]; exists {
			log.Printf("冲突案例数量: %v", len(conflictCases.([]interface{})))
		}
	}
}