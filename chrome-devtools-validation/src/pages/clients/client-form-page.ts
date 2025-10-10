/**
 * 客户表单页面Page Object
 */

import { BasePageObject, PageObjectConfig } from '../../core/base-page-object';
import { Logger } from '../../core/logger';

export interface ClientFormData {
  name: string;
  type: 'company' | 'individual' | 'government' | 'other';
  industry: string;
  contactPerson: string;
  email: string;
  phone: string;
  address: string;
  website?: string;
  description?: string;
  status: 'active' | 'inactive' | 'prospect' | 'archived';
  tags?: string[];
  taxId?: string;
  registrationNumber?: string;
  bankAccount?: string;
  billingAddress?: string;
}

export interface ContactFormData {
  name: string;
  position: string;
  email: string;
  phone: string;
  department?: string;
  isPrimary: boolean;
  notes?: string;
}

export interface ValidationRule {
  required?: boolean;
  minLength?: number;
  maxLength?: number;
  pattern?: string;
  custom?: (value: string) => boolean;
  message: string;
}

export interface ClientFormValidation {
  name: ValidationRule;
  contactPerson: ValidationRule;
  email: ValidationRule;
  phone: ValidationRule;
  address: ValidationRule;
  type: ValidationRule;
  industry: ValidationRule;
  status: ValidationRule;
}

export class ClientFormPage extends BasePageObject {
  protected override selectors = {
    clientFormContainer: '#client-form-container',
    formTitle: '#form-title',
    clientNameInput: '#client-name',
    clientTypeSelect: '#client-type',
    clientIndustrySelect: '#client-industry',
    clientContactPersonInput: '#client-contact-person',
    clientEmailInput: '#client-email',
    clientPhoneInput: '#client-phone',
    clientAddressInput: '#client-address',
    clientWebsiteInput: '#client-website',
    clientDescriptionTextarea: '#client-description',
    clientStatusSelect: '#client-status',
    clientTagsInput: '#client-tags',
    clientTaxIdInput: '#client-tax-id',
    clientRegistrationNumberInput: '#client-registration-number',
    clientBankAccountInput: '#client-bank-account',
    clientBillingAddressInput: '#client-billing-address',
    saveButton: '#save-client-button',
    saveAndNewButton: '#save-and-new-button',
    cancelButton: '#cancel-button',
    resetButton: '#reset-button',
    loadingSpinner: '.loading-spinner',
    errorMessage: '.error-message',
    successMessage: '.success-message',
    validationErrors: '.validation-error',
    contactsSection: '#contacts-section',
    addContactButton: '#add-contact-button',
    contactItems: '.contact-item',
    contactForm: '#contact-form',
    contactNameInput: '#contact-name',
    contactPositionInput: '#contact-position',
    contactEmailInput: '#contact-email',
    contactPhoneInput: '#contact-phone',
    contactDepartmentInput: '#contact-department',
    contactIsPrimaryCheckbox: '#contact-is-primary',
    contactNotesTextarea: '#contact-notes',
    saveContactButton: '#save-contact-button',
    cancelContactButton: '#cancel-contact-button',
    removeContactButton: (contactId: string) => `.contact-item[data-id="${contactId}"] .remove-contact-button`,
    duplicateButton: '#duplicate-client-button',
    bulkImportButton: '#bulk-import-button',
    autoFillButton: '#auto-fill-button',
    requiredFields: '.required-field',
    fieldLabels: {
      name: 'label[for="client-name"]',
      type: 'label[for="client-type"]',
      industry: 'label[for="client-industry"]',
      contactPerson: 'label[for="client-contact-person"]',
      email: 'label[for="client-email"]',
      phone: 'label[for="client-phone"]',
      address: 'label[for="client-address"]'
    },
    fieldHints: {
      email: '#email-hint',
      phone: '#phone-hint',
      website: '#website-hint',
      taxId: '#tax-id-hint'
    },
    progressIndicator: '.form-progress',
    stepIndicator: '.step-indicator',
    formSections: {
      basic: '#basic-info-section',
      contact: '#contact-info-section',
      financial: '#financial-info-section',
      additional: '#additional-info-section'
    }
  };

