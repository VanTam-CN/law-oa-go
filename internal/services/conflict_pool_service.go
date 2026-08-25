package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// ConflictPoolService 冲突池服务接口
type ConflictPoolService interface {
	// SyncLawyerPool 同步律师冲突池数据
	SyncLawyerPool(ctx context.Context, lawyerID uint, caseID uint) error
	// SyncAllLawyersPool 同步所有律师的冲突池数据
	SyncAllLawyersPool(ctx context.Context) (*SyncResult, error)
	// GetConflictPool 获取律师冲突池
	GetConflictPool(ctx context.Context, lawyerID uint) ([]*models.LawyerConflictPool, error)
	// SearchInPool 在冲突池中搜索
	SearchInPool(ctx context.Context, req *PoolSearchRequest) ([]*PoolMatchResult, error)
	// UpdatePoolEntry 更新冲突池条目
	UpdatePoolEntry(ctx context.Context, entryID uint, update *PoolUpdateRequest) error
	// DeletePoolEntry 删除冲突池条目
	DeletePoolEntry(ctx context.Context, entryID uint) error
	// RefreshFromAPI 从 API 刷新公司数据
	RefreshFromAPI(ctx context.Context, entryID uint, provider CompanyAPIProvider) error
}

// PoolSearchRequest 池搜索请求
type PoolSearchRequest struct {
	SearchTerm       string `json:"searchTerm"`       // 搜索词（公司名称/税号）
	SearchType       string `json:"searchType"`       // 搜索类型: standard/fuzzy
	IncludeAliases   bool   `json:"includeAliases"`   // 是否包含别名
	LawyerID         *uint  `json:"lawyerId"`         // 可选：限定律师ID
	RelationshipType string `json:"relationshipType"` // 可选：限定关系类型
}

// PoolMatchResult 池匹配结果
type PoolMatchResult struct {
	PoolEntry    *models.LawyerConflictPool `json:"poolEntry"`
	MatchScore   float64                    `json:"matchScore"`   // 匹配分数
	MatchReason  string                     `json:"matchReason"`  // 匹配原因
	ConflictType string                     `json:"conflictType"` // 冲突类型
	RiskLevel    string                     `json:"riskLevel"`    // 风险等级
}

// PoolUpdateRequest 池更新请求
type PoolUpdateRequest struct {
	RelationshipType string      `json:"relationshipType"`
	ShareholdingInfo models.JSON `json:"shareholdingInfo"`
	RelatedCompanies models.JSON `json:"relatedCompanies"`
	EntityAliases    models.JSON `json:"entityAliases"`
	DataSource       string      `json:"dataSource"`
}

// SyncResult 同步结果
type SyncResult struct {
	TotalLawyers   int       `json:"totalLawyers"`
	ProcessedCases int       `json:"processedCases"`
	AddedEntries   int       `json:"addedEntries"`
	UpdatedEntries int       `json:"updatedEntries"`
	FailedEntries  int       `json:"failedEntries"`
	Errors         []string  `json:"errors"`
	StartTime      time.Time `json:"startTime"`
	EndTime        time.Time `json:"endTime"`
	DurationMs     int64     `json:"durationMs"`
}

// conflictPoolService 冲突池服务实现
type conflictPoolService struct {
	db         *gorm.DB
	companyAPI CompanyAPIService
	caseRepo   CaseRepository
	clientRepo ClientRepository
}

// CaseRepository 案件仓库接口
type CaseRepository interface {
	GetByID(ctx context.Context, id uint) (*models.Case, error)
	ListByLawyer(ctx context.Context, lawyerID uint) ([]*models.Case, error)
	GetDB() *gorm.DB
}

// ClientRepository 客户仓库接口
type ClientRepository interface {
	FindByID(ctx context.Context, id uint) (*models.Client, error)
}

// NewConflictPoolService 创建新的冲突池服务
func NewConflictPoolService(
	db *gorm.DB,
	companyAPI CompanyAPIService,
	caseRepo CaseRepository,
	clientRepo ClientRepository,
) ConflictPoolService {
	return &conflictPoolService{
		db:         db,
		companyAPI: companyAPI,
		caseRepo:   caseRepo,
		clientRepo: clientRepo,
	}
}

