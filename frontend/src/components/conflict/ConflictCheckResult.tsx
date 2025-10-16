import React, { useState } from 'react';
import { Alert, Card, Tag, Typography, List, Button, Space, Divider, Timeline, Progress } from 'antd';
import {
  WarningOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  InfoCircleOutlined,
  ClockCircleOutlined,
  UserOutlined,
  FileTextOutlined,
  TeamOutlined,
  BankOutlined,
  EyeOutlined,
  RightOutlined
} from '@ant-design/icons';
import ConflictCaseDetail from './ConflictCaseDetail';

const { Title, Paragraph, Text } = Typography;

// 冲突案例接口
interface ConflictCase {
  id: string;
  caseId: string;
  caseName: string;
  caseNo?: string;
  conflictType: string;
  riskLevel: string;
  description: string;
  caseStatus: string;
  clientId: string;
  clientName: string;
  opposingParties: string[];
  conflictDetails: string;
  createdAt: string;
  lawyerName?: string;
  lawyerId?: string;
}

// 风险评估接口
interface RiskAssessment {
  overallRisk: string;
  riskScore: number;
  riskReason: string;
  requiresApproval: boolean;
  riskFactors: string[];
  mitigation: string[];
}

// 检查统计接口
interface CheckStatistics {
  totalCasesChecked: number;
  clientHistoryCases: number;
  relatedPartiesChecked: number;
  corporateRelationsChecked: number;
  timeRange: string;
  searchScope: string;
  startTime: string;
  endTime: string;
}

// 冲突检查结果属性
interface ConflictCheckResultProps {
  checkId: string;
  hasConflict: boolean;
  conflictCases: ConflictCase[];
  checkStatistics: CheckStatistics;
  riskAssessment: RiskAssessment;
  recommendations: string[];
  checkTime: string;
  duration: number;
  onConfirm?: () => void;
  onRetry?: () => void;
}

// 获取风险等级颜色
const getRiskLevelColor = (level: string) => {
  switch (level.toUpperCase()) {
    case 'CRITICAL': return '#dc3545';
    case 'HIGH': return '#fd7e14';
    case 'MEDIUM': return '#ffc107';
    case 'LOW': return '#28a745';
    case 'MINIMAL': return '#17a2b8';
    default: return '#6c757d';
  }
};

// 获取风险等级标签
const getRiskLevelTag = (level: string) => {
  const color = getRiskLevelColor(level);
  return (
    <Tag color={color} style={{ color: 'white', fontWeight: 'bold' }}>
      {level.toUpperCase()}
    </Tag>
  );
};

// 获取冲突类型图标
const getConflictTypeIcon = (type: string) => {
  switch (type) {
    case '代理冲突': return <UserOutlined style={{ color: '#fd7e14' }} />;
    case '当事人冲突': return <TeamOutlined style={{ color: '#dc3545' }} />;
    case '利益关联冲突': return <ExclamationCircleOutlined style={{ color: '#ffc107' }} />;
    default: return <WarningOutlined style={{ color: '#6c757d' }} />;
  }
};

