package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"law-oa-go/internal/models"
)

// 增强冲突检测仓库接口
type EnhancedConflictRepository interface {
	// 客户档案管理
	ClientProfileRepository

	// 冲突分类标准
	ConflictClassificationRepository

	// 豁免管理
	WaiverManagementRepository

	// 专业冲突检测
	ProfessionalConflictDetectionRepository

	// 统计和报告
	StatisticsRepository
}

// 客户档案仓库接口
type ClientProfileRepository interface {
	// 基础CRUD操作
	CreateClientProfile(ctx context.Context, profile *models.ClientProfile) error
	GetClientProfile(ctx context.Context, id string) (*models.ClientProfile, error)
	UpdateClientProfile(ctx context.Context, profile *models.ClientProfile) error
	DeleteClientProfile(ctx context.Context, id string) error

	// 查询操作
	GetClientProfileByNumber(ctx context.Context, clientNumber string) (*models.ClientProfile, error)
	GetClientProfilesByType(ctx context.Context, clientType string) ([]*models.ClientProfile, error)
	GetClientProfilesByStatus(ctx context.Context, status string) ([]*models.ClientProfile, error)
	GetClientsByPartner(ctx context.Context, partnerID string) ([]*models.ClientProfile, error)
	GetClientsByAttorney(ctx context.Context, attorneyID string) ([]*models.ClientProfile, error)

	// 高级查询
	SearchClients(ctx context.Context, query *ClientSearchQuery) (*ClientSearchResult, error)
	GetClientsWithRelationships(ctx context.Context, clientID string) (*models.ClientProfile, error)
	GetClientsWithRiskProfile(ctx context.Context, riskLevel string) ([]*models.ClientProfile, error)

	// 客户关系管理
	CreateClientRelationship(ctx context.Context, relationship *models.ClientRelationship) error
	GetClientRelationships(ctx context.Context, clientID string) ([]*models.ClientRelationship, error)
	UpdateClientRelationship(ctx context.Context, relationship *models.ClientRelationship) error
	DeleteClientRelationship(ctx context.Context, id string) error

	// 名称变体管理
	CreateClientNameVariant(ctx context.Context, variant *models.ClientNameVariant) error
	GetClientNameVariants(ctx context.Context, clientID string) ([]*models.ClientNameVariant, error)
	UpdateClientNameVariant(ctx context.Context, variant *models.ClientNameVariant) error
	DeleteClientNameVariant(ctx context.Context, id string) error

	// 行业分类管理
	CreateIndustryClassification(ctx context.Context, classification *models.IndustryClassification) error
	GetIndustryClassifications(ctx context.Context, clientID string) ([]*models.IndustryClassification, error)
	UpdateIndustryClassification(ctx context.Context, classification *models.IndustryClassification) error
	DeleteIndustryClassification(ctx context.Context, id string) error

	// 风险档案管理
	CreateClientRiskProfile(ctx context.Context, riskProfile *models.ClientRiskProfile) error
	GetClientRiskProfile(ctx context.Context, clientID string) (*models.ClientRiskProfile, error)
	UpdateClientRiskProfile(ctx context.Context, riskProfile *models.ClientRiskProfile) error
	GetClientsRequiringMonitoring(ctx context.Context) ([]*models.ClientProfile, error)
}

// 冲突分类标准仓库接口
type ConflictClassificationRepository interface {
	// 冲突类型管理
	CreateConflictType(ctx context.Context, conflictType *models.ConflictType) error
	GetConflictType(ctx context.Context, id string) (*models.ConflictType, error)
	GetConflictTypeByCode(ctx context.Context, code string) (*models.ConflictType, error)
	UpdateConflictType(ctx context.Context, conflictType *models.ConflictType) error
	DeleteConflictType(ctx context.Context, id string) error
	GetAllConflictTypes(ctx context.Context) ([]*models.ConflictType, error)
	GetActiveConflictTypes(ctx context.Context) ([]*models.ConflictType, error)

	// 分类标准引用管理
	CreateClassificationReference(ctx context.Context, ref *models.ClassificationReference) error
	GetClassificationReferences(ctx context.Context, conflictTypeID string) ([]*models.ClassificationReference, error)
	UpdateClassificationReference(ctx context.Context, ref *models.ClassificationReference) error
	DeleteClassificationReference(ctx context.Context, id string) error

	// 风险等级定义管理
	CreateRiskLevelDefinition(ctx context.Context, definition *models.RiskLevelDefinition) error
	GetRiskLevelDefinition(ctx context.Context, level string) (*models.RiskLevelDefinition, error)
	UpdateRiskLevelDefinition(ctx context.Context, definition *models.RiskLevelDefinition) error
	GetAllRiskLevelDefinitions(ctx context.Context) ([]*models.RiskLevelDefinition, error)

	// 查询方法
	SearchConflictTypes(ctx context.Context, query *ConflictTypeSearchQuery) (*ConflictTypeSearchResult, error)
	GetConflictTypesByStandard(ctx context.Context, standardType string) ([]*models.ConflictType, error)
	GetConflictTypesByRiskLevel(ctx context.Context, riskLevel string) ([]*models.ConflictType, error)
}

