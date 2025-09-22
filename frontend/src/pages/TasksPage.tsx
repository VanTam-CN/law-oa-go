import React, { useState, useEffect } from 'react';
import { getTasks, createTask, updateTask, deleteTask } from '../services/taskService';
import { Task, TaskListRequest, CreateTaskRequest, UpdateTaskRequest } from '../types';

const TasksPage: React.FC = () => {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');
  const [priorityFilter, setPriorityFilter] = useState('all');
  const [assigneeFilter, setAssigneeFilter] = useState('all');
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [formData, setFormData] = useState({
    title: '',
    description: '',
    dueDate: '',
    priority: 'medium',
    status: 'pending',
    assigneeId: 0
  });

  useEffect(() => {
    loadTasks();
  }, [currentPage, search, statusFilter, priorityFilter, assigneeFilter]);

  const loadTasks = async () => {
    setLoading(true);
    try {
      const params: TaskListRequest = {
        page: currentPage,
        page_size: 10,
        search: search || undefined,
        status: statusFilter !== 'all' ? statusFilter : undefined,
        priority: priorityFilter !== 'all' ? priorityFilter : undefined,
        assigned_to: assigneeFilter !== 'all' ? parseInt(assigneeFilter) : undefined
      };
      const response = await getTasks(params);
      setTasks(response.data);
      setTotalPages(Math.ceil(response.pagination.total / response.pagination.page_size));
    } catch (error) {
      console.error('Failed to load tasks', error);
    } finally {
      setLoading(false);
    }
  };

  const handleShowModal = (task?: Task) => {
    if (task) {
      setEditingTask(task);
      setFormData({
        title: task.title,
        description: task.description,
        dueDate: task.due_date ? task.due_date.split('T')[0] : '',
        priority: task.priority,
        status: task.status,
        assigneeId: task.assigned_to || 0
      });
    } else {
      setEditingTask(null);
      setFormData({
        title: '',
        description: '',
        dueDate: '',
        priority: 'medium',
        status: 'pending',
        assigneeId: 0
      });
    }
    setShowModal(true);
  };

  const handleCloseModal = () => {
    setShowModal(false);
    setEditingTask(null);
  };

  const handleChange = (e: React.ChangeEvent<any>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editingTask) {
        // Update existing task
        const data: UpdateTaskRequest = {
          title: formData.title,
          description: formData.description,
          due_date: formData.dueDate ? new Date(formData.dueDate).toISOString() : undefined,
          priority: formData.priority as 'low' | 'medium' | 'high' | 'urgent',
          status: formData.status as 'pending' | 'in_progress' | 'completed' | 'cancelled',
          assigned_to: formData.assigneeId || undefined
        };
        const updatedTask = await updateTask(editingTask.id, data);
        setTasks(prev => prev.map(t => t.id === updatedTask.id ? updatedTask : t));
      } else {
        // Create new task
        const data: CreateTaskRequest = {
          title: formData.title,
          description: formData.description,
          due_date: formData.dueDate ? new Date(formData.dueDate).toISOString() : undefined,
          priority: formData.priority as 'low' | 'medium' | 'high' | 'urgent',
          status: formData.status as 'pending' | 'in_progress' | 'completed' | 'cancelled',
          assigned_to: formData.assigneeId || undefined
        };
        const newTask = await createTask(data);
        setTasks(prev => [newTask, ...prev]);
      }
      handleCloseModal();
    } catch (error) {
      console.error('Failed to save task', error);
    }
  };

  const handleDelete = async (id: number) => {
    if (window.confirm('Are you sure you want to delete this task? This action cannot be undone.')) {
      try {
        await deleteTask(id);
        setTasks(prev => prev.filter(t => t.id !== id));
      } catch (error) {
        console.error('Failed to delete task', error);
      }
    }
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setCurrentPage(1);
    loadTasks();
  };

  const getStatusBadgeClass = (status: string) => {
    switch (status) {
      case 'pending': return 'bg-warning';
      case 'in_progress': return 'bg-primary';
      case 'completed': return 'bg-success';
      case 'cancelled': return 'bg-secondary';
      default: return 'bg-secondary';
    }
  };

  const getPriorityBadgeClass = (priority: string) => {
    switch (priority) {
      case 'low': return 'bg-info';
      case 'medium': return 'bg-warning';
      case 'high': return 'bg-danger';
      case 'urgent': return 'bg-danger';
      default: return 'bg-secondary';
    }
  };

  // Get status display text
  const getStatusText = (status: string) => {
    switch (status) {
      case 'pending': return 'Pending';
      case 'in_progress': return 'In Progress';
      case 'completed': return 'Completed';
      case 'cancelled': return 'Cancelled';
      default: return status;
    }
  };

  // Get priority display text
  const getPriorityText = (priority: string) => {
    switch (priority) {
      case 'low': return 'Low';
      case 'medium': return 'Medium';
      case 'high': return 'High';
      case 'urgent': return 'Urgent';
      default: return priority;
    }
  };

  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h1>Tasks</h1>
        <Button variant="primary" onClick={() => handleShowModal()}>
          <i className="fas fa-plus me-2"></i>
          Add Task
        </Button>
      </div>

      <Card className="mb-4">
        <Card.Body>
          <Row>
            <Col md={6}>
              <Form onSubmit={handleSearch}>
                <InputGroup>
                  <Form.Control
                    type="text"
                    placeholder="Search tasks by title or description..."
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                  />
                  <Button variant="outline-secondary" type="submit">
                    <i className="fas fa-search"></i>
                  </Button>
                </InputGroup>
              </Form>
            </Col>
            <Col md={6}>
              <div className="d-flex justify-content-end">
                <Dropdown className="me-2">
                  <Dropdown.Toggle variant="outline-secondary" id="status-dropdown">
                    Status: {statusFilter === 'all' ? 'All' : getStatusText(statusFilter)}
                  </Dropdown.Toggle>
                  <Dropdown.Menu>
                    <Dropdown.Item onClick={() => setStatusFilter('all')}>All</Dropdown.Item>
                    <Dropdown.Item onClick={() => setStatusFilter('pending')}>{getStatusText('pending')}</Dropdown.Item>
                    <Dropdown.Item onClick={() => setStatusFilter('in_progress')}>{getStatusText('in_progress')}</Dropdown.Item>
                    <Dropdown.Item onClick={() => setStatusFilter('completed')}>{getStatusText('completed')}</Dropdown.Item>
                    <Dropdown.Item onClick={() => setStatusFilter('cancelled')}>{getStatusText('cancelled')}</Dropdown.Item>
                  </Dropdown.Menu>
                </Dropdown>
                <Dropdown className="me-2">
                  <Dropdown.Toggle variant="outline-secondary" id="priority-dropdown">
                    Priority: {priorityFilter === 'all' ? 'All' : getPriorityText(priorityFilter)}
                  </Dropdown.Toggle>
                  <Dropdown.Menu>
                    <Dropdown.Item onClick={() => setPriorityFilter('all')}>All</Dropdown.Item>
                    <Dropdown.Item onClick={() => setPriorityFilter('low')}>{getPriorityText('low')}</Dropdown.Item>
                    <Dropdown.Item onClick={() => setPriorityFilter('medium')}>{getPriorityText('medium')}</Dropdown.Item>
                    <Dropdown.Item onClick={() => setPriorityFilter('high')}>{getPriorityText('high')}</Dropdown.Item>
                    <Dropdown.Item onClick={() => setPriorityFilter('urgent')}>{getPriorityText('urgent')}</Dropdown.Item>
                  </Dropdown.Menu>
                </Dropdown>
                <Dropdown className="me-2">
                  <Dropdown.Toggle variant="outline-secondary" id="assignee-dropdown">
                    Assignee: {assigneeFilter === 'all' ? 'All' : 'Selected'}
                  </Dropdown.Toggle>
                  <Dropdown.Menu>
                    <Dropdown.Item onClick={() => setAssigneeFilter('all')}>All</Dropdown.Item>
                    <Dropdown.Item onClick={() => setAssigneeFilter('1')}>Jane Smith</Dropdown.Item>
                    <Dropdown.Item onClick={() => setAssigneeFilter('2')}>John Doe</Dropdown.Item>
                    <Dropdown.Item onClick={() => setAssigneeFilter('3')}>Michael Johnson</Dropdown.Item>
                  </Dropdown.Menu>
                </Dropdown>
                <Button variant="outline-primary">
                  <i className="fas fa-filter me-2"></i>
                  More Filters
                </Button>
              </div>
            </Col>
          </Row>
        </Card.Body>
      </Card>

      {loading ? (
        <div className="d-flex justify-content-center align-items-center" style={{ height: '400px' }}>
          <Spinner animation="border" />
          <span className="ms-2">Loading tasks...</span>
        </div>
      ) : (
        <Card>
          <Card.Body>
            <Table striped bordered hover responsive>
              <thead>
                <tr>
                  <th>Task</th>
                  <th>Assignee</th>
                  <th>Due Date</th>
                  <th>Priority</th>
                  <th>Status</th>
                  <th>Created</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {tasks.map(task => (
                  <tr key={task.id}>
                    <td>
                      <div>
                        <div className="fw-bold">{task.title}</div>
                        <div className="small text-muted">{task.description}</div>
                      </div>
                    </td>
                    <td>
                      {task.assigned_to_name ? (
                        <div className="d-flex align-items-center">
                          <div className="bg-light rounded-circle d-flex align-items-center justify-content-center me-2" style={{ width: '24px', height: '24px' }}>
                            <i className="fas fa-user text-muted"></i>
                          </div>
                          <div>{task.assigned_to_name}</div>
                        </div>
                      ) : (
                        <span className="text-muted">Unassigned</span>
                      )}
                    </td>
                    <td>
                      {task.due_date ? (
                        <div>
                          <div>{new Date(task.due_date).toLocaleDateString()}</div>
                          <div className="small text-muted">
                            {new Date(task.due_date).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                          </div>
                        </div>
                      ) : (
                        <span className="text-muted">No due date</span>
                      )}
                    </td>
                    <td>
                      <Badge bg={getPriorityBadgeClass(task.priority)}>
                        {getPriorityText(task.priority)}
                      </Badge>
                    </td>
                    <td>
                      <Badge bg={getStatusBadgeClass(task.status)}>
                        {getStatusText(task.status)}
                      </Badge>
                    </td>
                    <td>{new Date(task.created_at).toLocaleDateString()}</td>
                    <td>
                      <div className="d-flex">
                        <Button
                          variant="outline-primary"
                          size="sm"
                          className="me-2"
                          onClick={() => handleShowModal(task)}
                        >
                          <i className="fas fa-edit"></i>
                        </Button>
                        <Button
                          variant="outline-info"
                          size="sm"
                          className="me-2"
                        >
                          <i className="fas fa-eye"></i>
                        </Button>
                        <Button
                          variant="outline-danger"
                          size="sm"
                          onClick={() => handleDelete(task.id)}
                        >
                          <i className="fas fa-trash"></i>
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </Table>
            
            {tasks.length === 0 && (
              <div className="text-center py-5">
                <i className="fas fa-tasks fa-3x text-muted mb-3"></i>
                <h5>No tasks found</h5>
                <p className="text-muted">Try adjusting your search or filter criteria</p>
                <Button variant="primary" onClick={() => handleShowModal()}>
                  <i className="fas fa-plus me-2"></i>
                  Add Your First Task
                </Button>
              </div>
            )}
          </Card.Body>
        </Card>
      )}

      {/* Task Modal */}
      <Modal show={showModal} onHide={handleCloseModal} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>
            {editingTask ? (
              <span><i className="fas fa-edit me-2"></i> Edit Task</span>
            ) : (
              <span><i className="fas fa-tasks me-2"></i> Add Task</span>
            )}
          </Modal.Title>
        </Modal.Header>
        <Form onSubmit={handleSubmit}>
          <Modal.Body>
            <Row>
              <Col md={12}>
                <Form.Group className="mb-3">
                  <Form.Label>Title <span className="text-danger">*</span></Form.Label>
                  <Form.Control
                    type="text"
                    name="title"
                    value={formData.title}
                    onChange={handleChange}
                    required
                    placeholder="Enter task title"
                  />
                </Form.Group>
              </Col>
            </Row>
            <Row>
              <Col md={12}>
                <Form.Group className="mb-3">
                  <Form.Label>Description</Form.Label>
                  <Form.Control
                    as="textarea"
                    rows={3}
                    name="description"
                    value={formData.description}
                    onChange={handleChange}
                    placeholder="Enter task description"
                  />
                </Form.Group>
              </Col>
            </Row>
            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Due Date</Form.Label>
                  <Form.Control
                    type="date"
                    name="dueDate"
                    value={formData.dueDate}
                    onChange={handleChange}
                  />
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Assignee</Form.Label>
                  <Form.Select
                    name="assigneeId"
                    value={formData.assigneeId}
                    onChange={handleChange}
                  >
                    <option value={0}>Unassigned</option>
                    <option value={1}>Jane Smith</option>
                    <option value={2}>John Doe</option>
                    <option value={3}>Michael Johnson</option>
                  </Form.Select>
                </Form.Group>
              </Col>
            </Row>
            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Priority</Form.Label>
                  <Form.Select
                    name="priority"
                    value={formData.priority}
                    onChange={handleChange}
                  >
                    <option value="low">{getPriorityText('low')}</option>
                    <option value="medium">{getPriorityText('medium')}</option>
                    <option value="high">{getPriorityText('high')}</option>
                    <option value="urgent">{getPriorityText('urgent')}</option>
                  </Form.Select>
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Status</Form.Label>
                  <Form.Select
                    name="status"
                    value={formData.status}
                    onChange={handleChange}
                  >
                    <option value="pending">{getStatusText('pending')}</option>
                    <option value="in_progress">{getStatusText('in_progress')}</option>
                    <option value="completed">{getStatusText('completed')}</option>
                    <option value="cancelled">{getStatusText('cancelled')}</option>
                  </Form.Select>
                </Form.Group>
              </Col>
            </Row>
          </Modal.Body>
          <Modal.Footer>
            <Button variant="secondary" onClick={handleCloseModal}>
              <i className="fas fa-times me-2"></i>
              Cancel
            </Button>
            <Button variant="primary" type="submit">
              {editingTask ? (
                <span><i className="fas fa-save me-2"></i> Update Task</span>
              ) : (
                <span><i className="fas fa-plus me-2"></i> Add Task</span>
              )}
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>
    </div>
  );
};

export default TasksPage;