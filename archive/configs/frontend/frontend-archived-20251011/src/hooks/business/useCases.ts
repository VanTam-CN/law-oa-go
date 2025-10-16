import { useState, useEffect, useCallback } from "react";
import { useSelector, useDispatch } from "react-redux";
import { RootState } from "../../store";
import { caseService } from "../../services/caseService";
import {
  Case,
  CaseListRequest,
  CreateCaseRequest,
  UpdateCaseRequest,
  AssignLawyerRequest,
  UpdateCaseStatusRequest,
  CaseStats,
} from "../../types";
import { AppError } from "../../types/errors";

// 案件列表返回类型
interface CaseListResponse {
  data: Case[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
}

// Hook 返回类型
interface UseCasesReturn {
  // 数据状态
  cases: Case[];
  loading: boolean;
  error: string | null;

  // 分页状态
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };

  // 统计信息
  stats: CaseStats | null;

  // 操作方法
  fetchCases: (params?: CaseListRequest) => Promise<void>;
  createCase: (data: CreateCaseRequest) => Promise<Case>;
  updateCase: (id: number, data: UpdateCaseRequest) => Promise<Case>;
  deleteCase: (id: number) => Promise<void>;
  getCase: (id: number) => Promise<Case>;
  searchCases: (
    query: string,
    params?: Omit<CaseListRequest, "search">,
  ) => Promise<CaseListResponse>;

  // 案件管理
  assignLawyer: (
    caseId: number,
    lawyer_id: number,
  ) => Promise<{ message: string }>;
  updateStatus: (
    caseId: number,
    status: string,
    reason?: string,
  ) => Promise<{ message: string }>;
  getCaseStats: () => Promise<CaseStats>;

  // 工具方法
  refresh: () => Promise<void>;
  clearError: () => void;
}

/**
 * 案件管理 Hook
 * 提供案件相关的数据获取、CRUD 操作和状态管理
 */
export const useCases = (defaultParams?: CaseListRequest): UseCasesReturn => {
  const dispatch = useDispatch();

  // Redux 状态
  const { user } = useSelector((state: RootState) => state.auth);

  // 本地状态
  const [cases, setCases] = useState<Case[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [pagination, setPagination] = useState({
    page: 1,
    page_size: 10,
    total: 0,
    total_pages: 0,
  });
  const [stats, setStats] = useState<CaseStats | null>(null);

  // 获取案件列表
  const fetchCases = useCallback(
    async (params?: CaseListRequest) => {
      try {
        setLoading(true);
        setError(null);

        const mergedParams = { ...defaultParams, ...params };
        const response = await caseService.getCases(mergedParams);

        setCases(response.data);
        setPagination(response.pagination);
      } catch (err: any) {
        const appError =
          err instanceof AppError ? err : AppError.fromApiError(err);
        setError(appError.message);
        console.error("获取案件列表失败:", err);
      } finally {
        setLoading(false);
      }
    },
    [defaultParams],
  );

  // 创建案件
  const createCase = useCallback(
    async (data: CreateCaseRequest): Promise<Case> => {
      try {
        setLoading(true);
        setError(null);

        const newCase = await caseService.createCase(data);

        // 更新本地列表
        setCases((prev) => [newCase, ...prev]);

        return newCase;
      } catch (err: any) {
        const appError =
          err instanceof AppError ? err : AppError.fromApiError(err);
        setError(appError.message);
        console.error("创建案件失败:", err);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  // 更新案件
  const updateCase = useCallback(
    async (id: number, data: UpdateCaseRequest): Promise<Case> => {
      try {
        setLoading(true);
        setError(null);

        const updatedCase = await caseService.updateCase(id, data);

        // 更新本地列表
        setCases((prev) =>
          prev.map((caseItem) => (caseItem.id === id ? updatedCase : caseItem)),
        );

        return updatedCase;
      } catch (err: any) {
        const appError =
          err instanceof AppError ? err : AppError.fromApiError(err);
        setError(appError.message);
        console.error("更新案件失败:", err);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  // 删除案件
  const deleteCase = useCallback(async (id: number): Promise<void> => {
    try {
      setLoading(true);
      setError(null);

      await caseService.deleteCase(id);

      // 更新本地列表
      setCases((prev) => prev.filter((caseItem) => caseItem.id !== id));
    } catch (err: any) {
      const appError =
        err instanceof AppError ? err : AppError.fromApiError(err);
      setError(appError.message);
      console.error("删除案件失败:", err);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // 获取单个案件
  const getCase = useCallback(async (id: number): Promise<Case> => {
    try {
      setLoading(true);
      setError(null);

      return await caseService.getCase(id);
    } catch (err: any) {
      const appError =
        err instanceof AppError ? err : AppError.fromApiError(err);
      setError(appError.message);
      console.error("获取案件详情失败:", err);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // 搜索案件
  const searchCases = useCallback(
    async (
      query: string,
      params?: Omit<CaseListRequest, "search">,
    ): Promise<CaseListResponse> => {
      try {
        setLoading(true);
        setError(null);

        const response = await caseService.searchCases(query, params);

        // 更新本地列表
        setCases(response.data);
        setPagination(response.pagination);

        return response;
      } catch (err: any) {
        const appError =
          err instanceof AppError ? err : AppError.fromApiError(err);
        setError(appError.message);
        console.error("搜索案件失败:", err);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  // 分配律师
  const assignLawyer = useCallback(
    async (caseId: number, lawyer_id: number): Promise<{ message: string }> => {
      try {
        setLoading(true);
        setError(null);

        const result = await caseService.assignLawyer(caseId, lawyer_id);

        // 刷新案件列表
        await refresh();

        return result;
      } catch (err: any) {
        const appError =
          err instanceof AppError ? err : AppError.fromApiError(err);
        setError(appError.message);
        console.error("分配律师失败:", err);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  // 更新状态
  const updateStatus = useCallback(
    async (
      caseId: number,
      status: string,
      reason?: string,
    ): Promise<{ message: string }> => {
      try {
        setLoading(true);
        setError(null);

        const result = await caseService.updateCaseStatus(caseId, status);

        // 刷新案件列表
        await refresh();

        return result;
      } catch (err: any) {
        const appError =
          err instanceof AppError ? err : AppError.fromApiError(err);
        setError(appError.message);
        console.error("更新案件状态失败:", err);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  // 获取案件统计
  const getCaseStats = useCallback(async (): Promise<CaseStats> => {
    try {
      setLoading(true);
      setError(null);

      const statsData = await caseService.getCaseStats();
      setStats(statsData);

      return statsData;
    } catch (err: any) {
      const appError =
        err instanceof AppError ? err : AppError.fromApiError(err);
      setError(appError.message);
      console.error("获取案件统计失败:", err);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // 刷新数据
  const refresh = useCallback(async (): Promise<void> => {
    await fetchCases();
    if (stats) {
      await getCaseStats();
    }
  }, [fetchCases, stats, getCaseStats]);

  // 清除错误
  const clearError = useCallback((): void => {
    setError(null);
  }, []);

  // 初始化加载
  useEffect(() => {
    if (user) {
      fetchCases();
    }
  }, [user, fetchCases]);

  return {
    // 数据状态
    cases,
    loading,
    error,

    // 分页状态
    pagination,

    // 统计信息
    stats,

    // 操作方法
    fetchCases,
    createCase,
    updateCase,
    deleteCase,
    getCase,
    searchCases,

    // 案件管理
    assignLawyer,
    updateStatus,
    getCaseStats,

    // 工具方法
    refresh,
    clearError,
  };
};

export default useCases;
export type { UseCasesReturn };
