/**
 * 文档详情页面
 */

import { BasePageObject } from '../core/base-page-object';
import { Logger } from '../core/logger';
import { DocumentDetail, DocumentUpdateData, DocumentListItem } from './document-list-page';

export interface DocumentComment {
  id: string;
  content: string;
  author: string;
  timestamp: Date;
  attachments?: string[];
}

export interface DocumentVersion {
  id: string;
  version: string;
  uploadDate: Date;
  uploadedBy: string;
  size: number;
  comment?: string;
  changes?: string[];
}

export interface DocumentShareSettings {
  isPublic: boolean;
  shareUrl?: string;
  password?: string;
  expirationDate?: Date;
  permissions: {
    canView: string[];
    canEdit: string[];
    canDelete: string[];
    canShare: string[];
  };
}

export class DocumentDetailPage extends BasePageObject {
  private baseUrl: string;

  constructor(config: { baseUrl: string; defaultTimeout?: number; screenshotOnFailure?: boolean }, logger?: Logger) {
    super(config, this.selectors, logger);
    this.baseUrl = config.baseUrl;
  }

  /**
   * 导航到文档详情页面
   */
  override async navigateToDocumentDetail(documentId: string): Promise<void> {
    await this.navigate(`${this.baseUrl}/documents/${documentId}`);
  }

  /**
   * 获取文档详情
   */
  override async getDocumentDetail(): Promise<DocumentDetail> {
    try {
      const detail = await this.executeScript(`
        (function() {
          const id = document.querySelector('.document-detail')?.getAttribute('data-document-id') || '';
          const name = document.querySelector('.document-name')?.gettextContent?.().trim() || '';
          const type = document.querySelector('.document-type')?.gettextContent?.().trim() || '';
          const sizeText = document.querySelector('.document-size')?.gettextContent?.().trim() || '0 B';
          const uploadDateText = document.querySelector('.document-upload-date')?.gettextContent?.().trim() || '';
          const uploadedBy = document.querySelector('.document-uploaded-by')?.gettextContent?.().trim() || '';
          const clientName = document.querySelector('.document-client-name')?.gettextContent?.().trim() || '';
          const clientId = document.querySelector('.document-client')?.getAttribute('data-client-id') || '';
          const status = document.querySelector('.document-status')?.gettextContent?.().trim() || '';
          const description = document.querySelector('.document-description')?.gettextContent?.().trim() || '';

          const tagsElements = document.querySelectorAll('.document-tag');
          const tags = Array.from(tagsElements).map(tag => tag.gettextContent?.().trim() || '');

          const isPublic = document.querySelector('.document-public') !== null;
          const isEncrypted = document.querySelector('.document-encrypted') !== null;
          const encryptionKey = document.querySelector('.document-encryption-key')?.gettextContent?.().trim();

          // 解析文件大小
          const parseSize = (sizeText) => {
            const units = { 'B': 1, 'KB': 1024, 'MB': 1024 * 1024, 'GB': 1024 * 1024 * 1024 };
            const match = sizeText.match(/^([\\d.]+)\\s+(B|KB|MB|GB)$/i);
            if (match) {
              return parseFloat(match[1]) * units[match[2].toUpperCase()];
            }
            return 0;
          };

          // 解析日期
          const parseDate = (dateText) => {
            const date = new Date(dateText);
            return isNaN(date.getTime()) ? new Date() : date;
          };

          // 解析元数据
          const metadata = {};
          const metadataElements = document.querySelectorAll('.document-metadata-item');
          metadataElements.forEach(item => {
            const key = item.querySelector('.metadata-key')?.gettextContent?.().trim() || '';
            const value = item.querySelector('.metadata-value')?.gettextContent?.().trim() || '';
            if (key && value) {
              metadata[key] = value;
            }
          });

          // 解析权限
          const permissions = {
            canView: JSON.parse(document.querySelector('#document-can-view')?.getAttribute('data-permissions') || '[]'),
            canEdit: JSON.parse(document.querySelector('#document-can-edit')?.getAttribute('data-permissions') || '[]'),
            canDelete: JSON.parse(document.querySelector('#document-can-delete')?.getAttribute('data-permissions') || '[]'),
            canShare: JSON.parse(document.querySelector('#document-can-share')?.getAttribute('data-permissions') || '[]')
          };

          // 解析版本历史
          const versions = Array.from(document.querySelectorAll('.document-version-item')).map(item => {
            const versionId = item.getAttribute('data-version-id') || '';
            const version = item.querySelector('.version-number')?.gettextContent?.().trim() || '';
            const uploadDateText = item.querySelector('.version-upload-date')?.gettextContent?.().trim() || '';
            const uploadedBy = item.querySelector('.version-uploaded-by')?.gettextContent?.().trim() || '';
            const sizeText = item.querySelector('.version-size')?.gettextContent?.().trim() || '0 B';
            const comment = item.querySelector('.version-comment')?.gettextContent?.().trim();

            return {
              id: versionId,
              version,
              uploadDate: parseDate(uploadDateText),
              uploadedBy,
              size: parseSize(sizeText),
              comment
            };
          });

          return {
            id,
            name,
            type,
            size: parseSize(sizeText),
            uploadDate: parseDate(uploadDateText),
            uploadedBy,
            client: {
              id: clientId,
              name: clientName
            },
            status,
            description,
            tags,
            metadata,
            permissions,
            versions,
            isPublic,
            isEncrypted,
            encryptionKey
          };
        })();
      `);

      return detail;
    } catch (error) {
      this.logger.error('获取文档详情失败', { error });
      throw error;
    }
  }

