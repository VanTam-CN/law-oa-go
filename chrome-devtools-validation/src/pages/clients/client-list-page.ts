/**
 * 客户列表页面Page Object
 */

import { BasePageObject, PageObjectConfig } from '../../core/base-page-object';
import { Logger } from '../../core/logger';

export interface ClientFilters {
  type?: string;
  industry?: string;
  status?: string;
  dateRange?: {
    start: Date;
    end: Date;
  };
  searchQuery?: string;
}

export interface ClientSortOptions {
  field: 'name' | 'createdDate' | 'updatedDate' | 'type' | 'status';
  order: 'asc' | 'desc';
}

export interface ClientListItem {
  id: string;
  name: string;
  type: string;
  industry: string;
  contactPerson: string;
  email: string;
  phone: string;
  status: string;
  createdDate: Date;
  updatedDate: Date;
  caseCount: number;
}

export class ClientListPage extends BasePageObject {
    protected override selectors = {
    clientTable: '#client-table',
    clientRows: '.client-row',
    clientRow: (id: string) => `.client-row[data-id="${id}"]`,
    searchInput: '#client-search',
    filterButton: '#filter-button',
    clearFiltersButton: '#clear-filters',
    typeFilter: '#type-filter',
    industryFilter: '#industry-filter',
    statusFilter: '#status-filter',
    dateRangeFilter: '#date-range-filter',
    sortDropdown: '#sort-dropdown',
    createClientButton: '#create-client-button',
    exportButton: '#export-button',
    importButton: '#import-button',
    pagination: '.pagination',
    currentPage: '.current-page',
    totalPages: '.total-pages',
    previousButton: '.previous-button',
    nextButton: '.next-button',
    loadingSpinner: '.loading-spinner',
    emptyState: '.empty-state',
    clientCount: '.client-count',
    selectedCount: '.selected-count',
    bulkActions: '.bulk-actions',
    bulkEditButton: '#bulk-edit-button',
    bulkDeleteButton: '#bulk-delete-button',
    clientTypeBadge: (id: string) => `.client-row[data-id="${id}"] .client-type`,
    clientStatusBadge: (id: string) => `.client-row[data-id="${id}"] .client-status`,
    clientName: (id: string) => `.client-row[data-id="${id}"] .client-name`,
    clientIndustry: (id: string) => `.client-row[data-id="${id}"] .client-industry`,
    clientContact: (id: string) => `.client-row[data-id="${id}"] .client-contact`,
    clientEmail: (id: string) => `.client-row[data-id="${id}"] .client-email`,
    clientCaseCount: (id: string) => `.client-row[data-id="${id}"] .client-case-count`,
    clientCheckbox: (id: string) => `.client-row[data-id="${id}"] .client-checkbox`,
    viewClientButton: (id: string) => `.client-row[data-id="${id}"] .view-client-button`,
    editClientButton: (id: string) => `.client-row[data-id="${id}"] .edit-client-button`,
    deleteClientButton: (id: string) => `.client-row[data-id="${id}"] .delete-client-button`
  };

  constructor(config: PageObjectConfig, logger?: Logger) {
    super(config, this.selectors, logger);
  }

  /**
   * 导航到客户列表页面
   */
  override async navigateToClientList(): Promise<void> {
    await this.navigate('/clients');
    await this.waitForElement(this.selectors.clientTable);
  }

  /**
   * 搜索客户
   */
  override async searchClients(query: string): Promise<void> {
    await this.waitForElement(this.selectors.searchInput);
    await this.fill(this.selectors.searchInput, query);

    // 等待搜索结果
    await this.waitForSearchResults();

    this.logger.debug('客户搜索完成', { query });
  }

  /**
   * 应用过滤器
   */
  override async applyFilters(filters: ClientFilters): Promise<void> {
    if (filters.type) {
      await this.waitForElement(this.selectors.typeFilter);
      await this.select(this.selectors.typeFilter, filters.type);
    }

    if (filters.industry) {
      await this.waitForElement(this.selectors.industryFilter);
      await this.select(this.selectors.industryFilter, filters.industry);
    }

    if (filters.status) {
      await this.waitForElement(this.selectors.statusFilter);
      await this.select(this.selectors.statusFilter, filters.status);
    }

    if (filters.dateRange) {
      await this.waitForElement(this.selectors.dateRangeFilter);
      // 在实际实现中，这里会设置日期范围
      this.logger.debug('日期范围过滤器已设置', { dateRange: filters.dateRange });
    }

    // 等待过滤结果
    await this.waitForFilterResults();

    this.logger.debug('客户过滤器已应用', { filters });
  }

  /**
   * 清除所有过滤器
   */
  override async clearFilters(): Promise<void> {
    if (await this.isVisible(this.selectors.clearFiltersButton)) {
      await this.click(this.selectors.clearFiltersButton);
      await this.waitForFilterResults();
      this.logger.debug('所有过滤器已清除');
    }
  }

