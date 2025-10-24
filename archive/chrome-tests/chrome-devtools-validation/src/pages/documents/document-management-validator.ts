/**
 * 文档管理验证工具
 */

import { TestExecutionResult } from '../types/test-engine-types';
import { Logger } from '../core/logger';
import { DocumentListPage } from './document-list-page';
import { DocumentDetailPage } from './document-detail-page';
import { DocumentFormPage } from './document-form-page';

export interface DocumentManagementValidationResult {
  valid: boolean;
  errors: string[];
  warnings: string[];
  details: {
    documentList: {
      valid: boolean;
      errors: string[];
      warnings: string[];
      stats: {
        totalDocuments: number;
        totalSize: number;
        byType: Record<string, number>;
        byClient: Record<string, number>;
        byStatus: Record<string, number>;
      };
    };
    documentDetail: {
      valid: boolean;
      errors: string[];
      warnings: string[];
      documentInfo?: any;
    };
    documentForm: {
      valid: boolean;
      errors: string[];
      warnings: string[];
      formValidation: any;
    };
  };
  timestamp: Date;
  executionTime: number;
}

export interface DocumentStatistics {
  total: number;
  totalSize: number;
  byType: Record<string, number>;
  byClient: Record<string, number>;
  byStatus: Record<string, number>;
  avgSize: number;
  recentUploads: number;
  encryptedCount: number;
  publicCount: number;
}

export class DocumentManagementValidator {
  private logger: Logger;
  private documentListPage: DocumentListPage;
  private documentDetailPage: DocumentDetailPage;
  private documentFormPage: DocumentFormPage;

  constructor(baseUrl: string, logger?: Logger) {
    this.logger = logger || new Logger('DocumentManagementValidator');
    this.documentListPage = new DocumentListPage({ baseUrl }, this.logger);
    this.documentDetailPage = new DocumentDetailPage({ baseUrl }, this.logger);
    this.documentFormPage = new DocumentFormPage({ baseUrl }, this.logger);
  }

  /**
   * 验证文档管理模块
   */
  override async validateDocumentManagementModule(
    baseUrl: string,
    executionResult?: TestExecutionResult
  ): Promise<DocumentManagementValidationResult> {
    const startTime = Date.now();
    this.logger.info('开始验证文档管理模块', { baseUrl });

    const result: DocumentManagementValidationResult = {
      valid: true,
      errors: [],
      warnings: [],
      details: {
        documentList: {
          valid: true,
          errors: [],
          warnings: [],
          stats: {
            totalDocuments: 0,
            totalSize: 0,
            byType: {},
            byClient: {},
            byStatus: {}
          }
        },
        documentDetail: {
          valid: true,
          errors: [],
          warnings: []
        },
        documentForm: {
          valid: true,
          errors: [],
          warnings: [],
          formValidation: null
        }
      },
      timestamp: new Date(),
      executionTime: 0
    };

    try {
      // 验证文档列表页面
      this.logger.debug('验证文档列表页面');
      const documentListResult = await this.validateDocumentListPage(baseUrl);
      result.details.documentList = documentListResult;

      if (!documentListResult.valid) {
        result.valid = false;
        result.errors.push(...documentListResult.errors);
      }
      result.warnings.push(...documentListResult.warnings);

      // 验证文档详情页面
      this.logger.debug('验证文档详情页面');
      const documentDetailResult = await this.validateDocumentDetailPage(baseUrl);
      result.details.documentDetail = documentDetailResult;

      if (!documentDetailResult.valid) {
        result.valid = false;
        result.errors.push(...documentDetailResult.errors);
      }
      result.warnings.push(...documentDetailResult.warnings);

      // 验证文档表单页面
      this.logger.debug('验证文档表单页面');
      const documentFormResult = await this.validateDocumentFormPage(baseUrl);
      result.details.documentForm = documentFormResult;

      if (!documentFormResult.valid) {
        result.valid = false;
        result.errors.push(...documentFormResult.errors);
      }
      result.warnings.push(...documentFormResult.warnings);

      // 集成验证结果
      if (executionResult) {
        await this.validateExecutionResults(result, executionResult);
      }

      // 执行综合验证
      await this.performComprehensiveValidation(result);

      result.executionTime = Date.now() - startTime;
      this.logger.info('文档管理模块验证完成', {
        valid: result.valid,
        errors: result.errors.length,
        warnings: result.warnings.length,
        executionTime: result.executionTime
      });

      return result;

    } catch (error) {
      result.valid = false;
      result.errors.push(`验证过程中发生错误: ${error instanceof Error ? error.message : String(error)}`);
      result.executionTime = Date.now() - startTime;

      this.logger.error('文档管理模块验证失败', { error });
      return result;
    }
  }

