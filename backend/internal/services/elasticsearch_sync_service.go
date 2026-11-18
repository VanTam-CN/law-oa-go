package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/models/elasticsearch"
	"law-oa-go/internal/repositories"
)

// ElasticsearchSyncService Elasticsearch同步服务
type ElasticsearchSyncService struct {
	statuteRepo  repositories.LegalStatuteRepository
	esRepo       repositories.ElasticsearchStatuteRepository
	batchSize    int
	syncInterval time.Duration
	stopChan     chan bool
}

// NewElasticsearchSyncService 创建Elasticsearch同步服务
func NewElasticsearchSyncService(
	statuteRepo repositories.LegalStatuteRepository,
	esRepo repositories.ElasticsearchStatuteRepository,
) *ElasticsearchSyncService {
	return &ElasticsearchSyncService{
		statuteRepo:  statuteRepo,
		esRepo:       esRepo,
		batchSize:    100,
		syncInterval: 30 * time.Minute,
		stopChan:     make(chan bool),
	}
}

// StartSyncWorker 启动同步工作器
func (s *ElasticsearchSyncService) StartSyncWorker(ctx context.Context) {
	log.Println("启动 Elasticsearch 数据同步工作器")

	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("收到停止信号，退出同步工作器")
			return
		case <-s.stopChan:
			log.Println("收到停止信号，退出同步工作器")
			return
		case <-ticker.C:
			if err := s.SyncAllStatutes(ctx); err != nil {
				log.Printf("同步失败: %v", err)
			}
		}
	}
}

// StopSyncWorker 停止同步工作器
func (s *ElasticsearchSyncService) StopSyncWorker() {
	select {
	case s.stopChan <- true:
	default:
		// 已经有停止信号在等待
	}
}

// SyncAllStatutes 同步所有法条数据
func (s *ElasticsearchSyncService) SyncAllStatutes(ctx context.Context) error {
	log.Println("开始同步所有法条数据到 Elasticsearch...")

	startTime := time.Now()
	totalSynced := 0
	offset := 0

	for {
		// 分批获取法条数据
		statutes, err := s.statuteRepo.List(ctx, offset, s.batchSize)
		if err != nil {
			return fmt.Errorf("获取法条数据失败: %v", err)
		}

		if len(statutes) == 0 {
			break
		}

		// 转换为ES文档格式
		esDocs := make([]*elasticsearch.LegalStatuteDocument, 0, len(statutes))
		for _, statute := range statutes {
			esDoc := s.convertToESDocument(statute)
			esDocs = append(esDocs, esDoc)
		}

		// 批量索引到ES
		if err := s.esRepo.BulkIndexDocuments(ctx, esDocs); err != nil {
			return fmt.Errorf("批量索引失败: %v", err)
		}

		totalSynced += len(esDocs)
		log.Printf("已同步 %d 条法条数据", totalSynced)

		offset += s.batchSize
	}

	duration := time.Since(startTime)
	log.Printf("数据同步完成，总共同步了 %d 条法条，耗时: %v", totalSynced, duration)

	return nil
}

// SyncStatute 同步单个法条
func (s *ElasticsearchSyncService) SyncStatute(ctx context.Context, statuteID int) error {
	log.Printf("同步法条数据到 Elasticsearch: %d", statuteID)

	// 获取法条数据
	statute, err := s.statuteRepo.GetByID(ctx, statuteID)
	if err != nil {
		return fmt.Errorf("获取法条失败: %v", err)
	}

	// 转换为ES文档格式
	esDoc := s.convertToESDocument(statute)

	// 索引到ES
	if err := s.esRepo.IndexDocument(ctx, esDoc); err != nil {
		return fmt.Errorf("索引法条失败: %v", err)
	}

	log.Printf("法条同步成功: %s", statute.StatuteNumber)
	return nil
}

// SyncRange 同步指定范围的法条
func (s *ElasticsearchSyncService) SyncRange(ctx context.Context, fromID, toID int) error {
	log.Printf("同步指定范围的法条数据: %d - %d", fromID, toID)

	// 这里可以实现更复杂的范围同步逻辑
	// 暂时使用同步单个法条的方式

	// 获取总数量
	total, err := s.statuteRepo.Count(ctx)
	if err != nil {
		return fmt.Errorf("获取法条总数失败: %v", err)
	}

	syncedCount := 0
	for id := fromID; id <= toID && id <= int(total); id++ {
		if err := s.SyncStatute(ctx, id); err != nil {
			log.Printf("同步法条 %d 失败: %v", id, err)
			continue
		}
		syncedCount++
	}

	log.Printf("范围同步完成，成功同步 %d 条法条", syncedCount)
	return nil
}

