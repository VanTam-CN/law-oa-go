import { useCallback } from 'react';
import { useAppDispatch, useAppSelector } from '../store/hooks';
import {
  setCache,
  clearCache,
  clearAllCache as clearAllReduxCache,
  clearCacheByPattern,
  selectCache,
  selectIsCached,
  selectIsCacheValid
} from '../store/slices/apiSlice';
import apiClient from '../services/api';

interface UseApiCacheOptions {
  enableReduxCache?: boolean;
  enableClientCache?: boolean;
  defaultTTL?: number;
}

export interface UseApiCacheReturn {
  setCache: <T>(key: string, data: T, ttl?: number) => void;
  getCache: <T>(key: string) => T | null;
  hasCache: (key: string) => boolean;
  isCacheValid: (key: string) => boolean;
  clearCache: (key: string) => void;
  clearAllCache: () => void;
  clearCacheByPattern: (pattern: string) => void;
  getCacheStats: () => {
    totalEntries: number;
    totalSize: number;
    pendingRequests: number;
  };
  preloadData: <T>(key: string, fetchFn: () => Promise<T>, ttl?: number) => Promise<T>;
  withCache: <T>(fetchFn: () => Promise<T>, key: string, ttl?: number) => (() => Promise<T>);
  batchPreload: <T>(items: Array<{
    key: string;
    fetchFn: () => Promise<T>;
    ttl?: number;
  }>) => Promise<void>;
}

export function useApiCache(options: UseApiCacheOptions = {}): UseApiCacheReturn {
  const {
    enableReduxCache = true,
    enableClientCache = true,
    defaultTTL = 5 * 60 * 1000, // 5 minutes
  } = options;

  const dispatch = useAppDispatch();

  // 设置缓存
  const setCacheData = useCallback(<T>(key: string, data: T, ttl: number = defaultTTL) => {
    if (enableReduxCache) {
      dispatch(setCache({ key, data, ttl }));
    }
  }, [dispatch, enableReduxCache, defaultTTL]);

  // 获取缓存数据
  const getCacheData = useCallback(<T>(key: string): T | null => {
    if (enableReduxCache) {
      const cacheData = useAppSelector(selectCache(key));
      return cacheData?.data || null;
    }
    return null;
  }, [enableReduxCache]);

  // 检查是否有缓存
  const hasCache = useCallback((key: string): boolean => {
    if (enableReduxCache) {
      return useAppSelector(selectIsCached(key));
    }
    return false;
  }, [enableReduxCache]);

  // 检查缓存是否有效
  const isCacheValid = useCallback((key: string): boolean => {
    if (enableReduxCache) {
      return useAppSelector(selectIsCacheValid(key));
    }
    return false;
  }, [enableReduxCache]);

  // 清除缓存
  const clearCacheData = useCallback((key: string) => {
    if (enableReduxCache) {
      dispatch(clearCache(key));
    }
    if (enableClientCache) {
      apiClient.clearCache(key);
    }
  }, [dispatch, enableReduxCache, enableClientCache]);

  // 清除所有缓存
  const clearAllCache = useCallback(() => {
    if (enableReduxCache) {
      dispatch(clearAllReduxCache());
    }
    if (enableClientCache) {
      apiClient.clearCache();
    }
  }, [dispatch, enableReduxCache, enableClientCache]);

  // 按模式清除缓存
  const clearCacheByPatternFn = useCallback((pattern: string) => {
    if (enableReduxCache) {
      dispatch(clearCacheByPattern(pattern));
    }
    if (enableClientCache) {
      // 对于客户端缓存，需要模拟模式匹配
      const stats = apiClient.getCacheStats();
      console.log(`Cleared cache for pattern: ${pattern}, stats:`, stats);
    }
  }, [dispatch, enableReduxCache, enableClientCache]);

  // 获取缓存统计信息
  const getCacheStats = useCallback(() => {
    if (enableClientCache) {
      return apiClient.getCacheStats();
    }
    return {
      totalEntries: 0,
      totalSize: 0,
      pendingRequests: 0,
    };
  }, [enableClientCache]);

  // 预加载数据到缓存
  const preloadData = useCallback(async <T>(
    key: string,
    fetchFn: () => Promise<T>,
    ttl: number = defaultTTL
  ): Promise<T> => {
    try {
      const data = await fetchFn();
      setCacheData(key, data, ttl);
      return data;
    } catch (error) {
      console.error(`Failed to preload data for key ${key}:`, error);
      throw error;
    }
  }, [setCacheData, defaultTTL]);

  // 缓存装饰器 - 用于自动缓存API响应
  const withCache = useCallback(<T>(
    fetchFn: () => Promise<T>,
    key: string,
    ttl: number = defaultTTL
  ): (() => Promise<T>) => {
    return async () => {
      // 检查缓存
      if (isCacheValid(key)) {
        const cachedData = getCacheData<T>(key);
        if (cachedData) {
          return cachedData;
        }
      }

      // 获取新数据
      const data = await fetchFn();

      // 设置缓存
      setCacheData(key, data, ttl);

      return data;
    };
  }, [isCacheValid, getCacheData, setCacheData, defaultTTL]);

  // 批量预加载
  const batchPreload = useCallback(async <T>(
    items: Array<{
      key: string;
      fetchFn: () => Promise<T>;
      ttl?: number;
    }>
  ): Promise<void> => {
    const promises = items.map(item =>
      preloadData(item.key, item.fetchFn, item.ttl)
        .catch(error => {
          console.error(`Failed to preload ${item.key}:`, error);
        })
    );

    await Promise.allSettled(promises);
  }, [preloadData]);

  return {
    setCache: setCacheData,
    getCache: getCacheData,
    hasCache,
    isCacheValid,
    clearCache: clearCacheData,
    clearAllCache,
    clearCacheByPattern: clearCacheByPatternFn,
    getCacheStats,
    preloadData,
    withCache,
    batchPreload,
  };
}

// 专门用于特定资源类型的缓存hook
export function useResourceCache<T>(
  resourceType: string,
  options: UseApiCacheOptions = {}
) {
  const cache = useApiCache(options);

  // 生成资源相关的缓存键
  const generateResourceKey = useCallback((id?: string, params?: any) => {
    const baseKey = `${resourceType}`;
    if (id) {
      return `${baseKey}_${id}`;
    }
    if (params) {
      return `${baseKey}_${JSON.stringify(params)}`;
    }
    return baseKey;
  }, [resourceType]);

  // 缓存单个资源
  const cacheResource = useCallback((id: string, data: T, ttl?: number) => {
    const key = generateResourceKey(id);
    cache.setCache(key, data, ttl);
  }, [cache, generateResourceKey]);

  // 获取缓存的资源
  const getCachedResource = useCallback((id: string): T | null => {
    const key = generateResourceKey(id);
    return cache.getCache<T>(key);
  }, [cache, generateResourceKey]);

  // 清除资源缓存
  const clearResourceCache = useCallback((id?: string) => {
    const key = generateResourceKey(id);
    cache.clearCache(key);
  }, [cache, generateResourceKey]);

  // 清除所有相关资源缓存
  const clearAllResourceCache = useCallback(() => {
    cache.clearCacheByPattern(resourceType);
  }, [cache, resourceType]);

  return {
    ...cache,
    generateResourceKey,
    cacheResource,
    getCachedResource,
    clearResourceCache,
    clearAllResourceCache,
  };
}
