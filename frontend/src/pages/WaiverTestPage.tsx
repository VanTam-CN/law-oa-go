import React from 'react'
import { Card, Button, Space, Typography, Divider, Tabs, Alert } from 'antd'
import { PlusOutlined, FileTextOutlined, AuditOutlined, MonitorOutlined } from '@ant-design/icons'
// import WaiverApplicationForm from '@/components/waiver/WaiverApplicationForm';
// import WaiverApprovalInterface from '@/components/waiver/WaiverApprovalInterface';
// import WaiverMonitoringDashboard from '@/components/waiver/WaiverMonitoringDashboard';
import type { WaiverApplication } from '@/types/waiverApproval'

const { Title, Text } = Typography
const { TabPane } = Tabs

const WaiverTestPage: React.FC = () => {
  const [applicationModalVisible, setApplicationModalVisible] = React.useState(false)
  const [activeTab, setActiveTab] = React.useState('form')

  const handleApplicationSuccess = (application: WaiverApplication) => {
    console.log('豁免申请创建成功:', application)
  }

  return (
    <div style={{ padding: 24 }}>
      <Title level={2}>豁免审批系统测试页面</Title>

      <Alert
        message='组件测试页面'
        description='这个页面用于测试所有豁免审批相关组件的功能和交互。'
        type='info'
        showIcon
        style={{ marginBottom: 24 }}
      />

      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane
          tab={
            <span>
              <FileTextOutlined />
              申请表单测试
            </span>
          }
          key='form'
        >
          <Card>
            <Title level={4}>豁免申请表单组件测试</Title>
            <Text>点击下方按钮测试豁免申请表单功能：</Text>
            <div style={{ marginTop: 16 }}>
              <Button
                type='primary'
                size='large'
                icon={<PlusOutlined />}
                onClick={() => setApplicationModalVisible(true)}
              >
                打开豁免申请表单
              </Button>
            </div>

            <Divider />

            <div>
              <Title level={5}>测试功能列表：</Title>
              <ul>
                <li>多步骤表单流程</li>
                <li>豁免类型选择和风险评估</li>
                <li>利益相关方管理</li>
                <li>文件上传功能</li>
                <li>表单验证</li>
                <li>模板应用</li>
                <li>草稿保存和提交</li>
              </ul>
            </div>
          </Card>

          {/* 组件暂时禁用 */}
          {/* <WaiverApplicationForm
            visible={applicationModalVisible}
            onCancel={() => setApplicationModalVisible(false)}
            onSuccess={handleApplicationSuccess}
            caseId="TEST_CASE_001"
            caseTitle="测试案件 - 豁免审批功能测试"
          /> */}
        </TabPane>

        <TabPane
          tab={
            <span>
              <AuditOutlined />
              审批界面测试
            </span>
          }
          key='approval'
        >
          <Card>
            <Title level={4}>豁免审批界面组件测试</Title>
            <Text>测试审批处理功能，包括待审批列表、审批操作、历史记录等。</Text>

            <Divider />

            {/* 组件暂时禁用 */}
            {/* <WaiverApprovalInterface userRole="LAWYER" userId="test_user_001" /> */}
            <Alert
              message='组件开发中'
              description='豁免审批界面组件正在完善中，敬请期待！'
              type='info'
              showIcon
            />
          </Card>
        </TabPane>

        <TabPane
          tab={
            <span>
              <MonitorOutlined />
              监控看板测试
            </span>
          }
          key='monitoring'
        >
          <Card>
            <Title level={4}>监控管理看板组件测试</Title>
            <Text>测试监控任务管理、合规分析、统计报表等功能。</Text>

            <Divider />

            {/* 组件暂时禁用 */}
            {/* <WaiverMonitoringDashboard userRole="LAWYER" userId="test_user_001" /> */}
            <Alert
              message='组件开发中'
              description='监控管理看板组件正在完善中，敬请期待！'
              type='info'
              showIcon
            />
          </Card>
        </TabPane>
      </Tabs>
    </div>
  )
}

export default WaiverTestPage
