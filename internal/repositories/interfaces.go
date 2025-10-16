package repositories

import (
	"errors"
	"fmt"

	"context"

	"law-oa-go/internal/models"
)

// Repository Sentinel Errors
var (
	// User Repository Errors
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserInvalid       = errors.New("invalid user data")

	// Client Repository Errors
	ErrClientNotFound      = errors.New("client not found")
	ErrClientAlreadyExists = errors.New("client already exists")
	ErrClientInvalid       = errors.New("invalid client data")

	// Case Repository Errors
	ErrCaseNotFound      = errors.New("case not found")
	ErrCaseAlreadyExists = errors.New("case already exists")
	ErrCaseInvalid       = errors.New("invalid case data")
	ErrCaseConflict      = errors.New("case conflict")
)

// RepositoryError provides more context for repository operations
type RepositoryError struct {
	Operation string
	Entity    string
	ID        interface{}
	Err       error
}

func (e *RepositoryError) Error() string {
	if e.ID != nil {
		return fmt.Sprintf("repository %s failed for %s with ID %v: %v", e.Operation, e.Entity, e.ID, e.Err)
	}
	return fmt.Sprintf("repository %s failed for %s: %v", e.Operation, e.Entity, e.Err)
}

func (e *RepositoryError) Unwrap() error {
	return e.Err
}

// NewRepositoryError creates a new repository error
func NewRepositoryError(operation, entity string, err error) *RepositoryError {
	return &RepositoryError{
		Operation: operation,
		Entity:    entity,
		Err:       err,
	}
}

// NewRepositoryErrorWithID creates a new repository error with ID
func NewRepositoryErrorWithID(operation, entity string, id interface{}, err error) *RepositoryError {
	return &RepositoryError{
		Operation: operation,
		Entity:    entity,
		ID:        id,
		Err:       err,
	}
}

// UserRepository 用户数据仓库接口
type UserRepository interface {
	// Create 创建用户
	Create(ctx context.Context, user *models.User) error
	// FindByID 根据ID查找用户
	FindByID(ctx context.Context, id uint) (*models.User, error)
	// FindByEmail 根据邮箱查找用户
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	// FindExistingEmails 批量查找已存在的邮箱（解决N+1查询问题）
	FindExistingEmails(ctx context.Context, emails []string) ([]string, error)
	// BatchCreate 批量创建用户（优化性能）
	BatchCreate(ctx context.Context, users []*models.User) error
	// Update 更新用户信息
	Update(ctx context.Context, user *models.User) error
	// Delete 删除用户
	Delete(ctx context.Context, id uint) error
	// List 用户列表查询
	List(ctx context.Context, params *UserListParams) ([]*models.User, int64, error)
}

// ClientRepository 客户数据仓库接口
type ClientRepository interface {
	// Create 创建客户
	Create(ctx context.Context, client *models.Client) error
	// FindByID 根据ID查找客户
	FindByID(ctx context.Context, id uint) (*models.Client, error)
	// FindByEmail 根据邮箱查找客户
	FindByEmail(ctx context.Context, email string) (*models.Client, error)
	// Update 更新客户信息
	Update(ctx context.Context, client *models.Client) error
	// Delete 删除客户
	Delete(ctx context.Context, id uint) error
	// List 客户列表查询
	List(ctx context.Context, params *ClientListParams) ([]*models.Client, int64, error)
	// GetStats 获取客户统计信息
	GetStats(ctx context.Context) (*ClientStats, error)
}

// CaseRepository 案件数据仓库接口
type CaseRepository interface {
	// Create 创建案件
	Create(ctx context.Context, caseModel *models.Case) error
	// FindByID 根据ID查找案件
	FindByID(ctx context.Context, id uint) (*models.Case, error)
	// Update 更新案件信息
	Update(ctx context.Context, caseModel *models.Case) error
	// Delete 删除案件
	Delete(ctx context.Context, id uint) error
	// List 案件列表查询
	List(ctx context.Context, params *CaseListParams) ([]*models.Case, int64, error)
	// GetStats 获取案件统计信息
	GetStats(ctx context.Context) (*CaseStats, error)
	// AssignLawyer 分配律师
	AssignLawyer(ctx context.Context, caseID, lawyerID uint) error
	// UpdateStatus 更新案件状态
	UpdateStatus(ctx context.Context, caseID uint, status string) error
}

// UserListParams 用户列表查询参数
type UserListParams struct {
	Page     int
	PageSize int
	Role     string
	Status   string
	Search   string
}

// ClientListParams 客户列表查询参数
type ClientListParams struct {
	Page     int
	PageSize int
	Status   string
	Search   string
	Type     string
	Company  string
}

// CaseListParams 案件列表查询参数
type CaseListParams struct {
	Page     int
	PageSize int
	Status   string
	CaseType string
	Priority string
	ClientID uint
	LawyerID uint
	Search   string
}

// ClientStats 客户统计信息
type ClientStats struct {
	TotalClients        int64
	ActiveClients       int64
	InactiveClients     int64
	NewClientsThisMonth int64
}

// CaseStats 案件统计信息
type CaseStats struct {
	TotalCases     int64
	PendingCases   int64
	ActiveCases    int64
	ClosedCases    int64
	SuspendedCases int64
	HighPriority   int64
	UrgentCases    int64
}


