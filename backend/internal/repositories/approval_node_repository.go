package repositories

import (
	"fmt"

	"law-oa-go/internal/models"

	"gorm.io/gorm"
)

// ApprovalNodeRepository 审批节点仓储
type ApprovalNodeRepository struct {
	db *gorm.DB
}

// NewApprovalNodeRepository 创建审批节点仓储
func NewApprovalNodeRepository(db *gorm.DB) *ApprovalNodeRepository {
	return &ApprovalNodeRepository{db: db}
}

// Create 创建节点
func (r *ApprovalNodeRepository) Create(node *models.ApprovalNode) error {
	if err := r.db.Create(node).Error; err != nil {
		return fmt.Errorf("创建审批节点失败: %w", err)
	}
	return nil
}

// CreateBatch 批量创建节点
func (r *ApprovalNodeRepository) CreateBatch(nodes []models.ApprovalNode) error {
	if err := r.db.Create(&nodes).Error; err != nil {
		return fmt.Errorf("批量创建审批节点失败: %w", err)
	}
	return nil
}

// FindByID 根据ID查找节点
func (r *ApprovalNodeRepository) FindByID(id uint) (*models.ApprovalNode, error) {
	var node models.ApprovalNode
	err := r.db.Preload("Children").First(&node, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查找审批节点失败: %w", err)
	}
	return &node, nil
}

// FindByApprovalID 查找审批的所有节点
func (r *ApprovalNodeRepository) FindByApprovalID(approvalID string) ([]models.ApprovalNode, error) {
	var nodes []models.ApprovalNode
	err := r.db.Where("approval_id = ?", approvalID).
		Order("step_order ASC, id ASC").
		Find(&nodes).Error
	if err != nil {
		return nil, fmt.Errorf("查找审批节点失败: %w", err)
	}
	return nodes, nil
}

// FindPendingByApprovalID 查找审批的待处理节点
func (r *ApprovalNodeRepository) FindPendingByApprovalID(approvalID string) ([]models.ApprovalNode, error) {
	var nodes []models.ApprovalNode
	err := r.db.Where("approval_id = ? AND action = ?", approvalID, models.NodeActionPending).
		Order("step_order ASC").
		Find(&nodes).Error
	if err != nil {
		return nil, fmt.Errorf("查找待处理节点失败: %w", err)
	}
	return nodes, nil
}

// FindPendingByApproverID 查找审批人的待处理节点
func (r *ApprovalNodeRepository) FindPendingByApproverID(approverID string) ([]models.ApprovalNode, error) {
	var nodes []models.ApprovalNode
	err := r.db.Joins("JOIN approval_requests ON approval_nodes.approval_id = approval_requests.id").
		Where("approval_nodes.approver_id = ? AND approval_nodes.action = ?", approverID, models.NodeActionPending).
		Order("approval_nodes.created_at ASC").
		Find(&nodes).Error
	if err != nil {
		return nil, fmt.Errorf("查找待处理节点失败: %w", err)
	}
	return nodes, nil
}

// Update 更新节点
func (r *ApprovalNodeRepository) Update(node *models.ApprovalNode) error {
	if err := r.db.Save(node).Error; err != nil {
		return fmt.Errorf("更新审批节点失败: %w", err)
	}
	return nil
}

// Delete 删除节点
func (r *ApprovalNodeRepository) Delete(id uint) error {
	if err := r.db.Delete(&models.ApprovalNode{}, id).Error; err != nil {
		return fmt.Errorf("删除审批节点失败: %w", err)
	}
	return nil
}

// DeleteByApprovalID 删除审批的所有节点
func (r *ApprovalNodeRepository) DeleteByApprovalID(approvalID string) error {
	if err := r.db.Where("approval_id = ?", approvalID).Delete(&models.ApprovalNode{}).Error; err != nil {
		return fmt.Errorf("删除审批节点失败: %w", err)
	}
	return nil
}

// CountPendingByApprovalID 统计审批的待处理节点数量
func (r *ApprovalNodeRepository) CountPendingByApprovalID(approvalID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.ApprovalNode{}).
		Where("approval_id = ? AND action = ?", approvalID, models.NodeActionPending).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("统计待处理节点失败: %w", err)
	}
	return count, nil
}

// CountApprovedByApprovalID 统计审批的已通过节点数量
func (r *ApprovalNodeRepository) CountApprovedByApprovalID(approvalID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.ApprovalNode{}).
		Where("approval_id = ? AND action = ?", approvalID, models.NodeActionApproved).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("统计已通过节点失败: %w", err)
	}
	return count, nil
}

// GetStepNodes 获取指定步骤的所有节点
func (r *ApprovalNodeRepository) GetStepNodes(approvalID string, stepOrder int) ([]models.ApprovalNode, error) {
	var nodes []models.ApprovalNode
	err := r.db.Where("approval_id = ? AND step_order = ?", approvalID, stepOrder).
		Order("id ASC").
		Find(&nodes).Error
	if err != nil {
		return nil, fmt.Errorf("查找步骤节点失败: %w", err)
	}
	return nodes, nil
}
