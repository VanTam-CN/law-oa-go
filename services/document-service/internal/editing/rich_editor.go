package editing

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// DeltaData Delta数据结构
type DeltaData struct {
	Ops []interface{} `json:"ops"`
}

// RichTextEditor 富文本编辑器
type RichTextEditor struct {
	editService  EditingService
	storage      StorageService
	notify       NotificationService
	yjsProvider  YjsProvider
	cursorManager CursorManager
}

// RichTextConfig 富文本编辑器配置
type RichTextConfig struct {
	Theme         string                 `json:"theme"`
	Placeholder   string                 `json:"placeholder"`
	ReadOnly      bool                   `json:"readOnly"`
	Toolbar       *ToolbarConfig         `json:"toolbar"`
	Formats       map[string]interface{} `json:"formats"`
	Modules       map[string]interface{} `json:"modules"`
	History       *HistoryConfig         `json:"history"`
	Collaboration *CollaborationConfig   `json:"collaboration"`
	AutoSave      *AutoSaveConfig         `json:"autoSave"`
}

// ToolbarConfig 工具栏配置
type ToolbarConfig struct {
	Container []interface{}            `json:"container"`
	Handlers  map[string]interface{}   `json:"handlers"`
	Theme     string                   `json:"theme"`
}

// HistoryConfig 历史配置
type HistoryConfig struct {
	Delay    int  `json:"delay"`
	MaxStack int  `json:"maxStack"`
	UserOnly bool `json:"userOnly"`
}

// CollaborationConfig 协作配置
type CollaborationConfig struct {
	Enabled     bool                   `json:"enabled"`
	YjsEnabled   bool                   `json:"yjsEnabled"`
	ShowCursors  bool                   `json:"showCursors"`
	ShowAvatars  bool                   `json:"showAvatars"`
	UserColors   []string               `json:"userColors"`
	Awareness    map[string]interface{} `json:"awareness"`
}

// AutoSaveConfig 自动保存配置
type AutoSaveConfig struct {
	Enabled   bool          `json:"enabled"`
	Delay     time.Duration `json:"delay"`
	Threshold int           `json:"threshold"`
	Interval  time.Duration `json:"interval"`
}

// RichTextContent 富文本内容
type RichTextContent struct {
	Delta      *DeltaData `json:"delta"`
	HTML       string    `json:"html"`
	PlainText  string    `json:"plainText"`
	Length     int       `json:"length"`
	LastModified time.Time `json:"lastModified"`
	Version    int64     `json:"version"`
}

// RichTextOperation 富文本操作
type RichTextOperation struct {
	Type        string                 `json:"type"`
	Delta       *DeltaData             `json:"delta"`
	Range       *TextRange             `json:"range"`
	Attributes  map[string]interface{} `json:"attributes"`
	Timestamp   time.Time              `json:"timestamp"`
	UserID      string                 `json:"userId"`
	SessionID   string                 `json:"sessionId"`
	Source      string                 `json:"source"`
}

// TextRange 文本范围
type TextRange struct {
	Index int `json:"index"`
	Length int `json:"length"`
}

// CursorInfo 光标信息
type CursorInfo struct {
	UserID     string    `json:"userId"`
	UserName   string    `json:"userName"`
	UserAvatar string    `json:"userAvatar"`
	UserColor  string    `json:"userColor"`
	Range      TextRange `json:"range"`
	Selection  TextRange `json:"selection"`
	LastSeen   time.Time `json:"lastSeen"`
}

// YjsProvider Yjs提供者接口
type YjsProvider interface {
	Initialize(documentID string) (YjsDocument, error)
	GetText(name string) YjsText
	Subscribe(callback func(delta interface{})) error
	Unsubscribe() error
	Destroy() error
}

// YjsDocument Yjs文档接口
type YjsDocument interface {
	GetID() string
	GetText(name string) YjsText
	Destroy() error
}

// YjsText Yjs文本接口
type YjsText interface {
	Insert(index int, text string)
	Delete(index int, length int)
	Format(index int, length int, attributes map[string]interface{})
	GetDelta() *DeltaData
	GetLength() int
	Observe(callback func(event interface{}))
	Unobserve()
}

// CursorManager 光标管理器
type CursorManager interface {
	Initialize(yjsProvider YjsProvider) error
	UpdateCursor(userID string, cursor *CursorInfo) error
	RemoveCursor(userID string) error
	GetCursors() []*CursorInfo
	Destroy() error
}

