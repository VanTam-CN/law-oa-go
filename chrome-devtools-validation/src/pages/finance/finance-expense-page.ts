/**
 * 费用管理页面
 */

import { BasePageObject } from '../core/base-page-object';
import { Logger } from '../core/logger';

export interface ExpenseFilters {
  status?: 'pending' | 'approved' | 'rejected' | 'paid' | 'cancelled';
  category?: string;
  employee?: string;
  project?: string;
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
  paymentMethod?: string;
}

export interface ExpenseSortOptions {
  field: 'expenseDate' | 'amount' | 'category' | 'status' | 'employee' | 'submitDate';
  order: 'asc' | 'desc';
}

export interface ExpenseSearchOptions {
  query?: string;
  filters?: ExpenseFilters;
  sortOptions?: ExpenseSortOptions;
}

export interface ExpenseItem {
  id: string;
  expenseNumber: string;
  title: string;
  description: string;
  category: string;
  amount: number;
  currency: string;
  expenseDate: string;
  submitDate: string;
  status: 'pending' | 'approved' | 'rejected' | 'paid' | 'cancelled';
  employee: string;
  employeeName: string;
  project?: string;
  projectName?: string;
  client?: string;
  clientName?: string;
  case?: string;
  caseName?: string;
  paymentMethod: string;
  receiptAttached: boolean;
  tags: string[];
  receiptCount: number;
  approver?: string;
  approverName?: string;
  approvedDate?: string;
  rejectionReason?: string;
  paidDate?: string;
  createdAt: string;
  updatedAt: string;
}

export interface ExpenseDetail extends ExpenseItem {
  items: ExpenseItemDetail[];
  receipts: ExpenseReceipt[];
  comments: ExpenseComment[];
  history: ExpenseHistory[];
  approvalWorkflow: ExpenseApprovalWorkflow;
  reimbursement?: ExpenseReimbursement;
}

export interface ExpenseItemDetail {
  id: string;
  description: string;
  category: string;
  amount: number;
  quantity: number;
  unitPrice: number;
  date: string;
  project?: string;
  client?: string;
  case?: string;
  notes?: string;
}

export interface ExpenseReceipt {
  id: string;
  fileName: string;
  fileSize: number;
  fileType: string;
  uploadedAt: string;
  uploadedBy: string;
  receiptDate: string;
  merchant: string;
  amount: number;
  currency: string;
  extractedData?: {
    merchant?: string;
    date?: string;
    amount?: number;
    taxAmount?: number;
    category?: string;
  };
}

export interface ExpenseComment {
  id: string;
  content: string;
  author: string;
  authorName: string;
  timestamp: string;
  type: 'internal' | 'public';
}

export interface ExpenseHistory {
  id: string;
  timestamp: string;
  action: string;
  user: string;
  userName: string;
  details?: string;
}

export interface ExpenseApprovalWorkflow {
  currentStep: number;
  totalSteps: number;
  steps: ExpenseApprovalStep[];
  status: 'pending' | 'in_progress' | 'approved' | 'rejected';
}

export interface ExpenseApprovalStep {
  step: number;
  type: 'manager' | 'finance' | 'partner' | 'custom';
  approver: string;
  approverName: string;
  status: 'pending' | 'approved' | 'rejected';
  actionDate?: string;
  comments?: string;
  required: boolean;
}

export interface ExpenseReimbursement {
  id: string;
  reimbursementNumber: string;
  status: 'pending' | 'processed' | 'paid' | 'cancelled';
  totalAmount: number;
  processedDate?: string;
  paidDate?: string;
  paymentMethod: string;
  paymentReference?: string;
  notes?: string;
}

export interface ExpenseFormData {
  title: string;
  description?: string;
  expenseDate: string;
  items: Omit<ExpenseItemDetail, 'id'>[];
  paymentMethod: string;
  project?: string;
  client?: string;
  case?: string;
  tags: string[];
  submitImmediately: boolean;
}

