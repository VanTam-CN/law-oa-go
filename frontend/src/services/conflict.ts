import { post } from './http';
import {
  ConflictCheckRequest,
  ConflictCheckResponse,
  ConflictCheckFormData,
  ValidationError,
  SearchDepth
} from '@/types/conflict';
import {
  transformToConflictCheckRequest,
  debugConflictCheckRequest
} from '@/utils/conflictTransform';
import {
  handleError,
  shouldRetry,
  getUserFriendlyMessage,
  StandardError,
  ErrorHandler
} from '@/utils/errorHandler';

export interface ConflictCheckParams {
  project_name: string;
  client_name: string;
  opposite_parties: string;
  project_type: string;
  team_members: string[];
  description?: string;
}

export interface ConflictResult {
  has_conflict: boolean;
  conflict_level: 'none' | 'low' | 'medium' | 'high';
  conflicts: {
    id: number;
    type: string;
    entity: string;
    project: string;
    level: 'low' | 'medium' | 'high';
    description: string;
  }[];
}

// 创建专用的冲突检查错误处理器
const conflictErrorHandler = new ErrorHandler({
  enableLogging: true,
  enableRetry: true,
  maxRetries: 2,
  retryDelay: 1000,
  logLevel: 'ERROR'
});

/**
 * 转换为增强冲突检测API格式
 */
const transformToEnhancedConflictCheckRequest = (request: any) => {
  // 解析对方当事人 - 取第一个作为主要对方
  const opposingParty = request.otherParties?.[0] || '';

  // 转换搜索深度 - 使用后端期望的大写格式
  const searchDepth = request.searchDepth === 'DEEP' ? 'DEEP' : 'BASIC';

  // 获取律师ID - 从用户信息获取
  const lawyerId = parseInt(request.userId) || 1;

  return {
    clientId: request.clientId || request.lawyerId?.toString() || lawyerId.toString(),
    clientName: request.clientName,
    caseName: request.caseName || `${request.clientName}诉${opposingParty}案件`,
    caseType: request.caseType.toLowerCase(), // 确保是小写
    clientType: request.clientType === 'COMPANY' ? 'COMPANY' : 'PERSON',
    otherParties: [opposingParty],
    searchYears: request.searchYears || 5,
    includeCorporateRelations: request.includeCorporateRelations || false,
    searchDepth: searchDepth,
    userId: request.userId?.toString() || lawyerId.toString(),
    requestTime: new Date().toISOString()
  };
};

/**
 * 检测客户行业
 */
const detectClientIndustry = (clientName: string): string => {
  const industryMappings: { [key: string]: string } = {
    '阿里巴巴': '科技、媒体和通信',
    '阿里': '科技、媒体和通信',
    '淘宝': '科技、媒体和通信',
    '天猫': '科技、媒体和通信',
    '支付宝': '科技、媒体和通信',
    '腾讯': '科技、媒体和通信',
    '微信': '科技、媒体和通信',
    '字节跳动': '科技、媒体和通信',
    '抖音': '科技、媒体和通信',
    'TikTok': '科技、媒体和通信',
    '百度': '科技、媒体和通信',
    '京东': '科技、媒体和通信',
    '美团': '科技、媒体和通信',
    '银行': '金融',
    '保险': '金融',
    '证券': '金融',
    '基金': '金融',
    '医院': '医疗健康',
    '诊所': '医疗健康',
    '药': '医疗健康',
    '学校': '教育',
    '大学': '教育',
    '电力': '能源',
    '石油': '能源',
    '建筑': '房地产',
    '房地产': '房地产'
  };

  for (const [keyword, industry] of Object.entries(industryMappings)) {
    if (clientName.includes(keyword)) {
      return industry;
    }
  }

  return '其他';
};

/**
 * 转换增强API响应为前端格式
 */
