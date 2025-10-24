/**
 * 客户管理验证工具
 */

import { TestExecutionResult } from '../../types/test-engine-types';
import { Logger } from '../../core/logger';
import { ClientListPage } from './client-list-page';
import { ClientDetailPage } from './client-detail-page';
import { ClientFormPage } from './client-form-page';

export interface ClientManagementValidationResult {
  valid: boolean;
  errors: string[];
  warnings: string[];
  details: {
    clientList: {
      valid: boolean;
      errors: string[];
      warnings: string[];
      stats: {
        totalClients: number;
        activeClients: number;
        byType: Record<string, number>;
        byIndustry: Record<string, number>;
      };
    };
    clientDetail: {
      valid: boolean;
      errors: string[];
      warnings: string[];
      clientInfo?: any;
    };
    clientForm: {
      valid: boolean;
      errors: string[];
      warnings: string[];
      formValidation: any;
    };
  };
  timestamp: Date;
  executionTime: number;
}

export interface ClientStatistics {
  total: number;
  active: number;
  inactive: number;
  byType: Record<string, number>;
  byIndustry: Record<string, number>;
  avgCaseCount: number;
  recentAdditions: number;
}

export class ClientManagementValidator {
  private logger: Logger;
  private clientListPage: ClientListPage;
  private clientDetailPage: ClientDetailPage;
  private clientFormPage: ClientFormPage;

  constructor(baseUrl: string, logger?: Logger) {
    this.logger = logger || new Logger('ClientManagementValidator');
    this.clientListPage = new ClientListPage({ baseUrl }, this.logger);
    this.clientDetailPage = new ClientDetailPage({ baseUrl }, this.logger);
    this.clientFormPage = new ClientFormPage({ baseUrl }, this.logger);
  }

