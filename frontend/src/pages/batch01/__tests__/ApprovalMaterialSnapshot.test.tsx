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
  metadata: {
    intake_id: 'INTAKE-1',
    materials: [
      { id: 5, name: '客户主体资料', material_type: 'evidence', status: 'uploaded' },
      { id: 7, name: '营业执照', material_type: 'identity', status: 'uploaded' },
    ],
  },
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

describe('Approval material snapshot versions', () => {
  beforeEach(() => {
    getUserInfoMock.mockReturnValue({ id: 2, role: 'director' })
    window.localStorage.clear()
    fetchMock.mockReset()
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      if (path.endsWith('/approvals/APPROVAL-1')) {
        return Promise.resolve(jsonResponse(approvalBase))
      }
      if (path.endsWith('/approvals/APPROVAL-1/snapshot')) {
        return Promise.resolve(
          jsonResponse({
            approval_id: 'APPROVAL-1',
            snapshot_id: 'snapshot-current',
            // The API returns snapshot_data itself, not the DB row's
            // snapshot_type or a nested metadata.material_fingerprints bag.
            snapshot: {
              intake_id: 'INTAKE-1',
              approval_id: 'APPROVAL-1',
                material_fingerprints: [
                  {
                    material_id: 5,
                    name: '客户主体资料',
                    material_type: 'evidence',
                    required: true,
                    current_file_id: 13,
                    status: 'uploaded',
                    files: [
                      { file_id: 12, file_name: '旧版主体资料.pdf', digest: 'aaa', size_bytes: 10 },
                      { file_id: 13, file_name: '新版主体资料.pdf', digest: 'bbb', size_bytes: 12 },
                    ],
                  },
                  {
                    material_id: 7,
                    name: '营业执照',
                    material_type: 'identity',
                    required: true,
                    status: 'pending',
                    current_file_id: null,
                    files: [],
                  },
                ],
            },
          }),
        )
      }
      if (path.endsWith('/case-materials/5/files/13/download')) {
        return Promise.resolve({
          ok: true,
          blob: async () => new Blob(['binary']),
        } as unknown as Response)
      }
      if (path.endsWith('/integration/approvals/APPROVAL-1/status')) {
        return Promise.resolve(jsonResponse({ overall_status: 'pending', case_creation: {} }))
      }
      return Promise.resolve(jsonResponse({}))
    })
  })

  it('shows frozen snapshot files and downloads the frozen current pointer via the authorized endpoint', async () => {
    const openSpy = jest.spyOn(window, 'open').mockImplementation(() => null)
    renderFlow()

    // Snapshot-frozen file names are listed; live-only names are not.
    expect(await screen.findByText('旧版主体资料.pdf')).toBeInTheDocument()
    expect(screen.getByText('新版主体资料.pdf')).toBeInTheDocument()

    // Download uses the authorized API path (Bearer header), frozen current file.
    fireEvent.click(screen.getByRole('button', { name: '下载快照版本' }))
    await waitFor(() => {
      const call = fetchMock.mock.calls.find(([input]) =>
        String(input).endsWith('/case-materials/5/files/13/download'),
      )
      expect(call).toBeDefined()
      const headers = (call?.[1] as RequestInit).headers as Record<string, string>
      expect(headers.Authorization).toBe('Bearer test-token')
    })
    expect(openSpy).not.toHaveBeenCalled()
    openSpy.mockRestore()
  })

  it('marks a frozen material with no file version as unknown instead of showing live state', async () => {
    renderFlow()
    expect(await screen.findByText('客户主体资料')).toBeInTheDocument()
    const unknowns = await screen.findAllByText('快照未记录文件版本')
    expect(unknowns.length).toBe(1)
    expect(screen.getByText('营业执照')).toBeInTheDocument()
    // No download action exists for an unknown-version row.
    expect(screen.getAllByRole('button', { name: '下载快照版本' })).toHaveLength(1)
  })
})
