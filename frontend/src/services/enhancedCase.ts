import { get, post, put, del } from './http';

// 增强案例接口定义
export interface EnhancedCase {
  id?: number;
  title: string;
  description: string;
  caseType: string;
  priority: string;
  status: string;
  startDate?: string;
  endDate?: string;
  createdAt: string;
  updatedAt: string;

  // 客户信息
  clientProfiles: EnhancedClientInfo[];
  clientProfileIds: string[];

  // 团队信息
  teamAssignment: TeamAssignmentInfo;

  // 冲突检测信息
  conflictDetection: ConflictDetectionInfo;
  riskLevel: string;

  // 豁免信息
  waiverInfo: WaiverInfo;

  // 信息屏障信息
  ethicalScreens: EthicalScreenInfo[];
}

export interface EnhancedClientInfo {
  clientProfileId: string;
  clientNumber: string;
  clientType: string;
  clientCategory: string;
  role: string;
}

export interface TeamAssignmentInfo {
  members: TeamMemberInfo[];
  leadLawyer: TeamMemberInfo;
  createdAt: string;
}

export interface TeamMemberInfo {
  user: {
    id: number;
    name: string;
    email: string;
    role: string;
  };
  role: string;
  capacity: number;
}

export interface ConflictDetectionInfo {
  status: string;
  lastChecked: string;
  conflicts: any[];
  riskLevel: string;
}

export interface WaiverInfo {
  status: string;
  type: string;
  requiredApprovals: string[];
  monitoringPlan: string;
  expiryDate: string;
  conditions: string[];
}

export interface EthicalScreenInfo {
  id: string;
  type: string;
  description: string;
  status: string;
  createdAt: string;
}

// 增强案例创建请求
export interface EnhancedCreateCaseRequest {
  // 基础字段
  title: string;
  description: string;
  caseType: string;
  priority: string;
  startDate?: string;
  practiceArea: string;
  estimatedDuration?: string;
  billingMethod: string;

  // 客户信息 - 支持多客户
  clientProfileIds: string[];
  clientRoles: Record<string, ClientRoleInfo>;

  // 团队分配
  lawyerId: number;
  assistingLawyerId?: number;
  teamMembers?: EnhancedTeamMemberRequest[];

  // 冲突检测配置
  conflictCheckConfig: ConflictCheckConfig;

  // 分配信息
  assignedBy: number;
  isMajorRisk: boolean;
}

export interface ClientRoleInfo {
  role: string;
  relationshipDescription?: string;
  contactInfo?: string;
}

export interface EnhancedTeamMemberRequest {
  userId: number;
  role: string;
  capacity?: number;
}

export interface ConflictCheckConfig {
  enabled: boolean;
  checkOnCreate: boolean;
  searchYears?: number;
  includeCorporateRelations?: boolean;
  searchDepth?: 'STANDARD' | 'DEEP' | 'COMPREHENSIVE';
  autoWaiverIfPossible?: boolean;
}

// 增强案例更新请求
export interface UpdateEnhancedCaseRequest {
  title?: string;
  description?: string;
  caseType?: string;
  priority?: string;
  status?: string;
  startDate?: string;
  endDate?: string;
}

// 增强案例列表请求
export interface ListEnhancedCasesRequest {
  page: number;
  pageSize: number;
  search?: string;
  status?: string;
  priority?: string;
  caseType?: string;
  clientId?: string;
  lawyerId?: string;
}

// 增强案例列表响应
export interface ListEnhancedCasesResponse {
  cases: EnhancedCase[];
  pagination: {
    page: number;
    pageSize: number;
    total: number;
    totalPages: number;
  };
}

// 冲突检测结果
export interface ConflictCheckResult {
  caseId: number;
  checkId: string;
  status: string;
  conflicts: any[];
  riskLevel: string;
  waiverRequired: boolean;
  checkedAt: string;
  nextCheckDate?: string;
}

// 添加客户到案件请求
export interface AddClientToCaseRequest {
  clientProfileId: string;
  role: string;
  conflictCheckConfig?: ConflictCheckConfig;
}

