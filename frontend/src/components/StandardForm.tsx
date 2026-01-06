import React, { useState, useCallback, useMemo } from 'react'
import {
  Form,
  Input,
  Select,
  DatePicker,
  TimePicker,
  Checkbox,
  Radio,
  Switch,
  Button,
  Space,
  Card,
  Row,
  Col,
  Divider,
  Typography,
  message,
  Upload,
  Rate,
  Slider,
  Cascader,
  TreeSelect,
  AutoComplete,
  InputNumber,
  Tooltip,
  Badge,
  Tag,
  Avatar,
  Image,
  Spin,
} from 'antd'
import {
  SaveOutlined,
  ReloadOutlined,
  PlusOutlined,
  DeleteOutlined,
  UploadOutlined,
  EyeOutlined,
  EditOutlined,
  CloseOutlined,
  CheckOutlined,
  ExclamationCircleOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons'
import type { FormInstance, FormProps } from 'antd/es/form'
import type { UploadProps } from 'antd/es/upload'
import type { DatePickerProps, TimePickerProps } from 'antd/es/date-picker'
import type { SelectProps } from 'antd/es/select'
import dayjs from 'dayjs'
import './StandardForm.less'

const { Text, Title, Paragraph } = Typography
const { TextArea, Password, Search } = Input
const { RangePicker } = DatePicker
const { Group: RadioGroup, Button: RadioButton } = Radio
const { Group: CheckboxGroup } = Checkbox
const { OptGroup, Option } = Select

export interface FormField {
  name: string
  label: string
  type:
    | 'text'
    | 'password'
    | 'textarea'
    | 'select'
    | 'multiple'
    | 'date'
    | 'time'
    | 'datetime'
    | 'daterange'
    | 'timerange'
    | 'checkbox'
    | 'radio'
    | 'switch'
    | 'number'
    | 'rate'
    | 'slider'
    | 'cascader'
    | 'tree'
    | 'upload'
    | 'autocomplete'
  required?: boolean
  disabled?: boolean
  readonly?: boolean
  hidden?: boolean
  placeholder?: string
  tooltip?: string
  rules?: any[]
  extra?: string
  dependencies?: string[]
  options?: Array<{ label: string; value: any; disabled?: boolean; children?: any[] }>
  props?: any
  addonBefore?: React.ReactNode
  addonAfter?: React.ReactNode
  prefix?: React.ReactNode
  suffix?: React.ReactNode
  maxLength?: number
  showCount?: boolean
  min?: number
  max?: number
  step?: number
  precision?: number
  format?: string
  showTime?: boolean
  allowClear?: boolean
  mode?: 'multiple' | 'tags' | 'combobox'
  showSearch?: boolean
  showArrow?: boolean
  bordered?: boolean
  size?: 'small' | 'middle' | 'large'
  variant?: 'outlined' | 'filled' | 'borderless'
  initialValue?: any
  normalize?: (value: any, prevValue: any, allValues: any) => any
  transform?: (value: any) => any
  getValueFromEvent?: (event: any) => any
  getValueProps?: (value: any) => any
  valuePropName?: string
  trigger?: string
  validateTrigger?: string | string[]
  validateFirst?: boolean
  noStyle?: boolean
  shouldUpdate?: boolean
  messageVariables?: Record<string, string>
  preserve?: boolean
}

export interface StandardFormProps extends Omit<FormProps, 'form'> {
  fields: FormField[]
  initialValues?: Record<string, any>
  layout?: 'horizontal' | 'vertical' | 'inline'
  labelAlign?: 'left' | 'right'
  labelWrap?: boolean
  labelCol?: any
  wrapperCol?: any
  colon?: boolean
  requiredMark?: boolean | 'optional'
  submitText?: string
  cancelText?: string
  resetText?: string
  showSubmitButton?: boolean
  showCancelButton?: boolean
  showResetButton?: boolean
  loading?: boolean
  readonly?: boolean
  disabled?: boolean
  size?: 'small' | 'middle' | 'large'
  variant?: 'outlined' | 'filled' | 'borderless'
  bordered?: boolean
  showActions?: boolean
  actionPosition?: 'top' | 'bottom' | 'both'
  actionAlign?: 'left' | 'center' | 'right'
  actionType?: 'primary' | 'default' | 'dashed' | 'link' | 'text'
  onSubmit?: (values: any) => void
  onCancel?: () => void
  onReset?: () => void
  onValuesChange?: (changedValues: any, allValues: any) => void
  onFieldsChange?: (changedFields: any, allFields: any) => void
  onFinish?: (values: any) => void
  onFinishFailed?: (errorInfo: any) => void
  extra?: React.ReactNode
  footer?: React.ReactNode
  header?: React.ReactNode
  title?: string
  description?: string
  help?: string
  validateMessages?: any
  scrollToFirstError?: boolean
  preserve?: boolean
  form?: FormInstance
  fieldId?: (name: string) => string
  name?: string
}

const StandardForm: React.FC<StandardFormProps> = ({
  fields,
  initialValues = {},
  layout = 'horizontal',
  labelAlign = 'right',
  labelWrap = false,
  labelCol,
  wrapperCol,
  colon = true,
  requiredMark = true,
  submitText = '提交',
  cancelText = '取消',
  resetText = '重置',
  showSubmitButton = true,
  showCancelButton = true,
  showResetButton = true,
  loading = false,
  readonly = false,
  disabled = false,
  size = 'middle',
  variant = 'outlined',
  bordered = true,
  showActions = true,
  actionPosition = 'bottom',
  actionAlign = 'right',
  actionType = 'primary',
  onSubmit,
  onCancel,
  onReset,
  onValuesChange,
  onFieldsChange,
  onFinish,
  onFinishFailed,
  extra,
  footer,
  header,
  title,
  description,
  help,
  validateMessages,
  scrollToFirstError = true,
  preserve = true,
  form: propForm,
  fieldId,
  name,
  ...restProps
}) => {
  const [form] = Form.useForm(propForm)
  const [submitLoading, setSubmitLoading] = useState(false)
  const [fileList, setFileList] = useState<any[]>([])

  // 初始化表单值
  React.useEffect(() => {
    if (Object.keys(initialValues).length > 0) {
      form.setFieldsValue(initialValues)
    }
  }, [initialValues, form])

  // 处理提交
  const handleSubmit = useCallback(async () => {
    try {
      setSubmitLoading(true)
      const values = await form.validateFields()

      // 处理日期时间格式
      const processedValues = { ...values }
      fields.forEach((field) => {
        if (field.type === 'date' || field.type === 'time' || field.type === 'datetime') {
          if (processedValues[field.name] && dayjs.isDayjs(processedValues[field.name])) {
            processedValues[field.name] = processedValues[field.name].format(
              field.format || 'YYYY-MM-DD',
            )
          }
        } else if (field.type === 'daterange' || field.type === 'timerange') {
          if (processedValues[field.name] && Array.isArray(processedValues[field.name])) {
            processedValues[field.name] = processedValues[field.name].map((item: any) =>
              dayjs.isDayjs(item) ? item.format(field.format || 'YYYY-MM-DD') : item,
            )
          }
        }
      })

      if (onSubmit) {
        await onSubmit(processedValues)
      }
      if (onFinish) {
        await onFinish(processedValues)
      }
      message.success('提交成功')
    } catch (error: any) {
      if (error.errorFields) {
        // 表单验证错误
        if (onFinishFailed) {
          onFinishFailed(error)
        }
      } else {
        // 其他错误
        message.error(error.message || '提交失败')
      }
    } finally {
      setSubmitLoading(false)
    }
  }, [form, fields, onSubmit, onFinish, onFinishFailed])

  // 处理取消
  const handleCancel = useCallback(() => {
    form.resetFields()
    if (onCancel) {
      onCancel()
    }
  }, [form, onCancel])

  // 处理重置
  const handleReset = useCallback(() => {
    form.resetFields()
    if (onReset) {
      onReset()
    }
    message.success('重置成功')
  }, [form, onReset])

  // 渲染表单项
  const renderFormItem = (field: FormField) => {
    if (field.hidden) {
      return null
    }

    const formItemProps: any = {
      name: field.name,
      label: field.label,
      required: field.required,
      disabled: disabled || field.disabled,
      hidden: field.hidden,
      tooltip: field.tooltip,
      extra: field.extra,
      dependencies: field.dependencies,
      rules:
        field.rules ||
        (field.required ? [{ required: true, message: `请输入${field.label}` }] : []),
      ...field.props,
    }

    // 处理不同类型的表单项
    let inputNode: React.ReactNode = null

    switch (field.type) {
      case 'text':
        inputNode = (
          <Input
            placeholder={field.placeholder || `请输入${field.label}`}
            maxLength={field.maxLength}
            showCount={field.showCount}
            addonBefore={field.addonBefore}
            addonAfter={field.addonAfter}
            prefix={field.prefix}
            suffix={field.suffix}
            size={size}
            variant={variant}
            bordered={bordered}
            readOnly={readonly || field.readonly}
          />
        )
        break

      case 'password':
        inputNode = (
          <Password
            placeholder={field.placeholder || `请输入${field.label}`}
            maxLength={field.maxLength}
            showCount={field.showCount}
            addonBefore={field.addonBefore}
            addonAfter={field.addonAfter}
            prefix={field.prefix}
            suffix={field.suffix}
            size={size}
            variant={variant}
            bordered={bordered}
            readOnly={readonly || field.readonly}
          />
        )
        break

      case 'textarea':
        inputNode = (
          <TextArea
            placeholder={field.placeholder || `请输入${field.label}`}
            maxLength={field.maxLength}
            showCount={field.showCount}
            rows={field.props?.rows || 4}
            size={size}
            variant={variant}
            bordered={bordered}
            readOnly={readonly || field.readonly}
          />
        )
        break

      case 'search':
        inputNode = (
          <Search
            placeholder={field.placeholder || `请搜索${field.label}`}
            allowClear={field.allowClear !== false}
            size={size}
            variant={variant}
            bordered={bordered}
            readOnly={readonly || field.readonly}
          />
        )
        break

      case 'select':
      case 'multiple':
        inputNode = (
          <Select
            placeholder={field.placeholder || `请选择${field.label}`}
            options={field.options}
            mode={field.type === 'multiple' ? 'multiple' : undefined}
            allowClear={field.allowClear !== false}
            showSearch={field.showSearch !== false}
            showArrow={field.showArrow !== false}
            size={size}
            variant={variant}
            bordered={bordered}
            disabled={disabled || field.disabled}
            fieldNames={field.props?.fieldNames}
          />
        )
        break

      case 'date':
      case 'datetime':
        inputNode = (
          <DatePicker
            placeholder={field.placeholder || `请选择${field.label}`}
            format={
              field.format || (field.type === 'datetime' ? 'YYYY-MM-DD HH:mm:ss' : 'YYYY-MM-DD')
            }
            showTime={field.type === 'datetime' || field.showTime}
            allowClear={field.allowClear !== false}
            size={size}
            variant={variant}
            bordered={bordered}
            disabled={disabled || field.disabled}
            {...field.props}
          />
        )
        break

      case 'time':
        inputNode = (
          <TimePicker
            placeholder={field.placeholder || `请选择${field.label}`}
            format={field.format || 'HH:mm:ss'}
            allowClear={field.allowClear !== false}
            size={size}
            variant={variant}
            bordered={bordered}
            disabled={disabled || field.disabled}
            {...field.props}
          />
        )
        break

      case 'daterange':
        inputNode = (
          <RangePicker
            placeholder={['开始时间', '结束时间']}
            format={field.format || 'YYYY-MM-DD'}
            showTime={field.showTime}
            allowClear={field.allowClear !== false}
            size={size}
            variant={variant}
            bordered={bordered}
            disabled={disabled || field.disabled}
            {...field.props}
          />
        )
        break

      case 'timerange':
        inputNode = (
          <RangePicker
            placeholder={['开始时间', '结束时间']}
            format={field.format || 'HH:mm:ss'}
            picker='time'
            allowClear={field.allowClear !== false}
            size={size}
            variant={variant}
            bordered={bordered}
            disabled={disabled || field.disabled}
            {...field.props}
          />
        )
        break

      case 'checkbox':
        inputNode = (
          <Checkbox disabled={disabled || field.disabled} {...field.props}>
            {field.extra}
          </Checkbox>
        )
        break

      case 'radio':
        inputNode = (
          <RadioGroup disabled={disabled || field.disabled} size={size} {...field.props}>
            {field.options?.map((option) => (
              <Radio key={option.value} value={option.value} disabled={option.disabled}>
                {option.label}
              </Radio>
            ))}
          </RadioGroup>
        )
        break

      case 'switch':
        inputNode = (
          <Switch
            disabled={disabled || field.disabled}
            size={size === 'small' ? 'small' : 'default'}
            {...field.props}
          />
        )
        break

      case 'number':
        inputNode = (
          <InputNumber
            placeholder={field.placeholder || `请输入${field.label}`}
            min={field.min}
            max={field.max}
            step={field.step}
            precision={field.precision}
            size={size}
            variant={variant}
            bordered={bordered}
            disabled={disabled || field.disabled}
            style={{ width: '100%' }}
            {...field.props}
          />
        )
        break

      case 'rate':
        inputNode = (
          <Rate
            disabled={disabled || field.disabled}
            count={field.props?.count || 5}
            allowHalf={field.props?.allowHalf || false}
            {...field.props}
          />
        )
        break

      case 'slider':
        inputNode = (
          <Slider
            disabled={disabled || field.disabled}
            min={field.min || 0}
            max={field.max || 100}
            step={field.step || 1}
            range={field.props?.range || false}
            {...field.props}
          />
        )
        break

      case 'cascader':
        inputNode = (
          <Cascader
            placeholder={field.placeholder || `请选择${field.label}`}
            options={field.options}
            allowClear={field.allowClear !== false}
            showSearch={field.showSearch !== false}
            size={size}
            variant={variant}
            bordered={bordered}
            disabled={disabled || field.disabled}
            {...field.props}
          />
        )
        break

      case 'tree':
        inputNode = (
          <TreeSelect
            placeholder={field.placeholder || `请选择${field.label}`}
            treeData={field.options}
            allowClear={field.allowClear !== false}
            showSearch={field.showSearch !== false}
            size={size}
            variant={variant}
            bordered={bordered}
            disabled={disabled || field.disabled}
            {...field.props}
          />
        )
        break

      case 'upload':
        const uploadProps: UploadProps = {
          name: 'file',
          action: field.props?.action || '/api/upload',
          headers: field.props?.headers || {},
          onChange: field.props?.onChange,
          beforeUpload: field.props?.beforeUpload,
          customRequest: field.props?.customRequest,
          multiple: field.props?.multiple || false,
          accept: field.props?.accept,
          disabled: disabled || field.disabled,
          ...field.props,
        }
        inputNode = (
          <Upload {...uploadProps}>
            <Button icon={<UploadOutlined />} disabled={disabled || field.disabled}>
              上传文件
            </Button>
          </Upload>
        )
        break

      case 'autocomplete':
        inputNode = (
          <AutoComplete
            placeholder={field.placeholder || `请输入${field.label}`}
            options={field.options}
            allowClear={field.allowClear !== false}
            size={size}
            variant={variant}
            bordered={bordered}
            disabled={disabled || field.disabled}
            {...field.props}
          />
        )
        break

      default:
        inputNode = (
          <Input
            placeholder={field.placeholder || `请输入${field.label}`}
            size={size}
            variant={variant}
            bordered={bordered}
            disabled={disabled || field.disabled}
          />
        )
    }

    return (
      <Form.Item key={field.name} {...formItemProps}>
        {inputNode}
      </Form.Item>
    )
  }

  // 渲染操作按钮
  const renderActions = (position: 'top' | 'bottom') => {
    if (!showActions || (actionPosition !== 'both' && actionPosition !== position)) {
      return null
    }

    return (
      <div className={`form-actions form-actions-${position} form-actions-align-${actionAlign}`}>
        <Space>
          {showResetButton && (
            <Button
              icon={<ReloadOutlined />}
              size={size}
              onClick={handleReset}
              disabled={loading || submitLoading}
            >
              {resetText}
            </Button>
          )}
          {showCancelButton && (
            <Button
              icon={<CloseOutlined />}
              size={size}
              onClick={handleCancel}
              disabled={loading || submitLoading}
            >
              {cancelText}
            </Button>
          )}
          {showSubmitButton && (
            <Button
              type={actionType}
              icon={<SaveOutlined />}
              size={size}
              loading={loading || submitLoading}
              onClick={handleSubmit}
            >
              {submitText}
            </Button>
          )}
        </Space>
      </div>
    )
  }

  // 计算表单布局
  const formLayoutProps = {
    layout,
    labelAlign,
    labelWrap,
    labelCol,
    wrapperCol,
    colon,
    requiredMark,
    size,
    variant,
    bordered,
    scrollToFirstError,
    preserve,
    form,
    fieldId,
    name,
    ...restProps,
  }

  return (
    <Card
      className={`standard-form standard-form-${layout} standard-form-size-${size} ${bordered ? 'standard-form-bordered' : ''}`}
      title={title}
      extra={extra}
      loading={loading}
    >
      {description && (
        <Paragraph type='secondary' className='form-description'>
          {description}
        </Paragraph>
      )}

      {help && (
        <div className='form-help'>
          <InfoCircleOutlined />
          <Text type='secondary'>{help}</Text>
        </div>
      )}

      {header}

      {/* 顶部操作按钮 */}
      {renderActions('top')}

      <Form
        {...formLayoutProps}
        initialValues={initialValues}
        onValuesChange={onValuesChange}
        onFieldsChange={onFieldsChange}
        onFinish={onFinish}
        onFinishFailed={onFinishFailed}
        validateMessages={validateMessages}
      >
        <Row gutter={[16, 16]}>
          {fields.map((field) => (
            <Col
              key={field.name}
              span={field.props?.colSpan || 24}
              xs={field.props?.xs || 24}
              sm={field.props?.sm || 24}
              md={field.props?.md || 12}
              lg={field.props?.lg || 8}
              xl={field.props?.xl || 6}
            >
              {renderFormItem(field)}
            </Col>
          ))}
        </Row>
      </Form>

      {/* 底部操作按钮 */}
      {renderActions('bottom')}

      {footer}
    </Card>
  )
}

export default StandardForm
