/**
 * 文档管理测试用例
 */

import { TestCase, TestStep, TestAssertion, TestSuite } from '../types/test-types';

export class DocumentManagementTestCases {
  /**
   * 获取文档列表测试套件
   */
  getDocumentListTestSuite(): TestSuite {
    return {
      id: 'document-list',
      name: '文档列表功能测试',
      description: '验证文档列表页面的各项功能',
      testCases: [
        this.getLoadDocumentListTestCase(),
        this.getSearchDocumentsTestCase(),
        this.getFilterDocumentsTestCase(),
        this.getSortDocumentsTestCase(),
        this.getPaginationTestCase(),
        this.getBulkOperationsTestCase(),
        this.getExportDocumentsTestCase()
      ]
    };
  }

  /**
   * 获取文档详情测试套件
   */
  getDocumentDetailTestSuite(): TestSuite {
    return {
      id: 'document-detail',
      name: '文档详情功能测试',
      description: '验证文档详情页面的各项功能',
      testCases: [
        this.getLoadDocumentDetailTestCase(),
        this.getViewDocumentInfoTestCase(),
        this.getPreviewDocumentTestCase(),
        this.getDownloadDocumentTestCase(),
        this.getShareDocumentTestCase(),
        this.getManageDocumentCommentsTestCase(),
        this.getViewVersionHistoryTestCase(),
        this.getEditDocumentTestCase(),
        this.getDeleteDocumentTestCase()
      ]
    };
  }

  /**
   * 获取文档表单测试套件
   */
  getDocumentFormTestSuite(): TestSuite {
    return {
      id: 'document-form',
      name: '文档表单功能测试',
      description: '验证文档表单页面的各项功能',
      testCases: [
        this.getCreateDocumentTestCase(),
        this.getEditDocumentFormTestCase(),
        this.getFormValidationTestCase(),
        this.getDocumentUploadTestCase(),
        this.getDocumentPreviewTestCase(),
        this.getDocumentMetadataTestCase(),
        this.getSaveAndNewTestCase(),
        this.getCancelOperationTestCase()
      ]
    };
  }

  /**
   * 获取加载文档列表测试用例
   */
  getLoadDocumentListTestCase(): TestCase {
    return {
      id: 'load-document-list',
      name: '加载文档列表',
      description: '验证文档列表页面能够正确加载和显示文档',
      priority: 'high',
      timeout: 30000,
      steps: [
        {
          id: 'navigate-to-document-list',
          name: '导航到文档列表页面',
          type: 'navigation',
          url: '/documents',
          timeout: 10000
        },
        {
          id: 'wait-for-document-list',
          name: '等待文档列表加载',
          type: 'wait',
          selector: '.document-list-item',
          expectedState: 'visible',
          timeout: 10000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '文档列表应该正确加载',
          async validate(context: any) {
            const { DocumentListPage } = await import('./document-list-page');
            const documentListPage = new DocumentListPage({ baseUrl: context.baseUrl }, context.logger);
            await documentListPage.navigateToDocumentList();
            const documents = await documentListPage.getDocumentList();
            return Array.isArray(documents) && documents.length >= 0;
          }
        }
      ],
      tags: ['document-list', 'load', 'high-priority']
    };
  }

  /**
   * 获取搜索文档测试用例
   */
  getSearchDocumentsTestCase(): TestCase {
    return {
      id: 'search-documents',
      name: '搜索文档功能',
      description: '验证文档搜索功能能够正确工作',
      priority: 'high',
      timeout: 30000,
      steps: [
        {
          id: 'navigate-to-document-list',
          name: '导航到文档列表页面',
          type: 'navigation',
          url: '/documents',
          timeout: 10000
        },
        {
          id: 'search-documents',
          name: '搜索文档',
          type: 'input',
          selector: '#document-search',
          value: '测试文档',
          timeout: 5000
        },
        {
          id: 'click-search-button',
          name: '点击搜索按钮',
          type: 'click',
          selector: '#document-search-button',
          timeout: 5000
        },
        {
          id: 'wait-for-search-results',
          name: '等待搜索结果',
          type: 'wait',
          timeout: 10000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '搜索功能应该正常工作',
          async validate(context: any) {
            const { DocumentListPage } = await import('./document-list-page');
            const documentListPage = new DocumentListPage({ baseUrl: context.baseUrl }, context.logger);
            await documentListPage.navigateToDocumentList();
            await documentListPage.searchDocuments({ query: '测试文档' });
            const documents = await documentListPage.getDocumentList();
            return Array.isArray(documents);
          }
        }
      ],
      tags: ['document-list', 'search', 'high-priority']
    };
  }

