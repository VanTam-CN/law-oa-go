/**
 * 财务管理仪表板页面
 */

import { BasePageObject } from '../core/base-page-object';
import { Logger } from '../core/logger';

export interface FinancialOverview {
  totalRevenue: number;
  totalExpenses: number;
  netProfit: number;
  receivables: number;
  payables: number;
  cashBalance: number;
  monthlyRevenue: number;
  monthlyExpenses: number;
  profitMargin: number;
}

export interface FinancialMetrics {
  revenueGrowth: number;
  expenseGrowth: number;
  profitGrowth: number;
  clientAcquisitionCost: number;
  clientLifetimeValue: number;
  casesPerLawyer: number;
  averageCaseValue: number;
  collectionRate: number;
}

export interface TransactionFilters {
  type?: 'income' | 'expense' | 'all';
  category?: string;
  client?: string;
  case?: string;
  dateRange?: {
    start: Date;
    end: Date;
  };
  amountRange?: {
    min: number;
    max: number;
  };
  status?: 'pending' | 'completed' | 'cancelled' | 'all';
}

export interface TransactionSortOptions {
  field: 'date' | 'amount' | 'type' | 'category' | 'client' | 'status';
  order: 'asc' | 'desc';
}

export interface TransactionItem {
  id: string;
  date: Date;
  type: 'income' | 'expense';
  category: string;
  description: string;
  amount: number;
  client?: string;
  case?: string;
  status: 'pending' | 'completed' | 'cancelled';
  reference?: string;
  attachments?: string[];
  createdAt: Date;
  updatedAt: Date;
}

export interface InvoiceItem {
  id: string;
  invoiceNumber: string;
  client: string;
  case?: string;
  issueDate: Date;
  dueDate: Date;
  amount: number;
  status: 'draft' | 'sent' | 'paid' | 'overdue' | 'cancelled';
  items: Array<{
    description: string;
    quantity: number;
    unitPrice: number;
    amount: number;
  }>;
  subtotal: number;
  tax: number;
  total: number;
  paidAmount: number;
  balance: number;
}

export interface ExpenseItem {
  id: string;
  date: Date;
  category: string;
  description: string;
  amount: number;
  vendor: string;
  receipt?: string;
  approvalStatus: 'pending' | 'approved' | 'rejected';
  reimbursable: boolean;
  reimbursed: boolean;
  notes?: string;
}

export class FinanceDashboardPage extends BasePageObject {
  private baseUrl: string;

  constructor(config: { baseUrl: string; defaultTimeout?: number; screenshotOnFailure?: boolean }, logger?: Logger) {
    super(config, this.selectors, logger);
    this.baseUrl = config.baseUrl;
  }

  /**
   * 导航到财务管理仪表板
   */
  override async navigateToFinanceDashboard(): Promise<void> {
    await this.navigate(`${this.baseUrl}/finance`);
  }

  /**
   * 获取财务概览
   */
  override async getFinancialOverview(): Promise<FinancialOverview> {
    try {
      const overview = await this.executeScript(`
        (function() {
          return {
            totalRevenue: parseFloat(document.querySelector('#total-revenue')?.getAttribute('data-value') || '0'),
            totalExpenses: parseFloat(document.querySelector('#total-expenses')?.getAttribute('data-value') || '0'),
            netProfit: parseFloat(document.querySelector('#net-profit')?.getAttribute('data-value') || '0'),
            receivables: parseFloat(document.querySelector('#receivables')?.getAttribute('data-value') || '0'),
            payables: parseFloat(document.querySelector('#payables')?.getAttribute('data-value') || '0'),
            cashBalance: parseFloat(document.querySelector('#cash-balance')?.getAttribute('data-value') || '0'),
            monthlyRevenue: parseFloat(document.querySelector('#monthly-revenue')?.getAttribute('data-value') || '0'),
            monthlyExpenses: parseFloat(document.querySelector('#monthly-expenses')?.getAttribute('data-value') || '0'),
            profitMargin: parseFloat(document.querySelector('#profit-margin')?.getAttribute('data-value') || '0')
          };
        })();
      `);

      return overview;

    } catch (error) {
      this.logger.error('获取财务概览失败', { error });
      throw error;
    }
  }

