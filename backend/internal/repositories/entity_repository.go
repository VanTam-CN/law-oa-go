package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
	"law-oa-go/internal/security"
)

// RelatedEntityNode 递归关联穿透结果节点
type RelatedEntityNode struct {
	EntityID     uint                `json:"entity_id"`
	Depth        int                 `json:"depth"` // 从起始实体算起的层级（1=直接关联）
	RelationType models.RelationType `json:"relation_type"`
	Direction    string              `json:"direction"` // "outgoing" 源→目标, "incoming" 目标→源
	Entity       *models.Entity      `json:"entity"`
	Path         []RelatedEdge       `json:"path"` // 从起始实体到当前节点的路径
}

// RelatedEdge 关联路径中的一条边
type RelatedEdge struct {
	FromID       uint                `json:"from_id"`
	ToID         uint                `json:"to_id"`
	RelationType models.RelationType `json:"relation_type"`
}

// EntityRepository 实体仓储接口
type EntityRepository interface {
	// CRUD 操作
	Create(ctx context.Context, entity *models.Entity) error
	GetByID(ctx context.Context, id uint) (*models.Entity, error)
	Update(ctx context.Context, id uint, updates map[string]interface{}) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int, conditions map[string]interface{}) ([]*models.Entity, int64, error)

	// 搜索功能
	SearchByName(ctx context.Context, name string, limit int) ([]*models.Entity, error)
	SearchByIdentityNumber(ctx context.Context, identityNumber string, limit int) ([]*models.Entity, error)
	SearchByAlias(ctx context.Context, alias string, limit int) ([]*models.Entity, error)
	SearchByFormerName(ctx context.Context, name string, limit int) ([]*models.Entity, error)

	// 关系查询
	GetRelations(ctx context.Context, entityID uint, activeOnly bool) ([]*models.EntityRelation, error)
	GetRelatedEntities(ctx context.Context, entityID uint, relationType models.RelationType, depth int) ([]*models.Entity, error)
	GetRelatedEntitiesRecursive(ctx context.Context, entityID uint, maxDepth int) ([]*RelatedEntityNode, error)

	// 名称历史
	GetNameHistory(ctx context.Context, entityID uint) ([]*models.EntityNameHistory, error)
	AddNameHistory(ctx context.Context, history *models.EntityNameHistory) error

	// 案件当事人
	GetCaseParties(ctx context.Context, caseID uint) ([]*models.CaseParty, error)
	AddCaseParty(ctx context.Context, party *models.CaseParty) error
	RemoveCaseParty(ctx context.Context, caseID, entityID uint) error

	// 批量操作
	BatchCreate(ctx context.Context, entities []*models.Entity) error
	BatchCreateRelations(ctx context.Context, relations []*models.EntityRelation) error

	// 高级搜索
	AdvancedSearch(ctx context.Context, params *EntitySearchParams) ([]*models.Entity, error)

	// 获取实体关联的活跃案件
	GetActiveCasesByEntity(ctx context.Context, entityID uint) ([]*models.Case, error)
	// 获取实体关联的全部历史案件（含已结案、已拒绝和已撤案记录）
	GetAllCasesByEntity(ctx context.Context, entityID uint) ([]*models.Case, error)
}

// entityRepository 实体仓储实现
type entityRepository struct {
	db *gorm.DB
}

// NewEntityRepository 创建实体仓储
func NewEntityRepository(db *gorm.DB) EntityRepository {
	return &entityRepository{db: db}
}

// Create 创建实体
func (r *entityRepository) Create(ctx context.Context, entity *models.Entity) error {
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		return fmt.Errorf("创建实体失败: %w", err)
	}
	return nil
}

// GetByID 根据ID获取实体
func (r *entityRepository) GetByID(ctx context.Context, id uint) (*models.Entity, error) {
	var entity models.Entity
	err := r.db.WithContext(ctx).
		Preload("Relations").
		Preload("NameHistory").
		First(&entity, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrRecordNotFound
		}
		return nil, fmt.Errorf("获取实体失败: %w", err)
	}
	return &entity, nil
}

