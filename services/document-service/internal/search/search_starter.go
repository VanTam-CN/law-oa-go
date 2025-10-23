package search

import (
	"context"
	"fmt"
	"time"

	"github.com/law-oa-go/document-service/internal/repositories"
	"github.com/sirupsen/logrus"
)

// SearchStarter 搜索服务启动器
type SearchStarter struct {
	searchManager *SearchManager
	logger         *logrus.Logger
	config         *SearchConfig
}

// NewSearchStarter 创建搜索服务启动器
func NewSearchStarter(config *SearchConfig, logger *logrus.Logger) *SearchStarter {
	return &SearchStarter{
		config: config,
		logger: logger,
	}
}

// Initialize 初始化搜索服务
func (ss *SearchStarter) Initialize(docRepo repositories.DocumentRepository, userRepo repositories.UserRepository) error {
	ss.logger.Info("Initializing search service...")

	// 记录配置
	LogConfig(ss.config, ss.logger)

	// 验证配置
	if err := ValidateSearchConfig(ss.config); err != nil {
		return fmt.Errorf("invalid search config: %w", err)
	}

	// 创建搜索管理器
	searchManager, err := NewSearchManager(ss.config, docRepo, userRepo, ss.logger)
	if err != nil {
		return fmt.Errorf("failed to create search manager: %w", err)
	}

	ss.searchManager = searchManager

	// 执行初始健康检查
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := searchManager.HealthCheck(ctx); err != nil {
		ss.logger.WithError(err).Error("Initial health check failed")
		return fmt.Errorf("initial health check failed: %w", err)
	}

	ss.logger.Info("Search service initialized successfully")
	return nil
}

// GetSearchManager 获取搜索管理器
func (ss *SearchStarter) GetSearchManager() *SearchManager {
	return ss.searchManager
}

// HealthCheck 执行健康检查
func (ss *SearchStarter) HealthCheck(ctx context.Context) error {
	if ss.searchManager == nil {
		return fmt.Errorf("search manager not initialized")
	}

	return ss.searchManager.HealthCheck(ctx)
}

// IsHealthy 检查服务是否健康
func (ss *SearchStarter) IsHealthy() bool {
	if ss.searchManager == nil {
		return false
	}
	return ss.searchManager.IsHealthy()
}

// StartPeriodicTasks 启动定期任务
func (ss *SearchStarter) StartPeriodicTasks() {
	if ss.searchManager == nil {
		ss.logger.Error("Cannot start periodic tasks: search manager not initialized")
		return
	}

	// 启动索引同步（如果配置了）
	if ss.config.Indexing.SyncInterval > 0 {
		go ss.startPeriodicSync()
	}

	// 启动缓存预热（如果启用了缓存）
	if ss.config.Caching.Enabled {
		go ss.startCacheWarmUp()
	}

	// 启动健康检查监控
	go ss.startHealthMonitoring()
}

// Close 关闭搜索服务
func (ss *SearchStarter) Close() error {
	if ss.searchManager == nil {
		return nil
	}

	ss.logger.Info("Closing search service...")
	return ss.searchManager.Close()
}

// startPeriodicSync 启动定期同步
func (ss *SearchStarter) startPeriodicSync() {
	ticker := time.NewTicker(ss.config.Indexing.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ss.logger.Info("Starting periodic index sync")
			// 这里可以实现定期同步逻辑
			// 例如：同步最近变更的文档
		}
	}
}

// startCacheWarmUp 启动缓存预热
func (ss *SearchStarter) startCacheWarmUp() {
	// 等待服务完全启动
	time.Sleep(30 * time.Second)

	ss.logger.Info("Starting cache warm-up")

	// 常见搜索查询列表
	commonQueries := []string{
		"document",
		"file",
		"report",
		"contract",
		"agreement",
		"presentation",
		"spreadsheet",
		"pdf",
		"image",
		"video",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := ss.searchManager.WarmUpCache(ctx, commonQueries); err != nil {
		ss.logger.WithError(err).Error("Failed to warm up cache")
	} else {
		ss.logger.Info("Cache warm-up completed")
	}
}

// startHealthMonitoring 启动健康监控
func (ss *SearchStarter) startHealthMonitoring() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			if err := ss.HealthCheck(ctx); err != nil {
				ss.logger.WithError(err).Error("Health check failed")
			}
			cancel()

			// 记录配置健康状态
			health := CheckConfigHealth(ss.config)
			if health.Overall != "healthy" {
				ss.logger.WithFields(logrus.Fields{
					"overall": health.Overall,
					"errors":  health.Errors,
				}).Error("Search configuration issues detected")
			}
		}
	}
}