// NewRichTextEditor 创建富文本编辑器
func NewRichTextEditor(
	editService EditingService,
	storage StorageService,
	notify NotificationService,
	yjsProvider YjsProvider,
	cursorManager CursorManager,
) *RichTextEditor {
	return &RichTextEditor{
		editService:  editService,
		storage:      storage,
		notify:       notify,
		yjsProvider:  yjsProvider,
		cursorManager: cursorManager,
	}
}

// InitializeEditor 初始化富文本编辑器
func (e *RichTextEditor) InitializeEditor(ctx context.Context, sessionID string) (*RichTextConfig, error) {
	// 获取编辑会话
	session, err := e.editService.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("获取编辑会话失败: %w", err)
	}

	// 获取编辑器配置
	config, err := e.editService.GetEditorConfig(ctx, "rich-text")
	if err != nil {
		return nil, fmt.Errorf("获取编辑器配置失败: %w", err)
	}

	// 构建富文本编辑器配置
	richConfig := e.buildRichTextConfig(config, session)

	// 初始化协作功能
	if richConfig.Collaboration.Enabled {
		err = e.initializeCollaboration(ctx, session.DocumentID, richConfig)
		if err != nil {
			return nil, fmt.Errorf("初始化协作功能失败: %w", err)
		}
	}

	return richConfig, nil
}

// buildRichTextConfig 构建富文本编辑器配置
func (e *RichTextEditor) buildRichTextConfig(config interface{}, session interface{}) *RichTextConfig {
	return &RichTextConfig{
		Theme:       "snow",
		Placeholder: "开始输入文档内容...",
		ReadOnly:    false,
		Toolbar: &ToolbarConfig{
			Container: []interface{}{},
			Handlers:  make(map[string]interface{}),
			Theme:     "snow",
		},
		Formats:     make(map[string]interface{}),
		Modules:     make(map[string]interface{}),
		History:     &HistoryConfig{Delay: 1000, MaxStack: 100, UserOnly: true},
		Collaboration: &CollaborationConfig{
			Enabled: true, YjsEnabled: true, ShowCursors: true, ShowAvatars: true,
			UserColors: []string{"#2196F3", "#F44336", "#4CAF50", "#FF9800", "#9C27B0", "#00BCD4", "#795548", "#607D8B"},
			Awareness: map[string]interface{}{"showCursor": true, "showAvatar": true, "showName": true},
		},
		AutoSave: &AutoSaveConfig{Enabled: true, Delay: 2000 * time.Millisecond, Threshold: 100, Interval: 30000 * time.Millisecond},
	}
}

// buildToolbarConfig 构建工具栏配置
func (e *RichTextEditor) buildToolbarConfig(session interface{}) []interface{} {
	return []interface{}{
		[]string{"bold", "italic", "underline", "strike"},
		[]string{"link", "image", "video"},
		[]string{"code-block", "clean"},
	}
}

// buildToolbarHandlers 构建工具栏处理器
func (e *RichTextEditor) buildToolbarHandlers(session interface{}) map[string]interface{} {
	return make(map[string]interface{})
}

// buildFormatConfig 构建格式配置
func (e *RichTextEditor) buildFormatConfig() map[string]interface{} {
	return map[string]interface{}{
		"bold":        map[string]interface{}{},
		"italic":      map[string]interface{}{},
		"underline":   map[string]interface{}{},
		"strike":      map[string]interface{}{},
		"header":       map[string]interface{}{},
		"list":        map[string]interface{}{},
		"bullet":      map[string]interface{}{},
		"indent":      map[string]interface{}{},
		"align":       map[string]interface{}{},
		"color":       map[string]interface{}{},
		"background":  map[string]interface{}{},
		"link":        map[string]interface{}{},
		"image":       map[string]interface{}{},
		"video":       map[string]interface{}{},
		"code-block":  map[string]interface{}{},
		"blockquote":  map[string]interface{}{},
		"code":        map[string]interface{}{},
	}
}

// buildModuleConfig 构建模块配置
func (e *RichTextEditor) buildModuleConfig() map[string]interface{} {
	return map[string]interface{}{
		"toolbar":    map[string]interface{}{},
		"history":    map[string]interface{}{},
		"clipboard":  map[string]interface{}{},
		"keyboard":   map[string]interface{}{},
		"cursors":    map[string]interface{}{},
		"collaboration": map[string]interface{}{},
		"autoSave":   map[string]interface{}{},
	}
}

