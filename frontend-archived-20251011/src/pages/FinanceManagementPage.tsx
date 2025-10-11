import React, { useState, useEffect } from 'react';
import {
  FaPlus,
  FaPenToSquare,
  FaTrash,
  FaMagnifyingGlass,
  FaDollarSign,
  FaCreditCard,
  FaChartBar,
  FaFileLines,
  FaCheck,
  FaTriangleExclamation,
  FaClock,
  FaCalendar,
  FaUser,
  FaFilter
} from 'react-icons/fa6';
import {
  Row,
  Col,
  Card,
  Table,
  Button,
  Form,
  InputGroup,
  Badge,
  Tabs,
  Tab,
  Alert
} from 'react-bootstrap';

interface Invoice {
  id: string;
  invoiceNumber: string;
  clientName: string;
  projectName: string;
  amount: number;
  status: 'pending' | 'paid' | 'overdue';
  issueDate: string;
  dueDate: string;
  paidDate?: string;
}

interface Expense {
  id: string;
  expenseNumber: string;
  category: string;
  description: string;
  amount: number;
  date: string;
  status: 'pending' | 'approved' | 'rejected';
  applicant: string;
  approveDate?: string;
  approver?: string;
}

const FinanceManagement: React.FC = () => {
  const [activeTab, setActiveTab] = useState<string>('overview');
  const [loading, setLoading] = useState<boolean>(false);

  // 模拟数据
  const [invoices, setInvoices] = useState<Invoice[]>([
    {
      id: '1',
      invoiceNumber: 'INV-2024-001',
      clientName: '张三',
      projectName: '借款纠纷案',
      amount: 50000,
      status: 'paid',
      issueDate: '2024-01-01',
      dueDate: '2024-01-15',
      paidDate: '2024-01-10'
    },
    {
      id: '2',
      invoiceNumber: 'INV-2024-002',
      clientName: 'ABC公司',
      projectName: '合同审查',
      amount: 30000,
      status: 'pending',
      issueDate: '2024-01-05',
      dueDate: '2024-01-20'
    },
    {
      id: '3',
      invoiceNumber: 'INV-2024-003',
      clientName: '科技公司',
      projectName: '知识产权咨询',
      amount: 25000,
      status: 'overdue',
      issueDate: '2023-12-15',
      dueDate: '2024-01-01'
    },
    {
      id: '4',
      invoiceNumber: 'INV-2024-004',
      clientName: '投资公司',
      projectName: '企业并购',
      amount: 200000,
      status: 'pending',
      issueDate: '2024-01-10',
      dueDate: '2024-01-25'
    }
  ]);

  const [expenses, setExpenses] = useState<Expense[]>([
    {
      id: '1',
      expenseNumber: 'EXP-2024-001',
      category: '办公耗材',
      description: '打印纸、文具等办公用品',
      amount: 1500,
      date: '2024-01-05',
      status: 'approved',
      applicant: '行政部',
      approveDate: '2024-01-06',
      approver: '财务部'
    },
    {
      id: '2',
      expenseNumber: 'EXP-2024-002',
      category: '差旅费用',
      description: '北京出差交通住宿费用',
      amount: 3500,
      date: '2024-01-08',
      status: 'pending',
      applicant: '王律师'
    },
    {
      id: '3',
      expenseNumber: 'EXP-2024-003',
      category: '培训费用',
      description: '法律培训课程费用',
      amount: 8000,
      date: '2024-01-10',
      status: 'rejected',
      applicant: '人事部',
      approveDate: '2024-01-12',
      approver: '财务部'
    },
    {
      id: '4',
      expenseNumber: 'EXP-2024-004',
      category: '软件费用',
      description: '法律数据库订阅费用',
      amount: 12000,
      date: '2024-01-12',
      status: 'pending',
      applicant: 'IT部'
    }
  ]);

  const [searchTerm, setSearchTerm] = useState<string>('');
  const [filterStatus, setFilterStatus] = useState<string>('all');

  // 计算统计数据
  const calculateStats = () => {
    const totalRevenue = invoices
      .filter(inv => inv.status === 'paid')
      .reduce((sum, inv) => sum + inv.amount, 0);

    const pendingRevenue = invoices
      .filter(inv => inv.status === 'pending')
      .reduce((sum, inv) => sum + inv.amount, 0);

    const overdueRevenue = invoices
      .filter(inv => inv.status === 'overdue')
      .reduce((sum, inv) => sum + inv.amount, 0);

    const totalExpenses = expenses
      .filter(exp => exp.status === 'approved')
      .reduce((sum, exp) => sum + exp.amount, 0);

    return {
      totalRevenue,
      pendingRevenue,
      overdueRevenue,
      totalExpenses,
      netIncome: totalRevenue - totalExpenses
    };
  };

  const stats = calculateStats();

  const getInvoiceStatusBadge = (status: Invoice['status']) => {
    switch (status) {
      case 'paid':
        return <Badge bg="success">已付款</Badge>;
      case 'pending':
        return <Badge bg="warning">待付款</Badge>;
      case 'overdue':
        return <Badge bg="danger">逾期</Badge>;
      default:
        return <Badge bg="secondary">未知</Badge>;
    }
  };

  const getExpenseStatusBadge = (status: Expense['status']) => {
    switch (status) {
      case 'approved':
        return <Badge bg="success">已批准</Badge>;
      case 'pending':
        return <Badge bg="warning">待审批</Badge>;
      case 'rejected':
        return <Badge bg="danger">已拒绝</Badge>;
      default:
        return <Badge bg="secondary">未知</Badge>;
    }
  };

  const formatCurrency = (amount: number): string => {
    return `¥${amount.toLocaleString()}`;
  };

  const filteredInvoices = invoices.filter(inv => {
    const matchesSearch = inv.clientName.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         inv.projectName.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesStatus = filterStatus === 'all' || inv.status === filterStatus;
    return matchesSearch && matchesStatus;
  });

  const filteredExpenses = expenses.filter(exp => {
    const matchesSearch = exp.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         exp.category.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesStatus = filterStatus === 'all' || exp.status === filterStatus;
    return matchesSearch && matchesStatus;
  });

  return (
    <div className="finance-management p-4">
      <Card className="mb-4">
        <Card.Header>
          <div className="d-flex justify-content-between align-items-center">
            <div>
              <h4 className="mb-0">财务管理</h4>
              <p className="text-muted mb-0">管理律所的收入、支出和财务报表</p>
            </div>
            <Badge bg="primary">
              <FaDollarSign className="w-4 h-4 me-1" />
              财务
            </Badge>
          </div>
        </Card.Header>
        <Card.Body>
          <Tabs
            activeKey={activeTab}
            onSelect={(key) => setActiveTab(key || 'overview')}
            className="mb-3"
          >
            <Tab
              eventKey="overview"
              title={
                <>
                  <FaChartBar className="w-4 h-4 me-2" />
                  财务概览
                </>
              }
            >
              {/* 财务概览 */}
              <Row className="mb-4">
                <Col md={3}>
                  <Card className="text-center">
                    <Card.Body>
                      <h3 className="text-success">{formatCurrency(stats.totalRevenue)}</h3>
                      <p className="text-muted mb-0">已收入金额</p>
                    </Card.Body>
                  </Card>
                </Col>
                <Col md={3}>
                  <Card className="text-center">
                    <Card.Body>
                      <h3 className="text-warning">{formatCurrency(stats.pendingRevenue)}</h3>
                      <p className="text-muted mb-0">待收入金额</p>
                    </Card.Body>
                  </Card>
                </Col>
                <Col md={3}>
                  <Card className="text-center">
                    <Card.Body>
                      <h3 className="text-danger">{formatCurrency(stats.overdueRevenue)}</h3>
                      <p className="text-muted mb-0">逾期金额</p>
                    </Card.Body>
                  </Card>
                </Col>
                <Col md={3}>
                  <Card className="text-center">
                    <Card.Body>
                      <h3 className="text-info">{formatCurrency(stats.totalExpenses)}</h3>
                      <p className="text-muted mb-0">总支出</p>
                    </Card.Body>
                  </Card>
                </Col>
              </Row>

              <Card className="mb-4">
                <Card.Header>
                  <h5 className="mb-0">净收入分析</h5>
                </Card.Header>
                <Card.Body>
                  <div className="text-center">
                    <h2 className={stats.netIncome >= 0 ? 'text-success' : 'text-danger'}>
                      {formatCurrency(stats.netIncome)}
                    </h2>
                    <p className="text-muted">
                      {stats.netIncome >= 0 ? '净利润' : '净亏损'}
                    </p>
                  </div>
                </Card.Body>
              </Card>

              <Card>
                <Card.Header>
                  <h5 className="mb-0">财务趋势图</h5>
                </Card.Header>
                <Card.Body>
                  <div className="text-center py-5">
                    <ChartBarIcon className="w-16 h-16 text-muted mx-auto mb-3" />
                    <h5>财务图表功能开发中...</h5>
                    <p className="text-muted">将展示收入支出趋势、项目盈利分析等</p>
                  </div>
                </Card.Body>
              </Card>
            </Tab>

            <Tab
              eventKey="invoices"
              title={
                <>
                  <FaFileLines className="w-4 h-4 me-2" />
                  发票管理
                </>
              }
            >
              <Card>
                <Card.Header>
                  <div className="d-flex justify-content-between align-items-center">
                    <div className="d-flex align-items-center">
                      <InputGroup className="me-3" style={{ width: '250px' }}>
                        <InputGroup.Text>
                          <FaMagnifyingGlass className="w-4 h-4" />
                        </InputGroup.Text>
                        <Form.Control
                          type="text"
                          placeholder="搜索客户或项目"
                          value={searchTerm}
                          onChange={(e) => setSearchTerm(e.target.value)}
                        />
                      </InputGroup>
                      <Form.Select
                        value={filterStatus}
                        onChange={(e) => setFilterStatus(e.target.value)}
                        style={{ width: '120px' }}
                      >
                        <option value="all">所有状态</option>
                        <option value="pending">待付款</option>
                        <option value="paid">已付款</option>
                        <option value="overdue">逾期</option>
                      </Form.Select>
                    </div>
                    <Button variant="primary">
                      <FaPlus className="w-4 h-4 me-2" />
                      新建发票
                    </Button>
                  </div>
                </Card.Header>
                <Card.Body>
                  <div className="table-responsive">
                    <Table striped hover>
                      <thead>
                        <tr>
                          <th>发票号</th>
                          <th>客户名称</th>
                          <th>项目名称</th>
                          <th>金额</th>
                          <th>状态</th>
                          <th>开票日期</th>
                          <th>到期日期</th>
                          <th>操作</th>
                        </tr>
                      </thead>
                      <tbody>
                        {filteredInvoices.map((invoice) => (
                          <tr key={invoice.id}>
                            <td>{invoice.invoiceNumber}</td>
                            <td>{invoice.clientName}</td>
                            <td>{invoice.projectName}</td>
                            <td>{formatCurrency(invoice.amount)}</td>
                            <td>{getInvoiceStatusBadge(invoice.status)}</td>
                            <td>{invoice.issueDate}</td>
                            <td>{invoice.dueDate}</td>
                            <td>
                              <div className="btn-group" role="group">
                                <Button variant="outline-primary" size="sm">
                                  <FaPenToSquare className="w-4 h-4" />
                                </Button>
                                <Button variant="outline-danger" size="sm">
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

            <Tab
              eventKey="expenses"
              title={
                <>
                  <FaCreditCard className="w-4 h-4 me-2" />
                  费用管理
                </>
              }
            >
              <Card>
                <Card.Header>
                  <div className="d-flex justify-content-between align-items-center">
                    <div className="d-flex align-items-center">
                      <InputGroup className="me-3" style={{ width: '250px' }}>
                        <InputGroup.Text>
                          <FaMagnifyingGlass className="w-4 h-4" />
                        </InputGroup.Text>
                        <Form.Control
                          type="text"
                          placeholder="搜索费用描述"
                          value={searchTerm}
                          onChange={(e) => setSearchTerm(e.target.value)}
                        />
                      </InputGroup>
                      <Form.Select
                        value={filterStatus}
                        onChange={(e) => setFilterStatus(e.target.value)}
                        style={{ width: '120px' }}
                      >
                        <option value="all">所有状态</option>
                        <option value="pending">待审批</option>
                        <option value="approved">已批准</option>
                        <option value="rejected">已拒绝</option>
                      </Form.Select>
                    </div>
                    <Button variant="primary">
                      <FaPlus className="w-4 h-4 me-2" />
                      申请费用
                    </Button>
                  </div>
                </Card.Header>
                <Card.Body>
                  <div className="table-responsive">
                    <Table striped hover>
                      <thead>
                        <tr>
                          <th>费用编号</th>
                          <th>类别</th>
                          <th>描述</th>
                          <th>金额</th>
                          <th>申请人</th>
                          <th>申请日期</th>
                          <th>状态</th>
                          <th>操作</th>
                        </tr>
                      </thead>
                      <tbody>
                        {filteredExpenses.map((expense) => (
                          <tr key={expense.id}>
                            <td>{expense.expenseNumber}</td>
                            <td>{expense.category}</td>
                            <td>{expense.description}</td>
                            <td>{formatCurrency(expense.amount)}</td>
                            <td>
                              <div className="d-flex align-items-center">
                                <FaUser className="w-4 h-4 me-1" />
                                {expense.applicant}
                              </div>
                            </td>
                            <td>{expense.date}</td>
                            <td>{getExpenseStatusBadge(expense.status)}</td>
                            <td>
                              <div className="btn-group" role="group">
                                <Button variant="outline-primary" size="sm">
                                  <FaPenToSquare className="w-4 h-4" />
                                </Button>
                                <Button variant="outline-danger" size="sm">
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

            <Tab
              eventKey="reports"
              title={
                <>
                  <FaFileLines className="w-4 h-4 me-2" />
                  财务报表
                </>
              }
            >
              <Card>
                <Card.Header>
                  <h5 className="mb-0">财务报表</h5>
                </Card.Header>
                <Card.Body>
                  <div className="text-center py-5">
                    <DocumentTextIcon className="w-16 h-16 text-muted mx-auto mb-3" />
                    <h5>财务报表功能开发中...</h5>
                    <p className="text-muted">将包含：月度报表、年度报表、项目盈利分析等</p>
                  </div>
                </Card.Body>
              </Card>
            </Tab>
          </Tabs>
        </Card.Body>
      </Card>

      {/* 通知区域 */}
      {stats.overdueRevenue > 0 && (
        <Alert variant="danger" className="mb-4">
          <FaTriangleExclamation className="w-5 h-5 me-2" />
          <strong>提醒：</strong>
          您有 {formatCurrency(stats.overdueRevenue)} 逾期款项未收回，请及时跟进。
        </Alert>
      )}

      {stats.pendingRevenue > 100000 && (
        <Alert variant="warning" className="mb-4">
          <FaClock className="w-5 h-5 me-2" />
          <strong>提醒：</strong>
          您有 {formatCurrency(stats.pendingRevenue)} 待收款项，建议跟进回款进度。
        </Alert>
      )}
    </div>
  );
};

export default FinanceManagement;