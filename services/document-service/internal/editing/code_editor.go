package editing

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CodeEditor 代码编辑器接口
type CodeEditor interface {
	// 初始化编辑器
	InitializeEditor(ctx context.Context, sessionID string) (*CodeEditorConfig, error)
	// 加载代码内容
	LoadCode(ctx context.Context, sessionID string) (*CodeContent, error)
	// 保存代码内容
	SaveCode(ctx context.Context, sessionID string, content *CodeContent) error
	// 处理编辑操作
	HandleOperation(ctx context.Context, sessionID string, operation *CodeOperation) error
	// 获取诊断信息
	GetDiagnostics(ctx context.Context, sessionID string) ([]*DiagnosticInfo, error)
	// 获取代码补全
	GetCompletions(ctx context.Context, sessionID string, position *Position) ([]*CompletionItem, error)
	// 获取悬停信息
	GetHoverInfo(ctx context.Context, sessionID string, position *Position) (*HoverInfo, error)
	// 格式化代码
	FormatCode(ctx context.Context, sessionID string, range_ *Range) (*TextEdit, error)
	// 获取语言服务配置
	GetLanguageServiceConfig(sessionID string) (*LanguageServiceConfig, error)
	// 销毁编辑器
	DestroyEditor(sessionID string) error
}

// CodeEditorImpl 代码编辑器实现
type CodeEditorImpl struct {
	editService    EditingService
	storageService StorageService
	notifyService  NotificationService

	// 会话管理
	sessions map[string]*CodeEditorSession
	mutex    sync.RWMutex

	// 语言服务管理
	languageServices map[string]*LanguageService
	lsMutex          sync.RWMutex
}

// CodeEditorSession 代码编辑器会话
type CodeEditorSession struct {
	ID            string
	UserID        string
	Document      *CodeDocument
	Language      string
	Theme         string
	Options       *EditorOptions
	Diagnostics   []*DiagnosticInfo
	CreatedAt     time.Time
	LastActivity  time.Time
	IsActive      bool
}

// CodeDocument 代码文档
type CodeDocument struct {
	URI     string
	Content string
	Version int
	Language string
	Changes []*DocumentChange
	mutex   sync.RWMutex
}

// DocumentChange 文档变更
type DocumentChange struct {
	Type        string    `json:"type"` // "insert", "delete", "replace"
	Range       *Range    `json:"range"`
	Text        string    `json:"text"`
	Version     int       `json:"version"`
	UserID      string    `json:"user_id"`
	Timestamp   time.Time `json:"timestamp"`
}

// CodeEditorConfig 代码编辑器配置
type CodeEditorConfig struct {
	SessionID    string            `json:"session_id"`
	Language     string            `json:"language"`
	Theme        string            `json:"theme"`
	Options      *EditorOptions    `json:"options"`
	Features     *EditorFeatures   `json:"features"`
	Workspace    *WorkspaceConfig  `json:"workspace,omitempty"`
	LanguageConfig *LanguageServiceConfig `json:"language_config,omitempty"`
}

// EditorOptions 编辑器选项
type EditorOptions struct {
	FontSize              int                    `json:"font_size"`
	FontFamily            string                 `json:"font_family"`
	TabSize               int                    `json:"tab_size"`
	InsertSpaces          bool                   `json:"insert_spaces"`
	WordWrap              string                 `json:"word_wrap"` // "on", "off", "wordWrapColumn", "bounded"
	Minimap               map[string]interface{} `json:"minimap"`
	LineNumbers           string                 `json:"line_numbers"` // "on", "off", "relative", "interval"
	ScrollBeyondLastLine  bool                   `json:"scroll_beyond_last_line"`
	AutomaticLayout       bool                   `json:"automatic_layout"`
	GlyphMargin           bool                   `json:"glyph_margin"`
	Folding               bool                   `json:"folding"`
	Brackets              map[string]interface{} `json:"brackets"`
	ColorDecorators       bool                   `json:"color_decorators"`
	LightBulb             map[string]interface{} `json:"lightbulb"`
	CodeLens              bool                   `json:"code_lens"`
	Suggest               *SuggestOptions        `json:"suggest"`
	Accessibility         *AccessibilityOptions  `json:"accessibility"`
}

// SuggestOptions 建议选项
type SuggestOptions struct {
	ShowSnippets bool `json:"show_snippets"`
	ShowKeywords  bool `json:"show_keywords"`
	ShowFunctions bool `json:"show_functions"`
	ShowTypes     bool `json:"show_types"`
}

