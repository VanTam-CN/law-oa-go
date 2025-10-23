/**
 * 现代化Mock索引文件 - 提供全面的Mock服务
 * 基于Jest 30.2和最新测试最佳实践
 */

// API Mocks
export { default as apiClientMock } from './api-client'
export { default as authServiceMock } from './auth-service'
export { default as caseServiceMock } from './case-service'

// React Router Mocks
export { default as routerMock } from './router'

// Ant Design Mocks
export { default as antdMock } from './antd'

// Utility Mocks
export { default as dateMock } from './date'
export { default as storageMock } from './storage'

// React Query Mocks
export { default as reactQueryMock } from './react-query'

// 全局Mock工厂
export * from './factory'