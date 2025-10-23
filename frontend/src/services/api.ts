import { message } from '@/utils/messageHelper';
import { getToken } from './auth';

// API配置
const API_BASE_URL = 'http://localhost:8080/api/v1'; // 直接指向后端服务
const DEFAULT_TIMEOUT = 30000;
const MAX_RETRIES = 3;

// 缓存配置
const CACHE_PREFIX = 'law_oa_cache_';
const CACHE_EXPIRY = 5 * 60 * 1000; // 5分钟

// 请求队列
let requestQueue: Map<string, Promise<any>> = new Map();

// 请求缓存
const requestCache = new Map<string, {
  data: any;
  timestamp: number;
}>();

// 性能监控
const performanceMetrics = {
  requestCount: 0,
  responseTime: 0,
  errorCount: 0,
  cacheHits: 0,
};

// 接口类型定义
interface RequestOptions extends RequestInit {
  timeout?: number;
  retry?: number;
  cache?: boolean;
  cacheKey?: string;
  showLoading?: boolean;
  showErrorMessage?: boolean;
}

interface ApiResponse<T = any> {
  code: number;
  data: T;
  message: string;
  success: boolean;
}

/**
 * 检查缓存
 */
const getCache = (key: string): any => {
  const cached = requestCache.get(key);
  if (cached && Date.now() - cached.timestamp < CACHE_EXPIRY) {
    performanceMetrics.cacheHits++;
    return cached.data;
  }
  return null;
};

/**
 * 设置缓存
 */
const setCache = (key: string, data: any): void => {
  requestCache.set(key, {
    data,
    timestamp: Date.now(),
  });
};

/**
 * 清理缓存
 */
const clearCache = (pattern?: string): void => {
  if (pattern) {
    for (const key of requestCache.keys()) {
      if (key.includes(pattern)) {
        requestCache.delete(key);
      }
    }
  } else {
    requestCache.clear();
  }
};

/**
 * 生成请求键
 */
const generateRequestKey = (url: string, options?: RequestOptions): string => {
  const method = options?.method || 'GET';
  const body = options?.body ? JSON.stringify(options.body) : '';
  return `${method}:${url}:${body}`;
};

/**
 * 请求重试
 */
const retryRequest = async (
  url: string,
  options: RequestOptions,
  retryCount: number
): Promise<Response> => {
  try {
    return await fetch(url, options);
  } catch (error) {
    if (retryCount > 0) {
      await new Promise(resolve => setTimeout(resolve, 1000 * (MAX_RETRIES - retryCount + 1)));
      return retryRequest(url, options, retryCount - 1);
    }
    throw error;
  }
};

/**
 * 超时控制
 */
const timeoutController = (timeout: number): AbortController => {
  const controller = new AbortController();
  setTimeout(() => controller.abort(), timeout);
  return controller;
};

/**
 * 统一请求处理
 */
const request = async <T = any>(
  url: string,
  options: RequestOptions = {}
): Promise<ApiResponse<T>> => {
  const {
    timeout = DEFAULT_TIMEOUT,
    retry = MAX_RETRIES,
    cache = false,
    cacheKey,
    showLoading = true,
    showErrorMessage = true,
    ...fetchOptions
  } = options;

  const startTime = performance.now();
  performanceMetrics.requestCount++;

  // 生成完整的URL
  const fullUrl = `${API_BASE_URL}${url}`;
  const requestKey = generateRequestKey(fullUrl, fetchOptions);

  // 检查缓存
  if (cache && (fetchOptions.method === 'GET' || !fetchOptions.method)) {
    const cachedData = getCache(cacheKey || requestKey);
    if (cachedData) {
      return cachedData;
    }
  }

  // 检查请求队列
  if (requestQueue.has(requestKey)) {
    return requestQueue.get(requestKey);
  }

  // 获取认证token
  const token = getToken();
  const headers = {
    'Content-Type': 'application/json',
    ...fetchOptions.headers,
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };

  const finalOptions: RequestOptions = {
    ...fetchOptions,
    headers,
    signal: timeoutController(timeout).signal,
  };

  // 创建请求Promise
  const requestPromise = (async () => {
    try {
      const response = await retryRequest(fullUrl, finalOptions, retry);
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data: ApiResponse<T> = await response.json();
      
      // 记录响应时间
      const responseTime = performance.now() - startTime;
      performanceMetrics.responseTime += responseTime;

      // 缓存成功响应
      if (data.success && cache && (fetchOptions.method === 'GET' || !fetchOptions.method)) {
        setCache(cacheKey || requestKey, data);
      }

      return data;
    } catch (error) {
      performanceMetrics.errorCount++;
      
      if (showErrorMessage) {
        message.error(error instanceof Error ? error.message : '请求失败');
      }
      
      throw error;
    } finally {
      requestQueue.delete(requestKey);
    }
  })();

  requestQueue.set(requestKey, requestPromise);
  return requestPromise;
};

