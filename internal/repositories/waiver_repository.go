package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"law-oa-go/internal/models"
)

// 豁免管理仓库实现
func (r *enhancedConflictRepository) CreateWaiverApplication(ctx context.Context, application *models.WaiverApplication) error {
	if err := r.db.WithContext(ctx).Create(application).Error; err != nil {
		return fmt.Errorf("创建豁免申请失败: %w", err)
	}
	return nil
}

func (r *enhancedConflictRepository) GetWaiverApplication(ctx context.Context, id string) (*models.WaiverApplication, error) {
	var application models.WaiverApplication
	if err := r.db.WithContext(ctx).
		Preload("ApprovalRecords").
		Preload("Signatures").
		Preload("MonitoringRecords").
		First(&application, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("豁免申请不存在: %s", id)
		}
		return nil, fmt.Errorf("获取豁免申请失败: %w", err)
	}
	return &application, nil
}

func (r *enhancedConflictRepository) GetWaiverApplicationByNumber(ctx context.Context, applicationNumber string) (*models.WaiverApplication, error) {
	var application models.WaiverApplication
	if err := r.db.WithContext(ctx).
		Preload("ApprovalRecords").
		Preload("Signatures").
		Preload("MonitoringRecords").
		First(&application, "application_number = ?", applicationNumber).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("豁免申请不存在: %s", applicationNumber)
		}
		return nil, fmt.Errorf("获取豁免申请失败: %w", err)
	}
	return &application, nil
}

func (r *enhancedConflictRepository) UpdateWaiverApplication(ctx context.Context, application *models.WaiverApplication) error {
	if err := r.db.WithContext(ctx).Save(application).Error; err != nil {
		return fmt.Errorf("更新豁免申请失败: %w", err)
	}
	return nil
}

func (r *enhancedConflictRepository) DeleteWaiverApplication(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.WaiverApplication{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("删除豁免申请失败: %w", err)
	}
	return nil
}

func (r *enhancedConflictRepository) GetWaiverApplicationsByStatus(ctx context.Context, status string) ([]*models.WaiverApplication, error) {
	var applications []*models.WaiverApplication
	if err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Find(&applications).Error; err != nil {
		return nil, fmt.Errorf("根据状态获取豁免申请失败: %w", err)
	}
	return applications, nil
}

func (r *enhancedConflictRepository) GetWaiverApplicationsByClient(ctx context.Context, clientID string) ([]*models.WaiverApplication, error) {
	var applications []*models.WaiverApplication
	if err := r.db.WithContext(ctx).
		Where("client_id = ?", clientID).
		Order("created_at DESC").
		Find(&applications).Error; err != nil {
		return nil, fmt.Errorf("根据客户获取豁免申请失败: %w", err)
	}
	return applications, nil
}

func (r *enhancedConflictRepository) GetWaiverApplicationsByLawyer(ctx context.Context, lawyerID string) ([]*models.WaiverApplication, error) {
	var applications []*models.WaiverApplication
	if err := r.db.WithContext(ctx).
		Where("lawyer_id = ?", lawyerID).
		Order("created_at DESC").
		Find(&applications).Error; err != nil {
		return nil, fmt.Errorf("根据律师获取豁免申请失败: %w", err)
	}
	return applications, nil
}

func (r *enhancedConflictRepository) GetWaiverApplicationsByConflictCheck(ctx context.Context, conflictCheckID string) ([]*models.WaiverApplication, error) {
	var applications []*models.WaiverApplication
	if err := r.db.WithContext(ctx).
		Where("conflict_check_id = ?", conflictCheckID).
		Order("created_at DESC").
		Find(&applications).Error; err != nil {
		return nil, fmt.Errorf("根据冲突检查获取豁免申请失败: %w", err)
	}
	return applications, nil
}

func (r *enhancedConflictRepository) GetPendingWaiverApplications(ctx context.Context) ([]*models.WaiverApplication, error) {
	var applications []*models.WaiverApplication
	if err := r.db.WithContext(ctx).
		Where("status IN ('SUBMITTED', 'UNDER_REVIEW')").
		Order("review_priority DESC, submission_date ASC").
		Find(&applications).Error; err != nil {
		return nil, fmt.Errorf("获取待处理豁免申请失败: %w", err)
	}
	return applications, nil
}

