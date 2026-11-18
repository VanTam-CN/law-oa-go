import React from 'react'
import { render, screen, waitFor, within } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router'
import { getUserInfo } from '@/utils/storage'
import { ApprovalDecisionFlow } from '../Batch01Prototype'

jest.mock('@/utils/storage', () => ({
  ...jest.requireActual('@/utils/storage'),
  getRoles: jest.fn(() => []),
  getToken: jest.fn(() => 'test-token'),
  getUserInfo: jest.fn(() => ({ id: 2, role: 'director' })),
}))

jest.mock('@/utils/messageHelper', () => ({
  message: { error: jest.fn(), success: jest.fn(), warning: jest.fn() },
  setAppMessage: jest.fn(),
}))

const getUserInfoMock = getUserInfo as jest.Mock

// A lawyer (not a conflict-review role): the server redacts legacy evidence
// payloads for this viewer, but the approval anchor columns stay current.
const LAWYER = { id: 2, role: 'lawyer' }

// Real anchors from the live resubmitted approval round.
const APPROVAL_ID = '74af5505-d006-4c67-9409-fc413ca2c184'
const ACTIVE_SNAPSHOT_ID = '9f23a879-a204-4f97-986f-48c432567360'
const NEW_CHECK_ID = 'CCT_8b28e98f-af19-42b0-a89b-de717712e2fe'
const LEGACY_CHECK_ID = 'CCT_892-legacy-payload'
const INTAKE_ID = 'INTAKE-RESUBMIT-1'
const PREVIOUS_SNAPSHOT_ID = '7abd7e3a-legacy-round1'
const STALE_REVIEWER_ID = 3

// Round-2 slim root, matching the resubmit contract in
// internal/services/approval_resubmit_contract.go: only the current
// check/review anchors plus structured_facts are frozen. No
// conflict_check_status, no conflict_coverage_status, no snapshot_type.
function resubmitSnapshotRoot(overrides: Record<string, unknown> = {}) {
  return {
    approval_id: APPROVAL_ID,
    intake_id: INTAKE_ID,
    round_number: 2,
    previous_snapshot_id: PREVIOUS_SNAPSHOT_ID,
    conflict_check_id: NEW_CHECK_ID,
    conflict_review_id: 88,
    conflict_reviewer_id: STALE_REVIEWER_ID,
    conflict_review_decision: 'no_conflict',
    conflict_review_state: 'independent_review_current',
    structured_facts: {
      title: '新版浏览器练习甲（虚构）',
      case_type: 'commercial',
      description: '第二轮重提后的真实案情描述（已按退回意见修改）',
    },
    case_creation_config: {
      source_snapshot_id: ACTIVE_SNAPSHOT_ID,
      source: 'intake_submission_snapshot',
    },
    material_fingerprints: [
      {
        material_id: 5,
        name: '客户主体资料',
        material_type: 'evidence',
        required: true,
        current_file_id: 13,
        status: 'uploaded',
        files: [
          { file_id: 12, file_name: '旧版主体资料.pdf', digest: 'aaa' },
          { file_id: 13, file_name: '新版主体资料.pdf', digest: 'bbb' },
        ],
      },
    ],
    ...overrides,
  }
}

// The GET /approvals/:id payload still carries legacy round-1 payloads in
// metadata.conflict_record / conflict_result. The anchor columns are always
// the current round's ids.
function approvalPayload(options: { metadataConflictRecord?: boolean; legacyConflictResult?: boolean } = {}) {
  const metadata: Record<string, unknown> = {
    intake_id: INTAKE_ID,
    materials: [{ id: 5, name: '客户主体资料', material_type: 'evidence', status: 'uploaded' }],
  }
  if (options.metadataConflictRecord) {
    metadata.conflict_record = {
      check_id: LEGACY_CHECK_ID,
      status: 'COMPLETED',
      case_name: '旧轮案件名（不得借用）',
      case_type: '其他',
      client_name: '旧轮客户（不得借用）',
      coverage_status: 'LIMITED',
      check_result: {
        decision: {
          status: 'BLOCKED',
          evidenceCount: 2,
          restrictedCount: 0,
          coverageStatus: 'COMPLETE',
        },
        review: { decision: 'confirmed_conflict' },
      },
    }
  }
  return {
    id: APPROVAL_ID,
    status: 'resubmitted',
    applicant_id: 9,
    applicant_name: '申请人甲',
    current_approver_id: 2,
    current_approver_name: '审批乙',
    available_actions: ['approve', 'reject', 'request_changes'],
    priority: 'high',
    workflow_type: 'CASE_APPROVAL',
    title: '新建案件审批 - 新版浏览器练习甲（虚构）',
    content: '第一轮旧案情内容（不得显示为步骤1）',
    conflict_check_id: NEW_CHECK_ID,
    active_submission_snapshot_id: ACTIVE_SNAPSHOT_ID,
    metadata,
    ...(options.legacyConflictResult
      ? {
          conflict_result: JSON.stringify({
            checkId: LEGACY_CHECK_ID,
            coverage_status: 'COMPLETE',
            riskAssessment: { overallRisk: 'LOW', riskScore: 5 },
            decision: { status: 'CLEAR', evidenceCount: 0, restrictedCount: 0 },
            review: { decision: 'no_conflict' },
          }),
        }
      : {}),
  }
}