  /**
   * 更新文档信息
   */
  override async updateDocument(updateData: DocumentUpdateData): Promise<void> {
    try {
      if (updateData.name) {
        await this.fill('#document-name', updateData.name);
      }

      if (updateData.description) {
        await this.fill('#document-description', updateData.description);
      }

      if (updateData.tags) {
        await this.fill('#document-tags', updateData.tags.join(', '));
      }

      if (updateData.isPublic !== undefined) {
        const publicCheckbox = await this.isVisible('#document-public-checkbox');
        const isChecked = await this.executeScript('return document.getElementById("document-public-checkbox").checked;');

        if (updateData.isPublic !== isChecked) {
          await this.click('#document-public-checkbox');
        }
      }

      if (updateData.permissions) {
        if (updateData.permissions.canView) {
          await this.fill('#document-can-view', updateData.permissions.canView.join(', '));
        }
        if (updateData.permissions.canEdit) {
          await this.fill('#document-can-edit', updateData.permissions.canEdit.join(', '));
        }
        if (updateData.permissions.canDelete) {
          await this.fill('#document-can-delete', updateData.permissions.canDelete.join(', '));
        }
        if (updateData.permissions.canShare) {
          await this.fill('#document-can-share', updateData.permissions.canShare.join(', '));
        }
      }

      // 保存更改
      await this.click('#document-save-button');
      await this.wait(2000); // 等待保存完成

    } catch (error) {
      this.logger.error('更新文档信息失败', { error, updateData });
      throw error;
    }
  }

  /**
   * 预览文档
   */
  override async previewDocument(): Promise<void> {
    try {
      await this.click('#document-preview-button');
      await this.waitForElement('.document-preview-modal', { timeout: 5000 });
    } catch (error) {
      this.logger.error('预览文档失败', { error });
      throw error;
    }
  }

  /**
   * 下载文档
   */
  override async downloadDocument(): Promise<void> {
    try {
      await this.click('#document-download-button');
      await this.wait(2000); // 等待下载开始
    } catch (error) {
      this.logger.error('下载文档失败', { error });
      throw error;
    }
  }

  /**
   * 删除文档
   */
  override async deleteDocument(): Promise<void> {
    try {
      await this.click('#document-delete-button');

      // 等待确认对话框
      await this.waitForElement('.document-delete-confirm-dialog', { timeout: 5000 });

      // 确认删除
      await this.click('.document-delete-confirm-button');
      await this.wait(1000); // 等待删除完成

    } catch (error) {
      this.logger.error('删除文档失败', { error });
      throw error;
    }
  }

