/**
 * E2E: 财务模块在当前 MVP 中的可用性边界.
 */

import { test, expect } from '@playwright/test'
import { seedAuthenticatedUser, waitForAppShell, waitForPageLoad } from './utils/test-helpers'

test.describe('财务管理 MVP 边界', () => {
  test('财务角色访问财务管理时应该看到 MVP 不可用说明', async ({ page }) => {
    await seedAuthenticatedUser(page, 'finance')
    await page.goto('/finance')
    await waitForPageLoad(page)

    await waitForAppShell(page)
    await expect(page.getByText('财务中心未纳入本次 MVP 试用范围')).toBeVisible()
    await expect(page.getByText('当前试用版聚焦主任工作台、案件、利益冲突、客户、审批和信托账户流程。')).toBeVisible()
  })

  test('律师访问财务管理时应该看到无权访问', async ({ page }) => {
    await seedAuthenticatedUser(page, 'lawyer')
    await page.goto('/finance')
    await waitForPageLoad(page)

    await expect(page.getByText('无权访问')).toBeVisible()
    await expect(page.getByText('需要财务角色或管理员授权')).toBeVisible()
  })

  test('不可用页可以返回工作台', async ({ page }) => {
    await seedAuthenticatedUser(page, 'finance')
    await page.goto('/finance')
    await page.getByRole('button', { name: '返回工作台' }).click()

    await expect(page).toHaveURL(/\/dashboard$/)
  })
})
