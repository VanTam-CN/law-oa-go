/**
 * 案件创建/编辑表单页面Page Object
 */

import { BasePageObject, PageObjectConfig } from '../../core/base-page-object';
import { Logger } from '../../core/logger';

export interface CaseFormData {
  title: string;
  description: string;
  type: string;
  priority: string;
  client: string;
  assignedTo: string;
  dueDate?: Date;
  estimatedValue?: number;
  tags?: string[];
}

export class CaseFormPage extends BasePageObject {
  protected override selectors = {
    caseForm: '#case-form',
    titleInput: '#case-title',
    descriptionTextarea: '#case-description',
    typeSelect: '#case-type',
    prioritySelect: '#case-priority',
    clientSelect: '#case-client',
    assignedToSelect: '#case-assigned-to',
    dueDateInput: '#case-due-date',
    estimatedValueInput: '#case-estimated-value',
    tagsInput: '#case-tags',
    saveButton: '#save-case-button',
    cancelButton: '#cancel-button',
    errorMessage: '.error-message',
    successMessage: '.success-message',
    loadingSpinner: '.loading-spinner',
    validationErrors: '.validation-error',
    tagSuggestions: '.tag-suggestions',
    clientSearch: '#client-search',
    assignedToSearch: '#assigned-to-search',
    formSections: '.form-section',
    requiredFields: '.required-field'
  };

  constructor(config: PageObjectConfig, logger?: Logger) {
    super(config, this.selectors, logger);
  }

  /**
   * 导航到创建案件页面
   */
  override async navigateToCreateCase(): Promise<void> {
    await this.navigate('/cases/create');
    await this.waitForElement(this.selectors.caseForm);
  }

  /**
   * 导航到编辑案件页面
   */
  override async navigateToEditCase(caseId: string): Promise<void> {
    await this.navigate(`/cases/${caseId}/edit`);
    await this.waitForElement(this.selectors.caseForm);
  }

  /**
   * 填写案件表单
   */
  override async fillCaseForm(formData: CaseFormData): Promise<void> {
    // 填写标题
    if (formData.title) {
      await this.waitForElement(this.selectors.titleInput);
      await this.fill(this.selectors.titleInput, formData.title);
    }

    // 填写描述
    if (formData.description) {
      await this.waitForElement(this.selectors.descriptionTextarea);
      await this.fill(this.selectors.descriptionTextarea, formData.description);
    }

    // 选择类型
    if (formData.type) {
      await this.waitForElement(this.selectors.typeSelect);
      await this.select(this.selectors.typeSelect, formData.type);
    }

    // 选择优先级
    if (formData.priority) {
      await this.waitForElement(this.selectors.prioritySelect);
      await this.select(this.selectors.prioritySelect, formData.priority);
    }

    // 选择客户
    if (formData.client) {
      await this.waitForElement(this.selectors.clientSelect);
      await this.select(this.selectors.clientSelect, formData.client);
    }

    // 选择分配给
    if (formData.assignedTo) {
      await this.waitForElement(this.selectors.assignedToSelect);
      await this.select(this.selectors.assignedToSelect, formData.assignedTo);
    }

    // 填写截止日期
    if (formData.dueDate) {
      await this.waitForElement(this.selectors.dueDateInput);
      await this.fill(this.selectors.dueDateInput, formData.dueDate.toISOString().split('T')[0]);
    }

    // 填写预估价值
    if (formData.estimatedValue) {
      await this.waitForElement(this.selectors.estimatedValueInput);
      await this.fill(this.selectors.estimatedValueInput, formData.estimatedValue.toString());
    }

    // 添加标签
    if (formData.tags && formData.tags.length > 0) {
      await this.addTags(formData.tags);
    }

    this.logger.debug('案件表单已填写', { title: formData.title });
  }

  /**
   * 添加标签
   */
  override async addTags(tags: string[]): Promise<void> {
    await this.waitForElement(this.selectors.tagsInput);

    for (const tag of tags) {
      await this.fill(this.selectors.tagsInput, tag);
      await this.pressKey('Enter');
      await this.wait(100); // 等待标签添加完成
    }

    this.logger.debug('标签已添加', { tags });
  }

  /**
   * 搜索客户
   */
  override async searchClient(query: string): Promise<void> {
    await this.waitForElement(this.selectors.clientSearch);
    await this.fill(this.selectors.clientSearch, query);
    await this.wait(500); // 等待搜索结果
    this.logger.debug('客户搜索完成', { query });
  }

  /**
   * 搜索分配人
   */
  override async searchAssignedTo(query: string): Promise<void> {
    await this.waitForElement(this.selectors.assignedToSearch);
    await this.fill(this.selectors.assignedToSearch, query);
    await this.wait(500); // 等待搜索结果
    this.logger.debug('分配人搜索完成', { query });
  }

  /**
   * 保存案件
   */
  override async saveCase(): Promise<void> {
    await this.waitForElement(this.selectors.saveButton);
    await this.click(this.selectors.saveButton);

    // 等待保存完成
    await this.waitForSaveComplete();

    this.logger.debug('案件保存完成');
  }

  /**
   * 取消编辑
   */
  override async cancelEdit(): Promise<void> {
    await this.waitForElement(this.selectors.cancelButton);
    await this.click(this.selectors.cancelButton);
    this.logger.debug('编辑已取消');
  }