  /**
   * 验证文档列表页面
   */
  private override async validateDocumentListPage(baseUrl: string): Promise<{
    valid: boolean;
    errors: string[];
    warnings: string[];
    stats: any;
  }> {
    const result = {
      valid: true,
      errors: [] as string[],
      warnings: [] as string[],
      stats: {
        totalDocuments: 0,
        totalSize: 0,
        byType: {} as Record<string, number>,
        byClient: {} as Record<string, number>,
        byStatus: {} as Record<string, number>
      }
    };

    try {
      // 导航到文档列表页面
      await this.documentListPage.navigateToDocumentList();

      // 验证页面元素
      const pageValidation = await this.documentListPage.validateDocumentListPage();
      if (!pageValidation.valid) {
        result.valid = false;
        result.errors.push(`文档列表页面缺少必要元素: ${pageValidation.missingElements.join(', ')}`);
      }

      // 获取文档统计数据
      try {
        const statistics = await this.documentListPage.getDocumentStatistics();
        result.stats = {
          totalDocuments: statistics.total,
          totalSize: statistics.totalSize,
          byType: statistics.byType,
          byClient: statistics.byClient,
          byStatus: statistics.byStatus
        };
      } catch (error) {
        result.warnings.push(`获取文档统计数据失败: ${error instanceof Error ? error.message : String(error)}`);
      }

      // 验证搜索功能
      try {
        await this.documentListPage.searchDocuments({ query: '测试' });
        await this.wait(1000); // 等待搜索结果
        await this.documentListPage.clearFilters();
      } catch (error) {
        result.warnings.push(`搜索功能验证失败: ${error instanceof Error ? error.message : String(error)}`);
      }

      // 验证过滤功能
      try {
        await this.documentListPage.applyFilters({
          type: 'PDF',
          status: 'active'
        });
        await this.wait(1000); // 等待过滤结果
        await this.documentListPage.clearFilters();
      } catch (error) {
        result.warnings.push(`过滤功能验证失败: ${error instanceof Error ? error.message : String(error)}`);
      }

      // 验证排序功能
      try {
        await this.documentListPage.sortDocuments({
          field: 'name',
          order: 'asc'
        });
        await this.wait(1000); // 等待排序结果
      } catch (error) {
        result.warnings.push(`排序功能验证失败: ${error instanceof Error ? error.message : String(error)}`);
      }

      this.logger.debug('文档列表页面验证完成', { valid: result.valid });

    } catch (error) {
      result.valid = false;
      result.errors.push(`文档列表页面验证失败: ${error instanceof Error ? error.message : String(error)}`);
      this.logger.error('文档列表页面验证错误', { error });
    }

    return result;
  }

