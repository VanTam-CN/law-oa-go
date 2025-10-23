package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
	log.Println("🚀 启动独立冲突检测服务...")
	log.Println("=" + "="*49)

	// 1. 初始化数据库
	log.Println("📋 初始化数据库...")
	db, err := initDatabase()
	if err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
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

	// 3. 创建服务实例
	log.Println("🔧 创建服务实例...")
	conflictService := createConflictDetectionService(db)

	// 4. 创建HTTP服务器
	log.Println("🌐 创建HTTP服务器...")
	app := createServer(conflictService)

	// 5. 启动服务器
	log.Println("🏃 启动服务器...")
	startServer(app)
}

// initDatabase 初始化数据库
func initDatabase() (*gorm.DB, error) {
	// 使用SQLite数据库文件
	db, err := gorm.Open(sqlite.Open("conflict_detection.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 自动迁移核心模型
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
		return nil, err
	}

	log.Println("✅ 数据库初始化完成: conflict_detection.db")
	return db, nil
}

// setupTestData 设置测试数据
func setupTestData(db *gorm.DB) error {
	log.Println("创建测试用户...")

	// 创建测试律师用户
	lawyers := []models.User{
		{
			Username:   "lawyer_zhang",
			Name:       "张律师",
			Email:      "zhang@law.com",
			Role:       "lawyer",
			Department: "诉讼部", // 测试新增字段
			Seniority:  "中级",   // 测试新增字段
			Phone:      "13800138001",
			Status:     "active",
		},
		{
			Username:   "lawyer_li",
			Name:       "李律师",
			Email:      "li@law.com",
			Role:       "lawyer",
			Department: "合规部",
			Seniority:  "高级",
			Phone:      "13800138002",
			Status:     "active",
		},
		{
			Username:   "lawyer_wang",
			Name:       "王律师",
			Email:      "wang@law.com",
			Role:       "lawyer",
			Department: "公司部",
			Seniority:  "初级",
			Phone:      "13800138003",
			Status:     "active",
		},
	}

	for _, lawyer := range lawyers {
		var existingUser models.User
		if err := db.Where("email = ?", lawyer.Email).First(&existingUser).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&lawyer).Error; err != nil {
				return err
			}
			log.Printf("✅ 创建律师: %s (%s)", lawyer.Name, lawyer.Department)
		} else {
			log.Printf("ℹ️ 律师已存在: %s", lawyer.Name)
		}
	}

	log.Println("创建测试客户...")

	// 创建测试客户
	clients := []models.Client{
		{
			Name:          "腾讯科技",
			Type:          "COMPANY",
			Email:         "tencent@example.com",
			Phone:         "0755-86013388",
			Address:       "深圳市南山区科技园",
			Company:       "腾讯科技有限公司",
			Industry:      "互联网科技",
			ContactPerson: "马总监",
			ContactPhone:  "13800138000",
			Source:        "推荐",
			Notes:         "大型互联网科技公司",
			Status:        "active",
		},
		{
			Name:          "字节跳动",
			Type:          "COMPANY",
			Email:         "bytedance@example.com",
			Phone:         "010-12345678",
			Address:       "北京市海淀区知春路",
			Company:       "字节跳动科技有限公司",
			Industry:      "互联网科技",
			ContactPerson: "张经理",
			ContactPhone:  "13900139000",
			Source:        "推荐",
			Notes:         "短视频和内容平台公司",
			Status:        "active",
		},
		{
			Name:          "阿里巴巴",
			Type:          "COMPANY",
			Email:         "alibaba@example.com",
			Phone:         "0571-85022088",
			Address:       "杭州市余杭区文一西路",
			Company:       "阿里巴巴集团控股有限公司",
			Industry:      "电子商务",
			ContactPerson: "刘总",
			ContactPhone:  "13700137000",
			Source:        "推荐",
			Notes:         "电子商务和云计算公司",
			Status:        "active",
		},
		{
			Name:          "美团",
			Type:          "COMPANY",
			Email:         "meituan@example.com",
			Phone:         "010-11111111",
			Address:       "北京市朝阳区望京东",
			Company:       "北京三快科技有限公司",
			Industry:      "本地生活",
			ContactPerson: "王总监",
			ContactPhone:  "13600136000",
			Source:        "推荐",
			Notes:         "本地生活服务平台",
			Status:        "active",
		},
	}

	for _, client := range clients {
		var existingClient models.Client
		if err := db.Where("email = ?", client.Email).First(&existingClient).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&client).Error; err != nil {
				return err
			}
			log.Printf("✅ 创建客户: %s (%s)", client.Name, client.Industry)
		} else {
			log.Printf("ℹ️ 客户已存在: %s", client.Name)
		}
	}

	log.Println("创建测试案件...")

	// 创建测试案件（用于冲突检测测试）
	cases := []models.Case{
		{
			Title:       "腾讯诉字节跳动不正当竞争案",
			Description: "涉及短视频平台的版权和不正当竞争纠纷",
			ClientID:    1, // 腾讯
			LawyerID:    1, // 张律师
			CaseType:    "知识产权",
			Priority:    "high",
			Status:      "active",
			StartDate:   &[]time.Time{time.Now().AddDate(0, -6, 0)}[0],
		},
		{
			Title:       "阿里巴巴商业秘密保护案",
			Description: "涉及电商平台商业秘密保护",
			ClientID:    3, // 阿里巴巴
			LawyerID:    1, // 张律师
			CaseType:    "商业纠纷",
			Priority:    "medium",
			Status:      "active",
			StartDate:   &[]time.Time{time.Now().AddDate(0, -3, 0)}[0],
		},
		{
			Title:       "美团平台合作协议纠纷",
			Description: "涉及外卖平台合作协议纠纷",
			ClientID:    4, // 美团
			LawyerID:    2, // 李律师
			CaseType:    "合同纠纷",
			Priority:    "medium",
			Status:      "active",
			StartDate:   &[]time.Time{time.Now().AddDate(0, -2, 0)}[0],
		},
	}

	for _, caseModel := range cases {
		var existingCase models.Case
		if err := db.Where("title = ?", caseModel.Title).First(&existingCase).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&caseModel).Error; err != nil {
				return err
			}
			log.Printf("✅ 创建案件: %s", caseModel.Title)
		} else {
			log.Printf("ℹ️ 案件已存在: %s", caseModel.Title)
		}
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

	// 创建冲突仓储（简化版本，不使用Redis）
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

