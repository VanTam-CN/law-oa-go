package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/law-oa-go/document-service/internal/models"
	"github.com/law-oa-go/document-service/internal/repositories"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// userService 用户服务实现
type userService struct {
	userRepo       repositories.UserRepository
	roleRepo       repositories.RoleRepository
	permissionRepo repositories.DocumentPermissionRepository
	auditRepo      repositories.DocumentAuditRepository
	logger         *logrus.Logger
}

// NewUserService 创建新的用户服务
func NewUserService(
	userRepo repositories.UserRepository,
	roleRepo repositories.RoleRepository,
	permissionRepo repositories.DocumentPermissionRepository,
	auditRepo repositories.DocumentAuditRepository,
	logger *logrus.Logger,
) UserService {
	return &userService{
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		auditRepo:      auditRepo,
		logger:         logger,
	}
}

// CreateUser 创建用户
func (s *userService) CreateUser(ctx context.Context, req *CreateUserRequest) (*UserResponse, error) {
	// 验证请求
	if err := s.validateCreateUserRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 检查邮箱是否已存在
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// 检查用户名是否已存在
	existingUser, err = s.userRepo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with username %s already exists", req.Username)
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 创建用户
	user := &models.User{
		Username:    req.Username,
		Email:       req.Email,
		Password:    string(hashedPassword),
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		DisplayName: req.DisplayName,
		Avatar:      req.Avatar,
		Status:      "active", // 默认激活状态
		TenantID:    req.TenantID,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 分配默认角色
	if req.RoleID != "" {
		roleID, err := s.parseRoleID(req.RoleID)
		if err != nil {
			s.logger.WithError(err).Warn("Invalid role ID provided")
		} else {
			if err := s.assignRoleToUser(ctx, user.ID, roleID); err != nil {
				s.logger.WithError(err).Error("Failed to assign default role to user")
			}
		}
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		UserID:     strconv.FormatUint(uint64(user.ID), 10),
		Action:     "create_user",
		Details:    fmt.Sprintf("Created user: %s (%s)", user.Username, user.Email),
		TenantID:   req.TenantID,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	return s.convertToUserResponse(user), nil
}

// GetUser 获取用户信息
func (s *userService) GetUser(ctx context.Context, userID string) (*UserResponse, error) {
	id, err := s.parseUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return s.convertToUserResponse(user), nil
}

// GetUserByEmail 通过邮箱获取用户
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*UserResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return s.convertToUserResponse(user), nil
}

// GetUserByUsername 通过用户名获取用户
func (s *userService) GetUserByUsername(ctx context.Context, username string) (*UserResponse, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return s.convertToUserResponse(user), nil
}

// UpdateUser 更新用户信息
func (s *userService) UpdateUser(ctx context.Context, req *UpdateUserRequest) (*UserResponse, error) {
	// 解析用户ID
	userID, err := s.parseUserID(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// 获取现有用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 更新字段
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.DisplayName != "" {
		user.DisplayName = req.DisplayName
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Status != "" {
		user.Status = req.Status
	}

	user.UpdatedAt = time.Now()

	// 保存更新
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		UserID:     req.UserID,
		Action:     "update_user",
		Details:    fmt.Sprintf("Updated user: %s", user.Username),
		TenantID:   req.TenantID,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	return s.convertToUserResponse(user), nil
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(ctx context.Context, req *DeleteUserRequest) error {
	userID, err := s.parseUserID(req.UserID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// 获取用户信息用于审计
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 删除用户
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		UserID:     req.UserID,
		Action:     "delete_user",
		Details:    fmt.Sprintf("Deleted user: %s (%s)", user.Username, user.Email),
		TenantID:   req.TenantID,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	return nil
}

// ListUsers 列出用户
func (s *userService) ListUsers(ctx context.Context, filter *UserFilter) (*UserListResponse, error) {
	// 设置默认值
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	// 计算偏移量
	offset := (filter.Page - 1) * filter.PageSize

	// 获取用户列表
	users, total, err := s.userRepo.List(ctx, repositories.UserListOptions{
		TenantID:  filter.TenantID,
		Status:    filter.Status,
		RoleID:    filter.RoleID,
		Search:    filter.Search,
		Limit:     filter.PageSize,
		Offset:    offset,
		SortBy:    filter.SortBy,
		SortOrder: filter.SortOrder,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// 转换为响应格式
	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = s.convertToUserResponse(user)
	}

	return &UserListResponse{
		Users:    responses,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

// ChangePassword 修改密码
func (s *userService) ChangePassword(ctx context.Context, req *ChangePasswordRequest) error {
	userID, err := s.parseUserID(req.UserID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 验证当前密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// 更新密码
	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		UserID:     req.UserID,
		Action:     "change_password",
		Details:    fmt.Sprintf("Changed password for user: %s", user.Username),
		TenantID:   req.TenantID,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	return nil
}

// ResetPassword 重置密码
func (s *userService) ResetPassword(ctx context.Context, req *ResetPasswordRequest) (string, error) {
	userID, err := s.parseUserID(req.UserID)
	if err != nil {
		return "", fmt.Errorf("invalid user ID: %w", err)
	}

	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// 生成临时密码
	tempPassword := s.generateTemporaryPassword()

	// 加密临时密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash temporary password: %w", err)
	}

	// 更新密码
	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return "", fmt.Errorf("failed to reset password: %w", err)
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		UserID:     req.UserID,
		Action:     "reset_password",
		Details:    fmt.Sprintf("Reset password for user: %s", user.Username),
		TenantID:   req.TenantID,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	return tempPassword, nil
}

// AssignRole 分配角色
func (s *userService) AssignRole(ctx context.Context, req *AssignRoleRequest) error {
	userID, err := s.parseUserID(req.UserID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	roleID, err := s.parseRoleID(req.RoleID)
	if err != nil {
		return fmt.Errorf("invalid role ID: %w", err)
	}

	// 验证用户存在
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 验证角色存在
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	// 分配角色
	if err := s.assignRoleToUser(ctx, userID, roleID); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		UserID:     req.UserID,
		Action:     "assign_role",
		Details:    fmt.Sprintf("Assigned role %s to user %s", role.Name, user.Username),
		TenantID:   req.TenantID,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	return nil
}

// RemoveRole 移除角色
func (s *userService) RemoveRole(ctx context.Context, req *RemoveRoleRequest) error {
	userID, err := s.parseUserID(req.UserID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	roleID, err := s.parseRoleID(req.RoleID)
	if err != nil {
		return fmt.Errorf("invalid role ID: %w", err)
	}

	// 验证用户存在
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 验证角色存在
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	// 移除角色
	if err := s.removeRoleFromUser(ctx, userID, roleID); err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		UserID:     req.UserID,
		Action:     "remove_role",
		Details:    fmt.Sprintf("Removed role %s from user %s", role.Name, user.Username),
		TenantID:   req.TenantID,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	return nil
}

// GetUserRoles 获取用户角色
func (s *userService) GetUserRoles(ctx context.Context, userID string) ([]*RoleResponse, error) {
	id, err := s.parseUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// 获取用户角色关联
	userRoles, err := s.userRepo.GetUserRoles(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// 获取角色详情
	roles := make([]*RoleResponse, len(userRoles))
	for i, userRole := range userRoles {
		role, err := s.roleRepo.GetByID(ctx, userRole.RoleID)
		if err != nil {
			s.logger.WithError(err).WithField("role_id", userRole.RoleID).Error("Failed to get role details")
			continue
		}
		roles[i] = &RoleResponse{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			TenantID:    role.TenantID,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
		}
	}

	return roles, nil
}

// ValidateUser 验证用户
func (s *userService) ValidateUser(ctx context.Context, username, password string) (*UserResponse, error) {
	// 通过用户名或邮箱查找用户
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		user, err = s.userRepo.GetByEmail(ctx, username)
		if err != nil {
			return nil, fmt.Errorf("user not found")
		}
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, fmt.Errorf("user account is not active")
	}

	// 更新最后登录时间
	user.LastLoginAt = &time.Time{}
	*user.LastLoginAt = time.Now()
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.WithError(err).Warn("Failed to update last login time")
	}

	return s.convertToUserResponse(user), nil
}

// GetActiveUsers 获取活跃用户
func (s *userService) GetActiveUsers(ctx context.Context, tenantID string) ([]*UserResponse, error) {
	users, err := s.userRepo.GetActiveUsers(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}

	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = s.convertToUserResponse(user)
	}

	return responses, nil
}

// GetUsersByRole 根据角色获取用户
func (s *userService) GetUsersByRole(ctx context.Context, roleID string) ([]*UserResponse, error) {
	id, err := s.parseRoleID(roleID)
	if err != nil {
		return nil, fmt.Errorf("invalid role ID: %w", err)
	}

	users, err := s.userRepo.GetUsersByRole(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by role: %w", err)
	}

	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = s.convertToUserResponse(user)
	}

	return responses, nil
}

// 辅助方法

// validateCreateUserRequest 验证创建用户请求
func (s *userService) validateCreateUserRequest(req *CreateUserRequest) error {
	if req.Username == "" {
		return fmt.Errorf("username is required")
	}
	if len(req.Username) < 3 || len(req.Username) > 50 {
		return fmt.Errorf("username must be between 3 and 50 characters")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if !s.isValidEmail(req.Email) {
		return fmt.Errorf("invalid email format")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	if len(req.Password) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}
	if req.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	return nil
}

// isValidEmail 验证邮箱格式
func (s *userService) isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// parseUserID 解析用户ID
func (s *userService) parseUserID(userID string) (uint, error) {
	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID format: %s", userID)
	}
	return uint(id), nil
}

// parseRoleID 解析角色ID
func (s *userService) parseRoleID(roleID string) (uint, error) {
	id, err := strconv.ParseUint(roleID, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid role ID format: %s", roleID)
	}
	return uint(id), nil
}

// assignRoleToUser 分配角色给用户
func (s *userService) assignRoleToUser(ctx context.Context, userID, roleID uint) error {
	userRole := &models.UserRole{
		UserID:    userID,
		RoleID:    roleID,
		CreatedAt: time.Now(),
	}

	return s.userRepo.AssignRole(ctx, userRole)
}

// removeRoleFromUser 从用户移除角色
func (s *userService) removeRoleFromUser(ctx context.Context, userID, roleID uint) error {
	return s.userRepo.RemoveRole(ctx, userID, roleID)
}

// generateTemporaryPassword 生成临时密码
func (s *userService) generateTemporaryPassword() string {
	// 简单实现，生成8位随机密码
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	password := make([]byte, 8)
	for i := range password {
		password[i] = chars[i%len(chars)]
	}
	return string(password)
}

// convertToUserResponse 转换为用户响应格式
func (s *userService) convertToUserResponse(user *models.User) *UserResponse {
	response := &UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		DisplayName: user.DisplayName,
		Avatar:      user.Avatar,
		Status:      user.Status,
		TenantID:    user.TenantID,
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	// 拼接全名
	if user.FirstName != "" && user.LastName != "" {
		response.FullName = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	} else if user.FirstName != "" {
		response.FullName = user.FirstName
	} else if user.LastName != "" {
		response.FullName = user.LastName
	} else {
		response.FullName = user.Username
	}

	return response
}

// logAudit 记录审计日志
func (s *userService) logAudit(ctx context.Context, req *LogActionRequest) error {
	// 解析用户ID
	var userID uint
	if req.UserID != "" {
		id, err := s.parseUserID(req.UserID)
		if err != nil {
			return err
		}
		userID = id
	}

	audit := &models.DocumentAudit{
		UserID:    userID,
		TenantID:  req.TenantID,
		Action:    req.Action,
		Details:   req.Details,
		IPAddress: req.IPAddress,
		UserAgent: req.UserAgent,
	}

	return s.auditRepo.Create(ctx, audit)
}