  /**
   * 验证文档详情页面
   */
  private override async validateDocumentDetailPage(baseUrl: string): Promise<{
    valid: boolean;
    errors: string[];
    warnings: string[];
    documentInfo?: any;
  }> {
    const result = {
      valid: true,
      errors: [] as string[],
      warnings: [] as string[],
      documentInfo: undefined as any
    };

    try {
      // 假设有一个文档ID为'doc-1'的文档
      const documentId = 'doc-1';

      // 导航到文档详情页面
      await this.documentDetailPage.navigateToDocumentDetail(documentId);

      // 验证页面元素
      const pageValidation = await this.documentDetailPage.validateDocumentDetailPage();
      if (!pageValidation.valid) {
        result.valid = false;
        result.errors.push(`文档详情页面缺少必要元素: ${pageValidation.missingElements.join(', ')}`);
      }

      // 获取文档详情
      try {
        const documentDetail = await this.documentDetailPage.getDocumentDetail();
        result.documentInfo = {
          id: documentDetail.id,
          name: documentDetail.name,
          type: documentDetail.type,
          size: documentDetail.size,
          uploadDate: documentDetail.uploadDate,
          uploadedBy: documentDetail.uploadedBy,
          client: documentDetail.client.name,
          status: documentDetail.status,
          hasComments: documentDetail.comments && documentDetail.comments.length > 0,
          hasVersions: documentDetail.versions && documentDetail.versions.length > 0,
          isPublic: documentDetail.isPublic,
          isEncrypted: documentDetail.isEncrypted
        };

        // 验证文档信息完整性
        if (!documentDetail.name) {
          result.warnings.push('文档名称为空');
        }
        if (!documentDetail.type) {
          result.warnings.push('文档类型为空');
        }
        if (documentDetail.size === 0) {
          result.warnings.push('文档大小为0');
        }

      } catch (error) {
        result.warnings.push(`获取文档详情失败: ${error instanceof Error ? error.message : String(error)}`);
      }

      // 验证评论功能
      try {
        const comments = await this.documentDetailPage.getComments();
        // 评论可以为空，所以只是检查功能是否可用
      } catch (error) {
        result.warnings.push(`评论功能验证失败: ${error instanceof Error ? error.message : String(error)}`);
      }

      // 验证版本历史功能
      try {
        const versions = await this.documentDetailPage.getVersionHistory();
        // 版本可以为空，所以只是检查功能是否可用
      } catch (error) {
        result.warnings.push(`版本历史功能验证失败: ${error instanceof Error ? error.message : String(error)}`);
      }

      // 验证关联文档功能
      try {
        const relatedDocuments = await this.documentDetailPage.getRelatedDocuments();
        // 关联文档可以为空，所以只是检查功能是否可用
      } catch (error) {
        result.warnings.push(`关联文档功能验证失败: ${error instanceof Error ? error.message : String(error)}`);
      }

      this.logger.debug('文档详情页面验证完成', { valid: result.valid });

    } catch (error) {
      result.valid = false;
      result.errors.push(`文档详情页面验证失败: ${error instanceof Error ? error.message : String(error)}`);
      this.logger.error('文档详情页面验证错误', { error });
    }

    return result;
  }

  /**
   * 验证文档表单页面
   */
  private override async validateDocumentFormPage(baseUrl: string): Promise<{
    valid: boolean;
    errors: string[];
    warnings: string[];
    formValidation: any;
  }> {
    const result = {
      valid: true,
      errors: [] as string[],
      warnings: [] as string[],
      formValidation: null as any
    };

    try {
      // 导航到创建文档页面
      await this.documentFormPage.navigateToCreateDocument();

      // 验证页面元素
      const pageValidation = await this.documentFormPage.validateDocumentFormPage();
      if (!pageValidation.valid) {
        result.valid = false;
        result.errors.push(`文档表单页面缺少必要元素: ${pageValidation.missingElements.join(', ')}`);
      }

      // 验证表单验证功能
      try {
        const validation = await this.documentFormPage.validateForm();
        result.formValidation = validation;

        if (validation.valid) {
          result.warnings.push('空表单应该有验证错误，但验证通过了');
        }
      } catch (error) {
        result.warnings.push(`表单验证功能检查失败: ${error instanceof Error ? error.message : String(error)}`);
      }

      // 验证字段级验证
      try {
        const nameValidation = await this.documentFormPage.validateFormField('name', '');
        if (nameValidation.valid) {
          result.warnings.push('空名称应该验证失败，但验证通过了');
        }

        const typeValidation = await this.documentFormPage.validateFormField('type', '');
        if (typeValidation.valid) {
          result.warnings.push('空类型应该验证失败，但验证通过了');
        }
      } catch (error) {
        result.warnings.push(`字段验证功能检查失败: ${error instanceof Error ? error.message : String(error)}`);
      }

      // 验证表单数据获取
      try {
        const formData = await this.documentFormPage.getFormData();
        if (!formData || typeof formData !== 'object') {
          result.warnings.push('表单数据获取功能异常');
        }
      } catch (error) {
        result.warnings.push(`表单数据获取功能检查失败: ${error instanceof Error ? error.message : String(error)}`);
      }

      // 验证表单重置功能
      try {
        await this.documentFormPage.resetForm();
      } catch (error) {
        result.warnings.push(`表单重置功能检查失败: ${error instanceof Error ? error.message : String(error)}`);
      }

      // 验证取消功能
      try {
        await this.documentFormPage.cancel();
        // 取消后应该返回文档列表页面，重新导航回来继续验证
        await this.documentFormPage.navigateToCreateDocument();
      } catch (error) {
        result.warnings.push(`取消功能检查失败: ${error instanceof Error ? error.message : String(error)}`);
      }

      this.logger.debug('文档表单页面验证完成', { valid: result.valid });

    } catch (error) {
      result.valid = false;
      result.errors.push(`文档表单页面验证失败: ${error instanceof Error ? error.message : String(error)}`);
      this.logger.error('文档表单页面验证错误', { error });
    }

    return result;
  }

