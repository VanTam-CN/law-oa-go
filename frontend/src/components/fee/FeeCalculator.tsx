import React, { useState, useEffect, useCallback } from 'react'
import {
  Card,
  Select,
  InputNumber,
  Button,
  Descriptions,
  Space,
  Spin,
  Empty,
  Alert,
  Divider,
  Statistic,
  Row,
  Col,
  Tag,
} from 'antd'
import {
  CalculatorOutlined,
  DollarOutlined,
} from '@ant-design/icons'
import {
  getFeeTemplatesByCaseType,
  calculateEstimatedFee,
  formatAmount,
} from '@/services/finance'
import './FeeCalculator.less'

interface FeeTemplate {
  id: number
  name: string
  case_type: string
  billing_type: string
  performance_bonus_rate: number
  min_amount: number
  max_amount: number
  cost_rate: number
  active: boolean
}

interface FeeCalculationResult {
  template_id: number
  template_name: string
  billing_type: string
  base_fee: number
  performance_bonus: number
  total_fee: number
  cost_deduction: number
  net_fee: number
}

interface FeeCalculatorProps {
  caseType: string
  expectedAmount: number
}

const billingTypeMap: Record<string, string> = {
  hourly: '计时收费',
  fixed: '固定收费',
  hybrid: '混合收费',
  retainer: '顾问费',
}

// 案件类型映射：CaseDetail使用CIVIL/COMMERCIAL等，后端费率模板使用litigation/non_litigation等
const caseTypeToFeeType: Record<string, string> = {
  CIVIL: 'litigation',
  COMMERCIAL: 'litigation',
  CRIMINAL: 'litigation',
  ADMINISTRATIVE: 'litigation',
  non_litigation: 'non_litigation',
  consulting: 'consulting',
}

const feeTypeLabelMap: Record<string, string> = {
  litigation: '诉讼/仲裁案件',
  non_litigation: '非诉案件',
  consulting: '咨询案件',
}

