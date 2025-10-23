package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockEditRepository 模拟编辑仓库
type MockEditRepository struct {
	mock.Mock
}

// MockDocumentRepository 模拟文档仓库
type MockDocumentRepository struct {
	mock.Mock
}

// MockUserRepository 模拟用户仓库
type MockUserRepository struct {
	mock.Mock
}

// MockStorageService 模拟存储服务
type MockStorageService struct {
	mock.Mock
}

// MockNotificationService 模拟通知服务
type MockNotificationService struct {
	mock.Mock
}

func (m *MockEditRepository) CreateEditSession(ctx context.Context, session *models.EditSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockEditRepository) GetEditSession(ctx context.Context, id uuid.UUID) (*models.EditSession, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.EditSession), args.Error(1)
}

func (m *MockEditRepository) GetEditSessionByToken(ctx context.Context, token string) (*models.EditSession, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EditSession), args.Error(1)
}

func (m *MockEditRepository) UpdateEditSession(ctx context.Context, session *models.EditSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockEditRepository) DeleteEditSession(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEditRepository) GetActiveEditSessionsByDocument(ctx context.Context, documentID uuid.UUID) ([]*models.EditSession, error) {
	args := m.Called(ctx, documentID)
	return args.Get(0).([]*models.EditSession), args.Error(1)
}

func (m *MockEditRepository) CleanExpiredSessions(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockEditRepository) CreateEditOperation(ctx context.Context, operation *models.EditOperation) error {
	args := m.Called(ctx, operation)
	return args.Error(0)
}

func (m *MockEditRepository) GetEditOperations(ctx context.Context, documentID uuid.UUID, limit, offset int) ([]*models.EditOperation, error) {
	args := m.Called(ctx, documentID, limit, offset)
	return args.Get(0).([]*models.EditOperation), args.Error(1)
}

func (m *MockEditRepository) GetPendingOperations(ctx context.Context, documentID uuid.UUID) ([]*models.EditOperation, error) {
	args := m.Called(ctx, documentID)
	return args.Get(0).([]*models.EditOperation), args.Error(1)
}

func (m *MockEditRepository) MarkOperationsApplied(ctx context.Context, operationIDs []uuid.UUID) error {
	args := m.Called(ctx, operationIDs)
	return args.Error(0)
}

func (m *MockEditRepository) DeleteOldOperations(ctx context.Context, beforeTime time.Time) error {
	args := m.Called(ctx, beforeTime)
	return args.Error(0)
}

func (m *MockEditRepository) CreateCollaborationSession(ctx context.Context, session *models.CollaborationSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockEditRepository) GetCollaborationSession(ctx context.Context, documentID uuid.UUID) (*models.CollaborationSession, error) {
	args := m.Called(ctx, documentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CollaborationSession), args.Error(1)
}

func (m *MockEditRepository) UpdateCollaborationSession(ctx context.Context, session *models.CollaborationSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockEditRepository) DeleteCollaborationSession(ctx context.Context, documentID uuid.UUID) error {
	args := m.Called(ctx, documentID)
	return args.Error(0)
}

func (m *MockEditRepository) GetActiveCollaborationSessions(ctx context.Context) ([]*models.CollaborationSession, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.CollaborationSession), args.Error(1)
}

func (m *MockEditRepository) AddCollaborationParticipant(ctx context.Context, participant *models.CollaborationParticipant) error {
	args := m.Called(ctx, participant)
	return args.Error(0)
}

func (m *MockEditRepository) RemoveCollaborationParticipant(ctx context.Context, sessionID, userID uuid.UUID) error {
	args := m.Called(ctx, sessionID, userID)
	return args.Error(0)
}

func (m *MockEditRepository) UpdateParticipantStatus(ctx context.Context, sessionID, userID uuid.UUID, status string) error {
	args := m.Called(ctx, sessionID, userID, status)
	return args.Error(0)
}