// Update 更新实体
func (r *entityRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	prepared, err := prepareEntityUpdates(updates)
	if err != nil {
		return err
	}
	result := r.db.WithContext(ctx).Model(&models.Entity{}).Where("id = ?", id).Updates(prepared)
	if result.Error != nil {
		return fmt.Errorf("更新实体失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// Delete 删除实体
func (r *entityRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Entity{}, id)
	if result.Error != nil {
		return fmt.Errorf("删除实体失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// List 列表查询
func (r *entityRepository) List(ctx context.Context, offset, limit int, conditions map[string]interface{}) ([]*models.Entity, int64, error) {
	var entities []*models.Entity
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Entity{})

	// 应用条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取实体总数失败: %w", err)
	}

	// 获取数据
	if err := query.Offset(offset).Limit(limit).Find(&entities).Error; err != nil {
		return nil, 0, fmt.Errorf("获取实体列表失败: %w", err)
	}

	return entities, total, nil
}

// SearchByName 按名称搜索。LOWER + LIKE 在 MySQL、PostgreSQL 和 SQLite 均可用。
func (r *entityRepository) SearchByName(ctx context.Context, name string, limit int) ([]*models.Entity, error) {
	var entities []*models.Entity

	query := r.db.WithContext(ctx).
		Where("LOWER(name) LIKE LOWER(?) OR LOWER(alias) LIKE LOWER(?)", "%"+name+"%", "%"+name+"%")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("按名称搜索实体失败: %w", err)
	}

	return entities, nil
}

// SearchByIdentityNumber 按证件号搜索
func (r *entityRepository) SearchByIdentityNumber(ctx context.Context, identityNumber string, limit int) ([]*models.Entity, error) {
	var entities []*models.Entity
	digest, err := security.IdentityDigest(identityNumber)
	if err != nil {
		return nil, fmt.Errorf("按证件号搜索实体失败: %w", err)
	}

	query := r.db.WithContext(ctx).
		Where("identity_number_digest = ?", digest)

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("按证件号搜索实体失败: %w", err)
	}

	return entities, nil
}

// SearchByAlias 按别名搜索
func (r *entityRepository) SearchByAlias(ctx context.Context, alias string, limit int) ([]*models.Entity, error) {
	var entities []*models.Entity

	query := r.db.WithContext(ctx).
		Where("LOWER(alias) LIKE LOWER(?)", "%"+alias+"%")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("按别名搜索实体失败: %w", err)
	}

	return entities, nil
}

// SearchByFormerName 按曾用名搜索
func (r *entityRepository) SearchByFormerName(ctx context.Context, name string, limit int) ([]*models.Entity, error) {
	var entities []*models.Entity

	query := r.db.WithContext(ctx).
		Joins("JOIN entity_name_history ON entity_name_history.entity_id = entities.id").
		Where("LOWER(entity_name_history.old_name) LIKE LOWER(?)", "%"+name+"%").
		Group("entities.id")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("按曾用名搜索实体失败: %w", err)
	}

	return entities, nil
}

// GetRelations 获取实体关联关系
func (r *entityRepository) GetRelations(ctx context.Context, entityID uint, activeOnly bool) ([]*models.EntityRelation, error) {
	var relations []*models.EntityRelation

	query := r.db.WithContext(ctx).
		Preload("SourceEntity").
		Preload("TargetEntity").
		Where("source_entity_id = ? OR target_entity_id = ?", entityID, entityID)

	if activeOnly {
		now := time.Now()
		query = query.Where("is_active = ? AND (end_date IS NULL OR end_date > ?)", true, now)
	}

	if err := query.Find(&relations).Error; err != nil {
		return nil, fmt.Errorf("获取实体关联关系失败: %w", err)
	}

	return relations, nil
}

