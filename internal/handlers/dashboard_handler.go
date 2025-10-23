package handlers

import (
	"law-oa-go/internal/common"
	"math/rand"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct{}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

type DashboardStats struct {
	TotalCases     int `json:"total_cases"`
	ActiveCases    int `json:"active_cases"`
	CompletedCases int `json:"completed_cases"`
	TotalClients   int `json:"total_clients"`
	NewClients     int `json:"new_clients"`
	TotalLawyers   int `json:"total_lawyers"`
	MonthlyRevenue float64 `json:"monthly_revenue"`
}

type TodoItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	DueDate     string `json:"due_date"`
	Status      string `json:"status"`
}

type ActivityItem struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	User        string `json:"user"`
}

// GetDashboardStatistics 获取仪表盘统计数据
func (h *DashboardHandler) GetDashboardStatistics(c *gin.Context) {
	stats := DashboardStats{
		TotalCases:     45,
		ActiveCases:    23,
		CompletedCases: 22,
		TotalClients:   67,
		NewClients:     5,
		TotalLawyers:   6,
		MonthlyRevenue: 245000.50,
	}

	common.APISuccess(c, stats)
}

// GetDashboardTodos 获取待办事项
func (h *DashboardHandler) GetDashboardTodos(c *gin.Context) {
	todos := []TodoItem{
		{
			ID:          "1",
			Title:       "完成合同审查",
			Description: "审阅ABC公司并购合同",
			Priority:    "high",
			DueDate:     "2025-10-22",
			Status:      "pending",
		},
		{
			ID:          "2",
			Title:       "法庭准备",
			Description: "准备明日庭审材料",
			Priority:    "high",
			DueDate:     "2025-10-21",
			Status:      "pending",
		},
		{
			ID:          "3",
			Title:       "客户会议",
			Description: "与客户XYZ讨论案件进展",
			Priority:    "medium",
			DueDate:     "2025-10-23",
			Status:      "pending",
		},
		{
			ID:          "4",
			Title:       "法律研究",
			Description: "研究最新劳动法相关法规",
			Priority:    "low",
			DueDate:     "2025-10-25",
			Status:      "completed",
		},
	}

	common.APISuccess(c, gin.H{
		"todos": todos,
		"total": len(todos),
	})
}

// GetDashboardActivities 获取活动记录
func (h *DashboardHandler) GetDashboardActivities(c *gin.Context) {
	activities := []ActivityItem{
		{
			ID:          "1",
			Type:        "case",
			Title:       "创建新案件",
			Description: "为客户张三创建民事纠纷案件",
			CreatedAt:   "2025-10-20T10:30:00Z",
			User:        "李律师",
		},
		{
			ID:          "2",
			Type:        "client",
			Title:       "新客户注册",
			Description: "ABC科技注册为系统用户",
			CreatedAt:   "2025-10-20T09:15:00Z",
			User:        "系统",
		},
		{
			ID:          "3",
			Type:        "document",
			Title:       "上传文档",
			Description: "上传合同扫描件至案件#123",
			CreatedAt:   "2025-10-19T16:45:00Z",
			User:        "王律师",
		},
		{
			ID:          "4",
			Type:        "case",
			Title:       "案件状态更新",
			Description: "案件#456状态更改为进行中",
			CreatedAt:   "2025-10-19T14:20:00Z",
			User:        "张律师",
		},
		{
			ID:          "5",
			Type:        "meeting",
			Title:       "安排会议",
			Description: "安排与客户的初步会议",
			CreatedAt:   "2025-10-19T11:00:00Z",
			User:        "李律师",
		},
	}

	common.APISuccess(c, gin.H{
		"activities": activities,
		"total":     len(activities),
	})
}

// GenerateMockData 生成模拟数据的辅助函数
func generateMockData(data interface{}) {
	// 这里可以添加更复杂的模拟数据生成逻辑
	rand.Seed(100)
}