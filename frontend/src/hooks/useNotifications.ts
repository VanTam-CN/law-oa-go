import { useState, useEffect, useCallback } from 'react'
import {
  notificationService,
  type Notification,
  type NotificationStats,
} from '@/services/notification'

/**
 * 消息通知钩子，用于在组件中获取和管理通知状态
 * @returns 通知状态和管理函数
 */
export const useNotifications = () => {
  const [notifications, setNotifications] = useState<Notification[]>([])
  const [stats, setStats] = useState<NotificationStats>({
    total: 0,
    unread: 0,
    byType: {},
  })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // 获取通知列表
  const fetchNotifications = useCallback(async () => {
    setLoading(true)
    setError(null)

    try {
      const data = await notificationService.getNotifications()
      setNotifications(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : '获取通知失败')
    } finally {
      setLoading(false)
    }
  }, [])

  // 获取通知统计
  const fetchNotificationStats = useCallback(async () => {
    try {
      const data = await notificationService.getNotificationStats()
      setStats(data)
    } catch (err) {
      console.error('获取通知统计失败:', err)
    }
  }, [])

  // 标记通知为已读
  const markAsRead = useCallback(async (id: number) => {
    try {
      await notificationService.markAsRead(id)
      // 更新本地状态
      setNotifications((prev) =>
        prev.map((notification) =>
          notification.id === id ? { ...notification, isRead: true } : notification,
        ),
      )
      // 更新统计
      setStats((prev) => ({
        ...prev,
        unread: Math.max(0, prev.unread - 1),
        total: Math.max(0, prev.total - 1),
      }))
    } catch (err) {
      console.error('标记已读失败:', err)
    }
  }, [])

  // 标记所有通知为已读
  const markAllAsRead = useCallback(async () => {
    try {
      await notificationService.markAllAsRead()
      // 更新本地状态
      setNotifications((prev) => prev.map((notification) => ({ ...notification, isRead: true })))
      // 更新统计
      setStats((prev) => ({
        ...prev,
        unread: 0,
      }))
    } catch (err) {
      console.error('标记全部已读失败:', err)
    }
  }, [])

  // 删除通知
  const deleteNotification = useCallback(async (id: number) => {
    try {
      await notificationService.deleteNotification(id)
      // 更新本地状态
      setNotifications((prev) => prev.filter((notification) => notification.id !== id))
      // 更新统计
      setStats((prev) => ({
        ...prev,
        total: Math.max(0, prev.total - 1),
        unread: Math.max(0, prev.unread - 1),
      }))
    } catch (err) {
      console.error('删除通知失败:', err)
    }
  }, [])

  // 刷新数据
  const refresh = useCallback(async () => {
    await Promise.all([fetchNotifications(), fetchNotificationStats()])
  }, [fetchNotifications, fetchNotificationStats])

  // 组件挂载时获取数据
  useEffect(() => {
    refresh()
  }, [refresh])

  return {
    notifications,
    stats,
    loading,
    error,
    fetchNotifications,
    fetchNotificationStats,
    markAsRead,
    markAllAsRead,
    deleteNotification,
    refresh,
  }
}

export default useNotifications
