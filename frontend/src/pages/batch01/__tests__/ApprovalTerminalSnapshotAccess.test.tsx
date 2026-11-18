import React from 'react'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router'
import { useAppStore, type User } from '@/stores/useAppStore'
import { getUserInfo } from '@/utils/storage'
import { ApprovalDecisionFlow } from '../Batch01Prototype'

jest.mock('@/utils/storage', () => ({
  ...jest.requireActual('@/utils/storage'),
  getRoles: jest.fn(() => []),
  getToken: jest.fn(() => 'test-token'),
  getUserInfo: jest.fn(),
}))

jest.mock('@/utils/messageHelper', () => ({
  message: { error: jest.fn(), success: jest.fn(), warning: jest.fn() },
  setAppMessage: jest.fn(),
}))

const APPROVAL_ID = '74af5505-d006-4c67-9409-fc413ca2c184'
const SNAPSHOT_ID = '9f23a879-a204-4f97-986f-48c432567360'
const CHECK_ID = 'CCT_8b28e98f-af19-42b0-a89b-de717712e2fe'
const INTAKE_ID = 'INTAKE-RESUBMIT-1'
const CASE_NUMBER = 'CASE-20261003223337'
const LEGACY_CHECK_ID = 'CCT_892-legacy-payload'
const fetchMock = jest.fn()
global.fetch = fetchMock as unknown as typeof fetch

function jsonResponse(data: unknown): Response {
  return { ok: true, status: 200, json: async () => ({ success: true, data }) } as Response
}

function failedResponse(status: number): Response {
  return {
    ok: false,
    status,
    json: async () => ({ success: false, message: '无法读取提交快照' }),
  } as Response
}

function routeUser(role: string, permissions: string[] = []): User {
  return {
    id: role === 'lawyer' ? '9' : '2',
    username: role,
    email: '',
    realName: '虚构试用用户',
    roles: [role],
    permissions,
    isActive: true,
    createdAt: '2026-10-03T00:00:00Z',
  }
}

function currentSnapshot() {
  // Real round-2 slim contract: status/coverage are NOT frozen here.
  return {
    approval_id: APPROVAL_ID,
    intake_id: INTAKE_ID,
    round_number: 2,
    conflict_check_id: CHECK_ID,
    conflict_review_id: 88,
    conflict_reviewer_id: 3,
    conflict_review_decision: 'no_conflict',
    conflict_review_state: 'independent_review_current',
    structured_facts: {
      title: '新版浏览器练习甲（虚构）',
      case_type: 'commercial',
      description: '第二轮实际冻结案情',
    },
    case_creation_config: { source_snapshot_id: SNAPSHOT_ID, source: 'intake_submission_snapshot' },
    material_fingerprints: [{
      material_id: 5,
      name: '本轮冻结主体资料',
      material_type: 'evidence',
      status: 'uploaded',
      current_file_id: 13,
      files: [{ file_id: 13, file_name: '本轮冻结资料.pdf', digest: 'bbb' }],
    }],
    round_diff: {
      round_diff_version: 2,
      baseline: { facts_baseline: 'frozen', material_baseline: 'frozen' },
      field_changes: [{ field: 'title', before: '第一轮冻结标题', after: '新版浏览器练习甲（虚构）' }],
      material_changes: [{
        material_id: 5,
        name: '本轮冻结主体资料',
        change_type: 'files_changed',
        new_files: ['本轮冻结资料.pdf'],
      }],
    },
  }
}

type SetupOptions = {
  legacyVerdict?: 'green' | 'red'
  snapshotIssue?: number | 'pointer' | 'intake' | 'approval' | null
  role?: string
  permissions?: string[]
  accessible?: boolean
  status?: string
  intakeScoped?: boolean
  hasCase?: boolean
}

