import React, { useState, useEffect } from "react";
import {
  Row,
  Col,
  Card,
  Table,
  Button,
  Form,
  InputGroup,
  Modal,
  Spinner,
  Dropdown,
  Badge,
  ProgressBar,
} from "react-bootstrap";
import {
  getFiles,
  uploadFile,
  deleteFile,
} from "../services/documentService";
import { FileInfo, UploadFileRequest } from "../types";

const DocumentsPage: React.FC = () => {
  const [documents, setDocuments] = useState<FileInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [showUploadModal, setShowUploadModal] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [search, setSearch] = useState("");
  const [typeFilter, setTypeFilter] = useState("all");
  const [statusFilter, setStatusFilter] = useState("all");
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [formData, setFormData] = useState({
    name: "",
    description: "",
    categoryId: 1,
    tags: "",
  });

  useEffect(() => {
    loadDocuments();
  }, [currentPage, search, typeFilter, statusFilter]);

  const loadDocuments = async () => {
    setLoading(true);
    try {
      const params = {
        page: currentPage,
        pageSize: 10,
        search: search || undefined,
        type: typeFilter !== "all" ? typeFilter : undefined,
        status: statusFilter !== "all" ? statusFilter : undefined,
      };
      const response = await getFiles(params);
      setDocuments(response.data);
      setTotalPages(
        Math.ceil(response.pagination.total / response.pagination.page_size),
      );
    } catch (error) {
      console.error("Failed to load documents", error);
    } finally {
      setLoading(false);
    }
  };

  const handleShowUploadModal = () => {
    setShowUploadModal(true);
    setFormData({
      name: "",
      description: "",
      categoryId: 1,
      tags: "",
    });
    setSelectedFile(null);
    setUploadProgress(0);
  };

  const handleCloseUploadModal = () => {
    setShowUploadModal(false);
    setUploading(false);
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const file = e.target.files[0];
      setSelectedFile(file);
      setFormData((prev) => ({ ...prev, name: file.name }));
    }
  };

  const handleChange = (e: React.ChangeEvent<any>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleUpload = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedFile) return;

    setUploading(true);

    try {
      const data: UploadFileRequest = {
        file: selectedFile,
        category: formData.categoryId.toString(),
        description: formData.description,
      };

      const newDocument = await uploadFile(data);
      setDocuments((prev) => [newDocument, ...prev]);
      handleCloseUploadModal();
    } catch (error) {
      console.error("Failed to upload document", error);
    } finally {
      setUploading(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (
      window.confirm(
        "Are you sure you want to delete this document? This action cannot be undone.",
      )
    ) {
      try {
        await deleteFile(id.toString());
        setDocuments((prev) => prev.filter((d) => d.id !== id.toString()));
      } catch (error) {
        console.error("Failed to delete document", error);
      }
    }
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setCurrentPage(1);
    loadDocuments();
  };

  const getFileIcon = (type: string) => {
    switch (type) {
      case "pdf":
        return <i className="fas fa-file-pdf text-danger"></i>;
      case "doc":
      case "docx":
        return <i className="fas fa-file-word text-primary"></i>;
      case "xls":
      case "xlsx":
        return <i className="fas fa-file-excel text-success"></i>;
      case "ppt":
      case "pptx":
        return <i className="fas fa-file-powerpoint text-warning"></i>;
      case "jpg":
      case "jpeg":
      case "png":
      case "gif":
        return <i className="fas fa-file-image text-info"></i>;
      case "txt":
        return <i className="fas fa-file-alt text-muted"></i>;
      default:
        return <i className="fas fa-file text-muted"></i>;
    }
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  const getTypeBadgeClass = (type: string) => {
    switch (type) {
      case "pdf":
        return "bg-danger";
      case "doc":
      case "docx":
        return "bg-primary";
      case "xls":
      case "xlsx":
        return "bg-success";
      case "ppt":
      case "pptx":
        return "bg-warning";
      case "jpg":
      case "jpeg":
      case "png":
      case "gif":
        return "bg-info";
      case "txt":
        return "bg-secondary";
      default:
        return "bg-secondary";
    }
  };

  const getStatusBadgeClass = (status: string) => {
    switch (status) {
      case "active":
        return "bg-success";
      case "inactive":
        return "bg-secondary";
      default:
        return "bg-secondary";
    }
  };

  // Get type display text
  const getTypeText = (type: string) => {
    switch (type) {
      case "pdf":
        return "PDF";
      case "doc":
        return "Word Document";
      case "docx":
        return "Word Document";
      case "xls":
        return "Excel Spreadsheet";
      case "xlsx":
        return "Excel Spreadsheet";
      case "ppt":
        return "PowerPoint Presentation";
      case "pptx":
        return "PowerPoint Presentation";
      case "jpg":
        return "JPEG Image";
      case "jpeg":
        return "JPEG Image";
      case "png":
        return "PNG Image";
      case "gif":
        return "GIF Image";
      case "txt":
        return "Text File";
      default:
        return type.toUpperCase();
    }
  };

  // Get status display text
  const getStatusText = (status: string) => {
    switch (status) {
      case "active":
        return "Active";
      case "inactive":
        return "Inactive";
      default:
        return status;
    }
  };

  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h1>Documents</h1>
        <Button variant="primary" onClick={handleShowUploadModal}>
          <i className="fas fa-upload me-2"></i>
          Upload Document
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
                    placeholder="Search documents by name, description, or tags..."
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
                    {typeFilter === "all" ? "All" : getTypeText(typeFilter)}
                  </Dropdown.Toggle>
                  <Dropdown.Menu>
                    <Dropdown.Item onClick={() => setTypeFilter("all")}>
                      All
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setTypeFilter("pdf")}>
                      {getTypeText("pdf")}
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setTypeFilter("doc")}>
                      {getTypeText("doc")}
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setTypeFilter("docx")}>
                      {getTypeText("docx")}
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setTypeFilter("xls")}>
                      {getTypeText("xls")}
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setTypeFilter("xlsx")}>
                      {getTypeText("xlsx")}
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
                      : getStatusText(statusFilter)}
                  </Dropdown.Toggle>
                  <Dropdown.Menu>
                    <Dropdown.Item onClick={() => setStatusFilter("all")}>
                      All
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setStatusFilter("active")}>
                      {getStatusText("active")}
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setStatusFilter("inactive")}>
                      {getStatusText("inactive")}
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

      {loading ? (
        <div
          className="d-flex justify-content-center align-items-center"
          style={{ height: "400px" }}
        >
          <Spinner animation="border" />
          <span className="ms-2">Loading documents...</span>
        </div>
      ) : (
        <Card>
          <Card.Body>
            <Table striped bordered hover responsive>
              <thead>
                <tr>
                  <th>Document</th>
                  <th>Type</th>
                  <th>Size</th>
                  <th>Category</th>
                  <th>Status</th>
                  <th>Created</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {documents.map((document) => (
                  <tr key={document.id}>
                    <td>
                      <div className="d-flex align-items-center">
                        <div className="bg-light rounded me-3 p-2">
                          {getFileIcon(document.type || 'unknown')}
                        </div>
                        <div>
                          <div className="fw-bold">{document.name}</div>
                          <div className="small text-muted">
                            {document.description}
                          </div>
                        </div>
                      </div>
                    </td>
                    <td>
                      <Badge bg={getTypeBadgeClass(document.type || 'unknown')}>
                        {getTypeText(document.type || 'unknown')}
                      </Badge>
                    </td>
                    <td>{formatFileSize(document.size)}</td>
                    <td>{document.categoryName}</td>
                    <td>
                      <Badge bg={getStatusBadgeClass(document.status || 'inactive')}>
                        {getStatusText(document.status || 'inactive')}
                      </Badge>
                    </td>
                    <td>
                      {document.created_at ? new Date(document.created_at).toLocaleDateString() : 'N/A'}
                    </td>
                    <td>
                      <div className="d-flex">
                        <Button
                          variant="outline-primary"
                          size="sm"
                          className="me-2"
                          onClick={() => window.open(document.url, "_blank")}
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
                        <Button
                          variant="outline-danger"
                          size="sm"
                          onClick={() => handleDelete(parseInt(document.id))}
                        >
                          <i className="fas fa-trash"></i>
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </Table>

            {documents.length === 0 && (
              <div className="text-center py-5">
                <i className="fas fa-file-alt fa-3x text-muted mb-3"></i>
                <h5>No documents found</h5>
                <p className="text-muted">
                  Try adjusting your search or filter criteria
                </p>
                <Button variant="primary" onClick={handleShowUploadModal}>
                  <i className="fas fa-upload me-2"></i>
                  Upload Your First Document
                </Button>
              </div>
            )}
          </Card.Body>
        </Card>
      )}

      {/* Upload Document Modal */}
      <Modal show={showUploadModal} onHide={handleCloseUploadModal} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>
            <i className="fas fa-upload me-2"></i>
            Upload Document
          </Modal.Title>
        </Modal.Header>
        <Form onSubmit={handleUpload}>
          <Modal.Body>
            {uploading && (
              <div className="mb-3">
                <div className="d-flex justify-content-between">
                  <span>Uploading...</span>
                  <span>{uploadProgress}%</span>
                </div>
                <ProgressBar now={uploadProgress} />
              </div>
            )}

            <Form.Group className="mb-3">
              <Form.Label>
                File <span className="text-danger">*</span>
              </Form.Label>
              <Form.Control
                type="file"
                onChange={handleFileChange}
                required
                disabled={uploading}
              />
            </Form.Group>

            <Form.Group className="mb-3">
              <Form.Label>
                Name <span className="text-danger">*</span>
              </Form.Label>
              <Form.Control
                type="text"
                name="name"
                value={formData.name}
                onChange={handleChange}
                required
                placeholder="Enter document name"
                disabled={uploading}
              />
            </Form.Group>

            <Form.Group className="mb-3">
              <Form.Label>Description</Form.Label>
              <Form.Control
                as="textarea"
                rows={3}
                name="description"
                value={formData.description}
                onChange={handleChange}
                placeholder="Enter document description"
                disabled={uploading}
              />
            </Form.Group>

            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Category</Form.Label>
                  <Form.Select
                    name="categoryId"
                    value={formData.categoryId}
                    onChange={handleChange}
                    disabled={uploading}
                  >
                    <option value={1}>Contracts</option>
                    <option value={2}>Evidence</option>
                    <option value={3}>Letters</option>
                    <option value={4}>Other</option>
                  </Form.Select>
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Tags</Form.Label>
                  <Form.Control
                    type="text"
                    name="tags"
                    value={formData.tags}
                    onChange={handleChange}
                    placeholder="Enter tags (comma separated)"
                    disabled={uploading}
                  />
                </Form.Group>
              </Col>
            </Row>
          </Modal.Body>
          <Modal.Footer>
            <Button
              variant="secondary"
              onClick={handleCloseUploadModal}
              disabled={uploading}
            >
              <i className="fas fa-times me-2"></i>
              Cancel
            </Button>
            <Button
              variant="primary"
              type="submit"
              disabled={uploading || !selectedFile}
            >
              {uploading ? (
                <span>
                  <i className="fas fa-spinner fa-spin me-2"></i>
                  Uploading...
                </span>
              ) : (
                <span>
                  <i className="fas fa-upload me-2"></i>
                  Upload Document
                </span>
              )}
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>
    </div>
  );
};

export default DocumentsPage;
