import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router';
import './Dashboard.module.css';
import { dataService } from '@/services/dataService';
import { useAppStore } from '@/stores/useAppStore';
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
  Progress,
  message
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
  RiseOutlined,
  FallOutlined,
  ExportOutlined,
  CustomerServiceOutlined,
  LineChartOutlined,
  CalendarOutlined
} from '@ant-design/icons';

const { Title, Text } = Typography;

// 设计系统配置 - 严格按照规范
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

// 初始空数据 - 将通过API获取真实数据
const initialStatistics = {
  totalProjects: 0,
  completedProjects: 0,
  pendingApprovals: 0,
  activeClients: 0,
  projectStatus: {},
  approvalStatus: {},
  financeStats: {
    totalRevenue: 0,
    totalExpenses: 0
  }
};

interface DashboardProps {
  statistics?: any;
  todos?: any[];
  activities?: any[];
  user?: any;
  kpis?: any;
  loading?: boolean;
}

const Dashboard: React.FC<DashboardProps> = () => {
  // 导航功能
  const navigate = useNavigate();

  // 全局状态管理
  const { user: currentUser } = useAppStore();

  // 状态管理
  const [statistics, setStatistics] = useState<any>(null);
  const [todos, setTodos] = useState<any[]>([]);
  const [activities, setActivities] = useState<any[]>([]);
  const [user, setUser] = useState<any>(null);
  const [kpis, setKpis] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  // 获取仪表盘数据
  useEffect(() => {
    const fetchDashboardData = async () => {
      try {
        setLoading(true);

        // 并行获取所有数据
        const [statsData, todosData, activitiesData] = await Promise.all([
          dataService.getDashboardStatistics(),
          dataService.getDashboardTodos(),
          dataService.getDashboardActivities()
        ]);

        // 使用全局状态的用户信息
        const userData = currentUser || { real_name: '用户', avatar: null };

        setStatistics(statsData);
        setTodos(todosData);
        setActivities(activitiesData);
        setUser(userData);

        // 计算KPI数据
        if (statsData) {
          const netProfit = statsData.financeStats?.totalRevenue - statsData.financeStats?.totalExpenses || 0;
          const profitMargin = statsData.financeStats?.totalRevenue > 0
            ? ((netProfit / statsData.financeStats?.totalRevenue) * 100).toFixed(1)
            : 0;
          const completionRate = statsData.totalProjects > 0
            ? ((statsData.completedProjects / statsData.totalProjects) * 100).toFixed(1)
            : 0;

          setKpis({
            netProfit,
            profitMargin: Number(profitMargin),
            completionRate: Number(completionRate)
          });
        }
      } catch (error) {
        console.error('获取仪表盘数据失败:', error);
        // 显示用户友好的错误信息
        message.error('加载仪表盘数据失败，请稍后重试');

        // 发生错误时使用空数据作为降级方案
        setStatistics(initialStatistics);
        setTodos([]);
        setActivities([]);
        setUser({ real_name: '用户', avatar: null });
        setKpis({
          netProfit: 0,
          profitMargin: 0,
          completionRate: 0
        });
      } finally {
        setLoading(false);
      }
    };

    fetchDashboardData();
  }, []);
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
      todo.deadline && new Date(todo.deadline) < new Date()
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

  // 处理常用功能按钮点击
  const handleQuickActionClick = (action: string) => {
    switch (action) {
      case 'new-project':
        navigate('/case');
        break;
      case 'client-management':
        navigate('/client');
        break;
      case 'schedule':
        navigate('/calendar');
        break;
      case 'team':
        navigate('/team');
        break;
      case 'finance':
        navigate('/finance');
        break;
      case 'analytics':
        navigate('/reports');
        break;
      default:
        console.log(`未知功能: ${action}`);
    }
  };

  // 骨架屏组件
  const renderSkeleton = () => (
    <div style={{
      padding: '12px',
      backgroundColor: '#f8f9fa',
      display: 'flex',
      flexDirection: 'column',
      gap: '16px'
    }}>
      <div style={{
        height: '120px',
        background: 'linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%)',
        borderRadius: '8px',
        animation: 'pulse 2s infinite'
      }}></div>
      <div style={{
        height: '300px',
        background: 'linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%)',
        borderRadius: '8px',
        animation: 'pulse 2s infinite'
      }}></div>
      <div style={{
        height: '200px',
        background: 'linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%)',
        borderRadius: '8px',
        animation: 'pulse 2s infinite'
      }}></div>
    </div>
  );

  const enhancedKPIs = getEnhancedKPIs();
  const dynamicWelcome = getDynamicWelcome();

  if (loading) {
    return renderSkeleton();
  }

  return (
    <div className="dashboard-balanced-professional" style={{ 
      padding: '12px', 
      backgroundColor: '#f8f9fa', 
      minHeight: '100vh',
      background: 'linear-gradient(180deg, #f8f9fa 0%, #e9ecef 100%)'
    }}>
      {/* 顶部欢迎区域 - 重新设计 */}
      <div style={{
        background: 'linear-gradient(135deg, #ffffff 0%, #f8f9fa 100%)',
        borderRadius: '12px',
        padding: '16px',
        marginBottom: '16px',
        boxShadow: '0 2px 8px rgba(0, 0, 0, 0.06)',
        border: '1px solid rgba(255, 255, 255, 0.9)'
      }}>
        <Row align="middle" gutter={[16, 16]}>
          <Col xs={24} lg={18}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
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
                  fontSize: '18px', 
                  lineHeight: '24px',
                  background: 'linear-gradient(135deg, #1677FF, #059669)',
                  WebkitBackgroundClip: 'text',
                  WebkitTextFillColor: 'transparent',
                  fontWeight: '600'
                }}>
                  {dynamicWelcome.greeting}
                </Title>
                <Text style={{ 
                  fontSize: '12px', 
                  color: '#666',
                  display: 'block',
                  marginBottom: '8px'
                }}>
                  {new Date().toLocaleDateString('zh-CN', { 
                    year: 'numeric', 
                    month: 'long', 
                    day: 'numeric',
                    weekday: 'long'
                  })}
                </Text>
                <div style={{
                  background: 'linear-gradient(135deg, #E8F4FD 0%, #E1F5FE 100%)',
                  border: '1px solid #B3E5FC',
                  borderRadius: '8px',
                  padding: '8px 12px',
                  fontSize: '12px',
                  color: '#0277BD',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '8px'
                }}>
                  <AlertOutlined style={{ fontSize: '14px' }} />
                  <span>今日待办 <strong>{todos.filter(todo => !todo.completed).length}</strong> 项</span>
                  <Divider type="vertical" style={{ margin: '0 4px', borderColor: '#81D4FA' }} />
                  <span>紧急事项 <strong>{todos.filter(todo => todo.priority === 'high' && !todo.completed).length}</strong> 项</span>
                  <Divider type="vertical" style={{ margin: '0 4px', borderColor: '#81D4FA' }} />
                  <span>待审批 <strong>{statistics?.pendingApprovals || 0}</strong> 项</span>
                </div>
              </div>
            </div>
          </Col>
          
          <Col xs={24} lg={6}>
            <div style={{ textAlign: 'right' }}>
              <Space direction="vertical" size={8} style={{ width: '100%' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <span style={{ fontSize: '11px', color: '#666' }}>系统状态</span>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                    <div style={{ 
                      width: '6px', 
                      height: '6px', 
                      borderRadius: '50%', 
                      backgroundColor: '#52C41A',
                      animation: 'pulse 2s infinite'
                    }}></div>
                    <span style={{ fontSize: '10px', color: '#52C41A', fontWeight: '500' }}>正常运行</span>
                  </div>
                </div>
                <div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
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
                      size="small"
                      style={{
                        borderRadius: '6px',
                        width: '32px',
                        height: '32px'
                      }}
                    />
                  </Badge>
                  <Button 
                    type="text" 
                    icon={<SettingOutlined />} 
                    size="small"
                    style={{
                      borderRadius: '6px',
                      width: '32px',
                      height: '32px'
                    }}
                  />
                </div>
              </Space>
            </div>
          </Col>
        </Row>
      </div>

      {/* 核心KPI指标 */}
      <div style={{
        background: 'linear-gradient(135deg, #ffffff 0%, #f8f9fa 100%)',
        borderRadius: '12px',
        padding: '16px',
        marginBottom: '16px',
        boxShadow: '0 2px 8px rgba(0, 0, 0, 0.06)',
        border: '1px solid rgba(255, 255, 255, 0.8)'
      }}>
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={12} lg={6}>
            <Card 
              hoverable
              onClick={() => handleKPIClick('profit')}
              style={{
                cursor: 'pointer',
                transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                borderLeft: '4px solid #059669',
                height: '90px'
              }}
              bodyStyle={{ padding: '12px', height: '100%', display: 'flex', flexDirection: 'column', justifyContent: 'center' }}
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
                borderLeft: '4px solid #1677FF',
                height: '90px'
              }}
              bodyStyle={{ padding: '12px', height: '100%', display: 'flex', flexDirection: 'column', justifyContent: 'center' }}
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
                borderLeft: '4px solid #FA8C16',
                height: '90px'
              }}
              bodyStyle={{ padding: '12px', height: '100%', display: 'flex', flexDirection: 'column', justifyContent: 'center' }}
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
                borderLeft: '4px solid #722ED1',
                height: '90px'
              }}
              bodyStyle={{ padding: '12px', height: '100%', display: 'flex', flexDirection: 'column', justifyContent: 'center' }}
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

      {/* 主要内容区域 - 16:8 布局 */}
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={16}>
          {/* 项目状态分布 */}
          <div style={{
            background: 'linear-gradient(135deg, #ffffff 0%, #f8f9fa 100%)',
            borderRadius: '8px',
            padding: '12px',
            height: '150px',
            marginBottom: '12px',
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
              bodyStyle={{ padding: '8px', background: 'transparent' }}
            >
              <Row gutter={[8, 8]}>
                {projectStatusData.length > 0 ? (
                  projectStatusData.map((item, index) => (
                    <Col xs={12} sm={6} key={index}>
                      <div 
                        onClick={() => handleStatusItemClick(item.name)}
                        style={{
                          cursor: 'pointer',
                          border: '1px solid #e5e7eb',
                          borderRadius: '6px',
                          padding: '8px',
                          background: '#ffffff',
                          transition: 'all 0.2s ease',
                          textAlign: 'center',
                          borderLeft: `3px solid ${item.color}`,
                          minHeight: '55px',
                          display: 'flex',
                          flexDirection: 'column',
                          justifyContent: 'center'
                        }}
                        onMouseEnter={(e) => {
                          e.currentTarget.style.backgroundColor = '#f8f9fa';
                          e.currentTarget.style.transform = 'translateY(-1px)';
                          e.currentTarget.style.boxShadow = '0 4px 8px rgba(0,0,0,0.1)';
                        }}
                        onMouseLeave={(e) => {
                          e.currentTarget.style.backgroundColor = '#ffffff';
                          e.currentTarget.style.transform = 'translateY(0)';
                          e.currentTarget.style.boxShadow = 'none';
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
            borderRadius: '8px',
            padding: '12px',
            height: '150px',
            marginBottom: '12px',
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
              style={{ border: 'none', background: 'transparent', height: 'calc(100% - 16px)' }}
              bodyStyle={{ padding: '8px', background: 'transparent', height: 'calc(100% - 40px)' }}
            >
              <Row gutter={[8, 8]}>
                {statistics?.approvalStatus ? 
                  Object.entries(statistics.approvalStatus).map(([status, count]) => (
                    <Col xs={6} key={status}>
                      <div style={{
                        textAlign: 'center',
                        padding: '8px',
                        border: '1px solid #f0f0f0',
                        borderRadius: '6px',
                        minHeight: '45px',
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

          {/* 常用功能模块 */}
          <div style={{
            background: 'linear-gradient(135deg, #ffffff 0%, #f8f9fa 100%)',
            borderRadius: '8px',
            padding: '12px',
            height: '150px',
            marginBottom: '12px',
            boxShadow: '0 2px 8px rgba(0, 0, 0, 0.06)',
            border: '1px solid rgba(255, 255, 255, 0.8)'
          }}>
            <Card 
              size="small" 
              title={
                <Space>
                  <ThunderboltOutlined style={{ color: DESIGN_SYSTEM.colors.warning }} />
                  <span style={{ fontSize: '14px', fontWeight: '600' }}>常用功能</span>
                </Space>
              }
              style={{ border: 'none', background: 'transparent', height: 'calc(100% - 16px)' }}
              bodyStyle={{ padding: '8px', background: 'transparent', height: 'calc(100% - 40px)' }}
            >
              <Row gutter={[6, 6]}>
                <Col xs={8}>
                  <Button
                    type="primary"
                    icon={<FileTextOutlined />}
                    size="small"
                    onClick={() => handleQuickActionClick('new-project')}
                    style={{
                      width: '100%',
                      height: '36px',
                      display: 'flex',
                      flexDirection: 'column',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontSize: '10px',
                      padding: '4px'
                    }}
                  >
                    新建项目
                  </Button>
                </Col>
                <Col xs={8}>
                  <Button
                    icon={<CustomerServiceOutlined />}
                    size="small"
                    onClick={() => handleQuickActionClick('client-management')}
                    style={{
                      width: '100%',
                      height: '36px',
                      display: 'flex',
                      flexDirection: 'column',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontSize: '10px',
                      padding: '4px'
                    }}
                  >
                    客户管理
                  </Button>
                </Col>
                <Col xs={8}>
                  <Button
                    icon={<CalendarOutlined />}
                    size="small"
                    onClick={() => handleQuickActionClick('schedule')}
                    style={{
                      width: '100%',
                      height: '36px',
                      display: 'flex',
                      flexDirection: 'column',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontSize: '10px',
                      padding: '4px'
                    }}
                  >
                    日程安排
                  </Button>
                </Col>
                <Col xs={8}>
                  <Button
                    icon={<TeamOutlined />}
                    size="small"
                    onClick={() => handleQuickActionClick('team')}
                    style={{
                      width: '100%',
                      height: '36px',
                      display: 'flex',
                      flexDirection: 'column',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontSize: '10px',
                      padding: '4px'
                    }}
                  >
                    团队协作
                  </Button>
                </Col>
                <Col xs={8}>
                  <Button
                    icon={<DollarCircleOutlined />}
                    size="small"
                    onClick={() => handleQuickActionClick('finance')}
                    style={{
                      width: '100%',
                      height: '36px',
                      display: 'flex',
                      flexDirection: 'column',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontSize: '10px',
                      padding: '4px'
                    }}
                  >
                    财务管理
                  </Button>
                </Col>
                <Col xs={8}>
                  <Button
                    icon={<LineChartOutlined />}
                    size="small"
                    onClick={() => handleQuickActionClick('analytics')}
                    style={{
                      width: '100%',
                      height: '36px',
                      display: 'flex',
                      flexDirection: 'column',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontSize: '10px',
                      padding: '4px'
                    }}
                  >
                    数据分析
                  </Button>
                </Col>
              </Row>
            </Card>
          </div>


        </Col>

        <Col xs={24} lg={8}>
          {/* 待办事项 - 高度280px */}
          <div style={{
            background: 'linear-gradient(135deg, #ffffff 0%, #f8f9fa 100%)',
            borderRadius: '8px',
            padding: '12px',
            height: '280px',
            marginBottom: '12px',
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
              style={{ border: 'none', background: 'transparent', height: 'calc(100% - 16px)' }}
              bodyStyle={{ padding: '4px', height: 'calc(100% - 40px)', overflow: 'auto', background: 'transparent' }}
            >
              <List
                size="small"
                dataSource={todos}
                renderItem={(item) => (
                  <List.Item style={{ padding: '8px 12px', border: 'none', borderBottom: '1px solid #f0f0f0' }}>
                    <div style={{ width: '100%' }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '4px' }}>
                        {renderPriorityTag(item.priority)}
                        <Button type="link" size="small" style={{ padding: 0, fontSize: '11px', height: 'auto', fontWeight: '500' }}>
                          处理
                        </Button>
                      </div>
                      <div style={{ fontSize: '13px', fontWeight: '500', color: '#1f2937', marginBottom: '4px', lineHeight: '1.3' }}>
                        {item.title}
                      </div>
                      <div style={{ fontSize: '11px', color: '#6b7280' }}>
                        <Space split={<span style={{ color: '#d1d5db' }}>|</span>}>
                          <span><CalendarOutlined style={{ fontSize: '9px' }} /> {item.deadline}</span>
                          <span><UserOutlined style={{ fontSize: '9px' }} /> {item.assignee}</span>
                        </Space>
                      </div>
                    </div>
                  </List.Item>
                )}
              />
            </Card>
          </div>

          {/* 最新动态 - 高度250px */}
          <div style={{
            background: 'linear-gradient(135deg, #ffffff 0%, #f8f9fa 100%)',
            borderRadius: '8px',
            padding: '12px',
            height: '250px',
            boxShadow: '0 2px 8px rgba(0, 0, 0, 0.06)',
            border: '1px solid rgba(255, 255, 255, 0.8)'
          }}>
            <Card 
              size="small"
              title={
                <Space>
                  <LineChartOutlined style={{ color: DESIGN_SYSTEM.colors.primary }} />
                  <span style={{ fontSize: '14px', fontWeight: '600' }}>最新动态</span>
                </Space>
              }
              style={{ border: 'none', background: 'transparent', height: 'calc(100% - 16px)' }}
              bodyStyle={{ padding: '4px', height: 'calc(100% - 40px)', overflow: 'auto', background: 'transparent' }}
            >
              <Timeline>
                {activities.slice(0, 6).map((activity) => (
                  <Timeline.Item 
                    key={activity.id}
                    color={STATUS_COLORS[activity.status as keyof typeof STATUS_COLORS] || DESIGN_SYSTEM.colors.neutral}
                    dot={
                      <div style={{ fontSize: '10px' }}>
                        {activity.type === 'approval' && <CheckCircleOutlined />}
                        {activity.type === 'project' && <FileTextOutlined />}
                        {activity.type === 'finance' && <DollarCircleOutlined />}
                      </div>
                    }
                  >
                    <div style={{ paddingBottom: '6px' }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '2px' }}>
                        <Text style={{ fontSize: '12px', fontWeight: '500', lineHeight: '1.3' }}>
                          {activity.title}
                        </Text>
                        {renderStatusTag(activity.status)}
                      </div>
                      <div style={{ fontSize: '10px', color: '#6b7280' }}>
                        <Space>
                          <Avatar size={12} icon={<UserOutlined />} style={{ fontSize: '6px' }} />
                          <span>{activity.user}</span>
                          <span>•</span>
                          <span>{activity.createdAt}</span>
                        </Space>
                      </div>
                    </div>
                  </Timeline.Item>
                ))}
              </Timeline>
            </Card>
          </div>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;