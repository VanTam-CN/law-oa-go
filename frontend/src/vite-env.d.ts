/// <reference types="vite/client" />
/// <reference types="@testing-library/jest-dom/jest-globals" />

declare global {
  const describe: typeof import('@jest/globals').describe
  const it: typeof import('@jest/globals').it
  const test: typeof import('@jest/globals').test
  const expect: typeof import('@jest/globals').expect
  const beforeEach: typeof import('@jest/globals').beforeEach
  const afterEach: typeof import('@jest/globals').afterEach
  const beforeAll: typeof import('@jest/globals').beforeAll
  const afterAll: typeof import('@jest/globals').afterAll
  const jest: typeof import('@jest/globals').jest

  interface Performance {
    memory?: {
      usedJSHeapSize: number
      totalJSHeapSize: number
      jsHeapSizeLimit: number
    }
  }
}

export {}
