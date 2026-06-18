import React from 'react'
import { Avatar, Badge, Button, Input, Modal, Progress, Select, Space, Switch, Tag, Tooltip, Upload } from 'antd'
import {
  AlertOutlined,
  ApartmentOutlined,
  ArrowLeftOutlined,
  AuditOutlined,
  BankOutlined,
  BellOutlined,
  CalendarOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  CloudUploadOutlined,
  DollarOutlined,
  DownloadOutlined,
  FileDoneOutlined,
  FileProtectOutlined,
  FileSearchOutlined,
  FileTextOutlined,
  FolderOpenOutlined,
  KeyOutlined,
  LockOutlined,
  MoreOutlined,
  PlusOutlined,
  PrinterOutlined,
  SafetyCertificateOutlined,
  SearchOutlined,
  SendOutlined,
  SettingOutlined,
  TeamOutlined,
  UserOutlined,
} from '@ant-design/icons'
import { useNavigate, useParams, useSearchParams } from 'react-router'
import { assignUserRoles, getAllRoles, getUserRoles } from '@/services/role'
import type { Role } from '@/services/role'
import { getToken, getUserInfo } from '@/utils/storage'
import { message } from '@/utils/messageHelper'
import './Batch01Prototype.less'

type Tone = 'blue' | 'teal' | 'red' | 'orange' | 'green' | 'slate'

interface MetricCardProps {
  icon: React.ReactNode
  label: string
  value: string | number
  delta: string
  tone: Tone
}

const metrics: MetricCardProps[] = [
  { icon: <FileDoneOutlined />, label: '待办事项', value: 0, delta: '正式 API', tone: 'blue' },
  { icon: <SafetyCertificateOutlined />, label: '利益冲突待复核', value: 0, delta: '正式 API', tone: 'red' },
  { icon: <AuditOutlined />, label: '待审批事项', value: 0, delta: '正式 API', tone: 'orange' },
  { icon: <FolderOpenOutlined />, label: '今日新增案件', value: 0, delta: '正式 API', tone: 'teal' },
  { icon: <FileTextOutlined />, label: '在办案件总数', value: 0, delta: '正式 API', tone: 'blue' },
  { icon: <ClockCircleOutlined />, label: '逾期任务', value: 0, delta: '正式 API', tone: 'red' },
  { icon: <DollarOutlined />, label: '合同回款预警', value: 0, delta: 'finance API', tone: 'orange' },
]

interface CommandCenterSummary {
  active_cases?: number
  clients?: number
  pending_approvals?: number
  open_conflict_tasks?: number
  unread_inbox?: number
}

interface CommandCenterWorkflow {
  intake?: number
  conflict?: number
  approval?: number
  activation?: number
}

interface CommandCenterTodo {
  id?: string | number
  type?: string
  title?: string
  content?: string
  priority?: string
  due_at?: string
  source_type?: string
  source_id?: string | number
}

interface CommandCenterRiskItem {
  id?: string
  case_id?: string | number
  case_number?: string
  case_no?: string
  title?: string
  case_type?: string
  client_id?: string | number
  client_name?: string
  matched_subject?: string
  matched_type?: string
  evidence_summary?: string
  status?: string
  risk_level?: string
  has_conflict?: boolean
  owner?: string | number
  duration?: number
  check_time?: string
  created_at?: string
  updated_at?: string
  search_parameters?: unknown
  check_result?: unknown
  conflict_cases?: Array<Record<string, unknown>>
}

interface CommandCenterApprovalItem {
  id?: string
  request_number?: string
  title?: string
  type?: string
  status?: string
  priority?: string
  current_stage?: string
  current_approver_name?: string
  created_at?: string
  timeout_at?: string
}

interface CommandCenterCaseRow {
  id?: string | number
  case_number?: string
  title?: string
  client_id?: string | number
  lawyer_id?: string | number
  client_name?: string
  case_type?: string
  status?: string
  priority?: string
  lawyer_name?: string
  updated_at?: string
}

interface CommandCenterCount {
  key?: string
  count?: number
}

interface CommandCenterActivity {
  id?: string | number
  title?: string
  type?: string
  created_at?: string
  actor_id?: string | number
}

interface CommandCenterPayload {
  summary?: CommandCenterSummary
  workflow?: CommandCenterWorkflow
  todo_items?: CommandCenterTodo[]
  risk_queue?: CommandCenterRiskItem[]
  approval_queue?: CommandCenterApprovalItem[]
  case_rows?: CommandCenterCaseRow[]
  case_stage_distribution?: CommandCenterCount[]
  risk_distribution?: CommandCenterCount[]
  overdue_tasks?: CommandCenterTodo[]
  recent_activities?: CommandCenterActivity[]
  generated_at?: string
}

interface FinanceOverviewPayload {
  invoice_stats?: {
    pending_invoice_amount?: number
  }
  payment_stats?: {
    pending_amount?: number
  }
}

interface LawyersResourcePayload {
  summary?: {
    lawyers?: number
    departments?: number
    active_cases?: number
    pending_tasks?: number
  }
  lawyers?: Record<string, unknown>[]
  capacity?: CommandCenterCount[]
  assignments?: CommandCenterCaseRow[]
  tasks?: CommandCenterTodo[]
}

interface AdminAccessPayload {
  summary?: {
    users?: number
    active_users?: number
    disabled_users?: number
    roles?: number
    permissions?: number
    pending_changes?: number
  }
  users?: Record<string, unknown>[]
  roles?: CommandCenterCount[]
  permission_changes?: Record<string, unknown>[]
  audit_events?: Record<string, unknown>[]
}

interface SettingsOverviewPayload {
  summary?: {
    settings?: number
    modules?: number
  }
  modules?: CommandCenterCount[]
  settings?: Record<string, unknown>[]
}

interface ApiEnvelope<T> {
  success?: boolean
  data?: T
  error?: { message?: string } | string
}

function listOf<T>(items: T[] | undefined | null): T[] {
  return Array.isArray(items) ? items : []
}

function textValue(value: unknown, fallback = '未填写') {
  if (value === null || value === undefined || value === '') {
    return fallback
  }
  return String(value)
}

function numberValue(value: unknown, fallback = 0) {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return value
  }
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : fallback
}

function formatApiDate(value?: string) {
  if (!value) {
    return '未设置'
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return date.toLocaleString()
}

function riskLabel(value?: string) {
  const normalized = (value || '').toUpperCase()
  if (normalized === 'CRITICAL') return '严重'
  if (normalized === 'HIGH') return '高风险'
  if (normalized === 'MEDIUM') return '中风险'
  if (normalized === 'LOW') return '低风险'
  if (normalized === 'MINIMAL') return '提示'
  return value || '未评级'
}

function conflictMatchesCaseContext(item: CommandCenterRiskItem, caseID: string, caseNumber: string, caseTitle: string) {
  const itemID = textValue(item.case_id || item.id, '')
  const itemNumber = textValue(item.case_number || item.case_no, '')
  const itemTitle = textValue(item.title, '')
  if ((caseID && itemID === caseID) || (caseNumber && itemNumber === caseNumber) || (caseTitle && itemTitle === caseTitle)) {
    return true
  }
  return listOf<Record<string, unknown>>(item.conflict_cases).some((conflictCase) => {
    const conflictCaseID = textValue(conflictCase.case_id || conflictCase.caseId || conflictCase.id, '')
    const conflictCaseNumber = textValue(conflictCase.case_no || conflictCase.case_number || conflictCase.caseNo || conflictCase.caseNumber, '')
    const conflictCaseTitle = textValue(conflictCase.case_name || conflictCase.case_title || conflictCase.caseName || conflictCase.title, '')
    return (caseID && conflictCaseID === caseID) ||
      (caseNumber && conflictCaseNumber === caseNumber) ||
      (caseTitle && conflictCaseTitle === caseTitle)
  })
}

function priorityLabel(value?: string) {
  const normalized = (value || '').toLowerCase()
  const labels: Record<string, string> = {
    critical: '紧急',
    urgent: '紧急',
    high: '高',
    medium: '中',
    normal: '中',
    low: '低',
  }
  return labels[normalized] || value || '中'
}

function statusLabel(value?: string) {
  const labels: Record<string, string> = {
    active: '办理中',
    pending: '待处理',
    in_progress: '办理中',
    completed: '已完成',
    cancelled: '已取消',
    archived: '已归档',
    draft: '草稿',
    submitted: '已提交',
    under_review: '审批中',
    approved: '已通过',
    rejected: '已拒绝',
    QUEUED: '排队中',
    RUNNING: '检测中',
    PROCESSING: '处理中',
    COMPLETED: '已完成',
    FAILED: '失败',
  }
  return labels[value || ''] || value || '未知'
}

function accountStatusLabel(value?: string) {
  const labels: Record<string, string> = {
    active: '启用',
    inactive: '停用',
    locked: '锁定',
  }
  return labels[value || ''] || statusLabel(value)
}

function roleLabel(value?: string) {
  const labels: Record<string, string> = {
    admin: '管理员',
    lawyer: '律师',
    user: '普通用户',
    partner: '合伙人',
  }
  return labels[value || ''] || value || '未授权'
}

function settingObject(value: unknown): Record<string, unknown> {
  if (value && typeof value === 'object' && !Array.isArray(value)) {
    return value as Record<string, unknown>
  }
  if (typeof value === 'string') {
    try {
      const parsed = JSON.parse(value) as unknown
      if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
        return parsed as Record<string, unknown>
      }
    } catch {
      return {}
    }
  }
  return {}
}

function recordValue(value: unknown): Record<string, any> {
  if (value && typeof value === 'object' && !Array.isArray(value)) {
    return value as Record<string, any>
  }
  if (typeof value === 'string') {
    try {
      const parsed = JSON.parse(value) as unknown
      if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
        return parsed as Record<string, any>
      }
    } catch {
      return {}
    }
  }
  return {}
}

function settingEnabled(row: Record<string, unknown>) {
  const value = settingObject(row.setting_value)
  return value.enabled !== false && textValue(row.setting_key, '').length > 0
}

interface IntakeFormState {
  title: string
  caseType: string
  priority: string
  clientId: number
  clientName: string
  opponentName: string
  description: string
  billingMethod: string
  lawyerId: number
}

interface IntakeRuntimeState {
  intake?: any
  conflict?: any
  approval?: any
  integrationStatus?: any
  apiTimings: Array<{ label: string; duration: number; at: string }>
}

interface ClientOption {
  id: number
  name: string
}

interface LawyerOption {
  id: number
  name: string
  department?: string
  seniority?: string
  position?: string
}

const defaultIntakeForm: IntakeFormState = {
  title: '',
  caseType: '',
  priority: 'medium',
  clientId: 0,
  clientName: '',
  opponentName: '',
  description: '',
  billingMethod: '',
  lawyerId: 0,
}

const caseIntakeDraftKey = 'law-oa-case-intake-draft-v1'

function loadCaseIntakeDraft(): IntakeFormState {
  if (typeof window === 'undefined') {
    return defaultIntakeForm
  }
  try {
    const raw = window.localStorage.getItem(caseIntakeDraftKey)
    if (!raw) {
      return defaultIntakeForm
    }
    const parsed = JSON.parse(raw) as Partial<IntakeFormState>
    return {
      ...defaultIntakeForm,
      ...parsed,
      clientId: Number(parsed.clientId || 0),
      lawyerId: Number(parsed.lawyerId || 0),
    }
  } catch {
    return defaultIntakeForm
  }
}

const defaultMaterials = [
  { name: '客户主体资料', material_type: 'identity', status: 'received', required: true },
  { name: '投资协议及补充协议', material_type: 'contract', status: 'received', required: true },
  { name: '初步证据目录', material_type: 'evidence', status: 'missing', required: true },
]

export const getConflictCheckFallbackMessage = () =>
  '试用版当前使用样例冲突复核流程，请在利益冲突工作台查看待复核事项。'

function dbCaseType(value: string) {
  const labels: Record<string, string> = {
    civil: '民事',
    civil_litigation: '民事诉讼',
    commercial: '商事',
    construction: '建设工程',
    ma: '商事',
    criminal: '刑事',
    administrative: '行政',
    labor: '劳动',
    intellectual: '知识产权',
    financial: '商事',
  }
  return labels[value] || '其他'
}

function normalizeClientOptions(data: any): ClientOption[] {
  const rows = Array.isArray(data)
    ? data
    : data?.clients || data?.list || data?.data?.clients || data?.data?.list || data?.data || []

  return rows
    .map((item: any) => ({
      id: numberValue(item.id),
      name: textValue(item.name, ''),
    }))
    .filter((item: ClientOption) => item.id > 0 && item.name)
}

function normalizeLawyerOptions(data: any): LawyerOption[] {
  const rows = Array.isArray(data)
    ? data
    : data?.lawyers || data?.list || data?.data?.lawyers || data?.data?.list || data?.data || []

  return rows
    .map((item: any) => ({
      id: numberValue(item.id ?? item.lawyerId ?? item.lawyer_id),
      name: textValue(item.name ?? item.lawyerName ?? item.lawyer_name, ''),
      department: textValue(item.department, ''),
      seniority: textValue(item.seniority, ''),
      position: textValue(item.position, ''),
    }))
    .filter((item: LawyerOption) => item.id > 0 && item.name)
}

function normalizeCaseRows(data: any): Record<string, unknown>[] {
  const rows = Array.isArray(data)
    ? data
    : data?.cases || data?.list || data?.data?.cases || data?.data?.list || data?.data || []

  return Array.isArray(rows) ? rows : []
}

async function apiRequest<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getToken()
  const isFormData = options.body instanceof FormData
  const response = await fetch(`/api/v1${path}`, {
    ...options,
    headers: {
      ...(isFormData ? {} : { 'Content-Type': 'application/json' }),
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(options.headers || {}),
    },
  })
  const body = (await response.json().catch(() => ({}))) as ApiEnvelope<T>
  if (!response.ok || body.success === false) {
    const error = typeof body.error === 'string' ? body.error : body.error?.message
    throw new Error(error || `API 请求失败：${response.status}`)
  }
  return (body.data ?? body) as T
}

async function fetchCommandCenter(signal: AbortSignal): Promise<CommandCenterPayload | null> {
  const token = getToken()
  const response = await fetch('/api/v1/dashboard/command-center', {
    signal,
    headers: token ? { Authorization: `Bearer ${token}` } : undefined,
  })
  if (!response.ok) {
    return null
  }
  const body = (await response.json()) as ApiEnvelope<CommandCenterPayload>
  return body.data ?? null
}

function formatDateTime(value?: string) {
  if (!value) {
    return '未连接'
  }
  return new Intl.DateTimeFormat('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  }).format(new Date(value))
}

function formatTodayText() {
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    weekday: 'long',
  }).format(new Date())
}

const toneIcon: Record<Tone, string> = {
  blue: '#1263d8',
  teal: '#12a89d',
  red: '#e8434e',
  orange: '#f59f2f',
  green: '#18a058',
  slate: '#50657d',
}

function MetricCard({ icon, label, value, delta, tone }: MetricCardProps) {
  return (
    <section className='batch-metric-card'>
      <span className={`batch-icon-badge ${tone}`}>{icon}</span>
      <div>
        <span>{label}</span>
        <strong>{value}</strong>
        <em className={delta.includes('↓') ? 'down' : 'up'}>{delta}</em>
      </div>
    </section>
  )
}

function RiskTag({ text }: { text: string }) {
  const tone =
    text.includes('高') || text.includes('直接') || text.includes('紧急')
      ? 'red'
      : text.includes('中') || text.includes('潜在')
        ? 'orange'
        : text.includes('低') || text.includes('正常') || text.includes('通过')
          ? 'green'
          : 'blue'
  return <span className={`batch-risk-tag ${tone}`}>{text}</span>
}

function PageHeader({
  eyebrow,
  title,
  subtitle,
  actions,
}: {
  eyebrow: string
  title: string
  subtitle?: string
  actions?: React.ReactNode
}) {
  return (
    <header className='batch-page-header'>
      <div>
        <div className='batch-breadcrumb'>{eyebrow}</div>
        <h1>{title}</h1>
        {subtitle && <p>{subtitle}</p>}
      </div>
      {actions && <div className='batch-header-actions'>{actions}</div>}
    </header>
  )
}

function SectionCard({
  title,
  extra,
  children,
  className = '',
}: {
  title: string
  extra?: React.ReactNode
  children: React.ReactNode
  className?: string
}) {
  return (
    <section className={`batch-card ${className}`}>
      <div className='batch-card-header'>
        <h2>{title}</h2>
        {extra}
      </div>
      {children}
    </section>
  )
}

function DataTable({ children }: { children: React.ReactNode }) {
  return <div className='batch-table-wrap'>{children}</div>
}

function normalizeApprovalAccess(approval: any) {
  if (!approval) {
    return { label: '加载中', approverIds: [] as string[], approverEmails: [] as string[], availableActions: [] as string[], canApprove: false, canReject: false, canReturn: false, canDecide: false, readonlyReason: '审批数据加载中' }
  }
  const userInfo = getUserInfo() || {}
  const currentUserId = textValue(userInfo.id, '')
  const approverName = textValue(approval.current_approver_name || approval.currentApproverName, '待分配')
  const approverIds = [textValue(approval.current_approver_id || approval.currentApproverId, '')].filter(Boolean)
  const approverEmails = [textValue(approval.current_approver_email || approval.currentApproverEmail, '')].filter(Boolean)
  const availableActions = listOf<string>(approval.available_actions || approval.availableActions)
  const canApprove = availableActions.length > 0 ? availableActions.includes('approve') : Boolean(currentUserId && approverIds.includes(currentUserId))
  const canReject = availableActions.length > 0 ? availableActions.includes('reject') : canApprove
  const canReturn = availableActions.length > 0 ? availableActions.includes('request_changes') : canApprove
  const canDecide = canApprove || canReject || canReturn
  const readonlyReason = canDecide ? undefined : `当前审批人：${approverName}。当前账号仅可查看审批进度。`
  return { label: approverName, approverIds, approverEmails, availableActions, canApprove, canReject, canReturn, canDecide, readonlyReason }
}

function StatusDot({ color }: { color: Tone }) {
  return <span className='batch-status-dot' style={{ backgroundColor: toneIcon[color] }} />
}

