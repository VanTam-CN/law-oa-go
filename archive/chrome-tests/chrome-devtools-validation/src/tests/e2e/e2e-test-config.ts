/**
 * 端到端测试配置
 */

import { CaseData, DocumentData } from '../../types/test-types';

export interface E2ETestConfig {
  baseUrl: string;
  defaultTimeout: number;
  screenshotOnFailure: boolean;
  parallelExecution: boolean;
  retryAttempts: number;
  testData: E2ETestData;
  environments: E2EEnvironmentConfig;
  workflows: E2EWorkflowConfig;
}

export interface E2ETestData {
  users: {
    attorney: E2ETestUser;
    paralegal: E2ETestUser;
    admin: E2ETestUser;
  };
  cases: CaseData[];
  documents: DocumentData[];
  clients: E2ETestClient[];
}

export interface E2ETestUser {
  id: string;
  username: string;
  email: string;
  password: string;
  firstName: string;
  lastName: string;
  role: 'attorney' | 'paralegal' | 'admin';
  department: string;
  permissions: string[];
}

export interface E2ETestClient {
  id: string;
  name: string;
  type: 'individual' | 'corporate';
  email?: string;
  phone?: string;
  address?: string;
  industry?: string;
  contactPerson?: string;
}

export interface E2EEnvironmentConfig {
  development: {
    baseUrl: string;
    apiBaseUrl: string;
    timeout: number;
    headless: boolean;
    screenshots: boolean;
  };
  staging: {
    baseUrl: string;
    apiBaseUrl: string;
    timeout: number;
    headless: boolean;
    screenshots: boolean;
  };
  production: {
    baseUrl: string;
    apiBaseUrl: string;
    timeout: number;
    headless: boolean;
    screenshots: boolean;
  };
}

export interface E2EWorkflowConfig {
  clientIntake: {
    enabled: boolean;
    steps: string[];
    expectedDuration: number;
  };
  caseManagement: {
    enabled: boolean;
    steps: string[];
    expectedDuration: number;
  };
  documentManagement: {
    enabled: boolean;
    steps: string[];
    expectedDuration: number;
  };
  financialTracking: {
    enabled: boolean;
    steps: string[];
    expectedDuration: number;
  };
  conflictCheck: {
    enabled: boolean;
    steps: string[];
    expectedDuration: number;
  };
  completeLifecycle: {
    enabled: boolean;
    steps: string[];
    expectedDuration: number;
  };
}