// GetRelatedEntities 递归获取关联实体
func (r *entityRepository) GetRelatedEntities(ctx context.Context, entityID uint, relationType models.RelationType, depth int) ([]*models.Entity, error) {
	if depth <= 0 {
		return []*models.Entity{}, nil
	}

	// 获取直接关联的实体ID
	var relatedIDs []uint
	query := r.db.WithContext(ctx).Model(&models.EntityRelation{}).
		Where("source_entity_id = ? AND relation_type = ? AND is_active = ?", entityID, relationType, true).
		Pluck("target_entity_id", &relatedIDs)

	if query.Error != nil {
		return nil, fmt.Errorf("获取关联实体ID失败: %w", query.Error)
	}

	if len(relatedIDs) == 0 {
		return []*models.Entity{}, nil
	}

	// 获取实体详情
	var entities []*models.Entity
	if err := r.db.WithContext(ctx).Where("id IN ?", relatedIDs).Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("获取关联实体详情失败: %w", err)
	}

	// 递归获取更深层的关联
	for _, entity := range entities {
		deeperEntities, err := r.GetRelatedEntities(ctx, entity.ID, relationType, depth-1)
		if err != nil {
			continue
		}
		entities = append(entities, deeperEntities...)
	}

	return entities, nil
}

// GetRelatedEntitiesRecursive 递归穿透获取关联实体（PostgreSQL Recursive CTE）
// 双向遍历所有关系类型，支持最大深度限制，返回完整路径信息
func (r *entityRepository) GetRelatedEntitiesRecursive(ctx context.Context, entityID uint, maxDepth int) ([]*RelatedEntityNode, error) {
	if maxDepth <= 0 {
		maxDepth = 3
	}
	if maxDepth > 5 {
		maxDepth = 5
	}

	// PostgreSQL Recursive CTE: 双向遍历 entity_relations，带路径追踪和环路检测
	cteSQL := `
	WITH RECURSIVE relation_graph AS (
		-- 基础查询: 获取起始实体的直接关联（双向）
		SELECT
			er.source_entity_id AS from_id,
			er.target_entity_id AS to_id,
			er.relation_type,
			1 AS depth,
			ARRAY[er.source_entity_id, er.target_entity_id] AS path_ids
		FROM entity_relations er
		WHERE er.deleted_at IS NULL AND er.is_active = true
		  AND (
		      (er.source_entity_id = ? AND er.target_entity_id != ?)
		      OR
		      (er.target_entity_id = ? AND er.source_entity_id != ?)
		  )

		UNION

		-- 递归查询: 沿着关联关系继续穿透
		SELECT
			CASE
				WHEN er.source_entity_id = rg.to_id THEN er.source_entity_id
				ELSE er.target_entity_id
			END AS from_id,
			CASE
				WHEN er.source_entity_id = rg.to_id THEN er.target_entity_id
				ELSE er.source_entity_id
			END AS to_id,
			er.relation_type,
			rg.depth + 1 AS depth,
			rg.path_ids || CASE
				WHEN er.source_entity_id = rg.to_id THEN er.target_entity_id
				ELSE er.source_entity_id
			END AS path_ids
		FROM entity_relations er
		JOIN relation_graph rg ON (
			er.source_entity_id = rg.to_id OR er.target_entity_id = rg.to_id
		)
		WHERE er.deleted_at IS NULL AND er.is_active = true
		  AND rg.depth < ?
		  AND NOT (
		      CASE
		          WHEN er.source_entity_id = rg.to_id THEN er.target_entity_id
		          ELSE er.source_entity_id
		      END = ANY(rg.path_ids)
		  )
	)
	SELECT DISTINCT ON (to_id, relation_type)
		from_id, to_id, relation_type, depth, path_ids
	FROM relation_graph
	ORDER BY to_id, relation_type, depth ASC
	`

	type cteRow struct {
		FromID       uint                `gorm:"column:from_id"`
		ToID         uint                `gorm:"column:to_id"`
		RelationType models.RelationType `gorm:"column:relation_type"`
		Depth        int                 `gorm:"column:depth"`
		PathIDs      []uint              `gorm:"column:path_ids;type:jsonb"`
	}

	var rows []cteRow
	err := r.db.WithContext(ctx).Raw(cteSQL, entityID, entityID, entityID, entityID, maxDepth).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("递归穿透查询失败: %w", err)
	}

	if len(rows) == 0 {
		return []*RelatedEntityNode{}, nil
	}

	// 收集所有关联实体ID，批量查询实体详情
	entityIDSet := make(map[uint]bool)
	for _, row := range rows {
		entityIDSet[row.ToID] = true
	}
	entityIDs := make([]uint, 0, len(entityIDSet))
	for id := range entityIDSet {
		entityIDs = append(entityIDs, id)
	}

	var entities []models.Entity
	if err := r.db.WithContext(ctx).Where("id IN ?", entityIDs).Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("批量获取实体详情失败: %w", err)
	}

	entityMap := make(map[uint]*models.Entity)
	for i := range entities {
		entityMap[entities[i].ID] = &entities[i]
	}

	// 构建结果节点
	nodes := make([]*RelatedEntityNode, 0, len(rows))
	for _, row := range rows {
		entity, ok := entityMap[row.ToID]
		if !ok {
			continue
		}

		// 构建路径边
		path := make([]RelatedEdge, 0, len(row.PathIDs)-1)
		for i := 0; i < len(row.PathIDs)-1; i++ {
			path = append(path, RelatedEdge{
				FromID:       row.PathIDs[i],
				ToID:         row.PathIDs[i+1],
				RelationType: row.RelationType,
			})
		}

		direction := "outgoing"
		if row.FromID == row.ToID {
			direction = "incoming"
		}

		nodes = append(nodes, &RelatedEntityNode{
			EntityID:     row.ToID,
			Depth:        row.Depth,
			RelationType: row.RelationType,
			Direction:    direction,
			Entity:       entity,
			Path:         path,
		})
	}

	return nodes, nil
}

