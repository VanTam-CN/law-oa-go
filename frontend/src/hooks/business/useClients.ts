import { useState, useEffect, useCallback } from "react";
import { useSelector, useDispatch } from "react-redux";
import { RootState } from "../../store";
import { clientService } from "../../services/clientService";
import {
  Client,
  ClientListRequest,
  CreateClientRequest,
  UpdateClientRequest,
} from "../../types";
import { AppError } from "../../types/errors";

// 客户列表返回类型
interface ClientListResponse {
  data: Client[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
}

// Hook 返回类型
interface UseClientsReturn {
  // 数据状态
  clients: Client[];
  loading: boolean;
  error: string | null;

  // 分页状态
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };

  // 操作方法
  fetchClients: (params?: ClientListRequest) => Promise<void>;
  createClient: (data: CreateClientRequest) => Promise<Client>;
  updateClient: (id: number, data: UpdateClientRequest) => Promise<Client>;
  deleteClient: (id: number) => Promise<void>;
  getClient: (id: number) => Promise<Client>;
  searchClients: (
    query: string,
    params?: Omit<ClientListRequest, "search">,
  ) => Promise<ClientListResponse>;

  // 工具方法
  refresh: () => Promise<void>;
  clearError: () => void;
}

/**
 * 客户管理 Hook
 * 提供客户相关的数据获取、CRUD 操作和状态管理
 */
export const useClients = (
  defaultParams?: ClientListRequest,
): UseClientsReturn => {
  const dispatch = useDispatch();

  // Redux 状态
  const { user } = useSelector((state: RootState) => state.auth);

  // 本地状态
  const [clients, setClients] = useState<Client[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [pagination, setPagination] = useState({
    page: 1,
    page_size: 10,
    total: 0,
    total_pages: 0,
  });

  // 获取客户列表
  const fetchClients = useCallback(
    async (params?: ClientListRequest) => {
      try {
        setLoading(true);
        setError(null);

        const mergedParams = { ...defaultParams, ...params };
        const response = await clientService.getClients(mergedParams);

        setClients(response.data);
        setPagination(response.pagination);
      } catch (err: any) {
        const appError =
          err instanceof AppError ? err : AppError.fromApiError(err);
        setError(appError.message);
        console.error("获取客户列表失败:", err);
      } finally {
        setLoading(false);
      }
    },
    [defaultParams],
  );

  // 创建客户
  const createClient = useCallback(
    async (data: CreateClientRequest): Promise<Client> => {
      try {
        setLoading(true);
        setError(null);

        const newClient = await clientService.createClient(data);

        // 更新本地列表
        setClients((prev) => [newClient, ...prev]);

        return newClient;
      } catch (err: any) {
        const appError =
          err instanceof AppError ? err : AppError.fromApiError(err);
        setError(appError.message);
        console.error("创建客户失败:", err);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  // 更新客户
  const updateClient = useCallback(
    async (id: number, data: UpdateClientRequest): Promise<Client> => {
      try {
        setLoading(true);
        setError(null);

        const updatedClient = await clientService.updateClient(id, data);

        // 更新本地列表
        setClients((prev) =>
          prev.map((client) => (client.id === id ? updatedClient : client)),
        );

        return updatedClient;
      } catch (err: any) {
        const appError =
          err instanceof AppError ? err : AppError.fromApiError(err);
        setError(appError.message);
        console.error("更新客户失败:", err);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  // 删除客户
  const deleteClient = useCallback(async (id: number): Promise<void> => {
    try {
      setLoading(true);
      setError(null);

      await clientService.deleteClient(id);

      // 更新本地列表
      setClients((prev) => prev.filter((client) => client.id !== id));
    } catch (err: any) {
      const appError =
        err instanceof AppError ? err : AppError.fromApiError(err);
      setError(appError.message);
      console.error("删除客户失败:", err);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // 获取单个客户
  const getClient = useCallback(async (id: number): Promise<Client> => {
    try {
      setLoading(true);
      setError(null);

      return await clientService.getClient(id);
    } catch (err: any) {
      const appError =
        err instanceof AppError ? err : AppError.fromApiError(err);
      setError(appError.message);
      console.error("获取客户详情失败:", err);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // 搜索客户
  const searchClients = useCallback(
    async (
      query: string,
      params?: Omit<ClientListRequest, "search">,
    ): Promise<ClientListResponse> => {
      try {
        setLoading(true);
        setError(null);

        const response = await clientService.searchClients(query, params);

        // 更新本地列表
        setClients(response.data);
        setPagination(response.pagination);

        return response;
      } catch (err: any) {
        const appError =
          err instanceof AppError ? err : AppError.fromApiError(err);
        setError(appError.message);
        console.error("搜索客户失败:", err);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  // 刷新数据
  const refresh = useCallback(async (): Promise<void> => {
    await fetchClients();
  }, [fetchClients]);

  // 清除错误
  const clearError = useCallback((): void => {
    setError(null);
  }, []);

  // 初始化加载
  useEffect(() => {
    if (user) {
      fetchClients();
    }
  }, [user, fetchClients]);

  return {
    // 数据状态
    clients,
    loading,
    error,

    // 分页状态
    pagination,

    // 操作方法
    fetchClients,
    createClient,
    updateClient,
    deleteClient,
    getClient,
    searchClients,

    // 工具方法
    refresh,
    clearError,
  };
};

export default useClients;
export type { UseClientsReturn };
