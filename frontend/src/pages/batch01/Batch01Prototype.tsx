import React from 'react'
import {
  Avatar,
  Alert,
  Badge,
  Button,
  Input,
  Modal,
  Progress,
  Segmented,
  Select,
  Space,
  Switch,
  Tag,
  Tooltip,
  Upload,
} from 'antd'
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
  DatabaseOutlined,
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
  SettingOutlined,
  TeamOutlined,
  UserOutlined,
} from '@ant-design/icons'
import { useNavigate, useParams, useSearchParams } from 'react-router'
import { assignUserRoles, getAllRoles, getUserRoles } from '@/services/role'
import type { Role } from '@/services/role'
import { getRoles, getToken, getUserInfo } from '@/utils/storage'
import { message } from '@/utils/messageHelper'
import { useAppStore } from '@/stores/useAppStore'
import { hasPermission } from '@/utils/accessControl'
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
  { icon: <FileDoneOutlined />, label: '待办事项', value: 0, delta: '', tone: 'blue' },
  {
    icon: <SafetyCertificateOutlined />,
    label: '利益冲突待复核',
    value: 0,
    delta: '',
    tone: 'red',
  },
  { icon: <AuditOutlined />, label: '待审批事项', value: 0, delta: '', tone: 'orange' },
  { icon: <FolderOpenOutlined />, label: '接案准备中', value: 0, delta: '', tone: 'teal' },
  { icon: <FileTextOutlined />, label: '在办案件总数', value: 0, delta: '', tone: 'blue' },
  { icon: <ClockCircleOutlined />, label: '逾期任务', value: 0, delta: '', tone: 'red' },
  {
    icon: <DollarOutlined />,
    label: '合同回款预警',
    value: 0,
    delta: '',
    tone: 'orange',
  },
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
  evidenceSummary?: string
  source_case?: string
  sourceCase?: string
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
  conflict_details?: string
  description?: string
  conflict_cases?: Array<Record<string, unknown>>
  evidence?: Array<Record<string, unknown>>
  coverage_status?: string
  approval_id?: string
  approval_status?: string
  approval_request_number?: string
  approval_current_approver_id?: string | number
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

interface SubjectPartyOption {
  entity_id: number
  name: string
  entity_type?: string
  role?: string
  party_type?: string
  identity_type?: string
  identity_present?: boolean
  identity_hint?: string
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
  message?: string
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

function conflictScopeLabel(coverageStatus: unknown) {
  return String(coverageStatus || '').toUpperCase() === 'COMPLETE'
    ? '已检索律所配置的权威数据源；最终覆盖范围以审计记录为准'
    : '系统已登记历史（覆盖完整性待确认，未登记档案需人工核查）'
}

function conflictQueueScopeLabel(canReviewConflict: boolean) {
  return canReviewConflict ? '全所冲突核查队列' : '当前账号可见的冲突任务'
}

function escapeHtml(value: unknown) {
  return String(value ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

function firstPresent(...values: unknown[]) {
  return values.find((value) => value !== null && value !== undefined && value !== '')
}

function numberValue(value: unknown, fallback = 0) {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return value
  }
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : fallback
}

function DiagnosticDetails({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className='batch-diagnostic-details'>
      <strong className='batch-diagnostic-title'>{label}</strong>
      <div className='batch-diagnostic-content'>{children}</div>
    </div>
  )
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
  if (normalized === 'REVIEW_REQUIRED') return '待人工复核'
  if (normalized === 'BLOCKED') return '暂停接案'
  if (normalized === 'WAIVER_PENDING') return '豁免评估中'
  if (normalized === 'WAIVED') return '已按条件豁免'
  return value || '未评级'
}

function formatRiskScore(value: unknown) {
  const raw = numberValue(value, 0)
  const normalized = raw > 0 && raw <= 1 ? raw * 100 : raw
  return normalized
    .toFixed(2)
    .replace(/\.00$/, '')
    .replace(/(\.\d)0$/, '$1')
}

function searchDepthLabel(value: unknown) {
  const labels: Record<string, string> = {
    BASIC: '基础检索',
    STANDARD: '标准检索',
    DEEP: '深度检索',
  }
  const normalized = textValue(value, 'STANDARD').toUpperCase()
  return labels[normalized] || '标准检索'
}

function conflictMatchTypeLabel(value: unknown) {
  const normalized = textValue(value, '').toUpperCase()
  const labels: Record<string, string> = {
    EXACT: '规范化名称完全一致',
    EXACT_NORMALIZED: '规范化名称完全一致',
    NAME_CANDIDATE: '名称候选，待人工核实',
    NORMALIZED_CONTAINS: '规范化名称包含关系',
    RELATED_PARTY: '关联方关系命中',
  }
  return labels[normalized] || textValue(value, '未提供')
}

function conflictSubjectRoleLabel(value: unknown) {
  const normalized = textValue(value, '').toUpperCase()
  const labels: Record<string, string> = {
    CLIENT: '客户/委托人',
    OPPOSING: '对方当事人',
    OPPOSING_PARTY: '对方当事人',
    RELATED_PARTY: '关联方',
    BENEFICIAL_OWNER: '实际控制人/受益所有人',
    LAWYER: '承办律师',
  }
  return labels[normalized] || textValue(value, '未提供')
}

function conflictRuleLabel(value: unknown) {
  const normalized = textValue(value, '').toUpperCase()
  const labels: Record<string, string> = {
    SUBJECT_CANDIDATE_REVIEW: '名称候选，需要核实主体身份',
    DIRECT_ADVERSE_CURRENT_CLIENT: '对方为本所现有客户',
    DIRECT_ADVERSE_FORMER_CLIENT: '对方为本所既往客户',
    STRUCTURED_IDENTITY_EXACT: '主体唯一标识一致',
    RELATED_PARTY_CONFLICT: '关联方关系命中',
  }
  return labels[normalized] || (normalized ? '其他规则（技术编号见审计信息）' : '未提供')
}

function conflictAlgorithmLabel(value: unknown) {
  const normalized = textValue(value, '').toUpperCase()
  const labels: Record<string, string> = {
    NORMALIZED_EXACT: '名称规范化后精确比对',
    NORMALIZED_CONTAINS: '名称规范化后包含比对',
    ENTITY_ID: '主体唯一标识比对',
    RELATION_GRAPH: '关联方关系图谱',
  }
  return labels[normalized] || textValue(value, '未提供')
}

function conflictTypeLabel(value: unknown) {
  const normalized = textValue(value, '').toUpperCase()
  const labels: Record<string, string> = {
    DIRECT_ADVERSE_CURRENT_CLIENT: '当前对方为本所现有客户',
    DIRECT_ADVERSE_FORMER_CLIENT: '当前对方为本所既往客户',
    RELATED_PARTY_CONFLICT: '关联方利益冲突',
    CONFLICT_OPPOSING_PARTY: '对方当事人与既有客户冲突',
    'CONFLICT-OPPOSING-001': '对方当事人与既有客户冲突',
    'CONFLICT-DIRECT-001': '直接利益冲突',
  }
  return labels[normalized] || textValue(value, '未标注冲突类型')
}

function conflictDispositionLabel(value: unknown) {
  const labels: Record<string, string> = {
    CLEAR: '未发现可识别冲突',
    REVIEW_REQUIRED: '待人工复核',
    BLOCKED: '暂停接案',
    WAIVER_PENDING: '豁免评估中',
    WAIVED: '已按条件豁免',
    no_conflict: '无冲突',
    confirmed_conflict: '确认冲突',
    false_positive: '误报',
    insufficient_information: '信息不足',
    waiver_requested: '申请豁免',
  }
  const raw = textValue(value, '')
  return labels[raw] || labels[raw.toUpperCase()] || raw || '未形成结论'
}

type ConflictDecisionStatus =
  | 'UNTESTED'
  | 'STALE'
  | 'CLEAR'
  | 'REVIEW_REQUIRED'
  | 'BLOCKED'
  | 'WAIVER_PENDING'
  | 'WAIVED'

interface ConflictDecisionViewModel {
  decision: ConflictDecisionStatus
  risk: string
  requiresApproval: boolean
  stale: boolean
  review: Record<string, unknown>
  waiver: Record<string, unknown>
  canSubmit: boolean
  canRequestWaiver: boolean
  showReviewAction: boolean
  needsHumanReview: boolean
  machineBlocked: boolean
  coverageLimited: boolean
  nonWaivableDirectConflict: boolean
  headline: string
  guidance: string
}

function deriveConflictDecisionViewModel(
  conflict: Record<string, any> | null | undefined,
  options: {
    stale?: boolean
    review?: Record<string, unknown>
    waiver?: Record<string, unknown>
  } = {},
): ConflictDecisionViewModel {
  const checkResult = recordValue(conflict?.check_result || conflict)
  const assessment = recordValue(checkResult.riskAssessment)
  const review =
    Object.keys(options.review || {}).length > 0
      ? recordValue(options.review)
      : recordValue(checkResult.review)
  const waiver =
    Object.keys(options.waiver || {}).length > 0
      ? recordValue(options.waiver)
      : recordValue(checkResult.waiver)
  const risk = textValue(
    assessment.overallRisk || conflict?.risk_level || conflict?.record?.risk_level,
    conflict ? 'LOW' : '',
  ).toUpperCase()
  const waiverStatus = textValue(waiver.status, '').toUpperCase()
  const reviewDecision = textValue(review.decision, '').toLowerCase()
  const rawDecision = textValue(
    recordValue(checkResult.decision).status || conflict?.decision?.status,
    '',
  ).toUpperCase()
  const nonWaivableDirectConflict = conflictHasNonWaivableDirectConflict(conflict)
  const coverageStatus = textValue(
    recordValue(checkResult.decision).coverageStatus ||
      recordValue(checkResult.decision).coverage_status ||
      conflict?.coverage_status,
    conflict ? 'COVERAGE_LIMITED' : '',
  ).toUpperCase()
  const coverageLimited = Boolean(conflict && coverageStatus !== 'COMPLETE')
  const conflictCases = listOf<Record<string, unknown>>(
    checkResult.conflictCases as Array<Record<string, unknown>> | undefined,
  )
  const legacyConflictCases = listOf<Record<string, unknown>>(conflict?.conflict_cases)
  const hitCount = conflictCases.length || legacyConflictCases.length
  const hasConflict = Boolean(conflict?.has_conflict ?? conflict?.hasConflict ?? hitCount > 0)
  const requiresApproval = Boolean(
    assessment.requiresApproval ??
      checkResult.requiresApproval ??
      ['HIGH', 'CRITICAL'].includes(risk),
  )

  let decision: ConflictDecisionStatus = conflict ? 'REVIEW_REQUIRED' : 'UNTESTED'
  if (options.stale) decision = 'STALE'
  else if (waiverStatus === 'APPROVED' || rawDecision === 'WAIVED') decision = 'WAIVED'
  else if (
    ['UNDER_REVIEW', 'SUBMITTED', 'PENDING'].includes(waiverStatus) ||
    rawDecision === 'WAIVER_PENDING'
  )
    decision = 'WAIVER_PENDING'
  else if (
    ['REJECTED', 'EXPIRED', 'REVOKED'].includes(waiverStatus) ||
    reviewDecision === 'confirmed_conflict'
  )
    decision = 'BLOCKED'
  else if (['no_conflict', 'false_positive'].includes(reviewDecision)) decision = 'CLEAR'
  else if (['NO_MATCH_FOUND', 'NO_CONFLICT', 'NO_MATCH'].includes(rawDecision) && !hasConflict)
    decision = 'CLEAR'
  else if (rawDecision === 'BLOCKED') decision = 'BLOCKED'
  else if (rawDecision === 'CLEAR') decision = 'CLEAR'
  else if (
    rawDecision === 'REVIEW_REQUIRED' ||
    reviewDecision === 'insufficient_information' ||
    conflict
  )
    decision = 'REVIEW_REQUIRED'

  // A reviewer cannot turn an incomplete archive scope into permission to
  // proceed. The backend enforces the same rule immediately before approval;
  // keeping it here prevents a misleading green action in the browser.
  if (coverageLimited && decision === 'CLEAR') decision = 'REVIEW_REQUIRED'

  const machineBlocked =
    decision === 'BLOCKED' && rawDecision === 'BLOCKED' && !reviewDecision && !waiverStatus
  const needsHumanReview = decision === 'REVIEW_REQUIRED' || machineBlocked

  const copy: Record<ConflictDecisionStatus, { headline: string; guidance: string }> = {
    UNTESTED: {
      headline: '尚未完成冲突检测',
      guidance: '保存最新输入并检测后，系统才会给出接案决策。',
    },
    STALE: {
      headline: '冲突检测结果已过期',
      guidance: '客户、对方、相关方、案件或负责律师已变化，请保存最新输入并重新检测。',
    },
    CLEAR: {
      headline: '可提交立案审批',
      guidance: '未发现需要阻止接案的冲突，可以继续提交立案审批。',
    },
    REVIEW_REQUIRED: {
      headline: '需要独立人工复核',
      guidance: '当前命中尚未形成最终结论，请进入本案冲突复核。',
    },
    BLOCKED: {
      headline: '已暂停接案',
      guidance: '已确认存在冲突，完成合规复核前不得提交立案审批。',
    },
    WAIVER_PENDING: {
      headline: '等待独立复核',
      guidance: '豁免申请正在由独立复核人处理，结果形成前无需再次申请或发起冲突审批。',
    },
    WAIVED: {
      headline: '按批准条件继续',
      guidance: '豁免已批准，请在批准条件和有效期限内继续办理。',
    },
  }

  const decisionCopy = machineBlocked
    ? {
        headline: '机器检测已自动阻断，等待独立复核',
        guidance: '精确命中已禁止立案，但尚未人工定案。请进入本案冲突复核，由独立复核人确认。',
      }
    : coverageLimited && decision === 'REVIEW_REQUIRED'
      ? {
          headline: '检索范围受限，需人工复核',
          guidance:
            '部分权威档案或关联信息尚未登记为完整覆盖，不能据此确认无冲突；请由冲突核查人补充核查或处理例外。',
        }
      : copy[decision]

  const resolvedDecisionCopy =
    reviewDecision === 'insufficient_information'
      ? {
          headline: '信息不足，暂停接案',
          guidance: '独立复核已完成，结论为信息不足；在审批流程完成处置前不得继续接案。',
        }
      : decisionCopy

  return {
    decision,
    risk,
    requiresApproval,
    stale: Boolean(options.stale),
    review,
    waiver,
    canSubmit: (decision === 'CLEAR' && !coverageLimited) || decision === 'WAIVED',
    canRequestWaiver:
      decision === 'BLOCKED' && !machineBlocked && !waiver.id && !nonWaivableDirectConflict,
    showReviewAction: needsHumanReview,
    needsHumanReview,
    machineBlocked,
    coverageLimited,
    nonWaivableDirectConflict,
    ...resolvedDecisionCopy,
  }
}

function conflictDecisionStatusLabel(
  view: Pick<ConflictDecisionViewModel, 'decision' | 'coverageLimited' | 'risk'>,
) {
  if (view.decision === 'STALE') return '结果已过期'
  if (view.decision === 'REVIEW_REQUIRED')
    return view.coverageLimited ? '范围受限，待人工复核' : '待独立人工复核'
  if (view.decision === 'BLOCKED') return '已暂停接案'
  if (view.decision === 'WAIVER_PENDING') return '等待独立复核'
  if (view.decision === 'WAIVED') return '已按条件豁免'
  if (view.decision === 'CLEAR') return '已复核：未发现冲突'
  return riskLabel(view.risk || 'LOW')
}

function conflictHasNonWaivableDirectConflict(conflict: Record<string, any> | null | undefined) {
  const checkResult = recordValue(conflict?.check_result || conflict)
  const directRuleCodes = new Set([
    'DIRECT_ADVERSE_CURRENT_CLIENT',
    'STRUCTURED_IDENTITY_EXACT',
    'DIRECT_CONFLICT',
  ])
  const visit = (value: unknown): boolean => {
    if (Array.isArray(value)) return value.some(visit)
    if (!value || typeof value !== 'object') return false
    const record = value as Record<string, unknown>
    const ruleCode = textValue(record.ruleCode || record.rule_code, '').toUpperCase()
    const conflictType = textValue(record.conflictType || record.conflict_type, '').toUpperCase()
    const matchType = textValue(record.matchType || record.match_type, '').toUpperCase()
    const partyRole = textValue(record.partyRole || record.party_role, '').toUpperCase()
    const historicalRole = textValue(
      record.historicalRole || record.historical_role,
      '',
    ).toUpperCase()
    if (
      directRuleCodes.has(ruleCode) ||
      conflictType.includes('DIRECT') ||
      conflictType.includes('直接冲突')
    )
      return true
    if (matchType === 'EXACT' && partyRole === 'OPPOSING_PARTY' && historicalRole === 'CLIENT')
      return true
    return Object.values(record).some(visit)
  }
  return visit(checkResult)
}

function primaryConflictEvidence(conflict: Record<string, any> | null | undefined) {
  const checkResult = recordValue(conflict?.check_result || conflict)
  const assessment = recordValue(checkResult.riskAssessment)
  const structured = recordValue(
    assessment.primaryEvidence ||
      checkResult.primaryEvidence ||
      assessment.matchEvidence ||
      checkResult.matchEvidence,
  )
  if (Object.keys(structured).length > 0) return structured
  const directEvidence = listOf<Record<string, unknown>>(
    (assessment.evidence || checkResult.evidence || conflict?.evidence) as
      | Array<Record<string, unknown>>
      | undefined,
  )
  if (directEvidence.length > 0) return directEvidence[0]
  const conflictCases = listOf<Record<string, unknown>>(
    checkResult.conflictCases as Array<Record<string, unknown>> | undefined,
  )
  const legacyCases =
    conflictCases.length > 0
      ? conflictCases
      : listOf<Record<string, unknown>>(conflict?.conflict_cases)
  return (
    legacyCases.flatMap((item) =>
      listOf<Record<string, unknown>>(item.evidence as Array<Record<string, unknown>> | undefined),
    )[0] || {}
  )
}

function conflictHitSubject(
  conflict: CommandCenterRiskItem | Record<string, any> | null | undefined,
) {
  const record = conflict as Record<string, any> | null | undefined
  const restricted =
    textValue(record?.source_case || record?.sourceCase, '') === '受限' ||
    textValue(record?.evidence_summary || record?.evidenceSummary, '').includes('受隔离')
  if (restricted) return '存在受限命中'
  const decision = deriveConflictDecisionViewModel(record).decision
  if (decision === 'CLEAR') {
    return '无命中主体'
  }
  const evidence = primaryConflictEvidence(record)
  const result = recordValue(record?.check_result)
  const explicitSubject = firstPresent(
    evidence.requestedParty,
    evidence.queryName,
    evidence.matchedSubject,
    evidence.subjectName,
    result.matchedSubject,
    result.hitSubject,
  )
  if (explicitSubject !== undefined && explicitSubject !== null && explicitSubject !== '') {
    return textValue(explicitSubject)
  }
  return (record?.has_conflict ?? record?.hasConflict) ? '待人工核实主体' : '未发现匹配记录'
}

export function conflictRecordMatchType(
  conflict: CommandCenterRiskItem | Record<string, any> | null | undefined,
) {
  const record = conflict as Record<string, any> | null | undefined
  const restricted =
    textValue(record?.source_case || record?.sourceCase, '') === '受限' ||
    textValue(record?.evidence_summary || record?.evidenceSummary, '').includes('受隔离')
  if (restricted) return '受限记录'
  const decision = deriveConflictDecisionViewModel(record).decision
  if (decision === 'CLEAR') return '无命中'
  const evidence = primaryConflictEvidence(record)
  const hasConflict = record?.has_conflict ?? record?.hasConflict
  if ((hasConflict === false || hasConflict === 0) && Object.keys(evidence).length === 0) {
    return '无匹配'
  }
  const result = recordValue(record?.check_result)
  const rawMatchType = firstPresent(
    evidence.matchType,
    evidence.algorithm,
    record?.matched_type,
    result.matchType,
  )
  return rawMatchType ? conflictMatchTypeLabel(rawMatchType) : '未标注匹配方式'
}

function conflictRecordEvidenceSummary(
  conflict: CommandCenterRiskItem | Record<string, any> | null | undefined,
) {
  const record = conflict as Record<string, any> | null | undefined
  if (deriveConflictDecisionViewModel(record).decision === 'CLEAR') return '未发现可识别冲突'
  const evidence = primaryConflictEvidence(record)
  return textValue(record?.evidence_summary || evidence.summary, '暂无结构化证据摘要')
}

function conflictConfidenceLabel(item: CommandCenterRiskItem) {
  const result = recordValue(item.check_result)
  const assessment = recordValue(result.riskAssessment)
  const evidence = recordValue(assessment.matchEvidence || result.matchEvidence)
  const parameters = recordValue(item.search_parameters)
  const automaticConclusion = evidence.automaticConclusion ?? parameters.automaticConclusion
  const matchType = textValue(evidence.matchType || parameters.matchMode, '').toUpperCase()
  if (automaticConclusion === false || matchType === 'NAME_CANDIDATE') return '待人工核实'
  const raw =
    evidence.similarity ?? evidence.confidence ?? parameters.similarity ?? parameters.confidence
  if (raw === undefined || raw === null || raw === '') return '-'
  const score = numberValue(raw, 0)
  return `${formatRiskScore(score > 0 && score <= 1 ? score * 100 : score)}%`
}

function contextRecordCandidates(item: CommandCenterRiskItem) {
  const itemRecord = item as Record<string, any>
  const checkResult = recordValue(itemRecord.check_result || itemRecord.checkResult)
  return [
    itemRecord,
    recordValue(itemRecord.search_parameters || itemRecord.searchParameters),
    recordValue(itemRecord.request || itemRecord.request_payload),
    recordValue(checkResult.searchParameters || checkResult.search_parameters),
    recordValue(checkResult.subjectCase || checkResult.subject_case),
  ]
}

function contextIdentity(record: Record<string, any>) {
  return {
    id: textValue(
      firstPresent(
        record.case_id,
        record.caseId,
        record.subject_case_id,
        record.subjectCaseId,
        record.subject_case?.id,
        record.subjectCase?.id,
      ),
      '',
    ),
    number: textValue(
      firstPresent(
        record.case_number,
        record.caseNumber,
        record.case_no,
        record.caseNo,
        record.subject_case_number,
        record.subjectCaseNumber,
        record.subject_case?.case_number,
        record.subjectCase?.caseNumber,
      ),
      '',
    ),
    title: textValue(
      firstPresent(
        record.title,
        record.case_title,
        record.caseTitle,
        record.subject_case_title,
        record.subjectCaseTitle,
        record.subject_case?.title,
        record.subjectCase?.title,
      ),
      '',
    ),
  }
}

export function conflictMatchesCaseContext(
  item: CommandCenterRiskItem,
  caseID: string,
  caseNumber: string,
  caseTitle: string,
) {
  const candidates = contextRecordCandidates(item).map(contextIdentity)
  if (caseID && candidates.some((candidate) => candidate.id === caseID)) return true
  if (caseNumber && candidates.some((candidate) => candidate.number === caseNumber)) return true
  // Once an explicit case identifier is present, a title-only match is unsafe.
  // conflictCases are deliberately excluded: their IDs identify historical hit
  // matters, not the subject case being checked.
  if (caseID || caseNumber) return false
  return Boolean(caseTitle && candidates.some((candidate) => candidate.title === caseTitle))
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
    overdue: '逾期',
  }
  return labels[normalized] || value || '中'
}

function statusLabel(value?: string) {
  const labels: Record<string, string> = {
    active: '办理中',
    pending: '待处理',
    in_progress: '办理中',
    completed: '已完成',
    received: '已收齐',
    missing: '待补充',
    uploaded: '已上传',
    cancelled: '已取消',
    archived: '已归档',
    draft: '草稿',
    submitted: '已提交',
    under_review: '审批中',
    approved: '已通过',
    rejected: '已拒绝',
    resubmitted: '已重新提交',
    needs_revision: '待补充',
    request_changes: '退回修改',
    approve: '同意',
    reject: '拒绝',
    conflict_ready: '待冲突复核',
    conflict_checking: '冲突检测中',
    risk_review: '冲突复核中',
    QUEUED: '排队中',
    RUNNING: '检测中',
    PROCESSING: '处理中',
    FAILED: '检测失败',
    COMPLETED: '已完成',
  }
  const raw = value || ''
  return labels[raw] || labels[raw.toLowerCase()] || value || '未知'
}

function approvalStageLabel(value: unknown) {
  const normalized = textValue(value, '').toLowerCase()
  const labels: Record<string, string> = {
    initial_review: '初审',
    director_review: '主任复核',
    compliance_review: '合规复核',
    management_approval: '管理层审批',
    department_review: '部门复核',
    conflict_review: '冲突复核',
    final_review: '最终复核',
    final_approval: '最终决定',
    conflict_approval: '冲突审批流程',
    case_intake: '接案审批流程',
  }
  return labels[normalized] || textValue(value, '审批节点')
}

function clientSourceLabel(value: unknown) {
  const normalized = textValue(value, '').toLowerCase()
  const labels: Record<string, string> = {
    trial_seed: '试用演示数据',
    lawyer_trial_acceptance_seed: '律师验收演示数据',
    referral: '客户转介',
    online: '线上咨询',
    manual: '人工录入',
  }
  return labels[normalized] || textValue(value)
}

function relationshipTypeLabel(value: unknown) {
  const normalized = textValue(value, '').toLowerCase()
  const labels: Record<string, string> = {
    opposing_party: '对方当事人',
    client: '客户',
    affiliate: '关联方',
    shareholder: '股东',
    controller: '实际控制人',
    legal_representative: '法定代表人',
  }
  return labels[normalized] || textValue(value, '未标注关系')
}

function workItemTypeLabel(value: unknown) {
  const normalized = textValue(value, '').toLowerCase()
  const labels: Record<string, string> = {
    case: '案件任务',
    task: '普通任务',
    approval: '审批待办',
    approval_request: '审批待办',
    deadline: '期限提醒',
    reminder: '事项提醒',
    conflict: '冲突复核',
  }
  return labels[normalized] || textValue(value, '待办事项')
}

function materialTypeLabel(value: unknown) {
  const normalized = textValue(value, '').toLowerCase()
  const labels: Record<string, string> = {
    identity: '主体身份证明',
    contract: '合同材料',
    evidence: '证据材料',
    application: '申请材料',
    conflict_report: '冲突检测报告',
    proof: '证明材料',
    document: '其他文档',
  }
  return labels[normalized] || textValue(value, '其他材料')
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

export interface IntakeFormState {
  intakeId: string
  intakeCode: string
  idempotencyKey: string
  title: string
  caseType: string
  businessArea: string
  subArea: string
  priority: string
  disputeAmount: string
  sourceChannel: string
  sourceContact: string
  investmentAgreementDate: string
  disputeDate: string
  breachDate: string
  proposedFilingDate: string
  jurisdiction: string
  clientId: number
  clientName: string
  opponentName: string
  opponentEntityType: string
  opponentIdentityType: string
  opponentIdentityNumber: string
  opponentAliases: string
  description: string
  billingMethod: string
  feeBase: string
  contingencyRate: string
  minimumFee: string
  lawyerId: number
}

export type ConflictCheckRequiredField =
  | 'title'
  | 'client'
  | 'opponentName'
  | 'opponentIdentityNumber'
  | 'lawyer'
  | 'caseType'
  | 'businessArea'
  | 'subArea'

export const conflictCheckRequiredFieldLabels: Record<ConflictCheckRequiredField, string> = {
  title: '案件名称',
  client: '客户',
  opponentName: '对方当事人',
  opponentIdentityNumber: '对方身份标识',
  lawyer: '负责律师',
  caseType: '案件类型',
  businessArea: '业务领域',
  subArea: '子领域',
}

const hasConflictCheckValue = (value: string | number | null | undefined) =>
  typeof value === 'number' ? value > 0 : String(value || '').trim().length > 0

export function getMissingConflictCheckFields(
  form: Pick<
    IntakeFormState,
    | 'title'
    | 'clientId'
    | 'clientName'
    | 'opponentName'
    | 'opponentIdentityNumber'
    | 'lawyerId'
    | 'caseType'
    | 'businessArea'
    | 'subArea'
  >,
): ConflictCheckRequiredField[] {
  const missing: ConflictCheckRequiredField[] = []
  if (!hasConflictCheckValue(form.title)) missing.push('title')
  if (!hasConflictCheckValue(form.clientId) || !hasConflictCheckValue(form.clientName)) {
    missing.push('client')
  }
  if (!hasConflictCheckValue(form.opponentName)) missing.push('opponentName')
  if (!hasConflictCheckValue(form.opponentIdentityNumber)) {
    missing.push('opponentIdentityNumber')
  }
  if (!hasConflictCheckValue(form.lawyerId)) missing.push('lawyer')
  if (!hasConflictCheckValue(form.caseType)) missing.push('caseType')
  if (!hasConflictCheckValue(form.businessArea)) missing.push('businessArea')
  if (!hasConflictCheckValue(form.subArea)) missing.push('subArea')
  return missing
}

function conflictCheckFieldForFormKey(
  key: keyof IntakeFormState,
): ConflictCheckRequiredField | undefined {
  if (key === 'clientId' || key === 'clientName') return 'client'
  if (key === 'lawyerId') return 'lawyer'
  if (
    key === 'title' ||
    key === 'opponentName' ||
    key === 'opponentIdentityNumber' ||
    key === 'caseType' ||
    key === 'businessArea' ||
    key === 'subArea'
  ) {
    return key
  }
  return undefined
}

interface IntakeRelatedParty {
  name: string
  role: string
  entityType: string
  identityType: string
  identityNumber: string
}

interface IntakeRuntimeState {
  intake?: any
  conflict?: any
  conflictTask?: any
  conflictInputFingerprint?: string
  approval?: any
  integrationStatus?: any
  apiTimings: Array<{ label: string; duration: number; at: string }>
}

interface ClientOption {
  id: number
  name: string
  email?: string
  phone?: string
  displayLabel: string
}

interface LawyerOption {
  id: number
  name: string
  department?: string
  seniority?: string
  position?: string
}

const defaultIntakeForm: IntakeFormState = {
  intakeId: '',
  intakeCode: '',
  idempotencyKey: '',
  title: '',
  caseType: '',
  businessArea: '',
  subArea: '',
  priority: 'medium',
  disputeAmount: '',
  sourceChannel: '',
  sourceContact: '',
  investmentAgreementDate: '',
  disputeDate: '',
  breachDate: '',
  proposedFilingDate: '',
  jurisdiction: '',
  clientId: 0,
  clientName: '',
  opponentName: '',
  opponentEntityType: 'LEGAL_PERSON',
  opponentIdentityType: 'SOCIAL_CREDIT_CODE',
  opponentIdentityNumber: '',
  opponentAliases: '',
  description: '',
  billingMethod: '',
  feeBase: '',
  contingencyRate: '',
  minimumFee: '',
  lawyerId: 0,
}

const legacyCaseIntakeDraftKey = 'law-oa-case-intake-draft-v1'

export function scopedCaseIntakeDraftKey(userID: string, caseID?: string) {
  return `law-oa-case-intake-draft-v2:${userID || 'anonymous'}:${caseID ? `case-${caseID}` : 'new'}`
}

export function caseIntakeConflictFingerprint(
  form: Pick<
    IntakeFormState,
    | 'clientId'
    | 'clientName'
    | 'opponentName'
    | 'opponentEntityType'
    | 'opponentIdentityType'
    | 'opponentIdentityNumber'
    | 'opponentAliases'
    | 'title'
    | 'caseType'
    | 'lawyerId'
  >,
  relatedParties: IntakeRelatedParty[],
) {
  return JSON.stringify({
    clientId: form.clientId,
    clientName: form.clientName.trim(),
    opponentName: form.opponentName.trim(),
    opponentEntityType: form.opponentEntityType,
    opponentIdentityType: form.opponentIdentityType,
    opponentIdentityNumber: form.opponentIdentityNumber.trim(),
    opponentAliases: form.opponentAliases.trim(),
    title: form.title.trim(),
    caseType: form.caseType,
    lawyerId: form.lawyerId,
    relatedParties: relatedParties
      .map((party) => ({
        name: party.name.trim(),
        role: party.role,
        entityType: party.entityType,
        identityType: party.identityType,
        identityNumber: party.identityNumber.trim(),
      }))
      .sort((a, b) => a.name.localeCompare(b.name)),
  })
}

function loadCaseIntakeDraft(draftKey: string): IntakeFormState | null {
  if (typeof window === 'undefined') {
    return null
  }
  try {
    const raw = window.localStorage.getItem(draftKey)
    if (!raw) {
      return null
    }
    const parsed = JSON.parse(raw) as Partial<IntakeFormState>
    return {
      ...defaultIntakeForm,
      ...parsed,
      // Identity values are deliberately never restored from browser storage.
      opponentIdentityNumber: '',
      clientId: Number(parsed.clientId || 0),
      lawyerId: Number(parsed.lawyerId || 0),
    }
  } catch {
    return null
  }
}

export function persistableCaseIntakeDraft(form: IntakeFormState) {
  return {
    ...form,
    // A shared workstation must not retain a client's or opponent's identity
    // number in localStorage. The lawyer re-enters it before confirmation.
    opponentIdentityNumber: '',
  }
}

const defaultMaterials = [
  { name: '客户主体资料', material_type: 'identity', status: 'missing', required: true },
  { name: '投资协议及补充协议', material_type: 'contract', status: 'missing', required: true },
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
    商事: '商事',
    民事: '民事',
    民事诉讼: '民事诉讼',
    建设工程: '建设工程',
    劳动: '劳动',
    知识产权: '知识产权',
    刑事: '刑事',
    行政: '行政',
  }
  return labels[value] || '其他'
}

const intakeCaseTypeOptions = [
  { value: 'commercial', label: '商事诉讼' },
  { value: 'civil', label: '民事' },
  { value: 'civil_litigation', label: '民事诉讼' },
  { value: 'construction', label: '建设工程' },
  { value: 'labor', label: '劳动争议' },
  { value: 'intellectual', label: '知识产权' },
  { value: 'criminal', label: '刑事' },
  { value: 'administrative', label: '行政' },
  { value: 'financial', label: '金融商事' },
  { value: 'ma', label: '并购重组' },
]

function normalizeClientOptions(data: any): ClientOption[] {
  const rows = Array.isArray(data)
    ? data
    : data?.clients || data?.list || data?.data?.clients || data?.data?.list || data?.data || []

  const clients: ClientOption[] = rows
    .map((item: any) => ({
      id: numberValue(item.id),
      name: textValue(item.name, ''),
      email: textValue(item.email, ''),
      phone: textValue(item.phone, ''),
      displayLabel: '',
    }))
    .filter((item: ClientOption) => item.id > 0 && item.name)
  const nameCounts = clients.reduce<Record<string, number>>((counts, client) => {
    counts[client.name] = (counts[client.name] || 0) + 1
    return counts
  }, {})
  return clients.map((client) => ({
    ...client,
    displayLabel:
      nameCounts[client.name] > 1
        ? `${client.name} · ${client.email || client.phone || `客户编号 ${client.id}`}`
        : client.name,
  }))
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
    const error =
      body.message || (typeof body.error === 'string' ? body.error : body.error?.message)
    throw new Error(error || `API 请求失败：${response.status}`)
  }
  return (body.data ?? body) as T
}

async function fetchCommandCenter(
  signal: AbortSignal,
  includeAllConflicts = false,
): Promise<CommandCenterPayload | null> {
  const token = getToken()
  const response = await fetch(
    `/api/v1/dashboard/command-center${includeAllConflicts ? '?conflict_scope=all' : ''}`,
    {
      signal,
      headers: token ? { Authorization: `Bearer ${token}` } : undefined,
    },
  )
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
  blue: 'var(--color-primary-500)',
  teal: '#12a89d',
  red: '#e8434e',
  orange: '#f59f2f',
  green: '#18a058',
  slate: '#50657d',
}

function MetricCard({ icon, label, value, delta, tone }: MetricCardProps) {
  return (
    <section
      className={`kpi-card ${
        tone === 'red'
          ? 'danger'
          : tone === 'orange'
            ? 'warn'
            : tone === 'green' || tone === 'teal'
              ? 'ok'
              : ''
      }`}
    >
      <div className='kpi-label'>
        {icon}
        {label}
      </div>
      <div className='kpi-val'>{value}</div>
      <div className='kpi-foot'>{delta}</div>
    </section>
  )
}

function RiskTag({ text }: { text: string }) {
  const cls =
    text.includes('高') || text.includes('直接') || text.includes('紧急')
      ? 'ng-risk h'
      : text.includes('中') || text.includes('潜在')
        ? 'ng-risk m'
        : 'ng-risk l'
  return <span className={cls}>{text}</span>
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
    <header className='ng-page-head'>
      <div>
        <div className='eyebrow'>{eyebrow}</div>
        <h1>{title}</h1>
        {subtitle && <p className='sub'>{subtitle}</p>}
      </div>
      {actions && <div className='ng-head-actions'>{actions}</div>}
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
    <section className={`ng-panel ${className}`}>
      <div className='ng-panel-head'>
        <h2 className='ng-panel-title'>{title}</h2>
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
    return {
      label: '加载中',
      approverIds: [] as string[],
      approverEmails: [] as string[],
      availableActions: [] as string[],
      canApprove: false,
      canReject: false,
      canReturn: false,
      canDecide: false,
      readonlyReason: '审批数据加载中',
    }
  }
  const userInfo = getUserInfo() || {}
  const currentUserId = textValue(userInfo.id, '')
  const approverName = textValue(
    approval.current_approver_name || approval.currentApproverName,
    '待分配',
  )
  const approverIds = [
    textValue(approval.current_approver_id || approval.currentApproverId, ''),
  ].filter(Boolean)
  const approverEmails = [
    textValue(approval.current_approver_email || approval.currentApproverEmail, ''),
  ].filter(Boolean)
  const availableActions = listOf<string>(approval.available_actions || approval.availableActions)
  const applicantId = textValue(approval.applicant_id || approval.applicantId, '')
  const isApplicant = Boolean(currentUserId && applicantId === currentUserId)
  const canApprove =
    !isApplicant &&
    (availableActions.length > 0
      ? availableActions.includes('approve')
      : Boolean(currentUserId && approverIds.includes(currentUserId)))
  const canReject = availableActions.length > 0 ? availableActions.includes('reject') : canApprove
  const canReturn =
    availableActions.length > 0 ? availableActions.includes('request_changes') : canApprove
  const canDecide = canApprove || canReject || canReturn
  const readonlyReason = canDecide
    ? undefined
    : isApplicant
      ? '申请人不能审批自己的申请，仅可查看审批进度。'
      : `当前审批人：${approverName}。当前账号仅可查看审批进度。`
  return {
    label: approverName,
    approverIds,
    approverEmails,
    availableActions,
    canApprove,
    canReject,
    canReturn,
    canDecide,
    readonlyReason,
  }
}