// SyncLawyerPool 同步律师冲突池数据
func (s *conflictPoolService) SyncLawyerPool(ctx context.Context, lawyerID uint, caseID uint) error {
	log.Printf("🔄 开始同步律师冲突池: lawyerID=%d, caseID=%d", lawyerID, caseID)

	// 获取案件信息
	case_, err := s.caseRepo.GetByID(ctx, caseID)
	if err != nil {
		return fmt.Errorf("获取案件失败: %w", err)
	}

	// 获取客户信息
	client, err := s.clientRepo.FindByID(ctx, case_.ClientID)
	if err != nil {
		return fmt.Errorf("获取客户失败: %w", err)
	}

	// 判断客户类型
	entityType := "individual"
	if client.Type == "企业" || strings.Contains(client.Name, "公司") ||
		strings.Contains(client.Name, "有限") || strings.Contains(client.Name, "集团") {
		entityType = "company"
	}

	// 标准化公司名称
	standardName := s.standardizeEntityName(client.Name)

	// 检查是否已存在
	var existingPool models.LawyerConflictPool
	err = s.db.WithContext(ctx).
		Where("lawyer_id = ? AND case_id = ? AND entity_name_standard = ?",
			lawyerID, caseID, standardName).
		First(&existingPool).Error

	now := time.Now()

	if err == nil {
		// 更新现有记录
		existingPool.EntityName = client.Name
		existingPool.EntityNameStandard = standardName
		existingPool.RelationshipType = s.determineRelationshipType(case_)
		existingPool.CaseTitle = case_.Title
		existingPool.UpdatedAt = now

		if err := s.db.WithContext(ctx).Save(&existingPool).Error; err != nil {
			return fmt.Errorf("更新冲突池记录失败: %w", err)
		}

		log.Printf("✅ 更新冲突池记录: ID=%d", existingPool.ID)
	} else if err == gorm.ErrRecordNotFound {
		// 创建新记录
		identityNumber, _ := client.DecryptedIdentity()
		newEntry := &models.LawyerConflictPool{
			LawyerID:           lawyerID,
			EntityType:         entityType,
			EntityName:         client.Name,
			EntityNameStandard: standardName,
			EntityTaxID:        identityNumber,
			RelationshipType:   s.determineRelationshipType(case_),
			CaseID:             caseID,
			CaseTitle:          case_.Title,
			DataSource:         "manual",
			LastVerifiedAt:     &now,
		}

		if err := s.db.WithContext(ctx).Create(newEntry).Error; err != nil {
			return fmt.Errorf("创建冲突池记录失败: %w", err)
		}

		log.Printf("✅ 创建冲突池记录: ID=%d", newEntry.ID)

		// 对于企业客户，异步获取详细信息
		if entityType == "company" {
			go s.enrichCompanyData(context.Background(), newEntry.ID, client.Name, identityNumber)
		}
	}

	return nil
}

// SyncAllLawyersPool 同步所有律师的冲突池数据
func (s *conflictPoolService) SyncAllLawyersPool(ctx context.Context) (*SyncResult, error) {
	log.Printf("🔄 开始同步所有律师冲突池")
	startTime := time.Now()

	result := &SyncResult{
		StartTime: startTime,
		Errors:    make([]string, 0),
	}

	// 获取所有有案件的律师
	var lawyerIDs []uint
	if err := s.db.WithContext(ctx).
		Model(&models.Case{}).
		Select("DISTINCT lawyer_id").
		Pluck("lawyer_id", &lawyerIDs).Error; err != nil {
		return nil, fmt.Errorf("获取律师列表失败: %w", err)
	}

	result.TotalLawyers = len(lawyerIDs)

	// 遍历每个律师
	for _, lawyerID := range lawyerIDs {
		// 获取律师的所有案件
		cases, err := s.caseRepo.ListByLawyer(ctx, lawyerID)
		if err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("律师 %d 获取案件失败: %v", lawyerID, err))
			result.FailedEntries++
			continue
		}

		// 处理每个案件
		for _, case_ := range cases {
			if err := s.SyncLawyerPool(ctx, lawyerID, case_.ID); err != nil {
				result.Errors = append(result.Errors,
					fmt.Sprintf("案件 %d 处理失败: %v", case_.ID, err))
				result.FailedEntries++
			} else {
				result.ProcessedCases++
			}
		}
	}

	result.EndTime = time.Now()
	result.DurationMs = result.EndTime.Sub(startTime).Milliseconds()

	log.Printf("✅ 同步完成: 处理案件=%d, 耗时=%dms",
		result.ProcessedCases, result.DurationMs)

	return result, nil
}

