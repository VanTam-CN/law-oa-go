package services

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"

	stderrors "errors"
	"golang.org/x/crypto/bcrypt"
	"law-oa-go/internal/concurrency"
	"law-oa-go/internal/errors"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// 预编译的正则表达式（避免重复编译，提升性能）
var (
	upperRegex  = regexp.MustCompile(`[A-Z]`)
	lowerRegex  = regexp.MustCompile(`[a-z]`)
	numberRegex = regexp.MustCompile(`[0-9]`)
	emailRegex  = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

type UserService struct {
	userRepo       repositories.UserRepository
	concurrentSvc  *concurrency.ConcurrentService
	concurrentSafe *concurrency.ConcurrentSafe
	mu             sync.RWMutex
}

func NewUserService(userRepo repositories.UserRepository) *UserService {
	config := &concurrency.ConcurrentConfig{
		MaxWorkers:     15,
		QueueSize:      500,
		TaskTimeout:    20 * time.Second,
		EnableMetrics:  true,
		CircuitBreaker: true,
		RateLimiter:    true,
		RetryPolicy: concurrency.RetryPolicy{
			MaxRetries:    3,
			RetryDelay:    100 * time.Millisecond,
			BackoffFactor: 2.0,
		},
	}

	concurrentSvc := concurrency.NewConcurrentService(config)
	concurrentSafe := concurrency.NewConcurrentSafe(true, true, 50, 25)

	return &UserService{
		userRepo:       userRepo,
		concurrentSvc:  concurrentSvc,
		concurrentSafe: concurrentSafe,
	}
}

// StartConcurrentService 启动并发服务
func (s *UserService) StartConcurrentService() {
	s.concurrentSvc.Start()
}

// StopConcurrentService 停止并发服务
func (s *UserService) StopConcurrentService() {
	s.concurrentSvc.Stop()
}

// GetConcurrentMetrics 获取并发指标
func (s *UserService) GetConcurrentMetrics() *concurrency.PoolMetricsSnapshot {
	return s.concurrentSvc.GetMetrics()
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=1,max=50"`
	Name     string `json:"name" binding:"required,min=1,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=admin lawyer user"`
	Phone    string `json:"phone" binding:"omitempty,max=20"`
}

type UpdateUserRequest struct {
	Name       *string `json:"name" binding:"omitempty,min=1,max=100"`
	Email      *string `json:"email" binding:"omitempty,email"`
	Phone      *string `json:"phone" binding:"omitempty,max=20"`
	Department *string `json:"department" binding:"omitempty,max=50"`
	Seniority  *string `json:"seniority" binding:"omitempty,max=20"`
}

type UserProfile struct {
	ID         uint      `json:"id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Role       string    `json:"role"`
	Phone      string    `json:"phone"`
	Avatar     string    `json:"avatar"`
	Status     string    `json:"status"`
	Department string    `json:"department"`
	Seniority  string    `json:"seniority"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*UserProfile, error) {
	if err := s.validateUserRequest(req); err != nil {
		return nil, err
	}

	// 检查邮箱是否已存在
	existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if stderrors.Is(err, repositories.ErrUserNotFound) {
			// 用户不存在，可以创建
		} else {
			return nil, errors.DatabaseError("check_email_existence", "Failed to check email existence", err)
		}
	}
	if existingUser != nil {
		return nil, errors.BusinessError("email", "exists", "Email address already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.InternalError("Failed to hash password", err)
	}

	user := &models.User{
		Username: req.Username,
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     req.Role,
		Phone:    req.Phone,
		Status:   "active",
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, errors.DatabaseError("create_user", "Failed to create user", err)
	}

	return s.toUserProfile(user), nil
}

func (s *UserService) AuthenticateUser(ctx context.Context, email, password string) (*UserProfile, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if stderrors.Is(err, repositories.ErrUserNotFound) {
			return nil, errors.NotFoundError("user", "User not found: "+email, nil)
		}
		return nil, errors.DatabaseError("authenticate_user", "Failed to authenticate user", err)
	}
	if user == nil || user.Status != "active" {
		return nil, errors.NotFoundError("user", "User not found: "+email, nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.ValidationErrorWithDetails("password", "Invalid password", "The provided password is incorrect", []string{"invalid_password"})
	}

	return s.toUserProfile(user), nil
}

func (s *UserService) GetUserProfile(ctx context.Context, userID uint) (*UserProfile, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.DatabaseError("get_user_profile", "Failed to get user profile", err)
	}
	if user == nil {
		return nil, errors.NotFoundError("user", "User not found", nil)
	}

	return s.toUserProfile(user), nil
}

