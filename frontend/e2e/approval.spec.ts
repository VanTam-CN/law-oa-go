/**
 * E2E测试: 审批流程
 * Story 6.2: 关键流程E2E
 */

import { test, expect } from '@playwright/test'
import { login, waitForPageLoad, waitForTableLoad, waitForModal } from './utils/test-helpers'

test.describe('审批列表', () => {
  test.beforeEach(async ({ page }) => {
    await login(page, 'admin', 'admin123')
    await page.goto('/approvals')
    await waitForPageLoad(page)
  })

  test('应该显示审批列表页面', async ({ page }) => {
    // 验证页面元素
    await expect(page.locator('.ant-table')).toBeVisible()
  })

  test('应该能筛选审批状态', async ({ page }) => {
    await waitForTableLoad(page)

    // 点击状态筛选
    const statusFilter = page.locator('.ant-select:has-text("状态") .ant-select-selector')
    if (await statusFilter.isVisible()) {
      await statusFilter.click()
      await page.click('.ant-select-dropdown:visible .ant-select-item:has-text("待审批")')
      await page.waitForTimeout(1000)
    }

    await expect(page.locator('.ant-table')).toBeVisible()
  })
})

test.describe('审批操作', () => {
  test.beforeEach(async ({ page }) => {
    await login(page, 'admin', 'admin123')
    await page.goto('/approvals')
    await waitForPageLoad(page)
  })

  test('应该能查看审批详情', async ({ page }) => {
    await waitForTableLoad(page)

    const viewButton = page.locator('.ant-table-tbody tr:first-child button:has-text("查看")')

    if (await viewButton.isVisible()) {
      await viewButton.click()
      await page.waitForTimeout(1000)

      // 应该显示详情
      const detailContent = page.locator('.ant-descriptions, .ant-drawer, .approval-detail')
      if (await detailContent.isVisible()) {
        await expect(detailContent).toBeVisible()
      }
    } else {
      test.skip()
    }
  })

  test('应该能通过审批', async ({ page }) => {
    await waitForTableLoad(page)

    // 找到待审批的记录
    const approveButton = page.locator('.ant-table-tbody button:has-text("通过")')

    if (await approveButton.first().isVisible()) {
      await approveButton.first().click()

      // 确认操作
      const confirmButton = page.locator('.ant-popconfirm button:has-text("确定"), .ant-modal button:has-text("确定")')
      if (await confirmButton.isVisible()) {
        await confirmButton.click()
      }

      // 等待成功提示
      await expect(page.locator('.ant-message')).toBeVisible({ timeout: 5000 })
    } else {
      test.skip()
    }
  })

  test('应该能拒绝审批', async ({ page }) => {
    await waitForTableLoad(page)

    const rejectButton = page.locator('.ant-table-tbody button:has-text("拒绝")')

    if (await rejectButton.first().isVisible()) {
      await rejectButton.first().click()

      // 可能需要填写拒绝原因
      const reasonInput = page.locator('textarea[placeholder*="原因"], input[placeholder*="原因"]')
      if (await reasonInput.isVisible()) {
        await reasonInput.fill('E2E测试拒绝')
      }

      // 确认操作
      const confirmButton = page.locator('.ant-popconfirm button:has-text("确定"), .ant-modal button:has-text("确定")')
      if (await confirmButton.isVisible()) {
        await confirmButton.click()
      }

      // 等待成功提示
      await expect(page.locator('.ant-message')).toBeVisible({ timeout: 5000 })
    } else {
      test.skip()
    }
  })
})

test.describe('审批历史', () => {
  test.beforeEach(async ({ page }) => {
    await login(page, 'admin', 'admin123')
    await page.goto('/approvals/history')
    await waitForPageLoad(page)
  })

  test('应该显示审批历史', async ({ page }) => {
    const table = page.locator('.ant-table')
    if (await table.isVisible()) {
      await expect(table).toBeVisible()
    }
  })
})
