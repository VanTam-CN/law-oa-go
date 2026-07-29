package services

import (
	"context"
	"fmt"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/security"
)

// EntityService 实体服务接口
type EntityService interface {
	// 实体CRUD
	CreateEntity(ctx context.Context, entity *models.Entity) error
	GetEntity(ctx context.Context, id uint) (*models.Entity, error)
	UpdateEntity(ctx context.Context, id uint, updates map[string]interface{}) error
	DeleteEntity(ctx context.Context, id uint) error
	ListEntities(ctx context.Context, params *EntityListParams) ([]*models.Entity, int64, error)

	// 搜索功能
	SearchEntities(ctx context.Context, params *EntitySearchParams) ([]*models.Entity, error)
	FindByName(ctx context.Context, name string) (*models.Entity, error)
	FindByIdentityNumber(ctx context.Context, identityNumber string) (*models.Entity, error)

	// 关系管理
	AddRelation(ctx context.Context, relation *models.EntityRelation) error
	GetRelations(ctx context.Context, entityID uint, activeOnly bool) ([]*models.EntityRelation, error)
	GetRelatedEntities(ctx context.Context, entityID uint, relationType models.RelationType, maxDepth int) ([]*models.Entity, error)
	RemoveRelation(ctx context.Context, relationID uint) error

	// 名称历史
	AddNameHistory(ctx context.Context, history *models.EntityNameHistory) error
	GetNameHistory(ctx context.Context, entityID uint) ([]*models.EntityNameHistory, error)

	// 案件当事人
	AddCaseParty(ctx context.Context, party *models.CaseParty) error
	GetCaseParties(ctx context.Context, caseID uint) ([]*models.CaseParty, error)
	RemoveCaseParty(ctx context.Context, caseID, entityID uint) error

	// 批量操作
	BatchCreateEntities(ctx context.Context, entities []*models.Entity) error
	BatchCreateRelations(ctx context.Context, relations []*models.EntityRelation) error

	// 冲突检测辅助
	FindPotentialConflicts(ctx context.Context, entityID uint, searchYears int) (*ConflictSearchResult, error)
}

// EntityListParams 实体列表参数
type EntityListParams struct {
	Page       int
	PageSize   int
	EntityType *models.EntityType
	Status     *models.EntityStatus
	Search     string
}

// EntitySearchParams 实体搜索参数
type EntitySearchParams struct {
	Name            string
	Alias           string
	IdentityNumber  string
	IdentityType    *models.IdentityType
	EntityType      *models.EntityType
	Status          *models.EntityStatus
	IncludeInactive bool
	Limit           int
}

// ConflictSearchResult 冲突搜索结果
type ConflictSearchResult struct {
	MatchedEntities []*models.Entity
	RelatedEntities []*models.Entity
	NameMatches     []*models.Entity
	TotalMatches    int
	HighRiskCount   int
	MediumRiskCount int
}

// entityService 实体服务实现
type entityService struct {
	entityRepo repositories.EntityRepository
}

// NewEntityService 创建实体服务
func NewEntityService(entityRepo repositories.EntityRepository) EntityService {
	return &entityService{
		entityRepo: entityRepo,
	}
}

// CreateEntity 创建实体
func (s *entityService) CreateEntity(ctx context.Context, entity *models.Entity) error {
	// 验证实体数据
	if entity.Name == "" {
		return fmt.Errorf("实体名称不能为空")
	}
	if entity.EntityType == "" {
		return fmt.Errorf("实体类型不能为空")
	}

	// 检查证件号是否已存在
	if entity.IdentityNumber != "" {
		existing, err := s.entityRepo.SearchByIdentityNumber(ctx, entity.IdentityNumber, 1)
		if err == nil && len(existing) > 0 {
			// 检查是否是同一个实体
			if existing[0].ID != entity.ID {
				return fmt.Errorf("证件号已被使用")
			}
		}
	}

	return s.entityRepo.Create(ctx, entity)
}

// GetEntity 获取实体
func (s *entityService) GetEntity(ctx context.Context, id uint) (*models.Entity, error) {
	return s.entityRepo.GetByID(ctx, id)
}

// UpdateEntity 更新实体
func (s *entityService) UpdateEntity(ctx context.Context, id uint, updates map[string]interface{}) error {
	// 如果更新名称，需要记录名称历史
	if newName, ok := updates["name"].(string); ok {
		entity, err := s.entityRepo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if entity.Name != newName {
			history := &models.EntityNameHistory{
				EntityID:     id,
				OldName:      entity.Name,
				NewName:      newName,
				ChangeDate:   time.Now(),
				ChangeReason: "更新",
			}
			_ = s.entityRepo.AddNameHistory(ctx, history)
		}
	}

	return s.entityRepo.Update(ctx, id, updates)
}

