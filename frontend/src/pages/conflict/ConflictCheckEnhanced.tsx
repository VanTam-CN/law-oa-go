import React, { useState, useEffect } from 'react'
import {
  Card,
  Form,
  Input,
  Button,
  Select,
  Alert,
  Spin,
  Space,
  Divider,
  Row,
  Col,
  Switch,
  Slider,
  Radio,
  message,
  Breadcrumb,
} from 'antd'
import {
  SearchOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  HomeOutlined,
  SafetyCertificateOutlined,
  DownloadOutlined,
  HistoryOutlined,
} from '@ant-design/icons'
import { useNavigate } from 'react-router'
import { performConflictCheck } from '@/services/conflict'
import type { ConflictCheckResponse, ConflictCheckFormData, SearchDepth } from '@/types/conflict'
import ConflictResult from './ConflictResult'
import EntitySelector from './EntitySelector'
import './ConflictCheckEnhanced.less'

const { Option } = Select
const { TextArea } = Input
const { Group: RadioGroup } = Radio

interface ConflictCheckEnhancedFormValues {
  clientId: string
  clientName: string
  clientTaxId?: string
  caseName: string
  caseType: string
  clientType: 'PERSON' | 'COMPANY'
  otherParties: string[]
  lawyerId?: string
  searchYears: number
  searchDepth: SearchDepth
  includeCorporateRelations: boolean
  description?: string
}

