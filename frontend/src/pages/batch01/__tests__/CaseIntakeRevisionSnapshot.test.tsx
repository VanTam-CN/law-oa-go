import React from 'react'
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { CaseIntakeWorkbench } from '../Batch01Prototype'

jest.mock('@/utils/storage', () => ({
  getRoles: jest.fn(() => []),
  getToken: jest.fn(() => 'test-token'),
  getUserInfo: jest.fn(() => ({ id: 7, role: 'lawyer', roles: ['lawyer'], name: '律师甲' })),
}))

jest.mock('@/utils/messageHelper', () => ({
  message: { error: jest.fn(), success: jest.fn(), warning: jest.fn(), info: jest.fn() },
  setAppMessage: jest.fn(),
}))

const fetchMock = jest.fn()
global.fetch = fetchMock as unknown as typeof fetch

function jsonResponse(data: unknown) {
  return { ok: true, json: async () => ({ success: true, data }) } as Response
}

const intakePayload = {
  intake: {
    id: 'intake-1',
    intake_code: 'INT-1',
    title: '虚构接案争议',
    case_type: 'commercial',
    priority: 'high',
    description: '虚构案件摘要',
    client_id: 42,
    created_by: '7',
    metadata: { lawyer_id: 7, business_area: '商事', sub_area: '合同' },
  },
  revision: 'intake-revision-1',
  parties: [
    { entity_name: '虚构客户甲', entity_type: 'LEGAL_PERSON', party_role: 'client' },
    { entity_name: '虚构对手乙', entity_type: 'LEGAL_PERSON', party_role: 'opposing_party' },
  ],
}

const approvalDetail = {
  id: 'approval-1',
  active_submission_snapshot_id: 'snapshot-current',
  records: [
    {
      id: 'return-1',
      decision: 'request_changes',
      approvalDate: '2026-09-27T08:00:00Z',
      imposed_requirements: JSON.stringify({
        items: [
          {
            item_id: 'reply-case-title',
            category: 'field_correction',
            field_key: 'case_title',
            description: '请说明案件名称对应的争议标的。',
            required: true,
          },
        ],
      }),
    },
  ],
}

let currentRevision = 'intake-revision-1'
let reviewApproved = true
let checkReady = true
function continuationPayload() {
  return {
    intake_id: 'intake-1', revision: currentRevision, status: checkReady ? 'conflict_ready' : 'draft',
    can_reuse_identity: checkReady, can_reuse_check: checkReady, can_submit: checkReady && reviewApproved,
    recheck_required: !checkReady, reasons: checkReady ? reviewApproved ? [] : ['REVIEW_MISSING'] : ['CONFLICT_CHECK_INCOMPLETE'],
    facts: { facts_version: 'facts-current', confirmed_at: '2026-10-02' },
    conflict_binding: { check_id: 'CHECK-current', check_status: 'COMPLETED', coverage_status: 'COMPLETE',
      ...(reviewApproved ? { review_id: 11, reviewer_id: 9, review_decision: 'no_conflict' } : {}) },
    approval: { approval_id: 'approval-1', status: 'needs_revision', snapshot_id: 'snapshot-current', round: 1 },
  }
}

function renderRevisionWorkbench() {
  return render(
    <MemoryRouter
      initialEntries={['/case/create?mode=revision&intake_id=intake-1&approval_id=approval-1']}
    >
      <CaseIntakeWorkbench />
    </MemoryRouter>,
  )
}

