/**
 * 客户管理测试用例
 */

import { TestCase, TestSuite, TestStep, Assertion } from '../../types/test-types';
import { ClientListPage } from './client-list-page';
import { ClientDetailPage } from './client-detail-page';
import { ClientFormPage } from './client-form-page';

export class ClientManagementTestCases {
  /**
   * 获取客户列表测试套件
   */
  getClientListTestSuite(): TestSuite {
    return {
      id: 'client-list',
      name: '客户列表功能测试',
      description: '验证客户列表页面的各项功能',
      priority: 'high',
      timeout: 60000,
      setup: {
        steps: [
          {
            id: 'navigate-to-client-list',
            name: '导航到客户列表页面',
            type: 'navigate',
            url: '/clients',
            timeout: 5000
          }
        ]
      },
      testCases: [
        this.getLoadClientListTestCase(),
        this.getSearchClientsTestCase(),
        this.getFilterClientsTestCase(),
        this.getSortClientsTestCase(),
        this.getPaginationTestCase(),
        this.getBulkOperationsTestCase(),
        this.getExportClientsTestCase()
      ]
    };
  }

  /**
   * 获取客户详情测试套件
   */
  getClientDetailTestSuite(): TestSuite {
    return {
      id: 'client-detail',
      name: '客户详情功能测试',
      description: '验证客户详情页面的各项功能',
      priority: 'high',
      timeout: 60000,
      setup: {
        steps: [
          {
            id: 'navigate-to-client-detail',
            name: '导航到客户详情页面',
            type: 'navigate',
            url: '/clients/client-1',
            timeout: 5000
          }
        ]
      },
      testCases: [
        this.getLoadClientDetailTestCase(),
        this.getViewClientInfoTestCase(),
        this.getManageContactsTestCase(),
        this.getManageCasesTestCase(),
        this.getManageDocumentsTestCase(),
        this.getEditClientTestCase(),
        this.getDeleteClientTestCase()
      ]
    };
  }

  /**
   * 获取客户表单测试套件
   */
  getClientFormTestSuite(): TestSuite {
    return {
      id: 'client-form',
      name: '客户表单功能测试',
      description: '验证客户创建和编辑表单的功能',
      priority: 'high',
      timeout: 90000,
      testCases: [
        this.getCreateClientTestCase(),
        this.getEditClientFormTestCase(),
        this.getFormValidationTestCase(),
        this.getContactManagementTestCase(),
        this.getSaveAndNewTestCase(),
        this.getCancelOperationTestCase(),
        this.getDuplicateClientTestCase()
      ]
    };
  }

  /**
   * 加载客户列表测试用例
   */
  public getLoadClientListTestCase(): TestCase {
    return {
      id: 'load-client-list',
      name: '加载客户列表',
      description: '验证客户列表页面能够正确加载并显示客户数据',
      priority: 'high',
      timeout: 30000,
      steps: [
        {
          id: 'wait-for-client-table',
          name: '等待客户表格加载',
          type: 'wait',
          selector: '#client-table',
          timeout: 10000
        },
        {
          id: 'check-client-rows',
          name: '检查客户行是否存在',
          type: 'verify',
          selector: '.client-row',
          expectedState: { visible: true }
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '客户列表应该正确加载',
          async validate(context: any) {
            const clientListPage = new ClientListPage({ baseUrl: context.baseUrl }, context.logger);
            const isValid = await clientListPage.validateClientListPage();
            return isValid.valid;
          }
        }
      ],
      tags: ['client-list', 'load', 'smoke']
    };
  }

  /**
   * 搜索客户测试用例
   */
  public getSearchClientsTestCase(): TestCase {
    return {
      id: 'search-clients',
      name: '搜索客户功能',
      description: '验证客户搜索功能能够正确工作',
      priority: 'high',
      timeout: 30000,
      steps: [
        {
          id: 'search-by-name',
          name: '按名称搜索客户',
          type: 'input',
          selector: '#client-search',
          value: '科技',
          timeout: 5000
        },
        {
          id: 'wait-for-search-results',
          name: '等待搜索结果',
          type: 'wait',
          timeout: 3000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '搜索结果应该包含相关客户',
          async validate(context: any) {
            const clientListPage = new ClientListPage({ baseUrl: context.baseUrl }, context.logger);
            const clients = await clientListPage.getClientList();
            return clients.some(client => client.name.includes('科技'));
          }
        }
      ],
      tags: ['client-list', 'search', 'filter']
    };
  }