// GetNameHistory 获取名称变更历史
func (r *entityRepository) GetNameHistory(ctx context.Context, entityID uint) ([]*models.EntityNameHistory, error) {
	var history []*models.EntityNameHistory

	if err := r.db.WithContext(ctx).
		Where("entity_id = ?", entityID).
		Order("change_date DESC").
		Find(&history).Error; err != nil {
		return nil, fmt.Errorf("获取名称历史失败: %w", err)
	}

	return history, nil
}

// AddNameHistory 添加名称变更记录
func (r *entityRepository) AddNameHistory(ctx context.Context, history *models.EntityNameHistory) error {
	if err := r.db.WithContext(ctx).Create(history).Error; err != nil {
		return fmt.Errorf("添加名称历史失败: %w", err)
	}
	return nil
}

// GetCaseParties 获取案件当事人
func (r *entityRepository) GetCaseParties(ctx context.Context, caseID uint) ([]*models.CaseParty, error) {
	var parties []*models.CaseParty

	if err := r.db.WithContext(ctx).
		Preload("Entity").
		Where("case_id = ?", caseID).
		Order("display_order ASC").
		Find(&parties).Error; err != nil {
		return nil, fmt.Errorf("获取案件当事人失败: %w", err)
	}

	return parties, nil
}

// AddCaseParty 添加案件当事人
func (r *entityRepository) AddCaseParty(ctx context.Context, party *models.CaseParty) error {
	if err := r.db.WithContext(ctx).Create(party).Error; err != nil {
		return fmt.Errorf("添加案件当事人失败: %w", err)
	}
	return nil
}

// RemoveCaseParty 移除案件当事人
func (r *entityRepository) RemoveCaseParty(ctx context.Context, caseID, entityID uint) error {
	result := r.db.WithContext(ctx).
		Where("case_id = ? AND entity_id = ?", caseID, entityID).
		Delete(&models.CaseParty{})

	if result.Error != nil {
		return fmt.Errorf("移除案件当事人失败: %w", result.Error)
	}

	return nil
}

// BatchCreate 批量创建实体
func (r *entityRepository) BatchCreate(ctx context.Context, entities []*models.Entity) error {
	if len(entities) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).CreateInBatches(entities, 100).Error; err != nil {
		return fmt.Errorf("批量创建实体失败: %w", err)
	}

	return nil
}