describe('CaseIntakeWorkbench structured resubmit snapshot contract', () => {
  beforeEach(() => {
    window.localStorage.clear()
    currentRevision = 'intake-revision-1'
    reviewApproved = true
    checkReady = true
    fetchMock.mockReset()
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      const method = String(init?.method || 'GET').toUpperCase()
      if (path.endsWith('/continuation')) return Promise.resolve(jsonResponse(continuationPayload()))
      if (path.endsWith('/clients?page=1&page_size=50')) {
        return Promise.resolve(jsonResponse({ clients: [{ id: 42, name: '虚构客户甲', type: '企业' }] }))
      }
      if (path.endsWith('/lawfirm/lawyers?page=1&page_size=50')) {
        return Promise.resolve(jsonResponse({ lawyers: [{ id: 7, name: '律师甲' }] }))
      }
      if (path.endsWith('/case-intakes/intake-1') && method === 'GET') {
        return Promise.resolve(jsonResponse(intakePayload))
      }
      if (path.endsWith('/case-intakes/intake-1') && method === 'PUT') {
        return Promise.resolve(jsonResponse({ revision: 'intake-revision-2', intake_code: 'INT-1' }))
      }
      if (path.endsWith('/approvals/approval-1/snapshot')) {
        return Promise.resolve(
          jsonResponse({
            approval_id: 'approval-1',
            snapshot_id: 'snapshot-current',
            snapshot: {},
          }),
        )
      }
      if (path.endsWith('/approvals/approval-1')) {
        return Promise.resolve(jsonResponse(approvalDetail))
      }
      if (path.endsWith('/approvals/approval-1/resubmit') && method === 'POST') {
        return Promise.resolve(jsonResponse({ status: 'resubmitted' }))
      }
      return Promise.resolve(jsonResponse({}))
    })
  })

  it('keeps resubmit enabled without a new conflict result and pins the request to the loaded snapshot', async () => {
    renderRevisionWorkbench()
    const reply = await screen.findByRole('textbox', { name: '回复：案件名称' })
    fireEvent.change(reply, { target: { value: '争议标的是虚构股权回购义务。' } })

    const submit = await screen.findByRole('button', { name: '重新提交审批' })
    await waitFor(() => expect(submit).toBeEnabled())
    await act(async () => {
      fireEvent.click(submit)
    })

    await waitFor(() => {
      expect(
        fetchMock.mock.calls.filter(([input, init]) =>
          String(input).endsWith('/case-intakes/intake-1') &&
          String(init?.method || '').toUpperCase() === 'PUT',
        ),
      ).toHaveLength(0)
      const call = fetchMock.mock.calls.find(([input, init]) =>
        String(input).endsWith('/approvals/approval-1/resubmit') &&
        String(init?.method || '').toUpperCase() === 'POST',
      )
      expect(call).toBeDefined()
      const request = call?.[1] as RequestInit
      expect(JSON.parse(String(request.body))).toMatchObject({
        expected_submission_snapshot_id: 'snapshot-current',
        expected_intake_revision: 'intake-revision-1',
        return_record_id: 'return-1',
        replies: [{ item_id: 'reply-case-title', response: '争议标的是虚构股权回购义务。' }],
      })
    })
  })
})

