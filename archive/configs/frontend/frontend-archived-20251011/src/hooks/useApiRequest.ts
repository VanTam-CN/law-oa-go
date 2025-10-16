import { useState, useEffect, useCallback } from "react";
import { useAppDispatch } from "../store/hooks";
import {
  setActionLoading,
  setApiError,
  clearApiError,
} from "../store/slices/apiSlice";
import apiClient from "../services/api";

interface UseApiRequestOptions {
  showError?: boolean;
  showLoading?: boolean;
  retryOnFailure?: boolean;
  cacheKey?: string;
  cacheTTL?: number;
}

export interface UseApiRequestReturn<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
  execute: (params?: any) => Promise<T>;
  refetch: () => Promise<T>;
  clearError: () => void;
  clearData: () => void;
}

export function useApiRequest<T>(
  requestFn: (params?: any) => Promise<T>,
  options: UseApiRequestOptions = {},
): UseApiRequestReturn<T> {
  const {
    showError = true,
    showLoading = true,
    retryOnFailure = false,
    cacheKey,
    cacheTTL,
  } = options;

  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<string | null>(null);
  const dispatch = useAppDispatch();

  const execute = useCallback(
    async (params?: any): Promise<T> => {
      try {
        if (showLoading) {
          dispatch(setActionLoading(cacheKey || requestFn.name, true));
        }

        dispatch(clearApiError(cacheKey || requestFn.name));
        setError(null);

        // 如果有缓存配置，使用缓存的API客户端
        let result: T;
        if (cacheKey) {
          result = (await apiClient.get(
            typeof params === "string" ? params : "",
            { useCache: true, cacheTTL },
          )) as T;
        } else {
          result = await requestFn(params);
        }

        setData(result);
        return result;
      } catch (error: any) {
        const errorMessage = error.message || "请求失败";
        setError(errorMessage);

        if (showError) {
          dispatch(
            setApiError({
              key: cacheKey || requestFn.name,
              error: errorMessage,
            }),
          );
        }

        if (retryOnFailure) {
          // 可以在这里添加重试逻辑
          console.warn("请求失败，将进行重试:", errorMessage);
        }

        throw error;
      } finally {
        if (showLoading) {
          dispatch(setActionLoading(cacheKey || requestFn.name, false));
        }
      }
    },
    [
      requestFn,
      showLoading,
      showError,
      retryOnFailure,
      cacheKey,
      cacheTTL,
      dispatch,
    ],
  );

  const refetch = useCallback(async (): Promise<T> => {
    return execute();
  }, [execute]);

  const clearError = useCallback(() => {
    setError(null);
    dispatch(clearApiError(cacheKey || requestFn.name));
  }, [cacheKey, requestFn.name, dispatch]);

  const clearData = useCallback(() => {
    setData(null);
  }, []);

  return {
    data,
    loading: false, // 可以从Redux状态中获取
    error,
    execute,
    refetch,
    clearError,
    clearData,
  };
}

// 用于自动触发API请求的hook
export function useApiAutoRequest<T>(
  requestFn: (params?: any) => Promise<T>,
  params?: any,
  options: UseApiRequestOptions & {
    deps?: any[];
    immediate?: boolean;
  } = {},
): UseApiRequestReturn<T> {
  const { deps = [], immediate = true, ...requestOptions } = options;

  const { execute, ...rest } = useApiRequest<T>(requestFn, requestOptions);

  useEffect(() => {
    if (immediate) {
      execute(params).catch(console.error);
    }
  }, deps); // eslint-disable-line react-hooks/exhaustive-deps

  return {
    execute,
    ...rest,
  };
}

// 用于分页请求的hook
export function usePaginatedApiRequest<T>(
  requestFn: (
    page: number,
    pageSize: number,
    params?: any,
  ) => Promise<{
    data: T[];
    pagination: {
      page: number;
      page_size: number;
      total: number;
      total_pages: number;
    };
  }>,
  options: UseApiRequestOptions & {
    initialPageSize?: number;
  } = {},
) {
  const { initialPageSize = 20, ...requestOptions } = options;

  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(initialPageSize);
  const [pagination, setPagination] = useState({
    page: 1,
    page_size: initialPageSize,
    total: 0,
    total_pages: 0,
  });

  const { data, loading, error, execute, clearError } = useApiRequest(
    (params?: any) => requestFn(currentPage, pageSize, params),
    requestOptions,
  );

  const loadData = useCallback(
    async (
      page: number = currentPage,
      size: number = pageSize,
      params?: any,
    ) => {
      const result = await execute({ ...params, page, page_size: size });
      setPagination(result.pagination);
      setCurrentPage(page);
      setPageSize(size);
      return result;
    },
    [execute, currentPage, pageSize],
  );

  const goToPage = useCallback(
    (page: number) => {
      return loadData(page, pageSize);
    },
    [loadData, pageSize],
  );

  const changePageSize = useCallback(
    (newSize: number) => {
      return loadData(1, newSize);
    },
    [loadData],
  );

  const nextPage = useCallback(() => {
    if (currentPage < pagination.total_pages) {
      return goToPage(currentPage + 1);
    }
  }, [currentPage, pagination.total_pages, goToPage]);

  const prevPage = useCallback(() => {
    if (currentPage > 1) {
      return goToPage(currentPage - 1);
    }
  }, [currentPage, goToPage]);

  return {
    data: data?.data || [],
    loading,
    error,
    pagination,
    currentPage,
    pageSize,
    loadData,
    goToPage,
    changePageSize,
    nextPage,
    prevPage,
    clearError,
  };
}