  private validationRules: ClientFormValidation = {
    name: {
      required: true,
      minLength: 2,
      maxLength: 100,
      message: '客户名称必须在2-100个字符之间'
    },
    contactPerson: {
      required: true,
      minLength: 2,
      maxLength: 50,
      message: '联系人姓名必须在2-50个字符之间'
    },
    email: {
      required: true,
      pattern: '^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$',
      message: '请输入有效的电子邮件地址'
    },
    phone: {
      required: true,
      pattern: '^[+]?[\\d\\s\\-()]{10,}$',
      message: '请输入有效的电话号码'
    },
    address: {
      required: true,
      minLength: 10,
      maxLength: 200,
      message: '地址必须在10-200个字符之间'
    },
    type: {
      required: true,
      message: '请选择客户类型'
    },
    industry: {
      required: true,
      message: '请选择所属行业'
    },
    status: {
      required: true,
      message: '请选择客户状态'
    }
  };

  constructor(config: PageObjectConfig, logger?: Logger) {
    super(config, this.selectors, logger);
  }

  /**
   * 导航到创建客户页面
   */
  override async navigateToCreateClient(): Promise<void> {
    await this.navigate('/clients/create');
    await this.waitForElement(this.selectors.clientFormContainer);
    this.logger.debug('已导航到创建客户页面');
  }

  /**
   * 导航到编辑客户页面
   */
  override async navigateToEditClient(clientId: string): Promise<void> {
    await this.navigate(`/clients/${clientId}/edit`);
    await this.waitForElement(this.selectors.clientFormContainer);
    this.logger.debug('已导航到编辑客户页面', { clientId });
  }

  /**
   * 获取表单标题
   */
  override async getFormTitle(): Promise<string> {
    return await this.getText(this.selectors.formTitle) || '客户信息';
  }

  /**
   * 填写基本信息
   */
  override async fillBasicInfo(data: {
    name: string;
    type: string;
    industry: string;
    status: string;
    description?: string;
    tags?: string[];
  }): Promise<void> {
    await this.waitForElement(this.selectors.clientNameInput);
    await this.fill(this.selectors.clientNameInput, data.name);

    await this.waitForElement(this.selectors.clientTypeSelect);
    await this.select(this.selectors.clientTypeSelect, data.type);

    await this.waitForElement(this.selectors.clientIndustrySelect);
    await this.select(this.selectors.clientIndustrySelect, data.industry);

    await this.waitForElement(this.selectors.clientStatusSelect);
    await this.select(this.selectors.clientStatusSelect, data.status);

    if (data.description) {
      await this.waitForElement(this.selectors.clientDescriptionTextarea);
      await this.fill(this.selectors.clientDescriptionTextarea, data.description);
    }

    if (data.tags && data.tags.length > 0) {
      await this.waitForElement(this.selectors.clientTagsInput);
      await this.fill(this.selectors.clientTagsInput, data.tags.join(', '));
    }

    this.logger.debug('基本信息已填写', data);
  }

  /**
   * 填写联系信息
   */
  override async fillContactInfo(data: {
    contactPerson: string;
    email: string;
    phone: string;
    address: string;
    website?: string;
    billingAddress?: string;
  }): Promise<void> {
    await this.waitForElement(this.selectors.clientContactPersonInput);
    await this.fill(this.selectors.clientContactPersonInput, data.contactPerson);

    await this.waitForElement(this.selectors.clientEmailInput);
    await this.fill(this.selectors.clientEmailInput, data.email);

    await this.waitForElement(this.selectors.clientPhoneInput);
    await this.fill(this.selectors.clientPhoneInput, data.phone);

    await this.waitForElement(this.selectors.clientAddressInput);
    await this.fill(this.selectors.clientAddressInput, data.address);

    if (data.website) {
      await this.waitForElement(this.selectors.clientWebsiteInput);
      await this.fill(this.selectors.clientWebsiteInput, data.website);
    }

    if (data.billingAddress) {
      await this.waitForElement(this.selectors.clientBillingAddressInput);
      await this.fill(this.selectors.clientBillingAddressInput, data.billingAddress);
    }

    this.logger.debug('联系信息已填写', data);
  }

