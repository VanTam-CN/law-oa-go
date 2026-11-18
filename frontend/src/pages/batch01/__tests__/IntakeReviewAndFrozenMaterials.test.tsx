import React from 'react'
import { act, cleanup, configure, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router'
import { ApprovalDecisionFlow, CaseIntakeWorkbench, scopedCaseIntakeDraftKey } from '../Batch01Prototype'

jest.mock('@/utils/storage', () => ({
  getRoles: jest.fn(() => []),
  getToken: jest.fn(() => 'frozen-materials-token'),
  getUserInfo: jest.fn(() => ({ id: 7, role: 'lawyer', roles: ['lawyer'], name: '律师甲' })),
}))
jest.mock('@/utils/messageHelper', () => ({
  message: { error: jest.fn(), success: jest.fn(), warning: jest.fn(), info: jest.fn() },
  setAppMessage: jest.fn(),
}))

configure({ asyncUtilTimeout: 4000 })

const fetchMock = jest.fn()
global.fetch = fetchMock as unknown as typeof fetch
const json = (data: unknown) => ({ ok: true, status: 200, json: async () => ({ success: true, data }) }) as Response

let activePointer = 'snapshot-round-1'
const resubmitPointer = 'snapshot-round-2'

const roundOneFingerprint = {
  material_id: 301,
  material_type: 'contract',
  name: '服务合同',
  required: true,
  status: 'uploaded',
  current_file_id: 3001,
  files: [{ file_id: 3001, file_name: '服务合同-v1.pdf' }],
}

const roundTwoFingerprint = {
  ...roundOneFingerprint,
  current_file_id: 3002,
  files: [
    { file_id: 3001, file_name: '服务合同-v1.pdf' },
    { file_id: 3002, file_name: '服务合同-v2.pdf' },
  ],
}

let snapshot: Record<string, unknown> = {}
let approval: Record<string, unknown> = {}
let continuationProof: Record<string, unknown> = {}
let continuationFailure = false

// Mirrors the real PG contract (GET /approvals/:id/snapshot projects the raw
// approval_snapshots.snapshot_data): the payload has NO snapshot_type — that
// value lives only on the DB column and is never returned by the endpoint.
// Frozen intake_submission content is therefore recognized by anchors only:
// payload intake_id/approval_id must match approval.metadata.intake_id and
// the routed approval id, on top of the verified active pointer.
function frozenSnapshot(approvalID: string, intakeID: string, fingerprints: unknown[]) {
  return {
    approval_id: approvalID,
    intake_id: intakeID,
    intake_revision: 'rev-1',
    intake_title: '股权争议',
    intake_status: 'conflict_ready',
    round_number: 1,
    previous_snapshot_id: '',
    submitted_by: 7,
    submitted_at: '2026-10-02T10:00:00Z',
    conflict_check_id: 'CHECK-1',
    conflict_check_status: 'COMPLETED',
    conflict_coverage_status: 'COMPLETE',
    lawyer_facts_version: 'fv1',
    client_id: 42,
    applicant_id: 42,
    applicant_name: '示例客户',
    submission_source: 'web',
    structured_facts: { facts_version: 'fv1' },
    case_creation_config: {},
    material_fingerprints: fingerprints,
  }
}

function continuation(intakeID: string, revision: string, canSubmit: boolean) {
  return {
    intake_id: intakeID,
    revision,
    status: 'conflict_ready',
    can_reuse_identity: true,
    can_reuse_check: true,
    can_submit: canSubmit,
    recheck_required: false,
    reasons: canSubmit ? [] : ['REVIEW_MISSING'],
    facts: { facts_version: 'fv1', confirmed_at: '2026-10-02T10:00:00Z' },
    conflict_binding: {
      check_id: 'CHECK-1',
      check_status: 'COMPLETED',
      coverage_status: 'COMPLETE',
      ...(canSubmit ? { review_id: 11, reviewer_id: 9, review_decision: 'no_conflict' } : {}),
    },
    approval: { approval_id: 'A-REVIEW-1', snapshot_id: activePointer, status: 'needs_revision', round: 1 },
  }
}

function mockApprovalAPIs() {
  // @ts-expect-error jest 30 mock typing (same pattern as sibling suites)
  fetchMock.mockImplementation((input: RequestInfo | URL) => {
    const path = String(input)
    if (path.endsWith('/approvals/A-REVIEW-1')) {
      return json({
        id: 'A-REVIEW-1',
        type: 'case_creation',
        status: approval.status || 'needs_revision',
        metadata: { intake_id: 'I-REVIEW-1', ...(approval.metadata || {}) },
        active_submission_snapshot_id: activePointer,
        records: approval.records || [],
      })
    }
    if (path.endsWith('/approvals/A-REVIEW-1/snapshot')) {
      return json({ approval_id: 'A-REVIEW-1', snapshot_id: activePointer, snapshot })
    }
    if (path.endsWith('/continuation')) {
      if (continuationFailure) return Promise.reject(new Error('offline'))
      return json(continuationProof)
    }
    if (path.endsWith('/case-intakes/I-REVIEW-1')) {
      return json({
        intake: {
          id: 'I-REVIEW-1',
          intake_code: 'INT-1',
          title: '股权争议',
          case_type: 'commercial',
          description: '本轮案情摘要',
          client_id: 42,
          created_by: '7',
          metadata: { lawyer_id: 7, business_area: '投资并购', sub_area: '合同' },
        },
        revision: 'rev-1',
        parties: [
          { entity_name: '示例客户', entity_type: 'LEGAL_PERSON', party_role: 'client' },
          { entity_name: '虚构对方', entity_type: 'LEGAL_PERSON', party_role: 'opposing_party' },
        ],
      })
    }
    if (path.includes('/clients?')) {
      return json({ clients: [{ id: 42, name: '示例客户', type: '企业' }] })
    }
    if (path.includes('/lawfirm/lawyers?')) {
      return json({ lawyers: [{ id: 7, name: '律师甲' }] })
    }
    if (path.includes('/case-materials/301/files/3001/download')) {
      return Promise.resolve({ ok: true, blob: async () => new Blob(['round-one']) } as Response)
    }
    if (path.includes('/case-materials/301/files/3002/download')) {
      return Promise.resolve({ ok: true, blob: async () => new Blob(['round-two']) } as Response)
    }
    return json({})
  })
}

async function renderApproval() {
  await renderApprovalExpecting('审批材料')
}

// Pending approvals with an unconfirmable snapshot anchor render the
// unconfirmable-retention card instead of the frozen materials section, so
// the fail-closed suites must wait on that card rather than 审批材料.
async function renderApprovalExpecting(expected: '审批材料' | 'unconfirmable') {
  render(
    <MemoryRouter initialEntries={['/approval/A-REVIEW-1/decision']}>
      <Routes>
        <Route path='/approval/:id/decision' element={<ApprovalDecisionFlow />} />
      </Routes>
    </MemoryRouter>,
  )
  await (expected === '审批材料'
    ? screen.findByText('审批材料')
    : screen.findByText(/本轮提交快照无法确认属于本审批当前轮次/))
  await act(async () => {})
}

async function openRevisionWorkbench() {
  window.localStorage.setItem(
    scopedCaseIntakeDraftKey('7'),
    JSON.stringify({
      intakeId: 'I-REVIEW-1',
      intakeCode: 'INT-1',
      title: '股权争议',
      caseType: 'commercial',
      businessArea: '投资并购',
      clientId: 42,
      clientName: '示例客户',
      lawyerId: 7,
      revision: 'rev-1',
    }),
  )
  render(
    <MemoryRouter initialEntries={['/case/create?mode=revision&intake_id=I-REVIEW-1&approval_id=A-REVIEW-1']}>
      <CaseIntakeWorkbench />
    </MemoryRouter>,
  )
  await act(async () => {})
}

describe('Intake continuation review copy and frozen approval materials', () => {
  beforeEach(() => {
    fetchMock.mockReset()
    window.localStorage.clear()
    window.sessionStorage.clear()
    snapshot = {}
    approval = { status: 'needs_revision', records: [{ id: 'R1', decision: 'request_changes', imposed_requirements: { items: [] } }] }
    continuationProof = {}
    continuationFailure = false
    activePointer = 'snapshot-round-1'
    mockApprovalAPIs()
  })

  it('renders only fingerprint rows for a strict anchored intake submission and downloads the frozen current file', async () => {
    snapshot = frozenSnapshot('A-REVIEW-1', 'I-REVIEW-1', [roundOneFingerprint])
    await renderApproval()
    expect(screen.getByTestId('approval-material-row-301')).toBeInTheDocument()
    expect(screen.getByTestId('approval-material-row-name-301')).toHaveTextContent('服务合同')
    expect(screen.getByTestId('approval-material-row-type-301')).toHaveTextContent('合同')
    expect(screen.getByText('服务合同-v1.pdf')).toBeInTheDocument()
    await act(async () => { fireEvent.click(screen.getByRole('button', { name: '下载快照版本' })) })
    await waitFor(() => {
      expect(fetchMock.mock.calls.some(([input]) =>
        String(input).includes('/case-materials/301/files/3001/download'),
      )).toBe(true)
    })
  })

  it('shows the newly frozen round after resubmission pointer changes and downloads only that current file', async () => {
    snapshot = frozenSnapshot('A-REVIEW-1', 'I-REVIEW-1', [roundOneFingerprint])
    await renderApproval()
    expect(screen.getByText('服务合同-v1.pdf')).toBeInTheDocument()
    activePointer = 'snapshot-round-2'
    snapshot = frozenSnapshot('A-REVIEW-1', 'I-REVIEW-1', [roundTwoFingerprint])
    cleanup()
    await renderApproval()
    expect(screen.getByText('服务合同-v2.pdf')).toBeInTheDocument()
    await act(async () => { fireEvent.click(screen.getByRole('button', { name: '下载快照版本' })) })
    await waitFor(() => expect(
      fetchMock.mock.calls.some(([input]) => String(input).includes('/files/3002/download')),
    ).toBe(true))
    expect(fetchMock.mock.calls.some(([input]) => String(input).includes('/files/3001/download'))).toBe(false)
  })

  it('fails closed when the active pointer cannot be verified even with anchor-shaped payload', async () => {
    activePointer = ''
    snapshot = frozenSnapshot('A-REVIEW-1', 'I-REVIEW-1', [roundOneFingerprint])
    await renderApprovalExpecting('unconfirmable')
    expect(screen.getByText(/本轮提交快照无法确认属于本审批当前轮次/)).toBeInTheDocument()
    expect(screen.queryByTestId('approval-material-row-301')).not.toBeInTheDocument()
    expect(screen.queryByText('服务合同-v1.pdf')).not.toBeInTheDocument()
    expect(screen.queryByText('示例客户')).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '下载快照版本' })).not.toBeInTheDocument()
  })

  it('fails closed when frozen fingerprints carry a foreign intake anchor', async () => {
    snapshot = frozenSnapshot('A-REVIEW-1', 'I-OTHER', [roundOneFingerprint])
     await renderApprovalExpecting('unconfirmable')
     expect(screen.getByText(/本轮提交快照无法确认属于本审批当前轮次/)).toBeInTheDocument()
     expect(screen.queryByTestId('approval-material-row-301')).not.toBeInTheDocument()
     expect(screen.queryByText('服务合同-v1.pdf')).not.toBeInTheDocument()
     expect(screen.queryByText('示例客户')).not.toBeInTheDocument()
     expect(screen.queryByRole('button', { name: '下载快照版本' })).not.toBeInTheDocument()
  })

  it('fails closed when frozen fingerprints carry a foreign approval anchor', async () => {
    snapshot = frozenSnapshot('A-OTHER', 'I-REVIEW-1', [roundOneFingerprint])
     await renderApprovalExpecting('unconfirmable')
     expect(screen.getByText(/本轮提交快照无法确认属于本审批当前轮次/)).toBeInTheDocument()
     expect(screen.queryByTestId('approval-material-row-301')).not.toBeInTheDocument()
     expect(screen.queryByText('服务合同-v1.pdf')).not.toBeInTheDocument()
     expect(screen.queryByText('示例客户')).not.toBeInTheDocument()
     expect(screen.queryByRole('button', { name: '下载快照版本' })).not.toBeInTheDocument()
  })

  it('hides decoy legacy materials and frozen downloads when the intake anchor mismatches', async () => {
    approval = {
      status: 'needs_revision',
      metadata: {
        intake_id: 'I-REVIEW-1',
        materials: [
          { id: 999, name: '旧元数据诱饵材料', material_type: 'contract', status: 'uploaded' },
          { id: 998, name: '旧元数据身份材料', material_type: 'identity', status: 'uploaded' },
        ],
      },
    records: [{ id: 'R1', decision: 'request_changes', imposed_requirements: { items: [] } }],
  }
    snapshot = frozenSnapshot('A-REVIEW-1', 'I-OTHER', [roundOneFingerprint])
    await renderApprovalExpecting('unconfirmable')
    expect(screen.getByText(/本轮提交快照无法确认属于本审批当前轮次/)).toBeInTheDocument()
    expect(screen.queryByTestId('approval-material-row-301')).not.toBeInTheDocument()
    expect(screen.queryByText('旧元数据诱饵材料')).not.toBeInTheDocument()
    expect(screen.queryByText('旧元数据身份材料')).not.toBeInTheDocument()
    expect(screen.queryByText('服务合同-v1.pdf')).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '下载快照版本' })).not.toBeInTheDocument()
    expect(fetchMock.mock.calls.some(([input]) => String(input).includes('/case-materials/'))).toBe(false)
  })

  it('applies the material filter to frozen fingerprint rows', async () => {
    snapshot = frozenSnapshot('A-REVIEW-1', 'I-REVIEW-1', [
      roundOneFingerprint,
      {
        ...roundOneFingerprint,
        material_id: 302,
        material_type: 'evidence',
        name: '主体证明',
        current_file_id: 3003,
        files: [{ file_id: 3003, file_name: '主体证明.pdf' }],
      },
    ])
    await renderApproval()
    expect(screen.getByTestId('approval-material-row-301')).toBeInTheDocument()
    expect(screen.getByTestId('approval-material-row-302')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: '申请文档' }))
    expect(screen.getByTestId('approval-material-row-301')).toBeInTheDocument()
    expect(screen.queryByTestId('approval-material-row-302')).not.toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: '全部材料' }))
    expect(screen.getByTestId('approval-material-row-302')).toBeInTheDocument()
  })

  it('keeps review status copy aligned with verified continuation', async () => {
    continuationProof = continuation('I-REVIEW-1', 'rev-1', true)
    await openRevisionWorkbench()
    await screen.findByDisplayValue('股权争议')
    const stepButtons = document.querySelectorAll('.batch-stepper button')
    expect(stepButtons.length).toBeGreaterThan(4)
    fireEvent.click(stepButtons[4])
    await waitFor(() => expect(screen.getByTestId('intake-review-copy')).toHaveTextContent('已复核：未发现冲突'))
    expect(screen.queryByText('冲突检查 未检测')).not.toBeInTheDocument()
    continuationFailure = true
    await act(async () => { window.dispatchEvent(new Event('focus')) })
    await waitFor(() => expect(screen.getByTestId('intake-review-copy')).toHaveTextContent(/待复核|未检测|无法确认/))
  })
})
