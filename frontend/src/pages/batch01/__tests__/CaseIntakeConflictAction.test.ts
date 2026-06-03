import { getConflictCheckFallbackMessage } from '../Batch01Prototype'

describe('CaseIntakeWorkbench conflict action', () => {
  it('keeps a clear MVP fallback message for unavailable conflict checks', () => {
    expect(getConflictCheckFallbackMessage()).toBe(
      '试用版当前使用样例冲突复核流程，请在利益冲突工作台查看待复核事项。',
    )
  })
})