// 豁免管理仓库接口
type WaiverManagementRepository interface {
	// 豁免申请管理
	CreateWaiverApplication(ctx context.Context, application *models.WaiverApplication) error
	GetWaiverApplication(ctx context.Context, id string) (*models.WaiverApplication, error)
	GetWaiverApplicationByNumber(ctx context.Context, applicationNumber string) (*models.WaiverApplication, error)
	UpdateWaiverApplication(ctx context.Context, application *models.WaiverApplication) error
	DeleteWaiverApplication(ctx context.Context, id string) error

	// 申请查询
	GetWaiverApplicationsByStatus(ctx context.Context, status string) ([]*models.WaiverApplication, error)
	GetWaiverApplicationsByClient(ctx context.Context, clientID string) ([]*models.WaiverApplication, error)
	GetWaiverApplicationsByLawyer(ctx context.Context, lawyerID string) ([]*models.WaiverApplication, error)
	GetWaiverApplicationsByConflictCheck(ctx context.Context, conflictCheckID string) ([]*models.WaiverApplication, error)
	GetPendingWaiverApplications(ctx context.Context) ([]*models.WaiverApplication, error)
	GetExpiringWaiverApplications(ctx context.Context, days int) ([]*models.WaiverApplication, error)

	// 审批记录管理
	CreateWaiverApprovalRecord(ctx context.Context, record *models.WaiverApprovalRecord) error
	GetWaiverApprovalRecords(ctx context.Context, applicationID string) ([]*models.WaiverApprovalRecord, error)
	GetLatestWaiverApprovalRecord(ctx context.Context, applicationID string) (*models.WaiverApprovalRecord, error)
	UpdateWaiverApprovalRecord(ctx context.Context, record *models.WaiverApprovalRecord) error

	// 电子签名管理
	CreateWaiverSignature(ctx context.Context, signature *models.WaiverSignature) error
	GetWaiverSignatures(ctx context.Context, applicationID string) ([]*models.WaiverSignature, error)
	GetWaiverSignature(ctx context.Context, id string) (*models.WaiverSignature, error)
	UpdateWaiverSignature(ctx context.Context, signature *models.WaiverSignature) error
	VerifyWaiverSignature(ctx context.Context, id string) (*models.WaiverSignature, error)

	// 监控记录管理
	CreateWaiverMonitoringRecord(ctx context.Context, record *models.WaiverMonitoringRecord) error
	GetWaiverMonitoringRecords(ctx context.Context, applicationID string) ([]*models.WaiverMonitoringRecord, error)
	GetLatestWaiverMonitoringRecord(ctx context.Context, applicationID string) (*models.WaiverMonitoringRecord, error)
	UpdateWaiverMonitoringRecord(ctx context.Context, record *models.WaiverMonitoringRecord) error
	GetOverdueMonitoringRecords(ctx context.Context) ([]*models.WaiverMonitoringRecord, error)

	// 豁免模板管理
	CreateWaiverTemplate(ctx context.Context, template *models.WaiverTemplate) error
	GetWaiverTemplate(ctx context.Context, id string) (*models.WaiverTemplate, error)
	GetWaiverTemplateByCode(ctx context.Context, templateCode string) (*models.WaiverTemplate, error)
	UpdateWaiverTemplate(ctx context.Context, template *models.WaiverTemplate) error
	DeleteWaiverTemplate(ctx context.Context, id string) error
	GetActiveWaiverTemplates(ctx context.Context) ([]*models.WaiverTemplate, error)
	GetWaiverTemplatesByType(ctx context.Context, templateType string) ([]*models.WaiverTemplate, error)
	IncrementTemplateUsage(ctx context.Context, templateID string) error
}

