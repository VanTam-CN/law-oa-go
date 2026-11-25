import { expect, test, type Page } from '@playwright/test'
import { readFileSync } from 'node:fs'
import { loginAndOpenDraft, seed, uuid } from '../e2e/utils/np-live-draft'

async function logout(page: Page) {
  await page.getByRole('button', { name: /^用户菜单/ }).click()
  await page.getByRole('menuitem', { name: '退出登录' }).click()
  await expect(page).toHaveURL(/\/login$/)
}

test('C08 换律师登录隔离来源与未决标识，原律师可再登录回查', async ({ page }) => {
  const other = JSON.parse(readFileSync(process.env.NP_BROWSER_OTHER_SEED!, 'utf8'))
  await loginAndOpenDraft(page)
  const text = 'C08合成摘要' + Date.now()
  await page.locator('#intake-description').fill(text)
  let key = ''
  await page.route('**/api/v1/case-intakes/' + seed.target, async route => {
    if (route.request().method() !== 'PUT') return route.continue()
    key = route.request().headers()['idempotency-key']
    await route.continue({ headers: { ...route.request().headers(), 'X-NP-Drop-Response': 'after-commit' } })
  })
  await page.getByRole('button', { name: '保存草稿', exact: true }).click()
  await expect(page.getByText(/原保存结果待核实；不能换键重发/).first()).toBeVisible()
  expect(key).toMatch(uuid)
  const record = 'np-save-v2:' + seed.actor_id + ':' + seed.target + ':' + key
  const stored = await page.evaluate(k => localStorage.getItem(k), record)
  expect(JSON.parse(stored!)).toEqual({ request_id: key, action: 'update', intake_id: seed.target })

  let confirmations = 0
  page.on('dialog', async dialog => {
    expect(dialog.type()).toBe('confirm')
    if (++confirmations === 1) await dialog.dismiss()
    else await dialog.accept()
  })
  await page.getByRole('button', { name: /^用户菜单/ }).click()
  await page.getByRole('menuitem', { name: '退出登录' }).click()
  await expect.poll(() => confirmations).toBe(1)
  await expect(page.locator('#intake-description')).toHaveValue(text)
  expect(await page.evaluate(k => localStorage.getItem(k), record)).toBe(stored)
  await logout(page)
  expect(confirmations).toBe(2)
  await loginAndOpenDraft(page, { manifest: other })
  await expect(page.getByRole('button', { name: '回查原保存结果' })).toHaveCount(0)
  await expect(page.locator('#intake-title')).toHaveValue(other.target_title)
  await expect(page.locator('#intake-description')).not.toHaveValue(text)
  const access = await page.evaluate(async ({ id, requestID }) => {
    const headers = { Authorization: 'Bearer ' + sessionStorage.getItem('auth_token') }
    const detail = await fetch('/api/v1/case-intakes/' + id, { headers })
    const receipt = await fetch('/api/v1/case-intakes/write-results/' + requestID, { headers })
    const data = (await receipt.json()).data
    return { detail: detail.status, receipt: receipt.status, state: data?.state, revealsOwner: Boolean(data?.id || data?.revision) }
  }, { id: seed.target, requestID: key })
  expect([403, 404]).toContain(access.detail)
  expect(access.receipt === 403 || access.receipt === 404 || (access.receipt === 200 && access.state === 'unknown')).toBe(true)
  expect(access.revealsOwner).toBe(false)
  expect(await page.evaluate(k => localStorage.getItem(k), record)).toBe(stored)
  await logout(page)
  await loginAndOpenDraft(page)
  const receipt = page.waitForResponse(r => r.url().endsWith('/write-results/' + key))
  await page.getByRole('button', { name: '回查原保存结果' }).click()
  expect((await (await receipt).json()).data).toMatchObject({ request_id: key, id: seed.target, state: 'committed' })
  await expect.poll(() => page.evaluate(k => localStorage.getItem(k), record)).toBeNull()
  await expect(page.locator('#intake-description')).toHaveValue(text)
})

test('C09 合成旧标签先核对仅内存与预填，再关闭并定向清理已确认的键', async ({ context, page }) => {
  const old = await context.newPage()
  // Inert legacy document under the same origin. It has no save or background writer.
  await old.route('**/__legacy_intake', route => route.fulfill({ contentType: 'text/html', body: '<textarea aria-label="旧页未保存文字"></textarea>' }))
  await old.goto('/__legacy_intake')
  await old.getByRole('textbox').fill('仅在旧标签内存的合成输入')
  const own = 'law-oa-case-intake-draft-v2:' + seed.actor_id + ':new'
  const foreign = 'law-oa-case-intake-draft-v2:other:new'
  const requestId = crypto.randomUUID()
  const pending = 'np-save-v2:other:' + seed.control + ':' + requestId
  await old.evaluate(({ own, foreign, pending, requestId, id }) => {
    localStorage.setItem('law-oa-case-intake-draft-v1', '归属待核对')
    localStorage.setItem(own, '{"conflictFingerprint":"合成旧指纹"}')
    localStorage.setItem(foreign, '另一账号的合成资料')
    localStorage.setItem(pending, JSON.stringify({ request_id: requestId, intake_id: id, action: 'update' }))
    sessionStorage.setItem('lawoa-intake-prefill-v1', '仅在旧标签的合成预填')
  }, { own, foreign, pending, requestId, id: seed.control })
  await loginAndOpenDraft(page)
  await expect(old.getByRole('textbox')).toHaveValue('仅在旧标签内存的合成输入')
  expect(await old.evaluate(() => sessionStorage.getItem('lawoa-intake-prefill-v1'))).toBe('仅在旧标签的合成预填')
  expect(await page.evaluate(() => sessionStorage.getItem('lawoa-intake-prefill-v1'))).toBeNull()
  expect(await page.evaluate(k => localStorage.getItem(k), own)).toContain('合成旧指纹')
  // Fixture operator explicitly discards this tab's verified input; the unidentified owner's items stay on hold.
  await old.close()
  await page.evaluate(k => localStorage.removeItem(k), own)
  expect(await page.evaluate(({ own, foreign, pending }) => [localStorage.getItem(own), Boolean(localStorage.getItem(foreign)),
    Boolean(localStorage.getItem(pending)), Boolean(localStorage.getItem('law-oa-case-intake-draft-v1'))], { own, foreign, pending }))
    .toEqual([null, true, true, true])
})
