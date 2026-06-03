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

// Mock localStorage
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
  key: jest.fn(),
  length: 0,
} as unknown as Storage

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
})

// Mock sessionStorage
const sessionStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
  key: jest.fn(),
  length: 0,
} as unknown as Storage

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

const originalGetComputedStyle = window.getComputedStyle.bind(window)

const getComputedStyleMock: typeof window.getComputedStyle = (element, pseudoElt) => {
  const computedStyle = originalGetComputedStyle(element, pseudoElt)
  if (computedStyle && typeof computedStyle.getPropertyValue === 'function') {
    return computedStyle
  }

  return {
    getPropertyValue: jest.fn().mockReturnValue(''),
  } as unknown as CSSStyleDeclaration
}

Object.defineProperty(window, 'getComputedStyle', {
  writable: true,
  value: getComputedStyleMock,
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
