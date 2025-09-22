import React, { useState, useEffect } from 'react';
import './Dashboard.module.css';
import { 
  Row, 
  Col, 
  Card, 
  Statistic, 
  List, 
  Timeline, 
  Tag, 
  Button, 
  Avatar, 
  Typography, 
  Space, 
  Divider,
  Badge,
  Dropdown,
  Menu,
  Progress
} from 'antd';
import {
  UserOutlined,
  BellOutlined,
  SettingOutlined,
  MoreOutlined,
  BarChartOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  CloseCircleOutlined,
  SyncOutlined,
  AlertOutlined,
  FileTextOutlined,
  DollarCircleOutlined,
  TeamOutlined,
  ThunderboltOutlined,
  TrophyOutlined,
  FireOutlined,
  RiseOutlined,
  FallOutlined,
  ExportOutlined,
  CustomerServiceOutlined,
  LineChartOutlined,
  CalendarOutlined
} from '@ant-design/icons';

const { Title, Text } = Typography;

// 设计系统配置
const DESIGN_SYSTEM = {
  colors: {
    primary: '#1677FF',
    info: '#1677FF',
    success: '#52C41A',
    warning: '#FA8C16',
    error: '#FF4D4F',
    neutral: '#8C8C8C',
    border: '#E5E7EB',
    background: '#F8F9FA'
  },
  spacing: {
    xs: 4,
    sm: 8,
    md: 12,
    lg: 16,
    xl: 20,
    xxl: 24
  },
  typography: {
    xs: { fontSize: '12px', lineHeight: '16px' },
    sm: { fontSize: '13px', lineHeight: '18px' },
    md: { fontSize: '14px', lineHeight: '20px' },
    lg: { fontSize: '16px', lineHeight: '24px' },
    xl: { fontSize: '18px', lineHeight: '28px' },
    xxl: { fontSize: '20px', lineHeight: '30px' }
  },
  radius: {
    sm: '6px',
    md: '8px',
    lg: '12px'
  }
};

// 状态颜色映射
const STATUS_COLORS = {
  '进行中': '#1677FF',
  '已完成': '#52C41A',
  '已暂停': '#FA8C16',
  '已取消': '#FF4D4F',
  '待审批': '#FA8C16',
  '已通过': '#52C41A',
  '已拒绝': '#FF4D4F',
  '已撤销': '#8C8C8C'
};

// 优先级映射
const priorityMap: Record<string, string> = {
  high: '高',
  medium: '中',
  low: '低'
};

// 模拟数据
const mockStatistics = {
  totalProjects: 24,
  completedProjects: 18,
  pendingApprovals: 5,
  activeClients: 8,
  projectStatus: {
    '进行中': 12,
    '已完成': 8,
    '已暂停': 3,
    '已取消': 1
  },
  approvalStatus: {
    '待审批': 5,
    '已通过': 15,
    '已拒绝': 2,
    '已撤销': 1
  },
  financeStats: {
    totalRevenue: 2500000,
    totalExpenses: 1800000
  }
};

const mockTodos = [
  {
    id: 1,
    title: '审批张三的请假申请',
    priority: 'high',
    deadline: '2024-01-15',
    assignee: '李四',
    completed: false
  },
  {
    id: 2,
    title: '更新项目进度报告',
    priority: 'medium',
    deadline: '2024-01-16',
    assignee: '王五',
    completed: false
  },
  {
    id: 3,
    title: '客户会议准备',
    priority: 'high',
    deadline: '2024-01-14',
    assignee: '赵六',
    completed: false
  },
  {
    id: 4,
    title: '整理合同文档',
    priority: 'low',
    deadline: '2024-01-18',
    assignee: '钱七',
    completed: false
  }
];