func (m *MockEditRepository) GetCollaborationParticipants(ctx context.Context, sessionID uuid.UUID) ([]*models.CollaborationParticipant, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).([]*models.CollaborationParticipant), args.Error(1)
}

func (m *MockEditRepository) GetParticipantBySocketID(ctx context.Context, socketID string) (*models.CollaborationParticipant, error) {
	args := m.Called(ctx, socketID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CollaborationParticipant), args.Error(1)
}

func (m *MockEditRepository) CreateDocumentVersion(ctx context.Context, version *models.DocumentVersion) error {
	args := m.Called(ctx, version)
	return args.Error(0)
}

func (m *MockEditRepository) GetDocumentVersions(ctx context.Context, documentID uuid.UUID, page, limit int) ([]*models.DocumentVersion, int64, error) {
	args := m.Called(ctx, documentID, page, limit)
	return args.Get(0).([]*models.DocumentVersion), args.Get(1).(int64), args.Error(2)
}

func (m *MockEditRepository) GetDocumentVersion(ctx context.Context, versionID uuid.UUID) (*models.DocumentVersion, error) {
	args := m.Called(ctx, versionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DocumentVersion), args.Error(1)
}

func (m *MockEditRepository) GetLatestVersion(ctx context.Context, documentID uuid.UUID) (*models.DocumentVersion, error) {
	args := m.Called(ctx, documentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DocumentVersion), args.Error(1)
}

func (m *MockEditRepository) CompareVersions(ctx context.Context, fromID, toID uuid.UUID) (*models.DocumentVersion, *models.DocumentVersion, error) {
	args := m.Called(ctx, fromID, toID)
	return args.Get(0).(*models.DocumentVersion), args.Get(1).(*models.DocumentVersion), args.Error(2)
}

func (m *MockEditRepository) DeleteOldVersions(ctx context.Context, documentID uuid.UUID, keepCount int) error {
	args := m.Called(ctx, documentID, keepCount)
	return args.Error(0)
}

func (m *MockEditRepository) CreateEditPermission(ctx context.Context, permission *models.EditPermission) error {
	args := m.Called(ctx, permission)
	return args.Error(0)
}

func (m *MockEditRepository) GetEditPermission(ctx context.Context, documentID, userID uuid.UUID) (*models.EditPermission, error) {
	args := m.Called(ctx, documentID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EditPermission), args.Error(1)
}

func (m *MockEditRepository) UpdateEditPermission(ctx context.Context, permission *models.EditPermission) error {
	args := m.Called(ctx, permission)
	return args.Error(0)
}

func (m *MockEditRepository) DeleteEditPermission(ctx context.Context, documentID, userID uuid.UUID) error {
	args := m.Called(ctx, documentID, userID)
	return args.Error(0)
}

func (m *MockEditRepository) GetDocumentPermissions(ctx context.Context, documentID uuid.UUID) ([]*models.EditPermission, error) {
	args := m.Called(ctx, documentID)
	return args.Get(0).([]*models.EditPermission), args.Error(1)
}

func (m *MockEditRepository) GetUserDocuments(ctx context.Context, userID uuid.UUID) ([]*models.EditPermission, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.EditPermission), args.Error(1)
}

func (m *MockEditRepository) GetEditorConfig(ctx context.Context, editorType string) (*models.EditorConfig, error) {
	args := m.Called(ctx, editorType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EditorConfig), args.Error(1)
}

func (m *MockEditRepository) CreateEditorConfig(ctx context.Context, config *models.EditorConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockEditRepository) UpdateEditorConfig(ctx context.Context, config *models.EditorConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockEditRepository) DeleteEditorConfig(ctx context.Context, editorType string) error {
	args := m.Called(ctx, editorType)
	return args.Error(0)
}

func (m *MockEditRepository) GetAllEditorConfigs(ctx context.Context) ([]*models.EditorConfig, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.EditorConfig), args.Error(1)
}

func (m *MockDocumentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Document, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Document), args.Error(1)
}

