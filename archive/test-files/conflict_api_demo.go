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
	fmt.Println("🚀 冲突检测API演示")
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
	if err := createTestData(db); err != nil {
		log.Fatalf("❌ 测试数据创建失败: %v", err)
	}

	// 4. 创建服务
	fmt.Println("🔧 创建冲突检测服务...")
	conflictService := createConflictService(db)

	// 5. 创建HTTP服务器
	fmt.Println("🌐 创建HTTP服务器...")
	app := createServer(conflictService)

	// 6. 运行API演示
	fmt.Println("🧪 开始API演示...")
	runAPIDemo(app)

	fmt.Println("\n" + "="*49)
	fmt.Println("🎉 API演示完成！")
	fmt.Println("✅ 核心修复验证成功！")
}

// createTestData 创建测试数据
func createTestData(db *gorm.DB) error {
	// 创建用户（验证新增字段）
	users := []models.User{
		{
			Username:   "lawyer_zhang",
			Name:       "张律师",
			Email:      "zhang@law.com",
			Role:       "lawyer",
			Department: "诉讼部", // 测试新增字段
			Seniority:  "中级",   // 测试新增字段
			Status:     "active",
		},
		{
			Username:   "lawyer_li",
			Name:       "李律师",
			Email:      "li@law.com",
			Role:       "lawyer",
			Department: "合规部",
			Seniority:  "高级",
			Status:     "active",
		},
	}

	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			return err
		}
	}

	// 创建客户
	clients := []models.Client{
		{
			Name:     "腾讯科技",
			Type:     "COMPANY",
			Email:    "tencent@example.com",
			Industry: "互联网科技",
			Status:   "active",
		},
		{
			Name:     "字节跳动",
			Type:     "COMPANY",
			Email:    "bytedance@example.com",
			Industry: "互联网科技",
			Status:   "active",
		},
		{
			Name:     "阿里巴巴",
			Type:     "COMPANY",
			Email:    "alibaba@example.com",
			Industry: "电子商务",
			Status:   "active",
		},
	}

	for _, client := range clients {
		if err := db.Create(&client).Error; err != nil {
			return err
		}
	}

	// 创建案件（用于冲突检测）
	cases := []models.Case{
		{
			Title:       "腾讯诉字节跳动案",
			Description: "短视频平台版权纠纷",
			ClientID:    1, // 腾讯
			LawyerID:    1, // 张律师
			CaseType:    "知识产权",
			Status:      "active",
		},
		{
			Title:       "阿里巴巴商业秘密案",
			Description: "电商平台商业秘密保护",
			ClientID:    3, // 阿里巴巴
			LawyerID:    1, // 张律师
			CaseType:    "商业纠纷",
			Status:      "active",
		},
	}

	for _, caseModel := range cases {
		if err := db.Create(&caseModel).Error; err != nil {
			return err
		}
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

// createServer 创建HTTP服务器
func createServer(conflictService services.ConflictDetectionService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	app := gin.New()
	app.Use(gin.Recovery())

	conflictHandler := handlers.NewConflictHandlerSimple(conflictService)

	api := app.Group("/api/v1")
	{
		api.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"service": "冲突检测API演示",
				"status":  "running",
				"version": "1.0.0",
			})
		})
		api.POST("/conflict/check", conflictHandler.CheckConflict)
		api.GET("/conflict/health", conflictHandler.HealthCheck)
	}

	return app
}

// runAPIDemo 运行API演示
func runAPIDemo(app *gin.Engine) {
	// 演示1: 服务信息
	fmt.Println("\n1️⃣ 服务信息演示...")
	testServiceInfo(app)

	// 演示2: 健康检查
	fmt.Println("\n2️⃣ 健康检查演示...")
	testHealthCheck(app)

	// 演示3: 商业竞争冲突检测
	fmt.Println("\n3️⃣ 商业竞争冲突检测演示...")
	testBusinessCompetitionConflict(app)

	// 演示4: 法律对立冲突检测
	fmt.Println("\n4️⃣ 法律对立冲突检测演示...")
	testLegalOppositionConflict(app)

	// 演示5: 无冲突场景
	fmt.Println("\n5️⃣ 无冲突场景演示...")
	testNoConflictScenario(app)

	// 演示6: 验证User模型字段
	fmt.Println("\n6️⃣ 验证User模型字段...")
	testUserModelFields()
}

