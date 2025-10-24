/**
 * 发票管理页面
 */

import { BasePageObject } from '../core/base-page-object';
import { Logger } from '../core/logger';

export interface InvoiceFilters {
  status?: 'draft' | 'sent' | 'paid' | 'overdue' | 'cancelled';
  client?: string;
  case?: string;
  dateRange?: {
    startDate: string;
    endDate: string;
  };
  amountRange?: {
    min: number;
    max: number;
  };
  tags?: string[];
}

export interface InvoiceSortOptions {
  field: 'invoiceNumber' | 'issueDate' | 'dueDate' | 'amount' | 'status' | 'client';
  order: 'asc' | 'desc';
}

export interface InvoiceSearchOptions {
  query?: string;
  filters?: InvoiceFilters;
  sortOptions?: InvoiceSortOptions;
}

export interface InvoiceItem {
  id: string;
  invoiceNumber: string;
  title: string;
  client: string;
  clientName: string;
  case?: string;
  caseName?: string;
  issueDate: string;
  dueDate: string;
  amount: number;
  currency: string;
  status: 'draft' | 'sent' | 'paid' | 'overdue' | 'cancelled';
  description?: string;
  tags: string[];
  paidAmount?: number;
  paidDate?: string;
  createdAt: string;
  updatedAt: string;
  createdBy: string;
}

export interface InvoiceDetail extends InvoiceItem {
  items: InvoiceItemDetail[];
  subtotal: number;
  taxRate: number;
  taxAmount: number;
  totalAmount: number;
  notes?: string;
  paymentTerms: string;
  bankInfo: {
    bankName: string;
    accountNumber: string;
    accountName: string;
    swiftCode?: string;
  };
  attachments: InvoiceAttachment[];
  comments: InvoiceComment[];
  history: InvoiceHistory[];
}

export interface InvoiceItemDetail {
  id: string;
  description: string;
  quantity: number;
  unitPrice: number;
  amount: number;
  category: string;
  caseRef?: string;
  timeEntries?: string[];
}

export interface InvoiceAttachment {
  id: string;
  fileName: string;
  fileSize: number;
  fileType: string;
  uploadedAt: string;
  uploadedBy: string;
}

export interface InvoiceComment {
  id: string;
  content: string;
  author: string;
  timestamp: string;
  type: 'internal' | 'client';
}

export interface InvoiceHistory {
  id: string;
  timestamp: string;
  action: string;
  user: string;
  details?: string;
}

export interface InvoiceFormData {
  title: string;
  client: string;
  case?: string;
  issueDate: string;
  dueDate: string;
  items: InvoiceItemDetail[];
  taxRate: number;
  notes?: string;
  paymentTerms: string;
  tags: string[];
  sendImmediately: boolean;
}

export interface InvoiceTemplate {
  id: string;
  name: string;
  description: string;
  items: InvoiceItemDetail[];
  taxRate: number;
  notes?: string;
  paymentTerms: string;
  isDefault: boolean;
  createdAt: string;
}

export class InvoicePage extends BasePageObject {
  private baseUrl: string;

  constructor(config: { baseUrl: string; defaultTimeout?: number; screenshotOnFailure?: boolean }, logger?: Logger) {
    super(config, this.selectors, logger);
    this.baseUrl = config.baseUrl;
  }

  /**
   * 导航到发票列表页面
   */
  override async navigateToInvoiceList(): Promise<void> {
    await this.navigate(`${this.baseUrl}/finance/invoices`);
  }

  /**
   * 导航到创建发票页面
   */
  override async navigateToCreateInvoice(): Promise<void> {
    await this.navigate(`${this.baseUrl}/finance/invoices/create`);
  }

  /**
   * 导航到发票详情页面
   */
  override async navigateToInvoiceDetail(invoiceId: string): Promise<void> {
    await this.navigate(`${this.baseUrl}/finance/invoices/${invoiceId}`);
  }

