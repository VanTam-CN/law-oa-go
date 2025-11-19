import React, { useState, useEffect } from 'react';
import {
  Card,
  Button,
  Space,
  Typography,
  Alert,
  Table,
  Tag,
  Badge,
  Descriptions,
  Timeline,
  Row,
  Col,
  Statistic,
  Tabs,
  Empty,
  message,
  Modal,
} from 'antd';
import {
  WarningOutlined,
  FileTextOutlined,
  UserOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
// import EnhancedWaiverApprovalInterface from '@/components/waiver/EnhancedWaiverApprovalInterface';
// import { conflictDetectionAdapter, createWaiverSuggestion } from '@/services/conflictDetectionAdapter';
import type { ConflictCase, ConflictCheckResponse } from '@/types/waiverApproval';

const { Title, Text, Paragraph } = Typography;
const { TabPane } = Tabs;

// 模拟的冲突检测数据 - 基于你提供的实际数据格式
const mockConflictData = [
  {
    CaseID: '24',
    CaseName: '中国中铁铁路工程纠纷案',
    ConflictType: '案件冲突',
    RiskLevel: 'LOW',
    ConflictDetail: '律师 王芳 同时代理了案件 \'中国中铁铁路工程纠纷案\'，存在潜在利益冲突',
    CreatedTime: '2025/11/5 10:22:31',
    Status: '进行中',
    LawyerName: '王芳',
  },
  {
    CaseID: '23',
    CaseName: '中国建筑集团建设工程合同纠纷',
    ConflictType: '案件冲突',
    RiskLevel: 'LOW',
    ConflictDetail: '律师 王芳 同时代理了案件 \'中国建筑集团建设工程合同纠纷\'，存在潜在利益冲突',
    CreatedTime: '2025/11/5 10:22:31',
    Status: '进行中',
    LawyerName: '王芳',
  },
  {
    CaseID: '35',
    CaseName: '中国建筑集团有限公司诉中国中铁股份有限公司项目竞标纠纷案',
    ConflictType: '案件冲突',
    RiskLevel: 'LOW',
    ConflictDetail: '律师 王芳 同时代理了案件 \'中国建筑集团有限公司诉中国中铁股份有限公司项目竞标纠纷案\'，存在潜在利益冲突',
    CreatedTime: '2025/11/5 10:22:31',
    Status: '进行中',
    LawyerName: '王芳',
  },
];

const ConflictDataDemo: React.FC = () => {
  const [activeTab, setActiveTab] = useState('demo');
  const [conflicts, setConflicts] = useState<ConflictCase[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedConflict, setSelectedConflict] = useState<ConflictCase | null>(null);
  const [demoModalVisible, setDemoModalVisible] = useState(false);

  useEffect(() => {
    // 转换模拟数据为标准格式
    const convertedConflicts = mockConflictData.map(legacyCase => {
      // 直接使用适配器的转换方法
      return {
        id: legacyCase.CaseID,
        caseId: legacyCase.CaseID,
        caseName: legacyCase.CaseName,
        conflictType: 'REPRESENTATION_CONFLICT',
        description: legacyCase.ConflictDetail,
        riskLevel: legacyCase.RiskLevel === 'LOW' ? 'LOW' : legacyCase.RiskLevel === 'MEDIUM' ? 'MEDIUM' : 'HIGH',
        status: 'ACTIVE',
        createdAt: legacyCase.CreatedTime,
        updatedAt: legacyCase.CreatedTime,
        parties: [
          {
            name: legacyCase.LawyerName || '未知律师',
            type: 'LAWYER',
            role: '代理律师'
          }
        ],
        mitigationMeasures: '信息隔离：确保不同案件的信息严格隔离；内部沟通：建立定期沟通机制',
        resolutionNotes: '',
        assignedTo: legacyCase.LawyerName || '',
        resolvedAt: null
      };
    });
    setConflicts(convertedConflicts);
  }, []);

  // 冲突案例表格列定义
  const conflictColumns: ColumnsType<ConflictCase> = [
    {
      title: '案件信息',
      key: 'case',
      render: (_, record) => (
        <div>
          <div style={{ fontWeight: 'bold', marginBottom: 4 }}>
            {record.caseName}
          </div>
          <div style={{ fontSize: '12px', color: '#666' }}>
            案件编号: {record.caseId}
          </div>
        </div>
      ),
    },
    {
      title: '冲突类型',
      dataIndex: 'conflictType',
      key: 'conflictType',
      render: (type) => (
        <Tag color="red">
          代理冲突
        </Tag>
      ),
    },
    {
      title: '风险等级',
      dataIndex: 'riskLevel',
      key: 'riskLevel',
      render: (level) => (
        <Tag color="green">
          低风险
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status) => (
        <Badge
          status="processing"
          text="进行中"
        />
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (time) => dayjs(time).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Button
            size="small"
            icon={<InfoCircleOutlined />}
            onClick={() => {
              setSelectedConflict(record);
              setDemoModalVisible(true);
            }}
          >
            查看详情
          </Button>
          <Button
            type="primary"
            size="small"
            icon={<FileTextOutlined />}
            onClick={() => {
              message.success('豁免申请创建成功');
            }}
          >
            申请豁免
          </Button>
        </Space>
      ),
    },
  ];

  // 渲染数据演示页面
  const renderDataDemo = () => (
    <div>
      <Alert
        message="冲突检测数据演示"
        description="这个页面展示了如何将实际的冲突检测数据（如你提供的格式）转换为豁免审批系统可用的格式。"
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
      />

      <Card
        title="原始冲突检测数据"
        extra={
          <Space>
            <Button icon={<WarningOutlined />}>
              模拟冲突检测
            </Button>
            <Button icon={<FileTextOutlined />} type="primary">
              批量申请豁免
            </Button>
          </Space>
        }
      >
        <div style={{ marginBottom: 16 }}>
          <Title level={4}>风险原因和处理措施</Title>
          <Paragraph>
            <strong>风险原因：</strong>记录冲突情况；定期复查；保持警惕；建立预防措施；针对代理冲突类型案件制定专门处理流程
          </Paragraph>
          <Paragraph>
            <strong>风险因素：</strong>中风险冲突3个
          </Paragraph>
        </div>

        <Table
          columns={conflictColumns}
          dataSource={conflicts}
          rowKey="id"
          pagination={false}
        />
      </Card>
    </div>
  );

  // 渲染集成功能演示
  const renderIntegrationDemo = () => (
    <div>
      <Alert
        message="集成功能演示"
        description="增强型豁免审批界面正在开发中，请先查看数据演示页面。"
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
      />

      <Card title="功能说明">
        <Paragraph>
          完整的集成功能演示将包括：
        </Paragraph>
        <ul>
          <li>冲突检测数据自动导入</li>
          <li>豁免申请一键生成</li>
          <li>审批流程完整追踪</li>
          <li>数据格式无缝转换</li>
        </ul>
        <Alert
          message="开发中"
          description="集成功能正在完善中，敬请期待！"
          type="warning"
          showIcon
          style={{ marginTop: 16 }}
        />
      </Card>
    </div>
  );

  // 渲染数据转换说明
  const renderConversionGuide = () => (
    <div>
      <Card title="数据格式转换指南">
        <Row gutter={16}>
          <Col span={12}>
            <Card size="small" title="原始数据格式">
              <pre style={{ background: '#f5f5f5', padding: 12, borderRadius: 4, fontSize: 12 }}>
{`{
  "CaseID": "24",
  "CaseName": "中国中铁铁路工程纠纷案",
  "ConflictType": "案件冲突",
  "RiskLevel": "LOW",
  "ConflictDetail": "律师 王芳 同时代理了案件...",
  "CreatedTime": "2025/11/5 10:22:31",
  "Status": "进行中",
  "LawyerName": "王芳"
}`}
              </pre>
            </Card>
          </Col>
          <Col span={12}>
            <Card size="small" title="标准格式">
              <pre style={{ background: '#f5f5f5', padding: 12, borderRadius: 4, fontSize: 12 }}>
{`{
  "id": "24",
  "caseId": "24",
  "caseName": "中国中铁铁路工程纠纷案",
  "conflictType": "REPRESENTATION_CONFLICT",
  "riskLevel": "LOW",
  "description": "律师 王芳 同时代理了案件...",
  "status": "ACTIVE",
  "createdAt": "2025/11/5 10:22:31",
  "parties": [
    { "name": "王芳", "type": "LAWYER" }
  ],
  "mitigationMeasures": "信息隔离：确保不同案件的信息严格隔离..."
}`}
              </pre>
            </Card>
          </Col>
        </Row>

        <Divider />

        <Title level={4}>转换要点</Title>
        <ul>
          <li><strong>字段映射：</strong>CaseID → id/caseId, CaseName → caseName</li>
          <li><strong>类型转换：</strong>案件冲突 → REPRESENTATION_CONFLICT</li>
          <li><strong>状态映射：</strong>进行中 → ACTIVE</li>
          <li><strong>数据结构：</strong>扁平结构 → 嵌套结构（parties数组）</li>
          <li><strong>默认值：</strong>自动生成缺失的字段（mitigationMeasures等）</li>
        </ul>

        <Title level={4}>适配器功能</Title>
        <ul>
          <li><strong>ConflictDetectionAdapter：</strong>处理API调用和数据转换</li>
          <li><strong>ConflictWaiverAdapter：</strong>后端数据适配器</li>
          <li><strong>ConflictWaiverService：</strong>业务逻辑集成服务</li>
          <li><strong>EnhancedWaiverApprovalInterface：</strong>增强型UI组件</li>
        </ul>
      </Card>
    </div>
  );

  return (
    <div style={{ padding: 24 }}>
      <Title level={2}>
        <WarningOutlined /> 冲突检测数据演示
      </Title>
      <Paragraph>
        基于你提供的实际冲突检测数据，演示如何与豁免审批系统进行集成。
      </Paragraph>

      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane
          tab={
            <span>
              <WarningOutlined />
              数据演示
            </span>
          }
          key="demo"
        >
          {renderDataDemo()}
        </TabPane>

        <TabPane
          tab={
            <span>
              <FileTextOutlined />
              集成演示
            </span>
          }
          key="integration"
        >
          {renderIntegrationDemo()}
        </TabPane>

        <TabPane
          tab={
            <span>
              <InfoCircleOutlined />
              转换指南
            </span>
          }
          key="guide"
        >
          {renderConversionGuide()}
        </TabPane>
      </Tabs>

      {/* 冲突详情模态框 */}
      <Modal
        title="冲突详情"
        open={demoModalVisible}
        onCancel={() => setDemoModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDemoModalVisible(false)}>
            关闭
          </Button>,
          <Button
            key="create"
            type="primary"
            icon={<FileTextOutlined />}
            onClick={() => {
              message.success('豁免申请创建成功');
              setDemoModalVisible(false);
            }}
          >
            申请豁免
          </Button>,
        ]}
        width={800}
      >
        {selectedConflict && (
          <div>
            <Descriptions column={2} bordered>
              <Descriptions.Item label="案件名称" span={2}>
                {selectedConflict.caseName}
              </Descriptions.Item>
              <Descriptions.Item label="案件编号">
                {selectedConflict.caseId}
              </Descriptions.Item>
              <Descriptions.Item label="冲突类型">
                <Tag color="red">代理冲突</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="风险等级">
                <Tag color="green">低风险</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="当前状态">
                <Badge status="processing" text="进行中" />
              </Descriptions.Item>
              <Descriptions.Item label="创建时间">
                {dayjs(selectedConflict.createdAt).format('YYYY-MM-DD HH:mm:ss')}
              </Descriptions.Item>
              <Descriptions.Item label="冲突描述" span={2}>
                <Paragraph>{selectedConflict.description}</Paragraph>
              </Descriptions.Item>
              <Descriptions.Item label="相关方" span={2}>
                <Tag style={{ margin: '4px 4px 4px 0' }}>
                  王芳 (律师)
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="风险控制措施" span={2}>
                <Paragraph>
                  信息隔离：确保不同案件的信息严格隔离；内部沟通：建立定期沟通机制
                </Paragraph>
              </Descriptions.Item>
            </Descriptions>

            <Divider />
            <Title level={4}>建议的豁免申请内容</Title>
            <Alert
              message="豁免建议"
              description={
                <div>
                  <p><strong>豁免类型：</strong>REPRESENTATION_CONFLICT（代理冲突）</p>
                  <p><strong>风险等级：</strong>LOW（低风险）</p>
                  <p><strong>建议措施：</strong>信息隔离，确保不同案件的信息严格隔离；建立定期沟通机制</p>
                </div>
              }
              type="info"
              showIcon
            />
          </div>
        )}
      </Modal>
    </div>
  );
};

export default ConflictDataDemo;