describe('CaseIntakeWorkbench revision fact edits', () => {
  beforeEach(() => {
    window.localStorage.clear()
    currentRevision = 'intake-revision-1'
    reviewApproved = true
    checkReady = true
    fetchMock.mockReset()
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      const method = String(init?.method || 'GET').toUpperCase()
      if (path.endsWith('/continuation')) return Promise.resolve(jsonResponse(continuationPayload()))
      if (path.endsWith('/clients?page=1&page_size=50')) {
        return Promise.resolve(jsonResponse({ clients: [{ id: 42, name: '虚构客户甲', type: '企业' }] }))
      }
      if (path.endsWith('/lawfirm/lawyers?page=1&page_size=50')) {
        return Promise.resolve(jsonResponse({ lawyers: [{ id: 7, name: '律师甲' }] }))
      }
      if (path.endsWith('/case-intakes/intake-1') && method === 'GET') {
        return Promise.resolve(jsonResponse(intakePayload))
      }
      if (path.endsWith('/case-intakes/intake-1') && method === 'PUT') {
        currentRevision = 'intake-revision-2'
        reviewApproved = false
        checkReady = false
        return Promise.resolve(jsonResponse({ revision: currentRevision, intake_code: 'INT-1' }))
      }
      if (path.endsWith('/facts-confirmation') && method === 'POST') {
        currentRevision = 'intake-revision-3'
        return Promise.resolve(jsonResponse({ revision: currentRevision }))
      }
      if (path.endsWith('/conflict-check') && method === 'POST') {
        checkReady = true
        return Promise.resolve(jsonResponse({ task_id: 'CHECK-current', status: 'COMPLETED', result: {
          checkId: 'CHECK-current', riskAssessment: { overallRisk: 'HIGH', requiresApproval: true },
          decision: { status: 'REVIEW_REQUIRED', coverageStatus: 'COMPLETE' },
        } }))
      }
      if (path.endsWith('/approvals/approval-1/snapshot')) {
        return Promise.resolve(
          jsonResponse({
            approval_id: 'approval-1',
            snapshot_id: 'snapshot-current',
            snapshot: {},
          }),
        )
      }
      if (path.endsWith('/approvals/approval-1')) {
        return Promise.resolve(jsonResponse(approvalDetail))
      }
      if (path.endsWith('/approvals/approval-1/resubmit') && method === 'POST') {
        return Promise.resolve(jsonResponse({ status: 'resubmitted' }))
      }
      return Promise.resolve(jsonResponse({}))
    })
  })

  it('saves changed facts and requires a new check/review before same-snapshot resubmission', async () => {
    renderRevisionWorkbench()
    const reply = await screen.findByRole('textbox', { name: '回复：案件名称' })
    fireEvent.change(reply, { target: { value: '争议标的是虚构股权回购义务。' } })
    const title = await screen.findByLabelText('案件名称 *')
    fireEvent.change(title, { target: { value: '虚构案件名称（修订版）' } })

    const submit = await screen.findByRole('button', { name: '重新提交审批' })
    expect(submit).toBeDisabled()
    expect(fetchMock.mock.calls.some(([path]) => String(path).endsWith('/resubmit'))).toBe(false)
    fireEvent.change(screen.getByLabelText('对方身份标识 *'), { target: { value: '91310000TEST000001' } })
    await act(async () => { fireEvent.click(screen.getByRole('button', { name: '保存并进行利益冲突检查' })) })
    await waitFor(() => expect(screen.getByTestId('intake-continuation')).toHaveTextContent('等待独立复核'))
    expect(submit).toBeDisabled()
    expect(fetchMock.mock.calls.some(([path]) => String(path).endsWith('/resubmit'))).toBe(false)
    reviewApproved = true
    await act(async () => { window.dispatchEvent(new Event('focus')) })
    await waitFor(() => expect(submit).toBeEnabled())
    await act(async () => {
      fireEvent.click(submit)
    })

    await waitFor(() => {
      expect(
        fetchMock.mock.calls.filter(([input, init]) =>
          String(input).endsWith('/case-intakes/intake-1') &&
          String(init?.method || '').toUpperCase() === 'PUT',
        ),
      ).toHaveLength(1)
      const call = fetchMock.mock.calls.find(([input, init]) =>
        String(input).endsWith('/approvals/approval-1/resubmit') &&
        String(init?.method || '').toUpperCase() === 'POST',
      )
      expect(call).toBeDefined()
      const request = call?.[1] as RequestInit
      expect(JSON.parse(String(request.body))).toMatchObject({
        expected_submission_snapshot_id: 'snapshot-current',
        expected_intake_revision: 'intake-revision-3',
        return_record_id: 'return-1',
        replies: [{ item_id: 'reply-case-title', response: '争议标的是虚构股权回购义务。' }],
      })
      const save = fetchMock.mock.calls.find(([path, init]) => String(path).endsWith('/case-intakes/intake-1') && init?.method === 'PUT')
      expect(JSON.parse(String(save?.[1]?.body))).toMatchObject({ title: '虚构案件名称（修订版）' })
      const confirmation = fetchMock.mock.calls.find(([path]) => String(path).endsWith('/facts-confirmation'))
      expect(confirmation?.[1]?.headers).toMatchObject({ 'X-Intake-Revision': 'intake-revision-2' })
      const check = fetchMock.mock.calls.find(([path]) => String(path).endsWith('/conflict-check'))
      expect(check?.[1]?.headers).toMatchObject({ 'X-Intake-Revision': 'intake-revision-3' })
    })
  })
})
