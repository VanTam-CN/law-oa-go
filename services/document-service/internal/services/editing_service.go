package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// EditingService 编辑服务接口
type EditingService interface {
	// 编辑会话管理
	CreateSession(ctx context.Context, req *CreateSessionRequest) (*SessionResponse, error)
	GetSession(ctx context.Context, sessionID string) (*SessionResponse, error)
	UpdateSessionStatus(ctx context.Context, sessionID string, status *SessionStatus) error
	CloseSession(ctx context.Context, sessionID string) error

	// 编辑操作处理
	HandleEditOperation(ctx context.Context, req *EditOperationRequest) error
	GetDocumentOperations(ctx context.Context, documentID string, page, limit int) (*OperationListResponse, error)

	// 协作功能
	JoinCollaboration(ctx context.Context, req *JoinCollaborationRequest) (*CollaborationResponse, error)
	LeaveCollaboration(ctx context.Context, documentID, userID string) error
	GetCollaborators(ctx context.Context, documentID string) (*CollaboratorListResponse, error)
	BroadcastOperation(ctx context.Context, req *BroadcastOperationRequest) error

	// 版本控制
	CreateVersion(ctx context.Context, req *CreateVersionRequest) (*VersionResponse, error)
	GetVersions(ctx context.Context, documentID string, page, limit int) (*VersionListResponse, error)
	CompareVersions(ctx context.Context, documentID, fromVersion, toVersion string) (*VersionComparisonResponse, error)
	RestoreVersion(ctx context.Context, documentID, versionID string) error

	// 权限管理
	GrantPermission(ctx context.Context, req *GrantPermissionRequest) error
	RevokePermission(ctx context.Context, documentID, userID string) error
	GetDocumentPermissions(ctx context.Context, documentID string) (*PermissionListResponse, error)

	// 编辑器配置
	GetEditorConfig(ctx context.Context, editorType string) (*EditorConfigResponse, error)
	UpdateEditorConfig(ctx context.Context, req *UpdateEditorConfigRequest) error
}

// CreateSessionRequest 创建编辑会话请求
type CreateSessionRequest struct {
	DocumentID string            `json:"document_id" validate:"required,uuid"`
	UserID     string            `json:"user_id" validate:"required,uuid"`
	EditorType string            `json:"editor_type" validate:"required,oneof=rich-text code markdown"`
	Permissions []string         `json:"permissions"`
	Settings   map[string]interface{} `json:"settings"`
}

// SessionResponse 编辑会话响应
type SessionResponse struct {
	SessionID     string                 `json:"session_id"`
	DocumentID    string                 `json:"document_id"`
	UserID        string                 `json:"user_id"`
	EditorType    string                 `json:"editor_type"`
	SessionToken  string                 `json:"session_token"`
	ExpiresAt     time.Time              `json:"expires_at"`
	Collaboration *CollaborationInfo     `json:"collaboration,omitempty"`
	Permissions   []string               `json:"permissions"`
	EditorConfig  *EditorConfigResponse  `json:"editor_config"`
}

// SessionStatus 会话状态
type SessionStatus struct {
	CursorPosition *models.Position `json:"cursor_position"`
	SelectionRange *models.Range    `json:"selection_range"`
	Activity       string           `json:"activity"`
}

// EditOperationRequest 编辑操作请求
type EditOperationRequest struct {
	SessionID     string                 `json:"session_id" validate:"required"`
	OperationType string                 `json:"operation_type" validate:"required,oneof=insert delete retain format cursor selection"`
	OperationData *models.OperationData  `json:"operation_data" validate:"required"`
	YjsState      map[string]uint64      `json:"yjs_state"`
}

// OperationListResponse 操作列表响应
type OperationListResponse struct {
	Operations []*models.EditOperation `json:"operations"`
	Total      int64                   `json:"total"`
	Page       int                     `json:"page"`
	Limit      int                     `json:"limit"`
}

// JoinCollaborationRequest 加入协作请求
type JoinCollaborationRequest struct {
	DocumentID string                `json:"document_id" validate:"required,uuid"`
	SessionID  string                `json:"session_id" validate:"required,uuid"`
	SocketID   string                `json:"socket_id" validate:"required"`
	UserInfo   *CollaborationUserInfo `json:"user_info" validate:"required"`
}

// CollaborationUserInfo 协作用户信息
type CollaborationUserInfo struct {
	Name     string   `json:"name" validate:"required"`
	Avatar   string   `json:"avatar"`
	Color    string   `json:"color"`
	Status   string   `json:"status" validate:"oneof=online away busy"`
	Permissions []string `json:"permissions"`
}

// CollaborationResponse 协作响应
type CollaborationResponse struct {
	RoomName     string                              `json:"room_name"`
	SessionID    string                              `json:"session_id"`
	Participants []*CollaborationParticipantResponse `json:"participants"`
	UserInfo     *CollaborationUserInfo              `json:"user_info"`
}

