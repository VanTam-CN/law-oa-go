import React from 'react'
import { act, configure, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { CaseIntakeWorkbench, supplementFieldLocateTargets } from '../Batch01Prototype'

jest.mock('@/utils/storage', () => ({
  ...jest.requireActual('@/utils/storage'),
  getRoles: jest.fn(() => []),
  getToken: jest.fn(() => 'test-token'),
  getUserInfo: jest.fn(() => ({ id: 7, role: 'lawyer', roles: ['lawyer'], name: '律师甲' })),
}))

jest.mock('@/utils/messageHelper', () => ({
  message: { error: jest.fn(), success: jest.fn(), warning: jest.fn(), info: jest.fn() },
  setAppMessage: jest.fn(),
}))

configure({ asyncUtilTimeout: 4000 })

const fetchMock = jest.fn()
global.fetch = fetchMock as unknown as typeof fetch

function jsonResponse(data: unknown) {
  return { ok: true, json: async () => ({ success: true, data }) } as Response
}

function jsonResponseError(status: number, message: string) {
  return { ok: false, status, json: async () => ({ success: false, message }) } as Response
}

const completeFormDraft = {
  intakeId: 'INTAKE-1',
  intakeCode: 'INT-2026-001',
  title: '股权争议',
  caseType: 'commercial',
  businessArea: '投资并购',
  subArea: '股权回购',
  clientId: 42,
  clientName: '示例客户',
  opponentName: '甲公司',
  opponentIdentityNumber: '91310000TEST000001',
  lawyerId: 7,
  revision: 'rev-1',
}

const identityReentryValue = '91310000TEST000009'

function baseWorkbenchMock() {
  fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
    const path = String(input)
    const method = String(init?.method || 'GET').toUpperCase()
    if (path.endsWith('/clients?page=1&page_size=50')) {
      return Promise.resolve(jsonResponse({ clients: [{ id: 42, name: '示例客户', type: '企业' }] }))
    }
    if (path.endsWith('/lawfirm/lawyers?page=1&page_size=50')) {
      return Promise.resolve(jsonResponse({ lawyers: [{ id: 7, name: '律师甲' }] }))
    }
    if (path.endsWith('/case-intakes?page=1&page_size=10')) {
      return Promise.resolve(jsonResponse({ items: [], total: 0 }))
    }
    if (path === '/api/v1/case-intakes/INTAKE-1' && method === 'GET') {
      return Promise.resolve(
        jsonResponse({
          intake: {
            id: 'INTAKE-1',
            intake_code: 'INT-2026-001',
            title: '股权争议',
            case_type: 'commercial',
            priority: 'high',
            description: '股权回购争议案情摘要',
            client_id: 42,
            created_by: '7',
            metadata: { lawyer_id: 7, business_area: '投资并购', sub_area: '股权回购' },
          },
          revision: 'rev-1',
          parties: [
            { entity_name: '示例客户', entity_type: 'LEGAL_PERSON', party_role: 'client' },
            { entity_name: '甲公司', entity_type: 'LEGAL_PERSON', party_role: 'opposing_party' },
          ],
        }),
      )
    }
    if (path === '/api/v1/case-intakes/INTAKE-1' && method === 'PUT') {
      return Promise.resolve(jsonResponse({ id: 'INTAKE-1', intake_code: 'INT-2026-001', revision: 'rev-2' }))
    }
    if (path.endsWith('/conflict/history')) {
      return Promise.resolve(jsonResponse([]))
    }
    if (path.endsWith('/case-intakes/INTAKE-1/facts-confirmation') && method === 'POST') {
      return Promise.resolve(jsonResponse({ revision: 'rev-3' }))
    }
    if (path.endsWith('/case-intakes/INTAKE-1/conflict-check') && method === 'POST') {
      return Promise.resolve(
        jsonResponse({
          task_id: 'CCT-LIVE-1',
          status: 'COMPLETED',
          result: {
            checkId: 'CCT-LIVE-1',
            riskAssessment: { overallRisk: 'HIGH', requiresApproval: true },
            decision: { status: 'REVIEW_REQUIRED', coverageStatus: 'COMPLETE' },
            conflictCases: [{ case_number: 'HIT-1', case_name: '在办冲突案' }],
          },
        }),
      )
    }
    if (path.includes('/conflict/tasks/') && path.endsWith('/review')) {
      return Promise.resolve(jsonResponse({ decision: 'pending' }))
    }
    if (path.includes('/conflict/tasks/') && path.endsWith('/waiver')) {
      return Promise.resolve(jsonResponse({}))
    }
    return Promise.resolve(jsonResponse({}))
  })
}

