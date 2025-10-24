/**
 * 文档管理页面基类
 */

import { BasePageObject } from '../core/base-page-object';
import { Logger } from '../core/logger';
import { ClientListItem } from '../clients/client-list-page';

export interface DocumentFilters {
  type?: string;
  client?: string;
  case?: string;
  status?: string;
  uploadedBy?: string;
  dateRange?: {
    start: Date;
    end: Date;
  };
}

export interface DocumentSortOptions {
  field: 'name' | 'type' | 'size' | 'uploadDate' | 'uploadedBy';
  order: 'asc' | 'desc';
}

export interface DocumentListItem {
  id: string;
  name: string;
  type: string;
  size: number;
  uploadDate: Date;
  uploadedBy: string;
  client: string;
  case?: string;
  status: string;
  tags: string[];
  isPublic: boolean;
  isEncrypted: boolean;
}

export interface DocumentDetail {
  id: string;
  name: string;
  type: string;
  size: number;
  uploadDate: Date;
  uploadedBy: string;
  client: ClientListItem;
  case?: any;
  status: string;
  description: string;
  tags: string[];
  metadata: {
    pages?: number;
    author?: string;
    createdDate?: Date;
    modifiedDate?: Date;
    [key: string]: any;
  };
  permissions: {
    canView: string[];
    canEdit: string[];
    canDelete: string[];
    canShare: string[];
  };
  versions: Array<{
    id: string;
    version: string;
    uploadDate: Date;
    uploadedBy: string;
    size: number;
    comment?: string;
  }>;
  isPublic: boolean;
  isEncrypted: boolean;
  encryptionKey?: string;
}

export interface DocumentUploadData {
  file: string; // 文件路径
  name: string;
  type: string;
  description?: string;
  client?: string;
  case?: string;
  tags: string[];
  isPublic: boolean;
  isEncrypted: boolean;
  permissions?: {
    canView?: string[];
    canEdit?: string[];
    canDelete?: string[];
    canShare?: string[];
  };
}

export interface DocumentUpdateData {
  name?: string;
  description?: string;
  tags?: string[];
  isPublic?: boolean;
  permissions?: {
    canView?: string[];
    canEdit?: string[];
    canDelete?: string[];
    canShare?: string[];
  };
}

export interface DocumentSearchOptions {
  query?: string;
  type?: string;
  client?: string;
  case?: string;
  uploadedBy?: string;
  tags?: string[];
  content?: string;
  metadata?: Record<string, any>;
}

export class DocumentListPage extends BasePageObject {
  private baseUrl: string;

  constructor(config: { baseUrl: string; defaultTimeout?: number; screenshotOnFailure?: boolean }, logger?: Logger) {
    super(config, this.selectors, logger);
    this.baseUrl = config.baseUrl;
  }

  /**
   * 导航到文档列表页面
   */
  override async navigateToDocumentList(): Promise<void> {
    await this.navigate(`${this.baseUrl}/documents`);
  }

  /**
   * 获取文档列表
   */
  override async getDocumentList(): Promise<DocumentListItem[]> {
    try {
      const documents = await this.executeScript(`
        (function() {
          const documentItems = Array.from(document.querySelectorAll('.document-list-item'));
          return documentItems.map(item => {
            const id = item.getAttribute('data-document-id') || '';
            const name = item.querySelector('.document-name')?.gettextContent?.().trim() || '';
            const type = item.querySelector('.document-type')?.gettextContent?.().trim() || '';
            const sizeText = item.querySelector('.document-size')?.gettextContent?.().trim() || '0 B';
            const uploadDateText = item.querySelector('.document-upload-date')?.gettextContent?.().trim() || '';
            const uploadedBy = item.querySelector('.document-uploaded-by')?.gettextContent?.().trim() || '';
            const client = item.querySelector('.document-client')?.gettextContent?.().trim() || '';
            const caseText = item.querySelector('.document-case')?.gettextContent?.().trim();
            const status = item.querySelector('.document-status')?.gettextContent?.().trim() || '';

            const tagsElements = item.querySelectorAll('.document-tag');
            const tags = Array.from(tagsElements).map(tag => tag.gettextContent?.().trim() || '');

            const isPublic = item.querySelector('.document-public') !== null;
            const isEncrypted = item.querySelector('.document-encrypted') !== null;

            // 解析文件大小
            const parseSize = (sizeText) => {
              const units = { 'B': 1, 'KB': 1024, 'MB': 1024 * 1024, 'GB': 1024 * 1024 * 1024 };
              const match = sizeText.match(/^([\\d.]+)\\s+(B|KB|MB|GB)$/i);
              if (match) {
                return parseFloat(match[1]) * units[match[2].toUpperCase()];
              }
              return 0;
            };

            // 解析日期
            const parseDate = (dateText) => {
              const date = new Date(dateText);
              return isNaN(date.getTime()) ? new Date() : date;
            };

            return {
              id,
              name,
              type,
              size: parseSize(sizeText),
              uploadDate: parseDate(uploadDateText),
              uploadedBy,
              client,
              case: caseText || undefined,
              status,
              tags,
              isPublic,
              isEncrypted
            };
          });
        })();
      `);

      return documents || [];
    } catch (error) {
      this.logger.error('获取文档列表失败', { error });
      throw error;
    }
  }

