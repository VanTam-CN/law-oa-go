package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/law-oa-go/document-service/internal/models"
	"github.com/law-oa-go/document-service/internal/repositories"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/sirupsen/logrus"
)

// IndexManager 索引管理器
type IndexManager struct {
	esClient      *ElasticsearchClient
	docRepo       repositories.DocumentRepository
	userRepo      repositories.UserRepository
	logger        *logrus.Logger
	syncBatchSize int
}

// NewIndexManager 创建索引管理器
func NewIndexManager(
	esClient *ElasticsearchClient,
	docRepo repositories.DocumentRepository,
	userRepo repositories.UserRepository,
	logger *logrus.Logger,
) *IndexManager {
	return &IndexManager{
		esClient:      esClient,
		docRepo:       docRepo,
		userRepo:      userRepo,
		logger:        logger,
		syncBatchSize: 100, // 批量同步大小
	}
}

// IndexDocument 索引单个文档
func (im *IndexManager) IndexDocument(ctx context.Context, document *models.Document) error {
	// 构建搜索文档
	searchDoc := im.buildSearchDocument(document)
	if searchDoc == nil {
		return fmt.Errorf("failed to build search document")
	}

	// 序列化文档
	docJSON, err := json.Marshal(searchDoc)
	if err != nil {
		return fmt.Errorf("failed to marshal search document: %w", err)
	}

	// 构建索引请求
	req := esapi.IndexRequest{
		Index:      im.esClient.GetIndexName(),
		DocumentID: fmt.Sprintf("%d", document.ID),
		Body:       bytes.NewReader(docJSON),
		Refresh:    "true",
		OpType:     "index",
	}

	// 执行索引请求
	res, err := req.Do(ctx, im.esClient.GetClient())
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("document indexing failed: %s", res.Status())
	}

	im.logger.WithFields(logrus.Fields{
		"document_id":   document.ID,
		"document_name": document.Name,
		"tenant_id":     document.TenantID,
	}).Debug("Document indexed successfully")

	return nil
}

// UpdateDocument 更新文档索引
func (im *IndexManager) UpdateDocument(ctx context.Context, document *models.Document) error {
	// 更新操作与索引操作相同，使用op_type=index会自动更新
	return im.IndexDocument(ctx, document)
}

// DeleteDocument 删除文档索引
func (im *IndexManager) DeleteDocument(ctx context.Context, documentID uint) error {
	req := esapi.DeleteRequest{
		Index:      im.esClient.GetIndexName(),
		DocumentID: fmt.Sprintf("%d", documentID),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, im.esClient.GetClient())
	if err != nil {
		return fmt.Errorf("failed to delete document from index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("document deletion failed: %s", res.Status())
	}

	im.logger.WithField("document_id", documentID).Debug("Document deleted from index")

	return nil
}

// BulkIndexDocuments 批量索引文档
func (im *IndexManager) BulkIndexDocuments(ctx context.Context, documents []*models.Document) error {
	if len(documents) == 0 {
		return nil
	}

	// 分批处理
	for i := 0; i < len(documents); i += im.syncBatchSize {
		end := i + im.syncBatchSize
		if end > len(documents) {
			end = len(documents)
		}

		batch := documents[i:end]
		if err := im.bulkIndexBatch(ctx, batch); err != nil {
			im.logger.WithError(err).WithField("batch_size", len(batch)).Error("Failed to index batch")
			return err
		}
	}

	im.logger.WithField("total_documents", len(documents)).Info("Bulk indexing completed")

	return nil
}

// bulkIndexBatch 批量索引单个批次
func (im *IndexManager) bulkIndexBatch(ctx context.Context, documents []*models.Document) error {
	var buf bytes.Buffer

	for _, document := range documents {
		// 构建搜索文档
		searchDoc := im.buildSearchDocument(document)
		if searchDoc == nil {
			continue
		}

		// 构建索引操作
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": im.esClient.GetIndexName(),
				"_id":    fmt.Sprintf("%d", document.ID),
			},
		}

		metaBytes, err := json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("failed to marshal meta: %w", err)
		}

		docBytes, err := json.Marshal(searchDoc)
		if err != nil {
			return fmt.Errorf("failed to marshal document: %w", err)
		}

		buf.Write(metaBytes)
		buf.WriteByte('\n')
		buf.Write(docBytes)
		buf.WriteByte('\n')
	}

	// 执行批量请求
	req := esapi.BulkRequest{
		Body:    &buf,
		Refresh: "true",
	}

	res, err := req.Do(ctx, im.esClient.GetClient())
	if err != nil {
		return fmt.Errorf("bulk request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk indexing failed: %s", res.Status())
	}

	// 解析响应以检查错误
	var bulkRes map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&bulkRes); err != nil {
		return fmt.Errorf("failed to decode bulk response: %w", err)
	}

	if errors, ok := bulkRes["errors"].(bool); ok && errors {
		if items, ok := bulkRes["items"].([]interface{}); ok {
			for i, item := range items {
				if index, ok := item.(map[string]interface{})["index"].(map[string]interface{}); ok {
					if status, ok := index["status"].(float64); ok && status >= 400 {
						errorMsg := "unknown error"
						if error, ok := index["error"].(map[string]interface{}); ok {
							if reason, ok := error["reason"].(string); ok {
								errorMsg = reason
							}
						}
						im.logger.WithFields(logrus.Fields{
							"document_id": i,
							"error":       errorMsg,
							"status":      status,
						}).Error("Document indexing failed")
					}
				}
			}
		}
	}

	return nil
}

