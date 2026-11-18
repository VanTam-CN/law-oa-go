import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Form, 
  Input, 
  Button, 
  message, 
  Tabs, 
  Upload,
  Avatar,
  Space,
  Divider
} from 'antd';
import { 
  UserOutlined, 
  MailOutlined, 
  PhoneOutlined, 
  HomeOutlined,
  LockOutlined,
  EditOutlined,
  SaveOutlined
} from '@ant-design/icons';
import { 
  getCurrentUser, 
  updateUserProfile, 
  changePassword 
} from '../../api/user';

const { TabPane } = Tabs;

export default function UserProfile() {
  const [form] = Form.useForm();
  const [passwordForm] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [user, setUser] = useState({});
  const [activeTab, setActiveTab] = useState('1');
  const [editing, setEditing] = useState(false);

  useEffect(() => {
    loadUserProfile();
  }, []);

  const loadUserProfile = async () => {
    try {
      const userData = await getCurrentUser();
      setUser(userData);
      form.setFieldsValue(userData);
    } catch (error) {
      message.error('加载用户信息失败');
    }
  };

  const handleProfileUpdate = async (values) => {
    setLoading(true);
    try {
      await updateUserProfile({ ...values, userId: user.userId });
      message.success('个人信息更新成功');
      setEditing(false);
      loadUserProfile();
    } catch (error) {
      message.error('更新失败');
    } finally {
      setLoading(false);
    }
  };

  const handlePasswordChange = async (values) => {
    setLoading(true);
    try {
      await changePassword({ ...values, userId: user.userId });
      message.success('密码修改成功');
      passwordForm.resetFields();
    } catch (error) {
      message.error('密码修改失败');
    } finally {
      setLoading(false);
    }
  };

  const handleAvatarChange = (info) => {
    if (info.file.status === 'done') {
      message.success('头像上传成功');
      loadUserProfile();
    } else if (info.file.status === 'error') {
      message.error('头像上传失败');
    }
  };

  return (
    <div className="user-profile">
      <Card>
        <div className="profile-header">
          <Space size="large" align="start">
            <div className="avatar-section">
              <Upload
                name="avatar"
                listType="picture-circle"
                className="avatar-uploader"
                showUploadList={false}
                action="/api/upload/avatar"
                onChange={handleAvatarChange}
              >
                {user.avatar ? (
                  <Avatar size={120} src={user.avatar} />
                ) : (
                  <Avatar size={120} icon={<UserOutlined />} />
                )}
              </Upload>
              <div className="avatar-tip">
                点击更换头像
              </div>
            </div>
            
            <div className="user-info">
              <h2>{user.realName || user.username}</h2>
              <p>{user.email}</p>
              <Space>
                <span>{user.departmentName}</span>
                <Divider type="vertical" />
                <span>{user.position}</span>
                <Divider type="vertical" />
                <span>{user.employeeId}</span>
              </Space>
            </div>
          </Space>
        </div>

        <Tabs activeKey={activeTab} onChange={setActiveTab}>
          <TabPane tab="基本信息" key="1">
            <Card title="个人信息" size="small">
              <Form
                form={form}
                layout="vertical"
                onFinish={handleProfileUpdate}
              >
                <Form.Item
                  name="username"
                  label="用户名"
                  rules={[{ required: true, message: '请输入用户名' }]}
                >
                  <Input 
                    prefix={<UserOutlined />} 
                    disabled={!editing}
                    placeholder="请输入用户名" 
                  />
                </Form.Item>

                <Form.Item
                  name="realName"
                  label="真实姓名"
                  rules={[{ required: true, message: '请输入真实姓名' }]}
                >
                  <Input 
                    prefix={<UserOutlined />} 
                    disabled={!editing}
                    placeholder="请输入真实姓名" 
                  />
                </Form.Item>

                <Form.Item
                  name="email"
                  label="邮箱"
                  rules={[
                    { required: true, message: '请输入邮箱' },
                    { type: 'email', message: '请输入有效的邮箱地址' }
                  ]}
                >
                  <Input 
                    prefix={<MailOutlined />} 
                    disabled={!editing}
                    placeholder="请输入邮箱" 
                  />
                </Form.Item>

                <Form.Item
                  name="phone"
                  label="手机号"
                  rules={[
                    { required: true, message: '请输入手机号' },
                    { pattern: /^1[3-9]\d{9}$/, message: '请输入有效的手机号' }
                  ]}
                >
                  <Input 
                    prefix={<PhoneOutlined />} 
                    disabled={!editing}
                    placeholder="请输入手机号" 
                  />
                </Form.Item>

                <Form.Item
                  name="employeeId"
                  label="员工编号"
                  rules={[{ required: true, message: '请输入员工编号' }]}
                >
                  <Input 
                    prefix={<UserOutlined />} 
                    disabled={!editing}
                    placeholder="请输入员工编号" 
                  />
                </Form.Item>

                <Form.Item
                  name="position"
                  label="职位"
                >
                  <Input 
                    prefix={<HomeOutlined />} 
                    disabled={!editing}
                    placeholder="请输入职位" 
                  />
                </Form.Item>

                <Form.Item>
                  <Space>
                    {editing ? (
                      <>
                        <Button 
                          type="primary" 
                          htmlType="submit" 
                          icon={<SaveOutlined />}
                          loading={loading}
                        >
                          保存
                        </Button>
                        <Button 
                          onClick={() => {
                            setEditing(false);
                            form.setFieldsValue(user);
                          }}
                        >
                          取消
                        </Button>
                      </>
                    ) : (
                      <Button 
                        type="primary" 
                        icon={<EditOutlined />}
                        onClick={() => setEditing(true)}
                      >
                        编辑
                      </Button>
                    )}
                  </Space>
                </Form.Item>
              </Form>
            </Card>
          </TabPane>

          <TabPane tab="修改密码" key="2">
            <Card title="修改密码" size="small">
              <Form
                form={passwordForm}
                layout="vertical"
                onFinish={handlePasswordChange}
              >
                <Form.Item
                  name="currentPassword"
                  label="当前密码"
                  rules={[{ required: true, message: '请输入当前密码' }]}
                >
                  <Input.Password 
                    prefix={<LockOutlined />} 
                    placeholder="请输入当前密码" 
                  />
                </Form.Item>

                <Form.Item
                  name="newPassword"
                  label="新密码"
                  rules={[
                    { required: true, message: '请输入新密码' },
                    { min: 8, max: 20, message: '密码长度为8-20个字符' },
                    { 
                      pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,20}$/,
                      message: '密码必须包含大小写字母、数字和特殊字符'
                    }
                  ]}
                >
                  <Input.Password 
                    prefix={<LockOutlined />} 
                    placeholder="请输入新密码" 
                  />
                </Form.Item>

                <Form.Item
                  name="confirmPassword"
                  label="确认密码"
                  dependencies={['newPassword']}
                  rules={[
                    { required: true, message: '请确认密码' },
                    ({ getFieldValue }) => ({
                      validator(_, value) {
                        if (!value || getFieldValue('newPassword') === value) {
                          return Promise.resolve();
                        }
                        return Promise.reject(new Error('两次输入的密码不一致'));
                      },
                    }),
                  ]}
                >
                  <Input.Password 
                    prefix={<LockOutlined />} 
                    placeholder="请确认密码" 
                  />
                </Form.Item>

                <Form.Item>
                  <Button 
                    type="primary" 
                    htmlType="submit" 
                    icon={<SaveOutlined />}
                    loading={loading}
                  >
                    修改密码
                  </Button>
                </Form.Item>
              </Form>
            </Card>
          </TabPane>

          <TabPane tab="账户信息" key="3">
            <Card title="账户信息" size="small">
              <div className="account-info">
                <div className="info-item">
                  <span className="label">用户ID：</span>
                  <span className="value">{user.userId}</span>
                </div>
                <div className="info-item">
                  <span className="label">用户名：</span>
                  <span className="value">{user.username}</span>
                </div>
                <div className="info-item">
                  <span className="label">状态：</span>
                  <span className="value">
                    {user.status === '1' ? (
                      <span style={{ color: 'green' }}>正常</span>
                    ) : user.status === '0' ? (
                      <span style={{ color: 'red' }}>禁用</span>
                    ) : (
                      <span style={{ color: 'orange' }}>冻结</span>
                    )}
                  </span>
                </div>
                <div className="info-item">
                  <span className="label">用户类型：</span>
                  <span className="value">
                    {user.userType === '1' ? '管理员' : 
                     user.userType === '2' ? '律师' : 
                     user.userType === '3' ? '助理' : 
                     user.userType === '4' ? '行政' : '未知'}
                  </span>
                </div>
                <div className="info-item">
                  <span className="label">创建时间：</span>
                  <span className="value">{user.createTime}</span>
                </div>
                <div className="info-item">
                  <span className="label">最后登录：</span>
                  <span className="value">{user.lastLoginTime || '从未登录'}</span>
                </div>
                <div className="info-item">
                  <span className="label">最后登录IP：</span>
                  <span className="value">{user.lastLoginIp || '未知'}</span>
                </div>
              </div>
            </Card>
          </TabPane>
        </Tabs>
      </Card>
    </div>
  );
}