import React, { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import {
  Row,
  Col,
  Card,
  Button,
  Spinner,
  Badge,
  Tabs,
  Tab,
  Dropdown,
  Form,
  InputGroup,
  ListGroup,
  Modal,
  Table,
  Alert
} from "react-bootstrap";
import {
  FaArrowLeft,
  FaPenToSquare,
  FaTrash,
  FaFileLines,
  FaFile,
  FaUser,
  FaCalendar,
  FaClock,
  FaDollarSign,
  FaPhone,
  FaEnvelope,
  FaLocationPin,
  FaComment,
  FaPaperclip,
  FaArrowsRotate,
  FaGavel,
  FaPeopleGroup,
  FaPlus,
  FaDownload,
  FaXmark,
  FaCircleInfo
} from "react-icons/fa6";
import { caseService } from "../services/caseService";
import { Case } from "../types";
import { useToast } from "../components/Toast";

// 使用后端API的Case类型，前端显示用的扩展属性
interface CaseDetail extends Case {
  clientName?: string;
  clientPhone?: string;
  clientEmail?: string;
  clientAddress?: string;
  lawyerName?: string;
  lawyerPhone?: string;
  lawyerEmail?: string;
  principalInfo?: string;
  opponentInfo?: string;
  caseId?: number; // 兼容旧代码
}

// 案件文档接口
interface CaseDocument {
  id: number;
  name: string;
  type: string;
  size: string;
  uploadTime: string;
  uploader: string;
}

// 案件时间线接口
interface CaseTimeline {
  id: number;
  time: string;
  event: string;
  user: string;
  description: string;
}

// 案件费用接口
interface CaseExpense {
  id: number;
  date: string;
  type: string;
  amount: number;
  description: string;
  status: string;
}

const CaseDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const toast = useToast();
  const [caseData, setCaseData] = useState<CaseDetail | null>(null);
  const [documents, setDocuments] = useState<CaseDocument[]>([]);
  const [timeline, setTimeline] = useState<CaseTimeline[]>([]);
  const [expenses, setExpenses] = useState<CaseExpense[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState("basic");
  const [showEditModal, setShowEditModal] = useState(false);
  const [editForm, setEditForm] = useState<Partial<CaseDetail>>({});

  // 获取案件详情
  useEffect(() => {
    if (id) {
      fetchCaseDetail(parseInt(id));
    }
  }, [id]);

  const fetchCaseDetail = async (caseId: number) => {
    setLoading(true);
    setError(null);

    try {
      // 调用真实的API获取案件详情
      const caseData = await caseService.getCase(caseId);

      // 转换数据格式以适配前端显示
      const formattedCaseData: CaseDetail = {
        ...caseData,
        caseId: caseData.id,
        clientName: caseData.client?.name || caseData.client?.company || caseData.clientName || "未知客户",
        clientPhone: caseData.client?.phone || caseData.clientPhone || "",
        clientEmail: caseData.client?.email || caseData.clientEmail || "",
        clientAddress: caseData.client?.address || caseData.clientAddress || "",
        lawyerName: caseData.lawyer?.name || caseData.lawyerName || "未分配",
        lawyerPhone: caseData.lawyer?.phone || caseData.lawyerPhone || "",
        lawyerEmail: caseData.lawyer?.email || caseData.lawyerEmail || "",
        principalInfo: caseData.principal_info || caseData.principalInfo || "",
        opponentInfo: caseData.opponent_info || caseData.opponentInfo || "",
      };

      // 设置案件数据
      setCaseData(formattedCaseData);
      setEditForm(formattedCaseData);

      // TODO: 后续添加获取文档、时间线、费用等数据的API调用
      // 目前先使用空数组，等后端接口完善后再集成
      setDocuments([]);
      setTimeline([]);
      setExpenses([]);

    } catch (error: any) {
      console.error("获取案件详情失败:", error);
      const errorMessage = error.message || "获取案件详情失败";
      setError(errorMessage);
      setCaseData(null);

      // 显示错误提示
      toast.showToast({
        type: 'error',
        title: '加载失败',
        message: errorMessage,
        persistent: true
      });
    } finally {
      setLoading(false);
    }
  };

  const handleEdit = () => {
    if (caseData) {
      setEditForm(caseData);
      setShowEditModal(true);
    }
  };

  const handleSave = async () => {
    if (!caseData) return;

    try {
      // 构建更新数据，只包含有变化的字段
      const updateData: any = {};

      if (editForm.title !== undefined && editForm.title !== caseData.title) {
        updateData.title = editForm.title;
      }
      if (editForm.description !== undefined && editForm.description !== caseData.description) {
        updateData.description = editForm.description;
      }
      if (editForm.case_type !== undefined && editForm.case_type !== caseData.case_type) {
        updateData.case_type = editForm.case_type;
      }
      if (editForm.status !== undefined && editForm.status !== caseData.status) {
        updateData.status = editForm.status;
      }
      if (editForm.priority !== undefined && editForm.priority !== caseData.priority) {
        updateData.priority = editForm.priority;
      }
      if (editForm.lawyer_id !== undefined && editForm.lawyer_id !== caseData.lawyer_id) {
        updateData.lawyer_id = editForm.lawyer_id;
      }

      // 如果没有变化，直接关闭模态框
      if (Object.keys(updateData).length === 0) {
        setShowEditModal(false);
        toast.showToast({
          type: 'info',
          title: '无变化',
          message: '案件信息没有发生变化',
          duration: 2000
        });
        return;
      }

      // 调用真实的API更新案件
      const updatedCase = await caseService.updateCase(caseData.id, updateData);

      // 更新本地数据，确保正确格式化
      const updatedFormattedData: CaseDetail = {
        ...caseData,
        ...updatedCase,
        caseId: updatedCase.id || caseData.id,
        clientName: updatedCase.client?.name || updatedCase.client?.company || caseData.clientName,
        clientPhone: updatedCase.client?.phone || caseData.clientPhone,
        clientEmail: updatedCase.client?.email || caseData.clientEmail,
        clientAddress: updatedCase.client?.address || caseData.clientAddress,
        lawyerName: updatedCase.lawyer?.name || caseData.lawyerName,
        lawyerPhone: updatedCase.lawyer?.phone || caseData.lawyerPhone,
        lawyerEmail: updatedCase.lawyer?.email || caseData.lawyerEmail,
        principalInfo: updatedCase.principal_info || caseData.principalInfo,
        opponentInfo: updatedCase.opponent_info || caseData.opponentInfo,
      };

      setCaseData(updatedFormattedData);
      setEditForm(updatedFormattedData);
      setShowEditModal(false);

      // 显示成功消息
      toast.showToast({
        type: 'success',
        title: '操作成功',
        message: '案件信息已更新',
        duration: 3000
      });
    } catch (error: any) {
      console.error("更新案件失败:", error);
      toast.showToast({
        type: 'error',
        title: '更新失败',
        message: error.message || "请重试",
        persistent: true
      });
    }
  };

  const handleDelete = async () => {
    if (!caseData) return;

    if (window.confirm(`确定要删除案件"${caseData.title}"吗？\n\n此操作不可撤销，请谨慎操作！`)) {
      try {
        // 调用真实的API删除案件
        await caseService.deleteCase(caseData.id);

        toast.showToast({
          type: 'success',
          title: '操作成功',
          message: '案件删除成功',
          duration: 2000
        });

        // 延迟导航，让用户看到成功提示
        setTimeout(() => {
          navigate("/cases");
        }, 1000);
      } catch (error: any) {
        console.error("删除案件失败:", error);
        toast.showToast({
          type: 'error',
          title: '删除失败',
          message: error.message || "请重试",
          persistent: true
        });
      }
    }
  };

  const getStatusBadge = (status: string) => {
    const statusMap = {
      'pending': { text: '待处理', variant: 'warning' },
      'active': { text: '进行中', variant: 'primary' },
      'closed': { text: '已结案', variant: 'success' },
      'suspended': { text: '已暂停', variant: 'secondary' },
      // 兼容旧的数字状态
      '0': { text: '未开始', variant: 'secondary' },
      '1': { text: '进行中', variant: 'primary' },
      '2': { text: '已结案', variant: 'success' },
      '3': { text: '已归档', variant: 'info' }
    };
    const config = statusMap[status as keyof typeof statusMap] || { text: '未知', variant: 'secondary' };
    return <Badge bg={config.variant}>{config.text}</Badge>;
  };

  const getCaseTypeBadge = (type: string) => {
    const typeMap = {
      'civil': { text: '民事案件', variant: 'primary' },
      'commercial': { text: '商事案件', variant: 'warning' },
      'criminal': { text: '刑事案件', variant: 'danger' },
      'administrative': { text: '行政案件', variant: 'info' }
    };
    const config = typeMap[type as keyof typeof typeMap] || { text: '其他', variant: 'secondary' };
    return <Badge bg={config.variant}>{config.text}</Badge>;
  };

  const getPriorityBadge = (priority: string) => {
    const priorityMap = {
      'low': { text: '低', variant: 'secondary' },
      'medium': { text: '中', variant: 'info' },
      'high': { text: '高', variant: 'warning' },
      'urgent': { text: '紧急', variant: 'danger' }
    };
    const config = priorityMap[priority as keyof typeof priorityMap] || { text: '普通', variant: 'secondary' };
    return <Badge bg={config.variant}>{config.text}</Badge>;
  };

  const formatDate = (dateString: string) => {
    if (!dateString) return "未知";
    try {
      return new Date(dateString).toLocaleString('zh-CN');
    } catch (error) {
      return dateString;
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setEditForm(prev => ({ ...prev, [name]: value }));
  };

  const handleRetry = () => {
    const caseId = parseInt(id || '0');
    if (caseId > 0) {
      toast.showToast({
        type: 'info',
        title: '正在重试',
        message: '正在重新获取案件信息...',
        duration: 1000
      });
      fetchCaseDetail(caseId);
    }
  };

  if (loading) {
    return (
      <div className="d-flex justify-content-center align-items-center" style={{ height: "50vh" }}>
        <Spinner animation="border" />
        <span className="ms-2">加载中...</span>
      </div>
    );
  }

  if (!caseData && !loading) {
    return (
      <div className="text-center py-5">
        {error ? (
          <>
            <FaCircleInfo className="fas fa-exclamation-circle fa-3x text-danger mb-3" />
            <h5>加载失败</h5>
            <p className="text-muted">{error}</p>
            <Button variant="primary" onClick={handleRetry} className="me-2">
              <FaArrowsRotate className="me-2" />
              重试
            </Button>
          </>
        ) : (
          <>
            <FaCircleInfo className="fas fa-exclamation-triangle fa-3x text-warning mb-3" />
            <h5>案件不存在</h5>
            <p className="text-muted">请求的案件无法找到</p>
          </>
        )}
        <Button variant="primary" onClick={() => navigate("/cases")}>
          <FaArrowLeft className="me-2" />
          返回案件列表
        </Button>
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center py-5">
        <FaCircleInfo className="fas fa-exclamation-circle fa-3x text-danger mb-3" />
        <h5>加载失败</h5>
        <p className="text-muted">{error}</p>
        <div className="mt-3">
          <Button variant="primary" onClick={handleRetry} className="me-2">
            <FaArrowsRotate className="me-2" />
            重试
          </Button>
          <Button variant="outline-secondary" onClick={() => navigate("/cases")}>
            <FaArrowLeft className="me-2" />
            返回案件列表
          </Button>
        </div>
      </div>
    );
  }

  // 在这一点上，caseData保证不为null，因为前面的条件检查已经处理了null情况
  if (!caseData) return null;

  return (
    <div>
      {/* 头部操作栏 */}
      <div className="d-flex justify-content-between align-items-center mb-4">
        <div>
          <Button variant="outline-secondary" onClick={() => navigate("/cases")} className="mb-2">
            <FaArrowLeft className="me-2" />
            返回案件列表
          </Button>
          <h1>案件 #{caseData.id}: {caseData.title}</h1>
        </div>
        <div className="d-flex">
          <Dropdown className="me-2">
            <Dropdown.Toggle variant="outline-secondary" id="actions-dropdown">
              操作
            </Dropdown.Toggle>
            <Dropdown.Menu>
              <Dropdown.Item onClick={handleEdit}>
                <FaPenToSquare className="me-2" />
                编辑案件
              </Dropdown.Item>
              <Dropdown.Item>
                <FaFileLines className="me-2" />
                导出报告
              </Dropdown.Item>
              <Dropdown.Item>
                <FaArrowsRotate className="me-2" />
                查看历史
              </Dropdown.Item>
              <Dropdown.Divider />
              <Dropdown.Item className="text-danger" onClick={handleDelete}>
                <FaTrash className="me-2" />
                删除案件
              </Dropdown.Item>
            </Dropdown.Menu>
          </Dropdown>
          <Button variant="primary" onClick={handleEdit}>
            <FaPenToSquare className="me-2" />
            编辑案件
          </Button>
        </div>
      </div>

      {/* 状态标签 */}
      <div className="mb-4">
        {getStatusBadge(caseData.status)}
        {getCaseTypeBadge(caseData.case_type)}
      </div>

      {/* 主要内容区域 */}
      <Row>
        <Col md={8}>
          <Tabs activeKey={activeTab} onSelect={(k) => setActiveTab(k || "basic")} className="mb-4">
            {/* 基本信息 */}
            <Tab eventKey="basic" title="基本信息">
              <Card>
                <Card.Body>
                  <Row>
                    <Col md={6}>
                      <div className="mb-3">
                        <small className="text-muted">案件编号</small>
                        <div className="fw-bold">#{caseData.id}</div>
                      </div>
                      <div className="mb-3">
                        <small className="text-muted">案件名称</small>
                        <div className="fw-bold">{caseData.title}</div>
                      </div>
                      <div className="mb-3">
                        <small className="text-muted">案件类型</small>
                        <div>{getCaseTypeBadge(caseData.case_type)}</div>
                      </div>
                      <div className="mb-3">
                        <small className="text-muted">状态</small>
                        <div>{getStatusBadge(caseData.status)}</div>
                      </div>
                      <div className="mb-3">
                        <small className="text-muted">优先级</small>
                        <div>{getPriorityBadge(caseData.priority)}</div>
                      </div>
                    </Col>
                    <Col md={6}>
                      <div className="mb-3">
                        <small className="text-muted">创建时间</small>
                        <div className="fw-bold">
                          <FaCalendar className="me-1" />
                          {formatDate(caseData.created_at)}
                        </div>
                      </div>
                      <div className="mb-3">
                        <small className="text-muted">更新时间</small>
                        <div className="fw-bold">
                          <FaClock className="me-1" />
                          {formatDate(caseData.updated_at)}
                        </div>
                      </div>
                      {caseData.lawyer_id && (
                        <div className="mb-3">
                          <small className="text-muted">律师ID</small>
                          <div className="fw-bold">{caseData.lawyer_id}</div>
                        </div>
                      )}
                      {caseData.client_id && (
                        <div className="mb-3">
                          <small className="text-muted">客户ID</small>
                          <div className="fw-bold">{caseData.client_id}</div>
                        </div>
                      )}
                    </Col>
                  </Row>

                  <Row>
                    <Col md={12}>
                      <div className="mb-3">
                        <small className="text-muted">案件描述</small>
                        <div className="p-3 bg-light rounded">
                          {caseData.description || "暂无描述"}
                        </div>
                      </div>
                    </Col>
                  </Row>
                </Card.Body>
              </Card>
            </Tab>

            {/* 客户信息 */}
            <Tab eventKey="client" title="客户信息">
              <Card>
                <Card.Header>
                  <FaUser className="me-2" />
                  客户信息
                </Card.Header>
                <Card.Body>
                  <Row>
                    <Col md={6}>
                      <div className="mb-3">
                        <small className="text-muted">客户姓名</small>
                        <div className="fw-bold">{caseData.clientName}</div>
                      </div>
                      <div className="mb-3">
                        <small className="text-muted">联系电话</small>
                        <div className="fw-bold">
                          <FaPhone className="me-1" />
                          {caseData.clientPhone}
                        </div>
                      </div>
                    </Col>
                    <Col md={6}>
                      <div className="mb-3">
                        <small className="text-muted">邮箱地址</small>
                        <div className="fw-bold">
                          <FaEnvelope className="me-1" />
                          {caseData.clientEmail}
                        </div>
                      </div>
                      <div className="mb-3">
                        <small className="text-muted">联系地址</small>
                        <div className="fw-bold">
                          <FaLocationPin className="me-1" />
                          {caseData.clientAddress}
                        </div>
                      </div>
                    </Col>
                  </Row>

                  <div className="d-grid gap-2">
                    <Button variant="primary">
                      <FaEnvelope className="me-2" />
                      发送邮件
                    </Button>
                    <Button variant="outline-primary">
                      <FaPhone className="me-2" />
                      拨打电话
                    </Button>
                  </div>
                </Card.Body>
              </Card>
            </Tab>

            {/* 当事人信息 */}
            <Tab eventKey="parties" title="当事人信息">
              <Card>
                <Card.Header>
                  <FaPeopleGroup className="me-2" />
                  当事人信息
                </Card.Header>
                <Card.Body>
                  <Row>
                    <Col md={6}>
                      <h6 className="text-primary mb-3">委托人信息</h6>
                      <div className="p-3 bg-light rounded">
                        <pre style={{ whiteSpace: 'pre-wrap', fontFamily: 'inherit', margin: 0 }}>
                          {caseData.principalInfo || '暂无委托人信息'}
                        </pre>
                      </div>
                    </Col>
                    <Col md={6}>
                      <h6 className="text-danger mb-3">对方当事人信息</h6>
                      <div className="p-3 bg-light rounded">
                        <pre style={{ whiteSpace: 'pre-wrap', fontFamily: 'inherit', margin: 0 }}>
                          {caseData.opponentInfo || '暂无对方当事人信息'}
                        </pre>
                      </div>
                    </Col>
                  </Row>
                </Card.Body>
              </Card>
            </Tab>

            {/* 律师信息 */}
            <Tab eventKey="lawyer" title="律师信息">
              <Card>
                <Card.Header>
                  <FaGavel className="me-2" />
                  律师信息
                </Card.Header>
                <Card.Body>
                  <Row>
                    <Col md={6}>
                      <div className="mb-3">
                        <small className="text-muted">负责律师</small>
                        <div className="fw-bold">{caseData.lawyerName}</div>
                      </div>
                      <div className="mb-3">
                        <small className="text-muted">联系电话</small>
                        <div className="fw-bold">
                          <FaPhone className="me-1" />
                          {caseData.lawyerPhone}
                        </div>
                      </div>
                    </Col>
                    <Col md={6}>
                      <div className="mb-3">
                        <small className="text-muted">邮箱地址</small>
                        <div className="fw-bold">
                          <FaEnvelope className="me-1" />
                          {caseData.lawyerEmail}
                        </div>
                      </div>
                    </Col>
                  </Row>

                  <div className="d-grid gap-2">
                    <Button variant="primary">
                      <FaEnvelope className="me-2" />
                      联系律师
                    </Button>
                  </div>
                </Card.Body>
              </Card>
            </Tab>

            {/* 案件文档 */}
            <Tab eventKey="documents" title="案件文档">
              <Card>
                <Card.Header className="d-flex justify-content-between align-items-center">
                  <span>
                    <FaFileLines className="me-2" />
                    案件文档
                  </span>
                  <Button variant="primary" size="sm">
                    <FaPlus className="me-1" />
                    上传文档
                  </Button>
                </Card.Header>
                <Card.Body>
                  {documents.length === 0 ? (
                    <Alert variant="info">
                      暂无文档
                    </Alert>
                  ) : (
                    <ListGroup>
                      {documents.map(doc => (
                        <ListGroup.Item key={doc.id} className="d-flex justify-content-between align-items-center">
                          <div className="d-flex align-items-center">
                            <FaFile className="me-3 text-primary" />
                            <div>
                              <div className="fw-bold">{doc.name}</div>
                              <small className="text-muted">
                                {doc.type} • {doc.size} • {doc.uploadTime}
                              </small>
                            </div>
                          </div>
                          <div>
                            <Button variant="outline-primary" size="sm" className="me-2">
                              <FaDownload />
                            </Button>
                            <Button variant="outline-danger" size="sm">
                              <FaTrash />
                            </Button>
                          </div>
                        </ListGroup.Item>
                      ))}
                    </ListGroup>
                  )}
                </Card.Body>
              </Card>
            </Tab>

            {/* 案件进展 */}
            <Tab eventKey="timeline" title="案件进展">
              <Card>
                <Card.Header>
                  <FaArrowsRotate className="me-2" />
                  案件进展
                </Card.Header>
                <Card.Body>
                  {timeline.length === 0 ? (
                    <Alert variant="info">
                      暂无进展记录
                    </Alert>
                  ) : (
                    <div className="timeline">
                      {timeline.map((item, index) => (
                        <div key={item.id} className="timeline-item mb-4">
                          <div className="d-flex">
                            <div className="timeline-icon bg-primary rounded-circle d-flex align-items-center justify-content-center me-3">
                              <FaClock className="text-white" />
                            </div>
                            <div className="timeline-content">
                              <h6 className="mb-1">{item.event}</h6>
                              <p className="mb-1 text-muted">{item.description}</p>
                              <small className="text-muted">
                                {item.time} • {item.user}
                              </small>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </Card.Body>
              </Card>
            </Tab>
          </Tabs>
        </Col>

        {/* 右侧信息栏 */}
        <Col md={4}>
          {/* 快速统计 */}
          <Card className="mb-4">
            <Card.Header>快速统计</Card.Header>
            <Card.Body>
              <Row className="text-center">
                <Col md={4}>
                  <div className="mb-3">
                    <div className="h3 text-primary">{documents.length}</div>
                    <small className="text-muted">文档</small>
                  </div>
                </Col>
                <Col md={4}>
                  <div className="mb-3">
                    <div className="h3 text-success">{timeline.length}</div>
                    <small className="text-muted">进展</small>
                  </div>
                </Col>
                <Col md={4}>
                  <div className="mb-3">
                    <div className="h3 text-warning">{expenses.length}</div>
                    <small className="text-muted">费用</small>
                  </div>
                </Col>
              </Row>
            </Card.Body>
          </Card>

          {/* 最近活动 */}
          <Card className="mb-4">
            <Card.Header>最近活动</Card.Header>
            <Card.Body>
              {timeline.slice(0, 3).map(item => (
                <div key={item.id} className="mb-3 pb-3 border-bottom">
                  <div className="fw-bold">{item.event}</div>
                  <small className="text-muted">{item.time}</small>
                </div>
              ))}
            </Card.Body>
          </Card>

          {/* 快速操作 */}
          <Card>
            <Card.Header>快速操作</Card.Header>
            <Card.Body>
              <div className="d-grid gap-2">
                <Button variant="outline-primary">
                  <FaPlus className="me-2" />
                  添加进展
                </Button>
                <Button variant="outline-primary">
                  <FaComment className="me-2" />
                  添加备注
                </Button>
                <Button variant="outline-primary">
                  <FaFileLines className="me-2" />
                  生成报告
                </Button>
              </div>
            </Card.Body>
          </Card>
        </Col>
      </Row>

      {/* 编辑模态框 */}
      <Modal show={showEditModal} onHide={() => setShowEditModal(false)} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>编辑案件</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <Form>
            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>案件名称</Form.Label>
                  <Form.Control
                    type="text"
                    name="title"
                    value={editForm.title || ''}
                    onChange={handleInputChange}
                  />
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>案件类型</Form.Label>
                  <Form.Select
                    name="case_type"
                    value={editForm.case_type || ''}
                    onChange={handleInputChange}
                  >
                    <option value="">请选择案件类型</option>
                    <option value="civil">民事案件</option>
                    <option value="commercial">商事案件</option>
                    <option value="criminal">刑事案件</option>
                    <option value="administrative">行政案件</option>
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
                    value={editForm.status || ''}
                    onChange={handleInputChange}
                  >
                    <option value="">请选择状态</option>
                    <option value="pending">待处理</option>
                    <option value="active">进行中</option>
                    <option value="closed">已结案</option>
                    <option value="suspended">已暂停</option>
                  </Form.Select>
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>优先级</Form.Label>
                  <Form.Select
                    name="priority"
                    value={editForm.priority || ''}
                    onChange={handleInputChange}
                  >
                    <option value="">请选择优先级</option>
                    <option value="low">低</option>
                    <option value="medium">中</option>
                    <option value="high">高</option>
                    <option value="urgent">紧急</option>
                  </Form.Select>
                </Form.Group>
              </Col>
            </Row>

            <Form.Group className="mb-3">
              <Form.Label>案件描述</Form.Label>
              <Form.Control
                as="textarea"
                rows={4}
                name="description"
                value={editForm.description || ''}
                onChange={handleInputChange}
              />
            </Form.Group>
          </Form>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => setShowEditModal(false)}>
            <FaXmark className="me-2" />
            取消
          </Button>
          <Button variant="primary" onClick={handleSave}>
            <FaDownload className="me-2" />
            保存
          </Button>
        </Modal.Footer>
      </Modal>
    </div>
  );
};

export default CaseDetailPage;