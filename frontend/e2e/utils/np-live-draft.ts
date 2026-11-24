import { expect, type Page } from '@playwright/test'
import { readFileSync } from 'node:fs'

const password = process.env.NP_BROWSER_PASSWORD
if (!password || password.length < 12) throw new Error('NP_BROWSER_PASSWORD must be set for the local fixture')

export const uuid = /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/i
const seedPath = process.env.NP_BROWSER_SEED
if (!seedPath) throw new Error('NP_BROWSER_SEED must point to the generated manifest')
export const seed = JSON.parse(readFileSync(seedPath, 'utf8'))
if (!uuid.test(seed.target) || !uuid.test(seed.control) || seed.target === seed.control ||
    !/^[1-9][0-9]*$/.test(seed.actor_id) || !seed.username || !seed.target_title) {
  throw new Error('invalid synthetic browser manifest')
}

export async function loginAndOpenDraft(page: Page) {
  await page.goto('/login')
  await page.getByPlaceholder('账号或邮箱').fill(seed.username)
  await page.getByPlaceholder('密码').fill(password!)
  await page.locator('button[type=submit]').click()
  await expect(page).toHaveURL(/dashboard/)
  await page.getByRole('button', { name: /新建立案$/ }).click()
  await page.getByRole('button', { name: '打开服务端草稿', exact: true }).click()
  const restored = page.waitForResponse(r => r.request().method() === 'GET' &&
    new URL(r.url()).pathname === '/api/v1/case-intakes/' + seed.target)
  await page.getByRole('button', { name: seed.target_title, exact: true }).click()
  expect((await restored).status()).toBe(200)
  await expect(page.locator('#intake-title')).toHaveValue(seed.target_title)
  await expect(page.locator('#intake-description')).not.toHaveValue('')
}

export async function applyCandidate(page: Page, text: string) {
  await page.getByRole('textbox', { name: '摘要对话输入', exact: true }).fill(text)
  await page.getByRole('button', { name: '开始摘要整理', exact: true }).click()
  await page.getByRole('checkbox', { name: '已核对来源', exact: true }).check()
  await page.getByRole('button', { name: '应用摘要', exact: true }).click()
  await expect(page.locator('#intake-description')).toHaveValue(new RegExp(text))
}
