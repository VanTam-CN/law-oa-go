package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

func main() {
	// 数据库连接配置
	dsn := "root:@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	fmt.Println("✅ 数据库连接成功")

	// 开始迁移
	fmt.Println("🔧 开始添加客户类型相关字段...")

	migrator := db.Migrator()

	// 添加 type 字段（如果不存在）
	if !migrator.HasColumn(&models.Client{}, "type") {
		if err := db.Exec("ALTER TABLE clients ADD COLUMN type VARCHAR(20) NOT NULL DEFAULT '个人' COMMENT '客户类型：个人/企业'").Error; err != nil {
			log.Fatal("添加 type 字段失败:", err)
		}
		fmt.Println("✅ 添加 type 字段成功")
	} else {
		fmt.Println("ℹ️ type 字段已存在")
	}

	// 添加 id_card 字段（如果不存在）
	if !migrator.HasColumn(&models.Client{}, "id_card") {
		if err := db.Exec("ALTER TABLE clients ADD COLUMN id_card VARCHAR(18) COMMENT '身份证号（个人客户）'").Error; err != nil {
			log.Fatal("添加 id_card 字段失败:", err)
		}
		fmt.Println("✅ 添加 id_card 字段成功")
	} else {
		fmt.Println("ℹ️ id_card 字段已存在")
	}

	// 添加 industry 字段（如果不存在）
	if !migrator.HasColumn(&models.Client{}, "industry") {
		if err := db.Exec("ALTER TABLE clients ADD COLUMN industry VARCHAR(50) COMMENT '所属行业（企业客户）'").Error; err != nil {
			log.Fatal("添加 industry 字段失败:", err)
		}
		fmt.Println("✅ 添加 industry 字段成功")
	} else {
		fmt.Println("ℹ️ industry 字段已存在")
	}

	// 添加 contact_person 字段（如果不存在）
	if !migrator.HasColumn(&models.Client{}, "contact_person") {
		if err := db.Exec("ALTER TABLE clients ADD COLUMN contact_person VARCHAR(50) COMMENT '联系人（企业客户）'").Error; err != nil {
			log.Fatal("添加 contact_person 字段失败:", err)
		}
		fmt.Println("✅ 添加 contact_person 字段成功")
	} else {
		fmt.Println("ℹ️ contact_person 字段已存在")
	}

	// 添加 contact_phone 字段（如果不存在）
	if !migrator.HasColumn(&models.Client{}, "contact_phone") {
		if err := db.Exec("ALTER TABLE clients ADD COLUMN contact_phone VARCHAR(20) COMMENT '联系电话（企业客户）'").Error; err != nil {
			log.Fatal("添加 contact_phone 字段失败:", err)
		}
		fmt.Println("✅ 添加 contact_phone 字段成功")
	} else {
		fmt.Println("ℹ️ contact_phone 字段已存在")
	}

	// 添加 source 字段（如果不存在）
	if !migrator.HasColumn(&models.Client{}, "source") {
		if err := db.Exec("ALTER TABLE clients ADD COLUMN source VARCHAR(50) COMMENT '客户来源'").Error; err != nil {
			log.Fatal("添加 source 字段失败:", err)
		}
		fmt.Println("✅ 添加 source 字段成功")
	} else {
		fmt.Println("ℹ️ source 字段已存在")
	}

	// 更新现有客户的默认类型
	fmt.Println("🔧 更新现有客户的默认类型...")
	result := db.Exec("UPDATE clients SET type = '个人' WHERE type IS NULL OR type = ''")
	if result.Error != nil {
		log.Fatal("更新现有客户类型失败:", result.Error)
	}

	fmt.Printf("✅ 更新了 %d 个现有客户的类型\n", result.RowsAffected)

	// 验证字段是否添加成功
	fmt.Println("\n🔍 验证字段添加结果...")
	columns, err := db.Migrator().ColumnTypes(&models.Client{})
	if err != nil {
		log.Fatal("获取列信息失败:", err)
	}

	expectedFields := []string{"type", "id_card", "industry", "contact_person", "contact_phone", "source"}
	for _, field := range expectedFields {
		found := false
		for _, column := range columns {
			if column.Name() == field {
				found = true
				break
			}
		}
		if found {
			fmt.Printf("✅ %s 字段已正确添加\n", field)
		} else {
			fmt.Printf("❌ %s 字段未找到\n", field)
		}
	}

	fmt.Println("\n🎉 客户类型字段迁移完成！")
	fmt.Println("现在可以正常使用客户管理功能了。")
}