  /**
   * 搜索发票
   */
  override async searchInvoices(options: InvoiceSearchOptions): Promise<InvoiceItem[]> {
    try {
      if (options.query) {
        await this.fill('#invoice-search-input', options.query);
      }

      if (options.filters) {
        await this.applyInvoiceFilters(options.filters);
      }

      if (options.sortOptions) {
        await this.sortInvoices(options.sortOptions);
      }

      await this.click('#invoice-search-button');
      await this.waitForElement('.invoice-list-container', { timeout: 5000 });

      const invoices = await this.getInvoiceList();
      this.logger.info('搜索发票完成', { query: options.query, count: invoices.length });
      return invoices;

    } catch (error) {
      this.logger.error('搜索发票失败', { error, options });
      throw error;
    }
  }

  /**
   * 应用发票过滤器
   */
  override async applyInvoiceFilters(filters: InvoiceFilters): Promise<void> {
    try {
      // 展开过滤器面板
      const filterPanel = await this.isVisible('.invoice-filter-panel');
      if (!filterPanel) {
        await this.click('#invoice-filter-toggle');
      }

      // 状态过滤
      if (filters.status) {
        await this.selectOption('#invoice-status-filter', [filters.status]);
      }

      // 客户过滤
      if (filters.client) {
        await this.selectOption('#invoice-client-filter', [filters.client]);
      }

      // 案件过滤
      if (filters.case) {
        await this.selectOption('#invoice-case-filter', [filters.case]);
      }

      // 日期范围
      if (filters.dateRange) {
        await this.fill('#invoice-date-start', filters.dateRange.startDate);
        await this.fill('#invoice-date-end', filters.dateRange.endDate);
      }

      // 金额范围
      if (filters.amountRange) {
        await this.fill('#invoice-amount-min', filters.amountRange.min.toString());
        await this.fill('#invoice-amount-max', filters.amountRange.max.toString());
      }

      // 标签过滤
      if (filters.tags && filters.tags.length > 0) {
        await this.fill('#invoice-tags-filter', filters.tags.join(', '));
      }

    } catch (error) {
      this.logger.error('应用发票过滤器失败', { error, filters });
      throw error;
    }
  }

  /**
   * 排序发票
   */
  override async sortInvoices(sortOptions: InvoiceSortOptions): Promise<void> {
    try {
      await this.click('#invoice-sort-button');
      await this.wait(500);

      // 选择排序字段
      const fieldOption = `#invoice-sort-field-${sortOptions.field}`;
      await this.click(fieldOption);

      // 选择排序顺序
      const orderOption = `#invoice-sort-order-${sortOptions.order}`;
      await this.click(orderOption);

      await this.click('#invoice-sort-apply');
      await this.wait(1000);

    } catch (error) {
      this.logger.error('排序发票失败', { error, sortOptions });
      throw error;
    }
  }

  /**
   * 获取发票列表
   */
  override async getInvoiceList(): Promise<InvoiceItem[]> {
    try {
      const invoiceElements = await this.executeScript(`
        (function() {
          const items = document.querySelectorAll('.invoice-item');
          return Array.from(items).map(item => {
            return {
              id: item.getAttribute('data-id') || '',
              invoiceNumber: item.getAttribute('data-invoice-number') || '',
              title: item.querySelector('.invoice-title')?.gettextContent?.().trim() || '',
              client: item.getAttribute('data-client-id') || '',
              clientName: item.querySelector('.invoice-client-name')?.gettextContent?.().trim() || '',
              case: item.getAttribute('data-case-id') || '',
              caseName: item.querySelector('.invoice-case-name')?.gettextContent?.().trim() || '',
              issueDate: item.getAttribute('data-issue-date') || '',
              dueDate: item.getAttribute('data-due-date') || '',
              amount: parseFloat(item.getAttribute('data-amount') || '0'),
              currency: item.getAttribute('data-currency') || '',
              status: item.getAttribute('data-status') || '',
              description: item.querySelector('.invoice-description')?.gettextContent?.().trim() || '',
              tags: item.getAttribute('data-tags')?.split(',').filter(tag => tag.length > 0) || [],
              paidAmount: parseFloat(item.getAttribute('data-paid-amount') || '0'),
              paidDate: item.getAttribute('data-paid-date') || '',
              createdAt: item.getAttribute('data-created-at') || '',
              updatedAt: item.getAttribute('data-updated-at') || '',
              createdBy: item.getAttribute('data-created-by') || ''
            };
          });
        })()
      `);

      return invoiceElements;

    } catch (error) {
      this.logger.error('获取发票列表失败', { error });
      throw error;
    }
  }

