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
} from "react-bootstrap";
import { getFileList, uploadFile, deleteFile } from "../services/fileService";
import { FileInfo, FileListRequest } from "../types";

const FilesPage: React.FC = () => {
  const [files, setFiles] = useState<FileInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [showUploadModal, setShowUploadModal] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [uploading, setUploading] = useState(false);
  const [search, setSearch] = useState("");
  const [typeFilter, setTypeFilter] = useState("all");
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  useEffect(() => {
    loadFiles();
  }, [currentPage, search, typeFilter]);

  const loadFiles = async () => {
    setLoading(true);
    try {
      const params: FileListRequest = {
        page: currentPage,
        page_size: 10,
        search: search || undefined,
        file_type: typeFilter !== "all" ? typeFilter : undefined,
      };
      const response = await getFileList(params);
      setFiles(response.data);
      setTotalPages(
        Math.ceil(response.pagination.total / response.pagination.page_size),
      );
    } catch (error) {
      console.error("Failed to load files", error);
    } finally {
      setLoading(false);
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      setSelectedFile(e.target.files[0]);
    }
  };

  const handleUpload = async () => {
    if (!selectedFile) return;

    setUploading(true);
    try {
      const response = await uploadFile(selectedFile, "document");
      setFiles((prev) => [response, ...prev]);
      handleCloseUploadModal();
    } catch (error) {
      console.error("Failed to upload file", error);
    } finally {
      setUploading(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (
      window.confirm(
        "Are you sure you want to delete this file? This action cannot be undone.",
      )
    ) {
      try {
        await deleteFile(id.toString());
        setFiles((prev) => prev.filter((f) => f.id !== id.toString()));
      } catch (error) {
        console.error("Failed to delete file", error);
      }
    }
  };

  const handleDownload = async (id: string, filename: string) => {
    try {
      // In a real implementation, this would download the file
      console.log(`Downloading file ${filename} with ID ${id}`);
      alert(`File download would start for ${filename}`);
    } catch (error) {
      console.error("Failed to download file", error);
    }
  };

  const handleCloseUploadModal = () => {
    setShowUploadModal(false);
    setSelectedFile(null);
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setCurrentPage(1);
    loadFiles();
  };

  const getFileIcon = (fileType: string) => {
    switch (fileType.toLowerCase()) {
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
      case "zip":
      case "rar":
        return <i className="fas fa-file-archive text-secondary"></i>;
      default:
        return <i className="fas fa-file text-muted"></i>;
    }
  };

  const getFileTypeBadgeClass = (fileType: string) => {
    switch (fileType.toLowerCase()) {
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
      case "zip":
      case "rar":
        return "bg-dark";
      default:
        return "bg-muted";
    }
  };

  // Get file type display text
  const getFileTypeText = (fileType: string) => {
    switch (fileType.toLowerCase()) {
      case "pdf":
        return "PDF Document";
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
      case "zip":
        return "ZIP Archive";
      case "rar":
        return "RAR Archive";
      default:
        return fileType.toUpperCase();
    }
  };

  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h1>Files</h1>
        <Button variant="primary" onClick={() => setShowUploadModal(true)}>
          <i className="fas fa-upload me-2"></i>
          Upload File
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
                    placeholder="Search files by name or type..."
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
                    {typeFilter === "all" ? "All" : getFileTypeText(typeFilter)}
                  </Dropdown.Toggle>
                  <Dropdown.Menu>
                    <Dropdown.Item onClick={() => setTypeFilter("all")}>
                      All
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setTypeFilter("pdf")}>
                      {getFileTypeText("pdf")}
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setTypeFilter("doc")}>
                      {getFileTypeText("doc")}
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setTypeFilter("docx")}>
                      {getFileTypeText("docx")}
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setTypeFilter("xls")}>
                      {getFileTypeText("xls")}
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setTypeFilter("xlsx")}>
                      {getFileTypeText("xlsx")}
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setTypeFilter("ppt")}>
                      {getFileTypeText("ppt")}
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setTypeFilter("pptx")}>
                      {getFileTypeText("pptx")}
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setTypeFilter("jpg")}>
                      {getFileTypeText("jpg")}
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setTypeFilter("png")}>
                      {getFileTypeText("png")}
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setTypeFilter("txt")}>
                      {getFileTypeText("txt")}
                    </Dropdown.Item>
                    <Dropdown.Item onClick={() => setTypeFilter("zip")}>
                      {getFileTypeText("zip")}
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
          <span className="ms-2">Loading files...</span>
        </div>
      ) : (
        <Card>
          <Card.Body>
            <Table striped bordered hover responsive>
              <thead>
                <tr>
                  <th>File</th>
                  <th>Type</th>
                  <th>Size</th>
                  <th>Uploaded</th>
                  <th>Owner</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {files.map((file) => (
                  <tr key={file.id}>
                    <td>
                      <div className="d-flex align-items-center">
                        <div
                          className="bg-light rounded-circle d-flex align-items-center justify-content-center me-3"
                          style={{ width: "32px", height: "32px" }}
                        >
                          {getFileIcon(file.type || 'unknown')}
                        </div>
                        <div>
                          <div className="fw-bold">{file.name}</div>
                          <div className="small text-muted">ID: {file.id}</div>
                        </div>
                      </div>
                    </td>
                    <td>
                      <Badge bg={getFileTypeBadgeClass(file.type || 'unknown')}>
                        {getFileTypeText(file.type || 'unknown')}
                      </Badge>
                    </td>
                    <td>{file.size} KB</td>
                    <td>{new Date(file.uploadTime).toLocaleDateString()}</td>
                    <td>System</td>
                    <td>
                      <div className="d-flex">
                        <Button
                          variant="outline-primary"
                          size="sm"
                          className="me-2"
                          onClick={() => handleDownload(file.id, file.name)}
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
                          onClick={() => handleDelete(parseInt(file.id))}
                        >
                          <i className="fas fa-trash"></i>
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </Table>

            {files.length === 0 && (
              <div className="text-center py-5">
                <i className="fas fa-file fa-3x text-muted mb-3"></i>
                <h5>No files found</h5>
                <p className="text-muted">
                  Try adjusting your search or filter criteria
                </p>
                <Button
                  variant="primary"
                  onClick={() => setShowUploadModal(true)}
                >
                  <i className="fas fa-upload me-2"></i>
                  Upload Your First File
                </Button>
              </div>
            )}
          </Card.Body>
        </Card>
      )}

      {/* Upload Modal */}
      <Modal show={showUploadModal} onHide={handleCloseUploadModal} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>
            <i className="fas fa-upload me-2"></i>
            Upload File
          </Modal.Title>
        </Modal.Header>
        <Form>
          <Modal.Body>
            <Form.Group className="mb-3">
              <Form.Label>
                Select File <span className="text-danger">*</span>
              </Form.Label>
              <Form.Control type="file" onChange={handleFileChange} required />
              <Form.Text className="text-muted">
                Supported file types: PDF, DOC, DOCX, XLS, XLSX, PPT, PPTX, JPG,
                PNG, GIF, TXT, ZIP, RAR
              </Form.Text>
            </Form.Group>

            {selectedFile && (
              <Card className="mb-3">
                <Card.Body>
                  <div className="d-flex align-items-center">
                    <div
                      className="bg-light rounded-circle d-flex align-items-center justify-content-center me-3"
                      style={{ width: "48px", height: "48px" }}
                    >
                      {getFileIcon(
                        selectedFile.type.split("/")[1] || "unknown",
                      )}
                    </div>
                    <div>
                      <div className="fw-bold">{selectedFile.name}</div>
                      <div className="small text-muted">
                        {(selectedFile.size / 1024).toFixed(2)} KB
                      </div>
                    </div>
                  </div>
                </Card.Body>
              </Card>
            )}

            <Form.Group className="mb-3">
              <Form.Label>Description</Form.Label>
              <Form.Control
                as="textarea"
                rows={3}
                placeholder="Enter file description (optional)"
              />
            </Form.Group>

            <Form.Group className="mb-3">
              <Form.Label>Category</Form.Label>
              <Form.Select>
                <option value="">Select category</option>
                <option value="contract">Contract</option>
                <option value="evidence">Evidence</option>
                <option value="letter">Letter</option>
                <option value="report">Report</option>
                <option value="other">Other</option>
              </Form.Select>
            </Form.Group>
          </Modal.Body>
          <Modal.Footer>
            <Button variant="secondary" onClick={handleCloseUploadModal}>
              <i className="fas fa-times me-2"></i>
              Cancel
            </Button>
            <Button
              variant="primary"
              onClick={handleUpload}
              disabled={!selectedFile || uploading}
            >
              {uploading ? (
                <>
                  <i className="fas fa-spinner fa-spin me-2"></i>
                  Uploading...
                </>
              ) : (
                <>
                  <i className="fas fa-upload me-2"></i>
                  Upload File
                </>
              )}
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>
    </div>
  );
};

export default FilesPage;
