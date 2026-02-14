import { get, post, put, del } from './api'

// ============================================================================
// 类型定义
// ============================================================================

export interface LockStatus {
  document_id: number
  document_name?: string
  is_locked: boolean
  locked_by?: number
  locked_by_name?: string
  locked_at?: string
  expires_at?: string
  is_checked_out?: boolean
  can_edit: boolean
  reason?: string
}

export interface DocumentVersion {
  id: number
  document_id: number
  version: number
  filename: string
  filepath: string
  file_hash: string
  file_size: number
  mime_type: string
  created_by: number
  created_by_name?: string
  change_description: string
  change_type: string
  is_current: boolean
  created_at: string
}

export interface VersionListResponse {
  data: DocumentVersion[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface VersionComparison {
  document_id: number
  from_version: {
    version: number
    filename: string
    created_at: string
  }
  to_version: {
    version: number
    filename: string
    created_at: string
  }
  changes: VersionChange[]
  summary: string
  compared_at: string
}

export interface VersionChange {
  type: string
  path: string
  old_content?: any
  new_content?: any
  description: string
}

export interface OnlyOfficeConfig {
  document: {
    key: string
    url: string
    title: string
    fileType: string
    info: {
      author: string
      created: string
      owner: string
      uploaded: string
    }
    permissions: {
      comment: boolean
      download: boolean
      edit: boolean
      fillForms: boolean
      modifyFilter: boolean
      modifyContentControl: boolean
      review: boolean
    }
  }
  documentType: 'word' | 'cell' | 'slide'
  editorConfig: {
    mode: 'edit' | 'view'
    callbackUrl: string
    customization: {
      about: boolean
      comments: boolean
      customer: Record<string, string>
      feedback: boolean
      forcesave: boolean
      help: boolean
      macro: boolean
      metrics: boolean
      plugins: boolean
      compactHeader: boolean
      compactToolbar: boolean
      zoom: number
    }
    user: {
      id: string
      name: string
      group: string
    }
  }
}

// ============================================================================
// API 函数
// ============================================================================

/**
 * 获取文档锁状态
 */
export const getDocumentLockStatus = (documentId: number): Promise<{ data: LockStatus }> => {
  return get(`/documents/${documentId}/lock`)
}

/**
 * 获取文档锁
 */
export const acquireDocumentLock = (
  documentId: number,
  options?: { is_checkout?: boolean }
): Promise<{ data: LockStatus }> => {
  return post(`/documents/${documentId}/lock`, options || {})
}

/**
 * 释放文档锁
 */
export const releaseDocumentLock = (
  documentId: number,
  options?: { force?: boolean }
): Promise<{ data: { message: string } }> => {
  return del(`/documents/${documentId}/lock`, { body: JSON.stringify(options || {}) })
}

/**
 * 续期文档锁
 */
export const renewDocumentLock = (documentId: number): Promise<{ data: LockStatus }> => {
  return put(`/documents/${documentId}/lock`)
}

/**
 * 获取用户持有的所有锁
 */
export const getUserLocks = (): Promise<{ data: LockStatus[]; count: number }> => {
  return get('/documents/locks/user')
}

/**
 * 获取文档的所有版本
 */
export const getDocumentVersions = (
  documentId: number,
  page = 1,
  pageSize = 20
): Promise<VersionListResponse> => {
  return get(`/documents/${documentId}/versions`, { page, page_size: pageSize })
}

/**
 * 获取文档的当前版本
 */
export const getCurrentDocumentVersion = (documentId: number): Promise<{ data: DocumentVersion }> => {
  return get(`/documents/${documentId}/versions/current`)
}

/**
 * 获取指定版本
 */
export const getDocumentVersion = (
  documentId: number,
  version: number
): Promise<{ data: DocumentVersion }> => {
  return get(`/documents/${documentId}/versions/${version}`)
}

/**
 * 创建新版本
 */
export const createDocumentVersion = (
  documentId: number,
  data: {
    change_description: string
    change_type?: string
  }
): Promise<{ data: DocumentVersion }> => {
  return post(`/documents/${documentId}/versions`, data)
}

/**
 * 恢复到指定版本
 */
export const restoreDocumentVersion = (
  documentId: number,
  version: number
): Promise<{ data: { message: string } }> => {
  return post(`/documents/${documentId}/versions/${version}/restore`, {})
}

/**
 * 比较两个版本
 */
export const compareDocumentVersions = (
  documentId: number,
  fromVersion: number,
  toVersion: number
): Promise<{ data: VersionComparison }> => {
  return get(`/documents/${documentId}/versions/compare`, {
    from: fromVersion,
    to: toVersion,
  })
}

/**
 * 删除版本
 */
export const deleteDocumentVersion = (
  documentId: number,
  version: number
): Promise<{ data: { message: string } }> => {
  return del(`/documents/${documentId}/versions/${version}`)
}

/**
 * 打开 OnlyOffice 编辑器
 */
export const openOnlyOfficeEditor = (
  documentId: number,
  mode: 'edit' | 'view' = 'edit'
): Promise<{ data: OnlyOfficeConfig }> => {
  return post('/documents/onlyoffice/open', { document_id: documentId, mode })
}

/**
 * 转换文档格式
 */
export const convertDocument = (
  documentId: number,
  outputType: string
): Promise<{ data: { task_id: string } }> => {
  return post('/documents/onlyoffice/convert', { document_id: documentId, output_type: outputType })
}

/**
 * 检查转换状态
 */
export const checkConversionStatus = (taskId: string): Promise<{ data: { status: string; url?: string } }> => {
  return get('/documents/onlyoffice/convert/status', { task_id: taskId })
}

// ============================================================================
// 导出
// ============================================================================

export default {
  // 锁相关
  getDocumentLockStatus,
  acquireDocumentLock,
  releaseDocumentLock,
  renewDocumentLock,
  getUserLocks,
  // 版本相关
  getDocumentVersions,
  getCurrentDocumentVersion,
  getDocumentVersion,
  createDocumentVersion,
  restoreDocumentVersion,
  compareDocumentVersions,
  deleteDocumentVersion,
  // OnlyOffice 相关
  openOnlyOfficeEditor,
  convertDocument,
  checkConversionStatus,
}