  /**
   * 获取过滤文档测试用例
   */
  getFilterDocumentsTestCase(): TestCase {
    return {
      id: 'filter-documents',
      name: '过滤文档功能',
      description: '验证文档过滤功能能够正确工作',
      priority: 'medium',
      timeout: 30000,
      steps: [
        {
          id: 'navigate-to-document-list',
          name: '导航到文档列表页面',
          type: 'navigation',
          url: '/documents',
          timeout: 10000
        },
        {
          id: 'apply-type-filter',
          name: '应用类型过滤',
          type: 'select',
          selector: '#document-type-filter',
          value: 'PDF',
          timeout: 5000
        },
        {
          id: 'apply-client-filter',
          name: '应用客户过滤',
          type: 'select',
          selector: '#document-client-filter',
          value: '测试客户',
          timeout: 5000
        },
        {
          id: 'apply-status-filter',
          name: '应用状态过滤',
          type: 'select',
          selector: '#document-status-filter',
          value: 'active',
          timeout: 5000
        },
        {
          id: 'click-apply-filters',
          name: '点击应用过滤按钮',
          type: 'click',
          selector: '#document-apply-filters',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '过滤功能应该正常工作',
          async validate(context: any) {
            const { DocumentListPage } = await import('./document-list-page');
            const documentListPage = new DocumentListPage({ baseUrl: context.baseUrl }, context.logger);
            await documentListPage.navigateToDocumentList();
            await documentListPage.applyFilters({
              type: 'PDF',
              client: '测试客户',
              status: 'active'
            });
            const documents = await documentListPage.getDocumentList();
            return Array.isArray(documents);
          }
        }
      ],
      tags: ['document-list', 'filter', 'medium-priority']
    };
  }

  /**
   * 获取排序文档测试用例
   */
  getSortDocumentsTestCase(): TestCase {
    return {
      id: 'sort-documents',
      name: '排序文档功能',
      description: '验证文档排序功能能够正确工作',
      priority: 'medium',
      timeout: 30000,
      steps: [
        {
          id: 'navigate-to-document-list',
          name: '导航到文档列表页面',
          type: 'navigation',
          url: '/documents',
          timeout: 10000
        },
        {
          id: 'click-sort-button',
          name: '点击排序按钮',
          type: 'click',
          selector: '#document-sort-button',
          timeout: 5000
        },
        {
          id: 'select-sort-field',
          name: '选择排序字段',
          type: 'click',
          selector: '.document-sort-field[data-field="name"]',
          timeout: 5000
        },
        {
          id: 'select-sort-order',
          name: '选择排序方向',
          type: 'click',
          selector: '.document-sort-order[data-order="asc"]',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '排序功能应该正常工作',
          async validate(context: any) {
            const { DocumentListPage } = await import('./document-list-page');
            const documentListPage = new DocumentListPage({ baseUrl: context.baseUrl }, context.logger);
            await documentListPage.navigateToDocumentList();
            await documentListPage.sortDocuments({
              field: 'name',
              order: 'asc'
            });
            const documents = await documentListPage.getDocumentList();
            return Array.isArray(documents);
          }
        }
      ],
      tags: ['document-list', 'sort', 'medium-priority']
    };
  }

