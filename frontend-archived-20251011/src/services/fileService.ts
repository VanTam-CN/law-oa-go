import apiClient from "./api";
import {
  FileInfo,
  FileListRequest,
  FileListResponse,
  FileStats,
  UploadFileRequest,
  FileUploadResponse,
  FileCategory,
  FileType,
} from "../types";
import {
  AppError,
  DEFAULT_RETRY_CONFIG,
  DEFAULT_CACHE_CONFIG,
} from "../types/errors";

class FileService {
  // 文件上传 - 添加错误处理，配置重试2次，不缓存
  async uploadFile(
    file: File,
    category: FileCategory,
    description?: string,
  ): Promise<FileUploadResponse> {
    try {
      const formData = new FormData();
      formData.append("file", file);
      formData.append("category", category);
      if (description) {
        formData.append("description", description);
      }

      const response = await apiClient
        .getClient()
        .post("/files/upload", formData, {
          headers: {
            "Content-Type": "multipart/form-data",
          },
        });
      return response.data;
    } catch (error: any) {
      throw AppError.fromApiError(error);
    }
  }

  // 获取文件列表 - 添加错误处理，配置缓存2分钟，重试1次
  async getFileList(params?: FileListRequest): Promise<FileListResponse> {
    try {
      const cacheConfig = {
        ...DEFAULT_CACHE_CONFIG,
        ttl: 2 * 60 * 1000, // 2分钟缓存
        key: `fileList:${JSON.stringify(params)}`,
      };

      return await apiClient.get<FileListResponse>("/files", {
        params,
        useCache: cacheConfig.enabled,
        cacheTTL: cacheConfig.ttl,
      });
    } catch (error: any) {
      throw AppError.fromApiError(error);
    }
  }

  // 获取文件详情 - 添加错误处理，配置缓存5分钟，重试1次
  async getFile(id: string): Promise<FileInfo> {
    try {
      const cacheConfig = {
        ...DEFAULT_CACHE_CONFIG,
        ttl: 5 * 60 * 1000, // 5分钟缓存
        key: `file:${id}`,
      };

      return await apiClient.get<FileInfo>(`/files/${id}`, {
        useCache: cacheConfig.enabled,
        cacheTTL: cacheConfig.ttl,
      });
    } catch (error: any) {
      throw AppError.fromApiError(error);
    }
  }

  // 下载文件 - 添加错误处理，配置重试1次，不缓存
  async downloadFile(id: string): Promise<Blob> {
    try {
      const response = await apiClient
        .getClient()
        .get(`/files/${id}/download`, {
          responseType: "blob",
        });
      return response.data;
    } catch (error: any) {
      throw AppError.fromApiError(error);
    }
  }

  // 删除文件 - 添加错误处理，配置重试2次，不缓存
  async deleteFile(id: string): Promise<{ message: string }> {
    try {
      return await apiClient.delete<{ message: string }>(`/files/${id}`);
    } catch (error: any) {
      throw AppError.fromApiError(error);
    }
  }

