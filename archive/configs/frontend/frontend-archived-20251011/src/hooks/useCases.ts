/**
 * 案件管理自定义Hook
 * 提供数据缓存、防抖搜索、智能重试和性能优化功能
 */

import { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import { getCases } from '../services/caseService';
import { CaseListRequest } from '../types';
import { Case } from '../types';
import { validateApiResponse } from '../utils/validation';
import errorHandler from '../utils/errorHandler';

interface UseCasesOptions {
  pageSize?: number;
  cacheTime?: number; // 缓存时间（毫秒）
  retryCount?: number; // 重试次数
  retryDelay?: number; // 重试延迟（毫秒）
  enableCache?: boolean; // 是否启用缓存
}

interface UseCasesResult {
  cases: Case[];
  loading: boolean;
  error: string | null;
  pagination: {
    current: number;
    pageSize: number;
    total: number;
    totalPages: number;
  };
  hasMore: boolean;
  refresh: () => Promise<void>;
  loadMore: () => Promise<void>;
  search: (params: CaseListRequest) => Promise<void>;
  reset: () => void;
}

// 缓存接口
interface CacheEntry<T> {
  data: T;
  timestamp: number;
  key: string;
}

class DataCache {
  private cache = new Map<string, CacheEntry<any>>();
  private defaultTTL = 5 * 60 * 1000; // 5分钟默认缓存时间

  set<T>(key: string, data: T, ttl?: number): void {
    const entry: CacheEntry<T> = {
      data,
      timestamp: Date.now(),
      key
    };

    this.cache.set(key, entry);

    // 设置过期时间
    const expirationTime = ttl || this.defaultTTL;
    setTimeout(() => {
      this.cache.delete(key);
    }, expirationTime);
  }

  get<T>(key: string, maxAge?: number): T | null {
    const entry = this.cache.get(key);
    if (!entry) {
      return null;
    }

    const age = Date.now() - entry.timestamp;
    const maxAgeMs = maxAge || this.defaultTTL;

    if (age > maxAgeMs) {
      this.cache.delete(key);
      return null;
    }

    return entry.data as T;
  }

  clear(): void {
    this.cache.clear();
  }

  delete(key: string): void {
    this.cache.delete(key);
  }

  size(): number {
    return this.cache.size;
  }

  // 清理过期缓存
  cleanup(): void {
    const now = Date.now();
    for (const [key, entry] of this.cache.entries()) {
      const age = now - entry.timestamp;
      if (age > this.defaultTTL) {
        this.cache.delete(key);
      }
    }
  }
}

// 全局缓存实例
const globalCache = new DataCache();

// 生成缓存键
function generateCacheKey(params: CaseListRequest): string {
  const keyParts = [
    'cases',
    `page:${params.page || 1}`,
    `pageSize:${params.page_size || 10}`,
    `search:${params.search || ''}`,
    `type:${params.case_type || ''}`,
    `status:${params.status || ''}`,
    `priority:${params.priority || ''}`,
    `client:${params.client_id || ''}`,
    `lawyer:${params.lawyer_id || ''}`
  ];
  return keyParts.join('|');
}

// 防抖函数
function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: NodeJS.Timeout;
  return (...args: Parameters<T>) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => func(...args), wait);
  };
}

// 智能重试函数
async function retryWithBackoff<T>(
  operation: () => Promise<T>,
  maxRetries: number = 3,
  baseDelay: number = 1000
): Promise<T> {
  let lastError: Error;

  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      return await operation();
    } catch (error) {
      lastError = error as Error;

      if (attempt === maxRetries) {
        throw lastError;
      }

      // 指数退避策略
      const delay = baseDelay * Math.pow(2, attempt - 1);
      await new Promise(resolve => setTimeout(resolve, delay));
    }
  }

  throw lastError!;
}

