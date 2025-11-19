import React, { useState, useEffect, useCallback } from 'react';
import {
  Card,
  Table,
  Button,
  Modal,
  Form,
  Input,
  Select,
  Row,
  Col,
  Space,
  Tag,
  Badge,
  Tooltip,
  Typography,
  Alert,
  Timeline,
  Descriptions,
  Divider,
  Tabs,
  Upload,
  message,
  Popconfirm,
  Progress,
  Statistic,
  List,
  Avatar,
  Steps,
  Drawer,
  Switch,
  DatePicker,
} from 'antd';
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  ExclamationCircleOutlined,
  UploadOutlined,
  DownloadOutlined,
  EyeOutlined,
  ClockCircleOutlined,
  UserOutlined,
  FileTextOutlined,
  WarningOutlined,
  InfoCircleOutlined,
  SendOutlined,
  HistoryOutlined,
  SettingOutlined,
  SearchOutlined,
  FilterOutlined,
  SignatureOutlined,
  AuditOutlined,
  TeamOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import type { UploadProps } from 'antd';
import dayjs from 'dayjs';
import { waiverApprovalService } from '@/services/waiverApproval';
import type {
  WaiverApplication,
  WaiverStatus,
  WaiverType,
  RiskLevel,
  ApproveWaiverApplicationRequest,
  RejectWaiverApplicationRequest,
  EscalateWaiverApplicationRequest,
  WaiverApprovalRecord,
  Stakeholder,
  WaiverAttachment,
  WaiverStatistics,
  PaginatedResponse,
} from '@/types/waiverApproval';

const { TextArea } = Input;
const { Option } = Select;
const { Title, Text, Paragraph } = Typography;
const { Step } = Steps;
const { TabPane } = Tabs;

interface WaiverApprovalInterfaceProps {
  userRole?: string;
  userId?: string;
  defaultView?: 'pending' | 'history' | 'statistics';
}

// 状态配置
const statusConfig: Record<WaiverStatus, {
  label: string;
  color: string;
  icon: React.ReactNode;
  description: string;
}> = {
  DRAFT: {
    label: '草稿',
    color: 'default',
    icon: <FileTextOutlined />,
    description: '申请尚未提交',
  },
  SUBMITTED: {
    label: '已提交',
    color: 'processing',
    icon: <SendOutlined />,
    description: '申请已提交，等待审查',
  },
  UNDER_REVIEW: {
    label: '审查中',
    color: 'warning',
    icon: <ClockCircleOutlined />,
    description: '正在审查中',
  },
  APPROVED: {
    label: '已批准',
    color: 'success',
    icon: <CheckCircleOutlined />,
    description: '申请已批准',
  },
  REJECTED: {
    label: '已拒绝',
    icon: <CloseCircleOutlined />,
    color: 'error',
    description: '申请已拒绝',
  },
  ESCALATED: {
    label: '已上报',
    color: 'purple',
    icon: <ExclamationCircleOutlined />,
    description: '已上报给上级',
  },
  EXPIRED: {
    label: '已过期',
    color: 'default',
    icon: <ClockCircleOutlined />,
    description: '申请已过期',
  },
  CANCELLED: {
    label: '已取消',
    color: 'default',
    icon: <CloseCircleOutlined />,
    description: '申请已取消',
  },
};

// 风险等级配置
const riskLevelConfig: Record<RiskLevel, {
  label: string;
  color: string;
  priority: number;
}> = {
  LOW: { label: '低风险', color: 'green', priority: 1 },
  MEDIUM: { label: '中风险', color: 'orange', priority: 2 },
  HIGH: { label: '高风险', color: 'red', priority: 3 },
  CRITICAL: { label: '关键风险', color: 'purple', priority: 4 },
};

