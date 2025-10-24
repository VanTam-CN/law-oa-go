/**
 * 财务报告页面
 */

import { BasePageObject } from '../core/base-page-object';
import { Logger } from '../core/logger';

export interface ReportFilters {
  reportType?: string;
  dateRange?: {
    startDate: string;
    endDate: string;
  };
  period?: 'month' | 'quarter' | 'year' | 'custom';
  client?: string;
  case?: string;
  employee?: string;
  department?: string;
  currency?: string;
  tags?: string[];
}

export interface ReportSortOptions {
  field: 'date' | 'amount' | 'client' | 'case' | 'employee' | 'category';
  order: 'asc' | 'desc';
}

export interface ReportSearchOptions {
  query?: string;
  filters?: ReportFilters;
  sortOptions?: ReportSortOptions;
}

export interface ReportItem {
  id: string;
  title: string;
  reportType: string;
  description: string;
  generatedAt: string;
  generatedBy: string;
  period: string;
  dateRange: {
    startDate: string;
    endDate: string;
  };
  format: 'pdf' | 'excel' | 'csv' | 'json';
  size: number;
  downloadUrl?: string;
  isScheduled: boolean;
  schedule?: ReportSchedule;
  tags: string[];
  status: 'generating' | 'ready' | 'failed';
  errorMessage?: string;
  dataSummary?: ReportDataSummary;
}

export interface ReportSchedule {
  id: string;
  frequency: 'daily' | 'weekly' | 'monthly' | 'quarterly' | 'yearly';
  nextRunDate: string;
  isActive: boolean;
  recipients: string[];
  format: 'pdf' | 'excel' | 'csv';
  filters: ReportFilters;
}

export interface ReportDataSummary {
  totalRecords: number;
  totalAmount: number;
  currency: string;
  byCategory: Record<string, number>;
  byClient: Record<string, number>;
  byMonth: Array<{ month: string; amount: number }>;
  keyMetrics: Record<string, number>;
}

export interface FinancialReportConfig {
  reportType: 'income_statement' | 'balance_sheet' | 'cash_flow' | 'profit_loss' | 'revenue' | 'expenses' | 'invoices' | 'payments' | 'tax' | 'custom';
  title: string;
  description?: string;
  dateRange: {
    startDate: string;
    endDate: string;
  };
  format: 'pdf' | 'excel' | 'csv';
  includeCharts: boolean;
  sections: ReportSection[];
  filters: ReportFilters;
  groupBy?: string;
  sortBy?: ReportSortOptions;
  currency?: string;
}

export interface ReportSection {
  id: string;
  title: string;
  type: 'table' | 'chart' | 'summary' | 'text';
  config: ReportSectionConfig;
  isVisible: boolean;
  order: number;
}

export interface ReportSectionConfig {
  dataFields: string[];
  chartType?: 'bar' | 'line' | 'pie' | 'doughnut' | 'area' | 'scatter';
  aggregation?: 'sum' | 'average' | 'count' | 'min' | 'max';
  groupBy?: string;
  filters?: Record<string, any>;
}

export interface ReportTemplate {
  id: string;
  name: string;
  description: string;
  reportType: string;
  config: FinancialReportConfig;
  isPublic: boolean;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
  usageCount: number;
}

export interface ReportExportOptions {
  format: 'pdf' | 'excel' | 'csv' | 'json';
  includeCharts: boolean;
  includeRawData: boolean;
  pageOrientation?: 'portrait' | 'landscape';
  paperSize?: 'A4' | 'A3' | 'Letter';
  customFields?: string[];
}

export class FinanceReportPage extends BasePageObject {
  private baseUrl: string;

  constructor(config: { baseUrl: string; defaultTimeout?: number; screenshotOnFailure?: boolean }, logger?: Logger) {
    super(config, this.selectors, logger);
    this.baseUrl = config.baseUrl;
  }

  /**
   * 导航到财务报告页面
   */
  override async navigateToFinanceReports(): Promise<void> {
    await this.navigate(`${this.baseUrl}/finance/reports`);
  }

  /**
   * 导航到创建报告页面
   */
  override async navigateToCreateReport(): Promise<void> {
    await this.navigate(`${this.baseUrl}/finance/reports/create`);
  }

  /**
   * 导航到报告详情页面
   */
  override async navigateToReportDetail(reportId: string): Promise<void> {
    await this.navigate(`${this.baseUrl}/finance/reports/${reportId}`);
  }

