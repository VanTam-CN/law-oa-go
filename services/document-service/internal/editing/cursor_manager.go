package editing

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// NotificationService 通知服务接口
type NotificationService interface {
	SendNotification(sessionID string, message interface{}) error
}

// CursorManager 光标管理器接口
type CursorManager interface {
	// 用户光标管理
	TrackCursor(sessionID, userID string, position *CursorPosition) error
	UntrackCursor(sessionID, userID string) error
	GetCursors(sessionID string) ([]*UserCursor, error)
	GetUserCursor(sessionID, userID string) (*UserCursor, error)

	// 用户在线状态
	UpdateUserPresence(sessionID, userID string, presence *UserPresence) error
	GetUserPresence(sessionID, userID string) (*UserPresence, error)
	GetActiveUsers(sessionID string) ([]*UserPresence, error)

	// 光标冲突检测和解决
	DetectCursorConflicts(sessionID string) ([]*CursorConflict, error)
	ResolveCursorConflict(sessionID string, conflict *CursorConflict) error

	// 光标历史
	GetCursorHistory(sessionID, userID string, limit int) ([]*CursorPosition, error)

	// 清理
	CleanupInactiveCursors(sessionID string, inactiveDuration time.Duration) error
	RemoveSession(sessionID string) error

	// 统计信息
	GetCursorStats(sessionID string) (*CursorStats, error)
}

// CursorPosition 光标位置
type CursorPosition struct {
	Index     int       `json:"index"`     // 光标在文档中的位置
	Length    int       `json:"length"`    // 选择的长度（0表示只有光标）
	Timestamp time.Time `json:"timestamp"` // 更新时间
}

// UserCursor 用户光标
type UserCursor struct {
	UserID      string          `json:"user_id"`
	UserName    string          `json:"user_name"`
	UserAvatar  string          `json:"user_avatar"`
	Color       string          `json:"color"`       // 用户颜色
	Position    *CursorPosition `json:"position"`
	Selection   *TextSelection  `json:"selection"`   // 文本选择
	IsActive    bool            `json:"is_active"`   // 是否活跃
	LastSeen    time.Time       `json:"last_seen"`   // 最后活跃时间
}

// TextSelection 文本选择
type TextSelection struct {
	Start int    `json:"start"`
	End   int    `json:"end"`
	Text  string `json:"text"`
}

// UserPresence 用户在线状态
type UserPresence struct {
	UserID        string            `json:"user_id"`
	UserName      string            `json:"user_name"`
	UserAvatar    string            `json:"user_avatar"`
	Status        UserStatus        `json:"status"`         // 在线状态
	CurrentEditor string            `json:"current_editor"` // 当前编辑器
	Activity      string            `json:"activity"`       // 当前活动
	LastActivity  time.Time         `json:"last_activity"`  // 最后活动时间
	Metadata      map[string]string `json:"metadata"`       // 额外元数据
}

// UserStatus 用户状态枚举
type UserStatus string

const (
	UserStatusOnline      UserStatus = "online"
	UserStatusAway        UserStatus = "away"
	UserStatusBusy        UserStatus = "busy"
	UserStatusOffline     UserStatus = "offline"
	UserStatusInEditor    UserStatus = "in_editor"
	UserStatusViewing     UserStatus = "viewing"
	UserStatusEditing     UserStatus = "editing"
	UserStatusSelecting   UserStatus = "selecting"
)

// CursorConflict 光标冲突
type CursorConflict struct {
	ConflictID     string        `json:"conflict_id"`
	Users          []string      `json:"users"`           // 冲突的用户ID列表
	ConflictType   ConflictType  `json:"conflict_type"`   // 冲突类型
	ConflictRange  *TextRange    `json:"conflict_range"`  // 冲突范围
	DetectedAt     time.Time     `json:"detected_at"`
	Resolved       bool          `json:"resolved"`
	ResolvedBy     string        `json:"resolved_by"`     // 解决者
	Resolution     string        `json:"resolution"`      // 解决方案
}

// ConflictType 冲突类型
type ConflictType string

const (
	ConflictTypeOverlap       ConflictType = "overlap"        // 重叠冲突
	ConflictTypeConcurrent    ConflictType = "concurrent"     // 并发编辑冲突
	ConflictTypeRace          ConflictType = "race"           // 竞争条件
	ConflictTypeSelection     ConflictType = "selection"      // 选择冲突
)

// TextRange 文本范围
type TextRange struct {
	Start int    `json:"start"`
	End   int    `json:"end"`
	Text  string `json:"text"`
}

