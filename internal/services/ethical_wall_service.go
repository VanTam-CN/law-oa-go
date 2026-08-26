package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

var (
	// ErrCaseNotFound 案件不存在
	ErrCaseNotFound = errors.New("case not found")
	// ErrUserNotFound 用户不存在
	ErrUserNotFound = errors.New("user not found")
	// ErrEthicalWallAlreadyEnabled 隔离墙已启用
	ErrEthicalWallAlreadyEnabled = errors.New("ethical wall already enabled")
	// ErrEthicalWallNotEnabled 隔离墙未启用
	ErrEthicalWallNotEnabled = errors.New("ethical wall not enabled")
	// ErrCannotAddToWhitelist 无法添加到白名单
	ErrCannotAddToWhitelist = errors.New("cannot add user to whitelist")
)

// EthicalWallService 隔离墙服务
type EthicalWallService struct {
	ethicalWallRepo repositories.EthicalWallRepository
	caseRepo        repositories.CaseRepository
	userRepo        repositories.UserRepository
}

// NewEthicalWallService 创建隔离墙服务实例
func NewEthicalWallService(
	ethicalWallRepo repositories.EthicalWallRepository,
	caseRepo repositories.CaseRepository,
	userRepo repositories.UserRepository,
) *EthicalWallService {
	return &EthicalWallService{
		ethicalWallRepo: ethicalWallRepo,
		caseRepo:        caseRepo,
		userRepo:        userRepo,
	}
}

// EnableEthicalWallRequest 启用隔离墙请求
type EnableEthicalWallRequest struct {
	CaseID      uint   `json:"case_id" binding:"required"`
	Description string `json:"description" binding:"max=500"`
}

// WhitelistEntryRequest 白名单条目请求
type WhitelistEntryRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Reason string `json:"reason" binding:"max=500"`
}

// WhitelistEntryResponse 白名单条目响应
type WhitelistEntryResponse struct {
	ID            uint      `json:"id"`
	CaseID        uint      `json:"case_id"`
	UserID        uint      `json:"user_id"`
	UserName      string    `json:"user_name"`
	GrantedBy     uint      `json:"granted_by"`
	GrantedByName string    `json:"granted_by_name"`
	GrantedAt     time.Time `json:"granted_at"`
	Reason        string    `json:"reason"`
}

// AccessLogResponse 访问日志响应
type AccessLogResponse struct {
	ID           uint      `json:"id"`
	CaseID       uint      `json:"case_id"`
	UserID       uint      `json:"user_id"`
	UserName     string    `json:"user_name"`
	AccessType   string    `json:"access_type"`
	AccessResult string    `json:"access_result"`
	IPAddress    string    `json:"ip_address"`
	AttemptedAt  time.Time `json:"attempted_at"`
}

// EnableEthicalWall 启用案件隔离墙
func (s *EthicalWallService) EnableEthicalWall(ctx context.Context, caseID uint, userID uint, description string) error {
	// 检查案件是否存在
	case_, err := s.caseRepo.FindByID(ctx, caseID)
	if err != nil {
		return fmt.Errorf("failed to find case: %w", err)
	}
	if case_ == nil {
		return ErrCaseNotFound
	}

	// 检查是否已启用
	enabled, err := s.ethicalWallRepo.IsEthicalWallEnabled(ctx, caseID)
	if err != nil {
		return fmt.Errorf("failed to check ethical wall status: %w", err)
	}
	if enabled {
		return ErrEthicalWallAlreadyEnabled
	}

	// 启用隔离墙
	return s.ethicalWallRepo.EnableEthicalWall(ctx, caseID, userID, description)
}

// DisableEthicalWall 禁用案件隔离墙
func (s *EthicalWallService) DisableEthicalWall(ctx context.Context, caseID uint) error {
	// 检查案件是否存在
	case_, err := s.caseRepo.FindByID(ctx, caseID)
	if err != nil {
		return fmt.Errorf("failed to find case: %w", err)
	}
	if case_ == nil {
		return ErrCaseNotFound
	}

	// 检查是否已启用
	enabled, err := s.ethicalWallRepo.IsEthicalWallEnabled(ctx, caseID)
	if err != nil {
		return fmt.Errorf("failed to check ethical wall status: %w", err)
	}
	if !enabled {
		return ErrEthicalWallNotEnabled
	}

	// 禁用隔离墙
	return s.ethicalWallRepo.DisableEthicalWall(ctx, caseID)
}

// AddToWhitelist 添加用户到案件白名单
func (s *EthicalWallService) AddToWhitelist(ctx context.Context, caseID, userID, grantedBy uint, reason string) error {
	// 检查案件是否存在
	case_, err := s.caseRepo.FindByID(ctx, caseID)
	if err != nil {
		return fmt.Errorf("failed to find case: %w", err)
	}
	if case_ == nil {
		return ErrCaseNotFound
	}

	// 检查用户是否存在
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// 检查隔离墙是否启用
	enabled, err := s.ethicalWallRepo.IsEthicalWallEnabled(ctx, caseID)
	if err != nil {
		return fmt.Errorf("failed to check ethical wall status: %w", err)
	}
	if !enabled {
		return ErrEthicalWallNotEnabled
	}

	// 添加到白名单
	return s.ethicalWallRepo.AddToWhitelist(ctx, caseID, userID, grantedBy, reason)
}

