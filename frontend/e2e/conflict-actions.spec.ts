import { expect, Page, test } from '@playwright/test'
import { seedAuthenticatedUser, waitForAppShell, waitForPageLoad } from './utils/test-helpers'

type Scenario = 'WAIVED' | 'CLEAR' | 'BLOCKED' | 'STALE' | 'COMPLETE_EVIDENCE' | 'PRINT_ESCAPED'

const scenarioItems: Record<Scenario, Record<string, any>> = {
  WAIVED: {
    id: 'qa-waived',
    title: '已批准豁免测试案件',
    client_name: '云帆科技有限公司',
    status: 'COMPLETED',
    risk_level: 'HIGH',
    has_conflict: true,
    check_result: {
      decision: {
        status: 'WAIVED',
        coverageStatus: 'COMPLETE',
        coverageNotice: '已完成全所客户与案件检索。',
      },
      waiver: {
        id: 'waiver-approved',
        status: 'APPROVED',
        approved_conditions: ['建立信息隔离墙', '每月由合规负责人复核'],
        expiry_date: '2026-12-31T00:00:00.000Z',
      },
      riskAssessment: {
        overallRisk: 'HIGH',
        riskScore: 82,
        requiresApproval: true,
        primaryEvidence: {
          requestedParty: '云帆科技有限公司',
          ruleCode: 'CONFLICT-DIRECT-001',
          matchType: 'EXACT',
          sourceCaseNumber: 'CASE-2025-0088',
        },
      },
    },
  },
  CLEAR: {
    id: 'qa-clear',
    title: '无冲突立案测试案件',
    client_name: '清源咨询有限公司',
    status: 'COMPLETED',
    risk_level: 'LOW',
    has_conflict: false,
    check_result: {
      decision: { status: 'CLEAR', coverageStatus: 'COMPLETE' },
      conflictCases: [],
      riskAssessment: {
        overallRisk: 'LOW',
        riskScore: 0,
        requiresApproval: false,
        primaryEvidence: {
          requestedParty: '清源咨询有限公司',
          queryName: '清源咨询有限公司',
        },
      },
    },
  },
  BLOCKED: {
    id: 'qa-blocked',
    title: '历史案件标题不得作为命中主体',
    client_name: '当前申请客户',
    status: 'COMPLETED',
    risk_level: 'CRITICAL',
    has_conflict: true,
    check_result: {
      decision: { status: 'BLOCKED', coverageStatus: 'COMPLETE' },
      review: { id: 'review-confirmed-blocked', decision: 'confirmed_conflict', reviewerName: '独立复核人' },
      riskAssessment: { overallRisk: 'CRITICAL', riskScore: 96, requiresApproval: true },
    },
    conflict_cases: [{
      id: 'source-case',
      case_no: 'CASE-2024-0066',
      case_name: '受限历史案件',
      risk_level: 'CRITICAL',
      evidence: [{
        requestedParty: '真实主命中实体',
        ruleCode: 'CONFLICT-OPPOSING-001',
        matchType: 'EXACT_NORMALIZED',
        sourceCaseNumber: 'CASE-2024-0066',
        summary: '当前对方是本所既有客户。',
      }],
    }],
  },
  STALE: {
    id: 'qa-stale',
    title: '检测结果过期测试案件',
    client_name: '时效测试客户',
    status: 'STALE',
    risk_level: 'CRITICAL',
    has_conflict: true,
    check_result: {
      stale: true,
      decision: { status: 'BLOCKED', coverageStatus: 'COMPLETE' },
      riskAssessment: { overallRisk: 'CRITICAL', riskScore: 99, requiresApproval: true },
    },
  },
  COMPLETE_EVIDENCE: {
    id: 'CCT_373e_complete_evidence',
    title: '完整证据回填测试案件',
    client_name: '青海云岭建设有限公司',
    matched_type: 'EXACT',
    status: 'COMPLETED',
    risk_level: 'CRITICAL',
    has_conflict: true,
    search_parameters: {
      query: '青海云岭建设有限公司',
      searchDepth: 'STANDARD',
      searchYears: 0,
    },
    check_result: {
      decision: { status: 'BLOCKED' },
      riskAssessment: { overallRisk: 'CRITICAL', requiresApproval: true },
      conflictCases: [{
        id: 'historical-source-case',
        case_no: 'CASE-2025-0373',
        case_name: '上海示例科技常年法律顾问案',
        evidence: [{
          requestedParty: '上海示例科技有限公司',
          historicalClientName: '上海示例科技有限公司（客户主档）',
          ruleCode: 'DIRECT_ADVERSE_CURRENT_CLIENT',
          matchType: 'EXACT',
          subjectRole: 'CLIENT',
          sourceCaseNumber: 'CASE-2025-0373',
          riskScore: 94,
          summary: '当前被检索对方与本所现有客户主档完全一致。',
        }],
      }],
    },
  },
  PRINT_ESCAPED: {
    id: 'CCT_PRINT_ESCAPED',
    title: '打印报告 <案件>',
    client_name: '客户 & 归档组',
    matched_subject: '主体 <待核验>',
    evidence_summary: '<img src=x onerror=alert(1)> & 未转义内容',
    status: 'COMPLETED',
    risk_level: 'MEDIUM',
    has_conflict: true,
    check_result: {
      decision: { status: 'REVIEW_REQUIRED' },
      riskAssessment: { overallRisk: 'MEDIUM', riskScore: 54, requiresApproval: false },
    },
  },
}

