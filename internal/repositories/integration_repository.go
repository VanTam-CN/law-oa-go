package repositories

import (
	"context"
	"time"

	"law-oa-go/internal/models"

	"gorm.io/gorm"
)

// IntegrationRepository 审批与冲突检测集成数据仓库
type IntegrationRepository struct {
	db *gorm.DB
}

// NewIntegrationRepository 创建集成仓库
func NewIntegrationRepository(db *gorm.DB) *IntegrationRepository {
	return &IntegrationRepository{db: db}
}

// ConflictAssociation 冲突检测关联数据模型
type ConflictAssociation struct {
	ID                 string    `gorm:"primaryKey" json:"id"`
	ApprovalRequestID  string    `gorm:"not null;index" json:"approval_request_id"`
	ConflictCheckID    string    `gorm:"not null;index" json:"conflict_check_id"`
	AssociationStatus  string    `gorm:"default:pending;check:association_status IN ('pending', 'active', 'superseded', 'cancelled')" json:"association_status"`
	AssociationType    string    `gorm:"default:required;check:association_type IN ('required', 'optional', 'conditional')" json:"association_type"`
	RiskLevel          string    `gorm:"check:risk_level IN ('CRITICAL', 'HIGH', 'MEDIUM', 'LOW', 'MINIMAL')" json:"risk_level"`
	RiskScore          *float64  `gorm:"type:decimal(5,2)" json:"risk_score"`
	ConflictCount      int       `gorm:"default:0" json:"conflict_count"`
	RequiresApproval   bool      `gorm:"default:false" json:"requires_approval"`
	AutoApproval       bool      `gorm:"default:false" json:"auto_approval"`
	ApprovalConditions string    `gorm:"type:jsonb" json:"approval_conditions"`
	MitigationMeasures string    `gorm:"type:jsonb" json:"mitigation_measures"`
	DataMapping        string    `gorm:"type:jsonb" json:"data_mapping"`
	MappedFields       string    `gorm:"type:jsonb" json:"mapped_fields"`
	ValidationErrors   string    `gorm:"type:jsonb" json:"validation_errors"`
	CreatedBy          string    `gorm:"not null" json:"created_by"`
	UpdatedBy          *string   `json:"updated_by"`
	CreatedAt          time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt          time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (ConflictAssociation) TableName() string {
	return "approval_conflict_associations"
}

// CaseCreationTracking 案件创建跟踪数据模型
type CaseCreationTracking struct {
	ID                  string     `gorm:"primaryKey" json:"id"`
	ApprovalRequestID   string     `gorm:"not null;index" json:"approval_request_id"`
	CaseID              *string    `gorm:"index" json:"case_id"`
	CaseNumber          *string    `gorm:"index" json:"case_number"`
	CaseType            *string    `json:"case_type"`
	CreationStatus      string     `gorm:"default:pending;check:creation_status IN ('pending', 'processing', 'completed', 'failed', 'retrying')" json:"creation_status"`
	CreationStep        *string    `json:"creation_step"`
	ProgressPercentage  float64    `gorm:"type:decimal(5,2);default:0.00" json:"progress_percentage"`
	ErrorCode           *string    `json:"error_code"`
	ErrorMessage        *string    `gorm:"type:text" json:"error_message"`
	ErrorDetails        string     `gorm:"type:jsonb" json:"error_details"`
	RetryCount          int        `gorm:"default:0" json:"retry_count"`
	MaxRetries          int        `gorm:"default:3" json:"max_retries"`
	DataMapping         string     `gorm:"type:jsonb" json:"data_mapping"`
	MappedFields        string     `gorm:"type:jsonb" json:"mapped_fields"`
	UnmappedFields      string     `gorm:"type:jsonb" json:"unmapped_fields"`
	AppliedConditions   string     `gorm:"type:jsonb" json:"applied_conditions"`
	ImposedRequirements string     `gorm:"type:jsonb" json:"imposed_requirements"`
	WorkflowActions     string     `gorm:"type:jsonb" json:"workflow_actions"`
	CreatedBy           string     `gorm:"not null" json:"created_by"`
	ProcessedBy         *string    `json:"processed_by"`
	CreatedAt           time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	ProcessedAt         *time.Time `json:"processed_at"`
	CompletedAt         *time.Time `json:"completed_at"`
}

func (CaseCreationTracking) TableName() string {
	return "approval_case_creation_tracking"
}

// IntegrationConfig 集成配置数据模型
type IntegrationConfig struct {
	ID                      string     `gorm:"primaryKey" json:"id"`
	ConfigName              string     `gorm:"uniqueIndex;not null" json:"config_name"`
	ConfigType              string     `gorm:"not null;check:config_type IN ('conflict_approval', 'case_creation', 'workflow_override')" json:"config_type"`
	ApplicableApprovalTypes string     `gorm:"type:jsonb" json:"applicable_approval_types"`
	ApplicableWorkflows     string     `gorm:"type:jsonb" json:"applicable_workflows"`
	ApplicableDepartments   string     `gorm:"type:jsonb" json:"applicable_departments"`
	ApplicableRoles         string     `gorm:"type:jsonb" json:"applicable_roles"`
	TriggerRules            string     `gorm:"not null;type:jsonb" json:"trigger_rules"`
	ProcessingRules         string     `gorm:"not null;type:jsonb" json:"processing_rules"`
	ValidationRules         string     `gorm:"type:jsonb" json:"validation_rules"`
	WorkflowConfig          string     `gorm:"not null;type:jsonb" json:"workflow_config"`
	ApprovalConfig          string     `gorm:"type:jsonb" json:"approval_config"`
	NotificationConfig      string     `gorm:"type:jsonb" json:"notification_config"`
	FieldMapping            string     `gorm:"type:jsonb" json:"field_mapping"`
	DataTransformation      string     `gorm:"type:jsonb" json:"data_transformation"`
	ValidationMapping       string     `gorm:"type:jsonb" json:"validation_mapping"`
	Status                  string     `gorm:"default:active;check:status IN ('active', 'inactive', 'testing')" json:"status"`
	Version                 int        `gorm:"default:1" json:"version"`
	Priority                int        `gorm:"default:0" json:"priority"`
	Conditions              string     `gorm:"type:jsonb" json:"conditions"`
	Limitations             string     `gorm:"type:jsonb" json:"limitations"`
	CreatedBy               string     `gorm:"not null" json:"created_by"`
	UpdatedBy               *string    `json:"updated_by"`
	CreatedAt               time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt               time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	EffectiveDate           *time.Time `json:"effective_date"`
	ExpiryDate              *time.Time `json:"expiry_date"`
	UsageCount              int        `gorm:"default:0" json:"usage_count"`
	LastUsedDate            *time.Time `json:"last_used_date"`
}

func (IntegrationConfig) TableName() string {
	return "approval_integration_configs"
}

// CreateConflictAssociation 创建冲突检测关联记录
func (r *IntegrationRepository) CreateConflictAssociation(ctx context.Context, association *ConflictAssociation) error {
	if err := r.db.WithContext(ctx).Create(association).Error; err != nil {
		return NewRepositoryError("create conflict association", "ConflictAssociation", err)
	}
	return nil
}

// GetConflictAssociationByApprovalID 根据审批ID获取冲突关联
func (r *IntegrationRepository) GetConflictAssociationByApprovalID(ctx context.Context, approvalID string) (*ConflictAssociation, error) {
	var association ConflictAssociation
	err := r.db.WithContext(ctx).
		Where("approval_request_id = ? AND association_status = ?", approvalID, "active").
		First(&association).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, NewRepositoryError("get conflict association by approval id", "ConflictAssociation", err)
	}
	return &association, nil
}

