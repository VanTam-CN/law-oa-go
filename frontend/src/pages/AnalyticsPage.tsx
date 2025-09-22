import React, { useState, useEffect } from 'react';
import {
  Card,
  Button,
  Row,
  Col,
  Table,
  Badge,
  Form,
  InputGroup,
  Tabs,
  Tab,
  Alert,
  ProgressBar,
  Dropdown,
  Modal,
} from 'react-bootstrap';
import {
  FaChartBar,
  FaFileLines,
  FaCalendar,
  FaDollarSign,
  FaUser,
  FaClock,
  FaDownload,
  FaEye,
  FaTrash,
  FaFilter,
  FaRotate,
  FaTriangleExclamation,
  FaCircleInfo
} from 'react-icons/fa6';

interface AnalyticsData {
  caseStats: {
    total: number;
    active: number;
    completed: number;
    successRate: number;
    monthlyTrend: { month: string; cases: number; completed: number }[];
  };
  financialStats: {
    totalRevenue: number;
    pendingRevenue: number;
    totalExpenses: number;
    netIncome: number;
    monthlyRevenue: { month: string; revenue: number; expenses: number }[];
  };
  clientStats: {
    total: number;
    active: number;
    newClients: number;
    topClients: { name: string; cases: number; revenue: number }[];
  };
  lawyerStats: {
    total: number;
    active: number;
    avgCases: number;
    performance: { name: string; cases: number; successRate: number; revenue: number }[];
  };
}

interface Report {
  id: string;
  title: string;
  type: string;
  status: 'pending' | 'generating' | 'completed' | 'failed';
  createdAt: string;
  size?: string;
  format: 'pdf' | 'excel' | 'html';
  description: string;
}

