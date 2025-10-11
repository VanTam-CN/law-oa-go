import React, { useState, useEffect } from 'react';
import { Card, Form, Input, Button, Upload, message, Avatar, Space, Divider, Modal } from 'antd';
import { UserOutlined, MailOutlined, PhoneOutlined, BankOutlined, EditOutlined, SaveOutlined, LockOutlined } from '@ant-design/icons';
import type { UploadFile } from 'antd/es/upload/interface';
import useAuth from '@/hooks/useAuth';
import { updateProfile, uploadAvatar, changePassword } from '@/api/auth';

const { Item: FormItem } = Form;

interface ProfileData {
  id?: number;
  username?: string;
  real_name?: string;
  email?: string;
  phone?: string;
  department?: string;
  position?: string;
  bio?: string;
  address?: string;
  avatar?: string;
  created_at?: string;
}

const Profile: React.FC = () => {
  const { user, updateUser } = useAuth();
  const [form] = Form.useForm();
  const [editing, setEditing] = useState(false);
  const [loading, setLoading] = useState(false);
  const [avatarUrl, setAvatarUrl] = useState<string>('');
  const [userData, setUserData] = useState<ProfileData>({});
  const [passwordModalVisible, setPasswordModalVisible] = useState(false);
  const [passwordLoading, setPasswordLoading] = useState(false);
  const [passwordForm] = Form.useForm();

  // 获取用户数据
  useEffect(() => {
    if (user) {
      const profileData: ProfileData = {
        id: user.id,
        username: user.username,
        real_name: user.real_name,
        email: user.email,
        phone: user.phone || '',
        department: user.department || '',
        position: user.position || '',
        bio: user.bio || '',
        address: user.address || '',
        avatar: user.avatar || '',
        created_at: user.created_at
      };
      setUserData(profileData);
      form.setFieldsValue(profileData);
      if (user.avatar) {
        setAvatarUrl(user.avatar);
      }
    }
  }, [user, form]);

  const handleEdit = () => {
    form.setFieldsValue(userData);
    setEditing(true);
  };

  const handleSave = async () => {
    try {
      setLoading(true);
      const values = await form.validateFields();
      
      // 调用后端API更新用户信息
      const updateData = {
        real_name: values.realName,
        phone: values.phone,
        department: values.department,
        position: values.position,
        bio: values.bio,
        address: values.address
      };
      
      await updateProfile(updateData);
      
      // 更新本地状态
      const updatedUser = { ...user, ...updateData };
      setUserData(prev => ({ ...prev, ...updateData }));
      updateUser(updatedUser);
      message.success('个人信息保存成功！');
      setEditing(false);
    } catch (error: any) {
      console.error('保存失败:', error);
      message.error(error.response?.data?.message || '保存失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = () => {
    setEditing(false);
    form.resetFields();
  };

  const handleAvatarChange = async (info: any) => {
    if (info.file.status === 'uploading') {
      return;
    }
    
    if (info.file.status === 'done') {
      try {
        const formData = new FormData();
        formData.append('avatar', info.file.originFileObj);
        
        const response = await uploadAvatar(formData);
        if (response.url) {
          setAvatarUrl(response.url);
          // 更新用户头像信息
          const updatedUser = { ...user, avatar: response.url };
          updateUser(updatedUser);
          message.success('头像上传成功');
        }
      } catch (error: any) {
        message.error(error.response?.data?.message || '头像上传失败');
      }
    } else if (info.file.status === 'error') {
      message.error('头像上传失败');
    }
  };

  const handleChangePassword = async () => {
    try {
      setPasswordLoading(true);
      const values = await passwordForm.validateFields();
      
      await changePassword({
        old_password: values.oldPassword,
        new_password: values.newPassword
      });
      
      message.success('密码修改成功');
      setPasswordModalVisible(false);
      passwordForm.resetFields();
    } catch (error: any) {
      console.error('密码修改失败:', error);
      message.error(error.response?.data?.message || '密码修改失败，请重试');
    } finally {
      setPasswordLoading(false);
    }
  };

  const showPasswordModal = () => {
    setPasswordModalVisible(true);
  };

  const handlePasswordModalCancel = () => {
    setPasswordModalVisible(false);
    passwordForm.resetFields();
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
              customRequest={({ file, onSuccess, onError }) => {
                // 自定义上传请求，使用我们的API
                const formData = new FormData();
                formData.append('avatar', file);
                
                uploadAvatar(formData)
                  .then(response => {
                    onSuccess(response);
                  })
                  .catch(error => {
                    onError(error);
                  });
              }}
              onChange={handleAvatarChange}
              disabled={!editing}
              accept="image/*"
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
            <Space>
              <Button type="primary" icon={<EditOutlined />} onClick={handleEdit}>
                编辑资料
              </Button>
              <Button icon={<LockOutlined />} onClick={showPasswordModal}>
                修改密码
              </Button>
            </Space>
          </div>
        )}
      </Card>

      {/* 修改密码模态框 */}
      <Modal
        title="修改密码"
        open={passwordModalVisible}
        onOk={handleChangePassword}
        onCancel={handlePasswordModalCancel}
        confirmLoading={passwordLoading}
        okText="确认修改"
        cancelText="取消"
      >
        <Form form={passwordForm} layout="vertical">
          <FormItem
            label="当前密码"
            name="oldPassword"
            rules={[{ required: true, message: '请输入当前密码' }]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="请输入当前密码" />
          </FormItem>
          <FormItem
            label="新密码"
            name="newPassword"
            rules={[
              { required: true, message: '请输入新密码' },
              { min: 6, message: '密码至少6位' }
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="请输入新密码" />
          </FormItem>
          <FormItem
            label="确认新密码"
            name="confirmPassword"
            dependencies={['newPassword']}
            rules={[
              { required: true, message: '请确认新密码' },
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
            <Input.Password prefix={<LockOutlined />} placeholder="请确认新密码" />
          </FormItem>
        </Form>
      </Modal>
    </div>
  );
};

export default Profile;