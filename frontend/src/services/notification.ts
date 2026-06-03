import { get, post, put, del } from './http'

// ============================================================================
// 旧版通知接口（兼容）
// ============================================================================

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

// ============================================================================
// 通知队列接口（新版）
// ============================================================================

export interface NotificationQueue {
  id: number
  created_at: string
  trigger_type: string
  trigger_id: number
  case_id?: number
  recipient_type: 'client' | 'lawyer' | 'admin'
  recipient_id: number
  recipient_name: string
  recipient_contact?: string
  channel: 'email' | 'sms' | 'wechat'
  subject?: string
  content: string
  template_id?: string
  status: 'pending' | 'approved' | 'sent' | 'cancelled' | 'failed'
  priority: 'urgent' | 'normal' | 'low'
  created_by: number
  approved_by?: number
  approved_at?: string
  sent_at?: string
  sent_retry_count: number
  error_message?: string
  contains_sensitive_info: boolean
  auto_send: boolean
  external_message_id?: string
}

export interface NotificationQueueStats {
  total: number
  pending: number
  approved: number
  sent: number
  failed: number
  cancelled: number
  pending_approval: number
  auto_send_count: number
}

export interface NotificationTemplate {
  id: number
  created_at: string
  updated_at: string
  template_code: string
  template_name: string
  channel: 'email' | 'sms' | 'wechat'
  recipient_type: 'client' | 'lawyer' | 'admin'
  trigger_event: string
  subject_template?: string
  content_template: string
  variables: Record<string, string>
  auto_send: boolean
  requires_approval: boolean
  is_active: boolean
}

export interface CreateNotificationRequest {
  trigger_type: string
  trigger_id: number
  case_id?: number
  recipient_type: 'client' | 'lawyer' | 'admin'
  recipient_id: number
  recipient_name: string
  recipient_contact?: string
  channel: 'email' | 'sms' | 'wechat'
  subject?: string
  content: string
  template_id?: string
  priority?: 'urgent' | 'normal' | 'low'
  contains_sensitive_info?: boolean
  auto_send?: boolean
}

export interface CreateTemplateRequest {
  template_code: string
  template_name: string
  channel: 'email' | 'sms' | 'wechat'
  recipient_type: 'client' | 'lawyer' | 'admin'
  trigger_event: string
  subject_template?: string
  content_template: string
  variables?: string[]
  auto_send?: boolean
  requires_approval?: boolean
}

// 通知队列服务
export const notificationQueueService = {
  // 获取通知队列列表
  getQueue: async (params: {
    page?: number
    page_size?: number
    status?: string
    priority?: string
    channel?: string
    recipient_type?: string
    trigger_type?: string
    search?: string
    date_from?: string
    date_to?: string
  } = {}) => {
    const queryParams = new URLSearchParams()
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== '') {
        queryParams.append(key, String(value))
      }
    })
    const query = queryParams.toString()
    return get<{
      data: NotificationQueue[]
      pagination: {
        page: number
        page_size: number
        total: number
        total_pages: number
      }
    }>(`/notifications${query ? `?${query}` : ''}`)
  },

  // 获取通知队列统计
  getQueueStats: () => get<NotificationQueueStats>('/notifications/stats'),

  // 创建通知
  createNotification: (data: CreateNotificationRequest) =>
    post<NotificationQueue>('/notifications', data),

  // 获取通知详情
  getNotificationById: (id: number) =>
    get<NotificationQueue>(`/notifications/${id}`),

  // 更新通知
  updateNotification: (id: number, data: {
    subject?: string
    content?: string
    priority?: 'urgent' | 'normal' | 'low'
  }) => put(`/notifications/${id}`, data),

  // 删除通知
  deleteNotification: (id: number) =>
    del(`/notifications/${id}`),

  // 审批通过通知
  approveNotification: (id: number) =>
    post(`/notifications/${id}/approve`),

  // 审批拒绝通知
  rejectNotification: (id: number, reason: string) =>
    post(`/notifications/${id}/reject`, { reason }),

  // 批量确认通知
  batchConfirm: (ids: number[]) =>
    post('/notifications/batch/confirm', { ids }),

  // 批量取消通知
  batchCancel: (ids: number[]) =>
    post('/notifications/batch/cancel', { ids }),

  // 发送通知
  sendNotification: (id: number) =>
    post(`/notifications/${id}/send`),
}

// 通知模板服务
export const notificationTemplateService = {
  // 获取模板列表
  getTemplates: async (params: {
    page?: number
    page_size?: number
    channel?: string
    recipient_type?: string
    trigger_event?: string
    is_active?: boolean
    search?: string
  } = {}) => {
    const queryParams = new URLSearchParams()
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== '') {
        queryParams.append(key, String(value))
      }
    })
    const query = queryParams.toString()
    return get<{
      data: NotificationTemplate[]
      pagination: {
        page: number
        page_size: number
        total: number
        total_pages: number
      }
    }>(`/notification-templates${query ? `?${query}` : ''}`)
  },

  // 获取启用的模板列表
  getActiveTemplates: () =>
    get<NotificationTemplate[]>('/notification-templates/active'),

  // 根据代码获取模板
  getTemplateByCode: (code: string) =>
    get<NotificationTemplate>(`/notification-templates/code/${code}`),

  // 创建模板
  createTemplate: (data: CreateTemplateRequest) =>
    post<NotificationTemplate>('/notification-templates', data),

  // 更新模板
  updateTemplate: (id: number, data: Partial<CreateTemplateRequest> & {
    is_active?: boolean
  }) => put(`/notification-templates/${id}`, data),

  // 删除模板
  deleteTemplate: (id: number) =>
    del(`/notification-templates/${id}`),

  // 切换模板启用状态
  toggleActive: (id: number) =>
    post<{ message: string; is_active: boolean }>(`/notification-templates/${id}/toggle`),
}

// 状态映射
export const notificationStatusMap: Record<string, { text: string; color: string }> = {
  pending: { text: '待处理', color: 'default' },
  approved: { text: '已审批', color: 'blue' },
  sent: { text: '已发送', color: 'green' },
  cancelled: { text: '已取消', color: 'default' },
  failed: { text: '发送失败', color: 'red' },
}

export const notificationPriorityMap: Record<string, { text: string; color: string }> = {
  urgent: { text: '紧急', color: 'red' },
  normal: { text: '普通', color: 'blue' },
  low: { text: '低', color: 'default' },
}

export const notificationChannelMap: Record<string, { text: string; icon: string }> = {
  email: { text: '邮件', icon: '✉️' },
  sms: { text: '短信', icon: '📱' },
  wechat: { text: '微信', icon: '💬' },
}
