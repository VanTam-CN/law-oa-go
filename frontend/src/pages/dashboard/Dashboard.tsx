import React, { useState, useEffect } from 'react'
import { useNavigate } from 'react-router'
import './Dashboard.module.css'
import { dataService } from '@/services/dataService'
import { useAppStore } from '@/stores/useAppStore'
import { getApprovalStats } from '@/services/approval'
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
  message,
} from 'antd'
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
  CalendarOutlined,
  ApartmentOutlined,
  SearchOutlined,
  FolderOutlined,
  SafetyCertificateOutlined,
  BookOutlined,
} from '@ant-design/icons'

const { Title, Text } = Typography

// 设计系统配置 - 使用新的专业配色方案
const DESIGN_SYSTEM = {
  colors: {
    primary: '#1E5A8D',
    info: '#1E5A8D',
    success: '#3FAF56',
    warning: '#F5A623',
    error: '#E8484C',
    neutral: '#6B7280',
    border: '#E8EBF0',
    background: '#F4F6F8',
  },
  spacing: {
    xs: 4,
    sm: 8,
    md: 12,
    lg: 16,
    xl: 20,
    xxl: 24,
  },
  typography: {
    xs: { fontSize: '12px', lineHeight: '16px' },
    sm: { fontSize: '13px', lineHeight: '18px' },
    md: { fontSize: '14px', lineHeight: '20px' },
    lg: { fontSize: '16px', lineHeight: '24px' },
    xl: { fontSize: '18px', lineHeight: '28px' },
    xxl: { fontSize: '20px', lineHeight: '30px' },
  },
  radius: {
    sm: '6px',
    md: '8px',
    lg: '12px',
  },
}

// 状态颜色映射 - 使用新的专业配色
const STATUS_COLORS = {
  进行中: '#1E5A8D',
  已完成: '#3FAF56',
  已暂停: '#F5A623',
  已取消: '#E8484C',
  待审批: '#F5A623',
  已通过: '#3FAF56',
  已拒绝: '#E8484C',
  已撤销: '#9CA3AF',
}

// 优先级映射
const priorityMap: Record<string, string> = {
  high: '高',
  medium: '中',
  low: '低',
}

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
    totalExpenses: 0,
  },
}

// MVP 快速操作配置
export const MVP_DASHBOARD_QUICK_ACTIONS = [
  { key: 'case-management', label: '案件管理', path: '/case' },
  { key: 'case-create', label: '新建立案', path: '/case/create' },
  { key: 'conflict-check', label: '冲突检测', path: '/conflict' },
  { key: 'client-management', label: '客户管理', path: '/client' },
  { key: 'approval-management', label: '审批管理', path: '/approval' },
  { key: 'trust-management', label: '信托账户', path: '/trust' },
] as const

interface DashboardProps {
  statistics?: any
  todos?: any[]
  activities?: any[]
  user?: any
  kpis?: any
  loading?: boolean
}