// AccessibilityOptions 辅助功能选项
type AccessibilityOptions struct {
	Support string `json:"support"` // "on", "off", "auto"
	HelpURL string `json:"help_url,omitempty"`
}

// EditorFeatures 编辑器特性
type EditorFeatures struct {
	AutoComplete     bool `json:"auto_complete"`
	IntelliSense     bool `json:"intellisense"`
	ErrorDiagnostics bool `json:"error_diagnostics"`
	FormatOnType     bool `json:"format_on_type"`
	FormatOnSave     bool `json:"format_on_save"`
	Snippets         bool `json:"snippets"`
	MultiCursor      bool `json:"multi_cursor"`
	Search           bool `json:"search"`
	FindReferences   bool `json:"find_references"`
	Rename            bool `json:"rename"`
	GoToDefinition   bool `json:"go_to_definition"`
	GoToImplementation bool `json:"go_to_implementation"`
	GoToTypeDefinition bool `json:"go_to_type_definition"`
	Hover             bool `json:"hover"`
	SignatureHelp     bool `json:"signature_help"`
	BraceMatching     bool `json:"brace_matching"`
	Indentation       bool `json:"indentation"`
	CodeLens          bool `json:"code_lens"`
	Folding           bool `json:"folding"`
	Minimap           bool `json:"minimap"`
	Theme             bool `json:"theme"`
	KeyBindings       bool `json:"key_bindings"`
	Commands          bool `json:"commands"`
	ContextMenu       bool `json:"context_menu"`
}

// WorkspaceConfig 工作区配置
type WorkspaceConfig struct {
	RootURI    string            `json:"root_uri"`
	Folders    []*WorkspaceFolder `json:"folders"`
	Settings   map[string]interface{} `json:"settings"`
	Extensions []*ExtensionInfo   `json:"extensions"`
}

// WorkspaceFolder 工作区文件夹
type WorkspaceFolder struct {
	URI  string `json:"uri"`
	Name string `json:"name"`
}

// ExtensionInfo 扩展信息
type ExtensionInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Enabled bool   `json:"enabled"`
}

// CodeContent 代码内容
type CodeContent struct {
	URI      string      `json:"uri"`
	Content  string      `json:"content"`
	Language string      `json:"language"`
	Version  int         `json:"version"`
	Metadata interface{} `json:"metadata,omitempty"`
}

