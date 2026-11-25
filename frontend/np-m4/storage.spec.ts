import { expect, test } from '@playwright/test'
import { readFileSync } from 'node:fs'
import { applyCandidate, loginAndOpenDraft, seed } from '../e2e/utils/np-live-draft'

for (const remember of [false, true]) {
  for (const assistant of [true, false]) {
    test(`C01 隔离业务正文，记住我=${remember} 助手=${assistant}`, async ({ page }) => {
      const account = assistant ? seed : JSON.parse(readFileSync(process.env.NP_BROWSER_OTHER_SEED!, 'utf8'))
      const oldKeys = ['law-oa-case-intake-draft-v1', 'law-oa-case-intake-draft-v2:' + account.actor_id + ':new', 'lawoa-intake-prefill-v1']
      await page.goto('/login')
      await page.evaluate(keys => {
        keys.forEach((key, i) => (i === 2 ? sessionStorage : localStorage).setItem(key, 'old-' + i))
      }, oldKeys)
      // Count business access as well as checking the final state. A write followed by delete still fails.
      await page.addInitScript(({ keys, fixtureValues }) => {
        const touched: string[] = []
        Object.assign(window, { npLegacyAccess: touched })
        for (const method of ['getItem', 'setItem', 'removeItem'] as const) {
          const original = Storage.prototype[method]
          Storage.prototype[method] = function (key: string, ...args: string[]) {
            const value = method === 'setItem' ? String(args[0]) : ''
            if (keys.includes(key) || fixtureValues.some(marker => value.includes(marker))) touched.push(method)
            return Reflect.apply(original, this, [key, ...args])
          }
        }
      }, { keys: oldKeys, fixtureValues: ['M4正文', account.target_title, '虚构初始摘要', '虚构对方', 'source_selection', 'conflictFingerprint'] })
      const capability = page.waitForResponse(r => new URL(r.url()).pathname === '/api/v1/case-intake-assist/capabilities')
      await loginAndOpenDraft(page, { manifest: account, remember })
      expect((await (await capability).json()).data.enabled).toBe(assistant)
      const marker = 'M4正文' + Date.now()
      if (assistant) await applyCandidate(page, marker)
      else await expect(page.getByRole('textbox', { name: '摘要对话输入' })).toHaveCount(0)
      await page.locator('#intake-description').fill(marker + '人工')
      await page.locator('#intake-title').fill(marker + '题名')
      // Drive a blur/render, then inspect both stores without reading the monitored keys in the app.
      await page.locator('#intake-description').focus()
      const observed = await page.evaluate(({ keys, marker }) => ({
        access: [...(window as Window & { npLegacyAccess: string[] }).npLegacyAccess],
        legacyUnchanged: keys.every((key, i) => (i === 2 ? sessionStorage : localStorage)[key] === 'old-' + i),
        hasText: [localStorage, sessionStorage].some(store => Object.values(store).some(value => String(value).includes(marker))),
        persistentLogin: Boolean(localStorage.auth_token),
        sessionLogin: Boolean(sessionStorage.auth_token),
      }), { keys: oldKeys, marker })
      expect(observed).toEqual({ access: [], legacyUnchanged: true, hasText: false,
        persistentLogin: remember, sessionLogin: !remember })
    })
  }
}

test('C06 恢复键写入遭拒时保留编辑，原生入口不发出保存请求', async ({ page }) => {
  await loginAndOpenDraft(page)
  const marker = '合成配额检查'
  await page.locator('#intake-description').fill(marker)
  let writes = 0
  page.on('request', request => {
    if (['PUT', 'POST', 'PATCH'].includes(request.method()) && new URL(request.url()).pathname.startsWith('/api/v1/case-intakes')) writes++
  })
  await page.evaluate(() => {
    const set = Storage.prototype.setItem
    Storage.prototype.setItem = function (key, value) {
      if (this === localStorage && key.startsWith('np-save-v2:')) throw new DOMException('fixture', 'QuotaExceededError')
      return set.call(this, key, value)
    }
  })
  await page.getByRole('button', { name: '保存草稿', exact: true }).click()
  await expect(page.getByText(/恢复标识无法写入/).first()).toBeVisible()
  await expect(page.locator('#intake-description')).toHaveValue(marker)
  expect(writes).toBe(0)
})
