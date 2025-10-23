package services

import (
	"context"
	"fmt"
	"time"

	"law-oa-go/internal/errors"
	"law-oa-go/internal/repositories"
)

// DocumentPermissionService handles document permission management
type DocumentPermissionService struct {
	docRepo  repositories.DocumentRepository
	userRepo repositories.UserRepository
}

// NewDocumentPermissionService creates a new document permission service
func NewDocumentPermissionService(docRepo repositories.DocumentRepository, userRepo repositories.UserRepository) *DocumentPermissionService {
	return &DocumentPermissionService{
		docRepo:  docRepo,
		userRepo: userRepo,
	}
}

// Permission represents a document permission
type Permission struct {
	ID           uint      `json:"id"`
	DocumentID   uint      `json:"document_id"`
	UserID       uint      `json:"user_id"`
	UserName     string    `json:"user_name"`
	Permission   string    `json:"permission"` // read, write, delete, share, admin
	GrantedBy    uint      `json:"granted_by"`
	GrantedByUser string   `json:"granted_by_user"`
	GrantedAt    time.Time `json:"granted_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	IsActive     bool      `json:"is_active"`
}

// PermissionRequest represents a permission grant request
type PermissionRequest struct {
	DocumentID uint      `json:"document_id" binding:"required"`
	UserID     uint      `json:"user_id" binding:"required"`
	Permission string    `json:"permission" binding:"required"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	GrantedBy  uint      `json:"granted_by" binding:"required"`
}

// PermissionCheckRequest represents a permission check request
type PermissionCheckRequest struct {
	DocumentID uint   `json:"document_id" binding:"required"`
	UserID     uint   `json:"user_id" binding:"required"`
	Permission string `json:"permission" binding:"required"`
}

// ShareDocumentRequest represents a document sharing request
type ShareDocumentRequest struct {
	DocumentID uint                `json:"document_id" binding:"required"`
	Users      []UserPermission    `json:"users" binding:"required"`
	Message    string              `json:"message"`
	ShareBy    uint                `json:"share_by" binding:"required"`
}

