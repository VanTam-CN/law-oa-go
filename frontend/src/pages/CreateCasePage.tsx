import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import {
  Card,
  Form,
  Button,
  Spinner,
  Alert,
  Badge,
  Tabs,
  Tab,
  Row,
  Col,
  ProgressBar,
  ListGroup
} from "react-bootstrap";
import {
  FaArrowLeft,
  FaSave,
  FaTimes,
  FaPlus,
  FaUpload,
  FaUser,
  FaUsers,
  FaFile,
  FaExclamationTriangle,
  FaCheckCircle,
  FaGavel,
  FaDollarSign,
  FaCalendar,
  FaCircleInfo
} from "react-icons/fa";
import {
  MagnifyingGlassIcon
} from "@heroicons/react/24/outline";

// 案件表单数据接口
interface CaseFormData {
  // 基本信息
  caseName: string;
  clientName: string;
  client_id: string;
  otherParties: string[];
  caseType: string;
  causeOfAction: string;
  caseDescription: string;

  // 内部管理
  leadLawyer: string;
  assistingLawyers: string[];
  billingMethod: string;
  contractAmount: number;
  estimatedDuration: number;

  // 合规风控
  conflictCheck: string;
  riskTags: string[];
  isHighRisk: boolean;
  approvalRequired: boolean;
}

// 客户接口
interface Client {
  id: string;
  name: string;
  type: 'PERSON' | 'COMPANY';
  phone: string;
  email: string;
}

// 律师接口
interface Lawyer {
  id: string;
  name: string;
  level: 'PARTNER' | 'SENIOR' | 'JUNIOR';
  specialties: string[];
}

