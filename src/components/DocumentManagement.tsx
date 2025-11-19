import React, { useState, useEffect } from 'react'
import {
  Card,
  Table,
  Button,
  Input,
  Select,
  Tag,
  Modal,
  Form,
  Upload,
  message,
  Space,
  Tooltip,
  Popconfirm,
  Badge,
  Dropdown,
  Menu,
  Typography,
  Row,
  Col,
  Statistic,
  Progress,
} from 'antd'
import {
  UploadOutlined,
  DownloadOutlined,
  EyeOutlined,
  EditOutlined,
  DeleteOutlined,
  ShareAltOutlined,
  HistoryOutlined,
  SearchOutlined,
  FileTextOutlined,
  FilePdfOutlined,
  FileWordOutlined,
  FileExcelOutlined,
  FileImageOutlined,
  FileUnknownOutlined,
  PlusOutlined,
  FilterOutlined,
  SortAscendingOutlined,
} from '@ant-design/icons'
import moment from 'moment'

const { Option } = Select
const { Search } = Input
const { Text } = Typography

interface Document {
  id: number
  docNo: string
  docName: string
  docType: string
  docCategory: string
  caseId?: number
  caseName?: string
  clientId?: number
  clientName?: string
  fileName: string
  fileSize: number
  fileType: string
  version: number
  isTemplate: boolean
  docStatus: 'DRAFT' | 'REVIEW' | 'APPROVED' | 'ARCHIVED'
  authorName: string
  createTime: string
  updateTime: string
  viewCount: number
  downloadCount: number
  accessLevel: number
}

interface DocumentVersion {
  id: number
  version: number
  versionName: string
  fileName: string
  fileSize: number
  authorName: string
  createTime: string
  changeLog: string
}

