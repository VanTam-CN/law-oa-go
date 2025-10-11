import React, { useState } from "react";
import {
  Row,
  Col,
  Card,
  Button,
  Form,
  InputGroup,
  Accordion,
  Badge,
} from "react-bootstrap";

const HelpPage: React.FC = () => {
  const [search, setSearch] = useState("");
  const [activeKey, setActiveKey] = useState("0");

  // FAQ 数据
  const faqs = [
    {
      id: "1",
      category: "getting_started",
      question: "How do I get started with the Law OA System?",
      answer:
        'To get started with the Law OA System, you need to create an account by clicking on the "Register" button on the login page. Once registered, you can log in using your email and password. After logging in, you\'ll be directed to the dashboard where you can start managing your clients and cases.',
    },
    {
      id: "2",
      category: "getting_started",
      question: "What are the system requirements for using the Law OA System?",
      answer:
        "The Law OA System is a web-based application that can be accessed through any modern web browser. We recommend using the latest versions of Chrome, Firefox, Safari, or Edge. The system works on Windows, macOS, and Linux operating systems. For optimal performance, ensure you have a stable internet connection.",
    },
    {
      id: "3",
      category: "clients",
      question: "How do I add a new client to the system?",
      answer:
        'To add a new client, navigate to the "Clients" section from the main menu and click on the "Add Client" button. Fill in the required information such as client name, email, phone number, and address. You can also add additional details like company information and notes. Once you\'ve filled in all the details, click "Save" to add the client to the system.',
    },
    {
      id: "4",
      category: "clients",
      question: "Can I import clients from a CSV file?",
      answer:
        'Yes, you can import clients from a CSV file. Go to the "Clients" section and click on the "Import Clients" button. Select your CSV file and ensure it follows the required format. The system will validate the data and import the clients. You\'ll receive a confirmation message once the import is complete.',
    },
    {
      id: "5",
      category: "cases",
      question: "How do I create a new case?",
      answer:
        'To create a new case, go to the "Cases" section and click on the "Add Case" button. Fill in the case details including title, description, client, case type, priority, and status. You can assign the case to a lawyer and set deadlines. Click "Save" to create the case in the system.',
    },
    {
      id: "6",
      category: "cases",
      question: "How do I assign a case to a lawyer?",
      answer:
        'To assign a case to a lawyer, go to the "Cases" section and find the case you want to assign. Click on the "Edit" button for that case. In the case details form, you\'ll see a dropdown for "Assigned Lawyer". Select the lawyer you want to assign the case to and save the changes.',
    },
    {
      id: "7",
      category: "users",
      question: "How do I add a new user to the system?",
      answer:
        'Only administrators can add new users. If you have admin privileges, go to the "Users" section and click on the "Add User" button. Fill in the user\'s name, email, role, and set a temporary password. The user will be prompted to change their password on first login.',
    },
    {
      id: "8",
      category: "users",
      question: "What are the different user roles in the system?",
      answer:
        "The system has three user roles: Administrator, Lawyer, and User. Administrators have full access to all system features including user management. Lawyers can manage cases and clients but cannot manage users. Users have limited access to view and update their own information.",
    },
    {
      id: "9",
      category: "reports",
      question: "How do I generate reports?",
      answer:
        'To generate reports, go to the "Reports" section and click on the "Generate Report" button. Select the report type, specify the date range, and choose the output format (PDF, Excel, or CSV). Click "Generate" and the system will create the report which you can then download.',
    },
    {
      id: "10",
      category: "reports",
      question: "Can I schedule automatic report generation?",
      answer:
        'Yes, administrators can schedule automatic report generation. In the "Reports" section, click on "Schedule Report". Choose the report type, frequency (daily, weekly, monthly), and specify the recipients who should receive the report via email.',
    },
    {
      id: "11",
      category: "security",
      question: "How secure is my data in the Law OA System?",
      answer:
        "The Law OA System takes data security seriously. All data is encrypted both in transit and at rest. We use industry-standard security protocols including HTTPS, AES-256 encryption, and secure authentication mechanisms. Regular security audits are performed to ensure the highest level of protection for your data.",
    },
    {
      id: "12",
      category: "security",
      question: "What should I do if I forget my password?",
      answer:
        'If you forget your password, click on the "Forgot Password" link on the login page. Enter your email address and you\'ll receive a password reset link. Follow the instructions in the email to set a new password. For security reasons, the reset link expires after 24 hours.',
    },
    {
      id: "13",
      category: "integrations",
      question: "Does the system integrate with other tools?",
      answer:
        "Yes, the Law OA System offers integrations with popular tools including Microsoft Office 365, Google Workspace, and various document management systems. Contact your system administrator for details on available integrations and how to set them up.",
    },
    {
      id: "14",
      category: "integrations",
      question: "Can I sync my calendar with the system?",
      answer:
        'Yes, you can sync your calendar with the Law OA System. The system supports integration with Google Calendar, Outlook, and other calendar applications through standard calendar protocols. Go to "Settings" > "Integrations" to configure calendar syncing.',
    },
    {
      id: "15",
      category: "troubleshooting",
      question: "What should I do if the system is running slowly?",
      answer:
        "If you experience slow performance, try clearing your browser cache and cookies. Ensure you're using a supported browser version. If the issue persists, contact your system administrator who can check server performance and network connectivity.",
    },
    {
      id: "16",
      category: "troubleshooting",
      question: "I'm getting an error message. What should I do?",
      answer:
        "If you encounter an error message, first note down the error code and message. Try refreshing the page or logging out and back in. If the error persists, contact technical support with the error details, steps to reproduce the issue, and your browser information.",
    },
  ];

  // 获取分类显示文本
  const getCategoryText = (category: string) => {
    switch (category) {
      case "getting_started":
        return "Getting Started";
      case "clients":
        return "Clients";
      case "cases":
        return "Cases";
      case "users":
        return "Users";
      case "reports":
        return "Reports";
      case "security":
        return "Security";
      case "integrations":
        return "Integrations";
      case "troubleshooting":
        return "Troubleshooting";
      default:
        return category;
    }
  };

  // 获取分类徽章类
  const getCategoryBadgeClass = (category: string) => {
    switch (category) {
      case "getting_started":
        return "bg-primary";
      case "clients":
        return "bg-success";
      case "cases":
        return "bg-warning";
      case "users":
        return "bg-danger";
      case "reports":
        return "bg-info";
      case "security":
        return "bg-dark";
      case "integrations":
        return "bg-secondary";
      case "troubleshooting":
        return "bg-light text-dark";
      default:
        return "bg-secondary";
    }
  };

  // 过滤FAQ
  const filteredFaqs = faqs.filter(
    (faq) =>
      faq.question.toLowerCase().includes(search.toLowerCase()) ||
      faq.answer.toLowerCase().includes(search.toLowerCase()) ||
      getCategoryText(faq.category)
        .toLowerCase()
        .includes(search.toLowerCase()),
  );

  // 按分类分组FAQ
  const groupedFaqs = filteredFaqs.reduce(
    (acc, faq) => {
      if (!acc[faq.category]) {
        acc[faq.category] = [];
      }
      acc[faq.category].push(faq);
      return acc;
    },
    {} as Record<string, typeof faqs>,
  );

  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h1>Help Center</h1>
        <Button variant="primary">
          <i className="fas fa-headset me-2"></i>
          Contact Support
        </Button>
      </div>

      <Card className="mb-4">
        <Card.Body>
          <div className="text-center mb-4">
            <h2 className="mb-3">How can we help you?</h2>
            <p className="text-muted mb-4">
              Find answers to common questions or contact our support team for
              assistance
            </p>
            <Form onSubmit={(e) => e.preventDefault()}>
              <InputGroup className="w-75 mx-auto">
                <Form.Control
                  type="text"
                  placeholder="Search help articles..."
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                />
                <Button variant="outline-secondary" type="submit">
                  <i className="fas fa-search"></i>
                </Button>
              </InputGroup>
            </Form>
          </div>

          <Row>
            <Col md={3}>
              <div className="d-grid gap-2">
                <Button
                  variant={activeKey === "0" ? "primary" : "outline-secondary"}
                  onClick={() => setActiveKey("0")}
                  className="text-start"
                >
                  <i className="fas fa-star me-2"></i>
                  Popular Articles
                </Button>
                <Button
                  variant={activeKey === "1" ? "primary" : "outline-secondary"}
                  onClick={() => setActiveKey("1")}
                  className="text-start"
                >
                  <i className="fas fa-user-plus me-2"></i>
                  Getting Started
                </Button>
                <Button
                  variant={activeKey === "2" ? "primary" : "outline-secondary"}
                  onClick={() => setActiveKey("2")}
                  className="text-start"
                >
                  <i className="fas fa-users me-2"></i>
                  Clients
                </Button>
                <Button
                  variant={activeKey === "3" ? "primary" : "outline-secondary"}
                  onClick={() => setActiveKey("3")}
                  className="text-start"
                >
                  <i className="fas fa-gavel me-2"></i>
                  Cases
                </Button>
                <Button
                  variant={activeKey === "4" ? "primary" : "outline-secondary"}
                  onClick={() => setActiveKey("4")}
                  className="text-start"
                >
                  <i className="fas fa-user-shield me-2"></i>
                  Users
                </Button>
                <Button
                  variant={activeKey === "5" ? "primary" : "outline-secondary"}
                  onClick={() => setActiveKey("5")}
                  className="text-start"
                >
                  <i className="fas fa-chart-bar me-2"></i>
                  Reports
                </Button>
                <Button
                  variant={activeKey === "6" ? "primary" : "outline-secondary"}
                  onClick={() => setActiveKey("6")}
                  className="text-start"
                >
                  <i className="fas fa-shield-alt me-2"></i>
                  Security
                </Button>
                <Button
                  variant={activeKey === "7" ? "primary" : "outline-secondary"}
                  onClick={() => setActiveKey("7")}
                  className="text-start"
                >
                  <i className="fas fa-plug me-2"></i>
                  Integrations
                </Button>
                <Button
                  variant={activeKey === "8" ? "primary" : "outline-secondary"}
                  onClick={() => setActiveKey("8")}
                  className="text-start"
                >
                  <i className="fas fa-tools me-2"></i>
                  Troubleshooting
                </Button>
              </div>
            </Col>
            <Col md={9}>
              <Accordion
                activeKey={activeKey}
                onSelect={(k) =>
                  setActiveKey(typeof k === "string" ? k : k?.[0] || "0")
                }
              >
                {Object.entries(groupedFaqs).map(
                  ([category, categoryFaqs], index) => (
                    <Accordion.Item eventKey={index.toString()} key={category}>
                      <Accordion.Header>
                        <span>
                          <Badge
                            variant={getCategoryBadgeClass(category)}
                            className="me-2"
                          >
                            {getCategoryText(category)}
                          </Badge>
                          {getCategoryText(category)} ({categoryFaqs.length})
                        </span>
                      </Accordion.Header>
                      <Accordion.Body>
                        <div className="faq-list">
                          {categoryFaqs.map((faq) => (
                            <Card key={faq.id} className="mb-3">
                              <Card.Header
                                className="faq-question fw-bold cursor-pointer"
                                onClick={() => {
                                  const element = document.getElementById(
                                    `faq-${faq.id}`,
                                  );
                                  if (element) {
                                    element.classList.toggle("d-none");
                                  }
                                }}
                              >
                                <i className="fas fa-question-circle me-2"></i>
                                {faq.question}
                              </Card.Header>
                              <Card.Body
                                id={`faq-${faq.id}`}
                                className="faq-answer"
                              >
                                <p>{faq.answer}</p>
                                <div className="d-flex justify-content-between align-items-center mt-3">
                                  <div>
                                    <small className="text-muted">
                                      <i className="fas fa-tags me-1"></i>
                                      <Badge
                                        variant={getCategoryBadgeClass(category)}
                                        className="me-1"
                                      >
                                        {getCategoryText(category)}
                                      </Badge>
                                    </small>
                                  </div>
                                  <div>
                                    <Button
                                      variant="outline-primary"
                                      size="sm"
                                      className="me-2"
                                    >
                                      <i className="fas fa-thumbs-up me-1"></i>
                                      Helpful
                                    </Button>
                                    <Button
                                      variant="outline-secondary"
                                      size="sm"
                                    >
                                      <i className="fas fa-flag me-1"></i>
                                      Report Issue
                                    </Button>
                                  </div>
                                </div>
                              </Card.Body>
                            </Card>
                          ))}
                        </div>
                      </Accordion.Body>
                    </Accordion.Item>
                  ),
                )}
              </Accordion>

              {filteredFaqs.length === 0 && (
                <div className="text-center py-5">
                  <i className="fas fa-search fa-3x text-muted mb-3"></i>
                  <h5>No articles found</h5>
                  <p className="text-muted">
                    We couldn't find any help articles matching your search. Try
                    different keywords or browse categories.
                  </p>
                  <Button variant="primary" onClick={() => setSearch("")}>
                    <i className="fas fa-sync me-2"></i>
                    Reset Search
                  </Button>
                </div>
              )}
            </Col>
          </Row>
        </Card.Body>
      </Card>

      <Card>
        <Card.Header>
          <i className="fas fa-life-ring me-2"></i>
          Need More Help?
        </Card.Header>
        <Card.Body>
          <Row>
            <Col md={4} className="mb-3">
              <Card className="h-100">
                <Card.Body className="text-center">
                  <div
                    className="bg-primary rounded-circle d-flex align-items-center justify-content-center mx-auto mb-3"
                    style={{ width: "60px", height: "60px" }}
                  >
                    <i className="fas fa-book text-white fa-2x"></i>
                  </div>
                  <h5>Documentation</h5>
                  <p className="text-muted">
                    Comprehensive guides and tutorials for all system features
                  </p>
                  <Button variant="outline-primary">
                    <i className="fas fa-external-link-alt me-2"></i>
                    View Docs
                  </Button>
                </Card.Body>
              </Card>
            </Col>
            <Col md={4} className="mb-3">
              <Card className="h-100">
                <Card.Body className="text-center">
                  <div
                    className="bg-success rounded-circle d-flex align-items-center justify-content-center mx-auto mb-3"
                    style={{ width: "60px", height: "60px" }}
                  >
                    <i className="fas fa-video text-white fa-2x"></i>
                  </div>
                  <h5>Video Tutorials</h5>
                  <p className="text-muted">
                    Step-by-step video guides for common tasks and workflows
                  </p>
                  <Button variant="outline-success">
                    <i className="fas fa-play-circle me-2"></i>
                    Watch Videos
                  </Button>
                </Card.Body>
              </Card>
            </Col>
            <Col md={4} className="mb-3">
              <Card className="h-100">
                <Card.Body className="text-center">
                  <div
                    className="bg-warning rounded-circle d-flex align-items-center justify-content-center mx-auto mb-3"
                    style={{ width: "60px", height: "60px" }}
                  >
                    <i className="fas fa-comments text-white fa-2x"></i>
                  </div>
                  <h5>Live Chat</h5>
                  <p className="text-muted">
                    Get instant help from our support team during business hours
                  </p>
                  <Button variant="outline-warning">
                    <i className="fas fa-comment-dots me-2"></i>
                    Start Chat
                  </Button>
                </Card.Body>
              </Card>
            </Col>
          </Row>
        </Card.Body>
      </Card>
    </div>
  );
};

export default HelpPage;