  /**
   * 创建发票
   */
  override async createInvoice(data: InvoiceFormData): Promise<string> {
    try {
      await this.navigateToCreateInvoice();
      await this.wait(2000);

      // 填充基本信息
      await this.fill('#invoice-title', data.title);
      await this.selectOption('#invoice-client', [data.client]);

      if (data.case) {
        await this.selectOption('#invoice-case', [data.case]);
      }

      await this.fill('#invoice-issue-date', data.issueDate);
      await this.fill('#invoice-due-date', data.dueDate);

      // 添加发票项目
      for (const item of data.items) {
        await this.addInvoiceItem(item);
      }

      // 设置税率
      await this.fill('#invoice-tax-rate', data.taxRate.toString());

      // 其他信息
      if (data.notes) {
        await this.fill('#invoice-notes', data.notes);
      }

      await this.fill('#invoice-payment-terms', data.paymentTerms);

      // 标签
      if (data.tags && data.tags.length > 0) {
        await this.fill('#invoice-tags', data.tags.join(', '));
      }

      // 创建发票
      await this.click('#invoice-create-button');
      await this.waitForElement('.invoice-detail-header', { timeout: 10000 });

      // 获取发票ID
      const invoiceId = await this.executeScript(`
        return window.location.pathname.split('/').pop();
      `);

      this.logger.info('创建发票成功', { invoiceId, title: data.title });
      return invoiceId;

    } catch (error) {
      this.logger.error('创建发票失败', { error, data });
      throw error;
    }
  }

  /**
   * 添加发票项目
   */
  override async addInvoiceItem(item: InvoiceItemDetail): Promise<void> {
    try {
      await this.click('#invoice-add-item-button');
      await this.wait(1000);

      // 填充项目信息
      const lastItem = await this.executeScript(`
        return document.querySelectorAll('.invoice-item-row').length - 1;
      `);

      await this.fill(`#invoice-item-description-${lastItem}`, item.description);
      await this.fill(`#invoice-item-quantity-${lastItem}`, item.quantity.toString());
      await this.fill(`#invoice-item-unit-price-${lastItem}`, item.unitPrice.toString());
      await this.selectOption(`#invoice-item-category-${lastItem}`, [item.category]);

      if (item.caseRef) {
        await this.selectOption(`#invoice-item-case-${lastItem}`, [item.caseRef]);
      }

    } catch (error) {
      this.logger.error('添加发票项目失败', { error, item });
      throw error;
    }
  }