func (s *UserService) UpdateUser(ctx context.Context, userID uint, req *UpdateUserRequest) (*UserProfile, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.DatabaseError("find_user", "Failed to find user", err)
	}
	if user == nil {
		return nil, errors.NotFoundError("user", "User not found", nil)
	}

	// 检查邮箱是否已被其他用户使用
	if req.Email != nil && *req.Email != user.Email {
		existingUser, err := s.userRepo.FindByEmail(ctx, *req.Email)
		if err != nil {
			if stderrors.Is(err, repositories.ErrUserNotFound) {
				// 邮箱不存在，可以使用
			} else {
				return nil, errors.DatabaseError("check_email_existence", "Failed to check email existence", err)
			}
		}
		if existingUser != nil && existingUser.ID != userID {
			return nil, errors.BusinessError("email", "exists", "Email already exists")
		}
		user.Email = *req.Email
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}
	if req.Department != nil {
		user.Department = *req.Department
	}
	if req.Seniority != nil {
		user.Seniority = *req.Seniority
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, errors.DatabaseError("update_user", "Failed to update user", err)
	}

	return s.GetUserProfile(ctx, userID)
}

func (s *UserService) ChangePassword(ctx context.Context, userID uint, currentPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.DatabaseError("find_user", "Failed to find user", err)
	}
	if user == nil {
		return errors.NotFoundError("user", "User not found", nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return errors.ValidationErrorWithDetails("password", "Current password is incorrect", "The current password provided is incorrect", []string{"invalid_current_password"})
	}

	if err := s.validatePassword(newPassword); err != nil {
		return errors.ValidationErrorWithDetails("password", "weak_password", "Password too weak", []string{"Password does not meet security requirements: " + err.Error()})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.InternalError("Failed to hash password", err)
	}

	user.Password = string(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return errors.DatabaseError("update_password", "Failed to update password", err)
	}

	return nil
}

// UpdateUserAvatar 更新用户头像
func (s *UserService) UpdateUserAvatar(ctx context.Context, userID uint, avatarPath string) (*UserProfile, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.DatabaseError("find_user", "Failed to find user", err)
	}
	if user == nil {
		return nil, errors.NotFoundError("user", "User not found", nil)
	}

	user.Avatar = avatarPath
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, errors.DatabaseError("update_avatar", "Failed to update avatar", err)
	}

	return s.toUserProfile(user), nil
}

func (s *UserService) validateUserRequest(req *CreateUserRequest) error {
	if err := s.validateEmail(req.Email); err != nil {
		return err
	}

	if err := s.validatePassword(req.Password); err != nil {
		return err
	}

	validRoles := map[string]bool{
		"admin":  true,
		"lawyer": true,
		"user":   true,
	}

	if !validRoles[req.Role] {
		return errors.ValidationErrorWithDetails("role", "invalid_role", "Invalid role", []string{"Role must be one of: admin, lawyer, user"})
	}

	return nil
}

func (s *UserService) validateEmail(email string) error {
	// 使用预编译的正则表达式，避免重复编译提升性能
	if !emailRegex.MatchString(email) {
		return errors.ValidationErrorWithDetails("email", "invalid_email_format", "Invalid email format", []string{"Please provide a valid email address"})
	}
	return nil
}

func (s *UserService) validatePassword(password string) error {
	if len(password) < 8 {
		return errors.ValidationErrorWithDetails("password", "password_too_short", "Password too short", []string{"Password must be at least 8 characters long"})
	}

	// 使用预编译的正则表达式，避免重复编译提升性能
	hasUpper := upperRegex.MatchString(password)
	hasLower := lowerRegex.MatchString(password)
	hasNumber := numberRegex.MatchString(password)

	if !hasUpper || !hasLower || !hasNumber {
		return errors.ValidationErrorWithDetails("password", "password_too_weak", "Password too weak", []string{"Password must contain at least one uppercase letter, one lowercase letter, and one number"})
	}

	return nil
}

func (s *UserService) toUserProfile(user *models.User) *UserProfile {
	return &UserProfile{
		ID:         user.ID,
		Name:       user.Name,
		Email:      user.Email,
		Role:       user.Role,
		Phone:      user.Phone,
		Avatar:     user.Avatar,
		Status:     user.Status,
		Department: user.Department,
		Seniority:  user.Seniority,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}
}

type UserListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Role     string `form:"role" binding:"omitempty,oneof=admin lawyer user"`
	Status   string `form:"status" binding:"omitempty,oneof=active inactive"`
	Search   string `form:"search"`
}

func (s *UserService) ListUsers(ctx context.Context, req *UserListRequest) ([]*UserProfile, int64, error) {
	page := 1
	pageSize := 20
	if req.Page > 0 {
		page = req.Page
	}
	if req.PageSize > 0 {
		pageSize = req.PageSize
	}

	params := &repositories.UserListParams{
		Page:     page,
		PageSize: pageSize,
		Role:     req.Role,
		Status:   req.Status,
		Search:   req.Search,
	}

	users, total, err := s.userRepo.List(ctx, params)
	if err != nil {
		return nil, 0, errors.DatabaseError("list_users", "Failed to list users", err)
	}

	profiles := make([]*UserProfile, len(users))
	for i, user := range users {
		profiles[i] = s.toUserProfile(user)
	}

	return profiles, total, nil
}

func (s *UserService) DeleteUser(ctx context.Context, userID uint) error {
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		if stderrors.Is(err, repositories.ErrUserNotFound) {
			return errors.NotFoundError("user", "User not found", nil)
		}
		return errors.DatabaseError("delete_user", "Failed to delete user", err)
	}
	return nil
}

// BatchCreateUsers 批量创建用户（优化版本 - 解决N+1查询问题）
func (s *UserService) BatchCreateUsers(ctx context.Context, requests []*CreateUserRequest) ([]*UserProfile, error) {
	if len(requests) == 0 {
		return nil, errors.ValidationErrorWithDetails("requests", "empty_request", "No users to create", []string{"Batch create request is empty"})
	}

	// 1. 批量验证所有请求
	if err := s.batchValidateRequests(requests); err != nil {
		return nil, err
	}

	// 2. 批量检查邮箱重复（解决N+1查询问题）
	emails := s.extractEmails(requests)
	existingEmails, err := s.userRepo.FindExistingEmails(ctx, emails)
	if err != nil {
		return nil, errors.DatabaseError("check_existing_emails", "Failed to check existing emails", err)
	}

	// 3. 检查是否有重复邮箱
	if len(existingEmails) > 0 {
		return nil, errors.BusinessError("email", "exist", fmt.Sprintf("Email addresses already exist: %v", existingEmails))
	}

	// 4. 批量创建用户
	users, err := s.batchPrepareUsers(requests)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.BatchCreate(ctx, users); err != nil {
		return nil, errors.DatabaseError("batch_create_users", "Failed to batch create users", err)
	}

	// 5. 转换为用户配置文件格式
	profiles := s.batchConvertToProfiles(users)
	return profiles, nil
}

// createSingleUser 创建单个用户（内部方法）
func (s *UserService) createSingleUser(ctx context.Context, req *CreateUserRequest, result **UserProfile, mu *sync.Mutex) error {
	if err := s.validateUserRequest(req); err != nil {
		return err
	}

	// 检查邮箱是否已存在
	existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if stderrors.Is(err, repositories.ErrUserNotFound) {
			// 用户不存在，可以创建
		} else {
			return errors.DatabaseError("check_email_existence", "Failed to check email existence", err)
		}
	}
	if existingUser != nil {
		return errors.BusinessError("email", "exists", "Email address already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.InternalError("Failed to hash password", err)
	}

	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     req.Role,
		Phone:    req.Phone,
		Status:   "active",
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return errors.DatabaseError("create_user", "Failed to create user", err)
	}

	userProfile, err := s.GetUserProfile(ctx, user.ID)
	if err != nil {
		return err
	}

	mu.Lock()
	*result = userProfile
	mu.Unlock()

	return nil
}

