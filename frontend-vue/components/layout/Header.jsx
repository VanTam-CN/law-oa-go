import React from 'react';
import { Layout, Dropdown, Space, Badge, Avatar } from 'antd';
import { BellOutlined, UserOutlined } from '@ant-design/icons';
import './header.less';

const { Header } = Layout;

const items = [
  {
    key: '1',
    label: '个人中心',
  },
  {
    key: '2',
    label: '退出登录',
  },
];

export default function AppHeader() {
  return (
    <Header className="app-header">
      <div className="header-right">
        <Space size="large">
          <Badge count={5}>
            <BellOutlined style={{ fontSize: '18px' }} />
          </Badge>
          <Dropdown menu={{ items }}>
            <Space>
              <Avatar icon={<UserOutlined />} />
              <span>管理员</span>
            </Space>
          </Dropdown>
        </Space>
      </div>
    </Header>
  );
}