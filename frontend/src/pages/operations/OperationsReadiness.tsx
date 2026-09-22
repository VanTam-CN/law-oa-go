import React from 'react'
import { Alert, Card, List, Progress, Space, Table, Tag, Typography } from 'antd'
import { summarizeOperationsReadiness } from '@/services/operationsReadiness'

const { Paragraph, Text, Title } = Typography

const OperationsReadiness: React.FC = () => {
  const summary = summarizeOperationsReadiness()

  return (
    <Space direction='vertical' size='large' style={{ width: '100%' }}>
      <Card>
        <Title level={3}>运维准备度</Title>
        <Paragraph>
          这个页面给律所主任和行政人员看：系统当前能用，不代表出事后能恢复。每一项都必须留下可复核的证据。
        </Paragraph>
        <Alert
          type='warning'
          showIcon
          message='健康检查通过不等于运维已就绪'
          description='健康检查只能说明当前服务可用。备份、恢复演练、故障负责人、升级和回滚仍需逐项补齐证据。'
        />
        <div style={{ marginTop: 24 }}>
          <Progress
            percent={Math.round((summary.verifiedCount / summary.total) * 100)}
            status='exception'
            format={() => `${summary.verifiedCount}/${summary.total} 项已有验证证据`}
          />
          <Text type='secondary'>
            当前未接入任何运维证据来源，所有事项均按“待补证据”处理，避免显示虚假通过。
          </Text>
        </div>
      </Card>

      <Card title='待补证据清单' data-testid='operations-readiness-requirements'>
        <Table
          dataSource={summary.items}
          rowKey='id'
          pagination={false}
          columns={[
            {
              title: '事项',
              dataIndex: 'title',
              render: (title: string, record) => (
                <Space direction='vertical' size='small'>
                  <Text strong>{title}</Text>
                  <Text type='secondary'>{record.userQuestion}</Text>
                </Space>
              ),
            },
            {
              title: '当前状态',
              dataIndex: 'status',
              width: 120,
              render: () => <Tag color='warning'>待补证据</Tag>,
            },
            {
              title: '需要留存',
              dataIndex: 'evidenceToKeep',
              render: (items: string[]) => (
                <List
                  size='small'
                  dataSource={items}
                  renderItem={(item) => <List.Item style={{ padding: '2px 0' }}>{item}</List.Item>}
                />
              ),
            },
            {
              title: '下一步动作',
              dataIndex: 'nextAction',
              render: (action: string) => <Text>{action}</Text>,
            },
          ]}
        />
      </Card>
    </Space>
  )
}

export default OperationsReadiness
