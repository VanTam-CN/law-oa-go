/**
 * 数据验证工具
 * 提供前端数据验证和类型安全检查功能
 */

import { Case, Client, UserProfile } from '../types';

export interface ValidationResult {
  isValid: boolean;
  errors: string[];
  warnings: string[];
}

export interface ValidationRule {
  field: string;
  required?: boolean;
  type?: 'string' | 'number' | 'email' | 'phone' | 'date' | 'url';
  minLength?: number;
  maxLength?: number;
  min?: number;
  max?: number;
  pattern?: RegExp;
  custom?: (value: any) => string | null;
}

export class ValidationUtils {
  /**
   * 验证案件数据
   */
  static validateCase(caseData: Partial<Case>): ValidationResult {
    const errors: string[] = [];
    const warnings: string[] = [];

    // 必填字段验证
    if (!caseData.title || caseData.title.trim().length === 0) {
      errors.push('案件标题不能为空');
    } else if (caseData.title.length < 2) {
      errors.push('案件标题至少需要2个字符');
    } else if (caseData.title.length > 200) {
      errors.push('案件标题不能超过200个字符');
    }

    if (!caseData.case_type) {
      errors.push('案件类型不能为空');
    }

    if (!caseData.status) {
      errors.push('案件状态不能为空');
    }

    if (!caseData.priority) {
      errors.push('案件优先级不能为空');
    }

    // 可选字段验证
    if (caseData.description && caseData.description.length > 5000) {
      warnings.push('案件描述过长，建议控制在5000字符以内');
    }

    if (caseData.case_amount !== undefined) {
      if (typeof caseData.case_amount !== 'number' || caseData.case_amount < 0) {
        errors.push('案件金额必须是非负数');
      } else if (caseData.case_amount > 999999999) {
        warnings.push('案件金额过大，请确认是否正确');
      }
    }

    // 日期验证
    if (caseData.start_date && !this.isValidDate(caseData.start_date)) {
      errors.push('开始日期格式不正确');
    }

    if (caseData.expected_end_date && !this.isValidDate(caseData.expected_end_date)) {
      errors.push('预期结束日期格式不正确');
    }

    if (caseData.start_date && caseData.expected_end_date) {
      const startDate = new Date(caseData.start_date);
      const endDate = new Date(caseData.expected_end_date);
      if (startDate > endDate) {
        errors.push('开始日期不能晚于结束日期');
      }
    }

    // ID验证
    if (caseData.client_id !== undefined && (!caseData.client_id || caseData.client_id <= 0)) {
      errors.push('客户ID无效');
    }

    if (caseData.lawyer_id !== undefined && caseData.lawyer_id && caseData.lawyer_id <= 0) {
      errors.push('律师ID无效');
    }

    return {
      isValid: errors.length === 0,
      errors,
      warnings
    };
  }

  /**
   * 验证客户数据
   */
  static validateClient(clientData: Partial<Client>): ValidationResult {
    const errors: string[] = [];
    const warnings: string[] = [];

    // 必填字段验证
    if (!clientData.name || clientData.name.trim().length === 0) {
      errors.push('客户姓名不能为空');
    } else if (clientData.name.length < 2) {
      errors.push('客户姓名至少需要2个字符');
    } else if (clientData.name.length > 100) {
      errors.push('客户姓名不能超过100个字符');
    }

    // 联系信息验证
    if (clientData.email && !this.isValidEmail(clientData.email)) {
      errors.push('邮箱格式不正确');
    }

    if (clientData.phone && !this.isValidPhone(clientData.phone)) {
      errors.push('电话号码格式不正确');
    }

    // 可选字段验证
    if (clientData.address && clientData.address.length > 500) {
      warnings.push('地址过长，建议控制在500字符以内');
    }

    if (clientData.company && clientData.company.length > 100) {
      warnings.push('公司名称过长，建议控制在100字符以内');
    }

    return {
      isValid: errors.length === 0,
      errors,
      warnings
    };
  }