// 专业冲突检测仓库接口
type ProfessionalConflictDetectionRepository interface {
	// 冲突检查请求管理
	CreateConflictCheckRequest(ctx context.Context, request *models.ProfessionalConflictCheckRequest) error
	GetConflictCheckRequest(ctx context.Context, id string) (*models.ProfessionalConflictCheckRequest, error)
	GetConflictCheckRequestByNumber(ctx context.Context, checkNumber string) (*models.ProfessionalConflictCheckRequest, error)
	UpdateConflictCheckRequest(ctx context.Context, request *models.ProfessionalConflictCheckRequest) error
	DeleteConflictCheckRequest(ctx context.Context, id string) error

	// 请求查询
	GetConflictCheckRequestsByStatus(ctx context.Context, status string) ([]*models.ProfessionalConflictCheckRequest, error)
	GetConflictCheckRequestsByPriority(ctx context.Context, priority string) ([]*models.ProfessionalConflictCheckRequest, error)
	GetConflictCheckRequestsByAssignee(ctx context.Context, assigneeID string) ([]*models.ProfessionalConflictCheckRequest, error)
	GetConflictCheckRequestsByDateRange(ctx context.Context, start, end time.Time) ([]*models.ProfessionalConflictCheckRequest, error)
	GetOverdueConflictCheckRequests(ctx context.Context) ([]*models.ProfessionalConflictCheckRequest, error)

	// 多维度冲突结果管理
	CreateMultidimensionalConflictResult(ctx context.Context, result *models.MultidimensionalConflictResult) error
	GetMultidimensionalConflictResults(ctx context.Context, checkRequestID string) ([]*models.MultidimensionalConflictResult, error)
	GetMultidimensionalConflictResult(ctx context.Context, id string) (*models.MultidimensionalConflictResult, error)
	UpdateMultidimensionalConflictResult(ctx context.Context, result *models.MultidimensionalConflictResult) error
	DeleteMultidimensionalConflictResult(ctx context.Context, id string) error

	// 冲突结果查询
	GetConflictsByType(ctx context.Context, conflictType string) ([]*models.MultidimensionalConflictResult, error)
	GetConflictsBySeverity(ctx context.Context, severity string) ([]*models.MultidimensionalConflictResult, error)
	GetConflictsByStatus(ctx context.Context, status string) ([]*models.MultidimensionalConflictResult, error)
	GetConflictsByEntity(ctx context.Context, entityType, entityID string) ([]*models.MultidimensionalConflictResult, error)
	GetConflictsRequiringMonitoring(ctx context.Context) ([]*models.MultidimensionalConflictResult, error)

	// 冲突检测规则管理
	CreateConflictDetectionRule(ctx context.Context, rule *models.ConflictDetectionRule) error
	GetConflictDetectionRule(ctx context.Context, id string) (*models.ConflictDetectionRule, error)
	GetConflictDetectionRuleByCode(ctx context.Context, ruleCode string) (*models.ConflictDetectionRule, error)
	UpdateConflictDetectionRule(ctx context.Context, rule *models.ConflictDetectionRule) error
	DeleteConflictDetectionRule(ctx context.Context, id string) error
	GetActiveConflictDetectionRules(ctx context.Context) ([]*models.ConflictDetectionRule, error)
	GetConflictDetectionRulesByType(ctx context.Context, ruleType string) ([]*models.ConflictDetectionRule, error)

	// 规则执行记录管理
	CreateConflictRuleExecution(ctx context.Context, execution *models.ConflictRuleExecution) error
	GetConflictRuleExecutions(ctx context.Context, checkRequestID string) ([]*models.ConflictRuleExecution, error)
	GetConflictRuleExecutionsByRule(ctx context.Context, ruleID string) ([]*models.ConflictRuleExecution, error)
	UpdateConflictRuleExecution(ctx context.Context, execution *models.ConflictRuleExecution) error
	GetRuleExecutionStatistics(ctx context.Context, ruleID string) (*RuleExecutionStats, error)
}

// 统计仓库接口
type StatisticsRepository interface {
	// 冲突检查统计
	GetConflictCheckStatistics(ctx context.Context, query *StatisticsQuery) (*models.ProfessionalConflictCheckStats, error)
	GetConflictTrends(ctx context.Context, period string) ([]*ConflictTrend, error)
	GetRiskDistribution(ctx context.Context, filters *RiskDistributionFilters) (*RiskDistribution, error)

	// 豁免管理统计
	GetWaiverStatistics(ctx context.Context, query *StatisticsQuery) (*WaiverStatistics, error)
	GetWaiverApprovalTrends(ctx context.Context, period string) ([]*WaiverApprovalTrend, error)

	// 客户风险统计
	GetClientRiskStatistics(ctx context.Context, query *StatisticsQuery) (*ClientRiskStatistics, error)
	GetHighRiskClients(ctx context.Context, limit int) ([]*HighRiskClient, error)

	// 效率统计
	GetProcessingEfficiencyStats(ctx context.Context, query *StatisticsQuery) (*ProcessingEfficiencyStats, error)
	GetSlaComplianceStats(ctx context.Context, query *StatisticsQuery) (*SlaComplianceStats, error)
}

// 增强冲突检测仓库实现
type enhancedConflictRepository struct {
	db *gorm.DB
}

// NewEnhancedConflictRepository 创建新的增强冲突检测仓库实例
func NewEnhancedConflictRepository(db *gorm.DB) EnhancedConflictRepository {
	return &enhancedConflictRepository{db: db}
}

// ==================== 客户档案管理实现 ====================

// CreateClientProfile 创建客户档案
func (r *enhancedConflictRepository) CreateClientProfile(ctx context.Context, profile *models.ClientProfile) error {
	if err := r.db.WithContext(ctx).Create(profile).Error; err != nil {
		return fmt.Errorf("创建客户档案失败: %w", err)
	}
	return nil
}

// GetClientProfile 获取客户档案
func (r *enhancedConflictRepository) GetClientProfile(ctx context.Context, id string) (*models.ClientProfile, error) {
	var profile models.ClientProfile
	if err := r.db.WithContext(ctx).
		Preload("RelatedClients.RelatedClient").
		Preload("NameVariants").
		Preload("IndustryClassifications").
		Preload("RiskProfile").
		First(&profile, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("客户档案不存在: %s", id)
		}
		return nil, fmt.Errorf("获取客户档案失败: %w", err)
	}
	return &profile, nil
}

