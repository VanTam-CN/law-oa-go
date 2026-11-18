import React, { useState, useEffect } from 'react'
import {
  Card,
  Table,
  Button,
  Upload,
  Modal,
  Form,
  Input,
  Select,
  Tag,
  Space,
  Popconfirm,
  message,
  Statistic,
  Row,
  Col,
  Tooltip,
} from 'antd'
import {
  UploadOutlined,
  DownloadOutlined,
  DeleteOutlined,
  InboxOutlined,
  FileTextOutlined,
  PictureOutlined,
  FileExcelOutlined,
  FilePdfOutlined,
  FilePptOutlined,
  FileUnknownOutlined,
  CloudUploadOutlined,
  CloudDownloadOutlined,
  SearchOutlined,
  FolderOpenOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
import type { UploadFile } from 'antd/es/upload/interface'
import {
  FileInfo,
  FileStats,
  FileListResponse,
  UploadResponse,
  BatchDeleteResponse,
  uploadFile,
  getFileList,
  downloadFile,
  deleteFile,
  getFileStats,
  batchDeleteFiles,
  formatFileSize,
  getFileIcon,
  getFileTypeColor,
} from '@/services/file'
import './FileManagement.less'

const { TextArea } = Input
const { Option } = Select

const FileManagement: React.FC = () => {
  const [loading, setLoading] = useState<boolean>(false)
  const [uploading, setUploading] = useState<boolean>(false)
  const [files, setFiles] = useState<FileInfo[]>([])
  const [stats, setStats] = useState<FileStats | null>(null)
  const [uploadModalVisible, setUploadModalVisible] = useState<boolean>(false)
  const [selectedCategory, setSelectedCategory] = useState<string>('')
  const [selectedRowKeys, setSelectedRowKeys] = useState<number[]>([])
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  })

  const [uploadForm] = Form.useForm()

  /**
   * 获取文件列表 - 使用正确的 API 路径和参数
   */
  const fetchFiles = async (page: number = 1, pageSize: number = 10, category?: string) => {
    try {
      setLoading(true)
      const response: FileListResponse = await getFileList({
        category: category || selectedCategory || undefined,
        page,
        page_size: pageSize,
      })

      setFiles(response.documents || [])
      setPagination({
        current: response.page || page,
        pageSize: response.page_size || pageSize,
        total: response.total || 0,
      })
    } catch (error) {
      message.error('获取文件列表失败')
      console.error('Failed to fetch files:', error)
      setFiles([])
      setPagination({
        current: page,
        pageSize,
        total: 0,
      })
    } finally {
      setLoading(false)
    }
  }

  /**
   * 获取统计信息 - 使用正确的 API 路径
   */
  const fetchStats = async () => {
    try {
      const response: FileStats = await getFileStats()
      setStats(response)
    } catch (error) {
      console.error('Failed to fetch file stats:', error)
      // 设置默认的统计信息，避免页面崩溃
      setStats({
        totalFiles: 0,
        totalSize: 0,
        totalSizeMB: 0,
        todayUploads: 0,
        typeStats: {
          文档: 0,
          图片: 0,
          其他: 0,
        },
      })
    }
  }

  // 初始化加载文件列表和统计信息
  useEffect(() => {
    fetchFiles()
    fetchStats()
  }, [])

  // 处理分类变化
  const handleCategoryChange = (category: string | null) => {
    setSelectedCategory(category || '')
    fetchFiles(1, pagination.pageSize, category || undefined)
  }

  /**
   * 处理文件上传 - 使用正确的参数名
   */
  const handleUpload = async (
    file: File,
    category?: string,
    description?: string,
    customName?: string,
  ) => {
    try {
      setUploading(true)
      await uploadFile(file, category, description, customName)
      message.success('文件上传成功')
      setUploadModalVisible(false)
      uploadForm.resetFields()

      // 重新加载数据
      fetchFiles()
      fetchStats()
    } catch (error: any) {
      // 更详细的错误处理
      if (error.response?.status === 403) {
        message.error('没有上传权限，请联系管理员')
      } else if (error.response?.status === 413) {
        message.error('文件太大，请选择较小的文件')
      } else if (error.message?.includes('Network Error')) {
        message.error('网络连接失败，请检查网络后重试')
      } else {
        message.error(`文件上传失败: ${error.message || '未知错误'}`)
      }
      console.error('Failed to upload file:', error)
    } finally {
      setUploading(false)
    }
  }

  /**
   * 处理文件下载 - 使用正确的 API 路径
   */
  const handleDownload = (id: number, displayName?: string) => {
    downloadFile(id, displayName)
  }

  /**
   * 处理文件删除 - 使用数字 ID
   */
  const handleDelete = async (id: number) => {
    try {
      await deleteFile(id)
      message.success('文件删除成功')
      fetchFiles()
      fetchStats()
    } catch (error: any) {
      if (error.response?.status === 403) {
        message.error('没有删除权限，请联系管理员')
      } else if (error.response?.status === 404) {
        message.error('文件不存在，可能已被删除')
      } else {
        message.error(`文件删除失败: ${error.message || '未知错误'}`)
      }
      console.error('Failed to delete file:', error)
    }
  }

  /**
   * 处理批量删除 - 使用数字 ID 数组
   */
  const handleBatchDelete = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请选择要删除的文件')
      return
    }

    try {
      const response: BatchDeleteResponse = await batchDeleteFiles(selectedRowKeys)
      if (response.failedCount === 0) {
        message.success(`成功删除 ${response.successCount} 个文件`)
      } else {
        message.warning(
          `成功删除 ${response.successCount} 个文件，失败 ${response.failedCount} 个文件`,
        )
      }

      setSelectedRowKeys([])
      fetchFiles()
      fetchStats()
    } catch (error: any) {
      if (error.response?.status === 403) {
        message.error('没有删除权限，请联系管理员')
      } else {
        message.error('批量删除失败，请重试')
      }
      console.error('Failed to batch delete files:', error)
    }
  }

  /**
   * 处理分页变化
   */
  const handleTableChange = (page: number, pageSize: number) => {
    fetchFiles(page, pageSize)
  }

  /**
   * 获取文件类型图标
   */
  const getFileTypeIcon = (fileName: string) => {
    const extension = fileName.split('.').pop()?.toLowerCase() || ''

    switch (extension) {
      case 'pdf':
        return <FilePdfOutlined style={{ fontSize: '20px', color: '#f5222d' }} />
      case 'ppt':
      case 'pptx':
        return <FilePptOutlined style={{ fontSize: '20px', color: '#fa8c16' }} />
      case 'doc':
      case 'docx':
        return <FileTextOutlined style={{ fontSize: '20px', color: '#1890ff' }} />
      case 'xls':
      case 'xlsx':
        return <FileExcelOutlined style={{ fontSize: '20px', color: '#52c41a' }} />
      case 'jpg':
      case 'jpeg':
      case 'png':
      case 'gif':
        return <PictureOutlined style={{ fontSize: '20px', color: '#722ed1' }} />
      default:
        return <FileUnknownOutlined style={{ fontSize: '20px', color: '#8c8c8c' }} />
    }
  }

  /**
   * 表格列定义 - 使用正确的后端字段名
   */
  const columns = [
    {
      title: '文件名',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: FileInfo) => (
        <Space>
          {getFileTypeIcon(text)}
          <div>
            <div>{text}</div>
            {record.filename && record.filename !== text && (
              <div style={{ fontSize: '12px', color: '#999' }}>
                原始文件名: {record.filename}
              </div>
            )}
          </div>
        </Space>
      ),
    },
    {
      title: '大小',
      dataIndex: 'filesize',
      key: 'filesize',
      render: (size: number) => formatFileSize(size),
      sorter: (a: FileInfo, b: FileInfo) => a.filesize - b.filesize,
    },
    {
      title: '类型',
      dataIndex: 'mime_type',
      key: 'mime_type',
      render: (type: string) => {
        const displayType = type.split('/')[1]?.toUpperCase() || type.toUpperCase()
        return <Tag color={getFileTypeColor(type)}>{displayType}</Tag>
      },
    },
    {
      title: '分类',
      dataIndex: 'category',
      key: 'category',
      render: (category: string) => <Tag color='blue'>{category}</Tag>,
    },
    {
      title: '上传时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (time: string) => new Date(time).toLocaleString(),
      sorter: (a: FileInfo, b: FileInfo) =>
        new Date(a.created_at).getTime() - new Date(b.created_at).getTime(),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: FileInfo) => (
        <Space size='middle'>
          <Tooltip title='下载文件'>
            <Button
              type='primary'
              size='small'
              icon={<DownloadOutlined />}
              onClick={() => handleDownload(record.id, record.name)}
            >
              下载
            </Button>
          </Tooltip>
          <Tooltip title='删除文件'>
            <Popconfirm
              title='确定要删除这个文件吗？'
              onConfirm={() => handleDelete(record.id)}
              okText='确定'
              cancelText='取消'
            >
              <Button type='primary' danger size='small' icon={<DeleteOutlined />}>
                删除
              </Button>
            </Popconfirm>
          </Tooltip>
        </Space>
      ),
    },
  ]

  /**
   * 行选择配置 - 使用数字 ID
   */
  const rowSelection = {
    selectedRowKeys,
    onChange: (keys: React.Key[], selectedRows: FileInfo[]) => {
      setSelectedRowKeys(keys as number[])
    },
    getCheckboxProps: (record: FileInfo) => ({
      disabled: false,
      name: String(record.id),
    }),
  }

  const hasSelected = selectedRowKeys.length > 0

  return (
    <div className='file-management'>
      {/* 统计信息 */}
      {stats && (
        <Row className='stats-row' gutter={16}>
          <Col xs={12} sm={12} md={6} lg={6}>
            <Card className='total-files-card'>
              <Statistic
                title='总文件数'
                value={stats.totalFiles}
                prefix={<InboxOutlined />}
              />
            </Card>
          </Col>
          <Col xs={12} sm={12} md={6} lg={6}>
            <Card className='storage-card'>
              <Statistic
                title='总存储空间'
                value={stats.totalSizeMB}
                suffix='MB'
                prefix={<CloudUploadOutlined />}
              />
            </Card>
          </Col>
          <Col xs={12} sm={12} md={6} lg={6}>
            <Card>
              <Statistic
                title='文档数量'
                value={stats.typeStats?.['文档'] || 0}
                prefix={<FileTextOutlined />}
              />
            </Card>
          </Col>
          <Col xs={12} sm={12} md={6} lg={6}>
            <Card className='today-uploads-card'>
              <Statistic
                title='今日上传'
                value={stats.todayUploads || 0}
                prefix={<FolderOpenOutlined />}
              />
            </Card>
          </Col>
        </Row>
      )}

      {/* 操作栏 */}
      <Card className='search-card' title='文件管理'>
        <Row justify='space-between' align='middle'>
          <Col>
            <Space>
              <Button
                type='primary'
                icon={<UploadOutlined />}
                onClick={() => setUploadModalVisible(true)}
              >
                上传文件
              </Button>
              {hasSelected && (
                <Popconfirm
                  title={`确定要删除选中的 ${selectedRowKeys.length} 个文件吗？`}
                  onConfirm={handleBatchDelete}
                  okText='确定'
                  cancelText='取消'
                >
                  <Button type='primary' danger icon={<DeleteOutlined />}>
                    批量删除 ({selectedRowKeys.length})
                  </Button>
                </Popconfirm>
              )}
            </Space>
          </Col>
          <Col>
            <Space>
              <Select
                placeholder='选择文件分类'
                allowClear
                style={{ width: 150 }}
                value={selectedCategory}
                onChange={handleCategoryChange}
              >
                <Option value='文档'>文档</Option>
                <Option value='图片'>图片</Option>
                <Option value='表格'>表格</Option>
                <Option value='其他'>其他</Option>
              </Select>
              <Button icon={<ReloadOutlined />} onClick={() => { fetchFiles(); fetchStats() }}>
                刷新
              </Button>
            </Space>
          </Col>
        </Row>
      </Card>

      {/* 文件列表 */}
      <Card title='文件列表'>
        <Table
          columns={columns}
          dataSource={files}
          rowKey={(record) => record.id}
          rowSelection={rowSelection}
          loading={loading}
          pagination={{
            current: pagination.current,
            pageSize: pagination.pageSize,
            total: pagination.total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
            onChange: handleTableChange,
            onShowSizeChange: handleTableChange,
          }}
        />
      </Card>

      {/* 上传模态框 */}
      <Modal
        title='上传文件'
        open={uploadModalVisible}
        onCancel={() => {
          setUploadModalVisible(false)
          uploadForm.resetFields()
        }}
        footer={null}
        width={600}
      >
        <Form
          form={uploadForm}
          layout='vertical'
          onFinish={(values) => {
            const file = values.file?.fileList?.[0]?.originFileObj
            if (file) {
              handleUpload(file, values.category, values.description, values.customName)
            }
          }}
        >
          <Form.Item
            name='file'
            label='选择文件'
            rules={[{ required: true, message: '请选择要上传的文件' }]}
          >
            <Upload.Dragger
              maxCount={1}
              beforeUpload={() => false}
              accept='.pdf,.ppt,.pptx,.doc,.docx,.xls,.xlsx,.jpg,.jpeg,.png,.gif,.txt'
            >
              <p className='ant-upload-drag-icon'>
                <InboxOutlined />
              </p>
              <p className='ant-upload-text'>点击或拖拽文件到此区域上传</p>
              <p className='ant-upload-hint'>
                支持单个文件上传，文件大小不超过100MB
                <br />
                支持的文件类型：PDF、PPT、Word、Excel、图片、文本文件
              </p>
            </Upload.Dragger>
          </Form.Item>

          <Form.Item
            name='customName'
            label='自定义文件名'
            extra='留空将使用原始文件名'
          >
            <Input
              placeholder='请输入自定义文件名（可包含扩展名）'
              onBlur={(e) => {
                const file = uploadForm.getFieldValue('file')?.fileList?.[0]
                if (file && e.target.value) {
                  const originalName = file.originFileObj.name
                  const extension = originalName.includes('.')
                    ? originalName.substring(originalName.lastIndexOf('.'))
                    : ''
                  const currentName = e.target.value

                  // 只有在没有扩展名时才添加，避免重复添加
                  if (extension && !currentName.includes('.')) {
                    uploadForm.setFieldsValue({
                      customName: currentName + extension,
                    })
                  }
                }
              }}
            />
          </Form.Item>

          <Form.Item
            name='category'
            label='文件分类'
            rules={[{ required: true, message: '请选择文件分类' }]}
          >
            <Select placeholder='请选择文件分类'>
              <Option value='文档'>文档</Option>
              <Option value='图片'>图片</Option>
              <Option value='表格'>表格</Option>
              <Option value='其他'>其他</Option>
            </Select>
          </Form.Item>

          <Form.Item name='description' label='文件描述'>
            <TextArea rows={3} placeholder='请输入文件描述（可选）' />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button
                type='primary'
                htmlType='submit'
                loading={uploading}
                icon={<UploadOutlined />}
              >
                {uploading ? '上传中...' : '开始上传'}
              </Button>
              <Button
                onClick={() => {
                  setUploadModalVisible(false)
                  uploadForm.resetFields()
                }}
              >
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default FileManagement