// CodeOperation 代码操作
type CodeOperation struct {
	Type      string      `json:"type"` // "insert", "delete", "replace", "format", "complete", "hover", "diagnostic"
	Range     *Range      `json:"range,omitempty"`
	Text      string      `json:"text,omitempty"`
	Position  *Position   `json:"position,omitempty"`
	UserID    string      `json:"user_id"`
	SessionID string      `json:"session_id"`
	Version   int         `json:"version"`
	Metadata  interface{} `json:"metadata,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// Position 位置
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range 范围
type Range struct {
	Start *Position `json:"start"`
	End   *Position `json:"end"`
}

// DiagnosticInfo 诊断信息
type DiagnosticInfo struct {
	Range     *Range     `json:"range"`
	Severity  int        `json:"severity"` // 1: Error, 2: Warning, 3: Information, 4: Hint
	Source    string     `json:"source"`
	Message   string     `json:"message"`
	Code      string     `json:"code,omitempty"`
	Tags      []int      `json:"tags,omitempty"`
	Related   []*RelatedDiagnostic `json:"related,omitempty"`
}

// RelatedDiagnostic 相关诊断
type RelatedDiagnostic struct {
	Range   *Range `json:"range"`
	Message string `json:"message"`
}

// CompletionItem 补全项
type CompletionItem struct {
	Label         string            `json:"label"`
	Kind          int               `json:"kind"`
	Detail        string            `json:"detail,omitempty"`
	Documentation string            `json:"documentation,omitempty"`
	SortText      string            `json:"sort_text,omitempty"`
	FilterText    string            `json:"filter_text,omitempty"`
	InsertText    string            `json:"insert_text,omitempty"`
	TextEdit      *TextEdit         `json:"text_edit,omitempty"`
	Additional    []*CompletionItem `json:"additional,omitempty"`
	Tags          []int             `json:"tags,omitempty"`
	Data          interface{}       `json:"data,omitempty"`
}

// TextEdit 文本编辑
type TextEdit struct {
	Range *Range `json:"range"`
	Text  string `json:"text"`
}

// HoverInfo 悬停信息
type HoverInfo struct {
	Contents interface{} `json:"contents"`
	Range    *Range     `json:"range,omitempty"`
}

// LanguageService 语言服务
type LanguageService struct {
	Language      string
	Configuration *LanguageServiceConfig
	Worker        *LanguageWorker
	Diagnostics   map[string][]*DiagnosticInfo
	mutex         sync.RWMutex
}

// LanguageServiceConfig 语言服务配置
type LanguageServiceConfig struct {
	Language             string                 `json:"language"`
	CompilerOptions      map[string]interface{} `json:"compiler_options"`
	DiagnosticsOptions   *DiagnosticsOptions    `json:"diagnostics_options"`
	ExtraLibraries       []*ExtraLibrary        `json:"extra_libraries"`
	TypeDefinitions      []*TypeDefinition      `json:"type_definitions"`
	Snippets             []*Snippet             `json:"snippets"`
	Themes               []*EditorTheme          `json:"themes"`
	KeyBindings          []*KeyBinding           `json:"key_bindings"`
	Commands             []*Command             `json:"commands"`
	ValidationSettings   map[string]interface{} `json:"validation_settings"`
	WorkspaceSettings    map[string]interface{} `json:"workspace_settings"`
}

// DiagnosticsOptions 诊断选项
type DiagnosticsOptions struct {
	NoSemanticValidation bool `json:"no_semantic_validation"`
	NoSyntaxValidation    bool `json:"no_syntax_validation"`
	NoUnusedParameters    bool `json:"no_unused_parameters"`
	NoUnusedLocals        bool `json:"no_unused_locals"`
	NoImplicitReturns     bool `json:"no_implicit_returns"`
}

// ExtraLibrary 扩展库
type ExtraLibrary struct {
	Name      string `json:"name"`
	Content   string `json:"content"`
	URI       string `json:"uri"`
	IsDefault bool   `json:"is_default"`
}

// TypeDefinition 类型定义
type TypeDefinition struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
	Location   *Range `json:"location"`
	Exported   bool   `json:"exported"`
}

// Snippet 代码片段
type Snippet struct {
	Name        string `json:"name"`
	Prefix      string `json:"prefix"`
	Body        string `json:"body"`
	Description string `json:"description"`
	Scope       string `json:"scope"`
}

// EditorTheme 编辑器主题
type EditorTheme struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"` // "light", "dark", "hc"
	Base       string                 `json:"base"`
	Rules      []*ThemeRule           `json:"rules"`
	Colors     map[string]interface{} `json:"colors"`
	TokenColors []*TokenColor          `json:"token_colors"`
}

// ThemeRule 主题规则
type ThemeRule struct {
	Token   string     `json:"token"`
	Foreground string  `json:"foreground"`
	Background string  `json:"background"`
	FontStyle   string  `json:"font_style"` // "italic", "bold", "underline"
}

// TokenColor 标记颜色
type TokenColor struct {
	Scope   []string          `json:"scope"`
	Settings map[string]interface{} `json:"settings"`
}

// KeyBinding 键绑定
type KeyBinding struct {
	Key      string   `json:"key"`
	Command  string   `json:"command"`
	Args     interface{} `json:"args,omitempty"`
	When     string   `json:"when,omitempty"`
}

// Command 命令
type Command struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description,omitempty"`
	Args        interface{} `json:"args,omitempty"`
	Category    string      `json:"category,omitempty"`
}

// LanguageWorker 语言工作器
type LanguageWorker struct {
	ID       string
	Language string
	URL      string
	Worker   interface{} // WebSocket连接或Web Worker实例
	Ready    bool
	mutex    sync.RWMutex
}

// NewCodeEditor 创建代码编辑器
func NewCodeEditor(
	editService EditingService,
	storageService StorageService,
	notifyService NotificationService,
) CodeEditor {
	editor := &CodeEditorImpl{
		editService:    editService,
		storageService: storageService,
		notifyService:  notifyService,
		sessions:       make(map[string]*CodeEditorSession),
		languageServices: make(map[string]*LanguageService),
	}

	// 初始化语言服务
	editor.initializeLanguageServices()

	return editor
}

