package main

import (
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
	fmt.Println("🚀 快速冲突检测服务测试")
	fmt.Println("=" + "="*49)

	// 1. 初始化数据库
	fmt.Println("📋 初始化数据库...")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// 2. 迁移模型
	if err := db.AutoMigrate(&models.User{}, &models.Client{}, &models.Case{}); err != nil {
		log.Fatalf("❌ 模型迁移失败: %v", err)
	}

	// 3. 创建测试数据
	fmt.Println("🌱 创建测试数据...")
	if err := createQuickTestData(db); err != nil {
		log.Fatalf("❌ 测试数据创建失败: %v", err)
	}

	// 4. 创建服务
	fmt.Println("🔧 创建服务...")
	conflictService := createConflictService(db)

	// 5. 创建HTTP服务器
	fmt.Println("🌐 创建HTTP服务器...")
	app := gin.New()
	app.Use(gin.Recovery())

	// 创建处理器
	conflictHandler := handlers.NewConflictHandlerSimple(conflictService)

	// 路由
	api := app.Group("/api/v1")
	{
		api.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "独立冲突检测服务运行中"})
		})
		api.POST("/conflict/check", conflictHandler.CheckConflict)
		api.GET("/conflict/health", conflictHandler.HealthCheck)
	}

	// 6. 运行测试
	fmt.Println("🧪 开始API测试...")
	runAPITests(app)

	fmt.Println("\n" + "="*49)
	fmt.Println("🎉 快速测试完成！")
}

// createQuickTestData 创建快速测试数据
func createQuickTestData(db *gorm.DB) error {
	// 创建用户（测试新增字段）
	user := models.User{
		Username:   "test_lawyer",
		Name:       "测试律师",
		Email:      "test@law.com",
		Role:       "lawyer",
		Department: "诉讼部", // 测试新增字段
		Seniority:  "中级",   // 测试新增字段
		Status:     "active",
	}
	if err := db.Create(&user).Error; err != nil {
		return err
	}

	// 创建客户
	client := models.Client{
		Name:     "测试客户",
		Type:     "COMPANY",
		Email:    "client@test.com",
		Industry: "互联网科技",
		Status:   "active",
	}
	if err := db.Create(&client).Error; err != nil {
		return err
	}

	// 创建案件（用于冲突检测）
	caseModel := models.Case{
		Title:       "现有案件",
		Description: "用于冲突检测测试的现有案件",
		ClientID:    1,
		LawyerID:    1,
		CaseType:    "商业纠纷",
		Status:      "active",
	}
	if err := db.Create(&caseModel).Error; err != nil {
		return err
	}

	fmt.Println("✅ 测试数据创建完成")
	return nil
}

// createConflictService 创建冲突检测服务
func createConflictService(db *gorm.DB) services.ConflictDetectionService {
	userRepo := repositories.NewUserRepository(db)
	clientRepo := repositories.NewClientRepository(db)
	caseRepo := repositories.NewCaseRepository(db)

	var conflictRepo repositories.BasicConflictRepository = repositories.NewConflictRepository(db, nil)
	riskAssessor := services.NewRiskAssessor(nil, nil)

	return services.NewConflictDetectionService(
		conflictRepo,
		riskAssessor,
		userRepo,
		clientRepo,
		caseRepo,
	)
}

// runAPITests 运行API测试
func runAPITests(app *gin.Engine) {
	// 测试1: 根路径
	fmt.Println("\n1️⃣ 测试根路径...")
	testRootPath(app)

	// 测试2: 健康检查
	fmt.Println("\n2️⃣ 测试健康检查...")
	testHealthCheck(app)

	// 测试3: 冲突检测
	fmt.Println("\n3️⃣ 测试冲突检测...")
	testConflictDetection(app)

	// 测试4: User模型字段验证
	fmt.Println("\n4️⃣ 验证User模型字段...")
	testUserModelFields(db)
}

// testRootPath 测试根路径
func testRootPath(app *gin.Engine) {
	req := httptest.NewRequest("GET", "/api/v1/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		fmt.Printf("❌ 根路径测试失败，状态码: %d\n", w.Code)
		return
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		fmt.Printf("❌ 根路径响应解析失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 根路径测试通过: %v\n", response["message"])
}

// testHealthCheck 测试健康检查
func testHealthCheck(app *gin.Engine) {
	req := httptest.NewRequest("GET", "/api/v1/conflict/health", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		fmt.Printf("❌ 健康检查失败，状态码: %d\n", w.Code)
		return
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		fmt.Printf("❌ 健康检查响应解析失败: %v\n", err)
		return
	}

	data := response["data"].(map[string]interface{})
	service := data["service"].(string)
	status := data["status"].(string)

	fmt.Printf("✅ 健康检查通过: 服务=%s, 状态=%s\n", service, status)
}

// testConflictDetection 测试冲突检测
func testConflictDetection(app *gin.Engine) {
	request := map[string]interface{}{
		"clientId":                  "1",
		"clientName":                "测试客户公司",
		"clientType":                "COMPANY",
		"otherParties":              []string{"竞争对手公司"},
		"caseName":                  "新案件",
		"caseType":                  "商业纠纷",
		"searchYears":               5,
		"includeCorporateRelations": true,
		"searchDepth":               "STANDARD",
		"userId":                    1,
		"requestTime":               time.Now(),
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/v1/conflict/check",
		strings.NewReader(string(requestBody)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		fmt.Printf("❌ 冲突检测失败，状态码: %d\n", w.Code)
		fmt.Printf("响应内容: %s\n", w.Body.String())
		return
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		fmt.Printf("❌ 冲突检测响应解析失败: %v\n", err)
		return
	}

	data := response["data"].(map[string]interface{})
	checkId := data["checkId"].(string)
	hasConflict := data["hasConflict"].(bool)

	fmt.Printf("✅ 冲突检测通过: ID=%s, 冲突=%t\n", checkId, hasConflict)

	// 显示详细信息
	if riskAssessment, ok := data["riskAssessment"].(map[string]interface{}); ok {
		if overallRisk, ok := riskAssessment["overallRisk"]; ok {
			fmt.Printf("   风险等级: %v\n", overallRisk)
		}
	}
}

// testUserModelFields 测试User模型字段
func testUserModelFields(db *gorm.DB) {
	var user models.User
	if err := db.Where("email = ?", "test@law.com").First(&user).Error; err != nil {
		fmt.Printf("❌ 查询用户失败: %v\n", err)
		return
	}

	fmt.Printf("✅ User模型字段验证:\n")
	fmt.Printf("   ID: %d\n", user.ID)
	fmt.Printf("   姓名: %s\n", user.Name)
	fmt.Printf("   邮箱: %s\n", user.Email)
	fmt.Printf("   角色: %s\n", user.Role)
	fmt.Printf("   部门: %s\n", user.Department) // 新增字段
	fmt.Printf("   职级: %s\n", user.Seniority)   // 新增字段
	fmt.Printf("   状态: %s\n", user.Status)

	// 验证字段是否正确设置
	if user.Department == "" {
		fmt.Printf("⚠️ Department字段为空\n")
	}
	if user.Seniority == "" {
		fmt.Printf("⚠️ Seniority字段为空\n")
	}
}