// BatchCreateRelations 批量创建关联关系
func (r *entityRepository) BatchCreateRelations(ctx context.Context, relations []*models.EntityRelation) error {
	if len(relations) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).CreateInBatches(relations, 100).Error; err != nil {
		return fmt.Errorf("批量创建关联关系失败: %w", err)
	}

	return nil
}

func prepareEntityUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	prepared := make(map[string]interface{}, len(updates)+2)
	for key, value := range updates {
		prepared[key] = value
	}
	rawValue, ok := prepared["identity_number"]
	if !ok {
		rawValue, ok = prepared["identityNumber"]
		if ok {
			delete(prepared, "identityNumber")
		}
	}
	if !ok {
		return prepared, nil
	}
	raw, ok := rawValue.(string)
	if !ok {
		return nil, fmt.Errorf("更新实体失败: identity_number必须是字符串")
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		prepared["identity_number"] = ""
		prepared["identity_number_digest"] = ""
		prepared["identity_number_ciphertext"] = ""
		return prepared, nil
	}
	ciphertext, digest, err := security.ProtectIdentityNumber(raw)
	if err != nil {
		return nil, fmt.Errorf("更新实体失败: %w", err)
	}
	prepared["identity_number"] = ""
	prepared["identity_number_digest"] = digest
	prepared["identity_number_ciphertext"] = ciphertext
	return prepared, nil
}

// GetActiveCasesByEntity 获取实体关联的活跃案件
func (r *entityRepository) GetActiveCasesByEntity(ctx context.Context, entityID uint) ([]*models.Case, error) {
	var cases []*models.Case

	query := r.db.WithContext(ctx).
		Joins("JOIN case_parties ON case_parties.case_id = cases.id").
		Where("case_parties.entity_id = ?", entityID).
		Where("cases.status IN ?", []string{"ACTIVE", "IN_PROGRESS", "PENDING"}).
		Distinct("cases.*")

	if err := query.Find(&cases).Error; err != nil {
		return nil, fmt.Errorf("获取实体关联案件失败: %w", err)
	}

	return cases, nil
}

// GetAllCasesByEntity returns every non-deleted matter in which the entity is
// a party. Conflict checks must use the full historical archive; limiting this
// query to active matters would create a systematic false-clear result for
// former clients and previously rejected matters.
func (r *entityRepository) GetAllCasesByEntity(ctx context.Context, entityID uint) ([]*models.Case, error) {
	var cases []*models.Case

	query := r.db.WithContext(ctx).
		Joins("JOIN case_parties ON case_parties.case_id = cases.id").
		Where("case_parties.entity_id = ?", entityID).
		Where("cases.deleted_at IS NULL").
		Distinct("cases.*")

	if err := query.Order("cases.created_at DESC").Find(&cases).Error; err != nil {
		return nil, fmt.Errorf("获取实体全部历史案件失败: %w", err)
	}

	return cases, nil
}

// ConflictCheckRepository 冲突检查仓储接口
type ConflictCheckRepository interface {
	// 冲突检查记录 CRUD
	CreateConflictCheck(ctx context.Context, check *models.ConflictCheck) error
	GetConflictCheck(ctx context.Context, id uint) (*models.ConflictCheck, error)
	ListConflictChecks(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]*models.ConflictCheck, int64, error)
	UpdateConflictCheckStatus(ctx context.Context, id uint, status string, result *models.CheckResult) error
	DeleteConflictCheck(ctx context.Context, id uint) error

	// 冲突详情 CRUD
	CreateConflictDetail(ctx context.Context, detail *models.ConflictDetail) error
	GetConflictDetails(ctx context.Context, checkID uint) ([]*models.ConflictDetail, error)
	UpdateConflictDetail(ctx context.Context, id uint, updates map[string]interface{}) error
	DeleteConflictDetail(ctx context.Context, id uint) error

	// 批量操作
	BatchCreateDetails(ctx context.Context, details []*models.ConflictDetail) error

	// 冲突检测辅助方法
	FindConflictingEntities(ctx context.Context, caseID uint, excludeEntityIDs []uint) ([]*models.Entity, error)
	FindEntitiesByCaseIDs(ctx context.Context, caseIDs []uint) ([]*models.Entity, error)
}

