/**
 * 冲突检测页面
 */

import { BasePageObject } from '../core/base-page-object';
import { Logger } from '../core/logger';

export interface ConflictCheckFilters {
  type?: 'client' | 'case' | 'party' | 'opposing_counsel' | 'judge' | 'witness' | 'expert';
  status?: 'pending' | 'reviewing' | 'resolved' | 'approved' | 'rejected';
  severity?: 'low' | 'medium' | 'high' | 'critical';
  dateRange?: {
    startDate: string;
    endDate: string;
  };
  reviewer?: string;
  client?: string;
  case?: string;
  tags?: string[];
}

export interface ConflictCheckSortOptions {
  field: 'checkDate' | 'severity' | 'type' | 'status' | 'client' | 'case' | 'reviewer';
  order: 'asc' | 'desc';
}

export interface ConflictCheckSearchOptions {
  query?: string;
  filters?: ConflictCheckFilters;
  sortOptions?: ConflictCheckSortOptions;
}

export interface ConflictCheckItem {
  id: string;
  checkNumber: string;
  title: string;
  description: string;
  type: 'client' | 'case' | 'party' | 'opposing_counsel' | 'judge' | 'witness' | 'expert';
  checkDate: string;
  checkedBy: string;
  checkedByName: string;
  status: 'pending' | 'reviewing' | 'resolved' | 'approved' | 'rejected';
  severity: 'low' | 'medium' | 'high' | 'critical';
  client: string;
  clientName: string;
  case?: string;
  caseName?: string;
  parties: ConflictParty[];
  conflicts: ConflictItem[];
  reviewer?: string;
  reviewerName?: string;
  reviewDate?: string;
  resolution?: string;
  rejectionReason?: string;
  tags: string[];
  approvedBy?: string;
  approvedDate?: string;
  createdAt: string;
  updatedAt: string;
}

export interface ConflictParty {
  id: string;
  name: string;
  type: 'individual' | 'company' | 'government' | 'organization';
  role: string;
  identifiers: PartyIdentifier[];
  relationships: PartyRelationship[];
}

export interface PartyIdentifier {
  type: 'name' | 'tin' | 'ssn' | 'passport' | 'business_id' | 'phone' | 'email' | 'address';
  value: string;
  country?: string;
  isPrimary: boolean;
}

export interface PartyRelationship {
  type: 'employee' | 'director' | 'shareholder' | 'family_member' | 'business_partner' | 'associate';
  relatedParty: string;
  relatedPartyName: string;
  relationship: string;
  confidence: number;
  source: string;
}

export interface ConflictItem {
  id: string;
  type: 'direct' | 'indirect' | 'potential' | 'historical';
  severity: 'low' | 'medium' | 'high' | 'critical';
  description: string;
  affectedCase?: string;
  affectedCaseName?: string;
  affectedClient?: string;
  affectedClientName?: string;
  conflictParty: string;
  conflictPartyName: string;
  conflictType: string;
  conflictDate?: string;
  details: string;
  sources: ConflictSource[];
  recommendations: string[];
  riskAssessment: RiskAssessment;
}

export interface ConflictSource {
  id: string;
  type: 'internal' | 'external' | 'public' | 'commercial';
  name: string;
  description: string;
  url?: string;
  retrievedAt: string;
  reliability: number;
}

export interface RiskAssessment {
  riskLevel: 'low' | 'medium' | 'high' | 'critical';
  probability: number;
  impact: number;
  mitigatingFactors: string[];
  escalationFactors: string[];
  recommendations: string[];
}

export interface ConflictCheckDetail extends ConflictCheckItem {
  checkMethod: 'manual' | 'automated' | 'hybrid';
  scope: ConflictCheckScope;
  results: ConflictCheckResult;
  documents: ConflictDocument[];
  comments: ConflictComment[];
  history: ConflictHistory[];
  approvalWorkflow: ConflictApprovalWorkflow;
}

export interface ConflictCheckScope {
  includeInternal: boolean;
  includeExternal: boolean;
  includeHistorical: boolean;
  includeRelated: boolean;
  databases: string[];
  searchDepth: 'basic' | 'standard' | 'comprehensive' | 'deep';
  timeRange?: {
    startDate: string;
    endDate: string;
  };
}

export interface ConflictCheckResult {
  totalChecks: number;
  conflictsFound: number;
  highRiskConflicts: number;
  mediumRiskConflicts: number;
  lowRiskConflicts: number;
  processingTime: number;
  databasesSearched: string[];
  confidenceScore: number;
  summary: string;
}

