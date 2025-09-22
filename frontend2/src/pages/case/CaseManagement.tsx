import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Table, 
  Button, 
  Space, 
  Tag, 
  Modal, 
  Form, 
  Input, 
  Select, 
  DatePicker, 
  message,
  Popconfirm,
  Tooltip,
  Badge,
  Drawer,
  Typography,
  Switch,
  InputNumber,
  Row,
  Col,
  Empty,
  List
} from 'antd';
import { 
  PlusOutlined, 
  EditOutlined, 
  DeleteOutlined, 
  EyeOutlined,
  SearchOutlined,
  FilterOutlined,
  FileTextOutlined,
  UserOutlined,
  CalendarOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  SyncOutlined,
  HistoryOutlined,
  SettingOutlined,
  UpOutlined,
  DownOutlined
} from '@ant-design/icons';
import CreateCaseWizard from '@/components/CreateCaseWizard';
import type { ColumnsType } from 'antd/es/table';
import { useNavigate } from 'react-router-dom';
import dayjs from 'dayjs';
import isBetween from 'dayjs/plugin/isBetween';
import { caseAPI } from '@/services/lawfirm';
import { get } from '@/services/http';

// 扩展dayjs插件
dayjs.extend(isBetween);

// 导入新组件
import AdvancedSearch from '@/components/AdvancedSearch';
import { SearchHighlight } from '@/components/SearchHighlight';
import SearchHistory from '@/components/SearchHistory';

const { Option } = Select;
const { Search } = Input;
const { TextArea } = Input;
const { Text } = Typography;
const { RangePicker } = DatePicker;

interface Case {
  caseId: number;
  caseNo: string;
  caseName: string;
  caseType: string;
  clientName: string;
  lawyerName: string;
  principalInfo?: string;
  opponentInfo?: string;
  status: string;
  description: string;
  createTime: string;
  updateTime: string;
  caseAmount?: number;
}

interface CaseFormData {
  caseNo: string;
  caseName: string;
  caseType: string;
  clientId: number | null;
  lawyerId: number | null;
  status: string;
  description: string;
  projectCode?: string;
  contractAmount?: number;
  startDate?: string;
  endDate?: string;
  teamMembers?: string;
  projectType?: string;
}

interface AdvancedSearchParams {
  searchText: string;
  caseType: string;
  status: string;
  projectType: string;
  lawyerId: string;
  clientId: string;
  dateRange: [dayjs.Dayjs, dayjs.Dayjs] | null;
  amountRange: [number, number | null] | null;
  sortBy: string;
  sortOrder: string;
}

