/**
 * 文档表单页面
 */

import { BasePageObject } from '../core/base-page-object';
import { Logger } from '../core/logger';
import { DocumentUploadData, DocumentUpdateData } from './document-list-page';

export interface DocumentFormData {
  name: string;
  type: string;
  description?: string;
  client?: string;
  case?: string;
  tags: string[];
  isPublic: boolean;
  isEncrypted: boolean;
  encryptionKey?: string;
  permissions?: {
    canView?: string[];
    canEdit?: string[];
    canDelete?: string[];
    canShare?: string[];
  };
}

export interface DocumentFormValidation {
  name?: string;
  type?: string;
  file?: string;
  encryptionKey?: string;
  permissions?: string;
}

export interface DocumentPreviewOptions {
  page?: number;
  zoom?: number;
  rotation?: number;
  showAnnotations?: boolean;
}

export class DocumentFormPage extends BasePageObject {
  private baseUrl: string;

  constructor(config: { baseUrl: string; defaultTimeout?: number; screenshotOnFailure?: boolean }, logger?: Logger) {
    super(config, this.selectors, logger);
    this.baseUrl = config.baseUrl;
  }

  /**
   * 导航到创建文档页面
   */
  override async navigateToCreateDocument(): Promise<void> {
    await this.navigate(`${this.baseUrl}/documents/create`);
  }

  /**
   * 导航到编辑文档页面
   */
  override async navigateToEditDocument(documentId: string): Promise<void> {
    await this.navigate(`${this.baseUrl}/documents/${documentId}/edit`);
  }

  /**
   * 上传文件
   */
  override async uploadFile(filePath: string): Promise<void> {
    try {
      // 点击上传按钮
      await this.click('#document-file-upload-button');
      await this.wait(1000); // 等待文件选择器打开

      // 这里需要使用特定的文件上传方法
      // 在实际实现中，这可能需要调用特定的文件上传API
      await this.executeScript(`
        (function() {
          const input = document.createElement('input');
          input.type = 'file';
          input.style.display = 'none';
          input.name = 'document-file';
          document.body.appendChild(input);
          input.click();
          return input;
        })();
      `);

      // 等待文件上传完成
      await this.wait(2000);

    } catch (error) {
      this.logger.error('上传文件失败', { error, filePath });
      throw error;
    }
  }

  /**
   * 填充文档表单
   */
  override async fillDocumentForm(data: DocumentFormData): Promise<void> {
    try {
      // 填充基本信息
      await this.fillBasicInfo(data);

      // 填充权限设置
      if (data.permissions) {
        await this.fillPermissions(data.permissions);
      }

      // 填充加密设置
      if (data.isEncrypted) {
        await this.fillEncryptionSettings(data.encryptionKey);
      }

    } catch (error) {
      this.logger.error('填充文档表单失败', { error, data });
      throw error;
    }
  }

  /**
   * 填充基本信息
   */
  override async fillBasicInfo(data: {
    name: string;
    type: string;
    description?: string;
    client?: string;
    case?: string;
    tags: string[];
  }): Promise<void> {
    try {
      if (data.name) {
        await this.fill('#document-name', data.name);
      }

      if (data.type) {
        await this.selectOption('#document-type', [data.type]);
      }

      if (data.description) {
        await this.fill('#document-description', data.description);
      }

      if (data.client) {
        await this.selectOption('#document-client', [data.client]);
      }

      if (data.case) {
        await this.selectOption('#document-case', [data.case]);
      }

      if (data.tags && data.tags.length > 0) {
        await this.fill('#document-tags', data.tags.join(', '));
      }

    } catch (error) {
      this.logger.error('填充文档基本信息失败', { error, data });
      throw error;
    }
  }

  /**
   * 填充权限设置
   */
  override async fillPermissions(permissions: {
    canView?: string[];
    canEdit?: string[];
    canDelete?: string[];
    canShare?: string[];
  }): Promise<void> {
    try {
      if (permissions.canView && permissions.canView.length > 0) {
        await this.fill('#document-permissions-view', permissions.canView.join(', '));
      }

      if (permissions.canEdit && permissions.canEdit.length > 0) {
        await this.fill('#document-permissions-edit', permissions.canEdit.join(', '));
      }

      if (permissions.canDelete && permissions.canDelete.length > 0) {
        await this.fill('#document-permissions-delete', permissions.canDelete.join(', '));
      }

      if (permissions.canShare && permissions.canShare.length > 0) {
        await this.fill('#document-permissions-share', permissions.canShare.join(', '));
      }

    } catch (error) {
      this.logger.error('填充文档权限设置失败', { error, permissions });
      throw error;
    }
  }