export function DashboardCommandCenter() {
  const navigate = useNavigate()
  const [commandCenter, setCommandCenter] = React.useState<CommandCenterPayload | null>(null)
  const [financeOverview, setFinanceOverview] = React.useState<FinanceOverviewPayload | null>(null)
  const [apiLoading, setApiLoading] = React.useState(false)
  const [apiError, setApiError] = React.useState(false)
  const [apiRefreshKey, setApiRefreshKey] = React.useState(0)
  const [globalSearchQuery, setGlobalSearchQuery] = React.useState('')
  const [globalSearchState, setGlobalSearchState] = React.useState<{
    status: 'idle' | 'results' | 'empty'
    message: string
  }>({ status: 'idle', message: '输入关键词后按 Enter 搜索' })

  React.useEffect(() => {
    const controller = new AbortController()
    setApiLoading(true)
    setApiError(false)
    fetchCommandCenter(controller.signal)
      .then((payload) => {
        if (payload) {
          setCommandCenter(payload)
          setApiError(false)
        } else {
          setApiError(true)
        }
      })
      .catch((error: unknown) => {
        if ((error as DOMException).name !== 'AbortError') {
          setApiError(true)
        }
      })
      .finally(() => setApiLoading(false))
    apiRequest<FinanceOverviewPayload>('/finance/overview')
      .then((payload) => setFinanceOverview(payload))
      .catch(() => setFinanceOverview(null))
    return () => controller.abort()
  }, [apiRefreshKey])

  const refreshCommandCenter = () => {
    setCommandCenter(null)
    setApiRefreshKey((current) => current + 1)
  }

  const summary = commandCenter?.summary
  const workflow = commandCenter?.workflow
  const todoItems = listOf(commandCenter?.todo_items)
  const riskItems = listOf(commandCenter?.risk_queue)
  const approvalItems = listOf(commandCenter?.approval_queue)
  const caseRowsLive = listOf(commandCenter?.case_rows)
  const stageCounts = listOf(commandCenter?.case_stage_distribution)
  const overdueItems = listOf(commandCenter?.overdue_tasks)
  const activities = listOf(commandCenter?.recent_activities)
  const currentUserInfo = getUserInfo()
  const currentUserName = textValue(currentUserInfo?.realName || currentUserInfo?.name || currentUserInfo?.username, '律师')
  const pendingApprovals = summary?.pending_approvals ?? 0
  const openConflicts = summary?.open_conflict_tasks ?? 0
  const activeCases = summary?.active_cases ?? 0
  const unreadInbox = summary?.unread_inbox ?? 0
  const financeWarningAmount =
    numberValue(financeOverview?.payment_stats?.pending_amount) ||
    numberValue(financeOverview?.invoice_stats?.pending_invoice_amount)
  const apiStatusText = apiError ? '接口异常' : apiLoading ? '连接中' : commandCenter ? '正式 API' : '等待连接'
  const apiStatusTone: Tone = apiError ? 'red' : commandCenter ? 'green' : 'slate'
  const dashboardMetrics = metrics.map((item) => {
    if (item.label === '待办事项') return { ...item, value: unreadInbox }
    if (item.label === '利益冲突待复核') return { ...item, value: openConflicts }
    if (item.label === '待审批事项') return { ...item, value: pendingApprovals }
    if (item.label === '今日新增案件') return { ...item, value: workflow?.intake ?? item.value }
    if (item.label === '在办案件总数') return { ...item, value: activeCases }
    if (item.label === '逾期任务') return { ...item, value: overdueItems.length }
    if (item.label === '合同回款预警') return { ...item, value: financeWarningAmount, delta: 'finance API' }
    return item
  })
  const urgentTodoCount = todoItems.filter((item) => ['critical', 'high'].includes((item.priority || '').toLowerCase())).length
  const approvalTodoCount = todoItems.filter((item) => ['approval', 'approval_request'].includes((item.type || item.source_type || '').toLowerCase())).length
  const runGlobalSearch = () => {
    const query = globalSearchQuery.trim().toLowerCase()
    if (!query) {
      setGlobalSearchState({ status: 'idle', message: '输入关键词后按 Enter 搜索' })
      return
    }
    const matches = [
      ...caseRowsLive.map((item) => `${item.title || ''} ${item.client_name || ''} ${item.case_number || ''}`),
      ...approvalItems.map((item) => `${item.title || ''} ${item.request_number || ''} ${item.current_approver_name || ''}`),
      ...riskItems.map((item) => `${item.title || ''} ${item.client_name || ''} ${item.id || ''}`),
    ].filter((text) => text.toLowerCase().includes(query))
    setGlobalSearchState(matches.length > 0
      ? { status: 'results', message: `找到 ${matches.length} 条相关案件、客户或审批记录` }
      : { status: 'empty', message: '未找到相关案件、客户或审批记录' })
  }

  return (
    <div className='batch-page'>
      <PageHeader
        eyebrow='工作台 / 指挥中心'
        title={`上午好，${currentUserName}`}
        subtitle={`今天是 ${formatTodayText()}`}
        actions={
          <>
            <Badge color={apiError ? '#e8434e' : '#12a89d'} text='正式 API' />
            <Button onClick={refreshCommandCenter} loading={apiLoading}>
              刷新接口
            </Button>
            <Input
              prefix={<SearchOutlined />}
              placeholder='全局搜索（客户、案件、文档、联系人...）'
              value={globalSearchQuery}
              onChange={(event) => setGlobalSearchQuery(event.target.value)}
              onPressEnter={runGlobalSearch}
            />
            <Button type='primary' icon={<PlusOutlined />} onClick={() => navigate('/case/create')}>
              新建立案
            </Button>
          </>
        }
      />

      {globalSearchState.status !== 'idle' && (
        <section className={`batch-search-feedback ${globalSearchState.status}`}>
          <SearchOutlined />
          <span>{globalSearchState.message}</span>
        </section>
      )}

      <section className='batch-api-status'>
        <span><StatusDot color={apiStatusTone} />当前数据源 <strong>{apiStatusText}</strong></span>
        <span>接口时间 <strong>{formatDateTime(commandCenter?.generated_at)}</strong></span>
        <span>数据库案件 <strong>{commandCenter ? activeCases : '-'}</strong></span>
        <span>冲突任务 <strong>{commandCenter ? openConflicts : '-'}</strong></span>
      </section>

      <div className='batch-metric-grid'>
        {dashboardMetrics.map((item) => (
          <MetricCard key={item.label} {...item} />
        ))}
      </div>

      <div className='batch-dashboard-layout'>
        <SectionCard
          title={`我的待办（${unreadInbox}）`}
          extra={
            <Space>
              <Button size='small' type='text'>全部</Button>
              <Button size='small'>紧急 {urgentTodoCount}</Button>
              <Button size='small'>审批 {approvalTodoCount}</Button>
              <Button size='small'>任务</Button>
            </Space>
          }
          className='span-2'
        >
          <DataTable>
            <table>
              <thead>
                <tr>
                  <th>事项</th>
                  <th>关联对象</th>
                  <th>发起人</th>
                  <th>截止时间</th>
                  <th>优先级</th>
                </tr>
              </thead>
              <tbody>
                {todoItems.map((item) => (
                  <tr key={item.id || item.title}>
                    <td>
                      <RiskTag text={textValue(item.type || item.source_type, '待办')} /> {textValue(item.title)}
                    </td>
                    <td>{textValue(item.content, '-')}</td>
                    <td>{textValue(item.source_type, '-')}</td>
                    <td>{formatApiDate(item.due_at)}</td>
                    <td>
                      <RiskTag text={textValue(item.priority, 'medium')} />
                    </td>
                  </tr>
                ))}
                {todoItems.length === 0 && (
                  <tr><td colSpan={5}>暂无数据库待办</td></tr>
                )}
              </tbody>
            </table>
          </DataTable>
          <Button type='link' aria-label="查看全部待办" onClick={() => navigate('/inbox')}>查看全部待办</Button>
        </SectionCard>

        <SectionCard title={`利益冲突待复核（${openConflicts}）`} extra={<Button type='link' aria-label="查看全部冲突任务" onClick={() => navigate('/conflict')}>查看全部冲突任务</Button>}>
          <div className='batch-list'>
            {riskItems.map((item) => (
              <article key={item.id || item.title}>
                <div>
                  <strong>{textValue(item.title)}</strong>
                  <p>发起时间：{formatApiDate(item.created_at)}</p>
                </div>
                <div>
                  <RiskTag text={riskLabel(item.risk_level)} />
                  <RiskTag text={statusLabel(item.status)} />
                </div>
                <span>客户：{textValue(item.client_name, '-')}</span>
              </article>
            ))}
            {riskItems.length === 0 && <p>暂无数据库冲突复核任务</p>}
          </div>
        </SectionCard>

        <SectionCard title={`审批提醒（${pendingApprovals}）`}>
          <div className='batch-donut-card'>
            <Progress type='circle' percent={pendingApprovals > 0 ? 100 : 0} format={() => String(pendingApprovals)} strokeColor='#1263d8' size={126} />
            <div className='batch-legend'>
              <span><StatusDot color='blue' />审批队列 <strong>{approvalItems.length}</strong></span>
              <span><StatusDot color='slate' />待处理 <strong>{pendingApprovals}</strong></span>
              <span><StatusDot color='orange' />冲突任务 <strong>{openConflicts}</strong></span>
            </div>
          </div>
        </SectionCard>

        <SectionCard title='在办案件阶段分布'>
          <div className='batch-donut-card'>
            <Progress type='circle' percent={72} format={() => String(activeCases)} strokeColor='#12a89d' trailColor='#e7edf4' size={126} />
            <div className='batch-legend compact'>
              {stageCounts.slice(0, 3).map((item) => (
                <span key={item.key}><StatusDot color='blue' />{statusLabel(item.key)} {item.count ?? 0}</span>
              ))}
              {stageCounts.length === 0 && <span><StatusDot color='slate' />暂无案件状态数据</span>}
            </div>
          </div>
        </SectionCard>

        <SectionCard title='案件趋势（近6个月）'>
          <div className='batch-line-chart'>
            {caseRowsLive.slice(0, 6).map((row, index) => (
              <span key={row.id || index} style={{ height: Math.max(16, (6 - index) * 16) }} />
            ))}
            {caseRowsLive.length === 0 && <span style={{ height: 16 }} />}
          </div>
          <div className='batch-chart-labels'>
            {caseRowsLive.slice(0, 6).map((row, index) => <span key={row.id || index}>{formatApiDate(row.updated_at).slice(0, 10)}</span>)}
          </div>
        </SectionCard>

        <SectionCard title='逾期任务 TOP5'>
          <div className='batch-overdue-list'>
            {overdueItems.map((item, index) => (
              <p key={item.id || item.title}><StatusDot color={index < 3 ? 'red' : 'orange'} />{textValue(item.title)}<RiskTag text={textValue(item.priority, 'overdue')} /></p>
            ))}
            {overdueItems.length === 0 && <p>暂无数据库逾期任务</p>}
          </div>
        </SectionCard>

        <SectionCard title='最近活动'>
          <div className='batch-activity-list'>
            {activities.map((activity, index) => (
              <p key={activity.id || activity.title || index}>
                <Avatar size='small' icon={activity.type === 'approval' ? <AuditOutlined /> : <UserOutlined />} />
                <span>{textValue(activity.title)}</span>
                <em>{formatApiDate(activity.created_at)}</em>
              </p>
            ))}
            {activities.length === 0 && <p>暂无数据库活动记录</p>}
          </div>
        </SectionCard>
      </div>
    </div>
  )
}

export function ClientMasterProfile() {
  const navigate = useNavigate()
  const [clientRows, setClientRows] = React.useState<any[]>([])
  const [selectedClientId, setSelectedClientId] = React.useState<string | number | null>(null)
  const [profile, setProfile] = React.useState<any>(null)
  const [loading, setLoading] = React.useState(false)
  const [createClientOpen, setCreateClientOpen] = React.useState(false)
  const [creatingClient, setCreatingClient] = React.useState(false)
  const [clientDraft, setClientDraft] = React.useState({
    name: '',
    type: '企业',
    phone: '',
    email: '',
    address: '',
  })
  const [contactModalOpen, setContactModalOpen] = React.useState(false)
  const [uploadModalOpen, setUploadModalOpen] = React.useState(false)
  const [contactSaving, setContactSaving] = React.useState(false)
  const [uploadSaving, setUploadSaving] = React.useState(false)
  const [uploadFile, setUploadFile] = React.useState<File | null>(null)
  const [contactDraft, setContactDraft] = React.useState({
    name: '',
    position: '',
    phone: '',
    email: '',
  })

  React.useEffect(() => {
    apiRequest<any>('/clients?page=1&page_size=20')
      .then((data) => {
        const rows = listOf<any>(data?.clients || data?.list || data)
        setClientRows(rows)
        if (!selectedClientId && rows[0]?.id) {
          setSelectedClientId(rows[0].id)
        }
      })
      .catch((error) => message.error(error instanceof Error ? error.message : '加载客户列表失败'))
  }, [selectedClientId])

  React.useEffect(() => {
    if (!selectedClientId) {
      return
    }
    setLoading(true)
    apiRequest<any>(`/clients/${selectedClientId}/master-profile`)
      .then((data) => setProfile(data))
      .catch((error) => {
        setProfile(null)
        message.error(error instanceof Error ? error.message : '加载客户档案失败')
      })
      .finally(() => setLoading(false))
  }, [selectedClientId])

  const client = profile?.client || {}
  const completeness = profile?.completeness || {}
  const relatedParties = listOf<any>(profile?.related_parties)
  const matterHistory = listOf<any>(profile?.matter_history)
  const conflictHistory = listOf<any>(profile?.conflict_history)
  const openCreateClient = () => {
    setClientDraft({ name: '', type: '企业', phone: '', email: '', address: '' })
    setCreateClientOpen(true)
  }
  const createClient = async () => {
    if (!clientDraft.name.trim()) {
      message.warning('请先填写客户名称')
      return
    }
    setCreatingClient(true)
    try {
      await apiRequest('/clients', {
        method: 'POST',
        body: JSON.stringify(clientDraft),
      })
      message.success('客户已创建')
      setCreateClientOpen(false)
      const data = await apiRequest<any>('/clients?page=1&page_size=20')
      setClientRows(listOf<any>(data?.clients || data?.list || data))
    } catch (error) {
      message.error(error instanceof Error ? error.message : '新增客户失败')
    } finally {
      setCreatingClient(false)
    }
  }
  const openContactModal = () => {
    setContactDraft({
      name: textValue(client.contact_person, ''),
      position: textValue(client.contact_position || client.position, ''),
      phone: textValue(client.contact_phone || client.phone, ''),
      email: textValue(client.email, ''),
    })
    setContactModalOpen(true)
  }
  const savePrimaryContact = async () => {
    if (!client.id) {
      message.warning('请先选择客户')
      return
    }
    if (!contactDraft.name.trim()) {
      message.warning('请填写联系人姓名')
      return
    }
    setContactSaving(true)
    try {
      const payload = {
        contact_person: contactDraft.name.trim(),
        contact_phone: contactDraft.phone.trim(),
        email: contactDraft.email.trim() || undefined,
        notes: contactDraft.position.trim()
          ? `${textValue(client.notes, '')}\n主联系人职位：${contactDraft.position.trim()}`.trim()
          : undefined,
      }
      const saveContact = (version: number) => apiRequest<any>(`/clients/${client.id}`, {
        method: 'PUT',
        body: JSON.stringify({
          version,
          ...payload,
        }),
      })
      let updatedClient: any
      try {
        updatedClient = await saveContact(Number(client.version || 1))
      } catch (error) {
        const latestClient = await apiRequest<any>(`/clients/${client.id}`)
        updatedClient = await saveContact(Number(latestClient.version || 1))
      }
      setProfile((current: any) => ({
        ...current,
        client: {
          ...(current?.client || {}),
          ...updatedClient,
          contact_person: contactDraft.name.trim(),
          contact_phone: contactDraft.phone.trim(),
          email: contactDraft.email.trim() || current?.client?.email,
          updated_at: new Date().toISOString(),
        },
      }))
      message.success('联系人已保存')
      setContactModalOpen(false)
    } catch (error) {
      message.error(error instanceof Error ? error.message : '联系人保存失败')
    } finally {
      setContactSaving(false)
    }
  }
  const uploadClientAttachment = async () => {
    if (!client.id) {
      message.warning('请先选择客户')
      return
    }
    if (!uploadFile) {
      message.warning('请先选择附件')
      return
    }
    setUploadSaving(true)
    try {
      const formData = new FormData()
      formData.append('name', uploadFile.name)
      formData.append('category', 'client_attachment')
      formData.append('entity_type', 'client')
      formData.append('entity_id', String(client.id))
      formData.append('file', uploadFile)
      await apiRequest('/documents', {
        method: 'POST',
        body: formData,
      })
      message.success('客户附件已上传')
      setUploadFile(null)
      setUploadModalOpen(false)
    } catch (error) {
      message.error(error instanceof Error ? error.message : '客户附件上传失败')
    } finally {
      setUploadSaving(false)
    }
  }

  return (
    <div className='batch-page client-profile-page'>
      <PageHeader
        eyebrow={`客户管理 / 客户档案 / ${textValue(client.name, '未选择客户')}`}
        title='客户主档案'
        actions={<Input prefix={<SearchOutlined />} placeholder='搜索客户、联系人、案件、文档...' />}
      />

      <div className='batch-client-layout'>
        <aside className='batch-client-list'>
          <div className='batch-panel-title'>
            <h2>客户列表</h2>
            <Button icon={<PlusOutlined />} onClick={openCreateClient}>新增客户</Button>
          </div>
          <Input prefix={<SearchOutlined />} placeholder='搜索客户名称/统一社会信用代码' />
          <div className='batch-client-filters'>
            <Select defaultValue='全部类型' options={[{ value: '全部类型' }]} />
            <Select defaultValue='全部状态' options={[{ value: '全部状态' }]} />
          </div>
          <div className='batch-client-items'>
            {clientRows.map((item) => (
              <article
                key={item.id || item.name}
                className={String(item.id) === String(selectedClientId) ? 'selected' : ''}
                onClick={() => setSelectedClientId(item.id)}
              >
                <BankOutlined />
                <div>
                  <strong>{textValue(item.name)}</strong>
                  <span>{textValue(item.email || item.company, '无邮箱/公司信息')}</span>
                </div>
                <RiskTag text={statusLabel(item.status)} />
              </article>
            ))}
            {clientRows.length === 0 && <p>暂无数据库客户</p>}
          </div>
          <div className='batch-pagination'>共 {clientRows.length} 条</div>
        </aside>

        <main className='batch-profile-main'>
          <section className='batch-profile-hero'>
            <span className='batch-company-avatar'><BankOutlined /></span>
            <div className='batch-profile-title'>
              <h1>{textValue(client.name, loading ? '加载中' : '未选择客户')} <Tag color='green'>{statusLabel(client.status)}</Tag></h1>
              <p>客户类型：{textValue(client.type)} · 行业：{textValue(client.industry)}</p>
            </div>
            <div className='batch-profile-metrics'>
              <div><Progress type='circle' percent={completeness.score || 0} size={64} strokeColor='#12a89d' /><span>数据完整度</span></div>
              <div><strong className={conflictHistory.length ? 'orange-text' : 'green-text'}>{conflictHistory.length ? '有检测记录' : '未见记录'}</strong><span>冲突记录</span></div>
              <div><strong className='green-text'>{statusLabel(client.status)}</strong><span>客户状态</span></div>
              <div><strong>{formatApiDate(client.created_at).slice(0, 10)}</strong><span>首次入库</span></div>
              <div><strong>{formatApiDate(client.updated_at).slice(0, 10)}</strong><span>最近更新</span></div>
            </div>
            <Space>
              <Button>编辑</Button>
              <Button>更多客户操作</Button>
              <Button type='primary' onClick={() => navigate('/case/create')}>发起新案件</Button>
            </Space>
          </section>

          <div className='batch-tabs'>
            {['基本信息', '关联方与穿透', '历史委托与案件', '冲突信息池', '联系人', '附件文档', '活动日志'].map((tab, index) => (
              <button key={tab} className={index === 0 ? 'active' : ''} aria-label={`${client.name || '客户'} ${tab}`}>{tab}</button>
            ))}
          </div>

          <div className='batch-profile-grid'>
            <SectionCard title='基本信息' className='span-2'>
              <div className='batch-info-grid'>
                {[
                  ['客户名称', textValue(client.name)],
                  ['客户类型', textValue(client.type)],
                  ['电子邮箱', textValue(client.email)],
                  ['联系电话', textValue(client.phone)],
                  ['所属行业', textValue(client.industry)],
                  ['联系人', textValue(client.contact_person)],
                  ['联系地址', textValue(client.address)],
                  ['客户来源', textValue(client.source)],
                  ['备注', textValue(client.notes)],
                ].map(([label, value]) => (
                  <p key={label}><span>{label}</span><strong>{value}</strong></p>
                ))}
              </div>
            </SectionCard>

            <SectionCard title='关系图谱（穿透预览）'>
              <div className='batch-relation-graph'>
                <span className='node main'>{textValue(client.name, '客户')}</span>
                {relatedParties.slice(0, 5).map((party: any, index: number) => (
                  <span key={party.name || index} className={`node small ${['top-left', 'top-right', 'bottom-left', 'bottom-mid', 'bottom-right'][index]}`}>{textValue(party.name)}<br />{textValue(party.relationship_type)}</span>
                ))}
              </div>
              <Button block>查看完整关系图谱</Button>
            </SectionCard>

            <SectionCard title='快速操作'>
              <div className='batch-action-list'>
                <Button icon={<PlusOutlined />} onClick={() => navigate('/case/create')}>发起新案件</Button>
                <Button icon={<FileSearchOutlined />} onClick={() => navigate('/conflict')}>发起冲突检查</Button>
                <Button icon={<UserOutlined />} onClick={openContactModal}>新增联系人</Button>
                <Button icon={<CloudUploadOutlined />} onClick={() => setUploadModalOpen(true)}>上传附件</Button>
                <Button icon={<DownloadOutlined />} onClick={() => { try { const blob = new Blob([JSON.stringify(client, null, 2)], { type: 'application/json' }); const url = URL.createObjectURL(blob); const a = document.createElement('a'); a.href = url; a.download = `client-${client.name || 'export'}.json`; a.click(); URL.revokeObjectURL(url); message.success('客户档案已导出'); } catch { message.error('导出失败，请稍后重试') } }}>导出客户档案</Button>
              </div>
            </SectionCard>

            <Modal title='新增联系人' open={contactModalOpen} onCancel={() => setContactModalOpen(false)} onOk={savePrimaryContact} okText='保存' confirmLoading={contactSaving}>
              <div className='batch-form-grid two'>
                <div className='batch-field'><span>姓名 *</span><Input placeholder='联系人姓名' value={contactDraft.name} onChange={(event) => setContactDraft((draft) => ({ ...draft, name: event.target.value }))} /></div>
                <div className='batch-field'><span>职位</span><Input placeholder='职位' value={contactDraft.position} onChange={(event) => setContactDraft((draft) => ({ ...draft, position: event.target.value }))} /></div>
                <div className='batch-field'><span>电话</span><Input placeholder='联系电话' value={contactDraft.phone} onChange={(event) => setContactDraft((draft) => ({ ...draft, phone: event.target.value }))} /></div>
                <div className='batch-field'><span>邮箱</span><Input placeholder='电子邮箱' value={contactDraft.email} onChange={(event) => setContactDraft((draft) => ({ ...draft, email: event.target.value }))} /></div>
              </div>
            </Modal>
            <Modal title='上传附件' open={uploadModalOpen} onCancel={() => setUploadModalOpen(false)} onOk={uploadClientAttachment} okText='上传' confirmLoading={uploadSaving}>
              <Upload
                beforeUpload={(file) => {
                  setUploadFile(file)
                  return false
                }}
                maxCount={1}
                fileList={uploadFile ? [{ uid: uploadFile.name, name: uploadFile.name, status: 'done' }] : []}
                onRemove={() => {
                  setUploadFile(null)
                  return true
                }}
              >
                <Button icon={<CloudUploadOutlined />}>选择附件</Button>
              </Upload>
              <p>附件将关联到当前客户档案：{textValue(client.name, '未选择客户')}</p>
            </Modal>

            <SectionCard title='别名 / 曾用名'>
              <div className='batch-key-list'>
                {listOf<string>(client.aliases).map((name, index) => <p key={`${name}-${index}`}>{name}<MoreOutlined /></p>)}
                {!client.aliases?.length && <p>数据库暂无别名记录</p>}
              </div>
            </SectionCard>

            <SectionCard title='关联方'>
              <DataTable>
                <table>
                  <tbody>
                    {relatedParties.map((party: any, index) => (
                      <tr key={`${textValue(party.id || party.name, 'related')}-${index}`}><td>{textValue(party.name)}</td><td>{textValue(party.relationship_type)}</td></tr>
                    ))}
                    {relatedParties.length === 0 && <tr><td colSpan={2}>暂无数据库关联方</td></tr>}
                  </tbody>
                </table>
              </DataTable>
            </SectionCard>

            <SectionCard title='关联案件（近12个月）'>
              <DataTable>
                <table>
                  <tbody>
                    {matterHistory.map((row: any) => (
                      <tr key={row.id}><td>{textValue(row.title)}</td><td>{textValue(row.case_type)}</td><td><RiskTag text={statusLabel(row.status)} /></td></tr>
                    ))}
                    {matterHistory.length === 0 && <tr><td colSpan={3}>暂无数据库关联案件</td></tr>}
                  </tbody>
                </table>
              </DataTable>
            </SectionCard>

            <SectionCard title='最近活动'>
              <div className='batch-activity-list profile'>
                {conflictHistory.map((activity: any) => (
                  <p key={activity.check_id}><StatusDot color='blue' /><span>{textValue(activity.case_name)}</span><em>{formatApiDate(activity.created_at)}</em></p>
                ))}
                {conflictHistory.length === 0 && <p>暂无数据库冲突活动</p>}
              </div>
            </SectionCard>
          </div>
        </main>
      </div>

      <Modal
        title='新增客户'
        open={createClientOpen}
        onCancel={() => setCreateClientOpen(false)}
        onOk={createClient}
        confirmLoading={creatingClient}
        okText='保存客户'
        cancelText='取消'
      >
        <div className='batch-client-create-form'>
          <div className='batch-client-create-field'>
            <span>客户类型 *</span>
            <Select
              value={clientDraft.type}
              options={[{ value: '企业', label: '企业' }, { value: '个人', label: '个人' }]}
              onChange={(value) => setClientDraft((current) => ({ ...current, type: value }))}
            />
          </div>
          <div className='batch-client-create-field'>
            <span>客户名称 *</span>
            <Input
              value={clientDraft.name}
              onChange={(event) => setClientDraft((current) => ({ ...current, name: event.target.value }))}
              placeholder='请输入客户名称'
            />
          </div>
          <div className='batch-client-create-field'>
            <span>联系电话</span>
            <Input
              value={clientDraft.phone}
              onChange={(event) => setClientDraft((current) => ({ ...current, phone: event.target.value }))}
              placeholder='请输入联系电话'
            />
          </div>
          <div className='batch-client-create-field'>
            <span>电子邮箱</span>
            <Input
              value={clientDraft.email}
              onChange={(event) => setClientDraft((current) => ({ ...current, email: event.target.value }))}
              placeholder='请输入电子邮箱'
            />
          </div>
          <div className='batch-client-create-field'>
            <span>联系地址</span>
            <Input
              value={clientDraft.address}
              onChange={(event) => setClientDraft((current) => ({ ...current, address: event.target.value }))}
              placeholder='请输入联系地址'
            />
          </div>
        </div>
      </Modal>
    </div>
  )
}