function StatusDot({ color }: { color: Tone }) {
  return <span className='batch-status-dot' style={{ backgroundColor: toneIcon[color] }} />
}

export function DashboardCommandCenter() {
  const navigate = useNavigate()
  const currentUserInfo = getUserInfo()
  const canViewFinance = ['admin', 'super_admin', 'finance'].includes(
    textValue(currentUserInfo?.role, '').toLowerCase(),
  )
  const [commandCenter, setCommandCenter] = React.useState<CommandCenterPayload | null>(null)
  const [financeOverview, setFinanceOverview] = React.useState<FinanceOverviewPayload | null>(null)
  const [apiLoading, setApiLoading] = React.useState(false)
  const [apiError, setApiError] = React.useState(false)
  const [apiRefreshKey, setApiRefreshKey] = React.useState(0)
  const [todoFilter, setTodoFilter] = React.useState('全部')
  const [globalSearchQuery, setGlobalSearchQuery] = React.useState('')
  const [globalSearchState, setGlobalSearchState] = React.useState<{
    status: 'idle' | 'results' | 'empty'
    message: string
    results: Array<{ key: string; type: string; label: string; detail: string; path: string }>
  }>({ status: 'idle', message: '输入关键词后按 Enter 搜索', results: [] })

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
    if (canViewFinance) {
      apiRequest<FinanceOverviewPayload>('/finance/overview')
        .then((payload) => setFinanceOverview(payload))
        .catch(() => setFinanceOverview(null))
    } else {
      setFinanceOverview(null)
    }
    return () => controller.abort()
  }, [apiRefreshKey, canViewFinance])

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
  const currentUserName = textValue(
    currentUserInfo?.realName || currentUserInfo?.name || currentUserInfo?.username,
    '律师',
  )
  const pendingApprovals = summary?.pending_approvals ?? 0
  const openConflicts = summary?.open_conflict_tasks ?? 0
  const activeCases = summary?.active_cases ?? 0
  const unreadInbox = summary?.unread_inbox ?? 0
  const financeWarningAmount =
    numberValue(financeOverview?.payment_stats?.pending_amount) ||
    numberValue(financeOverview?.invoice_stats?.pending_invoice_amount)
  const apiStatusText = apiError
    ? '接口异常'
    : apiLoading
      ? '连接中'
      : commandCenter
        ? '已连接'
        : '等待连接'
  const apiStatusTone: Tone = apiError ? 'red' : commandCenter ? 'green' : 'slate'
  const dashboardMetrics = metrics
    .filter((item) => canViewFinance || item.label !== '合同回款预警')
    .map((item) => {
      if (item.label === '待办事项') return { ...item, value: unreadInbox }
      if (item.label === '利益冲突待复核') return { ...item, value: openConflicts }
      if (item.label === '待审批事项') return { ...item, value: pendingApprovals }
      if (item.label === '接案准备中') return { ...item, value: workflow?.intake ?? item.value }
      if (item.label === '在办案件总数') return { ...item, value: activeCases }
      if (item.label === '逾期任务') return { ...item, value: overdueItems.length }
      if (item.label === '合同回款预警') return { ...item, value: financeWarningAmount, delta: '' }
      return item
    })
  const urgentTodoCount = todoItems.filter((item) =>
    ['critical', 'high'].includes((item.priority || '').toLowerCase()),
  ).length
  const approvalTodoCount = todoItems.filter((item) =>
    ['approval', 'approval_request'].includes((item.type || item.source_type || '').toLowerCase()),
  ).length
  const filteredTodoItems = todoItems.filter((item) => {
    if (todoFilter === '紧急')
      return ['critical', 'high'].includes(textValue(item.priority, '').toLowerCase())
    if (todoFilter === '审批')
      return ['approval', 'approval_request'].includes(
        textValue(item.type || item.source_type, '').toLowerCase(),
      )
    if (todoFilter === '任务')
      return !['approval', 'approval_request'].includes(
        textValue(item.type || item.source_type, '').toLowerCase(),
      )
    return true
  })
  const runGlobalSearch = (searchValue = globalSearchQuery) => {
    const query = searchValue.trim().toLowerCase()
    if (!query) {
      setGlobalSearchState({ status: 'idle', message: '输入关键词后按 Enter 搜索', results: [] })
      return
    }
    const results = [
      ...caseRowsLive
        .filter((item) =>
          `${item.title || ''} ${item.client_name || ''} ${item.case_number || ''}`
            .toLowerCase()
            .includes(query),
        )
        .map((item) => ({
          key: `case-${item.id}`,
          type: '案件',
          label: textValue(item.title, '未命名案件'),
          detail: textValue(item.case_number || item.client_name, ''),
          path: `/case/${item.id}`,
        })),
      ...approvalItems
        .filter((item) =>
          `${item.title || ''} ${item.request_number || ''} ${item.current_approver_name || ''}`
            .toLowerCase()
            .includes(query),
        )
        .map((item) => ({
          key: `approval-${item.id}`,
          type: '审批',
          label: textValue(item.title, '未命名审批'),
          detail: textValue(item.request_number, ''),
          path: `/approval/${item.id}`,
        })),
      ...riskItems
        .filter((item) =>
          `${item.title || ''} ${item.client_name || ''} ${item.id || ''}`
            .toLowerCase()
            .includes(query),
        )
        .map((item) => ({
          key: `conflict-${item.id}`,
          type: '冲突检测',
          label: textValue(item.title, '未命名检测'),
          detail: riskLabel(item.risk_level),
          path: `/conflict?task_id=${encodeURIComponent(textValue(item.id, ''))}`,
        })),
    ]
    setGlobalSearchState(
      results.length > 0
        ? {
            status: 'results',
            message: `找到 ${results.length} 条相关案件、冲突或审批记录`,
            results,
          }
        : { status: 'empty', message: '未找到相关案件、冲突或审批记录', results: [] },
    )
  }

  return (
    <div className='batch-page dashboard-page ng-content'>
      <header className='ng-hero'>
        <div>
          <h1>
            上午好，<em>{currentUserName}</em>
          </h1>
          <p>
            <span>数据库案件 {commandCenter ? activeCases : '-'}</span>
            <span className='sep' />
            <span>冲突任务 {commandCenter ? openConflicts : '-'}</span>
            <span className='sep' />
            <span>待审批 {commandCenter ? pendingApprovals : '-'}</span>
          </p>
        </div>
        <div className='ng-hero-date'>
          <b>{formatTodayText()}</b>
          <span>
            数据源 <StatusDot color={apiStatusTone} /> {apiStatusText} ·{' '}
            {formatDateTime(commandCenter?.generated_at)}
          </span>
          <span style={{ display: 'flex', gap: 8, marginTop: 8, justifyContent: 'flex-end' }}>
            <Badge color={apiError ? '#e8434e' : '#12a89d'} text='实时数据' />
            <Button size='small' onClick={refreshCommandCenter} loading={apiLoading}>
              刷新接口
            </Button>
            <Button
              size='small'
              type='primary'
              icon={<PlusOutlined />}
              onClick={() => navigate('/case/create')}
            >
              新建立案
            </Button>
          </span>
        </div>
      </header>

      <section className='ng-hero-search' style={{ display: 'flex', gap: 10, marginBottom: 18 }}>
        <Input.Search
          prefix={<SearchOutlined />}
          id='dashboard-global-search'
          name='globalSearch'
          placeholder='搜索案件、冲突检测或审批'
          value={globalSearchQuery}
          onChange={(event) => setGlobalSearchQuery(event.target.value)}
          onSearch={runGlobalSearch}
          enterButton='搜索'
          style={{ maxWidth: 460 }}
        />
      </section>

      {globalSearchState.status !== 'idle' && (
        <section
          className={`batch-search-feedback ${globalSearchState.status}`}
          style={{ marginBottom: 16 }}
        >
          <SearchOutlined />
          <span>{globalSearchState.message}</span>
          {globalSearchState.results.map((result) => (
            <Button key={result.key} type='link' onClick={() => navigate(result.path)}>
              {result.type}：{result.label}
              {result.detail ? ` · ${result.detail}` : ''}
            </Button>
          ))}
        </section>
      )}

      <div
        className='batch-metric-grid'
        style={{ gridTemplateColumns: 'repeat(4, 1fr)', marginBottom: 22 }}
      >
        {dashboardMetrics.map((item) => {
          const cardTone =
            item.tone === 'red'
              ? 'danger'
              : item.tone === 'orange'
                ? 'warn'
                : item.tone === 'green' || item.tone === 'teal'
                  ? 'ok'
                  : ''
          return (
            <section key={item.label} className={`kpi-card ${cardTone}`.trim()}>
              <div className='kpi-label'>
                <span className='tag-dot' />
                {item.label}
              </div>
              <div className='kpi-val'>{item.value}</div>
              <div className='kpi-foot'>{item.delta}</div>
            </section>
          )
        })}
      </div>

      <div className='ng-grid'>
        <section className='ng-panel'>
          <header className='ng-panel-head'>
            <div className='ng-panel-title'>
              我的待办
              <span className='count'>{unreadInbox}</span>
            </div>
            <div className='ng-panel-actions'>
              <button
                type='button'
                className={`ng-chip ${todoFilter === '全部' ? 'active' : ''}`}
                onClick={() => setTodoFilter('全部')}
              >
                全部
              </button>
              <button
                type='button'
                className={`ng-chip ${todoFilter === '紧急' ? 'active' : ''}`}
                onClick={() => setTodoFilter('紧急')}
              >
                紧急 {urgentTodoCount}
              </button>
              <button
                type='button'
                className={`ng-chip ${todoFilter === '审批' ? 'active' : ''}`}
                onClick={() => setTodoFilter('审批')}
              >
                审批 {approvalTodoCount}
              </button>
              <button
                type='button'
                className={`ng-chip ${todoFilter === '任务' ? 'active' : ''}`}
                onClick={() => setTodoFilter('任务')}
              >
                任务
              </button>
            </div>
          </header>
          <div className='ng-todo-list'>
            {filteredTodoItems.map((item) => {
              const rawType = textValue(item.type || item.source_type, '').toLowerCase()
              const markCls = ['approval', 'approval_request'].includes(rawType)
                ? 'approval'
                : rawType.includes('conflict')
                  ? 'conflict'
                  : rawType.includes('deadline')
                    ? 'deadline'
                    : 'doc'
              const pri = textValue(item.priority, '').toLowerCase()
              const urgCls = ['critical', 'high'].includes(pri)
                ? 'now'
                : pri === 'medium'
                  ? 'today'
                  : 'soon'
              const markText = workItemTypeLabel(item.type || item.source_type).slice(0, 1)
              return (
                <div key={item.id || item.title} className='ng-todo-item'>
                  <div className={`ng-todo-mark ${markCls}`}>{markText}</div>
                  <div className='ng-todo-body'>
                    <div className='ng-todo-title'>{textValue(item.title)}</div>
                    <div className='ng-todo-desc'>{textValue(item.content, '—')}</div>
                    <div className='ng-todo-foot'>
                      <span>{workItemTypeLabel(item.source_type)}</span>
                      <span className={`ng-urgency ${urgCls}`}>
                        截止 {formatApiDate(item.due_at)}
                      </span>
                      <RiskTag text={priorityLabel(item.priority)} />
                    </div>
                  </div>
                </div>
              )
            })}
            {filteredTodoItems.length === 0 && (
              <div className='ng-todo-item'>
                <div className='ng-todo-body'>
                  <div className='ng-todo-desc'>
                    {todoItems.length === 0 ? '暂无数据库待办' : `暂无${todoFilter}待办`}
                  </div>
                </div>
              </div>
            )}
          </div>
          <Button
            type='link'
            aria-label='查看全部待办'
            onClick={() => navigate('/inbox')}
            style={{ margin: '8px 22px' }}
          >
            查看全部待办
          </Button>
        </section>

        <section className='ng-panel'>
          <header className='ng-panel-head'>
            <div className='ng-panel-title'>
              利益冲突待复核
              <span className='count'>{openConflicts}</span>
            </div>
            <div className='ng-panel-actions'>
              <button
                type='button'
                className='ng-chip'
                aria-label='查看全部冲突任务'
                onClick={() => navigate('/conflict')}
              >
                查看全部冲突任务
              </button>
            </div>
          </header>
          <div className='ng-todo-list'>
            {riskItems.map((item) => (
              <div key={item.id || item.title} className='ng-todo-item'>
                <div className='ng-todo-mark conflict'>冲</div>
                <div className='ng-todo-body'>
                  <div className='ng-todo-title'>{textValue(item.title)}</div>
                  <div className='ng-todo-desc'>客户：{textValue(item.client_name, '-')}</div>
                  <div className='ng-todo-foot'>
                    <RiskTag text={riskLabel(item.risk_level)} />
                    <RiskTag text={statusLabel(item.status)} />
                    <span>发起 {formatApiDate(item.created_at)}</span>
                  </div>
                </div>
              </div>
            ))}
            {riskItems.length === 0 && (
              <div className='ng-todo-item'>
                <div className='ng-todo-body'>
                  <div className='ng-todo-desc'>暂无数据库冲突复核任务</div>
                </div>
              </div>
            )}
          </div>
        </section>
      </div>

      <div className='ng-grid-2'>
        <section className='ng-panel'>
          <header className='ng-panel-head'>
            <div className='ng-panel-title'>案件趋势（近6个月）</div>
            <div className='ng-panel-actions'>
              <span className='ng-chip'>月度</span>
            </div>
          </header>
          <div style={{ padding: 20 }}>
            <div className='ng-bars'>
              {caseRowsLive.slice(0, 6).map((row, index) => {
                const h = Math.max(16, (6 - index) * 16)
                return (
                  <div key={row.id || index} className='ng-bar-col'>
                    <div className={`ng-bar ${index === 0 ? 'peak' : ''}`} style={{ height: h }} />
                    <span className='ng-bar-x'>{formatApiDate(row.updated_at).slice(5, 10)}</span>
                  </div>
                )
              })}
              {caseRowsLive.length === 0 && (
                <div className='ng-bar-col'>
                  <div className='ng-bar' style={{ height: 16 }} />
                  <span className='ng-bar-x'>—</span>
                </div>
              )}
            </div>
          </div>
        </section>

        <section className='ng-panel'>
          <header className='ng-panel-head'>
            <div className='ng-panel-title'>
              在办案件阶段分布
              <span className='count'>{activeCases}</span>
            </div>
            <div className='ng-panel-actions'>
              <span className='ng-chip'>阶段</span>
            </div>
          </header>
          <div style={{ padding: '16px 22px' }}>
            {stageCounts.slice(0, 5).map((item, index) => {
              const total = stageCounts.reduce((sum, s) => sum + (s.count ?? 0), 0) || 1
              const pct = Math.round(((item.count ?? 0) / total) * 100)
              return (
                <div key={item.key} className='ng-distro-row'>
                  <span className='ng-distro-name'>{statusLabel(item.key)}</span>
                  <div className='ng-distro-track'>
                    <div className={`ng-distro-fill b${index + 1}`} style={{ width: `${pct}%` }} />
                  </div>
                  <span className='ng-distro-val'>{item.count ?? 0}</span>
                </div>
              )
            })}
            {stageCounts.length === 0 && (
              <div className='ng-distro-row'>
                <span className='ng-distro-name'>暂无数据</span>
                <div className='ng-distro-track' />
                <span className='ng-distro-val'>—</span>
              </div>
            )}
          </div>
        </section>
      </div>

      <div className='ng-grid-2'>
        <section className='ng-panel'>
          <header className='ng-panel-head'>
            <div className='ng-panel-title'>
              审批提醒
              <span className='count'>{pendingApprovals}</span>
            </div>
            <div className='ng-panel-actions'>
              <span className='ng-chip'>队列 {approvalItems.length}</span>
            </div>
          </header>
          <div style={{ padding: 20, display: 'flex', gap: 18, alignItems: 'center' }}>
            <Progress
              type='circle'
              percent={pendingApprovals > 0 ? 100 : 0}
              format={() => String(pendingApprovals)}
              strokeColor='var(--color-primary-500)'
              size={110}
            />
            <div className='ng-todo-body'>
              <div className='ng-todo-foot'>
                <StatusDot color='blue' /> 审批队列 <strong>{approvalItems.length}</strong>
              </div>
              <div className='ng-todo-foot'>
                <StatusDot color='slate' /> 待处理 <strong>{pendingApprovals}</strong>
              </div>
              <div className='ng-todo-foot'>
                <StatusDot color='orange' /> 冲突任务 <strong>{openConflicts}</strong>
              </div>
            </div>
          </div>
        </section>

        <section className='ng-panel'>
          <header className='ng-panel-head'>
            <div className='ng-panel-title'>逾期任务 TOP5</div>
            <div className='ng-panel-actions'>
              <span className='ng-chip'>{overdueItems.length}</span>
            </div>
          </header>
          <div className='ng-todo-list'>
            {overdueItems.map((item, index) => (
              <div key={item.id || item.title} className='ng-todo-item'>
                <div className={`ng-todo-mark ${index < 3 ? 'conflict' : 'deadline'}`}>
                  {index + 1}
                </div>
                <div className='ng-todo-body'>
                  <div className='ng-todo-title'>{textValue(item.title)}</div>
                  <div className='ng-todo-foot'>
                    <RiskTag text={priorityLabel(textValue(item.priority, 'overdue'))} />
                  </div>
                </div>
              </div>
            ))}
            {overdueItems.length === 0 && (
              <div className='ng-todo-item'>
                <div className='ng-todo-body'>
                  <div className='ng-todo-desc'>暂无数据库逾期任务</div>
                </div>
              </div>
            )}
          </div>
        </section>
      </div>

      <div className='ng-grid-2'>
        <section className='ng-panel'>
          <header className='ng-panel-head'>
            <div className='ng-panel-title'>最近活动</div>
            <div className='ng-panel-actions'>
              <span className='ng-chip'>{activities.length}</span>
            </div>
          </header>
          <div className='ng-todo-list'>
            {activities.map((activity, index) => (
              <div key={activity.id || activity.title || index} className='ng-todo-item'>
                <div
                  className={`ng-todo-mark ${activity.type === 'approval' ? 'approval' : 'doc'}`}
                >
                  {activity.type === 'approval' ? '审' : '动'}
                </div>
                <div className='ng-todo-body'>
                  <div className='ng-todo-title'>{textValue(activity.title)}</div>
                  <div className='ng-todo-foot'>
                    <span>{formatApiDate(activity.created_at)}</span>
                  </div>
                </div>
              </div>
            ))}
            {activities.length === 0 && (
              <div className='ng-todo-item'>
                <div className='ng-todo-body'>
                  <div className='ng-todo-desc'>暂无数据库活动记录</div>
                </div>
              </div>
            )}
          </div>
        </section>

        <section className='ng-panel'>
          <header className='ng-panel-head'>
            <div className='ng-panel-title'>数据概览</div>
          </header>
          <div className='ng-todo-list'>
            <div className='ng-todo-item'>
              <div className='ng-todo-mark doc'>案</div>
              <div className='ng-todo-body'>
                <div className='ng-todo-title'>{commandCenter ? activeCases : '-'}</div>
                <div className='ng-todo-desc'>数据库案件总数</div>
              </div>
            </div>
            <div className='ng-todo-item'>
              <div className='ng-todo-mark conflict'>冲</div>
              <div className='ng-todo-body'>
                <div className='ng-todo-title'>{commandCenter ? openConflicts : '-'}</div>
                <div className='ng-todo-desc'>待复核冲突任务</div>
              </div>
            </div>
          </div>
        </section>
      </div>

      <div className='ng-section-title'>快捷入口</div>
      <div className='ng-shortcuts'>
        <div className='ng-shortcut' onClick={() => navigate('/case/create')}>
          <div className='ng-shortcut-ico'>案</div>
          <div className='ng-shortcut-name'>新建立案</div>
          <div className='ng-shortcut-desc'>录入新案件信息</div>
        </div>
        <div className='ng-shortcut' onClick={() => navigate('/inbox')}>
          <div className='ng-shortcut-ico'>办</div>
          <div className='ng-shortcut-name'>待办中心</div>
          <div className='ng-shortcut-desc'>处理 inbox 待办</div>
        </div>
        <div className='ng-shortcut' onClick={() => navigate('/conflict')}>
          <div className='ng-shortcut-ico'>冲</div>
          <div className='ng-shortcut-name'>冲突检测</div>
          <div className='ng-shortcut-desc'>复核利益冲突</div>
        </div>
        <div className='ng-shortcut' onClick={() => navigate('/approval')}>
          <div className='ng-shortcut-ico'>审</div>
          <div className='ng-shortcut-name'>审批队列</div>
          <div className='ng-shortcut-desc'>待审批事项</div>
        </div>
        <div className='ng-shortcut' onClick={() => navigate('/dashboard')}>
          <div className='ng-shortcut-ico'>析</div>
          <div className='ng-shortcut-name'>数据看板</div>
          <div className='ng-shortcut-desc'>经营分析</div>
        </div>
      </div>
    </div>
  )
}

