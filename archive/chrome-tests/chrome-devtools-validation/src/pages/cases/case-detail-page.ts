/**
 * 案件详情页面Page Object
 */

import { BasePageObject, PageObjectConfig } from '../../core/base-page-object';
import { Logger } from '../../core/logger';

export interface CaseDetail {
  id: string;
  title: string;
  caseNumber: string;
  type: string;
  status: string;
  priority: string;
  client: string;
  assignedTo: string;
  description: string;
  estimatedValue: number;
  createdDate: Date;
  updatedDate: Date;
  dueDate?: Date;
  tags: string[];
  milestones: Array<{
    id: string;
    title: string;
    completed: boolean;
    dueDate?: Date;
  }>;
}

export interface CaseUpdateData {
  title?: string;
  description?: string;
  priority?: string;
  status?: string;
  assignedTo?: string;
  dueDate?: Date;
  estimatedValue?: number;
  tags?: string[];
}

export class CaseDetailPage extends BasePageObject {
    protected override selectors = {
    caseDetailContainer: '#case-detail-container',
    caseTitle: '#case-title',
    caseNumber: '#case-number',
    caseType: '#case-type',
    caseStatus: '#case-status',
    casePriority: '#case-priority',
    caseClient: '#case-client',
    caseAssignedTo: '#case-assigned-to',
    caseDescription: '#case-description',
    caseEstimatedValue: '#case-estimated-value',
    caseCreatedDate: '#case-created-date',
    caseUpdatedDate: '#case-updated-date',
    caseDueDate: '#case-due-date',
    caseTags: '.case-tags',
    editButton: '#edit-case-button',
    deleteButton: '#delete-case-button',
    backButton: '#back-button',
    milestonesSection: '#milestones-section',
    documentsSection: '#documents-section',
    timelineSection: '#timeline-section',
    financialSection: '#financial-section',
    addMilestoneButton: '#add-milestone-button',
    addDocumentButton: '#add-document-button',
    addFinancialRecordButton: '#add-financial-record-button',
    milestoneItems: '.milestone-item',
    documentItems: '.document-item',
    timelineItems: '.timeline-item',
    financialItems: '.financial-item',
    loadingSpinner: '.loading-spinner',
    errorMessage: '.error-message',
    successMessage: '.success-message',
    editForm: '#edit-case-form',
    saveButton: '#save-case-button',
    cancelButton: '#cancel-edit-button'
  };

  constructor(config: PageObjectConfig, logger?: Logger) {
    super(config, this.selectors, logger);
  }

  /**
   * 导航到案件详情页面
   */
  override async navigateToCaseDetail(caseId: string): Promise<void> {
    await this.navigate(`/cases/${caseId}`);
    await this.waitForElement(this.selectors.caseDetailContainer);
  }

  /**
   * 获取案件详情
   */
  override async getCaseDetail(): Promise<CaseDetail> {
    await this.waitForElement(this.selectors.caseTitle);

    // 在实际实现中，这里会从DOM获取案件详情
    // 现在返回模拟数据
    const mockCase: CaseDetail = {
      id: 'case-1',
      title: await this.getText(this.selectors.caseTitle) || '合同纠纷案件',
      caseNumber: await this.getText(this.selectors.caseNumber) || '(2024)京01民初123号',
      type: await this.getText(this.selectors.caseType) || 'litigation',
      status: await this.getText(this.selectors.caseStatus) || 'active',
      priority: await this.getText(this.selectors.casePriority) || 'high',
      client: await this.getText(this.selectors.caseClient) || '科技有限公司',
      assignedTo: await this.getText(this.selectors.caseAssignedTo) || '张律师',
      description: await this.getText(this.selectors.caseDescription) || '这是一个关于合同纠纷的案件，涉及金额较大，需要仔细处理。',
      estimatedValue: parseFloat(await this.getText(this.selectors.caseEstimatedValue) || '100000'),
      createdDate: new Date(await this.getText(this.selectors.caseCreatedDate) || '2024-01-15'),
      updatedDate: new Date(await this.getText(this.selectors.caseUpdatedDate) || '2024-01-20'),
      dueDate: await this.getText(this.selectors.caseDueDate) ? new Date(await this.getText(this.selectors.caseDueDate)) : undefined,
      tags: await this.getTags(),
      milestones: await this.getMilestones()
    };

    this.logger.debug('获取案件详情', { caseId: mockCase.id });
    return mockCase;
  }

  /**
   * 获取标签
   */
  private override async getTags(): Promise<string[]> {
    if (await this.isVisible(this.selectors.caseTags)) {
      // 在实际实现中，这里会解析标签元素
      return ['重要', '紧急', '待审核'];
    }
    return [];
  }

