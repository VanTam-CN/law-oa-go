package database

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// Manager 数据库管理器，统一管理所有数据存储连接
type Manager struct {
	MySQL        *Database
	Redis        *RedisClient
	Elasticsearch *ElasticsearchClient
	Cache        *CacheManager
	config       *Config
	mu           sync.RWMutex
}

// Config 统一配置
type Config struct {
	MySQL        *MySQLConfig        `mapstructure:"mysql"`
	Redis        *RedisConfig        `mapstructure:"redis"`
	Elasticsearch *ElasticsearchConfig `mapstructure:"elasticsearch"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		MySQL:        DefaultMySQLConfig(),
		Redis:        DefaultRedisConfig(),
		Elasticsearch: DefaultElasticsearchConfig(),
	}
}

// NewManager 创建新的数据库管理器
func NewManager(config *Config) (*Manager, error) {
	if config == nil {
		config = DefaultConfig()
	}

	manager := &Manager{
		config: config,
	}

	// 初始化MySQL
	if config.MySQL != nil {
		mysql, err := NewDatabase(config.MySQL)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize MySQL: %w", err)
		}
		manager.MySQL = mysql
		log.Println("MySQL connected successfully")
	}

	// 初始化Redis
	if config.Redis != nil {
		redis, err := NewRedisClient(config.Redis)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Redis: %w", err)
		}
		manager.Redis = redis
		manager.Cache = NewCacheManager(redis, "doc_service")
		log.Println("Redis connected successfully")
	}

	// 初始化Elasticsearch
	if config.Elasticsearch != nil {
		es, err := NewElasticsearchClient(config.Elasticsearch)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Elasticsearch: %w", err)
		}
		manager.Elasticsearch = es
		log.Println("Elasticsearch connected successfully")
	}

	return manager, nil
}

// Close 关闭所有数据库连接
func (m *Manager) Close() error {
	var errors []error

	if m.MySQL != nil {
		if err := m.MySQL.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close MySQL: %w", err))
		}
	}

	if m.Redis != nil {
		if err := m.Redis.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close Redis: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors occurred while closing databases: %v", errors)
	}

	log.Println("All database connections closed successfully")
	return nil
}

// Health 健康检查
func (m *Manager) Health() map[string]error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	health := make(map[string]error)
	wg := sync.WaitGroup{}
	results := make(chan map[string]error, 3)

	// 检查MySQL健康状态
	if m.MySQL != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := m.MySQL.Health()
			results <- map[string]error{"mysql": err}
		}()
	}

	// 检查Redis健康状态
	if m.Redis != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := m.Redis.Health()
			results <- map[string]error{"redis": err}
		}()
	}

	// 检查Elasticsearch健康状态
	if m.Elasticsearch != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := m.Elasticsearch.Health()
			results <- map[string]error{"elasticsearch": err}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		for key, err := range result {
			if err != nil {
				health[key] = err
			} else {
				health[key] = nil // 表示健康
			}
		}
	}

	return health
}

// Stats 获取所有数据库统计信息
func (m *Manager) Stats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]interface{})

	if m.MySQL != nil {
		stats["mysql"] = m.MySQL.Stats()
	}

	if m.Redis != nil {
		stats["redis"] = m.Redis.Stats()
	}

	if m.Elasticsearch != nil {
		esStats, err := m.Elasticsearch.Stats()
		if err != nil {
			stats["elasticsearch"] = map[string]interface{}{
				"error": err.Error(),
			}
		} else {
			stats["elasticsearch"] = esStats
		}
	}

	return stats
}

// GetMySQL 获取MySQL数据库实例
func (m *Manager) GetMySQL() *Database {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.MySQL
}

// GetRedis 获取Redis客户端实例
func (m *Manager) GetRedis() *RedisClient {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Redis
}

// GetElasticsearch 获取Elasticsearch客户端实例
func (m *Manager) GetElasticsearch() *ElasticsearchClient {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Elasticsearch
}

// GetCache 获取缓存管理器实例
func (m *Manager) GetCache() *CacheManager {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Cache
}

// GetDB 获取GORM数据库实例（为了向后兼容）
func (m *Manager) GetDB() *gorm.DB {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.MySQL != nil {
		return m.MySQL.GetDB()
	}
	return nil
}

// AutoMigrate 自动迁移数据库表
func (m *Manager) AutoMigrate(models ...interface{}) error {
	if m.MySQL == nil {
		return fmt.Errorf("MySQL not initialized")
	}
	return m.MySQL.AutoMigrate(models...)
}

// CreateIndexes 创建数据库索引
func (m *Manager) CreateIndexes() error {
	if m.MySQL == nil {
		return fmt.Errorf("MySQL not initialized")
	}
	return m.MySQL.CreateIndexes()
}

// InitializeElasticsearch 初始化Elasticsearch索引
func (m *Manager) InitializeElasticsearch(ctx context.Context) error {
	if m.Elasticsearch == nil {
		return fmt.Errorf("Elasticsearch not initialized")
	}

	// 文档索引映射
	documentMapping := map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
			"analysis": map[string]interface{}{
				"analyzer": map[string]interface{}{
					"document_analyzer": map[string]interface{}{
						"type":      "custom",
						"tokenizer": "ik_max_word",
						"filter":    []string{"lowercase", "stop"},
					},
				},
			},
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "keyword",
				},
				"tenant_id": map[string]interface{}{
					"type": "keyword",
				},
				"name": map[string]interface{}{
					"type":     "text",
					"analyzer": "document_analyzer",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"description": map[string]interface{}{
					"type":     "text",
					"analyzer": "document_analyzer",
				},
				"content": map[string]interface{}{
					"type":     "text",
					"analyzer": "document_analyzer",
				},
				"tags": map[string]interface{}{
					"type": "keyword",
				},
				"category": map[string]interface{}{
					"type": "keyword",
				},
				"entity_type": map[string]interface{}{
					"type": "keyword",
				},
				"entity_id": map[string]interface{}{
					"type": "integer",
				},
				"mime_type": map[string]interface{}{
					"type": "keyword",
				},
				"size": map[string]interface{}{
					"type": "long",
				},
				"created_by": map[string]interface{}{
					"type": "integer",
				},
				"created_at": map[string]interface{}{
					"type":   "date",
					"format": "yyyy-MM-dd HH:mm:ss",
				},
				"updated_at": map[string]interface{}{
					"type":   "date",
					"format": "yyyy-MM-dd HH:mm:ss",
				},
				"version": map[string]interface{}{
					"type": "integer",
				},
				"file_hash": map[string]interface{}{
					"type": "keyword",
				},
			},
		},
	}

	// 创建文档索引
	if err := m.Elasticsearch.CreateIndex(ctx, "documents", documentMapping); err != nil {
		return fmt.Errorf("failed to create documents index: %w", err)
	}

	log.Println("Elasticsearch indices created successfully")
	return nil
}

// HealthChecker 健康检查器
type HealthChecker struct {
	manager *Manager
	ctx     context.Context
	cancel  context.CancelFunc
	ticker  *time.Ticker
}

// NewHealthChecker 创建新的健康检查器
func NewHealthChecker(manager *Manager, interval time.Duration) *HealthChecker {
	ctx, cancel := context.WithCancel(context.Background())
	return &HealthChecker{
		manager: manager,
		ctx:     ctx,
		cancel:  cancel,
		ticker:  time.NewTicker(interval),
	}
}

// Start 开始健康检查
func (h *HealthChecker) Start() {
	go func() {
		for {
			select {
			case <-h.ctx.Done():
				return
			case <-h.ticker.C:
				health := h.manager.Health()
				for component, err := range health {
					if err != nil {
						log.Printf("Health check failed for %s: %v", component, err)
					} else {
						log.Printf("Health check passed for %s", component)
					}
				}
			}
		}
	}()
	log.Println("Health checker started")
}

// Stop 停止健康检查
func (h *HealthChecker) Stop() {
	h.cancel()
	h.ticker.Stop()
	log.Println("Health checker stopped")
}

// TransactionManager 事务管理器
type TransactionManager struct {
	db *gorm.DB
}

// NewTransactionManager 创建新的事务管理器
func NewTransactionManager(db *gorm.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

// WithTransaction 在事务中执行函数
func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	tx := tm.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// BeginTx 开始事务
func (tm *TransactionManager) BeginTx(ctx context.Context) (*gorm.DB, error) {
	tx := tm.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	return tx, nil
}