// UserPermission represents a user permission in a share request
type UserPermission struct {
	UserID     uint       `json:"user_id" binding:"required"`
	Permission string     `json:"permission" binding:"required"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

// DocumentPermissions represents all permissions for a document
type DocumentPermissions struct {
	DocumentID   uint                    `json:"document_id"`
	OwnerID      uint                    `json:"owner_id"`
	OwnerName    string                  `json:"owner_name"`
	Permissions  []Permission            `json:"permissions"`
	PublicAccess *PublicAccess            `json:"public_access,omitempty"`
	SharedWith   []SharedUser            `json:"shared_with,omitempty"`
	CreatedAt    time.Time               `json:"created_at"`
	UpdatedAt    time.Time               `json:"updated_at"`
}

// PublicAccess represents public access settings
type PublicAccess struct {
	Enabled    bool   `json:"enabled"`
	Permission string `json:"permission"`
	URL        string `json:"url,omitempty"`
	Password   string `json:"password,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

// SharedUser represents a user the document is shared with
type SharedUser struct {
	UserID     uint       `json:"user_id"`
	UserName   string     `json:"user_name"`
	UserEmail  string     `json:"user_email"`
	Permission string     `json:"permission"`
	SharedAt   time.Time  `json:"shared_at"`
	SharedBy   string     `json:"shared_by"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

// GrantPermission grants a permission to a user for a document
func (s *DocumentPermissionService) GrantPermission(ctx context.Context, req *PermissionRequest) (*Permission, error) {
	// Validate that the granter has permission to grant permissions
	if !s.hasPermission(ctx, req.DocumentID, req.GrantedBy, "admin") {
		return nil, errors.UnauthorizedError("grant_permission", "document")
	}

	// Validate that the target user exists
	targetUser, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, errors.DatabaseError("get_user", "Failed to get user", err)
	}
	if targetUser == nil {
		return nil, errors.NotFoundError("user", "User not found", req.UserID)
	}

	// Validate permission type
	validPermissions := map[string]bool{
		"read":   true,
		"write":  true,
		"delete": true,
		"share":  true,
		"admin":  true,
	}
	if !validPermissions[req.Permission] {
		return nil, errors.ValidationErrorWithDetails("permission", "Invalid permission", "Valid permissions are: read, write, delete, share, admin", []string{"invalid_permission"})
	}

	// Get granter user info
	granterUser, err := s.userRepo.FindByID(ctx, req.GrantedBy)
	if err != nil {
		return nil, errors.DatabaseError("get_granter", "Failed to get granter", err)
	}

	// Create permission
	permission := &Permission{
		DocumentID:   req.DocumentID,
		UserID:       req.UserID,
		UserName:     targetUser.Name,
		Permission:   req.Permission,
		GrantedBy:    req.GrantedBy,
		GrantedByUser: granterUser.Name,
		GrantedAt:    time.Now(),
		ExpiresAt:    req.ExpiresAt,
		IsActive:     true,
	}

	// In a real implementation, save to database
	// For now, return the constructed permission
	return permission, nil
}

// RevokePermission revokes a permission from a user
func (s *DocumentPermissionService) RevokePermission(ctx context.Context, documentID, userID, revokedBy uint) error {
	// Validate that the revoker has admin permission
	if !s.hasPermission(ctx, documentID, revokedBy, "admin") {
		return errors.UnauthorizedError("revoke_permission", "You don't have permission to manage this document")
	}

	// In a real implementation, delete from database
	// For now, just log the revocation
	fmt.Printf("Permission revoked for user %d on document %d by user %d\n", userID, documentID, revokedBy)

	return nil
}

// CheckPermission checks if a user has a specific permission for a document
func (s *DocumentPermissionService) CheckPermission(ctx context.Context, req *PermissionCheckRequest) (bool, error) {
	// First, check if user is the document owner
	if s.isDocumentOwner(ctx, req.DocumentID, req.UserID) {
		return true, nil
	}

	// Check explicit permissions
	if s.hasExplicitPermission(ctx, req.DocumentID, req.UserID, req.Permission) {
		return true, nil
	}

	// Check inherited permissions (e.g., from team membership)
	if s.hasInheritedPermission(ctx, req.DocumentID, req.UserID, req.Permission) {
		return true, nil
	}

	return false, nil
}

// GetDocumentPermissions retrieves all permissions for a document
func (s *DocumentPermissionService) GetDocumentPermissions(ctx context.Context, documentID uint, requestedBy uint) (*DocumentPermissions, error) {
	// Check if requester has permission to view permissions
	if !s.hasPermission(ctx, documentID, requestedBy, "admin") {
		return nil, errors.UnauthorizedError("view_permissions", "You don't have permission to view document permissions")
	}

	// Get document owner info
	ownerID, ownerName := s.getDocumentOwner(ctx, documentID)

	// In a real implementation, query from database
	// For now, return mock data
	permissions := &DocumentPermissions{
		DocumentID: documentID,
		OwnerID:   ownerID,
		OwnerName: ownerName,
		Permissions: []Permission{
			{
				ID:           1,
				DocumentID:   documentID,
				UserID:       2,
				UserName:     "User Two",
				Permission:   "read",
				GrantedBy:    ownerID,
				GrantedByUser: ownerName,
				GrantedAt:    time.Now().Add(-2 * time.Hour),
				ExpiresAt:    nil,
				IsActive:     true,
			},
			{
				ID:           2,
				DocumentID:   documentID,
				UserID:       3,
				UserName:     "User Three",
				Permission:   "write",
				GrantedBy:    ownerID,
				GrantedByUser: ownerName,
				GrantedAt:    time.Now().Add(-1 * time.Hour),
				ExpiresAt:    func() *time.Time { t := time.Now().Add(24 * time.Hour); return &t }(),
				IsActive:     true,
			},
		},
		SharedWith: []SharedUser{
			{
				UserID:     2,
				UserName:   "User Two",
				UserEmail:  "user2@example.com",
				Permission: "read",
				SharedAt:   time.Now().Add(-2 * time.Hour),
				SharedBy:   ownerName,
				ExpiresAt:  nil,
			},
			{
				UserID:     3,
				UserName:   "User Three",
				UserEmail:  "user3@example.com",
				Permission: "write",
				SharedAt:   time.Now().Add(-1 * time.Hour),
				SharedBy:   ownerName,
				ExpiresAt:  func() *time.Time { t := time.Now().Add(24 * time.Hour); return &t }(),
			},
		},
		CreatedAt: time.Now().Add(-3 * time.Hour),
		UpdatedAt: time.Now(),
	}

	return permissions, nil
}

// ShareDocument shares a document with multiple users
func (s *DocumentPermissionService) ShareDocument(ctx context.Context, req *ShareDocumentRequest) ([]*Permission, error) {
	// Check if user has permission to share
	if !s.hasPermission(ctx, req.DocumentID, req.ShareBy, "share") {
		return nil, errors.UnauthorizedError("share_document", "You don't have permission to share this document")
	}

	// Get sharer info
	_, err := s.userRepo.FindByID(ctx, req.ShareBy)
	if err != nil {
		return nil, errors.DatabaseError("get_sharer", "Failed to get sharer", err)
	}

	permissions := make([]*Permission, 0, len(req.Users))

	for _, userPerm := range req.Users {
		// Validate that the target user exists
		targetUser, err := s.userRepo.FindByID(ctx, userPerm.UserID)
		if err != nil {
			continue // Skip invalid users
		}
		if targetUser == nil {
			continue
		}

		// Grant permission
		permReq := &PermissionRequest{
			DocumentID: req.DocumentID,
			UserID:     userPerm.UserID,
			Permission: userPerm.Permission,
			ExpiresAt:  userPerm.ExpiresAt,
			GrantedBy:  req.ShareBy,
		}

		permission, err := s.GrantPermission(ctx, permReq)
		if err != nil {
			// Log error but continue with other users
			fmt.Printf("Failed to grant permission to user %d: %v\n", userPerm.UserID, err)
			continue
		}

		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// GetUserDocuments retrieves all documents a user has access to
func (s *DocumentPermissionService) GetUserDocuments(ctx context.Context, userID uint, permission string) ([]uint, error) {
	// In a real implementation, query from database
	// For now, return mock data
	documentIDs := []uint{1, 2, 3, 4, 5} // Mock document IDs

	return documentIDs, nil
}

// UpdatePermission updates an existing permission
func (s *DocumentPermissionService) UpdatePermission(ctx context.Context, documentID, userID uint, newPermission string, updatedBy uint) (*Permission, error) {
	// Check if updater has admin permission
	if !s.hasPermission(ctx, documentID, updatedBy, "admin") {
		return nil, errors.UnauthorizedError("update_permission", "You don't have permission to manage this document")
	}

	// Validate new permission
	validPermissions := map[string]bool{
		"read":   true,
		"write":  true,
		"delete": true,
		"share":  true,
		"admin":  true,
	}
	if !validPermissions[newPermission] {
		return nil, errors.ValidationErrorWithDetails("permission", "Invalid permission", "Valid permissions are: read, write, delete, share, admin", []string{"invalid_permission"})
	}

	// Get user info
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.DatabaseError("get_user", "Failed to get user", err)
	}
	if user == nil {
		return nil, errors.NotFoundError("user", "User not found", userID)
	}

	// Get updater info
	updaterUser, err := s.userRepo.FindByID(ctx, updatedBy)
	if err != nil {
		return nil, errors.DatabaseError("get_updater", "Failed to get updater", err)
	}

	// Create updated permission
	permission := &Permission{
		DocumentID:   documentID,
		UserID:       userID,
		UserName:     user.Name,
		Permission:   newPermission,
		GrantedBy:    updatedBy,
		GrantedByUser: updaterUser.Name,
		GrantedAt:    time.Now(),
		IsActive:     true,
	}

	// In a real implementation, update in database
	// For now, return the constructed permission
	return permission, nil
}

// Helper methods

// isDocumentOwner checks if a user is the owner of a document
func (s *DocumentPermissionService) isDocumentOwner(ctx context.Context, documentID, userID uint) bool {
	// Get document and check if it's owned by the user
	_, err := s.docRepo.FindByID(ctx, documentID)
	if err != nil {
		return false
	}

	// In a real implementation, check if document.owner_id == userID
	// For now, assume the first user is the owner
	return userID == 1 // Mock: assume user 1 is owner of all documents
}

// hasExplicitPermission checks if a user has an explicit permission
func (s *DocumentPermissionService) hasExplicitPermission(ctx context.Context, documentID uint, userID uint, permission string) bool {
	// In a real implementation, query permissions table
	// For now, return mock logic
	return userID == 2 && permission == "read" // Mock: user 2 has read permission
}

// hasInheritedPermission checks if a user has permission through inheritance
func (s *DocumentPermissionService) hasInheritedPermission(ctx context.Context, documentID uint, userID uint, permission string) bool {
	// In a real implementation, check team membership, role-based permissions, etc.
	// For now, return mock logic
	return userID == 3 && permission == "write" // Mock: user 3 has write permission through inheritance
}

// hasPermission checks if a user has any permission for a document
func (s *DocumentPermissionService) hasPermission(ctx context.Context, documentID uint, userID uint, permission string) bool {
	// Check owner first
	if s.isDocumentOwner(ctx, documentID, userID) {
		return true
	}

	// Check explicit permission
	if s.hasExplicitPermission(ctx, documentID, userID, permission) {
		return true
	}

	// Check inherited permission
	if s.hasInheritedPermission(ctx, documentID, userID, permission) {
		return true
	}

	// Check if user has higher-level permission
	// (e.g., admin permission includes all other permissions)
	permissionHierarchy := map[string][]string{
		"admin":  {"admin", "share", "delete", "write", "read"},
		"delete": {"delete", "write", "read"},
		"write":  {"write", "read"},
		"share":  {"share", "read"},
	}

	for _, higherPerm := range permissionHierarchy[permission] {
		if s.hasExplicitPermission(ctx, documentID, userID, higherPerm) ||
			s.hasInheritedPermission(ctx, documentID, userID, higherPerm) {
			return true
		}
	}

	return false
}

// getDocumentOwner gets the owner ID and name of a document
func (s *DocumentPermissionService) getDocumentOwner(ctx context.Context, documentID uint) (uint, string) {
	// In a real implementation, query document table
	// For now, return mock data
	return 1, "Document Owner"
}