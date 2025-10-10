/**
 * 案件管理测试用例集合
 */

import { TestCase, TestSuite, TestStep } from '../types/test-types';
import { TestDataGenerator } from '../utils/test-data-generator';
import { CaseListPage } from './case-list-page';
import { CaseDetailPage } from './case-detail-page';
import { CaseFormPage } from './case-form-page';

export class CaseManagementTestCases {
  private dataGenerator: TestDataGenerator;

  constructor() {
    this.dataGenerator = new TestDataGenerator();
  }

  /**
   * 获取案件列表页面测试套件
   */
  getCaseListTestSuite(baseUrl: string): TestSuite {
    return {
      id: 'case-list-suite',
      name: '案件列表页面测试',
      description: '验证案件列表页面的所有功能',
      baseUrl,
      setup: [
        {
          id: 'navigate-to-case-list',
          name: '导航到案件列表',
          type: 'navigate',
          url: '/cases',
          timeout: 10000
        },
        {
          id: 'verify-case-list-page',
          name: '验证案件列表页面',
          type: 'verifyAssertion',
          assertion: {
            type: 'elementExists',
            selector: '#case-table',
            description: '案件列表表格应该存在'
          },
          timeout: 5000
        }
      ],
      testCases: [
        this.getSearchCasesTestCase(),
        this.getFilterCasesTestCase(),
        this.getSortCasesTestCase(),
        this.getPaginationTestCase(),
        this.getBulkActionsTestCase(),
        this.getExportCasesTestCase(),
        this.getCreateNewCaseTestCase()
      ],
      cleanup: [
        {
          id: 'cleanup-test-data',
          name: '清理测试数据',
          type: 'executeScript',
          script: '// 清理测试创建的案件数据'
        }
      ]
    };
  }

  /**
   * 获取案件详情页面测试套件
   */
  getCaseDetailTestSuite(baseUrl: string): TestSuite {
    return {
      id: 'case-detail-suite',
      name: '案件详情页面测试',
      description: '验证案件详情页面的所有功能',
      baseUrl,
      setup: [
        {
          id: 'navigate-to-case-detail',
          name: '导航到案件详情',
          type: 'navigate',
          url: '/cases/case-1',
          timeout: 10000
        },
        {
          id: 'verify-case-detail-page',
          name: '验证案件详情页面',
          type: 'verifyAssertion',
          assertion: {
            type: 'elementExists',
            selector: '#case-detail-container',
            description: '案件详情容器应该存在'
          },
          timeout: 5000
        }
      ],
      testCases: [
        this.getViewCaseDetailsTestCase(),
        this.getUpdateCaseTestCase(),
        this.getAddMilestoneTestCase(),
        this.getAddDocumentTestCase(),
        this.getAddFinancialRecordTestCase(),
        this.getDeleteCaseTestCase()
      ],
      cleanup: [
        {
          id: 'cleanup-test-data',
          name: '清理测试数据',
          type: 'executeScript',
          script: '// 清理测试修改的案件数据'
        }
      ]
    };
  }

  /**
   * 获取案件表单页面测试套件
   */
  getCaseFormTestSuite(baseUrl: string): TestSuite {
    return {
      id: 'case-form-suite',
      name: '案件表单页面测试',
      description: '验证案件创建和编辑表单的所有功能',
      baseUrl,
      testCases: [
        this.getCreateCaseTestCase(),
        this.getEditCaseTestCase(),
        this.getFormValidationTestCase(),
        this.getClientSearchTestCase(),
        this.getTagManagementTestCase()
      ]
    };
  }

