package editing

import (
	"context"
	"fmt"
	"testing"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/services"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockEditService 模拟编辑服务
type MockEditService struct {
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

// MockYjsProvider 模拟Yjs提供者
type MockYjsProvider struct {
	mock.Mock
}

// MockYjsDocument 模拟Yjs文档
type MockYjsDocument struct {
	mock.Mock
}

// MockYjsText 模拟Yjs文本
type MockYjsText struct {
	mock.Mock
}

// MockCursorManager 模拟光标管理器
type MockCursorManager struct {
	mock.Mock
}

func (m *MockEditService) GetSession(ctx context.Context, sessionID string) (*services.SessionResponse, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.SessionResponse), args.Error(1)
}

func (m *MockEditService) GetEditorConfig(ctx context.Context, editorType string) (*services.EditorConfigResponse, error) {
	args := m.Called(ctx, editorType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.EditorConfigResponse), args.Error(1)
}

func (m *MockEditService) HandleEditOperation(ctx context.Context, req *services.EditOperationRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockEditService) CreateVersion(ctx context.Context, req *services.CreateVersionRequest) (*services.VersionResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.VersionResponse), args.Error(1)
}

func (m *MockEditService) GetVersions(ctx context.Context, documentID string, page, limit int) (*services.VersionListResponse, error) {
	args := m.Called(ctx, documentID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.VersionListResponse), args.Error(1)
}

func (m *MockEditService) GetCollaborators(ctx context.Context, documentID string) (*services.CollaboratorListResponse, error) {
	args := m.Called(ctx, documentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.CollaboratorListResponse), args.Error(1)
}

func (m *MockEditService) GetDocumentOperations(ctx context.Context, documentID string, page, limit int) (*services.OperationListResponse, error) {
	args := m.Called(ctx, documentID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.OperationListResponse), args.Error(1)
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

func (m *MockStorageService) UploadFile(ctx context.Context, req *services.UploadFileRequest) (*services.FileResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.FileResponse), args.Error(1)
}

func (m *MockStorageService) DownloadFile(ctx context.Context, fileID string) ([]byte, *services.FileMetadata, error) {
	args := m.Called(ctx, fileID)
	return args.Get(0).([]byte), args.Get(1).(*services.FileMetadata), args.Error(2)
}

func (m *MockNotificationService) SendNotification(ctx context.Context, req *services.SendNotificationRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockNotificationService) Send(ctx context.Context, notification *NotificationRequest) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func (m *MockYjsProvider) Initialize(documentID string) (YjsDocument, error) {
	args := m.Called(documentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(YjsDocument), args.Error(1)
}

func (m *MockYjsProvider) GetText(name string) YjsText {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(YjsText)
}

func (m *MockYjsProvider) Subscribe(callback func(delta interface{})) error {
	args := m.Called(callback)
	return args.Error(0)
}

func (m *MockYjsProvider) Unsubscribe() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockYjsProvider) Destroy() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockYjsDocument) GetID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockYjsDocument) GetText(name string) YjsText {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(YjsText)
}

func (m *MockYjsDocument) Destroy() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockYjsText) Insert(index int, text string) {
	m.Called(index, text)
}

func (m *MockYjsText) Delete(index int, length int) {
	m.Called(index, length)
}

func (m *MockYjsText) Format(index int, length int, attributes map[string]interface{}) {
	m.Called(index, length, attributes)
}

func (m *MockYjsText) GetDelta() *models.DeltaData {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.DeltaData)
}

func (m *MockYjsText) GetLength() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockYjsText) Observe(callback func(event interface{})) {
	m.Called(callback)
}

func (m *MockYjsText) Unobserve() {
	m.Called()
}

func (m *MockCursorManager) Initialize(yjsProvider YjsProvider) error {
	args := m.Called(yjsProvider)
	return args.Error(0)
}

func (m *MockCursorManager) UpdateCursor(userID string, cursor *CursorInfo) error {
	args := m.Called(userID, cursor)
	return args.Error(0)
}

