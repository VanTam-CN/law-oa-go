package migrations

import (
	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// AddClientTypeFields 添加客户类型相关字段
func AddClientTypeFields(db *gorm.DB) error {
	// 检查字段是否已存在
	migrator := db.Migrator()

	// 添加 type 字段（如果不存在）
	if !migrator.HasColumn(&models.Client{}, "type") {
		if err := db.Exec("ALTER TABLE clients ADD COLUMN type VARCHAR(20) NOT NULL DEFAULT '个人' COMMENT '客户类型：个人/企业'").Error; err != nil {
			return err
		}
	}

	// 添加 id_card 字段（如果不存在）
	if !migrator.HasColumn(&models.Client{}, "id_card") {
		if err := db.Exec("ALTER TABLE clients ADD COLUMN id_card VARCHAR(18) COMMENT '身份证号（个人客户）'").Error; err != nil {
			return err
		}
	}

	// 添加 industry 字段（如果不存在）
	if !migrator.HasColumn(&models.Client{}, "industry") {
		if err := db.Exec("ALTER TABLE clients ADD COLUMN industry VARCHAR(50) COMMENT '所属行业（企业客户）'").Error; err != nil {
			return err
		}
	}

	// 添加 contact_person 字段（如果不存在）
	if !migrator.HasColumn(&models.Client{}, "contact_person") {
		if err := db.Exec("ALTER TABLE clients ADD COLUMN contact_person VARCHAR(50) COMMENT '联系人（企业客户）'").Error; err != nil {
			return err
		}
	}

	// 添加 contact_phone 字段（如果不存在）
	if !migrator.HasColumn(&models.Client{}, "contact_phone") {
		if err := db.Exec("ALTER TABLE clients ADD COLUMN contact_phone VARCHAR(20) COMMENT '联系电话（企业客户）'").Error; err != nil {
			return err
		}
	}

	// 添加 source 字段（如果不存在）
	if !migrator.HasColumn(&models.Client{}, "source") {
		if err := db.Exec("ALTER TABLE clients ADD COLUMN source VARCHAR(50) COMMENT '客户来源'").Error; err != nil {
			return err
		}
	}

	return nil
}