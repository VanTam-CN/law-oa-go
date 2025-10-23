package handlers

import (
	"law-oa-go/internal/common"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct{}

func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{}
}

type Notification struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	ReadAt      *string `json:"read_at,omitempty"`
}

type NotificationStats struct {
	Total   int `json:"total"`
	Unread  int `json:"unread"`
	Read    int `json:"read"`
}

// GetNotifications 获取通知列表
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	// 模拟通知数据
	notifications := []Notification{
		{
			ID:        "1",
			Title:     "新案件分配",
			Content:   "您有一个新的案件需要处理",
			Type:      "case",
			Status:    "unread",
			CreatedAt: "2025-10-20T14:00:00Z",
		},
		{
			ID:        "2",
			Title:     "客户消息",
			Content:   "客户张三发来了一条新消息",
			Type:      "message",
			Status:    "unread",
			CreatedAt: "2025-10-20T13:30:00Z",
		},
		{
			ID:        "3",
			Title:     "系统通知",
			Content:   "系统将于今晚进行维护",
			Type:      "system",
			Status:    "read",
			CreatedAt: "2025-10-20T12:00:00Z",
			ReadAt:    stringPtr("2025-10-20T13:00:00Z"),
		},
	}

	common.APISuccess(c, gin.H{
		"notifications": notifications,
		"total":         len(notifications),
	})
}

// GetNotificationStats 获取通知统计
func (h *NotificationHandler) GetNotificationStats(c *gin.Context) {
	stats := NotificationStats{
		Total:  3,
		Unread: 2,
		Read:   1,
	}

	common.APISuccess(c, stats)
}

func stringPtr(s string) *string {
	return &s
}