import React, { useState, useEffect } from 'react'
import { Table, Card, Button, Tag, Space, Tabs, message, Badge } from 'antd'
import {
  PlusOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  SyncOutlined,
  ClockCircleOutlined,
  EyeOutlined,
} from '@ant-design/icons'
import { useNavigate } from 'react-router'
import { getApprovals, cancelApproval, getApprovalStats, ApprovalItem } from '@/services/approval'
import type { TabsProps } from 'antd'
import './ApprovalList.less'

interface ApprovalStats {
  totalRequests: number
  pendingRequests: number
  myPendingRequests: number
  approvedRequests: number
  rejectedRequests: number
}

const ApprovalList: React.FC = () => {
  const navigate = useNavigate()
  const [loading, setLoading] = useState<boolean>(true)
  const [myApprovals, setMyApprovals] = useState<ApprovalItem[]>([])
  const [pendingApprovals, setPendingApprovals] = useState<ApprovalItem[]>([])
  const [approvedApprovals, setApprovedApprovals] = useState<ApprovalItem[]>([])
  const [stats, setStats] = useState<ApprovalStats | null>(null)
  const [activeTab, setActiveTab] = useState<string>('pending')

  useEffect(() => {
    fetchApprovals()
    fetchStats()
  }, [activeTab])

  const fetchApprovals = async () => {
    try {
      setLoading(true)

      // 根据当前标签页获取数据
      if (activeTab === 'my') {
        const data = await getApprovals('my')
        setMyApprovals(data)
      } else if (activeTab === 'pending') {
        const data = await getApprovals('pending')
        setPendingApprovals(data)
      } else if (activeTab === 'approved') {
        // 获取已审批的列表（通过API或从待审批列表过滤）
        // 这里假设后端有相应的API，如果没有则从前端过滤
        const data = await getApprovals('pending')
        // 筛选出已经处理过的（状态为 approved 或 rejected）
        const processed = data.filter(
          (item) => item.status === 'approved' || item.status === 'rejected'
        )
        setApprovedApprovals(processed)
      }
    } catch (error) {
      console.error('Failed to fetch approvals:', error)
      message.error('获取审批列表失败')
    } finally {
      setLoading(false)
    }
  }

  const fetchStats = async () => {
    try {
      const data = await getApprovalStats()
      setStats(data)
    } catch (error) {
      console.error('Failed to fetch stats:', error)
    }
  }

  // 渲染状态标签
  const renderStatusTag = (status: string) => {
    switch (status) {
      case 'approved':
        return (
          <Tag icon={<CheckCircleOutlined />} color='success'>
            已通过
          </Tag>
        )
      case 'rejected':
        return (
          <Tag icon={<CloseCircleOutlined />} color='error'>
            已拒绝
          </Tag>
        )
      case 'submitted':
        return (
          <Tag icon={<ClockCircleOutlined />} color='processing'>
            已提交
          </Tag>
        )
      case 'under_review':
        return (
          <Tag icon={<SyncOutlined spin />} color='processing'>
            审核中
          </Tag>
        )
      case 'draft':
        return <Tag color='default'>草稿</Tag>
      case 'cancelled':
        return <Tag color='default'>已撤回</Tag>
      case 'expired':
        return <Tag color='warning'>已过期</Tag>
      default:
        return <Tag>未知</Tag>
    }
  }

  // 渲染紧急程度标签
  const renderUrgencyTag = (urgency: string) => {
    switch (urgency) {
      case 'very_urgent':
        return <Tag color='red'>特急</Tag>
      case 'urgent':
        return <Tag color='orange'>紧急</Tag>
      case 'normal':
        return <Tag color='blue'>普通</Tag>
      default:
        return <Tag>未知</Tag>
    }
  }

  const columns = [
    {
      title: '申请类型',
      dataIndex: 'type',
      key: 'type',
    },
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
      render: (text: string, record: ApprovalItem) => (
        <a onClick={() => navigate(`/approval/${record.id}`)}>{text}</a>
      ),
    },
    {
      title: '申请人',
      dataIndex: 'applicant',
      key: 'applicant',
    },
    {
      title: '部门',
      dataIndex: 'department',
      key: 'department',
    },
    {
      title: '申请时间',
      dataIndex: 'submission_date',
      key: 'submission_date',
      render: (text: string) => (text ? new Date(text).toLocaleString() : '-'),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => renderStatusTag(status),
    },
    {
      title: '紧急程度',
      dataIndex: 'urgency',
      key: 'urgency',
      render: (urgency: string) => renderUrgencyTag(urgency),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: ApprovalItem) => (
        <Space size='middle'>
          <a onClick={() => navigate(`/approval/${record.id}`)}>查看</a>
          {(record.status === 'submitted' || record.status === 'under_review') &&
            activeTab === 'my' && <a onClick={() => handleCancel(record.id)}>撤回</a>}
        </Space>
      ),
    },
  ]

  const handleCancel = async (id: string) => {
    try {
      await cancelApproval(id)
      message.success('审批已撤回')
      fetchApprovals()
      fetchStats()
    } catch (error) {
      console.error('Failed to cancel approval:', error)
      message.error('撤回失败')
    }
  }

  const items: TabsProps['items'] = [
    {
      key: 'pending',
      label: (
        <span>
          待我审批
          {stats && <Badge count={stats.myPendingRequests} style={{ marginLeft: 8 }} />}
        </span>
      ),
      children: (
        <Table
          rowKey='id'
          columns={columns}
          dataSource={pendingApprovals}
          loading={loading && activeTab === 'pending'}
          pagination={{ pageSize: 10 }}
          size='small'
        />
      ),
    },
    {
      key: 'my',
      label: (
        <span>
          我的申请
          {stats && <Badge count={stats.totalRequests} style={{ marginLeft: 8 }} />}
        </span>
      ),
      children: (
        <Table
          rowKey='id'
          columns={columns}
          dataSource={myApprovals}
          loading={loading && activeTab === 'my'}
          pagination={{ pageSize: 10 }}
          size='small'
        />
      ),
    },
    {
      key: 'approved',
      label: (
        <span>
          已审批
          {stats && (
            <Badge
              count={stats.approvedRequests + stats.rejectedRequests}
              style={{ marginLeft: 8 }}
            />
          )}
        </span>
      ),
      children: (
        <Table
          rowKey='id'
          columns={columns}
          dataSource={approvedApprovals}
          loading={loading && activeTab === 'approved'}
          pagination={{ pageSize: 10 }}
          size='small'
        />
      ),
    },
  ]

  return (
    <div className='approval-list-container'>
      <Card
        title='审批管理'
        extra={
          <Button
            type='primary'
            icon={<PlusOutlined />}
            onClick={() => navigate('/approval/create')}
          >
            新建审批
          </Button>
        }
      >
        <Tabs defaultActiveKey='pending' items={items} onChange={(key) => setActiveTab(key)} />
      </Card>
    </div>
  )
}

export default ApprovalList