const ConflictCheckEnhanced: React.FC = () => {
  const navigate = useNavigate()
  const [form] = Form.useForm<ConflictCheckEnhancedFormValues>()
  const [loading, setLoading] = useState<boolean>(false)
  const [result, setResult] = useState<ConflictCheckResponse | null>(null)
  const [generatingReport, setGeneratingReport] = useState<boolean>(false)

  // 案件类型选项
  const caseTypes = [
    { value: 'civil', label: '民事诉讼' },
    { value: 'commercial', label: '商业诉讼' },
    { value: 'criminal', label: '刑事诉讼' },
    { value: 'administrative', label: '行政诉讼' },
    { value: 'arbitration', label: '仲裁' },
    { value: 'consultation', label: '法律咨询' },
    { value: 'other', label: '其他' },
  ]

  // 搜索深度选项
  const searchDepthOptions = [
    { value: 'BASIC', label: '基础搜索', desc: '仅搜索直接匹配的案件' },
    { value: 'STANDARD', label: '标准搜索', desc: '包含模糊匹配和关联关系' },
    { value: 'DEEP', label: '深度搜索', desc: '包含企业关系和API查询' },
  ]

  // 处理表单提交
  const handleSubmit = async (values: ConflictCheckEnhancedFormValues) => {
    try {
      setLoading(true)
      setResult(null)

      const formData: ConflictCheckFormData = {
        clientId: values.clientId,
        clientName: values.clientName,
        caseName: values.caseName,
        caseType: values.caseType,
        clientType: values.clientType as any,
        opponentInfo: values.otherParties?.join(',') || '',
        lawyerId: values.lawyerId,
        searchYears: values.searchYears,
        searchDepth: values.searchDepth,
        includeCorporateRelations: values.includeCorporateRelations,
      }

      const response = await performConflictCheck(formData)
      setResult(response)

      if (response.data?.hasConflict) {
        message.warning('检测到潜在的利益冲突，请查看详细信息')
      } else {
        message.success('未检测到利益冲突')
      }
    } catch (error) {
      console.error('Conflict check error:', error)
      message.error('冲突检查失败，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  // 重置表单和结果
  const handleReset = () => {
    form.resetFields()
    setResult(null)
  }

  // 生成报告
  const handleGenerateReport = async () => {
    if (!result?.data) return

    try {
      setGeneratingReport(true)
      const report = {
        title: '利益冲突检测报告',
        generatedAt: new Date().toISOString(),
        result: result.data,
      }
      const blob = new Blob([JSON.stringify(report, null, 2)], {
        type: 'application/json;charset=utf-8',
      })
      const url = URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `conflict-report-${result.data.checkId}.json`
      document.body.appendChild(link)
      link.click()
      link.remove()
      URL.revokeObjectURL(url)
      message.success('报告生成成功')
    } catch (error) {
      message.error('报告生成失败')
    } finally {
      setGeneratingReport(false)
    }
  }

  // 查看案件详情
  const handleViewDetails = (caseId: string) => {
    navigate(`/case/${caseId}`)
  }

  return (
    <div className="conflict-check-enhanced-container">
      {/* 页面头部 */}
      <div className="page-header">
        <Breadcrumb>
          <Breadcrumb.Item>
            <HomeOutlined />
            <span>首页</span>
          </Breadcrumb.Item>
          <Breadcrumb.Item>
            <SafetyCertificateOutlined />
            <span>利益冲突审查</span>
          </Breadcrumb.Item>
        </Breadcrumb>
        <h2>利益冲突审查</h2>
        <p className="page-description">
          在接案前进行利益冲突检测，确保合规执业，避免潜在风险
        </p>
      </div>

      <Row gutter={[24, 24]}>
        {/* 左侧：检测表单 */}
        <Col xs={24} lg={10}>
          <Card
            title="检测信息"
            extra={<SafetyCertificateOutlined />}
            className="check-form-card"
          >
            <Spin spinning={false}>
              <Form
                form={form}
                layout="vertical"
                onFinish={handleSubmit}
                initialValues={{
                  clientType: 'COMPANY',
                  caseType: 'commercial',
                  searchYears: 5,
                  searchDepth: 'STANDARD',
                  includeCorporateRelations: true,
                  otherParties: [],
                }}
              >
                {/* 客户信息 */}
                <Divider orientation="left">客户信息</Divider>

                <Form.Item label="客户类型" name="clientType">
                  <RadioGroup>
                    <Radio.Button value="COMPANY">企业客户</Radio.Button>
                    <Radio.Button value="PERSON">个人客户</Radio.Button>
                  </RadioGroup>
                </Form.Item>

                <Form.Item
                  name="clientId"
                  label="客户ID"
                  rules={[{ required: true, message: '请输入客户ID' }]}
                >
                  <Input placeholder="请输入已建档客户ID" />
                </Form.Item>

                <Form.Item
                  name="clientName"
                  label="客户名称"
                  rules={[{ required: true, message: '请输入客户名称' }]}
                >
                  <Input placeholder="请输入客户/公司名称" prefix={<SearchOutlined />} />
                </Form.Item>

                <Form.Item name="clientTaxId" label="客户税号（可选）">
                  <Input placeholder="请输入统一社会信用代码" />
                </Form.Item>

                {/* 案件信息 */}
                <Divider orientation="left">案件信息</Divider>

                <Form.Item
                  name="caseName"
                  label="案件名称"
                  rules={[{ required: true, message: '请输入案件名称' }]}
                >
                  <Input placeholder="请输入案件名称" />
                </Form.Item>

                <Form.Item
                  name="caseType"
                  label="案件类型"
                  rules={[{ required: true, message: '请选择案件类型' }]}
                >
                  <Select placeholder="请选择案件类型">
                    {caseTypes.map((type) => (
                      <Option key={type.value} value={type.value}>
                        {type.label}
                      </Option>
                    ))}
                  </Select>
                </Form.Item>

                <Form.Item
                  name="lawyerId"
                  label="承办律师ID"
                  rules={[{ required: true, message: '请输入承办律师ID' }]}
                >
                  <Input placeholder="请输入承办律师用户ID" />
                </Form.Item>

                {/* 对方当事人 */}
                <Divider orientation="left">对方当事人</Divider>

                <Form.Item
                  name="otherParties"
                  label="对方当事人"
                  tooltip="可以添加多个对方当事人，支持搜索已有实体"
                >
                  <EntitySelector
                    placeholder="请输入对方当事人名称"
                    allowMultiple={true}
                    maxLength={5}
                  />
                </Form.Item>

                {/* 检测配置 */}
                <Divider orientation="left">检测配置</Divider>

                <Form.Item
                  name="searchYears"
                  label="搜索年限"
                  tooltip="搜索过去N年的案件记录"
                >
                  <Slider
                    min={1}
                    max={20}
                    marks={{
                      1: '1年',
                      5: '5年',
                      10: '10年',
                      15: '15年',
                      20: '20年',
                    }}
                  />
                </Form.Item>

                <Form.Item
                  name="searchDepth"
                  label="搜索深度"
                  tooltip="选择搜索深度，深度越高结果越全面但耗时越长"
                >
                  <RadioGroup>
                    <Space direction="vertical">
                      {searchDepthOptions.map((option) => (
                        <div key={option.value}>
                          <Radio value={option.value}>{option.label}</Radio>
                          <div className="search-depth-desc">{option.desc}</div>
                        </div>
                      ))}
                    </Space>
                  </RadioGroup>
                </Form.Item>

                <Form.Item
                  name="includeCorporateRelations"
                  label="包含企业关系"
                  valuePropName="checked"
                >
                  <Switch /> <span style={{ marginLeft: 8 }}>搜索企业关联关系（耗时较长）</span>
                </Form.Item>

                <Form.Item name="description" label="备注说明">
                  <TextArea rows={3} placeholder="请输入备注说明（可选）" />
                </Form.Item>

                {/* 操作按钮 */}
                <Form.Item>
                  <Space style={{ width: '100%' }}>
                    <Button
                      type="primary"
                      htmlType="submit"
                      icon={<SearchOutlined />}
                      loading={loading}
                      size="large"
                      block
                    >
                      开始检测
                    </Button>
                    <Button onClick={handleReset} size="large">
                      重置
                    </Button>
                  </Space>
                </Form.Item>
              </Form>
            </Spin>
          </Card>
        </Col>

        {/* 右侧：检测结果 */}
        <Col xs={24} lg={14}>
          {result?.data ? (
            <ConflictResult
              result={result.data}
              onGenerateReport={handleGenerateReport}
              onViewDetails={handleViewDetails}
              loading={generatingReport}
            />
          ) : (
            <Card className="initial-state-card">
              <div style={{ textAlign: 'center', padding: '60px 20px', color: '#999' }}>
                <SafetyCertificateOutlined style={{ fontSize: 64, marginBottom: 16, opacity: 0.5 }} />
                <div style={{ fontSize: 16, marginBottom: 8 }}>请填写检测信息并点击"开始检测"</div>
                <div style={{ fontSize: 14 }}>检测结果将在此处显示</div>
              </div>
            </Card>
          )}

          {/* 快捷操作 */}
          <Card
            title="快捷操作"
            style={{ marginTop: 16 }}
            className="quick-actions-card"
          >
            <Space>
              <Button icon={<HistoryOutlined />}>检测历史</Button>
              <Button icon={<DownloadOutlined />}>导出配置</Button>
              <Button onClick={() => navigate('/conflict')}>返回结果台</Button>
            </Space>
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export default ConflictCheckEnhanced