  /**
   * 搜索文档
   */
  override async searchDocuments(options: DocumentSearchOptions): Promise<void> {
    try {
      if (options.query) {
        await this.fill('#document-search', options.query);
      }

      if (options.type) {
        await this.selectOption('#document-type-filter', [options.type]);
      }

      if (options.client) {
        await this.selectOption('#document-client-filter', [options.client]);
      }

      if (options.case) {
        await this.selectOption('#document-case-filter', [options.case]);
      }

      if (options.uploadedBy) {
        await this.selectOption('#document-uploaded-by-filter', [options.uploadedBy]);
      }

      if (options.tags && options.tags.length > 0) {
        await this.fill('#document-tags-filter', options.tags.join(', '));
      }

      // 点击搜索按钮
      await this.click('#document-search-button');
      await this.wait(2000); // 等待搜索结果

    } catch (error) {
      this.logger.error('搜索文档失败', { error, options });
      throw error;
    }
  }

  /**
   * 应用过滤条件
   */
  override async applyFilters(filters: DocumentFilters): Promise<void> {
    try {
      if (filters.type) {
        await this.selectOption('#document-type-filter', [filters.type]);
      }

      if (filters.client) {
        await this.selectOption('#document-client-filter', [filters.client]);
      }

      if (filters.case) {
        await this.selectOption('#document-case-filter', [filters.case]);
      }

      if (filters.status) {
        await this.selectOption('#document-status-filter', [filters.status]);
      }

      if (filters.uploadedBy) {
        await this.selectOption('#document-uploaded-by-filter', [filters.uploadedBy]);
      }

      if (filters.dateRange) {
        const startDate = filters.dateRange.start.toISOString().split('T')[0];
        const endDate = filters.dateRange.end.toISOString().split('T')[0];

        await this.fill('#document-date-start', startDate);
        await this.fill('#document-date-end', endDate);
      }

      await this.click('#document-apply-filters');
      await this.wait(2000); // 等待过滤结果

    } catch (error) {
      this.logger.error('应用文档过滤条件失败', { error, filters });
      throw error;
    }
  }

  /**
   * 排序文档
   */
  override async sortDocuments(sortOptions: DocumentSortOptions): Promise<void> {
    try {
      const sortButtonSelector = '#document-sort-button';
      await this.click(sortButtonSelector);

      // 等待排序菜单出现
      await this.waitForElement('.document-sort-menu', { timeout: 5000 });

      // 选择排序字段
      const fieldSelector = `.document-sort-field[data-field="${sortOptions.field}"]`;
      await this.click(fieldSelector);

      // 选择排序方向
      const orderSelector = `.document-sort-order[data-order="${sortOptions.order}"]`;
      await this.click(orderSelector);

      await this.click(sortButtonSelector); // 关闭菜单
      await this.wait(1000); // 等待排序完成

    } catch (error) {
      this.logger.error('排序文档失败', { error, sortOptions });
      throw error;
    }
  }

  /**
   * 清除过滤条件
   */
  override async clearFilters(): Promise<void> {
    try {
      await this.click('#document-clear-filters');
      await this.wait(1000); // 等待过滤条件清除
    } catch (error) {
      this.logger.error('清除文档过滤条件失败', { error });
      throw error;
    }
  }

  /**
   * 选择文档
   */
  override async selectDocument(documentId: string): Promise<void> {
    try {
      const selector = `.document-list-item[data-document-id="${documentId}"] input[type="checkbox"]`;
      await this.click(selector);
      await this.wait(500);
    } catch (error) {
      this.logger.error('选择文档失败', { error, documentId });
      throw error;
    }
  }

  /**
   * 查看文档详情
   */
  override async viewDocument(documentId: string): Promise<void> {
    try {
      const selector = `.document-list-item[data-document-id="${documentId}"] .document-view-button`;
      await this.click(selector);
      await this.wait(1000); // 等待导航完成
    } catch (error) {
      this.logger.error('查看文档失败', { error, documentId });
      throw error;
    }
  }

  /**
   * 编辑文档
   */
  override async editDocument(documentId: string): Promise<void> {
    try {
      const selector = `.document-list-item[data-document-id="${documentId}"] .document-edit-button`;
      await this.click(selector);
      await this.wait(1000); // 等待导航完成
    } catch (error) {
      this.logger.error('编辑文档失败', { error, documentId });
      throw error;
    }
  }

  /**
   * 删除文档
   */
  override async deleteDocument(documentId: string): Promise<void> {
    try {
      const selector = `.document-list-item[data-document-id="${documentId}"] .document-delete-button`;
      await this.click(selector);

      // 等待确认对话框
      await this.waitForElement('.document-delete-confirm-dialog', { timeout: 5000 });

      // 确认删除
      await this.click('.document-delete-confirm-button');
      await this.wait(1000); // 等待删除完成

    } catch (error) {
      this.logger.error('删除文档失败', { error, documentId });
      throw error;
    }
  }

