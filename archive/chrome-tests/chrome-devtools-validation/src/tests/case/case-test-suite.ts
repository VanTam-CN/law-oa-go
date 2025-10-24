/**
 * 案件管理测试套件
 */

import { CaseListPage } from '../pages/case/case-list-page';
import { CaseDetailPage } from '../pages/case/case-detail-page';
import { CaseFormPage } from '../pages/case/case-form-page';
import { BasePageObject } from '../core/base-page-object';
import { Logger } from '../core/logger';
import { TestExecutionEngine } from '../core/test-execution-engine';
import { TestCase, TestStep } from '../types/test-types';

export interface CaseTestUser {
  username: string;
  password: string;
  role: string;
  department: string;
  permissions: string[];
}

export interface CaseTestConfig {
  baseUrl: string;
  user: CaseTestUser;
  defaultTimeout?: number;
  screenshotOnFailure?: boolean;
}

export interface CaseData {
  title: string;
  description: string;
  caseType: string;
  priority: 'low' | 'medium' | 'high' | 'urgent';
  client: string;
  assignedAttorney: string;
  assignedParalegal?: string;
  startDate: string;
  expectedEndDate?: string;
  estimatedValue?: number;
  tags: string[];
  jurisdiction: string;
  court?: string;
  judge?: string;
  opposingCounsel?: string[];
  status: 'draft' | 'active' | 'pending' | 'closed' | 'archived';
}

export class CaseTestSuite {
  private caseListPage: CaseListPage;
  private caseDetailPage: CaseDetailPage;
  private caseFormPage: CaseFormPage;
  private testEngine: TestExecutionEngine;
  private config: CaseTestConfig;
  private logger: Logger;

  constructor(config: CaseTestConfig, logger?: Logger) {
    this.config = config;
    this.logger = logger || new Logger('CaseTestSuite');

    const baseConfig = {
      baseUrl: config.baseUrl,
      defaultTimeout: config.defaultTimeout || 30000,
      screenshotOnFailure: config.screenshotOnFailure || true
    };

    this.caseListPage = new CaseListPage(baseConfig, this.logger);
    this.caseDetailPage = new CaseDetailPage(baseConfig, this.logger);
    this.caseFormPage = new CaseFormPage(baseConfig, this.logger);
    this.testEngine = new TestExecutionEngine(baseConfig, this.logger);
  }

  /**
   * 运行完整的案件管理测试套件
   */
  override async runFullCaseTestSuite(): Promise<{
    passed: number;
    failed: number;
    total: number;
    results: any[];
    summary: string;
  }> {
    this.logger.info('开始运行完整的案件管理测试套件');

    const testCases = this.getAllTestCases();
    const results = await this.testEngine.executeTestCases(testCases);

    const passed = results.filter(r => r.status === 'passed').length;
    const failed = results.filter(r => r.status === 'failed').length;
    const total = results.length;

    const summary = `
案件管理测试套件执行完成：
- 总计测试用例: ${total}
- 通过: ${passed}
- 失败: ${failed}
- 成功率: ${((passed / total) * 100).toFixed(2)}%
    `.trim();

    this.logger.info(summary);

    return {
      passed,
      failed,
      total,
      results,
      summary
    };
  }

  /**
   * 获取所有案件管理测试用例
   */
  private getAllTestCases(): TestCase[] {
    return [
      this.createCaseListPageValidationTestCase(),
      this.createCaseSearchTestCase(),
      this.createCaseFilterTestCase(),
      this.createCaseSortTestCase(),
      this.createCreateCaseTestCase(),
      this.createCaseFormValidationTestCase(),
      this.createViewCaseDetailTestCase(),
      this.createEditCaseTestCase(),
      this.createDeleteCaseTestCase(),
      this.createCaseStatusUpdateTestCase(),
      this.createCaseAssignmentTestCase(),
      this.createCaseDocumentsTestCase(),
      this.createCaseTimelineTestCase(),
      this.createCaseStatisticsTestCase(),
      this.createCaseExportTestCase(),
      this.createCasePermissionsTestCase(),
      this.createBulkCaseActionsTestCase(),
      this.createCaseDuplicateTestCase(),
      this.createCaseArchiveTestCase(),
      this.createCaseWorkflowTestCase()
    ];
  }

  /**
   * 创建案件列表页面验证测试用例
   */
  private createCaseListPageValidationTestCase(): TestCase {
    return {
      id: 'CASE-CL-001',
      name: '案件列表页面元素验证',
      description: '验证案件列表页面包含所有必需的元素',
      priority: 'high',
      tags: ['case', 'list', 'ui-validation'],
      steps: [
        {
          id: 'step-1',
          name: '用户登录',
          action: 'login',
          expected: '成功登录系统',
          credentials: {
            username: this.config.user.username,
            password: this.config.user.password
          }
        },
        {
          id: 'step-2',
          name: '导航到案件列表页面',
          action: 'navigate',
          expected: '成功加载案件列表页面',
          url: `${this.config.baseUrl}/cases`
        },
        {
          id: 'step-3',
          name: '验证页面标题',
          action: 'verify-element',
          expected: '页面标题包含"案件管理"',
          selector: '#case-list-title'
        },
        {
          id: 'step-4',
          name: '验证搜索框',
          action: 'verify-element',
          expected: '搜索框存在且可见',
          selector: '#case-search-input'
        },
        {
          id: 'step-5',
          name: '验证过滤器按钮',
          action: 'verify-element',
          expected: '过滤器按钮存在',
          selector: '#case-filter-toggle'
        },
        {
          id: 'step-6',
          name: '验证创建案件按钮',
          action: 'verify-element',
          expected: '创建案件按钮存在且可点击',
          selector: '#case-create-button'
        },
        {
          id: 'step-7',
          name: '验证案件列表容器',
          action: 'verify-element',
          expected: '案件列表容器存在',
          selector: '#case-list-container'
        },
        {
          id: 'step-8',
          name: '验证统计面板',
          action: 'verify-element',
          expected: '案件统计面板存在',
          selector: '#case-stats-container'
        },
        {
          id: 'step-9',
          name: '验证导出按钮',
          action: 'verify-element',
          expected: '导出按钮存在',
          selector: '#case-export-button'
        }
      ]
    };
  }

