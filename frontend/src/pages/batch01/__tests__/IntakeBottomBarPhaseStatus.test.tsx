import React from 'react'
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, useLocation, useNavigate } from 'react-router'
import { CaseIntakeWorkbench, scopedCaseIntakeDraftKey } from '../Batch01Prototype'

jest.mock('@/utils/storage', () => ({
  getRoles: jest.fn(() => []), getToken: jest.fn(() => 'test-token'),
  getUserInfo: jest.fn(() => ({ id: 7, role: 'lawyer' })),
}))
jest.mock('@/utils/messageHelper', () => ({
  message: { error: jest.fn(), success: jest.fn(), warning: jest.fn(), info: jest.fn() },
  setAppMessage: jest.fn(),
}))

const fetchMock = jest.fn()
global.fetch = fetchMock as typeof fetch
const json = (data: unknown) => ({ ok: true, json: async () => ({ success: true, data }) }) as Response
const intake = (id = 'I1') => ({
  intake: { id, intake_code: `INT-${id}`, title: `案件-${id}`, case_type: 'commercial',
    description: '本轮案情摘要', client_id: 42, created_by: '7',
    metadata: { lawyer_id: 7, business_area: '商事', sub_area: '合同' } },
  revision: 'rev-1',
  parties: [
    { entity_name: '虚构客户', entity_type: 'LEGAL_PERSON', party_role: 'client' },
    { entity_name: '虚构对方', entity_type: 'LEGAL_PERSON', party_role: 'opposing_party' },
  ],
})
const continuation = (id = 'I1', approved = false) => ({
  intake_id: id, revision: 'rev-1', status: 'conflict_ready', can_reuse_identity: true,
  can_reuse_check: true, can_submit: approved, recheck_required: false,
  reasons: approved ? [] : ['REVIEW_MISSING'], facts: { facts_version: 'fv1', confirmed_at: '2026-10-02' },
  conflict_binding: { check_id: `CHECK-${id}`, check_status: 'COMPLETED', coverage_status: 'COMPLETE',
    ...(approved ? { review_id: 11, reviewer_id: 9, review_decision: 'no_conflict' } : {}) },
  subjects: [{ kind: 'party', party_id: 12, name: '虚构对方', role: 'opposing_party',
    entity_type: 'LEGAL_PERSON', identity_type: 'SOCIAL_CREDIT_CODE', identity_masked: '****0001' }],
})

function LocationProbe() {
  const location = useLocation()
  const navigate = useNavigate()
  return <><output data-testid='location'>{location.pathname + location.search}</output>
    <button onClick={() => navigate('/case/create?draft_id=I2')}>打开第二份申请</button></>
}

async function open(path = '/case/create?draft_id=I1') {
  await act(async () => {
    render(<MemoryRouter initialEntries={[path]}><CaseIntakeWorkbench /><LocationProbe /></MemoryRouter>)
  })
}

let proof: unknown

beforeEach(() => {
  window.localStorage.clear()
  proof = { ...continuation(), approval: { approval_id: 'A1', snapshot_id: 'S1', status: 'needs_revision', round: 1 } }
  fetchMock.mockReset()
  fetchMock.mockImplementation((input: RequestInfo | URL) => {
    const path = String(input)
    if (path.endsWith('/continuation')) return Promise.resolve(json(proof))
    if (/\/case-intakes\/I[12]$/.test(path)) return Promise.resolve(json(intake(path.endsWith('I2') ? 'I2' : 'I1')))
    if (path.endsWith('/approvals/A1')) return Promise.resolve(json({ id: 'A1', active_submission_snapshot_id: 'S1', records: [
      { id: 'R1', decision: 'request_changes', imposed_requirements: { items: [] } },
    ] }))
    if (path.endsWith('/approvals/A1/snapshot')) return Promise.resolve(json({ approval_id: 'A1', snapshot_id: 'S1', snapshot: {} }))
    if (path.includes('/clients?')) return Promise.resolve(json({ clients: [{ id: 42, name: '虚构客户', type: '企业' }] }))
    if (path.includes('/lawfirm/lawyers?')) return Promise.resolve(json({ lawyers: [{ id: 7, name: '律师甲' }] }))
    return Promise.resolve(json({}))
  })
})

// The bottom bar must agree with the phase headline shown next to the
// primary button: closed/in_progress/unknown mean the ORIGINAL approval owns
// this intake now, so "待独立人工复核" is wrong and hides which approval is
// actually in play. Display-only: canSubmit stays false and no hidden POST.
it.each(['approved', 'under_review'] as const)(
  'bottom status agrees with the original %s approval phase',
  async (status) => {
    proof = {
      ...continuation('I1', true),
      can_submit: false,
      reasons: [status === 'approved' ? 'APPROVAL_CLOSED' : 'APPROVAL_IN_PROGRESS'],
      approval: { approval_id: 'A1', snapshot_id: 'S1', status, round: 1 },
    }
    await open('/case/create?draft_id=I1')
    await waitFor(() => expect(screen.getByRole('button', { name: '提交审批并等待成案' })).toBeDisabled())
    const bottom = document.querySelector('.batch-bottom-bar')!
    expect(bottom).not.toHaveTextContent('待独立人工复核')
    expect(bottom).toHaveTextContent(status === 'approved' ? '原审批已结束' : /等待原审批|原审批正在处理中/)
    expect(fetchMock.mock.calls.some(([url, init]) => String(url).includes('/approvals') && init?.method === 'POST')).toBe(false)
  },
)