func (m *MockCursorManager) RemoveCursor(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockCursorManager) GetCursors() []*CursorInfo {
	args := m.Called()
	return args.Get(0).([]*CursorInfo)
}

func (m *MockCursorManager) Destroy() error {
	args := m.Called()
	return args.Error(0)
}

// setupRichTextEditor 设置富文本编辑器测试
func setupRichTextEditor() (*RichTextEditor, *MockEditService, *MockStorageService, *MockNotificationService, *MockYjsProvider, *MockCursorManager) {
	mockEditService := new(MockEditService)
	mockStorage := new(MockStorageService)
	mockNotify := new(MockNotificationService)
	mockYjsProvider := new(MockYjsProvider)
	mockCursorManager := new(MockCursorManager)

	editor := &RichTextEditor{
		editService:  mockEditService,
		storage:      mockStorage,
		notify:       mockNotify,
		yjsProvider:  mockYjsProvider,
		cursorManager: mockCursorManager,
	}

	return editor, mockEditService, mockStorage, mockNotify, mockYjsProvider, mockCursorManager
}

// createTestSession 创建测试会话
func createTestSession() *services.SessionResponse {
	return &services.SessionResponse{
		SessionID:    uuid.New().String(),
		DocumentID:   uuid.New().String(),
		UserID:       uuid.New().String(),
		EditorType:   "rich-text",
		Permissions:  []string{"read", "write"},
		EditorConfig: &services.EditorConfigResponse{
			EditorType:  "rich-text",
			Theme:       "snow",
			FontFamily:  "Arial",
			FontSize:    14,
			TabSize:     4,
			WordWrap:    true,
			LineNumbers: true,
			Settings:    make(map[string]interface{}),
		},
	}
}

// createTestRichTextContent 创建测试富文本内容
func createTestRichTextContent() *RichTextContent {
	return &RichTextContent{
		Delta: &models.DeltaData{
			Ops: []interface{}{
				map[string]interface{}{
					"insert": "Hello World",
					"attributes": map[string]interface{}{
						"bold": true,
					},
				},
				map[string]interface{}{
					"insert": " ",
				},
				map[string]interface{}{
					"insert": "Rich Text Editor",
					"attributes": map[string]interface{}{
						"italic": true,
						"color":  "#ff0000",
					},
				},
				map[string]interface{}{
					"insert": "\n",
				},
			},
		},
		HTML:       "<strong>Hello World</strong> <em>Rich Text Editor</em><br>",
		PlainText:  "Hello World Rich Text Editor\n",
		Length:     26,
		LastModified: time.Now(),
		Version:    1,
	}
}

// createTestRichTextOperation 创建测试富文本操作
func createTestRichTextOperation() *RichTextOperation {
	return &RichTextOperation{
		Type:      "insert",
		Delta: &models.DeltaData{
			Ops: []interface{}{
				map[string]interface{}{
					"insert": "Test",
				},
			},
		},
		Range: TextRange{
			Index:  10,
			Length: 0,
		},
		Attributes: map[string]interface{}{
			"bold": true,
		},
		Timestamp: time.Now(),
		UserID:    uuid.New().String(),
		SessionID: uuid.New().String(),
		Source:    "user",
	}
}

// createTestCursorInfo 创建测试光标信息
func createTestCursorInfo() *CursorInfo {
	return &CursorInfo{
		UserID:     uuid.New().String(),
		UserName:   "Test User",
		UserAvatar: "https://example.com/avatar.jpg",
		UserColor:  "#2196F3",
		Range: TextRange{
			Index:  5,
			Length: 10,
		},
		Selection: TextRange{
			Index:  5,
			Length: 10,
		},
		LastSeen: time.Now(),
	}
}

