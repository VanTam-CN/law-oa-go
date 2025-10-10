import { get, post } from './http';

// 诉讼费计算
export interface LitigationFeeParams {
  amount: number;
}

export interface LitigationFeeResult {
  amount: number;
  fee: number;
  calculationTime: number;
}

export function calculateLitigationFee(params: LitigationFeeParams) {
  return post('/tools/litigation-fee', params);
}

// 利息计算器
export interface InterestCalculatorParams {
  principal: number;
  rate: number;
  days: number;
  type: 'simple' | 'compound' | 'penalty';
}

export interface InterestCalculatorResult {
  principal: number;
  rate: number;
  days: number;
  type: string;
  interest: number;
  total: number;
}

export function calculateInterest(params: InterestCalculatorParams) {
  return post('/tools/interest-calculator', params);
}

// 工期计算器
export interface DeadlineCalculatorParams {
  startDate: string;
  days: number;
  excludeWeekends: boolean;
  excludeHolidays: boolean;
}

export interface DeadlineCalculatorResult {
  startDate: string;
  days: number;
  excludeWeekends: boolean;
  excludeHolidays: boolean;
  endDate: string;
  workDays: number;
}

export function calculateDeadline(params: DeadlineCalculatorParams) {
  return post('/tools/deadline-calculator', params);
}

// 法条查询
export interface LawItem {
  id: number;
  title: string;
  category: string;
  content: string;
  effectiveDate: string;
}

export function getLaws(category?: string) {
  return get('/tools/laws', { category });
}

// 合同模板
export interface ContractTemplate {
  id: number;
  name: string;
  category: string;
  description: string;
  downloadUrl: string;
  updateTime: string;
}

export function getContractTemplates() {
  return get('/tools/contract-templates');
}

// 文档转换
export interface DocumentConvertParams {
  sourceFormat: string;
  targetFormat: string;
}

export interface DocumentConvertResult {
  sourceFormat: string;
  targetFormat: string;
  status: string;
  message: string;
  downloadUrl: string;
}

export function convertDocument(params: DocumentConvertParams) {
  return post('/tools/document-convert', params);
}

// 翻译服务
export interface TranslateParams {
  text: string;
  targetLang: string;
}

export interface TranslateResult {
  originalText: string;
  targetLang: string;
  translatedText: string;
  sourceLang: string;
}

export function translate(params: TranslateParams) {
  return post('/tools/translate', params);
}