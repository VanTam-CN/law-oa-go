package services

import (
	"context"
	"fmt"
	"sync"

	"github.com/law-oa-go/document-service/internal/repositories"
	"github.com/law-oa-go/document-service/internal/search"
	"github.com/sirupsen/logrus"
)

// serviceManager 服务管理器实现
type serviceManager struct {
	docService       DocumentService
	versionService    DocumentVersionService
	permissionService PermissionService
	auditService     AuditService
	userService      UserService
	roleService      RoleService
	searchService    SearchService
	storageService   StorageService
	notificationService NotificationService
	logger           *logrus.Logger
	once             sync.Once
}

// NewServiceManager 创建新的服务管理器
func NewServiceManager(
	repoManager repositories.RepositoryManager,
	storageConfig *StorageConfig,
	serviceConfig *ServiceConfig,
	logger *logrus.Logger,
) (*serviceManager, error) {
	// 创建存储服务
	var storageService StorageService
	var err error

	if storageConfig.Type == "minio" {
		storageService, err = NewMinioStorageService(
			storageConfig.MinIO.Endpoint,
			storageConfig.MinIO.AccessKey,
			storageConfig.MinIO.SecretKey,
			storageConfig.MinIO.Bucket,
			storageConfig.MinIO.Secure,
			logger,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create MinIO storage service: %w", err)
		}
	} else {
		// 使用内存存储（用于开发/测试）
		storageService = NewMemoryStorageService(logger)
	}

	sm := &serviceManager{
		storageService:   storageService,
		logger:           logger,
	}

	// 创建其他服务
	sm.docService = NewDocumentService(
		repoManager.Documents(),
		repoManager.Versions(),
		repoManager.Permissions(),
		repoManager.Audits(),
		repoManager.Users(),
		repoManager.Roles(),
		storageService,
		logger,
	)

	sm.permissionService = NewPermissionService(
		repoManager.Permissions(),
		repoManager.Documents(),
		repoManager.Users(),
		repoManager.Roles(),
		repoManager.Audits(),
		logger,
	)

	// 创建其他服务实例
	sm.userService = NewUserService(
		repoManager.Users(),
		repoManager.Roles(),
		repoManager.Permissions(),
		repoManager.Audits(),
		logger,
	)

	sm.roleService = NewRoleService(
		repoManager.Roles(),
		repoManager.Users(),
		repoManager.Audits(),
		logger,
	)

	sm.auditService = NewAuditService(
		repoManager.Audits(),
		repoManager.Users(),
		repoManager.Documents(),
		logger,
	)

	// 创建搜索服务（如果提供Elasticsearch配置）
	if serviceConfig != nil && serviceConfig.Search.Elasticsearch.Addresses != nil && len(serviceConfig.Search.Elasticsearch.Addresses) > 0 {
		// 初始化搜索服务
		searchStarter := &search.SearchStarter{}
		// 注意：这里简化处理，实际应该完整集成搜索服务
		logger.Info("Search service integration available (requires manual initialization)")
		sm.searchService = nil // 将在后续完整实现
	} else {
		sm.searchService = nil
	}

	sm.notificationService = NewNotificationService(
		repoManager.Notifications(),
		repoManager.Users(),
		repoManager.Audits(),
		logger,
	)

	// 文档版本服务需要单独实现
	// sm.versionService = NewVersionService(...)

	return sm, nil
}

// Documents 获取文档服务
func (sm *serviceManager) Documents() DocumentService {
	return sm.docService
}

// Versions 获取文档版本服务
func (sm *serviceManager) Versions() DocumentVersionService {
	return sm.versionService
}

// Permissions 获取权限服务
func (sm *serviceManager) Permissions() PermissionService {
	return sm.permissionService
}

// Audits 获取审计服务
func (sm *serviceManager) Audits() AuditService {
	return sm.auditService
}

// Users 获取用户服务
func (sm *serviceManager) Users() UserService {
	return sm.userService
}

// Roles 获取角色服务
func (sm *serviceManager) Roles() RoleService {
	return sm.roleService
}

// Search 获取搜索服务
func (sm *serviceManager) Search() SearchService {
	return sm.searchService
}