export const useCases = (options: UseCasesOptions = {}): UseCasesResult => {
  const {
    pageSize = 10,
    cacheTime = 5 * 60 * 1000, // 5分钟
    retryCount = 3,
    retryDelay = 1000,
    enableCache = true
  } = options;

  // 状态管理
  const [cases, setCases] = useState<Case[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize,
    total: 0,
    totalPages: 0
  });

  // Refs用于避免闭包问题
  const currentParamsRef = useRef<CaseListRequest>({});
  const abortControllerRef = useRef<AbortController | null>(null);

  // 计算是否还有更多数据
  const hasMore = useMemo(() => {
    return pagination.current < pagination.totalPages;
  }, [pagination.current, pagination.totalPages]);

  // 获取数据的核心函数
  const fetchCasesData = useCallback(async (
    params: CaseListRequest,
    useCache: boolean = enableCache
  ) => {
    // 取消之前的请求
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }

    // 创建新的AbortController
    abortControllerRef.current = new AbortController();

    setLoading(true);
    setError(null);

    const finalParams = {
      page: params.page || 1,
      page_size: params.page_size || pageSize,
      ...params
    };

    currentParamsRef.current = finalParams;

    try {
      // 检查缓存
      if (useCache) {
        const cacheKey = generateCacheKey(finalParams);
        const cachedData = globalCache.get(cacheKey, cacheTime);

        if (cachedData && typeof cachedData === 'object' && cachedData !== null && 'data' in cachedData) {
          const typedData = cachedData as { data: Case[], pagination: any };
          setCases(typedData.data);
          setPagination(typedData.pagination);
          setLoading(false);
          return cachedData;
        }
      }

      // 使用智能重试机制
      const response = await retryWithBackoff(
        () => getCases(finalParams),
        retryCount,
        retryDelay
      );

      // 检查请求是否被取消
      if (abortControllerRef.current?.signal.aborted) {
        return;
      }

      // 验证API响应
      const validation = validateApiResponse(response, 'list');
      if (!validation.isValid) {
        console.error('API响应验证失败:', validation.errors);
        errorHandler.showWarning('数据格式异常，部分功能可能受影响', '数据验证');
      }

      // 验证案件数据
      const invalidCases = response.data?.filter((caseItem: Case) => {
        // 简单验证：检查必要字段
        return !caseItem.id || !caseItem.title;
      }) || [];

      if (invalidCases.length > 0) {
        console.warn(`发现 ${invalidCases.length} 条无效数据`);
      }

      const resultData = {
        data: response.data || [],
        pagination: {
          current: response.pagination?.page || 1,
          pageSize: response.pagination?.page_size || pageSize,
          total: response.pagination?.total || 0,
          totalPages: response.pagination?.total_pages || 0
        }
      };

      // 更新状态
      setCases(resultData.data);
      setPagination(resultData.pagination);

      // 缓存结果
      if (useCache) {
        const cacheKey = generateCacheKey(finalParams);
        globalCache.set(cacheKey, resultData, cacheTime);
      }

      return resultData;

    } catch (err) {
      // 检查是否是取消错误
      if (err instanceof Error && err.name === 'AbortError') {
        return;
      }

      console.error('获取案件数据失败:', err);
      const errorMessage = err instanceof Error ? err.message : '获取数据失败';
      setError(errorMessage);

      // 使用全局错误处理器
      errorHandler.handleNetworkError(err as Error, '获取案件数据');

      throw err;
    } finally {
      setLoading(false);
    }
  }, [pageSize, cacheTime, retryCount, retryDelay, enableCache]);

  // 防抖搜索函数
  const debouncedSearch = useMemo(
    () => debounce(async (params: CaseListRequest) => {
      await fetchCasesData(params, true);
    }, 300),
    [fetchCasesData]
  );

  // 搜索函数
  const search = useCallback(async (params: CaseListRequest) => {
    const finalParams = {
      ...currentParamsRef.current,
      ...params,
      page: 1 // 搜索时重置到第一页
    };

    await debouncedSearch(finalParams);
  }, [debouncedSearch]);

  // 刷新函数
  const refresh = useCallback(async () => {
    await fetchCasesData(currentParamsRef.current, false); // 刷新时不使用缓存
  }, [fetchCasesData]);

  // 加载更多函数
  const loadMore = useCallback(async () => {
    if (!hasMore || loading) {
      return;
    }

    const nextPage = pagination.current + 1;
    const loadMoreParams = {
      ...currentParamsRef.current,
      page: nextPage
    };

    try {
      const newData = await fetchCasesData(loadMoreParams, true);

      if (newData && typeof newData === 'object' && newData !== null && 'data' in newData) {
        const typedData = newData as { data: Case[], pagination: any };
        // 追加新数据到现有数据
        setCases(prev => [...prev, ...typedData.data]);
        setPagination(typedData.pagination);
      }
    } catch (err) {
      console.error('加载更多数据失败:', err);
    }
  }, [hasMore, loading, pagination.current, fetchCasesData]);

  // 重置函数
  const reset = useCallback(() => {
    setCases([]);
    setError(null);
    setPagination({
      current: 1,
      pageSize,
      total: 0,
      totalPages: 0
    });
    currentParamsRef.current = {};
  }, [pageSize]);

  // 初始化加载
  useEffect(() => {
    fetchCasesData({ page: 1, page_size: pageSize }, true);
  }, [fetchCasesData, pageSize]);

  // 清理函数
  useEffect(() => {
    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, []);

  // 定期清理缓存
  useEffect(() => {
    const cleanupInterval = setInterval(() => {
      globalCache.cleanup();
    }, 60 * 1000); // 每分钟清理一次

    return () => {
      clearInterval(cleanupInterval);
    };
  }, []);

  return {
    cases,
    loading,
    error,
    pagination,
    hasMore,
    refresh,
    loadMore,
    search,
    reset
  };
};

// 导出缓存实例以供其他组件使用
export { globalCache };

// 导出缓存清理函数
export const clearAllCache = () => {
  globalCache.clear();
};

// 预加载函数
export const preloadCases = async (params: CaseListRequest, options?: UseCasesOptions) => {
  const cacheKey = generateCacheKey(params);
  const cachedData = globalCache.get(cacheKey);

  if (cachedData) {
    return cachedData;
  }

  try {
    const response = await getCases(params);
    const resultData = {
      data: response.data || [],
      pagination: {
        current: response.pagination?.page || 1,
        pageSize: response.pagination?.page_size || 10,
        total: response.pagination?.total || 0,
        totalPages: response.pagination?.total_pages || 0
      }
    };

    const cacheTime = options?.cacheTime || 5 * 60 * 1000;
    globalCache.set(cacheKey, resultData, cacheTime);

    return resultData;
  } catch (error) {
    console.error('预加载案件数据失败:', error);
    throw error;
  }
};

export default useCases;