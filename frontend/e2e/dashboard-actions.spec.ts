/**
 * E2E: 工作台查看全部入口反馈（IB-RT-002/003 回归）.
 */

import { test, expect } from '@playwright/test'
import { seedAuthenticatedUser, waitForAppShell, waitForPageLoad } from './utils/test-helpers'

test.describe('工作台查看全部操作', () => {
  test.beforeEach(async ({ page }) => {
    await seedAuthenticatedUser(page, 'lawyer')
    await page.goto('/dashboard')
    await waitForPageLoad(page)
    await waitForAppShell(page)
  })

  test('查看全部冲突任务应跳转到冲突页', async ({ page }) => {
    await page.getByRole('button', { name: '查看全部冲突任务' }).click()
    await expect(page).toHaveURL(/\/conflict/)
    await expect(page.getByRole('heading', { name: '利益冲突检测清单' })).toBeVisible()
    await expect(page.getByRole('heading', { name: '检测任务清单' })).toBeVisible()
    await expect(page.getByRole('dialog', { name: '冲突检测详情' })).not.toBeVisible()
  })

  test('查看全部待办应跳转到收件箱', async ({ page }) => {
    await page.getByRole('button', { name: '查看全部待办' }).click()
    await expect(page).toHaveURL(/\/inbox/)
  })
})