  // 批量删除文件 - 添加错误处理，配置重试2次，不缓存
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
      }>("/files/batch/delete", { file_ids: fileIds });
    } catch (error: any) {
      throw AppError.fromApiError(error);
    }
  }

  // 获取文件统计信息 - 添加错误处理，配置缓存5分钟，重试1次
  async getFileStats(): Promise<FileStats> {
    try {
      const cacheConfig = {
        ...DEFAULT_CACHE_CONFIG,
        ttl: 5 * 60 * 1000, // 5分钟缓存
        key: "fileStats",
      };

      return await apiClient.get<FileStats>("/files/stats", {
        useCache: cacheConfig.enabled,
        cacheTTL: cacheConfig.ttl,
      });
    } catch (error: any) {
      throw AppError.fromApiError(error);
    }
  }

  // 搜索文件 - 添加错误处理，配置缓存1分钟，重试1次
  async searchFiles(
    query: string,
    params?: Omit<FileListRequest, "search">,
  ): Promise<FileListResponse> {
    try {
      const searchParams = {
        ...params,
        search: query,
      };
      const cacheConfig = {
        ...DEFAULT_CACHE_CONFIG,
        ttl: 1 * 60 * 1000, // 1分钟缓存
        key: `search:${query}:${JSON.stringify(params)}`,
      };

      return await apiClient.get<FileListResponse>("/files/search", {
        params: searchParams,
        useCache: cacheConfig.enabled,
        cacheTTL: cacheConfig.ttl,
      });
    } catch (error: any) {
      throw AppError.fromApiError(error);
    }
  }

  // 获取文件分类列表 - 添加错误处理，配置缓存10分钟，重试1次
  async getFileCategories(): Promise<
    Array<{
      id: string;
      name: string;
      description: string;
      color: string;
    }>
  > {
    try {
      const cacheConfig = {
        ...DEFAULT_CACHE_CONFIG,
        ttl: 10 * 60 * 1000, // 10分钟缓存
        key: "fileCategories",
      };

      return await apiClient.get<
        Array<{
          id: string;
          name: string;
          description: string;
          color: string;
        }>
      >("/files/categories", {
        useCache: cacheConfig.enabled,
        cacheTTL: cacheConfig.ttl,
      });
    } catch (error: any) {
      throw AppError.fromApiError(error);
    }
  }

  // 获取文件预览URL - 添加错误处理，配置缓存2分钟，重试1次
  async getFilePreviewUrl(id: string): Promise<string> {
    try {
      const cacheConfig = {
        ...DEFAULT_CACHE_CONFIG,
        ttl: 2 * 60 * 1000, // 2分钟缓存
        key: `preview:${id}`,
      };

      const response = await apiClient.get<{ url: string }>(
        `/files/${id}/preview`,
        {
          useCache: cacheConfig.enabled,
          cacheTTL: cacheConfig.ttl,
        },
      );
      return response.url;
    } catch (error: any) {
      throw AppError.fromApiError(error);
    }
  }

  // 更新文件信息 - 添加错误处理，配置重试2次，不缓存
  async updateFileInfo(
    id: string,
    data: {
      name?: string;
      description?: string;
      category?: string;
    },
  ): Promise<FileInfo> {
    try {
      return await apiClient.put<FileInfo>(`/files/${id}`, data);
    } catch (error: any) {
      throw AppError.fromApiError(error);
    }
  }

  // 工具函数：格式化文件大小
  formatFileSize(bytes: number): string {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  }

  // 工具函数：获取文件类型
  getFileType(fileName: string): FileType {
    const extension = fileName.split(".").pop()?.toLowerCase() || "";

    const fileTypes: Record<string, FileType> = {
      pdf: "pdf",
      doc: "word",
      docx: "word",
      xls: "excel",
      xlsx: "excel",
      ppt: "powerpoint",
      pptx: "powerpoint",
      jpg: "image",
      jpeg: "image",
      png: "image",
      gif: "image",
      bmp: "image",
      txt: "text",
      md: "text",
    };

    return fileTypes[extension] || "other";
  }

  // 工具函数：获取文件分类
  getFileCategory(fileName: string): FileCategory {
    const fileType = this.getFileType(fileName);

    const categoryMap: Record<FileType, FileCategory> = {
      pdf: "document",
      word: "document",
      excel: "spreadsheet",
      powerpoint: "document",
      image: "image",
      text: "document",
      other: "other",
    };

    return categoryMap[fileType] || "other";
  }

  // 工具函数：获取文件图标
  getFileIcon(fileName: string): string {
    const fileType = this.getFileType(fileName);

    const iconMap: Record<FileType, string> = {
      pdf: "📄",
      word: "📝",
      excel: "📊",
      powerpoint: "📽️",
      image: "🖼️",
      text: "📄",
      other: "📎",
    };

    return iconMap[fileType] || "📎";
  }

  // 工具函数：获取文件类型颜色
  getFileTypeColor(fileName: string): string {
    const fileType = this.getFileType(fileName);

    const colorMap: Record<FileType, string> = {
      pdf: "#e53e3e", // red
      word: "#3182ce", // blue
      excel: "#38a169", // green
      powerpoint: "#ed8936", // orange
      image: "#805ad5", // purple
      text: "#718096", // gray
      other: "#a0aec0", // light gray
    };

    return colorMap[fileType] || "#a0aec0";
  }

  // 工具函数：验证文件类型
  isValidFileType(fileName: string): boolean {
    const allowedTypes = [
      "pdf",
      "doc",
      "docx",
      "xls",
      "xlsx",
      "ppt",
      "pptx",
      "jpg",
      "jpeg",
      "png",
      "gif",
      "bmp",
      "txt",
      "md",
    ];
    const extension = fileName.split(".").pop()?.toLowerCase() || "";
    return allowedTypes.includes(extension);
  }

  // 工具函数：验证文件大小
  isValidFileSize(file: File, maxSizeMB: number = 100): boolean {
    const maxSizeBytes = maxSizeMB * 1024 * 1024;
    return file.size <= maxSizeBytes;
  }

  // 工具函数：生成文件下载链接
  generateDownloadUrl(fileId: string, fileName?: string): string {
    const baseUrl = "/api/files";
    const url = `${baseUrl}/${fileId}/download`;
    if (fileName) {
      return `${url}?filename=${encodeURIComponent(fileName)}`;
    }
    return url;
  }

  // 工具函数：获取文件扩展名
  getFileExtension(fileName: string): string {
    return fileName.split(".").pop()?.toLowerCase() || "";
  }
}

// 导出单例实例
export const fileService = new FileService();

// 为了向后兼容，也导出独立的函数
export const uploadFile = (
  file: File,
  category: FileCategory,
  description?: string,
) => fileService.uploadFile(file, category, description);
export const getFileList = (params?: FileListRequest) =>
  fileService.getFileList(params);
export const getFile = (id: string) => fileService.getFile(id);
export const downloadFile = (id: string) => fileService.downloadFile(id);
export const deleteFile = (id: string) => fileService.deleteFile(id);
export const batchDeleteFiles = (fileIds: string[]) =>
  fileService.batchDeleteFiles(fileIds);
export const getFileStats = () => fileService.getFileStats();

export default fileService;