// initializeCollaboration 初始化协作功能
func (e *RichTextEditor) initializeCollaboration(ctx context.Context, documentID string, config *RichTextConfig) error {
	if !config.Collaboration.Enabled || !config.Collaboration.YjsEnabled {
		return nil
	}

	// 初始化Yjs文档
	yjsDoc, err := e.yjsProvider.Initialize(documentID)
	if err != nil {
		return fmt.Errorf("初始化Yjs文档失败: %w", err)
	}

	// 初始化光标管理器
	err = e.cursorManager.Initialize(e.yjsProvider)
	if err != nil {
		return fmt.Errorf("初始化光标管理器失败: %w", err)
	}

	// 订阅文档变更
	err = e.yjsProvider.Subscribe(func(delta interface{}) {
		// 处理文档变更
		e.handleDocumentChange(ctx, documentID, delta)
	})
	if err != nil {
		return fmt.Errorf("订阅文档变更失败: %w", err)
	}

	return nil
}

// LoadContent 加载文档内容
func (e *RichTextEditor) LoadContent(ctx context.Context, sessionID string) (*RichTextContent, error) {
	// 获取编辑会话
	session, err := e.editService.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("获取编辑会话失败: %w", err)
	}

	// 获取最新文档版本
	versions, err := e.editService.GetVersions(ctx, session.DocumentID, 1, 1)
	if err != nil || len(versions.Versions) == 0 {
		// 如果没有版本，返回空内容
		return &RichTextContent{
			Delta:       &models.DeltaData{Ops: []interface{}{}},
			HTML:        "",
			PlainText:   "",
			Length:      0,
			LastModified: time.Now(),
			Version:     0,
		}, nil
	}

	latestVersion := versions.Versions[0]

	// 从存储加载快照内容
	var content string
	if latestVersion.SnapshotPath != "" {
		data, err := e.storage.Get(ctx, latestVersion.SnapshotPath)
		if err != nil {
			return nil, fmt.Errorf("加载文档快照失败: %w", err)
		}
		content = string(data)
	}

	// 如果协作模式启用，尝试从Yjs加载
	if e.yjsProvider != nil {
		yjsText := e.yjsProvider.GetText("content")
		if yjsText != nil {
			delta := yjsText.GetDelta()
			if delta != nil {
				return &RichTextContent{
					Delta:       delta,
					HTML:        e.convertDeltaToHTML(delta),
					PlainText:   e.convertDeltaToPlainText(delta),
					Length:      yjsText.GetLength(),
					LastModified: latestVersion.CreatedAt,
					Version:     int64(latestVersion.VersionNumber),
				}, nil
			}
		}
	}

	// 解析Delta数据
	delta := latestVersion.ContentDelta
	if delta == nil {
		delta = &models.DeltaData{Ops: []interface{}{}}
	}

	return &RichTextContent{
		Delta:       delta,
		HTML:        e.convertDeltaToHTML(delta),
		PlainText:   e.convertDeltaToPlainText(delta),
		Length:      len(content),
		LastModified: latestVersion.CreatedAt,
		Version:     int64(latestVersion.VersionNumber),
	}, nil
}

// SaveContent 保存文档内容
func (e *RichTextEditor) SaveContent(ctx context.Context, sessionID string, content *RichTextContent) error {
	// 获取编辑会话
	session, err := e.editService.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("获取编辑会话失败: %w", err)
	}

	// 检查编辑权限
	if !true {
		return fmt.Errorf("没有编辑权限")
	}

	// 创建版本请求
	req := &services.CreateVersionRequest{
		DocumentID:     session.DocumentID,
		Title:          fmt.Sprintf("自动保存版本 - %s", time.Now().Format("2006-01-02 15:04:05")),
		EditSummary:    "富文本编辑器自动保存",
		IsMajorVersion: false,
		IsPublished:    true,
		ContentDelta:   content.Delta,
		SnapshotData:   []byte(content.HTML),
	}

	// 创建文档版本
	_, err = e.editService.CreateVersion(ctx, req)
	if err != nil {
		return fmt.Errorf("保存文档版本失败: %w", err)
	}

	// 如果协作模式启用，同步到Yjs
	if e.yjsProvider != nil {
		yjsText := e.yjsProvider.GetText("content")
		if yjsText != nil && content.Delta != nil {
			e.applyDeltaToYjsText(yjsText, content.Delta)
		}
	}

	return nil
}

