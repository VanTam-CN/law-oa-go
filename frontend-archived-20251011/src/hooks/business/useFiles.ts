import { useState, useEffect, useCallback } from "react";
import { useSelector, useDispatch } from "react-redux";
import { RootState } from "../../store";
import { fileService } from "../../services/fileService";
import {
  FileInfo,
  FileListRequest,
  FileStats,
  FileCategory,
  FileType,
} from "../../types";
import { AppError } from "../../types/errors";

// 文件上传请求类型
interface FileUploadRequest {
  file: File;
  category: FileCategory;
  description?: string;
}

// 文件列表返回类型
interface FileListResponse {
  data: FileInfo[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
}

// Hook 返回类型
interface UseFilesReturn {
  // 数据状态
  files: FileInfo[];
  loading: boolean;
  error: string | null;
  uploading: boolean;

  // 分页状态
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };

  // 统计信息
  stats: FileStats | null;

  // 操作方法
  fetchFiles: (params?: FileListRequest) => Promise<void>;
  uploadFile: (data: FileUploadRequest) => Promise<FileInfo>;
  updateFile: (id: string, data: Partial<FileInfo>) => Promise<FileInfo>;
  deleteFile: (id: string) => Promise<void>;
  getFile: (id: string) => Promise<FileInfo>;
  downloadFile: (id: string) => Promise<Blob>;

  // 批量操作
  batchDeleteFiles: (
    fileIds: string[],
  ) => Promise<{ success: number; failed: number; errors?: string[] }>;

  // 搜索和过滤
  searchFiles: (
    query: string,
    params?: FileListRequest,
  ) => Promise<FileListResponse>;
  getFilesByCategory: (category: FileCategory) => Promise<FileInfo[]>;
  getFilesByType: (type: FileType) => Promise<FileInfo[]>;

  // 预览和元数据
  getFilePreviewUrl: (id: string) => Promise<string>;
  getFileStats: () => Promise<FileStats>;

  // 工具方法
  refresh: () => Promise<void>;
  clearError: () => void;
  formatFileSize: (bytes: number) => string;
  getFileType: (fileName: string) => FileType;
}

/**
 * 文件管理 Hook
 * 提供文件相关的数据获取、上传、下载、CRUD 操作和状态管理
 */
