package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/models/elasticsearch"
	"law-oa-go/internal/repositories"

	"gorm.io/gorm"
)

// LegalStatuteService 法条服务接口
type LegalStatuteService interface {
	// 基础CRUD操作
	CreateStatute(ctx context.Context, req *models.LegalStatuteCreateRequest, userID int) (*models.LegalStatuteResponse, error)
	GetStatuteByID(ctx context.Context, id int, userID int) (*models.LegalStatuteResponse, error)
	GetStatuteByNumber(ctx context.Context, number string, userID int) (*models.LegalStatuteResponse, error)
	UpdateStatute(ctx context.Context, id int, req *models.LegalStatuteUpdateRequest, userID int) (*models.LegalStatuteResponse, error)
	DeleteStatute(ctx context.Context, id int, userID int) error
	ListStatutes(ctx context.Context, page, pageSize int, userID int) (*models.LegalSearchResponse, error)

	// 搜索功能
	SearchStatutes(ctx context.Context, req *models.LegalSearchRequest, userID int) (*models.LegalSearchResponse, error)
	GetSearchSuggestions(ctx context.Context, query string) ([]string, error)
	GetRelatedStatutes(ctx context.Context, statuteID int, limit int, userID int) ([]*models.LegalStatuteResponse, error)

	// 分类管理
	GetCategories(ctx context.Context) ([]*models.LegalCategoryResponse, error)
	GetCategoryTree(ctx context.Context) ([]*models.CategoryTreeNode, error)
	GetStatutesByCategory(ctx context.Context, categoryID int, page, pageSize int) (*models.LegalSearchResponse, error)

	// 标签管理
	GetTags(ctx context.Context) ([]*models.LegalTagResponse, error)
	GetPopularTags(ctx context.Context, limit int) ([]*models.LegalTagResponse, error)

	// 收藏功能
	AddToFavorites(ctx context.Context, statuteID int, userID int) error
	RemoveFromFavorites(ctx context.Context, statuteID int, userID int) error
	GetUserFavorites(ctx context.Context, userID int, page, pageSize int) (*models.LegalSearchResponse, error)

	// 搜索历史
	GetSearchHistory(ctx context.Context, userID int, limit int) ([]*models.SearchHistoryResponse, error)
	GetPopularSearches(ctx context.Context, limit int) ([]string, error)

	// 统计和推荐
	GetCategoryStats(ctx context.Context) ([]*models.CategoryStat, error)
	GetPopularStatutes(ctx context.Context, limit int, userID int) ([]*models.LegalStatuteResponse, error)
	GetRecentUpdates(ctx context.Context, days int) ([]*models.LegalStatuteResponse, error)

	// 管理功能
	SyncToElasticsearch(ctx context.Context) error
	RebuildSearchIndex(ctx context.Context) error
}

// legalStatuteService 法条服务实现
type legalStatuteService struct {
	db               *gorm.DB
	statuteRepo      repositories.LegalStatuteRepository
	categoryRepo     repositories.LegalCategoryRepository
	tagRepo          repositories.LegalTagRepository
	esRepo           repositories.ElasticsearchStatuteRepository
}

// NewLegalStatuteService 创建法条服务实例
func NewLegalStatuteService(
	db *gorm.DB,
	statuteRepo repositories.LegalStatuteRepository,
	categoryRepo repositories.LegalCategoryRepository,
	tagRepo repositories.LegalTagRepository,
	esRepo repositories.ElasticsearchStatuteRepository,
) LegalStatuteService {
	return &legalStatuteService{
		db:           db,
		statuteRepo:  statuteRepo,
		categoryRepo: categoryRepo,
		tagRepo:      tagRepo,
		esRepo:       esRepo,
	}
}

