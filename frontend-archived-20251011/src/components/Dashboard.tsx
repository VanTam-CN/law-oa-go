import React, { useState, useEffect } from "react";
import { Row, Col, Card, Spinner, Button, Dropdown } from "react-bootstrap";
import { getClientStats } from "../services/clientService";
import { getCaseStats } from "../services/caseService";
import { CaseStats, ClientStats } from "../types";

interface DashboardProps {
  timeRange: string;
  setTimeRange: (range: string) => void;
}

const Dashboard: React.FC<DashboardProps> = ({ timeRange, setTimeRange }) => {
  const [clientStats, setClientStats] = useState<ClientStats | null>(null);
  const [caseStats, setCaseStats] = useState<CaseStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchData();
  }, [timeRange]);

  const fetchData = async () => {
    setLoading(true);
    try {
      const [clientData, caseData] = await Promise.all([
        getClientStats(),
        getCaseStats(),
      ]);
      setClientStats(clientData);
      setCaseStats(caseData);
    } catch (error) {
      console.error("Failed to fetch dashboard data", error);
    } finally {
      setLoading(false);
    }
  };

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

  if (loading) {
    return (
      <div
        className="d-flex justify-content-center align-items-center"
        style={{ height: "50vh" }}
      >
        <Spinner animation="border" />
        <span className="ms-2">Loading dashboard data...</span>
      </div>
    );
  }

  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h1>Dashboard</h1>
        <div className="d-flex align-items-center">
          <Dropdown className="me-2">
            <Dropdown.Toggle
              variant="outline-secondary"
              id="time-range-dropdown"
            >
              Time Range: {timeRange === "today" && "Today"}
              {timeRange === "this_week" && "This Week"}
              {timeRange === "this_month" && "This Month"}
              {timeRange === "this_year" && "This Year"}
            </Dropdown.Toggle>
            <Dropdown.Menu>
              <Dropdown.Item onClick={() => setTimeRange("today")}>
                Today
              </Dropdown.Item>
              <Dropdown.Item onClick={() => setTimeRange("this_week")}>
                This Week
              </Dropdown.Item>
              <Dropdown.Item onClick={() => setTimeRange("this_month")}>
                This Month
              </Dropdown.Item>
              <Dropdown.Item onClick={() => setTimeRange("this_year")}>
                This Year
              </Dropdown.Item>
            </Dropdown.Menu>
          </Dropdown>
          <Button variant="outline-primary">
            <i className="fas fa-download me-2"></i>
            Export Report
          </Button>
        </div>
      </div>

      <Row>
        <Col md={3}>
          <Card className="stat-card bg-primary text-white">
            <Card.Body>
              <div className="d-flex justify-content-between">
                <div>
                  <Card.Title className="mb-0">Total Clients</Card.Title>
                  <div className="number">{clientStats?.total || 0}</div>
                </div>
                <div className="stat-icon">
                  <i className="fas fa-users fa-2x"></i>
                </div>
              </div>
              <div className="mt-2">
                <span className="small">
                  <i className="fas fa-arrow-up me-1"></i>
                  12% from last period
                </span>
              </div>
            </Card.Body>
          </Card>
        </Col>
        <Col md={3}>
          <Card className="stat-card bg-success text-white">
            <Card.Body>
              <div className="d-flex justify-content-between">
                <div>
                  <Card.Title className="mb-0">Active Cases</Card.Title>
                  <div className="number">{caseStats?.active || 0}</div>
                </div>
                <div className="stat-icon">
                  <i className="fas fa-gavel fa-2x"></i>
                </div>
              </div>
              <div className="mt-2">
                <span className="small">
                  <i className="fas fa-arrow-up me-1"></i>
                  8% from last period
                </span>
              </div>
            </Card.Body>
          </Card>
        </Col>
        <Col md={3}>
          <Card className="stat-card bg-warning text-white">
            <Card.Body>
              <div className="d-flex justify-content-between">
                <div>
                  <Card.Title className="mb-0">Pending Cases</Card.Title>
                  <div className="number">{caseStats?.pending || 0}</div>
                </div>
                <div className="stat-icon">
                  <i className="fas fa-clock fa-2x"></i>
                </div>
              </div>
              <div className="mt-2">
                <span className="small">
                  <i className="fas fa-arrow-down me-1"></i>
                  3% from last period
                </span>
              </div>
            </Card.Body>
          </Card>
        </Col>
        <Col md={3}>
          <Card className="stat-card bg-info text-white">
            <Card.Body>
              <div className="d-flex justify-content-between">
                <div>
                  <Card.Title className="mb-0">Closed Cases</Card.Title>
                  <div className="number">{caseStats?.closed || 0}</div>
                </div>
                <div className="stat-icon">
                  <i className="fas fa-check-circle fa-2x"></i>
                </div>
              </div>
              <div className="mt-2">
                <span className="small">
                  <i className="fas fa-arrow-up me-1"></i>
                  15% from last period
                </span>
              </div>
            </Card.Body>
          </Card>
        </Col>
      </Row>

      <Row className="mt-4">
        <Col md={8}>
          <Card>
            <Card.Header className="d-flex justify-content-between align-items-center">
              <span>Case Trends</span>
              <Button variant="outline-primary" size="sm">
                <i className="fas fa-chart-line me-1"></i>
                View Details
              </Button>
            </Card.Header>
            <Card.Body>
              <div
                className="chart-placeholder"
                style={{
                  height: "300px",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                }}
              >
                <div className="text-center">
                  <i className="fas fa-chart-bar fa-3x text-muted mb-3"></i>
                  <p className="text-muted">
                    Case trends chart would be displayed here
                  </p>
                </div>
              </div>
            </Card.Body>
          </Card>
        </Col>
        <Col md={4}>
          <Card>
            <Card.Header>Recent Activities</Card.Header>
            <Card.Body>
              <div className="activity-list">
                <div className="activity-item mb-3">
                  <div className="d-flex">
                    <div className="activity-icon bg-primary rounded-circle d-flex align-items-center justify-content-center me-3">
                      <i className="fas fa-user-plus text-white"></i>
                    </div>
                    <div className="activity-content">
                      <h6 className="mb-1">New Client Added</h6>
                      <p className="mb-0 text-muted">
                        John Doe was added as a new client
                      </p>
                      <small className="text-muted">2 hours ago</small>
                    </div>
                  </div>
                </div>
                <div className="activity-item mb-3">
                  <div className="d-flex">
                    <div className="activity-icon bg-success rounded-circle d-flex align-items-center justify-content-center me-3">
                      <i className="fas fa-gavel text-white"></i>
                    </div>
                    <div className="activity-content">
                      <h6 className="mb-1">Case Updated</h6>
                      <p className="mb-0 text-muted">
                        Case #1234 status changed to Active
                      </p>
                      <small className="text-muted">5 hours ago</small>
                    </div>
                  </div>
                </div>
                <div className="activity-item">
                  <div className="d-flex">
                    <div className="activity-icon bg-warning rounded-circle d-flex align-items-center justify-content-center me-3">
                      <i className="fas fa-clock text-white"></i>
                    </div>
                    <div className="activity-content">
                      <h6 className="mb-1">Pending Case</h6>
                      <p className="mb-0 text-muted">
                        Case #5678 requires attention
                      </p>
                      <small className="text-muted">1 day ago</small>
                    </div>
                  </div>
                </div>
              </div>
            </Card.Body>
          </Card>
        </Col>
      </Row>

      <Row className="mt-4">
        <Col md={6}>
          <Card>
            <Card.Header className="d-flex justify-content-between align-items-center">
              <span>Case Status Distribution</span>
              <Button variant="outline-primary" size="sm">
                <i className="fas fa-external-link-alt me-1"></i>
                View All
              </Button>
            </Card.Header>
            <Card.Body>
              <div
                className="chart-placeholder"
                style={{
                  height: "250px",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                }}
              >
                <div className="text-center">
                  <i className="fas fa-chart-pie fa-3x text-muted mb-3"></i>
                  <p className="text-muted">
                    Case status distribution chart would be displayed here
                  </p>
                </div>
              </div>
            </Card.Body>
          </Card>
        </Col>
        <Col md={6}>
          <Card>
            <Card.Header>Upcoming Deadlines</Card.Header>
            <Card.Body>
              <div className="deadline-list">
                <div className="deadline-item mb-3 p-3 border rounded">
                  <div className="d-flex justify-content-between">
                    <h6 className="mb-1">Case Hearing #1234</h6>
                    <span className="badge bg-danger">High</span>
                  </div>
                  <p className="mb-1 text-muted">Civil Law Case</p>
                  <div className="d-flex justify-content-between">
                    <small className="text-muted">
                      <i className="fas fa-calendar me-1"></i>
                      Tomorrow, 10:00 AM
                    </small>
                    <small className="text-muted">
                      <i className="fas fa-user me-1"></i>
                      Assigned to: Jane Smith
                    </small>
                  </div>
                </div>
                <div className="deadline-item mb-3 p-3 border rounded">
                  <div className="d-flex justify-content-between">
                    <h6 className="mb-1">Document Submission</h6>
                    <span className="badge bg-warning">Medium</span>
                  </div>
                  <p className="mb-1 text-muted">Criminal Case #5678</p>
                  <div className="d-flex justify-content-between">
                    <small className="text-muted">
                      <i className="fas fa-calendar me-1"></i>3 days from now
                    </small>
                    <small className="text-muted">
                      <i className="fas fa-user me-1"></i>
                      Assigned to: John Doe
                    </small>
                  </div>
                </div>
              </div>
            </Card.Body>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;
