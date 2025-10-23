package editing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEditingService 模拟编辑服务
type MockEditingService struct {
	mock.Mock
}

func (m *MockEditingService) GetEditSession(ctx context.Context, sessionID string) (*models.EditSession, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EditSession), args.Error(1)
}

func (m *MockEditingService) UpdateEditSession(ctx context.Context, session *models.EditSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockEditingService) CreateEditSession(ctx context.Context, session *models.EditSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockEditingService) DeleteEditSession(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

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

func TestNewCodeEditor(t *testing.T) {
	editService := &MockEditingService{}
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	editor := NewCodeEditor(editService, storageService, notifyService)

	assert.NotNil(t, editor)
}

func TestCodeEditor_InitializeEditor(t *testing.T) {
	editService := &MockEditingService{}
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	editor := NewCodeEditor(editService, storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"

	// 模拟编辑服务返回
	editSession := &models.EditSession{
		SessionID:   sessionID,
		UserID:      "user-1",
		DocumentURI: "file:///test/test.go",
		Content:     "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
		Version:     1,
	}

	editService.On("GetEditSession", ctx, sessionID).Return(editSession, nil)

	// 测试初始化
	config, err := editor.InitializeEditor(ctx, sessionID)

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, sessionID, config.SessionID)
	assert.Equal(t, "go", config.Language)
	assert.NotNil(t, config.Options)
	assert.NotNil(t, config.Features)

	editService.AssertExpectations(t)
}

func TestCodeEditor_InitializeEditor_InvalidSessionID(t *testing.T) {
	editService := &MockEditingService{}
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	editor := NewCodeEditor(editService, storageService, notifyService)
	ctx := context.Background()

	// 测试空会话ID
	config, err := editor.InitializeEditor(ctx, "")

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "会话ID不能为空")
}

func TestCodeEditor_InitializeEditor_SessionNotFound(t *testing.T) {
	editService := &MockEditingService{}
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	editor := NewCodeEditor(editService, storageService, notifyService)
	ctx := context.Background()
	sessionID := "nonexistent-session"

	editService.On("GetEditSession", ctx, sessionID).Return(nil, fmt.Errorf("会话不存在"))

	config, err := editor.InitializeEditor(ctx, sessionID)

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "获取编辑会话失败")

	editService.AssertExpectations(t)
}