export function CaseManagementCenter() {
  const navigate = useNavigate()
  const [commandCenter, setCommandCenter] = React.useState<CommandCenterPayload | null>(null)
  const [loading, setLoading] = React.useState(false)
  const [searchTerm, setSearchTerm] = React.useState('')
  const [statusFilter, setStatusFilter] = React.useState('全部')

  React.useEffect(() => {
    const controller = new AbortController()
    setLoading(true)
    fetchCommandCenter(controller.signal)
      .then((data) => setCommandCenter(data))
      .catch((error: unknown) => {
        if ((error as DOMException).name !== 'AbortError') {
          message.error('加载案件管理数据失败')
        }
      })
      .finally(() => setLoading(false))
    return () => controller.abort()
  }, [])

  const summary = commandCenter?.summary || {}
  const workflow = commandCenter?.workflow || {}
  const liveCaseRows = listOf(commandCenter?.case_rows)
  const normalizedSearchTerm = searchTerm.trim().toLowerCase()
  const statusFilterMap: Record<string, string[]> = {
    全部: [],
    潜在受理: ['draft'],
    冲突复核: ['risk_review', 'conflict_ready', 'conflict_checking'],
    审批中: ['submitted', 'under_review', 'pending'],
    办理中: ['active', 'in_progress'],
    已结案: ['completed', 'archived', 'closed'],
  }
  const filteredCaseRows = liveCaseRows.filter((row) => {
    const normalizedStatus = textValue(row.status, '').toLowerCase()
    const allowedStatuses = statusFilterMap[statusFilter] || []
    const matchesStatus = allowedStatuses.length === 0 || allowedStatuses.includes(normalizedStatus)
    const searchable = [
      row.case_number,
      row.title,
      row.client_name,
      row.case_type,
      row.status,
      row.priority,
      row.lawyer_name,
    ].map((value) => textValue(value, '').toLowerCase()).join(' ')
    return matchesStatus && (!normalizedSearchTerm || searchable.includes(normalizedSearchTerm))
  })
  const exportCases = () => {
    if (filteredCaseRows.length === 0) {
      message.warning('当前筛选条件下暂无可导出案件')
      return
    }
    const report = {
      title: '案件清单导出',
      generatedAt: new Date().toISOString(),
      filters: {
        search: searchTerm.trim(),
        status: statusFilter,
      },
      total: filteredCaseRows.length,
      rows: filteredCaseRows,
    }
    const blob = new Blob([JSON.stringify(report, null, 2)], {
      type: 'application/json;charset=utf-8',
    })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `cases-${new Date().toISOString().slice(0, 10)}.json`
    document.body.appendChild(link)
    link.click()
    link.remove()
    URL.revokeObjectURL(url)
    message.success(`已导出 ${filteredCaseRows.length} 条案件`)
  }

  return (
    <div className='batch-page case-management-page'>
      <PageHeader
        eyebrow='案件管理 / 案件清单'
        title='案件管理'
        subtitle='统一管理潜在接案、冲突复核、接案审批、在办案件和结案归档。'
        actions={
          <>
            <Input
              prefix={<SearchOutlined />}
              value={searchTerm}
              onChange={(event) => setSearchTerm(event.target.value)}
              allowClear
              placeholder='搜索案件编号、客户、对方、负责人...'
            />
            <Button icon={<DownloadOutlined />} onClick={exportCases}>导出</Button>
            <Button type='primary' icon={<PlusOutlined />} loading={loading} onClick={() => navigate('/case/create')}>新建案件</Button>
          </>
        }
      />

      <div className='batch-metric-grid case-metrics'>
        {[
          { icon: <FolderOpenOutlined />, label: '在办案件', value: summary.active_cases ?? 0, delta: '正式 API', tone: 'blue' as Tone },
          { icon: <FileSearchOutlined />, label: '冲突复核中', value: workflow.conflict ?? 0, delta: '正式 API', tone: 'red' as Tone },
          { icon: <AuditOutlined />, label: '接案审批中', value: workflow.approval ?? 0, delta: '正式 API', tone: 'orange' as Tone },
          { icon: <ClockCircleOutlined />, label: '待补充材料', value: workflow.intake ?? 0, delta: '正式 API', tone: 'red' as Tone },
          { icon: <CheckCircleOutlined />, label: '客户总数', value: summary.clients ?? 0, delta: '正式 API', tone: 'teal' as Tone },
        ].map((item) => <MetricCard key={item.label} {...item} />)}
      </div>

      <div className='batch-case-layout'>
        <SectionCard
          title='案件清单'
          className='span-2'
          extra={
            <Space className='batch-filter-bar'>
              {Object.keys(statusFilterMap).map((tab) => (
                <Button key={tab} type={statusFilter === tab ? 'primary' : 'default'} onClick={() => setStatusFilter(tab)}>{tab}</Button>
              ))}
            </Space>
          }
        >
          <DataTable>
            <table className='batch-case-table'>
              <colgroup>
                <col className='case-col-no' />
                <col className='case-col-title' />
                <col className='case-col-client' />
                <col className='case-col-type' />
                <col className='case-col-status' />
                <col className='case-col-risk' />
                <col className='case-col-owner' />
                <col className='case-col-time' />
                <col className='case-col-action' />
              </colgroup>
              <thead>
                <tr>
                  <th>案件编号</th>
                  <th>案件名称</th>
                  <th>客户</th>
                  <th>类型</th>
                  <th>状态</th>
                  <th>风险</th>
                  <th>负责人</th>
                  <th>更新时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {filteredCaseRows.map((row) => (
                  <tr key={row.id || row.case_number} className={['high', 'urgent', 'critical'].includes((row.priority || '').toLowerCase()) ? 'danger-row' : ''}>
                    <td className='mono-cell'>{textValue(row.case_number || row.id)}</td>
                    <td className='strong-cell'>{textValue(row.title)}</td>
                    <td>{textValue(row.client_name)}</td>
                    <td><RiskTag text={dbCaseType(textValue(row.case_type, ''))} /></td>
                    <td><RiskTag text={statusLabel(row.status)} /></td>
                    <td><RiskTag text={priorityLabel(row.priority)} /></td>
                    <td>{textValue(row.lawyer_name)}</td>
                    <td>{formatApiDate(row.updated_at)}</td>
                    <td><Button size='small' type='primary' ghost onClick={() => navigate(`/case/${row.id}`)}>查看</Button></td>
                  </tr>
                ))}
                {filteredCaseRows.length === 0 && <tr><td colSpan={9}>{liveCaseRows.length === 0 ? '暂无数据库案件' : '当前搜索或筛选条件下暂无案件'}</td></tr>}
              </tbody>
            </table>
          </DataTable>
        </SectionCard>

        <SectionCard title='接案漏斗'>
          <div className='batch-case-funnel'>
            {[
              ['接案准备', workflow.intake ?? 0, 'blue' as Tone],
              ['冲突复核', workflow.conflict ?? 0, 'red' as Tone],
              ['接案审批', workflow.approval ?? 0, 'orange' as Tone],
              ['办理中', summary.active_cases ?? 0, 'teal' as Tone],
              ['客户总数', summary.clients ?? 0, 'green' as Tone],
            ].map((item) => (
              <p key={item[0]}><StatusDot color={item[2] as Tone} /><span>{item[0]}</span><strong>{item[1]}</strong></p>
            ))}
          </div>
        </SectionCard>

        <SectionCard title='高风险案件'>
          <div className='batch-overdue-list'>
            {filteredCaseRows.filter((row) => (row.priority || '').toLowerCase() === 'high').map((row) => (
              <p key={row.id || row.title}><StatusDot color='red' />{textValue(row.title)}<RiskTag text={statusLabel(row.status)} /></p>
            ))}
            {filteredCaseRows.filter((row) => (row.priority || '').toLowerCase() === 'high').length === 0 && <p>暂无数据库高优先级案件</p>}
          </div>
        </SectionCard>

        <SectionCard title='团队负荷' className='span-2'>
          <div className='batch-policy-grid'>
            {filteredCaseRows.slice(0, 4).map((item) => (
              <article key={`${item.lawyer_name}-${item.id}`}>
                <UserOutlined />
                <div><strong>{textValue(item.lawyer_name)}</strong><p>{textValue(item.title)}</p></div>
                <RiskTag text={textValue(item.priority, 'medium')} />
              </article>
            ))}
            {filteredCaseRows.length === 0 && <p>暂无数据库团队负荷数据</p>}
          </div>
        </SectionCard>
      </div>
    </div>
  )
}

