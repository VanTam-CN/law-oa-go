import { act, renderHook } from '@testing-library/react'
import { useNarrowPilotSave } from '../useNarrowPilotSave'

const intake = 'aa812201-a10a-4c95-bd2b-449a2c28688e'
const original = '39753ad6-d0fd-465b-b86c-bd51c51a985e'
const another = 'be649c0c-bf2e-40c3-82e7-3bf9e7700ded'
const revision = 'fff8b25e-0a42-488d-9afb-02031dd937f2'
const text = '人工修订应保留'
const projection = JSON.stringify({ form: { description: text } })
const base = 'np-save-v2:71:' + intake + ':'
function remember(key: string) {
  localStorage.setItem(base + key, JSON.stringify({ action: 'update', intake_id: intake, request_id: key }))
}
beforeEach(() => localStorage.clear())

describe.each(['query', 'abandon'] as const)('%s 恢复回执', operation => {
  test.each([
    ['committed 缺少对象', { state: 'committed', revision }],
    ['no_change 缺少版本', { state: 'no_change', id: intake }],
    ['committed 非法版本', { state: 'committed', id: intake, revision: 'bad' }],
    ['committed 其他对象的旧版本', { state: 'committed', id: intake, revision: 'legacy-' + another }],
    ['abandoned 其他请求', { state: 'abandoned', id: intake, request_id: another }],
    ['abandoned 缺少对象', { state: 'abandoned' }],
    ['abandoned 错误动作', { state: 'abandoned', id: intake, action: 'create' }],
  ])('%s 保留恢复记录与人工输入', async (_label, fields) => {
    remember(original)
    const apiRequest = jest.fn().mockResolvedValue({ request_id: original, ...(fields as object) })
    const props = { actorID: '71', intakeID: intake, revision, description: text, baseline: projection, projection, apiRequest }
    const { result } = renderHook(useNarrowPilotSave, { initialProps: props })
    remember(another)
    await act(async () => { await result.current[operation]() })
    expect(apiRequest).toHaveBeenCalledTimes(1)
    expect(result.current.pending?.request_id).toBe(original)
    expect(localStorage.getItem(base + original)).not.toBeNull()
    expect(localStorage.getItem(base + another)).not.toBeNull()
  })
  test.each([
    ['committed', { revision }],
    ['no_change', { revision: 'legacy-' + intake, action: 'update' }],
    ['abandoned', {}],
  ])('%s 身份完整仅释放原请求', async (state, fields) => {
    remember(original)
    const apiRequest = jest.fn().mockResolvedValue({ id: intake, request_id: original, state, ...(fields as object) })
    const props = { actorID: '71', intakeID: intake, revision, description: text, baseline: projection, projection, apiRequest }
    const { result } = renderHook(useNarrowPilotSave, { initialProps: props })
    remember(another)
    await act(async () => { await result.current[operation]() })
    expect(apiRequest).toHaveBeenCalledTimes(1)
    expect(localStorage.getItem(base + original)).toBeNull()
    expect(localStorage.getItem(base + another)).not.toBeNull()
  })
})