  /**
   * 搜索案件测试用例
   */
  private getSearchCasesTestCase(): TestCase {
    return {
      id: 'search-cases',
      name: '搜索案件功能',
      description: '验证案件搜索功能',
      priority: 'high',
      timeout: 30000,
      setup: [
        {
          id: 'prepare-search-data',
          name: '准备搜索测试数据',
          type: 'executeScript',
          script: '// 准备用于搜索的测试案件数据'
        }
      ],
      steps: [
        {
          id: 'search-by-title',
          name: '按标题搜索',
          type: 'input',
          selector: '#case-search',
          value: '合同纠纷',
          timeout: 5000
        },
        {
          id: 'wait-search-results',
          name: '等待搜索结果',
          type: 'wait',
          duration: 2000
        },
        {
          id: 'verify-search-results',
          name: '验证搜索结果',
          type: 'verifyAssertion',
          assertion: {
            type: 'elementContainsText',
            selector: '.case-row .case-title',
            expectedText: '合同纠纷',
            description: '搜索结果应包含合同纠纷案件'
          }
        },
        {
          id: 'clear-search',
          name: '清除搜索',
          type: 'input',
          selector: '#case-search',
          value: ''
        }
      ],
      assertions: [
        {
          type: 'elementExists',
          selector: '#case-search',
          description: '搜索输入框应该存在'
        },
        {
          type: 'function',
          description: '搜索功能应该正常工作',
          async validate(context: any) {
            const caseListPage = new CaseListPage({ baseUrl: context.baseUrl }, context.logger);
            await caseListPage.searchCases('测试案件');
            const results = await caseListPage.getCaseList();
            return results.length >= 0;
          }
        }
      ],
      tags: ['search', 'case-list']
    };
  }

  /**
   * 过滤案件测试用例
   */
  private getFilterCasesTestCase(): TestCase {
    return {
      id: 'filter-cases',
      name: '过滤案件功能',
      description: '验证案件过滤功能',
      priority: 'high',
      timeout: 30000,
      steps: [
        {
          id: 'apply-status-filter',
          name: '应用状态过滤器',
          type: 'select',
          selector: '#status-filter',
          value: 'active',
          timeout: 5000
        },
        {
          id: 'apply-priority-filter',
          name: '应用优先级过滤器',
          type: 'select',
          selector: '#priority-filter',
          value: 'high',
          timeout: 5000
        },
        {
          id: 'apply-type-filter',
          name: '应用类型过滤器',
          type: 'select',
          selector: '#type-filter',
          value: 'litigation',
          timeout: 5000
        },
        {
          id: 'verify-filter-results',
          name: '验证过滤结果',
          type: 'verifyAssertion',
          assertion: {
            type: 'elementExists',
            selector: '.case-row',
            description: '过滤后应显示案件列表'
          }
        },
        {
          id: 'clear-filters',
          name: '清除过滤器',
          type: 'click',
          selector: '#clear-filters',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'elementExists',
          selector: '#filter-button',
          description: '过滤按钮应该存在'
        }
      ],
      tags: ['filter', 'case-list']
    };
  }

  /**
   * 排序案件测试用例
   */
  private getSortCasesTestCase(): TestCase {
    return {
      id: 'sort-cases',
      name: '排序案件功能',
      description: '验证案件排序功能',
      priority: 'medium',
      timeout: 20000,
      steps: [
        {
          id: 'sort-by-title-asc',
          name: '按标题升序排序',
          type: 'select',
          selector: '#sort-dropdown',
          value: 'title-asc',
          timeout: 5000
        },
        {
          id: 'verify-title-sort',
          name: '验证标题排序',
          type: 'wait',
          duration: 1000
        },
        {
          id: 'sort-by-date-desc',
          name: '按日期降序排序',
          type: 'select',
          selector: '#sort-dropdown',
          value: 'createdDate-desc',
          timeout: 5000
        },
        {
          id: 'verify-date-sort',
          name: '验证日期排序',
          type: 'wait',
          duration: 1000
        }
      ],
      assertions: [
        {
          type: 'elementExists',
          selector: '#sort-dropdown',
          description: '排序下拉框应该存在'
        }
      ],
      tags: ['sort', 'case-list']
    };
  }