// CursorStats 光标统计信息
type CursorStats struct {
	SessionID       string    `json:"session_id"`
	ActiveUsers     int       `json:"active_users"`
	TotalCursors    int       `json:"total_cursors"`
	ConflictCount   int       `json:"conflict_count"`
	LastActivity    time.Time `json:"last_activity"`
	AvgSessionTime  float64   `json:"avg_session_time"` // 平均会话时长（分钟）
	PeakUsers       int       `json:"peak_users"`       // 峰值用户数
	PeakTime        time.Time `json:"peak_time"`        // 峰值时间
}

// CursorManagerImpl 光标管理器实现
type CursorManagerImpl struct {
	sessions map[string]*CursorSession
	mutex    sync.RWMutex
	logger   *logrus.Logger
	notify   NotificationService
}

// CursorSession 光标会话
type CursorSession struct {
	SessionID    string                    `json:"session_id"`
	DocumentID   string                    `json:"document_id"`
	Cursors      map[string]*UserCursor    `json:"cursors"`       // userID -> cursor
	Presences    map[string]*UserPresence  `json:"presences"`     // userID -> presence
	History      map[string][]*CursorPosition `json:"history"`    // userID -> positions
	Conflicts    []*CursorConflict         `json:"conflicts"`
	Stats        *CursorStats              `json:"stats"`
	CreatedAt    time.Time                 `json:"created_at"`
	LastActivity time.Time                 `json:"last_activity"`
	mutex        sync.RWMutex
}

// NewCursorManager 创建光标管理器
func NewCursorManager(logger *logrus.Logger, notifyService NotificationService) CursorManager {
	return &CursorManagerImpl{
		sessions: make(map[string]*CursorSession),
		logger:   logger,
		notify:   notifyService,
	}
}

// TrackCursor 跟踪用户光标
func (cm *CursorManagerImpl) TrackCursor(sessionID, userID string, position *CursorPosition) error {
	if sessionID == "" || userID == "" {
		return fmt.Errorf("会话ID和用户ID不能为空")
	}
	if position == nil {
		return fmt.Errorf("光标位置不能为空")
	}

	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 获取或创建会话
	session := cm.getOrCreateSession(sessionID)

	session.mutex.Lock()
	defer session.mutex.Unlock()

	// 获取用户信息
	userName := fmt.Sprintf("用户_%s", userID[:min(8, len(userID))])
	userColor := cm.getUserColor(userID)

	// 更新或创建光标
	cursor, exists := session.Cursors[userID]
	if exists {
		// 更新现有光标
		cursor.Position = position
		cursor.IsActive = true
		cursor.LastSeen = time.Now()
	} else {
		// 创建新光标
		cursor = &UserCursor{
			UserID:   userID,
			UserName: userName,
			Color:    userColor,
			Position: position,
			IsActive: true,
			LastSeen: time.Now(),
		}
		session.Cursors[userID] = cursor
	}

	// 更新历史记录
	session.History[userID] = append(session.History[userID], position)
	if len(session.History[userID]) > 100 { // 限制历史记录数量
		session.History[userID] = session.History[userID][1:]
	}

	// 更新统计信息
	session.Stats.ActiveUsers = len(session.Cursors)
	session.Stats.TotalCursors = len(session.Cursors)
	session.Stats.LastActivity = time.Now()
	if session.Stats.ActiveUsers > session.Stats.PeakUsers {
		session.Stats.PeakUsers = session.Stats.ActiveUsers
		session.Stats.PeakTime = time.Now()
	}

	session.LastActivity = time.Now()

	// 检测光标冲突
	go cm.detectAndNotifyConflicts(sessionID)

	// 发送光标更新通知
	go cm.notifyCursorUpdate(sessionID, cursor)

	cm.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"user_id":    userID,
		"position":   position.Index,
	}).Debug("用户光标已更新")

	return nil
}

// UntrackCursor 取消跟踪用户光标
func (cm *CursorManagerImpl) UntrackCursor(sessionID, userID string) error {
	if sessionID == "" || userID == "" {
		return fmt.Errorf("会话ID和用户ID不能为空")
	}

	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	session, exists := cm.sessions[sessionID]
	if !exists {
		return nil // 会话不存在，无需处理
	}

	session.mutex.Lock()
	defer session.mutex.Unlock()

	// 删除光标
	delete(session.Cursors, userID)

	// 更新在线状态
	if presence, exists := session.Presences[userID]; exists {
		presence.Status = UserStatusOffline
		presence.LastActivity = time.Now()
	}

	// 更新统计信息
	session.Stats.ActiveUsers = len(session.Cursors)
	session.LastActivity = time.Now()

	// 发送光标离开通知
	go cm.notifyCursorLeave(sessionID, userID)

	cm.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"user_id":    userID,
	}).Debug("用户光标已移除")

	return nil
}

