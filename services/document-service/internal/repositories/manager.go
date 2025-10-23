package repositories

import (
	"context"
	"fmt"

	"github.com/law-oa-go/document-service/internal/database"
	"gorm.io/gorm"
)

// repositoryManager 仓库管理器实现
type repositoryManager struct {
	db                  *gorm.DB
	documentRepo        DocumentRepository
	versionRepo         DocumentVersionRepository
	permissionRepo      DocumentPermissionRepository
	auditRepo           DocumentAuditRepository
	userRepo            UserRepository
	roleRepo            RoleRepository
	userRoleRepo        UserRoleRepository
	searchRepo          SearchRepository
	transactionManager  TransactionManager
}

// NewRepositoryManager 创建新的仓库管理器
func NewRepositoryManager(db *gorm.DB, searchRepo SearchRepository) RepositoryManager {
	rm := &repositoryManager{
		db:                 db,
		documentRepo:       NewDocumentRepository(db),
		versionRepo:        NewDocumentVersionRepository(db),
		permissionRepo:     NewPermissionRepository(db),
		auditRepo:          NewAuditRepository(db),
		userRepo:           NewUserRepository(db),
		roleRepo:           NewRoleRepository(db),
		userRoleRepo:       NewUserRoleRepository(db),
		searchRepo:         searchRepo,
		transactionManager: NewTransactionManager(db),
	}

	return rm
}

// Documents 获取文档仓库
func (rm *repositoryManager) Documents() DocumentRepository {
	return rm.documentRepo
}

// Versions 获取文档版本仓库
func (rm *repositoryManager) Versions() DocumentVersionRepository {
	return rm.versionRepo
}

// Permissions 获取权限仓库
func (rm *repositoryManager) Permissions() DocumentPermissionRepository {
	return rm.permissionRepo
}

// Audits 获取审计仓库
func (rm *repositoryManager) Audits() DocumentAuditRepository {
	return rm.auditRepo
}

// Users 获取用户仓库
func (rm *repositoryManager) Users() UserRepository {
	return rm.userRepo
}

// Roles 获取角色仓库
func (rm *repositoryManager) Roles() RoleRepository {
	return rm.roleRepo
}

// UserRoles 获取用户角色关联仓库
func (rm *repositoryManager) UserRoles() UserRoleRepository {
	return rm.userRoleRepo
}

// Search 获取搜索仓库
func (rm *repositoryManager) Search() SearchRepository {
	return rm.searchRepo
}

// TransactionManager 获取事务管理器
func (rm *repositoryManager) TransactionManager() TransactionManager {
	return rm.transactionManager
}

// Health 健康检查
func (rm *repositoryManager) Health(ctx context.Context) error {
	// 检查数据库连接
	sqlDB, err := rm.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	// 检查搜索服务
	if rm.searchRepo != nil {
		if err := rm.searchRepo.Health(ctx); err != nil {
			return fmt.Errorf("search service health check failed: %w", err)
		}
	}

	return nil
}

// Close 关闭连接
func (rm *repositoryManager) Close() error {
	// GORM 数据库连接会自动关闭，但我们可以在这里做一些清理工作
	return nil
}

// NewRepositoryManagerFromDatabaseManager 从数据库管理器创建仓库管理器
func NewRepositoryManagerFromDatabaseManager(dbManager *database.Manager) RepositoryManager {
	var db *gorm.DB
	if dbManager != nil {
		db = dbManager.GetDB()
	}

	// 创建搜索仓库（这里先为nil，后面会实现）
	var searchRepo SearchRepository

	return NewRepositoryManager(db, searchRepo)
}

// WithTransaction 在事务中执行操作
func (rm *repositoryManager) WithTransaction(ctx context.Context, fn func(RepositoryManager) error) error {
	return rm.transactionManager.WithTransaction(ctx, func(txDB *gorm.DB) error {
		// 创建新的事务仓库管理器
		txRepoManager := &repositoryManager{
			db:                 txDB,
			documentRepo:       NewDocumentRepository(txDB),
			versionRepo:        NewDocumentVersionRepository(txDB),
			permissionRepo:     NewPermissionRepository(txDB),
			auditRepo:          NewAuditRepository(txDB),
			userRepo:           NewUserRepository(txDB),
			roleRepo:           NewRoleRepository(txDB),
			userRoleRepo:       NewUserRoleRepository(txDB),
			searchRepo:         rm.searchRepo, // 搜索仓库不使用事务
			transactionManager: NewTransactionManager(txDB),
		}

		return fn(txRepoManager)
	})
}

