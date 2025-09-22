import apiClient from "./api";
import {
  FileInfo,
  FileListRequest,
  FileListResponse,
  FileStats,
  UploadFileRequest,
  FileUploadResponse,
} from "../types";
import { AppError } from "../types/errors";

class DocumentService {
  // 获取文件列表
  async getFiles(params?: FileListRequest): Promise<FileListResponse> {
    try {
      return await apiClient.getPaginated<FileInfo>("/files", {
        params,
        
      });
    } catch (error: any) {
      console.error("获取文件列表失败:", error);
      throw new AppError(
        error.message || "获取文件列表失败",
        error.code || "GET_FILES_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取文件详情
  async getFile(id: string): Promise<FileInfo> {
    try {
      return await apiClient.get<FileInfo>(`/files/${id}`, {
        
      });
    } catch (error: any) {
      console.error("获取文件详情失败:", error);
      throw new AppError(
        error.message || "获取文件详情失败",
        error.code || "GET_FILE_ERROR",
        error.statusCode || 404,
      );
    }
  }

  // 上传文件
  async uploadFile(data: UploadFileRequest): Promise<FileUploadResponse> {
    try {
      const formData = new FormData();
      formData.append("file", data.file);
      formData.append("category", data.category);
      if (data.description) {
        formData.append("description", data.description);
      }

      const response = await apiClient
        .getClient()
        .post("/files/upload", formData, {
          headers: {
            "Content-Type": "multipart/form-data",
          },
          timeout: 30000, // 30秒超时
        });

      return response.data;
    } catch (error: any) {
      console.error("上传文件失败:", error);
      throw new AppError(
        error.message || "上传文件失败",
        error.code || "UPLOAD_FILE_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 删除文件
  async deleteFile(id: string): Promise<{ message: string }> {
    try {
      return await apiClient.delete<{ message: string }>(`/files/${id}`, {
      });
    } catch (error: any) {
      console.error("删除文件失败:", error);
      throw new AppError(
        error.message || "删除文件失败",
        error.code || "DELETE_FILE_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 获取文件统计信息
  async getFileStats(): Promise<FileStats> {
    try {
      return await apiClient.get<FileStats>("/files/stats", {
        
      });
    } catch (error: any) {
      console.error("获取文件统计信息失败:", error);
      throw new AppError(
        error.message || "获取文件统计信息失败",
        error.code || "GET_FILE_STATS_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 下载文件
  async downloadFile(id: string): Promise<Blob> {
    try {
      const response = await apiClient
        .getClient()
        .get(`/files/${id}/download`, {
          responseType: "blob",
          timeout: 60000, // 60秒超时
        });
      return response.data;
    } catch (error: any) {
      console.error("下载文件失败:", error);
      throw new AppError(
        error.message || "下载文件失败",
        error.code || "DOWNLOAD_FILE_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 预览文件
  async previewFile(id: string): Promise<string> {
    try {
      const response = await apiClient.getClient().get(`/files/${id}/preview`, {
        responseType: "blob",
        timeout: 30000, // 30秒超时
      });
      return URL.createObjectURL(response.data);
    } catch (error: any) {
      console.error("预览文件失败:", error);
      throw new AppError(
        error.message || "预览文件失败",
        error.code || "PREVIEW_FILE_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 更新文件信息
  async updateFile(
    id: string,
    data: { category?: string; description?: string },
  ): Promise<FileInfo> {
    try {
      return await apiClient.put<FileInfo>(`/files/${id}`, data, {
      });
    } catch (error: any) {
      console.error("更新文件信息失败:", error);
      throw new AppError(
        error.message || "更新文件信息失败",
        error.code || "UPDATE_FILE_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 批量删除文件
  async batchDeleteFiles(fileIds: string[]): Promise<{
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
        "/files/batch/delete",
        { file_ids: fileIds },
        {
        },
      );
    } catch (error: any) {
      console.error("批量删除文件失败:", error);
      throw new AppError(
        error.message || "批量删除文件失败",
        error.code || "BATCH_DELETE_FILES_ERROR",
        error.statusCode || 400,
      );
    }
  }

  // 获取文件分类统计
  async getFileCategories(): Promise<
    Array<{
      id: string;
      name: string;
      count: number;
      total_size: number;
    }>
  > {
    try {
      return await apiClient.get<
        Array<{
          id: string;
          name: string;
          count: number;
          total_size: number;
        }>
      >("/files/categories", {
        
      });
    } catch (error: any) {
      console.error("获取文件分类统计失败:", error);
      throw new AppError(
        error.message || "获取文件分类统计失败",
        error.code || "GET_FILE_CATEGORIES_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 获取文件类型统计
  async getFileTypes(): Promise<
    Array<{
      type: string;
      count: number;
      total_size: number;
    }>
  > {
    try {
      return await apiClient.get<
        Array<{
          type: string;
          count: number;
          total_size: number;
        }>
      >("/files/types", {
        
      });
    } catch (error: any) {
      console.error("获取文件类型统计失败:", error);
      throw new AppError(
        error.message || "获取文件类型统计失败",
        error.code || "GET_FILE_TYPES_ERROR",
        error.statusCode || 500,
      );
    }
  }

  // 搜索文件
  async searchFiles(
    query: string,
    params?: Omit<FileListRequest, "search">,
  ): Promise<FileListResponse> {
    try {
      const searchParams = {
        ...params,
        search: query,
      };
      return await apiClient.getPaginated<FileInfo>("/files/search", {
        params: searchParams,
        
      });
    } catch (error: any) {
      console.error("搜索文件失败:", error);
      throw new AppError(
        error.message || "搜索文件失败",
        error.code || "SEARCH_FILES_ERROR",
        error.statusCode || 500,
      );
    }
  }
}

// 导出单例实例
export const documentService = new DocumentService();

// 为了向后兼容，也导出独立的函数
export const getFiles = (params?: FileListRequest) =>
  documentService.getFiles(params);
export const getDocuments = (params?: FileListRequest) =>
  documentService.getFiles(params);
export const getFile = (id: string) => documentService.getFile(id);
export const uploadFile = (data: UploadFileRequest) =>
  documentService.uploadFile(data);
export const uploadDocument = (data: UploadFileRequest) =>
  documentService.uploadFile(data);
export const deleteFile = (id: string) => documentService.deleteFile(id);
export const deleteDocument = (id: string) => documentService.deleteFile(id);
export const getFileStats = () => documentService.getFileStats();
export const downloadFile = (id: string) => documentService.downloadFile(id);
export const previewFile = (id: string) => documentService.previewFile(id);
export const updateFile = (
  id: string,
  data: { category?: string; description?: string },
) => documentService.updateFile(id, data);
export const batchDeleteFiles = (fileIds: string[]) =>
  documentService.batchDeleteFiles(fileIds);
export const getFileCategories = () => documentService.getFileCategories();
export const getFileTypes = () => documentService.getFileTypes();
export const searchFiles = (
  query: string,
  params?: Omit<FileListRequest, "search">,
) => documentService.searchFiles(query, params);

export default documentService;
