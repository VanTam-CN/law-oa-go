import { narrowPilotIntakeEligibility, projectionWithDescription } from '../narrowPilotI13'

const base = {
  isAssistant: false,
  isSupplementMode: false,
  intakeID: 'intake-1',
  revision: 'legacy-intake-1',
  currentUserID: '7',
  role: 'lawyer',
  userIsActive: true,
  ownerID: 7,
  intakeStatus: 'draft',
  projection: JSON.stringify({ form: { title: 'A', description: '新摘要', opponentIdentityNumber: '911' }, relatedParties: [], caseTags: [] }),
  baseline: JSON.stringify({ form: { title: 'A', description: '旧摘要', opponentIdentityNumber: '911' }, relatedParties: [], caseTags: [] }),
  hasSourceHistory: true,
  sourceHistoryKnown: true,
}

test('在职律师 + 本人 draft + 仅摘要变化 + 有来源历史 → strict', () => {
  expect(narrowPilotIntakeEligibility(base)).toEqual({ mode: 'strict' })
})

test.each([
  ['title 混改', (b: typeof base) => ({ ...b, projection: JSON.stringify({ ...JSON.parse(b.projection), form: { ...JSON.parse(b.projection).form, title: 'B' } }) })],
  ['对方身份号混改（仅内存投影参与）', (b: typeof base) => ({ ...b, projection: JSON.stringify({ ...JSON.parse(b.projection), form: { ...JSON.parse(b.projection).form, opponentIdentityNumber: '912' } }) })],
  ['新增相关当事人', (b: typeof base) => ({ ...b, projection: JSON.stringify({ ...JSON.parse(b.projection), relatedParties: [{ name: '新主体' }] }) })],
  ['标签变化', (b: typeof base) => ({ ...b, projection: JSON.stringify({ ...JSON.parse(b.projection), caseTags: ['新标签'] }) })],
])('有来源历史时 %s 不能走 strict', (_label, build) => {
  expect(narrowPilotIntakeEligibility(build(base)).mode).toBe('blocked')
})

test.each([
  ['非律师角色', (b: typeof base) => ({ ...b, role: 'assistant' })],
  ['离职律师', (b: typeof base) => ({ ...b, userIsActive: false })],
  ['非本人草稿', (b: typeof base) => ({ ...b, ownerID: 9 })],
  ['非 draft 状态', (b: typeof base) => ({ ...b, intakeStatus: 'submitted' })],
  ['状态未知', (b: typeof base) => ({ ...b, intakeStatus: undefined })],
  ['补正模式', (b: typeof base) => ({ ...b, isSupplementMode: true })],
  ['无版本', (b: typeof base) => ({ ...b, revision: '' })],
  ['其他草稿的旧版本', (b: typeof base) => ({ ...b, revision: 'legacy-intake-2' })],
  ['无服务端锚点', (b: typeof base) => ({ ...b, baseline: '' })],
])('严格资格任一条件缺失时 %s → blocked', (_label, build) => {
  expect(narrowPilotIntakeEligibility(build(base)).mode).toBe('blocked')
})

test('无来源历史的本人草稿首次整理可走 strict', () => {
  expect(narrowPilotIntakeEligibility({ ...base, hasSourceHistory: false })).toEqual({ mode: 'strict' })
})

test('来源历史未知时阻止保存，且无历史证据不能当作无历史', () => {
  expect(narrowPilotIntakeEligibility({ ...base, sourceHistoryKnown: false })).toMatchObject({ mode: 'blocked' })
  expect(narrowPilotIntakeEligibility({ ...base, sourceHistoryKnown: false, hasSourceHistory: false }).mode).toBe('blocked')
})

test.each(['admin', 'assistant'])('%s 的无历史草稿走原生；已保护或未知历史阻止回退', role => {
  const input = { ...base, role, isAssistant: role === 'assistant', hasSourceHistory: false }
  expect(narrowPilotIntakeEligibility(input).mode).toBe('native')
  expect(narrowPilotIntakeEligibility({ ...input, hasSourceHistory: true }).mode).toBe('blocked')
  expect(narrowPilotIntakeEligibility({ ...input, sourceHistoryKnown: false }).mode).toBe('blocked')
})
test('账号信息缺失以及没有真实所有者的草稿均不走严格保存', () => {
  for (const missing of [{ userIsActive: undefined }, { role: undefined }, { ownerID: undefined }]) {
    expect(narrowPilotIntakeEligibility({ ...base, ...missing }).mode).toBe('blocked')
  }
})
test('新建继续使用原生创建', () => {
  expect(narrowPilotIntakeEligibility({ ...base, intakeID: '', sourceHistoryKnown: false }).mode).toBe('native')
})

test('projectionWithDescription 仅替换描述且保留敏感字段在投影对象内', () => {
  const next = projectionWithDescription(base.baseline, '回执摘要')
  const expected = JSON.parse(base.baseline)
  expected.form.description = '回执摘要'
  expect(JSON.parse(next)).toEqual(expected)
  expect(next).toContain('回执摘要')
  expect(next).toContain('911')
})
