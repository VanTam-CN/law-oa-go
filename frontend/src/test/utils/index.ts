/**
 * 测试工具集 - 现代化测试辅助函数和工具
 * 基于Jest 30.2和React Testing Library最佳实践
 */

import { RenderOptions, RenderResult } from '@testing-library/react'
import { ReactElement, ReactNode } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router'
import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'

// 自定义渲染函数
interface CustomRenderOptions extends Omit<RenderOptions, 'wrapper'> {
  queryClient?: QueryClient
  router?: {
    initialEntries?: string[]
    initialIndex?: number
  }
  antd?: boolean
}

// 默认QueryClient
const createDefaultQueryClient = () => new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
      gcTime: 0,
      staleTime: 0,
      refetchOnWindowFocus: false,
      refetchOnReconnect: false
    },
    mutations: {
      retry: false
    }
  }
})

// 测试Wrapper组件
interface TestWrapperProps {
  children: ReactNode
  queryClient?: QueryClient
  router?: {
    initialEntries?: string[]
    initialIndex?: number
  }
  antd?: boolean
}

const TestWrapper: React.FC<TestWrapperProps> = ({
  children,
  queryClient = createDefaultQueryClient(),
  router,
  antd = true
}) => {
  const content = (
    <QueryClientProvider client={queryClient}>
      {router ? (
        <MemoryRouter initialEntries={router.initialEntries} initialIndex={router.initialIndex}>
          {children}
        </MemoryRouter>
      ) : (
        children
      )}
    </QueryClientProvider>
  );

  if (antd) {
    return (
      <ConfigProvider locale={zhCN}>
        {content}
      </ConfigProvider>
    )
  }

  return content
}

// 自定义渲染函数
export const customRender = (
  ui: ReactElement,
  options: CustomRenderOptions = {}
): RenderResult => {
  const {
    queryClient,
    router,
    antd = true,
    ...renderOptions
  } = options

  return render(
    <TestWrapper
      queryClient={queryClient}
      router={router}
      antd={antd}
    >
      {ui}
    </TestWrapper>,
    renderOptions
  )
}

// 重新导出testing-library的所有内容
export * from '@testing-library/react'
export { customRender as render }

// ==============================
// 断言辅助函数
// ==============================

/**
 * 检查元素是否存在并可见
 */
export const expectElementVisible = (element: HTMLElement) => {
  expect(element).toBeInTheDocument()
  expect(element).toBeVisible()
}

/**
 * 检查元素是否被禁用
 */
export const expectElementDisabled = (element: HTMLElement) => {
  expect(element).toBeInTheDocument()
  expect(element).toBeDisabled()
}

/**
 * 检查元素是否被启用
 */
export const expectElementEnabled = (element: HTMLElement) => {
  expect(element).toBeInTheDocument()
  expect(element).not.toBeDisabled()
}

/**
 * 检查元素是否有特定的CSS类
 */
export const expectElementHasClass = (element: HTMLElement, className: string) => {
  expect(element).toBeInTheDocument()
  expect(element).toHaveClass(className)
}

/**
 * 检查元素是否有特定的属性
 */
export const expectElementHasAttribute = (
  element: HTMLElement,
  attribute: string,
  value?: string
) => {
  expect(element).toBeInTheDocument()
  if (value !== undefined) {
    expect(element).toHaveAttribute(attribute, value)
  } else {
    expect(element).toHaveAttribute(attribute)
  }
}

/**
 * 检查表单字段是否有效
 */
export const expectFieldValid = (field: HTMLElement) => {
  expect(field).toBeInTheDocument()
  expect(field).toBeValid()
}

/**
 * 检查表单字段是否无效
 */
export const expectFieldInvalid = (field: HTMLElement) => {
  expect(field).toBeInTheDocument()
  expect(field).toBeInvalid()
}

// ==============================
// 异步辅助函数
// ==============================

/**
 * 等待元素出现
 */
