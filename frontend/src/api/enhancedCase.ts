// 增强案例API - 统一导出接口
export * from '../services/enhancedCase';

// 为了保持向后兼容性，重新导出一些常用方法
export const enhancedCaseAPI = {
  createEnhancedCase: (data: any) => import('../services/enhancedCase').then(module => module.enhancedCaseService.createEnhancedCase(data)),
  getEnhancedCase: (id: number) => import('../services/enhancedCase').then(module => module.enhancedCaseService.getEnhancedCase(id)),
  updateEnhancedCase: (id: number, data: any) => import('../services/enhancedCase').then(module => module.enhancedCaseService.updateEnhancedCase(id, data)),
  listEnhancedCases: (params: any) => import('../services/enhancedCase').then(module => module.enhancedCaseService.listEnhancedCases(params)),
  deleteEnhancedCase: (id: number) => import('../services/enhancedCase').then(module => module.enhancedCaseService.deleteEnhancedCase(id)),
  performConflictCheck: (id: number) => import('../services/enhancedCase').then(module => module.enhancedCaseService.performConflictCheck(id)),
  addClientToCase: (id: number, data: any) => import('../services/enhancedCase').then(module => module.enhancedCaseService.addClientToCase(id, data)),
  removeClientFromCase: (id: number, clientId: string) => import('../services/enhancedCase').then(module => module.enhancedCaseService.removeClientFromCase(id, clientId))
};