export interface ExpensePolicy {
  id: string;
  name: string;
  description: string;
  rules: ExpensePolicyRule[];
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface ExpensePolicyRule {
  id: string;
  category: string;
  maxAmount: number;
  requiresReceipt: boolean;
  requiresApproval: boolean;
  approverLevel: string;
  conditions: string[];
}

export interface ExpenseReportData {
  employee: string;
  dateRange: {
    startDate: string;
    endDate: string;
  };
  expenses: ExpenseItem[];
  totalAmount: number;
  currency: string;
  status: 'draft' | 'submitted' | 'approved' | 'rejected' | 'processed';
  submittedDate?: string;
  approvedDate?: string;
  processedDate?: string;
  notes?: string;
}

export class ExpensePage extends BasePageObject {
  private baseUrl: string;

  constructor(config: { baseUrl: string; defaultTimeout?: number; screenshotOnFailure?: boolean }, logger?: Logger) {
    super(config, this.selectors, logger);
    this.baseUrl = config.baseUrl;
  }

  /**
   * 导航到费用列表页面
   */
  override async navigateToExpenseList(): Promise<void> {
    await this.navigate(`${this.baseUrl}/finance/expenses`);
  }

  /**
   * 导航到创建费用页面
   */
  override async navigateToCreateExpense(): Promise<void> {
    await this.navigate(`${this.baseUrl}/finance/expenses/create`);
  }

  /**
   * 导航到费用详情页面
   */
  override async navigateToExpenseDetail(expenseId: string): Promise<void> {
    await this.navigate(`${this.baseUrl}/finance/expenses/${expenseId}`);
  }

  /**
   * 搜索费用
   */
  override async searchExpenses(options: ExpenseSearchOptions): Promise<ExpenseItem[]> {
    try {
      if (options.query) {
        await this.fill('#expense-search-input', options.query);
      }

      if (options.filters) {
        await this.applyExpenseFilters(options.filters);
      }

      if (options.sortOptions) {
        await this.sortExpenses(options.sortOptions);
      }

      await this.click('#expense-search-button');
      await this.waitForElement('.expense-list-container', { timeout: 5000 });

      const expenses = await this.getExpenseList();
      this.logger.info('搜索费用完成', { query: options.query, count: expenses.length });
      return expenses;

    } catch (error) {
      this.logger.error('搜索费用失败', { error, options });
      throw error;
    }
  }

  /**
   * 应用费用过滤器
   */
  override async applyExpenseFilters(filters: ExpenseFilters): Promise<void> {
    try {
      // 展开过滤器面板
      const filterPanel = await this.isVisible('.expense-filter-panel');
      if (!filterPanel) {
        await this.click('#expense-filter-toggle');
      }

      // 状态过滤
      if (filters.status) {
        await this.selectOption('#expense-status-filter', [filters.status]);
      }

      // 分类过滤
      if (filters.category) {
        await this.selectOption('#expense-category-filter', [filters.category]);
      }

      // 员工过滤
      if (filters.employee) {
        await this.selectOption('#expense-employee-filter', [filters.employee]);
      }

      // 项目过滤
      if (filters.project) {
        await this.selectOption('#expense-project-filter', [filters.project]);
      }

      // 客户过滤
      if (filters.client) {
        await this.selectOption('#expense-client-filter', [filters.client]);
      }

      // 案件过滤
      if (filters.case) {
        await this.selectOption('#expense-case-filter', [filters.case]);
      }

      // 日期范围
      if (filters.dateRange) {
        await this.fill('#expense-date-start', filters.dateRange.startDate);
        await this.fill('#expense-date-end', filters.dateRange.endDate);
      }

      // 金额范围
      if (filters.amountRange) {
        await this.fill('#expense-amount-min', filters.amountRange.min.toString());
        await this.fill('#expense-amount-max', filters.amountRange.max.toString());
      }

      // 标签过滤
      if (filters.tags && filters.tags.length > 0) {
        await this.fill('#expense-tags-filter', filters.tags.join(', '));
      }

      // 支付方式过滤
      if (filters.paymentMethod) {
        await this.selectOption('#expense-payment-method-filter', [filters.paymentMethod]);
      }

    } catch (error) {
      this.logger.error('应用费用过滤器失败', { error, filters });
      throw error;
    }
  }