export const E2E_TEST_CONFIG: E2ETestConfig = {
  baseUrl: 'http://localhost:3000',
  defaultTimeout: 60000,
  screenshotOnFailure: true,
  parallelExecution: false,
  retryAttempts: 2,

  testData: {
    users: {
      attorney: {
        id: 'e2e-attorney-001',
        username: 'e2e.attorney',
        email: 'e2e.attorney@lawfirm.com',
        password: 'E2ETestPassword123!',
        firstName: 'E2E',
        lastName: 'Attorney',
        role: 'attorney',
        department: 'litigation',
        permissions: [
          'cases.read',
          'cases.write',
          'cases.delete',
          'clients.read',
          'clients.write',
          'documents.read',
          'documents.write',
          'finance.read',
          'finance.write'
        ]
      },
      paralegal: {
        id: 'e2e-paralegal-001',
        username: 'e2e.paralegal',
        email: 'e2e.paralegal@lawfirm.com',
        password: 'E2ETestPassword123!',
        firstName: 'E2E',
        lastName: 'Paralegal',
        role: 'paralegal',
        department: 'litigation',
        permissions: [
          'cases.read',
          'cases.write',
          'clients.read',
          'documents.read',
          'documents.write'
        ]
      },
      admin: {
        id: 'e2e-admin-001',
        username: 'e2e.admin',
        email: 'e2e.admin@lawfirm.com',
        password: 'E2ETestPassword123!',
        firstName: 'E2E',
        lastName: 'Admin',
        role: 'admin',
        department: 'it',
        permissions: [
          'admin.read',
          'admin.write',
          'cases.read',
          'cases.write',
          'cases.delete',
          'clients.read',
          'clients.write',
          'clients.delete',
          'documents.read',
          'documents.write',
          'documents.delete',
          'finance.read',
          'finance.write',
          'users.read',
          'users.write'
        ]
      }
    },

    cases: [
      {
        id: 'e2e-case-001',
        title: 'E2E测试案件 - 合同纠纷',
        caseNumber: 'E2E-2024-001',
        clientName: 'E2E测试客户公司',
        clientType: 'corporate',
        caseType: 'contract',
        priority: 'high',
        status: 'active',
        description: '用于端到端测试的合同纠纷案件',
        expectedOutcome: 'favorable',
        estimatedValue: 500000,
        startDate: '2024-01-15',
        expectedEndDate: '2024-06-30',
        assignedAttorney: 'e2e.attorney',
        teamMembers: ['e2e.paralegal'],
        tags: ['E2E', '测试', '合同'],
        documents: [],
        budget: 80000,
        actualCost: 0,
        notes: '端到端测试案件',
        court: '测试法院',
        judge: '测试法官',
        opposingCounsel: '对方律师事务所',
        keyDates: [],
        milestones: []
      }
    ] as CaseData[],

    documents: [
      {
        id: 'e2e-doc-001',
        name: 'E2E测试合同',
        fileName: 'e2e_test_contract.pdf',
        fileType: 'pdf',
        fileSize: 1024000,
        contentType: 'application/pdf',
        caseId: 'e2e-case-001',
        documentType: 'contract',
        category: 'legal',
        description: '用于端到端测试的合同文档',
        tags: ['E2E', '测试', '合同'],
        uploadDate: '2024-01-15',
        lastModified: '2024-01-15',
        uploadedBy: 'e2e.attorney',
        version: 1,
        isConfidential: false,
        isRequired: true,
        status: 'active',
        storagePath: '/cases/e2e-case-001/documents/e2e_test_contract.pdf',
        checksum: 'e2e-test-checksum-001'
      }
    ] as Document[],

    clients: [
      {
        id: 'e2e-client-001',
        name: 'E2E测试客户公司',
        type: 'corporate',
        email: 'e2e.client@company.com',
        phone: '+1-555-0123',
        address: '测试地址123号',
        industry: '科技',
        contactPerson: '测试联系人'
      },
      {
        id: 'e2e-client-002',
        name: 'E2E测试个人客户',
        type: 'individual',
        email: 'e2e.individual@email.com',
        phone: '+1-555-0456',
        address: '个人测试地址456号'
      }
    ] as E2ETestClient[]
  },

  environments: {
    development: {
      baseUrl: 'http://localhost:3000',
      apiBaseUrl: 'http://localhost:3001/api',
      timeout: 60000,
      headless: false,
      screenshots: true
    },
    staging: {
      baseUrl: 'https://staging.lawfirm.com',
      apiBaseUrl: 'https://staging-api.lawfirm.com/api',
      timeout: 90000,
      headless: true,
      screenshots: true
    },
    production: {
      baseUrl: 'https://app.lawfirm.com',
      apiBaseUrl: 'https://api.lawfirm.com/api',
      timeout: 120000,
      headless: true,
      screenshots: false
    }
  },

  workflows: {
    clientIntake: {
      enabled: true,
      steps: [
        '用户登录',
        '创建新客户',
        '客户信息验证',
        '记录初始咨询',
        '生成客户报告'
      ],
      expectedDuration: 300000 // 5分钟
    },
    caseManagement: {
      enabled: true,
      steps: [
        '用户登录',
        '创建新案件',
        '分配律师',
        '设置里程碑',
        '更新案件状态'
      ],
      expectedDuration: 420000 // 7分钟
    },
    documentManagement: {
      enabled: true,
      steps: [
        '用户登录',
        '上传案件文档',
        '文档分类和标记',
        '文档版本控制',
        '文档权限设置'
      ],
      expectedDuration: 360000 // 6分钟
    },
    financialTracking: {
      enabled: true,
      steps: [
        '用户登录',
        '创建财务记录',
        '预算管理',
        '开票和收费',
        '财务报告生成'
      ],
      expectedDuration: 480000 // 8分钟
    },
    conflictCheck: {
      enabled: true,
      steps: [
        '用户登录',
        '案前冲突检查',
        '分析冲突结果',
        '冲突解决流程',
        '审批和记录'
      ],
      expectedDuration: 300000 // 5分钟
    },
    completeLifecycle: {
      enabled: true,
      steps: [
        '用户登录',
        '客户 intake',
        '冲突检查',
        '案件创建',
        '团队分配',
        '文档管理',
        '案件处理',
        '财务管理',
        '状态更新',
        '案件结案'
      ],
      expectedDuration: 900000 // 15分钟
    }
  }
};