async function openConflictScenario(page: Page, scenario: Scenario, user: 'lawyer' | 'conflictOfficer' = 'lawyer') {
  await seedAuthenticatedUser(page, user)
  const item = scenarioItems[scenario]
  await page.route('**/api/v1/dashboard/command-center**', async (route) => {
    await route.fulfill({
      contentType: 'application/json',
      body: JSON.stringify({
        success: true,
        data: {
          generated_at: '2026-07-13T08:00:00.000Z',
          summary: { active_cases: 1, pending_approvals: 0 },
          risk_queue: [item],
        },
      }),
    })
  })
  await page.goto(`/conflict?task_id=${item.id}`)
  await waitForPageLoad(page)
  await waitForAppShell(page)
  await expect(page.getByRole('dialog', { name: '冲突检测详情' })).toBeVisible()
}

async function fillCaseIntakeBasics(page: Page) {
  await page.getByRole('textbox', { name: '案件名称 *' }).fill('冲突门禁 E2E 案件')
  await page.getByRole('combobox', { name: '案件类型 *' }).click()
  await page.getByTitle('商事诉讼').click()
  await page.getByRole('combobox', { name: '业务领域 *' }).click()
  await page.getByTitle('公司与并购').click()
  await page.getByRole('combobox', { name: '子领域 *' }).click()
  await page.getByTitle('投资与融资').click()
  await page.getByRole('combobox', { name: '我方当事人（客户）' }).click()
  await page.getByTitle('上海示例科技有限公司').click()
  await page.getByPlaceholder('输入对方当事人名称').fill('冲突测试对方')
  await page.getByRole('textbox', { name: '案情摘要 *' }).fill('验证冲突结果对立案提交的前置门禁。')
  await page.getByRole('combobox', { name: '负责律师预览' }).click()
  await page.getByTitle('张律师 · 争议解决部').click()
}

