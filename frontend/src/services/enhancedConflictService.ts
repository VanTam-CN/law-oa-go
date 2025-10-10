import Fuse from 'fuse.js';

// 冲突检测配置接口
export interface ConflictCheckConfig {
  searchYears: number;
  includeCorporateRelations: boolean;
  searchDepth: 'BASIC' | 'STANDARD' | 'DEEP';
  threshold: number; // 相似度阈值
  distance: number; // 编辑距离
  location: number; // 匹配位置权重
}

// 冲突匹配结果
export interface ConflictMatch {
  caseId: string;
  caseName: string;
  conflictType: string;
  riskLevel: 'CRITICAL' | 'HIGH' | 'MEDIUM' | 'LOW';
  matchScore: number;
  matchReasons: string[];
  conflictDetails: string;
  caseStatus: string;
  clientId: string;
  opposingParties: string[];
  matchedEntities: MatchedEntity[];
}

// 匹配的实体
export interface MatchedEntity {
  entityId: string;
  entityType: string;
  entityName: string;
  matchType: 'EXACT' | 'FUZZY' | 'PHONETIC' | 'RELATED';
  matchScore: number;
  originalName: string;
  matchedName: string;
}

// 冲突分析结果
export interface ConflictAnalysis {
  totalCasesChecked: number;
  conflictMatches: ConflictMatch[];
  highRiskMatches: ConflictMatch[];
  mediumRiskMatches: ConflictMatch[];
  lowRiskMatches: ConflictMatch[];
  clientHistoryCases: number;
  relatedPartiesChecked: number;
  corporateRelations: number;
  searchTimeRange: string;
  analysisDuration: number;
  riskAssessment: RiskAssessment;
  recommendations: string[];
}

// 风险评估
export interface RiskAssessment {
  overallRisk: string;
  riskScore: number;
  riskFactors: string[];
  requiresApproval: boolean;
  approvalLevel: string;
  mitigation: string[];
}

// 历史案例数据接口
export interface HistoricalCase {
  id: string;
  title: string;
  description: string;
  clientId: string;
  clientName: string;
  clientType: 'PERSON' | 'COMPANY';
  caseType: string;
  status: string;
  priority: string;
  lawyerId: string;
  lawyerName: string;
  createdAt: string;
  updatedAt: string;
  opposingParties?: string[];
  caseDetails?: string;
}

// 增强的冲突检测服务
export class EnhancedConflictService {
  private fuse: Fuse<HistoricalCase>;
  private historicalCases: HistoricalCase[] = [];
  private config: ConflictCheckConfig;

  constructor(config?: Partial<ConflictCheckConfig>) {
    this.config = {
      searchYears: 5,
      includeCorporateRelations: false,
      searchDepth: 'STANDARD',
      threshold: 0.6,
      distance: 100,
      location: 0,
      ...config
    };

    // 初始化Fuse.js配置
    this.fuse = new Fuse([], {
      keys: [
        {
          name: 'title',
          weight: 2.0
        },
        {
          name: 'clientName',
          weight: 3.0
        },
        {
          name: 'description',
          weight: 1.0
        },
        {
          name: 'opposingParties',
          weight: 2.5
        }
      ],
      threshold: this.config.threshold,
      includeScore: true,
      includeMatches: true,
      minMatchCharLength: 2,
      ignoreLocation: false,
      ignoreDiacritics: true,
      findAllMatches: true,
      location: this.config.location,
      distance: this.config.distance,
      useExtendedSearch: true,
      getFn: (obj, path) => {
        if (path === 'opposingParties' && obj.opposingParties) {
          return obj.opposingParties.join(' ');
        }
        return this.getNestedValue(obj, path);
      }
    });
  }

  // 设置历史案例数据
  setHistoricalCases(cases: HistoricalCase[]): void {
    this.historicalCases = cases;
    this.fuse.setCollection(cases);
  }

  // 获取嵌套属性值
  private getNestedValue(obj: any, path: string): string {
    return path.split('.').reduce((current, key) => current?.[key], obj) || '';
  }

  // 标准化实体名称
  private normalizeEntityName(name: string): string {
    if (!name) return '';

    return name
      .toLowerCase()
      .replace(/\s+/g, ' ')
      .replace(/[^\w\s\u4e00-\u9fff]/g, '') // 保留中英文字符
      .replace(/(?:co|ltd|inc|corp|llc|limited|company|corporation|公司|有限|股份|集团|控股|投资|实业|科技)$/g, '')
      .trim();
  }