// GetConflictAssociationByCheckID 根据冲突检测ID获取关联
func (r *IntegrationRepository) GetConflictAssociationByCheckID(ctx context.Context, checkID string) (*ConflictAssociation, error) {
	var association ConflictAssociation
	err := r.db.WithContext(ctx).
		Where("conflict_check_id = ? AND association_status = ?", checkID, "active").
		First(&association).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, NewRepositoryError("get conflict association by check id", "ConflictAssociation", err)
	}
	return &association, nil
}

// UpdateConflictAssociationStatus 更新冲突关联状态
func (r *IntegrationRepository) UpdateConflictAssociationStatus(ctx context.Context, id string, status string) error {
	result := r.db.WithContext(ctx).
		Model(&ConflictAssociation{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"association_status": status,
			"updated_at":         time.Now(),
		})

	if result.Error != nil {
		return NewRepositoryError("update conflict association status", "ConflictAssociation", result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("update conflict association status", "ConflictAssociation", id, ErrRecordNotFound)
	}
	return nil
}

// CreateCaseCreationTracking 创建案件创建跟踪记录
func (r *IntegrationRepository) CreateCaseCreationTracking(ctx context.Context, tracking *CaseCreationTracking) error {
	if err := r.db.WithContext(ctx).Create(tracking).Error; err != nil {
		return NewRepositoryError("create case creation tracking", "CaseCreationTracking", err)
	}
	return nil
}

// GetCaseCreationTrackingByApprovalID 根据审批ID获取案件创建跟踪
func (r *IntegrationRepository) GetCaseCreationTrackingByApprovalID(ctx context.Context, approvalID string) ([]*CaseCreationTracking, error) {
	var trackings []*CaseCreationTracking
	err := r.db.WithContext(ctx).
		Where("approval_request_id = ?", approvalID).
		Order("created_at DESC").
		Find(&trackings).Error
	if err != nil {
		return nil, NewRepositoryError("get case creation tracking by approval id", "CaseCreationTracking", err)
	}
	return trackings, nil
}

// GetLatestCaseCreationTracking 获取最新的案件创建跟踪记录
func (r *IntegrationRepository) GetLatestCaseCreationTracking(ctx context.Context, approvalID string) (*CaseCreationTracking, error) {
	var tracking CaseCreationTracking
	err := r.db.WithContext(ctx).
		Where("approval_request_id = ?", approvalID).
		Order("created_at DESC").
		First(&tracking).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, NewRepositoryError("get latest case creation tracking", "CaseCreationTracking", err)
	}
	return &tracking, nil
}

