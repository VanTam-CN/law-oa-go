/**
 * 案件管理测试配置
 */

import { TestUser, CaseData, Document } from '../../types/test-data-types';
import { TestUtils } from '../../utils/test-utils';

export interface CaseTestConfig {
  baseUrl: string;
  defaultTimeout: number;
  screenshotOnFailure: boolean;
  validUser: TestUser;
  testCases: CaseData[];
  testDocuments: Document[];
}

export const CASE_TEST_CONFIG: CaseTestConfig = {
  baseUrl: 'http://localhost:3000',
  defaultTimeout: 30000,
  screenshotOnFailure: true,

  // 有效测试用户
  validUser: {
    id: 'test-attorney-001',
    username: 'test.attorney',
    email: 'test.attorney@lawfirm.com',
    password: 'TestPassword123!',
    firstName: 'Test',
    lastName: 'Attorney',
    role: 'attorney',
    department: 'litigation',
    permissions: [
      'cases.read',
      'cases.write',
      'cases.delete',
      'documents.read',
      'documents.write',
      'clients.read'
    ]
  } as TestUser,

  // 测试案件数据
  testCases: [
    {
      id: 'case-001',
      title: '合同纠纷案',
      caseNumber: 'CL-2024-001',
      clientName: '张三科技有限公司',
      clientType: 'corporate',
      caseType: 'contract',
      priority: 'high',
      status: 'active',
      description: '涉及服务合同的违约纠纷，要求赔偿损失',
      expectedOutcome: 'favorable',
      estimatedValue: 500000,
      startDate: '2024-01-15',
      expectedEndDate: '2024-06-30',
      assignedAttorney: 'test.attorney',
      teamMembers: ['paralegal1', 'assistant1'],
      tags: ['合同', '违约', '赔偿'],
      documents: ['contract.pdf', 'evidence.pdf'],
      budget: 80000,
      actualCost: 35000,
      notes: '案件进展顺利，已收集关键证据',
      court: '北京市朝阳区人民法院',
      judge: '李法官',
      opposingCounsel: '对方律师事务所',
      keyDates: [
        { date: '2024-01-15', event: '立案', description: '正式立案' },
        { date: '2024-02-20', event: '证据交换', description: '完成证据交换' },
        { date: '2024-03-15', event: '庭前会议', description: '召开庭前会议' }
      ],
      milestones: [
        { name: '立案', date: '2024-01-15', completed: true },
        { name: '证据收集', date: '2024-02-28', completed: true },
        { name: '开庭审理', date: '2024-04-15', completed: false },
        { name: '判决', date: '2024-06-30', completed: false }
      ]
    },
    {
      id: 'case-002',
      title: '劳动争议案',
      caseNumber: 'LD-2024-002',
      clientName: '李四',
      clientType: 'individual',
      caseType: 'employment',
      priority: 'medium',
      status: 'pending',
      description: '员工与公司之间的劳动合同解除争议',
      expectedOutcome: 'settlement',
      estimatedValue: 100000,
      startDate: '2024-02-01',
      expectedEndDate: '2024-04-30',
      assignedAttorney: 'test.attorney',
      teamMembers: ['paralegal1'],
      tags: ['劳动', '合同解除', '赔偿'],
      documents: ['employment_contract.pdf'],
      budget: 25000,
      actualCost: 8000,
      notes: '正在准备调解材料',
      court: '北京市海淀区人民法院',
      judge: '王法官',
      opposingCounsel: '对方律师事务所',
      keyDates: [
        { date: '2024-02-01', event: '咨询', description: '初次咨询' },
        { date: '2024-02-15', event: '受理', description: '正式受理' }
      ],
      milestones: [
        { name: '受理', date: '2024-02-15', completed: true },
        { name: '调解', date: '2024-03-15', completed: false },
        { name: '仲裁', date: '2024-04-30', completed: false }
      ]
    },
    {
      id: 'case-003',
      title: '知识产权侵权案',
      caseNumber: 'IP-2024-003',
      clientName: '王五创意工作室',
      clientType: 'corporate',
      caseType: 'intellectual_property',
      priority: 'high',
      status: 'closed',
      description: '商标侵权和不正当竞争纠纷',
      expectedOutcome: 'favorable',
      estimatedValue: 800000,
      startDate: '2023-10-01',
      expectedEndDate: '2024-03-15',
      assignedAttorney: 'test.attorney',
      teamMembers: ['paralegal2', 'assistant2'],
      tags: ['商标', '侵权', '不正当竞争'],
      documents: ['trademark.pdf', 'evidence_collection.pdf', 'expert_report.pdf'],
      budget: 120000,
      actualCost: 95000,
      notes: '案件已胜诉，正在执行赔偿',
      court: '北京市知识产权法院',
      judge: '赵法官',
      opposingCounsel: '对方律师事务所',
      keyDates: [
        { date: '2023-10-01', event: '立案', description: '正式立案' },
        { date: '2024-01-20', event: '开庭', description: '正式开庭' },
        { date: '2024-02-28', event: '判决', description: '获得有利判决' },
        { date: '2024-03-15', event: '结案', description: '案件正式结案' }
      ],
      milestones: [
        { name: '立案', date: '2023-10-01', completed: true },
        { name: '证据收集', date: '2023-11-30', completed: true },
        { name: '开庭', date: '2024-01-20', completed: true },
        { name: '判决', date: '2024-02-28', completed: true },
        { name: '结案', date: '2024-03-15', completed: true }
      ]
    }
  ] as CaseData[],

  // 测试文档数据
  testDocuments: [
    {
      id: 'doc-001',
      name: '合同文件',
      fileName: 'contract.pdf',
      fileType: 'pdf',
      fileSize: 1024000,
      contentType: 'application/pdf',
      caseId: 'case-001',
      documentType: 'contract',
      category: 'legal',
      description: '服务合同原件',
      tags: ['合同', '原件'],
      uploadDate: '2024-01-15',
      lastModified: '2024-01-15',
      uploadedBy: 'test.attorney',
      version: 1,
      isConfidential: false,
      isRequired: true,
      status: 'active',
      storagePath: '/cases/case-001/documents/contract.pdf',
      checksum: 'abc123def456'
    },
    {
      id: 'doc-002',
      name: '证据材料',
      fileName: 'evidence.pdf',
      fileType: 'pdf',
      fileSize: 2048000,
      contentType: 'application/pdf',
      caseId: 'case-001',
      documentType: 'evidence',
      category: 'evidence',
      description: '案件相关证据材料',
      tags: ['证据', '材料'],
      uploadDate: '2024-02-01',
      lastModified: '2024-02-10',
      uploadedBy: 'paralegal1',
      version: 2,
      isConfidential: true,
      isRequired: true,
      status: 'active',
      storagePath: '/cases/case-001/documents/evidence_v2.pdf',
      checksum: 'def456ghi789'
    },
    {
      id: 'doc-003',
      name: '劳动合同',
      fileName: 'employment_contract.pdf',
      fileType: 'pdf',
      fileSize: 512000,
      contentType: 'application/pdf',
      caseId: 'case-002',
      documentType: 'contract',
      category: 'legal',
      description: '员工劳动合同',
      tags: ['合同', '劳动'],
      uploadDate: '2024-02-15',
      lastModified: '2024-02-15',
      uploadedBy: 'test.attorney',
      version: 1,
      isConfidential: false,
      isRequired: true,
      status: 'active',
      storagePath: '/cases/case-002/documents/employment_contract.pdf',
      checksum: 'ghi789jkl012'
    }
  ] as Document[]
};

