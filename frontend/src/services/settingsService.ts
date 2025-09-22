import apiClient from "./api";
import { AppError } from "../types/errors";

interface SystemSettings {
  id: number;
  site_name: string;
  site_description: string;
  logo_url?: string;
  favicon_url?: string;
  contact_email: string;
  contact_phone: string;
  contact_address: string;
  timezone: string;
  language: string;
  date_format: string;
  time_format: string;
  currency: string;
  theme: 'light' | 'dark' | 'auto';
  registration_enabled: boolean;
  email_verification_required: boolean;
  max_file_size: number;
  allowed_file_types: string[];
  session_timeout: number;
  password_policy: {
    min_length: number;
    require_uppercase: boolean;
    require_lowercase: boolean;
    require_numbers: boolean;
    require_special_chars: boolean;
    expiration_days: number;
  };
  notification_settings: {
    email_enabled: boolean;
    sms_enabled: boolean;
    push_enabled: boolean;
    email_templates: Record<string, any>;
  };
  smtp_settings: {
    host: string;
    port: number;
    username: string;
    password: string;
    use_tls: boolean;
    from_email: string;
    from_name: string;
  };
  created_at: string;
  updated_at: string;
}

interface UserSettings {
  id: number;
  user_id: number;
  theme: 'light' | 'dark' | 'auto';
  language: string;
  timezone: string;
  date_format: string;
  time_format: string;
  notifications: {
    email: boolean;
    sms: boolean;
    push: boolean;
    desktop: boolean;
  };
  privacy_settings: {
    profile_visible: boolean;
    show_email: boolean;
    show_phone: boolean;
    activity_tracking: boolean;
  };
  dashboard_settings: {
    layout: 'default' | 'compact' | 'detailed';
    widgets: string[];
    refresh_interval: number;
  };
  created_at: string;
  updated_at: string;
}

interface UpdateSystemSettingsRequest {
  site_name?: string;
  site_description?: string;
  logo_url?: string;
  favicon_url?: string;
  contact_email?: string;
  contact_phone?: string;
  contact_address?: string;
  timezone?: string;
  language?: string;
  date_format?: string;
  time_format?: string;
  currency?: string;
  theme?: 'light' | 'dark' | 'auto';
  registration_enabled?: boolean;
  email_verification_required?: boolean;
  max_file_size?: number;
  allowed_file_types?: string[];
  session_timeout?: number;
  password_policy?: Partial<SystemSettings['password_policy']>;
  notification_settings?: Partial<SystemSettings['notification_settings']>;
  smtp_settings?: Partial<SystemSettings['smtp_settings']>;
}

interface UpdateUserSettingsRequest {
  theme?: 'light' | 'dark' | 'auto';
  language?: string;
  timezone?: string;
  date_format?: string;
  time_format?: string;
  notifications?: Partial<UserSettings['notifications']>;
  privacy_settings?: Partial<UserSettings['privacy_settings']>;
  dashboard_settings?: Partial<UserSettings['dashboard_settings']>;
}