export function CaseDetailCenter() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const [caseDetail, setCaseDetail] = React.useState<any | null>(null)
  const [commandCenter, setCommandCenter] = React.useState<CommandCenterPayload | null>(null)
  const [loading, setLoading] = React.useState(false)

  React.useEffect(() => {
    if (!id) return

    let mounted = true
    setLoading(true)
    Promise.all([
      apiRequest<any>(`/cases/${id}`),
      fetchCommandCenter(new AbortController().signal).catch(() => null),
    ])
      .then(([detail, center]) => {
        if (!mounted) return
        setCaseDetail(detail)
        setCommandCenter(center)
      })
      .catch((error: unknown) => {
        message.error(error instanceof Error ? error.message : '加载案件详情失败')
      })
      .finally(() => mounted && setLoading(false))

    return () => {
      mounted = false
    }
  }, [id])

  const client = settingObject(caseDetail?.client)
  const lawyer = settingObject(caseDetail?.lawyer)
  const currentClientID = textValue(caseDetail?.client_id || client.id, '')
  const currentClientName = textValue(client.name || caseDetail?.client_name, '')
  const relatedRows = listOf(commandCenter?.case_rows)
    .filter((row) => String(row.id) !== String(caseDetail?.id))
    .filter((row) => {
      const rowClientID = textValue(row.client_id, '')
      const rowClientName = textValue(row.client_name, '')
      return (currentClientID && rowClientID === currentClientID) || (currentClientName && rowClientName === currentClientName)
    })
    .slice(0, 5)

  if (loading) {
    return (
      <div className='batch-page case-management-page'>
        <PageHeader
          eyebrow='案件管理 / 案件详情'
          title='正在加载案件详情'
          subtitle='正在从正式 API 读取案件、客户、律师与流程数据。'
        />
      </div>
    )
  }

  if (!caseDetail) {
    return (
      <div className='batch-page case-management-page'>
        <PageHeader
          eyebrow='案件管理 / 案件详情'
          title='案件不存在'
          subtitle='指定案件不存在或当前账号无权访问。'
          actions={<Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/case')}>返回案件清单</Button>}
        />
      </div>
    )
  }

  const metricsForCase = [
    { icon: <FileTextOutlined />, label: '案件编号', value: textValue(caseDetail.case_number || caseDetail.id), delta: '正式 API', tone: 'blue' as Tone },
    { icon: <TeamOutlined />, label: '客户', value: textValue(client.name || caseDetail.client_name), delta: '客户资料', tone: 'teal' as Tone },
    { icon: <UserOutlined />, label: '负责律师', value: textValue(lawyer.name || caseDetail.lawyer_name), delta: '团队分配', tone: 'green' as Tone },
    { icon: <FileSearchOutlined />, label: '优先级', value: textValue(caseDetail.priority, 'medium'), delta: '风险跟踪', tone: (String(caseDetail.priority).toLowerCase() === 'high' ? 'red' : 'orange') as Tone },
  ]
  const currentCaseStatus = textValue(caseDetail.status, '').toLowerCase()
  const caseStageDefinitions = [
    { label: '接案准备', statuses: ['draft', 'pending', 'todo', '待处理'] },
    { label: '冲突复核', statuses: ['risk_review', 'conflict_ready', 'conflict_checking'] },
    { label: '接案审批', statuses: ['submitted', 'under_review', 'approval_pending'] },
    { label: '办理中', statuses: ['active', 'in_progress', 'open'] },
    { label: '结案归档', statuses: ['completed', 'archived', 'closed'] },
  ]
  const needsConflictReview = ['draft', 'pending', 'todo', '待处理', 'risk_review', 'conflict_ready', 'conflict_checking'].includes(currentCaseStatus)
  const openConflictReview = () => {
    const params = new URLSearchParams()
    params.set('case_id', textValue(caseDetail.id, ''))
    params.set('case_number', textValue(caseDetail.case_number, ''))
    params.set('case_title', textValue(caseDetail.title, ''))
    navigate(`/conflict?${params.toString()}`)
  }

  return (
    <div className='batch-page case-management-page'>
      <PageHeader
        eyebrow='案件管理 / 案件清单 / 案件详情'
        title={textValue(caseDetail.title, '未命名案件')}
        subtitle={`案件类型：${textValue(caseDetail.case_type)} · 当前状态：${statusLabel(caseDetail.status)}`}
        actions={
          <>
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/case')}>返回案件清单</Button>
            <Button type='primary' onClick={() => navigate('/case/create')}>新建案件</Button>
          </>
        }
      />

      <div className='batch-metric-grid case-metrics'>
        {metricsForCase.map((item) => <MetricCard key={item.label} {...item} />)}
      </div>

      <div className='batch-case-layout'>
        <SectionCard title='案件概览' className='span-2'>
          <DataTable>
            <table>
              <tbody>
                <tr><td>案件名称</td><td>{textValue(caseDetail.title)}</td></tr>
                <tr><td>案件类型</td><td>{textValue(caseDetail.case_type)}</td></tr>
                <tr><td>案件状态</td><td><RiskTag text={statusLabel(caseDetail.status)} /></td></tr>
                <tr><td>优先级</td><td><RiskTag text={textValue(caseDetail.priority, 'medium')} /></td></tr>
                <tr><td>创建时间</td><td>{formatApiDate(caseDetail.created_at)}</td></tr>
                <tr><td>更新时间</td><td>{formatApiDate(caseDetail.updated_at)}</td></tr>
                <tr><td>案件描述</td><td>{textValue(caseDetail.description, '暂无案件描述')}</td></tr>
              </tbody>
            </table>
          </DataTable>
        </SectionCard>

        {needsConflictReview && (
          <SectionCard title='下一步操作'>
            <div className='batch-advice'>
              <strong>下一步：利益冲突复核</strong>
              <p>当前案件仍处于{statusLabel(caseDetail.status)}阶段。请先完成冲突复核，再进入接案审批或正式办理。</p>
              <Button type='primary' block icon={<FileSearchOutlined />} onClick={openConflictReview}>
                进入本案冲突复核
              </Button>
              <Button block onClick={() => navigate('/case/create')}>
                补充立案信息并重新检测
              </Button>
            </div>
          </SectionCard>
        )}

        <SectionCard title='办理阶段'>
          <div className='batch-case-funnel'>
            {caseStageDefinitions.map((stage) => {
              const isCurrent = stage.statuses.includes(currentCaseStatus)
              return (
              <p key={stage.label} className={isCurrent ? 'current-stage' : ''}>
                <StatusDot color={isCurrent ? 'teal' : 'blue'} />
                <span>{stage.label}</span>
                {isCurrent && <RiskTag text='当前' />}
              </p>
              )
            })}
          </div>
        </SectionCard>

        <SectionCard title='客户信息'>
          <div className='batch-policy-grid'>
            <article>
              <TeamOutlined />
              <div>
                <strong>{textValue(client.name || caseDetail.client_name)}</strong>
                <p>{textValue(client.email)}</p>
                <p>{textValue(client.phone)}</p>
              </div>
            </article>
          </div>
        </SectionCard>

        <SectionCard title='负责团队'>
          <div className='batch-policy-grid'>
            <article>
              <UserOutlined />
              <div>
                <strong>{textValue(lawyer.name || caseDetail.lawyer_name)}</strong>
                <p>{roleLabel(textValue(lawyer.role, 'lawyer'))}</p>
                <p>{textValue(lawyer.email)}</p>
              </div>
            </article>
          </div>
        </SectionCard>

        <SectionCard title='相关案件' className='span-2'>
          <DataTable>
            <table>
              <thead>
                <tr>
                  <th>案件名称</th>
                  <th>客户</th>
                  <th>类型</th>
                  <th>状态</th>
                  <th>负责人</th>
                  <th>更新时间</th>
                </tr>
              </thead>
              <tbody>
                {relatedRows.map((row) => (
                  <tr key={row.id || row.case_number}>
                    <td>{textValue(row.title)}</td>
                    <td>{textValue(row.client_name)}</td>
                    <td>{textValue(row.case_type)}</td>
                    <td><RiskTag text={statusLabel(row.status)} /></td>
                    <td>{textValue(row.lawyer_name)}</td>
                    <td>{formatApiDate(row.updated_at)}</td>
                  </tr>
                ))}
                {relatedRows.length === 0 && <tr><td colSpan={6}>暂无同客户相关案件</td></tr>}
              </tbody>
            </table>
          </DataTable>
        </SectionCard>
      </div>
    </div>
  )
}

export function CaseIntakeWorkbench() {
  const navigate = useNavigate()
  const [form, setForm] = React.useState<IntakeFormState>(() => loadCaseIntakeDraft())
  const [runtime, setRuntime] = React.useState<IntakeRuntimeState>({ apiTimings: [] })
  const [submitting, setSubmitting] = React.useState(false)
  const [submissionNotice, setSubmissionNotice] = React.useState('')
  const [activeStep, setActiveStep] = React.useState(0)
  const [relatedParties, setRelatedParties] = React.useState<Array<{ name: string; role: string }>>([])
  const [relatedPartyDraft, setRelatedPartyDraft] = React.useState('')
  const [caseTags, setCaseTags] = React.useState<string[]>([])
  const [tagDraft, setTagDraft] = React.useState('')
  const [clientOptions, setClientOptions] = React.useState<ClientOption[]>([])
  const [lawyerOptions, setLawyerOptions] = React.useState<LawyerOption[]>([])

  React.useEffect(() => {
    window.localStorage.setItem(caseIntakeDraftKey, JSON.stringify(form))
  }, [form])

  React.useEffect(() => {
    let mounted = true
    Promise.all([
      apiRequest<any>('/clients?page=1&page_size=50'),
      apiRequest<any>('/lawfirm/lawyers?page=1&page_size=50'),
    ])
      .then(([clientData, lawyerData]) => {
        if (!mounted) return

        const clients = normalizeClientOptions(clientData)
        const lawyers = normalizeLawyerOptions(lawyerData)
        setClientOptions(clients)
        setLawyerOptions(lawyers)

        setForm((current) => ({ ...current }))
      })
      .catch((error) => {
        message.error(error instanceof Error ? error.message : '加载客户或律师失败')
      })

    return () => {
      mounted = false
    }
  }, [])

  const recordTiming = (label: string, startedAt: number) => {
    setRuntime((current) => ({
      ...current,
      apiTimings: [
        { label, duration: Math.round(performance.now() - startedAt), at: new Date().toLocaleTimeString() },
        ...current.apiTimings,
      ].slice(0, 5),
    }))
  }

  const updateForm = (key: keyof IntakeFormState, value: string | number) => {
    setForm((current) => ({ ...current, [key]: value }))
  }

  const updateClient = (clientId: number) => {
    const selectedClient = clientOptions.find((client) => client.id === clientId)
    setForm((current) => ({
      ...current,
      clientId,
      clientName: selectedClient?.name || current.clientName,
    }))
  }

  const createIntake = async () => {
    if (!form.clientId || !form.clientName || !form.lawyerId) {
      throw new Error('请先选择数据库中的客户和负责律师')
    }
    const startedAt = performance.now()
    const intake = await apiRequest<any>('/case-intakes', {
      method: 'POST',
      body: JSON.stringify({
        client_id: form.clientId,
        title: form.title,
        case_type: form.caseType,
        status: 'conflict_ready',
        priority: form.priority,
        description: form.description,
        metadata: {
          source: 'batch01_real_api',
          billing_method: form.billingMethod,
          lawyer_id: form.lawyerId,
        },
        parties: [
          { entity_name: form.clientName, entity_type: 'company', party_role: 'client', relation_depth: 0 },
          { entity_name: form.opponentName, entity_type: 'company', party_role: 'opposing_party', relation_depth: 0 },
          ...relatedParties.map((party) => ({
            entity_name: party.name,
            entity_type: party.role.includes('自然人') ? 'person' : 'company',
            party_role: 'related_party',
            relation_depth: 1,
          })),
        ],
        materials: defaultMaterials,
      }),
    })
    recordTiming('创建接案', startedAt)
    setRuntime((current) => ({ ...current, intake }))
    message.success(`接案草稿已创建：${intake.intake_code}`)
    return intake
  }

  const runConflictCheck = async (intake = runtime.intake) => {
    if (!form.clientId || !form.clientName || !form.lawyerId) {
      throw new Error('请先选择数据库中的客户和负责律师')
    }
    const startedAt = performance.now()
    const conflict = await apiRequest<any>('/conflict/check', {
      method: 'POST',
      body: JSON.stringify({
        clientId: String(form.clientId),
        clientName: form.clientName,
        clientType: 'COMPANY',
        otherParties: [form.opponentName, ...relatedParties.map((party) => party.name)],
        parties: [
          { role: 'CLIENT', name: form.clientName, entityType: 'COMPANY' },
          { role: 'OPPOSING_PARTY', name: form.opponentName, entityType: 'COMPANY' },
          ...relatedParties.map((party) => ({
            role: party.role || 'RELATED_PARTY',
            name: party.name,
            entityType: 'COMPANY',
          })),
        ],
        caseName: form.title,
        caseType: form.caseType,
        searchYears: 5,
        includeCorporateRelations: true,
        searchDepth: 'STANDARD',
        userId: String(form.lawyerId),
      }),
    })
    recordTiming('冲突检查', startedAt)
    setRuntime((current) => ({ ...current, intake, conflict }))
    setSubmissionNotice('')
    message.success('利益冲突检查已完成')
    return conflict
  }

  const submitApproval = async () => {
    setSubmitting(true)
    try {
      if (missingSubmitFields.length > 0) {
        const notice = `以下必填项未完成：${missingSubmitFields.join('、')}。请补充后再提交审批。`
        setSubmissionNotice(notice)
        message.error(notice)
        return
      }
      const intake = runtime.intake || await createIntake()
      const conflict = runtime.conflict || await runConflictCheck(intake)
      const overallRisk = textValue(conflict?.riskAssessment?.overallRisk || conflict?.risk_level || conflict?.record?.risk_level, 'LOW').toUpperCase()
      if (['HIGH', 'CRITICAL'].includes(overallRisk)) {
        const notice = '检测到高风险或严重冲突，需先在利益冲突检测清单中发起冲突审批或豁免评估，暂不能提交立案审批。'
        setSubmissionNotice(notice)
        message.error(notice)
        return
      }
      const startedAt = performance.now()
      const approval = await apiRequest<any>('/integration/approvals/with-conflict', {
        method: 'POST',
        body: JSON.stringify({
          type: 'case_creation',
          title: `新建案件审批 - ${form.title}`,
          content: form.description,
          applicant_name: '当前登录用户',
          department_name: '公司业务部',
          urgency: 'urgent',
          priority: form.priority,
          workflow_type: 'CASE_APPROVAL',
          expected_duration: 3,
          category: 'case_intake',
          metadata: {
            intake_id: intake.id,
            intake_code: intake.intake_code,
            client: { id: form.clientId, name: form.clientName },
            case_creation_config: {
              title: form.title,
              description: form.description,
              client_id: form.clientId,
              case_type: dbCaseType(form.caseType),
              priority: form.priority,
              lawyer_id: form.lawyerId,
              billing_method: form.billingMethod,
            },
            parties: [
              { role: 'client', name: form.clientName },
              { role: 'opposing_party', name: form.opponentName },
              ...relatedParties.map((party) => ({ role: 'related_party', name: party.name, party_type: party.role })),
            ],
            materials: defaultMaterials,
            tags: caseTags,
            conflict_result: conflict,
          },
          case_creation_config: {
            title: form.title,
            description: form.description,
            client_id: form.clientId,
            case_type: dbCaseType(form.caseType),
            priority: form.priority,
            lawyer_id: form.lawyerId,
            billing_method: form.billingMethod,
            is_major_risk: conflict?.riskAssessment?.overallRisk === 'HIGH',
          },
        }),
      })
      recordTiming('提交审批', startedAt)
      setRuntime((current) => ({ ...current, approval }))
      setSubmissionNotice('')
      window.localStorage.removeItem(caseIntakeDraftKey)
      message.success('已提交真实审批')
      navigate(`/approval/${approval.approval_id}`)
    } catch (error) {
      message.error(error instanceof Error ? error.message : '提交审批失败')
    } finally {
      setSubmitting(false)
    }
  }

  const selectedLawyer = lawyerOptions.find((lawyer) => lawyer.id === form.lawyerId)
  const missingSubmitFields = [
    !form.title && '案件名称',
    !form.caseType && '案件类型',
    (!form.clientId || !form.clientName) && '客户',
    !form.opponentName && '对方当事人',
    !form.description && '案情摘要',
    !form.lawyerId && '负责律师',
    !runtime.conflict && '利益冲突检查',
  ].filter(Boolean) as string[]
  const completedSubmitFields = 7 - missingSubmitFields.length
  const intakeCompleteness = Math.round((completedSubmitFields / 7) * 100)
  const overviewHint = missingSubmitFields.length > 0
    ? `还需补充：${missingSubmitFields.join('、')}`
    : '必填信息已完成'
  const intakeSteps = [
    { title: '基本信息', desc: '案件与当事人信息' },
    { title: '利益冲突检查', desc: '自动检测与人工复核' },
    { title: '团队与费用', desc: '团队指派与收费安排' },
    { title: '文档与材料', desc: '材料清单与附件上传' },
    { title: '立案提交', desc: '提交审批并创建案件' },
  ]
  const currentStep = intakeSteps[activeStep] || intakeSteps[0]
  const currentConflictRisk = textValue(runtime.conflict?.riskAssessment?.overallRisk || runtime.conflict?.risk_level || runtime.conflict?.record?.risk_level, '').toUpperCase()
  const hasBlockingConflictRisk = ['HIGH', 'CRITICAL'].includes(currentConflictRisk)
  const conflictRiskText = runtime.conflict
    ? riskLabel(currentConflictRisk || 'LOW')
    : '未检测'
  const conflictStatusClass = runtime.conflict ? 'success-text' : 'danger-text'

  const addRelatedParty = () => {
    const name = relatedPartyDraft.trim() || `新增相关方 ${relatedParties.length + 1}`
    setRelatedParties((current) => [...current, { name, role: '待确认' }])
    setRelatedPartyDraft('')
    message.success('已添加相关方，将纳入冲突检索')
  }

  const addCaseTag = () => {
    const name = tagDraft.trim() || `新标签${caseTags.length + 1}`
    if (caseTags.includes(name)) {
      message.warning('标签已存在')
      return
    }
    setCaseTags((current) => [...current, name])
    setTagDraft('')
    message.success('已添加标签')
  }

  const handleCreateIntake = async () => {
    try {
      await createIntake()
    } catch (error) {
      message.error(error instanceof Error ? error.message : '暂存失败')
    }
  }

  const handleRunConflictCheck = async () => {
    const missing: string[] = []
    if (!form.title) missing.push('案件名称')
    if (!form.clientId || !form.clientName) missing.push('客户')
    if (!form.opponentName) missing.push('对方当事人')
    if (!form.lawyerId) missing.push('负责律师')
    if (!form.caseType) missing.push('案件类型')
    if (missing.length > 0) {
      message.error(`以下必填项未完成：${missing.join('、')}，请补充后再运行冲突检查`)
      return
    }
    try {
      await runConflictCheck(runtime.intake || await createIntake())
      setActiveStep(1)
      message.success('利益冲突检查已完成，当前草稿：' + (form.title || '未命名案件'))
    } catch (error) {
      message.error(error instanceof Error ? error.message : '冲突检查失败，请检查网络后重试')
    }
  }

  const handleCancelIntake = () => {
    Modal.confirm({
      title: '取消立案',
      content: '当前未提交的信息将不会进入正式案件流程。',
      okText: '返回案件清单',
      cancelText: '继续编辑',
      onOk: () => navigate('/case'),
    })
  }

  const handleSaveDraftAndExit = async () => {
    try {
      await createIntake()
      navigate('/case')
    } catch (error) {
      message.error(error instanceof Error ? error.message : '保存并退出失败')
    }
  }

  const showMaterialsUnavailable = () => {
    Modal.info({
      title: '文件材料归档暂未开放',
      content: '当前 MVP 试用版暂不跳转文档中心。你可以继续保留本次立案上下文，已填写内容不会丢失。',
      okText: '继续立案',
    })
  }

  return (
    <div className='batch-page intake-page'>
      <PageHeader
        eyebrow='案件管理 / 新建案件 / 立案工作台'
        title='新建案件立案工作台'
        actions={
          <>
            <span className='batch-autosave'>正式 API：{runtime.apiTimings[0] ? `${runtime.apiTimings[0].label} ${runtime.apiTimings[0].duration}ms` : '待调用'}</span>
            <Button onClick={handleCreateIntake}>暂存</Button>
            <Button type='primary' loading={submitting} onClick={submitApproval}>保存并提交审批</Button>
          </>
        }
      />

      <div className='batch-stepper'>
        {intakeSteps.map((step, index) => (
          <button
            type='button'
            key={step.title}
            className={index === activeStep ? 'active' : ''}
            onClick={() => setActiveStep(index)}
            onKeyDown={(event) => {
              if (event.key === 'Enter' || event.key === ' ') {
                event.preventDefault()
                setActiveStep(index)
              }
            }}
          >
            <span>{index + 1}</span>
            <strong>{step.title}</strong>
            <em>{step.desc}</em>
          </button>
        ))}
      </div>

      <div className='batch-intake-layout'>
        <main>
          {activeStep === 0 && (
            <>
          <SectionCard title='案件基本信息'>
            <div className='batch-tabs compact'>
              <button className='active'>{currentStep.title}</button>
              <button onClick={() => setActiveStep((current) => Math.max(0, current - 1))} disabled={activeStep === 0}>上一步</button>
              <button onClick={() => setActiveStep((current) => Math.min(intakeSteps.length - 1, current + 1))} disabled={activeStep === intakeSteps.length - 1}>下一步</button>
            </div>
            <div className='batch-form-grid four'>
              <div className='batch-field'><span>案件名称 *</span><Input value={form.title} onChange={(event) => updateForm('title', event.target.value)} /></div>
              <div className='batch-field'><span>案件类型 *</span><Select value={form.caseType} onChange={(value) => updateForm('caseType', value)} options={[{ value: 'commercial', label: '商事诉讼' }, { value: 'ma', label: '并购重组' }]} /></div>
              <div className='batch-field'><span>案件阶段 *</span><Select value='潜在受理' options={[{ value: '潜在受理' }]} /></div>
              <div className='batch-field'><span>预计争议金额</span><Input placeholder='请输入预计争议金额' /></div>
              <div className='batch-field'><span>业务领域 *</span><Select placeholder='请选择业务领域' options={[{ value: '公司与并购' }]} /></div>
              <div className='batch-field'><span>子领域 *</span><Select placeholder='请选择子领域' options={[{ value: '投资与融资' }]} /></div>
              <div className='batch-field'><span>案源渠道</span><Select placeholder='请选择案源渠道' options={[{ value: '现有客户介绍' }]} /></div>
              <div className='batch-field'><span>案源联系人</span><Select placeholder='请选择或录入案源联系人' options={[]} /></div>
            </div>
            <div className='batch-wide-label'>
              <span>案情摘要 *</span>
              <Input.TextArea rows={3} value={form.description} onChange={(event) => updateForm('description', event.target.value)} />
            </div>
            <div className='batch-form-grid five'>
              {['投资协议签署日', '争议发生日', '对方违约日', '拟起诉日', '诉讼地'].map((item) => (
                <div className='batch-field' key={item}><span>{item}</span><Input placeholder={`请输入${item}`} /></div>
              ))}
            </div>
          </SectionCard>

          <SectionCard title='当事人信息'>
            <div className='batch-party-grid'>
              <article>
                <strong>我方当事人（客户） *</strong>
                <div className='batch-party-card green'>
                  <TeamOutlined />
                  <div>
                    <Select
                      value={form.clientId || undefined}
                      onChange={updateClient}
                      showSearch
                      optionFilterProp='label'
                      placeholder='选择数据库客户'
                      options={clientOptions.map((client) => ({ value: client.id, label: client.name }))}
                      style={{ minWidth: 240 }}
                    />
                    <RiskTag text='现有客户' />
                    <p>客户 ID：{form.clientId || '未选择'}<br />正式 API 将用该客户创建案件</p>
                  </div>
                </div>
              </article>
              <span className='vs-badge'>VS</span>
              <article>
                <strong>对方当事人（被告/对方） *</strong>
                <div className='batch-party-card blue'>
                  <BankOutlined />
                  <div>
                    <Input
                      value={form.opponentName}
                      onChange={(event) => updateForm('opponentName', event.target.value)}
                      placeholder='输入对方当事人名称'
                    />
                    <RiskTag text='新增对方' />
                    <p>用于真实冲突检查 otherParties</p>
                  </div>
                </div>
              </article>
              <article>
                <strong>其他相关方（可选）</strong>
                <div className='batch-related-party'>
                  {relatedParties.map((party, index) => (
                    <p key={`${party.name}-${index}`}>
                      {party.name} <RiskTag text={party.role} />
                      <Button type='link' size='small' onClick={() => setRelatedParties((current) => current.filter((_, itemIndex) => itemIndex !== index))}>移除</Button>
                    </p>
                  ))}
                  <div className='batch-inline-add'>
                    <Input value={relatedPartyDraft} onChange={(event) => setRelatedPartyDraft(event.target.value)} placeholder='输入关联公司、实控人、保证人' />
                    <Button type='link' icon={<PlusOutlined />} onClick={addRelatedParty}>添加相关方</Button>
                  </div>
                </div>
              </article>
            </div>
            <div className='batch-tags'>
              {caseTags.map((tag) => (
                <Tag key={tag} closable onClose={() => setCaseTags((current) => current.filter((item) => item !== tag))}>{tag}</Tag>
              ))}
              <Input size='small' value={tagDraft} onChange={(event) => setTagDraft(event.target.value)} placeholder='案由/行业/风险标签' onPressEnter={addCaseTag} />
              <Button size='small' icon={<PlusOutlined />} onClick={addCaseTag}>添加标签</Button>
            </div>
          </SectionCard>
            </>
          )}

          {activeStep === 1 && (
            <SectionCard title='利益冲突检查'>
              <div className='batch-approval-risk'>
                <SafetyCertificateOutlined />
                <div>
                  <strong>检查状态：{runtime.conflict ? `已完成 · ${conflictRiskText}` : '待检测'}</strong>
                  <p>冲突检索会覆盖客户、对方当事人和其他相关方，并冻结到后续审批快照。</p>
                </div>
              </div>
              <div className='batch-intake-draft-summary' style={{ background: '#f6ffed', border: '1px solid #b7eb8f', borderRadius: 6, padding: '8px 12px', marginBottom: 12 }}>
                <strong>本次立案草稿</strong>
                {runtime.intake?.intake_code && <span style={{ marginLeft: 8 }}>接案草稿已创建：{runtime.intake.intake_code}</span>}
                <span style={{ marginLeft: 8 }}>案件：{form.title || '未填写'}</span>
                <span style={{ marginLeft: 8 }}>客户：{form.clientName || '未选择'}</span>
                <span style={{ marginLeft: 8 }}>对方：{form.opponentName || '未填写'}</span>
              </div>
              <DataTable>
                <table>
                  <tbody>
                    <tr><td>客户</td><td>{form.clientName || '未选择'}</td></tr>
                    <tr><td>对方当事人</td><td>{form.opponentName || '未填写'}</td></tr>
                    <tr><td>其他相关方</td><td>{relatedParties.map((party) => party.name).join('、') || '暂无'}</td></tr>
                    <tr><td>检测结果</td><td>{runtime.conflict ? `风险等级：${conflictRiskText}，评分：${runtime.conflict.riskAssessment?.riskScore || 0}` : '尚未调用正式 API'}</td></tr>
                    <tr><td>检测 ID</td><td>{runtime.conflict?.checkId || runtime.conflict?.record?.check_id || '检测后生成'}</td></tr>
                    <tr><td>提交提示</td><td>{submissionNotice || '完成必填项且无高风险/严重冲突后可提交立案审批'}</td></tr>
                  </tbody>
                </table>
              </DataTable>
              <Space>
                <Button icon={<ArrowLeftOutlined />} onClick={() => setActiveStep(0)}>返回基本信息</Button>
                <Button type='primary' icon={<SafetyCertificateOutlined />} onClick={handleRunConflictCheck}>运行利益冲突检查</Button>
                <Button onClick={() => setActiveStep(2)} disabled={!runtime.conflict}>进入团队与费用</Button>
              </Space>
            </SectionCard>
          )}

          {activeStep === 2 && (
          <SectionCard title='团队与费用'>
            <div className='batch-form-grid four'>
              <div className='batch-field'>
                <span>负责律师 *</span>
                <Select
                  value={form.lawyerId || undefined}
                  onChange={(value) => updateForm('lawyerId', value)}
                  showSearch
                  optionFilterProp='label'
                  placeholder='选择负责律师'
                  options={lawyerOptions.map((lawyer) => ({
                    value: lawyer.id,
                    label: `${lawyer.name}${lawyer.department ? ` · ${lawyer.department}` : ''}`,
                  }))}
                />
              </div>
              <div className='batch-field'><span>计费方式</span><Select value={form.billingMethod} onChange={(value) => updateForm('billingMethod', value)} options={[{ value: 'hourly', label: '小时计费' }, { value: 'fixed', label: '固定收费' }, { value: 'contingency', label: '风险代理' }]} /></div>
              <div className='batch-field'><span>收费基数</span><Select placeholder='请选择收费基数' options={[{ value: '争议标的金额' }]} /></div>
              <div className='batch-field'><span>风险代理比例</span><Input placeholder='请输入风险代理比例' /></div>
              <div className='batch-field'><span>最低收费保障</span><Input placeholder='请输入最低收费保障' /></div>
            </div>
          </SectionCard>
          )}

          {activeStep === 3 && (
            <SectionCard title='文档与材料'>
              <DataTable>
                <table>
                  <thead><tr><th>材料名称</th><th>类型</th><th>状态</th><th>操作</th></tr></thead>
                  <tbody>
                    {defaultMaterials.map((item) => (
                      <tr key={item.name}><td>{item.name}</td><td>{item.material_type}</td><td><RiskTag text={item.status} /></td><td><Button type='link' onClick={showMaterialsUnavailable}>进入文件管理</Button></td></tr>
                    ))}
                  </tbody>
                </table>
              </DataTable>
              <Button icon={<CloudUploadOutlined />} onClick={showMaterialsUnavailable}>打开文件材料归档</Button>
            </SectionCard>
          )}

          {activeStep === 4 && (
            <SectionCard title='立案提交确认'>
              <div className='batch-info-grid two'>
                {[
                  `案件名称 ${form.title}`,
                  `客户 ${form.clientName || '未选择'}`,
                  `对方当事人 ${form.opponentName}`,
                  `负责律师 ${selectedLawyer?.name || '未选择'}`,
                  `冲突检查 ${runtime.conflict ? conflictRiskText : '未检测'}`,
                  `标签 ${caseTags.join('、') || '暂无'}`,
                ].map((line) => (
                  <p key={line}><span>{line.split(' ')[0]}</span><strong>{line.substring(line.indexOf(' ') + 1)}</strong></p>
                ))}
              </div>
              <Space>
                <Button onClick={() => setActiveStep(3)}>返回材料</Button>
                <Button onClick={handleCreateIntake}>保存草稿</Button>
                <Button type='primary' loading={submitting} onClick={submitApproval}>提交审批并等待成案</Button>
              </Space>
              {submissionNotice && (
                <div className='batch-advice danger' style={{ marginTop: 12 }}>
                  <strong>暂不能提交审批</strong>
                  <p>{submissionNotice}</p>
                  {hasBlockingConflictRisk && (
                    <Button type='primary' danger onClick={() => navigate('/conflict')}>进入冲突清单发起复核</Button>
                  )}
                </div>
              )}
            </SectionCard>
          )}
        </main>

        <aside className='batch-intake-aside'>
          <SectionCard title='案件概览'>
            <div className='batch-overview-score'>
              <Progress type='circle' percent={intakeCompleteness} size={92} strokeColor='#12a89d' />
              <div><strong>信息完整度</strong><span>{overviewHint}</span></div>
            </div>
            {[
              `案件名称 ${form.title || '未填写'}`,
              `业务领域 ${form.caseType ? dbCaseType(form.caseType) : '未选择'}`,
              '案件阶段 潜在受理',
              '统一编号 待生成',
              `创建人 ${getUserInfo()?.name || getUserInfo()?.username || '当前用户'}`,
            ].map((line) => <p key={line}>{line}</p>)}
            {runtime.intake && <p>接案编号 <strong>{runtime.intake.intake_code}</strong></p>}
            {runtime.conflict && <p>冲突风险 <strong>{runtime.conflict.riskAssessment?.overallRisk || 'LOW'}</strong></p>}
          </SectionCard>
          <SectionCard title='团队指派（预览）'>
            <Select
              value={form.lawyerId || undefined}
              onChange={(value) => updateForm('lawyerId', value)}
              showSearch
              optionFilterProp='label'
              placeholder='选择负责律师'
              options={lawyerOptions.map((lawyer) => ({
                value: lawyer.id,
                label: `${lawyer.name}${lawyer.department ? ` · ${lawyer.department}` : ''}`,
              }))}
              style={{ width: '100%', marginBottom: 12 }}
            />
            <div className='batch-team-line'>
              <Avatar size='small' icon={<UserOutlined />} />
              负责人：{selectedLawyer?.name || '未选择'}（{selectedLawyer?.position || selectedLawyer?.seniority || '律师'}）
            </div>
          </SectionCard>
          <SectionCard title='正式 API 耗时'>
            {runtime.apiTimings.length === 0 ? <p>尚未调用正式 API</p> : runtime.apiTimings.map((item) => (
              <p key={`${item.label}-${item.at}`}>{item.label} <strong>{item.duration}ms</strong> <span>{item.at}</span></p>
            ))}
          </SectionCard>
        </aside>
      </div>

      <div className='batch-bottom-bar'>
        <Button onClick={handleCancelIntake}>取消立案</Button>
        <span><SafetyCertificateOutlined /> 利益冲突检查状态：<strong className={conflictStatusClass}>{conflictRiskText}</strong></span>
        {runtime.intake?.intake_code && <span>接案草稿已创建：<strong>{runtime.intake.intake_code}</strong></span>}
        {submissionNotice && <span className='danger-text'>{submissionNotice}</span>}
        {submissionNotice && hasBlockingConflictRisk && (
          <Button size='small' danger onClick={() => navigate('/conflict')}>进入冲突清单发起复核</Button>
        )}
        <Space>
          <Button onClick={handleCreateIntake}>保存草稿</Button>
          <Button onClick={handleSaveDraftAndExit}>保存并退出</Button>
          <Button icon={<SafetyCertificateOutlined />} onClick={handleRunConflictCheck}>保存并进行利益冲突检查</Button>
          <Button type='primary' loading={submitting} onClick={submitApproval}>提交审批并等待成案</Button>
        </Space>
      </div>
    </div>
  )
}

