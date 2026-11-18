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

const APPROVAL_ID = '74af5505-d006-4c67-9409-fc413ca2c184'
const ACTIVE_SNAPSHOT_ID = '9f23a879-a204-4f97-986f-48c432567360'
const NEW_CHECK_ID = 'CCT_8b28e98f-af19-42b0-a89b-de717712e2fe'
const INTAKE_ID = 'INTAKE-RESUBMIT-1'

function resubmitSnapshotRoot(structuredFacts: Record<string, unknown>) {
  return {
    approval_id: APPROVAL_ID,
    intake_id: INTAKE_ID,
    round_number: 2,
    conflict_check_id: NEW_CHECK_ID,
    conflict_review_id: 88,
    conflict_reviewer_id: 3,
    conflict_review_decision: 'no_conflict',
    conflict_review_state: 'independent_review_current',
    structured_facts: structuredFacts,
    case_creation_config: { source_snapshot_id: ACTIVE_SNAPSHOT_ID, source: 'intake_submission_snapshot' },
    material_fingerprints: [{
      material_id: 5,
      name: '客户主体资料',
      material_type: 'evidence',
      status: 'uploaded',
      current_file_id: 13,
      files: [{ file_id: 13, file_name: '新版主体资料.pdf', digest: 'bbb' }],
    }],
  }
}

function initialSnapshotRoot(structuredFacts: Record<string, unknown>) {
  // Round 1 contract (approval_resubmit_contract.go initial_submit): no
  // conflict_review_state marker, round_number 1, structured_facts still frozen.
  return {
    ...resubmitSnapshotRoot(structuredFacts),
    round_number: 1,
    conflict_review_state: undefined,
    submission_source: "initial_submit",
  }
}

function approvalPayload(metadata: Record<string, unknown>, status = 'resubmitted') {
  const terminal = ['approved', 'rejected', 'completed', 'cancelled', 'canceled'].includes(status)
  return {
    id: APPROVAL_ID,
    status,
    applicant_id: 9,
    applicant_name: '申请人甲',
    ...(terminal ? {} : { current_approver_id: 2, current_approver_name: '审批乙' }),
    available_actions: ['approve', 'reject', 'request_changes'],
    workflow_type: 'CASE_APPROVAL',
    title: '新建案件审批 - 新版浏览器练习甲（虚构）',
    content: '旧轮案情内容（不得显示）',
    conflict_check_id: NEW_CHECK_ID,
    active_submission_snapshot_id: ACTIVE_SNAPSHOT_ID,
    metadata,
  }
}

const fetchMock = jest.fn()
global.fetch = fetchMock as unknown as typeof fetch

function jsonResponse(data: unknown) {
  return { ok: true, json: async () => ({ success: true, data }) } as Response
}

function errorResponse(status: number) {
  return { ok: false, status, json: async () => ({ success: false, message: '无法读取' }) } as Response
}

function renderFlow() {
  render(
    <MemoryRouter initialEntries={['/approval/' + APPROVAL_ID]}>
      <Routes>
        <Route path='/approval/:id' element={<ApprovalDecisionFlow />} />
      </Routes>
    </MemoryRouter>,
  )
}

type SetupOptions = {
  structuredFacts?: Record<string, unknown>
  legacyMetadata?: Record<string, unknown>
  status?: string
  snapshotError?: number
  initialRound?: boolean
}

