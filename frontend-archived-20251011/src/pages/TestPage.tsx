import React from 'react';

const TestPage: React.FC = () => {
  return (
    <div style={{ padding: '20px', backgroundColor: '#f0f0f0', minHeight: '100vh' }}>
      <h1>测试页面</h1>
      <p>如果你能看到这个页面，说明React应用正常运行。</p>
      <div style={{ marginTop: '20px' }}>
        <h2>测试项目：</h2>
        <ul>
          <li>✅ React组件渲染正常</li>
          <li>✅ 样式加载正常</li>
          <li>✅ 路由工作正常</li>
        </ul>
      </div>
    </div>
  );
};

export default TestPage;