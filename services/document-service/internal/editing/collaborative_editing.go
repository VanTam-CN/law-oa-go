package editing

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"law-oa-go/internal/models"
)

// CollaborativeEditingService 协作编辑服务接口
type CollaborativeEditingService interface {
	// 创建协作会话
	CreateCollaborativeSession(ctx context.Context, sessionID string, userID string) (*CollaborativeSession, error)
	// 加入协作会话
	JoinCollaborativeSession(ctx context.Context, sessionID, userID, userName string) (*CollaborativeSession, error)
	// 离开协作会话
	LeaveCollaborativeSession(ctx context.Context, sessionID, userID string) error
	// 应用操作
	ApplyOperation(ctx context.Context, sessionID, userID string, operation *CRDTOperation) error
	// 获取文档状态
	GetDocumentState(ctx context.Context, sessionID string) (*DocumentState, error)
	// 获取活跃用户
	GetActiveUsers(ctx context.Context, sessionID string) ([]*CollaborativeUser, error)
	// 获取操作历史
	GetOperationHistory(ctx context.Context, sessionID string, limit int) ([]*CRDTOperation, error)
	// 冲突解决
	ResolveConflicts(ctx context.Context, sessionID string) (*ConflictResolution, error)
	// 同步文档状态
	SyncDocument(ctx context.Context, sessionID string) error
	// 销毁协作会话
	DestroyCollaborativeSession(ctx context.Context, sessionID string) error
}

// CollaborativeEditingServiceImpl 协作编辑服务实现
type CollaborativeEditingServiceImpl struct {
	// 存储服务
	storageService StorageService
	// 通知服务
	notifyService NotificationService

	// 协作会话管理
	sessions map[string]*CollaborativeSession
	mutex    sync.RWMutex

	// CRDT文档管理
	documents map[string]*CRDTDocument
	docMutex  sync.RWMutex

	// 操作历史管理
	histories map[string][]*CRDTOperation
	histMutex sync.RWMutex

	// WebSocket连接管理
	connections map[string][]*CollaborativeConnection
	connMutex   sync.RWMutex

	// 冲突检测器
	conflictDetector *ConflictDetector

	// 同步管理器
	syncManager *SyncManager
}