func (r *enhancedConflictRepository) GetExpiringWaiverApplications(ctx context.Context, days int) ([]*models.WaiverApplication, error) {
	var applications []*models.WaiverApplication
	expiryDate := time.Now().AddDate(0, 0, days)

	if err := r.db.WithContext(ctx).
		Where("status = 'APPROVED' AND requested_expiry_date <= ? AND requested_expiry_date >= ?",
			expiryDate, time.Now()).
		Order("requested_expiry_date ASC").
		Find(&applications).Error; err != nil {
		return nil, fmt.Errorf("获取即将到期豁免申请失败: %w", err)
	}
	return applications, nil
}

// 审批记录管理
func (r *enhancedConflictRepository) CreateWaiverApprovalRecord(ctx context.Context, record *models.WaiverApprovalRecord) error {
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return fmt.Errorf("创建豁免审批记录失败: %w", err)
	}
	return nil
}

func (r *enhancedConflictRepository) GetWaiverApprovalRecords(ctx context.Context, applicationID string) ([]*models.WaiverApprovalRecord, error) {
	var records []*models.WaiverApprovalRecord
	if err := r.db.WithContext(ctx).
		Where("waiver_application_id = ?", applicationID).
		Order("approval_date DESC").
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("获取豁免审批记录失败: %w", err)
	}
	return records, nil
}

func (r *enhancedConflictRepository) GetLatestWaiverApprovalRecord(ctx context.Context, applicationID string) (*models.WaiverApprovalRecord, error) {
	var record models.WaiverApprovalRecord
	if err := r.db.WithContext(ctx).
		Where("waiver_application_id = ?", applicationID).
		Order("approval_date DESC").
		First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("豁免审批记录不存在: %s", applicationID)
		}
		return nil, fmt.Errorf("获取最新豁免审批记录失败: %w", err)
	}
	return &record, nil
}

func (r *enhancedConflictRepository) UpdateWaiverApprovalRecord(ctx context.Context, record *models.WaiverApprovalRecord) error {
	if err := r.db.WithContext(ctx).Save(record).Error; err != nil {
		return fmt.Errorf("更新豁免审批记录失败: %w", err)
	}
	return nil
}

// 电子签名管理
func (r *enhancedConflictRepository) CreateWaiverSignature(ctx context.Context, signature *models.WaiverSignature) error {
	if err := r.db.WithContext(ctx).Create(signature).Error; err != nil {
		return fmt.Errorf("创建豁免签名失败: %w", err)
	}
	return nil
}

func (r *enhancedConflictRepository) GetWaiverSignatures(ctx context.Context, applicationID string) ([]*models.WaiverSignature, error) {
	var signatures []*models.WaiverSignature
	if err := r.db.WithContext(ctx).
		Where("waiver_application_id = ?", applicationID).
		Order("signature_timestamp ASC").
		Find(&signatures).Error; err != nil {
		return nil, fmt.Errorf("获取豁免签名失败: %w", err)
	}
	return signatures, nil
}

func (r *enhancedConflictRepository) GetWaiverSignature(ctx context.Context, id string) (*models.WaiverSignature, error) {
	var signature models.WaiverSignature
	if err := r.db.WithContext(ctx).
		First(&signature, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("豁免签名不存在: %s", id)
		}
		return nil, fmt.Errorf("获取豁免签名失败: %w", err)
	}
	return &signature, nil
}

func (r *enhancedConflictRepository) UpdateWaiverSignature(ctx context.Context, signature *models.WaiverSignature) error {
	if err := r.db.WithContext(ctx).Save(signature).Error; err != nil {
		return fmt.Errorf("更新豁免签名失败: %w", err)
	}
	return nil
}

func (r *enhancedConflictRepository) VerifyWaiverSignature(ctx context.Context, id string) (*models.WaiverSignature, error) {
	var signature models.WaiverSignature
	if err := r.db.WithContext(ctx).
		Where("id = ? AND verification_status = 'PENDING'", id).
		First(&signature).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("待验证的豁免签名不存在: %s", id)
		}
		return nil, fmt.Errorf("获取待验证豁免签名失败: %w", err)
	}

	// 更新验证状态
	signature.VerificationStatus = "VERIFIED"
	signature.VerifiedAt = &[]time.Time{time.Now()}[0]

	if err := r.db.WithContext(ctx).Save(&signature).Error; err != nil {
		return nil, fmt.Errorf("验证豁免签名失败: %w", err)
	}

	return &signature, nil
}

// 监控记录管理
func (r *enhancedConflictRepository) CreateWaiverMonitoringRecord(ctx context.Context, record *models.WaiverMonitoringRecord) error {
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return fmt.Errorf("创建豁免监控记录失败: %w", err)
	}
	return nil
}

