package repositories

import (
	"fmt"

	"law-oa-go/internal/models"

	"gorm.io/gorm"
)

// ApprovalWorkflowRepository 审批工作流仓储
type ApprovalWorkflowRepository struct {
	db *gorm.DB
}

// NewApprovalWorkflowRepository 创建工作流仓储
func NewApprovalWorkflowRepository(db *gorm.DB) *ApprovalWorkflowRepository {
	return &ApprovalWorkflowRepository{db: db}
}

// FindByType 根据类型查找工作流
func (r *ApprovalWorkflowRepository) FindByType(workflowType string) (*models.ApprovalWorkflow, error) {
	var workflow models.ApprovalWorkflow
	err := r.db.Where("workflow_type = ? AND status = ?", workflowType, models.WorkflowStatusActive).
		First(&workflow).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查找工作流失败: %v", err)
	}
	return &workflow, nil
}

// FindByID 根据ID查找工作流
func (r *ApprovalWorkflowRepository) FindByID(id string) (*models.ApprovalWorkflow, error) {
	var workflow models.ApprovalWorkflow
	err := r.db.Where("id = ?", id).First(&workflow).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查找工作流失败: %v", err)
	}
	return &workflow, nil
}

// FindAll 查找所有活跃工作流
func (r *ApprovalWorkflowRepository) FindAll() ([]models.ApprovalWorkflow, error) {
	var workflows []models.ApprovalWorkflow
	err := r.db.Where("status = ?", models.WorkflowStatusActive).
		Order("created_at DESC").
		Find(&workflows).Error
	if err != nil {
		return nil, fmt.Errorf("查找工作流列表失败: %v", err)
	}
	return workflows, nil
}

// Create 创建工作流
func (r *ApprovalWorkflowRepository) Create(workflow *models.ApprovalWorkflow) error {
	if err := r.db.Create(workflow).Error; err != nil {
		return fmt.Errorf("创建工作流失败: %v", err)
	}
	return nil
}

// Update 更新工作流
func (r *ApprovalWorkflowRepository) Update(workflow *models.ApprovalWorkflow) error {
	if err := r.db.Save(workflow).Error; err != nil {
		return fmt.Errorf("更新工作流失败: %v", err)
	}
	return nil
}

// Delete 删除工作流
func (r *ApprovalWorkflowRepository) Delete(id string) error {
	if err := r.db.Delete(&models.ApprovalWorkflow{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("删除工作流失败: %v", err)
	}
	return nil
}