export interface ConflictDocument {
  id: string;
  fileName: string;
  fileType: string;
  fileSize: number;
  uploadedAt: string;
  uploadedBy: string;
  documentType: string;
  description?: string;
  isRelevant: boolean;
}

export interface ConflictComment {
  id: string;
  content: string;
  author: string;
  authorName: string;
  timestamp: string;
  type: 'internal' | 'public';
  isResolution?: boolean;
}

export interface ConflictHistory {
  id: string;
  timestamp: string;
  action: string;
  user: string;
  userName: string;
  details?: string;
  attachments?: string[];
}

export interface ConflictApprovalWorkflow {
  currentStep: number;
  totalSteps: number;
  steps: ConflictApprovalStep[];
  status: 'pending' | 'in_progress' | 'approved' | 'rejected';
}

export interface ConflictApprovalStep {
  step: number;
  type: 'attorney' | 'partner' | 'compliance' | 'admin';
  approver: string;
  approverName: string;
  status: 'pending' | 'approved' | 'rejected';
  actionDate?: string;
  comments?: string;
  required: boolean;
}

export interface ConflictCheckFormData {
  title: string;
  description?: string;
  type: 'client' | 'case' | 'party' | 'opposing_counsel' | 'judge' | 'witness' | 'expert';
  client: string;
  case?: string;
  parties: Omit<ConflictParty, 'id'>[];
  checkMethod: 'manual' | 'automated' | 'hybrid';
  scope: Omit<ConflictCheckScope, 'databases'>;
  tags: string[];
  priority: 'low' | 'medium' | 'high' | 'urgent';
  requestImmediateReview: boolean;
}

export interface ConflictSearchTemplate {
  id: string;
  name: string;
  description: string;
  type: string;
  defaultScope: ConflictCheckScope;
  defaultFilters: ConflictCheckFilters;
  isPublic: boolean;
  createdBy: string;
  createdAt: string;
  usageCount: number;
}

export class ConflictCheckPage extends BasePageObject {
  private baseUrl: string;

  constructor(config: { baseUrl: string; defaultTimeout?: number; screenshotOnFailure?: boolean }, logger?: Logger) {
    super(config, this.selectors, logger);
    this.baseUrl = config.baseUrl;
  }

  /**
   * 导航到冲突检测列表页面
   */
  override async navigateToConflictCheckList(): Promise<void> {
    await this.navigate(`${this.baseUrl}/conflict/checks`);
  }

  /**
   * 导航到创建冲突检测页面
   */
  override async navigateToCreateConflictCheck(): Promise<void> {
    await this.navigate(`${this.baseUrl}/conflict/checks/create`);
  }

  /**
   * 导航到冲突检测详情页面
   */
  override async navigateToConflictCheckDetail(checkId: string): Promise<void> {
    await this.navigate(`${this.baseUrl}/conflict/checks/${checkId}`);
  }

  /**
   * 搜索冲突检测
   */
  override async searchConflictChecks(options: ConflictCheckSearchOptions): Promise<ConflictCheckItem[]> {
    try {
      if (options.query) {
        await this.fill('#conflict-search-input', options.query);
      }

      if (options.filters) {
        await this.applyConflictCheckFilters(options.filters);
      }

      if (options.sortOptions) {
        await this.sortConflictChecks(options.sortOptions);
      }

      await this.click('#conflict-search-button');
      await this.waitForElement('.conflict-check-list-container', { timeout: 5000 });

      const checks = await this.getConflictCheckList();
      this.logger.info('搜索冲突检测完成', { query: options.query, count: checks.length });
      return checks;

    } catch (error) {
      this.logger.error('搜索冲突检测失败', { error, options });
      throw error;
    }
  }

