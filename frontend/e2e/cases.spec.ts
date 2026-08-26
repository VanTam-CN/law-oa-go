/**
 * E2E: 当前 MVP 案件流程.
 */

import { test, expect } from '@playwright/test'
import {
  seedAuthenticatedUser,
  waitForAppShell,
  waitForNativeTable,
  waitForPageLoad,
} from './utils/test-helpers'

async function fillCaseIntakeBasics(page: any) {
  await page.locator('.batch-field', { hasText: '案件名称' }).getByRole('textbox').fill('示例科技服务合同纠纷案')
  await page.locator('.batch-field', { hasText: '案件类型' }).locator('.ant-select-selector').click()
  await page.getByTitle('商事诉讼').click()
  await page.locator('.batch-field', { hasText: '业务领域' }).locator('.ant-select-selector').click()
  await page.getByTitle('公司与并购').click()
  await page.locator('.batch-field', { hasText: '子领域' }).locator('.ant-select-selector').click()
  await page.getByTitle('投资与融资').click()
  await page.locator('article', { hasText: '我方当事人' }).locator('.ant-select-selector').click()
  await page.getByTitle('上海示例科技有限公司').click()
  await page.getByRole('textbox', { name: '法定名称或证件姓名' }).fill('上海华信建设集团有限公司')
  await page.getByPlaceholder('输入证件号或统一社会信用代码').fill('91310000TESTCASE0001')
  await page.locator('.batch-wide-label', { hasText: '案情摘要' }).getByRole('textbox').fill('客户拟就服务合同争议提起诉讼。')
  await page.locator('.batch-intake-aside').locator('.ant-select-selector').click()
  await page.getByTitle('张律师 · 争议解决部').click()
}

test.describe('案件管理', () => {
  test.beforeEach(async ({ page }) => {
    await seedAuthenticatedUser(page, 'lawyer')
    await page.goto('/case')
    await waitForPageLoad(page)
  })

  test('应该显示案件管理页和数据库案件清单', async ({ page }) => {
    await waitForAppShell(page)
    await expect(page.getByRole('heading', { name: '案件管理' })).toBeVisible()
    await waitForNativeTable(page)
    await expect(page.getByText('DEMO-2026-001')).toBeVisible()
    await expect(page.getByRole('cell', { name: '红杉资本投资管理咨询合同纠纷案' })).toBeVisible()
  })

  test('应该能搜索案件', async ({ page }) => {
    await page.getByPlaceholder('搜索案件编号、客户、对方、负责人...').fill('蓝海')

    await expect(page.getByRole('cell', { name: '蓝海公司股权转让争议' })).toBeVisible()
    await expect(page.getByRole('cell', { name: '红杉资本投资管理咨询合同纠纷案' })).not.toBeVisible()
  })

  test('应该能按状态筛选案件', async ({ page }) => {
    await page.getByRole('button', { name: '审批中' }).click()

    await expect(page.getByRole('cell', { name: '蓝海公司股权转让争议' })).toBeVisible()
    await expect(page.getByRole('cell', { name: '红杉资本投资管理咨询合同纠纷案' })).not.toBeVisible()
  })

  test('点击新建案件应该进入立案工作台', async ({ page }) => {
    await page.getByRole('button', { name: '新建案件' }).click()

    await expect(page).toHaveURL(/\/case\/create$/)
    await expect(page.getByRole('heading', { name: '新建案件立案工作台' })).toBeVisible()
  })

  test('应该能从列表进入案件详情', async ({ page }) => {
    await page.locator('tr', { hasText: 'DEMO-2026-001' }).getByRole('button', { name: /查\s*看/ }).click()

    await expect(page).toHaveURL(/\/case\/101$/)
    await expect(page.getByRole('heading', { name: '红杉资本投资管理咨询合同纠纷案' })).toBeVisible()
  })

  test('在办案件可以提报全新对方并进入核查岗确认状态', async ({ page }) => {
    await page.locator('tr', { hasText: 'DEMO-2026-001' }).getByRole('button', { name: /查\s*看/ }).click()
    await page.getByRole('button', { name: '报告主体变更并重新复核' }).click()
    const dialog = page.getByRole('dialog', { name: '报告案件主体变更' })
    await dialog.getByText('登记全新主体', { exact: true }).click()
    await dialog.getByLabel('主体法定名称或证件姓名 *').fill('虚构启明精密制造有限公司')
    await dialog.getByLabel('曾用名或别名').fill('虚构启明制造')
    await dialog.getByLabel('身份标识 *').fill('91310000TEST00A101')
    await dialog.getByLabel('变更原因 *').fill('法院通知追加该公司为共同被告')
    await dialog.getByRole('button', { name: '提交核查岗确认' }).click()

    await expect(page.locator('.ant-message')).toContainText('等待冲突核查岗确认主体档案')
    await expect(page.getByText('新主体等待核查岗确认，受控动作已暂停')).toBeVisible()
    await expect(page.getByText('无需重复提交')).toBeVisible()
  })

  test('待处理案件详情应提供本案冲突复核下一步', async ({ page }) => {
    await page.locator('tr', { hasText: 'CASE-20260513173242' }).getByRole('button', { name: /查\s*看/ }).click()

    await expect(page).toHaveURL(/\/case\/103$/)
    await expect(page.getByRole('heading', { name: '待处理冲突复核测试案件' })).toBeVisible()
    await expect(page.getByRole('heading', { name: '下一步操作' })).toBeVisible()
    await page.getByRole('button', { name: '进入本案冲突复核' }).click()

    await expect(page).toHaveURL(/\/conflict\?.*case_number=CASE-20260513173242/)
    await expect(page.getByRole('heading', { name: '本案复核上下文' })).toBeVisible()
    await expect(page.getByText(/已匹配到本案冲突检测记录/)).toBeVisible()
    await expect(page.getByText('CASE-20260513173242 待处理冲突复核测试案件')).toBeVisible()
    await expect(page.getByRole('heading', { name: '检测任务清单' })).toBeVisible()
    await expect(page.getByRole('dialog', { name: '冲突检测详情' })).not.toBeVisible()

    await page.getByRole('button', { name: '查看本案检测结果' }).click()
    const detailDialog = page.getByRole('dialog', { name: '冲突检测详情' })
    await expect(detailDialog).toBeVisible()
    await expect(detailDialog.getByRole('cell', { name: '受限历史事项' }).first()).toBeVisible()
    await expect(detailDialog.getByText('关联主体历史委托')).toHaveCount(0)
  })
})

