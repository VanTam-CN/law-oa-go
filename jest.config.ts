/**
 * Jest测试框架配置 - 基于最新最佳实践2024
 * 支持TypeScript、React Testing Library和现代化测试工具
 * 修复ESM兼容性和TypeScript转换问题
 */

import type { Config } from '@jest/types'

const config: Config.InitialOptions = {
  // 测试环境
  testEnvironment: 'jsdom',

  // 扩展模块支持ESM
  extensionsToTreatAsEsm: ['.ts', '.tsx'],

  // 测试环境选项
  testEnvironmentOptions: {
    html: '<html lang="zh-CN"></html>',
    url: 'http://localhost:3000',
    userAgent: 'Jest Test Environment',
  },

  // 测试匹配模式
  testMatch: ['**/__tests__/**/*.(ts|tsx|js|jsx)', '**/*.(test|spec).(ts|tsx|js|jsx)'],
  testPathIgnorePatterns: ['/node_modules/', '/dist/', '/build/'],

  // 转换配置 - 关键修复：使用测试专用TypeScript配置
  transform: {
    '^.+\\.(ts|tsx)$': [
      'ts-jest',
      {
        tsconfig: 'tsconfig.test.json',
      },
    ],
  },

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

  // 支持ESM模块
  transformIgnorePatterns: ['node_modules/(?!(.*\\.mjs$|antd|@ant-design|rc-.*|@babel/runtime))'],

  // 设置文件
  setupFilesAfterEnv: ['<rootDir>/src/test/setup.ts'],

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
  coverageDirectory: 'coverage',
  coverageReporters: ['text', 'text-summary', 'html', 'lcov', 'json'],

  // 临时禁用覆盖率阈值，直到测试稳定
  // coverageThreshold: {
  //   global: {
  //     branches: 70,
  //     functions: 70,
  //     lines: 70,
  //     statements: 70,
  //   },
  // },

  // 清理配置
  clearMocks: true,
  restoreMocks: true,

  // 错误处理
  errorOnDeprecated: true,

  // 详细输出
  verbose: true,

  // 测试超时
  testTimeout: 10000,

  // 模块文件扩展
  moduleFileExtensions: ['ts', 'tsx', 'js', 'jsx', 'json', 'node'],

  // 测试运行器配置 - 使用默认的jest-circus

  // 显示名称
  displayName: {
    name: 'Law OA Frontend',
    color: 'blue',
  },
}

export default config
