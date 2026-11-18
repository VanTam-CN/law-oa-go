import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes, useLocation } from 'react-router'
import { CaseIntakeWorkbench, ClientMasterProfile } from '../Batch01Prototype'

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

const approvalDetail = {
  id: 'approval-route-1',
  active_submission_snapshot_id: 'snapshot-route-current',
  records: [
    {
      id: 'return-route-1',
      decision: 'request_changes',
      approvalDate: '2026-09-27T08:00:00Z',
      imposed_requirements: JSON.stringify({
        items: [
          {
            item_id: 'reply-route-client-contact',
            category: 'field_correction',
            field_key: 'client_contact',
            description: '请更正客户档案联系方式。',
            required: true,
          },
        ],
      }),
    },
  ],
}

function LocationProbe() {
  const location = useLocation()
  return <div data-testid='route-location'>{location.pathname}</div>
}

function renderRevisionWorkbench() {
  return render(
    <MemoryRouter
      initialEntries={['/case/create?mode=revision&intake_id=intake-route-1&approval_id=approval-route-1']}
    >
     <Routes>
       <Route path='/case/create' element={<CaseIntakeWorkbench />} />
        <Route
          path='/client'
          element={
            <div data-testid='client-page'>
              <LocationProbe />
              <ClientMasterProfile />
            </div>
          }
        />
        <Route
          path='*'
          element={
            <div data-testid='not-found'>
              <LocationProbe />
              404 Not Found
            </div>
          }
        />
     </Routes>
    </MemoryRouter>,
  )
}

describe('client_contact locate detour reaches the real client page', () => {
  beforeAll(() => {
    Element.prototype.scrollIntoView = jest.fn()
  })
  beforeEach(() => {
    window.localStorage.clear()
    fetchMock.mockReset()
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      const url = new URL(path, 'http://localhost:3000')
      const method = String(init?.method || 'GET').toUpperCase()
      if (url.pathname === '/api/v1/clients') {
        return Promise.resolve(jsonResponse({ clients: [], pagination: { page: 1, page_size: 20, total: 0, total_page: 0 } }))
      }
      if (path.endsWith('/lawfirm/lawyers?page=1&page_size=50')) {
        return Promise.resolve(jsonResponse({ lawyers: [{ id: 7, name: '律师甲' }] }))
      }
      if (path.endsWith('/case-intakes/intake-route-1') && method === 'GET') {
        return Promise.resolve(jsonResponse({
          intake: {
            id: 'intake-route-1',
            intake_code: 'INT-ROUTE-1',
            title: '虚构接案争议',
            case_type: 'commercial',
            priority: 'high',
            description: '虚构案件摘要',
            client_id: 42,
            created_by: '7',
            metadata: { lawyer_id: 7, business_area: '商事', sub_area: '合同' },
          },
          revision: 'intake-route-revision-1',
          parties: [],
        }))
      }
      if (path.endsWith('/case-intakes/intake-route-1/materials')) {
        return Promise.resolve(jsonResponse({ materials: [] }))
      }
      if (path.endsWith('/approvals/approval-route-1/snapshot')) {
        return Promise.resolve(jsonResponse({ approval_id: 'approval-route-1', snapshot_id: 'snapshot-route-current', snapshot: {} }))
      }
      if (path.includes('/approvals/approval-route-1/return-items') && method === 'GET') {
        return Promise.resolve(jsonResponse({ items: [{ item_id: 'reply-route-client-contact' }], replies: [] }))
      }
      if (path.includes('/approvals/approval-route-1/return-items') && method === 'PUT') {
        return Promise.resolve(jsonResponse({ saved: true }))
      }
      if (path.endsWith('/approvals/approval-route-1')) {
        return Promise.resolve(jsonResponse(approvalDetail))
      }
      if (url.pathname.includes('/master-profile')) {
        return Promise.resolve(jsonResponse({ client: { id: 42, name: '虚构客户甲' }, completeness: { score: 80 }, primary_contact: null, related_parties: [], matter_history: [], conflict_history: [] }))
      }
      return Promise.resolve(jsonResponse({}))
    })
  })

  it('navigates to the registered /client route and mounts ClientMasterProfile', async () => {
    renderRevisionWorkbench()
    const locateButton = await screen.findByRole('button', { name: '定位并填写该字段' })
    fireEvent.click(locateButton)
    await screen.findByText('客户联系方式需在客户档案中更正')
    fireEvent.click(screen.getByRole('button', { name: '打开客户列表' }))

    await waitFor(() => {
      expect(screen.getByTestId('route-location').textContent).toBe('/client')
    })
    expect(screen.getByTestId('client-page')).toBeInTheDocument()
    expect(screen.queryByTestId('not-found')).not.toBeInTheDocument()
  })

  it('does not land on the /clients redirect alias', async () => {
    renderRevisionWorkbench()
    const locateButton = await screen.findByRole('button', { name: '定位并填写该字段' })
    fireEvent.click(locateButton)
    await screen.findByText('客户联系方式需在客户档案中更正')
    fireEvent.click(screen.getByRole('button', { name: '打开客户列表' }))

    await waitFor(() => {
      const pathname = screen.getByTestId('route-location').textContent
      expect(pathname).not.toBe('/clients')
    })
  })
})