  /**
   * 排序费用
   */
  override async sortExpenses(sortOptions: ExpenseSortOptions): Promise<void> {
    try {
      await this.click('#expense-sort-button');
      await this.wait(500);

      // 选择排序字段
      const fieldOption = `#expense-sort-field-${sortOptions.field}`;
      await this.click(fieldOption);

      // 选择排序顺序
      const orderOption = `#expense-sort-order-${sortOptions.order}`;
      await this.click(orderOption);

      await this.click('#expense-sort-apply');
      await this.wait(1000);

    } catch (error) {
      this.logger.error('排序费用失败', { error, sortOptions });
      throw error;
    }
  }

  /**
   * 获取费用列表
   */
  override async getExpenseList(): Promise<ExpenseItem[]> {
    try {
      const expenseElements = await this.executeScript(`
        (function() {
          const items = document.querySelectorAll('.expense-item');
          return Array.from(items).map(item => {
            return {
              id: item.getAttribute('data-id') || '',
              expenseNumber: item.getAttribute('data-expense-number') || '',
              title: item.querySelector('.expense-title')?.gettextContent?.().trim() || '',
              description: item.querySelector('.expense-description')?.gettextContent?.().trim() || '',
              category: item.getAttribute('data-category') || '',
              amount: parseFloat(item.getAttribute('data-amount') || '0'),
              currency: item.getAttribute('data-currency') || '',
              expenseDate: item.getAttribute('data-expense-date') || '',
              submitDate: item.getAttribute('data-submit-date') || '',
              status: item.getAttribute('data-status') || '',
              employee: item.getAttribute('data-employee-id') || '',
              employeeName: item.querySelector('.expense-employee-name')?.gettextContent?.().trim() || '',
              project: item.getAttribute('data-project-id') || '',
              projectName: item.querySelector('.expense-project-name')?.gettextContent?.().trim() || '',
              client: item.getAttribute('data-client-id') || '',
              clientName: item.querySelector('.expense-client-name')?.gettextContent?.().trim() || '',
              case: item.getAttribute('data-case-id') || '',
              caseName: item.querySelector('.expense-case-name')?.gettextContent?.().trim() || '',
              paymentMethod: item.getAttribute('data-payment-method') || '',
              receiptAttached: item.getAttribute('data-receipt-attached') === 'true',
              tags: item.getAttribute('data-tags')?.split(',').filter(tag => tag.length > 0) || [],
              receiptCount: parseInt(item.getAttribute('data-receipt-count') || '0'),
              approver: item.getAttribute('data-approver-id') || '',
              approverName: item.querySelector('.expense-approver-name')?.gettextContent?.().trim() || '',
              approvedDate: item.getAttribute('data-approved-date') || '',
              rejectionReason: item.getAttribute('data-rejection-reason') || '',
              paidDate: item.getAttribute('data-paid-date') || '',
              createdAt: item.getAttribute('data-created-at') || '',
              updatedAt: item.getAttribute('data-updated-at') || ''
            };
          });
        })()
      `);

      return expenseElements;

    } catch (error) {
      this.logger.error('获取费用列表失败', { error });
      throw error;
    }
  }