// RemoveFromWhitelist 从案件白名单移除用户
func (s *EthicalWallService) RemoveFromWhitelist(ctx context.Context, caseID, userID uint) error {
	// 检查案件是否存在
	case_, err := s.caseRepo.FindByID(ctx, caseID)
	if err != nil {
		return fmt.Errorf("failed to find case: %w", err)
	}
	if case_ == nil {
		return ErrCaseNotFound
	}

	// 从白名单移除
	return s.ethicalWallRepo.RemoveFromWhitelist(ctx, caseID, userID)
}

// GetWhitelist 获取案件白名单
func (s *EthicalWallService) GetWhitelist(ctx context.Context, caseID uint) ([]*WhitelistEntryResponse, error) {
	// 检查案件是否存在
	case_, err := s.caseRepo.FindByID(ctx, caseID)
	if err != nil {
		return nil, fmt.Errorf("failed to find case: %w", err)
	}
	if case_ == nil {
		return nil, ErrCaseNotFound
	}

	// 获取白名单
	entries, err := s.ethicalWallRepo.GetWhitelistByCase(ctx, caseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get whitelist: %w", err)
	}

	// 转换为响应格式
	responses := make([]*WhitelistEntryResponse, len(entries))
	for i, entry := range entries {
		responses[i] = s.convertWhitelistEntry(entry)
	}

	return responses, nil
}

// GetUserAccessibleCases 获取用户可访问的隔离墙案件
func (s *EthicalWallService) GetUserAccessibleCases(ctx context.Context, userID uint) ([]*WhitelistEntryResponse, error) {
	// 获取用户白名单
	entries, err := s.ethicalWallRepo.GetWhitelistByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user whitelist: %w", err)
	}

	// 转换为响应格式
	responses := make([]*WhitelistEntryResponse, len(entries))
	for i, entry := range entries {
		responses[i] = s.convertWhitelistEntry(entry)
	}

	return responses, nil
}

// CheckAccess 检查用户是否有权限访问案件
func (s *EthicalWallService) CheckAccess(ctx context.Context, caseID, userID uint) (bool, error) {
	// 检查隔离墙是否启用
	enabled, err := s.ethicalWallRepo.IsEthicalWallEnabled(ctx, caseID)
	if err != nil {
		return false, fmt.Errorf("failed to check ethical wall status: %w", err)
	}

	// 如果未启用隔离墙，允许访问
	if !enabled {
		return true, nil
	}

	// 检查用户是否在白名单中
	whitelisted, err := s.ethicalWallRepo.IsUserWhitelisted(ctx, caseID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check whitelist: %w", err)
	}

	return whitelisted, nil
}

// LogAccessAttempt 记录访问尝试
func (s *EthicalWallService) LogAccessAttempt(ctx context.Context, caseID, userID uint, accessType, accessResult, ipAddress, userAgent string) error {
	return s.ethicalWallRepo.LogAccessAttempt(ctx, caseID, userID, accessType, accessResult, ipAddress, userAgent)
}

// GetAccessLogs 获取访问日志
func (s *EthicalWallService) GetAccessLogs(ctx context.Context, caseID uint, limit int) ([]*AccessLogResponse, error) {
	logs, err := s.ethicalWallRepo.GetAccessLogs(ctx, caseID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get access logs: %w", err)
	}

	responses := make([]*AccessLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = &AccessLogResponse{
			ID:           log.ID,
			CaseID:       log.CaseID,
			UserID:       log.UserID,
			UserName:     "", // 需要预加载用户数据
			AccessType:   log.AccessType,
			AccessResult: log.AccessResult,
			IPAddress:    log.IPAddress,
			AttemptedAt:  log.AttemptedAt,
		}
	}

	return responses, nil
}

// ClearWhitelist 清空案件白名单
func (s *EthicalWallService) ClearWhitelist(ctx context.Context, caseID uint) error {
	return s.ethicalWallRepo.ClearWhitelist(ctx, caseID)
}

// convertWhitelistEntry 转换白名单条目为响应格式
func (s *EthicalWallService) convertWhitelistEntry(entry *models.CaseEthicalWallWhitelist) *WhitelistEntryResponse {
	resp := &WhitelistEntryResponse{
		ID:            entry.ID,
		CaseID:        entry.CaseID,
		UserID:        entry.UserID,
		UserName:      "",
		GrantedBy:     entry.GrantedBy,
		GrantedByName: "",
		GrantedAt:     entry.GrantedAt,
		Reason:        entry.Reason,
	}

	// 如果有预加载的用户数据，使用它
	if entry.User != nil {
		resp.UserName = entry.User.Name
		if resp.UserName == "" {
			resp.UserName = entry.User.Username
		}
	}

	if entry.GrantedByUser != nil {
		resp.GrantedByName = entry.GrantedByUser.Name
		if resp.GrantedByName == "" {
			resp.GrantedByName = entry.GrantedByUser.Username
		}
	}

	return resp
}