  /**
   * 应用冲突检测过滤器
   */
  override async applyConflictCheckFilters(filters: ConflictCheckFilters): Promise<void> {
    try {
      // 展开过滤器面板
      const filterPanel = await this.isVisible('.conflict-filter-panel');
      if (!filterPanel) {
        await this.click('#conflict-filter-toggle');
      }

      // 类型过滤
      if (filters.type) {
        await this.selectOption('#conflict-type-filter', [filters.type]);
      }

      // 状态过滤
      if (filters.status) {
        await this.selectOption('#conflict-status-filter', [filters.status]);
      }

      // 严重程度过滤
      if (filters.severity) {
        await this.selectOption('#conflict-severity-filter', [filters.severity]);
      }

      // 日期范围
      if (filters.dateRange) {
        await this.fill('#conflict-date-start', filters.dateRange.startDate);
        await this.fill('#conflict-date-end', filters.dateRange.endDate);
      }

      // 审核人过滤
      if (filters.reviewer) {
        await this.selectOption('#conflict-reviewer-filter', [filters.reviewer]);
      }

      // 客户过滤
      if (filters.client) {
        await this.selectOption('#conflict-client-filter', [filters.client]);
      }

      // 案件过滤
      if (filters.case) {
        await this.selectOption('#conflict-case-filter', [filters.case]);
      }

      // 标签过滤
      if (filters.tags && filters.tags.length > 0) {
        await this.fill('#conflict-tags-filter', filters.tags.join(', '));
      }

    } catch (error) {
      this.logger.error('应用冲突检测过滤器失败', { error, filters });
      throw error;
    }
  }

  /**
   * 排序冲突检测
   */
  override async sortConflictChecks(sortOptions: ConflictCheckSortOptions): Promise<void> {
    try {
      await this.click('#conflict-sort-button');
      await this.wait(500);

      // 选择排序字段
      const fieldOption = `#conflict-sort-field-${sortOptions.field}`;
      await this.click(fieldOption);

      // 选择排序顺序
      const orderOption = `#conflict-sort-order-${sortOptions.order}`;
      await this.click(orderOption);

      await this.click('#conflict-sort-apply');
      await this.wait(1000);

    } catch (error) {
      this.logger.error('排序冲突检测失败', { error, sortOptions });
      throw error;
    }
  }

  /**
   * 获取冲突检测列表
   */
  override async getConflictCheckList(): Promise<ConflictCheckItem[]> {
    try {
      const checkElements = await this.executeScript(`
        (function() {
          const items = document.querySelectorAll('.conflict-check-item');
          return Array.from(items).map(item => {
            return {
              id: item.getAttribute('data-id') || '',
              checkNumber: item.getAttribute('data-check-number') || '',
              title: item.querySelector('.conflict-check-title')?.gettextContent?.().trim() || '',
              description: item.querySelector('.conflict-check-description')?.gettextContent?.().trim() || '',
              type: item.getAttribute('data-type') || '',
              checkDate: item.getAttribute('data-check-date') || '',
              checkedBy: item.getAttribute('data-checked-by') || '',
              checkedByName: item.querySelector('.conflict-checked-by-name')?.gettextContent?.().trim() || '',
              status: item.getAttribute('data-status') || '',
              severity: item.getAttribute('data-severity') || '',
              client: item.getAttribute('data-client-id') || '',
              clientName: item.querySelector('.conflict-client-name')?.gettextContent?.().trim() || '',
              case: item.getAttribute('data-case-id') || '',
              caseName: item.querySelector('.conflict-case-name')?.gettextContent?.().trim() || '',
              reviewer: item.getAttribute('data-reviewer-id') || '',
              reviewerName: item.querySelector('.conflict-reviewer-name')?.gettextContent?.().trim() || '',
              reviewDate: item.getAttribute('data-review-date') || '',
              resolution: item.querySelector('.conflict-resolution')?.gettextContent?.().trim() || '',
              rejectionReason: item.querySelector('.conflict-rejection-reason')?.gettextContent?.().trim() || '',
              tags: item.getAttribute('data-tags')?.split(',').filter(tag => tag.length > 0) || [],
              approvedBy: item.getAttribute('data-approved-by') || '',
              approvedDate: item.getAttribute('data-approved-date') || '',
              createdAt: item.getAttribute('data-created-at') || '',
              updatedAt: item.getAttribute('data-updated-at') || '',
              parties: item.getAttribute('data-parties') ? JSON.parse(item.getAttribute('data-parties') || '[]') : [],
              conflicts: item.getAttribute('data-conflicts') ? JSON.parse(item.getAttribute('data-conflicts') || '[]') : []
            };
          });
        })()
      `);

      return checkElements;

    } catch (error) {
      this.logger.error('获取冲突检测列表失败', { error });
      throw error;
    }
  }