export const waitForElement = async (
  getElement: () => HTMLElement | null,
  timeout = 5000
): Promise<HTMLElement> => {
  return new Promise((resolve, reject) => {
    const startTime = Date.now()

    const checkElement = () => {
      const element = getElement()
      if (element) {
        resolve(element)
      } else if (Date.now() - startTime > timeout) {
        reject(new Error(`Element not found within ${timeout}ms`))
      } else {
        setTimeout(checkElement, 100)
      }
    }

    checkElement()
  })
}

/**
 * 等待元素消失
 */
export const waitForElementToBeRemoved = async (
  getElement: () => HTMLElement | null,
  timeout = 5000
): Promise<void> => {
  return new Promise((resolve, reject) => {
    const startTime = Date.now()

    const checkElement = () => {
      const element = getElement()
      if (!element) {
        resolve()
      } else if (Date.now() - startTime > timeout) {
        reject(new Error(`Element still present after ${timeout}ms`))
      } else {
        setTimeout(checkElement, 100)
      }
    }

    checkElement()
  })
}

/**
 * 等待加载状态完成
 */
export const waitForLoadingToFinish = async (
  timeout = 10000
): Promise<void> => {
  const startTime = Date.now()

  while (Date.now() - startTime < timeout) {
    const loadingElements = document.querySelectorAll('[aria-busy="true"], .loading, .ant-spin')
    if (loadingElements.length === 0) {
      return
    }
    await new Promise(resolve => setTimeout(resolve, 100))
  }

  throw new Error(`Loading did not finish within ${timeout}ms`)
}

// ==============================
// 表单辅助函数
// ==============================

/**
 * 填写表单字段
 */
export const fillFormField = async (
  field: HTMLElement,
  value: string
): Promise<void> => {
  const user = (await import('@testing-library/user-event')).default.setup()
  await user.clear(field)
  await user.type(field, value)
}

/**
 * 填写多个表单字段
 */
export const fillFormFields = async (
  fields: Record<string, { element: HTMLElement; value: string }>
): Promise<void> => {
  const user = (await import('@testing-library/user-event')).default.setup()

  for (const [name, { element, value }] of Object.entries(fields)) {
    await user.clear(element)
    await user.type(element, value)
  }
}

/**
 * 提交表单
 */
export const submitForm = async (
  form: HTMLElement,
  submitButton?: HTMLElement
): Promise<void> => {
  const user = (await import('@testing-library/user-event')).default.setup()

  if (submitButton) {
    await user.click(submitButton)
  } else {
    const button = form.querySelector('button[type="submit"], input[type="submit"]') as HTMLElement
    if (button) {
      await user.click(button)
    } else {
      await user.type(form, '{enter}')
    }
  }
}

// ==============================
// Mock辅助函数
// ==============================

/**
 * 创建fetch mock响应
 */
export const createMockFetchResponse = <T>(
  data: T,
  status = 200,
  headers: Record<string, string> = {}
): Response => {
  return new Response(JSON.stringify(data), {
    status,
    headers: {
      'Content-Type': 'application/json',
      ...headers
    }
  })
}

/**
 * 创建fetch mock错误
 */
export const createMockFetchError = (
  message: string,
  status = 500
): Response => {
  return new Response(JSON.stringify({ error: message }), {
    status,
    headers: { 'Content-Type': 'application/json' }
  })
}

/**
 * Mock fetch函数
 */
export const mockFetch = (responses: Array<{
  url: string | RegExp
  response: Response | Error
  delay?: number
}>): void => {
  const mockFn = jest.fn()

  responses.forEach(({ url, response, delay }) => {
    if (response instanceof Error) {
      mockFn.mockImplementationOnce(
        () => delay
          ? new Promise((_, reject) => setTimeout(() => reject(response), delay))
          : Promise.reject(response)
      )
    } else {
      mockFn.mockImplementationOnce(
        () => delay
          ? new Promise(resolve => setTimeout(() => resolve(response), delay))
          : Promise.resolve(response)
      )
    }
  })

  global.fetch = mockFn
}

// ==============================
// 时间辅助函数
// ==============================

/**
 * Mock定时器
 */
export const setupMockTimers = (): void => {
  jest.useFakeTimers()
}

/**
 * 清理Mock定时器
 */
export const cleanupMockTimers = (): void => {
  jest.runOnlyPendingTimers()
  jest.useRealTimers()
}

