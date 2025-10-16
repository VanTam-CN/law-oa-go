/**
 * 利益冲突检查参数转换工具
 * 用于将前端表单数据转换为后端API期望的格式
 */

import {
  ConflictCheckRequest,
  ConflictCheckFormData,
  CaseType,
  ClientType,
  SearchDepth,
  ValidationError
} from '@/types/conflict';

// 案件类型映射表 (前端 -> 后端)
const CASE_TYPE_MAPPING: Record<string, CaseType> = {
  'CIVIL': CaseType.CIVIL,
  'COMMERCIAL': CaseType.COMMERCIAL,
  'CRIMINAL': CaseType.CRIMINAL,
  'ADMINISTRATIVE': CaseType.ADMINISTRATIVE,
  'ARBITRATION': CaseType.ARBITRATION,
  'CONSULTATION': CaseType.CONSULTATION,
  'OTHER': CaseType.OTHER,
  '民事': CaseType.CIVIL,
  '商事': CaseType.COMMERCIAL,
  '刑事': CaseType.CRIMINAL,
  '行政': CaseType.ADMINISTRATIVE,
  '仲裁': CaseType.ARBITRATION,
  '咨询': CaseType.CONSULTATION,
  '其他': CaseType.OTHER
};

// 客户类型检测
export const detectClientType = (clientName: string, clientInfo?: any): ClientType => {
  // 如果有明确的客户类型信息，直接使用
  if (clientInfo?.type) {
    const type = clientInfo.type.toUpperCase();
    if (type === 'PERSON' || type === 'COMPANY') {
      return type as ClientType;
    }
  }

  // 根据客户名称智能推断
  if (!clientName) {
    return ClientType.PERSON; // 默认为个人
  }

  // 企业名称常见关键词
  const companyKeywords = [
    '有限公司', '股份有限公司', '集团', '公司', '企业', '实业',
    '科技', '投资', '控股', '发展', '建设', '工程', '贸易',
    'Co., Ltd', 'Ltd.', 'Inc.', 'LLC', 'Corp.', 'Group'
  ];

  const hasCompanyKeyword = companyKeywords.some(keyword =>
    clientName.includes(keyword) ||
    clientName.toLowerCase().includes(keyword.toLowerCase())
  );

  return hasCompanyKeyword ? ClientType.COMPANY : ClientType.PERSON;
};

// 案件类型转换
export const transformCaseType = (caseType: string): CaseType => {
  if (!caseType) {
    return CaseType.OTHER; // 默认值
  }

  const normalizedType = caseType.trim().toUpperCase();
  const mappedType = CASE_TYPE_MAPPING[normalizedType];

  if (mappedType) {
    return mappedType;
  }

  // 如果映射失败，尝试直接匹配枚举值
  if (Object.values(CaseType).includes(normalizedType as CaseType)) {
    return normalizedType as CaseType;
  }

  // 最后的默认值
  console.warn(`未知的案件类型: ${caseType}, 使用默认值 OTHER`);
  return CaseType.OTHER;
};

// 处理对方当事人信息
export const parseOtherParties = (opponentInfo?: string): string[] => {
  if (!opponentInfo) {
    return [];
  }

  // 先按换行符分割
  let lines = opponentInfo.split(/[\n\r]+/);

  // 提取可能的当事人名称
  const partyNames: string[] = [];

  for (const line of lines) {
    const trimmedLine = line.trim();
    if (!trimmedLine) continue;

    // 如果是数字开头的行（如"1. "），提取后面的内容
    if (/^\d+[\.、]\s*/.test(trimmedLine)) {
      const cleanLine = trimmedLine.replace(/^\d+[\.、]\s*/, '');
      partyNames.push(cleanLine);
    }
    // 如果包含常见的企业标识后缀，很可能是公司名称
    else if (isLikelyPartyName(trimmedLine)) {
      partyNames.push(trimmedLine);
    }
    // 如果包含冒号，提取冒号前的内容作为标签，冒号后的作为名称
    else if (trimmedLine.includes('：') || trimmedLine.includes(':')) {
      const parts = trimmedLine.split(/[：:]/);
      if (parts.length >= 2) {
        const label = parts[0].trim();
        const value = parts[1].trim();

        // 如果标签表明这是公司名称或姓名
        if (label.includes('公司') || label.includes('企业') || label.includes('姓名') ||
            label.includes('名称') || label.includes('单位') || isLikelyPartyName(value)) {
          partyNames.push(value);
        }
      }
    }
    // 否则，如果这行看起来像是一个合理的当事人名称，也加入
    else if (isLikelyPartyName(trimmedLine)) {
      partyNames.push(trimmedLine);
    }
  }

  // 如果没有找到任何合理的名称，则使用原始分割方法
  if (partyNames.length === 0) {
    // 支持多种分隔符
    const separators = [',', ';', '、', '，', '；', '|'];
    let parties = [opponentInfo];

    // 尝试各种分隔符进行分割
    for (const separator of separators) {
      if (opponentInfo.includes(separator)) {
        parties = opponentInfo.split(separator);
        break;
      }
    }

    return parties
      .map(party => party.trim())
      .filter(party => party.length > 0)
      .filter((party, index, arr) => arr.indexOf(party) === index); // 去重
  }

  // 清理和过滤空值
  return partyNames
    .map(party => {
      // 移除常见的无关后缀
      return party
        .replace(/(\s*纠纷|案件|诉讼|仲裁|争议|冲突|诉讼)$/, '') // 移除案件类型后缀
        .replace(/(\s*公司|企业|集团|有限公司|股份有限公司)$/, '') // 移除企业后缀（如果需要可以保留）
        .trim();
    })
    .filter(party => party.length > 0)
    .filter((party, index, arr) => arr.indexOf(party) === index); // 去重
};

