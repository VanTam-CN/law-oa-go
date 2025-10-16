import React, { useState, useEffect } from 'react';
import { Container, Row, Col, Badge, Spinner, Button, Card, Table, Alert, Tabs } from 'react-bootstrap';
import { authService } from '../services/authService';
import { clientService } from '../services/clientService';
import { caseService } from '../services/caseService';
import { apiTestRunner } from '../services/apiTest';

interface TestResult {
  name: string;
  status: 'pending' | 'running' | 'success' | 'error';
  message?: string;
  data?: any;
}

const APITestPage: React.FC = () => {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [user, setUser] = useState<any>(null);
  const [testResults, setTestResults] = useState<TestResult[]>([
    { name: '认证状态检查', status: 'pending' },
    { name: '客户管理测试', status: 'pending' },
    { name: '案件管理测试', status: 'pending' },
    { name: '错误处理测试', status: 'pending' }
  ]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [clients, setClients] = useState<any[]>([]);
  const [cases, setCases] = useState<any[]>([]);

  useEffect(() => {
    checkAuthStatus();
  }, []);

  const checkAuthStatus = async () => {
    try {
      const isAuthenticated = authService.isAuthenticated();
      setIsLoggedIn(isAuthenticated);

      if (isAuthenticated) {
        const currentUser = await authService.getCurrentUser();
        setUser(currentUser);
      }
    } catch (err) {
      console.error('检查认证状态失败:', err);
    }
  };

  const handleLogin = async () => {
    setLoading(true);
    setError(null);

    try {
      const result = await authService.login({
        email: 'admin@lawfirm.com',
        password: 'password123'
      });

      setIsLoggedIn(true);
      setUser(result.user);

      // 更新测试结果
      updateTestResult('认证状态检查', 'success', '登录成功');
    } catch (err: any) {
      setError(`登录失败: ${err.message}`);
      updateTestResult('认证状态检查', 'error', err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = async () => {
    try {
      await authService.logout();
      setIsLoggedIn(false);
      setUser(null);
      setClients([]);
      setCases([]);
      setError(null);

      // 重置测试结果
      setTestResults(testResults.map(result => ({ ...result, status: 'pending' as const, message: undefined })));
    } catch (err: any) {
      setError(`登出失败: ${err.message}`);
    }
  };

  const updateTestResult = (name: string, status: TestResult['status'], message?: string, data?: any) => {
    setTestResults(prev =>
      prev.map(result =>
        result.name === name
          ? { ...result, status, message, data }
          : result
      )
    );
  };

  const runClientTest = async () => {
    updateTestResult('客户管理测试', 'running');

    try {
      // 测试获取客户列表
      const clientsData = await clientService.getClients({ page: 1, page_size: 5 });
      setClients(clientsData.data);

      // 测试获取客户统计
      const stats = await clientService.getClientStats();

      updateTestResult('客户管理测试', 'success',
        `成功获取 ${clientsData.data.length} 个客户，总计 ${stats.total} 个客户`,
        { clients: clientsData.data, stats }
      );
    } catch (err: any) {
      updateTestResult('客户管理测试', 'error', err.message);
    }
  };

  const runCaseTest = async () => {
    updateTestResult('案件管理测试', 'running');

    try {
      // 测试获取案件列表
      const casesData = await caseService.getCases({ page: 1, page_size: 5 });
      setCases(casesData.data);

      // 测试获取案件统计
      const stats = await caseService.getCaseStats();

      updateTestResult('案件管理测试', 'success',
        `成功获取 ${casesData.data.length} 个案件，总计 ${stats.total} 个案件`,
        { cases: casesData.data, stats }
      );
    } catch (err: any) {
      updateTestResult('案件管理测试', 'error', err.message);
    }
  };

  const runAllTests = async () => {
    setLoading(true);
    setError(null);

    try {
      await runClientTest();
      await runCaseTest();
    } catch (err: any) {
      setError(`运行测试失败: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  const getStatusBadge = (status: TestResult['status']) => {
    switch (status) {
      case 'pending':
        return <Badge bg="secondary">待测试</Badge>;
      case 'running':
        return <Badge bg="warning">
          <Spinner as="span" size="sm" animation="border" /> 运行中
        </Badge>;
      case 'success':
        return <Badge bg="success">成功</Badge>;
      case 'error':
        return <Badge bg="danger">失败</Badge>;
      default:
        return <Badge bg="secondary">未知</Badge>;
    }
  };

  return (
    <Container className="py-4">
      <Row className="mb-4">
        <Col>
          <h1 className="mb-3">🚀 前后端API集成测试</h1>
          <p className="text-muted">
            此页面用于测试前端服务层与后端API的集成情况
          </p>
        </Col>
      </Row>

      {/* 认证状态卡片 */}
      <Row className="mb-4">
        <Col md={6}>
          <Card>
            <Card.Header as="h5">🔐 认证状态</Card.Header>
            <Card.Body>
              <div className="d-flex justify-content-between align-items-center">
                <div>
                  <p className="mb-1">
                    <strong>状态:</strong> {isLoggedIn ? '已登录' : '未登录'}
                  </p>
                  {user && (
                    <p className="mb-0">
                      <strong>用户:</strong> {user.name} ({user.role})
                    </p>
                  )}
                </div>
                <div>
                  {!isLoggedIn ? (
                    <Button onClick={handleLogin} disabled={loading}>
                      {loading ? <Spinner as="span" size="sm" animation="border" /> : '登录'}
                    </Button>
                  ) : (
                    <Button variant="outline-danger" onClick={handleLogout}>
                      登出
                    </Button>
                  )}
                </div>
              </div>
            </Card.Body>
          </Card>
        </Col>

        <Col md={6}>
          <Card>
            <Card.Header as="h5">🧪 测试控制</Card.Header>
            <Card.Body>
              <div className="d-grid gap-2">
                <Button
                  onClick={runAllTests}
                  disabled={!isLoggedIn || loading}
                  variant="primary"
                >
                  {loading ? <Spinner as="span" size="sm" animation="border" /> : '运行所有测试'}
                </Button>
                <Button
                  onClick={runClientTest}
                  disabled={!isLoggedIn || loading}
                  variant="outline-primary"
                >
                  测试客户管理
                </Button>
                <Button
                  onClick={runCaseTest}
                  disabled={!isLoggedIn || loading}
                  variant="outline-primary"
                >
                  测试案件管理
                </Button>
              </div>
            </Card.Body>
          </Card>
        </Col>
      </Row>

      {/* 错误提示 */}
      {error && (
        <Row className="mb-4">
          <Col>
            <Alert variant="danger">
              <strong>错误:</strong> {error}
            </Alert>
          </Col>
        </Row>
      )}

      {/* 测试结果 */}
      <Row className="mb-4">
        <Col>
          <Card>
            <Card.Header as="h5">📊 测试结果</Card.Header>
            <Card.Body>
              <Table striped bordered hover>
                <thead>
                  <tr>
                    <th>测试项目</th>
                    <th>状态</th>
                    <th>消息</th>
                  </tr>
                </thead>
                <tbody>
                  {testResults.map((result, index) => (
                    <tr key={index}>
                      <td>{result.name}</td>
                      <td>{getStatusBadge(result.status)}</td>
                      <td>{result.message || '-'}</td>
                    </tr>
                  ))}
                </tbody>
              </Table>
            </Card.Body>
          </Card>
        </Col>
      </Row>

      {/* 客户数据 */}
      {clients.length > 0 && (
        <Row className="mb-4">
          <Col>
            <Card>
              <Card.Header as="h5">👥 客户数据 (前5条)</Card.Header>
              <Card.Body>
                <Table striped bordered hover responsive>
                  <thead>
                    <tr>
                      <th>ID</th>
                      <th>姓名</th>
                      <th>邮箱</th>
                      <th>电话</th>
                      <th>公司</th>
                      <th>状态</th>
                    </tr>
                  </thead>
                  <tbody>
                    {clients.map((client) => (
                      <tr key={client.id}>
                        <td>{client.id}</td>
                        <td>{client.name}</td>
                        <td>{client.email}</td>
                        <td>{client.phone}</td>
                        <td>{client.company}</td>
                        <td>
                          <Badge bg={client.status === 'active' ? 'success' : 'secondary'}>
                            {client.status}
                          </Badge>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </Table>
              </Card.Body>
            </Card>
          </Col>
        </Row>
      )}

      {/* 案件数据 */}
      {cases.length > 0 && (
        <Row className="mb-4">
          <Col>
            <Card>
              <Card.Header as="h5">⚖️ 案件数据 (前5条)</Card.Header>
              <Card.Body>
                <Table striped bordered hover responsive>
                  <thead>
                    <tr>
                      <th>ID</th>
                      <th>标题</th>
                      <th>类型</th>
                      <th>优先级</th>
                      <th>状态</th>
                      <th>客户</th>
                    </tr>
                  </thead>
                  <tbody>
                    {cases.map((case_item) => (
                      <tr key={case_item.id}>
                        <td>{case_item.id}</td>
                        <td>{case_item.title}</td>
                        <td>{case_item.case_type}</td>
                        <td>
                          <Badge bg={
                            case_item.priority === 'urgent' ? 'danger' :
                            case_item.priority === 'high' ? 'warning' :
                            case_item.priority === 'medium' ? 'info' : 'secondary'
                          }>
                            {case_item.priority}
                          </Badge>
                        </td>
                        <td>
                          <Badge bg={
                            case_item.status === 'active' ? 'success' :
                            case_item.status === 'pending' ? 'warning' :
                            case_item.status === 'closed' ? 'secondary' : 'info'
                          }>
                            {case_item.status}
                          </Badge>
                        </td>
                        <td>{case_item.client?.name || '-'}</td>
                      </tr>
                    ))}
                  </tbody>
                </Table>
              </Card.Body>
            </Card>
          </Col>
        </Row>
      )}

      {/* 使用说明 */}
      <Row>
        <Col>
          <Card>
            <Card.Header as="h5">📝 使用说明</Card.Header>
            <Card.Body>
              <ol>
                <li>确保后端服务正在运行在 <code>http://localhost:8080</code></li>
                <li>点击"登录"按钮使用测试账号登录</li>
                <li>点击"运行所有测试"按钮进行完整的API集成测试</li>
                <li>查看测试结果和数据展示区域</li>
                <li>测试完成后可以点击"登出"按钮</li>
              </ol>
              <p className="mb-0">
                <strong>测试账号:</strong> admin@lawfirm.com / password123
              </p>
            </Card.Body>
          </Card>
        </Col>
      </Row>
    </Container>
  );
};

export default APITestPage;