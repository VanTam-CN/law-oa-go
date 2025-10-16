import React, { useEffect } from 'react';

const DevToolsReminder: React.FC = () => {
  useEffect(() => {
    // 只在开发环境中显示
    if (process.env.NODE_ENV === 'development') {
      // 创建一个样式化的提示，而不是使用console.warn
      const message = '%c💡 提示: 安装 React DevTools 以获得更好的开发体验\nhttps://reactjs.org/link/react-devtools';
      const styles = [
        'color: #0284c7',
        'font-weight: bold',
        'font-size: 12px',
        'padding: 4px 8px',
        'border-radius: 4px',
        'background: #e0f2fe'
      ].join(';');

      // 延迟显示，避免页面加载时的干扰
      const timer = setTimeout(() => {
        console.log(message, styles);
      }, 2000);

      return () => clearTimeout(timer);
    }
  }, []);

  return null; // 不渲染任何内容
};

export default DevToolsReminder;