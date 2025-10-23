package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// CollaborationServer 协作WebSocket服务器
type CollaborationServer struct {
	editRepo   repositories.EditingRepository
	editService services.EditingService
	logger     *logrus.Logger
	upgrader   websocket.Upgrader
	rooms      map[string]*Room
	roomsMutex sync.RWMutex
}

// Room 协作房间
type Room struct {
	DocumentID     string
	RoomName       string
	Participants   map[string]*Participant
	participantsMutex sync.RWMutex
	MessageChannel chan *Message
	DestroyChan    chan bool
	CreatedAt      time.Time
	LastActivity   time.Time
}

// Participant 参与者
type Participant struct {
	UserID      string
	UserName    string
	UserAvatar  string
	UserColor   string
	Socket      *websocket.Conn
	SocketID    string
	SessionID   string
	Status      string
	LastSeen    time.Time
	JoinedAt    time.Time
	Permissions []string
}

// Message WebSocket消息
type Message struct {
	Type      string                 `json:"type"`
	UserID    string                 `json:"user_id"`
	SessionID string                 `json:"session_id"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// MessageTypes 消息类型常量
const (
	MessageTypeOperation    = "operation"
	MessageTypeCursor       = "cursor"
	MessageTypeSelection    = "selection"
	MessageTypeAwareness    = "awareness"
	MessageTypeJoin         = "join"
	MessageTypeLeave        = "leave"
	MessageTypeError        = "error"
	MessageTypeHeartbeat    = "heartbeat"
	MessageTypeRoomInfo     = "room_info"
)

// NewCollaborationServer 创建协作服务器
func NewCollaborationServer(
	editRepo repositories.EditingRepository,
	editService services.EditingService,
	logger *logrus.Logger,
) *CollaborationServer {
	return &CollaborationServer{
		editRepo:    editRepo,
		editService: editService,
		logger:      logger,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 生产环境中需要检查origin
			},
		},
		rooms: make(map[string]*Room),
	}
}

// HandleWebSocket 处理WebSocket连接
func (s *CollaborationServer) HandleWebSocket(c *gin.Context) {
	documentID := c.Param("documentId")
	sessionToken := c.Query("session_token")
	socketID := c.Query("socket_id")

	if documentID == "" || sessionToken == "" || socketID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少必要参数"})
		return
	}

	// 验证编辑会话
	session, err := s.editRepo.GetEditSessionByToken(c.Request.Context(), sessionToken)
	if err != nil {
		s.logger.WithError(err).Error("获取编辑会话失败")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的会话令牌"})
		return
	}

	// 检查会话是否过期
	if session.IsExpired() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "会话已过期"})
		return
	}

	// 检查文档匹配
	if session.DocumentID.String() != documentID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文档不匹配"})
		return
	}

	// 升级到WebSocket连接
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.logger.WithError(err).Error("WebSocket升级失败")
		return
	}

	// 创建参与者
	participant := &Participant{
		UserID:      session.UserID.String(),
		UserName:    session.User.Username,
		UserAvatar:  session.User.Avatar,
		UserColor:   "#2196F3", // 默认颜色
		Socket:      conn,
		SocketID:    socketID,
		SessionID:   session.ID.String(),
		Status:      "online",
		LastSeen:    time.Now(),
		JoinedAt:    time.Now(),
		Permissions: []string{"read", "write"}, // 从权限获取
	}

	// 加入房间
	err = s.joinRoom(c.Request.Context(), documentID, participant)
	if err != nil {
		s.logger.WithError(err).Error("加入房间失败")
		conn.Close()
		return
	}

	// 启动消息处理
	go s.handleConnection(participant)
}

// joinRoom 加入房间
func (s *CollaborationServer) joinRoom(ctx context.Context, documentID string, participant *Participant) error {
	s.roomsMutex.Lock()
	defer s.roomsMutex.Unlock()

	// 获取或创建房间
	room, exists := s.rooms[documentID]
	if !exists {
		room = &Room{
			DocumentID:     documentID,
			RoomName:       fmt.Sprintf("doc_%s", documentID),
			Participants:   make(map[string]*Participant),
			MessageChannel: make(chan *Message, 100),
			DestroyChan:    make(chan bool, 1),
			CreatedAt:      time.Now(),
			LastActivity:   time.Now(),
		}
		s.rooms[documentID] = room

		// 启动房间消息处理
		go s.handleRoomMessages(room)
		go s.roomCleanup(room)
	}

	// 检查用户数量限制
	if len(room.Participants) >= 10 {
		return fmt.Errorf("房间人数已满")
	}

	// 检查用户是否已在房间中
	for _, p := range room.Participants {
		if p.UserID == participant.UserID {
			return fmt.Errorf("用户已在房间中")
		}
	}

	// 添加参与者
	room.Participants[participant.SocketID] = participant
	room.LastActivity = time.Now()

	// 发送加入消息
	joinMessage := &Message{
		Type:      MessageTypeJoin,
		UserID:    participant.UserID,
		SessionID: participant.SessionID,
		Timestamp: time.Now().UnixMilli(),
		Data: map[string]interface{}{
			"user_info": map[string]interface{}{
				"user_id":     participant.UserID,
				"user_name":   participant.UserName,
				"user_avatar": participant.UserAvatar,
				"user_color":  participant.UserColor,
				"status":      participant.Status,
				"joined_at":   participant.JoinedAt,
			},
		},
	}

	// 广播加入消息给其他参与者
	s.broadcastToRoom(room, joinMessage, participant.SocketID)

	// 发送房间信息给新加入者
	s.sendRoomInfo(room, participant)

	s.logger.WithFields(logrus.Fields{
		"document_id": documentID,
		"user_id":     participant.UserID,
		"socket_id":   participant.SocketID,
	}).Info("用户加入协作房间")

	return nil
}

// handleConnection 处理连接
func (s *CollaborationServer) handleConnection(participant *Participant) {
	defer func() {
		s.leaveRoom(participant)
		participant.Socket.Close()
	}()

	// 设置读取超时
	participant.Socket.SetReadDeadline(time.Now().Add(30 * time.Second))
	participant.Socket.SetPongHandler(func(string) error {
		participant.Socket.SetReadDeadline(time.Now().Add(30 * time.Second))
		participant.LastSeen = time.Now()
		return nil
	})

	// 发送心跳
	go s.sendHeartbeat(participant)

	for {
		// 读取消息
		var message Message
		err := participant.Socket.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.WithError(err).Warn("WebSocket连接异常关闭")
			}
			break
		}

		// 更新最后活动时间
		participant.LastSeen = time.Now()

		// 设置消息发送者信息
		message.UserID = participant.UserID
		message.SessionID = participant.SessionID
		message.Timestamp = time.Now().UnixMilli()

		// 处理消息
		err = s.handleMessage(participant, &message)
		if err != nil {
			s.logger.WithError(err).Error("处理消息失败")
			s.sendError(participant, err.Error())
		}
	}
}

// handleMessage 处理消息
func (s *CollaborationServer) handleMessage(participant *Participant, message *Message) error {
	switch message.Type {
	case MessageTypeOperation:
		return s.handleOperationMessage(participant, message)
	case MessageTypeCursor:
		return s.handleCursorMessage(participant, message)
	case MessageTypeSelection:
		return s.handleSelectionMessage(participant, message)
	case MessageTypeAwareness:
		return s.handleAwarenessMessage(participant, message)
	case MessageTypeHeartbeat:
		return s.handleHeartbeatMessage(participant, message)
	default:
		return fmt.Errorf("未知消息类型: %s", message.Type)
	}
}

// handleOperationMessage 处理操作消息
func (s *CollaborationServer) handleOperationMessage(participant *Participant, message *Message) error {
	// 获取房间
	room := s.getRoomByParticipant(participant)
	if room == nil {
		return fmt.Errorf("参与者不在任何房间中")
	}

	// 构建操作数据
	operationData := &models.OperationData{
		Type:    getStringFromMap(message.Data, "type"),
		Content: getStringFromMap(message.Data, "content"),
		Origin:  participant.UserID,
		Author:  participant.UserName,
	}

	// 解析位置和长度
	if pos, ok := message.Data["position"].(float64); ok {
		operationData.Position = int(pos)
	}
	if length, ok := message.Data["length"].(float64); ok {
		operationData.Length = int(length)
	}

	// 解析属性
	if attrs, ok := message.Data["attributes"].(map[string]interface{}); ok {
		operationData.Attributes = attrs
	}

	// 调用编辑服务处理操作
	req := &services.EditOperationRequest{
		SessionID:     participant.SessionID,
		OperationType: operationData.Type,
		OperationData: operationData,
		YjsState:      make(map[string]uint64),
	}

	if state, ok := message.Data["yjs_state"].(map[string]interface{}); ok {
		for k, v := range state {
			if val, ok := v.(float64); ok {
				req.YjsState[k] = uint64(val)
			}
		}
	}

	ctx := context.Background()
	err := s.editService.HandleEditOperation(ctx, req)
	if err != nil {
		return fmt.Errorf("处理编辑操作失败: %w", err)
	}

	// 广播操作给房间其他参与者
	s.broadcastToRoom(room, message, participant.SocketID)

	room.LastActivity = time.Now()

	return nil
}

// handleCursorMessage 处理光标消息
func (s *CollaborationServer) handleCursorMessage(participant *Participant, message *Message) error {
	room := s.getRoomByParticipant(participant)
	if room == nil {
		return fmt.Errorf("参与者不在任何房间中")
	}

	// 广播光标位置给其他参与者
	s.broadcastToRoom(room, message, participant.SocketID)

	return nil
}

// handleSelectionMessage 处理选择消息
func (s *CollaborationServer) handleSelectionMessage(participant *Participant, message *Message) error {
	room := s.getRoomByParticipant(participant)
	if room == nil {
		return fmt.Errorf("参与者不在任何房间中")
	}

	// 广播选择范围给其他参与者
	s.broadcastToRoom(room, message, participant.SocketID)

	return nil
}

// handleAwarenessMessage 处理用户状态消息
func (s *CollaborationServer) handleAwarenessMessage(participant *Participant, message *Message) error {
	room := s.getRoomByParticipant(participant)
	if room == nil {
		return fmt.Errorf("参与者不在任何房间中")
	}

	// 更新参与者状态
	if status, ok := message.Data["status"].(string); ok {
		participant.Status = status
	}

	// 广播用户状态给其他参与者
	s.broadcastToRoom(room, message, participant.SocketID)

	return nil
}

// handleHeartbeatMessage 处理心跳消息
func (s *CollaborationServer) handleHeartbeatMessage(participant *Participant, message *Message) error {
	participant.LastSeen = time.Now()
	return nil
}

// handleRoomMessages 处理房间消息
func (s *CollaborationServer) handleRoomMessages(room *Room) {
	for {
		select {
		case message := <-room.MessageChannel:
			s.broadcastToRoom(room, message, "")
		case <-room.DestroyChan:
			return
		}
	}
}

// leaveRoom 离开房间
func (s *CollaborationServer) leaveRoom(participant *Participant) {
	s.roomsMutex.Lock()
	defer s.roomsMutex.Unlock()

	// 查找房间
	for documentID, room := range s.rooms {
		if _, exists := room.Participants[participant.SocketID]; exists {
			// 从房间移除参与者
			delete(room.Participants, participant.SocketID)
			room.LastActivity = time.Now()

			// 发送离开消息
			leaveMessage := &Message{
				Type:      MessageTypeLeave,
				UserID:    participant.UserID,
				SessionID: participant.SessionID,
				Timestamp: time.Now().UnixMilli(),
				Data: map[string]interface{}{
					"user_id": participant.UserID,
				},
			}

			// 广播离开消息给其他参与者
			s.broadcastToRoom(room, leaveMessage, participant.SocketID)

			s.logger.WithFields(logrus.Fields{
				"document_id": documentID,
				"user_id":     participant.UserID,
				"socket_id":   participant.SocketID,
			}).Info("用户离开协作房间")

			// 如果房间为空，标记为可销毁
			if len(room.Participants) == 0 {
				go func() {
					time.Sleep(5 * time.Minute) // 5分钟后销毁空房间
					s.destroyRoom(documentID)
				}()
			}

			break
		}
	}
}

// destroyRoom 销毁房间
func (s *CollaborationServer) destroyRoom(documentID string) {
	s.roomsMutex.Lock()
	defer s.roomsMutex.Unlock()

	if room, exists := s.rooms[documentID]; exists {
		if len(room.Participants) == 0 {
			close(room.MessageChannel)
			close(room.DestroyChan)
			delete(s.rooms, documentID)
			s.logger.WithField("document_id", documentID).Info("销毁协作房间")
		}
	}
}

// roomCleanup 房间清理
func (s *CollaborationServer) roomCleanup(room *Room) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-room.DestroyChan:
			return
		case <-ticker.C:
			// 检查超时的参与者
			now := time.Now()
			timeoutParticipants := []string{}

			room.participantsMutex.RLock()
			for socketID, participant := range room.Participants {
				if now.Sub(participant.LastSeen) > 2*time.Minute {
					timeoutParticipants = append(timeoutParticipants, socketID)
				}
			}
			room.participantsMutex.RUnlock()

			// 移除超时的参与者
			for _, socketID := range timeoutParticipants {
				if participant, exists := room.Participants[socketID]; exists {
					participant.Socket.Close()
					s.leaveRoom(participant)
				}
			}
		}
	}
}

// getRoomByParticipant 根据参与者获取房间
func (s *CollaborationServer) getRoomByParticipant(participant *Participant) *Room {
	s.roomsMutex.RLock()
	defer s.roomsMutex.RUnlock()

	for _, room := range s.rooms {
		if _, exists := room.Participants[participant.SocketID]; exists {
			return room
		}
	}
	return nil
}

// broadcastToRoom 向房间广播消息
func (s *CollaborationServer) broadcastToRoom(room *Room, message *Message, excludeSocketID string) {
	room.participantsMutex.RLock()
	defer room.participantsMutex.RUnlock()

	for socketID, participant := range room.Participants {
		if socketID != excludeSocketID {
			err := participant.Socket.WriteJSON(message)
			if err != nil {
				s.logger.WithError(err).WithFields(logrus.Fields{
					"socket_id": socketID,
					"user_id":   participant.UserID,
				}).Error("发送消息失败")
				// 关闭有问题的连接
				participant.Socket.Close()
			}
		}
	}
}

// sendRoomInfo 发送房间信息
func (s *CollaborationServer) sendRoomInfo(room *Room, participant *Participant) {
	participants := make([]map[string]interface{}, 0)

	room.participantsMutex.RLock()
	for _, p := range room.Participants {
		if p.SocketID != participant.SocketID {
			participants = append(participants, map[string]interface{}{
				"user_id":     p.UserID,
				"user_name":   p.UserName,
				"user_avatar": p.UserAvatar,
				"user_color":  p.UserColor,
				"status":      p.Status,
				"last_seen":   p.LastSeen,
				"joined_at":   p.JoinedAt,
			})
		}
	}
	room.participantsMutex.RUnlock()

	roomInfoMessage := &Message{
		Type:      MessageTypeRoomInfo,
		UserID:    participant.UserID,
		SessionID: participant.SessionID,
		Timestamp: time.Now().UnixMilli(),
		Data: map[string]interface{}{
			"room_name":    room.RoomName,
			"participants": participants,
			"active_users": len(room.Participants),
		},
	}

	err := participant.Socket.WriteJSON(roomInfoMessage)
	if err != nil {
		s.logger.WithError(err).Error("发送房间信息失败")
	}
}

// sendError 发送错误消息
func (s *CollaborationServer) sendError(participant *Participant, errorMsg string) {
	errorMessage := &Message{
		Type:      MessageTypeError,
		UserID:    participant.UserID,
		SessionID: participant.SessionID,
		Timestamp: time.Now().UnixMilli(),
		Data: map[string]interface{}{
			"error": errorMsg,
		},
	}

	err := participant.Socket.WriteJSON(errorMessage)
	if err != nil {
		s.logger.WithError(err).Error("发送错误消息失败")
	}
}

// sendHeartbeat 发送心跳
func (s *CollaborationServer) sendHeartbeat(participant *Participant) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			heartbeatMessage := &Message{
				Type:      MessageTypeHeartbeat,
				UserID:    participant.UserID,
				SessionID: participant.SessionID,
				Timestamp: time.Now().UnixMilli(),
				Data:      map[string]interface{}{},
			}

			err := participant.Socket.WriteJSON(heartbeatMessage)
			if err != nil {
				s.logger.WithError(err).Debug("发送心跳失败")
				return
			}
		}
	}
}

// GetRoomStats 获取房间统计信息
func (s *CollaborationServer) GetRoomStats() map[string]interface{} {
	s.roomsMutex.RLock()
	defer s.roomsMutex.RUnlock()

	stats := make(map[string]interface{})

	totalRooms := len(s.rooms)
	totalParticipants := 0
	roomDetails := make([]map[string]interface{}, 0)

	for documentID, room := range s.rooms {
		room.participantsMutex.RLock()
		participantCount := len(room.Participants)
		totalParticipants += participantCount

		participantList := make([]map[string]interface{}, 0)
		for _, p := range room.Participants {
			participantList = append(participantList, map[string]interface{}{
				"user_id":   p.UserID,
				"user_name": p.UserName,
				"status":    p.Status,
				"last_seen": p.LastSeen,
			})
		}
		room.participantsMutex.RUnlock()

		roomDetails = append(roomDetails, map[string]interface{}{
			"document_id":   documentID,
			"room_name":     room.RoomName,
			"participants":  participantCount,
			"created_at":    room.CreatedAt,
			"last_activity": room.LastActivity,
			"users":         participantList,
		})
	}

	stats["total_rooms"] = totalRooms
	stats["total_participants"] = totalParticipants
	stats["rooms"] = roomDetails

	return stats
}

// GetActiveRoomCount 获取活跃房间数量
func (s *CollaborationServer) GetActiveRoomCount() int {
	s.roomsMutex.RLock()
	defer s.roomsMutex.RUnlock()

	count := 0
	for _, room := range s.rooms {
		if len(room.Participants) > 0 {
			count++
		}
	}
	return count
}

// GetStringFromMap 从map中获取字符串值
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}