// TestRichTextEditor_InitializeEditor 测试初始化编辑器
func TestRichTextEditor_InitializeEditor(t *testing.T) {
	editor, mockEditService, _, _, mockYjsProvider, mockCursorManager := setupRichTextEditor()

	session := createTestSession()
	ctx := context.Background()

	// 设置模拟返回
	mockEditService.On("GetSession", ctx, session.SessionID).Return(session, nil)
	mockEditService.On("GetEditorConfig", ctx, "rich-text").Return(session.EditorConfig, nil)
	mockYjsProvider.On("Initialize", session.DocumentID).Return(nil, nil)
	mockYjsProvider.On("Subscribe", mock.AnythingOfType("func(interface {})")).Return(nil)
	mockCursorManager.On("Initialize", mockYjsProvider).Return(nil)

	// 执行测试
	config, err := editor.InitializeEditor(ctx, session.SessionID)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "snow", config.Theme)
	assert.Equal(t, "开始输入文档内容...", config.Placeholder)
	assert.NotNil(t, config.Toolbar)
	assert.NotNil(t, config.History)
	assert.True(t, config.Collaboration.Enabled)
	assert.True(t, config.Collaboration.YjsEnabled)
	assert.True(t, config.AutoSave.Enabled)

	// 验证模拟调用
	mockEditService.AssertExpectations(t)
	mockYjsProvider.AssertExpectations(t)
	mockCursorManager.AssertExpectations(t)
}

// TestRichTextEditor_InitializeEditorNoPermission 测试初始化编辑器无权限
func TestRichTextEditor_InitializeEditorNoPermission(t *testing.T) {
	editor, mockEditService, _, _, _, _ := setupRichTextEditor()

	session := createTestSession()
	session.Permissions = []string{"read"} // 只有读取权限
	ctx := context.Background()

	// 设置模拟返回
	mockEditService.On("GetSession", ctx, session.SessionID).Return(session, nil)
	mockEditService.On("GetEditorConfig", ctx, "rich-text").Return(session.EditorConfig, nil)

	// 执行测试
	config, err := editor.InitializeEditor(ctx, session.SessionID)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.True(t, config.ReadOnly) // 应该是只读模式

	// 验证模拟调用
	mockEditService.AssertExpectations(t)
}

// TestRichTextEditor_LoadContent 测试加载内容
func TestRichTextEditor_LoadContent(t *testing.T) {
	editor, mockEditService, mockStorage, _, _, _ := setupRichTextEditor()

	session := createTestSession()
	content := createTestRichTextContent()
	versionResponse := &services.VersionResponse{
		VersionID:     uuid.New().String(),
		DocumentID:    session.DocumentID,
		VersionNumber: 1,
		CreatedAt:     time.Now(),
	}

	versionListResponse := &services.VersionListResponse{
		Versions: []*services.VersionResponse{versionResponse},
		Total:    1,
		Page:     1,
		Limit:    1,
	}

	ctx := context.Background()

	// 设置模拟返回
	mockEditService.On("GetSession", ctx, session.SessionID).Return(session, nil)
	mockEditService.On("GetVersions", ctx, session.DocumentID, 1, 1).Return(versionListResponse, nil)
	mockStorage.On("Get", ctx, versionResponse.SnapshotPath).Return([]byte(content.HTML), nil)

	// 执行测试
	loadedContent, err := editor.LoadContent(ctx, session.SessionID)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, loadedContent)
	assert.Equal(t, content.HTML, loadedContent.HTML)
	assert.Equal(t, content.PlainText, loadedContent.PlainText)
	assert.Equal(t, content.Length, loadedContent.Length)

	// 验证模拟调用
	mockEditService.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