function setupFlow(options: SetupOptions = {}) {
  const {
    legacyVerdict = 'green',
    snapshotIssue = 403,
    role = 'director',
    permissions = [],
    accessible = true,
    status = 'approved',
    intakeScoped = true,
    hasCase = true,
  } = options
  const user = routeUser(role, permissions)
  useAppStore.setState({ user, isAuthenticated: true })
  ;(getUserInfo as jest.Mock).mockReturnValue({ id: Number(user.id), role })
  const legacyGreen = legacyVerdict === 'green'
  const legacyConflict = {
    checkId: LEGACY_CHECK_ID,
    riskAssessment: { overallRisk: legacyGreen ? 'LOW' : 'HIGH', riskScore: legacyGreen ? 0 : 100 },
    decision: {
      status: legacyGreen ? 'CLEAR' : 'BLOCKED',
      coverageStatus: legacyGreen ? 'COMPLETE' : 'LIMITED',
      evidenceCount: legacyGreen ? 0 : 1,
      restrictedCount: 0,
    },
    review: { decision: legacyGreen ? 'no_conflict' : 'confirmed_conflict' },
    conflictCases: [{ caseName: '旧轮命中案件', riskLevel: 'HIGH' }],
  }
  const root = currentSnapshot()
  if (snapshotIssue === 'intake') root.intake_id = 'INTAKE-FOREIGN'
  if (snapshotIssue === 'approval') root.approval_id = 'APPROVAL-FOREIGN'
  const approvalPayload: Record<string, unknown> = {
    id: APPROVAL_ID,
    title: '新建案件审批 - 新版浏览器练习甲（虚构）',
    content: '旧轮案情摘要',
    status,
    applicant_id: 9,
    applicant_name: '申请律师',
    current_approver_id: 2,
    current_approver_name: '主任',
    available_actions: ['approve', 'reject', 'request_changes'],
    workflow_type: 'CASE_APPROVAL',
    conflict_check_id: CHECK_ID,
    active_submission_snapshot_id: SNAPSHOT_ID,
    conflict_result: JSON.stringify(legacyConflict),
    metadata: {
      ...(intakeScoped ? { intake_id: INTAKE_ID } : {}),
      conflict_check_id: CHECK_ID,
      case_creation_config: { title: '旧轮案件标题', case_type: 'other' },
      materials: [{ id: 99, name: '旧轮伪当前材料', status: 'uploaded' }],
      client_name: '旧轮客户',
      parties: [{ name: '旧轮对方主体', role: 'opposing_party' }],
      conflict_record: {
        check_id: LEGACY_CHECK_ID,
        status: 'COMPLETED',
        coverage_status: legacyGreen ? 'COMPLETE' : 'LIMITED',
        case_name: '旧轮案件标题',
        case_type: 'other',
        check_result: legacyConflict,
      },
    },
  }
  // A finished approval no longer has a pending approver slot; the real
  // payload carries no current_approver_* fields, so the terminal display
  // must not depend on a fabricated approver name.
  if (['approved', 'rejected', 'completed', 'cancelled', 'canceled'].includes(status)) {
    delete approvalPayload.current_approver_id
    delete approvalPayload.current_approver_name
  }
  fetchMock.mockImplementation((input: RequestInfo | URL) => {
    const path = String(input)
    if (path.endsWith('/approvals/' + APPROVAL_ID)) {
      return Promise.resolve(jsonResponse(approvalPayload))
    }
    if (path.endsWith('/approvals/' + APPROVAL_ID + '/snapshot')) {
      return Promise.resolve(typeof snapshotIssue === 'number'
        ? failedResponse(snapshotIssue)
        : jsonResponse({
            approval_id: APPROVAL_ID,
            snapshot_id: snapshotIssue === 'pointer' ? 'OTHER-SNAPSHOT' : SNAPSHOT_ID,
            snapshot: root,
          }))
    }
    if (path.endsWith('/conflict/tasks/' + CHECK_ID + '/review')) {
      return Promise.resolve(failedResponse(403))
    }
    if (path.endsWith('/conflict/tasks/' + CHECK_ID + '/result')) {
      return Promise.resolve(jsonResponse({
        task: { taskId: CHECK_ID, checkId: CHECK_ID, intakeId: INTAKE_ID, status: 'COMPLETED' },
        result: { checkId: CHECK_ID, decision: { status: 'REVIEW_REQUIRED', coverageStatus: 'COMPLETE' } },
      }))
    }
    if (path.endsWith('/integration/approvals/' + APPROVAL_ID + '/status')) {
      return Promise.resolve(jsonResponse({
        overall_status: status === 'approved' ? 'completed' : 'pending',
        case_creation: hasCase
          ? { status: 'completed', case_id: '38', case_number: CASE_NUMBER, accessible }
          : {},
      }))
    }
    return Promise.resolve(jsonResponse({}))
  })
  render(
    <MemoryRouter initialEntries={['/approval/' + APPROVAL_ID]}>
      <Routes>
        <Route path='/approval/:id' element={<ApprovalDecisionFlow />} />
        <Route path='/case/38' element={<p>已打开申请律师的关联案件 38</p>} />
      </Routes>
    </MemoryRouter>,
  )
}

