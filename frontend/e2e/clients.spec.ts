/**
 * E2E: 客户管理入口反馈.
 */

import { test, expect } from '@playwright/test'
import { seedAuthenticatedUser, waitForAppShell, waitForPageLoad } from './utils/test-helpers'

test.describe('客户主档案', () => {
  test.beforeEach(async ({ page }) => {
    await seedAuthenticatedUser(page, 'lawyer')
    await page.goto('/client')
    await waitForPageLoad(page)
  })

  test('新增客户入口应该打开新增客户表单', async ({ page }) => {
    await waitForAppShell(page)
    await page.getByRole('button', { name: '新增客户' }).click()

    await expect(page.getByRole('dialog', { name: '新增客户' })).toBeVisible()
    await expect(page.getByText('客户类型 *')).toBeVisible()
    await expect(page.getByPlaceholder('请输入客户名称')).toBeVisible()
  })
})

