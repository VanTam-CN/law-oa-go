/**
 * E2E: 律师工作台入口与搜索反馈.
 */

import { test, expect } from '@playwright/test'
import { seedAuthenticatedUser, waitForAppShell, waitForPageLoad } from './utils/test-helpers'

test.describe('律师工作台', () => {
  test.beforeEach(async ({ page }) => {
    await seedAuthenticatedUser(page, 'lawyer')
    await page.goto('/dashboard')
    await waitForPageLoad(page)
  })

  test('新建立案入口应该直达立案工作台', async ({ page }) => {
    await waitForAppShell(page)
    await page.getByRole('button', { name: '新建立案' }).click()

    await expect(page).toHaveURL(/\/case\/create$/)
    await expect(page.getByRole('heading', { name: '新建案件立案工作台' })).toBeVisible()
  })

  test('全局搜索应该展示结果反馈', async ({ page }) => {
    await waitForAppShell(page)
    await page.getByRole('searchbox', { name: '搜索案件、冲突检测或审批' }).fill('示例科技')
    await page.keyboard.press('Enter')

    await expect(page.getByText(/找到 \d+ 条相关案件、冲突或审批记录/)).toBeVisible()
  })

  test('390px 视口应移除隐藏侧栏占位并保留可用内容宽度', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 })

    const content = page.locator('.app-main-content')
    await expect(content).toBeVisible()
    const box = await content.boundingBox()

    expect(box).not.toBeNull()
    expect(box!.x).toBeLessThanOrEqual(20)
    expect(box!.width).toBeGreaterThanOrEqual(350)
    await expect(page.getByRole('heading', { name: /张律师/ })).toBeVisible()
  })
})
