import React from 'react'
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { NarrowPilotAssistant, type NarrowPilotApplySelection } from '../NarrowPilotAssistant'

type ApiFn = React.ComponentProps<typeof NarrowPilotAssistant>['apiRequest']

const candidate = (id: string, description: string) => ({
  id,
  description,
  sources: [{ id: `${id}-s1`, kind: 'user_input', text: '第一轮输入原文' }],
})
const contextView = (candidateView: ReturnType<typeof candidate> | null, version: number) => ({
  id: 'ctx-1',
  context_version: version,
  expires_at: new Date(Date.now() + 30 * 60_000).toISOString(),
  candidate: candidateView,
})

function SwitchHarness({
  apiMock,
  onApplySpy,
}: {
  apiMock: ApiFn
  onApplySpy?: jest.Mock
}) {
  const [intakeId, setIntakeId] = React.useState('I1')
  const [description, setDescription] = React.useState('')
  const [selection, setSelection] = React.useState<NarrowPilotApplySelection | null>(null)
  return (
    <>
      <button onClick={() => { setIntakeId('I2'); setDescription('I2保存版本摘要') }}>切换草稿</button>
      <button onClick={() => { setIntakeId('I1'); setDescription('I1新编辑摘要') }}>返回原草稿</button>
      <textarea
        aria-label='案情摘要'
        value={description}
        onChange={(event) => setDescription(event.target.value)}
      />
      <span data-testid='intake'>{intakeId}</span>
      <NarrowPilotAssistant
        intakeId={intakeId}
        revision='rev-1'
        description={description}
        apiRequest={apiMock}
        onApply={(text, value) => {
          setDescription(text)
          setSelection(value)
          onApplySpy?.(text, value)
        }}
        onInvalidate={() => setSelection(null)}
      />
      <span data-testid='selection'>{selection ? selection.fields.description.candidate_id : 'none'}</span>
    </>
  )
}

function scopeApiMock(gatedGet: Promise<unknown>): ApiFn {
  const mock = jest.fn((path: string, init?: RequestInit) => {
    const key = `${init?.method || 'GET'} ${path}`
    if (key === 'GET /case-intake-assist/capabilities') return Promise.resolve({ enabled: true })
    if (key === 'GET /case-intakes/I1/source-snapshots/rev-1') {
      return Promise.resolve({ revision: 'rev-1', source_revision: null, source: null })
    }
    if (key === 'GET /case-intakes/I2/source-snapshots/rev-1') {
      return Promise.resolve({
        revision: 'rev-1', source_revision: 'source-2',
        source: { description: 'I2保存版本摘要', sources: [{ id: 's2', kind: 'user_input', text: 'I2已保存来源' }] },
      })
    }
    if (key === 'POST /case-intake-assist/contexts') {
      return Promise.resolve(contextView(null, 1))
    }
    if (key === 'POST /case-intake-assist/contexts/ctx-1/turns') {
      return Promise.resolve(contextView(candidate('cand-1', '第一轮候选摘要'), 2))
    }
    if (key === 'GET /case-intake-assist/contexts/ctx-1') {
      return gatedGet
    }
    throw new Error(`unexpected api call: ${key}`)
  })
  return mock as unknown as ApiFn
}

