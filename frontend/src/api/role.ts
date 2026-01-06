import { get, post, put, del } from '../services/http'
import { roleService } from '../services/role'

export { roleService }

// 导出具体函数以保持向后兼容
export const getRoles = roleService.getRoleList
export const getRole = roleService.getRoleById
export const createRole = roleService.createRole
export const updateRole = roleService.updateRole
export const deleteRole = roleService.deleteRole
