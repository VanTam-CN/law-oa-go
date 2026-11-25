import { act, renderHook } from '@testing-library/react'
import { useNarrowPilotSave } from '../useNarrowPilotSave'

// C04-C07/C11: per-request v2 records, legacy dedup, storage failure,
// cross-page discovery and actor isolation. The real two-page race stays in C07.
const recordBase = "np-save-v2:lawyer-1:intake-1"
const otherAccountBase = "np-save-v2:lawyer-2:intake-1"
const legacyScope = "np-save:lawyer-1:intake-1"
const requestA = "ee23cb63-6169-4682-937c-17618ff28344"
const requestB = "e0d0e0e0-e0e0-e0e0-e0e0-e0e0e0e0e0e0"
const revision = "44444444-4444-4444-a444-444444444444"
const projection = JSON.stringify({ form: { title: "虚构标题", description: "虚构摘要" } })

type Props = Parameters<typeof useNarrowPilotSave>[0]
function setup(actorID = 'lawyer-1') {
  const api = jest.fn()
  const initialProps: Props = {
    actorID, intakeID: 'intake-1', revision, description: '虚构摘要',
    projection, baseline: projection, apiRequest: api as Props['apiRequest'],
  }
  return { api, initialProps, ...renderHook(useNarrowPilotSave, { initialProps }) }
}
function remember(base: string, requestId: string) {
  localStorage.setItem(base + ':' + requestId, JSON.stringify({ request_id: requestId, action: 'update', intake_id: 'intake-1' }))
}
beforeEach(() => { localStorage.clear(); jest.restoreAllMocks() })

test('旧单槽与本页新键同 request 时去重，终态只清该请求', async () => {
  const { api, result } = setup()
  api.mockRejectedValueOnce(new Error('虚构断线'))
  await act(async () => { await expect(result.current.save()).rejects.toThrow('虚构断线') })
  const requestId = result.current.pending!.request_id
  localStorage.setItem(legacyScope, JSON.stringify({ request_id: requestId, action: 'update', intake_id: 'intake-1' }))
  api.mockResolvedValueOnce({ state: 'committed', id: 'intake-1', revision, request_id: requestId })
  await act(async () => { await result.current.save() })
  expect(localStorage.getItem(recordBase + ':' + requestId)).toBeNull()
  expect(localStorage.getItem(legacyScope)).toBeNull()
})

test('畸形旧单槽不阻断终态清理', async () => {
  remember(recordBase, requestA)
  localStorage.setItem(legacyScope, '{malformed')
  const { api, result } = setup()
  api.mockResolvedValueOnce({ state: 'abandoned', id: 'intake-1', request_id: requestA })
  await act(async () => { await result.current.abandon() })
  expect(localStorage.getItem(recordBase + ':' + requestA)).toBeNull()
  expect(localStorage.getItem(legacyScope)).toBe('{malformed')
})

test('恢复标识写入失败阻止保存请求', async () => {
  const { api, result } = setup()
  jest.spyOn(localStorage, 'setItem').mockImplementationOnce(() => { throw new Error('quota') })
  await act(async () => { await expect(result.current.save()).rejects.toThrow('恢复标识无法写入') })
  expect(api).not.toHaveBeenCalled()
  expect(result.current.pending).toBeNull()
})

test('同账号在途记录可发现、互不误清；不同账号只见自己的键空间', async () => {
  remember(recordBase, requestA)
  remember(recordBase, requestB)
  const { api, result } = setup()
  await act(async () => { await expect(result.current.save()).rejects.toThrow('原保存结果待核实') })
  expect(api).not.toHaveBeenCalled()
  api.mockResolvedValueOnce({ state: 'unknown', request_id: requestA })
  await act(async () => { await result.current.query() })
  expect(localStorage.getItem(recordBase + ':' + requestA)).not.toBeNull()
  expect(localStorage.getItem(recordBase + ':' + requestB)).not.toBeNull()
  api.mockResolvedValueOnce({ state: 'abandoned', id: 'intake-1', request_id: requestA })
  await act(async () => { await result.current.abandon() })
  expect(localStorage.getItem(recordBase + ':' + requestA)).toBeNull()
  expect(localStorage.getItem(recordBase + ':' + requestB)).not.toBeNull()
  const other = setup('lawyer-2')
  expect(other.result.current.pending).toBeNull()
  other.api.mockRejectedValueOnce(new Error('other failed'))
  await act(async () => { await expect(other.result.current.save()).rejects.toThrow('other failed') })
  const created = other.result.current.pending!.request_id
  expect(created).not.toBe(requestB)
  expect(localStorage.getItem(otherAccountBase + ':' + created)).not.toBeNull()
  expect(localStorage.getItem(recordBase + ':' + requestB)).not.toBeNull()
})

