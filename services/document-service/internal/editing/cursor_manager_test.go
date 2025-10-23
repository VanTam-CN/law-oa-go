package editing

import (
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNotificationService 模拟通知服务
type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) SendNotification(sessionID string, message interface{}) error {
	args := m.Called(sessionID, message)
	return args.Error(0)
}

// TestNewCursorManager 测试创建光标管理器
func TestNewCursorManager(t *testing.T) {
	logger := logrus.New()
	mockNotify := &MockNotificationService{}

	manager := NewCursorManager(logger, mockNotify)

	assert.NotNil(t, manager)
	assert.Implements(t, (*CursorManager)(nil), manager)
}

// TestTrackCursor 测试跟踪光标
func TestTrackCursor(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // 减少测试输出
	mockNotify := &MockNotificationService{}

	manager := NewCursorManager(logger, mockNotify)
	sessionID := "test-session"
	userID := "user-123"
	position := &CursorPosition{
		Index:     100,
		Length:    0,
		Timestamp: time.Now(),
	}

	// 模拟通知服务调用
	mockNotify.On("SendNotification", sessionID, mock.AnythingOfType("map[string]interface {}")).Return(nil)

	// 测试正常跟踪
	err := manager.TrackCursor(sessionID, userID, position)
	assert.NoError(t, err)

	// 验证光标已创建
	cursor, err := manager.GetUserCursor(sessionID, userID)
	assert.NoError(t, err)
	assert.NotNil(t, cursor)
	assert.Equal(t, userID, cursor.UserID)
	assert.Equal(t, position.Index, cursor.Position.Index)

	// 测试更新光标位置
	newPosition := &CursorPosition{
		Index:     150,
		Length:    5,
		Timestamp: time.Now(),
	}

	err = manager.TrackCursor(sessionID, userID, newPosition)
	assert.NoError(t, err)

	// 验证光标已更新
	cursor, err = manager.GetUserCursor(sessionID, userID)
	assert.NoError(t, err)
	assert.Equal(t, newPosition.Index, cursor.Position.Index)
	assert.Equal(t, newPosition.Length, cursor.Position.Length)

	mockNotify.AssertExpectations(t)
}

// TestTrackCursorValidation 测试光标跟踪验证
func TestTrackCursorValidation(t *testing.T) {
	logger := logrus.New()
	mockNotify := &MockNotificationService{}
	manager := NewCursorManager(logger, mockNotify)
	position := &CursorPosition{Index: 100}

	// 测试空会话ID
	err := manager.TrackCursor("", "user123", position)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "会话ID和用户ID不能为空")

	// 测试空用户ID
	err = manager.TrackCursor("session123", "", position)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "会话ID和用户ID不能为空")

	// 测试空位置
	err = manager.TrackCursor("session123", "user123", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "光标位置不能为空")
}

// TestUntrackCursor 测试取消跟踪光标
func TestUntrackCursor(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	mockNotify := &MockNotificationService{}
	manager := NewCursorManager(logger, mockNotify)
	sessionID := "test-session"
	userID := "user-123"
	position := &CursorPosition{Index: 100}

	// 模拟通知服务调用
	mockNotify.On("SendNotification", sessionID, mock.AnythingOfType("map[string]interface {}")).Return(nil).Twice()

	// 先跟踪光标
	err := manager.TrackCursor(sessionID, userID, position)
	assert.NoError(t, err)

	// 验证光标存在
	cursor, err := manager.GetUserCursor(sessionID, userID)
	assert.NoError(t, err)
	assert.NotNil(t, cursor)

	// 取消跟踪
	err = manager.UntrackCursor(sessionID, userID)
	assert.NoError(t, err)

	// 验证光标已移除
	cursor, err = manager.GetUserCursor(sessionID, userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "用户光标不存在")

	mockNotify.AssertExpectations(t)
}

// TestGetCursors 测试获取所有光标
func TestGetCursors(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	mockNotify := &MockNotificationService{}
	manager := NewCursorManager(logger, mockNotify)
	sessionID := "test-session"

	// 模拟通知服务调用
	mockNotify.On("SendNotification", sessionID, mock.AnythingOfType("map[string]interface {}")).Return(nil).Times(3)

	// 添加多个用户光标
	users := []string{"user1", "user2", "user3"}
	positions := []*CursorPosition{
		{Index: 100, Length: 0},
		{Index: 200, Length: 5},
		{Index: 300, Length: 10},
	}

	for i, userID := range users {
		err := manager.TrackCursor(sessionID, userID, positions[i])
		assert.NoError(t, err)
	}

	// 获取所有光标
	cursors, err := manager.GetCursors(sessionID)
	assert.NoError(t, err)
	assert.Len(t, cursors, 3)

	// 验证光标信息
	cursorMap := make(map[string]*UserCursor)
	for _, cursor := range cursors {
		cursorMap[cursor.UserID] = cursor
	}

	for i, userID := range users {
		cursor, exists := cursorMap[userID]
		assert.True(t, exists)
		assert.Equal(t, positions[i].Index, cursor.Position.Index)
		assert.Equal(t, positions[i].Length, cursor.Position.Length)
	}

	mockNotify.AssertExpectations(t)
}