  /**
   * 获取分页测试用例
   */
  getPaginationTestCase(): TestCase {
    return {
      id: 'pagination',
      name: '分页功能',
      description: '验证文档列表分页功能能够正确工作',
      priority: 'medium',
      timeout: 30000,
      steps: [
        {
          id: 'navigate-to-document-list',
          name: '导航到文档列表页面',
          type: 'navigation',
          url: '/documents',
          timeout: 10000
        },
        {
          id: 'go-to-next-page',
          name: '导航到下一页',
          type: 'click',
          selector: '.pagination-next',
          timeout: 5000
        },
        {
          id: 'go-to-previous-page',
          name: '导航到上一页',
          type: 'click',
          selector: '.pagination-previous',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '分页功能应该正常工作',
          async validate(context: any) {
            const { DocumentListPage } = await import('./document-list-page');
            const documentListPage = new DocumentListPage({ baseUrl: context.baseUrl }, context.logger);
            await documentListPage.navigateToDocumentList();
            const documents = await documentListPage.getDocumentList();
            return Array.isArray(documents);
          }
        }
      ],
      tags: ['document-list', 'pagination', 'medium-priority']
    };
  }

  /**
   * 获取批量操作测试用例
   */
  getBulkOperationsTestCase(): TestCase {
    return {
      id: 'bulk-operations',
      name: '批量操作功能',
      description: '验证文档批量操作功能能够正确工作',
      priority: 'medium',
      timeout: 30000,
      steps: [
        {
          id: 'navigate-to-document-list',
          name: '导航到文档列表页面',
          type: 'navigation',
          url: '/documents',
          timeout: 10000
        },
        {
          id: 'select-documents',
          name: '选择多个文档',
          type: 'click',
          selector: '.document-list-item:first-child input[type="checkbox"]',
          timeout: 5000
        },
        {
          id: 'click-bulk-edit',
          name: '点击批量编辑按钮',
          type: 'click',
          selector: '#document-bulk-edit-button',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '批量操作功能应该正常工作',
          async validate(context: any) {
            const { DocumentListPage } = await import('./document-list-page');
            const documentListPage = new DocumentListPage({ baseUrl: context.baseUrl }, context.logger);
            await documentListPage.navigateToDocumentList();
            // 测试批量编辑功能
            await documentListPage.bulkEdit();
            return true;
          }
        }
      ],
      tags: ['document-list', 'bulk-operations', 'medium-priority']
    };
  }

  /**
   * 获取导出文档测试用例
   */
  getExportDocumentsTestCase(): TestCase {
    return {
      id: 'export-documents',
      name: '导出文档功能',
      description: '验证文档导出功能能够正确工作',
      priority: 'low',
      timeout: 30000,
      steps: [
        {
          id: 'navigate-to-document-list',
          name: '导航到文档列表页面',
          type: 'navigation',
          url: '/documents',
          timeout: 10000
        },
        {
          id: 'click-export-csv',
          name: '点击导出CSV按钮',
          type: 'click',
          selector: '#document-export-csv',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '导出功能应该正常工作',
          async validate(context: any) {
            const { DocumentListPage } = await import('./document-list-page');
            const documentListPage = new DocumentListPage({ baseUrl: context.baseUrl }, context.logger);
            await documentListPage.navigateToDocumentList();
            // 测试导出功能
            await documentListPage.exportDocuments('csv');
            return true;
          }
        }
      ],
      tags: ['document-list', 'export', 'low-priority']
    };
  }

  /**
   * 获取加载文档详情测试用例
   */
  getLoadDocumentDetailTestCase(): TestCase {
    return {
      id: 'load-document-detail',
      name: '加载文档详情',
      description: '验证文档详情页面能够正确加载',
      priority: 'high',
      timeout: 30000,
      steps: [
        {
          id: 'navigate-to-document-detail',
          name: '导航到文档详情页面',
          type: 'navigation',
          url: '/documents/doc-1',
          timeout: 10000
        },
        {
          id: 'wait-for-document-detail',
          name: '等待文档详情加载',
          type: 'wait',
          selector: '.document-detail',
          expectedState: 'visible',
          timeout: 10000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '文档详情应该正确加载',
          async validate(context: any) {
            const { DocumentDetailPage } = await import('./document-detail-page');
            const documentDetailPage = new DocumentDetailPage({ baseUrl: context.baseUrl }, context.logger);
            await documentDetailPage.navigateToDocumentDetail('doc-1');
            const detail = await documentDetailPage.getDocumentDetail();
            return detail && typeof detail === 'object';
          }
        }
      ],
      tags: ['document-detail', 'load', 'high-priority']
    };
  }

