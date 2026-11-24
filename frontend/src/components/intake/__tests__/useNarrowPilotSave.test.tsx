import { act, renderHook } from '@testing-library/react'
import { useNarrowPilotSave } from '../useNarrowPilotSave'

type Props = Parameters<typeof useNarrowPilotSave>[0]
// M3 ADAPT (cache-policy 6.2): per-request v2 records replace the legacy slot.
const recordBase = "np-save-v2:lawyer-1:intake-1"
const legacyScope = "np-save:lawyer-1:intake-1"
const recordKeyOf = (requestId: string) => `${recordBase}:${requestId}`
const projection = JSON.stringify({ form: { title: "虚构标题", description: "虚构摘要" } })
function setup() {
  const api = jest.fn()
  const initialProps: Props = {
    actorID: 'lawyer-1', intakeID: 'intake-1', revision: 'revision-1', description: '虚构摘要',
    projection, baseline: projection, apiRequest: api as Props['apiRequest'],
  }
  return { api, initialProps, ...renderHook(useNarrowPilotSave, { initialProps }) }
}
function idempotencyKeyOf(call: unknown[]): string {
  return ((call[1] as RequestInit).headers as Record<string, string>)['Idempotency-Key']
}
function storedRecord(requestId: string): Record<string, unknown> | null {
  const raw = localStorage.getItem(recordKeyOf(requestId))
  return raw ? JSON.parse(raw) : null
}
beforeEach(() => localStorage.clear())

test.each([
  [408, undefined], [408, "SOURCE_SELECTION_STALE"], [429, undefined], [499, undefined], [400, undefined], [400, "INVALID_REQUEST"],
  [401, undefined], [403, "OBJECT_ACCESS_DENIED"], [404, undefined],
  [401, "ACCESS_SESSION_REQUIRED"], [409, "IDEMPOTENCY_PAYLOAD_MISMATCH"],
  [409, "SAVE_GUARD_BUSY"], [409, "UNRECOGNIZED"], [503, "SAVE_RESULT_UNKNOWN"],
  [undefined, undefined],
])('不明保存结果保留原键、冻结体且不把正文存浏览器：%s %s', async (status, code) => {
  const { api, result, rerender, initialProps } = setup()
  const failure = Object.assign(new Error('虚构失败'), { status, code })
  api.mockRejectedValue(failure)
  await act(async () => { await expect(result.current.save()).rejects.toBe(failure) })
  const requestId = result.current.pending!.request_id
  const ref = storedRecord(requestId)
  expect(ref).toEqual({ request_id: expect.any(String), action: 'update', intake_id: 'intake-1' })
  // v2 key format: np-save-v2:<actor>:<intake>:<request>
  expect(recordKeyOf(requestId)).toMatch(/^np-save-v2:lawyer-1:intake-1:[0-9a-f-]{36}$/)
  expect(Object.keys(ref!).sort()).toEqual(['action', 'intake_id', 'request_id'])
  expect(result.current.notice).toContain('不能换键')
  const first = api.mock.calls[0][1]
  await act(async () => { await expect(result.current.save()).rejects.toBe(failure) })
  expect(api.mock.calls[1][1]).toEqual(first)
  rerender({ ...initialProps, description: '人工修改保留' })
  await act(async () => { await expect(result.current.save()).rejects.toThrow('原保存结果待核实') })
  expect(api).toHaveBeenCalledTimes(2)
  expect(storedRecord(requestId)!.request_id).toBe(ref!.request_id)
})

test.each(['INTAKE_DRAFT_REVISION_CONFLICT', 'SOURCE_SELECTION_STALE', 'SAVE_REQUEST_ABANDONED'])(
  '明确终态 409 %s 才清恢复键', async (code) => {
    const { api, result } = setup()
    api.mockRejectedValue(Object.assign(new Error('明确拒绝'), { status: 409, code }))
    await act(async () => { await expect(result.current.save()).rejects.toThrow('明确拒绝') })
    const key = idempotencyKeyOf(api.mock.calls[0])
    expect(localStorage.getItem(recordKeyOf(key))).toBeNull()
    expect(result.current.pending).toBeNull()
  },
)