class SettingsService {
  // 获取系统设置
  async getSystemSettings(): Promise<SystemSettings> {
    try {
      return await apiClient.get<SystemSettings>("/settings/system", {
        useCache: true,
        cacheTTL: 10 * 60 * 1000, // 10分钟缓存
      });
    } catch (error: any) {
      console.error("获取系统设置失败:", error);
      throw new AppError(
        error.message || "获取系统设置失败",
        error.code || "GET_SYSTEM_SETTINGS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 更新系统设置
  async updateSystemSettings(data: UpdateSystemSettingsRequest): Promise<SystemSettings> {
    try {
      return await apiClient.put<SystemSettings>("/settings/system", data, {
      });
    } catch (error: any) {
      console.error("更新系统设置失败:", error);
      throw new AppError(
        error.message || "更新系统设置失败",
        error.code || "UPDATE_SYSTEM_SETTINGS_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 获取用户设置
  async getUserSettings(): Promise<UserSettings> {
    try {
      return await apiClient.get<UserSettings>("/settings/user", {
        useCache: true,
        cacheTTL: 5 * 60 * 1000, // 5分钟缓存
      });
    } catch (error: any) {
      console.error("获取用户设置失败:", error);
      throw new AppError(
        error.message || "获取用户设置失败",
        error.code || "GET_USER_SETTINGS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 更新用户设置
  async updateUserSettings(data: UpdateUserSettingsRequest): Promise<UserSettings> {
    try {
      return await apiClient.put<UserSettings>("/settings/user", data, {
      });
    } catch (error: any) {
      console.error("更新用户设置失败:", error);
      throw new AppError(
        error.message || "更新用户设置失败",
        error.code || "UPDATE_USER_SETTINGS_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 重置用户设置
  async resetUserSettings(): Promise<UserSettings> {
    try {
      return await apiClient.post<UserSettings>("/settings/user/reset", {}, {
      });
    } catch (error: any) {
      console.error("重置用户设置失败:", error);
      throw new AppError(
        error.message || "重置用户设置失败",
        error.code || "RESET_USER_SETTINGS_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 测试SMTP设置
  async testSmtpSettings(settings: {
    host: string;
    port: number;
    username: string;
    password: string;
    use_tls: boolean;
    from_email: string;
    from_name: string;
  }): Promise<{ success: boolean; message: string }> {
    try {
      return await apiClient.post<{ success: boolean; message: string }>(
        "/settings/test-smtp",
        settings,
        {
          timeout: 30000, // 30秒超时
        },
      );
    } catch (error: any) {
      console.error("测试SMTP设置失败:", error);
      throw new AppError(
        error.message || "测试SMTP设置失败",
        error.code || "TEST_SMTP_SETTINGS_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 发送测试邮件
  async sendTestEmail(email: string): Promise<{ success: boolean; message: string }> {
    try {
      return await apiClient.post<{ success: boolean; message: string }>(
        "/settings/test-email",
        { email },
        {
          timeout: 30000, // 30秒超时
        },
      );
    } catch (error: any) {
      console.error("发送测试邮件失败:", error);
      throw new AppError(
        error.message || "发送测试邮件失败",
        error.code || "SEND_TEST_EMAIL_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 上传Logo
  async uploadLogo(file: File): Promise<{ logo_url: string }> {
    try {
      const formData = new FormData();
      formData.append("logo", file);

      const response = await apiClient.getClient().post("/settings/upload-logo", formData, {
        headers: {
          "Content-Type": "multipart/form-data",
        },
        timeout: 30000, // 30秒超时
      });

      return response.data;
    } catch (error: any) {
      console.error("上传Logo失败:", error);
      throw new AppError(
        error.message || "上传Logo失败",
        error.code || "UPLOAD_LOGO_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 上传Favicon
  async uploadFavicon(file: File): Promise<{ favicon_url: string }> {
    try {
      const formData = new FormData();
      formData.append("favicon", file);

      const response = await apiClient.getClient().post("/settings/upload-favicon", formData, {
        headers: {
          "Content-Type": "multipart/form-data",
        },
        timeout: 30000, // 30秒超时
      });

      return response.data;
    } catch (error: any) {
      console.error("上传Favicon失败:", error);
      throw new AppError(
        error.message || "上传Favicon失败",
        error.code || "UPLOAD_FAVICON_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 获取可用语言列表
  async getAvailableLanguages(): Promise<Array<{
    code: string;
    name: string;
    native_name: string;
    flag?: string;
  }>> {
    try {
      return await apiClient.get<Array<{
        code: string;
        name: string;
        native_name: string;
        flag?: string;
      }>>("/settings/languages", {
        useCache: true,
        cacheTTL: 60 * 60 * 1000, // 1小时缓存
      });
    } catch (error: any) {
      console.error("获取可用语言列表失败:", error);
      throw new AppError(
        error.message || "获取可用语言列表失败",
        error.code || "GET_AVAILABLE_LANGUAGES_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取可用时区列表
  async getAvailableTimezones(): Promise<Array<{
    id: string;
    name: string;
    offset: string;
    country: string;
  }>> {
    try {
      return await apiClient.get<Array<{
        id: string;
        name: string;
        offset: string;
        country: string;
      }>>("/settings/timezones", {
        useCache: true,
        cacheTTL: 60 * 60 * 1000, // 1小时缓存
      });
    } catch (error: any) {
      console.error("获取可用时区列表失败:", error);
      throw new AppError(
        error.message || "获取可用时区列表失败",
        error.code || "GET_AVAILABLE_TIMEZONES_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取可用货币列表
  async getAvailableCurrencies(): Promise<Array<{
    code: string;
    name: string;
    symbol: string;
    country: string;
  }>> {
    try {
      return await apiClient.get<Array<{
        code: string;
        name: string;
        symbol: string;
        country: string;
      }>>("/settings/currencies", {
        useCache: true,
        cacheTTL: 60 * 60 * 1000, // 1小时缓存
      });
    } catch (error: any) {
      console.error("获取可用货币列表失败:", error);
      throw new AppError(
        error.message || "获取可用货币列表失败",
        error.code || "GET_AVAILABLE_CURRENCIES_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 备份系统设置
  async backupSettings(): Promise<{ backup_data: string; timestamp: string }> {
    try {
      return await apiClient.post<{ backup_data: string; timestamp: string }>(
        "/settings/backup",
        {},
      );
    } catch (error: any) {
      console.error("备份系统设置失败:", error);
      throw new AppError(
        error.message || "备份系统设置失败",
        error.code || "BACKUP_SETTINGS_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 恢复系统设置
  async restoreSettings(backupData: string): Promise<SystemSettings> {
    try {
      return await apiClient.post<SystemSettings>(
        "/settings/restore",
        { backup_data: backupData },
      );
    } catch (error: any) {
      console.error("恢复系统设置失败:", error);
      throw new AppError(
        error.message || "恢复系统设置失败",
        error.code || "RESTORE_SETTINGS_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 获取系统健康状态
  async getSystemHealth(): Promise<{
    status: 'healthy' | 'warning' | 'error';
    checks: Array<{
      name: string;
      status: 'pass' | 'fail' | 'warning';
      message: string;
      timestamp: string;
    }>;
    last_check: string;
  }> {
    try {
      return await apiClient.get<{
        status: 'healthy' | 'warning' | 'error';
        checks: Array<{
          name: string;
          status: 'pass' | 'fail' | 'warning';
          message: string;
          timestamp: string;
        }>;
        last_check: string;
      }>("/settings/health", {
      });
    } catch (error: any) {
      console.error("获取系统健康状态失败:", error);
      throw new AppError(
        error.message || "获取系统健康状态失败",
        error.code || "GET_SYSTEM_HEALTH_ERROR",
        error.statusCode || 500,
      );
    }
  }
}

// 导出单例实例
export const settingsService = new SettingsService();

// 为了向后兼容，也导出独立的函数
export const getSystemSettings = () => settingsService.getSystemSettings();
export const updateSystemSettings = (data: UpdateSystemSettingsRequest) => settingsService.updateSystemSettings(data);
export const getUserSettings = () => settingsService.getUserSettings();
export const updateUserSettings = (data: UpdateUserSettingsRequest) => settingsService.updateUserSettings(data);
export const resetUserSettings = () => settingsService.resetUserSettings();
export const testSmtpSettings = (settings: {
  host: string;
  port: number;
  username: string;
  password: string;
  use_tls: boolean;
  from_email: string;
  from_name: string;
}) => settingsService.testSmtpSettings(settings);
export const sendTestEmail = (email: string) => settingsService.sendTestEmail(email);
export const uploadLogo = (file: File) => settingsService.uploadLogo(file);
export const uploadFavicon = (file: File) => settingsService.uploadFavicon(file);
export const getAvailableLanguages = () => settingsService.getAvailableLanguages();
export const getAvailableTimezones = () => settingsService.getAvailableTimezones();
export const getAvailableCurrencies = () => settingsService.getAvailableCurrencies();
export const backupSettings = () => settingsService.backupSettings();
export const restoreSettings = (backupData: string) => settingsService.restoreSettings(backupData);
export const getSystemHealth = () => settingsService.getSystemHealth();

export default settingsService;
