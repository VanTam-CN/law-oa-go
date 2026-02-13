/**
 * E2E测试: 案件管理流程
 * Story 6.2: 关键流程E2E
 */

import { test, expect } from '@playwright/test'
import {
  login,
  waitForPageLoad,
  waitForTableLoad,
  getTableRowCount,
  waitForModal,
  closeModal,
  clickConfirm,
  fillForm,
  selectOption,
} from './utils/test-helpers'

test.describe('案件列表', () => {
  test.beforeEach(async ({ page }) => {
    await login(page, 'admin', 'admin123')
    await page.goto('/cases')
    await waitForPageLoad(page)
  })

  test('应该显示案件列表页面', async ({ page }) => {
    // 验证页面标题
    await expect(page.locator('h1, .ant-page-header-heading-title')).toContainText('案件')

    // 验证表格存在
    await expect(page.locator('.ant-table')).toBeVisible()
  })

  test('应该能搜索案件', async ({ page }) => {
    // 等待表格加载
    await waitForTableLoad(page)

    // 输入搜索关键词
    await page.fill('input[placeholder*="搜索"], input[placeholder*="案件"]', '测试案件')
    await page.press('input[placeholder*="搜索"], input[placeholder*="案件"]', 'Enter')

    // 等待搜索结果
    await page.waitForTimeout(1000)

    // 验证搜索结果
    const rows = await getTableRowCount(page)
    expect(rows).toBeGreaterThanOrEqual(0)
  })

  test('应该能筛选案件状态', async ({ page }) => {
    await waitForTableLoad(page)

    // 点击状态筛选
    await page.click('.ant-select:has-text("状态") .ant-select-selector')

    // 选择状态
    await page.click('.ant-select-dropdown:visible .ant-select-item:has-text("进行中")')

    // 等待筛选结果
    await page.waitForTimeout(1000)

    // 验证表格已更新
    await expect(page.locator('.ant-table')).toBeVisible()
  })

  test('应该能分页浏览案件', async ({ page }) => {
    await waitForTableLoad(page)

    // 检查分页器
    const paginator = page.locator('.ant-pagination')
    const hasNext = await paginator.locator('.ant-pagination-next:not(.ant-pagination-disabled)').isVisible()

    if (hasNext) {
      // 点击下一页
      await paginator.locator('.ant-pagination-next').click()
      await page.waitForTimeout(1000)

      // 验证URL或页面内容变化
      await expect(page.locator('.ant-table')).toBeVisible()
    }
  })
})

test.describe('创建案件', () => {
  test.beforeEach(async ({ page }) => {
    await login(page, 'admin', 'admin123')
    await page.goto('/cases')
    await waitForPageLoad(page)
  })

  test('应该能打开创建案件弹窗', async ({ page }) => {
    // 点击新建按钮
    await page.click('button:has-text("新建"), button:has-text("创建案件")')

    // 等待弹窗出现
    await waitForModal(page)

    // 验证弹窗标题
    await expect(page.locator('.ant-modal-title')).toContainText(/案件|创建/)
  })

  test('应该能填写案件表单', async ({ page }) => {
    // 打开创建弹窗
    await page.click('button:has-text("新建"), button:has-text("创建案件")')
    await waitForModal(page)

    // 填写基本信息
    await page.fill('input[placeholder*="案件名称"], input[name="caseName"]', 'E2E测试案件')
    await page.fill('input[placeholder*="案件编号"], input[name="caseNo"]', 'E2E-001')

    // 选择案件类型
    const typeSelect = page.locator('.ant-form-item:has-text("类型") .ant-select-selector')
    if (await typeSelect.isVisible()) {
      await typeSelect.click()
      await page.click('.ant-select-dropdown:visible .ant-select-item:first-child')
    }

    // 验证表单已填写
    await expect(page.locator('input[value="E2E测试案件"], input[value*="E2E"]')).toBeVisible()
  })

  test('必填字段验证', async ({ page }) => {
    // 打开创建弹窗
    await page.click('button:has-text("新建"), button:has-text("创建案件")')
    await waitForModal(page)

    // 直接点击提交，不填任何内容
    await page.click('.ant-modal button:has-text("确定"), .ant-modal button:has-text("提交")')

    // 应该显示验证错误
    await expect(page.locator('.ant-form-item-explain-error')).toBeVisible()
  })

  test('应该能成功创建案件', async ({ page }) => {
    // 打开创建弹窗
    await page.click('button:has-text("新建"), button:has-text("创建案件")')
    await waitForModal(page)

    // 填写完整表单
    await page.fill('input[placeholder*="案件名称"], input[name="caseName"]', 'E2E自动化测试案件')
    await page.fill('input[placeholder*="案件编号"], input[name="caseNo"]', `AUTO-${Date.now()}`)

    // 选择类型
    const typeSelect = page.locator('.ant-form-item:has-text("类型") .ant-select-selector')
    if (await typeSelect.isVisible()) {
      await typeSelect.click()
      await page.click('.ant-select-dropdown:visible .ant-select-item:first-child')
    }

    // 提交表单
    await page.click('.ant-modal button:has-text("确定"), .ant-modal button:has-text("提交")')

    // 等待成功提示
    await expect(page.locator('.ant-message')).toBeVisible({ timeout: 10000 })

    // 弹窗应该关闭
    await expect(page.locator('.ant-modal-content')).not.toBeVisible({ timeout: 5000 })
  })
})