  /**
   * 搜索报告
   */
  override async searchReports(options: ReportSearchOptions): Promise<ReportItem[]> {
    try {
      if (options.query) {
        await this.fill('#report-search-input', options.query);
      }

      if (options.filters) {
        await this.applyReportFilters(options.filters);
      }

      if (options.sortOptions) {
        await this.sortReports(options.sortOptions);
      }

      await this.click('#report-search-button');
      await this.waitForElement('.report-list-container', { timeout: 5000 });

      const reports = await this.getReportList();
      this.logger.info('搜索报告完成', { query: options.query, count: reports.length });
      return reports;

    } catch (error) {
      this.logger.error('搜索报告失败', { error, options });
      throw error;
    }
  }

  /**
   * 应用报告过滤器
   */
  override async applyReportFilters(filters: ReportFilters): Promise<void> {
    try {
      // 展开过滤器面板
      const filterPanel = await this.isVisible('.report-filter-panel');
      if (!filterPanel) {
        await this.click('#report-filter-toggle');
      }

      // 报告类型过滤
      if (filters.reportType) {
        await this.selectOption('#report-type-filter', [filters.reportType]);
      }

      // 期间过滤
      if (filters.period) {
        await this.selectOption('#report-period-filter', [filters.period]);
      }

      // 日期范围
      if (filters.dateRange) {
        await this.fill('#report-date-start', filters.dateRange.startDate);
        await this.fill('#report-date-end', filters.dateRange.endDate);
      }

      // 客户过滤
      if (filters.client) {
        await this.selectOption('#report-client-filter', [filters.client]);
      }

      // 案件过滤
      if (filters.case) {
        await this.selectOption('#report-case-filter', [filters.case]);
      }

      // 员工过滤
      if (filters.employee) {
        await this.selectOption('#report-employee-filter', [filters.employee]);
      }

      // 部门过滤
      if (filters.department) {
        await this.selectOption('#report-department-filter', [filters.department]);
      }

      // 货币过滤
      if (filters.currency) {
        await this.selectOption('#report-currency-filter', [filters.currency]);
      }

      // 标签过滤
      if (filters.tags && filters.tags.length > 0) {
        await this.fill('#report-tags-filter', filters.tags.join(', '));
      }

    } catch (error) {
      this.logger.error('应用报告过滤器失败', { error, filters });
      throw error;
    }
  }

  /**
   * 排序报告
   */
  override async sortReports(sortOptions: ReportSortOptions): Promise<void> {
    try {
      await this.click('#report-sort-button');
      await this.wait(500);

      // 选择排序字段
      const fieldOption = `#report-sort-field-${sortOptions.field}`;
      await this.click(fieldOption);

      // 选择排序顺序
      const orderOption = `#report-sort-order-${sortOptions.order}`;
      await this.click(orderOption);

      await this.click('#report-sort-apply');
      await this.wait(1000);

    } catch (error) {
      this.logger.error('排序报告失败', { error, sortOptions });
      throw error;
    }
  }

  /**
   * 获取报告列表
   */
  override async getReportList(): Promise<ReportItem[]> {
    try {
      const reportElements = await this.executeScript(`
        (function() {
          const items = document.querySelectorAll('.report-item');
          return Array.from(items).map(item => {
            return {
              id: item.getAttribute('data-id') || '',
              title: item.querySelector('.report-title')?.gettextContent?.().trim() || '',
              reportType: item.getAttribute('data-report-type') || '',
              description: item.querySelector('.report-description')?.gettextContent?.().trim() || '',
              generatedAt: item.getAttribute('data-generated-at') || '',
              generatedBy: item.querySelector('.report-generated-by')?.gettextContent?.().trim() || '',
              period: item.getAttribute('data-period') || '',
              dateRange: {
                startDate: item.getAttribute('data-start-date') || '',
                endDate: item.getAttribute('data-end-date') || ''
              },
              format: item.getAttribute('data-format') || '',
              size: parseInt(item.getAttribute('data-size') || '0'),
              downloadUrl: item.getAttribute('data-download-url') || '',
              isScheduled: item.getAttribute('data-is-scheduled') === 'true',
              schedule: item.getAttribute('data-schedule') ? JSON.parse(item.getAttribute('data-schedule') || '{}') : undefined,
              tags: item.getAttribute('data-tags')?.split(',').filter(tag => tag.length > 0) || [],
              status: item.getAttribute('data-status') || '',
              errorMessage: item.querySelector('.report-error-message')?.gettextContent?.().trim() || '',
              dataSummary: item.getAttribute('data-summary') ? JSON.parse(item.getAttribute('data-summary') || '{}') : undefined
            };
          });
        })()
      `);

      return reportElements;

    } catch (error) {
      this.logger.error('获取报告列表失败', { error });
      throw error;
    }
  }

