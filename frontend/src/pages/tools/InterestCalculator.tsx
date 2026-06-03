import React, { useState } from 'react'
import { Card, Form, Input, Button, Result, Divider, Radio, Row, Col } from 'antd'
import { CalculatorOutlined } from '@ant-design/icons'
import {
  calculateInterest,
  InterestCalculatorParams,
  InterestCalculatorResult,
} from '@/services/tools'
import './InterestCalculator.less'

const { TextArea } = Input

const InterestCalculator: React.FC = () => {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState<boolean>(false)
  const [result, setResult] = useState<InterestCalculatorResult | null>(null)

  const handleSubmit = async (values: InterestCalculatorParams) => {
    try {
      setLoading(true)
      const response = await calculateInterest(values)
      setResult(response)
    } catch (error) {
      console.error('Failed to calculate interest:', error)
    } finally {
      setLoading(false)
    }
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('zh-CN', {
      style: 'currency',
      currency: 'CNY',
    }).format(amount)
  }

  const formatRate = (rate: number) => {
    return `${rate.toFixed(2)}%`
  }

  return (
    <div className='interest-calculator'>
      <Card title='利息计算器' className='calculator-card'>
        <Form
          form={form}
          layout='vertical'
          onFinish={handleSubmit}
          initialValues={{
            principal: 100000,
            rate: 4.35,
            days: 365,
            type: 'simple',
          }}
        >
          <Row gutter={16}>
            <Col xs={24} sm={12}>
              <Form.Item
                name='principal'
                label='本金（元）'
                rules={[
                  { required: true, message: '请输入本金' },
                  { type: 'number', min: 0, message: '本金必须大于0' },
                ]}
              >
                <Input type='number' placeholder='请输入本金金额' addonAfter='元' />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item
                name='rate'
                label='年利率（%）'
                rules={[
                  { required: true, message: '请输入年利率' },
                  { type: 'number', min: 0, message: '利率必须大于0' },
                ]}
              >
                <Input type='number' placeholder='请输入年利率' addonAfter='%' step='0.01' />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col xs={24} sm={12}>
              <Form.Item
                name='days'
                label='天数'
                rules={[
                  { required: true, message: '请输入天数' },
                  { type: 'number', min: 1, message: '天数必须大于0' },
                ]}
              >
                <Input type='number' placeholder='请输入计算天数' addonAfter='天' />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item
                name='type'
                label='计算方式'
                rules={[{ required: true, message: '请选择计算方式' }]}
              >
                <Radio.Group>
                  <Radio value='simple'>单利</Radio>
                  <Radio value='compound'>复利</Radio>
                  <Radio value='penalty'>违约金</Radio>
                </Radio.Group>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item>
            <Button
              type='primary'
              htmlType='submit'
              loading={loading}
              icon={<CalculatorOutlined />}
            >
              计算利息
            </Button>
          </Form.Item>
        </Form>

        {result && (
          <>
            <Divider />
            <Result
              status='success'
              title='计算完成'
              subTitle={
                <div className='result-content'>
                  <Row gutter={16}>
                    <Col xs={24} sm={8}>
                      <p>
                        <strong>本金：</strong>
                      </p>
                      <p className='amount'>{formatCurrency(result.principal)}</p>
                    </Col>
                    <Col xs={24} sm={8}>
                      <p>
                        <strong>利息：</strong>
                      </p>
                      <p className='amount interest'>{formatCurrency(result.interest)}</p>
                    </Col>
                    <Col xs={24} sm={8}>
                      <p>
                        <strong>本息合计：</strong>
                      </p>
                      <p className='amount total'>{formatCurrency(result.total)}</p>
                    </Col>
                  </Row>
                  <div className='calculation-details'>
                    <p>
                      <strong>计算详情：</strong>
                    </p>
                    <p>本金：{formatCurrency(result.principal)}</p>
                    <p>年利率：{formatRate(result.rate)}</p>
                    <p>天数：{result.days}天</p>
                    <p>
                      计算方式：
                      {result.type === 'simple'
                        ? '单利'
                        : result.type === 'compound'
                          ? '复利'
                          : '违约金'}
                    </p>
                  </div>
                </div>
              }
            />
          </>
        )}

        <Divider orientation='left'>计算公式说明</Divider>
        <div className='formula-explanation'>
          <TextArea
            value={`1. 单利计算：利息 = 本金 × 年利率 × 天数 ÷ 365
2. 复利计算：本息合计 = 本金 × (1 + 年利率 ÷ 365)^天数
   利息 = 本息合计 - 本金
3. 违约金计算：违约金 = 本金 × 日利率 × 天数
   （通常日利率按年利率 ÷ 365 计算，或按合同约定）`}
            rows={6}
            readOnly
          />
        </div>
      </Card>
    </div>
  )
}

export default InterestCalculator