  /**
   * 创建费用
   */
  override async createExpense(data: ExpenseFormData): Promise<string> {
    try {
      await this.navigateToCreateExpense();
      await this.wait(2000);

      // 填充基本信息
      await this.fill('#expense-title', data.title);

      if (data.description) {
        await this.fill('#expense-description', data.description);
      }

      await this.fill('#expense-date', data.expenseDate);
      await this.selectOption('#expense-payment-method', [data.paymentMethod]);

      // 添加费用项目
      for (const item of data.items) {
        await this.addExpenseItem(item);
      }

      // 关联信息
      if (data.project) {
        await this.selectOption('#expense-project', [data.project]);
      }

      if (data.client) {
        await this.selectOption('#expense-client', [data.client]);
      }

      if (data.case) {
        await this.selectOption('#expense-case', [data.case]);
      }

      // 标签
      if (data.tags && data.tags.length > 0) {
        await this.fill('#expense-tags', data.tags.join(', '));
      }

      // 提交或保存草稿
      if (data.submitImmediately) {
        await this.click('#expense-submit-button');
      } else {
        await this.click('#expense-save-draft-button');
      }

      await this.waitForElement('.expense-detail-header', { timeout: 10000 });

      // 获取费用ID
      const expenseId = await this.executeScript(`
        return window.location.pathname.split('/').pop();
      `);

      this.logger.info('创建费用成功', { expenseId, title: data.title, submitImmediately: data.submitImmediately });
      return expenseId;

    } catch (error) {
      this.logger.error('创建费用失败', { error, data });
      throw error;
    }
  }

  /**
   * 添加费用项目
   */
  override async addExpenseItem(item: Omit<ExpenseItemDetail, 'id'>): Promise<void> {
    try {
      await this.click('#expense-add-item-button');
      await this.wait(1000);

      // 填充项目信息
      const lastIndex = await this.executeScript(`
        return document.querySelectorAll('.expense-item-row').length - 1;
      `);

      await this.fill(`#expense-item-description-${lastIndex}`, item.description);
      await this.selectOption(`#expense-item-category-${lastIndex}`, [item.category]);
      await this.fill(`#expense-item-quantity-${lastIndex}`, item.quantity.toString());
      await this.fill(`#expense-item-unit-price-${lastIndex}`, item.unitPrice.toString());
      await this.fill(`#expense-item-date-${lastIndex}`, item.date);

      if (item.project) {
        await this.selectOption(`#expense-item-project-${lastIndex}`, [item.project]);
      }

      if (item.client) {
        await this.selectOption(`#expense-item-client-${lastIndex}`, [item.client]);
      }

      if (item.case) {
        await this.selectOption(`#expense-item-case-${lastIndex}`, [item.case]);
      }

      if (item.notes) {
        await this.fill(`#expense-item-notes-${lastIndex}`, item.notes);
      }

    } catch (error) {
      this.logger.error('添加费用项目失败', { error, item });
      throw error;
    }
  }

  /**
   * 上传收据
   */
  override async uploadReceipts(expenseId: string, receiptPaths: string[]): Promise<void> {
    try {
      await this.navigateToExpenseDetail(expenseId);

      await this.click('#expense-upload-receipt-button');
      await this.waitForElement('.receipt-upload-modal', { timeout: 5000 });

      // 这里需要实现实际的文件上传逻辑
      // 在实际实现中，这可能需要调用特定的文件上传API
      for (const receiptPath of receiptPaths) {
        await this.executeScript(`
          (function() {
            const input = document.createElement('input');
            input.type = 'file';
            input.style.display = 'none';
            input.name = 'receipt-files';
            document.body.appendChild(input);
            input.click();
            return input;
          })();
        `);
        await this.wait(2000);
      }

      await this.click('#receipt-upload-confirm');
      await this.waitForElement('.receipt-upload-success', { timeout: 10000 });

      this.logger.info('上传收据成功', { expenseId, count: receiptPaths.length });

    } catch (error) {
      this.logger.error('上传收据失败', { error, expenseId, receiptPaths });
      throw error;
    }
  }

