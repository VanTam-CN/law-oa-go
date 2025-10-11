import { get, post, put, del } from '../services/http';
import { caseService } from '../services/case';

export { caseService };

// 导出具体函数以保持向后兼容
export const getCaseList = caseService.getCaseList;
export const createCase = caseService.addCase;
export const updateCase = caseService.updateCase;
export const deleteCase = caseService.deleteCase;
export const assignLawyer = caseService.assignLawyer;
export const updateStatus = caseService.updateStatus;
export const getCaseStats = caseService.getCaseStats;
export const getCase = caseService.getCase;