  /**
   * 共享文档
   */
  override async shareDocument(shareSettings: DocumentShareSettings): Promise<void> {
    try {
      await this.click('#document-share-button');
      await this.waitForElement('.document-share-dialog', { timeout: 5000 });

      // 设置共享选项
      const publicCheckbox = await this.isVisible('#share-public-checkbox');
      const isChecked = await this.executeScript('return document.getElementById("share-public-checkbox").checked;');

      if (shareSettings.isPublic !== isChecked) {
        await this.click('#share-public-checkbox');
      }

      if (shareSettings.permissions.canView) {
        await this.fill('#share-can-view', shareSettings.permissions.canView.join(', '));
      }
      if (shareSettings.permissions.canEdit) {
        await this.fill('#share-can-edit', shareSettings.permissions.canEdit.join(', '));
      }
      if (shareSettings.permissions.canDelete) {
        await this.fill('#share-can-delete', shareSettings.permissions.canDelete.join(', '));
      }
      if (shareSettings.permissions.canShare) {
        await this.fill('#share-can-share', shareSettings.permissions.canShare.join(', '));
      }

      // 设置过期日期
      if (shareSettings.expirationDate) {
        const expirationDate = shareSettings.expirationDate.toISOString().split('T')[0];
        await this.fill('#share-expiration-date', expirationDate);
      }

      // 设置密码
      if (shareSettings.password) {
        await this.fill('#share-password', shareSettings.password);
      }

      // 生成共享链接
      await this.click('#share-generate-link-button');
      await this.wait(1000);

      // 保存设置
      await this.click('#share-save-button');
      await this.wait(1000);

    } catch (error) {
      this.logger.error('共享文档失败', { error, shareSettings });
      throw error;
    }
  }

  /**
   * 获取文档评论
   */
  override async getComments(): Promise<DocumentComment[]> {
    try {
      const comments = await this.executeScript(`
        (function() {
          const commentElements = Array.from(document.querySelectorAll('.document-comment'));
          return commentElements.map(comment => {
            const id = comment.getAttribute('data-comment-id') || '';
            const content = comment.querySelector('.comment-content')?.gettextContent?.().trim() || '';
            const author = comment.querySelector('.comment-author')?.gettextContent?.().trim() || '';
            const timestampText = comment.querySelector('.comment-timestamp')?.gettextContent?.().trim() || '';

            const attachments = Array.from(comment.querySelectorAll('.comment-attachment')).map(attachment =>
              attachment.getAttribute('data-attachment-id') || ''
            );

            return {
              id,
              content,
              author,
              timestamp: new Date(timestampText),
              attachments
            };
          });
        })();
      `);

      return comments || [];
    } catch (error) {
      this.logger.error('获取文档评论失败', { error });
      throw error;
    }
  }

  /**
   * 添加评论
   */
  override async addComment(content: string): Promise<void> {
    try {
      await this.fill('#document-comment-input', content);
      await this.click('#document-comment-submit-button');
      await this.wait(1000); // 等待评论提交
    } catch (error) {
      this.logger.error('添加文档评论失败', { error });
      throw error;
    }
  }

  /**
   * 删除评论
   */
  override async deleteComment(commentId: string): Promise<void> {
    try {
      const selector = `.document-comment[data-comment-id="${commentId}"] .comment-delete-button`;
      await this.click(selector);
      await this.wait(1000); // 等待删除完成
    } catch (error) {
      this.logger.error('删除文档评论失败', { error, commentId });
      throw error;
    }
  }

  /**
   * 获取版本历史
   */
  override async getVersionHistory(): Promise<DocumentVersion[]> {
    try {
      const versions = await this.executeScript(`
        (function() {
          const versionElements = Array.from(document.querySelectorAll('.document-version-item'));
          return versionElements.map(version => {
            const id = version.getAttribute('data-version-id') || '';
            const versionNumber = version.querySelector('.version-number')?.gettextContent?.().trim() || '';
            const uploadDateText = version.querySelector('.version-upload-date')?.gettextContent?.().trim() || '';
            const uploadedBy = version.querySelector('.version-uploaded-by')?.gettextContent?.().trim() || '';
            const sizeText = version.querySelector('.version-size')?.gettextContent?.().trim() || '0 B';
            const comment = version.querySelector('.version-comment')?.gettextContent?.().trim();

            const parseSize = (sizeText) => {
              const units = { 'B': 1, 'KB': 1024, 'MB': 1024 * 1024, 'GB': 1024 * 1024 * 1024 };
              const match = sizeText.match(/^([\\d.]+)\\s+(B|KB|MB|GB)$/i);
              if (match) {
                return parseFloat(match[1]) * units[match[2].toUpperCase()];
              }
              return 0;
            };

            return {
              id,
              version: versionNumber,
              uploadDate: new Date(uploadDateText),
              uploadedBy,
              size: parseSize(sizeText),
              comment
            };
          });
        })();
      `);

      return versions || [];
    } catch (error) {
      this.logger.error('获取文档版本历史失败', { error });
      throw error;
    }
  }