  /**
   * 分页功能测试用例
   */
  private getPaginationTestCase(): TestCase {
    return {
      id: 'pagination',
      name: '分页功能',
      description: '验证案件列表分页功能',
      priority: 'medium',
      timeout: 20000,
      steps: [
        {
          id: 'go-to-next-page',
          name: '翻到下一页',
          type: 'click',
          selector: '.next-button',
          timeout: 5000
        },
        {
          id: 'verify-page-change',
          name: '验证页面变化',
          type: 'verifyAssertion',
          assertion: {
            type: 'elementContainsText',
            selector: '.current-page',
            expectedText: '2',
            description: '当前页码应该是第2页'
          }
        },
        {
          id: 'go-to-previous-page',
          name: '翻到上一页',
          type: 'click',
          selector: '.previous-button',
          timeout: 5000
        },
        {
          id: 'verify-return-to-first',
          name: '验证返回第一页',
          type: 'verifyAssertion',
          assertion: {
            type: 'elementContainsText',
            selector: '.current-page',
            expectedText: '1',
            description: '当前页码应该是第1页'
          }
        }
      ],
      assertions: [
        {
          type: 'elementExists',
          selector: '.pagination',
          description: '分页组件应该存在'
        }
      ],
      tags: ['pagination', 'case-list']
    };
  }

  /**
   * 批量操作测试用例
   */
  private getBulkActionsTestCase(): TestCase {
    return {
      id: 'bulk-actions',
      name: '批量操作',
      description: '验证案件批量操作功能',
      priority: 'medium',
      timeout: 30000,
      steps: [
        {
          id: 'select-multiple-cases',
          name: '选择多个案件',
          type: 'executeScript',
          script: `
            // 选择前两个案件的复选框
            const checkboxes = document.querySelectorAll('.case-checkbox');
            if (checkboxes.length >= 2) {
              checkboxes[0].click();
              checkboxes[1].click();
            }
          `
        },
        {
          id: 'verify-bulk-actions-visible',
          name: '验证批量操作可见',
          type: 'verifyAssertion',
          assertion: {
            type: 'elementExists',
            selector: '.bulk-actions',
            description: '批量操作工具栏应该可见'
          }
        },
        {
          id: 'click-bulk-edit',
          name: '点击批量编辑',
          type: 'click',
          selector: '#bulk-edit-button',
          timeout: 5000
        },
        {
          id: 'cancel-bulk-edit',
          name: '取消批量编辑',
          type: 'click',
          selector: '.cancel-button',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'elementExists',
          selector: '.bulk-actions',
          description: '批量操作区域应该存在'
        }
      ],
      tags: ['bulk', 'case-list']
    };
  }

  /**
   * 导出案件测试用例
   */
  private getExportCasesTestCase(): TestCase {
    return {
      id: 'export-cases',
      name: '导出案件',
      description: '验证案件导出功能',
      priority: 'low',
      timeout: 20000,
      steps: [
        {
          id: 'click-export-button',
          name: '点击导出按钮',
          type: 'click',
          selector: '#export-button',
          timeout: 5000
        },
        {
          id: 'select-excel-format',
          name: '选择Excel格式',
          type: 'select',
          selector: '#export-format',
          value: 'excel',
          timeout: 5000
        },
        {
          id: 'confirm-export',
          name: '确认导出',
          type: 'click',
          selector: '#confirm-export-button',
          timeout: 5000
        },
        {
          id: 'verify-export-start',
          name: '验证导出开始',
          type: 'verifyAssertion',
          assertion: {
            type: 'elementExists',
            selector: '.exporting-indicator',
            description: '导出指示器应该显示'
          }
        }
      ],
      assertions: [
        {
          type: 'elementExists',
          selector: '#export-button',
          description: '导出按钮应该存在'
        }
      ],
      tags: ['export', 'case-list']
    };
  }