const mockActivities = [
  {
    id: 1,
    type: 'approval',
    title: '李四审批了项目A的预算申请',
    user: '李四',
    status: '已通过',
    createdAt: '10分钟前'
  },
  {
    id: 2,
    type: 'project',
    title: '王五更新了项目B的进度',
    user: '王五',
    status: '进行中',
    createdAt: '30分钟前'
  },
  {
    id: 3,
    type: 'finance',
    title: '赵六提交了费用报销申请',
    user: '赵六',
    status: '待审批',
    createdAt: '1小时前'
  },
  {
    id: 4,
    type: 'approval',
    title: '张三完成了客户合同审查',
    user: '张三',
    status: '已通过',
    createdAt: '2小时前'
  },
  {
    id: 5,
    type: 'project',
    title: '钱七启动了新的诉讼案件',
    user: '钱七',
    status: '进行中',
    createdAt: '3小时前'
  },
  {
    id: 6,
    type: 'finance',
    title: '孙八更新了财务报表',
    user: '孙八',
    status: '已完成',
    createdAt: '4小时前'
  }
];

// 扩展活动列表用于无缝滚动
const extendedActivities = [...mockActivities, ...mockActivities];

const mockUser = {
  real_name: '张三',
  avatar: null
};

const mockKPIs = {
  netProfit: 700000,
  profitMargin: 28.0,
  completionRate: 75.0
};

interface DashboardProps {
  statistics?: typeof mockStatistics;
  todos?: typeof mockTodos;
  activities?: typeof mockActivities;
  user?: typeof mockUser;
  kpis?: typeof mockKPIs;
  loading?: boolean;
}