/**
 * 快进时间
 */
export const advanceTimersByTime = (ms: number): void => {
  jest.advanceTimersByTime(ms)
}

/**
 * 快进到下一个定时器
 */
export const advanceTimersToNextTimer = (): void => {
  jest.advanceTimersToNextTimer()
}

// ==============================
// 存储辅助函数
// ==============================

/**
 * Mock localStorage
 */
export const mockLocalStorage = (): Storage => {
  const store: Record<string, string> = {}

  return {
    length: 0,
    clear(): void {
      Object.keys(store).forEach(key => delete store[key])
    },
    getItem(key: string): string | null {
      return store[key] || null
    },
    key(index: number): string | null {
      return Object.keys(store)[index] || null
    },
    removeItem(key: string): void {
      delete store[key]
    },
    setItem(key: string, value: string): void {
      store[key] = value
    }
  }
}

/**
 * Mock sessionStorage
 */
export const mockSessionStorage = (): Storage => {
  const store: Record<string, string> = {}

  return {
    length: 0,
    clear(): void {
      Object.keys(store).forEach(key => delete store[key])
    },
    getItem(key: string): string | null {
      return store[key] || null
    },
    key(index: number): string | null {
      return Object.keys(store)[index] || null
    },
    removeItem(key: string): void {
      delete store[key]
    },
    setItem(key: string, value: string): void {
      store[key] = value
    }
  }
}

// ==============================
// 网络辅助函数
// ==============================

/**
 * Mock网络状态
 */
export const mockNetworkStatus = (online: boolean): void => {
  Object.defineProperty(navigator, 'onLine', {
    writable: true,
    value: online
  })

  const event = new Event(online ? 'online' : 'offline')
  window.dispatchEvent(event)
}

/**
 * Mock地理位置
 */
export const mockGeolocation = (position?: GeolocationPosition): void => {
  const mockGeolocation = {
    getCurrentPosition: jest.fn(),
    watchPosition: jest.fn(),
    clearWatch: jest.fn()
  }

  if (position) {
    mockGeolocation.getCurrentPosition.mockImplementation(
      (success) => success(position)
    )
  } else {
    mockGeolocation.getCurrentPosition.mockImplementation(
      (_, error) => error(new Error('Geolocation not available'))
    )
  }

  Object.defineProperty(navigator, 'geolocation', {
    writable: true,
    value: mockGeolocation
  })
}

// ==============================
// 测试数据生成器
// ==============================

/**
 * 生成随机UUID
 */
export const generateUUID = (): string => {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = Math.random() * 16 | 0
    const v = c === 'x' ? r : (r & 0x3 | 0x8)
    return v.toString(16)
  })
}

/**
 * 生成随机字符串
 */
export const generateRandomString = (length = 10): string => {
  return Math.random().toString(36).substring(2, 2 + length)
}

/**
 * 生成随机数字
 */
export const generateRandomNumber = (min = 0, max = 100): number => {
  return Math.floor(Math.random() * (max - min + 1)) + min
}

/**
 * 生成随机邮箱
 */
export const generateRandomEmail = (): string => {
  return `${generateRandomString(8)}@example.com`
}

// ==============================
// 导出所有工具
// ==============================

export default {
  render: customRender,
  // 断言
  expectElementVisible,
  expectElementDisabled,
  expectElementEnabled,
  expectElementHasClass,
  expectElementHasAttribute,
  expectFieldValid,
  expectFieldInvalid,
  // 异步
  waitForElement,
  waitForElementToBeRemoved,
  waitForLoadingToFinish,
  // 表单
  fillFormField,
  fillFormFields,
  submitForm,
  // Mock
  createMockFetchResponse,
  createMockFetchError,
  mockFetch,
  // 时间
  setupMockTimers,
  cleanupMockTimers,
  advanceTimersByTime,
  advanceTimersToNextTimer,
  // 存储
  mockLocalStorage,
  mockSessionStorage,
  // 网络
  mockNetworkStatus,
  mockGeolocation,
  // 数据生成
  generateUUID,
  generateRandomString,
  generateRandomNumber,
  generateRandomEmail
}