export function ClientMasterProfile() {
  const navigate = useNavigate()
  const { user: activeUser } = useAppStore()
  const canManageClients = hasPermission(activeUser, 'client:manage')
  const [clientRows, setClientRows] = React.useState<any[]>([])
  const [clientSearch, setClientSearch] = React.useState('')
  const [clientTab, setClientTab] = React.useState('基本信息')
  const [selectedClientId, setSelectedClientId] = React.useState<string | number | null>(null)
  const [profile, setProfile] = React.useState<any>(null)
  const [loading, setLoading] = React.useState(false)
  const [createClientOpen, setCreateClientOpen] = React.useState(false)
  const [creatingClient, setCreatingClient] = React.useState(false)
  const [clientDraft, setClientDraft] = React.useState({
    name: '',
    type: '企业',
    identityNumber: '',
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
        setSelectedClientId((current) => current || rows[0]?.id || null)
      })
      .catch((error) => message.error(error instanceof Error ? error.message : '加载客户列表失败'))
  }, [])

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
  const primaryContact = profile?.primary_contact || null
  const completeness = profile?.completeness || {}
  const relatedParties = listOf<any>(profile?.related_parties)
  const matterHistory = listOf<any>(profile?.matter_history)
  const conflictHistory = listOf<any>(profile?.conflict_history)
  const filteredClientRows = clientRows.filter((item) =>
    `${item.name || ''} ${item.email || ''} ${item.company || ''}`
      .toLowerCase()
      .includes(clientSearch.trim().toLowerCase()),
  )
  const openCreateClient = () => {
    setClientDraft({ name: '', type: '企业', identityNumber: '', phone: '', email: '', address: '' })
    setCreateClientOpen(true)
  }
  const createClient = async () => {
    if (!clientDraft.name.trim()) {
      message.warning('请先填写客户名称')
      return
    }
    if (!clientDraft.identityNumber.trim()) {
      message.warning(
        clientDraft.type === '企业' ? '请填写统一社会信用代码' : '请填写身份证件号码',
      )
      return
    }
    setCreatingClient(true)
    try {
      const createdClient = await apiRequest<any>('/clients', {
        method: 'POST',
        body: JSON.stringify({
          ...clientDraft,
          identity_type:
            clientDraft.type === '企业' ? 'SOCIAL_CREDIT_CODE' : 'ID_CARD',
          identity_number: clientDraft.identityNumber.trim(),
          identityNumber: undefined,
        }),
      })
      message.success('客户已创建')
      setCreateClientOpen(false)
      const data = await apiRequest<any>('/clients?page=1&page_size=20')
      const rows = listOf<any>(data?.clients || data?.list || data)
      setClientRows(rows)
      if (createdClient?.id) {
		setClientTab('基本信息')
        setSelectedClientId(createdClient.id)
      }
    } catch (error) {
      message.error(error instanceof Error ? error.message : '新增客户失败')
    } finally {
      setCreatingClient(false)
    }
  }
  const openContactModal = () => {
	const legacyContact = Boolean(primaryContact?.legacy)
    setContactDraft({
      name: textValue(primaryContact?.name || client.contact_person, ''),
      position: textValue(primaryContact?.position, ''),
      // Legacy contact phones are masked by the aggregate API. Never persist a
      // masked display value as if it were the real phone number.
      phone: legacyContact ? '' : textValue(primaryContact?.phone, ''),
      email: textValue(primaryContact?.email, ''),
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
        version: numberValue(primaryContact?.version, 0),
        name: contactDraft.name.trim(),
        position: contactDraft.position.trim(),
        phone: contactDraft.phone.trim(),
        email: contactDraft.email.trim(),
      }
      const updatedContact = await apiRequest<any>(`/clients/${client.id}/primary-contact`, {
        method: 'PUT',
        body: JSON.stringify(payload),
      })
      setProfile((current: any) => ({
        ...current,
        primary_contact: updatedContact,
        client: {
          ...(current?.client || {}),
          contact_person: updatedContact.name,
          contact_phone: updatedContact.phone,
          updated_at: new Date().toISOString(),
        },
      }))
      message.success('主联系人已更新')
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
      />

      <div className='ng-content'>
        <div className='batch-client-layout'>
          <aside className='batch-client-list'>
            <div className='batch-panel-title'>
              <h2>客户列表</h2>
              {canManageClients && (
                <Button icon={<PlusOutlined />} onClick={openCreateClient}>
                  新增客户
                </Button>
              )}
            </div>
            <Input
              id='client-search'
              name='clientSearch'
              prefix={<SearchOutlined />}
              value={clientSearch}
              onChange={(event) => setClientSearch(event.target.value)}
              allowClear
              placeholder='搜索客户名称或邮箱'
            />
            <div className='batch-client-items'>
              {filteredClientRows.map((item) => (
                <button
                  type='button'
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
                </button>
              ))}
              {filteredClientRows.length === 0 && (
                <p>{clientRows.length === 0 ? '暂无数据库客户' : '未找到匹配客户'}</p>
              )}
            </div>
            <div className='batch-pagination'>共 {clientRows.length} 条</div>
          </aside>

          <main className='batch-profile-main'>
            <section className='batch-profile-hero'>
              <span className='batch-company-avatar'>
                <BankOutlined />
              </span>
              <div className='batch-profile-title'>
                <h1>
                  {textValue(client.name, loading ? '加载中' : '未选择客户')}{' '}
                  <Tag color='green'>{statusLabel(client.status)}</Tag>
                </h1>
                <p>
                  客户类型：{textValue(client.type)} · 行业：{textValue(client.industry)}
                </p>
              </div>
              <div className='batch-profile-metrics'>
                <div>
                  <Progress
                    type='circle'
                    percent={completeness.score || 0}
                    size={64}
                    strokeColor='#12a89d'
                  />
                  <span>数据完整度</span>
                </div>
                <div>
                  <strong className={conflictHistory.length ? 'orange-text' : 'green-text'}>
                    {conflictHistory.length ? '有检测记录' : '未见记录'}
                  </strong>
                  <span>冲突记录</span>
                </div>
                <div>
                  <strong className='green-text'>{statusLabel(client.status)}</strong>
                  <span>客户状态</span>
                </div>
                <div>
                  <strong>{formatApiDate(client.created_at).slice(0, 10)}</strong>
                  <span>首次入库</span>
                </div>
                <div>
                  <strong>{formatApiDate(client.updated_at).slice(0, 10)}</strong>
                  <span>最近更新</span>
                </div>
              </div>
            </section>

            <div className='batch-tabs'>
              {[
                '基本信息',
                '关联方与穿透',
                '历史委托与案件',
                '冲突信息池',
                '联系人',
                '附件文档',
                '活动日志',
              ].map((tab) => (
                <button
                  key={tab}
                  className={clientTab === tab ? 'active' : ''}
                  aria-label={`${client.name || '客户'} ${tab}`}
                  onClick={() => setClientTab(tab)}
                >
                  {tab}
                </button>
              ))}
            </div>

            <div className='batch-profile-grid'>
              {clientTab === '基本信息' && (
                <SectionCard title='基本信息' className='span-2'>
                  <div className='batch-info-grid'>
                    {[
                      ['客户名称', textValue(client.name)],
                      ['客户类型', textValue(client.type)],
                      [
                        textValue(client.identity_type).toUpperCase() === 'SOCIAL_CREDIT_CODE'
                          ? '统一社会信用代码'
                          : '身份证件标识',
                        textValue(client.identity_status, '未登记'),
                      ],
                      ['客户公共邮箱', textValue(client.email)],
                      ['联系电话', textValue(client.phone)],
                      ['所属行业', textValue(client.industry)],
                      ['主联系人', textValue(primaryContact?.name || client.contact_person)],
                      ['联系地址', textValue(client.address)],
                      ['客户来源', clientSourceLabel(client.source)],
                      ['备注', textValue(client.notes)],
                    ].map(([label, value]) => (
                      <p key={label}>
                        <span>{label}</span>
                        <strong>{value}</strong>
                      </p>
                    ))}
                  </div>
                </SectionCard>
              )}

              {clientTab === '关联方与穿透' && (
                <SectionCard title='关系图谱（穿透预览）'>
                  <div className='batch-relation-graph'>
                    <span className='node main'>{textValue(client.name, '客户')}</span>
                    {relatedParties.slice(0, 5).map((party: any, index: number) => (
                      <span
                        key={party.name || index}
                        className={`node small ${['top-left', 'top-right', 'bottom-left', 'bottom-mid', 'bottom-right'][index]}`}
                      >
                        {textValue(party.name)}
                        <br />
                        {relationshipTypeLabel(party.relationship_type)}
                      </span>
                    ))}
                  </div>
                </SectionCard>
              )}

              <SectionCard title='快速操作'>
                <div className='batch-action-list'>
                  <Button
                    icon={<PlusOutlined />}
                    onClick={() => navigate(`/case/create?client_id=${client.id}`)}
                    disabled={!client.id}
                  >
                    发起新案件
                  </Button>
                  <Button
                    icon={<FileSearchOutlined />}
                    onClick={() => navigate(`/case/create?client_id=${client.id}&intent=conflict`)}
                    disabled={!client.id}
                  >
                    以此客户新建立案并检查冲突
                  </Button>
                  <Button icon={<UserOutlined />} onClick={openContactModal}>
                    编辑主联系人
                  </Button>
                  <Button icon={<CloudUploadOutlined />} onClick={() => setUploadModalOpen(true)}>
                    上传附件
                  </Button>
                  <Button
                    icon={<DownloadOutlined />}
                    onClick={() => {
                      try {
                        const blob = new Blob([JSON.stringify(client, null, 2)], {
                          type: 'application/json',
                        })
                        const url = URL.createObjectURL(blob)
                        const a = document.createElement('a')
                        a.href = url
                        a.download = `client-${client.name || 'export'}.json`
                        a.click()
                        URL.revokeObjectURL(url)
                        message.success('客户档案已导出')
                      } catch {
                        message.error('导出失败，请稍后重试')
                      }
                    }}
                  >
                    导出客户档案
                  </Button>
                </div>
              </SectionCard>

              <Modal
                title='编辑主联系人'
                open={contactModalOpen}
                onCancel={() => setContactModalOpen(false)}
                onOk={savePrimaryContact}
                okText='保存'
                confirmLoading={contactSaving}
              >
                <div className='batch-form-grid two'>
                  <div className='batch-field'>
                    <label htmlFor='contact-name'>姓名 *</label>
                    <Input
                      id='contact-name'
                      name='contactName'
                      placeholder='联系人姓名'
                      value={contactDraft.name}
                      onChange={(event) =>
                        setContactDraft((draft) => ({ ...draft, name: event.target.value }))
                      }
                    />
                  </div>
                  <div className='batch-field'>
                    <label htmlFor='contact-position'>职位</label>
                    <Input
                      id='contact-position'
                      name='contactPosition'
                      placeholder='职位'
                      value={contactDraft.position}
                      onChange={(event) =>
                        setContactDraft((draft) => ({ ...draft, position: event.target.value }))
                      }
                    />
                  </div>
                  <div className='batch-field'>
                    <label htmlFor='contact-phone'>电话</label>
                    <Input
                      id='contact-phone'
                      name='contactPhone'
                      placeholder='联系电话'
                      value={contactDraft.phone}
                      onChange={(event) =>
                        setContactDraft((draft) => ({ ...draft, phone: event.target.value }))
                      }
                    />
                  </div>
                  <div className='batch-field'>
                    <label htmlFor='contact-email'>邮箱</label>
                    <Input
                      id='contact-email'
                      name='contactEmail'
                      placeholder='电子邮箱'
                      value={contactDraft.email}
                      onChange={(event) =>
                        setContactDraft((draft) => ({ ...draft, email: event.target.value }))
                      }
                    />
                  </div>
                </div>
              </Modal>
              <Modal
                title='上传附件'
                open={uploadModalOpen}
                onCancel={() => setUploadModalOpen(false)}
                onOk={uploadClientAttachment}
                okText='上传'
                confirmLoading={uploadSaving}
              >
                <Upload
                  beforeUpload={(file) => {
                    setUploadFile(file)
                    return false
                  }}
                  maxCount={1}
                  fileList={
                    uploadFile
                      ? [{ uid: uploadFile.name, name: uploadFile.name, status: 'done' }]
                      : []
                  }
                  onRemove={() => {
                    setUploadFile(null)
                    return true
                  }}
                >
                  <Button icon={<CloudUploadOutlined />}>选择附件</Button>
                </Upload>
                <p>附件将关联到当前客户档案：{textValue(client.name, '未选择客户')}</p>
              </Modal>

              {clientTab === '基本信息' && (
                <SectionCard title='别名 / 曾用名'>
                  <div className='batch-key-list'>
                    {listOf<string>(client.aliases).map((name, index) => (
                      <p key={`${name}-${index}`}>
                        {name}
                        <MoreOutlined />
                      </p>
                    ))}
                    {!client.aliases?.length && <p>数据库暂无别名记录</p>}
                  </div>
                </SectionCard>
              )}

              {clientTab === '关联方与穿透' && (
                <SectionCard title='关联方'>
                  <DataTable>
                    <table>
                      <tbody>
                        {relatedParties.map((party: any, index) => (
                          <tr key={`${textValue(party.id || party.name, 'related')}-${index}`}>
                            <td>{textValue(party.name)}</td>
                            <td>{relationshipTypeLabel(party.relationship_type)}</td>
                          </tr>
                        ))}
                        {relatedParties.length === 0 && (
                          <tr>
                            <td colSpan={2}>暂无数据库关联方</td>
                          </tr>
                        )}
                      </tbody>
                    </table>
                  </DataTable>
                </SectionCard>
              )}

              {clientTab === '历史委托与案件' && (
                <SectionCard title='关联案件（近12个月）'>
                  <DataTable>
                    <table>
                      <tbody>
                        {matterHistory.map((row: any) => (
                          <tr key={row.id}>
                            <td>{textValue(row.title)}</td>
                            <td>{dbCaseType(textValue(row.case_type))}</td>
                            <td>
                              <RiskTag text={statusLabel(row.status)} />
                            </td>
                          </tr>
                        ))}
                        {matterHistory.length === 0 && (
                          <tr>
                            <td colSpan={3}>暂无数据库关联案件</td>
                          </tr>
                        )}
                      </tbody>
                    </table>
                  </DataTable>
                </SectionCard>
              )}

              {(clientTab === '冲突信息池' || clientTab === '活动日志') && (
                <SectionCard title={clientTab === '冲突信息池' ? '冲突信息池' : '最近活动'}>
                  <div className='batch-activity-list profile'>
                    {conflictHistory.map((activity: any) => (
                      <p key={activity.check_id}>
                        <StatusDot color='blue' />
                        <span>{textValue(activity.case_name)}</span>
                        <em>{formatApiDate(activity.created_at)}</em>
                      </p>
                    ))}
                    {conflictHistory.length === 0 && <p>暂无数据库冲突活动</p>}
                  </div>
                </SectionCard>
              )}

              {clientTab === '联系人' && (
                <SectionCard title='联系人' className='span-2'>
                  <div className='batch-info-grid'>
                    <p>
                      <span>姓名</span>
                      <strong>{textValue(primaryContact?.name || client.contact_person)}</strong>
                    </p>
                    <p>
                      <span>电话</span>
                      <strong>{textValue(primaryContact?.phone || client.contact_phone)}</strong>
                    </p>
                    <p>
                      <span>邮箱</span>
                      <strong>{textValue(primaryContact?.email)}</strong>
                    </p>
                  </div>
                </SectionCard>
              )}

              {clientTab === '附件文档' && (
                <SectionCard title='附件文档' className='span-2'>
                  <p>当前档案附件通过文档服务保存并关联客户。</p>
                </SectionCard>
              )}
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
                options={[
                  { value: '企业', label: '企业' },
                  { value: '个人', label: '个人' },
                ]}
                onChange={(value) => setClientDraft((current) => ({ ...current, type: value }))}
              />
            </div>
            <div className='batch-client-create-field'>
              <span>客户名称 *</span>
              <Input
                value={clientDraft.name}
                onChange={(event) =>
                  setClientDraft((current) => ({ ...current, name: event.target.value }))
                }
                placeholder='请输入客户名称'
              />
            </div>
            <div className='batch-client-create-field'>
              <span>{clientDraft.type === '企业' ? '统一社会信用代码 *' : '身份证件号码 *'}</span>
              <Input.Password
                value={clientDraft.identityNumber}
                onChange={(event) =>
                  setClientDraft((current) => ({ ...current, identityNumber: event.target.value }))
                }
                placeholder={clientDraft.type === '企业' ? '按营业执照填写' : '按身份证件填写'}
                autoComplete='off'
              />
            </div>
            <div className='batch-client-create-field'>
              <span>联系电话</span>
              <Input
                value={clientDraft.phone}
                onChange={(event) =>
                  setClientDraft((current) => ({ ...current, phone: event.target.value }))
                }
                placeholder='请输入联系电话'
              />
            </div>
            <div className='batch-client-create-field'>
              <span>电子邮箱</span>
              <Input
                value={clientDraft.email}
                onChange={(event) =>
                  setClientDraft((current) => ({ ...current, email: event.target.value }))
                }
                placeholder='请输入电子邮箱'
              />
            </div>
            <div className='batch-client-create-field'>
              <span>联系地址</span>
              <Input
                value={clientDraft.address}
                onChange={(event) =>
                  setClientDraft((current) => ({ ...current, address: event.target.value }))
                }
                placeholder='请输入联系地址'
              />
            </div>
          </div>
        </Modal>
      </div>
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
  const visibleConflictRows = listOf<CommandCenterRiskItem>(commandCenter?.risk_queue)
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
    const matchesConflictQueue =
      statusFilter === '冲突复核' &&
      visibleConflictRows.some((item) =>
        conflictMatchesCaseContext(
          item,
          textValue(row.id, ''),
          textValue(row.case_number, ''),
          textValue(row.title, ''),
        ),
      )
    const matchesStatus =
      allowedStatuses.length === 0 ||
      allowedStatuses.includes(normalizedStatus) ||
      matchesConflictQueue
    const searchable = [
      row.case_number,
      row.title,
      row.client_name,
      row.case_type,
      row.status,
      row.priority,
      row.lawyer_name,
    ]
      .map((value) => textValue(value, '').toLowerCase())
      .join(' ')
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
    <div className='batch-page case-management-page ng-content'>
      <header className='ng-page-head'>
        <div>
          <div className='eyebrow'>案件管理 / 案件清单</div>
          <h1>案件管理</h1>
          <p className='sub'>统一管理潜在接案、冲突复核、接案审批、在办案件和结案归档。</p>
        </div>
        <div className='ng-head-actions'>
          <Input
            prefix={<SearchOutlined />}
            value={searchTerm}
            onChange={(event) => setSearchTerm(event.target.value)}
            allowClear
            placeholder='搜索案件编号、客户、对方、负责人...'
          />
          <Button icon={<DownloadOutlined />} onClick={exportCases}>
            导出
          </Button>
          <Button
            type='primary'
            icon={<PlusOutlined />}
            loading={loading}
            onClick={() => navigate('/case/create')}
          >
            新建案件
          </Button>
        </div>
      </header>

      <div
        className='batch-metric-grid case-metrics'
        style={{ gridTemplateColumns: 'repeat(5, 1fr)' }}
      >
        {[
          {
            icon: <FolderOpenOutlined />,
            label: '在办案件',
            value: summary.active_cases ?? 0,
            delta: '',
            tone: 'blue' as Tone,
          },
          {
            icon: <FileSearchOutlined />,
            label: '冲突复核中',
            value: workflow.conflict ?? 0,
            delta: '',
            tone: 'red' as Tone,
          },
          {
            icon: <AuditOutlined />,
            label: '接案审批中',
            value: workflow.approval ?? 0,
            delta: '',
            tone: 'orange' as Tone,
          },
          {
            icon: <ClockCircleOutlined />,
            label: '待补充材料',
            value: workflow.intake ?? 0,
            delta: '',
            tone: 'red' as Tone,
          },
          {
            icon: <CheckCircleOutlined />,
            label: '客户总数',
            value: summary.clients ?? 0,
            delta: '',
            tone: 'teal' as Tone,
          },
        ].map((item) => {
          const toneClass =
            item.tone === 'red'
              ? 'danger'
              : item.tone === 'orange'
                ? 'warn'
                : item.tone === 'green' || item.tone === 'teal'
                  ? 'ok'
                  : ''
          return (
            <section key={item.label} className={`kpi-card ${toneClass}`}>
              <div className='kpi-label'>
                <span className='tag-dot' />
                {item.icon}
                {item.label}
              </div>
              <div className='kpi-val'>{item.value}</div>
              <div className='kpi-foot'>{item.delta}</div>
            </section>
          )
        })}
      </div>

      <div className='batch-case-layout'>
        <section className='ng-panel span-2'>
          <div className='ng-filter-bar'>
            <div className='ng-panel-title'>案件清单</div>
            {Object.keys(statusFilterMap).map((tab) => (
              <button
                key={tab}
                type='button'
                className={`ng-filter-tab ${statusFilter === tab ? 'active' : ''}`}
                onClick={() => setStatusFilter(tab)}
              >
                {tab}
              </button>
            ))}
          </div>
          <div className='ng-table-wrap'>
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
                  <th>优先级</th>
                  <th>负责人</th>
                  <th>更新时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {filteredCaseRows.map((row) => (
                  <tr
                    key={row.id || row.case_number}
                    className={
                      ['high', 'urgent', 'critical'].includes((row.priority || '').toLowerCase())
                        ? 'danger-row'
                        : ''
                    }
                  >
                    <td className='mono-cell'>{textValue(row.case_number || row.id)}</td>
                    <td className='strong-cell'>{textValue(row.title)}</td>
                    <td>{textValue(row.client_name)}</td>
                    <td>
                      <RiskTag text={dbCaseType(textValue(row.case_type, ''))} />
                    </td>
                    <td>
                      <RiskTag text={statusLabel(row.status)} />
                    </td>
                    <td>
                      <RiskTag text={priorityLabel(row.priority)} />
                    </td>
                    <td>{textValue(row.lawyer_name)}</td>
                    <td>{formatApiDate(row.updated_at)}</td>
                    <td>
                      <Button
                        size='small'
                        type='primary'
                        ghost
                        onClick={() => navigate(`/case/${row.id}`)}
                      >
                        查看
                      </Button>
                    </td>
                  </tr>
                ))}
                {filteredCaseRows.length === 0 && (
                  <tr>
                    <td colSpan={9}>
                      {liveCaseRows.length === 0
                        ? '暂无数据库案件'
                        : '当前搜索或筛选条件下暂无案件'}
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </section>

        <section className='ng-panel'>
          <div className='ng-panel-head'>
            <div className='ng-panel-title'>接案漏斗</div>
          </div>
          <div className='batch-case-funnel'>
            {[
              ['接案准备', workflow.intake ?? 0, 'blue' as Tone],
              ['冲突复核', workflow.conflict ?? 0, 'red' as Tone],
              ['接案审批', workflow.approval ?? 0, 'orange' as Tone],
              ['办理中', summary.active_cases ?? 0, 'teal' as Tone],
              ['客户总数', summary.clients ?? 0, 'green' as Tone],
            ].map((item) => (
              <p key={item[0]}>
                <StatusDot color={item[2] as Tone} />
                <span>{item[0]}</span>
                <strong>{item[1]}</strong>
              </p>
            ))}
          </div>
        </section>

        <section className='ng-panel'>
          <div className='ng-panel-head'>
            <div className='ng-panel-title'>高优先级案件</div>
          </div>
          <div className='batch-overdue-list'>
            {filteredCaseRows
              .filter((row) => (row.priority || '').toLowerCase() === 'high')
              .map((row) => (
                <p key={row.id || row.title}>
                  <StatusDot color='red' />
                  {textValue(row.title)}
                  <RiskTag text={statusLabel(row.status)} />
                </p>
              ))}
            {filteredCaseRows.filter((row) => (row.priority || '').toLowerCase() === 'high')
              .length === 0 && <p>暂无高优先级案件</p>}
          </div>
        </section>

        <section className='ng-panel span-2'>
          <div className='ng-panel-head'>
            <div className='ng-panel-title'>团队负荷</div>
          </div>
          <div className='batch-policy-grid'>
            {filteredCaseRows.slice(0, 4).map((item) => (
              <article key={`${item.lawyer_name}-${item.id}`}>
                <UserOutlined />
                <div>
                  <strong>{textValue(item.lawyer_name)}</strong>
                  <p>{textValue(item.title)}</p>
                </div>
                <RiskTag text={priorityLabel(item.priority)} />
              </article>
            ))}
            {filteredCaseRows.length === 0 && <p>暂无数据库团队负荷数据</p>}
          </div>
        </section>
      </div>
    </div>
  )
}

export function CaseDetailCenter() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const { user: activeUser } = useAppStore()
  const [caseDetail, setCaseDetail] = React.useState<any | null>(null)
  const [commandCenter, setCommandCenter] = React.useState<CommandCenterPayload | null>(null)
  const [loading, setLoading] = React.useState(false)
  const [subjectRevisionOpen, setSubjectRevisionOpen] = React.useState(false)
  const [subjectRevisionLoading, setSubjectRevisionLoading] = React.useState(false)
  const [subjectParties, setSubjectParties] = React.useState<SubjectPartyOption[]>([])
  const [subjectEntityOptions, setSubjectEntityOptions] = React.useState<SubjectPartyOption[]>([])
  const [subjectEntityQuery, setSubjectEntityQuery] = React.useState('')
  const [subjectEntityLoading, setSubjectEntityLoading] = React.useState(false)
  const [subjectChangeType, setSubjectChangeType] = React.useState<
    'ADD_OPPOSING_PARTY' | 'ADD_THIRD_PARTY' | 'REMOVE_PARTY'
  >('ADD_OPPOSING_PARTY')
  const [selectedSubjectEntityID, setSelectedSubjectEntityID] = React.useState<number | null>(null)
  const [subjectChangeReason, setSubjectChangeReason] = React.useState('')
  const [subjectEntityMode, setSubjectEntityMode] = React.useState<'existing' | 'new'>('existing')
  const [newSubjectName, setNewSubjectName] = React.useState('')
  const [newSubjectAlias, setNewSubjectAlias] = React.useState('')
  const [newSubjectEntityType, setNewSubjectEntityType] = React.useState('LEGAL_PERSON')
  const [newSubjectIdentityType, setNewSubjectIdentityType] = React.useState('SOCIAL_CREDIT_CODE')
  const [newSubjectIdentityNumber, setNewSubjectIdentityNumber] = React.useState('')
  const [pendingSubjectRevision, setPendingSubjectRevision] = React.useState<Record<string, any>>({})

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

  React.useEffect(() => {
    if (!subjectRevisionOpen || !id) return
    let mounted = true
    apiRequest<SubjectPartyOption[]>(`/cases/${id}/subject-parties`)
      .then((rows) => {
        if (mounted) setSubjectParties(Array.isArray(rows) ? rows : [])
      })
      .catch((error: unknown) => {
        if (mounted) message.error(error instanceof Error ? error.message : '读取案件当事人失败')
      })
    return () => {
      mounted = false
    }
  }, [id, subjectRevisionOpen])

  React.useEffect(() => {
    if (
      !subjectRevisionOpen ||
      !id ||
      subjectChangeType === 'REMOVE_PARTY' ||
      subjectEntityMode === 'new'
    ) {
      setSubjectEntityOptions([])
      return
    }
    const query = subjectEntityQuery.trim()
    if (query.length < 2) {
      setSubjectEntityOptions([])
      return
    }
    let mounted = true
    const timer = window.setTimeout(() => {
      setSubjectEntityLoading(true)
      apiRequest<SubjectPartyOption[]>(
        `/cases/${id}/subject-entities?query=${encodeURIComponent(query)}`,
      )
        .then((rows) => {
          if (mounted) setSubjectEntityOptions(Array.isArray(rows) ? rows : [])
        })
        .catch((error: unknown) => {
          if (mounted) message.error(error instanceof Error ? error.message : '搜索结构化主体失败')
        })
        .finally(() => mounted && setSubjectEntityLoading(false))
    }, 250)
    return () => {
      mounted = false
      window.clearTimeout(timer)
    }
  }, [id, subjectChangeType, subjectEntityMode, subjectEntityQuery, subjectRevisionOpen])

  React.useEffect(() => {
    const revisionID = textValue(caseDetail?.pending_subject_revision_id, '')
    if (!id || !revisionID) {
      setPendingSubjectRevision({})
      return
    }
    let mounted = true
    apiRequest<any>(`/cases/${id}/subject-revisions/${revisionID}`)
      .then((result) => mounted && setPendingSubjectRevision(recordValue(result)))
      .catch(() => mounted && setPendingSubjectRevision({}))
    return () => {
      mounted = false
    }
  }, [caseDetail?.pending_subject_revision_id, id])

  const client = settingObject(caseDetail?.client)
  const lawyer = settingObject(caseDetail?.lawyer)
  const currentClientID = textValue(caseDetail?.client_id || client.id, '')
  const currentClientName = textValue(client.name || caseDetail?.client_name, '')
  const relatedRows = listOf(commandCenter?.case_rows)
    .filter((row) => String(row.id) !== String(caseDetail?.id))
    .filter((row) => {
      const rowClientID = textValue(row.client_id, '')
      const rowClientName = textValue(row.client_name, '')
      return (
        (currentClientID && rowClientID === currentClientID) ||
        (currentClientName && rowClientName === currentClientName)
      )
    })
    .slice(0, 5)

  if (loading) {
    return (
      <div className='batch-page case-management-page ng-content'>
        <header className='ng-page-head'>
          <div>
            <div className='eyebrow'>案件管理 / 案件详情</div>
            <h1>正在加载案件详情</h1>
            <p className='sub'>正在读取案件、客户、律师与流程数据。</p>
          </div>
        </header>
      </div>
    )
  }

  if (!caseDetail) {
    return (
      <div className='batch-page case-management-page ng-content'>
        <span className='ng-back' onClick={() => navigate('/case')}>
          <ArrowLeftOutlined /> 返回案件清单
        </span>
        <header className='ng-page-head'>
          <div>
            <div className='eyebrow'>案件管理 / 案件详情</div>
            <h1>案件不存在</h1>
            <p className='sub'>指定案件不存在或当前账号无权访问。</p>
          </div>
        </header>
      </div>
    )
  }

  const metricsForCase = [
    {
      icon: <FileTextOutlined />,
      label: '案件编号',
      value: textValue(caseDetail.case_number || caseDetail.id),
      delta: '',
      tone: 'blue' as Tone,
    },
    {
      icon: <TeamOutlined />,
      label: '客户',
      value: textValue(client.name || caseDetail.client_name),
      delta: '客户资料',
      tone: 'teal' as Tone,
    },
    {
      icon: <UserOutlined />,
      label: '负责律师',
      value: textValue(lawyer.name || caseDetail.lawyer_name),
      delta: '团队分配',
      tone: 'green' as Tone,
    },
    {
      icon: <FileSearchOutlined />,
      label: '优先级',
      value: priorityLabel(caseDetail.priority),
      delta: '风险跟踪',
      tone: (String(caseDetail.priority).toLowerCase() === 'high' ? 'red' : 'orange') as Tone,
    },
  ]
  const currentCaseStatus = textValue(caseDetail.status, '').toLowerCase()
  const caseConflictRecord = listOf<CommandCenterRiskItem>(commandCenter?.risk_queue).find(
    (item) => {
      const parameters = recordValue(item.search_parameters)
      const itemCaseID = firstPresent(
        item.case_id,
        parameters.subjectCaseId,
        parameters.subject_case_id,
      )
      const itemCaseNumber = firstPresent(
        item.case_number,
        item.case_no,
        parameters.subjectCaseNumber,
        parameters.subject_case_number,
      )
      return (
        (itemCaseID !== undefined && String(itemCaseID) === String(caseDetail.id)) ||
        (itemCaseNumber !== undefined && String(itemCaseNumber) === String(caseDetail.case_number))
      )
    },
  )
  const caseConflictDecision = caseConflictRecord
    ? deriveConflictDecisionViewModel(caseConflictRecord)
    : null
  const caseConflictReviewPending = Boolean(
    caseConflictDecision &&
      ['REVIEW_REQUIRED', 'BLOCKED', 'WAIVER_PENDING'].includes(caseConflictDecision.decision),
  )
  const caseStageDefinitions = [
    { label: '接案准备', statuses: ['draft', 'pending', 'todo', '待处理'] },
    { label: '冲突复核', statuses: ['risk_review', 'conflict_ready', 'conflict_checking'] },
    { label: '接案审批', statuses: ['submitted', 'under_review', 'approval_pending'] },
    { label: '办理中', statuses: ['active', 'in_progress', 'open'] },
    { label: '结案归档', statuses: ['completed', 'archived', 'closed'] },
  ]
  const needsConflictReview =
    caseConflictReviewPending ||
    [
      'draft',
      'pending',
      'todo',
      '待处理',
      'risk_review',
      'conflict_ready',
      'conflict_checking',
    ].includes(currentCaseStatus)
  const openConflictReview = () => {
    const params = new URLSearchParams()
    params.set('case_id', textValue(caseDetail.id, ''))
    params.set('case_number', textValue(caseDetail.case_number, ''))
    params.set('case_title', textValue(caseDetail.title, ''))
    navigate(`/conflict?${params.toString()}`)
  }

  const supplementCaseIntake = () => {
    const params = new URLSearchParams()
    params.set('mode', 'supplement')
    params.set('case_id', textValue(caseDetail.id, ''))
    navigate(`/case/create?${params.toString()}`)
  }

  const isAssistant =
    activeUser?.roles.some((role) =>
      ['assistant', 'intake_assistant'].includes(role.toLowerCase()),
    ) ||
    ['assistant', 'intake_assistant'].includes(textValue(getUserInfo()?.role, '').toLowerCase())
  const currentSubjectState = textValue(caseDetail?.subject_state, 'EFFECTIVE').toUpperCase()
  const subjectRevisionPending =
    currentSubjectState !== 'EFFECTIVE' ||
    Boolean(textValue(caseDetail?.pending_subject_revision_id, ''))
  const subjectRevisionID = textValue(caseDetail?.pending_subject_revision_id, '')
  const currentSubjectVersion = Math.max(1, numberValue(caseDetail?.subject_version, 1))
  const canReportSubjectRevision = !isAssistant && !needsConflictReview

  const resetSubjectRevisionForm = () => {
    setSubjectChangeType('ADD_OPPOSING_PARTY')
    setSelectedSubjectEntityID(null)
    setSubjectEntityQuery('')
    setSubjectEntityOptions([])
    setSubjectChangeReason('')
    setSubjectEntityMode('existing')
    setNewSubjectName('')
    setNewSubjectAlias('')
    setNewSubjectEntityType('LEGAL_PERSON')
    setNewSubjectIdentityType('SOCIAL_CREDIT_CODE')
    setNewSubjectIdentityNumber('')
  }

  const openSubjectRevision = () => {
    resetSubjectRevisionForm()
    setSubjectRevisionOpen(true)
  }

  const runSubjectRecheck = async (revisionID: string) => {
    if (!id || !revisionID) throw new Error('主体变更记录不存在')
    const result = await apiRequest<any>(`/cases/${id}/subject-revisions/${revisionID}/recheck`, {
      method: 'POST',
      body: JSON.stringify({}),
    })
    const latest = await apiRequest<any>(`/cases/${id}`)
    setCaseDetail(latest)
    return result
  }

  const submitSubjectRevision = async () => {
    if (!id) return
    const reason = subjectChangeReason.trim()
    if (reason.length < 5) {
      message.error('请说明主体发生变化的原因（至少 5 个字）')
      return
    }
    if (subjectChangeType !== 'REMOVE_PARTY' && subjectEntityMode === 'new') {
      if (newSubjectName.trim().length < 2 || newSubjectIdentityNumber.trim().length < 4) {
        message.error('请填写主体名称和可核验身份标识')
        return
      }
      setSubjectRevisionLoading(true)
      try {
        await apiRequest<any>(`/cases/${id}/subject-entity-registrations`, {
          method: 'POST',
          body: JSON.stringify({
            expected_subject_version: currentSubjectVersion,
            change_type: subjectChangeType,
            name: newSubjectName.trim(),
            alias: newSubjectAlias.trim(),
            entity_type: newSubjectEntityType,
            identity_type: newSubjectIdentityType,
            identity_number: newSubjectIdentityNumber.trim(),
            reason,
          }),
        })
        const latest = await apiRequest<any>(`/cases/${id}`)
        setCaseDetail(latest)
        setSubjectRevisionOpen(false)
        resetSubjectRevisionForm()
        message.success('新主体已提报，案件受控动作已暂停；等待冲突核查岗确认主体档案后再运行重检')
      } catch (error: unknown) {
        message.error(error instanceof Error ? error.message : '新主体登记申请失败')
      } finally {
        setSubjectRevisionLoading(false)
      }
      return
    }
    if (!selectedSubjectEntityID) {
      message.error(
        subjectChangeType === 'REMOVE_PARTY'
          ? '请选择要移除的已登记主体'
          : '请先搜索并选择已登记主体',
      )
      return
    }
    const payload =
      subjectChangeType === 'REMOVE_PARTY'
        ? { remove_party_ids: [selectedSubjectEntityID] }
        : {
            add_parties: [
              {
                entity_id: selectedSubjectEntityID,
                role: subjectChangeType === 'ADD_THIRD_PARTY' ? 'THIRD_PARTY' : 'DEFENDANT',
                party_type: subjectChangeType === 'ADD_THIRD_PARTY' ? 'THIRD_PARTY' : 'OPPOSING',
              },
            ],
          }
    setSubjectRevisionLoading(true)
    try {
      const created = await apiRequest<any>(`/cases/${id}/subject-revisions`, {
        method: 'POST',
        body: JSON.stringify({
          expected_subject_version: currentSubjectVersion,
          change_type: subjectChangeType,
          payload,
          reason,
        }),
      })
      const revisionID = textValue(created?.revision?.id, '')
      if (!revisionID) throw new Error('主体变更已登记，但未返回重检编号')
      await runSubjectRecheck(revisionID)
      setSubjectRevisionOpen(false)
      resetSubjectRevisionForm()
      message.success('主体变更已登记并完成重检，等待独立冲突复核；在此之前受控动作保持暂停')
    } catch (error: unknown) {
      message.error(error instanceof Error ? error.message : '主体变更或重检失败')
    } finally {
      setSubjectRevisionLoading(false)
    }
  }

  const continueSubjectRecheck = async () => {
    if (!subjectRevisionID) return
    setSubjectRevisionLoading(true)
    try {
      await runSubjectRecheck(subjectRevisionID)
      message.success('主体重检已完成，等待独立冲突复核')
    } catch (error: unknown) {
      message.error(error instanceof Error ? error.message : '继续主体重检失败')
    } finally {
      setSubjectRevisionLoading(false)
    }
  }

  return (
    <div className='batch-page case-management-page ng-content'>
      <span className='ng-back' onClick={() => navigate('/case')}>
        <ArrowLeftOutlined /> 返回案件清单
      </span>
      <header className='ng-detail-head'>
        <div>
          <div className='dh-tags'>
            <span>案件管理 / 案件清单 / 案件详情</span>
          </div>
          <h1>{textValue(caseDetail.title, '未命名案件')}</h1>
          <p className='sub'>
            {`案件类型：${dbCaseType(textValue(caseDetail.case_type))} · 当前状态：${statusLabel(caseDetail.status)}`}
          </p>
        </div>
        <div className='dh-actions'>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/case')}>
            返回案件清单
          </Button>
          <Button type='primary' onClick={() => navigate('/case/create')}>
            新建案件
          </Button>
        </div>
      </header>

      <div
        className='batch-metric-grid case-metrics'
        style={{ gridTemplateColumns: 'repeat(4, 1fr)' }}
      >
        {metricsForCase.map((item) => {
          const toneClass =
            item.tone === 'red'
              ? 'danger'
              : item.tone === 'orange'
                ? 'warn'
                : item.tone === 'green' || item.tone === 'teal'
                  ? 'ok'
                  : ''
          return (
            <section key={item.label} className={`kpi-card ${toneClass}`}>
              <div className='kpi-label'>
                <span className='tag-dot' />
                {item.icon}
                {item.label}
              </div>
              <div className='kpi-val'>{item.value}</div>
              <div className='kpi-foot'>{item.delta}</div>
            </section>
          )
        })}
      </div>

      <div className='batch-case-layout'>
        <section className='ng-panel span-2'>
          <div className='ng-panel-head'>
            <div className='ng-panel-title'>案件概览</div>
          </div>
          <div className='ng-table-wrap'>
            <table>
              <tbody>
                <tr>
                  <td>案件名称</td>
                  <td>{textValue(caseDetail.title)}</td>
                </tr>
                <tr>
                  <td>案件类型</td>
                  <td>{dbCaseType(textValue(caseDetail.case_type))}</td>
                </tr>
                <tr>
                  <td>案件状态</td>
                  <td>
                    <RiskTag text={statusLabel(caseDetail.status)} />
                  </td>
                </tr>
                <tr>
                  <td>优先级</td>
                  <td>
                    <RiskTag text={priorityLabel(caseDetail.priority)} />
                  </td>
                </tr>
                <tr>
                  <td>创建时间</td>
                  <td>{formatApiDate(caseDetail.created_at)}</td>
                </tr>
                <tr>
                  <td>更新时间</td>
                  <td>{formatApiDate(caseDetail.updated_at)}</td>
                </tr>
                <tr>
                  <td>案件描述</td>
                  <td>{textValue(caseDetail.description, '暂无案件描述')}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        {needsConflictReview && (
          <section className='ng-panel'>
            <div className='ng-panel-head'>
              <h2 className='ng-panel-title'>下一步操作</h2>
            </div>
            <div className='batch-advice'>
              <strong>下一步：利益冲突复核</strong>
              <p>
                当前案件仍处于{statusLabel(caseDetail.status)}
                阶段。请先完成冲突复核，再进入接案审批或正式办理。
              </p>
              <Button
                type='primary'
                block
                icon={<FileSearchOutlined />}
                onClick={openConflictReview}
              >
                进入本案冲突复核
              </Button>
              <Button block onClick={supplementCaseIntake}>
                补充立案信息并重新检测
              </Button>
            </div>
          </section>
        )}

        {canReportSubjectRevision && (
          <section className='ng-panel span-2'>
            <div className='ng-panel-head'>
              <div className='ng-panel-title'>案件主体与冲突门禁</div>
            </div>
            <div className='batch-advice'>
              <strong>
                {subjectRevisionPending
                  ? textValue(pendingSubjectRevision.status, '') ===
                    'ENTITY_REGISTRATION_PENDING'
                    ? '新主体等待核查岗确认，受控动作已暂停'
                    : '主体变更待独立复核，受控动作已暂停'
                  : `当前生效主体版本 V${currentSubjectVersion}`}
              </strong>
              <p>
                新增或移除对方、第三人等案件主体，必须先登记变更并重新做利益冲突检查；独立复核完成前，系统不会允许发函、出具法律意见、推进审批或其他受控动作。
              </p>
              {textValue(caseDetail.conflict_coverage_status, '').toUpperCase() !== 'COMPLETE' && (
                <p className='danger-text'>
                  冲突档案覆盖范围尚未确认完整，当前案件不能被当作“已确认无冲突”。
                </p>
              )}
              {subjectRevisionPending ? (
                textValue(pendingSubjectRevision.status, '') ===
                'ENTITY_REGISTRATION_PENDING' ? (
                  <div className='batch-advice'>
                    <strong>{textValue(pendingSubjectRevision.candidate_name, '候选新主体')}</strong>
                    <p>
                      已提交核查岗确认。确认前不能运行重检，也不能推进受控对外动作；无需重复提交。
                    </p>
                  </div>
                ) : (
                  <Button
                    type='primary'
                    block
                    loading={subjectRevisionLoading}
                    onClick={continueSubjectRecheck}
                    disabled={!subjectRevisionID}
                  >
                    继续运行主体重检
                  </Button>
                )
              ) : (
                <Button
                  type='primary'
                  block
                  icon={<SafetyCertificateOutlined />}
                  onClick={openSubjectRevision}
                >
                  报告主体变更并重新复核
                </Button>
              )}
            </div>
          </section>
        )}

        <section className='ng-panel'>
          <div className='ng-panel-head'>
            <div className='ng-panel-title'>办理阶段</div>
          </div>
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
        </section>

        <section className='ng-panel'>
          <div className='ng-panel-head'>
            <div className='ng-panel-title'>客户信息</div>
          </div>
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
        </section>

        <section className='ng-panel'>
          <div className='ng-panel-head'>
            <div className='ng-panel-title'>负责团队</div>
          </div>
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
        </section>

        <section className='ng-panel span-2'>
          <div className='ng-panel-head'>
            <div className='ng-panel-title'>相关案件</div>
          </div>
          <div className='ng-table-wrap'>
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
                    <td>{dbCaseType(textValue(row.case_type))}</td>
                    <td>
                      <RiskTag text={statusLabel(row.status)} />
                    </td>
                    <td>{textValue(row.lawyer_name)}</td>
                    <td>{formatApiDate(row.updated_at)}</td>
                  </tr>
                ))}
                {relatedRows.length === 0 && (
                  <tr>
                    <td colSpan={6}>暂无同客户相关案件</td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </section>
      </div>

      <Modal
        title='报告案件主体变更'
        open={subjectRevisionOpen}
        onCancel={() => !subjectRevisionLoading && setSubjectRevisionOpen(false)}
        onOk={submitSubjectRevision}
        okText={
          subjectChangeType !== 'REMOVE_PARTY' && subjectEntityMode === 'new'
            ? '提交核查岗确认'
            : '登记并重新检测'
        }
        cancelText='取消'
        confirmLoading={subjectRevisionLoading}
        okButtonProps={{
          disabled:
            subjectChangeReason.trim().length < 5 ||
            (subjectChangeType === 'REMOVE_PARTY' || subjectEntityMode === 'existing'
              ? !selectedSubjectEntityID
              : newSubjectName.trim().length < 2 || newSubjectIdentityNumber.trim().length < 4),
        }}
      >
        <p>
          {subjectChangeType !== 'REMOVE_PARTY' && subjectEntityMode === 'new'
            ? '请按身份材料提报全新主体。核查岗完成全所去重和身份确认后，申请律师才能运行新的利益冲突检查；确认前原主体版本继续生效，受控动作保持暂停。'
            : '请选择当前账号可访问案件中已经登记的主体。搜索结果受案件权限和隔离墙保护；未显示某个名称，不代表全所档案中不存在该主体。系统会先登记变更，再自动运行新的利益冲突检查；独立复核完成前，原主体版本继续生效。'}
        </p>
        <div className='batch-field'>
          <label htmlFor='subject-change-type'>变更内容 *</label>
          <Select
            id='subject-change-type'
            value={subjectChangeType}
            onChange={(value) => {
              setSubjectChangeType(value)
              if (value === 'REMOVE_PARTY') setSubjectEntityMode('existing')
              setSelectedSubjectEntityID(null)
              setSubjectEntityQuery('')
              setSubjectEntityOptions([])
            }}
            options={[
              { value: 'ADD_OPPOSING_PARTY', label: '新增对方当事人' },
              { value: 'ADD_THIRD_PARTY', label: '新增第三人' },
              { value: 'REMOVE_PARTY', label: '移除已登记当事人' },
            ]}
          />
        </div>
        {subjectChangeType === 'REMOVE_PARTY' ? (
          <div className='batch-field'>
            <label htmlFor='subject-remove-party'>选择要移除的主体 *</label>
            <Select
              id='subject-remove-party'
              value={selectedSubjectEntityID || undefined}
              onChange={setSelectedSubjectEntityID}
              placeholder='请选择案件当前当事人'
              options={subjectParties.map((party) => ({
                value: party.entity_id,
                label: `${party.name}${party.identity_hint ? `（标识 ${party.identity_hint}）` : ''}`,
              }))}
              notFoundContent='当前案件没有可移除的结构化主体'
            />
          </div>
        ) : (
          <>
            <div className='batch-field'>
              <label htmlFor='subject-entity-mode'>主体来源 *</label>
              <Segmented
                id='subject-entity-mode'
                block
                value={subjectEntityMode}
                onChange={(value) => {
                  setSubjectEntityMode(value as 'existing' | 'new')
                  setSelectedSubjectEntityID(null)
                }}
                options={[
                  { value: 'existing', label: '选择已有主体' },
                  { value: 'new', label: '登记全新主体' },
                ]}
              />
            </div>
            {subjectEntityMode === 'existing' ? (
              <>
                <div className='batch-field'>
                  <label htmlFor='subject-entity-search'>搜索当前可访问的已登记主体 *</label>
                  <Input
                    id='subject-entity-search'
                    value={subjectEntityQuery}
                    onChange={(event) => setSubjectEntityQuery(event.target.value)}
                    placeholder='输入主体名称或曾用名，至少两个字'
                    suffix={subjectEntityLoading ? <ClockCircleOutlined spin /> : <SearchOutlined />}
                    allowClear
                  />
                </div>
                <div className='batch-field'>
                  <label htmlFor='subject-entity-select'>选择主体 *</label>
                  <Select
                    id='subject-entity-select'
                    value={selectedSubjectEntityID || undefined}
                    onChange={setSelectedSubjectEntityID}
                    placeholder={
                      subjectEntityQuery.trim().length < 2
                        ? '先输入名称进行搜索'
                        : '请选择搜索结果'
                    }
                    options={subjectEntityOptions.map((entity) => ({
                      value: entity.entity_id,
                      label: `${entity.name}${entity.identity_hint ? `（标识 ${entity.identity_hint}）` : ''}`,
                    }))}
                    notFoundContent={
                      subjectEntityQuery.trim().length < 2
                        ? '请输入至少两个字'
                        : '当前可访问范围内未找到主体'
                    }
                  />
                </div>
                {subjectEntityQuery.trim().length >= 2 &&
                  !subjectEntityLoading &&
                  subjectEntityOptions.length === 0 && (
                    <p className='danger-text'>
                      当前可访问范围内未找到该主体，这不等于全所无记录。可切换到“登记全新主体”，由冲突核查岗先做全所去重和身份确认。
                    </p>
                  )}
              </>
            ) : (
              <>
                <div className='batch-field'>
                  <label htmlFor='new-subject-name'>主体法定名称或证件姓名 *</label>
                  <Input id='new-subject-name' value={newSubjectName} onChange={(event) => setNewSubjectName(event.target.value)} placeholder='请按营业执照或身份证件填写' />
                </div>
                <div className='batch-field'>
                  <label htmlFor='new-subject-alias'>曾用名或别名</label>
                  <Input id='new-subject-alias' value={newSubjectAlias} onChange={(event) => setNewSubjectAlias(event.target.value)} placeholder='多个名称用逗号分隔；没有可不填' />
                </div>
                <div className='batch-field'>
                  <label htmlFor='new-subject-type'>主体类型 *</label>
                  <Select
                    id='new-subject-type'
                    value={newSubjectEntityType}
                    onChange={(value) => {
                      setNewSubjectEntityType(value)
                      setNewSubjectIdentityType(value === 'INDIVIDUAL' ? 'ID_CARD' : 'SOCIAL_CREDIT_CODE')
                    }}
                    options={[
                      { value: 'LEGAL_PERSON', label: '法人或企业' },
                      { value: 'INDIVIDUAL', label: '自然人' },
                      { value: 'ORGANIZATION', label: '其他组织' },
                    ]}
                  />
                </div>
                <div className='batch-field'>
                  <label htmlFor='new-subject-identity-type'>身份标识类型 *</label>
                  <Select
                    id='new-subject-identity-type'
                    value={newSubjectIdentityType}
                    onChange={setNewSubjectIdentityType}
                    options={
                      newSubjectEntityType === 'INDIVIDUAL'
                        ? [
                            { value: 'ID_CARD', label: '身份证件号码' },
                            { value: 'PASSPORT', label: '护照号码' },
                            { value: 'OTHER', label: '其他有效证件' },
                          ]
                        : [
                            { value: 'SOCIAL_CREDIT_CODE', label: '统一社会信用代码' },
                            { value: 'BUSINESS_LICENSE', label: '营业执照号码' },
                            { value: 'ORGANIZATION_CODE', label: '组织机构代码' },
                            { value: 'OTHER', label: '其他有效登记号码' },
                          ]
                    }
                  />
                </div>
                <div className='batch-field'>
                  <label htmlFor='new-subject-identity-number'>身份标识 *</label>
                  <Input.Password id='new-subject-identity-number' value={newSubjectIdentityNumber} onChange={(event) => setNewSubjectIdentityNumber(event.target.value)} placeholder='提交后加密保存，仅核查岗可核验脱敏信息' />
                </div>
                <p>提交后先由冲突核查岗做全所去重与身份确认，确认完成后才能运行本案冲突重检。</p>
              </>
            )}
          </>
        )}
        <div className='batch-field'>
          <label htmlFor='subject-change-reason'>变更原因 *</label>
          <Input.TextArea
            id='subject-change-reason'
            value={subjectChangeReason}
            onChange={(event) => setSubjectChangeReason(event.target.value)}
            rows={4}
            placeholder='例如：收到法院追加第三人通知，需要将该主体纳入本案当事人清单。'
          />
        </div>
        <p className='danger-text'>
          提交后案件会进入“待独立复核”状态。系统不会把“检索完成”直接当成“无冲突确认”。
        </p>
      </Modal>
    </div>
  )
}

