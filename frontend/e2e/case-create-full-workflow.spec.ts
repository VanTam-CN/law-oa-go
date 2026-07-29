/**
 * E2E: 律师完整立案工作流（QA-RC-001 + IB-RT-001 回归）.
 */

import { test, expect } from '@playwright/test'
import { seedAuthenticatedUser, waitForAppShell, waitForPageLoad } from './utils/test-helpers'

test.describe('立案工作流：冲突检查留在当前上下文', () => {
  test.beforeEach(async ({ page }) => {
    await seedAuthenticatedUser(page, 'lawyer')
    await page.goto('/case/create')
    await waitForPageLoad(page)
    await waitForAppShell(page)
  })

  test('冲突检查后应留在 /case/create 并可进入团队与费用', async ({ page }) => {
    await expect(page.getByRole('heading', { name: '新建案件立案工作台' })).toBeVisible()

    const stepper = page.locator('.batch-stepper button')

    // Step 0 基本信息：案件名称
    const nameField = page.locator('.batch-field').filter({ hasText: /^案件名称/ }).locator('input')
    await nameField.fill('E2E复验测试案件')

    // Step 0 基本信息：案件类型
    const caseTypeField = page.locator('.batch-field').filter({ hasText: /^案件类型/ })
    await caseTypeField.locator('.ant-select-selector').click()
    await page.getByTitle('商事诉讼').click()

    const businessAreaField = page.locator('.batch-field').filter({ hasText: /^业务领域/ })
    await businessAreaField.locator('.ant-select-selector').click()
    await page.getByTitle('公司与并购').click()

    const subAreaField = page.locator('.batch-field').filter({ hasText: /^子领域/ })
    await subAreaField.locator('.ant-select-selector').click()
    await page.getByTitle('投资与融资').click()

    // Step 0 当事人：选择客户
    await page.locator('.batch-party-card.green .ant-select-selector').click()
    await page.getByTitle('上海示例科技有限公司').click()

    // Step 0 当事人：对方当事人
    await page.getByPlaceholder('输入对方当事人名称').fill('测试对方当事人')

    // 负责律师在 step 2（团队与费用），通过 stepper 跳到 step 2
    await stepper.nth(2).click()
    await expect(page.getByRole('heading', { name: '团队与费用' })).toBeVisible()

    // Step 2：选择负责律师
    const lawyerField = page.locator('.batch-field').filter({ hasText: /^负责律师/ })
    await lawyerField.locator('.ant-select-selector').click()
    await page.getByTitle(/张律师/).first().click()

    // 通过 stepper 回到 step 1（利益冲突检查）
    await stepper.nth(1).click()
    await expect(page.getByRole('heading', { name: '利益冲突检查' })).toBeVisible()

    // 运行冲突检查
    await page.getByRole('button', { name: '运行利益冲突检查' }).click()

    // 等待 API mock 响应
    await expect(page.getByText('利益冲突检查已完成，当前草稿：E2E复验测试案件')).toBeVisible({ timeout: 10000 })

    // 关键断言：URL 仍是 /case/create
    await expect(page).toHaveURL(/\/case\/create$/)

    // 关键断言：冲突状态显示已完成
    await expect(page.getByText(/检查状态：已完成/)).toBeVisible()

    // 关键断言：进入团队与费用按钮可点击
    const teamButton = page.getByRole('button', { name: '进入团队与费用' })
    await expect(teamButton).toBeEnabled()

    // 修改冲突检索关键输入后，旧结果必须立即过期，不能继续沿用。
    await page.getByRole('button', { name: '返回基本信息' }).click()
    await page.getByPlaceholder('输入对方当事人名称').fill('变更后的对方当事人')
    await stepper.nth(1).click()
    await expect(page.getByText('检查状态：冲突检测结果已过期')).toBeVisible()
    await expect(page.getByText('客户、对方、相关方、案件或负责律师已变化，请保存最新输入并重新检测。', { exact: true }).first()).toBeVisible()
    await expect(page.getByRole('button', { name: '进入团队与费用' })).toBeDisabled()
  })

  test('必填项缺失时运行冲突检查不得创建草稿', async ({ page }) => {
    await expect(page.getByRole('heading', { name: '新建案件立案工作台' })).toBeVisible()

    const stepper = page.locator('.batch-stepper button')

    // 只选择客户，不填案件名称和对方当事人
    await page.locator('.batch-party-card.green .ant-select-selector').click()
    await page.getByTitle('上海示例科技有限公司').click()

    // 选择负责律师
    await stepper.nth(2).click()
    await expect(page.getByRole('heading', { name: '团队与费用' })).toBeVisible()
    const lawyerField = page.locator('.batch-field').filter({ hasText: /^负责律师/ })
    await lawyerField.locator('.ant-select-selector').click()
    await page.getByTitle(/张律师/).first().click()

    // 回到冲突检查
    await stepper.nth(1).click()
    await expect(page.getByRole('heading', { name: '利益冲突检查' })).toBeVisible()

    // 点击运行冲突检查
    await page.getByRole('button', { name: '运行利益冲突检查' }).click()

    // 应显示校验错误
    await expect(page.getByText(/必填项未完成/)).toBeVisible()

    // 不应创建草稿
    await expect(page.getByText(/接案草稿已创建/)).not.toBeVisible()

    // URL 应仍在 /case/create
    await expect(page).toHaveURL(/\/case\/create$/)
  })
})
