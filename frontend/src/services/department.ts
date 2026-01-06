import { get, post, put, del } from './http'

export interface Department {
  id?: number
  name: string
  code: string
  parentId?: number
  parentName?: string
  description: string
  managerId?: number
  managerName?: string
  status: 'active' | 'inactive'
  createdAt?: string
  updatedAt?: string
}

export interface DepartmentListResponse {
  total: number
  pageNum: number
  pageSize: number
  list: Department[]
}

export interface DepartmentQueryParams {
  keyword?: string
  status?: string
  pageNum?: number
  pageSize?: number
}

export const departmentService = {
  // 获取部门列表
  getDepartmentList: (params: DepartmentQueryParams) =>
    get<DepartmentListResponse>('/departments', params),

  // 获取部门详情
  getDepartment: (id: number) => get<Department>(`/departments/${id}`),

  // 新增部门
  addDepartment: (data: Department) => post<Department>('/departments', data),

  // 更新部门信息
  updateDepartment: (id: number, data: Department) => put<Department>(`/departments/${id}`, data),

  // 删除部门
  deleteDepartment: (id: number) => del(`/departments/${id}`),

  // 获取部门树形结构
  getDepartmentTree: () => get<Department[]>('/departments/tree'),
}
