/**
 * 客户详情页面Page Object
 */

import { BasePageObject, PageObjectConfig } from '../../core/base-page-object';
import { Logger } from '../../core/logger';

export interface ClientDetail {
  id: string;
  name: string;
  type: string;
  industry: string;
  contactPerson: string;
  email: string;
  phone: string;
  address: string;
  website?: string;
  description?: string;
  status: string;
  createdDate: Date;
  updatedDate: Date;
  caseCount: number;
  totalValue: number;
  tags: string[];
  contacts: Array<{
    id: string;
    name: string;
    position: string;
    email: string;
    phone: string;
    isPrimary: boolean;
  }>;
  cases: Array<{
    id: string;
    title: string;
    status: string;
    type: string;
    assignedTo: string;
    createdDate: Date;
  }>;
  documents: Array<{
    id: string;
    name: string;
    type: string;
    uploadedBy: string;
    uploadedAt: Date;
  }>;
}

export interface ClientUpdateData {
  name?: string;
  type?: string;
  industry?: string;
  contactPerson?: string;
  email?: string;
  phone?: string;
  address?: string;
  website?: string;
  description?: string;
  status?: string;
  tags?: string[];
}

export class ClientDetailPage extends BasePageObject {
    protected override selectors = {
    clientDetailContainer: '#client-detail-container',
    clientName: '#client-name',
    clientType: '#client-type',
    clientIndustry: '#client-industry',
    clientStatus: '#client-status',
    clientContactPerson: '#client-contact-person',
    clientEmail: '#client-email',
    clientPhone: '#client-phone',
    clientAddress: '#client-address',
    clientWebsite: '#client-website',
    clientDescription: '#client-description',
    clientCaseCount: '#client-case-count',
    clientTotalValue: '#client-total-value',
    clientCreatedDate: '#client-created-date',
    clientUpdatedDate: '#client-updated-date',
    clientTags: '.client-tags',
    editButton: '#edit-client-button',
    deleteButton: '#delete-client-button',
    backButton: '#back-button',
    contactsSection: '#contacts-section',
    casesSection: '#cases-section',
    documentsSection: '#documents-section',
    statisticsSection: '#statistics-section',
    addContactButton: '#add-contact-button',
    addCaseButton: '#add-case-button',
    addDocumentButton: '#add-document-button',
    contactItems: '.contact-item',
    caseItems: '.case-item',
    documentItems: '.document-item',
    loadingSpinner: '.loading-spinner',
    errorMessage: '.error-message',
    successMessage: '.success-message',
    editForm: '#edit-client-form',
    saveButton: '#save-client-button',
    cancelButton: '#cancel-edit-button'
  };

  constructor(config: PageObjectConfig, logger?: Logger) {
    super(config, this.selectors, logger);
  }

  /**
   * 导航到客户详情页面
   */
  override async navigateToClientDetail(clientId: string): Promise<void> {
    await this.navigate(`/clients/${clientId}`);
    await this.waitForElement(this.selectors.clientDetailContainer);
  }

  /**
   * 获取客户详情
   */
  override async getClientDetail(): Promise<ClientDetail> {
    await this.waitForElement(this.selectors.clientName);

    // 在实际实现中，这里会从DOM获取客户详情
    // 现在返回模拟数据
    const mockClient: ClientDetail = {
      id: 'client-1',
      name: await this.getText(this.selectors.clientName) || '科技有限公司',
      type: await this.getText(this.selectors.clientType) || 'company',
      industry: await this.getText(this.selectors.clientIndustry) || 'technology',
      contactPerson: await this.getText(this.selectors.clientContactPerson) || '张经理',
      email: await this.getText(this.selectors.clientEmail) || 'contact@techcompany.com',
      phone: await this.getText(this.selectors.clientPhone) || '010-12345678',
      address: await this.getText(this.selectors.clientAddress) || '北京市朝阳区科技园区',
      website: await this.getText(this.selectors.clientWebsite) || 'https://techcompany.com',
      description: await this.getText(this.selectors.clientDescription) || '专注于软件开发和信息技术服务的科技公司',
      status: await this.getText(this.selectors.clientStatus) || 'active',
      createdDate: new Date(await this.getText(this.selectors.clientCreatedDate) || '2024-01-15'),
      updatedDate: new Date(await this.getText(this.selectors.clientUpdatedDate) || '2024-01-20'),
      caseCount: parseInt(await this.getText(this.selectors.clientCaseCount) || '3'),
      totalValue: parseFloat(await this.getText(this.selectors.clientTotalValue) || '500000'),
      tags: await this.getTags(),
      contacts: await this.getContacts(),
      cases: await this.getCases(),
      documents: await this.getDocuments()
    };

    this.logger.debug('获取客户详情', { clientId: mockClient.id });
    return mockClient;
  }

