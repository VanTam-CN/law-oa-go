/**
 * Mock工厂 - 提供灵活的Mock对象创建
 * 支持动态数据生成和复杂场景模拟
 */

import { mockUser, mockCase, mockClient } from './api-client'

// 通用Mock接口
export interface MockConfig {
  delay?: number
  error?: Error | string
  override?: Record<string, any>
}

// 数据生成器
export class DataFactory {
  // 生成随机字符串
  static randomString(length = 10): string {
    return Math.random()
      .toString(36)
      .substring(2, 2 + length)
  }

  // 生成随机数字
  static randomNumber(min = 0, max = 100): number {
    return Math.floor(Math.random() * (max - min + 1)) + min
  }

  // 生成随机日期
  static randomDate(start = new Date(2020, 0, 1), end = new Date()): Date {
    return new Date(start.getTime() + Math.random() * (end.getTime() - start.getTime()))
  }

  // 生成随机邮箱
  static randomEmail(): string {
    return `${this.randomString(8)}@example.com`
  }

  // 生成随机手机号
  static randomPhone(): string {
    return `1${this.randomNumber(3, 9)}${this.randomNumber(10000000, 99999999)}`
  }

  // 生成随机用户名
  static randomUsername(): string {
    return `user_${this.randomString(8)}`
  }

  // 生成随机中文名字
  static randomChineseName(): string {
    const surnames = ['张', '王', '李', '赵', '刘', '陈', '杨', '黄', '周', '吴']
    const names = ['伟', '芳', '娜', '秀英', '敏', '静', '丽', '强', '磊', '洋']
    return (
      surnames[this.randomNumber(0, surnames.length - 1)] +
      names[this.randomNumber(0, names.length - 1)]
    )
  }
}

// 用户Mock工厂
export class UserFactory {
  static create(override: Partial<typeof mockUser> = {}): typeof mockUser {
    return {
      id: DataFactory.randomNumber(1, 1000),
      username: DataFactory.randomUsername(),
      email: DataFactory.randomEmail(),
      name: DataFactory.randomChineseName(),
      role: 'lawyer',
      permissions: ['case.view', 'case.create', 'client.view'],
      createdAt: DataFactory.randomDate().toISOString(),
      updatedAt: DataFactory.randomDate().toISOString(),
      ...override,
    }
  }

  static createMany(count: number, override: Partial<typeof mockUser> = {}): (typeof mockUser)[] {
    return Array.from({ length: count }, () => this.create(override))
  }

  static createAdmin(override: Partial<typeof mockUser> = {}): typeof mockUser {
    return this.create({
      role: 'admin',
      permissions: ['*'],
      ...override,
    })
  }
}

// 案件Mock工厂
export class CaseFactory {
  static create(override: Partial<typeof mockCase> = {}): typeof mockCase {
    const now = new Date()
    const createdDate = DataFactory.randomDate(new Date(2023, 0, 1), now)

    return {
      id: DataFactory.randomNumber(1, 1000),
      title: `案件${DataFactory.randomString(8)}`,
      description: `案件描述：${DataFactory.randomString(20)}`,
      status: 'active',
      clientId: DataFactory.randomNumber(1, 100),
      lawyerId: DataFactory.randomNumber(1, 50),
      createdAt: createdDate.toISOString(),
      updatedAt: createdDate.toISOString(),
      ...override,
    }
  }

  static createMany(count: number, override: Partial<typeof mockCase> = {}): (typeof mockCase)[] {
    return Array.from({ length: count }, () => this.create(override))
  }

  static createClosed(override: Partial<typeof mockCase> = {}): typeof mockCase {
    return this.create({
      status: 'closed',
      ...override,
    })
  }

  static createPending(override: Partial<typeof mockCase> = {}): typeof mockCase {
    return this.create({
      status: 'pending',
      ...override,
    })
  }
}

// 客户Mock工厂
export class ClientFactory {
  static create(override: Partial<typeof mockClient> = {}): typeof mockClient {
    const now = new Date()
    const createdDate = DataFactory.randomDate(new Date(2023, 0, 1), now)

    return {
      id: DataFactory.randomNumber(1, 1000),
      name: DataFactory.randomChineseName(),
      email: DataFactory.randomEmail(),
      phone: DataFactory.randomPhone(),
      address: `${DataFactory.randomNumber(1, 100)}号测试地址`,
      createdAt: createdDate.toISOString(),
      updatedAt: createdDate.toISOString(),
      ...override,
    }
  }

  static createMany(
    count: number,
    override: Partial<typeof mockClient> = {},
  ): (typeof mockClient)[] {
    return Array.from({ length: count }, () => this.create(override))
  }
}