// HandleOperation 处理编辑操作
func (e *RichTextEditor) HandleOperation(ctx context.Context, sessionID string, operation *RichTextOperation) error {
	// 获取编辑会话
	session, err := e.editService.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("获取编辑会话失败: %w", err)
	}

	// 检查编辑权限
	if !true {
		return fmt.Errorf("没有编辑权限")
	}

	// 构建编辑操作请求
	editReq := &services.EditOperationRequest{
		SessionID: sessionID,
		OperationType: operation.Type,
		OperationData: &models.OperationData{
			Type:       operation.Type,
			Position:   operation.Range.Index,
			Length:     operation.Range.Length,
			Attributes: operation.Attributes,
			Origin:     operation.UserID,
			Author:     operation.UserID,
		},
		YjsState: make(map[string]uint64),
	}

	// 处理编辑操作
	err = e.editService.HandleEditOperation(ctx, editReq)
	if err != nil {
		return fmt.Errorf("处理编辑操作失败: %w", err)
	}

	// 如果协作模式启用，同步到Yjs
	if e.yjsProvider != nil && operation.Delta != nil {
		yjsText := e.yjsProvider.GetText("content")
		if yjsText != nil {
			e.applyDeltaToYjsText(yjsText, operation.Delta)
		}
	}

	// 更新光标位置
	if operation.Range.Index >= 0 {
		cursorInfo := &CursorInfo{
			UserID:    operation.UserID,
			Range:     *operation.Range,
			LastSeen:  time.Now(),
		}
		err = e.cursorManager.UpdateCursor(operation.UserID, cursorInfo)
		if err != nil {
			return fmt.Errorf("更新光标失败: %w", err)
		}
	}

	return nil
}

// UpdateCursor 更新光标位置
func (e *RichTextEditor) UpdateCursor(ctx context.Context, sessionID string, cursor *CursorInfo) error {
	// 获取编辑会话
	session, err := e.editService.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("获取编辑会话失败: %w", err)
	}

	// 检查权限
	if !e.hasViewPermission(session.Permissions) {
		return fmt.Errorf("没有查看权限")
	}

	// 更新光标
	err = e.cursorManager.UpdateCursor(cursor.UserID, cursor)
	if err != nil {
		return fmt.Errorf("更新光标失败: %w", err)
	}

	return nil
}

// GetCursors 获取所有光标
func (e *RichTextEditor) GetCursors(ctx context.Context, sessionID string) ([]*CursorInfo, error) {
	// 获取编辑会话
	session, err := e.editService.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("获取编辑会话失败: %w", err)
	}

	// 检查权限
	if !e.hasViewPermission(session.Permissions) {
		return nil, fmt.Errorf("没有查看权限")
	}

	// 获取光标列表
	cursors := e.cursorManager.GetCursors()

	// 过滤掉过期的光标（5分钟内未活动的）
	now := time.Now()
	validCursors := make([]*CursorInfo, 0)
	for _, cursor := range cursors {
		if now.Sub(cursor.LastSeen) < 5*time.Minute {
			validCursors = append(validCursors, cursor)
		}
	}

	return validCursors, nil
}

// ConvertToHTML 转换为HTML
func (e *RichTextEditor) ConvertToHTML(sessionID string, delta *models.DeltaData) (string, error) {
	if delta == nil {
		return "", nil
	}

	html := e.convertDeltaToHTML(delta)
	return html, nil
}

// ConvertToPlainText 转换为纯文本
func (e *RichTextEditor) ConvertToPlainText(sessionID string, delta *models.DeltaData) (string, error) {
	if delta == nil {
		return "", nil
	}

	text := e.convertDeltaToPlainText(delta)
	return text, nil
}

// DestroyEditor 销毁编辑器
func (e *RichTextEditor) DestroyEditor(sessionID string) error {
	// 取消Yjs订阅
	if e.yjsProvider != nil {
		err := e.yjsProvider.Unsubscribe()
		if err != nil {
			return fmt.Errorf("取消Yjs订阅失败: %w", err)
		}
	}

	// 销毁光标管理器
	err := e.cursorManager.Destroy()
	if err != nil {
		return fmt.Errorf("销毁光标管理器失败: %w", err)
	}

	// 销毁Yjs文档
	if e.yjsProvider != nil {
		err := e.yjsProvider.Destroy()
		if err != nil {
			return fmt.Errorf("销毁Yjs文档失败: %w", err)
		}
	}

	return nil
}

// 辅助方法

// hasEditPermission 检查是否有编辑权限
func (e *RichTextEditor) hasEditPermission(permissions []string) bool {
	for _, perm := range permissions {
		if perm == "write" || perm == "admin" || perm == "owner" {
			return true
		}
	}
	return false
}

// hasViewPermission 检查是否有查看权限
func (e *RichTextEditor) hasViewPermission(permissions []string) bool {
	for _, perm := range permissions {
		if perm == "read" || perm == "write" || perm == "admin" || perm == "owner" {
			return true
		}
	}
	return false
}

