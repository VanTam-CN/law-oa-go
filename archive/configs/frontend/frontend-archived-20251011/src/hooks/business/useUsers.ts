import { useState, useEffect, useCallback } from "react";
import { useSelector, useDispatch } from "react-redux";
import { RootState } from "../../store";
import { userService } from "../../services/userService";
import {
  UserProfile,
  UserListRequest,
  CreateUserRequest,
  UpdateUserRequest,
} from "../../types";
import { AppError } from "../../types/errors";

// Hook 返回类型
interface UseUsersReturn {
  // 数据状态
  users: UserProfile[];
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
  fetchUsers: (params: UserListRequest) => Promise<void>;
  createUser: (data: CreateUserRequest) => Promise<UserProfile>;
  updateUser: (id: number, data: UpdateUserRequest) => Promise<UserProfile>;
  deleteUser: (id: number) => Promise<void>;
  getUser: (id: number) => Promise<UserProfile>;

  // 批量操作
  batchDeleteUsers: (
    userIds: number[],
  ) => Promise<{ success: number; failed: number; errors?: string[] }>;
  batchUpdateStatus: (
    userIds: number[],
    status: string,
  ) => Promise<{ success: number; failed: number; errors?: string[] }>;

  // 工具方法
  refresh: () => Promise<void>;
  clearError: () => void;
}

/**
 * 用户管理 Hook
 * 提供用户相关的数据获取、CRUD 操作和状态管理
 */
export const useUsers = (defaultParams?: UserListRequest): UseUsersReturn => {
  const dispatch = useDispatch();

  // Redux 状态
  const { user } = useSelector((state: RootState) => state.auth);

  // 本地状态
  const [users, setUsers] = useState<UserProfile[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [pagination, setPagination] = useState({
    page: 1,
    page_size: 10,
    total: 0,
    total_pages: 0,
  });

  // 获取用户列表
  const fetchUsers = useCallback(
    async (params: UserListRequest) => {
      try {
        setLoading(true);
        setError(null);

        const mergedParams = { ...defaultParams, ...params };
        const response = await userService.getUsers(mergedParams);

        setUsers(response.data);
        setPagination(response.pagination);
      } catch (err: any) {
        const appError =
          err instanceof AppError ? err : AppError.fromApiError(err);
        setError(appError.message);
        console.error("获取用户列表失败:", err);
      } finally {
        setLoading(false);
      }
    },
    [defaultParams],
  );

  // 创建用户
  const createUser = useCallback(
    async (data: CreateUserRequest): Promise<UserProfile> => {
      try {
        setLoading(true);
        setError(null);

        const newUser = await userService.createUser(data);

        // 更新本地列表
        setUsers((prev) => [newUser, ...prev]);

        return newUser;
      } catch (err: any) {
        const appError =
          err instanceof AppError ? err : AppError.fromApiError(err);
        setError(appError.message);
        console.error("创建用户失败:", err);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  // 更新用户
  const updateUser = useCallback(
    async (id: number, data: UpdateUserRequest): Promise<UserProfile> => {
      try {
        setLoading(true);
        setError(null);

        const updatedUser = await userService.updateUser(id, data);

        // 更新本地列表
        setUsers((prev) =>
          prev.map((userItem) => (userItem.id === id ? updatedUser : userItem)),
        );

        return updatedUser;
      } catch (err: any) {
        const appError =
          err instanceof AppError ? err : AppError.fromApiError(err);
        setError(appError.message);
        console.error("更新用户失败:", err);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  // 删除用户
  const deleteUser = useCallback(async (id: number): Promise<void> => {
    try {
      setLoading(true);
      setError(null);

      await userService.deleteUser(id);

      // 更新本地列表
      setUsers((prev) => prev.filter((userItem) => userItem.id !== id));
    } catch (err: any) {
      const appError =
        err instanceof AppError ? err : AppError.fromApiError(err);
      setError(appError.message);
      console.error("删除用户失败:", err);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // 获取单个用户
  const getUser = useCallback(async (id: number): Promise<UserProfile> => {
    try {
      setLoading(true);
      setError(null);

      return await userService.getUser(id);
    } catch (err: any) {
      const appError =
        err instanceof AppError ? err : AppError.fromApiError(err);
      setError(appError.message);
      console.error("获取用户详情失败:", err);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // 批量删除用户
  const batchDeleteUsers = useCallback(
    async (
      userIds: number[],
    ): Promise<{ success: number; failed: number; errors?: string[] }> => {
      try {
        setLoading(true);
        setError(null);

        const result = await userService.batchDeleteUsers(userIds);

        // 更新本地列表
        if (result.success > 0) {
          setUsers((prev) =>
            prev.filter((userItem) => !userIds.includes(userItem.id)),
          );
        }

        return result;
      } catch (err: any) {
        const appError =
          err instanceof AppError ? err : AppError.fromApiError(err);
        setError(appError.message);
        console.error("批量删除用户失败:", err);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  // 批量更新用户状态
  const batchUpdateStatus = useCallback(
    async (
      userIds: number[],
      status: string,
    ): Promise<{ success: number; failed: number; errors?: string[] }> => {
      try {
        setLoading(true);
        setError(null);

        const result = await userService.batchUpdateUserStatus(userIds, status);

        // 更新本地列表
        if (result.success > 0) {
          setUsers((prev) =>
            prev.map((userItem) =>
              userIds.includes(userItem.id)
                ? { ...userItem, status: status as any }
                : userItem,
            ),
          );
        }

        return result;
      } catch (err: any) {
        const appError =
          err instanceof AppError ? err : AppError.fromApiError(err);
        setError(appError.message);
        console.error("批量更新用户状态失败:", err);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  // 刷新数据
  const refresh = useCallback(async (): Promise<void> => {
    if (defaultParams) {
      await fetchUsers(defaultParams);
    }
  }, [fetchUsers, defaultParams]);

  // 清除错误
  const clearError = useCallback((): void => {
    setError(null);
  }, []);

  return {
    // 数据状态
    users,
    loading,
    error,

    // 分页状态
    pagination,

    // 操作方法
    fetchUsers,
    createUser,
    updateUser,
    deleteUser,
    getUser,

    // 批量操作
    batchDeleteUsers,
    batchUpdateStatus,

    // 工具方法
    refresh,
    clearError,
  };
};

export default useUsers;
export type { UseUsersReturn };