// 增强案例服务
export const enhancedCaseService = {
  // 创建增强案例
  createEnhancedCase: (data: EnhancedCreateCaseRequest) => {
    console.log('创建增强案例数据:', data);
    return post<EnhancedCase>('/enhanced-cases', data);
  },

  // 获取增强案例详情
  getEnhancedCase: (id: number) => get<EnhancedCase>(`/enhanced-cases/${id}`),

  // 更新增强案例
  updateEnhancedCase: (id: number, data: UpdateEnhancedCaseRequest) => {
    console.log('更新增强案例数据:', { id, data });
    return put<EnhancedCase>(`/enhanced-cases/${id}`, data);
  },

  // 获取增强案例列表
  listEnhancedCases: (params: ListEnhancedCasesRequest) => {
    console.log('查询增强案例列表参数:', params);
    return get<ListEnhancedCasesResponse>('/enhanced-cases', params);
  },

  // 删除增强案例
  deleteEnhancedCase: (id: number) => del(`/enhanced-cases/${id}`),

  // 执行冲突检测
  performConflictCheck: (id: number) => {
    console.log('为案例执行冲突检测:', id);
    return post<ConflictCheckResult>(`/enhanced-cases/${id}/conflict-check`);
  },

  // 添加客户到案件
  addClientToCase: (id: number, data: AddClientToCaseRequest) => {
    console.log('添加客户到案件:', { caseId: id, data });
    return post<EnhancedCase>(`/enhanced-cases/${id}/clients`, data);
  },

  // 从案件移除客户
  removeClientFromCase: (id: number, clientId: string) => {
    console.log('从案件移除客户:', { caseId: id, clientId });
    return del(`/enhanced-cases/${id}/clients/${clientId}`);
  },

  // 获取冲突检测状态
  getConflictDetectionStatus: (id: number) => {
    return get<any>(`/enhanced-cases/${id}/conflict-status`);
  },

  // 触发冲突检测
  triggerConflictDetection: (id: number) => {
    return post<any>(`/enhanced-cases/${id}/conflict-check`);
  }
};

// 客户角色相关服务
export const clientRoleService = {
  // 获取可用的客户角色类型
  getClientRoles: () => [
    { value: 'PRIMARY', label: '主要委托人', description: '案件的主要委托人' },
    { value: 'SECONDARY', label: '次要委托人', description: '案件的次要委托人' },
    { value: 'GUARANTOR', label: '担保人', description: '为案件提供担保的个人或机构' },
    { value: 'INTERESTED_PARTY', label: '利益相关方', description: '与案件有利益关系的第三方' },
    { value: 'REPRESENTATIVE', label: '代表人', description: '代表其他主体参与案件的当事人' },
    { value: 'JOINT_APPLICANT', label: '共同申请人', description: '共同提出申请的多方当事人' }
  ],

  // 验证客户角色配置
  validateClientRoles: (clientIds: string[], clientRoles: Record<string, ClientRoleInfo>) => {
    const errors: string[] = [];

    // 验证是否所有客户都有角色配置
    clientIds.forEach(clientId => {
      if (!clientRoles[clientId]) {
        errors.push(`客户 ${clientId} 缺少角色配置`);
      }
    });

    // 验证是否至少有一个主要委托人
    const hasPrimaryClient = Object.values(clientRoles).some(role => role.role === 'PRIMARY');
    if (!hasPrimaryClient) {
      errors.push('案件必须有至少一个主要委托人');
    }

    // 验证主要委托人数量
    const primaryClients = Object.values(clientRoles).filter(role => role.role === 'PRIMARY');
    if (primaryClients.length > 1) {
      errors.push('主要委托人不能超过一个');
    }

    return {
      isValid: errors.length === 0,
      errors
    };
  }
};

// 冲突检测相关工具函数
export const conflictCheckUtils = {
  // 生成冲突检查配置
  generateConflictCheckConfig: (clientType: string, isMajorRisk: boolean): ConflictCheckConfig => {
    return {
      enabled: true,
      checkOnCreate: true,
      searchYears: clientType === 'COMPANY' ? 7 : 5,
      includeCorporateRelations: clientType === 'COMPANY',
      searchDepth: isMajorRisk ? 'COMPREHENSIVE' : 'STANDARD',
      autoWaiverIfPossible: !isMajorRisk
    };
  },

  // 验证冲突检查配置
  validateConflictCheckConfig: (config: ConflictCheckConfig) => {
    const errors: string[] = [];

    if (config.searchYears && (config.searchYears < 1 || config.searchYears > 20)) {
      errors.push('搜索年限必须在1-20年之间');
    }

    if (!['STANDARD', 'DEEP', 'COMPREHENSIVE'].includes(config.searchDepth || '')) {
      errors.push('搜索深度必须是STANDARD、DEEP或COMPREHENSIVE');
    }

    return {
      isValid: errors.length === 0,
      errors
    };
  }
};