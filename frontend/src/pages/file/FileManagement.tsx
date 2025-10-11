import React, { useState, useEffect } from "react";
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
  Progress,
  Statistic,
  Row,
  Col,
  Tooltip,
  Typography,
} from "antd";
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
  InfoCircleOutlined,
} from "@ant-design/icons";
import type { UploadFile } from "antd/es/upload/interface";
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
} from "@/services/file";

const { TextArea } = Input;
const { Option } = Select;
const { Text } = Typography;

const FileManagement: React.FC = () => {
  const [loading, setLoading] = useState<boolean>(false);
  const [uploading, setUploading] = useState<boolean>(false);
  const [files, setFiles] = useState<FileInfo[]>([]);
  const [fileInfoMap, setFileInfoMap] = useState<Map<string, FileInfo>>(
    new Map(),
  );
  const [stats, setStats] = useState<FileStats | null>(null);
  const [uploadModalVisible, setUploadModalVisible] = useState<boolean>(false);
  const [selectedCategory, setSelectedCategory] = useState<string>("");
  const [selectedRowKeys, setSelectedRowKeys] = useState<string[]>([]);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });

  const [uploadForm] = Form.useForm();

  // 最简单的函数定义
  const fetchFiles = async (
    pageNum: number = 1,
    pageSize: number = 10,
    category?: string,
  ) => {
    try {
      setLoading(true);
      const response: FileListResponse = await getFileList({
        category: category || selectedCategory || undefined,
        pageNum,
        pageSize,
      });

      // 合并文件信息映射中的自定义文件名
      const mergedFiles =
        response.rows?.map((file) => {
          const customInfo = fileInfoMap.get(file.id || file.name);
          return customInfo
            ? {
                ...file,
                name: customInfo.name,
                originalName: customInfo.originalName,
              }
            : file;
        }) || [];

      setFiles(mergedFiles);
      setPagination({
        current: pageNum,
        pageSize,
        total: response.total || 0,
      });
    } catch (error) {
      message.error("获取文件列表失败");
      console.error("Failed to fetch files:", error);
      // 设置空列表，避免页面崩溃
      setFiles([]);
      setPagination({
        current: pageNum,
        pageSize,
        total: 0,
      });
    } finally {
      setLoading(false);
    }
  };

  const fetchStats = async () => {
    try {
      const response: FileStats = await getFileStats();
      setStats(response);
    } catch (error) {
      console.error("Failed to fetch file stats:", error);
      // 设置默认的统计信息，避免页面崩溃
      setStats({
        totalFiles: 0,
        totalSize: 0,
        totalSizeMB: 0,
        typeStats: {
          文档: 0,
          图片: 0,
          其他: 0,
        },
      });
    }
  };

  // 初始化加载文件列表和统计信息
  useEffect(() => {
    fetchFiles();
    fetchStats();
  }, []);

  // 处理分类变化
  const handleCategoryChange = (category: string | null) => {
    setSelectedCategory(category || "");
    fetchFiles(1, pagination.pageSize, category || undefined);
  };

  const handleUpload = async (
    file: File,
    category?: string,
    description?: string,
    customName?: string,
  ) => {
    try {
      setUploading(true);
      const response: UploadResponse = await uploadFile(
        file,
        category,
        description,
        customName,
      );
      message.success("文件上传成功");
      setUploadModalVisible(false);
      uploadForm.resetFields();

      // 更新文件信息映射
      const newFileInfo: FileInfo = {
        id: response.id,
        name: response.name,
        originalName: response.originalName,
        size: response.size,
        contentType: response.contentType,
        category: response.category,
        description: response.description,
        uploadTime: response.uploadTime,
        uploadPath: response.uploadPath,
        type: response.contentType,
        path: response.uploadPath,
        url: response.uploadPath,
        lastModified: response.uploadTime,
      };

      setFileInfoMap((prev) => {
        const newMap = new Map(prev);
        newMap.set(response.id, newFileInfo);
        return newMap;
      });

      fetchFiles();
      fetchStats();
    } catch (error: any) {
      // 更详细的错误处理
      if (error.response?.status === 403) {
        message.error("没有上传权限，请联系管理员");
      } else if (error.response?.status === 413) {
        message.error("文件太大，请选择较小的文件");
      } else if (error.message?.includes("Network Error")) {
        message.error("网络连接失败，请检查网络后重试");
      } else {
        message.error("文件上传失败，请重试");
      }
      console.error("Failed to upload file:", error);
    } finally {
      setUploading(false);
    }
  };

  const handleDownload = (uniqueName: string, displayName?: string) => {
    downloadFile(uniqueName, displayName);
  };

  const handleDelete = async (fileName: string) => {
    try {
      await deleteFile(fileName);
      message.success("文件删除成功");

      // 从文件信息映射中删除对应的文件信息
      setFileInfoMap((prev) => {
        const newMap = new Map(prev);
        newMap.delete(fileName);
        return newMap;
      });

      fetchFiles();
      fetchStats();
    } catch (error: any) {
      if (error.response?.status === 403) {
        message.error("没有删除权限，请联系管理员");
      } else if (error.response?.status === 404) {
        message.error("文件不存在，可能已被删除");
      } else {
        message.error("文件删除失败，请重试");
      }
      console.error("Failed to delete file:", error);
    }
  };

  const handleBatchDelete = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning("请选择要删除的文件");
      return;
    }

    try {
      const response: BatchDeleteResponse =
        await batchDeleteFiles(selectedRowKeys);
      if (response.failedCount === 0) {
        message.success(`成功删除 ${response.successCount} 个文件`);
      } else {
        message.warning(
          `成功删除 ${response.successCount} 个文件，失败 ${response.failedCount} 个文件`,
        );
      }

      // 从文件信息映射中删除成功删除的文件信息
      setFileInfoMap((prev) => {
        const newMap = new Map(prev);
        response.successFiles.forEach((fileId) => {
          newMap.delete(fileId);
        });
        return newMap;
      });

      setSelectedRowKeys([]);
      fetchFiles();
      fetchStats();
    } catch (error: any) {
      if (error.response?.status === 403) {
        message.error("没有删除权限，请联系管理员");
      } else {
        message.error("批量删除失败，请重试");
      }
      console.error("Failed to batch delete files:", error);
    }
  };

  const handleTableChange = (page: number, pageSize: number) => {
    // 直接使用新的分页参数加载数据
    fetchFiles(page, pageSize);
  };

  const getFileTypeIcon = (fileName: string) => {
    const extension = fileName.split(".").pop()?.toLowerCase() || "";

    switch (extension) {
      case "pdf":
        return (
          <FilePdfOutlined style={{ fontSize: "20px", color: "#f5222d" }} />
        );
      case "ppt":
      case "pptx":
        return (
          <FilePptOutlined style={{ fontSize: "20px", color: "#fa8c16" }} />
        );
      case "doc":
      case "docx":
        return (
          <FileTextOutlined style={{ fontSize: "20px", color: "#1890ff" }} />
        );
      case "xls":
      case "xlsx":
        return (
          <FileExcelOutlined style={{ fontSize: "20px", color: "#52c41a" }} />
        );
      case "jpg":
      case "jpeg":
      case "png":
      case "gif":
        return (
          <PictureOutlined style={{ fontSize: "20px", color: "#722ed1" }} />
        );
      default:
        return (
          <FileUnknownOutlined style={{ fontSize: "20px", color: "#8c8c8c" }} />
        );
    }
  };

  const columns = [
    {
      title: "文件名",
      dataIndex: "name",
      key: "name",
      render: (text: string, record: FileInfo) => (
        <Space>
          {getFileTypeIcon(text)}
          <div>
            <Text>{text}</Text>
            {record.originalName && record.originalName !== text && (
              <div style={{ fontSize: "12px", color: "#999" }}>
                原始文件名: {record.originalName}
              </div>
            )}
          </div>
        </Space>
      ),
    },
    {
      title: "大小",
      dataIndex: "size",
      key: "size",
      render: (size: number) => formatFileSize(size),
      sorter: (a: FileInfo, b: FileInfo) => a.size - b.size,
    },
    {
      title: "类型",
      dataIndex: "type",
      key: "type",
      render: (type: string) => (
        <Tag color={getFileTypeColor(type)}>{type}</Tag>
      ),
    },
    {
      title: "分类",
      dataIndex: "category",
      key: "category",
      render: (category: string) => <Tag color="blue">{category}</Tag>,
    },
    {
      title: "上传时间",
      dataIndex: "uploadTime",
      key: "uploadTime",
      render: (time: string) => new Date(time).toLocaleString(),
      sorter: (a: FileInfo, b: FileInfo) =>
        new Date(a.uploadTime).getTime() - new Date(b.uploadTime).getTime(),
    },
    {
      title: "操作",
      key: "action",
      render: (_: any, record: FileInfo) => (
        <Space size="middle">
          <Tooltip title="下载文件">
            <Button
              type="primary"
              size="small"
              icon={<DownloadOutlined />}
              onClick={() => handleDownload(record.id, record.name)}
            >
              下载
            </Button>
          </Tooltip>
          <Tooltip title="删除文件">
            <Popconfirm
              title="确定要删除这个文件吗？"
              onConfirm={() => handleDelete(record.id)}
              okText="确定"
              cancelText="取消"
            >
              <Button
                type="primary"
                danger
                size="small"
                icon={<DeleteOutlined />}
              >
                删除
              </Button>
            </Popconfirm>
          </Tooltip>
        </Space>
      ),
    },
  ];

  const rowSelection = {
    selectedRowKeys,
    onChange: (selectedRowKeys: React.Key[], selectedRows: FileInfo[]) => {
      setSelectedRowKeys(selectedRowKeys as string[]);
    },
    getCheckboxProps: (record: FileInfo) => ({
      disabled: false,
      name: record.id,
    }),
  };

  const hasSelected = selectedRowKeys.length > 0;

  return (
    <div style={{ padding: "24px" }}>
      {/* 统计信息 */}
      {stats && (
        <Row gutter={16} style={{ marginBottom: "24px" }}>
          <Col span={6}>
            <Card>
              <Statistic
                title="总文件数"
                value={stats.totalFiles}
                prefix={<InboxOutlined />}
                valueStyle={{ color: "#3f8600" }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="总存储空间"
                value={stats.totalSizeMB}
                suffix="MB"
                prefix={<CloudUploadOutlined />}
                valueStyle={{ color: "#cf1322" }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="文档数量"
                value={stats.typeStats?.["文档"] || 0}
                prefix={<FileTextOutlined />}
                valueStyle={{ color: "#1890ff" }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="图片数量"
                value={stats.typeStats?.["图片"] || 0}
                prefix={<PictureOutlined />}
                valueStyle={{ color: "#722ed1" }}
              />
            </Card>
          </Col>
        </Row>
      )}

      {/* 操作栏 */}
      <Card style={{ marginBottom: "24px" }}>
        <Row justify="space-between" align="middle">
          <Col>
            <Space>
              <Button
                type="primary"
                icon={<UploadOutlined />}
                onClick={() => setUploadModalVisible(true)}
              >
                上传文件
              </Button>
              {hasSelected && (
                <Popconfirm
                  title={`确定要删除选中的 ${selectedRowKeys.length} 个文件吗？`}
                  onConfirm={handleBatchDelete}
                  okText="确定"
                  cancelText="取消"
                >
                  <Button type="primary" danger icon={<DeleteOutlined />}>
                    批量删除 ({selectedRowKeys.length})
                  </Button>
                </Popconfirm>
              )}
            </Space>
          </Col>
          <Col>
            <Space>
              <Select
                placeholder="选择文件分类"
                allowClear
                style={{ width: 150 }}
                value={selectedCategory}
                onChange={handleCategoryChange}
              >
                <Option value="文档">文档</Option>
                <Option value="图片">图片</Option>
                <Option value="表格">表格</Option>
                <Option value="其他">其他</Option>
              </Select>
              <Button icon={<CloudDownloadOutlined />} onClick={fetchStats}>
                刷新统计
              </Button>
            </Space>
          </Col>
        </Row>
      </Card>

      {/* 文件列表 */}
      <Card title="文件列表">
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
            showTotal: (total, range) =>
              `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
            onChange: handleTableChange,
            onShowSizeChange: handleTableChange,
          }}
        />
      </Card>

      {/* 上传模态框 */}
      <Modal
        title="上传文件"
        open={uploadModalVisible}
        onCancel={() => {
          setUploadModalVisible(false);
          uploadForm.resetFields();
        }}
        footer={null}
        width={600}
      >
        <Form
          form={uploadForm}
          layout="vertical"
          onFinish={(values) => {
            const file = values.file?.fileList?.[0]?.originFileObj;
            if (file) {
              handleUpload(
                file,
                values.category,
                values.description,
                values.customName,
              );
            }
          }}
        >
          <Form.Item
            name="file"
            label="选择文件"
            rules={[{ required: true, message: "请选择要上传的文件" }]}
          >
            <Upload.Dragger
              maxCount={1}
              beforeUpload={() => false}
              accept=".pdf,.ppt,.pptx,.doc,.docx,.xls,.xlsx,.jpg,.jpeg,.png,.gif,.txt"
            >
              <p className="ant-upload-drag-icon">
                <InboxOutlined />
              </p>
              <p className="ant-upload-text">点击或拖拽文件到此区域上传</p>
              <p className="ant-upload-hint">
                支持单个文件上传，文件大小不超过100MB
                <br />
                支持的文件类型：PDF、PPT、Word、Excel、图片、文本文件
              </p>
            </Upload.Dragger>
          </Form.Item>

          <Form.Item
            name="customName"
            label="自定义文件名"
            rules={[{ required: false, message: "请输入自定义文件名" }]}
            extra="留空将使用原始文件名"
          >
            <Input
              placeholder="请输入自定义文件名（可包含扩展名）"
              onBlur={(e) => {
                const file = uploadForm.getFieldValue("file")?.fileList?.[0];
                if (file && e.target.value) {
                  const originalName = file.originFileObj.name;
                  const extension = originalName.includes(".")
                    ? originalName.substring(originalName.lastIndexOf("."))
                    : "";
                  const currentName = e.target.value;

                  // 只有在没有扩展名时才添加，避免重复添加
                  if (extension && !currentName.includes(".")) {
                    uploadForm.setFieldsValue({
                      customName: currentName + extension,
                    });
                  }
                }
              }}
            />
          </Form.Item>

          <Form.Item
            name="category"
            label="文件分类"
            rules={[{ required: true, message: "请选择文件分类" }]}
          >
            <Select placeholder="请选择文件分类">
              <Option value="文档">文档</Option>
              <Option value="图片">图片</Option>
              <Option value="表格">表格</Option>
              <Option value="其他">其他</Option>
            </Select>
          </Form.Item>

          <Form.Item name="description" label="文件描述">
            <TextArea rows={3} placeholder="请输入文件描述（可选）" />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                loading={uploading}
                icon={<UploadOutlined />}
              >
                {uploading ? "上传中..." : "开始上传"}
              </Button>
              <Button
                onClick={() => {
                  setUploadModalVisible(false);
                  uploadForm.resetFields();
                }}
              >
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default FileManagement;
