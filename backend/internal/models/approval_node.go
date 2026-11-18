package models

import (
	"time"

	"gorm.io/gorm"
)

// ApprovalNode 审批节点
// 记录审批流程中每个节点的状态，支持会签、或签、转签等复杂场景
type ApprovalNode struct {
	ID           uint            `json:"id" gorm:"primarykey"`
	ApprovalID   string          `json:"approval_id" gorm:"not null;size:36;index:idx_approval_node"`
	StepOrder    int             `json:"step_order" gorm:"not null"`          // 步骤顺序
	NodeName     string          `json:"node_name" gorm:"size:100"`          // 节点名称
	ApproverType string          `json:"approver_type" gorm:"size:20"`       // ROLE, SPECIFIC_USER
	ApproverID   *string         `json:"approver_id"`                        // 具体审批人ID
	ApproverRole string          `json:"approver_role" gorm:"size:50"`       // 角色名
	Action       string          `json:"action" gorm:"size:20;default:'PENDING'"` // PENDING, APPROVED, REJECTED, TRANSFERRED, ADDED_SIGN
	Comment      string          `json:"comment" gorm:"type:text"`            // 审批意见
	SignImage    string          `json:"sign_image" gorm:"type:text"`         // 电子签章base64
	Duration     int             `json:"duration"`                            // 处理时长(秒)
	IsFinal      bool            `json:"is_final" gorm:"default:false"`       // 是否最终节点
	ParentNodeID *uint           `json:"parent_node_id"`                     // 父节点ID(用于会签、加签)
	SignType     string          `json:"sign_type" gorm:"size:20;default:'SINGLE'"` // SINGLE, COUNTERSIGN, OR_SIGN
	CreatedAt    time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt  `json:"-" gorm:"index"`

	// 关联
	Approval     *ApprovalRequest `json:"approval,omitempty" gorm:"foreignKey:ApprovalID"`
	Children     []ApprovalNode   `json:"children,omitempty" gorm:"foreignKey:ParentNodeID"`
}

// TableName 设置表名
func (ApprovalNode) TableName() string {
	return "approval_nodes"
}

// 审批节点动作常量
const (
	NodeActionPending     = "PENDING"      // 待处理
	NodeActionApproved    = "APPROVED"     // 已通过
	NodeActionRejected    = "REJECTED"     // 已拒绝
	NodeActionTransferred = "TRANSFERRED"  // 已转签
	NodeActionAddedSign   = "ADDED_SIGN"   // 已加签
)

// IsPending 检查节点是否待处理
func (n *ApprovalNode) IsPending() bool {
	return n.Action == NodeActionPending
}

// IsApproved 检查节点是否已通过
func (n *ApprovalNode) IsApproved() bool {
	return n.Action == NodeActionApproved
}

// IsRejected 检查节点是否已拒绝
func (n *ApprovalNode) IsRejected() bool {
	return n.Action == NodeActionRejected
}

// CanProcess 检查节点是否可处理
func (n *ApprovalNode) CanProcess() bool {
	return n.Action == NodeActionPending
}

// MarkAsApproved 标记为已通过
func (n *ApprovalNode) MarkAsApproved(comment string) {
	n.Action = NodeActionApproved
	n.Comment = comment
}

// MarkAsRejected 标记为已拒绝
func (n *ApprovalNode) MarkAsRejected(comment string) {
	n.Action = NodeActionRejected
	n.Comment = comment
}

// MarkAsTransferred 标记为已转签
func (n *ApprovalNode) MarkAsTransferred(comment string) {
	n.Action = NodeActionTransferred
	n.Comment = comment
}