  /**
   * 填充加密设置
   */
  override async fillEncryptionSettings(encryptionKey?: string): Promise<void> {
    try {
      // 启用加密
      const encryptionCheckbox = await this.isVisible('#document-encryption-checkbox');
      const isChecked = await this.executeScript('return document.getElementById("document-encryption-checkbox").checked;');

      if (!isChecked) {
        await this.click('#document-encryption-checkbox');
      }

      if (encryptionKey) {
        await this.fill('#document-encryption-key', encryptionKey);
      }

      // 生成密钥按钮
      const generateButton = await this.isVisible('#document-generate-key-button');
      if (generateButton) {
        await this.click('#document-generate-key-button');
        await this.wait(1000); // 等待密钥生成
      }

    } catch (error) {
      this.logger.error('填充文档加密设置失败', { error });
      throw error;
    }
  }

  /**
   * 预览文档
   */
  override async previewDocument(options?: DocumentPreviewOptions): Promise<void> {
    try {
      await this.click('#document-preview-button');
      await this.waitForElement('.document-preview-modal', { timeout: 5000 });

      if (options) {
        if (options.page) {
          await this.fill('#preview-page-number', options.page.toString());
        }

        if (options.zoom) {
          await this.selectOption('#preview-zoom', [options.zoom.toString()]);
        }

        if (options.rotation) {
          await this.click('#preview-rotate-button');
        }

        if (options.showAnnotations !== undefined) {
          const annotationCheckbox = await this.isVisible('#preview-show-annotations');
          const isChecked = await this.executeScript('return document.getElementById("preview-show-annotations").checked;');

          if (options.showAnnotations !== isChecked) {
            await this.click('#preview-show-annotations');
          }
        }
      }

    } catch (error) {
      this.logger.error('预览文档失败', { error, options });
      throw error;
    }
  }

  /**
   * 提取文档元数据
   */
  override async extractMetadata(): Promise<Record<string, any>> {
    try {
      await this.click('#document-extract-metadata-button');
      await this.wait(2000); // 等待元数据提取完成

      const metadata = await this.executeScript(`
        (function() {
          const metadataElements = document.querySelectorAll('.document-metadata-item');
          const metadata = {};

          metadataElements.forEach(item => {
            const key = item.querySelector('.metadata-key')?.gettextContent?.().trim() || '';
            const value = item.querySelector('.metadata-value')?.gettextContent?.().trim() || '';
            if (key && value) {
              metadata[key] = value;
            }
          });

          return metadata;
        })();
      `);

      return metadata;

    } catch (error) {
      this.logger.error('提取文档元数据失败', { error });
      throw error;
    }
  }

  /**
   * 验证表单
   */
  override async validateForm(): Promise<{
    valid: boolean;
    errors: DocumentFormValidation;
  }> {
    try {
      const errors: DocumentFormValidation = {};

      // 验证文件
      const fileInput = await this.isVisible('#document-file');
      if (fileInput) {
        const hasFile = await this.executeScript('return document.getElementById("document-file").files.length > 0;');
        if (!hasFile) {
          errors.file = '请选择文件';
        }
      }

      // 验证名称
      const name = await this.getText('#document-name');
      if (!name || name.trim().length === 0) {
        errors.name = '文档名称不能为空';
      }

      // 验证类型
      const type = await this.getText('#document-type');
      if (!type || type.trim().length === 0) {
        errors.type = '请选择文档类型';
      }

      // 验证加密密钥
      const isEncrypted = await this.executeScript('return document.getElementById("document-encryption-checkbox").checked;');
      if (isEncrypted) {
        const encryptionKey = await this.getText('#document-encryption-key');
        if (!encryptionKey || encryptionKey.trim().length === 0) {
          errors.encryptionKey = '加密密钥不能为空';
        }
      }

      // 验证权限
      const permissions = await this.validatePermissions();
      if (!permissions.valid) {
        errors.permissions = permissions.message;
      }

      return {
        valid: Object.keys(errors).length === 0,
        errors
      };

    } catch (error) {
      this.logger.error('验证文档表单失败', { error });
      throw error;
    }
  }

  /**
   * 验证权限设置
   */
  override async validatePermissions(): Promise<{
    valid: boolean;
    message?: string;
  }> {
    try {
      const canView = await this.getText('#document-permissions-view');
      const canEdit = await this.getText('#document-permissions-edit');
      const canDelete = await this.getText('#document-permissions-delete');
      const canShare = await this.getText('#document-permissions-share');

      const viewUsers = canView.split(',').map(u => u.trim()).filter(u => u.length > 0);
      const editUsers = canEdit.split(',').map(u => u.trim()).filter(u => u.length > 0);
      const deleteUsers = canDelete.split(',').map(u => u.trim()).filter(u => u.length > 0);
      const shareUsers = canShare.split(',').map(u => u.trim()).filter(u => u.length > 0);

      // 验证权限逻辑
      if (editUsers.length > 0 && viewUsers.length === 0) {
        return {
          valid: false,
          message: '设置编辑权限时必须设置查看权限'
        };
      }

      if (deleteUsers.length > 0 && viewUsers.length === 0) {
        return {
          valid: false,
          message: '设置删除权限时必须设置查看权限'
        };
      }

      if (shareUsers.length > 0 && viewUsers.length === 0) {
        return {
          valid: false,
          message: '设置共享权限时必须设置查看权限'
        };
      }

      return { valid: true };

    } catch (error) {
      this.logger.error('验证文档权限设置失败', { error });
      throw error;
    }
  }

