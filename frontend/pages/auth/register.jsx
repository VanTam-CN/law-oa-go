import React, { useState } from 'react';
import { Card, Form, Input, Button, Select, message, Spin } from 'antd';
import { UserOutlined, LockOutlined, MailOutlined, PhoneOutlined, HomeOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { registerUser, getDepartmentTree, getRoleOptions } from '../../api/user';
import './register.less';

const { Option } = Select;

export default function RegisterPage() {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [departments, setDepartments] = useState([]);
  const [roles, setRoles] = useState([]);

  React.useEffect(() => {
    loadDepartments();
    loadRoles();
  }, []);

  const loadDepartments = async () => {
    try {
      const data = await getDepartmentTree();
      setDepartments(data || []);
    } catch (error) {
      console.error('Failed to load departments:', error);
    }
  };

  const loadRoles = async () => {
    try {
      const data = await getRoleOptions();
      setRoles(data || []);
    } catch (error) {
      console.error('Failed to load roles:', error);
    }
  };

  const onFinish = async (values) => {
    setLoading(true);
    try {
      await registerUser(values);
      message.success('注册成功！请登录');
      navigate('/auth/login');
    } catch (error) {
      console.error('Registration failed:', error);
      message.error(error.message || '注册失败');
    } finally {
      setLoading(false);
    }
  };

  const renderDepartmentTree = (depts) => {
    return depts.map(dept => (
      <Option key={dept.deptId} value={dept.deptId}>
        {dept.deptName}
      </Option>
    ));
  };

  return (
    <div className="register-container">
      <Card className="register-card" title="用户注册">
        <Spin spinning={loading}>
          <Form
            form={form}
            name="register"
            onFinish={onFinish}
            layout="vertical"
            scrollToFirstError
          >
            <Form.Item
              name="username"
              label="用户名"
              rules={[
                { required: true, message: '请输入用户名' },
                { min: 4, max: 20, message: '用户名长度为4-20个字符' },
                { pattern: /^[a-zA-Z0-9_]+$/, message: '用户名只能包含字母、数字和下划线' }
              ]}
            >
              <Input prefix={<UserOutlined />} placeholder="请输入用户名" />
            </Form.Item>

            <Form.Item
              name="password"
              label="密码"
              rules={[
                { required: true, message: '请输入密码' },
                { min: 8, max: 20, message: '密码长度为8-20个字符' },
                { 
                  pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,20}$/,
                  message: '密码必须包含大小写字母、数字和特殊字符'
                }
              ]}
            >
              <Input.Password prefix={<LockOutlined />} placeholder="请输入密码" />
            </Form.Item>

            <Form.Item
              name="confirmPassword"
              label="确认密码"
              dependencies={['password']}
              rules={[
                { required: true, message: '请确认密码' },
                ({ getFieldValue }) => ({
                  validator(_, value) {
                    if (!value || getFieldValue('password') === value) {
                      return Promise.resolve();
                    }
                    return Promise.reject(new Error('两次输入的密码不一致'));
                  },
                }),
              ]}
            >
              <Input.Password prefix={<LockOutlined />} placeholder="请确认密码" />
            </Form.Item>

            <Form.Item
              name="realName"
              label="真实姓名"
              rules={[{ required: true, message: '请输入真实姓名' }]}
            >
              <Input placeholder="请输入真实姓名" />
            </Form.Item>

            <Form.Item
              name="email"
              label="邮箱"
              rules={[
                { required: true, message: '请输入邮箱' },
                { type: 'email', message: '请输入有效的邮箱地址' }
              ]}
            >
              <Input prefix={<MailOutlined />} placeholder="请输入邮箱" />
            </Form.Item>

            <Form.Item
              name="phone"
              label="手机号"
              rules={[
                { required: true, message: '请输入手机号' },
                { pattern: /^1[3-9]\d{9}$/, message: '请输入有效的手机号' }
              ]}
            >
              <Input prefix={<PhoneOutlined />} placeholder="请输入手机号" />
            </Form.Item>

            <Form.Item
              name="employeeId"
              label="员工编号"
              rules={[{ required: true, message: '请输入员工编号' }]}
            >
              <Input placeholder="请输入员工编号" />
            </Form.Item>

            <Form.Item
              name="deptId"
              label="所属部门"
              rules={[{ required: true, message: '请选择所属部门' }]}
            >
              <Select placeholder="请选择所属部门">
                {renderDepartmentTree(departments)}
              </Select>
            </Form.Item>

            <Form.Item
              name="userType"
              label="用户类型"
              rules={[{ required: true, message: '请选择用户类型' }]}
              initialValue="2"
            >
              <Select placeholder="请选择用户类型">
                <Option value="2">律师</Option>
                <Option value="3">助理</Option>
                <Option value="4">行政</Option>
              </Select>
            </Form.Item>

            <Form.Item
              name="roleIds"
              label="角色"
              rules={[{ required: true, message: '请选择角色' }]}
            >
              <Select mode="multiple" placeholder="请选择角色">
                {roles.map(role => (
                  <Option key={role.roleId} value={role.roleId}>
                    {role.roleName}
                  </Option>
                ))}
              </Select>
            </Form.Item>

            <Form.Item
              name="gender"
              label="性别"
              rules={[{ required: true, message: '请选择性别' }]}
              initialValue="1"
            >
              <Select placeholder="请选择性别">
                <Option value="1">男</Option>
                <Option value="2">女</Option>
                <Option value="0">保密</Option>
              </Select>
            </Form.Item>

            <Form.Item
              name="position"
              label="职位"
            >
              <Input placeholder="请输入职位" />
            </Form.Item>

            <Form.Item>
              <Button type="primary" htmlType="submit" block loading={loading}>
                注册
              </Button>
            </Form.Item>

            <Form.Item>
              <div className="register-footer">
                已有账号？{' '}
                <Button type="link" onClick={() => navigate('/auth/login')}>
                  立即登录
                </Button>
              </div>
            </Form.Item>
          </Form>
        </Spin>
      </Card>
    </div>
  );
}