const fetchMock = jest.fn()
global.fetch = fetchMock as unknown as typeof fetch

function jsonResponse(data: unknown) {
  return { ok: true, json: async () => ({ success: true, data }) } as Response
}

function errorResponse() {
  return {
    ok: false,
    status: 404,
    json: async () => ({ success: false, message: '冲突检测记录不存在' }),
  } as Response
}

function renderFlow() {
  render(
    <MemoryRouter initialEntries={['/approval/' + APPROVAL_ID + '/decision']}>
      <Routes>
        <Route path='/approval/:id/decision' element={<ApprovalDecisionFlow />} />
      </Routes>
    </MemoryRouter>,
  )
}

// The page legitimately renders prior-round content inside the audit card
// and the approval-record card; step-1 assertions must therefore stay
// scoped to the batch-approval-steps container.
function batchSteps() {
  const el = document.querySelector('.batch-approval-steps') as HTMLElement | null
  expect(el).not.toBeNull()
  return el as HTMLElement
}

// Async variant for load-time assertions: waits for the approval page (and
// its steps container) before scoping the query.
async function findBatchSteps() {
  await screen.findByText('审批流程')
  return waitFor(() => {
    const el = document.querySelector('.batch-approval-steps') as HTMLElement | null
    expect(el).not.toBeNull()
    return el as HTMLElement
  })
}

// antd injects an icon span before the label, so the accessible name of the
// approve button is icon text + 同意并成案 — match it with a regexp.
function queryApproveButton() {
  return screen.getByRole('button', { name: /同意并成案/ }) as HTMLButtonElement
}

function setupFlow(
  options: {
    metadataConflictRecord?: boolean
    legacyConflictResult?: boolean
    snapshotRoot?: Record<string, unknown>
    reviewPayload?: Record<string, unknown> | null
    resultMode?: 'ok' | 'no-coverage' | 'limited' | 'not-found' | 'foreign-check-id'
  } = {},
) {
  const {
    snapshotRoot = resubmitSnapshotRoot(),
    reviewPayload = { id: 88, checkId: NEW_CHECK_ID, decision: 'no_conflict', reviewerId: STALE_REVIEWER_ID },
    resultMode = 'ok',
    ...approvalOptions
  } = options
  fetchMock.mockReset()
  fetchMock.mockImplementation((input: RequestInfo | URL) => {
    const path = String(input)
    if (path.endsWith('/approvals/' + APPROVAL_ID)) {
      return Promise.resolve(jsonResponse(approvalPayload(approvalOptions)))
    }
    if (path.endsWith('/approvals/' + APPROVAL_ID + '/snapshot')) {
      return Promise.resolve(
        jsonResponse({
          approval_id: APPROVAL_ID,
          snapshot_id: ACTIVE_SNAPSHOT_ID,
          snapshot: snapshotRoot,
        }),
      )
    }
    if (path.includes('/conflict/tasks/') && path.endsWith('/review')) {
      return Promise.resolve(reviewPayload === null ? errorResponse() : jsonResponse(reviewPayload))
    }
    // GET /conflict/tasks/:id/result — real response shape confirmed live:
    // data.task + data.result, coverage only at result.decision.coverageStatus.
    if (path.includes('/conflict/tasks/') && path.endsWith('/result')) {
      if (resultMode === 'not-found') {
        return Promise.resolve(errorResponse())
      }
      const resultCheckId = resultMode === 'foreign-check-id' ? LEGACY_CHECK_ID : NEW_CHECK_ID
      const coverageStatus =
        resultMode === 'no-coverage'
          ? undefined
          : resultMode === 'limited'
            ? 'LIMITED'
            : 'COMPLETE'
      return Promise.resolve(
        jsonResponse({
          task: {
            taskId: resultCheckId,
            checkId: resultCheckId,
            intakeId: INTAKE_ID,
            status: 'COMPLETED',
          },
          result: {
            checkId: resultCheckId,
            decision: {
              status: 'REVIEW_REQUIRED',
              ...(coverageStatus ? { coverageStatus } : {}),
            },
            conflictCases: [],
          },
        }),
      )
    }
    if (path.endsWith('/integration/approvals/' + APPROVAL_ID + '/status')) {
      return Promise.resolve(jsonResponse({ overall_status: 'pending', case_creation: {} }))
    }
    return Promise.resolve(jsonResponse({}))
  })
  renderFlow()
}