  /**
   * 创建财务报告
   */
  override async createFinancialReport(config: FinancialReportConfig): Promise<string> {
    try {
      await this.navigateToCreateReport();
      await this.wait(2000);

      // 基本配置
      await this.selectOption('#report-type', [config.reportType]);
      await this.fill('#report-title', config.title);

      if (config.description) {
        await this.fill('#report-description', config.description);
      }

      // 日期范围
      await this.fill('#report-date-start', config.dateRange.startDate);
      await this.fill('#report-date-end', config.dateRange.endDate);

      // 输出格式
      await this.selectOption('#report-format', [config.format]);

      // 图表配置
      if (config.includeCharts) {
        await this.click('#report-include-charts');
      }

      // 货币设置
      if (config.currency) {
        await this.selectOption('#report-currency', [config.currency]);
      }

      // 分组设置
      if (config.groupBy) {
        await this.selectOption('#report-group-by', [config.groupBy]);
      }

      // 排序设置
      if (config.sortBy) {
        await this.click('#report-advanced-sort');
        await this.wait(500);

        await this.selectOption('#report-sort-field', [config.sortBy.field]);
        await this.selectOption('#report-sort-order', [config.sortBy.order]);
      }

      // 配置报告段
      for (const section of config.sections) {
        await this.addReportSection(section);
      }

      // 应用过滤器
      if (config.filters) {
        await this.applyReportFilters(config.filters);
      }

      // 生成报告
      await this.click('#report-generate-button');
      await this.waitForElement('.report-generation-progress', { timeout: 5000 });

      // 等待报告生成完成
      await this.waitForElement('.report-generation-complete', { timeout: 60000 });

      // 获取报告ID
      const reportId = await this.executeScript(`
        return window.location.pathname.split('/').pop();
      `);

      this.logger.info('创建财务报告成功', { reportId, reportType: config.reportType });
      return reportId;

    } catch (error) {
      this.logger.error('创建财务报告失败', { error, config });
      throw error;
    }
  }

  /**
   * 添加报告段
   */
  override async addReportSection(section: ReportSection): Promise<void> {
    try {
      await this.click('#report-add-section-button');
      await this.wait(1000);

      // 配置段信息
      const lastIndex = await this.executeScript(`
        return document.querySelectorAll('.report-section-row').length - 1;
      `);

      await this.fill(`#report-section-title-${lastIndex}`, section.title);
      await this.selectOption(`#report-section-type-${lastIndex}`, [section.type]);

      // 配置段类型特定设置
      if (section.type === 'chart' && section.config.chartType) {
        await this.selectOption(`#report-section-chart-type-${lastIndex}`, [section.config.chartType]);
      }

      if (section.config.aggregation) {
        await this.selectOption(`#report-section-aggregation-${lastIndex}`, [section.config.aggregation]);
      }

      if (section.config.groupBy) {
        await this.fill(`#report-section-group-by-${lastIndex}`, section.config.groupBy);
      }

      // 设置可见性和排序
      await this.click(`#report-section-visible-${lastIndex}`);
      await this.fill(`#report-section-order-${lastIndex}`, section.order.toString());

    } catch (error) {
      this.logger.error('添加报告段失败', { error, section });
      throw error;
    }
  }

  /**
   * 导出报告
   */
  override async exportReport(reportId: string, options: ReportExportOptions): Promise<void> {
    try {
      await this.navigateToReportDetail(reportId);

      await this.click('#report-export-button');
      await this.waitForElement('.report-export-modal', { timeout: 5000 });

      // 导出选项配置
      await this.selectOption('#export-format', [options.format]);

      if (options.includeCharts) {
        await this.click('#export-include-charts');
      }

      if (options.includeRawData) {
        await this.click('#export-include-raw-data');
      }

      if (options.pageOrientation) {
        await this.selectOption('#export-page-orientation', [options.pageOrientation]);
      }

      if (options.paperSize) {
        await this.selectOption('#export-paper-size', [options.paperSize]);
      }

      // 自定义字段
      if (options.customFields && options.customFields.length > 0) {
        await this.fill('#export-custom-fields', options.customFields.join(', '));
      }

      await this.click('#export-confirm-button');
      await this.waitForElement('.export-initiated', { timeout: 5000 });

      this.logger.info('导出报告成功', { reportId, format: options.format });

    } catch (error) {
      this.logger.error('导出报告失败', { error, reportId, options });
      throw error;
    }
  }