func (r *enhancedConflictRepository) GetWaiverMonitoringRecords(ctx context.Context, applicationID string) ([]*models.WaiverMonitoringRecord, error) {
	var records []*models.WaiverMonitoringRecord
	if err := r.db.WithContext(ctx).
		Where("waiver_application_id = ?", applicationID).
		Order("monitoring_date DESC").
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("获取豁免监控记录失败: %w", err)
	}
	return records, nil
}

func (r *enhancedConflictRepository) GetLatestWaiverMonitoringRecord(ctx context.Context, applicationID string) (*models.WaiverMonitoringRecord, error) {
	var record models.WaiverMonitoringRecord
	if err := r.db.WithContext(ctx).
		Where("waiver_application_id = ?", applicationID).
		Order("monitoring_date DESC").
		First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("豁免监控记录不存在: %s", applicationID)
		}
		return nil, fmt.Errorf("获取最新豁免监控记录失败: %w", err)
	}
	return &record, nil
}

func (r *enhancedConflictRepository) UpdateWaiverMonitoringRecord(ctx context.Context, record *models.WaiverMonitoringRecord) error {
	if err := r.db.WithContext(ctx).Save(record).Error; err != nil {
		return fmt.Errorf("更新豁免监控记录失败: %w", err)
	}
	return nil
}

func (r *enhancedConflictRepository) GetOverdueMonitoringRecords(ctx context.Context) ([]*models.WaiverMonitoringRecord, error) {
	var records []*models.WaiverMonitoringRecord
	if err := r.db.WithContext(ctx).
		Where("status = 'SCHEDULED' AND monitoring_date < ?", time.Now()).
		Order("monitoring_date ASC").
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("获取逾期监控记录失败: %w", err)
	}
	return records, nil
}

// 豁免模板管理
func (r *enhancedConflictRepository) CreateWaiverTemplate(ctx context.Context, template *models.WaiverTemplate) error {
	if err := r.db.WithContext(ctx).Create(template).Error; err != nil {
		return fmt.Errorf("创建豁免模板失败: %w", err)
	}
	return nil
}

func (r *enhancedConflictRepository) GetWaiverTemplate(ctx context.Context, id string) (*models.WaiverTemplate, error) {
	var template models.WaiverTemplate
	if err := r.db.WithContext(ctx).
		First(&template, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("豁免模板不存在: %s", id)
		}
		return nil, fmt.Errorf("获取豁免模板失败: %w", err)
	}
	return &template, nil
}

func (r *enhancedConflictRepository) GetWaiverTemplateByCode(ctx context.Context, templateCode string) (*models.WaiverTemplate, error) {
	var template models.WaiverTemplate
	if err := r.db.WithContext(ctx).
		First(&template, "template_code = ?", templateCode).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("豁免模板不存在: %s", templateCode)
		}
		return nil, fmt.Errorf("获取豁免模板失败: %w", err)
	}
	return &template, nil
}

func (r *enhancedConflictRepository) UpdateWaiverTemplate(ctx context.Context, template *models.WaiverTemplate) error {
	if err := r.db.WithContext(ctx).Save(template).Error; err != nil {
		return fmt.Errorf("更新豁免模板失败: %w", err)
	}
	return nil
}

func (r *enhancedConflictRepository) DeleteWaiverTemplate(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.WaiverTemplate{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("删除豁免模板失败: %w", err)
	}
	return nil
}

func (r *enhancedConflictRepository) GetActiveWaiverTemplates(ctx context.Context) ([]*models.WaiverTemplate, error) {
	var templates []*models.WaiverTemplate
	if err := r.db.WithContext(ctx).
		Where("status = 'ACTIVE' AND (expiry_date IS NULL OR expiry_date > ?)", time.Now()).
		Order("template_type, template_name").
		Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("获取活跃豁免模板失败: %w", err)
	}
	return templates, nil
}

func (r *enhancedConflictRepository) GetWaiverTemplatesByType(ctx context.Context, templateType string) ([]*models.WaiverTemplate, error) {
	var templates []*models.WaiverTemplate
	if err := r.db.WithContext(ctx).
		Where("template_type = ? AND status = 'ACTIVE'", templateType).
		Order("template_name").
		Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("根据类型获取豁免模板失败: %w", err)
	}
	return templates, nil
}

func (r *enhancedConflictRepository) IncrementTemplateUsage(ctx context.Context, templateID string) error {
	if err := r.db.WithContext(ctx).
		Model(&models.WaiverTemplate{}).
		Where("id = ?", templateID).
		Updates(map[string]interface{}{
			"usage_count":    gorm.Expr("usage_count + 1"),
			"last_used_date": time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("增加模板使用次数失败: %w", err)
	}
	return nil
}