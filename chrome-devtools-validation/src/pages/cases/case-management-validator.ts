/**
 * 案件管理验证工具
 */

import { TestCase, TestExecutionResult } from '../types/test-types';
import { CaseListPage } from './case-list-page';
import { CaseDetailPage } from './case-detail-page';
import { CaseFormPage } from './case-form-page';
import { TestDataGenerator } from '../utils/test-data-generator';
import { Logger } from '../core/logger';

export interface CaseManagementValidationResult {
  valid: boolean;
  errors: string[];
  warnings: string[];
  details: {
    caseListValidation: any;
    caseDetailValidation: any;
    caseFormValidation: any;
    testDataValidation: any;
  };
}

export class CaseManagementValidator {
  private logger: Logger;
  private dataGenerator: TestDataGenerator;

  constructor(logger?: Logger) {
    this.logger = logger || new Logger('CaseManagementValidator');
    this.dataGenerator = new TestDataGenerator();
  }

  /**
   * 验证整个案件管理模块
   */
  override async validateCaseManagementModule(
    baseUrl: string,
    executionResult?: TestExecutionResult
  ): Promise<CaseManagementValidationResult> {
    this.logger.info('开始验证案件管理模块');

    const result: CaseManagementValidationResult = {
      valid: true,
      errors: [],
      warnings: [],
      details: {
        caseListValidation: null,
        caseDetailValidation: null,
        caseFormValidation: null,
        testDataValidation: null
      }
    };

    try {
      // 验证案件列表页面
      result.details.caseListValidation = await this.validateCaseListPage(baseUrl);
      if (!result.details.caseListValidation.valid) {
        result.valid = false;
        result.errors.push(...result.details.caseListValidation.errors);
      }
      result.warnings.push(...result.details.caseListValidation.warnings);

      // 验证案件详情页面
      result.details.caseDetailValidation = await this.validateCaseDetailPage(baseUrl);
      if (!result.details.caseDetailValidation.valid) {
        result.valid = false;
        result.errors.push(...result.details.caseDetailValidation.errors);
      }
      result.warnings.push(...result.details.caseDetailValidation.warnings);

      // 验证案件表单页面
      result.details.caseFormValidation = await this.validateCaseFormPage(baseUrl);
      if (!result.details.caseFormValidation.valid) {
        result.valid = false;
        result.errors.push(...result.details.caseFormValidation.errors);
      }
      result.warnings.push(...result.details.caseFormValidation.warnings);

      // 验证测试数据
      result.details.testDataValidation = await this.validateTestData();
      if (!result.details.testDataValidation.valid) {
        result.valid = false;
        result.errors.push(...result.details.testDataValidation.errors);
      }
      result.warnings.push(...result.details.testDataValidation.warnings);

      // 验证测试执行结果（如果提供）
      if (executionResult) {
        const executionValidation = this.validateExecutionResult(executionResult);
        if (!executionValidation.valid) {
          result.valid = false;
          result.errors.push(...executionValidation.errors);
        }
        result.warnings.push(...executionValidation.warnings);
      }

    } catch (error) {
      result.valid = false;
      result.errors.push(`验证过程中发生错误: ${error instanceof Error ? error.message : error}`);
    }

    this.logger.info('案件管理模块验证完成', {
      valid: result.valid,
      errorsCount: result.errors.length,
      warningsCount: result.warnings.length
    });

    return result;
  }