/**
 * GET请求
 */
export const get = <T = any>(
  url: string,
  params?: Record<string, any>,
  options: RequestOptions = {}
): Promise<ApiResponse<T>> => {
  const queryString = params ? `?${new URLSearchParams(params).toString()}` : '';
  return request<T>(`${url}${queryString}`, {
    method: 'GET',
    cache: true,
    ...options,
  });
};

/**
 * POST请求
 */
export const post = <T = any>(
  url: string,
  data?: any,
  options: RequestOptions = {}
): Promise<ApiResponse<T>> => {
  return request<T>(url, {
    method: 'POST',
    body: JSON.stringify(data),
    cache: false,
    ...options,
  });
};

/**
 * PUT请求
 */
export const put = <T = any>(
  url: string,
  data?: any,
  options: RequestOptions = {}
): Promise<ApiResponse<T>> => {
  return request<T>(url, {
    method: 'PUT',
    body: JSON.stringify(data),
    cache: false,
    ...options,
  });
};

/**
 * DELETE请求
 */
export const del = <T = any>(
  url: string,
  options: RequestOptions = {}
): Promise<ApiResponse<T>> => {
  return request<T>(url, {
    method: 'DELETE',
    cache: false,
    ...options,
  });
};

/**
 * 文件上传
 */
export const upload = (
  url: string,
  file: File,
  options: RequestOptions = {}
): Promise<ApiResponse<{ url: string }>> => {
  const formData = new FormData();
  formData.append('file', file);

  return request<{ url: string }>(url, {
    method: 'POST',
    body: formData,
    headers: {}, // 让浏览器自动设置Content-Type
    cache: false,
    ...options,
  });
};

/**
 * 批量请求
 */
export const batchRequest = async <T = any>(
  requests: Array<() => Promise<ApiResponse<T>>>
): Promise<ApiResponse<T>[]> => {
  return Promise.all(requests.map(req => req()));
};

/**
 * 获取性能指标
 */
export const getPerformanceMetrics = () => ({
  ...performanceMetrics,
  averageResponseTime: performanceMetrics.responseCount > 0 
    ? performanceMetrics.responseTime / performanceMetrics.requestCount 
    : 0,
  errorRate: performanceMetrics.requestCount > 0 
    ? (performanceMetrics.errorCount / performanceMetrics.requestCount) * 100 
    : 0,
  cacheHitRate: performanceMetrics.requestCount > 0 
    ? (performanceMetrics.cacheHits / performanceMetrics.requestCount) * 100 
    : 0,
});

/**
 * 重置性能指标
 */
export const resetPerformanceMetrics = () => {
  performanceMetrics.requestCount = 0;
  performanceMetrics.responseTime = 0;
  performanceMetrics.errorCount = 0;
  performanceMetrics.cacheHits = 0;
};

/**
 * 清理请求缓存
 */
export const clearRequestCache = (pattern?: string) => {
  clearCache(pattern);
};

// 导出缓存清理函数
export { clearCache };

// 导出默认API服务
export default {
  get,
  post,
  put,
  del,
  upload,
  batchRequest,
  getPerformanceMetrics,
  resetPerformanceMetrics,
  clearRequestCache,
};