func (m *MockDocumentRepository) Create(ctx context.Context, document *models.Document) error {
	args := m.Called(ctx, document)
	return args.Error(0)
}

func (m *MockDocumentRepository) Update(ctx context.Context, document *models.Document) error {
	args := m.Called(ctx, document)
	return args.Error(0)
}

func (m *MockDocumentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockStorageService) Put(ctx context.Context, path string, data []byte) error {
	args := m.Called(ctx, path, data)
	return args.Error(0)
}

func (m *MockStorageService) Get(ctx context.Context, path string) ([]byte, error) {
	args := m.Called(ctx, path)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockStorageService) Delete(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

func (m *MockStorageService) UploadFile(ctx context.Context, req *UploadFileRequest) (*FileResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*FileResponse), args.Error(1)
}

func (m *MockStorageService) DownloadFile(ctx context.Context, fileID string) ([]byte, *FileMetadata, error) {
	args := m.Called(ctx, fileID)
	return args.Get(0).([]byte), args.Get(1).(*FileMetadata), args.Error(2)
}

func (m *MockNotificationService) SendNotification(ctx context.Context, req *SendNotificationRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockNotificationService) Send(ctx context.Context, notification *NotificationRequest) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

// setupEditingService 设置编辑服务测试
func setupEditingService() (*editingService, *MockEditRepository, *MockDocumentRepository, *MockUserRepository, *MockStorageService, *MockNotificationService) {
	mockEditRepo := new(MockEditRepository)
	mockDocRepo := new(MockDocumentRepository)
	mockUserRepo := new(MockUserRepository)
	mockStorage := new(MockStorageService)
	mockNotify := new(MockNotificationService)

	logger := &logrus.Logger{}
	logger.SetLevel(logrus.ErrorLevel)

	service := &editingService{
		editRepo: mockEditRepo,
		docRepo:  mockDocRepo,
		userRepo: mockUserRepo,
		storage:  mockStorage,
		notify:   mockNotify,
		logger:   logger,
	}

	return service, mockEditRepo, mockDocRepo, mockUserRepo, mockStorage, mockNotify
}

// createTestDocument 创建测试文档
func createTestDocument() *models.Document {
	return &models.Document{
		ID:        uuid.New(),
		TenantID:  "test-tenant",
		Title:     "测试文档",
		CreatedBy: 1,
		OwnerID:   1,
		Status:    "published",
	}
}

// createTestUser 创建测试用户
func createTestUser() *models.User {
	return &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Avatar:   "https://example.com/avatar.jpg",
	}
}

// createTestEditSession 创建测试编辑会话
func createTestEditSession() *models.EditSession {
	return &models.EditSession{
		ID:           uuid.New(),
		DocumentID:   uuid.New(),
		UserID:       uuid.New(),
		EditorType:   "rich-text",
		SessionToken: uuid.New().String(),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
	}
}

// createTestEditPermission 创建测试权限
func createTestEditPermission() *models.EditPermission {
	return &models.EditPermission{
		DocumentID:  uuid.New(),
		UserID:      uuid.New(),
		Role:        "editor",
		Permissions: []string{"read", "write"},
		IsActive:    true,
		GrantedAt:   time.Now(),
	}
}