// UpdateClientProfile 更新客户档案
func (r *enhancedConflictRepository) UpdateClientProfile(ctx context.Context, profile *models.ClientProfile) error {
	if err := r.db.WithContext(ctx).Save(profile).Error; err != nil {
		return fmt.Errorf("更新客户档案失败: %w", err)
	}
	return nil
}

// DeleteClientProfile 删除客户档案
func (r *enhancedConflictRepository) DeleteClientProfile(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.ClientProfile{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("删除客户档案失败: %w", err)
	}
	return nil
}

// GetClientProfileByNumber 根据客户编号获取客户档案
func (r *enhancedConflictRepository) GetClientProfileByNumber(ctx context.Context, clientNumber string) (*models.ClientProfile, error) {
	var profile models.ClientProfile
	if err := r.db.WithContext(ctx).
		Preload("RelatedClients.RelatedClient").
		Preload("NameVariants").
		Preload("IndustryClassifications").
		Preload("RiskProfile").
		First(&profile, "client_number = ?", clientNumber).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("客户档案不存在: %s", clientNumber)
		}
		return nil, fmt.Errorf("获取客户档案失败: %w", err)
	}
	return &profile, nil
}

// GetClientProfilesByType 根据客户类型获取客户档案列表
func (r *enhancedConflictRepository) GetClientProfilesByType(ctx context.Context, clientType string) ([]*models.ClientProfile, error) {
	var profiles []*models.ClientProfile
	if err := r.db.WithContext(ctx).
		Where("client_type = ?", clientType).
		Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("根据类型获取客户档案失败: %w", err)
	}
	return profiles, nil
}

// GetClientProfilesByStatus 根据客户状态获取客户档案列表
func (r *enhancedConflictRepository) GetClientProfilesByStatus(ctx context.Context, status string) ([]*models.ClientProfile, error) {
	var profiles []*models.ClientProfile
	if err := r.db.WithContext(ctx).
		Where("client_status = ?", status).
		Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("根据状态获取客户档案失败: %w", err)
	}
	return profiles, nil
}

// GetClientsByPartner 根据主管合伙人获取客户列表
func (r *enhancedConflictRepository) GetClientsByPartner(ctx context.Context, partnerID string) ([]*models.ClientProfile, error) {
	var profiles []*models.ClientProfile
	if err := r.db.WithContext(ctx).
		Where("assigned_partner = ?", partnerID).
		Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("根据主管合伙人获取客户列表失败: %w", err)
	}
	return profiles, nil
}

// GetClientsByAttorney 根据主管律师获取客户列表
func (r *enhancedConflictRepository) GetClientsByAttorney(ctx context.Context, attorneyID string) ([]*models.ClientProfile, error) {
	var profiles []*models.ClientProfile
	if err := r.db.WithContext(ctx).
		Where("assigned_attorney = ?", attorneyID).
		Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("根据主管律师获取客户列表失败: %w", err)
	}
	return profiles, nil
}

// SearchClients 搜索客户
func (r *enhancedConflictRepository) SearchClients(ctx context.Context, query *ClientSearchQuery) (*ClientSearchResult, error) {
	var profiles []*models.ClientProfile
	var total int64

	db := r.db.WithContext(ctx)

	// 应用搜索条件
	if query.Keyword != "" {
		keyword := "%" + query.Keyword + "%"
		db = db.Where("client_number LIKE ? OR client_number LIKE ?", keyword, keyword)
	}

	if query.ClientType != "" {
		db = db.Where("client_type = ?", query.ClientType)
	}

	if query.ClientStatus != "" {
		db = db.Where("client_status = ?", query.ClientStatus)
	}

	if query.AssignedPartner != "" {
		db = db.Where("assigned_partner = ?", query.AssignedPartner)
	}

	if query.AssignedAttorney != "" {
		db = db.Where("assigned_attorney = ?", query.AssignedAttorney)
	}

	// 获取总数
	if err := db.Model(&models.ClientProfile{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("获取客户总数失败: %w", err)
	}

	// 应用分页和排序
	if query.PageSize > 0 && query.Page > 0 {
		offset := (query.Page - 1) * query.PageSize
		db = db.Offset(offset).Limit(query.PageSize)
	}

	if query.OrderBy != "" {
		db = db.Order(query.OrderBy)
	} else {
		db = db.Order("created_at DESC")
	}

	// 执行查询
	if err := db.Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("搜索客户失败: %w", err)
	}

	return &ClientSearchResult{
		Profiles: profiles,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

// GetClientsWithRelationships 获取包含关系网络的客户档案
func (r *enhancedConflictRepository) GetClientsWithRelationships(ctx context.Context, clientID string) (*models.ClientProfile, error) {
	var profile models.ClientProfile
	if err := r.db.WithContext(ctx).
		Preload("RelatedClients.RelatedClient").
		Preload("RelatedClients.Client").
		Preload("NameVariants").
		Preload("IndustryClassifications").
		Preload("RiskProfile").
		First(&profile, "id = ?", clientID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("客户档案不存在: %s", clientID)
		}
		return nil, fmt.Errorf("获取客户档案失败: %w", err)
	}
	return &profile, nil
}

// GetClientsWithRiskProfile 获取根据风险等级筛选的客户列表
func (r *enhancedConflictRepository) GetClientsWithRiskProfile(ctx context.Context, riskLevel string) ([]*models.ClientProfile, error) {
	var profiles []*models.ClientProfile
	if err := r.db.WithContext(ctx).
		Joins("JOIN client_risk_profiles ON client_profiles.id = client_risk_profiles.client_id").
		Where("client_risk_profiles.overall_risk = ?", riskLevel).
		Preload("RiskProfile").
		Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("根据风险等级获取客户列表失败: %w", err)
	}
	return profiles, nil
}

