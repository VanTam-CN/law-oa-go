package models

import (
	"time"

	"gorm.io/gorm"
)

// Document 文档模型
type Document struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	UUID          string    `json:"uuid" gorm:"type:varchar(36);uniqueIndex;not null"`
	TenantID      string    `json:"tenant_id" gorm:"type:varchar(64);not null;index"`
	Name          string    `json:"name" gorm:"type:varchar(255);not null"`
	Description   string    `json:"description" gorm:"type:text"`
	OriginalName  string    `json:"original_name" gorm:"type:varchar(255)"`
	MIMEType      string    `json:"mime_type" gorm:"type:varchar(100);not null"`
	Size          int64     `json:"size" gorm:"not null"`
	Category      string    `json:"category" gorm:"type:varchar(50);index"`
	Tags          string    `json:"tags" gorm:"type:text"`
	EntityType    string    `json:"entity_type" gorm:"type:varchar(50);index"`
	EntityID      uint      `json:"entity_id" gorm:"index"`
	CurrentVersion uint    `json:"current_version" gorm:"default:1"`
	Status        string    `json:"status" gorm:"type:varchar(20);default:'active';index"`
	CreatedBy      uint      `json:"created_by" gorm:"not null"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" gorm:"index"`

	// 关联
	Versions []DocumentVersion `json:"versions,omitempty" gorm:"foreignKey:DocumentID"`
}

// DocumentVersion 文档版本模型
type DocumentVersion struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	DocumentID  uint      `json:"document_id" gorm:"not null;index"`
	Version     int       `json:"version" gorm:"not null"`
	UUID        string    `json:"uuid" gorm:"type:varchar(36);uniqueIndex;not null"`
	StoragePath string    `json:"storage_path" gorm:"type:varchar(512);not null"`
	FileHash    string    `json:"file_hash" gorm:"type:varchar(64);not null;index"`
	Size        int64     `json:"size" gorm:"not null"`
	Description string    `json:"description" gorm:"type:text"`
	CreatedBy   uint      `json:"created_by" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"`

	// 关联
	Document Document `json:"document" gorm:"foreignKey:DocumentID"`
}

// DocumentPermission 文档权限模型
type DocumentPermission struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	DocumentID uint      `json:"document_id" gorm:"not null;index"`
	UserID     *uint     `json:"user_id" gorm:"index"`
	RoleID     *uint     `json:"role_id" gorm:"index"`
	TenantID   string    `json:"tenant_id" gorm:"type:varchar(64);not null;index"`
	Permission string    `json:"permission" gorm:"type:varchar(50);not null"` // read, write, delete, admin
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// 关联
	Document *Document `json:"document,omitempty" gorm:"foreignKey:DocumentID"`
}

// DocumentAudit 文档审计日志模型
type DocumentAudit struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	DocumentID uint      `json:"document_id" gorm:"not null;index"`
	UserID     uint      `json:"user_id" gorm:"not null;index"`
	TenantID   string    `json:"tenant_id" gorm:"type:varchar(64);not null;index"`
	Action     string    `json:"action" gorm:"type:varchar(50);not null"` // create, read, update, delete, download, share
	Details    string    `json:"details" gorm:"type:text"`
	IPAddress  string    `json:"ip_address" gorm:"type:varchar(45)"`
	UserAgent  string    `json:"user_agent" gorm:"type:varchar(512)"`
	CreatedAt  time.Time `json:"created_at"`

	// 关联
	Document *Document `json:"document,omitempty" gorm:"foreignKey:DocumentID"`
}

// DocumentIndex 文档搜索索引模型 (用于Elasticsearch)
type DocumentIndex struct {
	ID           string    `json:"id"`                   // Document UUID
	TenantID     string    `json:"tenant_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Content      string    `json:"content"`              // OCR提取的内容
	Tags         []string  `json:"tags"`
	Category     string    `json:"category"`
	EntityType   string    `json:"entity_type"`
	EntityID     uint      `json:"entity_id"`
	MIMEType     string    `json:"mime_type"`
	Size         int64     `json:"size"`
	CreatedBy    uint      `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Version      int       `json:"version"`
	FileHash     string    `json:"file_hash"`
}

// User 用户模型 (简化版，用于权限控制)
type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Username string `json:"username" gorm:"type:varchar(64);uniqueIndex;not null"`
	Email    string `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
	TenantID string `json:"tenant_id" gorm:"type:varchar(64);not null;index"`
	IsActive bool   `json:"is_active" gorm:"default:true"`
}

// Role 角色模型
type Role struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Name      string `json:"name" gorm:"type:varchar(64);uniqueIndex;not null"`
	TenantID  string `json:"tenant_id" gorm:"type:varchar(64);not null;index"`
	IsDefault bool   `json:"is_default" gorm:"default:false"`
}

// UserRole 用户角色关联
type UserRole struct {
	ID     uint `json:"id" gorm:"primaryKey"`
	UserID uint `json:"user_id" gorm:"not null;index"`
	RoleID uint `json:"role_id" gorm:"not null;index"`
}

// 表名设置
func (Document) TableName() string {
	return "documents"
}

func (DocumentVersion) TableName() string {
	return "document_versions"
}

func (DocumentPermission) TableName() string {
	return "document_permissions"
}

func (DocumentAudit) TableName() string {
	return "document_audits"
}

func (User) TableName() string {
	return "users"
}

func (Role) TableName() string {
	return "roles"
}

func (UserRole) TableName() string {
	return "user_roles"
}

// BeforeCreate 创建前钩子
func (d *Document) BeforeCreate(tx *gorm.DB) error {
	if d.UUID == "" {
		// 生成UUID
		d.UUID = generateUUID()
	}
	return nil
}

func (dv *DocumentVersion) BeforeCreate(tx *gorm.DB) error {
	if dv.UUID == "" {
		dv.UUID = generateUUID()
	}
	return nil
}

// 生成UUID的简单实现 (实际项目中应使用更可靠的UUID库)
func generateUUID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}