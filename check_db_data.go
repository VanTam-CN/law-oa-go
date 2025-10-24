package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string
	Email     string
	Role      string
	Status    string
	CreatedAt time.Time
}

type Client struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string
	Type     string
	LawyerID uint
}

type Case struct {
	ID          uint       `gorm:"primaryKey"`
	Title       string
	ClientID    uint
	LawyerID    uint
	CaseType    string
	Description string
	DeletedAt   *time.Time `gorm:"index"`
	CreatedAt   time.Time
}

func main() {
	// 数据库连接 - 使用正确的配置
	dsn := "host=localhost port=5432 user=law_oa_user password=1q2w#E$R dbname=law_oa_db sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	// 检查用户数据
	var users []User
	if err := db.Where("role = ?", "lawyer").Find(&users).Error; err != nil {
		log.Fatal("查询用户失败:", err)
	}

	fmt.Println("=== 律师用户数据 ===")
	for _, user := range users {
		fmt.Printf("ID: %d, 姓名: %s, 邮箱: %s, 状态: %s\n", user.ID, user.Name, user.Email, user.Status)
	}

	// 检查客户数据
	var clients []Client
	if err := db.Find(&clients).Error; err != nil {
		log.Fatal("查询客户失败:", err)
	}

	fmt.Println("\n=== 客户数据 ===")
	for _, client := range clients {
		fmt.Printf("ID: %d, 名称: %s, 类型: %s, 律师ID: %d\n", client.ID, client.Name, client.Type, client.LawyerID)
	}

	// 检查案件数据
	var cases []Case
	if err := db.Where("deleted_at IS NULL").Find(&cases).Error; err != nil {
		log.Fatal("查询案件失败:", err)
	}

	fmt.Println("\n=== 案件数据 ===")
	for _, case_ := range cases {
		fmt.Printf("ID: %d, 标题: %s, 客户ID: %d, 律师ID: %d, 类型: %s\n", case_.ID, case_.Title, case_.ClientID, case_.LawyerID, case_.CaseType)
	}

	// 特别检查关键冲突案件
	fmt.Println("\n=== 关键冲突案件检查 ===")
	var conflictCases []Case
	if err := db.Where("deleted_at IS NULL AND (title ILIKE ? OR title ILIKE ? OR title ILIKE ?)",
		"%阿里巴巴%", "%字节跳动%", "%腾讯%").Find(&conflictCases).Error; err != nil {
		log.Fatal("查询冲突案件失败:", err)
	}

	for _, case_ := range conflictCases {
		fmt.Printf("冲突案件: ID=%d, 标题=%s, 律师ID=%d, 客户ID=%d\n", case_.ID, case_.Title, case_.LawyerID, case_.ClientID)
	}
}