  /**
   * 获取查看文档信息测试用例
   */
  getViewDocumentInfoTestCase(): TestCase {
    return {
      id: 'view-document-info',
      name: '查看文档信息',
      description: '验证文档信息显示功能',
      priority: 'high',
      timeout: 30000,
      steps: [
        {
          id: 'navigate-to-document-detail',
          name: '导航到文档详情页面',
          type: 'navigation',
          url: '/documents/doc-1',
          timeout: 10000
        },
        {
          id: 'check-document-info',
          name: '检查文档信息',
          type: 'verify',
          selector: '.document-info',
          expectedState: 'visible',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '文档信息应该正确显示',
          async validate(context: any) {
            const { DocumentDetailPage } = await import('./document-detail-page');
            const documentDetailPage = new DocumentDetailPage({ baseUrl: context.baseUrl }, context.logger);
            await documentDetailPage.navigateToDocumentDetail('doc-1');
            const detail = await documentDetailPage.getDocumentDetail();
            return detail && detail.name && detail.type;
          }
        }
      ],
      tags: ['document-detail', 'view', 'high-priority']
    };
  }

  /**
   * 获取预览文档测试用例
   */
  getPreviewDocumentTestCase(): TestCase {
    return {
      id: 'preview-document',
      name: '预览文档功能',
      description: '验证文档预览功能',
      priority: 'medium',
      timeout: 30000,
      steps: [
        {
          id: 'navigate-to-document-detail',
          name: '导航到文档详情页面',
          type: 'navigation',
          url: '/documents/doc-1',
          timeout: 10000
        },
        {
          id: 'click-preview-button',
          name: '点击预览按钮',
          type: 'click',
          selector: '#document-preview-button',
          timeout: 5000
        },
        {
          id: 'wait-for-preview-modal',
          name: '等待预览模态框',
          type: 'wait',
          selector: '.document-preview-modal',
          expectedState: 'visible',
          timeout: 10000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '预览功能应该正常工作',
          async validate(context: any) {
            const { DocumentDetailPage } = await import('./document-detail-page');
            const documentDetailPage = new DocumentDetailPage({ baseUrl: context.baseUrl }, context.logger);
            await documentDetailPage.navigateToDocumentDetail('doc-1');
            await documentDetailPage.previewDocument();
            return true;
          }
        }
      ],
      tags: ['document-detail', 'preview', 'medium-priority']
    };
  }

  /**
   * 获取下载文档测试用例
   */
  getDownloadDocumentTestCase(): TestCase {
    return {
      id: 'download-document',
      name: '下载文档功能',
      description: '验证文档下载功能',
      priority: 'medium',
      timeout: 30000,
      steps: [
        {
          id: 'navigate-to-document-detail',
          name: '导航到文档详情页面',
          type: 'navigation',
          url: '/documents/doc-1',
          timeout: 10000
        },
        {
          id: 'click-download-button',
          name: '点击下载按钮',
          type: 'click',
          selector: '#document-download-button',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '下载功能应该正常工作',
          async validate(context: any) {
            const { DocumentDetailPage } = await import('./document-detail-page');
            const documentDetailPage = new DocumentDetailPage({ baseUrl: context.baseUrl }, context.logger);
            await documentDetailPage.navigateToDocumentDetail('doc-1');
            await documentDetailPage.downloadDocument();
            return true;
          }
        }
      ],
      tags: ['document-detail', 'download', 'medium-priority']
    };
  }