// conflictCheckRepository 冲突检查仓储实现
type conflictCheckRepository struct {
	db *gorm.DB
}

// NewConflictCheckRepository 创建冲突检查仓储
func NewConflictCheckRepository(db *gorm.DB) ConflictCheckRepository {
	return &conflictCheckRepository{db: db}
}

// CreateConflictCheck 创建冲突检查记录
func (r *conflictCheckRepository) CreateConflictCheck(ctx context.Context, check *models.ConflictCheck) error {
	if err := r.db.WithContext(ctx).Create(check).Error; err != nil {
		return fmt.Errorf("创建冲突检查记录失败: %w", err)
	}
	return nil
}

// GetConflictCheck 获取冲突检查记录
func (r *conflictCheckRepository) GetConflictCheck(ctx context.Context, id uint) (*models.ConflictCheck, error) {
	var check models.ConflictCheck
	err := r.db.WithContext(ctx).
		Preload("ConflictDetails").
		Preload("ConflictDetails.MatchedEntity").
		Preload("ConflictDetails.MatchedCase").
		First(&check, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrRecordNotFound
		}
		return nil, fmt.Errorf("获取冲突检查记录失败: %w", err)
	}
	return &check, nil
}

// ListConflictChecks 列表查询冲突检查记录
func (r *conflictCheckRepository) ListConflictChecks(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]*models.ConflictCheck, int64, error) {
	var checks []*models.ConflictCheck
	var total int64

	query := r.db.WithContext(ctx).Model(&models.ConflictCheck{})
	if rawUserID, ok := filters["ethical_wall_user_id"]; ok {
		userID, valid := rawUserID.(uint)
		if !valid || userID == 0 {
			return nil, 0, fmt.Errorf("隔离墙查询用户无效")
		}
		query = query.Where(`conflict_checks.case_id IN (
			SELECT visible_cases.id FROM cases visible_cases
			WHERE visible_cases.deleted_at IS NULL
			  AND (
				  visible_cases.ethical_wall_enabled = ?
				  OR EXISTS (
					  SELECT 1 FROM case_ethical_wall_whitelist wall_access
					  WHERE wall_access.case_id = visible_cases.id
						AND wall_access.user_id = ?
				  )
			  )
		)`, false, userID)
	}

	// 应用过滤条件
	for key, value := range filters {
		if key == "ethical_wall_user_id" {
			continue
		}
		query = query.Where(key, value)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取冲突检查记录总数失败: %w", err)
	}

	// 获取数据
	if err := query.
		Preload("ConflictDetails").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&checks).Error; err != nil {
		return nil, 0, fmt.Errorf("获取冲突检查记录列表失败: %w", err)
	}

	return checks, total, nil
}

// UpdateConflictCheckStatus 更新冲突检查状态
func (r *conflictCheckRepository) UpdateConflictCheckStatus(ctx context.Context, id uint, status string, result *models.CheckResult) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if result != nil {
		updates["result"] = result
	}

	resultCheck := r.db.WithContext(ctx).Model(&models.ConflictCheck{}).Where("id = ?", id).Updates(updates)
	if resultCheck.Error != nil {
		return fmt.Errorf("更新冲突检查状态失败: %w", resultCheck.Error)
	}
	if resultCheck.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// DeleteConflictCheck 删除冲突检查记录
func (r *conflictCheckRepository) DeleteConflictCheck(ctx context.Context, id uint) error {
	return fmt.Errorf("冲突检查记录为审计证据，不允许删除；请通过追加更正或撤销记录处理")
}

// CreateConflictDetail 创建冲突详情
func (r *conflictCheckRepository) CreateConflictDetail(ctx context.Context, detail *models.ConflictDetail) error {
	if err := r.db.WithContext(ctx).Create(detail).Error; err != nil {
		return fmt.Errorf("创建冲突详情失败: %w", err)
	}
	return nil
}

