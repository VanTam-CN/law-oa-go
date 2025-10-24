package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/redis/go-redis/v9"
)

func main() {
	// 数据库连接 - 使用项目默认配置
	dsn := "host=localhost user=law_oa_user password=1q2w#E$R dbname=law_oa_db port=5432 sslmode=disable TimeZone=UTC"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// Redis连接（可选，用于测试）
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 直接创建冲突检测仓库
	conflictRepo := repositories.NewConflictRepository(db, rdb)

	// 创建测试请求 - 使用真实数据
	testRequest := &models.ConflictCheckRequest{
		ClientID:                "55",  // 阿里巴巴集团
		ClientName:              "阿里巴巴集团控股有限公司",
		ClientType:              "COMPANY",
		OtherParties:            []string{"字节跳动"},
		CaseName:                "测试案件",
		CaseType:                "civil",
		SearchYears:             5,
		IncludeCorporateRelations: true,
		SearchDepth:             "STANDARD",
		UserID:                  "45",  // 张伟律师
		RequestTime:             time.Now(),
	}

	log.Println("🧪 开始测试冲突检测仓库层修复...")
	log.Printf("测试请求: 客户ID=%s, 用户ID=%s, 案件类型=%s",
		testRequest.ClientID, testRequest.UserID, testRequest.CaseType)

	// 直接测试仓库层的GetPotentialConflicts方法
	ctx := context.Background()

	// 将UserID转换为uint（律师ID）
	var lawyerID uint
	if _, err := fmt.Sscanf(testRequest.UserID, "%d", &lawyerID); err != nil {
		log.Fatalf("❌ 用户ID格式错误: %s", testRequest.UserID)
	}

	conflicts, err := conflictRepo.GetPotentialConflicts(ctx, testRequest.ClientID, lawyerID, testRequest.OtherParties)
	if err != nil {
		log.Fatalf("❌ 冲突检测失败: %v", err)
	}

	// 输出结果
	log.Println("✅ 冲突检测仓库层测试成功!")
	log.Printf("找到的冲突案例数量: %d", len(conflicts))

	if len(conflicts) > 0 {
		log.Println("🔥 发现的冲突案例:")
		for i, conflict := range conflicts {
			log.Printf("  %d. ID: %s, 案件名: %s, 风险: %s",
				i+1, conflict.ID, conflict.CaseName, conflict.RiskLevel)
			log.Printf("     描述: %s", conflict.Description)
			log.Printf("     客户ID: %s, 冲突类型: %s", conflict.ClientID, conflict.ConflictType)
		}
	} else {
		log.Println("ℹ️ 未发现冲突案例")
	}

	// 验证数据库中的真实数据
	log.Println("\n📊 验证数据库中的相关数据...")

	// 查询律师45的案件
	var lawyerCases []struct {
		ID       uint   `gorm:"column:id"`
		Title    string `gorm:"column:title"`
		ClientID uint   `gorm:"column:client_id"`
		ClientName string `gorm:"column:clients>name"`
		CaseType string `gorm:"column:case_type"`
	}

	if err := db.WithContext(ctx).Table("cases").
		Select("cases.id, cases.title, cases.client_id, clients.name as client_name, cases.case_type").
		Joins("LEFT JOIN clients ON cases.client_id = clients.id").
		Where("cases.lawyer_id = ?", lawyerID).
		Where("cases.deleted_at IS NULL").
		Scan(&lawyerCases).Error; err != nil {
		log.Printf("⚠️ 查询律师案件失败: %v", err)
	} else {
		log.Printf("律师%d的案件数量: %d", lawyerID, len(lawyerCases))
		for _, case_ := range lawyerCases {
			log.Printf("  - ID:%d, 标题:%s, 客户:%s, 类型:%s",
				case_.ID, case_.Title, case_.ClientName, case_.CaseType)
		}
	}

	// 查询客户55的信息
	var clientInfo struct {
		ID   uint   `gorm:"column:id"`
		Name string `gorm:"column:name"`
		Type string `gorm:"column:type"`
	}

	if err := db.WithContext(ctx).Table("clients").
		Where("id = ?", 55).
		First(&clientInfo).Error; err != nil {
		log.Printf("⚠️ 查询客户信息失败: %v", err)
	} else {
		log.Printf("客户55信息: ID=%d, 名称=%s, 类型=%s",
			clientInfo.ID, clientInfo.Name, clientInfo.Type)
	}

	log.Println("\n🎯 测试完成！")
}