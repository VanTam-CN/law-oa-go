package editing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStorageService 模拟存储服务
type MockStorageService struct {
	mock.Mock
}

func (m *MockStorageService) StoreDocument(ctx context.Context, doc *models.Document) error {
	args := m.Called(ctx, doc)
	return args.Error(0)
}

func (m *MockStorageService) GetDocument(ctx context.Context, documentID string) (*models.Document, error) {
	args := m.Called(ctx, documentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Document), args.Error(1)
}

func (m *MockStorageService) UpdateDocument(ctx context.Context, doc *models.Document) error {
	args := m.Called(ctx, doc)
	return args.Error(0)
}

func (m *MockStorageService) DeleteDocument(ctx context.Context, documentID string) error {
	args := m.Called(ctx, documentID)
	return args.Error(0)
}

// MockNotificationService 模拟通知服务
type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) SendNotification(ctx context.Context, notification *models.Notification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func (m *MockNotificationService) GetNotifications(ctx context.Context, userID string) ([]*models.Notification, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Notification), args.Error(1)
}

func TestNewCollaborativeEditingService(t *testing.T) {
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	service := NewCollaborativeEditingService(storageService, notifyService)

	assert.NotNil(t, service)
}

func TestCollaborativeEditingService_CreateCollaborativeSession(t *testing.T) {
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	service := NewCollaborativeEditingService(storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"
	userID := "user-1"

	session, err := service.CreateCollaborativeSession(ctx, sessionID, userID)

	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, sessionID, session.ID)
	assert.Equal(t, userID, session.Users[0].ID)
	assert.True(t, session.IsActive)
	assert.NotNil(t, session.Settings)
}

func TestCollaborativeEditingService_CreateCollaborativeSession_InvalidParams(t *testing.T) {
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	service := NewCollaborativeEditingService(storageService, notifyService)
	ctx := context.Background()

	// 测试空会话ID
	session, err := service.CreateCollaborativeSession(ctx, "", "user-1")

	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "会话ID和用户ID不能为空")

	// 测试空用户ID
	session, err = service.CreateCollaborativeSession(ctx, "test-session-1", "")

	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "会话ID和用户ID不能为空")
}

