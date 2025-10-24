package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/redis/go-redis/v9"
)

func main() {
	fmt.Println("🔍 调试真实的请求数据...")

	// 数据库连接
	dsn := "host=localhost user=law_oa_user password=1q2w#E$R dbname=law_oa_db port=5432 sslmode=disable TimeZone=UTC"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// Redis连接
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 初始化仓库
	userRepo := repositories.NewUserRepository(db)
	clientRepo := repositories.NewClientRepository(db)
	caseRepo := repositories.NewCaseRepository(db)
	conflictRepo := repositories.NewConflictRepository(db, rdb)

	// 初始化风险评估器
	riskAssessor := services.NewRiskAssessor(nil, nil)

	// 初始化冲突检测服务
	conflictService := services.NewConflictDetectionService(
		conflictRepo,
		riskAssessor,
		userRepo,
		clientRepo,
		caseRepo,
	)

	// 使用前端发送的真实数据
	fmt.Println("\n📋 使用前端真实发送的数据:")

	realRequest := &models.ConflictCheckRequest{
		ClientID:                "57",  // 字节跳动科技有限公司
		ClientName:              "字节跳动科技有限公司",
		ClientType:              "PERSON",
		OtherParties:            []string{"腾讯\n垄断纠纷案"},
		CaseName:                "字节跳动诉腾讯垄断纠纷案",
		CaseType:                "commercial",
		SearchYears:             5,
		IncludeCorporateRelations: true,
		SearchDepth:             "DEEP",
		UserID:                  "45",  // 律师45
		RequestTime:             time.Now(),
	}

	requestJSON, _ := json.MarshalIndent(realRequest, "", "  ")
	fmt.Println("请求体:", string(requestJSON))

	// 验证用户ID转换
	fmt.Printf("\n🔍 验证用户ID转换:\n")
	fmt.Printf("前端发送的UserID: %s\n", realRequest.UserID)

	userIDUint, err := func(userID string) (uint64, error) {
		var result uint64
		_, err := fmt.Sscanf(userID, "%d", &result)
		return result, err
	}(realRequest.UserID)

	if err != nil {
		log.Fatalf("❌ 用户ID转换失败: %v", err)
	}

	lawyerID := uint(userIDUint)
	fmt.Printf("转换后的律师ID: %d\n", lawyerID)

	// 查询律师信息验证
	fmt.Printf("\n👤 验证律师%d的信息:\n", lawyerID)

	type LawyerInfo struct {
		ID   uint   `gorm:"column:id"`
		Name string `gorm:"column:name"`
		Role string `gorm:"column:role"`
	}

	var lawyerInfo LawyerInfo
	if err := db.WithContext(context.Background()).Table("users").
		Select("id, name, role").
		Where("id = ? AND deleted_at IS NULL", lawyerID).
		First(&lawyerInfo).Error; err != nil {
		fmt.Printf("⚠️ 查询律师信息失败: %v\n", err)
	} else {
		fmt.Printf("✅ 律师信息: ID=%d, 姓名=%s, 角色=%s\n", lawyerInfo.ID, lawyerInfo.Name, lawyerInfo.Role)
	}

	// 查询客户信息验证
	fmt.Printf("\n🏢 验证客户%s的信息:\n", realRequest.ClientID)

	clientIDUint, err := func(clientID string) (uint64, error) {
		var result uint64
		_, err := fmt.Sscanf(clientID, "%d", &result)
		return result, err
	}(realRequest.ClientID)

	if err != nil {
		log.Fatalf("❌ 客户ID转换失败: %v", err)
	}

	clientID := uint(clientIDUint)
	fmt.Printf("转换后的客户ID: %d\n", clientID)

	type ClientInfo struct {
		ID   uint   `gorm:"column:id"`
		Name string `gorm:"column:name"`
		Type string `gorm:"column:type"`
	}

	var clientInfo ClientInfo
	if err := db.WithContext(context.Background()).Table("clients").
		Select("id, name, type").
		Where("id = ? AND deleted_at IS NULL", clientID).
		First(&clientInfo).Error; err != nil {
		fmt.Printf("⚠️ 查询客户信息失败: %v\n", err)
	} else {
		fmt.Printf("✅ 客户信息: ID=%d, 名称=%s, 类型=%s\n", clientInfo.ID, clientInfo.Name, clientInfo.Type)
	}

	// 查询律师45的所有案件
	fmt.Printf("\n📁 查询律师%d的所有案件:\n", lawyerID)

	type LawyerCase struct {
		ID        uint      `gorm:"column:id"`
		Title     string    `gorm:"column:title"`
		ClientID  uint      `gorm:"column:client_id"`
		ClientName string   `gorm:"column:clients>name"`
		CaseType  string    `gorm:"column:case_type"`
		CreatedAt time.Time `gorm:"column:created_at"`
	}

	var lawyerCases []LawyerCase
	if err := db.WithContext(context.Background()).Table("cases").
		Select("cases.id, cases.title, cases.client_id, clients.name as client_name, cases.case_type, cases.created_at").
		Joins("LEFT JOIN clients ON cases.client_id = clients.id").
		Where("cases.lawyer_id = ? AND cases.deleted_at IS NULL", lawyerID).
		Order("cases.created_at DESC").
		Scan(&lawyerCases).Error; err != nil {
		fmt.Printf("⚠️ 查询律师案件失败: %v\n", err)
	} else {
		fmt.Printf("律师%d的案件数量: %d\n", lawyerID, len(lawyerCases))
		for i, case_ := range lawyerCases {
			fmt.Printf("  %d. ID:%d, 标题:%s, 客户:%s, 类型:%s\n",
				i+1, case_.ID, case_.Title, case_.ClientName, case_.CaseType)
		}
	}

	// 查询客户57的所有案件
	fmt.Printf("\n📁 查询客户%d的所有案件:\n", clientID)

	type ClientCase struct {
		ID        uint      `gorm:"column:id"`
		Title     string    `gorm:"column:title"`
		LawyerID  uint      `gorm:"column:lawyer_id"`
		LawyerName string   `gorm:"column:users>name"`
		CaseType  string    `gorm:"column:case_type"`
		CreatedAt time.Time `gorm:"column:created_at"`
	}

	var clientCases []ClientCase
	if err := db.WithContext(context.Background()).Table("cases").
		Select("cases.id, cases.title, cases.lawyer_id, users.name as lawyer_name, cases.case_type, cases.created_at").
		Joins("LEFT JOIN users ON cases.lawyer_id = users.id").
		Where("cases.client_id = ? AND cases.deleted_at IS NULL", clientID).
		Order("cases.created_at DESC").
		Scan(&clientCases).Error; err != nil {
		fmt.Printf("⚠️ 查询客户案件失败: %v\n", err)
	} else {
		fmt.Printf("客户%d的案件数量: %d\n", clientID, len(clientCases))
		for i, case_ := range clientCases {
			fmt.Printf("  %d. ID:%d, 标题:%s, 律师:%s, 类型:%s\n",
				i+1, case_.ID, case_.Title, case_.LawyerName, case_.CaseType)
		}
	}

	// 直接执行冲突检测
	fmt.Printf("\n🔍 执行冲突检测...\n")
	ctx := context.Background()
	response, err := conflictService.PerformConflictCheck(ctx, realRequest)
	if err != nil {
		log.Fatalf("❌ 冲突检测失败: %v", err)
	}

	// 输出结果
	fmt.Println("\n✅ 冲突检测结果:")
	fmt.Printf("检查ID: %s\n", response.CheckID)
	fmt.Printf("是否有冲突: %t\n", response.HasConflict)
	fmt.Printf("冲突案例数量: %d\n", len(response.ConflictCases))

	if response.HasConflict && len(response.ConflictCases) > 0 {
		fmt.Println("\n🔥 发现的冲突案例:")
		for i, conflict := range response.ConflictCases {
			fmt.Printf("  %d. 案件ID: %s, 名称: %s, 风险: %s\n",
				i+1, conflict.CaseID, conflict.CaseName, conflict.RiskLevel)
			fmt.Printf("     描述: %s\n", conflict.Description)
			fmt.Printf("     冲突类型: %s\n", conflict.ConflictType)
		}
	} else {
		fmt.Println("\n❌ 未发现任何冲突案例！")
	}

	fmt.Println("\n🎯 真实请求调试完成！")
}