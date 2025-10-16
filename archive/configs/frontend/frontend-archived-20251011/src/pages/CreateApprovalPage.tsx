import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  FaArrowLeft,
  FaCheck,
  FaTriangleExclamation,
  FaFileLines,
  FaUser,
  FaBuilding,
  FaClock,
  FaPaperPlane
} from 'react-icons/fa6';

interface ApprovalFormData {
  type: string;
  title: string;
  content: string;
  urgency: 'normal' | 'urgent' | 'very_urgent';
  department: string;
}

interface FormError {
  field: string;
  message: string;
}

const CreateApproval: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<boolean>(false);
  const [formData, setFormData] = useState<ApprovalFormData>({
    type: 'leave',
    title: '',
    content: '',
    urgency: 'normal',
    department: '技术部'
  });
  const [errors, setErrors] = useState<FormError[]>([]);

  const approvalTypes = [
    { value: 'leave', label: '请假申请' },
    { value: 'expense', label: '报销申请' },
    { value: 'purchase', label: '采购申请' },
    { value: 'project', label: '项目申请' },
    { value: 'other', label: '其他申请' }
  ];

  const departments = [
    '技术部',
    '销售部',
    '市场部',
    '人力资源部',
    '财务部',
    '行政部',
    '产品部',
    '运营部'
  ];

  const validateForm = (): boolean => {
    const newErrors: FormError[] = [];

    if (!formData.title.trim()) {
      newErrors.push({ field: 'title', message: '请输入申请标题' });
    }

    if (!formData.content.trim()) {
      newErrors.push({ field: 'content', message: '请输入申请内容' });
    } else if (formData.content.length < 10) {
      newErrors.push({ field: 'content', message: '申请内容至少需要10个字符' });
    }

    if (!formData.type) {
      newErrors.push({ field: 'type', message: '请选择申请类型' });
    }

    if (!formData.department) {
      newErrors.push({ field: 'department', message: '请选择申请部门' });
    }

    setErrors(newErrors);
    return newErrors.length === 0;
  };

  const getFieldError = (field: string): string | null => {
    const error = errors.find(e => e.field === field);
    return error ? error.message : null;
  };

  const handleInputChange = (field: keyof ApprovalFormData, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));

    // 清除该字段的错误
    setErrors(prev => prev.filter(e => e.field !== field));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    setLoading(true);
    setError(null);

    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 2000));

      console.log('提交审批申请:', formData);

      setSuccess(true);

      // 3秒后自动跳转到审批列表
      setTimeout(() => {
        navigate('/approval');
      }, 3000);

    } catch (error) {
      console.error('Failed to create approval:', error);
      setError('提交审批申请失败，请重试');
    } finally {
      setLoading(false);
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

  if (success) {
    return (
      <div className="container mt-4">
        <Card>
          <Card.Body className="text-center py-5">
            <FaCheck className="w-16 h-16 text-success mx-auto mb-4" />
            <h3>申请提交成功！</h3>
            <p className="text-muted">您的审批申请已成功提交，正在等待审批。</p>
            <p className="text-muted">页面将在3秒后自动跳转到审批列表...</p>
            <Button variant="primary" onClick={() => navigate('/approval')}>
              立即查看
            </Button>
          </Card.Body>
        </Card>
      </div>
    );
  }

  return (
    <div className="create-approval p-4">
      <Card>
        <Card.Header>
          <div className="d-flex justify-content-between align-items-center">
            <div className="d-flex align-items-center">
              <Button variant="outline-secondary" onClick={() => navigate('/approval')} className="me-3">
                <FaArrowLeft className="w-4 h-4" />
              </Button>
              <h4 className="mb-0">新建审批申请</h4>
            </div>
            <Badge bg="info">草稿</Badge>
          </div>
        </Card.Header>
        <Card.Body>
          {error && (
            <Alert variant="danger" onClose={() => setError(null)} dismissible>
              {error}
            </Alert>
          )}

          <Form onSubmit={handleSubmit}>
            {/* 基本信息 */}
            <Card className="mb-4">
              <Card.Header>
                <h5 className="mb-0">
                  <FaFileLines className="w-5 h-5 me-2" />
                  基本信息
                </h5>
              </Card.Header>
              <Card.Body>
                <div className="row">
                  <div className="col-md-6">
                    <Form.Group className="mb-3">
                      <Form.Label>
                        <FaUser className="w-4 h-4 me-1" />
                        申请类型
                      </Form.Label>
                      <Form.Select
                        value={formData.type}
                        onChange={(e) => handleInputChange('type', e.target.value)}
                        isInvalid={!!getFieldError('type')}
                      >
                        <option value="">请选择申请类型</option>
                        {approvalTypes.map(type => (
                          <option key={type.value} value={type.value}>
                            {type.label}
                          </option>
                        ))}
                      </Form.Select>
                      {getFieldError('type') && (
                        <Form.Control.Feedback type="invalid">
                          {getFieldError('type')}
                        </Form.Control.Feedback>
                      )}
                    </Form.Group>

                    <Form.Group className="mb-3">
                      <Form.Label>
                        <FaBuilding className="w-4 h-4 me-1" />
                        申请部门
                      </Form.Label>
                      <Form.Select
                        value={formData.department}
                        onChange={(e) => handleInputChange('department', e.target.value)}
                        isInvalid={!!getFieldError('department')}
                      >
                        <option value="">请选择申请部门</option>
                        {departments.map(dept => (
                          <option key={dept} value={dept}>
                            {dept}
                          </option>
                        ))}
                      </Form.Select>
                      {getFieldError('department') && (
                        <Form.Control.Feedback type="invalid">
                          {getFieldError('department')}
                        </Form.Control.Feedback>
                      )}
                    </Form.Group>
                  </div>

                  <div className="col-md-6">
                    <Form.Group className="mb-3">
                      <Form.Label>
                        <FaClock className="w-4 h-4 me-1" />
                        紧急程度
                      </Form.Label>
                      <div className="d-flex gap-2">
                        {(['normal', 'urgent', 'very_urgent'] as const).map(urgency => (
                          <Button
                            key={urgency}
                            variant={formData.urgency === urgency ? 'primary' : 'outline-primary'}
                            onClick={() => handleInputChange('urgency', urgency)}
                            className="d-flex align-items-center"
                          >
                            {renderUrgencyBadge(urgency)}
                            <span className="ms-2">
                              {urgency === 'normal' ? '普通' : urgency === 'urgent' ? '紧急' : '特急'}
                            </span>
                          </Button>
                        ))}
                      </div>
                    </Form.Group>

                    <Form.Group className="mb-3">
                      <Form.Label>
                        <FaFileLines className="w-4 h-4 me-1" />
                        申请标题
                      </Form.Label>
                      <Form.Control
                        type="text"
                        placeholder="请输入申请标题"
                        value={formData.title}
                        onChange={(e) => handleInputChange('title', e.target.value)}
                        isInvalid={!!getFieldError('title')}
                      />
                      {getFieldError('title') && (
                        <Form.Control.Feedback type="invalid">
                          {getFieldError('title')}
                        </Form.Control.Feedback>
                      )}
                    </Form.Group>
                  </div>
                </div>
              </Card.Body>
            </Card>

            {/* 申请内容 */}
            <Card className="mb-4">
              <Card.Header>
                <h5 className="mb-0">
                  <FaFileLines className="w-5 h-5 me-2" />
                  申请内容
                </h5>
              </Card.Header>
              <Card.Body>
                <Form.Group>
                  <Form.Label>详细说明</Form.Label>
                  <Form.Control
                    as="textarea"
                    rows={8}
                    placeholder="请详细描述您的申请内容，包括原因、时间、具体要求等..."
                    value={formData.content}
                    onChange={(e) => handleInputChange('content', e.target.value)}
                    isInvalid={!!getFieldError('content')}
                  />
                  <Form.Text className="text-muted">
                    请提供详细的申请信息，有助于审批人更好地理解您的需求。
                  </Form.Text>
                  {getFieldError('content') && (
                    <Form.Control.Feedback type="invalid">
                      {getFieldError('content')}
                    </Form.Control.Feedback>
                  )}
                </Form.Group>
              </Card.Body>
            </Card>

            {/* 注意事项 */}
            <Card className="mb-4">
              <Card.Header>
                <h5 className="mb-0">
                  <FaTriangleExclamation className="w-5 h-5 me-2" />
                  注意事项
                </h5>
              </Card.Header>
              <Card.Body>
                <ul className="mb-0">
                  <li>请确保申请信息真实准确</li>
                  <li>紧急申请请选择合适的紧急程度</li>
                  <li>提交后可在"我的申请"中查看审批进度</li>
                  <li>如有疑问，请联系相关审批人员</li>
                </ul>
              </Card.Body>
            </Card>

            {/* 操作按钮 */}
            <div className="d-flex justify-content-between">
              <Button variant="outline-secondary" onClick={() => navigate('/approval')}>
                <FaArrowLeft className="w-4 h-4 me-2" />
                返回列表
              </Button>

              <div className="d-flex gap-2">
                <Button variant="outline-primary" disabled={loading}>
                  保存草稿
                </Button>
                <Button
                  type="submit"
                  variant="primary"
                  disabled={loading}
                >
                  {loading ? (
                    <>
                      <Spinner as="span" animation="border" size="sm" role="status" aria-hidden="true" />
                      <span className="ms-2">提交中...</span>
                    </>
                  ) : (
                    <>
                      <FaPaperPlane className="w-4 h-4 me-2" />
                      提交申请
                    </>
                  )}
                </Button>
              </div>
            </div>
          </Form>
        </Card.Body>
      </Card>
    </div>
  );
};

export default CreateApproval;