// GetCursors 获取会话中的所有光标
func (cm *CursorManagerImpl) GetCursors(sessionID string) ([]*UserCursor, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("会话ID不能为空")
	}

	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	session, exists := cm.sessions[sessionID]
	if !exists {
		return []*UserCursor{}, nil
	}

	session.mutex.RLock()
	defer session.mutex.RUnlock()

	cursors := make([]*UserCursor, 0, len(session.Cursors))
	for _, cursor := range session.Cursors {
		if cursor.IsActive {
			cursors = append(cursors, cursor)
		}
	}

	return cursors, nil
}

// GetUserCursor 获取特定用户的光标
func (cm *CursorManagerImpl) GetUserCursor(sessionID, userID string) (*UserCursor, error) {
	if sessionID == "" || userID == "" {
		return nil, fmt.Errorf("会话ID和用户ID不能为空")
	}

	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	session, exists := cm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("会话不存在")
	}

	session.mutex.RLock()
	defer session.mutex.RUnlock()

	cursor, exists := session.Cursors[userID]
	if !exists {
		return nil, fmt.Errorf("用户光标不存在")
	}

	return cursor, nil
}

// UpdateUserPresence 更新用户在线状态
func (cm *CursorManagerImpl) UpdateUserPresence(sessionID, userID string, presence *UserPresence) error {
	if sessionID == "" || userID == "" {
		return fmt.Errorf("会话ID和用户ID不能为空")
	}
	if presence == nil {
		return fmt.Errorf("用户状态不能为空")
	}

	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	session := cm.getOrCreateSession(sessionID)

	session.mutex.Lock()
	defer session.mutex.Unlock()

	// 更新状态
	presence.LastActivity = time.Now()
	session.Presences[userID] = presence
	session.LastActivity = time.Now()

	// 发送状态更新通知
	go cm.notifyPresenceUpdate(sessionID, presence)

	cm.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"user_id":    userID,
		"status":     presence.Status,
	}).Debug("用户状态已更新")

	return nil
}

// GetUserPresence 获取用户在线状态
func (cm *CursorManagerImpl) GetUserPresence(sessionID, userID string) (*UserPresence, error) {
	if sessionID == "" || userID == "" {
		return nil, fmt.Errorf("会话ID和用户ID不能为空")
	}

	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	session, exists := cm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("会话不存在")
	}

	session.mutex.RLock()
	defer session.mutex.RUnlock()

	presence, exists := session.Presences[userID]
	if !exists {
		return nil, fmt.Errorf("用户状态不存在")
	}

	return presence, nil
}

// GetActiveUsers 获取活跃用户列表
func (cm *CursorManagerImpl) GetActiveUsers(sessionID string) ([]*UserPresence, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("会话ID不能为空")
	}

	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	session, exists := cm.sessions[sessionID]
	if !exists {
		return []*UserPresence{}, nil
	}

	session.mutex.RLock()
	defer session.mutex.RUnlock()

	presences := make([]*UserPresence, 0, len(session.Presences))
	now := time.Now()

	for _, presence := range session.Presences {
		// 只返回最近活跃的用户（5分钟内）
		if now.Sub(presence.LastActivity) < 5*time.Minute {
			presences = append(presences, presence)
		}
	}

	return presences, nil
}

// DetectCursorConflicts 检测光标冲突
func (cm *CursorManagerImpl) DetectCursorConflicts(sessionID string) ([]*CursorConflict, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("会话ID不能为空")
	}

	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	session, exists := cm.sessions[sessionID]
	if !exists {
		return []*CursorConflict{}, nil
	}

	session.mutex.RLock()
	defer session.mutex.RUnlock()

	conflicts := make([]*CursorConflict, 0)
	cursors := make([]*UserCursor, 0, len(session.Cursors))

	// 收集所有活跃光标
	for _, cursor := range session.Cursors {
		if cursor.IsActive && cursor.Position != nil {
			cursors = append(cursors, cursor)
		}
	}

	// 检测重叠冲突
	for i := 0; i < len(cursors); i++ {
		for j := i + 1; j < len(cursors); j++ {
			conflict := cm.detectOverlapConflict(cursors[i], cursors[j])
			if conflict != nil {
				conflicts = append(conflicts, conflict)
			}
		}
	}

	return conflicts, nil
}

