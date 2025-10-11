import React, { useState } from 'react';
import { Layout, Menu } from 'antd';
import {
  FileDoneOutlined,
  ProjectOutlined,
  SettingOutlined,
  ToolOutlined,
  MoneyCollectOutlined,
  UserOutlined,
  TeamOutlined,
  HomeOutlined,
  AppstoreOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import './sider.less';

const { Sider } = Layout;

const menuItems = [
  {
    key: 'approval',
    icon: <FileDoneOutlined />,
    label: '审批',
  },
  {
    key: 'business',
    icon: <ProjectOutlined />,
    label: '业务',
  },
  {
    key: 'admin',
    icon: <SettingOutlined />,
    label: '行政',
  },
  {
    key: 'tools',
    icon: <ToolOutlined />,
    label: '工具',
  },
  {
    key: 'finance',
    icon: <MoneyCollectOutlined />,
    label: '财务',
  },
  {
    key: 'system',
    icon: <AppstoreOutlined />,
    label: '系统管理',
    children: [
      {
        key: 'system/user-management',
        icon: <UserOutlined />,
        label: '用户管理',
      },
      {
        key: 'system/role-management',
        icon: <TeamOutlined />,
        label: '角色管理',
      },
      {
        key: 'system/department-management',
        icon: <HomeOutlined />,
        label: '部门管理',
      },
      {
        key: 'system/user-profile',
        icon: <UserOutlined />,
        label: '个人资料',
      },
    ],
  },
];

export default function AppSider() {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();

  const onMenuClick = ({ key }) => {
    navigate(`/${key}`);
  };

  return (
    <Sider 
      collapsible 
      collapsed={collapsed} 
      onCollapse={setCollapsed}
      width={220}
    >
      <div className="logo">OA系统</div>
      <Menu
        theme="dark"
        mode="inline"
        items={menuItems}
        onClick={onMenuClick}
      />
    </Sider>
  );
}