function setupFlow(options: SetupOptions = {}) {
  const {
    structuredFacts = {
      title: '新版浏览器练习甲（虚构）',
      case_type: 'commercial',
      description: '本轮主体更正后已重新检测和独立复核',
      parties: [
        { entity_name: '审核练习本轮客户（虚构）', entity_type: 'organization', party_role: 'client' },
        { entity_name: '审核练习本轮对方（虚构）', entity_type: 'organization', party_role: 'opposing_party' },
      ],
    },
    legacyMetadata = {
      intake_id: INTAKE_ID,
      client_name: '旧轮客户（不得借用）',
      parties: [{ name: '旧轮对方主体', role: 'opposing_party' }],
    },
    status = 'resubmitted',
    snapshotError,
    initialRound,
  } = options
  fetchMock.mockReset()
  fetchMock.mockImplementation((input: RequestInfo | URL) => {
    const path = String(input)
    if (path.endsWith('/approvals/' + APPROVAL_ID)) {
      return Promise.resolve(jsonResponse(approvalPayload(legacyMetadata, status)))
    }
    if (path.endsWith('/approvals/' + APPROVAL_ID + '/snapshot')) {
      if (snapshotError) return Promise.resolve(errorResponse(snapshotError))
      return Promise.resolve(jsonResponse({
        approval_id: APPROVAL_ID,
        snapshot_id: ACTIVE_SNAPSHOT_ID,
        snapshot: initialRound
          ? initialSnapshotRoot(structuredFacts)
          : resubmitSnapshotRoot(structuredFacts),
      }))
    }
    if (path.includes('/conflict/tasks/') && path.endsWith('/review')) {
      return Promise.resolve(jsonResponse({ id: 88, checkId: NEW_CHECK_ID, decision: 'no_conflict', reviewerId: 3 }))
    }
    if (path.includes('/conflict/tasks/') && path.endsWith('/result')) {
      return Promise.resolve(jsonResponse({
        task: { taskId: NEW_CHECK_ID, checkId: NEW_CHECK_ID, intakeId: INTAKE_ID, status: 'COMPLETED' },
        result: { checkId: NEW_CHECK_ID, decision: { status: 'REVIEW_REQUIRED', coverageStatus: 'COMPLETE' }, conflictCases: [] },
      }))
    }
    if (path.endsWith('/integration/approvals/' + APPROVAL_ID + '/status')) {
      return Promise.resolve(jsonResponse({ overall_status: 'pending', case_creation: {} }))
    }
    return Promise.resolve(jsonResponse({}))
  })
  renderFlow()
}

function queryApproveButton() {
  return screen.getByRole('button', { name: /同意并成案/ }) as HTMLButtonElement
}

function applicationInfo() {
  return screen.getByText('申请信息').closest('section') as HTMLElement
}

describe('approval current frozen parties', () => {
  beforeEach(() => {
    getUserInfoMock.mockReturnValue({ id: 2, role: 'director' })
    window.localStorage.clear()
  })

  it('renders both current frozen names from structured_facts and keeps approval enabled', async () => {
    setupFlow()
    await screen.findByText(/本轮独立复核结论/)
    const info = applicationInfo()
    expect(within(info).getByText('审核练习本轮客户（虚构）')).toBeInTheDocument()
    expect(within(info).getByText('审核练习本轮对方（虚构）')).toBeInTheDocument()
    expect(within(info).queryByText('旧轮客户（不得借用）')).not.toBeInTheDocument()
    expect(within(info).queryByText('旧轮对方主体')).not.toBeInTheDocument()
    await waitFor(() => expect(queryApproveButton()).not.toBeDisabled())
  })

  it('renders round-1 frozen parties from the initial_submit snapshot without the resubmit marker', async () => {
    setupFlow({ initialRound: true })
    await screen.findByText('总体风险等级', { exact: false, timeout: 3000 })
    const info = applicationInfo()
    expect(within(info).getByText('审核练习本轮客户（虚构）')).toBeInTheDocument()
    expect(within(info).getByText('审核练习本轮对方（虚构）')).toBeInTheDocument()
    expect(within(info).queryByText('旧轮客户（不得借用）')).not.toBeInTheDocument()
    expect(within(info).queryByText('旧轮对方主体')).not.toBeInTheDocument()
  })

it('never renders legacy metadata parties when the snapshot cannot be read', async () => {
  setupFlow({ snapshotError: 403 })
   await screen.findByText(/当前账号无法读取本轮提交留存|本轮提交快照无法确认属/, undefined, { timeout: 3000 })
    expect(document.body).not.toHaveTextContent('旧轮客户（不得借用）')
    expect(document.body).not.toHaveTextContent('旧轮对方主体')
 })

 it('blocks approve on a terminal approval while keeping current frozen facts visible', async () => {
   setupFlow({ status: 'approved' })
   await screen.findByText(/本轮独立复核结论/)
    await waitFor(() => {
      const approve = screen.queryByRole('button', { name: /同意并成案/ })
      if (approve) expect(approve).toBeDisabled()
    })
    const approve = screen.queryByRole('button', { name: /同意并成案/ })
    if (approve) expect(approve).toBeDisabled()
   const info = applicationInfo()
   expect(within(info).getByText('审核练习本轮对方（虚构）')).toBeInTheDocument()
 })
})