// API响应Mock工厂
export class ApiResponseFactory {
  static success<T>(data: T, meta?: any) {
    return {
      data,
      error: null,
      meta: {
        timestamp: Date.now(),
        requestId: `req_${DataFactory.randomString(12)}`,
        version: '1.0.0',
        ...meta,
      },
    }
  }

  static error(message: string, code = 'ERROR', status = 400) {
    return {
      data: null,
      error: {
        message,
        code,
        status,
      },
      meta: {
        timestamp: Date.now(),
        requestId: `req_${DataFactory.randomString(12)}`,
        version: '1.0.0',
      },
    }
  }

  static paginated<T>(data: T[], pagination: any = {}) {
    return {
      data,
      pagination: {
        page: 1,
        pageSize: 20,
        total: data.length,
        totalPages: Math.ceil(data.length / 20),
        hasNext: false,
        hasPrev: false,
        ...pagination,
      },
    }
  }
}

// 表单数据Mock工厂
export class FormDataFactory {
  static createCaseForm(override: any = {}) {
    return {
      title: `新案件${DataFactory.randomString(8)}`,
      description: '案件描述内容',
      clientId: DataFactory.randomNumber(1, 100),
      priority: 'normal',
      tags: ['标签1', '标签2'],
      ...override,
    }
  }

  static createClientForm(override: any = {}) {
    return {
      name: DataFactory.randomChineseName(),
      email: DataFactory.randomEmail(),
      phone: DataFactory.randomPhone(),
      address: '详细地址',
      company: DataFactory.randomString(10),
      ...override,
    }
  }

  static createLoginForm(override: any = {}) {
    return {
      username: DataFactory.randomUsername(),
      password: DataFactory.randomString(12),
      remember: false,
      ...override,
    }
  }
}

// Hook Mock工厂
export class HookFactory {
  static createAuthHook(override: any = {}) {
    return {
      user: UserFactory.create(),
      isAuthenticated: true,
      isLoading: false,
      login: jest.fn(),
      logout: jest.fn(),
      register: jest.fn(),
      updateProfile: jest.fn(),
      ...override,
    }
  }

  static createCaseHook(override: any = {}) {
    return {
      cases: CaseFactory.createMany(5),
      isLoading: false,
      error: null,
      createCase: jest.fn(),
      updateCase: jest.fn(),
      deleteCase: jest.fn(),
      fetchCases: jest.fn(),
      fetchCase: jest.fn(),
      ...override,
    }
  }

  static createClientHook(override: any = {}) {
    return {
      clients: ClientFactory.createMany(5),
      isLoading: false,
      error: null,
      createClient: jest.fn(),
      updateClient: jest.fn(),
      deleteClient: jest.fn(),
      fetchClients: jest.fn(),
      fetchClient: jest.fn(),
      ...override,
    }
  }
}

// 事件Mock工厂
export class EventFactory {
  static createClick(override: any = {}) {
    return {
      type: 'click',
      target: { value: '', ...override.target },
      preventDefault: jest.fn(),
      stopPropagation: jest.fn(),
      ...override,
    }
  }

  static createChange(override: any = {}) {
    return {
      target: {
        value: DataFactory.randomString(),
        name: DataFactory.randomString(6),
        ...override.target,
      },
      ...override,
    }
  }

  static createSubmit(override: any = {}) {
    return {
      preventDefault: jest.fn(),
      stopPropagation: jest.fn(),
      ...override,
    }
  }

  static createKeyboard(override: any = {}) {
    return {
      key: DataFactory.randomString(1),
      code: `Key${DataFactory.randomString(1).toUpperCase()}`,
      preventDefault: jest.fn(),
      stopPropagation: jest.fn(),
      ...override,
    }
  }
}

// 组件Props Mock工厂
export class PropsFactory {
  static createButton(override: any = {}) {
    return {
      type: 'button',
      disabled: false,
      loading: false,
      children: '按钮',
      onClick: jest.fn(),
      ...override,
    }
  }

  static createInput(override: any = {}) {
    return {
      type: 'text',
      placeholder: '请输入...',
      value: '',
      onChange: jest.fn(),
      onBlur: jest.fn(),
      onFocus: jest.fn(),
      ...override,
    }
  }

  static createModal(override: any = {}) {
    return {
      visible: true,
      title: '标题',
      onOk: jest.fn(),
      onCancel: jest.fn(),
      footer: null,
      ...override,
    }
  }
}

// 导出所有工厂
export {
  DataFactory as default,
  UserFactory,
  CaseFactory,
  ClientFactory,
  ApiResponseFactory,
  FormDataFactory,
  HookFactory,
  EventFactory,
  PropsFactory,
}
