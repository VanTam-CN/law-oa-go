package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	Username  string         `json:"username" gorm:"size:50;not null;uniqueIndex"`
	Name      string         `json:"name" gorm:"column:name;size:50"`
	Email     string         `json:"email" gorm:"size:100;not null;uniqueIndex"`
	Password  string         `json:"-" gorm:"size:255;not null"`
	Role      string         `json:"role" gorm:"size:50;not null;default:'user'"`
	Phone     string         `json:"phone" gorm:"size:20"`
	Avatar    string         `json:"avatar" gorm:"size:255"`
	Status    string         `json:"status" gorm:"size:20;default:'active'"`
	Department string        `json:"department" gorm:"column:department;size:50;default:'综合部'"` // 部门
	Seniority  string         `json:"seniority" gorm:"column:seniority;size:20;default:'初级'"`    // 职级：初级/中级/高级/合伙人
}

// Client 客户模型
type Client struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	Name      string         `json:"name" gorm:"column:name;size:100;not null"`
	Type      string         `json:"type" gorm:"column:type;size:20;not null;default:'个人'"` // 客户类型：个人/企业
	Email     string         `json:"email" gorm:"size:100;uniqueIndex"`
	Phone     string         `json:"phone" gorm:"size:20"`
	Address   string         `json:"address" gorm:"type:text"`
	Company   string         `json:"company" gorm:"size:100"`
	IDCard    string         `json:"id_card" gorm:"column:id_card;size:18"`           // 身份证号（个人客户）
	Industry  string         `json:"industry" gorm:"column:industry;size:50"`        // 所属行业（企业客户）
	ContactPerson string      `json:"contact_person" gorm:"column:contact_person;size:50"` // 联系人（企业客户）
	ContactPhone string      `json:"contact_phone" gorm:"column:contact_phone;size:20"` // 联系电话（企业客户）
	Source    string         `json:"source" gorm:"column:source;size:50"`           // 客户来源
	Notes     string         `json:"notes" gorm:"column:notes;type:text"`
	Status    string         `json:"status" gorm:"size:20;default:'active'"`
}

// Case 案件模型
type Case struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	Title       string         `json:"title" gorm:"size:200;not null"`
	Description string         `json:"description" gorm:"type:text"`
	ClientID    uint           `json:"client_id" gorm:"not null;index"`
	Client      *Client        `json:"client,omitempty" gorm:"foreignKey:ClientID"`
	LawyerID    uint           `json:"lawyer_id" gorm:"not null;index"`
	Lawyer      *User          `json:"lawyer,omitempty" gorm:"foreignKey:LawyerID"`
	CaseType    string         `json:"case_type" gorm:"size:50;not null"`
	Priority    string         `json:"priority" gorm:"size:20;default:'medium'"`
	Status      string         `json:"status" gorm:"size:20;default:'pending'"`
	StartDate   *time.Time     `json:"start_date"`
	EndDate     *time.Time     `json:"end_date"`

	// 隔离墙相关字段 (v2.2.0)
	EthicalWallEnabled     bool        `json:"ethical_wall_enabled" gorm:"column:ethical_wall_enabled;default:false;index:idx_ethical_wall;comment:是否启用隔离墙"`
	EthicalWallDescription string     `json:"ethical_wall_description" gorm:"column:ethical_wall_description;type:text;comment:隔离墙说明"`
	EthicalWallEnabledBy   *uint       `json:"ethical_wall_enabled_by,omitempty" gorm:"column:ethical_wall_enabled_by;comment:启用人ID"`
	EthicalWallEnabledByUser *User     `json:"ethical_wall_enabled_by_user,omitempty" gorm:"foreignKey:EthicalWallEnabledBy"`
	EthicalWallEnabledAt   *time.Time  `json:"ethical_wall_enabled_at,omitempty" gorm:"column:ethical_wall_enabled_at;comment:启用时间"`

	// 临时字段，不映射到数据库，仅用于增强冲突检测
	OpposingParty string `json:"opposing_party" gorm:"-"`
	ClientName    string `json:"client_name" gorm:"-"`
}
