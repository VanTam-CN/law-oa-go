/**
 * SmartFieldGroup组件的类型定义
 * 支持可折叠字段分组和条件显示
 */

import { ReactNode } from 'react';
import type { FormItemProps } from 'antd/es/form';
import type { CollapseProps } from 'antd/es/collapse';

// 字段条件类型
export interface FieldCondition {
  /** 依赖的字段名 */
  field: string;
  /** 条件操作符 */
  operator: 'equals' | 'notEquals' | 'contains' | 'notContains' | 'greaterThan' | 'lessThan' | 'exists' | 'notExists';
  /** 条件值 */
  value: any;
  /** 逻辑操作符 */
  logic?: 'and' | 'or';
}

// 字段配置
export interface FieldConfig {
  /** 字段名 */
  name: string;
  /** 字段标签 */
  label: string;
  /** 字段组件 */
  component: ReactNode;
  /** 字段属性 */
  props?: FormItemProps;
  /** 条件显示配置 */
  condition?: FieldCondition[];
  /** 是否必填 */
  required?: boolean;
  /** 是否在1080p优化时隐藏 */
  hideInCompact?: boolean;
  /** 字段优先级（用于空间优化时的排序） */
  priority?: 'high' | 'medium' | 'low';
}

// 分组配置
export interface FieldGroupConfig {
  /** 分组唯一标识 */
  key: string;
  /** 分组标题 */
  title: string;
  /** 分组描述 */
  description?: string;
  /** 分组内的字段配置 */
  fields: FieldConfig[];
  /** 是否默认展开 */
  defaultExpanded?: boolean;
  /** 是否可折叠 */
  collapsible?: boolean;
  /** 分组图标 */
  icon?: ReactNode;
  /** 分组优先级（用于自动排序） */
  priority?: number;
  /** 是否在1080p优化时自动折叠 */
  autoCollapseInCompact?: boolean;
}

// SmartFieldGroup组件属性
export interface SmartFieldGroupProps {
  /** 分组配置数组 */
  groups: FieldGroupConfig[];
  /** 当前表单数据（用于条件判断） */
  formData?: Record<string, any>;
  /** 是否启用1080p紧凑模式 */
  isCompact?: boolean;
  /** 是否显示分组描述 */
  showDescription?: boolean;
  /** 是否启用自动排序 */
  enableAutoSort?: boolean;
  /** 折叠面板变化回调 */
  onGroupChange?: (activeKeys: string | string[]) => void;
  /** 默认展开的分组 */
  defaultActiveGroups?: string[];
  /** 样式类名 */
  className?: string;
  /** 自定义样式 */
  style?: React.CSSProperties;
  /** Collapse组件的其他属性 */
  collapseProps?: Omit<CollapseProps, 'items' | 'onChange'>;
}

// 条件判断结果
export interface ConditionResult {
  /** 是否满足条件 */
  matched: boolean;
  /** 匹配的字段 */
  fields: string[];
}

// 分组状态
export interface GroupState {
  /** 分组key */
  key: string;
  /** 是否展开 */
  expanded: boolean;
  /** 可见字段数量 */
  visibleFields: number;
  /** 总字段数量 */
  totalFields: number;
  /** 是否有错误 */
  hasError: boolean;
}