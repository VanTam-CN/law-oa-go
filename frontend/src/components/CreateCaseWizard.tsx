import React, { useState, useEffect } from "react";
import {
  Modal,
  Form,
  Input,
  Select,
  InputNumber,
  DatePicker,
  Button,
  Steps,
  Card,
  Row,
  Col,
  Radio,
  Checkbox,
  ConfigProvider,
  Tag,
  Avatar,
  Space,
  Typography,
  Divider,
  Spin,
  Alert,
  Progress,
  Timeline,
  Descriptions,
  Statistic,
  Badge,
  Tooltip,
  List,
  Result,
  Collapse,
  Table,
  Drawer,
  Rate,
} from "antd";
import {
  FileTextOutlined,
  UserOutlined,
  TeamOutlined,
  SafetyOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  LoadingOutlined,
  WarningOutlined,
  InfoCircleOutlined,
  ClockCircleOutlined,
  SecurityScanOutlined,
  AlertOutlined,
  BulbOutlined,
  BarChartOutlined,
  EyeOutlined,
  HistoryOutlined,
  RocketOutlined,
  StarOutlined,
} from "@ant-design/icons";
import { get } from "@/services/api";
import { clientService } from "@/services/client";
import { conflictAPI } from "@/api/conflict";
import { message } from "@/utils/messageHelper";
import ConflictCheckResult from "./conflict/ConflictCheckResult";

interface CaseInfo {
  caseNo?: string;
  caseName?: string;
  caseType?: string;
  clientId?: number | null;
  lawyerId?: number | null;
  startDate?: string;
  endDate?: string;
  status?: string;
  description?: string;
  opponentInfo?: string;
  causeOfAction?: string;
  conflictCheckStatus?: string;
  isMajorRisk?: boolean;
  isMassCase?: boolean;
  isSensitiveCase?: boolean;
}

interface ClientInfo {
  clientId: number;
  clientName: string;
  phone: string;
  email: string;
  clientType: string;
  company?: string;
  address: string;
}

interface LawyerInfo {
  lawyerId: number;
  lawyerName: string;
  phone: string;
  email: string;
  position: string;
  department: string;
  specialty: string;
}

interface ConflictCheckResult {
  status: string;
  score: number;
  conflicts: any[];
  summary: string;
  checkTime?: string;
  checker?: string;
  totalChecked?: number;
  riskFactors?: any[];
  recommendations?: any[];
  relatedCases?: any[];
  complianceNotes?: string;
}
import dayjs from "dayjs";

const { TextArea } = Input;
const { Option } = Select;
const { Step } = Steps;
const { Panel } = Collapse;
const { Text, Title, Paragraph } = Typography;

interface CreateCaseWizardProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess: () => void;
}