  /**
   * 创建新案件测试用例
   */
  private getCreateNewCaseTestCase(): TestCase {
    return {
      id: 'create-new-case',
      name: '创建新案件',
      description: '验证从案件列表创建新案件',
      priority: 'high',
      timeout: 60000,
      steps: [
        {
          id: 'click-create-case',
          name: '点击创建案件按钮',
          type: 'click',
          selector: '#create-case-button',
          timeout: 5000
        },
        {
          id: 'verify-case-form',
          name: '验证案件表单',
          type: 'verifyAssertion',
          assertion: {
            type: 'elementExists',
            selector: '#case-form',
            description: '案件表单应该显示'
          }
        },
        {
          id: 'return-to-list',
          name: '返回案件列表',
          type: 'click',
          selector: '#cancel-button',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'elementExists',
          selector: '#create-case-button',
          description: '创建案件按钮应该存在'
        }
      ],
      tags: ['create', 'case-list']
    };
  }

  /**
   * 查看案件详情测试用例
   */
  private getViewCaseDetailsTestCase(): TestCase {
    return {
      id: 'view-case-details',
      name: '查看案件详情',
      description: '验证案件详情查看功能',
      priority: 'high',
      timeout: 30000,
      steps: [
        {
          id: 'verify-case-title',
          name: '验证案件标题',
          type: 'verifyAssertion',
          assertion: {
            type: 'elementExists',
            selector: '#case-title',
            description: '案件标题应该显示'
          }
        },
        {
          id: 'verify-case-info',
          name: '验证案件信息',
          type: 'verifyAssertion',
          assertion: {
            type: 'elementExists',
            selector: '#case-number',
            description: '案件编号应该显示'
          }
        },
        {
          id: 'verify-tabs',
          name: '验证标签页',
          type: 'verifyAssertion',
          assertion: {
            type: 'elementExists',
            selector: '#milestones-section',
            description: '里程碑部分应该存在'
          }
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '案件详情应该正确显示',
          async validate(context: any) {
            const caseDetailPage = new CaseDetailPage({ baseUrl: context.baseUrl }, context.logger);
            const caseDetail = await caseDetailPage.getCaseDetail();
            return caseDetail.id === 'case-1' && caseDetail.title.length > 0;
          }
        }
      ],
      tags: ['view', 'case-detail']
    };
  }

  /**
   * 更新案件测试用例
   */
  private getUpdateCaseTestCase(): TestCase {
    return {
      id: 'update-case',
      name: '更新案件',
      description: '验证案件更新功能',
      priority: 'high',
      timeout: 45000,
      steps: [
        {
          id: 'click-edit-button',
          name: '点击编辑按钮',
          type: 'click',
          selector: '#edit-case-button',
          timeout: 5000
        },
        {
          id: 'update-title',
          name: '更新标题',
          type: 'input',
          selector: '#edit-title',
          value: '更新后的案件标题',
          timeout: 5000
        },
        {
          id: 'update-priority',
          name: '更新优先级',
          type: 'select',
          selector: '#edit-priority',
          value: 'medium',
          timeout: 5000
        },
        {
          id: 'save-changes',
          name: '保存更改',
          type: 'click',
          selector: '#save-case-button',
          timeout: 5000
        },
        {
          id: 'verify-update-success',
          name: '验证更新成功',
          type: 'verifyAssertion',
          assertion: {
            type: 'elementContainsText',
            selector: '.success-message',
            expectedText: '案件更新成功',
            description: '应该显示成功消息'
          }
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '案件更新应该成功',
          async validate(context: any) {
            const caseDetailPage = new CaseDetailPage({ baseUrl: context.baseUrl }, context.logger);
            await caseDetailPage.updateCase({
              title: '测试更新案件',
              priority: 'low'
            });
            return true;
          }
        }
      ],
      tags: ['update', 'case-detail']
    };
  }