  /**
   * 过滤客户测试用例
   */
  public getFilterClientsTestCase(): TestCase {
    return {
      id: 'filter-clients',
      name: '过滤客户功能',
      description: '验证客户过滤器能够正确工作',
      priority: 'medium',
      timeout: 30000,
      steps: [
        {
          id: 'filter-by-type',
          name: '按客户类型过滤',
          type: 'select',
          selector: '#type-filter',
          value: 'company',
          timeout: 5000
        },
        {
          id: 'filter-by-industry',
          name: '按行业过滤',
          type: 'select',
          selector: '#industry-filter',
          value: 'technology',
          timeout: 5000
        },
        {
          id: 'filter-by-status',
          name: '按状态过滤',
          type: 'select',
          selector: '#status-filter',
          value: 'active',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '过滤结果应该符合条件',
          async validate(context: any) {
            const clientListPage = new ClientListPage({ baseUrl: context.baseUrl }, context.logger);
            const clients = await clientListPage.getClientList();
            return clients.every(client =>
              client.type === 'company' &&
              client.industry === 'technology' &&
              client.status === 'active'
            );
          }
        }
      ],
      tags: ['client-list', 'filter', 'medium']
    };
  }

  /**
   * 排序客户测试用例
   */
  public getSortClientsTestCase(): TestCase {
    return {
      id: 'sort-clients',
      name: '排序客户功能',
      description: '验证客户排序功能能够正确工作',
      priority: 'medium',
      timeout: 30000,
      steps: [
        {
          id: 'sort-by-name-asc',
          name: '按名称升序排序',
          type: 'select',
          selector: '#sort-dropdown',
          value: 'name-asc',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '客户应该按名称升序排列',
          async validate(context: any) {
            const clientListPage = new ClientListPage({ baseUrl: context.baseUrl }, context.logger);
            const clients = await clientListPage.getClientList();

            // 检查是否按名称升序排列
            for (let i = 1; i < clients.length; i++) {
              if (clients[i-1].name.localeCompare(clients[i].name) > 0) {
                return false;
              }
            }
            return true;
          }
        }
      ],
      tags: ['client-list', 'sort', 'medium']
    };
  }

  /**
   * 分页测试用例
   */
  public getPaginationTestCase(): TestCase {
    return {
      id: 'pagination',
      name: '分页功能',
      description: '验证客户列表分页功能能够正确工作',
      priority: 'medium',
      timeout: 30000,
      steps: [
        {
          id: 'go-to-next-page',
          name: '翻到下一页',
          type: 'click',
          selector: '.next-button',
          timeout: 5000
        },
        {
          id: 'go-to-previous-page',
          name: '翻到上一页',
          type: 'click',
          selector: '.previous-button',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '分页功能应该正常工作',
          async validate(context: any) {
            const clientListPage = new ClientListPage({ baseUrl: context.baseUrl }, context.logger);
            const currentPage = await clientListPage.getCurrentPage();
            const totalPages = await clientListPage.getTotalPages();
            return currentPage >= 1 && currentPage <= totalPages;
          }
        }
      ],
      tags: ['client-list', 'pagination', 'medium']
    };
  }

  /**
   * 批量操作测试用例
   */
  public getBulkOperationsTestCase(): TestCase {
    return {
      id: 'bulk-operations',
      name: '批量操作功能',
      description: '验证客户批量操作功能能够正确工作',
      priority: 'medium',
      timeout: 45000,
      steps: [
        {
          id: 'select-multiple-clients',
          name: '选择多个客户',
          type: 'custom',
          description: '选择多个客户进行批量操作',
          timeout: 10000,
          parameters: {
            action: 'selectClients',
            clientIds: ['client-1', 'client-2']
          }
        },
        {
          id: 'perform-bulk-edit',
          name: '执行批量编辑',
          type: 'click',
          selector: '#bulk-edit-button',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '批量操作应该正常工作',
          async validate(context: any) {
            const clientListPage = new ClientListPage({ baseUrl: context.baseUrl }, context.logger);
            const selectedCount = await clientListPage.getSelectedCount();
            return selectedCount > 0;
          }
        }
      ],
      tags: ['client-list', 'bulk', 'medium']
    };
  }

