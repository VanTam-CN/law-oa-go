import { get, post, del } from './http'

export interface Notification {
  id: number
  type: 'approval' | 'project' | 'system' | 'finance'
  title: string
  content: string
  isRead: boolean
  createdAt: string
  relatedId?: number
}

export interface NotificationStats {
  total: number
  unread: number
  byType: Record<string, number>
}

export const notificationService = {
  // 获取通知列表
  getNotifications: async () => {
    const response = await get<{ notifications: Notification[]; total: number }>('/notifications')
    return response.notifications || []
  },

  // 获取通知统计
  getNotificationStats: () => get<NotificationStats>('/notifications/stats'),

  // 标记通知为已读
  markAsRead: (id: number) => post(`/notifications/${id}/read`),

  // 标记所有通知为已读
  markAllAsRead: () => post('/notifications/read-all'),

  // 删除通知
  deleteNotification: (id: number) => del(`/notifications/${id}`),
}