// ResolveCursorConflict 解决光标冲突
func (cm *CursorManagerImpl) ResolveCursorConflict(sessionID string, conflict *CursorConflict) error {
	if sessionID == "" || conflict == nil {
		return fmt.Errorf("会话ID和冲突信息不能为空")
	}

	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	session, exists := cm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("会话不存在")
	}

	session.mutex.Lock()
	defer session.mutex.Unlock()

	// 标记冲突为已解决
	conflict.Resolved = true
	conflict.ResolvedBy = "system" // 可以从上下文获取实际解决者
	conflict.Resolution = "自动解决"

	// 记录解决时间
	for i, existingConflict := range session.Conflicts {
		if existingConflict.ConflictID == conflict.ConflictID {
			session.Conflicts[i] = conflict
			break
		}
	}

	// 发送冲突解决通知
	go cm.notifyConflictResolved(sessionID, conflict)

	cm.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"conflict_id": conflict.ConflictID,
		"resolution":  conflict.Resolution,
	}).Info("光标冲突已解决")

	return nil
}

// GetCursorHistory 获取光标历史
func (cm *CursorManagerImpl) GetCursorHistory(sessionID, userID string, limit int) ([]*CursorPosition, error) {
	if sessionID == "" || userID == "" {
		return nil, fmt.Errorf("会话ID和用户ID不能为空")
	}

	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	session, exists := cm.sessions[sessionID]
	if !exists {
		return []*CursorPosition{}, nil
	}

	session.mutex.RLock()
	defer session.mutex.RUnlock()

	history, exists := session.History[userID]
	if !exists {
		return []*CursorPosition{}, nil
	}

	// 限制返回数量
	if limit > 0 && len(history) > limit {
		history = history[len(history)-limit:]
	}

	return history, nil
}

// CleanupInactiveCursors 清理不活跃的光标
func (cm *CursorManagerImpl) CleanupInactiveCursors(sessionID string, inactiveDuration time.Duration) error {
	if sessionID == "" {
		return fmt.Errorf("会话ID不能为空")
	}

	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	session, exists := cm.sessions[sessionID]
	if !exists {
		return nil
	}

	session.mutex.Lock()
	defer session.mutex.Unlock()

	now := time.Now()
	inactiveUsers := make([]string, 0)

	// 查找不活跃用户
	for userID, cursor := range session.Cursors {
		if now.Sub(cursor.LastSeen) > inactiveDuration {
			inactiveUsers = append(inactiveUsers, userID)
		}
	}

	// 清理不活跃光标
	for _, userID := range inactiveUsers {
		delete(session.Cursors, userID)
		if presence, exists := session.Presences[userID]; exists {
			presence.Status = UserStatusOffline
		}
	}

	// 更新统计信息
	session.Stats.ActiveUsers = len(session.Cursors)

	if len(inactiveUsers) > 0 {
		cm.logger.WithFields(logrus.Fields{
			"session_id":      sessionID,
			"cleaned_users":   len(inactiveUsers),
			"inactive_duration": inactiveDuration,
		}).Debug("清理不活跃光标")
	}

	return nil
}

// RemoveSession 移除会话
func (cm *CursorManagerImpl) RemoveSession(sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("会话ID不能为空")
	}

	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	delete(cm.sessions, sessionID)

	cm.logger.WithField("session_id", sessionID).Debug("光标会话已移除")

	return nil
}

// GetCursorStats 获取光标统计信息
func (cm *CursorManagerImpl) GetCursorStats(sessionID string) (*CursorStats, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("会话ID不能为空")
	}

	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	session, exists := cm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("会话不存在")
	}

	session.mutex.RLock()
	defer session.mutex.RUnlock()

	// 计算平均会话时长
	var totalSessionTime float64
	userCount := 0
	now := time.Now()

	for _, presence := range session.Presences {
		sessionTime := now.Sub(presence.LastActivity).Minutes()
		totalSessionTime += sessionTime
		userCount++
	}

	avgSessionTime := float64(0)
	if userCount > 0 {
		avgSessionTime = totalSessionTime / float64(userCount)
	}

	stats := *session.Stats // 复制统计信息
	stats.AvgSessionTime = avgSessionTime

	return &stats, nil
}

// 辅助方法

