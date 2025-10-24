/**
 * 案件管理测试验证工具
 */

import { CaseData, Document } from '../../types/test-data-types';
import { CASE_TEST_CONFIG, VALIDATION_RULES, CaseTestUtils } from './case-test-config';
import { Logger } from '../../core/logger';

export interface ValidationResult {
  isValid: boolean;
  errors: string[];
  warnings: string[];
  details: {
    field: string;
    value: any;
    validation: 'passed' | 'failed' | 'warning';
    message: string;
  }[];
}

export interface TestCaseValidationResult {
  caseData: CaseData;
  validation: ValidationResult;
  suggestions: string[];
}

export interface DocumentValidationResult {
  document: Document;
  validation: ValidationResult;
  suggestions: string[];
}

export class CaseValidationUtils {
  private logger: Logger;

  constructor(logger?: Logger) {
    this.logger = logger || new Logger('CaseValidationUtils');
  }

  /**
   * 验证案件数据
   */
  validateCaseData(caseData: CaseData): ValidationResult {
    const errors: string[] | undefined = undefined;
    const warnings: string[] | undefined = undefined;
    const details: ValidationResult['details'] = [];

    // 验证必填字段
    const requiredFields = Object.keys(VALIDATION_RULES.required);
    for (const field of requiredFields) {
      const value = caseData[field as keyof CaseData];
      const rule = VALIDATION_RULES.required[field as keyof typeof VALIDATION_RULES.required];

      if (!value || (typeof value === 'string' && value.trim() === '')) {
        errors.push(rule.message);
        details.push({
          field,
          value,
          validation: 'failed',
          message: rule.message
        });
      } else {
        details.push({
          field,
          value,
          validation: 'passed',
          message: `${field} 验证通过`
        });
      }
    }

    // 验证格式
    this.validateFormats(caseData, errors, warnings, details);

    // 验证数值范围
    this.validateRanges(caseData, errors, warnings, details);

    // 验证日期逻辑
    this.validateDates(caseData, errors, warnings, details);

    // 验证业务规则
    this.validateBusinessRules(caseData, errors, warnings, details);

    return {
      isValid: errors.length === 0,
      errors,
      warnings,
      details
    };
  }

  /**
   * 验证格式
   */
  private validateFormats(
    caseData: CaseData,
    errors: string[],
    warnings: string[],
    details: ValidationResult['details']
  ): void {
    // 验证邮箱格式（如果有）
    if (caseData.clientEmail) {
      const emailRule = VALIDATION_RULES.format.email;
      if (!emailRule.pattern.test(caseData.clientEmail)) {
        errors.push(emailRule.message);
        details.push({
          field: 'clientEmail',
          value: caseData.clientEmail,
          validation: 'failed',
          message: emailRule.message
        });
      }
    }

    // 验证电话格式（如果有）
    if (caseData.clientPhone) {
      const phoneRule = VALIDATION_RULES.format.phone;
      if (!phoneRule.pattern.test(caseData.clientPhone)) {
        warnings.push(phoneRule.message);
        details.push({
          field: 'clientPhone',
          value: caseData.clientPhone,
          validation: 'warning',
          message: phoneRule.message
        });
      }
    }

    // 验证案件编号格式
    const caseNumberRule = VALIDATION_RULES.format.caseNumber;
    if (!caseNumberRule.pattern.test(caseData.caseNumber)) {
      errors.push(caseNumberRule.message);
      details.push({
        field: 'caseNumber',
        value: caseData.caseNumber,
        validation: 'failed',
        message: caseNumberRule.message
      });
    }
  }

  /**
   * 验证数值范围
   */
  private validateRanges(
    caseData: CaseData,
    errors: string[],
    warnings: string[],
    details: ValidationResult['details']
  ): void {
    const numericFields = ['estimatedValue', 'budget', 'actualCost'] as const;

    for (const field of numericFields) {
      const value = caseData[field];
      const rule = VALIDATION_RULES.range[field];

      if (value < rule.min) {
        errors.push(rule.message);
        details.push({
          field,
          value,
          validation: 'failed',
          message: rule.message
        });
      } else {
        details.push({
          field,
          value,
          validation: 'passed',
          message: `${field} 范围验证通过`
        });
      }
    }
  }

