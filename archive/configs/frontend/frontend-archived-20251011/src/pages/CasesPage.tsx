import React, { useState, useEffect } from "react";
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
} from "react-bootstrap";
import { useParams, useNavigate } from "react-router-dom";
import { getCase, updateCase, deleteCase } from "../services/caseService";
import { Case, UpdateCaseRequest } from "../types";

const CasesPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [caseData, setCaseData] = useState<Case | null>(null);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState("overview");
  const [editing, setEditing] = useState(false);
  const [formData, setFormData] = useState({
    title: "",
    description: "",
    client_id: 0,
    case_type: "civil",
    priority: "medium",
    status: "pending",
  });

  useEffect(() => {
    if (id) {
      loadCase(parseInt(id));
    }
  }, [id]);

  const loadCase = async (caseId: number) => {
    setLoading(true);
    try {
      const response = await getCase(caseId);
      setCaseData(response);
      setFormData({
        title: response.title,
        description: response.description,
        client_id: response.client_id,
        case_type: response.case_type,
        priority: response.priority,
        status: response.status,
      });
    } catch (error) {
      console.error("Failed to load case", error);
    } finally {
      setLoading(false);
    }
  };

  const handleEdit = () => {
    setEditing(true);
  };

  const handleCancelEdit = () => {
    setEditing(false);
    if (caseData) {
      setFormData({
        title: caseData.title,
        description: caseData.description,
        client_id: caseData.client_id,
        case_type: caseData.case_type,
        priority: caseData.priority,
        status: caseData.status,
      });
    }
  };

  const handleChange = (e: React.ChangeEvent<any>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!caseData) return;

    try {
      const data: UpdateCaseRequest = formData;
      const updatedCase = await updateCase(caseData.id, data);
      setCaseData(updatedCase);
      setEditing(false);
    } catch (error) {
      console.error("Failed to update case", error);
    }
  };

  const handleDelete = async () => {
    if (!caseData) return;

    if (
      window.confirm(
        "Are you sure you want to delete this case? This action cannot be undone.",
      )
    ) {
      try {
        await deleteCase(caseData.id);
        navigate("/cases");
      } catch (error) {
        console.error("Failed to delete case", error);
      }
    }
  };

  const getStatusBadgeClass = (status: string) => {
    switch (status) {
      case "pending":
        return "bg-warning";
      case "active":
        return "bg-primary";
      case "closed":
        return "bg-success";
      case "suspended":
        return "bg-secondary";
      default:
        return "bg-secondary";
    }
  };

  const getPriorityBadgeClass = (priority: string) => {
    switch (priority) {
      case "low":
        return "bg-info";
      case "medium":
        return "bg-warning";
      case "high":
        return "bg-danger";
      case "urgent":
        return "bg-danger";
      default:
        return "bg-secondary";
    }
  };

  const getCaseTypeBadgeClass = (caseType: string) => {
    switch (caseType) {
      case "civil":
        return "bg-primary";
      case "criminal":
        return "bg-danger";
      case "commercial":
        return "bg-success";
      case "administrative":
        return "bg-info";
      default:
        return "bg-secondary";
    }
  };

  // Get case type display text
  const getCaseTypeText = (caseType: string) => {
    switch (caseType) {
      case "civil":
        return "Civil";
      case "criminal":
        return "Criminal";
      case "commercial":
        return "Commercial";
      case "administrative":
        return "Administrative";
      default:
        return caseType;
    }
  };

  // Get priority display text
  const getPriorityText = (priority: string) => {
    switch (priority) {
      case "low":
        return "Low";
      case "medium":
        return "Medium";
      case "high":
        return "High";
      case "urgent":
        return "Urgent";
      default:
        return priority;
    }
  };

  // Get status display text
  const getStatusText = (status: string) => {
    switch (status) {
      case "pending":
        return "Pending";
      case "active":
        return "Active";
      case "closed":
        return "Closed";
      case "suspended":
        return "Suspended";
      default:
        return status;
    }
  };

  if (loading) {
    return (
      <div
        className="d-flex justify-content-center align-items-center"
        style={{ height: "50vh" }}
      >
        <Spinner animation="border" />
        <span className="ms-2">Loading case details...</span>
      </div>
    );
  }

  if (!caseData) {
    return (
      <div className="text-center py-5">
        <i className="fas fa-exclamation-triangle fa-3x text-warning mb-3"></i>
        <h5>Case not found</h5>
        <p className="text-muted">The requested case could not be found</p>
        <Button variant="primary" onClick={() => navigate("/cases")}>
          <i className="fas fa-arrow-left me-2"></i>
          Back to Cases
        </Button>
      </div>
    );
  }

  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <div>
          <Button
            variant="outline-secondary"
            onClick={() => navigate("/cases")}
            className="mb-2"
          >
            <i className="fas fa-arrow-left me-2"></i>
            Back to Cases
          </Button>
          <h1>
            Case #{caseData.id}: {caseData.title}
          </h1>
        </div>
        <div className="d-flex">
          <Dropdown className="me-2">
            <Dropdown.Toggle variant="outline-secondary" id="actions-dropdown">
              <i className="fas fa-ellipsis-h me-2"></i>
              Actions
            </Dropdown.Toggle>
            <Dropdown.Menu>
              <Dropdown.Item onClick={handleEdit}>
                <i className="fas fa-edit me-2"></i>
                Edit Case
              </Dropdown.Item>
              <Dropdown.Item>
                <i className="fas fa-copy me-2"></i>
                Duplicate Case
              </Dropdown.Item>
              <Dropdown.Item>
                <i className="fas fa-history me-2"></i>
                View History
              </Dropdown.Item>
              <Dropdown.Divider />
              <Dropdown.Item className="text-danger" onClick={handleDelete}>
                <i className="fas fa-trash me-2"></i>
                Delete Case
              </Dropdown.Item>
            </Dropdown.Menu>
          </Dropdown>
          <Button variant="primary" onClick={handleEdit}>
            <i className="fas fa-edit me-2"></i>
            Edit Case
          </Button>
        </div>
      </div>

      <Row>
        <Col md={8}>
          <Card className="mb-4">
            <Card.Header className="d-flex justify-content-between align-items-center">
              <span>Case Overview</span>
              <div>
                <Badge
                  variant={getStatusBadgeClass(caseData.status)}
                  className="me-2"
                >
                  {getStatusText(caseData.status)}
                </Badge>
                <Badge
                  variant={getPriorityBadgeClass(caseData.priority)}
                  className="me-2"
                >
                  {getPriorityText(caseData.priority)}
                </Badge>
                <Badge bg={getCaseTypeBadgeClass(caseData.case_type)}>
                  {getCaseTypeText(caseData.case_type)}
                </Badge>
              </div>
            </Card.Header>
            <Card.Body>
              {editing ? (
                <Form onSubmit={handleSubmit}>
                  <Row>
                    <Col md={12}>
                      <Form.Group className="mb-3">
                        <Form.Label>
                          Title <span className="text-danger">*</span>
                        </Form.Label>
                        <Form.Control
                          type="text"
                          name="title"
                          value={formData.title}
                          onChange={handleChange}
                          required
                          placeholder="Enter case title"
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
                          rows={4}
                          name="description"
                          value={formData.description}
                          onChange={handleChange}
                          placeholder="Enter case description"
                        />
                      </Form.Group>
                    </Col>
                  </Row>
                  <Row>
                    <Col md={6}>
                      <Form.Group className="mb-3">
                        <Form.Label>
                          Client ID <span className="text-danger">*</span>
                        </Form.Label>
                        <Form.Control
                          type="number"
                          name="client_id"
                          value={formData.client_id}
                          onChange={handleChange}
                          required
                          placeholder="Enter client ID"
                        />
                      </Form.Group>
                    </Col>
                    <Col md={6}>
                      <Form.Group className="mb-3">
                        <Form.Label>Case Type</Form.Label>
                        <Form.Select
                          name="caseType"
                          value={formData.case_type}
                          onChange={handleChange}
                        >
                          <option value="civil">
                            {getCaseTypeText("civil")}
                          </option>
                          <option value="criminal">
                            {getCaseTypeText("criminal")}
                          </option>
                          <option value="commercial">
                            {getCaseTypeText("commercial")}
                          </option>
                          <option value="administrative">
                            {getCaseTypeText("administrative")}
                          </option>
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
                          <option value="low">{getPriorityText("low")}</option>
                          <option value="medium">
                            {getPriorityText("medium")}
                          </option>
                          <option value="high">
                            {getPriorityText("high")}
                          </option>
                          <option value="urgent">
                            {getPriorityText("urgent")}
                          </option>
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
                          <option value="pending">
                            {getStatusText("pending")}
                          </option>
                          <option value="active">
                            {getStatusText("active")}
                          </option>
                          <option value="closed">
                            {getStatusText("closed")}
                          </option>
                          <option value="suspended">
                            {getStatusText("suspended")}
                          </option>
                        </Form.Select>
                      </Form.Group>
                    </Col>
                  </Row>
                  <div className="d-flex justify-content-end">
                    <Button
                      variant="secondary"
                      onClick={handleCancelEdit}
                      className="me-2"
                    >
                      <i className="fas fa-times me-2"></i>
                      Cancel
                    </Button>
                    <Button variant="primary" type="submit">
                      <i className="fas fa-save me-2"></i>
                      Save Changes
                    </Button>
                  </div>
                </Form>
              ) : (
                <>
                  <div className="mb-4">
                    <h5>Description</h5>
                    <p className="text-muted">{caseData.description}</p>
                  </div>

                  <Row>
                    <Col md={6}>
                      <div className="mb-3">
                        <small className="text-muted">Created</small>
                        <div className="fw-bold">
                          {new Date(caseData.created_at).toLocaleDateString()}{" "}
                          at{" "}
                          {new Date(caseData.created_at).toLocaleTimeString()}
                        </div>
                      </div>
                      <div className="mb-3">
                        <small className="text-muted">Last Updated</small>
                        <div className="fw-bold">
                          {new Date(caseData.updated_at).toLocaleDateString()}{" "}
                          at{" "}
                          {new Date(caseData.updated_at).toLocaleTimeString()}
                        </div>
                      </div>
                    </Col>
                    <Col md={6}>
                      <div className="mb-3">
                        <small className="text-muted">Assigned Lawyer</small>
                        <div className="fw-bold">
                          {caseData.lawyer_name || "Unassigned"}
                        </div>
                      </div>
                      <div className="mb-3">
                        <small className="text-muted">Client</small>
                        <div className="fw-bold">{caseData.client_name}</div>
                      </div>
                    </Col>
                  </Row>
                </>
              )}
            </Card.Body>
          </Card>

          <Tabs
            activeKey={activeTab}
            onSelect={(k) => setActiveTab(k || "overview")}
            className="mb-4"
          >
            <Tab eventKey="overview" title="Overview">
              <Card>
                <Card.Body>
                  <h5>Case Timeline</h5>
                  <div className="timeline">
                    <div className="timeline-item mb-4">
                      <div className="d-flex">
                        <div className="timeline-icon bg-primary rounded-circle d-flex align-items-center justify-content-center me-3">
                          <i className="fas fa-calendar-plus text-white"></i>
                        </div>
                        <div className="timeline-content">
                          <h6 className="mb-1">Case Created</h6>
                          <p className="mb-0 text-muted">
                            Case was created in the system
                          </p>
                          <small className="text-muted">
                            {new Date(caseData.created_at).toLocaleString()}
                          </small>
                        </div>
                      </div>
                    </div>
                    <div className="timeline-item mb-4">
                      <div className="d-flex">
                        <div className="timeline-icon bg-success rounded-circle d-flex align-items-center justify-content-center me-3">
                          <i className="fas fa-user-plus text-white"></i>
                        </div>
                        <div className="timeline-content">
                          <h6 className="mb-1">Lawyer Assigned</h6>
                          <p className="mb-0 text-muted">
                            Jane Smith was assigned to this case
                          </p>
                          <small className="text-muted">2 days ago</small>
                        </div>
                      </div>
                    </div>
                    <div className="timeline-item">
                      <div className="d-flex">
                        <div className="timeline-icon bg-warning rounded-circle d-flex align-items-center justify-content-center me-3">
                          <i className="fas fa-clock text-white"></i>
                        </div>
                        <div className="timeline-content">
                          <h6 className="mb-1">Status Updated</h6>
                          <p className="mb-0 text-muted">
                            Case status changed to Active
                          </p>
                          <small className="text-muted">1 day ago</small>
                        </div>
                      </div>
                    </div>
                  </div>
                </Card.Body>
              </Card>
            </Tab>
            <Tab eventKey="documents" title="Documents">
              <Card>
                <Card.Body>
                  <div className="d-flex justify-content-between align-items-center mb-4">
                    <h5>Case Documents</h5>
                    <Button variant="primary">
                      <i className="fas fa-plus me-2"></i>
                      Upload Document
                    </Button>
                  </div>

                  <div className="document-list">
                    <div className="document-item mb-3 p-3 border rounded">
                      <div className="d-flex justify-content-between align-items-center">
                        <div className="d-flex align-items-center">
                          <div className="bg-light rounded me-3 p-2">
                            <i className="fas fa-file-pdf fa-2x text-danger"></i>
                          </div>
                          <div>
                            <h6 className="mb-1">Contract Agreement.pdf</h6>
                            <small className="text-muted">
                              Uploaded 2 days ago
                            </small>
                          </div>
                        </div>
                        <div>
                          <Button
                            variant="outline-primary"
                            size="sm"
                            className="me-2"
                          >
                            <i className="fas fa-download"></i>
                          </Button>
                          <Button variant="outline-danger" size="sm">
                            <i className="fas fa-trash"></i>
                          </Button>
                        </div>
                      </div>
                    </div>

                    <div className="document-item mb-3 p-3 border rounded">
                      <div className="d-flex justify-content-between align-items-center">
                        <div className="d-flex align-items-center">
                          <div className="bg-light rounded me-3 p-2">
                            <i className="fas fa-file-word fa-2x text-primary"></i>
                          </div>
                          <div>
                            <h6 className="mb-1">Case Summary.docx</h6>
                            <small className="text-muted">
                              Uploaded 1 day ago
                            </small>
                          </div>
                        </div>
                        <div>
                          <Button
                            variant="outline-primary"
                            size="sm"
                            className="me-2"
                          >
                            <i className="fas fa-download"></i>
                          </Button>
                          <Button variant="outline-danger" size="sm">
                            <i className="fas fa-trash"></i>
                          </Button>
                        </div>
                      </div>
                    </div>
                  </div>
                </Card.Body>
              </Card>
            </Tab>
            <Tab eventKey="notes" title="Notes">
              <Card>
                <Card.Body>
                  <div className="d-flex justify-content-between align-items-center mb-4">
                    <h5>Case Notes</h5>
                    <Button variant="primary">
                      <i className="fas fa-plus me-2"></i>
                      Add Note
                    </Button>
                  </div>

                  <div className="note-list">
                    <div className="note-item mb-4 p-3 border rounded">
                      <div className="d-flex justify-content-between mb-2">
                        <div className="fw-bold">Initial Assessment</div>
                        <small className="text-muted">
                          2 days ago by Jane Smith
                        </small>
                      </div>
                      <p className="mb-2">
                        Initial assessment of the case indicates strong grounds
                        for litigation. Client has provided all necessary
                        documentation and evidence.
                      </p>
                      <div className="d-flex justify-content-end">
                        <Button
                          variant="outline-primary"
                          size="sm"
                          className="me-2"
                        >
                          <i className="fas fa-edit me-1"></i>
                          Edit
                        </Button>
                        <Button variant="outline-danger" size="sm">
                          <i className="fas fa-trash me-1"></i>
                          Delete
                        </Button>
                      </div>
                    </div>

                    <div className="note-item mb-4 p-3 border rounded">
                      <div className="d-flex justify-content-between mb-2">
                        <div className="fw-bold">Client Meeting</div>
                        <small className="text-muted">
                          1 day ago by John Doe
                        </small>
                      </div>
                      <p className="mb-2">
                        Met with client to discuss case progress. Client is
                        satisfied with current developments and has authorized
                        us to proceed with filing.
                      </p>
                      <div className="d-flex justify-content-end">
                        <Button
                          variant="outline-primary"
                          size="sm"
                          className="me-2"
                        >
                          <i className="fas fa-edit me-1"></i>
                          Edit
                        </Button>
                        <Button variant="outline-danger" size="sm">
                          <i className="fas fa-trash me-1"></i>
                          Delete
                        </Button>
                      </div>
                    </div>
                  </div>
                </Card.Body>
              </Card>
            </Tab>
            <Tab eventKey="activities" title="Activities">
              <Card>
                <Card.Body>
                  <h5>Case Activities</h5>
                  <div className="activity-list">
                    <div className="activity-item mb-3 p-3 border rounded">
                      <div className="d-flex">
                        <div className="activity-icon bg-primary rounded-circle d-flex align-items-center justify-content-center me-3">
                          <i className="fas fa-user-plus text-white"></i>
                        </div>
                        <div className="activity-content">
                          <h6 className="mb-1">New Client Added</h6>
                          <p className="mb-0 text-muted">
                            John Smith was added as a new client
                          </p>
                          <small className="text-muted">2 days ago</small>
                        </div>
                      </div>
                    </div>
                    <div className="activity-item mb-3 p-3 border rounded">
                      <div className="d-flex">
                        <div className="activity-icon bg-success rounded-circle d-flex align-items-center justify-content-center me-3">
                          <i className="fas fa-gavel text-white"></i>
                        </div>
                        <div className="activity-content">
                          <h6 className="mb-1">Case Updated</h6>
                          <p className="mb-0 text-muted">
                            Case #1234 status changed to Active
                          </p>
                          <small className="text-muted">1 day ago</small>
                        </div>
                      </div>
                    </div>
                    <div className="activity-item p-3 border rounded">
                      <div className="d-flex">
                        <div className="activity-icon bg-warning rounded-circle d-flex align-items-center justify-content-center me-3">
                          <i className="fas fa-clock text-white"></i>
                        </div>
                        <div className="activity-content">
                          <h6 className="mb-1">Pending Case</h6>
                          <p className="mb-0 text-muted">
                            Case #1234 requires attention
                          </p>
                          <small className="text-muted">12 hours ago</small>
                        </div>
                      </div>
                    </div>
                  </div>
                </Card.Body>
              </Card>
            </Tab>
          </Tabs>
        </Col>

        <Col md={4}>
          <Card className="mb-4">
            <Card.Header>Client Information</Card.Header>
            <Card.Body>
              <div className="d-flex align-items-center mb-3">
                <div
                  className="bg-light rounded-circle d-flex align-items-center justify-content-center me-3"
                  style={{ width: "48px", height: "48px" }}
                >
                  <i className="fas fa-user text-muted"></i>
                </div>
                <div>
                  <div className="fw-bold">{caseData.client_name}</div>
                  <div className="small text-muted">
                    Client ID: {caseData.client_id}
                  </div>
                </div>
              </div>

              <div className="mb-3">
                <small className="text-muted">Email</small>
                <div className="fw-bold">{caseData.client?.email || "-"}</div>
              </div>

              <div className="mb-3">
                <small className="text-muted">Phone</small>
                <div className="fw-bold">{caseData.client?.phone || "-"}</div>
              </div>

              <div className="mb-3">
                <small className="text-muted">Address</small>
                <div className="fw-bold">-</div>
              </div>

              <div className="mb-3">
                <small className="text-muted">Company</small>
                <div className="fw-bold">-</div>
              </div>

              <Button variant="outline-primary" className="w-100">
                <i className="fas fa-envelope me-2"></i>
                Contact Client
              </Button>
            </Card.Body>
          </Card>

          <Card className="mb-4">
            <Card.Header>Assigned Lawyer</Card.Header>
            <Card.Body>
              {caseData.lawyer_name ? (
                <>
                  <div className="d-flex align-items-center mb-3">
                    <div
                      className="bg-light rounded-circle d-flex align-items-center justify-content-center me-3"
                      style={{ width: "48px", height: "48px" }}
                    >
                      <i className="fas fa-user-tie text-muted"></i>
                    </div>
                    <div>
                      <div className="fw-bold">{caseData.lawyer_name}</div>
                      <div className="small text-muted">
                        Lawyer ID: {caseData.lawyer_id}
                      </div>
                    </div>
                  </div>

                  <div className="mb-3">
                    <small className="text-muted">Email</small>
                    <div className="fw-bold">
                      {caseData.lawyer?.email || "-"}
                    </div>
                  </div>

                  <div className="mb-3">
                    <small className="text-muted">Phone</small>
                    <div className="fw-bold">{"-"}</div>
                  </div>

                  <Button variant="outline-primary" className="w-100">
                    <i className="fas fa-envelope me-2"></i>
                    Contact Lawyer
                  </Button>
                </>
              ) : (
                <div className="text-center py-3">
                  <i className="fas fa-user-slash fa-2x text-muted mb-3"></i>
                  <h6>No Lawyer Assigned</h6>
                  <p className="text-muted">
                    This case has not been assigned to a lawyer yet
                  </p>
                  <Button variant="primary">
                    <i className="fas fa-user-plus me-2"></i>
                    Assign Lawyer
                  </Button>
                </div>
              )}
            </Card.Body>
          </Card>

          <Card>
            <Card.Header>Case Statistics</Card.Header>
            <Card.Body>
              <div className="stat-item mb-3">
                <div className="d-flex justify-content-between">
                  <span>Documents</span>
                  <span className="fw-bold">5</span>
                </div>
                <div className="progress mt-1">
                  <div
                    className="progress-bar bg-primary"
                    role="progressbar"
                    style={{ width: "60%" }}
                  ></div>
                </div>
              </div>

              <div className="stat-item mb-3">
                <div className="d-flex justify-content-between">
                  <span>Notes</span>
                  <span className="fw-bold">3</span>
                </div>
                <div className="progress mt-1">
                  <div
                    className="progress-bar bg-success"
                    role="progressbar"
                    style={{ width: "30%" }}
                  ></div>
                </div>
              </div>

              <div className="stat-item">
                <div className="d-flex justify-content-between">
                  <span>Activities</span>
                  <span className="fw-bold">12</span>
                </div>
                <div className="progress mt-1">
                  <div
                    className="progress-bar bg-warning"
                    role="progressbar"
                    style={{ width: "90%" }}
                  ></div>
                </div>
              </div>
            </Card.Body>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default CasesPage;
