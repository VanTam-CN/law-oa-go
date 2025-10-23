package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/law-oa-go/document-service/internal/models"
	"github.com/law-oa-go/document-service/internal/repositories"
	"github.com/sirupsen/logrus"
)

// permissionService 权限服务实现
type permissionService struct {
	permissionRepo repositories.DocumentPermissionRepository
	docRepo       repositories.DocumentRepository
	userRepo      repositories.UserRepository
	roleRepo      repositories.RoleRepository
	auditRepo     repositories.DocumentAuditRepository
	logger        *logrus.Logger
}

// NewPermissionService 创建新的权限服务
func NewPermissionService(
	permissionRepo repositories.DocumentPermissionRepository,
	docRepo repositories.DocumentRepository,
	userRepo repositories.UserRepository,
	roleRepo repositories.RoleRepository,
	auditRepo repositories.DocumentAuditRepository,
	logger *logrus.Logger,
) PermissionService {
	return &permissionService{
		permissionRepo: permissionRepo,
		docRepo:       docRepo,
		userRepo:      userRepo,
		roleRepo:      roleRepo,
		auditRepo:     auditRepo,
		logger:        logger,
	}
}

// GrantPermission 授予权限
func (s *permissionService) GrantPermission(ctx context.Context, req *GrantPermissionRequest) error {
	// 验证请求
	if err := s.validateGrantPermissionRequest(req); err != nil {
		return err
	}

	// 解析ID
	documentID, err := s.parseDocumentID(req.DocumentID)
	if err != nil {
		return fmt.Errorf("invalid document ID: %w", err)
	}

	// 验证文档存在
	document, err := s.docRepo.GetByID(ctx, documentID)
	if err != nil {
		return fmt.Errorf("document not found: %w", err)
	}

	// 检查权限是否已存在
	var existingPermission *models.DocumentPermission
	if req.UserID != "" {
		userID, err := s.parseUserID(req.UserID)
		if err != nil {
			return fmt.Errorf("invalid user ID: %w", err)
		}

		// 检查用户权限是否已存在
		exists, err := s.permissionRepo.CheckUserPermission(ctx, documentID, userID, req.Permission)
		if err != nil {
			return fmt.Errorf("failed to check existing permission: %w", err)
		}
		if exists {
			return fmt.Errorf("permission already exists for user")
		}

		existingPermission = &models.DocumentPermission{
			DocumentID: documentID,
			UserID:     &userID,
			TenantID:   req.TenantID,
			Permission: req.Permission,
		}
	} else if req.RoleID != "" {
		roleID, err := s.parseRoleID(req.RoleID)
		if err != nil {
			return fmt.Errorf("invalid role ID: %w", err)
		}

		// 检查角色权限是否已存在
		exists, err := s.permissionRepo.CheckRolePermission(ctx, documentID, roleID, req.Permission)
		if err != nil {
			return fmt.Errorf("failed to check existing permission: %w", err)
		}
		if exists {
			return fmt.Errorf("permission already exists for role")
		}

		existingPermission = &models.DocumentPermission{
			DocumentID: documentID,
			RoleID:     &roleID,
			TenantID:   req.TenantID,
			Permission: req.Permission,
		}
	} else {
		return fmt.Errorf("either user_id or role_id must be provided")
	}

	// 创建权限
	if err := s.permissionRepo.Create(ctx, existingPermission); err != nil {
		return fmt.Errorf("failed to grant permission: %w", err)
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		DocumentID: req.DocumentID,
		Action:     "grant_permission",
		Details:    fmt.Sprintf("Granted %s permission: %s", req.Permission, s.getPermissionTarget(req)),
		TenantID:   req.TenantID,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	return nil
}

// RevokePermission 撤销权限
func (s *permissionService) RevokePermission(ctx context.Context, req *RevokePermissionRequest) error {
	// 解析ID
	documentID, err := s.parseDocumentID(req.DocumentID)
	if err != nil {
		return fmt.Errorf("invalid document ID: %w", err)
	}

	// 构建删除条件
	query := s.permissionRepo.DeletePermission(ctx, documentID, 0, 0, "")

	if req.UserID != "" {
		userID, err := s.parseUserID(req.UserID)
		if err != nil {
			return fmt.Errorf("invalid user ID: %w", err)
		}
		query = s.permissionRepo.DeletePermission(ctx, documentID, userID, 0, "")
	} else if req.RoleID != "" {
		roleID, err := s.parseRoleID(req.RoleID)
		if err != nil {
			return fmt.Errorf("invalid role ID: %w", err)
		}
		query = s.permissionRepo.DeletePermission(ctx, documentID, 0, roleID, "")
	}

	if req.Permission != "" {
		query = s.permissionRepo.DeletePermission(ctx, documentID, 0, 0, req.Permission)
	}

	if err := query; err != nil {
		return fmt.Errorf("failed to revoke permission: %w", err)
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		DocumentID: req.DocumentID,
		Action:     "revoke_permission",
		Details:    fmt.Sprintf("Revoked permission: %s", s.getPermissionTarget(req)),
		TenantID:   req.TenantID,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	return nil
}

// UpdatePermission 更新权限
func (s *permissionService) UpdatePermission(ctx context.Context, req *UpdatePermissionRequest) error {
	// 解析ID
	documentID, err := s.parseDocumentID(req.DocumentID)
	if err != nil {
		return fmt.Errorf("invalid document ID: %w", err)
	}

	// 更新权限
	if req.UserID != "" {
		userID, err := s.parseUserID(req.UserID)
		if err != nil {
			return fmt.Errorf("invalid user ID: %w", err)
		}
		err = s.permissionRepo.UpdatePermission(ctx, documentID, userID, 0, req.NewPermission)
		if err != nil {
			return fmt.Errorf("failed to update user permission: %w", err)
		}
	} else if req.RoleID != "" {
		roleID, err := s.parseRoleID(req.RoleID)
		if err != nil {
			return fmt.Errorf("invalid role ID: %w", err)
		}
		err = s.permissionRepo.UpdatePermission(ctx, documentID, 0, roleID, req.NewPermission)
		if err != nil {
			return fmt.Errorf("failed to update role permission: %w", err)
		}
	} else {
		return fmt.Errorf("either user_id or role_id must be provided")
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		DocumentID: req.DocumentID,
		Action:     "update_permission",
		Details:    fmt.Sprintf("Updated permission from %s to %s: %s", req.OldPermission, req.NewPermission, s.getPermissionTarget(req)),
		TenantID:   req.TenantID,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	return nil
}

// GetDocumentPermissions 获取文档的所有权限
func (s *permissionService) GetDocumentPermissions(ctx context.Context, documentID string) (*PermissionListResponse, error) {
	docID, err := s.parseDocumentID(documentID)
	if err != nil {
		return nil, fmt.Errorf("invalid document ID: %w", err)
	}

	permissions, err := s.permissionRepo.FindByDocument(ctx, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document permissions: %w", err)
	}

	responses := make([]*PermissionResponse, len(permissions))
	for i, perm := range permissions {
		responses[i] = &PermissionResponse{
			ID:         perm.ID,
			DocumentID: perm.DocumentID,
			UserID:     perm.UserID,
			RoleID:     perm.RoleID,
			TenantID:   perm.TenantID,
			Permission: perm.Permission,
			CreatedAt:  perm.CreatedAt,
			UpdatedAt:  perm.UpdatedAt,
		}
	}

	return &PermissionListResponse{
		Permissions: responses,
		Total:       int64(len(responses)),
	}, nil
}

// GetUserPermissions 获取用户对文档的所有权限
func (s *permissionService) GetUserPermissions(ctx context.Context, userID, documentID string) ([]string, error) {
	userIDInt, err := s.parseUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	docID, err := s.parseDocumentID(documentID)
	if err != nil {
		return nil, fmt.Errorf("invalid document ID: %w", err)
	}

	// 获取用户直接权限
	userPermissions, err := s.permissionRepo.GetUserPermissions(ctx, docID, userIDInt)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	// 获取用户的角色权限
	userRoles, err := s.getUserRoles(ctx, userIDInt)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// 合并所有权限
	allPermissions := make(map[string]bool)
	for _, perm := range userPermissions {
		allPermissions[perm] = true
	}

	for _, roleID := range userRoles {
		rolePermissions, err := s.permissionRepo.GetRolePermissions(ctx, docID, roleID)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to get role permissions")
			continue
		}
		for _, perm := range rolePermissions {
			allPermissions[perm] = true
		}
	}

	// 转换为切片
	result := make([]string, 0, len(allPermissions))
	for perm := range allPermissions {
		result = append(result, perm)
	}

	return result, nil
}

// GetUserAccessibleDocuments 获取用户可访问的文档
func (s *permissionService) GetUserAccessibleDocuments(ctx context.Context, userID string, permission string) ([]*DocumentResponse, error) {
	userIDInt, err := s.parseUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// 获取用户有直接权限的文档
	documents, err := s.permissionRepo.GetUserAccessibleDocuments(ctx, userIDInt, permission)
	if err != nil {
		return nil, fmt.Errorf("failed to get user accessible documents: %w", err)
	}

	// 获取用户角色权限的文档
	userRoles, err := s.getUserRoles(ctx, userIDInt)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	for _, roleID := range userRoles {
		roleDocs, err := s.permissionRepo.GetRoleAccessibleDocuments(ctx, roleID, permission)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to get role accessible documents")
			continue
		}

		// 合并文档（去重）
		docMap := make(map[uint]*models.Document)
		for _, doc := range documents {
			docMap[doc.ID] = doc
		}
		for _, doc := range roleDocs {
			docMap[doc.ID] = doc
		}

		// 转换回切片
		documents = make([]*models.Document, 0, len(docMap))
		for _, doc := range docMap {
			documents = append(documents, doc)
		}
	}

	// 转换为响应
	responses := make([]*DocumentResponse, len(documents))
	for i, doc := range documents {
		responses[i] = &DocumentResponse{
			ID:             doc.ID,
			UUID:           doc.UUID,
			Name:           doc.Name,
			Description:    doc.Description,
			OriginalName:   doc.OriginalName,
			MIMEType:       doc.MIMEType,
			Size:           doc.Size,
			Category:       doc.Category,
			CurrentVersion: doc.CurrentVersion,
			CreatedBy:      doc.CreatedBy,
			CreatedAt:      doc.CreatedAt,
			UpdatedAt:      doc.UpdatedAt,
		}
	}

	return responses, nil
}

// CheckPermission 检查用户是否有指定权限
func (s *permissionService) CheckPermission(ctx context.Context, userID, documentID string, permission string) (bool, error) {
	userIDInt, err := s.parseUserID(userID)
	if err != nil {
		return false, fmt.Errorf("invalid user ID: %w", err)
	}

	docID, err := s.parseDocumentID(documentID)
	if err != nil {
		return false, fmt.Errorf("invalid document ID: %w", err)
	}

	// 检查用户直接权限
	hasPermission, err := s.permissionRepo.CheckUserPermission(ctx, docID, userIDInt, permission)
	if err != nil {
		return false, fmt.Errorf("failed to check user permission: %w", err)
	}

	if hasPermission {
		return true, nil
	}

	// 检查角色权限
	userRoles, err := s.getUserRoles(ctx, userIDInt)
	if err != nil {
		return false, fmt.Errorf("failed to get user roles: %w", err)
	}

	for _, roleID := range userRoles {
		hasRolePermission, err := s.permissionRepo.CheckRolePermission(ctx, docID, roleID, permission)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to check role permission")
			continue
		}
		if hasRolePermission {
			return true, nil
		}
	}

	return false, nil
}