  /**
   * 验证案件列表页面
   */
  private override async validateCaseListPage(baseUrl: string): Promise<{
    valid: boolean;
    errors: string[];
    warnings: string[];
    details: any;
  }> {
    this.logger.debug('验证案件列表页面');

    const errors: string[] | undefined = undefined;
    const warnings: string[] | undefined = undefined;
    const details: any = {};

    const caseListPage = new CaseListPage({ baseUrl }, this.logger);

    try {
      // 验证页面元素
      const elementValidation = await caseListPage.validateCaseListPage();
      details.elementValidation = elementValidation;

      if (!elementValidation.valid) {
        errors.push(`案件列表页面缺少必要元素: ${elementValidation.missingElements.join(', ')}`);
      }

      // 验证案件列表获取
      try {
        const caseList = await caseListPage.getCaseList();
        details.caseList = caseList;
        details.caseListCount = caseList.length;

        if (caseList.length === 0) {
          warnings.push('案件列表为空，可能影响某些测试用例');
        }

        // 验证案件数据结构
        const dataStructureValidation = this.validateCaseListDataStructure(caseList);
        if (!dataStructureValidation.valid) {
          errors.push(...dataStructureValidation.errors);
        }
        details.dataStructureValidation = dataStructureValidation;

      } catch (error) {
        errors.push(`获取案件列表失败: ${error instanceof Error ? error.message : error}`);
      }

      // 验证搜索功能
      try {
        await caseListPage.searchCases('测试');
        details.searchTest = { success: true };
      } catch (error) {
        errors.push(`搜索功能测试失败: ${error instanceof Error ? error.message : error}`);
        details.searchTest = { success: false, error };
      }

      // 验证过滤功能
      try {
        await caseListPage.applyFilters({
          status: 'active',
          priority: 'high'
        });
        details.filterTest = { success: true };
      } catch (error) {
        warnings.push(`过滤功能测试失败: ${error instanceof Error ? error.message : error}`);
        details.filterTest = { success: false, error };
      }

      // 验证分页功能
      try {
        const totalPages = await caseListPage.getTotalPages();
        const currentPage = await caseListPage.getCurrentPage();
        details.paginationTest = {
          success: true,
          totalPages,
          currentPage
        };

        if (totalPages > 1) {
          try {
            await caseListPage.nextPage();
            const newPage = await caseListPage.getCurrentPage();
            details.paginationTest.navigationSuccess = newPage === currentPage + 1;
          } catch (error) {
            warnings.push(`分页导航测试失败: ${error instanceof Error ? error.message : error}`);
          }
        }

      } catch (error) {
        warnings.push(`分页功能测试失败: ${error instanceof Error ? error.message : error}`);
        details.paginationTest = { success: false, error };
      }

    } catch (error) {
      errors.push(`案件列表页面验证失败: ${error instanceof Error ? error.message : error}`);
    }

    return {
      valid: errors.length === 0,
      errors,
      warnings,
      details
    };
  }

  /**
   * 验证案件详情页面
   */
  private override async validateCaseDetailPage(baseUrl: string): Promise<{
    valid: boolean;
    errors: string[];
    warnings: string[];
    details: any;
  }> {
    this.logger.debug('验证案件详情页面');

    const errors: string[] | undefined = undefined;
    const warnings: string[] | undefined = undefined;
    const details: any = {};

    const caseDetailPage = new CaseDetailPage({ baseUrl }, this.logger);

    try {
      // 验证页面元素
      const elementValidation = await caseDetailPage.validateCaseDetailPage();
      details.elementValidation = elementValidation;

      if (!elementValidation.valid) {
        errors.push(`案件详情页面缺少必要元素: ${elementValidation.missingElements.join(', ')}`);
      }

      // 验证案件详情获取
      try {
        const caseDetail = await caseDetailPage.getCaseDetail();
        details.caseDetail = caseDetail;

        // 验证案件详情数据结构
        const dataStructureValidation = this.validateCaseDetailDataStructure(caseDetail);
        if (!dataStructureValidation.valid) {
          errors.push(...dataStructureValidation.errors);
        }
        details.dataStructureValidation = dataStructureValidation;

      } catch (error) {
        errors.push(`获取案件详情失败: ${error instanceof Error ? error.message : error}`);
      }

      // 验证里程碑列表
      try {
        const milestones = await caseDetailPage.getMilestoneList();
        details.milestones = milestones;
        details.milestonesCount = milestones.length;
        details.milestonesTest = { success: true };
      } catch (error) {
        warnings.push(`里程碑功能测试失败: ${error instanceof Error ? error.message : error}`);
        details.milestonesTest = { success: false, error };
      }

      // 验证文档列表
      try {
        const documents = await caseDetailPage.getDocumentList();
        details.documents = documents;
        details.documentsCount = documents.length;
        details.documentsTest = { success: true };
      } catch (error) {
        warnings.push(`文档功能测试失败: ${error instanceof Error ? error.message : error}`);
        details.documentsTest = { success: false, error };
      }

      // 验证时间线
      try {
        const timeline = await caseDetailPage.getTimeline();
        details.timeline = timeline;
        details.timelineCount = timeline.length;
        details.timelineTest = { success: true };
      } catch (error) {
        warnings.push(`时间线功能测试失败: ${error instanceof Error ? error.message : error}`);
        details.timelineTest = { success: false, error };
      }

      // 验证财务记录
      try {
        const financialRecords = await caseDetailPage.getFinancialRecords();
        details.financialRecords = financialRecords;
        details.financialRecordsCount = financialRecords.length;
        details.financialTest = { success: true };
      } catch (error) {
        warnings.push(`财务记录功能测试失败: ${error instanceof Error ? error.message : error}`);
        details.financialTest = { success: false, error };
      }

    } catch (error) {
      errors.push(`案件详情页面验证失败: ${error instanceof Error ? error.message : error}`);
    }

    return {
      valid: errors.length === 0,
      errors,
      warnings,
      details
    };
  }

