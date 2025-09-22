import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Button, Table, Form, Badge, Modal, Row, Col, Tabs } from 'react-bootstrap';
import { getCases, createCase, updateCase, deleteCase } from '../services/caseService';
import {
  FaPlus,
  FaEdit,
  FaTrash,
  FaEye,
  FaFilter,
  FaClockRotateLeft,
  FaGear,
  FaFile,
  FaUser,
  FaCalendar,
  FaClock,
  FaCheckCircle,
  FaTimesCircle,
  FaSync
} from 'react-icons/fa';
import {
  MagnifyingGlassIcon
} from "@heroicons/react/24/outline";

interface Case {
  id: number;
  case_number?: string;
  title: string;
  description: string;
  client_id: number;
  client_name?: string;
  lawyer_id: number | null;
  lawyer_name?: string;
  case_type: string;
  priority: string;
  status: string;
  start_date?: string;
  end_date?: string | null;
  created_at: string;
  updated_at: string;
  case_amount?: number;
  expected_end_date?: string;
  principal_info?: string;
  opponent_info?: string;
}

interface CaseFormData {
  case_number?: string;
  title: string;
  description: string;
  client_id: number;
  lawyer_id?: number;
  case_type: string;
  priority: string;
  status: string;
  start_date?: string;
  expected_end_date?: string;
  case_amount?: number;
  principal_info?: string;
  opponent_info?: string;
}