  /**
   * 创建冲突检测
   */
  override async createConflictCheck(data: ConflictCheckFormData): Promise<string> {
    try {
      await this.navigateToCreateConflictCheck();
      await this.wait(2000);

      // 基本信息
      await this.fill('#conflict-check-title', data.title);

      if (data.description) {
        await this.fill('#conflict-check-description', data.description);
      }

      await this.selectOption('#conflict-check-type', [data.type]);
      await this.selectOption('#conflict-check-client', [data.client]);

      if (data.case) {
        await this.selectOption('#conflict-check-case', [data.case]);
      }

      // 检测方法
      await this.selectOption('#conflict-check-method', [data.checkMethod]);

      // 优先级
      await this.selectOption('#conflict-check-priority', [data.priority]);

      // 添加相关方
      for (const party of data.parties) {
        await this.addParty(party);
      }

      // 检测范围配置
      await this.configureCheckScope(data.scope);

      // 标签
      if (data.tags && data.tags.length > 0) {
        await this.fill('#conflict-check-tags', data.tags.join(', '));
      }

      // 立即审核请求
      if (data.requestImmediateReview) {
        await this.click('#conflict-immediate-review');
      }

      // 创建检测
      await this.click('#conflict-check-create-button');
      await this.waitForElement('.conflict-check-detail-header', { timeout: 10000 });

      // 获取检测ID
      const checkId = await this.executeScript(`
        return window.location.pathname.split('/').pop();
      `);

      this.logger.info('创建冲突检测成功', { checkId, title: data.title });
      return checkId;

    } catch (error) {
      this.logger.error('创建冲突检测失败', { error, data });
      throw error;
    }
  }

  /**
   * 添加相关方
   */
  override async addParty(party: Omit<ConflictParty, 'id'>): Promise<void> {
    try {
      await this.click('#conflict-add-party-button');
      await this.waitForElement('.party-form-modal', { timeout: 5000 });

      await this.fill('#party-name', party.name);
      await this.selectOption('#party-type', [party.type]);
      await this.fill('#party-role', party.role);

      // 添加标识符
      for (const identifier of party.identifiers) {
        await this.addPartyIdentifier(identifier);
      }

      // 添加关系
      for (const relationship of party.relationships) {
        await this.addPartyRelationship(relationship);
      }

      await this.click('#party-save-button');
      await this.wait(1000);

    } catch (error) {
      this.logger.error('添加相关方失败', { error, party });
      throw error;
    }
  }

  /**
   * 添加相关方标识符
   */
  override async addPartyIdentifier(identifier: PartyIdentifier): Promise<void> {
    try {
      await this.click('#party-add-identifier-button');
      await this.wait(500);

      const lastIndex = await this.executeScript(`
        return document.querySelectorAll('.party-identifier-row').length - 1;
      `);

      await this.selectOption(`#party-identifier-type-${lastIndex}`, [identifier.type]);
      await this.fill(`#party-identifier-value-${lastIndex}`, identifier.value);

      if (identifier.country) {
        await this.fill(`#party-identifier-country-${lastIndex}`, identifier.country);
      }

      if (identifier.isPrimary) {
        await this.click(`#party-identifier-primary-${lastIndex}`);
      }

    } catch (error) {
      this.logger.error('添加相关方标识符失败', { error, identifier });
      throw error;
    }
  }

  /**
   * 添加相关方关系
   */
  override async addPartyRelationship(relationship: PartyRelationship): Promise<void> {
    try {
      await this.click('#party-add-relationship-button');
      await this.wait(500);

      const lastIndex = await this.executeScript(`
        return document.querySelectorAll('.party-relationship-row').length - 1;
      `);

      await this.selectOption(`#party-relationship-type-${lastIndex}`, [relationship.type]);
      await this.fill(`#party-related-party-${lastIndex}`, relationship.relatedParty);
      await this.fill(`#party-relationship-details-${lastIndex}`, relationship.relationship);
      await this.fill(`#party-relationship-confidence-${lastIndex}`, relationship.confidence.toString());

    } catch (error) {
      this.logger.error('添加相关方关系失败', { error, relationship });
      throw error;
    }
  }

  /**
   * 配置检测范围
   */
  override async configureCheckScope(scope: Omit<ConflictCheckScope, 'databases'>): Promise<void> {
    try {
      // 展开 scope 配置面板
      await this.click('#conflict-scope-toggle');
      await this.wait(1000);

      // 基本范围设置
      if (scope.includeInternal) {
        await this.click('#scope-include-internal');
      }

      if (scope.includeExternal) {
        await this.click('#scope-include-external');
      }

      if (scope.includeHistorical) {
        await this.click('#scope-include-historical');
      }

      if (scope.includeRelated) {
        await this.click('#scope-include-related');
      }

      // 搜索深度
      await this.selectOption('#scope-search-depth', [scope.searchDepth]);

      // 时间范围
      if (scope.timeRange) {
        await this.click('#scope-time-range-toggle');
        await this.fill('#scope-time-start', scope.timeRange.startDate);
        await this.fill('#scope-time-end', scope.timeRange.endDate);
      }

    } catch (error) {
      this.logger.error('配置检测范围失败', { error, scope });
      throw error;
    }
  }