  /**
   * 获取里程碑
   */
  private override async getMilestones(): Promise<Array<{
    id: string;
    title: string;
    completed: boolean;
    dueDate?: Date;
  }>> {
    if (await this.isVisible(this.selectors.milestonesSection)) {
      // 在实际实现中，这里会解析里程碑元素
      return [
        {
          id: 'milestone-1',
          title: '案件受理',
          completed: true,
          dueDate: new Date('2024-01-20')
        },
        {
          id: 'milestone-2',
          title: '证据收集',
          completed: false,
          dueDate: new Date('2024-02-01')
        }
      ];
    }
    return [];
  }

  /**
   * 编辑案件
   */
  override async editCase(): Promise<void> {
    await this.waitForElement(this.selectors.editButton);
    await this.click(this.selectors.editButton);

    // 等待编辑表单出现
    await this.waitForElement(this.selectors.editForm);

    this.logger.debug('编辑案件模式已启动');
  }

  /**
   * 更新案件信息
   */
  override async updateCase(updateData: CaseUpdateData): Promise<void> {
    await this.editCase();

    if (updateData.title) {
      await this.waitForElement('#edit-title');
      await this.fill('#edit-title', updateData.title);
    }

    if (updateData.description) {
      await this.waitForElement('#edit-description');
      await this.fill('#edit-description', updateData.description);
    }

    if (updateData.priority) {
      await this.waitForElement('#edit-priority');
      await this.select('#edit-priority', updateData.priority);
    }

    if (updateData.status) {
      await this.waitForElement('#edit-status');
      await this.select('#edit-status', updateData.status);
    }

    if (updateData.assignedTo) {
      await this.waitForElement('#edit-assigned-to');
      await this.select('#edit-assigned-to', updateData.assignedTo);
    }

    if (updateData.dueDate) {
      await this.waitForElement('#edit-due-date');
      await this.fill('#edit-due-date', updateData.dueDate.toISOString().split('T')[0]);
    }

    if (updateData.estimatedValue) {
      await this.waitForElement('#edit-estimated-value');
      await this.fill('#edit-estimated-value', updateData.estimatedValue.toString());
    }

    // 保存更改
    await this.waitForElement(this.selectors.saveButton);
    await this.click(this.selectors.saveButton);

    // 等待保存完成
    await this.waitForSaveComplete();

    this.logger.debug('案件信息已更新', { updateData });
  }

  /**
   * 删除案件
   */
  override async deleteCase(): Promise<void> {
    await this.waitForElement(this.selectors.deleteButton);
    await this.click(this.selectors.deleteButton);

    // 确认删除
    await this.waitForElement('.confirm-delete-button');
    await this.click('.confirm-delete-button');

    this.logger.debug('案件已删除');
  }

  /**
   * 返回案件列表
   */
  override async backToList(): Promise<void> {
    await this.waitForElement(this.selectors.backButton);
    await this.click(this.selectors.backButton);
    this.logger.debug('返回案件列表');
  }

  /**
   * 添加里程碑
   */
  override async addMilestone(milestone: {
    title: string;
    dueDate?: Date;
  }): Promise<void> {
    await this.waitForElement(this.selectors.addMilestoneButton);
    await this.click(this.selectors.addMilestoneButton);

    // 填写里程碑表单
    await this.waitForElement('#milestone-title');
    await this.fill('#milestone-title', milestone.title);

    if (milestone.dueDate) {
      await this.waitForElement('#milestone-due-date');
      await this.fill('#milestone-due-date', milestone.dueDate.toISOString().split('T')[0]);
    }

    // 保存里程碑
    await this.waitForElement('#save-milestone-button');
    await this.click('#save-milestone-button');

    this.logger.debug('里程碑已添加', { title: milestone.title });
  }

  /**
   * 添加文档
   */
  override async addDocument(document: {
    name: string;
    type: string;
    description?: string;
  }): Promise<void> {
    await this.waitForElement(this.selectors.addDocumentButton);
    await this.click(this.selectors.addDocumentButton);

    // 填写文档表单
    await this.waitForElement('#document-name');
    await this.fill('#document-name', document.name);

    await this.waitForElement('#document-type');
    await this.select('#document-type', document.type);

    if (document.description) {
      await this.waitForElement('#document-description');
      await this.fill('#document-description', document.description);
    }

    // 保存文档
    await this.waitForElement('#save-document-button');
    await this.click('#save-document-button');

    this.logger.debug('文档已添加', { name: document.name });
  }

  /**
   * 添加财务记录
   */
  override async addFinancialRecord(record: {
    type: string;
    description: string;
    amount: number;
    date: Date;
  }): Promise<void> {
    await this.waitForElement(this.selectors.addFinancialRecordButton);
    await this.click(this.selectors.addFinancialRecordButton);

    // 填写财务记录表单
    await this.waitForElement('#financial-type');
    await this.select('#financial-type', record.type);

    await this.waitForElement('#financial-description');
    await this.fill('#financial-description', record.description);

    await this.waitForElement('#financial-amount');
    await this.fill('#financial-amount', record.amount.toString());

    await this.waitForElement('#financial-date');
    await this.fill('#financial-date', record.date.toISOString().split('T')[0]);

    // 保存财务记录
    await this.waitForElement('#save-financial-button');
    await this.click('#save-financial-button');

    this.logger.debug('财务记录已添加', { description: record.description });
  }

