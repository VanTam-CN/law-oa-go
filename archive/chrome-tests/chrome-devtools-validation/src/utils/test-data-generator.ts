/**
 * 测试数据生成器 - 生成各种类型的测试数据
 */

import { Logger } from '../core/logger';

export interface UserData {
  id: string;
  username: string;
  email: string;
  password: string;
  firstName: string;
  lastName: string;
  role: 'admin' | 'attorney' | 'paralegal' | 'client';
  department?: string;
  phone?: string;
  status: 'active' | 'inactive' | 'pending';
  createdAt: Date;
  updatedAt: Date;
}

export interface ClientData {
  id: string;
  name: string;
  type: 'individual' | 'corporate';
  registrationNumber?: string;
  industry?: string;
  contactPerson?: string;
  email?: string;
  phone?: string;
  address?: {
    street: string;
    city: string;
    state: string;
    zipCode: string;
    country: string;
  };
  status: 'active' | 'inactive' | 'prospect';
  createdAt: Date;
  updatedAt: Date;
}

export interface CaseData {
  id: string;
  title: string;
  caseNumber: string;
  type: 'litigation' | 'corporate' | 'family' | 'criminal' | 'other';
  status: 'active' | 'closed' | 'pending' | 'archived';
  priority: 'low' | 'medium' | 'high' | 'urgent';
  assignedTo: string;
  client: string;
  description?: string;
  estimatedValue?: number;
  createdDate: Date;
  updatedDate: Date;
  dueDate?: Date;
}

export interface DocumentData {
  id: string;
  name: string;
  type: 'contract' | 'evidence' | 'correspondence' | 'court_filing' | 'other';
  caseId: string;
  clientId?: string;
  uploadedBy: string;
  fileSize: number;
  mimeType: string;
  path: string;
  tags: string[];
  status: 'active' | 'archived' | 'deleted';
  uploadedAt: Date;
  updatedAt: Date;
}

export interface FinancialData {
  id: string;
  caseId: string;
  clientId: string;
  type: 'fee' | 'expense' | 'payment' | 'invoice';
  description: string;
  amount: number;
  currency: string;
  date: Date;
  status: 'pending' | 'paid' | 'overdue' | 'cancelled';
  createdAt: Date;
  updatedAt: Date;
}

export class TestDataGenerator {
  private logger: Logger;
  private randomSeed: number;

  constructor(logger?: Logger) {
    this.logger = logger || new Logger('TestDataGenerator');
    this.randomSeed = Date.now();
  }

  /**
   * 设置随机种子
   */
  setSeed(seed: number): void {
    this.randomSeed = seed;
    this.logger.debug('随机种子已设置', { seed });
  }

  /**
   * 生成随机数
   */
  private random(min: number, max: number): number {
    this.randomSeed = (this.randomSeed * 9301 + 49297) % 233280;
    const r = this.randomSeed / 233280;
    return Math.floor(r * (max - min + 1)) + min;
  }

  /**
   * 从数组中随机选择
   */
  private randomChoice<T>(array: T[]): T {
    return array[this.random(0, array.length - 1)];
  }

  /**
   * 生成随机字符串
   */
  private randomString(length: number, charset: string = 'abcdefghijklmnopqrstuvwxyz'): string {
    let result = '';
    for (let i = 0; i < length; i++) {
      result += charset.charAt(this.random(0, charset.length - 1));
    }
    return result;
  }

  /**
   * 生成随机邮箱
   */
  private randomEmail(): string {
    const username = this.randomString(8, 'abcdefghijklmnopqrstuvwxyz0123456789');
    const domains = ['example.com', 'test.org', 'demo.net', 'sample.io'];
    return `${username}@${this.randomChoice(domains)}`;
  }

  /**
   * 生成随机电话号码
   */
  private randomPhone(): string {
    const area = this.random(100, 999);
    const prefix = this.random(100, 999);
    const line = this.random(1000, 9999);
    return `(${area}) ${prefix}-${line}`;
  }

  /**
   * 生成随机日期
   */
  private randomDate(start: Date, end: Date): Date {
    return new Date(start.getTime() + Math.random() * (end.getTime() - start.getTime()));
  }

  /**
   * 生成随机金额
   */
  private randomAmount(min: number, max: number): number {
    return Math.round((min + Math.random() * (max - min)) * 100) / 100;
  }

  /**
   * 生成用户数据
   */
  generateUserData(overrides: Partial<UserData> = {}): UserData {
    const roles: Array<UserData['role']> = ['admin', 'attorney', 'paralegal', 'client'];
    const statuses: Array<UserData['status']> = ['active', 'inactive', 'pending'];
    const departments = [' litigation', 'corporate', 'family', 'criminal', 'general'];

    const firstName = this.randomChoice([
      '张', '李', '王', '刘', '陈', '杨', '赵', '黄', '周', '吴',
      '郑', '孙', '马', '朱', '胡', '林', '郭', '何', '高', '罗'
    ]);

    const lastName = this.randomChoice([
      '伟', '芳', '娜', '秀英', '敏', '静', '丽', '强', '磊', '军',
      '洋', '勇', '艳', '杰', '娟', '涛', '明', '超', '秀兰', '霞'
    ]);

    const baseData: UserData = {
      id: this.randomString(20, 'abcdefghijklmnopqrstuvwxyz0123456789'),
      username: `${firstName}${lastName}${this.random(1, 99)}`,
      email: this.randomEmail(),
      password: `Test@${this.randomString(8)}!`,
      firstName,
      lastName,
      role: this.randomChoice(roles),
      department: this.randomChoice(departments),
      phone: this.randomPhone(),
      status: this.randomChoice(statuses),
      createdAt: this.randomDate(new Date('2020-01-01'), new Date()),
      updatedAt: new Date()
    };

    return { ...baseData, ...overrides };
  }

