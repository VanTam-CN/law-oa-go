package repositories

import (
	"context"

	"law-oa-go/internal/models"
)

// IntegrationRepositoryInterface 集成数据仓库接口
type IntegrationRepositoryInterface interface {
	// 冲突检测关联管理
	CreateConflictAssociation(ctx context.Context, association *ConflictAssociation) error
	GetConflictAssociationByApprovalID(ctx context.Context, approvalID string) (*ConflictAssociation, error)
	GetConflictAssociationByCheckID(ctx context.Context, checkID string) (*ConflictAssociation, error)
	UpdateConflictAssociationStatus(ctx context.Context, id string, status string) error

	// 案件创建跟踪管理
	CreateCaseCreationTracking(ctx context.Context, tracking *CaseCreationTracking) error
	GetCaseCreationTrackingByApprovalID(ctx context.Context, approvalID string) ([]*CaseCreationTracking, error)
	GetLatestCaseCreationTracking(ctx context.Context, approvalID string) (*CaseCreationTracking, error)
	UpdateCaseCreationTracking(ctx context.Context, id string, updates map[string]interface{}) error

	// 集成配置管理
	CreateIntegrationConfig(ctx context.Context, config *IntegrationConfig) error
	GetIntegrationConfigByName(ctx context.Context, name string) (*IntegrationConfig, error)
	GetIntegrationConfigsByType(ctx context.Context, configType string) ([]*IntegrationConfig, error)
	UpdateIntegrationConfigUsage(ctx context.Context, configID string) error

	// 集成审批申请查询
	GetIntegratedApprovalsWithConflict(ctx context.Context, limit, offset int) ([]*models.ApprovalRequest, int64, error)
	GetApprovalsAwaitingCaseCreation(ctx context.Context, limit, offset int) ([]*models.ApprovalRequest, int64, error)

	// 统计信息
	GetIntegrationStatistics(ctx context.Context) (map[string]interface{}, error)

	// 审批申请关联更新
	UpdateApprovalConflictAssociation(ctx context.Context, approvalID string, conflictResult *models.ConflictCheckResponse) error
	UpdateApprovalCaseAssociation(ctx context.Context, approvalID string, caseResult *models.CaseCreationAssociation) error
}

// 验证接口实现
var _ IntegrationRepositoryInterface = (*IntegrationRepository)(nil)