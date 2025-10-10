package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Notification 通知结构体
type Notification struct {
	ID        int    `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	IsRead    bool   `json:"isRead"`
	CreatedAt string `json:"createdAt"`
	RelatedID *int   `json:"relatedId,omitempty"`
}

// NotificationStats 通知统计
type NotificationStats struct {
	Total  int               `json:"total"`
	Unread int               `json:"unread"`
	ByType map[string]int    `json:"byType"`
}

// NotificationHandler 通知处理器
type NotificationHandler struct{}

// NewNotificationHandler 创建通知处理器
func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{}
}

// GetNotifications 获取通知列表
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	// 返回空的通知列表，避免前端404错误
	notifications := []Notification{}

	c.JSON(http.StatusOK, gin.H{
		"data": notifications,
		"error": nil,
	})
}

// GetNotificationStats 获取通知统计
func (h *NotificationHandler) GetNotificationStats(c *gin.Context) {
	// 返回空的统计数据，避免前端404错误
	stats := NotificationStats{
		Total:  0,
		Unread: 0,
		ByType: map[string]int{
			"approval": 0,
			"project":  0,
			"system":   0,
			"finance":  0,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
		"error": nil,
	})
}

// MarkAsRead 标记通知为已读
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid notification ID",
		})
		return
	}

	// 暂时返回成功，不做实际处理
	c.JSON(http.StatusOK, gin.H{
		"data": map[string]interface{}{
			"id":     id,
			"isRead": true,
		},
		"error": nil,
	})
}

// MarkAllAsRead 标记所有通知为已读
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	// 暂时返回成功，不做实际处理
	c.JSON(http.StatusOK, gin.H{
		"data": map[string]interface{}{
			"message": "All notifications marked as read",
		},
		"error": nil,
	})
}

// DeleteNotification 删除通知
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid notification ID",
		})
		return
	}

	// 暂时返回成功，不做实际处理
	c.JSON(http.StatusOK, gin.H{
		"data": map[string]interface{}{
			"id":      id,
			"deleted": true,
		},
		"error": nil,
	})
}