const transformEnhancedApiResponse = (apiResponse: any): ConflictCheckResponse => {
  const data = apiResponse.data || {};

  return {
    success: apiResponse.success || false,
    message: apiResponse.message || '冲突检测完成',
    error: apiResponse.error || undefined,
    data: {
      hasConflict: data.hasConflict || false,
      riskAssessment: {
        overallRisk: data.riskAssessment?.overallRisk || 'NONE',
        riskScore: data.riskAssessment?.riskScore || 0,
        riskFactors: {
          conflictType: 'multiple',
          casePriority: 'medium',
          timeProximity: 30,
          industryOverlap: 0.5,
          financialImpact: 'medium'
        }
      },
      conflictCases: (data.conflictCases || []).map((item: any) => ({
        caseId: item.caseId || '-1',
        clientName: data.checkStatistics?.clientHistoryCases > 0 ? '历史客户' : '未知客户',
        caseName: item.caseName || '未知案件',
        conflictType: item.conflictType || '代理冲突',
        riskLevel: item.riskLevel || 'MEDIUM',
        riskScore: data.riskAssessment?.riskScore || 50,
        description: item.description || '检测到利益冲突'
      })),
      competitionAnalysis: data.competitionAnalysis,
      recommendations: data.recommendations || [],
      analysisSummary: {
        totalCasesChecked: data.checkStatistics?.totalCasesChecked || 0,
        directConflicts: data.conflictCases?.length || 0,
        industryConflicts: 0,
        nameSimilarityCases: 0,
        relatedConflicts: 0
      },
      detectionTime: data.checkTime || new Date().toISOString(),
      requestId: apiResponse.requestId || `REQ_${Date.now()}`
    },
    requestId: apiResponse.requestId || `REQ_${Date.now()}`,
    timestamp: apiResponse.timestamp || new Date().toISOString()
  };
};

/**
 * 执行利益冲突检查 (新版本 - 使用增强API)
 * @param formData 表单数据
 * @param clientInfo 客户信息 (可选)
 * @param userInfo 用户信息 (可选)
 * @param attempt 当前重试次数 (内部使用)
 * @returns 冲突检查结果
 */
export const performConflictCheck = async (
  formData: ConflictCheckFormData,
  clientInfo?: any,
  userInfo?: any,
  attempt: number = 1
): Promise<ConflictCheckResponse> => {
  try {
    // 1. 验证和转换参数
    const { request, errors } = transformToConflictCheckRequest(formData, clientInfo, userInfo);

    if (errors.length > 0) {
      const errorMessage = errors.map(e => `${e.field}: ${e.message}`).join('; ');
      console.error('🚫 冲突检查请求验证失败:', errorMessage);

      return {
        success: false,
        message: '请求参数验证失败',
        error: errorMessage,
        data: undefined,
        requestId: `VALIDATION_${Date.now()}`,
        timestamp: new Date().toISOString()
      } as ConflictCheckResponse;
    }

    debugConflictCheckRequest(request);

    // 2. 转换为增强API格式
    const enhancedRequest = transformToEnhancedConflictCheckRequest(request);

    console.group('🚀 发送到增强API的请求');
    console.log('转换后请求体:', JSON.stringify(enhancedRequest, null, 2));
    console.groupEnd();

    // 3. 发送增强API请求
    const response = await post<any>('/conflict/check', enhancedRequest);

    // 4. 转换响应格式
    const transformedResponse = transformEnhancedApiResponse(response);

    console.log('✅ 增强冲突检查API调用成功:', transformedResponse.requestId);
    return transformedResponse;

  } catch (error: any) {
    console.error(`💥 增强冲突检查API调用失败 (尝试 ${attempt}):`, error);

    // 如果增强API失败，回退到旧API
    if (attempt === 1) {
      console.log('🔄 回退到旧版API...');
      try {
        const legacyResponse = await performConflictCheckLegacy({
          project_name: formData.caseName,
          client_name: formData.clientName || clientInfo?.name,
          opposite_parties: formData.opponentInfo,
          project_type: formData.caseType,
          team_members: [],
          description: formData.description
        });

        return {
          success: true,
          message: '冲突检测完成（使用备用API）',
          error: undefined,
          data: {
            hasConflict: legacyResponse.has_conflict,
            riskAssessment: {
              overallRisk: legacyResponse.conflict_level.toUpperCase(),
              riskScore: legacyResponse.conflict_level === 'high' ? 80 :
                        legacyResponse.conflict_level === 'medium' ? 50 : 20,
              riskFactors: {
                conflictType: 'detected',
                casePriority: 'medium',
                timeProximity: 30,
                industryOverlap: 0.3,
                financialImpact: 'medium'
              }
            },
            conflictCases: legacyResponse.conflicts.map(c => ({
              caseId: c.id.toString(),
              clientName: c.entity,
              caseName: c.project,
              conflictType: c.type,
              riskLevel: c.level.toUpperCase(),
              riskScore: c.level === 'high' ? 80 : c.level === 'medium' ? 50 : 20,
              description: c.description
            })),
            recommendations: [],
            analysisSummary: {
              totalCasesChecked: legacyResponse.conflicts.length,
              directConflicts: legacyResponse.conflicts.length,
              industryConflicts: 0,
              nameSimilarityCases: 0,
              relatedConflicts: 0
            },
            detectionTime: new Date().toISOString(),
            requestId: `LEGACY_${Date.now()}`
          },
          requestId: `LEGACY_${Date.now()}`,
          timestamp: new Date().toISOString()
        };
      } catch (legacyError) {
        console.error('备用API也失败了:', legacyError);
      }
    }

    // 标准化错误处理
    const standardError = conflictErrorHandler.handleError(error, 'ConflictCheck');

    // 检查是否应该重试
    if (shouldRetry(standardError, attempt)) {
      console.log(`🔄 准备重试 (${attempt}/${conflictErrorHandler['config'].maxRetries})...`);

      const delay = conflictErrorHandler.getRetryDelay(attempt);
      await new Promise(resolve => setTimeout(resolve, delay));

      return performConflictCheck(formData, clientInfo, userInfo, attempt + 1);
    }

    // 返回用户友好的错误响应
    return {
      success: false,
      message: getUserFriendlyMessage(standardError),
      error: standardError.message,
      data: undefined,
      requestId: standardError.requestId,
      timestamp: standardError.timestamp
    } as ConflictCheckResponse;
  }
};

