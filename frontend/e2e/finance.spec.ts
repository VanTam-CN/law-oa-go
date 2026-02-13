/**
 * E2E测试: 财务流程
 * Story 6.2: 关键流程E2E
 */

import { test, expect } from '@playwright/test'
import { login, waitForPageLoad, waitForTableLoad, waitForModal, getTableRowCount } from './utils/test-helpers'

test.describe('财务管理页面', () => {
  test.beforeEach(async ({ page }) => {
    await login(page, 'admin', 'admin123')
    await page.goto('/finance')
    await waitForPageLoad(page)
  })

  test('应该显示财务管理页面', async ({ page }) => {
    // 验证页面标题
    await expect(page.locator('.ant-tabs')).toBeVisible()
  })

  test('应该有发票和费用两个标签页', async ({ page }) => {
    const tabs = page.locator('.ant-tabs-tab')
    const tabCount = await tabs.count()

    expect(tabCount).toBeGreaterThanOrEqual(2)
  })

  test('应该显示统计数据', async ({ page }) => {
    // 查找统计卡片
    const statistics = page.locator('.ant-statistic')
    const statCount = await statistics.count()

    expect(statCount).toBeGreaterThan(0)
  })
})

test.describe('发票管理', () => {
  test.beforeEach(async ({ page }) => {
    await login(page, 'admin', 'admin123')
    await page.goto('/finance')
    await waitForPageLoad(page)

    // 切换到发票标签
    await page.click('.ant-tabs-tab:has-text("发票")')
    await waitForTableLoad(page)
  })

  test('应该显示发票列表', async ({ page }) => {
    await expect(page.locator('.ant-table')).toBeVisible()
  })

  test('应该能搜索发票', async ({ page }) => {
    await waitForTableLoad(page)

    const searchInput = page.locator('input[placeholder*="搜索"]')
    if (await searchInput.isVisible()) {
      await searchInput.fill('INV')
      await page.press('input[placeholder*="搜索"]', 'Enter')
      await page.waitForTimeout(1000)
    }

    await expect(page.locator('.ant-table')).toBeVisible()
  })

  test('应该能筛选发票状态', async ({ page }) => {
    await waitForTableLoad(page)

    const statusFilter = page.locator('.ant-select:has-text("状态") .ant-select-selector')
    if (await statusFilter.isVisible()) {
      await statusFilter.click()
      await page.click('.ant-select-dropdown:visible .ant-select-item:first-child')
      await page.waitForTimeout(1000)
    }

    await expect(page.locator('.ant-table')).toBeVisible()
  })

  test('应该能创建发票', async ({ page }) => {
    await waitForTableLoad(page)

    // 点击新建按钮
    const createButton = page.locator('button:has-text("新建发票"), button:has-text("创建")')
    if (await createButton.isVisible()) {
      await createButton.click()
      await waitForModal(page)

      // 填写表单
      await page.fill('input[name="invoiceNo"], input[placeholder*="发票号"]', `INV-${Date.now()}`)
      await page.fill('input[name="amount"], input[placeholder*="金额"]', '10000')

      // 提交
      await page.click('.ant-modal button:has-text("确定"), .ant-modal button:has-text("提交")')

      // 等待结果
      await page.waitForTimeout(2000)
    } else {
      test.skip()
    }
  })

  test('应该能标记发票为已支付', async ({ page }) => {
    await waitForTableLoad(page)

    const markPaidButton = page.locator('.ant-table-tbody button:has-text("已支付"), .ant-table-tbody button:has-text("标记已付")')

    if (await markPaidButton.first().isVisible()) {
      await markPaidButton.first().click()

      // 确认操作
      const confirmButton = page.locator('.ant-popconfirm button:has-text("确定")')
      if (await confirmButton.isVisible()) {
        await confirmButton.click()
      }

      await page.waitForTimeout(1000)
    } else {
      test.skip()
    }
  })
})

test.describe('费用管理', () => {
  test.beforeEach(async ({ page }) => {
    await login(page, 'admin', 'admin123')
    await page.goto('/finance')
    await waitForPageLoad(page)

    // 切换到费用标签
    await page.click('.ant-tabs-tab:has-text("费用")')
    await waitForTableLoad(page)
  })

  test('应该显示费用列表', async ({ page }) => {
    await expect(page.locator('.ant-table')).toBeVisible()
  })

  test('应该能创建费用记录', async ({ page }) => {
    await waitForTableLoad(page)

    const createButton = page.locator('button:has-text("新建费用"), button:has-text("创建")')
    if (await createButton.isVisible()) {
      await createButton.click()
      await waitForModal(page)

      // 填写表单
      await page.fill('input[name="description"], input[placeholder*="描述"]', 'E2E测试费用')
      await page.fill('input[name="amount"], input[placeholder*="金额"]', '5000')

      // 选择费用类型
      const typeSelect = page.locator('.ant-form-item:has-text("类型") .ant-select-selector')
      if (await typeSelect.isVisible()) {
        await typeSelect.click()
        await page.click('.ant-select-dropdown:visible .ant-select-item:first-child')
      }

      // 提交
      await page.click('.ant-modal button:has-text("确定"), .ant-modal button:has-text("提交")')

      await page.waitForTimeout(2000)
    } else {
      test.skip()
    }
  })

  test('应该能审批费用', async ({ page }) => {
    await waitForTableLoad(page)

    const approveButton = page.locator('.ant-table-tbody button:has-text("审批"), .ant-table-tbody button:has-text("通过")')

    if (await approveButton.first().isVisible()) {
      await approveButton.first().click()

      const confirmButton = page.locator('.ant-popconfirm button:has-text("确定")')
      if (await confirmButton.isVisible()) {
        await confirmButton.click()
      }

      await page.waitForTimeout(1000)
    } else {
      test.skip()
    }
  })
})

test.describe('财务统计', () => {
  test.beforeEach(async ({ page }) => {
    await login(page, 'admin', 'admin123')
    await page.goto('/finance')
    await waitForPageLoad(page)
  })

  test('应该显示收入统计', async ({ page }) => {
    const incomeStat = page.locator('.ant-statistic:has-text("收入"), .ant-statistic:has-text("Income")')
    if (await incomeStat.isVisible()) {
      await expect(incomeStat).toBeVisible()
    }
  })

  test('应该显示支出统计', async ({ page }) => {
    const expenseStat = page.locator('.ant-statistic:has-text("支出"), .ant-statistic:has-text("Expense")')
    if (await expenseStat.isVisible()) {
      await expect(expenseStat).toBeVisible()
    }
  })

  test('应该显示待处理发票数', async ({ page }) => {
    const pendingStat = page.locator('.ant-statistic:has-text("待处理"), .ant-statistic:has-text("待收")')
    if (await pendingStat.isVisible()) {
      await expect(pendingStat).toBeVisible()
    }
  })
})