  /**
   * 获取发票详情
   */
  override async getInvoiceDetail(invoiceId?: string): Promise<InvoiceDetail> {
    try {
      if (invoiceId) {
        await this.navigateToInvoiceDetail(invoiceId);
      }

      const detail = await this.executeScript(`
        (function() {
          const items = document.querySelectorAll('.invoice-detail-item');
          const invoiceItems = Array.from(document.querySelectorAll('.invoice-item-detail')).map(item => {
            return {
              id: item.getAttribute('data-id') || '',
              description: item.querySelector('.item-description')?.gettextContent?.().trim() || '',
              quantity: parseInt(item.getAttribute('data-quantity') || '0'),
              unitPrice: parseFloat(item.getAttribute('data-unit-price') || '0'),
              amount: parseFloat(item.getAttribute('data-amount') || '0'),
              category: item.getAttribute('data-category') || '',
              caseRef: item.getAttribute('data-case-ref') || '',
              timeEntries: item.getAttribute('data-time-entries')?.split(',').filter(id => id.length > 0) || []
            };
          });

          const attachments = Array.from(document.querySelectorAll('.invoice-attachment')).forEach(attachment => {
            return {
              id: attachment.getAttribute('data-id') || '',
              fileName: attachment.querySelector('.attachment-name')?.gettextContent?.().trim() || '',
              fileSize: parseInt(attachment.getAttribute('data-size') || '0'),
              fileType: attachment.getAttribute('data-type') || '',
              uploadedAt: attachment.getAttribute('data-uploaded-at') || '',
              uploadedBy: attachment.getAttribute('data-uploaded-by') || ''
            };
          });

          const comments = Array.from(document.querySelectorAll('.invoice-comment')).forEach(comment => {
            return {
              id: comment.getAttribute('data-id') || '',
              content: comment.querySelector('.comment-content')?.gettextContent?.().trim() || '',
              author: comment.querySelector('.comment-author')?.gettextContent?.().trim() || '',
              timestamp: comment.getAttribute('data-timestamp') || '',
              type: comment.getAttribute('data-type') || 'internal'
            };
          });

          const history = Array.from(document.querySelectorAll('.invoice-history-item')).forEach(item => {
            return {
              id: item.getAttribute('data-id') || '',
              timestamp: item.getAttribute('data-timestamp') || '',
              action: item.querySelector('.history-action')?.gettextContent?.().trim() || '',
              user: item.querySelector('.history-user')?.gettextContent?.().trim() || '',
              details: item.querySelector('.history-details')?.gettextContent?.().trim() || ''
            };
          });

          return {
            id: document.getElementById('invoice-id')?.getAttribute('value') || '',
            invoiceNumber: document.getElementById('invoice-number')?.gettextContent?.().trim() || '',
            title: document.getElementById('invoice-title')?.gettextContent?.().trim() || '',
            client: document.getElementById('invoice-client-id')?.getAttribute('value') || '',
            clientName: document.getElementById('invoice-client-name')?.gettextContent?.().trim() || '',
            case: document.getElementById('invoice-case-id')?.getAttribute('value') || '',
            caseName: document.getElementById('invoice-case-name')?.gettextContent?.().trim() || '',
            issueDate: document.getElementById('invoice-issue-date')?.gettextContent?.().trim() || '',
            dueDate: document.getElementById('invoice-due-date')?.gettextContent?.().trim() || '',
            amount: parseFloat(document.getElementById('invoice-amount')?.getAttribute('data-value') || '0'),
            currency: document.getElementById('invoice-currency')?.gettextContent?.().trim() || '',
            status: document.getElementById('invoice-status')?.getAttribute('data-status') || '',
            description: document.getElementById('invoice-description')?.gettextContent?.().trim() || '',
            tags: document.getElementById('invoice-tags')?.getAttribute('data-tags')?.split(',').filter(tag => tag.length > 0) || [],
            paidAmount: parseFloat(document.getElementById('invoice-paid-amount')?.getAttribute('data-value') || '0'),
            paidDate: document.getElementById('invoice-paid-date')?.gettextContent?.().trim() || '',
            createdAt: document.getElementById('invoice-created-at')?.gettextContent?.().trim() || '',
            updatedAt: document.getElementById('invoice-updated-at')?.gettextContent?.().trim() || '',
            createdBy: document.getElementById('invoice-created-by')?.gettextContent?.().trim() || '',
            items: invoiceItems,
            subtotal: parseFloat(document.getElementById('invoice-subtotal')?.getAttribute('data-value') || '0'),
            taxRate: parseFloat(document.getElementById('invoice-tax-rate')?.getAttribute('data-value') || '0'),
            taxAmount: parseFloat(document.getElementById('invoice-tax-amount')?.getAttribute('data-value') || '0'),
            totalAmount: parseFloat(document.getElementById('invoice-total-amount')?.getAttribute('data-value') || '0'),
            notes: document.getElementById('invoice-notes')?.gettextContent?.().trim() || '',
            paymentTerms: document.getElementById('invoice-payment-terms')?.gettextContent?.().trim() || '',
            bankInfo: {
              bankName: document.getElementById('invoice-bank-name')?.gettextContent?.().trim() || '',
              accountNumber: document.getElementById('invoice-account-number')?.gettextContent?.().trim() || '',
              accountName: document.getElementById('invoice-account-name')?.gettextContent?.().trim() || '',
              swiftCode: document.getElementById('invoice-swift-code')?.gettextContent?.().trim() || ''
            },
            attachments: attachments,
            comments: comments,
            history: history
          };
        })()
      `);

      return detail;

    } catch (error) {
      this.logger.error('获取发票详情失败', { error, invoiceId });
      throw error;
    }
  }