const FeeCalculator: React.FC<FeeCalculatorProps> = ({
  caseType,
  expectedAmount,
}) => {
  const [templates, setTemplates] = useState<FeeTemplate[]>([])
  const [loading, setLoading] = useState(false)
  const [calculating, setCalculating] = useState(false)
  const [selectedTemplateId, setSelectedTemplateId] = useState<number | null>(null)
  const [amount, setAmount] = useState<number>(expectedAmount || 0)
  const [result, setResult] = useState<FeeCalculationResult | null>(null)
  const [error, setError] = useState<string>('')

  // 将CaseDetail的caseType映射为后端费率模板的case_type
  const feeCaseType = caseTypeToFeeType[caseType] || caseType

  // 同步expectedAmount变化
  useEffect(() => {
    if (expectedAmount > 0) {
      setAmount(expectedAmount)
    }
  }, [expectedAmount])

  // 根据案件类型加载费率模板
  useEffect(() => {
    const fetchTemplates = async () => {
      if (!feeCaseType) return
      // 重置之前的选择和结果
      setSelectedTemplateId(null)
      setResult(null)
      setLoading(true)
      setError('')
      try {
        const response = await getFeeTemplatesByCaseType(feeCaseType)
        const data = response.data || []
        setTemplates(data.filter((t: FeeTemplate) => t.active))
      } catch {
        setError('加载费率模板失败')
      } finally {
        setLoading(false)
      }
    }
    fetchTemplates()
  }, [feeCaseType])

  // 金额变化时自动重新计算
  const doCalculate = useCallback(async (templateId: number, calcAmount: number) => {
    if (!templateId || calcAmount <= 0) {
      setResult(null)
      return
    }
    setCalculating(true)
    setError('')
    try {
      const response = await calculateEstimatedFee(templateId, calcAmount)
      setResult(response.data || null)
    } catch {
      setError('费用计算失败，请重试')
      setResult(null)
    } finally {
      setCalculating(false)
    }
  }, [])

  const handleTemplateChange = (templateId: number) => {
    setSelectedTemplateId(templateId)
    setResult(null)
    if (templateId && amount > 0) {
      doCalculate(templateId, amount)
    }
  }

  const handleAmountChange = (value: number | null) => {
    const newAmount = value || 0
    setAmount(newAmount)
    if (selectedTemplateId && newAmount > 0) {
      doCalculate(selectedTemplateId, newAmount)
    }
  }

  const handleCalculate = () => {
    if (selectedTemplateId && amount > 0) {
      doCalculate(selectedTemplateId, amount)
    }
  }

  const selectedTemplate = templates.find((t) => t.id === selectedTemplateId)

  return (
    <div className='fee-calculator'>
      <Card title='费用测算' size='small'>
        {/* 案件类型提示 */}
        <div className='fee-calculator__info'>
          <Tag color='blue'>{feeTypeLabelMap[feeCaseType] || caseType}</Tag>
          <span className='fee-calculator__info-text'>
            可用模板: {loading ? '加载中...' : `${templates.length} 个`}
          </span>
        </div>

        <Divider style={{ margin: '12px 0' }} />

        {/* 输入区域 */}
        <Row gutter={16} align='middle'>
          <Col span={10}>
            <div className='fee-calculator__field'>
              <label className='fee-calculator__label' htmlFor='fee-template-select'>费率模板</label>
              {loading ? (
                <Spin size='small' />
              ) : (
                <Select
                  id='fee-template-select'
                  placeholder='请选择费率模板'
                  value={selectedTemplateId}
                  onChange={handleTemplateChange}
                  style={{ width: '100%' }}
                  options={templates.map((t) => ({
                    value: t.id,
                    label: `${t.name}（${billingTypeMap[t.billing_type] || t.billing_type}）`,
                  }))}
                />
              )}
            </div>
          </Col>
          <Col span={8}>
            <div className='fee-calculator__field'>
              <label className='fee-calculator__label' htmlFor='fee-amount-input'>计算金额（元）</label>
              <InputNumber
                id='fee-amount-input'
                value={amount}
                onChange={handleAmountChange}
                min={0}
                step={1000}
                style={{ width: '100%' }}
                formatter={(value) => `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')}
                parser={(value) => Number(value?.replace(/,/g, '') || 0)}
                addonBefore='¥'
              />
            </div>
          </Col>
          <Col span={6}>
            <Button
              type='primary'
              icon={<CalculatorOutlined />}
              onClick={handleCalculate}
              loading={calculating}
              disabled={!selectedTemplateId || amount <= 0}
              style={{ marginTop: 22 }}
            >
              计算费用
            </Button>
          </Col>
        </Row>

        {/* 错误提示 */}
        {error && (
          <Alert message={error} type='error' showIcon style={{ marginTop: 16 }} closable onClose={() => setError('')} />
        )}

        {/* 模板信息 */}
        {selectedTemplate && (
          <>
            <Divider style={{ margin: '16px 0 12px' }} />
            <Descriptions size='small' column={3} bordered>
              <Descriptions.Item label='计费方式'>
                {billingTypeMap[selectedTemplate.billing_type] || selectedTemplate.billing_type}
              </Descriptions.Item>
              <Descriptions.Item label='绩效奖金率'>
                {selectedTemplate.performance_bonus_rate}%
              </Descriptions.Item>
              <Descriptions.Item label='成本扣除率'>
                {selectedTemplate.cost_rate}%
              </Descriptions.Item>
              <Descriptions.Item label='最低金额'>
                {formatAmount(selectedTemplate.min_amount)}
              </Descriptions.Item>
              <Descriptions.Item label='最高金额'>
                {selectedTemplate.max_amount > 0 ? formatAmount(selectedTemplate.max_amount) : '不限'}
              </Descriptions.Item>
            </Descriptions>
          </>
        )}

        {/* 计算结果 */}
        {result && (
          <>
            <Divider style={{ margin: '16px 0 12px' }} />
            <div className='fee-calculator__result'>
              <Row gutter={16}>
                <Col span={6}>
                  <Statistic
                    title='基础费用'
                    value={result.base_fee}
                    precision={2}
                    prefix={<DollarOutlined />}
                    valueStyle={{ fontSize: 16 }}
                  />
                </Col>
                <Col span={6}>
                  <Statistic
                    title='绩效奖金'
                    value={result.performance_bonus}
                    precision={2}
                    prefix='+'
                    valueStyle={{ fontSize: 16, color: '#52c41a' }}
                  />
                </Col>
                <Col span={6}>
                  <Statistic
                    title='成本扣除'
                    value={result.cost_deduction}
                    precision={2}
                    prefix='-'
                    valueStyle={{ fontSize: 16, color: '#ff4d4f' }}
                  />
                </Col>
                <Col span={6}>
                  <Statistic
                    title='净费用'
                    value={result.net_fee}
                    precision={2}
                    prefix='¥'
                    valueStyle={{ fontSize: 18, color: '#1890ff', fontWeight: 'bold' }}
                  />
                </Col>
              </Row>
              <div className='fee-calculator__result-total'>
                <Space>
                  <span>费用合计（含成本）:</span>
                  <Tag color='blue' style={{ fontSize: 14 }}>{formatAmount(result.total_fee)}</Tag>
                </Space>
              </div>
            </div>
          </>
        )}

        {/* 空状态 */}
        {!loading && templates.length === 0 && (
          <Empty
            description={`暂无${feeTypeLabelMap[feeCaseType] || '该类型'}案件的费率模板`}
            style={{ marginTop: 24 }}
          />
        )}
      </Card>
    </div>
  )
}

export default FeeCalculator
