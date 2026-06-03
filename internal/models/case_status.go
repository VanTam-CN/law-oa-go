package models

import (
	"time"
)

// CaseStatus 案件状态常量
const (
	CaseStatusDraft      = "draft"       // 草拟
	CaseStatusRiskCheck  = "risk_check"  // 风控审查
	CaseStatusSigned     = "signed"      // 已签约
	CaseStatusInProgress = "in_progress" // 办案中
	CaseStatusTrial      = "trial"       // 庭审
	CaseStatusClosed     = "closed"      // 已结案
	CaseStatusArchived   = "archived"    // 已归档
)

// CaseStatusHistory 案件状态历史记录
type CaseStatusHistory struct {
	ID         uint      `json:"id" gorm:"primarykey"`
	CaseID     uint      `json:"case_id" gorm:"index;not null;comment:案件ID"`
	FromStatus string    `json:"from_status" gorm:"size:50;comment:变更前状态"`
	ToStatus   string    `json:"to_status" gorm:"size:50;not null;comment:变更后状态"`
	OperatorID uint      `json:"operator_id" gorm:"not null;comment:操作人ID"`
	Operator   *User     `json:"operator,omitempty" gorm:"foreignKey:OperatorID"`
	Reason     string    `json:"reason" gorm:"size:500;comment:变更原因"`
	CreatedAt  time.Time `json:"created_at" gorm:"comment:变更时间"`
}

// CaseStatusTransition 案件状态转换规则
type CaseStatusTransition struct {
	FromStatus    string `json:"from_status"`
	ToStatus      string `json:"to_status"`
	NeedsApproval bool   `json:"needs_approval"`
}

// TableName 设置表名
func (CaseStatusHistory) TableName() string {
	return "case_status_history"
}