// CheckPermissionBatch 批量检查权限
func (s *permissionService) CheckPermissionBatch(ctx context.Context, userID string, documentIDs []string, permission string) (map[string]bool, error) {
	result := make(map[string]bool)

	for _, docID := range documentIDs {
		hasPermission, err := s.CheckPermission(ctx, userID, docID, permission)
		if err != nil {
			s.logger.WithError(err).WithField("document_id", docID).Error("Failed to check permission")
			result[docID] = false
		} else {
			result[docID] = hasPermission
		}
	}

	return result, nil
}

// GrantPermissionsBatch 批量授予权限
func (s *permissionService) GrantPermissionsBatch(ctx context.Context, reqs []*GrantPermissionRequest) error {
	if len(reqs) == 0 {
		return nil
	}

	// 使用事务批量处理
	for _, req := range reqs {
		if err := s.GrantPermission(ctx, req); err != nil {
			return fmt.Errorf("failed to grant permission for document %s: %w", req.DocumentID, err)
		}
	}

	return nil
}

// RevokePermissionsBatch 批量撤销权限
func (s *permissionService) RevokePermissionsBatch(ctx context.Context, reqs []*RevokePermissionRequest) error {
	if len(reqs) == 0 {
		return nil
	}

	// 使用事务批量处理
	for _, req := range reqs {
		if err := s.RevokePermission(ctx, req); err != nil {
			return fmt.Errorf("failed to revoke permission for document %s: %w", req.DocumentID, err)
		}
	}

	return nil
}

