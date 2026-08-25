/**
 * E2E helpers for the current MVP frontend.
 *
 * The browser workflow is intentionally API-mocked so local QA does not depend
 * on a developer having PostgreSQL, Redis, Elasticsearch, or seeded data running.
 */

import { Page, expect } from '@playwright/test'

type TestUserKey = 'admin' | 'lawyer' | 'assistant' | 'finance' | 'conflictOfficer'

export const TEST_USERS: Record<TestUserKey, {
  email: string
  alias: string
  password: string
  role: string
  realName: string
}> = {
  admin: {
    email: 'demo.admin@example.test',
    alias: 'admin',
    password: 'Demo@2026',
    role: 'admin',
    realName: '示例管理员',
  },
  lawyer: {
    email: 'demo.lawyer@example.test',
    alias: 'lawyer',
    password: 'Demo@2026',
    role: 'lawyer',
    realName: '张律师',
  },
  assistant: {
    email: 'demo.assistant@example.test',
    alias: 'assistant',
    password: 'Demo@2026',
    role: 'assistant',
    realName: '示例助理',
  },
  finance: {
    email: 'demo.finance@example.test',
    alias: 'finance',
    password: 'Demo@2026',
    role: 'finance',
    realName: '示例财务',
  },
  conflictOfficer: {
    email: 'demo.conflict.officer@example.test',
    alias: 'conflict_officer',
    password: 'Demo@2026',
    role: 'conflict_officer',
    realName: '独立冲突核查人',
  },
}

const now = '2026-05-25T02:30:00.000Z'

const caseRows = [
  {
    id: 101,
    case_number: 'DEMO-2026-001',
    title: '红杉资本投资管理咨询合同纠纷案',
    client_name: '上海示例科技有限公司',
    case_type: 'commercial',
    status: 'active',
    priority: 'high',
    lawyer_name: '张律师',
    updated_at: now,
  },
  {
    id: 103,
    case_number: 'CASE-20260513173242',
    title: '待处理冲突复核测试案件',
    client_name: '上海示例科技有限公司',
    case_type: 'commercial',
    status: 'pending',
    priority: 'medium',
    lawyer_name: '张律师',
    updated_at: now,
  },
  {
    id: 102,
    case_number: 'DEMO-2026-002',
    title: '蓝海公司股权转让争议',
    client_name: '蓝海企业管理有限公司',
    case_type: 'ma',
    status: 'submitted',
    priority: 'medium',
    lawyer_name: '李律师',
    updated_at: now,
  },
]

const riskQueue = [
  {
    id: 301,
    case_id: 103,
    case_number: 'CASE-20260513173242',
    title: '待处理冲突复核测试案件',
    client_name: '上海示例科技有限公司',
    case_type: 'commercial',
    status: 'COMPLETED',
    risk_level: 'MEDIUM',
    has_conflict: true,
    matched_subject: '上海示例科技有限公司',
    matched_type: '名称相似待核实',
    evidence_summary: '名称候选只用于人工核实，不自动认定利益冲突。',
    owner: 1,
    duration: 96,
    created_at: now,
    updated_at: now,
    check_time: now,
    search_parameters: {
      searchDepth: 'STANDARD',
      searchYears: 0,
      query: '示例科技',
      matchedClientName: '上海示例科技有限公司',
      matchMode: 'NAME_CANDIDATE',
      automaticConclusion: false,
    },
    check_result: {
      riskAssessment: {
        overallRisk: 'MEDIUM',
        riskScore: 55,
        riskReason: '客户与既有案件存在关联主体',
        requiresApproval: true,
        matchEvidence: {
          queryName: '示例科技',
          candidateName: '上海示例科技有限公司',
          matchType: 'NAME_CANDIDATE',
          algorithm: 'NORMALIZED_CONTAINS',
          automaticConclusion: false,
          partyRole: '对方当事人',
          ruleId: 'CONFLICT-NAME-CANDIDATE-001',
        },
      },
      checkStatistics: {
        totalCasesChecked: 2,
        relatedPartiesChecked: 2,
      },
    },
    conflict_cases: [
      {
        id: 402,
        case_no: 'DEMO-2025-188',
        case_name: '关联主体历史委托',
        conflict_type: '关联冲突',
        risk_level: 'MEDIUM',
        case_status: 'active',
        description: '客户关联方与历史委托存在交集',
      },
    ],
  },
  {
    id: 302,
    title: '红杉资本投资管理咨询合同纠纷案',
    client_name: '上海示例科技有限公司',
    case_type: 'commercial',
    status: 'COMPLETED',
    risk_level: 'HIGH',
    has_conflict: true,
    owner: 1,
    duration: 128,
    created_at: now,
    updated_at: now,
    check_time: now,
    search_parameters: { searchDepth: 'STANDARD', searchYears: 0 },
    check_result: {
      riskAssessment: {
        overallRisk: 'HIGH',
        riskScore: 86,
        riskReason: '对方当事人与既有客户存在业务关联',
        requiresApproval: true,
      },
      checkStatistics: {
        totalCasesChecked: 2,
        relatedPartiesChecked: 3,
      },
    },
    conflict_cases: [
      {
        id: 401,
        case_no: 'DEMO-2025-099',
        case_name: '历史顾问合同',
        conflict_type: '既有客户冲突',
        risk_level: 'HIGH',
        case_status: 'active',
        description: '同一实控人关联企业',
      },
    ],
  },
]