export function ConflictCheckResults() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const [commandCenter, setCommandCenter] = React.useState<CommandCenterPayload | null>(null)
  const [loading, setLoading] = React.useState(false)
  const [selectedConflict, setSelectedConflict] = React.useState<CommandCenterRiskItem | null>(null)
  const [creatingApproval, setCreatingApproval] = React.useState(false)

  React.useEffect(() => {
    const controller = new AbortController()
    setLoading(true)
    fetchCommandCenter(controller.signal)
      .then((data) => setCommandCenter(data))
      .catch((error: unknown) => {
        if ((error as DOMException).name !== 'AbortError') {
          message.error('加载冲突检测数据失败')
        }
      })
      .finally(() => setLoading(false))
    return () => controller.abort()
  }, [])

  const riskItems = listOf(commandCenter?.risk_queue)
  const contextCaseID = searchParams.get('case_id') || ''
  const contextCaseNumber = searchParams.get('case_number') || ''
  const contextCaseTitle = searchParams.get('case_title') || ''
  const contextConflict = riskItems.find((item) => conflictMatchesCaseContext(item, contextCaseID, contextCaseNumber, contextCaseTitle))
  const criticalRiskCount = riskItems.filter((item) => (item.risk_level || '').toUpperCase() === 'CRITICAL').length
  const highRiskCount = riskItems.filter((item) => ['HIGH', 'CRITICAL'].includes((item.risk_level || '').toUpperCase())).length
  const mediumRiskCount = riskItems.filter((item) => (item.risk_level || '').toUpperCase() === 'MEDIUM').length
  const lowRiskCount = riskItems.filter((item) => (item.risk_level || '').toUpperCase() === 'LOW').length
  const queueRisk = criticalRiskCount > 0 ? 'CRITICAL' : highRiskCount > 0 ? 'HIGH' : mediumRiskCount > 0 ? 'MEDIUM' : lowRiskCount > 0 ? 'LOW' : ''
  const selectedCheckResult = recordValue(selectedConflict?.check_result)
  const selectedRiskAssessment = recordValue(selectedCheckResult.riskAssessment)
  const selectedRiskReason = textValue(
    selectedRiskAssessment.riskReason ||
    selectedCheckResult.riskReason ||
    selectedCheckResult.risk_reason ||
    selectedConflict?.conflict_details ||
    selectedConflict?.description,
    '暂无风险原因',
  )
  const selectedRequiresApproval = Boolean(
    selectedRiskAssessment.requiresApproval ??
    selectedCheckResult.requiresApproval ??
    ['HIGH', 'CRITICAL'].includes(textValue(selectedConflict?.risk_level, '').toUpperCase()),
  )
  const selectedStatistics = recordValue(selectedCheckResult.checkStatistics)
  const selectedSearchParameters = recordValue(selectedConflict?.search_parameters)
  const selectedConflictCases = listOf<Record<string, unknown>>(selectedConflict?.conflict_cases)
  const exportConflictReport = () => {
    if (riskItems.length === 0) {
      message.warning('暂无可导出的冲突检测记录')
      return
    }
    const report = {
      title: '利益冲突检测结果报告',
      generatedAt: new Date().toISOString(),
      source: 'conflict_check_records',
      summary: {
        total: riskItems.length,
        highRisk: highRiskCount,
        mediumRisk: mediumRiskCount,
        lowRisk: lowRiskCount,
      },
      records: riskItems,
    }
    const blob = new Blob([JSON.stringify(report, null, 2)], {
      type: 'application/json;charset=utf-8',
    })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `conflict-results-${new Date().toISOString().slice(0, 10)}.json`
    document.body.appendChild(link)
    link.click()
    link.remove()
    URL.revokeObjectURL(url)
    message.success('检测报告已导出')
  }

  const createConflictApproval = async (item?: CommandCenterRiskItem | null) => {
    if (!item?.id) {
      message.warning('请选择一条冲突检测记录')
      return
    }

    const userInfo = getUserInfo() || {}
    setCreatingApproval(true)
    try {
      const approval = await apiRequest<any>(`/conflict/tasks/${item.id}/approval`, {
        method: 'POST',
        body: JSON.stringify({
          title: `冲突审查审批 - ${textValue(item.title, textValue(item.id))}`,
          content: `客户 ${textValue(item.client_name, '未登记客户')} 的利益冲突检测结果为 ${riskLabel(item.risk_level)}，请合规复核。`,
          applicant_id: textValue(userInfo.id, '1'),
          applicant_name: textValue(userInfo.name || userInfo.username, '当前用户'),
          department_name: textValue(userInfo.department, '合规风控部'),
          priority: ['HIGH', 'CRITICAL'].includes(textValue(item.risk_level, '').toUpperCase()) ? 'high' : 'medium',
          current_approver_id: textValue(userInfo.id, '1'),
          current_approver_name: textValue(userInfo.name || userInfo.username, '合规负责人'),
        }),
      })
      message.success(`已创建冲突审批：${textValue(approval.request_number, approval.approval_id)}`)
      setSelectedConflict(null)
      navigate(`/approval/${approval.approval_id}`)
    } catch (error) {
      message.error(error instanceof Error ? error.message : '创建冲突审批失败')
    } finally {
      setCreatingApproval(false)
    }
  }

  return (
    <div className='batch-page conflict-page'>
      <PageHeader
        eyebrow='利益冲突 / 冲突检测 / 检测清单'
        title='利益冲突检测清单'
        subtitle={`来自正式 API 的冲突任务队列，最近刷新：${formatDateTime(commandCenter?.generated_at)}`}
        actions={
          <>
            <Button icon={<DownloadOutlined />} onClick={exportConflictReport}>导出检测报告</Button>
            <Button type='primary' loading={loading || creatingApproval} disabled={!selectedConflict?.id} onClick={() => createConflictApproval(selectedConflict)}>创建冲突审核</Button>
          </>
        }
      />

      {(contextCaseID || contextCaseNumber || contextCaseTitle) && (
        <SectionCard title='本案复核上下文'>
          <div className='batch-advice'>
            <strong>{contextCaseNumber || contextCaseID || '当前案件'} {contextCaseTitle}</strong>
            {contextConflict ? (
              <>
                <p>已匹配到本案冲突检测记录。请点击查看结果后再创建冲突审核。</p>
                <Button type='primary' onClick={() => setSelectedConflict(contextConflict)}>查看本案检测结果</Button>
              </>
            ) : (
              <p>暂未在冲突任务队列中匹配到本案检测记录。请确认是否已从立案工作台运行利益冲突检查；如未检测，请返回补充立案信息并重新检测。</p>
            )}
            {!contextConflict && (
              <Space wrap>
                <Button onClick={() => navigate(`/case/${contextCaseID}`)} disabled={!contextCaseID}>返回案件详情</Button>
                <Button type='primary' onClick={() => navigate('/case/create')}>补充立案信息并检测</Button>
              </Space>
            )}
          </div>
        </SectionCard>
      )}

      <div className='batch-conflict-top'>
        <SectionCard title='检测任务概览'>
          <div className='batch-info-grid compact'>
            {[
              `检测任务 ${riskItems.length} 条`,
              `高风险 ${highRiskCount} 条`,
              `中风险 ${mediumRiskCount} 条`,
              `低风险 ${lowRiskCount} 条`,
              `待处理审批 ${commandCenter?.summary?.pending_approvals ?? 0} 条`,
              `最近刷新 ${formatDateTime(commandCenter?.generated_at)}`,
            ].map((line) => (
              <p key={line}><span>{line.split(' ')[0]}</span><strong>{line.substring(line.indexOf(' ') + 1)}</strong></p>
            ))}
          </div>
        </SectionCard>
        <section className='batch-risk-banner'>
          <AlertOutlined />
          <div><span>队列最高风险</span><strong>{riskLabel(queueRisk)}</strong><em>发现 {highRiskCount} 项高风险/严重冲突</em></div>
          <div><span>任务数量</span><strong className='score'>{riskItems.length}</strong><em>条</em></div>
          <div className='batch-risk-counts'><p>高风险 <strong>{highRiskCount}</strong></p><p>中风险 <strong>{mediumRiskCount}</strong></p><p>低风险 <strong>{lowRiskCount}</strong></p><p>提示 <strong>{Math.max(0, riskItems.length - highRiskCount - mediumRiskCount - lowRiskCount)}</strong></p></div>
          <div><span>检测来源</span><p>conflict_check_records</p><p>无记录时显示空状态</p><p>不使用前端写死命中</p></div>
        </section>
      </div>

      <div className='batch-conflict-layout'>
        <SectionCard
          title='检测任务清单'
          extra={
            <Space>
              {['全部结果', '高风险', '中风险', '低风险', '提示'].map((tab, index) => <Button key={tab} type={index === 0 ? 'primary' : 'default'} aria-label={`筛选${tab}`}>{tab}</Button>)}
            </Space>
          }
          className='span-2'
        >
          <DataTable>
            <table>
              <thead>
                <tr>
                  <th>风险等级</th><th>命中主体</th><th>命中类型</th><th>命中范围</th><th>置信度</th><th>穿透层级</th><th>证据概要</th><th>来源</th><th>操作</th>
                </tr>
              </thead>
              <tbody>
                {riskItems.map((row) => (
                  <tr key={row.id || row.title} className={['HIGH', 'CRITICAL'].includes((row.risk_level || '').toUpperCase()) ? 'danger-row' : ''}>
                    <td><RiskTag text={riskLabel(row.risk_level)} /></td>
                    <td className='strong-cell'>{textValue(row.matched_subject || row.client_name, '未登记客户')}</td>
                    <td>{textValue(row.matched_type, statusLabel(row.status))}</td>
                    <td>{textValue(row.title, '未关联案件')}</td>
                    <td>{row.status === 'COMPLETED' ? '100%' : '-'}</td>
                    <td>-</td>
                    <td>{textValue(row.evidence_summary || row.title)}</td>
                    <td>conflict_check_records</td>
                    <td>
                      <Button size='small' type='primary' ghost onClick={() => setSelectedConflict(row)}>
                        查看详情
                      </Button>
                    </td>
                  </tr>
                ))}
                {riskItems.length === 0 && <tr><td colSpan={9}>暂无数据库冲突检测记录</td></tr>}
              </tbody>
            </table>
          </DataTable>
        </SectionCard>

        <aside className='batch-conflict-side'>
          <SectionCard title='合规建议'>
            <div className='batch-advice danger'>
              <strong>{highRiskCount > 0 ? '检测到高风险冲突，建议暂停承办' : '未发现高风险冲突记录'}</strong>
              <p>本建议基于数据库中的冲突检测记录生成。若记录为空，请先从接案工作台发起正式冲突检查。</p>
              <Button danger type='primary' block loading={creatingApproval} disabled={!selectedConflict?.id} onClick={() => createConflictApproval(selectedConflict)}>发起冲突审核（高风险）</Button>
              <Button block>申请豁免评估（中/低风险可豁免）</Button>
              <Button block>调整团队律师</Button>
              <Button block>退出案件</Button>
            </div>
          </SectionCard>
          <SectionCard title='豁免可能性评估'>
            <p><StatusDot color='red' />高风险项 {highRiskCount} 项<span className='danger-text'>{highRiskCount > 0 ? '需审批' : '无'}</span></p>
            <p><StatusDot color='orange' />中风险项 {mediumRiskCount} 项<span className='green-text'>{mediumRiskCount > 0 ? '可评估' : '无'}</span></p>
            <Button block>查看豁免评估详情</Button>
          </SectionCard>
          <SectionCard title='检测范围'>
            <div className='batch-scope-grid'>
              <span>冲突记录 {riskItems.length}</span><span>高风险 {highRiskCount}</span><span>中风险 {mediumRiskCount}</span><span>低风险 {lowRiskCount}</span><span>审批待办 {commandCenter?.summary?.pending_approvals ?? 0}</span><span>案件 {commandCenter?.summary?.active_cases ?? 0}</span>
            </div>
          </SectionCard>
        </aside>
      </div>

      <Modal
        title='冲突检测详情'
        open={Boolean(selectedConflict)}
        onCancel={() => setSelectedConflict(null)}
        footer={[
          <Button key='close' onClick={() => setSelectedConflict(null)}>关闭</Button>,
          <Button key='approval' type='primary' loading={creatingApproval} disabled={!selectedConflict?.id} onClick={() => createConflictApproval(selectedConflict)}>
            发起冲突审批
          </Button>,
        ]}
        width={760}
        destroyOnHidden
      >
        {selectedConflict && (
          <div className='batch-conflict-detail'>
            <div className='batch-info-grid two'>
                {[
                  ['检测任务', textValue(selectedConflict.id, '-')],
                  ['关联案件', textValue(selectedConflict.title, '-')],
                  ['案件类型', textValue(selectedConflict.case_type, '-')],
                  ['客户/委托人', textValue(selectedConflict.client_name, '-')],
                  ['主命中主体', textValue(selectedConflict.matched_subject || selectedConflict.client_name, '-')],
                  ['主冲突类型', textValue(selectedConflict.matched_type, '-')],
                  ['检测状态', statusLabel(selectedConflict.status)],
                  ['风险等级', riskLabel(selectedConflict.risk_level)],
                  ['是否冲突', selectedConflict.has_conflict ? '是' : '否'],
                  ['命中数量', String(selectedConflictCases.length || numberValue(selectedCheckResult.conflictCases?.length, 0))],
                  ['检测耗时', `${numberValue(selectedConflict.duration || selectedCheckResult.duration, 0)}ms`],
                  ['负责人ID', textValue(selectedConflict.owner, '-')],
                  ['检测时间', formatApiDate(textValue(selectedConflict.check_time, ''))],
                  ['创建时间', formatApiDate(selectedConflict.created_at)],
                  ['更新时间', formatApiDate(selectedConflict.updated_at)],
                ].map(([label, value]) => (
                <article key={label}>
                  <span>{label}</span>
                  <strong>{value}</strong>
                </article>
              ))}
            </div>

            <SectionCard title='记录来源'>
              <DataTable>
                <table>
                  <tbody>
                    <tr><td>数据表</td><td>conflict_check_records</td></tr>
                    <tr><td>风险判断</td><td>{riskLabel(selectedConflict.risk_level)} · {statusLabel(selectedConflict.status)}</td></tr>
                    <tr><td>处理建议</td><td>{['HIGH', 'CRITICAL'].includes(textValue(selectedConflict.risk_level, '').toUpperCase()) ? '建议暂停承办并发起冲突审批。' : '可继续合规评估，必要时发起豁免审批。'}</td></tr>
                  </tbody>
                </table>
              </DataTable>
            </SectionCard>

            <SectionCard title='风险评估结果'>
              <DataTable>
                <table>
                  <tbody>
                    <tr><td>总体风险</td><td>{riskLabel(textValue(selectedRiskAssessment.overallRisk || selectedConflict.risk_level, 'LOW'))}</td></tr>
                    <tr><td>风险评分</td><td>{numberValue(selectedRiskAssessment.riskScore, 0)}</td></tr>
                    <tr><td>风险原因</td><td>{selectedRiskReason}</td></tr>
                    <tr><td>需审批</td><td>{selectedRequiresApproval ? '是' : '否'}</td></tr>
                    <tr><td>检查范围</td><td>{textValue(selectedSearchParameters.searchDepth, 'STANDARD')} · {numberValue(selectedSearchParameters.searchYears, 5)}年</td></tr>
                    <tr><td>统计</td><td>检查案件 {numberValue(selectedStatistics.totalCasesChecked, selectedConflictCases.length)} 件，关联方 {numberValue(selectedStatistics.relatedPartiesChecked, 0)} 个</td></tr>
                  </tbody>
                </table>
              </DataTable>
            </SectionCard>

            <SectionCard title={`命中案件明细（${selectedConflictCases.length}）`}>
              <DataTable>
                <table>
                  <thead>
                    <tr><th>案件编号</th><th>案件名称</th><th>冲突类型</th><th>风险</th><th>状态</th><th>说明</th></tr>
                  </thead>
                  <tbody>
                    {selectedConflictCases.map((item) => (
                      <tr key={textValue(item.id || item.case_id)}>
                        <td>{textValue(item.case_no || item.case_number || item.caseNo || item.caseNumber || item.case_id)}</td>
                        <td>{textValue(item.case_name)}</td>
                        <td>{textValue(item.conflict_type)}</td>
                        <td><RiskTag text={riskLabel(textValue(item.risk_level, 'LOW'))} /></td>
                        <td>{statusLabel(textValue(item.case_status, '-'))}</td>
                        <td>{textValue(item.description, '-')}</td>
                      </tr>
                    ))}
                    {selectedConflictCases.length === 0 && <tr><td colSpan={6}>暂无命中案件明细</td></tr>}
                  </tbody>
                </table>
              </DataTable>
            </SectionCard>
          </div>
        )}
      </Modal>
    </div>
  )
}

