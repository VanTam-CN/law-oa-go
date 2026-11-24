package database

import (
	_ "embed"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

//go:embed narrow_pilot_schema.sql
var narrowPilotSchemaQuery string

// RequireNarrowPilotSchema verifies the same catalog contract as the controlled 86 upgrade.
func RequireNarrowPilotSchema(db *gorm.DB) error {
	if db == nil || db.Dialector.Name() != "postgres" {
		return errors.New("M1 接案候选需要 PostgreSQL")
	}
	var ready bool
	if err := db.Raw(narrowPilotSchemaQuery).Scan(&ready).Error; err != nil {
		return fmt.Errorf("读取接案来源与回执表结构失败: %w", err)
	}
	if !ready {
		return errors.New("接案来源或回执表结构与 000086 不一致")
	}
	return nil
}
