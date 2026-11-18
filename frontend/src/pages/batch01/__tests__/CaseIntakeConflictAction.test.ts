import React from 'react'
import { act, configure, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import {
  CaseIntakeWorkbench,
  caseIntakeConflictFingerprint,
  caseIntakeConflictResultStale,
  conflictMatchesCaseContext,
  conflictRecordMatchType,
  conflictCheckRequiredFieldLabels,
  getConflictCheckFallbackMessage,
  getMissingConflictCheckFields,
  firstRunGuidance,
  purgeLegacyConflictFingerprintStorage,
  persistableCaseIntakeDraft,
  scopedCaseIntakeDraftKey,
} from '../Batch01Prototype'

const snapshotLocalStorage = (): Array<[string, string]> => {
  const entries: Array<[string, string]> = []
  for (let index = 0; index < window.localStorage.length; index += 1) {
    const key = window.localStorage.key(index)
    if (key === null) continue
    const value = window.localStorage.getItem(key)
    if (value !== null) entries.push([key, value])
  }
  return entries
}

const restoreLocalStorage = (entries: Array<[string, string]>) => {
  for (let index = window.localStorage.length - 1; index >= 0; index -= 1) {
    const key = window.localStorage.key(index)
    if (key !== null) window.localStorage.removeItem(key)
  }
  for (const [key, value] of entries) {
    window.localStorage.setItem(key, value)
  }
}

describe('CaseIntakeWorkbench conflict action', () => {
  it('returns stable required fields before any draft or conflict request', () => {
    const missing = getMissingConflictCheckFields({
      title: ' ',
      clientId: 0,
      clientName: '',
      opponentName: '',
      opponentIdentityNumber: '  ',
      lawyerId: 0,
      caseType: '',
      businessArea: '',
      subArea: '',
    })

    expect(missing).toEqual([
      'title',
      'client',
      'opponentName',
      'opponentIdentityNumber',
      'lawyer',
      'caseType',
      'businessArea',
      'subArea',
    ])
    expect(missing.map((field) => conflictCheckRequiredFieldLabels[field])).toEqual([
      '案件名称',
      '客户',
      '对方当事人',
      '对方身份标识',
      '负责律师',
      '案件类型',
      '业务领域',
      '子领域',
    ])
  })

  it('returns no missing fields for a complete conflict-check input', () => {
    expect(getMissingConflictCheckFields({
      title: '股权争议',
      clientId: 42,
      clientName: '示例客户',
      opponentName: '甲公司',
      opponentIdentityNumber: '91310000TEST000001',
      lawyerId: 7,
      caseType: 'commercial',
      businessArea: '公司与并购',
      subArea: '投资与融资',
    })).toEqual([])
  })

  it('labels a coverage-limited record with no evidence as no match', () => {
    expect(conflictRecordMatchType({
      has_conflict: false,
      check_result: {
        decision: { status: 'REVIEW_REQUIRED', coverageStatus: 'COVERAGE_LIMITED' },
      },
    })).toBe('无匹配')
  })

  it('keeps a clear MVP fallback message for unavailable conflict checks', () => {
    expect(getConflictCheckFallbackMessage()).toBe(
      '试用版当前使用样例冲突复核流程，请在利益冲突工作台查看待复核事项。',
    )
  })

  it('provides a lawyer-readable first-run path and support guidance', () => {
    expect(firstRunGuidance.requiredChecklist).toEqual([
      '案件名称',
      '客户',
      '对方当事人',
      '对方身份标识',
      '负责律师',
      '案件类型',
      '业务领域',
      '子领域',
    ])
    expect(firstRunGuidance.nextStep).toContain('保存最新输入并检测')
    expect(firstRunGuidance.restoreNotice).toContain('不保留对方身份标识')
    expect(firstRunGuidance.restoreNotice).toContain('重新填写')
    expect(firstRunGuidance.restoreNotice).toContain('恢复后先看续办状态')
    expect(firstRunGuidance.restoreNotice).toContain('系统核验有效无需重复录入')
    expect(firstRunGuidance.restoreNotice).toContain('提示要求时才重新填写并检测')
    expect(firstRunGuidance.restoreNotice).not.toContain('恢复后请重新填写')
    expect(firstRunGuidance.saveFailure).toContain('已填写内容仍保留在本页')
    expect(firstRunGuidance.saveFailure).toContain('联系律所管理员')
    expect(firstRunGuidance.help.content).toContain('保存并退出')
    expect(firstRunGuidance.help.content).toContain('联系律所管理员')
    expect(JSON.stringify(firstRunGuidance)).not.toMatch(/API|ms|接口响应|诊断/)
  })

  it('does not treat a matched conflict case as the subject case', () => {
    const task = {
      id: 'CHECK-OTHER',
      title: '其他案件的检测任务',
      conflict_cases: [{ case_id: '35', case_no: 'DEMO-MVP-2026-002', case_name: '当前案件' }],
    }

    expect(conflictMatchesCaseContext(task, '35', 'DEMO-MVP-2026-002', '当前案件')).toBe(false)
  })

  it('matches only explicit subject case identifiers', () => {
    const task = {
      id: 'CHECK-35',
      case_id: '35',
      case_number: 'DEMO-MVP-2026-002',
      title: '当前案件',
      conflict_cases: [{ case_id: '99', case_name: '历史命中案件' }],
    }

    expect(conflictMatchesCaseContext(task, '35', 'DEMO-MVP-2026-002', '当前案件')).toBe(true)
    expect(conflictMatchesCaseContext(task, '99', '', '历史命中案件')).toBe(false)
  })

  it('matches subject identifiers nested in the persisted search parameters', () => {
    const task = {
      id: 'CHECK-44',
      title: '主体案件名称',
      search_parameters: {
        subjectCaseId: '44',
        subjectCaseNumber: 'DEMO-SUBJECT-2026-001',
      },
      conflict_cases: [{ case_id: '44', case_no: '历史命中案件' }],
    }

    expect(conflictMatchesCaseContext(task, '44', 'DEMO-SUBJECT-2026-001', '主体案件名称')).toBe(true)
  })

  it('matches the command-center task title used by a case-context link', () => {
    const task = {
      id: 'CHECK-CURRENT',
      title: '当前案件检测任务',
      conflict_cases: [{ case_id: '99', case_name: '历史命中案件' }],
    }

    expect(conflictMatchesCaseContext(task, '', '', '当前案件检测任务')).toBe(true)
  })

  it('does not use a historical hit case as the subject context', () => {
    const task = {
      id: 'CHECK-OTHER',
      title: '另一项检测',
      conflict_cases: [{ case_id: '44', case_no: 'DEMO-SUBJECT-2026-001', case_name: '当前案件' }],
    }

    expect(conflictMatchesCaseContext(task, '44', 'DEMO-SUBJECT-2026-001', '当前案件')).toBe(false)
  })

  it('isolates local drafts by user and case context', () => {
    expect(scopedCaseIntakeDraftKey('7')).toBe('law-oa-case-intake-draft-v2:7:new')
    expect(scopedCaseIntakeDraftKey('7', '35')).toBe('law-oa-case-intake-draft-v2:7:case-35')
    expect(scopedCaseIntakeDraftKey('8', '35')).not.toBe(scopedCaseIntakeDraftKey('7', '35'))
  })

  it('never persists an opponent identity number in a browser draft', () => {
    const draft = persistableCaseIntakeDraft({
      intakeId: '', intakeCode: '', idempotencyKey: '', title: '股权争议', caseType: 'commercial',
      businessArea: '', subArea: '', priority: 'medium', disputeAmount: '', sourceChannel: '',
      sourceContact: '', investmentAgreementDate: '', disputeDate: '', breachDate: '',
      proposedFilingDate: '', jurisdiction: '', clientId: 42, clientName: '示例客户',
      opponentName: '甲公司', opponentEntityType: 'LEGAL_PERSON',
      opponentIdentityType: 'SOCIAL_CREDIT_CODE', opponentIdentityNumber: '91310000TEST000001',
      opponentAliases: '', description: '', billingMethod: '', feeBase: '', contingencyRate: '',
      minimumFee: '', lawyerId: 7,
    })

    expect(draft.opponentIdentityNumber).toBe('')
    expect(JSON.stringify(draft)).not.toContain('91310000TEST000001')
  })

  it('invalidates a frozen result when a material conflict input changes', () => {
    const form = { clientId: 42, clientName: '示例客户', opponentName: '甲公司', opponentEntityType: 'LEGAL_PERSON', opponentIdentityType: 'SOCIAL_CREDIT_CODE', opponentIdentityNumber: '91310000TEST000001', opponentAliases: '', title: '股权争议', caseType: 'commercial', lawyerId: 7 }
    const related = { name: '保证人乙', role: 'GUARANTOR', entityType: 'LEGAL_PERSON', identityType: 'SOCIAL_CREDIT_CODE', identityNumber: '91310000TEST000002' }
    const frozen = caseIntakeConflictFingerprint(form, [related])
    expect(caseIntakeConflictFingerprint({ ...form, opponentName: '乙公司' }, [related])).not.toBe(frozen)
    expect(caseIntakeConflictFingerprint({ ...form, lawyerId: 8 }, [related])).not.toBe(frozen)
    expect(caseIntakeConflictFingerprint(form, [{ ...related, name: '新增实控人' }])).not.toBe(frozen)
  })

  it('never embeds identity numbers in the persistable fingerprint shape', () => {
    const form = { clientId: 42, clientName: '示例客户', opponentName: '甲公司', opponentEntityType: 'LEGAL_PERSON', opponentIdentityType: 'SOCIAL_CREDIT_CODE', opponentIdentityNumber: '91310000TEST000001', opponentAliases: '', title: '股权争议', caseType: 'commercial', lawyerId: 7 }
    const related = { name: '保证人乙', role: 'GUARANTOR', entityType: 'LEGAL_PERSON', identityType: 'SOCIAL_CREDIT_CODE', identityNumber: '91310000TEST000002' }
    expect(caseIntakeConflictFingerprint(form, [related])).not.toContain('91310000TEST000001')
    expect(caseIntakeConflictFingerprint(form, [related])).not.toContain('91310000TEST000002')
    expect(caseIntakeConflictFingerprint(form, [related])).not.toContain('identityNumber')
  })

  it('keeps the same fingerprint when related parties are only reordered', () => {
    const form = { clientId: 42, clientName: '示例客户', opponentName: '甲公司', opponentEntityType: 'LEGAL_PERSON', opponentIdentityType: 'SOCIAL_CREDIT_CODE', opponentIdentityNumber: '91310000TEST000001', opponentAliases: '', title: '股权争议', caseType: 'commercial', lawyerId: 7 }
    const partyA = { name: '乙', role: 'GUARANTOR', entityType: 'LEGAL_PERSON', identityType: 'SOCIAL_CREDIT_CODE', identityNumber: '91310000TEST000002' }
    const partyB = { name: '丙', role: 'CONTROLLER', entityType: 'INDIVIDUAL', identityType: 'ID_CARD', identityNumber: '110101199001010011' }
    const left = caseIntakeConflictFingerprint(form, [partyA, partyB])
    const right = caseIntakeConflictFingerprint(form, [partyB, partyA])
    expect(right).toBe(left)
  })

  it('marks a conflict result stale when only identity inputs change', () => {
    const form = { clientId: 42, clientName: '示例客户', opponentName: '甲公司', opponentEntityType: 'LEGAL_PERSON', opponentIdentityType: 'SOCIAL_CREDIT_CODE', opponentIdentityNumber: '91310000TEST000001', opponentAliases: '', title: '股权争议', caseType: 'commercial', lawyerId: 7 }
    const related = { name: '保证人乙', role: 'GUARANTOR', entityType: 'LEGAL_PERSON', identityType: 'SOCIAL_CREDIT_CODE', identityNumber: '91310000TEST000002' }
    const fingerprint = caseIntakeConflictFingerprint(form, [related])
    const snapshot = JSON.stringify({
      opponentIdentityNumber: '91310000TEST000001',
      relatedParties: [{ name: '保证人乙', identityNumber: '91310000TEST000002' }],
    })
    const runtime = {
      conflict: { task_id: 'TASK-1' },
      conflictInputFingerprint: fingerprint,
      conflictIdentitySnapshot: snapshot,
      intakeIdentityVerified: true,
    }
    expect(caseIntakeConflictResultStale(runtime, fingerprint, snapshot)).toBe(false)
    expect(
      caseIntakeConflictResultStale(
        { ...runtime, conflictIdentitySnapshot: JSON.stringify({ ...JSON.parse(snapshot), opponentIdentityNumber: '91310000TEST999999' }) },
        fingerprint,
        snapshot,
      ),
    ).toBe(true)
    expect(
      caseIntakeConflictResultStale(
        { ...runtime, intakeIdentityVerified: false },
        fingerprint,
        snapshot,
      ),
    ).toBe(true)
  })

  it('fails closed for a restored draft so identity is re-entered and the check re-run', () => {
    const restoredRuntime = {
      conflict: { task_id: 'TASK-1' },
      conflictInputFingerprint: 'fingerprint-from-past-session',
      intakeIdentityVerified: false,
    }
    expect(
      caseIntakeConflictResultStale(
        restoredRuntime,
        'fingerprint-from-past-session',
        'identity-snapshot-from-past-session',
      ),
    ).toBe(true)
    expect(
      caseIntakeConflictResultStale(
        { ...restoredRuntime, conflictInputFingerprint: undefined, conflictIdentitySnapshot: undefined },
        'current-fingerprint',
        'current-identity-snapshot',
      ),
    ).toBe(true)
  })

  it('sweeps legacy fingerprint keys inside the draft namespace only', () => {
    const previousStorage = snapshotLocalStorage()
    try {
      window.localStorage.setItem('law-oa-case-intake-draft-v2:7:new', '{"title":"kept"}')
      window.localStorage.setItem('law-oa-case-intake-draft-v2:7:new:conflictFingerprint', 'raw-1')
      window.localStorage.setItem('law-oa-case-intake-draft-v2:9:case-35:conflictFingerprint', 'raw-2')
      window.localStorage.setItem('other-app:conflictFingerprint', 'keep-other-app')
      window.localStorage.setItem('law-oa-case-intake-draft-v2:7:new:conflictResult', 'keep-result')

      purgeLegacyConflictFingerprintStorage()

      expect(window.localStorage.getItem('law-oa-case-intake-draft-v2:7:new:conflictFingerprint')).toBeNull()
      expect(window.localStorage.getItem('law-oa-case-intake-draft-v2:9:case-35:conflictFingerprint')).toBeNull()
      expect(window.localStorage.getItem('law-oa-case-intake-draft-v2:7:new')).toBe('{"title":"kept"}')
      expect(window.localStorage.getItem('other-app:conflictFingerprint')).toBe('keep-other-app')
      expect(window.localStorage.getItem('law-oa-case-intake-draft-v2:7:new:conflictResult')).toBe('keep-result')
    } finally {
      restoreLocalStorage(previousStorage)
    }
  })
})



// ---------------------------------------------------------------------------
// Manual reviewer-decision refresh (component path). The lawyer page must
// pick up a review verdict submitted after the last automatic load without
// re-running the conflict check and without touching facts or identity
// inputs. Extras are scoped to the check they were read for, so a new check
// round can never inherit the previous round's verdict.
// ---------------------------------------------------------------------------

jest.mock('@/utils/storage', () => ({
  ...jest.requireActual('@/utils/storage'),
  getRoles: jest.fn(() => []),
  getToken: jest.fn(() => 'test-token'),
  getUserInfo: jest.fn(() => ({ id: 7, role: 'lawyer' })),
}))

jest.mock('@/utils/messageHelper', () => ({
  message: { error: jest.fn(), success: jest.fn(), info: jest.fn(), warning: jest.fn() },
  setAppMessage: jest.fn(),
}))

const fetchMock = jest.fn()
global.fetch = fetchMock as unknown as typeof fetch

// The workbench mounts with many cascading async state updates; on a loaded
// CI runner the default 1s async-util window is too tight and produced
// flaky timeouts. The assertions themselves are unchanged.
configure({ asyncUtilTimeout: 4000 })

function componentJsonResponse(data: unknown) {
  return { ok: true, json: async () => ({ success: true, data }) } as Response
}

function componentErrorResponse(status: number) {
  return {
    ok: false,
    status,
    json: async () => ({ success: false, message: '复核记录不存在' }),
  } as Response
}

const draftFormForTest = {
  intakeId: 'INTAKE-1',
  intakeCode: 'INT-2026-001',
  idempotencyKey: '',
  title: '股权争议',
  caseType: 'commercial',
  businessArea: '投资并购',
  subArea: '股权回购',
  priority: 'high',
  disputeAmount: '',
  sourceChannel: '',
  sourceContact: '',
  investmentAgreementDate: '',
  disputeDate: '',
  breachDate: '',
  proposedFilingDate: '',
  jurisdiction: '',
  clientId: 42,
  clientName: '示例客户',
  opponentName: '甲公司',
  opponentEntityType: 'LEGAL_PERSON',
  opponentIdentityType: 'SOCIAL_CREDIT_CODE',
  opponentIdentityNumber: '',
  opponentAliases: '',
  description: '股权回购争议案情摘要',
  lawyerId: 7,
  revision: 'rev-1',
} as unknown as Parameters<typeof persistableCaseIntakeDraft>[0]

const conflictResultBase = (checkId: string, decisionStatus: string) => ({
  checkId,
  riskAssessment: { overallRisk: 'HIGH' },
  requiresApproval: true,
  decision: { status: decisionStatus, coverageStatus: 'COMPLETE' },
  conflictCases: [{ case_number: 'HIT-1', case_name: '在办冲突案' }],
})

function workbenchFetchImplementation(reviewResponse: () => Response) {
  let postCheckCounter = 1
  let revisionCounter = 1
  let checked = false
  return async (input: RequestInfo | URL, init?: RequestInit) => {
    const path = String(input)
    const method = String(init?.method || 'GET').toUpperCase()
    if (path.endsWith('/case-intakes/INTAKE-1') && method === 'GET') {
      return componentJsonResponse({ intake: {
        id: 'INTAKE-1', intake_code: 'INT-2026-001', title: '股权争议', case_type: 'commercial',
        priority: 'high', description: '股权回购争议案情摘要', client_id: 42,
        metadata: { lawyer_id: 7, business_area: '投资并购', sub_area: '股权回购' },
      }, revision: 'rev-' + revisionCounter, parties: [
        { entity_name: '示例客户', entity_type: 'LEGAL_PERSON', party_role: 'client' },
        { entity_name: '甲公司', entity_type: 'LEGAL_PERSON', party_role: 'opposing_party' },
      ] })
    }
    if (path.endsWith('/case-intakes/INTAKE-1') && method === 'PUT') {
      revisionCounter += 1
      checked = false
      return componentJsonResponse({ id: 'INTAKE-1', intake_code: 'INT-2026-001', revision: 'rev-' + revisionCounter })
    }
    if (path.endsWith('/facts-confirmation') && method === 'POST') {
      revisionCounter += 1
      return componentJsonResponse({ revision: 'rev-' + revisionCounter })
    }
    if (path.endsWith('/continuation')) {
      const response = reviewResponse()
      if (!response.ok) return response
      const body = await response.json()
      const reviewed = body.data?.decision === 'no_conflict'
      return componentJsonResponse({
        intake_id: 'INTAKE-1', revision: 'rev-' + revisionCounter, status: checked ? 'conflict_ready' : 'draft',
        can_reuse_identity: checked, can_reuse_check: checked, can_submit: checked && reviewed, recheck_required: !checked,
        reasons: checked ? reviewed ? [] : ['REVIEW_MISSING'] : ['FACTS_UNCONFIRMED'],
        facts: { facts_version: 'facts-' + postCheckCounter, confirmed_at: '2026-10-02' },
        conflict_binding: { check_id: 'CCT-' + postCheckCounter, check_status: 'COMPLETED', coverage_status: 'COMPLETE',
          ...(reviewed ? { review_id: 11, reviewer_id: 9, review_decision: 'no_conflict' } : {}) },
      })
    }
    if (path.includes('/conflict/tasks/') && path.endsWith('/review')) {
      return Promise.resolve(reviewResponse())
    }
    if (path.includes('/conflict/tasks/') && path.endsWith('/waiver')) {
      return Promise.resolve(componentJsonResponse({}))
    }
    if (path.includes('/conflict/history')) {
      return Promise.resolve(
        componentJsonResponse([
          {
            checkId: 'CCT-1',
            checkStatus: 'COMPLETED',
            checkTime: '2026-09-26T09:00:00Z',
            searchParameters: { intakeId: 'INTAKE-1' },
            checkResult: conflictResultBase('CCT-1', 'REVIEW_REQUIRED'),
          },
        ]),
      )
    }
    if (path === '/api/v1/case-intakes') {
      return Promise.resolve(
        componentJsonResponse({
          id: 'INTAKE-1',
          intake_code: 'INT-2026-001',
          revision: 'rev-2',
        }),
      )
    }
    if (path.endsWith('/conflict-check') && method === 'POST') {
      postCheckCounter += 1
      checked = true
      const taskID = 'CCT-' + postCheckCounter
      return Promise.resolve(
        componentJsonResponse({
          task_id: taskID,
          status: 'COMPLETED',
          result: conflictResultBase(taskID, 'REVIEW_REQUIRED'),
          review: null,
        }),
      )
    }
    return Promise.resolve(componentJsonResponse({}))
  }
}

const reviewGetCount = () =>
  fetchMock.mock.calls.filter(([input]) => String(input).endsWith('/review')).length

const continuationReadCount = () =>
  fetchMock.mock.calls.filter(([input]) => String(input).endsWith('/continuation')).length

const conflictCheckPostCount = () =>
  fetchMock.mock.calls.filter(
    ([input, init]) =>
      String(input).endsWith('/conflict-check') &&
      String(init?.method || '').toUpperCase() === 'POST',
  ).length

const bannerHeadline = (fragment: string) => {
  const matches = screen.getAllByText((content, element) => {
    if (element?.tagName.toLowerCase() !== 'strong') return false
    const hasText = (element.textContent || '').includes('检查状态：')
    const hasFragment = (element.textContent || '').includes(fragment)
    return hasText && hasFragment
  })
  expect(matches.length).toBeGreaterThanOrEqual(1)
  return matches[0]
}
// antd disables a Button while it is loading, and a click on a disabled
// button is silently dropped. Refresh clicks must wait for the previous
// (auto or manual) refresh to leave the loading state before firing.
async function clickRefreshWhenReady() {
  const refreshButton = await screen.findByRole('button', { name: '刷新复核状态' })
  await waitFor(() => expect(refreshButton).not.toHaveClass('ant-btn-loading'))
  fireEvent.click(refreshButton)
}
async function openWorkbenchWithRestoredDraft() {
  window.localStorage.setItem(
    scopedCaseIntakeDraftKey('7'),
    JSON.stringify(persistableCaseIntakeDraft(draftFormForTest)),
  )
  render(React.createElement(MemoryRouter, null, React.createElement(CaseIntakeWorkbench)))
  fireEvent.click(await screen.findByRole('button', { name: '继续未完成草稿' }))
  await waitFor(() =>
    expect(continuationReadCount()).toBeGreaterThan(0),
  )
  expect(fetchMock.mock.calls.some((c) => String(c[0]).includes('/conflict/history'))).toBe(false)
  // Local flags or client history never authorize this restored revision.
  const stepperButtons = document.querySelectorAll('.batch-stepper button')
  expect(stepperButtons.length).toBeGreaterThanOrEqual(2)
  fireEvent.click(stepperButtons[1])
  expect(screen.getByRole('button', { name: '提交审批并等待成案' })).toBeDisabled()
  expect(reviewGetCount()).toBe(0)
}

async function reenterIdentityAndRunCheck(options: { waitAutoRefresh?: boolean; expectedHeadline?: string } = {}) {
  // Step panels are conditionally rendered: the identity input exists only
  // on the basic-info step, so go back there before typing.
  const stepButtons = document.querySelectorAll('.batch-stepper button')
  fireEvent.click(stepButtons[0])
  const identityInput = document.querySelector(
    '#intake-opponent-identity-number',
  ) as HTMLInputElement | null
  expect(identityInput).not.toBeNull()
  fireEvent.change(identityInput as HTMLInputElement, { target: { value: '91310000TEST000001' } })
  // Move to the conflict step to run the check where its banner lives.
  fireEvent.click(document.querySelectorAll('.batch-stepper button')[1])
  fireEvent.click(await screen.findByRole('button', { name: '保存并进行利益冲突检查' }))
  await waitFor(() => expect(reviewGetCount()).toBe(1))
  await waitFor(() => bannerHeadline(options.expectedHeadline || '等待独立复核'))
  expect(screen.getByRole('button', { name: '刷新复核状态' })).toBeInTheDocument()
  // The auto refresh fired right after the check writes decision extras
  // only after its review + waiver reads settle. Waiting for the waiver
  // call keeps later refresh clicks from racing that in-flight update.
  // The 404 variant never reaches the waiver read, so callers opt in.
  if (options.waitAutoRefresh) {
    await waitFor(() =>
      expect(fetchMock.mock.calls.some(([input]) => String(input).endsWith('/waiver'))).toBe(true),
    )
  }
}

describe('CaseIntakeWorkbench manual review refresh', () => {
  let reviewPayload: Record<string, unknown>

  beforeEach(() => {
    reviewPayload = { decision: 'pending' }
    fetchMock.mockReset()
    fetchMock.mockImplementation(
      workbenchFetchImplementation(() => componentJsonResponse(reviewPayload)),
    )
  })

  it('unlocks submission after refreshing a pending check to no_conflict without re-running it', async () => {
    const previousStorage = snapshotLocalStorage()
    try {
      await openWorkbenchWithRestoredDraft()
      await reenterIdentityAndRunCheck({ waitAutoRefresh: true })
      await waitFor(() => bannerHeadline('等待独立复核'))

      const readsBefore = continuationReadCount()
      await clickRefreshWhenReady()
      await waitFor(() => expect(continuationReadCount()).toBe(readsBefore + 1))
      await waitFor(() => bannerHeadline('等待独立复核'))
      expect(screen.getByRole('button', { name: '提交审批并等待成案' })).toBeDisabled()

      reviewPayload = { decision: 'no_conflict' }
      await waitFor(() => bannerHeadline('等待独立复核'))

      await clickRefreshWhenReady()
      await waitFor(() => bannerHeadline('独立复核已通过，可继续提交'))
      await waitFor(() =>
        expect(screen.getByRole('button', { name: '提交审批并等待成案' })).toBeEnabled(),
      )

      expect(conflictCheckPostCount()).toBe(1)
      expect(
        fetchMock.mock.calls.filter(([input]) => String(input).endsWith('/facts-confirmation'))
          .length,
      ).toBe(1)
      const persistedDraft = JSON.parse(
        window.localStorage.getItem(scopedCaseIntakeDraftKey('7')) || '{}',
      )
      expect(persistedDraft.opponentIdentityNumber).toBe('')
      expect(persistedDraft.opponentName).toBe('甲公司')
      expect(persistedDraft.title).toBe('股权争议')
      expect(persistedDraft.clientId).toBe(42)
    } finally {
      restoreLocalStorage(previousStorage)
    }
  })

  it('never lets a previous round no_conflict verdict unlock a new check round', async () => {
    const previousStorage = snapshotLocalStorage()
    try {
      await openWorkbenchWithRestoredDraft()
      await reenterIdentityAndRunCheck({ waitAutoRefresh: true })

      // Round A: CCT-2 becomes reviewed no_conflict and unlocks submission.
      reviewPayload = { decision: 'no_conflict' }
      await clickRefreshWhenReady()
      await waitFor(() => bannerHeadline('独立复核已通过，可继续提交'))
      await waitFor(() =>
        expect(screen.getByRole('button', { name: '提交审批并等待成案' })).toBeEnabled(),
      )

      // Round B: the lawyer re-runs the check on unchanged facts. The new
      // task CCT-3 is unreviewed (pending), so the previous round's
      // verdict must not carry over and submission stays locked.
      reviewPayload = { decision: 'pending' }
      fireEvent.click(await screen.findByRole('button', { name: '运行利益冲突检查' }))
      await waitFor(() => bannerHeadline('等待独立复核'))
      expect(screen.getByRole('button', { name: '提交审批并等待成案' })).toBeDisabled()

      // A manual refresh while CCT-3 is still pending keeps the gate shut.
      await clickRefreshWhenReady()
      await waitFor(() => bannerHeadline('等待独立复核'))
      expect(screen.getByRole('button', { name: '提交审批并等待成案' })).toBeDisabled()

      // Only a no_conflict verdict fetched for CCT-3 itself unlocks.
      reviewPayload = { decision: 'no_conflict' }
      await clickRefreshWhenReady()
      await waitFor(() => bannerHeadline('独立复核已通过，可继续提交'))
      await waitFor(() =>
        expect(screen.getByRole('button', { name: '提交审批并等待成案' })).toBeEnabled(),
      )

      expect(conflictCheckPostCount()).toBe(2)
      expect(
        fetchMock.mock.calls.filter(([input]) => String(input).endsWith('/facts-confirmation'))
          .length,
      ).toBe(2)
      const persistedDraft = JSON.parse(
        window.localStorage.getItem(scopedCaseIntakeDraftKey('7')) || '{}',
      )
      expect(persistedDraft.opponentIdentityNumber).toBe('')
      expect(persistedDraft.title).toBe('股权争议')
      expect(persistedDraft.opponentName).toBe('甲公司')
      expect(persistedDraft.clientId).toBe(42)
    } finally {
      restoreLocalStorage(previousStorage)
    }
  })

  it('keeps the gate locked when the refreshed review cannot be read', async () => {
    const previousStorage = snapshotLocalStorage()
    try {
      fetchMock.mockImplementation(workbenchFetchImplementation(() => componentErrorResponse(404)))
      await openWorkbenchWithRestoredDraft()
      await reenterIdentityAndRunCheck({ expectedHeadline: '需要独立人工复核' })
      await waitFor(() => bannerHeadline('需要独立人工复核'))
      expect(screen.getByRole('button', { name: '提交审批并等待成案' })).toBeDisabled()
      expect(conflictCheckPostCount()).toBe(1)

      const readsBefore = continuationReadCount()
      await clickRefreshWhenReady()
      await waitFor(() => expect(continuationReadCount()).toBe(readsBefore + 1))
      await waitFor(() => bannerHeadline('需要独立人工复核'))
      expect(screen.getByRole('button', { name: '提交审批并等待成案' })).toBeDisabled()
      expect(conflictCheckPostCount()).toBe(1)
    } finally {
      restoreLocalStorage(previousStorage)
    }
  })
})