test.describe('新建立案工作台', () => {
  test.beforeEach(async ({ page }) => {
    await seedAuthenticatedUser(page, 'lawyer')
    await page.goto('/case/create')
    await waitForPageLoad(page)
  })

  test('应该显示五步立案流程', async ({ page }) => {
    await waitForAppShell(page)
    await expect(page.getByRole('heading', { name: '新建案件立案工作台' })).toBeVisible()
    const stepper = page.locator('.batch-stepper')
    await expect(stepper.getByRole('button', { name: /基本信息/ })).toBeVisible()
    await expect(stepper.getByRole('button', { name: /利益冲突检查/ })).toBeVisible()
    await expect(stepper.getByRole('button', { name: /团队与费用/ })).toBeVisible()
    await expect(stepper.getByRole('button', { name: /文档与材料/ })).toBeVisible()
    await expect(stepper.getByRole('button', { name: /立案提交/ })).toBeVisible()
  })

  test('初始新建页不应该预填样例案件数据', async ({ page }) => {
    await waitForAppShell(page)
    await expect(page.getByText('红杉资本投资管理咨询合同纠纷案')).toHaveCount(0)
    await expect(page.locator('.batch-field', { hasText: '案件名称' }).getByRole('textbox')).toHaveValue('')
  })

  test('首次新建应显示最小必填路径并隐藏工程诊断', async ({ page }) => {
    await waitForAppShell(page)

    await expect(page.getByText('首次上手最小路径')).toBeVisible()
    await expect(page.getByText(/案件名称、客户、对方当事人、对方身份标识/)).toBeVisible()
    await expect(page.getByText(/先完成上述最小必填项/)).toBeVisible()
    await expect(page.getByText('加载耗时')).toHaveCount(0)
    await expect(page.getByText('接口响应')).toHaveCount(0)
    await expect(page.getByText('数据保存中')).toHaveCount(0)

    await page.getByRole('button', { name: '帮助与支持' }).click()
    await expect(page.getByRole('dialog').getByText('立案帮助与支持')).toBeVisible()
    await expect(page.getByRole('dialog').getByText(/下次进入本页会提示恢复/)).toBeVisible()
    await expect(page.getByRole('dialog').getByText(/联系律所管理员/)).toBeVisible()
  })

  test('应该能暂存接案草稿', async ({ page }) => {
    await fillCaseIntakeBasics(page)
    await page.getByRole('button', { name: '保存草稿' }).click()

    await expect(page.locator('.ant-message')).toContainText('接案草稿已创建')
  })

  test('重复保存应更新同一接案且刷新后可恢复完整草稿', async ({ page }) => {
    let createCount = 0
    let updateCount = 0
    page.on('request', (request) => {
      const path = new URL(request.url()).pathname
      if (path === '/api/v1/case-intakes' && request.method() === 'POST') createCount += 1
      if (path === '/api/v1/case-intakes/801' && request.method() === 'PUT') updateCount += 1
    })

    await fillCaseIntakeBasics(page)
    await page.getByRole('button', { name: '保存草稿' }).click()
    await expect(page.locator('.ant-message')).toContainText('接案草稿已创建')
    await page.getByRole('button', { name: '保存草稿' }).click()
    await expect(page.locator('.ant-message')).toContainText('接案草稿已更新')
    expect(createCount).toBe(1)
    expect(updateCount).toBe(1)

    await page.reload()
    await page.getByRole('button', { name: '继续未完成草稿' }).click()
    await expect(page.getByRole('textbox', { name: '案件名称 *' })).toHaveValue('示例科技服务合同纠纷案')
    await expect(page.locator('.batch-field', { hasText: '案件类型' }).getByTitle('商事诉讼')).toBeVisible()
    await expect(page.locator('.batch-field', { hasText: '业务领域' }).getByTitle('公司与并购')).toBeVisible()
    await expect(page.locator('.batch-field', { hasText: '子领域' }).getByTitle('投资与融资')).toBeVisible()
    await expect(page.getByPlaceholder('输入证件号或统一社会信用代码')).toHaveValue('')
    await expect(page.getByText(/恢复后请重新填写/)).toBeVisible()
  })

  test('必填控件应具备可访问名称并可由标签定位', async ({ page }) => {
    await expect(page.getByRole('textbox', { name: '案件名称 *' })).toBeVisible()
    await expect(page.getByRole('combobox', { name: '案件类型 *' })).toBeVisible()
    await expect(page.getByRole('combobox', { name: '业务领域 *' })).toBeVisible()
    await expect(page.getByRole('combobox', { name: '子领域 *' })).toBeVisible()
    await page.locator('.batch-stepper').getByRole('button', { name: /团队与费用/ }).click()
    await expect(page.getByRole('combobox', { name: '负责律师 *' })).toBeVisible()
  })

  test('应该能运行利益冲突检查并进入团队与费用', async ({ page }) => {
    await fillCaseIntakeBasics(page)
    await page.locator('.batch-stepper').getByRole('button', { name: /利益冲突检查/ }).click()
    await page.getByRole('button', { name: '运行利益冲突检查' }).click()

    await expect(page.locator('.ant-message')).toContainText('利益冲突检查已完成')
    await expect(page.getByRole('button', { name: '进入团队与费用' })).toBeEnabled()
    await page.getByRole('button', { name: '进入团队与费用' }).click()
    await expect(page.getByText('负责律师 *')).toBeVisible()
  })

  test('文档材料归档入口不应该中断当前立案流程', async ({ page }) => {
    await page.locator('.batch-stepper').getByRole('button', { name: /文档与材料/ }).click()
    await page.getByRole('button', { name: '打开文件材料归档' }).click()

    await expect(page).toHaveURL(/\/case\/create$/)
    await expect(page.getByText('文件材料归档暂未开放')).toBeVisible()
  })

  test('提交审批后应该进入审批详情页', async ({ page }) => {
    await fillCaseIntakeBasics(page)
    await page.getByRole('button', { name: '保存并进行利益冲突检查' }).click()
    await expect(page.locator('.ant-message')).toContainText('利益冲突检查已完成')
    await page.getByRole('button', { name: '提交审批并等待成案' }).first().click()

    await expect(page).toHaveURL(/\/approval\/701$/, { timeout: 10000 })
  })
})