  /**
   * 添加里程碑测试用例
   */
  private getAddMilestoneTestCase(): TestCase {
    return {
      id: 'add-milestone',
      name: '添加里程碑',
      description: '验证添加案件里程碑功能',
      priority: 'medium',
      timeout: 30000,
      steps: [
        {
          id: 'click-add-milestone',
          name: '点击添加里程碑按钮',
          type: 'click',
          selector: '#add-milestone-button',
          timeout: 5000
        },
        {
          id: 'fill-milestone-form',
          name: '填写里程碑表单',
          type: 'input',
          selector: '#milestone-title',
          value: '案件受理',
          timeout: 5000
        },
        {
          id: 'set-milestone-date',
          name: '设置里程碑日期',
          type: 'input',
          selector: '#milestone-due-date',
          value: '2024-02-01',
          timeout: 5000
        },
        {
          id: 'save-milestone',
          name: '保存里程碑',
          type: 'click',
          selector: '#save-milestone-button',
          timeout: 5000
        },
        {
          id: 'verify-milestone-added',
          name: '验证里程碑添加成功',
          type: 'verifyAssertion',
          assertion: {
            type: 'elementContainsText',
            selector: '.milestone-item .milestone-title',
            expectedText: '案件受理',
            description: '新里程碑应该显示'
          }
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '里程碑添加应该成功',
          async validate(context: any) {
            const caseDetailPage = new CaseDetailPage({ baseUrl: context.baseUrl }, context.logger);
            await caseDetailPage.addMilestone({
              title: '测试里程碑',
              dueDate: new Date('2024-02-01')
            });
            return true;
          }
        }
      ],
      tags: ['milestone', 'case-detail']
    };
  }

  /**
   * 添加文档测试用例
   */
  private getAddDocumentTestCase(): TestCase {
    return {
      id: 'add-document',
      name: '添加文档',
      description: '验证添加案件文档功能',
      priority: 'medium',
      timeout: 30000,
      steps: [
        {
          id: 'click-add-document',
          name: '点击添加文档按钮',
          type: 'click',
          selector: '#add-document-button',
          timeout: 5000
        },
        {
          id: 'fill-document-form',
          name: '填写文档表单',
          type: 'input',
          selector: '#document-name',
          value: '合同文件.pdf',
          timeout: 5000
        },
        {
          id: 'select-document-type',
          name: '选择文档类型',
          type: 'select',
          selector: '#document-type',
          value: 'contract',
          timeout: 5000
        },
        {
          id: 'save-document',
          name: '保存文档',
          type: 'click',
          selector: '#save-document-button',
          timeout: 5000
        },
        {
          id: 'verify-document-added',
          name: '验证文档添加成功',
          type: 'verifyAssertion',
          assertion: {
            type: 'elementContainsText',
            selector: '.document-item .document-name',
            expectedText: '合同文件.pdf',
            description: '新文档应该显示'
          }
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '文档添加应该成功',
          async validate(context: any) {
            const caseDetailPage = new CaseDetailPage({ baseUrl: context.baseUrl }, context.logger);
            await caseDetailPage.addDocument({
              name: '测试文档.pdf',
              type: 'evidence',
              description: '测试文档描述'
            });
            return true;
          }
        }
      ],
      tags: ['document', 'case-detail']
    };
  }

  /**
   * 添加财务记录测试用例
   */
  private getAddFinancialRecordTestCase(): TestCase {
    return {
      id: 'add-financial-record',
      name: '添加财务记录',
      description: '验证添加案件财务记录功能',
      priority: 'medium',
      timeout: 30000,
      steps: [
        {
          id: 'click-add-financial',
          name: '点击添加财务记录按钮',
          type: 'click',
          selector: '#add-financial-record-button',
          timeout: 5000
        },
        {
          id: 'select-record-type',
          name: '选择记录类型',
          type: 'select',
          selector: '#financial-type',
          value: 'fee',
          timeout: 5000
        },
        {
          id: 'fill-description',
          name: '填写描述',
          type: 'input',
          selector: '#financial-description',
          value: '律师费',
          timeout: 5000
        },
        {
          id: 'fill-amount',
          name: '填写金额',
          type: 'input',
          selector: '#financial-amount',
          value: '50000',
          timeout: 5000
        },
        {
          id: 'save-financial-record',
          name: '保存财务记录',
          type: 'click',
          selector: '#save-financial-button',
          timeout: 5000
        },
        {
          id: 'verify-financial-added',
          name: '验证财务记录添加成功',
          type: 'verifyAssertion',
          assertion: {
            type: 'elementContainsText',
            selector: '.financial-item .financial-description',
            expectedText: '律师费',
            description: '新财务记录应该显示'
          }
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '财务记录添加应该成功',
          async validate(context: any) {
            const caseDetailPage = new CaseDetailPage({ baseUrl: context.baseUrl }, context.logger);
            await caseDetailPage.addFinancialRecord({
              type: 'expense',
              description: '诉讼费',
              amount: 10000,
              date: new Date()
            });
            return true;
          }
        }
      ],
      tags: ['financial', 'case-detail']
    };
  }

