/**
 * E2E: 审批详情权限一致性（QA-RC-004 回归）.
 */

import { test, expect } from '@playwright/test'
import { seedAuthenticatedUser, waitForAppShell, waitForPageLoad } from './utils/test-helpers'

test.describe('审批详情权限一致性', () => {
  test('非当前审批人的律师应看到只读提示，无决策按钮', async ({ page }) => {
    await seedAuthenticatedUser(page, 'lawyer')
    await page.goto('/approval/701')
    await waitForPageLoad(page)
    await waitForAppShell(page)

    // 律师 ID 为 2，审批的 current_approver_id 为 1，不匹配 → 只读
    await expect(page.getByText(/仅可查看审批进度/)).toBeVisible()
    await expect(page.getByRole('button', { name: '同意并成案' })).not.toBeVisible()
  })
})