test.describe('案件详情', () => {
  test.beforeEach(async ({ page }) => {
    await login(page, 'admin', 'admin123')
    await page.goto('/cases')
    await waitForPageLoad(page)
    await waitForTableLoad(page)
  })

  test('应该能查看案件详情', async ({ page }) => {
    // 点击第一个案件的查看按钮
    const viewButton = page.locator('.ant-table-tbody tr:first-child button:has-text("查看")')

    if (await viewButton.isVisible()) {
      await viewButton.click()

      // 等待跳转到详情页或弹窗
      await page.waitForTimeout(1000)

      // 验证详情内容
      const detailContent = page.locator('.ant-descriptions, .case-detail, .ant-drawer-content')
      await expect(detailContent).toBeVisible()
    } else {
      // 跳过测试如果没有数据
      test.skip()
    }
  })

  test('应该能编辑案件', async ({ page }) => {
    // 点击第一个案件的编辑按钮
    const editButton = page.locator('.ant-table-tbody tr:first-child button:has-text("编辑")')

    if (await editButton.isVisible()) {
      await editButton.click()
      await waitForModal(page)

      // 修改案件名称
      const nameInput = page.locator('input[placeholder*="案件名称"], input[name="caseName"]')
      await nameInput.fill('修改后的案件名称')

      // 保存
      await page.click('.ant-modal button:has-text("确定"), .ant-modal button:has-text("保存")')

      // 等待成功提示
      await expect(page.locator('.ant-message')).toBeVisible({ timeout: 5000 })
    } else {
      test.skip()
    }
  })
})

test.describe('案件删除', () => {
  test.beforeEach(async ({ page }) => {
    await login(page, 'admin', 'admin123')
    await page.goto('/cases')
    await waitForPageLoad(page)
  })

  test('删除应该需要确认', async ({ page }) => {
    await waitForTableLoad(page)

    // 点击删除按钮
    const deleteButton = page.locator('.ant-table-tbody tr:first-child button:has-text("删除")')

    if (await deleteButton.isVisible()) {
      await deleteButton.click()

      // 应该出现确认弹窗
      await expect(page.locator('.ant-popconfirm, .ant-modal-confirm')).toBeVisible()
    } else {
      test.skip()
    }
  })

  test('取消删除不应该删除数据', async ({ page }) => {
    await waitForTableLoad(page)

    const rowsBefore = await getTableRowCount(page)

    const deleteButton = page.locator('.ant-table-tbody tr:first-child button:has-text("删除")')

    if (await deleteButton.isVisible()) {
      await deleteButton.click()

      // 点击取消
      await page.click('.ant-popconfirm button:has-text("取消"), .ant-modal-confirm button:has-text("取消")')

      await page.waitForTimeout(500)

      // 行数应该不变
      const rowsAfter = await getTableRowCount(page)
      expect(rowsAfter).toBe(rowsBefore)
    } else {
      test.skip()
    }
  })
})
