import React, { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router'
import {
  Card,
  Descriptions,
  Tag,
  Button,
  Space,
  Timeline,
  message,
  Divider,
  Modal,
  Form,
  Input,
} from 'antd'
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  RollbackOutlined,
  EditOutlined,
  DeleteOutlined,
} from '@ant-design/icons'
import { getApprovalDetail, handleApproval, cancelApproval } from '@/services/approval'
import type { ApprovalDetail } from '@/services/approval'
import './ApprovalDetail.less'

const { TextArea } = Input

const ApprovalDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [loading, setLoading] = useState<boolean>(true)
  const [approval, setApproval] = useState<ApprovalDetail | null>(null)
  const [actionModalVisible, setActionModalVisible] = useState<boolean>(false)
  const [currentAction, setCurrentAction] = useState<'approve' | 'reject' | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    if (id) {
      fetchApprovalDetail(id)
    }
  }, [id])

  const fetchApprovalDetail = async (approvalId: string) => {
    try {
      setLoading(true)
      const data = await getApprovalDetail(approvalId)
      setApproval(data)
    } catch (error) {
      console.error('Failed to fetch approval detail:', error)
      message.error('获取审批详情失败')
    } finally {
      setLoading(false)
    }
  }

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
      case 'pending':
        return <Tag color='processing'>待审批</Tag>
      case 'cancelled':
        return <Tag color='default'>已撤回</Tag>
      default:
        return <Tag>未知</Tag>
    }
  }

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

  const handleAction = (action: 'approve' | 'reject') => {
    setCurrentAction(action)
    setActionModalVisible(true)
  }

  const submitAction = async () => {
    if (!approval || !currentAction) {
      return
    }

    try {
      const values = await form.validateFields()
      await handleApproval(approval.id, currentAction, values.comment)
      message.success(currentAction === 'approve' ? '审批通过' : '已拒绝')
      setActionModalVisible(false)
      form.resetFields()
      fetchApprovalDetail(approval.id)
    } catch (error) {
      console.error('Failed to handle approval:', error)
      message.error('操作失败')
    }
  }

  const handleCancel = async () => {
    if (!approval) {
      return
    }

    Modal.confirm({
      title: '确认撤回',
      content: '确定要撤回这个审批申请吗？',
      onOk: async () => {
        try {
          await cancelApproval(approval.id)
          message.success('撤回成功')
          fetchApprovalDetail(approval.id)
        } catch (error) {
          console.error('Failed to cancel approval:', error)
          message.error('撤回失败')
        }
      },
    })
  }

  if (loading) {
    return <div>加载中...</div>
  }

  if (!approval) {
    return <div>审批不存在</div>
  }

  const canApprove = approval.status === 'pending' && approval.currentApproverId === 1 // 假设当前用户ID为1
  const canCancel = approval.status === 'pending' && approval.applicantId === 1 // 假设当前用户ID为1

  return (
    <div className='approval-detail-container'>
      <Card>
        <div className='detail-header'>
          <h2>{approval.title}</h2>
          <Space>
            {renderStatusTag(approval.status)}
            {renderUrgencyTag(approval.urgency)}
          </Space>
        </div>

        <Divider />

        <Descriptions column={2}>
          <Descriptions.Item label='申请类型'>{approval.type}</Descriptions.Item>
          <Descriptions.Item label='申请部门'>{approval.department}</Descriptions.Item>
          <Descriptions.Item label='申请人'>{approval.applicant}</Descriptions.Item>
          <Descriptions.Item label='申请时间'>{approval.createTime}</Descriptions.Item>
          {approval.currentApprover && (
            <Descriptions.Item label='当前审批人'>{approval.currentApprover}</Descriptions.Item>
          )}
        </Descriptions>

        <Divider />

        <div className='approval-content'>
          <h3>申请内容</h3>
          <div className='content-text'>{approval.content}</div>
        </div>

        <Divider />

        {approval.records && approval.records.length > 0 && (
          <div className='approval-history'>
            <h3>审批记录</h3>
            <Timeline>
              {approval.records.map((record) => (
                <Timeline.Item
                  key={record.id}
                  color={record.action === 'approve' ? 'green' : 'red'}
                >
                  <div className='record-item'>
                    <div className='record-header'>
                      <span className='approver'>{record.approver}</span>
                      <span className='action'>
                        {record.action === 'approve' ? '通过' : '拒绝'}
                      </span>
                      <span className='time'>{record.createTime}</span>
                    </div>
                    {record.comment && (
                      <div className='record-comment'>审批意见：{record.comment}</div>
                    )}
                  </div>
                </Timeline.Item>
              ))}
            </Timeline>
          </div>
        )}

        <div className='detail-actions'>
          <Space>
            <Button onClick={() => navigate('/approval')}>返回列表</Button>
            {canApprove && (
              <>
                <Button
                  type='primary'
                  icon={<CheckCircleOutlined />}
                  onClick={() => handleAction('approve')}
                >
                  通过
                </Button>
                <Button
                  danger
                  icon={<CloseCircleOutlined />}
                  onClick={() => handleAction('reject')}
                >
                  拒绝
                </Button>
              </>
            )}
            {canCancel && (
              <Button icon={<RollbackOutlined />} onClick={handleCancel}>
                撤回
              </Button>
            )}
          </Space>
        </div>
      </Card>

      <Modal
        title={currentAction === 'approve' ? '审批通过' : '审批拒绝'}
        open={actionModalVisible}
        onOk={submitAction}
        onCancel={() => {
          setActionModalVisible(false)
          form.resetFields()
        }}
        okText={currentAction === 'approve' ? '通过' : '拒绝'}
      >
        <Form form={form} layout='vertical'>
          <Form.Item
            name='comment'
            label='审批意见'
            rules={[{ required: true, message: '请输入审批意见' }]}
          >
            <TextArea
              rows={4}
              placeholder={currentAction === 'approve' ? '请输入通过理由' : '请输入拒绝理由'}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default ApprovalDetail