// TestUpdateUserPresence 测试更新用户状态
func TestUpdateUserPresence(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	mockNotify := &MockNotificationService{}
	manager := NewCursorManager(logger, mockNotify)
	sessionID := "test-session"
	userID := "user-123"

	// 模拟通知服务调用
	mockNotify.On("SendNotification", sessionID, mock.AnythingOfType("map[string]interface {}")).Return(nil)

	presence := &UserPresence{
		UserID:        userID,
		UserName:      "测试用户",
		Status:        UserStatusOnline,
		CurrentEditor: "rich-text",
		Activity:      "正在编辑",
		LastActivity:  time.Now(),
		Metadata:      map[string]string{"role": "editor"},
	}

	// 更新用户状态
	err := manager.UpdateUserPresence(sessionID, userID, presence)
	assert.NoError(t, err)

	// 获取用户状态
	retrievedPresence, err := manager.GetUserPresence(sessionID, userID)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedPresence)
	assert.Equal(t, userID, retrievedPresence.UserID)
	assert.Equal(t, UserStatusOnline, retrievedPresence.Status)
	assert.Equal(t, "rich-text", retrievedPresence.CurrentEditor)

	mockNotify.AssertExpectations(t)
}

// TestGetActiveUsers 测试获取活跃用户
func TestGetActiveUsers(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	mockNotify := &MockNotificationService{}
	manager := NewCursorManager(logger, mockNotify)
	sessionID := "test-session"

	// 模拟通知服务调用
	mockNotify.On("SendNotification", sessionID, mock.AnythingOfType("map[string]interface {}")).Return(nil).Times(3)

	// 添加多个用户状态
	users := []string{"user1", "user2", "user3"}
	presences := []*UserPresence{
		{
			UserID:       "user1",
			Status:       UserStatusOnline,
			LastActivity: time.Now(),
		},
		{
			UserID:       "user2",
			Status:       UserStatusAway,
			LastActivity: time.Now().Add(-2 * time.Minute), // 2分钟前
		},
		{
			UserID:       "user3",
			Status:       UserStatusOffline,
			LastActivity: time.Now().Add(-10 * time.Minute), // 10分钟前，应该被过滤
		},
	}

	for i, userID := range users {
		err := manager.UpdateUserPresence(sessionID, userID, presences[i])
		assert.NoError(t, err)
	}

	// 获取活跃用户（5分钟内活跃）
	activeUsers, err := manager.GetActiveUsers(sessionID)
	assert.NoError(t, err)
	assert.Len(t, activeUsers, 2) // 只有user1和user2在5分钟内活跃

	mockNotify.AssertExpectations(t)
}

// TestDetectCursorConflicts 测试光标冲突检测
func TestDetectCursorConflicts(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	mockNotify := &MockNotificationService{}
	manager := NewCursorManager(logger, mockNotify)
	sessionID := "test-session"

	// 模拟通知服务调用
	mockNotify.On("SendNotification", sessionID, mock.AnythingOfType("map[string]interface {}")).Return(nil).Times(2)

	// 添加两个用户的光标，位置相近（会产生冲突）
	position1 := &CursorPosition{Index: 100, Length: 0, Timestamp: time.Now()}
	position2 := &CursorPosition{Index: 102, Length: 0, Timestamp: time.Now()}

	err := manager.TrackCursor(sessionID, "user1", position1)
	assert.NoError(t, err)

	err = manager.TrackCursor(sessionID, "user2", position2)
	assert.NoError(t, err)

	// 检测冲突
	conflicts, err := manager.DetectCursorConflicts(sessionID)
	assert.NoError(t, err)
	assert.Len(t, conflicts, 1) // 应该检测到一个冲突

	// 验证冲突信息
	conflict := conflicts[0]
	assert.Contains(t, conflict.Users, "user1")
	assert.Contains(t, conflict.Users, "user2")
	assert.Equal(t, ConflictTypeOverlap, conflict.ConflictType)
	assert.False(t, conflict.Resolved)

	mockNotify.AssertExpectations(t)
}