  /**
   * 获取财务指标
   */
  override async getFinancialMetrics(): Promise<FinancialMetrics> {
    try {
      const metrics = await this.executeScript(`
        (function() {
          return {
            revenueGrowth: parseFloat(document.querySelector('#revenue-growth')?.getAttribute('data-value') || '0'),
            expenseGrowth: parseFloat(document.querySelector('#expense-growth')?.getAttribute('data-value') || '0'),
            profitGrowth: parseFloat(document.querySelector('#profit-growth')?.getAttribute('data-value') || '0'),
            clientAcquisitionCost: parseFloat(document.querySelector('#client-acquisition-cost')?.getAttribute('data-value') || '0'),
            clientLifetimeValue: parseFloat(document.querySelector('#client-lifetime-value')?.getAttribute('data-value') || '0'),
            casesPerLawyer: parseFloat(document.querySelector('#cases-per-lawyer')?.getAttribute('data-value') || '0'),
            averageCaseValue: parseFloat(document.querySelector('#average-case-value')?.getAttribute('data-value') || '0'),
            collectionRate: parseFloat(document.querySelector('#collection-rate')?.getAttribute('data-value') || '0')
          };
        })();
      `);

      return metrics;

    } catch (error) {
      this.logger.error('获取财务指标失败', { error });
      throw error;
    }
  }

  /**
   * 获取交易记录
   */
  override async getTransactions(filters?: TransactionFilters, sortOptions?: TransactionSortOptions): Promise<TransactionItem[]> {
    try {
      // 应用过滤条件
      if (filters) {
        await this.applyTransactionFilters(filters);
      }

      // 应用排序
      if (sortOptions) {
        await this.sortTransactions(sortOptions);
      }

      const transactions = await this.executeScript(`
        (function() {
          const transactionElements = Array.from(document.querySelectorAll('.transaction-item'));
          return transactionElements.map(item => {
            const id = item.getAttribute('data-transaction-id') || '';
            const dateText = item.querySelector('.transaction-date')?.gettextContent?.().trim() || '';
            const type = item.getAttribute('data-transaction-type') || '';
            const category = item.querySelector('.transaction-category')?.gettextContent?.().trim() || '';
            const description = item.querySelector('.transaction-description')?.gettextContent?.().trim() || '';
            const amountText = item.querySelector('.transaction-amount')?.gettextContent?.().trim() || '0';
            const client = item.querySelector('.transaction-client')?.gettextContent?.().trim();
            const caseText = item.querySelector('.transaction-case')?.gettextContent?.().trim();
            const status = item.getAttribute('data-transaction-status') || '';
            const reference = item.querySelector('.transaction-reference')?.gettextContent?.().trim();
            const createdAtText = item.querySelector('.transaction-created-at')?.gettextContent?.().trim() || '';
            const updatedAtText = item.querySelector('.transaction-updated-at')?.gettextContent?.().trim() || '';

            const parseAmount = (amountText) => {
              const cleanAmount = amountText.replace(/[^\\d.-]/g, '');
              return parseFloat(cleanAmount) || 0;
            };

            const parseDate = (dateText) => {
              const date = new Date(dateText);
              return isNaN(date.getTime()) ? new Date() : date;
            };

            const attachments = Array.from(item.querySelectorAll('.transaction-attachment'))
              .map(attachment => attachment.getAttribute('data-attachment-id') || '');

            return {
              id,
              date: parseDate(dateText),
              type: type as 'income' | 'expense',
              category,
              description,
              amount: parseAmount(amountText),
              client,
              case: caseText,
              status: status as 'pending' | 'completed' | 'cancelled',
              reference,
              attachments,
              createdAt: parseDate(createdAtText),
              updatedAt: parseDate(updatedAtText)
            };
          });
        })();
      `);

      return transactions || [];

    } catch (error) {
      this.logger.error('获取交易记录失败', { error, filters, sortOptions });
      throw error;
    }
  }

