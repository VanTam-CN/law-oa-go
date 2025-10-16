package main

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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

	// 检查clients表结构
	fmt.Println("\n🔍 检查clients表结构:")
	var columns []struct {
		Field   string `gorm:"column:Field"`
		Type    string `gorm:"column:Type"`
		Null    string `gorm:"column:Null"`
		Key     string `gorm:"column:Key"`
		Default string `gorm:"column:Default"`
		Extra   string `gorm:"column:Extra"`
	}

	err = db.Raw("SHOW COLUMNS FROM clients").Scan(&columns).Error
	if err != nil {
		fmt.Println("查询表结构失败:", err)
		return
	}

	fmt.Printf("clients表有 %d 个字段:\n", len(columns))
	for _, col := range columns {
		fmt.Printf("- %s: %s %s\n", col.Field, col.Type, col.Null)
	}

	// 检查是否有其他表可能包含客户数据
	fmt.Println("\n🔍 查找可能包含客户数据的表:")
	var tables []struct {
	TableName string `gorm:"column:table_name"`
	}

	err = db.Raw("SHOW TABLES").Scan(&tables).Error
	if err != nil {
		fmt.Println("查询表列表失败:", err)
		return
	}

	fmt.Printf("数据库中有 %d 个表:\n", len(tables))
	for _, table := range tables {
		if len(table.TableName) > 0 {
			fmt.Printf("- %s\n", table.TableName)
		}
	}

	// 特别检查是否有以client开头的表
	fmt.Println("\n🔍 检查以client开头的表:")
	for _, table := range tables {
		if len(table.TableName) > 6 && table.TableName[:6] == "client" {
			fmt.Printf("发现表: %s\n", table.TableName)

			// 检查表结构
			var clientColumns []struct {
				Field string `gorm:"column:Field"`
				Type  string `gorm:"column:Type"`
			}
			err = db.Raw(fmt.Sprintf("SHOW COLUMNS FROM %s", table.TableName)).Scan(&clientColumns).Error
			if err == nil {
				fmt.Printf("  表 %s 的字段:\n", table.TableName)
				for _, col := range clientColumns {
					fmt.Printf("  - %s: %s\n", col.Field, col.Type)
				}
			}
		}
	}
}