// 辅助方法

// validateGrantPermissionRequest 验证授予权限请求
func (s *permissionService) validateGrantPermissionRequest(req *GrantPermissionRequest) error {
	if req.DocumentID == "" {
		return fmt.Errorf("document_id is required")
	}
	if req.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if req.Permission == "" {
		return fmt.Errorf("permission is required")
	}
	if !s.isValidPermission(req.Permission) {
		return fmt.Errorf("invalid permission: %s", req.Permission)
	}
	if req.UserID == "" && req.RoleID == "" {
		return fmt.Errorf("either user_id or role_id must be provided")
	}
	if req.UserID != "" && req.RoleID != "" {
		return fmt.Errorf("cannot specify both user_id and role_id")
	}

	return nil
}

// isValidPermission 检查权限是否有效
func (s *permissionService) isValidPermission(permission string) bool {
	validPermissions := []string{"read", "write", "delete", "admin"}
	for _, valid := range validPermissions {
		if permission == valid {
			return true
		}
	}
	return false
}

// parseDocumentID 解析文档ID
func (s *permissionService) parseDocumentID(documentID string) (uint, error) {
	// 尝试作为UUID查找
	// TODO: 实现UUID查找逻辑

	// 尝试解析为整数
	id, err := strconv.ParseUint(documentID, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid document ID format: %s", documentID)
	}
	return uint(id), nil
}