// CreateClientRelationship 创建客户关系
func (r *enhancedConflictRepository) CreateClientRelationship(ctx context.Context, relationship *models.ClientRelationship) error {
	if err := r.db.WithContext(ctx).Create(relationship).Error; err != nil {
		return fmt.Errorf("创建客户关系失败: %w", err)
	}
	return nil
}

// GetClientRelationships 获取客户关系列表
func (r *enhancedConflictRepository) GetClientRelationships(ctx context.Context, clientID string) ([]*models.ClientRelationship, error) {
	var relationships []*models.ClientRelationship
	if err := r.db.WithContext(ctx).
		Preload("Client").
		Preload("RelatedClient").
		Where("client_id = ? OR related_client_id = ?", clientID, clientID).
		Find(&relationships).Error; err != nil {
		return nil, fmt.Errorf("获取客户关系失败: %w", err)
	}
	return relationships, nil
}

// UpdateClientRelationship 更新客户关系
func (r *enhancedConflictRepository) UpdateClientRelationship(ctx context.Context, relationship *models.ClientRelationship) error {
	if err := r.db.WithContext(ctx).Save(relationship).Error; err != nil {
		return fmt.Errorf("更新客户关系失败: %w", err)
	}
	return nil
}

// DeleteClientRelationship 删除客户关系
func (r *enhancedConflictRepository) DeleteClientRelationship(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.ClientRelationship{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("删除客户关系失败: %w", err)
	}
	return nil
}

// CreateClientNameVariant 创建客户名称变体
func (r *enhancedConflictRepository) CreateClientNameVariant(ctx context.Context, variant *models.ClientNameVariant) error {
	if err := r.db.WithContext(ctx).Create(variant).Error; err != nil {
		return fmt.Errorf("创建客户名称变体失败: %w", err)
	}
	return nil
}

// GetClientNameVariants 获取客户名称变体列表
func (r *enhancedConflictRepository) GetClientNameVariants(ctx context.Context, clientID string) ([]*models.ClientNameVariant, error) {
	var variants []*models.ClientNameVariant
	if err := r.db.WithContext(ctx).
		Where("client_id = ?", clientID).
		Find(&variants).Error; err != nil {
		return nil, fmt.Errorf("获取客户名称变体失败: %w", err)
	}
	return variants, nil
}

// UpdateClientNameVariant 更新客户名称变体
func (r *enhancedConflictRepository) UpdateClientNameVariant(ctx context.Context, variant *models.ClientNameVariant) error {
	if err := r.db.WithContext(ctx).Save(variant).Error; err != nil {
		return fmt.Errorf("更新客户名称变体失败: %w", err)
	}
	return nil
}

// DeleteClientNameVariant 删除客户名称变体
func (r *enhancedConflictRepository) DeleteClientNameVariant(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.ClientNameVariant{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("删除客户名称变体失败: %w", err)
	}
	return nil
}

// CreateIndustryClassification 创建行业分类
func (r *enhancedConflictRepository) CreateIndustryClassification(ctx context.Context, classification *models.IndustryClassification) error {
	if err := r.db.WithContext(ctx).Create(classification).Error; err != nil {
		return fmt.Errorf("创建行业分类失败: %w", err)
	}
	return nil
}

// GetIndustryClassifications 获取客户行业分类列表
func (r *enhancedConflictRepository) GetIndustryClassifications(ctx context.Context, clientID string) ([]*models.IndustryClassification, error) {
	var classifications []*models.IndustryClassification
	if err := r.db.WithContext(ctx).
		Where("client_id = ?", clientID).
		Order("is_primary DESC, confidence DESC").
		Find(&classifications).Error; err != nil {
		return nil, fmt.Errorf("获取行业分类失败: %w", err)
	}
	return classifications, nil
}

// UpdateIndustryClassification 更新行业分类
func (r *enhancedConflictRepository) UpdateIndustryClassification(ctx context.Context, classification *models.IndustryClassification) error {
	if err := r.db.WithContext(ctx).Save(classification).Error; err != nil {
		return fmt.Errorf("更新行业分类失败: %w", err)
	}
	return nil
}

// DeleteIndustryClassification 删除行业分类
func (r *enhancedConflictRepository) DeleteIndustryClassification(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.IndustryClassification{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("删除行业分类失败: %w", err)
	}
	return nil
}

// CreateClientRiskProfile 创建客户风险档案
func (r *enhancedConflictRepository) CreateClientRiskProfile(ctx context.Context, riskProfile *models.ClientRiskProfile) error {
	if err := r.db.WithContext(ctx).Create(riskProfile).Error; err != nil {
		return fmt.Errorf("创建客户风险档案失败: %w", err)
	}
	return nil
}