const DocumentManagement: React.FC = () => {
  const [form] = Form.useForm()
  const [documents, setDocuments] = useState<Document[]>([])
  const [versions, setVersions] = useState<DocumentVersion[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [versionModalVisible, setVersionModalVisible] = useState(false)
  const [previewModalVisible, setPreviewModalVisible] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [selectedDocument, setSelectedDocument] = useState<Document | null>(null)
  const [filters, setFilters] = useState({
    docType: '',
    docCategory: '',
    docStatus: '',
    keyword: '',
  })

  // 文档类型选项
  const docTypes = [
    { value: 'CONTRACT', label: '合同类' },
    { value: 'LITIGATION', label: '诉讼类' },
    { value: 'ADVICE', label: '咨询类' },
    { value: 'RESEARCH', label: '研究类' },
    { value: 'TEMPLATE', label: '模板类' },
    { value: 'OTHER', label: '其他' },
  ]

  // 文档分类选项
  const docCategories = [
    { value: 'LEGAL', label: '法律文书' },
    { value: 'EVIDENCE', label: '证据材料' },
    { value: 'CORRESPONDENCE', label: '往来函件' },
    { value: 'RESEARCH', label: '研究报告' },
    { value: 'ADMIN', label: '行政管理' },
    { value: 'OTHER', label: '其他' },
  ]

  // 状态配置
  const statusConfig = {
    DRAFT: { color: 'default', text: '草稿' },
    REVIEW: { color: 'processing', text: '审核中' },
    APPROVED: { color: 'success', text: '已批准' },
    ARCHIVED: { color: 'purple', text: '已归档' },
  }

  // 访问级别配置
  const accessLevelConfig = {
    1: { color: 'green', text: '公开' },
    2: { color: 'blue', text: '内部' },
    3: { color: 'red', text: '机密' },
  }

  // 模拟数据
  useEffect(() => {
    fetchDocuments()
  }, [])

  const fetchDocuments = () => {
    setLoading(true)
    // 模拟API调用
    setTimeout(() => {
      const mockDocuments: Document[] = [
        {
          id: 1,
          docNo: 'DOC-2024-001',
          docName: '张三诉李四合同纠纷案起诉状',
          docType: 'LITIGATION',
          docCategory: 'LEGAL',
          caseId: 1,
          caseName: '张三诉李四合同纠纷案',
          fileName: '张三诉李四合同纠纷案起诉状.pdf',
          fileSize: 1024 * 256,
          fileType: 'pdf',
          version: 1,
          isTemplate: false,
          docStatus: 'APPROVED',
          authorName: '张律师',
          createTime: '2024-01-15 10:00:00',
          updateTime: '2024-01-15 10:00:00',
          viewCount: 15,
          downloadCount: 3,
          accessLevel: 2,
        },
        {
          id: 2,
          docNo: 'DOC-2024-002',
          docName: '法律咨询意见书',
          docType: 'ADVICE',
          docCategory: 'LEGAL',
          clientId: 1,
          clientName: '某科技有限公司',
          fileName: '法律咨询意见书.docx',
          fileSize: 1024 * 128,
          fileType: 'docx',
          version: 2,
          isTemplate: false,
          docStatus: 'REVIEW',
          authorName: '李律师',
          createTime: '2024-01-14 14:00:00',
          updateTime: '2024-01-15 09:00:00',
          viewCount: 8,
          downloadCount: 1,
          accessLevel: 2,
        },
      ]
      setDocuments(mockDocuments)
      setLoading(false)
    }, 1000)
  }

  // 获取文件图标
  const getFileIcon = (fileType: string) => {
    const iconMap: { [key: string]: any } = {
      pdf: <FilePdfOutlined style={{ fontSize: '24px', color: '#ff4d4f' }} />,
      doc: <FileWordOutlined style={{ fontSize: '24px', color: '#1890ff' }} />,
      docx: <FileWordOutlined style={{ fontSize: '24px', color: '#1890ff' }} />,
      xls: <FileExcelOutlined style={{ fontSize: '24px', color: '#52c41a' }} />,
      xlsx: <FileExcelOutlined style={{ fontSize: '24px', color: '#52c41a' }} />,
      jpg: <FileImageOutlined style={{ fontSize: '24px', color: '#722ed1' }} />,
      jpeg: <FileImageOutlined style={{ fontSize: '24px', color: '#722ed1' }} />,
      png: <FileImageOutlined style={{ fontSize: '24px', color: '#722ed1' }} />,
    }
    return (
      iconMap[fileType.toLowerCase()] || (
        <FileUnknownOutlined style={{ fontSize: '24px', color: '#d9d9d9' }} />
      )
    )
  }

  // 格式化文件大小
  const formatFileSize = (bytes: number) => {
    if (bytes === 0) {
      return '0 B'
    }
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
  }

  // 上传文件
  const handleUpload = async (file: File) => {
    setUploading(true)
    try {
      // 模拟上传过程
      await new Promise((resolve) => setTimeout(resolve, 2000))

      message.success('文件上传成功')
      setModalVisible(false)
      fetchDocuments() // 刷新列表
    } catch (error) {
      message.error('上传失败')
    }
    setUploading(false)
    return false // 阻止默认上传行为
  }

  // 查看版本历史
  const viewVersionHistory = (document: Document) => {
    setSelectedDocument(document)

    // 模拟版本数据
    const mockVersions: DocumentVersion[] = [
      {
        id: 1,
        version: 1,
        versionName: '初版',
        fileName: document.fileName,
        fileSize: document.fileSize,
        authorName: document.authorName,
        createTime: document.createTime,
        changeLog: '初始版本',
      },
      {
        id: 2,
        version: 2,
        versionName: '修订版',
        fileName: document.fileName,
        fileSize: document.fileSize + 1024,
        authorName: '李律师',
        createTime: '2024-01-16 10:00:00',
        changeLog: '修改了第3页的条款',
      },
    ]
    setVersions(mockVersions)
    setVersionModalVisible(true)
  }

  // 预览文档
  const previewDocument = (document: Document) => {
    setSelectedDocument(document)
    setPreviewModalVisible(true)
  }

  // 下载文档
  const downloadDocument = (document: Document) => {
    message.success(`正在下载: ${document.fileName}`)
    // 模拟下载
    setTimeout(() => {
      message.success('下载完成')
    }, 2000)
  }

  // 删除文档
  const deleteDocument = (id: number) => {
    setDocuments((prev) => prev.filter((doc) => doc.id !== id))
    message.success('删除成功')
  }

  // 分享文档
  const shareDocument = (document: Document) => {
    message.success('分享链接已复制到剪贴板')
  }

  // 筛选文档
  const filteredDocuments = documents.filter((doc) => {
    return (
      (!filters.docType || doc.docType === filters.docType) &&
      (!filters.docCategory || doc.docCategory === filters.docCategory) &&
      (!filters.docStatus || doc.docStatus === filters.docStatus) &&
      (!filters.keyword ||
        doc.docName.toLowerCase().includes(filters.keyword.toLowerCase()) ||
        doc.fileName.toLowerCase().includes(filters.keyword.toLowerCase()))
    )
  })

  // 统计数据
  const statistics = {
    totalDocs: documents.length,
    approvedDocs: documents.filter((doc) => doc.docStatus === 'APPROVED').length,
    reviewDocs: documents.filter((doc) => doc.docStatus === 'REVIEW').length,
    totalSize: documents.reduce((sum, doc) => sum + doc.fileSize, 0),
  }

  const columns = [
    {
      title: '文档名称',
      dataIndex: 'docName',
      key: 'docName',
      width: 250,
      render: (text: string, record: Document) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          {getFileIcon(record.fileType)}
          <div>
            <div style={{ fontWeight: 500 }}>{text}</div>
            <div style={{ fontSize: '12px', color: '#666' }}>
              {record.fileName} (v{record.version})
            </div>
          </div>
        </div>
      ),
    },
    {
      title: '类型',
      dataIndex: 'docType',
      key: 'docType',
      width: 100,
      render: (type: string) => docTypes.find((t) => t.value === type)?.label || type,
    },
    {
      title: '分类',
      dataIndex: 'docCategory',
      key: 'docCategory',
      width: 100,
      render: (category: string) =>
        docCategories.find((c) => c.value === category)?.label || category,
    },
    {
      title: '关联案件/客户',
      key: 'relation',
      width: 150,
      render: (text: any, record: Document) => (
        <div>
          {record.caseName && (
            <div style={{ fontSize: '12px' }}>
              <Tag color='blue'>案件</Tag>
              {record.caseName}
            </div>
          )}
          {record.clientName && (
            <div style={{ fontSize: '12px' }}>
              <Tag color='green'>客户</Tag>
              {record.clientName}
            </div>
          )}
        </div>
      ),
    },
    {
      title: '大小',
      dataIndex: 'fileSize',
      key: 'fileSize',
      width: 80,
      render: (size: number) => formatFileSize(size),
    },
    {
      title: '状态',
      dataIndex: 'docStatus',
      key: 'docStatus',
      width: 80,
      render: (status: string) => {
        const config = statusConfig[status as keyof typeof statusConfig]
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: '访问级别',
      dataIndex: 'accessLevel',
      key: 'accessLevel',
      width: 80,
      render: (level: number) => {
        const config = accessLevelConfig[level as keyof typeof accessLevelConfig]
        return <Badge color={config.color} text={config.text} />
      },
    },
    {
      title: '创建者',
      dataIndex: 'authorName',
      key: 'authorName',
      width: 80,
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      key: 'createTime',
      width: 120,
      render: (time: string) => moment(time).format('MM-DD HH:mm'),
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (text: any, record: Document) => (
        <Space>
          <Tooltip title='预览'>
            <Button
              type='link'
              size='small'
              icon={<EyeOutlined />}
              onClick={() => previewDocument(record)}
            />
          </Tooltip>

          <Tooltip title='下载'>
            <Button
              type='link'
              size='small'
              icon={<DownloadOutlined />}
              onClick={() => downloadDocument(record)}
            />
          </Tooltip>

          <Tooltip title='版本历史'>
            <Button
              type='link'
              size='small'
              icon={<HistoryOutlined />}
              onClick={() => viewVersionHistory(record)}
            />
          </Tooltip>

          <Dropdown
            overlay={
              <Menu>
                <Menu.Item
                  key='share'
                  icon={<ShareAltOutlined />}
                  onClick={() => shareDocument(record)}
                >
                  分享
                </Menu.Item>
                <Menu.Item key='edit' icon={<EditOutlined />}>
                  编辑
                </Menu.Item>
                <Menu.Item
                  key='delete'
                  icon={<DeleteOutlined />}
                  danger
                  onClick={() => deleteDocument(record.id)}
                >
                  删除
                </Menu.Item>
              </Menu>
            }
          >
            <Button type='link' size='small'>
              更多
            </Button>
          </Dropdown>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: '24px' }}>
      <Card>
        {/* 统计信息 */}
        <Row gutter={16} style={{ marginBottom: '24px' }}>
          <Col span={6}>
            <Statistic title='总文档数' value={statistics.totalDocs} />
          </Col>
          <Col span={6}>
            <Statistic title='已批准' value={statistics.approvedDocs} />
          </Col>
          <Col span={6}>
            <Statistic title='审核中' value={statistics.reviewDocs} />
          </Col>
          <Col span={6}>
            <Statistic
              title='总大小'
              value={statistics.totalSize / (1024 * 1024)}
              suffix='MB'
              precision={2}
            />
          </Col>
        </Row>

        {/* 操作栏 */}
        <div
          style={{
            marginBottom: '24px',
            display: 'flex',
            justifyContent: 'space-between',
          }}
        >
          <Space>
            <Button type='primary' icon={<PlusOutlined />} onClick={() => setModalVisible(true)}>
              上传文档
            </Button>
            <Button icon={<FilterOutlined />}>高级筛选</Button>
            <Button icon={<SortAscendingOutlined />}>排序</Button>
          </Space>

          <Space>
            <Select
              style={{ width: 120 }}
              placeholder='文档类型'
              allowClear
              value={filters.docType}
              onChange={(value) => setFilters((prev) => ({ ...prev, docType: value }))}
            >
              {docTypes.map((type) => (
                <Option key={type.value} value={type.value}>
                  {type.label}
                </Option>
              ))}
            </Select>

            <Select
              style={{ width: 120 }}
              placeholder='文档分类'
              allowClear
              value={filters.docCategory}
              onChange={(value) => setFilters((prev) => ({ ...prev, docCategory: value }))}
            >
              {docCategories.map((category) => (
                <Option key={category.value} value={category.value}>
                  {category.label}
                </Option>
              ))}
            </Select>

            <Select
              style={{ width: 120 }}
              placeholder='状态'
              allowClear
              value={filters.docStatus}
              onChange={(value) => setFilters((prev) => ({ ...prev, docStatus: value }))}
            >
              {Object.entries(statusConfig).map(([key, config]) => (
                <Option key={key} value={key}>
                  {config.text}
                </Option>
              ))}
            </Select>

            <Search
              placeholder='搜索文档'
              style={{ width: 200 }}
              value={filters.keyword}
              onChange={(e) => setFilters((prev) => ({ ...prev, keyword: e.target.value }))}
            />
          </Space>
        </div>

        {/* 文档列表 */}
        <Table
          columns={columns}
          dataSource={filteredDocuments}
          rowKey='id'
          loading={loading}
          pagination={{
            total: filteredDocuments.length,
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条/共 ${total} 条`,
          }}
        />
      </Card>

      {/* 上传文档弹窗 */}
      <Modal
        title='上传文档'
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={600}
      >
        <Form form={form} layout='vertical'>
          <Form.Item
            name='docName'
            label='文档名称'
            rules={[{ required: true, message: '请输入文档名称' }]}
          >
            <Input placeholder='输入文档名称' />
          </Form.Item>

          <Form.Item
            name='docType'
            label='文档类型'
            rules={[{ required: true, message: '请选择文档类型' }]}
          >
            <Select placeholder='选择文档类型'>
              {docTypes.map((type) => (
                <Option key={type.value} value={type.value}>
                  {type.label}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name='docCategory'
            label='文档分类'
            rules={[{ required: true, message: '请选择文档分类' }]}
          >
            <Select placeholder='选择文档分类'>
              {docCategories.map((category) => (
                <Option key={category.value} value={category.value}>
                  {category.label}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item name='accessLevel' label='访问级别' initialValue={2}>
            <Select>
              {Object.entries(accessLevelConfig).map(([key, config]) => (
                <Option key={key} value={parseInt(key)}>
                  {config.text}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name='file'
            label='选择文件'
            rules={[{ required: true, message: '请选择文件' }]}
          >
            <Upload.Dragger beforeUpload={handleUpload} showUploadList={false} multiple={false}>
              <p className='ant-upload-drag-icon'>
                <UploadOutlined />
              </p>
              <p className='ant-upload-text'>点击或拖拽文件到此处上传</p>
              <p className='ant-upload-hint'>支持单个文件上传，文件大小不超过50MB</p>
            </Upload.Dragger>
          </Form.Item>

          {uploading && (
            <div style={{ textAlign: 'center', padding: '20px' }}>
              <Progress percent={70} status='active' />
              <Text>正在上传文件...</Text>
            </div>
          )}
        </Form>
      </Modal>

      {/* 版本历史弹窗 */}
      <Modal
        title='版本历史'
        open={versionModalVisible}
        onCancel={() => setVersionModalVisible(false)}
        footer={null}
        width={800}
      >
        {selectedDocument && (
          <div>
            <div
              style={{
                marginBottom: '16px',
                padding: '16px',
                backgroundColor: '#f5f5f5',
                borderRadius: '4px',
              }}
            >
              <Text strong>当前文档：</Text>
              <Text>{selectedDocument.docName}</Text>
            </div>

            <Table
              columns={[
                {
                  title: '版本',
                  dataIndex: 'version',
                  key: 'version',
                  width: 80,
                },
                {
                  title: '版本名称',
                  dataIndex: 'versionName',
                  key: 'versionName',
                  width: 120,
                },
                {
                  title: '文件大小',
                  dataIndex: 'fileSize',
                  key: 'fileSize',
                  width: 100,
                  render: (size: number) => formatFileSize(size),
                },
                {
                  title: '创建者',
                  dataIndex: 'authorName',
                  key: 'authorName',
                  width: 100,
                },
                {
                  title: '创建时间',
                  dataIndex: 'createTime',
                  key: 'createTime',
                  width: 120,
                  render: (time: string) => moment(time).format('YYYY-MM-DD HH:mm'),
                },
                { title: '变更说明', dataIndex: 'changeLog', key: 'changeLog' },
                {
                  title: '操作',
                  key: 'action',
                  width: 120,
                  render: (text: any, record: DocumentVersion) => (
                    <Space>
                      <Button type='link' size='small'>
                        下载
                      </Button>
                      <Button type='link' size='small'>
                        对比
                      </Button>
                    </Space>
                  ),
                },
              ]}
              dataSource={versions}
              rowKey='id'
              pagination={false}
            />
          </div>
        )}
      </Modal>

      {/* 预览弹窗 */}
      <Modal
        title='文档预览'
        open={previewModalVisible}
        onCancel={() => setPreviewModalVisible(false)}
        footer={null}
        width={1000}
        style={{ top: 20 }}
      >
        {selectedDocument && (
          <div>
            <div
              style={{
                marginBottom: '16px',
                padding: '16px',
                backgroundColor: '#f5f5f5',
                borderRadius: '4px',
              }}
            >
              <Row gutter={16}>
                <Col span={12}>
                  <Text strong>文档名称：</Text>
                  <Text>{selectedDocument.docName}</Text>
                </Col>
                <Col span={12}>
                  <Text strong>文件大小：</Text>
                  <Text>{formatFileSize(selectedDocument.fileSize)}</Text>
                </Col>
              </Row>
            </div>

            <div
              style={{
                height: '600px',
                backgroundColor: '#f0f0f0',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                borderRadius: '4px',
              }}
            >
              <div style={{ textAlign: 'center' }}>
                {getFileIcon(selectedDocument.fileType)}
                <div style={{ marginTop: '16px' }}>
                  <Text>文档预览功能开发中...</Text>
                </div>
                <Button
                  type='primary'
                  style={{ marginTop: '16px' }}
                  onClick={() => downloadDocument(selectedDocument)}
                >
                  下载文档
                </Button>
              </div>
            </div>
          </div>
        )}
      </Modal>
    </div>
  )
}

export default DocumentManagement