// CreateStatute 创建法条
func (s *legalStatuteService) CreateStatute(ctx context.Context, req *models.LegalStatuteCreateRequest, userID int) (*models.LegalStatuteResponse, error) {
	// 验证请求
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("请求验证失败: %v", err)
	}

	// 检查法条编号是否已存在
	if existing, _ := s.statuteRepo.GetByStatuteNumber(ctx, req.StatuteNumber); existing != nil {
		return nil, fmt.Errorf("法条编号 %s 已存在", req.StatuteNumber)
	}

	// 验证分类是否存在
	if _, err := s.categoryRepo.GetByID(ctx, req.CategoryID); err != nil {
		return nil, fmt.Errorf("分类不存在: %v", err)
	}

	// 创建法条对象
	statute := &models.LegalStatute{
		StatuteNumber:       req.StatuteNumber,
		Title:               req.Title,
		Content:             req.Content,
		CategoryID:          req.CategoryID,
		LawName:             req.LawName,
		Chapter:             req.Chapter,
		Section:             req.Section,
		Part:                req.Part,
		EffectiveDate:       req.EffectiveDate,
		ExpiryDate:          req.ExpiryDate,
		PublishingAuthority: req.PublishingAuthority,
		Status:              req.Status,
		HierarchyLevel:      req.HierarchyLevel,
		ParentStatuteID:     req.ParentStatuteID,
		OrderInHierarchy:    req.OrderInHierarchy,
		Tags:                req.Tags,
		Keywords:            req.Keywords,
	}

	// 保存到数据库
	if err := s.statuteRepo.Create(ctx, statute); err != nil {
		return nil, fmt.Errorf("创建法条失败: %v", err)
	}

	// 同步到Elasticsearch
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		fullStatute, err := s.statuteRepo.GetByID(ctx, statute.ID)
		if err == nil && s.esRepo != nil {
			esDoc := s.convertToESDocument(fullStatute)
			if err := s.esRepo.IndexDocument(ctx, esDoc); err != nil {
				log.Printf("同步法条到Elasticsearch失败: %v", err)
			}
		}
	}()

	// 处理标签
	s.processTags(ctx, statute.Tags)

	// 转换为响应格式
	response := s.convertToResponse(statute, userID)
	return response, nil
}

// GetStatuteByID 根据ID获取法条
func (s *legalStatuteService) GetStatuteByID(ctx context.Context, id int, userID int) (*models.LegalStatuteResponse, error) {
	statute, err := s.statuteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取法条失败: %v", err)
	}

	// 更新浏览量（这里可以添加浏览量统计逻辑）

	// 转换为响应格式
	response := s.convertToResponse(statute, userID)
	return response, nil
}

// GetStatuteByNumber 根据法条编号获取法条
func (s *legalStatuteService) GetStatuteByNumber(ctx context.Context, number string, userID int) (*models.LegalStatuteResponse, error) {
	statute, err := s.statuteRepo.GetByStatuteNumber(ctx, number)
	if err != nil {
		return nil, fmt.Errorf("获取法条失败: %v", err)
	}

	// 转换为响应格式
	response := s.convertToResponse(statute, userID)
	return response, nil
}

// UpdateStatute 更新法条
func (s *legalStatuteService) UpdateStatute(ctx context.Context, id int, req *models.LegalStatuteUpdateRequest, userID int) (*models.LegalStatuteResponse, error) {
	// 验证请求
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("请求验证失败: %v", err)
	}

	// 获取现有法条
	statute, err := s.statuteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("法条不存在: %v", err)
	}

	// 检查是否有变更
	hasChanges := false
	if req.Title != "" && req.Title != statute.Title {
		statute.Title = req.Title
		hasChanges = true
	}
	if req.Content != "" && req.Content != statute.Content {
		statute.Content = req.Content
		hasChanges = true
	}
	if req.CategoryID > 0 && req.CategoryID != statute.CategoryID {
		statute.CategoryID = req.CategoryID
		hasChanges = true
	}
	// 其他字段更新...

	// 如果有变更，创建版本记录
	if hasChanges {
		version := &models.LegalStatuteVersion{
			StatuteID:         statute.ID,
			VersionNumber:     s.getNextVersionNumber(ctx, statute.ID),
			Title:             statute.Title,
			Content:           statute.Content,
			EffectiveDate:     statute.EffectiveDate,
			ExpiryDate:        statute.ExpiryDate,
			ChangeDescription: req.ChangeDescription,
			CreatedBy:         userID,
		}

		if err := s.statuteRepo.CreateVersion(ctx, version); err != nil {
			log.Printf("创建版本记录失败: %v", err)
		}
	}

	// 更新法条
	if err := s.statuteRepo.Update(ctx, statute); err != nil {
		return nil, fmt.Errorf("更新法条失败: %v", err)
	}

	// 同步到Elasticsearch
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if s.esRepo != nil {
			esDoc := s.convertToESDocument(statute)
			if err := s.esRepo.IndexDocument(ctx, esDoc); err != nil {
				log.Printf("同步法条到Elasticsearch失败: %v", err)
			}
		}
	}()

	// 处理标签
	if req.Tags != nil {
		s.processTags(ctx, req.Tags)
		statute.Tags = req.Tags
	}

	// 转换为响应格式
	response := s.convertToResponse(statute, userID)
	return response, nil
}

