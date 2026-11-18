/**
 * CompactCaseForm组件
 * 1080p优化的紧凑案件创建表单主组件
 */

import React, { useState, useEffect, useCallback, useRef, useMemo } from 'react'
import * as FormModule from 'antd/es/form'
import * as ButtonModule from 'antd/es/button'
import * as SpaceModule from 'antd/es/space'
import * as CardModule from 'antd/es/card'
import * as TypographyModule from 'antd/es/typography'
import * as MessageModule from 'antd/es/message'
import * as ModalModule from 'antd/es/modal'
import {
  SaveOutlined,
  ReloadOutlined,
  EyeOutlined,
  PrinterOutlined,
} from '@ant-design/icons'
import ResponsiveFormLayout from './ResponsiveFormLayout'
import SmartFieldGroup from './SmartFieldGroup'
import ProgressIndicator from './ProgressIndicator'
import ConflictCheckInline from './ConflictCheckInline'
import PerformanceOptimizer from './PerformanceOptimizer'
import type {
  CompactCaseFormProps,
  CaseInfo,
  FormConfig,
  FormState,
  FormStats,
  QuickAction,
  FormLayoutMode,
} from './types/CompactCaseForm.types'
import {
  DEFAULT_FORM_CONFIG,
  DEFAULT_RESPONSIVE_CONFIG,
  DEFAULT_SAVE_CONFIG,
  STEP_KEYS,
  FIELD_GROUP_KEYS,
} from './types/CompactCaseForm.types'
import './CompactCaseForm.less'

const resolveComponent = <T,>(module: Record<string, unknown>, exportName: string): T => {
  const defaultExport = module.default
  if (typeof defaultExport === 'function') {
    return defaultExport as T
  }

  return (module[exportName] || defaultExport) as T
}

const resolveObjectExport = <T,>(
  module: Record<string, unknown>,
  expectedKey: string,
): T => {
  const defaultExport = module.default as Record<string, unknown> | undefined
  if (defaultExport && expectedKey in defaultExport) {
    return defaultExport as T
  }

  return module as T
}

const Form = resolveComponent<typeof FormModule.default>(FormModule, 'Form')
const Button = resolveComponent<typeof ButtonModule.default>(ButtonModule, 'Button')
const Space = resolveComponent<typeof SpaceModule.default>(SpaceModule, 'Space')
const Card = resolveComponent<typeof CardModule.default>(CardModule, 'Card')
const Typography = resolveObjectExport<typeof TypographyModule.default>(TypographyModule, 'Title')
const message = resolveObjectExport<typeof MessageModule.default>(MessageModule, 'success')
const Modal = resolveObjectExport<typeof ModalModule.default>(ModalModule, 'confirm')
const { Title, Text } = Typography
const useForm = (Form.useForm || (FormModule as any).useForm) as typeof Form.useForm

const InlineAlert: React.FC<{
  message: string
  description?: string
  type: 'success' | 'error' | 'info'
  style?: React.CSSProperties
}> = ({ message: alertMessage, description, type, style }) => (
  <div
    role='alert'
    className={`compact-form-alert compact-form-alert-${type}`}
    style={style}
  >
    <strong>{alertMessage}</strong>
    {description && <div>{description}</div>}
  </div>
)

const StatusBadge: React.FC<{ status: 'success' | 'error'; text: string }> = ({
  status,
  text,
}) => <span className={`compact-form-badge compact-form-badge-${status}`}>{text}</span>

const SmartFieldGroupRenderer = SmartFieldGroup as React.ComponentType<Record<string, any>>

/**
 * 生成案件编号
 */
const generateCaseNumber = (): string => {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  const random = Math.floor(Math.random() * 1000)
    .toString()
    .padStart(3, '0')
  return `CASE${year}${month}${day}${random}`
}

/**
 * CompactCaseForm主组件
 */