  /**
   * 运行冲突检测
   */
  override async runConflictCheck(checkId: string): Promise<void> {
    try {
      await this.navigateToConflictCheckDetail(checkId);

      await this.click('#conflict-run-check-button');
      await this.waitForElement('.conflict-check-progress', { timeout: 5000 });

      // 等待检测完成
      await this.waitForElement('.conflict-check-complete', { timeout: 120000 });

      this.logger.info('运行冲突检测成功', { checkId });

    } catch (error) {
      this.logger.error('运行冲突检测失败', { error, checkId });
      throw error;
    }
  }

  /**
   * 获取冲突检测详情
   */
  override async getConflictCheckDetail(checkId?: string): Promise<ConflictCheckDetail> {
    try {
      if (checkId) {
        await this.navigateToConflictCheckDetail(checkId);
      }

      const detail = await this.executeScript(`
        (function() {
          const parties = Array.from(document.querySelectorAll('.conflict-party')).forEach(party => {
            return {
              id: party.getAttribute('data-id') || '',
              name: party.querySelector('.party-name')?.gettextContent?.().trim() || '',
              type: party.getAttribute('data-type') || '',
              role: party.querySelector('.party-role')?.gettextContent?.().trim() || '',
              identifiers: JSON.parse(party.getAttribute('data-identifiers') || '[]'),
              relationships: JSON.parse(party.getAttribute('data-relationships') || '[]')
            };
          });

          const conflicts = Array.from(document.querySelectorAll('.conflict-item')).forEach(conflict => {
            return {
              id: conflict.getAttribute('data-id') || '',
              type: conflict.getAttribute('data-type') || '',
              severity: conflict.getAttribute('data-severity') || '',
              description: conflict.querySelector('.conflict-description')?.gettextContent?.().trim() || '',
              affectedCase: conflict.getAttribute('data-affected-case') || '',
              affectedCaseName: conflict.querySelector('.conflict-affected-case-name')?.gettextContent?.().trim() || '',
              affectedClient: conflict.getAttribute('data-affected-client') || '',
              affectedClientName: conflict.querySelector('.conflict-affected-client-name')?.gettextContent?.().trim() || '',
              conflictParty: conflict.getAttribute('data-conflict-party') || '',
              conflictPartyName: conflict.querySelector('.conflict-party-name')?.gettextContent?.().trim() || '',
              conflictType: conflict.getAttribute('data-conflict-type') || '',
              conflictDate: conflict.getAttribute('data-conflict-date') || '',
              details: conflict.querySelector('.conflict-details')?.gettextContent?.().trim() || '',
              sources: JSON.parse(conflict.getAttribute('data-sources') || '[]'),
              recommendations: conflict.getAttribute('data-recommendations')?.split(',').filter(r => r.length > 0) || [],
              riskAssessment: JSON.parse(conflict.getAttribute('data-risk-assessment') || '{}')
            };
          });

          const documents = Array.from(document.querySelectorAll('.conflict-document')).forEach(doc => {
            return {
              id: doc.getAttribute('data-id') || '',
              fileName: doc.querySelector('.document-name')?.gettextContent?.().trim() || '',
              fileType: doc.getAttribute('data-type') || '',
              fileSize: parseInt(doc.getAttribute('data-size') || '0'),
              uploadedAt: doc.getAttribute('data-uploaded-at') || '',
              uploadedBy: doc.getAttribute('data-uploaded-by') || '',
              documentType: doc.getAttribute('data-document-type') || '',
              description: doc.querySelector('.document-description')?.gettextContent?.().trim() || '',
              isRelevant: doc.getAttribute('data-is-relevant') === 'true'
            };
          });

          const comments = Array.from(document.querySelectorAll('.conflict-comment')).forEach(comment => {
            return {
              id: comment.getAttribute('data-id') || '',
              content: comment.querySelector('.comment-content')?.gettextContent?.().trim() || '',
              author: comment.getAttribute('data-author') || '',
              authorName: comment.querySelector('.comment-author-name')?.gettextContent?.().trim() || '',
              timestamp: comment.getAttribute('data-timestamp') || '',
              type: comment.getAttribute('data-type') || 'internal',
              isResolution: comment.getAttribute('data-is-resolution') === 'true'
            };
          });

          const history = Array.from(document.querySelectorAll('.conflict-history-item')).forEach(item => {
            return {
              id: item.getAttribute('data-id') || '',
              timestamp: item.getAttribute('data-timestamp') || '',
              action: item.querySelector('.history-action')?.gettextContent?.().trim() || '',
              user: item.getAttribute('data-user') || '',
              userName: item.querySelector('.history-user-name')?.gettextContent?.().trim() || '',
              details: item.querySelector('.history-details')?.gettextContent?.().trim() || '',
              attachments: item.getAttribute('data-attachments')?.split(',').filter(a => a.length > 0) || []
            };
          });

          const workflow = {
            currentStep: parseInt(document.getElementById('conflict-workflow-current-step')?.gettextContent?.().trim() || '0'),
            totalSteps: parseInt(document.getElementById('conflict-workflow-total-steps')?.gettextContent?.().trim() || '0'),
            status: document.getElementById('conflict-workflow-status')?.gettextContent?.().trim() || 'pending',
            steps: Array.from(document.querySelectorAll('.conflict-approval-step')).map(step => ({
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

          const scope = {
            includeInternal: document.getElementById('scope-include-internal')?.gettextContent?.().trim() === 'true',
            includeExternal: document.getElementById('scope-include-external')?.gettextContent?.().trim() === 'true',
            includeHistorical: document.getElementById('scope-include-historical')?.gettextContent?.().trim() === 'true',
            includeRelated: document.getElementById('scope-include-related')?.gettextContent?.().trim() === 'true',
            databases: JSON.parse(document.getElementById('scope-databases')?.textContent || '[]'),
            searchDepth: document.getElementById('scope-search-depth')?.gettextContent?.().trim() || 'standard',
            timeRange: document.getElementById('scope-time-range') ? {
              startDate: document.getElementById('scope-time-start')?.gettextContent?.().trim() || '',
              endDate: document.getElementById('scope-time-end')?.gettextContent?.().trim() || ''
            } : undefined
          };

          const results = {
            totalChecks: parseInt(document.getElementById('conflict-total-checks')?.gettextContent?.().trim() || '0'),
            conflictsFound: parseInt(document.getElementById('conflict-conflicts-found')?.gettextContent?.().trim() || '0'),
            highRiskConflicts: parseInt(document.getElementById('conflict-high-risk-conflicts')?.gettextContent?.().trim() || '0'),
            mediumRiskConflicts: parseInt(document.getElementById('conflict-medium-risk-conflicts')?.gettextContent?.().trim() || '0'),
            lowRiskConflicts: parseInt(document.getElementById('conflict-low-risk-conflicts')?.gettextContent?.().trim() || '0'),
            processingTime: parseFloat(document.getElementById('conflict-processing-time')?.gettextContent?.().trim() || '0'),
            databasesSearched: JSON.parse(document.getElementById('conflict-databases-searched')?.textContent || '[]'),
            confidenceScore: parseFloat(document.getElementById('conflict-confidence-score')?.gettextContent?.().trim() || '0'),
            summary: document.getElementById('conflict-summary')?.gettextContent?.().trim() || ''
          };

          return {
            id: document.getElementById('conflict-check-id')?.gettextContent?.().trim() || '',
            checkNumber: document.getElementById('conflict-check-number')?.gettextContent?.().trim() || '',
            title: document.getElementById('conflict-check-title')?.gettextContent?.().trim() || '',
            description: document.getElementById('conflict-check-description')?.gettextContent?.().trim() || '',
            type: document.getElementById('conflict-check-type')?.gettextContent?.().trim() || '',
            checkDate: document.getElementById('conflict-check-date')?.gettextContent?.().trim() || '',
            checkedBy: document.getElementById('conflict-checked-by')?.gettextContent?.().trim() || '',
            checkedByName: document.getElementById('conflict-checked-by-name')?.gettextContent?.().trim() || '',
            status: document.getElementById('conflict-check-status')?.gettextContent?.().trim() || '',
            severity: document.getElementById('conflict-check-severity')?.gettextContent?.().trim() || '',
            client: document.getElementById('conflict-check-client-id')?.gettextContent?.().trim() || '',
            clientName: document.getElementById('conflict-check-client-name')?.gettextContent?.().trim() || '',
            case: document.getElementById('conflict-check-case-id')?.gettextContent?.().trim() || '',
            caseName: document.getElementById('conflict-check-case-name')?.gettextContent?.().trim() || '',
            reviewer: document.getElementById('conflict-check-reviewer-id')?.gettextContent?.().trim() || '',
            reviewerName: document.getElementById('conflict-check-reviewer-name')?.gettextContent?.().trim() || '',
            reviewDate: document.getElementById('conflict-check-review-date')?.gettextContent?.().trim() || '',
            resolution: document.getElementById('conflict-check-resolution')?.gettextContent?.().trim() || '',
            rejectionReason: document.getElementById('conflict-check-rejection-reason')?.gettextContent?.().trim() || '',
            tags: document.getElementById('conflict-check-tags')?.gettextContent?.().trim().split(',').filter(tag => tag.length > 0) || [],
            approvedBy: document.getElementById('conflict-check-approved-by')?.gettextContent?.().trim() || '',
            approvedDate: document.getElementById('conflict-check-approved-date')?.gettextContent?.().trim() || '',
            createdAt: document.getElementById('conflict-check-created-at')?.gettextContent?.().trim() || '',
            updatedAt: document.getElementById('conflict-check-updated-at')?.gettextContent?.().trim() || '',
            parties: parties,
            conflicts: conflicts,
            checkMethod: document.getElementById('conflict-check-method')?.gettextContent?.().trim() || '',
            scope: scope,
            results: results,
            documents: documents,
            comments: comments,
            history: history,
            approvalWorkflow: workflow
          };
        })()
      `);

      return detail;

    } catch (error) {
      this.logger.error('获取冲突检测详情失败', { error, checkId });
      throw error;
    }
  }

