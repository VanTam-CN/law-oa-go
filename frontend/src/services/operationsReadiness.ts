import { get, post } from './http'

export type OperationsRequirementId =
  | 'backup'
  | 'restoreDrill'
  | 'incidentOwner'
  | 'upgrade'
  | 'rollback'

export type OperationsHealthStatus = 'healthy' | 'unhealthy' | 'unknown'
export type OperationsEvidenceScope = 'qa' | 'controlled_pilot'
export type OperationsControlId = OperationsRequirementId

export interface VerifiedOperationsEvidence {
  verificationStatus: 'verified'
  reference: string
  verifiedAt: string
  verifiedBy: string
}

export interface OperationsRequirement {
  id: OperationsRequirementId
  title: string
  userQuestion: string
  evidenceToKeep: string[]
  nextAction: string
}

export interface OperationsReadinessView extends OperationsRequirement {
  status: 'pending-evidence' | 'verified'
}

export interface OperationsReadinessSummary {
  ready: boolean
  total: number
  verifiedCount: number
  pendingCount: number
  items: OperationsReadinessView[]
}

export interface RegisteredOperationsEvidence {
  id: string
  control: OperationsControlId
  scope: OperationsEvidenceScope
  evidenceReference: string
  reviewedBy: number
  reviewedAt: string
  notes?: string
  createdAt: string
}

export interface ServerOperationsReadinessSummary {
  scope: OperationsEvidenceScope
  ready: boolean
  score: number
  maximumScore: number
  verifiedCount: number
  total: number
  productionReady: boolean
  productionGate: 'production_external_evidence'
  items: Array<{
    control: OperationsControlId
    status: 'pending-evidence' | 'verified'
    evidence?: RegisteredOperationsEvidence
  }>
}

export interface OperationsEvidenceRegistrationInput {
  control: OperationsControlId
  scope: OperationsEvidenceScope
  evidenceReference: string
  reviewedAt: string
  notes?: string
}

const localControlToApi: Record<OperationsControlId, string> = {
  backup: 'backup',
  restoreDrill: 'restore_drill',
  incidentOwner: 'incident_owner',
  upgrade: 'upgrade',
  rollback: 'rollback',
}

const apiControlToLocal = Object.fromEntries(
  Object.entries(localControlToApi).map(([local, api]) => [api, local]),
) as Record<string, OperationsControlId>

const normalizeSummary = (summary: ServerOperationsReadinessSummary): ServerOperationsReadinessSummary => ({
  ...summary,
  items: summary.items.map((item) => ({
    ...item,
    control: apiControlToLocal[item.control] ?? item.control,
    evidence: item.evidence ? { ...item.evidence, control: apiControlToLocal[item.evidence.control] ?? item.evidence.control } : undefined,
  })),
})

export const OPERATIONS_READINESS_REQUIREMENTS: OperationsRequirement[] = [
  {
    id: 'backup',
    title: '数据备份',
    userQuestion: '案件、客户、代管款和文件数据是否已有备份？',
    evidenceToKeep: [
      '一次真实备份的执行时间和执行人',
      '备份编号或存放位置（不记录密码、密钥）',
      '文件大小或校验值',
    ],
    nextAction: '在受控环境执行一次数据库和上传文件备份，并留存备份记录。',
  },
  {
    id: 'restoreDrill',
    title: '恢复演练',
    userQuestion: '如果服务器损坏，能否用备份恢复业务数据？',
    evidenceToKeep: [
      '恢复演练使用的备份编号',
      '恢复到的隔离环境或端口',
      '恢复后核心数据抽查结果和完成时间',
    ],
    nextAction: '把最新备份恢复到隔离环境，抽查案件、客户和代管款数据后留存演练记录。',
  },
  {
    id: 'incidentOwner',
    title: '故障负责人与联系方式',
    userQuestion: '系统不可用时，谁来决定、联系和跟进？',
    evidenceToKeep: [
      '首席负责人和备用负责人姓名或岗位',
      '可执行的联系路径（不硬编码在系统页面中）',
      '最近一次联系路径测试记录',
    ],
    nextAction: '由律所确认首席/备用负责人和响应路径，并在隔离环境测试一次联系方式。',
  },
  {
    id: 'upgrade',
    title: '升级验证',
    userQuestion: '新版本升级前，谁确认哪些功能还能用？',
    evidenceToKeep: [
      '升级前后版本号和执行时间',
      '健康检查结果和核心业务功能抽查记录',
      '异常时的暂停标准',
    ],
    nextAction: '在受控环境执行一次版本升级，记录版本、健康检查和核心流程抽查结果。',
  },
  {
    id: 'rollback',
    title: '回滚方案',
    userQuestion: '升级失败后，如何退回可工作的版本？',
    evidenceToKeep: [
      '可回退的版本标识或制品编号',
      '数据兼容性说明和备份点',
      '一次回滚演练的结果、耗时和验证人',
    ],
    nextAction: '先在受控环境验证回到上一版本，并确认数据结构和业务数据仍可用。',
  },
]

export const summarizeOperationsReadiness = (
  evidence: Partial<Record<OperationsRequirementId, VerifiedOperationsEvidence>> = {},
  healthStatus: OperationsHealthStatus = 'unknown',
): OperationsReadinessSummary => {
  const items = OPERATIONS_READINESS_REQUIREMENTS.map((requirement) => ({
    ...requirement,
    status:
      evidence[requirement.id]?.verificationStatus === 'verified'
        ? ('verified' as const)
        : ('pending-evidence' as const),
  }))

  const verifiedCount = items.filter((item) => item.status === 'verified').length

  return {
    // A healthy service only proves current availability. It never substitutes for
    // restore, incident, upgrade, or rollback evidence.
    ready: healthStatus === 'healthy' && verifiedCount === items.length,
    total: items.length,
    verifiedCount,
    pendingCount: items.length - verifiedCount,
    items,
  }
}

export const getOperationsReadinessSummary = (
  scope: OperationsEvidenceScope = 'controlled_pilot',
): Promise<ServerOperationsReadinessSummary> =>
  get<ServerOperationsReadinessSummary>('/operations/readiness/evidence', { scope }).then(normalizeSummary)

export const registerOperationsEvidence = (
  input: OperationsEvidenceRegistrationInput,
): Promise<RegisteredOperationsEvidence> =>
  post<RegisteredOperationsEvidence>('/operations/readiness/evidence', {
    ...input,
    control: localControlToApi[input.control],
  }).then((evidence) => ({ ...evidence, control: apiControlToLocal[evidence.control] ?? evidence.control }))
