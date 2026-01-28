package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

func main() {
	fmt.Println("🔍 检查数据库中的用户")
	fmt.Println("===================")

	// 读取配置
	viper.SetConfigFile("config.postgresql.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	// 连接数据库
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		viper.GetString("database.host"),
		viper.GetInt("database.port"),
		viper.GetString("database.username"),
		viper.GetString("database.password"),
		viper.GetString("database.database"),
		viper.GetString("database.sslmode"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 查询用户
	query := "SELECT id, username, email, role, status FROM users LIMIT 10"
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("查询用户失败: %v", err)
	}
	defer rows.Close()

	fmt.Println("数据库中的用户:")
	fmt.Println("ID | Username | Email | Role | Status")
	fmt.Println("---|----------|-------|------|-------")

	for rows.Next() {
		var id int
		var username, email, role, status string
		if err := rows.Scan(&id, &username, &email, &role, &status); err != nil {
			log.Printf("扫描行失败: %v", err)
			continue
		}
		fmt.Printf("%d | %s | %s | %s | %s\n", id, username, email, role, status)
	}

	// 检查客户数据
	fmt.Println("\n检查客户数据:")
	clientQuery := "SELECT id, name, type, status FROM clients LIMIT 10"
	rows, err = db.Query(clientQuery)
	if err != nil {
		log.Printf("查询客户失败: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("ID | Name | Type | Status")
	fmt.Println("---|------|------|-------")

	for rows.Next() {
		var id int
		var name, clientType, status string
		if err := rows.Scan(&id, &name, &clientType, &status); err != nil {
			log.Printf("扫描客户行失败: %v", err)
			continue
		}
		fmt.Printf("%d | %s | %s | %s\n", id, name, clientType, status)
	}
}