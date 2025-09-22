package models

import (
	"time"

	"gorm.io/gorm"
)

// Document 文档模型
type Document struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	Name        string         `json:"name" gorm:"size:200;not null"`
	Description string         `json:"description" gorm:"type:text"`
	Filename    string         `json:"filename" gorm:"size:255;not null"`
	Filepath    string         `json:"filepath" gorm:"size:500;not null"`
	Filesize    int64          `json:"filesize"`
	MimeType    string         `json:"mime_type" gorm:"size:100"`
	Category    string         `json:"category" gorm:"size:100"`
	Tags        string         `json:"tags" gorm:"size:500"`
	EntityID    uint           `json:"entity_id" gorm:"index"`
	EntityType  string         `json:"entity_type" gorm:"size:50"`
	Status      string         `json:"status" gorm:"size:20;default:'active'"`
}

// TableName 指定表名
func (Document) TableName() string {
	return "documents"
}

// BeforeCreate 创建前钩子
func (d *Document) BeforeCreate(tx *gorm.DB) error {
	// 确保状态默认为active
	if d.Status == "" {
		d.Status = "active"
	}
	return nil
}
