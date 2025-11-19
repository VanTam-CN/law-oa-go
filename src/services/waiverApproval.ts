import { api } from '@/utils/request'
import type {
  WaiverApplication,
  WaiverStatus,
  WaiverType,
  RiskLevel,
  CreateWaiverApplicationRequest,
  UpdateWaiverApplicationRequest,
  SubmitWaiverApplicationRequest,
  ApproveWaiverApplicationRequest,
  RejectWaiverApplicationRequest,
  EscalateWaiverApplicationRequest,
  CreateMonitoringTaskRequest,
  WaiverApplicationQueryParams,
  WaiverApiResponse,
  PaginatedResponse,
  WaiverStatistics,
  WaiverApprovalRecord,
  WaiverMonitoringTask,
  WaiverTemplate,
  ApprovalHistory,
  Stakeholder,
  WaiverAttachment,
} from '@/types/waiverApproval'

/**
 * 豁免审批API服务
 */
class WaiverApprovalService {
  private baseUrl = '/api/v1/waiver-approval'

  // ========== 豁免申请相关 ==========

  /**
   * 创建豁免申请
   */
  async createWaiverApplication(
    request: CreateWaiverApplicationRequest,
  ): Promise<WaiverApiResponse<WaiverApplication>> {
    try {
      const formData = new FormData()

      // 添加基础字段
      formData.append('caseId', request.caseId)
      formData.append('waiverType', request.waiverType)
      formData.append('description', request.description)
      formData.append('justification', request.justification)
      formData.append('mitigationMeasures', request.mitigationMeasures)
      formData.append('affectedParties', JSON.stringify(request.affectedParties))
      formData.append('stakeholders', JSON.stringify(request.stakeholders))

      if (request.submitImmediately !== undefined) {
        formData.append('submitImmediately', request.submitImmediately.toString())
      }

      // 添加附件
      if (request.attachments && request.attachments.length > 0) {
        request.attachments.forEach((file, index) => {
          formData.append(`attachment_${index}`, file)
        })
      }

      const response = await api.post(`${this.baseUrl}/applications`, formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })

      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '创建豁免申请失败')
    }
  }

  /**
   * 获取豁免申请列表
   */
  async getWaiverApplications(
    params?: WaiverApplicationQueryParams,
  ): Promise<WaiverApiResponse<PaginatedResponse<WaiverApplication>>> {
    try {
      const response = await api.get(`${this.baseUrl}/applications`, { params })
      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '获取豁免申请列表失败')
    }
  }

  /**
   * 获取豁免申请详情
   */
  async getWaiverApplication(id: string): Promise<WaiverApiResponse<WaiverApplication>> {
    try {
      const response = await api.get(`${this.baseUrl}/applications/${id}`)
      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '获取豁免申请详情失败')
    }
  }

  /**
   * 更新豁免申请
   */
  async updateWaiverApplication(
    request: UpdateWaiverApplicationRequest,
  ): Promise<WaiverApiResponse<WaiverApplication>> {
    try {
      const formData = new FormData()
      formData.append('id', request.id)

      if (request.description !== undefined) {
        formData.append('description', request.description)
      }
      if (request.justification !== undefined) {
        formData.append('justification', request.justification)
      }
      if (request.mitigationMeasures !== undefined) {
        formData.append('mitigationMeasures', request.mitigationMeasures)
      }
      if (request.affectedParties !== undefined) {
        formData.append('affectedParties', JSON.stringify(request.affectedParties))
      }
      if (request.stakeholders !== undefined) {
        formData.append('stakeholders', JSON.stringify(request.stakeholders))
      }

      // 添加附件
      if (request.attachments && request.attachments.length > 0) {
        request.attachments.forEach((file, index) => {
          formData.append(`attachment_${index}`, file)
        })
      }

      const response = await api.put(`${this.baseUrl}/applications/${request.id}`, formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })

      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '更新豁免申请失败')
    }
  }

  /**
   * 提交豁免申请
   */
  async submitWaiverApplication(
    request: SubmitWaiverApplicationRequest,
  ): Promise<WaiverApiResponse<WaiverApplication>> {
    try {
      const response = await api.post(`${this.baseUrl}/applications/${request.id}/submit`, request)
      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '提交豁免申请失败')
    }
  }

  /**
   * 删除豁免申请
   */
  async deleteWaiverApplication(id: string): Promise<WaiverApiResponse<void>> {
    try {
      const response = await api.delete(`${this.baseUrl}/applications/${id}`)
      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '删除豁免申请失败')
    }
  }

  // ========== 审批相关 ==========

  /**
   * 批准豁免申请
   */
  async approveWaiverApplication(
    request: ApproveWaiverApplicationRequest,
  ): Promise<WaiverApiResponse<WaiverApprovalRecord>> {
    try {
      const response = await api.post(
        `${this.baseUrl}/applications/${request.applicationId}/approve`,
        request,
      )
      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '批准豁免申请失败')
    }
  }

  /**
   * 拒绝豁免申请
   */
  async rejectWaiverApplication(
    request: RejectWaiverApplicationRequest,
  ): Promise<WaiverApiResponse<WaiverApprovalRecord>> {
    try {
      const response = await api.post(
        `${this.baseUrl}/applications/${request.applicationId}/reject`,
        request,
      )
      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '拒绝豁免申请失败')
    }
  }

  /**
   * 上报豁免申请
   */
  async escalateWaiverApplication(
    request: EscalateWaiverApplicationRequest,
  ): Promise<WaiverApiResponse<WaiverApprovalRecord>> {
    try {
      const response = await api.post(
        `${this.baseUrl}/applications/${request.applicationId}/escalate`,
        request,
      )
      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '上报豁免申请失败')
    }
  }

  /**
   * 获取审批历史记录
   */
  async getApprovalHistory(
    applicationId: string,
  ): Promise<WaiverApiResponse<WaiverApprovalRecord[]>> {
    try {
      const response = await api.get(
        `${this.baseUrl}/applications/${applicationId}/approval-history`,
      )
      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '获取审批历史记录失败')
    }
  }

  /**
   * 获取待审批申请列表
   */
  async getPendingApplications(params?: {
    page?: number
    pageSize?: number
    riskLevel?: RiskLevel[]
    waiverType?: WaiverType[]
  }): Promise<WaiverApiResponse<PaginatedResponse<WaiverApplication>>> {
    try {
      const response = await api.get(`${this.baseUrl}/pending-applications`, { params })
      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '获取待审批申请列表失败')
    }
  }

  /**
   * 获取我的审批历史
   */
  async getMyApprovalHistory(params?: {
    page?: number
    pageSize?: number
    decision?: 'APPROVE' | 'REJECT' | 'ESCALATE'
    dateFrom?: string
    dateTo?: string
  }): Promise<WaiverApiResponse<PaginatedResponse<ApprovalHistory>>> {
    try {
      const response = await api.get(`${this.baseUrl}/my-approval-history`, { params })
      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '获取我的审批历史失败')
    }
  }

  // ========== 监控相关 ==========

  /**
   * 创建监控任务
   */
  async createMonitoringTask(
    request: CreateMonitoringTaskRequest,
  ): Promise<WaiverApiResponse<WaiverMonitoringTask>> {
    try {
      const response = await api.post(`${this.baseUrl}/monitoring-tasks`, request)
      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '创建监控任务失败')
    }
  }

  /**
   * 获取监控任务列表
   */
  async getMonitoringTasks(
    applicationId?: string,
  ): Promise<WaiverApiResponse<WaiverMonitoringTask[]>> {
    try {
      const params = applicationId ? { applicationId } : {}
      const response = await api.get(`${this.baseUrl}/monitoring-tasks`, { params })
      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '获取监控任务列表失败')
    }
  }

  /**
   * 更新监控任务状态
   */
  async updateMonitoringTaskStatus(
    taskId: string,
    status: 'COMPLETED' | 'PARTIALLY_COMPLETED',
    notes: string,
    evidenceAttachments?: File[],
  ): Promise<WaiverApiResponse<void>> {
    try {
      const formData = new FormData()
      formData.append('status', status)
      formData.append('notes', notes)

      if (evidenceAttachments && evidenceAttachments.length > 0) {
        evidenceAttachments.forEach((file, index) => {
          formData.append(`evidence_${index}`, file)
        })
      }

      const response = await api.put(
        `${this.baseUrl}/monitoring-tasks/${taskId}/status`,
        formData,
        {
          headers: {
            'Content-Type': 'multipart/form-data',
          },
        },
      )

      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '更新监控任务状态失败')
    }
  }

  // ========== 统计和分析 ==========

  /**
   * 获取豁免统计数据
   */
  async getWaiverStatistics(params?: {
    dateFrom?: string
    dateTo?: string
    groupBy?: 'month' | 'quarter' | 'year'
  }): Promise<WaiverApiResponse<WaiverStatistics>> {
    try {
      const response = await api.get(`${this.baseUrl}/statistics`, { params })
      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '获取豁免统计数据失败')
    }
  }

  /**
   * 获取豁免申请状态分布
   */
  async getStatusDistribution(params?: {
    dateFrom?: string
    dateTo?: string
  }): Promise<WaiverApiResponse<Record<WaiverStatus, number>>> {
    try {
      const response = await api.get(`${this.baseUrl}/statistics/status-distribution`, { params })
      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '获取状态分布数据失败')
    }
  }

  /**
   * 获取风险等级分布
   */
  async getRiskLevelDistribution(params?: {
    dateFrom?: string
    dateTo?: string
  }): Promise<WaiverApiResponse<Record<RiskLevel, number>>> {
    try {
      const response = await api.get(`${this.baseUrl}/statistics/risk-level-distribution`, {
        params,
      })
      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '获取风险等级分布数据失败')
    }
  }

  /**
   * 获取平均审批时间统计
   */
  async getAverageApprovalTime(params?: {
    dateFrom?: string
    dateTo?: string
    groupBy?: 'month' | 'quarter' | 'year'
  }): Promise<WaiverApiResponse<Array<{ period: string; averageDays: number }>>> {
    try {
      const response = await api.get(`${this.baseUrl}/statistics/average-approval-time`, { params })
      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '获取平均审批时间统计失败')
    }
  }

  // ========== 模板管理 ==========

  /**
   * 获取豁免模板列表
   */
  async getWaiverTemplates(params?: {
    waiverType?: WaiverType
    riskLevel?: RiskLevel
    isActive?: boolean
  }): Promise<WaiverApiResponse<WaiverTemplate[]>> {
    try {
      const response = await api.get(`${this.baseUrl}/templates`, { params })
      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '获取豁免模板列表失败')
    }
  }

  /**
   * 根据ID获取豁免模板
   */
  async getWaiverTemplate(id: string): Promise<WaiverApiResponse<WaiverTemplate>> {
    try {
      const response = await api.get(`${this.baseUrl}/templates/${id}`)
      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '获取豁免模板失败')
    }
  }

  /**
   * 基于模板创建豁免申请
   */
  async createApplicationFromTemplate(
    templateId: string,
    variables: Record<string, any>,
    additionalData: Partial<CreateWaiverApplicationRequest>,
  ): Promise<WaiverApiResponse<WaiverApplication>> {
    try {
      const response = await api.post(
        `${this.baseUrl}/templates/${templateId}/create-application`,
        {
          variables,
          ...additionalData,
        },
      )
      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '基于模板创建豁免申请失败')
    }
  }

  // ========== 附件管理 ==========

  /**
   * 下载豁免附件
   */
  async downloadAttachment(attachmentId: string): Promise<Blob> {
    try {
      const response = await api.get(`${this.baseUrl}/attachments/${attachmentId}/download`, {
        responseType: 'blob',
      })
      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '下载附件失败')
    }
  }

  /**
   * 删除豁免附件
   */
  async deleteAttachment(attachmentId: string): Promise<WaiverApiResponse<void>> {
    try {
      const response = await api.delete(`${this.baseUrl}/attachments/${attachmentId}`)
      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '删除附件失败')
    }
  }

  // ========== 电子签名 ==========

  /**
   * 获取签名证书信息
   */
  async getSignatureCertificates(): Promise<
    WaiverApiResponse<
      Array<{
        id: string
        name: string
        issuer: string
        expiresAt: string
        isActive: boolean
      }>
    >
  > {
    try {
      const response = await api.get(`${this.baseUrl}/signature-certificates`)
      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '获取签名证书信息失败')
    }
  }

  /**
   * 验证电子签名
   */
  async verifyElectronicSignature(
    applicationId: string,
    signatureData: {
      signatureBase64: string
      certificateId: string
      timestamp: string
    },
  ): Promise<
    WaiverApiResponse<{
      isValid: boolean
      verificationDetails: {
        certificateValid: boolean
        signatureValid: boolean
        timestampValid: boolean
        verifiedAt: string
      }
    }>
  > {
    try {
      const response = await api.post(
        `${this.baseUrl}/applications/${applicationId}/verify-signature`,
        signatureData,
      )
      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '验证电子签名失败')
    }
  }

  // ========== 通知和提醒 ==========

  /**
   * 发送审批提醒
   */
  async sendApprovalReminder(
    applicationId: string,
    message?: string,
  ): Promise<WaiverApiResponse<void>> {
    try {
      const response = await api.post(
        `${this.baseUrl}/applications/${applicationId}/send-reminder`,
        {
          message,
        },
      )
      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '发送审批提醒失败')
    }
  }

  /**
   * 获取待处理提醒列表
   */
  async getPendingReminders(): Promise<
    WaiverApiResponse<
      Array<{
        id: string
        applicationId: string
        applicationTitle: string
        type: 'APPROVAL_REMINDER' | 'MONITORING_DUE' | 'EXPIRY_WARNING'
        message: string
        createdAt: string
        isRead: boolean
      }>
    >
  > {
    try {
      const response = await api.get(`${this.baseUrl}/reminders`)
      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '获取待处理提醒列表失败')
    }
  }

  /**
   * 标记提醒为已读
   */
  async markReminderAsRead(reminderId: string): Promise<WaiverApiResponse<void>> {
    try {
      const response = await api.put(`${this.baseUrl}/reminders/${reminderId}/read`)
      return response.data
    } catch (error: any) {
      throw new Error(error.response?.data?.message || '标记提醒为已读失败')
    }
  }
}

// 导出单例实例
export const waiverApprovalService = new WaiverApprovalService()
export default waiverApprovalService