  /**
   * 创建新案件
   */
  override async createCase(formData: CaseFormData): Promise<void> {
    this.logger.info('开始创建案件', { title: formData.title });

    await this.navigateToCreateCase();
    await this.fillCaseForm(formData);
    await this.saveCase();

    this.logger.info('案件创建完成', { title: formData.title });
  }

  /**
   * 更新案件
   */
  override async updateCase(caseId: string, formData: CaseFormData): Promise<void> {
    this.logger.info('开始更新案件', { caseId, title: formData.title });

    await this.navigateToEditCase(caseId);
    await this.fillCaseForm(formData);
    await this.saveCase();

    this.logger.info('案件更新完成', { caseId, title: formData.title });
  }

  /**
   * 验证表单
   */
  override async validateForm(): Promise<{ valid: boolean; errors: string[] }> {
    const errors: string[] | undefined = undefined;

    // 检查必填字段
    if (!await this.isFieldValuePresent(this.selectors.titleInput)) {
      errors.push('案件标题不能为空');
    }

    if (!await this.isFieldValuePresent(this.selectors.typeSelect)) {
      errors.push('案件类型不能为空');
    }

    if (!await this.isFieldValuePresent(this.selectors.clientSelect)) {
      errors.push('客户不能为空');
    }

    if (!await this.isFieldValuePresent(this.selectors.assignedToSelect)) {
      errors.push('分配人不能为空');
    }

    // 检查验证错误
    const validationErrors = await this.getValidationErrors();
    errors.push(...validationErrors);

    return {
      valid: errors.length === 0,
      errors
    };
  }

  /**
   * 获取验证错误
   */
  override async getValidationErrors(): Promise<string[]> {
    const errors: string[] | undefined = undefined;

    if (await this.isVisible(this.selectors.validationErrors)) {
      // 在实际实现中，这里会解析具体的验证错误
      const errorText = await this.getText(this.selectors.validationErrors);
      errors.push(errorText);
    }

    return errors;
  }

  /**
   * 获取表单数据
   */
  override async getFormData(): Promise<CaseFormData> {
    const title = await this.getAttribute(this.selectors.titleInput, 'value') || '';
    const description = await this.getAttribute(this.selectors.descriptionTextarea, 'value') || '';
    const type = await this.getAttribute(this.selectors.typeSelect, 'value') || '';
    const priority = await this.getAttribute(this.selectors.prioritySelect, 'value') || '';
    const client = await this.getAttribute(this.selectors.clientSelect, 'value') || '';
    const assignedTo = await this.getAttribute(this.selectors.assignedToSelect, 'value') || '';
    const dueDate = await this.getAttribute(this.selectors.dueDateInput, 'value');
    const estimatedValue = await this.getAttribute(this.selectors.estimatedValueInput, 'value');
    const tags = await this.getTags();

    return {
      title: title || '',
      description: description || '',
      type: type || '',
      priority: priority || '',
      client: client || '',
      assignedTo: assignedTo || '',
      dueDate: dueDate ? new Date(dueDate) : undefined,
      estimatedValue: estimatedValue ? parseFloat(estimatedValue) : undefined,
      tags
    };
  }

  /**
   * 获取标签
   */
  override async getTags(): Promise<string[]> {
    // 在实际实现中，这里会解析标签元素
    return ['重要', '紧急'];
  }

  /**
   * 检查字段值是否存在
   */
  private override async isFieldValuePresent(selector: string): Promise<boolean> {
    const value = await this.getAttribute(selector, 'value');
    return value !== null && value.trim() !== '';
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
   * 检查是否正在加载
   */
  override async isLoading(): Promise<boolean> {
    return await this.isVisible(this.selectors.loadingSpinner);
  }

  /**
   * 清除表单
   */
  override async clearForm(): Promise<void> {
    const fields = [
      this.selectors.titleInput,
      this.selectors.descriptionTextarea,
      this.selectors.dueDateInput,
      this.selectors.estimatedValueInput,
      this.selectors.tagsInput
    ];

    for (const field of fields) {
      if (await this.isVisible(field)) {
        await this.fill(field, '');
      }
    }

    // 重置下拉选择
    const selects = [
      this.selectors.typeSelect,
      this.selectors.prioritySelect,
      this.selectors.clientSelect,
      this.selectors.assignedToSelect
    ];

    for (const select of selects) {
      if (await this.isVisible(select)) {
        await this.select(select, '');
      }
    }

    this.logger.debug('案件表单已清除');
  }

  /**
   * 检查表单是否可见
   */
  override async isFormVisible(): Promise<boolean> {
    return await this.isVisible(this.selectors.caseForm);
  }

  /**
   * 按下键盘按键
   */
  override async pressKey(key: string): Promise<void> {
    // 在实际实现中，这里会调用键盘操作
    this.logger.debug('按键按下', { key });
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
   * 验证案件表单页面元素
   */
  override async validateCaseFormPage(): Promise<{ valid: boolean; missingElements: string[] }> {
    const requiredElements = [
      { name: 'caseForm', selector: this.selectors.caseForm },
      { name: 'titleInput', selector: this.selectors.titleInput },
      { name: 'typeSelect', selector: this.selectors.typeSelect },
      { name: 'clientSelect', selector: this.selectors.clientSelect },
      { name: 'assignedToSelect', selector: this.selectors.assignedToSelect },
      { name: 'saveButton', selector: this.selectors.saveButton },
      { name: 'cancelButton', selector: this.selectors.cancelButton }
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