  /**
   * 应用交易过滤条件
   */
  private override async applyTransactionFilters(filters: TransactionFilters): Promise<void> {
    try {
      if (filters.type && filters.type !== 'all') {
        await this.selectOption('#transaction-type-filter', [filters.type]);
      }

      if (filters.category) {
        await this.selectOption('#transaction-category-filter', [filters.category]);
      }

      if (filters.client) {
        await this.selectOption('#transaction-client-filter', [filters.client]);
      }

      if (filters.case) {
        await this.selectOption('#transaction-case-filter', [filters.case]);
      }

      if (filters.status && filters.status !== 'all') {
        await this.selectOption('#transaction-status-filter', [filters.status]);
      }

      if (filters.dateRange) {
        const startDate = filters.dateRange.start.toISOString().split('T')[0];
        const endDate = filters.dateRange.end.toISOString().split('T')[0];

        await this.fill('#transaction-date-start', startDate);
        await this.fill('#transaction-date-end', endDate);
      }

      if (filters.amountRange) {
        await this.fill('#transaction-amount-min', filters.amountRange.min.toString());
        await this.fill('#transaction-amount-max', filters.amountRange.max.toString());
      }

      await this.click('#transaction-apply-filters');
      await this.wait(2000); // 等待过滤结果

    } catch (error) {
      this.logger.error('应用交易过滤条件失败', { error, filters });
      throw error;
    }
  }

  /**
   * 排序交易记录
   */
  private override async sortTransactions(sortOptions: TransactionSortOptions): Promise<void> {
    try {
      const sortButtonSelector = '#transaction-sort-button';
      await this.click(sortButtonSelector);

      // 等待排序菜单出现
      await this.waitForElement('.transaction-sort-menu', { timeout: 5000 });

      // 选择排序字段
      const fieldSelector = `.transaction-sort-field[data-field="${sortOptions.field}"]`;
      await this.click(fieldSelector);

      // 选择排序方向
      const orderSelector = `.transaction-sort-order[data-order="${sortOptions.order}"]`;
      await this.click(orderSelector);

      await this.click(sortButtonSelector); // 关闭菜单
      await this.wait(1000); // 等待排序完成

    } catch (error) {
      this.logger.error('排序交易记录失败', { error, sortOptions });
      throw error;
    }
  }

  /**
   * 获取发票列表
   */
  override async getInvoices(): Promise<InvoiceItem[]> {
    try {
      const invoices = await this.executeScript(`
        (function() {
          const invoiceElements = Array.from(document.querySelectorAll('.invoice-item'));
          return invoiceElements.map(item => {
            const id = item.getAttribute('data-invoice-id') || '';
            const invoiceNumber = item.querySelector('.invoice-number')?.gettextContent?.().trim() || '';
            const client = item.querySelector('.invoice-client')?.gettextContent?.().trim() || '';
            const caseText = item.querySelector('.invoice-case')?.gettextContent?.().trim();
            const issueDateText = item.querySelector('.invoice-issue-date')?.gettextContent?.().trim() || '';
            const dueDateText = item.querySelector('.invoice-due-date')?.gettextContent?.().trim() || '';
            const amountText = item.querySelector('.invoice-amount')?.gettextContent?.().trim() || '0';
            const status = item.getAttribute('data-invoice-status') || '';
            const paidAmountText = item.querySelector('.invoice-paid-amount')?.gettextContent?.().trim() || '0';

            const parseAmount = (amountText) => {
              const cleanAmount = amountText.replace(/[^\\d.-]/g, '');
              return parseFloat(cleanAmount) || 0;
            };

            const parseDate = (dateText) => {
              const date = new Date(dateText);
              return isNaN(date.getTime()) ? new Date() : date;
            };

            const invoiceItems = Array.from(item.querySelectorAll('.invoice-item-detail')).map(detail => ({
              description: detail.querySelector('.item-description')?.gettextContent?.().trim() || '',
              quantity: parseInt(detail.querySelector('.item-quantity')?.gettextContent?.().trim() || '0'),
              unitPrice: parseAmount(detail.querySelector('.item-unit-price')?.gettextContent?.().trim() || '0'),
              amount: parseAmount(detail.querySelector('.item-amount')?.gettextContent?.().trim() || '0')
            }));

            const subtotal = invoiceItems.reduce((sum, item) => sum + item.amount, 0);
            const taxRate = 0.1; // 假设税率10%
            const tax = subtotal * taxRate;
            const total = subtotal + tax;
            const paidAmount = parseAmount(paidAmountText);
            const balance = total - paidAmount;

            return {
              id,
              invoiceNumber,
              client,
              case: caseText,
              issueDate: parseDate(issueDateText),
              dueDate: parseDate(dueDateText),
              amount: total,
              status: status as 'draft' | 'sent' | 'paid' | 'overdue' | 'cancelled',
              items: invoiceItems,
              subtotal,
              tax,
              total,
              paidAmount,
              balance
            };
          });
        })();
      `);

      return invoices || [];

    } catch (error) {
      this.logger.error('获取发票列表失败', { error });
      throw error;
    }
  }

