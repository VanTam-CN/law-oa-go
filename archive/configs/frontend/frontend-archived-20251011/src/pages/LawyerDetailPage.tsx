import React, { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import {
  Card,
  Button,
  Spinner,
  Form,
  Badge,
  Modal,
  Alert,
  Row,
  Col,
  ListGroup
} from "react-bootstrap";
import {
  FaEdit,
  FaArrowLeft,
  FaUser,
  FaPhone,
  FaEnvelope,
  FaBriefcase,
  FaCheckCircle,
  FaUserTimes,
  FaPauseCircle,
  FaUserTie,
  FaCalendar,
  FaMapMarkerAlt,
  FaGraduationCap,
  FaAward,
  FaClock
} from "react-icons/fa";
import { toast } from "../components/Toast";

// 律师详情接口
interface LawyerDetail {
  id?: number;
  name?: string;
  username?: string;
  phone?: string;
  email?: string;
  licenseNumber?: string;
  department?: string;
  position?: string;
  status?: string;
  specialty?: string[];
  experience?: number;
  gender?: string;
  joinDate?: string;
  profile?: string;
  avatar?: string;
  address?: string;
  education?: string;
  achievements?: string[];
  hourlyRate?: number;
  consultationHours?: string;
  caseCount?: number;
  successRate?: number;
  createdAt?: string;
  updatedAt?: string;
}

const LawyerDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [lawyer, setLawyer] = useState<LawyerDetail | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [showEditModal, setShowEditModal] = useState<boolean>(false);
  const [editForm, setEditForm] = useState<LawyerDetail>({});
  const [saving, setSaving] = useState<boolean>(false);

  // 获取律师详情
  useEffect(() => {
    if (id) {
      fetchLawyerDetail(parseInt(id));
    }
  }, [id]);

  const fetchLawyerDetail = async (lawyerId: number) => {
    setLoading(true);
    setError(null);

    try {
      // 模拟API调用 - 在实际项目中应该调用真实的API
      console.log(`🛠️ 开发模式：加载律师详情 (ID: ${lawyerId})`);
      await new Promise(resolve => setTimeout(resolve, 1000));

      // 模拟数据
      const mockLawyer: LawyerDetail = {
        id: lawyerId,
        name: '张律师',
        username: 'zhang_lawyer',
        phone: '13800138001',
        email: 'zhang.lawyer@example.com',
        licenseNumber: '123456789012345',
        department: '民事诉讼部',
        position: '合伙人',
        status: 'active',
        specialty: ['合同纠纷', '侵权责任', '房产纠纷'],
        experience: 15,
        gender: 'male',
        joinDate: '2010-01-15',
        profile: '资深律师，专注于民事诉讼领域，具有丰富的庭审经验和调解技巧。擅长处理复杂的合同纠纷案件，为客户争取最大利益。',
        avatar: '',
        address: '北京市朝阳区建国门外大街1号',
        education: '中国政法大学 法学硕士',
        achievements: [
          '2023年度优秀律师',
          '处理案件成功率95%',
          '15年执业经验'
        ],
        hourlyRate: 800,
        consultationHours: '周一至周五 9:00-18:00',
        caseCount: 156,
        successRate: 95,
        createdAt: '2024-01-15T10:30:00Z',
        updatedAt: '2025-10-09T14:20:00Z'
      };

      setLawyer(mockLawyer);
      setEditForm(mockLawyer);
      console.log('✅ 律师详情加载完成');
    } catch (error) {
      console.error('获取律师详情失败:', error);
      setError(`加载律师详情失败: ${error instanceof Error ? error.message : '未知错误'}`);
    } finally {
      setLoading(false);
    }
  };

  // 获取状态标签
  const getStatusBadge = (status?: string) => {
    const statusMap = {
      'active': { text: '在职', variant: 'success', icon: <FaCheckCircle /> },
      'inactive': { text: '离职', variant: 'danger', icon: <FaUserTimes /> },
      'on_leave': { text: '休假', variant: 'warning', icon: <FaPauseCircle /> },
      // 兼容可能的数字状态
      '1': { text: '在职', variant: 'success', icon: <FaCheckCircle /> },
      '0': { text: '离职', variant: 'danger', icon: <FaUserTimes /> },
      '2': { text: '休假', variant: 'warning', icon: <FaPauseCircle /> }
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
  const getPositionBadge = (position?: string) => {
    const positionMap = {
      '合伙人': { variant: 'danger' },
      '资深律师': { variant: 'primary' },
      '律师': { variant: 'info' },
      '实习律师': { variant: 'secondary' }
    };
    const variant = positionMap[position as keyof typeof positionMap] || 'secondary';
    return <Badge bg={variant}>{position}</Badge>;
  };

  // 打开编辑模态框
  const handleEdit = () => {
    setShowEditModal(true);
    setEditForm(lawyer || {});
  };

  // 处理输入变化
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setEditForm(prev => ({
      ...prev,
      [name]: value
    }));
  };

  // 保存编辑
  const handleSave = async () => {
    if (!lawyer?.id) return;

    setSaving(true);
    try {
      // 构建更新数据，只包含有变化的字段
      const updateData: any = {};

      if (editForm.name !== undefined && editForm.name !== lawyer.name) {
        updateData.name = editForm.name;
      }
      if (editForm.phone !== undefined && editForm.phone !== lawyer.phone) {
        updateData.phone = editForm.phone;
      }
      if (editForm.email !== undefined && editForm.email !== lawyer.email) {
        updateData.email = editForm.email;
      }
      if (editForm.department !== undefined && editForm.department !== lawyer.department) {
        updateData.department = editForm.department;
      }
      if (editForm.position !== undefined && editForm.position !== lawyer.position) {
        updateData.position = editForm.position;
      }
      if (editForm.status !== undefined && editForm.status !== lawyer.status) {
        updateData.status = editForm.status;
      }
      if (editForm.profile !== undefined && editForm.profile !== lawyer.profile) {
        updateData.profile = editForm.profile;
      }
      if (editForm.address !== undefined && editForm.address !== lawyer.address) {
        updateData.address = editForm.address;
      }
      if (editForm.hourlyRate !== undefined && editForm.hourlyRate !== lawyer.hourlyRate) {
        updateData.hourlyRate = editForm.hourlyRate;
      }

      // 如果没有变化，直接关闭模态框
      if (Object.keys(updateData).length === 0) {
        setShowEditModal(false);
        toast.showToast({
          type: 'info',
          title: '无变化',
          message: '律师信息没有发生变化',
          duration: 2000
        });
        return;
      }

      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 1000));

      // 更新本地数据
      const updatedLawyer = { ...lawyer, ...updateData, updated_at: new Date().toISOString() };
      setLawyer(updatedLawyer);
      setEditForm(updatedLawyer);
      setShowEditModal(false);

      // 显示成功消息
      toast.showToast({
        type: 'success',
        title: '操作成功',
        message: '律师信息已更新',
        duration: 3000
      });
    } catch (error) {
      console.error('保存失败:', error);
      toast.showToast({
        type: 'error',
        title: '操作失败',
        message: '更新律师信息失败，请重试',
        duration: 3000
      });
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="text-center py-5">
        <Spinner animation="border" role="status" style={{ width: '3rem', height: '3rem' }}>
          <span className="visually-hidden">加载中...</span>
        </Spinner>
        <div className="mt-3">
          <h5 className="text-muted">正在加载律师详情...</h5>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="py-5">
        <Alert variant="danger">
          <Alert.Heading>加载失败</Alert.Heading>
          <p>{error}</p>
          <hr />
          <div className="d-flex justify-content-end">
            <Button variant="outline-danger" onClick={() => navigate(-1)}>
              <FaArrowLeft className="me-2" />
              返回
            </Button>
          </div>
        </Alert>
      </div>
    );
  }

  if (!lawyer) {
    return (
      <div className="text-center py-5">
        <h5 className="text-muted">未找到律师信息</h5>
        <Button variant="primary" onClick={() => navigate(-1)} className="mt-3">
          <FaArrowLeft className="me-2" />
          返回
        </Button>
      </div>
    );
  }

  return (
    <div>
      {/* 头部 */}
      <div className="d-flex justify-content-between align-items-center mb-4">
        <div className="d-flex align-items-center">
          <Button variant="outline-secondary" onClick={() => navigate(-1)} className="me-3">
            <FaArrowLeft className="me-2" />
            返回
          </Button>
          <div>
            <h1 className="mb-1">律师详情</h1>
            <p className="text-muted mb-0">查看和编辑律师详细信息</p>
          </div>
        </div>
        <Button variant="primary" onClick={handleEdit}>
          <FaEdit className="me-2" />
          编辑
        </Button>
      </div>

      {/* 律师基本信息 */}
      <Row className="mb-4">
        <Col md={4}>
          <Card className="text-center">
            <Card.Body>
              <div className="bg-light rounded-circle d-flex align-items-center justify-content-center mx-auto mb-3" style={{ width: '120px', height: '120px' }}>
                <FaUser className="text-muted" style={{ fontSize: '3rem' }} />
              </div>
              <h4>{lawyer.name}</h4>
              <p className="text-muted mb-2">{lawyer.position}</p>
              {getPositionBadge(lawyer.position)}
              <div className="mt-3">
                {getStatusBadge(lawyer.status)}
              </div>
            </Card.Body>
          </Card>
        </Col>
        <Col md={8}>
          <Card>
            <Card.Header>
              <h5 className="mb-0">基本信息</h5>
            </Card.Header>
            <Card.Body>
              <Row>
                <Col md={6}>
                  <div className="mb-3">
                    <small className="text-muted d-block">用户名</small>
                    <strong>{lawyer.username || '-'}</strong>
                  </div>
                  <div className="mb-3">
                    <small className="text-muted d-block">执业证号</small>
                    <strong>{lawyer.licenseNumber || '-'}</strong>
                  </div>
                  <div className="mb-3">
                    <small className="text-muted d-block">所属部门</small>
                    <strong>{lawyer.department || '-'}</strong>
                  </div>
                  <div className="mb-3">
                    <small className="text-muted d-block">工作经验</small>
                    <strong>{lawyer.experience || 0} 年</strong>
                  </div>
                </Col>
                <Col md={6}>
                  <div className="mb-3">
                    <small className="text-muted d-block">联系电话</small>
                    <div>
                      <FaPhone className="me-1 text-muted" />
                      {lawyer.phone || '-'}
                    </div>
                  </div>
                  <div className="mb-3">
                    <small className="text-muted d-block">邮箱地址</small>
                    <div>
                      <FaEnvelope className="me-1 text-muted" />
                      {lawyer.email || '-'}
                    </div>
                  </div>
                  <div className="mb-3">
                    <small className="text-muted d-block">办公地址</small>
                    <div>
                      <FaMapMarkerAlt className="me-1 text-muted" />
                      {lawyer.address || '-'}
                    </div>
                  </div>
                  <div className="mb-3">
                    <small className="text-muted d-block">加入时间</small>
                    <div>
                      <FaCalendar className="me-1 text-muted" />
                      {lawyer.joinDate || '-'}
                    </div>
                  </div>
                </Col>
              </Row>
            </Card.Body>
          </Card>
        </Col>
      </Row>

      {/* 详细信息 */}
      <Row>
        <Col md={6}>
          <Card className="mb-4">
            <Card.Header>
              <h5 className="mb-0">
                <FaGraduationCap className="me-2" />
                教育背景
              </h5>
            </Card.Header>
            <Card.Body>
              <p>{lawyer.education || '-'}</p>
            </Card.Body>
          </Card>
        </Col>
        <Col md={6}>
          <Card className="mb-4">
            <Card.Header>
              <h5 className="mb-0">
                <FaBriefcase className="me-2" />
                专业领域
              </h5>
            </Card.Header>
            <Card.Body>
              {lawyer.specialty && lawyer.specialty.length > 0 ? (
                <div className="d-flex flex-wrap gap-2">
                  {lawyer.specialty.map((specialty, index) => (
                    <Badge key={index} bg="info">{specialty}</Badge>
                  ))}
                </div>
              ) : (
                <p className="text-muted">-</p>
              )}
            </Card.Body>
          </Card>
        </Col>
      </Row>

      <Row>
        <Col md={12}>
          <Card className="mb-4">
            <Card.Header>
              <h5 className="mb-0">个人简介</h5>
            </Card.Header>
            <Card.Body>
              <p>{lawyer.profile || '-'}</p>
            </Card.Body>
          </Card>
        </Col>
      </Row>

      <Row>
        <Col md={6}>
          <Card className="mb-4">
            <Card.Header>
              <h5 className="mb-0">
                <FaAward className="me-2" />
                成就荣誉
              </h5>
            </Card.Header>
            <Card.Body>
              {lawyer.achievements && lawyer.achievements.length > 0 ? (
                <ListGroup variant="flush">
                  {lawyer.achievements.map((achievement, index) => (
                    <ListGroup.Item key={index}>
                      <FaCheckCircle className="me-2 text-success" />
                      {achievement}
                    </ListGroup.Item>
                  ))}
                </ListGroup>
              ) : (
                <p className="text-muted">-</p>
              )}
            </Card.Body>
          </Card>
        </Col>
        <Col md={6}>
          <Card className="mb-4">
            <Card.Header>
              <h5 className="mb-0">
                <FaClock className="me-2" />
                业务统计
              </h5>
            </Card.Header>
            <Card.Body>
              <Row>
                <Col md={6}>
                  <div className="text-center mb-3">
                    <h3 className="text-primary">{lawyer.caseCount || 0}</h3>
                    <small className="text-muted">处理案件</small>
                  </div>
                </Col>
                <Col md={6}>
                  <div className="text-center mb-3">
                    <h3 className="text-success">{lawyer.successRate || 0}%</h3>
                    <small className="text-muted">成功率</small>
                  </div>
                </Col>
              </Row>
              <hr />
              <div className="mb-2">
                <small className="text-muted">咨询费用</small>
                <strong className="d-block">¥{lawyer.hourlyRate || 0}/小时</strong>
              </div>
              <div>
                <small className="text-muted">咨询时间</small>
                <strong className="d-block">{lawyer.consultationHours || '-'}</strong>
              </div>
            </Card.Body>
          </Card>
        </Col>
      </Row>

      {/* 编辑模态框 */}
      <Modal show={showEditModal} onHide={() => setShowEditModal(false)} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>编辑律师信息</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <Row>
            <Col md={6}>
              <Form.Group className="mb-3">
                <Form.Label>姓名 *</Form.Label>
                <Form.Control
                  type="text"
                  name="name"
                  value={editForm.name || ''}
                  onChange={handleInputChange}
                  required
                />
              </Form.Group>
            </Col>
            <Col md={6}>
              <Form.Group className="mb-3">
                <Form.Label>联系电话 *</Form.Label>
                <Form.Control
                  type="tel"
                  name="phone"
                  value={editForm.phone || ''}
                  onChange={handleInputChange}
                  required
                />
              </Form.Group>
            </Col>
          </Row>

          <Row>
            <Col md={6}>
              <Form.Group className="mb-3">
                <Form.Label>邮箱地址 *</Form.Label>
                <Form.Control
                  type="email"
                  name="email"
                  value={editForm.email || ''}
                  onChange={handleInputChange}
                  required
                />
              </Form.Group>
            </Col>
            <Col md={6}>
              <Form.Group className="mb-3">
                <Form.Label>执业证号</Form.Label>
                <Form.Control
                  type="text"
                  name="licenseNumber"
                  value={editForm.licenseNumber || ''}
                  onChange={handleInputChange}
                />
              </Form.Group>
            </Col>
          </Row>

          <Row>
            <Col md={6}>
              <Form.Group className="mb-3">
                <Form.Label>所属部门</Form.Label>
                <Form.Select
                  name="department"
                  value={editForm.department || ''}
                  onChange={handleInputChange}
                >
                  <option value="">请选择部门</option>
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
                  value={editForm.position || ''}
                  onChange={handleInputChange}
                >
                  <option value="">请选择职位</option>
                  <option value="合伙人">合伙人</option>
                  <option value="资深律师">资深律师</option>
                  <option value="律师">律师</option>
                  <option value="实习律师">实习律师</option>
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
                  <option value="active">在职</option>
                  <option value="on_leave">休假</option>
                  <option value="inactive">离职</option>
                </Form.Select>
              </Form.Group>
            </Col>
            <Col md={6}>
              <Form.Group className="mb-3">
                <Form.Label>咨询费用（元/小时）</Form.Label>
                <Form.Control
                  type="number"
                  name="hourlyRate"
                  value={editForm.hourlyRate || ''}
                  onChange={handleInputChange}
                  min="0"
                />
              </Form.Group>
            </Col>
          </Row>

          <Row>
            <Col md={12}>
              <Form.Group className="mb-3">
                <Form.Label>办公地址</Form.Label>
                <Form.Control
                  type="text"
                  name="address"
                  value={editForm.address || ''}
                  onChange={handleInputChange}
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
                  rows={4}
                  name="profile"
                  value={editForm.profile || ''}
                  onChange={handleInputChange}
                  placeholder="请输入个人简介"
                />
              </Form.Group>
            </Col>
          </Row>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => setShowEditModal(false)}>
            取消
          </Button>
          <Button variant="primary" onClick={handleSave} disabled={saving}>
            {saving ? (
              <>
                <Spinner as="span" animation="border" size="sm" role="status" aria-hidden="true" className="me-2" />
                保存中...
              </>
            ) : (
              '保存'
            )}
          </Button>
        </Modal.Footer>
      </Modal>
    </div>
  );
};

export default LawyerDetailPage;