  /**
   * 验证客户管理模块
   */
  override async validateClientManagementModule(
    baseUrl: string,
    executionResult?: TestExecutionResult
  ): Promise<ClientManagementValidationResult> {
    const startTime = Date.now();
    this.logger.info('开始验证客户管理模块', { baseUrl });

    const result: ClientManagementValidationResult = {
      valid: true,
      errors: [],
      warnings: [],
      details: {
        clientList: {
          valid: true,
          errors: [],
          warnings: [],
          stats: {
            totalClients: 0,
            activeClients: 0,
            byType: {},
            byIndustry: {}
          }
        },
        clientDetail: {
          valid: true,
          errors: [],
          warnings: []
        },
        clientForm: {
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
      // 验证客户列表页面
      this.logger.debug('验证客户列表页面');
      const clientListResult = await this.validateClientListPage(baseUrl);
      result.details.clientList = clientListResult;

      if (!clientListResult.valid) {
        result.valid = false;
        result.errors.push(...clientListResult.errors);
      }
      result.warnings.push(...clientListResult.warnings);

      // 验证客户详情页面
      this.logger.debug('验证客户详情页面');
      const clientDetailResult = await this.validateClientDetailPage(baseUrl);
      result.details.clientDetail = clientDetailResult;

      if (!clientDetailResult.valid) {
        result.valid = false;
        result.errors.push(...clientDetailResult.errors);
      }
      result.warnings.push(...clientDetailResult.warnings);

      // 验证客户表单页面
      this.logger.debug('验证客户表单页面');
      const clientFormResult = await this.validateClientFormPage(baseUrl);
      result.details.clientForm = clientFormResult;

      if (!clientFormResult.valid) {
        result.valid = false;
        result.errors.push(...clientFormResult.errors);
      }
      result.warnings.push(...clientFormResult.warnings);

      // 集成验证结果
      if (executionResult) {
        await this.validateExecutionResults(result, executionResult);
      }

      // 执行综合验证
      await this.performComprehensiveValidation(result);

      result.executionTime = Date.now() - startTime;
      this.logger.info('客户管理模块验证完成', {
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

      this.logger.error('客户管理模块验证失败', { error });
      return result;
    }
  }

  /**
   * 验证客户列表页面
   */
  private override async validateClientListPage(baseUrl: string): Promise<{
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
        totalClients: 0,
        activeClients: 0,
        byType: {} as Record<string, number>,
        byIndustry: {} as Record<string, number>
      }
    };

    try {
      // 导航到客户列表页面
      await this.clientListPage.navigateToClientList();

      // 验证页面元素
      const pageValidation = await this.clientListPage.validateClientListPage();
      if (!pageValidation.valid) {
        result.valid = false;
        result.errors.push(`客户列表页面缺少必要元素: ${pageValidation.missingElements.join(', ')}`);
      }

      // 获取客户统计数据
      try {
        const statistics = await this.clientListPage.getClientStatistics();
        result.stats = {
          totalClients: statistics.total,
          activeClients: statistics.active,
          byType: statistics.byType,
          byIndustry: statistics.byIndustry
        };
      } catch (error) {
        result.warnings.push(`获取客户统计数据失败: ${error instanceof Error ? error.message : String(error)}`);
      }

      // 验证搜索功能
      try {
        await this.clientListPage.searchClients('测试');
        await this.wait(1000); // 等待搜索结果
        await this.clientListPage.clearFilters();
      } catch (error) {
        result.warnings.push(`搜索功能验证失败: ${error instanceof Error ? error.message : String(error)}`);
      }

      // 验证过滤功能
      try {
        await this.clientListPage.applyFilters({
          type: 'company',
          status: 'active'
        });
        await this.wait(1000); // 等待过滤结果
        await this.clientListPage.clearFilters();
      } catch (error) {
        result.warnings.push(`过滤功能验证失败: ${error instanceof Error ? error.message : String(error)}`);
      }

      // 验证排序功能
      try {
        await this.clientListPage.sortClients({
          field: 'name',
          order: 'asc'
        });
        await this.wait(1000); // 等待排序结果
      } catch (error) {
        result.warnings.push(`排序功能验证失败: ${error instanceof Error ? error.message : String(error)}`);
      }

      this.logger.debug('客户列表页面验证完成', { valid: result.valid });

    } catch (error) {
      result.valid = false;
      result.errors.push(`客户列表页面验证失败: ${error instanceof Error ? error.message : String(error)}`);
      this.logger.error('客户列表页面验证错误', { error });
    }

    return result;
  }

  /**
   * 验证客户详情页面
   */
  private override async validateClientDetailPage(baseUrl: string): Promise<{
    valid: boolean;
    errors: string[];
    warnings: string[];
    clientInfo?: any;
  }> {
    const result = {
      valid: true,
      errors: [] as string[],
      warnings: [] as string[],
      clientInfo: undefined as any
    };

    try {
      // 假设有一个客户ID为'client-1'的客户
      const clientId = 'client-1';

      // 导航到客户详情页面
      await this.clientDetailPage.navigateToClientDetail(clientId);

      // 验证页面元素
      const pageValidation = await this.clientDetailPage.validateClientDetailPage();
      if (!pageValidation.valid) {
        result.valid = false;
        result.errors.push(`客户详情页面缺少必要元素: ${pageValidation.missingElements.join(', ')}`);
      }

      // 获取客户详情
      try {
        const clientDetail = await this.clientDetailPage.getClientDetail();
        result.clientInfo = {
          id: clientDetail.id,
          name: clientDetail.name,
          type: clientDetail.type,
          status: clientDetail.status,
          caseCount: clientDetail.caseCount,
          hasContacts: clientDetail.contacts.length > 0,
          hasCases: clientDetail.cases.length > 0,
          hasDocuments: clientDetail.documents.length > 0
        };

        // 验证客户信息完整性
        if (!clientDetail.name) {
          result.warnings.push('客户名称为空');
        }
        if (!clientDetail.email) {
          result.warnings.push('客户邮箱为空');
        }
        if (!clientDetail.phone) {
          result.warnings.push('客户电话为空');
        }

      } catch (error) {
        result.warnings.push(`获取客户详情失败: ${error instanceof Error ? error.message : String(error)}`);
      }

      // 验证联系人功能
      try {
        const contacts = await this.clientDetailPage.getContactList();
        if (contacts.length === 0) {
          result.warnings.push('客户没有联系人信息');
        } else {
          // 检查主要联系人
          const primaryContacts = contacts.filter(contact => contact.isPrimary);
          if (primaryContacts.length === 0) {
            result.warnings.push('客户没有主要联系人');
          }
        }
      } catch (error) {
        result.warnings.push(`联系人功能验证失败: ${error instanceof Error ? error.message : String(error)}`);
      }

      // 验证案件功能
      try {
        const cases = await this.clientDetailPage.getCaseList();
        // 案件可以为空，所以只是检查功能是否可用
      } catch (error) {
        result.warnings.push(`案件功能验证失败: ${error instanceof Error ? error.message : String(error)}`);
      }

      // 验证文档功能
      try {
        const documents = await this.clientDetailPage.getDocumentList();
        // 文档可以为空，所以只是检查功能是否可用
      } catch (error) {
        result.warnings.push(`文档功能验证失败: ${error instanceof Error ? error.message : String(error)}`);
      }

      this.logger.debug('客户详情页面验证完成', { valid: result.valid });

    } catch (error) {
      result.valid = false;
      result.errors.push(`客户详情页面验证失败: ${error instanceof Error ? error.message : String(error)}`);
      this.logger.error('客户详情页面验证错误', { error });
    }

    return result;
  }

  /**
   * 验证客户表单页面
   */
  private override async validateClientFormPage(baseUrl: string): Promise<{
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
      // 导航到创建客户页面
      await this.clientFormPage.navigateToCreateClient();

      // 验证页面元素
      const pageValidation = await this.clientFormPage.validateClientFormPage();
      if (!pageValidation.valid) {
        result.valid = false;
        result.errors.push(`客户表单页面缺少必要元素: ${pageValidation.missingElements.join(', ')}`);
      }

      // 验证表单验证功能
      try {
        const validation = await this.clientFormPage.validateForm();
        result.formValidation = validation;

        if (validation.valid) {
          result.warnings.push('空表单应该有验证错误，但验证通过了');
        }
      } catch (error) {
        result.warnings.push(`表单验证功能检查失败: ${error instanceof Error ? error.message : String(error)}`);
      }

      // 验证字段级验证
      try {
        const emailValidation = await this.clientFormPage.validateFormField('email', 'invalid-email');
        if (emailValidation.valid) {
          result.warnings.push('无效邮箱地址应该验证失败，但验证通过了');
        }

        const phoneValidation = await this.clientFormPage.validateFormField('phone', '123');
        if (phoneValidation.valid) {
          result.warnings.push('无效电话号码应该验证失败，但验证通过了');
        }
      } catch (error) {
        result.warnings.push(`字段验证功能检查失败: ${error instanceof Error ? error.message : String(error)}`);
      }

      // 验证表单数据获取
      try {
        const formData = await this.clientFormPage.getFormData();
        if (!formData || typeof formData !== 'object') {
          result.warnings.push('表单数据获取功能异常');
        }
      } catch (error) {
        result.warnings.push(`表单数据获取功能检查失败: ${error instanceof Error ? error.message : String(error)}`);
      }

      // 验证表单重置功能
      try {
        await this.clientFormPage.resetForm();
      } catch (error) {
        result.warnings.push(`表单重置功能检查失败: ${error instanceof Error ? error.message : String(error)}`);
      }

      // 验证取消功能
      try {
        await this.clientFormPage.cancel();
        // 取消后应该返回客户列表页面，重新导航回来继续验证
        await this.clientFormPage.navigateToCreateClient();
      } catch (error) {
        result.warnings.push(`取消功能检查失败: ${error instanceof Error ? error.message : String(error)}`);
      }

      this.logger.debug('客户表单页面验证完成', { valid: result.valid });

    } catch (error) {
      result.valid = false;
      result.errors.push(`客户表单页面验证失败: ${error instanceof Error ? error.message : String(error)}`);
      this.logger.error('客户表单页面验证错误', { error });
    }

    return result;
  }

  /**
   * 验证执行结果
   */
  private override async validateExecutionResults(
    result: ClientManagementValidationResult,
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
  private override async performComprehensiveValidation(result: ClientManagementValidationResult): Promise<void> {
    try {
      // 检查数据一致性
      if (result.details.clientList.stats.totalClients === 0) {
        result.warnings.push('客户列表为空，可能影响测试结果');
      }

      // 检查客户详情信息
      if (result.details.clientDetail.clientInfo) {
        const clientInfo = result.details.clientDetail.clientInfo;
        if (clientInfo.caseCount === 0) {
          result.warnings.push('客户没有关联案件，可能影响相关功能测试');
        }
      }

      // 检查表单验证
      if (result.details.clientForm.formValidation) {
        const validation = result.details.clientForm.formValidation;
        if (Object.keys(validation.errors).length === 0) {
          result.warnings.push('表单验证可能不够严格');
        }
      }

      // 检查各模块之间的关联性
      if (result.details.clientList.valid && result.details.clientDetail.valid) {
        // 客户列表和详情都有效，检查数据一致性
        if (result.details.clientList.stats.totalClients > 0 && !result.details.clientDetail.clientInfo) {
          result.warnings.push('客户列表有数据但详情页面无法加载客户信息');
        }
      }

      // 性能相关检查
      if (result.executionTime > 10000) {
        result.warnings.push('验证执行时间较长，可能存在性能问题');
      }

      this.logger.debug('综合验证完成');

    } catch (error) {
      result.warnings.push(`综合验证失败: ${error instanceof Error ? error.message : String(error)}`);
    }
  }

  /**
   * 获取客户统计信息
   */
  override async getClientStatistics(): Promise<ClientStatistics> {
    try {
      await this.clientListPage.navigateToClientList();
      const stats = await this.clientListPage.getClientStatistics();

      const statistics: ClientStatistics = {
        total: stats.total,
        active: stats.active,
        inactive: stats.inactive,
        byType: stats.byType,
        byIndustry: stats.byIndustry,
        avgCaseCount: 0,
        recentAdditions: 0
      };

      // 计算平均案件数量
      const clients = await this.clientListPage.getClientList();
      if (clients.length > 0) {
        const totalCases = clients.reduce((sum, client) => sum + client.caseCount, 0);
        statistics.avgCaseCount = totalCases / clients.length;
      }

      // 计算最近添加的客户（假设7天内）
      const sevenDaysAgo = new Date();
      sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 7);
      statistics.recentAdditions = clients.filter(client =>
        client.createdDate >= sevenDaysAgo
      ).length;

      return statistics;

    } catch (error) {
      this.logger.error('获取客户统计信息失败', { error });
      throw error;
    }
  }

  /**
   * 生成验证报告
   */
  generateValidationReport(result: ClientManagementValidationResult): string {
    const report = [
      '# 客户管理模块验证报告',
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
      '### 客户列表页面',
      `状态: ${result.details.clientList.valid ? '✅ 通过' : '❌ 失败'}`,
      `错误: ${result.details.clientList.errors.length}`,
      `警告: ${result.details.clientList.warnings.length}`,
      '',
      '### 客户详情页面',
      `状态: ${result.details.clientDetail.valid ? '✅ 通过' : '❌ 失败'}`,
      `错误: ${result.details.clientDetail.errors.length}`,
      `警告: ${result.details.clientDetail.warnings.length}`,
      '',
      '### 客户表单页面',
      `状态: ${result.details.clientForm.valid ? '✅ 通过' : '❌ 失败'}`,
      `错误: ${result.details.clientForm.errors.length}`,
      `警告: ${result.details.clientForm.warnings.length}`,
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
      `客户总数: ${result.details.clientList.stats.totalClients}`,
      `活跃客户: ${result.details.clientList.stats.activeClients}`,
      '',
      '### 按类型分布',
      ...Object.entries(result.details.clientList.stats.byType).map(([type, count]) => `- ${type}: ${count}`),
      '',
      '### 按行业分布',
      ...Object.entries(result.details.clientList.stats.byIndustry).map(([industry, count]) => `- ${industry}: ${count}`)
    ];

    return report.join('\n');
  }

  /**
   * 等待指定时间
   */
  private override async wait(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}