export function CaseIntakeWorkbench() {
  const navigate = useNavigate()
  const { user: activeUser } = useAppStore()
  const isAssistant =
    activeUser?.roles.some((role) =>
      ['assistant', 'intake_assistant'].includes(role.toLowerCase()),
    ) ||
    ['assistant', 'intake_assistant'].includes(textValue(getUserInfo()?.role, '').toLowerCase())
  const [searchParams] = useSearchParams()
  const contextCaseID = searchParams.get('case_id') || ''
  const requestedClientID = numberValue(searchParams.get('client_id'), 0)
  const isSupplementMode = searchParams.get('mode') === 'supplement' && Boolean(contextCaseID)
  const currentUserID = textValue(getUserInfo()?.id, 'anonymous')
  const newIntakeIdempotencyKey = React.useRef(
    `case-intake-${currentUserID}-${Date.now()}-${Math.random().toString(36).slice(2)}`,
  )
  const draftKey = React.useMemo(
    () => scopedCaseIntakeDraftKey(currentUserID, isSupplementMode ? contextCaseID : undefined),
    [contextCaseID, currentUserID, isSupplementMode],
  )
  const [storedDraft, setStoredDraft] = React.useState<IntakeFormState | null>(() =>
    loadCaseIntakeDraft(draftKey),
  )
  const [form, setForm] = React.useState<IntakeFormState>({ ...defaultIntakeForm })
  const [draftActive, setDraftActive] = React.useState(false)
  const [sourceCase, setSourceCase] = React.useState<Record<string, any> | null>(null)
  const [runtime, setRuntime] = React.useState<IntakeRuntimeState>({ apiTimings: [] })
  const [submitting, setSubmitting] = React.useState(false)
  const [checkingConflict, setCheckingConflict] = React.useState(false)
  const [submissionNotice, setSubmissionNotice] = React.useState('')
  const [conflictCheckMissingFields, setConflictCheckMissingFields] = React.useState<
    ConflictCheckRequiredField[]
  >([])
  const [activeStep, setActiveStep] = React.useState(0)
  const [relatedParties, setRelatedParties] = React.useState<IntakeRelatedParty[]>([])
  const [relatedPartyDraft, setRelatedPartyDraft] = React.useState('')
  const [relatedPartyEntityType, setRelatedPartyEntityType] = React.useState('LEGAL_PERSON')
  const [relatedPartyIdentityType, setRelatedPartyIdentityType] = React.useState('SOCIAL_CREDIT_CODE')
  const [relatedPartyIdentityNumber, setRelatedPartyIdentityNumber] = React.useState('')
  const [caseTags, setCaseTags] = React.useState<string[]>([])
  const [tagDraft, setTagDraft] = React.useState('')
  const [clientOptions, setClientOptions] = React.useState<ClientOption[]>([])
  const [lawyerOptions, setLawyerOptions] = React.useState<LawyerOption[]>([])

  const conflictInputFingerprint = React.useMemo(
    () => caseIntakeConflictFingerprint(form, relatedParties),
    [form, relatedParties],
  )
  const isConflictResultStale = Boolean(
    runtime.conflict && runtime.conflictInputFingerprint !== conflictInputFingerprint,
  )
  const conflictCheckMissingSet = React.useMemo(
    () => new Set(conflictCheckMissingFields),
    [conflictCheckMissingFields],
  )
  const conflictCheckFieldError = (field: ConflictCheckRequiredField) =>
    conflictCheckMissingSet.has(field) ? '运行利益冲突检查前必须填写' : undefined

  React.useEffect(() => {
    window.localStorage.removeItem(legacyCaseIntakeDraftKey)
  }, [])

  React.useEffect(() => {
    if (!draftActive) return
    window.localStorage.setItem(draftKey, JSON.stringify(persistableCaseIntakeDraft(form)))
  }, [draftActive, draftKey, form])

  React.useEffect(() => {
    if (isAssistant) return
    if (!isSupplementMode) return
    let mounted = true
    apiRequest<any>(`/cases/${contextCaseID}`)
      .then((detail) => {
        if (!mounted) return
        const client = settingObject(detail.client)
        const lawyer = settingObject(detail.lawyer)
        setSourceCase(detail)
        setForm({
          ...defaultIntakeForm,
          title: textValue(detail.title, ''),
          caseType: textValue(detail.case_type, ''),
          priority: textValue(detail.priority, 'medium'),
          clientId: numberValue(detail.client_id || client.id, 0),
          clientName: textValue(detail.client_name || client.name, ''),
          lawyerId: numberValue(detail.lawyer_id || lawyer.id, 0),
          description: textValue(detail.description, ''),
          opponentName: textValue(detail.opposing_party, ''),
        })
        setDraftActive(true)
      })
      .catch((error) => message.error(error instanceof Error ? error.message : '加载当前案件失败'))
    return () => {
      mounted = false
    }
  }, [contextCaseID, isAssistant, isSupplementMode])

  React.useEffect(() => {
    if (isAssistant) return
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

        const requestedClient = clients.find((client) => client.id === requestedClientID)
        if (requestedClient) {
          setDraftActive(true)
          setForm((current) =>
            current.clientId
              ? current
              : { ...current, clientId: requestedClient.id, clientName: requestedClient.name },
          )
        }
      })
      .catch((error) => {
        message.error(error instanceof Error ? error.message : '加载客户或律师失败')
      })

    return () => {
      mounted = false
    }
  }, [isAssistant, requestedClientID])

  const recordTiming = (label: string, startedAt: number) => {
    setRuntime((current) => ({
      ...current,
      apiTimings: [
        {
          label,
          duration: Math.round(performance.now() - startedAt),
          at: new Date().toLocaleTimeString(),
        },
        ...current.apiTimings,
      ].slice(0, 5),
    }))
  }

  const updateForm = (key: keyof IntakeFormState, value: string | number) => {
    setDraftActive(true)
    const validationField = conflictCheckFieldForFormKey(key)
    if (validationField) {
      setConflictCheckMissingFields((current) =>
        current.filter((field) => field !== validationField),
      )
    }
    setForm((current) => ({ ...current, [key]: value }))
  }

  const updateClient = (clientId: number) => {
    const selectedClient = clientOptions.find((client) => client.id === clientId)
    setDraftActive(true)
    setConflictCheckMissingFields((current) => current.filter((field) => field !== 'client'))
    setForm((current) => ({
      ...current,
      clientId,
      clientName: selectedClient?.name || current.clientName,
    }))
  }

  const buildIntakePayload = () => {
    if (isAssistant) {
      return {
        title: form.title,
        case_type: form.caseType,
        priority: form.priority,
        description: form.description,
        materials: defaultMaterials,
      }
    }
    return {
      client_id: form.clientId,
      title: form.title,
      case_type: form.caseType,
      priority: form.priority,
      description: form.description,
      metadata: {
        source: 'batch01_real_api',
        subject_case_id: contextCaseID || undefined,
        subject_case_number: textValue(sourceCase?.case_number, '') || undefined,
        business_area: form.businessArea,
        sub_area: form.subArea,
        dispute_amount: form.disputeAmount,
        source_channel: form.sourceChannel,
        source_contact: form.sourceContact,
        investment_agreement_date: form.investmentAgreementDate,
        dispute_date: form.disputeDate,
        breach_date: form.breachDate,
        proposed_filing_date: form.proposedFilingDate,
        jurisdiction: form.jurisdiction,
        billing_method: form.billingMethod,
        fee_base: form.feeBase,
        contingency_rate: form.contingencyRate,
        minimum_fee: form.minimumFee,
        lawyer_id: form.lawyerId,
      },
      parties: [
        {
          entity_name: form.clientName,
          entity_type: 'company',
          party_role: 'client',
          relation_depth: 0,
        },
        {
          entity_name: form.opponentName,
          entity_type: form.opponentEntityType,
          party_role: 'opposing_party',
          relation_depth: 0,
          identity_type: form.opponentIdentityType,
          identity_number: form.opponentIdentityNumber,
          aliases: form.opponentAliases,
        },
        ...relatedParties.map((party) => ({
          entity_name: party.name,
          entity_type: party.entityType,
          party_role: 'related_party',
          relation_depth: 1,
          identity_type: party.identityType,
          identity_number: party.identityNumber,
        })),
      ],
      materials: defaultMaterials,
    }
  }

  const createIntake = async () => {
    if (!isAssistant && (!form.clientId || !form.clientName || !form.lawyerId)) {
      throw new Error('请先选择数据库中的客户和负责律师')
    }
    const startedAt = performance.now()
    const existingIntakeID = textValue(runtime.intake?.id || form.intakeId, '')
    const idempotencyKey = form.idempotencyKey || newIntakeIdempotencyKey.current
    if (!form.idempotencyKey) {
      setForm((current) => ({ ...current, idempotencyKey }))
    }
    const saved = await apiRequest<any>(
      existingIntakeID ? `/case-intakes/${existingIntakeID}` : '/case-intakes',
      {
        method: existingIntakeID ? 'PUT' : 'POST',
        headers: existingIntakeID ? undefined : { 'Idempotency-Key': idempotencyKey },
        body: JSON.stringify(buildIntakePayload()),
      },
    )
    const intake = existingIntakeID ? { ...runtime.intake, ...saved, id: existingIntakeID } : saved
    recordTiming(existingIntakeID ? '更新接案' : '创建接案', startedAt)
    setRuntime((current) => ({ ...current, intake }))
    setDraftActive(true)
    setForm((current) => ({
      ...current,
      intakeId: textValue(intake.id, existingIntakeID),
      intakeCode: textValue(intake.intake_code, current.intakeCode),
    }))
    message.success(
      existingIntakeID
        ? `接案草稿已更新：${intake.intake_code}`
        : `接案草稿已创建：${intake.intake_code}`,
    )
    return intake
  }

  const runConflictCheck = async (intake = runtime.intake) => {
    if (isAssistant) {
      throw new Error('助理不能运行利益冲突检查，请由负责律师确认当事人信息后操作')
    }
    if (!form.clientId || !form.clientName || !form.lawyerId) {
      throw new Error('请先选择数据库中的客户和负责律师')
    }
    const startedAt = performance.now()
    const intakeID = textValue(intake?.id, '')
    if (!intakeID) throw new Error('请先保存接案草稿，再运行利益冲突检查')
    await apiRequest(`/case-intakes/${intakeID}/facts-confirmation`, { method: 'POST', body: '{}' })
    const task = await apiRequest<any>(`/case-intakes/${intakeID}/conflict-check`, {
      method: 'POST',
      body: '{}',
    })
    setRuntime((current) => ({
      ...current,
      intake,
      conflict: undefined,
      conflictTask: task,
      conflictInputFingerprint,
    }))
    const conflict = task?.result
    if (!conflict) throw new Error('利益冲突检查仍在后台运行，请稍后在冲突清单查看结果')
    recordTiming('冲突检查', startedAt)
    setRuntime((current) => ({
      ...current,
      intake,
      conflict,
      conflictTask: { ...current.conflictTask, status: 'COMPLETED' },
      conflictInputFingerprint,
    }))
    setSubmissionNotice('')
    message.success('利益冲突检查已完成')
    return conflict
  }

  const submitApproval = async () => {
    if (isAssistant) {
      message.error('助理协作草稿不能提交审批，请由负责律师确认当事人信息并提交')
      return
    }
    setSubmitting(true)
    try {
      if (missingSubmitFields.length > 0) {
        const notice = `以下必填项未完成：${missingSubmitFields.join('、')}。请补充后再提交审批。`
        setSubmissionNotice(notice)
        message.error(notice)
        return
      }
      const intake = isConflictResultStale
        ? await createIntake()
        : runtime.intake || (await createIntake())
      let conflict =
        (!isConflictResultStale && runtime.conflict) || (await runConflictCheck(intake))
      const latestTaskID = textValue(
        conflict?.record?.id ||
          conflict?.record?.check_id ||
          conflict?.checkId ||
          runtime.conflictTask?.taskId,
        '',
      )
      if (latestTaskID) {
        try {
          const latest = await apiRequest<any>(`/conflict/tasks/${latestTaskID}/result`)
          if (latest?.result) conflict = latest.result
        } catch {
          // The frozen in-memory result remains authoritative if refreshing the task fails.
        }
      }
      let conflictReview = recordValue(conflict?.review)
      let conflictWaiver = recordValue(conflict?.waiver)
      const conflictTaskID = textValue(
        conflict?.record?.id || conflict?.record?.check_id || conflict?.checkId || conflict?.id,
        '',
      )
      if (conflictTaskID) {
        try {
          const reviewResponse = await apiRequest<any>(`/conflict/tasks/${conflictTaskID}/review`)
          conflictReview = recordValue(reviewResponse?.review || reviewResponse)
        } catch {
          // A missing review is an expected state and is handled by the gate below.
        }
        try {
          const waiverResponse = await apiRequest<any>(`/conflict/tasks/${conflictTaskID}/waiver`)
          conflictWaiver = recordValue(waiverResponse)
        } catch {
          // No waiver is the normal state for clear and manually reviewed checks.
        }
      }
      const decisionView = deriveConflictDecisionViewModel(conflict, {
        review: conflictReview,
        waiver: conflictWaiver,
      })
      if (!decisionView.canSubmit) {
        const notice = decisionView.guidance
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
              case_type: form.caseType,
              priority: form.priority,
              lawyer_id: form.lawyerId,
              billing_method: form.billingMethod,
            },
            parties: [
              { role: 'client', name: form.clientName },
              {
                role: 'opposing_party',
                name: form.opponentName,
                entity_type: form.opponentEntityType,
                identity_type: form.opponentIdentityType,
              },
              ...relatedParties.map((party) => ({
                role: 'related_party',
                name: party.name,
                party_type: party.role,
                entity_type: party.entityType,
                identity_type: party.identityType,
              })),
            ],
            materials: defaultMaterials,
            tags: caseTags,
            conflict_result: conflict,
          },
          case_creation_config: {
            title: form.title,
            description: form.description,
            client_id: form.clientId,
            case_type: form.caseType,
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
      window.localStorage.removeItem(draftKey)
      message.success('已提交真实审批')
      navigate(`/approval/${approval.approval_id}`)
    } catch (error) {
      message.error(error instanceof Error ? error.message : '提交审批失败')
    } finally {
      setSubmitting(false)
    }
  }

  const selectedLawyer = lawyerOptions.find((lawyer) => lawyer.id === form.lawyerId)
  const missingSubmitFields = (
    isAssistant
      ? []
      : [
          !form.title && '案件名称',
          !form.caseType && '案件类型',
          !form.businessArea && '业务领域',
          !form.subArea && '子领域',
          (!form.clientId || !form.clientName) && '客户',
          !form.opponentName && '对方当事人',
          !form.opponentIdentityNumber && '对方身份标识',
          !form.description && '案情摘要',
          !form.lawyerId && '负责律师',
          (!runtime.conflict || isConflictResultStale) && '利益冲突检查',
        ]
  ).filter(Boolean) as string[]
  const completedSubmitFields = 10 - missingSubmitFields.length
  const intakeCompleteness = Math.round((completedSubmitFields / 10) * 100)
  const overviewHint =
    missingSubmitFields.length > 0
      ? `还需补充：${missingSubmitFields.join('、')}`
      : '必填信息已完成'
  const intakeSteps = isAssistant
    ? [{ title: '协作草稿', desc: '整理案件摘要与材料清单' }]
    : [
        { title: '基本信息', desc: '案件与当事人信息' },
        { title: '利益冲突检查', desc: '自动检测与人工复核' },
        { title: '团队与费用', desc: '团队指派与收费安排' },
        { title: '文档与材料', desc: '材料清单与附件上传' },
        { title: '立案提交', desc: '提交审批并创建案件' },
      ]
  const currentStep = intakeSteps[activeStep] || intakeSteps[0]
  const conflictDecisionView = deriveConflictDecisionViewModel(runtime.conflict, {
    stale: isConflictResultStale,
  })
  const conflictRiskText = runtime.conflict
    ? conflictDecisionStatusLabel(conflictDecisionView)
    : '未检测'
  const conflictStatusClass = conflictDecisionView.canSubmit ? 'success-text' : 'danger-text'
  const conflictTaskID = textValue(
    runtime.conflict?.record?.id ||
      runtime.conflict?.record?.check_id ||
      runtime.conflict?.checkId ||
      runtime.conflictTask?.taskId,
    '',
  )
  const conflictEvidence = primaryConflictEvidence(runtime.conflict)
  const conflictReviewParams = new URLSearchParams()
  if (conflictTaskID) conflictReviewParams.set('task_id', conflictTaskID)
  if (runtime.intake?.id) conflictReviewParams.set('intake_id', textValue(runtime.intake.id))
  if (contextCaseID) conflictReviewParams.set('case_id', contextCaseID)
  if (sourceCase?.case_number)
    conflictReviewParams.set('case_number', textValue(sourceCase.case_number))
  if (form.title) conflictReviewParams.set('case_title', form.title)
  const conflictReviewPath = `/conflict?${conflictReviewParams.toString()}`
  const submitDisabled =
    submitting || missingSubmitFields.length > 0 || !conflictDecisionView.canSubmit
  const submitTooltip =
    missingSubmitFields.length > 0
      ? `请先完成：${missingSubmitFields.join('、')}`
      : conflictDecisionView.canSubmit
        ? ''
        : conflictDecisionView.guidance

  const addRelatedParty = () => {
    const name = relatedPartyDraft.trim()
    if (name.length < 2 || relatedPartyIdentityNumber.trim().length < 4) {
      message.error('请填写相关方名称和可核验身份标识后再添加')
      return
    }
    setRelatedParties((current) => [
      ...current,
      {
        name,
        role: '待确认',
        entityType: relatedPartyEntityType,
        identityType: relatedPartyIdentityType,
        identityNumber: relatedPartyIdentityNumber.trim(),
      },
    ])
    setRelatedPartyDraft('')
    setRelatedPartyIdentityNumber('')
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

  const continueStoredDraft = () => {
    if (!storedDraft) return
    setForm(storedDraft)
    if (storedDraft.intakeId) {
      setRuntime((current) => ({
        ...current,
        intake: { id: storedDraft.intakeId, intake_code: storedDraft.intakeCode },
      }))
    }
    setStoredDraft(null)
    setDraftActive(true)
    message.success('已恢复当前账号的未完成草稿')
  }

  const discardStoredDraft = () => {
    window.localStorage.removeItem(draftKey)
    setStoredDraft(null)
    if (!isSupplementMode) {
      setForm({ ...defaultIntakeForm })
      setDraftActive(false)
    }
    message.success('旧草稿已放弃')
  }

  const handleCreateIntake = async () => {
    try {
      await createIntake()
    } catch (error) {
      message.error(error instanceof Error ? error.message : '暂存失败')
    }
  }

  const handleRunConflictCheck = async () => {
    if (isAssistant) {
      message.error('助理不能运行利益冲突检查，请由负责律师确认当事人信息后操作')
      return
    }
    if (checkingConflict) return
    const missing = getMissingConflictCheckFields(form)
    if (missing.length > 0) {
      setConflictCheckMissingFields(missing)
      const labels = missing.map((field) => conflictCheckRequiredFieldLabels[field])
      const notice = `以下必填项未完成：${labels.join('、')}，请补充后再运行冲突检查`
      setSubmissionNotice(notice)
      message.error(notice)
      return
    }
    setConflictCheckMissingFields([])
    setSubmissionNotice('')
    setCheckingConflict(true)
    try {
      // Always persist the current form before confirming facts. Reusing an
      // existing intake here would let the UI label an old server-side check
      // as the result for newly edited parties.
      const intake = await createIntake()
      await runConflictCheck(intake)
      setActiveStep(1)
      message.success('利益冲突检查已完成，当前草稿：' + (form.title || '未命名案件'))
    } catch (error) {
      message.error(error instanceof Error ? error.message : '冲突检查失败，请检查网络后重试')
    } finally {
      setCheckingConflict(false)
    }
  }

  const handleCancelIntake = () => {
    Modal.confirm({
      title: '取消立案',
      content: '放弃后将清除当前账号在本案件下的本地草稿，未提交信息不会进入正式案件流程。',
      okText: '放弃草稿并返回',
      cancelText: '继续编辑',
      onOk: () => {
        window.localStorage.removeItem(draftKey)
        navigate('/case')
      },
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
      content:
        '当前 MVP 试用版暂不跳转文档中心。你可以继续保留本次立案上下文，已填写内容不会丢失。',
      okText: '继续立案',
    })
  }

  return (
    <div className='batch-page intake-page'>
      <PageHeader
        eyebrow='案件管理 / 新建案件 / 立案工作台'
        title={isSupplementMode ? '补充案件信息并重新检测' : '新建案件立案工作台'}
        subtitle={
          isSupplementMode
            ? `当前案件：${textValue(sourceCase?.case_number, contextCaseID)} ${textValue(sourceCase?.title, '')}`
            : '新建案件默认使用空白表单，未完成草稿需要手动恢复。'
        }
        actions={
          <span className='batch-autosave'>
            加载耗时：
            {runtime.apiTimings[0]
              ? `${runtime.apiTimings[0].label} ${runtime.apiTimings[0].duration}ms`
              : '待调用'}
          </span>
        }
      />

      {storedDraft && (
        <div className='batch-advice' role='status'>
          <strong>发现当前账号的未完成草稿：{storedDraft.title || '未命名案件'}</strong>
          <p>草稿不会自动覆盖当前表单。请选择继续草稿或放弃旧草稿。</p>
          <Space>
            <Button type='primary' onClick={continueStoredDraft}>
              继续未完成草稿
            </Button>
            <Button onClick={discardStoredDraft}>放弃旧草稿</Button>
          </Space>
        </div>
      )}

      {isAssistant && (
        <div className='batch-advice' role='status'>
          <strong>助理协作草稿</strong>
          <p>
            可整理案件摘要和材料清单；客户、对方、关联方及身份标识须由负责律师确认后录入。该草稿不能运行冲突检查或提交审批。
          </p>
        </div>
      )}

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
                  <button
                    onClick={() => setActiveStep((current) => Math.max(0, current - 1))}
                    disabled={activeStep === 0}
                  >
                    上一步
                  </button>
                  <button
                    onClick={() =>
                      setActiveStep((current) => Math.min(intakeSteps.length - 1, current + 1))
                    }
                    disabled={activeStep === intakeSteps.length - 1}
                  >
                    下一步
                  </button>
                </div>
                <div className='batch-form-grid four'>
                  <div className={`batch-field ${conflictCheckMissingSet.has('title') ? 'has-error' : ''}`}>
                    <label htmlFor='intake-title'>案件名称 *</label>
                    <Input
                      id='intake-title'
                      name='title'
                      value={form.title}
                      status={conflictCheckMissingSet.has('title') ? 'error' : undefined}
                      aria-invalid={conflictCheckMissingSet.has('title')}
                      onChange={(event) => updateForm('title', event.target.value)}
                    />
                    {conflictCheckFieldError('title') && (
                      <span className='batch-field-error'>{conflictCheckFieldError('title')}</span>
                    )}
                  </div>
                  <div className={`batch-field ${conflictCheckMissingSet.has('caseType') ? 'has-error' : ''}`}>
                    <label htmlFor='intake-case-type'>案件类型 *</label>
                    <Select
                      id='intake-case-type'
                      value={form.caseType || undefined}
                      status={conflictCheckMissingSet.has('caseType') ? 'error' : undefined}
                      aria-invalid={conflictCheckMissingSet.has('caseType')}
                      onChange={(value) => updateForm('caseType', value)}
                      options={intakeCaseTypeOptions}
                    />
                    {conflictCheckFieldError('caseType') && (
                      <span className='batch-field-error'>{conflictCheckFieldError('caseType')}</span>
                    )}
                  </div>
                  <div className='batch-field'>
                    <label htmlFor='intake-stage'>案件阶段 *</label>
                    <Select
                      id='intake-stage'
                      value='潜在受理'
                      disabled
                      options={[{ value: '潜在受理' }]}
                    />
                  </div>
                  <div className='batch-field'>
                    <label htmlFor='intake-dispute-amount'>预计争议金额</label>
                    <Input
                      id='intake-dispute-amount'
                      name='disputeAmount'
                      value={form.disputeAmount}
                      onChange={(event) => updateForm('disputeAmount', event.target.value)}
                      placeholder='请输入预计争议金额'
                    />
                  </div>
                  <div className={`batch-field ${conflictCheckMissingSet.has('businessArea') ? 'has-error' : ''}`}>
                    <label htmlFor='intake-business-area'>业务领域 *</label>
                    <Select
                      id='intake-business-area'
                      value={form.businessArea || undefined}
                      status={conflictCheckMissingSet.has('businessArea') ? 'error' : undefined}
                      aria-invalid={conflictCheckMissingSet.has('businessArea')}
                      onChange={(value) => updateForm('businessArea', value)}
                      placeholder='请选择业务领域'
                      options={[
                        { value: '公司与并购' },
                        { value: '争议解决' },
                        { value: '建设工程' },
                      ]}
                    />
                    {conflictCheckFieldError('businessArea') && (
                      <span className='batch-field-error'>{conflictCheckFieldError('businessArea')}</span>
                    )}
                  </div>
                  <div className={`batch-field ${conflictCheckMissingSet.has('subArea') ? 'has-error' : ''}`}>
                    <label htmlFor='intake-sub-area'>子领域 *</label>
                    <Select
                      id='intake-sub-area'
                      value={form.subArea || undefined}
                      status={conflictCheckMissingSet.has('subArea') ? 'error' : undefined}
                      aria-invalid={conflictCheckMissingSet.has('subArea')}
                      onChange={(value) => updateForm('subArea', value)}
                      placeholder='请选择子领域'
                      options={[
                        { value: '投资与融资' },
                        { value: '商事诉讼' },
                        { value: '工程合同' },
                      ]}
                    />
                    {conflictCheckFieldError('subArea') && (
                      <span className='batch-field-error'>{conflictCheckFieldError('subArea')}</span>
                    )}
                  </div>
                  <div className='batch-field'>
                    <label htmlFor='intake-source-channel'>案源渠道</label>
                    <Select
                      id='intake-source-channel'
                      value={form.sourceChannel || undefined}
                      onChange={(value) => updateForm('sourceChannel', value)}
                      placeholder='请选择案源渠道'
                      options={[
                        { value: '现有客户介绍' },
                        { value: '律师转介' },
                        { value: '公开咨询' },
                      ]}
                    />
                  </div>
                  <div className='batch-field'>
                    <label htmlFor='intake-source-contact'>案源联系人</label>
                    <Input
                      id='intake-source-contact'
                      name='sourceContact'
                      value={form.sourceContact}
                      onChange={(event) => updateForm('sourceContact', event.target.value)}
                      placeholder='请输入案源联系人'
                    />
                  </div>
                </div>
                <div className='batch-wide-label'>
                  <label htmlFor='intake-description'>案情摘要 *</label>
                  <Input.TextArea
                    id='intake-description'
                    name='description'
                    rows={3}
                    value={form.description}
                    onChange={(event) => updateForm('description', event.target.value)}
                  />
                </div>
                <div className='batch-form-grid five'>
                  <div className='batch-field'>
                    <label htmlFor='intake-investment-date'>投资协议签署日</label>
                    <Input
                      id='intake-investment-date'
                      name='investmentAgreementDate'
                      type='date'
                      value={form.investmentAgreementDate}
                      onChange={(event) =>
                        updateForm('investmentAgreementDate', event.target.value)
                      }
                    />
                  </div>
                  <div className='batch-field'>
                    <label htmlFor='intake-dispute-date'>争议发生日</label>
                    <Input
                      id='intake-dispute-date'
                      name='disputeDate'
                      type='date'
                      value={form.disputeDate}
                      onChange={(event) => updateForm('disputeDate', event.target.value)}
                    />
                  </div>
                  <div className='batch-field'>
                    <label htmlFor='intake-breach-date'>对方违约日</label>
                    <Input
                      id='intake-breach-date'
                      name='breachDate'
                      type='date'
                      value={form.breachDate}
                      onChange={(event) => updateForm('breachDate', event.target.value)}
                    />
                  </div>
                  <div className='batch-field'>
                    <label htmlFor='intake-filing-date'>拟起诉日</label>
                    <Input
                      id='intake-filing-date'
                      name='proposedFilingDate'
                      type='date'
                      value={form.proposedFilingDate}
                      onChange={(event) => updateForm('proposedFilingDate', event.target.value)}
                    />
                  </div>
                  <div className='batch-field'>
                    <label htmlFor='intake-jurisdiction'>诉讼地</label>
                    <Input
                      id='intake-jurisdiction'
                      name='jurisdiction'
                      value={form.jurisdiction}
                      onChange={(event) => updateForm('jurisdiction', event.target.value)}
                      placeholder='请输入诉讼地'
                    />
                  </div>
                </div>
              </SectionCard>

              {isAssistant ? (
                <SectionCard title='当事人信息'>
                  <p>
                    为避免未经复核的身份信息进入冲突检索，助理协作草稿不提供客户、对方或关联方录入。请由负责律师补充并确认。
                  </p>
                </SectionCard>
              ) : (
                <SectionCard title='当事人信息'>
                  <div className='batch-party-grid'>
                    <article>
                      <strong>我方当事人（客户） *</strong>
                      <div className='batch-party-card green'>
                        <TeamOutlined />
                        <div>
                          <Select
                            id='intake-client'
                            aria-label='我方当事人（客户）'
                            value={form.clientId || undefined}
                            status={conflictCheckMissingSet.has('client') ? 'error' : undefined}
                            aria-invalid={conflictCheckMissingSet.has('client')}
                            onChange={updateClient}
                            showSearch
                            optionFilterProp='label'
                            placeholder='选择数据库客户'
                            options={clientOptions.map((client) => ({
                              value: client.id,
                              label: client.displayLabel,
                            }))}
                            style={{ minWidth: 240 }}
                          />
                          {conflictCheckFieldError('client') && (
                            <span className='batch-field-error'>
                              {conflictCheckFieldError('client')}
                            </span>
                          )}
                          <RiskTag text='现有客户' />
                          <p>
                            {form.clientId ? '已选择数据库客户' : '尚未选择客户'}
                            <br />
                            提交时将关联客户主档案
                          </p>
                        </div>
                      </div>
                    </article>
                    <span className='vs-badge'>VS</span>
                    <article>
                      <strong>对方当事人（被告/对方） *</strong>
                      <div className='batch-party-card blue'>
                        <BankOutlined />
                        <div>
                          <label htmlFor='intake-opponent'>法定名称或证件姓名</label>
                          <Input
                            id='intake-opponent'
                            name='opponentName'
                            value={form.opponentName}
                            status={conflictCheckMissingSet.has('opponentName') ? 'error' : undefined}
                            aria-invalid={conflictCheckMissingSet.has('opponentName')}
                            onChange={(event) => updateForm('opponentName', event.target.value)}
                            placeholder='输入对方当事人名称'
                          />
                          {conflictCheckFieldError('opponentName') && (
                            <span className='batch-field-error'>
                              {conflictCheckFieldError('opponentName')}
                            </span>
                          )}
                          <label htmlFor='intake-opponent-entity-type'>主体类型</label>
                          <Select
                            id='intake-opponent-entity-type'
                            value={form.opponentEntityType}
                            onChange={(value) => {
                              updateForm('opponentEntityType', value)
                              updateForm(
                                'opponentIdentityType',
                                value === 'INDIVIDUAL' ? 'ID_CARD' : 'SOCIAL_CREDIT_CODE',
                              )
                            }}
                            options={[
                              { value: 'LEGAL_PERSON', label: '法人或企业' },
                              { value: 'INDIVIDUAL', label: '自然人' },
                              { value: 'ORGANIZATION', label: '其他组织' },
                            ]}
                          />
                          <label htmlFor='intake-opponent-identity-type'>身份标识类型</label>
                          <Select
                            id='intake-opponent-identity-type'
                            value={form.opponentIdentityType}
                            onChange={(value) => updateForm('opponentIdentityType', value)}
                            options={
                              form.opponentEntityType === 'INDIVIDUAL'
                                ? [
                                    { value: 'ID_CARD', label: '身份证件号码' },
                                    { value: 'PASSPORT', label: '护照号码' },
                                    { value: 'OTHER', label: '其他有效证件' },
                                  ]
                                : [
                                    { value: 'SOCIAL_CREDIT_CODE', label: '统一社会信用代码' },
                                    { value: 'BUSINESS_LICENSE', label: '营业执照号码' },
                                    { value: 'ORGANIZATION_CODE', label: '组织机构代码' },
                                    { value: 'OTHER', label: '其他有效登记号码' },
                                  ]
                            }
                          />
                          <label htmlFor='intake-opponent-identity-number'>身份标识</label>
                          <Input.Password
                            id='intake-opponent-identity-number'
                            value={form.opponentIdentityNumber}
                            status={
                              conflictCheckMissingSet.has('opponentIdentityNumber') ? 'error' : undefined
                            }
                            aria-invalid={conflictCheckMissingSet.has('opponentIdentityNumber')}
                            onChange={(event) =>
                              updateForm('opponentIdentityNumber', event.target.value)
                            }
                            placeholder='输入证件号或统一社会信用代码'
                          />
                          {conflictCheckFieldError('opponentIdentityNumber') && (
                            <span className='batch-field-error'>
                              {conflictCheckFieldError('opponentIdentityNumber')}
                            </span>
                          )}
                          <label htmlFor='intake-opponent-aliases'>曾用名或别名</label>
                          <Input
                            id='intake-opponent-aliases'
                            value={form.opponentAliases}
                            onChange={(event) => updateForm('opponentAliases', event.target.value)}
                            placeholder='多个名称用逗号分隔；没有可不填'
                          />
                          <RiskTag text='新增对方' />
                          <p>身份标识将加密保存，并与名称、曾用名共同用于冲突检索。</p>
                        </div>
                      </div>
                    </article>
                    <article>
                      <strong>其他相关方（可选）</strong>
                      <div className='batch-related-party'>
                        {relatedParties.map((party, index) => (
                          <p key={`${party.name}-${index}`}>
                            {party.name} <RiskTag text={party.role} />
                            <Button
                              type='link'
                              size='small'
                              onClick={() =>
                                setRelatedParties((current) =>
                                  current.filter((_, itemIndex) => itemIndex !== index),
                                )
                              }
                            >
                              移除
                            </Button>
                          </p>
                        ))}
                        <div className='batch-inline-add'>
                          <Input
                            id='intake-related-party'
                            name='relatedParty'
                            value={relatedPartyDraft}
                            onChange={(event) => setRelatedPartyDraft(event.target.value)}
                            placeholder='输入关联公司、实控人、保证人'
                          />
                          <Select
                            aria-label='相关方主体类型'
                            value={relatedPartyEntityType}
                            onChange={(value) => {
                              setRelatedPartyEntityType(value)
                              setRelatedPartyIdentityType(
                                value === 'INDIVIDUAL' ? 'ID_CARD' : 'SOCIAL_CREDIT_CODE',
                              )
                            }}
                            options={[
                              { value: 'LEGAL_PERSON', label: '法人或企业' },
                              { value: 'INDIVIDUAL', label: '自然人' },
                              { value: 'ORGANIZATION', label: '其他组织' },
                            ]}
                          />
                          <Select
                            aria-label='相关方身份标识类型'
                            value={relatedPartyIdentityType}
                            onChange={setRelatedPartyIdentityType}
                            options={
                              relatedPartyEntityType === 'INDIVIDUAL'
                                ? [
                                    { value: 'ID_CARD', label: '身份证件号码' },
                                    { value: 'PASSPORT', label: '护照号码' },
                                    { value: 'OTHER', label: '其他有效证件' },
                                  ]
                                : [
                                    { value: 'SOCIAL_CREDIT_CODE', label: '统一社会信用代码' },
                                    { value: 'BUSINESS_LICENSE', label: '营业执照号码' },
                                    { value: 'ORGANIZATION_CODE', label: '组织机构代码' },
                                    { value: 'OTHER', label: '其他有效登记号码' },
                                  ]
                            }
                          />
                          <Input.Password
                            aria-label='相关方身份标识'
                            value={relatedPartyIdentityNumber}
                            onChange={(event) => setRelatedPartyIdentityNumber(event.target.value)}
                            placeholder='证件号或统一社会信用代码'
                          />
                          <Button type='link' icon={<PlusOutlined />} onClick={addRelatedParty}>
                            添加相关方
                          </Button>
                        </div>
                      </div>
                    </article>
                  </div>
                  <div className='batch-tags'>
                    {caseTags.map((tag) => (
                      <Tag
                        key={tag}
                        closable
                        onClose={() =>
                          setCaseTags((current) => current.filter((item) => item !== tag))
                        }
                      >
                        {tag}
                      </Tag>
                    ))}
                    <Input
                      id='intake-tag'
                      name='caseTag'
                      size='small'
                      value={tagDraft}
                      onChange={(event) => setTagDraft(event.target.value)}
                      placeholder='案由/行业/风险标签'
                      onPressEnter={addCaseTag}
                    />
                    <Button size='small' icon={<PlusOutlined />} onClick={addCaseTag}>
                      添加标签
                    </Button>
                  </div>
                </SectionCard>
              )}
            </>
          )}

          {activeStep === 1 && (
            <SectionCard title='利益冲突检查'>
              {conflictCheckMissingFields.length > 0 && (
                <Alert
                  className='conflict-check-validation'
                  type='error'
                  showIcon
                  message='冲突检查前置校验未通过'
                  description={
                    <>
                      <p>
                        请补充：
                        {conflictCheckMissingFields
                          .map((field) => conflictCheckRequiredFieldLabels[field])
                          .join('、')}
                        。
                      </p>
                      <p>未创建接案草稿，也未发起冲突检查。</p>
                    </>
                  }
                />
              )}
              <div className='batch-approval-risk'>
                <SafetyCertificateOutlined />
                <div>
                  <strong>
                    检查状态：
                    {runtime.conflictTask && !runtime.conflict
                      ? `后台检测中 · ${statusLabel(runtime.conflictTask.status)}`
                      : conflictDecisionView.stale
                        ? conflictDecisionView.headline
                        : runtime.conflict
                          ? `已完成 · ${conflictDecisionView.headline}`
                          : conflictDecisionView.headline}
                  </strong>
                  <p>{conflictDecisionView.guidance}</p>
                </div>
              </div>
              <div
                className='batch-intake-draft-summary'
                style={{
                  background: '#f6ffed',
                  border: '1px solid #b7eb8f',
                  borderRadius: 6,
                  padding: '8px 12px',
                  marginBottom: 12,
                }}
              >
                <strong>本次立案草稿</strong>
                {runtime.intake?.intake_code && (
                  <span style={{ marginLeft: 8 }}>
                    接案草稿已创建：{runtime.intake.intake_code}
                  </span>
                )}
                <span style={{ marginLeft: 8 }}>案件：{form.title || '未填写'}</span>
                <span style={{ marginLeft: 8 }}>客户：{form.clientName || '未选择'}</span>
                <span style={{ marginLeft: 8 }}>对方：{form.opponentName || '未填写'}</span>
              </div>
              <DataTable>
                <table>
                  <tbody>
                    <tr>
                      <td>客户</td>
                      <td>{form.clientName || '未选择'}</td>
                    </tr>
                    <tr>
                      <td>对方当事人</td>
                      <td>{form.opponentName || '未填写'}</td>
                    </tr>
                    <tr>
                      <td>其他相关方</td>
                      <td>{relatedParties.map((party) => party.name).join('、') || '暂无'}</td>
                    </tr>
                    <tr>
                      <td>检测结果</td>
                      <td>
                        {conflictDecisionView.stale
                          ? '结果已过期'
                          : runtime.conflict
                            ? `处置状态：${conflictRiskText}。自动检索结果不能替代独立人工复核结论。`
                            : '尚未加载数据'}
                      </td>
                    </tr>
                    <tr>
                      <td>接案决策</td>
                      <td>{conflictDecisionView.headline}</td>
                    </tr>
                    <tr>
                      <td>检索范围</td>
                      <td>
                        {conflictDecisionView.stale
                          ? '等待按最新客户、对方、相关方和团队重新检索'
                          : runtime.conflict
                            ? textValue(
                                runtime.conflict.decision?.coverageNotice,
                                '以本所已录入数据为限',
                              )
                            : '待检测'}
                      </td>
                    </tr>
                    <tr>
                      <td>提交提示</td>
                      <td>{submissionNotice || conflictDecisionView.guidance}</td>
                    </tr>
                  </tbody>
                </table>
              </DataTable>
              {runtime.conflict && !conflictDecisionView.stale && (
                <DiagnosticDetails label='查看冲突检测审计信息'>
                  <p>检测任务编号：{conflictTaskID || '-'}</p>
                  <p>
                    主命中实体：
                    {textValue(
                      conflictEvidence.requestedParty ||
                        conflictEvidence.queryName ||
                        runtime.conflict.matched_subject,
                      '-',
                    )}
                  </p>
                  <p>
                    命中规则：{textValue(conflictEvidence.ruleCode || conflictEvidence.ruleId, '-')}
                  </p>
                  <p>
                    匹配方式：
                    {conflictMatchTypeLabel(
                      conflictEvidence.matchType || conflictEvidence.algorithm,
                    )}
                  </p>
                  <p>
                    来源案件：
                    {textValue(
                      conflictEvidence.sourceCaseNumber ||
                        conflictEvidence.sourceCaseName ||
                        conflictEvidence.caseNumber,
                      '-',
                    )}
                  </p>
                </DiagnosticDetails>
              )}
              <Space>
                <Button icon={<ArrowLeftOutlined />} onClick={() => setActiveStep(0)}>
                  返回基本信息
                </Button>
                <Button
                  type='primary'
                  aria-label={
                    conflictDecisionView.stale ? '保存最新输入并检测' : '运行利益冲突检查'
                  }
                  icon={<SafetyCertificateOutlined />}
                  loading={checkingConflict}
                  onClick={handleRunConflictCheck}
                >
                  {conflictDecisionView.stale ? '保存最新输入并检测' : '运行利益冲突检查'}
                </Button>
                <Button
                  onClick={() => setActiveStep(2)}
                  disabled={!runtime.conflict || isConflictResultStale}
                >
                  进入团队与费用
                </Button>
              </Space>
            </SectionCard>
          )}

          {activeStep === 2 && (
            <SectionCard title='团队与费用'>
              <div className='batch-form-grid four'>
                <div className={`batch-field ${conflictCheckMissingSet.has('lawyer') ? 'has-error' : ''}`}>
                  <label htmlFor='intake-lawyer'>负责律师 *</label>
                  <Select
                    id='intake-lawyer'
                    aria-label='负责律师 *'
                    value={form.lawyerId || undefined}
                    status={conflictCheckMissingSet.has('lawyer') ? 'error' : undefined}
                    aria-invalid={conflictCheckMissingSet.has('lawyer')}
                    onChange={(value) => updateForm('lawyerId', value)}
                    showSearch
                    optionFilterProp='label'
                    placeholder='选择负责律师'
                    options={lawyerOptions.map((lawyer) => ({
                      value: lawyer.id,
                      label: `${lawyer.name}${lawyer.department ? ` · ${lawyer.department}` : ''}`,
                    }))}
                  />
                  {conflictCheckFieldError('lawyer') && (
                    <span className='batch-field-error'>{conflictCheckFieldError('lawyer')}</span>
                  )}
                </div>
                <div className='batch-field'>
                  <label htmlFor='intake-billing-method'>计费方式</label>
                  <Select
                    id='intake-billing-method'
                    value={form.billingMethod || undefined}
                    onChange={(value) => updateForm('billingMethod', value)}
                    options={[
                      { value: 'hourly', label: '小时计费' },
                      { value: 'fixed', label: '固定收费' },
                      { value: 'contingency', label: '风险代理' },
                    ]}
                  />
                </div>
                <div className='batch-field'>
                  <label htmlFor='intake-fee-base'>收费基数</label>
                  <Select
                    id='intake-fee-base'
                    value={form.feeBase || undefined}
                    onChange={(value) => updateForm('feeBase', value)}
                    placeholder='请选择收费基数'
                    options={[{ value: '争议标的金额' }, { value: '固定金额' }]}
                  />
                </div>
                <div className='batch-field'>
                  <label htmlFor='intake-contingency-rate'>风险代理比例</label>
                  <Input
                    id='intake-contingency-rate'
                    name='contingencyRate'
                    value={form.contingencyRate}
                    onChange={(event) => updateForm('contingencyRate', event.target.value)}
                    placeholder='请输入风险代理比例'
                  />
                </div>
                <div className='batch-field'>
                  <label htmlFor='intake-minimum-fee'>最低收费保障</label>
                  <Input
                    id='intake-minimum-fee'
                    name='minimumFee'
                    value={form.minimumFee}
                    onChange={(event) => updateForm('minimumFee', event.target.value)}
                    placeholder='请输入最低收费保障'
                  />
                </div>
              </div>
            </SectionCard>
          )}

          {activeStep === 3 && (
            <SectionCard title='文档与材料'>
              <DataTable>
                <table>
                  <thead>
                    <tr>
                      <th>材料名称</th>
                      <th>类型</th>
                      <th>状态</th>
                      <th>操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    {defaultMaterials.map((item) => (
                      <tr key={item.name}>
                        <td>{item.name}</td>
                        <td>{materialTypeLabel(item.material_type)}</td>
                        <td>
                          <RiskTag text={statusLabel(item.status)} />
                        </td>
                        <td>
                          <Button type='link' disabled>
                            文件管理暂未开放
                          </Button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </DataTable>
              <Button icon={<CloudUploadOutlined />} onClick={showMaterialsUnavailable}>
                打开文件材料归档
              </Button>
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
                  <p key={line}>
                    <span>{line.split(' ')[0]}</span>
                    <strong>{line.substring(line.indexOf(' ') + 1)}</strong>
                  </p>
                ))}
              </div>
              <Space>
                <Button onClick={() => setActiveStep(3)}>返回材料</Button>
              </Space>
              {!conflictDecisionView.canSubmit && conflictDecisionView.decision !== 'UNTESTED' && (
                <div className='batch-advice danger' style={{ marginTop: 12 }}>
                  <strong>暂不能提交审批</strong>
                  <p>{conflictDecisionView.guidance}</p>
                </div>
              )}
            </SectionCard>
          )}
        </main>

        <aside className='batch-intake-aside'>
          <SectionCard title='案件概览'>
            <div className='batch-overview-score'>
              <Progress
                type='circle'
                percent={intakeCompleteness}
                size={92}
                strokeColor='#12a89d'
              />
              <div>
                <strong>信息完整度</strong>
                <span>{overviewHint}</span>
              </div>
            </div>
            {[
              `案件名称 ${form.title || '未填写'}`,
              `案件类型 ${form.caseType ? dbCaseType(form.caseType) : '未选择'}`,
              `业务领域 ${form.businessArea || '未选择'}`,
              '案件阶段 潜在受理',
              '统一编号 待生成',
              `创建人 ${getUserInfo()?.name || getUserInfo()?.username || '当前用户'}`,
            ].map((line) => (
              <p key={line}>{line}</p>
            ))}
            {runtime.intake && (
              <p>
                接案编号 <strong>{runtime.intake.intake_code}</strong>
              </p>
            )}
            {runtime.conflict && (
              <p>
                冲突处置 <strong>{conflictRiskText}</strong>
              </p>
            )}
          </SectionCard>
          <SectionCard title='团队指派（预览）'>
            {isAssistant ? (
              <p>负责律师将在接案草稿移交后由律师本人选择和确认。</p>
            ) : (
              <>
                <Select
                  aria-label='负责律师预览'
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
                  负责人：{selectedLawyer?.name || '未选择'}（
                  {selectedLawyer?.position || selectedLawyer?.seniority || '律师'}）
                </div>
              </>
            )}
          </SectionCard>
          <SectionCard title='接口响应'>
            {runtime.apiTimings.length === 0 ? (
              <p>尚未加载数据</p>
            ) : (
              runtime.apiTimings.map((item) => (
                <p key={`${item.label}-${item.at}`}>
                  {item.label} <strong>{item.duration}ms</strong> <span>{item.at}</span>
                </p>
              ))
            )}
          </SectionCard>
        </aside>
      </div>

      <div className='batch-bottom-bar'>
        <Button onClick={handleCancelIntake}>取消立案</Button>
        {isAssistant ? (
          <span>协作草稿：等待负责律师补充当事人信息并确认</span>
        ) : (
          <span>
            <SafetyCertificateOutlined /> 利益冲突检查状态：
            <strong className={conflictStatusClass}>{conflictRiskText}</strong>
          </span>
        )}
        {runtime.intake?.intake_code && (
          <span>
            接案草稿已创建：<strong>{runtime.intake.intake_code}</strong>
          </span>
        )}
        {!isAssistant && !conflictDecisionView.canSubmit && (
          <span className='danger-text'>{conflictDecisionView.guidance}</span>
        )}
        {!isAssistant && conflictDecisionView.showReviewAction && conflictTaskID && (
          <Button size='small' danger onClick={() => navigate(conflictReviewPath)}>
            进入本案冲突复核
          </Button>
        )}
        <Space>
          <Button onClick={handleCreateIntake}>保存草稿</Button>
          <Button onClick={handleSaveDraftAndExit}>保存并退出</Button>
          {!isAssistant && (
            <Button
              aria-label='保存并进行利益冲突检查'
              icon={<SafetyCertificateOutlined />}
              loading={checkingConflict}
              onClick={handleRunConflictCheck}
            >
              保存最新输入并检测
            </Button>
          )}
          {!isAssistant && (
            <Tooltip title={submitTooltip}>
              <span>
                <Button
                  type='primary'
                  loading={submitting}
                  disabled={submitDisabled}
                  onClick={submitApproval}
                >
                  提交审批并等待成案
                </Button>
              </span>
            </Tooltip>
          )}
        </Space>
      </div>
    </div>
  )
}