test.describe('冲突决策单一状态', () => {
  test('WAIVED 只允许按批准条件继续，不出现审批或再次豁免', async ({ page }) => {
    await openConflictScenario(page, 'WAIVED')
    const dialog = page.getByRole('dialog', { name: '冲突检测详情' })
    const decisionCard = dialog.locator('section.ng-panel', { hasText: '接案决策' })

    await expect(decisionCard.getByText('按批准条件继续', { exact: true })).toBeVisible()
    await expect(dialog.getByText(/建立信息隔离墙.*每月由合规负责人复核/)).toBeVisible()
    const expiry = dialog.locator('article', { hasText: '有效期限' })
    await expect(expiry).toContainText(/2026|2027/)
    await expect(expiry).not.toContainText('未设置')
    await expect(dialog.getByRole('button', { name: '申请豁免评估' })).toHaveCount(0)
    await expect(dialog.getByRole('button', { name: '发起冲突审批' })).toHaveCount(0)
  })

  test('CLEAR/LOW/零命中明确可提交立案审批且无冲突处置入口', async ({ page }) => {
    await openConflictScenario(page, 'CLEAR')
    const dialog = page.getByRole('dialog', { name: '冲突检测详情' })
    const decisionCard = dialog.locator('section.ng-panel', { hasText: '接案决策' })

    await expect(decisionCard.getByText('可提交立案审批', { exact: true })).toBeVisible()
    await expect(dialog.getByText(/不得提交/)).toHaveCount(0)
    await expect(dialog.getByRole('button', { name: '申请豁免评估' })).toHaveCount(0)
    await expect(dialog.getByRole('button', { name: '发起冲突审批' })).toHaveCount(0)
  })

  test('人工 confirmed_conflict 终态不重复发起复核且只保留一个豁免入口', async ({ page }) => {
    await openConflictScenario(page, 'BLOCKED')
    const dialog = page.getByRole('dialog', { name: '冲突检测详情' })
    const decisionCard = dialog.locator('section.ng-panel', { hasText: '接案决策' })

    await expect(decisionCard.getByText('已暂停接案', { exact: true })).toBeVisible()
    await expect(dialog.locator('article', { hasText: '命中历史主体' }).getByText('真实主命中实体', { exact: true })).toBeVisible()
    await expect(dialog.getByText('历史案件标题不得作为命中主体', { exact: true })).toHaveCount(1)
    const waiverAction = dialog.getByRole('button', { name: '申请豁免评估' })
    await expect(waiverAction).toHaveCount(1)
    await expect(waiverAction).toBeEnabled()
    await expect(dialog.getByRole('button', { name: '发起冲突审批' })).toHaveCount(0)
  })

  test('STALE 隐藏旧严重处置，只显示结果过期', async ({ page }) => {
    await openConflictScenario(page, 'STALE')
    const dialog = page.getByRole('dialog', { name: '冲突检测详情' })
    const decisionCard = dialog.locator('section.ng-panel', { hasText: '接案决策' })

    await expect(decisionCard.getByText('冲突检测结果已过期', { exact: true })).toBeVisible()
    await expect(dialog.getByText('已暂停接案', { exact: true })).toHaveCount(0)
    await expect(dialog.getByRole('button', { name: '申请豁免评估' })).toHaveCount(0)
    await expect(dialog.getByRole('button', { name: '发起冲突审批' })).toHaveCount(0)
  })

  test('完整 evidence 回填风险汇总且区分冲突类型与匹配方式', async ({ page }) => {
    await openConflictScenario(page, 'COMPLETE_EVIDENCE')
    const dialog = page.getByRole('dialog', { name: '冲突检测详情' })
    const conflictType = dialog.locator('article', { hasText: '主冲突类型' })
    await expect(conflictType).toContainText('当前对方为本所现有客户')
    await expect(conflictType).not.toContainText('EXACT')

    const riskSummary = dialog.locator('section.ng-panel').filter({ hasText: '风险评估结果' })
    const row = (label: string) => riskSummary.locator('tr', { hasText: label })
    await expect(row('风险评分')).toContainText('94 / 100')
    await expect(row('检索主体')).toContainText('上海示例科技有限公司')
    await expect(row('检索主体')).not.toContainText('青海云岭建设有限公司')
    await expect(row('匹配主体')).toContainText('上海示例科技有限公司（客户主档）')
    await expect(row('匹配方式')).toContainText('规范化名称完全一致')
    await expect(row('比对方法')).toContainText('规范化名称比对')
    await expect(row('主体角色')).toContainText('CLIENT')
    await expect(row('规则编号')).toContainText('DIRECT_ADVERSE_CURRENT_CLIENT')
    await expect(row('来源案件')).toContainText('受限历史事项')
    await expect(row('来源案件')).not.toContainText('CASE-2025-0373')
    for (const label of ['风险评分', '匹配方式', '比对方法', '主体角色', '规则编号', '来源案件']) {
      await expect(row(label)).not.toContainText('未提供')
    }

    await dialog.locator('.ant-modal-footer button').first().click()
    await expect(dialog).toBeHidden()
    const listRow = page.locator('.conflict-list-table tbody tr').filter({ hasText: '完整证据回填测试案件' })
    await expect(listRow.locator('td').nth(3)).toHaveText('规范化名称完全一致')
    await expect(listRow.locator('td').nth(7)).toContainText('当前被检索对方与本所现有客户主档完全一致。')
  })

  test('范围受限记录明确归入待人工复核，不显示为可忽略提示', async ({ page }) => {
    await openConflictScenario(page, 'PRINT_ESCAPED')
    const dialog = page.getByRole('dialog', { name: '冲突检测详情' })
    await dialog.locator('.ant-modal-footer button').first().click()
    await expect(dialog).toBeHidden()

    const overview = page.locator('.batch-info-grid.compact').filter({ hasText: '待人工复核' })
    await expect(overview.locator('p').filter({ hasText: '待人工复核' })).toContainText('1 条')
    await expect(page.getByText(/当前账号可见的冲突任务：存在 1 条待人工复核记录/)).toBeVisible()
    const complianceWarning = page
      .locator('.batch-conflict-side p')
      .filter({ hasText: '未发现高风险' })
    await expect(complianceWarning).toBeVisible()
    await expect(complianceWarning).toContainText('待人工复核记录')
    await expect(complianceWarning).toContainText('不得继续接案')

    await page.getByRole('button', { name: '筛选待人工复核' }).click()
    await expect(
      page.locator('.conflict-list-table tbody tr').filter({ hasText: '打印报告 <案件>' }),
    ).toHaveCount(1)
  })

  test('普通律师可为机器自动 BLOCKED 发起首次复核并进入后端复用的审批', async ({ page }) => {
    let approvalRequests = 0
    await openConflictScenario(page, 'COMPLETE_EVIDENCE')
    await page.route('**/api/v1/conflict/tasks/CCT_373e_complete_evidence/approval', async (route) => {
      approvalRequests += 1
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify({ success: true, data: { approval_id: 770, request_number: 'AP-EXISTING-770', reused: true } }),
      })
    })
    const dialog = page.getByRole('dialog', { name: '冲突检测详情' })
    const decisionCard = dialog.locator('section.ng-panel', { hasText: '接案决策' })
    await expect(decisionCard.getByText('机器检测已自动阻断，等待独立复核', { exact: true })).toBeVisible()
    await expect(dialog.getByRole('button', { name: '申请豁免评估' })).toHaveCount(0)
    const reviewAction = dialog.getByRole('button', { name: '发起冲突审批' })
    await expect(reviewAction).toHaveCount(1)
    await reviewAction.click()

    await expect(page.locator('.ant-message')).toContainText('已进入现有冲突审批：AP-EXISTING-770')
    await expect(page).toHaveURL(/\/approval\/770$/)
    expect(approvalRequests).toBe(1)
  })

  test('首次创建冲突审批显示新建提示并保持审批导航', async ({ page }) => {
    await openConflictScenario(page, 'COMPLETE_EVIDENCE')
    await page.route('**/api/v1/conflict/tasks/CCT_373e_complete_evidence/approval', async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify({ success: true, data: { approval_id: 771, request_number: 'AP-NEW-771', reused: false } }),
      })
    })
    await page.getByRole('dialog', { name: '冲突检测详情' }).getByRole('button', { name: '发起冲突审批' }).click()
    await expect(page.locator('.ant-message')).toContainText('已创建冲突审批：AP-NEW-771')
    await expect(page).toHaveURL(/\/approval\/771$/)
  })

  test('打印检测报告包含归档字段并转义动态文本', async ({ page }) => {
    await openConflictScenario(page, 'PRINT_ESCAPED')
    await page.evaluate(() => {
      ;(window as any).__printReportHtml = ''
      ;(window as any).__printCalled = false
      ;(window as any).open = () => {
        const reportWindow = {
          document: {
            open: () => undefined,
            write: (html: string) => { ;(window as any).__printReportHtml = html },
            close: () => undefined,
          },
          focus: () => undefined,
          print: () => { ;(window as any).__printCalled = true },
        }
        return reportWindow
      }
    })
    const dialog = page.getByRole('dialog', { name: '冲突检测详情' })
    await dialog.locator('.ant-modal-footer button').first().click()
    await expect(dialog).toBeHidden()
    await page.getByRole('button', { name: '打印/导出 PDF' }).click()

    const reportHtml = await page.evaluate(() => (window as any).__printReportHtml as string)
    expect(reportHtml).toContain('利益冲突检测报告')
    expect(reportHtml).toContain('生成时间')
    expect(reportHtml).toContain('检测范围')
    expect(reportHtml).toContain('总数')
    expect(reportHtml).toContain('打印报告 &lt;案件&gt;')
    expect(reportHtml).toContain('&lt;img src=x onerror=alert(1)&gt; &amp; 未转义内容')
    expect(reportHtml).not.toContain('<img src=x onerror=alert(1)>')
    expect(await page.evaluate(() => Boolean((window as any).__printCalled))).toBe(true)
  })

  test('打印窗口被浏览器拦截时显示明确反馈', async ({ page }) => {
    await openConflictScenario(page, 'PRINT_ESCAPED')
    await page.evaluate(() => {
      ;(window as any).open = () => null
    })
    const dialog = page.getByRole('dialog', { name: '冲突检测详情' })
    await dialog.locator('.ant-modal-footer button').first().click()
    await expect(dialog).toBeHidden()
    await page.getByRole('button', { name: '打印/导出 PDF' }).click()
    await expect(page.locator('.ant-message')).toContainText('无法打开打印窗口，请允许浏览器弹出窗口后重试')
  })

  test('CLEAR/LOW/零命中在清单和打印报告中不把客户显示为命中主体', async ({ page }) => {
    await openConflictScenario(page, 'CLEAR')
    const dialog = page.getByRole('dialog', { name: '冲突检测详情' })
    await dialog.locator('.ant-modal-footer button').first().click()
    await expect(dialog).toBeHidden()

    const row = page.locator('.conflict-list-table tbody tr').filter({ hasText: '无冲突立案测试案件' })
    await expect(row.locator('td').nth(2)).toHaveText('无命中主体')
    await expect(row.locator('td').nth(2)).not.toContainText('清源咨询有限公司')
    await expect(row.locator('td').nth(3)).toHaveText('无命中')
    await expect(row.locator('td').nth(7)).toHaveText('未发现可识别冲突')

    await page.evaluate(() => {
      ;(window as any).__printReportHtml = ''
      ;(window as any).open = () => ({
        document: {
          open: () => undefined,
          write: (html: string) => { ;(window as any).__printReportHtml = html },
          close: () => undefined,
        },
        focus: () => undefined,
        print: () => undefined,
      })
    })
    await page.getByRole('button', { name: '打印/导出 PDF' }).click()
    const reportHtml = await page.evaluate(() => (window as any).__printReportHtml as string)
    expect(reportHtml).toContain('<td>无冲突立案测试案件</td>')
    expect(reportHtml).toContain('<td>清源咨询有限公司</td>\n        <td>无命中主体</td>')
    expect(reportHtml).toContain('<td>未发现可识别冲突</td>')
    expect(reportHtml).not.toContain('<td>清源咨询有限公司</td>\n        <td>清源咨询有限公司</td>')
  })

  test('具有复核权限的角色对机器 BLOCKED 显示页面内人工复核区', async ({ page }) => {
    await openConflictScenario(page, 'COMPLETE_EVIDENCE', 'conflictOfficer')
    const dialog = page.getByRole('dialog', { name: '冲突检测详情' })
    await expect(dialog.getByRole('heading', { name: '人工复核记录' })).toBeVisible()
    await expect(dialog.getByRole('button', { name: '提交人工复核结论' })).toBeVisible()
    await expect(dialog.getByRole('button', { name: '发起冲突审批' })).toHaveCount(0)
    await expect(dialog.getByRole('button', { name: '申请豁免评估' })).toHaveCount(0)
  })

  test('技术管理员不能代替专业冲突复核人查看冲突详情', async ({ page }) => {
    await seedAuthenticatedUser(page, 'admin')
    await page.goto('/conflict?task_id=CCT_373e_complete_evidence')
    await waitForPageLoad(page)
    await waitForAppShell(page)

    await expect(page.getByText('无权访问', { exact: true })).toBeVisible()
    await expect(page.getByRole('dialog', { name: '冲突检测详情' })).toHaveCount(0)
  })
})