  /**
   * 审核冲突检测
   */
  override async reviewConflictCheck(checkId: string, reviewData: {
    action: 'approve' | 'reject' | 'request_changes';
    resolution?: string;
    comments?: string;
    assignReviewer?: string;
  }): Promise<void> {
    try {
      await this.navigateToConflictCheckDetail(checkId);

      if (reviewData.action === 'approve') {
        await this.click('#conflict-approve-button');
      } else if (reviewData.action === 'reject') {
        await this.click('#conflict-reject-button');
      } else {
        await this.click('#conflict-request-changes-button');
      }

      await this.waitForElement('.conflict-review-modal', { timeout: 5000 });

      if (reviewData.resolution) {
        await this.fill('#conflict-resolution', reviewData.resolution);
      }

      if (reviewData.comments) {
        await this.fill('#conflict-review-comments', reviewData.comments);
      }

      if (reviewData.assignReviewer) {
        await this.selectOption('#conflict-assign-reviewer', [reviewData.assignReviewer]);
      }

      await this.click('#conflict-review-confirm');
      await this.waitForElement('.conflict-review-confirmation', { timeout: 5000 });

      this.logger.info('审核冲突检测成功', { checkId, action: reviewData.action });

    } catch (error) {
      this.logger.error('审核冲突检测失败', { error, checkId, reviewData });
      throw error;
    }
  }