  /**
   * 验证特定字段
   */
  override async validateFormField(fieldName: keyof DocumentFormValidation, value: string): Promise<{
    valid: boolean;
    message?: string;
  }> {
    try {
      switch (fieldName) {
        case 'name':
          if (!value || value.trim().length === 0) {
            return { valid: false, message: '文档名称不能为空' };
          }
          if (value.length > 200) {
            return { valid: false, message: '文档名称不能超过200个字符' };
          }
          return { valid: true };

        case 'type':
          if (!value || value.trim().length === 0) {
            return { valid: false, message: '请选择文档类型' };
          }
          return { valid: true };

        case 'encryptionKey':
          if (!value || value.trim().length === 0) {
            return { valid: false, message: '加密密钥不能为空' };
          }
          if (value.length < 8) {
            return { valid: false, message: '加密密钥长度至少8位' };
          }
          return { valid: true };

        default:
          return { valid: true };
      }

    } catch (error) {
      this.logger.error('验证文档表单字段失败', { error, fieldName, value });
      throw error;
    }
  }

  /**
   * 获取表单数据
   */
  override async getFormData(): Promise<DocumentFormData> {
    try {
      const data = await this.executeScript(`
        (function() {
          const name = document.getElementById('document-name')?.value || '';
          const type = document.getElementById('document-type')?.value || '';
          const description = document.getElementById('document-description')?.value || '';
          const client = document.getElementById('document-client')?.value || '';
          const caseValue = document.getElementById('document-case')?.value || '';
          const tagsText = document.getElementById('document-tags')?.value || '';
          const isPublic = document.getElementById('document-public-checkbox')?.checked || false;
          const isEncrypted = document.getElementById('document-encryption-checkbox')?.checked || false;
          const encryptionKey = document.getElementById('document-encryption-key')?.value || '';

          const canView = document.getElementById('document-permissions-view')?.value || '';
          const canEdit = document.getElementById('document-permissions-edit')?.value || '';
          const canDelete = document.getElementById('document-permissions-delete')?.value || '';
          const canShare = document.getElementById('document-permissions-share')?.value || '';

          const tags = tagsText.split(',').map(tag => tag.trim()).filter(tag => tag.length > 0);

          const permissions = {
            canView: canView ? canView.split(',').map(u => u.trim()).filter(u => u.length > 0) : [],
            canEdit: canEdit ? canEdit.split(',').map(u => u.trim()).filter(u => u.length > 0) : [],
            canDelete: canDelete ? canDelete.split(',').map(u => u.trim()).filter(u => u.length > 0) : [],
            canShare: canShare ? canShare.split(',').map(u => u.trim()).filter(u => u.length > 0) : []
          };

          return {
            name,
            type,
            description,
            client,
            case: caseValue || undefined,
            tags,
            isPublic,
            isEncrypted,
            encryptionKey: encryptionKey || undefined,
            permissions
          };
        })();
      `);

      return data;

    } catch (error) {
      this.logger.error('获取文档表单数据失败', { error });
      throw error;
    }
  }

  /**
   * 重置表单
   */
  override async resetForm(): Promise<void> {
    try {
      await this.click('#document-reset-button');
      await this.wait(1000); // 等待表单重置
    } catch (error) {
      this.logger.error('重置文档表单失败', { error });
      throw error;
    }
  }

  /**
   * 保存文档
   */
  override async saveDocument(): Promise<void> {
    try {
      await this.click('#document-save-button');
      await this.wait(2000); // 等待保存完成
    } catch (error) {
      this.logger.error('保存文档失败', { error });
      throw error;
    }
  }

  /**
   * 保存并新建
   */
  override async saveAndNew(): Promise<void> {
    try {
      await this.click('#document-save-and-new-button');
      await this.wait(2000); // 等待保存并跳转到新页面
    } catch (error) {
      this.logger.error('保存并新建文档失败', { error });
      throw error;
    }
  }

  /**
   * 取消操作
   */
  override async cancel(): Promise<void> {
    try {
      await this.click('#document-cancel-button');
      await this.wait(1000); // 等待跳转
    } catch (error) {
      this.logger.error('取消文档操作失败', { error });
      throw error;
    }
  }

  /**
   * 验证文档表单页面
   */
  override async validateDocumentFormPage(): Promise<{
    valid: boolean;
    missingElements: string[];
    availableElements: string[];
  }> {
    const requiredElements = [
      '#document-file-upload-button',
      '#document-name',
      '#document-type',
      '#document-description',
      '#document-client',
      '#document-case',
      '#document-tags',
      '#document-public-checkbox',
      '#document-encryption-checkbox',
      '#document-encryption-key',
      '#document-generate-key-button',
      '#document-permissions-view',
      '#document-permissions-edit',
      '#document-permissions-delete',
      '#document-permissions-share',
      '#document-preview-button',
      '#document-extract-metadata-button',
      '#document-save-button',
      '#document-save-and-new-button',
      '#document-cancel-button',
      '#document-reset-button'
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