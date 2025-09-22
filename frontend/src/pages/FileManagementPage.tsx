import React, { useState } from 'react';
import {
  FaPlus,
  FaTrash,
  FaMagnifyingGlass,
  FaFolder,
  FaFileLines,
  FaImage,
  FaMusic,
  FaFilm,
  FaUpload,
  FaDownload,
  FaEye,
  FaShare,
  FaFilter,
  FaCalendar,
  FaUser,
  FaTriangleExclamation
} from 'react-icons/fa6';
import {
  Row,
  Col,
  Card,
  Table,
  Button,
  Form,
  InputGroup,
  Badge,
  Tabs,
  Alert,
  Modal
} from 'react-bootstrap';

interface FileItem {
  id: string;
  name: string;
  type: string;
  size: number;
  category: string;
  uploadedBy: string;
  uploadedDate: string;
  lastModified: string;
  description: string;
  isFavorite: boolean;
  isShared: boolean;
  tags: string[];
}

const FileManagement: React.FC = () => {
  const [files, setFiles] = useState<FileItem[]>([
    {
      id: '1',
      name: '张三借款纠纷合同.pdf',
      type: 'pdf',
      size: 2048576,
      category: '合同文件',
      uploadedBy: '王律师',
      uploadedDate: '2024-01-15',
      lastModified: '2024-01-15',
      description: '张三借款纠纷案件的委托合同',
      isFavorite: true,
      isShared: false,
      tags: ['合同', '借款纠纷', '张三']
    },
    {
      id: '2',
      name: 'ABC公司服务协议.docx',
      type: 'docx',
      size: 1024000,
      category: '合同文件',
      uploadedBy: '李律师',
      uploadedDate: '2024-01-14',
      lastModified: '2024-01-16',
      description: 'ABC公司法律服务协议',
      isFavorite: false,
      isShared: true,
      tags: ['协议', 'ABC公司', '服务']
    },
    {
      id: '3',
      name: '证据材料汇总.xlsx',
      type: 'xlsx',
      size: 512000,
      category: '案件材料',
      uploadedBy: '赵律师',
      uploadedDate: '2024-01-13',
      lastModified: '2024-01-13',
      description: '知识产权案件证据材料汇总',
      isFavorite: true,
      isShared: false,
      tags: ['证据', '知识产权', '汇总']
    },
    {
      id: '4',
      name: '庭审照片.jpg',
      type: 'jpg',
      size: 3072000,
      category: '图片资料',
      uploadedBy: '王律师',
      uploadedDate: '2024-01-12',
      lastModified: '2024-01-12',
      description: '庭审现场照片',
      isFavorite: false,
      isShared: true,
      tags: ['照片', '庭审']
    },
    {
      id: '5',
      name: '法律意见书.pdf',
      type: 'pdf',
      size: 1536000,
      category: '法律文书',
      uploadedBy: '钱律师',
      uploadedDate: '2024-01-11',
      lastModified: '2024-01-11',
      description: '企业并购法律意见书',
      isFavorite: true,
      isShared: false,
      tags: ['意见书', '并购', '法律']
    }
  ]);

  const [loading, setLoading] = useState<boolean>(false);
  const [uploadModalVisible, setUploadModalVisible] = useState<boolean>(false);
  const [selectedFiles, setSelectedFiles] = useState<string[]>([]);
  const [searchTerm, setSearchTerm] = useState<string>('');
  const [filterCategory, setFilterCategory] = useState<string>('all');
  const [filterType, setFilterType] = useState<string>('all');
  const [activeTab, setActiveTab] = useState<string>('all');

  const fileCategories = [
    '合同文件',
    '案件材料',
    '法律文书',
    '图片资料',
    '音频资料',
    '视频资料',
    '其他'
  ];

  const fileTypes = [
    'pdf', 'doc', 'docx', 'xls', 'xlsx', 'ppt', 'pptx',
    'jpg', 'jpeg', 'png', 'gif', 'bmp',
    'mp3', 'wav', 'mp4', 'avi', 'mov',
    'zip', 'rar', '7z', 'txt'
  ];

  const getFileIcon = (fileType: string) => {
    if (['jpg', 'jpeg', 'png', 'gif', 'bmp'].includes(fileType)) {
      return <FaImage className="w-8 h-8 text-success" />;
    } else if (['mp3', 'wav'].includes(fileType)) {
      return <FaMusic className="w-8 h-8 text-info" />;
    } else if (['mp4', 'avi', 'mov'].includes(fileType)) {
      return <FaFilm className="w-8 h-8 text-warning" />;
    } else {
      return <FaFileLines className="w-8 h-8 text-primary" />;
    }
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const getFilteredFiles = () => {
    let filtered = files;

    // 根据标签页筛选
    if (activeTab === 'favorites') {
      filtered = filtered.filter(file => file.isFavorite);
    } else if (activeTab === 'shared') {
      filtered = filtered.filter(file => file.isShared);
    } else if (activeTab === 'recent') {
      filtered = filtered.slice(0, 10);
    }

    // 根据搜索词筛选
    if (searchTerm) {
      filtered = filtered.filter(file =>
        file.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        file.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
        file.tags.some(tag => tag.toLowerCase().includes(searchTerm.toLowerCase()))
      );
    }

    // 根据分类筛选
    if (filterCategory !== 'all') {
      filtered = filtered.filter(file => file.category === filterCategory);
    }

    // 根据文件类型筛选
    if (filterType !== 'all') {
      filtered = filtered.filter(file => file.type === filterType);
    }

    return filtered;
  };

  const handleFileSelect = (fileId: string) => {
    setSelectedFiles(prev => {
      if (prev.includes(fileId)) {
        return prev.filter(id => id !== fileId);
      } else {
        return [...prev, fileId];
      }
    });
  };

  const handleSelectAll = () => {
    const filteredFiles = getFilteredFiles();
    if (selectedFiles.length === filteredFiles.length) {
      setSelectedFiles([]);
    } else {
      setSelectedFiles(filteredFiles.map(file => file.id));
    }
  };

  const handleUpload = () => {
    setUploadModalVisible(true);
  };

  const handleDelete = () => {
    if (selectedFiles.length === 0) {
      alert('请选择要删除的文件');
      return;
    }

    if (window.confirm(`确定要删除选中的 ${selectedFiles.length} 个文件吗？`)) {
      setFiles(files.filter(file => !selectedFiles.includes(file.id)));
      setSelectedFiles([]);
      alert('删除成功');
    }
  };

  const handleDownload = () => {
    if (selectedFiles.length === 0) {
      alert('请选择要下载的文件');
      return;
    }
    alert(`开始下载 ${selectedFiles.length} 个文件`);
  };

  const toggleFavorite = (fileId: string) => {
    setFiles(files.map(file =>
      file.id === fileId ? { ...file, isFavorite: !file.isFavorite } : file
    ));
  };

  const filteredFiles = getFilteredFiles();

  // 统计数据
  const stats = {
    totalFiles: files.length,
    totalSize: files.reduce((sum, file) => sum + file.size, 0),
    favoriteFiles: files.filter(f => f.isFavorite).length,
    sharedFiles: files.filter(f => f.isShared).length,
    categoryCounts: fileCategories.reduce((acc, category) => {
      acc[category] = files.filter(f => f.category === category).length;
      return acc;
    }, {} as Record<string, number>)
  };

  return (
    <div className="file-management p-4">
      {/* 头部统计 */}
      <Card className="mb-4">
        <Card.Header>
          <div className="d-flex justify-content-between align-items-center">
            <div>
              <h4 className="mb-0">文件管理</h4>
              <p className="text-muted mb-0">管理律所的所有文档和资料</p>
            </div>
            <Badge bg="primary">
              <FaFolder className="w-4 h-4 me-1" />
              文件
            </Badge>
          </div>
        </Card.Header>
        <Card.Body>
          <div className="row mb-3">
            <div className="col-md-3">
              <Card className="text-center">
                <Card.Body>
                  <h3>{stats.totalFiles}</h3>
                  <p className="text-muted mb-0">总文件数</p>
                </Card.Body>
              </Card>
            </div>
            <div className="col-md-3">
              <Card className="text-center bg-primary text-white">
                <Card.Body>
                  <h3>{formatFileSize(stats.totalSize)}</h3>
                  <p className="mb-0">总存储空间</p>
                </Card.Body>
              </Card>
            </div>
            <div className="col-md-3">
              <Card className="text-center bg-warning text-white">
                <Card.Body>
                  <h3>{stats.favoriteFiles}</h3>
                  <p className="mb-0">收藏文件</p>
                </Card.Body>
              </Card>
            </div>
            <div className="col-md-3">
              <Card className="text-center bg-info text-white">
                <Card.Body>
                  <h3>{stats.sharedFiles}</h3>
                  <p className="mb-0">共享文件</p>
                </Card.Body>
              </Card>
            </div>
          </div>
        </Card.Body>
      </Card>

      {/* 主内容 */}
      <Card>
        <Card.Header>
          <div className="d-flex justify-content-between align-items-center">
            <div className="d-flex align-items-center">
              <InputGroup className="me-3" style={{ width: '250px' }}>
                <InputGroup.Text>
                  <FaMagnifyingGlass className="w-4 h-4" />
                </InputGroup.Text>
                <Form.Control
                  type="text"
                  placeholder="搜索文件名或标签..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                />
              </InputGroup>
              <Form.Select
                value={filterCategory}
                onChange={(e) => setFilterCategory(e.target.value)}
                className="me-2"
                style={{ width: '120px' }}
              >
                <option value="all">所有分类</option>
                {fileCategories.map(category => (
                  <option key={category} value={category}>{category}</option>
                ))}
              </Form.Select>
              <Form.Select
                value={filterType}
                onChange={(e) => setFilterType(e.target.value)}
                style={{ width: '100px' }}
              >
                <option value="all">所有类型</option>
                {fileTypes.map(type => (
                  <option key={type} value={type}>{type.toUpperCase()}</option>
                ))}
              </Form.Select>
            </div>
            <div className="d-flex gap-2">
              {selectedFiles.length > 0 && (
                <>
                  <Button variant="outline-danger" onClick={handleDelete}>
                    <FaTrash className="w-4 h-4 me-1" />
                    删除 ({selectedFiles.length})
                  </Button>
                  <Button variant="outline-primary" onClick={handleDownload}>
                    <FaDownload className="w-4 h-4 me-1" />
                    下载 ({selectedFiles.length})
                  </Button>
                </>
              )}
              <Button variant="primary" onClick={handleUpload}>
                <FaUpload className="w-4 h-4 me-2" />
                上传文件
              </Button>
            </div>
          </div>
        </Card.Header>
        <Card.Body>
          {/* 标签页 */}
          <Tabs
            activeKey={activeTab}
            onSelect={(key) => setActiveTab(key || 'all')}
            className="mb-3"
          >
            <Tab eventKey="all" title={`全部文件 (${stats.totalFiles})`} />
            <Tab eventKey="favorites" title={`收藏 (${stats.favoriteFiles})`} />
            <Tab eventKey="shared" title={`共享 (${stats.sharedFiles})`} />
            <Tab eventKey="recent" title="最近文件" />
          </Tabs>

          {/* 文件列表 */}
          {filteredFiles.length === 0 ? (
            <div className="text-center py-5">
              <FaFolder className="w-16 h-16 text-muted mx-auto mb-3" />
              <h5>没有找到匹配的文件</h5>
              <p className="text-muted">请尝试调整搜索条件或上传新文件</p>
              <Button variant="primary" onClick={handleUpload}>
                <FaUpload className="w-4 h-4 me-2" />
                上传文件
              </Button>
            </div>
          ) : (
            <div className="table-responsive">
              <Table striped hover>
                <thead>
                  <tr>
                    <th style={{ width: '40px' }}>
                      <Form.Check
                        type="checkbox"
                        checked={selectedFiles.length === filteredFiles.length}
                        onChange={handleSelectAll}
                      />
                    </th>
                    <th>文件名</th>
                    <th>分类</th>
                    <th>大小</th>
                    <th>上传者</th>
                    <th>上传时间</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredFiles.map((file) => (
                    <tr key={file.id}>
                      <td>
                        <Form.Check
                          type="checkbox"
                          checked={selectedFiles.includes(file.id)}
                          onChange={() => handleFileSelect(file.id)}
                        />
                      </td>
                      <td>
                        <div className="d-flex align-items-center">
                          <div className="me-3">
                            {getFileIcon(file.type)}
                          </div>
                          <div>
                            <div className="d-flex align-items-center">
                              <strong>{file.name}</strong>
                              {file.isFavorite && (
                                <span className="text-warning ms-2">★</span>
                              )}
                              {file.isShared && (
                                <Badge bg="info" className="ms-2">共享</Badge>
                              )}
                            </div>
                            <div className="text-muted small">
                              {file.description}
                            </div>
                            <div className="mt-1">
                              {file.tags.map(tag => (
                                <Badge key={tag} bg="light" text="dark" className="me-1" style={{ fontSize: '0.75rem' }}>
                                  {tag}
                                </Badge>
                              ))}
                            </div>
                          </div>
                        </div>
                      </td>
                      <td>{file.category}</td>
                      <td>{formatFileSize(file.size)}</td>
                      <td>
                        <div className="d-flex align-items-center">
                          <FaUser className="w-4 h-4 me-1" />
                          {file.uploadedBy}
                        </div>
                      </td>
                      <td>
                        <div className="d-flex align-items-center">
                          <FaCalendar className="w-4 h-4 me-1" />
                          {file.uploadedDate}
                        </div>
                      </td>
                      <td>
                        <div className="btn-group" role="group">
                          <Button
                            variant="outline-primary"
                            size="sm"
                            onClick={() => alert(`预览文件：${file.name}`)}
                          >
                            <FaEye className="w-4 h-4" />
                          </Button>
                          <Button
                            variant="outline-warning"
                            size="sm"
                            onClick={() => toggleFavorite(file.id)}
                          >
                            <span className={file.isFavorite ? 'text-warning' : ''}>★</span>
                          </Button>
                          <Button
                            variant="outline-info"
                            size="sm"
                            onClick={() => alert(`分享文件：${file.name}`)}
                          >
                            <FaShare className="w-4 h-4" />
                          </Button>
                          <Button
                            variant="outline-danger"
                            size="sm"
                            onClick={() => {
                              if (window.confirm(`确定要删除文件 ${file.name} 吗？`)) {
                                setFiles(files.filter(f => f.id !== file.id));
                              }
                            }}
                          >
                            <FaTrash className="w-4 h-4" />
                          </Button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </Table>
            </div>
          )}
        </Card.Body>
      </Card>

      {/* 上传文件模态框 */}
      <Modal show={uploadModalVisible} onHide={() => setUploadModalVisible(false)} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>上传文件</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <div className="text-center py-4">
            <FaUpload className="w-16 h-16 text-primary mx-auto mb-3" />
            <h5>拖拽文件到此处或点击选择</h5>
            <p className="text-muted">支持 PDF、DOC、XLS、PPT、JPG、PNG 等格式</p>
            <Button variant="primary">
              选择文件
            </Button>
          </div>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => setUploadModalVisible(false)}>
            取消
          </Button>
          <Button variant="primary">
            开始上传
          </Button>
        </Modal.Footer>
      </Modal>

      {/* 存储空间提醒 */}
      {stats.totalSize > 10 * 1024 * 1024 * 1024 && (
        <Alert variant="warning" className="mt-4">
          <FaTriangleExclamation className="w-5 h-5 me-2" />
          <strong>存储空间提醒：</strong>
          您已使用 {formatFileSize(stats.totalSize)} 存储空间，建议清理不必要的文件。
        </Alert>
      )}
    </div>
  );
};

export default FileManagement;