export function ApprovalDecisionFlow() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [approval, setApproval] = React.useState<any>(null)
  const [snapshot, setSnapshot] = React.useState<any>(null)
  const [integrationStatus, setIntegrationStatus] = React.useState<any>(null)
  const [apiTimings, setApiTimings] = React.useState<Array<{ label: string; duration: number; at: string }>>([])
  const [deciding, setDeciding] = React.useState(false)

  const recordTiming = (label: string, startedAt: number) => {
    setApiTimings((current) => [
      { label, duration: Math.round(performance.now() - startedAt), at: new Date().toLocaleTimeString() },
      ...current,
    ].slice(0, 5))
  }

  const loadApproval = React.useCallback(async () => {
    if (!id) {
      return
    }
    try {
      const approvalStartedAt = performance.now()
      const approvalData = await apiRequest<any>(`/approvals/${id}`)
      recordTiming('审批详情', approvalStartedAt)
      setApproval(approvalData)

      try {
        const snapshotStartedAt = performance.now()
        const snapshotData = await apiRequest<any>(`/approvals/${id}/snapshot`)
        recordTiming('审批快照', snapshotStartedAt)
        setSnapshot(recordValue(snapshotData.snapshot))
      } catch {
        setSnapshot({})
      }

      const statusStartedAt = performance.now()
      const statusData = await apiRequest<any>(`/integration/approvals/${id}/status`)
      recordTiming('集成状态', statusStartedAt)
      setIntegrationStatus(statusData)
    } catch (error) {
      message.error(error instanceof Error ? error.message : '加载审批详情失败')
    }
  }, [id])

  React.useEffect(() => {
    loadApproval()
  }, [loadApproval])

  const decideApproval = async (
    decision: 'approve' | 'reject' | 'request_changes',
    reason: string,
    comments: string,
  ) => {
    if (!id) {
      return
    }
    setDeciding(true)
    try {
      const startedAt = performance.now()
      await apiRequest<any>(`/integration/approvals/${id}/decision`, {
        method: 'POST',
        body: JSON.stringify({
          decision,
          decision_reason: reason,
          decision_comments: comments,
        }),
      })
      recordTiming(decision === 'approve' ? '审批通过并成案' : decision === 'reject' ? '审批拒绝' : '退回修改', startedAt)
      await loadApproval()
      const statusData = await apiRequest<any>(`/integration/approvals/${id}/status`)
      if (decision === 'approve') {
        message.success(statusData?.case_creation?.case_id ? `已成案：${statusData.case_creation.case_number}` : '审批已通过')
      } else if (decision === 'reject') {
        message.success('审批已拒绝，未创建正式案件')
      } else {
        message.success('已退回修改，未创建正式案件')
      }
    } catch (error) {
      message.error(error instanceof Error ? error.message : '审批处理失败')
    } finally {
      setDeciding(false)
    }
  }

  const approve = () => decideApproval(
    'approve',
    '接案材料完整，冲突风险已完成复核，同意承办并创建正式案件。',
    '通过正式 API 审批并触发自动成案。',
  )

  const reject = () => decideApproval(
    'reject',
    '冲突风险或接案条件不满足，拒绝本次新建案件申请。',
    '审批拒绝后不会创建正式案件。',
  )

  const requestChanges = () => decideApproval(
    'request_changes',
    '接案资料或冲突说明需要补充，退回申请人修改后重新提交。',
    '退回修改后暂不创建正式案件。',
  )

  const approvalMetadata = recordValue(approval?.metadata)
  const snapshotMetadata = recordValue(snapshot?.metadata)
  const metadata = { ...approvalMetadata, ...snapshotMetadata }
  const conflictResult = metadata.conflict_result || snapshot?.conflict_result
  const conflictRecord = recordValue(metadata.conflict_record || snapshot?.conflict_record || conflictResult?.record)
  const approvalConflictCases = listOf<any>(metadata.conflict_cases || snapshot?.conflict_cases || conflictResult?.conflictCases)
  const conflictRisk = textValue(
    conflictResult?.riskAssessment?.overallRisk || conflictRecord.risk_level,
    'LOW',
  )
  const conflictCheckID = textValue(conflictResult?.checkId || conflictRecord.check_id || metadata.conflict_task_id, '已随审批快照冻结')
  const caseCreation = integrationStatus?.case_creation
  const applicationCaseTitle = textValue(
    metadata.case_creation_config?.title || snapshot?.case_creation_config?.title || conflictRecord.case_name || approval?.title,
    '未命名案件',
  )
  const approvalMaterialRows = listOf<any>(metadata.materials || snapshot?.materials)
  const approvalCommentRows = listOf<any>(approval?.records)
  const approvalTraceRows = approvalCommentRows.length > 0
    ? approvalCommentRows
    : [{
      id: 'submitted',
      approver_name: approval?.applicant_name || snapshot?.applicant?.submitted_name || '申请人',
      approver_role: '提交人',
      decision: approval?.status || 'submitted',
      decision_comments: approval?.content || `已提交立案审批：${applicationCaseTitle}`,
      created_at: approval?.created_at,
    }]
  const relatedInfoRows = [
    `关联冲突检测 ${conflictCheckID}`,
    `关联客户 ${textValue(conflictRecord.client_name || metadata.client?.name, '来自审批快照')}`,
    `关联案件 ${applicationCaseTitle}`,
    `关联流程 ${approval?.workflow_type || 'CONFLICT_APPROVAL'}`,
  ]
  const approvalAccess = normalizeApprovalAccess(approval)
  const canDecideApproval = approvalAccess.canDecide

  return (
    <div className='batch-page approval-page'>
      <PageHeader
        eyebrow='审批中心 / 我的审批 / 审批详情'
        title={approval?.title || '新建案件审批'}
        subtitle={`审批编号：${approval?.request_number || id || '加载中'} 状态：${approval?.status || '加载中'} 当前审批人：${approvalAccess.label}`}
        actions={
          <>
            <Badge count={caseCreation?.case_id ? `已成案 ${caseCreation.case_number}` : '正式 API'} color={caseCreation?.case_id ? '#12a89d' : '#f59f2f'} />
            <Button icon={<PrinterOutlined />}>打印</Button>
            <Button>更多审批操作</Button>
            <Tooltip title={caseCreation?.case_id ? '跳转到关联案件详情' : '暂无关联案件，审批通过后生成案件'}>
              <Button type='primary' disabled={!caseCreation?.case_id} onClick={() => caseCreation?.case_id && navigate(`/case/${caseCreation.case_id}`)}>查看关联案件</Button>
            </Tooltip>
          </>
        }
      />

      <div className='batch-approval-layout'>
        <SectionCard title='审批流程'>
          <div className='batch-approval-steps'>
            {[
              ['1. 利益冲突初检', '李助理（法务专员）', '已通过', '初检意见：未发现直接冲突，转合规复核。'],
              ['2. 合规复核', '刘合规（合规专员）', '已通过', '复核意见：存在潜在冲突，建议提交合伙人审议。'],
              ['3. 合伙人决策（当前节点）', `${approvalAccess.label}（当前节点）`, '审批中', `SLA 剩余：2天 03:45:18`],
              ['4. 管理合伙人终审', '管理合伙人', '待审批', '预计 2026-04-27'],
              ['5. 通知与归档', '系统自动处理', '待处理', ''],
            ].map((step, index) => (
              <article key={step[0]} className={index < 2 ? 'done' : index === 2 ? 'active' : ''}>
                <span>{index < 2 ? <CheckCircleOutlined /> : index + 1}</span>
                <div><strong>{step[0]}</strong><p>{step[1]}</p><em>{step[3]}</em></div>
                <RiskTag text={step[2]} />
              </article>
            ))}
          </div>
        </SectionCard>

        <main>
          <div className='batch-approval-main-grid'>
            <SectionCard title='申请信息'>
              <div className='batch-info-grid two'>
                {[
                  `申请类型 ${approval?.type || 'case_creation'}`,
                  `申请人 ${approval?.applicant_name || snapshot?.applicant?.submitted_name || '当前用户'}`,
                  `申请部门 ${approval?.department_name || snapshot?.applicant?.department_name || '公司业务部'}`,
                  `案件名称 ${applicationCaseTitle}`,
                  `关联客户 ${metadata.client?.name || '来自审批快照'}`,
                  `对方当事人 ${metadata.parties?.find?.((party: any) => party.role === 'opposing_party')?.name || '来自审批快照'}`,
                  `案件类型 ${snapshot?.case_creation_config?.case_type || 'commercial'}`,
                ].map((line) => (
                  <p key={line}><span>{line.split(' ')[0]}</span><strong>{line.substring(line.indexOf(' ') + 1)}</strong></p>
                ))}
              </div>
            </SectionCard>

            <SectionCard title='冲突检测摘要'>
              <div className='batch-approval-risk'>
                <AlertOutlined />
                <div><strong>总体风险等级：{riskLabel(conflictRisk)}</strong><p>风险评分：{conflictResult?.riskAssessment?.riskScore || 0} 检测 ID：{conflictCheckID}</p></div>
              </div>
              <div className='batch-hit-list'>
                <p><RiskTag text={riskLabel(conflictRisk)} />客户：{textValue(conflictRecord.client_name || metadata.client?.name, '来自审批快照')}</p>
                <p><RiskTag text={statusLabel(textValue(conflictRecord.status, approval?.status))} />案件：{applicationCaseTitle}</p>
                <p><RiskTag text={`${approvalConflictCases.length} 条命中`} />明细：{approvalConflictCases.slice(0, 2).map((item: any) => textValue(item.case_name || item.caseName, '')).filter(Boolean).join('、') || '审批快照暂无命中明细'}</p>
                <p><RiskTag text='正式 API' />来源：conflict_check_records / approval_snapshots</p>
              </div>
              <Button type='link' onClick={() => navigate('/conflict')}>返回利益冲突检测台</Button>
            </SectionCard>

            <SectionCard title={`冲突命中明细（${approvalConflictCases.length}）`}>
              <DataTable>
                <table>
                  <thead>
                    <tr><th>案件编号</th><th>案件名称</th><th>冲突类型</th><th>风险</th><th>说明</th></tr>
                  </thead>
                  <tbody>
                    {approvalConflictCases.map((item: any) => (
                      <tr key={textValue(item.id || item.case_id || item.caseId)}>
                        <td>{textValue(item.case_no || item.caseNo || item.case_id || item.caseId)}</td>
                        <td>{textValue(item.case_name || item.caseName)}</td>
                        <td>{textValue(item.conflict_type || item.conflictType)}</td>
                        <td><RiskTag text={riskLabel(textValue(item.risk_level || item.riskLevel, 'LOW'))} /></td>
                        <td>{textValue(item.description, '-')}</td>
                      </tr>
                    ))}
                    {approvalConflictCases.length === 0 && <tr><td colSpan={5}>暂无冲突命中明细</td></tr>}
                  </tbody>
                </table>
              </DataTable>
            </SectionCard>

            <SectionCard title='审批材料'>
              <div className='batch-tabs compact'>
                {['全部材料', '冲突报告', '申请文档', '证明材料', '其他'].map((tab, index) => <button key={tab} className={index === 0 ? 'active' : ''}>{tab}</button>)}
              </div>
              <DataTable>
                <table>
                  <tbody>
                    {approvalMaterialRows.map((row) => (
                      <tr key={row.name || row.id}><td><FileTextOutlined /> {textValue(row.name)}</td><td>{textValue(row.material_type || row.type)}</td><td>{textValue(row.status)}</td><td>{formatApiDate(row.created_at)}</td><td><Button type='link'>预览</Button></td></tr>
                    ))}
                    {approvalMaterialRows.length === 0 && <tr><td colSpan={5}>审批快照暂无材料记录</td></tr>}
                  </tbody>
                </table>
              </DataTable>
            </SectionCard>

            <SectionCard title={`意见记录（${approvalCommentRows.length}）`}>
              <div className='batch-comment-list'>
                {approvalCommentRows.map((item) => (
                  <article key={item.id || `${item.approver_name}-${item.created_at}`}>
                    <Avatar icon={<UserOutlined />} />
                    <div><strong>{textValue(item.approver_name)} <span>{textValue(item.approver_role)}</span></strong><p>{textValue(item.decision_comments || item.decision_reason)}</p></div>
                    <RiskTag text={textValue(item.decision)} />
                    <em>{formatApiDate(item.approval_date || item.created_at)}</em>
                  </article>
                ))}
                {approvalCommentRows.length === 0 && <p>暂无数据库审批意见</p>}
                <div className='batch-comment-input'>
                  <Input placeholder='添加审批意见（对所有可见）...' />
                  <Button type='primary' icon={<SendOutlined />}>发送</Button>
                </div>
              </div>
            </SectionCard>

            <SectionCard title={`审批记录（${approvalTraceRows.length}）`}>
              <div className='batch-comment-list'>
                {approvalTraceRows.map((item) => (
                  <article key={item.id || `${item.approver_name}-${item.created_at}`}>
                    <Avatar icon={<UserOutlined />} />
                    <div>
                      <strong>{textValue(item.approver_name)} <span>{textValue(item.approver_role)}</span></strong>
                      <p>{textValue(item.decision_comments || item.decision_reason)}</p>
                    </div>
                    <RiskTag text={statusLabel(textValue(item.decision, approval?.status))} />
                    <em>{formatApiDate(item.approval_date || item.created_at)}</em>
                  </article>
                ))}
              </div>
            </SectionCard>
          </div>
        </main>

        <aside>
          <SectionCard title='基本信息'>
            {[
              `审批状态 ${approval?.status || '加载中'}`,
              `优先级 ${approval?.priority || 'medium'}`,
              `发起时间 ${approval?.created_at ? new Date(approval.created_at).toLocaleString() : '加载中'}`,
              `成案状态 ${caseCreation?.status || 'pending'}`,
              `正式案件 ${caseCreation?.case_number || '审批通过后生成'}`,
            ].map((line) => <p key={line}>{line}</p>)}
          </SectionCard>
          <SectionCard title='正式 API 耗时'>
            {apiTimings.length === 0 ? <p>正在加载正式 API</p> : apiTimings.map((item) => (
              <p key={`${item.label}-${item.at}`}>{item.label} <strong>{item.duration}ms</strong> <span>{item.at}</span></p>
            ))}
          </SectionCard>
          <SectionCard title='关联信息'>
            {relatedInfoRows.map((line) => <p key={line}>{line}</p>)}
          </SectionCard>
        </aside>
      </div>

      <div className='batch-bottom-bar approval-actions'>
        <Button onClick={() => navigate('/approval')}>返回</Button>
        {(approval?.status === 'submitted' || approval?.status === 'under_review' || approval?.status === 'pending') && canDecideApproval && (
        <Space>
          <Button type='primary' className='approve-btn' icon={<CheckCircleOutlined />} loading={deciding} onClick={approve}>同意并成案</Button>
          <Button danger type='primary' loading={deciding} onClick={reject}>拒绝</Button>
          <Button className='return-btn' loading={deciding} onClick={requestChanges}>退回修改</Button>
          <Button>更多处理方式</Button>
        </Space>
        )}
        {(approval?.status === 'submitted' || approval?.status === 'under_review' || approval?.status === 'pending') && !canDecideApproval && (
          <Space>
            <Button onClick={() => message.info(approvalAccess.readonlyReason || '当前账号仅可查看审批进度，暂无更多处理方式')}>
              更多处理方式
            </Button>
            <span className='batch-approval-readonly'>
              {approvalAccess.readonlyReason}
            </span>
          </Space>
        )}
      </div>
    </div>
  )
}