  /**
   * 使用模板创建冲突检测
   */
  override async createFromTemplate(templateId: string, overrides: {
    title?: string;
    client?: string;
    case?: string;
    parties?: Omit<ConflictParty, 'id'>[];
    scope?: Omit<ConflictCheckScope, 'databases'>;
  } = {}): Promise<string> {
    try {
      await this.navigateToCreateConflictCheck();

      await this.click('#conflict-use-template-button');
      await this.waitForElement('.conflict-template-selector', { timeout: 5000 });

      await this.click(`#conflict-template-${templateId}`);
      await this.click('#conflict-template-apply');
      await this.wait(2000);

      // 应用覆盖设置
      if (overrides.title) {
        await this.fill('#conflict-check-title', overrides.title);
      }

      if (overrides.client) {
        await this.selectOption('#conflict-check-client', [overrides.client]);
      }

      if (overrides.case) {
        await this.selectOption('#conflict-check-case', [overrides.case]);
      }

      if (overrides.parties) {
        for (const party of overrides.parties) {
          await this.addParty(party);
        }
      }

      if (overrides.scope) {
        await this.configureCheckScope(overrides.scope);
      }

      await this.click('#conflict-check-create-button');
      await this.waitForElement('.conflict-check-detail-header', { timeout: 10000 });

      const checkId = await this.executeScript(`
        return window.location.pathname.split('/').pop();
      `);

      this.logger.info('从模板创建冲突检测成功', { templateId, checkId });
      return checkId;

    } catch (error) {
      this.logger.error('从模板创建冲突检测失败', { error, templateId, overrides });
      throw error;
    }
  }