  /**
   * 获取报告详情
   */
  override async getReportDetail(reportId?: string): Promise<ReportItem> {
    try {
      if (reportId) {
        await this.navigateToReportDetail(reportId);
      }

      const detail = await this.executeScript(`
        (function() {
          return {
            id: document.getElementById('report-id')?.gettextContent?.().trim() || '',
            title: document.getElementById('report-title')?.gettextContent?.().trim() || '',
            reportType: document.getElementById('report-type')?.gettextContent?.().trim() || '',
            description: document.getElementById('report-description')?.gettextContent?.().trim() || '',
            generatedAt: document.getElementById('report-generated-at')?.gettextContent?.().trim() || '',
            generatedBy: document.getElementById('report-generated-by')?.gettextContent?.().trim() || '',
            period: document.getElementById('report-period')?.gettextContent?.().trim() || '',
            dateRange: {
              startDate: document.getElementById('report-start-date')?.gettextContent?.().trim() || '',
              endDate: document.getElementById('report-end-date')?.gettextContent?.().trim() || ''
            },
            format: document.getElementById('report-format')?.gettextContent?.().trim() || '',
            size: parseInt(document.getElementById('report-size')?.gettextContent?.().trim() || '0'),
            downloadUrl: document.getElementById('report-download-url')?.getAttribute('href') || '',
            isScheduled: document.getElementById('report-is-scheduled')?.gettextContent?.().trim() === 'true',
            schedule: document.getElementById('report-schedule') ? JSON.parse(document.getElementById('report-schedule')?.textContent || '{}') : undefined,
            tags: document.getElementById('report-tags')?.gettextContent?.().trim().split(',').filter(tag => tag.length > 0) || [],
            status: document.getElementById('report-status')?.gettextContent?.().trim() || '',
            errorMessage: document.getElementById('report-error-message')?.gettextContent?.().trim() || '',
            dataSummary: document.getElementById('report-data-summary') ? JSON.parse(document.getElementById('report-data-summary')?.textContent || '{}') : undefined
          };
        })()
      `);

      return detail;

    } catch (error) {
      this.logger.error('获取报告详情失败', { error, reportId });
      throw error;
    }
  }

  /**
   * 使用模板创建报告
   */
  override async createFromTemplate(templateId: string, overrides: {
    title?: string;
    dateRange?: {
      startDate: string;
      endDate: string;
    };
    filters?: ReportFilters;
    format?: 'pdf' | 'excel' | 'csv';
  } = {}): Promise<string> {
    try {
      await this.navigateToCreateReport();

      await this.click('#report-use-template-button');
      await this.waitForElement('.report-template-selector', { timeout: 5000 });

      await this.click(`#report-template-${templateId}`);
      await this.click('#report-template-apply');
      await this.wait(2000);

      // 应用覆盖设置
      if (overrides.title) {
        await this.fill('#report-title', overrides.title);
      }

      if (overrides.dateRange) {
        await this.fill('#report-date-start', overrides.dateRange.startDate);
        await this.fill('#report-date-end', overrides.dateRange.endDate);
      }

      if (overrides.filters) {
        await this.applyReportFilters(overrides.filters);
      }

      if (overrides.format) {
        await this.selectOption('#report-format', [overrides.format]);
      }

      await this.click('#report-generate-button');
      await this.waitForElement('.report-generation-complete', { timeout: 60000 });

      const reportId = await this.executeScript(`
        return window.location.pathname.split('/').pop();
      `);

      this.logger.info('从模板创建报告成功', { templateId, reportId });
      return reportId;

    } catch (error) {
      this.logger.error('从模板创建报告失败', { error, templateId, overrides });
      throw error;
    }
  }

  /**
   * 获取报告模板
   */
  override async getReportTemplates(): Promise<ReportTemplate[]> {
    try {
      const templates = await this.executeScript(`
        (function() {
          return Array.from(document.querySelectorAll('.report-template-item')).map(item => {
            return {
              id: item.getAttribute('data-id') || '',
              name: item.querySelector('.template-name')?.gettextContent?.().trim() || '',
              description: item.querySelector('.template-description')?.gettextContent?.().trim() || '',
              reportType: item.getAttribute('data-report-type') || '',
              config: JSON.parse(item.getAttribute('data-config') || '{}'),
              isPublic: item.getAttribute('data-is-public') === 'true',
              createdBy: item.getAttribute('data-created-by') || '',
              createdAt: item.getAttribute('data-created-at') || '',
              updatedAt: item.getAttribute('data-updated-at') || '',
              usageCount: parseInt(item.getAttribute('data-usage-count') || '0')
            };
          });
        })()
      `);

      return templates;

    } catch (error) {
      this.logger.error('获取报告模板失败', { error });
      throw error;
    }
  }

