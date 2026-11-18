/**
 * E2E: 个人中心表单校验.
 */

import { test, expect } from '@playwright/test'
import { seedAuthenticatedUser, waitForAppShell, waitForPageLoad } from './utils/test-helpers'

test.describe('个人中心', () => {
  test('修改密码空表单应该显示字段级校验', async ({ page }) => {
    await seedAuthenticatedUser(page, 'lawyer')
    await page.goto('/profile')
    await waitForPageLoad(page)

    await waitForAppShell(page)
    await page.getByRole('button', { name: '修改密码' }).click()
    await page.getByRole('button', { name: '确认修改' }).click()

    await expect(page.getByText('请输入当前密码')).toBeVisible()
    await expect(page.getByText('请输入新密码')).toBeVisible()
    await expect(page.getByText('请确认新密码')).toBeVisible()
  })
})