// TestRichTextEditor_LoadContentWithYjs 测试从Yjs加载内容
func TestRichTextEditor_LoadContentWithYjs(t *testing.T) {
	editor, mockEditService, _, _, mockYjsProvider, _ := setupRichTextEditor()

	session := createTestSession()
	content := createTestRichTextContent()
	yjsText := new(MockYjsText)

	ctx := context.Background()

	// 设置模拟返回
	mockEditService.On("GetSession", ctx, session.SessionID).Return(session, nil)
	mockEditService.On("GetVersions", ctx, session.DocumentID, 1, 1).Return(nil, fmt.Errorf("no versions"))
	mockYjsProvider.On("GetText", "content").Return(yjsText)
	yjsText.On("GetDelta").Return(content.Delta)
	yjsText.On("GetLength").Return(content.Length)

	// 执行测试
	loadedContent, err := editor.LoadContent(ctx, session.SessionID)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, loadedContent)
	assert.Equal(t, content.Delta, loadedContent.Delta)
	assert.Equal(t, content.HTML, loadedContent.HTML)
	assert.Equal(t, content.Length, loadedContent.Length)

	// 验证模拟调用
	mockEditService.AssertExpectations(t)
	mockYjsProvider.AssertExpectations(t)
	yjsText.AssertExpectations(t)
}

// TestRichTextEditor_SaveContent 测试保存内容
func TestRichTextEditor_SaveContent(t *testing.T) {
	editor, mockEditService, mockStorage, _, _, _ := setupRichTextEditor()

	session := createTestSession()
	content := createTestRichTextContent()
	versionResponse := &services.VersionResponse{
		VersionID:     uuid.New().String(),
		DocumentID:    session.DocumentID,
		VersionNumber: 2,
		CreatedAt:     time.Now(),
	}

	ctx := context.Background()

	// 设置模拟返回
	mockEditService.On("GetSession", ctx, session.SessionID).Return(session, nil)
	mockEditService.On("CreateVersion", ctx, mock.AnythingOfType("*services.CreateVersionRequest")).Return(versionResponse, nil)
	mockStorage.On("Put", ctx, mock.AnythingOfType("string"), []byte(content.HTML)).Return(nil)

	// 执行测试
	err := editor.SaveContent(ctx, session.SessionID, content)

	// 验证结果
	assert.NoError(t, err)

	// 验证模拟调用
	mockEditService.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

// TestRichTextEditor_SaveContentNoPermission 测试保存内容无权限
func TestRichTextEditor_SaveContentNoPermission(t *testing.T) {
	editor, mockEditService, _, _, _, _ := setupRichTextEditor()

	session := createTestSession()
	session.Permissions = []string{"read"} // 只有读取权限
	content := createTestRichTextContent()

	ctx := context.Background()

	// 设置模拟返回
	mockEditService.On("GetSession", ctx, session.SessionID).Return(session, nil)

	// 执行测试
	err := editor.SaveContent(ctx, session.SessionID, content)

	// 验证结果
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "没有编辑权限")

	// 验证模拟调用
	mockEditService.AssertExpectations(t)
}

// TestRichTextEditor_HandleOperation 测试处理操作
func TestRichTextEditor_HandleOperation(t *testing.T) {
	editor, mockEditService, _, _, _, _ := setupRichTextEditor()

	session := createTestSession()
	operation := createTestRichTextOperation()

	ctx := context.Background()

	// 设置模拟返回
	mockEditService.On("GetSession", ctx, operation.SessionID).Return(session, nil)
	mockEditService.On("HandleEditOperation", ctx, mock.AnythingOfType("*services.EditOperationRequest")).Return(nil)

	// 执行测试
	err := editor.HandleOperation(ctx, operation.SessionID, operation)

	// 验证结果
	assert.NoError(t, err)

	// 验证模拟调用
	mockEditService.AssertExpectations(t)
}

// TestRichTextEditor_UpdateCursor 测试更新光标
func TestRichTextEditor_UpdateCursor(t *testing.T) {
	editor, mockEditService, _, _, _, mockCursorManager := setupRichTextEditor()

	session := createTestSession()
	cursor := createTestCursorInfo()

	ctx := context.Background()

	// 设置模拟返回
	mockEditService.On("GetSession", ctx, session.SessionID).Return(session, nil)
	mockCursorManager.On("UpdateCursor", cursor.UserID, cursor).Return(nil)

	// 执行测试
	err := editor.UpdateCursor(ctx, session.SessionID, cursor)

	// 验证结果
	assert.NoError(t, err)

	// 验证模拟调用
	mockEditService.AssertExpectations(t)
	mockCursorManager.AssertExpectations(t)
}