// BatchUpdateUsers 批量更新用户（并发版本）
func (s *UserService) BatchUpdateUsers(ctx context.Context, updates map[uint]*UpdateUserRequest) ([]*UserProfile, error) {
	if len(updates) == 0 {
		return nil, errors.ValidationErrorWithDetails("updates", "empty_request", "No users to update", []string{"Batch update request is empty"})
	}

	results := make([]*UserProfile, 0, len(updates))

	// 使用并发安全的批量更新
	if err := s.executeBatchUpdates(ctx, updates, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// executeBatchUpdates 执行并发批量更新
func (s *UserService) executeBatchUpdates(ctx context.Context, updates map[uint]*UpdateUserRequest, results *[]*UserProfile) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(updates))

	for userID, req := range updates {
		wg.Add(1)
		userID, req := userID, req // 闭包捕获

		go func(id uint, request *UpdateUserRequest) {
			defer wg.Done()
			if err := s.updateSingleUserSafe(ctx, id, request, results, &mu); err != nil {
				errChan <- err
			}
		}(userID, req)
	}

	wg.Wait()
	close(errChan)

	// 检查错误
	for err := range errChan {
		if err != nil {
			return errors.InternalError("Batch update users failed", err)
		}
	}

	return nil
}

// updateSingleUserSafe 并发安全的单用户更新
func (s *UserService) updateSingleUserSafe(ctx context.Context, userID uint, req *UpdateUserRequest, results *[]*UserProfile, mu *sync.Mutex) error {
	return s.concurrentSafe.Execute(ctx, func() error {
		return s.updateSingleUser(ctx, userID, req, results, mu)
	})
}

// updateSingleUser 更新单个用户（内部方法）
func (s *UserService) updateSingleUser(ctx context.Context, userID uint, req *UpdateUserRequest, results *[]*UserProfile, mu *sync.Mutex) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.DatabaseError("find_user", "Failed to find user", err)
	}
	if user == nil {
		return errors.NotFoundError("user", "User not found", nil)
	}

	// 检查邮箱是否已被其他用户使用
	if req.Email != nil && *req.Email != user.Email {
		existingUser, err := s.userRepo.FindByEmail(ctx, *req.Email)
		if err != nil {
			if stderrors.Is(err, repositories.ErrUserNotFound) {
				// 邮箱不存在，可以使用
			} else {
				return errors.DatabaseError("check_email_existence", "Failed to check email existence", err)
			}
		}
		if existingUser != nil && existingUser.ID != userID {
			return errors.BusinessError("email", "exists", "Email already exists")
		}
		user.Email = *req.Email
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return errors.DatabaseError("update_user", "Failed to update user", err)
	}

	userProfile, err := s.GetUserProfile(ctx, userID)
	if err != nil {
		return err
	}

	mu.Lock()
	*results = append(*results, userProfile)
	mu.Unlock()

	return nil
}

// BatchDeleteUsers 批量删除用户（并发版本）
func (s *UserService) BatchDeleteUsers(ctx context.Context, userIDs []uint) error {
	if len(userIDs) == 0 {
		return errors.ValidationErrorWithDetails("user_ids", "empty_request", "No users to delete", []string{"Batch delete request is empty"})
	}

	tasks := make([]concurrency.Task, len(userIDs))
	var wg sync.WaitGroup
	errChan := make(chan error, len(userIDs))

	for i, userID := range userIDs {
		wg.Add(1)
		userID := userID // 闭包捕获
		task := &concurrency.DatabaseTask{
			TaskID:       fmt.Sprintf("delete_user_%d", userID),
			TaskType:     "delete_user",
			TaskPriority: 1,
			Operation: func(ctx context.Context) error {
				return s.concurrentSafe.Execute(ctx, func() error {
					return s.DeleteUser(ctx, userID)
				})
			},
			Context: ctx,
		}
		tasks[i] = task

		go func(task concurrency.Task) {
			defer wg.Done()
			if err := task.Execute(ctx); err != nil {
				errChan <- err
			}
		}(task)
	}

	wg.Wait()
	close(errChan)

	// 检查错误
	for err := range errChan {
		if err != nil {
			return errors.InternalError("Batch delete users failed", err)
		}
	}

	return nil
}

