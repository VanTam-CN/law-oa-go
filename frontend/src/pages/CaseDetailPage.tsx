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

// 案件详情接口
interface CaseDetail {
  caseId: number;
  caseNo: string;
  caseName: string;
  caseType: string;
  clientName: string;
  clientPhone: string;
  clientEmail: string;
  clientAddress: string;
  lawyerName: string;
  lawyerPhone: string;
  lawyerEmail: string;
  status: string;
  description: string;
  createTime: string;
  updateTime: string;
  expectedAmount: number;
  actualAmount: number;
  principalInfo?: string;
  opponentInfo?: string;
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
  const [caseData, setCaseData] = useState<CaseDetail | null>(null);
  const [documents, setDocuments] = useState<CaseDocument[]>([]);
  const [timeline, setTimeline] = useState<CaseTimeline[]>([]);
  const [expenses, setExpenses] = useState<CaseExpense[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState("basic");
  const [showEditModal, setShowEditModal] = useState(false);
  const [editForm, setEditForm] = useState<Partial<CaseDetail>>({});

  // 模拟API调用 - 获取案件详情
  useEffect(() => {
    if (id) {
      fetchCaseDetail(parseInt(id));
    }
  }, [id]);

  const fetchCaseDetail = async (caseId: number) => {
    setLoading(true);
    try {
      // 模拟API调用 - 实际项目中应该调用真实的API
      await new Promise(resolve => setTimeout(resolve, 1000));

      // 模拟数据
      const mockCaseData: CaseDetail = {
        caseId: caseId,
        caseNo: `CASE-${String(caseId).padStart(4, '0')}`,
        caseName: "合同纠纷案",
        caseType: "CIVIL",
        clientName: "张三",
        clientPhone: "13800138000",
        clientEmail: "zhangsan@example.com",
        clientAddress: "北京市朝阳区某某街道123号",
        lawyerName: "李律师",
        lawyerPhone: "13900139000",
        lawyerEmail: "li.lawyer@example.com",
        status: "1",
        description: "这是一起关于合同纠纷的案件，涉及金额较大，需要仔细处理。",
        createTime: "2024-01-15 10:30:00",
        updateTime: "2024-01-20 15:45:00",
        expectedAmount: 500000,
        actualAmount: 450000,
        principalInfo: "委托人：张三\n身份证号：110101199001011234\n联系方式：13800138000\n住址：北京市朝阳区某某街道123号",
        opponentInfo: "对方当事人：某公司\n统一社会信用代码：91110108XXXXXXXXXX\n法定代表人：王五\n联系方式：010-12345678\n地址：北京市海淀区某某大厦10层"
      };

      const mockDocuments: CaseDocument[] = [
        {
          id: 1,
          name: "合同原件.pdf",
          type: "PDF",
          size: "2.5MB",
          uploadTime: "2024-01-15 10:30:00",
          uploader: "李律师"
        },
        {
          id: 2,
          name: "证据清单.docx",
          type: "Word",
          size: "1.2MB",
          uploadTime: "2024-01-16 14:20:00",
          uploader: "张三"
        }
      ];

      const mockTimeline: CaseTimeline[] = [
        {
          id: 1,
          time: "2024-01-15 10:30:00",
          event: "案件创建",
          user: "系统",
          description: "案件已在系统中创建"
        },
        {
          id: 2,
          time: "2024-01-16 09:00:00",
          event: "律师指派",
          user: "管理员",
          description: "李律师被指派处理此案件"
        },
        {
          id: 3,
          time: "2024-01-17 14:30:00",
          event: "客户会面",
          user: "李律师",
          description: "与客户进行首次会面，了解案件详情"
        }
      ];

      const mockExpenses: CaseExpense[] = [
        {
          id: 1,
          date: "2024-01-16",
          type: "诉讼费",
          amount: 10000,
          description: "法院诉讼费",
          status: "已支付"
        },
        {
          id: 2,
          date: "2024-01-17",
          type: "律师费",
          amount: 50000,
          description: "律师代理费",
          status: "未支付"
        }
      ];

      setCaseData(mockCaseData);
      setDocuments(mockDocuments);
      setTimeline(mockTimeline);
      setExpenses(mockExpenses);
      setEditForm(mockCaseData);
    } catch (error) {
      console.error("获取案件详情失败:", error);
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
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 1000));

      setCaseData({ ...caseData, ...editForm, updateTime: new Date().toLocaleString() });
      setShowEditModal(false);

      // 显示成功消息
      alert("案件信息已更新");
    } catch (error) {
      console.error("更新案件失败:", error);
      alert("更新失败，请重试");
    }
  };

  const handleDelete = async () => {
    if (!caseData) return;

    if (window.confirm("确定要删除这个案件吗？此操作不可撤销。")) {
      try {
        // 模拟API调用
        await new Promise(resolve => setTimeout(resolve, 1000));

        alert("删除成功");
        navigate("/cases");
      } catch (error) {
        console.error("删除案件失败:", error);
        alert("删除失败，请重试");
      }
    }
  };

  const getStatusBadge = (status: string) => {
    const statusMap = {
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
      'CIVIL': { text: '民事案件', variant: 'primary' },
      'COMMERCIAL': { text: '商事案件', variant: 'warning' },
      'CRIMINAL': { text: '刑事案件', variant: 'danger' },
      'ADMINISTRATIVE': { text: '行政案件', variant: 'info' }
    };
    const config = typeMap[type as keyof typeof typeMap] || { text: '其他', variant: 'secondary' };
    return <Badge bg={config.variant}>{config.text}</Badge>;
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setEditForm(prev => ({ ...prev, [name]: value }));
  };

  if (loading) {
    return (
      <div className="d-flex justify-content-center align-items-center" style={{ height: "50vh" }}>
        <Spinner animation="border" />
        <span className="ms-2">加载中...</span>
      </div>
    );
  }

  if (!caseData) {
    return (
      <div className="text-center py-5">
        <FaCircleInfo className="fas fa-exclamation-triangle fa-3x text-warning mb-3" />
        <h5>案件不存在</h5>
        <p className="text-muted">请求的案件无法找到</p>
        <Button variant="primary" onClick={() => navigate("/cases")}>
          <FaArrowLeft className="me-2" />
          返回案件列表
        </Button>
      </div>
    );
  }

  return (
    <div>
      {/* 头部操作栏 */}
      <div className="d-flex justify-content-between align-items-center mb-4">
        <div>
          <Button variant="outline-secondary" onClick={() => navigate("/cases")} className="mb-2">
            <FaArrowLeft className="me-2" />
            返回案件列表
          </Button>
          <h1>案件 #{caseData.caseNo}: {caseData.caseName}</h1>
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
        {getCaseTypeBadge(caseData.caseType)}
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
                        <div className="fw-bold">{caseData.caseNo}</div>
                      </div>
                      <div className="mb-3">
                        <small className="text-muted">案件名称</small>
                        <div className="fw-bold">{caseData.caseName}</div>
                      </div>
                      <div className="mb-3">
                        <small className="text-muted">案件类型</small>
                        <div>{getCaseTypeBadge(caseData.caseType)}</div>
                      </div>
                      <div className="mb-3">
                        <small className="text-muted">状态</small>
                        <div>{getStatusBadge(caseData.status)}</div>
                      </div>
                    </Col>
                    <Col md={6}>
                      <div className="mb-3">
                        <small className="text-muted">预计金额</small>
                        <div className="fw-bold">
                          <FaDollarSign className="me-1" />
                          ¥{caseData.expectedAmount.toLocaleString()}
                        </div>
                      </div>
                      <div className="mb-3">
                        <small className="text-muted">实际金额</small>
                        <div className="fw-bold">
                          <FaDollarSign className="me-1" />
                          ¥{caseData.actualAmount.toLocaleString()}
                        </div>
                      </div>
                      <div className="mb-3">
                        <small className="text-muted">创建时间</small>
                        <div className="fw-bold">
                          <FaCalendar className="me-1" />
                          {caseData.createTime}
                        </div>
                      </div>
                      <div className="mb-3">
                        <small className="text-muted">更新时间</small>
                        <div className="fw-bold">
                          <FaClock className="me-1" />
                          {caseData.updateTime}
                        </div>
                      </div>
                    </Col>
                  </Row>

                  <Row>
                    <Col md={12}>
                      <div className="mb-3">
                        <small className="text-muted">案件描述</small>
                        <div className="p-3 bg-light rounded">
                          {caseData.description}
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
                    name="caseName"
                    value={editForm.caseName || ''}
                    onChange={handleInputChange}
                  />
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>案件类型</Form.Label>
                  <Form.Select
                    name="caseType"
                    value={editForm.caseType || ''}
                    onChange={handleInputChange}
                  >
                    <option value="CIVIL">民事案件</option>
                    <option value="COMMERCIAL">商事案件</option>
                    <option value="CRIMINAL">刑事案件</option>
                    <option value="ADMINISTRATIVE">行政案件</option>
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
                    <option value="0">未开始</option>
                    <option value="1">进行中</option>
                    <option value="2">已结案</option>
                    <option value="3">已归档</option>
                  </Form.Select>
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>预计金额</Form.Label>
                  <Form.Control
                    type="number"
                    name="expectedAmount"
                    value={editForm.expectedAmount || ''}
                    onChange={handleInputChange}
                  />
                </Form.Group>
              </Col>
            </Row>

            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>客户姓名</Form.Label>
                  <Form.Control
                    type="text"
                    name="clientName"
                    value={editForm.clientName || ''}
                    onChange={handleInputChange}
                  />
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>律师姓名</Form.Label>
                  <Form.Control
                    type="text"
                    name="lawyerName"
                    value={editForm.lawyerName || ''}
                    onChange={handleInputChange}
                  />
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