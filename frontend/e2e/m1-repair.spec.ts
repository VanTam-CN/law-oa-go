/**
 * M1 修复回归：恢复后主体保全（仅虚构数据，浏览器内 mock API）。
 * - 有版本本机恢复：重填身份后保存不覆盖其余主体。
 * - 服务端 20 行投影恢复：保存不整体覆盖主体行。
 * 浏览器从模拟 API 回执取得版本，数据库保全由独立 PG 用例验证。
 */

import { test, expect, type Page } from '@playwright/test'
import {
  seedAuthenticatedUser,
  selectIntakeOption,
  waitForPageLoad,
} from './utils/test-helpers'

// The isolated Go/Chrome test supplies a loopback approval endpoint and fixture IDs.
const nativeApproval = process.env.NP_M1_BROWSER_FIXTURE
  ? JSON.parse(process.env.NP_M1_BROWSER_FIXTURE) as { url: string; intake: string; client: number; check: string }
  : undefined
const INTAKE_ID = nativeApproval?.intake ?? 801
const checkID = nativeApproval?.check ?? 'CHECK-801'
const OPPONENT_NAME = '上海华信建设集团有限公司'
const ORIGINAL_IDENTITY = '91310000TESTCASE0001'
const UPDATED_IDENTITY = '91310000TESTCASE0002'

type Party = Record<string, any>

function makeParty(role: string, name: string, extra: Party = {}): Party {
  return { party_role: role, entity_name: name, entity_type: 'LEGAL_PERSON', identity_type: 'SOCIAL_CREDIT_CODE', ...extra }
}

function truncatedParties(): Party[] {
  const rows: Party[] = [makeParty('client', '上海示例科技有限公司'), makeParty('opposing_party', OPPONENT_NAME)]
  for (let index = 2; index < 20; index += 1) {
    rows.push(makeParty('related_party', `虚构关联主体${index}`))
  }
  return rows
}

function detailPayload(revision: string, parties: Party[]) {
  return {
    intake: {
      id: INTAKE_ID,
      intake_code: 'IN-2026-001',
      title: '示例科技服务合同纠纷案',
      case_type: 'commercial',
      priority: 'medium',
      description: '客户拟就服务合同争议提起诉讼。',
      client_id: nativeApproval?.client ?? 1,
      status: 'draft',
      metadata: { business_area: '公司与并购', sub_area: '投资与融资', lawyer_id: 2 },
    },
    revision,
    parties,
    materials: [],
  }
}

async function fillIntakeBasics(page: Page) {
  await page.locator('.batch-field', { hasText: '案件名称' }).getByRole('textbox').fill('示例科技服务合同纠纷案')
  await selectIntakeOption(page.locator('.batch-field', { hasText: '案件类型' }).getByRole('combobox'), '商事诉讼')
  await selectIntakeOption(page.locator('.batch-field', { hasText: '业务领域' }).getByRole('combobox'), '公司与并购')
  await selectIntakeOption(page.locator('.batch-field', { hasText: '子领域' }).getByRole('combobox'), '投资与融资')
  await selectIntakeOption(page.locator('article', { hasText: '我方当事人' }).getByRole('combobox'), '上海示例科技有限公司')
  await page.getByRole('textbox', { name: '法定名称或证件姓名' }).fill(OPPONENT_NAME)
  await page.getByPlaceholder('输入证件号或统一社会信用代码').fill(ORIGINAL_IDENTITY)
  await page.locator('.batch-wide-label', { hasText: '案情摘要' }).getByRole('textbox').fill('客户拟就服务合同争议提起诉讼。')
  await selectIntakeOption(page.locator('.batch-intake-aside').getByRole('combobox'), '张律师 · 争议解决部')
}

async function fillUpdatedOpponentIdentity(page: Page) {
  const identityInput = page.getByPlaceholder('输入证件号或统一社会信用代码')
  await expect(identityInput).toHaveValue('')
  await identityInput.fill(UPDATED_IDENTITY)
}

async function runFactsConfirmation(page: Page) {
  await page.locator('.batch-stepper').getByRole('button', { name: /利益冲突检查/ }).click()
  await page.getByRole('button', { name: '运行利益冲突检查' }).click()
  await expect(page.locator('.ant-message')).toContainText('利益冲突检查已完成')
  await expect(page.getByRole('button', { name: '进入团队与费用' })).toBeEnabled()
}

