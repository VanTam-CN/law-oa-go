import apiClient from "./api";
import { AppError } from "../types/errors";
import { Message, MessageListRequest, SendMessageRequest, UpdateMessageRequest } from "../types";

interface MessageStats {
  total: number;
  unread: number;
  sent: number;
  received: number;
  by_type: {
    text: number;
    file: number;
    system: number;
  };
  by_status: {
    sent: number;
    delivered: number;
    read: number;
  };
}

class MessageService {
  // 获取消息列表
  async getMessages(params?: MessageListRequest): Promise<{
    data: Message[];
    pagination: {
      page: number;
      page_size: number;
      total: number;
      total_pages: number;
    };
  }> {
    try {
      return await apiClient.getPaginated<Message>("/messages", {
        params,
        
      });
    } catch (error: any) {
      console.error("获取消息列表失败:", error);
      throw new AppError(
        error.message || "获取消息列表失败",
        error.code || "GET_MESSAGES_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取消息详情
  async getMessage(id: number): Promise<Message> {
    try {
      return await apiClient.get<Message>(`/messages/${id}`, {
        
      });
    } catch (error: any) {
      console.error("获取消息详情失败:", error);
      throw new AppError(
        error.message || "获取消息详情失败",
        error.code || "GET_MESSAGE_ERROR",
        error.statusCode || 404,
      );
    }
  }

  // 发送消息
  async sendMessage(data: SendMessageRequest): Promise<Message> {
    try {
      return await apiClient.post<Message>("/messages", data, {
      });
    } catch (error: any) {
      console.error("发送消息失败:", error);
      throw new AppError(
        error.message || "发送消息失败",
        error.code || "SEND_MESSAGE_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 更新消息
  async updateMessage(id: number, data: UpdateMessageRequest): Promise<Message> {
    try {
      return await apiClient.put<Message>(`/messages/${id}`, data, {
      });
    } catch (error: any) {
      console.error("更新消息失败:", error);
      throw new AppError(
        error.message || "更新消息失败",
        error.code || "UPDATE_MESSAGE_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 标记消息为已读
  async markAsRead(id: number): Promise<Message> {
    try {
      return await apiClient.put<Message>(`/messages/${id}/read`, {}, {
      });
    } catch (error: any) {
      console.error("标记消息已读失败:", error);
      throw new AppError(
        error.message || "标记消息已读失败",
        error.code || "MARK_AS_READ_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 删除消息
  async deleteMessage(id: number): Promise<{ message: string }> {
    try {
      return await apiClient.delete<{ message: string }>(`/messages/${id}`, {
      });
    } catch (error: any) {
      console.error("删除消息失败:", error);
      throw new AppError(
        error.message || "删除消息失败",
        error.code || "DELETE_MESSAGE_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 获取消息统计信息
  async getMessageStats(): Promise<MessageStats> {
    try {
      return await apiClient.get<MessageStats>("/messages/stats", {
        
      });
    } catch (error: any) {
      console.error("获取消息统计信息失败:", error);
      throw new AppError(
        error.message || "获取消息统计信息失败",
        error.code || "GET_MESSAGE_STATS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 批量标记消息为已读
  async batchMarkAsRead(messageIds: number[]): Promise<{
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
        "/messages/batch/read",
        { message_ids: messageIds },
        {
        },
      );
    } catch (error: any) {
      console.error("批量标记消息为已读失败:", error);
      throw new AppError(
        error.message || "批量标记消息为已读失败",
        error.code || "BATCH_MARK_AS_READ_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 获取未读消息
  async getUnreadMessages(params?: Omit<MessageListRequest, "status">): Promise<{
    data: Message[];
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
        status: "sent",
      };
      return await apiClient.getPaginated<Message>("/messages/unread", {
        params: unreadParams,
        
      });
    } catch (error: any) {
      console.error("获取未读消息失败:", error);
      throw new AppError(
        error.message || "获取未读消息失败",
        error.code || "GET_UNREAD_MESSAGES_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取已发送消息
  async getSentMessages(params?: Omit<MessageListRequest, "sender_id">): Promise<{
    data: Message[];
    pagination: {
      page: number;
      page_size: number;
      total: number;
      total_pages: number;
    };
  }> {
    try {
      return await apiClient.getPaginated<Message>("/messages/sent", {
        params,
        
      });
    } catch (error: any) {
      console.error("获取已发送消息失败:", error);
      throw new AppError(
        error.message || "获取已发送消息失败",
        error.code || "GET_SENT_MESSAGES_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 搜索消息
  async searchMessages(
    query: string,
    params?: Omit<MessageListRequest, "search">,
  ): Promise<{
    data: Message[];
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
      return await apiClient.getPaginated<Message>("/messages/search", {
        params: searchParams,
        
      });
    } catch (error: any) {
      console.error("搜索消息失败:", error);
      throw new AppError(
        error.message || "搜索消息失败",
        error.code || "SEARCH_MESSAGES_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取与特定用户的对话
  async getConversation(userId: number, params?: Omit<MessageListRequest, "receiver_id" | "sender_id">): Promise<{
    data: Message[];
    pagination: {
      page: number;
      page_size: number;
      total: number;
      total_pages: number;
    };
  }> {
    try {
      return await apiClient.getPaginated<Message>(`/messages/conversation/${userId}`, {
        params,
        
      });
    } catch (error: any) {
      console.error("获取对话失败:", error);
      throw new AppError(
        error.message || "获取对话失败",
        error.code || "GET_CONVERSATION_ERROR",
        error.statusCode || 500,
      );
    }
  }
}

// 导出单例实例
export const messageService = new MessageService();

// 为了向后兼容，也导出独立的函数
export const getMessages = (params?: MessageListRequest) =>
  messageService.getMessages(params);
export const getMessage = (id: number) => messageService.getMessage(id);
export const sendMessage = (data: SendMessageRequest) =>
  messageService.sendMessage(data);
export const updateMessage = (id: number, data: UpdateMessageRequest) =>
  messageService.updateMessage(id, data);
export const deleteMessage = (id: number) => messageService.deleteMessage(id);
export const getMessageStats = () => messageService.getMessageStats();
export const markAsRead = (id: number) => messageService.markAsRead(id);
export const batchMarkAsRead = (messageIds: number[]) =>
  messageService.batchMarkAsRead(messageIds);
export const getUnreadMessages = (params?: Omit<MessageListRequest, "status">) =>
  messageService.getUnreadMessages(params);
export const getSentMessages = (params?: Omit<MessageListRequest, "sender_id">) =>
  messageService.getSentMessages(params);
export const searchMessages = (query: string, params?: Omit<MessageListRequest, "search">) =>
  messageService.searchMessages(query, params);
export const getConversation = (userId: number, params?: Omit<MessageListRequest, "receiver_id" | "sender_id">) =>
  messageService.getConversation(userId, params);

export default messageService;