  /**
   * 验证案件表单页面
   */
  private override async validateCaseFormPage(baseUrl: string): Promise<{
    valid: boolean;
    errors: string[];
    warnings: string[];
    details: any;
  }> {
    this.logger.debug('验证案件表单页面');

    const errors: string[] | undefined = undefined;
    const warnings: string[] | undefined = undefined;
    const details: any = {};

    const caseFormPage = new CaseFormPage({ baseUrl }, this.logger);

    try {
      // 验证表单页面导航
      try {
        await caseFormPage.navigateToCreateCase();
        details.navigationTest = { success: true };
      } catch (error) {
        errors.push(`表单页面导航失败: ${error instanceof Error ? error.message : error}`);
        details.navigationTest = { success: false, error };
      }

      // 验证表单验证功能
      try {
        const validation = await caseFormPage.validateForm();
        details.formValidation = validation;

        if (validation.valid) {
          warnings.push('空表单验证应该失败，但返回成功');
        } else {
          details.formValidationTest = { success: true };
        }
      } catch (error) {
        errors.push(`表单验证功能测试失败: ${error instanceof Error ? error.message : error}`);
        details.formValidationTest = { success: false, error };
      }

      // 验证表单数据获取
      try {
        const formData = await caseFormPage.getFormData();
        details.formData = formData;
        details.formDataTest = { success: true };

        // 验证表单数据结构
        const dataStructureValidation = this.validateFormDataStructure(formData);
        if (!dataStructureValidation.valid) {
          errors.push(...dataStructureValidation.errors);
        }
        details.formDataStructureValidation = dataStructureValidation;

      } catch (error) {
        errors.push(`获取表单数据失败: ${error instanceof Error ? error.message : error}`);
        details.formDataTest = { success: false, error };
      }

      // 验证客户搜索功能
      try {
        await caseFormPage.searchClient('测试');
        details.clientSearchTest = { success: true };
      } catch (error) {
        warnings.push(`客户搜索功能测试失败: ${error instanceof Error ? error.message : error}`);
        details.clientSearchTest = { success: false, error };
      }

      // 验证分配人搜索功能
      try {
        await caseFormPage.searchAssignedTo('律师');
        details.assignedToSearchTest = { success: true };
      } catch (error) {
        warnings.push(`分配人搜索功能测试失败: ${error instanceof Error ? error.message : error}`);
        details.assignedToSearchTest = { success: false, error };
      }

      // 验证标签功能
      try {
        const initialTags = await caseFormPage.getTags();
        details.initialTags = initialTags;

        await caseFormPage.addTags(['测试标签']);
        const updatedTags = await caseFormPage.getTags();
        details.updatedTags = updatedTags;
        details.tagsTest = {
          success: true,
          tagsAdded: updatedTags.length > initialTags.length
        };
      } catch (error) {
        warnings.push(`标签功能测试失败: ${error instanceof Error ? error.message : error}`);
        details.tagsTest = { success: false, error };
      }

      // 验证表单清除功能
      try {
        await caseFormPage.clearForm();
        const clearedFormData = await caseFormPage.getFormData();
        details.clearedFormData = clearedFormData;
        details.clearFormTest = { success: true };
      } catch (error) {
        warnings.push(`表单清除功能测试失败: ${error instanceof Error ? error.message : error}`);
        details.clearFormTest = { success: false, error };
      }

    } catch (error) {
      errors.push(`案件表单页面验证失败: ${error instanceof Error ? error.message : error}`);
    }

    return {
      valid: errors.length === 0,
      errors,
      warnings,
      details
    };
  }

