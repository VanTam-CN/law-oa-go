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

    const bannerChildren = await page.locator('.batch-risk-banner > *').evaluateAll((elements) =>
      elements.map((element) => {
        const box = element.getBoundingClientRect()
        return { left: box.left, right: box.right, top: box.top, bottom: box.bottom }
      }),
    )
    for (let left = 0; left < bannerChildren.length; left += 1) {
      for (let right = left + 1; right < bannerChildren.length; right += 1) {
        const a = bannerChildren[left]
        const b = bannerChildren[right]
        const horizontal = Math.min(a.right, b.right) - Math.max(a.left, b.left)
        const vertical = Math.min(a.bottom, b.bottom) - Math.max(a.top, b.top)
        expect(horizontal > 1 && vertical > 1).toBe(false)
      }
    }
  })

  test('窄屏下侧栏自动折叠且冲突清单仍保留可用宽度', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 })
    await seedAuthenticatedUser(page, 'lawyer')
    await page.goto('/conflict')
    await waitForPageLoad(page)
    await waitForAppShell(page)

    await expectNoPageHorizontalOverflow(page)
    const siderWidth = await page.locator('.ant-layout-sider').evaluate((element) =>
      element.getBoundingClientRect().width,
    )
    const contentWidth = await page.locator('main.ant-layout-content').evaluate((element) =>
      element.getBoundingClientRect().width,
    )
    expect(siderWidth).toBeLessThanOrEqual(80)
    expect(contentWidth).toBeGreaterThanOrEqual(260)
  })

  test('客户主档案页面不应该撑破视口', async ({ page }) => {
    await page.setViewportSize({ width: 1200, height: 900 })
    await seedAuthenticatedUser(page, 'lawyer')
    await page.goto('/client')
    await waitForPageLoad(page)

    await waitForAppShell(page)
    await expectNoPageHorizontalOverflow(page)
  })

  test('平板与小窗口下侧栏偏移和正文宽度保持一致', async ({ page }) => {
    await seedAuthenticatedUser(page, 'lawyer')

    await page.setViewportSize({ width: 960, height: 760 })
    await page.goto('/dashboard')
    await waitForPageLoad(page)
    await waitForAppShell(page)
    await expectNoPageHorizontalOverflow(page)
    let siderBox = await page.locator('.ant-layout-sider').evaluate((element) => {
      const box = element.getBoundingClientRect()
      return { left: box.left, right: box.right, width: box.width }
    })
    let contentBox = await page.locator('main.ant-layout-content').evaluate((element) => {
      const box = element.getBoundingClientRect()
      return { left: box.left, right: box.right, width: box.width }
    })
    expect(siderBox.width).toBeLessThanOrEqual(80)
    expect(contentBox.left).toBeLessThanOrEqual(104)
    expect(contentBox.right).toBeLessThanOrEqual(960)

    await page.setViewportSize({ width: 640, height: 760 })
    await page.reload()
    await waitForPageLoad(page)
    await waitForAppShell(page)
    await page.waitForTimeout(500)
    await expectNoPageHorizontalOverflow(page)
    siderBox = await page.locator('.ant-layout-sider').evaluate((element) => {
      const box = element.getBoundingClientRect()
      return { left: box.left, right: box.right, width: box.width }
    })
    contentBox = await page.locator('main.ant-layout-content').evaluate((element) => {
      const box = element.getBoundingClientRect()
      return { left: box.left, right: box.right, width: box.width }
    })
    expect(siderBox.right).toBeLessThanOrEqual(0)
    expect(contentBox.left).toBeLessThanOrEqual(16)
    expect(contentBox.width).toBeGreaterThanOrEqual(600)
  })
})