// DeleteStatute 删除法条
func (s *legalStatuteService) DeleteStatute(ctx context.Context, id int, userID int) error {
	// 检查法条是否存在
	_, err := s.statuteRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("法条不存在: %v", err)
	}

	// 删除法条
	if err := s.statuteRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("删除法条失败: %v", err)
	}

	// 从Elasticsearch删除
	go func() {
		if s.esRepo != nil {
			// 这里需要实现从ES删除文档的逻辑
			log.Printf("TODO: 从Elasticsearch删除法条 %d", id)
		}
	}()

	return nil
}

// ListStatutes 获取法条列表
func (s *legalStatuteService) ListStatutes(ctx context.Context, page, pageSize int, userID int) (*models.LegalSearchResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// 获取法条列表
	statutes, err := s.statuteRepo.List(ctx, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("获取法条列表失败: %v", err)
	}

	// 获取总数
	total, err := s.statuteRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取法条总数失败: %v", err)
	}

	// 转换为响应格式
	statuteResponses := make([]*models.LegalStatuteResponse, 0, len(statutes))
	for _, statute := range statutes {
		response := s.convertToResponse(statute, userID)
		statuteResponses = append(statuteResponses, response)
	}

	return &models.LegalSearchResponse{
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
		Statutes:   statutes,
		SearchTime: 0,
	}, nil
}

// SearchStatutes 搜索法条
func (s *legalStatuteService) SearchStatutes(ctx context.Context, req *models.LegalSearchRequest, userID int) (*models.LegalSearchResponse, error) {
	// 验证请求
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("请求验证失败: %v", err)
	}

	start := time.Now()

	// 临时禁用Elasticsearch，直接使用数据库搜索进行调试
	// if s.esRepo != nil && req.Query != "" {
	// 	return s.searchWithElasticsearch(ctx, req, userID)
	// }

	// 使用数据库搜索
	return s.searchWithDatabase(ctx, req, userID, start)
}

// searchWithElasticsearch 使用Elasticsearch搜索
func (s *legalStatuteService) searchWithElasticsearch(ctx context.Context, req *models.LegalSearchRequest, userID int) (*models.LegalSearchResponse, error) {
	// 构建ES搜索请求
	esReq := &elasticsearch.LegalSearchRequest{
		Query:             req.Query,
		CategoryCodes:     s.getCategoryCodes(ctx, req.CategoryID),
		Status:            req.Status,
		Tags:              req.Tags,
		SortBy:            req.SortBy,
		SortOrder:         req.SortOrder,
		Page:              req.Page,
		PageSize:          req.PageSize,
		Highlight:         true,
	}

	if req.LawName != "" {
		esReq.LawNames = []string{req.LawName}
	}

	// 执行搜索
	esResp, err := s.esRepo.Search(ctx, esReq)
	if err != nil {
		log.Printf("Elasticsearch搜索失败，回退到数据库搜索: %v", err)
		return s.searchWithDatabase(ctx, req, userID, time.Now())
	}

	// 转换为响应格式
	statutes := make([]*models.LegalStatute, 0, len(esResp.Documents))
	for _, esDoc := range esResp.Documents {
		statute := s.convertESDocumentToModel(&esDoc)
		statutes = append(statutes, statute)
	}

	// 获取分类信息
	s.loadCategoriesForStatutes(ctx, statutes)

	// 转换为响应格式
	statuteResponses := make([]*models.LegalStatuteResponse, 0, len(statutes))
	for _, statute := range statutes {
		response := s.convertToResponse(statute, userID)
		statuteResponses = append(statuteResponses, response)
	}

	return &models.LegalSearchResponse{
		Total:       esResp.Total,
		Page:        req.Page,
		PageSize:    req.PageSize,
		TotalPages:  int((esResp.Total + int64(req.PageSize) - 1) / int64(req.PageSize)),
		Statutes:    statutes,
		Suggestions: esResp.Suggestions,
		SearchTime:  esResp.SearchTime,
	}, nil
}