// GetClientRiskProfile 获取客户风险档案
func (r *enhancedConflictRepository) GetClientRiskProfile(ctx context.Context, clientID string) (*models.ClientRiskProfile, error) {
	var riskProfile models.ClientRiskProfile
	if err := r.db.WithContext(ctx).
		First(&riskProfile, "client_id = ?", clientID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("客户风险档案不存在: %s", clientID)
		}
		return nil, fmt.Errorf("获取客户风险档案失败: %w", err)
	}
	return &riskProfile, nil
}

// UpdateClientRiskProfile 更新客户风险档案
func (r *enhancedConflictRepository) UpdateClientRiskProfile(ctx context.Context, riskProfile *models.ClientRiskProfile) error {
	if err := r.db.WithContext(ctx).Save(riskProfile).Error; err != nil {
		return fmt.Errorf("更新客户风险档案失败: %w", err)
	}
	return nil
}

// GetClientsRequiringMonitoring 获取需要监控的客户列表
func (r *enhancedConflictRepository) GetClientsRequiringMonitoring(ctx context.Context) ([]*models.ClientProfile, error) {
	var profiles []*models.ClientProfile
	if err := r.db.WithContext(ctx).
		Joins("JOIN client_risk_profiles ON client_profiles.id = client_risk_profiles.client_id").
		Where("client_risk_profiles.monitoring_required = ? AND client_profiles.client_status IN ('ACTIVE', 'DORMANT')", true).
		Preload("RiskProfile").
		Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("获取需要监控的客户列表失败: %w", err)
	}
	return profiles, nil
}

// ==================== 冲突分类标准实现 ====================

// CreateConflictType 创建冲突类型
func (r *enhancedConflictRepository) CreateConflictType(ctx context.Context, conflictType *models.ConflictType) error {
	if err := r.db.WithContext(ctx).Create(conflictType).Error; err != nil {
		return fmt.Errorf("创建冲突类型失败: %w", err)
	}
	return nil
}

// GetConflictType 获取冲突类型
func (r *enhancedConflictRepository) GetConflictType(ctx context.Context, id string) (*models.ConflictType, error) {
	var conflictType models.ConflictType
	if err := r.db.WithContext(ctx).
		Preload("ClassificationRefs").
		First(&conflictType, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("冲突类型不存在: %s", id)
		}
		return nil, fmt.Errorf("获取冲突类型失败: %w", err)
	}
	return &conflictType, nil
}

// GetConflictTypeByCode 根据代码获取冲突类型
func (r *enhancedConflictRepository) GetConflictTypeByCode(ctx context.Context, code string) (*models.ConflictType, error) {
	var conflictType models.ConflictType
	if err := r.db.WithContext(ctx).
		Preload("ClassificationRefs").
		First(&conflictType, "code = ?", code).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("冲突类型不存在: %s", code)
		}
		return nil, fmt.Errorf("获取冲突类型失败: %w", err)
	}
	return &conflictType, nil
}

// UpdateConflictType 更新冲突类型
func (r *enhancedConflictRepository) UpdateConflictType(ctx context.Context, conflictType *models.ConflictType) error {
	if err := r.db.WithContext(ctx).Save(conflictType).Error; err != nil {
		return fmt.Errorf("更新冲突类型失败: %w", err)
	}
	return nil
}

// DeleteConflictType 删除冲突类型
func (r *enhancedConflictRepository) DeleteConflictType(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.ConflictType{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("删除冲突类型失败: %w", err)
	}
	return nil
}

// GetAllConflictTypes 获取所有冲突类型
func (r *enhancedConflictRepository) GetAllConflictTypes(ctx context.Context) ([]*models.ConflictType, error) {
	var conflictTypes []*models.ConflictType
	if err := r.db.WithContext(ctx).
		Order("category, sub_category, name").
		Find(&conflictTypes).Error; err != nil {
		return nil, fmt.Errorf("获取冲突类型列表失败: %w", err)
	}
	return conflictTypes, nil
}

// GetActiveConflictTypes 获取活跃的冲突类型
func (r *enhancedConflictRepository) GetActiveConflictTypes(ctx context.Context) ([]*models.ConflictType, error) {
	var conflictTypes []*models.ConflictType
	if err := r.db.WithContext(ctx).
		Where("status = ?", "ACTIVE").
		Order("category, sub_category, name").
		Find(&conflictTypes).Error; err != nil {
		return nil, fmt.Errorf("获取活跃冲突类型列表失败: %w", err)
	}
	return conflictTypes, nil
}

// CreateClassificationReference 创建分类标准引用
func (r *enhancedConflictRepository) CreateClassificationReference(ctx context.Context, ref *models.ClassificationReference) error {
	if err := r.db.WithContext(ctx).Create(ref).Error; err != nil {
		return fmt.Errorf("创建分类标准引用失败: %w", err)
	}
	return nil
}

// GetClassificationReferences 获取分类标准引用列表
func (r *enhancedConflictRepository) GetClassificationReferences(ctx context.Context, conflictTypeID string) ([]*models.ClassificationReference, error) {
	var refs []*models.ClassificationReference
	if err := r.db.WithContext(ctx).
		Where("conflict_type_id = ?", conflictTypeID).
		Order("standard_type, standard_name").
		Find(&refs).Error; err != nil {
		return nil, fmt.Errorf("获取分类标准引用失败: %w", err)
	}
	return refs, nil
}