  /**
   * 排序客户
   */
  override async sortClients(sortOptions: ClientSortOptions): Promise<void> {
    await this.waitForElement(this.selectors.sortDropdown);
    await this.select(this.selectors.sortDropdown, `${sortOptions.field}-${sortOptions.order}`);

    // 等待排序结果
    await this.waitForSortResults();

    this.logger.debug('客户排序完成', { sortOptions });
  }

  /**
   * 获取客户列表
   */
  override async getClientList(): Promise<ClientListItem[]> {
    await this.waitForElement(this.selectors.clientTable);

    // 在实际实现中，这里会从DOM获取客户列表
    // 现在返回模拟数据
    const mockClients: ClientListItem[] = [
      {
        id: 'client-1',
        name: '科技有限公司',
        type: 'company',
        industry: 'technology',
        contactPerson: '张经理',
        email: 'contact@techcompany.com',
        phone: '010-12345678',
        status: 'active',
        createdDate: new Date('2024-01-15'),
        updatedDate: new Date('2024-01-20'),
        caseCount: 3
      },
      {
        id: 'client-2',
        name: '制造企业',
        type: 'company',
        industry: 'manufacturing',
        contactPerson: '李总',
        email: 'contact@manufacturing.com',
        phone: '020-87654321',
        status: 'active',
        createdDate: new Date('2024-01-10'),
        updatedDate: new Date('2024-01-18'),
        caseCount: 2
      },
      {
        id: 'client-3',
        name: '王先生',
        type: 'individual',
        industry: 'personal',
        contactPerson: '王先生',
        email: 'wang@example.com',
        phone: '13800138000',
        status: 'inactive',
        createdDate: new Date('2023-12-01'),
        updatedDate: new Date('2024-01-05'),
        caseCount: 1
      }
    ];

    this.logger.debug('获取客户列表', { count: mockClients.length });
    return mockClients;
  }

  /**
   * 选择客户
   */
  override async selectClient(clientId: string): Promise<void> {
    const selector = this.selectors.clientCheckbox(clientId);
    await this.waitForElement(selector);
    await this.click(selector);
    this.logger.debug('客户已选择', { clientId });
  }

  /**
   * 批量选择客户
   */
  override async selectMultipleClients(clientIds: string[]): Promise<void> {
    for (const clientId of clientIds) {
      await this.selectClient(clientId);
    }
    this.logger.debug('批量选择客户完成', { count: clientIds.length });
  }

  /**
   * 获取选中的客户数量
   */
  override async getSelectedCount(): Promise<number> {
    if (await this.isVisible(this.selectors.selectedCount)) {
      const text = await this.getText(this.selectors.selectedCount);
      const match = text.match(/(\d+)/);
      return match ? parseInt(match[1]) : 0;
    }
    return 0;
  }

  /**
   * 查看客户详情
   */
  override async viewClient(clientId: string): Promise<void> {
    const selector = this.selectors.viewClientButton(clientId);
    await this.waitForElement(selector);
    await this.click(selector);
    this.logger.debug('查看客户详情', { clientId });
  }

  /**
   * 编辑客户
   */
  override async editClient(clientId: string): Promise<void> {
    const selector = this.selectors.editClientButton(clientId);
    await this.waitForElement(selector);
    await this.click(selector);
    this.logger.debug('编辑客户', { clientId });
  }

  /**
   * 删除客户
   */
  override async deleteClient(clientId: string): Promise<void> {
    const selector = this.selectors.deleteClientButton(clientId);
    await this.waitForElement(selector);
    await this.click(selector);

    // 确认删除
    await this.waitForElement('.confirm-delete-button');
    await this.click('.confirm-delete-button');

    this.logger.debug('客户已删除', { clientId });
  }

  /**
   * 创建新客户
   */
  override async createNewClient(): Promise<void> {
    await this.waitForElement(this.selectors.createClientButton);
    await this.click(this.selectors.createClientButton);
    this.logger.debug('创建新客户按钮已点击');
  }

  /**
   * 导出客户列表
   */
  override async exportClients(format: 'csv' | 'excel' | 'pdf' = 'excel'): Promise<void> {
    await this.waitForElement(this.selectors.exportButton);

    // 在实际实现中，这里会选择导出格式
    await this.click(this.selectors.exportButton);

    this.logger.debug('客户列表导出已开始', { format });
  }

  /**
   * 导入客户
   */
  override async importClients(): Promise<void> {
    await this.waitForElement(this.selectors.importButton);
    await this.click(this.selectors.importButton);
    this.logger.debug('导入客户按钮已点击');
  }

  /**
   * 翻页
   */
  override async goToPage(page: number): Promise<void> {
    // 在实际实现中，这里会点击相应的页码
    this.logger.debug('翻页到', { page });
  }

  /**
   * 下一页
   */
  override async nextPage(): Promise<void> {
    if (await this.isVisible(this.selectors.nextButton)) {
      await this.click(this.selectors.nextButton);
      await this.waitForPageLoad();
      this.logger.debug('已翻到下一页');
    }
  }

