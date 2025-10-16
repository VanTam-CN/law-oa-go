package main

import (
	"context"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
)

func main() {
	// 数据库连接
	dsn := "law_oa:law_oa_password@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 创建案件服务
	caseService := services.NewCaseService(db, nil, false)

	// 获取第一个案件
	var firstCase models.Case
	if err := db.First(&firstCase).Error; err != nil {
		log.Fatal("获取案件失败:", err)
	}

	fmt.Printf("测试案件ID: %d\n", firstCase.ID)

	// 调用GetCaseByID方法
	caseResponse, err := caseService.GetCaseByID(context.Background(), firstCase.ID)
	if err != nil {
		log.Fatal("获取案件详情失败:", err)
	}

	// 打印完整的响应数据
	fmt.Printf("\n=== CaseResponse 原始数据 ===\n")
	fmt.Printf("ID: %d\n", caseResponse.ID)
	fmt.Printf("Title: %s\n", caseResponse.Title)
	fmt.Printf("ClientID: %d\n", caseResponse.ClientID)
	fmt.Printf("LawyerID: %d\n", caseResponse.LawyerID)
	fmt.Printf("ClientName: '%s'\n", caseResponse.ClientName)
	fmt.Printf("LawyerName: '%s'\n", caseResponse.LawyerName)

	fmt.Printf("\n=== Client 对象 ===\n")
	if caseResponse.Client != nil {
		fmt.Printf("Client.ID: %d\n", caseResponse.Client.ID)
		fmt.Printf("Client.Name: '%s'\n", caseResponse.Client.Name)
		fmt.Printf("Client.Company: '%s'\n", caseResponse.Client.Company)
		fmt.Printf("Client.Email: '%s'\n", caseResponse.Client.Email)
		fmt.Printf("Client.Phone: '%s'\n", caseResponse.Client.Phone)
	} else {
		fmt.Println("Client 对象为 nil")
	}

	fmt.Printf("\n=== Lawyer 对象 ===\n")
	if caseResponse.Lawyer != nil {
		fmt.Printf("Lawyer.ID: %d\n", caseResponse.Lawyer.ID)
		fmt.Printf("Lawyer.Name: '%s'\n", caseResponse.Lawyer.Name)
		fmt.Printf("Lawyer.Email: '%s'\n", caseResponse.Lawyer.Email)
		fmt.Printf("Lawyer.Role: '%s'\n", caseResponse.Lawyer.Role)
	} else {
		fmt.Println("Lawyer 对象为 nil")
	}

	// 检查JSON序列化
	fmt.Printf("\n=== JSON序列化测试 ===\n")

	// 创建简化的结构体用于测试
	type SimpleCaseResponse struct {
		ID         uint                   `json:"id"`
		Title      string                 `json:"title"`
		ClientID   uint                   `json:"client_id"`
		LawyerID   uint                   `json:"lawyer_id"`
		ClientName string                 `json:"client_name,omitempty"`
		LawyerName string                 `json:"lawyer_name,omitempty"`
		Client     *models.Client         `json:"client,omitempty"`
		Lawyer     *models.User           `json:"lawyer,omitempty"`
	}

	simpleResponse := &SimpleCaseResponse{
		ID:         caseResponse.ID,
		Title:      caseResponse.Title,
		ClientID:   caseResponse.ClientID,
		LawyerID:   caseResponse.LawyerID,
		ClientName: caseResponse.ClientName,
		LawyerName: caseResponse.LawyerName,
		Client:     caseResponse.Client,
		Lawyer:     caseResponse.Lawyer,
	}

	// 手动序列化
	if simpleResponse.Client != nil {
		fmt.Printf("Simple.Client.ID: %d\n", simpleResponse.Client.ID)
		fmt.Printf("Simple.Client.Name: '%s'\n", simpleResponse.Client.Name)
		fmt.Printf("Simple.Client.Company: '%s'\n", simpleResponse.Client.Company)
	}
	if simpleResponse.Lawyer != nil {
		fmt.Printf("Simple.Lawyer.ID: %d\n", simpleResponse.Lawyer.ID)
		fmt.Printf("Simple.Lawyer.Name: '%s'\n", simpleResponse.Lawyer.Name)
		fmt.Printf("Simple.Lawyer.Role: '%s'\n", simpleResponse.Lawyer.Role)
	}
}