  /**
   * 获取标签
   */
  private override async getTags(): Promise<string[]> {
    if (await this.isVisible(this.selectors.clientTags)) {
      // 在实际实现中，这里会解析标签元素
      return ['重要客户', '长期合作', '技术公司'];
    }
    return [];
  }

  /**
   * 获取联系人列表
   */
  private override async getContacts(): Promise<Array<{
    id: string;
    name: string;
    position: string;
    email: string;
    phone: string;
    isPrimary: boolean;
  }>> {
    if (await this.isVisible(this.selectors.contactsSection)) {
      // 在实际实现中，这里会解析联系人元素
      return [
        {
          id: 'contact-1',
          name: '张经理',
          position: '总经理',
          email: 'zhang@techcompany.com',
          phone: '010-12345678',
          isPrimary: true
        },
        {
          id: 'contact-2',
          name: '李财务',
          position: '财务经理',
          email: 'li@techcompany.com',
          phone: '010-87654321',
          isPrimary: false
        }
      ];
    }
    return [];
  }

  /**
   * 获取案件列表
   */
  private override async getCases(): Promise<Array<{
    id: string;
    title: string;
    status: string;
    type: string;
    assignedTo: string;
    createdDate: Date;
  }>> {
    if (await this.isVisible(this.selectors.casesSection)) {
      // 在实际实现中，这里会解析案件元素
      return [
        {
          id: 'case-1',
          title: '合同纠纷案件',
          status: 'active',
          type: 'litigation',
          assignedTo: '张律师',
          createdDate: new Date('2024-01-15')
        },
        {
          id: 'case-2',
          title: '知识产权保护',
          status: 'pending',
          type: 'intellectual-property',
          assignedTo: '李律师',
          createdDate: new Date('2024-01-10')
        }
      ];
    }
    return [];
  }

  /**
   * 获取文档列表
   */
  private override async getDocuments(): Promise<Array<{
    id: string;
    name: string;
    type: string;
    uploadedBy: string;
    uploadedAt: Date;
  }>> {
    if (await this.isVisible(this.selectors.documentsSection)) {
      // 在实际实现中，这里会解析文档元素
      return [
        {
          id: 'doc-1',
          name: '合作协议.pdf',
          type: 'contract',
          uploadedBy: '张律师',
          uploadedAt: new Date('2024-01-15')
        },
        {
          id: 'doc-2',
          name: '营业执照.pdf',
          type: 'certificate',
          uploadedBy: '李律师',
          uploadedAt: new Date('2024-01-16')
        }
      ];
    }
    return [];
  }

  /**
   * 编辑客户
   */
  override async editClient(): Promise<void> {
    await this.waitForElement(this.selectors.editButton);
    await this.click(this.selectors.editButton);

    // 等待编辑表单出现
    await this.waitForElement(this.selectors.editForm);

    this.logger.debug('编辑客户模式已启动');
  }

  /**
   * 更新客户信息
   */
  override async updateClient(updateData: ClientUpdateData): Promise<void> {
    await this.editClient();

    if (updateData.name) {
      await this.waitForElement('#edit-name');
      await this.fill('#edit-name', updateData.name);
    }

    if (updateData.type) {
      await this.waitForElement('#edit-type');
      await this.select('#edit-type', updateData.type);
    }

    if (updateData.industry) {
      await this.waitForElement('#edit-industry');
      await this.select('#edit-industry', updateData.industry);
    }

    if (updateData.contactPerson) {
      await this.waitForElement('#edit-contact-person');
      await this.fill('#edit-contact-person', updateData.contactPerson);
    }

    if (updateData.email) {
      await this.waitForElement('#edit-email');
      await this.fill('#edit-email', updateData.email);
    }

    if (updateData.phone) {
      await this.waitForElement('#edit-phone');
      await this.fill('#edit-phone', updateData.phone);
    }

    if (updateData.address) {
      await this.waitForElement('#edit-address');
      await this.fill('#edit-address', updateData.address);
    }

    if (updateData.website) {
      await this.waitForElement('#edit-website');
      await this.fill('#edit-website', updateData.website);
    }

    if (updateData.description) {
      await this.waitForElement('#edit-description');
      await this.fill('#edit-description', updateData.description);
    }

    if (updateData.status) {
      await this.waitForElement('#edit-status');
      await this.select('#edit-status', updateData.status);
    }

    // 保存更改
    await this.waitForElement(this.selectors.saveButton);
    await this.click(this.selectors.saveButton);

    // 等待保存完成
    await this.waitForSaveComplete();

    this.logger.debug('客户信息已更新', { updateData });
  }

