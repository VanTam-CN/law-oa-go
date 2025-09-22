package repositories

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// UserRepositoryImpl 用户数据仓库的GORM实现
type UserRepositoryImpl struct {
	*BaseRepository[models.User]
	db *gorm.DB
}

// NewUserRepository 创建用户数据仓库实例
func NewUserRepository(db *gorm.DB) UserRepository {
	return &UserRepositoryImpl{
		BaseRepository: NewBaseRepository[models.User](db),
		db:             db,
	}
}

// Create 创建用户
func (r *UserRepositoryImpl) Create(ctx context.Context, user *models.User) error {
	// 先检查邮箱是否已存在
	var existingUser models.User
	err := r.db.WithContext(ctx).Where("email = ?", user.Email).First(&existingUser).Error
	if err == nil {
		return NewRepositoryError("create", "user", ErrUserAlreadyExists)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return NewRepositoryError("create", "user", err)
	}

	return r.BaseRepository.Create(ctx, user)
}

// FindByID 根据ID查找用户
func (r *UserRepositoryImpl) FindByID(ctx context.Context, id uint) (*models.User, error) {
	user, err := r.BaseRepository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return nil, NewRepositoryErrorWithID("find", "user", id, ErrUserNotFound)
		}
		return nil, NewRepositoryErrorWithID("find", "user", id, err)
	}
	return user, nil
}

// FindByEmail 根据邮箱查找用户
func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewRepositoryError("find by email", "user", ErrUserNotFound)
		}
		return nil, NewRepositoryError("find by email", "user", err)
	}
	return &user, nil
}

// Update 更新用户信息
func (r *UserRepositoryImpl) Update(ctx context.Context, user *models.User) error {
	return r.BaseRepository.Update(ctx, user.ID, map[string]interface{}{
		"name":     user.Name,
		"email":    user.Email,
		"role":     user.Role,
		"phone":    user.Phone,
		"avatar":   user.Avatar,
		"status":   user.Status,
		"password": user.Password,
	})
}

// Delete 删除用户
func (r *UserRepositoryImpl) Delete(ctx context.Context, id uint) error {
	err := r.BaseRepository.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return NewRepositoryErrorWithID("delete", "user", id, ErrUserNotFound)
		}
		return NewRepositoryErrorWithID("delete", "user", id, err)
	}
	return nil
}

// List 用户列表查询
func (r *UserRepositoryImpl) List(ctx context.Context, params *UserListParams) ([]*models.User, int64, error) {
	page := params.Page
	if page <= 0 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	queryBuilder := NewQueryBuilder[models.User](r.db)

	if params.Role != "" {
		queryBuilder = queryBuilder.Where("role = ?", params.Role)
	}
	if params.Status != "" {
		queryBuilder = queryBuilder.Where("status = ?", params.Status)
	}
	if params.Search != "" {
		// 性能优化：智能搜索策略
		searchLower := strings.ToLower(params.Search)

		if len(params.Search) >= 3 {
			// 对于3个字符以上的搜索，使用后缀匹配以利用索引
			suffixTerm := searchLower + "%"
			queryBuilder = queryBuilder.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ?", suffixTerm, suffixTerm)
		} else {
			// 对于短搜索词，优先搜索邮箱字段（通常更精确）
			fullSearchTerm := "%" + searchLower + "%"
			queryBuilder = queryBuilder.Where("LOWER(email) LIKE ? OR LOWER(name) LIKE ?", fullSearchTerm, fullSearchTerm)
		}
	}

	// 获取总数
	total, err := queryBuilder.Count(ctx)
	if err != nil {
		return nil, 0, NewRepositoryError("list", "user", err)
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	users, err := queryBuilder.Offset(offset).Limit(pageSize).OrderDesc("created_at").Find(ctx)
	if err != nil {
		return nil, 0, NewRepositoryError("list", "user", err)
	}

	return users, total, nil
}