  // 计算相似度分数
  private calculateSimilarityScore(str1: string, str2: string): number {
    if (!str1 || !str2) return 0;

    const normalized1 = this.normalizeEntityName(str1);
    const normalized2 = this.normalizeEntityName(str2);

    if (normalized1 === normalized2) return 1.0;

    // 简化的Levenshtein距离计算
    const distance = this.levenshteinDistance(normalized1, normalized2);
    const maxLen = Math.max(normalized1.length, normalized2.length);

    return maxLen > 0 ? 1 - (distance / maxLen) : 0;
  }

  // Levenshtein距离计算
  private levenshteinDistance(str1: string, str2: string): number {
    const matrix = Array(str2.length + 1).fill(null).map(() =>
      Array(str1.length + 1).fill(null)
    );

    for (let i = 0; i <= str1.length; i++) matrix[0][i] = i;
    for (let j = 0; j <= str2.length; j++) matrix[j][0] = j;

    for (let j = 1; j <= str2.length; j++) {
      for (let i = 1; i <= str1.length; i++) {
        const indicator = str1[i - 1] === str2[j - 1] ? 0 : 1;
        matrix[j][i] = Math.min(
          matrix[j][i - 1] + 1, // deletion
          matrix[j - 1][i] + 1, // insertion
          matrix[j - 1][i - 1] + indicator // substitution
        );
      }
    }

    return matrix[str2.length][str1.length];
  }

  // 语音匹配（简化的Soundex算法）
  private isPhoneticMatch(name1: string, name2: string): boolean {
    if (!name1 || !name2) return false;

    const soundex1 = this.simpleSoundex(name1);
    const soundex2 = this.simpleSoundex(name2);

    return soundex1 === soundex2;
  }

  private simpleSoundex(name: string): string {
    if (!name) return '0000';

    const normalized = this.normalizeEntityName(name);
    if (normalized.length === 0) return '0000';

    const soundexMap: { [key: string]: string } = {
      'b': '1', 'f': '1', 'p': '1', 'v': '1',
      'c': '2', 'g': '2', 'j': '2', 'k': '2', 'q': '2', 's': '2', 'x': '2', 'z': '2',
      'd': '3', 't': '3',
      'l': '4',
      'm': '5', 'n': '5',
      'r': '6'
    };

    let result = normalized[0].toUpperCase();
    let code = '';

    for (let i = 1; i < normalized.length && code.length < 3; i++) {
      const char = normalized[i];
      const mapped = soundexMap[char];

      if (mapped && (code.length === 0 || code[code.length - 1] !== mapped)) {
        code += mapped;
      }
    }

    while (code.length < 3) {
      code += '0';
    }

    return result + code;
  }

  // 执行增强的冲突检测
  async checkConflict(request: {
    clientId: string;
    clientName: string;
    clientType: 'PERSON' | 'COMPANY';
    otherParties: string[];
    caseName: string;
    caseType: string;
  }): Promise<ConflictAnalysis> {
    const startTime = Date.now();

    // 1. 精确匹配
    const exactMatches = this.performExactMatching(request);

    // 2. 模糊匹配
    const fuzzyMatches = this.performFuzzyMatching(request);

    // 3. 语音匹配
    const phoneticMatches = this.performPhoneticMatching(request);

    // 4. 实体关联匹配
    const entityMatches = this.performEntityMatching(request);

    // 合并和去重
    const allMatches = this.deduplicateMatches([
      ...exactMatches,
      ...fuzzyMatches,
      ...phoneticMatches,
      ...entityMatches
    ]);

    // 风险评估
    const riskAssessment = this.performRiskAssessment(allMatches);

    // 生成建议
    const recommendations = this.generateRecommendations(riskAssessment, allMatches);

    return {
      totalCasesChecked: this.historicalCases.length,
      conflictMatches: allMatches,
      highRiskMatches: allMatches.filter(m => m.riskLevel === 'CRITICAL' || m.riskLevel === 'HIGH'),
      mediumRiskMatches: allMatches.filter(m => m.riskLevel === 'MEDIUM'),
      lowRiskMatches: allMatches.filter(m => m.riskLevel === 'LOW'),
      clientHistoryCases: this.historicalCases.filter(c => c.clientId === request.clientId).length,
      relatedPartiesChecked: request.otherParties.length,
      corporateRelations: 0, // TODO: 实现企业关联检查
      searchTimeRange: `${this.config.searchYears}年`,
      analysisDuration: Date.now() - startTime,
      riskAssessment,
      recommendations
    };
  }

