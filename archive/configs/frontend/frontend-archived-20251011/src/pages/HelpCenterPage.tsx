import React, { useState } from 'react';
import { Row, Col, Card, Button, Form, InputGroup, Accordion, Badge, Spinner } from 'react-bootstrap',  Tab

const HelpCenterPage: React.FC = () => {
  const [search, setSearch] = useState('');
  const [activeCategory, setActiveCategory] = useState('getting_started');
  const [loading, setLoading] = useState(false);

  // 帮助文章数据
  const helpArticles = [
    {
      id: '1',
      category: 'getting_started',
      title: 'Getting Started with Law OA System',
      content: 'Learn how to set up your account and start using the Law OA System.',
      tags: ['beginner', 'setup', 'account']
    },
    {
      id: '2',
      category: 'getting_started',
      title: 'Navigating the Dashboard',
      content: 'Understand how to use the dashboard to monitor your cases and clients.',
      tags: ['dashboard', 'navigation', 'overview']
    },
    {
      id: '3',
      category: 'clients',
      title: 'Managing Clients',
      content: 'How to add, edit, and delete client information in the system.',
      tags: ['clients', 'crm', 'management']
    },
    {
      id: '4',
      category: 'clients',
      title: 'Client Search and Filtering',
      content: 'Learn how to search and filter clients using advanced search options.',
      tags: ['clients', 'search', 'filter']
    },
    {
      id: '5',
      category: 'cases',
      title: 'Creating and Managing Cases',
      content: 'Step-by-step guide to creating and managing legal cases in the system.',
      tags: ['cases', 'legal', 'management']
    },
    {
      id: '6',
      category: 'cases',
      title: 'Assigning Lawyers to Cases',
      content: 'How to assign lawyers to cases and track their progress.',
      tags: ['cases', 'lawyers', 'assignment']
    },
    {
      id: '7',
      category: 'cases',
      title: 'Tracking Case Deadlines',
      content: 'Learn how to set and track important case deadlines and hearings.',
      tags: ['cases', 'deadlines', 'tracking']
    },
    {
      id: '8',
      category: 'documents',
      title: 'Document Management',
      content: 'How to upload, organize, and share documents with your team.',
      tags: ['documents', 'files', 'sharing']
    },
    {
      id: '9',
      category: 'documents',
      title: 'Document Templates',
      content: 'Using and customizing document templates for common legal documents.',
      tags: ['documents', 'templates', 'customization']
    },
    {
      id: '10',
      category: 'reports',
      title: 'Generating Reports',
      content: 'How to generate and export various reports from the system.',
      tags: ['reports', 'analytics', 'export']
    },
    {
      id: '11',
      category: 'reports',
      title: 'Understanding Analytics',
      content: 'Interpreting the analytics and insights provided by the system.',
      tags: ['reports', 'analytics', 'insights']
    },
    {
      id: '12',
      category: 'settings',
      title: 'Account Settings',
      content: 'Managing your account settings, preferences, and security options.',
      tags: ['settings', 'account', 'security']
    },
    {
      id: '13',
      category: 'settings',
      title: 'Notification Preferences',
      content: 'Configuring how and when you receive notifications from the system.',
      tags: ['settings', 'notifications', 'preferences']
    },
    {
      id: '14',
      category: 'troubleshooting',
      title: 'Common Issues and Solutions',
      content: 'Solutions to common problems users encounter with the system.',
      tags: ['troubleshooting', 'issues', 'solutions']
    },
    {
      id: '15',
      category: 'troubleshooting',
      title: 'Contacting Support',
      content: 'How to get help from our support team when you need it.',
      tags: ['troubleshooting', 'support', 'contact']
    }
  ];

  // 类别数据
  const categories = [
    { id: 'getting_started', name: 'Getting Started', icon: 'fas fa-rocket', count: 2 },
    { id: 'clients', name: 'Clients', icon: 'fas fa-users', count: 2 },
    { id: 'cases', name: 'Cases', icon: 'fas fa-gavel', count: 3 },
    { id: 'documents', name: 'Documents', icon: 'fas fa-file-contract', count: 2 },
    { id: 'reports', name: 'Reports', icon: 'fas fa-chart-bar', count: 2 },
    { id: 'settings', name: 'Settings', icon: 'fas fa-cog', count: 2 },
    { id: 'troubleshooting', name: 'Troubleshooting', icon: 'fas fa-exclamation-triangle', count: 2 }
  ];

  // 过滤文章
  const filteredArticles = helpArticles.filter(article => 
    (activeCategory === 'all' || article.category === activeCategory) &&
    (article.title.toLowerCase().includes(search.toLowerCase()) ||
     article.content.toLowerCase().includes(search.toLowerCase()) ||
     article.tags.some(tag => tag.toLowerCase().includes(search.toLowerCase())))
  );

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    // 模拟搜索延迟
    setTimeout(() => {
      setLoading(false);
    }, 500);
  };

  const getCategoryIcon = (category: string) => {
    const cat = categories.find(c => c.id === category);
    return cat ? cat.icon : 'fas fa-question-circle';
  };

  const getCategoryName = (category: string) => {
    const cat = categories.find(c => c.id === category);
    return cat ? cat.name : category;
  };

  const getCategoryBadgeClass = (category: string) => {
    switch (category) {
      case 'getting_started': return 'bg-primary';
      case 'clients': return 'bg-success';
      case 'cases': return 'bg-danger';
      case 'documents': return 'bg-info';
      case 'reports': return 'bg-warning';
      case 'settings': return 'bg-secondary';
      case 'troubleshooting': return 'bg-dark';
      default: return 'bg-secondary';
    }
  };

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
              Find answers to common questions or contact our support team for assistance
            </p>
            <Form onSubmit={handleSearch}>
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
                  variant={activeCategory === 'all' ? 'primary' : 'outline-secondary'} 
                  onClick={() => setActiveCategory('all')}
                  className="text-start"
                >
                  <i className="fas fa-star me-2"></i>
                  Popular Articles
                </Button>
                {categories.map(category => (
                  <Button 
                    key={category.id}
                    variant={activeCategory === category.id ? 'primary' : 'outline-secondary'} 
                    onClick={() => setActiveCategory(category.id)}
                    className="text-start d-flex justify-content-between align-items-center"
                  >
                    <span>
                      <i className={`${category.icon} me-2`}></i>
                      {category.name}
                    </span>
                    <Badge bg={activeCategory === category.id ? 'light' : 'secondary'} className="text-dark">
                      {category.count}
                    </Badge>
                  </Button>
                ))}
              </div>
            </Col>
            <Col md={9}>
              <Accordion activeKey={activeCategory}>
                {filteredArticles.length > 0 ? (
                  filteredArticles.map((article, index) => (
                    <Accordion.Item eventKey={`${index}`} key={article.id}>
                      <Accordion.Header>
                        <span>
                          <i className={`${getCategoryIcon(article.category)} me-2`}></i>
                          {article.title}
                        </span>
                      </Accordion.Header>
                      <Accordion.Body>
                        <p>{article.content}</p>
                        <div className="d-flex justify-content-between align-items-center mt-3">
                          <div>
                            <small className="text-muted">
                              <i className="fas fa-tags me-1"></i>
                              {article.tags.map(tag => (
                                <Badge 
                                  key={tag} 
                                  variant={getCategoryBadgeClass(article.category)} 
                                  className="me-1"
                                >
                                  {tag}
                                </Badge>
                              ))}
                            </small>
                          </div>
                          <div>
                            <Button variant="outline-primary" size="sm" className="me-2">
                              <i className="fas fa-thumbs-up me-1"></i>
                              Helpful
                            </Button>
                            <Button variant="outline-secondary" size="sm">
                              <i className="fas fa-flag me-1"></i>
                              Report Issue
                            </Button>
                          </div>
                        </div>
                      </Accordion.Body>
                    </Accordion.Item>
                  ))
                ) : (
                  <div className="text-center py-5">
                    <i className="fas fa-search fa-3x text-muted mb-3"></i>
                    <h5>No articles found</h5>
                    <p className="text-muted">
                      We couldn't find any help articles matching your search. Try different keywords or browse categories.
                    </p>
                    <Button variant="primary" onClick={() => setSearch('')}>
                      <i className="fas fa-sync me-2"></i>
                      Reset Search
                    </Button>
                  </div>
                )}
              </Accordion>
              
              {filteredArticles.length > 0 && (
                <div className="mt-4">
                  <h5 className="mb-3">Need More Help?</h5>
                  <Row>
                    <Col md={4} className="mb-3">
                      <Card className="h-100">
                        <Card.Body className="text-center">
                          <div className="bg-primary rounded-circle d-flex align-items-center justify-content-center mx-auto mb-3" style={{ width: '60px', height: '60px' }}>
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
                          <div className="bg-success rounded-circle d-flex align-items-center justify-content-center mx-auto mb-3" style={{ width: '60px', height: '60px' }}>
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
                          <div className="bg-warning rounded-circle d-flex align-items-center justify-content-center mx-auto mb-3" style={{ width: '60px', height: '60px' }}>
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
                </div>
              )}
            </Col>
          </Row>
        </Card.Body>
      </Card>
    </div>
  );
};

export default HelpCenterPage;