  /**
   * 获取费用列表
   */
  override async getExpenses(): Promise<ExpenseItem[]> {
    try {
      const expenses = await this.executeScript(`
        (function() {
          const expenseElements = Array.from(document.querySelectorAll('.expense-item'));
          return expenseElements.map(item => {
            const id = item.getAttribute('data-expense-id') || '';
            const dateText = item.querySelector('.expense-date')?.gettextContent?.().trim() || '';
            const category = item.querySelector('.expense-category')?.gettextContent?.().trim() || '';
            const description = item.querySelector('.expense-description')?.gettextContent?.().trim() || '';
            const amountText = item.querySelector('.expense-amount')?.gettextContent?.().trim() || '0';
            const vendor = item.querySelector('.expense-vendor')?.gettextContent?.().trim() || '';
            const receipt = item.querySelector('.expense-receipt')?.getAttribute('data-receipt-id');
            const approvalStatus = item.getAttribute('data-approval-status') || '';
            const reimbursable = item.querySelector('.expense-reimbursable') !== null;
            const reimbursed = item.querySelector('.expense-reimbursed') !== null;
            const notes = item.querySelector('.expense-notes')?.gettextContent?.().trim();

            const parseAmount = (amountText) => {
              const cleanAmount = amountText.replace(/[^\\d.-]/g, '');
              return parseFloat(cleanAmount) || 0;
            };

            const parseDate = (dateText) => {
              const date = new Date(dateText);
              return isNaN(date.getTime()) ? new Date() : date;
            };

            return {
              id,
              date: parseDate(dateText),
              category,
              description,
              amount: parseAmount(amountText),
              vendor,
              receipt,
              approvalStatus: approvalStatus as 'pending' | 'approved' | 'rejected',
              reimbursable,
              reimbursed,
              notes
            };
          });
        })();
      `);

      return expenses || [];

    } catch (error) {
      this.logger.error('获取费用列表失败', { error });
      throw error;
    }
  }

  /**
   * 添加收入交易
   */
  override async addIncomeTransaction(data: {
    client?: string;
    case?: string;
    amount: number;
    category: string;
    description: string;
    date: Date;
    reference?: string;
  }): Promise<void> {
    try {
      await this.click('#add-income-button');
      await this.waitForElement('#income-transaction-modal', { timeout: 5000 });

      // 填写表单
      if (data.client) {
        await this.selectOption('#income-client', [data.client]);
      }
      if (data.case) {
        await this.selectOption('#income-case', [data.case]);
      }
      await this.fill('#income-amount', data.amount.toString());
      await this.selectOption('#income-category', [data.category]);
      await this.fill('#income-description', data.description);
      await this.fill('#income-date', data.date.toISOString().split('T')[0]);
      if (data.reference) {
        await this.fill('#income-reference', data.reference);
      }

      // 保存
      await this.click('#income-save-button');
      await this.wait(2000); // 等待保存完成

    } catch (error) {
      this.logger.error('添加收入交易失败', { error, data });
      throw error;
    }
  }