  /**
   * 验证执行结果
   */
  private override async validateExecutionResults(
    result: DocumentManagementValidationResult,
    executionResult: TestExecutionResult
  ): Promise<void> {
    try {
      // 检查是否有失败的测试用例
      const failedTests = executionResult.results.filter(r => !r.passed);
      if (failedTests.length > 0) {
        result.warnings.push(`${failedTests.length} 个测试用例失败`);
        failedTests.forEach(test => {
          result.warnings.push(`  - ${test.name}: ${test.geterror?.().message || '未知错误'}`);
        });
      }

      // 检查执行时间
      if (executionResult.duration > 30000) {
        result.warnings.push('测试执行时间较长，可能存在性能问题');
      }

      // 检查错误率
      const errorRate = failedTests.length / executionResult.results.length;
      if (errorRate > 0.2) {
        result.warnings.push(`测试错误率较高: ${(errorRate * 100).toFixed(1)}%`);
      }

      this.logger.debug('执行结果验证完成', {
        totalTests: executionResult.results.length,
        failedTests: failedTests.length,
        errorRate
      });

    } catch (error) {
      result.warnings.push(`执行结果验证失败: ${error instanceof Error ? error.message : String(error)}`);
    }
  }

  /**
   * 执行综合验证
   */
  private override async performComprehensiveValidation(result: DocumentManagementValidationResult): Promise<void> {
    try {
      // 检查数据一致性
      if (result.details.documentList.stats.totalDocuments === 0) {
        result.warnings.push('文档列表为空，可能影响测试结果');
      }

      // 检查文档详情信息
      if (result.details.documentDetail.documentInfo) {
        const documentInfo = result.details.documentDetail.documentInfo;
        if (documentInfo.size === 0) {
          result.warnings.push('文档大小为0，可能影响相关功能测试');
        }
      }

      // 检查表单验证
      if (result.details.documentForm.formValidation) {
        const validation = result.details.documentForm.formValidation;
        if (Object.keys(validation.errors).length === 0) {
          result.warnings.push('表单验证可能不够严格');
        }
      }

      // 检查各模块之间的关联性
      if (result.details.documentList.valid && result.details.documentDetail.valid) {
        // 文档列表和详情都有效，检查数据一致性
        if (result.details.documentList.stats.totalDocuments > 0 && !result.details.documentDetail.documentInfo) {
          result.warnings.push('文档列表有数据但详情页面无法加载文档信息');
        }
      }

      // 检查文档存储空间
      if (result.details.documentList.stats.totalSize > 1024 * 1024 * 1024) { // 1GB
        result.warnings.push('文档存储空间较大，可能需要清理或优化');
      }

      // 性能相关检查
      if (result.executionTime > 15000) {
        result.warnings.push('验证执行时间较长，可能存在性能问题');
      }

      // 检查加密文档比例
      const encryptedCount = result.details.documentList.stats.byType['encrypted'] || 0;
      const totalCount = result.details.documentList.stats.totalDocuments;
      if (totalCount > 0 && encryptedCount / totalCount > 0.8) {
        result.warnings.push('加密文档比例较高，可能影响性能');
      }

      this.logger.debug('综合验证完成');

    } catch (error) {
      result.warnings.push(`综合验证失败: ${error instanceof Error ? error.message : String(error)}`);
    }
  }