// ReindexAll 重新索引所有数据
func (ss *SearchStarter) ReindexAll(tenantID string) error {
	if ss.searchManager == nil {
		return fmt.Errorf("search manager not initialized")
	}

	ss.logger.WithField("tenant_id", tenantID).Info("Starting full reindex")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if err := ss.searchManager.Reindex(ctx, tenantID); err != nil {
		return fmt.Errorf("reindex failed: %w", err)
	}

	ss.logger.WithField("tenant_id", tenantID).Info("Reindex completed")
	return nil
}

// GetStats 获取搜索统计信息
func (ss *SearchStarter) GetStats(tenantID string) (*SearchStats, error) {
	if ss.searchManager == nil {
		return nil, fmt.Errorf("search manager not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stats, err := ss.searchManager.GetSearchStats(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get search stats: %w", err)
	}

	return stats, nil
}

// GetIndexStats 获取索引统计信息
func (ss *SearchStarter) GetIndexStats() (*IndexStats, error) {
	if ss.searchManager == nil {
		return nil, fmt.Errorf("search manager not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stats, err := ss.searchManager.GetIndexStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get index stats: %w", err)
	}

	return stats, nil
}

// GetConfig 获取搜索配置
func (ss *SearchStarter) GetConfig() *SearchConfig {
	return ss.config
}

// UpdateConfig 更新搜索配置
func (ss *SearchStarter) UpdateConfig(newConfig *SearchConfig) error {
	if err := ValidateSearchConfig(newConfig); err != nil {
		return fmt.Errorf("invalid search config: %w", err)
	}

	oldConfig := ss.config
	ss.config = newConfig

	// 如果关键配置发生变化，需要重启服务
	configChanged := false

	if oldConfig.Elasticsearch.Addresses != newConfig.Elasticsearch.Addresses ||
		oldConfig.Elasticsearch.IndexName != newConfig.Elasticsearch.IndexName ||
		oldConfig.Redis.Host != newConfig.Redis.Host ||
		oldConfig.Redis.Port != newConfig.Redis.Port ||
		oldConfig.Caching.Enabled != newConfig.Caching.Enabled {
		configChanged = true
	}

	if configChanged {
		ss.logger.Warn("Critical search configuration changed, service restart required")
		return fmt.Errorf("critical configuration changed, service restart required")
	}

	// 更新非关键配置
	if ss.searchManager != nil {
		// 这里可以实现动态配置更新逻辑
		ss.logger.Info("Search configuration updated")
	}

	return nil
}

// CreateTenantIndex 为租户创建专用索引
func (ss *SearchStarter) CreateTenantIndex(tenantID string) error {
	if ss.searchManager == nil {
		return fmt.Errorf("search manager not initialized")
	}

	ss.logger.WithField("tenant_id", tenantID).Info("Creating tenant index")

	// 这里可以实现租户专用索引创建逻辑
	// 例如：为每个租户创建单独的索引模板

	return nil
}

// DropTenantIndex 删除租户索引
func (ss *SearchStarter) DropTenantIndex(tenantID string) error {
	if ss.searchManager == nil {
		return fmt.Errorf("search manager not initialized")
	}

	ss.logger.WithField("tenant_id", tenantID).Info("Dropping tenant index")

	// 这里可以实现租户索引删除逻辑

	return nil
}

// BackupIndex 备份索引
func (ss *SearchStarter) BackupIndex(sourceIndex, targetIndex string) error {
	if ss.searchManager == nil {
		return fmt.Errorf("search manager not initialized")
	}

	ss.logger.WithFields(logrus.Fields{
		"source_index": sourceIndex,
		"target_index": targetIndex,
	}).Info("Starting index backup")

	// 这里可以实现索引备份逻辑
	// 例如：使用Elasticsearch的reindex API

	return nil
}

// RestoreIndex 恢复索引
func (ss *SearchStarter) RestoreIndex(sourceIndex, targetIndex string) error {
	if ss.searchManager == nil {
		return fmt.Errorf("search manager not initialized")
	}

	ss.logger.WithFields(logrus.Fields{
		"source_index": sourceIndex,
		"target_index": targetIndex,
	}).Info("Starting index restore")

	// 这里可以实现索引恢复逻辑

	return nil
}