import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Badge, Spinner, Alert, Button, Card, Table, Modal, Form } from 'react-bootstrap';
import {
  FaArrowLeft,
  FaCheck,
  FaXmark,
  FaRotate,
  FaFileLines,
  FaClock,
  FaUser,
  FaBuilding,
  FaCalendar,
  FaComment,
  FaTriangleExclamation
} from 'react-icons/fa6';

interface ApprovalRecord {
  id: number;
  approvalId: number;
  approver: string;
  approverId: number;
  action: 'approve' | 'reject';
  comment: string;
  createTime: string;
}

interface ApprovalDetail {
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
  records: ApprovalRecord[];
}

const ApprovalDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [approval, setApproval] = useState<ApprovalDetail | null>(null);
  const [actionModalVisible, setActionModalVisible] = useState<boolean>(false);
  const [currentAction, setCurrentAction] = useState<'approve' | 'reject' | null>(null);
  const [comment, setComment] = useState<string>('');

  useEffect(() => {
    if (id) {
      fetchApprovalDetail(parseInt(id));
    }
  }, [id]);

  const fetchApprovalDetail = async (approvalId: number) => {
    try {
      setLoading(true);
      setError(null);

      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 1000));

      const mockData: ApprovalDetail = {
        id: approvalId,
        type: '请假申请',
        title: '事假申请 - 家中有事',
        content: '因家中有急事需要请假3天，从1月16日到1月18日。家中老人需要照顾，希望领导能够批准。',
        applicant: '张三',
        applicantId: 1,
        department: '技术部',
        createTime: '2024-01-15 09:00:00',
        status: 'pending',
        urgency: 'normal',
        currentApprover: '李四',
        currentApproverId: 2,
        records: [
          {
            id: 1,
            approvalId: approvalId,
            approver: '王五',
            approverId: 3,
            action: 'approve',
            comment: '同意申请，请安排好工作交接',
            createTime: '2024-01-15 10:30:00'
          },
          {
            id: 2,
            approvalId: approvalId,
            approver: '赵六',
            approverId: 4,
            action: 'reject',
            comment: '当前项目比较紧急，建议调整请假时间',
            createTime: '2024-01-15 11:45:00'
          }
        ]
      };

      setApproval(mockData);
    } catch (error) {
      console.error('Failed to fetch approval detail:', error);
      setError('获取审批详情失败');
    } finally {
      setLoading(false);
    }
  };

  const renderStatusBadge = (status: string) => {
    switch (status) {
      case 'approved':
        return <Badge bg="success">已通过</Badge>;
      case 'rejected':
        return <Badge bg="danger">已拒绝</Badge>;
      case 'pending':
        return <Badge bg="warning">待审批</Badge>;
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

  const handleAction = (action: 'approve' | 'reject') => {
    setCurrentAction(action);
    setActionModalVisible(true);
  };

  const submitAction = async () => {
    if (!approval || !currentAction || !comment.trim()) return;

    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 1000));

      // 更新本地状态
      const updatedApproval = {
        ...approval,
        status: (currentAction === 'approve' ? 'approved' : 'rejected') as 'approved' | 'rejected',
        records: [
          ...approval.records,
          {
            id: approval.records.length + 1,
            approvalId: approval.id,
            approver: '当前用户',
            approverId: 1,
            action: currentAction,
            comment: comment,
            createTime: new Date().toLocaleString()
          }
        ]
      };

      setApproval(updatedApproval);
      setActionModalVisible(false);
      setComment('');

      alert(currentAction === 'approve' ? '审批通过' : '已拒绝');
    } catch (error) {
      console.error('Failed to handle approval:', error);
      alert('操作失败');
    }
  };

  const handleCancel = async () => {
    if (!approval) return;

    if (window.confirm('确定要撤回这个审批申请吗？')) {
      try {
        // 模拟撤回操作
        await new Promise(resolve => setTimeout(resolve, 500));

        const updatedApproval = {
          ...approval,
          status: 'cancelled' as const
        };

        setApproval(updatedApproval);
        alert('撤回成功');
      } catch (error) {
        console.error('Failed to cancel approval:', error);
        alert('撤回失败');
      }
    }
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

  if (error) {
    return (
      <div className="container mt-4">
        <Alert variant="danger">{error}</Alert>
        <Button variant="primary" onClick={() => navigate('/approval')}>
          返回列表
        </Button>
      </div>
    );
  }

  if (!approval) {
    return (
      <div className="container mt-4">
        <Alert variant="warning">审批不存在</Alert>
        <Button variant="primary" onClick={() => navigate('/approval')}>
          返回列表
        </Button>
      </div>
    );
  }

  const canApprove = approval.status === 'pending' && approval.currentApproverId === 1;
  const canCancel = approval.status === 'pending' && approval.applicantId === 1;

  return (
    <div className="approval-detail p-4">
      <Card>
        <Card.Header>
          <div className="d-flex justify-content-between align-items-center">
            <div>
              <h3 className="mb-2">{approval.title}</h3>
              <div className="d-flex gap-2">
                {renderStatusBadge(approval.status)}
                {renderUrgencyBadge(approval.urgency)}
              </div>
            </div>
            <Button variant="outline-secondary" onClick={() => navigate('/approval')}>
              <FaArrowLeft className="w-4 h-4 me-2" />
              返回列表
            </Button>
          </div>
        </Card.Header>
        <Card.Body>
          {/* 基本信息 */}
          <div className="row mb-4">
            <div className="col-md-6">
              <div className="d-flex align-items-center mb-3">
                <FaFileLines className="w-5 h-5 text-primary me-2" />
                <strong>申请类型：</strong>
                <span className="ms-2">{approval.type}</span>
              </div>
              <div className="d-flex align-items-center mb-3">
                <FaBuilding className="w-5 h-5 text-primary me-2" />
                <strong>申请部门：</strong>
                <span className="ms-2">{approval.department}</span>
              </div>
              <div className="d-flex align-items-center mb-3">
                <FaUser className="w-5 h-5 text-primary me-2" />
                <strong>申请人：</strong>
                <span className="ms-2">{approval.applicant}</span>
              </div>
              <div className="d-flex align-items-center mb-3">
                <FaCalendar className="w-5 h-5 text-primary me-2" />
                <strong>申请时间：</strong>
                <span className="ms-2">{approval.createTime}</span>
              </div>
            </div>
            <div className="col-md-6">
              {approval.currentApprover && (
                <div className="d-flex align-items-center mb-3">
                  <FaUser className="w-5 h-5 text-primary me-2" />
                  <strong>当前审批人：</strong>
                  <span className="ms-2">{approval.currentApprover}</span>
                </div>
              )}
              <div className="d-flex align-items-center mb-3">
                <FaClock className="w-5 h-5 text-primary me-2" />
                <strong>状态：</strong>
                <span className="ms-2">{renderStatusBadge(approval.status)}</span>
              </div>
              <div className="d-flex align-items-center mb-3">
                <span className="text-primary me-2">⚠️</span>
                <strong>紧急程度：</strong>
                <span className="ms-2">{renderUrgencyBadge(approval.urgency)}</span>
              </div>
            </div>
          </div>

          {/* 申请内容 */}
          <Card className="mb-4">
            <Card.Header>
              <h5 className="mb-0">
                <FaFileLines className="w-5 h-5 me-2" />
                申请内容
              </h5>
            </Card.Header>
            <Card.Body>
              <div className="p-3 bg-light rounded">
                {approval.content}
              </div>
            </Card.Body>
          </Card>

          {/* 审批记录 */}
          {approval.records && approval.records.length > 0 && (
            <Card className="mb-4">
              <Card.Header>
                <h5 className="mb-0">
                  <FaComment className="w-5 h-5 me-2" />
                  审批记录
                </h5>
              </Card.Header>
              <Card.Body>
                <Table striped>
                  <thead>
                    <tr>
                      <th>审批人</th>
                      <th>操作</th>
                      <th>时间</th>
                      <th>意见</th>
                    </tr>
                  </thead>
                  <tbody>
                    {approval.records.map((record) => (
                      <tr key={record.id}>
                        <td>
                          <div className="d-flex align-items-center">
                            <FaUser className="w-4 h-4 me-2" />
                            {record.approver}
                          </div>
                        </td>
                        <td>
                          <Badge bg={record.action === 'approve' ? 'success' : 'danger'}>
                            {record.action === 'approve' ? '通过' : '拒绝'}
                          </Badge>
                        </td>
                        <td>{record.createTime}</td>
                        <td>{record.comment}</td>
                      </tr>
                    ))}
                  </tbody>
                </Table>
              </Card.Body>
            </Card>
          )}

          {/* 操作按钮 */}
          <div className="d-flex gap-2">
            <Button variant="outline-secondary" onClick={() => navigate('/approval')}>
              返回列表
            </Button>

            {canApprove && (
              <>
                <Button
                  variant="success"
                  onClick={() => handleAction('approve')}
                >
                  <FaCheck className="w-4 h-4 me-2" />
                  通过
                </Button>
                <Button
                  variant="danger"
                  onClick={() => handleAction('reject')}
                >
                  <FaXmark className="w-4 h-4 me-2" />
                  拒绝
                </Button>
              </>
            )}

            {canCancel && (
              <Button
                variant="warning"
                onClick={handleCancel}
              >
                <FaRotate className="w-4 h-4 me-2" />
                撤回
              </Button>
            )}
          </div>
        </Card.Body>
      </Card>

      {/* 审批操作模态框 */}
      <Modal show={actionModalVisible} onHide={() => {
        setActionModalVisible(false);
        setComment('');
      }}>
        <Modal.Header closeButton>
          <Modal.Title>
            {currentAction === 'approve' ? '审批通过' : '审批拒绝'}
          </Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <Form>
            <Form.Group className="mb-3">
              <Form.Label>审批意见</Form.Label>
              <Form.Control
                as="textarea"
                rows={4}
                placeholder={currentAction === 'approve' ? '请输入通过理由' : '请输入拒绝理由'}
                value={comment}
                onChange={(e) => setComment(e.target.value)}
              />
            </Form.Group>
          </Form>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => {
            setActionModalVisible(false);
            setComment('');
          }}>
            取消
          </Button>
          <Button
            variant={currentAction === 'approve' ? 'success' : 'danger'}
            onClick={submitAction}
            disabled={!comment.trim()}
          >
            {currentAction === 'approve' ? '通过' : '拒绝'}
          </Button>
        </Modal.Footer>
      </Modal>
    </div>
  );
};

export default ApprovalDetail;