  /**
   * 提交费用审批
   */
  override async submitForApproval(expenseId: string, message?: string): Promise<void> {
    try {
      await this.navigateToExpenseDetail(expenseId);

      await this.click('#expense-submit-approval-button');
      await this.waitForElement('.expense-submit-modal', { timeout: 5000 });

      if (message) {
        await this.fill('#expense-submit-message', message);
      }

      await this.click('#expense-submit-confirm');
      await this.waitForElement('.expense-submitted-confirmation', { timeout: 5000 });

      this.logger.info('提交费用审批成功', { expenseId });

    } catch (error) {
      this.logger.error('提交费用审批失败', { error, expenseId, message });
      throw error;
    }
  }

  /**
   * 审批费用
   */
  override async approveExpense(expenseId: string, approvalData: {
    action: 'approve' | 'reject';
    comments?: string;
    nextApprover?: string;
  }): Promise<void> {
    try {
      await this.navigateToExpenseDetail(expenseId);

      if (approvalData.action === 'approve') {
        await this.click('#expense-approve-button');
      } else {
        await this.click('#expense-reject-button');
      }

      await this.waitForElement('.expense-approval-modal', { timeout: 5000 });

      if (approvalData.comments) {
        await this.fill('#expense-approval-comments', approvalData.comments);
      }

      if (approvalData.nextApprover && approvalData.action === 'approve') {
        await this.selectOption('#expense-next-approver', [approvalData.nextApprover]);
      }

      await this.click('#expense-approval-confirm');
      await this.waitForElement('.expense-approval-confirmation', { timeout: 5000 });

      this.logger.info('审批费用成功', { expenseId, action: approvalData.action });

    } catch (error) {
      this.logger.error('审批费用失败', { error, expenseId, approvalData });
      throw error;
    }
  }