  /**
   * 发送发票
   */
  override async sendInvoice(invoiceId: string, sendOptions: {
    method: 'email' | 'download' | 'print';
    recipients?: string[];
    subject?: string;
    message?: string;
  }): Promise<void> {
    try {
      await this.navigateToInvoiceDetail(invoiceId);

      if (sendOptions.method === 'email') {
        await this.click('#invoice-send-email-button');
        await this.waitForElement('.invoice-send-modal', { timeout: 5000 });

        if (sendOptions.recipients && sendOptions.recipients.length > 0) {
          await this.fill('#invoice-email-recipients', sendOptions.recipients.join(', '));
        }

        if (sendOptions.subject) {
          await this.fill('#invoice-email-subject', sendOptions.subject);
        }

        if (sendOptions.message) {
          await this.fill('#invoice-email-message', sendOptions.message);
        }

        await this.click('#invoice-send-email-confirm');
        await this.waitForElement('.invoice-sent-confirmation', { timeout: 10000 });

      } else if (sendOptions.method === 'download') {
        await this.click('#invoice-download-button');
        await this.wait(3000); // 等待下载

      } else if (sendOptions.method === 'print') {
        await this.click('#invoice-print-button');
        await this.wait(2000); // 等待打印对话框
      }

      this.logger.info('发送发票成功', { invoiceId, method: sendOptions.method });

    } catch (error) {
      this.logger.error('发送发票失败', { error, invoiceId, sendOptions });
      throw error;
    }
  }

  /**
   * 记录付款
   */
  override async recordPayment(invoiceId: string, paymentData: {
    amount: number;
    paymentDate: string;
    paymentMethod: string;
    reference?: string;
    notes?: string;
  }): Promise<void> {
    try {
      await this.navigateToInvoiceDetail(invoiceId);

      await this.click('#invoice-record-payment-button');
      await this.waitForElement('.invoice-payment-modal', { timeout: 5000 });

      await this.fill('#invoice-payment-amount', paymentData.amount.toString());
      await this.fill('#invoice-payment-date', paymentData.paymentDate);
      await this.selectOption('#invoice-payment-method', [paymentData.paymentMethod]);

      if (paymentData.reference) {
        await this.fill('#invoice-payment-reference', paymentData.reference);
      }

      if (paymentData.notes) {
        await this.fill('#invoice-payment-notes', paymentData.notes);
      }

      await this.click('#invoice-payment-confirm');
      await this.waitForElement('.payment-recorded-confirmation', { timeout: 5000 });

      this.logger.info('记录付款成功', { invoiceId, amount: paymentData.amount });

    } catch (error) {
      this.logger.error('记录付款失败', { error, invoiceId, paymentData });
      throw error;
    }
  }

  /**
   * 取消发票
   */
  override async cancelInvoice(invoiceId: string, reason: string): Promise<void> {
    try {
      await this.navigateToInvoiceDetail(invoiceId);

      await this.click('#invoice-cancel-button');
      await this.waitForElement('.invoice-cancel-modal', { timeout: 5000 });

      await this.fill('#invoice-cancel-reason', reason);
      await this.click('#invoice-cancel-confirm');
      await this.waitForElement('.invoice-cancelled-confirmation', { timeout: 5000 });

      this.logger.info('取消发票成功', { invoiceId, reason });

    } catch (error) {
      this.logger.error('取消发票失败', { error, invoiceId, reason });
      throw error;
    }
  }

  /**
   * 复制发票
   */
  override async duplicateInvoice(invoiceId: string): Promise<string> {
    try {
      await this.navigateToInvoiceDetail(invoiceId);

      await this.click('#invoice-duplicate-button');
      await this.waitForElement('.invoice-create-form', { timeout: 5000 });

      const newInvoiceId = await this.executeScript(`
        return window.location.pathname.split('/').pop();
      `);

      this.logger.info('复制发票成功', { originalId: invoiceId, newId: newInvoiceId });
      return newInvoiceId;

    } catch (error) {
      this.logger.error('复制发票失败', { error, invoiceId });
      throw error;
    }
  }