// 案件类型配置
export const CASE_TYPES = {
  contract: {
    name: '合同纠纷',
    description: '涉及各类合同的纠纷案件',
    requiredFields: ['contractDate', 'contractValue', 'breachType'],
    typicalDocuments: ['contract', 'evidence', 'correspondence']
  },
  employment: {
    name: '劳动争议',
    description: '员工与用人单位之间的劳动纠纷',
    requiredFields: ['employmentDate', 'terminationDate', 'disputeType'],
    typicalDocuments: ['employment_contract', 'payroll', 'correspondence']
  },
  intellectual_property: {
    name: '知识产权',
    description: '商标、专利、著作权等知识产权纠纷',
    requiredFields: ['ipType', 'registrationNumber', 'infringementType'],
    typicalDocuments: ['registration_certificate', 'evidence', 'expert_report']
  },
  real_estate: {
    name: '房地产',
    description: '房地产买卖、租赁等纠纷',
    requiredFields: ['propertyAddress', 'propertyValue', 'transactionType'],
    typicalDocuments: ['property_deed', 'contract', 'appraisal']
  },
  family: {
    name: '婚姻家庭',
    description: '离婚、继承、抚养等家庭纠纷',
    requiredFields: ['familyMembers', 'disputeType', 'assetsInvolved'],
    typicalDocuments: ['marriage_certificate', 'asset_list', 'agreement']
  },
  criminal: {
    name: '刑事案件',
    description: '各类刑事案件',
    requiredFields: ['charges', 'arrestDate', 'courtDate'],
    typicalDocuments: ['police_report', 'evidence', 'witness_statements']
  }
};