  /**
   * 创建案件搜索测试用例
   */
  private createCaseSearchTestCase(): TestCase {
    return {
      id: 'CASE-SR-001',
      name: '案件搜索功能',
      description: '测试案件搜索功能的正常工作',
      priority: 'high',
      tags: ['case', 'search', 'feature'],
      testData: {
        searchTerm: '合同纠纷',
        expectedResults: 3
      },
      steps: [
        {
          id: 'step-1',
          name: '用户登录',
          action: 'login',
          expected: '成功登录系统',
          credentials: {
            username: this.config.user.username,
            password: this.config.user.password
          }
        },
        {
          id: 'step-2',
          name: '导航到案件列表页面',
          action: 'navigate',
          expected: '成功加载案件列表页面',
          url: `${this.config.baseUrl}/cases`
        },
        {
          id: 'step-3',
          name: '输入搜索关键词',
          action: 'fill',
          selector: '#case-search-input',
          value: '合同纠纷',
          expected: '搜索关键词输入成功'
        },
        {
          id: 'step-4',
          name: '点击搜索按钮',
          action: 'click',
          selector: '#case-search-button',
          expected: '搜索结果加载完成'
        },
        {
          id: 'step-5',
          name: '验证搜索结果',
          action: 'verify-search-results',
          expected: '显示相关搜索结果',
          selector: '.case-item',
          expectedResultCount: 3
        },
        {
          id: 'step-6',
          name: '验证搜索结果相关性',
          action: 'verify-search-relevance',
          expected: '搜索结果包含搜索关键词',
          searchTerm: '合同纠纷'
        },
        {
          id: 'step-7',
          name: '清空搜索',
          action: 'click',
          selector: '#case-search-clear',
          expected: '搜索框清空，显示所有案件'
        }
      ]
    };
  }

  /**
   * 创建案件过滤器测试用例
   */
  private createCaseFilterTestCase(): TestCase {
    return {
      id: 'CASE-FL-001',
      name: '案件过滤功能',
      description: '测试案件过滤器的正常工作',
      priority: 'high',
      tags: ['case', 'filter', 'feature'],
      testData: {
        filters: {
          status: 'active',
          priority: 'high',
          caseType: 'litigation',
          assignedAttorney: this.config.user.username
        }
      },
      steps: [
        {
          id: 'step-1',
          name: '用户登录',
          action: 'login',
          expected: '成功登录系统',
          credentials: {
            username: this.config.user.username,
            password: this.config.user.password
          }
        },
        {
          id: 'step-2',
          name: '导航到案件列表页面',
          action: 'navigate',
          expected: '成功加载案件列表页面',
          url: `${this.config.baseUrl}/cases`
        },
        {
          id: 'step-3',
          name: '打开过滤器面板',
          action: 'click',
          selector: '#case-filter-toggle',
          expected: '过滤器面板展开'
        },
        {
          id: 'step-4',
          name: '设置状态过滤器',
          action: 'select-option',
          selector: '#case-status-filter',
          value: ['active'],
          expected: '状态过滤器设置成功'
        },
        {
          id: 'step-5',
          name: '设置优先级过滤器',
          action: 'select-option',
          selector: '#case-priority-filter',
          value: ['high'],
          expected: '优先级过滤器设置成功'
        },
        {
          id: 'step-6',
          name: '设置案件类型过滤器',
          action: 'select-option',
          selector: '#case-type-filter',
          value: ['litigation'],
          expected: '案件类型过滤器设置成功'
        },
        {
          id: 'step-7',
          name: '设置负责人过滤器',
          action: 'select-option',
          selector: '#case-attorney-filter',
          value: [this.config.user.username],
          expected: '负责人过滤器设置成功'
        },
        {
          id: 'step-8',
          name: '应用过滤器',
          action: 'click',
          selector: '#case-filter-apply',
          expected: '过滤器应用成功，显示过滤结果'
        },
        {
          id: 'step-9',
          name: '验证过滤结果',
          action: 'verify-filter-results',
          expected: '显示符合过滤条件的案件',
          filters: {
            status: 'active',
            priority: 'high',
            caseType: 'litigation',
            assignedAttorney: this.config.user.username
          }
        },
        {
          id: 'step-10',
          name: '重置过滤器',
          action: 'click',
          selector: '#case-filter-reset',
          expected: '过滤器重置，显示所有案件'
        }
      ]
    };
  }

  /**
   * 创建案件排序测试用例
   */
  private createCaseSortTestCase(): TestCase {
    return {
      id: 'CASE-SO-001',
      name: '案件排序功能',
      description: '测试案件排序功能的正常工作',
      priority: 'medium',
      tags: ['case', 'sort', 'feature'],
      testData: {
        sortOptions: [
          { field: 'createdDate', order: 'desc' },
          { field: 'title', order: 'asc' },
          { field: 'priority', order: 'desc' }
        ]
      },
      steps: [
        {
          id: 'step-1',
          name: '用户登录',
          action: 'login',
          expected: '成功登录系统',
          credentials: {
            username: this.config.user.username,
            password: this.config.user.password
          }
        },
        {
          id: 'step-2',
          name: '导航到案件列表页面',
          action: 'navigate',
          expected: '成功加载案件列表页面',
          url: `${this.config.baseUrl}/cases`
        },
        {
          id: 'step-3',
          name: '按创建日期降序排序',
          action: 'sort-cases',
          expected: '案件按创建日期降序排列',
          sortBy: { field: 'createdDate', order: 'desc' }
        },
        {
          id: 'step-4',
          name: '验证排序结果',
          action: 'verify-sort-order',
          expected: '案件按创建日期降序排列',
          sortField: 'createdDate',
          sortOrder: 'desc'
        },
        {
          id: 'step-5',
          name: '按标题升序排序',
          action: 'sort-cases',
          expected: '案件按标题升序排列',
          sortBy: { field: 'title', order: 'asc' }
        },
        {
          id: 'step-6',
          name: '验证排序结果',
          action: 'verify-sort-order',
          expected: '案件按标题升序排列',
          sortField: 'title',
          sortOrder: 'asc'
        },
        {
          id: 'step-7',
          name: '按优先级降序排序',
          action: 'sort-cases',
          expected: '案件按优先级降序排列',
          sortBy: { field: 'priority', order: 'desc' }
        },
        {
          id: 'step-8',
          name: '验证排序结果',
          action: 'verify-sort-order',
          expected: '案件按优先级降序排列',
          sortField: 'priority',
          sortOrder: 'desc'
        }
      ]
    };
  }

