import React, { useState } from 'react'
import { Card, Row, Col, Button, Space, Typography, Tabs, Badge, Alert, Divider } from 'antd'
import {
  PlusOutlined,
  FileTextOutlined,
  AuditOutlined,
  MonitorOutlined,
  SettingOutlined,
  BookOutlined,
} from '@ant-design/icons'
import WaiverApplicationForm from './WaiverApplicationForm'
import WaiverApprovalInterface from './WaiverApprovalInterface'
import WaiverMonitoringDashboard from './WaiverMonitoringDashboard'
import type { WaiverApplication } from '@/types/waiverApproval'

const { Title, Text, Paragraph } = Typography
const { TabPane } = Tabs

interface WaiverApprovalExampleProps {
  userRole?: string
  userId?: string
}

const WaiverApprovalExample: React.FC<WaiverApprovalExampleProps> = ({
  userRole = 'LAWYER',
  userId = 'user_001',
}) => {
  const [applicationModalVisible, setApplicationModalVisible] = useState(false)
  const [activeTab, setActiveTab] = useState('overview')

  // 处理豁免申请创建成功
  const handleApplicationSuccess = (application: WaiverApplication) => {
    console.log('豁免申请创建成功:', application)
    // 可以在这里添加成功后的逻辑，比如刷新列表、显示通知等
  }

  // 渲染功能概览
  const renderOverview = () => (
    <div>
      <Alert
        message='豁免审批管理系统'
        description='完整的律师事务所豁免审批工作流解决方案，包括申请创建、审批处理、监控管理等全流程功能。'
        type='info'
        showIcon
        style={{ marginBottom: 24 }}
      />

      <Row gutter={[16, 16]}>
        <Col span={8}>
          <Card
            hoverable
            actions={[
              <Button
                type='primary'
                icon={<PlusOutlined />}
                onClick={() => setApplicationModalVisible(true)}
              >
                创建申请
              </Button>,
            ]}
          >
            <Card.Meta
              avatar={<FileTextOutlined style={{ fontSize: 24, color: '#1890ff' }} />}
              title='豁免申请管理'
              description={
                <div>
                  <Paragraph style={{ marginBottom: 8 }}>
                    支持多种豁免类型申请，包括利益冲突、业务关系、代理冲突等。
                  </Paragraph>
                  <Space wrap>
                    <Badge color='blue' text='多类型支持' />
                    <Badge color='green' text='智能表单' />
                    <Badge color='orange' text='风险评估' />
                    <Badge color='purple' text='电子签名' />
                  </Space>
                </div>
              }
            />
          </Card>
        </Col>

        <Col span={8}>
          <Card
            hoverable
            actions={[
              <Button
                type='primary'
                icon={<AuditOutlined />}
                onClick={() => setActiveTab('approval')}
              >
                审批处理
              </Button>,
            ]}
          >
            <Card.Meta
              avatar={<AuditOutlined style={{ fontSize: 24, color: '#52c41a' }} />}
              title='审批工作流'
              description={
                <div>
                  <Paragraph style={{ marginBottom: 8 }}>
                    完整的审批流程管理，支持多级审批、电子签名、审批记录等。
                  </Paragraph>
                  <Space wrap>
                    <Badge color='blue' text='多级审批' />
                    <Badge color='green' text='电子签名' />
                    <Badge color='orange' text='审批追溯' />
                    <Badge color='purple' text='智能提醒' />
                  </Space>
                </div>
              }
            />
          </Card>
        </Col>

        <Col span={8}>
          <Card
            hoverable
            actions={[
              <Button
                type='primary'
                icon={<MonitorOutlined />}
                onClick={() => setActiveTab('monitoring')}
              >
                监控管理
              </Button>,
            ]}
          >
            <Card.Meta
              avatar={<MonitorOutlined style={{ fontSize: 24, color: '#fa8c16' }} />}
              title='监控与合规'
              description={
                <div>
                  <Paragraph style={{ marginBottom: 8 }}>
                    豁免批准后的持续监控，确保合规执行和风险控制。
                  </Paragraph>
                  <Space wrap>
                    <Badge color='blue' text='任务监控' />
                    <Badge color='green' text='合规检查' />
                    <Badge color='orange' text='风险预警' />
                    <Badge color='purple' text='数据分析' />
                  </Space>
                </div>
              }
            />
          </Card>
        </Col>
      </Row>

      <Divider />

      <Title level={4}>系统特性</Title>
      <Row gutter={[16, 16]}>
        <Col span={12}>
          <Card title='核心功能' size='small'>
            <ul>
              <li>多类型豁免申请支持</li>
              <li>智能风险评估和等级划分</li>
              <li>多级审批工作流</li>
              <li>电子签名和数字证书</li>
              <li>审批记录和审计追踪</li>
              <li>自动化监控任务</li>
              <li>合规性分析和报告</li>
              <li>实时通知和提醒</li>
            </ul>
          </Card>
        </Col>
        <Col span={12}>
          <Card title='技术特点' size='small'>
            <ul>
              <li>React + TypeScript 前端架构</li>
              <li>Go 后端服务支持</li>
              <li>RESTful API 接口设计</li>
              <li>组件化模块开发</li>
              <li>响应式用户界面</li>
              <li>实时数据更新</li>
              <li>安全的权限控制</li>
              <li>完善的错误处理</li>
            </ul>
          </Card>
        </Col>
      </Row>
    </div>
  )

  return (
    <div style={{ padding: 24 }}>
      <Title level={2}>豁免审批系统演示</Title>

      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane
          tab={
            <span>
              <BookOutlined />
              功能概览
            </span>
          }
          key='overview'
        >
          {renderOverview()}
        </TabPane>

        <TabPane
          tab={
            <span>
              <PlusOutlined />
              申请创建
            </span>
          }
          key='application'
        >
          <Card title='创建豁免申请'>
            <Paragraph>
              点击下方按钮创建新的豁免申请。系统支持多种豁免类型，会根据类型自动进行风险评估。
            </Paragraph>
            <Button
              type='primary'
              size='large'
              icon={<PlusOutlined />}
              onClick={() => setApplicationModalVisible(true)}
            >
              创建豁免申请
            </Button>
          </Card>
        </TabPane>

        <TabPane
          tab={
            <span>
              <AuditOutlined />
              审批处理
            </span>
          }
          key='approval'
        >
          <WaiverApprovalInterface userRole={userRole} userId={userId} />
        </TabPane>

        <TabPane
          tab={
            <span>
              <MonitorOutlined />
              监控管理
            </span>
          }
          key='monitoring'
        >
          <WaiverMonitoringDashboard userRole={userRole} userId={userId} />
        </TabPane>

        <TabPane
          tab={
            <span>
              <SettingOutlined />
              系统配置
            </span>
          }
          key='settings'
        >
          <Card title='系统配置'>
            <Alert
              message='配置功能开发中'
              description='系统配置功能包括用户权限管理、审批流程配置、模板管理等，正在开发中。'
              type='info'
              showIcon
            />

            <div style={{ marginTop: 16 }}>
              <Title level={5}>计划配置功能</Title>
              <ul>
                <li>用户角色和权限管理</li>
                <li>审批流程自定义配置</li>
                <li>豁免申请模板管理</li>
                <li>通知规则设置</li>
                <li>监控频率配置</li>
                <li>报表和分析设置</li>
                <li>系统参数管理</li>
                <li>数据备份和恢复</li>
              </ul>
            </div>
          </Card>
        </TabPane>
      </Tabs>

      {/* 豁免申请表单模态框 */}
      <WaiverApplicationForm
        visible={applicationModalVisible}
        onCancel={() => setApplicationModalVisible(false)}
        onSuccess={handleApplicationSuccess}
        caseId='DEMO_CASE_001'
        caseTitle='演示案件 - 公司合同纠纷'
      />
    </div>
  )
}

export default WaiverApprovalExample
