import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
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

function Harness({
  apiMock,
  onApplySpy,
  onInvalidateSpy,
}: {
  apiMock: ApiFn
  onApplySpy?: jest.Mock
  onInvalidateSpy?: jest.Mock
}) {
  const [description, setDescription] = React.useState('第一轮输入原文')
  const [selection, setSelection] = React.useState<NarrowPilotApplySelection | null>(null)
  return (
    <>
      <textarea
        aria-label='案情摘要'
        value={description}
        onChange={(event) => setDescription(event.target.value)}
      />
      <NarrowPilotAssistant
        intakeId='I1'
        revision='rev-1'
        description={description}
        apiRequest={apiMock}
        onApply={(text, value) => {
          setDescription(text)
          setSelection(value)
          onApplySpy?.(text, value)
        }}
        onInvalidate={() => {
          setSelection(null)
          onInvalidateSpy?.()
        }}
      />
      <span data-testid='selection'>{selection ? selection.fields.description.candidate_id : 'none'}</span>
    </>
  )
}

function baseApiMock(overrides: Record<string, (init?: RequestInit) => unknown> = {}): ApiFn {
  const mock = jest.fn((path: string, init?: RequestInit) => {
    const key = `${init?.method || 'GET'} ${path}`
    if (key === 'GET /case-intake-assist/capabilities') return Promise.resolve({ enabled: true })
    if (key === 'GET /case-intakes/I1/source-snapshots/rev-1') {
      return Promise.resolve({ revision: 'rev-1', source_revision: null, source: null })
    }
    const override = overrides[key]
    if (override) return override(init)
    throw new Error(`unexpected api call: ${key}`)
  })
  return mock as unknown as ApiFn
}

test('应用等待GET期间人工改摘要：恢复后进入对比确认，不得用旧闭包覆盖', async () => {
  let release!: (value: unknown) => void
  const gated = new Promise((resolve) => {
    release = resolve
  })
  const applySpy = jest.fn()
  const apiMock = baseApiMock({
    'POST /case-intake-assist/contexts': () => Promise.resolve(contextView(null, 1)),
    'POST /case-intake-assist/contexts/ctx-1/turns': () =>
      Promise.resolve(contextView(candidate('cand-1', '第一轮候选摘要'), 2)),
    'GET /case-intake-assist/contexts/ctx-1': () => gated,
  })
  render(<Harness apiMock={apiMock} onApplySpy={applySpy} />)

  await screen.findByLabelText('摘要对话输入')
  fireEvent.change(screen.getByLabelText('摘要对话输入'), { target: { value: '第一轮输入原文' } })
  fireEvent.click(screen.getByRole('button', { name: '开始摘要整理' }))
  await screen.findByText('第一轮候选摘要')
  fireEvent.click(screen.getByLabelText('已核对来源'))
  fireEvent.click(screen.getByRole('button', { name: '应用摘要' }))

  // GET 重验尚未返回；此刻人工修改摘要（含清空场景同路径）。
  fireEvent.change(screen.getByLabelText('案情摘要'), { target: { value: '人工改动' } })
  release(contextView(candidate('cand-1', '第一轮候选摘要'), 2))

  // 恢复后不得静默覆盖：必须出现人工对比确认界面。
  await screen.findByRole('alertdialog')
  expect(screen.getByLabelText('案情摘要')).toHaveValue('人工改动')
  expect(applySpy).not.toHaveBeenCalled()
})

