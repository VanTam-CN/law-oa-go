import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Button, Table, Form, Badge, Modal, Row, Col, Tabs } from 'react-bootstrap';
import { getCases, createCase, updateCase, deleteCase } from '../services/caseService';
import { getClients } from '../services/clientService';
import { getLawyers } from '../services/userService';
import { Case, Client, UserProfile } from '../types';
import { logApiCallStart, logApiCall, logDataValidation, logRender, logPerformance, logError } from '../utils/devToolsValidation';
import DevToolsPanel from '../components/DevToolsPanel';
import { useToast } from '../components/Toast';
import errorHandler from '../utils/errorHandler';
import { validateCase, validateWithRules, ValidationRule } from '../utils/validation';
import useCases, { preloadCases } from '../hooks/useCases';
import PerformanceMonitor from '../components/PerformanceMonitor';
import {
  FaPlus,
  FaEdit,
  FaTrash,
  FaEye,
  FaFilter,
  FaCog,
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
  const toast = useToast();

  // 使用useCases hook管理案件数据
  const {
    cases,
    loading: casesLoading,
    error: casesError,
    pagination,
    hasMore,
    refresh: refreshCases,
    search: searchCases,
    reset: resetCases
  } = useCases({
    pageSize: 10,
    cacheTime: 5 * 60 * 1000, // 5分钟缓存
    retryCount: 3,
    retryDelay: 1000,
    enableCache: true
  });

  const [showModal, setShowModal] = useState(false);
  const [editingCase, setEditingCase] = useState<Case | null>(null);
  const [clients, setClients] = useState<Client[]>([]);
  const [lawyers, setLawyers] = useState<UserProfile[]>([]);
  const [dataLoading, setDataLoading] = useState(false);

  // 错误状态管理
  const [error, setError] = useState<string | null>(null);
  const [clientsError, setClientsError] = useState<string | null>(null);
  const [lawyersError, setLawyersError] = useState<string | null>(null);

  // 操作状态管理
  const [submitting, setSubmitting] = useState(false);
  const [deleting, setDeleting] = useState<number | null>(null);

  // 表单验证状态
  const [formErrors, setFormErrors] = useState<Record<string, string>>({});
  const [formTouched, setFormTouched] = useState<Record<string, boolean>>({});

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
  const [debouncedSearchText, setDebouncedSearchText] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [typeFilter, setTypeFilter] = useState('');
  const [lawyerFilter, setLawyerFilter] = useState('');
  const [clientFilter, setClientFilter] = useState('');
  const [dateRangeFilter, setDateRangeFilter] = useState<[string, string] | null>(null);

  

  useEffect(() => {
    fetchClients();
    fetchLawyers();
  }, []);

  // 搜索防抖处理
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearchText(searchText);
    }, 500); // 500ms防抖

    return () => clearTimeout(timer);
  }, [searchText]);

  // 当防抖搜索文本或分页变化时，重新获取数据
  useEffect(() => {
    if (pagination.current === 1) {
      fetchCases();
    } else {
      // 当搜索时重置到第一页
      setPagination(prev => ({ ...prev, current: 1 }));
    }
  }, [debouncedSearchText, statusFilter, typeFilter, lawyerFilter, clientFilter]);

  // 当分页变化时获取数据
  useEffect(() => {
    if (pagination.current > 1) {
      fetchCases();
    }
  }, [pagination.current]);

  const fetchCases = async () => {
    const startTime = performance.now();

    const requestParams = {
      page: pagination.current,
      page_size: pagination.pageSize,
      search: debouncedSearchText,
      status: statusFilter,
      case_type: typeFilter,
      lawyer_id: lawyerFilter ? parseInt(lawyerFilter) : undefined,
      client_id: clientFilter ? parseInt(clientFilter) : undefined
    };

    logApiCallStart('GET', '/cases', requestParams);

    try {
      // 使用useCases hook的搜索功能
      await searchCases(requestParams);

      const endTime = performance.now();
      logPerformance('Fetch Cases via Hook', endTime - startTime, 'ms');
      logApiCall('GET', '/cases', requestParams, { data: cases }, 200);

      // 验证数据
      logDataValidation('Cases', cases, ['id', 'title', 'client_name', 'lawyer_name', 'status', 'priority']);
    } catch (error) {
      const endTime = performance.now();
      logError('API Error', error instanceof Error ? error : String(error), { requestParams, duration: endTime - startTime });

      console.error('获取案件列表失败:', error);
      // 错误处理已经在useCases hook中完成
    }
  };

  const fetchClients = async () => {
    const startTime = performance.now();
    setDataLoading(true);
    setClientsError(null);

    const requestParams = { page: 1, page_size: 100 };
    logApiCallStart('GET', '/clients', requestParams);

    try {
      const response = await getClients(requestParams);

      const endTime = performance.now();
      logApiCall('GET', '/clients', requestParams, response, 200);
      logPerformance('Fetch Clients', endTime - startTime, 'ms');

      setClients(response.data);
      logDataValidation('Clients', response.data, ['id', 'name', 'email', 'phone']);
    } catch (error) {
      const endTime = performance.now();
      logError('API Error', error instanceof Error ? error : String(error), { requestParams, duration: endTime - startTime });

      console.error('获取客户列表失败:', error);
      const errorMessage = error instanceof Error ? error.message : '获取客户列表失败';
      setClientsError(errorMessage);
      setClients([]); // 清空数据

      // 使用全局错误处理器显示用户友好的错误提示
      errorHandler.handleNetworkError(error, '获取客户列表');
    } finally {
      setDataLoading(false);
    }
  };

  const fetchLawyers = async () => {
    const startTime = performance.now();
    setDataLoading(true);
    setLawyersError(null);

    const requestParams = { role: 'lawyer', page_size: 100 };
    logApiCallStart('GET', '/admin/users', requestParams);

    try {
      const response = await getLawyers();

      const endTime = performance.now();
      logApiCall('GET', '/admin/users', requestParams, response, 200);
      logPerformance('Fetch Lawyers', endTime - startTime, 'ms');

      setLawyers(response.data);
      logDataValidation('Lawyers', response.data, ['id', 'name', 'email', 'role']);
    } catch (error) {
      const endTime = performance.now();
      logError('API Error', error instanceof Error ? error : String(error), { requestParams, duration: endTime - startTime });

      console.error('获取律师列表失败:', error);
      const errorMessage = error instanceof Error ? error.message : '获取律师列表失败';
      setLawyersError(errorMessage);
      setLawyers([]); // 清空数据

      // 使用全局错误处理器显示用户友好的错误提示
      errorHandler.handleNetworkError(error, '获取律师列表');
    } finally {
      setDataLoading(false);
    }
  };

  // 表单验证函数
  const validateForm = (): boolean => {
    const errors: Record<string, string> = {};

    // 使用新的验证工具进行详细验证
    const validationRules: ValidationRule[] = [
      {
        field: 'title',
        required: true,
        type: 'string',
        minLength: 2,
        maxLength: 200
      },
      {
        field: 'description',
        type: 'string',
        maxLength: 5000
      },
      {
        field: 'client_id',
        required: true,
        type: 'number',
        min: 1
      },
      {
        field: 'lawyer_id',
        type: 'number',
        min: 1
      },
      {
        field: 'case_type',
        required: true,
        type: 'string'
      },
      {
        field: 'priority',
        required: true,
        type: 'string'
      },
      {
        field: 'status',
        required: true,
        type: 'string'
      },
      {
        field: 'case_amount',
        type: 'number',
        min: 0,
        max: 999999999
      },
      {
        field: 'start_date',
        type: 'date'
      },
      {
        field: 'expected_end_date',
        type: 'date',
        custom: (value) => {
          if (form.start_date && value) {
            const startDate = new Date(form.start_date);
            const endDate = new Date(value);
            if (startDate > endDate) {
              return '结束日期不能早于开始日期';
            }
          }
          return null;
        }
      }
    ];

    const validationResult = validateWithRules(form, validationRules);

    // 转换验证错误为表单错误格式
    if (validationResult.errors.length > 0) {
      validationResult.errors.forEach(error => {
        // 尝试从错误信息中提取字段名
        const fieldName = error.split('不能为空')[0].split('格式不正确')[0].split('至少需要')[0].split('不能超过')[0].split('不能小于')[0].split('不能大于')[0];

        // 映射字段名到表单字段
        const fieldMapping: Record<string, string> = {
          '案件标题': 'title',
          '客户': 'client_id',
          '律师': 'lawyer_id',
          '案件类型': 'case_type',
          '优先级': 'priority',
          '状态': 'status',
          '案件金额': 'case_amount',
          '开始日期': 'start_date',
          '预期结束日期': 'expected_end_date',
          '案件描述': 'description'
        };

        const formField = fieldMapping[fieldName] || 'unknown';
        errors[formField] = error;
      });
    }

    // 如果有警告信息，显示给用户但不阻止提交
    if (validationResult.warnings.length > 0) {
      validationResult.warnings.forEach(warning => {
        console.warn('表单验证警告:', warning);
      });
    }

    setFormErrors(errors);

    // 记录验证结果
    logDataValidation('Case Form', form, validationResult.errors.length === 0 ? Object.keys(form) : validationResult.errors);

    return validationResult.isValid;
  };

  // 重置表单错误
  const resetFormErrors = () => {
    setFormErrors({});
    setFormTouched({});
  };

  // 处理表单字段变化
  const handleFormFieldChange = (field: string, value: any) => {
    setForm(prev => ({ ...prev, [field]: value }));
    setFormTouched(prev => ({ ...prev, [field]: true }));

    // 清除该字段的错误信息
    if (formErrors[field]) {
      setFormErrors(prev => ({ ...prev, [field]: '' }));
    }
  };

  const handleCreate = () => {
    setEditingCase(null);
    resetFormErrors();
    setError(null);
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
    resetFormErrors();
    setError(null);
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
      setDeleting(id);

      const apiStartTime = performance.now();
      logApiCallStart('DELETE', `/cases/${id}`);

      try {
        await deleteCase(id);

        const apiEndTime = performance.now();
        logApiCall('DELETE', `/cases/${id}`, {}, {}, 200);
        logPerformance('Delete Case', apiEndTime - apiStartTime, 'ms');

        // 刷新案件列表而不是本地删除
        refreshCases();
        setError(null);

        // 显示成功提示
        errorHandler.showSuccess('案件删除成功', '案件管理');
      } catch (error) {
        const apiEndTime = performance.now();
        logError('Delete Case Error', error instanceof Error ? error : String(error), {
          id,
          duration: apiEndTime - apiStartTime
        });

        console.error('删除失败:', error);
        const errorMessage = error instanceof Error ? error.message : '删除失败';
        setError(errorMessage);

        // 使用全局错误处理器显示用户友好的错误提示
        errorHandler.handleNetworkError(error, '删除案件');
      } finally {
        setDeleting(null);
      }
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // 表单验证
    if (!validateForm()) {
      return;
    }

    setSubmitting(true);
    setError(null);

    try {
      if (editingCase) {
        // 更新案件
        const updateData = {
          title: form.title,
          description: form.description,
          lawyer_id: form.lawyer_id || undefined,
          case_type: form.case_type,
          priority: form.priority,
          status: form.status
        };

        const apiStartTime = performance.now();
        logApiCallStart('PUT', `/cases/${editingCase.id}`, updateData);

        await updateCase(editingCase.id, updateData);

        const apiEndTime = performance.now();
        logApiCall('PUT', `/cases/${editingCase.id}`, updateData, {}, 200);
        logPerformance('Update Case', apiEndTime - apiStartTime, 'ms');

        // 成功后更新本地数据
        setCases(prev => prev.map(item =>
          item.id === editingCase.id
            ? { ...item, ...updateData }
            : item
        ));
      } else {
        // 新增案件
        const createData = {
          title: form.title,
          description: form.description,
          client_id: form.client_id,
          lawyer_id: form.lawyer_id || undefined,
          case_type: form.case_type,
          priority: form.priority,
          status: form.status
        };

        const apiStartTime = performance.now();
        logApiCallStart('POST', '/cases', createData);

        await createCase(createData);

        const apiEndTime = performance.now();
        logApiCall('POST', '/cases', createData, {}, 200);
        logPerformance('Create Case', apiEndTime - apiStartTime, 'ms');
      }

      setShowModal(false);
      refreshCases(); // 使用useCases hook的刷新功能

      // 显示成功提示
      const successMessage = editingCase ? '案件更新成功' : '案件创建成功';
      errorHandler.showSuccess(successMessage, '案件管理');
    } catch (error) {
      console.error('保存案件失败:', error);
      const errorMessage = error instanceof Error ? error.message : '保存失败，请重试';
      setError(errorMessage);

      // 使用全局错误处理器显示用户友好的错误提示
      errorHandler.handleNetworkError(error, '保存案件');
    } finally {
      setSubmitting(false);
    }
  };

  // 重试获取数据的函数
  const handleRetry = () => {
    setError(null);
    setCasesError(null);
    setClientsError(null);
    setLawyersError(null);
    refreshCases(); // 使用useCases hook的刷新功能
    fetchClients();
    fetchLawyers();
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
            disabled={deleting === record.id}
          >
            {deleting === record.id ? (
              <span className="spinner-border spinner-border-sm" role="status" />
            ) : (
              <FaTrash />
            )}
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

  // 记录组件渲染
  logRender('CaseManagement', {
    casesCount: cases.length,
    loading,
    error,
    pagination,
    filters: { searchText, statusFilter, typeFilter, lawyerFilter, clientFilter }
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
          {/* 错误显示区域 */}
          {error && (
            <div className="alert alert-danger d-flex justify-content-between align-items-center mb-3">
              <span>{error}</span>
              <Button variant="outline-danger" size="sm" onClick={handleRetry}>
                重试
              </Button>
            </div>
          )}

          {casesError && (
            <div className="alert alert-warning d-flex justify-content-between align-items-center mb-3">
              <span>案件数据加载失败: {casesError}</span>
              <Button variant="outline-warning" size="sm" onClick={() => {
                setCasesError(null);
                refreshCases(); // 使用useCases hook的刷新功能
              }}>
                重试
              </Button>
            </div>
          )}

          {(clientsError || lawyersError) && (
            <div className="alert alert-info d-flex justify-content-between align-items-center mb-3">
              <span>
                {clientsError && `客户数据加载失败: ${clientsError}`}
                {clientsError && lawyersError && ' | '}
                {lawyersError && `律师数据加载失败: ${lawyersError}`}
              </span>
              <Button variant="outline-info" size="sm" onClick={() => {
                setClientsError(null);
                setLawyersError(null);
                fetchClients();
                fetchLawyers();
              }}>
                重试
              </Button>
            </div>
          )}

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
          <div
            className="table-responsive"
            data-cases-container="true"
            ref={(el) => {
              if (el) {
                (el as any).__casesData = cases;
              }
            }}
          >
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
                    onChange={(e) => handleFormFieldChange('title', e.target.value)}
                    required
                    isInvalid={!!formErrors.title}
                  />
                  {formErrors.title && (
                    <Form.Control.Feedback type="invalid">
                      {formErrors.title}
                    </Form.Control.Feedback>
                  )}
                </Form.Group>
              </Col>
            </Row>

            <Row className="mb-3">
              <Col md={4}>
                <Form.Group>
                  <Form.Label>案件类型 *</Form.Label>
                  <Form.Select
                    value={form.case_type}
                    onChange={(e) => handleFormFieldChange('case_type', e.target.value)}
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
                    onChange={(e) => handleFormFieldChange('case_type', e.target.value)}
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
                    onChange={(e) => handleFormFieldChange('client_id', e.target.value ? parseInt(e.target.value) : 0)}
                    required
                    disabled={dataLoading}
                    isInvalid={!!formErrors.client_id}
                    data-clients-container="true"
                    ref={(el) => {
                      if (el) {
                        (el as any).__clientsData = clients;
                      }
                    }}
                  >
                    <option value="">请选择客户</option>
                    {clients.map((client) => (
                      <option key={client.id} value={client.id.toString()}>
                        {client.company ? client.company : client.name}
                      </option>
                    ))}
                  </Form.Select>
                  {formErrors.client_id && (
                    <Form.Control.Feedback type="invalid">
                      {formErrors.client_id}
                    </Form.Control.Feedback>
                  )}
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group>
                  <Form.Label>负责律师 *</Form.Label>
                  <Form.Select
                    value={form.lawyer_id?.toString() || ''}
                    onChange={(e) => handleFormFieldChange('lawyer_id', e.target.value ? parseInt(e.target.value) : null)}
                    required
                    disabled={dataLoading}
                    data-lawyers-container="true"
                    ref={(el) => {
                      if (el) {
                        (el as any).__lawyersData = lawyers;
                      }
                    }}
                  >
                    <option value="">请选择负责律师</option>
                    {lawyers.map((lawyer) => (
                      <option key={lawyer.id} value={lawyer.id.toString()}>
                        {lawyer.name}
                      </option>
                    ))}
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
              <Button variant="primary" type="submit" disabled={submitting}>
                {submitting ? '保存中...' : (editingCase ? '更新案件' : '创建案件')}
              </Button>
            </div>
          </Form>
        </Modal.Body>
      </Modal>

      {/* 开发工具面板 */}
      <DevToolsPanel />

      {/* 性能监控组件 */}
      <PerformanceMonitor enabled={process.env.NODE_ENV === 'development'} showDetails={false} />
    </div>
  );
};

export default CaseManagement;