// API配置文件 - 生产环境

export interface ApiConfig {
  baseURL: string;
  timeout: number;
  retryTimes: number;
  retryDelay: number;
}

// 环境配置
const ENV = (import.meta as any).env?.MODE || 'development';

// 生产环境API配置
const PRODUCTION_CONFIG: ApiConfig = {
  baseURL: '/api', 
  timeout: 30000, // 30秒超时
  retryTimes: 2,
  retryDelay: 1000
};

// 开发环境API配置
const DEVELOPMENT_CONFIG: ApiConfig = {
  baseURL: '/api', // 开发环境使用代理
  timeout: 10000,
  retryTimes: 1,
  retryDelay: 500
};

// 根据环境选择配置
export const API_CONFIG: ApiConfig = ENV === 'production' 
  ? PRODUCTION_CONFIG 
  : DEVELOPMENT_CONFIG;

// API端点配置
export const API_ENDPOINTS = {
  // 冲突检索相关
  CONFLICT_CHECK: {
    PERFORM: '/conflict-check/perform-with-progress',
    PRE_CHECK: '/conflict-check/pre-check',
    HISTORY: '/conflict-check/history'
  },
  // 案件管理相关
  CASE_MANAGEMENT: {
    CREATE: '/case/create',
    LIST: '/case/list',
    DETAIL: '/case/detail'
  }
};

// HTTP状态码映射
export const HTTP_STATUS_MESSAGES = {
  400: '请求参数错误',
  401: '身份验证失败，请重新登录',
  403: '权限不足，无法访问该资源',
  404: '请求的资源不存在',
  408: '请求超时，请检查网络连接',
  429: '请求过于频繁，请稍后再试',
  500: '服务器内部错误，请联系技术支持',
  502: '网关错误，服务暂时不可用',
  503: '服务暂时不可用，请稍后重试',
  504: '网关超时，请稍后重试'
};

// API调用工具函数
export class ApiClient {
  private static async makeRequest(
    endpoint: string, 
    options: RequestInit = {},
    retryCount = 0
  ): Promise<Response> {
    const url = `${API_CONFIG.baseURL}${endpoint}`;
    const controller = new AbortController();
    
    // 设置超时
    const timeoutId = setTimeout(() => controller.abort(), API_CONFIG.timeout);
    
    try {
      const response = await fetch(url, {
        ...options,
        signal: controller.signal,
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
          ...options.headers
        }
      });
      
      clearTimeout(timeoutId);
      
      // 如果请求失败且还有重试次数
      if (!response.ok && retryCount < API_CONFIG.retryTimes) {
        await new Promise(resolve => setTimeout(resolve, API_CONFIG.retryDelay));
        return ApiClient.makeRequest(endpoint, options, retryCount + 1);
      }
      
      return response;
    } catch (error) {
      clearTimeout(timeoutId);
      
      if (error instanceof DOMException && error.name === 'AbortError') {
        throw new Error(`请求超时（${API_CONFIG.timeout / 1000}秒），请检查网络连接`);
      }
      
      // 网络错误重试
      if (retryCount < API_CONFIG.retryTimes) {
        await new Promise(resolve => setTimeout(resolve, API_CONFIG.retryDelay));
        return ApiClient.makeRequest(endpoint, options, retryCount + 1);
      }
      
      throw error;
    }
  }
  
  static async post(endpoint: string, data: any): Promise<any> {
    try {
      const response = await this.makeRequest(endpoint, {
        method: 'POST',
        body: JSON.stringify(data)
      });
      
      if (!response.ok) {
        const errorMessage = HTTP_STATUS_MESSAGES[response.status as keyof typeof HTTP_STATUS_MESSAGES] 
          || `HTTP错误: ${response.status}`;
        throw new Error(errorMessage);
      }
      
      // 尝试解析JSON响应
      try {
        return await response.json();
      } catch (jsonError) {
        console.error('JSON解析错误:', jsonError);
        throw new Error('服务器返回数据格式错误');
      }
    } catch (error) {
      // 提供更详细的错误信息
      if (error instanceof TypeError && error.message.includes('Failed to fetch')) {
        throw new Error('网络连接失败，请检查后端服务是否启动');
      }
      throw error;
    }
  }
  
  static async get(endpoint: string): Promise<any> {
    try {
      const response = await this.makeRequest(endpoint, {
        method: 'GET'
      });
      
      if (!response.ok) {
        const errorMessage = HTTP_STATUS_MESSAGES[response.status as keyof typeof HTTP_STATUS_MESSAGES] 
          || `HTTP错误: ${response.status}`;
        throw new Error(errorMessage);
      }
      
      // 尝试解析JSON响应
      try {
        return await response.json();
      } catch (jsonError) {
        console.error('JSON解析错误:', jsonError);
        throw new Error('服务器返回数据格式错误');
      }
    } catch (error) {
      // 提供更详细的错误信息
      if (error instanceof TypeError && error.message.includes('Failed to fetch')) {
        throw new Error('网络连接失败，请检查后端服务是否启动');
      }
      throw error;
    }
  }
}