const CompactCaseForm: React.FC<CompactCaseFormProps> = ({
  form: externalForm,
  initialData,
  config = {},
  responsiveConfig = {},
  saveConfig = {},
  steps,
  fieldGroups,
  validationRules = [],
  quickActions = [],
  events = {},
  className = '',
  style,
  readonly = false,
  disabled = false,
  loading = false,
}) => {
  // 合并配置
  const finalConfig = useMemo<FormConfig>(
    () => ({
      ...DEFAULT_FORM_CONFIG,
      ...config,
    }),
    [config],
  )

  const finalResponsiveConfig = useMemo(
    () => ({
      ...DEFAULT_RESPONSIVE_CONFIG,
      ...responsiveConfig,
    }),
    [responsiveConfig],
  )

  const finalSaveConfig = useMemo(
    () => ({
      ...DEFAULT_SAVE_CONFIG,
      ...saveConfig,
    }),
    [saveConfig],
  )

  // 内部表单实例
  const [internalForm] = useForm()
  const form = externalForm || internalForm

  // 表单状态
  const [formState, setFormState] = useState<FormState>(() => ({
    currentStep: steps[0]?.key || STEP_KEYS.BASIC,
    formData: {
      basic: {
        caseNumber: finalConfig.mode === 'create' ? generateCaseNumber() : undefined,
        title: '',
        caseType: '',
        caseStatus: '新建',
        priority: '普通',
        description: '',
        tags: [],
      },
      client: {
        clientId: '',
        clientName: '',
        contact: '',
        address: '',
      },
      lawyer: {
        lawyerId: '',
        lawyerName: '',
        assistingLawyerId: '',
        assistingLawyerName: '',
      },
      details: {
        amount: undefined,
        disputeAmount: undefined,
        source: '',
        urgency: '普通',
        expectedCompletion: '',
        remarks: '',
      },
      conflictCheck: {
        checked: false,
      },
    },
    validationState: {
      isValid: false,
      errors: {},
      touched: {},
    },
    saveState: {
      status: 'idle',
      isSaving: false,
    },
    stepStates: {},
    conflictState: {
      isChecked: false,
      hasConflict: false,
    },
  }))

  // UI状态
  const [isCompact, setIsCompact] = useState(false)
  const [layoutMode, setLayoutMode] = useState<FormLayoutMode>(finalConfig.layoutMode)
  const [currentBreakpoint, setCurrentBreakpoint] = useState<string>('lg')

  // 自动保存定时器
  const autoSaveTimerRef = useRef<NodeJS.Timeout>()

  /**
   * 检测屏幕分辨率和断点
   */
  useEffect(() => {
    const checkResolution = () => {
      const width = window.innerWidth
      let breakpoint = 'lg'
      let compact = false
      let layoutMode: FormLayoutMode = finalConfig.layoutMode

      if (width <= 768) {
        breakpoint = 'xs'
        layoutMode = 'compact'
      } else if (width <= 992) {
        breakpoint = 'md'
        layoutMode = 'compact'
      } else if (width <= 1200) {
        breakpoint = 'lg'
        compact = finalResponsiveConfig.enable1080pOptimization && width <= 1920
        if (compact) {
          layoutMode = 'compact'
        }
      } else {
        breakpoint = 'xl'
        compact = false
      }

      setCurrentBreakpoint(breakpoint)
      setIsCompact(compact)
      setLayoutMode(layoutMode)
    }

    checkResolution()
    window.addEventListener('resize', checkResolution)
    return () => window.removeEventListener('resize', checkResolution)
  }, [finalConfig.layoutMode, finalResponsiveConfig.enable1080pOptimization])

  /**
   * 初始化表单数据
   */
  useEffect(() => {
    if (initialData) {
      const mergedData = {
        ...formState.formData,
        ...initialData,
      }
      setFormState((prev) => ({
        ...prev,
        formData: mergedData,
      }))
      form.setFieldsValue(mergedData)
    }

    events.onInit?.(form)
  }, [initialData, form, events.onInit])

  /**
   * 自动保存
   */
  useEffect(() => {
    if (finalConfig.enableAutoSave && finalConfig.mode !== 'view') {
      if (autoSaveTimerRef.current) {
        clearTimeout(autoSaveTimerRef.current)
      }

      autoSaveTimerRef.current = setTimeout(() => {
        handleAutoSave()
      }, finalConfig.autoSaveInterval * 1000)

      return () => {
        if (autoSaveTimerRef.current) {
          clearTimeout(autoSaveTimerRef.current)
        }
      }
    }
  }, [formState.formData, finalConfig.enableAutoSave, finalConfig.autoSaveInterval])

  /**
   * 计算表单统计信息
   */
  const formStats = useMemo<FormStats>(() => {
    const flatFormData = Object.values(formState.formData).flatMap((section) =>
      section && typeof section === 'object' ? Object.values(section) : [section],
    )
    const totalFields = flatFormData.length
    const filledFields = flatFormData.filter(
      (field) => field !== undefined && field !== null && field !== '',
    ).length

    const errorFields = Object.keys(formState.validationState.errors).length
    const completionPercentage =
      totalFields > 0 ? Math.round((filledFields / totalFields) * 100) : 0

    const currentStepIndex = steps.findIndex((step) => step.key === formState.currentStep)
    const totalSteps = steps.length

    return {
      totalFields,
      filledFields,
      validatedFields: totalFields - errorFields,
      errorFields,
      completionPercentage,
      currentStepIndex: currentStepIndex >= 0 ? currentStepIndex : 0,
      totalSteps,
    }
  }, [formState.formData, formState.validationState.errors, steps, formState.currentStep])

  /**
   * 生成步骤配置
   */
  const stepConfigs = useMemo(() => {
    return steps.map((step) => ({
      key: step.key,
      title: step.title,
      description: step.description,
      status: (formState.stepStates[step.key]?.isCompleted
        ? 'finish'
        : formState.currentStep === step.key
          ? 'process'
          : 'wait') as 'finish' | 'process' | 'wait',
      completed: formState.stepStates[step.key]?.isCompleted || false,
      validation: step.validation
        ? {
            required: true,
            fields: step.validation.map((rule) => rule.field),
            validator: () => Object.keys(formState.validationState.errors).length === 0,
          }
        : undefined,
    }))
  }, [steps, formState.stepStates, formState.currentStep, formState.validationState.errors])

  /**
   * 自动保存处理
   */
  const handleAutoSave = useCallback(async () => {
    if (finalConfig.mode === 'view' || readonly) {
      return
    }

    try {
      // 这里应该调用实际的保存API
      console.log('Auto saving:', formState.formData)
      setFormState((prev) => ({
        ...prev,
        saveState: {
          ...prev.saveState,
          status: 'saved',
          lastSaved: new Date(),
        },
      }))
    } catch (error) {
      console.error('Auto save failed:', error)
    }
  }, [formState.formData, finalConfig.mode, readonly])

  /**
   * 步骤变化处理
   */
  const handleStepChange = useCallback(
    (stepKey: string, stepIndex: number) => {
      const prevStep = formState.currentStep

      // 验证当前步骤
      const currentStepConfig = steps.find((step) => step.key === prevStep)
      if (currentStepConfig && finalConfig.enableFieldValidation) {
        const isValid = validateStep(prevStep)
        if (!isValid) {
          message.warning('请完成当前步骤的必填字段')
          return
        }
      }

      setFormState((prev) => ({
        ...prev,
        currentStep: stepKey,
      }))

      events.onStepChange?.(stepKey, prevStep)
    },
    [formState.currentStep, steps, finalConfig.enableFieldValidation, events.onStepChange],
  )

  /**
   * 字段变化处理
   */
  const handleFieldChange = useCallback(
    (changedFields: any, allFields: any) => {
      const newFormData = {
        ...formState.formData,
        ...changedFields,
      }

      setFormState((prev) => ({
        ...prev,
        formData: newFormData,
      }))

      // 触发字段变化事件
      Object.keys(changedFields).forEach((field) => {
        events.onFieldChange?.(field, changedFields[field], newFormData)
      })
    },
    [formState.formData, events.onFieldChange],
  )

  /**
   * 验证单个步骤
   */
  const validateStep = useCallback(
    (stepKey: string): boolean => {
      const stepConfig = steps.find((step) => step.key === stepKey)
      if (!stepConfig) {
        return true
      }

      const stepFields = stepConfig.fieldGroups.flatMap(
        (groupKey) => fieldGroups[groupKey]?.fields || [],
      )

      let isValid = true
      const errors: Record<string, string> = {}

      stepFields.forEach((field) => {
        const fieldValue = form.getFieldValue(field.name)
        if (field.required && (!fieldValue || fieldValue === '')) {
          errors[field.name] = `${field.label}是必填项`
          isValid = false
        }
      })

      // 更新步骤状态
      setFormState((prev) => ({
        ...prev,
        stepStates: {
          ...prev.stepStates,
          [stepKey]: {
            isValid,
            isCompleted: isValid,
            error: isValid ? undefined : '请完成必填字段',
          },
        },
      }))

      return isValid
    },
    [steps, fieldGroups, form],
  )

  /**
   * 表单验证
   */
  const validateForm = useCallback(async (): Promise<boolean> => {
    try {
      await form.validateFields()
      setFormState((prev) => ({
        ...prev,
        validationState: {
          isValid: true,
          errors: {},
          touched: prev.validationState.touched,
        },
      }))
      return true
    } catch (error) {
      const errors: Record<string, string> = {}
      const validationError = error as { errorFields?: Array<{ name: string; errors: string[] }> }
      if (validationError.errorFields) {
        validationError.errorFields.forEach((field) => {
          errors[field.name] = field.errors[0]
        })
      }

      setFormState((prev) => ({
        ...prev,
        validationState: {
          isValid: false,
          errors,
          touched: prev.validationState.touched,
        },
      }))
      return false
    }
  }, [form])

  /**
   * 保存处理
   */
  const handleSave = useCallback(
    async (isAutoSave = false) => {
      if (readonly || disabled) {
        return
      }

      setFormState((prev) => ({
        ...prev,
        saveState: {
          ...prev.saveState,
          status: 'saving',
          isSaving: true,
        },
      }))

      try {
        // 验证表单
        if (finalSaveConfig.validateBeforeSave) {
          const isValid = await validateForm()
          if (!isValid) {
            message.error('请完成必填字段')
            setFormState((prev) => ({
              ...prev,
              saveState: {
                ...prev.saveState,
                status: 'error',
                isSaving: false,
                error: '表单验证失败',
              },
            }))
            return
          }
        }

        // 保存前事件
        const canSave = await events.onBeforeSave?.(formState.formData as CaseInfo)
        if (canSave === false) {
          return
        }

        // 执行保存逻辑
        if (finalSaveConfig.customSaveLogic) {
          await finalSaveConfig.customSaveLogic(formState.formData as CaseInfo)
        } else {
          // 默认保存逻辑
          console.log('Saving form data:', formState.formData)
        }

        setFormState((prev) => ({
          ...prev,
          saveState: {
            status: 'saved',
            isSaving: false,
            lastSaved: new Date(),
          },
        }))

        if (!isAutoSave) {
          message.success('保存成功')
          events.onAfterSave?.(formState.formData)

          if (finalSaveConfig.resetAfterSave) {
            handleReset()
          }

          if (finalSaveConfig.redirectAfterSave) {
            // 这里应该进行路由跳转
            console.log('Redirecting to:', finalSaveConfig.redirectPath)
          }
        }
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : '保存失败'
        setFormState((prev) => ({
          ...prev,
          saveState: {
            ...prev.saveState,
            status: 'error',
            isSaving: false,
            error: errorMessage,
          },
        }))

        if (!isAutoSave) {
          message.error(errorMessage)
          events.onSaveError?.(error)
        }
      }
    },
    [formState.formData, readonly, disabled, finalSaveConfig, validateForm, events],
  )

  /**
   * 重置处理
   */
  const handleReset = useCallback(() => {
    Modal.confirm({
      title: '确认重置',
      content: '确定要重置表单吗？所有未保存的数据将丢失。',
      onOk: () => {
        form.resetFields()
        setFormState((prev) => ({
          ...prev,
          formData: {
            basic: {
              caseNumber: finalConfig.mode === 'create' ? generateCaseNumber() : undefined,
              title: '',
              caseType: '',
              caseStatus: '新建',
              priority: '普通',
              description: '',
              tags: [],
            },
            client: {
              clientId: '',
              clientName: '',
              contact: '',
              address: '',
            },
            lawyer: {
              lawyerId: '',
              lawyerName: '',
              assistingLawyerId: '',
              assistingLawyerName: '',
            },
            details: {
              amount: undefined,
              disputeAmount: undefined,
              source: '',
              urgency: '普通',
              expectedCompletion: '',
              remarks: '',
            },
            conflictCheck: {
              checked: false,
            },
          },
          validationState: {
            isValid: false,
            errors: {},
            touched: {},
          },
          saveState: {
            status: 'idle',
            isSaving: false,
          },
          stepStates: {},
          conflictState: {
            isChecked: false,
            hasConflict: false,
          },
        }))
        events.onReset?.()
      },
    })
  }, [form, finalConfig.mode, events.onReset])

  /**
   * 取消处理
   */
  const handleCancel = useCallback(() => {
    Modal.confirm({
      title: '确认取消',
      content: '确定要取消吗？所有未保存的数据将丢失。',
      onOk: () => {
        events.onCancel?.()
      },
    })
  }, [events.onCancel])

  /**
   * 冲突检测处理
   */
  const handleConflictCheck = useCallback(
    (params: any) => {
      setFormState((prev) => ({
        ...prev,
        conflictState: {
          isChecked: true,
          hasConflict: params.hasConflict || false,
          result: params.result,
        },
      }))
      events.onConflictCheck?.(params)
    },
    [events.onConflictCheck],
  )

  /**
   * 渲染快捷操作
   */
  const renderQuickActions = () => {
    const defaultActions: QuickAction[] = [
      {
        key: 'save',
        label: '保存草稿',
        icon: <SaveOutlined />,
        type: 'primary',
        handler: () => handleSave(),
        disabled: readonly || disabled || formState.saveState.isSaving,
      },
      {
        key: 'reset',
        label: '重置表单',
        icon: <ReloadOutlined />,
        type: 'default',
        handler: handleReset,
        disabled: readonly || disabled,
      },
      {
        key: 'preview',
        label: '预览',
        icon: <EyeOutlined />,
        type: 'default',
        handler: () => console.log('Preview mode'),
      },
      {
        key: 'print',
        label: '打印',
        icon: <PrinterOutlined />,
        type: 'default',
        handler: () => window.print(),
      },
    ]

    const defaultActionKeys = new Set(defaultActions.map((action) => action.key))
    const actions = [
      ...defaultActions,
      ...quickActions.filter((action) => !defaultActionKeys.has(action.key)),
    ].filter(
      (action) => action.visible !== false,
    )

    if (actions.length === 0) {
      return null
    }

    return (
      <div className='compact-form-actions'>
        <Space size={isCompact ? 'small' : 'middle'}>
          {actions.map((action) => (
            <span key={action.key} title={action.label}>
              <Button
                type={action.type === 'danger' ? 'default' : action.type}
                danger={action.type === 'danger'}
                icon={action.icon}
                onClick={() => action.handler()}
                disabled={action.disabled}
                loading={action.key === 'save' && formState.saveState.isSaving}
                size={isCompact ? 'small' : 'middle'}
              >
                {!isCompact && action.label}
              </Button>
            </span>
          ))}
        </Space>
      </div>
    )
  }

  /**
   * 渲染当前步骤内容
   */
  const renderStepContent = () => {
    const currentStepConfig = steps.find((step) => step.key === formState.currentStep)
    if (!currentStepConfig) {
      return null
    }

    return (
      <div className='step-content'>
        {currentStepConfig.fieldGroups.map((groupKey) => {
          const groupConfig = fieldGroups[groupKey]
          if (!groupConfig) {
            return null
          }

          return (
            <SmartFieldGroupRenderer
              key={groupKey}
              config={groupConfig}
              groups={[groupConfig]}
              form={form}
              formData={formState.formData}
              isCompact={isCompact}
              readonly={readonly}
              disabled={disabled}
              onChange={handleFieldChange}
            />
          )
        })}
      </div>
    )
  }

  /**
   * 渲染冲突检测组件
   */
  const renderConflictCheck = () => {
    if (!finalConfig.enableConflictCheck || finalConfig.mode === 'view') {
      return null
    }

    const conflictParams = {
      clientName: formState.formData.client?.clientName || '',
      opposingParty: '', // 从表单中获取
      caseType: formState.formData.basic?.caseType,
      lawyerId: formState.formData.lawyer?.lawyerId,
    }

    return (
      <div className='conflict-check-section'>
        <ConflictCheckInline
          checkParams={conflictParams}
          result={formState.conflictState.result}
          onCheckComplete={handleConflictCheck}
          displayConfig={{
            isCompact,
            showDetails: true,
            showStats: true,
            showActions: true,
            maxDisplayCount: isCompact ? 2 : 3,
          }}
        />
      </div>
    )
  }

  const containerClassName = [
    'compact-case-form',
    `layout-${layoutMode}`,
    `breakpoint-${currentBreakpoint}`,
    isCompact ? 'compact' : '',
    readonly ? 'readonly' : '',
    disabled ? 'disabled' : '',
    loading ? 'loading' : '',
    className,
  ]
    .filter(Boolean)
    .join(' ')

  return (
    <div className={containerClassName} style={style}>
      {/* 表单头部 */}
      <div className='form-header'>
        <div className='compact-form-row compact-form-row-between'>
          <div>
            <Space>
              <Title level={isCompact ? 4 : 3} style={{ margin: 0 }}>
                {finalConfig.mode === 'create'
                  ? '新建案件'
                  : finalConfig.mode === 'edit'
                    ? '编辑案件'
                    : '查看案件'}
              </Title>
              {formState.formData.basic?.caseNumber && (
                <Text type='secondary'>案件编号: {formState.formData.basic.caseNumber}</Text>
              )}
            </Space>
          </div>
          <div>{renderQuickActions()}</div>
        </div>
      </div>

      {/* 进度指示器 */}
      {finalConfig.showProgressIndicator && (
        <div className='progress-section'>
          <ProgressIndicator
            steps={stepConfigs}
            currentStepKey={formState.currentStep}
            onStepChange={handleStepChange}
            showProgressBar
            showStats
            enableStepNavigation={!readonly}
            isCompact={isCompact}
            direction={isCompact ? 'vertical' : 'horizontal'}
          />
        </div>
      )}

      {/* 表单主体 */}
      <div className='form-body'>
        <ResponsiveFormLayout
          columns={isCompact ? 1 : 2}
          spacing={isCompact ? 'small' : 'medium'}
        >
          <Card
            size={isCompact ? 'small' : 'default'}
            className='form-card'
            title={steps.find((step) => step.key === formState.currentStep)?.title}
            extra={
              <Space>
                {formState.saveState.status === 'saved' && (
                  <StatusBadge status='success' text='已自动保存' />
                )}
                {formState.saveState.status === 'error' && (
                  <StatusBadge status='error' text='保存失败' />
                )}
              </Space>
            }
          >
            {/* 表单字段 */}
            <Form
              form={form}
              layout='vertical'
              initialValues={formState.formData}
              onValuesChange={handleFieldChange}
              disabled={disabled}
              size={isCompact ? 'small' : 'middle'}
            >
              {renderStepContent()}
            </Form>

            {/* 冲突检测 */}
            {renderConflictCheck()}

            {/* 表单状态提示 */}
            {!readonly &&
              formState.validationState.errors &&
              Object.keys(formState.validationState.errors).length > 0 && (
                <InlineAlert
                  message='表单验证错误'
                  description={`还有 ${Object.keys(formState.validationState.errors).length} 个字段需要完善`}
                  type='error'
                  style={{ marginTop: 16 }}
                />
              )}

            {/* 保存状态提示 */}
            {formState.saveState.status === 'saved' && (
              <InlineAlert
                message='自动保存成功'
                description={`最后保存时间: ${formState.saveState.lastSaved?.toLocaleString()}`}
                type='success'
                style={{ marginTop: 16 }}
              />
            )}
          </Card>
        </ResponsiveFormLayout>
      </div>

      {/* 表单底部 */}
      <div className='form-footer'>
        <div className='compact-form-row compact-form-row-between'>
          <div>
            <Text type='secondary' style={{ fontSize: isCompact ? 12 : 14 }}>
              完成度: {formStats.completionPercentage}% | 已填写: {formStats.filledFields}/
              {formStats.totalFields} 字段
            </Text>
          </div>
          <div>
            <Space>
              <Button
                size={isCompact ? 'small' : 'middle'}
                onClick={handleCancel}
                disabled={formState.saveState.isSaving}
              >
                取消
              </Button>
              <Button
                type='primary'
                size={isCompact ? 'small' : 'middle'}
                onClick={() => handleSave()}
                loading={formState.saveState.isSaving}
                disabled={readonly || disabled}
              >
                {formState.saveState.isSaving ? '保存中...' : '保存'}
              </Button>
            </Space>
          </div>
        </div>
      </div>

      {/* 性能优化器 */}
      {process.env.NODE_ENV === 'development' && (
        <div className='performance-optimizer-section' style={{ marginTop: 16 }}>
          <PerformanceOptimizer enableAutoOptimization showMetrics={false} />
        </div>
      )}

      {/* 1080p优化提示 */}
      {isCompact && (
        <div className='optimization-hint'>
          <InlineAlert
            message='已为1080p显示器优化'
            description='表单已自动调整为紧凑布局，提升空间利用率'
            type='info'
            style={{ marginTop: 16 }}
          />
        </div>
      )}
    </div>
  )
}

export default CompactCaseForm
