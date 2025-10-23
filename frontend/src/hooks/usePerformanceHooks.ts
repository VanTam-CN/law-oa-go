/**
 * 性能优化的自定义Hook示例
 * 展示React最佳实践：useMemo, useCallback, 和稳定引用
 */

import { useMemo, useCallback, useRef, useEffect, useState } from 'react'
import { debounce, throttle } from 'lodash'

// 通用的异步操作Hook
export const useAsyncOperation = <T, E = Error>(
  asyncFn: () => Promise<T>
) => {
  const [data, setData] = useState<T | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<E | null>(null)

  const execute = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const result = await asyncFn()
      setData(result)
      return result
    } catch (err) {
      setError(err as E)
      throw err
    } finally {
      setLoading(false)
    }
  }, [asyncFn])

  const reset = useCallback(() => {
    setData(null)
    setError(null)
    setLoading(false)
  }, [])

  return {
    data,
    loading,
    error,
    execute,
    reset
  }
}

// 防抖Hook
export const useDebounce = <T>(value: T, delay: number): T => {
  const [debouncedValue, setDebouncedValue] = useState<T>(value)

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value)
    }, delay)

    return () => {
      clearTimeout(handler)
    }
  }, [value, delay])

  return debouncedValue
}

// 节流Hook
export const useThrottle = <T extends (...args: any[]) => any>(
  callback: T,
  delay: number
): T => {
  const callbackRef = useRef(callback)
  callbackRef.current = callback

  return useMemo(
    () =>
      throttle((...args: Parameters<T>) => {
        callbackRef.current(...args)
      }, delay) as T,
    [delay]
  )
}

// 防抖回调Hook
export const useDebouncedCallback = <T extends (...args: any[]) => any>(
  callback: T,
  delay: number,
  deps: React.DependencyList = []
): T => {
  const callbackRef = useRef(callback)
  callbackRef.current = callback

  const debouncedCallback = useMemo(
    () =>
      debounce((...args: Parameters<T>) => {
        callbackRef.current(...args)
      }, delay),
    [delay, ...deps]
  )

  useEffect(() => {
    return () => {
      debouncedCallback.cancel()
    }
  }, [debouncedCallback])

  return debouncedCallback as T
}

// 本地存储Hook（带类型安全）
export const useLocalStorage = <T>(
  key: string,
  initialValue: T
): [T, (value: T | ((val: T) => T)) => void] => {
  const [storedValue, setStoredValue] = useState<T>(() => {
    try {
      const item = window.localStorage.getItem(key)
      return item ? JSON.parse(item) : initialValue
    } catch (error) {
      console.error(`Error reading localStorage key "${key}":`, error)
      return initialValue
    }
  })

  const setValue = useCallback(
    (value: T | ((val: T) => T)) => {
      try {
        const valueToStore =
          value instanceof Function ? value(storedValue) : value
        setStoredValue(valueToStore)
        window.localStorage.setItem(key, JSON.stringify(valueToStore))
      } catch (error) {
        console.error(`Error setting localStorage key "${key}":`, error)
      }
    },
    [key, storedValue]
  )

  return [storedValue, setValue]
}

// 窗口大小Hook
export const useWindowSize = () => {
  const [windowSize, setWindowSize] = useState({
    width: 0,
    height: 0,
  })

  useEffect(() => {
    const handleResize = () => {
      setWindowSize({
        width: window.innerWidth,
        height: window.innerHeight,
      })
    }

    handleResize()
    window.addEventListener('resize', handleResize)

    return () => window.removeEventListener('resize', handleResize)
  }, [])

  return useMemo(() => ({
    ...windowSize,
    isMobile: windowSize.width < 768,
    isTablet: windowSize.width >= 768 && windowSize.width < 1024,
    isDesktop: windowSize.width >= 1024,
  }), [windowSize])
}

// 媒体查询Hook
export const useMediaQuery = (query: string): boolean => {
  const [matches, setMatches] = useState(false)

  useEffect(() => {
    const media = window.matchMedia(query)
    if (media.matches !== matches) {
      setMatches(media.matches)
    }

    const listener = () => setMatches(media.matches)
    media.addEventListener('change', listener)

    return () => media.removeEventListener('change', listener)
  }, [matches, query])

  return matches
}

// 上次值Hook
export const usePrevious = <T>(value: T): T | undefined => {
  const ref = useRef<T>()
  useEffect(() => {
    ref.current = value
  })
  return ref.current
}