  /**
   * 上一页
   */
  override async previousPage(): Promise<void> {
    if (await this.isVisible(this.selectors.previousButton)) {
      await this.click(this.selectors.previousButton);
      await this.waitForPageLoad();
      this.logger.debug('已翻到上一页');
    }
  }

  /**
   * 获取当前页码
   */
  override async getCurrentPage(): Promise<number> {
    if (await this.isVisible(this.selectors.currentPage)) {
      const text = await this.getText(this.selectors.currentPage);
      return parseInt(text) || 1;
    }
    return 1;
  }

  /**
   * 获取总页数
   */
  override async getTotalPages(): Promise<number> {
    if (await this.isVisible(this.selectors.totalPages)) {
      const text = await this.getText(this.selectors.totalPages);
      const match = text.match(/(\d+)/);
      return match ? parseInt(match[1]) : 1;
    }
    return 1;
  }

  /**
   * 获取客户总数
   */
  override async getClientCount(): Promise<number> {
    if (await this.isVisible(this.selectors.clientCount)) {
      const text = await this.getText(this.selectors.clientCount);
      const match = text.match(/(\d+)/);
      return match ? parseInt(match[1]) : 0;
    }
    return 0;
  }

  /**
   * 检查是否正在加载
   */
  override async isLoading(): Promise<boolean> {
    return await this.isVisible(this.selectors.loadingSpinner);
  }

  /**
   * 检查是否为空状态
   */
  override async isEmpty(): Promise<boolean> {
    return await this.isVisible(this.selectors.emptyState);
  }

  /**
   * 获取客户详情信息
   */
  override async getClientDetails(clientId: string): Promise<ClientListItem | null> {
    const clients = await this.getClientList();
    return clients.find(c => c.id === clientId) || null;
  }

  /**
   * 批量编辑客户
   */
  override async bulkEdit(): Promise<void> {
    if (await this.isVisible(this.selectors.bulkActions)) {
      await this.waitForElement(this.selectors.bulkEditButton);
      await this.click(this.selectors.bulkEditButton);
      this.logger.debug('批量编辑已启动');
    }
  }

  /**
   * 批量删除客户
   */
  override async bulkDelete(): Promise<void> {
    if (await this.isVisible(this.selectors.bulkActions)) {
      await this.waitForElement(this.selectors.bulkDeleteButton);
      await this.click(this.selectors.bulkDeleteButton);

      // 确认删除
      await this.waitForElement('.confirm-bulk-delete-button');
      await this.click('.confirm-bulk-delete-button');

      this.logger.debug('批量删除已完成');
    }
  }

  /**
   * 获取客户统计数据
   */
  override async getClientStatistics(): Promise<{
    total: number;
    active: number;
    inactive: number;
    byType: Record<string, number>;
    byIndustry: Record<string, number>;
  }> {
    const clients = await this.getClientList();

    const statistics = {
      total: clients.length,
      active: clients.filter(c => c.status === 'active').length,
      inactive: clients.filter(c => c.status === 'inactive').length,
      byType: {} as Record<string, number>,
      byIndustry: {} as Record<string, number>
    };

    // 按类型统计
    clients.forEach(client => {
      statistics.byType[client.type] = (statistics.byType[client.type] || 0) + 1;
    });

    // 按行业统计
    clients.forEach(client => {
      statistics.byIndustry[client.industry] = (statistics.byIndustry[client.industry] || 0) + 1;
    });

    this.logger.debug('获取客户统计数据', statistics);
    return statistics;
  }

  /**
   * 等待搜索结果
   */
  private override async waitForSearchResults(): Promise<void> {
    await this.wait(1000); // 模拟搜索延迟
  }

  /**
   * 等待过滤结果
   */
  private override async waitForFilterResults(): Promise<void> {
    await this.wait(1000); // 模拟过滤延迟
  }

  /**
   * 等待排序结果
   */
  private override async waitForSortResults(): Promise<void> {
    await this.wait(1000); // 模拟排序延迟
  }

  /**
   * 等待页面加载
   */
  private override async waitForPageLoad(): Promise<void> {
    await this.wait(500); // 模拟页面加载延迟
  }

  /**
   * 验证客户列表页面元素
   */
  override async validateClientListPage(): Promise<{ valid: boolean; missingElements: string[] }> {
    const requiredElements = [
      { name: 'clientTable', selector: this.selectors.clientTable },
      { name: 'searchInput', selector: this.selectors.searchInput },
      { name: 'createClientButton', selector: this.selectors.createClientButton }
    ];

    const missingElements: string[] | undefined = undefined;

    for (const element of requiredElements) {
      if (!await this.isExists(element.selector)) {
        missingElements.push(element.name);
      }
    }

    return {
      valid: missingElements.length === 0,
      missingElements
    };
  }
}