  /**
   * 获取冲突检测模板
   */
  override async getConflictCheckTemplates(): Promise<ConflictSearchTemplate[]> {
    try {
      const templates = await this.executeScript(`
        (function() {
          return Array.from(document.querySelectorAll('.conflict-template-item')).map(item => {
            return {
              id: item.getAttribute('data-id') || '',
              name: item.querySelector('.template-name')?.gettextContent?.().trim() || '',
              description: item.querySelector('.template-description')?.gettextContent?.().trim() || '',
              type: item.getAttribute('data-type') || '',
              defaultScope: JSON.parse(item.getAttribute('data-default-scope') || '{}'),
              defaultFilters: JSON.parse(item.getAttribute('data-default-filters') || '{}'),
              isPublic: item.getAttribute('data-is-public') === 'true',
              createdBy: item.getAttribute('data-created-by') || '',
              createdAt: item.getAttribute('data-created-at') || '',
              usageCount: parseInt(item.getAttribute('data-usage-count') || '0')
            };
          });
        })()
      `);

      return templates;

    } catch (error) {
      this.logger.error('获取冲突检测模板失败', { error });
      throw error;
    }
  }

  /**
   * 获取冲突检测统计
   */
  override async getConflictCheckStatistics(): Promise<{
    totalChecks: number;
    pendingChecks: number;
    highRiskConflicts: number;
    mediumRiskConflicts: number;
    lowRiskConflicts: number;
    averageProcessingTime: number;
    byType: Record<string, number>;
    byStatus: Record<string, number>;
    byMonth: Array<{ month: string; count: number; conflictsFound: number }>;
    topReviewers: Array<{ reviewer: string; count: number; averageTime: number }>;
  }> {
    try {
      const statistics = await this.executeScript(`
        (function() {
          return {
            totalChecks: parseInt(document.getElementById('conflict-total-checks')?.gettextContent?.().trim() || '0'),
            pendingChecks: parseInt(document.getElementById('conflict-pending-checks')?.gettextContent?.().trim() || '0'),
            highRiskConflicts: parseInt(document.getElementById('conflict-high-risk-conflicts')?.gettextContent?.().trim() || '0'),
            mediumRiskConflicts: parseInt(document.getElementById('conflict-medium-risk-conflicts')?.gettextContent?.().trim() || '0'),
            lowRiskConflicts: parseInt(document.getElementById('conflict-low-risk-conflicts')?.gettextContent?.().trim() || '0'),
            averageProcessingTime: parseFloat(document.getElementById('conflict-avg-processing-time')?.gettextContent?.().trim() || '0'),
            byType: JSON.parse(document.getElementById('conflict-stats-by-type')?.textContent || '{}'),
            byStatus: JSON.parse(document.getElementById('conflict-stats-by-status')?.textContent || '{}'),
            byMonth: JSON.parse(document.getElementById('conflict-stats-by-month')?.textContent || '[]'),
            topReviewers: JSON.parse(document.getElementById('conflict-top-reviewers')?.textContent || '[]')
          };
        })()
      `);

      return statistics;

    } catch (error) {
      this.logger.error('获取冲突检测统计失败', { error });
      throw error;
    }
  }

  /**
   * 验证冲突检测列表页面
   */
  override async validateConflictCheckListPage(): Promise<{
    valid: boolean;
    missingElements: string[];
    availableElements: string[];
  }> {
    const requiredElements = [
      '#conflict-search-input',
      '#conflict-search-button',
      '#conflict-filter-toggle',
      '#conflict-sort-button',
      '#conflict-create-button',
      '.conflict-check-list-container',
      '.conflict-check-item',
      '#conflict-stats-container',
      '#conflict-templates-button',
      '#conflict-export-button'
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
   * 验证冲突检测详情页面
   */
  override async validateConflictCheckDetailPage(): Promise<{
    valid: boolean;
    missingElements: string[];
    availableElements: string[];
  }> {
    const requiredElements = [
      '.conflict-check-detail-header',
      '#conflict-check-title',
      '#conflict-check-type',
      '#conflict-check-date',
      '#conflict-check-status',
      '#conflict-check-severity',
      '#conflict-check-client-name',
      '#conflict-parties-section',
      '#conflict-conflicts-section',
      '#conflict-run-check-button',
      '#conflict-approve-button',
      '#conflict-reject-button',
      '#conflict-request-changes-button',
      '.conflict-check-results',
      '.conflict-check-documents',
      '.conflict-check-comments',
      '.conflict-check-history'
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