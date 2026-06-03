import { get, post, put, del } from './http'

// 角色接口
export interface Role {
  id: number
  name: string
  code: string
  description: string
  status: 'active' | 'inactive'
  sort_order: number
  created_at: string
  updated_at: string
}

// 权限接口
export interface Permission {
  id: number
  name: string
  code: string
  type: 'menu' | 'button' | 'api'
  parent_id: number | null
  path: string
  icon: string
  component: string
  sort_order: number
  status: 'active' | 'inactive'
  created_at: string
  updated_at: string
  children?: Permission[]
}

// 用户角色关联接口
export interface UserRole {
  id: number
  user_id: number
  role_id: number
  created_at: string
  role?: Role
}

// 角色权限关联接口
export interface RolePermission {
  id: number
  role_id: number
  permission_id: number
  created_at: string
  permission?: Permission
}

// 角色查询参数
export interface RoleQueryParams {
  name?: string
  code?: string
  status?: string
  page?: number
  page_num?: number
  page_size?: number
}

// 角色分页响应
export interface RolePageResponse {
  list: Role[]
  total: number
  page_num: number
  page_size: number
}

// 权限查询参数
export interface PermissionQueryParams {
  name?: string
  code?: string
  type?: string
  status?: string
}

// ============ 角色管理API ============

/**
 * 获取角色列表
 * @param params 查询参数
 * @returns 角色列表
 */
const normalizeRoleQuery = (params?: RoleQueryParams) => {
  if (!params) {
    return params
  }
  const { page_num, ...rest } = params
  return {
    ...rest,
    page: params.page ?? page_num,
  }
}

export const getRoleList = (params?: RoleQueryParams): Promise<RolePageResponse> => {
  return get<RolePageResponse>('/admin/roles', normalizeRoleQuery(params))
}

/**
 * 获取所有角色（不分页）
 * @returns 角色列表
 */
export const getAllRoles = (): Promise<Role[]> => {
  return get<Role[]>('/admin/roles/all')
}

/**
 * 获取角色详情
 * @param id 角色ID
 * @returns 角色详情
 */
export const getRoleById = (id: number): Promise<Role> => {
  return get<Role>(`/admin/roles/${id}`)
}

/**
 * 创建角色
 * @param data 角色数据
 * @returns 创建的角色
 */
export const createRole = (data: Partial<Role>): Promise<Role> => {
  return post<Role>('/admin/roles', data)
}

/**
 * 更新角色
 * @param id 角色ID
 * @param data 角色数据
 * @returns 更新的角色
 */
export const updateRole = (id: number, data: Partial<Role>): Promise<Role> => {
  return put<Role>(`/admin/roles/${id}`, data)
}

/**
 * 删除角色
 * @param id 角色ID
 * @returns 删除结果
 */
export const deleteRole = (id: number): Promise<void> => {
  return del<void>(`/admin/roles/${id}`)
}

/**
 * 更新角色状态
 * @param id 角色ID
 * @param status 状态
 * @returns 更新结果
 */
export const updateRoleStatus = (id: number, status: string): Promise<void> => {
  return put<void>(`/admin/roles/${id}/status`, { status })
}

// ============ 权限管理API ============

/**
 * 获取权限列表
 * @param params 查询参数
 * @returns 权限列表（树形结构）
 */
export const getPermissionList = (params?: PermissionQueryParams): Promise<Permission[]> => {
  return get<Permission[]>('/admin/permissions', params)
}

/**
 * 获取所有权限（扁平结构）
 * @returns 权限列表
 */
export const getAllPermissions = (): Promise<Permission[]> => {
  return get<Permission[]>('/admin/permissions/all')
}

/**
 * 获取权限详情
 * @param id 权限ID
 * @returns 权限详情
 */
export const getPermissionById = (id: number): Promise<Permission> => {
  return get<Permission>(`/admin/permissions/${id}`)
}

/**
 * 创建权限
 * @param data 权限数据
 * @returns 创建的权限
 */
export const createPermission = (data: Partial<Permission>): Promise<Permission> => {
  return post<Permission>('/admin/permissions', data)
}

/**
 * 更新权限
 * @param id 权限ID
 * @param data 权限数据
 * @returns 更新的权限
 */
export const updatePermission = (id: number, data: Partial<Permission>): Promise<Permission> => {
  return put<Permission>(`/admin/permissions/${id}`, data)
}

/**
 * 删除权限
 * @param id 权限ID
 * @returns 删除结果
 */
export const deletePermission = (id: number): Promise<void> => {
  return del<void>(`/admin/permissions/${id}`)
}

// ============ 角色权限关联API ============

/**
 * 获取角色的权限列表
 * @param roleId 角色ID
 * @returns 权限ID列表
 */
export const getRolePermissions = (roleId: number): Promise<number[]> => {
  return get<number[]>(`/admin/roles/${roleId}/permissions`)
}

/**
 * 为角色分配权限
 * @param roleId 角色ID
 * @param permissionIds 权限ID列表
 * @returns 分配结果
 */
export const assignRolePermissions = (roleId: number, permissionIds: number[]): Promise<void> => {
  return post<void>(`/admin/roles/${roleId}/permissions`, { permission_ids: permissionIds })
}

// ============ 用户角色关联API ============

/**
 * 获取用户的角色列表
 * @param userId 用户ID
 * @returns 角色列表
 */
export const getUserRoles = (userId: number): Promise<Role[]> => {
  return get<Role[]>(`/admin/users/${userId}/roles`)
}

/**
 * 为用户分配角色
 * @param userId 用户ID
 * @param roleIds 角色ID列表
 * @returns 分配结果
 */
export const assignUserRoles = (userId: number, roleIds: number[]): Promise<void> => {
  return post<void>(`/admin/users/${userId}/roles`, { role_ids: roleIds })
}

/**
 * 获取当前用户的权限列表
 * @returns 权限列表
 */
export const getCurrentUserPermissions = (): Promise<Permission[]> => {
  return get<Permission[]>('/admin/current-user/permissions')
}

/**
 * 获取当前用户的角色列表
 * @returns 角色列表
 */
export const getCurrentUserRoles = (): Promise<Role[]> => {
  return get<Role[]>('/admin/current-user/roles')
}

// 角色服务统一导出
export const roleService = {
  getRoleList,
  getAllRoles,
  getRoleById,
  createRole,
  updateRole,
  deleteRole,
  updateRoleStatus,
  getPermissionList,
  getAllPermissions,
  getPermissionById,
  createPermission,
  updatePermission,
  deletePermission,
  getRolePermissions,
  assignRolePermissions,
  getUserRoles,
  assignUserRoles,
  getCurrentUserPermissions,
  getCurrentUserRoles,
}