const WaiverApprovalInterface: React.FC<WaiverApprovalInterfaceProps> = ({
  userRole = 'LAWYER',
  userId,
  defaultView = 'pending',
}) => {
  const [activeTab, setActiveTab] = useState(defaultView);
  const [pendingApplications, setPendingApplications] = useState<WaiverApplication[]>([]);
  const [historyApplications, setHistoryApplications] = useState<WaiverApplication[]>([]);
  const [statistics, setStatistics] = useState<WaiverStatistics | null>(null);
  const [loading, setLoading] = useState(false);
  const [selectedApplication, setSelectedApplication] = useState<WaiverApplication | null>(null);
  const [approvalModalVisible, setApprovalModalVisible] = useState(false);
  const [rejectionModalVisible, setRejectionModalVisible] = useState(false);
  const [escalationModalVisible, setEscalationModalVisible] = useState(false);
  const [detailDrawerVisible, setDetailDrawerVisible] = useState(false);
  const [signatureData, setSignatureData] = useState<string>('');
  const [approvalForm] = Form.useForm();
  const [rejectionForm] = Form.useForm();
  const [escalationForm] = Form.useForm();

  // 加载待审批申请
  const loadPendingApplications = useCallback(async () => {
    try {
      setLoading(true);
      const response = await waiverApprovalService.getPendingApplications({
        page: 1,
        pageSize: 100,
      });
      if (response.success) {
        setPendingApplications(response.data.items);
      }
    } catch (error) {
      message.error('加载待审批申请失败');
    } finally {
      setLoading(false);
    }
  }, []);

  // 加载审批历史
  const loadHistoryApplications = useCallback(async () => {
    try {
      setLoading(true);
      const response = await waiverApprovalService.getMyApprovalHistory({
        page: 1,
        pageSize: 100,
      });
      if (response.success) {
        setHistoryApplications(response.data.items.map(item => ({
          id: item.applicationId,
          caseTitle: item.applicationTitle,
          status: item.currentStatus,
          submittedAt: item.submissionDate,
          reviewedAt: item.lastReviewedDate,
          approvedAt: item.approvalDate,
          reviewerName: item.reviewerName,
          daysInReview: item.daysInReview,
        } as WaiverApplication)));
      }
    } catch (error) {
      message.error('加载审批历史失败');
    } finally {
      setLoading(false);
    }
  }, []);

  // 加载统计数据
  const loadStatistics = useCallback(async () => {
    try {
      setLoading(true);
      const response = await waiverApprovalService.getWaiverStatistics();
      if (response.success) {
        setStatistics(response.data);
      }
    } catch (error) {
      message.error('加载统计数据失败');
    } finally {
      setLoading(false);
    }
  }, []);

  // 初始加载
  useEffect(() => {
    if (activeTab === 'pending') {
      loadPendingApplications();
    } else if (activeTab === 'history') {
      loadHistoryApplications();
    } else if (activeTab === 'statistics') {
      loadStatistics();
    }
  }, [activeTab, loadPendingApplications, loadHistoryApplications, loadStatistics]);

  // 处理批准申请
  const handleApprove = async (values: any) => {
    if (!selectedApplication) return;

    try {
      setLoading(true);
      const request: ApproveWaiverApplicationRequest = {
        applicationId: selectedApplication.id,
        comments: values.comments,
        conditions: values.conditions,
        nextReviewDate: values.nextReviewDate?.format('YYYY-MM-DD'),
        isElectronicSignature: values.isElectronicSignature,
        signatureData: values.isElectronicSignature ? {
          signatureBase64: signatureData,
          certificateId: 'default_certificate', // 实际项目中从用户配置获取
        } : undefined,
      };

      const response = await waiverApprovalService.approveWaiverApplication(request);
      if (response.success) {
        message.success('豁免申请批准成功');
        setApprovalModalVisible(false);
        approvalForm.resetFields();
        setSignatureData('');
        loadPendingApplications();
        setSelectedApplication(null);
      }
    } catch (error: any) {
      message.error(error.message || '批准失败');
    } finally {
      setLoading(false);
    }
  };

  // 处理拒绝申请
  const handleReject = async (values: any) => {
    if (!selectedApplication) return;

    try {
      setLoading(true);
      const request: RejectWaiverApplicationRequest = {
        applicationId: selectedApplication.id,
        comments: values.comments,
        reason: values.reason,
        isElectronicSignature: values.isElectronicSignature,
        signatureData: values.isElectronicSignature ? {
          signatureBase64: signatureData,
          certificateId: 'default_certificate',
        } : undefined,
      };

      const response = await waiverApprovalService.rejectWaiverApplication(request);
      if (response.success) {
        message.success('豁免申请拒绝成功');
        setRejectionModalVisible(false);
        rejectionForm.resetFields();
        setSignatureData('');
        loadPendingApplications();
        setSelectedApplication(null);
      }
    } catch (error: any) {
      message.error(error.message || '拒绝失败');
    } finally {
      setLoading(false);
    }
  };

  // 处理上报申请
  const handleEscalate = async (values: any) => {
    if (!selectedApplication) return;

    try {
      setLoading(true);
      const request: EscalateWaiverApplicationRequest = {
        applicationId: selectedApplication.id,
        escalationReason: values.escalationReason,
        escalatedToUserId: values.escalatedToUserId,
        comments: values.comments,
        isElectronicSignature: values.isElectronicSignature,
        signatureData: values.isElectronicSignature ? {
          signatureBase64: signatureData,
          certificateId: 'default_certificate',
        } : undefined,
      };

      const response = await waiverApprovalService.escalateWaiverApplication(request);
      if (response.success) {
        message.success('豁免申请上报成功');
        setEscalationModalVisible(false);
        escalationForm.resetFields();
        setSignatureData('');
        loadPendingApplications();
        setSelectedApplication(null);
      }
    } catch (error: any) {
      message.error(error.message || '上报失败');
    } finally {
      setLoading(false);
    }
  };

  // 查看申请详情
  const viewApplicationDetail = async (application: WaiverApplication) => {
    try {
      setLoading(true);
      const response = await waiverApprovalService.getWaiverApplication(application.id);
      if (response.success) {
        setSelectedApplication(response.data);
        setDetailDrawerVisible(true);
      }
    } catch (error) {
      message.error('获取申请详情失败');
    } finally {
      setLoading(false);
    }
  };

  // 发送提醒
  const sendReminder = async (applicationId: string) => {
    try {
      await waiverApprovalService.sendApprovalReminder(applicationId);
      message.success('提醒发送成功');
    } catch (error: any) {
      message.error(error.message || '发送提醒失败');
    }
  };

  // 待审批申请表格列
  const pendingColumns: ColumnsType<WaiverApplication> = [
    {
      title: '案件编号',
      dataIndex: 'caseId',
      key: 'caseId',
      width: 120,
      render: (text, record) => (
        <Button type="link" onClick={() => viewApplicationDetail(record)}>
          {text}
        </Button>
      ),
    },
    {
      title: '案件标题',
      dataIndex: 'caseTitle',
      key: 'caseTitle',
      ellipsis: true,
    },
    {
      title: '豁免类型',
      dataIndex: 'waiverType',
      key: 'waiverType',
      width: 120,
      render: (type) => {
        const typeLabels: Record<WaiverType, string> = {
          CONFLICT_OF_INTEREST: '利益冲突',
          BUSINESS_RELATIONSHIP: '业务关系',
          REPRESENTATION_CONFLICT: '代理冲突',
          FINANCIAL_INTEREST: '财务利益',
          PERSONAL_RELATIONSHIP: '个人关系',
          ORGANIZATIONAL: '组织冲突',
        };
        return <Tag color="blue">{typeLabels[type]}</Tag>;
      },
    },
    {
      title: '风险等级',
      dataIndex: 'riskLevel',
      key: 'riskLevel',
      width: 100,
      render: (level) => {
        const config = riskLevelConfig[level];
        return <Badge color={config.color} text={config.label} />;
      },
      sorter: (a, b) => riskLevelConfig[a.riskLevel].priority - riskLevelConfig[b.riskLevel].priority,
    },
    {
      title: '申请人',
      dataIndex: 'applicantName',
      key: 'applicantName',
      width: 100,
    },
    {
      title: '提交时间',
      dataIndex: 'submittedAt',
      key: 'submittedAt',
      width: 120,
      render: (date) => dayjs(date).format('MM-DD HH:mm'),
      sorter: (a, b) => dayjs(a.submittedAt).unix() - dayjs(b.submittedAt).unix(),
    },
    {
      title: '等待时间',
      key: 'waitingTime',
      width: 100,
      render: (_, record) => {
        const days = dayjs().diff(dayjs(record.submittedAt), 'day');
        return (
          <Tag color={days > 3 ? 'red' : days > 1 ? 'orange' : 'green'}>
            {days}天
          </Tag>
        );
      },
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_, record) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => viewApplicationDetail(record)}
            />
          </Tooltip>
          <Tooltip title="批准">
            <Button
              type="text"
              icon={<CheckCircleOutlined />}
              onClick={() => {
                setSelectedApplication(record);
                setApprovalModalVisible(true);
              }}
            />
          </Tooltip>
          <Tooltip title="拒绝">
            <Button
              type="text"
              danger
              icon={<CloseCircleOutlined />}
              onClick={() => {
                setSelectedApplication(record);
                setRejectionModalVisible(true);
              }}
            />
          </Tooltip>
          <Tooltip title="上报">
            <Button
              type="text"
              icon={<ExclamationCircleOutlined />}
              onClick={() => {
                setSelectedApplication(record);
                setEscalationModalVisible(true);
              }}
            />
          </Tooltip>
          <Tooltip title="发送提醒">
            <Button
              type="text"
              icon={<SendOutlined />}
              onClick={() => sendReminder(record.id)}
            />
          </Tooltip>
        </Space>
      ),
    },
  ];

  // 渲染待审批申请列表
  const renderPendingApplications = () => (
    <Card
      title="待审批申请"
      extra={
        <Space>
          <Button icon={<FilterOutlined />}>筛选</Button>
          <Button icon={<SearchOutlined />}>搜索</Button>
          <Button icon={<SettingOutlined />}>设置</Button>
        </Space>
      }
    >
      <Table
        columns={pendingColumns}
        dataSource={pendingApplications}
        loading={loading}
        pagination={{
          total: pendingApplications.length,
          pageSize: 10,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 条记录`,
        }}
        rowKey="id"
        size="small"
      />
    </Card>
  );

  // 渲染审批历史
  const renderApprovalHistory = () => (
    <Card title="审批历史">
      <List
        dataSource={historyApplications}
        loading={loading}
        renderItem={(item) => (
          <List.Item
            actions={[
              <Button type="link" icon={<EyeOutlined />}>
                查看
              </Button>,
            ]}
          >
            <List.Item.Meta
              avatar={<Avatar icon={<UserOutlined />} />}
              title={item.caseTitle}
              description={
                <Space>
                  <Tag color={statusConfig[item.status].color}>
                    {statusConfig[item.status].label}
                  </Tag>
                  <Text type="secondary">
                    提交时间：{dayjs(item.submittedAt).format('YYYY-MM-DD HH:mm')}
                  </Text>
                  {item.reviewedAt && (
                    <Text type="secondary">
                      审查时间：{dayjs(item.reviewedAt).format('YYYY-MM-DD HH:mm')}
                    </Text>
                  )}
                  <Text type="secondary">
                    处理用时：{item.daysInReview}天
                  </Text>
                </Space>
              }
            />
          </List.Item>
        )}
      />
    </Card>
  );

  // 渲染统计信息
  const renderStatistics = () => {
    if (!statistics) return null;

    return (
      <div>
        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col span={6}>
            <Card>
              <Statistic
                title="总申请数"
                value={statistics.totalApplications}
                prefix={<FileTextOutlined />}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="待审批"
                value={statistics.pendingApplications}
                valueStyle={{ color: '#fa8c16' }}
                prefix={<ClockCircleOutlined />}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="已批准"
                value={statistics.approvedApplications}
                valueStyle={{ color: '#52c41a' }}
                prefix={<CheckCircleOutlined />}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="已拒绝"
                value={statistics.rejectedApplications}
                valueStyle={{ color: '#f5222d' }}
                prefix={<CloseCircleOutlined />}
              />
            </Card>
          </Col>
        </Row>

        <Row gutter={16}>
          <Col span={12}>
            <Card title="风险等级分布">
              <div style={{ height: 200 }}>
                {Object.entries(statistics.riskLevelDistribution).map(([level, count]) => (
                  <div key={level} style={{ marginBottom: 8 }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                      <Badge color={riskLevelConfig[level as RiskLevel].color} text={riskLevelConfig[level as RiskLevel].label} />
                      <Text strong>{count}</Text>
                    </div>
                    <Progress
                      percent={statistics.totalApplications > 0 ? (count / statistics.totalApplications) * 100 : 0}
                      strokeColor={riskLevelConfig[level as RiskLevel].color}
                      showInfo={false}
                      size="small"
                    />
                  </div>
                ))}
              </div>
            </Card>
          </Col>
          <Col span={12}>
            <Card title="审批效率">
              <Row gutter={16}>
                <Col span={12}>
                  <Statistic
                    title="批准率"
                    value={statistics.approvalRate}
                    precision={1}
                    suffix="%"
                    valueStyle={{ color: '#52c41a' }}
                  />
                </Col>
                <Col span={12}>
                  <Statistic
                    title="平均审批时间"
                    value={statistics.averageReviewTime}
                    precision={1}
                    suffix="天"
                    valueStyle={{ color: '#1890ff' }}
                  />
                </Col>
              </Row>
            </Card>
          </Col>
        </Row>
      </div>
    );
  };

  // 渲染申请详情抽屉
  const renderApplicationDetail = () => {
    if (!selectedApplication) return null;

    return (
      <Drawer
        title="申请详情"
        placement="right"
        onClose={() => setDetailDrawerVisible(false)}
        open={detailDrawerVisible}
        width={800}
      >
        <Tabs defaultActiveKey="basic">
          <TabPane tab="基本信息" key="basic">
            <Descriptions column={1} bordered>
              <Descriptions.Item label="案件编号">
                {selectedApplication.caseId}
              </Descriptions.Item>
              <Descriptions.Item label="案件标题">
                {selectedApplication.caseTitle}
              </Descriptions.Item>
              <Descriptions.Item label="豁免类型">
                <Tag color="blue">
                  {selectedApplication.waiverType}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="风险等级">
                <Badge color={riskLevelConfig[selectedApplication.riskLevel].color} text={riskLevelConfig[selectedApplication.riskLevel].label} />
              </Descriptions.Item>
              <Descriptions.Item label="申请人">
                {selectedApplication.applicantName}
              </Descriptions.Item>
              <Descriptions.Item label="提交时间">
                {dayjs(selectedApplication.submittedAt).format('YYYY-MM-DD HH:mm:ss')}
              </Descriptions.Item>
              <Descriptions.Item label="状态">
                <Badge color={statusConfig[selectedApplication.status].color} text={statusConfig[selectedApplication.status].label} />
              </Descriptions.Item>
            </Descriptions>

            <Divider />

            <Title level={5}>豁免描述</Title>
            <Paragraph>
              {selectedApplication.description}
            </Paragraph>

            <Title level={5}>申请理由</Title>
            <Paragraph>
              {selectedApplication.justification}
            </Paragraph>

            <Title level={5}>风险控制措施</Title>
            <Paragraph>
              {selectedApplication.mitigationMeasures}
            </Paragraph>
          </TabPane>

          <TabPane tab="利益相关方" key="stakeholders">
            <List
              dataSource={selectedApplication.stakeholders}
              renderItem={(stakeholder) => (
                <List.Item>
                  <List.Item.Meta
                    avatar={<Avatar icon={<UserOutlined />} />}
                    title={stakeholder.name}
                    description={
                      <Space direction="vertical" size="small">
                        <Text type="secondary">类型：{stakeholder.type}</Text>
                        <Text type="secondary">关系：{stakeholder.relationshipDescription}</Text>
                        <Text type="secondary">冲突详情：{stakeholder.conflictDetails}</Text>
                      </Space>
                    }
                  />
                </List.Item>
              )}
            />
          </TabPane>

          <TabPane tab="附件" key="attachments">
            <List
              dataSource={selectedApplication.attachments}
              renderItem={(attachment) => (
                <List.Item
                  actions={[
                    <Button type="link" icon={<DownloadOutlined />}>
                      下载
                    </Button>,
                  ]}
                >
                  <List.Item.Meta
                    avatar={<FileTextOutlined />}
                    title={attachment.originalName}
                    description={
                      <Space>
                        <Text type="secondary">
                          大小：{(attachment.fileSize / 1024 / 1024).toFixed(2)} MB
                        </Text>
                        <Text type="secondary">
                          上传时间：{dayjs(attachment.uploadedAt).format('YYYY-MM-DD HH:mm')}
                        </Text>
                      </Space>
                    }
                  />
                </List.Item>
              )}
            />
          </TabPane>

          <TabPane tab="审批记录" key="approval">
            <Timeline>
              {selectedApplication.approvalRecords.map((record) => (
                <Timeline.Item
                  key={record.id}
                  color={
                    record.decision === 'APPROVE' ? 'green' :
                    record.decision === 'REJECT' ? 'red' :
                    record.decision === 'ESCALATE' ? 'purple' : 'blue'
                  }
                >
                  <div>
                    <Text strong>{record.reviewerName}</Text>
                    <Tag color="blue" style={{ marginLeft: 8 }}>
                      {record.decision === 'APPROVE' ? '批准' :
                       record.decision === 'REJECT' ? '拒绝' :
                       record.decision === 'ESCALATE' ? '上报' : '要求信息'}
                    </Tag>
                  </div>
                  <div>{record.comments}</div>
                  {record.conditions && (
                    <div style={{ marginTop: 4 }}>
                      <Text type="secondary">批准条件：</Text>
                      <Text>{record.conditions}</Text>
                    </div>
                  )}
                  <div style={{ marginTop: 4 }}>
                    <Text type="secondary">
                      {dayjs(record.reviewedAt).format('YYYY-MM-DD HH:mm:ss')}
                    </Text>
                  </div>
                </Timeline.Item>
              ))}
            </Timeline>
          </TabPane>
        </Tabs>
      </Drawer>
    );
  };

  return (
    <div style={{ padding: 24 }}>
      <Title level={2}>豁免审批管理</Title>

      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane
          tab={
            <span>
              <ClockCircleOutlined />
              待审批 ({pendingApplications.length})
            </span>
          }
          key="pending"
        >
          {renderPendingApplications()}
        </TabPane>

        <TabPane
          tab={
            <span>
              <HistoryOutlined />
              审批历史
            </span>
          }
          key="history"
        >
          {renderApprovalHistory()}
        </TabPane>

        <TabPane
          tab={
            <span>
              <AuditOutlined />
              统计分析
            </span>
          }
          key="statistics"
        >
          {renderStatistics()}
        </TabPane>
      </Tabs>

      {/* 批准模态框 */}
      <Modal
        title="批准豁免申请"
        open={approvalModalVisible}
        onCancel={() => setApprovalModalVisible(false)}
        footer={null}
        width={600}
      >
        <Form form={approvalForm} layout="vertical" onFinish={handleApprove}>
          <Form.Item
            label="审批意见"
            name="comments"
            rules={[{ required: true, message: '请输入审批意见' }]}
          >
            <TextArea rows={4} placeholder="请输入审批意见" />
          </Form.Item>

          <Form.Item label="批准条件" name="conditions">
            <TextArea rows={3} placeholder="如有特殊条件，请在此说明" />
          </Form.Item>

          <Form.Item label="下次审查日期" name="nextReviewDate">
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item name="isElectronicSignature" valuePropName="checked" initialValue={true}>
            <Checkbox>使用电子签名</Checkbox>
          </Form.Item>

          <Form.Item>
            <Space>
              <Button onClick={() => setApprovalModalVisible(false)}>
                取消
              </Button>
              <Button type="primary" htmlType="submit" loading={loading}>
                确认批准
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 拒绝模态框 */}
      <Modal
        title="拒绝豁免申请"
        open={rejectionModalVisible}
        onCancel={() => setRejectionModalVisible(false)}
        footer={null}
        width={600}
      >
        <Form form={rejectionForm} layout="vertical" onFinish={handleReject}>
          <Form.Item
            label="拒绝原因"
            name="reason"
            rules={[{ required: true, message: '请输入拒绝原因' }]}
          >
            <Select placeholder="请选择拒绝原因">
              <Option value="信息不完整">信息不完整</Option>
              <Option value="风险过高">风险过高</Option>
              <Option value="不符合条件">不符合条件</Option>
              <Option value="其他">其他</Option>
            </Select>
          </Form.Item>

          <Form.Item
            label="详细说明"
            name="comments"
            rules={[{ required: true, message: '请输入详细说明' }]}
          >
            <TextArea rows={4} placeholder="请详细说明拒绝原因" />
          </Form.Item>

          <Form.Item name="isElectronicSignature" valuePropName="checked" initialValue={true}>
            <Checkbox>使用电子签名</Checkbox>
          </Form.Item>

          <Form.Item>
            <Space>
              <Button onClick={() => setRejectionModalVisible(false)}>
                取消
              </Button>
              <Button danger type="primary" htmlType="submit" loading={loading}>
                确认拒绝
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 上报模态框 */}
      <Modal
        title="上报豁免申请"
        open={escalationModalVisible}
        onCancel={() => setEscalationModalVisible(false)}
        footer={null}
        width={600}
      >
        <Form form={escalationForm} layout="vertical" onFinish={handleEscalate}>
          <Form.Item
            label="上报对象"
            name="escalatedToUserId"
            rules={[{ required: true, message: '请选择上报对象' }]}
          >
            <Select placeholder="请选择上报对象">
              <Option value="partner_001">张合伙人</Option>
              <Option value="partner_002">李合伙人</Option>
              <Option value="committee_001">管理委员会</Option>
            </Select>
          </Form.Item>

          <Form.Item
            label="上报原因"
            name="escalationReason"
            rules={[{ required: true, message: '请输入上报原因' }]}
          >
            <TextArea rows={4} placeholder="请说明上报原因" />
          </Form.Item>

          <Form.Item
            label="补充说明"
            name="comments"
          >
            <TextArea rows={3} placeholder="其他需要说明的情况" />
          </Form.Item>

          <Form.Item name="isElectronicSignature" valuePropName="checked" initialValue={true}>
            <Checkbox>使用电子签名</Checkbox>
          </Form.Item>

          <Form.Item>
            <Space>
              <Button onClick={() => setEscalationModalVisible(false)}>
                取消
              </Button>
              <Button type="primary" htmlType="submit" loading={loading}>
                确认上报
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 申请详情抽屉 */}
      {renderApplicationDetail()}
    </div>
  );
};

export default WaiverApprovalInterface;