// ConcurrentBatchGetUsers 并发批量获取用户信息
func (s *UserService) ConcurrentBatchGetUsers(ctx context.Context, userIDs []uint) ([]*UserProfile, error) {
	if len(userIDs) == 0 {
		return nil, errors.ValidationErrorWithDetails("user_ids", "empty_request", "No users to get", []string{"Batch get request is empty"})
	}

	tasks := make([]concurrency.Task, len(userIDs))
	results := make([]*UserProfile, len(userIDs))
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(userIDs))

	for i, userID := range userIDs {
		wg.Add(1)
		userID := userID // 闭包捕获
		i := i           // 闭包捕获
		task := &concurrency.DatabaseTask{
			TaskID:       fmt.Sprintf("get_user_%d", userID),
			TaskType:     "get_user",
			TaskPriority: 4,
			Operation: func(ctx context.Context) error {
				return s.concurrentSafe.Execute(ctx, func() error {
					profile, err := s.GetUserProfile(ctx, userID)
					if err != nil {
						return err
					}
					mu.Lock()
					results[i] = profile
					mu.Unlock()
					return nil
				})
			},
			Context: ctx,
		}
		tasks[i] = task

		go func(task concurrency.Task) {
			defer wg.Done()
			if err := task.Execute(ctx); err != nil {
				errChan <- err
			}
		}(task)
	}

	wg.Wait()
	close(errChan)

	// 检查错误
	for err := range errChan {
		if err != nil {
			return nil, errors.InternalError("Batch get users failed", err)
		}
	}

	// 过滤掉nil结果
	var validResults []*UserProfile
	for _, result := range results {
		if result != nil {
			validResults = append(validResults, result)
		}
	}

	return validResults, nil
}

// BatchChangePassword 批量修改密码（并发版本）
func (s *UserService) BatchChangePassword(ctx context.Context, changes map[uint]map[string]string) error {
	if len(changes) == 0 {
		return errors.ValidationErrorWithDetails("changes", "empty_request", "No password changes to process", []string{"Batch password change request is empty"})
	}

	tasks := make([]concurrency.Task, 0, len(changes))
	var wg sync.WaitGroup
	errChan := make(chan error, len(changes))

	for userID, changeData := range changes {
		wg.Add(1)
		userID, changeData := userID, changeData // 闭包捕获
		task := &concurrency.DatabaseTask{
			TaskID:       fmt.Sprintf("change_password_%d", userID),
			TaskType:     "change_password",
			TaskPriority: 3,
			Operation: func(ctx context.Context) error {
				return s.concurrentSafe.Execute(ctx, func() error {
					currentPassword, ok1 := changeData["current_password"]
					newPassword, ok2 := changeData["new_password"]
					if !ok1 || !ok2 {
						return errors.ValidationErrorWithDetails("password_change_data", "Invalid password change data", fmt.Sprintf("Invalid password change data for user %d", userID), []string{"invalid_password_change_data"})
					}
					return s.ChangePassword(ctx, userID, currentPassword, newPassword)
				})
			},
			Context: ctx,
		}
		tasks = append(tasks, task)

		go func(task concurrency.Task) {
			defer wg.Done()
			if err := task.Execute(ctx); err != nil {
				errChan <- err
			}
		}(task)
	}

	wg.Wait()
	close(errChan)

	// 检查错误
	for err := range errChan {
		if err != nil {
			return errors.InternalError("Batch change password failed", err)
		}
	}

	return nil
}

// batchValidateRequests 批量验证用户请求
func (s *UserService) batchValidateRequests(requests []*CreateUserRequest) error {
	for i, req := range requests {
		if err := s.validateUserRequest(req); err != nil {
			return errors.ValidationErrorWithDetails(
				fmt.Sprintf("request_%d", i),
				fmt.Sprintf("Invalid request at index %d: %v", i, err),
				"One or more requests are invalid",
				[]string{"invalid_request"},
			)
		}
	}
	return nil
}

// extractEmails 提取所有邮箱地址
func (s *UserService) extractEmails(requests []*CreateUserRequest) []string {
	emails := make([]string, len(requests))
	for i, req := range requests {
		emails[i] = req.Email
	}
	return emails
}

// batchPrepareUsers 批量准备用户数据
func (s *UserService) batchPrepareUsers(requests []*CreateUserRequest) ([]*models.User, error) {
	users := make([]*models.User, len(requests))

	for i, req := range requests {
		// 密码哈希
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.InternalError("Failed to hash password", err)
		}

		users[i] = &models.User{
			Name:     req.Name,
			Email:    req.Email,
			Password: string(hashedPassword),
			Role:     req.Role,
			Phone:    req.Phone,
			Status:   "active",
		}
	}

	return users, nil
}

// batchConvertToProfiles 批量转换为用户配置文件
func (s *UserService) batchConvertToProfiles(users []*models.User) []*UserProfile {
	profiles := make([]*UserProfile, len(users))
	for i, user := range users {
		profiles[i] = s.toUserProfile(user)
	}
	return profiles
}
