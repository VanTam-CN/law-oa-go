/**
 * Jest测试设置文件 - 简化兼容版
 * 基于ESM兼容性优先考虑
 */

import '@testing-library/jest-dom'
import { TextDecoder, TextEncoder } from 'util'

Object.assign(globalThis, {
  TextDecoder,
  TextEncoder,
})

function createStorageMock(): Storage {
  let values = new Map<string, string>()

  return {
    get length() {
      return values.size
    },
    clear: jest.fn(() => {
      values = new Map<string, string>()
    }),
    getItem: jest.fn((key: string) => values.get(key) ?? null),
    key: jest.fn((index: number) => Array.from(values.keys())[index] ?? null),
    removeItem: jest.fn((key: string) => {
      values.delete(key)
    }),
    setItem: jest.fn((key: string, value: string) => {
      values.set(key, String(value))
    }),
  }
}

// Use independent in-memory stores so persistence behavior remains testable.
const localStorageMock = createStorageMock()

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
})

const sessionStorageMock = createStorageMock()

Object.defineProperty(window, 'sessionStorage', {
  value: sessionStorageMock,
})

// Mock matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: jest.fn().mockImplementation((query) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(),
    removeListener: jest.fn(),
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
  })),
})

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}

// 设置测试超时
jest.setTimeout(10000)

// 清理函数
afterEach(() => {
  jest.clearAllMocks()
  document.body.innerHTML = ''
})
