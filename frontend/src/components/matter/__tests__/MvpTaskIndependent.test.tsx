import React from 'react'
import { act, fireEvent, render, screen } from '@testing-library/react'
import CaseTaskPanel from '../CaseTaskPanel'
import { message } from '@/utils/messageHelper'

jest.mock('@/utils/storage', () => ({ getToken: () => 'independent-test-token' }))
jest.mock('@/utils/messageHelper', () => ({
  message: { error: jest.fn(), success: jest.fn(), warning: jest.fn(), info: jest.fn() },
  setAppMessage: jest.fn(),
}))

const fetchMock = jest.fn()
global.fetch = fetchMock as unknown as typeof fetch

function response(data: unknown) {
  return { ok: true, status: 200, json: async () => ({ success: true, data }) } as Response
}

async function prepareDraft() {
  render(<CaseTaskPanel caseID={12} assigneeOptions={[{ id: 7, name: '独立测试律师' }]} currentUserId={7} />)
  await screen.findByText('暂无案件任务。可在下方创建通用任务事项。')
  fireEvent.change(screen.getByPlaceholderText('任务标题（必填）'), { target: { value: '不能丢失的原始任务' } })
  fireEvent.mouseDown(screen.getAllByRole('combobox')[0])
  fireEvent.click(await screen.findByText('独立测试律师'))
  await act(async () => { fireEvent.click(screen.getByRole('button', { name: /创建任务/ })) })
}

describe('Independent task creation proof', () => {
  beforeEach(() => {
    fetchMock.mockReset()
    jest.clearAllMocks()
  })

  it('does not accept a creation response without proof of the same case', async () => {
    fetchMock.mockImplementation((url: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === 'POST') return Promise.resolve(response({ task: { id: 71 } }))
      return Promise.resolve(response({ tasks: [] }))
    })
    await prepareDraft()
    expect(screen.getByPlaceholderText('任务标题（必填）')).toHaveValue('不能丢失的原始任务')
    expect(message.success).not.toHaveBeenCalled()
    expect(screen.getByRole('button', { name: /核对创建结果/ })).toBeInTheDocument()
  })

  it.each([
    ['missing case', { id: 71 }],
    ['zero task id', { id: 0, case_id: 12 }],
    ['different case', { id: 71, case_id: 99 }],
  ])('keeps the original draft when found returns %s', async (_label, task) => {
    fetchMock.mockImplementation((url: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === 'POST') return Promise.reject(new TypeError('isolated unknown outcome'))
      if (String(url).includes('/creation-result?')) return Promise.resolve(response({ state: 'found', task }))
      return Promise.resolve(response({ tasks: [] }))
    })
    await prepareDraft()
    await act(async () => { fireEvent.click(screen.getByRole('button', { name: /核对创建结果/ })) })
    expect(screen.getByPlaceholderText('任务标题（必填）')).toHaveValue('不能丢失的原始任务')
    expect(message.success).not.toHaveBeenCalled()
    expect(screen.getByRole('button', { name: /核对创建结果/ })).toBeInTheDocument()
  })

  it.each([400, 403])('allows correcting an explicit HTTP %s rejection without an unknown-outcome lock', async (status) => {
    fetchMock.mockImplementation((_url: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === 'POST') return Promise.resolve({
        ok: false,
        status,
        json: async () => ({ success: false, error: { code: 'CASE_TASK_ASSIGNEE_INVALID', message: '负责人不在该案件协作范围' } }),
      } as Response)
      return Promise.resolve(response({ tasks: [] }))
    })
    await prepareDraft()
    expect(screen.getByPlaceholderText('任务标题（必填）')).toHaveValue('不能丢失的原始任务')
    expect(screen.getByPlaceholderText('任务标题（必填）')).not.toBeDisabled()
    expect(screen.queryByRole('button', { name: /核对创建结果/ })).not.toBeInTheDocument()
    expect(message.success).not.toHaveBeenCalled()
    expect(screen.getByText(/负责人不在该案件协作范围/)).toBeInTheDocument()
  })

  it('does not abandon an unknown write merely because a momentary lookup has not found it', async () => {
    fetchMock.mockImplementation((url: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === 'POST') return Promise.reject(new TypeError('response lost while original server write may still be in flight'))
      if (String(url).includes('/creation-result?')) return Promise.resolve(response({ state: 'not_found' }))
      return Promise.resolve(response({ tasks: [] }))
    })
    await prepareDraft()
    await act(async () => { fireEvent.click(screen.getByRole('button', { name: /核对创建结果/ })) })
    // A read of committed rows does not cancel an original in-flight write.
    // Retry must therefore keep that original body and request, not expose
    // a new-key reset that could create a duplicate when the old write lands.
    expect(screen.getByPlaceholderText('任务标题（必填）')).toHaveValue('不能丢失的原始任务')
    expect(screen.getByPlaceholderText('任务标题（必填）')).toBeDisabled()
    expect(screen.getByRole('button', { name: '用相同内容重试' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '改为开始新任务' })).not.toBeInTheDocument()
    await act(async () => { fireEvent.click(screen.getByRole('button', { name: '用相同内容重试' })) })
    const writes = fetchMock.mock.calls.filter(([, init]) => init?.method === 'POST')
    expect(writes).toHaveLength(2)
    expect(writes[1][1].body).toBe(writes[0][1].body)
  })
})