  /**
   * 计划报告
   */
  override async scheduleReport(reportId: string, schedule: Omit<ReportSchedule, 'id'>): Promise<string> {
    try {
      await this.navigateToReportDetail(reportId);

      await this.click('#report-schedule-button');
      await this.waitForElement('.report-schedule-modal', { timeout: 5000 });

      // 计划配置
      await this.selectOption('#schedule-frequency', [schedule.frequency]);
      await this.fill('#schedule-next-run', schedule.nextRunDate);

      if (schedule.isActive) {
        await this.click('#schedule-is-active');
      }

      // 收件人
      if (schedule.recipients && schedule.recipients.length > 0) {
        await this.fill('#schedule-recipients', schedule.recipients.join(', '));
      }

      // 格式设置
      await this.selectOption('#schedule-format', [schedule.format]);

      // 高级过滤器
      await this.click('#schedule-advanced-filters');
      await this.wait(500);

      if (schedule.filters) {
        await this.applyReportFilters(schedule.filters);
      }

      await this.click('#schedule-confirm-button');
      await this.waitForElement('.schedule-created', { timeout: 5000 });

      const scheduleId = await this.executeScript(`
        return document.querySelector('.schedule-id')?.gettextContent?.().trim() || '';
      `);

      this.logger.info('计划报告成功', { reportId, scheduleId, frequency: schedule.frequency });

      return scheduleId;

    } catch (error) {
      this.logger.error('计划报告失败', { error, reportId, schedule });
      throw error;
    }
  }

  /**
   * 删除报告
   */
  override async deleteReport(reportId: string): Promise<void> {
    try {
      await this.navigateToReportDetail(reportId);

      await this.click('#report-delete-button');
      await this.waitForElement('.report-delete-modal', { timeout: 5000 });

      await this.click('#delete-confirm-button');
      await this.waitForElement('.report-deleted-confirmation', { timeout: 5000 });

      this.logger.info('删除报告成功', { reportId });

    } catch (error) {
      this.logger.error('删除报告失败', { error, reportId });
      throw error;
    }
  }

  /**
   * 获取报告统计数据
   */
  override async getReportStatistics(): Promise<{
    totalReports: number;
    totalSize: number;
    byType: Record<string, number>;
    byFormat: Record<string, number>;
    byStatus: Record<string, number>;
    recentGenerations: Array<{ date: string; count: number }>;
    topUsers: Array<{ user: string; count: number }>;
    scheduledReports: number;
  }> {
    try {
      const statistics = await this.executeScript(`
        (function() {
          return {
            totalReports: parseInt(document.getElementById('report-total-count')?.gettextContent?.().trim() || '0'),
            totalSize: parseFloat(document.getElementById('report-total-size')?.gettextContent?.().trim() || '0'),
            byType: JSON.parse(document.getElementById('report-stats-by-type')?.textContent || '{}'),
            byFormat: JSON.parse(document.getElementById('report-stats-by-format')?.textContent || '{}'),
            byStatus: JSON.parse(document.getElementById('report-stats-by-status')?.textContent || '{}'),
            recentGenerations: JSON.parse(document.getElementById('report-recent-generations')?.textContent || '[]'),
            topUsers: JSON.parse(document.getElementById('report-top-users')?.textContent || '[]'),
            scheduledReports: parseInt(document.getElementById('report-scheduled-count')?.gettextContent?.().trim() || '0')
          };
        })()
      `);

      return statistics;

    } catch (error) {
      this.logger.error('获取报告统计数据失败', { error });
      throw error;
    }
  }

  /**
   * 验证报告列表页面
   */
  override async validateReportListPage(): Promise<{
    valid: boolean;
    missingElements: string[];
    availableElements: string[];
  }> {
    const requiredElements = [
      '#report-search-input',
      '#report-search-button',
      '#report-filter-toggle',
      '#report-sort-button',
      '#report-create-button',
      '.report-list-container',
      '.report-item',
      '#report-stats-container',
      '#report-templates-button',
      '#report-export-button'
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
   * 验证报告详情页面
   */
  override async validateReportDetailPage(): Promise<{
    valid: boolean;
    missingElements: string[];
    availableElements: string[];
  }> {
    const requiredElements = [
      '.report-detail-header',
      '#report-title',
      '#report-type',
      '#report-generated-at',
      '#report-generated-by',
      '#report-date-range',
      '#report-format',
      '#report-size',
      '#report-status',
      '#report-preview-section',
      '#report-export-button',
      '#report-download-button',
      '#report-schedule-button',
      '#report-delete-button',
      '.report-data-summary',
      '.report-charts-section'
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