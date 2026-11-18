/**
 * E2E: 当前 MVP 审批流程.
 */

import { test, expect } from '@playwright/test'
import {
  seedAuthenticatedUser,
  waitForAppShell,
  waitForNativeTable,
  waitForPageLoad,
} from './utils/test-helpers'

test.describe('审批工作台', () => {
  test.beforeEach(async ({ page }) => {
    await seedAuthenticatedUser(page, 'lawyer')
    await page.goto('/approval')
    await waitForPageLoad(page)
  })

  test('应该显示审批队列', async ({ page }) => {
    await waitForAppShell(page)
    await expect(page.getByRole('heading', { name: '审批工作台' })).toBeVisible()
    await waitForNativeTable(page)
    await expect(page.getByRole('columnheader', { name: '优先级' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: '风险' })).toHaveCount(0)
    await expect(page.getByText('AP-2026-001')).toBeVisible()
    await expect(page.getByRole('cell', { name: '冲突审查审批 - 红杉资本投资管理咨询合同纠纷案' })).toBeVisible()
  })

  test('审批分类按钮应该可见', async ({ page }) => {
    await expect(page.getByRole('button', { name: /全\s*部/ })).toBeVisible()
    await expect(page.getByRole('button', { name: '冲突审查' })).toBeVisible()
    await expect(page.getByRole('button', { name: '豁免披露' })).toBeVisible()
    await expect(page.getByRole('button', { name: '待补充' })).toBeVisible()
  })

  test('应该能进入审批详情', async ({ page }) => {
    await page.locator('tr', { hasText: 'AP-2026-001' }).getByRole('button', { name: '进入审批' }).click()

    await expect(page).toHaveURL(/\/approval\/701$/)
    await expect(page.getByRole('heading', { name: '冲突审查审批 - 红杉资本投资管理咨询合同纠纷案' })).toBeVisible()
  })
})

test.describe('审批决策台', () => {
  test.beforeEach(async ({ page }) => {
    await seedAuthenticatedUser(page, 'lawyer')
    await page.goto('/approval/701')
    await waitForPageLoad(page)
  })

  test('律师不是当前审批人时不应该显示审批决策按钮', async ({ page }) => {
    await waitForAppShell(page)
    await expect(page.getByRole('button', { name: '同意并成案' })).toHaveCount(0)
    await expect(page.getByRole('button', { name: /^拒绝$/ })).toHaveCount(0)
    await expect(page.getByRole('button', { name: '退回修改' })).toHaveCount(0)
    await expect(page.locator('.batch-approval-readonly')).toHaveText('申请人不能审批自己的申请，仅可查看审批进度。')
  })
})

test.describe('审批人决策台', () => {
  test.beforeEach(async ({ page }) => {
    await seedAuthenticatedUser(page, 'admin')
    await page.goto('/approval/701')
    await waitForPageLoad(page)
  })

  test('当前审批人同意审批后应该显示成案状态', async ({ page }) => {
    await waitForAppShell(page)
    await page.getByRole('button', { name: '同意并成案' }).click()
    const dialog = page.getByRole('dialog', { name: '确认同意并成案' })
    await expect(dialog).toBeVisible()
    await dialog.getByRole('textbox', { name: '处理依据' }).fill('冲突复核已经完成，申请材料齐全，同意正式成案。')
    await dialog.getByRole('button', { name: '确认同意并成案' }).click()

    await expect(page.locator('.ant-message')).toContainText('已成案：DEMO-2026-001')
    await expect(page.getByRole('button', { name: '查看关联案件' })).toBeEnabled()
  })

  test('拒绝审批必须填写依据并二次确认，首次点击不得直接提交', async ({ page }) => {
    let decisionRequests = 0
    page.on('request', (request) => {
      if (request.url().includes('/integration/approvals/701/decision')) decisionRequests += 1
    })

    await waitForAppShell(page)
    await page.getByRole('button', { name: '拒 绝' }).click()
    const dialog = page.getByRole('dialog', { name: '确认拒绝申请' })
    await expect(dialog).toBeVisible()
    expect(decisionRequests).toBe(0)

    await dialog.getByRole('textbox', { name: '处理依据' }).fill('理由太短')
    await dialog.getByRole('button', { name: '确认拒绝' }).click()
    await expect(page.locator('.ant-message')).toContainText('请填写不少于 10 个字的处理依据')
    expect(decisionRequests).toBe(0)

    await dialog.getByRole('button', { name: '取 消' }).click()
    await expect(dialog).toBeHidden()
    expect(decisionRequests).toBe(0)
  })
})

