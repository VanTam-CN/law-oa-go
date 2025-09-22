import React from 'react';
import { Card, Form, Input, Button, Checkbox, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { login } from '@/services/auth';
import useAuth from '@/hooks/useAuth';
import './Login.less';

interface LoginFormValues {
  username: string;
  password: string;
  remember: boolean;
}

const LoginPage: React.FC = () => {
  const navigate = useNavigate();
  const { login: authLogin } = useAuth();
  const [form] = Form.useForm<LoginFormValues>();
  const [loading, setLoading] = React.useState(false);
  
  const onFinish = async (values: LoginFormValues) => {
    try {
      setLoading(true);
      const response = await login(values);
      
      // 处理后端返回的token格式
      const token = response.token || response.data?.token;
      const user = response.user || { 
        id: 1, 
        username: values.username, 
        real_name: 'Admin', 
        email: 'admin@example.com', 
        role: 'admin', 
        department: 'IT' 
      };
      
      if (token) {
        authLogin(token, user);
        message.success('登录成功');
        navigate('/');
      } else {
        throw new Error('未获取到有效的登录凭证');
      }
    } catch (error) {
      console.error('Login failed:', error);
      message.error('登录失败，请检查用户名和密码');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-container">
      <Card className="login-card" title="律所OA系统登录">
        <Form
          form={form}
          name="login"
          initialValues={{ remember: true }}
          onFinish={onFinish}
          size="large"
        >
          <Form.Item
            name="username"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input 
              prefix={<UserOutlined />} 
              placeholder="用户名" 
              autoComplete="username"
            />
          </Form.Item>
          <Form.Item
            name="password"
            rules={[{ required: true, message: '请输入密码' }]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="密码"
              autoComplete="current-password"
            />
          </Form.Item>
          <Form.Item name="remember" valuePropName="checked">
            <Checkbox>记住我</Checkbox>
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block loading={loading}>
              登录
            </Button>
          </Form.Item>
        </Form>
      </Card>
      <div className="login-footer">
        © {new Date().getFullYear()} 律所OA系统 - 版权所有
      </div>
    </div>
  );
};

export default LoginPage;