// UpdateClassificationReference 更新分类标准引用
func (r *enhancedConflictRepository) UpdateClassificationReference(ctx context.Context, ref *models.ClassificationReference) error {
	if err := r.db.WithContext(ctx).Save(ref).Error; err != nil {
		return fmt.Errorf("更新分类标准引用失败: %w", err)
	}
	return nil
}

// DeleteClassificationReference 删除分类标准引用
func (r *enhancedConflictRepository) DeleteClassificationReference(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.ClassificationReference{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("删除分类标准引用失败: %w", err)
	}
	return nil
}

// CreateRiskLevelDefinition 创建风险等级定义
func (r *enhancedConflictRepository) CreateRiskLevelDefinition(ctx context.Context, definition *models.RiskLevelDefinition) error {
	if err := r.db.WithContext(ctx).Create(definition).Error; err != nil {
		return fmt.Errorf("创建风险等级定义失败: %w", err)
	}
	return nil
}

// GetRiskLevelDefinition 获取风险等级定义
func (r *enhancedConflictRepository) GetRiskLevelDefinition(ctx context.Context, level string) (*models.RiskLevelDefinition, error) {
	var definition models.RiskLevelDefinition
	if err := r.db.WithContext(ctx).
		First(&definition, "level = ?", level).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("风险等级定义不存在: %s", level)
		}
		return nil, fmt.Errorf("获取风险等级定义失败: %w", err)
	}
	return &definition, nil
}

// UpdateRiskLevelDefinition 更新风险等级定义
func (r *enhancedConflictRepository) UpdateRiskLevelDefinition(ctx context.Context, definition *models.RiskLevelDefinition) error {
	if err := r.db.WithContext(ctx).Save(definition).Error; err != nil {
		return fmt.Errorf("更新风险等级定义失败: %w", err)
	}
	return nil
}

// GetAllRiskLevelDefinitions 获取所有风险等级定义
func (r *enhancedConflictRepository) GetAllRiskLevelDefinitions(ctx context.Context) ([]*models.RiskLevelDefinition, error) {
	var definitions []*models.RiskLevelDefinition
	if err := r.db.WithContext(ctx).
		Order("score_range").
		Find(&definitions).Error; err != nil {
		return nil, fmt.Errorf("获取风险等级定义列表失败: %w", err)
	}
	return definitions, nil
}

