import React from 'react'
import { Card, Table, Tag, Space, Button, Descriptions, Alert, Progress, Divider, Tooltip } from 'antd'
import {
  ExclamationCircleOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  CloseCircleOutlined,
  FileTextOutlined,
  DownloadOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import type { ConflictCheckResult } from '@/types/conflict'
import './ConflictResult.less'

interface ConflictResultProps {
  result: ConflictCheckResult
  onGenerateReport?: () => void
  onViewDetails?: (caseId: string) => void
  loading?: boolean
}

// 风险等级配置
const RISK_LEVEL_CONFIG = {
  CRITICAL: {
    color: 'error',
    text: '严重风险',
    icon: <CloseCircleOutlined />,
    scoreRange: [85, 100],
    progressColor: '#a8071a',
  },
  HIGH: {
    color: 'error',
    text: '高风险',
    icon: <CloseCircleOutlined />,
    scoreRange: [70, 100],
    progressColor: '#ff4d4f',
  },
  MEDIUM: {
    color: 'warning',
    text: '中风险',
    icon: <WarningOutlined />,
    scoreRange: [40, 69],
    progressColor: '#faad14',
  },
  LOW: {
    color: 'processing',
    text: '低风险',
    icon: <ExclamationCircleOutlined />,
    scoreRange: [10, 39],
    progressColor: '#1890ff',
  },
  MINIMAL: {
    color: 'success',
    text: '极低风险',
    icon: <CheckCircleOutlined />,
    scoreRange: [0, 9],
    progressColor: '#52c41a',
  },
  PASS: {
    color: 'success',
    text: '无冲突',
    icon: <CheckCircleOutlined />,
    scoreRange: [0, 9],
    progressColor: '#52c41a',
  },
}

// 冲突类型配置
const CONFLICT_TYPE_CONFIG: Record<string, { color: string; text: string }> = {
  direct_conflict: { color: 'error', text: '直接冲突' },
  indirect_conflict: { color: 'warning', text: '间接冲突' },
  related_conflict: { color: 'processing', text: '关联冲突' },
  opposing_party: { color: 'error', text: '对方当事人' },
  corporate_relation: { color: 'default', text: '企业关联' },
  name_similarity: { color: 'default', text: '名称相似' },
}

const ConflictResult: React.FC<ConflictResultProps> = ({
  result,
  onGenerateReport,
  onViewDetails,
  loading = false,
}) => {
  const { hasConflict, conflictCases, checkStatistics, riskAssessment, recommendations } = result

  // 获取风险等级配置
  const getRiskConfig = () => {
    const level = riskAssessment?.overallRisk || 'PASS'
    return RISK_LEVEL_CONFIG[level as keyof typeof RISK_LEVEL_CONFIG] || RISK_LEVEL_CONFIG.PASS
  }

  // 获取冲突类型标签
  const getConflictTypeTag = (type: string) => {
    const config = CONFLICT_TYPE_CONFIG[type] || { color: 'default', text: type }
    return <Tag color={config.color}>{config.text}</Tag>
  }

  // 表格列定义
  const columns: ColumnsType<any> = [
    {
      title: '案件编号',
      dataIndex: 'caseId',
      key: 'caseId',
      width: 120,
      render: (text) => (
        <Tooltip title={text}>
          <span style={{ fontFamily: 'monospace' }}>{text?.slice(-8)}</span>
        </Tooltip>
      ),
    },
    {
      title: '案件名称',
      dataIndex: 'caseName',
      key: 'caseName',
      ellipsis: true,
      render: (text, record) => (
        <a onClick={() => onViewDetails?.(record.caseId)}>{text}</a>
      ),
    },
    {
      title: '客户名称',
      dataIndex: 'clientName',
      key: 'clientName',
      width: 150,
      ellipsis: true,
    },
    {
      title: '冲突类型',
      dataIndex: 'conflictType',
      key: 'conflictType',
      width: 100,
      render: (type: string) => getConflictTypeTag(type),
    },
    {
      title: '风险等级',
      dataIndex: 'riskLevel',
      key: 'riskLevel',
      width: 100,
      render: (level: string) => {
        const config = RISK_LEVEL_CONFIG[level as keyof typeof RISK_LEVEL_CONFIG]
        return config ? (
          <Tag color={config.color} icon={config.icon}>
            {config.text}
          </Tag>
        ) : (
          <Tag>{level}</Tag>
        )
      },
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
  ]

  const riskConfig = getRiskConfig()

  return (
    <div className="conflict-result-container">
      {/* 风险评估摘要 */}
      <Card
        title={
          <Space>
            {riskConfig.icon}
            <span>风险评估结果</span>
          </Space>
        }
        extra={
          onGenerateReport && (
            <Button
              icon={<DownloadOutlined />}
              onClick={onGenerateReport}
              loading={loading}
              size="small"
            >
              生成报告
            </Button>
          )
        }
        className="risk-summary-card"
      >
        <Alert
          message={
            <Space size="large">
              <span>风险等级: </span>
              <Tag color={riskConfig.color} icon={riskConfig.icon} style={{ fontSize: 14 }}>
                {riskConfig.text}
              </Tag>
              {riskAssessment?.riskScore !== undefined && (
                <span>风险分数: {riskAssessment.riskScore}/100</span>
              )}
            </Space>
          }
          description={
            <div style={{ marginTop: 12 }}>
              {riskAssessment?.requiresApproval && (
                <div style={{ marginBottom: 8 }}>
                  <Tag color="warning">需要合规部门审批</Tag>
                </div>
              )}
              {recommendations && recommendations.length > 0 && (
                <div>
                  <div style={{ fontWeight: 500, marginBottom: 8 }}>处理建议:</div>
                  {recommendations.map((rec, idx) => (
                    <div key={idx} style={{ marginLeft: 16 }}>
                      • {rec}
                    </div>
                  ))}
                </div>
              )}
            </div>
          }
          type={hasConflict ? 'warning' : 'success'}
          showIcon
        />

        {/* 风险分数进度条 */}
        {riskAssessment?.riskScore !== undefined && (
          <div style={{ marginTop: 16 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
              <span>风险分数</span>
              <span>{riskAssessment.riskScore}/100</span>
            </div>
            <Progress
              percent={Math.round(riskAssessment.riskScore)}
              status={hasConflict ? 'exception' : 'success'}
              strokeColor={riskConfig.progressColor}
            />
          </div>
        )}
      </Card>

      {/* 检测统计 */}
      {checkStatistics && (
        <Card title="检测统计" className="stats-card" style={{ marginTop: 16 }}>
          <Descriptions column={2} size="small">
            <Descriptions.Item label="检测案件总数">
              {checkStatistics.totalCasesChecked}
            </Descriptions.Item>
            <Descriptions.Item label="客户历史案件">
              {checkStatistics.clientHistoryCases}
            </Descriptions.Item>
            <Descriptions.Item label="关联方检测">
              {checkStatistics.relatedPartiesChecked}
            </Descriptions.Item>
            <Descriptions.Item label="企业关系检测">
              {checkStatistics.corporateRelationsChecked}
            </Descriptions.Item>
            <Descriptions.Item label="检索时间范围">
              {checkStatistics.timeRange}
            </Descriptions.Item>
            <Descriptions.Item label="检索范围">
              {checkStatistics.searchScope}
            </Descriptions.Item>
          </Descriptions>
        </Card>
      )}

      {/* 冲突案件列表 */}
      {hasConflict && conflictCases && conflictCases.length > 0 && (
        <Card
          title={`冲突案件列表 (${conflictCases.length})`}
          className="conflict-list-card"
          style={{ marginTop: 16 }}
        >
          <Table
            columns={columns}
            dataSource={conflictCases}
            rowKey="caseId"
            pagination={{ pageSize: 10 }}
            size="small"
            className="conflict-table"
          />
        </Card>
      )}

      {/* 无冲突状态 */}
      {!hasConflict && (
        <Card className="no-conflict-card" style={{ marginTop: 16 }}>
          <div style={{ textAlign: 'center', padding: '40px 0' }}>
            <CheckCircleOutlined style={{ fontSize: 64, color: '#52c41a' }} />
            <div style={{ marginTop: 16, fontSize: 16, color: '#52c41a', fontWeight: 500 }}>
              未检测到利益冲突
            </div>
            <div style={{ marginTop: 8, color: '#999' }}>
              可以正常处理此案件
            </div>
          </div>
        </Card>
      )}

      {/* 风险因素详情 */}
      {riskAssessment?.riskFactors && riskAssessment.riskFactors.length > 0 && (
        <Card title="风险因素分析" style={{ marginTop: 16 }}>
          <Space wrap>
            {riskAssessment.riskFactors.map((factor, idx) => (
              <Tag key={idx} color="warning" style={{ marginBottom: 8, padding: '4px 12px' }}>
                {factor}
              </Tag>
            ))}
          </Space>
        </Card>
      )}
    </div>
  )
}

export default ConflictResult