export function ConflictCheckResults() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const [commandCenter, setCommandCenter] = React.useState<CommandCenterPayload | null>(null)
  const [loading, setLoading] = React.useState(false)
  const [selectedConflict, setSelectedConflict] = React.useState<CommandCenterRiskItem | null>(null)
  const requestedRiskFilter = searchParams.get('risk') || '全部结果'
  const [riskFilter, setRiskFilter] = React.useState(
    ['全部结果', '高风险', '中风险', '低风险', '待人工复核', '提示'].includes(requestedRiskFilter)
      ? requestedRiskFilter
      : '全部结果',
  )
  const [creatingApproval, setCreatingApproval] = React.useState(false)
  const [latestReview, setLatestReview] = React.useState<Record<string, unknown>>({})
  const [reviewerAssignment, setReviewerAssignment] = React.useState<Record<string, unknown>>({})
  const [reviewerCandidates, setReviewerCandidates] = React.useState<Array<Record<string, any>>>([])
  const [selectedReviewerID, setSelectedReviewerID] = React.useState<number | undefined>()
  const [assigningReviewer, setAssigningReviewer] = React.useState(false)
  const [reviewDecision, setReviewDecision] = React.useState('')
  const [reviewNotes, setReviewNotes] = React.useState('')
  const [submittingReview, setSubmittingReview] = React.useState(false)
  const [latestWaiver, setLatestWaiver] = React.useState<Record<string, unknown>>({})
  const [waiverModalOpen, setWaiverModalOpen] = React.useState(false)
  const [waiverRationale, setWaiverRationale] = React.useState('')
  const [waiverConditions, setWaiverConditions] = React.useState(
    '建立信息隔离墙\n限制敏感资料访问\n定期合规复核',
  )
  const [waiverDurationDays, setWaiverDurationDays] = React.useState('180')
  const [submittingWaiver, setSubmittingWaiver] = React.useState(false)
  const [subjectRegistrations, setSubjectRegistrations] = React.useState<Array<Record<string, any>>>([])
  const [selectedSubjectRegistration, setSelectedSubjectRegistration] = React.useState<Record<string, any> | null>(null)
  const [subjectRegistrationDecision, setSubjectRegistrationDecision] = React.useState('CREATE_NEW')
  const [subjectRegistrationNotes, setSubjectRegistrationNotes] = React.useState('')
  const [subjectRegistrationSubmitting, setSubjectRegistrationSubmitting] = React.useState(false)
  const [registryEntityQuery, setRegistryEntityQuery] = React.useState('')
  const [registryEntityOptions, setRegistryEntityOptions] = React.useState<Array<Record<string, any>>>([])
  const [registryEntityLoading, setRegistryEntityLoading] = React.useState(false)
  const [selectedRegistryEntityID, setSelectedRegistryEntityID] = React.useState<number | undefined>()

  React.useEffect(() => {
    const controller = new AbortController()
    setLoading(true)
    fetchCommandCenter(controller.signal, true)
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
  const requestedTaskID = searchParams.get('task_id') || ''
  const riskFilterLevels: Record<string, string[]> = {
    高风险: ['HIGH', 'CRITICAL'],
    中风险: ['MEDIUM'],
    低风险: ['LOW'],
    待人工复核: ['REVIEW_REQUIRED'],
    提示: ['MINIMAL'],
  }
  const filteredRiskItems =
    riskFilter === '全部结果'
      ? riskItems
      : riskFilter === '待人工复核'
        ? riskItems.filter((item) => {
            const view = deriveConflictDecisionViewModel(item, {
              stale: textValue(item.status, '').toUpperCase() === 'STALE',
            })
            return view.decision === 'REVIEW_REQUIRED'
          })
        : riskItems.filter((item) =>
            riskFilterLevels[riskFilter]?.includes(textValue(item.risk_level, '').toUpperCase()),
          )

  React.useEffect(() => {
    if (!requestedTaskID || !commandCenter) return
    const requested = listOf(commandCenter.risk_queue).find(
      (item) => textValue(item.id, '') === requestedTaskID,
    )
    if (requested) {
      setSelectedConflict(requested)
    }
  }, [commandCenter, requestedTaskID])
  const contextCaseID = searchParams.get('case_id') || ''
  const contextCaseNumber = searchParams.get('case_number') || ''
  const contextCaseTitle = searchParams.get('case_title') || ''
  const contextConflict =
    riskItems.find((item) => requestedTaskID && textValue(item.id, '') === requestedTaskID) ||
    riskItems.find((item) =>
      conflictMatchesCaseContext(item, contextCaseID, contextCaseNumber, contextCaseTitle),
    )
  const requestedTaskUnavailable = Boolean(
    requestedTaskID && commandCenter && !contextConflict,
  )
  const criticalRiskCount = riskItems.filter(
    (item) => (item.risk_level || '').toUpperCase() === 'CRITICAL',
  ).length
  const highRiskCount = riskItems.filter((item) =>
    ['HIGH', 'CRITICAL'].includes((item.risk_level || '').toUpperCase()),
  ).length
  const mediumRiskCount = riskItems.filter(
    (item) => (item.risk_level || '').toUpperCase() === 'MEDIUM',
  ).length
  const lowRiskCount = riskItems.filter(
    (item) => (item.risk_level || '').toUpperCase() === 'LOW',
  ).length
  const reviewRequiredCount = riskItems.filter((item) => {
    const view = deriveConflictDecisionViewModel(item, {
      stale: textValue(item.status, '').toUpperCase() === 'STALE',
    })
    return view.decision === 'REVIEW_REQUIRED'
  }).length
  const minimalRiskCount = riskItems.filter(
    (item) => (item.risk_level || '').toUpperCase() === 'MINIMAL',
  ).length
  const queueRisk =
    criticalRiskCount > 0
      ? 'CRITICAL'
      : highRiskCount > 0
        ? 'HIGH'
        : mediumRiskCount > 0
          ? 'MEDIUM'
          : lowRiskCount > 0
            ? 'LOW'
            : reviewRequiredCount > 0
              ? 'REVIEW_REQUIRED'
              : minimalRiskCount > 0
                ? 'MINIMAL'
                : ''
  const selectedCheckResult = recordValue(selectedConflict?.check_result)
  const selectedDecision = recordValue(selectedCheckResult.decision)
  const selectedNormalizedSubjects = listOf<Record<string, unknown>>(
    selectedCheckResult.normalizedSubjects,
  )
  const selectedRiskAssessment = recordValue(selectedCheckResult.riskAssessment)
  const selectedMatchEvidence = recordValue(
    selectedRiskAssessment.matchEvidence || selectedCheckResult.matchEvidence,
  )
  const selectedStatistics = recordValue(selectedCheckResult.checkStatistics)
  const selectedSearchParameters = recordValue(selectedConflict?.search_parameters)
  const frozenConflictCases = listOf<Record<string, unknown>>(
    selectedCheckResult.conflictCases as Array<Record<string, unknown>> | undefined,
  )
  const selectedConflictCases =
    frozenConflictCases.length > 0
      ? frozenConflictCases
      : listOf<Record<string, unknown>>(selectedConflict?.conflict_cases)
  const selectedEvidence = selectedConflictCases.flatMap((item) =>
    listOf<Record<string, unknown>>(item.evidence as Array<Record<string, unknown>> | undefined),
  )
  const selectedStructuredPrimaryEvidence = recordValue(
    selectedRiskAssessment.primaryEvidence || selectedCheckResult.primaryEvidence,
  )
  const selectedDirectEvidence =
    listOf<Record<string, unknown>>(
      (selectedRiskAssessment.evidence ||
        selectedCheckResult.evidence ||
        selectedConflict?.evidence) as Array<Record<string, unknown>> | undefined,
    )[0] || {}
  const selectedPrimaryConflictCase = selectedConflictCases[0] || {}
  const selectedEvidenceRows =
    selectedEvidence.length > 0
      ? selectedEvidence
      : Object.keys(selectedMatchEvidence).length > 0
        ? [
            {
              ...selectedMatchEvidence,
              ruleCode: firstPresent(
                selectedMatchEvidence.ruleCode,
                selectedRiskAssessment.ruleCode,
                selectedPrimaryConflictCase.ruleCode,
                '-',
              ),
              matchType: firstPresent(
                selectedMatchEvidence.matchType,
                selectedRiskAssessment.matchType,
                selectedPrimaryConflictCase.matchType,
                '-',
              ),
              requestedParty: firstPresent(
                selectedMatchEvidence.requestedParty,
                selectedRiskAssessment.requestedParty,
                selectedConflict?.matched_subject,
                '-',
              ),
              sourceCaseName: firstPresent(
                selectedMatchEvidence.sourceCaseName,
                selectedPrimaryConflictCase.caseName,
                selectedConflict?.title,
                '-',
              ),
              summary: firstPresent(
                selectedMatchEvidence.summary,
                selectedRiskAssessment.riskReason,
                selectedConflict?.evidence_summary,
                '-',
              ),
            },
          ]
        : []
  const selectedCaseEvidence = selectedEvidenceRows[0] || {}
  const selectedPrimaryEvidence = primaryConflictEvidence(selectedConflict)
  const selectedHasMatchEvidence =
    selectedConflictCases.length > 0 ||
    selectedEvidenceRows.length > 0 ||
    Object.keys(selectedRiskAssessment.matchEvidence || {}).length > 0 ||
    Object.keys(selectedRiskAssessment.primaryEvidence || {}).length > 0 ||
    Object.keys(selectedCheckResult.matchEvidence || {}).length > 0 ||
    Object.keys(selectedCheckResult.primaryEvidence || {}).length > 0
  const selectedSearchSubject = textValue(
    firstPresent(
      selectedRiskAssessment.requestedParty,
      selectedRiskAssessment.queryName,
      selectedStructuredPrimaryEvidence.requestedParty,
      selectedStructuredPrimaryEvidence.queryName,
      selectedMatchEvidence.requestedParty,
      selectedMatchEvidence.queryName,
      selectedDirectEvidence.requestedParty,
      selectedDirectEvidence.queryName,
      selectedCaseEvidence.requestedParty,
      selectedCaseEvidence.queryName,
      selectedPrimaryConflictCase.requestedParty,
      selectedSearchParameters.requestedParty,
      selectedSearchParameters.query,
    ),
    '-',
  )
  const selectedRestricted =
    textValue(selectedConflict?.source_case || selectedConflict?.sourceCase, '') === '受限' ||
    textValue(selectedConflict?.evidence_summary || selectedConflict?.evidenceSummary, '').includes(
      '受隔离',
    )
  const selectedHasConflict =
    selectedConflict?.has_conflict ??
    (selectedConflict as Record<string, any> | null | undefined)?.hasConflict
  const selectedNoMatch =
    !selectedRestricted &&
    !selectedHasMatchEvidence &&
    (selectedHasConflict === false || selectedHasConflict === 0)
  const selectedHistoricalSubject = selectedRestricted
    ? '存在受限命中（详情受隔离保护）'
    : selectedHasMatchEvidence
      ? textValue(
          firstPresent(
            selectedRiskAssessment.matchedClientName,
            selectedRiskAssessment.historicalClientName,
            selectedStructuredPrimaryEvidence.matchedClientName,
            selectedStructuredPrimaryEvidence.historicalClientName,
            selectedMatchEvidence.matchedClientName,
            selectedMatchEvidence.historicalClientName,
            selectedMatchEvidence.candidateName,
            selectedDirectEvidence.matchedClientName,
            selectedDirectEvidence.historicalClientName,
            selectedCaseEvidence.matchedClientName,
            selectedCaseEvidence.historicalClientName,
            selectedCaseEvidence.historicalParty,
            selectedCaseEvidence.requestedParty,
            selectedPrimaryConflictCase.matched_subject,
            selectedSearchParameters.matchedClientName,
            selectedConflict?.matched_subject,
          ),
          '无命中主体',
        )
      : '未发现匹配记录'
  const selectedMatchType = textValue(
    firstPresent(
      selectedRiskAssessment.matchType,
      selectedStructuredPrimaryEvidence.matchType,
      selectedMatchEvidence.matchType,
      selectedDirectEvidence.matchType,
      selectedCaseEvidence.matchType,
      selectedPrimaryConflictCase.matchType,
      selectedSearchParameters.matchMode,
    ),
    '',
  )
  const selectedAlgorithm = textValue(
    firstPresent(
      selectedRiskAssessment.algorithm,
      selectedStructuredPrimaryEvidence.algorithm,
      selectedMatchEvidence.algorithm,
      selectedDirectEvidence.algorithm,
      selectedCaseEvidence.algorithm,
    ),
    '',
  )
  const selectedAlgorithmLabel = selectedAlgorithm
    ? conflictAlgorithmLabel(selectedAlgorithm)
    : ['EXACT', 'EXACT_NORMALIZED'].includes(selectedMatchType.toUpperCase())
      ? '规范化名称比对'
      : '未提供'
  const selectedSubjectRole = textValue(
    firstPresent(
      selectedRiskAssessment.subjectRole,
      selectedRiskAssessment.partyRole,
      selectedStructuredPrimaryEvidence.subjectRole,
      selectedStructuredPrimaryEvidence.partyRole,
      selectedMatchEvidence.subjectRole,
      selectedMatchEvidence.partyRole,
      selectedDirectEvidence.subjectRole,
      selectedDirectEvidence.partyRole,
      selectedCaseEvidence.subjectRole,
      selectedCaseEvidence.partyRole,
    ),
    '未提供',
  )
  const selectedRuleCode = textValue(
    firstPresent(
      selectedRiskAssessment.ruleCode,
      selectedRiskAssessment.ruleId,
      selectedStructuredPrimaryEvidence.ruleCode,
      selectedStructuredPrimaryEvidence.ruleId,
      selectedMatchEvidence.ruleCode,
      selectedMatchEvidence.ruleId,
      selectedDirectEvidence.ruleCode,
      selectedDirectEvidence.ruleId,
      selectedCaseEvidence.ruleCode,
      selectedCaseEvidence.ruleId,
      selectedPrimaryConflictCase.ruleCode,
    ),
    '',
  )
  const selectedSourceCase = textValue(
    firstPresent(
      selectedRiskAssessment.sourceCaseNumber,
      selectedStructuredPrimaryEvidence.sourceCaseNumber,
      selectedStructuredPrimaryEvidence.sourceCaseName,
      selectedMatchEvidence.sourceCaseNumber,
      selectedMatchEvidence.sourceCaseName,
      selectedDirectEvidence.sourceCaseNumber,
      selectedDirectEvidence.sourceCaseName,
      selectedCaseEvidence.sourceCaseNumber,
      selectedCaseEvidence.sourceCaseName,
      selectedPrimaryConflictCase.case_no,
      selectedPrimaryConflictCase.case_number,
      selectedPrimaryConflictCase.caseName,
    ),
    '未提供',
  )
  const selectedRiskScore = firstPresent(
    selectedRiskAssessment.riskScore,
    selectedStructuredPrimaryEvidence.riskScore,
    selectedMatchEvidence.riskScore,
    selectedDirectEvidence.riskScore,
    selectedCaseEvidence.riskScore,
    selectedPrimaryConflictCase.risk_score,
    selectedPrimaryConflictCase.riskScore,
  )
  const selectedRawMatchedType = textValue(selectedConflict?.matched_type, '').toUpperCase()
  const selectedConflictType = conflictTypeLabel(
    firstPresent(
      selectedRiskAssessment.conflictType,
      selectedStructuredPrimaryEvidence.conflictType,
      selectedMatchEvidence.conflictType,
      selectedDirectEvidence.conflictType,
      selectedCaseEvidence.conflictType,
      selectedPrimaryConflictCase.conflict_type,
      selectedPrimaryConflictCase.conflictType,
      selectedRuleCode,
      ['EXACT', 'EXACT_NORMALIZED', 'NAME_CANDIDATE'].includes(selectedRawMatchedType)
        ? undefined
        : selectedConflict?.matched_type,
    ),
  )
  const selectedAutomaticConclusion = firstPresent(
    selectedRiskAssessment.automaticConclusion,
    selectedStructuredPrimaryEvidence.automaticConclusion,
    selectedMatchEvidence.automaticConclusion,
    selectedDirectEvidence.automaticConclusion,
    selectedCaseEvidence.automaticConclusion,
    selectedSearchParameters.automaticConclusion,
  )
  const selectedResolvedRiskReason = textValue(
    firstPresent(
      selectedRiskAssessment.riskReason,
      selectedCheckResult.riskReason,
      selectedCheckResult.risk_reason,
      selectedStructuredPrimaryEvidence.summary,
      selectedMatchEvidence.summary,
      selectedDirectEvidence.summary,
      selectedCaseEvidence.summary,
      selectedConflict?.conflict_details,
      selectedConflict?.description,
    ),
    '暂无风险原因',
  )
  const selectedIsStale = Boolean(
    selectedCheckResult.stale ||
      selectedCheckResult.isStale ||
      selectedCheckResult.is_stale ||
      textValue(selectedConflict?.status, '').toUpperCase() === 'STALE' ||
      textValue(selectedDecision.status, '').toUpperCase() === 'STALE',
  )
  const selectedDecisionView = deriveConflictDecisionViewModel(selectedConflict, {
    stale: selectedIsStale,
    review: latestReview,
    waiver: latestWaiver,
  })
  const currentUserInfo = getUserInfo()
  const currentRoleCodes = [
    textValue(currentUserInfo?.role, ''),
    ...listOf<any>(currentUserInfo?.roles).map((role) =>
      typeof role === 'string' ? role : textValue(role.code || role.name, ''),
    ),
    ...listOf(getRoles()).map((role) => textValue(role.code, '')),
  ]
    .map((role) => role.toLowerCase())
    .filter(Boolean)
  const canReviewConflict = currentRoleCodes.some((role) =>
    [
      'director',
      'partner',
      'compliance',
      'risk',
      'risk_control',
      'management',
      'conflict_officer',
    ].includes(role),
  )
  const hasExistingReview = Boolean(latestReview.id)

  const loadSubjectRegistrations = React.useCallback(() => {
    if (!canReviewConflict) {
      setSubjectRegistrations([])
      return
    }
    apiRequest<Array<Record<string, any>>>('/conflict/subject-entity-registrations')
      .then((rows) => setSubjectRegistrations(Array.isArray(rows) ? rows : []))
      .catch(() => setSubjectRegistrations([]))
  }, [canReviewConflict])

  React.useEffect(() => {
    loadSubjectRegistrations()
  }, [loadSubjectRegistrations])

  React.useEffect(() => {
    if (
      !selectedSubjectRegistration ||
      subjectRegistrationDecision !== 'LINK_EXISTING' ||
      registryEntityQuery.trim().length < 2
    ) {
      setRegistryEntityOptions([])
      return
    }
    let mounted = true
    const timer = window.setTimeout(() => {
      setRegistryEntityLoading(true)
      apiRequest<Array<Record<string, any>>>(
        `/conflict-v2/entities/search?query=${encodeURIComponent(registryEntityQuery.trim())}`,
      )
        .then((rows) => mounted && setRegistryEntityOptions(Array.isArray(rows) ? rows : []))
        .catch(() => mounted && setRegistryEntityOptions([]))
        .finally(() => mounted && setRegistryEntityLoading(false))
    }, 250)
    return () => {
      mounted = false
      window.clearTimeout(timer)
    }
  }, [registryEntityQuery, selectedSubjectRegistration, subjectRegistrationDecision])

  React.useEffect(() => {
    setReviewDecision('')
    setReviewNotes('')
    setLatestReview(recordValue(selectedCheckResult.review))
    setReviewerAssignment({})
    setSelectedReviewerID(undefined)
    setLatestWaiver(recordValue(selectedCheckResult.waiver))
    if (!selectedConflict?.id) return
    apiRequest<any>(`/conflict/tasks/${selectedConflict.id}/review`)
      .then((response) => {
        const review = recordValue(response?.review || response)
        setLatestReview(
          Object.keys(review).length > 0 ? review : recordValue(selectedCheckResult.review),
        )
      })
      .catch(() => setLatestReview(recordValue(selectedCheckResult.review)))
    const selectedWaiver = recordValue(selectedCheckResult.waiver)
    const shouldLoadWaiver =
      Object.keys(selectedWaiver).length > 0 ||
      ['WAIVER_PENDING', 'WAIVED'].includes(selectedDecisionView.decision)
    if (shouldLoadWaiver) {
      apiRequest<any>(`/conflict/tasks/${selectedConflict.id}/waiver`)
        .then((response) => {
          const waiver = recordValue(response)
          setLatestWaiver(Object.keys(waiver).length > 0 ? waiver : selectedWaiver)
        })
        .catch(() => setLatestWaiver(selectedWaiver))
    } else {
      setLatestWaiver(selectedWaiver)
    }
    if (canReviewConflict) {
      apiRequest<any>(`/conflict/tasks/${selectedConflict.id}/reviewer-assignment`)
        .then((response) => {
          const assignment = recordValue(response?.assignment || response)
          setReviewerAssignment(assignment)
          if (assignment.reviewer_id || assignment.reviewerId) {
            setSelectedReviewerID(
              numberValue(assignment.reviewer_id || assignment.reviewerId, 0) || undefined,
            )
          }
        })
        .catch(() => setReviewerAssignment({}))
      apiRequest<any>('/conflict/reviewer-candidates')
        .then((response) => {
          const candidates = Array.isArray(response) ? response : listOf(response?.data)
          setReviewerCandidates(candidates as Array<Record<string, any>>)
        })
        .catch(() => setReviewerCandidates([]))
    }
  }, [selectedConflict?.id, canReviewConflict])

  const assignConflictReviewer = async () => {
    if (!selectedConflict?.id || !selectedReviewerID) {
      message.warning('请先选择独立复核人')
      return
    }
    setAssigningReviewer(true)
    try {
      const response = await apiRequest<any>(
        `/conflict/tasks/${selectedConflict.id}/reviewer-assignment`,
        {
          method: 'POST',
          body: JSON.stringify({
            reviewer_id: selectedReviewerID,
            recusal_declared: true,
            independence_reason: '已确认与申请律师及承办律师不存在直接管理关系，并完成回避声明。',
          }),
        },
      )
      setReviewerAssignment(recordValue(response?.assignment || response))
      message.success('已指定独立复核人；完成复核前不能形成无冲突结论')
    } catch (error) {
      message.error(error instanceof Error ? error.message : '指定独立复核人失败')
    } finally {
      setAssigningReviewer(false)
    }
  }

  const submitConflictReview = async () => {
    if (!selectedConflict?.id || !reviewDecision) {
      message.warning('请选择人工复核结论')
      return
    }
    if (!reviewNotes.trim()) {
      message.warning('请填写复核依据，确保结论可以审计')
      return
    }
    if (canReviewConflict && !reviewerAssignment.id) {
      message.warning('请先指定独立复核人并完成回避声明')
      return
    }
    setSubmittingReview(true)
    try {
      const response = await apiRequest<any>(`/conflict/tasks/${selectedConflict.id}/review`, {
        method: 'POST',
        body: JSON.stringify({ decision: reviewDecision, notes: reviewNotes.trim() }),
      })
      const review = recordValue(response?.review || response)
      setLatestReview(review)
      setSelectedConflict((current) =>
        current
          ? {
              ...current,
              check_result: { ...recordValue(current.check_result), review },
            }
          : current,
      )
      setReviewDecision('')
      setReviewNotes('')
      fetchCommandCenter(new AbortController().signal, true)
        .then(setCommandCenter)
        .catch(() => null)
      message.success('人工复核结论已记录')
    } catch (error) {
      message.error(error instanceof Error ? error.message : '提交人工复核失败')
    } finally {
      setSubmittingReview(false)
    }
  }

  const submitWaiverRequest = async () => {
    if (!selectedConflict?.id || !waiverRationale.trim()) {
      message.warning('请填写具体的豁免理由和风险控制依据')
      return
    }
    setSubmittingWaiver(true)
    try {
      const waiver = await apiRequest<any>(`/conflict/tasks/${selectedConflict.id}/waiver`, {
        method: 'POST',
        body: JSON.stringify({
          rationale: waiverRationale.trim(),
          waiver_type: 'INFORMED_CONSENT',
          waiver_category: 'CLIENT_CONSENT',
          proposed_conditions: waiverConditions
            .split('\n')
            .map((line) => line.trim())
            .filter(Boolean),
          duration_days: Math.max(1, numberValue(waiverDurationDays, 180)),
          review_priority: ['HIGH', 'CRITICAL'].includes(
            textValue(selectedConflict.risk_level, '').toUpperCase(),
          )
            ? 'HIGH'
            : 'MEDIUM',
        }),
      })
      setLatestWaiver(recordValue(waiver))
      setSelectedConflict((current) =>
        current
          ? {
              ...current,
              check_result: {
                ...recordValue(current.check_result),
                waiver,
                decision: {
                  ...recordValue(recordValue(current.check_result).decision),
                  status: 'WAIVER_PENDING',
                  recommendation: '豁免申请正在复核，批准前不得继续接案。',
                },
              },
            }
          : current,
      )
      setWaiverModalOpen(false)
      setWaiverRationale('')
      fetchCommandCenter(new AbortController().signal, true)
        .then(setCommandCenter)
        .catch(() => null)
      message.success(`豁免申请已提交：${textValue(waiver.application_number, waiver.id)}`)
    } catch (error) {
      message.error(error instanceof Error ? error.message : '提交豁免申请失败')
    } finally {
      setSubmittingWaiver(false)
    }
  }

  const exportConflictReport = () => {
    if (riskItems.length === 0) {
      message.warning('暂无可导出的冲突检测记录')
      return
    }
    const generatedAt = new Date()
    const generatedAtText = generatedAt.toLocaleString()
    const reportRows = riskItems
      .map((row) => {
        const evidence = primaryConflictEvidence(row)
        const evidenceSummary = conflictRecordEvidenceSummary(row)
        return `<tr>
        <td>${escapeHtml(textValue(row.title, '未关联案件'))}</td>
        <td>${escapeHtml(textValue(row.client_name, '未登记客户'))}</td>
        <td>${escapeHtml(conflictHitSubject(row))}</td>
        <td>${escapeHtml(riskLabel(row.risk_level))}</td>
        <td>${escapeHtml(conflictDispositionLabel(deriveConflictDecisionViewModel(row, { stale: textValue(row.status, '').toUpperCase() === 'STALE' }).decision))}</td>
        <td>${escapeHtml(evidenceSummary)}</td>
        <td>${escapeHtml('系统冲突检测记录')}</td>
      </tr>`
      })
      .join('')
    const reportHtml = `<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><title>${escapeHtml('利益冲突检测报告')}</title>
      <style>
        @page { size: A4 landscape; margin: 16mm; }
        body { color: #172033; font: 14px/1.55 -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }
        h1 { margin: 0 0 8px; font-size: 24px; } h2 { margin: 22px 0 8px; font-size: 16px; }
        .meta, .stats { display: grid; grid-template-columns: repeat(4, 1fr); gap: 8px; margin: 12px 0; }
        .meta div, .stats div { border: 1px solid #d8e1ee; padding: 8px; } .label { color: #64748b; display: block; font-size: 12px; }
        table { width: 100%; border-collapse: collapse; table-layout: fixed; } th, td { border: 1px solid #cbd5e1; padding: 7px; text-align: left; vertical-align: top; word-break: break-word; }
        th { background: #f1f5f9; } th:nth-child(1) { width: 16%; } th:nth-child(2) { width: 15%; } th:nth-child(3) { width: 15%; } th:nth-child(6) { width: 25%; }
        .note { color: #64748b; font-size: 12px; }
        @media print { .no-print { display: none; } }
      </style></head><body>
      <h1>${escapeHtml('利益冲突检测报告')}</h1>
      <div class="meta"><div><span class="label">生成时间</span>${escapeHtml(generatedAtText)}</div><div><span class="label">检测范围</span>${escapeHtml('当前利益冲突检测任务队列')}</div></div>
      <h2>检测统计</h2><div class="stats"><div><span class="label">总数</span>${riskItems.length}</div><div><span class="label">高风险</span>${highRiskCount}</div><div><span class="label">中风险</span>${mediumRiskCount}</div><div><span class="label">低风险</span>${lowRiskCount}</div></div>
      <h2>检测记录</h2><table><thead><tr><th>案件</th><th>客户</th><th>命中主体</th><th>风险</th><th>状态</th><th>证据摘要</th><th>来源</th></tr></thead><tbody>${reportRows}</tbody></table>
      <p class="note">本报告基于当前系统冲突检测记录生成，最终处置以风险规则与人工复核为准。</p>
      </body></html>`
    const printWindow = window.open('', '_blank')
    if (!printWindow) {
      message.error('无法打开打印窗口，请允许浏览器弹出窗口后重试')
      return
    }
    try {
      printWindow.opener = null
    } catch {
      // Some browsers expose opener as read-only; the report contains escaped text and remains safe to print.
    }
    printWindow.document.open()
    printWindow.document.write(reportHtml)
    printWindow.document.close()
    printWindow.focus()
    printWindow.print()
  }

  const createConflictApproval = async (item?: CommandCenterRiskItem | null) => {
    if (!item?.id) {
      message.warning('请选择一条冲突检测记录')
      return
    }
    if (
      (contextCaseID || contextCaseNumber) &&
      !conflictMatchesCaseContext(item, contextCaseID, contextCaseNumber, contextCaseTitle)
    ) {
      message.error('所选检测记录不属于当前案件，已阻止创建错误的冲突审核')
      return
    }

    setCreatingApproval(true)
    try {
      const approval = await apiRequest<any>(`/conflict/tasks/${item.id}/approval`, {
        method: 'POST',
        body: JSON.stringify({
          title: `冲突审查审批 - ${textValue(item.title, textValue(item.id))}`,
          content: `客户 ${textValue(item.client_name, '未登记客户')} 的利益冲突检测结果为 ${riskLabel(item.risk_level)}，请合规复核。`,
          priority: ['HIGH', 'CRITICAL'].includes(textValue(item.risk_level, '').toUpperCase())
            ? 'high'
            : 'medium',
          expected_subject_case_id: contextCaseID || undefined,
        }),
      })
      const approvalNumber = textValue(approval.request_number, approval.approval_id)
      message.success(
        approval.reused
          ? `已进入现有冲突审批：${approvalNumber}`
          : `已创建冲突审批：${approvalNumber}`,
      )
      setSelectedConflict(null)
      navigate(`/approval/${approval.approval_id}`)
    } catch (error) {
      message.error(error instanceof Error ? error.message : '创建冲突审批失败')
    } finally {
      setCreatingApproval(false)
    }
  }

  const openConflictApproval = (item?: CommandCenterRiskItem | null) => {
    if (!item?.approval_id) {
      message.warning('当前记录尚未生成可处理的冲突审批')
      return
    }
    setSelectedConflict(null)
    navigate(`/approval/${item.approval_id}`)
  }

  const openSubjectRegistrationReview = (registration: Record<string, any>) => {
    setSelectedSubjectRegistration(registration)
    setSubjectRegistrationDecision('CREATE_NEW')
    setSubjectRegistrationNotes('')
    setRegistryEntityQuery('')
    setRegistryEntityOptions([])
    setSelectedRegistryEntityID(undefined)
  }

  const submitSubjectRegistrationReview = async () => {
    if (!selectedSubjectRegistration) return
    if (subjectRegistrationNotes.trim().length < 10) {
      message.error('请填写至少 10 个字的身份核验或驳回依据')
      return
    }
    if (subjectRegistrationDecision === 'LINK_EXISTING' && !selectedRegistryEntityID) {
      message.error('请选择要合并的已有主体')
      return
    }
    setSubjectRegistrationSubmitting(true)
    try {
      await apiRequest<any>(
        `/cases/${selectedSubjectRegistration.case_id}/subject-revisions/${selectedSubjectRegistration.revision_id}/entity-registration-review`,
        {
          method: 'POST',
          body: JSON.stringify({
            decision: subjectRegistrationDecision,
            existing_entity_id:
              subjectRegistrationDecision === 'LINK_EXISTING'
                ? selectedRegistryEntityID
                : undefined,
            notes: subjectRegistrationNotes.trim(),
          }),
        },
      )
      setSelectedSubjectRegistration(null)
      loadSubjectRegistrations()
      message.success(
        subjectRegistrationDecision === 'REJECT'
          ? '主体登记已驳回，原主体版本继续生效'
          : '主体身份已确认，申请律师待办已生成，可继续运行冲突重检',
      )
    } catch (error) {
      message.error(error instanceof Error ? error.message : '处理主体登记失败')
    } finally {
      setSubjectRegistrationSubmitting(false)
    }
  }

  return (
    <div className='batch-page conflict-page'>
      <PageHeader
        eyebrow='利益冲突 / 冲突检测 / 检测清单'
        title='利益冲突检测清单'
        subtitle={`冲突任务队列，最近刷新：${loading ? '正在加载' : formatDateTime(commandCenter?.generated_at)}`}
        actions={
          <Button
            className='conflict-report-print-button'
            icon={<PrinterOutlined />}
            onClick={exportConflictReport}
          >
            打印/导出 PDF
          </Button>
        }
      />

      <div className='ng-content'>
        {canReviewConflict && (
          <SectionCard title={`新主体登记待确认（${subjectRegistrations.length}）`}>
            <DataTable>
              <table>
                <thead>
                  <tr>
                    <th>案件</th>
                    <th>候选主体</th>
                    <th>主体类型</th>
                    <th>身份标识</th>
                    <th>申请人</th>
                    <th>变更原因</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  {subjectRegistrations.map((registration) => (
                    <tr key={textValue(registration.revision_id)}>
                      <td>
                        {textValue(registration.case_number)}<br />
                        {textValue(registration.case_title)}
                      </td>
                      <td>{textValue(registration.candidate_name)}</td>
                      <td>{textValue(registration.entity_type)}</td>
                      <td>
                        {textValue(registration.identity_type)} · {textValue(registration.identity_hint)}
                      </td>
                      <td>{textValue(registration.requested_by_name, registration.requested_by)}</td>
                      <td>{textValue(registration.reason)}</td>
                      <td>
                        <Button type='primary' size='small' onClick={() => openSubjectRegistrationReview(registration)}>
                          核验主体身份
                        </Button>
                      </td>
                    </tr>
                  ))}
                  {subjectRegistrations.length === 0 && (
                    <tr><td colSpan={7}>当前没有等待确认的新主体登记</td></tr>
                  )}
                </tbody>
              </table>
            </DataTable>
          </SectionCard>
        )}
        {requestedTaskUnavailable && (
          <SectionCard title='指定冲突任务不可访问'>
            <div className='batch-advice'>
              <strong>未找到该任务，或当前账号无权查看。</strong>
              <p>
                系统不会披露无权访问的案件或历史客户信息。请从当前账号的检测任务清单进入，或联系独立冲突核查人确认分派范围。
              </p>
            </div>
          </SectionCard>
        )}
        {(contextCaseID || contextCaseNumber || contextCaseTitle) && (
          <SectionCard title='本案复核上下文'>
            <div className='batch-advice'>
              <strong>
                {contextCaseNumber || contextCaseID || '当前案件'} {contextCaseTitle}
              </strong>
              {contextConflict ? (
                <>
                  <p>已匹配到本案冲突检测记录。请点击查看结果后再创建冲突审核。</p>
                  <Button type='primary' onClick={() => setSelectedConflict(contextConflict)}>
                    查看本案检测结果
                  </Button>
                </>
              ) : (
                <p>
                  暂未在冲突任务队列中匹配到本案检测记录。请确认是否已从立案工作台运行利益冲突检查；如未检测，请返回补充立案信息并重新检测。
                </p>
              )}
              {!contextConflict && (
                <Space wrap>
                  <Button
                    onClick={() => navigate(`/case/${contextCaseID}`)}
                    disabled={!contextCaseID}
                  >
                    返回案件详情
                  </Button>
                  <Button
                    type='primary'
                    onClick={() =>
                      navigate(
                        `/case/create?mode=supplement&case_id=${encodeURIComponent(contextCaseID)}`,
                      )
                    }
                    disabled={!contextCaseID}
                  >
                    补充立案信息并检测
                  </Button>
                </Space>
              )}
            </div>
          </SectionCard>
        )}

        <div className='batch-conflict-top'>
          <SectionCard title='检测任务概览'>
            <div className='batch-info-grid compact'>
              {[
                loading ? '检测任务 正在加载' : `检测任务 ${riskItems.length} 条`,
                loading ? '高风险 正在加载' : `高风险 ${highRiskCount} 条`,
                loading ? '中风险 正在加载' : `中风险 ${mediumRiskCount} 条`,
                loading ? '低风险 正在加载' : `低风险 ${lowRiskCount} 条`,
                loading ? '待人工复核 正在加载' : `待人工复核 ${reviewRequiredCount} 条`,
                loading
                  ? '待处理审批 正在加载'
                  : `待处理审批 ${commandCenter?.summary?.pending_approvals ?? 0} 条`,
                `最近刷新 ${loading ? '正在加载' : formatDateTime(commandCenter?.generated_at)}`,
              ].map((line) => (
                <p key={line}>
                  <span>{line.split(' ')[0]}</span>
                  <strong>{line.substring(line.indexOf(' ') + 1)}</strong>
                </p>
              ))}
            </div>
          </SectionCard>
          <section className='batch-risk-banner'>
            <AlertOutlined />
            <div>
              <span>队列最高风险</span>
              <strong>{loading ? '加载中' : riskLabel(queueRisk)}</strong>
              <em>
                {loading
                  ? '正在读取数据库记录'
                  : reviewRequiredCount > 0
                    ? `${reviewRequiredCount} 项记录需要独立人工复核，不能据此确认无冲突`
                    : riskItems.length === 0
                      ? canReviewConflict
                        ? '全所核查队列当前没有待处理记录'
                        : '当前账号没有待处理任务；不代表全所无冲突'
                      : `当前队列有 ${highRiskCount} 项高风险/严重记录`}
              </em>
            </div>
            <div>
              <span>任务数量</span>
              <strong className='score'>{loading ? '—' : riskItems.length}</strong>
              <em>条</em>
            </div>
            <div className='batch-risk-counts'>
              <p>
                高风险 <strong>{loading ? '—' : highRiskCount}</strong>
              </p>
              <p>
                中风险 <strong>{loading ? '—' : mediumRiskCount}</strong>
              </p>
              <p>
                低风险 <strong>{loading ? '—' : lowRiskCount}</strong>
              </p>
              <p>
                待人工复核 <strong>{loading ? '—' : reviewRequiredCount}</strong>
              </p>
              <p>
                提示 <strong>{loading ? '—' : minimalRiskCount}</strong>
              </p>
            </div>
            <div>
              <span>检测来源</span>
              <p>系统冲突检测记录</p>
              <p>{loading ? '正在读取数据库记录' : '无记录时显示空状态'}</p>
              <p>结果来自本地数据库</p>
            </div>
          </section>
        </div>

        <div className='batch-conflict-layout'>
          <SectionCard
            title='检测任务清单'
            extra={
              <Space>
                {['全部结果', '高风险', '中风险', '低风险', '待人工复核', '提示'].map((tab) => (
                  <Button
                    key={tab}
                    type={riskFilter === tab ? 'primary' : 'default'}
                    aria-label={`筛选${tab}`}
                    onClick={() => {
                      setRiskFilter(tab)
                      setSelectedConflict(null)
                      const next = new URLSearchParams(searchParams)
                      if (tab === '全部结果') next.delete('risk')
                      else next.set('risk', tab)
                      setSearchParams(next, { replace: true })
                    }}
                  >
                    {tab}
                  </Button>
                ))}
              </Space>
            }
            className='span-2'
          >
            <DataTable>
              <table className='conflict-list-table'>
                <thead>
                  <tr>
                    <th>风险等级</th>
                    <th>处置状态</th>
                    <th>命中主体</th>
                    <th>命中类型</th>
                    <th>命中范围</th>
                    <th>置信度</th>
                    <th>穿透层级</th>
                    <th>证据概要</th>
                    <th>来源</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredRiskItems.map((row) => (
                    <tr
                      key={row.id || row.title}
                      className={
                        ['HIGH', 'CRITICAL'].includes((row.risk_level || '').toUpperCase())
                          ? 'danger-row'
                          : ''
                      }
                    >
                      <td>
                        <RiskTag text={riskLabel(row.risk_level)} />
                      </td>
                      <td>
                        {conflictDispositionLabel(
                          deriveConflictDecisionViewModel(row, {
                            stale: textValue(row.status, '').toUpperCase() === 'STALE',
                          }).decision,
                        )}
                      </td>
                      <td className='strong-cell'>{conflictHitSubject(row)}</td>
                      <td>{conflictRecordMatchType(row)}</td>
                      <td>{textValue(row.title, '未关联案件')}</td>
                      <td>{conflictConfidenceLabel(row)}</td>
                      <td>-</td>
                      <td className='conflict-evidence-cell'>
                        <Tooltip title={conflictRecordEvidenceSummary(row)}>
                          <span>{conflictRecordEvidenceSummary(row)}</span>
                        </Tooltip>
                      </td>
                      <td>系统冲突检测记录</td>
                      <td className='conflict-action-cell'>
                        <Button
                          size='small'
                          type='primary'
                          ghost
                          onClick={() => setSelectedConflict(row)}
                        >
                          查看详情
                        </Button>
                      </td>
                    </tr>
                  ))}
                  {filteredRiskItems.length === 0 && (
                    <tr>
                      <td colSpan={10}>
                        {loading
                          ? '正在加载数据库冲突检测记录'
                          : riskItems.length === 0
                            ? canReviewConflict
                              ? '全所冲突核查队列暂无检测记录'
                              : '当前账号暂无待处理冲突检测任务'
                            : `暂无${riskFilter}记录`}
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </DataTable>
          </SectionCard>

          <aside className='batch-conflict-side'>
            <SectionCard title='合规建议'>
              <div className='batch-advice danger'>
                <strong>
                  {conflictQueueScopeLabel(canReviewConflict)}：
                  {reviewRequiredCount > 0
                    ? `存在 ${reviewRequiredCount} 条待人工复核记录`
                    : highRiskCount > 0
                      ? '存在高风险或严重记录'
                      : canReviewConflict
                        ? '当前队列未发现高风险记录'
                        : '暂无待处理冲突检测任务'}
                </strong>
                <p>
                  该判断汇总{canReviewConflict ? '全所核查队列' : '当前账号可见范围'}内的{' '}
                  {riskItems.length} 条记录，不代表当前选中记录的处置结论。
                </p>
                {!canReviewConflict && (
                  <p>本页不代表全所已完成冲突检查；如需确认全所范围，请联系独立冲突核查人。</p>
                )}
                {reviewRequiredCount > 0 && (
                  <p>
                    “未发现高风险”不等于“已确认无冲突”；待人工复核记录在独立核查人形成结论前不得继续接案。
                  </p>
                )}
                {selectedConflict && (
                  <p>
                    <strong>当前记录建议：</strong>
                    {selectedDecisionView.headline}。{selectedDecisionView.guidance}
                  </p>
                )}
              </div>
            </SectionCard>
            <SectionCard title='检测范围'>
              <div className='batch-scope-grid'>
                <span>冲突记录 {riskItems.length}</span>
                <span>高风险 {highRiskCount}</span>
                <span>中风险 {mediumRiskCount}</span>
                <span>低风险 {lowRiskCount}</span>
                <span>待人工复核 {reviewRequiredCount}</span>
                <span>审批待办 {commandCenter?.summary?.pending_approvals ?? 0}</span>
                <span>案件 {commandCenter?.summary?.active_cases ?? 0}</span>
                <span>范围 {conflictQueueScopeLabel(canReviewConflict)}</span>
              </div>
            </SectionCard>
          </aside>
        </div>

        <Modal
          title='冲突检测详情'
          open={Boolean(selectedConflict)}
          onCancel={() => setSelectedConflict(null)}
          footer={[
            <Button key='close' onClick={() => setSelectedConflict(null)}>
              关闭
            </Button>,
            ...(selectedDecisionView.decision === 'BLOCKED' && selectedDecisionView.canRequestWaiver
              ? [
                  <Button key='waiver' type='primary' onClick={() => setWaiverModalOpen(true)}>
                    申请豁免评估
                  </Button>,
                ]
              : []),
            ...(selectedConflict?.approval_id
              ? [
                  <Button
                    key='review'
                    type='primary'
                    onClick={() => openConflictApproval(selectedConflict)}
                  >
                    {canReviewConflict ? '处理冲突审批' : '进入冲突审批'}
                  </Button>,
                ]
              : selectedDecisionView.needsHumanReview && !canReviewConflict
                ? [
                    <Button
                      key='review'
                      type='primary'
                      loading={creatingApproval}
                      disabled={!selectedConflict?.id}
                      onClick={() => createConflictApproval(selectedConflict)}
                    >
                      发起冲突审批
                    </Button>,
                  ]
                : []),
          ]}
          width={960}
          destroyOnHidden
        >
          {selectedConflict && (
            <div className='batch-conflict-detail'>
              <div className='batch-info-grid two'>
                {[
                  ['关联案件', textValue(selectedConflict.title, '-')],
                  ['案件类型', dbCaseType(textValue(selectedConflict.case_type, '-'))],
                  ['客户/委托人', textValue(selectedConflict.client_name, '-')],
                  ['命中历史主体', selectedHistoricalSubject],
                  ['主冲突类型', selectedConflictType],
                  [
                    '自动检索状态',
                    textValue(selectedConflict.status, '').toUpperCase() === 'COMPLETED'
                      ? '自动检索已完成；接案结论以复核状态为准'
                      : statusLabel(selectedConflict.status),
                  ],
                  ['风险等级', riskLabel(selectedConflict.risk_level)],
                  ['接案状态', selectedDecisionView.headline],
                  [
                    '命中数量',
                    selectedRestricted && !canReviewConflict
                      ? '受隔离保护'
                      : String(
                          selectedConflictCases.length ||
                            numberValue(selectedCheckResult.conflictCases?.length, 0),
                        ),
                  ],
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
                      <tr>
                        <td>数据来源</td>
                        <td>系统冲突检测记录</td>
                      </tr>
                      <tr>
                        <td>风险判断</td>
                        <td>{conflictDecisionStatusLabel(selectedDecisionView)}</td>
                      </tr>
                      <tr>
                        <td>当前记录建议</td>
                        <td>
                          {selectedDecisionView.headline}。{selectedDecisionView.guidance}
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </DataTable>
              </SectionCard>

              <SectionCard title='接案决策'>
                <div
                  className={`batch-advice ${['BLOCKED', 'STALE'].includes(selectedDecisionView.decision) ? 'danger' : ''}`}
                >
                  <strong>{selectedDecisionView.headline}</strong>
                  <p>{selectedDecisionView.guidance}</p>
                  {!selectedDecisionView.stale && (
                    <p>
                      {textValue(
                        selectedDecision.coverageNotice,
                        '检索范围受现有客户、案件和关联方数据完整度限制。',
                      )}
                    </p>
                  )}
                </div>
              </SectionCard>

              {selectedDecisionView.decision === 'WAIVED' && (
                <SectionCard title='批准条件与期限'>
                  <div className='batch-info-grid two'>
                    <article>
                      <span>批准条件</span>
                      <strong>
                        {listOf<string>(
                          (latestWaiver.approved_conditions ||
                            latestWaiver.proposed_conditions ||
                            latestWaiver.conditions) as string[] | undefined,
                        ).join('；') || '按书面批准文件执行'}
                      </strong>
                    </article>
                    <article>
                      <span>有效期限</span>
                      <strong>
                        {formatApiDate(
                          textValue(
                            latestWaiver.expiry_date || latestWaiver.requested_expiry_date,
                            '',
                          ),
                        )}
                      </strong>
                    </article>
                  </div>
                </SectionCard>
              )}

              {selectedDecisionView.decision === 'WAIVER_PENDING' && (
                <SectionCard title='独立复核状态'>
                  <div className='batch-advice'>
                    <strong>等待独立复核</strong>
                    <p>
                      申请编号：{textValue(latestWaiver.application_number || latestWaiver.id, '-')}
                      ；复核截止：{formatApiDate(textValue(latestWaiver.review_deadline, ''))}
                    </p>
                  </div>
                </SectionCard>
              )}

              <SectionCard title={`规范化检索主体（${selectedNormalizedSubjects.length}）`}>
                <DataTable>
                  <table>
                    <thead>
                      <tr>
                        <th>角色</th>
                        <th>原始名称</th>
                        <th>规范化名称</th>
                        <th>别名数量</th>
                        <th>身份标识数量</th>
                      </tr>
                    </thead>
                    <tbody>
                      {selectedNormalizedSubjects.map((subject, index) => (
                        <tr key={`${textValue(subject.normalizedName)}-${index}`}>
                          <td>{conflictSubjectRoleLabel(subject.role)}</td>
                          <td>{textValue(subject.originalName, '-')}</td>
                          <td>{textValue(subject.normalizedName, '-')}</td>
                          <td>{listOf(subject.aliases as unknown[] | undefined).length}</td>
                          <td>{Object.keys(recordValue(subject.identifiers)).length}</td>
                        </tr>
                      ))}
                      {selectedNormalizedSubjects.length === 0 && (
                        <tr>
                          <td colSpan={5}>历史记录没有规范化主体快照，请重新运行检测</td>
                        </tr>
                      )}
                    </tbody>
                  </table>
                </DataTable>
              </SectionCard>

              <SectionCard title='风险评估结果'>
                <DataTable>
                  <table>
                    <tbody>
                      <tr>
                        <td>总体风险</td>
                        <td>
                          {selectedDecisionView.coverageLimited
                            ? '范围受限，待人工复核'
                            : riskLabel(
                                textValue(
                                  selectedRiskAssessment.overallRisk || selectedConflict.risk_level,
                                  'LOW',
                                ),
                              )}
                        </td>
                      </tr>
                      <tr>
                        <td>风险评分</td>
                        <td>
                          {selectedDecisionView.needsHumanReview
                            ? '暂不评分（待独立人工复核）'
                            : selectedRiskScore === undefined
                              ? '未提供'
                              : selectedDecisionView.coverageLimited
                                ? `不作为无冲突结论（机器评分 ${formatRiskScore(selectedRiskScore)} / 100）`
                                : `${formatRiskScore(selectedRiskScore)} / 100`}
                        </td>
                      </tr>
                      <tr>
                        <td>风险原因</td>
                        <td>{selectedResolvedRiskReason}</td>
                      </tr>
                      <tr>
                        <td>检索主体</td>
                        <td>{selectedSearchSubject}</td>
                      </tr>
                      <tr>
                        <td>匹配主体</td>
                        <td>{selectedHistoricalSubject}</td>
                      </tr>
                      <tr>
                        <td>匹配方式</td>
                        <td>
                          {selectedNoMatch ? '无匹配' : conflictMatchTypeLabel(selectedMatchType)}
                        </td>
                      </tr>
                      <tr>
                        <td>比对方法</td>
                        <td>{selectedAlgorithmLabel}</td>
                      </tr>
                      {selectedAutomaticConclusion !== undefined && (
                        <tr>
                          <td>自动结论</td>
                          <td>
                            {selectedAutomaticConclusion === false
                              ? '否，必须人工核实主体身份'
                              : '是'}
                          </td>
                        </tr>
                      )}
                      <tr>
                        <td>主体角色</td>
                        <td>{conflictSubjectRoleLabel(selectedSubjectRole)}</td>
                      </tr>
                      <tr>
                        <td>规则编号</td>
                        <td>{conflictRuleLabel(selectedRuleCode)}</td>
                      </tr>
                      <tr>
                        <td>来源案件</td>
                        <td>
                          {selectedNoMatch
                            ? '不适用（未发现匹配记录）'
                            : canReviewConflict
                            ? selectedSourceCase
                            : '受限历史事项（请联系独立冲突核查人）'}
                        </td>
                      </tr>
                      <tr>
                        <td>需审批</td>
                        <td>{selectedDecisionView.requiresApproval ? '是' : '否'}</td>
                      </tr>
                      <tr>
                        <td>检查范围</td>
                        <td>
                          {conflictScopeLabel(
                            selectedDecision.coverageStatus || selectedConflict?.coverage_status,
                          )}
                        </td>
                      </tr>
                      <tr>
                        <td>统计</td>
                        <td>
                          检查案件{' '}
                          {numberValue(
                            selectedStatistics.totalCasesChecked,
                            selectedConflictCases.length,
                          )}{' '}
                          件，关联方 {numberValue(selectedStatistics.relatedPartiesChecked, 0)} 个
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </DataTable>
              </SectionCard>

              <DiagnosticDetails label='查看审计技术信息'>
                <p>检测任务编号：{textValue(selectedConflict.id, '-')}</p>
                <p>数据来源：系统冲突检测记录</p>
                <p>
                  检测耗时：
                  {numberValue(selectedConflict.duration || selectedCheckResult.duration, 0)}ms
                </p>
                {Boolean(latestReview.evidenceHash) && (
                  <p>证据指纹：{textValue(latestReview.evidenceHash, '-')}</p>
                )}
              </DiagnosticDetails>

              <SectionCard
                title={
                  selectedRestricted && !canReviewConflict
                    ? '命中案件明细（受隔离保护）'
                    : `命中案件明细（${selectedConflictCases.length}）`
                }
              >
                <DataTable>
                  <table>
                    <thead>
                      <tr>
                        <th>案件编号</th>
                        <th>案件名称</th>
                        <th>冲突类型</th>
                        <th>风险</th>
                        <th>状态</th>
                        <th>说明</th>
                      </tr>
                    </thead>
                    <tbody>
                      {selectedConflictCases.map((item, index) => (
                        <tr
                          key={`conflict-case-${textValue(
                            item.id || item.case_id || item.caseId,
                            'unknown',
                          )}-${index}`}
                        >
                          <td>
                            {canReviewConflict
                              ? textValue(
                                  item.case_no ||
                                    item.case_number ||
                                    item.caseNo ||
                                    item.caseNumber ||
                                    item.case_id ||
                                    item.caseId,
                                )
                              : '受限'}
                          </td>
                          <td>
                            {canReviewConflict
                              ? textValue(item.case_name || item.caseName)
                              : '受限历史事项'}
                          </td>
                          <td>
                            {textValue(item.conflict_type || item.conflictType, '待人工复核')}
                          </td>
                          <td>
                            <RiskTag
                              text={
                                canReviewConflict
                                  ? riskLabel(textValue(item.risk_level || item.riskLevel, 'LOW'))
                                  : '受限'
                              }
                            />
                          </td>
                          <td>
                            {canReviewConflict
                              ? statusLabel(textValue(item.case_status || item.caseStatus, '-'))
                              : '受限'}
                          </td>
                          <td>
                            {canReviewConflict
                              ? textValue(item.conflict_details || item.description, '-')
                              : '存在受限命中，请联系独立冲突核查人。'}
                          </td>
                        </tr>
                      ))}
                      {selectedConflictCases.length === 0 && (
                        <tr>
                          <td colSpan={6}>
                            {selectedRestricted && !canReviewConflict
                              ? '存在受限记录，具体明细仅对独立冲突核查人可见'
                              : '暂无命中案件明细'}
                          </td>
                        </tr>
                      )}
                    </tbody>
                  </table>
                </DataTable>
              </SectionCard>

              <SectionCard
                title={
                  selectedRestricted && !canReviewConflict
                    ? '证据链（受隔离保护）'
                    : `证据链（${selectedEvidenceRows.length}）`
                }
              >
                <DataTable>
                  <table>
                    <thead>
                      <tr>
                        <th>规则</th>
                        <th>匹配方式</th>
                        <th>检索主体</th>
                        <th>历史角色</th>
                        <th>来源案件</th>
                        <th>证据摘要</th>
                      </tr>
                    </thead>
                    <tbody>
                      {selectedEvidenceRows.map((evidence, index) => (
                        <tr key={`${textValue(evidence.ruleCode)}-${index}`}>
                          <td>{conflictRuleLabel(evidence.ruleCode)}</td>
                          <td>{conflictMatchTypeLabel(evidence.matchType)}</td>
                          <td>{textValue(evidence.requestedParty, '-')}</td>
                          <td>{conflictSubjectRoleLabel(evidence.historicalRole)}</td>
                          <td>
                            {canReviewConflict && !evidence.restricted
                              ? textValue(evidence.sourceCaseNumber || evidence.sourceCaseName, '-')
                              : '受限历史事项'}
                          </td>
                          <td>
                            {canReviewConflict
                              ? textValue(evidence.summary, '-')
                              : '受限命中，请联系独立冲突核查人。'}
                          </td>
                        </tr>
                      ))}
                      {selectedEvidenceRows.length === 0 && (
                        <tr>
                          <td colSpan={6}>
                            {selectedRestricted && !canReviewConflict
                              ? '存在受限证据，具体内容仅对独立冲突核查人可见'
                              : selectedNoMatch
                                ? '未发现匹配证据；当前仅因检索覆盖受限进入人工复核'
                              : '历史记录没有结构化证据，请重新运行检测'}
                          </td>
                        </tr>
                      )}
                    </tbody>
                  </table>
                </DataTable>
              </SectionCard>

              {selectedDecisionView.needsHumanReview && (
                <SectionCard title='人工复核记录'>
                  {latestReview.id ? (
                    <div className='batch-info-grid two'>
                      <article>
                        <span>复核结论</span>
                        <strong>{conflictDispositionLabel(latestReview.decision)}</strong>
                      </article>
                      <article>
                        <span>复核人</span>
                        <strong>
                          {textValue(latestReview.reviewerName || latestReview.reviewerId, '-')}
                        </strong>
                      </article>
                      <article>
                        <span>复核时间</span>
                        <strong>{formatApiDate(textValue(latestReview.createdAt, ''))}</strong>
                      </article>
                      <article>
                        <span>复核依据</span>
                        <strong>{textValue(latestReview.notes, '-')}</strong>
                      </article>
                    </div>
                  ) : (
                    <p>
                      {selectedNoMatch
                        ? '尚未完成人工复核。当前未发现匹配记录，但检索覆盖完整性尚未确认；独立核查确认前不得继续提交立案。'
                        : '尚未完成人工复核。候选命中在复核前不得作为“已确认冲突”，也不得继续提交立案。'}
                    </p>
                  )}
                  {canReviewConflict ? (
                    <Space direction='vertical' style={{ width: '100%', marginTop: 12 }}>
                      <div className='batch-advice'>
                        <strong>独立复核指定</strong>
                        {reviewerAssignment.id ? (
                          <p>
                            已指定复核人：
                            {textValue(
                              reviewerCandidates.find(
                                (candidate) =>
                                  numberValue(candidate.id, 0) ===
                                  numberValue(
                                    reviewerAssignment.reviewer_id || reviewerAssignment.reviewerId,
                                    0,
                                  ),
                              )?.name,
                              textValue(
                                reviewerAssignment.reviewer_id || reviewerAssignment.reviewerId,
                                '-',
                              ),
                            )}
                            。已完成回避声明，结论将写入审计记录。
                          </p>
                        ) : (
                          <>
                            <p>先指定与申请律师、承办律师不存在直接管理关系的业务复核人。</p>
                            <Space.Compact block>
                              <Select
                                aria-label='选择独立复核人'
                                value={selectedReviewerID}
                                onChange={setSelectedReviewerID}
                                placeholder='选择独立复核人'
                                options={reviewerCandidates.map((candidate) => ({
                                  value: numberValue(candidate.id, 0),
                                  label: `${textValue(candidate.name, candidate.username)} · ${textValue(candidate.role, '业务复核角色')}`,
                                }))}
                              />
                              <Button
                                type='primary'
                                loading={assigningReviewer}
                                disabled={!selectedReviewerID}
                                onClick={assignConflictReviewer}
                              >
                                指定并声明回避
                              </Button>
                            </Space.Compact>
                          </>
                        )}
                      </div>
                      <Select
                        value={reviewDecision || undefined}
                        placeholder='选择复核结论'
                        style={{ width: '100%' }}
                        disabled={hasExistingReview}
                        onChange={setReviewDecision}
                        options={[
                          { value: 'no_conflict', label: '无冲突' },
                          { value: 'confirmed_conflict', label: '确认冲突' },
                          { value: 'false_positive', label: '误报' },
                          { value: 'insufficient_information', label: '信息不足' },
                          { value: 'waiver_requested', label: '申请豁免' },
                        ]}
                      />
                      <Input.TextArea
                        value={reviewNotes}
                        onChange={(event) => setReviewNotes(event.target.value)}
                        rows={3}
                        disabled={hasExistingReview}
                        placeholder='填写核对对象、数据来源和判断依据'
                      />
                      <Button
                        type='primary'
                        loading={submittingReview}
                        disabled={hasExistingReview}
                        onClick={submitConflictReview}
                      >
                        {hasExistingReview ? '已完成复核' : '提交人工复核结论'}
                      </Button>
                      {hasExistingReview && (
                        <p>当前检测记录已有复核结论。如有新证据，请重新运行冲突检测后再复核。</p>
                      )}
                    </Space>
                  ) : (
                    <p>
                      {latestReview.id
                        ? '独立复核已完成，但当前结论仍不足以放行接案。请进入冲突审批查看后续处置。'
                        : '当前账号不能直接下人工复核结论。请点击下方“发起冲突审批”，由独立冲突核查人完成复核。'}
                    </p>
                  )}
                </SectionCard>
              )}

              {selectedDecisionView.decision === 'BLOCKED' && Boolean(latestWaiver.id) && (
                <SectionCard title='豁免记录'>
                  <div className='batch-info-grid two'>
                    <article>
                      <span>申请编号</span>
                      <strong>{textValue(latestWaiver.application_number, '-')}</strong>
                    </article>
                    <article>
                      <span>状态</span>
                      <strong>{statusLabel(textValue(latestWaiver.status, '-'))}</strong>
                    </article>
                    <article>
                      <span>申请理由</span>
                      <strong>{textValue(latestWaiver.rationale, '-')}</strong>
                    </article>
                    <article>
                      <span>到期时间</span>
                      <strong>
                        {formatApiDate(textValue(latestWaiver.requested_expiry_date, ''))}
                      </strong>
                    </article>
                  </div>
                </SectionCard>
              )}
            </div>
          )}
        </Modal>

        <Modal
          title='申请冲突豁免评估'
          open={waiverModalOpen}
          onCancel={() => setWaiverModalOpen(false)}
          onOk={submitWaiverRequest}
          okText='提交独立复核'
          confirmLoading={submittingWaiver}
          destroyOnHidden
        >
          <Space direction='vertical' style={{ width: '100%' }}>
            <p>
              豁免不会自动消除冲突。申请提交后将由独立的合规、主任或管理人员复核，批准前仍禁止接案。
            </p>
            <Input.TextArea
              rows={4}
              value={waiverRationale}
              onChange={(event) => setWaiverRationale(event.target.value)}
              placeholder='说明知情同意基础、为何仍可代理、替代方案及剩余风险'
            />
            <Input.TextArea
              rows={4}
              value={waiverConditions}
              onChange={(event) => setWaiverConditions(event.target.value)}
              placeholder='每行一项风险控制条件'
            />
            <Input
              value={waiverDurationDays}
              onChange={(event) => setWaiverDurationDays(event.target.value)}
              placeholder='豁免有效期（天）'
            />
          </Space>
        </Modal>
      </div>

      <Modal
        title='核验新主体登记'
        open={Boolean(selectedSubjectRegistration)}
        onCancel={() => !subjectRegistrationSubmitting && setSelectedSubjectRegistration(null)}
        onOk={submitSubjectRegistrationReview}
        okText='确认处理结果'
        cancelText='取消'
        confirmLoading={subjectRegistrationSubmitting}
        okButtonProps={{
          disabled:
            subjectRegistrationNotes.trim().length < 10 ||
            (subjectRegistrationDecision === 'LINK_EXISTING' && !selectedRegistryEntityID),
        }}
      >
        <div className='batch-advice'>
          <strong>{textValue(selectedSubjectRegistration?.candidate_name, '候选主体')}</strong>
          <p>
            {textValue(selectedSubjectRegistration?.identity_type)} · 标识尾号
            {textValue(selectedSubjectRegistration?.identity_hint)}。请依据原始身份材料核验，不得只凭名称判断。
          </p>
        </div>
        <div className='batch-field'>
          <label htmlFor='subject-registration-decision'>处理方式 *</label>
          <Select
            id='subject-registration-decision'
            value={subjectRegistrationDecision}
            onChange={(value) => {
              setSubjectRegistrationDecision(value)
              setSelectedRegistryEntityID(undefined)
            }}
            options={[
              { value: 'CREATE_NEW', label: '建立新的正式主体档案' },
              { value: 'LINK_EXISTING', label: '合并到已有主体档案' },
              { value: 'REJECT', label: '驳回登记申请' },
            ]}
          />
        </div>
        {subjectRegistrationDecision === 'LINK_EXISTING' && (
          <>
            <div className='batch-field'>
              <label htmlFor='registry-entity-query'>搜索全所主体库 *</label>
              <Input
                id='registry-entity-query'
                value={registryEntityQuery}
                onChange={(event) => setRegistryEntityQuery(event.target.value)}
                placeholder='输入法定名称或曾用名，至少两个字'
                suffix={registryEntityLoading ? <ClockCircleOutlined spin /> : <SearchOutlined />}
              />
            </div>
            <div className='batch-field'>
              <label htmlFor='registry-entity-select'>选择已有主体 *</label>
              <Select
                id='registry-entity-select'
                value={selectedRegistryEntityID}
                onChange={setSelectedRegistryEntityID}
                placeholder='选择身份材料对应的已有主体'
                options={registryEntityOptions.map((entity) => ({
                  value: numberValue(entity.id, 0),
                  label: `${textValue(entity.name)} · ${textValue(entity.entity_type)}`,
                }))}
                notFoundContent='未找到可合并主体'
              />
            </div>
            <p className='danger-text'>系统会再次比对加密身份摘要；名称相同但身份不一致时无法合并。</p>
          </>
        )}
        <div className='batch-field'>
          <label htmlFor='subject-registration-notes'>核验或驳回依据 *</label>
          <Input.TextArea
            id='subject-registration-notes'
            value={subjectRegistrationNotes}
            onChange={(event) => setSubjectRegistrationNotes(event.target.value)}
            rows={4}
            placeholder='例如：已核验营业执照原件，统一社会信用代码与主体库记录一致。'
          />
        </div>
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
  const [approvalConflictReview, setApprovalConflictReview] = React.useState<Record<string, any>>(
    {},
  )
  const [apiTimings, setApiTimings] = React.useState<
    Array<{ label: string; duration: number; at: string }>
  >([])
  const [deciding, setDeciding] = React.useState(false)
  const [decisionDialog, setDecisionDialog] = React.useState<
    'approve' | 'reject' | 'request_changes' | null
  >(null)
  const [decisionReason, setDecisionReason] = React.useState('')
  const [materialFilter, setMaterialFilter] = React.useState('全部材料')
  const [approvalLoadState, setApprovalLoadState] = React.useState<'loading' | 'ready' | 'error'>(
    'loading',
  )
  const [approvalLoadError, setApprovalLoadError] = React.useState('')

  const approvalViewerRoleCodes = [
    textValue(getUserInfo()?.role, ''),
    ...listOf<any>(getUserInfo()?.roles).map((role) =>
      typeof role === 'string' ? role : textValue(role.code || role.name, ''),
    ),
    ...listOf(getRoles()).map((role) => textValue(role.code, '')),
  ]
    .map((role) => role.toLowerCase())
    .filter(Boolean)
  const canViewConflictEvidence = approvalViewerRoleCodes.some((role) =>
    [
      'director',
      'partner',
      'compliance',
      'risk',
      'risk_control',
      'management',
      'conflict_officer',
    ].includes(role),
  )

  const recordTiming = (label: string, startedAt: number) => {
    setApiTimings((current) => [
      {
        label,
        duration: Math.round(performance.now() - startedAt),
        at: new Date().toLocaleTimeString(),
      },
      ...current.filter((item) => item.label !== label),
    ])
  }

  const loadApproval = React.useCallback(async () => {
    if (!id) {
      setApprovalLoadError('审批编号缺失，无法加载审批详情。')
      setApprovalLoadState('error')
      return
    }
    setApprovalLoadState('loading')
    setApprovalLoadError('')
    try {
      const approvalStartedAt = performance.now()
      const approvalData = await apiRequest<any>(`/approvals/${id}`)
      recordTiming('审批详情', approvalStartedAt)
      setApproval(approvalData)

      const approvalMetadata = recordValue(approvalData?.metadata)
      const approvalConflictCheckID = textValue(
        approvalData?.conflict_check_id ||
          approvalMetadata.conflict_check_id ||
          approvalMetadata.conflict_task_id ||
          recordValue(approvalMetadata.conflict_record).check_id ||
          recordValue(approvalMetadata.conflict_result).checkId,
        '',
      )
      setApprovalConflictReview({})
      if (approvalConflictCheckID) {
        try {
          const reviewData = await apiRequest<any>(
            `/conflict/tasks/${approvalConflictCheckID}/review`,
          )
          setApprovalConflictReview(recordValue(reviewData?.review || reviewData))
        } catch {
          // A lawyer may not read protected review details. The server-side
          // approval gate remains authoritative when the review is hidden.
          setApprovalConflictReview({})
        }
      }

      try {
        const snapshotStartedAt = performance.now()
        const snapshotData = await apiRequest<any>(`/approvals/${id}/snapshot`)
        recordTiming('审批快照', snapshotStartedAt)
        setSnapshot(recordValue(snapshotData.snapshot))
      } catch {
        setSnapshot({})
      }

      try {
        const statusStartedAt = performance.now()
        const statusData = await apiRequest<any>(`/integration/approvals/${id}/status`)
        recordTiming('集成状态', statusStartedAt)
        setIntegrationStatus(statusData)
      } catch {
        setIntegrationStatus(null)
      }
      setApprovalLoadState('ready')
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '加载审批详情失败'
      setApproval(null)
      setSnapshot(null)
      setIntegrationStatus(null)
      setApprovalLoadError(errorMessage)
      setApprovalLoadState('error')
      message.error(errorMessage)
    }
  }, [id])

  React.useEffect(() => {
    loadApproval()
  }, [loadApproval])

  if (approvalLoadState === 'loading') {
    return (
      <div className='batch-page approval-page'>
        <PageHeader
          eyebrow='审批中心 / 我的审批 / 审批详情'
          title='正在加载审批详情'
          subtitle={`审批编号：${id || '未提供'}`}
        />
        <SectionCard title='加载中'>
          <p>正在读取审批、冲突证据和成案状态，请稍候。</p>
        </SectionCard>
      </div>
    )
  }

  if (approvalLoadState === 'error' || !approval) {
    return (
      <div className='batch-page approval-page'>
        <PageHeader
          eyebrow='审批中心 / 我的审批 / 审批详情'
          title='审批详情加载失败'
          subtitle={`审批编号：${id || '未提供'}`}
          actions={<Button onClick={loadApproval}>重新加载</Button>}
        />
        <SectionCard title='无法展示审批数据'>
          <strong>{approvalLoadError || '未取得审批详情'}</strong>
          <p>
            为避免把占位内容误认为真实审批结论，本页不会在接口失败时展示风险、申请人或成案状态。
          </p>
          <Button type='primary' onClick={() => navigate('/approval')}>
            返回审批中心
          </Button>
        </SectionCard>
      </div>
    )
  }

  const decideApproval = async (
    decision: 'approve' | 'reject' | 'request_changes',
    reason: string,
    comments: string,
  ): Promise<boolean> => {
    if (!id) {
      return false
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
      recordTiming(
        decision === 'approve' ? '审批通过并成案' : decision === 'reject' ? '审批拒绝' : '退回修改',
        startedAt,
      )
      await loadApproval()
      const statusData = await apiRequest<any>(`/integration/approvals/${id}/status`)
      if (decision === 'approve') {
        message.success(
          statusData?.case_creation?.case_id
            ? `已成案：${statusData.case_creation.case_number}`
            : '审批已通过',
        )
      } else if (decision === 'reject') {
        message.success('审批已拒绝，未创建正式案件')
      } else {
        message.success('已退回修改，未创建正式案件')
      }
      return true
    } catch (error) {
      message.error(error instanceof Error ? error.message : '审批处理失败')
      return false
    } finally {
      setDeciding(false)
    }
  }

  const openDecisionDialog = (decision: 'approve' | 'reject' | 'request_changes') => {
    setDecisionDialog(decision)
    setDecisionReason('')
  }

  const closeDecisionDialog = () => {
    if (deciding) return
    setDecisionDialog(null)
    setDecisionReason('')
  }

  const submitDecision = async () => {
    if (!decisionDialog) return
    const reason = decisionReason.trim()
    if (reason.length < 10) {
      message.warning('请填写不少于 10 个字的处理依据，便于申请人修改和后续审计。')
      return
    }
    const comments =
      decisionDialog === 'approve'
        ? '确认通过审批并触发成案。'
        : decisionDialog === 'reject'
          ? '确认拒绝，本次申请不会创建正式案件。'
          : '确认退回申请人补充材料，当前不会创建正式案件。'
    if (await decideApproval(decisionDialog, reason, comments)) {
      closeDecisionDialog()
    }
  }

  const decisionDialogCopy = {
    approve: {
      title: '确认同意并成案',
      notice: '通过后系统将创建正式案件。请确认冲突复核结论、材料和承办条件均已满足。',
      placeholder: '填写同意依据、已核验材料和必要的承办条件',
      okText: '确认同意并成案',
    },
    reject: {
      title: '确认拒绝申请',
      notice: '拒绝后本次申请终止且不会成案。请写明事实依据，避免申请人无法理解或错误重提。',
      placeholder: '填写拒绝所依据的冲突事实、规则或接案条件',
      okText: '确认拒绝',
    },
    request_changes: {
      title: '确认退回修改',
      notice: '退回后申请人可补充材料并重新提交。请明确列出缺失信息和复核要求。',
      placeholder: '填写需要补充的主体身份材料、冲突说明或其他事项',
      okText: '确认退回修改',
    },
  }

  const approvalMetadata = recordValue(approval?.metadata)
  const snapshotMetadata = recordValue(snapshot?.metadata)
  const metadata = { ...approvalMetadata, ...snapshotMetadata }
  const snapshotRoot = recordValue(snapshot)
  const approvalConflictResult = recordValue(approval?.conflict_result)
  const conflictResult =
    Object.keys(approvalConflictResult).length > 0
      ? approvalConflictResult
      : metadata.conflict_result || snapshot?.conflict_result
  const conflictRecord = recordValue(
    metadata.conflict_record || snapshot?.conflict_record || conflictResult?.record,
  )
  const approvalConflictCases = listOf<any>(
    metadata.conflict_cases || snapshot?.conflict_cases || conflictResult?.conflictCases,
  )
  const snapshotClient = recordValue(snapshotRoot.client)
  const snapshotMetadataClient = recordValue(snapshotMetadata.client)
  const approvalMetadataClient = recordValue(approvalMetadata.client)
  const snapshotClientName = textValue(
    snapshotRoot.client_name ||
      snapshotMetadata.client_name ||
      snapshotClient.name ||
      snapshotMetadataClient.name ||
      approvalMetadata.client_name ||
      approvalMetadataClient.name ||
      conflictRecord.client_name,
    '',
  )
  const newOpposingParties = listOf<any>(snapshotRoot.opposing_parties as any[] | undefined)
  const metadataOpposingParties = listOf<any>(
    snapshotMetadata.opposing_parties as any[] | undefined,
  )
  const legacyParties = listOf<any>(
    (snapshotRoot.parties || snapshotMetadata.parties || approvalMetadata.parties) as
      | any[]
      | undefined,
  )
  const snapshotOpposingPartyNames = (
    newOpposingParties.length > 0
      ? newOpposingParties
      : metadataOpposingParties.length > 0
        ? metadataOpposingParties
        : legacyParties.filter(
            (party) => textValue(recordValue(party).role, '').toLowerCase() === 'opposing_party',
          )
  )
    .map((party) =>
      typeof party === 'string'
        ? party
        : textValue(
            recordValue(party).name ||
              recordValue(party).originalName ||
              recordValue(party).normalizedName,
            '',
          ),
    )
    .filter(Boolean)
  const rawSnapshotEvidence =
    snapshotRoot.evidence || snapshotMetadata.evidence || approvalMetadata.evidence
  const snapshotEvidence = Array.isArray(rawSnapshotEvidence)
    ? listOf<Record<string, unknown>>(rawSnapshotEvidence)
    : Object.keys(recordValue(rawSnapshotEvidence)).length > 0
      ? [recordValue(rawSnapshotEvidence)]
      : []
  const primarySnapshotEvidence = snapshotEvidence[0] || primaryConflictEvidence(conflictResult)
  const approvalHasRestrictedHit =
    textValue(conflictRecord.source_case || conflictRecord.sourceCase, '') === '受限' ||
    textValue(conflictRecord.evidence_summary || conflictRecord.evidenceSummary, '').includes(
      '受隔离',
    ) ||
    approvalConflictCases.some(
      (item) =>
        textValue(item.source_case || item.sourceCase, '') === '受限' ||
        textValue(item.description || item.conflict_details, '').includes('受隔离'),
    )
  const conflictDecision = recordValue(
    conflictResult?.decision ||
      recordValue(conflictRecord.check_result).decision ||
      snapshotRoot.decision ||
      snapshotMetadata.decision,
  )
  const approvalNoMatch =
    !approvalHasRestrictedHit &&
    approvalConflictCases.length === 0 &&
    snapshotEvidence.length === 0 &&
    numberValue(conflictDecision.evidenceCount, 0) === 0 &&
    numberValue(conflictDecision.restrictedCount, 0) === 0
  const requestedEvidenceSubject = textValue(
    primarySnapshotEvidence.requestedParty || primarySnapshotEvidence.requested_party,
    '',
  )
  const primarySnapshotSubject = textValue(
    approvalNoMatch
      ? snapshotOpposingPartyNames[0]
      : approvalHasRestrictedHit
      ? '存在受限命中（详情受隔离保护）'
      : primarySnapshotEvidence.matchedSubject ||
          primarySnapshotEvidence.matched_subject ||
          primarySnapshotEvidence.historicalSubject ||
          primarySnapshotEvidence.historical_subject ||
          primarySnapshotEvidence.matchedClientName ||
          primarySnapshotEvidence.historicalClientName ||
          (requestedEvidenceSubject && requestedEvidenceSubject !== snapshotClientName
            ? requestedEvidenceSubject
            : undefined) ||
          recordValue(snapshotRoot.decision || snapshotMetadata.decision).primarySubject ||
          conflictRecord.matched_subject ||
          snapshotOpposingPartyNames[0],
    '待核实主体',
  )
  const rawSnapshotRuleCode = textValue(
    primarySnapshotEvidence.ruleCode ||
      primarySnapshotEvidence.rule_code ||
      primarySnapshotEvidence.ruleId,
    '',
  )
  const snapshotEvidenceSource = [
    textValue(
      primarySnapshotEvidence.sourceCaseNumber ||
        primarySnapshotEvidence.source_case_number ||
        primarySnapshotEvidence.sourceCaseName ||
        primarySnapshotEvidence.source_case_name,
      '',
    ),
    rawSnapshotRuleCode ? conflictRuleLabel(rawSnapshotRuleCode) : '',
    textValue(primarySnapshotEvidence.summary || primarySnapshotEvidence.description, ''),
  ]
    .filter(Boolean)
    .join(' · ')
  const conflictDecisionStatus = textValue(conflictDecision.status, '').toUpperCase()
  const currentConflictReview =
    Object.keys(approvalConflictReview).length > 0
      ? approvalConflictReview
      : recordValue(recordValue(conflictRecord.check_result).review)
  const conflictReviewDecision = textValue(currentConflictReview.decision, '').toLowerCase()
  const conflictCoverageStatus = textValue(
    conflictDecision.coverageStatus ||
      conflictDecision.coverage_status ||
      conflictResult?.coverage_status ||
      conflictRecord.coverage_status ||
      metadata.conflict_coverage_status,
    '',
  ).toUpperCase()
  const conflictNeedsHumanReview =
    (conflictDecisionStatus === 'REVIEW_REQUIRED' &&
      !['no_conflict', 'false_positive'].includes(conflictReviewDecision)) ||
    ['insufficient_information', 'waiver_requested'].includes(conflictReviewDecision) ||
    (conflictCoverageStatus !== '' && conflictCoverageStatus !== 'COMPLETE')
  const conflictApprovalBlocked =
    conflictNeedsHumanReview ||
    ['confirmed_conflict', 'insufficient_information', 'waiver_requested'].includes(
      conflictReviewDecision,
    )
  const conflictRisk = textValue(
    conflictNeedsHumanReview
      ? 'REVIEW_REQUIRED'
      : conflictResult?.riskAssessment?.overallRisk || conflictRecord.risk_level,
    approval?.type === 'conflict_approval' ? 'REVIEW_REQUIRED' : '',
  )
  const conflictCheckID = textValue(
    conflictResult?.checkId || conflictRecord.check_id || metadata.conflict_task_id,
    '已随审批快照冻结',
  )
  const expectedConflictCheckID = textValue(
    approval?.conflict_check_id || approvalMetadata.conflict_task_id,
    '',
  )
  const evidenceConflictCheckIDs = [
    textValue(approvalConflictResult.checkId, ''),
    textValue(snapshotMetadata.conflict_task_id || snapshot?.conflict_task_id, ''),
    textValue(recordValue(snapshot?.conflict_result).checkId, ''),
  ].filter(Boolean)
  const hasConflictEvidenceMismatch = Boolean(
    expectedConflictCheckID &&
      evidenceConflictCheckIDs.some((checkID) => checkID !== expectedConflictCheckID),
  )
  const caseCreation = integrationStatus?.case_creation
  const canOpenRelatedCase = Boolean(caseCreation?.case_id && caseCreation?.accessible !== false)
  const caseCreationLabel = caseCreation?.case_id
    ? `已成案 ${caseCreation.case_number || ''}`.trim()
    : approval?.status === 'approved'
      ? '审批已通过，待生成正式案件'
      : '尚未成案'
  const applicationCaseTitle = textValue(
    metadata.case_creation_config?.title ||
      snapshot?.case_creation_config?.title ||
      conflictRecord.case_name ||
      approval?.title,
    '未命名案件',
  )
  const approvalMaterialRows = listOf<any>(metadata.materials || snapshot?.materials)
  const filteredApprovalMaterials = approvalMaterialRows.filter((row) => {
    if (materialFilter === '全部材料') return true
    const materialType = textValue(row.material_type || row.type, '').toLowerCase()
    if (materialFilter === '冲突报告') return materialType.includes('conflict')
    if (materialFilter === '申请文档')
      return ['application', 'contract'].some((item) => materialType.includes(item))
    if (materialFilter === '证明材料')
      return ['evidence', 'identity', 'proof'].some((item) => materialType.includes(item))
    return !['conflict', 'application', 'contract', 'evidence', 'identity', 'proof'].some((item) =>
      materialType.includes(item),
    )
  })
  const approvalCommentRows = listOf<any>(approval?.records)
  const approvalTraceRows =
    approvalCommentRows.length > 0
      ? approvalCommentRows
      : [
          {
            id: 'submitted',
            approver_name:
              approval?.applicant_name || snapshot?.applicant?.submitted_name || '申请人',
            approver_role: '提交人',
            decision: approval?.status || 'submitted',
            decision_comments: approval?.content || `已提交立案审批：${applicationCaseTitle}`,
            created_at: approval?.created_at,
          },
        ]
  const approvalHasPreviousTerminalRecords =
    ['submitted', 'under_review', 'pending', 'resubmitted'].includes(
      textValue(approval?.status, '').toLowerCase(),
    ) &&
    approvalTraceRows.some((item) =>
      ['approve', 'approved', 'reject', 'rejected', 'request_changes'].includes(
        textValue(item.decision, '').toLowerCase(),
      ),
    )
  const relatedInfoRows = [
    `关联客户 ${snapshotClientName || '未记录'}`,
    `关联案件 ${applicationCaseTitle}`,
    `关联流程 ${approvalStageLabel(approval?.workflow_type || 'CONFLICT_APPROVAL')}`,
  ]
  const approvalAccess = normalizeApprovalAccess(approval)
  const canApproveApproval =
    approvalAccess.canApprove && !hasConflictEvidenceMismatch && !conflictApprovalBlocked
  const canOtherDecision =
    !hasConflictEvidenceMismatch && (approvalAccess.canReject || approvalAccess.canReturn)
  const canDecisionActions = canApproveApproval || canOtherDecision

  return (
    <div className='batch-page approval-page'>
      <PageHeader
        eyebrow='审批中心 / 我的审批 / 审批详情'
        title={approval?.title || '新建案件审批'}
        subtitle={`审批编号：${approval?.request_number || id || '加载中'} 状态：${statusLabel(approval?.status || '加载中')} 当前审批人：${approvalAccess.label}`}
        actions={
          <>
            <Badge
              count={caseCreationLabel}
              color={caseCreation?.case_id ? '#12a89d' : '#f59f2f'}
            />
            <Button icon={<PrinterOutlined />} onClick={() => window.print()}>
              打印
            </Button>
            <Tooltip
              title={
                canOpenRelatedCase
                  ? '跳转到关联案件详情'
                  : caseCreation?.case_id
                    ? '关联案件权限尚未同步，请联系管理员处理'
                    : '暂无关联案件，审批通过后生成案件'
              }
            >
              <Button
                type='primary'
                disabled={!canOpenRelatedCase}
                onClick={() => canOpenRelatedCase && navigate(`/case/${caseCreation.case_id}`)}
              >
                查看关联案件
              </Button>
            </Tooltip>
          </>
        }
      />

      {hasConflictEvidenceMismatch && (
        <SectionCard title='审批证据不一致'>
          <strong>审批关联的检测编号与冻结快照不一致，系统已阻止审批决定。</strong>
          <p>请联系管理员执行数据一致性修复后重新打开本审批，避免依据错误的冲突记录作出决定。</p>
        </SectionCard>
      )}

      <div className='batch-approval-layout'>
        <SectionCard title='审批流程'>
          <div className='batch-approval-steps'>
            {[
              [
                '1. 申请提交',
                approval?.applicant_name || '申请人',
                approval?.submission_date ? '已提交' : '草稿',
                approval?.content || '已提交审批申请',
              ],
              [
                '2. 当前审批',
                approvalAccess.label,
                statusLabel(approval?.status),
                approvalAccess.readonlyReason || '等待当前审批人处理',
              ],
              [
                '3. 成案与归档',
                '系统自动处理',
                caseCreation?.case_id ? '已完成' : '待处理',
                caseCreation?.case_number || '审批通过后生成正式案件',
              ],
            ].map((step, index) => (
              <article
                key={step[0]}
                className={
                  index === 0 || (index === 2 && caseCreation?.case_id)
                    ? 'done'
                    : index === 1
                      ? 'active'
                      : ''
                }
              >
                <span>
                  {index === 0 || (index === 2 && caseCreation?.case_id) ? (
                    <CheckCircleOutlined />
                  ) : (
                    index + 1
                  )}
                </span>
                <div>
                  <strong>{step[0]}</strong>
                  <p>{step[1]}</p>
                  <em>{step[3]}</em>
                </div>
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
                  `申请类型 ${approval?.type === 'conflict_approval' ? '冲突审核' : '立案审批'}`,
                  `申请人 ${approval?.applicant_name || snapshot?.applicant?.submitted_name || '当前用户'}`,
                  `申请部门 ${approval?.department_name || snapshot?.applicant?.department_name || '公司业务部'}`,
                  `案件名称 ${applicationCaseTitle}`,
                  `关联客户 ${snapshotClientName || '未记录'}`,
                  `对方当事人 ${snapshotOpposingPartyNames.join('、') || '未记录'}`,
                  `案件类型 ${dbCaseType(
                    textValue(
                      snapshot?.case_creation_config?.case_type ||
                        metadata.case_creation_config?.case_type ||
                        conflictRecord.case_type ||
                        metadata.conflict_record?.case_type,
                      '未记录',
                    ),
                  )}`,
                ].map((line) => (
                  <p key={line}>
                    <span>{line.split(' ')[0]}</span>
                    <strong>{line.substring(line.indexOf(' ') + 1)}</strong>
                  </p>
                ))}
              </div>
            </SectionCard>

            <SectionCard title='冲突检测摘要'>
              <div className='batch-approval-risk'>
                <AlertOutlined />
                <div>
                  <strong>总体风险等级：{riskLabel(conflictRisk)}</strong>
                  <p>
                    {conflictNeedsHumanReview
                      ? '当前结果不能作为无冲突结论，须由独立冲突核查人完成人工复核。'
                      : `风险评分：${formatRiskScore(conflictResult?.riskAssessment?.riskScore)} / 100`}
                    {' · '}检测记录：证据已冻结
                  </p>
                </div>
              </div>
              <div className='batch-hit-list'>
                <p>
                  <RiskTag text={riskLabel(conflictRisk)} />
                  客户：{snapshotClientName || '未记录'}
                </p>
                <p>
                  <RiskTag text={approvalNoMatch ? '检索对方' : '历史主体'} />
                  {primarySnapshotSubject}
                </p>
                <p>
                  <RiskTag text='证据来源' />
                  {approvalNoMatch
                    ? '未发现匹配证据（检索覆盖受限）'
                    : snapshotEvidenceSource || '未记录'}
                </p>
                <p>
                  <RiskTag text='自动检索' />
                  状态：{statusLabel(textValue(conflictRecord.status, approval?.status))}
                </p>
                <p>
                  <RiskTag text={`${approvalConflictCases.length} 条命中`} />
                  明细：
                  {approvalConflictCases
                    .slice(0, 2)
                    .map((item: any) => textValue(item.case_name || item.caseName, ''))
                    .filter(Boolean)
                    .join('、') || '审批快照暂无命中明细'}
                </p>
                <p>
                  <RiskTag text='数据来源' />
                  系统冲突检测记录与审批快照
                </p>
              </div>
              <DiagnosticDetails label='查看审批审计信息'>
                <p>检测记录编号：{conflictCheckID}</p>
                {apiTimings.map((item) => (
                  <p key={`${item.label}-${item.at}`}>
                    {item.label}：{item.duration}ms，读取时间 {item.at}
                  </p>
                ))}
              </DiagnosticDetails>
              <Button type='link' onClick={() => navigate('/conflict')}>
                返回利益冲突检测台
              </Button>
              {conflictApprovalBlocked && (
                <p className='batch-approval-readonly'>
                  {canDecisionActions
                    ? '当前不能同意并成案：冲突复核尚未形成可成案结论。请核对补充要求，或选择“退回修改”。'
                    : '当前不能成案：请等待独立冲突核查人处理；如申请被退回，再按意见补充主体身份材料并重新提交。'}
                </p>
              )}
            </SectionCard>

            <SectionCard title={`冲突命中明细（${approvalConflictCases.length}）`}>
              <DataTable>
                <table>
                  <thead>
                    <tr>
                      <th>案件编号</th>
                      <th>案件名称</th>
                      <th>冲突类型</th>
                      <th>风险</th>
                      <th>说明</th>
                    </tr>
                  </thead>
                  <tbody>
                    {approvalConflictCases.map((item: any, index: number) => (
                      <tr
                        key={
                          textValue(item.id || item.case_id || item.caseId, '') ||
                          `restricted-conflict-${index}`
                        }
                      >
                        <td>
                          {textValue(item.case_no || item.caseNo || item.case_id || item.caseId)}
                        </td>
                        <td>{textValue(item.case_name || item.caseName)}</td>
                        <td>{textValue(item.conflict_type || item.conflictType)}</td>
                        <td>
                          <RiskTag
                            text={
                              canViewConflictEvidence
                                ? riskLabel(textValue(item.risk_level || item.riskLevel, 'LOW'))
                                : '受限'
                            }
                          />
                        </td>
                        <td>{textValue(item.description, '-')}</td>
                      </tr>
                    ))}
                    {approvalConflictCases.length === 0 && (
                      <tr>
                        <td colSpan={5}>暂无冲突命中明细</td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </DataTable>
            </SectionCard>

            <SectionCard title='审批材料'>
              <div className='batch-tabs compact'>
                {['全部材料', '冲突报告', '申请文档', '证明材料', '其他'].map((tab) => (
                  <button
                    key={tab}
                    className={materialFilter === tab ? 'active' : ''}
                    onClick={() => setMaterialFilter(tab)}
                  >
                    {tab}
                  </button>
                ))}
              </div>
              <DataTable>
                <table>
                  <tbody>
                    {filteredApprovalMaterials.map((row) => (
                      <tr key={row.name || row.id}>
                        <td>
                          <FileTextOutlined /> {textValue(row.name)}
                        </td>
                        <td>{materialTypeLabel(row.material_type || row.type)}</td>
                        <td>{statusLabel(textValue(row.status))}</td>
                        <td>{formatApiDate(row.created_at)}</td>
                        <td>
                          <Button
                            type='link'
                            disabled={!row.storage_url}
                            onClick={() =>
                              row.storage_url &&
                              window.open(
                                textValue(row.storage_url),
                                '_blank',
                                'noopener,noreferrer',
                              )
                            }
                          >
                            预览
                          </Button>
                        </td>
                      </tr>
                    ))}
                    {filteredApprovalMaterials.length === 0 && (
                      <tr>
                        <td colSpan={5}>
                          {approvalMaterialRows.length === 0
                            ? '审批快照暂无材料记录'
                            : `暂无${materialFilter}`}
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </DataTable>
            </SectionCard>

            <SectionCard title={`审批记录（${approvalTraceRows.length}）`}>
              {approvalHasPreviousTerminalRecords && (
                <p className='batch-history-notice'>
                  该申请已重新进入审批流程。以下包含前次处理记录；本轮状态以页面顶部为准。
                </p>
              )}
              <div className='batch-comment-list'>
                {approvalTraceRows.map((item) => (
                  <article key={item.id || `${item.approver_name}-${item.created_at}`}>
                    <Avatar icon={<UserOutlined />} />
                    <div>
                      <strong>
                        {textValue(item.approver_name, '审批人（历史记录未保存姓名）')}{' '}
                        {textValue(item.approver_role, '') && (
                          <span>{roleLabel(textValue(item.approver_role, ''))}</span>
                        )}
                      </strong>
                      <p>{textValue(item.decision_reason || item.decision_comments)}</p>
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
              `审批状态 ${statusLabel(approval?.status || '加载中')}`,
              `优先级 ${priorityLabel(approval?.priority)}`,
              `发起时间 ${approval?.created_at ? new Date(approval.created_at).toLocaleString() : '加载中'}`,
              `成案状态 ${caseCreation?.case_id ? '已生成正式案件' : approval?.status === 'approved' ? '审批已通过，待生成正式案件' : '尚未成案（等待审批通过）'}`,
              `正式案件 ${caseCreation?.case_number || '审批通过后生成'}`,
            ].map((line) => (
              <p key={line}>{line}</p>
            ))}
          </SectionCard>
          <SectionCard title='关联信息'>
            {relatedInfoRows.map((line) => (
              <p key={line}>{line}</p>
            ))}
          </SectionCard>
        </aside>
      </div>

      <div className='batch-bottom-bar approval-actions'>
        <Button onClick={() => navigate('/approval')}>返回</Button>
        {(approval?.status === 'submitted' ||
          approval?.status === 'under_review' ||
          approval?.status === 'pending') &&
          canDecisionActions && (
            <Space>
              <Tooltip
                title={
                  conflictApprovalBlocked
                    ? '冲突复核未形成可成案结论，补充材料并重新复核后才能同意并成案'
                    : undefined
                }
              >
                <span>
                  <Button
                    type='primary'
                    className='approve-btn'
                    icon={<CheckCircleOutlined />}
                    loading={deciding}
                    disabled={!canApproveApproval}
                    onClick={() => openDecisionDialog('approve')}
                  >
                    同意并成案
                  </Button>
                </span>
              </Tooltip>
              {approvalAccess.canReject && (
                <Button
                  danger
                  type='primary'
                  loading={deciding}
                  onClick={() => openDecisionDialog('reject')}
                >
                  拒绝
                </Button>
              )}
              {approvalAccess.canReturn && (
                <Button
                  className='return-btn'
                  loading={deciding}
                  onClick={() => openDecisionDialog('request_changes')}
                >
                  退回修改
                </Button>
              )}
            </Space>
          )}
        {(approval?.status === 'submitted' ||
          approval?.status === 'under_review' ||
          approval?.status === 'pending') &&
          !canDecisionActions && (
            <Space>
              <span className='batch-approval-readonly'>
                {hasConflictEvidenceMismatch
                  ? '审批证据不一致，已禁止处理'
                  : conflictApprovalBlocked
                    ? '冲突复核尚未形成可成案结论，当前不能同意并成案。'
                    : approvalAccess.readonlyReason}
              </span>
            </Space>
          )}
      </div>

      <Modal
        title={decisionDialog ? decisionDialogCopy[decisionDialog].title : '审批处理确认'}
        open={Boolean(decisionDialog)}
        onCancel={closeDecisionDialog}
        onOk={submitDecision}
        okText={decisionDialog ? decisionDialogCopy[decisionDialog].okText : '确认'}
        cancelText='取消'
        confirmLoading={deciding}
        okButtonProps={{ danger: decisionDialog === 'reject' }}
        destroyOnHidden
      >
        {decisionDialog && (
          <Space direction='vertical' style={{ width: '100%' }}>
            <p>{decisionDialogCopy[decisionDialog].notice}</p>
            <label htmlFor='approval-decision-reason'>处理依据（必填）</label>
            <Input.TextArea
              id='approval-decision-reason'
              aria-label='处理依据'
              rows={5}
              maxLength={1000}
              showCount
              value={decisionReason}
              onChange={(event) => setDecisionReason(event.target.value)}
              placeholder={decisionDialogCopy[decisionDialog].placeholder}
            />
            <p className='batch-muted'>
              处理依据将写入审批记录，提交后不能删除，只能追加更正说明。
            </p>
          </Space>
        )}
      </Modal>
    </div>
  )
}

export function ApprovalWorkbench() {
  const navigate = useNavigate()
  const [workbench, setWorkbench] = React.useState<any>({ items: [], stats: {} })
  const [approvalSearch, setApprovalSearch] = React.useState('')
  const [approvalFilter, setApprovalFilter] = React.useState('全部')

  React.useEffect(() => {
    apiRequest<any>('/approvals/workbench')
      .then((data) => {
        setWorkbench(data || { items: [], stats: {} })
      })
      .catch(() => setWorkbench({ items: [], stats: {} }))
  }, [])

  const approvalRows = listOf<Record<string, unknown>>(workbench.items)
  const filteredApprovalRows = approvalRows.filter((item) => {
    const query = approvalSearch.trim().toLowerCase()
    const matchesSearch =
      !query ||
      [item.request_number, item.title, item.applicant_name, item.current_approver_name]
        .map((value) => textValue(value, ''))
        .join(' ')
        .toLowerCase()
        .includes(query)
    if (!matchesSearch) return false
    if (approvalFilter === '冲突审查') return item.type === 'conflict_approval'
    if (approvalFilter === '豁免披露') return item.type === 'waiver' || item.category === 'waiver'
    if (approvalFilter === '待补充') return item.status === 'needs_revision'
    if (approvalFilter === '已超时')
      return Boolean(
        item.timeout_at &&
          new Date(String(item.timeout_at)).getTime() < Date.now() &&
          item.status !== 'approved',
      )
    return true
  })
  const approvalItems = filteredApprovalRows.length
    ? filteredApprovalRows.map((item) => [
        item.request_number || item.id,
        item.title || item.content || '未命名审批',
        priorityLabel(textValue(item.priority, 'medium')),
        item.current_stage
          ? approvalStageLabel(item.current_stage)
          : statusLabel(textValue(item.status, 'pending')),
        item.current_approver_name || item.applicant_name || '未分配',
        item.status === 'approved'
          ? '已完成'
          : item.timeout_at && new Date(String(item.timeout_at)).getTime() < Date.now()
            ? '已超时'
            : '正常',
        item.id,
      ])
    : []
  const conflictApprovalCount = numberValue(
    workbench.queues?.find?.((queue: any) => queue.key === 'conflict')?.count,
  )
  const waiverReviewCount = numberValue(workbench.stats?.waiver_review)
  const riskDistributionTotal = Math.max(
    conflictApprovalCount + waiverReviewCount,
    approvalItems.length,
  )
  const conflictApprovalPercent =
    riskDistributionTotal > 0
      ? Math.round((conflictApprovalCount / riskDistributionTotal) * 100)
      : 0
  const slaWarningItems = approvalItems.filter((row: any[]) => row[5] !== '正常')
  const exportApprovals = () => {
    if (filteredApprovalRows.length === 0) {
      message.warning('当前搜索或筛选条件下暂无可导出审批')
      return
    }
    const blob = new Blob(
      [
        JSON.stringify(
          { generatedAt: new Date().toISOString(), rows: filteredApprovalRows },
          null,
          2,
        ),
      ],
      { type: 'application/json;charset=utf-8' },
    )
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `approvals-${new Date().toISOString().slice(0, 10)}.json`
    link.click()
    URL.revokeObjectURL(url)
    message.success(`已导出 ${filteredApprovalRows.length} 条审批`)
  }

  return (
    <div className='batch-page approval-workbench-page'>
      <PageHeader
        eyebrow='审批中心 / 审批工作台'
        title='审批工作台'
        subtitle='聚合冲突审查、豁免披露、接案审批、费用审批和退回补充，形成审批闭环入口。'
        actions={
          <>
            <Input
              id='approval-search'
              name='approvalSearch'
              prefix={<SearchOutlined />}
              value={approvalSearch}
              onChange={(event) => setApprovalSearch(event.target.value)}
              allowClear
              placeholder='搜索审批编号、标题、发起人或审批人'
            />
            <Button icon={<DownloadOutlined />} onClick={exportApprovals}>
              导出
            </Button>
          </>
        }
      />

      <div className='ng-content'>
        <div className='batch-metric-grid approval-metrics'>
          {[
            {
              icon: <AuditOutlined />,
              label: '待我审批',
              value: workbench.stats?.pending ?? 0,
              delta: '',
              tone: 'blue' as Tone,
            },
            {
              icon: <SafetyCertificateOutlined />,
              label: '冲突审查',
              value: workbench.queues?.find?.((queue: any) => queue.key === 'conflict')?.count ?? 0,
              delta: '',
              tone: 'red' as Tone,
            },
            {
              icon: <FileProtectOutlined />,
              label: '豁免披露',
              value: workbench.stats?.waiver_review ?? 0,
              delta: '',
              tone: 'orange' as Tone,
            },
            {
              icon: <ClockCircleOutlined />,
              label: '需补充',
              value: workbench.stats?.needs_revision ?? 0,
              delta: '',
              tone: 'red' as Tone,
            },
            {
              icon: <CheckCircleOutlined />,
              label: '队列总数',
              value: approvalItems.length,
              delta: '',
              tone: 'teal' as Tone,
            },
          ].map((item) => (
            <MetricCard key={item.label} {...item} />
          ))}
        </div>

        <div className='batch-approval-board'>
          <SectionCard
            title='审批队列'
            extra={
              <Space>
                {['全部', '冲突审查', '豁免披露', '待补充', '已超时'].map((tab) => (
                  <Button
                    key={tab}
                    type={approvalFilter === tab ? 'primary' : 'default'}
                    onClick={() => setApprovalFilter(tab)}
                  >
                    {tab}
                  </Button>
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
                    <th>优先级</th>
                    <th>当前节点</th>
                    <th>负责人</th>
                    <th>SLA</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  {approvalItems.map((row: any[]) => (
                    <tr
                      key={row[0]}
                      className={['高', '紧急'].includes(row[2]) ? 'danger-row' : ''}
                    >
                      <td>{row[0]}</td>
                      <td>{row[1]}</td>
                      <td>
                        <RiskTag text={row[2]} />
                      </td>
                      <td>{row[3]}</td>
                      <td>{row[4]}</td>
                      <td className={row[5].includes('超时') ? 'danger-text' : ''}>{row[5]}</td>
                      <td>
                        <Button
                          size='small'
                          aria-label={`进入审批 ${row[0]}`}
                          onClick={() => navigate(`/approval/${row[6]}`)}
                        >
                          进入审批
                        </Button>
                      </td>
                    </tr>
                  ))}
                  {approvalItems.length === 0 && (
                    <tr>
                      <td colSpan={7}>
                        {approvalRows.length === 0
                          ? '暂无数据库审批队列'
                          : '当前搜索或筛选条件下暂无审批'}
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </DataTable>
          </SectionCard>

          <SectionCard title='审批风险分布'>
            <div className='batch-donut-card'>
              <Progress
                type='circle'
                percent={conflictApprovalPercent}
                format={() => String(riskDistributionTotal)}
                strokeColor='#e8434e'
                trailColor='#f4d8dc'
                size={126}
              />
              <div className='batch-legend'>
                <span>
                  <StatusDot color='red' />
                  冲突审批 <strong>{conflictApprovalCount}</strong>
                </span>
                <span>
                  <StatusDot color='orange' />
                  豁免评估 <strong>{waiverReviewCount}</strong>
                </span>
                <span>
                  <StatusDot color='teal' />
                  待补充 <strong>{workbench.stats?.needs_revision ?? 0}</strong>
                </span>
                <span>
                  <StatusDot color='blue' />
                  全部 <strong>{approvalItems.length}</strong>
                </span>
              </div>
            </div>
          </SectionCard>

          <SectionCard title='SLA 预警'>
            <div className='batch-overdue-list'>
              {slaWarningItems.slice(0, 4).map((row: any[], index: number) => (
                <p key={row[0]}>
                  <StatusDot color={index === 0 ? 'red' : 'orange'} />
                  {row[1]}
                  <span className='danger-text'>{row[5]}</span>
                </p>
              ))}
              {slaWarningItems.length === 0 && <p>暂无数据库 SLA 预警</p>}
            </div>
          </SectionCard>

          <SectionCard title='豁免与披露进度'>
            <p>
              {waiverReviewCount > 0
                ? `当前有 ${waiverReviewCount} 项豁免申请待复核，详情以审批队列为准。`
                : '当前没有进行中的豁免与披露流程。'}
            </p>
          </SectionCard>

          <SectionCard title='最近审批意见' className='span-2'>
            <div className='batch-comment-list'>
              {approvalRows.slice(0, 5).map((item) => (
                <article key={`approval-${textValue(item.id || item.request_number)}`}>
                  <Avatar icon={<UserOutlined />} />
                  <div>
                    <strong>
                      {textValue(item.current_approver_name || item.applicant_name)}
                      <span>{approvalStageLabel(item.current_stage || item.type)}</span>
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
            <span className='batch-autosave'>
              加载耗时：{apiTiming === null ? '加载中' : `${apiTiming}ms`}
            </span>
            <Button icon={<CalendarOutlined />}>排期视图</Button>
            <Button type='primary' icon={<PlusOutlined />}>
              新增律师
            </Button>
          </>
        }
      />

      <div className='batch-metric-grid lawyer-metrics'>
        {[
          {
            icon: <TeamOutlined />,
            label: '执业律师',
            value: resource.summary?.lawyers ?? lawyerRows.length,
            delta: '',
            tone: 'blue' as Tone,
          },
          {
            icon: <ApartmentOutlined />,
            label: '部门覆盖',
            value: resource.summary?.departments ?? capacityRows.length,
            delta: '来自 users.department',
            tone: 'teal' as Tone,
          },
          {
            icon: <FolderOpenOutlined />,
            label: '在办案件',
            value: resource.summary?.active_cases ?? assignmentRows.length,
            delta: '来自 cases',
            tone: 'orange' as Tone,
          },
          {
            icon: <SafetyCertificateOutlined />,
            label: '可分配账号',
            value: lawyerRows.filter((row) => textValue(row.status, '').toLowerCase() === 'active')
              .length,
            delta: 'active 状态',
            tone: 'green' as Tone,
          },
          {
            icon: <FileDoneOutlined />,
            label: '待处理事项',
            value: resource.summary?.pending_tasks ?? taskRows.length,
            delta: '来自 inbox_items',
            tone: 'red' as Tone,
          },
        ].map((item) => (
          <MetricCard key={item.label} {...item} />
        ))}
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
                  <tr
                    key={textValue(row.id || row.email)}
                    className={
                      textValue(row.status, '').toLowerCase() !== 'active' ? 'danger-row' : ''
                    }
                  >
                    <td>
                      <Avatar size='small' icon={<UserOutlined />} />{' '}
                      {textValue(row.name || row.username)}
                    </td>
                    <td>{roleLabel(textValue(row.role, ''))}</td>
                    <td>{textValue(row.department)}</td>
                    <td>{textValue(row.seniority)}</td>
                    <td>{textValue(row.email)}</td>
                    <td>
                      <RiskTag text={accountStatusLabel(textValue(row.status, ''))} />
                    </td>
                    <td>{formatApiDate(textValue(row.created_at, ''))}</td>
                    <td>
                      <Button size='small' onClick={() => navigate(`/lawyer/${textValue(row.id)}`)}>
                        查看档案
                      </Button>
                    </td>
                  </tr>
                ))}
                {lawyerRows.length === 0 && (
                  <tr>
                    <td colSpan={8}>暂无数据库律师账号</td>
                  </tr>
                )}
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
                <Progress
                  percent={Math.min(100, numberValue(item.count) * 20)}
                  size='small'
                  showInfo={false}
                />
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
                <p>
                  负责人：{textValue(row.lawyer_name)} 客户：{textValue(row.client_name)}
                </p>
                <RiskTag text={statusLabel(row.status)} />
                <Button type='primary' onClick={() => navigate('/conflict')}>
                  进入冲突检查
                </Button>
              </article>
            ))}
            {assignmentRows.length === 0 && <p>暂无数据库案件指派</p>}
          </div>
        </SectionCard>

        <SectionCard title='待处理律师事项'>
          <div className='batch-overdue-list'>
            {taskRows.map((task) => (
              <p key={textValue(task.id || task.title)}>
                <StatusDot
                  color={
                    task.priority === 'critical' || task.priority === 'high' ? 'red' : 'orange'
                  }
                />
                {workItemTypeLabel(task.type || task.source_type)} · {textValue(task.title)}
                <RiskTag text={priorityLabel(textValue(task.priority, 'medium'))} />
              </p>
            ))}
            {taskRows.length === 0 && <p>暂无数据库待处理事项</p>}
          </div>
        </SectionCard>

        <SectionCard title='业务领域覆盖'>
          <div className='batch-scope-grid'>
            {capacityRows.map((item) => (
              <span key={textValue(item.key)}>
                {textValue(item.key, '未分配部门')} {numberValue(item.count)}人
              </span>
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
          actions={
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/lawyer')}>
              返回律师资源
            </Button>
          }
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
          subtitle='未找到该律师账号。'
          actions={
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/lawyer')}>
              返回律师资源
            </Button>
          }
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
            <span className='batch-autosave'>
              加载耗时：{apiTiming === null ? '加载中' : `${apiTiming}ms`}
            </span>
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/lawyer')}>
              返回律师资源
            </Button>
          </>
        }
      />

      <div className='batch-metric-grid lawyer-metrics'>
        {[
          {
            icon: <UserOutlined />,
            label: '账号状态',
            value: accountStatusLabel(textValue(lawyer.status, 'active')),
            delta: 'users.status',
            tone:
              textValue(lawyer.status, '').toLowerCase() === 'active'
                ? ('green' as Tone)
                : ('orange' as Tone),
          },
          {
            icon: <FolderOpenOutlined />,
            label: '负责案件',
            value: caseRows.length,
            delta: 'cases.lawyer_id',
            tone: 'blue' as Tone,
          },
          {
            icon: <FileDoneOutlined />,
            label: '在办案件',
            value: activeCases.length,
            delta: 'active/pending',
            tone: 'teal' as Tone,
          },
          {
            icon: <AlertOutlined />,
            label: '高优先级',
            value: highPriorityCases.length,
            delta: 'priority',
            tone: highPriorityCases.length > 0 ? ('red' as Tone) : ('green' as Tone),
          },
          {
            icon: <ClockCircleOutlined />,
            label: '入库时间',
            value: formatApiDate(textValue(lawyer.created_at, '')),
            delta: '',
            tone: 'slate' as Tone,
          },
        ].map((item) => (
          <MetricCard key={item.label} {...item} />
        ))}
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
            <p>
              <StatusDot
                color={textValue(lawyer.status, '').toLowerCase() === 'active' ? 'green' : 'orange'}
              />
              账号状态
              <RiskTag text={accountStatusLabel(textValue(lawyer.status, 'active'))} />
            </p>
            <p>
              <StatusDot color='blue' />
              所属部门
              <RiskTag text={textValue(lawyer.department, '未分配')} />
            </p>
            <p>
              <StatusDot color='slate' />
              最近更新
              <RiskTag
                text={formatApiDate(textValue(lawyer.updated_at || lawyer.created_at, ''))}
              />
            </p>
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
                    <td>
                      <RiskTag text={dbCaseType(textValue(row.case_type, ''))} />
                    </td>
                    <td>
                      <RiskTag text={statusLabel(textValue(row.status, ''))} />
                    </td>
                    <td>
                      <RiskTag text={priorityLabel(textValue(row.priority, 'medium'))} />
                    </td>
                    <td>
                      <Button size='small' onClick={() => navigate(`/case/${textValue(row.id)}`)}>
                        查看案件
                      </Button>
                    </td>
                  </tr>
                ))}
                {caseRows.length === 0 && (
                  <tr>
                    <td colSpan={7}>暂无数据库案件指派</td>
                  </tr>
                )}
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
            <span className='batch-autosave'>
              加载耗时：{apiTiming === null ? '加载中' : `${apiTiming}ms`}
            </span>
            <Button icon={<DownloadOutlined />}>导出</Button>
            <Button type='primary' icon={<PlusOutlined />}>
              新增用户
            </Button>
          </>
        }
      />

      <div className='batch-metric-grid user-metrics'>
        {[
          {
            icon: <UserOutlined />,
            label: '系统用户',
            value: access.summary?.users ?? userRows.length,
            delta: '',
            tone: 'blue' as Tone,
          },
          {
            icon: <TeamOutlined />,
            label: '活跃账号',
            value:
              access.summary?.active_users ??
              userRows.filter((row) => textValue(row.status, '').toLowerCase() === 'active').length,
            delta: 'users.status',
            tone: 'teal' as Tone,
          },
          {
            icon: <KeyOutlined />,
            label: '角色数量',
            value: access.summary?.roles ?? roleRows.length,
            delta: 'RBAC 角色',
            tone: 'orange' as Tone,
          },
          {
            icon: <LockOutlined />,
            label: '停用/锁定',
            value: access.summary?.disabled_users ?? 0,
            delta: '非 active 状态',
            tone: 'green' as Tone,
          },
          {
            icon: <AuditOutlined />,
            label: '权限变更待审',
            value: access.summary?.pending_changes ?? changeRows.length,
            delta: 'approval_requests',
            tone: 'red' as Tone,
          },
        ].map((item) => (
          <MetricCard key={item.label} {...item} />
        ))}
      </div>

      <div className='batch-admin-layout'>
        <SectionCard
          title='账号清单'
          className='span-2'
          extra={
            <Space>
              {['全部', '管理员', '律师', '助理', '合规', '停用'].map((tab, index) => (
                <Button key={tab} type={index === 0 ? 'primary' : 'default'}>
                  {tab}
                </Button>
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
                  <tr
                    key={textValue(row.id || row.email)}
                    className={
                      textValue(row.status, '').toLowerCase() !== 'active' ? 'danger-row' : ''
                    }
                  >
                    <td>
                      <Avatar size='small' icon={<UserOutlined />} />{' '}
                      {textValue(row.name || row.username)}
                    </td>
                    <td>{textValue(row.seniority || row.department)}</td>
                    <td>{textValue(row.email)}</td>
                    <td>
                      <RiskTag text={roleLabel(textValue(row.role, ''))} />
                    </td>
                    <td>
                      <RiskTag text={accountStatusLabel(textValue(row.status, ''))} />
                    </td>
                    <td>{formatApiDate(textValue(row.updated_at || row.created_at, ''))}</td>
                    <td>{textValue(row.department, '未分配部门')}</td>
                    <td>
                      <Button size='small' onClick={() => openRoleEditor(row)}>
                        编辑角色
                      </Button>
                    </td>
                  </tr>
                ))}
                {userRows.length === 0 && (
                  <tr>
                    <td colSpan={8}>暂无数据库用户账号</td>
                  </tr>
                )}
              </tbody>
            </table>
          </DataTable>
        </SectionCard>

        <SectionCard title='角色矩阵'>
          <div className='batch-role-list'>
            {roleRows.map((role) => (
              <article key={textValue(role.key)}>
                <div>
                  <strong>
                    {roleLabel(textValue(role.key, ''))} <span>{numberValue(role.count)}人</span>
                  </strong>
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
                <StatusDot
                  color={textValue(item.priority, '').toLowerCase() === 'high' ? 'red' : 'orange'}
                />
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
                <div>
                  <strong>{textValue(item.event_type || item.action, '审计事件')}</strong>
                  <p>{textValue(item.summary || item.description, '暂无审计说明')}</p>
                </div>
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
  const approvalSettings = settingRows.filter((row) =>
    textValue(row.category, '').includes('approval'),
  )
  const notificationSettings = settingRows.filter(
    (row) =>
      textValue(row.category, '').includes('notification') ||
      textValue(row.category, '').includes('sla'),
  )
  const auditSettings = settingRows.filter(
    (row) =>
      textValue(row.category, '').includes('audit') ||
      textValue(row.category, '').includes('security') ||
      textValue(row.category, '').includes('file'),
  )

  return (
    <div className='batch-page system-settings-page'>
      <PageHeader
        eyebrow='系统管理 / 配置中心'
        title='系统设置'
        subtitle='围绕接案、冲突、审批、通知、审计和文件归档配置系统级规则。'
        actions={
          <>
            <span className='batch-autosave'>
              加载耗时：{apiTiming === null ? '加载中' : `${apiTiming}ms`}
            </span>
            <Button icon={<DownloadOutlined />}>导出配置</Button>
            <Button>恢复默认</Button>
            <Button type='primary' icon={<SettingOutlined />}>
              保存设置
            </Button>
          </>
        }
      />

      <div className='batch-metric-grid settings-metrics'>
        {[
          {
            icon: <SettingOutlined />,
            label: '配置项',
            value: overview.summary?.settings ?? settingRows.length,
            delta: 'system_settings',
            tone: 'blue' as Tone,
          },
          {
            icon: <SafetyCertificateOutlined />,
            label: '配置分组',
            value: overview.summary?.modules ?? moduleRows.length,
            delta: 'category 聚合',
            tone: 'red' as Tone,
          },
          {
            icon: <BellOutlined />,
            label: '通知策略',
            value: notificationSettings.length,
            delta: '',
            tone: 'orange' as Tone,
          },
          {
            icon: <FileProtectOutlined />,
            label: '审计策略',
            value: auditSettings.length,
            delta: '',
            tone: 'teal' as Tone,
          },
          {
            icon: <CloudUploadOutlined />,
            label: '启用配置',
            value: settingRows.filter(settingEnabled).length,
            delta: 'setting_value.enabled',
            tone: 'green' as Tone,
          },
        ].map((item) => (
          <MetricCard key={item.label} {...item} />
        ))}
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
                    <td>
                      <RiskTag text='数据库配置' />
                    </td>
                    <td>
                      <RiskTag text={numberValue(row.count) > 3 ? '高优先级' : '标准'} />
                    </td>
                    <td>
                      <Switch defaultChecked={numberValue(row.count) > 0} />
                    </td>
                    <td>
                      <Button size='small'>配置</Button>
                    </td>
                  </tr>
                ))}
                {moduleRows.length === 0 && (
                  <tr>
                    <td colSpan={6}>暂无数据库配置分组</td>
                  </tr>
                )}
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
            {approvalSettings.length === 0 && (
              <p>
                <span>暂无数据库审批规则</span>
              </p>
            )}
          </div>
        </SectionCard>

        <SectionCard title='通知与 SLA'>
          <div className='batch-settings-stack'>
            {notificationSettings.map((item) => (
              <p key={textValue(item.id || item.setting_key)}>
                <span>{textValue(item.description || item.setting_key)}</span>
                <RiskTag
                  text={textValue(
                    settingObject(item.setting_value).channel || item.category,
                    '已配置',
                  )}
                />
              </p>
            ))}
            {notificationSettings.length === 0 && (
              <p>
                <span>暂无数据库通知策略</span>
              </p>
            )}
          </div>
        </SectionCard>

        <SectionCard title='数据与审计' className='span-2'>
          <div className='batch-policy-grid'>
            {auditSettings.map((item) => (
              <article key={textValue(item.id || item.setting_key)}>
                <LockOutlined />
                <div>
                  <strong>{textValue(item.setting_key)}</strong>
                  <p>{textValue(item.description, '暂无配置说明')}</p>
                </div>
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

interface ConflictPolicyPackageView {
  package: Record<string, any>
  endorsements: Array<Record<string, any>>
  status: string
}

const emptyPolicyForm = () => {
  const now = new Date()
  const review = new Date(now)
  review.setFullYear(review.getFullYear() + 1)
  const localValue = (date: Date) => {
    const offset = date.getTimezoneOffset() * 60_000
    return new Date(date.getTime() - offset).toISOString().slice(0, 16)
  }
  return {
    policy_version: '',
    jurisdiction: '',
    applicable_rule_name: '',
    applicable_rule_version: '',
    applicable_rule_authority: '',
    applicable_rule_reference: '',
    data_source_policy_reference: '',
    privacy_basis_matrix_reference: '',
    retention_policy_reference: '',
    waiver_policy_reference: '',
    controlled_actions_reference: '',
    external_review_reference: '',
    effective_at: localValue(now),
    next_review_at: localValue(review),
    expires_at: '',
    integrity_hash: '',
  }
}

function localDateTimeValue(value?: string) {
  const date = value ? new Date(value) : new Date()
  if (Number.isNaN(date.getTime())) return ''
  const offset = date.getTimezoneOffset() * 60_000
  return new Date(date.getTime() - offset).toISOString().slice(0, 16)
}

const emptyScopeForm = (scopeType = '', scope?: Record<string, any>) => ({
  id: textValue(scope?.id, `scope-${scopeType.toLowerCase()}`),
  scope_type: scopeType,
  status: textValue(scope?.status, 'ACTIVE'),
  coverage_status: textValue(scope?.coverage_status, 'COVERAGE_LIMITED'),
  source_version: textValue(scope?.source_version, ''),
  evidence_reference: textValue(scope?.evidence_reference, ''),
  covered_from: localDateTimeValue(scope?.covered_from),
  covered_to: localDateTimeValue(scope?.covered_to),
  missing_sources: (() => {
    const value = scope?.missing_sources
    if (Array.isArray(value)) return value.join('\n')
    try {
      const parsed = JSON.parse(textValue(value, '[]'))
      return Array.isArray(parsed) ? parsed.join('\n') : textValue(value, '')
    } catch {
      return textValue(value, '')
    }
  })(),
  index_run_id: textValue(scope?.index_run_id, ''),
  source_of_truth: Boolean(scope?.source_of_truth),
  sync_mode: textValue(scope?.sync_mode, 'BATCH'),
  max_sync_lag_minutes: numberValue(scope?.max_sync_lag_minutes, 1440),
  last_successful_sync_at: localDateTimeValue(scope?.last_successful_sync_at),
  minimum_field_coverage_percent: numberValue(scope?.minimum_field_coverage_bps, 10000) / 100,
  measured_field_coverage_percent: numberValue(scope?.measured_field_coverage_bps) / 100,
  maximum_duplicate_rate_percent: numberValue(scope?.maximum_duplicate_rate_bps) / 100,
  measured_duplicate_rate_percent: numberValue(scope?.measured_duplicate_rate_bps) / 100,
  quality_owner_id: numberValue(scope?.quality_owner_id),
  quality_reviewed_at: localDateTimeValue(scope?.quality_reviewed_at),
  max_quality_review_age_days: numberValue(scope?.max_quality_review_age_days, 31),
  failure_alert_reference: textValue(scope?.failure_alert_reference, ''),
  correction_procedure_reference: textValue(scope?.correction_procedure_reference, ''),
})

function policyStatusLabel(status: string) {
  switch (status) {
    case 'APPROVED':
      return '双人确认完成'
    case 'PENDING_MANAGEMENT':
      return '待主任/管理合伙人确认'
    case 'PENDING_COMPLIANCE':
      return '待合规负责人确认'
    default:
      return '待双方确认'
  }
}

const conflictScopeLabels: Record<string, string> = {
  CASE_ARCHIVE: '案件档案',
  CLIENT_ARCHIVE: '客户档案',
  SUBJECT_REGISTRY: '客户及相关主体名册',
  RELATION_ARCHIVE: '关联关系档案',
}

function formatPolicyDate(value?: string) {
  if (!value) return '未设置'
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value))
}

export function ConflictGovernanceCenter() {
  const [policies, setPolicies] = React.useState<ConflictPolicyPackageView[]>([])
  const [scopes, setScopes] = React.useState<Array<Record<string, any>>>([])
  const [qualityOwners, setQualityOwners] = React.useState<Array<Record<string, any>>>([])
  const [loading, setLoading] = React.useState(true)
  const [submitting, setSubmitting] = React.useState(false)
  const [modalOpen, setModalOpen] = React.useState(false)
  const [form, setForm] = React.useState(emptyPolicyForm)
  const [scopeModalOpen, setScopeModalOpen] = React.useState(false)
  const [scopeForm, setScopeForm] = React.useState(() => emptyScopeForm())
  const roles = listOf(getRoles()).map((role) => textValue(role.code).toLowerCase())
  const isManagement = roles.some((role) => ['director', 'partner', 'management'].includes(role))
  const isCompliance = roles.some((role) => ['compliance', 'risk', 'risk_control'].includes(role))
  const canManagePolicies = isManagement || isCompliance

  const load = React.useCallback(async () => {
    setLoading(true)
    try {
      const [policyRows, scopeRows, ownerRows] = await Promise.all([
        canManagePolicies
          ? apiRequest<ConflictPolicyPackageView[]>('/conflict-v2/governance/policies')
          : Promise.resolve([]),
        apiRequest<Array<Record<string, any>>>('/conflict-v2/search-scopes'),
        apiRequest<Array<Record<string, any>>>('/conflict/reviewer-candidates'),
      ])
      setPolicies(Array.isArray(policyRows) ? policyRows : [])
      setScopes(Array.isArray(scopeRows) ? scopeRows : [])
      setQualityOwners(Array.isArray(ownerRows) ? ownerRows : [])
    } catch (error) {
      message.error(error instanceof Error ? error.message : '冲突治理信息加载失败')
    } finally {
      setLoading(false)
    }
  }, [canManagePolicies])

  React.useEffect(() => {
    void load()
  }, [load])

  const updateField = (field: keyof ReturnType<typeof emptyPolicyForm>, value: string) => {
    setForm((current) => ({ ...current, [field]: value }))
  }

  const submitPackage = async () => {
    setSubmitting(true)
    try {
      await apiRequest('/conflict-v2/governance/policies', {
        method: 'POST',
        body: JSON.stringify({
          ...form,
          effective_at: new Date(form.effective_at).toISOString(),
          next_review_at: new Date(form.next_review_at).toISOString(),
          expires_at: form.expires_at ? new Date(form.expires_at).toISOString() : null,
        }),
      })
      message.success('政策材料包已提交，需主任/管理合伙人与合规负责人分别确认')
      setModalOpen(false)
      setForm(emptyPolicyForm())
      await load()
    } catch (error) {
      message.error(error instanceof Error ? error.message : '政策材料包提交失败')
    } finally {
      setSubmitting(false)
    }
  }

  const openScopeEditor = (scopeType: string, scope?: Record<string, any>) => {
    setScopeForm(emptyScopeForm(scopeType, scope))
    setScopeModalOpen(true)
  }

  const updateScopeField = (field: string, value: unknown) => {
    setScopeForm((current) => ({ ...current, [field]: value }))
  }

  const submitScope = async () => {
    setSubmitting(true)
    try {
      const missingSources = String(scopeForm.missing_sources || '')
        .split(/[\n,，]/)
        .map((value) => value.trim())
        .filter(Boolean)
      await apiRequest(`/conflict-v2/search-scopes/${encodeURIComponent(scopeForm.id)}`, {
        method: 'PUT',
        body: JSON.stringify({
          ...scopeForm,
          covered_from: scopeForm.covered_from ? new Date(scopeForm.covered_from).toISOString() : null,
          covered_to: scopeForm.covered_to ? new Date(scopeForm.covered_to).toISOString() : null,
          last_successful_sync_at: scopeForm.last_successful_sync_at ? new Date(scopeForm.last_successful_sync_at).toISOString() : null,
          quality_reviewed_at: scopeForm.quality_reviewed_at ? new Date(scopeForm.quality_reviewed_at).toISOString() : null,
          missing_sources: missingSources,
          minimum_field_coverage_bps: Math.round(numberValue(scopeForm.minimum_field_coverage_percent) * 100),
          measured_field_coverage_bps: Math.round(numberValue(scopeForm.measured_field_coverage_percent) * 100),
          maximum_duplicate_rate_bps: Math.round(numberValue(scopeForm.maximum_duplicate_rate_percent) * 100),
          measured_duplicate_rate_bps: Math.round(numberValue(scopeForm.measured_duplicate_rate_percent) * 100),
          quality_owner_id: numberValue(scopeForm.quality_owner_id) || null,
        }),
      })
      message.success('权威档案来源登记已保存并留痕')
      setScopeModalOpen(false)
      await load()
    } catch (error) {
      message.error(error instanceof Error ? error.message : '权威档案来源登记失败')
    } finally {
      setSubmitting(false)
    }
  }

  const endorse = (item: ConflictPolicyPackageView) => {
    Modal.confirm({
      title: '确认当前政策材料包',
      content: `您将以当前账号确认政策版本 ${textValue(item.package.policy_version)}，完整性摘要为 ${textValue(item.package.integrity_hash)}。确认记录不可修改或删除。`,
      okText: '确认并留痕',
      cancelText: '取消',
      onOk: async () => {
        try {
          await apiRequest(`/conflict-v2/governance/policies/${item.package.id}/endorsements`, {
            method: 'POST',
            body: '{}',
          })
          message.success('确认记录已保存')
          await load()
        } catch (error) {
          message.error(error instanceof Error ? error.message : '确认记录保存失败')
        }
      },
    })
  }

  const canEndorse = (item: ConflictPolicyPackageView) => {
    if (item.status === 'APPROVED') return false
    const types = item.endorsements.map((entry) => textValue(entry.endorsement_type))
    return (isManagement && !types.includes('MANAGEMENT')) || (isCompliance && !types.includes('COMPLIANCE'))
  }

  const completeScopeTypes = new Set(
    scopes
      .filter((scope) => textValue(scope.status) === 'ACTIVE' && textValue(scope.coverage_status) === 'COMPLETE')
      .map((scope) => textValue(scope.scope_type)),
  )
  const requiredScopes = ['CASE_ARCHIVE', 'CLIENT_ARCHIVE', 'SUBJECT_REGISTRY', 'RELATION_ARCHIVE']

  return (
    <div className='batch-page'>
      <PageHeader
        eyebrow='利益冲突 / 正式放行治理'
        title='冲突治理'
        subtitle='政策签署与权威档案覆盖均完成后，系统才允许正式客户案件使用完整结论。'
        actions={
          <Space>
            <Button onClick={() => void load()} loading={loading}>刷新</Button>
            {canManagePolicies && (
              <Button type='primary' icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>
                提交新政策材料包
              </Button>
            )}
          </Space>
        }
      />

      <div className='batch-metric-grid'>
        {canManagePolicies && <MetricCard icon={<FileProtectOutlined />} label='政策材料包' value={policies.length} delta='' tone='blue' />}
        {canManagePolicies && <MetricCard icon={<CheckCircleOutlined />} label='双人确认完成' value={policies.filter((item) => item.status === 'APPROVED').length} delta='' tone='green' />}
        <MetricCard icon={<DatabaseOutlined />} label='权威来源完整' value={`${completeScopeTypes.size}/4`} delta='' tone={completeScopeTypes.size === 4 ? 'green' : 'orange'} />
      </div>

      {canManagePolicies && <SectionCard title='律所冲突政策签署记录'>
        <DataTable>
          <table>
            <thead><tr><th>政策版本</th><th>适用范围</th><th>生效/复核</th><th>确认状态</th><th>材料摘要</th><th>操作</th></tr></thead>
            <tbody>
              {policies.map((item) => (
                <tr key={textValue(item.package.id)}>
                  <td><strong>{textValue(item.package.policy_version)}</strong><br /><small>{textValue(item.package.applicable_rule_name)}</small></td>
                  <td>{textValue(item.package.jurisdiction)}<br /><small>{textValue(item.package.applicable_rule_authority)}</small></td>
                  <td>{formatPolicyDate(item.package.effective_at)}<br /><small>复核：{formatPolicyDate(item.package.next_review_at)}</small></td>
                  <td>
                    <RiskTag text={policyStatusLabel(item.status)} />
                    <div className='policy-endorser-list'>
                      {item.endorsements.map((entry) => (
                        <small key={textValue(entry.id)}>
                          {textValue(entry.endorsement_type) === 'MANAGEMENT' ? '管理确认' : '合规确认'}：{textValue(entry.endorser_name, '账号已停用')}
                        </small>
                      ))}
                    </div>
                  </td>
                  <td><code>{textValue(item.package.integrity_hash).slice(0, 12)}…</code></td>
                  <td><Button size='small' disabled={!canEndorse(item)} onClick={() => endorse(item)}>确认政策材料</Button></td>
                </tr>
              ))}
              {!loading && policies.length === 0 && <tr><td colSpan={6}>尚未提交政策材料包，正式客户案件保持未放行。</td></tr>}
            </tbody>
          </table>
        </DataTable>
      </SectionCard>}

      <SectionCard title='权威档案来源覆盖'>
        <div className='batch-policy-grid'>
          {requiredScopes.map((scopeType) => {
            const scope = scopes.find((row) => textValue(row.scope_type) === scopeType)
            const complete = scope && textValue(scope.status) === 'ACTIVE' && textValue(scope.coverage_status) === 'COMPLETE'
            return (
              <article key={scopeType}>
                <DatabaseOutlined />
                <div>
                  <strong>{conflictScopeLabels[scopeType]}</strong>
                  <p>{scope ? (complete ? '已完成索引和数量对账' : '已登记，尚未完成核对') : '未登记权威来源'}</p>
                </div>
                <div className='scope-card-actions'>
                  <RiskTag text={complete ? '完整并已核对' : '未完成'} />
                  <Button size='small' onClick={() => openScopeEditor(scopeType, scope)}>登记/更新</Button>
                </div>
              </article>
            )
          })}
        </div>
      </SectionCard>

      <Modal title='提交不可修改的政策材料包' open={modalOpen} onCancel={() => setModalOpen(false)} onOk={() => void submitPackage()} okText='提交材料包' cancelText='取消' confirmLoading={submitting} width={760} destroyOnHidden>
        <div className='conflict-policy-form'>
          <div><span>政策版本</span><Input aria-label='政策版本' value={form.policy_version} onChange={(event) => updateField('policy_version', event.target.value)} /></div>
          <div><span>执业登记地</span><Input aria-label='执业登记地' value={form.jurisdiction} onChange={(event) => updateField('jurisdiction', event.target.value)} /></div>
          <div><span>适用规则名称</span><Input aria-label='适用规则名称' value={form.applicable_rule_name} onChange={(event) => updateField('applicable_rule_name', event.target.value)} /></div>
          <div><span>规则版本</span><Input aria-label='规则版本' value={form.applicable_rule_version} onChange={(event) => updateField('applicable_rule_version', event.target.value)} /></div>
          <div><span>发布/确认机关</span><Input aria-label='发布或确认机关' value={form.applicable_rule_authority} onChange={(event) => updateField('applicable_rule_authority', event.target.value)} /></div>
          <div><span>适用规则材料引用</span><Input aria-label='适用规则材料引用' value={form.applicable_rule_reference} onChange={(event) => updateField('applicable_rule_reference', event.target.value)} /></div>
          <div><span>权威来源政策引用</span><Input aria-label='权威来源政策引用' value={form.data_source_policy_reference} onChange={(event) => updateField('data_source_policy_reference', event.target.value)} /></div>
          <div><span>个人信息处理依据引用</span><Input aria-label='个人信息处理依据引用' value={form.privacy_basis_matrix_reference} onChange={(event) => updateField('privacy_basis_matrix_reference', event.target.value)} /></div>
          <div><span>档案保留政策引用</span><Input aria-label='档案保留政策引用' value={form.retention_policy_reference} onChange={(event) => updateField('retention_policy_reference', event.target.value)} /></div>
          <div><span>冲突豁免政策引用</span><Input aria-label='冲突豁免政策引用' value={form.waiver_policy_reference} onChange={(event) => updateField('waiver_policy_reference', event.target.value)} /></div>
          <div><span>受控对外动作清单引用</span><Input aria-label='受控对外动作清单引用' value={form.controlled_actions_reference} onChange={(event) => updateField('controlled_actions_reference', event.target.value)} /></div>
          <div><span>外部或内部专项审阅引用</span><Input aria-label='外部或内部专项审阅引用' value={form.external_review_reference} onChange={(event) => updateField('external_review_reference', event.target.value)} /></div>
          <div><span>生效时间</span><Input aria-label='生效时间' type='datetime-local' value={form.effective_at} onChange={(event) => updateField('effective_at', event.target.value)} /></div>
          <div><span>下次复核时间</span><Input aria-label='下次复核时间' type='datetime-local' value={form.next_review_at} onChange={(event) => updateField('next_review_at', event.target.value)} /></div>
          <div><span>到期时间（可不填）</span><Input aria-label='到期时间' type='datetime-local' value={form.expires_at} onChange={(event) => updateField('expires_at', event.target.value)} /></div>
          <div className='span-2'><span>签署材料包 SHA-256 摘要</span><Input aria-label='签署材料包SHA-256摘要' value={form.integrity_hash} maxLength={64} onChange={(event) => updateField('integrity_hash', event.target.value)} /></div>
        </div>
      </Modal>
      <Modal title={`登记${conflictScopeLabels[scopeForm.scope_type] || '权威档案来源'}`} open={scopeModalOpen} onCancel={() => setScopeModalOpen(false)} onOk={() => void submitScope()} okText='保存并留痕' cancelText='取消' confirmLoading={submitting} width={820} destroyOnHidden>
        <div className='conflict-policy-form'>
          <div><span>来源状态</span><Select aria-label='来源状态' value={scopeForm.status} onChange={(value) => updateScopeField('status', value)} options={[{ value: 'ACTIVE', label: '当前有效' }, { value: 'INACTIVE', label: '已停用' }]} /></div>
          <div><span>覆盖结论</span><Select aria-label='覆盖结论' value={scopeForm.coverage_status} onChange={(value) => updateScopeField('coverage_status', value)} options={[{ value: 'COMPLETE', label: '完整覆盖' }, { value: 'COVERAGE_LIMITED', label: '覆盖受限' }]} /></div>
          <div><span>来源数据版本</span><Input aria-label='来源数据版本' value={scopeForm.source_version} onChange={(event) => updateScopeField('source_version', event.target.value)} /></div>
          <div><span>索引构建运行编号</span><Input aria-label='索引构建运行编号' value={scopeForm.index_run_id} onChange={(event) => updateScopeField('index_run_id', event.target.value)} /></div>
          <div className='span-2'><span>核对凭证引用</span><Input aria-label='核对凭证引用' value={scopeForm.evidence_reference} onChange={(event) => updateScopeField('evidence_reference', event.target.value)} /></div>
          <div><span>覆盖资料起始时间</span><Input aria-label='覆盖资料起始时间' type='datetime-local' value={scopeForm.covered_from} onChange={(event) => updateScopeField('covered_from', event.target.value)} /></div>
          <div><span>覆盖资料截止时间</span><Input aria-label='覆盖资料截止时间' type='datetime-local' value={scopeForm.covered_to} onChange={(event) => updateScopeField('covered_to', event.target.value)} /></div>
          <div className='span-2'><span>未纳入的资料来源（每行一项；完整覆盖时必须为空）</span><Input.TextArea aria-label='未纳入的资料来源' rows={3} value={scopeForm.missing_sources} onChange={(event) => updateScopeField('missing_sources', event.target.value)} /></div>
          <div><span>是否经律所确认为权威来源</span><Switch aria-label='是否经律所确认为权威来源' checked={scopeForm.source_of_truth} onChange={(value) => updateScopeField('source_of_truth', value)} checkedChildren='是' unCheckedChildren='否' /></div>
          <div><span>同步方式</span><Select aria-label='同步方式' value={scopeForm.sync_mode} onChange={(value) => updateScopeField('sync_mode', value)} options={[{ value: 'REALTIME', label: '实时同步' }, { value: 'BATCH', label: '定时批量同步' }, { value: 'MANUAL_IMPORT', label: '受控人工导入' }]} /></div>
          <div><span>允许最大同步延迟（分钟）</span><Input aria-label='允许最大同步延迟' type='number' min={1} value={scopeForm.max_sync_lag_minutes} onChange={(event) => updateScopeField('max_sync_lag_minutes', numberValue(event.target.value))} /></div>
          <div><span>最近成功同步时间</span><Input aria-label='最近成功同步时间' type='datetime-local' value={scopeForm.last_successful_sync_at} onChange={(event) => updateScopeField('last_successful_sync_at', event.target.value)} /></div>
          <div><span>最低字段覆盖率（%）</span><Input aria-label='最低字段覆盖率' type='number' min={0.01} max={100} step={0.01} value={scopeForm.minimum_field_coverage_percent} onChange={(event) => updateScopeField('minimum_field_coverage_percent', numberValue(event.target.value))} /></div>
          <div><span>实测字段覆盖率（%）</span><Input aria-label='实测字段覆盖率' type='number' min={0} max={100} step={0.01} value={scopeForm.measured_field_coverage_percent} onChange={(event) => updateScopeField('measured_field_coverage_percent', numberValue(event.target.value))} /></div>
          <div><span>允许最高重复率（%）</span><Input aria-label='允许最高重复率' type='number' min={0} max={100} step={0.01} value={scopeForm.maximum_duplicate_rate_percent} onChange={(event) => updateScopeField('maximum_duplicate_rate_percent', numberValue(event.target.value))} /></div>
          <div><span>实测重复率（%）</span><Input aria-label='实测重复率' type='number' min={0} max={100} step={0.01} value={scopeForm.measured_duplicate_rate_percent} onChange={(event) => updateScopeField('measured_duplicate_rate_percent', numberValue(event.target.value))} /></div>
          <div><span>数据质量责任人</span><Select aria-label='数据质量责任人' showSearch optionFilterProp='label' value={scopeForm.quality_owner_id || undefined} onChange={(value) => updateScopeField('quality_owner_id', numberValue(value))} options={qualityOwners.map((owner) => ({ value: numberValue(owner.id), label: `${textValue(owner.name || owner.username, '未命名账号')} · ${textValue(owner.department, textValue(owner.role))}` }))} /></div>
          <div><span>最近质量复核时间</span><Input aria-label='最近质量复核时间' type='datetime-local' value={scopeForm.quality_reviewed_at} onChange={(event) => updateScopeField('quality_reviewed_at', event.target.value)} /></div>
          <div><span>质量复核有效期（天）</span><Input aria-label='质量复核有效期' type='number' min={1} value={scopeForm.max_quality_review_age_days} onChange={(event) => updateScopeField('max_quality_review_age_days', numberValue(event.target.value))} /></div>
          <div><span>同步失败告警流程引用</span><Input aria-label='同步失败告警流程引用' value={scopeForm.failure_alert_reference} onChange={(event) => updateScopeField('failure_alert_reference', event.target.value)} /></div>
          <div className='span-2'><span>数据更正与留痕流程引用</span><Input aria-label='数据更正与留痕流程引用' value={scopeForm.correction_procedure_reference} onChange={(event) => updateScopeField('correction_procedure_reference', event.target.value)} /></div>
        </div>
      </Modal>
    </div>
  )
}