// 案件状态配置
export const CASE_STATUS = {
  draft: {
    name: '草稿',
    description: '案件信息正在准备中',
    color: '#gray',
    nextStatus: ['pending', 'active']
  },
  pending: {
    name: '待处理',
    description: '案件已受理，等待分配律师',
    color: '#orange',
    nextStatus: ['active', 'rejected']
  },
  active: {
    name: '进行中',
    description: '案件正在积极处理中',
    color: '#blue',
    nextStatus: ['paused', 'completed', 'closed']
  },
  paused: {
    name: '暂停',
    description: '案件处理暂时停止',
    color: '#yellow',
    nextStatus: ['active', 'closed']
  },
  completed: {
    name: '已完成',
    description: '案件主要工作已完成',
    color: '#green',
    nextStatus: ['closed']
  },
  closed: {
    name: '已结案',
    description: '案件已正式结案',
    color: '#purple',
    nextStatus: []
  },
  rejected: {
    name: '已拒绝',
    description: '案件申请被拒绝',
    color: '#red',
    nextStatus: []
  }
};

// 案件优先级配置
export const CASE_PRIORITY = {
  low: {
    name: '低',
    description: '一般优先级案件',
    color: '#green',
    urgency: 'normal'
  },
  medium: {
    name: '中',
    description: '中等优先级案件',
    color: '#yellow',
    urgency: 'medium'
  },
  high: {
    name: '高',
    description: '高优先级案件',
    color: '#orange',
    urgency: 'high'
  },
  urgent: {
    name: '紧急',
    description: '紧急案件，需要立即处理',
    color: '#red',
    urgency: 'critical'
  }
};

// 搜索和筛选配置
export const SEARCH_CONFIG = {
  fields: [
    'title',
    'caseNumber',
    'clientName',
    'caseType',
    'status',
    'priority',
    'assignedAttorney',
    'tags'
  ],
  operators: {
    equals: '等于',
    contains: '包含',
    startsWith: '开头是',
    endsWith: '结尾是',
    greaterThan: '大于',
    lessThan: '小于',
    between: '介于'
  },
  dateFields: [
    'startDate',
    'expectedEndDate',
    'createdDate',
    'lastModifiedDate'
  ],
  numericFields: [
    'estimatedValue',
    'budget',
    'actualCost'
  ]
};

// 排序配置
export const SORT_CONFIG = {
  fields: [
    { field: 'title', label: '案件标题', type: 'string' },
    { field: 'caseNumber', label: '案件编号', type: 'string' },
    { field: 'clientName', label: '客户名称', type: 'string' },
    { field: 'startDate', label: '开始日期', type: 'date' },
    { field: 'expectedEndDate', label: '预计结束日期', type: 'date' },
    { field: 'priority', label: '优先级', type: 'enum' },
    { field: 'status', label: '状态', type: 'enum' },
    { field: 'estimatedValue', label: '预估价值', type: 'number' },
    { field: 'createdDate', label: '创建日期', type: 'date' },
    { field: 'lastModifiedDate', label: '最后修改', type: 'date' }
  ],
  defaultSort: { field: 'createdDate', direction: 'desc' }
};

// 导出配置
export const EXPORT_CONFIG = {
  formats: ['pdf', 'excel', 'csv', 'json'],
  defaultFormat: 'pdf',
  includeFields: [
    'caseNumber',
    'title',
    'clientName',
    'caseType',
    'status',
    'priority',
    'assignedAttorney',
    'startDate',
    'expectedEndDate',
    'estimatedValue',
    'budget',
    'actualCost'
  ],
  dateRangeOptions: [
    { label: '今天', value: 'today' },
    { label: '本周', value: 'thisWeek' },
    { label: '本月', value: 'thisMonth' },
    { label: '本季度', value: 'thisQuarter' },
    { label: '本年', value: 'thisYear' },
    { label: '自定义', value: 'custom' }
  ]
};

// 权限配置
export const PERMISSION_CONFIG = {
  roles: {
    admin: {
      permissions: ['cases.read', 'cases.write', 'cases.delete', 'cases.export', 'cases.assign']
    },
    attorney: {
      permissions: ['cases.read', 'cases.write', 'cases.export', 'cases.assign']
    },
    paralegal: {
      permissions: ['cases.read', 'cases.write']
    },
    assistant: {
      permissions: ['cases.read']
    }
  },
  actions: {
    read: {
      label: '查看',
      description: '查看案件信息'
    },
    write: {
      label: '编辑',
      description: '编辑案件信息'
    },
    delete: {
      label: '删除',
      description: '删除案件'
    },
    export: {
      label: '导出',
      description: '导出案件数据'
    },
    assign: {
      label: '分配',
      description: '分配案件给律师'
    }
  }
};

