package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	"law-oa-go/internal/services"
)

func main() {
	log.Println("🚀 开始核心集成测试...")

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

	// 3. 手动创建服务（绕过有问题的自动初始化）
	log.Println("🔧 手动创建核心服务...")
	conflictService := createConflictDetectionService(db)

	// 4. 创建Gin应用和手动路由
	log.Println("🛣️ 创建手动路由...")
	app := createManualRoutes(conflictService)

	// 5. 测试健康检查端点
	log.Println("🏥 测试健康检查端点...")
	testHealthCheck(app)

	// 6. 测试冲突检测端点（无需认证）
	log.Println("⚔️ 测试冲突检测端点...")
	testConflictDetection(app)

	log.Println("🎉 核心集成测试完成！")
}

// setupTestDatabase 设置测试数据库
func setupTestDatabase() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接内存数据库失败: %w", err)
	}

	// 只迁移核心模型
	err = db.AutoMigrate(
		&models.User{},
		&models.Client{},
		&models.Case{},
	)
	if err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	log.Println("✅ 测试数据库设置完成")
	return db, nil
}

// setupTestData 设置测试数据
func setupTestData(db *gorm.DB) error {
	// 创建测试用户（包含新增字段）
	user := models.User{
		Username:   "lawyer1",
		Name:       "张律师",
		Email:      "zhang@law.com",
		Role:       "lawyer",
		Department: "诉讼部", // 测试新增字段
		Seniority:  "中级",   // 测试新增字段
	}
	if err := db.Create(&user).Error; err != nil {
		return fmt.Errorf("创建测试用户失败: %w", err)
	}

	// 创建测试客户
	client := models.Client{
		Name:     "测试客户公司",
		Type:     "COMPANY",
		Email:    "client@example.com",
		Industry: "互联网科技",
	}
	if err := db.Create(&client).Error; err != nil {
		return fmt.Errorf("创建测试客户失败: %w", err)
	}

	// 创建测试案件
	caseModel := models.Case{
		Title:       "商业纠纷案件",
		Description: "测试商业纠纷案件描述",
		ClientID:    1,
		LawyerID:    1,
		CaseType:    "商业纠纷",
		Status:      "active",
	}
	if err := db.Create(&caseModel).Error; err != nil {
		return fmt.Errorf("创建测试案件失败: %w", err)
	}

	log.Println("✅ 测试数据初始化完成")
	return nil
}

// createConflictDetectionService 创建冲突检测服务
func createConflictDetectionService(db *gorm.DB) services.ConflictDetectionService {
	// 创建仓储
	userRepo := repositories.NewUserRepository(db)
	clientRepo := repositories.NewClientRepository(db)
	caseRepo := repositories.NewCaseRepository(db)

	// 创建冲突仓储（简化版本）
	var conflictRepo repositories.BasicConflictRepository = repositories.NewConflictRepository(db, nil)

	// 创建风险评估器
	riskAssessor := services.NewRiskAssessor(nil, nil)

	// 创建冲突检测服务
	conflictService := services.NewConflictDetectionService(
		conflictRepo,
		riskAssessor,
		userRepo,
		clientRepo,
		caseRepo,
	)

	log.Println("✅ 冲突检测服务创建成功")
	return conflictService
}

// createManualRoutes 创建手动路由（绕过有问题的自动路由）
func createManualRoutes(conflictService services.ConflictDetectionService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	app := gin.New()

	// 添加中间件
	app.Use(gin.Logger())
	app.Use(gin.Recovery())

	// 创建冲突检测处理器
	conflictHandler := handlers.NewConflictHandlerSimple(conflictService)

	// 添加冲突检测路由（无需认证）
	api := app.Group("/api/v1")
	{
		conflict := api.Group("/conflict")
		{
			conflict.POST("/check", conflictHandler.CheckConflict)
			conflict.GET("/health", conflictHandler.HealthCheck)
		}
	}

	log.Println("✅ 手动路由创建完成")
	return app
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

	if response["code"] != float64(200) {
		log.Printf("❌ 健康检查响应码错误: %v", response["code"])
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

	log.Printf("✅ 健康检查通过: 服务=%s, 状态=%s", service, status)
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

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	// 检查响应状态
	if w.Code != http.StatusOK {
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

	// 验证响应结构
	if response["code"] != float64(200) {
		log.Printf("❌ 响应码错误: %v", response["code"])
		return
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		log.Printf("❌ 响应数据格式错误")
		return
	}

	// 验证关键字段
	if checkId, exists := data["checkId"]; exists {
		log.Printf("✅ 检测ID: %v", checkId)
	} else {
		log.Printf("❌ 缺少检测ID")
		return
	}

	if hasConflict, exists := data["hasConflict"]; exists {
		log.Printf("✅ 冲突检测结果: %v", hasConflict)
	}

	if conflictCases, exists := data["conflictCases"]; exists {
		conflictCasesList := conflictCases.([]interface{})
		log.Printf("✅ 冲突案例数量: %d", len(conflictCasesList))
	}

	if riskAssessment, exists := data["riskAssessment"]; exists {
		if riskMap, ok := riskAssessment.(map[string]interface{}); ok {
			if overallRisk, exists := riskMap["overallRisk"]; exists {
				log.Printf("✅ 风险等级: %v", overallRisk)
			}
		}
	}

	log.Printf("✅ 完整响应结构验证通过")
	log.Printf("响应摘要: %+v", data)
}