// DeleteEntity 删除实体
func (s *entityService) DeleteEntity(ctx context.Context, id uint) error {
	return s.entityRepo.Delete(ctx, id)
}

// ListEntities 列出实体
func (s *entityService) ListEntities(ctx context.Context, params *EntityListParams) ([]*models.Entity, int64, error) {
	conditions := make(map[string]interface{})
	offset := (params.Page - 1) * params.PageSize

	if params.EntityType != nil {
		conditions["entity_type"] = *params.EntityType
	}
	if params.Status != nil {
		conditions["status"] = *params.Status
	}
	if params.Search != "" {
		// 搜索条件需要特殊处理，这里简化
		conditions["name LIKE"] = "%" + params.Search + "%"
	}

	return s.entityRepo.List(ctx, offset, params.PageSize, conditions)
}

// SearchEntities 搜索实体
func (s *entityService) SearchEntities(ctx context.Context, params *EntitySearchParams) ([]*models.Entity, error) {
	// 使用仓储的 AdvancedSearch 方法
	searchParams := &repositories.EntitySearchParams{
		Name:           params.Name,
		IdentityNumber: params.IdentityNumber,
		Limit:          params.Limit,
	}

	if params.EntityType != nil {
		searchParams.EntityType = string(*params.EntityType)
	}
	if params.Status != nil {
		searchParams.Status = string(*params.Status)
	}

	// 需要在仓储中实现 AdvancedSearch 方法或使用现有方法组合
	// 这里简化处理
	var entities []*models.Entity

	// 按名称搜索
	if params.Name != "" {
		byName, _ := s.entityRepo.SearchByName(ctx, params.Name, params.Limit)
		entities = append(entities, byName...)
	}

	// 按别名搜索
	if params.Alias != "" {
		byAlias, _ := s.entityRepo.SearchByAlias(ctx, params.Alias, params.Limit)
		entities = append(entities, byAlias...)
	}

	// 按曾用名搜索
	if params.Name != "" {
		byFormerName, _ := s.entityRepo.SearchByFormerName(ctx, params.Name, params.Limit)
		entities = append(entities, byFormerName...)
	}

	// 去重
	uniqueEntities := make(map[uint]*models.Entity)
	for _, e := range entities {
		uniqueEntities[e.ID] = e
	}

	result := make([]*models.Entity, 0, len(uniqueEntities))
	for _, e := range uniqueEntities {
		result = append(result, e)
	}

	return result, nil
}

// FindByName 按名称查找实体
func (s *entityService) FindByName(ctx context.Context, name string) (*models.Entity, error) {
	entities, err := s.entityRepo.SearchByName(ctx, name, 1)
	if err != nil {
		return nil, err
	}
	if len(entities) == 0 {
		return nil, repositories.ErrRecordNotFound
	}
	return entities[0], nil
}

// FindByIdentityNumber 按证件号查找实体
func (s *entityService) FindByIdentityNumber(ctx context.Context, identityNumber string) (*models.Entity, error) {
	entities, err := s.entityRepo.SearchByIdentityNumber(ctx, identityNumber, 1)
	if err != nil {
		return nil, err
	}
	if len(entities) == 0 {
		return nil, repositories.ErrRecordNotFound
	}
	return entities[0], nil
}

// AddRelation 添加关联关系
func (s *entityService) AddRelation(ctx context.Context, relation *models.EntityRelation) error {
	// 验证关系
	if relation.SourceEntityID == relation.TargetEntityID {
		return fmt.Errorf("源实体和目标实体不能相同")
	}
	if relation.RelationType == "" {
		return fmt.Errorf("关系类型不能为空")
	}

	// 验证实体存在
	_, err := s.entityRepo.GetByID(ctx, relation.SourceEntityID)
	if err != nil {
		return fmt.Errorf("源实体不存在: %w", err)
	}
	_, err = s.entityRepo.GetByID(ctx, relation.TargetEntityID)
	if err != nil {
		return fmt.Errorf("目标实体不存在: %w", err)
	}

	return nil // 需要仓储实现 CreateRelation 方法
}

// GetRelations 获取关联关系
func (s *entityService) GetRelations(ctx context.Context, entityID uint, activeOnly bool) ([]*models.EntityRelation, error) {
	return s.entityRepo.GetRelations(ctx, entityID, activeOnly)
}