// InitializeEditor 初始化编辑器
func (e *CodeEditorImpl) InitializeEditor(ctx context.Context, sessionID string) (*CodeEditorConfig, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("会话ID不能为空")
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	// 检查会话是否已存在
	if session, exists := e.sessions[sessionID]; exists {
		session.LastActivity = time.Now()
		session.IsActive = true

		return e.buildConfigFromSession(session), nil
	}

	// 获取编辑会话
	editSession, err := e.editService.GetEditSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("获取编辑会话失败: %w", err)
	}

	// 创建新会话
	session := &CodeEditorSession{
		ID:           sessionID,
		UserID:       editSession.UserID,
		Language:     detectLanguageFromURI(editSession.DocumentURI),
		Theme:        "vs-dark",
		Options:      e.getDefaultEditorOptions(),
		Diagnostics:  make([]*DiagnosticInfo, 0),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		IsActive:     true,
	}

	// 创建文档
	session.Document = &CodeDocument{
		URI:      editSession.DocumentURI,
		Content:  editSession.Content,
		Version:  1,
		Language: session.Language,
		Changes:  make([]*DocumentChange, 0),
	}

	e.sessions[sessionID] = session

	return e.buildConfigFromSession(session), nil
}

// LoadCode 加载代码内容
func (e *CodeEditorImpl) LoadCode(ctx context.Context, sessionID string) (*CodeContent, error) {
	e.mutex.RLock()
	session, exists := e.sessions[sessionID]
	e.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("会话不存在")
	}

	session.Document.mutex.RLock()
	defer session.Document.mutex.RUnlock()

	return &CodeContent{
		URI:      session.Document.URI,
		Content:  session.Document.Content,
		Language: session.Document.Language,
		Version:  session.Document.Version,
	}, nil
}

// SaveCode 保存代码内容
func (e *CodeEditorImpl) SaveCode(ctx context.Context, sessionID string, content *CodeContent) error {
	if content == nil {
		return fmt.Errorf("内容不能为空")
	}

	e.mutex.RLock()
	session, exists := e.sessions[sessionID]
	e.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("会话不存在")
	}

	// 验证权限
	if err := e.checkEditPermission(ctx, session.UserID); err != nil {
		return err
	}

	// 更新文档内容
	session.Document.mutex.Lock()
	session.Document.Content = content.Content
	session.Document.Version++
	session.Document.Language = content.Language
	session.Document.mutex.Unlock()

	// 保存到存储
	if err := e.editService.UpdateEditSession(ctx, sessionID, content.Content, session.Document.Version); err != nil {
		return fmt.Errorf("保存编辑会话失败: %w", err)
	}

	// 运行诊断
	go e.runDiagnostics(sessionID)

	return nil
}

// HandleOperation 处理编辑操作
func (e *CodeEditorImpl) HandleOperation(ctx context.Context, sessionID string, operation *CodeOperation) error {
	if operation == nil {
		return fmt.Errorf("操作不能为空")
	}

	e.mutex.RLock()
	session, exists := e.sessions[sessionID]
	e.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("会话不存在")
	}

	// 验证权限
	if err := e.checkEditPermission(ctx, session.UserID); err != nil {
		return err
	}

	// 处理不同类型的操作
	switch operation.Type {
	case "insert", "delete", "replace":
		return e.handleTextOperation(sessionID, operation)
	case "format":
		return e.handleFormatOperation(ctx, sessionID, operation)
	case "complete":
		return e.handleCompletionOperation(ctx, sessionID, operation)
	case "hover":
		return e.handleHoverOperation(ctx, sessionID, operation)
	case "diagnostic":
		return e.handleDiagnosticOperation(ctx, sessionID, operation)
	default:
		return fmt.Errorf("不支持的操作类型: %s", operation.Type)
	}
}

// GetDiagnostics 获取诊断信息
func (e *CodeEditorImpl) GetDiagnostics(ctx context.Context, sessionID string) ([]*DiagnosticInfo, error) {
	e.mutex.RLock()
	session, exists := e.sessions[sessionID]
	e.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("会话不存在")
	}

	// 获取语言服务
	ls, exists := e.getLanguageService(session.Document.Language)
	if !exists || !ls.Ready {
		return make([]*DiagnosticInfo, 0), nil
	}

	ls.mutex.RLock()
	diagnostics := session.Diagnostics
	ls.mutex.RUnlock()

	return diagnostics, nil
}