// GetConflictPool 获取律师冲突池
func (s *conflictPoolService) GetConflictPool(ctx context.Context, lawyerID uint) ([]*models.LawyerConflictPool, error) {
	var pool []*models.LawyerConflictPool

	if err := s.db.WithContext(ctx).
		Where("lawyer_id = ?", lawyerID).
		Order("created_at DESC").
		Find(&pool).Error; err != nil {
		return nil, fmt.Errorf("获取冲突池失败: %w", err)
	}

	return pool, nil
}

// SearchInPool 在冲突池中搜索
func (s *conflictPoolService) SearchInPool(ctx context.Context, req *PoolSearchRequest) ([]*PoolMatchResult, error) {
	log.Printf("🔍 在冲突池中搜索: term=%s, type=%s", req.SearchTerm, req.SearchType)

	// 构建查询
	query := s.db.WithContext(ctx).Model(&models.LawyerConflictPool{})

	// 标准化搜索词
	standardSearchTerm := s.standardizeEntityName(req.SearchTerm)

	if req.SearchType == "fuzzy" {
		// 模糊搜索
		query = query.Where(
			"entity_name_standard LIKE ? OR entity_name LIKE ? OR entity_tax_id = ?",
			"%"+standardSearchTerm+"%", "%"+req.SearchTerm+"%", req.SearchTerm,
		)
	} else {
		// 精确搜索（优先）
		query = query.Where(
			"entity_name_standard = ? OR entity_name = ? OR entity_tax_id = ?",
			standardSearchTerm, req.SearchTerm, req.SearchTerm,
		)
	}

	// 添加过滤条件
	if req.LawyerID != nil {
		query = query.Where("lawyer_id = ?", *req.LawyerID)
	}
	if req.RelationshipType != "" {
		query = query.Where("relationship_type = ?", req.RelationshipType)
	}

	// 执行查询
	var poolEntries []*models.LawyerConflictPool
	if err := query.Find(&poolEntries).Error; err != nil {
		return nil, fmt.Errorf("搜索失败: %w", err)
	}

	// 构建匹配结果
	results := make([]*PoolMatchResult, 0, len(poolEntries))
	for _, entry := range poolEntries {
		matchScore := s.calculateMatchScore(req.SearchTerm, entry, req.SearchType)
		conflictType := s.determineConflictType(entry)
		riskLevel := s.assessRiskLevel(entry, conflictType)

		results = append(results, &PoolMatchResult{
			PoolEntry:    entry,
			MatchScore:   matchScore,
			MatchReason:  s.getMatchReason(entry, matchScore),
			ConflictType: conflictType,
			RiskLevel:    riskLevel,
		})
	}

	// 按匹配分数排序
	s.sortResults(results)

	log.Printf("✅ 搜索完成: 找到 %d 条匹配结果", len(results))
	return results, nil
}

// UpdatePoolEntry 更新冲突池条目
func (s *conflictPoolService) UpdatePoolEntry(ctx context.Context, entryID uint, update *PoolUpdateRequest) error {
	var entry models.LawyerConflictPool
	if err := s.db.WithContext(ctx).First(&entry, entryID).Error; err != nil {
		return fmt.Errorf("找不到冲突池条目: %w", err)
	}

	// 更新字段
	if update.RelationshipType != "" {
		entry.RelationshipType = update.RelationshipType
	}
	if update.ShareholdingInfo != nil {
		entry.ShareholdingInfo = update.ShareholdingInfo
	}
	if update.RelatedCompanies != nil {
		entry.RelatedCompanies = update.RelatedCompanies
	}
	if update.EntityAliases != nil {
		entry.EntityAliases = update.EntityAliases
	}
	if update.DataSource != "" {
		entry.DataSource = update.DataSource
	}

	entry.UpdatedAt = time.Now()

	if err := s.db.WithContext(ctx).Save(&entry).Error; err != nil {
		return fmt.Errorf("更新冲突池条目失败: %w", err)
	}

	log.Printf("✅ 更新冲突池条目: ID=%d", entryID)
	return nil
}