test.describe('立案冲突门禁', () => {
  test.beforeEach(async ({ page }) => {
    await seedAuthenticatedUser(page, 'lawyer')
  })

  test('BLOCKED 检测完成后提交按钮预先禁用并显示原因', async ({ page }) => {
    await page.route('**/api/v1/case-intakes/801/conflict-check', async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            taskId: 'CHK-2026-ASYNC-001', status: 'COMPLETED', result: {
              checkId: 'CHK-2026-ASYNC-001',
              record: { id: 'CHK-2026-ASYNC-001' },
              hasConflict: true,
              decision: { status: 'BLOCKED' },
              riskAssessment: { overallRisk: 'CRITICAL', riskScore: 98, requiresApproval: true },
            },
          },
        }),
      })
    })
    await page.goto('/case/create')
    await fillCaseIntakeBasics(page)
    await page.getByRole('button', { name: '保存并进行利益冲突检查' }).click()
    await expect(page.locator('.ant-message')).toContainText('利益冲突检查已完成')

    const submit = page.getByRole('button', { name: '提交审批并等待成案' })
    await expect(submit).toBeDisabled()
    await submit.locator('..').hover()
    await expect(page.getByRole('tooltip')).toContainText('精确命中已禁止立案')
    await expect(page.getByRole('button', { name: '进入本案冲突复核' })).toHaveCount(1)
  })

  test('STALE 在步骤区和底栏统一过期，并可保存最新输入重新检测', async ({ page }) => {
    let conflictCheckRequests = 0
    let intakeUpdateRequests = 0
    let latestIntakeUpdateBody = ''
    page.on('request', (request) => {
      if (new URL(request.url()).pathname === '/api/v1/case-intakes/801/conflict-check' && request.method() === 'POST') conflictCheckRequests += 1
      if (new URL(request.url()).pathname === '/api/v1/case-intakes/801' && request.method() === 'PUT') {
        intakeUpdateRequests += 1
        latestIntakeUpdateBody = request.postData() || ''
      }
    })
    await page.goto('/case/create')
    await fillCaseIntakeBasics(page)
    await page.getByRole('button', { name: '保存并进行利益冲突检查' }).click()
    await expect(page.locator('.ant-message')).toContainText('利益冲突检查已完成')

    await page.getByRole('button', { name: '返回基本信息' }).click()
    await page.getByPlaceholder('输入对方当事人名称').fill('变更后的冲突测试对方')
    await page.locator('.batch-stepper').getByRole('button', { name: /利益冲突检查/ }).click()
    await expect(page.getByText('结果已过期', { exact: true })).toHaveCount(3)
    await expect(page.getByRole('button', { name: '提交审批并等待成案' })).toBeDisabled()
    await expect(page.getByRole('button', { name: '进入本案冲突复核' })).toHaveCount(0)

    await page.getByRole('button', { name: '保存最新输入并检测', exact: true }).click()
    await expect.poll(() => intakeUpdateRequests).toBe(1)
    expect(latestIntakeUpdateBody).toContain('变更后的冲突测试对方')
    await expect.poll(() => conflictCheckRequests).toBe(2)
    await expect(page.getByText('结果已过期', { exact: true })).toHaveCount(0)
  })
})

