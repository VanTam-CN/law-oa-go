import React, { useState, useEffect } from "react";
import {
  Row,
  Col,
  Card,
  Button,
  Form,
  InputGroup,
  Spinner,
  Dropdown,
  Badge,
  ProgressBar,
  Table,
} from "react-bootstrap";
import { getReports, generateReport } from "../services/reportService";
import { Report, CreateReportRequest } from "../types";

const ReportsPage: React.FC = () => {
  const [reports, setReports] = useState<Report[]>([]);
  const [loading, setLoading] = useState(true);
  const [generating, setGenerating] = useState(false);
  const [search, setSearch] = useState("");
  const [typeFilter, setTypeFilter] = useState("all");
  const [statusFilter, setStatusFilter] = useState("all");
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [formData, setFormData] = useState({
    title: "",
    description: "",
    type: "client_summary",
    startDate: "",
    endDate: "",
    format: "pdf",
    parameters: {} as Record<string, any>,
  });

  useEffect(() => {
    loadReports();
  }, [currentPage, search, typeFilter, statusFilter]);

  const loadReports = async () => {
    setLoading(true);
    try {
      const params = {
        page: currentPage,
        pageSize: 10,
        search: search || undefined,
        type: typeFilter !== "all" ? typeFilter : undefined,
        status: statusFilter !== "all" ? statusFilter : undefined,
      };
      const response = await getReports(params);
      setReports(response.data);
      setTotalPages(
        Math.ceil(response.pagination.total / response.pagination.page_size),
      );
    } catch (error) {
      console.error("Failed to load reports", error);
    } finally {
      setLoading(false);
    }
  };

  const handleShowGenerateModal = () => {
    setFormData({
      title: "",
      description: "",
      type: "client_summary",
      startDate: "",
      endDate: "",
      format: "pdf",
      parameters: {} as Record<string, any>,
    });
    setGenerating(false);
  };

  const handleGenerateReport = async (e: React.FormEvent) => {
    e.preventDefault();
    setGenerating(true);
    try {
      const data: CreateReportRequest = {
        title: formData.title,
        description: formData.description,
        type: formData.type,
        parameters: {
          start_date: formData.startDate,
          end_date: formData.endDate,
        },
        format: formData.format as "pdf" | "excel" | "csv",
      };
      const newReport = await generateReport(data);
      setReports((prev) => [newReport, ...prev]);
      handleShowGenerateModal();
    } catch (error) {
      console.error("Failed to generate report", error);
    } finally {
      setGenerating(false);
    }
  };

  const handleDownloadReport = async (id: number, format: string) => {
    try {
      // 这里应该调用下载报告的服务
      console.log(`Downloading report ${id} in ${format} format`);
    } catch (error) {
      console.error("Failed to download report", error);
    }
  };

  const handleChange = (e: React.ChangeEvent<any>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setCurrentPage(1);
    loadReports();
  };

  const getReportTypeBadgeClass = (type: string) => {
    switch (type) {
      case "client_summary":
        return "bg-primary";
      case "case_summary":
        return "bg-success";
      case "financial":
        return "bg-warning";
      case "performance":
        return "bg-info";
      case "activity":
        return "bg-secondary";
      default:
        return "bg-secondary";
    }
  };

  const getReportStatusBadgeClass = (status: string) => {
    switch (status) {
      case "generated":
        return "bg-success";
      case "pending":
        return "bg-warning";
      case "failed":
        return "bg-danger";
      default:
        return "bg-secondary";
    }
  };

  // Get report type display text
  const getReportTypeText = (type: string) => {
    switch (type) {
      case "client_summary":
        return "Client Summary";
      case "case_summary":
        return "Case Summary";
      case "financial":
        return "Financial Report";
      case "performance":
        return "Performance Report";
      case "activity":
        return "Activity Report";
      default:
        return type;
    }
  };

  // Get report status display text
  const getReportStatusText = (status: string) => {
    switch (status) {
      case "generated":
        return "Generated";
      case "pending":
        return "Pending";
      case "failed":
        return "Failed";
      default:
        return status;
    }
  };

  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h1>Reports</h1>
        <Button variant="primary" onClick={handleShowGenerateModal}>
          <i className="fas fa-plus me-2"></i>
          Generate Report
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
                    placeholder="Search reports by title or description..."
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
                  <Dropdown.Toggle
                    variant="outline-secondary"
                    id="type-dropdown"
                  >
                    Type:{" "}
                    {typeFilter === "all"
                      ? "All"
                      : getReportTypeText(typeFilter)}
                  </Dropdown.Toggle>
                  <Dropdown.Menu>
                    <Dropdown.Item onClick={() => setTypeFilter("all")}>
                      All
                    </Dropdown.Item>
                    <Dropdown.Item
                      onClick={() => setTypeFilter("client_summary")}
                    >
                      {getReportTypeText("client_summary")}
                    </Dropdown.Item>
                    <Dropdown.Item
                      onClick={() => setTypeFilter("case_summary")}
                    >
                      {getReportTypeText("case_summary")}
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setTypeFilter("financial")}>
                      {getReportTypeText("financial")}
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setTypeFilter("performance")}>
                      {getReportTypeText("performance")}
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setTypeFilter("activity")}>
                      {getReportTypeText("activity")}
                    </Dropdown.Item>
                  </Dropdown.Menu>
                </Dropdown>
                <Dropdown className="me-2">
                  <Dropdown.Toggle
                    variant="outline-secondary"
                    id="status-dropdown"
                  >
                    Status:{" "}
                    {statusFilter === "all"
                      ? "All"
                      : getReportStatusText(statusFilter)}
                  </Dropdown.Toggle>
                  <Dropdown.Menu>
                    <Dropdown.Item onClick={() => setStatusFilter("all")}>
                      All
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setStatusFilter("generated")}>
                      {getReportStatusText("generated")}
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setStatusFilter("pending")}>
                      {getReportStatusText("pending")}
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setStatusFilter("failed")}>
                      {getReportStatusText("failed")}
                    </Dropdown.Item>
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

      <Card className="mb-4">
        <Card.Header className="d-flex justify-content-between align-items-center">
          <span>
            <i className="fas fa-file-contract me-2"></i> Generate New Report
          </span>
          <Button variant="outline-primary" size="sm">
            <i className="fas fa-history me-2"></i>
            View Templates
          </Button>
        </Card.Header>
        <Card.Body>
          <Form onSubmit={handleGenerateReport}>
            <Row>
              <Col md={6}>
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
                    placeholder="Enter report title"
                  />
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>
                    Report Type <span className="text-danger">*</span>
                  </Form.Label>
                  <Form.Select
                    name="type"
                    value={formData.type}
                    onChange={handleChange}
                    required
                  >
                    <option value="client_summary">
                      {getReportTypeText("client_summary")}
                    </option>
                    <option value="case_summary">
                      {getReportTypeText("case_summary")}
                    </option>
                    <option value="financial">
                      {getReportTypeText("financial")}
                    </option>
                    <option value="performance">
                      {getReportTypeText("performance")}
                    </option>
                    <option value="activity">
                      {getReportTypeText("activity")}
                    </option>
                  </Form.Select>
                </Form.Group>
              </Col>
            </Row>
            <Row>
              <Col md={12}>
                <Form.Group className="mb-3">
                  <Form.Label>Description</Form.Label>
                  <Form.Control
                    as="textarea"
                    rows={2}
                    name="description"
                    value={formData.description}
                    onChange={handleChange}
                    placeholder="Enter report description"
                  />
                </Form.Group>
              </Col>
            </Row>
            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Start Date</Form.Label>
                  <Form.Control
                    type="date"
                    name="startDate"
                    value={formData.startDate}
                    onChange={handleChange}
                  />
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>End Date</Form.Label>
                  <Form.Control
                    type="date"
                    name="endDate"
                    value={formData.endDate}
                    onChange={handleChange}
                  />
                </Form.Group>
              </Col>
            </Row>
            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Format</Form.Label>
                  <Form.Select
                    name="format"
                    value={formData.format}
                    onChange={handleChange}
                  >
                    <option value="pdf">PDF</option>
                    <option value="excel">Excel</option>
                    <option value="csv">CSV</option>
                    <option value="html">HTML</option>
                  </Form.Select>
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>&nbsp;</Form.Label>
                  <div>
                    <Button
                      variant="primary"
                      type="submit"
                      disabled={generating}
                      className="w-100"
                    >
                      {generating ? (
                        <span>
                          <i className="fas fa-spinner fa-spin me-2"></i>
                          Generating Report...
                        </span>
                      ) : (
                        <span>
                          <i className="fas fa-file-contract me-2"></i>
                          Generate Report
                        </span>
                      )}
                    </Button>
                  </div>
                </Form.Group>
              </Col>
            </Row>
          </Form>
        </Card.Body>
      </Card>

      {loading ? (
        <div
          className="d-flex justify-content-center align-items-center"
          style={{ height: "400px" }}
        >
          <Spinner animation="border" />
          <span className="ms-2">Loading reports...</span>
        </div>
      ) : (
        <Card>
          <Card.Body>
            <Table striped bordered hover responsive>
              <thead>
                <tr>
                  <th>Report</th>
                  <th>Type</th>
                  <th>Period</th>
                  <th>Status</th>
                  <th>Generated</th>
                  <th>Size</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {reports.map((report) => (
                  <tr key={report.id}>
                    <td>
                      <div>
                        <div className="fw-bold">{report.title}</div>
                        <div className="small text-muted">
                          {report.description}
                        </div>
                      </div>
                    </td>
                    <td>
                      <Badge bg={getReportTypeBadgeClass(report.type)}>
                        {getReportTypeText(report.type)}
                      </Badge>
                    </td>
                    <td>
                      <div>
                        <div>
                          {report.startDate
                            ? new Date(report.startDate).toLocaleDateString()
                            : "N/A"}
                        </div>
                        <div className="small text-muted">
                          to{" "}
                          {report.endDate
                            ? new Date(report.endDate).toLocaleDateString()
                            : "N/A"}
                        </div>
                      </div>
                    </td>
                    <td>
                      <Badge bg={getReportStatusBadgeClass(report.status)}>
                        {getReportStatusText(report.status)}
                      </Badge>
                    </td>
                    <td>
                      {report.created_at
                        ? new Date(report.created_at).toLocaleDateString()
                        : "N/A"}
                    </td>
                    <td>
                      {report.size
                        ? `${(report.size / 1024 / 1024).toFixed(2)} MB`
                        : "N/A"}
                    </td>
                    <td>
                      <div className="d-flex">
                        <Button
                          variant="outline-primary"
                          size="sm"
                          className="me-2"
                          onClick={() =>
                            handleDownloadReport(report.id, report.format)
                          }
                          disabled={report.status !== "generated"}
                        >
                          <i className="fas fa-download"></i>
                        </Button>
                        <Button
                          variant="outline-info"
                          size="sm"
                          className="me-2"
                        >
                          <i className="fas fa-eye"></i>
                        </Button>
                        <Button variant="outline-danger" size="sm">
                          <i className="fas fa-trash"></i>
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </Table>

            {reports.length === 0 && (
              <div className="text-center py-5">
                <i className="fas fa-file-contract fa-3x text-muted mb-3"></i>
                <h5>No reports found</h5>
                <p className="text-muted">
                  Try adjusting your search or filter criteria
                </p>
                <Button variant="primary" onClick={handleShowGenerateModal}>
                  <i className="fas fa-plus me-2"></i>
                  Generate Your First Report
                </Button>
              </div>
            )}
          </Card.Body>
        </Card>
      )}

      {/* Report Generation Progress */}
      {generating && (
        <Card className="mt-4">
          <Card.Header>
            <i className="fas fa-spinner fa-spin me-2"></i>
            Report Generation in Progress
          </Card.Header>
          <Card.Body>
            <div className="d-flex align-items-center">
              <div className="flex-grow-1 me-3">
                <ProgressBar animated now={75} />
              </div>
              <div className="text-muted">75%</div>
            </div>
            <div className="mt-2 text-muted">
              <i className="fas fa-info-circle me-2"></i>
              Generating report: {formData.title}
            </div>
          </Card.Body>
        </Card>
      )}
    </div>
  );
};

export default ReportsPage;