  /**
   * 获取费用详情
   */
  override async getExpenseDetail(expenseId?: string): Promise<ExpenseDetail> {
    try {
      if (expenseId) {
        await this.navigateToExpenseDetail(expenseId);
      }

      const detail = await this.executeScript(`
        (function() {
          const items = Array.from(document.querySelectorAll('.expense-item-detail')).map(item => {
            return {
              id: item.getAttribute('data-id') || '',
              description: item.querySelector('.item-description')?.gettextContent?.().trim() || '',
              category: item.getAttribute('data-category') || '',
              amount: parseFloat(item.getAttribute('data-amount') || '0'),
              quantity: parseInt(item.getAttribute('data-quantity') || '0'),
              unitPrice: parseFloat(item.getAttribute('data-unit-price') || '0'),
              date: item.getAttribute('data-date') || '',
              project: item.getAttribute('data-project') || '',
              client: item.getAttribute('data-client') || '',
              case: item.getAttribute('data-case') || '',
              notes: item.querySelector('.item-notes')?.gettextContent?.().trim() || ''
            };
          });

          const receipts = Array.from(document.querySelectorAll('.expense-receipt')).forEach(receipt => {
            return {
              id: receipt.getAttribute('data-id') || '',
              fileName: receipt.querySelector('.receipt-name')?.gettextContent?.().trim() || '',
              fileSize: parseInt(receipt.getAttribute('data-size') || '0'),
              fileType: receipt.getAttribute('data-type') || '',
              uploadedAt: receipt.getAttribute('data-uploaded-at') || '',
              uploadedBy: receipt.getAttribute('data-uploaded-by') || '',
              receiptDate: receipt.getAttribute('data-receipt-date') || '',
              merchant: receipt.getAttribute('data-merchant') || '',
              amount: parseFloat(receipt.getAttribute('data-amount') || '0'),
              currency: receipt.getAttribute('data-currency') || '',
              extractedData: receipt.getAttribute('data-extracted-data') ? JSON.parse(receipt.getAttribute('data-extracted-data') || '{}') : undefined
            };
          });

          const comments = Array.from(document.querySelectorAll('.expense-comment')).forEach(comment => {
            return {
              id: comment.getAttribute('data-id') || '',
              content: comment.querySelector('.comment-content')?.gettextContent?.().trim() || '',
              author: comment.getAttribute('data-author') || '',
              authorName: comment.querySelector('.comment-author-name')?.gettextContent?.().trim() || '',
              timestamp: comment.getAttribute('data-timestamp') || '',
              type: comment.getAttribute('data-type') || 'internal'
            };
          });

          const history = Array.from(document.querySelectorAll('.expense-history-item')).forEach(item => {
            return {
              id: item.getAttribute('data-id') || '',
              timestamp: item.getAttribute('data-timestamp') || '',
              action: item.querySelector('.history-action')?.gettextContent?.().trim() || '',
              user: item.getAttribute('data-user') || '',
              userName: item.querySelector('.history-user-name')?.gettextContent?.().trim() || '',
              details: item.querySelector('.history-details')?.gettextContent?.().trim() || ''
            };
          });

          const workflow = {
            currentStep: parseInt(document.getElementById('expense-workflow-current-step')?.gettextContent?.().trim() || '0'),
            totalSteps: parseInt(document.getElementById('expense-workflow-total-steps')?.gettextContent?.().trim() || '0'),
            status: document.getElementById('expense-workflow-status')?.gettextContent?.().trim() || 'pending',
            steps: Array.from(document.querySelectorAll('.expense-approval-step')).map(step => ({
              step: parseInt(step.getAttribute('data-step') || '0'),
              type: step.getAttribute('data-type') || '',
              approver: step.getAttribute('data-approver') || '',
              approverName: step.querySelector('.step-approver-name')?.gettextContent?.().trim() || '',
              status: step.getAttribute('data-status') || '',
              actionDate: step.getAttribute('data-action-date') || '',
              comments: step.querySelector('.step-comments')?.gettextContent?.().trim() || '',
              required: step.getAttribute('data-required') === 'true'
            }))
          };

          const reimbursement = document.getElementById('expense-reimbursement-info') ? {
            id: document.getElementById('reimbursement-id')?.gettextContent?.().trim() || '',
            reimbursementNumber: document.getElementById('reimbursement-number')?.gettextContent?.().trim() || '',
            status: document.getElementById('reimbursement-status')?.gettextContent?.().trim() || '',
            totalAmount: parseFloat(document.getElementById('reimbursement-amount')?.gettextContent?.().trim() || '0'),
            processedDate: document.getElementById('reimbursement-processed-date')?.gettextContent?.().trim() || '',
            paidDate: document.getElementById('reimbursement-paid-date')?.gettextContent?.().trim() || '',
            paymentMethod: document.getElementById('reimbursement-payment-method')?.gettextContent?.().trim() || '',
            paymentReference: document.getElementById('reimbursement-reference')?.gettextContent?.().trim() || '',
            notes: document.getElementById('reimbursement-notes')?.gettextContent?.().trim() || ''
          } : undefined;

          return {
            id: document.getElementById('expense-id')?.gettextContent?.().trim() || '',
            expenseNumber: document.getElementById('expense-number')?.gettextContent?.().trim() || '',
            title: document.getElementById('expense-title')?.gettextContent?.().trim() || '',
            description: document.getElementById('expense-description')?.gettextContent?.().trim() || '',
            category: document.getElementById('expense-category')?.gettextContent?.().trim() || '',
            amount: parseFloat(document.getElementById('expense-amount')?.gettextContent?.().trim() || '0'),
            currency: document.getElementById('expense-currency')?.gettextContent?.().trim() || '',
            expenseDate: document.getElementById('expense-date')?.gettextContent?.().trim() || '',
            submitDate: document.getElementById('expense-submit-date')?.gettextContent?.().trim() || '',
            status: document.getElementById('expense-status')?.gettextContent?.().trim() || '',
            employee: document.getElementById('expense-employee-id')?.gettextContent?.().trim() || '',
            employeeName: document.getElementById('expense-employee-name')?.gettextContent?.().trim() || '',
            project: document.getElementById('expense-project-id')?.gettextContent?.().trim() || '',
            projectName: document.getElementById('expense-project-name')?.gettextContent?.().trim() || '',
            client: document.getElementById('expense-client-id')?.gettextContent?.().trim() || '',
            clientName: document.getElementById('expense-client-name')?.gettextContent?.().trim() || '',
            case: document.getElementById('expense-case-id')?.gettextContent?.().trim() || '',
            caseName: document.getElementById('expense-case-name')?.gettextContent?.().trim() || '',
            paymentMethod: document.getElementById('expense-payment-method')?.gettextContent?.().trim() || '',
            receiptAttached: document.getElementById('expense-receipt-attached')?.gettextContent?.().trim() === 'true',
            tags: document.getElementById('expense-tags')?.gettextContent?.().trim().split(',').filter(tag => tag.length > 0) || [],
            receiptCount: parseInt(document.getElementById('expense-receipt-count')?.gettextContent?.().trim() || '0'),
            approver: document.getElementById('expense-approver-id')?.gettextContent?.().trim() || '',
            approverName: document.getElementById('expense-approver-name')?.gettextContent?.().trim() || '',
            approvedDate: document.getElementById('expense-approved-date')?.gettextContent?.().trim() || '',
            rejectionReason: document.getElementById('expense-rejection-reason')?.gettextContent?.().trim() || '',
            paidDate: document.getElementById('expense-paid-date')?.gettextContent?.().trim() || '',
            createdAt: document.getElementById('expense-created-at')?.gettextContent?.().trim() || '',
            updatedAt: document.getElementById('expense-updated-at')?.gettextContent?.().trim() || '',
            items: items,
            receipts: receipts,
            comments: comments,
            history: history,
            approvalWorkflow: workflow,
            reimbursement: reimbursement
          };
        })()
      `);

      return detail;

    } catch (error) {
      this.logger.error('获取费用详情失败', { error, expenseId });
      throw error;
    }
  }