// TestRichTextEditor_UpdateCursorNoPermission 测试更新光标无权限
func TestRichTextEditor_UpdateCursorNoPermission(t *testing.T) {
	editor, mockEditService, _, _, _, _ := setupRichTextEditor()

	session := createTestSession()
	session.Permissions = []string{} // 没有任何权限
	cursor := createTestCursorInfo()

	ctx := context.Background()

	// 设置模拟返回
	mockEditService.On("GetSession", ctx, session.SessionID).Return(session, nil)

	// 执行测试
	err := editor.UpdateCursor(ctx, session.SessionID, cursor)

	// 验证结果
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "没有查看权限")

	// 验证模拟调用
	mockEditService.AssertExpectations(t)
}

// TestRichTextEditor_GetCursors 测试获取光标
func TestRichTextEditor_GetCursors(t *testing.T) {
	editor, mockEditService, _, _, _, mockCursorManager := setupRichTextEditor()

	session := createTestSession()
	cursor1 := createTestCursorInfo()
	cursor2 := createTestCursorInfo()
	cursor2.LastSeen = time.Now().Add(-1 * time.Minute) // 1分钟前

	ctx := context.Background()

	// 设置模拟返回
	mockEditService.On("GetSession", ctx, session.SessionID).Return(session, nil)
	mockCursorManager.On("GetCursors").Return([]*CursorInfo{cursor1, cursor2})

	// 执行测试
	cursors, err := editor.GetCursors(ctx, session.SessionID)

	// 验证结果
	assert.NoError(t, err)
	assert.Len(t, cursors, 2) // 两个光标都应该在5分钟内，所以都有效

	// 验证模拟调用
	mockEditService.AssertExpectations(t)
	mockCursorManager.AssertExpectations(t)
}

// TestRichTextEditor_GetCursorsWithExpired 测试获取光标包含过期光标
func TestRichTextEditor_GetCursorsWithExpired(t *testing.T) {
	editor, mockEditService, _, _, _, mockCursorManager := setupRichTextEditor()

	session := createTestSession()
	cursor1 := createTestCursorInfo()
	cursor2 := createTestCursorInfo()
	cursor2.LastSeen = time.Now().Add(-10 * time.Minute) // 10分钟前，应该被过滤掉

	ctx := context.Background()

	// 设置模拟返回
	mockEditService.On("GetSession", ctx, session.SessionID).Return(session, nil)
	mockCursorManager.On("GetCursors").Return([]*CursorInfo{cursor1, cursor2})

	// 执行测试
	cursors, err := editor.GetCursors(ctx, session.SessionID)

	// 验证结果
	assert.NoError(t, err)
	assert.Len(t, cursors, 1) // 只有一个有效光标
	assert.Equal(t, cursor1.UserID, cursors[0].UserID)

	// 验证模拟调用
	mockEditService.AssertExpectations(t)
	mockCursorManager.AssertExpectations(t)
}

// TestRichTextEditor_ConvertToHTML 测试转换为HTML
func TestRichTextEditor_ConvertToHTML(t *testing.T) {
	editor := &RichTextEditor{}

	sessionID := uuid.New().String()
	delta := &models.DeltaData{
		Ops: []interface{}{
			map[string]interface{}{
				"insert": "Hello",
				"attributes": map[string]interface{}{
					"bold": true,
				},
			},
			map[string]interface{}{
				"insert": " World",
			},
		},
	}

	// 执行测试
	html, err := editor.ConvertToHTML(sessionID, delta)

	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, "<strong>Hello</strong> World", html)
}

