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
    await expect(page.getByText('我的待办')).toBeVisible()
    await expect(page.getByText('审批红杉资本新接案')).toBeVisible()
  })

  test('律师工作台不得请求或展示财务指标', async ({ page }) => {
    let financeRequestCount = 0
    page.on('request', (request) => {
      if (new URL(request.url()).pathname === '/api/v1/finance/overview') financeRequestCount += 1
    })
    await page.reload()
    await waitForPageLoad(page)

    expect(financeRequestCount).toBe(0)
    await expect(page.getByText('合同回款预警')).toHaveCount(0)
  })

  test('全局搜索应展示可进入的业务结果', async ({ page }) => {
    const search = page.getByRole('textbox', { name: '搜索案件、冲突检测或审批' })
    await search.fill('红杉资本')
    await search.press('Enter')

    const result = page.getByRole('button', { name: /案件：红杉资本投资管理咨询合同纠纷案/ })
    await expect(result).toBeVisible()
    await result.click()
    await expect(page).toHaveURL(/\/case\/101$/)
  })
})