  /**
   * 验证测试数据
   */
  private override async validateTestData(): Promise<{
    valid: boolean;
    errors: string[];
    warnings: string[];
    details: any;
  }> {
    this.logger.debug('验证测试数据');

    const errors: string[] | undefined = undefined;
    const warnings: string[] | undefined = undefined;
    const details: any = {};

    try {
      // 验证用户数据生成
      try {
        const userData = this.dataGenerator.generateUserData();
        details.userData = userData;
        details.userDataTest = { success: true };

        const userDataValidation = this.validateUserDataStructure(userData);
        if (!userDataValidation.valid) {
          errors.push(...userDataValidation.errors);
        }
        details.userDataValidation = userDataValidation;

      } catch (error) {
        errors.push(`用户数据生成失败: ${error instanceof Error ? error.message : error}`);
        details.userDataTest = { success: false, error };
      }

      // 验证客户数据生成
      try {
        const clientData = this.dataGenerator.generateClientData();
        details.clientData = clientData;
        details.clientDataTest = { success: true };

        const clientDataValidation = this.validateClientDataStructure(clientData);
        if (!clientDataValidation.valid) {
          errors.push(...clientDataValidation.errors);
        }
        details.clientDataValidation = clientDataValidation;

      } catch (error) {
        errors.push(`客户数据生成失败: ${error instanceof Error ? error.message : error}`);
        details.clientDataTest = { success: false, error };
      }

      // 验证案件数据生成
      try {
        const caseData = this.dataGenerator.generateCaseData();
        details.caseData = caseData;
        details.caseDataTest = { success: true };

        const caseDataValidation = this.validateCaseDataStructure(caseData);
        if (!caseDataValidation.valid) {
          errors.push(...caseDataValidation.errors);
        }
        details.caseDataValidation = caseDataValidation;

      } catch (error) {
        errors.push(`案件数据生成失败: ${error instanceof Error ? error.message : error}`);
        details.caseDataTest = { success: false, error };
      }

      // 验证测试数据集生成
      try {
        const testDataSet = this.dataGenerator.generateTestDataSet({
          userCount: 2,
          clientCount: 3,
          caseCount: 5,
          documentCount: 10
        });
        details.testDataSet = testDataSet;
        details.testDataSetTest = { success: true };

        if (testDataSet.users.length !== 2) {
          warnings.push(`用户数量不匹配: 期望2，实际${testDataSet.users.length}`);
        }
        if (testDataSet.clients.length !== 3) {
          warnings.push(`客户数量不匹配: 期望3，实际${testDataSet.clients.length}`);
        }
        if (testDataSet.cases.length !== 5) {
          warnings.push(`案件数量不匹配: 期望5，实际${testDataSet.cases.length}`);
        }

      } catch (error) {
        errors.push(`测试数据集生成失败: ${error instanceof Error ? error.message : error}`);
        details.testDataSetTest = { success: false, error };
      }

    } catch (error) {
      errors.push(`测试数据验证失败: ${error instanceof Error ? error.message : error}`);
    }

    return {
      valid: errors.length === 0,
      errors,
      warnings,
      details
    };
  }

  /**
   * 验证测试执行结果
   */
  private validateExecutionResult(executionResult: TestExecutionResult): {
    valid: boolean;
    errors: string[];
    warnings: string[];
  } {
    const errors: string[] | undefined = undefined;
    const warnings: string[] | undefined = undefined;

    // 验证测试结果结构
    if (!executionResult.suites || executionResult.suites.length === 0) {
      errors.push('测试执行结果缺少测试套件信息');
    }

    // 验证测试套件结果
    executionResult.suites.forEach(suite => {
      if (!suite.id || !suite.name) {
        errors.push(`测试套件缺少必要信息: ${JSON.stringify(suite)}`);
      }

      if (suite.testCases && suite.testCases.length > 0) {
        suite.testCases.forEach(testCase => {
          if (!testCase.id || !testCase.name) {
            errors.push(`测试用例缺少必要信息: ${JSON.stringify(testCase)}`);
          }

          if (testCase.result === 'failed' && !testCase.error) {
            warnings.push(`失败的测试用例缺少错误信息: ${testCase.name}`);
          }

          if (testCase.result === 'passed' && testCase.error) {
            warnings.push(`通过的测试用例包含错误信息: ${testCase.name}`);
          }
        });
      }
    });

    // 验证总体统计
    if (executionResult.summary) {
      const { passed, failed, skipped } = executionResult.summary;
      const total = passed + failed + skipped;

      if (total === 0) {
        warnings.push('没有执行任何测试用例');
      }

      if (failed > 0) {
        errors.push(`${failed} 个测试用例失败`);
      }

      if (passed + failed + skipped !== total) {
        warnings.push('测试统计数量不匹配');
      }
    } else {
      warnings.push('测试执行结果缺少统计信息');
    }

    // 验证性能指标
    if (executionResult.performance) {
      const { duration, memoryUsage } = executionResult.performance;

      if (duration > 60000) {
        warnings.push(`测试执行时间过长: ${duration}ms`);
      }

      if (memoryUsage && memoryUsage.peak > 500 * 1024 * 1024) {
        warnings.push(`内存使用过高: ${Math.round(memoryUsage.peak / 1024 / 1024)}MB`);
      }
    } else {
      warnings.push('测试执行结果缺少性能信息');
    }

    return {
      valid: errors.length === 0,
      errors,
      warnings
    };
  }

