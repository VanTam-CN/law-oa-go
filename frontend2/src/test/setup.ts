import '@testing-library/jest-dom'
import { vi } from 'vitest'

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(), // deprecated
    removeListener: vi.fn(), // deprecated
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
})

// Mock ResizeObserver
global.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}))

// Mock IntersectionObserver
global.IntersectionObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}))

// Mock localStorage with actual storage behavior
const localStorageStore = new Map<string, string>()
const localStorageMock = {
  getItem: vi.fn((key: string) => localStorageStore.get(key) || null),
  setItem: vi.fn((key: string, value: string) => localStorageStore.set(key, value)),
  removeItem: vi.fn((key: string) => localStorageStore.delete(key)),
  clear: vi.fn(() => localStorageStore.clear()),
}
global.localStorage = localStorageMock

// Mock sessionStorage with actual storage behavior
const sessionStorageStore = new Map<string, string>()
const sessionStorageMock = {
  getItem: vi.fn((key: string) => sessionStorageStore.get(key) || null),
  setItem: vi.fn((key: string, value: string) => sessionStorageStore.set(key, value)),
  removeItem: vi.fn((key: string) => sessionStorageStore.delete(key)),
  clear: vi.fn(() => sessionStorageStore.clear()),
}
global.sessionStorage = sessionStorageMock

// Mock fetch
global.fetch = vi.fn()

// Mock scrollTo
window.scrollTo = vi.fn()

// Mock alert
window.alert = vi.fn()

// Mock confirm
window.confirm = vi.fn()

// Mock prompt
window.prompt = vi.fn()

// 设置测试环境变量
process.env.NODE_ENV = 'test'

// 清理函数
afterEach(() => {
  vi.clearAllMocks()
  localStorageStore.clear()
  sessionStorageStore.clear()
})