test.describe('审批证据一致性', () => {
  test('检测编号与旧快照不一致时应该显示权威结果并阻止审批', async ({ page }) => {
    await seedAuthenticatedUser(page, 'admin')
    await page.route('**/api/v1/approvals/701', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            id: 701,
            request_number: 'AP-2026-001',
            title: '冲突审查审批 - 名称候选核实',
            type: 'conflict_approval',
            status: 'pending',
            applicant_id: '2',
            applicant_name: '张律师',
            current_approver_id: '1',
            current_approver_name: '示例管理员',
            conflict_check_id: 'CHECK-MEDIUM-001',
            conflict_result: {
              checkId: 'CHECK-MEDIUM-001',
              riskAssessment: { overallRisk: 'MEDIUM', riskScore: 58 },
            },
            metadata: {
              conflict_task_id: 'CHECK-HIGH-OLD',
              conflict_result: {
                checkId: 'CHECK-HIGH-OLD',
                riskAssessment: { overallRisk: 'HIGH', riskScore: 92 },
              },
            },
          },
        }),
      })
    })
    await page.route('**/api/v1/approvals/701/snapshot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            snapshot: {
              metadata: {
                conflict_task_id: 'CHECK-HIGH-OLD',
                conflict_result: {
                  checkId: 'CHECK-HIGH-OLD',
                  riskAssessment: { overallRisk: 'HIGH', riskScore: 92 },
                },
              },
            },
          },
        }),
      })
    })

    await page.goto('/approval/701')
    await waitForPageLoad(page)

    await expect(page.getByRole('heading', { name: '审批证据不一致' })).toBeVisible()
    await expect(page.getByText('总体风险等级：中风险')).toBeVisible()
    await expect(page.getByText('检测记录：证据已冻结')).toBeVisible()
    await expect(page.getByText('查看审批审计信息')).toBeVisible()
    await expect(page.getByRole('button', { name: '同意并成案' })).toHaveCount(0)
    await expect(page.locator('.batch-approval-readonly')).toHaveText('审批证据不一致，已禁止处理')
  })
})

test.describe('审批审计记录', () => {
  test('终态审批应显示实际核查人和不可删除的处理依据', async ({ page }) => {
    await seedAuthenticatedUser(page, 'conflictOfficer')
    await page.route('**/api/v1/approvals/701', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            id: 701,
            request_number: 'AP-2026-001',
            title: '冲突审查审批 - 主体身份材料不足',
            type: 'conflict_approval',
            status: 'rejected',
            applicant_id: '2',
            applicant_name: '张律师',
            current_approver_id: '',
            current_approver_name: '',
            records: [
              {
                id: 'record-rejected',
                approver_id: '3',
                approver_name: '独立冲突核查人',
                approver_role: '冲突核查人',
                decision: 'reject',
                decision_reason: '缺少统一社会信用代码，现有名称候选不足以确认同一主体。',
                approval_date: '2026-07-31T15:00:00.000Z',
              },
            ],
          },
        }),
      })
    })

    await page.goto('/approval/701')
    await waitForPageLoad(page)
    const auditRecords = page.locator('section.ng-panel').filter({ hasText: '审批记录（1）' })
    await expect(auditRecords.getByText('独立冲突核查人')).toBeVisible()
    const decisionReason = auditRecords.getByText(
      '缺少统一社会信用代码，现有名称候选不足以确认同一主体。',
    )
    await decisionReason.scrollIntoViewIfNeeded()
    await expect(decisionReason).toBeVisible()
    await expect(page.getByRole('heading', { name: /意见记录/ })).toHaveCount(0)
  })
})