// searchWithDatabase 使用数据库搜索
func (s *legalStatuteService) searchWithDatabase(ctx context.Context, req *models.LegalSearchRequest, userID int, start time.Time) (*models.LegalSearchResponse, error) {
	var statutes []*models.LegalStatute
	var total int64
	var err error

	offset := (req.Page - 1) * req.PageSize

	// 根据查询条件选择不同的搜索方法
	if req.Query != "" {
		statutes, err = s.statuteRepo.SearchByKeyword(ctx, req.Query, offset, req.PageSize)
		if err != nil {
			return nil, fmt.Errorf("关键词搜索失败: %v", err)
		}
		// 估算总数（简化实现）
		total = int64(len(statutes))
	} else if req.CategoryID > 0 {
		statutes, err = s.statuteRepo.FindByCategory(ctx, req.CategoryID, offset, req.PageSize)
		if err != nil {
			return nil, fmt.Errorf("分类搜索失败: %v", err)
		}
		// 获取分类下的法条总数（简化实现）
		total = int64(len(statutes))
	} else if req.LawName != "" {
		statutes, err = s.statuteRepo.FindByLawName(ctx, req.LawName, offset, req.PageSize)
		if err != nil {
			return nil, fmt.Errorf("法律名称搜索失败: %v", err)
		}
		total = int64(len(statutes))
	} else {
		statutes, err = s.statuteRepo.List(ctx, offset, req.PageSize)
		if err != nil {
			return nil, fmt.Errorf("获取法条列表失败: %v", err)
		}
		total, err = s.statuteRepo.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("获取法条总数失败: %v", err)
		}
	}

	// 转换为响应格式
	statuteResponses := make([]*models.LegalStatuteResponse, 0, len(statutes))
	for _, statute := range statutes {
		response := s.convertToResponse(statute, userID)
		statuteResponses = append(statuteResponses, response)
	}

	return &models.LegalSearchResponse{
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: int((total + int64(req.PageSize) - 1) / int64(req.PageSize)),
		Statutes:   statutes,
		SearchTime: int(time.Since(start).Milliseconds()),
	}, nil
}

// GetSearchSuggestions 获取搜索建议
func (s *legalStatuteService) GetSearchSuggestions(ctx context.Context, query string) ([]string, error) {
	if s.esRepo != nil {
		return s.esRepo.GetSuggestion(ctx, query)
	}

	// 使用数据库提供基础建议
	suggestions := []string{
		query + "相关法条",
		"民法典" + query,
		"公司法" + query,
		"劳动合同法" + query,
	}

	return suggestions, nil
}