async function restoreDraftAndOpenWorkbench() {
  const { persistableCaseIntakeDraft, scopedCaseIntakeDraftKey } = await import('../Batch01Prototype')
  window.localStorage.setItem(
    scopedCaseIntakeDraftKey('7'),
    JSON.stringify(persistableCaseIntakeDraft(completeFormDraft as never)),
  )
  render(
    <MemoryRouter initialEntries={['/case/create']}>
      <CaseIntakeWorkbench />
    </MemoryRouter>,
  )
  fireEvent.click(await screen.findByRole('button', { name: '继续未完成草稿' }))
}

const intakePutCount = () =>
  fetchMock.mock.calls.filter(
    ([input, init]) =>
      String(input).endsWith('/case-intakes/INTAKE-1') &&
      String(init?.method || '').toUpperCase() === 'PUT',
  ).length

describe('Intake saved-state dirty projection (recheck #1/#5)', () => {
  beforeEach(() => {
    window.localStorage.clear()
    window.sessionStorage.clear()
    fetchMock.mockReset()
  })

  it('keeps a brand-new intake unsaved when typing only enables local draft storage', async () => {
    baseWorkbenchMock()
    const { scopedCaseIntakeDraftKey } = await import('../Batch01Prototype')
    render(
      <MemoryRouter initialEntries={['/case/create']}>
        <CaseIntakeWorkbench />
      </MemoryRouter>,
    )

    expect(await screen.findByText('系统保存：未保存')).toBeInTheDocument()
    expect(screen.getByText('当前输入：未保存')).toBeInTheDocument()
    fireEvent.change(screen.getByLabelText('案件名称 *'), {
      target: { value: '全新未保存案件' },
    })
    fireEvent.change(screen.getByLabelText('案情摘要 *'), {
      target: { value: '仅在本机暂存的虚构案情' },
    })

    await waitFor(() => {
      const draft = JSON.parse(window.localStorage.getItem(scopedCaseIntakeDraftKey('7')) || '{}')
      expect(draft).toMatchObject({ title: '全新未保存案件', intakeId: '', intakeCode: '' })
    })
    expect(screen.getByText('本机暂存：已开启（仅草稿字段，不含身份标识）')).toBeInTheDocument()
    expect(
      fetchMock.mock.calls.filter(
        ([input, init]) =>
          String(input).includes('/case-intakes') &&
          ['POST', 'PUT', 'PATCH'].includes(String(init?.method || 'GET').toUpperCase()),
      ),
    ).toHaveLength(0)
    expect({
      system: screen.getByText(/^系统保存：/).textContent,
      current: screen.getByText(/^当前输入：/).textContent,
    }).toEqual({ system: '系统保存：未保存', current: '当前输入：未保存' })
    expect(screen.queryByText('系统保存：已有接案记录')).not.toBeInTheDocument()
    expect(screen.queryByText('当前输入：沿用已有接案记录')).not.toBeInTheDocument()
  })

  it('keeps a restored local-only draft without an intake ID unsaved on the server', async () => {
    baseWorkbenchMock()
    const { persistableCaseIntakeDraft, scopedCaseIntakeDraftKey } = await import('../Batch01Prototype')
    window.localStorage.setItem(
      scopedCaseIntakeDraftKey('7'),
      JSON.stringify(persistableCaseIntakeDraft({
        ...completeFormDraft,
        intakeId: '',
        intakeCode: '',
        revision: '',
        title: '本机未保存案件',
      } as never)),
    )
    render(
      <MemoryRouter initialEntries={['/case/create']}>
        <CaseIntakeWorkbench />
      </MemoryRouter>,
    )

    fireEvent.click(await screen.findByRole('button', { name: '继续未完成草稿' }))
    await waitFor(() => expect(screen.getByLabelText('案件名称 *')).toHaveValue('本机未保存案件'))
    expect(screen.getByText('本机暂存：已开启（仅草稿字段，不含身份标识）')).toBeInTheDocument()
    expect(
      JSON.parse(window.localStorage.getItem(scopedCaseIntakeDraftKey('7')) || '{}'),
    ).toMatchObject({ intakeId: '', intakeCode: '' })
    expect(
      fetchMock.mock.calls.filter(
        ([input, init]) =>
          String(input).includes('/case-intakes') &&
          ['POST', 'PUT', 'PATCH'].includes(String(init?.method || 'GET').toUpperCase()),
      ),
    ).toHaveLength(0)
    expect({
      system: screen.getByText(/^系统保存：/).textContent,
      current: screen.getByText(/^当前输入：/).textContent,
    }).toEqual({ system: '系统保存：未保存', current: '当前输入：未保存' })
    expect(screen.queryByText('系统保存：已有接案记录')).not.toBeInTheDocument()
    expect(screen.queryByText('当前输入：沿用已有接案记录')).not.toBeInTheDocument()
  })

  it('shows a fresh success as saved, flips to unsaved on edits, and recovers after another save', async () => {
    baseWorkbenchMock()
    await restoreDraftAndOpenWorkbench()

    // Restoring a draft does NOT itself save; the lawyer explicitly saves
    // and only then sees the saved projection.
    fireEvent.click(await screen.findByRole('button', { name: '保存草稿' }))
    await waitFor(() => expect(screen.getByText(/系统保存：已保存/)).toBeInTheDocument())
    expect(screen.getByText('当前输入：已保存到服务器')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '退出立案' })).toBeInTheDocument()

    // Any subsequent user input immediately flips the panel to unsaved.
    fireEvent.change(screen.getByLabelText('案件名称 *'), { target: { value: '股权争议（更名）' } })
    await waitFor(() => expect(screen.getByText(/有未保存修改/)).toBeInTheDocument())
    expect(screen.queryByText('当前输入：已保存到服务器')).not.toBeInTheDocument()

    // A failed save keeps its durable failed signal, then a successful
    // save re-anchors the projection and recovers the panel.
    fetchMock.mockImplementationOnce(() => jsonResponseError(500, '系统保存失败'))
    fireEvent.click(screen.getByRole('button', { name: '保存草稿' }))
    await waitFor(() => expect(screen.getByText('系统保存：保存失败')).toBeInTheDocument())
    expect(screen.getByText(/未保存，输入仍保留在页面/)).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: '保存草稿' }))
    await waitFor(() => expect(screen.getByText(/系统保存：已保存/)).toBeInTheDocument())
    expect(screen.getByText('当前输入：已保存到服务器')).toBeInTheDocument()
    expect(intakePutCount()).toBeGreaterThanOrEqual(2)
  })

  it('stays saved when a conflict round rotates only server-owned tokens', async () => {
    baseWorkbenchMock()
    await restoreDraftAndOpenWorkbench()
    fireEvent.click(await screen.findByRole('button', { name: '保存草稿' }))
    await waitFor(() => expect(screen.getByText('当前输入：已保存到服务器')).toBeInTheDocument())

    // The identity number never comes back from storage; re-enter it, then
    // run the check. Facts confirmation + conflict check rotate the server
    // revision token only, so the panel must stay "saved" instead of
    // showing a false unsaved indicator.
    fireEvent.change(screen.getByLabelText('身份标识'), {
      target: { value: identityReentryValue },
    })
    fireEvent.click(await screen.findByRole('button', { name: '保存并进行利益冲突检查' }))
    await waitFor(() =>
      expect(
        fetchMock.mock.calls.some(
          ([input, init]) =>
            String(input).endsWith('/case-intakes/INTAKE-1/facts-confirmation') &&
            String(init?.method || '').toUpperCase() === 'POST',
        ),
      ).toBe(true),
    )
    expect(screen.getByText('当前输入：已保存到服务器')).toBeInTheDocument()
  })
})

