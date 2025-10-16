import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from "axios";
import { ApiResponse, ApiError } from "../types";

interface CacheEntry<T> {
  data: T;
  timestamp: number;
  ttl: number;
}

interface PendingRequest {
  promise: Promise<any>;
  timestamp: number;
}

class APIClient {
  private client: AxiosInstance;
  private cache: Map<string, CacheEntry<any>> = new Map();
  private pendingRequests: Map<string, PendingRequest> = new Map();
  private defaultCacheTTL: number = 5 * 60 * 1000; // 5 minutes
  private maxRetryAttempts: number = 3;
  private retryDelay: number = 1000; // 1 second base delay

  constructor() {
    this.client = axios.create({
      baseURL:
        process.env.REACT_APP_API_BASE_URL || "http://localhost:8080/api",
      timeout: 30000,
      headers: {
        "Content-Type": "application/json",
      },
    });

    this.setupInterceptors();
  }

  // 生成缓存键
  private generateCacheKey(url: string, params?: any, data?: any): string {
    const keyData = { url, params, data };
    return btoa(JSON.stringify(keyData));
  }

  // 检查缓存是否有效
  private isCacheValid(entry: CacheEntry<any>): boolean {
    return Date.now() - entry.timestamp < entry.ttl;
  }

  // 清理过期缓存
  private cleanExpiredCache(): void {
    const now = Date.now();
    for (const [key, entry] of this.cache.entries()) {
      if (now - entry.timestamp > entry.ttl) {
        this.cache.delete(key);
      }
    }
  }

  // 指数退避重试
  private async retryRequest<T>(
    requestFn: () => Promise<T>,
    attempt: number = 0,
  ): Promise<T> {
    try {
      return await requestFn();
    } catch (error: any) {
      // 如果是认证错误或客户端错误，不重试
      if (
        error.code === "AUTHENTICATION_ERROR" ||
        error.code === "AUTHORIZATION_ERROR" ||
        error.code === "VALIDATION_ERROR" ||
        error.response?.status < 500
      ) {
        throw error;
      }

      if (attempt >= this.maxRetryAttempts) {
        throw error;
      }

      const delay = this.retryDelay * Math.pow(2, attempt);
      await new Promise((resolve) => setTimeout(resolve, delay));

      return this.retryRequest(requestFn, attempt + 1);
    }
  }

  // 获取缓存数据
  private getCacheData<T>(key: string): T | null {
    const entry = this.cache.get(key);
    if (entry && this.isCacheValid(entry)) {
      return entry.data;
    }
    return null;
  }

  // 设置缓存数据
  private setCacheData<T>(key: string, data: T, ttl?: number): void {
    this.cache.set(key, {
      data,
      timestamp: Date.now(),
      ttl: ttl || this.defaultCacheTTL,
    });
  }

  // 检查是否有重复请求
  private getPendingRequest(key: string): PendingRequest | null {
    const pending = this.pendingRequests.get(key);
    if (pending && Date.now() - pending.timestamp < 30000) {
      // 30秒内认为有效
      return pending;
    }
    return null;
  }

  // 设置等待中的请求
  private setPendingRequest(key: string, promise: Promise<any>): void {
    this.pendingRequests.set(key, {
      promise,
      timestamp: Date.now(),
    });
  }

  // 移除等待中的请求
  private removePendingRequest(key: string): void {
    this.pendingRequests.delete(key);
  }

  // 清理过期的等待请求
  private cleanExpiredPendingRequests(): void {
    const now = Date.now();
    for (const [key, pending] of this.pendingRequests.entries()) {
      if (now - pending.timestamp > 30000) {
        this.pendingRequests.delete(key);
      }
    }
  }

  private setupInterceptors(): void {
    // 请求拦截器
    this.client.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem("token");
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }

        // 添加请求ID用于追踪
        config.headers["X-Request-ID"] = this.generateRequestId();

