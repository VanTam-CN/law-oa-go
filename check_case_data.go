package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

func main() {
	// 数据库连接
	dsn := "law_oa:law_oa_password@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 检查案件数据
	fmt.Println("=== 检查案件数据 ===")
	var cases []models.Case
	if err := db.Find(&cases).Error; err != nil {
		log.Fatal("查询案件失败:", err)
	}

	if len(cases) == 0 {
		fmt.Println("没有找到案件数据")
		return
	}

	// 检查第一个案件的详细信息
	caseItem := cases[0]
	fmt.Printf("案件ID: %d\n", caseItem.ID)
	fmt.Printf("案件标题: %s\n", caseItem.Title)
	fmt.Printf("客户ID: %d\n", caseItem.ClientID)
	fmt.Printf("律师ID: %d\n", caseItem.LawyerID)
	fmt.Printf("案件类型: %s\n", caseItem.CaseType)
	fmt.Printf("状态: %s\n", caseItem.Status)
	fmt.Printf("优先级: %s\n", caseItem.Priority)

	// 检查关联的客户数据
	fmt.Println("\n=== 检查客户数据 ===")
	var client models.Client
	if err := db.First(&client, caseItem.ClientID).Error; err != nil {
		fmt.Printf("客户数据查询失败: %v\n", err)
	} else {
		fmt.Printf("客户ID: %d\n", client.ID)
		fmt.Printf("客户姓名: %s\n", client.Name)
		fmt.Printf("客户公司: %s\n", client.Company)
		fmt.Printf("客户邮箱: %s\n", client.Email)
		fmt.Printf("客户电话: %s\n", client.Phone)
	}

	// 检查关联的律师数据
	fmt.Println("\n=== 检查律师数据 ===")
	var lawyer models.User
	if err := db.First(&lawyer, caseItem.LawyerID).Error; err != nil {
		fmt.Printf("律师数据查询失败: %v\n", err)
	} else {
		fmt.Printf("律师ID: %d\n", lawyer.ID)
		fmt.Printf("律师姓名: %s\n", lawyer.Name)
		fmt.Printf("律师邮箱: %s\n", lawyer.Email)
		fmt.Printf("律师角色: %s\n", lawyer.Role)
	}

	// 测试带预加载的查询
	fmt.Println("\n=== 测试预加载查询 ===")
	var caseWithRelations models.Case
	if err := db.Preload("Client").Preload("Lawyer").First(&caseWithRelations, caseItem.ID).Error; err != nil {
		fmt.Printf("预加载查询失败: %v\n", err)
	} else {
		fmt.Printf("案件ID: %d\n", caseWithRelations.ID)
		if caseWithRelations.Client != nil {
			fmt.Printf("预加载的客户姓名: %s\n", caseWithRelations.Client.Name)
			fmt.Printf("预加载的客户公司: %s\n", caseWithRelations.Client.Company)
		} else {
			fmt.Println("预加载的客户为空")
		}
		if caseWithRelations.Lawyer != nil {
			fmt.Printf("预加载的律师姓名: %s\n", caseWithRelations.Lawyer.Name)
		} else {
			fmt.Println("预加载的律师为空")
		}
	}
}