  /**
   * 填写财务信息
   */
  override async fillFinancialInfo(data: {
    taxId?: string;
    registrationNumber?: string;
    bankAccount?: string;
  }): Promise<void> {
    if (data.taxId) {
      await this.waitForElement(this.selectors.clientTaxIdInput);
      await this.fill(this.selectors.clientTaxIdInput, data.taxId);
    }

    if (data.registrationNumber) {
      await this.waitForElement(this.selectors.clientRegistrationNumberInput);
      await this.fill(this.selectors.clientRegistrationNumberInput, data.registrationNumber);
    }

    if (data.bankAccount) {
      await this.waitForElement(this.selectors.clientBankAccountInput);
      await this.fill(this.selectors.clientBankAccountInput, data.bankAccount);
    }

    this.logger.debug('财务信息已填写', data);
  }

  /**
   * 填写完整客户表单
   */
  override async fillClientForm(data: ClientFormData): Promise<void> {
    await this.fillBasicInfo({
      name: data.name,
      type: data.type,
      industry: data.industry,
      status: data.status,
      description: data.description,
      tags: data.tags
    });

    await this.fillContactInfo({
      contactPerson: data.contactPerson,
      email: data.email,
      phone: data.phone,
      address: data.address,
      website: data.website,
      billingAddress: data.billingAddress
    });

    await this.fillFinancialInfo({
      taxId: data.taxId,
      registrationNumber: data.registrationNumber,
      bankAccount: data.bankAccount
    });

    this.logger.debug('客户表单已完整填写', data);
  }

  /**
   * 添加联系人
   */
  override async addContact(contact: ContactFormData): Promise<void> {
    await this.waitForElement(this.selectors.addContactButton);
    await this.click(this.selectors.addContactButton);

    await this.waitForElement(this.selectors.contactForm);

    await this.waitForElement(this.selectors.contactNameInput);
    await this.fill(this.selectors.contactNameInput, contact.name);

    await this.waitForElement(this.selectors.contactPositionInput);
    await this.fill(this.selectors.contactPositionInput, contact.position);

    await this.waitForElement(this.selectors.contactEmailInput);
    await this.fill(this.selectors.contactEmailInput, contact.email);

    await this.waitForElement(this.selectors.contactPhoneInput);
    await this.fill(this.selectors.contactPhoneInput, contact.phone);

    if (contact.department) {
      await this.waitForElement(this.selectors.contactDepartmentInput);
      await this.fill(this.selectors.contactDepartmentInput, contact.department);
    }

    if (contact.isPrimary) {
      await this.waitForElement(this.selectors.contactIsPrimaryCheckbox);
      await this.click(this.selectors.contactIsPrimaryCheckbox);
    }

    if (contact.notes) {
      await this.waitForElement(this.selectors.contactNotesTextarea);
      await this.fill(this.selectors.contactNotesTextarea, contact.notes);
    }

    await this.waitForElement(this.selectors.saveContactButton);
    await this.click(this.selectors.saveContactButton);

    this.logger.debug('联系人已添加', contact);
  }

  /**
   * 保存客户
   */
  override async saveClient(): Promise<void> {
    await this.waitForElement(this.selectors.saveButton);
    await this.click(this.selectors.saveButton);

    await this.waitForSaveComplete();

    this.logger.debug('客户保存操作已执行');
  }

  /**
   * 保存并新建
   */
  override async saveAndNew(): Promise<void> {
    await this.waitForElement(this.selectors.saveAndNewButton);
    await this.click(this.selectors.saveAndNewButton);

    await this.waitForSaveComplete();

    this.logger.debug('保存并新建操作已执行');
  }