  /**
   * 验证用户数据
   */
  static validateUser(userData: Partial<UserProfile>): ValidationResult {
    const errors: string[] = [];
    const warnings: string[] = [];

    // 必填字段验证
    if (!userData.name || userData.name.trim().length === 0) {
      errors.push('用户姓名不能为空');
    } else if (userData.name.length < 2) {
      errors.push('用户姓名至少需要2个字符');
    } else if (userData.name.length > 50) {
      errors.push('用户姓名不能超过50个字符');
    }

    if (!userData.email) {
      errors.push('邮箱不能为空');
    } else if (!this.isValidEmail(userData.email)) {
      errors.push('邮箱格式不正确');
    }

    if (!userData.role) {
      errors.push('用户角色不能为空');
    }

    // 可选字段验证
    if (userData.phone && !this.isValidPhone(userData.phone)) {
      errors.push('电话号码格式不正确');
    }

    if (userData.department && userData.department.length > 100) {
      warnings.push('部门名称过长，建议控制在100字符以内');
    }

    return {
      isValid: errors.length === 0,
      errors,
      warnings
    };
  }

  /**
   * 使用规则验证数据
   */
  static validateWithRules(data: any, rules: ValidationRule[]): ValidationResult {
    const errors: string[] = [];
    const warnings: string[] = [];

    for (const rule of rules) {
      const value = data[rule.field];
      const fieldName = this.getFieldDisplayName(rule.field);

      // 必填验证
      if (rule.required && (value === undefined || value === null || value === '')) {
        errors.push(`${fieldName}不能为空`);
        continue;
      }

      // 如果值为空且不是必填，跳过其他验证
      if (value === undefined || value === null || value === '') {
        continue;
      }

      // 类型验证
      if (rule.type) {
        const typeError = this.validateType(value, rule.type, fieldName);
        if (typeError) {
          errors.push(typeError);
        }
      }

      // 长度验证
      if (typeof value === 'string') {
        if (rule.minLength && value.length < rule.minLength) {
          errors.push(`${fieldName}至少需要${rule.minLength}个字符`);
        }
        if (rule.maxLength && value.length > rule.maxLength) {
          errors.push(`${fieldName}不能超过${rule.maxLength}个字符`);
        }
      }

      // 数值范围验证
      if (typeof value === 'number') {
        if (rule.min !== undefined && value < rule.min) {
          errors.push(`${fieldName}不能小于${rule.min}`);
        }
        if (rule.max !== undefined && value > rule.max) {
          errors.push(`${fieldName}不能大于${rule.max}`);
        }
      }

      // 正则表达式验证
      if (rule.pattern && !rule.pattern.test(String(value))) {
        errors.push(`${fieldName}格式不正确`);
      }

      // 自定义验证
      if (rule.custom) {
        const customError = rule.custom(value);
        if (customError) {
          errors.push(customError);
        }
      }
    }

    return {
      isValid: errors.length === 0,
      errors,
      warnings
    };
  }

