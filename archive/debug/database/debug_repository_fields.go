package main

import (
	"context"
	"encoding/json"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

func main() {
	// 数据库连接配置
	dsn := "root:@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("数据库连接失败:", err)
		return
	}

	fmt.Println("✅ 数据库连接成功")

	// 方法1: 直接使用GORM查询
	fmt.Println("\n🔍 方法1: 直接GORM查询")
	var directClient models.Client
	result := db.First(&directClient, 20)
	if result.Error != nil {
		fmt.Println("直接查询失败:", result.Error)
		return
	}

	fmt.Printf("直接查询结果: Type='%s'\n", directClient.Type)
	jsonDirect, _ := json.MarshalIndent(directClient, "", "  ")
	fmt.Println("直接查询JSON:")
	fmt.Println(string(jsonDirect))

	// 方法2: 使用Repository
	fmt.Println("\n🔍 方法2: 使用Repository查询")
	clientRepo := repositories.NewClientRepository(db)
	repoClients, _, err := clientRepo.List(context.Background(), &repositories.ClientListParams{
		Page:     1,
		PageSize: 1,
	})
	if err != nil {
		fmt.Println("Repository查询失败:", err)
		return
	}

	if len(repoClients) > 0 {
		repoClient := repoClients[0]
		fmt.Printf("Repository查询结果: Type='%s'\n", repoClient.Type)
		jsonRepo, _ := json.MarshalIndent(repoClient, "", "  ")
		fmt.Println("Repository查询JSON:")
		fmt.Println(string(jsonRepo))
	}

	// 方法3: 使用Repository的特定查询
	fmt.Println("\n🔍 方法3: Repository特定字段查询")
	var specificClient models.Client
	err = db.WithContext(context.Background()).Select("id", "name", "type", "email", "phone").First(&specificClient, 20).Error
	if err != nil {
		fmt.Println("特定字段查询失败:", err)
		return
	}

	fmt.Printf("特定字段查询结果: Type='%s'\n", specificClient.Type)
	jsonSpecific, _ := json.MarshalIndent(specificClient, "", "  ")
	fmt.Println("特定字段查询JSON:")
	fmt.Println(string(jsonSpecific))

	// 方法4: 检查Repository的查询SQL
	fmt.Println("\n🔍 方法4: 检查Repository查询的SQL")
	var sqlClient models.Client
	err = db.WithContext(context.Background()).Where("id = ?", 20).First(&sqlClient).Error
	if err != nil {
		fmt.Println("SQL查询失败:", err)
		return
	}

	fmt.Printf("SQL查询结果: Type='%s'\n", sqlClient.Type)
	jsonSQL, _ := json.MarshalIndent(sqlClient, "", "  ")
	fmt.Println("SQL查询JSON:")
	fmt.Println(string(jsonSQL))
}