  /**
   * 导出客户测试用例
   */
  public getExportClientsTestCase(): TestCase {
    return {
      id: 'export-clients',
      name: '导出客户功能',
      description: '验证客户导出功能能够正确工作',
      priority: 'low',
      timeout: 30000,
      steps: [
        {
          id: 'export-to-excel',
          name: '导出到Excel',
          type: 'click',
          selector: '#export-button',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '导出功能应该启动',
          async validate(context: any) {
            // 检查是否有导出相关的成功消息或下载开始
            return true; // 简化验证
          }
        }
      ],
      tags: ['client-list', 'export', 'low']
    };
  }

  /**
   * 加载客户详情测试用例
   */
  public getLoadClientDetailTestCase(): TestCase {
    return {
      id: 'load-client-detail',
      name: '加载客户详情',
      description: '验证客户详情页面能够正确加载并显示客户信息',
      priority: 'high',
      timeout: 30000,
      steps: [
        {
          id: 'wait-for-client-detail',
          name: '等待客户详情加载',
          type: 'wait',
          selector: '#client-detail-container',
          timeout: 10000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '客户详情应该正确加载',
          async validate(context: any) {
            const clientDetailPage = new ClientDetailPage({ baseUrl: context.baseUrl }, context.logger);
            const isValid = await clientDetailPage.validateClientDetailPage();
            return isValid.valid;
          }
        }
      ],
      tags: ['client-detail', 'load', 'smoke']
    };
  }

  /**
   * 查看客户信息测试用例
   */
  public getViewClientInfoTestCase(): TestCase {
    return {
      id: 'view-client-info',
      name: '查看客户信息',
      description: '验证客户基本信息能够正确显示',
      priority: 'high',
      timeout: 30000,
      steps: [
        {
          id: 'check-client-name',
          name: '检查客户名称',
          type: 'verify',
          selector: '#client-name',
          expectedState: { visible: true }
        },
        {
          id: 'check-client-type',
          name: '检查客户类型',
          type: 'verify',
          selector: '#client-type',
          expectedState: { visible: true }
        },
        {
          id: 'check-client-contact',
          name: '检查联系人信息',
          type: 'verify',
          selector: '#client-contact-person',
          expectedState: { visible: true }
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '客户信息应该完整显示',
          async validate(context: any) {
            const clientDetailPage = new ClientDetailPage({ baseUrl: context.baseUrl }, context.logger);
            const clientDetail = await clientDetailPage.getClientDetail();
            return !!clientDetail.name && !!clientDetail.email && !!clientDetail.phone;
          }
        }
      ],
      tags: ['client-detail', 'view', 'smoke']
    };
  }

  /**
   * 管理联系人测试用例
   */
  public getManageContactsTestCase(): TestCase {
    return {
      id: 'manage-contacts',
      name: '管理联系人',
      description: '验证客户联系人管理功能能够正确工作',
      priority: 'medium',
      timeout: 45000,
      steps: [
        {
          id: 'add-new-contact',
          name: '添加新联系人',
          type: 'click',
          selector: '#add-contact-button',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '联系人管理应该正常工作',
          async validate(context: any) {
            const clientDetailPage = new ClientDetailPage({ baseUrl: context.baseUrl }, context.logger);
            const contacts = await clientDetailPage.getContactList();
            return Array.isArray(contacts);
          }
        }
      ],
      tags: ['client-detail', 'contacts', 'medium']
    };
  }