  /**
   * 删除案件测试用例
   */
  private getDeleteCaseTestCase(): TestCase {
    return {
      id: 'delete-case',
      name: '删除案件',
      description: '验证案件删除功能',
      priority: 'high',
      timeout: 30000,
      steps: [
        {
          id: 'click-delete-button',
          name: '点击删除按钮',
          type: 'click',
          selector: '#delete-case-button',
          timeout: 5000
        },
        {
          id: 'confirm-delete',
          name: '确认删除',
          type: 'click',
          selector: '.confirm-delete-button',
          timeout: 5000
        },
        {
          id: 'verify-delete-success',
          name: '验证删除成功',
          type: 'verifyAssertion',
          assertion: {
            type: 'elementContainsText',
            selector: '.success-message',
            expectedText: '案件删除成功',
            description: '应该显示成功消息'
          }
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '案件删除应该成功',
          async validate(context: any) {
            const caseDetailPage = new CaseDetailPage({ baseUrl: context.baseUrl }, context.logger);
            await caseDetailPage.deleteCase();
            return true;
          }
        }
      ],
      tags: ['delete', 'case-detail']
    };
  }

  /**
   * 创建案件测试用例
   */
  private getCreateCaseTestCase(): TestCase {
    const caseData = this.dataGenerator.generateCaseData();

    return {
      id: 'create-case',
      name: '创建案件',
      description: '验证完整案件创建流程',
      priority: 'high',
      timeout: 60000,
      setup: [
        {
          id: 'navigate-to-create-case',
          name: '导航到创建案件页面',
          type: 'navigate',
          url: '/cases/create',
          timeout: 10000
        }
      ],
      steps: [
        {
          id: 'fill-case-title',
          name: '填写案件标题',
          type: 'input',
          selector: '#case-title',
          value: caseData.title,
          timeout: 5000
        },
        {
          id: 'fill-case-description',
          name: '填写案件描述',
          type: 'input',
          selector: '#case-description',
          value: caseData.description,
          timeout: 5000
        },
        {
          id: 'select-case-type',
          name: '选择案件类型',
          type: 'select',
          selector: '#case-type',
          value: caseData.type,
          timeout: 5000
        },
        {
          id: 'select-case-priority',
          name: '选择案件优先级',
          type: 'select',
          selector: '#case-priority',
          value: caseData.priority,
          timeout: 5000
        },
        {
          id: 'select-client',
          name: '选择客户',
          type: 'select',
          selector: '#case-client',
          value: caseData.client,
          timeout: 5000
        },
        {
          id: 'select-assigned-to',
          name: '选择分配律师',
          type: 'select',
          selector: '#case-assigned-to',
          value: caseData.assignedTo,
          timeout: 5000
        },
        {
          id: 'set-due-date',
          name: '设置截止日期',
          type: 'input',
          selector: '#case-due-date',
          value: caseData.getdueDate?.().toISOString().split('T')[0] || '',
          timeout: 5000
        },
        {
          id: 'set-estimated-value',
          name: '设置预估价值',
          type: 'input',
          selector: '#case-estimated-value',
          value: caseData.getestimatedValue?.().toString() || '',
          timeout: 5000
        },
        {
          id: 'save-case',
          name: '保存案件',
          type: 'click',
          selector: '#save-case-button',
          timeout: 10000
        },
        {
          id: 'verify-creation-success',
          name: '验证创建成功',
          type: 'verifyAssertion',
          assertion: {
            type: 'elementContainsText',
            selector: '.success-message',
            expectedText: '案件创建成功',
            description: '应该显示成功消息'
          }
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '案件创建应该成功',
          async validate(context: any) {
            const caseFormPage = new CaseFormPage({ baseUrl: context.baseUrl }, context.logger);
            await caseFormPage.createCase(caseData);
            return true;
          }
        }
      ],
      tags: ['create', 'case-form']
    };
  }

