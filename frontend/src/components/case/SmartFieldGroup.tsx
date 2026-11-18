import React, { useState, useEffect, useMemo, useCallback } from 'react'
import { Collapse, Form, Badge, Typography, Space, Tooltip, Alert } from 'antd'
import {
  DownOutlined,
  RightOutlined,
  InfoCircleOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons'
import type { CollapseProps } from 'antd/es/collapse'
import type { FormInstance } from 'antd/es/form'
import {
  SmartFieldGroupProps,
  FieldGroupConfig,
  FieldConfig,
  FieldCondition,
  GroupState,
} from './types/SmartFieldGroup.types'
import './SmartFieldGroup.less'

const { Text, Paragraph } = Typography

/**
 * SmartFieldGroup - 智能字段分组组件
 * 支持可折叠字段分组、条件显示和1080p优化
 */
const SmartFieldGroup: React.FC<SmartFieldGroupProps> = ({
  groups,
  formData = {},
  isCompact = false,
  showDescription = true,
  enableAutoSort = true,
  onGroupChange,
  defaultActiveGroups = [],
  className = '',
  style,
  collapseProps = {},
}) => {
  const [form] = Form.useForm()
  const [activeGroups, setActiveGroups] = useState<string[]>(defaultActiveGroups)
  const [groupStates, setGroupStates] = useState<Record<string, GroupState>>({})

  // 条件判断函数
  const evaluateCondition = useCallback((fieldValue: any, condition: FieldCondition): boolean => {
    const { operator, value, logic = 'and' } = condition

    let result = false

    switch (operator) {
      case 'equals':
        result = fieldValue === value
        break
      case 'notEquals':
        result = fieldValue !== value
        break
      case 'contains':
        result = String(fieldValue || '')
          .toLowerCase()
          .includes(String(value).toLowerCase())
        break
      case 'notContains':
        result = !String(fieldValue || '')
          .toLowerCase()
          .includes(String(value).toLowerCase())
        break
      case 'greaterThan':
        result = Number(fieldValue) > Number(value)
        break
      case 'lessThan':
        result = Number(fieldValue) < Number(value)
        break
      case 'exists':
        result = fieldValue !== undefined && fieldValue !== null && fieldValue !== ''
        break
      case 'notExists':
        result = fieldValue === undefined || fieldValue === null || fieldValue === ''
        break
      default:
        result = false
    }

    return result
  }, [])

  // 检查字段是否应该显示
  const shouldFieldShow = useCallback(
    (field: FieldConfig, currentFormData: Record<string, any>): boolean => {
      if (!field.condition || field.condition.length === 0) {
        return true
      }

      // 在紧凑模式下检查是否应该隐藏
      if (isCompact && field.hideInCompact) {
        return false
      }

      // 评估所有条件
      const results = field.condition.map((condition) =>
        evaluateCondition(currentFormData[condition.field], condition),
      )

      // 处理逻辑操作符
      if (field.condition.length === 1) {
        return results[0]
      }

      // 简单的逻辑处理（只支持同一种逻辑操作符）
      const firstLogic = field.condition[0]?.logic || 'and'
      return firstLogic === 'and' ? results.every((r) => r) : results.some((r) => r)
    },
    [evaluateCondition, isCompact],
  )

  // 过滤并排序字段
  const processFields = useCallback(
    (fields: FieldConfig[], currentFormData: Record<string, any>): FieldConfig[] => {
      // 过滤可见字段
      const visibleFields = fields.filter((field) => shouldFieldShow(field, currentFormData))

      // 如果启用自动排序，按优先级排序
      if (enableAutoSort) {
        const priorityOrder = { high: 3, medium: 2, low: 1 }
        return visibleFields.sort(
          (a, b) =>
            (priorityOrder[b.priority || 'medium'] || 2) -
            (priorityOrder[a.priority || 'medium'] || 2),
        )
      }

      return visibleFields
    },
    [shouldFieldShow, enableAutoSort],
  )

  // 处理分组数据
  const processedGroups = useMemo(() => {
    // 如果启用自动排序，按优先级排序分组
    const sortedGroups = [...groups]
    if (enableAutoSort) {
      sortedGroups.sort((a, b) => (b.priority || 0) - (a.priority || 0))
    }

    return sortedGroups
      .map((group) => {
        const processedFields = processFields(group.fields, formData)

        // 1080p紧凑模式下的自动折叠逻辑
        let shouldCollapse = false
        if (isCompact && group.autoCollapseInCompact && activeGroups.includes(group.key)) {
          // 在紧凑模式下，如果高优先级字段很少，考虑自动折叠
          const highPriorityFields = processedFields.filter((f) => f.priority === 'high')
          shouldCollapse = highPriorityFields.length <= 1
        }

        return {
          ...group,
          fields: processedFields,
          visibleFieldsCount: processedFields.length,
          totalFieldsCount: group.fields.length,
          shouldCollapse,
        }
      })
      .filter((group) => group.visibleFieldsCount > 0) // 过滤掉没有可见字段的分组
  }, [groups, formData, processFields, enableAutoSort, isCompact, activeGroups])

  // 更新分组状态
  useEffect(() => {
    const newStates: Record<string, GroupState> = {}

    processedGroups.forEach((group) => {
      newStates[group.key] = {
        key: group.key,
        expanded: activeGroups.includes(group.key),
        visibleFields: group.visibleFieldsCount,
        totalFields: group.totalFieldsCount,
        hasError: false, // 这里可以添加错误检查逻辑
      }
    })

    setGroupStates(newStates)
  }, [processedGroups, activeGroups])

  // 处理折叠状态变化
  const handleCollapseChange = useCallback(
    (keys: string | string[]) => {
      const activeKeys = Array.isArray(keys) ? keys : [keys]
      setActiveGroups(activeKeys)

      if (onGroupChange) {
        onGroupChange(activeKeys)
      }
    },
    [onGroupChange],
  )

  // 渲染字段
  const renderField = useCallback((field: FieldConfig) => {
    const { name, label, component, props, required } = field

    return (
      <Form.Item
        key={name}
        name={name}
        label={label}
        required={required}
        {...props}
        className={`smart-field smart-field-${name}`}
      >
        {component}
      </Form.Item>
    )
  }, [])

  // 渲染分组内容
  const renderGroupContent = useCallback(
    (group: FieldGroupConfig) => {
      const { fields, description } = group

      return (
        <div className='smart-group-content'>
          {showDescription && description && (
            <div className='smart-group-description'>
              <Paragraph
                type='secondary'
                ellipsis={{ rows: 2, expandable: true }}
                className='group-description-text'
              >
                {description}
              </Paragraph>
            </div>
          )}

          <div className='smart-fields-container'>{fields.map(renderField)}</div>
        </div>
      )
    },
    [showDescription, renderField],
  )

  // 生成Collapse items
  const collapseItems: CollapseProps['items'] = useMemo(() => {
    return processedGroups.map((group) => {
      const state = groupStates[group.key]
      const isActive = activeGroups.includes(group.key)

      // 状态指示器
      let statusIcon = null
      if (state?.hasError) {
        statusIcon = (
          <Tooltip title='该分组包含错误'>
            <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />
          </Tooltip>
        )
      } else if (state?.visibleFields && state.visibleFields < state?.totalFields) {
        statusIcon = (
          <Tooltip title={`${state.visibleFields}/${state.totalFields} 字段可见`}>
            <InfoCircleOutlined style={{ color: '#1890ff' }} />
          </Tooltip>
        )
      }

      // 标题内容
      const labelContent = (
        <div className='smart-group-header'>
          <Space size='small' className='group-title-content'>
            {group.icon && <span className='group-icon'>{group.icon}</span>}
            <span className='group-title'>{group.title}</span>
            {statusIcon}
            {state?.visibleFields && (
              <Badge
                count={state.visibleFields}
                size='small'
                style={{
                  backgroundColor: isActive ? '#1890ff' : '#d9d9d9',
                  minWidth: '20px',
                  height: '20px',
                  lineHeight: '20px',
                  fontSize: '12px',
                }}
              />
            )}
          </Space>

          {/* 1080p优化指示器 */}
          {isCompact && group.autoCollapseInCompact && (
            <div className='compact-indicator'>
              <Text type='secondary' style={{ fontSize: '12px' }}>
                📺 已优化
              </Text>
            </div>
          )}
        </div>
      )

      return {
        key: group.key,
        label: labelContent,
        children: renderGroupContent(group),
        extra: group.collapsible !== false && (
          <span className='collapse-toggle-icon'>
            {isActive ? <DownOutlined /> : <RightOutlined />}
          </span>
        ),
        className: `smart-group-item ${isActive ? 'active' : ''} ${isCompact ? 'compact' : ''}`,
      }
    })
  }, [processedGroups, groupStates, activeGroups, isCompact, renderGroupContent])

  // 如果没有分组或没有可见字段
  if (processedGroups.length === 0) {
    return (
      <Alert
        message='没有可显示的字段'
        description='当前配置下没有符合条件的字段可显示'
        type='info'
        showIcon
        className='smart-field-group-empty'
      />
    )
  }

  return (
    <div
      className={`smart-field-group ${className} ${isCompact ? 'smart-field-group-compact' : ''}`}
      style={style}
      data-testid="smart-field-group"
    >
      <Collapse
        items={collapseItems}
        activeKey={activeGroups}
        onChange={handleCollapseChange}
        bordered={false}
        size={isCompact ? 'small' : 'middle'}
        className='smart-field-collapse'
        {...collapseProps}
      />

      {/* 1080p优化提示 */}
      {isCompact && (
        <div className='smart-field-optimization-hint'>
          <Alert
            message='表单已为1080p显示器优化'
            description='部分字段已自动折叠或隐藏，点击分组标题可查看更多字段'
            type='info'
            showIcon
            closable
          />
        </div>
      )}
    </div>
  )
}

export default SmartFieldGroup