for (const source of ['本机草稿', '服务端20行投影', '本机别名', '投影外对方', '检测后保存', '投影对方有歧义', '本机版本已过期', '本机未保存的对方名称', '检测回执晚于新保存'] as const) {
  test(`${source}恢复后以原生 PUT 更新指定对方并确认检测`, async ({ page }) => {
    await seedAuthenticatedUser(page, 'lawyer')
    let approvalID = '701'
    if (nativeApproval) {
      const endpoint = new URL(nativeApproval.url)
      expect(endpoint.hostname).toBe('127.0.0.1')
      await page.route(url => url.pathname === '/api/v1/clients', route => route.fulfill({
        json: { success: true, data: { list: [{ id: nativeApproval.client, name: '上海示例科技有限公司' }], total: 1 } },
      }))
      await page.route(url => url.pathname === endpoint.pathname, async route => {
        const response = await route.fetch({ url: endpoint.href, maxRedirects: 0 })
        expect(response.status()).toBe(200)
        approvalID = (await response.json()).data.approval_id
        await route.fulfill({ response })
      })
    }
    const path = '/api/v1/case-intakes'
    const detail = path + '/' + INTAKE_ID
    const writes: Record<string, any>[] = []
    const reads: string[] = []
    const headers: string[] = []
    let releaseCheck!: () => void
    const delayedCheck = new Promise<void>(resolve => { releaseCheck = resolve })
    const local = source.startsWith('本机')
    const alias = 'SYNTHETIC-EDIT-ONLY'
    const revision = local ? 'created801' : 'restored801'
    const ok = (data: unknown) => ({ json: { success: true, data } })
    await page.route(url => url.pathname === path || url.pathname.startsWith(detail), async route => {
      const req = route.request()
      const url = new URL(req.url()).pathname
      if (req.method() === 'GET' && url === path) {
        return route.fulfill(ok({ items: [{ id: INTAKE_ID, title: '打开合成样例' }], total: 1 }))
      }
      if (req.method() === 'GET' && url === detail) {
        reads.push(url)
        const parties = truncatedParties()
        const opponent = parties[1]
        if (source === '投影外对方') parties[1] = makeParty('related_party', '合成关联人')
        const allRelated = source === '本机草稿'
          ? [...parties.slice(2), makeParty('related_party', '投影外合成关联方')]
          : undefined
        return route.fulfill(ok({
          ...detailPayload(source === '本机版本已过期' ? 'newer' : revision, parties),
          opponent_party: source === '投影对方有歧义' ? null : opponent,
          related_parties: allRelated,
        }))
      }
      if (req.method() === 'POST' && url === path) {
        expect(req.postDataJSON().parties.find((p: Party) => p.party_role === 'opposing_party').identity_number).toBe(ORIGINAL_IDENTITY)
        return route.fulfill(ok({ id: INTAKE_ID, revision, intake_code: 'IN-801' }))
      }
      if (req.method() === 'PUT' && url === detail) {
        writes.push(req.postDataJSON())
        return route.fulfill(ok({ id: INTAKE_ID, revision: 'saved-' + writes.length, status: 'draft', intake_code: 'IN-801' }))
      }
      headers.push(req.headers()['x-intake-revision'])
      if (url === detail + '/facts-confirmation') {
        return route.fulfill(ok({ id: INTAKE_ID, revision: 'confirmed', status: 'lawyer_facts_confirmed' }))
      }
      expect(url).toBe(detail + '/conflict-check')
      if (source === '检测回执晚于新保存') await delayedCheck
      return route.fulfill(ok({ taskId: checkID, status: 'COMPLETED', revision: 'checked',
        result: { checkId: checkID, hasConflict: false, conflictCases: [], riskAssessment: { overallRisk: 'LOW' },
          decision: { status: 'CLEAR' }, coverage_status: 'COMPLETE' } }))
    })
    await page.goto('/case/create')
    await waitForPageLoad(page)
    if (local) {
      await fillIntakeBasics(page)
      await page.getByRole('button', { name: '保存草稿', exact: true }).click()
      await expect(page.locator('.ant-message')).toContainText('接案草稿已创建')
      if (source === '本机未保存的对方名称') {
        await page.getByRole('textbox', { name: '法定名称或证件姓名' }).fill('未保存的合成主体')
        await expect.poll(() => page.evaluate(() => Object.values(localStorage).some(v => v.includes('未保存的合成主体')))).toBe(true)
      }
      if (source === '本机别名') {
        await page.locator('#intake-opponent-aliases').fill(alias)
        await expect.poll(() => page.evaluate(value => Object.values(localStorage).some(v => v.includes(value)), alias)).toBe(true)
      }
      await page.reload()
      await page.getByRole('button', { name: '继续未完成草稿' }).click()
    } else {
      await page.getByRole('button', { name: '打开服务端草稿' }).click()
      await page.getByRole('button', { name: '打开合成样例' }).click()
    }
    await expect(page.getByRole('textbox', { name: '案件名称 *' })).toHaveValue('示例科技服务合同纠纷案')
    if (source === '投影对方有歧义') {
      await expect(page.getByRole('textbox', { name: '法定名称或证件姓名' })).toHaveValue('')
      expect(writes).toHaveLength(0)
      expect(headers).toHaveLength(0)
      return
    }
    if (source === '本机版本已过期' || source === '本机未保存的对方名称') {
      const changedName = source === '本机未保存的对方名称'
      await expect(page.getByRole('textbox', { name: '法定名称或证件姓名' })).toHaveValue(changedName ? '未保存的合成主体' : OPPONENT_NAME)
      await fillUpdatedOpponentIdentity(page)
      await page.getByRole('button', { name: '保存草稿', exact: true }).click()
      await expect(page.locator('.ant-message')).toContainText(changedName ? '无法保存主体变更' : '核对对象和版本')
      expect(writes).toHaveLength(0)
      expect(headers).toHaveLength(0)
      return
    }
    await expect(page.getByRole('textbox', { name: '法定名称或证件姓名' })).toHaveValue(OPPONENT_NAME)
    if (source === '本机别名') await expect(page.locator('#intake-opponent-aliases')).toHaveValue(alias)
    await fillUpdatedOpponentIdentity(page)
    await page.getByRole('button', { name: '保存草稿', exact: true }).click()
    await expect(page.locator('.ant-message')).toContainText('接案草稿已更新')
    if (source === '检测回执晚于新保存') {
      await page.locator('.batch-stepper').getByRole('button', { name: /利益冲突检查/ }).click()
      await page.getByRole('button', { name: '运行利益冲突检查' }).click()
      await expect.poll(() => headers.length).toBe(2)
      await page.getByRole('button', { name: '保存草稿', exact: true }).click()
      await expect.poll(() => writes.length).toBe(3)
      const response = page.waitForResponse(r => new URL(r.url()).pathname === detail + '/conflict-check')
      releaseCheck()
      await response
      await expect(page.getByRole('button', { name: '运行利益冲突检查' })).toBeEnabled()
      await expect(page.locator('.batch-bottom-bar')).toContainText('未检测')
      await expect(page.getByRole('button', { name: '进入团队与费用' })).toBeDisabled()
      return
    }
    await runFactsConfirmation(page)
    expect(writes).toHaveLength(2)
    expect(writes[0].expected_revision).toBe(revision)
    expect(writes[0].opponent_identity_update).toEqual({ entity_name: OPPONENT_NAME, entity_type: 'LEGAL_PERSON',
      identity_type: 'SOCIAL_CREDIT_CODE', identity_number: UPDATED_IDENTITY, ...(source === '本机别名' ? { aliases: alias } : {}) })
    expect(writes[0].parties).toBeUndefined()
    expect(writes[1].expected_revision).toBe('saved-1')
    expect(writes[1].parties).toBeUndefined()
    expect(writes[1].opponent_identity_update).toBeUndefined()
    expect(headers).toEqual(['saved-2', 'confirmed'])
    if (!local) expect(reads).toHaveLength(1)
    if (source === '本机草稿') {
      const posted = page.waitForRequest(req => req.method() === 'POST' &&
        new URL(req.url()).pathname === '/api/v1/integration/approvals/with-conflict')
      await page.getByRole('button', { name: '提交审批并等待成案' }).first().click()
      const { metadata } = (await posted).postDataJSON()
      expect(metadata.intake_revision).toBe('checked')
      const names = metadata.parties.map((party: { name: string }) => party.name)
      expect(names).toEqual([...truncatedParties().map(party => party.entity_name), '投影外合成关联方'])
      await expect(page).toHaveURL(url => url.pathname === '/approval/' + approvalID)
    }

    if (source === '检测后保存') {
      await page.getByRole('button', { name: '保存草稿', exact: true }).click()
      await expect.poll(() => writes.length).toBe(3)
      expect(writes[2].expected_revision).toBe('checked')
      await expect(page.locator('.batch-bottom-bar')).toContainText('未检测')
      await expect(page.getByRole('button', { name: '进入团队与费用' })).toBeDisabled()
    }
  })
}
