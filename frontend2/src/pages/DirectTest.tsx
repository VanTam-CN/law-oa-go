import React from 'react';

const DirectTest: React.FC = () => {
  return (
    <div style={{ 
      padding: '50px', 
      backgroundColor: '#f0f2f5', 
      minHeight: '100vh',
      textAlign: 'center' 
    }}>
      <h1>直接测试页面</h1>
      <p>如果你能看到这个页面，说明React应用正在工作！</p>
      <p>当前时间: {new Date().toLocaleString()}</p>
      <div style={{ 
        backgroundColor: '#fff', 
        padding: '20px', 
        borderRadius: '8px',
        marginTop: '20px',
        display: 'inline-block'
      }}>
        <h2>✅ React渲染成功</h2>
        <p>这个页面不依赖任何路由或复杂组件</p>
      </div>
    </div>
  );
};

export default DirectTest;