const commandCenterPayload = {
  generated_at: now,
  summary: {
    active_cases: 2,
    clients: 2,
    pending_approvals: 1,
  },
  workflow: {
    intake: 1,
    conflict: 1,
    approval: 1,
  },
  case_rows: caseRows,
  risk_queue: riskQueue,
  inbox_items: [
    {
      id: 501,
      title: '审批红杉资本新接案',
      content: '红杉资本投资管理咨询合同纠纷案',
      type: 'approval',
      source_type: 'approval',
      priority: 'high',
      due_at: now,
    },
  ],
}

const approvalWorkbenchPayload = {
  stats: {
    pending: 1,
    waiver_review: 0,
    needs_revision: 0,
  },
  queues: [
    { key: 'conflict', count: 1 },
  ],
  items: [
    {
      id: 701,
      request_number: 'AP-2026-001',
      title: '冲突审查审批 - 红杉资本投资管理咨询合同纠纷案',
      priority: 'high',
      current_stage: '合规复核',
      current_approver_id: '1',
      current_approver_name: '示例管理员',
      applicant_id: '2',
      applicant_name: '张律师',
      status: 'pending',
      content: '请确认是否可继续承办。',
      updated_at: now,
      created_at: now,
    },
  ],
}

const clientsPayload = {
  list: [
    { id: 1, name: '上海示例科技有限公司' },
    { id: 2, name: '蓝海企业管理有限公司' },
  ],
  total: 2,
}

const lawyersPayload = {
  list: [
    { id: 1, name: '张律师', department: '争议解决部', position: '合伙人' },
    { id: 2, name: '李律师', department: '公司业务部', position: '律师' },
  ],
  total: 2,
}

function ok(data: unknown) {
  return {
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify({ success: true, data }),
  }
}

function createTestToken(role: string) {
  const header = Buffer.from(JSON.stringify({ alg: 'none', typ: 'JWT' })).toString('base64url')
  const payload = Buffer.from(JSON.stringify({
    sub: `e2e-${role}`,
    role,
    exp: Math.floor(Date.now() / 1000) + 60 * 60,
  })).toString('base64url')
  return `${header}.${payload}.`
}

function userPayload(key: TestUserKey) {
  const user = TEST_USERS[key]
  return {
    id: key === 'admin' ? 1 : key === 'lawyer' ? 2 : key === 'finance' ? 4 : 3,
    email: user.email,
    name: user.realName,
    real_name: user.realName,
    username: user.alias,
    role: user.role,
    status: 'active',
    created_at: now,
  }
}

function matchLoginUser(emailOrAlias: string): TestUserKey | null {
  const normalized = emailOrAlias.trim().toLowerCase()
  return (Object.keys(TEST_USERS) as TestUserKey[]).find((key) => {
    const user = TEST_USERS[key]
    return normalized === user.alias || normalized === user.email
  }) ?? null
}

