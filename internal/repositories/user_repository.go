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
		// 输入验证和清理（防止SQL注入和XSS）
		cleanSearch := r.sanitizeSearchInput(params.Search)
		searchLower := strings.ToLower(cleanSearch)

		if len(cleanSearch) >= 3 {
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

// FindExistingEmails 批量查找已存在的邮箱（解决N+1查询问题）
func (r *UserRepositoryImpl) FindExistingEmails(ctx context.Context, emails []string) ([]string, error) {
	if len(emails) == 0 {
		return []string{}, nil
	}

	var existingEmails []string
	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("email IN ?", emails).
		Pluck("email", &existingEmails).Error

	if err != nil {
		return nil, NewRepositoryError("find_existing_emails", "user", err)
	}

	return existingEmails, nil
}

// BatchCreate 批量创建用户（优化性能）
func (r *UserRepositoryImpl) BatchCreate(ctx context.Context, users []*models.User) error {
	if len(users) == 0 {
		return nil
	}

	// 使用GORM的Create方法进行批量插入
	if err := r.db.WithContext(ctx).Create(&users).Error; err != nil {
		return NewRepositoryError("batch_create", "user", err)
	}

	return nil
}

// sanitizeSearchInput 清理搜索输入，防止SQL注入和XSS攻击
func (r *UserRepositoryImpl) sanitizeSearchInput(input string) string {
	// 移除危险字符
	dangerousChars := []string{
		"'", "\"", ";", "--", "/*", "*/", "xp_", "sp_",
		"<", ">", "(", ")", "{", "}", "[", "]",
		"=", "!=", "<>", ">", "<", ">=", "<=",
	}

	cleaned := input
	for _, char := range dangerousChars {
		cleaned = strings.ReplaceAll(cleaned, char, "")
	}

	// 限制搜索长度，防止过长输入
	if len(cleaned) > 100 {
		cleaned = cleaned[:100]
	}

	return strings.TrimSpace(cleaned)
}

// GetLawyers 获取律师用户列表
func (r *UserRepositoryImpl) GetLawyers(ctx context.Context, page, pageSize int) ([]models.User, error) {
	var lawyers []models.User

	offset := (page - 1) * pageSize

	err := r.db.WithContext(ctx).
		Where("role = ?", "lawyer").
		Where("status = ?", "active").
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&lawyers).Error

	if err != nil {
		return nil, NewRepositoryError("get_lawyers", "user", err)
	}

	return lawyers, nil
}

// FindByStringID 根据字符串ID查找用户
func (r *UserRepositoryImpl) FindByStringID(id string) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByRole 根据角色查找用户
func (r *UserRepositoryImpl) FindByRole(role string, limit int) ([]models.User, error) {
	var users []models.User
	query := r.db.Where("role = ?", role).Where("status = ?", "active")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Order("created_at ASC").Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// FindDepartmentHead 查找部门主管
func (r *UserRepositoryImpl) FindDepartmentHead(deptID string, limit int) ([]models.User, error) {
	var users []models.User
	query := r.db.Where("department_id = ?", deptID).
		Where("role IN ?", []string{"department_head", "admin"}).
		Where("status = ?", "active")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Order("created_at ASC").Find(&users).Error
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		// 如果没有部门主管，返回管理员
		return r.FindByRole("admin", limit)
	}
	return users, nil
}