  /**
   * 生成客户数据
   */
  generateClientData(overrides: Partial<ClientData> = {}): ClientData {
    const types: Array<ClientData['type']> = ['individual', 'corporate'];
    const statuses: Array<ClientData['status']> = ['active', 'inactive', 'prospect'];
    const industries = ['technology', 'finance', 'healthcare', 'education', 'retail', 'manufacturing'];

    const companyNames = [
      '科技有限公司', '信息有限公司', '网络科技', '数字科技', '智能科技',
      '创新科技', '未来科技', '智慧科技', '云端科技', '数据科技'
    ];

    const individualNames = [
      '张伟', '李娜', '王芳', '刘敏', '陈静', '杨丽', '赵强', '黄磊', '周军', '吴勇'
    ];

    const baseData: ClientData = {
      id: this.randomString(20, 'abcdefghijklmnopqrstuvwxyz0123456789'),
      name: this.randomChoice(types) === 'corporate'
        ? this.randomChoice(companyNames)
        : this.randomChoice(individualNames),
      type: this.randomChoice(types),
      registrationNumber: this.randomString(10, '0123456789'),
      industry: this.randomChoice(industries),
      contactPerson: this.randomChoice(individualNames),
      email: this.randomEmail(),
      phone: this.randomPhone(),
      address: {
        street: `${this.random(1, 999)}号街道`,
        city: this.randomChoice(['北京', '上海', '广州', '深圳', '杭州', '南京', '成都', '武汉']),
        state: this.randomChoice(['北京', '上海', '广东', '江苏', '浙江', '四川', '湖北']),
        zipCode: this.randomString(6, '0123456789'),
        country: '中国'
      },
      status: this.randomChoice(statuses),
      createdAt: this.randomDate(new Date('2020-01-01'), new Date()),
      updatedAt: new Date()
    };

    return { ...baseData, ...overrides };
  }

  /**
   * 生成案件数据
   */
  generateCaseData(overrides: Partial<CaseData> = {}): CaseData {
    const types: Array<CaseData['type']> = ['litigation', 'corporate', 'family', 'criminal', 'other'];
    const statuses: Array<CaseData['status']> = ['active', 'closed', 'pending', 'archived'];
    const priorities: Array<CaseData['priority']> = ['low', 'medium', 'high', 'urgent'];

    const caseTitles = [
      '合同纠纷', '知识产权侵权', '劳动争议', '股权纠纷', '债务追偿',
      '房地产纠纷', '婚姻家庭', '继承纠纷', '侵权责任', '公司法律事务'
    ];

    const attorneys = ['张律师', '李律师', '王律师', '刘律师', '陈律师', '杨律师'];

    const baseData: CaseData = {
      id: this.randomString(20, 'abcdefghijklmnopqrstuvwxyz0123456789'),
      title: this.randomChoice(caseTitles),
      caseNumber: `(${new Date().getFullYear()})${this.randomString(4, 'ABCDEFGHIJKLMNOPQRSTUVWXYZ')}${this.randomString(6, '0123456789')}`,
      type: this.randomChoice(types),
      status: this.randomChoice(statuses),
      priority: this.randomChoice(priorities),
      assignedTo: this.randomChoice(attorneys),
      client: this.randomChoice(['科技公司', '制造企业', '服务公司', '个人客户']),
      description: `这是一个关于${this.randomChoice(caseTitles)}的案件，需要专业的法律服务。`,
      estimatedValue: this.randomAmount(10000, 1000000),
      createdDate: this.randomDate(new Date('2020-01-01'), new Date()),
      updatedDate: new Date(),
      dueDate: this.randomDate(new Date(), new Date(Date.now() + 90 * 24 * 60 * 60 * 1000))
    };

    return { ...baseData, ...overrides };
  }

