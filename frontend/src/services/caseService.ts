import apiClient from "./api";
import {
  Case,
  CaseListRequest,
  CreateCaseRequest,
  UpdateCaseRequest,
  CaseStats,
  AssignLawyerRequest,
  UpdateCaseStatusRequest,
} from "../types";
import { AppError } from "../types/errors";

class CaseService {
  // 获取案件列表（分页）
  async getCases(params?: CaseListRequest): Promise<{
    data: Case[];
    pagination: {
      page: number;
      page_size: number;
      total: number;
      total_pages: number;
    };
  }> {
    try {
      return await apiClient.getPaginated<Case>("/cases", {
        params,
        
      });
    } catch (error: any) {
      console.error("获取案件列表失败:", error);
      throw new AppError(
        error.message || "获取案件列表失败",
        error.code || "GET_CASES_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取案件详情
  async getCase(id: number): Promise<Case> {
    try {
      return await apiClient.get<Case>(`/cases/${id}`, {
        
      });
    } catch (error: any) {
      console.error("获取案件详情失败:", error);
      throw new AppError(
        error.message || "获取案件详情失败",
        error.code || "GET_CASE_ERROR",
        error.statusCode || 404,
      );
    }
  }

  // 创建案件
  async createCase(data: CreateCaseRequest): Promise<Case> {
    try {
      return await apiClient.post<Case>("/cases", data, {
      });
    } catch (error: any) {
      console.error("创建案件失败:", error);
      throw new AppError(
        error.message || "创建案件失败",
        error.code || "CREATE_CASE_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 更新案件
  async updateCase(id: number, data: UpdateCaseRequest): Promise<Case> {
    try {
      return await apiClient.put<Case>(`/cases/${id}`, data, {
      });
    } catch (error: any) {
      console.error("更新案件失败:", error);
      throw new AppError(
        error.message || "更新案件失败",
        error.code || "UPDATE_CASE_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 删除案件
  async deleteCase(id: number): Promise<{ message: string }> {
    try {
      return await apiClient.delete<{ message: string }>(`/cases/${id}`, {
      });
    } catch (error: any) {
      console.error("删除案件失败:", error);
      throw new AppError(
        error.message || "删除案件失败",
        error.code || "DELETE_CASE_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 获取案件统计信息
  async getCaseStats(): Promise<CaseStats> {
    try {
      return await apiClient.get<CaseStats>("/cases/stats", {
        
      });
    } catch (error: any) {
      console.error("获取案件统计信息失败:", error);
      throw new AppError(
        error.message || "获取案件统计信息失败",
        error.code || "GET_CASE_STATS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 分配律师
  async assignLawyer(
    caseId: number,
    lawyer_id: number,
  ): Promise<{ message: string }> {
    try {
      const data: AssignLawyerRequest = { lawyer_id: lawyer_id };
      return await apiClient.post<{ message: string }>(
        `/cases/${caseId}/assign`,
        data,
        {
        },
      );
    } catch (error: any) {
      console.error("分配律师失败:", error);
      throw new AppError(
        error.message || "分配律师失败",
        error.code || "ASSIGN_LAWYER_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 更新案件状态
  async updateCaseStatus(
    caseId: number,
    status: string,
  ): Promise<{ message: string }> {
    try {
      const data: UpdateCaseStatusRequest = { status };
      return await apiClient.post<{ message: string }>(
        `/cases/${caseId}/status`,
        data,
        {
        },
      );
    } catch (error: any) {
      console.error("更新案件状态失败:", error);
      throw new AppError(
        error.message || "更新案件状态失败",
        error.code || "UPDATE_CASE_STATUS_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 获取我的案件
  async getMyCases(params?: Omit<CaseListRequest, "lawyer_id">): Promise<{
    data: Case[];
    pagination: {
      page: number;
      page_size: number;
      total: number;
      total_pages: number;
    };
  }> {
    try {
      return await apiClient.getPaginated<Case>("/cases/my", {
        params,
        
      });
    } catch (error: any) {
      console.error("获取我的案件失败:", error);
      throw new AppError(
        error.message || "获取我的案件失败",
        error.code || "GET_MY_CASES_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取客户的案件
  async getClientCases(
    client_id: number,
    params?: Omit<CaseListRequest, "client_id">,
  ): Promise<{
    data: Case[];
    pagination: {
      page: number;
      page_size: number;
      total: number;
      total_pages: number;
    };
  }> {
    try {
      return await apiClient.getPaginated<Case>(`/clients/${client_id}/cases`, {
        params,
        
      });
    } catch (error: any) {
      console.error("获取客户案件失败:", error);
      throw new AppError(
        error.message || "获取客户案件失败",
        error.code || "GET_CLIENT_CASES_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 搜索案件
  async searchCases(
    query: string,
    params?: Omit<CaseListRequest, "search">,
  ): Promise<{
    data: Case[];
    pagination: {
      page: number;
      page_size: number;
      total: number;
      total_pages: number;
    };
  }> {
    try {
      const searchParams = {
        ...params,
        search: query,
      };
      return await apiClient.getPaginated<Case>("/cases/search", {
        params: searchParams,
        
      });
    } catch (error: any) {
      console.error("搜索案件失败:", error);
      throw new AppError(
        error.message || "搜索案件失败",
        error.code || "SEARCH_CASES_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 批量更新案件状态
  async batchUpdateStatus(
    caseIds: number[],
    status: string,
  ): Promise<{
    success: number;
    failed: number;
    errors?: string[];
  }> {
    try {
      return await apiClient.post<{
        success: number;
        failed: number;
        errors?: string[];
      }>(
        "/cases/batch/status",
        { case_ids: caseIds, status },
        {
        },
      );
    } catch (error: any) {
      console.error("批量更新案件状态失败:", error);
      throw new AppError(
        error.message || "批量更新案件状态失败",
        error.code || "BATCH_UPDATE_STATUS_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 批量分配律师
  async batchAssignLawyer(
    caseIds: number[],
    lawyer_id: number,
  ): Promise<{
    success: number;
    failed: number;
    errors?: string[];
  }> {
    try {
      return await apiClient.post<{
        success: number;
        failed: number;
        errors?: string[];
      }>(
        "/cases/batch/assign",
        { case_ids: caseIds, lawyer_id: lawyer_id },
        {
        },
      );
    } catch (error: any) {
      console.error("批量分配律师失败:", error);
      throw new AppError(
        error.message || "批量分配律师失败",
        error.code || "BATCH_ASSIGN_LAWYER_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 批量删除案件
  async batchDeleteCases(caseIds: number[]): Promise<{
    success: number;
    failed: number;
    errors?: string[];
  }> {
    try {
      return await apiClient.post<{
        success: number;
        failed: number;
        errors?: string[];
      }>(
        "/cases/batch/delete",
        { case_ids: caseIds },
        {
        },
      );
    } catch (error: any) {
      console.error("批量删除案件失败:", error);
      throw new AppError(
        error.message || "批量删除案件失败",
        error.code || "BATCH_DELETE_CASES_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 获取案件时间线
  async getCaseTimeline(caseId: number): Promise<
    Array<{
      id: number;
      action: string;
      description: string;
      user: string;
      timestamp: string;
    }>
  > {
    try {
      return await apiClient.get<
        Array<{
          id: number;
          action: string;
          description: string;
          user: string;
          timestamp: string;
        }>
      >(`/cases/${caseId}/timeline`, {
        
      });
    } catch (error: any) {
      console.error("获取案件时间线失败:", error);
      throw new AppError(
        error.message || "获取案件时间线失败",
        error.code || "GET_CASE_TIMELINE_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 添加案件备注
  async addCaseNote(
    caseId: number,
    note: string,
  ): Promise<{
    id: number;
    note: string;
    user: string;
    timestamp: string;
  }> {
    try {
      return await apiClient.post<{
        id: number;
        note: string;
        user: string;
        timestamp: string;
      }>(
        `/cases/${caseId}/notes`,
        { note },
        {
        },
      );
    } catch (error: any) {
      console.error("添加案件备注失败:", error);
      throw new AppError(
        error.message || "添加案件备注失败",
        error.code || "ADD_CASE_NOTE_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 获取案件备注列表
  async getCaseNotes(caseId: number): Promise<
    Array<{
      id: number;
      note: string;
      user: string;
      timestamp: string;
    }>
  > {
    try {
      return await apiClient.get<
        Array<{
          id: number;
          note: string;
          user: string;
          timestamp: string;
        }>
      >(`/cases/${caseId}/notes`, {
        
      });
    } catch (error: any) {
      console.error("获取案件备注列表失败:", error);
      throw new AppError(
        error.message || "获取案件备注列表失败",
        error.code || "GET_CASE_NOTES_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 导出案件数据
  async exportCases(params?: CaseListRequest): Promise<Blob> {
    try {
      const response = await apiClient.getClient().get("/cases/export", {
        params,
        responseType: "blob",
      });
      return response.data;
    } catch (error: any) {
      console.error("导出案件数据失败:", error);
      throw new AppError(
        error.message || "导出案件数据失败",
        error.code || "EXPORT_CASES_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取案件类型列表
  async getCaseTypes(): Promise<
    Array<{
      id: string;
      name: string;
      description: string;
    }>
  > {
    try {
      return await apiClient.get<
        Array<{
          id: string;
          name: string;
          description: string;
        }>
      >("/cases/types", {
        
      });
    } catch (error: any) {
      console.error("获取案件类型列表失败:", error);
      throw new AppError(
        error.message || "获取案件类型列表失败",
        error.code || "GET_CASE_TYPES_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取案件优先级列表
  async getCasePriorities(): Promise<
    Array<{
      id: string;
      name: string;
      description: string;
      color: string;
    }>
  > {
    try {
      return await apiClient.get<
        Array<{
          id: string;
          name: string;
          description: string;
          color: string;
        }>
      >("/cases/priorities", {
        
      });
    } catch (error: any) {
      console.error("获取案件优先级列表失败:", error);
      throw new AppError(
        error.message || "获取案件优先级列表失败",
        error.code || "GET_CASE_PRIORITIES_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取案件状态列表
  async getCaseStatuses(): Promise<
    Array<{
      id: string;
      name: string;
      description: string;
      color: string;
    }>
  > {
    try {
      return await apiClient.get<
        Array<{
          id: string;
          name: string;
          description: string;
          color: string;
        }>
      >("/cases/statuses", {
        
      });
    } catch (error: any) {
      console.error("获取案件状态列表失败:", error);
      throw new AppError(
        error.message || "获取案件状态列表失败",
        error.code || "GET_CASE_STATUSES_ERROR",
        error.statusCode || 500,
      );
    }
  }
}

// 导出单例实例
export const caseService = new CaseService();

// 为了向后兼容，也导出独立的函数
export const getCases = (params?: CaseListRequest) =>
  caseService.getCases(params);
export const getCase = (id: number) => caseService.getCase(id);
export const getCaseById = (id: number) => caseService.getCase(id);
export const createCase = (data: CreateCaseRequest) =>
  caseService.createCase(data);
export const updateCase = (id: number, data: UpdateCaseRequest) =>
  caseService.updateCase(id, data);
export const deleteCase = (id: number) => caseService.deleteCase(id);
export const getCaseStats = () => caseService.getCaseStats();
export const assignLawyer = (caseId: number, lawyer_id: number) =>
  caseService.assignLawyer(caseId, lawyer_id);
export const updateCaseStatus = (caseId: number, status: string) =>
  caseService.updateCaseStatus(caseId, status);
export const batchDeleteCases = (caseIds: number[]) =>
  caseService.batchDeleteCases(caseIds);
export const batchUpdateStatus = (caseIds: number[], status: string) =>
  caseService.batchUpdateStatus(caseIds, status);
export const batchAssignLawyer = (caseIds: number[], lawyer_id: number) =>
  caseService.batchAssignLawyer(caseIds, lawyer_id);
export const getMyCases = (params?: Omit<CaseListRequest, "lawyer_id">) =>
  caseService.getMyCases(params);
export const getClientCases = (
  client_id: number,
  params?: Omit<CaseListRequest, "client_id">,
) => caseService.getClientCases(client_id, params);
export const searchCases = (
  query: string,
  params?: Omit<CaseListRequest, "search">,
) => caseService.searchCases(query, params);
export const exportCases = (params?: CaseListRequest) =>
  caseService.exportCases(params);
export const getCaseTypes = () => caseService.getCaseTypes();
export const getCasePriorities = () => caseService.getCasePriorities();
export const getCaseStatuses = () => caseService.getCaseStatuses();
export const getCaseTimeline = (caseId: number) =>
  caseService.getCaseTimeline(caseId);
export const addCaseNote = (caseId: number, note: string) =>
  caseService.addCaseNote(caseId, note);
export const getCaseNotes = (caseId: number) =>
  caseService.getCaseNotes(caseId);

export default caseService;
