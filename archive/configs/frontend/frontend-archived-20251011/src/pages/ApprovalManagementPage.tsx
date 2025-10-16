import React, { useState, useEffect } from 'react';
import {
  Card,
  Button,
  Table,
  Badge,
  Form,
  InputGroup,
  Tabs,
  Tab,
  Modal,
  Row,
  Col,
  Alert,
  FormControl,
  Spinner
} from 'react-bootstrap';
import {
  FaEye,
  FaXmark,
  FaCheck,
  FaRotate,
  FaPlus,
  FaMagnifyingGlass,
  FaFilter,
  FaFile
} from 'react-icons/fa6';
import { useNavigate } from 'react-router-dom';

interface ApprovalItem {
  id: number;
  type: string;
  title: string;
  content: string;
  applicant: string;
  applicantId: number;
  department: string;
  createTime: string;
  status: 'pending' | 'approved' | 'rejected' | 'cancelled';
  urgency: 'normal' | 'urgent' | 'very_urgent';
  currentApprover?: string;
  currentApproverId?: number;
}

interface ApprovalStats {
  pendingCount: number;
  myPendingCount: number;
  myTotalCount: number;
}

const ApprovalManagement: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [myApprovals, setMyApprovals] = useState<ApprovalItem[]>([]);
  const [pendingApprovals, setPendingApprovals] = useState<ApprovalItem[]>([]);
  const [stats, setStats] = useState<ApprovalStats | null>(null);
  const [activeTab, setActiveTab] = useState<string>('pending');
  const [searchTerm, setSearchTerm] = useState<string>('');
  const [filterType, setFilterType] = useState<string>('all');

  useEffect(() => {
    fetchApprovals();
    fetchStats();
  }, [activeTab]);

  const fetchApprovals = async () => {
    try {
      setLoading(true);
      setError(null);

      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 1000));

      const mockData: ApprovalItem[] = [
        {
          id: 1,
          type: '请假申请',
          title: '事假申请 - 家中有事',
          content: '因家中有急事需要请假3天',
          applicant: '张三',
          applicantId: 1,
          department: '技术部',
          createTime: '2024-01-15 09:00:00',
          status: 'pending',
          urgency: 'normal',
          currentApprover: '李四',
          currentApproverId: 2
        },
        {
          id: 2,
          type: '报销申请',
          title: '差旅费用报销 - 北京出差',
          content: '北京出差期间的交通和住宿费用报销',
          applicant: '王五',
          applicantId: 3,
          department: '销售部',
          createTime: '2024-01-14 14:30:00',
          status: 'pending',
          urgency: 'urgent',
          currentApprover: '赵六',
          currentApproverId: 4
        },
        {
          id: 3,
          type: '采购申请',
          title: '办公用品采购 - 打印机',
          content: '申请采购新的激光打印机一台',
          applicant: '钱七',
          applicantId: 5,
          department: '行政部',
          createTime: '2024-01-13 11:15:00',
          status: 'approved',
          urgency: 'normal'
        },
        {
          id: 4,
          type: '项目申请',
          title: '新项目启动 - 客户管理系统',
          content: '申请启动新的客户管理系统开发项目',
          applicant: '孙八',
          applicantId: 6,
          department: '产品部',
          createTime: '2024-01-12 16:45:00',
          status: 'rejected',
          urgency: 'very_urgent'
        }
      ];

      const myApprovalsData = mockData.filter(item => item.applicantId === 1);
      const pendingApprovalsData = mockData.filter(item => item.status === 'pending');

      setMyApprovals(myApprovalsData);
      setPendingApprovals(pendingApprovalsData);
    } catch (error) {
      console.error('Failed to fetch approvals:', error);
      setError('获取审批列表失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchStats = async () => {
    try {
      // 模拟统计数据
      await new Promise(resolve => setTimeout(resolve, 500));

      setStats({
        pendingCount: 2,
        myPendingCount: 1,
        myTotalCount: 3
      });
    } catch (error) {
      console.error('Failed to fetch stats:', error);
    }
  };

  const renderStatusBadge = (status: string) => {
    switch (status) {
      case 'approved':
        return <Badge bg="success">已通过</Badge>;
      case 'rejected':
        return <Badge bg="danger">已拒绝</Badge>;
      case 'pending':
        return <Badge bg="warning">处理中</Badge>;
      case 'cancelled':
        return <Badge bg="secondary">已撤回</Badge>;
      default:
        return <Badge bg="light">未知</Badge>;
    }
  };

  const renderUrgencyBadge = (urgency: string) => {
    switch (urgency) {
      case 'very_urgent':
        return <Badge bg="danger">特急</Badge>;
      case 'urgent':
        return <Badge bg="warning">紧急</Badge>;
      case 'normal':
        return <Badge bg="info">普通</Badge>;
      default:
        return <Badge bg="light">未知</Badge>;
    }
  };

  const handleView = (id: number) => {
    navigate(`/approval/${id}`);
  };

  const handleCancel = async (id: number) => {
    try {
      // 模拟撤回操作
      await new Promise(resolve => setTimeout(resolve, 500));

      // 更新本地状态
      const updatedApprovals = myApprovals.map(item =>
        item.id === id ? { ...item, status: 'cancelled' as const } : item
      );
      setMyApprovals(updatedApprovals);

      alert('审批已撤回');
      fetchApprovals();
      fetchStats();
    } catch (error) {
      console.error('Failed to cancel approval:', error);
      alert('撤回失败');
    }
  };

  const getFilteredData = () => {
    let data = activeTab === 'my' ? myApprovals : pendingApprovals;

    if (searchTerm) {
      data = data.filter(item =>
        item.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
        item.applicant.toLowerCase().includes(searchTerm.toLowerCase()) ||
        item.type.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }

    if (filterType !== 'all') {
      data = data.filter(item => item.type === filterType);
    }

    return data;
  };

  const getTypeOptions = () => {
    const types = [...new Set([...myApprovals, ...pendingApprovals].map(item => item.type))];
    return types.map(type => ({ value: type, label: type }));
  };

  if (loading) {
    return (
      <div className="d-flex min-vh-100 align-items-center justify-content-center">
        <Spinner animation="border" role="status">
          <span className="visually-hidden">加载中...</span>
        </Spinner>
      </div>
    );
  }

  return (
    <div className="approval-management p-4">
      <Card className="mb-4">
        <Card.Header className="d-flex justify-content-between align-items-center">
          <h4 className="mb-0">审批管理</h4>
          <Button
            variant="primary"
            onClick={() => navigate('/approval/create')}
          >
            <FaPlus className="w-4 h-4 me-2" />
            新建审批
          </Button>
        </Card.Header>
        <Card.Body>
          {error && (
            <Alert variant="danger" onClose={() => setError(null)} dismissible>
              {error}
            </Alert>
          )}

          {/* 统计信息 */}
          <div className="row mb-4">
            <div className="col-md-3">
              <Card className="text-center" bg="light">
                <Card.Body>
                  <h3>{stats?.pendingCount || 0}</h3>
                  <p className="text-muted mb-0">待审批总数</p>
                </Card.Body>
              </Card>
            </div>
            <div className="col-md-3">
              <Card className="text-center" bg="warning">
                <Card.Body>
                  <h3>{stats?.myPendingCount || 0}</h3>
                  <p className="mb-0">待我审批</p>
                </Card.Body>
              </Card>
            </div>
            <div className="col-md-3">
              <Card className="text-center" bg="info">
                <Card.Body>
                  <h3>{stats?.myTotalCount || 0}</h3>
                  <p className="mb-0">我的申请</p>
                </Card.Body>
              </Card>
            </div>
            <div className="col-md-3">
              <Card className="text-center" bg="success">
                <Card.Body>
                  <h3>{myApprovals.filter(a => a.status === 'approved').length}</h3>
                  <p className="mb-0">已通过</p>
                </Card.Body>
              </Card>
            </div>
          </div>

          {/* 搜索和筛选 */}
          <div className="row mb-3">
            <div className="col-md-4">
              <InputGroup>
                <InputGroup.Text>
                  <FaMagnifyingGlass className="w-4 h-4" />
                </InputGroup.Text>
                <FormControl
                  placeholder="搜索申请标题、申请人或类型..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                />
              </InputGroup>
            </div>
            <div className="col-md-3">
              <Form.Select
                value={filterType}
                onChange={(e) => setFilterType(e.target.value)}
              >
                <option value="all">所有类型</option>
                {getTypeOptions().map(option => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </Form.Select>
            </div>
            <div className="col-md-2">
              <Button variant="outline-secondary" onClick={() => {
                setSearchTerm('');
                setFilterType('all');
              }}>
                <FaFilter className="w-4 h-4 me-2" />
                重置筛选
              </Button>
            </div>
          </div>

          {/* 标签页 */}
          <Tabs
            activeKey={activeTab}
            onSelect={(key) => setActiveTab(key || 'pending')}
            className="mb-3"
          >
            <Tab
              eventKey="pending"
              title={
                <>
                  待我审批
                  {stats && <Badge bg="danger" className="ms-2">{stats.myPendingCount}</Badge>}
                </>
              }
            >
              <div className="table-responsive">
                <Table striped hover>
                  <thead>
                    <tr>
                      <th>申请类型</th>
                      <th>标题</th>
                      <th>申请人</th>
                      <th>部门</th>
                      <th>申请时间</th>
                      <th>状态</th>
                      <th>紧急程度</th>
                      <th>操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    {getFilteredData().map((item) => (
                      <tr key={item.id}>
                        <td>{item.type}</td>
                        <td>
                          <button
                            className="btn btn-link p-0 text-decoration-none"
                            onClick={() => handleView(item.id)}
                          >
                            {item.title}
                          </button>
                        </td>
                        <td>{item.applicant}</td>
                        <td>{item.department}</td>
                        <td>{item.createTime}</td>
                        <td>{renderStatusBadge(item.status)}</td>
                        <td>{renderUrgencyBadge(item.urgency)}</td>
                        <td>
                          <Button
                            variant="outline-primary"
                            size="sm"
                            onClick={() => handleView(item.id)}
                            className="me-2"
                          >
                            <FaEye className="w-4 h-4" />
                          </Button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </Table>
              </div>
            </Tab>

            <Tab
              eventKey="my"
              title={
                <>
                  我的申请
                  {stats && <Badge bg="info" className="ms-2">{stats.myTotalCount}</Badge>}
                </>
              }
            >
              <div className="table-responsive">
                <Table striped hover>
                  <thead>
                    <tr>
                      <th>申请类型</th>
                      <th>标题</th>
                      <th>申请人</th>
                      <th>部门</th>
                      <th>申请时间</th>
                      <th>状态</th>
                      <th>紧急程度</th>
                      <th>操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    {getFilteredData().map((item) => (
                      <tr key={item.id}>
                        <td>{item.type}</td>
                        <td>
                          <button
                            className="btn btn-link p-0 text-decoration-none"
                            onClick={() => handleView(item.id)}
                          >
                            {item.title}
                          </button>
                        </td>
                        <td>{item.applicant}</td>
                        <td>{item.department}</td>
                        <td>{item.createTime}</td>
                        <td>{renderStatusBadge(item.status)}</td>
                        <td>{renderUrgencyBadge(item.urgency)}</td>
                        <td>
                          <Button
                            variant="outline-primary"
                            size="sm"
                            onClick={() => handleView(item.id)}
                            className="me-2"
                          >
                            <FaEye className="w-4 h-4" />
                          </Button>
                          {item.status === 'pending' && (
                            <Button
                              variant="outline-danger"
                              size="sm"
                              onClick={() => handleCancel(item.id)}
                            >
                              <FaXmark className="w-4 h-4" />
                            </Button>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </Table>
              </div>
            </Tab>
          </Tabs>
        </Card.Body>
      </Card>
    </div>
  );
};

export default ApprovalManagement;