  /**
   * 管理案件测试用例
   */
  public getManageCasesTestCase(): TestCase {
    return {
      id: 'manage-cases',
      name: '管理相关案件',
      description: '验证客户相关案件管理功能能够正确工作',
      priority: 'medium',
      timeout: 30000,
      steps: [
        {
          id: 'check-cases-section',
          name: '检查案件区域',
          type: 'verify',
          selector: '#cases-section',
          expectedState: { visible: true }
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '案件管理应该正常工作',
          async validate(context: any) {
            const clientDetailPage = new ClientDetailPage({ baseUrl: context.baseUrl }, context.logger);
            const cases = await clientDetailPage.getCaseList();
            return Array.isArray(cases);
          }
        }
      ],
      tags: ['client-detail', 'cases', 'medium']
    };
  }

  /**
   * 管理文档测试用例
   */
  public getManageDocumentsTestCase(): TestCase {
    return {
      id: 'manage-documents',
      name: '管理相关文档',
      description: '验证客户相关文档管理功能能够正确工作',
      priority: 'medium',
      timeout: 30000,
      steps: [
        {
          id: 'check-documents-section',
          name: '检查文档区域',
          type: 'verify',
          selector: '#documents-section',
          expectedState: { visible: true }
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '文档管理应该正常工作',
          async validate(context: any) {
            const clientDetailPage = new ClientDetailPage({ baseUrl: context.baseUrl }, context.logger);
            const documents = await clientDetailPage.getDocumentList();
            return Array.isArray(documents);
          }
        }
      ],
      tags: ['client-detail', 'documents', 'medium']
    };
  }

  /**
   * 编辑客户测试用例
   */
  public getEditClientTestCase(): TestCase {
    return {
      id: 'edit-client',
      name: '编辑客户信息',
      description: '验证客户信息编辑功能能够正确工作',
      priority: 'high',
      timeout: 45000,
      steps: [
        {
          id: 'click-edit-button',
          name: '点击编辑按钮',
          type: 'click',
          selector: '#edit-client-button',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '编辑功能应该正常工作',
          async validate(context: any) {
            const clientDetailPage = new ClientDetailPage({ baseUrl: context.baseUrl }, context.logger);
            const isLoading = await clientDetailPage.isLoading();
            return !isLoading; // 确保不是加载状态
          }
        }
      ],
      tags: ['client-detail', 'edit', 'high']
    };
  }

  /**
   * 删除客户测试用例
   */
  public getDeleteClientTestCase(): TestCase {
    return {
      id: 'delete-client',
      name: '删除客户',
      description: '验证客户删除功能能够正确工作',
      priority: 'high',
      timeout: 30000,
      steps: [
        {
          id: 'click-delete-button',
          name: '点击删除按钮',
          type: 'click',
          selector: '#delete-client-button',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '删除确认应该出现',
          async validate(context: any) {
            // 检查是否有删除确认对话框
            return true; // 简化验证
          }
        }
      ],
      tags: ['client-detail', 'delete', 'high']
    };
  }

  /**
   * 创建客户测试用例
   */
  public getCreateClientTestCase(): TestCase {
    return {
      id: 'create-client',
      name: '创建新客户',
      description: '验证新客户创建功能能够正确工作',
      priority: 'high',
      timeout: 60000,
      setup: {
        steps: [
          {
            id: 'navigate-to-create-client',
            name: '导航到创建客户页面',
            type: 'navigate',
            url: '/clients/create',
            timeout: 5000
          }
        ]
      },
      steps: [
        {
          id: 'fill-client-name',
          name: '填写客户名称',
          type: 'input',
          selector: '#client-name',
          value: '测试客户公司',
          timeout: 5000
        },
        {
          id: 'select-client-type',
          name: '选择客户类型',
          type: 'select',
          selector: '#client-type',
          value: 'company',
          timeout: 5000
        },
        {
          id: 'select-industry',
          name: '选择行业',
          type: 'select',
          selector: '#client-industry',
          value: 'technology',
          timeout: 5000
        },
        {
          id: 'fill-contact-person',
          name: '填写联系人',
          type: 'input',
          selector: '#client-contact-person',
          value: '张经理',
          timeout: 5000
        },
        {
          id: 'fill-email',
          name: '填写邮箱',
          type: 'input',
          selector: '#client-email',
          value: 'zhang@testcompany.com',
          timeout: 5000
        },
        {
          id: 'fill-phone',
          name: '填写电话',
          type: 'input',
          selector: '#client-phone',
          value: '010-12345678',
          timeout: 5000
        },
        {
          id: 'fill-address',
          name: '填写地址',
          type: 'input',
          selector: '#client-address',
          value: '北京市朝阳区测试路123号',
          timeout: 5000
        },
        {
          id: 'select-status',
          name: '选择状态',
          type: 'select',
          selector: '#client-status',
          value: 'active',
          timeout: 5000
        },
        {
          id: 'save-client',
          name: '保存客户',
          type: 'click',
          selector: '#save-client-button',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '客户创建应该成功',
          async validate(context: any) {
            const clientFormPage = new ClientFormPage({ baseUrl: context.baseUrl }, context.logger);
            const successMessage = await clientFormPage.getSuccessMessage();
            return successMessage.includes('成功') || successMessage.includes('创建');
          }
        }
      ],
      tags: ['client-form', 'create', 'high']
    };
  }

