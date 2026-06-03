/**
 * E2E: 律师端主要页面布局不应产生页面级横向溢出.
 */

import { test, expect } from '@playwright/test'
import { seedAuthenticatedUser, waitForAppShell, waitForPageLoad } from './utils/test-helpers'

async function expectNoPageHorizontalOverflow(page: any) {
  const sizes = await page.evaluate(() => ({
    scrollWidth: document.documentElement.scrollWidth,
    innerWidth: window.innerWidth,
  }))
  expect(sizes.scrollWidth).toBeLessThanOrEqual(sizes.innerWidth + 4)
}

test.describe('律师端响应式布局', () => {
  test('冲突检测页面不应该撑破视口', async ({ page }) => {
    await page.setViewportSize({ width: 1366, height: 900 })
    await seedAuthenticatedUser(page, 'lawyer')
    await page.goto('/conflict')
    await waitForPageLoad(page)

    await waitForAppShell(page)
    await expectNoPageHorizontalOverflow(page)
  })

  test('客户主档案页面不应该撑破视口', async ({ page }) => {
    await page.setViewportSize({ width: 1200, height: 900 })
    await seedAuthenticatedUser(page, 'lawyer')
    await page.goto('/client')
    await waitForPageLoad(page)

    await waitForAppShell(page)
    await expectNoPageHorizontalOverflow(page)
  })
})