// DeletePoolEntry 删除冲突池条目
func (s *conflictPoolService) DeletePoolEntry(ctx context.Context, entryID uint) error {
	if err := s.db.WithContext(ctx).Delete(&models.LawyerConflictPool{}, entryID).Error; err != nil {
		return fmt.Errorf("删除冲突池条目失败: %w", err)
	}

	log.Printf("✅ 删除冲突池条目: ID=%d", entryID)
	return nil
}

// RefreshFromAPI 从 API 刷新公司数据
func (s *conflictPoolService) RefreshFromAPI(ctx context.Context, entryID uint, provider CompanyAPIProvider) error {
	var entry models.LawyerConflictPool
	if err := s.db.WithContext(ctx).First(&entry, entryID).Error; err != nil {
		return fmt.Errorf("找不到冲突池条目: %w", err)
	}

	log.Printf("🔄 从 API 刷新数据: entryID=%d, provider=%s", entryID, provider)

	// 获取公司详细信息
	_, err := s.companyAPI.GetCompanyDetail(ctx, entry.EntityName, entry.EntityTaxID, provider)
	if err != nil {
		return fmt.Errorf("获取公司详情失败: %w", err)
	}

	// 获取股权穿透信息
	var shareholdingInfo models.JSON
	if entry.EntityTaxID != "" {
		shareholdingData, err := s.companyAPI.GetShareholding(ctx, entry.EntityTaxID, provider)
		if err == nil {
			shareholdingInfo = models.JSON(map[string]interface{}{
				"directShareholders": shareholdingData.DirectShareholders,
				"beneficialOwners":   shareholdingData.BeneficialOwners,
				"ultimateController": shareholdingData.UltimateController,
				"penetrationDepth":   shareholdingData.PenetrationDepth,
			})
		}
	}

	// 获取关联公司
	var relatedCompanies models.JSON
	if entry.EntityTaxID != "" {
		relatedData, err := s.companyAPI.GetRelatedCompanies(ctx, entry.EntityTaxID, provider)
		if err == nil {
			relatedCompanies = models.JSON(map[string]interface{}{
				"companies": relatedData,
			})
		}
	}

	// 更新条目
	_ = &PoolUpdateRequest{
		ShareholdingInfo: shareholdingInfo,
		RelatedCompanies: relatedCompanies,
		DataSource:       string(provider),
	}

	now := time.Now()
	entry.ShareholdingInfo = shareholdingInfo
	entry.RelatedCompanies = relatedCompanies
	entry.APIProvider = string(provider)
	entry.DataSource = string(provider)
	entry.LastVerifiedAt = &now
	entry.UpdatedAt = now

	if err := s.db.WithContext(ctx).Save(&entry).Error; err != nil {
		return fmt.Errorf("更新冲突池条目失败: %w", err)
	}

	log.Printf("✅ 从 API 刷新数据完成: entryID=%d", entryID)
	return nil
}

// ============================================================================
// 私有辅助方法
// ============================================================================

// standardizeEntityName 标准化实体名称
func (s *conflictPoolService) standardizeEntityName(name string) string {
	// 移除空格
	name = strings.TrimSpace(name)

	// 移除常见的公司后缀
	suffixes := []string{
		"有限公司", "股份有限公司", "集团有限公司",
		"有限责任公司", "控股有限公司", "科技有限公司",
		"（集团）有限公司", "投资有限公司",
	}

	for _, suffix := range suffixes {
		name = strings.TrimSuffix(name, suffix)
	}

	// 统一括号
	name = strings.ReplaceAll(name, "（", "(")
	name = strings.ReplaceAll(name, "）", ")")

	return strings.ToLower(strings.TrimSpace(name))
}

// determineRelationshipType 确定关系类型
func (s *conflictPoolService) determineRelationshipType(case_ *models.Case) string {
	// 根据案件类型判断关系
	caseType := strings.ToLower(case_.CaseType)

	if strings.Contains(caseType, "民事") || strings.Contains(caseType, "civil") {
		return "client"
	} else if strings.Contains(caseType, "刑事") || strings.Contains(caseType, "criminal") {
		return "client"
	} else if strings.Contains(caseType, "商事") || strings.Contains(caseType, "commercial") {
		return "client"
	}

	// 默认为客户关系
	return "client"
}