export const useFiles = (defaultParams?: FileListRequest): UseFilesReturn => {
  const dispatch = useDispatch();

  // Redux 状态
  const { user } = useSelector((state: RootState) => state.auth);

  // 本地状态
  const [files, setFiles] = useState<FileInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [pagination, setPagination] = useState({
    page: 1,
    page_size: 10,
    total: 0,
    total_pages: 0,
  });
  const [stats, setStats] = useState<FileStats | null>(null);

  // 获取文件列表
  const fetchFiles = useCallback(
    async (params?: FileListRequest) => {
      try {
        setLoading(true);
        setError(null);

        const mergedParams = { ...defaultParams, ...params };
        const response = await fileService.getFileList(mergedParams);

        setFiles(response.data);
        setPagination(response.pagination);
      } catch (err: any) {
        const appError =
          err instanceof AppError ? err : AppError.fromApiError(err);
        setError(appError.message);
        console.error("获取文件列表失败:", err);
      } finally {
        setLoading(false);
      }
    },
    [defaultParams],
  );

  // 上传文件
  const uploadFile = useCallback(
    async (data: FileUploadRequest): Promise<FileInfo> => {
      try {
        setUploading(true);
        setError(null);

        const newFile = await fileService.uploadFile(
          data.file,
          data.category,
          data.description,
        );

        // 更新本地列表
        setFiles((prev) => [newFile, ...prev]);

        return newFile;
      } catch (err: any) {
        const appError =
          err instanceof AppError ? err : AppError.fromApiError(err);
        setError(appError.message);
        console.error("上传文件失败:", err);
        throw err;
      } finally {
        setUploading(false);
      }
    },
    [],
  );

  // 更新文件
  const updateFile = useCallback(
    async (id: string, data: Partial<FileInfo>): Promise<FileInfo> => {
      try {
        setLoading(true);
        setError(null);

        const updatedFile = await fileService.updateFileInfo(id, data);

        // 更新本地列表
        setFiles((prev) =>
          prev.map((file) => (file.id === id ? updatedFile : file)),
        );

        return updatedFile;
      } catch (err: any) {
        const appError =
          err instanceof AppError ? err : AppError.fromApiError(err);
        setError(appError.message);
        console.error("更新文件失败:", err);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  // 删除文件
  const deleteFile = useCallback(async (id: string): Promise<void> => {
    try {
      setLoading(true);
      setError(null);

      await fileService.deleteFile(id);

      // 更新本地列表
      setFiles((prev) => prev.filter((file) => file.id !== id));
    } catch (err: any) {
      const appError =
        err instanceof AppError ? err : AppError.fromApiError(err);
      setError(appError.message);
      console.error("删除文件失败:", err);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // 获取单个文件
  const getFile = useCallback(async (id: string): Promise<FileInfo> => {
    try {
      setLoading(true);
      setError(null);

      return await fileService.getFile(id);
    } catch (err: any) {
      const appError =
        err instanceof AppError ? err : AppError.fromApiError(err);
      setError(appError.message);
      console.error("获取文件详情失败:", err);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // 下载文件
  const downloadFile = useCallback(async (id: string): Promise<Blob> => {
    try {
      setLoading(true);
      setError(null);

      return await fileService.downloadFile(id);
    } catch (err: any) {
      const appError =
        err instanceof AppError ? err : AppError.fromApiError(err);
      setError(appError.message);
      console.error("下载文件失败:", err);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // 批量删除文件
  const batchDeleteFiles = useCallback(
    async (
      fileIds: string[],
    ): Promise<{ success: number; failed: number; errors?: string[] }> => {
      try {
        setLoading(true);
        setError(null);

        const result = await fileService.batchDeleteFiles(fileIds);

        // 更新本地列表
        if (result.success > 0) {
          setFiles((prev) => prev.filter((file) => !fileIds.includes(file.id)));
        }

        return result;
      } catch (err: any) {
        const appError =
          err instanceof AppError ? err : AppError.fromApiError(err);
        setError(appError.message);
        console.error("批量删除文件失败:", err);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  // 搜索文件
  const searchFiles = useCallback(
    async (
      query: string,
      params?: FileListRequest,
    ): Promise<FileListResponse> => {
      try {
        setLoading(true);
        setError(null);

        const response = await fileService.searchFiles(query, params);

        // 更新本地列表
        setFiles(response.data);
        setPagination(response.pagination);

        return response;
      } catch (err: any) {
        const appError =
          err instanceof AppError ? err : AppError.fromApiError(err);
        setError(appError.message);
        console.error("搜索文件失败:", err);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  // 按分类获取文件
  const getFilesByCategory = useCallback(
    async (category: FileCategory): Promise<FileInfo[]> => {
      try {
        setLoading(true);
        setError(null);

        const response = await fileService.getFileList({
          category,
          page: 1,
          page_size: 1000,
        });

        return response.data;
      } catch (err: any) {
        const appError =
          err instanceof AppError ? err : AppError.fromApiError(err);
        setError(appError.message);
        console.error("获取分类文件失败:", err);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  // 按类型获取文件
  const getFilesByType = useCallback(
    async (type: FileType): Promise<FileInfo[]> => {
      try {
        setLoading(true);
        setError(null);

        const response = await fileService.getFileList({
          file_type: type,
          page: 1,
          page_size: 1000,
        });

        return response.data;
      } catch (err: any) {
        const appError =
          err instanceof AppError ? err : AppError.fromApiError(err);
        setError(appError.message);
        console.error("获取类型文件失败:", err);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  // 获取文件预览URL
  const getFilePreviewUrl = useCallback(async (id: string): Promise<string> => {
    try {
      setLoading(true);
      setError(null);

      return await fileService.getFilePreviewUrl(id);
    } catch (err: any) {
      const appError =
        err instanceof AppError ? err : AppError.fromApiError(err);
      setError(appError.message);
      console.error("获取文件预览URL失败:", err);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // 获取文件统计
  const getFileStats = useCallback(async (): Promise<FileStats> => {
    try {
      setLoading(true);
      setError(null);

      const statsData = await fileService.getFileStats();
      setStats(statsData);

      return statsData;
    } catch (err: any) {
      const appError =
        err instanceof AppError ? err : AppError.fromApiError(err);
      setError(appError.message);
      console.error("获取文件统计失败:", err);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // 格式化文件大小
  const formatFileSize = useCallback((bytes: number): string => {
    if (bytes === 0) return "0 Bytes";

    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));

    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  }, []);

  // 获取文件类型
  const getFileType = useCallback((fileName: string): FileType => {
    const extension = fileName.split(".").pop()?.toLowerCase() || "";

    const imageTypes = ["jpg", "jpeg", "png", "gif", "bmp", "svg", "webp"];
    const videoTypes = ["mp4", "avi", "mov", "wmv", "flv", "webm"];
    const audioTypes = ["mp3", "wav", "aac", "flac", "ogg"];

    if (imageTypes.includes(extension)) return "image";
    if (extension === "pdf") return "pdf";
    if (extension === "doc" || extension === "docx") return "word";
    if (extension === "xls" || extension === "xlsx") return "excel";
    if (extension === "ppt" || extension === "pptx") return "powerpoint";
    if (extension === "txt") return "text";
    if (videoTypes.includes(extension)) return "other";
    if (audioTypes.includes(extension)) return "other";

    return "other";
  }, []);

  // 刷新数据
  const refresh = useCallback(async (): Promise<void> => {
    await fetchFiles();
    if (stats) {
      await getFileStats();
    }
  }, [fetchFiles, stats, getFileStats]);

  // 清除错误
  const clearError = useCallback((): void => {
    setError(null);
  }, []);

  // 初始化加载
  useEffect(() => {
    if (user) {
      fetchFiles();
    }
  }, [user, fetchFiles]);

  return {
    // 数据状态
    files,
    loading,
    error,
    uploading,

    // 分页状态
    pagination,

    // 统计信息
    stats,

    // 操作方法
    fetchFiles,
    uploadFile,
    updateFile,
    deleteFile,
    getFile,
    downloadFile,

    // 批量操作
    batchDeleteFiles,

    // 搜索和过滤
    searchFiles,
    getFilesByCategory,
    getFilesByType,

    // 预览和元数据
    getFilePreviewUrl,
    getFileStats,

    // 工具方法
    refresh,
    clearError,
    formatFileSize,
    getFileType,
  };
};

export default useFiles;
export type { UseFilesReturn };