// 工作流配置
export const WORKFLOW_CONFIG = {
  // 超时设置
  timeouts: {
    default: 60000,
    pageLoad: 30000,
    elementWait: 10000,
    actionWait: 5000,
    assertionWait: 3000
  },

  // 重试设置
  retries: {
    default: 2,
    critical: 3,
    nonCritical: 1
  },

  // 截图设置
  screenshots: {
    onFailure: true,
    onSuccess: false,
    onStep: false,
    path: './test-screenshots/e2e/'
  },

  // 报告设置
  reporting: {
    format: 'json',
    includeScreenshots: true,
    includeLogs: true,
    includePerformance: true,
    outputDir: './test-reports/e2e/'
  },

  // 性能监控
  performance: {
    enabled: true,
    metrics: [
      'pageLoadTime',
      'stepExecutionTime',
      'totalWorkflowTime',
      'memoryUsage',
      'networkRequests'
    ],
    thresholds: {
      pageLoadTime: 5000,
      stepExecutionTime: 10000,
      totalWorkflowTime: 300000
    }
  },

  // 数据清理
  cleanup: {
    enabled: true,
    autoCleanup: true,
    entities: [
      'cases',
      'documents',
      'clients',
      'financial_records',
      'user_sessions'
    ]
  }
};

// 测试场景配置
export const TEST_SCENARIOS = {
  happyPath: {
    name: '正常流程',
    description: '标准的业务流程测试',
    priority: 'high'
  },
  edgeCases: {
    name: '边界情况',
    description: '测试边界条件和异常情况',
    priority: 'medium'
  },
  errorHandling: {
    name: '错误处理',
    description: '测试系统错误处理能力',
    priority: 'high'
  },
  performance: {
    name: '性能测试',
    description: '测试系统性能和响应时间',
    priority: 'medium'
  },
  security: {
    name: '安全测试',
    description: '测试系统安全性',
    priority: 'high'
  }
};

// 断言配置
export const ASSERTION_CONFIG = {
  timeout: 10000,
  retry: 3,
  retryDelay: 1000,
  strict: true,
  verbose: true,
  includeStackTrace: true
};

// 数据生成器配置
export const DATA_GENERATOR_CONFIG = {
  useRealisticData: true,
  locale: 'zh-CN',
  seed: 'e2e-test-seed',
  customGenerators: {
    caseNumber: () => `E2E-${Date.now()}-${Math.floor(Math.random() * 1000)}`,
    clientName: () => `E2E客户${Math.floor(Math.random() * 1000)}`,
    documentName: () => `E2E文档${Math.floor(Math.random() * 1000)}.pdf`
  }
};

// 集成配置
export const INTEGRATION_CONFIG = {
  externalServices: {
    email: {
      enabled: true,
      mock: true,
      config: {
        host: 'smtp.test.com',
        port: 587,
        secure: false
      }
    },
    storage: {
      enabled: true,
      mock: true,
      config: {
        provider: 'local',
        path: './test-storage/'
      }
    },
    payment: {
      enabled: false,
      mock: true
    }
  }
};

export default E2E_TEST_CONFIG;