// GetRelatedEntities 获取关联实体
func (s *entityService) GetRelatedEntities(ctx context.Context, entityID uint, relationType models.RelationType, maxDepth int) ([]*models.Entity, error) {
	return s.entityRepo.GetRelatedEntities(ctx, entityID, relationType, maxDepth)
}

// RemoveRelation 移除关联关系
func (s *entityService) RemoveRelation(ctx context.Context, relationID uint) error {
	// 需要仓储实现 RemoveRelation 方法
	return fmt.Errorf("方法待实现")
}

// AddNameHistory 添加名称历史
func (s *entityService) AddNameHistory(ctx context.Context, history *models.EntityNameHistory) error {
	return s.entityRepo.AddNameHistory(ctx, history)
}

// GetNameHistory 获取名称历史
func (s *entityService) GetNameHistory(ctx context.Context, entityID uint) ([]*models.EntityNameHistory, error) {
	return s.entityRepo.GetNameHistory(ctx, entityID)
}

// AddCaseParty 添加案件当事人
func (s *entityService) AddCaseParty(ctx context.Context, party *models.CaseParty) error {
	return s.entityRepo.AddCaseParty(ctx, party)
}

// GetCaseParties 获取案件当事人
func (s *entityService) GetCaseParties(ctx context.Context, caseID uint) ([]*models.CaseParty, error) {
	return s.entityRepo.GetCaseParties(ctx, caseID)
}

// RemoveCaseParty 移除案件当事人
func (s *entityService) RemoveCaseParty(ctx context.Context, caseID, entityID uint) error {
	return s.entityRepo.RemoveCaseParty(ctx, caseID, entityID)
}

// BatchCreateEntities 批量创建实体
func (s *entityService) BatchCreateEntities(ctx context.Context, entities []*models.Entity) error {
	return s.entityRepo.BatchCreate(ctx, entities)
}

// BatchCreateRelations 批量创建关联关系
func (s *entityService) BatchCreateRelations(ctx context.Context, relations []*models.EntityRelation) error {
	return s.entityRepo.BatchCreateRelations(ctx, relations)
}

// FindPotentialConflicts 查找潜在冲突
func (s *entityService) FindPotentialConflicts(ctx context.Context, entityID uint, searchYears int) (*ConflictSearchResult, error) {
	result := &ConflictSearchResult{
		MatchedEntities: make([]*models.Entity, 0),
		RelatedEntities: make([]*models.Entity, 0),
		NameMatches:     make([]*models.Entity, 0),
	}

	// 获取实体信息
	entity, err := s.entityRepo.GetByID(ctx, entityID)
	if err != nil {
		return nil, err
	}

	// 1. 按名称搜索
	nameMatches, _ := s.entityRepo.SearchByName(ctx, entity.Name, 100)
	for _, e := range nameMatches {
		if e.ID != entityID {
			result.NameMatches = append(result.NameMatches, e)
		}
	}

	// 2. 按证件号搜索。主体档案已密文保存时，只在服务端内存中解密，仓储
	// 会立即转换为 keyed digest，不能把明文写回响应或审计记录。
	identityNumber := entity.IdentityNumber
	if identityNumber == "" && entity.IdentityNumberCiphertext != "" {
		identityNumber, _ = security.DecryptIdentityNumber(entity.IdentityNumberCiphertext)
	}
	if identityNumber != "" {
		idMatches, _ := s.entityRepo.SearchByIdentityNumber(ctx, identityNumber, 100)
		for _, e := range idMatches {
			if e.ID != entityID {
				result.MatchedEntities = append(result.MatchedEntities, e)
			}
		}
	}

	// 3. 获取关联实体
	relations, _ := s.entityRepo.GetRelations(ctx, entityID, true)
	for _, rel := range relations {
		var relatedID uint
		if rel.SourceEntityID == entityID {
			relatedID = rel.TargetEntityID
		} else {
			relatedID = rel.SourceEntityID
		}
		related, _ := s.entityRepo.GetByID(ctx, relatedID)
		if related != nil {
			result.RelatedEntities = append(result.RelatedEntities, related)
		}
	}

	// 统计
	result.TotalMatches = len(result.NameMatches) + len(result.MatchedEntities)
	result.HighRiskCount = len(result.MatchedEntities) // 证件号匹配为高风险
	result.MediumRiskCount = len(result.NameMatches)   // 名称匹配为中风险

	return result, nil
}