  /**
   * 生成文档数据
   */
  generateDocumentData(caseId: string, overrides: Partial<DocumentData> = {}): DocumentData {
    const types: Array<DocumentData['type']> = ['contract', 'evidence', 'correspondence', 'court_filing', 'other'];
    const statuses: Array<DocumentData['status']> = ['active', 'archived', 'deleted'];
    const mimeTypes = ['application/pdf', 'application/msword', 'text/plain', 'image/jpeg'];

    const documentNames = [
      '合同文件.pdf', '证据材料.pdf', '法院文书.pdf', '邮件往来.pdf',
      '会议记录.pdf', '法律意见书.pdf', '起诉状.pdf', '答辩状.pdf'
    ];

    const baseData: DocumentData = {
      id: this.randomString(20, 'abcdefghijklmnopqrstuvwxyz0123456789'),
      name: this.randomChoice(documentNames),
      type: this.randomChoice(types),
      caseId,
      uploadedBy: this.randomChoice(['张律师', '李律师', '王律师']),
      fileSize: this.randomAmount(1024, 10485760), // 1KB to 10MB
      mimeType: this.randomChoice(mimeTypes),
      path: `/documents/${caseId}/${this.randomString(10)}_${this.randomChoice(documentNames)}`,
      tags: [this.randomChoice(['重要', '紧急', '待审核', '已归档'])],
      status: this.randomChoice(statuses),
      uploadedAt: this.randomDate(new Date('2020-01-01'), new Date()),
      updatedAt: new Date()
    };

    return { ...baseData, ...overrides };
  }

  /**
   * 生成财务数据
   */
  generateFinancialData(caseId: string, clientId: string, overrides: Partial<FinancialData> = {}): FinancialData {
    const types: Array<FinancialData['type']> = ['fee', 'expense', 'payment', 'invoice'];
    const statuses: Array<FinancialData['status']> = ['pending', 'paid', 'overdue', 'cancelled'];

    const descriptions = [
      '律师费', '诉讼费', '咨询费', '文件处理费', '差旅费', '调查费',
      '专家费', '翻译费', '公证费', '执行费', '保全费', '其他费用'
    ];

    const baseData: FinancialData = {
      id: this.randomString(20, 'abcdefghijklmnopqrstuvwxyz0123456789'),
      caseId,
      clientId,
      type: this.randomChoice(types),
      description: this.randomChoice(descriptions),
      amount: this.randomAmount(100, 50000),
      currency: 'CNY',
      date: this.randomDate(new Date('2020-01-01'), new Date()),
      status: this.randomChoice(statuses),
      createdAt: this.randomDate(new Date('2020-01-01'), new Date()),
      updatedAt: new Date()
    };

    return { ...baseData, ...overrides };
  }

  /**
   * 批量生成用户数据
   */
  generateUsers(count: number): UserData[] {
    const users: UserData[] | undefined = undefined;
    for (let i = 0; i < count; i++) {
      users.push(this.generateUserData());
    }
    return users;
  }

  /**
   * 批量生成客户数据
   */
  generateClients(count: number): ClientData[] {
    const clients: ClientData[] | undefined = undefined;
    for (let i = 0; i < count; i++) {
      clients.push(this.generateClientData());
    }
    return clients;
  }

  /**
   * 批量生成案件数据
   */
  generateCases(count: number): CaseData[] {
    const cases: CaseData[] | undefined = undefined;
    for (let i = 0; i < count; i++) {
      cases.push(this.generateCaseData());
    }
    return cases;
  }

  /**
   * 生成完整的测试数据集
   */
  generateTestDataSet(options: {
    userCount?: number;
    clientCount?: number;
    caseCount?: number;
    documentsPerCase?: number;
    financialRecordsPerCase?: number;
  } = {}): {
    users: UserData[];
    clients: ClientData[];
    cases: CaseData[];
    documents: DocumentData[];
    financial: FinancialData[];
  } {
    const {
      userCount = 10,
      clientCount = 20,
      caseCount = 15,
      documentsPerCase = 3,
      financialRecordsPerCase = 2
    } = options;

    this.logger.info('开始生成测试数据集', {
      userCount,
      clientCount,
      caseCount,
      documentsPerCase,
      financialRecordsPerCase
    });

    const users = this.generateUsers(userCount);
    const clients = this.generateClients(clientCount);
    const cases = this.generateCases(caseCount);

    // 为每个案件生成文档
    const documents: DocumentData[] | undefined = undefined;
    cases.forEach(caseData => {
      for (let i = 0; i < documentsPerCase; i++) {
        documents.push(this.generateDocumentData(caseData.id));
      }
    });

    // 为每个案件生成财务记录
    const financial: FinancialData[] | undefined = undefined;
    cases.forEach(caseData => {
      const client = this.randomChoice(clients);
      for (let i = 0; i < financialRecordsPerCase; i++) {
        financial.push(this.generateFinancialData(caseData.id, client.id));
      }
    });

    this.logger.info('测试数据集生成完成', {
      users: users.length,
      clients: clients.length,
      cases: cases.length,
      documents: documents.length,
      financial: financial.length
    });

    return {
      users,
      clients,
      cases,
      documents,
      financial
    };
  }

  /**
   * 导出测试数据为JSON
   */
  exportToJson(data: any, filename: string): void {
    try {
      const jsonData = JSON.stringify(data, null, 2);
      // 在实际实现中，这里会写入文件系统
      this.logger.info('测试数据已导出', { filename, size: jsonData.length });
    } catch (error) {
      this.logger.error('导出测试数据失败', {
        filename,
        error: error instanceof Error ? error.message : error
      });
      throw error;
    }
  }
}