// TestEditingService_CreateSession 测试创建编辑会话
func TestEditingService_CreateSession(t *testing.T) {
	service, mockEditRepo, mockDocRepo, _, _, _ := setupEditingService()

	document := createTestDocument()
	permission := createTestEditPermission()
	session := createTestEditSession()

	ctx := context.Background()

	req := &CreateSessionRequest{
		DocumentID: document.ID.String(),
		UserID:     permission.UserID.String(),
		EditorType: "rich-text",
		Permissions: []string{"read", "write"},
	}

	// 设置模拟返回
	mockDocRepo.On("GetByID", ctx, document.ID).Return(document, nil)
	mockEditRepo.On("GetEditPermission", ctx, document.ID, permission.UserID).Return(permission, nil)
	mockEditRepo.On("CreateEditSession", ctx, mock.AnythingOfType("*models.EditSession")).Return(nil)
	mockEditRepo.On("GetCollaborationSession", ctx, document.ID).Return(nil, fmt.Errorf("not found"))
	mockEditRepo.On("GetEditorConfig", ctx, "rich-text").Return(nil, fmt.Errorf("not found"))

	// 执行测试
	response, err := service.CreateSession(ctx, req)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, document.ID.String(), response.DocumentID)
	assert.Equal(t, permission.UserID.String(), response.UserID)
	assert.Equal(t, "rich-text", response.EditorType)
	assert.NotEmpty(t, response.SessionToken)
	assert.NotNil(t, response.EditorConfig)

	// 验证模拟调用
	mockDocRepo.AssertExpectations(t)
	mockEditRepo.AssertExpectations(t)
}

// TestEditingService_GetSession 测试获取编辑会话
func TestEditingService_GetSession(t *testing.T) {
	service, mockEditRepo, _, _, _, _ := setupEditingService()

	session := createTestEditSession()
	permission := createTestEditPermission()

	ctx := context.Background()

	// 设置模拟返回
	mockEditRepo.On("GetEditSession", ctx, session.ID).Return(session, nil)
	mockEditRepo.On("GetEditPermission", ctx, session.DocumentID, session.UserID).Return(permission, nil)
	mockEditRepo.On("GetCollaborationSession", ctx, session.DocumentID).Return(nil, fmt.Errorf("not found"))
	mockEditRepo.On("GetEditorConfig", ctx, session.EditorType).Return(nil, fmt.Errorf("not found"))

	// 执行测试
	response, err := service.GetSession(ctx, session.ID.String())

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, session.ID.String(), response.SessionID)
	assert.Equal(t, session.DocumentID.String(), response.DocumentID)
	assert.Equal(t, session.UserID.String(), response.UserID)
	assert.Equal(t, session.EditorType, response.EditorType)

	// 验证模拟调用
	mockEditRepo.AssertExpectations(t)
}

// TestEditingService_GetSessionExpired 测试获取过期会话
func TestEditingService_GetSessionExpired(t *testing.T) {
	service, mockEditRepo, _, _, _, _ := setupEditingService()

	session := createTestEditSession()
	session.ExpiresAt = time.Now().Add(-1 * time.Hour) // 过期会话

	ctx := context.Background()

	// 设置模拟返回
	mockEditRepo.On("GetEditSession", ctx, session.ID).Return(session, nil)

	// 执行测试
	response, err := service.GetSession(ctx, session.ID.String())

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "编辑会话已过期")

	// 验证模拟调用
	mockEditRepo.AssertExpectations(t)
}

// TestEditingService_UpdateSessionStatus 测试更新会话状态
func TestEditingService_UpdateSessionStatus(t *testing.T) {
	service, mockEditRepo, _, _, _, _ := setupEditingService()

	session := createTestEditSession()

	ctx := context.Background()

	status := &SessionStatus{
		CursorPosition: &models.Position{
			Line:   1,
			Column: 10,
		},
		SelectionRange: &models.Range{
			Start: models.Position{Line: 1, Column: 5},
			End:   models.Position{Line: 1, Column: 15},
		},
		Activity: "editing",
	}

	// 设置模拟返回
	mockEditRepo.On("GetEditSession", ctx, session.ID).Return(session, nil)
	mockEditRepo.On("UpdateEditSession", ctx, mock.AnythingOfType("*models.EditSession")).Return(nil)

	// 执行测试
	err := service.UpdateSessionStatus(ctx, session.ID.String(), status)

	// 验证结果
	assert.NoError(t, err)

	// 验证模拟调用
	mockEditRepo.AssertExpectations(t)
}