const CreateCasePage: React.FC = () => {
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState("basic");
  const [loading, setLoading] = useState(false);
  const [conflictCheckLoading, setConflictCheckLoading] = useState(false);
  const [conflictCheckProgress, setConflictCheckProgress] = useState<{
    step: number;
    stepName: string;
    progress: number;
    details: string[];
    isCompleted: boolean;
  } | null>(null);
  const [conflictCheckResult, setConflictCheckResult] = useState<any>(null);

  const [formData, setFormData] = useState<CaseFormData>({
    caseName: '',
    clientName: '',
    client_id: '',
    otherParties: [],
    caseType: '',
    causeOfAction: '',
    caseDescription: '',
    leadLawyer: '',
    assistingLawyers: [],
    billingMethod: '',
    contractAmount: 0,
    estimatedDuration: 6,
    conflictCheck: '',
    riskTags: [],
    isHighRisk: false,
    approvalRequired: false
  });

  // 模拟数据
  const clients: Client[] = [
    { id: '1', name: '张三', type: 'PERSON', phone: '13800138001', email: 'zhangsan@example.com' },
    { id: '2', name: 'ABC科技有限公司', type: 'COMPANY', phone: '010-12345678', email: 'contact@abc.com' },
    { id: '3', name: '王五', type: 'PERSON', phone: '13800138002', email: 'wangwu@example.com' }
  ];

  const lawyers: Lawyer[] = [
    { id: '1', name: '张律师', level: 'SENIOR', specialties: ['民事诉讼', '合同纠纷'] },
    { id: '2', name: '李律师', level: 'PARTNER', specialties: ['刑事辩护', '知识产权'] },
    { id: '3', name: '王律师', level: 'JUNIOR', specialties: ['劳动争议', '公司法务'] }
  ];

  const caseTypes = [
    {
      value: 'CIVIL',
      label: '民事案件',
      causes: [
        '合同纠纷', '侵权责任', '婚姻家庭', '继承纠纷',
        '物权纠纷', '人格权纠纷', '劳动争议', '负欠纠纷',
        '不当得利纠纷', '无因管理纠纷', '医疗纠纷', '交通事故赔偿纠纷'
      ]
    },
    {
      value: 'COMMERCIAL',
      label: '商事案件',
      causes: [
        '公司纠纷', '金融纠纷', '知识产权', '投资纠纷',
        '证券纠纷', '保险纠纷', '票据纠纷', '信托纠纷',
        '担保物权纠纷', '不正当竞争纠纷', '破产纠纷', '租赁合同纠纷'
      ]
    },
    {
      value: 'CRIMINAL',
      label: '刑事案件',
      causes: [
        '经济犯罪', '职务犯罪', '暴力犯罪', '网络犯罪',
        '毒品犯罪', '财产犯罪', '政务犯罪', '破坏社会主义市场经济秩序罪',
        '侵犯公民人身权利、民主权利罪', '危害公共安全罪'
      ]
    },
    {
      value: 'ADMINISTRATIVE',
      label: '行政案件',
      causes: [
        '行政处罚', '行政许可', '信息公开', '征收补偿',
        '政府采购', '土地使用权', '环保执法', '税务争议',
        '城市管理', '教育行政', '卫生行政', '公安行政'
      ]
    }
  ];

  const riskCategories = [
    { value: 'HIGH_VALUE', label: '重大标的项目', description: '标的额超过1000万元' },
    { value: 'GROUP_CASE', label: '群体性项目', description: '涉及多方当事人' },
    { value: 'SENSITIVE', label: '敏感性项目', description: '涉及政府机关或公共利益' },
    { value: 'MEDIA_ATTENTION', label: '媒体关注项目', description: '可能引起媒体关注' },
    { value: 'CROSS_BORDER', label: '涉外项目', description: '涉及境外当事人或法律' }
  ];

  const billingMethods = [
    { value: 'FIXED', label: '定额收费' },
    { value: 'HOURLY', label: '按时收费' },
    { value: 'CONTINGENCY', label: '风险代理' },
    { value: 'MIXED', label: '混合收费' }
  ];

  // 利益冲突检查
  const performConflictCheck = async () => {
    if (!formData.client_id) {
      alert('请先选择委托人');
      return;
    }

    setConflictCheckLoading(true);
    setConflictCheckResult(null);
    setConflictCheckProgress(null);

    try {
      const selectedClient = clients.find(c => c.id === formData.client_id);

      // 模拟检查步骤
      const checkSteps = [
        {
          step: 1,
          stepName: '检索委托人信息',
          progress: 10,
          details: [
            `正在检索委托人: ${selectedClient?.name || '未知'}`,
            `委托人类型: ${selectedClient?.type === 'COMPANY' ? '企业客户' : '个人客户'}`,
            `联系方式: ${selectedClient?.phone || '未提供'}`
          ]
        },
        {
          step: 2,
          stepName: '检索历史案件',
          progress: 30,
          details: [
            '正在检索该委托人的历史案件...',
            '查询时间范围: 2020年1月 - 至今',
            '检索字段: 委托人姓名、证件号码、企业统一社会信用代码'
          ]
        },
        {
          step: 3,
          stepName: '检索对方当事人',
          progress: 50,
          details: [
            `正在检索对方当事人: ${formData.otherParties.join('、') || '未填写'}`,
            '检索范围: 所有历史案件的当事人信息',
            '匹配方式: 姓名完全匹配 + 模糊匹配'
          ]
        },
        {
          step: 4,
          stepName: '分析潜在冲突',
          progress: 90,
          details: [
            '正在分析潜在利益冲突...',
            '检查同一案件中的对立当事人',
            '检查关联案件中的利益关系',
            '应用冲突检索规则引擎'
          ]
        },
        {
          step: 5,
          stepName: '生成检索报告',
          progress: 100,
          details: [
            '正在生成详细检索报告...',
            '汇总检索结果',
            '标记风险等级'
          ]
        }
      ];

      // 模拟逐步检索过程
      for (let i = 0; i < checkSteps.length; i++) {
        const step = checkSteps[i];
        setConflictCheckProgress({
          ...step,
          isCompleted: i === checkSteps.length - 1
        });

        await new Promise(resolve => setTimeout(resolve, 800));
      }

      // 模拟检查结果
      const mockResult = {
        hasConflict: false,
        checkDetails: {
          totalCasesChecked: 156,
          clientHistoryCases: 3,
          relatedPartiesChecked: formData.otherParties.length,
          corporateRelationsChecked: selectedClient?.type === 'COMPANY' ? 5 : 0,
          timeRange: '2020年1月 - 至今',
          riskAssessment: '低风险'
        },
        conflictCases: [],
        recommendations: [
          '未发现明显利益冲突',
          '建议在案件处理过程中持续监控潜在冲突',
          '定期更新利益冲突检查记录'
        ]
      };

      setConflictCheckResult(mockResult);
      setFormData(prev => ({ ...prev, conflictCheck: 'NO_CONFLICT' }));

    } catch (error) {
      console.error('冲突检查失败:', error);
      alert('冲突检查失败，请重试');
    } finally {
      setConflictCheckLoading(false);
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const handleCheckboxChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, checked } = e.target;
    setFormData(prev => ({ ...prev, [name]: checked }));
  };

  const handleClientChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const client_id = e.target.value;
    const selectedClient = clients.find(c => c.id === client_id);
    setFormData(prev => ({
      ...prev,
      client_id,
      clientName: selectedClient?.name || ''
    }));
  };

  const handleCaseTypeChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const caseType = e.target.value;
    setFormData(prev => ({
      ...prev,
      caseType,
      causeOfAction: '' // 重置案由
    }));
  };

  const handleRiskTagsChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { value, checked } = e.target;
    setFormData(prev => ({
      ...prev,
      riskTags: checked
        ? [...prev.riskTags, value]
        : prev.riskTags.filter(tag => tag !== value)
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // 基本验证
    if (!formData.caseName || !formData.client_id || !formData.caseType) {
      alert('请填写必要信息');
      return;
    }

    if (!formData.conflictCheck) {
      alert('请完成利益冲突检查');
      return;
    }

    setLoading(true);
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 2000));

      alert('案件创建成功！');
      navigate('/cases');
    } catch (error) {
      console.error('创建失败:', error);
      alert('创建失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const renderBasicInfo = () => (
    <Card>
      <Card.Header>
        <FaFile className="me-2" />
        基本信息
      </Card.Header>
      <Card.Body>
        <Row>
          <Col md={12}>
            <Form.Group className="mb-3">
              <Form.Label>案件名称 *</Form.Label>
              <Form.Control
                type="text"
                name="caseName"
                value={formData.caseName}
                onChange={handleInputChange}
                placeholder="例如：张三 Vs 李四合同纠纷案"
                required
              />
              <Form.Text className="text-muted">
                格式建议：委托人 Vs 对方当事人
              </Form.Text>
            </Form.Group>
          </Col>
        </Row>

        <Row>
          <Col md={6}>
            <Form.Group className="mb-3">
              <Form.Label>委托人 *</Form.Label>
              <Form.Select
                name="client_id"
                value={formData.client_id}
                onChange={handleClientChange}
                required
              >
                <option value="">选择委托人</option>
                {clients.map(client => (
                  <option key={client.id} value={client.id}>
                    {client.name} ({client.type === 'COMPANY' ? '企业' : '个人'})
                  </option>
                ))}
              </Form.Select>
            </Form.Group>
          </Col>
          <Col md={6}>
            <Form.Group className="mb-3">
              <Form.Label>案件类型 *</Form.Label>
              <Form.Select
                name="caseType"
                value={formData.caseType}
                onChange={handleCaseTypeChange}
                required
              >
                <option value="">选择案件类型</option>
                {caseTypes.map(type => (
                  <option key={type.value} value={type.value}>
                    {type.label}
                  </option>
                ))}
              </Form.Select>
            </Form.Group>
          </Col>
        </Row>

        <Row>
          <Col md={6}>
            <Form.Group className="mb-3">
              <Form.Label>案由 *</Form.Label>
              <Form.Select
                name="causeOfAction"
                value={formData.causeOfAction}
                onChange={handleInputChange}
                required
              >
                <option value="">选择或输入案由</option>
                {formData.caseType && caseTypes
                  .find(type => type.value === formData.caseType)
                  ?.causes.map(cause => (
                    <option key={cause} value={cause}>{cause}</option>
                  ))}
              </Form.Select>
            </Form.Group>
          </Col>
          <Col md={6}>
            <Form.Group className="mb-3">
              <Form.Label>其他当事人</Form.Label>
              <Form.Control
                type="text"
                name="otherParties"
                value={formData.otherParties.join(', ')}
                onChange={(e) => setFormData(prev => ({
                  ...prev,
                  otherParties: e.target.value.split(',').map(s => s.trim()).filter(Boolean)
                }))}
                placeholder="输入对方当事人，多个用逗号分隔"
              />
            </Form.Group>
          </Col>
        </Row>

        <Row>
          <Col md={12}>
            <Form.Group className="mb-3">
              <Form.Label>案件描述 *</Form.Label>
              <Form.Control
                as="textarea"
                rows={4}
                name="caseDescription"
                value={formData.caseDescription}
                onChange={handleInputChange}
                placeholder="请详细描述案件背景、争议焦点等关键信息..."
                required
              />
            </Form.Group>
          </Col>
        </Row>
      </Card.Body>
    </Card>
  );

  const renderManagementInfo = () => (
    <Card>
      <Card.Header>
        <FaUsers className="me-2" />
        内部管理信息
      </Card.Header>
      <Card.Body>
        <Row>
          <Col md={6}>
            <Form.Group className="mb-3">
              <Form.Label>主办律师 *</Form.Label>
              <Form.Select
                name="leadLawyer"
                value={formData.leadLawyer}
                onChange={handleInputChange}
                required
              >
                <option value="">选择主办律师</option>
                {lawyers.map(lawyer => (
                  <option key={lawyer.id} value={lawyer.id}>
                    {lawyer.name} ({lawyer.level === 'PARTNER' ? '合伙人' : lawyer.level === 'SENIOR' ? '资深律师' : '律师'})
                  </option>
                ))}
              </Form.Select>
            </Form.Group>
          </Col>
          <Col md={6}>
            <Form.Group className="mb-3">
              <Form.Label>收费方式 *</Form.Label>
              <Form.Select
                name="billingMethod"
                value={formData.billingMethod}
                onChange={handleInputChange}
                required
              >
                <option value="">选择收费方式</option>
                {billingMethods.map(method => (
                  <option key={method.value} value={method.value}>
                    {method.label}
                  </option>
                ))}
              </Form.Select>
            </Form.Group>
          </Col>
        </Row>

        <Row>
          <Col md={6}>
            <Form.Group className="mb-3">
              <Form.Label>合同金额 *</Form.Label>
              <Form.Control
                type="number"
                name="contractAmount"
                value={formData.contractAmount}
                onChange={handleInputChange}
                placeholder="请输入金额"
                min="0"
                required
              />
            </Form.Group>
          </Col>
          <Col md={6}>
            <Form.Group className="mb-3">
              <Form.Label>预估工期（月）</Form.Label>
              <Form.Control
                type="number"
                name="estimatedDuration"
                value={formData.estimatedDuration}
                onChange={handleInputChange}
                placeholder="预估工期"
                min="1"
                max="60"
              />
            </Form.Group>
          </Col>
        </Row>
      </Card.Body>
    </Card>
  );

  const renderComplianceInfo = () => (
    <Card>
      <Card.Header>
        <FaExclamationTriangle className="me-2" />
        合规与风险控制
      </Card.Header>
      <Card.Body>
        <Alert variant="warning">
          <FaExclamationTriangle className="me-2" />
          请认真进行利益冲突检查和风险评估，这是律师执业的基本要求。
        </Alert>

        <Form.Group className="mb-4">
          <Form.Label>利益冲突检查 *</Form.Label>
          <div className="d-grid gap-2">
            <Button
              variant="primary"
              onClick={performConflictCheck}
              disabled={conflictCheckLoading || !formData.client_id}
            >
              {conflictCheckLoading ? (
                <>
                  <Spinner as="span" animation="border" size="sm" className="me-2" />
                  正在执行冲突检索...
                </>
              ) : (
                <>
                  <MagnifyingGlassIcon className="w-4 h-4 me-2" />
                  执行冲突检索
                </>
              )}
            </Button>

            {/* 检索进度显示 */}
            {conflictCheckProgress && (
              <Card className="bg-light">
                <Card.Body>
                  <div className="mb-2">
                    <strong className="text-primary">
                      步骤 {conflictCheckProgress.step}/5: {conflictCheckProgress.stepName}
                    </strong>
                    <ProgressBar
                      now={conflictCheckProgress.progress}
                      className="mt-2"
                      variant={conflictCheckProgress.isCompleted ? 'success' : 'primary'}
                    />
                  </div>
                  <div>
                    {conflictCheckProgress.details.map((detail, index) => (
                      <div key={index} className="text-muted small mb-1">
                        • {detail}
                      </div>
                    ))}
                  </div>
                </Card.Body>
              </Card>
            )}

            {/* 检索结果显示 */}
            {conflictCheckResult && (
              <Card className={conflictCheckResult.hasConflict ? 'border-danger' : 'border-success'}>
                <Card.Body>
                  <Alert variant={conflictCheckResult.hasConflict ? 'danger' : 'success'}>
                    <div className="d-flex align-items-center">
                      {conflictCheckResult.hasConflict ? (
                        <FaExclamationTriangle className="me-2" />
                      ) : (
                        <FaCheckCircle className="me-2" />
                      )}
                      <strong>
                        {conflictCheckResult.hasConflict ? '发现潜在冲突' : '无冲突'}
                      </strong>
                    </div>
                  </Alert>

                  <div className="mt-3">
                    <h6>检索统计：</h6>
                    <ListGroup variant="flush">
                      <ListGroup.Item>
                        已检索案件数量：{conflictCheckResult.checkDetails.totalCasesChecked} 件
                      </ListGroup.Item>
                      <ListGroup.Item>
                        委托人历史案件：{conflictCheckResult.checkDetails.clientHistoryCases} 件
                      </ListGroup.Item>
                      <ListGroup.Item>
                        风险评估：{conflictCheckResult.checkDetails.riskAssessment}
                      </ListGroup.Item>
                    </ListGroup>
                  </div>

                  <div className="mt-3">
                    <h6>建议措施：</h6>
                    <ListGroup variant="flush">
                      {conflictCheckResult.recommendations.map((rec: string, index: number) => (
                        <ListGroup.Item key={index}>{rec}</ListGroup.Item>
                      ))}
                    </ListGroup>
                  </div>
                </Card.Body>
              </Card>
            )}
          </div>

          <Form.Check
            type="radio"
            id="conflict-no"
            name="conflictCheck"
            value="NO_CONFLICT"
            label="无冲突（已完成检索）"
            checked={formData.conflictCheck === 'NO_CONFLICT'}
            onChange={handleInputChange}
            className="mb-2"
          />
          <Form.Check
            type="radio"
            id="conflict-resolved"
            name="conflictCheck"
            value="CONFLICT_RESOLVED"
            label="有冲突但已解决"
            checked={formData.conflictCheck === 'CONFLICT_RESOLVED'}
            onChange={handleInputChange}
            className="mb-2"
          />
          <Form.Check
            type="radio"
            id="conflict-manual"
            name="conflictCheck"
            value="MANUAL_CHECK"
            label="手动检索确认"
            checked={formData.conflictCheck === 'MANUAL_CHECK'}
            onChange={handleInputChange}
          />
        </Form.Group>

        <Form.Group className="mb-3">
          <Form.Label>风险标签</Form.Label>
          {riskCategories.map(category => (
            <Form.Check
              key={category.value}
              type="checkbox"
              id={`risk-${category.value}`}
              value={category.value}
              label={
                <div>
                  <strong>{category.label}</strong>
                  <br />
                  <small className="text-muted">{category.description}</small>
                </div>
              }
              checked={formData.riskTags.includes(category.value)}
              onChange={handleRiskTagsChange}
              className="mb-2"
            />
          ))}
        </Form.Group>

        <Form.Check
          type="checkbox"
          id="high-risk"
          name="isHighRisk"
          label={
            <>
              <FaExclamationTriangle className="text-danger me-2" />
              标记为重大风险项目（需要合伙人审批）
            </>
          }
          checked={formData.isHighRisk}
          onChange={handleCheckboxChange}
        />
      </Card.Body>
    </Card>
  );

  return (
    <div>
      {/* 头部 */}
      <div className="d-flex justify-content-between align-items-center mb-4">
        <div>
          <Button variant="outline-secondary" onClick={() => navigate('/cases')} className="mb-2">
            <FaArrowLeft className="me-2" />
            返回案件列表
          </Button>
          <h1>新建立案</h1>
        </div>
      </div>

      <Form onSubmit={handleSubmit}>
        {/* 标签页 */}
        <Tabs activeKey={activeTab} onSelect={(k) => setActiveTab(k || "basic")} className="mb-4">
          <Tab eventKey="basic" title="基本信息">
            {renderBasicInfo()}
          </Tab>
          <Tab eventKey="management" title="内部管理">
            {renderManagementInfo()}
          </Tab>
          <Tab eventKey="compliance" title="合规风控">
            {renderComplianceInfo()}
          </Tab>
        </Tabs>

        {/* 操作按钮 */}
        <div className="d-flex justify-content-between">
          <Button variant="secondary" onClick={() => navigate('/cases')}>
            <FaTimes className="me-2" />
            取消
          </Button>
          <Button variant="primary" type="submit" disabled={loading}>
            {loading ? (
              <>
                <Spinner as="span" animation="border" size="sm" className="me-2" />
                创建中...
              </>
            ) : (
              <>
                <FaSave className="me-2" />
                创建案件
              </>
            )}
          </Button>
        </div>
      </Form>
    </div>
  );
};

export default CreateCasePage;