export function ApprovalWorkbench() {
  const navigate = useNavigate()
  const [workbench, setWorkbench] = React.useState<any>({ items: [], stats: {} })
  const [apiTiming, setApiTiming] = React.useState<number | null>(null)

  React.useEffect(() => {
    const startedAt = performance.now()
    apiRequest<any>('/approvals/workbench')
      .then((data) => {
        setWorkbench(data || { items: [], stats: {} })
        setApiTiming(Math.round(performance.now() - startedAt))
      })
      .catch(() => {
        setApiTiming(null)
      })
  }, [])

  const approvalRows = listOf<Record<string, unknown>>(workbench.items)
  const approvalItems = approvalRows.length
    ? approvalRows.map((item) => [
      item.request_number || item.id,
      item.title || item.content || '未命名审批',
      item.priority === 'high' || item.priority === 'critical' ? '高风险' : '正常',
      item.current_stage || item.status || '待处理',
      item.current_approver_name || item.applicant_name || '未分配',
      item.status === 'approved' ? '已完成' : '正常',
      item.id,
    ])
    : []

  return (
    <div className='batch-page approval-workbench-page'>
      <PageHeader
        eyebrow='审批中心 / 审批工作台'
        title='审批工作台'
        subtitle='聚合冲突审查、豁免披露、接案审批、费用审批和退回补充，形成审批闭环入口。'
        actions={
          <>
            <Input prefix={<SearchOutlined />} placeholder='搜索审批编号、案件、客户、发起人...' />
            <span className='batch-autosave'>正式 API：{apiTiming === null ? '加载中' : `${apiTiming}ms`}</span>
            <Button>导出</Button>
            <Button type='primary' icon={<PlusOutlined />}>新建审批</Button>
          </>
        }
      />

      <div className='batch-metric-grid approval-metrics'>
        {[
          { icon: <AuditOutlined />, label: '待我审批', value: workbench.stats?.pending ?? 0, delta: '正式 API', tone: 'blue' as Tone },
          { icon: <SafetyCertificateOutlined />, label: '冲突审查', value: workbench.queues?.find?.((queue: any) => queue.key === 'conflict')?.count ?? 0, delta: '正式 API', tone: 'red' as Tone },
          { icon: <FileProtectOutlined />, label: '豁免披露', value: workbench.stats?.waiver_review ?? 0, delta: '正式 API', tone: 'orange' as Tone },
          { icon: <ClockCircleOutlined />, label: '需补充', value: workbench.stats?.needs_revision ?? 0, delta: '正式 API', tone: 'red' as Tone },
          { icon: <CheckCircleOutlined />, label: '队列总数', value: approvalItems.length, delta: '正式 API', tone: 'teal' as Tone },
        ].map((item) => <MetricCard key={item.label} {...item} />)}
      </div>

      <div className='batch-approval-board'>
        <SectionCard
          title='审批队列'
          extra={
            <Space>
              {['全部', '冲突审查', '豁免披露', '待补充', '已超时'].map((tab, index) => (
                <Button key={tab} type={index === 0 ? 'primary' : 'default'}>{tab}</Button>
              ))}
            </Space>
          }
          className='span-2'
        >
          <DataTable>
            <table>
              <thead>
                <tr>
                  <th>审批编号</th>
                  <th>标题</th>
                  <th>风险</th>
                  <th>当前节点</th>
                  <th>负责人</th>
                  <th>SLA</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {approvalItems.map((row: any[]) => (
                  <tr key={row[0]} className={row[2] === '高风险' ? 'danger-row' : ''}>
                    <td>{row[0]}</td>
                    <td>{row[1]}</td>
                    <td><RiskTag text={row[2]} /></td>
                    <td>{row[3]}</td>
                    <td>{row[4]}</td>
                    <td className={row[5].includes('超时') ? 'danger-text' : ''}>{row[5]}</td>
                    <td><Button size='small' aria-label={`进入审批 ${row[0]}`} onClick={() => navigate(`/approval/${row[6]}`)}>进入审批</Button></td>
                  </tr>
                ))}
                {approvalItems.length === 0 && <tr><td colSpan={7}>暂无数据库审批队列</td></tr>}
              </tbody>
            </table>
          </DataTable>
        </SectionCard>

        <SectionCard title='审批风险分布'>
          <div className='batch-donut-card'>
            <Progress type='circle' percent={74} format={() => '28'} strokeColor='#e8434e' trailColor='#f4d8dc' size={126} />
            <div className='batch-legend'>
              <span><StatusDot color='red' />冲突审批 <strong>{workbench.queues?.find?.((queue: any) => queue.key === 'conflict')?.count ?? 0}</strong></span>
              <span><StatusDot color='orange' />豁免评估 <strong>{workbench.stats?.waiver_review ?? 0}</strong></span>
              <span><StatusDot color='teal' />待补充 <strong>{workbench.stats?.needs_revision ?? 0}</strong></span>
              <span><StatusDot color='blue' />全部 <strong>{approvalItems.length}</strong></span>
            </div>
          </div>
        </SectionCard>

        <SectionCard title='SLA 预警'>
          <div className='batch-overdue-list'>
            {approvalItems.slice(0, 4).map((row: any[], index: number) => (
              <p key={row[0]}><StatusDot color={index === 0 ? 'red' : 'orange'} />{row[1]}<span className='danger-text'>{row[5]}</span></p>
            ))}
            {approvalItems.length === 0 && <p>暂无数据库 SLA 预警</p>}
          </div>
        </SectionCard>

        <SectionCard title='豁免与披露进度'>
          <div className='batch-approval-mini-flow'>
            {['生成披露文件', '内部合规审批', '向客户发送披露文件', '客户签署', '我所签署确认', '生效与归档'].map((step, index) => (
              <p key={step} className={index < 3 ? 'done' : index === 3 ? 'active' : ''}>
                <span>{index + 1}</span>{step}<RiskTag text={index < 3 ? '已完成' : index === 3 ? '进行中' : '待处理'} />
              </p>
            ))}
          </div>
        </SectionCard>

        <SectionCard title='最近审批意见' className='span-2'>
          <div className='batch-comment-list'>
            {approvalRows.slice(0, 5).map((item) => (
              <article key={`approval-${textValue(item.id || item.request_number)}`}>
                <Avatar icon={<UserOutlined />} />
                <div>
                  <strong>
                    {textValue(item.current_approver_name || item.applicant_name)}
                    <span>{textValue(item.current_stage || item.type, '审批节点')}</span>
                  </strong>
                  <p>{textValue(item.content || item.title, '暂无审批说明')}</p>
                </div>
                <RiskTag text={statusLabel(textValue(item.status, ''))} />
                <em>{formatApiDate(textValue(item.updated_at || item.created_at, ''))}</em>
              </article>
            ))}
            {approvalRows.length === 0 && <p>暂无数据库审批意见</p>}
          </div>
        </SectionCard>
      </div>
    </div>
  )
}

export function LawyerResourceCenter() {
  const navigate = useNavigate()
  const [resource, setResource] = React.useState<LawyersResourcePayload>({})
  const [apiTiming, setApiTiming] = React.useState<number | null>(null)

  React.useEffect(() => {
    const startedAt = performance.now()
    apiRequest<LawyersResourcePayload>('/lawyers/resource-center')
      .then((data) => {
        setResource(data || {})
        setApiTiming(Math.round(performance.now() - startedAt))
      })
      .catch(() => {
        setResource({})
        setApiTiming(null)
      })
  }, [])

  const lawyerRows = listOf<Record<string, unknown>>(resource.lawyers)
  const capacityRows = listOf<CommandCenterCount>(resource.capacity)
  const assignmentRows = listOf<CommandCenterCaseRow>(resource.assignments)
  const taskRows = listOf<CommandCenterTodo>(resource.tasks)

  return (
    <div className='batch-page lawyer-resource-page'>
      <PageHeader
        eyebrow='律师管理 / 团队与人效'
        title='律师资源管理'
        subtitle='围绕接案团队、审批责任、冲突复核和工时负荷管理律师资源。'
        actions={
          <>
            <Input prefix={<SearchOutlined />} placeholder='搜索律师、专长、案件、审批...' />
            <span className='batch-autosave'>正式 API：{apiTiming === null ? '加载中' : `${apiTiming}ms`}</span>
            <Button icon={<CalendarOutlined />}>排期视图</Button>
            <Button type='primary' icon={<PlusOutlined />}>新增律师</Button>
          </>
        }
      />

      <div className='batch-metric-grid lawyer-metrics'>
        {[
          { icon: <TeamOutlined />, label: '执业律师', value: resource.summary?.lawyers ?? lawyerRows.length, delta: '正式 API', tone: 'blue' as Tone },
          { icon: <ApartmentOutlined />, label: '部门覆盖', value: resource.summary?.departments ?? capacityRows.length, delta: '来自 users.department', tone: 'teal' as Tone },
          { icon: <FolderOpenOutlined />, label: '在办案件', value: resource.summary?.active_cases ?? assignmentRows.length, delta: '来自 cases', tone: 'orange' as Tone },
          { icon: <SafetyCertificateOutlined />, label: '可分配账号', value: lawyerRows.filter((row) => textValue(row.status, '').toLowerCase() === 'active').length, delta: 'active 状态', tone: 'green' as Tone },
          { icon: <FileDoneOutlined />, label: '待处理事项', value: resource.summary?.pending_tasks ?? taskRows.length, delta: '来自 inbox_items', tone: 'red' as Tone },
        ].map((item) => <MetricCard key={item.label} {...item} />)}
      </div>

      <div className='batch-lawyer-layout'>
        <SectionCard title='律师负荷与可分配能力' className='span-2'>
          <DataTable>
            <table>
              <thead>
                <tr>
                  <th>律师</th>
                  <th>角色</th>
                  <th>部门</th>
                  <th>职级</th>
                  <th>邮箱</th>
                  <th>状态</th>
                  <th>入库时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {lawyerRows.map((row) => (
                  <tr key={textValue(row.id || row.email)} className={textValue(row.status, '').toLowerCase() !== 'active' ? 'danger-row' : ''}>
                    <td><Avatar size='small' icon={<UserOutlined />} /> {textValue(row.name || row.username)}</td>
                    <td>{roleLabel(textValue(row.role, ''))}</td>
                    <td>{textValue(row.department)}</td>
                    <td>{textValue(row.seniority)}</td>
                    <td>{textValue(row.email)}</td>
                    <td><RiskTag text={accountStatusLabel(textValue(row.status, ''))} /></td>
                    <td>{formatApiDate(textValue(row.created_at, ''))}</td>
                    <td>
                      <Button size='small' onClick={() => navigate(`/lawyer/${textValue(row.id)}`)}>
                        查看档案
                      </Button>
                    </td>
                  </tr>
                ))}
                {lawyerRows.length === 0 && <tr><td colSpan={8}>暂无数据库律师账号</td></tr>}
              </tbody>
            </table>
          </DataTable>
        </SectionCard>

        <SectionCard title='部门覆盖分布'>
          <div className='batch-ranking-list'>
            {capacityRows.slice(0, 5).map((item, index) => (
              <div className='batch-ranking-row' key={textValue(item.key)}>
                <span>{index + 1}</span>
                <strong>{textValue(item.key, '未分配部门')}</strong>
                <em>{numberValue(item.count)}人</em>
                <Progress percent={Math.min(100, numberValue(item.count) * 20)} size='small' showInfo={false} />
              </div>
            ))}
            {capacityRows.length === 0 && <p>暂无数据库部门分布</p>}
          </div>
        </SectionCard>

        <SectionCard title='团队指派与冲突责任' className='span-2'>
          <div className='batch-lawyer-assignment'>
            {assignmentRows.map((row) => (
              <article key={textValue(row.id || row.case_number)}>
                <h3>{textValue(row.title)}</h3>
                <p>负责人：{textValue(row.lawyer_name)} 客户：{textValue(row.client_name)}</p>
                <RiskTag text={statusLabel(row.status)} />
                <Button type='primary' onClick={() => navigate('/conflict')}>进入冲突检查</Button>
              </article>
            ))}
            {assignmentRows.length === 0 && <p>暂无数据库案件指派</p>}
          </div>
        </SectionCard>

        <SectionCard title='待处理律师事项'>
          <div className='batch-overdue-list'>
            {taskRows.map((task) => (
              <p key={textValue(task.id || task.title)}>
                <StatusDot color={task.priority === 'critical' || task.priority === 'high' ? 'red' : 'orange'} />
                {textValue(task.type || task.source_type, '待办')} · {textValue(task.title)}
                <RiskTag text={textValue(task.priority, '普通')} />
              </p>
            ))}
            {taskRows.length === 0 && <p>暂无数据库待处理事项</p>}
          </div>
        </SectionCard>

        <SectionCard title='业务领域覆盖'>
          <div className='batch-scope-grid'>
            {capacityRows.map((item) => (
              <span key={textValue(item.key)}>{textValue(item.key, '未分配部门')} {numberValue(item.count)}人</span>
            ))}
            {capacityRows.length === 0 && <span>暂无数据库业务领域</span>}
          </div>
        </SectionCard>
      </div>
    </div>
  )
}