// calculateMatchScore 计算匹配分数
func (s *conflictPoolService) calculateMatchScore(searchTerm string, entry *models.LawyerConflictPool, searchType string) float64 {
	standardSearchTerm := s.standardizeEntityName(searchTerm)

	// 完全匹配
	if standardSearchTerm == entry.EntityNameStandard {
		return 1.0
	}

	// 税号匹配
	if searchTerm == entry.EntityTaxID {
		return 1.0
	}

	// 原名称匹配
	if searchTerm == entry.EntityName {
		return 0.95
	}

	// 包含匹配
	if strings.Contains(entry.EntityNameStandard, standardSearchTerm) {
		return 0.7
	}

	if strings.Contains(entry.EntityName, searchTerm) {
		return 0.6
	}

	// 别名匹配
	if entry.EntityAliases != nil {
		aliases := entry.EntityAliases.ToMap()
		for _, alias := range aliases {
			if aliasStr, ok := alias.(string); ok {
				if aliasStr == searchTerm || strings.Contains(aliasStr, searchTerm) {
					return 0.5
				}
			}
		}
	}

	return 0.3
}

// determineConflictType 确定冲突类型
func (s *conflictPoolService) determineConflictType(entry *models.LawyerConflictPool) string {
	switch entry.RelationshipType {
	case "client":
		return "客户冲突"
	case "opposing":
		return "对方当事人冲突"
	case "witness":
		return "证人冲突"
	default:
		return "一般冲突"
	}
}

// assessRiskLevel 评估风险等级
func (s *conflictPoolService) assessRiskLevel(entry *models.LawyerConflictPool, conflictType string) string {
	// 根据冲突类型和实体类型评估
	if conflictType == "对方当事人冲突" {
		return "CRITICAL"
	}

	if entry.EntityType == "company" {
		// 检查股权穿透信息
		if entry.ShareholdingInfo != nil {
			if data, ok := entry.ShareholdingInfo["directShareholders"].([]interface{}); ok {
				if len(data) > 0 {
					// 有股东信息的企业客户
					return "HIGH"
				}
			}
		}
		return "MEDIUM"
	}

	return "MEDIUM"
}

// getMatchReason 获取匹配原因
func (s *conflictPoolService) getMatchReason(entry *models.LawyerConflictPool, score float64) string {
	if score >= 1.0 {
		return "完全匹配"
	} else if score >= 0.9 {
		return "名称完全匹配"
	} else if score >= 0.7 {
		return "名称包含匹配"
	} else if score >= 0.5 {
		return "别名匹配"
	} else {
		return "部分匹配"
	}
}

// sortResults 按匹配分数排序结果
func (s *conflictPoolService) sortResults(results []*PoolMatchResult) {
	// 使用简单冒泡排序
	n := len(results)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if results[j].MatchScore < results[j+1].MatchScore {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}
}

// enrichCompanyData 丰富公司数据
func (s *conflictPoolService) enrichCompanyData(ctx context.Context, entryID uint, companyName, taxID string) {
	log.Printf("🔄 丰富公司数据: entryID=%d, name=%s", entryID, companyName)

	// 使用模拟数据获取详情
	_, err := s.companyAPI.GetCompanyDetail(ctx, companyName, taxID, ProviderMock)
	if err != nil {
		log.Printf("⚠️ 获取公司详情失败: %v", err)
		return
	}

	// 获取股权穿透
	var shareholdingInfo models.JSON
	if taxID != "" {
		shareholdingData, err := s.companyAPI.GetShareholding(ctx, taxID, ProviderMock)
		if err == nil {
			data, _ := json.Marshal(shareholdingData)
			shareholdingInfo = models.JSON(map[string]interface{}{
				"data": string(data),
			})
		}
	}

	// 获取关联公司
	var relatedCompanies models.JSON
	if taxID != "" {
		relatedData, err := s.companyAPI.GetRelatedCompanies(ctx, taxID, ProviderMock)
		if err == nil {
			data, _ := json.Marshal(relatedData)
			relatedCompanies = models.JSON(map[string]interface{}{
				"data": string(data),
			})
		}
	}

	// 更新数据库
	if err := s.db.Model(&models.LawyerConflictPool{}).
		Where("id = ?", entryID).
		Updates(map[string]interface{}{
			"shareholding_info": shareholdingInfo,
			"related_companies": relatedCompanies,
			"api_provider":      "mock",
			"data_source":       "api",
			"last_verified_at":  time.Now(),
		}).Error; err != nil {
		log.Printf("⚠️ 更新公司数据失败: %v", err)
	}

	log.Printf("✅ 丰富公司数据完成: entryID=%d", entryID)
}

