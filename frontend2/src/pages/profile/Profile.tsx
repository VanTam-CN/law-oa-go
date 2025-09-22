import React, { useState } from 'react';
import { Card, Form, Input, Button, Upload, message, Avatar, Space, Divider } from 'antd';
import { UserOutlined, MailOutlined, PhoneOutlined, BankOutlined, EditOutlined, SaveOutlined } from '@ant-design/icons';
import type { UploadFile } from 'antd/es/upload/interface';

const { Item: FormItem } = Form;

const Profile: React.FC = () => {
  const [form] = Form.useForm();
  const [editing, setEditing] = useState(false);
  const [loading, setLoading] = useState(false);
  const [avatarUrl, setAvatarUrl] = useState<string>('');

  // 模拟用户数据
  const userData = {
    username: 'dev_user',
    realName: '开发用户',
    email: 'dev@example.com',
    phone: '13800138000',
    department: '技术部',
    position: '高级开发工程师',
    bio: '专注于前端开发和用户体验设计',
    address: '北京市朝阳区',
    joinDate: '2024-01-01'
  };

  const handleEdit = () => {
    form.setFieldsValue(userData);
    setEditing(true);
  };

  const handleSave = async () => {
    try {
      setLoading(true);
      const values = await form.validateFields();
      console.log('保存用户信息:', values);
      message.success('个人信息保存成功！');
      setEditing(false);
    } catch (error) {
      console.error('保存失败:', error);
      message.error('保存失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = () => {
    setEditing(false);
    form.resetFields();
  };

  const handleAvatarChange = (info: any) => {
    if (info.file.status === 'done') {
      setAvatarUrl(info.file.response?.url || '');
      message.success('头像上传成功');
    } else if (info.file.status === 'error') {
      message.error('头像上传失败');
    }
  };

  return (
    <div style={{ padding: '24px' }}>
      <Card title="个人中心" style={{ maxWidth: 800, margin: '0 auto' }}>
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <Avatar
            size={100}
            src={avatarUrl}
            icon={<UserOutlined />}
            style={{ marginBottom: 16 }}
          />
          <div>
            <Upload
              name="avatar"
              listType="picture-circle"
              className="avatar-uploader"
              showUploadList={false}
              action="/api/upload"
              onChange={handleAvatarChange}
              disabled={!editing}
            >
              {editing ? (
                <Button type="link" icon={<EditOutlined />}>
                  更换头像
                </Button>
              ) : null}
            </Upload>
          </div>
        </div>

        <Divider />

        <Form
          form={form}
          layout="vertical"
          initialValues={userData}
        >
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
            <FormItem
              label="用户名"
              name="username"
              rules={[{ required: true, message: '请输入用户名' }]}
            >
              <Input 
                prefix={<UserOutlined />} 
                disabled={!editing}
                placeholder="请输入用户名"
              />
            </FormItem>

            <FormItem
              label="真实姓名"
              name="realName"
              rules={[{ required: true, message: '请输入真实姓名' }]}
            >
              <Input 
                prefix={<UserOutlined />} 
                disabled={!editing}
                placeholder="请输入真实姓名"
              />
            </FormItem>

            <FormItem
              label="邮箱"
              name="email"
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
            </FormItem>

            <FormItem
              label="手机号"
              name="phone"
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
            </FormItem>

            <FormItem
              label="部门"
              name="department"
            >
              <Input 
                prefix={<BankOutlined />} 
                disabled={!editing}
                placeholder="请输入部门"
              />
            </FormItem>

            <FormItem
              label="职位"
              name="position"
            >
              <Input 
                disabled={!editing}
                placeholder="请输入职位"
              />
            </FormItem>
          </div>

          <FormItem
            label="地址"
            name="address"
          >
            <Input 
              disabled={!editing}
              placeholder="请输入地址"
            />
          </FormItem>

          <FormItem
            label="个人简介"
            name="bio"
          >
            <Input.TextArea 
              rows={4}
              disabled={!editing}
              placeholder="请输入个人简介"
            />
          </FormItem>

          {editing && (
            <FormItem>
              <Space>
                <Button 
                  type="primary" 
                  icon={<SaveOutlined />}
                  onClick={handleSave}
                  loading={loading}
                >
                  保存
                </Button>
                <Button onClick={handleCancel}>
                  取消
                </Button>
              </Space>
            </FormItem>
          )}
        </Form>

        {!editing && (
          <div style={{ textAlign: 'center', marginTop: 24 }}>
            <Button type="primary" icon={<EditOutlined />} onClick={handleEdit}>
              编辑资料
            </Button>
          </div>
        )}
      </Card>
    </div>
  );
};

export default Profile;