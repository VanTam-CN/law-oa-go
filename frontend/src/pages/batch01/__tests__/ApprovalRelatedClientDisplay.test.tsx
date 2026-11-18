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
const LAWYER = { id: 2, role: 'lawyer' }

const APPROVAL_ID = 'ap-related-client-1'
const ACTIVE_SNAPSHOT_ID = 'snap-related-client-1'
const NEW_CHECK_ID = 'CCT_related_client_1'
const INTAKE_ID = 'INTAKE-RELATED-CLIENT-1'
const CURRENT_CLIENT = '六项复测A3客户甲公司'
const LEGACY_CLIENT = '旧轮客户（不得借用）'

function resubmitSnapshotRoot(overrides: Record<string, unknown> = {}) {
  return {
    approval_id: APPROVAL_ID,
    intake_id: INTAKE_ID,
    round_number: 2,
    conflict_check_id: NEW_CHECK_ID,
    conflict_review_decision: 'no_conflict',
    conflict_review_state: 'independent_review_current',
    material_fingerprints: [
      {
        material_id: 1,
        name: '委托合同',
        material_type: 'application',
        required: true,
        current_file_id: 2,
        status: 'uploaded',
        files: [{ file_id: 1, file_name: '委托合同-初稿.pdf', digest: 'aaa' }],
      },
    ],
    structured_facts: {
      title: '新版浏览器练习甲（虚构）',
      case_type: 'commercial',
      parties: [
        { party_role: 'client', entity_name: CURRENT_CLIENT, entity_type: 'company' },
        { party_role: 'opposing_party', entity_name: '对方乙公司', entity_type: 'company' },
      ],
    },
    ...overrides,
  }
}

function approvalPayload(options: { withIntakeAnchor?: boolean } = {}) {
  const metadata: Record<string, unknown> = {
    materials: [{ id: 1, name: '委托合同', material_type: 'application', status: 'uploaded' }],
  }
  if (options.withIntakeAnchor !== false) {
    metadata.intake_id = INTAKE_ID
  }
  return {
    id: APPROVAL_ID,
    status: 'resubmitted',
    applicant_id: 9,
    applicant_name: '申请人甲',
    current_approver_id: 2,
    workflow_type: 'CASE_APPROVAL',
    title: '新建案件审批 - 新版浏览器练习甲（虚构）',
    content: '旧轮内容（不得显示）',
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

function renderFlow() {
  render(
    <MemoryRouter initialEntries={['/approval/' + APPROVAL_ID + '/decision']}>
      <Routes>
        <Route path='/approval/:id/decision' element={<ApprovalDecisionFlow />} />
      </Routes>
    </MemoryRouter>,
  )
}

async function asideCard() {
  await screen.findByText('审批流程')
  return waitFor(() => {
    const el = document.querySelector('aside') as HTMLElement | null
    expect(el).not.toBeNull()
    return el as HTMLElement
  })
}

function setupFlow(options: { snapshotRoot?: Record<string, unknown>; snapshotOk?: boolean } = {}) {
  const {
    snapshotRoot = resubmitSnapshotRoot(),
    snapshotOk = true,
  } = options
  fetchMock.mockReset()
  fetchMock.mockImplementation((input: RequestInfo | URL) => {
    const path = String(input)
    if (path.endsWith('/approvals/' + APPROVAL_ID)) {
      return Promise.resolve(jsonResponse(approvalPayload()))
    }
    if (path.endsWith('/approvals/' + APPROVAL_ID + '/snapshot')) {
      if (!snapshotOk) {
        return Promise.resolve({
          ok: false,
          status: 404,
          json: async () => ({ success: false, message: '快照不存在' }),
        } as Response)
      }
      return Promise.resolve(
        jsonResponse({
          approval_id: APPROVAL_ID,
          snapshot_id: ACTIVE_SNAPSHOT_ID,
          snapshot: snapshotRoot,
        }),
      )
    }
    return Promise.resolve(jsonResponse({}))
  })
  renderFlow()
}

describe('Approval related-info client display (resubmit rounds)', () => {
  beforeEach(() => {
    getUserInfoMock.mockReturnValue(LAWYER)
    window.localStorage.clear()
  })

  it('shows the current-round frozen client instead of 未记录 in the related-info card', async () => {
    setupFlow()
    const aside = await asideCard()
    expect(within(aside).getByText('关联客户 ' + CURRENT_CLIENT)).toBeInTheDocument()
    expect(within(aside).queryByText(/关联客户 未记录/)).not.toBeInTheDocument()
  })

  it('never falls back to a legacy round client name in the related-info card', async () => {
    setupFlow()
    const aside = await asideCard()
    expect(within(aside).getByText('关联客户 ' + CURRENT_CLIENT)).toBeInTheDocument()
    expect(within(aside).queryByText(new RegExp(LEGACY_CLIENT))).not.toBeInTheDocument()
  })

  it('keeps the unreadable-snapshot placeholder when the snapshot cannot be verified', async () => {
    setupFlow({ snapshotOk: false })
    const aside = await asideCard()
    expect(within(aside).getByText('关联客户 本轮客户信息暂不可读取')).toBeInTheDocument()
    expect(within(aside).queryByText(new RegExp(CURRENT_CLIENT))).not.toBeInTheDocument()
  })

  it('keeps the legacy approval display path when no intake anchor exists', async () => {
    fetchMock.mockReset()
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const path = String(input)
      if (path.endsWith('/approvals/' + APPROVAL_ID)) {
        return Promise.resolve(
          jsonResponse({
            ...approvalPayload({ withIntakeAnchor: false }),
            active_submission_snapshot_id: undefined,
          }),
        )
      }
      if (path.endsWith('/approvals/' + APPROVAL_ID + '/snapshot')) {
        return Promise.resolve(
          jsonResponse({
            approval_id: APPROVAL_ID,
            snapshot_id: ACTIVE_SNAPSHOT_ID,
            snapshot: { client_name: LEGACY_CLIENT },
          }),
        )
      }
      return Promise.resolve(jsonResponse({}))
    })
    renderFlow()
    const aside = await asideCard()
    expect(within(aside).getByText('关联客户 ' + LEGACY_CLIENT)).toBeInTheDocument()
    expect(within(aside).queryByText(/关联客户 本轮快照未记录/)).not.toBeInTheDocument()
  })
})