// GetCompletions 获取代码补全
func (e *CodeEditorImpl) GetCompletions(ctx context.Context, sessionID string, position *Position) ([]*CompletionItem, error) {
	if position == nil {
		return nil, fmt.Errorf("位置不能为空")
	}

	e.mutex.RLock()
	session, exists := e.sessions[sessionID]
	e.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("会话不存在")
	}

	// 获取语言服务
	ls, exists := e.getLanguageService(session.Document.Language)
	if !exists || !ls.Ready {
		return e.getDefaultCompletions(session.Document.Language), nil
	}

	// 从语言服务获取补全项
	return e.getLanguageServiceCompletions(ls, sessionID, position)
}

// GetHoverInfo 获取悬停信息
func (e *CodeEditorImpl) GetHoverInfo(ctx context.Context, sessionID string, position *Position) (*HoverInfo, error) {
	if position == nil {
		return nil, fmt.Errorf("位置不能为空")
	}

	e.mutex.RLock()
	session, exists := e.sessions[sessionID]
	e.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("会话不存在")
	}

	// 获取语言服务
	ls, exists := e.getLanguageService(session.Document.Language)
	if !exists || !ls.Ready {
		return &HoverInfo{
			Contents: "悬停信息不可用",
		}, nil
	}

	// 从语言服务获取悬停信息
	return e.getLanguageServiceHover(ls, sessionID, position)
}

// FormatCode 格式化代码
func (e *CodeEditorImpl) FormatCode(ctx context.Context, sessionID string, range_ *Range) (*TextEdit, error) {
	e.mutex.RLock()
	session, exists := e.sessions[sessionID]
	e.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("会话不存在")
	}

	// 验证权限
	if err := e.checkEditPermission(ctx, session.UserID); err != nil {
		return nil, err
	}

	// 获取语言服务
	ls, exists := e.getLanguageService(session.Document.Language)
	if !exists || !ls.Ready {
		return nil, fmt.Errorf("格式化功能不可用")
	}

	// 执行格式化
	return e.formatCodeWithLanguageService(ls, sessionID, range_)
}

// GetLanguageServiceConfig 获取语言服务配置
func (e *CodeEditorImpl) GetLanguageServiceConfig(sessionID string) (*LanguageServiceConfig, error) {
	e.mutex.RLock()
	session, exists := e.sessions[sessionID]
	e.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("会话不存在")
	}

	ls, exists := e.getLanguageService(session.Document.Language)
	if !exists {
		return nil, fmt.Errorf("语言服务不存在")
	}

	return ls.Configuration, nil
}

// DestroyEditor 销毁编辑器
func (e *CodeEditorImpl) DestroyEditor(sessionID string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if session, exists := e.sessions[sessionID]; exists {
		session.IsActive = false
		// 清理资源，但不立即删除，给用户重连的机会
		go func() {
			time.Sleep(5 * time.Minute)
			e.mutex.Lock()
			if s, stillExists := e.sessions[sessionID]; stillExists && !s.IsActive {
				delete(e.sessions, sessionID)
			}
			e.mutex.Unlock()
		}()
	}

	return nil
}

// 辅助方法

// detectLanguageFromURI 从URI检测语言
func detectLanguageFromURI(uri string) string {
	// 简化实现，根据文件扩展名检测语言
	switch {
	case len(uri) > 3 && uri[len(uri)-3:] == ".go":
		return "go"
	case len(uri) > 4 && uri[len(uri)-4:] == ".tsx":
		return "typescript"
	case len(uri) > 3 && uri[len(uri)-3:] == ".ts":
		return "typescript"
	case len(uri) > 4 && uri[len(uri)-4:] == ".jsx":
		return "javascript"
	case len(uri) > 3 && uri[len(uri)-3:] == ".js":
		return "javascript"
	case len(uri) > 5 && uri[len(uri)-5:] == ".json":
		return "json"
	case len(uri) > 4 && uri[len(uri)-4:] == ".css":
		return "css"
	case len(uri) > 5 && uri[len(uri)-5:] == ".html":
		return "html"
	default:
		return "plaintext"
	}
}

// getDefaultEditorOptions 获取默认编辑器选项
func (e *CodeEditorImpl) getDefaultEditorOptions() *EditorOptions {
	return &EditorOptions{
		FontSize:              14,
		FontFamily:            "Consolas, Monaco, 'Courier New', monospace",
		TabSize:               4,
		InsertSpaces:          true,
		WordWrap:              "off",
		LineNumbers:           "on",
		ScrollBeyondLastLine:  true,
		AutomaticLayout:       true,
		GlyphMargin:           true,
		Folding:               true,
		ColorDecorators:       true,
		Suggest: &SuggestOptions{
			ShowSnippets: true,
			ShowKeywords:  true,
			ShowFunctions: true,
			ShowTypes:     true,
		},
		Accessibility: &AccessibilityOptions{
			Support: "auto",
		},
	}
}

