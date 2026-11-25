import { expect, test } from '@playwright/test'
import { applyCandidate, loginAndOpenDraft, seed } from '../e2e/utils/np-live-draft'

test('尚无历史来源的首个候选手改后混改标题，保存须停在人工确认', async ({ page }) => {
  // Exercise the first-ever selection while no source history can mask a legacy fallback.
  await loginAndOpenDraft(page)
  await applyCandidate(page, '虚构未确认候选')
  await page.locator('#intake-description').fill('待核对的人工内容')
  await page.locator('#intake-title').fill('待核对的新标题')
  let mixedWrites = 0
  page.on('request', request => {
    if (request.method() === 'PUT' && new URL(request.url()).pathname === '/api/v1/case-intakes/' + seed.target) mixedWrites++
  })
  await page.getByRole('button', { name: '保存草稿', exact: true }).click()
  await expect(page.getByText(/来源已失效，请重新应用候选/).first()).toBeVisible()
  expect(mixedWrites).toBe(0)
})

test('来源草稿重开后直接检测沿用已保存版本，空白证件框不触发重写', async ({ page, browser }) => {
  await loginAndOpenDraft(page)
  await applyCandidate(page, '虚构来源，用于核对重开后的操作')
  const saved = page.waitForResponse(r => r.request().method() === 'PUT' && new URL(r.url()).pathname.endsWith('/' + seed.target))
  await page.getByRole('button', { name: '保存草稿', exact: true }).click()
  const receipt = await saved
  expect(receipt.status()).toBe(200)
  const original = (await receipt.json()).data.revision
  const fresh = await browser.newContext()
  const restored = await fresh.newPage()
  const path = '**/api/v1/case-intakes/' + seed.target
  // Supply the registration status and case classification absent from the four-scenario seed.
  await restored.route(path, async route => {
    if (route.request().method() !== 'GET') return route.continue()
    const response = await route.fetch()
    const payload = await response.json()
    expect(payload.data.has_source_history).toBe(true)
    Object.assign(payload.data.intake.metadata, { business_area: '公司与并购', sub_area: '投资与融资' })
    const opponent = payload.data.opponent_party ?? payload.data.parties.find((party: {party_role: string}) => party.party_role === 'opposing_party')
    expect(opponent).toBeTruthy()
    opponent.identity_status = '已登记（受保护）'
    await route.fulfill({ response, json: payload })
  })
  await loginAndOpenDraft(restored)
  const confirmed = 'f0000000-0000-4000-8000-000000000001'
  const checked = 'f0000000-0000-4000-8000-000000000002'
  const writes: string[][] = []
  restored.on('request', req => {
    if (['PUT', 'POST'].includes(req.method())) {
      writes.push([new URL(req.url()).pathname.split('/').pop()!, req.headers()['x-intake-revision']])
    }
  })
  await restored.route(path + '/facts-confirmation', route => route.fulfill({
    json: { success: true, data: { id: seed.target, revision: confirmed } },
  }))
  await restored.route(path + '/conflict-check', route => route.fulfill({
    json: { success: true, data: { intake_id: seed.target, revision: checked, result: { decision: { status: 'CLEAR' } } } },
  }))
  await expect(restored.getByPlaceholder('输入证件号或统一社会信用代码')).toHaveValue('')
  await restored.getByRole('button', { name: '保存并进行利益冲突检查' }).click()
  await expect(restored.locator('.ant-message')).toContainText('利益冲突检查已完成')
  expect(writes).toEqual([['facts-confirmation', original], ['conflict-check', confirmed]])
  await fresh.close()
})