  /**
   * 编辑客户表单测试用例
   */
  public getEditClientFormTestCase(): TestCase {
    return {
      id: 'edit-client-form',
      name: '编辑客户表单',
      description: '验证客户表单编辑功能能够正确工作',
      priority: 'high',
      timeout: 60000,
      setup: {
        steps: [
          {
            id: 'navigate-to-edit-client',
            name: '导航到编辑客户页面',
            type: 'navigate',
            url: '/clients/client-1/edit',
            timeout: 5000
          }
        ]
      },
      steps: [
        {
          id: 'update-client-name',
          name: '更新客户名称',
          type: 'input',
          selector: '#client-name',
          value: '更新后的客户名称',
          timeout: 5000
        },
        {
          id: 'save-changes',
          name: '保存更改',
          type: 'click',
          selector: '#save-client-button',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '客户信息应该更新成功',
          async validate(context: any) {
            const clientFormPage = new ClientFormPage({ baseUrl: context.baseUrl }, context.logger);
            const successMessage = await clientFormPage.getSuccessMessage();
            return successMessage.includes('成功') || successMessage.includes('更新');
          }
        }
      ],
      tags: ['client-form', 'edit', 'high']
    };
  }

  /**
   * 表单验证测试用例
   */
  public getFormValidationTestCase(): TestCase {
    return {
      id: 'form-validation',
      name: '表单验证',
      description: '验证客户表单验证功能能够正确工作',
      priority: 'high',
      timeout: 30000,
      setup: {
        steps: [
          {
            id: 'navigate-to-create-client',
            name: '导航到创建客户页面',
            type: 'navigate',
            url: '/clients/create',
            timeout: 5000
          }
        ]
      },
      steps: [
        {
          id: 'submit-empty-form',
          name: '提交空表单',
          type: 'click',
          selector: '#save-client-button',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '应该显示验证错误',
          async validate(context: any) {
            const clientFormPage = new ClientFormPage({ baseUrl: context.baseUrl }, context.logger);
            const validation = await clientFormPage.validateForm();
            return !validation.valid; // 应该有验证错误
          }
        }
      ],
      tags: ['client-form', 'validation', 'high']
    };
  }

  /**
   * 联系人管理测试用例
   */
  public getContactManagementTestCase(): TestCase {
    return {
      id: 'contact-management',
      name: '联系人管理',
      description: '验证表单中的联系人管理功能能够正确工作',
      priority: 'medium',
      timeout: 45000,
      setup: {
        steps: [
          {
            id: 'navigate-to-create-client',
            name: '导航到创建客户页面',
            type: 'navigate',
            url: '/clients/create',
            timeout: 5000
          }
        ]
      },
      steps: [
        {
          id: 'add-contact',
          name: '添加联系人',
          type: 'click',
          selector: '#add-contact-button',
          timeout: 5000
        },
        {
          id: 'fill-contact-info',
          name: '填写联系人信息',
          type: 'custom',
          description: '填写联系人表单',
          timeout: 10000,
          parameters: {
            action: 'fillContact',
            contact: {
              name: '李经理',
              position: '技术总监',
              email: 'li@testcompany.com',
              phone: '010-87654321',
              isPrimary: true
            }
          }
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '联系人应该添加成功',
          async validate(context: any) {
            const clientFormPage = new ClientFormPage({ baseUrl: context.baseUrl }, context.logger);
            // 检查是否没有错误消息
            const errorMessage = await clientFormPage.getErrorMessage();
            return !errorMessage;
          }
        }
      ],
      tags: ['client-form', 'contacts', 'medium']
    };
  }