  // 精确匹配
  private performExactMatching(request: any): ConflictMatch[] {
    const matches: ConflictMatch[] = [];

    for (const historicalCase of this.historicalCases) {
      // 客户名称精确匹配
      if (this.normalizeEntityName(request.clientName) ===
          this.normalizeEntityName(historicalCase.clientName)) {
        matches.push({
          caseId: historicalCase.id,
          caseName: historicalCase.title,
          conflictType: 'CLIENT_NAME_EXACT',
          riskLevel: 'HIGH',
          matchScore: 1.0,
          matchReasons: ['客户名称完全匹配'],
          conflictDetails: `客户"${request.clientName}"与历史案件中的客户完全匹配`,
          caseStatus: historicalCase.status,
          clientId: historicalCase.clientId,
          opposingParties: historicalCase.opposingParties || [],
          matchedEntities: [{
            entityId: request.clientId,
            entityType: request.clientType,
            entityName: request.clientName,
            matchType: 'EXACT',
            matchScore: 1.0,
            originalName: request.clientName,
            matchedName: historicalCase.clientName
          }]
        });
      }

      // 对方当事人精确匹配
      for (const party of request.otherParties) {
        if (historicalCase.opposingParties?.some(op =>
            this.normalizeEntityName(party) === this.normalizeEntityName(op))) {
          matches.push({
            caseId: historicalCase.id,
            caseName: historicalCase.title,
            conflictType: 'OPPOSING_PARTY_EXACT',
            riskLevel: 'CRITICAL',
            matchScore: 1.0,
            matchReasons: [`对方当事人"${party}"完全匹配`],
            conflictDetails: `对方当事人"${party}"与历史案件中的当事人完全匹配`,
            caseStatus: historicalCase.status,
            clientId: historicalCase.clientId,
            opposingParties: historicalCase.opposingParties || [],
            matchedEntities: []
          });
        }
      }
    }

    return matches;
  }

  // 模糊匹配
  private performFuzzyMatching(request: any): ConflictMatch[] {
    const matches: ConflictMatch[] = [];

    // 使用Fuse.js进行模糊匹配
    const fuseResults = this.fuse.search(request.clientName, {
      threshold: 0.8,
      includeScore: true
    });

    for (const result of fuseResults) {
      if (result.score && result.score < 0.4) { // Fuse.js的score越低越好
        const similarity = 1 - result.score;
        if (similarity >= 0.8) {
          matches.push({
            caseId: result.item.id,
            caseName: result.item.title,
            conflictType: 'CLIENT_NAME_FUZZY',
            riskLevel: 'MEDIUM',
            matchScore: similarity,
            matchReasons: [`客户名称相似度: ${(similarity * 100).toFixed(1)}%`],
            conflictDetails: `客户名称"${request.clientName}"与历史客户名称相似`,
            caseStatus: result.item.status,
            clientId: result.item.clientId,
            opposingParties: result.item.opposingParties || [],
            matchedEntities: [{
              entityId: request.clientId,
              entityType: request.clientType,
              entityName: request.clientName,
              matchType: 'FUZZY',
              matchScore: similarity,
              originalName: request.clientName,
              matchedName: result.item.clientName
            }]
          });
        }
      }
    }

    return matches;
  }

  // 语音匹配
  private performPhoneticMatching(request: any): ConflictMatch[] {
    const matches: ConflictMatch[] = [];

    for (const historicalCase of this.historicalCases) {
      if (this.isPhoneticMatch(request.clientName, historicalCase.clientName)) {
        matches.push({
          caseId: historicalCase.id,
          caseName: historicalCase.title,
          conflictType: 'CLIENT_NAME_PHONETIC',
          riskLevel: 'LOW',
          matchScore: 0.7,
          matchReasons: ['客户名称语音相似'],
          conflictDetails: `客户名称"${request.clientName}"与历史客户名称语音相似`,
          caseStatus: historicalCase.status,
          clientId: historicalCase.clientId,
          opposingParties: historicalCase.opposingParties || [],
          matchedEntities: [{
            entityId: request.clientId,
            entityType: request.clientType,
            entityName: request.clientName,
            matchType: 'PHONETIC',
            matchScore: 0.7,
            originalName: request.clientName,
            matchedName: historicalCase.clientName
          }]
        });
      }
    }

    return matches;
  }

