import { act, renderHook } from '@testing-library/react'
import { useNarrowPilotSave } from '../useNarrowPilotSave'

type Props = Parameters<typeof useNarrowPilotSave>[0]
const scope = 'np-save:lawyer-1:intake-1'
const projection = JSON.stringify({ form: { title: '虚构标题', description: '虚构摘要' } })
function setup() {
  const api = jest.fn()
  const initialProps: Props = {
    actorID: 'lawyer-1', intakeID: 'intake-1', revision: 'revision-1', description: '虚构摘要',
    projection, baseline: projection, apiRequest: api as Props['apiRequest'],
  }
  return { api, initialProps, ...renderHook(useNarrowPilotSave, { initialProps }) }
}
beforeEach(() => localStorage.clear())

test.each([
  [408, undefined], [408, 'SOURCE_SELECTION_STALE'], [429, undefined], [499, undefined], [400, undefined], [400, 'INVALID_REQUEST'],
  [401, undefined], [403, 'OBJECT_ACCESS_DENIED'], [404, undefined],
  [401, 'ACCESS_SESSION_REQUIRED'], [409, 'IDEMPOTENCY_PAYLOAD_MISMATCH'],
  [409, 'SAVE_GUARD_BUSY'], [409, 'UNRECOGNIZED'], [503, 'SAVE_RESULT_UNKNOWN'],
  [undefined, undefined],
])('不明保存结果保留原键、冻结体且不把正文存浏览器：%s %s', async (status, code) => {
  const { api, result, rerender, initialProps } = setup()
  const failure = Object.assign(new Error('虚构失败'), { status, code })
  api.mockRejectedValue(failure)
  await act(async () => { await expect(result.current.save()).rejects.toBe(failure) })
  const ref = JSON.parse(localStorage.getItem(scope) || 'null')
  expect(ref).toEqual({ request_id: expect.any(String), action: 'update', intake_id: 'intake-1' })
  expect(Object.keys(ref).sort()).toEqual(['action', 'intake_id', 'request_id'])
  expect(result.current.notice).toContain('不能换键')
  const first = api.mock.calls[0][1]
  await act(async () => { await expect(result.current.save()).rejects.toBe(failure) })
  expect(api.mock.calls[1][1]).toEqual(first)
  rerender({ ...initialProps, description: '人工修改保留' })
  await act(async () => { await expect(result.current.save()).rejects.toThrow('原保存结果待核实') })
  expect(api).toHaveBeenCalledTimes(2)
  expect(JSON.parse(localStorage.getItem(scope)!).request_id).toBe(ref.request_id)
})

test.each(['INTAKE_DRAFT_REVISION_CONFLICT', 'SOURCE_SELECTION_STALE', 'SAVE_REQUEST_ABANDONED'])(
  '明确终态 409 %s 才清恢复键', async (code) => {
    const { api, result } = setup()
    api.mockRejectedValue(Object.assign(new Error('明确拒绝'), { status: 409, code }))
    await act(async () => { await expect(result.current.save()).rejects.toThrow('明确拒绝') })
    expect(localStorage.getItem(scope)).toBeNull()
    expect(result.current.pending).toBeNull()
  },
)

test('重挂载无冻结正文时不能新键保存，unknown 查询不放弃，明确放弃后才解锁', async () => {
  localStorage.setItem(scope, JSON.stringify({ request_id: 'original-key', action: 'update', intake_id: 'intake-1' }))
  const { api, result } = setup()
  await act(async () => { await expect(result.current.save()).rejects.toThrow('原保存结果待核实') })
  expect(api).not.toHaveBeenCalled()
  api.mockResolvedValueOnce({ request_id: 'original-key', state: 'unknown' })
  await act(async () => { await result.current.query() })
  expect(api.mock.calls[0]).toEqual(['/case-intakes/write-results/original-key', {}])
  expect(localStorage.getItem(scope)).not.toBeNull()
  api.mockResolvedValueOnce({ request_id: 'original-key', state: 'abandoned' })
  await act(async () => { await result.current.abandon() })
  expect(api.mock.calls[1][0]).toBe('/case-intakes/write-results/original-key/abandon')
  expect(api.mock.calls[1][1].method).toBe('POST')
  expect(localStorage.getItem(scope)).toBeNull()
})

test.each(['committed', 'no_change'])('成功回执 %s 清理恢复键', async (state) => {
  const { api, result } = setup()
  api.mockResolvedValueOnce({ state, id: 'intake-1', revision: 'revision-2' })
  await act(async () => { await result.current.save() })
  expect(localStorage.getItem(scope)).toBeNull()
  expect(result.current.pending).toBeNull()
})

test('带来源的未决请求只缓存恢复标识，正文和选择仅保留在内存', async () => {
  const { api, result } = setup()
  const selection = {
    context_id: 'context-1', expected_context_version: 2,
    fields: { description: { candidate_id: 'candidate-1', requested_mode: 'candidate_sources' as const, reviewed: true } },
  }
  act(() => result.current.apply('虚构摘要', selection))
  api.mockRejectedValue(new Error('虚构断线'))
  await act(async () => { await expect(result.current.save()).rejects.toThrow('虚构断线') })
  expect(JSON.parse(api.mock.calls[0][1].body).source_selection).toEqual(selection)
  const ref = JSON.parse(localStorage.getItem(scope)!)
  expect(Object.keys(ref).sort()).toEqual(['action', 'intake_id', 'request_id'])
  expect(localStorage.getItem(scope)).not.toContain('虚构摘要')
  expect(localStorage.getItem(scope)).not.toContain('candidate-1')
})