  /**
   * 验证API响应数据格式
   */
  static validateApiResponse(response: any, expectedType: 'case' | 'client' | 'user' | 'list'): ValidationResult {
    const errors: string[] = [];
    const warnings: string[] = [];

    if (!response || typeof response !== 'object') {
      errors.push('API响应格式不正确：不是有效的对象');
      return { isValid: false, errors, warnings };
    }

    switch (expectedType) {
      case 'list':
        if (!response.data || !Array.isArray(response.data)) {
          errors.push('列表响应缺少data字段或data不是数组');
        }
        if (!response.pagination || typeof response.pagination !== 'object') {
          warnings.push('列表响应缺少pagination字段');
        } else {
          if (typeof response.pagination.page !== 'number') {
            errors.push('pagination.page必须是数字');
          }
          if (typeof response.pagination.page_size !== 'number') {
            errors.push('pagination.page_size必须是数字');
          }
          if (typeof response.pagination.total !== 'number') {
            errors.push('pagination.total必须是数字');
          }
        }
        break;

      case 'case':
        const caseErrors = this.validateCase(response);
        errors.push(...caseErrors.errors);
        warnings.push(...caseErrors.warnings);
        if (!response.id || typeof response.id !== 'number') {
          errors.push('案件响应缺少有效的id字段');
        }
        break;

      case 'client':
        const clientErrors = this.validateClient(response);
        errors.push(...clientErrors.errors);
        warnings.push(...clientErrors.warnings);
        if (!response.id || typeof response.id !== 'number') {
          errors.push('客户响应缺少有效的id字段');
        }
        break;

      case 'user':
        const userErrors = this.validateUser(response);
        errors.push(...userErrors.errors);
        warnings.push(...userErrors.warnings);
        if (!response.id || typeof response.id !== 'number') {
          errors.push('用户响应缺少有效的id字段');
        }
        break;
    }

    return {
      isValid: errors.length === 0,
      errors,
      warnings
    };
  }

  /**
   * 验证邮箱格式
   */
  private static isValidEmail(email: string): boolean {
    const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailPattern.test(email);
  }

  /**
   * 验证电话号码格式
   */
  private static isValidPhone(phone: string): boolean {
    // 支持中国手机号和固话格式
    const phonePattern = /^1[3-9]\d{9}$|^0\d{2,3}-?\d{7,8}$/;
    const cleanPhone = phone.replace(/[\s-]/g, '');
    return phonePattern.test(cleanPhone);
  }

  /**
   * 验证日期格式
   */
  private static isValidDate(dateString: string): boolean {
    const date = new Date(dateString);
    return date instanceof Date && !isNaN(date.getTime());
  }

  /**
   * 验证数据类型
   */
  private static validateType(value: any, type: string, fieldName: string): string | null {
    switch (type) {
      case 'string':
        if (typeof value !== 'string') {
          return `${fieldName}必须是字符串`;
        }
        break;
      case 'number':
        if (typeof value !== 'number' || isNaN(value)) {
          return `${fieldName}必须是数字`;
        }
        break;
      case 'email':
        if (!this.isValidEmail(value)) {
          return `${fieldName}必须是有效的邮箱地址`;
        }
        break;
      case 'phone':
        if (!this.isValidPhone(value)) {
          return `${fieldName}必须是有效的电话号码`;
        }
        break;
      case 'date':
        if (!this.isValidDate(value)) {
          return `${fieldName}必须是有效的日期`;
        }
        break;
      case 'url':
        try {
          new URL(value);
        } catch {
          return `${fieldName}必须是有效的URL`;
        }
        break;
    }
    return null;
  }

  /**
   * 获取字段显示名称
   */
  private static getFieldDisplayName(field: string): string {
    const fieldNames: Record<string, string> = {
      title: '案件标题',
      description: '案件描述',
      case_type: '案件类型',
      priority: '优先级',
      status: '状态',
      client_id: '客户',
      lawyer_id: '律师',
      case_amount: '案件金额',
      start_date: '开始日期',
      expected_end_date: '预期结束日期',
      name: '姓名',
      email: '邮箱',
      phone: '电话',
      address: '地址',
      company: '公司',
      department: '部门',
      role: '角色'
    };
    return fieldNames[field] || field;
  }
}

// 便捷的验证函数
export const validateCase = (caseData: Partial<Case>) => ValidationUtils.validateCase(caseData);
export const validateClient = (clientData: Partial<Client>) => ValidationUtils.validateClient(clientData);
export const validateUser = (userData: Partial<UserProfile>) => ValidationUtils.validateUser(userData);
export const validateApiResponse = (response: any, expectedType: 'case' | 'client' | 'user' | 'list') =>
  ValidationUtils.validateApiResponse(response, expectedType);
export const validateWithRules = (data: any, rules: ValidationRule[]) =>
  ValidationUtils.validateWithRules(data, rules);

export default ValidationUtils;