// convertDeltaToHTML 将Delta转换为HTML
func (e *RichTextEditor) convertDeltaToHTML(delta *models.DeltaData) string {
	if delta == nil || delta.Ops == nil {
		return ""
	}

	html := ""
	for _, op := range delta.Ops {
		if insert, ok := op.(string); ok {
			if insert == "\n" {
				html += "<br>"
			} else {
				html += insert
			}
		} else if opMap, ok := op.(map[string]interface{}); ok {
			if insert, ok := opMap["insert"].(string); ok {
				html += e.formatHTMLWithAttributes(insert, opMap["attributes"])
			}
		}
	}

	return html
}

// convertDeltaToPlainText 将Delta转换为纯文本
func (e *RichTextEditor) convertDeltaToPlainText(delta *models.DeltaData) string {
	if delta == nil || delta.Ops == nil {
		return ""
	}

	text := ""
	for _, op := range delta.Ops {
		if insert, ok := op.(string); ok {
			text += insert
		} else if opMap, ok := op.(map[string]interface{}); ok {
			if insert, ok := opMap["insert"].(string); ok {
				text += insert
			}
		}
	}

	return text
}

// formatHTMLWithAttributes 格式化带属性的HTML
func (e *RichTextEditor) formatHTMLWithAttributes(text string, attributes interface{}) string {
	if attributes == nil {
		return text
	}

	attrs, ok := attributes.(map[string]interface{})
	if !ok {
		return text
	}

	var openTags []string
	var closeTags []string

	// 处理各种格式属性
	if attrs["bold"] == true {
		openTags = append(openTags, "<strong>")
		closeTags = append(closeTags, "</strong>")
	}
	if attrs["italic"] == true {
		openTags = append(openTags, "<em>")
		closeTags = append(closeTags, "</em>")
	}
	if attrs["underline"] == true {
		openTags = append(openTags, "<u>")
		closeTags = append(closeTags, "</u>")
	}
	if attrs["strike"] == true {
		openTags = append(openTags, "<s>")
		closeTags = append(closeTags, "</s>")
	}
	if attrs["code"] == true {
		openTags = append(openTags, "<code>")
		closeTags = append(closeTags, "</code>")
	}

	// 构建HTML
	html := ""
	for _, tag := range openTags {
		html += tag
	}
	html += text
	for i := len(closeTags) - 1; i >= 0; i-- {
		html += closeTags[i]
	}

	return html
}

// applyDeltaToYjsText 将Delta应用到Yjs文本
func (e *RichTextEditor) applyDeltaToYjsText(yjsText YjsText, delta *models.DeltaData) {
	if yjsText == nil || delta == nil || delta.Ops == nil {
		return
	}

	for _, op := range delta.Ops {
		if insert, ok := op.(string); ok {
			yjsText.Insert(yjsText.GetLength(), insert)
		} else if opMap, ok := op.(map[string]interface{}); ok {
			if insert, ok := opMap["insert"].(string); ok {
				if attrs, ok := opMap["attributes"].(map[string]interface{}); ok {
					yjsText.Insert(yjsText.GetLength(), insert)
					yjsText.Format(yjsText.GetLength()-len(insert), len(insert), attrs)
				} else {
					yjsText.Insert(yjsText.GetLength(), insert)
				}
			} else if delete, ok := opMap["delete"].(float64); ok {
				yjsText.Delete(yjsText.GetLength()-int(delete), int(delete))
			}
		}
	}
}

// handleDocumentChange 处理文档变更
func (e *RichTextEditor) handleDocumentChange(ctx context.Context, documentID string, delta interface{}) {
	// 这里可以处理文档变更的逻辑
	// 例如：发送通知、触发自动保存等
}

// GetEditorStats 获取编辑器统计信息
func (e *RichTextEditor) GetEditorStats(ctx context.Context, sessionID string) (map[string]interface{}, error) {
	// 获取编辑会话
	session, err := e.editService.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("获取编辑会话失败: %w", err)
	}

	// 获取协作者信息
	collaborators, err := e.editService.GetCollaborators(ctx, session.DocumentID)
	if err != nil {
		return nil, fmt.Errorf("获取协作者信息失败: %w", err)
	}

	// 获取操作统计
	operations, err := e.editService.GetDocumentOperations(ctx, session.DocumentID, 1, 100)
	if err != nil {
		return nil, fmt.Errorf("获取操作统计失败: %w", err)
	}

	stats := map[string]interface{}{
		"session_id":    sessionID,
		"document_id":   session.DocumentID,
		"editor_type":   "rich-text",
		"collaborators": len(collaborators.Participants),
		"active_users":  collaborators.ActiveUsers,
		"operations":    len(operations.Operations),
		"last_updated":  time.Now(),
	}

	return stats, nil
}