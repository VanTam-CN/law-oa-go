/**
 * E2E: 审批列表操作列布局 + 审批详情按钮唯一性（QA-RC-003 + IB-RT-004 回归）.
 */

import { test, expect } from '@playwright/test'
import { seedAuthenticatedUser, waitForAppShell, waitForPageLoad } from './utils/test-helpers'

test.describe('审批列表操作列可达性', () => {
  test.beforeEach(async ({ page }) => {
    await seedAuthenticatedUser(page, 'lawyer')
  })

  for (const width of [1366, 1200]) {
    test(`在 ${width}px 下操作列按钮可见或可滚动访问`, async ({ page }) => {
      await page.setViewportSize({ width, height: 768 })
      await page.goto('/approval')
      await waitForPageLoad(page)
      await waitForAppShell(page)

      const actionButton = page.getByRole('button', { name: /进入审批/ }).first()
      await expect(actionButton).toBeVisible()
    })
  }
})

test.describe('审批详情按钮唯一性（IB-RT-004）', () => {
  test('更多审批操作和更多处理方式名称唯一', async ({ page }) => {
    await seedAuthenticatedUser(page, 'lawyer')
    await page.goto('/approval/701')
    await waitForPageLoad(page)
    await waitForAppShell(page)

    await expect(page.getByRole('button', { name: '更多审批操作' })).toBeVisible()
    await expect(page.getByRole('button', { name: '更多处理方式' })).toBeVisible()
  })
})