// TestRichTextEditor_ConvertToPlainText 测试转换为纯文本
func TestRichTextEditor_ConvertToPlainText(t *testing.T) {
	editor := &RichTextEditor{}

	sessionID := uuid.New().String()
	delta := &models.DeltaData{
	Ops: []interface{}{
			map[string]interface{}{
				"insert": "Hello",
				"attributes": map[string]interface{}{
					"bold": true,
				},
			},
			map[string]interface{}{
				"insert": " World",
			},
		},
	}

	// 执行测试
	text, err := editor.ConvertToPlainText(sessionID, delta)

	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, "Hello World", text)
}

// TestRichTextEditor_DestroyEditor 测试销毁编辑器
func TestRichTextEditor_DestroyEditor(t *testing.T) {
	editor, _, _, _, mockYjsProvider, mockCursorManager := setupRichTextEditor()

	sessionID := uuid.New().String()

	// 设置模拟返回
	mockYjsProvider.On("Unsubscribe").Return(nil)
	mockCursorManager.On("Destroy").Return(nil)
	mockYjsProvider.On("Destroy").Return(nil)

	// 执行测试
	err := editor.DestroyEditor(sessionID)

	// 验证结果
	assert.NoError(t, err)

	// 验证模拟调用
	mockYjsProvider.AssertExpectations(t)
	mockCursorManager.AssertExpectations(t)
}

// TestRichTextEditor_GetEditorStats 测试获取编辑器统计
func TestRichTextEditor_GetEditorStats(t *testing.T) {
	editor, mockEditService, _, _, _, _ := setupRichTextEditor()

	session := createTestSession()
	collaborators := &services.CollaboratorListResponse{
		Participants: []*services.CollaborationParticipantResponse{
			{
				UserID:    uuid.New().String(),
				UserName:  "User 1",
				Status:    "online",
				JoinedAt:   time.Now(),
			},
			{
				UserID:    uuid.New().String(),
				UserName:  "User 2",
				Status:    "away",
				JoinedAt:   time.Now(),
			},
		},
		ActiveUsers: 2,
	}

	operations := &services.OperationListResponse{
		Operations: []*models.EditOperation{},
		Total:      0,
		Page:       1,
		Limit:      100,
	}

	ctx := context.Background()

	// 设置模拟返回
	mockEditService.On("GetSession", ctx, session.SessionID).Return(session, nil)
	mockEditService.On("GetCollaborators", ctx, session.DocumentID).Return(collaborators, nil)
	mockEditService.On("GetDocumentOperations", ctx, session.DocumentID, 1, 100).Return(operations, nil)

	// 执行测试
	stats, err := editor.GetEditorStats(ctx, session.SessionID)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, sessionID, stats["session_id"])
	assert.Equal(t, session.DocumentID, stats["document_id"])
	assert.Equal(t, "rich-text", stats["editor_type"])
	assert.Equal(t, 2, stats["collaborators"])
	assert.Equal(t, 2, stats["active_users"])
	assert.Equal(t, 0, stats["operations"])

	// 验证模拟调用
	mockEditService.AssertExpectations(t)
}

// 辅助测试函数

