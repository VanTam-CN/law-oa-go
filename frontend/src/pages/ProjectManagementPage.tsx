import React, { useState } from 'react';
import { Card, Button, Table, Form, Badge, Modal, InputGroup } from 'react-bootstrap';
import {
  PlusIcon,
  PencilIcon,
  TrashIcon,
  MagnifyingGlassIcon,
  FunnelIcon,
  UserIcon,
  CalendarIcon,
  BriefcaseIcon,
  CheckCircleIcon,
  ClockIcon,
  PauseIcon,
  XMarkIcon
} from '@heroicons/react/outline';

interface Project {
  id: string;
  name: string;
  client: string;
  type: string;
  status: 'planning' | 'active' | 'completed' | 'suspended';
  startDate: string;
  endDate?: string;
  lawyer: string;
  description: string;
  budget?: number;
  progress?: number;
}

const ProjectManagement: React.FC = () => {
  const [projects, setProjects] = useState<Project[]>([
    {
      id: '1',
      name: '张三诉李四借款纠纷案',
      client: '张三',
      type: '民事诉讼',
      status: 'active',
      startDate: '2024-01-15',
      lawyer: '王律师',
      description: '民间借贷纠纷，涉及金额50万元',
      budget: 50000,
      progress: 65
    },
    {
      id: '2',
      name: 'ABC公司合同审查',
      client: 'ABC公司',
      type: '合同审查',
      status: 'planning',
      startDate: '2024-02-01',
      lawyer: '李律师',
      description: '企业合同条款审查和风险评估',
      budget: 30000,
      progress: 20
    },
    {
      id: '3',
      name: '知识产权保护咨询',
      client: '科技公司',
      type: '法律咨询',
      status: 'completed',
      startDate: '2023-12-01',
      endDate: '2024-01-10',
      lawyer: '赵律师',
      description: '商标注册和专利申请咨询',
      budget: 25000,
      progress: 100
    },
    {
      id: '4',
      name: '企业并购法律服务',
      client: '投资公司',
      type: '公司法务',
      status: 'active',
      startDate: '2024-01-20',
      lawyer: '钱律师',
      description: '企业并购全程法律服务',
      budget: 200000,
      progress: 40
    },
    {
      id: '5',
      name: '劳动争议仲裁',
      client: '制造企业',
      type: '劳动仲裁',
      status: 'suspended',
      startDate: '2024-01-10',
      lawyer: '孙律师',
      description: '员工劳动争议仲裁代理',
      budget: 40000,
      progress: 30
    }
  ]);

  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingProject, setEditingProject] = useState<Project | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterType, setFilterType] = useState('all');
  const [filterStatus, setFilterStatus] = useState('all');

  const projectTypes = [
    '民事诉讼',
    '刑事辩护',
    '合同审查',
    '法律咨询',
    '知识产权',
    '公司法务',
    '劳动仲裁',
    '行政诉讼'
  ];

  const getStatusBadge = (status: Project['status']) => {
    switch (status) {
      case 'planning':
        return <Badge bg="info">计划中</Badge>;
      case 'active':
        return <Badge bg="success">进行中</Badge>;
      case 'completed':
        return <Badge bg="secondary">已完成</Badge>;
      case 'suspended':
        return <Badge bg="warning">暂停</Badge>;
      default:
        return <Badge bg="light">未知</Badge>;
    }
  };

  const getStatusIcon = (status: Project['status']) => {
    switch (status) {
      case 'planning':
        return <ClockIcon className="w-4 h-4" />;
      case 'active':
        return <CheckCircleIcon className="w-4 h-4" />;
      case 'completed':
        return <CheckCircleIcon className="w-4 h-4" />;
      case 'suspended':
        return <PauseIcon className="w-4 h-4" />;
      default:
        return <XMarkIcon className="w-4 h-4" />;
    }
  };

  const getFilteredProjects = () => {
    let filtered = projects;

    if (searchTerm) {
      filtered = filtered.filter(project =>
        project.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        project.client.toLowerCase().includes(searchTerm.toLowerCase()) ||
        project.lawyer.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }

    if (filterType !== 'all') {
      filtered = filtered.filter(project => project.type === filterType);
    }

    if (filterStatus !== 'all') {
      filtered = filtered.filter(project => project.status === filterStatus);
    }

    return filtered;
  };

  const handleAdd = () => {
    setEditingProject(null);
    setModalVisible(true);
  };

  const handleEdit = (project: Project) => {
    setEditingProject(project);
    setModalVisible(true);
  };

  const handleDelete = (id: string) => {
    if (window.confirm('确定要删除这个项目吗？')) {
      setProjects(projects.filter(p => p.id !== id));
      alert('删除成功');
    }
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);

    setLoading(true);
    try {
      const projectData = {
        name: formData.get('name') as string,
        client: formData.get('client') as string,
        type: formData.get('type') as string,
        status: formData.get('status') as Project['status'],
        startDate: formData.get('startDate') as string,
        endDate: formData.get('endDate') as string || undefined,
        lawyer: formData.get('lawyer') as string,
        description: formData.get('description') as string,
        budget: formData.get('budget') ? parseInt(formData.get('budget') as string) : undefined,
        progress: formData.get('progress') ? parseInt(formData.get('progress') as string) : undefined
      };

      if (editingProject) {
        // 编辑项目
        setProjects(projects.map(p =>
          p.id === editingProject.id ? { ...p, ...projectData } : p
        ));
        alert('项目更新成功');
      } else {
        // 新增项目
        const newProject: Project = {
          ...projectData,
          id: Date.now().toString(),
        };
        setProjects([...projects, newProject]);
        alert('项目创建成功');
      }

      setModalVisible(false);
    } catch (error) {
      alert('操作失败');
    } finally {
      setLoading(false);
    }
  };

  const filteredProjects = getFilteredProjects();

  // 统计数据
  const stats = {
    total: projects.length,
    active: projects.filter(p => p.status === 'active').length,
    completed: projects.filter(p => p.status === 'completed').length,
    planning: projects.filter(p => p.status === 'planning').length,
    totalBudget: projects.reduce((sum, p) => sum + (p.budget || 0), 0)
  };

  return (
    <div className="project-management p-4">
      {/* 头部统计 */}
      <div className="row mb-4">
        <div className="col-md-2">
          <Card className="text-center">
            <Card.Body>
              <h3>{stats.total}</h3>
              <p className="text-muted mb-0">总项目数</p>
            </Card.Body>
          </Card>
        </div>
        <div className="col-md-2">
          <Card className="text-center bg-success text-white">
            <Card.Body>
              <h3>{stats.active}</h3>
              <p className="mb-0">进行中</p>
            </Card.Body>
          </Card>
        </div>
        <div className="col-md-2">
          <Card className="text-center bg-secondary text-white">
            <Card.Body>
              <h3>{stats.completed}</h3>
              <p className="mb-0">已完成</p>
            </Card.Body>
          </Card>
        </div>
        <div className="col-md-2">
          <Card className="text-center bg-info text-white">
            <Card.Body>
              <h3>{stats.planning}</h3>
              <p className="mb-0">计划中</p>
            </Card.Body>
          </Card>
        </div>
        <div className="col-md-4">
          <Card className="text-center bg-warning">
            <Card.Body>
              <h3>¥{stats.totalBudget.toLocaleString()}</h3>
              <p className="mb-0">总预算</p>
            </Card.Body>
          </Card>
        </div>
      </div>

      {/* 主内容 */}
      <Card>
        <Card.Header>
          <div className="d-flex justify-content-between align-items-center">
            <div>
              <h4 className="mb-0">项目管理</h4>
              <p className="text-muted mb-0">管理律所的各类法律服务项目</p>
            </div>
            <Button variant="primary" onClick={handleAdd}>
              <PlusIcon className="w-4 h-4 me-2" />
              新建项目
            </Button>
          </div>
        </Card.Header>
        <Card.Body>
          {/* 搜索和筛选 */}
          <div className="row mb-3">
            <div className="col-md-4">
              <InputGroup>
                <InputGroup.Text>
                  <MagnifyingGlassIcon className="w-4 h-4" />
                </InputGroup.Text>
                <Form.Control
                  type="text"
                  placeholder="搜索项目名称或客户..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                />
              </InputGroup>
            </div>
            <div className="col-md-3">
              <Form.Select
                value={filterType}
                onChange={(e) => setFilterType(e.target.value)}
              >
                <option value="all">所有类型</option>
                {projectTypes.map(type => (
                  <option key={type} value={type}>{type}</option>
                ))}
              </Form.Select>
            </div>
            <div className="col-md-3">
              <Form.Select
                value={filterStatus}
                onChange={(e) => setFilterStatus(e.target.value)}
              >
                <option value="all">所有状态</option>
                <option value="planning">计划中</option>
                <option value="active">进行中</option>
                <option value="completed">已完成</option>
                <option value="suspended">暂停</option>
              </Form.Select>
            </div>
            <div className="col-md-2">
              <Button variant="outline-secondary" onClick={() => {
                setSearchTerm('');
                setFilterType('all');
                setFilterStatus('all');
              }}>
                <FunnelIcon className="w-4 h-4 me-2" />
                重置
              </Button>
            </div>
          </div>

          {/* 项目列表 */}
          <div className="table-responsive">
            <Table striped hover>
              <thead>
                <tr>
                  <th>项目名称</th>
                  <th>客户</th>
                  <th>类型</th>
                  <th>状态</th>
                  <th>负责律师</th>
                  <th>开始日期</th>
                  <th>预算</th>
                  <th>进度</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {filteredProjects.map((project) => (
                  <tr key={project.id}>
                    <td>
                      <div>
                        <strong>{project.name}</strong>
                        <br />
                        <small className="text-muted">{project.description}</small>
                      </div>
                    </td>
                    <td>
                      <div className="d-flex align-items-center">
                        <UserIcon className="w-4 h-4 me-1" />
                        {project.client}
                      </div>
                    </td>
                    <td>
                      <Badge bg="light" text="dark">
                        <BriefcaseIcon className="w-3 h-3 me-1" />
                        {project.type}
                      </Badge>
                    </td>
                    <td>
                      <div className="d-flex align-items-center">
                        {getStatusIcon(project.status)}
                        <span className="ms-2">{getStatusBadge(project.status)}</span>
                      </div>
                    </td>
                    <td>{project.lawyer}</td>
                    <td>
                      <div className="d-flex align-items-center">
                        <CalendarIcon className="w-4 h-4 me-1" />
                        {project.startDate}
                      </div>
                    </td>
                    <td>
                      {project.budget ? `¥${project.budget.toLocaleString()}` : '-'}
                    </td>
                    <td>
                      {project.progress !== undefined && (
                        <div className="d-flex align-items-center">
                          <div className="progress flex-grow-1 me-2" style={{ height: '8px' }}>
                            <div
                              className="progress-bar"
                              role="progressbar"
                              style={{ width: `${project.progress}%` }}
                              aria-valuenow={project.progress}
                              aria-valuemin={0}
                              aria-valuemax={100}
                            ></div>
                          </div>
                          <small>{project.progress}%</small>
                        </div>
                      )}
                    </td>
                    <td>
                      <div className="btn-group" role="group">
                        <Button
                          variant="outline-primary"
                          size="sm"
                          onClick={() => handleEdit(project)}
                        >
                          <PencilIcon className="w-4 h-4" />
                        </Button>
                        <Button
                          variant="outline-danger"
                          size="sm"
                          onClick={() => handleDelete(project.id)}
                        >
                          <TrashIcon className="w-4 h-4" />
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </Table>
          </div>

          {filteredProjects.length === 0 && (
            <div className="text-center py-4">
              <BriefcaseIcon className="w-16 h-16 text-muted mx-auto mb-3" />
              <h5>没有找到匹配的项目</h5>
              <p className="text-muted">请尝试调整搜索条件或创建新项目</p>
            </div>
          )}
        </Card.Body>
      </Card>

      {/* 项目创建/编辑模态框 */}
      <Modal show={modalVisible} onHide={() => setModalVisible(false)} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>
            {editingProject ? '编辑项目' : '新建项目'}
          </Modal.Title>
        </Modal.Header>
        <Form onSubmit={handleSubmit}>
          <Modal.Body>
            <div className="row">
              <div className="col-md-6">
                <Form.Group className="mb-3">
                  <Form.Label>项目名称</Form.Label>
                  <Form.Control
                    type="text"
                    name="name"
                    defaultValue={editingProject?.name}
                    required
                  />
                </Form.Group>
              </div>
              <div className="col-md-6">
                <Form.Group className="mb-3">
                  <Form.Label>客户</Form.Label>
                  <Form.Control
                    type="text"
                    name="client"
                    defaultValue={editingProject?.client}
                    required
                  />
                </Form.Group>
              </div>
            </div>

            <div className="row">
              <div className="col-md-6">
                <Form.Group className="mb-3">
                  <Form.Label>项目类型</Form.Label>
                  <Form.Select
                    name="type"
                    defaultValue={editingProject?.type}
                    required
                  >
                    <option value="">请选择项目类型</option>
                    {projectTypes.map(type => (
                      <option key={type} value={type}>{type}</option>
                    ))}
                  </Form.Select>
                </Form.Group>
              </div>
              <div className="col-md-6">
                <Form.Group className="mb-3">
                  <Form.Label>项目状态</Form.Label>
                  <Form.Select
                    name="status"
                    defaultValue={editingProject?.status}
                    required
                  >
                    <option value="planning">计划中</option>
                    <option value="active">进行中</option>
                    <option value="completed">已完成</option>
                    <option value="suspended">暂停</option>
                  </Form.Select>
                </Form.Group>
              </div>
            </div>

            <div className="row">
              <div className="col-md-6">
                <Form.Group className="mb-3">
                  <Form.Label>负责律师</Form.Label>
                  <Form.Control
                    type="text"
                    name="lawyer"
                    defaultValue={editingProject?.lawyer}
                    required
                  />
                </Form.Group>
              </div>
              <div className="col-md-3">
                <Form.Group className="mb-3">
                  <Form.Label>开始日期</Form.Label>
                  <Form.Control
                    type="date"
                    name="startDate"
                    defaultValue={editingProject?.startDate}
                    required
                  />
                </Form.Group>
              </div>
              <div className="col-md-3">
                <Form.Group className="mb-3">
                  <Form.Label>结束日期</Form.Label>
                  <Form.Control
                    type="date"
                    name="endDate"
                    defaultValue={editingProject?.endDate}
                  />
                </Form.Group>
              </div>
            </div>

            <div className="row">
              <div className="col-md-6">
                <Form.Group className="mb-3">
                  <Form.Label>预算</Form.Label>
                  <Form.Control
                    type="number"
                    name="budget"
                    defaultValue={editingProject?.budget}
                    min="0"
                  />
                </Form.Group>
              </div>
              <div className="col-md-6">
                <Form.Group className="mb-3">
                  <Form.Label>进度 (%)</Form.Label>
                  <Form.Control
                    type="number"
                    name="progress"
                    defaultValue={editingProject?.progress}
                    min="0"
                    max="100"
                  />
                </Form.Group>
              </div>
            </div>

            <Form.Group className="mb-3">
              <Form.Label>项目描述</Form.Label>
              <Form.Control
                as="textarea"
                rows={3}
                name="description"
                defaultValue={editingProject?.description}
              />
            </Form.Group>
          </Modal.Body>
          <Modal.Footer>
            <Button variant="secondary" onClick={() => setModalVisible(false)}>
              取消
            </Button>
            <Button type="submit" variant="primary" disabled={loading}>
              {loading ? '处理中...' : (editingProject ? '更新' : '创建')}
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>
    </div>
  );
};

export default ProjectManagement;