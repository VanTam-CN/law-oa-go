package models

import (
	"time"

	"gorm.io/gorm"
)

// Role 角色模型
type Role struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	Name        string         `json:"name" gorm:"size:50;not null;uniqueIndex"`
	Code        string         `json:"code" gorm:"size:50;not null;uniqueIndex"`
	Description string         `json:"description" gorm:"size:255"`
	Status      string         `json:"status" gorm:"size:20;default:'active'"`
	SortOrder   int            `json:"sort_order" gorm:"default:0"`
}

// Permission 权限模型
type Permission struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	Name      string         `json:"name" gorm:"size:100;not null"`
	Code      string         `json:"code" gorm:"size:100;not null;uniqueIndex"`
	Type      string         `json:"type" gorm:"size:20;not null;default:'menu'"` // menu, button, api
	ParentID  *uint          `json:"parent_id" gorm:"column:parent_id"`
	Path      string         `json:"path" gorm:"size:255"`
	Icon      string         `json:"icon" gorm:"size:100"`
	Component string         `json:"component" gorm:"size:255"`
	SortOrder int            `json:"sort_order" gorm:"default:0"`
	Status    string         `json:"status" gorm:"size:20;default:'active'"`

	// 关联
	Parent   *Permission  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []Permission `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}

// RolePermission 角色权限关联模型
type RolePermission struct {
	ID           uint      `json:"id" gorm:"primarykey"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	RoleID       uint      `json:"role_id" gorm:"not null;index"`
	PermissionID uint      `json:"permission_id" gorm:"not null;index"`

	// 关联
	Role       Role       `json:"role,omitempty" gorm:"foreignKey:RoleID"`
	Permission Permission `json:"permission,omitempty" gorm:"foreignKey:PermissionID"`
}

// UserRole 用户角色关联模型
type UserRole struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	RoleID    uint      `json:"role_id" gorm:"not null;index"`

	// 关联
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Role Role `json:"role,omitempty" gorm:"foreignKey:RoleID"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "roles"
}

func (Permission) TableName() string {
	return "permissions"
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

func (UserRole) TableName() string {
	return "user_roles"
}
