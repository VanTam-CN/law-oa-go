/**
 * 表单验证工具函数
 * 为SmartFieldGroup组件提供验证支持
 */

// 基础验证规则
export interface ValidationRule {
  type: 'required' | 'email' | 'phone' | 'idCard' | 'url' | 'number' | 'min' | 'max' | 'pattern';
  message?: string;
  value?: any;
}

// 验证结果
export interface ValidationResult {
  valid: boolean;
  message?: string;
}

// 验证器映射
const validators: Record<ValidationRule['type'], (value: any, rule: ValidationRule) => ValidationResult> = {
  required: (value: any, rule: ValidationRule): ValidationResult => {
    const isValid = value !== undefined && value !== null && value !== '' &&
                   (typeof value === 'string' ? value.trim() !== '' : true);
    return {
      valid: isValid,
      message: rule.message || '此字段为必填项'
    };
  },

  email: (value: string, rule: ValidationRule): ValidationResult => {
    if (!value) return { valid: true };

    const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    const isValid = emailPattern.test(value);
    return {
      valid: isValid,
      message: rule.message || '请输入有效的邮箱地址'
    };
  },

  phone: (value: string, rule: ValidationRule): ValidationResult => {
    if (!value) return { valid: true };

    const phonePattern = /^1[3-9]\d{9}$/;
    const isValid = phonePattern.test(value.replace(/\D/g, ''));
    return {
      valid: isValid,
      message: rule.message || '请输入有效的手机号码'
    };
  },

  idCard: (value: string, rule: ValidationRule): ValidationResult => {
    if (!value) return { valid: true };

    const idCardPattern = /(^\d{15}$)|(^\d{18}$)|(^\d{17}(\d|X|x)$)/;
    const isValid = idCardPattern.test(value);
    return {
      valid: isValid,
      message: rule.message || '请输入有效的身份证号码'
    };
  },

  url: (value: string, rule: ValidationRule): ValidationResult => {
    if (!value) return { valid: true };

    try {
      new URL(value);
      return { valid: true };
    } catch {
      return {
        valid: false,
        message: rule.message || '请输入有效的URL地址'
      };
    }
  },

  number: (value: any, rule: ValidationRule): ValidationResult => {
    if (value === '' || value === null || value === undefined) return { valid: true };

    const isValid = !isNaN(Number(value));
    return {
      valid: isValid,
      message: rule.message || '请输入有效的数字'
    };
  },

  min: (value: number | string, rule: ValidationRule): ValidationResult => {
    if (value === '' || value === null || value === undefined) return { valid: true };

    const numValue = Number(value);
    const isValid = !isNaN(numValue) && numValue >= Number(rule.value);
    return {
      valid: isValid,
      message: rule.message || `最小值为 ${rule.value}`
    };
  },

  max: (value: number | string, rule: ValidationRule): ValidationResult => {
    if (value === '' || value === null || value === undefined) return { valid: true };

    const numValue = Number(value);
    const isValid = !isNaN(numValue) && numValue <= Number(rule.value);
    return {
      valid: isValid,
      message: rule.message || `最大值为 ${rule.value}`
    };
  },

  pattern: (value: string, rule: ValidationRule): ValidationResult => {
    if (!value) return { valid: true };

    const pattern = new RegExp(rule.value);
    const isValid = pattern.test(value);
    return {
      valid: isValid,
      message: rule.message || '格式不正确'
    };
  }
};

/**
 * 验证单个字段
 * @param value 字段值
 * @param rules 验证规则数组
 * @returns 验证结果
 */
export function validateField(value: any, rules: ValidationRule[]): ValidationResult {
  if (!rules || rules.length === 0) {
    return { valid: true };
  }

  for (const rule of rules) {
    const validator = validators[rule.type];
    if (validator) {
      const result = validator(value, rule);
      if (!result.valid) {
        return result;
      }
    }
  }

  return { valid: true };
}

