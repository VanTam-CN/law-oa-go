package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"law-oa-go/internal/common"
	"law-oa-go/internal/models"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required,min=1,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=admin lawyer user"`
	Phone    string `json:"phone" binding:"omitempty,max=20"`
}

type UpdateUserRequest struct {
	Name  *string `json:"name" binding:"omitempty,min=1,max=100"`
	Email *string `json:"email" binding:"omitempty,email"`
	Phone *string `json:"phone" binding:"omitempty,max=20"`
}

type UserProfile struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Phone     string    `json:"phone"`
	Avatar    string    `json:"avatar"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*UserProfile, error) {
	if err := s.validateUserRequest(req); err != nil {
		return nil, err
	}

	var existingUser models.User
	if err := s.db.WithContext(ctx).Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, fmt.Errorf("%w: %s", common.ErrEmailExists, req.Email)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, common.NewInternalError("failed to hash password", err)
	}

	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     req.Role,
		Phone:    req.Phone,
		Status:   "active",
	}

	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, common.NewDatabaseError("create user", err)
	}

	return s.toUserProfile(user), nil
}

func (s *UserService) AuthenticateUser(ctx context.Context, email, password string) (*UserProfile, error) {
	var user models.User
	err := s.db.WithContext(ctx).Where("email = ? AND status = ?", email, "active").First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", common.ErrUserNotFound, email)
		}
		return nil, common.NewDatabaseError("authenticate user", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("%w", common.ErrInvalidPassword)
	}

	return s.toUserProfile(&user), nil
}

func (s *UserService) GetUserProfile(ctx context.Context, userID uint) (*UserProfile, error) {
	var user models.User
	err := s.db.WithContext(ctx).First(&user, userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NewNotFoundError("user")
		}
		return nil, common.NewDatabaseError("get user profile", err)
	}

	return s.toUserProfile(&user), nil
}

func (s *UserService) UpdateUser(ctx context.Context, userID uint, req *UpdateUserRequest) (*UserProfile, error) {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NewNotFoundError("user")
		}
		return nil, common.NewDatabaseError("find user", err)
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Email != nil {
		if *req.Email != user.Email {
			var existingUser models.User
			if err := s.db.WithContext(ctx).Where("email = ? AND id != ?", *req.Email, userID).First(&existingUser).Error; err == nil {
				return nil, common.NewValidationError("email already exists", "The email address is already in use")
			}
		}
		updates["email"] = *req.Email
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}

	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(&user).Updates(updates).Error; err != nil {
			return nil, common.NewDatabaseError("update user", err)
		}
	}

	return s.GetUserProfile(ctx, userID)
}

func (s *UserService) ChangePassword(ctx context.Context, userID uint, currentPassword, newPassword string) error {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NewNotFoundError("user")
		}
		return common.NewDatabaseError("find user", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return common.NewValidationError("invalid current password", "Current password is incorrect")
	}

	if err := s.validatePassword(newPassword); err != nil {
		return fmt.Errorf("%w: %v", common.ErrWeakPassword, err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return common.NewInternalError("failed to hash password", err)
	}

	if err := s.db.WithContext(ctx).Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		return common.NewDatabaseError("update password", err)
	}

	return nil
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
		return fmt.Errorf("%w: %s (allowed: admin, lawyer, user)", common.ErrInvalidRole, req.Role)
	}

	return nil
}

func (s *UserService) validateEmail(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return common.NewValidationError("invalid email format", "Please provide a valid email address")
	}
	return nil
}

func (s *UserService) validatePassword(password string) error {
	if len(password) < 8 {
		return common.NewValidationError("password too weak", "Password must be at least 8 characters long")
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)  // 修复：使用正确的小写字母正则
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)

	if !hasUpper || !hasLower || !hasNumber {
		return common.NewValidationError("password too weak", "Password must contain at least one uppercase letter, one lowercase letter, and one number")
	}

	return nil
}

func (s *UserService) toUserProfile(user *models.User) *UserProfile {
	return &UserProfile{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		Phone:     user.Phone,
		Avatar:    user.Avatar,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
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

	query := s.db.WithContext(ctx).Model(&models.User{})

	if req.Role != "" {
		query = query.Where("role = ?", req.Role)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.Search != "" {
		searchTerm := "%" + strings.ToLower(req.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ?", searchTerm, searchTerm)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, common.NewDatabaseError("count users", err)
	}

	var users []models.User
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, common.NewDatabaseError("list users", err)
	}

	profiles := make([]*UserProfile, len(users))
	for i, user := range users {
		profiles[i] = s.toUserProfile(&user)
	}

	return profiles, total, nil
}

func (s *UserService) DeleteUser(ctx context.Context, userID uint) error {
	result := s.db.WithContext(ctx).Delete(&models.User{}, userID)
	if result.Error != nil {
		return common.NewDatabaseError("delete user", result.Error)
	}
	if result.RowsAffected == 0 {
		return common.NewNotFoundError("user")
	}
	return nil
}