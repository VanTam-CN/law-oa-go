import React from 'react';

const MinimalTest: React.FC = () => {
  return (
    <div style={{ padding: '20px', backgroundColor: '#f0f0f0' }}>
      <h1>最简单的测试组件</h1>
      <p>如果你能看到这个内容，说明React和路由都工作正常。</p>
      <p>当前时间: {new Date().toLocaleString()}</p>
    </div>
  );
};

export default MinimalTest;