// UpdateCaseCreationTracking 更新案件创建跟踪
func (r *IntegrationRepository) UpdateCaseCreationTracking(ctx context.Context, id string, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).
		Model(&CaseCreationTracking{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return NewRepositoryError("update case creation tracking", "CaseCreationTracking", result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("update case creation tracking", "CaseCreationTracking", id, ErrRecordNotFound)
	}
	return nil
}

// ClaimCaseCreation atomically moves an approved request into processing.
// The conditional update is the cross-instance lock; an in-process mutex
// would not protect deployments with more than one backend replica.
func (r *IntegrationRepository) ClaimCaseCreation(ctx context.Context, approvalID, actorID string) (bool, error) {
	result := r.db.WithContext(ctx).Model(&models.ApprovalRequest{}).
		Where("id = ? AND status = ? AND case_created = ?", approvalID, models.ApprovalStatusApproved, false).
		Where("(case_creation_status IS NULL OR case_creation_status IN ?)", []string{"", "pending", "failed", "retrying"}).
		Updates(map[string]interface{}{
			"case_creation_status": "processing",
			"case_creation_time":   time.Now(),
			"updated_by":           actorID,
			"updated_at":           time.Now(),
		})
	if result.Error != nil {
		return false, NewRepositoryError("claim case creation", "ApprovalRequest", result.Error)
	}
	return result.RowsAffected == 1, nil
}

func (r *IntegrationRepository) MarkCaseCreationFailed(ctx context.Context, approvalID, message string) error {
	result := r.db.WithContext(ctx).Model(&models.ApprovalRequest{}).
		Where("id = ?", approvalID).
		Updates(map[string]interface{}{
			"case_created":         false,
			"case_creation_status": "failed",
			"case_creation_time":   time.Now(),
			"updated_at":           time.Now(),
		})
	if result.Error != nil {
		return NewRepositoryError("mark case creation failed", "ApprovalRequest", result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("mark case creation failed", "ApprovalRequest", approvalID, ErrRecordNotFound)
	}
	return nil
}

// GetConflictCheckRecord 获取集成审批关联的冲突检测记录
func (r *IntegrationRepository) GetConflictCheckRecord(ctx context.Context, checkID string) (*models.ConflictCheckRecord, error) {
	var record models.ConflictCheckRecord
	err := r.db.WithContext(ctx).Where("check_id = ?", checkID).First(&record).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, NewRepositoryError("get conflict check record", "ConflictCheckRecord", err)
	}
	return &record, nil
}

// GetLatestConflictReview reads the append-only professional conclusion that
// is linked to an approval's frozen conflict check.
func (r *IntegrationRepository) GetLatestConflictReview(ctx context.Context, checkID string) (*models.ConflictReview, error) {
	var review models.ConflictReview
	err := r.db.WithContext(ctx).
		Where("check_id = ?", checkID).
		Order("created_at DESC").
		First(&review).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, NewRepositoryError("get latest conflict review", "ConflictReview", err)
	}
	return &review, nil
}