// 验证规则
export const VALIDATION_RULES = {
  required: {
    title: { message: '案件标题不能为空' },
    caseNumber: { message: '案件编号不能为空' },
    clientName: { message: '客户名称不能为空' },
    caseType: { message: '案件类型不能为空' },
    priority: { message: '优先级不能为空' },
    status: { message: '状态不能为空' },
    startDate: { message: '开始日期不能为空' }
  },
  format: {
    email: {
      pattern: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
      message: '邮箱格式不正确'
    },
    phone: {
      pattern: /^[\+]?[1-9][\d]{0,15}$/,
      message: '电话号码格式不正确'
    },
    caseNumber: {
      pattern: /^[A-Z]{2}-\d{4}-\d{3}$/,
      message: '案件编号格式不正确（如：CL-2024-001）'
    }
  },
  range: {
    estimatedValue: {
      min: 0,
      message: '预估价值不能为负数'
    },
    budget: {
      min: 0,
      message: '预算不能为负数'
    }
  },
  date: {
    startDate: {
      before: 'expectedEndDate',
      message: '开始日期不能晚于预计结束日期'
    }
  }
};

// 测试工具函数
export const CaseTestUtils = {
  /**
   * 生成随机案件编号
   */
  generateCaseNumber(prefix: string = 'CL'): string {
    const year = new Date().getFullYear();
    const sequence = Math.floor(Math.random() * 999) + 1;
    return `${prefix}-${year}-${sequence.toString().padStart(3, '0')}`;
  },

  /**
   * 生成测试案件数据
   */
  generateTestCase(overrides: Partial<CaseData> = {}): CaseData {
    const baseCase: CaseData = {
      id: TestUtils.generateRandomUsername('case'),
      title: '测试案件 ' + Math.random().toString(36).substring(7),
      caseNumber: this.generateCaseNumber(),
      clientName: '测试客户 ' + Math.random().toString(36).substring(7),
      clientType: 'individual',
      caseType: 'contract',
      priority: 'medium',
      status: 'active',
      description: '这是一个测试案件',
      expectedOutcome: 'favorable',
      estimatedValue: Math.floor(Math.random() * 1000000),
      startDate: new Date().toISOString().split('T')[0],
      expectedEndDate: new Date(Date.now() + 90 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
      assignedAttorney: 'test.attorney',
      teamMembers: [],
      tags: ['测试'],
      documents: [],
      budget: Math.floor(Math.random() * 100000),
      actualCost: Math.floor(Math.random() * 50000),
      notes: '测试案件备注',
      court: '测试法院',
      judge: '测试法官',
      opposingCounsel: '对方律师事务所',
      keyDates: [],
      milestones: []
    };

    return { ...baseCase, ...overrides };
  },

  /**
   * 生成测试文档数据
   */
  generateDocument(caseId: string, overrides: Partial<Document> = {}): Document {
    const baseDocument: Document = {
      id: TestUtils.generateRandomUsername('doc'),
      name: '测试文档.pdf',
      fileName: 'test_document.pdf',
      fileType: 'pdf',
      fileSize: Math.floor(Math.random() * 10000000),
      contentType: 'application/pdf',
      caseId,
      documentType: 'evidence',
      category: 'evidence',
      description: '这是一个测试文档',
      tags: ['测试'],
      uploadDate: new Date().toISOString().split('T')[0],
      lastModified: new Date().toISOString().split('T')[0],
      uploadedBy: 'test.attorney',
      version: 1,
      isConfidential: false,
      isRequired: true,
      status: 'active',
      storagePath: `/cases/${caseId}/documents/test_document.pdf`,
      checksum: Math.random().toString(36).substring(7)
    };

    return { ...baseDocument, ...overrides };
  },

  /**
   * 验证案件数据
   */
  validateCaseData(caseData: CaseData): { isValid: boolean; errors: string[] } {
    const errors: string[] | undefined = undefined;

    // 验证必填字段
    const requiredFields = ['title', 'caseNumber', 'clientName', 'caseType', 'priority', 'status', 'startDate'];
    for (const field of requiredFields) {
      if (!caseData[field as keyof CaseData]) {
        errors.push(`${field} is required`);
      }
    }

    // 验证日期格式
    const dateFields = ['startDate', 'expectedEndDate'];
    for (const field of dateFields) {
      const value = caseData[field as keyof CaseData];
      if (value && isNaN(Date.parse(value as string))) {
        errors.push(`${field} has invalid date format`);
      }
    }

    // 验证数值字段
    const numericFields = ['estimatedValue', 'budget', 'actualCost'];
    for (const field of numericFields) {
      const value = caseData[field as keyof CaseData];
      if (value && typeof value === 'number' && value < 0) {
        errors.push(`${field} cannot be negative`);
      }
    }

    return {
      isValid: errors.length === 0,
      errors
    };
  }
};

export default CASE_TEST_CONFIG;