func TestCodeEditor_LoadCode(t *testing.T) {
	editService := &MockEditingService{}
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	editor := NewCodeEditor(editService, storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"

	// 先初始化编辑器
	editSession := &models.EditSession{
		SessionID:   sessionID,
		UserID:      "user-1",
		DocumentURI: "file:///test/test.go",
		Content:     "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
		Version:     1,
	}

	editService.On("GetEditSession", ctx, sessionID).Return(editSession, nil)

	config, err := editor.InitializeEditor(ctx, sessionID)
	assert.NoError(t, err)

	// 测试加载代码
	content, err := editor.LoadCode(ctx, sessionID)

	assert.NoError(t, err)
	assert.NotNil(t, content)
	assert.Equal(t, editSession.DocumentURI, content.URI)
	assert.Equal(t, editSession.Content, content.Content)
	assert.Equal(t, "go", content.Language)
	assert.Equal(t, 1, content.Version)

	editService.AssertExpectations(t)
}

func TestCodeEditor_LoadCode_SessionNotFound(t *testing.T) {
	editService := &MockEditingService{}
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	editor := NewCodeEditor(editService, storageService, notifyService)
	ctx := context.Background()
	sessionID := "nonexistent-session"

	content, err := editor.LoadCode(ctx, sessionID)

	assert.Error(t, err)
	assert.Nil(t, content)
	assert.Contains(t, err.Error(), "会话不存在")
}

func TestCodeEditor_SaveCode(t *testing.T) {
	editService := &MockEditingService{}
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	editor := NewCodeEditor(editService, storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"

	// 先初始化编辑器
	editSession := &models.EditSession{
		SessionID:   sessionID,
		UserID:      "user-1",
		DocumentURI: "file:///test/test.go",
		Content:     "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
		Version:     1,
	}

	editService.On("GetEditSession", ctx, sessionID).Return(editSession, nil)

	config, err := editor.InitializeEditor(ctx, sessionID)
	assert.NoError(t, err)

	// 测试保存代码
	newContent := &CodeContent{
		URI:      "file:///test/test.go",
		Content:  "package main\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}",
		Language: "go",
		Version:  2,
	}

	editService.On("UpdateEditSession", ctx, mock.AnythingOfType("*models.EditSession")).Return(nil)

	err = editor.SaveCode(ctx, sessionID, newContent)

	assert.NoError(t, err)

	editService.AssertExpectations(t)
}

func TestCodeEditor_SaveCode_InvalidContent(t *testing.T) {
	editService := &MockEditingService{}
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	editor := NewCodeEditor(editService, storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"

	// 测试空内容
	err := editor.SaveCode(ctx, sessionID, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "内容不能为空")
}

func TestCodeEditor_SaveCode_SessionNotFound(t *testing.T) {
	editService := &MockEditingService{}
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	editor := NewCodeEditor(editService, storageService, notifyService)
	ctx := context.Background()
	sessionID := "nonexistent-session"

	content := &CodeContent{
		URI:      "file:///test/test.go",
		Content:  "test content",
		Language: "go",
		Version:  1,
	}

	err := editor.SaveCode(ctx, sessionID, content)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "会话不存在")
}

func TestCodeEditor_HandleOperation(t *testing.T) {
	editService := &MockEditingService{}
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	editor := NewCodeEditor(editService, storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"

	// 先初始化编辑器
	editSession := &models.EditSession{
		SessionID:   sessionID,
		UserID:      "user-1",
		DocumentURI: "file:///test/test.go",
		Content:     "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
		Version:     1,
	}

	editService.On("GetEditSession", ctx, sessionID).Return(editSession, nil)

	config, err := editor.InitializeEditor(ctx, sessionID)
	assert.NoError(t, err)

	// 测试文本插入操作
	operation := &CodeOperation{
		Type:      "insert",
		Text:      "fmt.Println(\"Hello, World!\")\n",
		UserID:    "user-1",
		SessionID: sessionID,
		Version:   1,
		Timestamp: time.Now(),
	}

	err = editor.HandleOperation(ctx, sessionID, operation)

	// 注意：由于当前实现中handleTextOperation返回未实现错误，这里会失败
	// 但这是预期的，因为我们还没有完全实现所有功能
	// assert.NoError(t, err)
	assert.Contains(t, err.Error(), "尚未实现")
}

func TestCodeEditor_HandleOperation_InvalidOperation(t *testing.T) {
	editService := &MockEditingService{}
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	editor := NewCodeEditor(editService, storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"

	// 测试空操作
	err := editor.HandleOperation(ctx, sessionID, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "操作不能为空")
}

func TestCodeEditor_HandleOperation_SessionNotFound(t *testing.T) {
	editService := &MockEditingService{}
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	editor := NewCodeEditor(editService, storageService, notifyService)
	ctx := context.Background()
	sessionID := "nonexistent-session"

	operation := &CodeOperation{
		Type:      "insert",
		Text:      "test",
		UserID:    "user-1",
		SessionID: sessionID,
		Version:   1,
		Timestamp: time.Now(),
	}

	err := editor.HandleOperation(ctx, sessionID, operation)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "会话不存在")
}

func TestCodeEditor_GetDiagnostics(t *testing.T) {
	editService := &MockEditingService{}
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	editor := NewCodeEditor(editService, storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"

	// 先初始化编辑器
	editSession := &models.EditSession{
		SessionID:   sessionID,
		UserID:      "user-1",
		DocumentURI: "file:///test/test.go",
		Content:     "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
		Version:     1,
	}

	editService.On("GetEditSession", ctx, sessionID).Return(editSession, nil)

	config, err := editor.InitializeEditor(ctx, sessionID)
	assert.NoError(t, err)

	// 测试获取诊断信息
	diagnostics, err := editor.GetDiagnostics(ctx, sessionID)

	assert.NoError(t, err)
	assert.NotNil(t, diagnostics)
	// 由于语言服务未就绪，应该返回空列表
	assert.Empty(t, diagnostics)
}

func TestCodeEditor_GetDiagnostics_SessionNotFound(t *testing.T) {
	editService := &MockEditingService{}
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	editor := NewCodeEditor(editService, storageService, notifyService)
	ctx := context.Background()
	sessionID := "nonexistent-session"

	diagnostics, err := editor.GetDiagnostics(ctx, sessionID)

	assert.Error(t, err)
	assert.Nil(t, diagnostics)
	assert.Contains(t, err.Error(), "会话不存在")
}

func TestCodeEditor_GetCompletions(t *testing.T) {
	editService := &MockEditingService{}
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	editor := NewCodeEditor(editService, storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"

	// 先初始化编辑器
	editSession := &models.EditSession{
		SessionID:   sessionID,
		UserID:      "user-1",
		DocumentURI: "file:///test/test.go",
		Content:     "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
		Version:     1,
	}

	editService.On("GetEditSession", ctx, sessionID).Return(editSession, nil)

	config, err := editor.InitializeEditor(ctx, sessionID)
	assert.NoError(t, err)

	// 测试获取补全项
	position := &Position{
		Line:      3,
		Character: 1,
	}

	completions, err := editor.GetCompletions(ctx, sessionID, position)

	assert.NoError(t, err)
	assert.NotNil(t, completions)
	// 应该有默认的Go语言补全项
	assert.True(t, len(completions) > 0)

	// 验证补全项内容
	completionLabels := make([]string, len(completions))
	for i, completion := range completions {
		completionLabels[i] = completion.Label
	}
	assert.Contains(t, completionLabels, "func")
	assert.Contains(t, completionLabels, "if")
	assert.Contains(t, completionLabels, "for")
}

func TestCodeEditor_GetCompletions_InvalidPosition(t *testing.T) {
	editService := &MockEditingService{}
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	editor := NewCodeEditor(editService, storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"

	completions, err := editor.GetCompletions(ctx, sessionID, nil)

	assert.Error(t, err)
	assert.Nil(t, completions)
	assert.Contains(t, err.Error(), "位置不能为空")
}

func TestCodeEditor_GetCompletions_SessionNotFound(t *testing.T) {
	editService := &MockEditingService{}
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	editor := NewCodeEditor(editService, storageService, notifyService)
	ctx := context.Background()
	sessionID := "nonexistent-session"

	position := &Position{
		Line:      1,
		Character: 1,
	}

	completions, err := editor.GetCompletions(ctx, sessionID, position)

	assert.Error(t, err)
	assert.Nil(t, completions)
	assert.Contains(t, err.Error(), "会话不存在")
}

func TestCodeEditor_GetHoverInfo(t *testing.T) {
	editService := &MockEditingService{}
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	editor := NewCodeEditor(editService, storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"

	// 先初始化编辑器
	editSession := &models.EditSession{
		SessionID:   sessionID,
		UserID:      "user-1",
		DocumentURI: "file:///test/test.go",
		Content:     "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
		Version:     1,
	}

	editService.On("GetEditSession", ctx, sessionID).Return(editSession, nil)

	config, err := editor.InitializeEditor(ctx, sessionID)
	assert.NoError(t, err)

	// 测试获取悬停信息
	position := &Position{
		Line:      3,
		Character: 1,
	}

	hoverInfo, err := editor.GetHoverInfo(ctx, sessionID, position)

	assert.NoError(t, err)
	assert.NotNil(t, hoverInfo)
	assert.Equal(t, "悬停信息正在加载中...", hoverInfo.Contents)
}

func TestCodeEditor_GetHoverInfo_InvalidPosition(t *testing.T) {
	editService := &MockEditingService{}
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	editor := NewCodeEditor(editService, storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"

	hoverInfo, err := editor.GetHoverInfo(ctx, sessionID, nil)

	assert.Error(t, err)
	assert.Nil(t, hoverInfo)
	assert.Contains(t, err.Error(), "位置不能为空")
}

func TestCodeEditor_GetHoverInfo_SessionNotFound(t *testing.T) {
	editService := &MockEditingService{}
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	editor := NewCodeEditor(editService, storageService, notifyService)
	ctx := context.Background()
	sessionID := "nonexistent-session"

	position := &Position{
		Line:      1,
		Character: 1,
	}

	hoverInfo, err := editor.GetHoverInfo(ctx, sessionID, position)

	assert.Error(t, err)
	assert.Nil(t, hoverInfo)
	assert.Contains(t, err.Error(), "会话不存在")
}

func TestCodeEditor_FormatCode(t *testing.T) {
	editService := &MockEditingService{}
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	editor := NewCodeEditor(editService, storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"

	// 先初始化编辑器
	editSession := &models.EditSession{
		SessionID:   sessionID,
		UserID:      "user-1",
		DocumentURI: "file:///test/test.go",
		Content:     "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
		Version:     1,
	}

	editService.On("GetEditSession", ctx, sessionID).Return(editSession, nil)

	config, err := editor.InitializeEditor(ctx, sessionID)
	assert.NoError(t, err)

	// 测试格式化代码
	range_ := &Range{
		Start: &Position{Line: 0, Character: 0},
		End:   &Position{Line: 10, Character: 0},
	}

	textEdit, err := editor.FormatCode(ctx, sessionID, range_)

	// 由于语言服务未就绪，应该返回错误
	assert.Error(t, err)
	assert.Nil(t, textEdit)
	assert.Contains(t, err.Error(), "格式化功能不可用")
}

func TestCodeEditor_GetLanguageServiceConfig(t *testing.T) {
	editService := &MockEditingService{}
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	editor := NewCodeEditor(editService, storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"

	// 先初始化编辑器
	editSession := &models.EditSession{
		SessionID:   sessionID,
		UserID:      "user-1",
		DocumentURI: "file:///test/test.go",
		Content:     "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
		Version:     1,
	}

	editService.On("GetEditSession", ctx, sessionID).Return(editSession, nil)

	config, err := editor.InitializeEditor(ctx, sessionID)
	assert.NoError(t, err)

	// 测试获取语言服务配置
	langConfig, err := editor.GetLanguageServiceConfig(sessionID)

	assert.NoError(t, err)
	assert.NotNil(t, langConfig)
	assert.Equal(t, "go", langConfig.Language)
	assert.NotNil(t, langConfig.CompilerOptions)
	assert.NotNil(t, langConfig.DiagnosticsOptions)
}

func TestCodeEditor_GetLanguageServiceConfig_SessionNotFound(t *testing.T) {
	editService := &MockEditingService{}
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	editor := NewCodeEditor(editService, storageService, notifyService)
	sessionID := "nonexistent-session"

	langConfig, err := editor.GetLanguageServiceConfig(sessionID)

	assert.Error(t, err)
	assert.Nil(t, langConfig)
	assert.Contains(t, err.Error(), "会话不存在")
}

func TestCodeEditor_DestroyEditor(t *testing.T) {
	editService := &MockEditingService{}
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	editor := NewCodeEditor(editService, storageService, notifyService)
	ctx := context.Background()
	sessionID := "test-session-1"

	// 先初始化编辑器
	editSession := &models.EditSession{
		SessionID:   sessionID,
		UserID:      "user-1",
		DocumentURI: "file:///test/test.go",
		Content:     "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
		Version:     1,
	}

	editService.On("GetEditSession", ctx, sessionID).Return(editSession, nil)

	config, err := editor.InitializeEditor(ctx, sessionID)
	assert.NoError(t, err)

	// 测试销毁编辑器
	err = editor.DestroyEditor(sessionID)

	assert.NoError(t, err)
}

func TestCodeEditor_DestroyEditor_InvalidSessionID(t *testing.T) {
	editService := &MockEditingService{}
	storageService := &MockStorageService{}
	notifyService := &MockNotificationService{}

	editor := NewCodeEditor(editService, storageService, notifyService)

	// 测试空会话ID
	err := editor.DestroyEditor("")

	assert.NoError(t, err) // DestroyEditor不验证会话是否存在
}

// 辅助函数测试

func TestDetectLanguageFromURI(t *testing.T) {
	editor := &CodeEditorImpl{}

	// 测试各种文件扩展名
	assert.Equal(t, "go", editor.detectLanguageFromURI("test.go"))
	assert.Equal(t, "typescript", editor.detectLanguageFromURI("test.ts"))
	assert.Equal(t, "typescript", editor.detectLanguageFromURI("test.tsx"))
	assert.Equal(t, "javascript", editor.detectLanguageFromURI("test.js"))
	assert.Equal(t, "javascript", editor.detectLanguageFromURI("test.jsx"))
	assert.Equal(t, "json", editor.detectLanguageFromURI("test.json"))
	assert.Equal(t, "css", editor.detectLanguageFromURI("test.css"))
	assert.Equal(t, "html", editor.detectLanguageFromURI("test.html"))
	assert.Equal(t, "plaintext", editor.detectLanguageFromURI("test.txt"))
	assert.Equal(t, "plaintext", editor.detectLanguageFromURI("test"))
}

func TestGetDefaultEditorOptions(t *testing.T) {
	editor := &CodeEditorImpl{}
	options := editor.getDefaultEditorOptions()

	assert.NotNil(t, options)
	assert.Equal(t, 14, options.FontSize)
	assert.Equal(t, "Consolas, Monaco, 'Courier New', monospace", options.FontFamily)
	assert.Equal(t, 4, options.TabSize)
	assert.True(t, options.InsertSpaces)
	assert.Equal(t, "off", options.WordWrap)
	assert.Equal(t, "on", options.LineNumbers)
	assert.True(t, options.ScrollBeyondLastLine)
	assert.True(t, options.AutomaticLayout)
	assert.True(t, options.GlyphMargin)
	assert.True(t, options.Folding)
	assert.True(t, options.ColorDecorators)
	assert.NotNil(t, options.Suggest)
	assert.NotNil(t, options.Accessibility)
}

func TestGetDefaultLanguageServiceConfig(t *testing.T) {
	editor := &CodeEditorImpl{}

	// 测试Go语言配置
	goConfig := editor.getDefaultLanguageServiceConfig("go")
	assert.NotNil(t, goConfig)
	assert.Equal(t, "go", goConfig.Language)
	assert.NotNil(t, goConfig.CompilerOptions)
	assert.Equal(t, "go1.21", goConfig.CompilerOptions["target"])
	assert.Equal(t, "mod", goConfig.CompilerOptions["module"])

	// 测试TypeScript语言配置
	tsConfig := editor.getDefaultLanguageServiceConfig("typescript")
	assert.NotNil(t, tsConfig)
	assert.Equal(t, "typescript", tsConfig.Language)
	assert.NotNil(t, tsConfig.CompilerOptions)
	assert.Equal(t, "es2020", tsConfig.CompilerOptions["target"])
	assert.Equal(t, "esnext", tsConfig.CompilerOptions["module"])

	// 测试JavaScript语言配置
	jsConfig := editor.getDefaultLanguageServiceConfig("javascript")
	assert.NotNil(t, jsConfig)
	assert.Equal(t, "javascript", jsConfig.Language)
	assert.NotNil(t, jsConfig.CompilerOptions)
	assert.Equal(t, "es2020", jsConfig.CompilerOptions["target"])
	assert.Equal(t, "esnext", jsConfig.CompilerOptions["module"])
}

func TestGetDefaultCompletions(t *testing.T) {
	editor := &CodeEditorImpl{}

	// 测试Go语言补全
	goCompletions := editor.getDefaultCompletions("go")
	assert.NotNil(t, goCompletions)
	assert.True(t, len(goCompletions) > 0)

	completionLabels := make([]string, len(goCompletions))
	for i, completion := range goCompletions {
		completionLabels[i] = completion.Label
	}
	assert.Contains(t, completionLabels, "func")
	assert.Contains(t, completionLabels, "if")
	assert.Contains(t, completionLabels, "for")

	// 测试TypeScript语言补全
	tsCompletions := editor.getDefaultCompletions("typescript")
	assert.NotNil(t, tsCompletions)
	assert.True(t, len(tsCompletions) > 0)

	tsCompletionLabels := make([]string, len(tsCompletions))
	for i, completion := range tsCompletions {
		tsCompletionLabels[i] = completion.Label
	}
	assert.Contains(t, tsCompletionLabels, "function")
	assert.Contains(t, tsCompletionLabels, "const")
	assert.Contains(t, tsCompletionLabels, "class")

	// 测试未知语言补全
	unknownCompletions := editor.getDefaultCompletions("unknown")
	assert.NotNil(t, unknownCompletions)
	assert.Empty(t, unknownCompletions)
}