  // 实体关联匹配
  private performEntityMatching(request: any): ConflictMatch[] {
    const matches: ConflictMatch[] = [];

    for (const party of request.otherParties) {
      for (const historicalCase of this.historicalCases) {
        // 检查相关方与历史案例的匹配
        const clientSimilarity = this.calculateSimilarityScore(party, historicalCase.clientName);

        if (clientSimilarity >= 0.8) {
          matches.push({
            caseId: historicalCase.id,
            caseName: historicalCase.title,
            conflictType: 'RELATED_PARTY_MATCH',
            riskLevel: 'MEDIUM',
            matchScore: clientSimilarity,
            matchReasons: [`相关方"${party}"匹配度: ${(clientSimilarity * 100).toFixed(1)}%`],
            conflictDetails: `相关方"${party}"与历史案件的客户名称匹配`,
            caseStatus: historicalCase.status,
            clientId: historicalCase.clientId,
            opposingParties: historicalCase.opposingParties || [],
            matchedEntities: []
          });
        }
      }
    }

    return matches;
  }

  // 去重
  private deduplicateMatches(matches: ConflictMatch[]): ConflictMatch[] {
    const seen = new Set<string>();
    return matches.filter(match => {
      const key = `${match.caseId}_${match.conflictType}`;
      if (seen.has(key)) {
        return false;
      }
      seen.add(key);
      return true;
    });
  }

  // 风险评估
  private performRiskAssessment(matches: ConflictMatch[]): RiskAssessment {
    if (matches.length === 0) {
      return {
        overallRisk: 'LOW',
        riskScore: 0.0,
        riskFactors: [],
        requiresApproval: false,
        approvalLevel: '',
        mitigation: ['未发现明显冲突，建议正常处理']
      };
    }

    let riskScore = 0.0;
    const riskFactors: string[] = [];

    for (const match of matches) {
      switch (match.riskLevel) {
        case 'CRITICAL':
          riskScore += 0.4;
          riskFactors.push(`关键冲突: ${match.caseName}`);
          break;
        case 'HIGH':
          riskScore += 0.3;
          riskFactors.push(`高风险冲突: ${match.caseName}`);
          break;
        case 'MEDIUM':
          riskScore += 0.2;
          riskFactors.push(`中等风险冲突: ${match.caseName}`);
          break;
        case 'LOW':
          riskScore += 0.1;
          break;
      }
    }

    if (riskScore > 1.0) riskScore = 1.0;

    let overallRisk = 'LOW';
    let requiresApproval = false;
    let approvalLevel = '';

    if (riskScore >= 0.7) {
      overallRisk = 'CRITICAL';
      requiresApproval = true;
      approvalLevel = 'SENIOR_PARTNER';
    } else if (riskScore >= 0.5) {
      overallRisk = 'HIGH';
      requiresApproval = true;
      approvalLevel = 'PARTNER';
    } else if (riskScore >= 0.3) {
      overallRisk = 'MEDIUM';
      requiresApproval = false;
    }

    return {
      overallRisk,
      riskScore,
      riskFactors,
      requiresApproval,
      approvalLevel,
      mitigation: this.generateMitigation(overallRisk)
    };
  }

  // 生成缓解措施
  private generateMitigation(riskLevel: string): string[] {
    switch (riskLevel) {
      case 'CRITICAL':
        return [
          '立即停止案件受理',
          '要求高级合伙人审查',
          '考虑是否需要拒绝代理',
          '详细记录冲突情况'
        ];
      case 'HIGH':
        return [
          '要求合伙人级别审查',
          '获取客户书面同意',
          '建立信息隔离墙',
          '持续监控潜在冲突'
        ];
      case 'MEDIUM':
        return [
          '要求主管律师审查',
          '加强内部信息管理',
          '定期更新冲突检查'
        ];
      default:
        return ['未发现明显冲突，建议正常处理'];
    }
  }

  // 生成建议
  private generateRecommendations(
    assessment: RiskAssessment,
    matches: ConflictMatch[]
  ): string[] {
    const recommendations = [...assessment.mitigation];

    if (assessment.requiresApproval) {
      recommendations.push(`需要${assessment.approvalLevel}级别批准`);
    }

    if (assessment.riskScore >= 0.5) {
      recommendations.push('建议咨询风险管理委员会');
    }

    return recommendations;
  }

  // 配置更新
  updateConfig(newConfig: Partial<ConflictCheckConfig>): void {
    this.config = { ...this.config, ...newConfig };

    // 更新Fuse.js配置
    this.fuse.options = {
      ...this.fuse.options,
      threshold: this.config.threshold,
      location: this.config.location,
      distance: this.config.distance
    };
  }

  // 获取配置
  getConfig(): ConflictCheckConfig {
    return { ...this.config };
  }
}

// 创建单例实例
export const enhancedConflictService = new EnhancedConflictService({
  searchYears: 5,
  includeCorporateRelations: false,
  searchDepth: 'STANDARD',
  threshold: 0.6,
  distance: 100,
  location: 0
});