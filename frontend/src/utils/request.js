import axios from 'axios';
import { message } from 'antd';
import { getToken } from './storage';

const service = axios.create({
  baseURL: '/api/v1',
  timeout: 10000
});

// 请求拦截器
service.interceptors.request.use(
  config => {
    const token = getToken();
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`;
    }
    return config;
  },
  error => {
    return Promise.reject(error);
  }
);

// 响应拦截器
service.interceptors.response.use(
  response => {
    let res = response.data;
    
    // 检查是否为base64编码的响应（后端压缩）
    if (res && res.data && typeof res.data === 'string') {
      try {
        // 尝试解析base64编码的响应
        const decodedData = atob(res.data);
        const parsedData = JSON.parse(decodedData);
        res = parsedData;
      } catch (e) {
        // 如果解析失败，使用原始响应
        console.warn('Failed to decode base64 response:', e);
      }
    }
    
    // 处理新的后端响应格式 {success: true, data: {...}, meta: {...}}
    if (res.success === true) {
      return res.data;
    } else if (res.success === false) {
      // 处理错误响应
      const errorMessage = res.error?.message || res.error?.details || '请求失败';
      message.error(errorMessage);
      return Promise.reject(new Error(errorMessage));
    } else if (res.code !== 0) {
      // 兼容旧的响应格式
      message.error(res.msg || '请求失败');
      return Promise.reject(new Error(res.msg || 'Error'));
    } else {
      // 兼容旧的响应格式
      return res.data;
    }
  },
  error => {
    if (error.response) {
      switch (error.response.status) {
        case 401:
          message.error('未授权，请重新登录');
          // 清除token并跳转到登录页
          localStorage.removeItem('auth_token');
          window.location.href = '/login';
          break;
        case 403:
          message.error('拒绝访问');
          break;
        case 404:
          message.error('请求资源不存在');
          break;
        case 429:
          message.error('请求过于频繁，请稍后再试');
          break;
        case 500:
          message.error('服务器错误');
          break;
        default:
          message.error(error.message || '网络错误');
      }
    } else if (error.code === 'ECONNABORTED') {
      message.error('请求超时，请检查网络连接');
    } else {
      message.error('网络错误，请检查网络连接');
    }
    return Promise.reject(error);
  }
);

export default service;