  /**
   * 创建创建案件测试用例
   */
  private createCreateCaseTestCase(): TestCase {
    return {
      id: 'CASE-CR-001',
      name: '创建新案件',
      description: '测试创建新案件的完整流程',
      priority: 'critical',
      tags: ['case', 'create', 'happy-path'],
      testData: {
        caseData: {
          title: '测试案件-合同纠纷',
          description: '这是一个测试用的合同纠纷案件',
          caseType: 'litigation',
          priority: 'medium',
          client: '测试客户公司',
          assignedAttorney: this.config.user.username,
          startDate: new Date().toISOString().split('T')[0],
          expectedEndDate: new Date(Date.now() + 90 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
          estimatedValue: 50000,
          tags: ['测试', '合同纠纷', '重要'],
          jurisdiction: '北京市',
          court: '北京市第一中级人民法院',
          status: 'active'
        }
      },
      steps: [
        {
          id: 'step-1',
          name: '用户登录',
          action: 'login',
          expected: '成功登录系统',
          credentials: {
            username: this.config.user.username,
            password: this.config.user.password
          }
        },
        {
          id: 'step-2',
          name: '导航到案件列表页面',
          action: 'navigate',
          expected: '成功加载案件列表页面',
          url: `${this.config.baseUrl}/cases`
        },
        {
          id: 'step-3',
          name: '点击创建案件按钮',
          action: 'click',
          selector: '#case-create-button',
          expected: '跳转到创建案件页面'
        },
        {
          id: 'step-4',
          name: '验证创建案件页面',
          action: 'verify-element',
          expected: '创建案件页面标题显示',
          selector: '#case-create-title'
        },
        {
          id: 'step-5',
          name: '填写案件基本信息',
          action: 'fill-case-basic-info',
          expected: '基本信息填写成功',
          caseData: {
            title: '测试案件-合同纠纷',
            description: '这是一个测试用的合同纠纷案件',
            caseType: 'litigation',
            priority: 'medium',
            client: '测试客户公司',
            assignedAttorney: this.config.user.username,
            startDate: new Date().toISOString().split('T')[0],
            expectedEndDate: new Date(Date.now() + 90 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
            estimatedValue: 50000,
            tags: ['测试', '合同纠纷', '重要']
          }
        },
        {
          id: 'step-6',
          name: '填写案件详细信息',
          action: 'fill-case-detailed-info',
          expected: '详细信息填写成功',
          caseData: {
            jurisdiction: '北京市',
            court: '北京市第一中级人民法院',
            status: 'active'
          }
        },
        {
          id: 'step-7',
          name: '验证表单数据',
          action: 'validate-case-form',
          expected: '表单数据验证通过',
          expectedFields: ['title', 'description', 'caseType', 'priority', 'client', 'assignedAttorney']
        },
        {
          id: 'step-8',
          name: '提交创建案件',
          action: 'click',
          selector: '#case-create-submit',
          expected: '案件创建成功'
        },
        {
          id: 'step-9',
          name: '验证重定向到案件详情',
          action: 'verify-url',
          expected: '重定向到案件详情页面',
          pattern: '.*cases/.*'
        },
        {
          id: 'step-10',
          name: '验证案件信息显示',
          action: 'verify-case-info',
          expected: '案件信息正确显示',
          expectedTitle: '测试案件-合同纠纷',
          expectedClient: '测试客户公司'
        }
      ]
    };
  }

  /**
   * 创建案件表单验证测试用例
   */
  private createCaseFormValidationTestCase(): TestCase {
    return {
      id: 'CASE-FV-001',
      name: '案件表单验证',
      description: '测试案件创建表单的验证功能',
      priority: 'high',
      tags: ['case', 'validation', 'negative'],
      steps: [
        {
          id: 'step-1',
          name: '用户登录',
          action: 'login',
          expected: '成功登录系统',
          credentials: {
            username: this.config.user.username,
            password: this.config.user.password
          }
        },
        {
          id: 'step-2',
          name: '导航到创建案件页面',
          action: 'navigate',
          expected: '成功加载创建案件页面',
          url: `${this.config.baseUrl}/cases/create`
        },
        {
          id: 'step-3',
          name: '不填写必填字段直接提交',
          action: 'click',
          selector: '#case-create-submit',
          expected: '显示必填字段错误消息'
        },
        {
          id: 'step-4',
          name: '验证标题必填错误',
          action: 'verify-element',
          expected: '显示"案件标题不能为空"错误',
          selector: '#case-title-error'
        },
        {
          id: 'step-5',
          name: '验证客户必填错误',
          action: 'verify-element',
          expected: '显示"请选择客户"错误',
          selector: '#case-client-error'
        },
        {
          id: 'step-6',
          name: '输入无效的预计价值',
          action: 'fill',
          selector: '#case-estimated-value',
          value: 'invalid-value',
          expected: '显示无效数值错误'
        },
        {
          id: 'step-7',
          name: '验证数值格式错误',
          action: 'verify-element',
          expected: '显示"请输入有效的数值"错误',
          selector: '#case-value-error'
        },
        {
          id: 'step-8',
          name: '输入无效的日期格式',
          action: 'fill',
          selector: '#case-start-date',
          value: 'invalid-date',
          expected: '显示日期格式错误'
        },
        {
          id: 'step-9',
          name: '验证日期格式错误',
          action: 'verify-element',
          expected: '显示"请输入有效的日期"错误',
          selector: '#case-date-error'
        }
      ]
    };
  }

  /**
   * 创建查看案件详情测试用例
   */
  private createViewCaseDetailTestCase(): TestCase {
    return {
      id: 'CASE-VD-001',
      name: '查看案件详情',
      description: '测试查看案件详情页面的功能',
      priority: 'high',
      tags: ['case', 'view', 'feature'],
      testData: {
        caseId: '123',
        expectedTitle: '合同纠纷案件',
        expectedClient: '测试客户公司'
      },
      steps: [
        {
          id: 'step-1',
          name: '用户登录',
          action: 'login',
          expected: '成功登录系统',
          credentials: {
            username: this.config.user.username,
            password: this.config.user.password
          }
        },
        {
          id: 'step-2',
          name: '导航到案件详情页面',
          action: 'navigate',
          expected: '成功加载案件详情页面',
          url: `${this.config.baseUrl}/cases/123`
        },
        {
          id: 'step-3',
          name: '验证案件基本信息',
          action: 'verify-case-basic-info',
          expected: '案件基本信息正确显示',
          expectedData: {
            title: '合同纠纷案件',
            client: '测试客户公司',
            status: 'active',
            priority: 'medium'
          }
        },
        {
          id: 'step-4',
          name: '验证案件详情标签页',
          action: 'verify-element',
          expected: '案件详情标签页存在',
          selector: '#case-detail-tabs'
        },
        {
          id: 'step-5',
          name: '验证概览标签页',
          action: 'click',
          selector: '#case-tab-overview',
          expected: '概览标签页内容显示'
        },
        {
          id: 'step-6',
          name: '验证文档标签页',
          action: 'click',
          selector: '#case-tab-documents',
          expected: '文档标签页内容显示'
        },
        {
          id: 'step-7',
          name: '验证时间线标签页',
          action: 'click',
          selector: '#case-tab-timeline',
          expected: '时间线标签页内容显示'
        },
        {
          id: 'step-8',
          name: '验证团队标签页',
          action: 'click',
          selector: '#case-tab-team',
          expected: '团队标签页内容显示'
        },
        {
          id: 'step-9',
          name: '验证统计信息',
          action: 'verify-element',
          expected: '案件统计信息显示',
          selector: '#case-statistics'
        }
      ]
    };
  }

  /**
   * 创建编辑案件测试用例
   */
  private createEditCaseTestCase(): TestCase {
    return {
      id: 'CASE-ED-001',
      name: '编辑案件信息',
      description: '测试编辑案件信息的功能',
      priority: 'high',
      tags: ['case', 'edit', 'feature'],
      testData: {
        caseId: '123',
        updateData: {
          title: '更新后的案件标题',
          description: '更新后的案件描述',
          priority: 'high',
          estimatedValue: 75000,
          tags: ['重要', '合同纠纷']
        }
      },
      steps: [
        {
          id: 'step-1',
          name: '用户登录',
          action: 'login',
          expected: '成功登录系统',
          credentials: {
            username: this.config.user.username,
            password: this.config.user.password
          }
        },
        {
          id: 'step-2',
          name: '导航到案件详情页面',
          action: 'navigate',
          expected: '成功加载案件详情页面',
          url: `${this.config.baseUrl}/cases/123`
        },
        {
          id: 'step-3',
          name: '点击编辑按钮',
          action: 'click',
          selector: '#case-edit-button',
          expected: '跳转到编辑案件页面'
        },
        {
          id: 'step-4',
          name: '验证编辑表单',
          action: 'verify-element',
          expected: '编辑表单加载成功，包含当前案件数据',
          selector: '#case-edit-form'
        },
        {
          id: 'step-5',
          name: '更新案件标题',
          action: 'fill',
          selector: '#case-title',
          value: '更新后的案件标题',
          expected: '标题更新成功'
        },
        {
          id: 'step-6',
          name: '更新案件描述',
          action: 'fill',
          selector: '#case-description',
          value: '更新后的案件描述',
          expected: '描述更新成功'
        },
        {
          id: 'step-7',
          name: '更新优先级',
          action: 'select-option',
          selector: '#case-priority',
          value: ['high'],
          expected: '优先级更新成功'
        },
        {
          id: 'step-8',
          name: '更新预计价值',
          action: 'fill',
          selector: '#case-estimated-value',
          value: '75000',
          expected: '预计价值更新成功'
        },
        {
          id: 'step-9',
          name: '更新标签',
          action: 'fill',
          selector: '#case-tags',
          value: '重要,合同纠纷',
          expected: '标签更新成功'
        },
        {
          id: 'step-10',
          name: '保存更改',
          action: 'click',
          selector: '#case-save-button',
          expected: '案件信息更新成功'
        },
        {
          id: 'step-11',
          name: '验证更新结果',
          action: 'verify-case-updates',
          expected: '案件信息已更新',
          expectedData: {
            title: '更新后的案件标题',
            priority: 'high',
            estimatedValue: 75000
          }
        }
      ]
    };
  }

  /**
   * 创建删除案件测试用例
   */
  private createDeleteCaseTestCase(): TestCase {
    return {
      id: 'CASE-DL-001',
      name: '删除案件',
      description: '测试删除案件的功能',
      priority: 'medium',
      tags: ['case', 'delete', 'feature'],
      testData: {
        caseId: '999', // 假设这是一个测试案件
        caseTitle: '待删除的测试案件'
      },
      steps: [
        {
          id: 'step-1',
          name: '用户登录',
          action: 'login',
          expected: '成功登录系统',
          credentials: {
            username: this.config.user.username,
            password: this.config.user.password
          }
        },
        {
          id: 'step-2',
          name: '导航到案件详情页面',
          action: 'navigate',
          expected: '成功加载案件详情页面',
          url: `${this.config.baseUrl}/cases/999`
        },
        {
          id: 'step-3',
          name: '点击删除按钮',
          action: 'click',
          selector: '#case-delete-button',
          expected: '显示删除确认对话框'
        },
        {
          id: 'step-4',
          name: '验证删除确认对话框',
          action: 'verify-element',
          expected: '删除确认对话框显示',
          selector: '#case-delete-modal'
        },
        {
          id: 'step-5',
          name: '确认删除',
          action: 'click',
          selector: '#case-delete-confirm',
          expected: '案件删除成功'
        },
        {
          id: 'step-6',
          name: '验证重定向到案件列表',
          action: 'verify-url',
          expected: '重定向到案件列表页面',
          pattern: '.*cases$'
        },
        {
          id: 'step-7',
          name: '验证删除成功消息',
          action: 'verify-element',
          expected: '显示案件删除成功消息',
          selector: '#case-delete-success-message'
        },
        {
          id: 'step-8',
          name: '验证案件不在列表中',
          action: 'verify-case-not-exists',
          expected: '已删除的案件不再显示在列表中',
          caseId: '999'
        }
      ]
    };
  }

  /**
   * 创建案件状态更新测试用例
   */
  private createCaseStatusUpdateTestCase(): TestCase {
    return {
      id: 'CASE-SU-001',
      name: '案件状态更新',
      description: '测试更新案件状态的功能',
      priority: 'high',
      tags: ['case', 'status', 'feature'],
      testData: {
        caseId: '123',
        statusTransitions: [
          { from: 'draft', to: 'active' },
          { from: 'active', to: 'pending' },
          { from: 'pending', to: 'closed' },
          { from: 'closed', to: 'archived' }
        ]
      },
      steps: [
        {
          id: 'step-1',
          name: '用户登录',
          action: 'login',
          expected: '成功登录系统',
          credentials: {
            username: this.config.user.username,
            password: this.config.user.password
          }
        },
        {
          id: 'step-2',
          name: '导航到案件详情页面',
          action: 'navigate',
          expected: '成功加载案件详情页面',
          url: `${this.config.baseUrl}/cases/123`
        },
        {
          id: 'step-3',
          name: '更新案件状态为进行中',
          action: 'update-case-status',
          expected: '案件状态更新成功',
          newStatus: 'active',
          comment: '案件正式开始'
        },
        {
          id: 'step-4',
          name: '验证状态更新',
          action: 'verify-case-status',
          expected: '案件状态显示为"进行中"',
          expectedStatus: 'active'
        },
        {
          id: 'step-5',
          name: '更新案件状态为待处理',
          action: 'update-case-status',
          expected: '案件状态更新成功',
          newStatus: 'pending',
          comment: '等待进一步处理'
        },
        {
          id: 'step-6',
          name: '验证状态更新',
          action: 'verify-case-status',
          expected: '案件状态显示为"待处理"',
          expectedStatus: 'pending'
        },
        {
          id: 'step-7',
          name: '更新案件状态为已关闭',
          action: 'update-case-status',
          expected: '案件状态更新成功',
          newStatus: 'closed',
          comment: '案件已成功结案'
        },
        {
          id: 'step-8',
          name: '验证状态更新',
          action: 'verify-case-status',
          expected: '案件状态显示为"已关闭"',
          expectedStatus: 'closed'
        },
        {
          id: 'step-9',
          name: '验证状态历史记录',
          action: 'verify-status-history',
          expected: '状态变更历史记录正确显示',
          expectedTransitions: 3
        }
      ]
    };
  }

  /**
   * 创建案件分配测试用例
   */
  private createCaseAssignmentTestCase(): TestCase {
    return {
      id: 'CASE-AS-001',
      name: '案件分配管理',
      description: '测试案件分配和团队成员管理功能',
      priority: 'high',
      tags: ['case', 'assignment', 'feature'],
      testData: {
        caseId: '123',
        assignments: [
          { role: 'attorney', user: 'attorney2', message: '请协助处理此案件' },
          { role: 'paralegal', user: 'paralegal1', message: '请准备相关文件' }
        ]
      },
      steps: [
        {
          id: 'step-1',
          name: '用户登录',
          action: 'login',
          expected: '成功登录系统',
          credentials: {
            username: this.config.user.username,
            password: this.config.user.password
          }
        },
        {
          id: 'step-2',
          name: '导航到案件详情页面',
          action: 'navigate',
          expected: '成功加载案件详情页面',
          url: `${this.config.baseUrl}/cases/123`
        },
        {
          id: 'step-3',
          name: '切换到团队标签页',
          action: 'click',
          selector: '#case-tab-team',
          expected: '团队标签页显示'
        },
        {
          id: 'step-4',
          name: '添加律师分配',
          action: 'add-case-assignment',
          expected: '律师分配添加成功',
          assignment: {
            role: 'attorney',
            user: 'attorney2',
            message: '请协助处理此案件'
          }
        },
        {
          id: 'step-5',
          name: '验证律师分配',
          action: 'verify-assignment',
          expected: '律师正确分配到案件',
            role: 'attorney',
            user: 'attorney2'
          },
        {
          id: 'step-6',
          name: '添加助理分配',
          action: 'add-case-assignment',
          expected: '助理分配添加成功',
          assignment: {
            role: 'paralegal',
            user: 'paralegal1',
            message: '请准备相关文件'
          }
        },
        {
          id: 'step-7',
          name: '验证助理分配',
          action: 'verify-assignment',
          expected: '助理正确分配到案件',
          role: 'paralegal',
          user: 'paralegal1'
        },
        {
          id: 'step-8',
          name: '发送分配通知',
          action: 'send-assignment-notification',
          expected: '分配通知发送成功',
          notification: {
            recipients: ['attorney2', 'paralegal1'],
            message: '您已被分配到新案件'
          }
        },
        {
          id: 'step-9',
          name: '验证团队成员列表',
          action: 'verify-team-members',
          expected: '团队成员列表显示所有分配的人员',
          expectedMembers: [this.config.user.username, 'attorney2', 'paralegal1']
        }
      ]
    };
  }

  /**
   * 创建案件文档测试用例
   */
  private createCaseDocumentsTestCase(): TestCase {
    return {
      id: 'CASE-DC-001',
      name: '案件文档管理',
      description: '测试案件相关文档的管理功能',
      priority: 'high',
      tags: ['case', 'documents', 'feature'],
      testData: {
        caseId: '123',
        documents: [
          {
            title: '起诉状',
            description: '正式起诉状文件',
            type: 'legal_document',
            category: 'pleading',
            confidentiality: 'internal'
          },
          {
            title: '证据清单',
            description: '案件相关证据清单',
            type: 'evidence_list',
            category: 'evidence',
            confidentiality: 'internal'
          }
        ]
      },
      steps: [
        {
          id: 'step-1',
          name: '用户登录',
          action: 'login',
          expected: '成功登录系统',
          credentials: {
            username: this.config.user.username,
            password: this.config.user.password
          }
        },
        {
          id: 'step-2',
          name: '导航到案件详情页面',
          action: 'navigate',
          expected: '成功加载案件详情页面',
          url: `${this.config.baseUrl}/cases/123`
        },
        {
          id: 'step-3',
          name: '切换到文档标签页',
          action: 'click',
          selector: '#case-tab-documents',
          expected: '文档标签页显示'
        },
        {
          id: 'step-4',
          name: '上传起诉状文档',
          action: 'upload-case-document',
          expected: '文档上传成功',
          document: {
            title: '起诉状',
            description: '正式起诉状文件',
            type: 'legal_document',
            category: 'pleading',
            confidentiality: 'internal'
          }
        },
        {
          id: 'step-5',
          name: '验证文档上传',
          action: 'verify-document-upload',
          expected: '起诉状文档显示在列表中',
          documentTitle: '起诉状'
        },
        {
          id: 'step-6',
          name: '设置文档权限',
          action: 'set-document-permissions',
          expected: '文档权限设置成功',
          permissions: {
            canView: [this.config.user.username, 'attorney2'],
            canEdit: [this.config.user.username],
            canDelete: [this.config.user.username]
          }
        },
        {
          id: 'step-7',
          name: '上传证据清单文档',
          action: 'upload-case-document',
          expected: '文档上传成功',
          document: {
            title: '证据清单',
            description: '案件相关证据清单',
            type: 'evidence_list',
            category: 'evidence',
            confidentiality: 'internal'
          }
        },
        {
          id: 'step-8',
          name: '验证文档列表',
          action: 'verify-document-list',
          expected: '文档列表显示所有上传的文档',
          expectedDocuments: ['起诉状', '证据清单']
        },
        {
          id: 'step-9',
          name: '测试文档下载',
          action: 'download-document',
          expected: '文档下载成功',
          documentTitle: '起诉状'
        },
        {
          id: 'step-10',
          name: '测试文档预览',
          action: 'preview-document',
          expected: '文档预览功能正常',
          documentTitle: '证据清单'
        }
      ]
    };
  }

  /**
   * 创建案件时间线测试用例
   */
  private createCaseTimelineTestCase(): TestCase {
    return {
      id: 'CASE-TL-001',
      name: '案件时间线管理',
      description: '测试案件时间线和事件记录功能',
      priority: 'medium',
      tags: ['case', 'timeline', 'feature'],
      testData: {
        caseId: '123',
        events: [
          {
            title: '案件受理',
            description: '法院正式受理案件',
            date: new Date().toISOString(),
            type: 'milestone',
            priority: 'high'
          },
          {
            title: '第一次庭审',
            description: '与对方进行第一次庭审',
            date: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
            type: 'hearing',
            priority: 'medium'
          }
        ]
      },
      steps: [
        {
          id: 'step-1',
          name: '用户登录',
          action: 'login',
          expected: '成功登录系统',
          credentials: {
            username: this.config.user.username,
            password: this.config.user.password
          }
        },
        {
          id: 'step-2',
          name: '导航到案件详情页面',
          action: 'navigate',
          expected: '成功加载案件详情页面',
          url: `${this.config.baseUrl}/cases/123`
        },
        {
          id: 'step-3',
          name: '切换到时间线标签页',
          action: 'click',
          selector: '#case-tab-timeline',
          expected: '时间线标签页显示'
        },
        {
          id: 'step-4',
          name: '添加案件受理事件',
          action: 'add-timeline-event',
          expected: '时间线事件添加成功',
          event: {
            title: '案件受理',
            description: '法院正式受理案件',
            date: new Date().toISOString(),
            type: 'milestone',
            priority: 'high'
          }
        },
        {
          id: 'step-5',
          name: '验证事件显示',
          action: 'verify-timeline-event',
          expected: '案件受理事件正确显示在时间线上',
          eventTitle: '案件受理'
        },
        {
          id: 'step-6',
          name: '添加第一次庭审事件',
          action: 'add-timeline-event',
          expected: '时间线事件添加成功',
          event: {
            title: '第一次庭审',
            description: '与对方进行第一次庭审',
            date: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
            type: 'hearing',
            priority: 'medium'
          }
        },
        {
          id: 'step-7',
          name: '验证时间线排序',
          action: 'verify-timeline-order',
          expected: '时间线事件按日期正确排序'
        },
        {
          id: 'step-8',
          name: '编辑时间线事件',
          action: 'edit-timeline-event',
          expected: '时间线事件编辑成功',
          eventId: '案件受理',
          updates: {
            description: '法院正式受理并立案'
          }
        },
        {
          id: 'step-9',
          name: '验证事件更新',
          action: 'verify-event-update',
          expected: '事件描述已更新',
          eventId: '案件受理',
            expectedDescription: '法院正式受理并立案'
        },
        {
          id: 'step-10',
          name: '设置事件提醒',
          action: 'set-event-reminder',
          expected: '事件提醒设置成功',
          eventId: '第一次庭审',
          reminder: {
            before: 24 * 60 * 60 * 1000, // 24小时前
            message: '第一次庭审即将开始'
          }
        }
      ]
    };
  }

  /**
   * 创建案件统计测试用例
   */
  private createCaseStatisticsTestCase(): TestCase {
    return {
      id: 'CASE-ST-001',
      name: '案件统计功能',
      description: '测试案件统计和数据可视化功能',
      priority: 'medium',
      tags: ['case', 'statistics', 'feature'],
      steps: [
        {
          id: 'step-1',
          name: '用户登录',
          action: 'login',
          expected: '成功登录系统',
          credentials: {
            username: this.config.user.username,
            password: this.config.user.password
          }
        },
        {
          id: 'step-2',
          name: '导航到案件列表页面',
          action: 'navigate',
          expected: '成功加载案件列表页面',
          url: `${this.config.baseUrl}/cases`
        },
        {
          id: 'step-3',
          name: '查看总体统计',
          action: 'verify-case-statistics',
          expected: '案件总体统计数据正确显示',
          statistics: ['totalCases', 'activeCases', 'closedCases', 'pendingCases']
        },
        {
          id: 'step-4',
          name: '查看按类型统计',
          action: 'verify-statistics-by-type',
          expected: '按案件类型统计显示正确',
          expectedTypes: ['litigation', 'consultation', 'transaction']
        },
        {
          id: 'step-5',
          name: '查看按优先级统计',
          action: 'verify-statistics-by-priority',
          expected: '按优先级统计显示正确',
          expectedPriorities: ['low', 'medium', 'high', 'urgent']
        },
        {
          id: 'step-6',
          name: '查看按状态统计',
          action: 'verify-statistics-by-status',
          expected: '按状态统计显示正确',
          expectedStatuses: ['draft', 'active', 'pending', 'closed', 'archived']
        },
        {
          id: 'step-7',
          name: '查看月度趋势图',
          action: 'verify-monthly-trend-chart',
          expected: '月度案件趋势图显示正确'
        },
        {
          id: 'step-8',
          name: '查看案件分布图',
          action: 'verify-case-distribution-chart',
          expected: '案件类型分布图显示正确'
        },
        {
          id: 'step-9',
          name: '导出统计报告',
          action: 'export-statistics-report',
          expected: '统计报告导出成功',
          format: 'pdf'
        }
      ]
    };
  }

  /**
   * 创建案件导出测试用例
   */
  private createCaseExportTestCase(): TestCase {
    return {
      id: 'CASE-EX-001',
      name: '案件导出功能',
      description: '测试案件数据导出功能',
      priority: 'medium',
      tags: ['case', 'export', 'feature'],
      steps: [
        {
          id: 'step-1',
          name: '用户登录',
          action: 'login',
          expected: '成功登录系统',
          credentials: {
            username: this.config.user.username,
            password: this.config.user.password
          }
        },
        {
          id: 'step-2',
          name: '导航到案件列表页面',
          action: 'navigate',
          expected: '成功加载案件列表页面',
          url: `${this.config.baseUrl}/cases`
        },
        {
          id: 'step-3',
          name: '点击导出按钮',
          action: 'click',
          selector: '#case-export-button',
          expected: '导出选项对话框显示'
        },
        {
          id: 'step-4',
          name: '选择导出格式',
          action: 'select-option',
          selector: '#export-format',
          value: ['excel'],
          expected: 'Excel格式被选中'
        },
        {
          id: 'step-5',
          name: '选择导出范围',
          action: 'select-option',
          selector: '#export-scope',
          value: ['filtered'],
          expected: '过滤结果导出被选中'
        },
        {
          id: 'step-6',
          name: '选择包含字段',
          action: 'select-export-fields',
          expected: '导出字段选择成功',
          fields: ['title', 'client', 'status', 'priority', 'createdDate', 'assignedAttorney']
        },
        {
          id: 'step-7',
          name: '开始导出',
          action: 'click',
          selector: '#export-start-button',
          expected: '导出过程开始'
        },
        {
          id: 'step-8',
          name: '等待导出完成',
          action: 'wait-for-export',
          expected: '导出完成',
          timeout: 30000
        },
        {
          id: 'step-9',
          name: '验证导出文件',
          action: 'verify-export-file',
          expected: '导出文件下载成功',
          format: 'excel'
        },
        {
          id: 'step-10',
          name: '测试PDF导出',
          action: 'export-cases',
          expected: 'PDF格式导出成功',
          format: 'pdf'
        }
      ]
    };
  }

  /**
   * 创建案件权限测试用例
   */
  private createCasePermissionsTestCase(): TestCase {
    return {
      id: 'CASE-PM-001',
      name: '案件权限管理',
      description: '测试案件访问权限控制功能',
      priority: 'high',
      tags: ['case', 'permissions', 'security'],
      testData: {
        caseId: '123',
        users: {
          owner: this.config.user.username,
          attorney: 'attorney2',
          paralegal: 'paralegal1',
          external: 'external.user'
        }
      },
      steps: [
        {
          id: 'step-1',
          name: '用户登录',
          action: 'login',
          expected: '成功登录系统',
          credentials: {
            username: this.config.user.username,
            password: this.config.user.password
          }
        },
        {
          id: 'step-2',
          name: '导航到案件详情页面',
          action: 'navigate',
          expected: '成功加载案件详情页面',
          url: `${this.config.baseUrl}/cases/123`
        },
        {
          id: 'step-3',
          name: '切换到权限设置标签页',
          action: 'click',
          selector: '#case-tab-permissions',
          expected: '权限设置标签页显示'
        },
        {
          id: 'step-4',
          name: '添加律师查看权限',
          action: 'add-user-permission',
          expected: '权限添加成功',
          permission: {
            user: 'attorney2',
            permissions: ['view', 'edit'],
            message: '协助处理此案件'
          }
        },
        {
          id: 'step-5',
          name: '添加助理只读权限',
          action: 'add-user-permission',
          expected: '权限添加成功',
          permission: {
            user: 'paralegal1',
            permissions: ['view'],
            message: '可查看案件信息'
          }
        },
        {
          id: 'step-6',
          name: '验证权限列表',
          action: 'verify-permission-list',
          expected: '权限列表显示所有用户权限',
          expectedUsers: [this.config.user.username, 'attorney2', 'paralegal1']
        },
        {
          id: 'step-7',
          name: '测试权限限制访问',
          action: 'test-restricted-access',
          expected: '无权限用户无法访问案件',
          restrictedUser: 'external.user',
          caseId: '123'
        },
        {
          id: 'step-8',
          name: '更新权限',
          action: 'update-user-permission',
          expected: '权限更新成功',
          permission: {
            user: 'paralegal1',
            newPermissions: ['view', 'edit'],
            reason: '需要编辑案件文档'
          }
        },
        {
          id: 'step-9',
          name: '验证权限更新',
          action: 'verify-permission-update',
          expected: '助理权限已更新',
          user: 'paralegal1',
          expectedPermissions: ['view', 'edit']
        },
        {
          id: 'step-10',
          name: '移除用户权限',
          action: 'remove-user-permission',
          expected: '用户权限移除成功',
          user: 'attorney2',
          reason: '案件已重新分配'
        }
      ]
    };
  }

  /**
   * 创建批量操作测试用例
   */
  private createBulkCaseActionsTestCase(): TestCase {
    return {
      id: 'CASE-BA-001',
      name: '案件批量操作',
      description: '测试案件批量操作功能',
      priority: 'medium',
      tags: ['case', 'bulk', 'feature'],
      testData: {
        caseIds: ['101', '102', '103'],
        bulkAction: 'assign',
        targetUser: 'attorney2'
      },
      steps: [
        {
          id: 'step-1',
          name: '用户登录',
          action: 'login',
          expected: '成功登录系统',
          credentials: {
            username: this.config.user.username,
            password: this.config.user.password
          }
        },
        {
          id: 'step-2',
          name: '导航到案件列表页面',
          action: 'navigate',
          expected: '成功加载案件列表页面',
          url: `${this.config.baseUrl}/cases`
        },
        {
          id: 'step-3',
          name: '选择多个案件',
          action: 'select-cases',
          expected: '多个案件被选中',
          caseIds: ['101', '102', '103']
        },
        {
          id: 'step-4',
          name: '验证批量操作按钮',
          action: 'verify-element',
          expected: '批量操作按钮可用',
          selector: '#bulk-actions-button'
        },
        {
          id: 'step-5',
          name: '打开批量操作菜单',
          action: 'click',
          selector: '#bulk-actions-button',
          expected: '批量操作菜单显示'
        },
        {
          id: 'step-6',
          name: '选择批量分配',
          action: 'click',
          selector: '#bulk-assign',
          expected: '批量分配对话框显示'
        },
        {
          id: 'step-7',
          name: '选择目标律师',
          action: 'select-option',
          selector: '#bulk-assign-target',
          value: ['attorney2'],
          expected: '目标律师选择成功'
        },
        {
          id: 'step-8',
          name: '输入分配消息',
          action: 'fill',
          selector: '#bulk-assign-message',
          value: '请协助处理这些案件',
          expected: '分配消息输入成功'
        },
        {
          id: 'step-9',
          name: '执行批量分配',
          action: 'click',
          selector: '#bulk-assign-confirm',
          expected: '批量分配执行成功'
        },
        {
          id: 'step-10',
          name: '验证批量操作结果',
          action: 'verify-bulk-operation-result',
          expected: '所有选定案件已分配给目标律师',
          operation: 'assign',
          targetUser: 'attorney2',
          affectedCases: 3
        }
      ]
    };
  }

  /**
   * 创建案件复制测试用例
   */
  private createCaseDuplicateTestCase(): TestCase {
    return {
      id: 'CASE-DP-001',
      name: '案件复制功能',
      description: '测试案件复制和模板功能',
      priority: 'medium',
      tags: ['case', 'duplicate', 'feature'],
      testData: {
        sourceCaseId: '123',
        newCaseData: {
          title: '复制的合同纠纷案件',
          client: '新客户公司',
          assignedAttorney: 'attorney2'
        }
      },
      steps: [
        {
          id: 'step-1',
          name: '用户登录',
          action: 'login',
          expected: '成功登录系统',
          credentials: {
            username: this.config.user.username,
            password: this.config.user.password
          }
        },
        {
          id: 'step-2',
          name: '导航到案件详情页面',
          action: 'navigate',
          expected: '成功加载案件详情页面',
          url: `${this.config.baseUrl}/cases/123`
        },
        {
          id: 'step-3',
          name: '点击复制案件按钮',
          action: 'click',
          selector: '#case-duplicate-button',
          expected: '案件复制对话框显示'
        },
        {
          id: 'step-4',
          name: '填写新案件信息',
          action: 'fill-duplicate-form',
          expected: '新案件信息填写成功',
          newCaseData: {
            title: '复制的合同纠纷案件',
            client: '新客户公司',
            assignedAttorney: 'attorney2'
          }
        },
        {
          id: 'step-5',
          name: '选择要复制的内容',
          action: 'select-duplicate-content',
          expected: '复制内容选择成功',
          content: {
            basicInfo: true,
            team: false,
            documents: true,
            timeline: false,
            customFields: true
          }
        },
        {
          id: 'step-6',
          name: '确认复制',
          action: 'click',
          selector: '#duplicate-confirm',
          expected: '案件复制成功'
        },
        {
          id: 'step-7',
          name: '验证新案件创建',
          action: 'verify-duplicated-case',
          expected: '新案件创建成功并包含复制的内容',
          expectedTitle: '复制的合同纠纷案件',
          expectedClient: '新客户公司'
        },
        {
          id: 'step-8',
          name: '验证复制的内容',
          action: 'verify-duplicated-content',
          expected: '复制的内容正确显示在新案件中',
          expectedContent: ['basicInfo', 'documents', 'customFields']
        }
      ]
    };
  }

  /**
   * 创建案件归档测试用例
   */
  private createCaseArchiveTestCase(): TestCase {
    return {
      id: 'CASE-AR-001',
      name: '案件归档功能',
      description: '测试案件归档和恢复功能',
      priority: 'medium',
      tags: ['case', 'archive', 'feature'],
      testData: {
        caseId: '456',
        archiveReason: '案件已成功结案，归档保存'
      },
      steps: [
        {
          id: 'step-1',
          name: '用户登录',
          action: 'login',
          expected: '成功登录系统',
          credentials: {
            username: this.config.user.username,
            password: this.config.user.password
          }
        },
        {
          id: 'step-2',
          name: '导航到案件详情页面',
          action: 'navigate',
          expected: '成功加载案件详情页面',
          url: `${this.config.baseUrl}/cases/456`
        },
        {
          id: 'step-3',
          name: '更新案件状态为已关闭',
          action: 'update-case-status',
          expected: '案件状态更新为已关闭',
          newStatus: 'closed',
          comment: '案件已成功结案'
        },
        {
          id: 'step-4',
          name: '点击归档按钮',
          action: 'click',
          selector: '#case-archive-button',
          expected: '归档确认对话框显示'
        },
        {
          id: 'step-5',
          name: '填写归档原因',
          action: 'fill',
          selector: '#archive-reason',
          value: '案件已成功结案，归档保存',
          expected: '归档原因填写成功'
        },
        {
          id: 'step-6',
          name: '确认归档',
          action: 'click',
          selector: '#archive-confirm',
          expected: '案件归档成功'
        },
        {
          id: 'step-7',
          name: '验证归档状态',
          action: 'verify-case-status',
          expected: '案件状态显示为"已归档"',
          expectedStatus: 'archived'
        },
        {
          id: 'step-8',
          name: '验证归档后访问',
          action: 'verify-archived-case-access',
          expected: '归档案件只能查看，不能编辑',
          caseId: '456'
        },
        {
          id: 'step-9',
          name: '测试从普通列表中移除',
          action: 'verify-archived-not-in-list',
          expected: '归档案件不在普通案件列表中显示',
          caseId: '456'
        },
        {
          id: 'step-10',
          name: '测试恢复归档案件',
          action: 'restore-archived-case',
          expected: '归档案件恢复成功',
          caseId: '456',
          reason: '需要重新处理案件'
        }
      ]
    };
  }

  /**
   * 创建案件工作流测试用例
   */
  private createCaseWorkflowTestCase(): TestCase {
    return {
      id: 'CASE-WF-001',
      name: '案件工作流管理',
      description: '测试案件工作流和审批流程',
      priority: 'high',
      tags: ['case', 'workflow', 'feature'],
      testData: {
        caseId: '123',
        workflowSteps: [
          {
            step: 1,
            name: '案件受理',
            assignee: this.config.user.username,
            status: 'completed'
          },
          {
            step: 2,
            name: '法律研究',
            assignee: 'paralegal1',
            status: 'in_progress'
          },
          {
            step: 3,
            name: '文件准备',
            assignee: 'paralegal1',
            status: 'pending'
          },
          {
            step: 4,
            name: '庭前准备',
            assignee: 'attorney2',
            status: 'pending'
          }
        ]
      },
      steps: [
        {
          id: 'step-1',
          name: '用户登录',
          action: 'login',
          expected: '成功登录系统',
          credentials: {
            username: this.config.user.username,
            password: this.config.user.password
          }
        },
        {
          id: 'step-2',
          name: '导航到案件详情页面',
          action: 'navigate',
          expected: '成功加载案件详情页面',
          url: `${this.config.baseUrl}/cases/123`
        },
        {
          id: 'step-3',
          name: '切换到工作流标签页',
          action: 'click',
          selector: '#case-tab-workflow',
          expected: '工作流标签页显示'
        },
        {
          id: 'step-4',
          name: '验证工作流步骤',
          action: 'verify-workflow-steps',
          expected: '工作流步骤正确显示',
          expectedSteps: 4
        },
        {
          id: 'step-5',
          name: '验证当前步骤',
          action: 'verify-current-step',
          expected: '当前步骤为"法律研究"',
          expectedStep: 2
        },
        {
          id: 'step-6',
          name: '完成当前步骤',
          action: 'complete-workflow-step',
          expected: '工作流步骤完成成功',
          stepId: 2,
          comment: '法律研究已完成'
        },
        {
          id: 'step-7',
          name: '验证步骤状态更新',
          action: 'verify-step-status',
          expected: '步骤状态更新为已完成',
          stepId: 2,
          expectedStatus: 'completed'
        },
        {
          id: 'step-8',
          name: '验证下一步激活',
          action: 'verify-next-step-activated',
          expected: '下一步骤"文件准备"已激活',
          expectedStep: 3
        },
        {
          id: 'step-9',
          name: '分配步骤负责人',
          action: 'assign-step-owner',
          expected: '步骤负责人分配成功',
          stepId: 3,
          assignee: 'paralegal1'
        },
        {
          id: 'step-10',
          name: '设置步骤截止日期',
          action: 'set-step-deadline',
          expected: '步骤截止日期设置成功',
          stepId: 3,
          deadline: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString()
        },
        {
          id: 'step-11',
          name: '添加步骤说明',
          action: 'add-step-instructions',
          expected: '步骤说明添加成功',
          stepId: 3,
          instructions: '请准备所有必要的法律文件和证据材料'
        },
        {
          id: 'step-12',
          name: '验证工作流进度',
          action: 'verify-workflow-progress',
          expected: '工作流进度正确显示',
          expectedProgress: 50 // 2/4 完成
        }
      ]
    };
  }

  /**
   * 运行特定的案件测试
   */
  override async runSpecificTest(testId: string): Promise<any> {
    const testCases = this.getAllTestCases();
    const testCase = testCases.find(tc => tc.id === testId);

    if (!testCase) {
      throw new Error(`测试用例 ${testId} 未找到`);
    }

    this.logger.info(`运行特定测试: ${testCase.name} (${testId})`);
    const results = await this.testEngine.executeTestCases([testCase]);

    return results[0];
  }

  /**
   * 运行特定类别的案件测试
   */
  override async runTestsByCategory(category: string): Promise<{
    passed: number;
    failed: number;
    total: number;
    results: any[];
  }> {
    const testCases = this.getAllTestCases();
    const categoryTests = testCases.filter(tc => tc.tags.includes(category));

    this.logger.info(`运行 ${category} 类别测试，共 ${categoryTests.length} 个测试用例`);
    const results = await this.testEngine.executeTestCases(categoryTests);

    const passed = results.filter(r => r.status === 'passed').length;
    const failed = results.filter(r => r.status === 'failed').length;
    const total = results.length;

    return {
      passed,
      failed,
      total,
      results
    };
  }

  /**
   * 生成测试报告
   */
  override async generateTestReport(results: any[]): Promise<string> {
    const report = {
      timestamp: new Date().toISOString(),
      suite: '案件管理测试',
      summary: {
        total: results.length,
        passed: results.filter(r => r.status === 'passed').length,
        failed: results.filter(r => r.status === 'failed').length,
        skipped: results.filter(r => r.status === 'skipped').length
      },
      results: results.map(result => ({
        id: result.testCase.id,
        name: result.testCase.name,
        status: result.status,
        duration: result.duration,
        error: result.error,
        steps: result.getstepResults?.().map((step: any) => ({
          id: step.step.id,
          name: step.step.name,
          status: step.status,
          error: step.error
        }))
      }))
    };

    return JSON.stringify(report, null, 2);
  }
}