  /**
   * 下载文档
   */
  override async downloadDocument(documentId: string): Promise<void> {
    try {
      const selector = `.document-list-item[data-document-id="${documentId}"] .document-download-button`;
      await this.click(selector);
      await this.wait(2000); // 等待下载开始
    } catch (error) {
      this.logger.error('下载文档失败', { error, documentId });
      throw error;
    }
  }

  /**
   * 共享文档
   */
  override async shareDocument(documentId: string): Promise<void> {
    try {
      const selector = `.document-list-item[data-document-id="${documentId}"] .document-share-button`;
      await this.click(selector);

      // 等待共享对话框
      await this.waitForElement('.document-share-dialog', { timeout: 5000 });

      // 这里可以添加共享逻辑
      // await this.fill('#share-email', 'email@example.com');
      // await this.click('#share-button');

    } catch (error) {
      this.logger.error('共享文档失败', { error, documentId });
      throw error;
    }
  }

  /**
   * 批量操作
   */
  override async bulkEdit(): Promise<void> {
    try {
      await this.click('#document-bulk-edit-button');
      await this.wait(1000); // 等待批量编辑模式
    } catch (error) {
      this.logger.error('进入批量编辑模式失败', { error });
      throw error;
    }
  }

  /**
   * 批量删除
   */
  override async bulkDelete(): Promise<void> {
    try {
      await this.click('#document-bulk-delete-button');

      // 等待确认对话框
      await this.waitForElement('.document-bulk-delete-confirm-dialog', { timeout: 5000 });

      // 确认删除
      await this.click('.document-bulk-delete-confirm-button');
      await this.wait(1000); // 等待删除完成

    } catch (error) {
      this.logger.error('批量删除文档失败', { error });
      throw error;
    }
  }

  /**
   * 上传文档
   */
  override async uploadDocument(): Promise<void> {
    try {
      await this.click('#document-upload-button');
      await this.wait(1000); // 等待上传对话框
    } catch (error) {
      this.logger.error('打开文档上传对话框失败', { error });
      throw error;
    }
  }

  /**
   * 获取文档统计信息
   */
  override async getDocumentStatistics(): Promise<{
    total: number;
    totalSize: number;
    byType: Record<string, number>;
    byClient: Record<string, number>;
    byStatus: Record<string, number>;
    recentUploads: number;
  }> {
    try {
      const statistics = await this.executeScript(`
        (function() {
          return {
            total: parseInt(document.querySelector('#document-total-count')?.textContent || '0'),
            totalSize: document.querySelector('#document-total-size')?.getAttribute('data-size') || '0',
            byType: JSON.parse(document.querySelector('#document-by-type-stats')?.getAttribute('data-stats') || '{}'),
            byClient: JSON.parse(document.querySelector('#document-by-client-stats')?.getAttribute('data-stats') || '{}'),
            byStatus: JSON.parse(document.querySelector('#document-by-status-stats')?.getAttribute('data-stats') || '{}'),
            recentUploads: parseInt(document.querySelector('#document-recent-uploads')?.textContent || '0')
          };
        })();
      `);

      return {
        total: statistics.total || 0,
        totalSize: parseInt(statistics.totalSize) || 0,
        byType: statistics.byType || {},
        byClient: statistics.byClient || {},
        byStatus: statistics.byStatus || {},
        recentUploads: statistics.recentUploads || 0
      };

    } catch (error) {
      this.logger.error('获取文档统计信息失败', { error });
      throw error;
    }
  }

  /**
   * 导出文档列表
   */
  override async exportDocuments(format: 'csv' | 'excel' | 'pdf' = 'csv'): Promise<void> {
    try {
      const exportButton = format === 'excel' ? '#document-export-excel' :
                           format === 'pdf' ? '#document-export-pdf' : '#document-export-csv';

      await this.click(exportButton);
      await this.wait(2000); // 等待导出开始
    } catch (error) {
      this.logger.error('导出文档列表失败', { error, format });
      throw error;
    }
  }

  /**
   * 验证文档列表页面
   */
  override async validateDocumentListPage(): Promise<{
    valid: boolean;
    missingElements: string[];
    availableElements: string[];
  }> {
    const requiredElements = [
      '#document-search',
      '#document-type-filter',
      '#document-client-filter',
      '#document-case-filter',
      '#document-status-filter',
      '#document-uploaded-by-filter',
      '#document-date-start',
      '#document-date-end',
      '#document-apply-filters',
      '#document-clear-filters',
      '#document-sort-button',
      '#document-upload-button',
      '#document-bulk-edit-button',
      '#document-bulk-delete-button',
      '#document-export-csv',
      '#document-export-excel',
      '#document-export-pdf',
      '.document-list-item'
    ];

    const missingElements: string[] | undefined = undefined;
    const availableElements: string[] | undefined = undefined;

    for (const selector of requiredElements) {
      const isPresent = await this.isVisible(selector);
      if (isPresent) {
        availableElements.push(selector);
      } else {
        missingElements.push(selector);
      }
    }

    return {
      valid: missingElements.length === 0,
      missingElements,
      availableElements
    };
  }
}