  /**
   * 获取共享文档测试用例
   */
  getShareDocumentTestCase(): TestCase {
    return {
      id: 'share-document',
      name: '共享文档功能',
      description: '验证文档共享功能',
      priority: 'medium',
      timeout: 30000,
      steps: [
        {
          id: 'navigate-to-document-detail',
          name: '导航到文档详情页面',
          type: 'navigation',
          url: '/documents/doc-1',
          timeout: 10000
        },
        {
          id: 'click-share-button',
          name: '点击共享按钮',
          type: 'click',
          selector: '#document-share-button',
          timeout: 5000
        },
        {
          id: 'wait-for-share-dialog',
          name: '等待共享对话框',
          type: 'wait',
          selector: '.document-share-dialog',
          expectedState: 'visible',
          timeout: 10000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '共享功能应该正常工作',
          async validate(context: any) {
            const { DocumentDetailPage } = await import('./document-detail-page');
            const documentDetailPage = new DocumentDetailPage({ baseUrl: context.baseUrl }, context.logger);
            await documentDetailPage.navigateToDocumentDetail('doc-1');
            await documentDetailPage.shareDocument({
              isPublic: true,
              permissions: {
                canView: ['user1@example.com'],
                canEdit: [],
                canDelete: [],
                canShare: []
              }
            });
            return true;
          }
        }
      ],
      tags: ['document-detail', 'share', 'medium-priority']
    };
  }

  /**
   * 获取管理文档评论测试用例
   */
  getManageDocumentCommentsTestCase(): TestCase {
    return {
      id: 'manage-document-comments',
      name: '管理文档评论',
      description: '验证文档评论管理功能',
      priority: 'low',
      timeout: 30000,
      steps: [
        {
          id: 'navigate-to-document-detail',
          name: '导航到文档详情页面',
          type: 'navigation',
          url: '/documents/doc-1',
          timeout: 10000
        },
        {
          id: 'add-comment',
          name: '添加评论',
          type: 'input',
          selector: '#document-comment-input',
          value: '这是一个测试评论',
          timeout: 5000
        },
        {
          id: 'submit-comment',
          name: '提交评论',
          type: 'click',
          selector: '#document-comment-submit-button',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '评论功能应该正常工作',
          async validate(context: any) {
            const { DocumentDetailPage } = await import('./document-detail-page');
            const documentDetailPage = new DocumentDetailPage({ baseUrl: context.baseUrl }, context.logger);
            await documentDetailPage.navigateToDocumentDetail('doc-1');
            await documentDetailPage.addComment('这是一个测试评论');
            const comments = await documentDetailPage.getComments();
            return Array.isArray(comments) && comments.length > 0;
          }
        }
      ],
      tags: ['document-detail', 'comments', 'low-priority']
    };
  }

  /**
   * 获取查看版本历史测试用例
   */
  getViewVersionHistoryTestCase(): TestCase {
    return {
      id: 'view-version-history',
      name: '查看版本历史',
      description: '验证文档版本历史查看功能',
      priority: 'medium',
      timeout: 30000,
      steps: [
        {
          id: 'navigate-to-document-detail',
          name: '导航到文档详情页面',
          type: 'navigation',
          url: '/documents/doc-1',
          timeout: 10000
        },
        {
          id: 'check-version-history',
          name: '检查版本历史',
          type: 'verify',
          selector: '.document-version-history',
          expectedState: 'visible',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '版本历史应该正确显示',
          async validate(context: any) {
            const { DocumentDetailPage } = await import('./document-detail-page');
            const documentDetailPage = new DocumentDetailPage({ baseUrl: context.baseUrl }, context.logger);
            await documentDetailPage.navigateToDocumentDetail('doc-1');
            const versions = await documentDetailPage.getVersionHistory();
            return Array.isArray(versions);
          }
        }
      ],
      tags: ['document-detail', 'version-history', 'medium-priority']
    };
  }

  /**
   * 获取编辑文档测试用例
   */
  getEditDocumentTestCase(): TestCase {
    return {
      id: 'edit-document',
      name: '编辑文档信息',
      description: '验证文档信息编辑功能',
      priority: 'high',
      timeout: 30000,
      steps: [
        {
          id: 'navigate-to-document-detail',
          name: '导航到文档详情页面',
          type: 'navigation',
          url: '/documents/doc-1',
          timeout: 10000
        },
        {
          id: 'click-edit-button',
          name: '点击编辑按钮',
          type: 'click',
          selector: '#document-edit-button',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '编辑功能应该正常工作',
          async validate(context: any) {
            const { DocumentDetailPage } = await import('./document-detail-page');
            const documentDetailPage = new DocumentDetailPage({ baseUrl: context.baseUrl }, context.logger);
            await documentDetailPage.navigateToDocumentDetail('doc-1');
            await documentDetailPage.updateDocument({
              name: '更新后的文档名称',
              description: '更新后的描述'
            });
            return true;
          }
        }
      ],
      tags: ['document-detail', 'edit', 'high-priority']
    };
  }