// SearchConflictTypes 搜索冲突类型
func (r *enhancedConflictRepository) SearchConflictTypes(ctx context.Context, query *ConflictTypeSearchQuery) (*ConflictTypeSearchResult, error) {
	var conflictTypes []*models.ConflictType
	var total int64

	db := r.db.WithContext(ctx)

	// 应用搜索条件
	if query.Keyword != "" {
		keyword := "%" + query.Keyword + "%"
		db = db.Where("name LIKE ? OR description LIKE ? OR category LIKE ?", keyword, keyword, keyword)
	}

	if query.Category != "" {
		db = db.Where("category = ?", query.Category)
	}

	if query.RiskLevel != "" {
		db = db.Where("default_risk_level = ?", query.RiskLevel)
	}

	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	if query.WaiverPossible != nil {
		db = db.Where("waiver_possible = ?", *query.WaiverPossible)
	}

	// 获取总数
	if err := db.Model(&models.ConflictType{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("获取冲突类型总数失败: %w", err)
	}

	// 应用分页和排序
	if query.PageSize > 0 && query.Page > 0 {
		offset := (query.Page - 1) * query.PageSize
		db = db.Offset(offset).Limit(query.PageSize)
	}

	if query.OrderBy != "" {
		db = db.Order(query.OrderBy)
	} else {
		db = db.Order("category, name")
	}

	// 执行查询
	if err := db.Find(&conflictTypes).Error; err != nil {
		return nil, fmt.Errorf("搜索冲突类型失败: %w", err)
	}

	return &ConflictTypeSearchResult{
		ConflictTypes: conflictTypes,
		Total:         total,
		Page:          query.Page,
		PageSize:      query.PageSize,
	}, nil
}

// GetConflictTypesByStandard 根据标准获取冲突类型
func (r *enhancedConflictRepository) GetConflictTypesByStandard(ctx context.Context, standardType string) ([]*models.ConflictType, error) {
	var conflictTypes []*models.ConflictType
	if err := r.db.WithContext(ctx).
		Joins("JOIN classification_references ON conflict_types.id = classification_references.conflict_type_id").
		Where("classification_references.standard_type = ?", standardType).
		Distinct().
		Find(&conflictTypes).Error; err != nil {
		return nil, fmt.Errorf("根据标准获取冲突类型失败: %w", err)
	}
	return conflictTypes, nil
}

// GetConflictTypesByRiskLevel 根据风险等级获取冲突类型
func (r *enhancedConflictRepository) GetConflictTypesByRiskLevel(ctx context.Context, riskLevel string) ([]*models.ConflictType, error) {
	var conflictTypes []*models.ConflictType
	if err := r.db.WithContext(ctx).
		Where("default_risk_level = ?", riskLevel).
		Order("category, name").
		Find(&conflictTypes).Error; err != nil {
		return nil, fmt.Errorf("根据风险等级获取冲突类型失败: %w", err)
	}
	return conflictTypes, nil
}

// ==================== 查询结构体定义 ====================

// ClientSearchQuery 客户搜索查询
type ClientSearchQuery struct {
	Keyword          string `json:"keyword"`
	ClientType       string `json:"client_type"`
	ClientStatus     string `json:"client_status"`
	AssignedPartner  string `json:"assigned_partner"`
	AssignedAttorney string `json:"assigned_attorney"`
	Page             int    `json:"page"`
	PageSize         int    `json:"page_size"`
	OrderBy          string `json:"order_by"`
}

// ClientSearchResult 客户搜索结果
type ClientSearchResult struct {
	Profiles []*models.ClientProfile `json:"profiles"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

// ConflictTypeSearchQuery 冲突类型搜索查询
type ConflictTypeSearchQuery struct {
	Keyword       string `json:"keyword"`
	Category      string `json:"category"`
	RiskLevel     string `json:"risk_level"`
	Status        string `json:"status"`
	WaiverPossible *bool  `json:"waiver_possible"`
	Page          int    `json:"page"`
	PageSize      int    `json:"page_size"`
	OrderBy       string `json:"order_by"`
}

// ConflictTypeSearchResult 冲突类型搜索结果
type ConflictTypeSearchResult struct {
	ConflictTypes []*models.ConflictType `json:"conflict_types"`
	Total         int64                  `json:"total"`
	Page          int                    `json:"page"`
	PageSize      int                    `json:"page_size"`
}

// StatisticsQuery 统计查询
type StatisticsQuery struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Filters   map[string]interface{} `json:"filters"`
}

// RuleExecutionStats 规则执行统计
type RuleExecutionStats struct {
	TotalExecutions    int64     `json:"total_executions"`
	SuccessRate        float64 `json:"success_rate"`
	AvgExecutionTime   int64     `json:"avg_execution_time"`
	LastExecuted       *time.Time `json:"last_executed"`
	ErrorRate          float64 `json:"error_rate"`
}

// ConflictTrend 冲突趋势
type ConflictTrend struct {
	Period     string `json:"period"`
	Conflicts  int64    `json:"conflicts"`
	Resolutions int64   `json:"resolutions"`
	Pending    int64    `json:"pending"`
}

// RiskDistributionFilters 风险分布筛选器
type RiskDistributionFilters struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	ClientIDs []string  `json:"client_ids"`
}

// RiskDistribution 风险分布
type RiskDistribution struct {
	Low      int64 `json:"low"`
	Medium   int64 `json:"medium"`
	High     int64 `json:"high"`
	Critical int64 `json:"critical"`
}

// WaiverStatistics 豁免统计
type WaiverStatistics struct {
	TotalRequests  int64     `json:"total_requests"`
	ApprovedCount  int64     `json:"approved_count"`
	RejectedCount  int64     `json:"rejected_count"`
	PendingCount   int64     `json:"pending_count"`
	ApprovalRate   float64 `json:"approval_rate"`
	AvgProcessingDays int64  `json:"avg_processing_days"`
}

// WaiverApprovalTrend 豁免审批趋势
type WaiverApprovalTrend struct {
	Period      string `json:"period"`
	Requests    int64    `json:"requests"`
	Approvals   int64    `json:"approvals"`
	Rejections  int64    `json:"rejections"`
}

// ClientRiskStatistics 客户风险统计
type ClientRiskStatistics struct {
	TotalClients      int64     `json:"total_clients"`
	LowRiskClients    int64     `json:"low_risk_clients"`
	MediumRiskClients int64     `json:"medium_risk_clients"`
	HighRiskClients   int64     `json:"high_risk_clients"`
	CriticalRiskClients int64   `json:"critical_risk_clients"`
	MonitoringRequiredClients int64 `json:"monitoring_required_clients"`
}

// HighRiskClient 高风险客户
type HighRiskClient struct {
	ClientID   string  `json:"client_id"`
	ClientName string  `json:"client_name"`
	RiskScore  float64 `json:"risk_score"`
	RiskLevel  string  `json:"risk_level"`
	LastUpdate time.Time `json:"last_update"`
}

// ProcessingEfficiencyStats 处理效率统计
type ProcessingEfficiencyStats struct {
	AvgProcessingTime   int64     `json:"avg_processing_time"`
	FastProcessingRate  float64 `json:"fast_processing_rate"`
	OnTimeCompletionRate float64 `json:"on_time_completion_rate"`
	QueueLength         int64     `json:"queue_length"`
}

// SlaComplianceStats SLA合规统计
type SlaComplianceStats struct {
	OverallComplianceRate float64 `json:"overall_compliance_rate"`
	PriorityCompliance    map[string]float64 `json:"priority_compliance"`
	BreachedRequests     int64     `json:"breached_requests"`
	TotalRequests        int64     `json:"total_requests"`
}