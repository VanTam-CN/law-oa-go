/**
 * E2E: 当前 MVP 案件流程.
 */

import { test, expect, type Page } from '@playwright/test'
import {
  seedAuthenticatedUser,
  selectIntakeOption,
  waitForAppShell,
  waitForNativeTable,
  waitForPageLoad,
} from './utils/test-helpers'

async function fillCaseIntakeBasics(page: Page) {
  await page.locator('.batch-field', { hasText: '案件名称' }).getByRole('textbox').fill('示例科技服务合同纠纷案')
  await selectIntakeOption(page.locator('.batch-field', { hasText: '案件类型' }).getByRole('combobox'), '商事诉讼')
  await selectIntakeOption(page.locator('.batch-field', { hasText: '业务领域' }).getByRole('combobox'), '公司与并购')
  await selectIntakeOption(page.locator('.batch-field', { hasText: '子领域' }).getByRole('combobox'), '投资与融资')
  await selectIntakeOption(page.locator('article', { hasText: '我方当事人' }).getByRole('combobox'), '上海示例科技有限公司')
  await page.getByRole('textbox', { name: '法定名称或证件姓名' }).fill('上海华信建设集团有限公司')
  await page.getByPlaceholder('输入证件号或统一社会信用代码').fill('91310000TESTCASE0001')
  await page.locator('.batch-wide-label', { hasText: '案情摘要' }).getByRole('textbox').fill('客户拟就服务合同争议提起诉讼。')
  await selectIntakeOption(page.locator('.batch-intake-aside').getByRole('combobox'), '张律师 · 争议解决部')
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
    const intakeWrites: { method: string; url: string; body: any; headers: Record<string, string> }[] = []
    page.on('request', (request) => {
      const url = new URL(request.url())
      if (url.pathname === '/api/v1/case-intakes' && request.method() === 'POST') {
        intakeWrites.push({ method: 'POST', url: url.pathname, body: request.postDataJSON(), headers: request.headers() })
      }
      if (url.pathname === '/api/v1/case-intakes/801' && request.method() === 'PUT') {
        intakeWrites.push({ method: 'PUT', url: url.pathname, body: request.postDataJSON(), headers: request.headers() })
      }
      if (
        (url.pathname === '/api/v1/case-intakes/801/facts-confirmation' ||
          url.pathname === '/api/v1/case-intakes/801/conflict-check') &&
        request.method() === 'POST'
      ) {
        intakeWrites.push({ method: 'POST', url: url.pathname, body: request.postDataJSON(), headers: request.headers() })
      }
    })

    await fillCaseIntakeBasics(page)
    await page.getByRole('button', { name: '保存草稿' }).click()
    await expect(page.locator('.ant-message')).toContainText('接案草稿已创建')
    await page.getByRole('button', { name: '保存草稿' }).click()
    await expect(page.locator('.ant-message')).toContainText('接案草稿已更新')
    // POST creates the row; PUT must carry the exact token returned by the
    // POST receipt, so a concurrent save cannot be clobbered by this tab.
    expect(intakeWrites.filter((write) => write.method === 'POST' && write.url === '/api/v1/case-intakes')).toHaveLength(1)
    const firstPut = intakeWrites.find((write) => write.method === 'PUT')
    expect(firstPut?.body?.expected_revision).toBe('rev-801-1')
    expect(firstPut?.body?.parties).toBeUndefined()
    expect(intakeWrites[0].body.parties.map((party: any) => party.party_role)).toEqual(['client', 'opposing_party'])

    await page.reload()
    await page.getByRole('button', { name: '继续未完成草稿' }).click()
    await expect(page.getByRole('textbox', { name: '案件名称 *' })).toHaveValue('示例科技服务合同纠纷案')
    await expect(page.locator('.batch-field', { hasText: '案件类型' }).getByTitle('商事诉讼')).toBeVisible()
    await expect(page.locator('.batch-field', { hasText: '业务领域' }).getByTitle('公司与并购')).toBeVisible()
    await expect(page.locator('.batch-field', { hasText: '子领域' }).getByTitle('投资与融资')).toBeVisible()
    await expect(page.getByPlaceholder('输入证件号或统一社会信用代码')).toHaveValue('')
    await expect(page.getByText(/恢复后请重新填写/)).toBeVisible()

    // Restoring is read-only until the user explicitly saves.
    expect(intakeWrites.filter((write) => write.method === 'PUT')).toHaveLength(1)
  })

  for (const failure of ['none', 'save', 'result', 'review', 'waiver']) {
    test(`摘要修改后直接审批：${failure}`, async ({ page }) => {
      await fillCaseIntakeBasics(page)
      await page.getByRole('button', { name: '保存并进行利益冲突检查' }).click()
      await expect(page.locator('.ant-message')).toContainText('利益冲突检查已完成')
      await page.getByRole('button', { name: '返回基本信息' }).click()
      const summary = '合成资料：律师核对后的第二版摘要'
      await page.locator('#intake-description').fill(summary)
      if (failure === 'save') await page.route('**/api/v1/case-intakes/801', async route => {
        if (route.request().method() !== 'PUT') return route.fallback()
        await route.fulfill({ status: 409, json: { success: false, error: { message: '合成草稿版本冲突' } } })
      })
      if (['result', 'review', 'waiver'].includes(failure)) {
        await page.route('**/api/v1/conflict/tasks/*/' + failure, route =>
          route.fulfill({ status: 503, json: { success: false, error: { message: '合成读取故障 ' + failure } } }))
      }
      const writes: string[][] = []
      page.on('request', request => {
        if (!['POST', 'PUT'].includes(request.method())) return
        const path = new URL(request.url()).pathname.split('/').pop()!
        const body = request.postDataJSON() || {}
        if (path === '801') writes.push([path, body.expected_revision, body.description])
        if (['facts-confirmation', 'conflict-check'].includes(path)) writes.push([path, request.headers()['x-intake-revision']])
        if (path === 'with-conflict') writes.push([path, body.metadata.intake_revision, body.content])
      })
      await page.getByRole('button', { name: '提交审批并等待成案' }).click()
      if (failure === 'save') {
        await expect(page.locator('.ant-message')).toContainText('合成草稿版本冲突')
        await expect(page.locator('#intake-description')).toHaveValue(summary)
        expect(writes).toEqual([['801', 'rev-801-3', summary]])
      } else if (failure !== 'none') {
        await expect(page.locator('.ant-message')).toContainText('合成读取故障 ' + failure)
        await expect(page.locator('#intake-description')).toHaveValue(summary)
        expect(writes.map(([action]) => action)).toEqual(['801', 'facts-confirmation', 'conflict-check'])
      } else {
        await expect(page).toHaveURL(/\/approval\/701$/)
        expect(writes).toEqual([
          ['801', 'rev-801-3', summary], ['facts-confirmation', 'rev-801-4'],
          ['conflict-check', 'rev-801-5'], ['with-conflict', 'rev-801-6', summary],
        ])
      }
    })
  }

  test('创建重放保留本地输入，重开前不能用服务端新版本继续写', async ({ page }) => {
    let creates = 0
    let writes = 0
    await page.route('**/api/v1/case-intakes', async route => {
      if (route.request().method() !== 'POST') return route.fallback()
      creates++
      await route.fulfill({ json: { success: true, data: {
        id: 801, intake_code: 'IN-2026-001', revision: '10000000-0000-4000-8000-000000000003',
        title: '服务端较新标题', description: '其他标签已保存', status: 'draft',
        created_by: 2, has_source_history: false, idempotent_replay: true,
      } } })
    })
    page.on('request', request => {
      if (request.method() === 'PUT' || request.url().endsWith('/facts-confirmation')) writes++
    })
    await fillCaseIntakeBasics(page)
    await page.getByRole('button', { name: '保存草稿', exact: true }).click()
    await expect(page.locator('.ant-message')).toContainText('此创建请求已有草稿')
    await page.getByRole('button', { name: '保存并进行利益冲突检查' }).click()
    await expect(page.locator('.ant-message')).toContainText('本机草稿缺少版本号')
    await expect(page.locator('#intake-title')).toHaveValue('示例科技服务合同纠纷案')
    expect([creates, writes]).toEqual([1, 0])
  })

  test('缺少版本的旧本机草稿保留输入并阻止原生 PUT', async ({ page }) => {
    await fillCaseIntakeBasics(page)
    await page.getByRole('button', { name: '保存草稿' }).click()
    await expect(page.locator('.ant-message')).toContainText('接案草稿已创建')
    await page.evaluate(() => {
      const key = Object.keys(localStorage).find(item => item.startsWith('law-oa-case-intake-draft-v2:'))
      if (!key) throw new Error('缺少本机草稿')
      const draft = JSON.parse(localStorage.getItem(key)!)
      if (!draft.intakeId || !draft.revision) throw new Error('缺少保存回执')
      delete draft.revision
      localStorage.setItem(key, JSON.stringify(draft))
    })
    await page.reload()
    await page.getByRole('button', { name: '继续未完成草稿' }).click()
    const title = page.getByRole('textbox', { name: '案件名称 *' })
    await title.fill('本机尚未确认的修改')
    let puts = 0
    page.on('request', request => { if (request.method() === 'PUT') puts++ })
    await page.getByRole('button', { name: '保存草稿' }).click()
    await expect(page.locator('.ant-message')).toContainText('本机草稿缺少版本号')
    await expect(title).toHaveValue('本机尚未确认的修改')
    expect(puts).toBe(0)
  })

  test('重复保存后服务端草稿恢复不应回传身份号或覆盖主体与材料', async ({ page }) => {
    const serverDetailBodies: any[] = []
    const savedPutBodies: any[] = []
    page.on('response', async (response) => {
      const url = new URL(response.url())
      if (url.pathname === '/api/v1/case-intakes/801' && response.request().method() === 'GET') {
        serverDetailBodies.push(await response.json().catch(() => ({})))
      }
    })
    page.on('request', (request) => {
      const url = new URL(request.url())
      if (url.pathname === '/api/v1/case-intakes/801' && request.method() === 'PUT') {
        savedPutBodies.push(request.postDataJSON())
      }
    })

    await fillCaseIntakeBasics(page)
    await page.getByRole('button', { name: '保存草稿' }).click()
    await expect(page.locator('.ant-message')).toContainText('接案草稿已创建')

    await page.getByRole('button', { name: '打开服务端草稿' }).click()
    await page.getByRole('button', { name: '示例科技服务合同纠纷案' }).click()
    await expect(page.getByRole('textbox', { name: '案件名称 *' })).toHaveValue('示例科技服务合同纠纷案')

    // The redacted related row keeps its role tag on the basics step but
    // must not carry a readable identity number back through the read
    // boundary; the opponent password field stays empty as before.
    await page.getByRole('button', { name: '切换并恢复所选草稿' }).click()
    await expect(page.locator('.ant-message')).toContainText('已恢复所选服务端草稿')
    await expect(page.getByPlaceholder('输入证件号或统一社会信用代码')).toHaveValue('')

    await expect.poll(() => serverDetailBodies.length).toBeGreaterThanOrEqual(1)
    const detail = serverDetailBodies[serverDetailBodies.length - 1]?.data
    const detailParties = detail?.parties ?? []
    expect(detailParties.map((party: any) => party.party_role)).toEqual(['client', 'opposing_party'])
    expect(JSON.stringify(detailParties)).not.toContain('91310000TESTCASE0001')
    expect(detail?.revision).toBe('rev-801-1')

    // An unedited restore keeps its saved projection, so an ordinary save
    // must not send a partial parties array or reseed the material list.
    await page.getByRole('button', { name: '保存草稿' }).click()
    await expect(page.locator('.ant-message')).toContainText('接案草稿已更新')
    const restoreSave = savedPutBodies.at(-1)
    expect(restoreSave.expected_revision).toBe(detail.revision)
    expect(restoreSave.materials).toBeUndefined()
    expect(restoreSave.parties).toBeUndefined()
    expect(restoreSave?.metadata?.case_tags).toBeUndefined()
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