test('畸形同scope记录 mount 后 storageBlocked 并阻止严格保存', async () => {
  localStorage.setItem(recordBase + ':broken-key', '{"request_id":"ee23cb63-6169-4682-937c-17618ff28344","action":"update","intake_id":"intake-1","extra":123}')
  const { api, result } = setup()
  expect(result.current.storageBlocked).toBe(true)
  expect(result.current.pending).toBeNull()
  await act(async () => { await expect(result.current.save()).rejects.toThrow('严格保存已阻止') })
  expect(api).not.toHaveBeenCalled()
})

test('混合字段旧键不能被忽略，也不自动删除', async () => {
  const mixedKey = recordBase + ':ee23cb63-6169-4682-937c-17618ff28344'
  localStorage.setItem(mixedKey, JSON.stringify({ request_id: 'ee23cb63-6169-4682-937c-17618ff28344', action: 'update', intake_id: 'intake-1', memo: 'x' }))
  const { api, result } = setup()
  expect(result.current.storageBlocked).toBe(true)
  await act(async () => { await expect(result.current.save()).rejects.toThrow('严格保存已阻止') })
  expect(api).not.toHaveBeenCalled()
  expect(localStorage.getItem(mixedKey)).not.toBeNull()
})

test('挂载枚举异常 save 前置拦截时也标 blocked', async () => {
  localStorage.setItem(recordBase + ':x', '{bad')
  const { api, result } = setup()
  await act(async () => { await expect(result.current.save()).rejects.toThrow('恢复标识无法核对') })
  expect(api).not.toHaveBeenCalled()
  expect(result.current.storageBlocked).toBe(true)
})

test('同scope畸形记录跨代保留并保持 blocked', async () => {
  localStorage.setItem(recordBase + ':malformed', '{bad')
  const first = setup()
  expect(first.result.current.storageBlocked).toBe(true)
  await act(async () => { first.rerender({ ...first.initialProps, revision: '55555555-5555-5555-a555-555555555555' }) })
  expect(first.result.current.storageBlocked).toBe(true)
  expect(localStorage.getItem(recordBase + ':malformed')).toBe('{bad')
  expect(first.result.current.pending).toBeNull()
})

test.each([
  ['多字段', { request_id: requestA, action: 'update', intake_id: 'intake-1', extra: 1 }],
  ['action不符', { request_id: requestA, action: 'create', intake_id: 'intake-1' }],
  ['intake不符', { request_id: requestA, action: 'update', intake_id: 'intake-9' }],
])('clearResolved legacy %s 时保留旧键', async (_label, legacyValue) => {
  localStorage.setItem(legacyScope, JSON.stringify(legacyValue))
  remember(recordBase, requestA)
  const { api, result } = setup()
  api.mockResolvedValueOnce({ state: 'abandoned', id: 'intake-1', request_id: requestA })
  await act(async () => { await result.current.abandon() })
  expect(JSON.parse(localStorage.getItem(legacyScope)!)).toEqual(legacyValue)
  expect(localStorage.getItem(recordBase + ':' + requestA)).toBeNull()
})

test('legacy 三字段一致终态才清旧键', async () => {
  localStorage.setItem(legacyScope, JSON.stringify({ request_id: requestA, action: 'update', intake_id: 'intake-1' }))
  remember(recordBase, requestA)
  const { api, result } = setup()
  api.mockResolvedValueOnce({ state: 'abandoned', id: 'intake-1', request_id: requestA })
  await act(async () => { await result.current.abandon() })
  expect(localStorage.getItem(legacyScope)).toBeNull()
  expect(localStorage.getItem(recordBase + ':' + requestA)).toBeNull()
})