// GetRelatedStatutes 获取相关法条
func (s *legalStatuteService) GetRelatedStatutes(ctx context.Context, statuteID int, limit int, userID int) ([]*models.LegalStatuteResponse, error) {
	var statutes []*models.LegalStatute

	if s.esRepo != nil {
		esDocs, err := s.esRepo.GetRelatedStatutes(ctx, statuteID, limit)
		if err != nil {
			log.Printf("ES获取相关法条失败，使用数据库: %v", err)
		} else {
			// 转换ES文档到模型
			for _, esDoc := range esDocs {
				statute := s.convertESDocumentToModel(esDoc)
				statutes = append(statutes, statute)
			}
		}
	}

	// 如果ES没有返回结果，使用数据库逻辑
	if len(statutes) == 0 {
		// 获取当前法条
		current, err := s.statuteRepo.GetByID(ctx, statuteID)
		if err != nil {
			return nil, fmt.Errorf("获取当前法条失败: %v", err)
		}

		// 查找同分类的其他法条
		statutes, err = s.statuteRepo.FindByCategory(ctx, current.CategoryID, 0, limit+1)
		if err != nil {
			return nil, fmt.Errorf("查找相关法条失败: %v", err)
		}

		// 过滤掉当前法条
		filtered := make([]*models.LegalStatute, 0, len(statutes))
		for _, statute := range statutes {
			if statute.ID != statuteID {
				filtered = append(filtered, statute)
			}
		}
		statutes = filtered
	}

	// 转换为响应格式
	responses := make([]*models.LegalStatuteResponse, 0, len(statutes))
	for _, statute := range statutes {
		response := s.convertToResponse(statute, userID)
		responses = append(responses, response)
	}

	return responses, nil
}

