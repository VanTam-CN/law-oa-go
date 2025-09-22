import React, { useEffect, useState } from 'react';
import { Card, Alert, Button, Spin } from 'antd';

const AuthTest: React.FC = () => {
  const [authState, setAuthState] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // 测试AuthContext状态
    const checkAuth = () => {
      try {
        // 检查localStorage状态
        const token = localStorage.getItem('token');
        const userInfo = localStorage.getItem('userInfo');
        const roles = localStorage.getItem('roles');
        const permissions = localStorage.getItem('permissions');
        
        setAuthState({
          token: token ? 'present' : 'missing',
          userInfo: userInfo ? 'present' : 'missing',
          roles: roles ? 'present' : 'missing',
          permissions: permissions ? 'present' : 'missing',
          env: process.env.NODE_ENV
        });
      } catch (error) {
        setAuthState({
          error: error.message,
          env: process.env.NODE_ENV
        });
      } finally {
        setLoading(false);
      }
    };

    checkAuth();
  }, []);

  const clearAuth = () => {
    localStorage.clear();
    window.location.reload();
  };

  if (loading) {
    return (
      <div style={{ padding: '24px', textAlign: 'center' }}>
        <Spin size="large" />
        <p>正在检查认证状态...</p>
      </div>
    );
  }

  return (
    <div style={{ padding: '24px' }}>
      <Card title="AuthContext 测试" style={{ marginBottom: '16px' }}>
        <Alert
          message="认证系统状态"
          description="检查AuthContext和localStorage状态"
          type="info"
          showIcon
        />
        
        <div style={{ marginTop: '16px' }}>
          <h4>认证状态:</h4>
          <pre style={{ 
            background: '#f5f5f5', 
            padding: '8px', 
            borderRadius: '4px',
            fontSize: '12px'
          }}>
            {JSON.stringify(authState, null, 2)}
          </pre>
        </div>

        <div style={{ marginTop: '16px' }}>
          <Button type="primary" onClick={clearAuth}>
            清除认证数据并刷新
          </Button>
        </div>

        {authState?.error && (
          <Alert
            message="错误"
            description={authState.error}
            type="error"
            style={{ marginTop: '16px' }}
          />
        )}
      </Card>

      <Card title="环境信息">
        <p><strong>当前环境:</strong> {process.env.NODE_ENV}</p>
        <p><strong>当前时间:</strong> {new Date().toLocaleString()}</p>
        <p><strong>User Agent:</strong> {navigator.userAgent}</p>
      </Card>
    </div>
  );
};

export default AuthTest;