// SyncAllDocuments 同步所有文档到索引
func (im *IndexManager) SyncAllDocuments(ctx context.Context, tenantID string) error {
	im.logger.WithField("tenant_id", tenantID).Info("Starting full document sync")

	// 先删除现有索引（重新创建）
	if err := im.recreateIndex(ctx); err != nil {
		return fmt.Errorf("failed to recreate index: %w", err)
	}

	offset := 0
	totalSynced := 0

	for {
		// 分批获取文档
		documents, total, err := im.docRepo.List(ctx, repositories.DocumentListOptions{
			TenantID: tenantID,
			Limit:    im.syncBatchSize,
			Offset:   offset,
		})

		if err != nil {
			return fmt.Errorf("failed to fetch documents batch: %w", err)
		}

		if len(documents) == 0 {
			break
		}

		// 批量索引
		if err := im.BulkIndexDocuments(ctx, documents); err != nil {
			return fmt.Errorf("failed to index batch at offset %d: %w", offset, err)
		}

		totalSynced += len(documents)
		offset += im.syncBatchSize

		im.logger.WithFields(logrus.Fields{
			"synced":     totalSynced,
			"total":      total,
			"progress":   fmt.Sprintf("%.2f%%", float64(totalSynced)/float64(total)*100),
		}).Info("Sync progress")

		// 检查是否完成
		if int64(totalSynced) >= total {
			break
		}
	}

	im.logger.WithFields(logrus.Fields{
		"tenant_id":   tenantID,
		"total_synced": totalSynced,
	}).Info("Full document sync completed")

	return nil
}

// SyncTenantDocuments 同步指定租户的文档
func (im *IndexManager) SyncTenantDocuments(ctx context.Context, tenantID string) error {
	return im.SyncAllDocuments(ctx, tenantID)
}

// recreateIndex 重新创建索引
func (im *IndexManager) recreateIndex(ctx context.Context) error {
	// 删除现有索引
	req := esapi.IndicesDeleteRequest{
		Index: []string{im.esClient.GetIndexName()},
	}

	res, err := req.Do(ctx, im.esClient.GetClient())
	if err != nil {
		return fmt.Errorf("failed to delete index: %w", err)
	}
	defer res.Body.Close()

	// 忽略404错误（索引不存在）
	if res.IsError() && res.StatusCode != 404 {
		im.logger.WithField("status", res.Status()).Warn("Index deletion returned error")
	}

	// 重新创建索引
	if err := im.reinitializeIndex(); err != nil {
		return fmt.Errorf("failed to reinitialize index: %w", err)
	}

	return nil
}

// reinitializeIndex 重新初始化索引
func (im *IndexManager) reinitializeIndex() error {
	// 使用elasticsearch_client.go中的初始化逻辑
	// 这里简化处理，实际应该委托给esClient
	indexName := im.esClient.GetIndexName()
	client := im.esClient.GetClient()

	// 构建索引配置
	settings := map[string]interface{}{
		"number_of_shards":   1,
		"number_of_replicas": 0, // 单节点环境
		"refresh_interval":   "1s",
		"max_result_window":  50000,
		"analysis": map[string]interface{}{
			"analyzer": map[string]interface{}{
				"document_analyzer": map[string]interface{}{
					"type":      "custom",
					"tokenizer": "standard",
					"filter": []string{
						"lowercase",
						"stop",
						"snowball",
					},
				},
				"search_analyzer": map[string]interface{}{
					"type":      "custom",
					"tokenizer": "standard",
					"filter": []string{
						"lowercase",
						"stop",
					},
				},
			},
		},
	}

	mappings := im.buildDocumentMapping()

	indexConfig := map[string]interface{}{
		"settings": settings,
		"mappings": mappings,
	}

	configJSON, err := json.Marshal(indexConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal index config: %w", err)
	}

	req := esapi.IndicesCreateRequest{
		Index: indexName,
		Body:  bytes.NewReader(configJSON),
	}

	res, err := req.Do(context.Background(), client)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("index creation failed: %s", res.Status())
	}

	im.logger.Info("Index recreated successfully")

	return nil
}