  /**
   * 删除客户
   */
  override async deleteClient(): Promise<void> {
    await this.waitForElement(this.selectors.deleteButton);
    await this.click(this.selectors.deleteButton);

    // 确认删除
    await this.waitForElement('.confirm-delete-button');
    await this.click('.confirm-delete-button');

    this.logger.debug('客户已删除');
  }

  /**
   * 返回客户列表
   */
  override async backToList(): Promise<void> {
    await this.waitForElement(this.selectors.backButton);
    await this.click(this.selectors.backButton);
    this.logger.debug('返回客户列表');
  }

  /**
   * 添加联系人
   */
  override async addContact(contact: {
    name: string;
    position: string;
    email: string;
    phone: string;
    isPrimary?: boolean;
  }): Promise<void> {
    await this.waitForElement(this.selectors.addContactButton);
    await this.click(this.selectors.addContactButton);

    // 填写联系人表单
    await this.waitForElement('#contact-name');
    await this.fill('#contact-name', contact.name);

    await this.waitForElement('#contact-position');
    await this.fill('#contact-position', contact.position);

    await this.waitForElement('#contact-email');
    await this.fill('#contact-email', contact.email);

    await this.waitForElement('#contact-phone');
    await this.fill('#contact-phone', contact.phone);

    if (contact.isPrimary) {
      await this.waitForElement('#contact-is-primary');
      await this.click('#contact-is-primary');
    }

    // 保存联系人
    await this.waitForElement('#save-contact-button');
    await this.click('#save-contact-button');

    this.logger.debug('联系人已添加', { name: contact.name });
  }

  /**
   * 添加案件
   */
  override async addCase(caseData: {
    title: string;
    type: string;
    priority: string;
    description?: string;
  }): Promise<void> {
    await this.waitForElement(this.selectors.addCaseButton);
    await this.click(this.selectors.addCaseButton);

    // 填写案件表单
    await this.waitForElement('#case-title');
    await this.fill('#case-title', caseData.title);

    await this.waitForElement('#case-type');
    await this.select('#case-type', caseData.type);

    await this.waitForElement('#case-priority');
    await this.select('#case-priority', caseData.priority);

    if (caseData.description) {
      await this.waitForElement('#case-description');
      await this.fill('#case-description', caseData.description);
    }

    // 保存案件
    await this.waitForElement('#save-case-button');
    await this.click('#save-case-button');

    this.logger.debug('案件已添加', { title: caseData.title });
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
   * 获取联系人列表
   */
  override async getContactList(): Promise<Array<{
    id: string;
    name: string;
    position: string;
    email: string;
    phone: string;
    isPrimary: boolean;
  }>> {
    if (await this.isVisible(this.selectors.contactItems)) {
      // 在实际实现中，这里会解析联系人元素
      return this.getContacts();
    }
    return [];
  }

  /**
   * 获取案件列表
   */
  override async getCaseList(): Promise<Array<{
    id: string;
    title: string;
    status: string;
    type: string;
    assignedTo: string;
    createdDate: Date;
  }>> {
    if (await this.isVisible(this.selectors.caseItems)) {
      // 在实际实现中，这里会解析案件元素
      return this.getCases();
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
      return this.getDocuments();
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
   * 验证客户详情页面元素
   */
  override async validateClientDetailPage(): Promise<{ valid: boolean; missingElements: string[] }> {
    const requiredElements = [
      { name: 'clientDetailContainer', selector: this.selectors.clientDetailContainer },
      { name: 'clientName', selector: this.selectors.clientName },
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