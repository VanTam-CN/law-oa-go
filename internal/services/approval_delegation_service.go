package services

import (
	"context"
	"fmt"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"

	"github.com/google/uuid"
)

// ApprovalDelegationService 代理审批服务接口
type ApprovalDelegationService interface {
	CreateDelegation(ctx context.Context, req *CreateDelegationRequest) (*models.ApprovalDelegation, error)
	GetActiveDelegations(ctx context.Context, delegatorID string) ([]*models.ApprovalDelegation, error)
	RevokeDelegation(ctx context.Context, id string) error
	GetEffectiveApprover(ctx context.Context, originalApproverID string) (effectiveApproverID string, delegationID string, isDelegated bool, err error)
	ListDelegations(ctx context.Context, params *repositories.DelegationListParams) ([]*models.ApprovalDelegation, int64, error)
	GetDelegation(ctx context.Context, id string) (*models.ApprovalDelegation, error)
}

// CreateDelegationRequest 创建代理配置请求
type CreateDelegationRequest struct {
	DelegatorID string     `json:"delegator_id" binding:"required"`
	DelegateID  string     `json:"delegate_id" binding:"required"`
	ValidFrom   time.Time  `json:"valid_from" binding:"required"`
	ValidUntil  *time.Time `json:"valid_until"`
	Reason      string     `json:"reason"`
	CreatedBy   string     `json:"created_by" binding:"required"`
}

// approvalDelegationService 代理审批服务实现
type approvalDelegationService struct {
	delegationRepo repositories.ApprovalDelegationRepository
}

// NewApprovalDelegationService 创建代理审批服务
func NewApprovalDelegationService(delegationRepo repositories.ApprovalDelegationRepository) ApprovalDelegationService {
	return &approvalDelegationService{delegationRepo: delegationRepo}
}

// checkCircularDelegation 检查是否会形成循环代理
// 递归遍历代理人(delegateID)的代理链，检查是否会回到原委托人(delegatorID)
func (s *approvalDelegationService) checkCircularDelegation(ctx context.Context, delegatorID, delegateID string) (bool, error) {
	const maxDepth = 5
	visited := make(map[string]bool)
	currentID := delegateID

	for i := 0; i < maxDepth; i++ {
		// 如果找到回到原委托人，形成循环
		if currentID == delegatorID {
			return true, nil
		}

		// 防止重复访问同一节点
		if visited[currentID] {
			return false, fmt.Errorf("检测到代理链异常：存在重复节点")
		}
		visited[currentID] = true

		// 获取当前节点的有效代理
		delegation, err := s.delegationRepo.GetValidDelegate(ctx, currentID)
		if err != nil {
			return false, fmt.Errorf("检查代理链失败: %w", err)
		}
		if delegation == nil {
			// 链条结束，无循环
			return false, nil
		}

		currentID = delegation.DelegateID
	}

	// 超过最大深度，返回错误提示
	return false, fmt.Errorf("代理链过长：超过%d层，请简化代理关系", maxDepth)
}

// CreateDelegation 创建代理配置
func (s *approvalDelegationService) CreateDelegation(ctx context.Context, req *CreateDelegationRequest) (*models.ApprovalDelegation, error) {
	// 验证：不能自己代理自己
	if req.DelegatorID == req.DelegateID {
		return nil, fmt.Errorf("不能为自己配置代理审批")
	}

	// 验证：有效时间范围
	if req.ValidUntil != nil && req.ValidUntil.Before(req.ValidFrom) {
		return nil, fmt.Errorf("结束时间不能早于开始时间")
	}

	// 验证：检查循环代理（包括多节点环路如 A->B->C->A）
	isCircular, err := s.checkCircularDelegation(ctx, req.DelegatorID, req.DelegateID)
	if err != nil {
		return nil, fmt.Errorf("检查循环代理失败: %w", err)
	}
	if isCircular {
		return nil, fmt.Errorf("不允许循环代理: 形成闭环代理链")
	}

	delegation := &models.ApprovalDelegation{
		ID:          uuid.New().String(),
		DelegatorID: req.DelegatorID,
		DelegateID:  req.DelegateID,
		ValidFrom:   req.ValidFrom,
		ValidUntil:  req.ValidUntil,
		IsActive:    true,
		Reason:      req.Reason,
		CreatedBy:   req.CreatedBy,
	}

	if err := s.delegationRepo.Create(ctx, delegation); err != nil {
		return nil, err
	}

	return delegation, nil
}

// GetActiveDelegations 获取活跃的代理配置
func (s *approvalDelegationService) GetActiveDelegations(ctx context.Context, delegatorID string) ([]*models.ApprovalDelegation, error) {
	return s.delegationRepo.GetActiveDelegations(ctx, delegatorID)
}

// RevokeDelegation 撤销代理配置
func (s *approvalDelegationService) RevokeDelegation(ctx context.Context, id string) error {
	delegation, err := s.delegationRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if delegation == nil {
		return fmt.Errorf("代理配置不存在")
	}

	return s.delegationRepo.Update(ctx, id, map[string]interface{}{
		"is_active": false,
	})
}

// GetEffectiveApprover 获取有效的审批人（考虑代理替换）
// 支持递归查找代理链，最大深度5层
// 返回: 实际审批人ID, 代理配置ID, 是否为代理, 错误
func (s *approvalDelegationService) GetEffectiveApprover(ctx context.Context, originalApproverID string) (string, string, bool, error) {
	const maxDepth = 5
	currentID := originalApproverID
	var delegationID string
	var isDelegated bool

	// 递归查找最终有效审批人
	for i := 0; i < maxDepth; i++ {
		delegation, err := s.delegationRepo.GetValidDelegate(ctx, currentID)
		if err != nil {
			return originalApproverID, "", false, err
		}
		if delegation == nil {
			// 没有更多代理，返回当前找到的审批人
			break
		}

		currentID = delegation.DelegateID
		delegationID = delegation.ID
		isDelegated = true
	}

	return currentID, delegationID, isDelegated, nil
}

// ListDelegations 列表查询代理配置
func (s *approvalDelegationService) ListDelegations(ctx context.Context, params *repositories.DelegationListParams) ([]*models.ApprovalDelegation, int64, error) {
	return s.delegationRepo.List(ctx, params)
}

// GetDelegation 获取代理配置详情
func (s *approvalDelegationService) GetDelegation(ctx context.Context, id string) (*models.ApprovalDelegation, error) {
	return s.delegationRepo.GetByID(ctx, id)
}