test('开始新一轮立即清除旧来源选择并保留表单文字；原句原样提交', async () => {
  const applySpy = jest.fn()
  const invalidateSpy = jest.fn()
  const turnBodies: Array<Record<string, unknown>> = []
  const apiMock = baseApiMock({
    'POST /case-intake-assist/contexts': () => Promise.resolve(contextView(null, 1)),
    'POST /case-intake-assist/contexts/ctx-1/turns': (init?: RequestInit) => {
      const body = JSON.parse(String(init?.body)) as Record<string, unknown>
      const index = turnBodies.push(body) - 1
      return Promise.resolve(
        index === 0
          ? contextView(candidate('cand-1', '第一轮候选摘要'), 2)
          : contextView(candidate('cand-2', '第二轮候选摘要'), 3),
      )
    },
    'GET /case-intake-assist/contexts/ctx-1': () =>
      Promise.resolve(contextView(candidate('cand-1', '第一轮候选摘要'), 2)),
  })
  render(<Harness apiMock={apiMock} onApplySpy={applySpy} onInvalidateSpy={invalidateSpy} />)

  await screen.findByLabelText('摘要对话输入')
  fireEvent.change(screen.getByLabelText('摘要对话输入'), { target: { value: '第一轮输入原文' } })
  fireEvent.click(screen.getByRole('button', { name: '开始摘要整理' }))
  await screen.findByText('第一轮候选摘要')
  fireEvent.click(screen.getByLabelText('已核对来源'))
  fireEvent.click(screen.getByRole('button', { name: '应用摘要' }))
  await waitFor(() => expect(screen.getByTestId('selection')).toHaveTextContent('cand-1'))
  expect(screen.getByTestId('selection')).toHaveTextContent('cand-1')
  expect(screen.getByLabelText('案情摘要')).toHaveValue('第一轮候选摘要')

  fireEvent.change(screen.getByLabelText('摘要对话输入'), {
    target: { value: '  原句含首尾空格  ' },
  })
  fireEvent.click(screen.getByRole('button', { name: '生成摘要候选' }))

  // 新一轮发出即清旧 selection；表单文字保持不动。
  expect(screen.getByTestId('selection')).toHaveTextContent('none')
  expect(screen.getByLabelText('案情摘要')).toHaveValue('第一轮候选摘要')

  await screen.findByText('第二轮候选摘要')
  expect(applySpy).toHaveBeenCalledTimes(1)
  // trim 只判空；提交正文保留原句首尾空格。
  expect(turnBodies[1]?.text).toBe('  原句含首尾空格  ')
  expect(screen.getByLabelText('案情摘要')).toHaveValue('第一轮候选摘要')
})

test('来源权限被收回后清除已显示的来源，不再应用候选', async () => {
  const applySpy=jest.fn()
  const base=baseApiMock({
    'POST /case-intake-assist/contexts':()=>Promise.resolve(contextView(null,1)),
    'POST /case-intake-assist/contexts/ctx-1/turns':()=>Promise.resolve(contextView(candidate('cand-1','第一轮候选摘要'),2)),
    'GET /case-intake-assist/contexts/ctx-1':()=>Promise.reject(Object.assign(new Error('denied'),{code:'OBJECT_ACCESS_DENIED'})),
  })
  const api=((path:string,init?:RequestInit)=>path.includes('/source-snapshots/')
    ? Promise.resolve({revision:'rev-1',source_revision:'source-1',source:{description:'保存版本摘要',sources:[{id:'s',kind:'user_input',text:'保存版本敏感原句'}]}})
    : base(path,init)) as ApiFn
  render(<Harness apiMock={api} onApplySpy={applySpy}/>)
  await screen.findByText('保存版本敏感原句')
  fireEvent.change(screen.getByLabelText('摘要对话输入'),{target:{value:'第一轮输入原文'}})
  fireEvent.click(screen.getByRole('button',{name:'开始摘要整理'}))
  await screen.findByText('第一轮候选摘要')
  fireEvent.click(screen.getByLabelText('已核对来源'))
  fireEvent.click(screen.getByRole('button',{name:'应用摘要'}))
  await waitFor(()=>expect(screen.queryByText('保存版本敏感原句')).not.toBeInTheDocument())
  expect(screen.queryByText('第一轮候选摘要')).not.toBeInTheDocument()
  expect(applySpy).not.toHaveBeenCalled()
})