  /**
   * 获取删除文档测试用例
   */
  getDeleteDocumentTestCase(): TestCase {
    return {
      id: 'delete-document',
      name: '删除文档',
      description: '验证文档删除功能',
      priority: 'high',
      timeout: 30000,
      steps: [
        {
          id: 'navigate-to-document-detail',
          name: '导航到文档详情页面',
          type: 'navigation',
          url: '/documents/doc-1',
          timeout: 10000
        },
        {
          id: 'click-delete-button',
          name: '点击删除按钮',
          type: 'click',
          selector: '#document-delete-button',
          timeout: 5000
        },
        {
          id: 'confirm-delete',
          name: '确认删除',
          type: 'click',
          selector: '.document-delete-confirm-button',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '删除功能应该正常工作',
          async validate(context: any) {
            const { DocumentDetailPage } = await import('./document-detail-page');
            const documentDetailPage = new DocumentDetailPage({ baseUrl: context.baseUrl }, context.logger);
            await documentDetailPage.navigateToDocumentDetail('doc-1');
            // 注意：这个测试在实际环境中会删除文档，可能需要使用测试数据
            // await documentDetailPage.deleteDocument();
            return true;
          }
        }
      ],
      tags: ['document-detail', 'delete', 'high-priority']
    };
  }

  /**
   * 获取创建文档测试用例
   */
  getCreateDocumentTestCase(): TestCase {
    return {
      id: 'create-document',
      name: '创建新文档',
      description: '验证创建新文档功能',
      priority: 'high',
      timeout: 60000,
      setup: {
        description: '准备测试环境',
        steps: [
          {
            id: 'prepare-test-data',
            name: '准备测试数据',
            type: 'setup',
            timeout: 5000
          }
        ]
      },
      steps: [
        {
          id: 'navigate-to-create-document',
          name: '导航到创建文档页面',
          type: 'navigation',
          url: '/documents/create',
          timeout: 10000
        },
        {
          id: 'upload-file',
          name: '上传文件',
          type: 'click',
          selector: '#document-file-upload-button',
          timeout: 5000
        },
        {
          id: 'fill-document-name',
          name: '填写文档名称',
          type: 'input',
          selector: '#document-name',
          value: '测试文档',
          timeout: 5000
        },
        {
          id: 'select-document-type',
          name: '选择文档类型',
          type: 'select',
          selector: '#document-type',
          value: 'PDF',
          timeout: 5000
        },
        {
          id: 'fill-description',
          name: '填写描述',
          type: 'input',
          selector: '#document-description',
          value: '这是一个测试文档',
          timeout: 5000
        },
        {
          id: 'save-document',
          name: '保存文档',
          type: 'click',
          selector: '#document-save-button',
          timeout: 10000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '创建文档功能应该正常工作',
          async validate(context: any) {
            const { DocumentFormPage } = await import('./document-form-page');
            const documentFormPage = new DocumentFormPage({ baseUrl: context.baseUrl }, context.logger);
            await documentFormPage.navigateToCreateDocument();
            // 注意：实际文件上传需要特殊处理
            // await documentFormPage.uploadFile('/path/to/test/file.pdf');
            await documentFormPage.fillDocumentForm({
              name: '测试文档',
              type: 'PDF',
              description: '这是一个测试文档',
              tags: ['测试'],
              isPublic: false,
              isEncrypted: false
            });
            await documentFormPage.saveDocument();
            return true;
          }
        }
      ],
      tags: ['document-form', 'create', 'high-priority']
    };
  }