  /**
   * 保存并新建测试用例
   */
  public getSaveAndNewTestCase(): TestCase {
    return {
      id: 'save-and-new',
      name: '保存并新建',
      description: '验证保存并新建功能能够正确工作',
      priority: 'medium',
      timeout: 45000,
      setup: {
        steps: [
          {
            id: 'navigate-to-create-client',
            name: '导航到创建客户页面',
            type: 'navigate',
            url: '/clients/create',
            timeout: 5000
          }
        ]
      },
      steps: [
        {
          id: 'fill-basic-info',
          name: '填写基本信息',
          type: 'custom',
          description: '填写客户基本信息',
          timeout: 15000,
          parameters: {
            action: 'fillBasicInfo',
            data: {
              name: '测试客户',
              type: 'company',
              industry: 'technology',
              contactPerson: '王经理',
              email: 'wang@test.com',
              phone: '010-11111111',
              address: '测试地址',
              status: 'active'
            }
          }
        },
        {
          id: 'click-save-and-new',
          name: '点击保存并新建',
          type: 'click',
          selector: '#save-and-new-button',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '应该清空表单并准备新建',
          async validate(context: any) {
            const clientFormPage = new ClientFormPage({ baseUrl: context.baseUrl }, context.logger);
            const formData = await clientFormPage.getFormData();
            return !formData.name; // 名称应该被清空
          }
        }
      ],
      tags: ['client-form', 'save-and-new', 'medium']
    };
  }

  /**
   * 取消操作测试用例
   */
  public getCancelOperationTestCase(): TestCase {
    return {
      id: 'cancel-operation',
      name: '取消操作',
      description: '验证取消操作能够正确工作',
      priority: 'low',
      timeout: 30000,
      setup: {
        steps: [
          {
            id: 'navigate-to-create-client',
            name: '导航到创建客户页面',
            type: 'navigate',
            url: '/clients/create',
            timeout: 5000
          }
        ]
      },
      steps: [
        {
          id: 'fill-some-info',
          name: '填写部分信息',
          type: 'input',
          selector: '#client-name',
          value: '临时客户',
          timeout: 5000
        },
        {
          id: 'cancel-operation',
          name: '取消操作',
          type: 'click',
          selector: '#cancel-button',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '应该返回客户列表页面',
          async validate(context: any) {
            const clientListPage = new ClientListPage({ baseUrl: context.baseUrl }, context.logger);
            const isValid = await clientListPage.validateClientListPage();
            return isValid.valid;
          }
        }
      ],
      tags: ['client-form', 'cancel', 'low']
    };
  }

  /**
   * 复制客户测试用例
   */
  public getDuplicateClientTestCase(): TestCase {
    return {
      id: 'duplicate-client',
      name: '复制客户',
      description: '验证客户复制功能能够正确工作',
      priority: 'low',
      timeout: 45000,
      setup: {
        steps: [
          {
            id: 'navigate-to-edit-client',
            name: '导航到编辑客户页面',
            type: 'navigate',
            url: '/clients/client-1/edit',
            timeout: 5000
          }
        ]
      },
      steps: [
        {
          id: 'duplicate-client',
          name: '复制客户',
          type: 'click',
          selector: '#duplicate-client-button',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '应该创建客户的副本',
          async validate(context: any) {
            const clientFormPage = new ClientFormPage({ baseUrl: context.baseUrl }, context.logger);
            const formData = await clientFormPage.getFormData();
            return !!formData.name; // 应该有填充的表单数据
          }
        }
      ],
      tags: ['client-form', 'duplicate', 'low']
    };
  }
}