// 计数器Hook
export const useCounter = (initialValue = 0) => {
  const [count, setCount] = useState(initialValue)

  const increment = useCallback(() => {
    setCount(prev => prev + 1)
  }, [])

  const decrement = useCallback(() => {
    setCount(prev => prev - 1)
  }, [])

  const reset = useCallback(() => {
    setCount(initialValue)
  }, [initialValue])

  const set = useCallback((value: number) => {
    setCount(value)
  }, [])

  return useMemo(() => ({
    count,
    increment,
    decrement,
    reset,
    set,
    isPositive: count > 0,
    isNegative: count < 0,
    isZero: count === 0,
  }), [count, increment, decrement, reset, set])
}

// 数组操作Hook
export const useArray = <T>(initialValue: T[] = []) => {
  const [array, setArray] = useState<T[]>(initialValue)

  const push = useCallback((element: T) => {
    setArray(prev => [...prev, element])
  }, [])

  const filter = useCallback((predicate: (item: T) => boolean) => {
    setArray(prev => prev.filter(predicate))
  }, [])

  const update = useCallback((index: number, element: T) => {
    setArray(prev => {
      const newArray = [...prev]
      newArray[index] = element
      return newArray
    })
  }, [])

  const remove = useCallback((index: number) => {
    setArray(prev => prev.filter((_, i) => i !== index))
  }, [])

  const clear = useCallback(() => {
    setArray([])
  }, [])

  return useMemo(() => ({
    array,
    set: setArray,
    push,
    filter,
    update,
    remove,
    clear,
    isEmpty: array.length === 0,
    length: array.length,
  }), [array, push, filter, update, remove, clear])
}

// 布尔值切换Hook
export const useToggle = (initialValue = false) => {
  const [value, setValue] = useState(initialValue)

  const toggle = useCallback(() => {
    setValue(prev => !prev)
  }, [])

  const setTrue = useCallback(() => {
    setValue(true)
  }, [])

  const setFalse = useCallback(() => {
    setValue(false)
  }, [])

  return useMemo(() => ({
    value,
    setValue,
    toggle,
    setTrue,
    setFalse,
  }), [value, setValue, toggle, setTrue, setFalse])
}

// 表单状态Hook
export const useForm = <T extends Record<string, any>>(initialValues: T) => {
  const [values, setValues] = useState<T>(initialValues)
  const [errors, setErrors] = useState<Partial<Record<keyof T, string>>>({})
  const [touched, setTouched] = useState<Partial<Record<keyof T, boolean>>>({})

  const setValue = useCallback((name: keyof T, value: T[keyof T]) => {
    setValues(prev => ({ ...prev, [name]: value }))
    setTouched(prev => ({ ...prev, [name]: true }))
  }, [])

  const setError = useCallback((name: keyof T, error: string) => {
    setErrors(prev => ({ ...prev, [name]: error }))
  }, [])

  const clearError = useCallback((name: keyof T) => {
    setErrors(prev => {
      const newErrors = { ...prev }
      delete newErrors[name]
      return newErrors
    })
  }, [])

  const validate = useCallback((validator: (values: T) => Partial<Record<keyof T, string>>) => {
    const newErrors = validator(values)
    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }, [values])

  const reset = useCallback(() => {
    setValues(initialValues)
    setErrors({})
    setTouched({})
  }, [initialValues])

  const isFieldTouched = useCallback((name: keyof T) => {
    return touched[name] || false
  }, [touched])

  const hasFieldError = useCallback((name: keyof T) => {
    return Boolean(errors[name] && touched[name])
  }, [errors, touched])

  return useMemo(() => ({
    values,
    errors,
    touched,
    setValue,
    setError,
    clearError,
    validate,
    reset,
    isFieldTouched,
    hasFieldError,
    isValid: Object.keys(errors).length === 0,
    isDirty: Object.keys(touched).length > 0,
  }), [values, errors, touched, setValue, setError, clearError, validate, reset, isFieldTouched, hasFieldError])
}

// 无限滚动Hook
export const useInfiniteScroll = (
  fetchMore: (page: number) => Promise<void>,
  hasMore: boolean
) => {
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)

  const loadMore = useCallback(async () => {
    if (loading || !hasMore) return

    setLoading(true)
    try {
      await fetchMore(page + 1)
      setPage(prev => prev + 1)
    } finally {
      setLoading(false)
    }
  }, [fetchMore, page, loading, hasMore])

  useEffect(() => {
    const handleScroll = () => {
      if (
        window.innerHeight + document.documentElement.scrollTop
        >= document.documentElement.offsetHeight - 1000
      ) {
        loadMore()
      }
    }

    window.addEventListener('scroll', handleScroll)
    return () => window.removeEventListener('scroll', handleScroll)
  }, [loadMore])

  return {
    loading,
    loadMore,
    page
  }
}