const CaseManagement: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [cases, setCases] = useState<Case[]>([]);
  const [showModal, setShowModal] = useState(false);
  const [editingCase, setEditingCase] = useState<Case | null>(null);
  const [form, setForm] = useState({
    case_number: '',
    title: '',
    case_type: 'CIVIL',
    client_id: 0 as number,
    lawyer_id: null as number | null,
    priority: 'medium',
    status: 'pending',
    description: '',
    start_date: '',
    expected_end_date: '',
    case_amount: null as number | null,
    principal_info: '',
    opponent_info: ''
  });

  // 搜索和过滤状态
  const [searchText, setSearchText] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [typeFilter, setTypeFilter] = useState('');
  const [lawyerFilter, setLawyerFilter] = useState('');
  const [clientFilter, setClientFilter] = useState('');
  const [dateRangeFilter, setDateRangeFilter] = useState<[string, string] | null>(null);

  // 分页状态
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });

  
  useEffect(() => {
    fetchCases();
  }, []);

  const fetchCases = async () => {
    setLoading(true);
    try {
      // 使用真实API调用
      const response = await getCases({
        page: pagination.current,
        page_size: pagination.pageSize,
        search: searchText,
        status: statusFilter,
        case_type: typeFilter
      });

      setCases(response.data);
      setPagination({
        current: response.pagination.page,
        pageSize: response.pagination.page_size,
        total: response.pagination.total
      });
    } catch (error) {
      console.error('获取案件列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setEditingCase(null);
    setForm({
      case_number: '',
      title: '',
      case_type: 'CIVIL',
      client_id: 0,
      lawyer_id: null,
      priority: 'medium',
      status: 'pending',
      description: '',
      start_date: '',
      expected_end_date: '',
      case_amount: null,
      principal_info: '',
      opponent_info: ''
    });
    setShowModal(true);
  };

  const handleEdit = (record: Case) => {
    setEditingCase(record);
    setForm({
      case_number: record.case_number || '',
      title: record.title,
      case_type: record.case_type,
      client_id: record.client_id,
      lawyer_id: record.lawyer_id,
      priority: record.priority,
      status: record.status,
      description: record.description,
      start_date: record.start_date || '',
      expected_end_date: record.expected_end_date || '',
      case_amount: record.case_amount || null,
      principal_info: record.principal_info || '',
      opponent_info: record.opponent_info || ''
    });
    setShowModal(true);
  };

  const handleDelete = async (id: number) => {
    if (window.confirm('确定要删除这个案件吗？')) {
      try {
        await deleteCase(id);
        setCases(cases.filter(item => item.id !== id));
        alert('删除成功');
      } catch (error) {
        console.error('删除失败:', error);
        alert('删除失败');
      }
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setLoading(true);

      if (editingCase) {
        // 更新案件
        const updateData = {
          ...form,
          client_id: form.client_id || undefined,
          lawyer_id: form.lawyer_id || undefined,
          case_amount: form.case_amount || undefined
        };
        await updateCase(editingCase.id, updateData);
        alert('案件更新成功');
      } else {
        // 新增案件
        await createCase({
          ...form,
          lawyer_id: form.lawyer_id || undefined,
          case_amount: form.case_amount || undefined
        });
        alert('案件创建成功');
      }

      setShowModal(false);
      fetchCases();
    } catch (error) {
      console.error('保存案件失败:', error);
      alert('保存失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const getStatusBadge = (status: string) => {
    const statusMap = {
      '0': { text: '未开始', color: 'secondary' },
      '1': { text: '进行中', color: 'primary' },
      '2': { text: '已结案', color: 'success' },
      '3': { text: '已归档', color: 'secondary' }
    };
    const config = statusMap[status as keyof typeof statusMap] || { text: '未知', color: 'secondary' };
    return <Badge bg={config.color}>{config.text}</Badge>;
  };

  const getCaseTypeTag = (type: string) => {
    const typeMap = {
      'CIVIL': { text: '民事案件', color: 'primary' },
      'COMMERCIAL': { text: '商事案件', color: 'warning' },
      'CRIMINAL': { text: '刑事案件', color: 'danger' },
      'ADMINISTRATIVE': { text: '行政案件', color: 'info' },
      'ADVISORY': { text: '咨询项目', color: 'success' },
      'REVIEW': { text: '审查项目', color: 'secondary' }
    };
    const config = typeMap[type as keyof typeof typeMap] || { text: '其他', color: 'secondary' };
    return <Badge bg={config.color}>{config.text}</Badge>;
  };

  const columns = [
    {
      header: '案件编号',
      field: 'case_number',
      render: (text: string) => (
        <div className="d-flex align-items-center">
          <FaFile className="me-2" />
          <span>{text}</span>
        </div>
      )
    },
    {
      header: '案件名称',
      field: 'title',
      render: (text: string) => <span title={text}>{text}</span>
    },
    {
      header: '案件类型',
      field: 'case_type',
      render: (text: string) => getCaseTypeTag(text)
    },
    {
      header: '客户',
      field: 'client_name',
      render: (text: string) => (
        <div className="d-flex align-items-center">
          <FaUser className="me-2" />
          <span>{text}</span>
        </div>
      )
    },
    {
      header: '负责律师',
      field: 'lawyer_name',
      render: (text: string) => <span>{text}</span>
    },
    {
      header: '状态',
      field: 'status',
      render: (text: string) => getStatusBadge(text)
    },
    {
      header: '创建时间',
      field: 'created_at',
      render: (text: string) => (
        <div className="d-flex align-items-center">
          <FaCalendar className="me-2" />
          <span>{text}</span>
        </div>
      )
    },
    {
      header: '操作',
      field: 'action',
      render: (_: any, record: Case) => (
        <div className="d-flex gap-2">
          <Button
            variant="outline-primary"
            size="sm"
            onClick={() => navigate(`/cases/${record.id}`)}
            title="查看详情"
          >
            <FaEye />
          </Button>
          <Button
            variant="outline-secondary"
            size="sm"
            onClick={() => handleEdit(record)}
            title="编辑"
          >
            <FaEdit />
          </Button>
          <Button
            variant="outline-danger"
            size="sm"
            onClick={() => handleDelete(record.id)}
            title="删除"
          >
            <FaTrash />
          </Button>
        </div>
      )
    }
  ];

  // 应用搜索和过滤
  const filteredCases = cases.filter(case_item => {
    const matchesSearch = !searchText ||
      case_item.title.toLowerCase().includes(searchText.toLowerCase()) ||
      (case_item.case_number && case_item.case_number.toLowerCase().includes(searchText.toLowerCase())) ||
      (case_item.client_name && case_item.client_name.toLowerCase().includes(searchText.toLowerCase()));

    const matchesStatus = !statusFilter || case_item.status === statusFilter;
    const matchesType = !typeFilter || case_item.case_type === typeFilter;
    const matchesLawyer = !lawyerFilter || (case_item.lawyer_name && case_item.lawyer_name.includes(lawyerFilter));
    const matchesClient = !clientFilter || (case_item.client_name && case_item.client_name.includes(clientFilter));

    return matchesSearch && matchesStatus && matchesType && matchesLawyer && matchesClient;
  });

  return (
    <div className="case-management">
      <Card>
        <Card.Header>
          <div className="d-flex justify-content-between align-items-center">
            <h4 className="mb-0">案件管理</h4>
            <Button variant="primary" onClick={handleCreate}>
              <FaPlus className="me-2" />
              新建案件
            </Button>
          </div>
        </Card.Header>
        <Card.Body>
          {/* 搜索和过滤区域 */}
          <Row className="mb-3">
            <Col md={6}>
              <div className="input-group">
                <span className="input-group-text">
                  <MagnifyingGlassIcon className="w-4 h-4" />
                </span>
                <Form.Control
                  type="text"
                  placeholder="搜索案件名称、编号或客户"
                  value={searchText}
                  onChange={(e) => setSearchText(e.target.value)}
                />
              </div>
            </Col>
            <Col md={2}>
              <Form.Select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
              >
                <option value="">所有状态</option>
                <option value="0">未开始</option>
                <option value="1">进行中</option>
                <option value="2">已结案</option>
                <option value="3">已归档</option>
              </Form.Select>
            </Col>
            <Col md={2}>
              <Form.Select
                value={typeFilter}
                onChange={(e) => setTypeFilter(e.target.value)}
              >
                <option value="">所有类型</option>
                <option value="CIVIL">民事案件</option>
                <option value="COMMERCIAL">商事案件</option>
                <option value="CRIMINAL">刑事案件</option>
                <option value="ADMINISTRATIVE">行政案件</option>
                <option value="ADVISORY">咨询项目</option>
                <option value="REVIEW">审查项目</option>
              </Form.Select>
            </Col>
            <Col md={2}>
              <Button variant="outline-secondary" className="w-100">
                <FaFilter className="me-2" />
                高级搜索
              </Button>
            </Col>
          </Row>

          {/* 快速过滤标签 */}
          <div className="mb-3">
            <span className="fw-bold me-2">快速过滤：</span>
            {['0', '1', '2'].map(status => (
              <Button
                key={status}
                variant={statusFilter === status ? 'primary' : 'outline-primary'}
                size="sm"
                className="me-2 mb-1"
                onClick={() => setStatusFilter(statusFilter === status ? '' : status)}
              >
                {status === '0' ? '未开始' : status === '1' ? '进行中' : '已结案'}
              </Button>
            ))}
            {['CIVIL', 'COMMERCIAL', 'CRIMINAL'].map(type => (
              <Button
                key={type}
                variant={typeFilter === type ? 'warning' : 'outline-warning'}
                size="sm"
                className="me-2 mb-1"
                onClick={() => setTypeFilter(typeFilter === type ? '' : type)}
              >
                {type === 'CIVIL' ? '民事' : type === 'COMMERCIAL' ? '商事' : '刑事'}
              </Button>
            ))}
          </div>

          {/* 搜索条件显示 */}
          {(searchText || statusFilter || typeFilter) && (
            <div className="alert alert-info d-flex justify-content-between align-items-center mb-3">
              <div>
                <strong>当前搜索条件：</strong>
                <div className="d-inline-block ms-2">
                  {searchText && <Badge bg="info" className="me-2">关键词: {searchText}</Badge>}
                  {statusFilter && (
                    <Badge bg="success" className="me-2">
                      状态: {statusFilter === '0' ? '未开始' : statusFilter === '1' ? '进行中' : statusFilter === '2' ? '已结案' : '已归档'}
                    </Badge>
                  )}
                  {typeFilter && (
                    <Badge bg="warning" className="me-2">
                      类型: {typeFilter === 'CIVIL' ? '民事' : typeFilter === 'COMMERCIAL' ? '商事' : typeFilter === 'CRIMINAL' ? '刑事' : typeFilter === 'ADMINISTRATIVE' ? '行政' : typeFilter === 'ADVISORY' ? '咨询' : '审查'}
                    </Badge>
                  )}
                </div>
              </div>
              <Button
                variant="link"
                size="sm"
                onClick={() => {
                  setSearchText('');
                  setStatusFilter('');
                  setTypeFilter('');
                }}
              >
                清除搜索
              </Button>
            </div>
          )}

          {/* 案件表格 */}
          <div className="table-responsive">
            <Table striped bordered hover>
              <thead>
                <tr>
                  {columns.map(col => (
                    <th key={col.field}>{col.header}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <tr>
                    <td colSpan={columns.length} className="text-center">
                      <div className="spinner-border" role="status">
                        <span className="visually-hidden">加载中...</span>
                      </div>
                    </td>
                  </tr>
                ) : filteredCases.length === 0 ? (
                  <tr>
                    <td colSpan={columns.length} className="text-center">
                      暂无数据
                    </td>
                  </tr>
                ) : (
                  filteredCases.map(case_item => (
                    <tr key={case_item.id}>
                      {columns.map(col => (
                        <td key={col.field}>
                          {col.render ? col.render(case_item[col.field as keyof Case], case_item) : case_item[col.field as keyof Case]}
                        </td>
                      ))}
                    </tr>
                  ))
                )}
              </tbody>
            </Table>
          </div>

          {/* 分页 */}
          <div className="d-flex justify-content-between align-items-center mt-3">
            <div>
              显示 {(pagination.current - 1) * pagination.pageSize + 1} 到{' '}
              {Math.min(pagination.current * pagination.pageSize, pagination.total)} 条，共 {pagination.total} 条
            </div>
            <div className="d-flex gap-2">
              <Button
                variant="outline-primary"
                disabled={pagination.current === 1}
                onClick={() => setPagination({...pagination, current: pagination.current - 1})}
              >
                上一页
              </Button>
              <Button
                variant="outline-primary"
                disabled={pagination.current * pagination.pageSize >= pagination.total}
                onClick={() => setPagination({...pagination, current: pagination.current + 1})}
              >
                下一页
              </Button>
            </div>
          </div>
        </Card.Body>
      </Card>

      {/* 案件创建/编辑模态框 */}
      <Modal show={showModal} onHide={() => setShowModal(false)} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>{editingCase ? '编辑案件' : '新建案件'}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <Form onSubmit={handleSubmit}>
            <Row className="mb-3">
              <Col md={6}>
                <Form.Group>
                  <Form.Label>案件编号 *</Form.Label>
                  <Form.Control
                    type="text"
                    value={form.case_number}
                    onChange={(e) => setForm({...form, case_number: e.target.value})}
                    required
                  />
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group>
                  <Form.Label>案件名称 *</Form.Label>
                  <Form.Control
                    type="text"
                    value={form.title}
                    onChange={(e) => setForm({...form, title: e.target.value})}
                    required
                  />
                </Form.Group>
              </Col>
            </Row>

            <Row className="mb-3">
              <Col md={4}>
                <Form.Group>
                  <Form.Label>案件类型 *</Form.Label>
                  <Form.Select
                    value={form.case_type}
                    onChange={(e) => setForm({...form, case_type: e.target.value})}
                    required
                  >
                    <option value="CIVIL">民事案件</option>
                    <option value="COMMERCIAL">商事案件</option>
                    <option value="CRIMINAL">刑事案件</option>
                    <option value="ADMINISTRATIVE">行政案件</option>
                    <option value="ADVISORY">咨询项目</option>
                    <option value="REVIEW">审查项目</option>
                  </Form.Select>
                </Form.Group>
              </Col>
              <Col md={4}>
                <Form.Group>
                  <Form.Label>状态 *</Form.Label>
                  <Form.Select
                    value={form.status}
                    onChange={(e) => setForm({...form, status: e.target.value})}
                    required
                  >
                    <option value="0">未开始</option>
                    <option value="1">进行中</option>
                    <option value="2">已结案</option>
                    <option value="3">已归档</option>
                  </Form.Select>
                </Form.Group>
              </Col>
              <Col md={4}>
                <Form.Group>
                  <Form.Label>项目类型</Form.Label>
                  <Form.Select
                    value={form.case_type}
                    onChange={(e) => setForm({...form, case_type: e.target.value})}
                  >
                    <option value="">请选择项目类型</option>
                    <option value="CIVIL">民事诉讼</option>
                    <option value="COMMERCIAL">商业诉讼</option>
                    <option value="CRIMINAL">刑事诉讼</option>
                    <option value="ADVISORY">法律顾问</option>
                    <option value="REVIEW">合同审查</option>
                  </Form.Select>
                </Form.Group>
              </Col>
            </Row>

            <Row className="mb-3">
              <Col md={6}>
                <Form.Group>
                  <Form.Label>客户 *</Form.Label>
                  <Form.Select
                    value={form.client_id?.toString() || ''}
                    onChange={(e) => setForm({...form, client_id: e.target.value ? parseInt(e.target.value) : 0})}
                    required
                  >
                    <option value="">请选择客户</option>
                    <option value="1">张三</option>
                    <option value="2">王五</option>
                    <option value="3">ABC科技有限公司</option>
                  </Form.Select>
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group>
                  <Form.Label>负责律师 *</Form.Label>
                  <Form.Select
                    value={form.lawyer_id?.toString() || ''}
                    onChange={(e) => setForm({...form, lawyer_id: e.target.value ? parseInt(e.target.value) : null})}
                    required
                  >
                    <option value="">请选择负责律师</option>
                    <option value="1">张律师</option>
                    <option value="2">杨律师</option>
                    <option value="3">王律师</option>
                  </Form.Select>
                </Form.Group>
              </Col>
            </Row>

            <Row className="mb-3">
              <Col md={4}>
                <Form.Group>
                  <Form.Label>项目编号</Form.Label>
                  <Form.Control
                    type="text"
                    value={form.case_number}
                    onChange={(e) => setForm({...form, case_number: e.target.value})}
                  />
                </Form.Group>
              </Col>
              <Col md={4}>
                <Form.Group>
                  <Form.Label>项目金额</Form.Label>
                  <Form.Control
                    type="number"
                    value={form.case_amount || ''}
                    onChange={(e) => setForm({...form, case_amount: e.target.value ? parseFloat(e.target.value) : null})}
                    placeholder="请输入案件金额"
                  />
                </Form.Group>
              </Col>
              <Col md={4}>
                <Form.Group>
                  <Form.Label>优先级</Form.Label>
                  <Form.Select
                    value={form.priority}
                    onChange={(e) => setForm({...form, priority: e.target.value})}
                  >
                    <option value="low">低</option>
                    <option value="medium">中</option>
                    <option value="high">高</option>
                  </Form.Select>
                </Form.Group>
              </Col>
            </Row>

            <Row className="mb-3">
              <Col md={6}>
                <Form.Group>
                  <Form.Label>开始日期</Form.Label>
                  <Form.Control
                    type="date"
                    value={form.start_date}
                    onChange={(e) => setForm({...form, start_date: e.target.value})}
                  />
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group>
                  <Form.Label>预期结束日期</Form.Label>
                  <Form.Control
                    type="date"
                    value={form.expected_end_date}
                    onChange={(e) => setForm({...form, expected_end_date: e.target.value})}
                  />
                </Form.Group>
              </Col>
            </Row>

            <Form.Group className="mb-3">
              <Form.Label>案件描述 *</Form.Label>
              <Form.Control
                as="textarea"
                rows={4}
                value={form.description}
                onChange={(e) => setForm({...form, description: e.target.value})}
                required
              />
            </Form.Group>

            <div className="d-flex justify-content-end gap-2">
              <Button variant="secondary" onClick={() => setShowModal(false)}>
                取消
              </Button>
              <Button variant="primary" type="submit" disabled={loading}>
                {loading ? '保存中...' : (editingCase ? '更新案件' : '创建案件')}
              </Button>
            </div>
          </Form>
        </Modal.Body>
      </Modal>
    </div>
  );
};

export default CaseManagement;