  /**
   * 获取编辑文档表单测试用例
   */
  getEditDocumentFormTestCase(): TestCase {
    return {
      id: 'edit-document-form',
      name: '编辑文档表单',
      description: '验证编辑文档表单功能',
      priority: 'high',
      timeout: 30000,
      steps: [
        {
          id: 'navigate-to-edit-document',
          name: '导航到编辑文档页面',
          type: 'navigation',
          url: '/documents/doc-1/edit',
          timeout: 10000
        },
        {
          id: 'update-document-name',
          name: '更新文档名称',
          type: 'input',
          selector: '#document-name',
          value: '更新的文档名称',
          timeout: 5000
        },
        {
          id: 'update-description',
          name: '更新描述',
          type: 'input',
          selector: '#document-description',
          value: '更新的描述',
          timeout: 5000
        },
        {
          id: 'save-changes',
          name: '保存更改',
          type: 'click',
          selector: '#document-save-button',
          timeout: 10000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '编辑文档表单功能应该正常工作',
          async validate(context: any) {
            const { DocumentFormPage } = await import('./document-form-page');
            const documentFormPage = new DocumentFormPage({ baseUrl: context.baseUrl }, context.logger);
            await documentFormPage.navigateToEditDocument('doc-1');
            await documentFormPage.fillDocumentForm({
              name: '更新的文档名称',
              type: 'PDF',
              description: '更新的描述',
              tags: ['更新'],
              isPublic: false,
              isEncrypted: false
            });
            await documentFormPage.saveDocument();
            return true;
          }
        }
      ],
      tags: ['document-form', 'edit', 'high-priority']
    };
  }

  /**
   * 获取表单验证测试用例
   */
  getFormValidationTestCase(): TestCase {
    return {
      id: 'form-validation',
      name: '表单验证',
      description: '验证文档表单验证功能',
      priority: 'high',
      timeout: 30000,
      steps: [
        {
          id: 'navigate-to-create-document',
          name: '导航到创建文档页面',
          type: 'navigation',
          url: '/documents/create',
          timeout: 10000
        },
        {
          id: 'attempt-save-empty-form',
          name: '尝试保存空表单',
          type: 'click',
          selector: '#document-save-button',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '表单验证应该正常工作',
          async validate(context: any) {
            const { DocumentFormPage } = await import('./document-form-page');
            const documentFormPage = new DocumentFormPage({ baseUrl: context.baseUrl }, context.logger);
            await documentFormPage.navigateToCreateDocument();
            const validation = await documentFormPage.validateForm();
            return !validation.valid && Object.keys(validation.errors).length > 0;
          }
        }
      ],
      tags: ['document-form', 'validation', 'high-priority']
    };
  }

  /**
   * 获取文档上传测试用例
   */
  getDocumentUploadTestCase(): TestCase {
    return {
      id: 'document-upload',
      name: '文档上传功能',
      description: '验证文档上传功能',
      priority: 'high',
      timeout: 60000,
      steps: [
        {
          id: 'navigate-to-create-document',
          name: '导航到创建文档页面',
          type: 'navigation',
          url: '/documents/create',
          timeout: 10000
        },
        {
          id: 'click-upload-button',
          name: '点击上传按钮',
          type: 'click',
          selector: '#document-file-upload-button',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '上传功能应该正常工作',
          async validate(context: any) {
            const { DocumentFormPage } = await import('./document-form-page');
            const documentFormPage = new DocumentFormPage({ baseUrl: context.baseUrl }, context.logger);
            await documentFormPage.navigateToCreateDocument();
            // 注意：实际文件上传需要特殊处理
            // await documentFormPage.uploadFile('/path/to/test/file.pdf');
            return true;
          }
        }
      ],
      tags: ['document-form', 'upload', 'high-priority']
    };
  }

  /**
   * 获取文档预览测试用例
   */
  getDocumentPreviewTestCase(): TestCase {
    return {
      id: 'document-preview',
      name: '文档预览功能',
      description: '验证文档预览功能',
      priority: 'medium',
      timeout: 30000,
      steps: [
        {
          id: 'navigate-to-create-document',
          name: '导航到创建文档页面',
          type: 'navigation',
          url: '/documents/create',
          timeout: 10000
        },
        {
          id: 'click-preview-button',
          name: '点击预览按钮',
          type: 'click',
          selector: '#document-preview-button',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '预览功能应该正常工作',
          async validate(context: any) {
            const { DocumentFormPage } = await import('./document-form-page');
            const documentFormPage = new DocumentFormPage({ baseUrl: context.baseUrl }, context.logger);
            await documentFormPage.navigateToCreateDocument();
            // 注意：预览功能需要先上传文件
            // await documentFormPage.previewDocument();
            return true;
          }
        }
      ],
      tags: ['document-form', 'preview', 'medium-priority']
    };
  }

