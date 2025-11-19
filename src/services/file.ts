import { get, post, del } from '@/services/http'

export interface FileInfo {
  id: string
  name: string
  originalName?: string
  size: number
  contentType: string
  category: string
  description: string
  uploadTime: string
  uploadPath: string
  type?: string
  path?: string
  url?: string
  lastModified?: number
}

export interface FileStats {
  totalFiles: number
  totalSize: number
  totalSizeMB: number
  typeStats: Record<string, number>
}

export interface FileListResponse {
  total: number
  rows: FileInfo[]
}

export interface UploadResponse {
  id: string
  name: string
  uniqueName: string
  originalName: string
  size: number
  contentType: string
  category: string
  description: string
  uploadTime: string
  uploadPath: string
}

export interface BatchDeleteResponse {
  successCount: number
  failedCount: number
  successFiles: string[]
  failedFiles: string[]
}

/**
 * 上传文件
 */
export const uploadFile = (
  file: File,
  category?: string,
  description?: string,
  customName?: string,
): Promise<UploadResponse> => {
  const formData = new FormData()
  formData.append('file', file)
  if (category) {
    formData.append('category', category)
  }
  if (description) {
    formData.append('description', description)
  }
  if (customName) {
    formData.append('customName', customName)
  }

  return post<UploadResponse>('/file/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  })
}

/**
 * 获取文件列表
 */
export const getFileList = (params?: {
  category?: string
  pageNum?: number
  pageSize?: number
}): Promise<FileListResponse> => {
  return get<FileListResponse>('/file/list', { params })
}

/**
 * 下载文件
 */
export const downloadFile = (uniqueName: string, displayName?: string): void => {
  const link = document.createElement('a')
  link.href = `/file/download/${uniqueName}`
  link.download = displayName || uniqueName
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

/**
 * 删除文件
 */
export const deleteFile = (fileName: string): Promise<void> => {
  return del<void>(`/file/delete/${fileName}`)
}

/**
 * 获取文件统计信息
 */
export const getFileStats = (): Promise<FileStats> => {
  return get<FileStats>('/file/stats')
}

/**
 * 批量删除文件
 */
export const batchDeleteFiles = (fileNames: string[]): Promise<BatchDeleteResponse> => {
  return del<BatchDeleteResponse>('/file/batch-delete', { data: fileNames })
}

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
 * 获取文件图标
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
      return '#f5222d'
    case 'PPT':
      return '#fa8c16'
    case 'Word':
      return '#1890ff'
    case 'Excel':
      return '#52c41a'
    case '图片':
      return '#722ed1'
    default:
      return '#8c8c8c'
  }
}
