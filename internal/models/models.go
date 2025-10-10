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
}

// Client 客户模型
type Client struct {
	ID           uint           `json:"id" gorm:"primarykey"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
	ClientName   string         `json:"client_name" gorm:"column:client_name;size:100;not null"`
	Name         string         `json:"name" gorm:"column:client_name;size:100;not null"`
	Email        string         `json:"email" gorm:"size:100;uniqueIndex"`
	Phone        string         `json:"phone" gorm:"size:20"`
	Address      string         `json:"address" gorm:"type:text"`
	Company      string         `json:"company" gorm:"size:100"`
	Notes        string         `json:"notes" gorm:"column:remark;type:text"`
	Status       string         `json:"status" gorm:"size:20;default:'active'"`
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
}