// CollaborationParticipantResponse 协作参与者响应
type CollaborationParticipantResponse struct {
	UserID      string    `json:"user_id"`
	UserName    string    `json:"user_name"`
	UserAvatar  string    `json:"user_avatar"`
	UserColor   string    `json:"user_color"`
	Status      string    `json:"status"`
	LastSeen    time.Time `json:"last_seen"`
	JoinedAt    time.Time `json:"joined_at"`
	Permissions []string  `json:"permissions"`
}

// CollaborationInfo 协作信息
type CollaborationInfo struct {
	RoomName     string `json:"room_name"`
	ActiveUsers  int    `json:"active_users"`
	MaxUsers     int    `json:"max_users"`
	CanCollaborate bool  `json:"can_collaborate"`
}

// CollaboratorListResponse 协作者列表响应
type CollaboratorListResponse struct {
	Participants []*CollaborationParticipantResponse `json:"participants"`
	ActiveUsers  int                                 `json:"active_users"`
	MaxUsers     int                                 `json:"max_users"`
}

// BroadcastOperationRequest 广播操作请求
type BroadcastOperationRequest struct {
	DocumentID string                 `json:"document_id" validate:"required,uuid"`
	UserID     string                 `json:"user_id" validate:"required,uuid"`
	OperationType string              `json:"operation_type" validate:"required"`
	OperationData *models.OperationData `json:"operation_data" validate:"required"`
	ExcludedUsers []string            `json:"excluded_users"`
}

// CreateVersionRequest 创建版本请求
type CreateVersionRequest struct {
	DocumentID     string                 `json:"document_id" validate:"required,uuid"`
	Title          string                 `json:"title" validate:"required,max=255"`
	EditSummary    string                 `json:"edit_summary"`
	IsMajorVersion bool                   `json:"is_major_version"`
	IsPublished    bool                   `json:"is_published"`
	ContentDelta   *models.DeltaData      `json:"content_delta"`
	SnapshotData   []byte                 `json:"snapshot_data"`
}

// VersionResponse 版本响应
type VersionResponse struct {
	VersionID      string                 `json:"version_id"`
	DocumentID     string                 `json:"document_id"`
	VersionNumber  int                    `json:"version_number"`
	Title          string                 `json:"title"`
	ContentHash    string                 `json:"content_hash"`
	EditorID       string                 `json:"editor_id,omitempty"`
	EditSummary    string                 `json:"edit_summary"`
	IsMajorVersion bool                   `json:"is_major_version"`
	IsPublished    bool                   `json:"is_published"`
	FileSize       int64                  `json:"file_size"`
	CharacterCount int                    `json:"character_count"`
	CreatedAt      time.Time              `json:"created_at"`
	Editor         *UserResponse          `json:"editor,omitempty"`
}