// 判断文本是否可能是当事人名称
function isLikelyPartyName(text: string): boolean {
  if (!text || text.length < 2) return false;

  // 包含常见的企业标识符
  const companySuffixes = ['有限公司', '股份有限公司', '集团', '公司', '企业', '实业',
                           '科技', '投资', '控股', '发展', '建设', '工程', '贸易',
                           'Co., Ltd', 'Ltd.', 'Inc.', 'LLC', 'Corp.', 'Group'];

  const hasCompanySuffix = companySuffixes.some(suffix =>
    text.includes(suffix) || text.toLowerCase().includes(suffix.toLowerCase())
  );

  // 如果文本太长，不太可能是纯名称
  if (text.length > 50) return false;

  // 如果包含数字编号，很可能是标签而不是名称
  if (/^\d+[\.、]/.test(text)) return false;

  // 如果包含常见的描述性词汇，不太可能是纯名称
  const descriptionWords = ['纠纷', '案件', '诉讼', '仲裁', '争议', '地址', '联系', '电话',
                           '邮箱', '传真', '邮编', '网址', '说明', '备注'];
  const hasDescriptionWord = descriptionWords.some(word => text.includes(word));

  return hasCompanySuffix || (!hasDescriptionWord && text.length <= 20);
}

// 验证请求数据
export const validateConflictCheckRequest = (
  formData: ConflictCheckFormData,
  clientInfo?: any
): ValidationError[] => {
  const errors: ValidationError[] = [];

  // 验证必填字段 - 如果有clientInfo则可以不验证formData中的clientName
  if ((!formData.clientName || formData.clientName.trim().length === 0) && !clientInfo?.name) {
    errors.push({
      field: 'clientName',
      message: '客户名称不能为空',
      code: 'REQUIRED_FIELD'
    });
  }

  if (!formData.caseName || formData.caseName.trim().length === 0) {
    errors.push({
      field: 'caseName',
      message: '案件名称不能为空',
      code: 'REQUIRED_FIELD'
    });
  }

  if (!formData.caseType) {
    errors.push({
      field: 'caseType',
      message: '案件类型不能为空',
      code: 'REQUIRED_FIELD'
    });
  }

  // 验证数据长度
  if (formData.caseName && formData.caseName.length > 200) {
    errors.push({
      field: 'caseName',
      message: '案件名称不能超过200个字符',
      code: 'MAX_LENGTH_EXCEEDED'
    });
  }

  if (formData.clientName && formData.clientName.length > 100) {
    errors.push({
      field: 'clientName',
      message: '客户名称不能超过100个字符',
      code: 'MAX_LENGTH_EXCEEDED'
    });
  }

  // 验证搜索年限
  if (formData.searchYears && (formData.searchYears < 1 || formData.searchYears > 20)) {
    errors.push({
      field: 'searchYears',
      message: '搜索年限必须在1-20年之间',
      code: 'INVALID_RANGE'
    });
  }

  return errors;
};

// 主要转换函数
export const transformToConflictCheckRequest = (
  formData: ConflictCheckFormData,
  clientInfo?: any,
  userInfo?: any
): { request: ConflictCheckRequest; errors: ValidationError[] } => {
  const errors = validateConflictCheckRequest(formData, clientInfo);

  if (errors.length > 0) {
    return {
      request: {} as ConflictCheckRequest,
      errors
    };
  }

  const caseType = transformCaseType(formData.caseType);
  const clientType = formData.clientType || detectClientType(formData.clientName || '', clientInfo);
  const otherParties = parseOtherParties(formData.opponentInfo);

  const request: ConflictCheckRequest = {
    clientId: formData.clientId || (clientInfo?.id?.toString() || '1'),
    clientName: formData.clientName || clientInfo?.name || '未知客户',
    caseName: formData.caseName || '未知案件',
    caseType: caseType,
    clientType: clientType,
    otherParties: otherParties,
    searchYears: formData.searchYears || 5,
    includeCorporateRelations: formData.includeCorporateRelations !== false,
    searchDepth: formData.searchDepth || SearchDepth.STANDARD,
    userId: userInfo?.id?.toString() || userInfo?.userId || '1',
    requestTime: new Date().toISOString()
  };

  return { request, errors: [] };
};

// 生成默认的冲突检查请求 (用于测试和调试)
export const createDefaultConflictCheckRequest = (
  overrides: Partial<ConflictCheckRequest> = {}
): ConflictCheckRequest => {
  const defaultRequest: ConflictCheckRequest = {
    clientId: '1',
    clientName: '测试客户',
    caseName: '测试案件',
    caseType: CaseType.CIVIL,
    clientType: ClientType.PERSON,
    otherParties: [],
    searchYears: 5,
    includeCorporateRelations: true,
    searchDepth: SearchDepth.STANDARD,
    userId: '1',
    requestTime: new Date().toISOString(),
    ...overrides
  };

  return defaultRequest;
};

// 调试工具函数
export const debugConflictCheckRequest = (request: ConflictCheckRequest): void => {
  if (process.env.NODE_ENV === 'development') {
    console.group('🔍 利益冲突检查请求详情');
    console.log('客户ID:', request.clientId);
    console.log('客户名称:', request.clientName);
    console.log('案件名称:', request.caseName);
    console.log('案件类型:', request.caseType);
    console.log('客户类型:', request.clientType);
    console.log('对方当事人:', request.otherParties);
    console.log('搜索年限:', request.searchYears);
    console.log('搜索深度:', request.searchDepth);
    console.log('包含企业关系:', request.includeCorporateRelations);
    console.log('用户ID:', request.userId);
    console.log('请求时间:', request.requestTime);
    console.groupEnd();
  }
};