  /**
   * 验证日期逻辑
   */
  private validateDates(
    caseData: CaseData,
    errors: string[],
    warnings: string[],
    details: ValidationResult['details']
  ): void {
    const dateRule = VALIDATION_RULES.date.startDate;

    if (caseData.startDate && caseData.expectedEndDate) {
      const start = new Date(caseData.startDate);
      const end = new Date(caseData.expectedEndDate);

      if (start >= end) {
        errors.push(dateRule.message);
        details.push({
          field: 'startDate',
          value: caseData.startDate,
          validation: 'failed',
          message: dateRule.message
        });
      } else {
        details.push({
          field: 'startDate',
          value: caseData.startDate,
          validation: 'passed',
          message: '日期逻辑验证通过'
        });
      }
    }

    // 验证日期格式
    const dateFields = ['startDate', 'expectedEndDate'] as const;
    for (const field of dateFields) {
      const value = caseData[field];
      if (value && isNaN(Date.parse(value))) {
        errors.push(`${field} 日期格式无效`);
        details.push({
          field,
          value,
          validation: 'failed',
          message: `${field} 日期格式无效`
        });
      }
    }
  }

  /**
   * 验证业务规则
   */
  private validateBusinessRules(
    caseData: CaseData,
    errors: string[],
    warnings: string[],
    details: ValidationResult['details']
  ): void {
    // 验证预算与实际成本的关系
    if (caseData.budget && caseData.actualCost) {
      if (caseData.actualCost > caseData.budget * 1.5) {
        warnings.push('实际成本超过预算50%，请关注');
        details.push({
          field: 'actualCost',
          value: caseData.actualCost,
          validation: 'warning',
          message: '实际成本超过预算50%'
        });
      }
    }

    // 验证预估价值的合理性
    if (caseData.estimatedValue && caseData.estimatedValue < 1000) {
      warnings.push('预估价值较低，请确认是否正确');
      details.push({
        field: 'estimatedValue',
        value: caseData.estimatedValue,
        validation: 'warning',
        message: '预估价值较低'
      });
    }

    // 验证案件状态的合理性
    if (caseData.status === 'closed' && caseData.expectedEndDate) {
      const expectedEnd = new Date(caseData.expectedEndDate);
      const now = new Date();
      if (now < expectedEnd) {
        warnings.push('案件已结案但未到预计结束日期');
        details.push({
          field: 'status',
          value: caseData.status,
          validation: 'warning',
          message: '案件提前结案'
        });
      }
    }

    // 验证团队成员配置
    if (caseData.teamMembers && caseData.teamMembers.length > 10) {
      warnings.push('团队成员较多，建议考虑管理效率');
      details.push({
        field: 'teamMembers',
        value: caseData.teamMembers,
        validation: 'warning',
        message: '团队成员数量较多'
      });
    }
  }

  /**
   * 验证文档数据
   */
  validateDocumentData(document: Document): ValidationResult {
    const errors: string[] | undefined = undefined;
    const warnings: string[] | undefined = undefined;
    const details: ValidationResult['details'] = [];

    // 验证必填字段
    const requiredFields = ['name', 'fileName', 'fileType', 'caseId', 'documentType', 'category'];
    for (const field of requiredFields) {
      const value = document[field as keyof Document];
      if (!value || (typeof value === 'string' && value.trim() === '')) {
        errors.push(`${field} 不能为空`);
        details.push({
          field,
          value,
          validation: 'failed',
          message: `${field} 不能为空`
        });
      } else {
        details.push({
          field,
          value,
          validation: 'passed',
          message: `${field} 验证通过`
        });
      }
    }

    // 验证文件大小
    if (document.fileSize <= 0) {
      errors.push('文件大小必须大于0');
      details.push({
        field: 'fileSize',
        value: document.fileSize,
        validation: 'failed',
        message: '文件大小必须大于0'
      });
    } else if (document.fileSize > 100 * 1024 * 1024) { // 100MB
      warnings.push('文件大小超过100MB，可能影响上传性能');
      details.push({
        field: 'fileSize',
        value: document.fileSize,
        validation: 'warning',
        message: '文件较大'
      });
    }

    // 验证版本号
    if (document.version <= 0) {
      errors.push('版本号必须大于0');
      details.push({
        field: 'version',
        value: document.version,
        validation: 'failed',
        message: '版本号必须大于0'
      });
    }

    // 验证内容类型
    const allowedContentTypes = [
      'application/pdf',
      'application/msword',
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      'image/jpeg',
      'image/png',
      'text/plain'
    ];

    if (!allowedContentTypes.includes(document.contentType)) {
      warnings.push('文件类型不被推荐，请确认是否支持');
      details.push({
        field: 'contentType',
        value: document.contentType,
        validation: 'warning',
        message: '文件类型不常见'
      });
    }

    return {
      isValid: errors.length === 0,
      errors,
      warnings,
      details
    };
  }

