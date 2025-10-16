package handlers

import (
	"law-oa-go/internal/common"
	"law-oa-go/internal/services"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	userService    *services.UserService
	clientService  *services.ClientService
	caseService    *services.CaseService
}

func NewDashboardHandler(
	userService *services.UserService,
	clientService *services.ClientService,
	caseService *services.CaseService,
) *DashboardHandler {
	return &DashboardHandler{
		userService:   userService,
		clientService: clientService,
		caseService:   caseService,
	}
}

// GetStatistics 获取仪表盘统计数据
func (h *DashboardHandler) GetStatistics(c *gin.Context) {
	// 并行获取客户和案件统计数据
	type StatsResult struct {
		ClientStats *services.ClientStatsResponse
		CaseStats   *services.CaseStatsResponse
		Error       error
	}

	results := make(chan StatsResult, 2)

	// 获取客户统计
	go func() {
		clientStats, err := h.clientService.GetClientStats(c.Request.Context())
		results <- StatsResult{ClientStats: clientStats, Error: err}
	}()

	// 获取案件统计
	go func() {
		caseStats, err := h.caseService.GetCaseStats(c.Request.Context())
		results <- StatsResult{CaseStats: caseStats, Error: err}
	}()

	// 收集结果
	var clientStats *services.ClientStatsResponse
	var caseStats *services.CaseStatsResponse
	var hasError bool

	for i := 0; i < 2; i++ {
		result := <-results
		if result.Error != nil {
			hasError = true
		}
		switch {
		case result.ClientStats != nil:
			clientStats = result.ClientStats
		case result.CaseStats != nil:
			caseStats = result.CaseStats
		}
	}

	if hasError {
		common.APIInternalServerError(c, "获取统计数据失败")
		return
	}

	// 构建响应数据
	response := map[string]interface{}{
		"totalProjects": caseStats.TotalCases,
		"completedProjects": caseStats.ClosedCases,
		"pendingApprovals": 0, // 暂时设为0，后续可以添加审批系统
		"activeClients": clientStats.ActiveClients,
		"totalClients": clientStats.Total,
		"projectStatus": map[string]interface{}{
			"进行中": caseStats.ActiveCases,
			"已完成": caseStats.ClosedCases,
			"已暂停": caseStats.SuspendedCases,
			"已取消": 0,
		},
		"approvalStatus": map[string]interface{}{
			"待审批": 0,
			"已通过": 0,
			"已拒绝": 0,
			"已撤销": 0,
		},
		"financeStats": map[string]interface{}{
			"totalRevenue":  2500000,  // 模拟数据
			"pendingRevenue": 500000,
			"overdueRevenue": 100000,
			"totalExpenses": 1800000,
		},
		"monthlyRevenueTrend": []map[string]interface{}{
			{"month": "1月", "revenue": 200000},
			{"month": "2月", "revenue": 220000},
			{"month": "3月", "revenue": 180000},
			{"month": "4月", "revenue": 250000},
			{"month": "5月", "revenue": 280000},
			{"month": "6月", "revenue": 300000},
		},
	}

	common.APISuccess(c, response)
}

// GetTodos 获取待办事项
func (h *DashboardHandler) GetTodos(c *gin.Context) {
	// 模拟待办事项数据
	// 实际项目中这些数据应该从数据库的待办事项表中获取
	todos := []map[string]interface{}{
		{
			"id":       1,
			"type":     "approval",
			"title":    "审批张三的请假申请",
			"priority": "high",
			"deadline": "2024-01-15",
			"assignee": "李四",
		},
		{
			"id":       2,
			"type":     "project",
			"title":    "更新项目进度报告",
			"priority": "medium",
			"deadline": "2024-01-16",
			"assignee": "王五",
		},
		{
			"id":       3,
			"type":     "client",
			"title":    "客户会议准备",
			"priority": "high",
			"deadline": "2024-01-14",
			"assignee": "赵六",
		},
	}

	common.APISuccess(c, todos)
}

// GetActivities 获取最新动态
func (h *DashboardHandler) GetActivities(c *gin.Context) {
	// 模拟活动数据
	// 实际项目中这些数据应该从数据库的活动日志表中获取
	activities := []map[string]interface{}{
		{
			"id":        1,
			"type":      "approval",
			"title":     "李四审批了张三的请假申请",
			"status":    "已完成",
			"createdAt": "2小时前",
			"user":      "李四",
		},
		{
			"id":        2,
			"type":      "project",
			"title":     "王五更新了项目进度",
			"status":    "进行中",
			"createdAt": "3小时前",
			"user":      "王五",
		},
		{
			"id":        3,
			"type":      "client",
			"title":     "赵六创建了新客户档案",
			"status":    "已完成",
			"createdAt": "5小时前",
			"user":      "赵六",
		},
		{
			"id":        4,
			"type":      "finance",
			"title":     "孙八更新了财务报表",
			"status":    "已完成",
			"createdAt": "1天前",
			"user":      "孙八",
		},
		{
			"id":        5,
			"type":      "project",
			"title":     "周九完成了案件归档",
			"status":    "已完成",
			"createdAt": "1天前",
			"user":      "周九",
		},
		{
			"id":        6,
			"type":      "finance",
			"title":     "孙八更新了财务报表",
			"status":    "已完成",
			"createdAt": "4小时前",
			"user":      "孙八",
		},
	}

	common.APISuccess(c, activities)
}