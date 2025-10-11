import apiClient from "./api";
import { AppError, Report, ReportListRequest, CreateReportRequest, UpdateReportRequest } from "../types";

// 确保Report接口包含服务需要的字段
interface ServiceReport extends Report {
  created_by: number;
  created_by_name: string;
  error_message?: string;
}

interface ReportStats {
  total: number;
  pending: number;
  generating: number;
  completed: number;
  failed: number;
  by_type: Record<string, number>;
  by_format: Record<string, number>;
  by_status: Record<string, number>;
}

interface ReportTemplate {
  id: string;
  name: string;
  description: string;
  type: string;
  supported_formats: ("pdf" | "excel" | "csv")[];
  parameters: Array<{
    name: string;
    type: string;
    required: boolean;
    description: string;
    default_value?: any;
  }>;
}

class ReportService {
  // 获取报告列表
  async getReports(params?: ReportListRequest): Promise<{
    data: Report[];
    pagination: {
      page: number;
      page_size: number;
      total: number;
      total_pages: number;
    };
  }> {
    try {
      return await apiClient.getPaginated<Report>("/reports", {
        params,
        
      });
    } catch (error: any) {
      console.error("获取报告列表失败:", error);
      throw new AppError(
        error.message || "获取报告列表失败",
        error.code || "GET_REPORTS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取报告详情
  async getReport(id: number): Promise<Report> {
    try {
      return await apiClient.get<Report>(`/reports/${id}`, {
        
      });
    } catch (error: any) {
      console.error("获取报告详情失败:", error);
      throw new AppError(
        error.message || "获取报告详情失败",
        error.code || "GET_REPORT_ERROR",
        error.statusCode || 404,
      );
    }
  }

  // 创建报告
  async createReport(data: CreateReportRequest): Promise<Report> {
    try {
      return await apiClient.post<Report>("/reports", data, {
      });
    } catch (error: any) {
      console.error("创建报告失败:", error);
      throw new AppError(
        error.message || "创建报告失败",
        error.code || "CREATE_REPORT_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 更新报告
  async updateReport(id: number, data: UpdateReportRequest): Promise<Report> {
    try {
      return await apiClient.put<Report>(`/reports/${id}`, data, {
      });
    } catch (error: any) {
      console.error("更新报告失败:", error);
      throw new AppError(
        error.message || "更新报告失败",
        error.code || "UPDATE_REPORT_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 删除报告
  async deleteReport(id: number): Promise<{ message: string }> {
    try {
      return await apiClient.delete<{ message: string }>(`/reports/${id}`, {
      });
    } catch (error: any) {
      console.error("删除报告失败:", error);
      throw new AppError(
        error.message || "删除报告失败",
        error.code || "DELETE_REPORT_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 下载报告
  async downloadReport(id: number): Promise<Blob> {
    try {
      const response = await apiClient
        .getClient()
        .get(`/reports/${id}/download`, {
          responseType: "blob",
          timeout: 60000, // 60秒超时
        });
      return response.data;
    } catch (error: any) {
      console.error("下载报告失败:", error);
      throw new AppError(
        error.message || "下载报告失败",
        error.code || "DOWNLOAD_REPORT_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 重新生成报告
  async regenerateReport(id: number): Promise<Report> {
    try {
      return await apiClient.post<Report>(
        `/reports/${id}/regenerate`,
        {},
      );
    } catch (error: any) {
      console.error("重新生成报告失败:", error);
      throw new AppError(
        error.message || "重新生成报告失败",
        error.code || "REGENERATE_REPORT_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 获取报告统计信息
  async getReportStats(): Promise<ReportStats> {
    try {
      return await apiClient.get<ReportStats>("/reports/stats", {
        
      });
    } catch (error: any) {
      console.error("获取报告统计信息失败:", error);
      throw new AppError(
        error.message || "获取报告统计信息失败",
        error.code || "GET_REPORT_STATS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取报告模板列表
  async getReportTemplates(): Promise<ReportTemplate[]> {
    try {
      return await apiClient.get<ReportTemplate[]>("/reports/templates", {
        
      });
    } catch (error: any) {
      console.error("获取报告模板列表失败:", error);
      throw new AppError(
        error.message || "获取报告模板列表失败",
        error.code || "GET_REPORT_TEMPLATES_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取报告模板详情
  async getReportTemplate(id: string): Promise<ReportTemplate> {
    try {
      return await apiClient.get<ReportTemplate>(`/reports/templates/${id}`, {
        
      });
    } catch (error: any) {
      console.error("获取报告模板详情失败:", error);
      throw new AppError(
        error.message || "获取报告模板详情失败",
        error.code || "GET_REPORT_TEMPLATE_ERROR",
        error.statusCode || 404,
      );
    }
  }

  // 预览报告
  async previewReport(id: number): Promise<string> {
    try {
      const response = await apiClient
        .getClient()
        .get(`/reports/${id}/preview`, {
          responseType: "blob",
          timeout: 30000, // 30秒超时
        });
      return URL.createObjectURL(response.data);
    } catch (error: any) {
      console.error("预览报告失败:", error);
      throw new AppError(
        error.message || "预览报告失败",
        error.code || "PREVIEW_REPORT_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 获取我的报告
  async getMyReports(params?: Omit<ReportListRequest, "created_by">): Promise<{
    data: Report[];
    pagination: {
      page: number;
      page_size: number;
      total: number;
      total_pages: number;
    };
  }> {
    try {
      return await apiClient.getPaginated<Report>("/reports/my", {
        params,
        
      });
    } catch (error: any) {
      console.error("获取我的报告失败:", error);
      throw new AppError(
        error.message || "获取我的报告失败",
        error.code || "GET_MY_REPORTS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 搜索报告
  async searchReports(
    query: string,
    params?: Omit<ReportListRequest, "search">,
  ): Promise<{
    data: Report[];
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
      return await apiClient.getPaginated<Report>("/reports/search", {
        params: searchParams,
        
      });
    } catch (error: any) {
      console.error("搜索报告失败:", error);
      throw new AppError(
        error.message || "搜索报告失败",
        error.code || "SEARCH_REPORTS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 批量删除报告
  async batchDeleteReports(reportIds: number[]): Promise<{
    success: number;
    failed: number;
    errors?: string[];
  }> {
    try {
      return await apiClient.post<{
        success: number;
        failed: number;
        errors?: string[];
      }>("/reports/batch/delete", { report_ids: reportIds }, {
      });
    } catch (error: any) {
      console.error("批量删除报告失败:", error);
      throw new AppError(
        error.message || "批量删除报告失败",
        error.code || "BATCH_DELETE_REPORTS_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 获取报告类型列表
  async getReportTypes(): Promise<
    Array<{
      id: string;
      name: string;
      description: string;
      supported_formats: ("pdf" | "excel" | "csv")[];
    }>
  > {
    try {
      return await apiClient.get<
        Array<{
          id: string;
          name: string;
          description: string;
          supported_formats: ("pdf" | "excel" | "csv")[];
        }>
      >("/reports/types", {
        
      });
    } catch (error: any) {
      console.error("获取报告类型列表失败:", error);
      throw new AppError(
        error.message || "获取报告类型列表失败",
        error.code || "GET_REPORT_TYPES_ERROR",
        error.statusCode || 500,
      );
    }
  }
}

// 导出单例实例
export const reportService = new ReportService();

// 为了向后兼容，也导出独立的函数
export const getReports = (params?: ReportListRequest) =>
  reportService.getReports(params);
export const getReport = (id: number) => reportService.getReport(id);
export const createReport = (data: CreateReportRequest) =>
  reportService.createReport(data);
export const updateReport = (id: number, data: UpdateReportRequest) =>
  reportService.updateReport(id, data);
export const deleteReport = (id: number) => reportService.deleteReport(id);
export const downloadReport = (id: number) => reportService.downloadReport(id);
export const regenerateReport = (id: number) =>
  reportService.regenerateReport(id);
export const getReportStats = () => reportService.getReportStats();
export const getReportTemplates = () => reportService.getReportTemplates();
export const getReportTemplate = (id: string) =>
  reportService.getReportTemplate(id);
export const previewReport = (id: number) => reportService.previewReport(id);
export const getMyReports = (params?: Omit<ReportListRequest, "created_by">) =>
  reportService.getMyReports(params);
export const searchReports = (
  query: string,
  params?: Omit<ReportListRequest, "search">,
) => reportService.searchReports(query, params);
export const batchDeleteReports = (reportIds: number[]) =>
  reportService.batchDeleteReports(reportIds);
export const getReportTypes = () => reportService.getReportTypes();
export const generateReport = (data: CreateReportRequest) =>
  reportService.createReport(data);

export default reportService;
