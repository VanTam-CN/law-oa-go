import React from 'react'
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, useNavigate } from 'react-router'
import { Modal } from 'antd'
import { CaseIntakeWorkbench } from '../Batch01Prototype'

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
function SwitchIntake() {
  const navigate = useNavigate()
  return <button onClick={() => navigate('/case/create?mode=revision&intake_id=I2&approval_id=A2')}>切换到另一份申请</button>
}
beforeEach(() => {
  window.localStorage.clear()
  Element.prototype.scrollIntoView = jest.fn()
  fetchMock.mockImplementation((input: RequestInfo | URL) => {
    const path = String(input)
    if (/\/case-intakes\/I[12]$/.test(path)) return Promise.resolve(json({
      intake: { id: path.endsWith('I2') ? 'I2' : 'I1', title: path.endsWith('I2') ? '第二份申请' : '第一份申请',
        client_id: 42, case_type: 'commercial', metadata: { lawyer_id: 7 } }, revision: 'rev-1', parties: [],
    }))
    if (/\/approvals\/A[12]$/.test(path)) return Promise.resolve(json({
      active_submission_snapshot_id: 'S1', records: [{ id: 'R1', decision: 'request_changes', imposed_requirements: { items: [
        { item_id: 'title', category: 'field_correction', field_key: 'case_title', description: '更正名称', required: true },
        { item_id: 'contact', category: 'field_correction', field_key: 'client_contact', description: '更正电话', required: true },
      ] } }],
    }))
    if (path.endsWith('/snapshot')) return Promise.resolve(json({ approval_id: path.includes('/A2/') ? 'A2' : 'A1', snapshot_id: 'S1', snapshot: {} }))
    return Promise.resolve(json({}))
  })
})
afterEach(() => { Modal.destroyAll() })

async function openAndArmTitle() {
  await act(async () => {
    render(<MemoryRouter initialEntries={['/case/create?mode=revision&intake_id=I1&approval_id=A1']}>
      <CaseIntakeWorkbench /><SwitchIntake />
    </MemoryRouter>)
  })
  const callbacks: Array<() => void> = []
  const realSetTimeout = window.setTimeout.bind(window)
  jest.spyOn(window, 'setTimeout').mockImplementation(((handler: TimerHandler, delay?: number, ...args: unknown[]) => {
    if (delay === 150 && typeof handler === 'function') {
      callbacks.push(handler as () => void)
      return 9000 + callbacks.length
    }
    return realSetTimeout(handler, delay, ...args)
  }) as typeof window.setTimeout)
  const buttons = await screen.findAllByRole('button', { name: '定位并填写该字段' })
  fireEvent.click(buttons[0])
  expect(callbacks).toHaveLength(1)
  return { callback: callbacks[0], buttons }
}

it('a client-archive detour cancels the previously armed field focus', async () => {
  const { callback, buttons } = await openAndArmTitle()
  fireEvent.click(buttons[1])
  expect(await screen.findByText('客户联系方式需在客户档案中更正')).toBeInTheDocument()
  act(() => callback())
  expect(document.activeElement?.id).not.toBe('intake-title')
})

it('an old locate callback cannot focus the next intake', async () => {
  const { callback } = await openAndArmTitle()
  await act(async () => { fireEvent.click(screen.getByRole('button', { name: '切换到另一份申请' })) })
  await waitFor(() => expect(screen.getByLabelText('案件名称 *')).toHaveValue('第二份申请'))
  act(() => callback())
  expect(document.activeElement?.id).not.toBe('intake-title')
})

it('clicking the current step manually cancels pending field focus too', async () => {
  const { callback } = await openAndArmTitle()
  fireEvent.click(document.querySelectorAll('.batch-stepper button')[0])
  act(() => callback())
  expect(document.activeElement?.id).not.toBe('intake-title')
})