  /**
   * 创建费用报告
   */
  override async createExpenseReport(data: ExpenseReportData): Promise<string> {
    try {
      await this.navigateToExpenseList();

      await this.click('#expense-create-report-button');
      await this.waitForElement('.expense-report-modal', { timeout: 5000 });

      await this.selectOption('#report-employee', [data.employee]);
      await this.fill('#report-date-start', data.dateRange.startDate);
      await this.fill('#report-date-end', data.dateRange.endDate);

      if (data.notes) {
        await this.fill('#report-notes', data.notes);
      }

      await this.click('#report-create-confirm');
      await this.waitForElement('.report-created-confirmation', { timeout: 5000 });

      const reportId = await this.executeScript(`
        return document.querySelector('.report-id')?.gettextContent?.().trim() || '';
      `);

      this.logger.info('创建费用报告成功', { reportId, employee: data.employee });

      return reportId;

    } catch (error) {
      this.logger.error('创建费用报告失败', { error, data });
      throw error;
    }
  }

  /**
   * 获取费用政策
   */
  override async getExpensePolicies(): Promise<ExpensePolicy[]> {
    try {
      const policies = await this.executeScript(`
        (function() {
          return Array.from(document.querySelectorAll('.expense-policy-item')).map(item => {
            return {
              id: item.getAttribute('data-id') || '',
              name: item.querySelector('.policy-name')?.gettextContent?.().trim() || '',
              description: item.querySelector('.policy-description')?.gettextContent?.().trim() || '',
              isActive: item.getAttribute('data-is-active') === 'true',
              createdAt: item.getAttribute('data-created-at') || '',
              updatedAt: item.getAttribute('data-updated-at') || '',
              rules: Array.from(item.querySelectorAll('.policy-rule')).map(rule => ({
                id: rule.getAttribute('data-id') || '',
                category: rule.getAttribute('data-category') || '',
                maxAmount: parseFloat(rule.getAttribute('data-max-amount') || '0'),
                requiresReceipt: rule.getAttribute('data-requires-receipt') === 'true',
                requiresApproval: rule.getAttribute('data-requires-approval') === 'true',
                approverLevel: rule.getAttribute('data-approver-level') || '',
                conditions: rule.getAttribute('data-conditions')?.split(',').filter(c => c.length > 0) || []
              }))
            };
          });
        })()
      `);

      return policies;

    } catch (error) {
      this.logger.error('获取费用政策失败', { error });
      throw error;
    }
  }