// CreateIntegrationConfig 创建集成配置
func (r *IntegrationRepository) CreateIntegrationConfig(ctx context.Context, config *IntegrationConfig) error {
	if err := r.db.WithContext(ctx).Create(config).Error; err != nil {
		return NewRepositoryError("create integration config", "IntegrationConfig", err)
	}
	return nil
}

// GetIntegrationConfigByName 根据名称获取集成配置
func (r *IntegrationRepository) GetIntegrationConfigByName(ctx context.Context, name string) (*IntegrationConfig, error) {
	var config IntegrationConfig
	err := r.db.WithContext(ctx).
		Where("config_name = ? AND status = ?", name, "active").
		First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, NewRepositoryError("get integration config by name", "IntegrationConfig", err)
	}
	return &config, nil
}

// GetIntegrationConfigsByType 根据类型获取集成配置列表
func (r *IntegrationRepository) GetIntegrationConfigsByType(ctx context.Context, configType string) ([]*IntegrationConfig, error) {
	var configs []*IntegrationConfig
	err := r.db.WithContext(ctx).
		Where("config_type = ? AND status = ?", configType, "active").
		Order("priority DESC, created_at ASC").
		Find(&configs).Error
	if err != nil {
		return nil, NewRepositoryError("get integration configs by type", "IntegrationConfig", err)
	}
	return configs, nil
}

// UpdateIntegrationConfigUsage 更新集成配置使用统计
func (r *IntegrationRepository) UpdateIntegrationConfigUsage(ctx context.Context, configID string) error {
	result := r.db.WithContext(ctx).
		Model(&IntegrationConfig{}).
		Where("id = ?", configID).
		Updates(map[string]interface{}{
			"usage_count":    gorm.Expr("usage_count + 1"),
			"last_used_date": time.Now(),
			"updated_at":     time.Now(),
		})

	if result.Error != nil {
		return NewRepositoryError("update integration config usage", "IntegrationConfig", result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("update integration config usage", "IntegrationConfig", configID, ErrRecordNotFound)
	}
	return nil
}

// GetIntegratedApprovalsWithConflict 获取包含冲突检测的审批申请
func (r *IntegrationRepository) GetIntegratedApprovalsWithConflict(ctx context.Context, limit, offset int) ([]*models.ApprovalRequest, int64, error) {
	var approvals []*models.ApprovalRequest
	var total int64

	query := r.db.WithContext(ctx).
		Model(&models.ApprovalRequest{}).
		Where("conflict_check_id IS NOT NULL AND deleted_at IS NULL")

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, NewRepositoryError("count integrated approvals with conflict", "ApprovalRequest", err)
	}

	// 获取分页数据
	err := query.Preload("ConflictCheck").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&approvals).Error
	if err != nil {
		return nil, 0, NewRepositoryError("get integrated approvals with conflict", "ApprovalRequest", err)
	}

	return approvals, total, nil
}

