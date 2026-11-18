import { get, post, put, del } from './http'

export interface Lawyer {
  lawyerId?: number
  lawyerName: string
  phone: string
  email: string
  licenseNo: string
  specialty: string
  department: string
  position: string
  delFlag: string
  // 前端需要的额外字段
  id?: number
  name?: string
  licenseNumber?: string
  gender?: 'male' | 'female'
  experience?: number
  status?: 'active' | 'inactive' | 'on_leave'
  joinDate?: string
  profile?: string
  avatar?: string
}

export interface LawyerStats {
  total: number
  active: number
  inactive: number
  onLeave: number
  departmentStats: Record<string, number>
  specialtyStats: Record<string, number>
}

const normalizeStatus = (status?: string): 'active' | 'inactive' | 'on_leave' => {
  if (status === 'inactive' || status === 'on_leave') {
    return status
  }
  return 'active'
}

const normalizeLawyer = (raw: any): Lawyer => {
  const id = raw.lawyerId ?? raw.lawyer_id ?? raw.id
  const name = raw.lawyerName ?? raw.lawyer_name ?? raw.name ?? raw.username ?? ''
  const specialty = Array.isArray(raw.specialty)
    ? raw.specialty.join(',')
    : raw.specialty || raw.seniority || '综合法律服务'

  return {
    lawyerId: id,
    lawyerName: name,
    phone: raw.phone || '',
    email: raw.email || '',
    licenseNo: raw.licenseNo ?? raw.license_no ?? raw.licenseNumber ?? raw.license_number ?? '',
    specialty,
    department: raw.department || '综合部',
    position: raw.position || raw.seniority || '执业律师',
    delFlag: raw.delFlag ?? raw.del_flag ?? '0',
    id,
    name,
    licenseNumber: raw.licenseNumber ?? raw.license_number ?? raw.licenseNo ?? raw.license_no ?? '',
    gender: raw.gender === 'female' ? 'female' : 'male',
    experience: Number(raw.experience ?? 0),
    status: normalizeStatus(raw.status),
    joinDate: raw.joinDate ?? raw.join_date ?? raw.created_at?.slice?.(0, 10) ?? '',
    profile: raw.profile || `${name || '律师'}，负责律所法律服务工作。`,
    avatar: raw.avatar || '',
  }
}

/**
 * 获取律师列表
 */
export const getLawyerList = (params?: any): Promise<{ list: Lawyer[]; total: number }> => {
  return get<any>('/lawfirm/lawyers', params).then((response) => {
    const list = Array.isArray(response)
      ? response
      : response?.list || response?.data?.list || response?.data || []
    const total = response?.total ?? response?.pagination?.total ?? list.length

    return {
      list: list.map(normalizeLawyer),
      total,
    }
  })
}

/**
 * 获取律师详情
 */
export const getLawyerDetail = (id: number): Promise<Lawyer> => {
  return get<any>(`/lawfirm/lawyers/${id}`).then(normalizeLawyer)
}

/**
 * 新增律师
 */
export const addLawyer = (data: Lawyer): Promise<Lawyer> => {
  return post<Lawyer>('/lawfirm/lawyers', data)
}

// 兼容前端页面调用的方法名
export const createLawyer = (data: Lawyer): Promise<Lawyer> => {
  return post<Lawyer>('/lawfirm/lawyers', data)
}

/**
 * 更新律师信息
 */
export const updateLawyer = (id: number, data: Lawyer): Promise<Lawyer> => {
  return put<Lawyer>(`/lawfirm/lawyers/${id}`, data)
}

/**
 * 删除律师
 */
export const deleteLawyer = (id: number): Promise<void> => {
  return del<void>(`/lawfirm/lawyers/${id}`)
}

/**
 * 获取律师统计信息
 */
export const getLawyerStats = (): Promise<LawyerStats> => {
  return get<LawyerStats>('/lawfirm/lawyers/stats')
}

/**
 * 更新律师状态
 */
export const updateLawyerStatus = (id: number, status: string): Promise<void> => {
  return put<void>(`/lawfirm/lawyers/${id}/status`, { status })
}

// 导出所有服务函数作为lawyerService对象
export const lawyerService = {
  getLawyerList,
  getLawyerDetail,
  addLawyer,
  createLawyer,
  updateLawyer,
  deleteLawyer,
  getLawyerStats,
  updateLawyerStatus,
}