const CreateCaseWizard: React.FC<CreateCaseWizardProps> = ({
  visible,
  onCancel,
  onSuccess,
}) => {
  const [form] = Form.useForm();
  const [currentStep, setCurrentStep] = useState(0);
  const [loading, setLoading] = useState(false);
  const [conflictChecking, setConflictChecking] = useState(false);

  // 数据状态
  const [clients, setClients] = useState<ClientInfo[]>([]);
  const [lawyers, setLawyers] = useState<LawyerInfo[]>([]);
  const [caseTypes, setCaseTypes] = useState<any[]>([]);
  const [projectTypes, setProjectTypes] = useState<any[]>([]);
  const [billingMethods, setBillingMethods] = useState<any[]>([]);

  // 表单数据状态
  const [formData, setFormData] = useState<Partial<CaseInfo>>({});
  const [conflictResult, setConflictResult] =
    useState<ConflictCheckResult | null>(null);
  const [conflictConfirmed, setConflictConfirmed] = useState(false);

  // 步骤定义
  const steps = [
    {
      title: "基本信息",
      description: "填写案件基本信息",
      icon: <FileTextOutlined />,
    },
    {
      title: "当事人信息",
      description: "选择客户和对方信息",
      icon: <UserOutlined />,
    },
    {
      title: "团队分配",
      description: "分配律师团队",
      icon: <TeamOutlined />,
    },
    {
      title: "利益冲突检查",
      description: "检查潜在利益冲突",
      icon: <SafetyOutlined />,
    },
    {
      title: "确认创建",
      description: "确认信息并创建案件",
      icon: <CheckCircleOutlined />,
    },
  ];

  useEffect(() => {
    if (visible) {
      loadInitialData();
      resetWizard();
    }
  }, [visible]);

  const resetWizard = () => {
    setCurrentStep(0);
    setFormData({});
    setConflictResult(null);
    form.resetFields();
  };

  const loadInitialData = async () => {
    console.log('CreateCaseWizard开始加载数据...');
    setLoading(true);

    try {
      // 获取客户数据 - 使用clientService确保与客户管理界面一致
      console.log('获取客户数据...');
      const clientsResponse = await clientService.getClientList({
        pageNum: 1,
        pageSize: 9999
      });
      console.log('客户API响应:', clientsResponse);

      // 客户数据处理 - 基于实际API响应格式
      let clientData = [];
      let totalCount = 0;

      // 🔧 修复：HTTP拦截器处理后的实际格式：{clients: [...], pagination: {total: 14, ...}}
      if (clientsResponse.clients && Array.isArray(clientsResponse.clients)) {
        // HTTP拦截器转换后的标准格式：数据在clients字段中
        clientData = clientsResponse.clients;
        totalCount = clientsResponse.pagination?.total || clientsResponse.total || 0;
        console.log('使用HTTP拦截器标准格式解析，clients数量:', clientData.length);
      } else if (clientsResponse.success && clientsResponse.data) {
        // 备用格式：数据在res.data.clients中
        clientData = clientsResponse.data.clients || [];
        totalCount = clientsResponse.data.pagination?.total || 0;
        console.log('使用备用API格式解析，clients数量:', clientData.length);
      } else if (clientsResponse.data && clientsResponse.data.clients) {
        // 兼容格式：直接从data.clients获取
        clientData = clientsResponse.data.clients;
        totalCount = clientsResponse.data.pagination?.total || 0;
        console.log('使用兼容格式解析，clients数量:', clientData.length);
      } else if (clientsResponse.data && Array.isArray(clientsResponse.data)) {
        // 直接是数组数据
        clientData = clientsResponse.data;
        totalCount = clientsResponse.data.length;
        console.log('使用数组格式解析，clients数量:', clientData.length);
      } else if (clientsResponse.list && Array.isArray(clientsResponse.list)) {
        // 备用格式：数据在list字段中
        clientData = clientsResponse.list;
        totalCount = clientsResponse.total || 0;
        console.log('使用list格式解析，clients数量:', clientData.length);
      } else if (clientsResponse.code === 0 && clientsResponse.data) {
        // 旧格式：数据在data中
        clientData = clientsResponse.data.list || clientsResponse.data.clients || [];
        totalCount = clientsResponse.data.total || 0;
        console.log('使用旧API格式解析，clients数量:', clientData.length);
      } else {
        console.warn('未识别的响应格式:', clientsResponse);
        console.log('响应结构分析:', {
          hasClients: !!clientsResponse?.clients,
          hasData: !!clientsResponse?.data,
          hasList: !!clientsResponse?.list,
          hasPagination: !!clientsResponse?.pagination,
          dataKeys: clientsResponse?.data ? Object.keys(clientsResponse.data) : [],
          responseType: typeof clientsResponse,
          isArray: Array.isArray(clientsResponse)
        });
        clientData = [];
        totalCount = 0;
      }

      if (clientData.length > 0) {
        // 字段映射 - 将后端的下划线字段名映射为前端的驼峰命名
        const mappedClientData = clientData.map((client: any) => ({
          ...client,
          idCard: client.id_card,        // 后端 id_card -> 前端 idCard
          contactPerson: client.contact_person,  // 后端 contact_person -> 前端 contactPerson
          contactPhone: client.contact_phone,    // 后端 contact_phone -> 前端 contactPhone
          createdAt: client.created_at,          // 后端 created_at -> 前端 createdAt
          updatedAt: client.updated_at,          // 后端 updated_at -> 前端 updatedAt
        }));

        // 格式化数据以适配组件需求
        const formattedClients = mappedClientData.map((client: any) => ({
          clientId: client.id,
          clientName: client.name || client.company || `客户${client.id}`,
          phone: client.phone || '未提供',
          email: client.email || '',
          clientType: client.type || (client.company ? '企业' : '个人'),
          company: client.company || '',
          address: client.address || '',
          status: client.status || 'active',
          // 保留原始字段以备需要
          idCard: client.idCard,
          contactPerson: client.contactPerson,
          contactPhone: client.contactPhone,
          source: client.source,
          notes: client.notes,
          createdAt: client.createdAt,
          updatedAt: client.updatedAt
        }));
        console.log('格式化客户数据:', formattedClients);
        console.log('客户总数:', totalCount);
        setClients(formattedClients);
      } else {
        console.log('客户数据格式不正确或为空:', clientsResponse);
        console.log('响应结构分析:', {
          hasSuccess: !!clientsResponse?.success,
          hasData: !!clientsResponse?.data,
          dataKeys: clientsResponse?.data ? Object.keys(clientsResponse.data) : [],
          responseType: typeof clientsResponse,
          isArray: Array.isArray(clientsResponse)
        });
        message.error('客户数据加载失败，请检查API连接和响应格式');
      }

      // 获取律师数据
      console.log('获取律师数据...');
      let lawyersResponse = null;

      try {
        console.log('尝试律师端点: /lawfirm/lawyers');
        lawyersResponse = await get('/lawfirm/lawyers', { page: 1, page_size: 100 });
        console.log('律师API成功:', lawyersResponse);
      } catch (error) {
        console.error('律师API失败:', error);
        lawyersResponse = null;
      }

      // 律师数据处理
      if (lawyersResponse?.data?.list && Array.isArray(lawyersResponse.data.list)) {
        const formattedLawyers = lawyersResponse.data.list.map((lawyer: any) => ({
          lawyerId: lawyer.id,
          lawyerName: lawyer.name || `律师${lawyer.id}`,
          phone: lawyer.phone || '未提供',
          email: lawyer.email || '',
          position: lawyer.position || '律师',
          department: lawyer.department || '律师事务所',
          specialty: lawyer.specialty || '法律咨询'
        }));
        console.log('格式化律师数据:', formattedLawyers);
        setLawyers(formattedLawyers);
      } else if (lawyersResponse?.data && Array.isArray(lawyersResponse.data)) {
        const formattedLawyers = lawyersResponse.data.map((lawyer: any) => ({
          lawyerId: lawyer.id,
          lawyerName: lawyer.name || `律师${lawyer.id}`,
          phone: lawyer.phone || '未提供',
          email: lawyer.email || '',
          position: lawyer.position || '律师',
          department: lawyer.department || '律师事务所',
          specialty: lawyer.specialty || '法律咨询'
        }));
        console.log('格式化律师数据:', formattedLawyers);
        setLawyers(formattedLawyers);
      } else if (lawyersResponse && Array.isArray(lawyersResponse)) {
        // 🔧 新增：处理HTTP拦截器转换后的格式 - 直接是律师数组
        const formattedLawyers = lawyersResponse.map((lawyer: any) => ({
          lawyerId: lawyer.id,
          lawyerName: lawyer.name || `律师${lawyer.id}`,
          phone: lawyer.phone || '未提供',
          email: lawyer.email || '',
          position: lawyer.position || '律师',
          department: lawyer.department || '律师事务所',
          specialty: lawyer.specialty || '法律咨询'
        }));
        console.log('使用HTTP拦截器转换格式解析，律师数量:', formattedLawyers.length);
        console.log('格式化律师数据:', formattedLawyers);
        setLawyers(formattedLawyers);
      } else {
        console.log('律师数据格式不正确或为空:', lawyersResponse);
        console.log('律师响应结构分析:', {
          hasSuccess: !!lawyersResponse?.success,
          hasData: !!lawyersResponse?.data,
          dataKeys: lawyersResponse?.data ? Object.keys(lawyersResponse.data) : [],
          responseType: typeof lawyersResponse,
          isArray: Array.isArray(lawyersResponse),
          code: lawyersResponse?.code,
          msg: lawyersResponse?.msg
        });

        // 临时处理：律师API不可用时的提示
        message.warning('律师数据暂时无法加载，您可以稍后再试或联系管理员修复律师API');
        console.warn('律师API端点问题：所有尝试的端点都返回错误。请检查以下端点：');
        console.warn('- /lawfirm/lawyers');
        console.warn('- /lawyers');
        console.warn('- /users/lawyers');
      }

      // 设置案件类型数据 - 优先确保有默认值
      setCaseTypes([
        { id: 1, name: "民事案件", code: "civil" },
        { id: 2, name: "商事案件", code: "commercial" },
        { id: 3, name: "刑事案件", code: "criminal" },
        { id: 4, name: "行政案件", code: "administrative" },
        { id: 5, name: "劳动案件", code: "labor" },
        { id: 6, name: "知识产权案件", code: "intellectual" },
        { id: 7, name: "金融案件", code: "financial" },
      ]);

      // 尝试从后端获取案件类型数据（可选）
      try {
        const caseTypesResponse = await get('/case-types');
        if (caseTypesResponse?.data && Array.isArray(caseTypesResponse.data)) {
          const formattedCaseTypes = caseTypesResponse.data.map((type: any) => ({
            id: type.id,
            name: type.name,
            code: type.code || type.name.toLowerCase()
          }));
          console.log('使用后端案件类型数据:', formattedCaseTypes);
          setCaseTypes(formattedCaseTypes);
        } else if (caseTypesResponse && Array.isArray(caseTypesResponse)) {
          // 🔧 新增：处理HTTP拦截器转换后的格式 - 直接是案件类型数组
          const formattedCaseTypes = caseTypesResponse.map((type: any) => ({
            id: type.id,
            name: type.name,
            code: type.code || type.name.toLowerCase()
          }));
          console.log('使用HTTP拦截器转换格式解析案件类型:', formattedCaseTypes);
          setCaseTypes(formattedCaseTypes);
        }
      } catch (error) {
        console.log('案件类型API调用失败，使用默认类型:', error);
      }

      // 项目类型和收费方式使用固定数据
      setProjectTypes([
        { id: 1, name: "新项目", code: "new" },
        { id: 2, name: "进行中项目", code: "ongoing" },
        { id: 3, name: "已完成项目", code: "completed" },
        { id: 4, name: "暂停项目", code: "suspended" },
      ]);

      setBillingMethods([
        { id: 1, name: "定额收费", code: "FIXED" },
        { id: 2, name: "按时收费", code: "HOURLY" },
        { id: 3, name: "风险代理", code: "RISK" },
        { id: 4, name: "混合收费", code: "MIXED" },
      ]);

      console.log('数据加载完成');

    } catch (error) {
      console.error('加载数据失败:', error);
      message.error('数据加载失败，请刷新页面重试');

      // 只设置基本的静态数据，不设置模拟的客户和律师
      setCaseTypes([
        { id: 1, name: "民事案件", code: "civil" },
        { id: 2, name: "商事案件", code: "commercial" },
        { id: 3, name: "刑事案件", code: "criminal" },
        { id: 4, name: "行政案件", code: "administrative" },
        { id: 5, name: "劳动案件", code: "labor" },
        { id: 6, name: "知识产权案件", code: "intellectual" },
      ]);

      setProjectTypes([
        { id: 1, name: "新项目", code: "new" },
        { id: 2, name: "进行中项目", code: "ongoing" },
        { id: 3, name: "已完成项目", code: "completed" },
        { id: 4, name: "暂停项目", code: "suspended" },
      ]);

      setBillingMethods([
        { id: 1, name: "定额收费", code: "FIXED" },
        { id: 2, name: "按时收费", code: "HOURLY" },
        { id: 3, name: "风险代理", code: "RISK" },
        { id: 4, name: "混合收费", code: "MIXED" },
      ]);
    } finally {
      setLoading(false);
    }
  };

  // 下一步
  const handleNext = async () => {
    try {
      const values = await form.validateFields();
      const newFormData = { ...formData, ...values };
      setFormData(newFormData);

      if (currentStep === 2) {
        // 在团队分配步骤后，自动进行利益冲突检查
        await performConflictCheck(newFormData);
        // 检查完成后，自动跳转到冲突分析结果页面（第3步）
        setCurrentStep(3);
        return;
      }

      // 如果当前是利益冲突检查步骤（第3步），需要确认用户已经确认了冲突结果
      if (currentStep === 3 && !conflictConfirmed) {
        message.warning("请先确认利益冲突检查结果后再继续");
        return;
      }

      setCurrentStep(currentStep + 1);
    } catch (error) {
      console.error("表单验证失败:", error);
    }
  };

  // 上一步
  const handlePrev = () => {
    setCurrentStep(currentStep - 1);
  };

  // 处理利益冲突检查确认
  const handleConflictConfirm = () => {
    setConflictConfirmed(true);
    setCurrentStep(3); // 确保停留在冲突结果页面
    message.success('利益冲突检查结果已确认');
  };

  // 执行利益冲突检查
  const performConflictCheck = async (caseData: Partial<CaseInfo>) => {
    setConflictChecking(true);
    setConflictResult({
      status: "checking",
      score: 0,
      conflicts: [],
      summary: "正在检查...",
    });

    try {
      // 调用真实的后端API进行冲突检查
      const result = await conflictAPI.check({
        clientId: caseData.clientId || undefined,
        clientName: clients.find(c => c.clientId === caseData.clientId)?.clientName,
        caseName: caseData.caseName,
        caseType: caseData.caseType, // API内部会自动转换为小写
        opponentInfo: caseData.opponentInfo,
        lawyerId: caseData.lawyerId || undefined,
        causeOfAction: caseData.causeOfAction,
        searchYears: 5,
        searchDepth: 'deep',
        includeCorporateRelations: true
      });

      // 转换后端响应为前端格式
      const convertedResult = {
        status: result.hasConflict ? "warning" : "passed",
        score: Math.round(result.riskAssessment?.riskScore || 0),
        checkTime: result.checkTime,
        checker: "智能冲突检查系统",
        totalChecked: result.checkStatistics?.totalCasesChecked || 0,
        conflicts: (result.conflictCases || []).map(c => ({
          id: c.caseId || 'unknown',
          type: c.conflictType || 'unknown',
          level: c.riskLevel || 'low',
          description: c.description || '无描述',
          relatedCase: `案件: ${c.caseName || '未知案件'}`,
          recommendation: "建议详细了解相关情况",
          details: c.description || '无详细信息',
          foundTime: result.checkTime || new Date().toLocaleString(),
          impact: (c.riskLevel || 'low') === 'HIGH' ? '高影响' : (c.riskLevel || 'low') === 'MEDIUM' ? '中等影响' : '轻微影响',
          probability: (c.riskLevel || 'low') === 'HIGH' ? 85 : (c.riskLevel || 'low') === 'MEDIUM' ? 65 : 30,
          severity: (c.riskLevel || 'low') === 'HIGH' ? 4 : (c.riskLevel || 'low') === 'MEDIUM' ? 3 : 2,
          evidence: [{
            type: "系统检查记录",
            description: c.description || '系统检查记录',
            date: (result.checkTime || new Date().toISOString()).split(' ')[0],
            caseNumber: c.caseId || 'unknown'
          }]
        })),
        summary: (result.recommendations || []).join('；'),
        riskFactors: (result.riskAssessment?.riskFactors || []).map((factor, index) => ({
          factor: factor,
          weight: 25,
          score: Math.round(result.riskAssessment?.riskScore || 0),
          description: factor
        })),
        recommendations: (result.recommendations || []).map(rec => ({
          priority: (result.riskAssessment?.overallRisk || 'LOW') === 'HIGH' ? 'high' : 'medium',
          action: rec,
          description: rec,
          timeline: "接受委托前"
        })),
        relatedCases: (result.conflictCases || []).map(c => ({
          caseId: c.caseId || 'unknown',
          caseName: c.caseName || '未知案件',
          status: c.status || 'unknown',
          relationship: c.conflictType || 'unknown',
          riskLevel: (c.riskLevel || 'low') === 'HIGH' ? '高' : (c.riskLevel || 'low') === 'MEDIUM' ? '中' : '低'
        })),
        complianceNotes: "根据《律师执业管理办法》相关规定进行检查"
      };

      setConflictResult(convertedResult);
    } catch (error) {
      // API调用失败时，提供清晰的错误信息，不使用假数据
      console.error("冲突检查API调用失败:", error);

      // 设置检查失败状态，只提供基本信息
      setConflictResult({
        status: "error",
        score: 0,
        checkTime: new Date().toLocaleString(),
        checker: "智能冲突检查系统",
        totalChecked: 0,
        conflicts: [],
        summary: "冲突检查服务暂时不可用，请稍后重试",
        riskFactors: [],
        recommendations: [
          {
            priority: "high",
            action: "稍后重试检查",
            description: "网络连接或服务可能暂时不可用",
            timeline: "立即"
          }
        ],
        relatedCases: [],
        complianceNotes: "检查服务暂时不可用"
      });
    } finally {
      setConflictChecking(false);
    }
  };

  // 提交创建案件
  const handleSubmit = async () => {
    try {
      setLoading(true);

      const caseData: CaseInfo = {
        ...formData,
        caseNo: formData.caseNo || "",
        caseName: formData.caseName || "",
        caseType: formData.caseType || "",
        clientId: formData.clientId || undefined,
        lawyerId: formData.lawyerId ?? null,
        startDate: formData.startDate
          ? dayjs(formData.startDate).format("YYYY-MM-DD")
          : undefined,
        endDate: formData.endDate
          ? dayjs(formData.endDate).format("YYYY-MM-DD")
          : undefined,
        status: "0",
        conflictCheckStatus:
          conflictResult?.status === "passed" ? "PASSED" : "WARNING",
        isMajorRisk: formData.isMajorRisk || false,
        isMassCase: formData.isMassCase || false,
        isSensitiveCase: formData.isSensitiveCase || false,
      };

      // 调用案件创建API
      const response = await fetch('/api/cases', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          title: caseData.caseName,
          description: caseData.description || '',
          client_id: caseData.clientId,
          lawyer_id: caseData.lawyerId,
          case_type: caseData.caseType,
          priority: 'medium',
          status: caseData.status || 'pending'
        })
      });
      
      if (!response.ok) {
        throw new Error('创建案件失败');
      }
      message.success("案件创建成功");
      onSuccess();
      onCancel();
    } catch (error) {
      message.error("创建案件失败");
    } finally {
      setLoading(false);
    }
  };

  // 渲染基本信息步骤
  const renderBasicInfo = () => (
    <div
      style={{ background: "#fafafa", padding: "16px", borderRadius: "8px" }}
    >
      <div style={{ marginBottom: "16px", textAlign: "center" }}>
        <Avatar
          size={48}
          icon={<FileTextOutlined />}
          style={{ backgroundColor: "#1890ff", marginBottom: "12px" }}
        />
        <Title level={5} style={{ margin: 0, color: "#1890ff", fontSize: "16px" }}>
          案件基本信息
        </Title>
        <Text type="secondary" style={{ fontSize: "12px" }}>
          请填写案件的基础信息，这些信息将用于后续的冲突检查
        </Text>
      </div>

      <Card
        title={
          <Space>
            <FileTextOutlined style={{ color: "#1890ff" }} />
            <span>基本信息</span>
          </Space>
        }
        size="small"
        style={{ boxShadow: "0 2px 8px rgba(0,0,0,0.1)" }}
      >
        <Row gutter={[16, 16]}>
          <Col span={12}>
            <Form.Item
              label={
                <Space>
                  <span>案件名称</span>
                  <Tooltip title="案件的完整名称，建议包含当事人和争议焦点">
                    <InfoCircleOutlined style={{ color: "#1890ff" }} />
                  </Tooltip>
                </Space>
              }
              name="caseName"
              rules={[{ required: true, message: "请输入案件名称" }]}
            >
              <Input
                placeholder="例：张三诉李四合同纠纷案"
                size="large"
                prefix={<FileTextOutlined style={{ color: "#bfbfbf" }} />}
              />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item
              label={
                <Space>
                  <span>案件编号</span>
                  <Tooltip title="系统自动生成或手动输入唯一编号">
                    <InfoCircleOutlined style={{ color: "#1890ff" }} />
                  </Tooltip>
                </Space>
              }
              name="caseNo"
              rules={[{ required: true, message: "请输入案件编号" }]}
            >
              <Input
                placeholder="2024-CIVIL-001"
                size="large"
                prefix={<BulbOutlined style={{ color: "#bfbfbf" }} />}
              />
            </Form.Item>
          </Col>
        </Row>

        <Row gutter={[16, 16]}>
          <Col span={12}>
            <Form.Item
              label="案件类型"
              name="caseType"
              rules={[{ required: true, message: "请选择案件类型" }]}
            >
              <Select
                placeholder="请选择案件类型"
                size="large"
                suffixIcon={<TeamOutlined />}
              >
                {caseTypes?.map((type) => (
                  <Option key={type.id} value={type.code}>
                    <Space>
                      <Tag color="blue">{type.code}</Tag>
                      {type.name}
                    </Space>
                  </Option>
                ))}
              </Select>
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="项目类型" name="projectType">
              <Select
                placeholder="请选择项目类型"
                size="large"
                suffixIcon={<RocketOutlined />}
              >
                {projectTypes?.map((type) => (
                  <Option key={type.id} value={type.code}>
                    <Space>
                      <Tag color="green">{type.code}</Tag>
                      {type.name}
                    </Space>
                  </Option>
                ))}
              </Select>
            </Form.Item>
          </Col>
        </Row>

        <Row gutter={[16, 16]}>
          <Col span={12}>
            <Form.Item label="开始日期" name="startDate">
              <DatePicker
                style={{ width: "100%" }}
                size="large"
                placeholder="选择开始日期"
                suffixIcon={<ClockCircleOutlined />}
              />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="预计结束日期" name="endDate">
              <DatePicker
                style={{ width: "100%" }}
                size="large"
                placeholder="选择预计结束日期"
                suffixIcon={<ClockCircleOutlined />}
              />
            </Form.Item>
          </Col>
        </Row>

        <Form.Item
          label={
            <Space>
              <span>案件描述</span>
              <Text type="secondary">(详细描述有助于更准确的冲突检查)</Text>
            </Space>
          }
          name="description"
        >
          <TextArea
            rows={4}
            placeholder="请详细描述案件的基本情况、争议焦点、涉及金额等关键信息..."
            showCount
            maxLength={500}
          />
        </Form.Item>
      </Card>
    </div>
  );

  // 渲染当事人信息步骤
  const renderPartyInfo = () => (
    <div
      style={{ background: "#fafafa", padding: "16px", borderRadius: "8px" }}
    >
      <div style={{ marginBottom: "16px", textAlign: "center" }}>
        <Avatar
          size={48}
          icon={<UserOutlined />}
          style={{ backgroundColor: "#52c41a", marginBottom: "12px" }}
        />
        <Title level={5} style={{ margin: 0, color: "#52c41a", fontSize: "16px" }}>
          当事人信息
        </Title>
        <Text type="secondary" style={{ fontSize: "12px" }}>
          准确的当事人信息是进行利益冲突检查的关键依据
        </Text>
      </div>

      <Card
        title={
          <Space>
            <UserOutlined style={{ color: "#52c41a" }} />
            <span>当事人信息</span>
          </Space>
        }
        size="small"
        style={{ boxShadow: "0 2px 8px rgba(0,0,0,0.1)" }}
      >
        <Row gutter={[16, 16]}>
          <Col span={12}>
            <Form.Item
              label={
                <Space>
                  <span>委托人</span>
                  <Tooltip title="选择现有客户或联系管理员添加新客户">
                    <InfoCircleOutlined style={{ color: "#52c41a" }} />
                  </Tooltip>
                </Space>
              }
              name="clientId"
              rules={[{ required: true, message: "请选择委托人" }]}
            >
              <Select
                placeholder="请选择委托人"
                showSearch
                size="large"
                filterOption={(input, option) =>
                  !!(
                    option?.children &&
                    option.children
                      .toString()
                      .toLowerCase()
                      .indexOf(input.toLowerCase()) >= 0
                  )
                }
              >
                {clients?.map((client) => (
                  <Option key={client.clientId} value={client.clientId}>
                    <div
                      style={{
                        display: "flex",
                        justifyContent: "space-between",
                        alignItems: "center",
                      }}
                    >
                      <Space>
                        <Avatar size="small" icon={<UserOutlined />} />
                        <span>{client.clientName}</span>
                      </Space>
                      <Tag
                        color={
                          client.clientType === "个人" ? "blue" : "orange"
                        }
                      >
                        {client.clientType}
                      </Tag>
                    </div>
                  </Option>
                ))}
              </Select>
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item
              label={
                <Space>
                  <span>合同金额</span>
                  <Text type="secondary">(用于风险评估)</Text>
                </Space>
              }
              name="contractAmount"
            >
              <InputNumber
                style={{ width: "100%" }}
                size="large"
                placeholder="请输入合同金额"
                formatter={(value) =>
                  `¥ ${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ",")
                }
                parser={(value) => value!.replace(/¥\s?|(,*)/g, "") as any}
                min={0}
                precision={2}
              />
            </Form.Item>
          </Col>
        </Row>

        <Row gutter={[16, 16]}>
          <Col span={24}>
            <Form.Item
              label={
                <Space>
                  <span>对方当事人信息</span>
                  <Tooltip title="包括姓名/公司名称、联系方式、地址等">
                    <InfoCircleOutlined style={{ color: "#fa8c16" }} />
                  </Tooltip>
                </Space>
              }
              name="opponentInfo"
              rules={[
                { required: true, message: "请输入对方当事人信息" },
                { min: 5, message: "请至少输入5个字符" },
              ]}
            >
              <TextArea
                rows={4}
                placeholder="请详细输入对方当事人信息，包括：&#10;1. 姓名/公司全称&#10;2. 联系方式&#10;3. 地址&#10;4. 其他相关信息"
                showCount
                maxLength={300}
              />
            </Form.Item>
          </Col>
        </Row>

        <Row gutter={[16, 16]}>
          <Col span={24}>
            <Form.Item
              label={
                <Space>
                  <span>案由</span>
                  <Tooltip title="法律关系的性质和争议的焦点">
                    <InfoCircleOutlined style={{ color: "#fa8c16" }} />
                  </Tooltip>
                </Space>
              }
              name="causeOfAction"
              rules={[{ required: true, message: "请输入案由" }]}
            >
              <Input
                placeholder="例：合同纠纷、侵权责任纠纷、劳动争议等"
                size="large"
                prefix={<FileTextOutlined style={{ color: "#bfbfbf" }} />}
              />
            </Form.Item>
          </Col>
        </Row>
      </Card>
    </div>
  );

  // 渲染团队分配步骤
  const renderTeamAssignment = () => (
    <div
      style={{ background: "#fafafa", padding: "16px", borderRadius: "8px" }}
    >
      <div style={{ marginBottom: "16px", textAlign: "center" }}>
        <Avatar
          size={48}
          icon={<TeamOutlined />}
          style={{ backgroundColor: "#722ed1", marginBottom: "12px" }}
        />
        <Title level={5} style={{ margin: 0, color: "#722ed1", fontSize: "16px" }}>
          团队分配
        </Title>
        <Text type="secondary" style={{ fontSize: "12px" }}>
          合理的团队配置是案件成功的重要保障
        </Text>
      </div>

      <Card
        title={
          <Space>
            <TeamOutlined style={{ color: "#722ed1" }} />
            <span>团队分配</span>
          </Space>
        }
        size="small"
        style={{ boxShadow: "0 2px 8px rgba(0,0,0,0.1)" }}
      >
        <Row gutter={[16, 16]}>
          <Col span={12}>
            <Form.Item
              label={
                <Space>
                  <span>主办律师</span>
                  <Tooltip title="负责案件整体把控和主要工作">
                    <InfoCircleOutlined style={{ color: "#722ed1" }} />
                  </Tooltip>
                </Space>
              }
              name="lawyerId"
              rules={[{ required: true, message: "请选择主办律师" }]}
            >
              <Select
                placeholder="请选择主办律师"
                size="large"
                showSearch
                filterOption={(input, option) =>
                  !!(
                    option?.children &&
                    option.children
                      .toString()
                      .toLowerCase()
                      .indexOf(input.toLowerCase()) >= 0
                  )
                }
              >
                {lawyers?.map((lawyer) => (
                  <Option key={lawyer.lawyerId} value={lawyer.lawyerId}>
                    <div
                      style={{
                        display: "flex",
                        justifyContent: "space-between",
                        alignItems: "center",
                      }}
                    >
                      <Space>
                        <Avatar size="small" icon={<UserOutlined />} />
                        <div>
                          <div>{lawyer.lawyerName}</div>
                          <Text type="secondary" style={{ fontSize: "12px" }}>
                            {lawyer.department} · {lawyer.specialty}
                          </Text>
                        </div>
                      </Space>
                      <Tag color="purple">{lawyer.position}</Tag>
                    </div>
                  </Option>
                ))}
              </Select>
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item
              label={
                <Space>
                  <span>协办律师</span>
                  <Text type="secondary">(可选)</Text>
                </Space>
              }
              name="assistingLawyerId"
            >
              <Select
                placeholder="请选择协办律师"
                allowClear
                size="large"
                showSearch
                filterOption={(input, option) =>
                  !!(
                    option?.children &&
                    option.children
                      ?.toString()
                      .toLowerCase()
                      .indexOf(input.toLowerCase()) >= 0
                  )
                }
              >
                {lawyers?.map((lawyer) => (
                  <Option key={lawyer.lawyerId} value={lawyer.lawyerId}>
                    <div
                      style={{
                        display: "flex",
                        justifyContent: "space-between",
                        alignItems: "center",
                      }}
                    >
                      <Space>
                        <Avatar size="small" icon={<UserOutlined />} />
                        <div>
                          <div>{lawyer.lawyerName}</div>
                          <Text type="secondary" style={{ fontSize: "12px" }}>
                            {lawyer.department} · {lawyer.specialty}
                          </Text>
                        </div>
                      </Space>
                      <Tag color="cyan">{lawyer.position}</Tag>
                    </div>
                  </Option>
                ))}
              </Select>
            </Form.Item>
          </Col>
        </Row>

        <Row gutter={[16, 16]}>
          <Col span={12}>
            <Form.Item
              label="收费方式"
              name="billingMethod"
              rules={[{ required: true, message: "请选择收费方式" }]}
            >
              <Select placeholder="请选择收费方式" size="large">
                {billingMethods?.map((method) => (
                  <Option key={method.id} value={method.code}>
                    <Space>
                      <Tag color="orange">{method.code}</Tag>
                      {method.name}
                    </Space>
                  </Option>
                ))}
              </Select>
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item
              label={
                <Space>
                  <span>其他团队成员</span>
                  <Tooltip title="律师助理、实习生等其他参与人员">
                    <InfoCircleOutlined style={{ color: "#722ed1" }} />
                  </Tooltip>
                </Space>
              }
              name="teamMembers"
            >
              <Input
                placeholder="例：张助理、李实习生（用逗号分隔）"
                size="large"
                prefix={<TeamOutlined style={{ color: "#bfbfbf" }} />}
              />
            </Form.Item>
          </Col>
        </Row>

        <Row gutter={[16, 16]}>
          <Col span={8}>
            <Form.Item
              name="isMajorRisk"
              valuePropName="checked"
              style={{ marginBottom: 0 }}
            >
              <Checkbox>
                <Space direction="vertical" size="small" style={{ textAlign: 'center' }}>
                  <AlertOutlined
                    style={{ fontSize: "20px", color: "#ff4d4f" }}
                  />
                  <Text strong>重大风险案件</Text>
                  <Text type="secondary" style={{ fontSize: "11px" }}>
                    涉及重大利益或复杂法律问题
                  </Text>
                </Space>
              </Checkbox>
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item
              name="isMassCase"
              valuePropName="checked"
              style={{ marginBottom: 0 }}
            >
              <Checkbox>
                <Space direction="vertical" size="small" style={{ textAlign: 'center' }}>
                  <TeamOutlined
                    style={{ fontSize: "20px", color: "#fa8c16" }}
                  />
                  <Text strong>群体性案件</Text>
                  <Text type="secondary" style={{ fontSize: "11px" }}>
                    涉及多个当事人或社会影响较大
                  </Text>
                </Space>
              </Checkbox>
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item
              name="isSensitiveCase"
              valuePropName="checked"
              style={{ marginBottom: 0 }}
            >
              <Checkbox>
                <Space direction="vertical" size="small" style={{ textAlign: 'center' }}>
                  <EyeOutlined
                    style={{ fontSize: "20px", color: "#722ed1" }}
                  />
                  <Text strong>敏感案件</Text>
                  <Text type="secondary" style={{ fontSize: "11px" }}>
                    涉及政治、宗教或其他敏感因素
                  </Text>
                </Space>
              </Checkbox>
            </Form.Item>
          </Col>
        </Row>
      </Card>
    </div>
  );

  // 渲染利益冲突检查步骤
  const renderConflictCheck = () => (
    <div
      style={{ background: "#fafafa", padding: "24px", borderRadius: "8px" }}
    >
      <div style={{ marginBottom: "24px", textAlign: "center" }}>
        <Avatar
          size={64}
          icon={<SafetyOutlined />}
          style={{ backgroundColor: "#fa541c", marginBottom: "16px" }}
        />
        <Title level={4} style={{ margin: 0, color: "#fa541c" }}>
          利益冲突检查
        </Title>
        <Text type="secondary">智能分析系统正在进行全方位的利益冲突检查</Text>
      </div>

      {conflictChecking ? (
        <Card
          style={{
            textAlign: "center",
            padding: "40px 20px",
            boxShadow: "0 4px 12px rgba(0,0,0,0.1)",
          }}
        >
          <Space direction="vertical" size="large" style={{ width: "100%" }}>
            <Spin
              size="large"
              indicator={
                <LoadingOutlined
                  style={{ fontSize: 48, color: "#fa541c" }}
                  spin
                />
              }
            />
            <div>
              <Title level={4} style={{ color: "#fa541c", margin: 0 }}>
                正在进行利益冲突检查
              </Title>
              <Paragraph type="secondary">
                系统正在分析客户关系、律师历史、案件关联等多个维度...
              </Paragraph>
            </div>
            <Progress
              percent={60}
              status="active"
              strokeColor={{
                "0%": "#108ee9",
                "100%": "#87d068",
              }}
              style={{ width: "80%" }}
            />
            <Space>
              <Tag color="processing">客户冲突检查</Tag>
              <Tag color="processing">律师冲突检查</Tag>
              <Tag color="processing">案件关联分析</Tag>
              <Tag color="processing">对方当事人检查</Tag>
            </Space>
          </Space>
        </Card>
      ) : conflictResult ? (
        <ConflictCheckResult
          checkId={conflictResult.checkId || `check-${Date.now()}`}
          hasConflict={conflictResult.status !== "passed"}
          conflictCases={conflictResult.conflicts?.map(conflict => ({
            id: conflict.id,
            caseId: conflict.id,
            caseName: conflict.relatedCase?.replace('案件: ', '') || conflict.description,
            caseNo: conflict.evidence?.[0]?.caseNumber,
            conflictType: conflict.type === "client" ? "客户冲突" :
                        conflict.type === "opponent" ? "对方冲突" :
                        conflict.type === "lawyer" ? "律师冲突" : "案件冲突",
            riskLevel: conflict.level === "high" ? "HIGH" :
                       conflict.level === "medium" ? "MEDIUM" : "LOW",
            description: conflict.description,
            caseStatus: "进行中",
            clientId: conflict.id,
            clientName: conflict.relatedCase || "未知客户",
            opposingParties: [],
            conflictDetails: conflict.details || conflict.description,
            createdAt: conflict.foundTime || new Date().toISOString(),
            lawyerName: conflictResult.checker,
            lawyerId: conflictResult.lawyerId
          })) || []}
          checkStatistics={{
            totalCasesChecked: conflictResult.totalChecked || 0,
            clientHistoryCases: conflictResult.totalChecked || 0,
            relatedPartiesChecked: 0,
            corporateRelationsChecked: 0,
            timeRange: "5年",
            searchScope: "全面搜索",
            startTime: conflictResult.checkTime || new Date().toISOString(),
            endTime: new Date().toISOString()
          }}
          riskAssessment={{
            overallRisk: conflictResult.status === "passed" ? "MINIMAL" :
                         conflictResult.status === "warning" ? "MEDIUM" : "HIGH",
            riskScore: conflictResult.score / 100,
            riskReason: conflictResult.summary,
            requiresApproval: conflictResult.status !== "passed",
            riskFactors: conflictResult.riskFactors?.map(f => f.factor) || [],
            mitigation: conflictResult.recommendations?.map(r => r.action) || []
          }}
          recommendations={conflictResult.recommendations?.map(r => r.action || r.description) || []}
          checkTime={conflictResult.checkTime || new Date().toISOString()}
          duration={3000}
          onConfirm={handleConflictConfirm}
          onRetry={() => performConflictCheck(formData)}
        />
      ) : (
        <Card
          style={{
            textAlign: "center",
            padding: "40px 20px",
            boxShadow: "0 4px 12px rgba(0,0,0,0.1)",
          }}
        >
          <Space direction="vertical" size="large">
            <SafetyOutlined style={{ fontSize: "64px", color: "#d9d9d9" }} />
            <div>
              <Title level={4} style={{ color: "#8c8c8c", margin: 0 }}>
                等待进行利益冲突检查
              </Title>
              <Paragraph type="secondary">
                完成前面步骤后，系统将自动进行全面的利益冲突分析
              </Paragraph>
            </div>
          </Space>
        </Card>
      )}
    </div>
  );

  // 渲染确认信息步骤
  const renderConfirmation = () => (
    <div
      style={{ background: "#fafafa", padding: "24px", borderRadius: "8px" }}
    >
      <div style={{ marginBottom: "24px", textAlign: "center" }}>
        <Avatar
          size={64}
          icon={<CheckCircleOutlined />}
          style={{ backgroundColor: "#52c41a", marginBottom: "16px" }}
        />
        <Title level={4} style={{ margin: 0, color: "#52c41a" }}>
          确认创建案件
        </Title>
        <Text type="secondary">请仔细核对以下信息，确认无误后即可创建案件</Text>
      </div>

      <Space direction="vertical" style={{ width: "100%" }} size="large">
        {/* 案件基本信息确认 */}
        <Card
          title={
            <Space>
              <FileTextOutlined style={{ color: "#1890ff" }} />
              <span>案件基本信息</span>
            </Space>
          }
          size="small"
          style={{ boxShadow: "0 2px 8px rgba(0,0,0,0.1)" }}
        >
          <Descriptions column={2} bordered size="small">
            <Descriptions.Item label="案件名称">
              <Text strong>{formData.caseName}</Text>
            </Descriptions.Item>
            <Descriptions.Item label="案件编号">
              <Text code>{formData.caseNo}</Text>
            </Descriptions.Item>
            <Descriptions.Item label="案件类型">
              <Tag color="blue">
                {caseTypes.find((t) => t.code === formData.caseType)?.name ||
                  formData.caseType}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="项目类型">
              <Tag color="green">
                {projectTypes.find((t) => t.code === formData.projectType)
                  ?.name || formData.projectType}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="开始日期">
              {formData.startDate
                ? dayjs(formData.startDate).format("YYYY-MM-DD")
                : "未设置"}
            </Descriptions.Item>
            <Descriptions.Item label="预计结束日期">
              {formData.endDate
                ? dayjs(formData.endDate).format("YYYY-MM-DD")
                : "未设置"}
            </Descriptions.Item>
            <Descriptions.Item label="案件描述" span={2}>
              <Paragraph ellipsis={{ rows: 2, expandable: true }}>
                {formData.description || "无描述"}
              </Paragraph>
            </Descriptions.Item>
          </Descriptions>
        </Card>

        {/* 当事人信息确认 */}
        <Card
          title={
            <Space>
              <UserOutlined style={{ color: "#52c41a" }} />
              <span>当事人信息</span>
            </Space>
          }
          size="small"
          style={{ boxShadow: "0 2px 8px rgba(0,0,0,0.1)" }}
        >
          <Descriptions column={2} bordered size="small">
            <Descriptions.Item label="委托人">
              <Space>
                <Avatar size="small" icon={<UserOutlined />} />
                <Text strong>
                  {clients.find((c) => c.clientId === formData.clientId)
                    ?.clientName || "未选择"}
                </Text>
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label="合同金额">
              <Text strong style={{ color: "#fa541c" }}>
                {formData.contractAmount
                  ? `¥ ${formData.contractAmount.toLocaleString()}`
                  : "未填写"}
              </Text>
            </Descriptions.Item>
            <Descriptions.Item label="案由">
              <Tag color="orange">{formData.causeOfAction || "未填写"}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="对方当事人" span={1}>
              <Paragraph ellipsis={{ rows: 1, expandable: true }}>
                {formData.opponentInfo || "未填写"}
              </Paragraph>
            </Descriptions.Item>
          </Descriptions>
        </Card>

        {/* 团队信息确认 */}
        <Card
          title={
            <Space>
              <TeamOutlined style={{ color: "#722ed1" }} />
              <span>团队配置</span>
            </Space>
          }
          size="small"
          style={{ boxShadow: "0 2px 8px rgba(0,0,0,0.1)" }}
        >
          <Descriptions column={2} bordered size="small">
            <Descriptions.Item label="主办律师">
              <Space>
                <Avatar size="small" icon={<UserOutlined />} />
                <Text strong>
                  {lawyers.find((l) => l.lawyerId === formData.lawyerId)
                    ?.lawyerName || "未选择"}
                </Text>
                <Tag color="purple">
                  {
                    lawyers.find((l) => l.lawyerId === formData.lawyerId)
                      ?.position
                  }
                </Tag>
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label="协办律师">
              {formData.assistingLawyerId ? (
                <Space>
                  <Avatar size="small" icon={<UserOutlined />} />
                  <Text>
                    {
                      lawyers.find(
                        (l) => l.id === formData.assistingLawyerId,
                      )?.name
                    }
                  </Text>
                  <Tag color="cyan">
                    {
                      lawyers.find(
                        (l) => l.id === formData.assistingLawyerId,
                      )?.position
                    }
                  </Tag>
                </Space>
              ) : (
                <Text type="secondary">未指定</Text>
              )}
            </Descriptions.Item>
            <Descriptions.Item label="收费方式">
              <Tag color="orange">
                {billingMethods.find((m) => m.code === formData.billingMethod)
                  ?.name || formData.billingMethod}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="团队成员">
              {formData.teamMembers || <Text type="secondary">无其他成员</Text>}
            </Descriptions.Item>
          </Descriptions>

          {/* 风险标记 */}
          <Divider orientation="left" style={{ margin: "16px 0 12px 0" }}>
            风险标记
          </Divider>
          <Space wrap>
            {formData.isMajorRisk && <Tag color="red">重大风险案件</Tag>}
            {formData.isMassCase && <Tag color="orange">群体性案件</Tag>}
            {formData.isSensitiveCase && <Tag color="purple">敏感案件</Tag>}
            {!formData.isMajorRisk &&
              !formData.isMassCase &&
              !formData.isSensitiveCase && (
                <Text type="secondary">无特殊风险标记</Text>
              )}
          </Space>
        </Card>

        {/* 冲突检查结果确认 */}
        <Card
          title={
            <Space>
              <SafetyOutlined style={{ color: "#fa541c" }} />
              <span>利益冲突检查结果</span>
            </Space>
          }
          size="small"
          style={{
            boxShadow: "0 2px 8px rgba(0,0,0,0.1)",
            border: `2px solid ${
              conflictResult?.status === "passed"
                ? "#52c41a"
                : conflictResult?.status === "warning"
                  ? "#faad14"
                  : "#ff4d4f"
            }`,
          }}
        >
          {conflictResult ? (
            <Space direction="vertical" style={{ width: "100%" }} size="middle">
              <Row gutter={24} align="middle">
                <Col span={8} style={{ textAlign: "center" }}>
                  <div
                    style={{
                      width: "80px",
                      height: "80px",
                      borderRadius: "50%",
                      background: `linear-gradient(135deg, ${
                        conflictResult.status === "passed"
                          ? "#52c41a, #73d13d"
                          : conflictResult.status === "warning"
                            ? "#faad14, #ffc53d"
                            : "#ff4d4f, #ff7875"
                      })`,
                      display: "flex",
                      alignItems: "center",
                      justifyContent: "center",
                      margin: "0 auto",
                    }}
                  >
                    <div style={{ color: "white", textAlign: "center" }}>
                      <div style={{ fontSize: "24px", fontWeight: "bold" }}>
                        {conflictResult.score}
                      </div>
                      <div style={{ fontSize: "10px" }}>评分</div>
                    </div>
                  </div>
                </Col>
                <Col span={16}>
                  <Space
                    direction="vertical"
                    size="small"
                    style={{ width: "100%" }}
                  >
                    <div>
                      <Tag
                        color={
                          conflictResult.status === "passed"
                            ? "green"
                            : conflictResult.status === "warning"
                              ? "orange"
                              : "red"
                        }
                        style={{ fontSize: "14px", padding: "4px 12px" }}
                      >
                        {conflictResult.status === "passed"
                          ? "✅ 检查通过"
                          : conflictResult.status === "warning"
                            ? "⚠️ 需要注意"
                            : "❌ 存在冲突"}
                      </Tag>
                    </div>
                    <Text>{conflictResult.summary}</Text>
                    <div>
                      <Text type="secondary" style={{ fontSize: "12px" }}>
                        发现冲突: {conflictResult.conflicts?.length || 0} 项 |
                        检查时间: {conflictResult.checkTime || "刚刚"}
                      </Text>
                    </div>
                  </Space>
                </Col>
              </Row>

              {conflictResult.status !== "passed" && (
                <Alert
                  message={
                    conflictResult.status === "warning"
                      ? "注意事项"
                      : "严重警告"
                  }
                  description={
                    conflictResult.status === "warning"
                      ? "该案件存在潜在利益冲突，请确认已充分了解相关风险并做好披露工作。建议在充分披露后谨慎接受委托。"
                      : "该案件存在严重利益冲突，强烈建议拒绝委托或采取回避措施。"
                  }
                  type={
                    conflictResult.status === "warning" ? "warning" : "error"
                  }
                  showIcon
                />
              )}
            </Space>
          ) : (
            <Alert
              message="未进行冲突检查"
              description="请返回上一步完成利益冲突检查"
              type="warning"
              showIcon
            />
          )}
        </Card>

        {/* 最终确认 */}
        <Card
          style={{
            background: "linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%)",
            border: "2px dashed #1890ff",
            boxShadow: "0 4px 12px rgba(24, 144, 255, 0.1)",
          }}
        >
          <div style={{ textAlign: "center", padding: "20px 0" }}>
            <CheckCircleOutlined
              style={{
                fontSize: "48px",
                color: "#52c41a",
                marginBottom: "16px",
              }}
            />
            <Title level={4} style={{ color: "#1890ff", margin: 0 }}>
              确认创建案件
            </Title>
            <Paragraph
              type="secondary"
              style={{ marginTop: 8, marginBottom: 0 }}
            >
              点击"创建案件"按钮将正式创建此案件，并开始案件管理流程
            </Paragraph>
          </div>
        </Card>
      </Space>
    </div>
  );

  // 渲染当前步骤内容
  const renderStepContent = () => {
    switch (currentStep) {
      case 0:
        return renderBasicInfo();
      case 1:
        return renderPartyInfo();
      case 2:
        return renderTeamAssignment();
      case 3:
        return renderConflictCheck();
      case 4:
        return renderConfirmation();
      default:
        return null;
    }
  };

  return (
    <>
      {/* 🎯 Modal层级修复说明：现在使用全局CSS方案，移除内联样式以避免冲突 */}
      <ConfigProvider
        theme={{
          token: {
            // 🎯 统一字体大小和间距，确保界面比例协调
            fontSize: 14,
            fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
            borderRadius: 6,
            // 🎯 优化组件间距 - 紧凑布局
            padding: 12,
            margin: 12,
            // 🎯 统一颜色主题
            colorPrimary: '#1890ff',
            colorSuccess: '#52c41a',
            colorWarning: '#faad14',
            colorError: '#ff4d4f',
          },
          components: {
            // 🎯 针对Modal组件的特殊配置
            Modal: {
              contentBg: '#ffffff',
              headerBg: '#fafafa',
              borderRadiusLG: 8,
              zIndexPopup: 2002,  // 确保Modal弹窗的z-index
            },
            // 🎯 针对Form组件的特殊配置
            Form: {
              labelColor: '#262626',
              itemMarginBottom: 16,
            },
          },
        }}
      >
      <Modal
        title={
          <div style={{ textAlign: "center", padding: "10px 0" }}>
            <Space>
              <Avatar
                icon={<RocketOutlined />}
                style={{ backgroundColor: "#1890ff" }}
              />
              <div>
                <Title level={4} style={{ margin: 0, color: "#1890ff" }}>
                  智能案件创建向导
                </Title>
                <Text type="secondary">分步骤创建案件，智能冲突检查</Text>
              </div>
            </Space>
          </div>
        }
        open={visible}
        onCancel={onCancel}
        // 🎯 响应式宽度：优化适配不同屏幕尺寸
        width={{
          xs: '95vw',
          sm: '85vw',
          md: '75vw',
          lg: '800px',    // 1080p优化：800px宽度
          xl: '850px',
          xxl: '900px'
        }}
        // 🎯 使用 centered 属性实现完美居中
        centered
        footer={null}
        destroyOnClose
        // 🎯 Modal层级修复：使用全局CSS方案
        getContainer={() => document.body}
        // 🎯 简化样式配置，让全局CSS生效
        styles={{
          body: {
            padding: "0 24px 24px 24px",
            height: '85vh',            // 使用固定高度
            maxHeight: '90vh',         // 设置上限防止超出屏幕
            overflowY: 'auto',
            display: 'flex',
            flexDirection: 'column'
          }
        }}
        // 🎯 添加自定义CSS类，便于全局CSS定位
        className="create-case-wizard-modal"
    >
      <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
        {/* 步骤指示器 */}
        <div
          style={{
            background: "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
            margin: "0 -24px 16px -24px",  // 进一步减少下边距
            padding: "12px 24px",          // 减少垂直padding
            borderRadius: "0 0 12px 12px",
            flexShrink: 0                   // 防止压缩
          }}
        >
          <Steps
            current={currentStep}
            size="small"                     // 使用small尺寸减少占用空间
            style={{
              background: "rgba(255,255,255,0.1)",
              padding: "8px 12px",         // 进一步减少padding
              borderRadius: "8px",
              backdropFilter: "blur(10px)",
            }}
            progressDot={(dot, { status, index }) => (
              <div
                style={{
                  width: "20px",                   // 进一步缩小圆点尺寸
                  height: "20px",
                  borderRadius: "50%",
                  background:
                    status === "finish"
                      ? "#52c41a"
                      : status === "process"
                        ? "#1890ff"
                        : "rgba(255,255,255,0.3)",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  color: "white",
                  fontWeight: "bold",
                  fontSize: "10px",                // 进一步缩小字体尺寸
                }}
              >
                {status === "finish" ? (
                  <CheckCircleOutlined />
                ) : status === "process" ? (
                  steps[index].icon
                ) : (
                  index + 1
                )}
              </div>
            )}
          >
            {steps.map((step, index) => (
              <Step
                key={index}
                title={
                  <span style={{ color: "white", fontWeight: "bold", fontSize: "12px" }}>
                    {step.title}
                  </span>
                }
                description={
                  <span style={{ color: "rgba(255,255,255,0.8)", fontSize: "10px" }}>
                    {step.description}
                  </span>
                }
              />
            ))}
          </Steps>
        </div>

        {/* 表单内容区域 - 可滚动 */}
        <div style={{ flex: 1, overflowY: 'auto', marginBottom: '16px' }}>
          <Form form={form} layout="vertical" initialValues={formData}>
            {renderStepContent()}
          </Form>
        </div>

        {/* 操作按钮 - 固定在底部 */}
        <div
          style={{
            padding: "12px 0",               // 减少padding
            borderTop: "1px solid #f0f0f0",
            background: "linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%)",
            margin: "0 -24px -24px -24px",  // 调整边距
            flexShrink: 0,                  // 防止压缩
            borderRadius: "0 0 8px 8px",
          }}
        >
          <div style={{ textAlign: "center" }}>
            <Space size="large">
              {currentStep > 0 && (
                <Button
                  size="large"
                  onClick={handlePrev}
                  style={{ minWidth: "120px" }}
                >
                  <Space>
                    <span>上一步</span>
                  </Space>
                </Button>
              )}
              {currentStep < steps.length - 1 && (currentStep !== 3 || conflictConfirmed) && (
                <Button
                  type="primary"
                  size="large"
                  onClick={handleNext}
                  loading={currentStep === 2 && conflictChecking}
                  style={{
                    minWidth: "160px",
                    background:
                      "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
                    border: "none",
                  }}
                >
                  <Space>
                    {currentStep === 2 ? (
                      <>
                        <SafetyOutlined />
                        <span>检查冲突并继续</span>
                      </>
                    ) : currentStep === 3 && conflictConfirmed ? (
                      <>
                        <span>下一步</span>
                        <RocketOutlined />
                      </>
                    ) : (
                      <>
                        <span>下一步</span>
                        <RocketOutlined />
                      </>
                    )}
                  </Space>
                </Button>
              )}
              {currentStep === steps.length - 1 && (
                <Button
                  type="primary"
                  size="large"
                  onClick={handleSubmit}
                  loading={loading}
                  style={{
                    minWidth: "160px",
                    background:
                      "linear-gradient(135deg, #52c41a 0%, #73d13d 100%)",
                    border: "none",
                  }}
                >
                  <Space>
                    <CheckCircleOutlined />
                    <span>创建案件</span>
                  </Space>
                </Button>
              )}
              <Button
                size="large"
                onClick={onCancel}
                style={{ minWidth: "120px" }}
              >
                取消
              </Button>
            </Space>
          </div>

          {/* 进度提示 */}
          <div style={{ textAlign: "center", marginTop: "16px" }}>
            <Text type="secondary">
              步骤 {currentStep + 1} / {steps.length} |
              {currentStep === 0 && " 填写案件基本信息"}
              {currentStep === 1 && " 确认当事人信息"}
              {currentStep === 2 && " 配置团队和风险评估"}
              {currentStep === 3 && " 进行利益冲突检查"}
              {currentStep === 4 && " 最终确认并创建"}
            </Text>
          </div>
        </div>
      </div>
    </Modal>
    </ConfigProvider>
    </>
  );
};

export default CreateCaseWizard;
