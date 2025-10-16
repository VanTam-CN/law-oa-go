package main

import (
	"context"
	"fmt"
	"log"

	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 数据库连接字符串
	dsn := "root:@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local"

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	fmt.Println("✅ 数据库连接成功")

	// 创建repository和服务
	lawyerRepo := repositories.NewLawyerRepository(db)
	lawyerService := services.NewLawyerService(lawyerRepo)

	fmt.Println("=== 测试律师服务 ===")

	// 测试获取律师列表
	req := &services.LawyerListRequest{
		Page:     1,
		PageSize: 5,
		Search:   "",
	}

	lawyers, total, err := lawyerService.ListLawyers(context.Background(), req)
	if err != nil {
		fmt.Printf("❌ 获取律师列表失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 成功获取律师列表，共 %d 条记录，总数 %d\n", len(lawyers), total)

	// 显示律师数据
	for i, lawyer := range lawyers {
		fmt.Printf("\n律师 %d:\n", i+1)
		fmt.Printf("  ID: %d\n", lawyer.ID)
		fmt.Printf("  姓名: '%s'\n", lawyer.Name)
		fmt.Printf("  邮箱: '%s'\n", lawyer.Email)
		fmt.Printf("  电话: '%s'\n", lawyer.Phone)
		fmt.Printf("  执业证号: '%s'\n", lawyer.LicenseNumber)
		fmt.Printf("  专业领域: '%s'\n", lawyer.Specialty)
		fmt.Printf("  部门: '%s'\n", lawyer.Department)
		fmt.Printf("  职位: '%s'\n", lawyer.Position)
		fmt.Printf("  状态: '%s'\n", lawyer.Status)
		fmt.Printf("  从业年限: %d年\n", lawyer.Experience)
		fmt.Printf("  入职日期: '%s'\n", lawyer.JoinDate)
		fmt.Printf("  性别: '%s'\n", lawyer.Gender)
		fmt.Printf("  头像: '%s'\n", lawyer.Avatar)
		fmt.Printf("  个人简介: '%s'\n", lawyer.Profile)
		fmt.Printf("  创建时间: %s\n", lawyer.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println("---")
	}

	// 测试搜索功能
	fmt.Println("\n=== 测试搜索功能 ===")
	searchReq := &services.LawyerListRequest{
		Page:     1,
		PageSize: 5,
		Search:   "张",
	}

	searchLawyers, searchTotal, err := lawyerService.ListLawyers(context.Background(), searchReq)
	if err != nil {
		fmt.Printf("❌ 搜索律师失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 搜索'张'结果，共 %d 条记录，总数 %d\n", len(searchLawyers), searchTotal)
	for _, lawyer := range searchLawyers {
		fmt.Printf("  - %s (%s)\n", lawyer.Name, lawyer.Email)
	}

	fmt.Println("\n🎉 律师服务测试完成！")
}