// VersionListResponse 版本列表响应
type VersionListResponse struct {
	Versions []*VersionResponse `json:"versions"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	Limit    int                `json:"limit"`
}

// VersionComparisonResponse 版本比较响应
type VersionComparisonResponse struct {
	FromVersion *VersionResponse `json:"from_version"`
	ToVersion   *VersionResponse `json:"to_version"`
	Differences []VersionDiff    `json:"differences"`
}

// VersionDiff 版本差异
type VersionDiff struct {
	Type        string `json:"type"`
	Position    int    `json:"position"`
	Content     string `json:"content"`
	Length      int    `json:"length"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

// GrantPermissionRequest 授予权限请求
type GrantPermissionRequest struct {
	DocumentID   string   `json:"document_id" validate:"required,uuid"`
	UserID       string   `json:"user_id" validate:"required,uuid"`
	Role         string   `json:"role" validate:"required,oneof=viewer commenter editor admin"`
	Permissions  []string `json:"permissions"`
	ExpiresAt    *time.Time `json:"expires_at"`
}

// PermissionListResponse 权限列表响应
type PermissionListResponse struct {
	Permissions []*PermissionResponse `json:"permissions"`
}

// PermissionResponse 权限响应
type PermissionResponse struct {
	UserID      string     `json:"user_id"`
	UserName    string     `json:"user_name"`
	Role        string     `json:"role"`
	Permissions []string   `json:"permissions"`
	GrantedBy   string     `json:"granted_by"`
	GrantedAt   time.Time  `json:"granted_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	IsActive    bool       `json:"is_active"`
}

// EditorConfigResponse 编辑器配置响应
type EditorConfigResponse struct {
	EditorType string                 `json:"editor_type"`
	Theme      string                 `json:"theme"`
	FontFamily string                 `json:"font_family"`
	FontSize   int                    `json:"font_size"`
	TabSize    int                    `json:"tab_size"`
	WordWrap   bool                   `json:"word_wrap"`
	LineNumbers bool                  `json:"line_numbers"`
	Settings   map[string]interface{} `json:"settings"`
}

// UpdateEditorConfigRequest 更新编辑器配置请求
type UpdateEditorConfigRequest struct {
	EditorType string                 `json:"editor_type" validate:"required"`
	Theme      string                 `json:"theme"`
	FontFamily string                 `json:"font_family"`
	FontSize   int                    `json:"font_size"`
	TabSize    int                    `json:"tab_size"`
	WordWrap   bool                   `json:"word_wrap"`
	LineNumbers bool                  `json:"line_numbers"`
	Settings   map[string]interface{} `json:"settings"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
}

// editingService 编辑服务实现
type editingService struct {
	editRepo repositories.EditingRepository
	docRepo  repositories.DocumentRepository
	userRepo repositories.UserRepository
	storage  StorageService
	notify   NotificationService
	logger   *logrus.Logger
}

// NewEditingService 创建编辑服务实例
func NewEditingService(
	editRepo repositories.EditingRepository,
	docRepo repositories.DocumentRepository,
	userRepo repositories.UserRepository,
	storage StorageService,
	notify NotificationService,
	logger *logrus.Logger,
) EditingService {
	return &editingService{
		editRepo: editRepo,
		docRepo:  docRepo,
		userRepo: userRepo,
		storage:  storage,
		notify:   notify,
		logger:   logger,
	}
}

// CreateSession 创建编辑会话
func (s *editingService) CreateSession(ctx context.Context, req *CreateSessionRequest) (*SessionResponse, error) {
	// 验证用户和文档权限
	documentUUID := uuid.MustParse(req.DocumentID)
	userUUID := uuid.MustParse(req.UserID)

	// 检查文档是否存在
	document, err := s.docRepo.GetByID(ctx, documentUUID)
	if err != nil {
		return nil, fmt.Errorf("获取文档失败: %w", err)
	}

	// 检查用户权限
	permission, err := s.editRepo.GetEditPermission(ctx, documentUUID, userUUID)
	if err != nil && err.Error() != "record not found" {
		return nil, fmt.Errorf("获取权限失败: %w", err)
	}

	// 如果没有权限，创建默认读取权限
	if permission == nil {
		permission = &models.EditPermission{
			DocumentID: documentUUID,
			UserID:     userUUID,
			Role:       "viewer",
			Permissions: []string{"read"},
			IsActive:   true,
		}
	}

	// 检查权限是否允许编辑
	canEdit := false
	for _, perm := range permission.Permissions {
		if perm == "write" || perm == "admin" || perm == "owner" {
			canEdit = true
			break
		}
	}

	// 创建编辑会话
	session := &models.EditSession{
		DocumentID: documentUUID,
		UserID:     userUUID,
		EditorType: req.EditorType,
		ExpiresAt:  time.Now().Add(24 * time.Hour), // 24小时过期
	}

	err = s.editRepo.CreateEditSession(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("创建编辑会话失败: %w", err)
	}

	// 获取协作会话信息
	collaborationInfo, _ := s.getCollaborationInfo(ctx, documentUUID)

	// 获取编辑器配置
	editorConfig, _ := s.editRepo.GetEditorConfig(ctx, req.EditorType)

	// 构建响应
	response := &SessionResponse{
		SessionID:     session.ID.String(),
		DocumentID:    session.DocumentID.String(),
		UserID:        session.UserID.String(),
		EditorType:    session.EditorType,
		SessionToken:  session.SessionToken,
		ExpiresAt:     session.ExpiresAt,
		Collaboration: collaborationInfo,
		Permissions:   permission.Permissions,
		EditorConfig:  s.toEditorConfigResponse(editorConfig),
	}

	s.logger.WithFields(logrus.Fields{
		"session_id":  session.ID,
		"document_id": documentUUID,
		"user_id":     userUUID,
		"editor_type": req.EditorType,
	}).Info("创建编辑会话成功")

	return response, nil
}

// GetSession 获取编辑会话
func (s *editingService) GetSession(ctx context.Context, sessionID string) (*SessionResponse, error) {
	sessionUUID := uuid.MustParse(sessionID)

	session, err := s.editRepo.GetEditSession(ctx, sessionUUID)
	if err != nil {
		return nil, fmt.Errorf("获取编辑会话失败: %w", err)
	}

	// 检查会话是否过期
	if session.IsExpired() {
		return nil, fmt.Errorf("编辑会话已过期")
	}

	// 获取权限信息
	permission, err := s.editRepo.GetEditPermission(ctx, session.DocumentID, session.UserID)
	if err != nil {
		permission = &models.EditPermission{
			Permissions: []string{"read"},
		}
	}

	// 获取协作信息
	collaborationInfo, _ := s.getCollaborationInfo(ctx, session.DocumentID)

	// 获取编辑器配置
	editorConfig, _ := s.editRepo.GetEditorConfig(ctx, session.EditorType)

	response := &SessionResponse{
		SessionID:     session.ID.String(),
		DocumentID:    session.DocumentID.String(),
		UserID:        session.UserID.String(),
		EditorType:    session.EditorType,
		SessionToken:  session.SessionToken,
		ExpiresAt:     session.ExpiresAt,
		Collaboration: collaborationInfo,
		Permissions:   permission.Permissions,
		EditorConfig:  s.toEditorConfigResponse(editorConfig),
	}

	return response, nil
}

// UpdateSessionStatus 更新会话状态
func (s *editingService) UpdateSessionStatus(ctx context.Context, sessionID string, status *SessionStatus) error {
	sessionUUID := uuid.MustParse(sessionID)

	session, err := s.editRepo.GetEditSession(ctx, sessionUUID)
	if err != nil {
		return fmt.Errorf("获取编辑会话失败: %w", err)
	}

	// 更新会话状态
	session.CursorPosition = status.CursorPosition
	session.SelectionRange = status.SelectionRange
	session.ActivityAt = time.Now()

	err = s.editRepo.UpdateEditSession(ctx, session)
	if err != nil {
		return fmt.Errorf("更新编辑会话失败: %w", err)
	}

	return nil
}

// CloseSession 关闭编辑会话
func (s *editingService) CloseSession(ctx context.Context, sessionID string) error {
	sessionUUID := uuid.MustParse(sessionID)

	// 获取会话信息
	session, err := s.editRepo.GetEditSession(ctx, sessionUUID)
	if err != nil {
		return fmt.Errorf("获取编辑会话失败: %w", err)
	}

	// 如果是协作会话，离开协作
	_, err = s.editRepo.GetCollaborationSession(ctx, session.DocumentID)
	if err == nil {
		err = s.LeaveCollaboration(ctx, session.DocumentID.String(), session.UserID.String())
		if err != nil {
			s.logger.WithError(err).Warn("离开协作会话失败")
		}
	}

	// 删除编辑会话
	err = s.editRepo.DeleteEditSession(ctx, sessionUUID)
	if err != nil {
		return fmt.Errorf("删除编辑会话失败: %w", err)
	}

	s.logger.WithField("session_id", sessionUUID).Info("关闭编辑会话成功")

	return nil
}

// HandleEditOperation 处理编辑操作
func (s *editingService) HandleEditOperation(ctx context.Context, req *EditOperationRequest) error {
	sessionUUID := uuid.MustParse(req.SessionID)

	// 获取会话信息
	session, err := s.editRepo.GetEditSession(ctx, sessionUUID)
	if err != nil {
		return fmt.Errorf("获取编辑会话失败: %w", err)
	}

	// 检查会话是否过期
	if session.IsExpired() {
		return fmt.Errorf("编辑会话已过期")
	}

	// 检查编辑权限
	permission, err := s.editRepo.GetEditPermission(ctx, session.DocumentID, session.UserID)
	if err != nil || !permission.CanEdit() {
		return fmt.Errorf("没有编辑权限")
	}

	// 创建编辑操作
	operation := &models.EditOperation{
		DocumentID:     session.DocumentID,
		UserID:         session.UserID,
		SessionID:      session.ID,
		OperationType:  req.OperationType,
		OperationData:  req.OperationData,
		YjsStateVector: req.YjsState,
		Timestamp:      time.Now(),
		Applied:        false,
	}

	err = s.editRepo.CreateEditOperation(ctx, operation)
	if err != nil {
		return fmt.Errorf("创建编辑操作失败: %w", err)
	}

	// 广播操作给其他协作者
	err = s.BroadcastOperation(ctx, &BroadcastOperationRequest{
		DocumentID:     session.DocumentID.String(),
		UserID:         session.UserID.String(),
		OperationType:  req.OperationType,
		OperationData:  req.OperationData,
		ExcludedUsers:  []string{session.UserID.String()},
	})
	if err != nil {
		s.logger.WithError(err).Warn("广播编辑操作失败")
	}

	return nil
}

// GetDocumentOperations 获取文档操作列表
func (s *editingService) GetDocumentOperations(ctx context.Context, documentID string, page, limit int) (*OperationListResponse, error) {
	documentUUID := uuid.MustParse(documentID)

	operations, err := s.editRepo.GetEditOperations(ctx, documentUUID, limit, (page-1)*limit)
	if err != nil {
		return nil, fmt.Errorf("获取编辑操作失败: %w", err)
	}

	// 计算总数
	total, err := s.countDocumentOperations(ctx, documentUUID)
	if err != nil {
		return nil, fmt.Errorf("计算操作总数失败: %w", err)
	}

	response := &OperationListResponse{
		Operations: operations,
		Total:      total,
		Page:       page,
		Limit:      limit,
	}

	return response, nil
}

// JoinCollaboration 加入协作
func (s *editingService) JoinCollaboration(ctx context.Context, req *JoinCollaborationRequest) (*CollaborationResponse, error) {
	documentUUID := uuid.MustParse(req.DocumentID)
	sessionUUID := uuid.MustParse(req.SessionID)

	// 获取编辑会话
	session, err := s.editRepo.GetEditSession(ctx, sessionUUID)
	if err != nil {
		return nil, fmt.Errorf("获取编辑会话失败: %w", err)
	}

	// 检查编辑权限
	permission, err := s.editRepo.GetEditPermission(ctx, documentUUID, session.UserID)
	if err != nil || !permission.CanEdit() {
		return nil, fmt.Errorf("没有协作权限")
	}

	// 获取或创建协作会话
	collabSession, err := s.editRepo.GetCollaborationSession(ctx, documentUUID)
	if err != nil {
		// 创建新的协作会话
		collabSession = &models.CollaborationSession{
			DocumentID:   documentUUID,
			RoomName:     "doc_" + documentUUID.String(),
			ActiveUsers:  0,
			MaxUsers:     10,
			LastActivity: time.Now(),
		}
		err = s.editRepo.CreateCollaborationSession(ctx, collabSession)
		if err != nil {
			return nil, fmt.Errorf("创建协作会话失败: %w", err)
		}
	}

	// 检查用户数量限制
	if collabSession.ActiveUsers >= collabSession.MaxUsers {
		return nil, fmt.Errorf("协作用户数量已达上限")
	}

	// 检查用户是否已经在协作中
	existingParticipants, _ := s.editRepo.GetCollaborationParticipants(ctx, collabSession.ID)
	for _, participant := range existingParticipants {
		if participant.UserID == session.UserID {
			return nil, fmt.Errorf("用户已在协作中")
		}
	}

	// 创建协作参与者
	participant := &models.CollaborationParticipant{
		SessionID:   collabSession.ID,
		UserID:      session.UserID,
		SocketID:    req.SocketID,
		UserName:    req.UserInfo.Name,
		UserAvatar:  req.UserInfo.Avatar,
		UserColor:   req.UserInfo.Color,
		Status:      req.UserInfo.Status,
		Permissions: req.UserInfo.Permissions,
		LastSeen:    time.Now(),
	}

	err = s.editRepo.AddCollaborationParticipant(ctx, participant)
	if err != nil {
		return nil, fmt.Errorf("添加协作参与者失败: %w", err)
	}

	// 更新协作会话活跃用户数
	collabSession.ActiveUsers++
	collabSession.LastActivity = time.Now()
	err = s.editRepo.UpdateCollaborationSession(ctx, collabSession)
	if err != nil {
		s.logger.WithError(err).Warn("更新协作会话失败")
	}

	// 获取所有参与者
	participants, err := s.editRepo.GetCollaborationParticipants(ctx, collabSession.ID)
	if err != nil {
		s.logger.WithError(err).Warn("获取协作参与者失败")
	}

	response := &CollaborationResponse{
		RoomName:     collabSession.RoomName,
		SessionID:    collabSession.ID.String(),
		Participants: s.toCollaborationParticipantResponses(participants),
		UserInfo:     req.UserInfo,
	}

	s.logger.WithFields(logrus.Fields{
		"document_id": documentUUID,
		"user_id":     session.UserID,
		"room_name":   collabSession.RoomName,
	}).Info("用户加入协作成功")

	return response, nil
}

// LeaveCollaboration 离开协作
func (s *editingService) LeaveCollaboration(ctx context.Context, documentID, userID string) error {
	documentUUID := uuid.MustParse(documentID)
	userUUID := uuid.MustParse(userID)

	// 获取协作会话
	collabSession, err := s.editRepo.GetCollaborationSession(ctx, documentUUID)
	if err != nil {
		return fmt.Errorf("获取协作会话失败: %w", err)
	}

	// 移除协作参与者
	err = s.editRepo.RemoveCollaborationParticipant(ctx, collabSession.ID, userUUID)
	if err != nil {
		return fmt.Errorf("移除协作参与者失败: %w", err)
	}

	// 更新协作会话活跃用户数
	if collabSession.ActiveUsers > 0 {
		collabSession.ActiveUsers--
		collabSession.LastActivity = time.Now()
		err = s.editRepo.UpdateCollaborationSession(ctx, collabSession)
		if err != nil {
			s.logger.WithError(err).Warn("更新协作会话失败")
		}
	}

	// 如果没有活跃用户，删除协作会话
	if collabSession.ActiveUsers == 0 {
		err = s.editRepo.DeleteCollaborationSession(ctx, documentUUID)
		if err != nil {
			s.logger.WithError(err).Warn("删除协作会话失败")
		}
	}

	s.logger.WithFields(logrus.Fields{
		"document_id": documentUUID,
		"user_id":     userUUID,
	}).Info("用户离开协作成功")

	return nil
}

// GetCollaborators 获取协作者列表
func (s *editingService) GetCollaborators(ctx context.Context, documentID string) (*CollaboratorListResponse, error) {
	documentUUID := uuid.MustParse(documentID)

	// 获取协作会话
	collabSession, err := s.editRepo.GetCollaborationSession(ctx, documentUUID)
	if err != nil {
		return &CollaboratorListResponse{
			Participants: []*CollaborationParticipantResponse{},
			ActiveUsers:  0,
			MaxUsers:     10,
		}, nil
	}

	// 获取参与者列表
	participants, err := s.editRepo.GetCollaborationParticipants(ctx, collabSession.ID)
	if err != nil {
		return nil, fmt.Errorf("获取协作参与者失败: %w", err)
	}

	response := &CollaboratorListResponse{
		Participants: s.toCollaborationParticipantResponses(participants),
		ActiveUsers:  collabSession.ActiveUsers,
		MaxUsers:     collabSession.MaxUsers,
	}

	return response, nil
}

// BroadcastOperation 广播操作
func (s *editingService) BroadcastOperation(ctx context.Context, req *BroadcastOperationRequest) error {
	documentUUID := uuid.MustParse(req.DocumentID)

	// 获取协作会话
	collabSession, err := s.editRepo.GetCollaborationSession(ctx, documentUUID)
	if err != nil {
		return fmt.Errorf("获取协作会话失败: %w", err)
	}

	// 获取参与者列表
	participants, err := s.editRepo.GetCollaborationParticipants(ctx, collabSession.ID)
	if err != nil {
		return fmt.Errorf("获取协作参与者失败: %w", err)
	}

	// 过滤排除的用户
	targetUsers := make([]string, 0)
	for _, participant := range participants {
		shouldExclude := false
		for _, excludedUser := range req.ExcludedUsers {
			if participant.UserID.String() == excludedUser {
				shouldExclude = true
				break
			}
		}
		if !shouldExclude {
			targetUsers = append(targetUsers, participant.UserID.String())
		}
	}

	// 通过WebSocket广播操作
	// 这里需要与WebSocket服务集成
	// 实际实现中会调用WebSocket服务的广播方法

	s.logger.WithFields(logrus.Fields{
		"document_id":   documentUUID,
		"user_id":       req.UserID,
		"operation_type": req.OperationType,
		"target_users":  len(targetUsers),
	}).Info("广播编辑操作")

	return nil
}

// CreateVersion 创建文档版本
func (s *editingService) CreateVersion(ctx context.Context, req *CreateVersionRequest) (*VersionResponse, error) {
	documentUUID := uuid.MustParse(req.DocumentID)

	// 获取文档信息
	document, err := s.docRepo.GetByID(ctx, documentUUID)
	if err != nil {
		return nil, fmt.Errorf("获取文档失败: %w", err)
	}

	// 获取最新版本号
	latestVersion, err := s.editRepo.GetLatestVersion(ctx, documentUUID)
	versionNumber := 1
	if err == nil && latestVersion != nil {
		versionNumber = latestVersion.VersionNumber + 1
	}

	// 计算内容哈希
	contentHash := s.calculateContentHash(req.ContentDelta)

	// 如果有快照数据，保存到存储
	var snapshotPath string
	if len(req.SnapshotData) > 0 {
		snapshotPath = fmt.Sprintf("documents/%s/versions/%d/snapshot.bin",
			documentUUID.String(), versionNumber)
		err = s.storage.Put(ctx, snapshotPath, req.SnapshotData)
		if err != nil {
			s.logger.WithError(err).Warn("保存版本快照失败")
		}
	}

	// 创建版本记录
	version := &models.DocumentVersion{
		DocumentID:     documentUUID,
		VersionNumber:  versionNumber,
		Title:          req.Title,
		ContentHash:    contentHash,
		ContentDelta:   req.ContentDelta,
		SnapshotPath:   snapshotPath,
		EditSummary:    req.EditSummary,
		IsMajorVersion: req.IsMajorVersion,
		IsPublished:    req.IsPublished,
		FileSize:       int64(len(req.SnapshotData)),
		CharacterCount: s.calculateCharacterCount(req.ContentDelta),
		CreatedAt:      time.Now(),
	}

	// 设置编辑者（这里从上下文获取，实际实现中需要从JWT获取）
	// version.EditorID = getCurrentUserID(ctx)

	err = s.editRepo.CreateDocumentVersion(ctx, version)
	if err != nil {
		return nil, fmt.Errorf("创建文档版本失败: %w", err)
	}

	// 构建响应
	response := &VersionResponse{
		VersionID:      version.ID.String(),
		DocumentID:     version.DocumentID.String(),
		VersionNumber:  version.VersionNumber,
		Title:          version.Title,
		ContentHash:    version.ContentHash,
		EditSummary:    version.EditSummary,
		IsMajorVersion: version.IsMajorVersion,
		IsPublished:    version.IsPublished,
		FileSize:       version.FileSize,
		CharacterCount: version.CharacterCount,
		CreatedAt:      version.CreatedAt,
	}

	s.logger.WithFields(logrus.Fields{
		"document_id":    documentUUID,
		"version_number": versionNumber,
		"is_major":       req.IsMajorVersion,
	}).Info("创建文档版本成功")

	return response, nil
}

// GetVersions 获取版本列表
func (s *editingService) GetVersions(ctx context.Context, documentID string, page, limit int) (*VersionListResponse, error) {
	documentUUID := uuid.MustParse(documentID)

	versions, total, err := s.editRepo.GetDocumentVersions(ctx, documentUUID, page, limit)
	if err != nil {
		return nil, fmt.Errorf("获取文档版本失败: %w", err)
	}

	response := &VersionListResponse{
		Versions: s.toVersionResponses(versions),
		Total:    total,
		Page:     page,
		Limit:    limit,
	}

	return response, nil
}

// CompareVersions 比较版本
func (s *editingService) CompareVersions(ctx context.Context, documentID, fromVersion, toVersion string) (*VersionComparisonResponse, error) {
	fromUUID := uuid.MustParse(fromVersion)
	toUUID := uuid.MustParse(toVersion)

	fromVersionModel, toVersionModel, err := s.editRepo.CompareVersions(ctx, fromUUID, toUUID)
	if err != nil {
		return nil, fmt.Errorf("比较版本失败: %w", err)
	}

	// 计算差异（这里简化实现，实际需要实现完整的差异算法）
	differences := s.calculateDifferences(fromVersionModel.ContentDelta, toVersionModel.ContentDelta)

	response := &VersionComparisonResponse{
		FromVersion: s.toVersionResponse(fromVersionModel),
		ToVersion:   s.toVersionResponse(toVersionModel),
		Differences: differences,
	}

	return response, nil
}

// RestoreVersion 恢复版本
func (s *editingService) RestoreVersion(ctx context.Context, documentID, versionID string) error {
	documentUUID := uuid.MustParse(documentID)
	versionUUID := uuid.MustParse(versionID)

	// 获取版本信息
	version, err := s.editRepo.GetDocumentVersion(ctx, versionUUID)
	if err != nil {
		return fmt.Errorf("获取版本信息失败: %w", err)
	}

	// 这里需要实现版本恢复逻辑
	// 1. 恢复文档内容
	// 2. 创建新的版本记录
	// 3. 更新文档状态

	s.logger.WithFields(logrus.Fields{
		"document_id": documentUUID,
		"version_id":  versionUUID,
	}).Info("恢复文档版本")

	return nil
}

// GrantPermission 授予权限
func (s *editingService) GrantPermission(ctx context.Context, req *GrantPermissionRequest) error {
	documentUUID := uuid.MustParse(req.DocumentID)
	userUUID := uuid.MustParse(req.UserID)

	// 检查文档是否存在
	_, err := s.docRepo.GetByID(ctx, documentUUID)
	if err != nil {
		return fmt.Errorf("获取文档失败: %w", err)
	}

	// 检查用户是否存在
	_, err = s.userRepo.GetByID(ctx, userUUID)
	if err != nil {
		return fmt.Errorf("获取用户失败: %w", err)
	}

	// 创建权限记录
	permission := &models.EditPermission{
		DocumentID:  documentUUID,
		UserID:      userUUID,
		Role:        req.Role,
		Permissions: req.Permissions,
		ExpiresAt:   req.ExpiresAt,
		IsActive:    true,
		GrantedAt:   time.Now(),
	}

	err = s.editRepo.CreateEditPermission(ctx, permission)
	if err != nil {
		return fmt.Errorf("创建权限失败: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"document_id": documentUUID,
		"user_id":     userUUID,
		"role":        req.Role,
	}).Info("授予权限成功")

	return nil
}

// RevokePermission 撤销权限
func (s *editingService) RevokePermission(ctx context.Context, documentID, userID string) error {
	documentUUID := uuid.MustParse(documentID)
	userUUID := uuid.MustParse(userID)

	err := s.editRepo.DeleteEditPermission(ctx, documentUUID, userUUID)
	if err != nil {
		return fmt.Errorf("撤销权限失败: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"document_id": documentUUID,
		"user_id":     userUUID,
	}).Info("撤销权限成功")

	return nil
}

// GetDocumentPermissions 获取文档权限
func (s *editingService) GetDocumentPermissions(ctx context.Context, documentID string) (*PermissionListResponse, error) {
	documentUUID := uuid.MustParse(documentID)

	permissions, err := s.editRepo.GetDocumentPermissions(ctx, documentUUID)
	if err != nil {
		return nil, fmt.Errorf("获取文档权限失败: %w", err)
	}

	response := &PermissionListResponse{
		Permissions: s.toPermissionResponses(permissions),
	}

	return response, nil
}

// GetEditorConfig 获取编辑器配置
func (s *editingService) GetEditorConfig(ctx context.Context, editorType string) (*EditorConfigResponse, error) {
	config, err := s.editRepo.GetEditorConfig(ctx, editorType)
	if err != nil {
		// 返回默认配置
		return &EditorConfigResponse{
			EditorType: editorType,
			Theme:      "default",
			FontFamily: "monospace",
			FontSize:   14,
			TabSize:    4,
			WordWrap:   true,
			LineNumbers: true,
			Settings:   make(map[string]interface{}),
		}, nil
	}

	return s.toEditorConfigResponse(config), nil
}

// UpdateEditorConfig 更新编辑器配置
func (s *editingService) UpdateEditorConfig(ctx context.Context, req *UpdateEditorConfigRequest) error {
	config, err := s.editRepo.GetEditorConfig(ctx, req.EditorType)
	if err != nil {
		// 创建新配置
		config = &models.EditorConfig{
			EditorType: req.EditorType,
		}
	}

	// 更新配置
	config.Theme = req.Theme
	config.FontFamily = req.FontFamily
	config.FontSize = req.FontSize
	config.TabSize = req.TabSize
	config.WordWrap = req.WordWrap
	config.LineNumbers = req.LineNumbers
	config.Settings = req.Settings

	if config.ID == uuid.Nil {
		err = s.editRepo.CreateEditorConfig(ctx, config)
	} else {
		err = s.editRepo.UpdateEditorConfig(ctx, config)
	}

	if err != nil {
		return fmt.Errorf("更新编辑器配置失败: %w", err)
	}

	s.logger.WithField("editor_type", req.EditorType).Info("更新编辑器配置成功")

	return nil
}

// 辅助方法

// getCollaborationInfo 获取协作信息
func (s *editingService) getCollaborationInfo(ctx context.Context, documentID uuid.UUID) (*CollaborationInfo, error) {
	collabSession, err := s.editRepo.GetCollaborationSession(ctx, documentID)
	if err != nil {
		return nil, err
	}

	info := &CollaborationInfo{
		RoomName:      collabSession.RoomName,
		ActiveUsers:   collabSession.ActiveUsers,
		MaxUsers:      collabSession.MaxUsers,
		CanCollaborate: collabSession.ActiveUsers < collabSession.MaxUsers,
	}

	return info, nil
}

// calculateContentHash 计算内容哈希
func (s *editingService) calculateContentHash(delta *models.DeltaData) string {
	if delta == nil {
		return ""
	}

	h := sha256.New()
	// 这里简化实现，实际需要序列化delta数据
	h.Write([]byte(fmt.Sprintf("%v", delta)))
	return hex.EncodeToString(h.Sum(nil))
}

// calculateCharacterCount 计算字符数
func (s *editingService) calculateCharacterCount(delta *models.DeltaData) int {
	if delta == nil {
		return 0
	}

	// 简化实现，实际需要解析delta数据计算字符数
	count := 0
	for _, op := range delta.Ops {
		if str, ok := op.(string); ok {
			count += len(str)
		}
	}
	return count
}

// countDocumentOperations 计算文档操作总数
func (s *editingService) countDocumentOperations(ctx context.Context, documentID uuid.UUID) (int64, error) {
	// 这里需要在仓库层添加计数方法
	// 简化实现
	return 0, nil
}

// calculateDifferences 计算版本差异
func (s *editingService) calculateDifferences(fromDelta, toDelta *models.DeltaData) []VersionDiff {
	// 这里需要实现完整的差异算法
	// 简化实现
	return []VersionDiff{}
}

// toEditorConfigResponse 转换编辑器配置响应
func (s *editingService) toEditorConfigResponse(config *models.EditorConfig) *EditorConfigResponse {
	if config == nil {
		return nil
	}

	return &EditorConfigResponse{
		EditorType:  config.EditorType,
		Theme:       config.Theme,
		FontFamily:  config.FontFamily,
		FontSize:    config.FontSize,
		TabSize:     config.TabSize,
		WordWrap:    config.WordWrap,
		LineNumbers: config.LineNumbers,
		Settings:    config.Settings,
	}
}

// toCollaborationParticipantResponses 转换协作参与者响应
func (s *editingService) toCollaborationParticipantResponses(participants []*models.CollaborationParticipant) []*CollaborationParticipantResponse {
	responses := make([]*CollaborationParticipantResponse, len(participants))
	for i, participant := range participants {
		responses[i] = &CollaborationParticipantResponse{
			UserID:      participant.UserID.String(),
			UserName:    participant.UserName,
			UserAvatar:  participant.UserAvatar,
			UserColor:   participant.UserColor,
			Status:      participant.Status,
			LastSeen:    participant.LastSeen,
			JoinedAt:    participant.JoinedAt,
			Permissions: participant.Permissions,
		}
	}
	return responses
}

// toVersionResponses 转换版本响应
func (s *editingService) toVersionResponses(versions []*models.DocumentVersion) []*VersionResponse {
	responses := make([]*VersionResponse, len(versions))
	for i, version := range versions {
		responses[i] = s.toVersionResponse(version)
	}
	return responses
}

// toVersionResponse 转换单个版本响应
func (s *editingService) toVersionResponse(version *models.DocumentVersion) *VersionResponse {
	response := &VersionResponse{
		VersionID:      version.ID.String(),
		DocumentID:     version.DocumentID.String(),
		VersionNumber:  version.VersionNumber,
		Title:          version.Title,
		ContentHash:    version.ContentHash,
		EditSummary:    version.EditSummary,
		IsMajorVersion: version.IsMajorVersion,
		IsPublished:    version.IsPublished,
		FileSize:       version.FileSize,
		CharacterCount: version.CharacterCount,
		CreatedAt:      version.CreatedAt,
	}

	if version.Editor != nil {
		response.EditorID = version.Editor.ID.String()
		response.Editor = &UserResponse{
			ID:       version.Editor.ID.String(),
			Username: version.Editor.Username,
			Email:    version.Editor.Email,
			Avatar:   version.Editor.Avatar,
		}
	}

	return response
}

// toPermissionResponses 转换权限响应
func (s *editingService) toPermissionResponses(permissions []*models.EditPermission) []*PermissionResponse {
	responses := make([]*PermissionResponse, len(permissions))
	for i, permission := range permissions {
		responses[i] = &PermissionResponse{
			UserID:      permission.UserID.String(),
			UserName:    permission.User.Username,
			Role:        permission.Role,
			Permissions: permission.Permissions,
			GrantedBy:   permission.Granter.Username,
			GrantedAt:   permission.GrantedAt,
			ExpiresAt:   permission.ExpiresAt,
			IsActive:    permission.IsActive,
		}
	}
	return responses
}