// GetDatabaseStats 获取数据库统计信息
func (rm *repositoryManager) GetDatabaseStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取连接统计
	sqlDB, err := rm.db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	dbStats := sqlDB.Stats()
	stats["database"] = map[string]interface{}{
		"open_connections": dbStats.OpenConnections,
		"in_use":          dbStats.InUse,
		"idle":            dbStats.Idle,
		"max_open":        dbStats.MaxOpenConnections,
	}

	// 获取文档统计
	var docCount int64
	if err := rm.db.WithContext(ctx).
		Model(&Document{}).
		Count(&docCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count documents: %w", err)
	}
	stats["total_documents"] = docCount

	// 获取版本统计
	var versionCount int64
	if err := rm.db.WithContext(ctx).
		Model(&DocumentVersion{}).
		Count(&versionCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count versions: %w", err)
	}
	stats["total_versions"] = versionCount

	// 获取权限统计
	var permissionCount int64
	if err := rm.db.WithContext(ctx).
		Model(&DocumentPermission{}).
		Count(&permissionCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count permissions: %w", err)
	}
	stats["total_permissions"] = permissionCount

	// 获取审计统计
	var auditCount int64
	if err := rm.db.WithContext(ctx).
		Model(&DocumentAudit{}).
		Count(&auditCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count audits: %w", err)
	}
	stats["total_audits"] = auditCount

	// 获取用户统计
	var userCount int64
	if err := rm.db.WithContext(ctx).
		Model(&User{}).
		Count(&userCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}
	stats["total_users"] = userCount

	// 获取角色统计
	var roleCount int64
	if err := rm.db.WithContext(ctx).
		Model(&Role{}).
		Count(&roleCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count roles: %w", err)
	}
	stats["total_roles"] = roleCount

	return stats, nil
}

// CleanupOldData 清理旧数据
func (rm *repositoryManager) CleanupOldData(ctx context.Context, beforeDate string) error {
	// 使用事务进行清理
	return rm.WithTransaction(ctx, func(txRepoManager RepositoryManager) error {
		// 清理旧的审计记录
		if err := txRepoManager.Audits().DeleteOldRecords(ctx, beforeDate); err != nil {
			return fmt.Errorf("failed to cleanup old audit records: %w", err)
		}

		// 可以在这里添加其他清理逻辑

		return nil
	})
}

// ValidateConstraints 验证数据库约束
func (rm *repositoryManager) ValidateConstraints(ctx context.Context) error {
	// 检查外键约束
	constraints := []string{
		"documents_ibfk_1",     // documents.created_by -> users.id
		"document_versions_ibfk_1", // document_versions.document_id -> documents.id
		"document_versions_ibfk_2", // document_versions.created_by -> users.id
		"document_permissions_ibfk_1", // document_permissions.document_id -> documents.id
		"document_permissions_ibfk_2", // document_permissions.user_id -> users.id
		"document_permissions_ibfk_3", // document_permissions.role_id -> roles.id
		"document_audits_ibfk_1",  // document_audits.document_id -> documents.id
		"document_audits_ibfk_2",  // document_audits.user_id -> users.id
		"user_roles_ibfk_1",       // user_roles.user_id -> users.id
		"user_roles_ibfk_2",       // user_roles.role_id -> roles.id
	}

	for _, constraint := range constraints {
		// 这里可以添加具体的约束验证逻辑
		_ = constraint
	}

	return nil
}

// BackupData 备份数据
func (rm *repositoryManager) BackupData(ctx context.Context, backupPath string) error {
	// 这个方法可以实现数据备份逻辑
	// 例如导出SQL文件或生成数据快照
	return fmt.Errorf("backup functionality not implemented yet")
}

// RestoreData 恢复数据
func (rm *repositoryManager) RestoreData(ctx context.Context, backupPath string) error {
	// 这个方法可以实现数据恢复逻辑
	// 例如从SQL文件导入或从快照恢复
	return fmt.Errorf("restore functionality not implemented yet")
}