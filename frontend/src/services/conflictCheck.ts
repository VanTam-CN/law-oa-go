import { post, get } from '@/services/http';
import { message } from '@/utils/messageHelper';

// 冲突检索请求接口
export interface ConflictCheckRequest {
  clientId: string;
  clientName: string;
  clientType: 'PERSON' | 'COMPANY' | 'ANY';
  otherParties: string[];
  caseName: string;
  caseType: string;
  searchYears?: number;
  includeCorporateRelations?: boolean;
  searchDepth?: 'BASIC' | 'STANDARD' | 'DEEP';
  userId?: string; // 用户ID，后端需要
}

// 冲突检索响应接口
export interface ConflictCheckResponse {
  checkId: string;
  hasConflict: boolean;
  conflictCases: ConflictCase[];
  checkStatistics: CheckStatistics;
  riskAssessment: RiskAssessment;
  recommendations: string[];
  checkTime: string;
  duration: number;
}

export interface ConflictCase {
  caseId: string;
  caseName: string;
  caseNo: string;
  conflictType: string;
  riskLevel: string;
  description: string;
  caseStatus: string;
  createTime: string;
  conflictDetails: string;
}

export interface CheckStatistics {
  totalCasesChecked: number;
  clientHistoryCases: number;
  relatedPartiesChecked: number;
  corporateRelationsChecked: number;
  timeRange: string;
  searchScope: string;
}

export interface RiskAssessment {
  overallRisk: string;
  riskReason: string;
  requiresApproval: boolean;
  approvalLevel?: string;
  riskFactors: string[];
}

// 冲突检索服务
export class ConflictCheckService {
  /**
   * 执行完整的利益冲突检索
   */
  static async performConflictCheck(request: ConflictCheckRequest): Promise<ConflictCheckResponse> {
    try {
      // 验证必要参数
      this.validateRequest(request);
      
      // 开发环境下的备用方案：如果后端不可用，使用模拟数据
      const isDevelopment = typeof window !== 'undefined' && window.location.hostname === 'localhost';
      
      try {
        // 调用后端API
        const response = await post<ConflictCheckResponse>('/conflict/check', request);
        
        // 验证响应基本结构
        if (!response) {
          throw new Error('后端服务无响应');
        }
        
        // 🔧 修复：HTTP拦截器已经处理了响应格式，直接使用response作为结果
        // HTTP拦截器会自动提取data字段，所以response就是我们需要的数据
        const result = response as ConflictCheckResponse;

        if (typeof result.hasConflict === 'undefined') {
          console.warn('后端响应缺少hasConflict字段，基于conflictCases判断');
          result.hasConflict = (result.conflictCases && result.conflictCases.length > 0);
        }

        return result;
      } catch (apiError) {
        console.error('API调用错误:', apiError);
        
        // 如果是开发环境且后端不可用，使用模拟数据
        if (isDevelopment && this.isConnectionError(apiError)) {
          console.warn('开发环境：后端API不可用，使用模拟数据');
          return this.getMockResponse(request);
        }
        
        // 重新抛出API错误
        throw apiError;
      }
      
    } catch (error) {
      console.error('冲突检索服务错误:', error);
      
      // 确保错误信息清晰
      if (error instanceof Error) {
        throw new Error(`冲突检索失败: ${error.message}`);
      } else {
        throw new Error('冲突检索服务异常');
      }
    }
  }
  
  /**
   * 快速预检
   */
  static async performPreCheck(clientName: string, otherParties: string[]): Promise<any> {
    try {
      const response = await post('/conflict/pre-check', { clientName, otherParties });
      return response;
    } catch (error) {
      console.error('预检服务错误:', error);
      throw error;
    }
  }
  
  /**
   * 获取检索历史
   */
  static async getCheckHistory(clientId?: string): Promise<any[]> {
    try {
      const endpoint = clientId 
        ? `/conflict/history?clientId=${clientId}`
        : '/conflict/history';
        
      const response = await get(endpoint);
      return response || [];
    } catch (error) {
      console.error('获取检索历史错误:', error);
      return [];
    }
  }
  
  /**
   * 检查是否为网络连接错误
   */
  private static isConnectionError(error: any): boolean {
    // 检查各种网络连接错误的情况
    if (error instanceof TypeError) {
      return error.message.includes('Failed to fetch') ||
             error.message.includes('Network request failed') ||
             error.message.includes('网络连接失败') ||
             error.message.includes('ERR_CONNECTION_REFUSED') ||
             error.message.includes('ERR_NETWORK');
    }
    
    // 检查Fetch API错误
    if (error instanceof Error) {
      return error.message.includes('fetch') ||
             error.message.includes('ECONNREFUSED') ||
             error.message.includes('net::ERR_') ||
             error.message.includes('请求超时') ||
             error.message.includes('连接被拒绝');
    }
    
    // 检查状态码相关错误
    if (typeof error === 'object' && error !== null) {
      // 检查是否有网络相关的属性
      return error.name === 'NetworkError' ||
             error.name === 'AbortError' ||
             error.code === 'NETWORK_ERROR' ||
             error.type === 'network';
    }
    
    return false;
  }
  