/**
 * 验证整个表单
 * @param formData 表单数据
 * @param validationRules 验证规则映射
 * @returns 验证结果
 */
export function validateForm(
  formData: Record<string, any>,
  validationRules: Record<string, ValidationRule[]>
): {
  valid: boolean;
  errors: Record<string, string>;
  fields: string[];
} {
  const errors: Record<string, string> = {};
  const invalidFields: string[] = [];

  for (const [fieldName, rules] of Object.entries(validationRules)) {
    const fieldValue = formData[fieldName];
    const result = validateField(fieldValue, rules);

    if (!result.valid) {
      errors[fieldName] = result.message || '验证失败';
      invalidFields.push(fieldName);
    }
  }

  return {
    valid: Object.keys(errors).length === 0,
    errors,
    fields: invalidFields
  };
}

/**
 * 创建Ant Design表单验证规则
 * @param rules 自定义验证规则
 * @returns Ant Design表单验证规则
 */
export function createAntdRules(rules: ValidationRule[]): Array<{
  required?: boolean;
  message?: string;
  pattern?: RegExp;
  min?: number;
  max?: number;
  type?: string;
  validator?: (rule: any, value: any) => Promise<void>;
}> {
  return rules.map(rule => {
    const antdRule: any = {};

    switch (rule.type) {
      case 'required':
        antdRule.required = true;
        antdRule.message = rule.message || '此字段为必填项';
        break;

      case 'email':
        antdRule.type = 'email';
        antdRule.message = rule.message || '请输入有效的邮箱地址';
        break;

      case 'url':
        antdRule.type = 'url';
        antdRule.message = rule.message || '请输入有效的URL地址';
        break;

      case 'number':
        antdRule.type = 'number';
        antdRule.message = rule.message || '请输入有效的数字';
        antdRule.validator = async (_: any, value: any) => {
          const result = validators.number(value, rule);
          if (!result.valid) {
            throw new Error(result.message);
          }
        };
        break;

      case 'min':
        antdRule.min = rule.value;
        antdRule.message = rule.message || `最小值为 ${rule.value}`;
        antdRule.type = 'number';
        break;

      case 'max':
        antdRule.max = rule.value;
        antdRule.message = rule.message || `最大值为 ${rule.value}`;
        antdRule.type = 'number';
        break;

      case 'pattern':
        antdRule.pattern = new RegExp(rule.value);
        antdRule.message = rule.message || '格式不正确';
        break;

      case 'phone':
        antdRule.validator = async (_: any, value: any) => {
          const result = validators.phone(value, rule);
          if (!result.valid) {
            throw new Error(result.message);
          }
        };
        break;

      case 'idCard':
        antdRule.validator = async (_: any, value: any) => {
          const result = validators.idCard(value, rule);
          if (!result.valid) {
            throw new Error(result.message);
          }
        };
        break;

      default:
        break;
    }

    return antdRule;
  });
}

/**
 * 预定义的常用验证规则
 */
export const commonRules = {
  required: { type: 'required' as const, message: '此字段为必填项' },
  email: { type: 'email' as const, message: '请输入有效的邮箱地址' },
  phone: { type: 'phone' as const, message: '请输入有效的手机号码' },
  idCard: { type: 'idCard' as const, message: '请输入有效的身份证号码' },
  url: { type: 'url' as const, message: '请输入有效的URL地址' },
  number: { type: 'number' as const, message: '请输入有效的数字' },
  positiveNumber: { type: 'min' as const, value: 0, message: '请输入正数' },
  age: { type: 'max' as const, value: 150, message: '年龄不能超过150岁' },
  adultAge: { type: 'min' as const, value: 18, message: '年龄不能小于18岁' },
  chineseName: {
    type: 'pattern' as const,
    value: /^[\u4e00-\u9fa5]{2,8}$/,
    message: '请输入2-8个中文字符'
  }
};