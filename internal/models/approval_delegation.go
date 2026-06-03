package models

import (
	"time"
)

// ApprovalDelegation 审批代理配置
// 当原审批人暂时无法处理审批时，可配置代理人代为审批
type ApprovalDelegation struct {
	ID          string     `json:"id" gorm:"column:id;type:varchar(36);primaryKey;default:(uuid_generate_v4())"`
	DelegatorID string     `json:"delegator_id" gorm:"column:delegator_id;type:varchar(36);not null;index:idx_delegation_delegator"`
	DelegateID  string     `json:"delegate_id" gorm:"column:delegate_id;type:varchar(36);not null;index:idx_delegation_delegate"`
	ValidFrom   time.Time  `json:"valid_from" gorm:"column:valid_from;type:timestamp;not null"`
	ValidUntil  *time.Time `json:"valid_until,omitempty" gorm:"column:valid_until;type:timestamp"`
	IsActive   bool       `json:"is_active" gorm:"column:is_active;default:true;not null;index:idx_delegation_active"`
	Reason     string     `json:"reason" gorm:"column:reason;type:text"`
	CreatedAt  time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	CreatedBy  string     `json:"created_by" gorm:"column:created_by;type:varchar(36);not null"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`

	// 关联数据（不持久化，仅查询时加载）
	Delegator *User `json:"delegator,omitempty" gorm:"-"`
	Delegate  *User `json:"delegate,omitempty" gorm:"-"`
}

func (ApprovalDelegation) TableName() string {
	return "approval_delegations"
}
