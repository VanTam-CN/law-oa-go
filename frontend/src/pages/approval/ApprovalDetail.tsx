import React, { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router'
import {
  Card,
  Descriptions,
  Tag,
  Button,
  Space,
  Timeline,
  Steps,
  message,
  Divider,
  Modal,
  Form,
  Input,
  Select,
} from 'antd'
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  RollbackOutlined,
  EditOutlined,
  DeleteOutlined,
  ExclamationCircleOutlined,
  SyncOutlined,
  ClockCircleOutlined,
  LoadingOutlined,
} from '@ant-design/icons'
import {
  getApprovalDetail,
  cancelApproval,
  processApprovalDecision,
  submitApproval,
  resubmitApproval,
  updateApproval,
} from '@/services/approval'
import type { ApprovalDetail } from '@/services/approval'
import { getUserInfo } from '@/utils/storage'
import './ApprovalDetail.less'

const { TextArea } = Input
const { Option } = Select

type ApprovalAction = 'approve' | 'reject' | 'request_changes'

const ApprovalDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [loading, setLoading] = useState<boolean>(true)
  const [approval, setApproval] = useState<ApprovalDetail | null>(null)
  const [actionModalVisible, setActionModalVisible] = useState<boolean>(false)
  const [currentAction, setCurrentAction] = useState<ApprovalAction | null>(null)
  const [resubmitModalVisible, setResubmitModalVisible] = useState<boolean>(false)
  const [editModalVisible, setEditModalVisible] = useState<boolean>(false)
  const [form] = Form.useForm()
  const [editForm] = Form.useForm()

  const currentUserId = getUserInfo()?.id?.toString() || '1'

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
      case 'needs_revision':
        return (
          <Tag icon={<EditOutlined />} color='warning'>
            需要修改
          </Tag>
        )
      case 'resubmitted':
        return (
          <Tag icon={<SyncOutlined />} color='processing'>
            重新提交
          </Tag>
        )
      case 'submitted':
      case 'under_review':
        return <Tag color='processing'>待审批</Tag>
      case 'draft':
        return <Tag color='default'>草稿</Tag>
      case 'cancelled':
        return <Tag color='default'>已撤回</Tag>
      case 'expired':
        return <Tag color='default'>已过期</Tag>
      case 'pending':
        return <Tag color='processing'>待审批</Tag>
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

  const getActionLabel = (action: ApprovalAction) => {
    switch (action) {
      case 'approve':
        return '通过'
      case 'reject':
        return '拒绝'
      case 'request_changes':
        return '要求修改'
      default:
        return ''
    }
  }

  const getActionPlaceholder = (action: ApprovalAction) => {
    switch (action) {
      case 'approve':
        return '请输入通过理由（可选）'
      case 'reject':
        return '请输入拒绝理由'
      case 'request_changes':
        return '请说明需要修改的内容'
      default:
        return ''
    }
  }

  // 获取审批流程当前步骤
  const getApprovalCurrentStep = (records: any[], status: string) => {
    if (status === 'draft') return 0
    const completedSteps = records.filter(
      (r) => r.decision === 'approve' || r.decision === 'request_changes'
    ).length
    return completedSteps
  }

  // 获取审批流程状态
  const getApprovalStepStatus = (status: string) => {
    if (status === 'approved') return 'finish'
    if (status === 'rejected') return 'error'
    if (status === 'cancelled') return 'wait'
    return 'process'
  }

  // 获取步骤状态
  const getStepStatus = (decision: string, recordStatus: string) => {
    if (decision === 'approve') return 'finish'
    if (decision === 'reject') return 'error'
    if (decision === 'request_changes') return 'wait'
    return 'process'
  }

  // 获取步骤图标
  const getStepIcon = (decision: string, recordStatus: string) => {
    if (decision === 'approve') return <CheckCircleOutlined />
    if (decision === 'reject') return <CloseCircleOutlined />
    if (decision === 'request_changes') return <EditOutlined />
    if (recordStatus === 'pending') return <ClockCircleOutlined />
    return <LoadingOutlined />
  }

  const handleAction = (action: ApprovalAction) => {
    setCurrentAction(action)
    setActionModalVisible(true)
  }

  const submitAction = async () => {
    if (!approval || !currentAction) {
      return
    }

    try {
      const values = await form.validateFields()
      await processApprovalDecision(approval.id, {
        decision: currentAction,
        decisionReason: values.comment,
        decisionComments: values.comment,
      })
      message.success(`${getActionLabel(currentAction)}成功`)
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
      icon: <ExclamationCircleOutlined />,
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

  const handleSubmit = async () => {
    if (!approval) {
      return
    }

    Modal.confirm({
      title: '确认提交',
      content: '提交后将进入审批流程，确定要提交吗？',
      icon: <ExclamationCircleOutlined />,
      onOk: async () => {
        try {
          await submitApproval(approval.id)
          message.success('提交成功')
          fetchApprovalDetail(approval.id)
        } catch (error) {
          console.error('Failed to submit approval:', error)
          message.error('提交失败')
        }
      },
    })
  }

  const handleResubmit = async () => {
    if (!approval) {
      return
    }

    try {
      const values = await form.validateFields()
      await resubmitApproval(approval.id, values.comment)
      message.success('重新提交成功')
      setResubmitModalVisible(false)
      form.resetFields()
      fetchApprovalDetail(approval.id)
    } catch (error) {
      console.error('Failed to resubmit approval:', error)
      message.error('重新提交失败')
    }
  }

  const handleEdit = () => {
    if (!approval) return
    editForm.setFieldsValue({
      title: approval.title,
      content: approval.content,
    })
    setEditModalVisible(true)
  }

  const handleEditSubmit = async () => {
    if (!approval) {
      return
    }

    try {
      const values = await editForm.validateFields()
      await updateApproval(approval.id, {
        title: values.title,
        content: values.content,
      })
      message.success('更新成功')
      setEditModalVisible(false)
      editForm.resetFields()
      fetchApprovalDetail(approval.id)
    } catch (error) {
      console.error('Failed to update approval:', error)
      message.error('更新失败')
    }
  }

  if (loading) {
    return <div>加载中...</div>
  }

  if (!approval) {
    return <div>审批不存在</div>
  }

  // 判断权限
  const isApprover = approval.currentApproverId === currentUserId
  const isApplicant = approval.applicantId === currentUserId
  const canApprove = (approval.status === 'submitted' || approval.status === 'under_review') && isApprover
  const canCancel = (approval.status === 'submitted' || approval.status === 'draft') && isApplicant
  const canSubmit = approval.status === 'draft' && isApplicant
  const canEdit = (approval.status === 'draft' || approval.status === 'needs_revision') && isApplicant
  const canResubmit = (approval.status === 'rejected' || approval.status === 'needs_revision') && isApplicant
  const hasDecisionActions = canSubmit || canEdit || canResubmit || canCancel || canApprove

  const handleMoreApprovalActions = () => {
    message.info('暂无更多审批操作')
  }

  const handleMoreHandlingActions = () => {
    if (!hasDecisionActions) {
      message.info('当前仅可查看审批进度，暂无更多处理方式')
      return
    }

    message.info('暂无更多处理方式')
  }

  return (
    <div className='approval-detail-container'>
      <Card>
        <div className='detail-header'>
          <h2>{approval.title}</h2>
          <Space>
            {renderStatusTag(approval.status)}
            {renderUrgencyTag(approval.urgency)}
            <Button onClick={handleMoreApprovalActions}>更多审批操作</Button>
          </Space>
        </div>

        <Divider />

        <Descriptions column={2}>
          <Descriptions.Item label='申请编号'>{approval.requestNumber}</Descriptions.Item>
          <Descriptions.Item label='申请类型'>{approval.type}</Descriptions.Item>
          <Descriptions.Item label='申请部门'>{approval.department}</Descriptions.Item>
          <Descriptions.Item label='申请人'>{approval.applicant}</Descriptions.Item>
          <Descriptions.Item label='申请时间'>
            {approval.submissionDate || approval.createTime}
          </Descriptions.Item>
          {approval.currentStage && (
            <Descriptions.Item label='当前阶段'>{approval.currentStage}</Descriptions.Item>
          )}
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

        {/* 审批流程步骤条 */}
        {approval.records && approval.records.length > 0 && (
          <div className='approval-flow-steps'>
            <h3>审批流程</h3>
            <Steps
              current={getApprovalCurrentStep(approval.records, approval.status)}
              status={getApprovalStepStatus(approval.status)}
              size="small"
            >
              {approval.records.map((record, index) => (
                <Steps.Step
                  key={record.id}
                  title={record.stage || `步骤 ${index + 1}`}
                  description={
                    <div className='step-description'>
                      <div className='step-approver'>{record.approver}</div>
                      <div className='step-time'>{record.approvalDate}</div>
                    </div>
                  }
                  icon={getStepIcon(record.decision, record.status)}
                  status={getStepStatus(record.decision, record.status)}
                />
              ))}
            </Steps>
          </div>
        )}

        {approval.records && approval.records.length > 0 && (
          <div className='approval-history'>
            <h3>审批记录</h3>
            <Timeline>
              {approval.records.map((record) => (
                <Timeline.Item
                  key={record.id}
                  color={
                    record.decision === 'approve'
                      ? 'green'
                      : record.decision === 'request_changes'
                        ? 'orange'
                        : 'red'
                  }
                >
                  <div className='record-item'>
                    <div className='record-header'>
                      <span className='approver'>{record.approver}</span>
                      <span className='action'>
                        {record.decision === 'approve'
                          ? '通过'
                          : record.decision === 'request_changes'
                            ? '要求修改'
                            : '拒绝'}
                      </span>
                      <span className='time'>{record.approvalDate}</span>
                    </div>
                    <div className='record-reason'>审批理由：{record.decisionReason}</div>
                    {record.decisionComments && (
                      <div className='record-comment'>审批意见：{record.decisionComments}</div>
                    )}
                  </div>
                </Timeline.Item>
              ))}
            </Timeline>
          </div>
        )}

        <Divider />

        {/* Metadata Display Section */}
        {(() => {
          if (!approval.metadata) return null
          try {
            const metadata = JSON.parse(approval.metadata)
            const { conflict_check_config, case_creation_config } = metadata

            return (
              <div className='metadata-section'>
                {conflict_check_config && (
                  <div className='conflict-check-info' style={{ marginBottom: 24 }}>
                    <h3>利益冲突检测报告</h3>
                    <Card size='small' type='inner' title='检测配置'>
                      <Descriptions column={2} size='small'>
                        <Descriptions.Item label='委托人'>
                          {conflict_check_config.clientName}
                        </Descriptions.Item>
                        <Descriptions.Item label='对方当事人'>
                          {Array.isArray(conflict_check_config.otherParties)
                            ? conflict_check_config.otherParties.join(', ')
                            : conflict_check_config.otherParties}
                        </Descriptions.Item>
                        <Descriptions.Item label='检索深度'>
                          {conflict_check_config.searchDepth}
                        </Descriptions.Item>
                        <Descriptions.Item label='检索年份'>
                          {conflict_check_config.searchYears}年
                        </Descriptions.Item>
                      </Descriptions>
                    </Card>
                  </div>
                )}

                {case_creation_config && (
                  <div className='case-creation-info'>
                    <h3>案件创建预览</h3>
                    <Card size='small' type='inner' title='拟创建案件信息'>
                      <Descriptions column={2} size='small'>
                        <Descriptions.Item label='案件名称'>
                          {case_creation_config.caseName}
                        </Descriptions.Item>
                        <Descriptions.Item label='案件类型'>
                          {case_creation_config.caseType}
                        </Descriptions.Item>
                        <Descriptions.Item label='收费方式'>
                          {case_creation_config.billingMethod}
                        </Descriptions.Item>
                        <Descriptions.Item label='合同金额'>
                          {case_creation_config.contractAmount}
                        </Descriptions.Item>
                        <Descriptions.Item label='主办律师'>
                          {case_creation_config.leadLawyer}
                        </Descriptions.Item>
                        <Descriptions.Item label='预估工期'>
                          {case_creation_config.estimatedDuration}个月
                        </Descriptions.Item>
                      </Descriptions>
                      <div style={{ marginTop: 12 }}>
                        <strong>案件描述：</strong>
                        <div
                          style={{
                            marginTop: 8,
                            padding: 8,
                            background: '#f5f5f5',
                            borderRadius: 4,
                          }}
                        >
                          {case_creation_config.caseDescription}
                        </div>
                      </div>
                    </Card>
                  </div>
                )}
              </div>
            )
          } catch (e) {
            console.error('Failed to parse metadata:', e)
            return null
          }
        })()}

        <div className='detail-actions'>
          <Space wrap>
            <Button onClick={() => navigate('/approval')}>返回列表</Button>

            {/* 申请人操作 */}
            {canSubmit && (
              <Button type='primary' icon={<CheckCircleOutlined />} onClick={handleSubmit}>
                提交审批
              </Button>
            )}
            {canEdit && (
              <Button icon={<EditOutlined />} onClick={handleEdit}>
                编辑
              </Button>
            )}
            {canResubmit && (
              <Button type='primary' icon={<SyncOutlined />} onClick={() => setResubmitModalVisible(true)}>
                重新提交
              </Button>
            )}
            {canCancel && (
              <Button icon={<RollbackOutlined />} onClick={handleCancel}>
                撤回
              </Button>
            )}

            {/* 审批人操作 */}
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
                  icon={<EditOutlined />}
                  onClick={() => handleAction('request_changes')}
                >
                  要求修改
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

            <Button onClick={handleMoreHandlingActions}>更多处理方式</Button>
          </Space>
        </div>
      </Card>

      {/* 审批决定弹窗 */}
      <Modal
        title={`${getActionLabel(currentAction || 'approve')}审批`}
        open={actionModalVisible}
        onOk={submitAction}
        onCancel={() => {
          setActionModalVisible(false)
          form.resetFields()
        }}
        okText={getActionLabel(currentAction || 'approve')}
      >
        <Form form={form} layout='vertical'>
          <Form.Item
            name='comment'
            label='审批意见'
            rules={[{ required: true, message: '请输入审批意见' }]}
          >
            <TextArea
              rows={4}
              placeholder={getActionPlaceholder(currentAction || 'approve')}
            />
          </Form.Item>
        </Form>
      </Modal>

      {/* 重新提交弹窗 */}
      <Modal
        title='重新提交审批'
        open={resubmitModalVisible}
        onOk={handleResubmit}
        onCancel={() => {
          setResubmitModalVisible(false)
          form.resetFields()
        }}
        okText='重新提交'
      >
        <Form form={form} layout='vertical'>
          <Form.Item
            name='comment'
            label='修改说明'
            rules={[{ required: true, message: '请说明修改内容' }]}
          >
            <TextArea
              rows={4}
              placeholder='请说明已修改的内容'
            />
          </Form.Item>
        </Form>
      </Modal>

      {/* 编辑弹窗 */}
      <Modal
        title='编辑审批申请'
        open={editModalVisible}
        onOk={handleEditSubmit}
        onCancel={() => {
          setEditModalVisible(false)
          editForm.resetFields()
        }}
        okText='保存'
        width={600}
      >
        <Form form={editForm} layout='vertical'>
          <Form.Item
            name='title'
            label='申请标题'
            rules={[{ required: true, message: '请输入申请标题' }]}
          >
            <Input placeholder='请输入申请标题' />
          </Form.Item>
          <Form.Item
            name='content'
            label='申请内容'
            rules={[{ required: true, message: '请输入申请内容' }]}
          >
            <TextArea rows={6} placeholder='请输入申请内容' />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default ApprovalDetail