// GetConflictDetails 获取冲突检查的所有详情
func (r *conflictCheckRepository) GetConflictDetails(ctx context.Context, checkID uint) ([]*models.ConflictDetail, error) {
	var details []*models.ConflictDetail

	if err := r.db.WithContext(ctx).
		Preload("MatchedEntity").
		Preload("MatchedCase").
		Where("conflict_check_id = ?", checkID).
		Order("created_at ASC").
		Find(&details).Error; err != nil {
		return nil, fmt.Errorf("获取冲突详情失败: %w", err)
	}

	return details, nil
}

// UpdateConflictDetail 更新冲突详情
func (r *conflictCheckRepository) UpdateConflictDetail(ctx context.Context, id uint, updates map[string]interface{}) error {
	return fmt.Errorf("冲突详情为审计证据，不允许更新；请追加新的证据或更正记录")
}

// DeleteConflictDetail 删除冲突详情
func (r *conflictCheckRepository) DeleteConflictDetail(ctx context.Context, id uint) error {
	return fmt.Errorf("冲突详情为审计证据，不允许删除；请追加更正记录")
}

// BatchCreateDetails 批量创建冲突详情
func (r *conflictCheckRepository) BatchCreateDetails(ctx context.Context, details []*models.ConflictDetail) error {
	if len(details) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).CreateInBatches(details, 100).Error; err != nil {
		return fmt.Errorf("批量创建冲突详情失败: %w", err)
	}
	return nil
}

// FindConflictingEntities 查找可能冲突的实体
func (r *conflictCheckRepository) FindConflictingEntities(ctx context.Context, caseID uint, excludeEntityIDs []uint) ([]*models.Entity, error) {
	var entities []*models.Entity

	// 查找与该案件相关的其他案件中的实体
	query := r.db.WithContext(ctx).
		Joins("JOIN case_parties ON case_parties.entity_id = entities.id").
		Joins("JOIN cases ON case_parties.case_id = cases.id").
		Where("cases.id != ?", caseID)

	if len(excludeEntityIDs) > 0 {
		query = query.Where("entities.id NOT IN ?", excludeEntityIDs)
	}

	if err := query.Distinct("entities.*").Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("查找冲突实体失败: %w", err)
	}

	return entities, nil
}

// FindEntitiesByCaseIDs 根据案件ID列表查找实体
func (r *conflictCheckRepository) FindEntitiesByCaseIDs(ctx context.Context, caseIDs []uint) ([]*models.Entity, error) {
	var entities []*models.Entity

	if len(caseIDs) == 0 {
		return entities, nil
	}

	query := r.db.WithContext(ctx).
		Joins("JOIN case_parties ON case_parties.entity_id = entities.id").
		Where("case_parties.case_id IN ?", caseIDs)

	if err := query.Distinct("entities.*").Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("根据案件查找实体失败: %w", err)
	}

	return entities, nil
}

// EntitySearchParams 实体搜索参数
type EntitySearchParams struct {
	Name           string
	EntityType     string
	Status         string
	IdentityNumber string
	Alias          string
	Limit          int
}

// AdvancedSearch 高级搜索
func (r *entityRepository) AdvancedSearch(ctx context.Context, params *EntitySearchParams) ([]*models.Entity, error) {
	var entities []*models.Entity
	query := r.db.WithContext(ctx)

	if params.Name != "" {
		query = query.Where("LOWER(name) LIKE LOWER(?) OR LOWER(alias) LIKE LOWER(?)", "%"+params.Name+"%", "%"+params.Name+"%")
	}
	if params.IdentityNumber != "" {
		digest, err := security.IdentityDigest(params.IdentityNumber)
		if err != nil {
			return nil, fmt.Errorf("高级搜索失败: %w", err)
		}
		query = query.Where("identity_number_digest = ?", digest)
	}
	if params.Alias != "" {
		query = query.Where("LOWER(alias) LIKE LOWER(?)", "%"+params.Alias+"%")
	}
	if params.EntityType != "" {
		query = query.Where("entity_type = ?", params.EntityType)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}

	if err := query.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("高级搜索失败: %w", err)
	}

	return entities, nil
}
