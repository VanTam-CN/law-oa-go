import apiClient from "./api";
import {
  UserProfile,
  UserListRequest,
  UserListResponse,
  CreateUserRequest,
  UpdateUserRequest,
} from "../types";
import { AppError } from "../types/errors";

class UserService {
  // 获取用户列表
  async getUsers(params: UserListRequest): Promise<UserListResponse> {
    try {
      return await apiClient.get<UserListResponse>("/admin/users", {
        params,
        useCache: true,
        cacheTTL: 2 * 60 * 1000, // 2分钟缓存
      });
    } catch (error: any) {
      console.error("获取用户列表失败:", error);
      throw new AppError(
        error.message || "获取用户列表失败",
        error.code || "GET_USERS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取用户详情
  async getUser(id: number): Promise<UserProfile> {
    try {
      return await apiClient.get<UserProfile>(`/admin/users/${id}`, {
        useCache: true,
        cacheTTL: 5 * 60 * 1000, // 5分钟缓存
      });
    } catch (error: any) {
      console.error("获取用户详情失败:", error);
      throw new AppError(
        error.message || "获取用户详情失败",
        error.code || "GET_USER_ERROR",
        error.statusCode || 404,
      );
    }
  }

  // 创建用户
  async createUser(data: CreateUserRequest): Promise<UserProfile> {
    try {
      return await apiClient.post<UserProfile>("/admin/users", data, {
      });
    } catch (error: any) {
      console.error("创建用户失败:", error);
      throw new AppError(
        error.message || "创建用户失败",
        error.code || "CREATE_USER_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 更新用户
  async updateUser(id: number, data: UpdateUserRequest): Promise<UserProfile> {
    try {
      return await apiClient.put<UserProfile>(`/admin/users/${id}`, data, {
      });
    } catch (error: any) {
      console.error("更新用户失败:", error);
      throw new AppError(
        error.message || "更新用户失败",
        error.code || "UPDATE_USER_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 删除用户
  async deleteUser(id: number): Promise<{ message: string }> {
    try {
      return await apiClient.delete<{ message: string }>(`/admin/users/${id}`, {
      });
    } catch (error: any) {
      console.error("删除用户失败:", error);
      throw new AppError(
        error.message || "删除用户失败",
        error.code || "DELETE_USER_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 获取用户统计信息
  async getUserStats(): Promise<{
    total_users: number;
    active_users: number;
    inactive_users: number;
    new_users_this_month: number;
    users_by_role: Array<{
      role: string;
      count: number;
    }>;
  }> {
    try {
      return await apiClient.get<{
        total_users: number;
        active_users: number;
        inactive_users: number;
        new_users_this_month: number;
        users_by_role: Array<{
          role: string;
          count: number;
        }>;
      }>("/admin/users/stats", {
        useCache: true,
        cacheTTL: 3 * 60 * 1000, // 3分钟缓存
      });
    } catch (error: any) {
      console.error("获取用户统计信息失败:", error);
      throw new AppError(
        error.message || "获取用户统计信息失败",
        error.code || "GET_USER_STATS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 搜索用户
  async searchUsers(
    query: string,
    params?: Omit<UserListRequest, "search">,
  ): Promise<UserListResponse> {
    try {
      const searchParams = {
        ...params,
        search: query,
      };
      return await apiClient.get<UserListResponse>("/admin/users/search", {
        params: searchParams,
        useCache: true,
        cacheTTL: 1 * 60 * 1000, // 1分钟缓存
      });
    } catch (error: any) {
      console.error("搜索用户失败:", error);
      throw new AppError(
        error.message || "搜索用户失败",
        error.code || "SEARCH_USERS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取用户角色列表
  async getUserRoles(): Promise<
    Array<{
      id: string;
      name: string;
      description: string;
      permissions: string[];
    }>
  > {
    try {
      return await apiClient.get<
        Array<{
          id: string;
          name: string;
          description: string;
          permissions: string[];
        }>
      >("/admin/users/roles", {
        useCache: true,
        cacheTTL: 10 * 60 * 1000, // 10分钟缓存
      });
    } catch (error: any) {
      console.error("获取用户角色列表失败:", error);
      throw new AppError(
        error.message || "获取用户角色列表失败",
        error.code || "GET_USER_ROLES_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 更新用户角色
  async updateUserRole(
    userId: number,
    roleId: string,
  ): Promise<{ message: string }> {
    try {
      return await apiClient.post<{ message: string }>(
        `/admin/users/${userId}/role`,
        { role_id: roleId },
        {
        },
      );
    } catch (error: any) {
      console.error("更新用户角色失败:", error);
      throw new AppError(
        error.message || "更新用户角色失败",
        error.code || "UPDATE_USER_ROLE_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 重置用户密码
  async resetUserPassword(
    userId: number,
    newPassword: string,
  ): Promise<{ message: string }> {
    try {
      return await apiClient.post<{ message: string }>(
        `/admin/users/${userId}/reset-password`,
        { new_password: newPassword },
        {
        },
      );
    } catch (error: any) {
      console.error("重置用户密码失败:", error);
      throw new AppError(
        error.message || "重置用户密码失败",
        error.code || "RESET_USER_PASSWORD_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 批量删除用户
  async batchDeleteUsers(userIds: number[]): Promise<{
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
        "/admin/users/batch/delete",
        { user_ids: userIds },
        {
        },
      );
    } catch (error: any) {
      console.error("批量删除用户失败:", error);
      throw new AppError(
        error.message || "批量删除用户失败",
        error.code || "BATCH_DELETE_USERS_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 批量更新用户状态
  async batchUpdateUserStatus(
    userIds: number[],
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
        "/admin/users/batch/status",
        { user_ids: userIds, status },
        {
        },
      );
    } catch (error: any) {
      console.error("批量更新用户状态失败:", error);
      throw new AppError(
        error.message || "批量更新用户状态失败",
        error.code || "BATCH_UPDATE_USER_STATUS_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 导出用户数据
  async exportUsers(params?: UserListRequest): Promise<Blob> {
    try {
      const response = await apiClient.getClient().get("/admin/users/export", {
        params,
        responseType: "blob",
      });
      return response.data;
    } catch (error: any) {
      console.error("导出用户数据失败:", error);
      throw new AppError(
        error.message || "导出用户数据失败",
        error.code || "EXPORT_USERS_ERROR",
        error.statusCode || 500,
      );
    }
  }
}

// 导出单例实例
export const userService = new UserService();

// 为了向后兼容，也导出独立的函数
export const getUsers = (params: UserListRequest) =>
  userService.getUsers(params);
export const getUser = (id: number) => userService.getUser(id);
export const getUserById = (id: number) => userService.getUser(id); // 别名，为了向后兼容
export const createUser = (data: CreateUserRequest) =>
  userService.createUser(data);
export const updateUser = (id: number, data: UpdateUserRequest) =>
  userService.updateUser(id, data);
export const deleteUser = (id: number) => userService.deleteUser(id);
export const getUserStats = () => userService.getUserStats();
export const searchUsers = (
  query: string,
  params?: Omit<UserListRequest, "search">,
) => userService.searchUsers(query, params);
export const getUserRoles = () => userService.getUserRoles();
export const updateUserRole = (userId: number, roleId: string) =>
  userService.updateUserRole(userId, roleId);
export const resetUserPassword = (userId: number, newPassword: string) =>
  userService.resetUserPassword(userId, newPassword);
export const batchDeleteUsers = (userIds: number[]) =>
  userService.batchDeleteUsers(userIds);
export const batchUpdateUserStatus = (userIds: number[], status: string) =>
  userService.batchUpdateUserStatus(userIds, status);
export const exportUsers = (params?: UserListRequest) =>
  userService.exportUsers(params);

export default userService;
