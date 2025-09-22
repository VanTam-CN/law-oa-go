import React, { useState, useEffect } from 'react';
import { Row, Col, Card,  Tab
import { searchClients, searchCases, searchUsers } from '../services/searchService';
import { Client, Case, UserProfile } from '../types';

const SearchPage: React.FC = () => {
  const [searchTerm, setSearchTerm] = useState('');
  const [searchType, setSearchType] = useState('all');
  const [results, setResults] = useState<{
    clients: Client[];
    cases: Case[];
    users: UserProfile[];
  }>({
    clients: [],
    cases: [],
    users: []
  });
  const [loading, setLoading] = useState(false);
  const [activeTab, setActiveTab] = useState('all');
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  useEffect(() => {
    if (searchTerm.trim() !== '') {
      performSearch();
    } else {
      setResults({ clients: [], cases: [], users: [] });
    }
  }, [searchTerm, searchType, activeTab]);

  const performSearch = async () => {
    setLoading(true);
    try {
      const [clientResults, caseResults, userResults] = await Promise.all([
        searchType === 'all' || searchType === 'clients' ? searchClients({ query: searchTerm, page: currentPage, page_size: 10 }) : Promise.resolve({ data: [], pagination: { total: 0, page: 1, page_size: 10, total_pages: 1 } }),
        searchType === 'all' || searchType === 'cases' ? searchCases({ query: searchTerm, page: currentPage, page_size: 10 }) : Promise.resolve({ data: [], pagination: { total: 0, page: 1, page_size: 10, total_pages: 1 } }),
        searchType === 'all' || searchType === 'users' ? searchUsers({ query: searchTerm, page: currentPage, page_size: 10 }) : Promise.resolve({ data: [], pagination: { total: 0, page: 1, page_size: 10, total_pages: 1 } })
      ]);

      setResults({
        clients: clientResults?.data || [],
        cases: caseResults?.data || [],
        users: userResults?.data || []
      });

      // Calculate total pages based on the largest result set
      const maxTotal = Math.max(
        clientResults?.pagination?.total || 0,
        caseResults?.pagination?.total || 0,
        userResults?.pagination?.total || 0
      );
      setTotalPages(Math.ceil(maxTotal / 10));
    } catch (error) {
      console.error('Search failed', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setCurrentPage(1);
    performSearch();
  };

  const getStatusBadgeClass = (status: string) => {
    switch (status) {
      case 'active': return 'bg-success';
      case 'inactive': return 'bg-secondary';
      case 'pending': return 'bg-warning';
      case 'closed': return 'bg-success';
      case 'suspended': return 'bg-secondary';
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

  const getRoleBadgeClass = (role: string) => {
    switch (role) {
      case 'admin': return 'bg-danger';
      case 'lawyer': return 'bg-primary';
      case 'user': return 'bg-info';
      default: return 'bg-secondary';
    }
  };

  // Get status display text
  const getStatusText = (status: string) => {
    switch (status) {
      case 'active': return 'Active';
      case 'inactive': return 'Inactive';
      case 'pending': return 'Pending';
      case 'closed': return 'Closed';
      case 'suspended': return 'Suspended';
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

  // Get role display text
  const getRoleText = (role: string) => {
    switch (role) {
      case 'admin': return 'Administrator';
      case 'lawyer': return 'Lawyer';
      case 'user': return 'User';
      default: return role;
    }
  };

  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h1>Search</h1>
        <Button variant="outline-primary">
          <i className="fas fa-external-link-alt me-2"></i>
          Advanced Search
        </Button>
      </div>

      <Card className="mb-4">
        <Card.Body>
          <Form onSubmit={handleSearch}>
            <Row>
              <Col md={8}>
                <InputGroup>
                  <Form.Control
                    type="text"
                    placeholder="Search across clients, cases, and users..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                  />
                  <Button variant="outline-secondary" type="submit">
                    <i className="fas fa-search"></i>
                  </Button>
                </InputGroup>
              </Col>
              <Col md={4}>
                <div className="d-flex justify-content-end">
                  <Dropdown className="me-2">
                    <Dropdown.Toggle variant="outline-secondary" id="search-type-dropdown">
                      Search Type: {searchType === 'all' ? 'All' : searchType.charAt(0).toUpperCase() + searchType.slice(1)}
                    </Dropdown.Toggle>
                    <Dropdown.Menu>
                      <Dropdown.Item onClick={() => setSearchType('all')}>All</Dropdown.Item>
                      <Dropdown.Item onClick={() => setSearchType('clients')}>Clients</Dropdown.Item>
                      <Dropdown.Item onClick={() => setSearchType('cases')}>Cases</Dropdown.Item>
                      <Dropdown.Item onClick={() => setSearchType('users')}>Users</Dropdown.Item>
                    </Dropdown.Menu>
                  </Dropdown>
                  <Button variant="outline-primary">
                    <i className="fas fa-filter me-2"></i>
                    Filters
                  </Button>
                </div>
              </Col>
            </Row>
          </Form>
        </Card.Body>
      </Card>

      {searchTerm.trim() === '' ? (
        <Card>
          <Card.Body>
            <div className="text-center py-5">
              <i className="fas fa-search fa-3x text-muted mb-3"></i>
              <h5>Search the Law OA System</h5>
              <p className="text-muted">
                Enter a search term above to find clients, cases, and users across the system
              </p>
              <div className="mt-4">
                <h6 className="mb-3">Search Tips:</h6>
                <ul className="text-start d-inline-block text-muted">
                  <li>Use specific keywords for better results</li>
                  <li>Try searching by name, email, or case title</li>
                  <li>Use filters to narrow down your search</li>
                  <li>Search across all entities or specific types</li>
                </ul>
              </div>
            </div>
          </Card.Body>
        </Card>
      ) : loading ? (
        <div className="d-flex justify-content-center align-items-center" style={{ height: '400px' }}>
          <Spinner animation="border" />
          <span className="ms-2">Searching...</span>
        </div>
      ) : (
        <Card>
          <Card.Body>
            <div className="d-flex justify-content-between align-items-center mb-3">
              <div>
                <h5>Search Results</h5>
                <p className="text-muted mb-0">
                  Found {results.clients.length + results.cases.length + results.users.length} results for "{searchTerm}"
                </p>
              </div>
              <div>
                <Dropdown>
                  <Dropdown.Toggle variant="outline-secondary" id="sort-dropdown">
                    Sort By: Relevance
                  </Dropdown.Toggle>
                  <Dropdown.Menu>
                    <Dropdown.Item>Relevance</Dropdown.Item>
                    <Dropdown.Item>Date Created</Dropdown.Item>
                    <Dropdown.Item>Name</Dropdown.Item>
                  </Dropdown.Menu>
                </Dropdown>
              </div>
            </div>

            <ul className="nav nav-tabs mb-4">
              <li className="nav-item">
                <button
                  className={`nav-link ${activeTab === 'all' ? 'active' : ''}`}
                  onClick={() => setActiveTab('all')}
                >
                  All ({results.clients.length + results.cases.length + results.users.length})
                </button>
              </li>
              <li className="nav-item">
                <button
                  className={`nav-link ${activeTab === 'clients' ? 'active' : ''}`}
                  onClick={() => setActiveTab('clients')}
                >
                  Clients ({results.clients.length})
                </button>
              </li>
              <li className="nav-item">
                <button
                  className={`nav-link ${activeTab === 'cases' ? 'active' : ''}`}
                  onClick={() => setActiveTab('cases')}
                >
                  Cases ({results.cases.length})
                </button>
              </li>
              <li className="nav-item">
                <button
                  className={`nav-link ${activeTab === 'users' ? 'active' : ''}`}
                  onClick={() => setActiveTab('users')}
                >
                  Users ({results.users.length})
                </button>
              </li>
            </ul>

            {(activeTab === 'all' || activeTab === 'clients') && results.clients.length > 0 && (
              <div className="mb-5">
                <h6 className="mb-3">
                  <i className="fas fa-users me-2"></i>
                  Clients
                </h6>
                <Table striped bordered hover responsive>
                  <thead>
                    <tr>
                      <th>Name</th>
                      <th>Email</th>
                      <th>Phone</th>
                      <th>Company</th>
                      <th>Status</th>
                      <th>Created</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {results.clients.map(client => (
                      <tr key={client.id}>
                        <td>
                          <div className="d-flex align-items-center">
                            <div className="bg-light rounded-circle d-flex align-items-center justify-content-center me-2" style={{ width: '32px', height: '32px' }}>
                              <i className="fas fa-user text-muted"></i>
                            </div>
                            <div>
                              <div className="fw-bold">{client.name}</div>
                              <div className="small text-muted">ID: {client.id}</div>
                            </div>
                          </div>
                        </td>
                        <td>{client.email}</td>
                        <td>{client.phone}</td>
                        <td>{client.company}</td>
                        <td>
                          <Badge bg={getStatusBadgeClass(client.status)}>
                            {getStatusText(client.status)}
                          </Badge>
                        </td>
                        <td>{new Date(client.created_at).toLocaleDateString()}</td>
                        <td>
                          <div className="d-flex">
                            <Button
                              variant="outline-primary"
                              size="sm"
                              className="me-2"
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
                            >
                              <i className="fas fa-trash"></i>
                            </Button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </Table>
              </div>
            )}

            {(activeTab === 'all' || activeTab === 'cases') && results.cases.length > 0 && (
              <div className="mb-5">
                <h6 className="mb-3">
                  <i className="fas fa-gavel me-2"></i>
                  Cases
                </h6>
                <Table striped bordered hover responsive>
                  <thead>
                    <tr>
                      <th>Case Title</th>
                      <th>Client</th>
                      <th>Type</th>
                      <th>Priority</th>
                      <th>Status</th>
                      <th>Lawyer</th>
                      <th>Created</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {results.cases.map(caseItem => (
                      <tr key={caseItem.id}>
                        <td>
                          <div>
                            <div className="fw-bold">#{caseItem.id} {caseItem.title}</div>
                            <div className="small text-muted">{caseItem.description.substring(0, 50)}...</div>
                          </div>
                        </td>
                        <td>{caseItem.client_name}</td>
                        <td>
                          <Badge bg="primary">
                            {caseItem.case_type}
                          </Badge>
                        </td>
                        <td>
                          <Badge bg={getPriorityBadgeClass(caseItem.priority)}>
                            {getPriorityText(caseItem.priority)}
                          </Badge>
                        </td>
                        <td>
                          <Badge bg={getStatusBadgeClass(caseItem.status)}>
                            {getStatusText(caseItem.status)}
                          </Badge>
                        </td>
                        <td>{caseItem.lawyer_name || 'Unassigned'}</td>
                        <td>{new Date(caseItem.created_at).toLocaleDateString()}</td>
                        <td>
                          <div className="d-flex">
                            <Button
                              variant="outline-primary"
                              size="sm"
                              className="me-2"
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
                            >
                              <i className="fas fa-trash"></i>
                            </Button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </Table>
              </div>
            )}

            {(activeTab === 'all' || activeTab === 'users') && results.users.length > 0 && (
              <div>
                <h6 className="mb-3">
                  <i className="fas fa-user-shield me-2"></i>
                  Users
                </h6>
                <Table striped bordered hover responsive>
                  <thead>
                    <tr>
                      <th>User</th>
                      <th>Email</th>
                      <th>Role</th>
                      <th>Status</th>
                      <th>Created</th>
                      <th>Last Login</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {results.users.map(user => (
                      <tr key={user.id}>
                        <td>
                          <div className="d-flex align-items-center">
                            <div className="bg-light rounded-circle d-flex align-items-center justify-content-center me-2" style={{ width: '32px', height: '32px' }}>
                              <i className="fas fa-user text-muted"></i>
                            </div>
                            <div>
                              <div className="fw-bold">{user.name}</div>
                              <div className="small text-muted">ID: {user.id}</div>
                            </div>
                          </div>
                        </td>
                        <td>{user.email}</td>
                        <td>
                          <Badge bg={getRoleBadgeClass(user.role)}>
                            {getRoleText(user.role)}
                          </Badge>
                        </td>
                        <td>
                          <Badge bg={getStatusBadgeClass(user.status)}>
                            {getStatusText(user.status)}
                          </Badge>
                        </td>
                        <td>{new Date(user.created_at).toLocaleDateString()}</td>
                        <td>
                          {user.updated_at ? new Date(user.updated_at).toLocaleDateString() : 'Never'}
                        </td>
                        <td>
                          <div className="d-flex">
                            <Button
                              variant="outline-primary"
                              size="sm"
                              className="me-2"
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
                            {user.role !== 'admin' && (
                              <Button
                                variant="outline-danger"
                                size="sm"
                              >
                                <i className="fas fa-trash"></i>
                              </Button>
                            )}
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </Table>
              </div>
            )}

            {results.clients.length === 0 && results.cases.length === 0 && results.users.length === 0 && (
              <div className="text-center py-5">
                <i className="fas fa-search fa-3x text-muted mb-3"></i>
                <h5>No results found</h5>
                <p className="text-muted">
                  We couldn't find any matches for "{searchTerm}". Try different keywords or adjust your filters.
                </p>
                <Button variant="primary" onClick={() => setSearchTerm('')}>
                  <i className="fas fa-sync me-2"></i>
                  Clear Search
                </Button>
              </div>
            )}

            {totalPages > 1 && (
              <div className="d-flex justify-content-center mt-4">
                <Pagination>
                  <Pagination.First onClick={() => setCurrentPage(1)} disabled={currentPage === 1} />
                  <Pagination.Prev onClick={() => setCurrentPage(prev => Math.max(prev - 1, 1))} disabled={currentPage === 1} />
                  {[...Array(totalPages)].map((_, i) => (
                    <Pagination.Item
                      key={i + 1}
                      active={i + 1 === currentPage}
                      onClick={() => setCurrentPage(i + 1)}
                    >
                      {i + 1}
                    </Pagination.Item>
                  ))}
                  <Pagination.Next onClick={() => setCurrentPage(prev => Math.min(prev + 1, totalPages))} disabled={currentPage === totalPages} />
                  <Pagination.Last onClick={() => setCurrentPage(totalPages)} disabled={currentPage === totalPages} />
                </Pagination>
              </div>
            )}
          </Card.Body>
        </Card>
      )}
    </div>
  );
};

export default SearchPage;