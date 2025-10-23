/**
 * 安全的消息工具 - 使用App组件的message上下文
 */

import { App } from 'antd';

let appMessage: any = null;

// 设置App组件的message实例
export const setAppMessage = (app: any) => {
  appMessage = app;
};

// 导出安全的message实例
export const message = {
  success: (content: string, duration?: number) => {
    if (appMessage) {
      try {
        return appMessage.success(content, duration);
      } catch (error) {
        console.error('App message success failed:', error);
        console.log('SUCCESS:', content);
      }
    } else {
      // 作为后备方案，使用console
      console.log('SUCCESS:', content);
    }
  },

  error: (content: string, duration?: number) => {
    if (appMessage) {
      try {
        return appMessage.error(content, duration);
      } catch (error) {
        console.error('App message error failed:', error);
        console.error('ERROR:', content);
      }
    } else {
      // 作为后备方案，使用console
      console.error('ERROR:', content);
    }
  },

  info: (content: string, duration?: number) => {
    if (appMessage) {
      try {
        return appMessage.info(content, duration);
      } catch (error) {
        console.error('App message info failed:', error);
        console.log('INFO:', content);
      }
    } else {
      // 作为后备方案，使用console
      console.log('INFO:', content);
    }
  },

  warning: (content: string, duration?: number) => {
    if (appMessage) {
      try {
        return appMessage.warning(content, duration);
      } catch (error) {
        console.error('App message warning failed:', error);
        console.warn('WARNING:', content);
      }
    } else {
      // 作为后备方案，使用console
      console.warn('WARNING:', content);
    }
  },

  loading: (content: string, duration?: number) => {
    if (appMessage) {
      try {
        return appMessage.loading(content, duration);
      } catch (error) {
        console.error('App message loading failed:', error);
        console.log('LOADING:', content);
        // 作为后备方案，返回一个空的hide函数
        return () => {};
      }
    } else {
      // 作为后备方案，返回一个空的hide函数
      return () => {};
    }
  }
};

// 提供默认导出
export default message;