const CaseManagement: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [cases, setCases] = useState<Case[]>([]);
  const [visible, setVisible] = useState(false);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [editingCase, setEditingCase] = useState<Case | null>(null);
  const [form] = Form.useForm();
  
  // 基础搜索状态
  const [searchText, setSearchText] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [typeFilter, setTypeFilter] = useState<string>('');
  const [lawyerFilter, setLawyerFilter] = useState<string>('');
  const [clientFilter, setClientFilter] = useState<string>('');
  // 动态下拉选项
  const [lawyerOptions, setLawyerOptions] = useState<{ label: string; value: string | number }[]>([]);
  const [clientOptions, setClientOptions] = useState<{ label: string; value: string | number }[]>([]);
  const [dateRangeFilter, setDateRangeFilter] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null);
  const [amountRangeFilter, setAmountRangeFilter] = useState<[number, number | null] | null>(null);
  
  // 高级搜索状态
  const [showAdvancedSearch, setShowAdvancedSearch] = useState(false);
  const [advancedParams, setAdvancedParams] = useState<AdvancedSearchParams | null>(null);
  const [showHistory, setShowHistory] = useState(false);
  
  // 保存的搜索条件状态
  const [savedSearches, setSavedSearches] = useState<any[]>([]);
  const [saveSearchModalVisible, setSaveSearchModalVisible] = useState(false);
  const [saveSearchForm] = Form.useForm();
  
  // 搜索高亮状态
  const [searchTerms, setSearchTerms] = useState<string[]>([]);
  const [currentSearchParams, setCurrentSearchParams] = useState<{searchText?: string}>({});
  
  // 分页状态
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });

  // 丰富的模拟数据 - 基于样例数据
  const mockCases: Case[] = [
    {
      caseId: 1,
      caseNo: 'CIV2025010001',
      caseName: '张三与李四借款纠纷',
      caseType: 'CIVIL',
      clientName: '张三',
      lawyerName: '张律师',
      status: '2',
      description: '民间借贷纠纷案件，涉及金额50万元',
      createTime: '2024-12-01 10:30:00',
      updateTime: '2025-01-15 15:45:00'
    },
    {
      caseId: 2,
      caseNo: 'CIV2025010002',
      caseName: '王五房屋买卖合同纠纷',
      caseType: 'CIVIL',
      clientName: '王五',
      lawyerName: '杨律师',
      status: '1',
      description: '房屋买卖合同纠纷，涉及金额300万元',
      createTime: '2024-12-15 09:15:00',
      updateTime: '2024-12-20 14:30:00'
    },
    {
      caseId: 3,
      caseNo: 'CIV2025010003',
      caseName: 'ABC公司劳动合同争议',
      caseType: 'CIVIL',
      clientName: 'ABC科技有限公司',
      lawyerName: '刘律师',
      status: '1',
      description: '劳动合同纠纷案件，涉及赔偿金额80万元',
      createTime: '2024-12-20 08:00:00',
      updateTime: '2024-12-25 16:00:00'
    },
    {
      caseId: 4,
      caseNo: 'CIV2025010004',
      caseName: '赵六离婚财产分割',
      caseType: 'CIVIL',
      clientName: '赵六',
      lawyerName: '陈律师',
      status: '0',
      description: '离婚财产分割案件，涉及夫妻共同财产约500万元',
      createTime: '2025-01-05 11:00:00',
      updateTime: '2025-01-05 11:00:00'
    },
    {
      caseId: 5,
      caseNo: 'CIV2025010005',
      caseName: '孙七交通事故损害赔偿',
      caseType: 'CIVIL',
      clientName: '孙七',
      lawyerName: '张律师',
      status: '2',
      description: '交通事故损害赔偿案件，涉及赔偿金额120万元',
      createTime: '2024-11-20 15:30:00',
      updateTime: '2025-01-10 10:00:00'
    },
    {
      caseId: 6,
      caseNo: 'COM2025010001',
      caseName: 'ABC公司合同纠纷',
      caseType: 'COMMERCIAL',
      clientName: 'ABC科技有限公司',
      lawyerName: '王律师',
      status: '1',
      description: '商业合同纠纷案件，涉及金额200万元',
      createTime: '2024-12-10 09:00:00',
      updateTime: '2024-12-18 17:00:00'
    },
    {
      caseId: 7,
      caseNo: 'COM2025010002',
      caseName: 'DEF公司股权纠纷',
      caseType: 'COMMERCIAL',
      clientName: 'DEF贸易集团',
      lawyerName: '赵律师',
      status: '1',
      description: '股权转让纠纷案件，涉及金额1000万元',
      createTime: '2024-12-05 14:00:00',
      updateTime: '2024-12-15 16:30:00'
    },
    {
      caseId: 8,
      caseNo: 'COM2025010003',
      caseName: 'GHI公司投资协议纠纷',
      caseType: 'COMMERCIAL',
      clientName: 'GHI投资公司',
      lawyerName: '赵律师',
      status: '0',
      description: '投资协议纠纷案件，涉及金额5000万元',
      createTime: '2025-01-08 10:00:00',
      updateTime: '2025-01-08 10:00:00'
    },
    {
      caseId: 9,
      caseNo: 'COM2025010004',
      caseName: 'JKL律所服务合同纠纷',
      caseType: 'COMMERCIAL',
      clientName: 'JKL律师事务所',
      lawyerName: '王律师',
      status: '3',
      description: '法律服务合同纠纷案件，涉及金额150万元',
      createTime: '2024-10-15 09:30:00',
      updateTime: '2024-12-20 11:00:00'
    },
    {
      caseId: 10,
      caseNo: 'COM2025010005',
      caseName: 'MNO公司商业秘密侵权',
      caseType: 'COMMERCIAL',
      clientName: 'MNO咨询集团',
      lawyerName: '孙律师',
      status: '1',
      description: '商业秘密侵权案件，涉及金额800万元',
      createTime: '2024-12-25 13:00:00',
      updateTime: '2024-12-30 15:00:00'
    },
    {
      caseId: 11,
      caseNo: 'CRI2025010001',
      caseName: '张三诈骗案',
      caseType: 'CRIMINAL',
      clientName: '张三',
      lawyerName: '李律师',
      status: '1',
      description: '诈骗案件，涉案金额100万元',
      createTime: '2024-11-15 08:30:00',
      updateTime: '2024-12-01 16:00:00'
    },
    {
      caseId: 12,
      caseNo: 'CRI2025010002',
      caseName: '李四职务侵占案',
      caseType: 'CRIMINAL',
      clientName: 'ABC科技有限公司',
      lawyerName: '李律师',
      status: '2',
      description: '职务侵占案件，涉案金额200万元',
      createTime: '2024-10-20 10:00:00',
      updateTime: '2024-12-30 14:30:00'
    },
    {
      caseId: 13,
      caseNo: 'CRI2025010003',
      caseName: '王五故意伤害案',
      caseType: 'CRIMINAL',
      clientName: '王五',
      lawyerName: '李律师',
      status: '0',
      description: '故意伤害案件，被害人轻伤',
      createTime: '2025-01-12 09:00:00',
      updateTime: '2025-01-12 09:00:00'
    },
    {
      caseId: 14,
      caseNo: 'ADM2025010001',
      caseName: 'ABC公司行政处罚纠纷',
      caseType: 'ADMINISTRATIVE',
      clientName: 'ABC科技有限公司',
      lawyerName: '刘律师',
      status: '1',
      description: '行政处罚纠纷案件，涉及罚款金额50万元',
      createTime: '2024-12-18 11:00:00',
      updateTime: '2024-12-28 15:30:00'
    },
    {
      caseId: 15,
      caseNo: 'ADM2025010002',
      caseName: '赵六行政复议案',
      caseType: 'ADMINISTRATIVE',
      clientName: '赵六',
      lawyerName: '刘律师',
      status: '2',
      description: '行政复议案件，涉及土地使用权争议',
      createTime: '2024-11-25 14:00:00',
      updateTime: '2025-01-05 10:00:00'
    },
    {
      caseId: 16,
      caseNo: 'ADV2025010001',
      caseName: 'GHI公司合规咨询',
      caseType: 'ADVISORY',
      clientName: 'GHI投资公司',
      lawyerName: '王律师',
      status: '3',
      description: '企业合规咨询项目，帮助GHI公司建立完善的合规管理体系',
      createTime: '2024-12-01 09:00:00',
      updateTime: '2024-12-31 17:00:00'
    },
    {
      caseId: 17,
      caseNo: 'ADV2025010002',
      caseName: 'MNO公司并购咨询',
      caseType: 'ADVISORY',
      clientName: 'MNO咨询集团',
      lawyerName: '赵律师',
      status: '1',
      description: '企业并购咨询项目，为MNO公司提供目标公司尽职调查和交易结构设计服务',
      createTime: '2025-01-01 10:00:00',
      updateTime: '2025-01-10 16:00:00'
    },
    {
      caseId: 18,
      caseNo: 'REV2025010001',
      caseName: 'JKL律所合同审查',
      caseType: 'REVIEW',
      clientName: 'JKL律师事务所',
      lawyerName: '张律师',
      status: '3',
      description: '合同审查项目，为JKL律所审查各类业务合同，提供法律风险分析和修改建议',
      createTime: '2024-11-10 08:00:00',
      updateTime: '2024-12-15 18:00:00'
    },
    {
      caseId: 19,
      caseNo: 'REV2025010002',
      caseName: 'DEF公司文件审查',
      caseType: 'REVIEW',
      clientName: 'DEF贸易集团',
      lawyerName: '张律师',
      status: '2',
      description: '法律文件审查项目，审查DEF公司的重要合同和法律文件，确保法律合规性',
      createTime: '2024-12-20 13:00:00',
      updateTime: '2025-01-08 11:00:00'
    },
    {
      caseId: 20,
      caseNo: 'REV2025010003',
      caseName: 'ABC公司上市文件审查',
      caseType: 'REVIEW',
      clientName: 'ABC科技有限公司',
      lawyerName: '赵律师',
      status: '1',
      description: '上市文件审查项目，协助ABC公司准备IPO相关的法律文件和合规审查',
      createTime: '2025-01-03 09:30:00',
      updateTime: '2025-01-15 14:00:00'
    }
  ];

  useEffect(() => {
    fetchCases();
    loadSavedSearches();
    // 并行加载律师/客户下拉：直接请求后端已实现接口，解析正确的数据结构
    (async () => {
      try {
        const [clientRes, lawyerRes] = await Promise.all([
          get<any>('/clients', { pageNum: 1, pageSize: 9999 }).catch(() => ({ list: [] })),
          get<any>('/lawfirm/lawyers', { pageNum: 1, pageSize: 9999 }).catch(() => ({ list: [] })),
        ]);
        
        const cOpts = (clientRes?.data?.list ?? []).map((c: any) => ({
          label: c.name ?? c.clientName ?? '',
          value: c.id ?? c.clientId,
        })).filter((o: any) => o.label && o.value !== undefined);
        
        const lOpts = (lawyerRes?.data?.list ?? []).map((l: any) => ({
          label: l.lawyerName ?? l.name ?? '',
          value: l.lawyerId ?? l.id,
        })).filter((o: any) => o.label && o.value !== undefined);
        
        setClientOptions(cOpts);
        setLawyerOptions(lOpts);
      } catch (error) {
        console.error('加载客户和律师选项失败:', error);
      }
    })();
  }, []);

  // 加载保存的搜索条件
  const loadSavedSearches = () => {
    try {
      const saved = localStorage.getItem('case-saved-searches');
      if (saved) {
        setSavedSearches(JSON.parse(saved));
      }
    } catch (error) {
      console.error('加载保存的搜索条件失败:', error);
    }
  };

  // 保存搜索条件
  const saveCurrentSearch = async (name: string) => {
    try {
      const searchCondition = {
        id: `saved_${Date.now()}`,
        name,
        searchText,
        statusFilter,
        typeFilter,
        lawyerFilter,
        clientFilter,
        dateRangeFilter: dateRangeFilter ? [
          dateRangeFilter[0].toISOString(),
          dateRangeFilter[1].toISOString()
        ] : null,
        amountRangeFilter,
        createTime: new Date().toISOString()
      };
      
      const newSavedSearches = [searchCondition, ...savedSearches];
      localStorage.setItem('case-saved-searches', JSON.stringify(newSavedSearches));
      setSavedSearches(newSavedSearches);
      message.success('搜索条件保存成功');
    } catch (error) {
      message.error('保存搜索条件失败');
      console.error('保存搜索条件失败:', error);
    }
  };

  // 应用保存的搜索条件
  const applySavedSearch = (search: any) => {
    setSearchText(search.searchText || '');
    setStatusFilter(search.statusFilter || '');
    setTypeFilter(search.typeFilter || '');
    setLawyerFilter(search.lawyerFilter || '');
    setClientFilter(search.clientFilter || '');
    
    if (search.dateRangeFilter) {
      setDateRangeFilter([
        dayjs(search.dateRangeFilter[0]),
        dayjs(search.dateRangeFilter[1])
      ]);
    } else {
      setDateRangeFilter(null);
    }
    
    setAmountRangeFilter(search.amountRangeFilter || null);
    setCurrentSearchParams({ searchText: search.searchText });
    
    // 重新发起搜索
    setTimeout(() => fetchCases(), 100);
  };

  // 删除保存的搜索条件
  const deleteSavedSearch = (id: string) => {
    const newSavedSearches = savedSearches.filter(s => s.id !== id);
    localStorage.setItem('case-saved-searches', JSON.stringify(newSavedSearches));
    setSavedSearches(newSavedSearches);
    message.success('删除成功');
  };

  // 监听搜索和筛选条件变化
  useEffect(() => {
    const timer = setTimeout(() => {
      // 更新当前搜索参数用于高亮显示
      setCurrentSearchParams({ searchText });
      setPagination(prev => ({
        ...prev,
        current: 1, // 重置到第一页
      }));
    }, 300); // 防抖处理
    
    return () => clearTimeout(timer);
  }, [searchText, statusFilter, typeFilter, lawyerFilter, clientFilter, dateRangeFilter, amountRangeFilter]);

  // 监听分页变化
  useEffect(() => {
    if (pagination.total > 0) { // 避免初始加载时重复调用
      fetchCases();
    }
  }, [pagination.current, pagination.pageSize]);

  // 键盘快捷键支持
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      // Ctrl+F 或 Cmd+F 打开高级搜索
      if ((event.ctrlKey || event.metaKey) && event.key === 'f') {
        event.preventDefault();
        setShowAdvancedSearch(true);
      }
      // Escape 关闭弹出的面板
      if (event.key === 'Escape') {
        setShowAdvancedSearch(false);
        setShowHistory(false);
      }
      // Ctrl+K 或 Cmd+K 聚焦搜索框
      if ((event.ctrlKey || event.metaKey) && event.key === 'k') {
        event.preventDefault();
        const searchInput = document.querySelector('input[placeholder*="搜索"]') as HTMLInputElement;
        if (searchInput) {
          searchInput.focus();
        }
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, []);

  const fetchCases = async (searchParams?: AdvancedSearchParams) => {
    setLoading(true);
    try {
      // 构建查询参数
      const params: any = {
        pageNum: pagination.current,
        pageSize: pagination.pageSize,
      };
      
      let currentSearchTerms: string[] = [];
      
      // 使用高级搜索参数或基础搜索参数
      if (searchParams) {
        // 高级搜索
        Object.assign(params, searchParams);
        
        // 构建搜索词列表用于高亮
        if (searchParams.searchText) {
          currentSearchTerms.push(searchParams.searchText);
        }
        
        // 更新高级搜索状态
        setAdvancedParams(searchParams);
      } else {
        // 基础搜索
        if (searchText) {
          params.searchText = searchText;
          currentSearchTerms.push(searchText);
        }
        
        if (statusFilter) {
          params.status = statusFilter;
        }
        
        if (typeFilter) {
          params.caseType = typeFilter;
        }
        
        if (lawyerFilter) {
          // 找到律师选项中匹配的ID
          const lawyerOption = lawyerOptions.find(opt => opt.label === lawyerFilter);
          if (lawyerOption) {
            params.lawyerId = lawyerOption.value;
          }
        }
        
        if (clientFilter) {
          // 找到客户选项中匹配的ID
          const clientOption = clientOptions.find(opt => opt.label === clientFilter);
          if (clientOption) {
            params.clientId = clientOption.value;
          }
        }
        
        if (dateRangeFilter) {
          params.dateRange = dateRangeFilter;
        }
        
        if (amountRangeFilter) {
          params.amountRange = amountRangeFilter;
        }
        
        setAdvancedParams(null);
      }
      
      // 更新搜索词用于高亮
      setSearchTerms(currentSearchTerms);
      
      // 如果有搜索文本，将其分解为多个词进行高亮
      const allSearchTerms = [];
      if (searchParams?.searchText) {
        allSearchTerms.push(...searchParams.searchText.split(/\s+/).filter(term => term.trim()));
      } else if (searchText) {
        allSearchTerms.push(...searchText.split(/\s+/).filter(term => term.trim()));
      }
      setSearchTerms(allSearchTerms);
      
      // 调用真实的API
      // 通过开关控制是否请求客户/律师列表，避免本地后端未实现导致 404
      const useAuxiliaryLists = false;
      if (!useAuxiliaryLists) {
        const caseRes = await caseAPI.getList(params);
        const mappedRows = (caseRes.rows || []).map((item: any) => ({
          caseId: item.caseId ?? 0,
          caseNo: item.caseNo ?? '',
          caseName: item.caseName ?? '',
          caseType: item.caseType ?? '',
          clientName: (item as any).clientName ?? '',
          lawyerName: (item as any).lawyerName ?? '',
          principalInfo: (item as any).principalInfo ?? '',
          opponentInfo: (item as any).opponentInfo ?? '',
          status: item.status ?? '',
          description: item.description ?? '',
          createTime: (item as any).createTime ?? '',
          updateTime: (item as any).updateTime ?? '',
          caseAmount: (item as any).caseAmount,
        })) as Case[];
        setCases(mappedRows);
        setPagination({
          ...pagination,
          total: caseRes.total
        });
      } else {
        // 并行获取案件、客户、律师列表，构建映射后合并姓名字段
        const [caseRes, clientRes, lawyerRes] = await Promise.all([
          caseAPI.getList(params),
          get<any>('/clients', { pageNum: 1, pageSize: 9999 }).catch(() => ({ data: { list: [] } })),
          get<any>('/lawfirm/lawyers', { pageNum: 1, pageSize: 9999 }).catch(() => ({ data: { list: [] } })),
        ]);
        const clientMap = new Map<string | number, string>();
        for (const c of (clientRes?.data?.list ?? [])) {
          const id = (c as any).id ?? (c as any).clientId;
          const name = (c as any).name ?? (c as any).clientName ?? '';
          if (id != null) clientMap.set(id, name);
        }
        const lawyerMap = new Map<string | number, string>();
        for (const l of (lawyerRes?.data?.list ?? [])) {
          const id = (l as any).lawyerId ?? (l as any).id;
          const name = (l as any).lawyerName ?? (l as any).name ?? '';
          if (id != null) lawyerMap.set(id, name);
        }
        const mappedRows = (caseRes.rows || []).map((item: any) => {
          const clientId = item.clientId ?? (item as any).client_id;
          const lawyerId = item.lawyerId ?? (item as any).lawyer_id;
          const clientNameFromApi = (item as any).clientName;
          const lawyerNameFromApi = (item as any).lawyerName;
          return {
            caseId: item.caseId ?? 0,
            caseNo: item.caseNo ?? '',
            caseName: item.caseName ?? '',
            caseType: item.caseType ?? '',
            clientName: clientNameFromApi ?? (clientId != null ? (clientMap.get(clientId) ?? '') : ''),
            lawyerName: lawyerNameFromApi ?? (lawyerId != null ? (lawyerMap.get(lawyerId) ?? '') : ''),
            principalInfo: (item as any).principalInfo ?? '',
            opponentInfo: (item as any).opponentInfo ?? '',
            status: item.status ?? '',
            description: item.description ?? '',
            createTime: (item as any).createTime ?? '',
            updateTime: (item as any).updateTime ?? '',
            caseAmount: (item as any).caseAmount,
          } as Case;
        });
        setCases(mappedRows);
        setPagination({
          ...pagination,
          total: caseRes.total
        });
      }
    } catch (error) {
      message.error('获取案件列表失败');
      console.error('获取案件列表失败:', error);
    } finally {
      setLoading(false);
    }
  };
  const handleCreate = () => {
    setCreateModalVisible(true);
  };

  const handleEdit = (record: Case) => {
    setEditingCase(record);
    setVisible(true);
    
    // 延迟设置表单值，确保Modal已经打开
    setTimeout(() => {
      // 查找对应的客户和律师ID
      const clientOpt = clientOptions.find(opt => opt.label === record.clientName);
      const lawyerOpt = lawyerOptions.find(opt => opt.label === record.lawyerName);
      
      form.setFieldsValue({
        caseNo: record.caseNo,
        caseName: record.caseName,
        caseType: record.caseType,
        clientId: clientOpt?.value || null,
        lawyerId: lawyerOpt?.value || null,
        status: record.status,
        description: record.description,
        projectCode: (record as any).projectCode || '',
        contractAmount: (record as any).contractAmount || null,
        startDate: (record as any).startDate ? dayjs((record as any).startDate) : null,
        endDate: (record as any).endDate ? dayjs((record as any).endDate) : null,
        teamMembers: (record as any).teamMembers || '',
        projectType: (record as any).projectType || record.caseType,
      });
    }, 0);
  };

  const handleDelete = async (id: number) => {
    try {
      // 这里应该调用API删除数据
      // await caseService.deleteCase(id);
      
      // 模拟删除
      setCases(cases.filter(item => item.caseId !== id));
      message.success('删除成功');
    } catch (error) {
      message.error('删除失败');
    }
  };

  const handleSubmit = async (values: CaseFormData) => {
    try {
      setLoading(true);
      
      // 构建提交数据
      const submitData = {
        ...values,
        clientId: values.clientId || null,
        lawyerId: values.lawyerId || null,
      };
      
      if (editingCase) {
        // 更新案件 - 调用真实API，需要包含caseId
        const updateData = {
          ...submitData,
          caseId: editingCase.caseId,
        };
        await caseAPI.update(updateData);
        message.success('案件更新成功');
      } else {
        // 新增案件 - 调用真实API
        await caseAPI.create(submitData);
        message.success('案件创建成功');
      }
      
      // 关闭弹窗并刷新列表
      setVisible(false);
      form.resetFields();
      fetchCases();
    } catch (error) {
      console.error('保存案件失败:', error);
      message.error('保存失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const getStatusBadge = (status: string) => {
    const statusMap = {
      '0': { text: '未开始', color: 'default' },
      '1': { text: '进行中', color: 'processing' },
      '2': { text: '已结案', color: 'success' },
      '3': { text: '已归档', color: 'default' }
    };
    const config = statusMap[status as keyof typeof statusMap] || { text: '未知', color: 'default' };
    return <Badge status={config.color as any} text={config.text} />;
  };

  const getCaseTypeTag = (type: string) => {
    const typeMap = {
      'CIVIL': { text: '民事案件', color: 'blue' },
      'COMMERCIAL': { text: '商事案件', color: 'orange' },
      'CRIMINAL': { text: '刑事案件', color: 'red' },
      'ADMINISTRATIVE': { text: '行政案件', color: 'purple' },
      'ADVISORY': { text: '咨询项目', color: 'green' },
      'REVIEW': { text: '审查项目', color: 'cyan' }
    };
    const config = typeMap[type as keyof typeof typeMap] || { text: '其他', color: 'default' };
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  const columns: ColumnsType<Case> = [
    {
      title: '案件编号',
      dataIndex: 'caseNo',
      key: 'caseNo',
      render: (text) => (
        <Space>
          <FileTextOutlined />
          <span>{text}</span>
        </Space>
      )
    },
    {
      title: '案件名称',
      dataIndex: 'caseName',
      key: 'caseName',
      ellipsis: true,
      render: (text) => (
        <span title={text}>
          <SearchHighlight 
            content={text} 
            searchTerms={searchTerms}
          />
        </span>
      )
    },
    {
      title: '案件类型',
      dataIndex: 'caseType',
      key: 'caseType',
      render: (text) => getCaseTypeTag(text)
    },
    {
      title: '客户',
      dataIndex: 'clientName',
      key: 'clientName',
      render: (text) => (
        <Space>
          <UserOutlined />
          <SearchHighlight 
            content={text} 
            searchTerms={searchTerms}
          />
        </Space>
      )
    },
    {
      title: '负责律师',
      dataIndex: 'lawyerName',
      key: 'lawyerName',
      render: (text) => (
        <SearchHighlight 
          content={text} 
          searchTerms={searchTerms}
        />
      )
    },
    {
      title: '委托人',
      dataIndex: 'principalInfo',
      key: 'principalInfo',
      ellipsis: true,
      render: (text) => (
        <SearchHighlight 
          content={text ? text.split('\n')[0] : ''} 
          searchTerms={searchTerms}
        />
      )
    },
    {
      title: '对方当事人',
      dataIndex: 'opponentInfo',
      key: 'opponentInfo',
      ellipsis: true,
      render: (text) => (
        <SearchHighlight 
          content={text ? text.split('\n')[0] : ''} 
          searchTerms={searchTerms}
        />
      )
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (text) => getStatusBadge(text)
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      key: 'createTime',
      render: (text) => (
        <Space>
          <CalendarOutlined />
          <span>{text}</span>
        </Space>
      )
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space size="middle">
          <Button 
            type="text" 
            icon={<EyeOutlined />}
            title="查看详情"
            onClick={() => {
              navigate(`/case/${record.caseId}`);
            }}
          />
          <Button 
            type="text" 
            icon={<EditOutlined />}
            title="编辑"
            onClick={() => handleEdit(record)}
          />
          <Popconfirm
            title="确定要删除这个案件吗？"
            onConfirm={() => handleDelete(record.caseId)}
            okText="确定"
            cancelText="取消"
          >
            <Button 
              type="text" 
              icon={<DeleteOutlined />}
              title="删除"
              danger
            />
          </Popconfirm>
        </Space>
      )
    }
  ];

  // 服务端分页处理 - 移除客户端过滤
  const displayCases = cases; // 直接使用服务端返回的数据

  return (
    <div className="case-management">
      <Card>
        <div className="table-header">
          <div className="left">
            <h2>案件管理</h2>
          </div>
          <div className="right">
            <Space wrap>
              <Search
                placeholder="搜索案件名称、编号或客户"
                allowClear
                enterButton={<SearchOutlined />}
                style={{ width: 300 }}
                value={searchText}
                onChange={(e) => setSearchText(e.target.value)}
                onSearch={() => fetchCases()}
              />
              <Button 
                type="text" 
                icon={<HistoryOutlined />} 
                title="搜索历史"
                onClick={() => setShowHistory(!showHistory)}
                style={{ 
                  color: showHistory ? '#1890ff' : undefined,
                  backgroundColor: showHistory ? '#e6f7ff' : undefined
                }}
              />
              <Button 
                type="text" 
                icon={<FilterOutlined />} 
                title="高级搜索 (Ctrl+F)"
                onClick={() => setShowAdvancedSearch(!showAdvancedSearch)}
                style={{ 
                  color: showAdvancedSearch ? '#1890ff' : undefined,
                  backgroundColor: showAdvancedSearch ? '#e6f7ff' : undefined
                }}
              />
              <Button 
                type="text" 
                icon={<SettingOutlined />} 
                title="保存搜索条件"
                onClick={() => setSaveSearchModalVisible(true)}
                disabled={!searchText && !statusFilter && !typeFilter && !lawyerFilter && !clientFilter && !dateRangeFilter && !amountRangeFilter}
              />
              <Button 
                type="primary" 
                icon={<PlusOutlined />}
                onClick={handleCreate}
              >
                新建案件
              </Button>
            </Space>
          </div>
        </div>

        {/* 快速过滤栏 */}
        <div style={{ marginBottom: 16 }}>
          {/* 快速过滤标签 */}
          <div style={{ marginBottom: 12 }}>
            <Text strong style={{ marginRight: 12 }}>快速过滤：</Text>
            <Space wrap>
              <Tag.CheckableTag
                checked={statusFilter === '1'}
                onChange={(checked) => {
                  if (checked) {
                    setStatusFilter('1');
                  } else {
                    setStatusFilter('');
                  }
                }}
              >
                进行中案件
              </Tag.CheckableTag>
              <Tag.CheckableTag
                checked={statusFilter === '2'}
                onChange={(checked) => {
                  if (checked) {
                    setStatusFilter('2');
                  } else {
                    setStatusFilter('');
                  }
                }}
              >
                已结案案件
              </Tag.CheckableTag>
              <Tag.CheckableTag
                checked={typeFilter === 'CIVIL'}
                onChange={(checked) => {
                  if (checked) {
                    setTypeFilter('CIVIL');
                  } else {
                    setTypeFilter('');
                  }
                }}
              >
                民事案件
              </Tag.CheckableTag>
              <Tag.CheckableTag
                checked={typeFilter === 'COMMERCIAL'}
                onChange={(checked) => {
                  if (checked) {
                    setTypeFilter('COMMERCIAL');
                  } else {
                    setTypeFilter('');
                  }
                }}
              >
                商事案件
              </Tag.CheckableTag>
              <Tag.CheckableTag
                checked={dateRangeFilter !== null && dateRangeFilter.length === 2 && dayjs().diff(dateRangeFilter[0], 'days') <= 7}
                onChange={(checked) => {
                  if (checked) {
                    setDateRangeFilter([dayjs().subtract(7, 'days'), dayjs()]);
                  } else {
                    setDateRangeFilter(null);
                  }
                }}
              >
                迗七天新增
              </Tag.CheckableTag>
              <Tag.CheckableTag
                checked={dateRangeFilter !== null && dateRangeFilter.length === 2 && dayjs().diff(dateRangeFilter[0], 'days') <= 30}
                onChange={(checked) => {
                  if (checked) {
                    setDateRangeFilter([dayjs().subtract(30, 'days'), dayjs()]);
                  } else {
                    setDateRangeFilter(null);
                  }
                }}
              >
                本月新增
              </Tag.CheckableTag>
              <Tag.CheckableTag
                checked={amountRangeFilter !== null && amountRangeFilter[0] === 1000000}
                onChange={(checked) => {
                  if (checked) {
                    setAmountRangeFilter([1000000, null]);
                  } else {
                    setAmountRangeFilter(null);
                  }
                }}
              >
                高金额案件(&gt;100万)
              </Tag.CheckableTag>
              <Tag.CheckableTag
                checked={lawyerFilter === '张律师'}
                onChange={(checked) => {
                  if (checked) {
                    setLawyerFilter('张律师');
                  } else {
                    setLawyerFilter('');
                  }
                }}
              >
                张律师负责
              </Tag.CheckableTag>
            </Space>
          </div>
          
          {/* 已保存的搜索条件 */}
          {savedSearches.length > 0 && (
            <div style={{ marginBottom: 12 }}>
              <Text strong style={{ marginRight: 12 }}>快速搜索：</Text>
              <Space wrap>
                {savedSearches.slice(0, 6).map((search) => (
                  <Tag
                    key={search.id}
                    color="volcano"
                    style={{ cursor: 'pointer' }}
                    onClick={() => applySavedSearch(search)}
                  >
                    {search.name}
                  </Tag>
                ))}
                {savedSearches.length > 6 && (
                  <Tag
                    color="default"
                    style={{ cursor: 'pointer' }}
                    onClick={() => setSaveSearchModalVisible(true)}
                  >
                    +{savedSearches.length - 6}个更多...
                  </Tag>
                )}
              </Space>
            </div>
          )}
          
          {/* 详细过滤选项 */}
          <Row gutter={[16, 8]}>
            <Col span={4}>
              <Select
                placeholder="案件状态"
                allowClear
                style={{ width: '100%' }}
                value={statusFilter || undefined}
                onChange={setStatusFilter}
              >
                <Option value="0">未开始</Option>
                <Option value="1">进行中</Option>
                <Option value="2">已结案</Option>
                <Option value="3">已归档</Option>
              </Select>
            </Col>
            <Col span={4}>
              <Select
                placeholder="案件类型"
                allowClear
                style={{ width: '100%' }}
                value={typeFilter || undefined}
                onChange={setTypeFilter}
              >
                <Option value="CIVIL">民事案件</Option>
                <Option value="COMMERCIAL">商事案件</Option>
                <Option value="CRIMINAL">刑事案件</Option>
                <Option value="ADMINISTRATIVE">行政案件</Option>
                <Option value="ADVISORY">咨询项目</Option>
                <Option value="REVIEW">审查项目</Option>
              </Select>
            </Col>
            <Col span={4}>
              <Select
                placeholder="负责律师"
                allowClear
                style={{ width: '100%' }}
                value={lawyerFilter || undefined}
                onChange={(v) => setLawyerFilter(v || '')}
                showSearch
                optionFilterProp="label"
                options={lawyerOptions}
              />
            </Col>
            <Col span={4}>
              <Select
                placeholder="客户筛选"
                allowClear
                style={{ width: '100%' }}
                value={clientFilter || undefined}
                onChange={(v) => setClientFilter(v || '')}
                showSearch
                optionFilterProp="label"
                options={clientOptions}
              />
            </Col>
            <Col span={5}>
              <RangePicker
                placeholder={['开始日期', '结束日期']}
                style={{ width: '100%' }}
                value={dateRangeFilter}
                onChange={(dates) => setDateRangeFilter(dates as [dayjs.Dayjs, dayjs.Dayjs] | null)}
                presets={[
                  { label: '今日', value: [dayjs().startOf('day'), dayjs().endOf('day')] },
                  { label: '迗七天', value: [dayjs().subtract(7, 'd'), dayjs()] },
                  { label: '本月', value: [dayjs().startOf('month'), dayjs().endOf('month')] },
                ]}
              />
            </Col>
            <Col span={3}>
              <Button 
                onClick={() => {
                  setSearchText('');
                  setStatusFilter('');
                  setTypeFilter('');
                  setLawyerFilter('');
                  setClientFilter('');
                  setDateRangeFilter(null);
                  setAmountRangeFilter(null);
                  setAdvancedParams(null);
                  setCurrentSearchParams({});
                  fetchCases();
                }}
                style={{ width: '100%' }}
              >
                清除过滤
              </Button>
            </Col>
          </Row>
        </div>

        {/* 搜索状态提示 */}
        {(searchText || statusFilter || typeFilter || lawyerFilter || clientFilter || dateRangeFilter || amountRangeFilter || advancedParams) && (
          <div style={{ 
            marginBottom: 16, 
            padding: '12px', 
            backgroundColor: '#f6ffed', 
            border: '1px solid #b7eb8f',
            borderRadius: '6px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between'
          }}>
            <div>
              <Text strong style={{ color: '#52c41a' }}>
                <SearchOutlined /> 当前搜索条件：
              </Text>
              <Space style={{ marginLeft: 8 }} wrap>
                {searchText && (
                  <Tag color="blue">关键词: {searchText}</Tag>
                )}
                {statusFilter && (
                  <Tag color="green">状态: {statusFilter === '0' ? '未开始' : statusFilter === '1' ? '进行中' : statusFilter === '2' ? '已结案' : '已归档'}</Tag>
                )}
                {typeFilter && (
                  <Tag color="orange">类型: {typeFilter === 'CIVIL' ? '民事' : typeFilter === 'COMMERCIAL' ? '商事' : typeFilter === 'CRIMINAL' ? '刑事' : typeFilter === 'ADMINISTRATIVE' ? '行政' : typeFilter === 'ADVISORY' ? '咨询' : '审查'}</Tag>
                )}
                {lawyerFilter && (
                  <Tag color="purple">律师: {lawyerFilter}</Tag>
                )}
                {clientFilter && (
                  <Tag color="cyan">客户: {clientFilter}</Tag>
                )}
                {dateRangeFilter && (
                  <Tag color="geekblue">日期: {dateRangeFilter[0].format('MM-DD')} ~ {dateRangeFilter[1].format('MM-DD')}</Tag>
                )}
                {amountRangeFilter && (amountRangeFilter[0] || amountRangeFilter[1]) && (
                  <Tag color="magenta">金额: {amountRangeFilter[0] || 0}万 ~ {amountRangeFilter[1] || '无限'}万</Tag>
                )}
                {advancedParams && (
                  <>
                    {advancedParams.dateRange && (
                      <Tag color="purple">日期范围</Tag>
                    )}
                    {advancedParams.amountRange && (advancedParams.amountRange[0] || advancedParams.amountRange[1]) && (
                      <Tag color="gold">金额范围</Tag>
                    )}
                  </>
                )}
              </Space>
            </div>
            <Button 
              type="link" 
              size="small" 
              onClick={() => {
                setSearchText('');
                setStatusFilter('');
                setTypeFilter('');
                setLawyerFilter('');
                setClientFilter('');
                setDateRangeFilter(null);
                setAmountRangeFilter(null);
                setAdvancedParams(null);
                setCurrentSearchParams({});
                fetchCases();
              }}
            >
              清除搜索
            </Button>
          </div>
        )}

        <Table
          columns={columns}
          dataSource={displayCases}
          rowKey="caseId"
          loading={loading}
          pagination={{
            current: pagination.current,
            pageSize: pagination.pageSize,
            total: pagination.total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条记录`,
            onChange: (page, pageSize) => {
              setPagination({
                ...pagination,
                current: page,
                pageSize: pageSize || 10,
              });
            },
            onShowSizeChange: (page, pageSize) => {
              setPagination({
                ...pagination,
                current: page,
                pageSize: pageSize,
              });
            },
          }}
        />
        
        {showHistory && (
          <div style={{ marginTop: 16 }}>
            <SearchHistory 
              onSelect={(history) => {
                // 应用搜索历史
                setSearchText(history.filters.searchText || '');
                setStatusFilter(history.filters.status || '');
                setTypeFilter(history.filters.caseType || '');
                setLawyerFilter(history.filters.lawyerId || '');
                setClientFilter(history.filters.clientId || '');
                
                // 应用日期范围
                if (history.filters.dateRange && history.filters.dateRange.length === 2) {
                  setDateRangeFilter([
                    dayjs(history.filters.dateRange[0]),
                    dayjs(history.filters.dateRange[1])
                  ]);
                } else {
                  setDateRangeFilter(null);
                }
                
                // 应用金额范围
                setAmountRangeFilter(history.filters.amountRange || null);
                
                setCurrentSearchParams({ searchText: history.filters.searchText });
                setShowHistory(false);
                
                // 重新发起搜索
                setTimeout(() => fetchCases(), 100);
              }}
            />
          </div>
        )}
        
        {showAdvancedSearch && (
          <div style={{ marginTop: 16 }}>
            <AdvancedSearch 
              onSearch={(params) => {
                setAdvancedParams(params);
                fetchCases(params);
                setCurrentSearchParams({ searchText: params.searchText });
                setShowAdvancedSearch(false);
              }}
              onReset={() => {
                setAdvancedParams(null);
                setSearchText('');
                setStatusFilter('');
                setTypeFilter('');
                setLawyerFilter('');
                setClientFilter('');
                setDateRangeFilter(null);
                setAmountRangeFilter(null);
                setCurrentSearchParams({});
                fetchCases();
              }}
              loading={loading}
            />
          </div>
        )}
      </Card>

      <CreateCaseWizard
        visible={createModalVisible}
        onCancel={() => setCreateModalVisible(false)}
        onSuccess={() => {
          setCreateModalVisible(false);
          fetchCases();
        }}
      />
      
      {/* 原有的编辑案件 Modal 保留用于编辑功能 */}
      <Modal
        title={editingCase ? '编辑案件' : '新建案件'}
        open={visible}
        confirmLoading={loading}
        onCancel={() => {
          setVisible(false);
          form.resetFields();
        }}
        footer={null}
        width={800}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            status: '0',
            caseType: 'CIVIL',
          }}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="案件编号"
                name="caseNo"
                rules={[{ required: true, message: '请输入案件编号' }]}
              >
                <Input placeholder="请输入案件编号" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="案件名称"
                name="caseName"
                rules={[{ required: true, message: '请输入案件名称' }]}
              >
                <Input placeholder="请输入案件名称" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={8}>
              <Form.Item
                label="案件类型"
                name="caseType"
                rules={[{ required: true, message: '请选择案件类型' }]}
              >
                <Select placeholder="请选择案件类型">
                  <Option value="CIVIL">民事案件</Option>
                  <Option value="COMMERCIAL">商事案件</Option>
                  <Option value="CRIMINAL">刑事案件</Option>
                  <Option value="ADMINISTRATIVE">行政案件</Option>
                  <Option value="ADVISORY">咨询项目</Option>
                  <Option value="REVIEW">审查项目</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                label="项目类型"
                name="projectType"
              >
                <Select placeholder="请选择项目类型（可选）">
                  <Option value="CIVIL">民事诉讼</Option>
                  <Option value="COMMERCIAL">商业诉讼</Option>
                  <Option value="CRIMINAL">刑事诉讼</Option>
                  <Option value="ADVISORY">法律顾问</Option>
                  <Option value="REVIEW">合同审查</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                label="状态"
                name="status"
                rules={[{ required: true, message: '请选择状态' }]}
              >
                <Select placeholder="请选择状态">
                  <Option value="0">未开始</Option>
                  <Option value="1">进行中</Option>
                  <Option value="2">已结案</Option>
                  <Option value="3">已归档</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="客户"
                name="clientId"
                rules={[{ required: true, message: '请选择客户' }]}
              >
                <Select
                  placeholder="请选择客户"
                  showSearch
                  optionFilterProp="label"
                  options={clientOptions}
                  notFoundContent="暂无客户数据"
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="负责律师"
                name="lawyerId"
                rules={[{ required: true, message: '请选择负责律师' }]}
              >
                <Select
                  placeholder="请选择负责律师"
                  showSearch
                  optionFilterProp="label"
                  options={lawyerOptions}
                  notFoundContent="暂无律师数据"
                />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={8}>
              <Form.Item
                label="项目编号"
                name="projectCode"
              >
                <Input placeholder="请输入项目编号（可选）" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                label="合同金额"
                name="contractAmount"
              >
                <InputNumber
                  placeholder="请输入合同金额"
                  style={{ width: '100%' }}
                  min={0}
                  precision={2}
                  addonAfter="元"
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                label="团队成员"
                name="teamMembers"
              >
                <Input placeholder="请输入团队成员（可选）" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="开始日期"
                name="startDate"
              >
                <DatePicker 
                  placeholder="请选择开始日期" 
                  style={{ width: '100%' }} 
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="结束日期"
                name="endDate"
              >
                <DatePicker 
                  placeholder="请选择结束日期" 
                  style={{ width: '100%' }} 
                />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            label="案件描述"
            name="description"
            rules={[{ required: true, message: '请输入案件描述' }]}
          >
            <TextArea 
              rows={4} 
              placeholder="请输入案件描述" 
              showCount 
              maxLength={1000}
            />
          </Form.Item>

          <Form.Item>
            <Space style={{ width: '100%', justifyContent: 'flex-end' }}>
              <Button 
                onClick={() => {
                  setVisible(false);
                  form.resetFields();
                }}
              >
                取消
              </Button>
              <Button 
                type="primary" 
                htmlType="submit"
                loading={loading}
              >
                {editingCase ? '更新案件' : '创建案件'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 保存搜索条件模态框 */}
      <Modal
        title="管理搜索条件"
        open={saveSearchModalVisible}
        onCancel={() => {
          setSaveSearchModalVisible(false);
          saveSearchForm.resetFields();
        }}
        footer={null}
        width={700}
      >
        <div>
          {/* 保存当前搜索条件 */}
          {(searchText || statusFilter || typeFilter || lawyerFilter || clientFilter || dateRangeFilter || amountRangeFilter) && (
            <Card 
              title="保存当前搜索" 
              size="small" 
              style={{ marginBottom: 16 }}
            >
              <Form
                form={saveSearchForm}
                layout="inline"
                onFinish={(values) => {
                  saveCurrentSearch(values.name);
                  setSaveSearchModalVisible(false);
                  saveSearchForm.resetFields();
                }}
              >
                <Form.Item
                  name="name"
                  rules={[{ required: true, message: '请输入搜索条件名称' }]}
                  style={{ flex: 1 }}
                >
                  <Input placeholder="请输入搜索条件名称" />
                </Form.Item>
                <Form.Item>
                  <Button type="primary" htmlType="submit">
                    保存
                  </Button>
                </Form.Item>
              </Form>
              
              {/* 显示当前搜索条件 */}
              <div style={{ marginTop: 12 }}>
                <Text type="secondary">当前搜索条件：</Text>
                <div style={{ marginTop: 8 }}>
                  <Space wrap>
                    {searchText && <Tag color="blue">关键词: {searchText}</Tag>}
                    {statusFilter && <Tag color="green">状态: {statusFilter === '0' ? '未开始' : statusFilter === '1' ? '进行中' : statusFilter === '2' ? '已结案' : '已归档'}</Tag>}
                    {typeFilter && <Tag color="orange">类型: {typeFilter === 'CIVIL' ? '民事' : typeFilter === 'COMMERCIAL' ? '商事' : typeFilter === 'CRIMINAL' ? '刑事' : typeFilter === 'ADMINISTRATIVE' ? '行政' : typeFilter === 'ADVISORY' ? '咨询' : '审查'}</Tag>}
                    {lawyerFilter && <Tag color="purple">律师: {lawyerFilter}</Tag>}
                    {clientFilter && <Tag color="cyan">客户: {clientFilter}</Tag>}
                    {dateRangeFilter && <Tag color="geekblue">日期范围</Tag>}
                    {amountRangeFilter && <Tag color="magenta">金额范围</Tag>}
                  </Space>
                </div>
              </div>
            </Card>
          )}
          
          {/* 已保存的搜索条件列表 */}
          <Card title="已保存的搜索条件" size="small">
            {savedSearches.length === 0 ? (
              <Empty 
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description="暂无保存的搜索条件"
              />
            ) : (
              <List
                size="small"
                dataSource={savedSearches}
                renderItem={(search) => (
                  <List.Item
                    key={search.id}
                    actions={[
                      <Button 
                        key="apply"
                        type="text" 
                        icon={<SearchOutlined />} 
                        size="small"
                        title="应用搜索"
                        onClick={() => {
                          applySavedSearch(search);
                          setSaveSearchModalVisible(false);
                        }}
                      />,
                      <Popconfirm
                        key="delete"
                        title="确定要删除这个搜索条件吗？"
                        onConfirm={() => deleteSavedSearch(search.id)}
                        okText="确定"
                        cancelText="取消"
                      >
                        <Button 
                          type="text" 
                          icon={<DeleteOutlined />} 
                          size="small"
                          title="删除"
                          danger
                        />
                      </Popconfirm>
                    ]}
                  >
                    <List.Item.Meta
                      title={search.name}
                      description={
                        <div>
                          <div>
                            <Space wrap size="small">
                              {search.searchText && <Tag color="blue">关键词: {search.searchText}</Tag>}
                              {search.statusFilter && <Tag color="green">状态筛选</Tag>}
                              {search.typeFilter && <Tag color="orange">类型筛选</Tag>}
                              {search.lawyerFilter && <Tag color="purple">律师筛选</Tag>}
                              {search.clientFilter && <Tag color="cyan">客户筛选</Tag>}
                              {search.dateRangeFilter && <Tag color="geekblue">日期范围</Tag>}
                              {search.amountRangeFilter && <Tag color="magenta">金额范围</Tag>}
                            </Space>
                          </div>
                          <Text type="secondary" style={{ fontSize: '12px' }}>
                            <ClockCircleOutlined /> {dayjs(search.createTime).format('YYYY-MM-DD HH:mm')}
                          </Text>
                        </div>
                      }
                    />
                  </List.Item>
                )}
              />
            )}
          </Card>
        </div>
      </Modal>
    </div>
  );
};

export default CaseManagement;