const AnalyticsPage: React.FC = () => {
  const [activeTab, setActiveTab] = useState<string>('overview');
  const [analyticsData, setAnalyticsData] = useState<AnalyticsData | null>(null);
  const [reports, setReports] = useState<Report[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [selectedDateRange, setSelectedDateRange] = useState<string>('30days');
  const [showGenerateModal, setShowGenerateModal] = useState<boolean>(false);

  // 模拟数据
  const mockAnalyticsData: AnalyticsData = {
    caseStats: {
      total: 156,
      active: 89,
      completed: 67,
      successRate: 78.5,
      monthlyTrend: [
        { month: '1月', cases: 12, completed: 10 },
        { month: '2月', cases: 15, completed: 12 },
        { month: '3月', cases: 18, completed: 14 },
        { month: '4月', cases: 22, completed: 17 },
        { month: '5月', cases: 20, completed: 16 },
        { month: '6月', cases: 25, completed: 20 },
      ]
    },
    financialStats: {
      totalRevenue: 2850000,
      pendingRevenue: 850000,
      totalExpenses: 1200000,
      netIncome: 1650000,
      monthlyRevenue: [
        { month: '1月', revenue: 380000, expenses: 180000 },
        { month: '2月', revenue: 420000, expenses: 190000 },
        { month: '3月', revenue: 480000, expenses: 200000 },
        { month: '4月', revenue: 520000, expenses: 210000 },
        { month: '5月', revenue: 550000, expenses: 220000 },
        { month: '6月', revenue: 500000, expenses: 200000 },
      ]
    },
    clientStats: {
      total: 124,
      active: 89,
      newClients: 23,
      topClients: [
        { name: 'ABC科技有限公司', cases: 8, revenue: 450000 },
        { name: '张三', cases: 3, revenue: 150000 },
        { name: 'DEF投资集团', cases: 5, revenue: 380000 },
        { name: '李四', cases: 2, revenue: 120000 },
        { name: 'GHI制造企业', cases: 4, revenue: 280000 }
      ]
    },
    lawyerStats: {
      total: 12,
      active: 10,
      avgCases: 8.9,
      performance: [
        { name: '王律师', cases: 15, successRate: 85, revenue: 580000 },
        { name: '李律师', cases: 12, successRate: 78, revenue: 450000 },
        { name: '赵律师', cases: 10, successRate: 82, revenue: 380000 },
        { name: '钱律师', cases: 8, successRate: 75, revenue: 320000 },
        { name: '孙律师', cases: 7, successRate: 71, revenue: 280000 }
      ]
    }
  };

  const mockReports: Report[] = [
    {
      id: '1',
      title: '2024年上半年案件统计报告',
      type: 'case_summary',
      status: 'completed',
      createdAt: '2024-06-30',
      size: '2.5MB',
      format: 'pdf',
      description: '包含案件类型分布、成功率分析、律师绩效等'
    },
    {
      id: '2',
      title: '月度财务分析报告',
      type: 'financial',
      status: 'completed',
      createdAt: '2024-06-28',
      size: '1.8MB',
      format: 'excel',
      description: '收入支出分析、客户贡献度、费用趋势等'
    },
    {
      id: '3',
      title: '客户满意度调查报告',
      type: 'client_summary',
      status: 'generating',
      createdAt: '2024-06-29',
      format: 'pdf',
      description: '客户反馈分析、服务满意度、改进建议等'
    },
    {
      id: '4',
      title: '律师绩效评估报告',
      type: 'performance',
      status: 'pending',
      createdAt: '2024-06-27',
      format: 'html',
      description: '律师工作负荷、案件成功率、客户评价等'
    }
  ];

  useEffect(() => {
    loadAnalyticsData();
    loadReports();
  }, [selectedDateRange]);

  const loadAnalyticsData = () => {
    setLoading(true);
    // 模拟API调用
    setTimeout(() => {
      setAnalyticsData(mockAnalyticsData);
      setLoading(false);
    }, 1000);
  };

  const loadReports = () => {
    // 模拟API调用
    setReports(mockReports);
  };

  const formatCurrency = (amount: number): string => {
    return `¥${amount.toLocaleString()}`;
  };

  const getReportTypeText = (type: string): string => {
    switch (type) {
      case 'case_summary': return '案件统计';
      case 'financial': return '财务分析';
      case 'client_summary': return '客户分析';
      case 'performance': return '绩效评估';
      case 'activity': return '活动报告';
      default: return type;
    }
  };

  const getReportStatusBadge = (status: Report['status']) => {
    switch (status) {
      case 'completed':
        return <Badge bg="success">已完成</Badge>;
      case 'generating':
        return <Badge bg="warning">生成中</Badge>;
      case 'pending':
        return <Badge bg="info">等待中</Badge>;
      case 'failed':
        return <Badge bg="danger">失败</Badge>;
      default:
        return <Badge bg="secondary">未知</Badge>;
    }
  };

  return (
    <div className="analytics-page p-4">
      {/* 头部 */}
      <Card className="mb-4">
        <Card.Header>
          <div className="d-flex justify-content-between align-items-center">
            <div>
              <h4 className="mb-0">数据分析与报告</h4>
              <p className="text-muted mb-0">律所运营数据分析和报告生成</p>
            </div>
            <div className="d-flex gap-2">
              <Form.Select
                value={selectedDateRange}
                onChange={(e) => setSelectedDateRange(e.target.value)}
                style={{ width: '120px' }}
              >
                <option value="7days">最近7天</option>
                <option value="30days">最近30天</option>
                <option value="90days">最近90天</option>
                <option value="1year">最近1年</option>
              </Form.Select>
              <Button variant="outline-secondary" onClick={loadAnalyticsData}>
                <FaRotate className="w-4 h-4 me-1" />
                刷新
              </Button>
              <Button variant="primary" onClick={() => setShowGenerateModal(true)}>
                <span className="w-4 h-4 me-1">➕</span>
                生成报告
              </Button>
            </div>
          </div>
        </Card.Header>
      </Card>

      {/* 主要内容 */}
      <Tabs
        activeKey={activeTab}
        onSelect={(key) => setActiveTab(key || 'overview')}
        className="mb-4"
      >
        {/* 概览标签页 */}
        <Tab
          eventKey="overview"
          title={
            <>
              <FaChartBar className="w-4 h-4 me-2" />
              数据概览
            </>
          }
        >
          <Row className="mb-4">
            <Col md={3}>
              <Card className="text-center">
                <Card.Body>
                  <h3 className="text-primary">{analyticsData?.caseStats.total || 0}</h3>
                  <p className="text-muted mb-0">总案件数</p>
                  <small className="text-success">
                    <FaChartBar className="w-3 h-3" />
                    本月+{analyticsData?.caseStats.monthlyTrend[5]?.cases || 0}
                  </small>
                </Card.Body>
              </Card>
            </Col>
            <Col md={3}>
              <Card className="text-center">
                <Card.Body>
                  <h3 className="text-success">{formatCurrency(analyticsData?.financialStats.totalRevenue || 0)}</h3>
                  <p className="text-muted mb-0">总收入</p>
                  <small className="text-success">
                    <FaChartBar className="w-3 h-3" />
                    净利润{formatCurrency(analyticsData?.financialStats.netIncome || 0)}
                  </small>
                </Card.Body>
              </Card>
            </Col>
            <Col md={3}>
              <Card className="text-center">
                <Card.Body>
                  <h3 className="text-info">{analyticsData?.clientStats.total || 0}</h3>
                  <p className="text-muted mb-0">客户总数</p>
                  <small className="text-success">
                    <FaChartBar className="w-3 h-3" />
                    本月新增{analyticsData?.clientStats.newClients || 0}
                  </small>
                </Card.Body>
              </Card>
            </Col>
            <Col md={3}>
              <Card className="text-center">
                <Card.Body>
                  <h3 className="text-warning">{(analyticsData?.caseStats.successRate || 0).toFixed(1)}%</h3>
                  <p className="text-muted mb-0">案件成功率</p>
                  <small className="text-muted">行业领先水平</small>
                </Card.Body>
              </Card>
            </Col>
          </Row>

          <Row className="mb-4">
            <Col md={6}>
              <Card>
                <Card.Header>
                  <h6 className="mb-0">案件趋势</h6>
                </Card.Header>
                <Card.Body>
                  <div className="table-responsive">
                    <Table size="sm">
                      <thead>
                        <tr>
                          <th>月份</th>
                          <th>新增案件</th>
                          <th>完成案件</th>
                          <th>完成率</th>
                        </tr>
                      </thead>
                      <tbody>
                        {analyticsData?.caseStats.monthlyTrend.map((trend, index) => (
                          <tr key={index}>
                            <td>{trend.month}</td>
                            <td>{trend.cases}</td>
                            <td>{trend.completed}</td>
                            <td>{((trend.completed / trend.cases) * 100).toFixed(1)}%</td>
                          </tr>
                        ))}
                      </tbody>
                    </Table>
                  </div>
                </Card.Body>
              </Card>
            </Col>
            <Col md={6}>
              <Card>
                <Card.Header>
                  <h6 className="mb-0">财务趋势</h6>
                </Card.Header>
                <Card.Body>
                  <div className="table-responsive">
                    <Table size="sm">
                      <thead>
                        <tr>
                          <th>月份</th>
                          <th>收入</th>
                          <th>支出</th>
                          <th>净利润</th>
                        </tr>
                      </thead>
                      <tbody>
                        {analyticsData?.financialStats.monthlyRevenue.map((revenue, index) => (
                          <tr key={index}>
                            <td>{revenue.month}</td>
                            <td>{formatCurrency(revenue.revenue)}</td>
                            <td>{formatCurrency(revenue.expenses)}</td>
                            <td className={revenue.revenue - revenue.expenses > 0 ? 'text-success' : 'text-danger'}>
                              {formatCurrency(revenue.revenue - revenue.expenses)}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </Table>
                  </div>
                </Card.Body>
              </Card>
            </Col>
          </Row>
        </Tab>

        {/* 案件分析标签页 */}
        <Tab
          eventKey="cases"
          title={
            <>
              <FaFileLines className="w-4 h-4 me-2" />
              案件分析
            </>
          }
        >
          <Row className="mb-4">
            <Col md={8}>
              <Card>
                <Card.Header>
                  <h6 className="mb-0">案件类型分布</h6>
                </Card.Header>
                <Card.Body>
                  <div className="text-center py-5">
                    <FaChartBar className="w-16 h-16 text-muted mx-auto mb-3" />
                    <h5>图表功能开发中...</h5>
                    <p className="text-muted">将展示案件类型分布图、案件趋势图等</p>
                  </div>
                </Card.Body>
              </Card>
            </Col>
            <Col md={4}>
              <Card>
                <Card.Header>
                  <h6 className="mb-0">案件状态</h6>
                </Card.Header>
                <Card.Body>
                  <div className="g">
                    <div>
                      <div className="d-flex justify-content-between mb-1">
                        <span>进行中</span>
                        <span>{analyticsData?.caseStats.active}</span>
                      </div>
                      <ProgressBar now={57} variant="primary" />
                    </div>
                    <div>
                      <div className="d-flex justify-content-between mb-1">
                        <span>已完成</span>
                        <span>{analyticsData?.caseStats.completed}</span>
                      </div>
                      <ProgressBar now={43} variant="success" />
                    </div>
                  </div>
                </Card.Body>
              </Card>
            </Col>
          </Row>
        </Tab>

        {/* 财务分析标签页 */}
        <Tab
          eventKey="financial"
          title={
            <>
              <FaDollarSign className="w-4 h-4 me-2" />
              财务分析
            </>
          }
        >
          <Row className="mb-4">
            <Col md={6}>
              <Card>
                <Card.Header>
                  <h6 className="mb-0">收入构成</h6>
                </Card.Header>
                <Card.Body>
                  <div className="text-center py-5">
                    <FaDollarSign className="w-16 h-16 text-muted mx-auto mb-3" />
                    <h5>图表功能开发中...</h5>
                    <p className="text-muted">将展示收入构成饼图、收入趋势图等</p>
                  </div>
                </Card.Body>
              </Card>
            </Col>
            <Col md={6}>
              <Card>
                <Card.Header>
                  <h6 className="mb-0">财务摘要</h6>
                </Card.Header>
                <Card.Body>
                  <div className="g">
                    <div className="d-flex justify-content-between">
                      <span>总收入：</span>
                      <strong className="text-success">{formatCurrency(analyticsData?.financialStats.totalRevenue || 0)}</strong>
                    </div>
                    <div className="d-flex justify-content-between">
                      <span>待收款项：</span>
                      <strong className="text-warning">{formatCurrency(analyticsData?.financialStats.pendingRevenue || 0)}</strong>
                    </div>
                    <div className="d-flex justify-content-between">
                      <span>总支出：</span>
                      <strong className="text-danger">{formatCurrency(analyticsData?.financialStats.totalExpenses || 0)}</strong>
                    </div>
                    <hr />
                    <div className="d-flex justify-content-between">
                      <span>净利润：</span>
                      <strong className={(analyticsData?.financialStats.netIncome || 0) > 0 ? 'text-success' : 'text-danger'}>
                        {formatCurrency(analyticsData?.financialStats.netIncome || 0)}
                      </strong>
                    </div>
                  </div>
                </Card.Body>
              </Card>
            </Col>
          </Row>
        </Tab>

        {/* 客户分析标签页 */}
        <Tab
          eventKey="clients"
          title={
            <>
              <FaUser className="w-4 h-4 me-2" />
              客户分析
            </>
          }
        >
          <Card>
            <Card.Header>
              <h6 className="mb-0">重点客户排行</h6>
            </Card.Header>
            <Card.Body>
              <div className="table-responsive">
                <Table>
                  <thead>
                    <tr>
                      <th>客户名称</th>
                      <th>案件数量</th>
                      <th>贡献收入</th>
                      <th>占比</th>
                    </tr>
                  </thead>
                  <tbody>
                    {analyticsData?.clientStats.topClients.map((client, index) => (
                      <tr key={index}>
                        <td>{client.name}</td>
                        <td>{client.cases}</td>
                        <td>{formatCurrency(client.revenue)}</td>
                        <td>
                          {((client.revenue / analyticsData.financialStats.totalRevenue) * 100).toFixed(1)}%
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </Table>
              </div>
            </Card.Body>
          </Card>
        </Tab>

        {/* 律师绩效标签页 */}
        <Tab
          eventKey="lawyers"
          title={
            <>
              <FaCalendar className="w-4 h-4 me-2" />
              律师绩效
            </>
          }
        >
          <Card>
            <Card.Header>
              <h6 className="mb-0">律师绩效排行</h6>
            </Card.Header>
            <Card.Body>
              <div className="table-responsive">
                <Table>
                  <thead>
                    <tr>
                      <th>律师姓名</th>
                      <th>负责案件</th>
                      <th>成功率</th>
                      <th>创收金额</th>
                      <th>状态</th>
                    </tr>
                  </thead>
                  <tbody>
                    {analyticsData?.lawyerStats.performance.map((lawyer, index) => (
                      <tr key={index}>
                        <td>
                          <div className="d-flex align-items-center">
                            <FaUser className="w-4 h-4 me-2" />
                            {lawyer.name}
                          </div>
                        </td>
                        <td>{lawyer.cases}</td>
                        <td>
                          <Badge bg={lawyer.successRate > 80 ? 'success' : lawyer.successRate > 70 ? 'warning' : 'danger'}>
                            {lawyer.successRate}%
                          </Badge>
                        </td>
                        <td>{formatCurrency(lawyer.revenue)}</td>
                        <td>
                          <Badge bg="success">活跃</Badge>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </Table>
              </div>
            </Card.Body>
          </Card>
        </Tab>

        {/* 报告管理标签页 */}
        <Tab
          eventKey="reports"
          title={
            <>
              <FaDownload className="w-4 h-4 me-2" />
              报告管理
            </>
          }
        >
          <Card>
            <Card.Header>
              <div className="d-flex justify-content-between align-items-center">
                <h6 className="mb-0">生成报告列表</h6>
                <Button variant="primary" size="sm" onClick={() => setShowGenerateModal(true)}>
                  <span className="w-4 h-4 me-1">➕</span>
                  新建报告
                </Button>
              </div>
            </Card.Header>
            <Card.Body>
              <div className="table-responsive">
                <Table>
                  <thead>
                    <tr>
                      <th>报告名称</th>
                      <th>类型</th>
                      <th>状态</th>
                      <th>生成时间</th>
                      <th>大小</th>
                      <th>操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    {reports.map((report) => (
                      <tr key={report.id}>
                        <td>
                          <div>
                            <div className="fw-bold">{report.title}</div>
                            <div className="small text-muted">{report.description}</div>
                          </div>
                        </td>
                        <td>
                          <Badge bg="light" text="dark">
                            {getReportTypeText(report.type)}
                          </Badge>
                        </td>
                        <td>{getReportStatusBadge(report.status)}</td>
                        <td>{report.createdAt}</td>
                        <td>{report.size || '-'}</td>
                        <td>
                          <div className="btn-group" role="group">
                            <Button
                              variant="outline-primary"
                              size="sm"
                              disabled={report.status !== 'completed'}
                              onClick={() => alert(`下载报告：${report.title}`)}
                            >
                              <FaDownload className="w-4 h-4" />
                            </Button>
                            <Button
                              variant="outline-info"
                              size="sm"
                              disabled={report.status !== 'completed'}
                              onClick={() => alert(`预览报告：${report.title}`)}
                            >
                              <FaEye className="w-4 h-4" />
                            </Button>
                            <Button
                              variant="outline-danger"
                              size="sm"
                              onClick={() => {
                                if (window.confirm('确定要删除这个报告吗？')) {
                                  setReports(reports.filter(r => r.id !== report.id));
                                }
                              }}
                            >
                              <FaTrash className="w-4 h-4" />
                            </Button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </Table>
              </div>
            </Card.Body>
          </Card>
        </Tab>
      </Tabs>

      {/* 生成报告模态框 */}
      <Modal show={showGenerateModal} onHide={() => setShowGenerateModal(false)} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>生成新报告</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <Alert variant="info">
            <span className="me-2">ℹ️</span>
            <strong>提示：</strong>
            报告生成可能需要几分钟时间，生成完成后会自动出现在报告列表中。
          </Alert>

          <Form>
            <Row className="mb-3">
              <Col md={6}>
                <Form.Group>
                  <Form.Label>报告类型</Form.Label>
                  <Form.Select>
                    <option value="case_summary">案件统计报告</option>
                    <option value="financial">财务分析报告</option>
                    <option value="client_summary">客户分析报告</option>
                    <option value="performance">律师绩效报告</option>
                    <option value="activity">活动统计报告</option>
                  </Form.Select>
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group>
                  <Form.Label>输出格式</Form.Label>
                  <Form.Select>
                    <option value="pdf">PDF</option>
                    <option value="excel">Excel</option>
                    <option value="html">HTML</option>
                  </Form.Select>
                </Form.Group>
              </Col>
            </Row>

            <Row className="mb-3">
              <Col md={12}>
                <Form.Group>
                  <Form.Label>报告标题</Form.Label>
                  <Form.Control
                    type="text"
                    placeholder="请输入报告标题"
                  />
                </Form.Group>
              </Col>
            </Row>

            <Row className="mb-3">
              <Col md={6}>
                <Form.Group>
                  <Form.Label>开始日期</Form.Label>
                  <Form.Control type="date" />
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group>
                  <Form.Label>结束日期</Form.Label>
                  <Form.Control type="date" />
                </Form.Group>
              </Col>
            </Row>
          </Form>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => setShowGenerateModal(false)}>
            取消
          </Button>
          <Button variant="primary">
            开始生成
          </Button>
        </Modal.Footer>
      </Modal>
    </div>
  );
};

export default AnalyticsPage;