// buildSearchDocument 构建搜索文档
func (im *IndexManager) buildSearchDocument(document *models.Document) map[string]interface{} {
	if document == nil {
		return nil
	}

	// 获取创建者信息
	var creatorName string
	if document.CreatedBy > 0 {
		if creator, err := im.userRepo.GetByID(context.Background(), document.CreatedBy); err == nil {
			creatorName = creator.Username
		}
	}

	// 构建基础搜索文档
	searchDoc := map[string]interface{}{
		"id":           document.ID,
		"uuid":         document.UUID,
		"name":         document.Name,
		"description":  document.Description,
		"category":     document.Category,
		"tenant_id":    document.TenantID,
		"created_by":   document.CreatedBy,
		"creator_name": creatorName,
		"created_at":   document.CreatedAt.Format(time.RFC3339),
		"updated_at":   document.UpdatedAt.Format(time.RFC3339),
		"mime_type":    document.MIMEType,
		"size":         document.Size,
		"version":      document.CurrentVersion,
		"status":       "active", // 默认状态
		"access_level": "private", // 默认访问级别
		"priority":     0, // 默认优先级
	}

	// 添加标签（如果有）
	if document.Tags != nil {
		searchDoc["tags"] = document.Tags
	}

	// 添加文件哈希
	if document.FileHash != "" {
		searchDoc["file_hash"] = document.FileHash
	}

	// 添加元数据
	if document.Metadata != nil {
		searchDoc["metadata"] = document.Metadata
	}

	return searchDoc
}

// buildDocumentMapping 构建文档映射
func (im *IndexManager) buildDocumentMapping() map[string]interface{} {
	return map[string]interface{}{
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type": "integer",
			},
			"uuid": map[string]interface{}{
				"type": "keyword",
			},
			"name": map[string]interface{}{
				"type": "text",
				"fields": map[string]interface{}{
					"keyword": map[string]interface{}{
						"type":         "keyword",
						"ignore_above": 256,
					},
					"suggest": map[string]interface{}{
						"type":           "completion",
						"preserve_separators": true,
						"preserve_position_increments": true,
						"max_input_length": 50,
					},
					"edge_ngram": map[string]interface{}{
						"type":          "text",
						"analyzer":      "document_analyzer",
						"search_analyzer": "search_analyzer",
					},
				},
				"analyzer": "document_analyzer",
				"search_analyzer": "search_analyzer",
			},
			"description": map[string]interface{}{
				"type": "text",
				"fields": map[string]interface{}{
					"keyword": map[string]interface{}{
						"type":         "keyword",
						"ignore_above": 512,
					},
				},
				"analyzer": "document_analyzer",
				"search_analyzer": "search_analyzer",
			},
			"category": map[string]interface{}{
				"type": "keyword",
				"fields": map[string]interface{}{
					"text": map[string]interface{}{
						"type":     "text",
						"analyzer": "keyword",
					},
				},
			},
			"tags": map[string]interface{}{
				"type": "keyword",
			},
			"tenant_id": map[string]interface{}{
				"type": "keyword",
			},
			"created_by": map[string]interface{}{
				"type": "integer",
			},
			"creator_name": map[string]interface{}{
				"type": "text",
				"fields": map[string]interface{}{
					"keyword": map[string]interface{}{
						"type": "keyword",
					},
				},
				"analyzer": "document_analyzer",
			},
			"created_at": map[string]interface{}{
				"type":   "date",
				"format": "strict_date_optional_time||epoch_millis",
			},
			"updated_at": map[string]interface{}{
				"type":   "date",
				"format": "strict_date_optional_time||epoch_millis",
			},
			"mime_type": map[string]interface{}{
				"type": "keyword",
			},
			"size": map[string]interface{}{
				"type": "long",
			},
			"file_hash": map[string]interface{}{
				"type": "keyword",
			},
			"version": map[string]interface{}{
				"type": "integer",
			},
			"status": map[string]interface{}{
				"type": "keyword",
			},
			"access_level": map[string]interface{}{
				"type": "keyword",
			},
			"priority": map[string]interface{}{
				"type": "integer",
			},
			"metadata": map[string]interface{}{
				"type":    "object",
				"dynamic": true,
			},
		},
	}
}

// GetIndexStats 获取索引统计
func (im *IndexManager) GetIndexStats(ctx context.Context) (*IndexStats, error) {
	req := esapi.IndicesStatsRequest{
		Index: []string{im.esClient.GetIndexName()},
	}

	res, err := req.Do(ctx, im.esClient.GetClient())
	if err != nil {
		return nil, fmt.Errorf("failed to get index stats: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("index stats request failed: %s", res.Status())
	}

	var stats map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode index stats: %w", err)
	}

	indexStats := &IndexStats{
		IndexName:    im.esClient.GetIndexName(),
		DocumentCount: 0,
		StoreSize:     0,
	}

	if indices, ok := stats["indices"].(map[string]interface{}); ok {
		if index, ok := indices[im.esClient.GetIndexName()].(map[string]interface{}); ok {
			if total, ok := index["total"].(map[string]interface{}); ok {
				if docs, ok := total["docs"].(map[string]interface{}); ok {
					if count, ok := docs["count"].(float64); ok {
						indexStats.DocumentCount = int64(count)
					}
				}
				if store, ok := total["store"].(map[string]interface{}); ok {
					if size, ok := store["size_in_bytes"].(float64); ok {
						indexStats.StoreSize = int64(size)
					}
				}
			}
		}
	}

	return indexStats, nil
}

// IndexStats 索引统计
type IndexStats struct {
	IndexName    string `json:"index_name"`
	DocumentCount int64 `json:"document_count"`
	StoreSize     int64 `json:"store_size"`
}