  /**
   * 获取文档统计信息
   */
  override async getDocumentStatistics(): Promise<DocumentStatistics> {
    try {
      await this.documentListPage.navigateToDocumentList();
      const stats = await this.documentListPage.getDocumentStatistics();

      const statistics: DocumentStatistics = {
        total: stats.total,
        totalSize: stats.totalSize,
        byType: stats.byType,
        byClient: stats.byClient,
        byStatus: stats.byStatus,
        avgSize: 0,
        recentUploads: 0,
        encryptedCount: 0,
        publicCount: 0
      };

      // 计算平均大小
      const documents = await this.documentListPage.getDocumentList();
      if (documents.length > 0) {
        statistics.avgSize = statistics.totalSize / documents.length;

        // 计算最近添加的文档（假设7天内）
        const sevenDaysAgo = new Date();
        sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 7);
        statistics.recentUploads = documents.filter(doc =>
          doc.uploadDate >= sevenDaysAgo
        ).length;

        // 计算加密和公开文档数量
        statistics.encryptedCount = documents.filter(doc => doc.isEncrypted).length;
        statistics.publicCount = documents.filter(doc => doc.isPublic).length;
      }

      return statistics;

    } catch (error) {
      this.logger.error('获取文档统计信息失败', { error });
      throw error;
    }
  }

  /**
   * 生成验证报告
   */
  generateValidationReport(result: DocumentManagementValidationResult): string {
    const report = [
      '# 文档管理模块验证报告',
      '',
      `验证时间: ${result.timestamp.toISOString()}`,
      `执行时间: ${result.executionTime}ms`,
      `验证结果: ${result.valid ? '✅ 通过' : '❌ 失败'}`,
      '',
      '## 验证摘要',
      '',
      `- 错误数量: ${result.errors.length}`,
      `- 警告数量: ${result.warnings.length}`,
      '',
      '## 模块验证结果',
      '',
      '### 文档列表页面',
      `状态: ${result.details.documentList.valid ? '✅ 通过' : '❌ 失败'}`,
      `错误: ${result.details.documentList.errors.length}`,
      `警告: ${result.details.documentList.warnings.length}`,
      '',
      '### 文档详情页面',
      `状态: ${result.details.documentDetail.valid ? '✅ 通过' : '❌ 失败'}`,
      `错误: ${result.details.documentDetail.errors.length}`,
      `警告: ${result.details.documentDetail.warnings.length}`,
      '',
      '### 文档表单页面',
      `状态: ${result.details.documentForm.valid ? '✅ 通过' : '❌ 失败'}`,
      `错误: ${result.details.documentForm.errors.length}`,
      `警告: ${result.details.documentForm.warnings.length}`,
      '',
      '## 详细信息',
      '',
      '### 错误列表',
      ...result.errors.map(error => `- ${error}`),
      '',
      '### 警告列表',
      ...result.warnings.map(warning => `- ${warning}`),
      '',
      '## 统计信息',
      '',
      `文档总数: ${result.details.documentList.stats.totalDocuments}`,
      `总大小: ${this.formatFileSize(result.details.documentList.stats.totalSize)}`,
      `平均大小: ${this.formatFileSize(result.details.documentList.stats.totalSize / Math.max(1, result.details.documentList.stats.totalDocuments))}`,
      '',
      '### 按类型分布',
      ...Object.entries(result.details.documentList.stats.byType).map(([type, count]) => `- ${type}: ${count}`),
      '',
      '### 按客户分布',
      ...Object.entries(result.details.documentList.stats.byClient).map(([client, count]) => `- ${client}: ${count}`),
      '',
      '### 按状态分布',
      ...Object.entries(result.details.documentList.stats.byStatus).map(([status, count]) => `- ${status}: ${count}`)
    ];

    return report.join('\n');
  }

  /**
   * 格式化文件大小
   */
  private formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  /**
   * 等待指定时间
   */
  private override async wait(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}