  /**
   * 验证案件列表数据结构
   */
  private validateCaseListDataStructure(caseList: any[]): { valid: boolean; errors: string[] } {
    const errors: string[] | undefined = undefined;

    if (!Array.isArray(caseList)) {
      errors.push('案件列表不是数组');
      return { valid: false, errors };
    }

    caseList.forEach((caseItem, index) => {
      const prefix = `案件列表项[${index}]`;

      if (!caseItem.id) {
        errors.push(`${prefix} 缺少id字段`);
      }

      if (!caseItem.title) {
        errors.push(`${prefix} 缺少title字段`);
      }

      if (!caseItem.caseNumber) {
        errors.push(`${prefix} 缺少caseNumber字段`);
      }

      if (!caseItem.type) {
        errors.push(`${prefix} 缺少type字段`);
      }

      if (!caseItem.status) {
        errors.push(`${prefix} 缺少status字段`);
      }

      if (!caseItem.priority) {
        errors.push(`${prefix} 缺少priority字段`);
      }

      if (!caseItem.client) {
        errors.push(`${prefix} 缺少client字段`);
      }

      if (!caseItem.assignedTo) {
        errors.push(`${prefix} 缺少assignedTo字段`);
      }

      if (!caseItem.createdDate) {
        errors.push(`${prefix} 缺少createdDate字段`);
      }
    });

    return { valid: errors.length === 0, errors };
  }

  /**
   * 验证案件详情数据结构
   */
  private validateCaseDetailDataStructure(caseDetail: any): { valid: boolean; errors: string[] } {
    const errors: string[] | undefined = undefined;

    const requiredFields = [
      'id', 'title', 'caseNumber', 'type', 'status', 'priority',
      'client', 'assignedTo', 'description', 'estimatedValue',
      'createdDate', 'updatedDate', 'tags', 'milestones'
    ];

    requiredFields.forEach(field => {
      if (caseDetail[field] === undefined || caseDetail[field] === null) {
        errors.push(`案件详情缺少${field}字段`);
      }
    });

    // 验证日期字段
    const dateFields = ['createdDate', 'updatedDate'];
    dateFields.forEach(field => {
      if (caseDetail[field] && !(caseDetail[field] instanceof Date)) {
        errors.push(`案件详情${field}字段不是Date类型`);
      }
    });

    // 验证数组字段
    const arrayFields = ['tags', 'milestones'];
    arrayFields.forEach(field => {
      if (caseDetail[field] && !Array.isArray(caseDetail[field])) {
        errors.push(`案件详情${field}字段不是数组类型`);
      }
    });

    return { valid: errors.length === 0, errors };
  }

  /**
   * 验证表单数据结构
   */
  private validateFormDataStructure(formData: any): { valid: boolean; errors: string[] } {
    const errors: string[] | undefined = undefined;

    const expectedFields = [
      'title', 'description', 'type', 'priority', 'client',
      'assignedTo', 'dueDate', 'estimatedValue', 'tags'
    ];

    expectedFields.forEach(field => {
      if (formData[field] === undefined) {
        errors.push(`表单数据缺少${field}字段`);
      }
    });

    // 验证日期字段
    if (formData.dueDate && !(formData.dueDate instanceof Date)) {
      errors.push('表单数据dueDate字段不是Date类型');
    }

    // 验证数字字段
    if (formData.estimatedValue && typeof formData.estimatedValue !== 'number') {
      errors.push('表单数据estimatedValue字段不是数字类型');
    }

    // 验证数组字段
    if (formData.tags && !Array.isArray(formData.tags)) {
      errors.push('表单数据tags字段不是数组类型');
    }

    return { valid: errors.length === 0, errors };
  }

  /**
   * 验证用户数据结构
   */
  private validateUserDataStructure(userData: any): { valid: boolean; errors: string[] } {
    const errors: string[] | undefined = undefined;

    const requiredFields = [
      'id', 'username', 'email', 'firstName', 'lastName',
      'role', 'department', 'phone', 'status', 'createdAt'
    ];

    requiredFields.forEach(field => {
      if (userData[field] === undefined || userData[field] === null) {
        errors.push(`用户数据缺少${field}字段`);
      }
    });

    return { valid: errors.length === 0, errors };
  }