// getOrCreateSession 获取或创建会话
func (cm *CursorManagerImpl) getOrCreateSession(sessionID string) *CursorSession {
	session, exists := cm.sessions[sessionID]
	if !exists {
		session = &CursorSession{
			SessionID:    sessionID,
			DocumentID:   fmt.Sprintf("doc_%s", sessionID),
			Cursors:      make(map[string]*UserCursor),
			Presences:    make(map[string]*UserPresence),
			History:      make(map[string][]*CursorPosition),
			Conflicts:    make([]*CursorConflict, 0),
			Stats: &CursorStats{
				SessionID: sessionID,
			},
			CreatedAt:    time.Now(),
			LastActivity: time.Now(),
		}
		cm.sessions[sessionID] = session
	}
	return session
}

// getUserColor 获取用户颜色
func (cm *CursorManagerImpl) getUserColor(userID string) string {
	// 根据用户ID生成一致的颜色
	colors := []string{
		"#FF6B6B", "#4ECDC4", "#45B7D1", "#96CEB4", "#FFEAA7",
		"#DDA0DD", "#98D8C8", "#FFD93D", "#6BCB77", "#FF6B9D",
	}

	hash := 0
	for _, char := range userID {
		hash = int(char) + ((hash << 5) - hash)
	}

	return colors[hash%len(colors)]
}

// detectOverlapConflict 检测重叠冲突
func (cm *CursorManagerImpl) detectOverlapConflict(cursor1, cursor2 *UserCursor) *CursorConflict {
	if cursor1.Position == nil || cursor2.Position == nil {
		return nil
	}

	pos1, pos2 := cursor1.Position, cursor2.Position

	// 检测位置重叠
	distance := abs(pos1.Index - pos2.Index)
	if distance <= 5 { // 5个字符内认为是冲突
		return &CursorConflict{
			ConflictID:    fmt.Sprintf("%s_%s", cursor1.UserID, cursor2.UserID),
			Users:         []string{cursor1.UserID, cursor2.UserID},
			ConflictType:  ConflictTypeOverlap,
			ConflictRange: &TextRange{
				Start: min(pos1.Index, pos2.Index),
				End:   max(pos1.Index, pos2.Index),
			},
			DetectedAt: time.Now(),
			Resolved:   false,
		}
	}

	return nil
}

// detectAndNotifyConflicts 检测并通知冲突
func (cm *CursorManagerImpl) detectAndNotifyConflicts(sessionID string) {
	conflicts, err := cm.DetectCursorConflicts(sessionID)
	if err != nil {
		cm.logger.WithError(err).Error("检测光标冲突失败")
		return
	}

	for _, conflict := range conflicts {
		cm.notifyConflictDetected(sessionID, conflict)
	}
}

// 通知方法

// notifyCursorUpdate 通知光标更新
func (cm *CursorManagerImpl) notifyCursorUpdate(sessionID string, cursor *UserCursor) {
	if cm.notify != nil {
		message := map[string]interface{}{
			"type":   "cursor_update",
			"cursor": cursor,
		}
		cm.notify.SendNotification(sessionID, message)
	}
}

// notifyCursorLeave 通知光标离开
func (cm *CursorManagerImpl) notifyCursorLeave(sessionID, userID string) {
	if cm.notify != nil {
		message := map[string]interface{}{
			"type":    "cursor_leave",
			"user_id": userID,
		}
		cm.notify.SendNotification(sessionID, message)
	}
}

// notifyPresenceUpdate 通知状态更新
func (cm *CursorManagerImpl) notifyPresenceUpdate(sessionID string, presence *UserPresence) {
	if cm.notify != nil {
		message := map[string]interface{}{
			"type":     "presence_update",
			"presence": presence,
		}
		cm.notify.SendNotification(sessionID, message)
	}
}

// notifyConflictDetected 通知冲突检测
func (cm *CursorManagerImpl) notifyConflictDetected(sessionID string, conflict *CursorConflict) {
	if cm.notify != nil {
		message := map[string]interface{}{
			"type":     "conflict_detected",
			"conflict": conflict,
		}
		cm.notify.SendNotification(sessionID, message)
	}
}

// notifyConflictResolved 通知冲突解决
func (cm *CursorManagerImpl) notifyConflictResolved(sessionID string, conflict *CursorConflict) {
	if cm.notify != nil {
		message := map[string]interface{}{
			"type":     "conflict_resolved",
			"conflict": conflict,
		}
		cm.notify.SendNotification(sessionID, message)
	}
}

// 工具函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}