// buildConfigFromSession 从会话构建配置
func (e *CodeEditorImpl) buildConfigFromSession(session *CodeEditorSession) *CodeEditorConfig {
	config := &CodeEditorConfig{
		SessionID: session.ID,
		Language:  session.Document.Language,
		Theme:     session.Theme,
		Options:   session.Options,
		Features: &EditorFeatures{
			AutoComplete:        true,
			IntelliSense:        true,
			ErrorDiagnostics:   true,
			FormatOnType:        true,
			FormatOnSave:        true,
			Snippets:            true,
			MultiCursor:         true,
			Search:              true,
			FindReferences:      true,
			Rename:              true,
			GoToDefinition:      true,
			GoToImplementation:  true,
			GoToTypeDefinition:  true,
			Hover:               true,
			SignatureHelp:       true,
			BraceMatching:       true,
			Indentation:         true,
			CodeLens:            true,
			Folding:             true,
			Minimap:             true,
			Theme:               true,
			KeyBindings:         true,
			Commands:            true,
			ContextMenu:        true,
		},
	}

	// 添加语言服务配置
	if ls, exists := e.getLanguageService(session.Document.Language); exists {
		config.LanguageConfig = ls.Configuration
	}

	return config
}

// checkEditPermission 检查编辑权限
func (e *CodeEditorImpl) checkEditPermission(ctx context.Context, userID string) error {
	// 这里应该实现权限检查逻辑
	// 暂时返回nil，表示有权限
	return nil
}

// initializeLanguageServices 初始化语言服务
func (e *CodeEditorImpl) initializeLanguageServices() {
	// 初始化支持的语言服务
	supportedLanguages := []string{"go", "typescript", "javascript", "json", "css", "html"}

	for _, lang := range supportedLanguages {
		ls := &LanguageService{
			Language:      lang,
			Configuration: e.getDefaultLanguageServiceConfig(lang),
			Worker:        nil,
			Diagnostics:   make(map[string][]*DiagnosticInfo),
		}

		e.lsMutex.Lock()
		e.languageServices[lang] = ls
		e.lsMutex.Unlock()
	}
}

// getDefaultLanguageServiceConfig 获取默认语言服务配置
func (e *CodeEditorImpl) getDefaultLanguageServiceConfig(language string) *LanguageServiceConfig {
	config := &LanguageServiceConfig{
		Language: language,
		CompilerOptions: make(map[string]interface{}),
		DiagnosticsOptions: &DiagnosticsOptions{
			NoSemanticValidation: false,
			NoSyntaxValidation:    false,
		},
		ExtraLibraries:     make([]*ExtraLibrary, 0),
		TypeDefinitions:    make([]*TypeDefinition, 0),
		Snippets:           make([]*Snippet, 0),
		Themes:             make([]*EditorTheme, 0),
		KeyBindings:        make([]*KeyBinding, 0),
		Commands:           make([]*Command, 0),
		ValidationSettings: make(map[string]interface{}),
		WorkspaceSettings:  make(map[string]interface{}),
	}

	// 根据语言设置特定配置
	switch language {
	case "go":
		config.CompilerOptions = map[string]interface{}{
			"target": "go1.21",
			"module": "mod",
		}
	case "typescript", "javascript":
		config.CompilerOptions = map[string]interface{}{
			"target":                "es2020",
			"module":               "esnext",
			"moduleResolution":      "node",
			"lib":                   []string{"es2020", "dom"},
			"strict":                true,
			"esModuleInterop":       true,
			"skipLibCheck":          true,
			"forceConsistentCasingInFileNames": true,
		}
	}

	return config
}

// getLanguageService 获取语言服务
func (e *CodeEditorImpl) getLanguageService(language string) (*LanguageService, bool) {
	e.lsMutex.RLock()
	defer e.lsMutex.RUnlock()

	ls, exists := e.languageServices[language]
	return ls, exists
}

// runDiagnostics 运行诊断
func (e *CodeEditorImpl) runDiagnostics(sessionID string) {
	// 这里应该实现实际的诊断逻辑
	// 暂时为空实现
}