  /**
   * 验证客户数据结构
   */
  private validateClientDataStructure(clientData: any): { valid: boolean; errors: string[] } {
    const errors: string[] | undefined = undefined;

    const requiredFields = [
      'id', 'name', 'type', 'industry', 'contactPerson',
      'email', 'phone', 'address', 'status', 'createdAt'
    ];

    requiredFields.forEach(field => {
      if (clientData[field] === undefined || clientData[field] === null) {
        errors.push(`客户数据缺少${field}字段`);
      }
    });

    return { valid: errors.length === 0, errors };
  }

  /**
   * 验证案件数据结构
   */
  private validateCaseDataStructure(caseData: any): { valid: boolean; errors: string[] } {
    const errors: string[] | undefined = undefined;

    const requiredFields = [
      'id', 'title', 'caseNumber', 'type', 'status', 'priority',
      'client', 'assignedTo', 'description', 'estimatedValue',
      'createdDate', 'updatedDate', 'tags'
    ];

    requiredFields.forEach(field => {
      if (caseData[field] === undefined || caseData[field] === null) {
        errors.push(`案件数据缺少${field}字段`);
      }
    });

    return { valid: errors.length === 0, errors };
  }

  /**
   * 生成验证报告
   */
  generateValidationReport(result: CaseManagementValidationResult): string {
    const report = [
      '# 案件管理模块验证报告',
      '',
      `## 总体状态: ${result.valid ? '✅ 通过' : '❌ 失败'}`,
      '',
      `**错误数量**: ${result.errors.length}`,
      `**警告数量**: ${result.warnings.length}`,
      '',
      '## 详细结果',
      ''
    ];

    // 案件列表页面验证结果
    if (result.details.caseListValidation) {
      report.push(
        '### 案件列表页面',
        `- **状态**: ${result.details.caseListValidation.valid ? '✅ 通过' : '❌ 失败'}`,
        `- **错误**: ${result.details.caseListValidation.errors.length}`,
        `- **警告**: ${result.details.caseListValidation.warnings.length}`,
        ''
      );
    }

    // 案件详情页面验证结果
    if (result.details.caseDetailValidation) {
      report.push(
        '### 案件详情页面',
        `- **状态**: ${result.details.caseDetailValidation.valid ? '✅ 通过' : '❌ 失败'}`,
        `- **错误**: ${result.details.caseDetailValidation.errors.length}`,
        `- **警告**: ${result.details.caseDetailValidation.warnings.length}`,
        ''
      );
    }

    // 案件表单页面验证结果
    if (result.details.caseFormValidation) {
      report.push(
        '### 案件表单页面',
        `- **状态**: ${result.details.caseFormValidation.valid ? '✅ 通过' : '❌ 失败'}`,
        `- **错误**: ${result.details.caseFormValidation.errors.length}`,
        `- **警告**: ${result.details.caseFormValidation.warnings.length}`,
        ''
      );
    }

    // 测试数据验证结果
    if (result.details.testDataValidation) {
      report.push(
        '### 测试数据',
        `- **状态**: ${result.details.testDataValidation.valid ? '✅ 通过' : '❌ 失败'}`,
        `- **错误**: ${result.details.testDataValidation.errors.length}`,
        `- **警告**: ${result.details.testDataValidation.warnings.length}`,
        ''
      );
    }

    // 错误详情
    if (result.errors.length > 0) {
      report.push('## 错误详情', '');
      result.errors.forEach((error, index) => {
        report.push(`${index + 1}. ${error}`);
      });
      report.push('');
    }

    // 警告详情
    if (result.warnings.length > 0) {
      report.push('## 警告详情', '');
      result.warnings.forEach((warning, index) => {
        report.push(`${index + 1}. ${warning}`);
      });
      report.push('');
    }

    // 建议
    report.push('## 建议', '');
    if (result.valid) {
      report.push('- 案件管理模块验证通过，可以开始执行测试');
      report.push('- 建议定期运行验证以确保模块功能正常');
    } else {
      report.push('- 请修复所有错误后再次运行验证');
      report.push('- 建议优先修复关键功能相关的错误');
      report.push('- 检查页面元素选择器是否正确');
    }

    return report.join('\n');
  }
}