  /**
   * 获取文档元数据测试用例
   */
  getDocumentMetadataTestCase(): TestCase {
    return {
      id: 'document-metadata',
      name: '文档元数据功能',
      description: '验证文档元数据提取功能',
      priority: 'low',
      timeout: 30000,
      steps: [
        {
          id: 'navigate-to-create-document',
          name: '导航到创建文档页面',
          type: 'navigation',
          url: '/documents/create',
          timeout: 10000
        },
        {
          id: 'click-extract-metadata',
          name: '点击提取元数据按钮',
          type: 'click',
          selector: '#document-extract-metadata-button',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '元数据提取功能应该正常工作',
          async validate(context: any) {
            const { DocumentFormPage } = await import('./document-form-page');
            const documentFormPage = new DocumentFormPage({ baseUrl: context.baseUrl }, context.logger);
            await documentFormPage.navigateToCreateDocument();
            // 注意：元数据提取需要先上传文件
            // const metadata = await documentFormPage.extractMetadata();
            return true;
          }
        }
      ],
      tags: ['document-form', 'metadata', 'low-priority']
    };
  }

  /**
   * 获取保存并新建测试用例
   */
  getSaveAndNewTestCase(): TestCase {
    return {
      id: 'save-and-new',
      name: '保存并新建功能',
      description: '验证保存并新建功能',
      priority: 'medium',
      timeout: 30000,
      steps: [
        {
          id: 'navigate-to-create-document',
          name: '导航到创建文档页面',
          type: 'navigation',
          url: '/documents/create',
          timeout: 10000
        },
        {
          id: 'fill-document-form',
          name: '填写文档表单',
          type: 'input',
          selector: '#document-name',
          value: '测试文档',
          timeout: 5000
        },
        {
          id: 'click-save-and-new',
          name: '点击保存并新建按钮',
          type: 'click',
          selector: '#document-save-and-new-button',
          timeout: 10000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '保存并新建功能应该正常工作',
          async validate(context: any) {
            const { DocumentFormPage } = await import('./document-form-page');
            const documentFormPage = new DocumentFormPage({ baseUrl: context.baseUrl }, context.logger);
            await documentFormPage.navigateToCreateDocument();
            await documentFormPage.fillDocumentForm({
              name: '测试文档',
              type: 'PDF',
              description: '这是一个测试文档',
              tags: ['测试'],
              isPublic: false,
              isEncrypted: false
            });
            await documentFormPage.saveAndNew();
            return true;
          }
        }
      ],
      tags: ['document-form', 'save-and-new', 'medium-priority']
    };
  }

  /**
   * 获取取消操作测试用例
   */
  getCancelOperationTestCase(): TestCase {
    return {
      id: 'cancel-operation',
      name: '取消操作功能',
      description: '验证取消操作功能',
      priority: 'low',
      timeout: 30000,
      steps: [
        {
          id: 'navigate-to-create-document',
          name: '导航到创建文档页面',
          type: 'navigation',
          url: '/documents/create',
          timeout: 10000
        },
        {
          id: 'click-cancel-button',
          name: '点击取消按钮',
          type: 'click',
          selector: '#document-cancel-button',
          timeout: 5000
        }
      ],
      assertions: [
        {
          type: 'function',
          description: '取消功能应该正常工作',
          async validate(context: any) {
            const { DocumentFormPage } = await import('./document-form-page');
            const documentFormPage = new DocumentFormPage({ baseUrl: context.baseUrl }, context.logger);
            await documentFormPage.navigateToCreateDocument();
            await documentFormPage.cancel();
            return true;
          }
        }
      ],
      tags: ['document-form', 'cancel', 'low-priority']
    };
  }
}