// handleTextOperation 处理文本操作
func (e *CodeEditorImpl) handleTextOperation(sessionID string, operation *CodeOperation) error {
	e.mutex.RLock()
	session, exists := e.sessions[sessionID]
	e.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("会话不存在")
	}

	// 创建文档变更
	change := &DocumentChange{
		Type:      operation.Type,
		Range:     operation.Range,
		Text:      operation.Text,
		Version:   session.Document.Version + 1,
		UserID:    operation.UserID,
		Timestamp: time.Now(),
	}

	// 应用变更
	session.Document.mutex.Lock()
	defer session.Document.mutex.Unlock()

	// 简化实现：直接应用文本变更
	if operation.Type == "insert" {
		session.Document.Content = operation.Text + session.Document.Content
	} else if operation.Type == "delete" && operation.Range != nil {
		// 这里应该实现范围删除逻辑
	}

	session.Document.Version++
	session.Document.Changes = append(session.Document.Changes, change)

	return nil
}

// handleFormatOperation 处理格式化操作
func (e *CodeEditorImpl) handleFormatOperation(ctx context.Context, sessionID string, operation *CodeOperation) error {
	// 实现格式化操作
	return fmt.Errorf("格式化操作尚未实现")
}

// handleCompletionOperation 处理补全操作
func (e *CodeEditorImpl) handleCompletionOperation(ctx context.Context, sessionID string, operation *CodeOperation) error {
	// 实现补全操作
	return fmt.Errorf("补全操作尚未实现")
}

// handleHoverOperation 处理悬停操作
func (e *CodeEditorImpl) handleHoverOperation(ctx context.Context, sessionID string, operation *CodeOperation) error {
	// 实现悬停操作
	return fmt.Errorf("悬停操作尚未实现")
}

// handleDiagnosticOperation 处理诊断操作
func (e *CodeEditorImpl) handleDiagnosticOperation(ctx context.Context, sessionID string, operation *CodeOperation) error {
	// 实现诊断操作
	return fmt.Errorf("诊断操作尚未实现")
}

// getDefaultCompletions 获取默认补全项
func (e *CodeEditorImpl) getDefaultCompletions(language string) []*CompletionItem {
	completions := make([]*CompletionItem, 0)

	// 根据语言提供基本的补全项
	switch language {
	case "go":
		completions = append(completions,
			&CompletionItem{Label: "func", Kind: 15, InsertText: "func ${1:name}() {\n\t${2:// TODO}\n}"},
			&CompletionItem{Label: "if", Kind: 15, InsertText: "if ${1:condition} {\n\t${2:// TODO}\n}"},
			&CompletionItem{Label: "for", Kind: 15, InsertText: "for ${1:i := 0}; ${2:i < ${3:10}}; ${4:i++} {\n\t${5:// TODO}\n}"},
		)
	case "typescript", "javascript":
		completions = append(completions,
			&CompletionItem{Label: "function", Kind: 15, InsertText: "function ${1:name}(${2:params}) {\n\t${3:// TODO}\n}"},
			&CompletionItem{Label: "const", Kind: 15, InsertText: "const ${1:name} = ${2:value}"},
			&CompletionItem{Label: "class", Kind: 15, InsertText: "class ${1:Name} {\n\tconstructor(${2:params}) {\n\t\t${3:// TODO}\n\t}\n}"},
		)
	}

	return completions
}

// getLanguageServiceCompletions 从语言服务获取补全
func (e *CodeEditorImpl) getLanguageServiceCompletions(ls *LanguageService, sessionID string, position *Position) ([]*CompletionItem, error) {
	// 这里应该实现与语言服务的通信
	return e.getDefaultCompletions(ls.Language), nil
}

// getLanguageServiceHover 从语言服务获取悬停信息
func (e *CodeEditorImpl) getLanguageServiceHover(ls *LanguageService, sessionID string, position *Position) (*HoverInfo, error) {
	// 这里应该实现与语言服务的通信
	return &HoverInfo{
		Contents: "悬停信息正在加载中...",
	}, nil
}

// formatCodeWithLanguageService 使用语言服务格式化代码
func (e *CodeEditorImpl) formatCodeWithLanguageService(ls *LanguageService, sessionID string, range_ *Range) (*TextEdit, error) {
	// 这里应该实现与语言服务的通信
	return nil, fmt.Errorf("格式化功能需要语言服务支持")
}