  /**
   * 添加支出交易
   */
  override async addExpenseTransaction(data: {
    vendor: string;
    amount: number;
    category: string;
    description: string;
    date: Date;
    reimbursable?: boolean;
    receipt?: string;
  }): Promise<void> {
    try {
      await this.click('#add-expense-button');
      await this.waitForElement('#expense-transaction-modal', { timeout: 5000 });

      // 填写表单
      await this.fill('#expense-vendor', data.vendor);
      await this.fill('#expense-amount', data.amount.toString());
      await this.selectOption('#expense-category', [data.category]);
      await this.fill('#expense-description', data.description);
      await this.fill('#expense-date', data.date.toISOString().split('T')[0]);

      if (data.reimbursable !== undefined) {
        const reimbursableCheckbox = await this.isVisible('#expense-reimbursable');
        const isChecked = await this.executeScript('return document.getElementById("expense-reimbursable").checked;');
        if (data.reimbursable !== isChecked) {
          await this.click('#expense-reimbursable');
        }
      }

      // 保存
      await this.click('#expense-save-button');
      await this.wait(2000); // 等待保存完成

    } catch (error) {
      this.logger.error('添加支出交易失败', { error, data });
      throw error;
    }
  }

  /**
   * 创建发票
   */
  override async createInvoice(data: {
    client: string;
    case?: string;
    items: Array<{
      description: string;
      quantity: number;
      unitPrice: number;
    }>;
    dueDate: Date;
    notes?: string;
  }): Promise<void> {
    try {
      await this.click('#create-invoice-button');
      await this.waitForElement('#invoice-modal', { timeout: 5000 });

      // 填写表单
      await this.selectOption('#invoice-client', [data.client]);
      if (data.case) {
        await this.selectOption('#invoice-case', [data.case]);
      }
      await this.fill('#invoice-due-date', data.dueDate.toISOString().split('T')[0]);
      if (data.notes) {
        await this.fill('#invoice-notes', data.notes);
      }

      // 添加发票项目
      for (const item of data.items) {
        await this.click('#add-invoice-item-button');
        await this.fill('#invoice-item-description', item.description);
        await this.fill('#invoice-item-quantity', item.quantity.toString());
        await this.fill('#invoice-item-unit-price', item.unitPrice.toString());
      }

      // 保存
      await this.click('#invoice-save-button');
      await this.wait(2000); // 等待保存完成

    } catch (error) {
      this.logger.error('创建发票失败', { error, data });
      throw error;
    }
  }

  /**
   * 导出财务报表
   */
  override async exportFinancialReport(type: 'income-statement' | 'balance-sheet' | 'cash-flow' | 'transaction-list' | 'invoice-summary', format: 'pdf' | 'excel' = 'pdf'): Promise<void> {
    try {
      const exportButton = `#export-${type}-${format}`;
      await this.click(exportButton);
      await this.wait(3000); // 等待导出开始
    } catch (error) {
      this.logger.error('导出财务报表失败', { error, type, format });
      throw error;
    }
  }

  /**
   * 验证财务管理仪表板页面
   */
  override async validateFinanceDashboardPage(): Promise<{
    valid: boolean;
    missingElements: string[];
    availableElements: string[];
  }> {
    const requiredElements = [
      '#total-revenue',
      '#total-expenses',
      '#net-profit',
      '#receivables',
      '#payables',
      '#cash-balance',
      '#revenue-growth',
      '#expense-growth',
      '#profit-growth',
      '#client-acquisition-cost',
      '#client-lifetime-value',
      '#cases-per-lawyer',
      '#average-case-value',
      '#collection-rate',
      '#transaction-list',
      '#transaction-type-filter',
      '#transaction-category-filter',
      '#transaction-client-filter',
      '#transaction-case-filter',
      '#transaction-status-filter',
      '#transaction-apply-filters',
      '#transaction-sort-button',
      '#add-income-button',
      '#add-expense-button',
      '#create-invoice-button',
      '#export-income-statement-pdf',
      '#export-income-statement-excel',
      '#export-balance-sheet-pdf',
      '#export-balance-sheet-excel',
      '#export-cash-flow-pdf',
      '#export-cash-flow-excel'
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