const Dashboard: React.FC<DashboardProps> = ({
  statistics = mockStatistics,
  todos = mockTodos,
  activities = mockActivities,
  user = mockUser,
  kpis = mockKPIs,
  loading = false
}) => {
  // 渲染优先级标签
  const renderPriorityTag = (priority: string) => {
    const color = priority === 'high' ? 'red' : priority === 'medium' ? 'orange' : 'green';
    return (
      <Tag 
        color={color} 
        style={{
          fontSize: DESIGN_SYSTEM.typography.xs.fontSize,
          fontWeight: 500,
          borderRadius: DESIGN_SYSTEM.radius.sm
        }}
      >
        {priorityMap[priority] || priority}
      </Tag>
    );
  };

  // 渲染状态标签
  const renderStatusTag = (status: string) => {
    const color = STATUS_COLORS[status as keyof typeof STATUS_COLORS] || DESIGN_SYSTEM.colors.neutral;
    return (
      <Tag 
        color={color} 
        style={{
          fontSize: DESIGN_SYSTEM.typography.xs.fontSize,
          fontWeight: 500,
          borderRadius: DESIGN_SYSTEM.radius.sm
        }}
      >
        {status}
      </Tag>
    );
  };

  // 项目状态数据转换为图表格式
  const projectStatusData = statistics?.projectStatus ? 
    Object.entries(statistics.projectStatus).map(([name, value]) => ({
      name,
      value,
      color: STATUS_COLORS[name as keyof typeof STATUS_COLORS] || DESIGN_SYSTEM.colors.neutral,
      percentage: statistics.totalProjects > 0 ? ((value / statistics.totalProjects) * 100).toFixed(1) : '0'
    })) : [];

  // 处理状态项点击跳转
  const handleStatusItemClick = (status: string) => {
    console.log(`跳转到${status}项目列表`);
  };

  // 动态欢迎语生成
  const getDynamicWelcome = () => {
    const hour = new Date().getHours();
    const userName = user?.real_name || '用户';
    const urgentTodos = todos.filter(todo => 
      !todo.completed && 
      todo.priority === 'high' && 
      new Date(todo.deadline) < new Date()
    ).length;
    const pendingApprovals = statistics?.pendingApprovals || 0;
    
    let timeGreeting = '';
    if (hour < 12) {
      timeGreeting = '早上好';
    } else if (hour < 18) {
      timeGreeting = '下午好';
    } else {
      timeGreeting = '晚上好';
    }
    
    let guidance = '';
    if (urgentTodos > 0 || pendingApprovals > 0) {
      guidance = `您今天有 ${urgentTodos} 个紧急事项和 ${pendingApprovals} 个待审批需要处理。`;
    } else if (todos.filter(todo => !todo.completed).length > 0) {
      guidance = `您今天有 ${todos.filter(todo => !todo.completed).length} 个待办事项需要处理。`;
    } else {
      guidance = '今天没有紧急事项，祝您工作愉快！';
    }
    
    return {
      greeting: `${timeGreeting}，${userName}`,
      guidance: guidance
    };
  };

  // KPI数据增强
  const getEnhancedKPIs = () => {
    if (!statistics) return null;
    
    const totalRevenue = statistics.financeStats?.totalRevenue || 0;
    const totalExpenses = statistics.financeStats?.totalExpenses || 0;
    const netProfit = totalRevenue - totalExpenses;
    const profitMargin = totalRevenue > 0 ? (netProfit / totalRevenue) * 100 : 0;
    
    const completionRate = statistics.totalProjects > 0 
      ? ((statistics.projectStatus?.['已完成'] || 0) / statistics.totalProjects) * 100 
      : 0;
    
    const profitTrend = netProfit > 100000 ? '+5.2%' : '-2.1%';
    const completionTrend = completionRate > 40 ? '+3.8%' : '-1.2%';
    
    return {
      netProfit: {
        value: netProfit,
        trend: profitTrend,
        trendUp: profitTrend.startsWith('+'),
        period: '本月'
      },
      completionRate: {
        value: completionRate,
        trend: completionTrend,
        trendUp: completionTrend.startsWith('+'),
        period: '年度'
      },
      pendingApprovals: statistics.pendingApprovals || 0,
      activeClients: statistics.activeClients || 2
    };
  };

  // 处理KPI卡片点击
  const handleKPIClick = (kpiType: string) => {
    console.log(`跳转到${kpiType}详情页面`);
  };

  // 骨架屏组件
  const renderSkeleton = () => (
    <div style={{ padding: '12px', backgroundColor: '#f8f9fa' }}>
      <div>加载中...</div>
    </div>
  );

  const enhancedKPIs = getEnhancedKPIs();
  const dynamicWelcome = getDynamicWelcome();

  if (loading) {
    return renderSkeleton();
  }

  return (
    <div className="dashboard-balanced-professional" style={{ 
      padding: '16px', 
      backgroundColor: '#f8f9fa', 
      minHeight: '100vh',
      background: 'linear-gradient(180deg, #f8f9fa 0%, #e9ecef 100%)'
    }}>
      {/* 顶部欢迎区域 */}
      <div style={{
        background: 'linear-gradient(135deg, #ffffff 0%, #f8f9fa 100%)',
        borderRadius: '12px',
        padding: '24px',
        marginBottom: '20px',
        boxShadow: '0 4px 16px rgba(0, 0, 0, 0.08)',
        border: '1px solid rgba(255, 255, 255, 0.9)'
      }}>
        <Row align="middle" gutter={[DESIGN_SYSTEM.spacing.xl, DESIGN_SYSTEM.spacing.lg]}>
          <Col xs={24} lg={10}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '20px' }}>
              <Avatar 
                size={56} 
                icon={<UserOutlined />} 
                style={{ 
                  background: 'linear-gradient(135deg, #1677FF 0%, #059669 100%)',
                  border: '3px solid rgba(255, 255, 255, 0.9)',
                  boxShadow: '0 4px 12px rgba(22, 119, 255, 0.25)',
                  flexShrink: 0
                }}
              />
              <div style={{ flex: 1 }}>
                <Title level={3} style={{ 
                  margin: 0, 
                  fontSize: '20px', 
                  lineHeight: '28px',
                  background: 'linear-gradient(135deg, #1677FF, #059669)',
                  WebkitBackgroundClip: 'text',
                  WebkitTextFillColor: 'transparent',
                  fontWeight: '600'
                }}>
                  {dynamicWelcome.greeting}
                </Title>
                <Text style={{ 
                  fontSize: '13px', 
                  color: '#666',
                  display: 'block',
                  marginTop: '2px',
                  marginBottom: '8px'
                }}>
                  {new Date().toLocaleDateString('zh-CN', { 
                    year: 'numeric', 
                    month: 'long', 
                    day: 'numeric',
                    weekday: 'long'
                  })}
                </Text>
              </div>
            </div>
          </Col>
          
          <Col xs={24} lg={8}>
            <div>
              <div style={{ 
                marginBottom: '12px',
                display: 'flex',
                alignItems: 'center',
                gap: '8px'
              }}>
                <ThunderboltOutlined style={{ color: DESIGN_SYSTEM.colors.primary, fontSize: '16px' }} />
                <span style={{ fontSize: '14px', fontWeight: '600', color: '#1f2937' }}>快捷操作</span>
              </div>
            </div>
          </Col>
          
          <Col xs={24} lg={6}>
            <div style={{ textAlign: 'right' }}>
              <Space size={DESIGN_SYSTEM.spacing.md}>
                <Badge 
                  count={statistics?.pendingApprovals || 0} 
                  size="small"
                  style={{ 
                    backgroundColor: statistics?.pendingApprovals > 0 ? '#FF4D4F' : '#52C41A'
                  }}
                >
                  <Button 
                    type="text" 
                    icon={<BellOutlined />} 
                    size="middle"
                    style={{
                      borderRadius: '8px',
                      width: '40px',
                      height: '40px'
                    }}
                  />
                </Badge>
              </Space>
            </div>
          </Col>
        </Row>
      </div>

      {/* 核心KPI指标 */}
      <div style={{
        background: 'linear-gradient(135deg, #ffffff 0%, #f8f9fa 100%)',
        borderRadius: '16px',
        padding: '20px',
        marginBottom: '16px',
        boxShadow: '0 2px 8px rgba(0, 0, 0, 0.06)',
        border: '1px solid rgba(255, 255, 255, 0.8)'
      }}>
        <Row gutter={[DESIGN_SYSTEM.spacing.lg, DESIGN_SYSTEM.spacing.lg]}>
          <Col xs={24} sm={12} lg={6}>
            <Card 
              hoverable
              onClick={() => handleKPIClick('profit')}
              style={{
                cursor: 'pointer',
                transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                borderLeft: '4px solid #059669'
              }}
            >
              <Statistic
                title="净利润"
                value={kpis?.netProfit || 0}
                suffix="元"
                valueStyle={{ color: '#059669' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card 
              hoverable
              onClick={() => handleKPIClick('completion')}
              style={{
                cursor: 'pointer',
                transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                borderLeft: '4px solid #1677FF'
              }}
            >
              <Statistic
                title="项目完成率"
                value={kpis?.completionRate || 0}
                suffix="%"
                valueStyle={{ color: '#1677FF' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card 
              hoverable
              onClick={() => handleKPIClick('approval')}
              style={{
                cursor: 'pointer',
                transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                borderLeft: '4px solid #FA8C16'
              }}
            >
              <Statistic
                title="待审批"
                value={statistics?.pendingApprovals || 0}
                valueStyle={{ color: '#FA8C16' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card 
              hoverable
              onClick={() => handleKPIClick('client')}
              style={{
                cursor: 'pointer',
                transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                borderLeft: '4px solid #722ED1'
              }}
            >
              <Statistic
                title="活跃客户"
                value={statistics?.activeClients || 0}
                valueStyle={{ color: '#722ED1' }}
              />
            </Card>
          </Col>
        </Row>
      </div>

      {/* 主要内容区域 */}
      <Row gutter={[DESIGN_SYSTEM.spacing.lg, DESIGN_SYSTEM.spacing.lg]}>
        <Col xs={24} lg={16}>
          {/* 项目状态分布 */}
          <div style={{
            background: 'linear-gradient(135deg, #ffffff 0%, #f8f9fa 100%)',
            borderRadius: '16px',
            padding: '16px',
            marginBottom: DESIGN_SYSTEM.spacing.lg,
            boxShadow: '0 2px 8px rgba(0, 0, 0, 0.06)',
            border: '1px solid rgba(255, 255, 255, 0.8)'
          }}>
            <Card 
              size="small" 
              title={
                <Space>
                  <BarChartOutlined style={{ color: DESIGN_SYSTEM.colors.primary }} />
                  <span style={{ fontSize: '14px', fontWeight: '600' }}>项目状态分布</span>
                  <Badge count={statistics?.totalProjects || 0} style={{ backgroundColor: DESIGN_SYSTEM.colors.primary }} />
                </Space>
              }
              style={{ marginBottom: 0, border: 'none', background: 'transparent' }}
              bodyStyle={{ padding: '12px', background: 'transparent' }}
            >
              <Row gutter={[DESIGN_SYSTEM.spacing.sm, DESIGN_SYSTEM.spacing.sm]}>
                {projectStatusData.length > 0 ? (
                  projectStatusData.map((item, index) => (
                    <Col xs={12} sm={6} key={index}>
                      <div 
                        onClick={() => handleStatusItemClick(item.name)}
                        style={{
                          cursor: 'pointer',
                          border: '1px solid #e5e7eb',
                          borderRadius: '8px',
                          padding: '8px',
                          background: '#ffffff',
                          transition: 'all 0.2s ease',
                          textAlign: 'center',
                          borderLeft: `3px solid ${item.color}`,
                          minHeight: '60px',
                          display: 'flex',
                          flexDirection: 'column',
                          justifyContent: 'center'
                        }}
                      >
                        <div style={{ fontSize: '16px', fontWeight: 'bold', color: item.color, lineHeight: 1 }}>
                          {item.value}
                        </div>
                        <div style={{ fontSize: '11px', color: '#666', marginTop: '2px' }}>
                          {item.name}
                        </div>
                        <div style={{ fontSize: '10px', color: '#999', marginTop: '1px' }}>
                          {item.percentage}%
                        </div>
                      </div>
                    </Col>
                  ))
                ) : (
                  <Col span={24}>
                    <div style={{ textAlign: 'center', color: '#999', fontSize: '12px', padding: '20px' }}>
                      暂无项目数据
                    </div>
                  </Col>
                )}
              </Row>
            </Card>
          </div>

          {/* 审批状态概览 */}
          <div style={{
            background: 'linear-gradient(135deg, #ffffff 0%, #f8f9fa 100%)',
            borderRadius: '16px',
            padding: '16px',
            marginBottom: DESIGN_SYSTEM.spacing.lg,
            boxShadow: '0 2px 8px rgba(0, 0, 0, 0.06)',
            border: '1px solid rgba(255, 255, 255, 0.8)'
          }}>
            <Card 
              size="small" 
              title={
                <Space>
                  <CheckCircleOutlined style={{ color: DESIGN_SYSTEM.colors.success }} />
                  <span style={{ fontSize: '14px', fontWeight: '600' }}>审批状态概览</span>
                  <Badge 
                    count={Object.values(statistics?.approvalStatus || {}).reduce((sum, count) => sum + count, 0)} 
                    style={{ backgroundColor: DESIGN_SYSTEM.colors.success }}
                  />
                </Space>
              }
              style={{ border: 'none', background: 'transparent' }}
              bodyStyle={{ padding: '12px', background: 'transparent' }}
            >
              <Row gutter={[DESIGN_SYSTEM.spacing.sm, DESIGN_SYSTEM.spacing.sm]}>
                {statistics?.approvalStatus ? 
                  Object.entries(statistics.approvalStatus).map(([status, count]) => (
                    <Col xs={6} key={status}>
                      <div style={{
                        textAlign: 'center',
                        padding: '8px',
                        border: '1px solid #f0f0f0',
                        borderRadius: '6px',
                        minHeight: '50px',
                        display: 'flex',
                        flexDirection: 'column',
                        justifyContent: 'center'
                      }}>
                        <div style={{ 
                          fontSize: '14px', 
                          fontWeight: 'bold', 
                          color: STATUS_COLORS[status as keyof typeof STATUS_COLORS] || '#666',
                          lineHeight: 1
                        }}>
                          {count}
                        </div>
                        <div style={{ fontSize: '10px', color: '#666', marginTop: '2px' }}>
                          {status}
                        </div>
                      </div>
                    </Col>
                  )) : (
                  <Col span={24}>
                    <div style={{ textAlign: 'center', color: '#999', fontSize: '12px' }}>暂无审批数据</div>
                  </Col>
                )}
              </Row>
            </Card>
          </div>
        </Col>

        {/* 最新动态 */}
        <Col xs={24}>
          <div style={{
            background: 'linear-gradient(135deg, #ffffff 0%, #f8f9fa 100%)',
            borderRadius: '16px',
            padding: '16px',
            height: '300px',
            boxShadow: '0 2px 8px rgba(0, 0, 0, 0.06)',
            border: '1px solid rgba(255, 255, 255, 0.8)'
          }}>
            <Card 
              size="small"
              title={
                <Space>
                  <LineChartOutlined style={{ color: DESIGN_SYSTEM.colors.primary }} />
                  <span style={{ fontSize: '14px', fontWeight: '600' }}>最新动态</span>
                  <Badge count={activities.length} style={{ backgroundColor: DESIGN_SYSTEM.colors.primary }} />
                </Space>
              }
              style={{ border: 'none', background: 'transparent', height: 'calc(100% - 32px)' }}
              bodyStyle={{ padding: '0', height: 'calc(100% - 47px)', background: 'transparent' }}
            >
              <div style={{ overflow: 'auto', height: '100%' }}>
                {activities.map((activity) => (
                  <div key={activity.id} style={{ padding: '8px 0', borderBottom: '1px solid #f0f0f0' }}>
                    <div style={{ display: 'flex', alignItems: 'flex-start', gap: '8px' }}>
                      <Avatar size={24} icon={<UserOutlined />} />
                      <div style={{ flex: 1 }}>
                        <div style={{ fontSize: '12px', fontWeight: '500', marginBottom: '2px' }}>
                          {activity.title}
                        </div>
                        <div style={{ fontSize: '10px', color: '#666' }}>
                          {activity.user} · {activity.createdAt}
                        </div>
                      </div>
                      {renderStatusTag(activity.status)}
                    </div>
                  </div>
                ))}
              </div>
            </Card>
          </div>
        </Col>

        {/* 待办事项 */}
        <Col xs={24} lg={8}>
          <div style={{
            background: 'linear-gradient(135deg, #ffffff 0%, #f8f9fa 100%)',
            borderRadius: '16px',
            padding: '16px',
            height: '400px',
            boxShadow: '0 2px 8px rgba(0, 0, 0, 0.06)',
            border: '1px solid rgba(255, 255, 255, 0.8)'
          }}>
            <Card 
              size="small"
              title={
                <Space>
                  <AlertOutlined style={{ color: DESIGN_SYSTEM.colors.warning }} />
                  <span style={{ fontSize: '14px', fontWeight: '600' }}>待办事项</span>
                  <Badge count={todos.length} style={{ backgroundColor: DESIGN_SYSTEM.colors.warning }} />
                </Space>
              }
              style={{ border: 'none', background: 'transparent', height: 'calc(100% - 32px)' }}
              bodyStyle={{ padding: '8px', height: 'calc(100% - 47px)', overflow: 'auto', background: 'transparent' }}
            >
              <List
                size="small"
                dataSource={todos}
                renderItem={(item) => (
                  <List.Item style={{ padding: '12px 16px', border: 'none', borderBottom: '1px solid #f0f0f0' }}>
                    <div style={{ width: '100%' }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '6px' }}>
                        {renderPriorityTag(item.priority)}
                        <Button type="link" size="small" style={{ padding: 0, fontSize: '12px', height: 'auto', fontWeight: '500' }}>
                          处理
                        </Button>
                      </div>
                      <div style={{ fontSize: '14px', fontWeight: '500', color: '#1f2937', marginBottom: '6px', lineHeight: '1.4' }}>
                        {item.title}
                      </div>
                      <div style={{ fontSize: '12px', color: '#6b7280' }}>
                        <Space split={<span style={{ color: '#d1d5db' }}>|</span>}>
                          <span><CalendarOutlined style={{ fontSize: '8px' }} /> {item.deadline}</span>
                          <span><UserOutlined style={{ fontSize: '8px' }} /> {item.assignee}</span>
                        </Space>
                      </div>
                    </div>
                  </List.Item>
                )}
              />
            </Card>
          </div>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;