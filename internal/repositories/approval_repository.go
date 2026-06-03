package repositories

import (
	"fmt"

	"law-oa-go/internal/models"

	"gorm.io/gorm"
)

type ApprovalRepository struct {
	db *gorm.DB
}

func NewApprovalRepository(db *gorm.DB) *ApprovalRepository {
	return &ApprovalRepository{db: db}
}

// DB 返回底层数据库连接
func (r *ApprovalRepository) DB() *gorm.DB {
	return r.db
}

// FindByID 根据ID查找审批记录
func (r *ApprovalRepository) FindByID(id string) (*models.ApprovalRequest, error) {
	var approval models.ApprovalRequest
	err := r.db.Where("id = ?", id).First(&approval).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查找审批记录失败: %v", err)
	}
	return &approval, nil
}

// FindByUserID 根据用户ID查找审批记录
func (r *ApprovalRepository) FindByUserID(userID string, limit, offset int) ([]models.ApprovalRequest, error) {
	var approvals []models.ApprovalRequest
	err := r.db.Where("applicant_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&approvals).Error
	if err != nil {
		return nil, fmt.Errorf("查找用户审批记录失败: %v", err)
	}
	return approvals, nil
}

// FindPendingByApproverID 查找待审批人审批记录
func (r *ApprovalRepository) FindPendingByApproverID(approverID string, limit, offset int) ([]models.ApprovalRequest, error) {
	var approvals []models.ApprovalRequest
	err := r.db.Where("current_approver_id = ? AND status IN ?", approverID, []string{
		models.ApprovalStatusSubmitted,
		models.ApprovalStatusUnderReview,
	}).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&approvals).Error
	if err != nil {
		return nil, fmt.Errorf("查找待审批记录失败: %v", err)
	}
	return approvals, nil
}

// CountByUserID 统计用户审批记录数量
func (r *ApprovalRepository) CountByUserID(userID string, status string) (int64, error) {
	var count int64
	query := r.db.Model(&models.ApprovalRequest{}).Where("applicant_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("统计用户审批记录失败: %v", err)
	}
	return count, nil
}

// CountPendingByApproverID 统计待审批人待审批数量
func (r *ApprovalRepository) CountPendingByApproverID(approverID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.ApprovalRequest{}).
		Where("current_approver_id = ? AND status IN ?", approverID, []string{
			models.ApprovalStatusSubmitted,
			models.ApprovalStatusUnderReview,
		}).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("统计待审批记录失败: %v", err)
	}
	return count, nil
}

// CountPending 统计所有待审批数量
func (r *ApprovalRepository) CountPending() (int64, error) {
	var count int64
	err := r.db.Model(&models.ApprovalRequest{}).
		Where("status IN ?", []string{
			models.ApprovalStatusSubmitted,
			models.ApprovalStatusUnderReview,
		}).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("统计待审批记录失败: %v", err)
	}
	return count, nil
}

// Create 创建审批记录
func (r *ApprovalRepository) Create(approval *models.ApprovalRequest) error {
	if err := r.db.Create(approval).Error; err != nil {
		return fmt.Errorf("创建审批记录失败: %v", err)
	}
	return nil
}

// Update 更新审批记录
func (r *ApprovalRepository) Update(approval *models.ApprovalRequest) error {
	if err := r.db.Save(approval).Error; err != nil {
		return fmt.Errorf("更新审批记录失败: %v", err)
	}
	return nil
}

// Delete 删除审批记录
func (r *ApprovalRepository) Delete(id string) error {
	if err := r.db.Delete(&models.ApprovalRequest{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("删除审批记录失败: %v", err)
	}
	return nil
}

// FindWorkflows 查找工作流列表
func (r *ApprovalRepository) FindWorkflows() ([]models.ApprovalWorkflow, error) {
	var workflows []models.ApprovalWorkflow
	err := r.db.Where("status = ?", models.WorkflowStatusActive).
		Order("created_at DESC").
		Find(&workflows).Error
	if err != nil {
		return nil, fmt.Errorf("查找工作流失败: %v", err)
	}
	return workflows, nil
}

// FindTemplates 查找模板列表
func (r *ApprovalRepository) FindTemplates(templateType, category string) ([]models.ApprovalTemplateV2, error) {
	query := r.db.Where("status = ?", models.TemplateStatusActive)

	if templateType != "" {
		query = query.Where("template_type = ?", templateType)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}

	var templates []models.ApprovalTemplateV2
	err := query.Order("usage_count DESC, created_at DESC").
		Find(&templates).Error
	if err != nil {
		return nil, fmt.Errorf("查找模板失败: %v", err)
	}
	return templates, nil
}

// CreateApprovalRecord 创建审批记录
func (r *ApprovalRepository) CreateRecord(record *models.ApprovalRecord) error {
	if err := r.db.Create(record).Error; err != nil {
		return fmt.Errorf("创建审批记录失败: %v", err)
	}
	return nil
}

// FindRecordsByApprovalID 查找审批的所有记录
func (r *ApprovalRepository) FindRecordsByApprovalID(approvalRequestID string) ([]models.ApprovalRecord, error) {
	var records []models.ApprovalRecord
	err := r.db.Where("approval_request_id = ?", approvalRequestID).
		Order("created_at ASC").
		Find(&records).Error
	if err != nil {
		return nil, fmt.Errorf("查找审批记录失败: %v", err)
	}
	return records, nil
}