test('终态清理后重新枚举遇畸形：本请求已清且转 blocked 不误报未知', async () => {
  remember(recordBase, requestA)
  const { api, result } = setup()
  expect(result.current.pending?.request_id).toBe(requestA)
  const receipt = { state: 'abandoned', id: 'intake-1', request_id: requestA }
  api.mockImplementationOnce(async () => {
    localStorage.setItem(recordBase + ':newpage', '{bad')
    return receipt
  })
  await act(async () => { await result.current.abandon() })
  expect(localStorage.getItem(recordBase + ':' + requestA)).toBeNull()
  expect(result.current.storageBlocked).toBe(true)
  expect(result.current.pending).toBeNull()
  expect(result.current.notice).not.toContain('结果尚未核实')
})

test('A→B→A 迟到终态只清存储不动当前视图', async () => {
  const { api, result, rerender, initialProps } = setup()
  api.mockRejectedValueOnce(new Error('A断线'))
  await act(async () => { await expect(result.current.save()).rejects.toThrow('A断线') })
  const requestAId = result.current.pending!.request_id
  // Generation 2 (simulated by rerender to B and back to A) re-enumerates;
  // the still-pending generation-1 query must settle storage only.
  await act(async () => { rerender({ ...initialProps, actorID: 'lawyer-2' }) })
  await act(async () => { rerender(initialProps) })
  expect(result.current.pending?.request_id).toBe(requestAId)
  const receipt = { state: 'committed', id: 'intake-1', revision, request_id: requestAId }
  api.mockImplementationOnce(async () => receipt)
  await act(async () => { await result.current.query() })
  // The resolved record is gone from storage; the current view re-enumerates
  // to the same record only if storage still holds it, which it must not.
  expect(localStorage.getItem(recordBase + ':' + requestAId)).toBeNull()
  expect(result.current.pending).toBeNull()
  expect(result.current.storageBlocked).toBe(false)
})

test('跨账号迟到回包不把旧scope其他在途记录带进新页面', async () => {
  // Lawyer A already has two pending records; the page resolves the first,
  // then switches to lawyer B before the receipt arrives.
  remember(recordBase, requestA)
  remember(recordBase, requestB)
  const { api, result, rerender, initialProps } = setup()
  expect(result.current.pending?.request_id).toBe(requestA)
  let releaseReceipt: (value: unknown) => void = () => {}
  api.mockImplementationOnce(() => new Promise(resolve => { releaseReceipt = resolve }))
  let queryPromise: Promise<void> = Promise.resolve()
  act(() => { queryPromise = result.current.query() })
  await act(async () => { rerender({ ...initialProps, actorID: 'lawyer-2' }) })
  releaseReceipt({ state: 'abandoned', id: 'intake-1', request_id: requestA })
  await act(async () => { await queryPromise })
  // Storage is settled for request A; lawyer B's view must stay empty.
  expect(localStorage.getItem(recordBase + ':' + requestA)).toBeNull()
  expect(localStorage.getItem(recordBase + ':' + requestB)).not.toBeNull()
  expect(result.current.pending).toBeNull()
})

test('同scope新代允许迟到回包刷新本scope视图', async () => {
  remember(recordBase, requestA)
  remember(recordBase, requestB)
  const { api, result, rerender, initialProps } = setup()
  expect(result.current.pending?.request_id).toBe(requestA)
  let releaseReceipt: (value: unknown) => void = () => {}
  api.mockImplementationOnce(() => new Promise(resolve => { releaseReceipt = resolve }))
  let queryPromise: Promise<void> = Promise.resolve()
  act(() => { queryPromise = result.current.query() })
  await act(async () => { rerender({ ...initialProps, revision: '55555555-5555-5555-a555-555555555555' }) })
  releaseReceipt({ state: 'abandoned', id: 'intake-1', request_id: requestA })
  await act(async () => { await queryPromise })
  expect(localStorage.getItem(recordBase + ':' + requestA)).toBeNull()
  // Same scope, newer generation: the view refreshes to the remaining record.
  expect(result.current.pending?.request_id).toBe(requestB)
})