  /**
   * 取消操作
   */
  override async cancel(): Promise<void> {
    await this.waitForElement(this.selectors.cancelButton);
    await this.click(this.selectors.cancelButton);

    this.logger.debug('操作已取消');
  }

  /**
   * 重置表单
   */
  override async resetForm(): Promise<void> {
    if (await this.isVisible(this.selectors.resetButton)) {
      await this.waitForElement(this.selectors.resetButton);
      await this.click(this.selectors.resetButton);
      this.logger.debug('表单已重置');
    }
  }

  /**
   * 复制客户
   */
  override async duplicateClient(): Promise<void> {
    if (await this.isVisible(this.selectors.duplicateButton)) {
      await this.waitForElement(this.selectors.duplicateButton);
      await this.click(this.selectors.duplicateButton);
      this.logger.debug('客户复制操作已执行');
    }
  }

  /**
   * 验证表单字段
   */
  override async validateFormField(fieldName: keyof ClientFormValidation, value: string): Promise<{ valid: boolean; message?: string }> {
    const rule = this.validationRules[fieldName];

    if (rule.required && !value.trim()) {
      return { valid: false, message: rule.message };
    }

    if (rule.minLength && value.length < rule.minLength) {
      return { valid: false, message: rule.message };
    }

    if (rule.maxLength && value.length > rule.maxLength) {
      return { valid: false, message: rule.message };
    }

    if (rule.pattern && !new RegExp(rule.pattern).test(value)) {
      return { valid: false, message: rule.message };
    }

    if (rule.custom && !rule.custom(value)) {
      return { valid: false, message: rule.message };
    }

    return { valid: true };
  }

  /**
   * 验证整个表单
   */
  override async validateForm(): Promise<{ valid: boolean; errors: Record<string, string> }> {
    const errors: Record<string, string> = {};

    // 获取表单数据
    const formData = await this.getFormData();

    // 验证每个字段
    for (const [fieldName, rule] of Object.entries(this.validationRules)) {
      const value = formData[fieldName as keyof ClientFormData] as string;
      const validation = await this.validateFormField(fieldName as keyof ClientFormValidation, value);

      if (!validation.valid) {
        errors[fieldName] = validation.message || '验证失败';
      }
    }

    return {
      valid: Object.keys(errors).length === 0,
      errors
    };
  }

  /**
   * 获取表单数据
   */
  override async getFormData(): Promise<Partial<ClientFormData>> {
    const formData: Partial<ClientFormData> = {};

    if (await this.isVisible(this.selectors.clientNameInput)) {
      formData.name = await this.getText(this.selectors.clientNameInput);
    }

    if (await this.isVisible(this.selectors.clientTypeSelect)) {
      formData.type = await this.getSelectedValue(this.selectors.clientTypeSelect) as any;
    }

    if (await this.isVisible(this.selectors.clientIndustrySelect)) {
      formData.industry = await this.getSelectedValue(this.selectors.clientIndustrySelect);
    }

    if (await this.isVisible(this.selectors.clientStatusSelect)) {
      formData.status = await this.getSelectedValue(this.selectors.clientStatusSelect) as any;
    }

    if (await this.isVisible(this.selectors.clientContactPersonInput)) {
      formData.contactPerson = await this.getText(this.selectors.clientContactPersonInput);
    }

    if (await this.isVisible(this.selectors.clientEmailInput)) {
      formData.email = await this.getText(this.selectors.clientEmailInput);
    }

    if (await this.isVisible(this.selectors.clientPhoneInput)) {
      formData.phone = await this.getText(this.selectors.clientPhoneInput);
    }

    if (await this.isVisible(this.selectors.clientAddressInput)) {
      formData.address = await this.getText(this.selectors.clientAddressInput);
    }

    if (await this.isVisible(this.selectors.clientWebsiteInput)) {
      formData.website = await this.getText(this.selectors.clientWebsiteInput);
    }

    if (await this.isVisible(this.selectors.clientDescriptionTextarea)) {
      formData.description = await this.getText(this.selectors.clientDescriptionTextarea);
    }

    if (await this.isVisible(this.selectors.clientTagsInput)) {
      const tagsText = await this.getText(this.selectors.clientTagsInput);
      formData.tags = tagsText.split(',').map(tag => tag.trim()).filter(tag => tag);
    }

    return formData;
  }

