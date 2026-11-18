import React from 'react'
import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import CaseTaskPanel from '../CaseTaskPanel'
import { message } from '@/utils/messageHelper'

jest.mock('@/utils/storage', () => ({
  getToken: () => 'test-token',
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

const fullCapabilities = {
  can_adjust: true,
  can_cancel: true,
  can_complete: true,
  can_view_changes: true,
}

async function fillCreateDraft() {
  await screen.findByText('暂无案件任务。可在下方创建通用任务事项。')
  fireEvent.change(screen.getByPlaceholderText('任务标题（必填）'), {
    target: { value: '保留的任务标题' },
  })
  fireEvent.change(screen.getByPlaceholderText('补充说明（可选）'), {
    target: { value: '保留的补充说明' },
  })
  fireEvent.mouseDown(screen.getAllByRole('combobox')[0])
  fireEvent.click(await screen.findByText('律师甲'))
}

function expectCreateInputsLocked(locked: boolean) {
  const inputs = [
    screen.getByPlaceholderText('任务标题（必填）'),
    screen.getByPlaceholderText('补充说明（可选）'),
    screen.getAllByRole('combobox')[0],
    screen.getByPlaceholderText('截止（律师确认）'),
  ]
  for (const input of inputs) {
    if (locked) expect(input).toBeDisabled()
    else expect(input).not.toBeDisabled()
  }
}

describe('CaseTaskPanel', () => {
  beforeEach(() => {
    fetchMock.mockReset()
    ;(message.success as jest.Mock).mockClear()
    ;(message.error as jest.Mock).mockClear()
  })

  it('lists tasks with assignee, due date, and completion record', async () => {
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const path = String(input)
      if (path.endsWith('/cases/12/tasks')) {
        return Promise.resolve(
          jsonResponse({
            tasks: [
              {
                id: 5,
                case_id: 12,
                title: '整理证据目录',
                detail: '按时间顺序',
                assignee_id: 7,
                created_by: 7,
                due_date: '2026-10-05T00:00:00Z',
                status: 'completed',
                completed_by: 7,
                completed_at: '2026-09-29T10:00:00Z',
              },
              {
                id: 6,
                case_id: 12,
                title: '联系客户确认开庭时间',
                assignee_id: 8,
                created_by: 7,
                status: 'open',
              },
            ],
          }),
        )
      }
      return Promise.resolve(jsonResponse({}))
    })

    render(
      <CaseTaskPanel
        caseID={12}
        assigneeOptions={[
          { id: 7, name: '律师甲' },
          { id: 8, name: '律师乙' },
        ]}
        currentUserId={7}
      />,
    )

    expect(await screen.findByText('整理证据目录')).toBeInTheDocument()
    expect(screen.getByText('联系客户确认开庭时间')).toBeInTheDocument()
    expect(screen.getByText('律师甲')).toBeInTheDocument()
    expect(screen.getByText('律师乙')).toBeInTheDocument()
    expect(screen.getByText('2026-10-05')).toBeInTheDocument()
    // Completion record renders with who and when (day is TZ-independent;
    // the hour part depends on the runner timezone).
    expect(screen.getByText(/律师甲 · 2026-09-29/)).toBeInTheDocument()
    // Open task shows an enabled complete button for the assignee.
    const complete = screen.getByRole('button', { name: /去完成/ })
    expect(complete).not.toBeDisabled()
    // Completed task shows no complete button.
    expect(screen.getAllByRole('button', { name: /去完成/ })).toHaveLength(1)
  })

  it('creates a task with idempotency key and reloads', async () => {
    let created = false
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      const method = String(init?.method || 'GET').toUpperCase()
      if (path.endsWith('/cases/12/tasks') && method === 'GET') {
        return Promise.resolve(
          jsonResponse({ tasks: created ? [{ id: 9, case_id: 12, title: '新任务', assignee_id: 7, created_by: 7, status: 'open' }] : [] }),
        )
      }
      if (path.endsWith('/cases/12/tasks') && method === 'POST') {
        created = true
        const body = JSON.parse(String(init?.body))
        expect(body.title).toBe('新任务')
        expect(body.assignee_id).toBe(7)
        expect(body.idempotency_key).toBeTruthy()
        return Promise.resolve(
          jsonResponse({
            task: { id: 9, case_id: 12, title: '新任务', assignee_id: 7, created_by: 7, status: 'open' },
          }),
        )
      }
      return Promise.resolve(jsonResponse({}))
    })

    render(
      <CaseTaskPanel
        caseID={12}
        assigneeOptions={[{ id: 7, name: '律师甲' }]}
        currentUserId={7}
      />,
    )
    await screen.findByText('暂无案件任务。可在下方创建通用任务事项。')

    fireEvent.change(screen.getByPlaceholderText('任务标题（必填）'), {
      target: { value: '新任务' },
    })
    const assigneeCombo = screen.getAllByRole('combobox')[0]
    fireEvent.mouseDown(assigneeCombo)
    const option = await screen.findByText('律师甲')
    fireEvent.click(option)
    fireEvent.click(screen.getByRole('button', { name: /创建任务/ }))

    await waitFor(() =>
      expect(fetchMock.mock.calls.some(([input, init]) => String(input).endsWith('/cases/12/tasks') && String(init?.method) === 'POST')).toBe(true),
    )
    // findByText returns a promise: await it, then assert against the real DOM.
    // findByText('新任务') would also match the form label, so assert the
    // actual reloaded row instead.
    expect(await screen.findByTestId('task-row-9')).toBeInTheDocument()
    expect(within(screen.getByTestId('task-row-9')).getByText('新任务')).toBeInTheDocument()
  })

  it('sends completion_note when completing a task', async () => {
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      const method = String(init?.method || 'GET').toUpperCase()
      if (path.endsWith('/cases/12/tasks') && method === 'GET') {
        return Promise.resolve(
          jsonResponse({
            tasks: [
              { id: 5, case_id: 12, title: '待完成任务', assignee_id: 7, created_by: 7, status: 'open' },
            ],
          }),
        )
      }
      if (path.endsWith('/case-tasks/5/complete') && method === 'POST') {
        const body = JSON.parse(String(init?.body))
        expect(body.completion_note).toBe('已整理并交付客户确认')
        return Promise.resolve(jsonResponse({ recorded: true }))
      }
      return Promise.resolve(jsonResponse({ tasks: [] }))
    })

    render(
      <CaseTaskPanel
        caseID={12}
        assigneeOptions={[{ id: 7, name: '律师甲' }]}
        currentUserId={7}
      />,
    )
    await screen.findByText('待完成任务')

    // The completion window opens directly from one click: the former
    // Popconfirm intermediate step was merged into the modal.
    fireEvent.click(screen.getByRole('button', { name: /去完成/ }))
    const noteBox = await screen.findByPlaceholderText('完成记录（可选）：做了什么、结论或下一步')
    fireEvent.change(noteBox, { target: { value: '已整理并交付客户确认' } })
    fireEvent.click(screen.getByRole('button', { name: '确认完成' }))

    await waitFor(() =>
      expect(
        fetchMock.mock.calls.some(([input, init]) =>
          String(input).endsWith('/case-tasks/5/complete') &&
          String(init?.method) === 'POST',
        ),
      ).toBe(true),
    )
  })

  it('shows the real error envelope message and keeps the idempotency key on 409', async () => {
    const postBodies: string[] = []
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      const method = String(init?.method || 'GET').toUpperCase()
      if (path.endsWith('/cases/12/tasks') && method === 'GET') {
        return Promise.resolve(jsonResponse({ tasks: [] }))
      }
      if (path.endsWith('/cases/12/tasks') && method === 'POST') {
        postBodies.push(String(init?.body))
        return Promise.resolve({
          ok: false,
          status: 409,
          json: async () => ({
            success: false,
            error: { code: 'IDEMPOTENCY_CONFLICT', message: '相同幂等键的请求已处理' },
          }),
        } as Response)
      }
      return Promise.resolve(jsonResponse({}))
    })

    render(
      <CaseTaskPanel
        caseID={12}
        assigneeOptions={[{ id: 7, name: '律师甲' }]}
        currentUserId={7}
      />,
    )
    await screen.findByText('暂无案件任务。可在下方创建通用任务事项。')

    fireEvent.change(screen.getByPlaceholderText('任务标题（必填）'), {
      target: { value: '幂等冲突任务' },
    })
    fireEvent.mouseDown(screen.getAllByRole('combobox')[0])
    fireEvent.click(await screen.findByText('律师甲'))
    fireEvent.click(screen.getByRole('button', { name: /创建任务/ }))

    // The real server message surfaces, and a 409 is treated as a possibly
    // consumed key: the verification entry appears instead of an instant reset.
    expect(await screen.findByText(/相同幂等键的请求已处理/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /核对创建结果/ })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '开始新任务' })).not.toBeInTheDocument()

    // The form is locked after a 409: clicking 创建任务 again cannot silently
    // resend with a stale key (button disabled), so no second POST fires.
    const createButton = screen.getByRole('button', { name: /创建任务/ })
    expect(createButton).toBeDisabled()
    fireEvent.click(createButton)
    await waitFor(() => expect(postBodies).toHaveLength(1))

    // The original input is preserved for the verification flow.
    expect(screen.getByPlaceholderText('任务标题（必填）')).toHaveValue('幂等冲突任务')
  })

  it.each([400, 403])('allows correcting a structured HTTP %s rejection before manual resubmission', async (status) => {
    const postBodies: string[] = []
    fetchMock.mockImplementation((_input: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === 'POST') {
        postBodies.push(String(init.body))
        if (postBodies.length === 1) {
          return Promise.resolve({
            ok: false,
            status,
            json: async () => ({
              success: false,
              error: { code: 'CASE_TASK_ASSIGNEE_INVALID', message: '负责人不在该案件协作范围' },
            }),
          } as Response)
        }
        return Promise.resolve(jsonResponse({ task: { id: 71, case_id: 12 } }))
      }
      return Promise.resolve(jsonResponse({ tasks: [] }))
    })

    render(
      <CaseTaskPanel caseID={12} assigneeOptions={[{ id: 7, name: '律师甲' }, { id: 8, name: '律师乙' }]} currentUserId={7} />,
    )
    await fillCreateDraft()
    fireEvent.click(screen.getByRole('button', { name: /创建任务/ }))

    expect(await screen.findByText('负责人不在该案件协作范围')).toBeInTheDocument()
    expectCreateInputsLocked(false)
    expect(screen.getByPlaceholderText('任务标题（必填）')).toHaveValue('保留的任务标题')
    expect(screen.getByPlaceholderText('补充说明（可选）')).toHaveValue('保留的补充说明')
    expect(screen.getByRole('button', { name: /创建任务/ })).not.toBeDisabled()
    expect(screen.queryByRole('button', { name: /核对创建结果/ })).not.toBeInTheDocument()
    expect(message.error).toHaveBeenCalledWith('负责人不在该案件协作范围')
    expect(message.success).not.toHaveBeenCalled()

    fireEvent.change(screen.getByPlaceholderText('任务标题（必填）'), {
      target: { value: '更正后的任务标题' },
    })
    fireEvent.mouseDown(screen.getAllByRole('combobox')[0])
    fireEvent.click(await screen.findByText('律师乙'))
    expect(postBodies).toHaveLength(1)
    fireEvent.click(screen.getByRole('button', { name: /创建任务/ }))
    await waitFor(() => expect(message.success).toHaveBeenCalledWith('任务已创建，待办同步生成'))
    expect(JSON.parse(postBodies[1])).toEqual({
      ...JSON.parse(postBodies[0]),
      title: '更正后的任务标题',
      assignee_id: 8,
    })
  })

  it.each([400, 403])('unlocks corrections after a verified not_found retry receives HTTP %s', async (status) => {
    const postBodies: string[] = []
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === 'POST') {
        postBodies.push(String(init.body))
        if (postBodies.length === 1) return Promise.reject(new TypeError('Failed to fetch'))
        if (postBodies.length === 2) {
          return Promise.resolve({
            ok: false,
            status,
            json: async () => ({
              success: false,
              error: { code: 'CASE_TASK_ASSIGNEE_INVALID', message: '负责人不在该案件协作范围' },
            }),
          } as Response)
        }
        return Promise.resolve(jsonResponse({ task: { id: 71, case_id: 12 } }))
      }
      if (String(input).includes('/creation-result?')) {
        return Promise.resolve(jsonResponse({ state: 'not_found' }))
      }
      return Promise.resolve(jsonResponse({ tasks: [] }))
    })

    render(
      <CaseTaskPanel caseID={12} assigneeOptions={[{ id: 7, name: '律师甲' }]} currentUserId={7} />,
    )
    await fillCreateDraft()
    fireEvent.click(screen.getByRole('button', { name: /创建任务/ }))
    fireEvent.click(await screen.findByRole('button', { name: /核对创建结果/ }))
    fireEvent.click(await screen.findByRole('button', { name: '用相同内容重试' }))
    expect(await screen.findByText('负责人不在该案件协作范围')).toBeInTheDocument()
    expectCreateInputsLocked(false)
    expect(postBodies[1]).toBe(postBodies[0])
    expect(screen.getByPlaceholderText('任务标题（必填）')).toHaveValue('保留的任务标题')
    expect(screen.getByPlaceholderText('补充说明（可选）')).toHaveValue('保留的补充说明')
    expect(screen.getByRole('button', { name: /创建任务/ })).not.toBeDisabled()
    expect(screen.queryByRole('button', { name: /核对创建结果/ })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '用相同内容重试' })).not.toBeInTheDocument()

    fireEvent.change(screen.getByPlaceholderText('补充说明（可选）'), {
      target: { value: '手动更正后的说明' },
    })
    expect(postBodies).toHaveLength(2)
    fireEvent.click(screen.getByRole('button', { name: /创建任务/ }))
    await waitFor(() => expect(postBodies).toHaveLength(3))
    expect(JSON.parse(postBodies[2])).toEqual({
      ...JSON.parse(postBodies[0]),
      detail: '手动更正后的说明',
    })
  })

  it.each([500, 504, 'network', 'malformed 200', 409])('preserves the original request and key for %s until verification', async (failure) => {
    const postBodies: string[] = []
    const checkUrls: string[] = []
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === 'POST') {
        postBodies.push(String(init.body))
        if (postBodies.length > 1) {
          return Promise.resolve(jsonResponse({ task: { id: 71, case_id: 12 } }))
        }
        if (failure === 'network') return Promise.reject(new TypeError('网络失败（400）'))
        if (failure === 'malformed 200') {
          return Promise.resolve(jsonResponse({ task: { id: 0, case_id: 12 } }))
        }
        return Promise.resolve({
          ok: false,
          status: failure,
          json: async () => ({
            success: false,
            error: { code: 'CASE_TASK_ASSIGNEE_INVALID', message: '错误消息不能决定 HTTP 状态（400）' },
          }),
        } as Response)
      }
      if (String(input).includes('/creation-result?')) {
        checkUrls.push(String(input))
        return Promise.resolve(jsonResponse({ state: checkUrls.length === 1 ? 'unknown' : 'not_found' }))
      }
      return Promise.resolve(jsonResponse({ tasks: [] }))
    })

    render(
      <CaseTaskPanel caseID={12} assigneeOptions={[{ id: 7, name: '律师甲' }]} currentUserId={7} />,
    )
    await fillCreateDraft()
    fireEvent.click(screen.getByRole('button', { name: /创建任务/ }))
    await screen.findByRole('button', { name: /核对创建结果/ })
    expectCreateInputsLocked(true)
    expect(screen.getByPlaceholderText('任务标题（必填）')).toHaveValue('保留的任务标题')
    expect(screen.getByPlaceholderText('补充说明（可选）')).toHaveValue('保留的补充说明')
    expect(screen.getByRole('button', { name: /创建任务/ })).toBeDisabled()
    expect(message.success).not.toHaveBeenCalled()
    const originalKey = JSON.parse(postBodies[0]).idempotency_key

    fireEvent.click(screen.getByRole('button', { name: /核对创建结果/ }))
    expect(await screen.findByText(/暂时无法确认创建结果/)).toBeInTheDocument()
    expect(checkUrls[0]).toContain(encodeURIComponent(originalKey))
    expect(postBodies).toHaveLength(1)
    expect(screen.queryByRole('button', { name: '用相同内容重试' })).not.toBeInTheDocument()

    await userEvent.type(screen.getByPlaceholderText('任务标题（必填）'), '核对期间编辑的标题')
    await userEvent.type(screen.getByPlaceholderText('补充说明（可选）'), '核对期间编辑的说明')
    expect(screen.getByPlaceholderText('任务标题（必填）')).toHaveValue('保留的任务标题')
    expect(screen.getByPlaceholderText('补充说明（可选）')).toHaveValue('保留的补充说明')
    fireEvent.click(screen.getByRole('button', { name: /核对创建结果/ }))
    await screen.findByRole('button', { name: '用相同内容重试' })
    expectCreateInputsLocked(true)
    expect(screen.queryByRole('button', { name: '改为开始新任务' })).not.toBeInTheDocument()
    expect(checkUrls[1]).toBe(checkUrls[0])
    expect(postBodies).toHaveLength(1)
    fireEvent.click(screen.getByRole('button', { name: '用相同内容重试' }))
    await waitFor(() => expect(message.success).toHaveBeenCalledWith('任务已创建，待办同步生成'))
    expect(postBodies[1]).toBe(postBodies[0])
  })

  it('keeps pending inputs and the original request locked through momentary not_found until a late write is found', async () => {
    const postBodies: string[] = []
    const checkUrls: string[] = []
    const notFoundMessage = '本次查询暂未找到原任务。可用原内容重试；原请求可能仍在处理，请勿重新建同一任务。'
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === 'POST') {
        postBodies.push(String(init.body))
        return Promise.reject(new TypeError('response lost while original write is still processing'))
      }
      if (String(input).includes('/creation-result?')) {
        checkUrls.push(String(input))
        return Promise.resolve(jsonResponse(checkUrls.length <= 2
          ? { state: 'not_found' }
          : { state: 'found', task: { id: 71, case_id: 12 } }))
      }
      return Promise.resolve(jsonResponse({ tasks: [] }))
    })

    render(
      <CaseTaskPanel caseID={12} assigneeOptions={[{ id: 7, name: '律师甲' }]} currentUserId={7} />,
    )
    await fillCreateDraft()
    fireEvent.click(screen.getByRole('button', { name: /创建任务/ }))
    await screen.findByRole('button', { name: /核对创建结果/ })
    expectCreateInputsLocked(true)
    const originalKey = JSON.parse(postBodies[0]).idempotency_key

    for (let attempt = 1; attempt <= 2; attempt += 1) {
      await act(async () => {
        fireEvent.click(screen.getByRole('button', { name: /核对创建结果/ }))
      })
      expect(screen.getByText(notFoundMessage)).toBeInTheDocument()
      expect(checkUrls).toHaveLength(attempt)
      expect(checkUrls[attempt - 1]).toContain(encodeURIComponent(originalKey))
      expectCreateInputsLocked(true)
      expect(screen.getByPlaceholderText('任务标题（必填）')).toHaveValue('保留的任务标题')
      expect(screen.getByPlaceholderText('补充说明（可选）')).toHaveValue('保留的补充说明')
      expect(screen.getByRole('button', { name: /创建任务/ })).toBeDisabled()
      expect(screen.getByRole('button', { name: '用相同内容重试' })).toBeInTheDocument()
      expect(screen.queryByRole('button', { name: '改为开始新任务' })).not.toBeInTheDocument()
      expect(postBodies).toHaveLength(1)
      expect(message.success).not.toHaveBeenCalled()
    }

    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: '用相同内容重试' }))
    })
    expect(postBodies).toHaveLength(2)
    expect(postBodies[1]).toBe(postBodies[0])
    expect(JSON.parse(postBodies[1]).idempotency_key).toBe(originalKey)
    expectCreateInputsLocked(true)
    expect(screen.getByPlaceholderText('任务标题（必填）')).toHaveValue('保留的任务标题')
    expect(screen.getByPlaceholderText('补充说明（可选）')).toHaveValue('保留的补充说明')

    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: /核对创建结果/ }))
    })
    expect(checkUrls).toHaveLength(3)
    expect(new Set(checkUrls).size).toBe(1)
    expect(message.success).toHaveBeenCalledWith('已找到先前创建的任务')
    expectCreateInputsLocked(false)
    expect(screen.getByPlaceholderText('任务标题（必填）')).toHaveValue('')
    expect(screen.getByPlaceholderText('补充说明（可选）')).toHaveValue('')
    expect(screen.getByRole('button', { name: /创建任务/ })).not.toBeDisabled()
    expect(postBodies).toHaveLength(2)
  })

  it('disables actions by server capabilities even for the creator', async () => {
    // Server truth: this creator can neither complete nor adjust; the old
    // created_by inference would have rendered an enabled button.
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const path = String(input)
      if (path.endsWith('/cases/12/tasks')) {
        return Promise.resolve(
          jsonResponse({
            tasks: [
              {
                id: 5,
                case_id: 12,
                title: '受限任务',
                assignee_id: 7,
                created_by: 7,
                status: 'open',
                version: 1,
                capabilities: {
                  can_adjust: false,
                  can_cancel: false,
                  can_complete: false,
                  can_view_changes: false,
                },
              },
            ],
          }),
        )
      }
      return Promise.resolve(jsonResponse({}))
    })

    render(
      <CaseTaskPanel
        caseID={12}
        assigneeOptions={[{ id: 7, name: '律师甲' }]}
        currentUserId={7}
      />,
    )
    await screen.findByText('受限任务')
    expect(screen.queryByRole('button', { name: /去完成/ })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /调整/ })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /取消/ })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /变更记录/ })).not.toBeInTheDocument()
  })

  it('renders manager actions from server capabilities beyond created_by', async () => {
    // currentUserId=9 is neither assignee nor creator; the server granted
    // case-manager capabilities, so the buttons must render.
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const path = String(input)
      if (path.endsWith('/cases/12/tasks')) {
        return Promise.resolve(
          jsonResponse({
            tasks: [
              {
                id: 6,
                case_id: 12,
                title: '管理者任务',
                assignee_id: 7,
                created_by: 7,
                status: 'open',
                version: 2,
                capabilities: {
                  can_adjust: true,
                  can_cancel: true,
                  can_complete: true,
                  can_view_changes: true,
                },
              },
            ],
          }),
        )
      }
      if (path.endsWith('/case-tasks/6/changes')) {
        return Promise.resolve(
          jsonResponse({
            task: { id: 6 },
            version: 2,
            changes: [
              {
                id: 1,
                created_at: '2026-10-02T09:00:00Z',
                change_type: 'reassign',
                actor_id: 9,
                old_assignee_id: 7,
                new_assignee_id: 8,
                old_version: 1,
                new_version: 2,
                reason: '工作分工调整',
              },
            ],
          }),
        )
      }
      return Promise.resolve(jsonResponse({}))
    })

    render(
      <CaseTaskPanel
        caseID={12}
        assigneeOptions={[
          { id: 7, name: '律师甲' },
          { id: 8, name: '律师乙' },
          { id: 9, name: '主管' },
        ]}
        currentUserId={9}
      />,
    )
    await screen.findByText('管理者任务')
    expect(screen.getByRole('button', { name: /去完成/ })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /调整/ })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /取消/ })).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /变更记录/ }))
    expect(await screen.findByText(/工作分工调整/)).toBeInTheDocument()
    expect(screen.getByText(/负责人：律师甲 → 律师乙/)).toBeInTheDocument()
  })

  it('opens the completion window in one click', async () => {
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      const method = String(init?.method || 'GET').toUpperCase()
      if (path.endsWith('/cases/12/tasks') && method === 'GET') {
        return Promise.resolve(
          jsonResponse({
            tasks: [
              {
                id: 5,
                case_id: 12,
                title: '单击完成任务',
                assignee_id: 7,
                created_by: 7,
                status: 'open',
                capabilities: {
                  can_adjust: false,
                  can_cancel: false,
                  can_complete: true,
                  can_view_changes: false,
                },
              },
            ],
          }),
        )
      }
      return Promise.resolve(jsonResponse({}))
    })

    render(
      <CaseTaskPanel
        caseID={12}
        assigneeOptions={[{ id: 7, name: '律师甲' }]}
        currentUserId={7}
      />,
    )
    await screen.findByText('单击完成任务')
    fireEvent.click(screen.getByRole('button', { name: /去完成/ }))
    // No intermediate Popconfirm step: the note window is already open.
    expect(await screen.findByText('完成后关联待办会一并关闭。')).toBeInTheDocument()
  })

  it('keeps the same idempotency key after a network timeout and retries the same body once confirmed not_found', async () => {
    const postBodies: string[] = []
    let creationResultUrl = ''
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      const method = String(init?.method || 'GET').toUpperCase()
      if (path.endsWith('/cases/12/tasks') && method === 'GET') {
        return Promise.resolve(jsonResponse({ tasks: [] }))
      }
      if (path.endsWith('/cases/12/tasks') && method === 'POST') {
        postBodies.push(String(init?.body))
        if (postBodies.length === 1) {
          // Network-level failure: the server outcome is unknown.
          throw new TypeError('Failed to fetch')
        }
        return Promise.resolve(jsonResponse({ task: { id: 9 } }))
      }
      if (path.includes('/cases/12/tasks/creation-result')) {
        creationResultUrl = path
        return Promise.resolve(jsonResponse({ state: 'not_found' }))
      }
      return Promise.resolve(jsonResponse({}))
    })

    render(
      <CaseTaskPanel
        caseID={12}
        assigneeOptions={[{ id: 7, name: '律师甲' }]}
        currentUserId={7}
      />,
    )
    await screen.findByText('暂无案件任务。可在下方创建通用任务事项。')
    fireEvent.change(screen.getByPlaceholderText('任务标题（必填）'), {
      target: { value: '超时核对任务' },
    })
    fireEvent.mouseDown(screen.getAllByRole('combobox')[0])
    fireEvent.click(await screen.findByText('律师甲'))
    fireEvent.click(screen.getByRole('button', { name: /创建任务/ }))

    // The inputs and the pending key are kept; no silent retry happens.
    expect(await screen.findByText(/创建请求未收到服务器确认/)).toBeInTheDocument()
    expect(screen.getByPlaceholderText('任务标题（必填）')).toHaveValue('超时核对任务')
    // getByRole matches include the icon's accessible label, so anchor on
    // the full accessible name "reload 核对创建结果".
    expect(postBodies).toHaveLength(1)

    fireEvent.click(screen.getByRole('button', { name: /核对创建结果/ }))
    expect(await screen.findByText(/本次查询暂未找到原任务/)).toBeInTheDocument()
    const firstKey = JSON.parse(postBodies[0]).idempotency_key
    expect(creationResultUrl).toContain(encodeURIComponent(firstKey))

    // Retry with the IDENTICAL body and key: only the explicit user action
    // after not_found may replay the original in-flight request.
    fireEvent.click(screen.getByRole('button', { name: '用相同内容重试' }))
    await waitFor(() => expect(postBodies).toHaveLength(2))
    expect(JSON.parse(postBodies[1])).toEqual(JSON.parse(postBodies[0]))
  })

  it('never rotates the key or auto-retries on an unknown outcome', async () => {
    const postBodies: string[] = []
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      const method = String(init?.method || 'GET').toUpperCase()
      if (path.endsWith('/cases/12/tasks') && method === 'GET') {
        return Promise.resolve(jsonResponse({ tasks: [] }))
      }
      if (path.endsWith('/cases/12/tasks') && method === 'POST') {
        postBodies.push(String(init?.body))
        throw new TypeError('Failed to fetch')
      }
      if (path.includes('/cases/12/tasks/creation-result')) {
        return Promise.resolve(jsonResponse({ state: 'unknown' }))
      }
      return Promise.resolve(jsonResponse({}))
    })

    render(
      <CaseTaskPanel
        caseID={12}
        assigneeOptions={[{ id: 7, name: '律师甲' }]}
        currentUserId={7}
      />,
    )
    await screen.findByText('暂无案件任务。可在下方创建通用任务事项。')
    fireEvent.change(screen.getByPlaceholderText('任务标题（必填）'), {
      target: { value: '未知结果任务' },
    })
    fireEvent.mouseDown(screen.getAllByRole('combobox')[0])
    fireEvent.click(await screen.findByText('律师甲'))
    fireEvent.click(screen.getByRole('button', { name: /创建任务/ }))

    expect(await screen.findByText(/创建请求未收到服务器确认/)).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: /核对创建结果/ }))
    expect(await screen.findByText(/暂时无法确认创建结果/)).toBeInTheDocument()
    // Unknown outcome: no retry entry, no automatic second POST. The same
    // key stays pending for the next manual check.
    expect(screen.queryByRole('button', { name: '用相同内容重试' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '改为开始新任务' })).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: /核对创建结果/ })).toBeInTheDocument()
    await waitFor(() => expect(postBodies).toHaveLength(1))
  })

  it('aborts stale list responses when switching cases quickly', async () => {
    let resolveCase12: ((value: unknown) => void) | undefined
    const case12Promise = new Promise((resolve) => {
      resolveCase12 = resolve
    })
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const path = String(input)
      if (path.endsWith('/cases/12/tasks')) {
        return case12Promise as Promise<Response>
      }
      if (path.endsWith('/cases/13/tasks')) {
        return Promise.resolve(
          jsonResponse({
            tasks: [
              {
                id: 31,
                case_id: 13,
                title: '新案任务',
                assignee_id: 7,
                created_by: 7,
                status: 'open',
              },
            ],
          }),
        )
      }
      return Promise.resolve(jsonResponse({}))
    })

    const { rerender } = render(
      <CaseTaskPanel
        caseID={12}
        assigneeOptions={[{ id: 7, name: '律师甲' }]}
        currentUserId={7}
      />,
    )
    rerender(
      <CaseTaskPanel
        caseID={13}
        assigneeOptions={[{ id: 7, name: '律师甲' }]}
        currentUserId={7}
      />,
    )
    expect(await screen.findByText('新案任务')).toBeInTheDocument()
    // The slow case-12 response lands AFTER the switch; it must be dropped
    // instead of rendering cross-case rows.
    resolveCase12!(
      jsonResponse({
        tasks: [
          {
            id: 99,
            case_id: 12,
            title: '旧案串扰任务',
            assignee_id: 7,
            created_by: 7,
            status: 'open',
          },
        ],
      }),
    )
    await waitFor(() => {
      expect(screen.queryByText('旧案串扰任务')).not.toBeInTheDocument()
    })
    expect(screen.getByText('新案任务')).toBeInTheDocument()
  })

  it('keeps the newest history when an older changes response lands late', async () => {
    let resolveFirst: ((value: unknown) => void) | undefined
    const firstPromise = new Promise((resolve) => {
      resolveFirst = resolve
    })
    let changesCalls = 0
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const path = String(input)
      if (path.endsWith('/cases/12/tasks')) {
        return Promise.resolve(
          jsonResponse({
            tasks: [
              {
                id: 5,
                case_id: 12,
                title: '历史逆序任务',
                assignee_id: 7,
                created_by: 7,
                status: 'open',
                version: 3,
                capabilities: { ...fullCapabilities, can_complete: false },
              },
            ],
          }),
        )
      }
      if (path.endsWith('/case-tasks/5/changes')) {
        changesCalls += 1
        // Only the FIRST request hangs so the second one can finish first;
        // sharing one promise between both requests would deadlock the test.
        if (changesCalls === 1) {
          return firstPromise as unknown as Response
        }
        return Promise.resolve(jsonResponse({ changes: [] }))
      }
      return Promise.resolve(jsonResponse({}))
    })

    render(
      <CaseTaskPanel caseID={12} assigneeOptions={[{ id: 7, name: '律师甲' }]} currentUserId={7} />,
    )
    await screen.findByText('历史逆序任务')

    fireEvent.click(screen.getByRole('button', { name: /变更记录/ }))
    // The first request hangs; collapse and expand again to fire a second,
    // newer request.
    fireEvent.click(screen.getByRole('button', { name: /收起变更/ }))
    fireEvent.click(screen.getByRole('button', { name: /变更记录/ }))
    await waitFor(() => expect(screen.getByText('暂无变更记录。')).toBeInTheDocument())

    // The first (stale) response lands after the newest one finished.
    resolveFirst!(
      jsonResponse({
        changes: [
          {
            id: 1,
            created_at: '2026-10-01T00:00:00Z',
            change_type: 'reassign',
            actor_id: 7,
            old_assignee_id: 7,
            new_assignee_id: 8,
          },
        ],
      }),
    )
    await waitFor(() => {
      expect(screen.queryByText(/律师甲 → 律师乙/)).not.toBeInTheDocument()
    })
    expect(screen.getByText('暂无变更记录。')).toBeInTheDocument()
  })

  it('drops late create-verification responses after switching cases', async () => {
    let resolveCheck: ((value: unknown) => void) | undefined
    const checkPromise = new Promise((resolve) => {
      resolveCheck = resolve
    })
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      const method = String(init?.method || 'GET').toUpperCase()
      if (path.endsWith('/cases/12/tasks') && method === 'GET') {
        return Promise.resolve(jsonResponse({ tasks: [] }))
      }
      if (path.endsWith('/cases/12/tasks') && method === 'POST') {
        throw new TypeError('Failed to fetch')
      }
      if (path.includes('/cases/12/tasks/creation-result')) {
        return checkPromise as Promise<Response>
      }
      if (path.endsWith('/cases/13/tasks') && method === 'GET') {
        return Promise.resolve(jsonResponse({ tasks: [] }))
      }
      return Promise.resolve(jsonResponse({}))
    })

    const { rerender } = render(
      <CaseTaskPanel caseID={12} assigneeOptions={[{ id: 7, name: '律师甲' }]} currentUserId={7} />,
    )
    await screen.findByText('暂无案件任务。可在下方创建通用任务事项。')
    fireEvent.change(screen.getByPlaceholderText('任务标题（必填）'), {
      target: { value: '跨案核对任务' },
    })
    fireEvent.mouseDown(screen.getAllByRole('combobox')[0])
    fireEvent.click(await screen.findByText('律师甲'))
    fireEvent.click(screen.getByRole('button', { name: /创建任务/ }))
    expect(await screen.findByText(/创建请求未收到服务器确认/)).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: /核对创建结果/ }))

    // Switch cases while the check is still hanging.
    rerender(
      <CaseTaskPanel caseID={13} assigneeOptions={[{ id: 7, name: '律师甲' }]} currentUserId={7} />,
    )
    await act(async () => {
      resolveCheck!(
        jsonResponse({
          state: 'found',
          task: { id: 77, case_id: 12, title: '跨案核对任务' },
        }),
      )
      await Promise.resolve()
    })
    expect(screen.queryByText(/已找到先前创建的任务/)).not.toBeInTheDocument()
    expect(screen.queryByTestId('task-row-77')).not.toBeInTheDocument()
  })

  it('treats HTTP 504 and malformed 200 as unknown without clearing the form', async () => {
    const postBodies: string[] = []
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      const method = String(init?.method || 'GET').toUpperCase()
      if (path.endsWith('/cases/12/tasks') && method === 'GET') {
        return Promise.resolve(jsonResponse({ tasks: [] }))
      }
      if (path.endsWith('/cases/12/tasks') && method === 'POST') {
        postBodies.push(String(init?.body))
        if (postBodies.length === 1) {
          return Promise.resolve({
            ok: false,
            status: 504,
            json: async () => ({}),
          } as Response)
        }
        return Promise.resolve({ ok: true, json: async () => ({ success: true }) } as Response)
      }
      if (path.includes('/cases/12/tasks/creation-result')) {
        return Promise.resolve(jsonResponse({ state: 'not_found' }))
      }
      return Promise.resolve(jsonResponse({}))
    })

    render(
      <CaseTaskPanel caseID={12} assigneeOptions={[{ id: 7, name: '律师甲' }]} currentUserId={7} />,
    )
    await screen.findByText('暂无案件任务。可在下方创建通用任务事项。')
    fireEvent.change(screen.getByPlaceholderText('任务标题（必填）'), {
      target: { value: '504与坏200任务' },
    })
    fireEvent.mouseDown(screen.getAllByRole('combobox')[0])
    fireEvent.click(await screen.findByText('律师甲'))
    fireEvent.click(screen.getByRole('button', { name: /创建任务/ }))

    // 504 is a gateway-level failure: unknown, form kept, key kept.
    expect(await screen.findByText(/创建请求未收到服务器确认/)).toBeInTheDocument()
    expect(screen.getByPlaceholderText('任务标题（必填）')).toHaveValue('504与坏200任务')
    expect(screen.getByRole('button', { name: /核对创建结果/ })).toBeInTheDocument()
    expect(postBodies).toHaveLength(1)

    // After a momentary not_found the user may replay the same body and
    // key; a 200 whose payload lacks a usable task is also unknown, not
    // success, and must not clear the form.
    fireEvent.click(screen.getByRole('button', { name: /核对创建结果/ }))
    expect(await screen.findByText(/本次查询暂未找到原任务/)).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: '用相同内容重试' }))
    expect(await screen.findByText(/服务器返回内容异常/)).toBeInTheDocument()
    expect(screen.getByPlaceholderText('任务标题（必填）')).toHaveValue('504与坏200任务')
    expect(postBodies).toHaveLength(2)
    expect(JSON.parse(postBodies[1])).toEqual(JSON.parse(postBodies[0]))
  })

  it('keeps unknown and the original key when found carries a bad payload', async () => {
    const postBodies: string[] = []
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      const method = String(init?.method || 'GET').toUpperCase()
      if (path.endsWith('/cases/12/tasks') && method === 'GET') {
        return Promise.resolve(jsonResponse({ tasks: [] }))
      }
      if (path.endsWith('/cases/12/tasks') && method === 'POST') {
        postBodies.push(String(init?.body))
        throw new TypeError('Failed to fetch')
      }
      if (path.includes('/cases/12/tasks/creation-result')) {
        // state=found but the payload is not a usable same-case task.
        return Promise.resolve(jsonResponse({ state: 'found', task: { title: '坏数据' } }))
      }
      return Promise.resolve(jsonResponse({}))
    })

    render(
      <CaseTaskPanel caseID={12} assigneeOptions={[{ id: 7, name: '律师甲' }]} currentUserId={7} />,
    )
    await screen.findByText('暂无案件任务。可在下方创建通用任务事项。')
    fireEvent.change(screen.getByPlaceholderText('任务标题（必填）'), {
      target: { value: 'found坏负载任务' },
    })
    fireEvent.mouseDown(screen.getAllByRole('combobox')[0])
    fireEvent.click(await screen.findByText('律师甲'))
    fireEvent.click(screen.getByRole('button', { name: /创建任务/ }))
    expect(await screen.findByText(/创建请求未收到服务器确认/)).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: /核对创建结果/ }))

    expect(await screen.findByText(/核对结果缺少有效任务信息/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /核对创建结果/ })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '用相同内容重试' })).not.toBeInTheDocument()
    // The pending key is untouched: no retry may fire with a rotated key.
    await waitFor(() => expect(postBodies).toHaveLength(1))
  })

  it('rejects create and found responses that omit case_id', async () => {
    const postBodies: string[] = []
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      const method = String(init?.method || 'GET').toUpperCase()
      if (path.endsWith('/cases/12/tasks') && method === 'GET') {
        return Promise.resolve(jsonResponse({ tasks: [] }))
      }
      if (path.endsWith('/cases/12/tasks') && method === 'POST') {
        postBodies.push(String(init?.body))
        return Promise.resolve(jsonResponse({ task: { id: 71 } }))
      }
      if (path.includes('/cases/12/tasks/creation-result')) {
        return Promise.resolve(jsonResponse({ state: 'found', task: { id: 71 } }))
      }
      return Promise.resolve(jsonResponse({}))
    })

    render(
      <CaseTaskPanel caseID={12} assigneeOptions={[{ id: 7, name: '律师甲' }]} currentUserId={7} />,
    )
    await screen.findByText('暂无案件任务。可在下方创建通用任务事项。')
    fireEvent.change(screen.getByPlaceholderText('任务标题（必填）'), {
      target: { value: '缺失case_id任务' },
    })
    fireEvent.mouseDown(screen.getAllByRole('combobox')[0])
    fireEvent.click(await screen.findByText('律师甲'))
    fireEvent.click(screen.getByRole('button', { name: /创建任务/ }))

    expect(await screen.findByText(/服务器返回内容异常/)).toBeInTheDocument()
    expect(message.success).not.toHaveBeenCalled()
    expect(screen.getByPlaceholderText('任务标题（必填）')).toHaveValue('缺失case_id任务')

    fireEvent.click(screen.getByRole('button', { name: /核对创建结果/ }))
    expect(await screen.findByText(/核对结果缺少有效任务信息/)).toBeInTheDocument()
    expect(message.success).not.toHaveBeenCalled()
    expect(screen.getByPlaceholderText('任务标题（必填）')).toHaveValue('缺失case_id任务')
    expect(postBodies).toHaveLength(1)
  })

  it('rejects a found task whose id is not a positive integer', async () => {
    const postBodies: string[] = []
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      const method = String(init?.method || 'GET').toUpperCase()
      if (path.endsWith('/cases/12/tasks') && method === 'GET') {
        return Promise.resolve(jsonResponse({ tasks: [] }))
      }
      if (path.endsWith('/cases/12/tasks') && method === 'POST') {
        postBodies.push(String(init?.body))
        throw new TypeError('Failed to fetch')
      }
      if (path.includes('/cases/12/tasks/creation-result')) {
        return Promise.resolve(jsonResponse({ state: 'found', task: { id: 0, case_id: 12 } }))
      }
      return Promise.resolve(jsonResponse({}))
    })

    render(
      <CaseTaskPanel caseID={12} assigneeOptions={[{ id: 7, name: '律师甲' }]} currentUserId={7} />,
    )
    await screen.findByText('暂无案件任务。可在下方创建通用任务事项。')
    fireEvent.change(screen.getByPlaceholderText('任务标题（必填）'), {
      target: { value: '零ID任务' },
    })
    fireEvent.mouseDown(screen.getAllByRole('combobox')[0])
    fireEvent.click(await screen.findByText('律师甲'))
    fireEvent.click(screen.getByRole('button', { name: /创建任务/ }))
    expect(await screen.findByText(/创建请求未收到服务器确认/)).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /核对创建结果/ }))
    expect(await screen.findByText(/核对结果缺少有效任务信息/)).toBeInTheDocument()
    expect(message.success).not.toHaveBeenCalled()
    expect(screen.getByPlaceholderText('任务标题（必填）')).toHaveValue('零ID任务')
    expect(postBodies).toHaveLength(1)
  })

  it('keeps the create state live when a stale case switch aborts an old create', async () => {
    let resolveOldCase: ((value: unknown) => void) | undefined
    const oldCasePromise = new Promise((resolve) => {
      resolveOldCase = resolve
    })
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      const method = String(init?.method || 'GET').toUpperCase()
      if (path.endsWith('/cases/12/tasks') && method === 'GET') {
        return Promise.resolve(jsonResponse({ tasks: [] }))
      }
      if (path.endsWith('/cases/12/tasks') && method === 'POST') {
        return oldCasePromise as Promise<Response>
      }
      if (path.endsWith('/cases/13/tasks') && method === 'GET') {
        return Promise.resolve(jsonResponse({ tasks: [] }))
      }
      return Promise.resolve(jsonResponse({}))
    })

    const { rerender } = render(
      <CaseTaskPanel caseID={12} assigneeOptions={[{ id: 7, name: '律师甲' }]} currentUserId={7} />,
    )
    await screen.findByText('暂无案件任务。可在下方创建通用任务事项。')
    fireEvent.change(screen.getByPlaceholderText('任务标题（必填）'), {
      target: { value: '跨案守卫任务' },
    })
    fireEvent.mouseDown(screen.getAllByRole('combobox')[0])
    fireEvent.click(await screen.findByText('律师甲'))
    fireEvent.click(screen.getByRole('button', { name: /创建任务/ }))
    rerender(
      <CaseTaskPanel caseID={13} assigneeOptions={[{ id: 7, name: '律师甲' }]} currentUserId={7} />,
    )
    await screen.findByText('暂无案件任务。可在下方创建通用任务事项。')

    // Fill and submit in the new case while the old-case POST still hangs.
    fireEvent.change(screen.getByPlaceholderText('任务标题（必填）'), {
      target: { value: '新案进行中任务' },
    })
    fireEvent.mouseDown(screen.getAllByRole('combobox')[0])
    fireEvent.click(await screen.findByText('律师甲'))
    fireEvent.click(screen.getByRole('button', { name: /创建任务/ }))
    expect(screen.getByRole('button', { name: /创建任务/ })).toBeDisabled()
    resolveOldCase!(
      jsonResponse({ task: { id: 71, case_id: 12, title: '跨案守卫任务' } }),
    )
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /创建任务/ })).toBeDisabled()
    })
    expect(screen.getByPlaceholderText('任务标题（必填）')).toHaveValue('新案进行中任务')
  })

  it('does not resubmit automatically after 409 stale but the manual retry uses the fresh version', async () => {
    const putBodies: string[] = []
    let latestVersion = 1
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      const method = String(init?.method || 'GET').toUpperCase()
      if (path.endsWith('/cases/12/tasks') && method === 'GET') {
        return Promise.resolve(
          jsonResponse({
            tasks: [
              {
                id: 5,
                case_id: 12,
                title: '手动重提版本任务',
                assignee_id: 7,
                created_by: 7,
                status: 'open',
                version: latestVersion,
                due_date: latestVersion === 1 ? '2026-10-08T00:00:00Z' : '2026-10-09T00:00:00Z',
                capabilities: { ...fullCapabilities, can_complete: false },
              },
            ],
          }),
        )
      }
      if (path.endsWith('/case-tasks/5') && method === 'PUT') {
        putBodies.push(String(init?.body))
        if (putBodies.length === 1) {
          latestVersion = 2
          return Promise.resolve({
            ok: false,
            status: 409,
            json: async () => ({
              success: false,
              error: { code: 'CASE_TASK_CHANGE_STALE', message: '任务已被他人修改' },
            }),
          } as Response)
        }
        return Promise.resolve(jsonResponse({ task: { id: 5 } }))
      }
      return Promise.resolve(jsonResponse({}))
    })

    render(
      <CaseTaskPanel caseID={12} assigneeOptions={[{ id: 7, name: '律师甲' }]} currentUserId={7} />,
    )
    await screen.findByText('手动重提版本任务')
    fireEvent.click(screen.getByRole('button', { name: /调整/ }))
    const reasonBox = await screen.findByPlaceholderText('调整原因（可选）')
    fireEvent.change(reasonBox, { target: { value: '期限顺延' } })
    fireEvent.click(screen.getByRole('button', { name: '提交变更' }))

    expect(await screen.findByText(/只有下一次手动提交才按最新任务内容处理/)).toBeInTheDocument()
    expect(await screen.findByText(/最新负责人：律师甲；最新期限：2026-10-09/)).toBeInTheDocument()
    expect(reasonBox).toHaveValue('期限顺延')
    expect(JSON.parse(putBodies[0]).version).toBe(1)
    expect(putBodies).toHaveLength(1)
    fireEvent.click(screen.getByRole('button', { name: '提交变更' }))
    await waitFor(() => expect(putBodies).toHaveLength(2))
    expect(JSON.parse(putBodies[1])).toEqual({
      version: 2,
      due_date: expect.any(String),
      reason: '期限顺延',
    })
    expect(JSON.parse(putBodies[1]).due_date).toBe(JSON.parse(putBodies[0]).due_date)
  })

  it('shows the latest server row after 409 stale without auto-upgrading or resubmitting', async () => {
    const putBodies: string[] = []
    let latestVersion = 1
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      const method = String(init?.method || 'GET').toUpperCase()
      if (path.endsWith('/cases/12/tasks') && method === 'GET') {
        return Promise.resolve(
          jsonResponse({
            tasks: [
              {
                id: 5,
                case_id: 12,
                title: '并发调整任务',
                assignee_id: 7,
                created_by: 7,
                status: 'open',
                version: latestVersion,
                due_date: latestVersion === 1 ? '2026-10-08T00:00:00Z' : '2026-10-09T00:00:00Z',
                capabilities: { ...fullCapabilities, can_complete: false },
              },
            ],
          }),
        )
      }
      if (path.endsWith('/case-tasks/5') && method === 'PUT') {
        putBodies.push(String(init?.body))
        if (putBodies.length === 1) {
          // Someone else advanced the task while the modal was open; the
          // recovery GET must now observe version 2.
          latestVersion = 2
          return Promise.resolve({
            ok: false,
            status: 409,
            json: async () => ({
              success: false,
              error: { code: 'CASE_TASK_CHANGE_STALE', message: '任务已被他人修改' },
            }),
          } as Response)
        }
        return Promise.resolve(jsonResponse({ task: { id: 5 } }))
      }
      return Promise.resolve(jsonResponse({}))
    })

    render(
      <CaseTaskPanel caseID={12} assigneeOptions={[{ id: 7, name: '律师甲' }]} currentUserId={7} />,
    )
    await screen.findByText('并发调整任务')
    fireEvent.click(screen.getByRole('button', { name: /调整/ }))
    const reasonBox = await screen.findByPlaceholderText('调整原因（可选）')
    fireEvent.change(reasonBox, { target: { value: '期限顺延' } })
    fireEvent.click(screen.getByRole('button', { name: '提交变更' }))

    expect(await screen.findByText(/任务已被他人修改，请对照最新内容后重试/)).toBeInTheDocument()
    expect(await screen.findByText(/最新负责人：律师甲；最新期限：2026-10-09/)).toBeInTheDocument()
    expect(reasonBox).toHaveValue('期限顺延')
    expect(JSON.parse(putBodies[0]).version).toBe(1)
    expect(putBodies).toHaveLength(1)
    // The modal now carries the fresh version, but nothing is auto-resubmitted.
    await waitFor(() => expect(putBodies).toHaveLength(1))

    // Explicit user retry submits with the fresh version.
    fireEvent.click(screen.getByRole('button', { name: '提交变更' }))
    await waitFor(() => expect(putBodies).toHaveLength(2))
    expect(JSON.parse(putBodies[1])).toEqual({
      version: 2,
      due_date: expect.any(String),
      reason: '期限顺延',
    })
    expect(JSON.parse(putBodies[1]).due_date).toBe(JSON.parse(putBodies[0]).due_date)
  })

  it('shows old and new due dates in the change history', async () => {
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const path = String(input)
      if (path.endsWith('/cases/12/tasks')) {
        return Promise.resolve(
          jsonResponse({
            tasks: [
              {
                id: 5,
                case_id: 12,
                title: '期限审计任务',
                assignee_id: 7,
                created_by: 7,
                status: 'open',
                version: 2,
                capabilities: { ...fullCapabilities, can_complete: false },
              },
            ],
          }),
        )
      }
      if (path.endsWith('/case-tasks/5/changes')) {
        return Promise.resolve(
          jsonResponse({
            changes: [
              {
                id: 1,
                created_at: '2026-10-02T09:00:00Z',
                change_type: 'adjust',
                actor_id: 7,
                old_due_date: '2026-10-05T00:00:00Z',
                new_due_date: '2026-10-09T00:00:00Z',
                reason: '庭审延期',
              },
              {
                id: 2,
                created_at: '2026-10-01T09:00:00Z',
                change_type: 'adjust',
                actor_id: 7,
                old_due_date: '2026-10-05T00:00:00Z',
                new_due_date: null,
              },
            ],
          }),
        )
      }
      return Promise.resolve(jsonResponse({}))
    })

    render(
      <CaseTaskPanel caseID={12} assigneeOptions={[{ id: 7, name: '律师甲' }]} currentUserId={7} />,
    )
    await screen.findByText('期限审计任务')
    fireEvent.click(screen.getByRole('button', { name: /变更记录/ }))
    expect(await screen.findByText(/期限：2026-10-05 → 2026-10-09/)).toBeInTheDocument()
    expect(screen.getByText(/期限：2026-10-05 → 已清除/)).toBeInTheDocument()
  })
})
