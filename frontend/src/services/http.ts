import axios, { AxiosRequestConfig, AxiosResponse, AxiosError } from 'axios';
import { message } from 'antd';
import { getToken } from '@/utils/storage';
import { getAuthToken } from '@/utils/auth';

// 创建axios实例
const service = axios.create({
  baseURL: 'http://localhost:8080/api',
  timeout: 10000,
  paramsSerializer: {
    indexes: null,
    serialize: (params) => {
      const parts: string[] = [];
      Object.entries(params).forEach(([key, value]) => {
        if (value === undefined || value === null) return;
        if (Array.isArray(value)) {
          value.forEach(item => {
            parts.push(`${encodeURIComponent(key)}=${encodeURIComponent(item)}`);
          });
        } else {
          parts.push(`${encodeURIComponent(key)}=${encodeURIComponent(value)}`);
        }
      });
      return parts.join('&');
    }
  }
});

// 请求拦截器
service.interceptors.request.use(
  (config: AxiosRequestConfig): AxiosRequestConfig => {
    const token = getToken() || getAuthToken();
    if (token && config.headers) {
      config.headers['Authorization'] = `Bearer ${token}`;
    }
    return config;
  },
  (error: AxiosError) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
service.interceptors.response.use(
  (response: AxiosResponse) => {
    const res = response.data;
    
    // 如果是文件下载，直接返回
    if (response.config.responseType === 'blob') {
      return response;
    }
    
    // 根据后端API的响应格式进行处理 (code: 0 或 200 表示成功)
    if (res.code !== undefined && res.code !== 0 && res.code !== 200) {
      // 对特定API路径不显示错误消息
      const noErrorPaths = ['/admin/current-user/roles', '/admin/current-user/permissions', '/notifications', '/notifications/stats', '/file/stats', '/api/file/stats'];
      const shouldShowError = !noErrorPaths.some(path => response.config.url?.includes(path));
      
      if (shouldShowError) {
        message.error(res.msg || res.message || '请求失败');
      }
      
      // 401: 未登录或token过期
      if (res.code === 401) {
        // 可以在这里处理登出逻辑
      }
      
      return Promise.reject(new Error(res.msg || res.message || 'Error'));
    } else {
      // 如果有data字段，返回data，否则返回整个响应
      return res.data !== undefined ? res.data : res;
    }
  },
  (error: AxiosError) => {
    if (error.response) {
      // 对特定API路径和状态码不显示错误消息
      const noErrorPaths = ['/admin/current-user/roles', '/admin/current-user/permissions', '/notifications', '/notifications/stats', '/file/stats', '/api/file/stats'];
      const shouldShowError = !noErrorPaths.some(path => error.config?.url?.includes(path));
      
      if (shouldShowError) {
        switch (error.response.status) {
          case 401:
            // 在开发模式下提供更友好的错误提示
            const isDevMode = process.env.NODE_ENV === 'development';
            if (isDevMode) {
              console.warn('🛠️ 开发者模式：认证错误，但不自动重定向');
              message.error('开发模式认证错误，请检查token设置或重新登录');
            } else {
              message.error('未授权，请重新登录');
            }
            break;
          case 403:
            // 只对非通知API和文件统计API显示403错误
            if (!error.config?.url?.includes('/notifications') && !error.config?.url?.includes('/file/stats') && !error.config?.url?.includes('/api/file/stats')) {
              message.error('拒绝访问');
            }
            break;
          case 404:
            // 只对非特定API显示404错误
            if (!error.config?.url?.includes('/admin/current-user')) {
              message.error('请求资源不存在');
            }
            break;
          case 500:
            message.error('服务器错误');
            break;
          default:
            message.error(`请求错误: ${error.message}`);
        }
      }
    } else {
      message.error(`网络错误: ${error.message}`);
    }
    return Promise.reject(error);
  }
);

// 封装GET请求
export const get = <T>(url: string, params?: any, config?: AxiosRequestConfig): Promise<T> => {
  return service.get(url, { params, ...config });
};

// 封装POST请求
export const post = <T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> => {
  return service.post(url, data, config);
};

// 封装PUT请求
export const put = <T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> => {
  return service.put(url, data, config);
};

// 封装DELETE请求
export const del = <T>(url: string, config?: AxiosRequestConfig): Promise<T> => {
  return service.delete(url, config);
};

export default service;