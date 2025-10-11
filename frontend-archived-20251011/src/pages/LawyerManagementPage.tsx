import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import {
  Card,
  Button,
  Spinner,
  Form,
  Table,
  Badge,
  Modal,
  InputGroup,
  Row,
  Col,
  Alert
} from "react-bootstrap";
import {
  FaPlus,
  FaEdit,
  FaTrash,
  FaEye,
  FaSync,
  FaUser,
  FaPhone,
  FaEnvelope,
  FaBriefcase,
  FaCheckCircle,
  FaUserTimes,
  FaPauseCircle,
  FaUsers,
  FaUserTie,
  FaCalendar
} from "react-icons/fa";
import {
  MagnifyingGlassIcon
} from "@heroicons/react/24/outline";
import lawyerService, { Lawyer, LawyerStats } from "../services/lawyerService";


const LawyerManagementPage: React.FC = () => {
  const navigate = useNavigate();
  const [lawyers, setLawyers] = useState<Lawyer[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [modalVisible, setModalVisible] = useState<boolean>(false);
  const [modalTitle, setModalTitle] = useState<string>('');
  const [editingLawyer, setEditingLawyer] = useState<Lawyer | null>(null);
  const [stats, setStats] = useState<LawyerStats | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [retryCount, setRetryCount] = useState<number>(0);

  // 查询参数
  const [queryParams, setQueryParams] = useState({
    name: '',
    department: '',
    status: '',
    specialty: '',
    pageNum: 1,
    pageSize: 10
  });

  const [total, setTotal] = useState<number>(0);

  // 获取律师列表
  const fetchLawyers = async (isRetry: boolean = false) => {
    setLoading(true);
    setError(null);

    try {
      console.log('📡 调用律师列表API');
      const response = await lawyerService.getLawyers({
        page: queryParams.pageNum,
        pageSize: queryParams.pageSize,
        name: queryParams.name || undefined,
        department: queryParams.department || undefined,
        status: queryParams.status || undefined
      });

      setLawyers(response.data);
      setTotal(response.total);
      setRetryCount(0); // 重置重试计数
      console.log(`✅ 律师列表加载完成：${response.data.length} 条记录`);
    } catch (error) {
      console.error('获取律师列表失败:', error);

      // 重试机制
      if (retryCount < 2 && !isRetry) {
        console.log(`🔄 第 ${retryCount + 1} 次重试...`);
        setRetryCount(retryCount + 1);
        setTimeout(() => {
          fetchLawyers(true);
        }, 1000 * (retryCount + 1)); // 递增延迟
        return;
      }

      setError(`加载律师列表失败: ${error instanceof Error ? error.message : '未知错误'}`);
    } finally {
      setLoading(false);
    }
  };

  // 获取律师统计
  const fetchStats = async () => {
    try {
      console.log('📡 调用律师统计API');
      const statsData = await lawyerService.getLawyerStats();
      setStats(statsData);
      console.log('✅ 律师统计数据加载完成');
    } catch (error) {
      console.error('获取统计数据失败:', error);
      setError(`获取统计数据失败: ${error instanceof Error ? error.message : '未知错误'}`);
    }
  };

  useEffect(() => {
    fetchLawyers();
    fetchStats();
  }, [queryParams]);

  // 搜索
  const handleSearch = () => {
    setQueryParams({ ...queryParams, pageNum: 1 });
  };

  // 重置搜索
  const handleReset = () => {
    setQueryParams({
      name: '',
      department: '',
      status: '',
      specialty: '',
      pageNum: 1,
      pageSize: 10
    });
    setError(null);
  };

  // 手动重试
  const handleRetry = () => {
    console.log('🔄 用户手动重试...');
    setRetryCount(0);
    setError(null);
    fetchLawyers();
  };

  // 获取状态标签
  const getStatusBadge = (status: string) => {
    const statusMap = {
      'active': { text: '在职', variant: 'success', icon: <FaCheckCircle /> },
      'inactive': { text: '离职', variant: 'danger', icon: <FaUserTimes /> },
      'on_leave': { text: '休假', variant: 'warning', icon: <FaPauseCircle /> }
    };
    const config = statusMap[status as keyof typeof statusMap] || { text: '未知', variant: 'secondary' };
    return (
      <Badge bg={config.variant} className="d-flex align-items-center">
        {config.icon}
        <span className="ms-1">{config.text}</span>
      </Badge>
    );
  };

  // 获取职位标签
  const getPositionBadge = (position: string) => {
    const positionMap = {
      '合伙人': { variant: 'danger' },
      '资深律师': { variant: 'primary' },
      '律师': { variant: 'info' }
    };
    const variant = positionMap[position as keyof typeof positionMap] || 'secondary';
    return <Badge bg={variant}>{position}</Badge>;
  };

  // 打开新增律师弹窗
  const handleAdd = () => {
    setModalTitle('新增律师');
    setEditingLawyer(null);
    setModalVisible(true);
  };

  // 打开编辑律师弹窗
  const handleEdit = (lawyer: Lawyer) => {
    setModalTitle('编辑律师');
    setEditingLawyer(lawyer);
    setModalVisible(true);
  };

  // 删除律师
  const handleDelete = async (lawyer: Lawyer) => {
    if (window.confirm(`确定要删除${lawyer.name}吗？`)) {
      try {
        if (lawyer.id) {
          await lawyerService.deleteLawyer(lawyer.id);
          console.log(`✅ 律师 ${lawyer.name} 删除成功`);
          fetchLawyers();
          fetchStats();
        }
      } catch (error) {
        console.error('删除失败:', error);
        alert(`删除失败: ${error instanceof Error ? error.message : '未知错误'}`);
      }
    }
  };

  // 查看详情
  const handleView = (lawyer: Lawyer) => {
    navigate(`/lawyer/${lawyer.id}`);
  };

  // 提交表单
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!editingLawyer) return;

    try {
      if (editingLawyer.id) {
        // 更新律师
        await lawyerService.updateLawyer(editingLawyer.id, {
          name: editingLawyer.name,
          phone: editingLawyer.phone,
          email: editingLawyer.email,
          licenseNumber: editingLawyer.licenseNumber,
          department: editingLawyer.department,
          position: editingLawyer.position,
          status: editingLawyer.status,
          experience: editingLawyer.experience,
          profile: editingLawyer.profile,
          address: editingLawyer.address,
          education: editingLawyer.education,
          achievements: editingLawyer.achievements,
          hourlyRate: editingLawyer.hourlyRate,
          consultationHours: editingLawyer.consultationHours
        });
        console.log('✅ 律师更新成功');
      } else {
        // 创建律师
        await lawyerService.createLawyer({
          name: editingLawyer.name || '',
          username: editingLawyer.name?.toLowerCase().replace(/\s+/g, '_') || '',
          phone: editingLawyer.phone || '',
          email: editingLawyer.email || '',
          licenseNumber: editingLawyer.licenseNumber || '',
          department: editingLawyer.department,
          position: editingLawyer.position,
          status: editingLawyer.status || 'active',
          experience: editingLawyer.experience,
          profile: editingLawyer.profile,
          address: editingLawyer.address,
          education: editingLawyer.education,
          achievements: editingLawyer.achievements,
          hourlyRate: editingLawyer.hourlyRate,
          consultationHours: editingLawyer.consultationHours
        });
        console.log('✅ 律师创建成功');
      }

      alert(editingLawyer.id ? '更新成功' : '新增成功');
      setModalVisible(false);
      fetchLawyers();
      fetchStats();
    } catch (error) {
      console.error('保存失败:', error);
      alert(`保存失败: ${error instanceof Error ? error.message : '未知错误'}`);
    }
  };

  // 处理输入变化
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    if (editingLawyer) {
      setEditingLawyer({
        ...editingLawyer,
        [name]: value
      });
    }
  };

  return (
    <div>

      {/* 头部 */}
      <div className="d-flex justify-content-between align-items-center mb-4">
        <div>
          <h1>律师管理</h1>
          <p className="text-muted">
            管理律师事务所的律师信息
          </p>
        </div>
        <Button variant="primary" onClick={handleAdd}>
          <FaPlus className="me-2" />
          新增律师
        </Button>
      </div>

      {/* 统计卡片 */}
      {stats && (
        <Row className="mb-4">
          <Col md={3}>
            <Card className="text-center">
              <Card.Body>
                <FaUsers className="text-primary mb-2" style={{ fontSize: '2rem' }} />
                <h3>{stats.total}</h3>
                <p className="text-muted mb-0">律师总数</p>
              </Card.Body>
            </Card>
          </Col>
          <Col md={3}>
            <Card className="text-center">
              <Card.Body>
                <FaCheckCircle className="text-success mb-2" style={{ fontSize: '2rem' }} />
                <h3>{stats.active}</h3>
                <p className="text-muted mb-0">在职律师</p>
              </Card.Body>
            </Card>
          </Col>
          <Col md={3}>
            <Card className="text-center">
              <Card.Body>
                <FaPauseCircle className="text-warning mb-2" style={{ fontSize: '2rem' }} />
                <h3>{stats.onLeave}</h3>
                <p className="text-muted mb-0">休假律师</p>
              </Card.Body>
            </Card>
          </Col>
          <Col md={3}>
            <Card className="text-center">
              <Card.Body>
                <FaUserTimes className="text-danger mb-2" style={{ fontSize: '2rem' }} />
                <h3>{stats.inactive}</h3>
                <p className="text-muted mb-0">离职律师</p>
              </Card.Body>
            </Card>
          </Col>
        </Row>
      )}

      {/* 搜索表单 */}
      <Card className="mb-4">
        <Card.Body>
          <Form>
            <Row>
              <Col md={4}>
                <Form.Group className="mb-3">
                  <Form.Label>姓名</Form.Label>
                  <InputGroup>
                    <InputGroup.Text>
                      <FaUser />
                    </InputGroup.Text>
                    <Form.Control
                      type="text"
                      placeholder="请输入律师姓名"
                      value={queryParams.name}
                      onChange={(e) => setQueryParams({ ...queryParams, name: e.target.value })}
                    />
                  </InputGroup>
                </Form.Group>
              </Col>
              <Col md={4}>
                <Form.Group className="mb-3">
                  <Form.Label>部门</Form.Label>
                  <Form.Select
                    value={queryParams.department}
                    onChange={(e) => setQueryParams({ ...queryParams, department: e.target.value })}
                  >
                    <option value="">全部部门</option>
                    <option value="民事诉讼部">民事诉讼部</option>
                    <option value="刑事辩护部">刑事辩护部</option>
                    <option value="公司法务部">公司法务部</option>
                    <option value="行政诉讼部">行政诉讼部</option>
                  </Form.Select>
                </Form.Group>
              </Col>
              <Col md={4}>
                <Form.Group className="mb-3">
                  <Form.Label>状态</Form.Label>
                  <Form.Select
                    value={queryParams.status}
                    onChange={(e) => setQueryParams({ ...queryParams, status: e.target.value })}
                  >
                    <option value="">全部状态</option>
                    <option value="active">在职</option>
                    <option value="on_leave">休假</option>
                    <option value="inactive">离职</option>
                  </Form.Select>
                </Form.Group>
              </Col>
            </Row>
            <Row>
              <Col md={12}>
                <div className="d-flex gap-2">
                  <Button variant="primary" onClick={handleSearch}>
                    <MagnifyingGlassIcon className="w-4 h-4 me-2" />
                    搜索
                  </Button>
                  <Button variant="outline-secondary" onClick={handleReset}>
                    <FaSync className="me-2" />
                    重置
                  </Button>
                </div>
              </Col>
            </Row>
          </Form>
        </Card.Body>
      </Card>

      {/* 错误提示和重试 */}
      {error && (
        <Alert variant="warning" className="mb-4">
          <div className="d-flex justify-content-between align-items-center">
            <div>
              <strong>⚠️ 加载错误:</strong> {error}
              <div className="small text-muted mt-1">
                {retryCount > 0 && `已自动重试 ${retryCount} 次`}
              </div>
            </div>
            <Button variant="outline-warning" size="sm" onClick={handleRetry}>
              <FaSync className="me-1" />
              手动重试
            </Button>
          </div>
        </Alert>
      )}

      {/* 律师列表 */}
      <Card>
        <Card.Body>
          {loading ? (
            <div className="text-center py-5">
              <div className="mb-3">
                <Spinner animation="border" role="status" style={{ width: '3rem', height: '3rem' }}>
                  <span className="visually-hidden">加载中...</span>
                </Spinner>
              </div>
              <h5 className="text-muted mb-2">
                {retryCount > 0 ? `正在重试... (${retryCount}/2)` : '正在加载律师列表...'}
              </h5>
              <p className="text-muted small mb-0">
                请稍候，正在获取最新数据
              </p>
              {retryCount > 0 && (
                <div className="mt-2">
                  <small className="text-info">
                    <FaSync className="me-1" />
                    自动重试第 {retryCount} 次
                  </small>
                </div>
              )}
            </div>
          ) : (
            <>
              <Table striped bordered hover responsive>
                <thead>
                  <tr>
                    <th>姓名</th>
                    <th>联系方式</th>
                    <th>执业证号</th>
                    <th>部门</th>
                    <th>职位</th>
                    <th>状态</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  {lawyers.map((lawyer) => (
                    <tr key={lawyer.id}>
                      <td>
                        <div className="d-flex align-items-center">
                          <div className="bg-light rounded-circle d-flex align-items-center justify-content-center me-3" style={{ width: '40px', height: '40px' }}>
                            <FaUser className="text-muted" />
                          </div>
                          <div>
                            <div className="fw-bold">{lawyer.name}</div>
                            <small className="text-muted">{lawyer.specialty?.join(', ')}</small>
                          </div>
                        </div>
                      </td>
                      <td>
                        <div className="mb-1">
                          <FaPhone className="me-1 text-muted" />
                          {lawyer.phone}
                        </div>
                        <div>
                          <FaEnvelope className="me-1 text-muted" />
                          {lawyer.email}
                        </div>
                      </td>
                      <td>
                        <small className="text-muted">{lawyer.licenseNumber}</small>
                      </td>
                      <td>{lawyer.department}</td>
                      <td>{getPositionBadge(lawyer.position || '')}</td>
                      <td>{getStatusBadge(lawyer.status || '')}</td>
                      <td>
                        <div className="d-flex gap-1">
                          <Button
                            variant="outline-primary"
                            size="sm"
                            onClick={() => handleView(lawyer)}
                            title="查看详情"
                          >
                            <FaEye />
                          </Button>
                          <Button
                            variant="outline-warning"
                            size="sm"
                            onClick={() => handleEdit(lawyer)}
                            title="编辑"
                          >
                            <FaEdit />
                          </Button>
                          <Button
                            variant="outline-danger"
                            size="sm"
                            onClick={() => handleDelete(lawyer)}
                            title="删除"
                          >
                            <FaTrash />
                          </Button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </Table>

              {/* 空数据状态 */}
              {!loading && lawyers.length === 0 && (
                <div className="text-center py-5">
                  <FaUserTie className="text-muted mb-3" style={{ fontSize: '3rem' }} />
                  <h5 className="text-muted mb-2">暂无律师数据</h5>
                  <p className="text-muted small mb-3">
                    {queryParams.name || queryParams.department || queryParams.status
                      ? '没有找到符合筛选条件的律师，请尝试调整搜索条件'
                      : '系统还没有律师数据，点击上方"新增律师"按钮添加第一位律师'
                    }
                  </p>
                  {!queryParams.name && !queryParams.department && !queryParams.status && (
                    <Button variant="primary" onClick={handleAdd}>
                      <FaPlus className="me-2" />
                      新增律师
                    </Button>
                  )}
                </div>
              )}

              {/* 分页 */}
              <div className="d-flex justify-content-between align-items-center mt-3">
                <div>
                  显示 {(queryParams.pageNum - 1) * queryParams.pageSize + 1} - {Math.min(queryParams.pageNum * queryParams.pageSize, total)} 条，共 {total} 条
                </div>
                <div className="d-flex gap-2">
                  <Button
                    variant="outline-primary"
                    disabled={queryParams.pageNum === 1}
                    onClick={() => setQueryParams({ ...queryParams, pageNum: queryParams.pageNum - 1 })}
                  >
                    上一页
                  </Button>
                  <Button
                    variant="outline-primary"
                    disabled={queryParams.pageNum * queryParams.pageSize >= total}
                    onClick={() => setQueryParams({ ...queryParams, pageNum: queryParams.pageNum + 1 })}
                  >
                    下一页
                  </Button>
                </div>
              </div>
            </>
          )}
        </Card.Body>
      </Card>

      {/* 编辑律师模态框 */}
      <Modal show={modalVisible} onHide={() => setModalVisible(false)} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>{modalTitle}</Modal.Title>
        </Modal.Header>
        <Form onSubmit={handleSubmit}>
          <Modal.Body>
            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>姓名 *</Form.Label>
                  <Form.Control
                    type="text"
                    name="name"
                    value={editingLawyer?.name || ''}
                    onChange={handleInputChange}
                    required
                  />
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>执业证号 *</Form.Label>
                  <Form.Control
                    type="text"
                    name="licenseNumber"
                    value={editingLawyer?.licenseNumber || ''}
                    onChange={handleInputChange}
                    required
                  />
                </Form.Group>
              </Col>
            </Row>

            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>联系电话 *</Form.Label>
                  <InputGroup>
                    <InputGroup.Text>
                      <FaPhone />
                    </InputGroup.Text>
                    <Form.Control
                      type="tel"
                      name="phone"
                      value={editingLawyer?.phone || ''}
                      onChange={handleInputChange}
                      required
                    />
                  </InputGroup>
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>邮箱 *</Form.Label>
                  <InputGroup>
                    <InputGroup.Text>
                      <FaEnvelope />
                    </InputGroup.Text>
                    <Form.Control
                      type="email"
                      name="email"
                      value={editingLawyer?.email || ''}
                      onChange={handleInputChange}
                      required
                    />
                  </InputGroup>
                </Form.Group>
              </Col>
            </Row>

            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>部门</Form.Label>
                  <Form.Select
                    name="department"
                    value={editingLawyer?.department || ''}
                    onChange={handleInputChange}
                  >
                    <option value="">选择部门</option>
                    <option value="民事诉讼部">民事诉讼部</option>
                    <option value="刑事辩护部">刑事辩护部</option>
                    <option value="公司法务部">公司法务部</option>
                    <option value="行政诉讼部">行政诉讼部</option>
                  </Form.Select>
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>职位</Form.Label>
                  <Form.Select
                    name="position"
                    value={editingLawyer?.position || ''}
                    onChange={handleInputChange}
                  >
                    <option value="">选择职位</option>
                    <option value="合伙人">合伙人</option>
                    <option value="资深律师">资深律师</option>
                    <option value="律师">律师</option>
                  </Form.Select>
                </Form.Group>
              </Col>
            </Row>

            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>状态</Form.Label>
                  <Form.Select
                    name="status"
                    value={editingLawyer?.status || ''}
                    onChange={handleInputChange}
                  >
                    <option value="active">在职</option>
                    <option value="on_leave">休假</option>
                    <option value="inactive">离职</option>
                  </Form.Select>
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>工作经验（年）</Form.Label>
                  <Form.Control
                    type="number"
                    name="experience"
                    value={editingLawyer?.experience || ''}
                    onChange={handleInputChange}
                    min="0"
                  />
                </Form.Group>
              </Col>
            </Row>

            <Row>
              <Col md={12}>
                <Form.Group className="mb-3">
                  <Form.Label>个人简介</Form.Label>
                  <Form.Control
                    as="textarea"
                    rows={3}
                    name="profile"
                    value={editingLawyer?.profile || ''}
                    onChange={handleInputChange}
                    placeholder="请输入个人简介"
                  />
                </Form.Group>
              </Col>
            </Row>
          </Modal.Body>
          <Modal.Footer>
            <Button variant="secondary" onClick={() => setModalVisible(false)}>
              取消
            </Button>
            <Button variant="primary" type="submit">
              {editingLawyer?.id ? '更新' : '新增'}
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>
    </div>
  );
};

export default LawyerManagementPage;