/**
 * CompactCaseForm组件的类型定义
 * 支持1080p优化的紧凑案件创建表单
 */

import { ReactNode } from 'react';
import type { FormInstance } from 'antd/es/form';
import type { StepConfig } from './ProgressIndicator.types';
import type { FieldGroupConfig } from './SmartFieldGroup.types';
import type { ConflictCheckParams } from './ConflictCheckInline.types';

// 表单模式
export type FormMode = 'create' | 'edit' | 'view';

// 表单布局模式
export type FormLayoutMode = 'compact' | 'standard' | 'expanded';

// 表单状态
export type FormStatus = 'idle' | 'validating' | 'saving' | 'saved' | 'error';

// 验证模式
export type ValidationMode = 'realtime' | 'onblur' | 'onsubmit';

// 案件信息类型
export interface CaseInfo {
  /** 案件基本信息 */
  basic: {
    /** 案件编号 */
    caseNumber?: string;
    /** 案件标题 */
    title: string;
    /** 案件类型 */
    caseType: string;
    /** 案件状态 */
    caseStatus: string;
    /** 优先级 */
    priority: string;
    /** 案件描述 */
    description?: string;
    /** 标签 */
    tags?: string[];
  };
  /** 客户信息 */
  client: {
    /** 客户ID */
    clientId: string;
    /** 客户名称 */
    clientName: string;
    /** 联系方式 */
    contact?: string;
    /** 地址 */
    address?: string;
  };
  /** 律师信息 */
  lawyer: {
    /** 主办律师ID */
    lawyerId: string;
    /** 主办律师姓名 */
    lawyerName: string;
    /** 助理律师ID */
    assistingLawyerId?: string;
    /** 助理律师姓名 */
    assistingLawyerName?: string;
  };
  /** 案件详情 */
  details: {
    /** 案件金额 */
    amount?: number;
    /** 争议金额 */
    disputeAmount?: number;
    /** 案件来源 */
    source?: string;
    /** 紧急程度 */
    urgency?: string;
    /** 预计完成时间 */
    expectedCompletion?: string;
    /** 备注 */
    remarks?: string;
  };
  /** 冲突检测结果 */
  conflictCheck?: {
    /** 是否已检测 */
    checked: boolean;
    /** 检测结果 */
    result?: any;
    /** 检测时间 */
    checkedAt?: string;
  };
}

// 表单配置
export interface FormConfig {
  /** 表单模式 */
  mode: FormMode;
  /** 布局模式 */
  layoutMode: FormLayoutMode;
  /** 验证模式 */
  validationMode: ValidationMode;
  /** 是否启用自动保存 */
  enableAutoSave: boolean;
  /** 自动保存间隔(秒) */
  autoSaveInterval: number;
  /** 是否启用字段验证 */
  enableFieldValidation: boolean;
  /** 是否启用冲突检测 */
  enableConflictCheck: boolean;
  /** 是否显示进度指示器 */
  showProgressIndicator: boolean;
  /** 是否显示帮助信息 */
  showHelpInfo: boolean;
  /** 是否显示快捷操作 */
  showQuickActions: boolean;
  /** 最大步骤数 */
  maxSteps: number;
}

// 响应式配置
export interface ResponsiveConfig {
  /** 是否启用1080p优化 */
  enable1080pOptimization: boolean;
  /** 1080p断点 */
  p1080Breakpoint: number;
  /** 移动端断点 */
  mobileBreakpoint: number;
  /** 平板端断点 */
  tabletBreakpoint: number;
  /** 是否自动调整布局 */
  autoAdjustLayout: boolean;
}

// 保存配置
export interface SaveConfig {
  /** 保存前验证 */
  validateBeforeSave: boolean;
  /** 保存后重置 */
  resetAfterSave: boolean;
  /** 保存后跳转 */
  redirectAfterSave: boolean;
  /** 跳转路径 */
  redirectPath?: string;
  /** 自定义保存逻辑 */
  customSaveLogic?: (data: CaseInfo) => Promise<void>;
}

// 验证规则
export interface ValidationRule {
  /** 字段名 */
  field: string;
  /** 规则名称 */
  rule: string;
  /** 验证函数 */
  validator?: (value: any, formData: any) => boolean | string;
  /** 错误消息 */
  message: string;
  /** 是否必填 */
  required?: boolean;
  /** 触发条件 */
  trigger?: 'change' | 'blur' | 'submit';
}

// 快捷操作
export interface QuickAction {
  /** 操作key */
  key: string;
  /** 操作标签 */
  label: string;
  /** 操作图标 */
  icon?: ReactNode;
  /** 操作类型 */
  type: 'primary' | 'default' | 'danger' | 'link';
  /** 处理函数 */
  handler: () => void | Promise<void>;
  /** 确认文本 */
  confirmText?: string;
  /** 是否禁用 */
  disabled?: boolean;
  /** 条件显示 */
  visible?: boolean;
}

// 表单步骤配置
export interface FormStepConfig {
  /** 步骤key */
  key: string;
  /** 步骤标题 */
  title: string;
  /** 步骤描述 */
  description?: string;
  /** 步骤图标 */
  icon?: ReactNode;
  /** 包含的字段组 */
  fieldGroups: string[];
  /** 验证规则 */
  validation?: ValidationRule[];
  /** 是否可选 */
  optional?: boolean;
  /** 完成条件 */
  completionCondition?: (formData: any) => boolean;
}

