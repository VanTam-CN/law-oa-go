import { get } from './http'

export interface DashboardStatistics {
  pendingApprovals: number
  activeProjects: number
  completedProjects: number
  totalClients: number
  activeClients: number
  monthlyRevenue: number
  yearlyRevenue: number
  totalProjects: number
  projectStatus: Record<string, number>
  approvalStatus: Record<string, number>
  monthlyRevenueTrend: Array<{
    month: string
    revenue: number
  }>
  // 财务统计数据
  financeStats: {
    totalRevenue: number
    pendingRevenue: number
    overdueRevenue: number
    totalExpenses: number
  }
}

export interface RecentActivity {
  id: number
  type: string
  title: string
  status: string
  createdAt: string
  user: string
}

export interface TodoItem {
  id: number
  type: string
  title: string
  priority: string
  deadline: string
  assignee: string
}

export const dashboardService = {
  // 获取仪表盘统计数据
  getStatistics: () => get<DashboardStatistics>('/dashboard/statistics'),

  // 获取最近活动
  getActivities: () => get<RecentActivity[]>('/dashboard/activities'),

  // 获取待办事项
  getTodos: () => get<TodoItem[]>('/dashboard/todos'),

  // 获取当前用户信息
  getCurrentUser: () => get<any>('/users/profile'),
}
