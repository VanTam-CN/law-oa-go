package repositories

import (
	"context"
	"fmt"
	"time"

	"law-oa-go/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EditingRepository 编辑仓库接口
type EditingRepository interface {
	// 编辑会话管理
	CreateEditSession(ctx context.Context, session *models.EditSession) error
	GetEditSession(ctx context.Context, id uuid.UUID) (*models.EditSession, error)
	GetEditSessionByToken(ctx context.Context, token string) (*models.EditSession, error)
	UpdateEditSession(ctx context.Context, session *models.EditSession) error
	DeleteEditSession(ctx context.Context, id uuid.UUID) error
	GetActiveEditSessionsByDocument(ctx context.Context, documentID uuid.UUID) ([]*models.EditSession, error)
	CleanExpiredSessions(ctx context.Context) error

	// 编辑操作管理
	CreateEditOperation(ctx context.Context, operation *models.EditOperation) error
	GetEditOperations(ctx context.Context, documentID uuid.UUID, limit, offset int) ([]*models.EditOperation, error)
	GetPendingOperations(ctx context.Context, documentID uuid.UUID) ([]*models.EditOperation, error)
	MarkOperationsApplied(ctx context.Context, operationIDs []uuid.UUID) error
	DeleteOldOperations(ctx context.Context, beforeTime time.Time) error

	// 协作会话管理
	CreateCollaborationSession(ctx context.Context, session *models.CollaborationSession) error
	GetCollaborationSession(ctx context.Context, documentID uuid.UUID) (*models.CollaborationSession, error)
	UpdateCollaborationSession(ctx context.Context, session *models.CollaborationSession) error
	DeleteCollaborationSession(ctx context.Context, documentID uuid.UUID) error
	GetActiveCollaborationSessions(ctx context.Context) ([]*models.CollaborationSession, error)

	// 协作参与者管理
	AddCollaborationParticipant(ctx context.Context, participant *models.CollaborationParticipant) error
	RemoveCollaborationParticipant(ctx context.Context, sessionID, userID uuid.UUID) error
	UpdateParticipantStatus(ctx context.Context, sessionID, userID uuid.UUID, status string) error
	GetCollaborationParticipants(ctx context.Context, sessionID uuid.UUID) ([]*models.CollaborationParticipant, error)
	GetParticipantBySocketID(ctx context.Context, socketID string) (*models.CollaborationParticipant, error)

	// 文档版本管理
	CreateDocumentVersion(ctx context.Context, version *models.DocumentVersion) error
	GetDocumentVersions(ctx context.Context, documentID uuid.UUID, page, limit int) ([]*models.DocumentVersion, int64, error)
	GetDocumentVersion(ctx context.Context, versionID uuid.UUID) (*models.DocumentVersion, error)
	GetLatestVersion(ctx context.Context, documentID uuid.UUID) (*models.DocumentVersion, error)
	CompareVersions(ctx context.Context, fromID, toID uuid.UUID) (*models.DocumentVersion, *models.DocumentVersion, error)
	DeleteOldVersions(ctx context.Context, documentID uuid.UUID, keepCount int) error

	// 编辑权限管理
	CreateEditPermission(ctx context.Context, permission *models.EditPermission) error
	GetEditPermission(ctx context.Context, documentID, userID uuid.UUID) (*models.EditPermission, error)
	UpdateEditPermission(ctx context.Context, permission *models.EditPermission) error
	DeleteEditPermission(ctx context.Context, documentID, userID uuid.UUID) error
	GetDocumentPermissions(ctx context.Context, documentID uuid.UUID) ([]*models.EditPermission, error)
	GetUserDocuments(ctx context.Context, userID uuid.UUID) ([]*models.EditPermission, error)

	// 编辑器配置管理
	GetEditorConfig(ctx context.Context, editorType string) (*models.EditorConfig, error)
	CreateEditorConfig(ctx context.Context, config *models.EditorConfig) error
	UpdateEditorConfig(ctx context.Context, config *models.EditorConfig) error
	DeleteEditorConfig(ctx context.Context, editorType string) error
	GetAllEditorConfigs(ctx context.Context) ([]*models.EditorConfig, error)
}

// editingRepository 编辑仓库实现
type editingRepository struct {
	db *gorm.DB
}

// NewEditingRepository 创建编辑仓库实例
func NewEditingRepository(db *gorm.DB) EditingRepository {
	return &editingRepository{db: db}
}

// CreateEditSession 创建编辑会话
func (r *editingRepository) CreateEditSession(ctx context.Context, session *models.EditSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// GetEditSession 获取编辑会话
func (r *editingRepository) GetEditSession(ctx context.Context, id uuid.UUID) (*models.EditSession, error) {
	var session models.EditSession
	err := r.db.WithContext(ctx).
		Preload("Document").
		Preload("User").
		First(&session, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetEditSessionByToken 根据令牌获取编辑会话
func (r *editingRepository) GetEditSessionByToken(ctx context.Context, token string) (*models.EditSession, error) {
	var session models.EditSession
	err := r.db.WithContext(ctx).
		Preload("Document").
		Preload("User").
		First(&session, "session_token = ?", token).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// UpdateEditSession 更新编辑会话
func (r *editingRepository) UpdateEditSession(ctx context.Context, session *models.EditSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

// DeleteEditSession 删除编辑会话
func (r *editingRepository) DeleteEditSession(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.EditSession{}, "id = ?", id).Error
}

// GetActiveEditSessionsByDocument 获取文档的活跃编辑会话
func (r *editingRepository) GetActiveEditSessionsByDocument(ctx context.Context, documentID uuid.UUID) ([]*models.EditSession, error) {
	var sessions []*models.EditSession
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("document_id = ? AND expires_at > ?", documentID, time.Now()).
		Find(&sessions).Error
	return sessions, err
}

// CleanExpiredSessions 清理过期会话
func (r *editingRepository) CleanExpiredSessions(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Delete(&models.EditSession{}, "expires_at < ?", time.Now()).Error
}

// CreateEditOperation 创建编辑操作
func (r *editingRepository) CreateEditOperation(ctx context.Context, operation *models.EditOperation) error {
	return r.db.WithContext(ctx).Create(operation).Error
}

// GetEditOperations 获取编辑操作列表
func (r *editingRepository) GetEditOperations(ctx context.Context, documentID uuid.UUID, limit, offset int) ([]*models.EditOperation, error) {
	var operations []*models.EditOperation
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("document_id = ?", documentID).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&operations).Error
	return operations, err
}

// GetPendingOperations 获取待处理的操作
func (r *editingRepository) GetPendingOperations(ctx context.Context, documentID uuid.UUID) ([]*models.EditOperation, error) {
	var operations []*models.EditOperation
	err := r.db.WithContext(ctx).
		Where("document_id = ? AND applied = false", documentID).
		Order("timestamp ASC").
		Find(&operations).Error
	return operations, err
}

// MarkOperationsApplied 标记操作已应用
func (r *editingRepository) MarkOperationsApplied(ctx context.Context, operationIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.EditOperation{}).
		Where("id IN ?", operationIDs).
		Update("applied", true).Error
}

// DeleteOldOperations 删除旧操作记录
func (r *editingRepository) DeleteOldOperations(ctx context.Context, beforeTime time.Time) error {
	return r.db.WithContext(ctx).
		Delete(&models.EditOperation{}, "timestamp < ?", beforeTime).Error
}

// CreateCollaborationSession 创建协作会话
func (r *editingRepository) CreateCollaborationSession(ctx context.Context, session *models.CollaborationSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// GetCollaborationSession 获取协作会话
func (r *editingRepository) GetCollaborationSession(ctx context.Context, documentID uuid.UUID) (*models.CollaborationSession, error) {
	var session models.CollaborationSession
	err := r.db.WithContext(ctx).
		Preload("Document").
		First(&session, "document_id = ?", documentID).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// UpdateCollaborationSession 更新协作会话
func (r *editingRepository) UpdateCollaborationSession(ctx context.Context, session *models.CollaborationSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

// DeleteCollaborationSession 删除协作会话
func (r *editingRepository) DeleteCollaborationSession(ctx context.Context, documentID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Delete(&models.CollaborationSession{}, "document_id = ?", documentID).Error
}

// GetActiveCollaborationSessions 获取活跃的协作会话
func (r *editingRepository) GetActiveCollaborationSessions(ctx context.Context) ([]*models.CollaborationSession, error) {
	var sessions []*models.CollaborationSession
	err := r.db.WithContext(ctx).
		Where("active_users > 0 AND last_activity > ?", time.Now().Add(-5*time.Minute)).
		Find(&sessions).Error
	return sessions, err
}

// AddCollaborationParticipant 添加协作参与者
func (r *editingRepository) AddCollaborationParticipant(ctx context.Context, participant *models.CollaborationParticipant) error {
	return r.db.WithContext(ctx).Create(participant).Error
}

// RemoveCollaborationParticipant 移除协作参与者
func (r *editingRepository) RemoveCollaborationParticipant(ctx context.Context, sessionID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Delete(&models.CollaborationParticipant{}, "session_id = ? AND user_id = ?", sessionID, userID).Error
}

// UpdateParticipantStatus 更新参与者状态
func (r *editingRepository) UpdateParticipantStatus(ctx context.Context, sessionID, userID uuid.UUID, status string) error {
	return r.db.WithContext(ctx).
		Model(&models.CollaborationParticipant{}).
		Where("session_id = ? AND user_id = ?", sessionID, userID).
		Updates(map[string]interface{}{
			"status":    status,
			"last_seen": time.Now(),
		}).Error
}

// GetCollaborationParticipants 获取协作参与者列表
func (r *editingRepository) GetCollaborationParticipants(ctx context.Context, sessionID uuid.UUID) ([]*models.CollaborationParticipant, error) {
	var participants []*models.CollaborationParticipant
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("session_id = ?", sessionID).
		Find(&participants).Error
	return participants, err
}

// GetParticipantBySocketID 根据Socket ID获取参与者
func (r *editingRepository) GetParticipantBySocketID(ctx context.Context, socketID string) (*models.CollaborationParticipant, error) {
	var participant models.CollaborationParticipant
	err := r.db.WithContext(ctx).
		Preload("Session").
		Preload("User").
		First(&participant, "socket_id = ?", socketID).Error
	if err != nil {
		return nil, err
	}
	return &participant, nil
}

// CreateDocumentVersion 创建文档版本
func (r *editingRepository) CreateDocumentVersion(ctx context.Context, version *models.DocumentVersion) error {
	return r.db.WithContext(ctx).Create(version).Error
}

// GetDocumentVersions 获取文档版本列表
func (r *editingRepository) GetDocumentVersions(ctx context.Context, documentID uuid.UUID, page, limit int) ([]*models.DocumentVersion, int64, error) {
	var versions []*models.DocumentVersion
	var total int64

	// 获取总数
	err := r.db.WithContext(ctx).
		Model(&models.DocumentVersion{}).
		Where("document_id = ?", documentID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取版本列表
	offset := (page - 1) * limit
	err = r.db.WithContext(ctx).
		Preload("Editor").
		Where("document_id = ?", documentID).
		Order("version_number DESC").
		Limit(limit).
		Offset(offset).
		Find(&versions).Error

	return versions, total, err
}

// GetDocumentVersion 获取指定版本
func (r *editingRepository) GetDocumentVersion(ctx context.Context, versionID uuid.UUID) (*models.DocumentVersion, error) {
	var version models.DocumentVersion
	err := r.db.WithContext(ctx).
		Preload("Document").
		Preload("Editor").
		First(&version, "id = ?", versionID).Error
	if err != nil {
		return nil, err
	}
	return &version, nil
}

// GetLatestVersion 获取最新版本
func (r *editingRepository) GetLatestVersion(ctx context.Context, documentID uuid.UUID) (*models.DocumentVersion, error) {
	var version models.DocumentVersion
	err := r.db.WithContext(ctx).
		Preload("Editor").
		Where("document_id = ? AND is_published = true", documentID).
		Order("version_number DESC").
		First(&version).Error
	if err != nil {
		return nil, err
	}
	return &version, nil
}

// CompareVersions 比较两个版本
func (r *editingRepository) CompareVersions(ctx context.Context, fromID, toID uuid.UUID) (*models.DocumentVersion, *models.DocumentVersion, error) {
	var fromVersion, toVersion models.DocumentVersion

	// 获取源版本
	err := r.db.WithContext(ctx).First(&fromVersion, "id = ?", fromID).Error
	if err != nil {
		return nil, nil, fmt.Errorf("获取源版本失败: %w", err)
	}

	// 获取目标版本
	err = r.db.WithContext(ctx).First(&toVersion, "id = ?", toID).Error
	if err != nil {
		return nil, nil, fmt.Errorf("获取目标版本失败: %w", err)
	}

	return &fromVersion, &toVersion, nil
}

// DeleteOldVersions 删除旧版本
func (r *editingRepository) DeleteOldVersions(ctx context.Context, documentID uuid.UUID, keepCount int) error {
	// 获取要保留的版本ID
	var versionIDs []uuid.UUID
	err := r.db.WithContext(ctx).
		Model(&models.DocumentVersion{}).
		Where("document_id = ?", documentID).
		Order("version_number DESC").
		Limit(keepCount).
		Pluck("id", &versionIDs).Error
	if err != nil {
		return err
	}

	// 删除不在保留列表中的版本
	if len(versionIDs) > 0 {
		return r.db.WithContext(ctx).
			Delete(&models.DocumentVersion{}, "document_id = ? AND id NOT IN ?", documentID, versionIDs).Error
	} else {
		return r.db.WithContext(ctx).
			Delete(&models.DocumentVersion{}, "document_id = ?", documentID).Error
	}
}

// CreateEditPermission 创建编辑权限
func (r *editingRepository) CreateEditPermission(ctx context.Context, permission *models.EditPermission) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

// GetEditPermission 获取编辑权限
func (r *editingRepository) GetEditPermission(ctx context.Context, documentID, userID uuid.UUID) (*models.EditPermission, error) {
	var permission models.EditPermission
	err := r.db.WithContext(ctx).
		Preload("Document").
		Preload("User").
		Preload("Granter").
		First(&permission, "document_id = ? AND user_id = ?", documentID, userID).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

// UpdateEditPermission 更新编辑权限
func (r *editingRepository) UpdateEditPermission(ctx context.Context, permission *models.EditPermission) error {
	return r.db.WithContext(ctx).Save(permission).Error
}

// DeleteEditPermission 删除编辑权限
func (r *editingRepository) DeleteEditPermission(ctx context.Context, documentID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Delete(&models.EditPermission{}, "document_id = ? AND user_id = ?", documentID, userID).Error
}

// GetDocumentPermissions 获取文档的所有权限
func (r *editingRepository) GetDocumentPermissions(ctx context.Context, documentID uuid.UUID) ([]*models.EditPermission, error) {
	var permissions []*models.EditPermission
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Granter").
		Where("document_id = ? AND is_active = true", documentID).
		Find(&permissions).Error
	return permissions, err
}

// GetUserDocuments 获取用户有权限的文档
func (r *editingRepository) GetUserDocuments(ctx context.Context, userID uuid.UUID) ([]*models.EditPermission, error) {
	var permissions []*models.EditPermission
	err := r.db.WithContext(ctx).
		Preload("Document").
		Where("user_id = ? AND is_active = true", userID).
		Find(&permissions).Error
	return permissions, err
}

// GetEditorConfig 获取编辑器配置
func (r *editingRepository) GetEditorConfig(ctx context.Context, editorType string) (*models.EditorConfig, error) {
	var config models.EditorConfig
	err := r.db.WithContext(ctx).First(&config, "editor_type = ?", editorType).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// CreateEditorConfig 创建编辑器配置
func (r *editingRepository) CreateEditorConfig(ctx context.Context, config *models.EditorConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

// UpdateEditorConfig 更新编辑器配置
func (r *editingRepository) UpdateEditorConfig(ctx context.Context, config *models.EditorConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

// DeleteEditorConfig 删除编辑器配置
func (r *editingRepository) DeleteEditorConfig(ctx context.Context, editorType string) error {
	return r.db.WithContext(ctx).
		Delete(&models.EditorConfig{}, "editor_type = ?", editorType).Error
}

// GetAllEditorConfigs 获取所有编辑器配置
func (r *editingRepository) GetAllEditorConfigs(ctx context.Context) ([]*models.EditorConfig, error) {
	var configs []*models.EditorConfig
	err := r.db.WithContext(ctx).Find(&configs).Error
	return configs, err
}