// TestEditingService_CloseSession 测试关闭编辑会话
func TestEditingService_CloseSession(t *testing.T) {
	service, mockEditRepo, _, _, _, _ := setupEditingService()

	session := createTestEditSession()
	collabSession := &models.CollaborationSession{
		ID:         uuid.New(),
		DocumentID: session.DocumentID,
		RoomName:   "doc_" + session.DocumentID.String(),
	}

	ctx := context.Background()

	// 设置模拟返回
	mockEditRepo.On("GetEditSession", ctx, session.ID).Return(session, nil)
	mockEditRepo.On("GetCollaborationSession", ctx, session.DocumentID).Return(collabSession, nil)
	mockEditRepo.On("RemoveCollaborationParticipant", ctx, collabSession.ID, session.UserID).Return(nil)
	mockEditRepo.On("UpdateCollaborationSession", ctx, mock.AnythingOfType("*models.CollaborationSession")).Return(nil)
	mockEditRepo.On("DeleteEditSession", ctx, session.ID).Return(nil)

	// 执行测试
	err := service.CloseSession(ctx, session.ID.String())

	// 验证结果
	assert.NoError(t, err)

	// 验证模拟调用
	mockEditRepo.AssertExpectations(t)
}

// TestEditingService_HandleEditOperation 测试处理编辑操作
func TestEditingService_HandleEditOperation(t *testing.T) {
	service, mockEditRepo, _, _, _, _ := setupEditingService()

	session := createTestEditSession()
	permission := createTestEditPermission()
	permission.Permissions = []string{"read", "write"}

	ctx := context.Background()

	req := &EditOperationRequest{
		SessionID: session.ID.String(),
		OperationType: "insert",
		OperationData: &models.OperationData{
			Type:     "insert",
			Position: 10,
			Content:  "测试内容",
			Origin:   "user",
			Author:   "testuser",
		},
		YjsState: map[string]uint64{
			"client": 1,
		},
	}

	// 设置模拟返回
	mockEditRepo.On("GetEditSession", ctx, session.ID).Return(session, nil)
	mockEditRepo.On("GetEditPermission", ctx, session.DocumentID, session.UserID).Return(permission, nil)
	mockEditRepo.On("CreateEditOperation", ctx, mock.AnythingOfType("*models.EditOperation")).Return(nil)
	mockEditRepo.On("GetCollaborationSession", ctx, session.DocumentID).Return(nil, fmt.Errorf("not found"))

	// 执行测试
	err := service.HandleEditOperation(ctx, req)

	// 验证结果
	assert.NoError(t, err)

	// 验证模拟调用
	mockEditRepo.AssertExpectations(t)
}

// TestEditingService_HandleEditOperationNoPermission 测试处理编辑操作无权限
func TestEditingService_HandleEditOperationNoPermission(t *testing.T) {
	service, mockEditRepo, _, _, _, _ := setupEditingService()

	session := createTestEditSession()
	permission := createTestEditPermission()
	permission.Permissions = []string{"read"} // 只有读取权限

	ctx := context.Background()

	req := &EditOperationRequest{
		SessionID: session.ID.String(),
		OperationType: "insert",
		OperationData: &models.OperationData{
			Type:     "insert",
			Position: 10,
			Content:  "测试内容",
		},
	}

	// 设置模拟返回
	mockEditRepo.On("GetEditSession", ctx, session.ID).Return(session, nil)
	mockEditRepo.On("GetEditPermission", ctx, session.DocumentID, session.UserID).Return(permission, nil)

	// 执行测试
	err := service.HandleEditOperation(ctx, req)

	// 验证结果
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "没有编辑权限")

	// 验证模拟调用
	mockEditRepo.AssertExpectations(t)
}