  /**
   * 获取费用统计
   */
  override async getExpenseStatistics(): Promise<{
    totalExpenses: number;
    totalAmount: number;
    pendingAmount: number;
    approvedAmount: number;
    rejectedAmount: number;
    paidAmount: number;
    byCategory: Record<string, number>;
    byEmployee: Record<string, { count: number; amount: number }>;
    byMonth: Array<{ month: string; count: number; amount: number }>;
    byStatus: Record<string, number>;
    averageProcessingTime: number;
  }> {
    try {
      const statistics = await this.executeScript(`
        (function() {
          return {
            totalExpenses: parseInt(document.getElementById('expense-total-count')?.gettextContent?.().trim() || '0'),
            totalAmount: parseFloat(document.getElementById('expense-total-amount')?.gettextContent?.().trim() || '0'),
            pendingAmount: parseFloat(document.getElementById('expense-pending-amount')?.gettextContent?.().trim() || '0'),
            approvedAmount: parseFloat(document.getElementById('expense-approved-amount')?.gettextContent?.().trim() || '0'),
            rejectedAmount: parseFloat(document.getElementById('expense-rejected-amount')?.gettextContent?.().trim() || '0'),
            paidAmount: parseFloat(document.getElementById('expense-paid-amount')?.gettextContent?.().trim() || '0'),
            byCategory: JSON.parse(document.getElementById('expense-stats-by-category')?.textContent || '{}'),
            byEmployee: JSON.parse(document.getElementById('expense-stats-by-employee')?.textContent || '{}'),
            byMonth: JSON.parse(document.getElementById('expense-stats-by-month')?.textContent || '[]'),
            byStatus: JSON.parse(document.getElementById('expense-stats-by-status')?.textContent || '{}'),
            averageProcessingTime: parseFloat(document.getElementById('expense-avg-processing-time')?.gettextContent?.().trim() || '0')
          };
        })()
      `);

      return statistics;

    } catch (error) {
      this.logger.error('获取费用统计失败', { error });
      throw error;
    }
  }

  /**
   * 验证费用列表页面
   */
  override async validateExpenseListPage(): Promise<{
    valid: boolean;
    missingElements: string[];
    availableElements: string[];
  }> {
    const requiredElements = [
      '#expense-search-input',
      '#expense-search-button',
      '#expense-filter-toggle',
      '#expense-sort-button',
      '#expense-create-button',
      '.expense-list-container',
      '.expense-item',
      '#expense-stats-container',
      '#expense-create-report-button',
      '#expense-export-button'
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
   * 验证费用详情页面
   */
  override async validateExpenseDetailPage(): Promise<{
    valid: boolean;
    missingElements: string[];
    availableElements: string[];
  }> {
    const requiredElements = [
      '.expense-detail-header',
      '#expense-title',
      '#expense-employee-name',
      '#expense-date',
      '#expense-amount',
      '#expense-status',
      '#expense-category',
      '#expense-payment-method',
      '#expense-items-table',
      '#expense-receipts-section',
      '#expense-submit-approval-button',
      '#expense-edit-button',
      '#expense-delete-button',
      '.expense-approval-workflow',
      '.expense-comments-section',
      '.expense-history-section'
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