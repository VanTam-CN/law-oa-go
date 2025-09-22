import { get, post, put, del } from './http';

// 案件相关接口
export interface CaseInfo {
  caseId?: number;
  caseNo: string;
  caseName: string;
  caseType: string;
  clientId: number | null;
  lawyerId: number | null;
  status: string;
  description?: string;
  projectCode?: string;
  contractAmount?: number;
  startDate?: string;
  endDate?: string;
  teamMembers?: string;
  projectType?: string;
  principalInfo?: string;
  opponentInfo?: string;
  causeOfAction?: string;
  assistingLawyerId?: number | null;
  billingMethod?: string;
  conflictCheckStatus?: string;
  isMajorRisk?: boolean;
  isMassCase?: boolean;
  isSensitiveCase?: boolean;
  contractDocument?: string;
  legalLetterDocument?: string;
  otherDocuments?: string;
}

// 客户相关接口
export interface ClientInfo {
  clientId?: number;
  clientName: string;
  phone: string;
  email: string;
  clientType: string;
  company?: string;
  idCard?: string;
  address: string;
}

// 律师相关接口
export interface LawyerInfo {
  lawyerId?: number;
  lawyerName: string;
  phone: string;
  email: string;
  licenseNo: string;
  position: string;
  department?: string;
  specialty?: string;
}

// 文档相关接口
export interface DocumentInfo {
  docId?: number;
  docName: string;
  docType: string;
  caseId?: number;
  clientId?: number;
  filePath: string;
  fileSize: number;
  fileType: string;
  description?: string;
}

/**
 * 案件管理API
 */
export const caseAPI = {
  // 获取案件列表
  getList: (params?: any): Promise<{ total: number; rows: CaseInfo[] }> => {
    return get('/lawfirm/case/list', params);
  },

  // 获取单个案件
  getById: (id: number): Promise<CaseInfo> => {
    return get(`/lawfirm/case/${id}`);
  },

  // 创建案件
  create: (data: CaseInfo): Promise<any> => {
    return post('/lawfirm/case', data);
  },

  // 更新案件
  update: (data: CaseInfo): Promise<any> => {
    return put('/lawfirm/case', data);
  },

  // 删除案件
  delete: (id: number): Promise<any> => {
    return del(`/lawfirm/case/${id}`);
  }
};

/**
 * 客户管理API
 */
export const clientAPI = {
  // 获取客户列表
  getList: (params?: any): Promise<{ total: number; rows: ClientInfo[] }> => {
    return get('/lawfirm/client/list', params);
  },

  // 获取单个客户
  getById: (id: number): Promise<ClientInfo> => {
    return get(`/lawfirm/client/${id}`);
  },

  // 创建客户
  create: (data: ClientInfo): Promise<any> => {
    return post('/lawfirm/client', data);
  },

  // 更新客户
  update: (data: ClientInfo): Promise<any> => {
    return put('/lawfirm/client', data);
  },

  // 删除客户
  delete: (id: number): Promise<any> => {
    return del(`/lawfirm/client/${id}`);
  }
};

/**
 * 律师管理API
 */
export const lawyerAPI = {
  // 获取律师列表
  getList: (params?: any): Promise<{ total: number; rows: LawyerInfo[] }> => {
    return get('/lawfirm/lawyers', params);
  },

  // 获取单个律师
  getById: (id: number): Promise<LawyerInfo> => {
    return get(`/lawfirm/lawyers/${id}`);
  },

  // 创建律师
  create: (data: LawyerInfo): Promise<any> => {
    return post('/lawfirm/lawyers', data);
  },

  // 更新律师
  update: (data: LawyerInfo): Promise<any> => {
    return put(`/lawfirm/lawyers/${data.lawyerId}`, data);
  },

  // 删除律师
  delete: (id: number): Promise<any> => {
    return del(`/lawfirm/lawyers/${id}`);
  }
};

/**
 * 文档管理API
 */
export const documentAPI = {
  // 获取文档列表
  getList: (params?: any): Promise<{ total: number; rows: DocumentInfo[] }> => {
    return get('/lawfirm/document/list', params);
  },

  // 获取单个文档
  getById: (id: number): Promise<DocumentInfo> => {
    return get(`/lawfirm/document/${id}`);
  },

  // 上传文档
  upload: (formData: FormData): Promise<DocumentInfo> => {
    return post('/lawfirm/document/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    });
  },

  // 下载文档
  download: (id: number): Promise<Blob> => {
    return get(`/lawfirm/document/download/${id}`, null, {
      responseType: 'blob'
    });
  },

  // 删除文档
  delete: (id: number): Promise<any> => {
    return del(`/lawfirm/document/${id}`);
  }
};

/**
 * 利益冲突检查API
 */
export interface ConflictCheckRequest {
  clientId?: number;
  opponentInfo?: string;
  lawyerId?: number;
  caseType?: string;
  causeOfAction?: string;
}

export interface ConflictCheckResult {
  status: 'checking' | 'passed' | 'warning' | 'failed';
  score: number;
  conflicts: Array<{
    id: string;
    type: 'client' | 'opponent' | 'lawyer' | 'case';
    level: 'low' | 'medium' | 'high';
    description: string;
    relatedCase?: string;
    recommendation: string;
    details?: string;
    foundTime?: string;
    impact?: string;
    probability?: number;
    severity?: number;
    evidence?: Array<{
      type: string;
      description: string;
      date?: string;
      caseNumber?: string;
    }>;
  }>;
  summary: string;
  checkTime?: string;
  checker?: string;
  totalChecked?: number;
  riskFactors?: Array<{
    factor: string;
    weight: number;
    score: number;
    description: string;
  }>;
  recommendations?: Array<{
    priority: 'high' | 'medium' | 'low';
    action: string;
    description: string;
    timeline?: string;
  }>;
  relatedCases?: Array<{
    caseId: string;
    caseName: string;
    status: string;
    relationship: string;
    riskLevel: string;
  }>;
  complianceNotes?: string;
}

export const conflictAPI = {
  // 执行利益冲突检查
  check: (data: ConflictCheckRequest): Promise<ConflictCheckResult> => {
    return post('/conflict-check/analyze', data);
  },

  // 获取冲突检查历史
  getHistory: (caseId: number): Promise<ConflictCheckResult[]> => {
    return get(`/conflict-check/history/${caseId}`);
  },

  // 更新冲突检查状态
  updateStatus: (caseId: number, status: string, notes?: string): Promise<any> => {
    return put(`/conflict-check/status/${caseId}`, { status, notes });
  }
};