// RebuildIndex 重建搜索索引
func (s *ElasticsearchSyncService) RebuildIndex(ctx context.Context) error {
	log.Println("开始重建 Elasticsearch 搜索索引...")

	// 1. 删除现有索引
	if err := s.esRepo.DeleteIndex(ctx); err != nil {
		log.Printf("删除索引失败: %v", err)
		// 不返回错误，继续创建新索引
	}

	// 2. 创建新索引
	if err := s.esRepo.CreateIndex(ctx); err != nil {
		return fmt.Errorf("创建索引失败: %v", err)
	}

	// 3. 同步所有数据
	if err := s.SyncAllStatutes(ctx); err != nil {
		return fmt.Errorf("同步数据失败: %v", err)
	}

	log.Println("搜索索引重建完成")
	return nil
}

// convertToESDocument 将数据库模型转换为ES文档
func (s *ElasticsearchSyncService) convertToESDocument(statute *models.LegalStatute) *elasticsearch.LegalStatuteDocument {
	doc := &elasticsearch.LegalStatuteDocument{
		ID:             statute.ID,
		StatuteNumber:  statute.StatuteNumber,
		Title:          statute.Title,
		Content:        statute.Content,
		LawName:        statute.LawName,
		Chapter:        statute.Chapter,
		Section:        statute.Section,
		Part:           statute.Part,
		Status:         statute.Status,
		HierarchyLevel: statute.HierarchyLevel,
		Tags:           statute.Tags,
		Keywords:       statute.Keywords,
		CreatedAt:      statute.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      statute.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		ContentLength:  len(statute.Content),
		ViewCount:      0,
		FavoriteCount:  0,
		SearchWeight:   s.calculateSearchWeight(statute),
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

// calculateSearchWeight 计算搜索权重
func (s *ElasticsearchSyncService) calculateSearchWeight(statute *models.LegalStatute) float64 {
	weight := 1.0

	// 根据层级调整权重
	if statute.HierarchyLevel <= 2 {
		weight += 0.5
	}

	// 根据标签数量调整权重
	if len(statute.Tags) > 0 {
		weight += float64(len(statute.Tags)) * 0.1
	}

	// 根据关键词数量调整权重
	if len(statute.Keywords) > 0 {
		weight += float64(len(statute.Keywords)) * 0.1
	}

	// 根据内容长度调整权重
	contentLength := len(statute.Content)
	if contentLength > 100 && contentLength < 1000 {
		weight += 0.2
	} else if contentLength >= 1000 {
		weight += 0.3
	}

	return weight
}

// GetSyncStatus 获取同步状态
func (s *ElasticsearchSyncService) GetSyncStatus(ctx context.Context) (*SyncStatus, error) {
	if s.esRepo == nil {
		return nil, fmt.Errorf("Elasticsearch 未配置")
	}

	// 检查Elasticsearch状态
	// 这里可以添加更多的状态检查逻辑
	isAvailable := true

	// 获取数据库法条总数
	dbCount, err := s.statuteRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取数据库法条总数失败: %v", err)
	}

	// 获取ES索引中的法条数量（这里需要实现ES查询逻辑）
	esCount := dbCount // 暂时假设数量一致

	lastSyncTime := time.Now() // 这里可以从配置或数据库中获取最后同步时间

	return &SyncStatus{
		IsAvailable:   isAvailable,
		DatabaseCount: dbCount,
		ESCount:       esCount,
		LastSyncTime:  lastSyncTime,
		SyncInterval:  s.syncInterval,
	}, nil
}

// SyncStatus 同步状态
type SyncStatus struct {
	IsAvailable   bool          `json:"isAvailable"`
	DatabaseCount int64         `json:"databaseCount"`
	ESCount       int64         `json:"esCount"`
	LastSyncTime  time.Time     `json:"lastSyncTime"`
	SyncInterval  time.Duration `json:"syncInterval"`
}

// SyncProgress 同步进度
type SyncProgress struct {
	TotalCount    int     `json:"totalCount"`
	SyncedCount   int     `json:"syncedCount"`
	Progress      float64 `json:"progress"`
	CurrentID     int     `json:"currentId"`
	EstimatedTime int     `json:"estimatedTimeSeconds"`
}

// GetSyncProgress 获取同步进度
func (s *ElasticsearchSyncService) GetSyncProgress(ctx context.Context) (*SyncProgress, error) {
	if s.statuteRepo == nil {
		return nil, fmt.Errorf("法条仓储未配置")
	}

	total, err := s.statuteRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取法条总数失败: %v", err)
	}

	// 这里可以实现更精确的进度计算
	// 暂时返回估算进度
	return &SyncProgress{
		TotalCount:  int(total),
		SyncedCount: 0,
		Progress:    0,
		CurrentID:   0,
	}, nil
}
