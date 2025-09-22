import apiClient from "./api";
import {
  LoginRequest,
  RegisterRequest,
  LoginResponse,
  RefreshTokenResponse,
  UserProfile,
} from "../types";
import { AppError } from "../types/errors";

class AuthService {
  // 用户登录
  async login(data: LoginRequest): Promise<LoginResponse> {
    try {
      const response = await apiClient.post<LoginResponse>(
        "/auth/login",
        data,
        {
        },
      );

      // 登录成功后保存令牌
      if (response.token) {
        localStorage.setItem("token", response.token);
        apiClient.setAuthToken(response.token);
      }

      return response;
    } catch (error: any) {
      console.error("登录失败:", error);
      throw new AppError(
        error.message || "登录失败，请检查用户名和密码",
        error.code || "LOGIN_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 用户注册
  async register(data: RegisterRequest): Promise<LoginResponse> {
    try {
      const response = await apiClient.post<LoginResponse>(
        "/auth/register",
        data,
        {
        },
      );

      // 注册成功后保存令牌
      if (response.token) {
        localStorage.setItem("token", response.token);
        apiClient.setAuthToken(response.token);
      }

      return response;
    } catch (error: any) {
      console.error("注册失败:", error);
      throw new AppError(
        error.message || "注册失败，请稍后重试",
        error.code || "REGISTER_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 刷新令牌
  async refreshToken(token: string): Promise<RefreshTokenResponse> {
    try {
      const response = await apiClient.post<RefreshTokenResponse>(
        "/auth/refresh",
        { token },
        {
        },
      );

      // 更新本地令牌
      if (response.token) {
        localStorage.setItem("token", response.token);
        apiClient.setAuthToken(response.token);
      }

      return response;
    } catch (error: any) {
      console.error("刷新令牌失败:", error);
      throw new AppError(
        error.message || "令牌刷新失败，请重新登录",
        error.code || "TOKEN_REFRESH_ERROR",
        error.statusCode || 401,
      );
    }
  }

  // 获取当前用户资料
  async getCurrentUser(): Promise<UserProfile> {
    try {
      return await apiClient.get<UserProfile>("/users/profile", {
        useCache: true,
        cacheTTL: 5 * 60 * 1000, // 5分钟缓存
      });
    } catch (error: any) {
      console.error("获取用户资料失败:", error);
      throw new AppError(
        error.message || "获取用户资料失败",
        error.code || "GET_USER_ERROR",
        error.statusCode || 401,
      );
    }
  }

  // 更新用户资料
  async updateProfile(data: Partial<UserProfile>): Promise<UserProfile> {
    try {
      return await apiClient.put<UserProfile>("/users/me", data, {
      });
    } catch (error: any) {
      console.error("更新用户资料失败:", error);
      throw new AppError(
        error.message || "更新用户资料失败",
        error.code || "UPDATE_USER_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 修改密码
  async changePassword(
    currentPassword: string,
    newPassword: string,
  ): Promise<void> {
    try {
      await apiClient.post(
        "/users/change-password",
        {
          current_password: currentPassword,
          new_password: newPassword,
        },
        {
        },
      );
    } catch (error: any) {
      console.error("修改密码失败:", error);
      throw new AppError(
        error.message || "修改密码失败",
        error.code || "CHANGE_PASSWORD_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 用户登出
  async logout(): Promise<void> {
    try {
      await apiClient.post("/auth/logout", {}, {
        timeout: 5000, // 5秒超时
      });
    } catch (error) {
      // 即使登出请求失败，也要清除本地令牌
      console.warn("登出请求失败:", error);
    } finally {
      this.clearAuth();
    }
  }

  // 清除认证信息
  clearAuth(): void {
    localStorage.removeItem("token");
    localStorage.removeItem("refreshToken");
    apiClient.setAuthToken(null);
  }

  // 检查是否已认证
  isAuthenticated(): boolean {
    const token = localStorage.getItem("token");
    return !!token;
  }

  // 获取存储的令牌
  getToken(): string | null {
    return localStorage.getItem("token");
  }

  // 设置令牌
  setToken(token: string): void {
    localStorage.setItem("token", token);
    apiClient.setAuthToken(token);
  }

  // 获取用户权限信息
  async getUserPermissions(): Promise<string[]> {
    // 暂时返回默认权限，等后端支持权限系统
    try {
      // 这里可以基于用户角色返回默认权限
      const userStr = localStorage.getItem("user");
      if (userStr) {
        const user = JSON.parse(userStr);
        if (user.role === "admin") {
          return ["read", "write", "delete", "manage_users"];
        } else if (user.role === "lawyer") {
          return ["read", "write", "manage_cases"];
        }
      }
      return ["read"];
    } catch (error) {
      console.error("获取用户权限失败:", error);
      return ["read"];
    }
  }

  // 检查用户是否有特定权限
  async hasPermission(permission: string): Promise<boolean> {
    const permissions = await this.getUserPermissions();
    return permissions.includes(permission);
  }

  // 检查用户是否有特定角色
  hasRole(role: string): boolean {
    const userStr = localStorage.getItem("user");
    if (!userStr) return false;

    try {
      const user = JSON.parse(userStr);
      return user.role === role;
    } catch {
      return false;
    }
  }

  // 验证令牌是否有效
  async validateToken(): Promise<boolean> {
    try {
      await this.getCurrentUser();
      return true;
    } catch {
      return false;
    }
  }

  // 忘记密码 - 发送重置链接
  async forgotPassword(email: string): Promise<void> {
    try {
      await apiClient.post("/auth/forgot-password", { email }, {
      });
    } catch (error: any) {
      console.error("发送重置密码邮件失败:", error);
      throw new AppError(
        error.message || "发送重置密码邮件失败",
        error.code || "FORGOT_PASSWORD_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 重置密码
  async resetPassword(token: string, newPassword: string): Promise<void> {
    try {
      await apiClient.post(
        "/auth/reset-password",
        { token, new_password: newPassword },
        {
        },
      );
    } catch (error: any) {
      console.error("重置密码失败:", error);
      throw new AppError(
        error.message || "重置密码失败",
        error.code || "RESET_PASSWORD_ERROR",
        error.statusCode || 400,
      );
    }
  }
}

// 导出单例实例
export const authService = new AuthService();

// 为了向后兼容，也导出独立的函数
export const login = (data: LoginRequest) => authService.login(data);
export const register = (data: RegisterRequest) => authService.register(data);
export const refreshTokens = (token: string) => authService.refreshToken(token);
export const getUserProfile = () => authService.getCurrentUser();
export const updateUserProfile = (data: Partial<UserProfile>) =>
  authService.updateProfile(data);
export const changePassword = (currentPassword: string, newPassword: string) =>
  authService.changePassword(currentPassword, newPassword);
export const logout = () => authService.logout();
export const forgotPassword = (email: string) =>
  authService.forgotPassword(email);
export const resetPassword = (token: string, newPassword: string) =>
  authService.resetPassword(token, newPassword);

export default authService;