// TestResolveCursorConflict 测试解决光标冲突
func TestResolveCursorConflict(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	mockNotify := &MockNotificationService{}
	manager := NewCursorManager(logger, mockNotify)
	sessionID := "test-session"

	// 模拟通知服务调用
	mockNotify.On("SendNotification", sessionID, mock.AnythingOfType("map[string]interface {}")).Return(nil).Times(3)

	// 添加冲突光标
	position1 := &CursorPosition{Index: 100, Length: 0, Timestamp: time.Now()}
	position2 := &CursorPosition{Index: 102, Length: 0, Timestamp: time.Now()}

	manager.TrackCursor(sessionID, "user1", position1)
	manager.TrackCursor(sessionID, "user2", position2)

	// 检测冲突
	conflicts, err := manager.DetectCursorConflicts(sessionID)
	assert.NoError(t, err)
	assert.Len(t, conflicts, 1)

	// 解决冲突
	conflict := conflicts[0]
	conflict.Resolution = "手动调整"
	err = manager.ResolveCursorConflict(sessionID, conflict)
	assert.NoError(t, err)

	// 验证冲突已解决
	resolvedConflicts, err := manager.DetectCursorConflicts(sessionID)
	assert.NoError(t, err)
	assert.Len(t, resolvedConflicts, 0) // 冲突已解决，应该返回空列表

	mockNotify.AssertExpectations(t)
}

// TestGetCursorHistory 测试获取光标历史
func TestGetCursorHistory(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	mockNotify := &MockNotificationService{}
	manager := NewCursorManager(logger, mockNotify)
	sessionID := "test-session"
	userID := "user-123"

	// 模拟通知服务调用
	mockNotify.On("SendNotification", sessionID, mock.AnythingOfType("map[string]interface {}")).Return(nil).Times(5)

	// 添加多个光标位置
	positions := []*CursorPosition{
		{Index: 100, Length: 0, Timestamp: time.Now()},
		{Index: 150, Length: 5, Timestamp: time.Now()},
		{Index: 200, Length: 0, Timestamp: time.Now()},
		{Index: 250, Length: 10, Timestamp: time.Now()},
		{Index: 300, Length: 0, Timestamp: time.Now()},
	}

	for _, position := range positions {
		err := manager.TrackCursor(sessionID, userID, position)
		assert.NoError(t, err)
	}

	// 获取完整历史
	history, err := manager.GetCursorHistory(sessionID, userID, 0)
	assert.NoError(t, err)
	assert.Len(t, history, 5)

	// 验证历史顺序
	for i, position := range positions {
		assert.Equal(t, position.Index, history[i].Index)
	}

	// 获取限制数量的历史
	limitedHistory, err := manager.GetCursorHistory(sessionID, userID, 3)
	assert.NoError(t, err)
	assert.Len(t, limitedHistory, 3)

	// 验证获取的是最新的3个位置
	for i, position := range positions[2:] {
		assert.Equal(t, position.Index, limitedHistory[i].Index)
	}

	mockNotify.AssertExpectations(t)
}

// TestCleanupInactiveCursors 测试清理不活跃光标
func TestCleanupInactiveCursors(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	mockNotify := &MockNotificationService{}
	manager := NewCursorManager(logger, mockNotify)
	sessionID := "test-session"

	// 模拟通知服务调用
	mockNotify.On("SendNotification", sessionID, mock.AnythingOfType("map[string]interface {}")).Return(nil).Times(2)

	// 添加两个用户光标
	position1 := &CursorPosition{Index: 100, Length: 0, Timestamp: time.Now()}
	position2 := &CursorPosition{Index: 200, Length: 0, Timestamp: time.Now()}

	err := manager.TrackCursor(sessionID, "user1", position1)
	assert.NoError(t, err)

	err = manager.TrackCursor(sessionID, "user2", position2)
	assert.NoError(t, err)

	// 验证两个光标都存在
	cursors, err := manager.GetCursors(sessionID)
	assert.NoError(t, err)
	assert.Len(t, cursors, 2)

	// 等待一小段时间，然后清理不活跃光标
	time.Sleep(10 * time.Millisecond)
	err = manager.CleanupInactiveCursors(sessionID, 1*time.Millisecond)
	assert.NoError(t, err)

	// 验证光标已被清理
	cursors, err = manager.GetCursors(sessionID)
	assert.NoError(t, err)
	assert.Len(t, cursors, 0)

	mockNotify.AssertExpectations(t)
}

