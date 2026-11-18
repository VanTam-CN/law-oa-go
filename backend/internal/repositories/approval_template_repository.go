package repositories

import (
	"encoding/json"
	"fmt"

	"law-oa-go/internal/models"

	"gorm.io/gorm"
)

// ApprovalTemplateRepository 审批模板仓储
type ApprovalTemplateRepository struct {
	db *gorm.DB
}

// NewApprovalTemplateRepository 创建审批模板仓储
func NewApprovalTemplateRepository(db *gorm.DB) *ApprovalTemplateRepository {
	return &ApprovalTemplateRepository{db: db}
}

// FindByName 根据名称查找模板
func (r *ApprovalTemplateRepository) FindByName(name string) (*models.ApprovalTemplate, error) {
	var template models.ApprovalTemplate
	err := r.db.Where("name = ? AND is_active = ?", name, true).First(&template).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查找审批模板失败: %w", err)
	}
	return &template, nil
}

// FindAll 查找所有启用的模板
func (r *ApprovalTemplateRepository) FindAll() ([]models.ApprovalTemplate, error) {
	var templates []models.ApprovalTemplate
	err := r.db.Where("is_active = ?", true).Order("id").Find(&templates).Error
	if err != nil {
		return nil, fmt.Errorf("查找审批模板列表失败: %w", err)
	}
	return templates, nil
}

// Create 创建模板
func (r *ApprovalTemplateRepository) Create(template *models.ApprovalTemplate) error {
	if err := r.db.Create(template).Error; err != nil {
		return fmt.Errorf("创建审批模板失败: %w", err)
	}
	return nil
}

// Update 更新模板
func (r *ApprovalTemplateRepository) Update(template *models.ApprovalTemplate) error {
	if err := r.db.Save(template).Error; err != nil {
		return fmt.Errorf("更新审批模板失败: %w", err)
	}
	return nil
}

// Delete 删除模板(软删除)
func (r *ApprovalTemplateRepository) Delete(id uint) error {
	if err := r.db.Delete(&models.ApprovalTemplate{}, id).Error; err != nil {
		return fmt.Errorf("删除审批模板失败: %w", err)
	}
	return nil
}

// GetSteps 获取模板的审批步骤
func (r *ApprovalTemplateRepository) GetSteps(template *models.ApprovalTemplate) ([]models.ApprovalStep, error) {
	var steps []models.ApprovalStep
	if err := json.Unmarshal([]byte(template.Steps), &steps); err != nil {
		return nil, fmt.Errorf("解析审批步骤失败: %w", err)
	}
	return steps, nil
}

// GetConditions 获取模板的条件配置
func (r *ApprovalTemplateRepository) GetConditions(template *models.ApprovalTemplate) ([]models.ApprovalCondition, error) {
	if template.Conditions == "" || template.Conditions == "null" {
		return []models.ApprovalCondition{}, nil
	}
	var conditions []models.ApprovalCondition
	if err := json.Unmarshal([]byte(template.Conditions), &conditions); err != nil {
		return nil, fmt.Errorf("解析审批条件失败: %w", err)
	}
	return conditions, nil
}