// Storage 获取存储服务
func (sm *serviceManager) Storage() StorageService {
	return sm.storageService
}

// Notifications 获取通知服务
func (sm *serviceManager) Notifications() NotificationService {
	return sm.notificationService
}

// Health 健康检查
func (sm *serviceManager) Health(ctx context.Context) (map[string]error, error) {
	health := make(map[string]error)

	// 检查文档服务
	if sm.docService != nil {
		health["document_service"] = nil
	} else {
		health["document_service"] = fmt.Errorf("document service not initialized")
	}

	// 检查存储服务
	if sm.storageService != nil {
		health["storage_service"] = nil
	} else {
		health["storage_service"] = fmt.Errorf("storage service not initialized")
	}

	// 检查权限服务
	if sm.permissionService != nil {
		health["permission_service"] = nil
	} else {
		health["permission_service"] = fmt.Errorf("permission service not initialized")
	}

	// 检查用户服务
	if sm.userService != nil {
		health["user_service"] = nil
	} else {
		health["user_service"] = fmt.Errorf("user service not initialized")
	}

	// 检查角色服务
	if sm.roleService != nil {
		health["role_service"] = nil
	} else {
		health["role_service"] = fmt.Errorf("role service not initialized")
	}

	// 检查审计服务
	if sm.auditService != nil {
		health["audit_service"] = nil
	} else {
		health["audit_service"] = fmt.Errorf("audit service not initialized")
	}

	// 检查通知服务
	if sm.notificationService != nil {
		health["notification_service"] = nil
	} else {
		health["notification_service"] = fmt.Errorf("notification service not initialized")
	}

	// 检查搜索服务
	if sm.searchService != nil {
		health["search_service"] = nil
	} else {
		health["search_service"] = fmt.Errorf("search service not initialized")
	}

	return health, nil
}

// Close 关闭服务
func (sm *serviceManager) Close() error {
	// 清理资源
	sm.logger.Info("Closing service manager")
	return nil
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type string `yaml:"type" json:"type"`
	MinIO struct {
		Endpoint  string `yaml:"endpoint" json:"endpoint"`
	AccessKey string `yaml:"access_key" json:"access_key"`
		SecretKey string `yaml:"secret_key" json:"secret_key"`
		Bucket    string `yaml:"bucket" json:"bucket"`
		Secure    bool   `yaml:"secure" json:"secure"`
	} `yaml:"minio" json:"minio"`
	Local struct {
		BasePath string `yaml:"base_path" json:"base_path"`
	} `yaml:"local" json:"local"`
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	Storage StorageConfig `yaml:"storage" json:"storage"`
	Auth    AuthConfig    `yaml:"auth" json:"auth"`
	Search  SearchConfig  `yaml:"search" json:"search"`
	Cache   CacheConfig   `yaml:"cache" json:"cache"`
	Logging LoggingConfig `yaml:"logging" json:"logging"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWT JWTConfig `yaml:"jwt" json:"jwt"`
}

type JWTConfig struct {
	Secret     string `yaml:"secret" json:"secret"`
	Expiry     int    `yaml:"expiry" json:"expiry"`
	Issuer     string `yaml:"issuer" json:"issuer"`
	RefreshExpiry int   `yaml:"refresh_expiry" json:"refresh_expiry"`
}

// SearchConfig 搜索配置
type SearchConfig struct {
	Elasticsearch ElasticsearchConfig `yaml:"elasticsearch" json:"elasticsearch"`
}

type ElasticsearchConfig struct {
	Addresses []string `yaml:"addresses" json:"addresses"`
	Username   string   `yaml:"username" json:"username"`
	Password   string   `yaml:"password" json:"password"`
	IndexName  string   `yaml:"index_name" json:"index_name"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Redis RedisConfig `yaml:"redis" json:"redis"`
}

type RedisConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Password string `yaml:"password" json:"password"`
	Database int    `yaml:"database" json:"database"`
	PoolSize int    `yaml:"pool_size" json:"pool_size"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string `yaml:"level" json:"level"`
	Format string `yaml:"format" json:"format"`
	Output string `yaml:"output" json:"output"`
}