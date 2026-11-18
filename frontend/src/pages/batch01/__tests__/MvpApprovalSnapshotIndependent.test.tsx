import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router'
import { ApprovalDecisionFlow } from '../Batch01Prototype'

jest.mock('@/utils/storage', () => ({ getRoles: () => [], getToken: () => 'independent-snapshot-token', getUserInfo: () => ({ id: 2, role: 'director' }) }))
jest.mock('@/utils/messageHelper', () => ({ message: { error: jest.fn(), success: jest.fn(), warning: jest.fn(), info: jest.fn() }, setAppMessage: jest.fn() }))

const fetchMock = jest.fn()
global.fetch = fetchMock as unknown as typeof fetch
const json = (data: unknown) => ({ ok: true, status: 200, json: async () => ({ success: true, data }) }) as Response

async function openSnapshot(currentFileID: number) {
  fetchMock.mockImplementation((input: RequestInfo | URL) => {
    const path = String(input)
    if (path.endsWith('/approvals/INDEPENDENT-A1')) return Promise.resolve(json({
      id: 'INDEPENDENT-A1', status: 'resubmitted', applicant_id: 9, current_approver_id: 2,
      available_actions: ['approve', 'reject', 'request_changes'], active_submission_snapshot_id: 'INDEPENDENT-S1',
      metadata: { intake_id: 'INDEPENDENT-I1', materials: [{ id: 5, name: '独立核对原合同', status: 'uploaded' }] },
    }))
    if (path.endsWith('/approvals/INDEPENDENT-A1/snapshot')) return Promise.resolve(json({
      approval_id: 'INDEPENDENT-A1', snapshot_id: 'INDEPENDENT-S1', snapshot: {
        intake_id: 'INDEPENDENT-I1', approval_id: 'INDEPENDENT-A1',
        material_fingerprints: [{ material_id: 5, name: '独立核对原合同', material_type: 'contract', required: true, status: 'uploaded', current_file_id: currentFileID, files: [
          { file_id: 12, file_name: '实际选定版本.pdf', digest: 'a'.repeat(64), size_bytes: 11 },
          { file_id: 13, file_name: '另一个历史版本.pdf', digest: 'b'.repeat(64), size_bytes: 12 },
        ] }],
      },
    }))
    if (path.includes('/case-materials/5/files/')) return Promise.resolve({ ok: true, blob: async () => new Blob(['isolated mock file']) } as Response)
    return Promise.resolve(json({ overall_status: 'pending', case_creation: {} }))
  })
  render(<MemoryRouter initialEntries={['/approval/INDEPENDENT-A1/decision']}><Routes><Route path='/approval/:id/decision' element={<ApprovalDecisionFlow />} /></Routes></MemoryRouter>)
  await screen.findByText('实际选定版本.pdf')
}

describe('Independent frozen current attachment selection', () => {
  beforeEach(() => { fetchMock.mockReset(); window.localStorage.clear() })

  it('downloads the explicitly frozen selected version, not the last file in history', async () => {
    await openSnapshot(12)
    fireEvent.click(screen.getByRole('button', { name: '下载快照版本' }))
    await waitFor(() => {
      expect(fetchMock.mock.calls.some(([url]) => String(url).endsWith('/case-materials/5/files/12/download'))).toBe(true)
    })
    expect(fetchMock.mock.calls.some(([url]) => String(url).endsWith('/case-materials/5/files/13/download'))).toBe(false)
    const download = fetchMock.mock.calls.find(([url]) => String(url).endsWith('/case-materials/5/files/12/download'))
    expect((download?.[1] as RequestInit).headers).toEqual(expect.objectContaining({ Authorization: 'Bearer independent-snapshot-token' }))
  })

  it('does not offer a guessed attachment when the explicit frozen pointer is invalid', async () => {
    await openSnapshot(999)
    expect(screen.queryByRole('button', { name: '下载快照版本' })).not.toBeInTheDocument()
    expect(fetchMock.mock.calls.some(([url]) => String(url).includes('/case-materials/5/files/'))).toBe(false)
  })
})
