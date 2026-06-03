import { get, post, put, del } from './http'

// 部门接口
export interface Department {
  id: number
  name: string
  code: string
  description: string
}

// 员工接口
export interface Employee {
  id: number
  name: string
  position: string
  department: string
  email: string
  phone: string
  status: 'active' | 'inactive'
  join_date: string
  created_at: string
}

// 文档接口
export interface Document {
  id: number
  title: string
  type: string
  category: string
  content: string
  create_date: string
  creator: string
  status: 'draft' | 'published' | 'archived'
  created_at: string
}

// 系统设置接口
export interface SystemSettings {
  system: {
    site_name: string
    logo_url: string
    theme: string
    language: string
    timezone: string
  }
  security: {
    password_min_length: number
    session_timeout: number
    login_attempts: number
  }
  email: {
    smtp_host: string
    smtp_port: number
    smtp_username: string
    from_name: string
  }
}

/**
 * 获取部门列表
 * @returns 部门列表
 */
export const getDepartments = (): Promise<Department[]> => {
  return get<Department[]>('/admin/departments')
}

/**
 * 获取员工列表
 * @param params 查询参数
 * @returns 员工列表
 */
export const getEmployees = (params?: {
  department?: string
  status?: string
}): Promise<Employee[]> => {
  return get<Employee[]>('/admin/employees', params)
}

/**
 * 获取员工详情
 * @param id 员工ID
 * @returns 员工详情
 */
export const getEmployeeById = (id: number): Promise<Employee> => {
  return get<Employee>(`/admin/employees/${id}`)
}

/**
 * 创建员工
 * @param data 员工数据
 * @returns 创建的员工
 */
export const createEmployee = (data: Partial<Employee>): Promise<Employee> => {
  return post<Employee>('/admin/employees', data)
}

/**
 * 更新员工
 * @param id 员工ID
 * @param data 员工数据
 * @returns 更新的员工
 */
export const updateEmployee = (id: number, data: Partial<Employee>): Promise<Employee> => {
  return put<Employee>(`/admin/employees/${id}`, data)
}

/**
 * 删除员工
 * @param id 员工ID
 * @returns 删除结果
 */
export const deleteEmployee = (id: number): Promise<void> => {
  return del<void>(`/admin/employees/${id}`)
}

/**
 * 获取文档列表
 * @param params 查询参数
 * @returns 文档列表
 */
export const getDocuments = (params?: {
  type?: string
  category?: string
  status?: string
}): Promise<Document[]> => {
  return get<Document[]>('/admin/documents', params)
}

/**
 * 获取文档详情
 * @param id 文档ID
 * @returns 文档详情
 */
export const getDocumentById = (id: number): Promise<Document> => {
  return get<Document>(`/admin/documents/${id}`)
}

/**
 * 创建文档
 * @param data 文档数据
 * @returns 创建的文档
 */
export const createDocument = (data: Partial<Document>): Promise<Document> => {
  return post<Document>('/admin/documents', data)
}

/**
 * 更新文档
 * @param id 文档ID
 * @param data 文档数据
 * @returns 更新的文档
 */
export const updateDocument = (id: number, data: Partial<Document>): Promise<Document> => {
  return put<Document>(`/admin/documents/${id}`, data)
}

/**
 * 删除文档
 * @param id 文档ID
 * @returns 删除结果
 */
export const deleteDocument = (id: number): Promise<void> => {
  return del<void>(`/admin/documents/${id}`)
}

/**
 * 获取系统设置
 * @returns 系统设置
 */
export const getSettings = (): Promise<SystemSettings> => {
  return get<SystemSettings>('/admin/settings')
}

/**
 * 更新系统设置
 * @param settings 系统设置
 * @returns 更新结果
 */
export const updateSettings = (settings: Partial<SystemSettings>): Promise<void> => {
  return put<void>('/admin/settings', settings)
}