// TestGetCursorStats 测试获取光标统计
func TestGetCursorStats(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	mockNotify := &MockNotificationService{}
	manager := NewCursorManager(logger, mockNotify)
	sessionID := "test-session"

	// 模拟通知服务调用
	mockNotify.On("SendNotification", sessionID, mock.AnythingOfType("map[string]interface {}")).Return(nil).Times(3)

	// 添加用户光标和状态
	users := []string{"user1", "user2", "user3"}
	positions := []*CursorPosition{
		{Index: 100, Length: 0},
		{Index: 200, Length: 5},
		{Index: 300, Length: 10},
	}

	for i, userID := range users {
		err := manager.TrackCursor(sessionID, userID, positions[i])
		assert.NoError(t, err)

		presence := &UserPresence{
			UserID:       userID,
			Status:       UserStatusOnline,
			LastActivity: time.Now(),
		}
		err = manager.UpdateUserPresence(sessionID, userID, presence)
		assert.NoError(t, err)
	}

	// 获取统计信息
	stats, err := manager.GetCursorStats(sessionID)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, sessionID, stats.SessionID)
	assert.Equal(t, 3, stats.ActiveUsers)
	assert.Equal(t, 3, stats.TotalCursors)
	assert.Equal(t, 3, stats.PeakUsers)
	assert.False(t, stats.LastActivity.IsZero())

	mockNotify.AssertExpectations(t)
}

// TestRemoveSession 测试移除会话
func TestRemoveSession(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	mockNotify := &MockNotificationService{}
	manager := NewCursorManager(logger, mockNotify)
	sessionID := "test-session"

	// 模拟通知服务调用
	mockNotify.On("SendNotification", sessionID, mock.AnythingOfType("map[string]interface {}")).Return(nil).Times(2)

	// 添加用户光标
	position := &CursorPosition{Index: 100, Length: 0}
	err := manager.TrackCursor(sessionID, "user1", position)
	assert.NoError(t, err)

	// 验证会话存在
	cursors, err := manager.GetCursors(sessionID)
	assert.NoError(t, err)
	assert.Len(t, cursors, 1)

	// 移除会话
	err = manager.RemoveSession(sessionID)
	assert.NoError(t, err)

	// 验证会话已移除
	cursors, err = manager.GetCursors(sessionID)
	assert.NoError(t, err)
	assert.Len(t, cursors, 0)

	// 尝试获取统计信息应该失败
	_, err = manager.GetCursorStats(sessionID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "会话不存在")

	mockNotify.AssertExpectations(t)
}

// TestGetUserColor 测试用户颜色生成
func TestGetUserColor(t *testing.T) {
	logger := logrus.New()
	mockNotify := &MockNotificationService{}
	manager := NewCursorManager(logger, mockNotify)

	// 测试相同用户ID生成相同颜色
	userID := "test-user-123"
	color1 := manager.(*CursorManagerImpl).getUserColor(userID)
	color2 := manager.(*CursorManagerImpl).getUserColor(userID)
	assert.Equal(t, color1, color2)

	// 测试不同用户ID生成不同颜色
	userID2 := "different-user-456"
	color3 := manager.(*CursorManagerImpl).getUserColor(userID2)
	assert.NotEqual(t, color1, color3)

	// 验证颜色格式（应该是十六进制颜色代码）
	assert.Regexp(t, "^#[0-9A-Fa-f]{6}$", color1)
	assert.Regexp(t, "^#[0-9A-Fa-f]{6}$", color3)
}

// BenchmarkTrackCursor 性能测试：跟踪光标
func BenchmarkTrackCursor(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	mockNotify := &MockNotificationService{}
	mockNotify.On("SendNotification", mock.AnythingOfType("string"), mock.Anything).Return(nil)

	manager := NewCursorManager(logger, mockNotify)
	sessionID := "benchmark-session"
	position := &CursorPosition{Index: 100, Length: 0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		userID := fmt.Sprintf("user-%d", i)
		manager.TrackCursor(sessionID, userID, position)
	}
}

// BenchmarkGetCursors 性能测试：获取光标
func BenchmarkGetCursors(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	mockNotify := &MockNotificationService{}
	mockNotify.On("SendNotification", mock.AnythingOfType("string"), mock.Anything).Return(nil)

	manager := NewCursorManager(logger, mockNotify)
	sessionID := "benchmark-session"

	// 预先添加1000个光标
	for i := 0; i < 1000; i++ {
		userID := fmt.Sprintf("user-%d", i)
		position := &CursorPosition{Index: i, Length: 0}
		manager.TrackCursor(sessionID, userID, position)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.GetCursors(sessionID)
	}
}