// TestEditingService_JoinCollaboration 测试加入协作
func TestEditingService_JoinCollaboration(t *testing.T) {
	service, mockEditRepo, mockDocRepo, _, _, _ := setupEditingService()

	document := createTestDocument()
	session := createTestEditSession()
	permission := createTestEditPermission()
	permission.Permissions = []string{"read", "write"}

	ctx := context.Background()

	req := &JoinCollaborationRequest{
		DocumentID: document.ID.String(),
		SessionID:  session.ID.String(),
		SocketID:   "socket-123",
		UserInfo: &CollaborationUserInfo{
			Name:        "测试用户",
			Avatar:      "https://example.com/avatar.jpg",
			Color:       "#FF5722",
			Status:      "online",
			Permissions: []string{"read", "write"},
		},
	}

	// 设置模拟返回
	mockEditRepo.On("GetEditSession", ctx, session.ID).Return(session, nil)
	mockEditRepo.On("GetEditPermission", ctx, document.ID, session.UserID).Return(permission, nil)
	mockDocRepo.On("GetByID", ctx, document.ID).Return(document, nil)
	mockEditRepo.On("GetCollaborationSession", ctx, document.ID).Return(nil, fmt.Errorf("not found"))
	mockEditRepo.On("CreateCollaborationSession", ctx, mock.AnythingOfType("*models.CollaborationSession")).Return(nil)
	mockEditRepo.On("AddCollaborationParticipant", ctx, mock.AnythingOfType("*models.CollaborationParticipant")).Return(nil)
	mockEditRepo.On("GetCollaborationParticipants", ctx, mock.AnythingOfType("uuid.UUID")).Return([]*models.CollaborationParticipant{}, nil)

	// 执行测试
	response, err := service.JoinCollaboration(ctx, req)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.RoomName)
	assert.Equal(t, req.UserInfo, response.UserInfo)

	// 验证模拟调用
	mockEditRepo.AssertExpectations(t)
	mockDocRepo.AssertExpectations(t)
}

// TestEditingService_CreateVersion 测试创建文档版本
func TestEditingService_CreateVersion(t *testing.T) {
	service, mockEditRepo, mockDocRepo, _, mockStorage, _ := setupEditingService()

	document := createTestDocument()

	ctx := context.Background()

	req := &CreateVersionRequest{
		DocumentID:     document.ID.String(),
		Title:          "测试版本",
		EditSummary:    "版本说明",
		IsMajorVersion: false,
		IsPublished:    true,
		ContentDelta: &models.DeltaData{
			Ops:      []interface{}{"insert", "测试内容"},
			ClientID: 1,
			Clock:    1,
			Origin:   "user",
			State:    map[string]uint64{},
		},
		SnapshotData: []byte("snapshot data"),
	}

	// 设置模拟返回
	mockDocRepo.On("GetByID", ctx, document.ID).Return(document, nil)
	mockEditRepo.On("GetLatestVersion", ctx, document.ID).Return(nil, fmt.Errorf("not found"))
	mockStorage.On("Put", ctx, mock.AnythingOfType("string"), req.SnapshotData).Return(nil)
	mockEditRepo.On("CreateDocumentVersion", ctx, mock.AnythingOfType("*models.DocumentVersion")).Return(nil)

	// 执行测试
	response, err := service.CreateVersion(ctx, req)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, req.DocumentID, response.DocumentID)
	assert.Equal(t, 1, response.VersionNumber) // 第一个版本
	assert.Equal(t, req.Title, response.Title)
	assert.Equal(t, req.EditSummary, response.EditSummary)
	assert.Equal(t, req.IsMajorVersion, response.IsMajorVersion)
	assert.Equal(t, req.IsPublished, response.IsPublished)

	// 验证模拟调用
	mockDocRepo.AssertExpectations(t)
	mockEditRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

// TestEditingService_GetEditorConfig 测试获取编辑器配置
func TestEditingService_GetEditorConfig(t *testing.T) {
	service, mockEditRepo, _, _, _, _ := setupEditingService()

	ctx := context.Background()
	editorType := "rich-text"

	// 设置模拟返回 - 配置不存在
	mockEditRepo.On("GetEditorConfig", ctx, editorType).Return(nil, fmt.Errorf("not found"))

	// 执行测试
	response, err := service.GetEditorConfig(ctx, editorType)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, editorType, response.EditorType)
	assert.Equal(t, "default", response.Theme)
	assert.Equal(t, "monospace", response.FontFamily)
	assert.Equal(t, 14, response.FontSize)
	assert.Equal(t, 4, response.TabSize)
	assert.True(t, response.WordWrap)
	assert.True(t, response.LineNumbers)

	// 验证模拟调用
	mockEditRepo.AssertExpectations(t)
}

