import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
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

const approvalBase = {
  id: 'APPROVAL-1',
  status: 'resubmitted',
  applicant_id: 9,
  applicant_name: '申请人甲',
  current_approver_id: 2,
  current_approver_name: '审批乙',
  available_actions: ['approve', 'reject', 'request_changes'],
  priority: 'high',
  workflow_type: 'CASE_APPROVAL',
  content: '重提后的立案审批',
  created_at: '2026-10-01T09:00:00Z',
  metadata: { intake_id: 'INTAKE-1' },
  active_submission_snapshot_id: 'snapshot-current',
}

const fetchMock = jest.fn()
global.fetch = fetchMock as unknown as typeof fetch

function jsonResponse(data: unknown) {
  return { ok: true, json: async () => ({ success: true, data }) } as Response
}

function renderFlow() {
  return render(
    <MemoryRouter initialEntries={['/approval/APPROVAL-1/decision']}>
      <Routes>
        <Route path='/approval/:id/decision' element={<ApprovalDecisionFlow />} />
      </Routes>
    </MemoryRouter>,
  )
}

describe('ApprovalDecisionFlow mixed structured return', () => {
  beforeEach(() => {
    getUserInfoMock.mockReturnValue({ id: 2, role: 'director' })
    window.localStorage.clear()
    fetchMock.mockReset()
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const path = String(input)
      if (path.endsWith('/approvals/APPROVAL-1')) {
        return Promise.resolve(jsonResponse(approvalBase))
      }
      if (path.endsWith('/approvals/APPROVAL-1/snapshot')) {
        return Promise.resolve(
          jsonResponse({
            approval_id: 'APPROVAL-1',
            snapshot_id: 'snapshot-current',
            snapshot: {},
          }),
        )
      }
      if (path.endsWith('/case-intakes/INTAKE-1/materials')) {
        return Promise.resolve(
          jsonResponse({
            materials: [{ id: 5, name: '客户主体资料', status: 'missing' }],
          }),
        )
      }
      if (path.endsWith('/integration/approvals/APPROVAL-1/status')) {
        return Promise.resolve(jsonResponse({ overall_status: 'pending', case_creation: {} }))
      }
      return Promise.resolve(jsonResponse({}))
    })
  })

  it('sends field corrections and material requests as one whitelist item list', async () => {
    renderFlow()
    fireEvent.click(await screen.findByText('退回修改'))

    // Add one structured field correction and fill it in.
    fireEvent.click(screen.getByRole('button', { name: /添加字段更正项/ }))
    const fieldSelectInput = document.querySelector(
      '[data-testid="material-request-editor"] .ant-select-selector input',
    ) as HTMLElement | null
    expect(fieldSelectInput).not.toBeNull()
    fireEvent.mouseDown(fieldSelectInput as HTMLElement)
    const option = await waitFor(() => {
      const el = Array.from(document.querySelectorAll('.ant-select-item-option')).find(
        (node) => node.textContent === '案件名称',
      )
      expect(el).toBeDefined()
      return el as HTMLElement
    })
    fireEvent.click(option)
    fireEvent.change(screen.getByLabelText('字段更正说明：1'), {
      target: { value: '请将案件名称更正为股权回购协议争议。' },
    })

    // Check one real material row from the intake ledger and explain it.
    fireEvent.click(await screen.findByRole('checkbox', { name: '退回材料：客户主体资料' }))
    fireEvent.change(screen.getByLabelText('退回说明：客户主体资料'), {
      target: { value: '请补传加盖公章的主体资格证明。' },
    })

    fireEvent.change(screen.getByLabelText('处理依据'), {
      target: { value: '案件名称与申请材料不一致，且主体资料缺失，需要一并退回补充。' },
    })
    fireEvent.click(screen.getByRole('button', { name: '确认退回修改' }))

    await waitFor(() => {
      const call = fetchMock.mock.calls.find(([input]) =>
        String(input).endsWith('/integration/approvals/APPROVAL-1/decision'),
      )
      expect(call).toBeDefined()
      const request = call?.[1] as RequestInit
      const body = JSON.parse(String(request.body))
      expect(body.decision).toBe('request_changes')
      const items = body.imposed_requirements?.items
      expect(Array.isArray(items)).toBe(true)
      expect(items).toEqual(
        expect.arrayContaining([
          expect.objectContaining({
            category: 'field_correction',
            field_key: 'case_title',
            description: '请将案件名称更正为股权回购协议争议。',
            required: true,
          }),
          expect.objectContaining({
            category: 'material_request',
            field_key: 'material',
            material_id: 5,
            description: '请补传加盖公章的主体资格证明。',
            required: true,
          }),
        ]),
      )
      expect(items).toHaveLength(2)
    })
  })

  it('blocks the return when a checked item has no explanation instead of dropping it', async () => {
    renderFlow()
    fireEvent.click(await screen.findByText('退回修改'))

    // One material is checked but its explanation stays empty.
    fireEvent.click(await screen.findByRole('checkbox', { name: '退回材料：客户主体资料' }))
    fireEvent.change(screen.getByLabelText('处理依据'), {
      target: { value: '主体资料缺失，需要申请人补充后重新提交审批。' },
    })
    fireEvent.click(screen.getByRole('button', { name: '确认退回修改' }))

    // The POST must never fire, and the warning names the unfinished item.
    await waitFor(() => {
      const warning = require('@/utils/messageHelper').message.warning as jest.Mock
      expect(warning).toHaveBeenCalledWith(
        expect.stringContaining('客户主体资料'),
      )
    })
    expect(
      fetchMock.mock.calls.some(([input]) =>
        String(input).endsWith('/integration/approvals/APPROVAL-1/decision'),
      ),
    ).toBe(false)
  })
})