test.describe('冲突审批快照兼容', () => {
  test('优先展示新快照顶层客户、对方和主要来源证据', async ({ page }) => {
    await seedAuthenticatedUser(page, 'lawyer')
    await page.route('**/api/v1/approvals/702', async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            id: 702,
            request_number: 'AP-CONFLICT-702',
            title: '新形状冲突审批快照',
            type: 'conflict_approval',
            status: 'pending',
            applicant_name: '张律师',
            metadata: {
              client: { name: '旧版客户不应展示' },
              parties: [{ role: 'opposing_party', name: '旧版对方不应展示' }],
            },
          },
        }),
      })
    })
    await page.route('**/api/v1/approvals/702/snapshot', async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            snapshot: {
              client_name: '新快照真实客户有限公司',
              opposing_parties: [{ name: '新快照真实对方集团' }],
              subjects: [{ role: 'opposing_party', name: '新快照真实对方集团' }],
              normalizedSubjects: [{ role: 'opposing_party', originalName: '新快照真实对方集团', normalizedName: '新快照真实对方集团有限公司' }],
              decision: { status: 'REVIEW_REQUIRED', primarySubject: '决策层候选主体' },
              evidence: [{
                requestedParty: '结构化主要命中主体',
                sourceCaseNumber: 'CASE-SOURCE-2026-009',
                ruleCode: 'CONFLICT-CLIENT-EXACT-001',
                summary: '该对方是本所现有客户。',
              }],
              metadata: {
                client_name: 'metadata 客户不应覆盖顶层',
                opposing_parties: [{ name: 'metadata 对方不应覆盖顶层' }],
                subjects: [{ name: 'metadata 主体不应覆盖顶层' }],
                normalizedSubjects: [{ normalizedName: 'metadata 规范名不应覆盖顶层' }],
                decision: { status: 'BLOCKED' },
                evidence: [{ requestedParty: 'metadata 证据不应覆盖顶层' }],
              },
            },
          },
        }),
      })
    })
    await page.route('**/api/v1/integration/approvals/702/status', async (route) => {
      await route.fulfill({ contentType: 'application/json', body: JSON.stringify({ success: true, data: { status: 'pending' } }) })
    })

    await page.goto('/approval/702')
    await waitForPageLoad(page)
    await waitForAppShell(page)
    await expect(page.getByRole('heading', { name: '新形状冲突审批快照' })).toBeVisible()

    const applicationInfo = page.locator('section.ng-panel', { has: page.getByRole('heading', { name: '申请信息' }) })
    const clientRow = applicationInfo.locator('p').filter({ hasText: '关联客户' })
    const opposingPartyRow = applicationInfo.locator('p').filter({ hasText: '对方当事人' })
    await expect(clientRow).toBeVisible()
    await expect(clientRow).toContainText('新快照真实客户有限公司')
    await expect(opposingPartyRow).toBeVisible()
    await expect(opposingPartyRow).toContainText('新快照真实对方集团')
    const conflictSummary = page.locator('section.ng-panel', { has: page.getByRole('heading', { name: '冲突检测摘要' }) })
    await expect(page.getByText('总体风险等级：待人工复核')).toBeVisible()
    await expect(page.getByText(/不能作为无冲突结论/)).toBeVisible()
    const primarySubjectRow = conflictSummary.locator('p').filter({ hasText: '主体：结构化主要命中主体' })
    const evidenceRow = conflictSummary.locator('p').filter({ hasText: 'CASE-SOURCE-2026-009' })
    await expect(primarySubjectRow).toBeVisible()
    await expect(evidenceRow).toBeVisible()
    await expect(evidenceRow).toContainText(/CASE-SOURCE-2026-009.*CONFLICT-CLIENT-EXACT-001.*该对方是本所现有客户/)
    await expect(page.getByText('来自审批快照', { exact: true })).toHaveCount(0)
    await expect(page.getByText(/不应展示|不应覆盖顶层/)).toHaveCount(0)
  })
})
