import { expect, test } from '@playwright/test'
import { loginAndOpenDraft, seed } from '../e2e/utils/np-live-draft'

test('C12 生成失败后人工编辑仍只留在内存', async ({ page }) => {
  const marker = 'C12合成内容' + Date.now()
  await page.addInitScript(({ marker, draftTitle }) => {
    const writes: string[] = []
    Object.assign(window, { npSensitiveWrites: writes })
    const put = Storage.prototype.setItem
    Storage.prototype.setItem = function (key: string, value: string) {
      if ([marker, draftTitle, '虚构初始摘要', 'source_selection'].some(word => (key + value).includes(word)) ||
          /law-oa-case-intake-draft|lawoa-intake-prefill/.test(key)) writes.push('business write')
      return put.call(this, key, value)
    }
  }, { marker, draftTitle: seed.target_title })
  await loginAndOpenDraft(page)
  let failedTurns = 0
  await page.route('**/api/v1/case-intake-assist/contexts/*/turns', async route => {
    expect(route.request().method()).toBe('POST')
    expect(route.request().postDataJSON().text).toBe(marker + '口述')
    failedTurns++
    await route.fulfill({ status: 503, contentType: 'application/json', json: {
      success: false, message: '合成生成不可用', error: { code: 'GENERATION_UNAVAILABLE', message: '合成生成不可用' },
    } })
  })
  const dialog = page.getByRole('textbox', {name: '摘要对话输入', exact: true})
  await dialog.fill(marker + '口述')
  const context = page.waitForResponse(r => r.request().method() === 'POST' &&
    new URL(r.url()).pathname === '/api/v1/case-intake-assist/contexts')
  await page.getByRole('button', { name: '开始摘要整理', exact: true }).click()
  expect((await context).status()).toBe(200)
  await expect(page.getByText('合成生成不可用', {exact: true})).toBeVisible()
  await expect(page.getByRole('checkbox', {name:'已核对来源',exact:true})).toHaveCount(0)
  const description = page.locator('#intake-description')
  await description.fill(marker + '手工摘要')
  await page.locator('#intake-title').fill(marker + '手工标题')
  await description.focus()
  await expect(description).toHaveValue(marker + '手工摘要')
  expect(failedTurns).toBe(1)
  expect(await page.evaluate(word => {
    const hasText = (store: Storage) => Object.entries(store).some(([k,v]) => (k + v).includes(word))
    return {
      writes: (window as Window & {npSensitiveWrites: string[]}).npSensitiveWrites.length,
      local: hasText(localStorage), session: hasText(sessionStorage),
    }
  }, marker)).toEqual({ writes: 0, local: false, session: false })
})