  /**
   * 验证案件与文档的关联性
   */
  validateCaseDocumentAssociation(
    caseData: CaseData,
    documents: Document[]
  ): ValidationResult {
    const errors: string[] | undefined = undefined;
    const warnings: string[] | undefined = undefined;
    const details: ValidationResult['details'] = [];

    // 验证文档是否属于当前案件
    const caseDocuments = documents.filter(doc => doc.caseId === caseData.id);
    const otherCaseDocuments = documents.filter(doc => doc.caseId !== caseData.id);

    if (otherCaseDocuments.length > 0) {
      errors.push(`发现 ${otherCaseDocuments.length} 个不属于此案件的文档`);
      otherCaseDocuments.forEach(doc => {
        details.push({
          field: 'caseId',
          value: doc.caseId,
          validation: 'failed',
          message: `文档 ${doc.name} 不属于案件 ${caseData.id}`
        });
      });
    }

    // 验证必要文档类型
    const requiredDocumentTypes = ['contract', 'evidence'];
    const existingTypes = caseDocuments.map(doc => doc.documentType);
    const missingTypes = requiredDocumentTypes.filter(type => !existingTypes.includes(type));

    if (missingTypes.length > 0) {
      warnings.push(`缺少必要文档类型: ${missingTypes.join(', ')}`);
      missingTypes.forEach(type => {
        details.push({
          field: 'documentType',
          value: type,
          validation: 'warning',
          message: `缺少 ${type} 类型文档`
        });
      });
    }

    // 验证文档大小总和
    const totalSize = caseDocuments.reduce((sum, doc) => sum + doc.fileSize, 0);
    if (totalSize > 50 * 1024 * 1024) { // 50MB
      warnings.push('案件文档总大小超过50MB，可能影响系统性能');
      details.push({
        field: 'totalSize',
        value: totalSize,
        validation: 'warning',
        message: '文档总大小较大'
      });
    }

    return {
      isValid: errors.length === 0,
      errors,
      warnings,
      details
    };
  }

  /**
   * 验证案件搜索功能
   */
  validateSearchFunctionality(
    searchQuery: string,
    results: CaseData[],
    allCases: CaseData[]
  ): ValidationResult {
    const errors: string[] | undefined = undefined;
    const warnings: string[] | undefined = undefined;
    const details: ValidationResult['details'] = [];

    if (!searchQuery || searchQuery.trim() === '') {
      errors.push('搜索查询不能为空');
      details.push({
        field: 'searchQuery',
        value: searchQuery,
        validation: 'failed',
        message: '搜索查询不能为空'
      });
    }

    if (!Array.isArray(results)) {
      errors.push('搜索结果必须是数组');
      details.push({
        field: 'results',
        value: results,
        validation: 'failed',
        message: '搜索结果格式错误'
      });
    }

    // 验证搜索结果的相关性
    if (Array.isArray(results) && results.length > 0) {
      const queryLower = searchQuery.toLowerCase();
      const relevantResults = results.filter(caseData => {
        return Object.values(caseData).some(value =>
          typeof value === 'string' && value.toLowerCase().includes(queryLower)
        );
      });

      if (relevantResults.length !== results.length) {
        warnings.push('部分搜索结果可能不相关');
        details.push({
          field: 'relevance',
          value: `${relevantResults.length}/${results.length}`,
          validation: 'warning',
          message: '搜索结果相关性待提高'
        });
      }
    }

    return {
      isValid: errors.length === 0,
      errors,
      warnings,
      details
    };
  }

  /**
   * 生成测试建议
   */
  generateSuggestions(caseData: CaseData): string[] {
    const suggestions: string[] | undefined = undefined;

    // 基于验证结果生成建议
    if (!caseData.clientEmail && !caseData.clientPhone) {
      suggestions.push('建议添加客户联系方式');
    }

    if (!caseData.description || caseData.description.length < 50) {
      suggestions.push('建议补充更详细的案件描述');
    }

    if (!caseData.tags || caseData.tags.length === 0) {
      suggestions.push('建议添加标签以便分类和搜索');
    }

    if (!caseData.milestones || caseData.milestones.length === 0) {
      suggestions.push('建议设置案件里程碑以便跟踪进度');
    }

    if (!caseData.budget) {
      suggestions.push('建议设置预算以便控制成本');
    }

    if (caseData.priority === 'high' && !caseData.assignedAttorney) {
      suggestions.push('高优先级案件建议尽快分配律师');
    }

    if (caseData.status === 'active' && caseData.teamMembers.length === 0) {
      suggestions.push('进行中的案件建议分配团队成员');
    }

    return suggestions;
  }

