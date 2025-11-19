/**
 * Jest测试框架配置 - 基于最新最佳实践2024
 * 支持TypeScript、React Testing Library和现代化测试工具
 */

import type { Config } from '@jest/types'

const config: Config.InitialOptions = {
  // 测试环境
  testEnvironment: 'jsdom',

  // 测试环境选项
  testEnvironmentOptions: {
    html: '<html lang="zh-CN"></html>',
    url: 'http://localhost:3000',
    userAgent: 'Jest Test Environment',
  },

  // 测试匹配模式
  testMatch: ['**/__tests__/**/*.(ts|tsx|js|jsx)', '**/*.(test|spec).(ts|tsx|js|jsx)'],

  // 文件转换配置
  preset: 'ts-jest',
  testPathIgnorePatterns: ['/node_modules/', '/dist/', '/build/'],

  // 模块名称映射
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/src/$1',
    '^@components/(.*)$': '<rootDir>/src/components/$1',
    '^@pages/(.*)$': '<rootDir>/src/pages/$1',
    '^@hooks/(.*)$': '<rootDir>/src/hooks/$1',
    '^@utils/(.*)$': '<rootDir>/src/utils/$1',
    '^@services/(.*)$': '<rootDir>/src/services/$1',
    '^@stores/(.*)$': '<rootDir>/src/stores/$1',
    '^@types/(.*)$': '<rootDir>/src/types/$1',
    '^@assets/(.*)$': '<rootDir>/src/assets/$1',
    '\\.(css|less|scss|sass)$': 'identity-obj-proxy',
  },

  // 设置文件 - 在测试框架安装之后、测试代码执行之前运行
  setupFilesAfterEnv: ['<rootDir>/src/test/setup.ts', '<rootDir>/src/test/mocks/index.ts'],

  // 覆盖率配置
  collectCoverage: true,
  collectCoverageFrom: [
    'src/**/*.{ts,tsx}',
    '!src/**/*.d.ts',
    '!src/**/*.stories.{ts,tsx}',
    '!src/test/**',
    '!src/**/__tests__/**',
    '!src/**/index.ts',
    '!src/main.tsx',
    '!src/vite-env.d.ts',
    '!src/assets/**',
  ],

  // 覆盖率报告格式
  coverageReporters: ['text', 'text-summary', 'html', 'lcov', 'json', 'clover'],

  // 覆盖率阈值
  coverageThreshold: {
    global: {
      branches: 70,
      functions: 70,
      lines: 70,
      statements: 70,
    },
    // 特定模块的更高覆盖率要求
    'src/components/**': {
      branches: 80,
      functions: 80,
      lines: 80,
      statements: 80,
    },
    'src/hooks/**': {
      branches: 80,
      functions: 80,
      lines: 80,
      statements: 80,
    },
    'src/services/**': {
      branches: 85,
      functions: 85,
      lines: 85,
      statements: 85,
    },
    'src/stores/**': {
      branches: 80,
      functions: 80,
      lines: 80,
      statements: 80,
    },
  },

  // 清理配置
  clearMocks: true,
  restoreMocks: true,

  // 错误处理
  errorOnDeprecated: true,

  // 详细输出
  verbose: true,

  // 并行执行
  maxWorkers: '50%',

  // 显示名称（用于多项目）
  displayName: {
    name: 'Law OA Frontend',
    color: 'blue',
  },

  // 全局变量
  globals: {
    'ts-jest': {
      tsconfig: 'tsconfig.json',
    },
  },

  // 模块文件扩展
  moduleFileExtensions: ['ts', 'tsx', 'js', 'jsx', 'json', 'node'],

  // 测试超时
  testTimeout: 10000,

  // 快照配置
  snapshotSerializers: [],

  // 监视模式配置
  watchPlugins: ['jest-watch-typeahead/filename', 'jest-watch-typeahead/testname'],

  // 转换忽略模式
  transformIgnorePatterns: ['node_modules/(?!(antd|@ant-design|rc-.*|@babel/runtime))'],

  // 项目配置（用于多包项目）
  projects: [
    {
      displayName: 'unit',
      testMatch: ['<rootDir>/src/**/__tests__/**/*.unit.{test,spec}.{ts,tsx}'],
      setupFilesAfterEnv: ['<rootDir>/src/test/setup.ts'],
    },
    {
      displayName: 'integration',
      testMatch: ['<rootDir>/src/**/__tests__/**/*.integration.{test,spec}.{ts,tsx}'],
      setupFilesAfterEnv: ['<rootDir>/src/test/setup.ts'],
    },
    {
      displayName: 'e2e',
      testMatch: ['<rootDir>/src/**/__tests__/**/*.e2e.{test,spec}.{ts,tsx}'],
      setupFilesAfterEnv: ['<rootDir>/src/test/setup.ts'],
      testTimeout: 30000,
    },
  ],
}

export default config