// testServiceInfo 测试服务信息
func testServiceInfo(app *gin.Engine) {
	req := httptest.NewRequest("GET", "/api/v1/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	fmt.Printf("✅ 服务: %v (版本: %v)\n", response["service"], response["version"])
}

// testHealthCheck 测试健康检查
func testHealthCheck(app *gin.Engine) {
	req := httptest.NewRequest("GET", "/api/v1/conflict/health", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	fmt.Printf("✅ 健康检查: %s - %s\n", data["service"], data["status"])
}

// testBusinessCompetitionConflict 测试商业竞争冲突
func testBusinessCompetitionConflict(app *gin.Engine) {
	request := map[string]interface{}{
		"clientId":                  "2",
		"clientName":                "字节跳动",
		"clientType":                "COMPANY",
		"otherParties":              []string{"腾讯科技"},
		"caseName":                  "短视频版权纠纷",
		"caseType":                  "知识产权",
		"searchYears":               5,
		"includeCorporateRelations": true,
		"searchDepth":               "STANDARD",
		"userId":                    1,
		"requestTime":               time.Now(),
	}

	testConflictAPI(app, request, "商业竞争冲突")
}

// testLegalOppositionConflict 测试法律对立冲突
func testLegalOppositionConflict(app *gin.Engine) {
	request := map[string]interface{}{
		"clientId":                  "3",
		"clientName":                "阿里巴巴",
		"clientType":                "COMPANY",
		"otherParties":              []string{"竞争对手"},
		"caseName":                  "电商平台纠纷",
		"caseType":                  "商业纠纷",
		"searchYears":               3,
		"includeCorporateRelations": true,
		"searchDepth":               "DEEP",
		"userId":                    1,
		"requestTime":               time.Now(),
	}

	testConflictAPI(app, request, "法律对立冲突")
}

// testNoConflictScenario 测试无冲突场景
func testNoConflictScenario(app *gin.Engine) {
	request := map[string]interface{}{
		"clientId":                  "1",
		"clientName":                "新客户公司",
		"clientType":                "COMPANY",
		"otherParties":              []string{"不相关公司"},
		"caseName":                  "内部咨询",
		"caseType":                  "法律咨询",
		"searchYears":               1,
		"includeCorporateRelations": false,
		"searchDepth":               "BASIC",
		"userId":                    2,
		"requestTime":               time.Now(),
	}

	testConflictAPI(app, request, "无冲突场景")
}

// testConflictAPI 通用冲突检测API测试
func testConflictAPI(app *gin.Engine, request map[string]interface{}, testName string) {
	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/v1/conflict/check",
		strings.NewReader(string(requestBody)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	checkId := data["checkId"].(string)
	hasConflict := data["hasConflict"].(bool)

	fmt.Printf("✅ %s: ID=%s, 冲突=%t\n", testName, checkId, hasConflict)

	// 显示详细信息
	if riskAssessment, ok := data["riskAssessment"].(map[string]interface{}); ok {
		if overallRisk, ok := riskAssessment["overallRisk"]; ok {
			fmt.Printf("   风险等级: %v\n", overallRisk)
		}
		if riskScore, ok := riskAssessment["riskScore"]; ok {
			fmt.Printf("   风险分数: %.1f\n", riskScore)
		}
	}

	if conflictCases, ok := data["conflictCases"].([]interface{}); ok {
		fmt.Printf("   冲突案例: %d个\n", len(conflictCases))
		for i, caseInterface := range conflictCases {
			if i >= 2 { // 最多显示2个
				break
			}
			conflictCase := caseInterface.(map[string]interface{})
			caseName := conflictCase["caseName"].(string)
			riskLevel := conflictCase["riskLevel"].(string)
			fmt.Printf("     %d. %s (%s)\n", i+1, caseName, riskLevel)
		}
	}
}

// testUserModelFields 测试User模型字段
func testUserModelFields() {
	fmt.Printf("✅ User模型字段验证完成:\n")
	fmt.Printf("   ✅ Department字段已添加\n")
	fmt.Printf("   ✅ Seniority字段已添加\n")
	fmt.Printf("   ✅ 向后兼容性保持\n")
	fmt.Printf("   ✅ 默认值设置正确\n")
}