// TestRichTextEditor_hasEditPermission 测试编辑权限检查
func TestRichTextEditor_hasEditPermission(t *testing.T) {
	editor := &RichTextEditor{}

	tests := []struct {
		name        string
		permissions []string
		expected    bool
	}{
		{
			name:        "有编辑权限",
			permissions: []string{"read", "write"},
			expected:    true,
		},
		{
			name:        "管理员权限",
			permissions: []string{"read", "admin"},
			expected:    true,
		},
		{
			name:        "所有者权限",
			permissions: []string{"read", "owner"},
			expected:    true,
		},
		{
			name:        "只有读取权限",
			permissions: []string{"read"},
			expected:    false,
		},
		{
			name:        "没有权限",
			permissions: []string{},
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := editor.hasEditPermission(tt.permissions)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRichTextEditor_hasViewPermission 测试查看权限检查
func TestRichTextEditor_hasViewPermission(t *testing.T) {
	editor := &RichTextEditor{}

	tests := []struct {
		name        string
		permissions []string
		expected    bool
	}{
		{
			name:        "有读取权限",
			permissions: []string{"read"},
			expected:    true,
		},
		{
			name:        "有编辑权限",
			permissions: []string{"read", "write"},
			expected:    true,
		},
		{
			name:        "管理员权限",
			permissions: []string{"read", "admin"},
			expected:    true,
		},
		{
			name:        "所有者权限",
			permissions: []string{"read", "owner"},
			expected:    true,
		},
		{
			name:        "没有权限",
			permissions: []string{},
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := editor.hasViewPermission(tt.permissions)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRichTextEditor_convertDeltaToHTML 测试Delta转HTML
func TestRichTextEditor_convertDeltaToHTML(t *testing.T) {
	editor := &RichTextEditor{}

	tests := []struct {
		name     string
		delta    *models.DeltaData
		expected string
	}{
		{
			name: "空Delta",
			delta: &models.DeltaData{Ops: []interface{}{}},
			expected: "",
		},
		{
			name: "纯文本",
			delta: &models.DeltaData{
				Ops: []interface{}{
					map[string]interface{}{"insert": "Hello World"},
				},
			},
			expected: "Hello World",
		},
		{
			name: "换行符",
			delta: &models.DeltaData{
				Ops: []interface{}{
					map[string]interface{}{"insert": "Hello"},
					map[string]interface{}{"insert": "\n"},
					map[string]interface{}{"insert": "World"},
				},
			},
			expected: "Hello<br>World",
		},
		{
			name: "粗体文本",
			delta: &models.DeltaData{
				Ops: []interface{}{
					map[string]interface{}{
						"insert": "Bold",
						"attributes": map[string]interface{}{
							"bold": true,
						},
					},
				},
			},
			expected: "<strong>Bold</strong>",
		},
		{
			name: "斜体文本",
			delta: &models.DeltaData{
				Ops: []interface{}{
					map[string]interface{}{
						"insert": "Italic",
						"attributes": map[string]interface{}{
							"italic": true,
						},
					},
				},
			},
			expected: "<em>Italic</em>",
		},
		{
			name: "组合格式",
			delta: &models.DeltaData{
				Ops: []interface{}{
					map[string]interface{}{
						"insert": "Bold",
						"attributes": map[string]interface{}{
							"bold": true,
						},
					},
					map[string]interface{}{
						"insert": " ",
					},
					map[string]interface{}{
						"insert": "Italic",
						"attributes": map[string]interface{}{
							"italic": true,
						},
					},
				},
			},
			expected: "<strong>Bold</strong> <em>Italic</em>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := editor.convertDeltaToHTML(tt.delta)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRichTextEditor_convertDeltaToPlainText 测试Delta转纯文本
func TestRichTextEditor_convertDeltaToPlainText(t *testing.T) {
	editor := &RichTextEditor{}

	tests := []struct {
		name     string
		delta    *models.DeltaData
		expected string
	}{
		{
			name:     "空Delta",
			delta:    &models.DeltaData{Ops: []interface{}{}},
			expected: "",
		},
		{
			name: "纯文本",
			delta: &models.DeltaData{
				Ops: []interface{}{
					map[string]interface{}{"insert": "Hello World"},
				},
			},
			expected: "Hello World",
		},
		{
			name: "带格式的文本",
			delta: &models.DeltaData{
				Ops: []interface{}{
					map[string]interface{}{
						"insert": "Bold",
						"attributes": map[string]interface{}{
							"bold": true,
						},
					},
					map[string]interface{}{
						"insert": " ",
					},
					map[string]interface{}{
						"insert": "Italic",
						"attributes": map[string]interface{}{
							"italic": true,
						},
					},
				},
			},
			expected: "Bold Italic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := editor.convertDeltaToPlainText(tt.delta)
			assert.Equal(t, tt.expected, result)
		})
	}
}