// createServer 创建HTTP服务器
func createServer(conflictService services.ConflictDetectionService) *gin.Engine {
	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()

	// 添加中间件
	app.Use(gin.Logger())
	app.Use(gin.Recovery())
	app.Use(corsMiddleware())

	// 创建冲突检测处理器
	conflictHandler := handlers.NewConflictHandlerSimple(conflictService)

	// API路由组
	api := app.Group("/api/v1")
	{
		// 公开路由（无需认证）
		public := api.Group("/")
		{
			conflict := public.Group("conflict")
			{
				conflict.POST("/check", conflictHandler.CheckConflict)
				conflict.GET("/health", conflictHandler.HealthCheck)
				conflict.GET("/history", conflictHandler.GetCheckHistory)
				conflict.GET("/stats", conflictHandler.GetConflictStats)
			}
		}
	}

	// 根路径
	app.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":     "独立冲突检测服务",
			"version":     "1.0.0",
			"status":      "running",
			"endpoints": []string{
				"POST /api/v1/conflict/check - 执行冲突检测",
				"GET /api/v1/conflict/health - 健康检查",
				"GET /api/v1/conflict/history - 获取检测历史",
				"GET /api/v1/conflict/stats - 获取统计数据",
			},
		})
	})

	log.Println("✅ HTTP服务器创建成功")
	return app
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

// startServer 启动服务器
func startServer(app *gin.Engine) {
	server := &http.Server{
		Addr:    ":8080",
		Handler: app,
	}

	// 启动服务器
	go func() {
		log.Printf("🌐 服务器启动在: http://localhost:8080")
		log.Println("📋 可用端点:")
		log.Println("  GET  / - 服务信息")
		log.Println("  POST /api/v1/conflict/check - 冲突检测")
		log.Println("  GET  /api/v1/conflict/health - 健康检查")
		log.Println("  GET  /api/v1/conflict/history - 检测历史")
		log.Println("  GET  /api/v1/conflict/stats - 统计数据")
		log.Println("=" + "="*49)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ 服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号以优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 正在关闭服务器...")

	// 创建一个超时上下文进行关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("❌ 服务器强制关闭: %v", err)
	}

	log.Println("✅ 服务器已关闭")
}