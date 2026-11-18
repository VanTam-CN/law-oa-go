import { get, post, del } from '@/services/http'

// =============================================================================
// 数据类型定义 - 与后端 API 保持一致
// =============================================================================

/**
 * 文档信息 - 对应后端 Document 结构
 */
export interface FileInfo {
  id: number // 后端返回的是数字 ID
  name: string // 显示名称
  filename: string // 原始文件名
  description?: string
  filesize: number
  mime_type: string
  category: string
  tags?: string[]
  entity_id?: number
  entity_type?: string
  filepath: string
  created_at: string
  updated_at: string
}

/**
 * 文档列表响应 - 对应后端返回结构
 */
export interface FileListResponse {
  documents: FileInfo[]
  total: number
  page: number
  page_size: number
}

/**
 * 文档上传响应
 */
export interface UploadResponse {
  id: number
  name: string
  filename: string
  description?: string
  filesize: number
  mime_type: string
  category: string
  filepath: string
  created_at: string
}

/**
 * 文档统计信息 - 对应后端 DashboardStatsResponse
 */
export interface FileStats {
  totalFiles: number
  totalSize: number
  totalSizeMB: number
  todayUploads?: number
  typeStats?: Record<string, number>
}

/**
 * 批量删除响应
 */
export interface BatchDeleteResponse {
  successCount: number
  failedCount: number
  successFiles: string[]
  failedFiles: string[]
}

// =============================================================================
// API 函数 - 与后端路由完全对应
// =============================================================================

/**
 * 上传文档
 * POST /api/v1/documents
 *
 * 后端参数: name, description, category, tags, entity_id, entity_type, file
 */
export const uploadFile = async (
  file: File,
  category?: string,
  description?: string,
  customName?: string,
): Promise<UploadResponse> => {
  const formData = new FormData()

  // 使用 name 字段而不是 customName
  formData.append('name', customName || file.name)
  formData.append('file', file)

  if (category) {
    formData.append('category', category)
  }
  if (description) {
    formData.append('description', description)
  }

  // 调用正确的后端 API 路径
  return post<UploadResponse>('/documents', formData)
}

/**
 * 获取文档列表
 * GET /api/v1/documents
 *
 * 后端参数: page, page_size, category, entity_type, entity_id, search, sort_by, sort_order
 */
export const getFileList = async (params?: {
  category?: string
  page?: number
  page_size?: number
  search?: string
}): Promise<FileListResponse> => {
  // 转换参数名：pageNum -> page, pageSize -> page_size
  const requestParams: Record<string, any> = {
    page: params?.page || 1,
    page_size: params?.page_size || 10,
  }

  if (params?.category) {
    requestParams.category = params.category
  }
  if (params?.search) {
    requestParams.search = params.search
  }

  // 调用正确的后端 API 路径
  const response = await get<{
    documents: FileInfo[]
    total: number
    page: number
    page_size: number
  }>('/documents', requestParams)

  return response
}

/**
 * 下载文档
 * GET /api/v1/documents/:id/download
 */
export const downloadFile = (id: number, displayName?: string): void => {
  // 使用正确的后端 API 路径
  const link = document.createElement('a')
  link.href = `/api/v1/documents/${id}/download`
  link.download = displayName || `document-${id}`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

/**
 * 删除文档
 * DELETE /api/v1/documents/:id
 *
 * 注意：后端是软删除，会将文档移至回收站
 */
export const deleteFile = async (id: number): Promise<void> => {
  // 使用正确的后端 API 路径
  return del<void>(`/documents/${id}`)
}

/**
 * 获取文档统计信息
 * GET /api/v1/documents/stats/dashboard
 */
export const getFileStats = async (): Promise<FileStats> => {
  // 使用正确的后端 API 路径
  const response = await get<{
    total_documents: number
    total_storage: number
    storage_by_category: Record<string, number>
    recent_uploads: number
  }>('/documents/stats/dashboard')

  // 转换后端响应格式到前端期望的格式
  return {
    totalFiles: response.total_documents || 0,
    totalSize: response.total_storage || 0,
    totalSizeMB: Math.round((response.total_storage || 0) / 1024 / 1024),
    todayUploads: response.recent_uploads || 0,
    typeStats: response.storage_by_category || {},
  }
}

/**
 * 批量删除文档
 *
 * 注意：后端没有批量删除接口，这里通过多次调用单个删除实现
 */
export const batchDeleteFiles = async (
  ids: number[],
): Promise<BatchDeleteResponse> => {
  const successFiles: string[] = []
  const failedFiles: string[] = []

  // 逐个删除文档
  for (const id of ids) {
    try {
      await deleteFile(id)
      successFiles.push(String(id))
    } catch (error) {
      failedFiles.push(String(id))
    }
  }

  return {
    successCount: successFiles.length,
    failedCount: failedFiles.length,
    successFiles,
    failedFiles,
  }
}

// =============================================================================
// 工具函数
// =============================================================================

/**
 * 格式化文件大小
 */
export const formatFileSize = (bytes: number): string => {
  if (bytes === 0) {
    return '0 B'
  }
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
}

/**
 * 获取文件图标 (emoji)
 */
export const getFileIcon = (fileName: string): string => {
  const extension = fileName.split('.').pop()?.toLowerCase() || ''

  switch (extension) {
    case 'pdf':
      return '📄'
    case 'ppt':
    case 'pptx':
      return '📽️'
    case 'doc':
    case 'docx':
      return '📝'
    case 'xls':
    case 'xlsx':
      return '📊'
    case 'jpg':
    case 'jpeg':
    case 'png':
    case 'gif':
      return '🖼️'
    case 'zip':
    case 'rar':
    case '7z':
      return '📦'
    case 'txt':
      return '📃'
    default:
      return '📎'
  }
}

/**
 * 获取文件类型颜色
 */
export const getFileTypeColor = (type: string): string => {
  switch (type) {
    case 'PDF':
    case 'application/pdf':
      return '#f5222d'
    case 'PPT':
    case 'PowerPoint':
      return '#fa8c16'
    case 'Word':
    case 'application/msword':
    case 'application/vnd.openxmlformats-officedocument.wordprocessingml.document':
      return '#1890ff'
    case 'Excel':
    case 'application/vnd.ms-excel':
    case 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet':
      return '#52c41a'
    case '图片':
    case 'image':
      return '#722ed1'
    default:
      return '#8c8c8c'
  }
}
