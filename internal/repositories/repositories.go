package repositories

import (
	"context"

	"law-oa-go/internal/models"
)

// Repositories 统一的仓储集合，遵循Clean Architecture原则
type Repositories struct {
	UserRepo        UserRepository
	ClientRepo      ClientRepository
	DocumentRepo    DocumentRepository
	LawyerRepo      LawyerRepository
	ConflictRepo    BasicConflictRepository
	ConflictExtRepo *ConflictExtendedRepository
}

// NewRepositories 创建新的仓储集合
func NewRepositories(
	db interface{},
	redis RedisClient,
	esClient ElasticsearchClient,
) *Repositories {
	adapter := NewDBAdapter(db)
	gormDB := adapter.GetGormDB()

	return &Repositories{
		UserRepo:        NewUserRepository(gormDB),
		ClientRepo:      NewClientRepository(gormDB),
		DocumentRepo:    NewDocumentRepository(gormDB),
		LawyerRepo:      NewLawyerRepository(gormDB),
		ConflictRepo:    NewConflictRepository(gormDB, ToRedisClient(redis)),
		ConflictExtRepo: NewConflictExtendedRepository(db),
	}
}

// 基础接口定义
type DB interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) error
	QueryContext(ctx context.Context, query string, args ...interface{}) (interface{}, error)
	PrepareContext(ctx context.Context, query string) (interface{}, error)
}

type RedisClient interface {
	// Redis操作接口
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration interface{}) error
	Delete(ctx context.Context, key string) error
}

type ElasticsearchClient interface {
	// ES操作接口
	Index(index string, body interface{}) error
	Search(index string, query map[string]interface{}) (interface{}, error)
	Delete(index, id string) error
}

// 基础仓储接口
type BasicUserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uint) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, filter map[string]interface{}) ([]*models.User, error)
}

type BasicClientRepository interface {
	Create(ctx context.Context, client *models.Client) error
	GetByID(ctx context.Context, id uint) (*models.Client, error)
	Update(ctx context.Context, client *models.Client) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, filter map[string]interface{}) ([]*models.Client, error)
}

type BasicDocumentRepository interface {
	Upload(ctx context.Context, doc *models.Document) error
	GetByID(ctx context.Context, id uint) (*models.Document, error)
	Update(ctx context.Context, doc *models.Document) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, filter map[string]interface{}) ([]*models.Document, error)
}

type BasicLawyerRepository interface {
	Create(ctx context.Context, lawyer *models.User) error
	GetByID(ctx context.Context, id uint) (*models.User, error)
	Update(ctx context.Context, lawyer *models.User) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, filter map[string]interface{}) ([]*models.User, error)
}

// 事务管理接口
type TransactionManager interface {
	ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// 扩展的冲突检测仓储接口
type ConflictRepository interface {
	// 保存冲突检测记录
	SaveCheckRecord(ctx context.Context, record *models.ConflictCheckRecord) error
	// 获取冲突检测历史
	GetCheckHistory(ctx context.Context, clientID string, limit int) ([]*models.ConflictCheckRecord, error)
	// 获取潜在的冲突案件
	GetPotentialConflicts(ctx context.Context, lawyerID int, clientName string, opposingParty string) ([]*models.Case, error)
	// 获取冲突统计
	GetConflictStats(ctx context.Context, lawyerID int) (map[string]interface{}, error)
}

// 注意：EnhancedConflictRepository 接口已移动到 enhanced_conflict_repository.go