export async function installApiMocks(page: Page) {
  let pendingSubjectRegistration = false
  await page.route('**/api/v1/**', async (route) => {
    const request = route.request()
    const url = new URL(request.url())
    const path = url.pathname.replace('/api/v1', '')

    if (path === '/auth/login' && request.method() === 'POST') {
      const body = request.postDataJSON() as { email?: string; password?: string }
      const userKey = body.email ? matchLoginUser(body.email) : null
      if (!userKey || body.password !== TEST_USERS[userKey].password) {
        await route.fulfill({
          status: 401,
          contentType: 'application/json',
          body: JSON.stringify({
            success: false,
            error: { message: '账号或密码错误' },
          }),
        })
        return
      }
      await route.fulfill(ok({ token: createTestToken(TEST_USERS[userKey].role), user: userPayload(userKey) }))
      return
    }

    if (path === '/admin/current-user/roles') {
      await route.fulfill(ok([{ code: 'lawyer' }]))
      return
    }

    if (path === '/admin/current-user/permissions') {
      await route.fulfill(ok([]))
      return
    }

    if (path === '/dashboard/command-center') {
      await route.fulfill(ok(commandCenterPayload))
      return
    }

    if (path === '/finance/overview') {
      await route.fulfill(ok({ warning_amount: 120000 }))
      return
    }

    if (path === '/clients') {
      await route.fulfill(ok(clientsPayload))
      return
    }

    if (path.match(/^\/clients\/\d+$/) && request.method() === 'PUT') {
      await route.fulfill(ok({
        id: 1,
        name: '上海示例科技有限公司',
        contact_person: '王总',
        contact_phone: '021-55550000',
      }))
      return
    }

    if (path.match(/^\/clients\/\d+\/master-profile$/)) {
      await route.fulfill(ok({
        client: {
          id: 1,
          version: 3,
          name: '上海示例科技有限公司',
          status: 'active',
          type: '企业',
          identity_type: 'SOCIAL_CREDIT_CODE',
          identity_status: '已登记（受保护）',
          aliases: ['示例科技'],
          industry: '科技服务',
          email: 'legal@example.com',
          phone: '021-55550000',
          contact_person: '王总',
          address: '上海市浦东新区',
          source: '现有客户介绍',
          created_at: now,
          updated_at: now,
        },
        related_parties: [],
        conflict_history: riskQueue,
        completeness: { score: 92 },
      }))
      return
    }

    if (path === '/documents/upload' && request.method() === 'POST') {
      await route.fulfill(ok({
        id: 901,
        name: 'client-note.txt',
        filename: 'client-note.txt',
        entity_type: 'client',
        entity_id: 1,
      }))
      return
    }

    if (path === '/admin/access-center') {
      await route.fulfill(ok({
        summary: { users: 2, active_users: 2, roles: 2, disabled_users: 0, pending_changes: 0 },
        users: [
          { id: 1, name: '示例管理员', email: TEST_USERS.admin.email, status: 'active', role: 'admin' },
          { id: 2, name: '张律师', email: TEST_USERS.lawyer.email, status: 'active', role: 'lawyer' },
        ],
        roles: [
          { key: 'admin', label: '管理员', count: 1 },
          { key: 'lawyer', label: '律师', count: 1 },
        ],
        permission_changes: [],
        audit_events: [],
      }))
      return
    }

    if (path === '/lawfirm/lawyers') {
      await route.fulfill(ok(lawyersPayload))
      return
    }

    if (path === '/case-intakes' && request.method() === 'POST') {
      await route.fulfill(ok({ id: 801, intake_code: 'IN-2026-001' }))
      return
    }

    if (path.match(/^\/case-intakes\/\d+$/) && request.method() === 'PUT') {
      await route.fulfill(ok({ id: 801, intake_code: 'IN-2026-001', status: 'conflict_ready' }))
      return
    }

    if (path === '/case-intakes/801/facts-confirmation' && request.method() === 'POST') {
      await route.fulfill(ok({ id: 801, status: 'lawyer_facts_confirmed' }))
      return
    }

    if (path === '/case-intakes/801/conflict-check' && request.method() === 'POST') {
      await route.fulfill(ok({
        taskId: 'CHK-2026-ASYNC-001',
        checkId: 'CHK-2026-ASYNC-001',
        status: 'COMPLETED',
        result: {
          checkId: 'CHK-2026-ASYNC-001',
          hasConflict: false,
          conflictCases: [],
          riskAssessment: { overallRisk: 'LOW', riskScore: 18, riskReason: '未发现直接冲突' },
          decision: { status: 'CLEAR', recommendation: '未发现可识别的冲突线索，可继续进入人工确认环节。' },
          coverage_status: 'COMPLETE',
          checkStatistics: { totalCasesChecked: 2, relatedPartiesChecked: 1 },
        },
      }))
      return
    }

    if (path === '/inbox') {
      await route.fulfill(ok({ items: commandCenterPayload.inbox_items, pagination: { total: commandCenterPayload.inbox_items.length, page: 1, page_size: 20 } }))
      return
    }

    if (path === '/inbox/stats') {
      await route.fulfill(ok({ total: 1, unread: 1, pending: 1, completed: 0, overdue: 0 }))
      return
    }

    if (path === '/conflict/check' && request.method() === 'POST') {
      await route.fulfill(ok({
        checkId: 'CHK-2026-001',
        riskAssessment: {
          overallRisk: 'LOW',
          riskScore: 18,
          riskReason: '未发现直接冲突',
        },
        checkStatistics: {
          totalCasesChecked: 2,
          relatedPartiesChecked: 1,
        },
      }))
      return
    }

    if (path === '/conflict/subject-entity-registrations') {
      await route.fulfill(ok([{
        revision_id: 'revision-new-subject-101',
        case_id: 101,
        case_number: 'DEMO-2026-001',
        case_title: '红杉资本投资管理咨询合同纠纷案',
        change_type: 'ADD_OPPOSING_PARTY',
        candidate_name: '虚构启明精密制造有限公司',
        entity_type: 'LEGAL_PERSON',
        identity_type: 'SOCIAL_CREDIT_CODE',
        identity_hint: '**************A101',
        requested_by: 2,
        requested_by_name: '张律师',
        reason: '法院通知追加该公司为共同被告',
      }]))
      return
    }

    if (path === '/conflict-v2/entities/search') {
      await route.fulfill(ok([
        { id: 901, name: '虚构启明精密制造有限公司', entity_type: 'LEGAL_PERSON' },
      ]))
      return
    }

    if (path === '/conflict/tasks' && request.method() === 'POST') {
      await route.fulfill(ok({
        taskId: 'CHK-2026-ASYNC-001',
        checkId: 'CHK-2026-ASYNC-001',
        status: 'QUEUED',
        recommendedPollingInterval: 0.1,
      }))
      return
    }

    if (path === '/conflict/tasks/CHK-2026-ASYNC-001/result') {
      await route.fulfill(ok({
        task: { taskId: 'CHK-2026-ASYNC-001', checkId: 'CHK-2026-ASYNC-001', status: 'COMPLETED' },
        result: {
          checkId: 'CHK-2026-ASYNC-001',
          hasConflict: false,
          conflictCases: [],
          riskAssessment: { overallRisk: 'LOW', riskScore: 18, riskReason: '未发现直接冲突' },
          decision: { status: 'CLEAR', recommendation: '未发现可识别的冲突线索，可继续进入人工确认环节。' },
          coverage_status: 'COMPLETE',
          checkStatistics: { totalCasesChecked: 2, relatedPartiesChecked: 1 },
        },
      }))
      return
    }

    if (path === '/integration/approvals/with-conflict' && request.method() === 'POST') {
      await route.fulfill(ok({ approval_id: 701, request_number: 'AP-2026-001' }))
      return
    }

    if (path === '/approvals/workbench') {
      await route.fulfill(ok(approvalWorkbenchPayload))
      return
    }

    if (path.match(/^\/approvals\/\d+$/)) {
      await route.fulfill(ok(approvalWorkbenchPayload.items[0]))
      return
    }

    if (path.match(/^\/approvals\/\d+\/snapshot$/)) {
      await route.fulfill(ok({ approval: approvalWorkbenchPayload.items[0], conflict: riskQueue[0] }))
      return
    }

    if (path.match(/^\/integration\/approvals\/\d+\/status$/)) {
      await route.fulfill(ok({
        case_creation: { case_id: 101, case_number: 'DEMO-2026-001', status: 'created' },
        status: 'approved',
      }))
      return
    }

    if (path.match(/^\/integration\/approvals\/\d+\/decision$/)) {
      await route.fulfill(ok({ status: 'approved', case_id: 101, case_number: 'DEMO-2026-001' }))
      return
    }

    if (path.match(/^\/conflict\/tasks\/\d+\/approval$/)) {
      await route.fulfill(ok({ approval_id: 701, request_number: 'AP-2026-001' }))
      return
    }

    if (path.match(/^\/cases\/\d+\/subject-parties$/)) {
      await route.fulfill(ok([]))
      return
    }

    if (path.match(/^\/cases\/\d+\/subject-entities$/)) {
      await route.fulfill(ok([]))
      return
    }

    if (path.match(/^\/cases\/\d+\/subject-entity-registrations$/) && request.method() === 'POST') {
      pendingSubjectRegistration = true
      await route.fulfill(ok({
        revision: { id: 'revision-new-subject-101', status: 'ENTITY_REGISTRATION_PENDING' },
        case_subject_state: 'RECHECK_REQUIRED',
        action_gate_message: '新主体等待核查岗确认，受控动作已暂停',
      }))
      return
    }

    if (path.match(/^\/cases\/\d+\/subject-revisions\/[^/]+$/) && request.method() === 'GET') {
      await route.fulfill(ok({
        revision_id: 'revision-new-subject-101',
        status: pendingSubjectRegistration ? 'ENTITY_REGISTRATION_PENDING' : 'CHANGE_PROPOSED',
        change_type: 'ADD_OPPOSING_PARTY',
        candidate_name: pendingSubjectRegistration ? '虚构启明精密制造有限公司' : undefined,
        identity_type: pendingSubjectRegistration ? 'SOCIAL_CREDIT_CODE' : undefined,
        identity_hint: pendingSubjectRegistration ? '**************A101' : undefined,
      }))
      return
    }

    if (path.match(/^\/cases\/\d+\/subject-revisions\/[^/]+\/entity-registration-review$/) && request.method() === 'POST') {
      await route.fulfill(ok({
        revision: { id: 'revision-new-subject-101', status: 'CHANGE_PROPOSED' },
        case_subject_state: 'RECHECK_REQUIRED',
        action_gate_message: '主体登记已确认，等待申请律师运行冲突重检',
      }))
      return
    }

    if (path.match(/^\/cases\/\d+$/)) {
      const caseId = Number(path.split('/').pop())
      const row = caseRows.find((item) => item.id === caseId) || caseRows[0]
      await route.fulfill(ok({
        ...row,
        description: 'E2E 案件详情',
        created_at: now,
        subject_version: 1,
        subject_state: pendingSubjectRegistration ? 'RECHECK_REQUIRED' : 'EFFECTIVE',
        pending_subject_revision_id: pendingSubjectRegistration ? 'revision-new-subject-101' : '',
        conflict_coverage_status: 'COMPLETE',
      }))
      return
    }

    if (path === '/notifications' || path === '/notifications/stats') {
      await route.fulfill(ok(path === '/notifications/stats' ? { unread: 0, total: 0 } : []))
      return
    }

    if (path.startsWith('/trust/')) {
      await route.fulfill(ok({ list: [], total: 0, summary: {} }))
      return
    }

    await route.fulfill(ok({}))
  })
}