export function LawyerProfileCenter() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const [lawyer, setLawyer] = React.useState<Record<string, unknown> | null>(null)
  const [caseRows, setCaseRows] = React.useState<Record<string, unknown>[]>([])
  const [loading, setLoading] = React.useState(false)
  const [apiTiming, setApiTiming] = React.useState<number | null>(null)

  React.useEffect(() => {
    if (!id) return
    const startedAt = performance.now()
    setLoading(true)
    Promise.all([
      apiRequest<Record<string, unknown>>(`/lawfirm/lawyers/${id}`),
      apiRequest<any>(`/cases?page=1&page_size=20&lawyer_id=${id}`),
    ])
      .then(([lawyerData, casesData]) => {
        setLawyer(lawyerData || null)
        setCaseRows(normalizeCaseRows(casesData))
        setApiTiming(Math.round(performance.now() - startedAt))
      })
      .catch(() => {
        setLawyer(null)
        setCaseRows([])
        setApiTiming(null)
      })
      .finally(() => setLoading(false))
  }, [id])

  const lawyerName = textValue(lawyer?.name || lawyer?.username, '律师档案')
  const activeCases = caseRows.filter((row) =>
    ['active', 'in_progress', 'pending'].includes(textValue(row.status, '').toLowerCase()),
  )
  const highPriorityCases = caseRows.filter((row) =>
    ['high', 'urgent', 'critical'].includes(textValue(row.priority, '').toLowerCase()),
  )

  if (loading && !lawyer) {
    return (
      <div className='batch-page lawyer-resource-page'>
        <PageHeader
          eyebrow='律师管理 / 律师档案'
          title='正在加载律师档案'
          actions={<Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/lawyer')}>返回律师资源</Button>}
        />
      </div>
    )
  }

  if (!lawyer) {
    return (
      <div className='batch-page lawyer-resource-page'>
        <PageHeader
          eyebrow='律师管理 / 律师档案'
          title='律师档案不存在'
          subtitle='正式 API 未返回该律师账号。'
          actions={<Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/lawyer')}>返回律师资源</Button>}
        />
      </div>
    )
  }

  return (
    <div className='batch-page lawyer-resource-page'>
      <PageHeader
        eyebrow='律师管理 / 团队与人效 / 律师档案'
        title={lawyerName}
        subtitle={`${textValue(lawyer.department, '未分配部门')} · ${roleLabel(textValue(lawyer.role, 'lawyer'))}`}
        actions={
          <>
            <span className='batch-autosave'>正式 API：{apiTiming === null ? '加载中' : `${apiTiming}ms`}</span>
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/lawyer')}>返回律师资源</Button>
          </>
        }
      />

      <div className='batch-metric-grid lawyer-metrics'>
        {[
          { icon: <UserOutlined />, label: '账号状态', value: accountStatusLabel(textValue(lawyer.status, 'active')), delta: 'users.status', tone: textValue(lawyer.status, '').toLowerCase() === 'active' ? 'green' as Tone : 'orange' as Tone },
          { icon: <FolderOpenOutlined />, label: '负责案件', value: caseRows.length, delta: 'cases.lawyer_id', tone: 'blue' as Tone },
          { icon: <FileDoneOutlined />, label: '在办案件', value: activeCases.length, delta: 'active/pending', tone: 'teal' as Tone },
          { icon: <AlertOutlined />, label: '高优先级', value: highPriorityCases.length, delta: 'priority', tone: highPriorityCases.length > 0 ? 'red' as Tone : 'green' as Tone },
          { icon: <ClockCircleOutlined />, label: '入库时间', value: formatApiDate(textValue(lawyer.created_at, '')), delta: '正式 API', tone: 'slate' as Tone },
        ].map((item) => <MetricCard key={item.label} {...item} />)}
      </div>

      <div className='batch-lawyer-layout'>
        <SectionCard title='律师基本档案'>
          <div className='batch-list'>
            {[
              ['姓名', lawyerName],
              ['邮箱', textValue(lawyer.email, '-')],
              ['电话', textValue(lawyer.phone, '-')],
              ['部门', textValue(lawyer.department, '-')],
              ['职级', textValue(lawyer.seniority || lawyer.position, '-')],
              ['角色', roleLabel(textValue(lawyer.role, 'lawyer'))],
            ].map(([label, value]) => (
              <article key={label}>
                <div>
                  <strong>{label}</strong>
                  <p>{value}</p>
                </div>
              </article>
            ))}
          </div>
        </SectionCard>

        <SectionCard title='执业与权限状态'>
          <div className='batch-overdue-list'>
            <p><StatusDot color={textValue(lawyer.status, '').toLowerCase() === 'active' ? 'green' : 'orange'} />账号状态<RiskTag text={accountStatusLabel(textValue(lawyer.status, 'active'))} /></p>
            <p><StatusDot color='blue' />所属部门<RiskTag text={textValue(lawyer.department, '未分配')} /></p>
            <p><StatusDot color='slate' />最近更新<RiskTag text={formatApiDate(textValue(lawyer.updated_at || lawyer.created_at, ''))} /></p>
          </div>
        </SectionCard>

        <SectionCard title='负责案件' className='span-2'>
          <DataTable>
            <table>
              <thead>
                <tr>
                  <th>案件编号</th>
                  <th>案件名称</th>
                  <th>客户</th>
                  <th>类型</th>
                  <th>状态</th>
                  <th>优先级</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {caseRows.map((row) => (
                  <tr key={textValue(row.id || row.case_number)}>
                    <td>{textValue(row.case_number || row.id)}</td>
                    <td>{textValue(row.title)}</td>
                    <td>{textValue(recordValue(row.client).name || row.client_name, '-')}</td>
                    <td><RiskTag text={dbCaseType(textValue(row.case_type, ''))} /></td>
                    <td><RiskTag text={statusLabel(textValue(row.status, ''))} /></td>
                    <td><RiskTag text={textValue(row.priority, 'medium')} /></td>
                    <td><Button size='small' onClick={() => navigate(`/case/${textValue(row.id)}`)}>查看案件</Button></td>
                  </tr>
                ))}
                {caseRows.length === 0 && <tr><td colSpan={7}>暂无数据库案件指派</td></tr>}
              </tbody>
            </table>
          </DataTable>
        </SectionCard>
      </div>
    </div>
  )
}

export function UserAccessCenter() {
  const [access, setAccess] = React.useState<AdminAccessPayload>({})
  const [apiTiming, setApiTiming] = React.useState<number | null>(null)
  const [roleModalOpen, setRoleModalOpen] = React.useState(false)
  const [selectedUser, setSelectedUser] = React.useState<Record<string, unknown> | null>(null)
  const [allRoles, setAllRoles] = React.useState<Role[]>([])
  const [selectedRoleIds, setSelectedRoleIds] = React.useState<number[]>([])
  const [roleSaving, setRoleSaving] = React.useState(false)

  const loadAccess = React.useCallback(() => {
    const startedAt = performance.now()
    apiRequest<AdminAccessPayload>('/admin/access-center')
      .then((data) => {
        setAccess(data || {})
        setApiTiming(Math.round(performance.now() - startedAt))
      })
      .catch(() => {
        setAccess({})
        setApiTiming(null)
      })
  }, [])

  React.useEffect(() => {
    loadAccess()
  }, [loadAccess])

  const userRows = listOf<Record<string, unknown>>(access.users)
  const roleRows = listOf<CommandCenterCount>(access.roles)
  const changeRows = listOf<Record<string, unknown>>(access.permission_changes)
  const auditRows = listOf<Record<string, unknown>>(access.audit_events)
  const selectedUserId = Number(textValue(selectedUser?.id, '0'))

  const openRoleEditor = async (row: Record<string, unknown>) => {
    const userID = Number(textValue(row.id, '0'))
    if (!userID) {
      message.error('用户ID无效')
      return
    }

    setSelectedUser(row)
    setRoleModalOpen(true)
    try {
      const [roles, userRoles] = await Promise.all([getAllRoles(), getUserRoles(userID)])
      setAllRoles(roles || [])
      setSelectedRoleIds((userRoles || []).map((role) => role.id))
    } catch (error) {
      message.error('加载用户角色失败')
    }
  }

  const saveUserRoles = async () => {
    if (!selectedUserId) {
      message.error('用户ID无效')
      return
    }

    setRoleSaving(true)
    try {
      await assignUserRoles(selectedUserId, selectedRoleIds)
      message.success('用户角色已更新')
      setRoleModalOpen(false)
      loadAccess()
    } catch (error) {
      message.error('保存用户角色失败')
    } finally {
      setRoleSaving(false)
    }
  }

  return (
    <div className='batch-page user-access-page'>
      <PageHeader
        eyebrow='系统管理 / 用户与权限'
        title='用户管理'
        subtitle='集中管理账号状态、角色授权、数据域权限和高敏操作审计。'
        actions={
          <>
            <Input prefix={<SearchOutlined />} placeholder='搜索姓名、邮箱、角色、权限...' />
            <span className='batch-autosave'>正式 API：{apiTiming === null ? '加载中' : `${apiTiming}ms`}</span>
            <Button icon={<DownloadOutlined />}>导出</Button>
            <Button type='primary' icon={<PlusOutlined />}>新增用户</Button>
          </>
        }
      />

      <div className='batch-metric-grid user-metrics'>
        {[
          { icon: <UserOutlined />, label: '系统用户', value: access.summary?.users ?? userRows.length, delta: '正式 API', tone: 'blue' as Tone },
          { icon: <TeamOutlined />, label: '活跃账号', value: access.summary?.active_users ?? userRows.filter((row) => textValue(row.status, '').toLowerCase() === 'active').length, delta: 'users.status', tone: 'teal' as Tone },
          { icon: <KeyOutlined />, label: '角色数量', value: access.summary?.roles ?? roleRows.length, delta: 'RBAC 角色', tone: 'orange' as Tone },
          { icon: <LockOutlined />, label: '停用/锁定', value: access.summary?.disabled_users ?? 0, delta: '非 active 状态', tone: 'green' as Tone },
          { icon: <AuditOutlined />, label: '权限变更待审', value: access.summary?.pending_changes ?? changeRows.length, delta: 'approval_requests', tone: 'red' as Tone },
        ].map((item) => <MetricCard key={item.label} {...item} />)}
      </div>

      <div className='batch-admin-layout'>
        <SectionCard
          title='账号清单'
          className='span-2'
          extra={
            <Space>
              {['全部', '管理员', '律师', '助理', '合规', '停用'].map((tab, index) => (
                <Button key={tab} type={index === 0 ? 'primary' : 'default'}>{tab}</Button>
              ))}
            </Space>
          }
        >
          <DataTable>
            <table>
              <thead>
                <tr>
                  <th>用户</th>
                  <th>岗位</th>
                  <th>邮箱</th>
                  <th>角色</th>
                  <th>状态</th>
                  <th>最近登录</th>
                  <th>权限域</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {userRows.map((row) => (
                  <tr key={textValue(row.id || row.email)} className={textValue(row.status, '').toLowerCase() !== 'active' ? 'danger-row' : ''}>
                    <td><Avatar size='small' icon={<UserOutlined />} /> {textValue(row.name || row.username)}</td>
                    <td>{textValue(row.seniority || row.department)}</td>
                    <td>{textValue(row.email)}</td>
                    <td><RiskTag text={roleLabel(textValue(row.role, ''))} /></td>
                    <td><RiskTag text={accountStatusLabel(textValue(row.status, ''))} /></td>
                    <td>{formatApiDate(textValue(row.updated_at || row.created_at, ''))}</td>
                    <td>{textValue(row.department, '未分配部门')}</td>
                    <td><Button size='small' onClick={() => openRoleEditor(row)}>编辑角色</Button></td>
                  </tr>
                ))}
                {userRows.length === 0 && <tr><td colSpan={8}>暂无数据库用户账号</td></tr>}
              </tbody>
            </table>
          </DataTable>
        </SectionCard>

        <SectionCard title='角色矩阵'>
          <div className='batch-role-list'>
            {roleRows.map((role) => (
              <article key={textValue(role.key)}>
                <div>
                  <strong>{roleLabel(textValue(role.key, ''))} <span>{numberValue(role.count)}人</span></strong>
                  <p>来自 RBAC 角色授权</p>
                </div>
                <RiskTag text={role.key === 'admin' ? '高敏' : '标准'} />
              </article>
            ))}
            {roleRows.length === 0 && <p>暂无数据库角色</p>}
          </div>
        </SectionCard>

        <SectionCard title='权限变更审批'>
          <div className='batch-overdue-list'>
            {changeRows.map((item) => (
              <p key={textValue(item.id || item.request_number)}>
                <StatusDot color={textValue(item.priority, '').toLowerCase() === 'high' ? 'red' : 'orange'} />
                {textValue(item.applicant_name)} · {textValue(item.title || item.content)}
                <RiskTag text={statusLabel(textValue(item.status, ''))} />
              </p>
            ))}
            {changeRows.length === 0 && <p>暂无数据库权限变更审批</p>}
          </div>
        </SectionCard>

        <SectionCard title='安全审计事件' className='span-2'>
          <div className='batch-policy-grid'>
            {auditRows.map((item) => (
              <article key={textValue(item.id || item.subject_id || item.created_at)}>
                <KeyOutlined />
                <div><strong>{textValue(item.event_type || item.action, '审计事件')}</strong><p>{textValue(item.summary || item.description, '暂无审计说明')}</p></div>
                <RiskTag text={riskLabel(textValue(item.risk_level, ''))} />
              </article>
            ))}
            {auditRows.length === 0 && <p>暂无数据库安全审计事件</p>}
          </div>
        </SectionCard>
      </div>

      <Modal
        title={`编辑角色 - ${textValue(selectedUser?.name || selectedUser?.username, '')}`}
        open={roleModalOpen}
        onOk={saveUserRoles}
        confirmLoading={roleSaving}
        onCancel={() => setRoleModalOpen(false)}
        destroyOnHidden
      >
        <Select
          mode='multiple'
          value={selectedRoleIds}
          onChange={setSelectedRoleIds}
          placeholder='选择用户角色'
          style={{ width: '100%' }}
          options={allRoles.map((role) => ({
            label: `${role.name} (${role.code})`,
            value: role.id,
          }))}
        />
      </Modal>
    </div>
  )
}

export function SystemSettingsCenter() {
  const [overview, setOverview] = React.useState<SettingsOverviewPayload>({})
  const [apiTiming, setApiTiming] = React.useState<number | null>(null)

  React.useEffect(() => {
    const startedAt = performance.now()
    apiRequest<SettingsOverviewPayload>('/settings/overview')
      .then((data) => {
        setOverview(data || {})
        setApiTiming(Math.round(performance.now() - startedAt))
      })
      .catch(() => {
        setOverview({})
        setApiTiming(null)
      })
  }, [])

  const moduleRows = listOf<CommandCenterCount>(overview.modules)
  const settingRows = listOf<Record<string, unknown>>(overview.settings)
  const approvalSettings = settingRows.filter((row) => textValue(row.category, '').includes('approval'))
  const notificationSettings = settingRows.filter((row) => textValue(row.category, '').includes('notification') || textValue(row.category, '').includes('sla'))
  const auditSettings = settingRows.filter((row) => textValue(row.category, '').includes('audit') || textValue(row.category, '').includes('security') || textValue(row.category, '').includes('file'))

  return (
    <div className='batch-page system-settings-page'>
      <PageHeader
        eyebrow='系统管理 / 配置中心'
        title='系统设置'
        subtitle='围绕接案、冲突、审批、通知、审计和文件归档配置系统级规则。'
        actions={
          <>
            <span className='batch-autosave'>正式 API：{apiTiming === null ? '加载中' : `${apiTiming}ms`}</span>
            <Button icon={<DownloadOutlined />}>导出配置</Button>
            <Button>恢复默认</Button>
            <Button type='primary' icon={<SettingOutlined />}>保存设置</Button>
          </>
        }
      />

      <div className='batch-metric-grid settings-metrics'>
        {[
          { icon: <SettingOutlined />, label: '配置项', value: overview.summary?.settings ?? settingRows.length, delta: 'system_settings', tone: 'blue' as Tone },
          { icon: <SafetyCertificateOutlined />, label: '配置分组', value: overview.summary?.modules ?? moduleRows.length, delta: 'category 聚合', tone: 'red' as Tone },
          { icon: <BellOutlined />, label: '通知策略', value: notificationSettings.length, delta: '正式 API', tone: 'orange' as Tone },
          { icon: <FileProtectOutlined />, label: '审计策略', value: auditSettings.length, delta: '正式 API', tone: 'teal' as Tone },
          { icon: <CloudUploadOutlined />, label: '启用配置', value: settingRows.filter(settingEnabled).length, delta: 'setting_value.enabled', tone: 'green' as Tone },
        ].map((item) => <MetricCard key={item.label} {...item} />)}
      </div>

      <div className='batch-settings-layout'>
        <SectionCard title='配置分组' className='span-2'>
          <DataTable>
            <table>
              <thead>
                <tr>
                  <th>分组</th>
                  <th>说明</th>
                  <th>状态</th>
                  <th>优先级</th>
                  <th>启用</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {moduleRows.map((row) => (
                  <tr key={textValue(row.key)}>
                    <td>{textValue(row.key, '未分类')}</td>
                    <td>该分组包含 {numberValue(row.count)} 个数据库配置项</td>
                    <td><RiskTag text='数据库配置' /></td>
                    <td><RiskTag text={numberValue(row.count) > 3 ? '高优先级' : '标准'} /></td>
                    <td><Switch defaultChecked={numberValue(row.count) > 0} /></td>
                    <td><Button size='small'>配置</Button></td>
                  </tr>
                ))}
                {moduleRows.length === 0 && <tr><td colSpan={6}>暂无数据库配置分组</td></tr>}
              </tbody>
            </table>
          </DataTable>
        </SectionCard>

        <SectionCard title='审批规则'>
          <div className='batch-settings-stack'>
            {approvalSettings.map((item) => (
              <p key={textValue(item.id || item.setting_key)}>
                <span>{textValue(item.description || item.setting_key)}</span>
                <Switch defaultChecked={settingEnabled(item)} />
              </p>
            ))}
            {approvalSettings.length === 0 && <p><span>暂无数据库审批规则</span></p>}
          </div>
        </SectionCard>

        <SectionCard title='通知与 SLA'>
          <div className='batch-settings-stack'>
            {notificationSettings.map((item) => (
              <p key={textValue(item.id || item.setting_key)}>
                <span>{textValue(item.description || item.setting_key)}</span>
                <RiskTag text={textValue(settingObject(item.setting_value).channel || item.category, '已配置')} />
              </p>
            ))}
            {notificationSettings.length === 0 && <p><span>暂无数据库通知策略</span></p>}
          </div>
        </SectionCard>

        <SectionCard title='数据与审计' className='span-2'>
          <div className='batch-policy-grid'>
            {auditSettings.map((item) => (
              <article key={textValue(item.id || item.setting_key)}>
                <LockOutlined />
                <div><strong>{textValue(item.setting_key)}</strong><p>{textValue(item.description, '暂无配置说明')}</p></div>
                <RiskTag text={settingEnabled(item) ? '已启用' : '未启用'} />
              </article>
            ))}
            {auditSettings.length === 0 && <p>暂无数据库数据与审计策略</p>}
          </div>
        </SectionCard>
      </div>
    </div>
  )
}