// GetApprovalsAwaitingCaseCreation 获取等待案件创建的已批准申请
func (r *IntegrationRepository) GetApprovalsAwaitingCaseCreation(ctx context.Context, limit, offset int) ([]*models.ApprovalRequest, int64, error) {
	var approvals []*models.ApprovalRequest
	var total int64

	query := r.db.WithContext(ctx).
		Model(&models.ApprovalRequest{}).
		Where("status = ? AND case_created = ? AND deleted_at IS NULL",
			models.ApprovalStatusApproved, false)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, NewRepositoryError("count approvals awaiting case creation", "ApprovalRequest", err)
	}

	// 获取分页数据
	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&approvals).Error
	if err != nil {
		return nil, 0, NewRepositoryError("get approvals awaiting case creation", "ApprovalRequest", err)
	}

	return approvals, total, nil
}

// GetIntegrationStatistics 获取集成统计信息
func (r *IntegrationRepository) GetIntegrationStatistics(ctx context.Context) (map[string]interface{}, error) {
	var stats map[string]interface{}

	// 调用数据库函数获取统计信息
	err := r.db.WithContext(ctx).Raw("SELECT get_integration_stats() as stats").Scan(&stats).Error
	if err != nil {
		return nil, NewRepositoryError("get integration statistics", "Integration", err)
	}

	return stats, nil
}

// UpdateApprovalConflictAssociation 更新审批申请的冲突关联信息
func (r *IntegrationRepository) UpdateApprovalConflictAssociation(ctx context.Context, approvalID string, conflictResult *models.ConflictCheckResponse) error {
	updates := map[string]interface{}{
		"conflict_check_id":   conflictResult.CheckID,
		"conflict_risk_level": conflictResult.RiskAssessment.OverallRisk,
		"conflict_check_time": conflictResult.CheckTime,
		"conflict_result":     conflictResult,
		"updated_at":          time.Now(),
	}

	result := r.db.WithContext(ctx).
		Model(&models.ApprovalRequest{}).
		Where("id = ?", approvalID).
		Updates(updates)

	if result.Error != nil {
		return NewRepositoryError("update approval conflict association", "ApprovalRequest", result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("update approval conflict association", "ApprovalRequest", approvalID, ErrRecordNotFound)
	}
	return nil
}

// UpdateApprovalCaseAssociation 更新审批申请的案件关联信息
func (r *IntegrationRepository) UpdateApprovalCaseAssociation(ctx context.Context, approvalID string, caseResult *models.CaseCreationAssociation) error {
	updates := map[string]interface{}{
		"case_created":         true,
		"created_case_id":      caseResult.CaseID,
		"case_creation_time":   caseResult.CreationTime,
		"case_creation_status": caseResult.Status,
		"updated_at":           time.Now(),
	}

	result := r.db.WithContext(ctx).
		Model(&models.ApprovalRequest{}).
		Where("id = ?", approvalID).
		Updates(updates)

	if result.Error != nil {
		return NewRepositoryError("update approval case association", "ApprovalRequest", result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("update approval case association", "ApprovalRequest", approvalID, ErrRecordNotFound)
	}
	return nil
}