func TestCollaborativeEditingService_CreateCollaborativeSession_SessionExists(t *testing.T) {
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	service := NewCollaborativeEditingService(storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"
	userID := "user-1"

	// 创建第一个会话
	session1, err := service.CreateCollaborativeSession(ctx, sessionID, userID)
	assert.NoError(t, err)

	// 尝试创建相同会话
	session2, err := service.CreateCollaborativeSession(ctx, sessionID, userID)

	assert.Error(t, err)
	assert.Nil(t, session2)
	assert.Contains(t, err.Error(), "协作会话已存在")
	assert.Equal(t, session1.ID, session1.ID)
}

func TestCollaborativeEditingService_JoinCollaborativeSession(t *testing.T) {
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	service := NewCollaborativeEditingService(storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"
	userID := "user-1"
	userName := "Test User"

	// 创建会话
	session, err := service.CreateCollaborativeSession(ctx, sessionID, userID)
	assert.NoError(t, err)

	// 用户加入会话
	updatedSession, err := service.JoinCollaborativeSession(ctx, sessionID, userID, userName)

	assert.NoError(t, err)
	assert.NotNil(t, updatedSession)
	assert.Equal(t, 1, len(updatedSession.Users))
	assert.Equal(t, userID, updatedSession.Users[0].ID)
	assert.Equal(t, userName, updatedSession.Users[0].Name)
	assert.True(t, updatedSession.Users[0].IsActive)
}

func TestCollaborativeEditingService_JoinCollaborativeSession_NewUser(t *testing.T) {
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	service := NewCollaborativeEditingService(storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"
	user1ID := "user-1"
	user2ID := "user-2"
	user2Name := "Test User 2"

	// 创建会话
	session, err := service.CreateCollaborativeSession(ctx, sessionID, user1ID)
	assert.NoError(t, err)

	// 第一个用户加入
	_, err = service.JoinCollaborativeSession(ctx, sessionID, user1ID, "User 1")
	assert.NoError(t, err)

	// 第二个用户加入
	updatedSession, err := service.JoinCollaborativeSession(ctx, sessionID, user2ID, user2Name)

	assert.NoError(t, err)
	assert.NotNil(t, updatedSession)
	assert.Equal(t, 2, len(updatedSession.Users))
	assert.Equal(t, user2ID, updatedSession.Users[1].ID)
	assert.Equal(t, user2Name, updatedSession.Users[1].Name)
}

func TestCollaborativeService_JoinCollaborativeSession_SessionNotFound(t *testing.T) {
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	service := NewCollaborativeEditingService(storageService, notifyService)
	ctx := context.Background()
	sessionID := "nonexistent-session"
	userID := "user-1"
	userName := "Test User"

	session, err := service.JoinCollaborativeSession(ctx, sessionID, userID, userName)

	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "协作会话不存在")
}

func TestCollaborativeEditingService_LeaveCollaborativeSession(t *testing.T) {
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	service := NewCollaborativeEditingService(storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"
	userID := "user-1"
	userName := "Test User"

	// 创建会话
	session, err := service.CreateCollaborativeSession(ctx, sessionID, userID)
	assert.NoError(t, err)

	// 用户加入会话
	_, err = service.JoinCollaborativeSession(ctx, sessionID, userID, userName)
	assert.NoError(t, err)

	// 验证用户在会话中
	assert.Equal(t, 1, len(session.Users))

	// 用户离开会话
	err = service.LeaveCollaborativeSession(ctx, sessionID, userID)
	assert.NoError(t, err)

	// 验证用户已从会话中移除
	assert.Equal(t, 0, len(session.Users))
}

func TestCollaborativeEditingService_LeaveCollaborativeSession_SessionNotFound(t *testing.T) {
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	service := NewCollaborativeEditingService(storageService, notifyService)
	ctx := context.Background()
	sessionID := "nonexistent-session"
	userID := "user-1"

	err := service.LeaveCollaborativeSession(ctx, sessionID, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "协作会话不存在")
}

func TestCollaborativeEditingService_ApplyOperation(t *testing.T) {
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	service := NewCollaborativeEditingService(storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"
	userID := "user-1"

	// 创建会话
	session, err := service.CreateCollaborativeSession(ctx, sessionID, userID)
	assert.NoError(t, err)

	// 创建插入操作
	operation := &CRDTOperation{
		Type:     "insert",
		Position: 0,
		Text:     "package main\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}",
	}

	// 应用操作
	err = service.ApplyOperation(ctx, sessionID, userID, operation)

	assert.NoError(t, err)

	// 验证文档已更新
	docState, err := service.GetDocumentState(ctx, sessionID)
	assert.NoError(t, err)
	assert.Equal(t, operation.Text, docState.Content)
}

func TestCollaborativeEditingService_ApplyOperation_Insert(t *testing.T) {
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	service := NewCollaborativeEditingService(storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"
	userID := "user-1"

	// 创建会话
	session, err := service.CreateCollaborativeSession(ctx, sessionID, userID)
	assert.NoError(t, err)

	// 初始内容为空
	docState, err := service.GetDocumentState(ctx, sessionID)
	assert.NoError(t, err)
	assert.Empty(t, docState.Content)

	// 在位置0插入文本
	operation := &CRDTOperation{
		Type:     "insert",
		Position: 0,
		Text:     "Hello",
	}

	err = service.ApplyOperation(ctx, sessionID, userID, operation)
	assert.NoError(t, err)

	// 验证内容
	docState, err = service.GetDocumentState(ctx, sessionID)
	assert.NoError(t, err)
	assert.Equal(t, "Hello", docState.Content)
}

func TestCollaborativeEditingService_ApplyOperation_Delete(t *testing.T) {
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	service := NewCollaborativeEditingService(storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"
	userID := "user-1"

	// 创建会话
	session, err := service.CreateCollaborativeSession(ctx, sessionID, userID)
	assert.NoError(t, err)

	// 先插入文本
	insertOperation := &CRDTOperation{
		Type:     "insert",
		Position: 0,
		Text:     "Hello World",
	}
	err = service.ApplyOperation(ctx, sessionID, userID, insertOperation)
	assert.NoError(t, err)

	// 删除部分文本
	deleteOperation := &CRDTOperation{
		Type:     "delete",
		Position: 6,
		Length:   5,
	}

	err = service.ApplyOperation(ctx, sessionID, userID, deleteOperation)
	assert.NoError(t, err)

	// 验证内容
	docState, err := service.GetDocumentState(ctx, sessionID)
	assert.NoError(t, err)
	assert.Equal(t, "Hello ", docState.Content)
}

func TestCollaborativeEditingService_ApplyOperation_InvalidParams(t *testing.T) {
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	service := NewCollaborativeEditingService(storageService, notifyService)
	ctx := context.Background()

	// 测试空会话ID
	operation := &CRDTOperation{
		Type:     "insert",
		Position: 0,
		Text:     "test",
	}
	err := service.ApplyOperation(ctx, "", "user-1", operation)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "会话ID、用户ID和操作不能为空")

	// 测试空用户ID
	err = service.ApplyOperation(ctx, "session-id", "", operation)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "会话ID、用户ID和操作不能为空")

	// 测试空操作
	err = service.ApplyOperation(ctx, "session-id", "user-1", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "会话ID、用户ID和操作不能为空")
}

func TestCollaborativeEditingService_ApplyOperation_SessionNotFound(t *testing.T) {
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	service := NewCollaborativeEditingService(storageService, notifyService)
	ctx := context.Background()
	sessionID := "nonexistent-session"
	userID := "user-1"

	operation := &CRDTOperation{
		Type:     "insert",
		Position: 0,
		Text:     "test",
	}

	err := service.ApplyOperation(ctx, sessionID, userID, operation)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "协作会话不存在或已关闭")
}

func TestCollaborativeEditingService_GetDocumentState(t *testing.T) {
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	service := NewCollaborativeEditingService(storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"
	userID := "user-1"

	// 创建会话
	session, err := service.CreateCollaborativeSession(ctx, sessionID, userID)
	assert.NoError(t, err)

	// 获取文档状态
	state, err := service.GetDocumentState(ctx, sessionID)

	assert.NoError(t, err)
	assert.NotNil(t, state)
	assert.Equal(t, sessionID, state.URI)
	assert.Empty(t, state.Content)
	assert.Equal(t, 1, state.Version)
	assert.Equal(t, 0, len(state.Operations))
	assert.Equal(t, len(session.Users), len(state.Users))
}

func TestCollaborativeEditingService_GetDocumentState_InvalidSessionID(t *testing.T) {
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	service := NewCollaborativeEditingService(storageService, notifyService)
	ctx := context.Background()

	state, err := service.GetDocumentState(ctx, "")

	assert.Error(t, err)
	assert.Nil(t, state)
	assert.Contains(t, err.Error(), "会话ID不能为空")
}

func TestCollaborativeEditingService_GetDocumentState_SessionNotFound(t *testing.T) {
	storageService := &MockStorageService{}
	notifyService &MockNotificationService{}

	service := NewCollaborativeEditingService(storageService, notifyService)
	ctx := context.Background()
	sessionID := "nonexistent-session"

	state, err := service.GetDocumentState(ctx, sessionID)

	assert.Error(t, err)
	assert.Nil(t, state)
	assert.Contains(t, err.Error(), "协作会话不存在或已关闭")
}

func TestCollaborativeEditingService_GetActiveUsers(t *testing.T) {
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	service := NewCollaborativeService(storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"
	userID1 := "user-1"
	userID2 := "user-2"

	// 创建会话
	session, err := service.CreateCollaborativeSession(ctx, sessionID, userID1)
	assert.NoError(t, err)

	// 用户1加入
	_, err = service.JoinCollaborativeSession(ctx, sessionID, userID1, "User 1")
	assert.NoError(t, err)

	// 用户2加入
	_, err = service.JoinCollaborativeSession(ctx, sessionID, userID2, "User 2")
	assert.NoError(t, err)

	// 获取活跃用户
	activeUsers, err := service.GetActiveUsers(ctx, sessionID)

	assert.NoError(t, err)
	assert.NotNil(t, activeUsers)
	assert.Equal(t, 2, len(activeUsers))

	// 验证用户列表
	userIDs := make([]string, len(activeUsers))
	for i, user := range activeUsers {
		userIDs[i] = user.ID
	}
	assert.Contains(t, userIDs, userID1)
	assert.Contains(t, userIDs, userID2)
}

func TestCollaborativeEditingService_GetActiveUsers_NoActiveUsers(t *testing.T) {
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	service := NewCollaborativeEditingService(storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"
	userID := "user-1"

	// 创建会话但用户不加入
	session, err := service.CreateCollaborativeSession(ctx, sessionID, userID)
	assert.NoError(t, err)

	// 获取活跃用户
	activeUsers, err := service.GetActiveUsers(ctx, sessionID)

	assert.NoError(t, err)
	assert.NotNil(t, activeUsers)
	assert.Empty(t, activeUsers)
}

func TestCollaborativeEditingService_GetOperationHistory(t *testing.T) {
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	service := NewCollaborativeEditingService(storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"
	userID := "user-1"

	// 创建会话
	session, err := service.CreateCollaborativeSession(ctx, sessionID, userID)
	assert.NoError(t, err)

	// 应用多个操作
	for i := 0; i < 5; i++ {
		operation := &CRDTOperation{
			Type:     "insert",
			Position: i * 10,
			Text:     fmt.Sprintf("Operation %d", i+1),
		}
		err = service.ApplyOperation(ctx, sessionID, userID, operation)
		assert.NoError(t, err)
	}

	// 获取操作历史
	history, err := service.GetOperationHistory(ctx, sessionID, 10)

	assert.NoError(t, err)
	assert.NotNil(t, history)
	assert.True(t, len(history) > 0)
	assert.Equal(t, 5, len(history))
}

func TestCollaborativeEditingService_GetOperationHistory_WithLimit(t *testing.T) {
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	service := NewCollaborativeEditingService(storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"
	userID := "user-1"

	// 创建会话
	session, err := service.CreateCollaborativeSession(ctx, sessionID, userID)
	assert.NoError(t, err)

	// 应用多个操作
	for i := 0; i < 10; i++ {
		operation := &CRDTOperation{
			Type:     "insert",
			Position: i * 10,
			Text:     fmt.Sprintf("Operation %d", i+1),
		}
		err = service.ApplyOperation(ctx, sessionID, userID, operation)
		assert.NoError(t, err)
	}

	// 获取限制为5的操作历史
	history, err := service.GetOperationHistory(ctx, sessionID, 5)

	assert.NoError(t, err)
	assert.NotNil(t, history)
	assert.Equal(t, 5, len(history))

	// 验证返回的是最近的5个操作
	assert.Equal(t, "Operation 6", history[0].Text)
	assert.Equal(t, "Operation 10", history[4].Text)
}

func TestCollaborativeEditingService_ResolveConflicts(t *testing.T) {
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	service := NewCollaborativeEditingService(storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"
	userID := "user-1"

	// 创建会话
	session, err := service.CreateCollaborativeSession(ctx, sessionID, userID)
	assert.NoError(t, err)

	// 获取初始冲突解决结果
	resolution, err := service.ResolveConflicts(ctx, sessionID)

	assert.NoError(t, err)
	assert.NotNil(t, resolution)
	assert.Equal(t, "no_conflicts", resolution.Resolution)
	assert.Empty(t, resolution.Conflicts)
}

func TestCollaborativeEditingService_SyncDocument(t *testing.T) {
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	service := NewCollaborativeService(storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"
	userID := "user-1"

	// 创建会话
	session, err := service.CreateCollaborativeSession(ctx, sessionID, userID)
	assert.NoError(t, err)

	// 同步文档
	err = service.SyncDocument(ctx, sessionID)

	assert.NoError(t, err)
}

func TestCollaborativeEditingService_DestroyCollaborativeSession(t *testing.T) {
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	service := NewCollaborativeEditingService(storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"
	userID := "user-1"

	// 创建会话
	session, err := service.CreateCollaborativeSession(ctx, sessionID, userID)
	assert.NoError(t)

	assert.True(t, session.IsActive)

	// 销毁会话
	err = service.DestroyCollaborativeSession(ctx, sessionID)

	assert.NoError(t, err)

	// 验证会话已被标记为非活跃
	assert.False(t, session.IsActive)
}

func TestCollaborativeEditingService_DestroyCollaborativeSession_InvalidSessionID(t *testing.T) {
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	service := NewCollaborativeService(storageService, notifyService)
	ctx := context.Background()

	err := service.DestroyCollaborativeSession(ctx, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "会话ID不能为空")
}

// 辅助函数测试

func TestGenerateUserAvatar(t *testing.T) {
	service := &CollaborativeEditingServiceImpl{}

	avatar1 := service.generateUserAvatar("user1")
	avatar2 := service.generateUserAvatar("user2")

	assert.NotEmpty(t, avatar1)
	assert.NotEmpty(t, avatar2)
	assert.Contains(t, avatar1, "user1"[:2])
	assert.Contains(t, avatar2, "user2"[:2])
	assert.NotEqual(t, avatar1, avatar2)
}

func TestGenerateUserColor(t *testing.T) {
	service := &CollaborativeEditingServiceImpl{}

	color1 := service.generateUserColor("user1")
	color2 := service.generateUserColor("user2")
	color3 := service.generateUserColor("user1") // 相同用户ID应返回相同颜色

	assert.NotEmpty(t, color1)
	assert.NotEmpty(t, color2)
	assert.Equal(t, color1, color3)
}

func TestGenerateOperationID(t *testing.T) {
	service := &CollaborativeEditingServiceImpl{}

	id1 := service.generateOperationID()
	id2 := service.generateOperationID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "op_")
	assert.Contains(t, id2, "op_")
}

func TestApplyCRDTOperation_Insert(t *testing.T) {
	service := &CollaborativeEditingServiceImpl{}

	// 测试在开头插入
	content := "World"
	operation := &CRDTOperation{
		Type:     "insert",
		Position: 0,
		Text:     "Hello ",
	}

	result, err := service.applyCRDTOperation(content, operation)

	assert.NoError(t, err)
	assert.Equal(t, "Hello World", result)

	// 测试在中间插入
	content = "Hello World"
	operation = &CRDTOperation{
		Type:     "insert",
		Position: 5,
		Text:     ",",
	}

	result, err = service.applyCRDTOperation(content, operation)

	assert.NoError(t, err)
	assert.Equal(t, "Hello, World", result)

	// 测试在末尾插入
	content = "Hello"
	operation = &CRDTOperation{
		Type:     "insert",
		Position: 5,
		Text:     " World",
	}

	result, err = service.applyCRDTOperation(content, operation)

	assert.NoError(t, err)
	assert.Equal(t, "Hello World", result)

	// 测试位置超出范围
	content = "Hello"
	operation = &CRDTOperation{
		Type:     "insert",
		Position: 10,
		Text:     "!",
	}

	result, err = service.applyCRDTOperation(content, operation)

	assert.NoError(t, err)
	assert.Equal(t, "Hello!", result)
}

func TestApplyCRDTOperation_Delete(t *testing.T) {
	service := &CollaborativeEditingServiceImpl{}

	// 测试删除中间部分
	content := "Hello World"
	operation := &CRDTOperation{
		Type:     "delete",
		Position: 5,
		Length:   6,
	}

	result, err := service.applyCRDTOOperation(content, operation)

	assert.NoError(t, err)
	assert.Equal(t, "Hello", result)

	// 测试删除超出范围
	content = "Hello"
	operation = &CRDTOperation{
		Type:     "delete",
		Position: 10,
		Length:   5,
	}

	result, err = service.applyCRDTOperation(content, operation)

	assert.NoError(t, err)
	assert.Equal(t, "Hello", result)

	// 测试删除空内容
	content = "Hello"
	operation = &CRDTOperation{
		Type:     "delete",
		Position: 0,
		Length:   0,
	}

	result, err = service.applyCRDTOperation(content, operation)

	assert.NoError(t, err)
	assert.Equal(t, "Hello", result)
}

func TestApplyCRDTOperation_Retain(t *testing.T) {
	service := &CollaborativeEditingServiceImpl{}

	content := "Hello World"
	operation := &CRDTOperation{
		Type:     "retain",
		Position: 5,
	}

	result, err := service.applyCRDTOperation(content, operation)

	assert.NoError(t, err)
	assert.Equal(t, "Hello World", result)
}

func TestApplyCRDTOperation_UnsupportedType(t *testing.T) {
	service := &CollaborativeServiceImpl{}

	content := "Hello"
	operation := &CRDTOperation{
		Type: "unsupported",
	}

	result, err := service.applyCRDTOperation(content, operation)

	assert.Error(t, err)
	assert.Equal(t, "Hello", result)
	assert.Contains(t, err.Error(), "不支持的操作类型")
}

func TestGetDefaultCollaborativeSettings(t *testing.T) {
	service := &CollaborativeEditingServiceImpl{}

	settings := service.getDefaultCollaborativeSettings()

	assert.NotNil(t, settings)
	assert.True(t, settings.AutoSave)
	assert.Equal(t, 5*time.Second, settings.AutoSaveDelay)
	assert.Equal(t, "merge", settings.ConflictPolicy)
	assert.Equal(t, 10, settings.MaxUsers)
	assert.True(t, settings.EnableChat)
	assert.True(t, settings.EnableComments)
	assert.True(t, settings.EnableTracking)
	assert.Equal(t, "private", settings.PrivacyMode)
}

func TestNewConflictDetector(t *testing.T) {
	detector := NewConflictDetector()

	assert.NotNil(t, detector)
	assert.NotNil(t, detector.rules)
	assert.True(t, len(detector.rules) > 0)
}

func TestConflictDetector_AddRule(t *testing.T) {
	detector := NewConflictDetector()
	initialRuleCount := len(detector.rules)

	rule := ConflictRule{
		Type:     "test_rule",
		Checker:  func(operations []*CRDTOperation) []*Conflict {
			return []*Conflict{}
		},
		Priority: 1,
		Enabled:  true,
	}

	detector.addRule(rule)

	assert.Equal(t, initialRuleCount+1, len(detector.rules))
}

func TestConflictDetector_DetectConflicts(t *testing.T) {
	detector := NewConflictDetector()

	operations := []*CRDTOperation{
		{
			Type:     "insert",
			Position: 0,
			Text:     "test",
		},
	}

	conflicts := detector.DetectConflicts(operations)

	assert.NotNil(t, conflicts)
	// 由于没有实际的冲突检测逻辑，应该返回空列表
	assert.Empty(t, conflicts)
}

func TestNewSyncManager(t *testing.T) {
	manager := NewSyncManager()

	assert.NotNil(t, manager)
	assert.Equal(t, 1*time.Second, manager.syncInterval)
	assert.NotNil(t, manager.syncBuffer)
}

func TestSyncManager_GetSyncBuffer(t *testing.T) {
	manager := NewSyncManager()

	sessionID := "test-session-1"

	// 初始时没有缓冲区
	buffer, exists := manager.syncBuffer[sessionID]
	assert.False(t, exists)

	// 添加缓冲区后
	manager.syncManager.mutex.Lock()
	manager.syncBuffer[sessionID] = &SyncBuffer{
		SessionID:  sessionID,
		Operations: make([]*CRDTOperation, 0),
		Timestamp:  time.Now(),
		Flushed:    false,
	}
	manager.syncManager.mutex.Unlock()

	buffer, exists = manager.syncBuffer[sessionID]
	assert.True(t, exists)
	assert.Equal(t, sessionID, buffer.SessionID)
	assert.Empty(t, buffer.Operations)
	assert.False(t, buffer.Flushed)
}