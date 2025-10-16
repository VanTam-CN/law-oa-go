package main

import (
	"fmt"
	"log"
	"strings"

	"law-oa-go/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 数据库连接字符串 - 使用调试工具中验证成功的连接
	dsns := []string{
		"root:@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local",
	}

	var db *gorm.DB
	var err error
	var successfulDSN string

	for _, dsn := range dsns {
		fmt.Printf("尝试连接数据库: %s\n", maskPassword(dsn))
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err == nil {
			fmt.Println("✅ 数据库连接成功!")
			successfulDSN = dsn
			break
		}
		fmt.Printf("❌ 连接失败: %v\n", err)
	}

	if err != nil {
		log.Fatal("所有数据库连接尝试均失败")
	}

	fmt.Printf("使用连接: %s\n", maskPassword(successfulDSN))

	fmt.Println("=== 检查律师相关数据 ===")

	// 检查是否有专门的律师表
	var lawyerModels []struct {
	TableName string
	}

	// 查询所有表名
	if err := db.Raw("SHOW TABLES").Scan(&lawyerModels).Error; err != nil {
		fmt.Printf("查询表名失败: %v\n", err)
	} else {
		fmt.Println("数据库中的表:")
		for _, table := range lawyerModels {
			tableName := fmt.Sprintf("%v", table.TableName)
			if strings.Contains(strings.ToLower(tableName), "lawyer") || strings.Contains(strings.ToLower(tableName), "user") {
				fmt.Printf("  - %s\n", tableName)
			}
		}
	}

	// 检查用户表中是否有律师数据
	fmt.Println("\n=== 检查用户表中的律师数据 ===")
	var users []models.User
	if err := db.Where("role = ? OR role = ?", "lawyer", "律师").Find(&users).Error; err != nil {
		fmt.Printf("查询用户表失败: %v\n", err)
	} else {
		fmt.Printf("找到 %d 个律师用户:\n\n", len(users))

		if len(users) == 0 {
			// 如果没有找到律师用户，检查所有用户
			fmt.Println("没有找到标记为律师的用户，检查所有用户...")
			if err := db.Find(&users).Error; err != nil {
				fmt.Printf("查询所有用户失败: %v\n", err)
			} else {
				fmt.Printf("数据库中共有 %d 个用户:\n\n", len(users))
			}
		}

		for i, user := range users {
			fmt.Printf("用户 %d:\n", i+1)
			fmt.Printf("  ID: %d\n", user.ID)
			fmt.Printf("  用户名: '%s'\n", user.Username)
			fmt.Printf("  姓名: '%s'\n", user.Name)
			fmt.Printf("  邮箱: '%s'\n", user.Email)
			fmt.Printf("  电话: '%s'\n", user.Phone)
			fmt.Printf("  角色: '%s'\n", user.Role)
			fmt.Printf("  状态: '%s'\n", user.Status)
			fmt.Printf("  头像: '%s'\n", user.Avatar)
			fmt.Printf("  创建时间: %s\n", user.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Println("---")
		}
	}

	// 统计用户角色
	fmt.Println("\n=== 用户角色统计 ===")
	var roleStats []struct {
		Role  string
		Count int64
	}

	if err := db.Raw("SELECT role, COUNT(*) as count FROM users GROUP BY role").Scan(&roleStats).Error; err != nil {
		fmt.Printf("统计用户角色失败: %v\n", err)
	} else {
		for _, stat := range roleStats {
			fmt.Printf("角色 '%s': %d 个用户\n", stat.Role, stat.Count)
		}
	}

	// 检查是否有专门的律师表
	fmt.Println("\n=== 检查是否有专门的律师表 ===")

	// 尝试查询一个可能的律师表结构
	var lawyerExists bool
	if err := db.Raw("SHOW TABLES LIKE '%lawyer%'").Scan(&lawyerExists).Error; err == nil && lawyerExists {
		fmt.Println("发现律师相关的表，尝试查询...")
		// 这里可以添加具体的律师表查询逻辑
	} else {
		fmt.Println("没有发现专门的律师表，律师数据可能在用户表中")
	}

	// 尝试创建一个简单的律师用户作为测试
	fmt.Println("\n=== 创建测试律师用户（如果需要）===")
	testLawyer := models.User{
		Username:  "test_lawyer_001",
		Name:      "测试律师",
		Email:     "test.lawyer@example.com",
		Password:  "test_password_hash", // 实际应用中需要加密
		Role:      "lawyer",
		Phone:     "13800138000",
		Avatar:    "",
		Status:    "active",
	}

	// 检查是否已存在
	var existingUser models.User
	result := db.Where("username = ?", testLawyer.Username).First(&existingUser)
	if result.Error != nil {
		fmt.Println("创建测试律师用户...")
		if err := db.Create(&testLawyer).Error; err != nil {
			fmt.Printf("创建测试律师用户失败: %v\n", err)
		} else {
			fmt.Printf("✅ 成功创建测试律师用户: %s (ID: %d)\n", testLawyer.Username, testLawyer.ID)
		}
	} else {
		fmt.Printf("测试律师用户已存在: %s (ID: %d)\n", existingUser.Username, existingUser.ID)
	}
}

// maskPassword 隐藏密码显示
func maskPassword(dsn string) string {
	if strings.Contains(dsn, ":password@") {
		return strings.Replace(dsn, "password", "***", 1)
	} else if strings.Contains(dsn, ":root@") {
		return strings.Replace(dsn, "root", "***", 1)
	} else if strings.Contains(dsn, ":123456@") {
		return strings.Replace(dsn, "123456", "***", 1)
	}
	return dsn
}