  /**
   * 获取里程碑列表
   */
  override async getMilestoneList(): Promise<Array<{
    id: string;
    title: string;
    completed: boolean;
    dueDate?: Date;
  }>> {
    if (await this.isVisible(this.selectors.milestoneItems)) {
      // 在实际实现中，这里会解析里程碑元素
      return this.getMilestones();
    }
    return [];
  }

  /**
   * 获取文档列表
   */
  override async getDocumentList(): Promise<Array<{
    id: string;
    name: string;
    type: string;
    uploadedBy: string;
    uploadedAt: Date;
  }>> {
    if (await this.isVisible(this.selectors.documentItems)) {
      // 在实际实现中，这里会解析文档元素
      return [
        {
          id: 'doc-1',
          name: '合同文件.pdf',
          type: 'contract',
          uploadedBy: '张律师',
          uploadedAt: new Date('2024-01-15')
        },
        {
          id: 'doc-2',
          name: '证据材料.pdf',
          type: 'evidence',
          uploadedBy: '李律师',
          uploadedAt: new Date('2024-01-16')
        }
      ];
    }
    return [];
  }

  /**
   * 获取时间线
   */
  override async getTimeline(): Promise<Array<{
    id: string;
    action: string;
    description: string;
    timestamp: Date;
    user: string;
  }>> {
    if (await this.isVisible(this.selectors.timelineItems)) {
      // 在实际实现中，这里会解析时间线元素
      return [
        {
          id: 'timeline-1',
          action: '创建案件',
          description: '案件已创建并分配给张律师',
          timestamp: new Date('2024-01-15'),
          user: '系统'
        },
        {
          id: 'timeline-2',
          action: '更新状态',
          description: '案件状态已更新为进行中',
          timestamp: new Date('2024-01-16'),
          user: '张律师'
        }
      ];
    }
    return [];
  }

  /**
   * 获取财务记录
   */
  override async getFinancialRecords(): Promise<Array<{
    id: string;
    type: string;
    description: string;
    amount: number;
    date: Date;
    status: string;
  }>> {
    if (await this.isVisible(this.selectors.financialItems)) {
      // 在实际实现中，这里会解析财务记录元素
      return [
        {
          id: 'financial-1',
          type: 'fee',
          description: '律师费',
          amount: 50000,
          date: new Date('2024-01-15'),
          status: 'pending'
        },
        {
          id: 'financial-2',
          type: 'expense',
          description: '诉讼费',
          amount: 10000,
          date: new Date('2024-01-16'),
          status: 'paid'
        }
      ];
    }
    return [];
  }

  /**
   * 检查是否正在加载
   */
  override async isLoading(): Promise<boolean> {
    return await this.isVisible(this.selectors.loadingSpinner);
  }

  /**
   * 获取错误消息
   */
  override async getErrorMessage(): Promise<string> {
    if (await this.isVisible(this.selectors.errorMessage)) {
      return await this.getText(this.selectors.errorMessage);
    }
    return '';
  }

  /**
   * 获取成功消息
   */
  override async getSuccessMessage(): Promise<string> {
    if (await this.isVisible(this.selectors.successMessage)) {
      return await this.getText(this.selectors.successMessage);
    }
    return '';
  }

  /**
   * 等待保存完成
   */
  private override async waitForSaveComplete(): Promise<void> {
    try {
      await this.waitForEither([
        { selector: this.selectors.successMessage, timeout: 10000 },
        { selector: this.selectors.errorMessage, timeout: 5000 }
      ]);
    } catch (error) {
      throw new Error('等待保存完成超时');
    }
  }

  /**
   * 等待任一元素出现
   */
  private override async waitForEither(selectors: Array<{ selector: string; timeout?: number }>): Promise<void> {
    const defaultTimeout = 10000;

    for (const { selector, timeout = defaultTimeout } of selectors) {
      try {
        await this.waitForElement(selector, { timeout });
        return;
      } catch {
        // 继续尝试下一个选择器
      }
    }

    throw new Error('等待元素超时');
  }

  /**
   * 验证案件详情页面元素
   */
  override async validateCaseDetailPage(): Promise<{ valid: boolean; missingElements: string[] }> {
    const requiredElements = [
      { name: 'caseDetailContainer', selector: this.selectors.caseDetailContainer },
      { name: 'caseTitle', selector: this.selectors.caseTitle },
      { name: 'caseNumber', selector: this.selectors.caseNumber },
      { name: 'editButton', selector: this.selectors.editButton },
      { name: 'backButton', selector: this.selectors.backButton }
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