// ============================================================================
// 批量操作辅助方法
// ============================================================================

// BatchSyncLawyers 批量同步多个律师的冲突池
func (s *conflictPoolService) BatchSyncLawyers(ctx context.Context, lawyerIDs []uint) (*SyncResult, error) {
	log.Printf("🔄 开始批量同步律师冲突池: count=%d", len(lawyerIDs))
	startTime := time.Now()

	result := &SyncResult{
		StartTime: startTime,
		Errors:    make([]string, 0),
	}

	result.TotalLawyers = len(lawyerIDs)

	for _, lawyerID := range lawyerIDs {
		// 获取律师的所有案件
		cases, err := s.caseRepo.ListByLawyer(ctx, lawyerID)
		if err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("律师 %d 获取案件失败: %v", lawyerID, err))
			result.FailedEntries++
			continue
		}

		// 处理每个案件
		for _, case_ := range cases {
			if err := s.SyncLawyerPool(ctx, lawyerID, case_.ID); err != nil {
				result.Errors = append(result.Errors,
					fmt.Sprintf("案件 %d 处理失败: %v", case_.ID, err))
				result.FailedEntries++
			} else {
				result.ProcessedCases++
			}
		}
	}

	result.EndTime = time.Now()
	result.DurationMs = result.EndTime.Sub(startTime).Milliseconds()

	log.Printf("✅ 批量同步完成: 处理案件=%d, 失败=%d, 耗时=%dms",
		result.ProcessedCases, result.FailedEntries, result.DurationMs)

	return result, nil
}

// GetPoolStats 获取冲突池统计信息
func (s *conflictPoolService) GetPoolStats(ctx context.Context, lawyerID uint) (*PoolStats, error) {
	stats := &PoolStats{}

	// 总条目数
	if err := s.db.WithContext(ctx).
		Model(&models.LawyerConflictPool{}).
		Where("lawyer_id = ?", lawyerID).
		Count(&stats.TotalEntries).Error; err != nil {
		return nil, fmt.Errorf("获取总条目数失败: %w", err)
	}

	// 按关系类型统计
	var relationshipStats []struct {
		RelationshipType string
		Count            int64
	}
	if err := s.db.WithContext(ctx).
		Model(&models.LawyerConflictPool{}).
		Select("relationship_type, count(*) as count").
		Where("lawyer_id = ?", lawyerID).
		Group("relationship_type").
		Scan(&relationshipStats).Error; err != nil {
		return nil, fmt.Errorf("获取关系类型统计失败: %w", err)
	}

	stats.ByRelationship = make(map[string]int64)
	for _, stat := range relationshipStats {
		stats.ByRelationship[stat.RelationshipType] = stat.Count
	}

	// 按实体类型统计
	var entityStats []struct {
		EntityType string
		Count      int64
	}
	if err := s.db.WithContext(ctx).
		Model(&models.LawyerConflictPool{}).
		Select("entity_type, count(*) as count").
		Where("lawyer_id = ?", lawyerID).
		Group("entity_type").
		Scan(&entityStats).Error; err != nil {
		return nil, fmt.Errorf("获取实体类型统计失败: %w", err)
	}

	stats.ByEntityType = make(map[string]int64)
	for _, stat := range entityStats {
		stats.ByEntityType[stat.EntityType] = stat.Count
	}

	// API 数据覆盖率
	var apiDataCount int64
	if err := s.db.WithContext(ctx).
		Model(&models.LawyerConflictPool{}).
		Where("lawyer_id = ? AND data_source = ?", lawyerID, "api").
		Count(&apiDataCount).Error; err != nil {
		return nil, fmt.Errorf("获取 API 数据统计失败: %w", err)
	}

	if stats.TotalEntries > 0 {
		stats.APIDataCoverage = float64(apiDataCount) / float64(stats.TotalEntries)
	}

	return stats, nil
}

// PoolStats 冲突池统计信息
type PoolStats struct {
	TotalEntries    int64            `json:"totalEntries"`
	ByRelationship  map[string]int64 `json:"byRelationship"`
	ByEntityType    map[string]int64 `json:"byEntityType"`
	APIDataCoverage float64          `json:"apiDataCoverage"`
}