  /**
   * 获取表单验证错误
   */
  override async getValidationErrors(): Promise<Record<string, string[]>> {
    const errors: Record<string, string[]> = {};

    if (await this.isVisible(this.selectors.validationErrors)) {
      const errorElements = await this.querySelectorAll(this.selectors.validationErrors);

      for (const element of errorElements) {
        const fieldName = element.getAttribute('data-field') || 'unknown';
        const errorMessage = element.textContent || '';

        if (!errors[fieldName]) {
          errors[fieldName] = [];
        }
        errors[fieldName].push(errorMessage);
      }
    }

    return errors;
  }

  /**
   * 检查字段是否显示验证错误
   */
  override async hasFieldError(fieldName: string): Promise<boolean> {
    const errors = await this.getValidationErrors();
    return errors[fieldName] && errors[fieldName].length > 0;
  }

  /**
   * 获取字段验证错误消息
   */
  override async getFieldError(fieldName: string): Promise<string> {
    const errors = await this.getValidationErrors();
    return errors[fieldName] ? errors[fieldName][0] : '';
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
   * 检查字段是否为必填
   */
  override async isFieldRequired(fieldName: string): Promise<boolean> {
    const selector = `${this.selectors.fieldLabels[fieldName as keyof typeof this.selectors.fieldLabels]} .required-marker`;
    return await this.isVisible(selector);
  }

  /**
   * 获取字段提示信息
   */
  override async getFieldHint(fieldName: string): Promise<string> {
    const selector = this.selectors.fieldHints[fieldName as keyof typeof this.selectors.fieldHints];
    if (selector && await this.isVisible(selector)) {
      return await this.getText(selector);
    }
    return '';
  }

  /**
   * 获取表单进度
   */
  override async getFormProgress(): Promise<number> {
    if (await this.isVisible(this.selectors.progressIndicator)) {
      const progressText = await this.getText(this.selectors.progressIndicator);
      const match = progressText.match(/(\d+)%/);
      return match ? parseInt(match[1]) : 0;
    }
    return 0;
  }

  /**
   * 获取当前步骤
   */
  override async getCurrentStep(): Promise<number> {
    if (await this.isVisible(this.selectors.stepIndicator)) {
      const stepText = await this.getText(this.selectors.stepIndicator);
      const match = stepText.match(/步骤\s*(\d+)/);
      return match ? parseInt(match[1]) : 1;
    }
    return 1;
  }

  /**
   * 导航到指定步骤
   */
  override async goToStep(step: number): Promise<void> {
    const stepButton = `.step-button[data-step="${step}"]`;
    if (await this.isVisible(stepButton)) {
      await this.waitForElement(stepButton);
      await this.click(stepButton);
      this.logger.debug('已导航到步骤', { step });
    }
  }

  /**
   * 检查表单是否有效
   */
  override async isFormValid(): Promise<boolean> {
    const validation = await this.validateForm();
    return validation.valid;
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
   * 获取选中的值
   */
  private override async getSelectedValue(selector: string): Promise<string> {
    const selectElement = await this.querySelector(selector);
    return selectElement ? selectElement.value : '';
  }

  /**
   * 验证客户表单页面元素
   */
  override async validateClientFormPage(): Promise<{ valid: boolean; missingElements: string[] }> {
    const requiredElements = [
      { name: 'clientFormContainer', selector: this.selectors.clientFormContainer },
      { name: 'clientNameInput', selector: this.selectors.clientNameInput },
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