// GetCategories 获取所有分类
func (s *legalStatuteService) GetCategories(ctx context.Context) ([]*models.LegalCategoryResponse, error) {
	categories, err := s.categoryRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取分类列表失败: %v", err)
	}

	responses := make([]*models.LegalCategoryResponse, 0, len(categories))
	for _, category := range categories {
		response := &models.LegalCategoryResponse{
			ID:           category.ID,
			Name:         category.Name,
			Code:         category.Code,
			ParentID:     category.ParentID,
			Level:        category.Level,
			Description:  category.Description,
			IsActive:     category.IsActive,
			StatuteCount: 0, // 需要统计
			CreatedAt:    category.CreatedAt,
			UpdatedAt:    category.UpdatedAt,
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// GetCategoryTree 获取分类树
func (s *legalStatuteService) GetCategoryTree(ctx context.Context) ([]*models.CategoryTreeNode, error) {
	return s.categoryRepo.GetTree(ctx)
}

// GetStatutesByCategory 根据分类获取法条
func (s *legalStatuteService) GetStatutesByCategory(ctx context.Context, categoryID int, page, pageSize int) (*models.LegalSearchResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	statutes, err := s.statuteRepo.FindByCategory(ctx, categoryID, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("获取分类法条失败: %v", err)
	}

	// 简化总数计算
	total := int64(len(statutes))

	statuteResponses := make([]*models.LegalStatuteResponse, 0, len(statutes))
	for _, statute := range statutes {
		response := s.convertToResponse(statute, 0)
		statuteResponses = append(statuteResponses, response)
	}

	return &models.LegalSearchResponse{
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
		Statutes:   statutes,
		SearchTime: 0,
	}, nil
}

// GetTags 获取所有标签
func (s *legalStatuteService) GetTags(ctx context.Context) ([]*models.LegalTagResponse, error) {
	tags, err := s.tagRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取标签列表失败: %v", err)
	}

	responses := make([]*models.LegalTagResponse, 0, len(tags))
	for _, tag := range tags {
		response := &models.LegalTagResponse{
			ID:          tag.ID,
			Name:        tag.Name,
			Color:       tag.Color,
			Description: tag.Description,
			UsageCount:  tag.UsageCount,
			CreatedAt:   tag.CreatedAt,
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// GetPopularTags 获取热门标签
func (s *legalStatuteService) GetPopularTags(ctx context.Context, limit int) ([]*models.LegalTagResponse, error) {
	tags, err := s.tagRepo.GetPopularTags(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("获取热门标签失败: %v", err)
	}

	responses := make([]*models.LegalTagResponse, 0, len(tags))
	for _, tag := range tags {
		response := &models.LegalTagResponse{
			ID:          tag.ID,
			Name:        tag.Name,
			Color:       tag.Color,
			Description: tag.Description,
			UsageCount:  tag.UsageCount,
			CreatedAt:   tag.CreatedAt,
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// AddToFavorites 添加到收藏
func (s *legalStatuteService) AddToFavorites(ctx context.Context, statuteID int, userID int) error {
	// 检查法条是否存在
	_, err := s.statuteRepo.GetByID(ctx, statuteID)
	if err != nil {
		return fmt.Errorf("法条不存在: %v", err)
	}

	// 检查是否已收藏
	favorited, err := s.statuteRepo.IsFavorited(ctx, userID, statuteID)
	if err != nil {
		return fmt.Errorf("检查收藏状态失败: %v", err)
	}
	if favorited {
		return fmt.Errorf("法条已在收藏中")
	}

	// 添加收藏
	return s.statuteRepo.AddToFavorites(ctx, userID, statuteID)
}

// RemoveFromFavorites 移除收藏
func (s *legalStatuteService) RemoveFromFavorites(ctx context.Context, statuteID int, userID int) error {
	return s.statuteRepo.RemoveFromFavorites(ctx, userID, statuteID)
}

// GetUserFavorites 获取用户收藏
func (s *legalStatuteService) GetUserFavorites(ctx context.Context, userID int, page, pageSize int) (*models.LegalSearchResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	statutes, err := s.statuteRepo.GetUserFavorites(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("获取收藏列表失败: %v", err)
	}

	// 简化总数计算
	total := int64(len(statutes))

	statuteResponses := make([]*models.LegalStatuteResponse, 0, len(statutes))
	for _, statute := range statutes {
		response := s.convertToResponse(statute, userID)
		response.IsFavorited = true
		statuteResponses = append(statuteResponses, response)
	}

	return &models.LegalSearchResponse{
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
		Statutes:   statutes,
		SearchTime: 0,
	}, nil
}

// GetSearchHistory 获取搜索历史
func (s *legalStatuteService) GetSearchHistory(ctx context.Context, userID int, limit int) ([]*models.SearchHistoryResponse, error) {
	histories, err := s.statuteRepo.GetUserSearchHistory(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("获取搜索历史失败: %v", err)
	}

	responses := make([]*models.SearchHistoryResponse, 0, len(histories))
	for _, history := range histories {
		response := &models.SearchHistoryResponse{
			ID:            history.ID,
			UserID:        history.UserID,
			SearchQuery:   history.SearchQuery,
			SearchFilters: history.SearchFilters,
			ResultCount:   history.ResultCount,
			SearchDuration: history.SearchDuration,
			CreatedAt:     history.CreatedAt,
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// GetPopularSearches 获取热门搜索
func (s *legalStatuteService) GetPopularSearches(ctx context.Context, limit int) ([]string, error) {
	return s.statuteRepo.GetPopularSearches(ctx, limit)
}

// GetCategoryStats 获取分类统计
func (s *legalStatuteService) GetCategoryStats(ctx context.Context) ([]*models.CategoryStat, error) {
	return s.statuteRepo.GetCategoryStats(ctx)
}

// GetPopularStatutes 获取热门法条
func (s *legalStatuteService) GetPopularStatutes(ctx context.Context, limit int, userID int) ([]*models.LegalStatuteResponse, error) {
	statutes, err := s.statuteRepo.GetPopularStatutes(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("获取热门法条失败: %v", err)
	}

	responses := make([]*models.LegalStatuteResponse, 0, len(statutes))
	for _, statute := range statutes {
		response := s.convertToResponse(statute, userID)
		responses = append(responses, response)
	}

	return responses, nil
}

// GetRecentUpdates 获取最近更新
func (s *legalStatuteService) GetRecentUpdates(ctx context.Context, days int) ([]*models.LegalStatuteResponse, error) {
	statutes, err := s.statuteRepo.GetRecentUpdates(ctx, days)
	if err != nil {
		return nil, fmt.Errorf("获取最近更新失败: %v", err)
	}

	responses := make([]*models.LegalStatuteResponse, 0, len(statutes))
	for _, statute := range statutes {
		response := s.convertToResponse(statute, 0)
		responses = append(responses, response)
	}

	return responses, nil
}

// SyncToElasticsearch 同步到Elasticsearch
func (s *legalStatuteService) SyncToElasticsearch(ctx context.Context) error {
	if s.esRepo == nil {
		return fmt.Errorf("Elasticsearch未配置")
	}

	// TODO: 实现同步功能
	log.Println("Elasticsearch同步功能待实现")
	return nil
}

// RebuildSearchIndex 重建搜索索引
func (s *legalStatuteService) RebuildSearchIndex(ctx context.Context) error {
	if s.esRepo == nil {
		return fmt.Errorf("Elasticsearch未配置")
	}

	// TODO: 实现重建索引功能
	log.Println("Elasticsearch重建索引功能待实现")
	return nil
}

// 辅助方法

// convertToResponse 转换为响应格式
func (s *legalStatuteService) convertToResponse(statute *models.LegalStatute, userID int) *models.LegalStatuteResponse {
	response := &models.LegalStatuteResponse{
		ID:                 statute.ID,
		StatuteNumber:      statute.StatuteNumber,
		Title:              statute.Title,
		Content:            statute.Content,
		LawName:            statute.LawName,
		Chapter:            statute.Chapter,
		Section:            statute.Section,
		Part:               statute.Part,
		EffectiveDate:      statute.EffectiveDate,
		ExpiryDate:         statute.ExpiryDate,
		PublishingAuthority: statute.PublishingAuthority,
		Status:             statute.Status,
		HierarchyLevel:     statute.HierarchyLevel,
		ParentStatuteID:    statute.ParentStatuteID,
		OrderInHierarchy:   statute.OrderInHierarchy,
		Tags:               statute.Tags,
		Keywords:           statute.Keywords,
		ViewCount:          0, // 需要从统计表获取
		FavoriteCount:      0, // 需要从统计表获取
		CreatedAt:          statute.CreatedAt,
		UpdatedAt:          statute.UpdatedAt,
		FullPath:           statute.GetFullPath(),
		IsActive:           statute.IsActive(),
	}

	// 处理分类信息
	if statute.Category != nil {
		response.Category = &models.LegalCategoryResponse{
			ID:          statute.Category.ID,
			Name:        statute.Category.Name,
			Code:        statute.Category.Code,
			Level:       statute.Category.Level,
			Description: statute.Category.Description,
			IsActive:    statute.Category.IsActive,
		}
	}

	// 检查是否已收藏
	if userID > 0 {
		if favorited, _ := s.statuteRepo.IsFavorited(context.Background(), userID, statute.ID); favorited {
			response.IsFavorited = true
		}
	}

	return response
}

// convertToESDocument 转换为ES文档
func (s *legalStatuteService) convertToESDocument(statute *models.LegalStatute) *elasticsearch.LegalStatuteDocument {
	doc := &elasticsearch.LegalStatuteDocument{
		ID:                 statute.ID,
		StatuteNumber:      statute.StatuteNumber,
		Title:              statute.Title,
		Content:            statute.Content,
		LawName:            statute.LawName,
		Chapter:            statute.Chapter,
		Section:            statute.Section,
		Part:               statute.Part,
		Status:             statute.Status,
		HierarchyLevel:     statute.HierarchyLevel,
		Tags:               statute.Tags,
		Keywords:           statute.Keywords,
		CreatedAt:          statute.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:          statute.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		ContentLength:      len(statute.Content),
		ViewCount:          0,
		FavoriteCount:      0,
		SearchWeight:       1.0,
	}

	// 处理可选字段
	if statute.EffectiveDate != nil {
		doc.EffectiveDate = statute.EffectiveDate.Format("2006-01-02")
	}
	if statute.ExpiryDate != nil {
		doc.ExpiryDate = statute.ExpiryDate.Format("2006-01-02")
	}
	if statute.ParentStatuteID != nil {
		doc.ParentStatuteID = *statute.ParentStatuteID
	}

	// 处理分类信息
	if statute.Category != nil {
		doc.Category = elasticsearch.Category{
			ID:   statute.Category.ID,
			Name: statute.Category.Name,
			Code: statute.Category.Code,
		}
	}

	return doc
}

// convertESDocumentToModel 转换ES文档到模型
func (s *legalStatuteService) convertESDocumentToModel(doc *elasticsearch.LegalStatuteDocument) *models.LegalStatute {
	statute := &models.LegalStatute{
		ID:              doc.ID,
		StatuteNumber:   doc.StatuteNumber,
		Title:           doc.Title,
		Content:         doc.Content,
		LawName:         doc.LawName,
		Chapter:         doc.Chapter,
		Section:         doc.Section,
		Part:            doc.Part,
		Status:          doc.Status,
		HierarchyLevel:  doc.HierarchyLevel,
		Tags:            doc.Tags,
		Keywords:        doc.Keywords,
		ParentStatuteID: &doc.ParentStatuteID,
	}

	// 解析日期
	if doc.EffectiveDate != "" {
		if effectiveDate, err := time.Parse("2006-01-02", doc.EffectiveDate); err == nil {
			statute.EffectiveDate = &effectiveDate
		}
	}
	if doc.ExpiryDate != "" {
		if expiryDate, err := time.Parse("2006-01-02", doc.ExpiryDate); err == nil {
			statute.ExpiryDate = &expiryDate
		}
	}

	// 解析创建时间
	if createdAt, err := time.Parse(time.RFC3339, doc.CreatedAt); err == nil {
		statute.CreatedAt = createdAt
	}
	if updatedAt, err := time.Parse(time.RFC3339, doc.UpdatedAt); err == nil {
		statute.UpdatedAt = updatedAt
	}

	return statute
}

// getNextVersionNumber 获取下一个版本号
func (s *legalStatuteService) getNextVersionNumber(ctx context.Context, statuteID int) int {
	versions, err := s.statuteRepo.GetVersions(ctx, statuteID)
	if err != nil || len(versions) == 0 {
		return 1
	}

	return versions[0].VersionNumber + 1
}

// getCategoryCodes 获取分类代码
func (s *legalStatuteService) getCategoryCodes(ctx context.Context, categoryID int) []string {
	if categoryID <= 0 {
		return []string{}
	}

	category, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return []string{}
	}

	return []string{category.Code}
}

// processTags 处理标签
func (s *legalStatuteService) processTags(ctx context.Context, tags []string) {
	for _, tagName := range tags {
		// 检查标签是否存在
		tag, err := s.tagRepo.GetByName(ctx, tagName)
		if err != nil {
			// 创建新标签
			tag = &models.LegalTag{
				Name:  tagName,
				Color: "#1890ff",
			}
			if err := s.tagRepo.Create(ctx, tag); err != nil {
				log.Printf("创建标签失败: %v", err)
				continue
			}
		}

		// 更新使用次数
		if err := s.tagRepo.UpdateUsageCount(ctx, tag.ID); err != nil {
			log.Printf("更新标签使用次数失败: %v", err)
		}
	}
}

// loadCategoriesForStatutes 为法条加载分类信息
func (s *legalStatuteService) loadCategoriesForStatutes(ctx context.Context, statutes []*models.LegalStatute) {
	categoryIDs := make([]int, 0, len(statutes))
	for _, statute := range statutes {
		if statute.CategoryID > 0 {
			categoryIDs = append(categoryIDs, statute.CategoryID)
		}
	}

	// 批量获取分类信息
	categories := make(map[int]*models.LegalCategory)
	for _, categoryID := range categoryIDs {
		if category, err := s.categoryRepo.GetByID(ctx, categoryID); err == nil {
			categories[categoryID] = category
		}
	}

	// 设置分类信息
	for _, statute := range statutes {
		if category, exists := categories[statute.CategoryID]; exists {
			statute.Category = category
		}
	}
}