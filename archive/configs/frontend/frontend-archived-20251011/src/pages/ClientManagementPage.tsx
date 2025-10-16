import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { getClients, createClient, updateClient, deleteClient } from "../services/clientService";
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
  FaBuilding,
  FaPhone,
  FaEnvelope,
  FaMapMarkerAlt,
  FaUsers,
  FaUserCheck,
  FaUserTimes,
  FaIdCard,
  FaBriefcase,
  FaStar,
  FaCircleInfo
} from "react-icons/fa";
import {
  MagnifyingGlassIcon
} from "@heroicons/react/24/outline";

// 客户接口
interface Client {
  id?: number;
  name?: string;
  type?: 'PERSON' | 'COMPANY';
  phone?: string;
  email?: string;
  address?: string;
  status?: 'active' | 'inactive';
  source?: string;
  remark?: string;
  // 个人客户特有字段
  idCard?: string;
  // 企业客户特有字段
  company?: string;
  industry?: string;
  contactPerson?: string;
  contactPhone?: string;
}

// 客户统计接口
interface ClientStats {
  total: number;
  active: number;
  inactive: number;
  person: number;
  company: number;
}

const ClientManagementPage: React.FC = () => {
  const navigate = useNavigate();
  const [clients, setClients] = useState<Client[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [modalVisible, setModalVisible] = useState<boolean>(false);
  const [modalTitle, setModalTitle] = useState<string>('');
  const [editingClient, setEditingClient] = useState<Client | null>(null);
  const [stats, setStats] = useState<ClientStats | null>(null);

  // 查询参数
  const [queryParams, setQueryParams] = useState({
    name: '',
    type: '',
    status: '',
    pageNum: 1,
    pageSize: 10
  });

  const [total, setTotal] = useState<number>(0);

  
  // 获取客户列表
  const fetchClients = async () => {
    setLoading(true);
    try {
      // 构建查询参数
      const params = new URLSearchParams({
        page: queryParams.pageNum.toString(),
        page_size: queryParams.pageSize.toString()
      });

      if (queryParams.name) {
        params.append('search', queryParams.name);
      }

      if (queryParams.status) {
        params.append('status', queryParams.status);
      }

      const response = await getClients(params.toString());
      setClients(response.data);
      setTotal(response.pagination?.total || 0);
    } catch (error) {
      console.error('获取客户列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  // 获取客户统计
  const fetchStats = async () => {
    try {
      const statsResponse = await getClientStats();
      setStats({
        total: statsResponse.total_clients,
        active: statsResponse.active_clients,
        inactive: statsResponse.inactive_clients,
        person: Math.floor(statsResponse.total_clients * 0.6), // 估算
        company: Math.floor(statsResponse.total_clients * 0.4)  // 估算
      });
    } catch (error) {
      console.error('获取统计数据失败:', error);
    }
  };

  useEffect(() => {
    fetchClients();
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
      type: '',
      status: '',
      pageNum: 1,
      pageSize: 10
    });
  };

  // 获取状态标签
  const getStatusBadge = (status: string) => {
    const statusMap = {
      'active': { text: '活跃', variant: 'success', icon: <FaUserCheck /> },
      'inactive': { text: '非活跃', variant: 'danger', icon: <FaUserTimes /> }
    };
    const config = statusMap[status as keyof typeof statusMap] || { text: '未知', variant: 'secondary' };
    return (
      <Badge bg={config.variant} className="d-flex align-items-center">
        {config.icon}
        <span className="ms-1">{config.text}</span>
      </Badge>
    );
  };

  // 获取客户类型标签
  const getTypeBadge = (type: string) => {
    const typeMap = {
      'PERSON': { text: '个人', variant: 'primary', icon: <FaUser /> },
      'COMPANY': { text: '企业', variant: 'info', icon: <FaBuilding /> }
    };
    const config = typeMap[type as keyof typeof typeMap] || { text: '未知', variant: 'secondary' };
    return (
      <Badge bg={config.variant} className="d-flex align-items-center">
        {config.icon}
        <span className="ms-1">{config.text}</span>
      </Badge>
    );
  };

  // 打开新增客户弹窗
  const handleAdd = () => {
    setModalTitle('新增客户');
    setEditingClient({
      type: 'PERSON',
      status: 'active'
    });
    setModalVisible(true);
  };

  // 打开编辑客户弹窗
  const handleEdit = (client: Client) => {
    setModalTitle('编辑客户');
    setEditingClient(client);
    setModalVisible(true);
  };

  // 删除客户
  const handleDelete = async (client: Client) => {
    if (window.confirm(`确定要删除${client.name}吗？`)) {
      try {
        // 模拟API调用
        await new Promise(resolve => setTimeout(resolve, 500));

        alert('删除成功');
        fetchClients();
        fetchStats();
      } catch (error) {
        console.error('删除失败:', error);
        alert('删除失败，请重试');
      }
    }
  };

  // 查看详情
  const handleView = (client: Client) => {
    const detailContent = `
      <div class="client-detail">
        <p><strong>客户名称：</strong>${client.name}</p>
        <p><strong>客户类型：</strong>${client.type === 'PERSON' ? '个人' : '企业'}</p>
        <p><strong>联系电话：</strong>${client.phone}</p>
        <p><strong>电子邮箱：</strong>${client.email}</p>
        ${client.type === 'PERSON' && client.idCard ? `<p><strong>身份证号：</strong>${client.idCard}</p>` : ''}
        <p><strong>地址：</strong>${client.address}</p>
        ${client.type === 'COMPANY' ? `
          <p><strong>公司名称：</strong>${client.company}</p>
          <p><strong>所属行业：</strong>${client.industry}</p>
          <p><strong>联系人：</strong>${client.contactPerson}</p>
          <p><strong>联系电话：</strong>${client.contactPhone}</p>
        ` : ''}
        <p><strong>客户来源：</strong>${client.source || '-'}</p>
        <p><strong>客户状态：</strong>${client.status === 'active' ? '活跃' : '非活跃'}</p>
        <p><strong>备注：</strong>${client.remark || '-'}</p>
      </div>
    `;

    // 创建模态框显示详情
    const modal = document.createElement('div');
    modal.innerHTML = `
      <div class="modal fade" id="clientDetailModal" tabindex="-1">
        <div class="modal-dialog modal-lg">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title">客户详情</h5>
              <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
            </div>
            <div class="modal-body">
              ${detailContent}
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">关闭</button>
            </div>
          </div>
        </div>
      </div>
    `;
    document.body.appendChild(modal);

    // 使用Bootstrap的Modal
    const modalElement = new (window as any).bootstrap.Modal(document.getElementById('clientDetailModal'));
    modalElement.show();

    // 清理
    modalElement._element.addEventListener('hidden.bs.modal', () => {
      document.body.removeChild(modal);
    });
  };

  // 提交表单
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!editingClient) return;

    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 1000));

      alert(editingClient.id ? '更新成功' : '新增成功');
      setModalVisible(false);
      fetchClients();
      fetchStats();
    } catch (error) {
      console.error('保存失败:', error);
      alert('保存失败，请重试');
    }
  };

  // 处理输入变化
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    if (editingClient) {
      setEditingClient({
        ...editingClient,
        [name]: value
      });
    }
  };

  return (
    <div>
      {/* 头部 */}
      <div className="d-flex justify-content-between align-items-center mb-4">
        <div>
          <h1>客户管理</h1>
          <p className="text-muted">管理律师事务所的客户信息</p>
        </div>
        <Button variant="primary" onClick={handleAdd}>
          <FaPlus className="me-2" />
          新增客户
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
                <p className="text-muted mb-0">客户总数</p>
              </Card.Body>
            </Card>
          </Col>
          <Col md={3}>
            <Card className="text-center">
              <Card.Body>
                <FaUserCheck className="text-success mb-2" style={{ fontSize: '2rem' }} />
                <h3>{stats.active}</h3>
                <p className="text-muted mb-0">活跃客户</p>
              </Card.Body>
            </Card>
          </Col>
          <Col md={3}>
            <Card className="text-center">
              <Card.Body>
                <FaUser className="text-info mb-2" style={{ fontSize: '2rem' }} />
                <h3>{stats.person}</h3>
                <p className="text-muted mb-0">个人客户</p>
              </Card.Body>
            </Card>
          </Col>
          <Col md={3}>
            <Card className="text-center">
              <Card.Body>
                <FaBuilding className="text-warning mb-2" style={{ fontSize: '2rem' }} />
                <h3>{stats.company}</h3>
                <p className="text-muted mb-0">企业客户</p>
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
                  <Form.Label>客户名称</Form.Label>
                  <InputGroup>
                    <InputGroup.Text>
                      <MagnifyingGlassIcon className="w-4 h-4" />
                    </InputGroup.Text>
                    <Form.Control
                      type="text"
                      placeholder="请输入客户名称"
                      value={queryParams.name}
                      onChange={(e) => setQueryParams({ ...queryParams, name: e.target.value })}
                    />
                  </InputGroup>
                </Form.Group>
              </Col>
              <Col md={4}>
                <Form.Group className="mb-3">
                  <Form.Label>客户类型</Form.Label>
                  <Form.Select
                    value={queryParams.type}
                    onChange={(e) => setQueryParams({ ...queryParams, type: e.target.value })}
                  >
                    <option value="">全部类型</option>
                    <option value="PERSON">个人</option>
                    <option value="COMPANY">企业</option>
                  </Form.Select>
                </Form.Group>
              </Col>
              <Col md={4}>
                <Form.Group className="mb-3">
                  <Form.Label>客户状态</Form.Label>
                  <Form.Select
                    value={queryParams.status}
                    onChange={(e) => setQueryParams({ ...queryParams, status: e.target.value })}
                  >
                    <option value="">全部状态</option>
                    <option value="active">活跃</option>
                    <option value="inactive">非活跃</option>
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

      {/* 客户列表 */}
      <Card>
        <Card.Body>
          {loading ? (
            <div className="text-center py-5">
              <Spinner animation="border" />
              <p className="mt-2">加载中...</p>
            </div>
          ) : (
            <>
              <Table striped bordered hover responsive>
                <thead>
                  <tr>
                    <th>客户名称</th>
                    <th>客户类型</th>
                    <th>联系方式</th>
                    <th>客户状态</th>
                    <th>客户来源</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  {clients.map((client) => (
                    <tr key={client.id}>
                      <td>
                        <div className="d-flex align-items-center">
                          <div className="bg-light rounded-circle d-flex align-items-center justify-content-center me-3" style={{ width: '40px', height: '40px' }}>
                            {client.type === 'PERSON' ?
                              <FaUser className="text-muted" /> :
                              <FaBuilding className="text-muted" />
                            }
                          </div>
                          <div>
                            <div className="fw-bold">{client.name}</div>
                            {client.type === 'COMPANY' && client.contactPerson && (
                              <small className="text-muted">联系人: {client.contactPerson}</small>
                            )}
                          </div>
                        </div>
                      </td>
                      <td>{getTypeBadge(client.type || '')}</td>
                      <td>
                        <div className="mb-1">
                          <FaPhone className="me-1 text-muted" />
                          {client.phone}
                        </div>
                        <div>
                          <FaEnvelope className="me-1 text-muted" />
                          {client.email}
                        </div>
                      </td>
                      <td>{getStatusBadge(client.status || '')}</td>
                      <td>
                        {client.source && (
                          <Badge bg="secondary">
                            <FaStar className="me-1" />
                            {client.source}
                          </Badge>
                        )}
                      </td>
                      <td>
                        <div className="d-flex gap-1">
                          <Button
                            variant="outline-primary"
                            size="sm"
                            onClick={() => handleView(client)}
                            title="查看详情"
                          >
                            <FaEye />
                          </Button>
                          <Button
                            variant="outline-warning"
                            size="sm"
                            onClick={() => handleEdit(client)}
                            title="编辑"
                          >
                            <FaEdit />
                          </Button>
                          <Button
                            variant="outline-danger"
                            size="sm"
                            onClick={() => handleDelete(client)}
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

      {/* 编辑客户模态框 */}
      <Modal show={modalVisible} onHide={() => setModalVisible(false)} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>{modalTitle}</Modal.Title>
        </Modal.Header>
        <Form onSubmit={handleSubmit}>
          <Modal.Body>
            <Row>
              <Col md={12}>
                <Form.Group className="mb-3">
                  <Form.Label>客户类型 *</Form.Label>
                  <div>
                    <Form.Check
                      inline
                      type="radio"
                      label="个人"
                      name="type"
                      value="PERSON"
                      checked={editingClient?.type === 'PERSON'}
                      onChange={handleInputChange}
                      id="type-person"
                    />
                    <Form.Check
                      inline
                      type="radio"
                      label="企业"
                      name="type"
                      value="COMPANY"
                      checked={editingClient?.type === 'COMPANY'}
                      onChange={handleInputChange}
                      id="type-company"
                    />
                  </div>
                </Form.Group>
              </Col>
            </Row>

            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>
                    {editingClient?.type === 'COMPANY' ? '公司名称 *' : '客户姓名 *'}
                  </Form.Label>
                  <Form.Control
                    type="text"
                    name="name"
                    value={editingClient?.name || ''}
                    onChange={handleInputChange}
                    required
                  />
                </Form.Group>
              </Col>
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
                      value={editingClient?.phone || ''}
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
                  <Form.Label>电子邮箱 *</Form.Label>
                  <InputGroup>
                    <InputGroup.Text>
                      <FaEnvelope />
                    </InputGroup.Text>
                    <Form.Control
                      type="email"
                      name="email"
                      value={editingClient?.email || ''}
                      onChange={handleInputChange}
                      required
                    />
                  </InputGroup>
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>客户状态</Form.Label>
                  <Form.Select
                    name="status"
                    value={editingClient?.status || ''}
                    onChange={handleInputChange}
                  >
                    <option value="active">活跃</option>
                    <option value="inactive">非活跃</option>
                  </Form.Select>
                </Form.Group>
              </Col>
            </Row>

            {editingClient?.type === 'PERSON' && (
              <Row>
                <Col md={12}>
                  <Form.Group className="mb-3">
                    <Form.Label>身份证号</Form.Label>
                    <InputGroup>
                      <InputGroup.Text>
                        <FaIdCard />
                      </InputGroup.Text>
                      <Form.Control
                        type="text"
                        name="idCard"
                        value={editingClient?.idCard || ''}
                        onChange={handleInputChange}
                        placeholder="请输入身份证号"
                      />
                    </InputGroup>
                  </Form.Group>
                </Col>
              </Row>
            )}

            {editingClient?.type === 'COMPANY' && (
              <>
                <Row>
                  <Col md={6}>
                    <Form.Group className="mb-3">
                      <Form.Label>所属行业</Form.Label>
                      <InputGroup>
                        <InputGroup.Text>
                          <FaBriefcase />
                        </InputGroup.Text>
                        <Form.Control
                          type="text"
                          name="industry"
                          value={editingClient?.industry || ''}
                          onChange={handleInputChange}
                          placeholder="请输入所属行业"
                        />
                      </InputGroup>
                    </Form.Group>
                  </Col>
                  <Col md={6}>
                    <Form.Group className="mb-3">
                      <Form.Label>联系人</Form.Label>
                      <Form.Control
                        type="text"
                        name="contactPerson"
                        value={editingClient?.contactPerson || ''}
                        onChange={handleInputChange}
                        placeholder="请输入联系人"
                      />
                    </Form.Group>
                  </Col>
                </Row>
                <Row>
                  <Col md={12}>
                    <Form.Group className="mb-3">
                      <Form.Label>联系电话</Form.Label>
                      <InputGroup>
                        <InputGroup.Text>
                          <FaPhone />
                        </InputGroup.Text>
                        <Form.Control
                          type="tel"
                          name="contactPhone"
                          value={editingClient?.contactPhone || ''}
                          onChange={handleInputChange}
                          placeholder="请输入联系电话"
                        />
                      </InputGroup>
                    </Form.Group>
                  </Col>
                </Row>
              </>
            )}

            <Row>
              <Col md={12}>
                <Form.Group className="mb-3">
                  <Form.Label>地址 *</Form.Label>
                  <InputGroup>
                    <InputGroup.Text>
                      <FaMapMarkerAlt />
                    </InputGroup.Text>
                    <Form.Control
                      as="textarea"
                      rows={2}
                      name="address"
                      value={editingClient?.address || ''}
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
                  <Form.Label>客户来源</Form.Label>
                  <Form.Select
                    name="source"
                    value={editingClient?.source || ''}
                    onChange={handleInputChange}
                  >
                    <option value="">请选择客户来源</option>
                    <option value="推荐">推荐</option>
                    <option value="自主开发">自主开发</option>
                    <option value="网络推广">网络推广</option>
                    <option value="合作机构">合作机构</option>
                    <option value="其他">其他</option>
                  </Form.Select>
                </Form.Group>
              </Col>
            </Row>

            <Row>
              <Col md={12}>
                <Form.Group className="mb-3">
                  <Form.Label>备注</Form.Label>
                  <Form.Control
                    as="textarea"
                    rows={3}
                    name="remark"
                    value={editingClient?.remark || ''}
                    onChange={handleInputChange}
                    placeholder="请输入备注信息"
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
              {editingClient?.id ? '更新' : '新增'}
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>
    </div>
  );
};

export default ClientManagementPage;