  /**
   * 查看特定版本
   */
  override async viewVersion(versionId: string): Promise<void> {
    try {
      const selector = `.document-version-item[data-version-id="${versionId}"] .version-view-button`;
      await this.click(selector);
      await this.wait(1000); // 等待版本加载
    } catch (error) {
      this.logger.error('查看文档版本失败', { error, versionId });
      throw error;
    }
  }

  /**
   * 恢复到特定版本
   */
  override async restoreVersion(versionId: string): Promise<void> {
    try {
      const selector = `.document-version-item[data-version-id="${versionId}"] .version-restore-button`;
      await this.click(selector);

      // 等待确认对话框
      await this.waitForElement('.version-restore-confirm-dialog', { timeout: 5000 });

      // 确认恢复
      await this.click('.version-restore-confirm-button');
      await this.wait(2000); // 等待恢复完成

    } catch (error) {
      this.logger.error('恢复文档版本失败', { error, versionId });
      throw error;
    }
  }

  /**
   * 获取关联文档
   */
  override async getRelatedDocuments(): Promise<DocumentListItem[]> {
    try {
      const documents = await this.executeScript(`
        (function() {
          const relatedElements = Array.from(document.querySelectorAll('.related-document-item'));
          return relatedElements.map(item => {
            const id = item.getAttribute('data-document-id') || '';
            const name = item.querySelector('.related-document-name')?.gettextContent?.().trim() || '';
            const type = item.querySelector('.related-document-type')?.gettextContent?.().trim() || '';
            const sizeText = item.querySelector('.related-document-size')?.gettextContent?.().trim() || '0 B';
            const uploadDateText = item.querySelector('.related-document-upload-date')?.gettextContent?.().trim() || '';

            const parseSize = (sizeText) => {
              const units = { 'B': 1, 'KB': 1024, 'MB': 1024 * 1024, 'GB': 1024 * 1024 * 1024 };
              const match = sizeText.match(/^([\\d.]+)\\s+(B|KB|MB|GB)$/i);
              if (match) {
                return parseFloat(match[1]) * units[match[2].toUpperCase()];
              }
              return 0;
            };

            return {
              id,
              name,
              type,
              size: parseSize(sizeText),
              uploadDate: new Date(uploadDateText),
              uploadedBy: '',
              client: '',
              status: '',
              tags: [],
              isPublic: false,
              isEncrypted: false
            };
          });
        })();
      `);

      return documents || [];
    } catch (error) {
      this.logger.error('获取关联文档失败', { error });
      throw error;
    }
  }

  /**
   * 获取活动日志
   */
  override async getActivityLog(): Promise<Array<{
    id: string;
    action: string;
    user: string;
    timestamp: Date;
    details?: string;
  }>> {
    try {
      const activities = await this.executeScript(`
        (function() {
          const activityElements = Array.from(document.querySelectorAll('.document-activity-item'));
          return activityElements.map(activity => {
            const id = activity.getAttribute('data-activity-id') || '';
            const action = activity.querySelector('.activity-action')?.gettextContent?.().trim() || '';
            const user = activity.querySelector('.activity-user')?.gettextContent?.().trim() || '';
            const timestampText = activity.querySelector('.activity-timestamp')?.gettextContent?.().trim() || '';
            const details = activity.querySelector('.activity-details')?.gettextContent?.().trim();

            return {
              id,
              action,
              user,
              timestamp: new Date(timestampText),
              details
            };
          });
        })();
      `);

      return activities || [];
    } catch (error) {
      this.logger.error('获取文档活动日志失败', { error });
      throw error;
    }
  }

  /**
   * 验证文档详情页面
   */
  override async validateDocumentDetailPage(): Promise<{
    valid: boolean;
    missingElements: string[];
    availableElements: string[];
  }> {
    const requiredElements = [
      '.document-detail',
      '.document-name',
      '.document-type',
      '.document-size',
      '.document-upload-date',
      '.document-uploaded-by',
      '.document-client',
      '.document-status',
      '.document-description',
      '.document-tags',
      '.document-preview-button',
      '.document-download-button',
      '.document-edit-button',
      '.document-delete-button',
      '.document-share-button',
      '.document-comment-section',
      '.document-version-history',
      '.document-activity-log',
      '.related-documents'
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