// 冲突检查结果展示组件
const ConflictCheckResult: React.FC<ConflictCheckResultProps> = ({
  checkId,
  hasConflict,
  conflictCases,
  checkStatistics,
  riskAssessment,
  recommendations,
  checkTime,
  duration,
  onConfirm,
  onRetry
}) => {
  // 状态管理
  const [selectedCase, setSelectedCase] = useState<ConflictCase | null>(null);
  const [detailVisible, setDetailVisible] = useState(false);

  // 格式化时间
  const formatTime = (timeString: string) => {
    return new Date(timeString).toLocaleString('zh-CN');
  };

  // 格式化持续时间
  const formatDuration = (ms: number) => {
    return `${ms}ms`;
  };

  // 查看案例详情
  const handleViewCaseDetail = (conflictCase: ConflictCase) => {
    setSelectedCase(conflictCase);
    setDetailVisible(true);
  };

  // 关闭详情弹窗
  const handleCloseDetail = () => {
    setDetailVisible(false);
    setSelectedCase(null);
  };

  // 跳转到案件详情
  const handleNavigateCase = (caseId: string) => {
    // 这里可以实现路由跳转逻辑
    console.log('跳转到案件详情:', caseId);
    // 可以使用 navigate 函数或者其他路由方式
  };

  return (
    <div style={{ maxWidth: 1200, margin: '0 auto', padding: '20px' }}>
      {/* 头部状态展示 */}
      <Card style={{ marginBottom: '20px' }}>
        <div style={{ textAlign: 'center', padding: '20px 0' }}>
          {hasConflict ? (
            <>
              <WarningOutlined style={{ fontSize: '48px', color: '#fd7e14', marginBottom: '16px' }} />
              <Title level={3} style={{ color: '#fd7e14', margin: '16px 0 8px 0' }}>
                发现潜在利益冲突
              </Title>
              <Paragraph type="secondary">
                系统检测到 {conflictCases.length} 个潜在冲突案例，请仔细查看详情
              </Paragraph>
            </>
          ) : (
            <>
              <CheckCircleOutlined style={{ fontSize: '48px', color: '#28a745', marginBottom: '16px' }} />
              <Title level={3} style={{ color: '#28a745', margin: '16px 0 8px 0' }}>
                未发现明显冲突
              </Title>
              <Paragraph type="secondary">
                经过全面检测，未发现明显的利益冲突问题
              </Paragraph>
              <div style={{ backgroundColor: '#f6ffed', padding: '12px', borderRadius: '6px', marginTop: '12px' }}>
                <Text type="secondary" style={{ fontSize: '12px' }}>
                  <strong>检查说明：</strong>
                  <br />• 系统已搜索了同一律师代理的所有历史案件
                  <br />• 未发现与当前案件存在利益冲突的案例
                  <br />• 该律师可以安全代理此案件
                  <br />• 建议定期进行利益冲突检查以确保合规性
                </Text>
              </div>
            </>
          )}
        </div>
      </Card>

      {/* 检查摘要信息 */}
      <Card title="检查摘要" style={{ marginBottom: '20px' }}>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '16px' }}>
          <div>
            <Text type="secondary">检查ID</Text>
            <div><Text strong>{checkId}</Text></div>
          </div>
          <div>
            <Text type="secondary">检查时间</Text>
            <div><Text strong>{formatTime(checkTime)}</Text></div>
          </div>
          <div>
            <Text type="secondary">检查耗时</Text>
            <div><Text strong>{formatDuration(duration)}</Text></div>
          </div>
          <div>
            <Text type="secondary">搜索范围</Text>
            <div><Text strong>{checkStatistics.timeRange} · {checkStatistics.searchScope}</Text></div>
          </div>
        </div>
      </Card>

      {/* 风险评估结果 */}
      <Card title="风险评估" style={{ marginBottom: '20px' }}>
        <div style={{ marginBottom: '16px' }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '8px' }}>
            <Text strong>综合风险等级</Text>
            {getRiskLevelTag(riskAssessment.overallRisk)}
          </div>
          <Progress
            percent={Math.round(riskAssessment.riskScore * 100)}
            strokeColor={getRiskLevelColor(riskAssessment.overallRisk)}
            showInfo={false}
          />
          <div style={{ textAlign: 'center', marginTop: '8px' }}>
            <Text type="secondary">
              风险评分: {(riskAssessment.riskScore * 100).toFixed(1)}/100
            </Text>
          </div>
        </div>

        <Divider />

        <div style={{ marginBottom: '16px' }}>
          <Text strong>风险原因</Text>
          <Paragraph style={{ marginTop: '8px' }}>
            {riskAssessment.riskReason}
          </Paragraph>
        </div>

        {riskAssessment.riskFactors.length > 0 && (
          <div style={{ marginBottom: '16px' }}>
            <Text strong>风险因素</Text>
            <div style={{ marginTop: '8px' }}>
              {riskAssessment.riskFactors.map((factor, index) => (
                <Tag key={index} style={{ margin: '4px 4px 4px 0' }}>
                  {factor}
                </Tag>
              ))}
            </div>
          </div>
        )}

        {riskAssessment.requiresApproval && (
          <Alert
            message="需要审批"
            description="此案件的风险等级需要提交给合规部门审批"
            type="warning"
            showIcon
            style={{ marginTop: '16px' }}
          />
        )}
      </Card>

      {/* 冲突案例详情 */}
      {hasConflict && conflictCases.length > 0 && (
        <Card title={`冲突案例详情 (${conflictCases.length}个)`} style={{ marginBottom: '20px' }}>
          <List
            dataSource={conflictCases}
            renderItem={(conflictCase) => (
              <List.Item key={conflictCase.id}>
                <Card size="small" style={{ width: '100%' }}>
                  <div style={{ display: 'flex', alignItems: 'flex-start', gap: '12px' }}>
                    <div style={{ marginTop: '4px' }}>
                      {getConflictTypeIcon(conflictCase.conflictType)}
                    </div>
                    <div style={{ flex: 1 }}>
                      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '8px' }}>
                        <Text strong style={{ fontSize: '16px' }}>
                          {conflictCase.caseName}
                        </Text>
                        <Space>
                          {getRiskLevelTag(conflictCase.riskLevel)}
                          <Tag color="blue">{conflictCase.conflictType}</Tag>
                        </Space>
                      </div>

                      {conflictCase.caseNo && (
                        <div style={{ marginBottom: '8px' }}>
                          <Text type="secondary">案件编号: </Text>
                          <Text>{conflictCase.caseNo}</Text>
                        </div>
                      )}

                      <div style={{ marginBottom: '8px' }}>
                        <Text type="secondary">冲突详情: </Text>
                        <Text>{conflictCase.conflictDetails || conflictCase.description}</Text>
                      </div>

                      {conflictCase.opposingParties && conflictCase.opposingParties.length > 0 && (
                        <div style={{ marginBottom: '8px' }}>
                          <Text type="secondary">对方当事人: </Text>
                          <div style={{ marginTop: '4px' }}>
                            {conflictCase.opposingParties.map((party, index) => (
                              <Tag key={index} style={{ margin: '2px 4px 2px 0' }}>
                                {party}
                              </Tag>
                            ))}
                          </div>
                        </div>
                      )}

                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <Text type="secondary" style={{ fontSize: '12px' }}>
                          <ClockCircleOutlined style={{ marginRight: '4px' }} />
                          创建时间: {formatTime(conflictCase.createdAt)}
                        </Text>
                        <Space>
                          <Text type="secondary" style={{ fontSize: '12px' }}>
                            状态: <Tag size="small" color="green">{conflictCase.caseStatus}</Tag>
                          </Text>
                          <Button
                            type="link"
                            size="small"
                            icon={<EyeOutlined />}
                            onClick={() => handleViewCaseDetail(conflictCase)}
                          >
                            查看详情
                          </Button>
                        </Space>
                      </div>
                    </div>
                  </div>
                </Card>
              </List.Item>
            )}
          />
        </Card>
      )}

      {/* 检查统计信息 */}
      <Card title="检查统计" style={{ marginBottom: '20px' }}>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(150px, 1fr))', gap: '16px' }}>
          <div style={{ textAlign: 'center', padding: '16px', backgroundColor: '#f8f9fa', borderRadius: '8px' }}>
            <div style={{ fontSize: '24px', fontWeight: 'bold', color: '#1890ff' }}>
              {checkStatistics.totalCasesChecked}
            </div>
            <div style={{ color: '#666', marginTop: '4px' }}>总检查案件</div>
          </div>
          <div style={{ textAlign: 'center', padding: '16px', backgroundColor: '#f8f9fa', borderRadius: '8px' }}>
            <div style={{ fontSize: '24px', fontWeight: 'bold', color: '#52c41a' }}>
              {checkStatistics.clientHistoryCases}
            </div>
            <div style={{ color: '#666', marginTop: '4px' }}>客户历史案件</div>
          </div>
          <div style={{ textAlign: 'center', padding: '16px', backgroundColor: '#f8f9fa', borderRadius: '8px' }}>
            <div style={{ fontSize: '24px', fontWeight: 'bold', color: '#faad14' }}>
              {checkStatistics.relatedPartiesChecked}
            </div>
            <div style={{ color: '#666', marginTop: '4px' }}>关联方检查</div>
          </div>
          <div style={{ textAlign: 'center', padding: '16px', backgroundColor: '#f8f9fa', borderRadius: '8px' }}>
            <div style={{ fontSize: '24px', fontWeight: 'bold', color: '#722ed1' }}>
              {checkStatistics.corporateRelationsChecked}
            </div>
            <div style={{ color: '#666', marginTop: '4px' }}>企业关系检查</div>
          </div>
        </div>
      </Card>

      {/* 处理建议 */}
      {recommendations.length > 0 && (
        <Card title="处理建议" style={{ marginBottom: '20px' }}>
          <Timeline>
            {recommendations.map((recommendation, index) => (
              <Timeline.Item
                key={index}
                dot={<InfoCircleOutlined style={{ fontSize: '16px' }} />}
                color="blue"
              >
                <Text>{recommendation}</Text>
              </Timeline.Item>
            ))}
          </Timeline>
        </Card>
      )}

      {/* 操作按钮 */}
      <Card style={{ textAlign: 'center', backgroundColor: '#fafafa' }}>
        <Space size="large">
          {hasConflict && riskAssessment.requiresApproval && (
            <Alert
              message="此案件需要合规部门审批后才能继续"
              type="warning"
              showIcon
              style={{ marginBottom: '16px' }}
            />
          )}

          {!riskAssessment.requiresApproval && (
            <Button
              type="primary"
              size="large"
              icon={<CheckCircleOutlined />}
              onClick={onConfirm}
              style={{ minWidth: '120px' }}
            >
              确认继续
            </Button>
          )}

          {riskAssessment.requiresApproval && (
            <Button
              type="primary"
              size="large"
              icon={<ExclamationCircleOutlined />}
              onClick={onConfirm}
              disabled
              style={{ minWidth: '120px' }}
            >
              待审批
            </Button>
          )}

          <Button
            size="large"
            icon={<ClockCircleOutlined />}
            onClick={onRetry}
          >
            重新检查
          </Button>
        </Space>

        {hasConflict && (
          <div style={{ marginTop: '16px' }}>
            <Text type="secondary" style={{ fontSize: '12px' }}>
              请仔细查看上述冲突案例详情，确认是否继续处理此案件。
              如有疑问，请联系合规部门。
            </Text>
          </div>
        )}
      </Card>

      {/* 冲突案例详情弹窗 */}
      <ConflictCaseDetail
        conflictCase={selectedCase}
        visible={detailVisible}
        onClose={handleCloseDetail}
        onNavigateCase={handleNavigateCase}
      />
    </div>
  );
};

export default ConflictCheckResult;