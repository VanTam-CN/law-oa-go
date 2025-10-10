import React, { useState, useEffect } from 'react';
import { Card, Button, Result, Spin, Alert } from 'antd';

const SystemTest: React.FC = () => {
  const [apiStatus, setApiStatus] = useState<'loading' | 'success' | 'error'>('loading');
  const [apiResponse, setApiResponse] = useState<any>(null);

  useEffect(() => {
    // 测试API连接
    const testApi = async () => {
      try {
        const response = await fetch('/captchaImage');
        const data = await response.json();
        if (data.code === 0) {
          setApiStatus('success');
          setApiResponse(data);
        } else {
          setApiStatus('error');
        }
      } catch (error) {
        console.error('API test failed:', error);
        setApiStatus('error');
      }
    };

    testApi();
  }, []);

  return (
    <div style={{ padding: '24px', backgroundColor: '#f5f5f5', minHeight: '100vh' }}>
      <Card title="系统测试页面" style={{ marginBottom: '24px' }}>
        <Alert
          message="系统状态检查"
          description="这个页面用于验证前端和后端系统是否正常工作"
          type="info"
          showIcon
          style={{ marginBottom: '16px' }}
        />
        
        <div style={{ marginBottom: '16px' }}>
          <h3>前端状态</h3>
          <Result
            status="success"
            title="前端运行正常"
            subTitle="React应用已成功加载并渲染"
          />
        </div>

        <div style={{ marginBottom: '16px' }}>
          <h3>后端API状态</h3>
          {apiStatus === 'loading' && (
            <div style={{ textAlign: 'center', padding: '24px' }}>
              <Spin size="large" />
              <p style={{ marginTop: '16px' }}>正在测试后端API连接...</p>
            </div>
          )}
          
          {apiStatus === 'success' && (
            <Result
              status="success"
              title="后端API正常"
              subTitle={`API响应成功，验证码功能正常 (UUID: ${apiResponse?.uuid})`}
            />
          )}
          
          {apiStatus === 'error' && (
            <Result
              status="error"
              title="后端API连接失败"
              subTitle="无法连接到后端API，请检查后端服务是否正常运行"
              extra={[
                <Button type="primary" key="retry" onClick={() => window.location.reload()}>
                  重试
                </Button>
              ]}
            />
          )}
        </div>
      </Card>

      <Card title="系统信息">
        <p><strong>前端服务器:</strong> http://localhost:3002</p>
        <p><strong>后端服务器:</strong> http://localhost:8082</p>
        <p><strong>React版本:</strong> 18.2.0</p>
        <p><strong>Ant Design版本:</strong> 5.x</p>
        <p><strong>测试时间:</strong> {new Date().toLocaleString()}</p>
      </Card>
    </div>
  );
};

export default SystemTest;