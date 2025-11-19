import { get, post, put, del } from '../services/http'
import { departmentService } from '../services/department'

export { departmentService }

// 导出具体函数以保持向后兼容
export const getDepartments = departmentService.getDepartmentList
export const getDepartment = departmentService.getDepartment
export const createDepartment = departmentService.addDepartment
export const updateDepartment = departmentService.updateDepartment
export const deleteDepartment = departmentService.deleteDepartment