  /**
   * 获取模拟响应数据（仅开发环境）- 诚实的数据展示
   */
  private static getMockResponse(request: ConflictCheckRequest): ConflictCheckResponse {
    // 开发环境诚实显示：没有真实数据
    const hasConflict = false; // 开发环境默认无冲突，避免误导
    
    return {
      checkId: 'MOCK_' + Date.now(),
      hasConflict,
      conflictCases: [], // 开发环境显示空数组
      checkStatistics: {
        totalCasesChecked: 0, // 诚实显示0，没有真实数据
        clientHistoryCases: 0, // 诚实显示0
        relatedPartiesChecked: request.otherParties.length + 1,
        corporateRelationsChecked: 0, // 诚实显示0
        timeRange: '开发环境模拟数据',
        searchScope: '模拟数据库'
      },
      riskAssessment: {
        overallRisk: 'LOW',
        riskReason: '开发环境：未连接真实数据库',
        requiresApproval: false,
        riskFactors: ['开发环境模拟']
      },
      recommendations: [
        '开发环境：此为模拟数据，非真实冲突检测结果',
        '请连接真实数据库进行实际冲突检测',
        '开发模式下不提供真实的冲突检测服务'
      ],
      checkTime: new Date().toISOString(),
      duration: 100 // 固定响应时间，避免造假
    };
  }
  
  /**
   * 验证请求参数
   */
  private static validateRequest(request: ConflictCheckRequest): void {
    if (!request.clientId || !request.clientName) {
      throw new Error('委托人信息不能为空');
    }
    
    if (!request.caseName || !request.caseType) {
      throw new Error('案件名称和类型不能为空');
    }
    
    if (!['PERSON', 'COMPANY'].includes(request.clientType)) {
      throw new Error('委托人类型无效');
    }
  }
  
  /**
   * 格式化冲突类型标签
   */
  static formatConflictType(type: string): string {
    const labels: { [key: string]: string } = {
      'SAME_PARTY': '同一当事人',
      'RELATED_PARTY': '关联方',
      'OPPOSING_PARTY': '对方当事人'
    };
    return labels[type] || type;
  }
  
  /**
   * 格式化风险等级
   */
  static formatRiskLevel(level: string): { text: string, color: string } {
    const formats: { [key: string]: { text: string, color: string } } = {
      'LOW': { text: '低风险', color: 'green' },
      'MEDIUM': { text: '中等风险', color: 'orange' },
      'HIGH': { text: '高风险', color: 'red' },
      'CRITICAL': { text: '严重风险', color: 'red' }
    };
    return formats[level] || { text: level, color: 'gray' };
  }
}

// 冲突检索结果处理工具
export class ConflictCheckResultProcessor {
  /**
   * 转换后端数据为前端展示格式
   */
  static processResult(backendResult: any): any {
    return {
      hasConflict: backendResult.hasConflict,
      conflictCases: (backendResult.conflictCases || []).map((conflict: any) => ({
        id: conflict.caseId,
        caseName: conflict.caseName,
        caseNo: conflict.caseNo,
        conflictType: ConflictCheckService.formatConflictType(conflict.conflictType),
        riskLevel: conflict.riskLevel,
        description: conflict.description,
        caseStatus: conflict.caseStatus,
        createTime: conflict.createTime
      })),
      checkTime: backendResult.checkTime,
      checkDetails: {
        totalCasesChecked: backendResult.checkStatistics?.totalCasesChecked || 0,
        clientHistoryCases: backendResult.checkStatistics?.clientHistoryCases || 0,
        relatedPartiesChecked: backendResult.checkStatistics?.relatedPartiesChecked || 0,
        corporateRelationsChecked: backendResult.checkStatistics?.corporateRelationsChecked || 0,
        timeRange: backendResult.checkStatistics?.timeRange || '未知',
        riskAssessment: backendResult.riskAssessment?.overallRisk || 'UNKNOWN'
      },
      recommendations: backendResult.recommendations || []
    };
  }
  
  /**
   * 生成检索统计摘要
   */
  static generateSummary(checkDetails: any): string {
    const { totalCasesChecked, clientHistoryCases, relatedPartiesChecked } = checkDetails;
    return `检索了${totalCasesChecked}个案件，发现${clientHistoryCases}个历史案件，检查了${relatedPartiesChecked}个当事人`;
  }
}