const Dashboard: React.FC<DashboardProps> = () => {
  // 导航功能
  const navigate = useNavigate()

  // 全局状态管理
  const { user: currentUser } = useAppStore()

  // 状态管理
  const [statistics, setStatistics] = useState<any>(null)
  const [todos, setTodos] = useState<any[]>([])
  const [activities, setActivities] = useState<any[]>([])
  const [user, setUser] = useState<any>(null)
  const [kpis, setKpis] = useState<any>(null)
  const [loading, setLoading] = useState(true)

  // 获取仪表盘数据
  useEffect(() => {
    const fetchDashboardData = async () => {
      try {
        setLoading(true)

        // 并行获取所有数据
        const [statsData, todosData, activitiesData, approvalStatsData] = await Promise.all([
          dataService.getDashboardStatistics(),
          dataService.getDashboardTodos(),
          dataService.getDashboardActivities(),
          getApprovalStats().catch((err) => {
            console.warn('获取审批统计失败，使用空统计:', err)
            return {
              pendingRequests: 0,
              totalRequests: 0,
              approvedRequests: 0,
              rejectedRequests: 0,
            }
          }),
        ])

        // 使用全局状态的用户信息
        const userData = currentUser || { real_name: '用户', avatar: null }

        setStatistics(statsData)
        setTodos(todosData)
        setActivities(activitiesData)
        setUser(userData)

        // 设置审批统计数据
        if (approvalStatsData) {
          setStatistics((prev: any) => ({
            ...prev,
            pendingApprovals: approvalStatsData.pendingRequests || 0,
            approvalStatus: {
              待审批: approvalStatsData.pendingRequests || 0,
              已通过: approvalStatsData.approvedRequests || 0,
              已拒绝: approvalStatsData.rejectedRequests || 0,
              已撤销: 0,
            },
          }))
        }

        // 计算KPI数据
        if (statsData) {
          const netProfit =
            statsData.financeStats?.totalRevenue - statsData.financeStats?.totalExpenses || 0
          const profitMargin =
            statsData.financeStats?.totalRevenue > 0
              ? ((netProfit / statsData.financeStats?.totalRevenue) * 100).toFixed(1)
              : 0
          const completionRate =
            statsData.totalProjects > 0
              ? ((statsData.completedProjects / statsData.totalProjects) * 100).toFixed(1)
              : 0

          setKpis({
            netProfit,
            profitMargin: Number(profitMargin),
            completionRate: Number(completionRate),
          })
        }
      } catch (error) {
        console.error('获取仪表盘数据失败:', error)
        // 显示用户友好的错误信息
        message.error('加载仪表盘数据失败，请稍后重试')

        // 发生错误时使用空数据作为降级方案
        setStatistics(initialStatistics)
        setTodos([])
        setActivities([])
        setUser({ real_name: '用户', avatar: null })
        setKpis({
          netProfit: 0,
          profitMargin: 0,
          completionRate: 0,
        })
      } finally {
        setLoading(false)
      }
    }

    fetchDashboardData()
  }, [])
  // 渲染优先级标签
  const renderPriorityTag = (priority: string) => {
    const color = priority === 'high' ? 'red' : priority === 'medium' ? 'orange' : 'green'
    return (
      <Tag
        color={color}
        style={{
          fontSize: DESIGN_SYSTEM.typography.xs.fontSize,
          fontWeight: 500,
          borderRadius: DESIGN_SYSTEM.radius.sm,
        }}
      >
        {priorityMap[priority] || priority}
      </Tag>
    )
  }

  // 渲染状态标签
  const renderStatusTag = (status: string) => {
    const color =
      STATUS_COLORS[status as keyof typeof STATUS_COLORS] || DESIGN_SYSTEM.colors.neutral
    return (
      <Tag
        color={color}
        style={{
          fontSize: DESIGN_SYSTEM.typography.xs.fontSize,
          fontWeight: 500,
          borderRadius: DESIGN_SYSTEM.radius.sm,
        }}
      >
        {status}
      </Tag>
    )
  }

  // 项目状态数据转换为图表格式
  const projectStatusData = statistics?.projectStatus
    ? Object.entries(statistics.projectStatus).map(([name, value]: [string, any]) => ({
        name,
        value: Number(value) || 0,
        color: STATUS_COLORS[name as keyof typeof STATUS_COLORS] || DESIGN_SYSTEM.colors.neutral,
        percentage:
          statistics.totalProjects > 0
            ? ((Number(value) / statistics.totalProjects) * 100).toFixed(1)
            : '0',
      }))
    : []

  // 处理状态项点击跳转
  const handleStatusItemClick = (status: string) => {
    console.log(`跳转到${status}项目列表`)
  }

  // 动态欢迎语生成
  const getDynamicWelcome = () => {
    const hour = new Date().getHours()
    const userName = user?.real_name || '用户'
    const urgentTodos = todos.filter(
      (todo) =>
        !todo.completed &&
        todo.priority === 'high' &&
        todo.deadline &&
        new Date(todo.deadline) < new Date(),
    ).length
    const pendingApprovals = 0

    let timeGreeting = ''
    if (hour < 12) {
      timeGreeting = '早上好'
    } else if (hour < 18) {
      timeGreeting = '下午好'
    } else {
      timeGreeting = '晚上好'
    }

    let guidance = ''
    if (urgentTodos > 0 || pendingApprovals > 0) {
      guidance = `您今天有 ${urgentTodos} 个紧急事项和 ${pendingApprovals} 个待审批需要处理。`
    } else if (todos.filter((todo) => !todo.completed).length > 0) {
      guidance = `您今天有 ${todos.filter((todo) => !todo.completed).length} 个待办事项需要处理。`
    } else {
      guidance = '今天没有紧急事项，祝您工作愉快！'
    }

    return {
      greeting: `${timeGreeting}，${userName}`,
      guidance,
    }
  }

  // KPI数据增强
  const getEnhancedKPIs = () => {
    if (!statistics) {
      return null
    }

    const totalRevenue = statistics.financeStats?.totalRevenue || 0
    const totalExpenses = statistics.financeStats?.totalExpenses || 0
    const netProfit = totalRevenue - totalExpenses
    const profitMargin = totalRevenue > 0 ? (netProfit / totalRevenue) * 100 : 0

    const completionRate =
      statistics.totalProjects > 0
        ? ((statistics.projectStatus?.['已完成'] || 0) / statistics.totalProjects) * 100
        : 0

    const profitTrend = netProfit > 100000 ? '+5.2%' : '-2.1%'
    const completionTrend = completionRate > 40 ? '+3.8%' : '-1.2%'

    return {
      netProfit: {
        value: netProfit,
        trend: profitTrend,
        trendUp: profitTrend.startsWith('+'),
        period: '本月',
      },
      completionRate: {
        value: completionRate,
        trend: completionTrend,
        trendUp: completionTrend.startsWith('+'),
        period: '年度',
      },
      pendingApprovals: statistics.pendingApprovals || 0,
      activeClients: statistics.activeClients || 2,
    }
  }

  // 处理KPI卡片点击
  const handleKPIClick = (kpiType: string) => {
    console.log(`跳转到${kpiType}详情页面`)
  }

  // 处理常用功能按钮点击
  const handleQuickActionClick = (action: string) => {
    const quickAction = MVP_DASHBOARD_QUICK_ACTIONS.find((item) => item.key === action)
    if (quickAction) {
      navigate(quickAction.path)
      return
    }
    console.log(`未知功能: ${action}`)
  }

  // 骨架屏组件
  const renderSkeleton = () => (
    <div
      style={{
        padding: '12px',
        backgroundColor: '#f8f9fa',
        display: 'flex',
        flexDirection: 'column',
        gap: '16px',
      }}
    >
      <div
        style={{
          height: '120px',
          background: 'linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%)',
          borderRadius: '8px',
          animation: 'pulse 2s infinite',
        }}
      />
      <div
        style={{
          height: '300px',
          background: 'linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%)',
          borderRadius: '8px',
          animation: 'pulse 2s infinite',
        }}
      />
      <div
        style={{
          height: '200px',
          background: 'linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%)',
          borderRadius: '8px',
          animation: 'pulse 2s infinite',
        }}
      />
    </div>
  )

  const enhancedKPIs = getEnhancedKPIs()
  const dynamicWelcome = getDynamicWelcome()

  if (loading) {
    return renderSkeleton()
  }

  return (
    <div
      className='dashboard-balanced-professional'
      style={{
        padding: '16px',
        backgroundColor: 'transparent',
      }}
    >
      {/* 顶部欢迎区域 */}
      <div
        style={{
          background: 'linear-gradient(135deg, #1E5A8D 0%, #163C5D 100%)',
          borderRadius: '12px',
          padding: '16px 20px',
          marginBottom: '16px',
          boxShadow: '0 4px 12px rgba(30, 90, 141, 0.2)',
        }}
      >
        <Row align='middle' gutter={[12, 8]}>
          {/* 左侧：用户信息 */}
          <Col xs={24} lg={10}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
              <Avatar
                size={48}
                icon={<UserOutlined />}
                style={{
                  background: 'rgba(255, 255, 255, 0.2)',
                  border: '2px solid rgba(255, 255, 255, 0.3)',
                  flexShrink: 0,
                }}
              />
              <div style={{ flex: 1, minWidth: 0 }}>
                <Title
                  level={4}
                  style={{
                    margin: 0,
                    fontSize: '18px',
                    lineHeight: '26px',
                    color: '#FFFFFF',
                    fontWeight: '600',
                    whiteSpace: 'nowrap',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                  }}
                >
                  {dynamicWelcome.greeting}
                </Title>
                <Text
                  style={{
                    fontSize: '15px',
                    color: 'rgba(255, 255, 255, 0.85)',
                    display: 'block',
                  }}
                >
                  {new Date().toLocaleDateString('zh-CN', {
                    month: 'short',
                    day: 'numeric',
                    weekday: 'short',
                  })}
                </Text>

                {/* 提示框 */}
                <div
                  style={{
                    background: 'rgba(15, 35, 57, 0.5)',
                    border: '1px solid rgba(197, 165, 114, 0.4)',
                    borderRadius: '8px',
                    padding: '6px 16px',
                    fontSize: '14px',
                    color: '#FFFFFF',
                    display: 'inline-flex',
                    alignItems: 'center',
                    gap: '6px',
                    marginTop: '8px',
                  }}
                >
                  <Space size={10} split={<Divider type="vertical" style={{ margin: 0, borderColor: 'rgba(197, 165, 114, 0.4)', height: '14px' }} />}>
                    <span style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                      <AlertOutlined style={{ fontSize: '15px', color: '#C5A572' }} />
                      待办 <strong style={{ fontSize: '16px', fontWeight: '700', color: '#C5A572' }}>{todos.filter((todo) => !todo.completed).length}</strong>
                    </span>
                    <span style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                      <ClockCircleOutlined style={{ fontSize: '15px', color: '#E8484C' }} />
                      紧急 <strong style={{ fontSize: '16px', fontWeight: '700', color: '#E8484C' }}>{todos.filter((todo) => todo.priority === 'high' && !todo.completed).length}</strong>
                    </span>
                  </Space>
                </div>
              </div>
            </div>
          </Col>

          {/* 右侧：常用功能快捷入口 */}
          <Col xs={24} lg={14}>
            <Row gutter={[8, 8]}>
              {MVP_DASHBOARD_QUICK_ACTIONS.map((action) => (
                <Col xs={6} sm={6} lg={6} key={action.key}>
                  <div
                    onClick={() => handleQuickActionClick(action.key)}
                    style={{
                      background: 'rgba(255, 255, 255, 0.15)',
                      border: '1px solid rgba(255, 255, 255, 0.25)',
                      borderRadius: '8px',
                      padding: '10px 6px',
                      cursor: 'pointer',
                      transition: 'all 0.2s ease',
                      display: 'flex',
                      flexDirection: 'column',
                      alignItems: 'center',
                      gap: '4px',
                      minHeight: '56px',
                      justifyContent: 'center',
                    }}
                    onMouseEnter={(e) => {
                      e.currentTarget.style.background = 'rgba(255, 255, 255, 0.25)'
                      e.currentTarget.style.borderColor = 'rgba(255, 255, 255, 0.4)'
                      e.currentTarget.style.transform = 'translateY(-2px)'
                    }}
                    onMouseLeave={(e) => {
                      e.currentTarget.style.background = 'rgba(255, 255, 255, 0.15)'
                      e.currentTarget.style.borderColor = 'rgba(255, 255, 255, 0.25)'
                      e.currentTarget.style.transform = 'translateY(0)'
                    }}
                  >
                    {action.key === 'case-management' && <FileTextOutlined style={{ fontSize: '18px', color: '#C5A572' }} />}
                    {action.key === 'case-create' && <ThunderboltOutlined style={{ fontSize: '18px', color: '#F59E0B' }} />}
                    {action.key === 'conflict-check' && <SafetyCertificateOutlined style={{ fontSize: '18px', color: '#E8484C' }} />}
                    {action.key === 'client-management' && <CustomerServiceOutlined style={{ fontSize: '18px', color: '#3FAF56' }} />}
                    {action.key === 'approval-management' && <CheckCircleOutlined style={{ fontSize: '18px', color: '#8B5CF6' }} />}
                    {action.key === 'trust-management' && <DollarCircleOutlined style={{ fontSize: '18px', color: '#0EA5E9' }} />}
                    <span style={{ fontSize: '13px', color: '#FFFFFF', fontWeight: '500' }}>{action.label}</span>
                  </div>
                </Col>
              ))}
            </Row>
          </Col>
        </Row>
      </div>

      {/* 核心KPI指标 */}
      <div
        style={{
          background: '#FFFFFF',
          borderRadius: '12px',
          padding: '16px 20px',
          marginBottom: '16px',
          boxShadow: '0 2px 8px rgba(0, 0, 0, 0.08)',
          border: '1px solid #E8EBF0',
        }}
      >
        <Row gutter={[12, 12]}>
          <Col xs={12} sm={12} lg={6}>
            <Card
              hoverable
              onClick={() => handleKPIClick('profit')}
              style={{
                cursor: 'pointer',
                transition: 'all 0.2s ease',
                borderLeft: '4px solid #3FAF56',
                borderRadius: '8px',
                border: '1px solid #E8EBF0',
                borderLeftWidth: '4px',
              }}
              bodyStyle={{
                padding: '16px 20px',
              }}
            >
              <Statistic
                title={<span style={{ fontSize: '15px', color: '#374151', fontWeight: '600' }}>净利润</span>}
                value={kpis?.netProfit || 0}
                suffix='元'
                valueStyle={{ color: '#3FAF56', fontSize: '24px', fontWeight: '600' }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={12} lg={6}>
            <Card
              hoverable
              onClick={() => handleKPIClick('completion')}
              style={{
                cursor: 'pointer',
                transition: 'all 0.2s ease',
                borderLeft: '4px solid #1E5A8D',
                borderRadius: '8px',
                border: '1px solid #E8EBF0',
                borderLeftWidth: '4px',
              }}
              bodyStyle={{
                padding: '16px 20px',
              }}
            >
              <Statistic
                title={<span style={{ fontSize: '15px', color: '#374151', fontWeight: '600' }}>完成率</span>}
                value={kpis?.completionRate || 0}
                suffix='%'
                valueStyle={{ color: '#1E5A8D', fontSize: '24px', fontWeight: '600' }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={12} lg={6}>
            <Card
              hoverable
              onClick={() => handleKPIClick('approval')}
              style={{
                cursor: 'pointer',
                transition: 'all 0.2s ease',
                borderLeft: '4px solid #F5A623',
                borderRadius: '8px',
                border: '1px solid #E8EBF0',
                borderLeftWidth: '4px',
              }}
              bodyStyle={{
                padding: '16px 20px',
              }}
            >
              <Statistic
                title={<span style={{ fontSize: '15px', color: '#374151', fontWeight: '600' }}>待审批</span>}
                value={statistics?.pendingApprovals || 0}
                valueStyle={{ color: '#F5A623', fontSize: '24px', fontWeight: '600' }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={12} lg={6}>
            <Card
              hoverable
              onClick={() => handleKPIClick('client')}
              style={{
                cursor: 'pointer',
                transition: 'all 0.2s ease',
                borderLeft: '4px solid #E8484C',
                borderRadius: '8px',
                border: '1px solid #E8EBF0',
                borderLeftWidth: '4px',
              }}
              bodyStyle={{
                padding: '16px 20px',
              }}
            >
              <Statistic
                title={<span style={{ fontSize: '15px', color: '#374151', fontWeight: '600' }}>活跃客户</span>}
                value={statistics?.activeClients || 0}
                valueStyle={{ color: '#E8484C', fontSize: '24px', fontWeight: '600' }}
              />
            </Card>
          </Col>
        </Row>
      </div>

      {/* 主要内容区域 - 16:8 布局 */}
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={16}>
          {/* 审批状态概览 */}
          <div
            style={{
              background: '#FFFFFF',
              borderRadius: '12px',
              padding: '16px 20px',
              marginBottom: '0px',
              boxShadow: '0 2px 8px rgba(0, 0, 0, 0.08)',
              border: '1px solid #E8EBF0',
            }}
          >
            <Card
              size='small'
              title={
                <Space size={6}>
                  <CheckCircleOutlined style={{ color: DESIGN_SYSTEM.colors.success, fontSize: '15px' }} />
                  <span style={{ fontSize: '15px', fontWeight: '600' }}>审批状态</span>
                  <Badge
                    count={Object.values(statistics?.approvalStatus || {}).reduce(
                      (sum: number, count: any) => sum + (Number(count) || 0),
                      0,
                    )}
                    style={{ backgroundColor: DESIGN_SYSTEM.colors.success, fontSize: '12px' }}
                  />
                </Space>
              }
              style={{ border: 'none', background: 'transparent' }}
              bodyStyle={{ padding: '12px 0 0 0', background: 'transparent' }}
            >
              <Row gutter={[12, 12]}>
                {statistics?.approvalStatus ? (
                  Object.entries(statistics.approvalStatus).map(([status, count]: [string, any]) => (
                    <Col xs={12} sm={8} md={6} key={status}>
                      <div
                        style={{
                          textAlign: 'center',
                          padding: '12px 8px',
                          border: '1px solid #E8EBF0',
                          borderRadius: '8px',
                          minHeight: '64px',
                          display: 'flex',
                          flexDirection: 'column',
                          justifyContent: 'center',
                          cursor: 'pointer',
                          transition: 'all 0.2s',
                        }}
                        onMouseEnter={(e) => {
                          e.currentTarget.style.borderColor = STATUS_COLORS[status as keyof typeof STATUS_COLORS]
                          e.currentTarget.style.transform = 'translateY(-3px)'
                          e.currentTarget.style.boxShadow = '0 4px 12px rgba(0,0,0,0.15)'
                        }}
                        onMouseLeave={(e) => {
                          e.currentTarget.style.borderColor = '#E8EBF0'
                          e.currentTarget.style.transform = 'translateY(0)'
                          e.currentTarget.style.boxShadow = 'none'
                        }}
                      >
                        <div
                          style={{
                            fontSize: '20px',
                            fontWeight: 'bold',
                            color: STATUS_COLORS[status as keyof typeof STATUS_COLORS] || '#374151',
                            lineHeight: 1,
                          }}
                        >
                          {Number(count) || 0}
                        </div>
                        <div style={{ fontSize: '14px', color: '#374151', marginTop: '4px', fontWeight: '500' }}>
                          {status}
                        </div>
                      </div>
                    </Col>
                  ))
                ) : (
                  <Col span={24}>
                    <div style={{ textAlign: 'center', color: '#9CA3AF', fontSize: '14px', padding: '20px' }}>
                      暂无审批数据
                    </div>
                  </Col>
                )}
              </Row>
            </Card>
          </div>

        </Col>

        <Col xs={24} lg={8}>
          {/* 待办事项 */}
          <div
            style={{
              background: '#FFFFFF',
              borderRadius: '12px',
              padding: '16px 20px',
              marginBottom: '0px',
              boxShadow: '0 2px 8px rgba(0, 0, 0, 0.08)',
              border: '1px solid #E8EBF0',
            }}
          >
            <Card
              size='small'
              title={
                <Space size={6}>
                  <AlertOutlined style={{ color: DESIGN_SYSTEM.colors.warning, fontSize: '15px' }} />
                  <span style={{ fontSize: '15px', fontWeight: '600' }}>待办事项</span>
                  <Badge
                    count={todos.length}
                    style={{ backgroundColor: DESIGN_SYSTEM.colors.warning, fontSize: '12px' }}
                  />
                </Space>
              }
              style={{ border: 'none', background: 'transparent' }}
              bodyStyle={{
                padding: '12px 0 0 0',
                background: 'transparent',
                maxHeight: '280px',
                overflow: 'auto',
              }}
            >
              <List
                size='small'
                dataSource={todos}
                renderItem={(item) => (
                  <List.Item
                    style={{
                      padding: '12px 16px',
                      border: 'none',
                      borderBottom: '1px solid #F0F0F0',
                    }}
                  >
                    <div style={{ width: '100%' }}>
                      <div
                        style={{
                          display: 'flex',
                          justifyContent: 'space-between',
                          alignItems: 'flex-start',
                          marginBottom: '6px',
                        }}
                      >
                        {renderPriorityTag(item.priority)}
                        <Button
                          type='link'
                          size='small'
                          style={{
                            padding: 0,
                            fontSize: '13px',
                            height: 'auto',
                            fontWeight: '500',
                          }}
                        >
                          处理
                        </Button>
                      </div>
                      <div
                        style={{
                          fontSize: '14px',
                          fontWeight: '500',
                          color: '#374151',
                        }}
                      >
                        {item.title}
                      </div>
                    </div>
                  </List.Item>
                )}
              />
            </Card>
          </div>

          {/* 最新动态 */}
          <div
            style={{
              background: '#FFFFFF',
              borderRadius: '12px',
              padding: '16px 20px',
              marginTop: '16px',
              boxShadow: '0 2px 8px rgba(0, 0, 0, 0.08)',
              border: '1px solid #E8EBF0',
            }}
          >
            <Card
              size='small'
              title={
                <Space size={6}>
                  <LineChartOutlined style={{ color: DESIGN_SYSTEM.colors.primary, fontSize: '15px' }} />
                  <span style={{ fontSize: '15px', fontWeight: '600' }}>最新动态</span>
                </Space>
              }
              style={{ border: 'none', background: 'transparent' }}
              bodyStyle={{
                padding: '12px 0 0 0',
                background: 'transparent',
                maxHeight: '240px',
                overflow: 'auto',
              }}
            >
              <Timeline>
                {activities.slice(0, 5).map((activity) => (
                  <Timeline.Item
                    key={activity.id}
                    color={
                      STATUS_COLORS[activity.status as keyof typeof STATUS_COLORS] ||
                      DESIGN_SYSTEM.colors.neutral
                    }
                    dot={<div style={{ fontSize: '12px' }}>{
                      activity.type === 'approval' && <CheckCircleOutlined />
                    }
                    {
                      activity.type === 'project' && <FileTextOutlined />
                    }
                    {
                      activity.type === 'finance' && <DollarCircleOutlined />
                    }</div>}
                  >
                    <div style={{ paddingBottom: '8px' }}>
                      <div
                        style={{
                          display: 'flex',
                          justifyContent: 'space-between',
                          alignItems: 'flex-start',
                          marginBottom: '6px',
                        }}
                      >
                        <Text style={{ fontSize: '14px', fontWeight: '500', color: '#374151' }}>
                          {activity.title}
                        </Text>
                        {renderStatusTag(activity.status)}
                      </div>
                      <div style={{ fontSize: '13px', color: '#374151' }}>
                        <Space size={4}>
                          <Avatar size={18} icon={<UserOutlined />} style={{ fontSize: '10px' }} />
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
  )
}

export default Dashboard
