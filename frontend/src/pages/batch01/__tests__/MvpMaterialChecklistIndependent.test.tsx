import React from 'react'
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { CaseIntakeWorkbench } from '../Batch01Prototype'

jest.mock('@/utils/storage', () => ({ getRoles: () => [], getToken: () => 'independent-material-token', getUserInfo: () => ({ id: 7, role: 'lawyer' }) }))
jest.mock('@/utils/messageHelper', () => ({ message: { error: jest.fn(), success: jest.fn(), warning: jest.fn(), info: jest.fn() }, setAppMessage: jest.fn() }))

const fetchMock = jest.fn()
global.fetch = fetchMock as unknown as typeof fetch
const json = (data: unknown) => ({ ok: true, status: 200, json: async () => ({ success: true, data }) }) as Response
const item = { item_id: 'material-original', category: 'material_request', field_key: 'material', material_id: 200, description: '请在原合同材料行补交签章页', required: true }

async function open(evidence: unknown) {
  fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
    const path = String(input)
    if (path.endsWith('/continuation')) return Promise.resolve(json({
      intake_id: 'I1', revision: 'rev1', status: 'conflict_ready', can_reuse_identity: true, can_reuse_check: true, can_submit: true, recheck_required: false,
      reasons: [], facts: { facts_version: 'fv1', confirmed_at: '2026-10-02' },
      conflict_binding: { check_id: 'C1', check_status: 'COMPLETED', coverage_status: 'COMPLETE', review_id: 11, reviewer_id: 9, review_decision: 'no_conflict' },
      approval: { approval_id: 'A1', status: 'needs_revision', snapshot_id: 'S1', round: 1 },
    }))
    if (path.endsWith('/case-intakes/I1')) return Promise.resolve(json({
      intake: { id: 'I1', title: '虚构合同争议', case_type: 'commercial', description: '虚构摘要', client_id: 42, created_by: '7', metadata: { lawyer_id: 7, business_area: '商事', sub_area: '合同' } },
      revision: 'rev1', parties: [{ entity_name: '虚构客户', entity_type: 'LEGAL_PERSON', party_role: 'client' }, { entity_name: '虚构对方', entity_type: 'LEGAL_PERSON', party_role: 'opposing_party' }],
    }))
    if (path.endsWith('/approvals/A1/snapshot')) return Promise.resolve(json({ approval_id: 'A1', snapshot_id: 'S1', snapshot: {} }))
    if (path.endsWith('/approvals/A1')) return Promise.resolve(json({ id: 'A1', status: 'needs_revision', active_submission_snapshot_id: 'S1', records: [{ id: 'R1', decision: 'request_changes', imposed_requirements: { items: [item] } }] }))
    if (path.includes('/approvals/A1/return-items') && !init?.method) return Promise.resolve(json({ items: [item], replies: [], material_evidence: evidence }))
    if (path.endsWith('/case-intakes/I1/materials')) return Promise.resolve(json({ materials: [{ id: 200, name: '原合同', status: 'uploaded', current_file_id: 10, current_file: { id: 10, file_name: '退回前合同.pdf' } }] }))
    if (path.includes('/clients?')) return Promise.resolve(json({ clients: [{ id: 42, name: '虚构客户', type: '企业' }] }))
    if (path.includes('/lawfirm/lawyers?')) return Promise.resolve(json({ lawyers: [{ id: 7, name: '律师甲' }] }))
    return Promise.resolve(json({}))
  })
  await act(async () => {
    render(<MemoryRouter initialEntries={['/case/create?mode=revision&intake_id=I1&approval_id=A1']}><CaseIntakeWorkbench /></MemoryRouter>)
  })
  await waitFor(() => expect(screen.getByTestId('intake-continuation')).toHaveTextContent('可继续提交审批'))
  const reply = await screen.findByRole('textbox', { name: '回复：原合同' })
  fireEvent.change(reply, { target: { value: '我已填写说明，但说明不代表上传新附件。' } })
  return reply
}

describe('Independent returned-material submission guidance', () => {
  beforeEach(() => { fetchMock.mockReset(); window.localStorage.clear() })

  it.each(['missing', 'unknown', 'missing_response', 'wrong_item', 'complete_without_id'])(
    'retains the reply and blocks resubmit when new original-row evidence is %s', async (variant) => {
      const evidence = variant === 'missing_response' ? undefined : [{
        item_id: variant === 'wrong_item' ? 'other-item' : item.item_id, material_id: 200,
        state: variant === 'missing' ? 'missing' : variant === 'unknown' ? 'unknown' : 'complete',
        ...(variant === 'complete_without_id' ? {} : { current_file_id: 10 }), message: '',
      }]
      const reply = await open(evidence)
      expect(reply).toHaveValue('我已填写说明，但说明不代表上传新附件。')
      expect(screen.getByRole('button', { name: '重新提交审批' })).toBeDisabled()
      expect(screen.getByTestId('supplement-status-material-original')).not.toHaveTextContent('补正附件已上传')
      expect(fetchMock.mock.calls.some(([path]) => String(path).endsWith('/resubmit'))).toBe(false)
    },
  )

  it('enables same-approval resubmit for a server-confirmed new original-row file, without a new check', async () => {
    await open([{ item_id: item.item_id, material_id: 200, state: 'complete', current_file_id: 11, message: '已上传补正附件' }])
    expect(screen.getByTestId('supplement-status-material-original')).toHaveTextContent('补正附件已上传，待审批复核')
    expect(screen.getByRole('button', { name: '重新提交审批' })).toBeEnabled()
    await act(async () => { fireEvent.click(screen.getByRole('button', { name: '重新提交审批' })) })
    expect(fetchMock.mock.calls.some(([path]) => String(path).endsWith('/approvals/A1/resubmit'))).toBe(true)
    expect(fetchMock.mock.calls.some(([path]) => /facts-confirmation|conflict-check/.test(String(path)))).toBe(false)
  })
})