        return config;
      },
      (error) => {
        console.error("请求拦截器错误:", error);
        return Promise.reject(error);
      },
    );

    // 响应拦截器
    this.client.interceptors.response.use(
      (response: AxiosResponse<ApiResponse>) => {
        // 统一处理后端API响应格式
        const apiResponse = response.data;

        if (apiResponse.success === false) {
          // 业务错误，抛出给调用者处理
          const error: ApiError = {
            code: apiResponse.error?.code || "UNKNOWN_ERROR",
            message: apiResponse.error?.message || "请求失败",
            details: apiResponse.error?.details,
            suggestions: apiResponse.error?.suggestions,
          };
          return Promise.reject(error);
        }

        // 成功响应，返回data字段
        return apiResponse.data;
      },
      (error) => {
        if (error.response) {
          // HTTP错误响应
          const status = error.response.status;
          const apiResponse = error.response.data as ApiResponse;

          if (status === 401) {
            // 认证失败，清除令牌并重定向
            this.handleAuthError();
          }

          // 构造标准错误对象
          const apiError: ApiError = {
            code: apiResponse?.error?.code || this.getErrorCodeByStatus(status),
            message:
              apiResponse?.error?.message ||
              this.getErrorMessageByStatus(status),
            details: apiResponse?.error?.details,
            suggestions:
              apiResponse?.error?.suggestions ||
              this.getErrorSuggestions(status),
          };

          return Promise.reject(apiError);
        } else if (error.request) {
          // 请求发出但没有收到响应
          return Promise.reject({
            code: "NETWORK_ERROR",
            message: "网络连接失败，请检查网络设置",
            suggestions: ["检查网络连接", "确认服务器正常运行"],
          });
        } else {
          // 请求配置错误
          return Promise.reject({
            code: "REQUEST_ERROR",
            message: "请求配置错误",
            details: error.message,
          });
        }
      },
    );
  }

  private generateRequestId(): string {
    return `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  private handleAuthError(): void {
    localStorage.removeItem("token");
    localStorage.removeItem("refreshToken");

    // 在开发模式下避免自动重定向到登录页面
    const isDevMode = process.env.NODE_ENV === 'development';
    if (isDevMode) {
      console.warn('🛠️ 开发者模式：认证错误但跳过自动重定向到登录页面');
      return;
    }

    // 避免在登录页面无限重定向
    if (window.location.pathname !== "/login") {
      console.log('认证失败，重定向到登录页面');
      window.location.href = "/login";
    }
  }

  private getErrorCodeByStatus(status: number): string {
    const errorCodes: Record<number, string> = {
      400: "VALIDATION_ERROR",
      401: "AUTHENTICATION_ERROR",
      403: "AUTHORIZATION_ERROR",
      404: "NOT_FOUND",
      409: "CONFLICT",
      429: "RATE_LIMIT_ERROR",
      500: "INTERNAL_ERROR",
    };
    return errorCodes[status] || "UNKNOWN_ERROR";
  }

  private getErrorMessageByStatus(status: number): string {
    const messages: Record<number, string> = {
      400: "请求参数验证失败",
      401: "认证失败",
      403: "权限不足",
      404: "请求的资源不存在",
      409: "资源冲突",
      429: "请求频率超限",
      500: "服务器内部错误",
    };
    return messages[status] || "未知错误";
  }

  private getErrorSuggestions(status: number): string[] {
    const suggestions: Record<number, string[]> = {
      400: ["检查请求参数格式", "确认必填字段已填写"],
      401: ["重新登录", "检查用户名和密码"],
      403: ["联系管理员获取权限", "确认账号状态正常"],
      404: ["检查资源ID是否正确", "确认资源存在"],
      429: ["稍后重试", "降低请求频率"],
      500: ["联系技术支持", "稍后重试"],
    };
    return suggestions[status] || ["稍后重试"];
  }

  // GET请求 - 带缓存和去重
  public async get<T>(
    url: string,
    config?: AxiosRequestConfig & { useCache?: boolean; cacheTTL?: number },
  ): Promise<T> {
    const useCache = config?.useCache ?? true;
    const cacheTTL = config?.cacheTTL;
    const cacheKey = this.generateCacheKey(url, config?.params);

    // 清理过期缓存和等待请求
    this.cleanExpiredCache();
    this.cleanExpiredPendingRequests();

    // 检查缓存
    if (useCache) {
      const cachedData = this.getCacheData<T>(cacheKey);
      if (cachedData) {
        return cachedData;
      }
    }

    // 检查重复请求
    const pendingRequest = this.getPendingRequest(cacheKey);
    if (pendingRequest) {
      return pendingRequest.promise;
    }

    // 创建请求
    const requestPromise = this.retryRequest(async () => {
      const response = await this.client.get<ApiResponse<T>>(url, config);
      const data = response as T;

      // 缓存结果
      if (useCache) {
        this.setCacheData(cacheKey, data, cacheTTL);
      }

      this.removePendingRequest(cacheKey);
      return data;
    });

    this.setPendingRequest(cacheKey, requestPromise);
    return requestPromise;
  }

  // POST请求 - 带重试
  public async post<T>(
    url: string,
    data?: any,
    config?: AxiosRequestConfig & { useRetry?: boolean },
  ): Promise<T> {
    const useRetry = config?.useRetry ?? true;
    const cacheKey = this.generateCacheKey(url, undefined, data);

    // 清理过期等待请求
    this.cleanExpiredPendingRequests();

    // 检查重复请求
    const pendingRequest = this.getPendingRequest(cacheKey);
    if (pendingRequest) {
      return pendingRequest.promise;
    }

    // 创建请求
    const requestPromise = this.retryRequest(async () => {
      const response = await this.client.post<ApiResponse<T>>(
        url,
        data,
        config,
      );
      const result = response as T;

      // POST请求成功后，清除相关的GET缓存
      this.clearRelatedCache(url);

      this.removePendingRequest(cacheKey);
      return result;
    });

    if (useRetry) {
      this.setPendingRequest(cacheKey, requestPromise);
    }
    return requestPromise;
  }

  // PUT请求
  public async put<T>(
    url: string,
    data?: any,
    config?: AxiosRequestConfig,
  ): Promise<T> {
    const response = await this.client.put<ApiResponse<T>>(url, data, config);
    return response as T;
  }

  // DELETE请求
  public async delete<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.delete<ApiResponse<T>>(url, config);
    return response as T;
  }

  // 分页GET请求
  public async getPaginated<T>(
    url: string,
    config?: AxiosRequestConfig,
  ): Promise<{
    data: T[];
    pagination: {
      page: number;
      page_size: number;
      total: number;
      total_pages: number;
    };
  }> {
    const response = await this.client.get<
      ApiResponse<{
        data: T[];
        pagination: {
          page: number;
          page_size: number;
          total: number;
          total_pages: number;
        };
      }>
    >(url, config);

    // 响应拦截器已经处理了ApiResponse格式，这里直接返回处理后的数据
    const responseData = response as any;

    return {
      data: responseData.data || [],
      pagination: responseData.pagination || {
        page: 1,
        page_size: 20,
        total: 0,
        total_pages: 0,
      },
    };
  }

  // 设置认证令牌
  public setAuthToken(token: string | null): void {
    if (token) {
      this.client.defaults.headers.common["Authorization"] = `Bearer ${token}`;
    } else {
      delete this.client.defaults.headers.common["Authorization"];
    }
  }

  // 获取客户端实例（用于高级操作）
  public getClient(): AxiosInstance {
    return this.client;
  }

  // 缓存管理方法
  public clearCache(url?: string): void {
    if (url) {
      // 清除特定URL的缓存
      const keysToDelete: string[] = [];
      for (const key of this.cache.keys()) {
        if (key.includes(btoa(url))) {
          keysToDelete.push(key);
        }
      }
      keysToDelete.forEach((key) => this.cache.delete(key));
    } else {
      // 清除所有缓存
      this.cache.clear();
    }
  }

  // 清除相关缓存（用于POST/PUT/DELETE后）
  private clearRelatedCache(url: string): void {
    // 清除相同URL的GET请求缓存
    this.clearCache(url);

    // 如果是资源列表，清除整个列表的缓存
    const resourcePattern = url.match(/\/api\/v1\/([^\/]+)/);
    if (resourcePattern) {
      const resource = resourcePattern[1];
      const resourceUrl = `/api/v1/${resource}`;
      this.clearCache(resourceUrl);
    }
  }

  // 获取缓存统计信息
  public getCacheStats() {
    return {
      totalEntries: this.cache.size,
      totalSize: Array.from(this.cache.values()).reduce(
        (sum, entry) => sum + JSON.stringify(entry.data).length,
        0,
      ),
      pendingRequests: this.pendingRequests.size,
    };
  }

  // 强制清除所有等待请求
  public clearAllPendingRequests(): void {
    this.pendingRequests.clear();
  }

  // 设置缓存配置
  public setCacheConfig(config: {
    defaultTTL?: number;
    maxRetryAttempts?: number;
    retryDelay?: number;
  }): void {
    if (config.defaultTTL) this.defaultCacheTTL = config.defaultTTL;
    if (config.maxRetryAttempts)
      this.maxRetryAttempts = config.maxRetryAttempts;
    if (config.retryDelay) this.retryDelay = config.retryDelay;
  }
}

const apiClient = new APIClient();
export default apiClient;
