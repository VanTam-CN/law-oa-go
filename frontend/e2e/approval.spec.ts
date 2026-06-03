/**
 * E2E: 当前 MVP 审批流程.
 */

import { test, expect } from '@playwright/test'
import {
  seedAuthenticatedUser,
  waitForAppShell,
  waitForNativeTable,
  waitForPageLoad,
} from './utils/test-helpers'

test.describe('审批工作台', () => {
  test.beforeEach(async ({ page }) => {
    await seedAuthenticatedUser(page, 'lawyer')
    await page.goto('/approval')
    await waitForPageLoad(page)
  })

  test('应该显示审批队列', async ({ page }) => {
    await waitForAppShell(page)
    await expect(page.getByRole('heading', { name: '审批工作台' })).toBeVisible()
    await waitForNativeTable(page)
    await expect(page.getByText('AP-2026-001')).toBeVisible()
    await expect(page.getByRole('cell', { name: '冲突审查审批 - 红杉资本投资管理咨询合同纠纷案' })).toBeVisible()
  })

  test('审批分类按钮应该可见', async ({ page }) => {
    await expect(page.getByRole('button', { name: /全\s*部/ })).toBeVisible()
    await expect(page.getByRole('button', { name: '冲突审查' })).toBeVisible()
    await expect(page.getByRole('button', { name: '豁免披露' })).toBeVisible()
    await expect(page.getByRole('button', { name: '待补充' })).toBeVisible()
  })

  test('应该能进入审批详情', async ({ page }) => {
    await page.locator('tr', { hasText: 'AP-2026-001' }).getByRole('button', { name: '进入审批' }).click()

    await expect(page).toHaveURL(/\/approval\/701$/)
    await expect(page.getByRole('heading', { name: '冲突审查审批 - 红杉资本投资管理咨询合同纠纷案' })).toBeVisible()
  })
})

test.describe('审批决策台', () => {
  test.beforeEach(async ({ page }) => {
    await seedAuthenticatedUser(page, 'lawyer')
    await page.goto('/approval/701')
    await waitForPageLoad(page)
  })

  test('律师不是当前审批人时不应该显示审批决策按钮', async ({ page }) => {
    await waitForAppShell(page)
    await expect(page.getByRole('button', { name: '同意并成案' })).toHaveCount(0)
    await expect(page.getByRole('button', { name: /^拒绝$/ })).toHaveCount(0)
    await expect(page.getByRole('button', { name: '退回修改' })).toHaveCount(0)
    await expect(page.getByText('当前账号仅可查看审批进度')).toBeVisible()
  })
})

test.describe('审批人决策台', () => {
  test.beforeEach(async ({ page }) => {
    await seedAuthenticatedUser(page, 'admin')
    await page.goto('/approval/701')
    await waitForPageLoad(page)
  })

  test('当前审批人同意审批后应该显示成案状态', async ({ page }) => {
    await waitForAppShell(page)
    await page.getByRole('button', { name: '同意并成案' }).click()

    await expect(page.locator('.ant-message')).toContainText('已成案：HD-2026-001')
    await expect(page.getByRole('button', { name: '查看关联案件' })).toBeEnabled()
  })
})