describe('Approval terminal intake snapshot access', () => {
  const originalUser = useAppStore.getState().user
  const originalAuthenticated = useAppStore.getState().isAuthenticated

  beforeEach(() => {
    fetchMock.mockReset()
    window.localStorage.clear()
  })

  afterEach(() => {
    cleanup()
    useAppStore.setState({ user: originalUser, isAuthenticated: originalAuthenticated })
  })

  it.each(['green', 'red'] as const)('shows approved-case status and unavailable retention after snapshot403, never stale %s facts', async (legacyVerdict) => {
    setupFlow({ legacyVerdict })
    await screen.findByText('已成案 ' + CASE_NUMBER)
    expect(screen.getByText(/当前账号无法读取本轮提交留存/)).toHaveTextContent('请联系项目组')
    expect(screen.getByText('成案状态 已生成正式案件')).toBeInTheDocument()
    expect(document.body).not.toHaveTextContent(LEGACY_CHECK_ID)
    expect(document.body).not.toHaveTextContent('旧轮案情摘要')
    expect(document.body).not.toHaveTextContent('旧轮案件标题')
    expect(document.body).not.toHaveTextContent('旧轮客户')
    expect(document.body).not.toHaveTextContent('旧轮对方主体')
    expect(document.body).not.toHaveTextContent('旧轮命中案件')
    expect(document.body).not.toHaveTextContent('旧轮伪当前材料')
    expect(screen.queryByText('其他')).not.toBeInTheDocument()
    expect(document.body).not.toHaveTextContent('风险评分：')
    expect(document.body).not.toHaveTextContent('未发现匹配证据')
    expect(document.body).not.toHaveTextContent('字段基线未知')
    expect(document.body).not.toHaveTextContent('审批快照暂无材料记录')
    expect(document.body).not.toHaveTextContent('已禁止同意并成案')
    expect(document.body).not.toHaveTextContent('当前不能成案')
    expect(screen.queryByRole('button', { name: /下载快照版本/ })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /同意并成案/ })).not.toBeInTheDocument()
    expect(document.body).not.toHaveTextContent('待分配')
    expect(document.body).not.toHaveTextContent('等待当前审批人处理')
    expect(screen.getAllByText('本次审批已结束').length).toBeGreaterThan(0)
  })

  it('disables the related-case entry for a director without the actual case route capability even when the server allows it', async () => {
    setupFlow({ accessible: true })
    await screen.findByText('已成案 ' + CASE_NUMBER)
    const openCase = screen.getByRole('button', { name: '查看关联案件' })
    expect(openCase).toBeDisabled()
    expect(screen.getByText('当前账号没有查看案件的权限，请联系项目组。')).toBeInTheDocument()
    fireEvent.click(openCase)
    expect(screen.queryByText('已打开申请律师的关联案件 38')).not.toBeInTheDocument()
  })

  it('does not turn an unreadable completed-case status into pending-case instructions on a terminal approval', async () => {
    setupFlow({ hasCase: false })
    await screen.findByText(/当前账号无法读取本轮提交留存/)
    expect(screen.getAllByText('审批已通过，成案信息暂不可读取').length).toBeGreaterThan(0)
    expect(document.body).not.toHaveTextContent('待生成正式案件')
    expect(document.body).not.toHaveTextContent('审批通过后生成')
    expect(document.body).not.toHaveTextContent('等待审批通过')
    expect(document.body).not.toHaveTextContent('等待当前审批人处理')
    // The terminal label 无待处理人 intentionally contains 待处理 as a
    // substring; forbid the pending-queue copy itself, not that label.
    expect(document.body.textContent).not.toMatch(/(?<!无)待处理/)
    expect(document.body).not.toHaveTextContent('已禁止同意并成案')
  })

  it.each([404, 500, 'pointer', 'intake', 'approval'] as const)('does not present another round as current when snapshot access fails with %s', async (snapshotIssue) => {
    setupFlow({ snapshotIssue, legacyVerdict: 'red' })
    await screen.findByText('已成案 ' + CASE_NUMBER)
    expect(screen.getByTestId('retention-status-line').textContent).toMatch(
      /无法读取本轮提交留存|无法确认属于本审批当前轮次/,
    )
    expect(document.body).not.toHaveTextContent(LEGACY_CHECK_ID)
    expect(document.body).not.toHaveTextContent('旧轮案情摘要')
    expect(document.body).not.toHaveTextContent('旧轮案件标题')
    expect(document.body).not.toHaveTextContent('旧轮伪当前材料')
    expect(document.body).not.toHaveTextContent('第二轮实际冻结案情')
    expect(document.body).not.toHaveTextContent('本轮冻结资料.pdf')
    expect(document.body).not.toHaveTextContent('第一轮冻结标题')
    expect(document.body).not.toHaveTextContent('字段基线未知')
    expect(document.body).not.toHaveTextContent('已禁止同意并成案')
    expect(document.body).not.toHaveTextContent('当前不能成案')
    expect(document.body).not.toHaveTextContent('审批证据不一致')
    expect(screen.queryByRole('button', { name: /下载快照版本/ })).not.toBeInTheDocument()
    expect(document.body).not.toHaveTextContent('待分配')
    expect(document.body).not.toHaveTextContent('等待当前审批人处理')
    expect(screen.getAllByText('本次审批已结束').length).toBeGreaterThan(0)
  })

  it.each(['lawyer', 'admin'])('preserves the %s route capability to open an accessible case even when approval retention is unreadable', async (role) => {
    setupFlow({ role, accessible: true })
    await screen.findByText('已成案 ' + CASE_NUMBER)
    const openCase = screen.getByRole('button', { name: '查看关联案件' })
    expect(openCase).not.toBeDisabled()
    fireEvent.click(openCase)
    expect(await screen.findByText('已打开申请律师的关联案件 38')).toBeInTheDocument()
  })

  it.each(['lawyer', 'admin'])('does not let the %s route capability override an inaccessible server case', async (role) => {
    setupFlow({ role, accessible: false })
    await screen.findByText('已成案 ' + CASE_NUMBER)
    const openCase = screen.getByRole('button', { name: '查看关联案件' })
    expect(openCase).toBeDisabled()
    expect(screen.getByText('当前账号无法查看该关联案件，请联系项目组。')).toBeInTheDocument()
    fireEvent.click(openCase)
    expect(screen.queryByText('已打开申请律师的关联案件 38')).not.toBeInTheDocument()
  })

  it('uses existing explicit case-view grants instead of hardcoding related-case roles', async () => {
    setupFlow({ role: 'director', permissions: ['case:view'], accessible: true })
    await screen.findByText('已成案 ' + CASE_NUMBER)
    expect(screen.getByRole('button', { name: '查看关联案件' })).not.toBeDisabled()
    fireEvent.click(screen.getByRole('button', { name: '查看关联案件' }))
    expect(await screen.findByText('已打开申请律师的关联案件 38')).toBeInTheDocument()
  })

  it('retains the current frozen facts, actual round diff and frozen downloads on a readable terminal approval', async () => {
    setupFlow({ snapshotIssue: null, legacyVerdict: 'red' })
    await screen.findByText('已成案 ' + CASE_NUMBER)
    expect(screen.getByText('商事')).toBeInTheDocument()
    expect(screen.getByText('第二轮实际冻结案情')).toBeInTheDocument()
    expect(screen.getByText(/本轮独立复核结论：未发现冲突/)).toHaveTextContent('检索覆盖：完整')
    expect(screen.getByTestId('approval-round-diff')).toHaveTextContent('第一轮冻结标题 → 新版浏览器练习甲（虚构）')
    expect(screen.getByTestId('approval-materials-section')).toHaveTextContent('本轮冻结资料.pdf')
    expect(screen.getByRole('button', { name: /下载快照版本/ })).not.toBeDisabled()
    expect(screen.queryByText(/当前账号无法读取本轮提交留存/)).not.toBeInTheDocument()
    expect(document.body).not.toHaveTextContent(LEGACY_CHECK_ID)
    expect(document.body).not.toHaveTextContent('当前不能成案')
    expect(document.body).not.toHaveTextContent('已禁止同意并成案')
    expect(document.body).not.toHaveTextContent('待分配')
    expect(document.body).not.toHaveTextContent('等待当前审批人处理')
    expect(screen.getAllByText('本次审批已结束').length).toBeGreaterThan(0)
  })

  it('keeps an unreadable pending intake approval closed to approval even with stale green metadata', async () => {
    setupFlow({ status: 'resubmitted', hasCase: false })
    await screen.findByText(/当前账号无法读取本轮提交留存/)
    const approve = screen.getByRole('button', { name: /同意并成案/ })
    expect(approve).toBeDisabled()
    expect(document.body).not.toHaveTextContent(LEGACY_CHECK_ID)
    expect(document.body).not.toHaveTextContent('旧轮案情摘要')
    fireEvent.click(approve)
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(fetchMock.mock.calls.some(([, options]) => options?.method === 'POST')).toBe(false)
  })

  it.each([
    ['resubmitted', 'green', 'pointer'],
    ['resubmitted', 'green', 'intake'],
    ['resubmitted', 'green', 'approval'],
    ['resubmitted', 'red', 'pointer'],
    ['resubmitted', 'red', 'intake'],
    ['resubmitted', 'red', 'approval'],
    ['needs_revision', 'green', 'pointer'],
    ['needs_revision', 'green', 'intake'],
    ['needs_revision', 'green', 'approval'],
    ['needs_revision', 'red', 'pointer'],
    ['needs_revision', 'red', 'intake'],
    ['needs_revision', 'red', 'approval'],
  ] as const)('never falls back to legacy facts when a readable pending %s snapshot has an unconfirmable %s anchor (%s)', async (status, legacyVerdict, snapshotIssue) => {
    setupFlow({ status, legacyVerdict, snapshotIssue, hasCase: false })
    await screen.findByText(/本轮提交快照无法确认/)
    expect(screen.queryByText(/本轮独立复核结论/)).not.toBeInTheDocument()
    expect(document.body).not.toHaveTextContent('旧轮案情摘要')
    expect(document.body).not.toHaveTextContent(LEGACY_CHECK_ID)
    expect(document.body).not.toHaveTextContent('旧轮伪当前材料')
    expect(document.body).not.toHaveTextContent('旧轮客户')
    expect(document.body).not.toHaveTextContent('第二轮实际冻结案情')
    expect(document.body).not.toHaveTextContent('本轮冻结资料.pdf')
    expect(screen.queryByRole('button', { name: /下载快照版本/ })).not.toBeInTheDocument()
    const approve = screen.queryByRole('button', { name: /同意并成案/ })
    if (approve) {
      expect(approve).toBeDisabled()
      fireEvent.click(approve)
    }
    expect(fetchMock.mock.calls.some(([, options]) => options?.method === 'POST')).toBe(false)
  })

  it('preserves the legacy non-intake display when its snapshot is unreadable', async () => {
    setupFlow({ intakeScoped: false, role: 'lawyer' })
    await screen.findByText('已成案 ' + CASE_NUMBER)
    expect(screen.queryByText(/当前账号无法读取本轮提交留存/)).not.toBeInTheDocument()
    expect(document.body).toHaveTextContent('旧轮案件标题')
    expect(document.body).toHaveTextContent('旧轮案情摘要')
    expect(document.body).toHaveTextContent(LEGACY_CHECK_ID)
    expect(screen.getByRole('button', { name: '查看关联案件' })).not.toBeDisabled()
  })
})