  /**
   * 验证测试数据的完整性
   */
  validateTestDataIntegrity(): ValidationResult {
    const errors: string[] | undefined = undefined;
    const warnings: string[] | undefined = undefined;
    const details: ValidationResult['details'] = [];

    // 验证测试用户
    if (!CASE_TEST_CONFIG.validUser) {
      errors.push('缺少有效的测试用户');
    }

    // 验证测试案件数据
    if (!CASE_TEST_CONFIG.testCases || CASE_TEST_CONFIG.testCases.length === 0) {
      errors.push('缺少测试案件数据');
    } else {
      CASE_TEST_CONFIG.testCases.forEach((testCase, index) => {
        const validation = this.validateCaseData(testCase);
        if (!validation.isValid) {
          errors.push(`测试案件 ${index + 1} 验证失败: ${validation.errors.join(', ')}`);
        }
      });
    }

    // 验证测试文档数据
    if (!CASE_TEST_CONFIG.testDocuments || CASE_TEST_CONFIG.testDocuments.length === 0) {
      warnings.push('缺少测试文档数据');
    } else {
      CASE_TEST_CONFIG.testDocuments.forEach((document, index) => {
        const validation = this.validateDocumentData(document);
        if (!validation.isValid) {
          errors.push(`测试文档 ${index + 1} 验证失败: ${validation.errors.join(', ')}`);
        }
      });
    }

    return {
      isValid: errors.length === 0,
      errors,
      warnings,
      details
    };
  }

  /**
   * 生成验证报告
   */
  generateValidationReport(
    caseValidations: TestCaseValidationResult[],
    documentValidations: DocumentValidationResult[]
  ): string {
    const report = {
      timestamp: new Date().toISOString(),
      summary: {
        totalCases: caseValidations.length,
        validCases: caseValidations.filter(v => v.validation.isValid).length,
        invalidCases: caseValidations.filter(v => !v.validation.isValid).length,
        totalDocuments: documentValidations.length,
        validDocuments: documentValidations.filter(v => v.validation.isValid).length,
        invalidDocuments: documentValidations.filter(v => !v.validation.isValid).length
      },
      caseValidations: caseValidations.map(v => ({
        caseId: v.caseData.id,
        title: v.caseData.title,
        isValid: v.validation.isValid,
        errors: v.validation.errors,
        warnings: v.validation.warnings,
        suggestions: v.suggestions
      })),
      documentValidations: documentValidations.map(v => ({
        documentId: v.document.id,
        name: v.document.name,
        isValid: v.validation.isValid,
        errors: v.validation.errors,
        warnings: v.validation.warnings,
        suggestions: v.suggestions
      })),
      recommendations: [
        '定期验证数据完整性',
        '修复所有验证错误',
        '关注验证警告信息',
        '优化数据质量'
      ]
    };

    return JSON.stringify(report, null, 2);
  }

  /**
   * 批量验证案件数据
   */
  override async batchValidateCases(cases: CaseData[]): Promise<TestCaseValidationResult[]> {
    const results: TestCaseValidationResult[] | undefined = undefined;

    for (const caseData of cases) {
      const validation = this.validateCaseData(caseData);
      const suggestions = this.generateSuggestions(caseData);

      results.push({
        caseData,
        validation,
        suggestions
      });
    }

    return results;
  }

  /**
   * 批量验证文档数据
   */
  override async batchValidateDocuments(documents: Document[]): Promise<DocumentValidationResult[]> {
    const results: DocumentValidationResult[] | undefined = undefined;

    for (const document of documents) {
      const validation = this.validateDocumentData(document);
      const suggestions = [];

      // 基于验证结果生成建议
      if (!document.description || document.description.length < 20) {
        suggestions.push('建议补充更详细的文档描述');
      }

      if (!document.tags || document.tags.length === 0) {
        suggestions.push('建议添加标签以便分类和搜索');
      }

      if (document.fileSize > 10 * 1024 * 1024) { // 10MB
        suggestions.push('大文件建议压缩或分割');
      }

      results.push({
        document,
        validation,
        suggestions
      });
    }

    return results;
  }
}

export default CaseValidationUtils;