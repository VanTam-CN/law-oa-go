import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { CaseIntakeWorkbench, supplementFieldLocateTargets } from '../Batch01Prototype'

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
          {
            item_id: 'reply-client-contact',
            category: 'field_correction',
            field_key: 'client_contact',
            description: '请更正客户档案联系方式。',
            required: true,
          },
          {
            item_id: 'reply-material-1',
            category: 'material_request',
            field_key: 'material',
            material_id: 9001,
            description: '请上传盖章版委托代理协议。',
            required: true,
        },
        ],
      }),
    },
  ],
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

describe('Supplement field locate targeting', () => {
  beforeAll(() => {
    // JSDOM does not implement scrollIntoView.
    Element.prototype.scrollIntoView = jest.fn()
  })
  beforeEach(() => {
    window.localStorage.clear()
    fetchMock.mockReset()
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      const method = String(init?.method || 'GET').toUpperCase()
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
      if (path.endsWith('/case-intakes/intake-1/materials')) {
        return Promise.resolve(jsonResponse({ materials: [{ id: 9001, name: '委托代理协议', status: 'missing' }] }))
      }
      if (path.endsWith('/approvals/approval-1/snapshot')) {
        return Promise.resolve(jsonResponse({ approval_id: 'approval-1', snapshot_id: 'snapshot-current', snapshot: {} }))
      }
      if (path.includes('/approvals/approval-1/return-items') && method === 'GET') {
        return Promise.resolve(jsonResponse({ items: [{ item_id: 'reply-case-title' }, { item_id: 'reply-client-contact' }, { item_id: 'reply-material-1' }], replies: [] }))
      }
      if (path.includes('/approvals/approval-1/return-items') && method === 'PUT') {
        return Promise.resolve(jsonResponse({ saved: true }))
      }
      if (path.endsWith('/approvals/approval-1')) {
        return Promise.resolve(jsonResponse(approvalDetail))
      }
      return Promise.resolve(jsonResponse({}))
    })
  })


  it('locates the real field by selector and routes client_contact to the client-archive detour', async () => {
    renderRevisionWorkbench()
    const locateButtons = await screen.findAllByRole('button', { name: '定位并填写该字段' })
    expect(locateButtons.length).toBeGreaterThanOrEqual(2)

    // case_title focuses the real input.
    fireEvent.click(locateButtons[0])
    await waitFor(() => expect(document.querySelector('#intake-title')).not.toBeNull())
    await waitFor(() => expect(document.activeElement?.id).toBe('intake-title'))

    // client_contact never touches the lead-source contact input.
    fireEvent.click(locateButtons[1])
    const dialog = await screen.findByText('客户联系方式需在客户档案中更正')
    expect(dialog).toBeInTheDocument()
    expect(screen.getAllByText(/案源联系人/).length).toBeGreaterThan(0)

    // "打开客户列表" navigates away without focusing the wrong field.
    fireEvent.click(screen.getByRole('button', { name: '打开客户列表' }))
    await waitFor(() => expect(screen.queryByText('客户联系方式需在客户档案中更正')).not.toBeInTheDocument())
    expect(document.activeElement?.id).not.toBe('intake-source-contact')
  })

  it('cancels an earlier field retry loop when a material locate starts', async () => {
    renderRevisionWorkbench()
    const locateButtons = await screen.findAllByRole('button', { name: '定位并填写该字段' })

    // Field A is armed; its first retry would fire ~150ms later.
    fireEvent.click(locateButtons[0])
    // Material B starts BEFORE A's first tick: it must invalidate A's loop.
    const materialButton = await screen.findByRole('button', { name: '去第 4 步上传该材料' })
    fireEvent.click(materialButton)

    // B lands on the real material row.
    await waitFor(() => {
      expect(document.querySelector('[data-testid="material-row-9001"]')).not.toBeNull()
    })
    const scrollMock = Element.prototype.scrollIntoView as unknown as jest.Mock
    await waitFor(() => {
      expect(
        scrollMock.mock.contexts.some(
          (el) => el?.getAttribute?.('data-testid') === 'material-row-9001',
        ),
      ).toBe(true)
    })

    // A was canceled: its timer never focused the title field.
    await new Promise((resolve) => setTimeout(resolve, 350))
    expect(document.activeElement?.id).not.toBe('intake-title')
  })

  it('cancels pending retries when the user switches steps manually', async () => {
    renderRevisionWorkbench()
    const locateButtons = await screen.findAllByRole('button', { name: '定位并填写该字段' })
    fireEvent.click(locateButtons[0])

    // A manual step switch must invalidate the locate loop.
    const stepperButtons = document.querySelectorAll('.batch-stepper button')
    expect(stepperButtons.length).toBeGreaterThanOrEqual(2)
    fireEvent.click(stepperButtons[1])

    await new Promise((resolve) => setTimeout(resolve, 350))
    expect(document.activeElement?.id).not.toBe('intake-title')
  })
})