// CollaborativeSession 协作会话
type CollaborativeSession struct {
	ID          string                `json:"id"`
	DocumentURI string                `json:"document_uri"`
	Users       []*CollaborativeUser   `json:"users"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	IsActive    bool                  `json:"is_active"`
	Settings    *CollaborativeSettings `json:"settings"`
	mutex       sync.RWMutex
}

// CollaborativeUser 协作用户
type CollaborativeUser struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Avatar    string                 `json:"avatar"`
	Email     string                 `json:"email"`
	Color     string                 `json:"color"`
	JoinedAt   time.Time             `json:"joined_at"`
	LastSeen  time.Time             `json:"last_seen"`
	IsActive   bool                   `json:"is_active"`
	Selection *SelectionInfo          `json:"selection,omitempty"`
	Cursor    *CursorPosition       `json:"cursor,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// CollaborativeSettings 协作设置
type CollaborativeSettings struct {
	AutoSave        bool          `json:"auto_save"`
	AutoSaveDelay   time.Duration `json:"auto_save_delay"`
	ConflictPolicy  string        `json:"conflict_policy"` // "last-writer-wins", "merge", "manual"
	MaxUsers        int           `json:"max_users"`
	Permissions     []string      `json:"permissions"`
	EnableChat      bool          `json:"enable_chat"`
	EnableComments  bool          `json:"enable_comments"`
	EnableTracking  bool          `json:"enable_tracking"`
	PrivacyMode     string        `json:"privacy_mode"` // "public", "private", "restricted"
}

// CRDTDocument CRDT文档
type CRDTDocument struct {
	URI       string      `json:"uri"`
	Content   string      `json:"content"`
	Version   int         `json:"version"`
	Operations []*CRDTOperation `json:"operations"`
	Authors   map[string]string `json:"authors"`
	Timestamps map[int64]time.Time `json:"timestamps"`
	mutex     sync.RWMutex
}

// CRDTOperation CRDT操作
type CRDTOperation struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"` // "insert", "delete", "retain", "format"
	Position  int                    `json:"position"`
	Length    int                    `json:"length,omitempty"`
	Text      string                 `json:"text,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	AuthorID   string                 `json:"author_id"`
	AuthorName string                 `json:"author_name"`
	AuthorColor string                 `json:"author_color"`
	Timestamp time.Time              `json:"timestamp"`
	Version   int                    `json:"version"`
	ParentID   string                 `json:"parent_id,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// DocumentState 文档状态
type DocumentState struct {
	URI        string           `json:"uri"`
	Content    string           `json:"content"`
	Version    int              `json:"version"`
	Users      []*CollaborativeUser `json:"users"`
	Operations []*CRDTOperation   `json:"operations"`
	LastSync   time.Time         `json:"last_sync"`
	mutex      sync.RWMutex
}

// ConflictResolution 冲突解决
type ConflictResolution struct {
	Conflicts  []*Conflict          `json:"conflicts"`
	Resolution string                 `json:"resolution"` // "auto", "manual", "merged"
	Timestamp  time.Time              `json:"timestamp"`
	AuthorID   string                 `json:"author_id"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Conflict 冲突
type Conflict struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"` // "content", "format", "version"
	Operations   []*CRDTOperation `json:"operations"`
	Description  string      `json:"description"`
	Severity    int         `json:"severity"` // 1: low, 2: medium, 3: high
	Resolved    bool        `json:"resolved"`
	Resolver    string      `json:"resolver,omitempty"`
}

// CollaborativeConnection 协作连接
type CollaborativeConnection struct {
	UserID      string        `json:"user_id"`
	SessionID   string        `json:"session_id"`
	WebSocket   interface{}   `json:"websocket"`
	LastPing    time.Time     `json:"last_ping"`
	IsActive    bool          `json:"is_active"`
	Subscriptions []string     `json:"subscriptions"`
	mutex       sync.RWMutex
}

// ConflictDetector 冲突检测器
type ConflictDetector struct {
	// 冲突检测规则
	rules []ConflictRule
	mutex sync.RWMutex
}

// ConflictRule 冲突规则
type ConflictRule struct {
	Type     string        `json:"type"`
	Checker  func([]*CRDTOperation) []*Conflict
	Priority int           `json:"priority"`
	Enabled  bool          `json:"enabled"`
}

// SyncManager 同步管理器
type SyncManager struct {
	syncInterval time.Duration
	syncBuffer   map[string]*SyncBuffer
	mutex        sync.RWMutex
}

// SyncBuffer 同步缓冲区
type SyncBuffer struct {
	SessionID string
	Operations []*CRDTOperation
	Timestamp  time.Time
	Flushed    bool
	mutex      sync.RWMutex
}

// CollaborativeEvent 协作事件
type CollaborativeEvent struct {
	Type      string                 `json:"type"`
	SessionID string                 `json:"session_id"`
	UserID    string                 `json:"user_id"`
	Data      interface{}            `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewCollaborativeEditingService 创建协作编辑服务
func NewCollaborativeEditingService(
	storageService StorageService,
	notifyService NotificationService,
) CollaborativeEditingService {
	service := &CollaborativeEditingServiceImpl{
		storageService:   storageService,
		notifyService:    notifyService,
		sessions:          make(map[string]*CollaborativeSession),
		documents:         make(map[string]*CRDTDocument),
		histories:         make(map[string][]*CRDTOperation),
		connections:       make(map[string][]*CollaborativeConnection),
		conflictDetector:  NewConflictDetector(),
		syncManager:       NewSyncManager(),
	}

	// 启动后台同步
	go service.startBackgroundSync()

	return service
}

// CreateCollaborativeSession 创建协作会话
func (s *CollaborativeEditingServiceImpl) CreateCollaborativeSession(
	ctx context.Context,
	sessionID string,
	userID string,
) (*CollaborativeSession, error) {
	if sessionID == "" || userID == "" {
		return nil, fmt.Errorf("会话ID和用户ID不能为空")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查会话是否已存在
	if _, exists := s.sessions[sessionID]; exists {
		return nil, fmt.Errorf("协作会话已存在")
	}

	// 创建新会话
	session := &CollaborativeSession{
		ID:          sessionID,
		Users:       make([]*CollaborativeUser, 0),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		IsActive:    true,
		Settings:    s.getDefaultCollaborativeSettings(),
	}

	s.sessions[sessionID] = session

	// 创建文档
	s.docMutex.Lock()
	s.documents[sessionID] = &CRDTDocument{
		URI:         sessionID,
		Content:     "",
		Version:     1,
		Operations:  make([]*CRDTOperation, 0),
		Authors:     make(map[string]string),
		Timestamps:  make(map[int64]time.Time),
	}
	s.docMutex.Unlock()

	// 创建操作历史
	s.histMutex.Lock()
	s.histories[sessionID] = make([]*CRDTOperation, 0)
	s.histMutex.Unlock()

	// 创建连接管理
	s.connMutex.Lock()
	s.connections[sessionID] = make([]*CollaborativeConnection, 0)
	s.connMutex.Unlock()

	// 创建同步缓冲区
	s.syncManager.mutex.Lock()
	s.syncManager.syncBuffer[sessionID] = &SyncBuffer{
		SessionID:  sessionID,
		Operations: make([]*CRDTOperation, 0),
		Timestamp:  time.Now(),
		Flushed:    false,
	}
	s.syncManager.mutex.Unlock()

	return session, nil
}

// JoinCollaborativeSession 加入协作会话
func (s *CollaborativeEditingServiceImpl) JoinCollaborativeSession(
	ctx context.Context,
	sessionID,
	userID,
	userName string,
) (*CollaborativeSession, error) {
	if sessionID == "" || userID == "" {
		return nil, fmt.Errorf("会话ID和用户ID不能为空")
	}

	s.mutex.RLock()
	session, exists := s.sessions[sessionID]
	s.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("协作会话不存在")
	}

	session.mutex.Lock()
	defer session.mutex.Unlock()

	// 检查用户是否已在会话中
	for _, user := range session.Users {
		if user.ID == userID {
			user.IsActive = true
			user.LastSeen = time.Now()
			return session, nil
		}
	}

	// 检查用户数量限制
	if len(session.Users) >= session.Settings.MaxUsers {
		return nil, fmt.Errorf("协作会话用户数量已达到上限")
	}

	// 创建新用户
	user := &CollaborativeUser{
		ID:        userID,
		Name:      userName,
		Avatar:    s.generateUserAvatar(userID),
		Color:     s.generateUserColor(userID),
		JoinedAt:   time.Now(),
		LastSeen:  time.Now(),
		IsActive:   true,
		Metadata:  make(map[string]interface{}),
	}

	session.Users = append(session.Users, user)
	session.UpdatedAt = time.Now()

	// 发送用户加入事件
	s.broadcastEvent(sessionID, &CollaborativeEvent{
		Type:      "user_joined",
		SessionID: sessionID,
		UserID:    userID,
		Data:      user,
		Timestamp: time.Now(),
	})

	return session, nil
}

// LeaveCollaborativeSession 离开协作会话
func (s *CollaborativeEditingServiceImpl) LeaveCollaborativeSession(
	ctx context.Context,
	sessionID,
	userID string,
) error {
	if sessionID == "" || userID == "" {
		return fmt.Errorf("会话ID和用户ID不能为空")
	}

	s.mutex.RLock()
	session, exists := s.sessions[sessionID]
	s.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("协作会话不存在")
	}

	session.mutex.Lock()
	defer session.mutex.Unlock()

	// 标记用户为非活跃
	for i, user := range session.Users {
		if user.ID == userID {
			user.IsActive = false
			user.LastSeen = time.Now()
			// 从用户列表中移除
			session.Users = append(session.Users[:i], session.Users[i+1:]...)
			break
		}
	}

	session.UpdatedAt = time.Now()

	// 发送用户离开事件
	s.broadcastEvent(sessionID, &CollaborativeEvent{
		Type:      "user_left",
		SessionID: sessionID,
		UserID:    userID,
		Timestamp: time.Now(),
	})

	return nil
}

// ApplyOperation 应用操作
func (s *CollaborativeEditingServiceImpl) ApplyOperation(
	ctx context.Context,
	sessionID,
	userID string,
	operation *CRDTOperation,
) error {
	if sessionID == "" || userID == "" || operation == nil {
		return fmt.Errorf("会话ID、用户ID和操作不能为空")
	}

	s.mutex.RLock()
	session, exists := s.sessions[sessionID]
	s.mutex.RUnlock()

	if !exists || !session.IsActive {
		return fmt.Errorf("协作会话不存在或已关闭")
	}

	// 设置操作信息
	operation.ID = s.generateOperationID()
	operation.AuthorID = userID
	operation.AuthorName = s.getUserName(session, userID)
	operation.AuthorColor = s.getUserColor(session, userID)
	operation.Timestamp = time.Now()

	// 获取文档
	s.docMutex.RLock()
	doc, docExists := s.documents[sessionID]
	s.docMutex.RUnlock()

	if !docExists {
		return fmt.Errorf("文档不存在")
	}

	// 应用操作到文档
	doc.mutex.Lock()
	defer doc.mutex.Unlock()

	// 执行CRDT操作转换
	updatedContent, err := s.applyCRDTOperation(doc.Content, operation)
	if err != nil {
		return fmt.Errorf("应用CRDT操作失败: %w", err)
	}

	doc.Content = updatedContent
	doc.Version++
	doc.Operations = append(doc.Operations, operation)
	doc.Authors[operation.AuthorID] = operation.AuthorName
	doc.Timestamps[int64(doc.Version)] = operation.Timestamp

	// 添加到操作历史
	s.histMutex.Lock()
	s.histories[sessionID] = append(s.histories[sessionID], operation)
	s.histMutex.Unlock()

	// 检测冲突
	conflicts := s.conflictDetector.DetectConflicts(doc.Operations)
	if len(conflicts) > 0 {
		// 尝试自动解决冲突
		resolution := s.resolveConflicts(doc, conflicts)
		if resolution != nil {
			doc.Content = resolution.Content
			doc.Version++
		}
	}

	// 广播操作事件
	s.broadcastOperation(sessionID, operation)

	return nil
}

// GetDocumentState 获取文档状态
func (s *CollaborativeEditingServiceImpl) GetDocumentState(
	ctx context.Context,
	sessionID string,
) (*DocumentState, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("会话ID不能为空")
	}

	s.mutex.RLock()
	session, exists := s.sessions[sessionID]
	s.mutex.RUnlock()

	if !exists || !session.IsActive {
		return nil, fmt.Errorf("协作会话不存在或已关闭")
	}

	s.docMutex.RLock()
	doc, docExists := s.documents[sessionID]
	s.docMutex.RUnlock()

	if !docExists {
		return nil, fmt.Errorf("文档不存在")
	}

	doc.mutex.RLock()
	defer doc.mutex.RUnlock()

	// 复制用户列表
	users := make([]*CollaborativeUser, len(session.Users))
	copy(users, session.Users)

	// 复制操作列表
	operations := make([]*CRDTOperation, len(doc.Operations))
	copy(operations, doc.Operations)

	state := &DocumentState{
		URI:        doc.URI,
		Content:    doc.Content,
		Version:    doc.Version,
		Users:      users,
		Operations: operations,
		LastSync:   time.Now(),
	}

	return state, nil
}

// GetActiveUsers 获取活跃用户
func (s *CollaborativeEditingServiceImpl) GetActiveUsers(
	ctx context.Context,
	sessionID string,
) ([]*CollaborativeUser, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("会话ID不能为空")
	}

	s.mutex.RLock()
	session, exists := s.sessions[sessionID]
	s.mutex.RUnlock()

	if !exists || !session.IsActive {
		return nil, fmt.Errorf("协作会话不存在或已关闭")
	}

	session.mutex.RLock()
	defer session.mutex.RUnlock()

	// 过滤活跃用户
	activeUsers := make([]*CollaborativeUser, 0)
	for _, user := range session.Users {
		if user.IsActive {
			activeUsers = append(activeUsers, user)
		}
	}

	return activeUsers, nil
}

// GetOperationHistory 获取操作历史
func (s *CollaborativeEditingServiceImpl) GetOperationHistory(
	ctx context.Context,
	sessionID string,
	limit int,
) ([]*CRDTOperation, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("会话ID不能为空")
	}

	s.histMutex.RLock()
	history, exists := s.histories[sessionID]
	s.histMutex.RUnlock()

	if !exists {
		return make([]*CRDTOperation, 0), nil
	}

	// 返回最近的操作
	if limit > 0 && len(history) > limit {
		start := len(history) - limit
		return history[start:], nil
	}

	return history, nil
}

// ResolveConflicts 解决冲突
func (s *CollaborativeEditingServiceImpl) ResolveConflicts(
	ctx context.Context,
	sessionID string,
) (*ConflictResolution, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("会话ID不能为空")
	}

	s.docMutex.RLock()
	doc, docExists := s.documents[sessionID]
	s.docMutex.RUnlock()

	if !docExists {
		return nil, fmt.Errorf("文档不存在")
	}

	doc.mutex.RLock()
	defer doc.mutex.RUnlock()

	// 检测冲突
	conflicts := s.conflictDetector.DetectConflicts(doc.Operations)
	if len(conflicts) == 0 {
		return &ConflictResolution{
			Conflicts:  make([]*Conflict, 0),
			Resolution: "no_conflicts",
			Timestamp:  time.Now(),
		}, nil
	}

	// 解决冲突
	resolution := s.resolveConflicts(doc, conflicts)

	// 广播冲突解决事件
	s.broadcastEvent(sessionID, &CollaborativeEvent{
		Type:      "conflicts_resolved",
		SessionID: sessionID,
		Data:      resolution,
		Timestamp: time.Now(),
	})

	return resolution, nil
}

// SyncDocument 同步文档
func (s *CollaborativeEditingServiceImpl) SyncDocument(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("会话ID不能为空")
	}

	s.syncManager.mutex.RLock()
	buffer, exists := s.syncManager.syncBuffer[sessionID]
	s.syncManager.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("同步缓冲区不存在")
	}

	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()

	if len(buffer.Operations) == 0 {
		return nil
	}

	// 执行同步
	doc := &CRDTDocument{
		URI:       buffer.SessionID,
		Content:   "", // 将从缓冲区重建
		Version:   1,
		Operations: make([]*CRDTOperation, 0),
		Authors:   make(map[string]string),
		Timestamps: make(map[int64]time.Time),
	}

	// 应用所有操作
	for _, op := range buffer.Operations {
		content, err := s.applyCRDTOperation(doc.Content, op)
		if err != nil {
			return fmt.Errorf("同步操作失败: %w", err)
		}
		doc.Content = content
		doc.Operations = append(doc.Operations, op)
		doc.Authors[op.AuthorID] = op.AuthorName
		doc.Timestamps[int64(doc.Version)] = op.Timestamp
		doc.Version++
	}

	// 更新文档
	s.docMutex.Lock()
	s.documents[sessionID] = doc
	s.docMutex.Unlock()

	// 清空缓冲区
	buffer.Operations = make([]*CRDTOperation, 0)
	buffer.Timestamp = time.Now()
	buffer.Flushed = true

	return nil
}

// DestroyCollaborativeSession 销毁协作会话
func (s *CollaborativeEditingServiceImpl) DestroyCollaborativeSession(
	ctx context.Context,
	sessionID string,
) error {
	if sessionID == "" {
		return fmt.Errorf("会话ID不能为空")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 标记会话为非活跃
	if session, exists := s.sessions[sessionID]; exists {
		session.IsActive = false
		session.UpdatedAt = time.Now()

		// 断开所有连接
		s.connMutex.RLock()
	connections := s.connections[sessionID]
		s.connMutex.RUnlock()

		for _, conn := range connections {
			conn.mutex.Lock()
			conn.IsActive = false
			conn.mutex.Unlock()
		}
	}

	// 发送会话结束事件
	s.broadcastEvent(sessionID, &CollaborativeEvent{
		Type:      "session_ended",
		SessionID: sessionID,
		Timestamp: time.Now(),
	})

	// 延迟清理资源
	go func() {
		time.Sleep(10 * time.Minute)
		s.mutex.Lock()
		delete(s.sessions, sessionID)
		s.docMutex.Lock()
		delete(s.documents, sessionID)
		s.histMutex.Lock()
		delete(s.histories, sessionID)
		s.connMutex.Lock()
		delete(s.connections, sessionID)
		s.syncManager.mutex.Lock()
		delete(s.syncManager.syncBuffer, sessionID)
		s.syncManager.mutex.Unlock()
		s.mutex.Unlock()
	}()

	return nil
}

// 辅助方法

// getDefaultCollaborativeSettings 获取默认协作设置
func (s *CollaborativeEditingServiceImpl) getDefaultCollaborativeSettings() *CollaborativeSettings {
	return &CollaborativeSettings{
	AutoSave:       true,
		AutoSaveDelay:  5 * time.Second,
	ConflictPolicy:  "merge",
	MaxUsers:        10,
	Permissions:    []string{"read", "write", "comment"},
	EnableChat:      true,
		EnableComments:  true,
		EnableTracking:  true,
		PrivacyMode:     "private",
	}
}

// generateUserAvatar 生成用户头像
func (s *CollaborativeEditingServiceImpl) generateUserAvatar(userID string) string {
	// 简化实现：返回Gravatar URL或默认头像
	avatarColors := []string{"#FF6B6B", "#4ECDC4", "#45B7D1", "#96CEB4", "#FECA57"}
	hash := 0
	for _, char := range userID {
		hash = hash*31 + int(char)
	}
	color := avatarColors[hash%len(avatarColors)]

	return fmt.Sprintf("data:image/svg+xml,%s",
		fmt.Sprintf(`<svg width="40" height="40" xmlns="http://www.w3.org/2000/svg"><rect width="40" height="40" fill="%s"/><text x="20" y="25" text-anchor="middle" fill="white" font-family="Arial" font-size="14">%s</text></svg>`, color, userID[:2]))
}

// generateUserColor 生成用户颜色
func (s *CollaborativeEditingServiceImpl) generateUserColor(userID string) string {
	colors := []string{
		"#FF6B6B", "#4ECDC4", "#45B7D1", "#96CEB4", "#FECA57",
		"#48C9B0", "#AF7AC5", "#F8B739", "#EC7063", "#5DADE2",
		"#58D68D", "#F4D03F", "#BB8FCE", "#85C1E2", "#F8C471",
	}
	hash := 0
	for _, char := range userID {
		hash = hash*31 + int(char)
	}
	return colors[hash%len(colors)]
}

// generateOperationID 生成操作ID
func (s *CollaborativeEditingServiceImpl) generateOperationID() string {
	return fmt.Sprintf("op_%d_%d", time.Now().UnixNano(), time.Now().Nanosecond())
}

// getUserName 获取用户名
func (s *CollaborativeEditingServiceImpl) getUserName(session *CollaborativeSession, userID string) string {
	for _, user := range session.Users {
		if user.ID == userID {
			return user.Name
		}
	}
	return "Unknown User"
}

// getUserColor 获取用户颜色
func (s *CollaborativeEditingServiceImpl) getUserColor(session *CollaborativeSession, userID string) string {
	for _, user := range session.Users {
		if user.ID == userID {
			return user.Color
		}
	}
	return "#808080"
}

// applyCRDTOperation 应用CRDT操作
func (s *CollaborativeEditingServiceImpl) applyCRDTOperation(
	content string,
	operation *CRDTOperation,
) (string, error) {
	switch operation.Type {
	case "insert":
		if operation.Position < 0 || operation.Position > len(content) {
			operation.Position = len(content)
		}
		return content[:operation.Position] + operation.Text + content[operation.Position:], nil
	case "delete":
		if operation.Position < 0 || operation.Position >= len(content) {
			return content, nil
		}
		endPos := operation.Position + operation.Length
		if endPos > len(content) {
			endPos = len(content)
		}
		return content[:operation.Position] + content[endPos:], nil
	case "retain":
		if operation.Position < 0 || operation.Position > len(content) {
			return content, nil
		}
		// retain操作不改变内容
		return content, nil
	case "format":
		// 格式化操作需要解析和应用格式化规则
		// 这里简化处理，实际实现会更复杂
		return content, nil
	default:
		return content, fmt.Errorf("不支持的操作类型: %s", operation.Type)
	}
}

// broadcastEvent 广播事件
func (s *CollaborativeEditingServiceImpl) broadcastEvent(sessionID string, event *CollaborativeEvent) {
	s.connMutex.RLock()
	connections := s.connections[sessionID]
	s.connMutex.RUnlock()

	eventData, _ := json.Marshal(event)
	message := string(eventData)

	for _, conn := range connections {
		conn.mutex.RLock()
		if conn.IsActive {
			// 这里应该通过WebSocket发送消息
			// 简化实现
			_ = message
		}
		conn.mutex.RUnlock()
	}
}

// broadcastOperation 广播操作
func (s *CollaborativeEditingServiceImpl) broadcastOperation(sessionID string, operation *CRDTOperation) {
	event := &CollaborativeEvent{
		Type:      "operation_applied",
		SessionID: sessionID,
		UserID:    operation.AuthorID,
		Data:      operation,
		Timestamp: time.Now(),
	}

	s.broadcastEvent(sessionID, event)
}

// resolveConflicts 解决冲突
func (s *CollaborativeEditingServiceImpl) resolveConflicts(
	doc *CRDTDocument,
	conflicts []*Conflict,
) *ConflictResolution {
	// 简化的冲突解决实现
	resolution := &ConflictResolution{
		Conflicts:  conflicts,
		Resolution: "auto",
		Timestamp:  time.Now(),
	}

	// 根据冲突策略解决
	// 这里会根据设置选择不同的解决策略
	return resolution
}

// startBackgroundSync 启动后台同步
func (s *CollaborativeEditingServiceImpl) startBackgroundSync() {
	ticker := time.NewTicker(s.syncManager.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.performSync()
		}
	}
}

// performSync 执行同步
func (s *CollaborativeEditingServiceImpl) performSync() {
	s.syncManager.mutex.RLock()
	buffers := s.syncManager.syncBuffer
	s.syncManager.mutex.RUnlock()

	for sessionID, buffer := range buffers {
		if len(buffer.Operations) > 0 && !buffer.Flushed {
			// 这里应该异步执行同步
			go func(sid string) {
				_ = s.SyncDocument(context.Background(), sid)
			}(sessionID)
		}
	}
}

// NewConflictDetector 创建冲突检测器
func NewConflictDetector() *ConflictDetector {
	detector := &ConflictDetector{
		rules: make([]ConflictRule, 0),
	}

	// 添加默认冲突检测规则
	detector.addRule(ConflictRule{
		Type:     "concurrent_insert",
		Checker:  s.checkConcurrentInsert,
		Priority: 1,
		Enabled:  true,
	})

	detector.addRule(ConflictRule{
		Type:     "concurrent_delete",
		Checker:  s.checkConcurrentDelete,
		Priority: 1,
		Enabled:  true,
	})

	return detector
}

// addRule 添加冲突规则
func (d *ConflictDetector) addRule(rule ConflictRule) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.rules = append(d.rules, rule)
}

// DetectConflicts 检测冲突
func (d *ConflictDetector) DetectConflicts(operations []*CRDTOperation) []*Conflict {
	d.mutex.RLock()
	rules := d.rules
	d.mutex.RUnlock()

	var allConflicts []*Conflict

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		conflicts := rule.Checker(operations)
		allConflicts = append(allConflicts, conflicts...)
	}

	// 按优先级排序冲突
	// 简化实现
	return allConflicts
}

// checkConcurrentInsert 检查并发插入冲突
func (s *CollaborativeEditingServiceImpl) checkConcurrentInsert(operations []*CRDTOperation) []*Conflict {
	// 简化的并发插入检测
	// 实际实现需要更复杂的时间和位置分析
	return make([]*Conflict, 0)
}

// checkConcurrentDelete 检查并发删除冲突
func (s *CollaborativeEditingServiceImpl) checkConcurrentDelete(operations []*CRDTOperation) []*Conflict {
	// 简化的并发删除检测
	// 实际实现需要更复杂的范围分析
	return make([]*Conflict, 0)
}

// NewSyncManager 创建同步管理器
func NewSyncManager() *SyncManager {
	return &SyncManager{
		syncInterval: 1 * time.Second,
		syncBuffer:   make(map[string]*SyncBuffer),
	}
}