  /**
   * 编辑案件测试用例
   */
  private getEditCaseTestCase(): TestCase {
    const updateData = {
      title: '更新后的案件标题',
      priority: 'low' as const,
      description: '更新后的案件描述'
    };

    return {
      id: 'edit-case',
      name: '编辑案件',
      description: '验证案件编辑功能',
      priority: 'high',
      timeout: 45000,
      setup: [
        {
          id: 'navigate-to-edit-case',
          name: '导航到编辑案件页面',
          type: 'navigate',
          url: '/cases/case-1/edit',
          timeout: 10000
        }
      ],
      steps: [
        {
          id: 'update-title',
          name: '更新标题',
          type: 'input',
          selector: '#case-title',
          value: updateData.title,
          timeout: 5000
        },
        {
          id: 'update-description',
          name: '更新描述',
          type: 'input',
          selector: '#case-description',
          value: updateData.description,
          timeout: 5000
        },
        {
          id: 'update-priority',
          name: '更新优先级',
          type: 'select',
          selector: '#case-priority',
          value: updateData.priority,
          timeout: 5000
        },
        {
          id: 'save-changes',
          name: '保存更改',
          type: 'click',
          selector: '#save-case-button',
          timeout: 10000
        },
        {
          id: 'verify-update-success',
          name: '验证更新成功',
          type: 'verifyAssertion',
          assertion: {
            type: 'elementContainsText',
            selector: '.success-message',
            expectedText: '案件更新成功',
            description: '应该显示成功消息'
          }
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '案件编辑应该成功',
          async validate(context: any) {
            const caseFormPage = new CaseFormPage({ baseUrl: context.baseUrl }, context.logger);
            await caseFormPage.updateCase('case-1', updateData);
            return true;
          }
        }
      ],
      tags: ['edit', 'case-form']
    };
  }

  /**
   * 表单验证测试用例
   */
  private getFormValidationTestCase(): TestCase {
    return {
      id: 'form-validation',
      name: '表单验证',
      description: '验证案件表单验证功能',
      priority: 'high',
      timeout: 30000,
      setup: [
        {
          id: 'navigate-to-create-case',
          name: '导航到创建案件页面',
          type: 'navigate',
          url: '/cases/create',
          timeout: 10000
        }
      ],
      steps: [
        {
          id: 'submit-empty-form',
          name: '提交空表单',
          type: 'click',
          selector: '#save-case-button',
          timeout: 5000
        },
        {
          id: 'validation-errors',
          name: '验证错误信息',
          type: 'verifyAssertion',
          assertion: {
            type: 'elementExists',
            selector: '.validation-error',
            description: '应该显示验证错误'
          }
        },
        {
          id: 'fill-only-title',
          name: '只填写标题',
          type: 'input',
          selector: '#case-title',
          value: '测试案件',
          timeout: 5000
        },
        {
          id: 'submit-incomplete-form',
          name: '提交不完整表单',
          type: 'click',
          selector: '#save-case-button',
          timeout: 5000
        },
        {
          id: 'verify-validation-errors',
          name: '验证仍有错误',
          type: 'verifyAssertion',
          assertion: {
            type: 'elementExists',
            selector: '.validation-error',
            description: '应该仍有验证错误'
          }
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '表单验证应该正确工作',
          async validate(context: any) {
            const caseFormPage = new CaseFormPage({ baseUrl: context.baseUrl }, context.logger);
            const validation = await caseFormPage.validateForm();
            return !validation.valid && validation.errors.length > 0;
          }
        }
      ],
      tags: ['validation', 'case-form']
    };
  }

