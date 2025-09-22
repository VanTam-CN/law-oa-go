import { useCallback } from "react";
import { useAppDispatch, useAppSelector } from "../store/hooks";
import {
  setGlobalLoading,
  setActionLoading,
  clearAllLoading,
  selectGlobalLoading,
  selectActionLoading,
} from "../store/slices/uiSlice";

interface UseLoadingOptions {
  timeout?: number;
  showTimeout?: boolean;
}

export interface UseLoadingReturn {
  // 状态
  globalLoading: boolean;

  // 全局loading方法
  setGlobal: (loading: boolean) => void;
  showGlobalLoading: () => void;
  hideGlobalLoading: () => void;

  // Action loading方法
  setAction: (key: string, loading: boolean) => void;
  showActionLoading: (key: string) => void;
  hideActionLoading: (key: string) => void;
  getActionLoading: (key: string) => boolean;

  // 工具方法
  clearAll: () => void;
  withLoading: <T>(asyncFn: () => Promise<T>, key?: string) => (() => Promise<T>);
  withParallelLoading: <T>(asyncFns: Array<() => Promise<T>>, key?: string) => (() => Promise<T[]>);
  withDebouncedLoading: <T>(asyncFn: () => Promise<T>, delay?: number, key?: string) => (() => Promise<T>);
}

export function useLoading(options: UseLoadingOptions = {}): UseLoadingReturn {
  const { timeout = 30000, showTimeout = true } = options;
  const dispatch = useAppDispatch();

  // 全局loading状态
  const globalLoading = useAppSelector(selectGlobalLoading);

  // 设置全局loading状态
  const setGlobal = useCallback(
    (loading: boolean) => {
      dispatch(setGlobalLoading(loading));
    },
    [dispatch],
  );

  // 显示全局loading
  const showGlobalLoading = useCallback(() => {
    setGlobal(true);

    if (showTimeout) {
      setTimeout(() => {
        setGlobal(false);
      }, timeout);
    }
  }, [setGlobal, timeout, showTimeout]);

  // 隐藏全局loading
  const hideGlobalLoading = useCallback(() => {
    setGlobal(false);
  }, [setGlobal]);

  // 设置特定action的loading状态
  const setAction = useCallback(
    (key: string, loading: boolean) => {
      dispatch(setActionLoading(key, loading));
    },
    [dispatch],
  );

  // 显示action loading
  const showActionLoading = useCallback(
    (key: string) => {
      setAction(key, true);

      if (showTimeout) {
        setTimeout(() => {
          setAction(key, false);
        }, timeout);
      }
    },
    [setAction, timeout, showTimeout],
  );

  // 隐藏action loading
  const hideActionLoading = useCallback(
    (key: string) => {
      setAction(key, false);
    },
    [setAction],
  );

  // 获取action的loading状态
  const getActionLoading = useCallback((key: string) => {
    return useAppSelector(selectActionLoading(key));
  }, []);

  // 清除所有loading状态
  const clearAll = useCallback(() => {
    dispatch(clearAllLoading());
  }, [dispatch]);

  // 包装异步函数，自动管理loading状态
  const withLoading = useCallback(
    <T>(asyncFn: () => Promise<T>, key?: string) => {
      return async (): Promise<T> => {
        try {
          if (key) {
            showActionLoading(key);
          } else {
            showGlobalLoading();
          }

          const result = await asyncFn();
          return result;
        } finally {
          if (key) {
            hideActionLoading(key);
          } else {
            hideGlobalLoading();
          }
        }
      };
    },
    [
      showActionLoading,
      showGlobalLoading,
      hideActionLoading,
      hideGlobalLoading,
    ],
  );

  // 包装多个异步函数，并行执行
  const withParallelLoading = useCallback(
    <T>(asyncFns: Array<() => Promise<T>>, key?: string) => {
      return async (): Promise<T[]> => {
        try {
          if (key) {
            showActionLoading(key);
          } else {
            showGlobalLoading();
          }

          const results = await Promise.all(asyncFns.map((fn) => fn()));
          return results;
        } finally {
          if (key) {
            hideActionLoading(key);
          } else {
            hideGlobalLoading();
          }
        }
      };
    },
    [
      showActionLoading,
      showGlobalLoading,
      hideActionLoading,
      hideGlobalLoading,
    ],
  );

  // 创建防抖loading
  const withDebouncedLoading = useCallback(
    <T>(asyncFn: () => Promise<T>, delay: number = 300, key?: string) => {
      let timeoutId: NodeJS.Timeout;

      return async (): Promise<T> => {
        try {
          if (timeoutId) {
            clearTimeout(timeoutId);
          }

          return new Promise((resolve, reject) => {
            timeoutId = setTimeout(async () => {
              try {
                if (key) {
                  showActionLoading(key);
                } else {
                  showGlobalLoading();
                }

                const result = await asyncFn();
                resolve(result);
              } catch (error) {
                reject(error);
              } finally {
                if (key) {
                  hideActionLoading(key);
                } else {
                  hideGlobalLoading();
                }
              }
            }, delay);
          });
        } catch (error) {
          throw error;
        }
      };
    },
    [
      showActionLoading,
      showGlobalLoading,
      hideActionLoading,
      hideGlobalLoading,
    ],
  );

  return {
    // 状态
    globalLoading,

    // 全局loading方法
    setGlobal,
    showGlobalLoading,
    hideGlobalLoading,

    // Action loading方法
    setAction,
    showActionLoading,
    hideActionLoading,
    getActionLoading,

    // 工具方法
    clearAll,
    withLoading,
    withParallelLoading,
    withDebouncedLoading,
  };
}

// 专门用于表单提交的loading hook
export function useFormLoading(
  formKey: string,
  options: UseLoadingOptions = {},
) {
  const loading = useLoading(options);

  const showFormLoading = useCallback(() => {
    loading.showActionLoading(`form_${formKey}`);
  }, [loading, formKey]);

  const hideFormLoading = useCallback(() => {
    loading.hideActionLoading(`form_${formKey}`);
  }, [loading, formKey]);

  const isFormLoading = useCallback(() => {
    return loading.getActionLoading(`form_${formKey}`);
  }, [loading, formKey]);

  const withFormLoading = useCallback(
    <T>(asyncFn: () => Promise<T>) => {
      return loading.withLoading(asyncFn, `form_${formKey}`);
    },
    [loading, formKey],
  );

  return {
    ...loading,
    showFormLoading,
    hideFormLoading,
    isFormLoading,
    withFormLoading,
  };
}

// 专门用于API请求的loading hook
export function useApiLoading(apiKey: string, options: UseLoadingOptions = {}) {
  const loading = useLoading(options);

  const showApiLoading = useCallback(() => {
    loading.showActionLoading(`api_${apiKey}`);
  }, [loading, apiKey]);

  const hideApiLoading = useCallback(() => {
    loading.hideActionLoading(`api_${apiKey}`);
  }, [loading, apiKey]);

  const isApiLoading = useCallback(() => {
    return loading.getActionLoading(`api_${apiKey}`);
  }, [loading, apiKey]);

  const withApiLoading = useCallback(
    <T>(asyncFn: () => Promise<T>) => {
      return loading.withLoading(asyncFn, `api_${apiKey}`);
    },
    [loading, apiKey],
  );

  return {
    ...loading,
    showApiLoading,
    hideApiLoading,
    isApiLoading,
    withApiLoading,
  };
}