// 表单事件
export interface FormEvents {
  /** 表单初始化 */
  onInit?: (form: FormInstance) => void;
  /** 步骤变化 */
  onStepChange?: (currentStep: string, prevStep: string) => void;
  /** 字段变化 */
  onFieldChange?: (field: string, value: any, formData: any) => void;
  /** 验证状态变化 */
  onValidationChange?: (isValid: boolean, errors: any) => void;
  /** 保存前 */
  onBeforeSave?: (data: CaseInfo) => Promise<boolean>;
  /** 保存后 */
  onAfterSave?: (result: any) => void;
  /** 保存错误 */
  onSaveError?: (error: any) => void;
  /** 取消 */
  onCancel?: () => void;
  /** 重置 */
  onReset?: () => void;
  /** 冲突检测 */
  onConflictCheck?: (params: ConflictCheckParams) => void;
}

// 组件属性
export interface CompactCaseFormProps {
  /** 表单实例 */
  form?: FormInstance;
  /** 初始数据 */
  initialData?: Partial<CaseInfo>;
  /** 表单配置 */
  config?: Partial<FormConfig>;
  /** 响应式配置 */
  responsiveConfig?: Partial<ResponsiveConfig>;
  /** 保存配置 */
  saveConfig?: Partial<SaveConfig>;
  /** 步骤配置 */
  steps: FormStepConfig[];
  /** 字段组配置 */
  fieldGroups: Record<string, FieldGroupConfig>;
  /** 验证规则 */
  validationRules?: ValidationRule[];
  /** 快捷操作 */
  quickActions?: QuickAction[];
  /** 事件处理 */
  events?: FormEvents;
  /** 自定义样式 */
  className?: string;
  style?: React.CSSProperties;
  /** 是否只读 */
  readonly?: boolean;
  /** 是否禁用 */
  disabled?: boolean;
  /** 加载状态 */
  loading?: boolean;
}

// 表单状态
export interface FormState {
  /** 当前步骤 */
  currentStep: string;
  /** 表单数据 */
  formData: Partial<CaseInfo>;
  /** 验证状态 */
  validationState: {
    isValid: boolean;
    errors: Record<string, string>;
    touched: Record<string, boolean>;
  };
  /** 保存状态 */
  saveState: {
    status: FormStatus;
    isSaving: boolean;
    lastSaved?: Date;
    error?: string;
  };
  /** 步骤状态 */
  stepStates: Record<string, {
    isValid: boolean;
    isCompleted: boolean;
    error?: string;
  }>;
  /** 冲突检测状态 */
  conflictState: {
    isChecked: boolean;
    hasConflict: boolean;
    result?: any;
  };
}

// 统计信息
export interface FormStats {
  /** 总字段数 */
  totalFields: number;
  /** 已填写字段数 */
  filledFields: number;
  /** 已验证字段数 */
  validatedFields: number;
  /** 错误字段数 */
  errorFields: number;
  /** 完成百分比 */
  completionPercentage: number;
  /** 当前步骤 */
  currentStepIndex: number;
  /** 总步骤数 */
  totalSteps: number;
}

// 字段配置
export interface FieldConfig {
  /** 字段名 */
  name: string;
  /** 字段标签 */
  label: string;
  /** 字段类型 */
  type: 'input' | 'textarea' | 'select' | 'date' | 'number' | 'upload' | 'custom';
  /** 是否必填 */
  required?: boolean;
  /** 默认值 */
  defaultValue?: any;
  /** 占位符 */
  placeholder?: string;
  /** 帮助信息 */
  help?: string;
  /** 验证规则 */
  rules?: ValidationRule[];
  /** 依赖字段 */
  dependencies?: string[];
  /** 条件显示 */
  condition?: (formData: any) => boolean;
  /** 自定义组件 */
  component?: ReactNode;
  /** 组件属性 */
  componentProps?: Record<string, any>;
}

// 默认配置
export const DEFAULT_FORM_CONFIG: FormConfig = {
  mode: 'create',
  layoutMode: 'compact',
  validationMode: 'realtime',
  enableAutoSave: true,
  autoSaveInterval: 30,
  enableFieldValidation: true,
  enableConflictCheck: true,
  showProgressIndicator: true,
  showHelpInfo: true,
  showQuickActions: true,
  maxSteps: 5
};

export const DEFAULT_RESPONSIVE_CONFIG: ResponsiveConfig = {
  enable1080pOptimization: true,
  p1080Breakpoint: 1920,
  mobileBreakpoint: 768,
  tabletBreakpoint: 992,
  autoAdjustLayout: true
};

export const DEFAULT_SAVE_CONFIG: SaveConfig = {
  validateBeforeSave: true,
  resetAfterSave: false,
  redirectAfterSave: true,
  redirectPath: '/cases'
};

// 步骤映射
export const STEP_KEYS = {
  BASIC: 'basic',
  CLIENT: 'client',
  LAWYER: 'lawyer',
  DETAILS: 'details',
  REVIEW: 'review'
} as const;

// 字段组映射
export const FIELD_GROUP_KEYS = {
  BASIC_INFO: 'basic_info',
  CASE_DETAILS: 'case_details',
  CLIENT_INFO: 'client_info',
  LAWYER_INFO: 'lawyer_info',
  FINANCIAL_INFO: 'financial_info',
  ADDITIONAL_INFO: 'additional_info'
} as const;