export async function waitForPageLoad(page: Page) {
  await page.waitForLoadState('domcontentloaded')
}

export async function login(page: Page, userKey: TestUserKey = 'lawyer') {
  await installApiMocks(page)
  const user = TEST_USERS[userKey]
  await page.goto('/login')
  await waitForPageLoad(page)
  await page.getByPlaceholder('账号或邮箱').fill(user.email)
  await page.getByPlaceholder('密码').fill(user.password)
  await page.locator('button[type="submit"]').click()
  await expect(page).toHaveURL(/\/dashboard$/, { timeout: 10000 })
}

export async function seedAuthenticatedUser(page: Page, userKey: TestUserKey = 'lawyer') {
  await installApiMocks(page)
  await page.goto('/login')
  const user = TEST_USERS[userKey]
  await page.evaluate((payload) => {
    localStorage.setItem('auth_token', payload.token)
    localStorage.setItem('law_oa_user_info', JSON.stringify(payload.user))
    localStorage.setItem('law_oa_roles', JSON.stringify([{ code: payload.user.role }]))
    localStorage.setItem('law_oa_permissions', JSON.stringify([]))
  }, {
    token: createTestToken(user.role),
    user: {
      id: String(userKey === 'admin' ? 1 : userKey === 'lawyer' ? 2 : userKey === 'finance' ? 4 : userKey === 'conflictOfficer' ? 5 : 3),
      username: user.email,
      email: user.email,
      realName: user.realName,
      roles: [user.role],
      permissions: [],
      isActive: true,
      createdAt: now,
    },
  })
}

export async function logout(page: Page) {
  await page.locator('.user-menu').click()
  await page.getByText('退出登录').click()
  await expect(page).toHaveURL(/\/login$/)
}

export async function isLoggedIn(page: Page): Promise<boolean> {
  return page.locator('.user-menu').isVisible()
}

export async function waitForAppShell(page: Page) {
  await expect(page.locator('.app-header')).toBeVisible()
  await expect(page.locator('.app-sidebar')).toBeVisible()
}

export async function waitForNativeTable(page: Page) {
  await expect(page.getByRole('table')).toBeVisible()
}