// parseUserID 解析用户ID
func (s *permissionService) parseUserID(userID string) (uint, error) {
	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID format: %s", userID)
	}
	return uint(id), nil
}

// parseRoleID 解析角色ID
func (s *permissionService) parseRoleID(roleID string) (uint, error) {
	id, err := strconv.ParseUint(roleID, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid role ID format: %s", roleID)
	}
	return uint(id), nil
}

// getUserRoles 获取用户的所有角色
func (s *permissionService) getUserRoles(ctx context.Context, userID uint) ([]uint, error) {
	// TODO: 实现获取用户角色的逻辑
	// 这里需要调用用户角色关联仓库
	return []uint{}, nil
}

// getPermissionTarget 获取权限目标描述
func (s *permissionService) getPermissionTarget(req interface{}) string {
	switch v := req.(type) {
	case *GrantPermissionRequest:
		if v.UserID != "" {
			return fmt.Sprintf("user %s", v.UserID)
		} else if v.RoleID != "" {
			return fmt.Sprintf("role %s", v.RoleID)
		}
	case *RevokePermissionRequest:
		if v.UserID != "" {
			return fmt.Sprintf("user %s", v.UserID)
		} else if v.RoleID != "" {
			return fmt.Sprintf("role %s", v.RoleID)
		}
	case *UpdatePermissionRequest:
		if v.UserID != "" {
			return fmt.Sprintf("user %s", v.UserID)
		} else if v.RoleID != "" {
			return fmt.Sprintf("role %s", v.RoleID)
		}
	}
	return "unknown"
}

// logAudit 记录审计日志
func (s *permissionService) logAudit(ctx context.Context, req *LogActionRequest) error {
	// 解析用户ID
	var userID uint
	if req.UserID != "" {
		id, err := strconv.ParseUint(req.UserID, 10, 32)
		if err != nil {
			return err
		}
		userID = uint(id)
	}

	// 解析文档ID
	var documentID uint
	if req.DocumentID != "" {
		id, err := s.parseDocumentID(req.DocumentID)
		if err != nil {
			return err
		}
		documentID = id
	}

	audit := &models.DocumentAudit{
		DocumentID: documentID,
		UserID:     userID,
		TenantID:   req.TenantID,
		Action:     req.Action,
		Details:    req.Details,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	return s.auditRepo.Create(ctx, audit)
}