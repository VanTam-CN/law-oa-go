import apiClient from "./api";
import {
  Client,
  ClientListRequest,
  CreateClientRequest,
  UpdateClientRequest,
  ClientStats,
} from "../types";
import { AppError } from "../types/errors";

class ClientService {
  // 获取客户列表（分页）
  async getClients(params?: ClientListRequest): Promise<{
    data: Client[];
    pagination: {
      page: number;
      page_size: number;
      total: number;
      total_pages: number;
    };
  }> {
    try {
      return await apiClient.getPaginated<Client>("/clients", {
        params,
        
      });
    } catch (error: any) {
      console.error("获取客户列表失败:", error);
      throw new AppError(
        error.message || "获取客户列表失败",
        error.code || "GET_CLIENTS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取客户详情
  async getClient(id: number): Promise<Client> {
    try {
      return await apiClient.get<Client>(`/clients/${id}`, {
        
      });
    } catch (error: any) {
      console.error("获取客户详情失败:", error);
      throw new AppError(
        error.message || "获取客户详情失败",
        error.code || "GET_CLIENT_ERROR",
        error.statusCode || 404,
      );
    }
  }

  // 创建客户
  async createClient(data: CreateClientRequest): Promise<Client> {
    try {
      return await apiClient.post<Client>("/clients", data, {
      });
    } catch (error: any) {
      console.error("创建客户失败:", error);
      throw new AppError(
        error.message || "创建客户失败",
        error.code || "CREATE_CLIENT_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 更新客户
  async updateClient(id: number, data: UpdateClientRequest): Promise<Client> {
    try {
      return await apiClient.put<Client>(`/clients/${id}`, data, {
      });
    } catch (error: any) {
      console.error("更新客户失败:", error);
      throw new AppError(
        error.message || "更新客户失败",
        error.code || "UPDATE_CLIENT_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 删除客户
  async deleteClient(id: number): Promise<{ message: string }> {
    try {
      return await apiClient.delete<{ message: string }>(`/clients/${id}`, {
      });
    } catch (error: any) {
      console.error("删除客户失败:", error);
      throw new AppError(
        error.message || "删除客户失败",
        error.code || "DELETE_CLIENT_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 获取客户统计信息
  async getClientStats(): Promise<ClientStats> {
    try {
      return await apiClient.get<ClientStats>("/clients/stats", {
        
      });
    } catch (error: any) {
      console.error("获取客户统计信息失败:", error);
      throw new AppError(
        error.message || "获取客户统计信息失败",
        error.code || "GET_CLIENT_STATS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 搜索客户
  async searchClients(
    query: string,
    params?: Omit<ClientListRequest, "search">,
  ): Promise<{
    data: Client[];
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
      return await apiClient.getPaginated<Client>("/clients/search", {
        params: searchParams,
        
      });
    } catch (error: any) {
      console.error("搜索客户失败:", error);
      throw new AppError(
        error.message || "搜索客户失败",
        error.code || "SEARCH_CLIENTS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取客户的案件
  async getClientCases(
    client_id: number,
    params?: {
      page?: number;
      page_size?: number;
      status?: string;
      case_type?: string;
    },
  ): Promise<{
    data: Array<{
      id: number;
      title: string;
      case_type: string;
      priority: string;
      status: string;
      start_date: string;
      lawyer_id: number | null;
      lawyer_name: string | null;
    }>;
    pagination: {
      page: number;
      page_size: number;
      total: number;
      total_pages: number;
    };
  }> {
    try {
      return await apiClient.getPaginated<{
        id: number;
        title: string;
        case_type: string;
        priority: string;
        status: string;
        start_date: string;
        lawyer_id: number | null;
        lawyer_name: string | null;
      }>(`/clients/${client_id}/cases`, {
        params,
        
      });
    } catch (error: any) {
      console.error("获取客户案件失败:", error);
      throw new AppError(
        error.message || "获取客户案件失败",
        error.code || "GET_CLIENT_CASES_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 批量删除客户
  async batchDeleteClients(client_ids: number[]): Promise<{
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
        "/clients/batch/delete",
        { client_ids: client_ids },
        {
        },
      );
    } catch (error: any) {
      console.error("批量删除客户失败:", error);
      throw new AppError(
        error.message || "批量删除客户失败",
        error.code || "BATCH_DELETE_CLIENTS_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 批量更新客户状态
  async batchUpdateStatus(
    client_ids: number[],
    status: string,
  ): Promise<{
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
        "/clients/batch/status",
        { client_ids: client_ids, status },
        {
        },
      );
    } catch (error: any) {
      console.error("批量更新客户状态失败:", error);
      throw new AppError(
        error.message || "批量更新客户状态失败",
        error.code || "BATCH_UPDATE_STATUS_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 导出客户数据
  async exportClients(params?: ClientListRequest): Promise<Blob> {
    try {
      const response = await apiClient.getClient().get("/clients/export", {
        params,
        responseType: "blob",
      });
      return response.data;
    } catch (error: any) {
      console.error("导出客户数据失败:", error);
      throw new AppError(
        error.message || "导出客户数据失败",
        error.code || "EXPORT_CLIENTS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取客户状态列表
  async getClientStatuses(): Promise<
    Array<{
      id: string;
      name: string;
      description: string;
      color: string;
    }>
  > {
    try {
      return await apiClient.get<
        Array<{
          id: string;
          name: string;
          description: string;
          color: string;
        }>
      >("/clients/statuses", {
        
      });
    } catch (error: any) {
      console.error("获取客户状态列表失败:", error);
      throw new AppError(
        error.message || "获取客户状态列表失败",
        error.code || "GET_CLIENT_STATUSES_ERROR",
        error.statusCode || 500,
      );
    }
  }
}

// 导出单例实例
export const clientService = new ClientService();

// 为了向后兼容，也导出独立的函数
export const getClients = (params?: ClientListRequest) =>
  clientService.getClients(params);
export const getClient = (id: number) => clientService.getClient(id);
export const createClient = (data: CreateClientRequest) =>
  clientService.createClient(data);
export const updateClient = (id: number, data: UpdateClientRequest) =>
  clientService.updateClient(id, data);
export const deleteClient = (id: number) => clientService.deleteClient(id);
export const getClientStats = () => clientService.getClientStats();
export const getClientCases = (
  client_id: number,
  params?: {
    page?: number;
    page_size?: number;
    status?: string;
    case_type?: string;
  },
) => clientService.getClientCases(client_id, params);

export default clientService;