describe('Approval resubmitted slim-round frozen context', () => {
  beforeEach(() => {
    getUserInfoMock.mockReturnValue(LAWYER)
    window.localStorage.clear()
  })

  // Shared render gate: the strict slim context must be active before any
  // button/coverage assertions in these tests mean anything.
  async function expectResubmitSummaryActive() {
    // The frozen-review sentence renders only in the strict slim branch;
    // the round-diff card also renders on legacy paths and is not a marker.
    await screen.findByText(/本轮独立复核结论/)
  }

  it('renders this round frozen facts and anchors without borrowing legacy round-1 payload', async () => {
    setupFlow({ metadataConflictRecord: true })
    expect(await screen.findByText('新版浏览器练习甲（虚构）')).toBeInTheDocument()
    await expectResubmitSummaryActive()
    expect(screen.getByText('商事')).toBeInTheDocument()
    expect(screen.getByText(/第二轮重提后的真实案情描述/)).toBeInTheDocument()
    const audit = screen.getByText('查看审批审计信息').closest(
      '.batch-diagnostic-details',
    ) as HTMLElement
    within(audit).getByText(/检测记录编号：/)
    within(audit).getByText(/CCT_8b28e98f/)
    // Coverage comes from the current check record (same id): COMPLETE.
    expect(screen.getByText(/检索覆盖：完整/)).toBeInTheDocument()
    const approveBtn = queryApproveButton()
    expect(approveBtn).not.toBeDisabled()
    // Legacy number, legacy round-1 description, legacy case name/client.
    expect(screen.queryByText(/CCT_892/)).not.toBeInTheDocument()
    // The legacy round-1 description may legitimately appear inside the
    // approval-record history card, but never as this round's step-1 copy.
    expect(within(batchSteps()).queryByText(/第一轮旧案情内容/)).not.toBeInTheDocument()
    expect(screen.queryByText(/旧轮案件名/)).not.toBeInTheDocument()
    expect(screen.queryByText(/旧轮客户/)).not.toBeInTheDocument()
    // The info row renders 关联案件 label and value in separate elements.
    expect(screen.getByText('新版浏览器练习甲（虚构）')).toBeInTheDocument()
    expect(screen.getByText('商事')).toBeInTheDocument()
  })

  it('does not present legacy green evidence as this round zero-hit result', async () => {
    setupFlow({ legacyConflictResult: true })
    expect(await screen.findByText('新版浏览器练习甲（虚构）')).toBeInTheDocument()
    await expectResubmitSummaryActive()
    expect(screen.queryByText(/未发现匹配证据/)).not.toBeInTheDocument()
    expect(screen.getByText(/本轮快照未记录检索证据/)).toBeInTheDocument()
    expect(screen.queryByText(/CCT_892/)).not.toBeInTheDocument()
    const approveBtn = queryApproveButton()
    expect(approveBtn).not.toBeDisabled()
  })

  it('keeps the frozen no_conflict decision decisive over legacy blocked metadata', async () => {
    setupFlow({ metadataConflictRecord: true })
    expect(await screen.findByText('新版浏览器练习甲（虚构）')).toBeInTheDocument()
    await expectResubmitSummaryActive()
    // The legacy payload here is a BLOCKED/confirmed_conflict record: the
    // current frozen no_conflict must stay decisive in BOTH directions.
    expect(screen.getByText(/总体风险等级：已复核·无冲突/)).toBeInTheDocument()
    expect(screen.queryByText(/暂停接案/)).not.toBeInTheDocument()
    const approveBtn = queryApproveButton()
    expect(approveBtn).not.toBeDisabled()
    expect(screen.queryByText(/旧轮案件名/)).not.toBeInTheDocument()
  })

  it('keeps the slim gate on frozen review + anchored coverage when the review API is unreadable and legacy metadata says blocked', async () => {
    // The frozen root says no_conflict, the anchored current check proves
    // COMPLETE coverage, but the review endpoint 404s and the legacy
    // metadata carries a confirmed_conflict/LIMITED verdict. The slim gate
    // must use ONLY this round's frozen review + anchored coverage.
    setupFlow({ reviewPayload: null, metadataConflictRecord: true })
    await expectResubmitSummaryActive()
    expect(screen.getByText(/本轮独立复核结论：未发现冲突（已冻结）/)).toBeInTheDocument()
    expect(screen.getByText(/检索覆盖：完整/)).toBeInTheDocument()
    // No legacy "needs human review" copy may leak into the slim summary.
    expect(screen.queryByText(/待人工复核/)).not.toBeInTheDocument()
    expect(screen.queryByText(/暂停接案/)).not.toBeInTheDocument()
    expect(queryApproveButton()).not.toBeDisabled()
    expect(screen.queryByText(/旧轮案件名/)).not.toBeInTheDocument()
  })

  it('stays blocked on an unanchored slim payload even when legacy metadata says blocked', async () => {
    // Same legacy-blocked environment, but the snapshot intake anchor does
    // not match the approval: nothing about this round is proven, so the
    // page fails closed instead of trusting either payload.
    setupFlow({
      reviewPayload: null,
      metadataConflictRecord: true,
      snapshotRoot: resubmitSnapshotRoot({ intake_id: 'INTAKE-OTHER' }),
    })
    expect(await screen.findByText(/本轮提交快照无法确认/)).toBeInTheDocument()
    expect(screen.queryByText(/本轮独立复核结论/)).not.toBeInTheDocument()
    const steps = await findBatchSteps()
    expect(within(steps).queryByText(/第一轮旧案情内容/)).not.toBeInTheDocument()
    expect(screen.queryByText(/旧轮案件名/)).not.toBeInTheDocument()
    expect(screen.queryByText(/旧轮客户/)).not.toBeInTheDocument()
    expect(queryApproveButton()).toBeDisabled()
    expect(screen.getByText(/本轮提交快照锚点校验未通过，已禁止同意并成案/)).toBeInTheDocument()
  })

  it('blocks approval and shows unknown copy when this round froze no review decision', async () => {
    setupFlow({ snapshotRoot: resubmitSnapshotRoot({ conflict_review_decision: '' }) })
    expect(await screen.findByText('新版浏览器练习甲（虚构）')).toBeInTheDocument()
    await expectResubmitSummaryActive()
    const approveBtn = queryApproveButton()
    expect(approveBtn).toBeDisabled()
    expect(screen.getByText(/本轮独立复核结论：本轮快照未记录/)).toBeInTheDocument()
    expect(screen.getByText(/总体风险等级：待人工复核/)).toBeInTheDocument()
  })

  it('fails closed when the current check result is missing coverage', async () => {
    setupFlow({ resultMode: 'no-coverage' })
    expect(await screen.findByText('新版浏览器练习甲（虚构）')).toBeInTheDocument()
    await expectResubmitSummaryActive()
    const approveBtn = queryApproveButton()
    expect(approveBtn).toBeDisabled()
    expect(screen.getByText(/检索覆盖：本轮快照未记录/)).toBeInTheDocument()
  })

  it('fails closed when the current check result coverage is LIMITED', async () => {
    setupFlow({ resultMode: 'limited' })
    expect(await screen.findByText('新版浏览器练习甲（虚构）')).toBeInTheDocument()
    await expectResubmitSummaryActive()
    const approveBtn = queryApproveButton()
    expect(approveBtn).toBeDisabled()
    expect(screen.getByText(/检索覆盖：受限/)).toBeInTheDocument()
  })

  it('fails closed when the current check result cannot be read (404)', async () => {
    setupFlow({ resultMode: 'not-found' })
    expect(await screen.findByText('新版浏览器练习甲（虚构）')).toBeInTheDocument()
    await expectResubmitSummaryActive()
    const approveBtn = queryApproveButton()
    expect(approveBtn).toBeDisabled()
    expect(screen.getByText(/检索覆盖：本轮快照未记录/)).toBeInTheDocument()
  })

  it('fails closed when the result endpoint returns a foreign check id; legacy green cannot unlock', async () => {
    setupFlow({ resultMode: 'foreign-check-id', legacyConflictResult: true })
    expect(await screen.findByText('新版浏览器练习甲（虚构）')).toBeInTheDocument()
    await expectResubmitSummaryActive()
    const approveBtn = queryApproveButton()
    expect(approveBtn).toBeDisabled()
    expect(screen.getByText(/检索覆盖：本轮快照未记录/)).toBeInTheDocument()
    expect(screen.queryByText(/未发现匹配证据/)).not.toBeInTheDocument()
  })

  it('shows an unconfirmable snapshot and never legacy facts when the snapshot intake anchor is wrong', async () => {
    setupFlow({ snapshotRoot: resubmitSnapshotRoot({ intake_id: 'INTAKE-OTHER' }) })
    expect(await screen.findByText(/本轮提交快照无法确认/)).toBeInTheDocument()
    expect(screen.queryByText(/本轮独立复核结论/)).not.toBeInTheDocument()
    // Nothing about this round is proven: the legacy round-1 payload must
    // not appear as current facts anywhere on the page.
    const steps = await findBatchSteps()
    expect(within(steps).queryByText(/第一轮旧案情内容/)).not.toBeInTheDocument()
    expect(screen.queryByText(/旧轮案件名/)).not.toBeInTheDocument()
    expect(screen.queryByText(/旧轮客户/)).not.toBeInTheDocument()
    // The anchor failure disables APPROVAL (case creation must not ride an
    // unproven snapshot) and names the reason; reject/return stay available
    // because they do not advance the round and stay pinned to the active
    // snapshot id by decideApproval.
    expect(queryApproveButton()).toBeDisabled()
    expect(screen.getByText(/本轮提交快照锚点校验未通过，已禁止同意并成案/)).toBeInTheDocument()
  })

  it('shows an unconfirmable snapshot and never legacy facts when the snapshot approval anchor is wrong', async () => {
    setupFlow({ snapshotRoot: resubmitSnapshotRoot({ approval_id: 'OTHER-APPROVAL' }) })
    expect(await screen.findByText(/本轮提交快照无法确认/)).toBeInTheDocument()
    expect(screen.queryByText(/本轮独立复核结论/)).not.toBeInTheDocument()
    const steps = await findBatchSteps()
    expect(within(steps).queryByText(/第一轮旧案情内容/)).not.toBeInTheDocument()
    expect(screen.queryByText(/旧轮案件名/)).not.toBeInTheDocument()
    expect(screen.queryByText(/旧轮客户/)).not.toBeInTheDocument()
    expect(queryApproveButton()).toBeDisabled()
    expect(screen.getByText(/本轮提交快照锚点校验未通过，已禁止同意并成案/)).toBeInTheDocument()
  })

  it('shows an unconfirmable snapshot and never legacy facts when the active pointer does not verify', async () => {
    fetchMock.mockReset()
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const path = String(input)
      if (path.endsWith('/approvals/' + APPROVAL_ID)) {
        return Promise.resolve(jsonResponse(approvalPayload()))
      }
      if (path.endsWith('/approvals/' + APPROVAL_ID + '/snapshot')) {
        return Promise.resolve(
          jsonResponse({
            approval_id: APPROVAL_ID,
            // Envelope id diverges from approval.active_submission_snapshot_id.
            snapshot_id: 'snapshot-stale',
            snapshot: resubmitSnapshotRoot(),
          }),
        )
      }
      if (path.endsWith('/integration/approvals/' + APPROVAL_ID + '/status')) {
        return Promise.resolve(jsonResponse({ overall_status: 'pending', case_creation: {} }))
    }
    return Promise.resolve(jsonResponse({}))
  })
  renderFlow()
    expect(await screen.findByText(/本轮提交快照无法确认/)).toBeInTheDocument()
    expect(screen.queryByText(/本轮独立复核结论/)).not.toBeInTheDocument()
    const steps = await findBatchSteps()
    expect(within(steps).queryByText(/第一轮旧案情内容/)).not.toBeInTheDocument()
    expect(screen.queryByText(/旧轮案件名/)).not.toBeInTheDocument()
    expect(screen.queryByText(/旧轮客户/)).not.toBeInTheDocument()
    expect(queryApproveButton()).toBeDisabled()
    expect(screen.getByText(/本轮提交快照锚点校验未通过，已禁止同意并成案/)).toBeInTheDocument()
  })

  it('blocks with the evidence-mismatch gate when the snapshot check anchor diverges', async () => {
    setupFlow({ snapshotRoot: resubmitSnapshotRoot({ conflict_check_id: 'CCT_other-check' }) })
    expect(await screen.findByText('审批证据不一致')).toBeInTheDocument()
    // With the mismatch gate on, the action bar shows the read-only reason
    // instead of any decision buttons.
    expect(screen.queryByRole('button', { name: /同意并成案/ })).not.toBeInTheDocument()
    expect(screen.getByText('审批证据不一致，已禁止处理')).toBeInTheDocument()
  })

  it('keeps legacy non-intake flows on the original display path', async () => {
    fetchMock.mockReset()
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const path = String(input)
      if (path.endsWith('/approvals/' + APPROVAL_ID)) {
        return Promise.resolve(
          jsonResponse({
            ...approvalPayload(),
            metadata: {},
            title: '传统审批（无接案绑定）',
            content: '传统审批申请内容',
          }),
        )
      }
      if (path.endsWith('/approvals/' + APPROVAL_ID + '/snapshot')) {
        return Promise.resolve(
          jsonResponse({
            approval_id: APPROVAL_ID,
            snapshot_id: ACTIVE_SNAPSHOT_ID,
            snapshot: resubmitSnapshotRoot(),
          }),
        )
      }
      if (path.endsWith('/integration/approvals/' + APPROVAL_ID + '/status')) {
        return Promise.resolve(jsonResponse({ overall_status: 'pending', case_creation: {} }))
      }
      return Promise.resolve(jsonResponse({}))
    })
    renderFlow()
    const legacyTitleMatches = await screen.findAllByText('传统审批（无接案绑定）')
    expect(legacyTitleMatches.length).toBeGreaterThanOrEqual(2)
    expect(screen.queryByText(/本轮独立复核结论/)).not.toBeInTheDocument()
    // Without approval.metadata.intake_id the strict anchor gate is off, so
    // the slim branch must not activate; the legacy title path renders.
    // Legacy info rows keep the title in the header and the 关联案件 row.
    expect(screen.getAllByText('传统审批（无接案绑定）').length).toBeGreaterThanOrEqual(2)
  })

  it('shows round-1 frozen facts via the legacy chain without blocking on coverage', async () => {
    const round1Root = {
      approval_id: APPROVAL_ID,
      intake_id: INTAKE_ID,
      round_number: 1,
      conflict_check_id: NEW_CHECK_ID,
      conflict_check_status: 'COMPLETED',
      conflict_coverage_status: 'COMPLETE',
      conflict_review_id: 7,
      conflict_review_decision: 'no_conflict',
      conflict_reviewer_id: 3,
      // Round-1 roots freeze status/coverage fields but NOT the
      // independent_review_current marker; they keep the legacy display path.
      structured_facts: {
        title: '新版浏览器练习甲（虚构）',
        case_type: 'commercial',
        description: '第一轮案情',
      },
      material_fingerprints: [],
    }
    fetchMock.mockReset()
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const path = String(input)
      if (path.endsWith('/approvals/' + APPROVAL_ID)) {
        return Promise.resolve(jsonResponse(approvalPayload()))
      }
      if (path.endsWith('/approvals/' + APPROVAL_ID + '/snapshot')) {
        return Promise.resolve(
          jsonResponse({
            approval_id: APPROVAL_ID,
            snapshot_id: ACTIVE_SNAPSHOT_ID,
            snapshot: round1Root,
          }),
        )
      }
      if (path.endsWith('/integration/approvals/' + APPROVAL_ID + '/status')) {
        return Promise.resolve(jsonResponse({ overall_status: 'pending', case_creation: {} }))
      }
      return Promise.resolve(jsonResponse({}))
    })
    renderFlow()
    // The round-1 title only appears inside the composed header string via
    // the legacy chain, so match it with a regexp.
    const round1TitleMatches = await screen.findAllByText(/新版浏览器练习甲（虚构）/)
    expect(round1TitleMatches.length).toBeGreaterThanOrEqual(1)
    expect(screen.queryByText(/本轮独立复核结论/)).not.toBeInTheDocument()
    // Round-1 description renders via the legacy chain, current approval
    // content stays the step-1 copy, and coverage does NOT block approval.
    expect(within(batchSteps()).getByText(/第一轮旧案情内容/)).toBeInTheDocument()
    expect(screen.queryByText(/待人工复核/)).not.toBeInTheDocument()
    expect(queryApproveButton()).not.toBeDisabled()
  })
})