  /**
   * 使用模板创建发票
   */
  override async createFromTemplate(templateId: string, overrides: {
    client?: string;
    case?: string;
    issueDate?: string;
    dueDate?: string;
    notes?: string;
    tags?: string[];
  } = {}): Promise<string> {
    try {
      await this.navigateToCreateInvoice();

      await this.click('#invoice-use-template-button');
      await this.waitForElement('.invoice-template-selector', { timeout: 5000 });

      await this.click(`#invoice-template-${templateId}`);
      await this.click('#invoice-template-apply');
      await this.wait(2000);

      // 应用覆盖设置
      if (overrides.client) {
        await this.fill('#invoice-client', overrides.client);
      }

      if (overrides.case) {
        await this.fill('#invoice-case', overrides.case);
      }

      if (overrides.issueDate) {
        await this.fill('#invoice-issue-date', overrides.issueDate);
      }

      if (overrides.dueDate) {
        await this.fill('#invoice-due-date', overrides.dueDate);
      }

      if (overrides.notes) {
        await this.fill('#invoice-notes', overrides.notes);
      }

      if (overrides.tags && overrides.tags.length > 0) {
        await this.fill('#invoice-tags', overrides.tags.join(', '));
      }

      await this.click('#invoice-create-button');
      await this.waitForElement('.invoice-detail-header', { timeout: 10000 });

      const invoiceId = await this.executeScript(`
        return window.location.pathname.split('/').pop();
      `);

      this.logger.info('从模板创建发票成功', { templateId, invoiceId });
      return invoiceId;

    } catch (error) {
      this.logger.error('从模板创建发票失败', { error, templateId, overrides });
      throw error;
    }
  }

  /**
   * 获取发票统计
   */
  override async getInvoiceStatistics(): Promise<{
    totalInvoices: number;
    totalAmount: number;
    paidAmount: number;
    outstandingAmount: number;
    overdueAmount: number;
    byStatus: Record<string, number>;
    byMonth: Array<{ month: string; count: number; amount: number }>;
    byClient: Record<string, { count: number; amount: number }>;
    topClients: Array<{ client: string; amount: number; count: number }>;
  }> {
    try {
      const statistics = await this.executeScript(`
        (function() {
          return {
            totalInvoices: parseInt(document.getElementById('invoice-total-count')?.gettextContent?.().trim() || '0'),
            totalAmount: parseFloat(document.getElementById('invoice-total-amount')?.gettextContent?.().trim() || '0'),
            paidAmount: parseFloat(document.getElementById('invoice-paid-amount')?.gettextContent?.().trim() || '0'),
            outstandingAmount: parseFloat(document.getElementById('invoice-outstanding-amount')?.gettextContent?.().trim() || '0'),
            overdueAmount: parseFloat(document.getElementById('invoice-overdue-amount')?.gettextContent?.().trim() || '0'),
            byStatus: JSON.parse(document.getElementById('invoice-stats-by-status')?.textContent || '{}'),
            byMonth: JSON.parse(document.getElementById('invoice-stats-by-month')?.textContent || '[]'),
            byClient: JSON.parse(document.getElementById('invoice-stats-by-client')?.textContent || '{}'),
            topClients: JSON.parse(document.getElementById('invoice-top-clients')?.textContent || '[]')
          };
        })()
      `);

      return statistics;

    } catch (error) {
      this.logger.error('获取发票统计失败', { error });
      throw error;
    }
  }

  /**
   * 验证发票列表页面
   */
  override async validateInvoiceListPage(): Promise<{
    valid: boolean;
    missingElements: string[];
    availableElements: string[];
  }> {
    const requiredElements = [
      '#invoice-search-input',
      '#invoice-search-button',
      '#invoice-filter-toggle',
      '#invoice-sort-button',
      '#invoice-create-button',
      '.invoice-list-container',
      '.invoice-item',
      '#invoice-stats-container',
      '#invoice-export-button'
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

  /**
   * 验证发票详情页面
   */
  override async validateInvoiceDetailPage(): Promise<{
    valid: boolean;
    missingElements: string[];
    availableElements: string[];
  }> {
    const requiredElements = [
      '.invoice-detail-header',
      '#invoice-title',
      '#invoice-client-name',
      '#invoice-issue-date',
      '#invoice-due-date',
      '#invoice-amount',
      '#invoice-status',
      '#invoice-items-table',
      '#invoice-send-email-button',
      '#invoice-download-button',
      '#invoice-print-button',
      '#invoice-record-payment-button',
      '#invoice-edit-button',
      '#invoice-duplicate-button',
      '#invoice-cancel-button',
      '.invoice-comments-section',
      '.invoice-history-section'
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