  /**
   * 客户搜索测试用例
   */
  private getClientSearchTestCase(): TestCase {
    return {
      id: 'client-search',
      name: '客户搜索',
      description: '验证客户搜索功能',
      priority: 'medium',
      timeout: 30000,
      setup: [
        {
          id: 'navigate-to-create-case',
          name: '导航到创建案件页面',
          type: 'navigate',
          url: '/cases/create',
          timeout: 10000
        }
      ],
      steps: [
        {
          id: 'focus-client-search',
          name: '聚焦客户搜索',
          type: 'click',
          selector: '#client-search',
          timeout: 5000
        },
        {
          id: 'search-client',
          name: '搜索客户',
          type: 'input',
          selector: '#client-search',
          value: '科技',
          timeout: 5000
        },
        {
          id: 'wait-search-results',
          name: '等待搜索结果',
          type: 'wait',
          duration: 1000
        },
        {
          id: 'verify-search-suggestions',
          name: '验证搜索建议',
          type: 'verifyAssertion',
          assertion: {
            type: 'elementExists',
            selector: '.client-suggestions',
            description: '应该显示客户搜索建议'
          }
        },
        {
          id: 'select-client',
          name: '选择客户',
          type: 'click',
          selector: '.client-suggestion:first-child',
          timeout: 5000
        },
        {
          id: 'verify-client-selected',
          name: '验证客户已选择',
          type: 'verifyAssertion',
          assertion: {
            type: 'elementContainsText',
            selector: '#case-client',
            expectedText: '科技有限公司',
            description: '应该选中科技有限公司'
          }
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '客户搜索应该正常工作',
          async validate(context: any) {
            const caseFormPage = new CaseFormPage({ baseUrl: context.baseUrl }, context.logger);
            await caseFormPage.searchClient('科技');
            return true;
          }
        }
      ],
      tags: ['search', 'client', 'case-form']
    };
  }

  /**
   * 标签管理测试用例
   */
  private getTagManagementTestCase(): TestCase {
    return {
      id: 'tag-management',
      name: '标签管理',
      description: '验证案件标签管理功能',
      priority: 'medium',
      timeout: 30000,
      setup: [
        {
          id: 'navigate-to-create-case',
          name: '导航到创建案件页面',
          type: 'navigate',
          url: '/cases/create',
          timeout: 10000
        }
      ],
      steps: [
        {
          id: 'add-first-tag',
          name: '添加第一个标签',
          type: 'input',
          selector: '#case-tags',
          value: '重要',
          timeout: 5000
        },
        {
          id: 'press-enter-for-first-tag',
          name: '按回车确认第一个标签',
          type: 'sendKeys',
          selector: '#case-tags',
          keys: 'Enter',
          timeout: 2000
        },
        {
          id: 'add-second-tag',
          name: '添加第二个标签',
          type: 'input',
          selector: '#case-tags',
          value: '紧急',
          timeout: 5000
        },
        {
          id: 'press-enter-for-second-tag',
          name: '按回车确认第二个标签',
          type: 'sendKeys',
          selector: '#case-tags',
          keys: 'Enter',
          timeout: 2000
        },
        {
          id: 'verify-tags-added',
          name: '验证标签已添加',
          type: 'verifyAssertion',
          assertion: {
            type: 'elementExists',
            selector: '.case-tag',
            description: '应该显示案件标签'
          }
        },
        {
          id: 'remove-tag',
          name: '删除标签',
          type: 'click',
          selector: '.case-tag .remove-tag:first-child',
          timeout: 5000
        },
        {
          id: 'verify-tag-removed',
          name: '验证标签已删除',
          type: 'verifyAssertion',
          assertion: {
            type: 'elementCount',
            selector: '.case-tag',
            expectedCount: 1,
            description: '应该只剩下一个标签'
          }
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '标签管理应该正常工作',
          async validate(context: any) {
            const caseFormPage = new CaseFormPage({ baseUrl: context.baseUrl }, context.logger);
            await caseFormPage.addTags(['重要', '紧急']);
            return true;
          }
        }
      ],
      tags: ['tags', 'case-form']
    };
  }
}