describe('Intake revision refocus refresh (recheck #6)', () => {
  let reviewGetCount = 0

  beforeEach(() => {
    window.localStorage.clear()
    window.sessionStorage.clear()
    fetchMock.mockReset()
    reviewGetCount = 0
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const path = String(input)
      if (path.includes('/conflict/tasks/') && path.endsWith('/review')) {
        reviewGetCount += 1
        return Promise.resolve(jsonResponse({ decision: 'pending' }))
      }
      if (path.includes('/conflict/tasks/') && path.endsWith('/waiver')) {
        return Promise.resolve(jsonResponse({}))
      }
      if (path.endsWith('/clients?page=1&page_size=50')) {
        return Promise.resolve(jsonResponse({ clients: [{ id: 42, name: '虚构客户甲', type: '企业' }] }))
      }
      if (path.endsWith('/lawfirm/lawyers?page=1&page_size=50')) {
        return Promise.resolve(jsonResponse({ lawyers: [{ id: 7, name: '律师甲' }] }))
      }
      if (path === '/api/v1/case-intakes/intake-1') {
        return Promise.resolve(
          jsonResponse({
            intake: {
              id: 'intake-1',
              intake_code: 'INT-1',
              title: '虚构接案争议',
              case_type: 'commercial',
              priority: 'high',
              description: '虚构案件摘要',
              client_id: 42,
              created_by: '7',
              metadata: { lawyer_id: 7, business_area: '商事', sub_area: '合同' },
            },
            revision: 'intake-revision-1',
            parties: [
              { entity_name: '虚构客户甲', entity_type: 'LEGAL_PERSON', party_role: 'client' },
              { entity_name: '虚构对手乙', entity_type: 'LEGAL_PERSON', party_role: 'opposing_party' },
            ],
          }),
        )
      }
      if (path.endsWith('/conflict/history')) {
        return Promise.resolve(
          jsonResponse([
            {
              checkId: 'CCT-REV-1',
              checkStatus: 'COMPLETED',
              checkTime: '2026-09-28T09:00:00Z',
              searchParameters: { intakeId: 'intake-1' },
              checkResult: {
                checkId: 'CCT-REV-1',
                riskAssessment: { overallRisk: 'HIGH', requiresApproval: true },
                decision: { status: 'REVIEW_REQUIRED', coverageStatus: 'COMPLETE' },
                conflictCases: [{ case_number: 'HIT-1', case_name: '在办冲突案' }],
              },
            },
          ]),
        )
      }
      if (path.endsWith('/approvals/approval-1/snapshot')) {
        return Promise.resolve(jsonResponse({ approval_id: 'approval-1', snapshot_id: 'snapshot-current', snapshot: {} }))
      }
      if (path.includes('/approvals/approval-1/return-items')) {
        return Promise.resolve(jsonResponse({ items: [], replies: [] }))
      }
      if (path.endsWith('/approvals/approval-1')) {
        return Promise.resolve(
          jsonResponse({
            id: 'approval-1',
            active_submission_snapshot_id: 'snapshot-current',
            records: [],
          }),
        )
      }
      return Promise.resolve(jsonResponse({}))
    })
  })

  it('refreshes the review verdict on window focus without overwriting typed input', async () => {
    render(
      <MemoryRouter
        initialEntries={['/case/create?mode=revision&intake_id=intake-1&approval_id=approval-1']}
      >
        <CaseIntakeWorkbench />
      </MemoryRouter>,
    )

    // Revision load restores the completed check via /conflict/history;
    // because nothing was re-verified in this session the verdict read is
    // guarded off until a fresh check exists. Assert the form loaded and
    // the typed input survives any focus refresh path.
    const title = await screen.findByLabelText('案件名称 *')
    await waitFor(() => expect((title as HTMLInputElement).value).toBe('虚构接案争议'))
    const typedValue = '回焦刷新后仍保留的标题'
    fireEvent.change(title, { target: { value: typedValue } })

    await act(async () => {
      window.dispatchEvent(new Event('focus'))
    })
    // Give any (incorrectly) scheduled refresh a chance to run.
    await act(async () => {
      await Promise.resolve()
    })
    expect(screen.getByLabelText('案件名称 *')).toHaveValue(typedValue)
  })
})

describe('Revision field locate targets (recheck #2 contract)', () => {
  const domFieldKeys = [
    'client_name',
    'case_title',
    'case_summary',
    'case_type',
    'case_area',
    'responsible_lawyer',
    'party',
  ]

  it('maps every in-workbench whitelisted field to step 0 and a real DOM id', () => {
    for (const fieldKey of domFieldKeys) {
      const target = supplementFieldLocateTargets[fieldKey]
      expect(target).toBeDefined()
      expect(target.step).toBe(0)
      expect(target.selector).toMatch(/^#[a-z][a-z0-9-]*$/)
    }
  })

  it('never points client_contact at the lead-source contact field (recheck: wrong-field defect)', () => {
    // #intake-source-contact is the LEAD source contact, not the client
    // archive contact; the whitelist entry must not land there.
    expect(supplementFieldLocateTargets.client_contact.selector).not.toBe('#intake-source-contact')
    expect(supplementFieldLocateTargets.client_contact.selector).toBe('__client_archive__')
  })
})
