import { createSlice, createAsyncThunk, PayloadAction } from "@reduxjs/toolkit";
import apiClient from "../../services/api";

interface CacheEntry<T> {
  data: T;
  timestamp: number;
  ttl: number; // Time to live in milliseconds
}

interface ApiState {
  cache: Record<string, CacheEntry<any>>;
  pendingRequests: Record<string, Promise<any>>;
  errors: Record<string, string>;
  lastFetch: Record<string, number>;
}

const initialState: ApiState = {
  cache: {},
  pendingRequests: {},
  errors: {},
  lastFetch: {},
};

// 清理过期缓存
const cleanExpiredCache = (cache: Record<string, CacheEntry<any>>) => {
  const now = Date.now();
  Object.keys(cache).forEach((key) => {
    if (now - cache[key].timestamp > cache[key].ttl) {
      delete cache[key];
    }
  });
};

// 生成缓存键
const generateCacheKey = (endpoint: string, params?: any): string => {
  return params ? `${endpoint}_${JSON.stringify(params)}` : endpoint;
};

const apiSlice = createSlice({
  name: "api",
  initialState,
  reducers: {
    // Cache management
    setCache: (
      state,
      action: PayloadAction<{ key: string; data: any; ttl?: number }>,
    ) => {
      const { key, data, ttl = 5 * 60 * 1000 } = action.payload; // Default 5 minutes
      state.cache[key] = {
        data,
        timestamp: Date.now(),
        ttl,
      };
    },
    clearCache: (state, action: PayloadAction<string>) => {
      delete state.cache[action.payload];
    },
    clearAllCache: (state) => {
      state.cache = {};
    },
    clearCacheByPattern: (state, action: PayloadAction<string>) => {
      const pattern = action.payload;
      Object.keys(state.cache).forEach((key) => {
        if (key.includes(pattern)) {
          delete state.cache[key];
        }
      });
    },

    // Pending requests
    setPendingRequest: (
      state,
      action: PayloadAction<{ key: string; request: Promise<any> }>,
    ) => {
      state.pendingRequests[action.payload.key] = action.payload.request;
    },
    removePendingRequest: (state, action: PayloadAction<string>) => {
      delete state.pendingRequests[action.payload];
    },
    clearAllPendingRequests: (state) => {
      state.pendingRequests = {};
    },

    // Error management
    setApiError: (
      state,
      action: PayloadAction<{ key: string; error: string }>,
    ) => {
      state.errors[action.payload.key] = action.payload.error;
    },
    clearApiError: (state, action: PayloadAction<string>) => {
      delete state.errors[action.payload];
    },
    clearAllApiErrors: (state) => {
      state.errors = {};
    },

    // Last fetch tracking
    setLastFetch: (
      state,
      action: PayloadAction<{ key: string; timestamp: number }>,
    ) => {
      state.lastFetch[action.payload.key] = action.payload.timestamp;
    },
    clearLastFetch: (state, action: PayloadAction<string>) => {
      delete state.lastFetch[action.payload];
    },
  },
  extraReducers: (builder) => {
    // 可以在这里添加通用的API action处理
  },
});

export const {
  setCache,
  clearCache,
  clearAllCache,
  clearCacheByPattern,
  setPendingRequest,
  removePendingRequest,
  clearAllPendingRequests,
  setApiError,
  clearApiError,
  clearAllApiErrors,
  setLastFetch,
  clearLastFetch,
} = apiSlice.actions;

// Selectors
export const selectCache = (key: string) => (state: { api: ApiState }) =>
  state.api.cache[key];
export const selectCacheData = (key: string) => (state: { api: ApiState }) =>
  state.api.cache[key]?.data;
export const selectIsCached = (key: string) => (state: { api: ApiState }) =>
  !!state.api.cache[key];
export const selectIsCacheValid =
  (key: string) => (state: { api: ApiState }) => {
    const cache = state.api.cache[key];
    if (!cache) return false;
    return Date.now() - cache.timestamp < cache.ttl;
  };
export const selectPendingRequest =
  (key: string) => (state: { api: ApiState }) =>
    state.api.pendingRequests[key];
export const selectApiError = (key: string) => (state: { api: ApiState }) =>
  state.api.errors[key];
export const selectLastFetch = (key: string) => (state: { api: ApiState }) =>
  state.api.lastFetch[key];

// Action creators for loading state
export const setActionLoading = (key: string, loading: boolean) => ({
  type: "api/setActionLoading",
  payload: { key, loading },
});

export const setGlobalLoading = (loading: boolean) => ({
  type: "api/setGlobalLoading",
  payload: loading,
});

export const clearAllLoading = () => ({
  type: "api/clearAllLoading",
});

export const selectGlobalLoading = (state: { api: ApiState }) =>
  Object.keys(state.api.pendingRequests).length > 0;

export const selectActionLoading =
  (key: string) => (state: { api: ApiState }) =>
    !!state.api.pendingRequests[key];

// Async thunk factory for API calls with caching
export const createApiThunk = <T, P = any>(
  typePrefix: string,
  apiCall: (params: P) => Promise<T>,
  options: {
    cacheKey?: (params: P) => string;
    cacheTTL?: number;
    shouldCache?: boolean;
    shouldDeduplicate?: boolean;
  } = {},
) => {
  const {
    cacheKey = (params: P) => generateCacheKey(typePrefix, params),
    cacheTTL = 5 * 60 * 1000, // 5 minutes default
    shouldCache = true,
    shouldDeduplicate = true,
  } = options;

  return createAsyncThunk(
    `api/${typePrefix}`,
    async (params: P, { getState, dispatch, rejectWithValue }) => {
      const key = cacheKey(params);
      const state = getState() as { api: ApiState };

      // 清理过期缓存
      cleanExpiredCache(state.api.cache);

      // 检查是否有有效的缓存
      if (shouldCache && selectIsCacheValid(key)(state)) {
        return selectCacheData(key)(state);
      }

      // 检查是否有重复请求
      if (shouldDeduplicate) {
        const pendingRequest = selectPendingRequest(key)(state);
        if (pendingRequest) {
          try {
            return await pendingRequest;
          } catch (error) {
            // 如果pending请求失败，继续执行新的请求
          }
        }
      }

      // 创建请求
      const requestPromise = apiCall(params).catch((error) => {
        dispatch(removePendingRequest(key));
        dispatch(setApiError({ key, error: error.message }));
        throw error;
      });

      // 设置pending请求
      dispatch(setPendingRequest({ key, request: requestPromise }));

      try {
        const result = await requestPromise;

        // 缓存结果
        if (shouldCache) {
          dispatch(setCache({ key, data: result, ttl: cacheTTL }));
        }

        // 清理pending和error
        dispatch(removePendingRequest(key));
        dispatch(clearApiError(key));
        dispatch(setLastFetch({ key, timestamp: Date.now() }));

        return result;
      } catch (error: any) {
        return rejectWithValue(error.message || "API请求失败");
      }
    },
  );
};

export default apiSlice.reducer;
