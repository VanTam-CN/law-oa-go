import apiClient from "./api";
import { AppError, Notification } from "../types";

interface NotificationListRequest {
  page?: number;
  page_size?: number;
  type?: string;
  read?: boolean;
  search?: string;
}

interface CreateNotificationRequest {
  title: string;
  message: string;
  type: 'info' | 'warning' | 'error' | 'success' | 'case_update' | 'client_update' | 'system' | 'reminder' | 'deadline' | 'document' | 'message';
  action_url?: string;
  expires_at?: string;
}

interface UpdateNotificationRequest {
  title?: string;
  message?: string;
  type?: 'info' | 'warning' | 'error' | 'success' | 'case_update' | 'client_update' | 'system' | 'reminder' | 'deadline' | 'document' | 'message';
  read?: boolean;
  action_url?: string;
  expires_at?: string;
}

interface NotificationStats {
  total: number;
  unread: number;
  read: number;
  by_type: {
    info: number;
    warning: number;
    error: number;
    success: number;
    case_update: number;
    client_update: number;
    system: number;
    reminder: number;
    deadline: number;
    document: number;
    message: number;
  };
  by_status: {
    unread: number;
    read: number;
  };
}

class NotificationService {
  // 获取通知列表
  async getNotifications(params?: NotificationListRequest): Promise<{
    data: Notification[];
    pagination: {
      page: number;
      page_size: number;
      total: number;
      total_pages: number;
    };
  }> {
    try {
      return await apiClient.getPaginated<Notification>("/notifications", {
        params,
        
      });
    } catch (error: any) {
      console.error("获取通知列表失败:", error);
      throw new AppError(
        error.message || "获取通知列表失败",
        error.code || "GET_NOTIFICATIONS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取通知详情
  async getNotification(id: number): Promise<Notification> {
    try {
      return await apiClient.get<Notification>(`/notifications/${id}`, {
        
      });
    } catch (error: any) {
      console.error("获取通知详情失败:", error);
      throw new AppError(
        error.message || "获取通知详情失败",
        error.code || "GET_NOTIFICATION_ERROR",
        error.statusCode || 404,
      );
    }
  }

  // 创建通知
  async createNotification(data: CreateNotificationRequest): Promise<Notification> {
    try {
      return await apiClient.post<Notification>("/notifications", data, {
      });
    } catch (error: any) {
      console.error("创建通知失败:", error);
      throw new AppError(
        error.message || "创建通知失败",
        error.code || "CREATE_NOTIFICATION_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 更新通知
  async updateNotification(id: number, data: UpdateNotificationRequest): Promise<Notification> {
    try {
      return await apiClient.put<Notification>(`/notifications/${id}`, data, {
      });
    } catch (error: any) {
      console.error("更新通知失败:", error);
      throw new AppError(
        error.message || "更新通知失败",
        error.code || "UPDATE_NOTIFICATION_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 删除通知
  async deleteNotification(id: number): Promise<{ message: string }> {
    try {
      return await apiClient.delete<{ message: string }>(`/notifications/${id}`, {
      });
    } catch (error: any) {
      console.error("删除通知失败:", error);
      throw new AppError(
        error.message || "删除通知失败",
        error.code || "DELETE_NOTIFICATION_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 获取通知统计信息
  async getNotificationStats(): Promise<NotificationStats> {
    try {
      return await apiClient.get<NotificationStats>("/notifications/stats", {
        
      });
    } catch (error: any) {
      console.error("获取通知统计信息失败:", error);
      throw new AppError(
        error.message || "获取通知统计信息失败",
        error.code || "GET_NOTIFICATION_STATS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 标记通知为已读
  async markAsRead(id: number): Promise<Notification> {
    try {
      return await apiClient.put<Notification>(
        `/notifications/${id}/read`,
        {},
        {
        },
      );
    } catch (error: any) {
      console.error("标记通知为已读失败:", error);
      throw new AppError(
        error.message || "标记通知为已读失败",
        error.code || "MARK_AS_READ_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 批量标记通知为已读
  async batchMarkAsRead(notificationIds: number[]): Promise<{
    success: number;
    failed: number;
    errors?: string[];
  }> {
    try {
      return await apiClient.post<{
        success: number;
        failed: number;
        errors?: string[];
      }>(
        "/notifications/batch/read",
        { notification_ids: notificationIds },
        {
        },
      );
    } catch (error: any) {
      console.error("批量标记通知为已读失败:", error);
      throw new AppError(
        error.message || "批量标记通知为已读失败",
        error.code || "BATCH_MARK_AS_READ_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 获取未读通知
  async getUnreadNotifications(params?: Omit<NotificationListRequest, "read">): Promise<{
    data: Notification[];
    pagination: {
      page: number;
      page_size: number;
      total: number;
      total_pages: number;
    };
  }> {
    try {
      const unreadParams = {
        ...params,
        read: false,
      };
      return await apiClient.getPaginated<Notification>("/notifications/unread", {
        params: unreadParams,
        
      });
    } catch (error: any) {
      console.error("获取未读通知失败:", error);
      throw new AppError(
        error.message || "获取未读通知失败",
        error.code || "GET_UNREAD_NOTIFICATIONS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取已读通知
  async getReadNotifications(params?: Omit<NotificationListRequest, "read">): Promise<{
    data: Notification[];
    pagination: {
      page: number;
      page_size: number;
      total: number;
      total_pages: number;
    };
  }> {
    try {
      const readParams = {
        ...params,
        read: true,
      };
      return await apiClient.getPaginated<Notification>("/notifications/read", {
        params: readParams,
        
      });
    } catch (error: any) {
      console.error("获取已读通知失败:", error);
      throw new AppError(
        error.message || "获取已读通知失败",
        error.code || "GET_READ_NOTIFICATIONS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 搜索通知
  async searchNotifications(
    query: string,
    params?: Omit<NotificationListRequest, "search">,
  ): Promise<{
    data: Notification[];
    pagination: {
      page: number;
      page_size: number;
      total: number;
      total_pages: number;
    };
  }> {
    try {
      const searchParams = {
        ...params,
        search: query,
      };
      return await apiClient.getPaginated<Notification>("/notifications/search", {
        params: searchParams,
        
      });
    } catch (error: any) {
      console.error("搜索通知失败:", error);
      throw new AppError(
        error.message || "搜索通知失败",
        error.code || "SEARCH_NOTIFICATIONS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 删除过期通知
  async deleteExpiredNotifications(): Promise<{
    success: number;
    failed: number;
    errors?: string[];
  }> {
    try {
      return await apiClient.post<{
        success: number;
        failed: number;
        errors?: string[];
      }>(
        "/notifications/cleanup/expired",
        {},
        {
        },
      );
    } catch (error: any) {
      console.error("删除过期通知失败:", error);
      throw new AppError(
        error.message || "删除过期通知失败",
        error.code || "DELETE_EXPIRED_NOTIFICATIONS_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 批量删除通知
  async batchDeleteNotifications(notificationIds: number[]): Promise<{
    success: number;
    failed: number;
    errors?: string[];
  }> {
    try {
      return await apiClient.post<{
        success: number;
        failed: number;
        errors?: string[];
      }>(
        "/notifications/batch/delete",
        { notification_ids: notificationIds },
        {
        },
      );
    } catch (error: any) {
      console.error("批量删除通知失败:", error);
      throw new AppError(
        error.message || "批量删除通知失败",
        error.code || "BATCH_DELETE_NOTIFICATIONS_ERROR",
        error.statusCode || 400,
      );
    }
  }
}

// 导出单例实例
export const notificationService = new NotificationService();

// 为了向后兼容，也导出独立的函数
export const getNotifications = (params?: NotificationListRequest) =>
  notificationService.getNotifications(params);
export const getNotification = (id: number) => notificationService.getNotification(id);
export const createNotification = (data: CreateNotificationRequest) =>
  notificationService.createNotification(data);
export const updateNotification = (id: number, data: UpdateNotificationRequest) =>
  notificationService.updateNotification(id, data);
export const deleteNotification = (id: number) => notificationService.deleteNotification(id);
export const getNotificationStats = () => notificationService.getNotificationStats();
export const markAsRead = (id: number) => notificationService.markAsRead(id);
export const batchMarkAsRead = (notificationIds: number[]) =>
  notificationService.batchMarkAsRead(notificationIds);
export const getUnreadNotifications = (params?: Omit<NotificationListRequest, "read">) =>
  notificationService.getUnreadNotifications(params);
export const getReadNotifications = (params?: Omit<NotificationListRequest, "read">) =>
  notificationService.getReadNotifications(params);
export const searchNotifications = (query: string, params?: Omit<NotificationListRequest, "search">) =>
  notificationService.searchNotifications(query, params);
export const deleteExpiredNotifications = () => notificationService.deleteExpiredNotifications();
export const batchDeleteNotifications = (notificationIds: number[]) =>
  notificationService.batchDeleteNotifications(notificationIds);

export default notificationService;
