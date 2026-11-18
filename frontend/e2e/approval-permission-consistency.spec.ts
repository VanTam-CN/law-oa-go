/**
 * E2E: 审批详情权限一致性（QA-RC-004 回归）.
 */

import { test, expect } from '@playwright/test'
import { seedAuthenticatedUser, waitForAppShell, waitForPageLoad } from './utils/test-helpers'

test.describe('审批详情权限一致性', () => {
  test('独立冲突核查人可以进入审批中心处理被分配事项', async ({ page }) => {
    await seedAuthenticatedUser(page, 'conflictOfficer')
    await page.goto('/dashboard')
    await waitForPageLoad(page)
    await waitForAppShell(page)

    const approvalMenu = page.getByRole('menuitem', { name: /审批中心/ })
    await expect(approvalMenu).toBeVisible()
    await approvalMenu.click()
    await expect(page).toHaveURL(/\/approval$/)
  })

  test('非当前审批人的律师应看到只读提示，无决策按钮', async ({ page }) => {
    await seedAuthenticatedUser(page, 'lawyer')
    await page.goto('/approval/701')
    await waitForPageLoad(page)
    await waitForAppShell(page)

    // 律师 ID 为 2，审批的 current_approver_id 为 1，不匹配 → 只读
    await expect(page.locator('.batch-approval-readonly')).toHaveText('申请人不能审批自己的申请，仅可查看审批进度。')
    await expect(page.getByRole('button', { name: '同意并成案' })).not.toBeVisible()
  })
})