test.each([false, true])('同挂载切换intakeId：旧 GET 成功或失败均不影响新草稿，拒绝=%s', async (denied) => {
  let release!: (value: unknown) => void
  let reject!: (error: unknown) => void
  const gatedGet = new Promise((resolve, fail) => {
    release = resolve
    reject = fail
  })
  const applySpy = jest.fn()
  render(<SwitchHarness apiMock={scopeApiMock(gatedGet)} onApplySpy={applySpy} />)

  await screen.findByLabelText('摘要对话输入')
  fireEvent.change(screen.getByLabelText('摘要对话输入'), { target: { value: '第一轮输入原文' } })
  fireEvent.click(screen.getByRole('button', { name: '开始摘要整理' }))
  await screen.findByText('第一轮候选摘要')
  fireEvent.click(screen.getByLabelText('已核对来源'))
  fireEvent.click(screen.getByRole('button', { name: '应用摘要' }))

  // GET 仍在等待；同一挂载不换 key，只把 intakeId 切到 I2。
  fireEvent.click(screen.getByRole('button', { name: '切换草稿' }))
  expect(screen.getByTestId('intake')).toHaveTextContent('I2')

  // 旧 GET 迟到返回：不得调用 onApply，也不得覆盖新草稿输入或渲染旧候选。
  await act(async () => {
    if (denied) reject(Object.assign(new Error('旧请求已失去权限'), { code: 'OBJECT_ACCESS_DENIED' }))
    else release(contextView(candidate('cand-1', '第一轮候选摘要'), 2))
  })
  await waitFor(() => expect(screen.getByText('I2已保存来源')).toBeInTheDocument())
  expect(applySpy).not.toHaveBeenCalled()
  expect(screen.getByTestId('selection')).toHaveTextContent('none')
  expect(screen.getByLabelText('案情摘要')).toHaveValue('I2保存版本摘要')
  expect(screen.queryByText('第一轮候选摘要')).not.toBeInTheDocument()
})

test('切换草稿后迟到的生成响应不能显示旧候选，新草稿不被旧 busy 锁住', async () => {
  let release!: (value: unknown) => void
  const gatedTurn = new Promise(resolve => { release = resolve })
  const fallback = scopeApiMock(Promise.resolve(contextView(null, 1)))
  const apiMock = jest.fn((path: string, init?: RequestInit) => path.endsWith('/turns')
    ? gatedTurn : fallback(path, init))
  render(<SwitchHarness apiMock={apiMock as ApiFn} />)
  await screen.findByLabelText('摘要对话输入')
  fireEvent.change(screen.getByLabelText('摘要对话输入'), { target: { value: '第一轮输入原文' } })
  fireEvent.click(screen.getByRole('button', { name: '开始摘要整理' }))
  await waitFor(() => expect(apiMock).toHaveBeenCalledWith('/case-intake-assist/contexts/ctx-1/turns', expect.any(Object)))
  fireEvent.click(screen.getByRole('button', { name: '切换草稿' }))
  await screen.findByText('I2已保存来源')
  await act(async () => { release(contextView(candidate('old', '迟到的旧草稿候选'), 2)) })
  expect(screen.queryByText('迟到的旧草稿候选')).not.toBeInTheDocument()
  expect(screen.getByLabelText('摘要对话输入')).toBeEnabled()
  fireEvent.change(screen.getByLabelText('摘要对话输入'), { target: { value: '新草稿继续整理' } })
  expect(screen.getByRole('button', { name: '开始摘要整理' })).toBeEnabled()
  expect(screen.getByLabelText('案情摘要')).toHaveValue('I2保存版本摘要')
})

test('切换后又返回同一草稿，旧 GET 仍失效且不能调用应用回调', async () => {
  let release!: (value: unknown) => void
  const gatedGet = new Promise(resolve => { release = resolve })
  const applySpy = jest.fn()
  render(<SwitchHarness apiMock={scopeApiMock(gatedGet)} onApplySpy={applySpy} />)
  await screen.findByLabelText('摘要对话输入')
  fireEvent.change(screen.getByLabelText('摘要对话输入'), { target: { value: '第一轮输入原文' } })
  fireEvent.click(screen.getByRole('button', { name: '开始摘要整理' }))
  await screen.findByText('第一轮候选摘要')
  fireEvent.click(screen.getByLabelText('已核对来源'))
  fireEvent.click(screen.getByRole('button', { name: '应用摘要' }))
  fireEvent.click(screen.getByRole('button', { name: '切换草稿' }))
  await screen.findByText('I2已保存来源')
  fireEvent.click(screen.getByRole('button', { name: '返回原草稿' }))
  await act(async () => { release(contextView(candidate('cand-1', '第一轮候选摘要'), 2)) })
  expect(applySpy).not.toHaveBeenCalled()
  expect(screen.queryByText('第一轮候选摘要')).not.toBeInTheDocument()
  expect(screen.getByLabelText('案情摘要')).toHaveValue('I1新编辑摘要')
})