// TestEditingService_GetEditorConfigExisting 测试获取已存在的编辑器配置
func TestEditingService_GetEditorConfigExisting(t *testing.T) {
	service, mockEditRepo, _, _, _, _ := setupEditingService()

	ctx := context.Background()
	editorType := "rich-text"

	config := &models.EditorConfig{
		EditorType:  editorType,
		Theme:       "dark",
		FontFamily:  "Arial",
		FontSize:    16,
		TabSize:     2,
		WordWrap:    false,
		LineNumbers: false,
		Settings:    map[string]interface{}{"key": "value"},
	}

	// 设置模拟返回
	mockEditRepo.On("GetEditorConfig", ctx, editorType).Return(config, nil)

	// 执行测试
	response, err := service.GetEditorConfig(ctx, editorType)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, editorType, response.EditorType)
	assert.Equal(t, config.Theme, response.Theme)
	assert.Equal(t, config.FontFamily, response.FontFamily)
	assert.Equal(t, config.FontSize, response.FontSize)
	assert.Equal(t, config.TabSize, response.TabSize)
	assert.Equal(t, config.WordWrap, response.WordWrap)
	assert.Equal(t, config.LineNumbers, response.LineNumbers)

	// 验证模拟调用
	mockEditRepo.AssertExpectations(t)
}

// TestEditingService_UpdateEditorConfig 测试更新编辑器配置
func TestEditingService_UpdateEditorConfig(t *testing.T) {
	service, mockEditRepo, _, _, _, _ := setupEditingService()

	ctx := context.Background()

	req := &UpdateEditorConfigRequest{
		EditorType:  "rich-text",
		Theme:       "dark",
		FontFamily:  "Arial",
		FontSize:    16,
		TabSize:     2,
		WordWrap:    false,
		LineNumbers: false,
		Settings:    map[string]interface{}{"key": "value"},
	}

	// 设置模拟返回 - 配置不存在，需要创建
	mockEditRepo.On("GetEditorConfig", ctx, req.EditorType).Return(nil, fmt.Errorf("not found"))
	mockEditRepo.On("CreateEditorConfig", ctx, mock.AnythingOfType("*models.EditorConfig")).Return(nil)

	// 执行测试
	err := service.UpdateEditorConfig(ctx, req)

	// 验证结果
	assert.NoError(t, err)

	// 验证模拟调用
	mockEditRepo.AssertExpectations(t)
}

// TestEditingService_UpdateEditorConfigExisting 测试更新已存在的编辑器配置
func TestEditingService_UpdateEditorConfigExisting(t *testing.T) {
	service, mockEditRepo, _, _, _, _ := setupEditingService()

	ctx := context.Background()

	req := &UpdateEditorConfigRequest{
		EditorType:  "rich-text",
		Theme:       "dark",
		FontFamily:  "Arial",
		FontSize:    16,
		TabSize:     2,
		WordWrap:    false,
		LineNumbers: false,
		Settings:    map[string]interface{}{"key": "value"},
	}

	config := &models.EditorConfig{
		ID:         uuid.New(),
		EditorType: req.EditorType,
	}

	// 设置模拟返回 - 配置已存在，需要更新
	mockEditRepo.On("GetEditorConfig", ctx, req.EditorType).Return(config, nil)
	mockEditRepo.On("UpdateEditorConfig", ctx, mock.AnythingOfType("*models.EditorConfig")).Return(nil)

	// 执行测试
	err := service.UpdateEditorConfig(ctx, req)

	// 验证结果
	assert.NoError(t, err)

	// 验证模拟调用
	mockEditRepo.AssertExpectations(t)
}