test('重挂载无冻结正文时不能新键保存，unknown 查询不放弃，明确放弃后才解锁', async () => {
  localStorage.setItem(legacyScope, JSON.stringify({ request_id: "ee23cb63-6169-4682-937c-17618ff28344", action: "update", intake_id: "intake-1" }))
  const { api, result } = setup()
  await act(async () => { await expect(result.current.save()).rejects.toThrow('原保存结果待核实') })
  expect(api).not.toHaveBeenCalled()
  api.mockResolvedValueOnce({ request_id: "ee23cb63-6169-4682-937c-17618ff28344", state: "unknown" })
  await act(async () => { await result.current.query() })
  expect(api.mock.calls[0]).toEqual(['/case-intakes/write-results/ee23cb63-6169-4682-937c-17618ff28344', {}])
  expect(localStorage.getItem(legacyScope)).not.toBeNull()
  api.mockResolvedValueOnce({ request_id: "ee23cb63-6169-4682-937c-17618ff28344", state: "abandoned", id: "intake-1" })
  await act(async () => { await result.current.abandon() })
  expect(api.mock.calls[1][0]).toBe('/case-intakes/write-results/ee23cb63-6169-4682-937c-17618ff28344/abandon')
  expect(api.mock.calls[1][1].method).toBe('POST')
  expect(localStorage.getItem(legacyScope)).toBeNull()
})

test.each(['committed', 'no_change'])('成功回执 %s 清本请求记录且保留其他未决请求', async (state) => {
  const { api, result } = setup()
  // 断线先留下本页 pending，再预置其他页记录，验证终态清理只清本请求。
  api.mockRejectedValueOnce(new Error('虚构断线'))
  await act(async () => { await expect(result.current.save()).rejects.toThrow('虚构断线') })
  localStorage.setItem(recordKeyOf("e0d3e52b-079d-40dd-a5bc-d6c10b2e66ee"), JSON.stringify({ request_id: "e0d3e52b-079d-40dd-a5bc-d6c10b2e66ee", action: "update", intake_id: "intake-1" }))
  api.mockImplementationOnce(async (_path: string, init?: RequestInit) => ({
    state,
    id: "intake-1",
    revision: "44444444-4444-4444-a444-444444444444",
    request_id: ((init?.headers ?? {}) as Record<string, string>)['Idempotency-Key'],
  }))
  await act(async () => { await result.current.save() })
  expect(api).toHaveBeenCalledTimes(2)
  const key = idempotencyKeyOf(api.mock.calls[1])
  expect(idempotencyKeyOf(api.mock.calls[0])).toBe(key)
  expect(localStorage.getItem(recordKeyOf(key))).toBeNull()
  // Other pending records must survive a different request resolution.
  expect(JSON.parse(localStorage.getItem(recordKeyOf("e0d3e52b-079d-40dd-a5bc-d6c10b2e66ee"))!).request_id).toBe("e0d3e52b-079d-40dd-a5bc-d6c10b2e66ee")
  expect(result.current.pending?.request_id).toBe("e0d3e52b-079d-40dd-a5bc-d6c10b2e66ee")
})

test.each([
  ['缺 request_id', () => ({ state: 'committed', id: 'intake-1', revision: '44444444-4444-4444-a444-444444444444' })],
  ['request_id 不符', () => ({ state: 'committed', request_id: 'wrong-key', id: 'intake-1', revision: '44444444-4444-4444-a444-444444444444' })],
  ['id 对象不符', (key: string) => ({ state: 'committed', request_id: key, id: 'intake-9', revision: '44444444-4444-4444-a444-444444444444' })],
  ['缺 revision', (key: string) => ({ state: 'committed', request_id: key, id: 'intake-1' })],
  ['非终态', (key: string) => ({ state: 'unknown', request_id: key, id: 'intake-1', revision: '44444444-4444-4444-a444-444444444444' })],
  ['action 不符', (key: string) => ({ state: 'committed', request_id: key, id: 'intake-1', revision: '44444444-4444-4444-a444-444444444444', action: 'delete' })],
])('错回执 %s 保留记录且不能换键重发', async (_label, buildReceipt) => {
  const { api, result } = setup()
  api.mockRejectedValue(new Error("稍后再试"))
  api.mockImplementationOnce(async (_path: string, init?: RequestInit) =>
    (buildReceipt as (key: string) => Record<string, unknown>)(((init?.headers ?? {}) as Record<string, string>)["Idempotency-Key"]))
  await act(async () => { await expect(result.current.save()).rejects.toThrow("保存回执无法确认") })
  const key = idempotencyKeyOf(api.mock.calls[0])
  expect(storedRecord(key)!.request_id).toBe(key)
  // Retry reuses the frozen key; a new key must not be unlocked.
  await act(async () => { await expect(result.current.save()).rejects.toThrow("稍后再试") })
  expect(idempotencyKeyOf(api.mock.calls[1])).toBe(key)
  expect(storedRecord(key)!.request_id).toBe(key)
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
  const requestId = result.current.pending!.request_id
  const ref = storedRecord(requestId)
  expect(Object.keys(ref!).sort()).toEqual(['action', 'intake_id', 'request_id'])
  expect(localStorage.getItem(recordKeyOf(requestId))).not.toContain("虚构摘要")
  expect(localStorage.getItem(recordKeyOf(requestId))).not.toContain("candidate-1")
})