/**
 * @deprecated 请使用新的 performConflictCheck 方法，此方法为向后兼容保留
 * 执行利益冲突检查 (旧版本)
 * @param params 旧的参数格式
 * @returns 冲突检查结果
 */
export const performConflictCheckLegacy = (params: ConflictCheckParams): Promise<ConflictResult> => {
  console.warn('⚠️  使用了已弃用的performConflictCheckLegacy方法，建议使用新的performConflictCheck方法');

  // 转换为新的表单数据格式
  const formData: ConflictCheckFormData = {
    caseName: params.project_name,
    clientName: params.client_name,
    caseType: params.project_type,
    opponentInfo: params.opposite_parties,
    // 其他字段使用默认值
    searchYears: 5,
    searchDepth: SearchDepth.DEEP,
    includeCorporateRelations: true
  };

  // 调用新方法
  return performConflictCheck(formData).then(response => {
    // 转换为旧格式
    return {
      has_conflict: response.data?.hasConflict || false,
      conflict_level: response.data?.riskAssessment?.overallRisk === 'HIGH' ? 'high' :
                    response.data?.riskAssessment?.overallRisk === 'MEDIUM' ? 'medium' :
                    response.data?.riskAssessment?.overallRisk === 'LOW' ? 'low' : 'none',
      conflicts: response.data?.conflictCases?.map(c => ({
        id: parseInt(c.caseId) || 0,
        type: c.conflictType,
        entity: c.clientName,
        project: c.caseName,
        level: c.riskLevel.toLowerCase() as 'low' | 'medium' | 'high',
        description: c.description
      })) || []
    };
  }).catch(error => {
    console.error('Legacy conflict check failed:', error);
    return {
      has_conflict: false,
      conflict_level: 'none',
      conflicts: []
    };
  });
};

/**
 * 执行利益冲突预检
 * @param params 预检参数
 * @returns 预检结果
 */
export const performPreScreen = (params: {
  our_client_ids: number[];
  opponent_parties: string[];
  third_parties?: string[];
}): Promise<any> => {
  return post<any>('/conflict-check/pre-screen', params);
};