import { expect, test, type Page } from '@playwright/test'
import { readFileSync } from 'node:fs'

const password = process.env.NP_BROWSER_PASSWORD
if (!password || password.length < 12) throw new Error('NP_BROWSER_PASSWORD must be set for the local fixture')

const uuid = /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/i
const seedPath = process.env.NP_BROWSER_SEED
if (!seedPath) throw new Error('NP_BROWSER_SEED must point to the generated manifest')
const seed = JSON.parse(readFileSync(seedPath, 'utf8'))
if (!uuid.test(seed.target) || !uuid.test(seed.control) || seed.target === seed.control ||
    !/^[1-9][0-9]*$/.test(seed.actor_id) || !seed.username || !seed.target_title) {
  throw new Error('invalid synthetic browser manifest')
}

async function loginAndOpenDraft(page: Page) {
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

async function applyCandidate(page: Page, text: string) {
  await page.getByRole('textbox', { name: '摘要对话输入', exact: true }).fill(text)
  await page.getByRole('button', { name: '开始摘要整理', exact: true }).click()
  await page.getByRole('checkbox', { name: '已核对来源', exact: true }).check()
  await page.getByRole('button', { name: '应用摘要', exact: true }).click()
  await expect(page.locator('#intake-description')).toHaveValue(new RegExp(text))
}

async function saveResponse(page: Page) {
  const response = page.waitForResponse(r => r.request().method() === 'PUT' && r.url().endsWith('/case-intakes/' + seed.target))
  await page.getByRole('button', { name: '保存草稿', exact: true }).click()
  return response
}

test('真实服务：多轮核对，仅应用摘要，原生保存后新会话从原菜单重开来源', async ({ page, browser }) => {
  await loginAndOpenDraft(page)
  const originalTitle = await page.locator('#intake-title').inputValue()
  await page.getByRole('textbox', { name: '摘要对话输入', exact: true }).fill('这是虚构资料。因对方未按时交货，我方要求确认交付时间。')
  await page.getByRole('button', { name: '开始摘要整理', exact: true }).click()
  await expect(page.getByRole('button', { name: '生成摘要候选', exact: true })).toBeVisible()
  await page.getByRole('textbox', { name: '摘要对话输入', exact: true }).fill('补充：交付日期尚未核对，不能写成已确认。')
  await page.getByRole('button', { name: '生成摘要候选', exact: true }).click()
  await page.getByRole('checkbox', { name: '已核对来源', exact: true }).check()
  await page.getByRole('button', { name: '应用摘要', exact: true }).click()
  await expect(page.locator('#intake-description')).toHaveValue(/交付日期尚未核对/)
  const applied = await page.locator('#intake-description').inputValue()
  await expect(page.locator('#intake-title')).toHaveValue(originalTitle)
  const response = await saveResponse(page)
  expect(response.status()).toBe(200)
  const body = response.request().postDataJSON()
  expect(body.changed_fields).toEqual(['description'])
  expect(Object.keys(body).sort()).toEqual(['changed_fields', 'description', 'expected_revision', 'source_selection'])
  expect(response.request().headers()['idempotency-key']).toBeTruthy()
  const fresh = await browser.newContext()
  const reopened = await fresh.newPage()
  await loginAndOpenDraft(reopened)
  await expect(reopened.locator('#intake-description')).toHaveValue(applied)
  await expect(reopened.locator('#intake-title')).toHaveValue(originalTitle)
  await expect(reopened.getByText('已保存摘要来源', { exact: true })).toBeVisible()
  await expect(reopened.getByText('补充：交付日期尚未核对，不能写成已确认。', { exact: true }).first()).toBeVisible()
  await fresh.close()
})

test('真实服务：A来源保存后，B旧版纯手工摘要不得覆盖并保留输入', async ({ page, browser }) => {
  await loginAndOpenDraft(page)
  const other = await browser.newContext(); const b = await other.newPage()
  await loginAndOpenDraft(b)
  const manual = 'B标签人工输入应保留 ' + Date.now()
  await b.locator('#intake-description').fill(manual)
  const input = '虚构并发来源 ' + Date.now()
  await applyCandidate(page, input)
  expect((await saveResponse(page)).status()).toBe(200)
  const rejected = await saveResponse(b)
  expect(rejected.status()).toBe(409)
  expect(rejected.request().postDataJSON().source_selection).toBeUndefined()
  expect((await rejected.json()).error.code).toBe('INTAKE_DRAFT_REVISION_CONFLICT')
  await expect(b.locator('#intake-description')).toHaveValue(manual)
  await expect(page.locator('#intake-description')).toHaveValue(new RegExp(input))
  await other.close()
})

test('真实上游断线：提交后断开TCP，重登按原键回查且重开来源', async ({ page, browser }) => {
  await loginAndOpenDraft(page)
  const input = '虚构断线来源 ' + Date.now()
  await applyCandidate(page, input)
  const expected = await page.locator('#intake-description').inputValue()
  let droppedKey = ''
  await page.route('**/api/v1/case-intakes/' + seed.target, async route => {
    const request = route.request()
    if (request.method() === 'PUT' && !droppedKey) {
      droppedKey = request.headers()['idempotency-key']
      expect(droppedKey).toMatch(uuid)
      await route.continue({ headers: { ...request.headers(), 'X-NP-Drop-Response': 'after-commit' } })
    } else await route.continue()
  })
  await page.getByRole('button', { name: '保存草稿', exact: true }).click()
  await expect(page.getByRole('button', { name: '回查原保存结果', exact: true })).toBeVisible()
  await expect(page.getByText('原保存结果待核实；不能换键重发，请回查结果或明确结束原保存。', { exact: true })).toBeVisible()
  expect(droppedKey).toMatch(uuid)
  const key = `np-save-v2:${seed.actor_id}:${seed.target}:${droppedKey}`
  const pending = await page.evaluate(k => localStorage.getItem(k), key)
  expect(pending, 'missing per-request v2 recovery record').not.toBeNull()
  expect(JSON.parse(pending!)).toEqual({ action: 'update', intake_id: seed.target, request_id: droppedKey })
  expect(await page.evaluate(() => Object.keys(localStorage).some(k => k.startsWith('np-save:')))).toBe(false)

  const fresh = await browser.newContext({ storageState: { cookies: [], origins: [{
    origin: 'http://127.0.0.1:3003', localStorage: [{ name: key, value: pending! }],
  }] } })
  const reopened = await fresh.newPage()
  await loginAndOpenDraft(reopened)
  const query = reopened.waitForResponse(r => r.request().method() === 'GET' &&
    new URL(r.url()).pathname === '/api/v1/case-intakes/write-results/' + droppedKey)
  await reopened.getByRole('button', { name: '回查原保存结果', exact: true }).click()
  const response = await query
  expect(response.status()).toBe(200)
  const { data } = await response.json()
  expect(data).toMatchObject({ state: 'committed', request_id: droppedKey, id: seed.target })
  expect(data.revision).toMatch(uuid)
  if (data.action !== undefined) expect(data.action).toBe('update')
  await expect.poll(() => reopened.evaluate(k => localStorage.getItem(k), key)).toBeNull()
  await expect(reopened.locator('#intake-description')).toHaveValue(expected)
  await expect(reopened.getByText('已保存摘要来源', { exact: true })).toBeVisible()
  await expect(reopened.getByText(input, { exact: true }).first()).toBeVisible()
  await fresh.close()
})

test('生成期间保留未发送输入；人工摘要不被候选静默覆盖', async ({ page }) => {
  await loginAndOpenDraft(page)
  let release!: () => void
  let arrived!: () => void
  const blocked = new Promise<void>(resolve => { arrived = resolve })
  const gate = new Promise<void>(resolve => { release = resolve })
  await page.route('**/api/v1/case-intake-assist/contexts/*/turns', async route => {
    if (route.request().method() === 'POST') {
      arrived()
      await gate
    }
    await route.continue()
  })
  const turn = page.waitForResponse(r => r.request().method() === 'POST' &&
    /^\/api\/v1\/case-intake-assist\/contexts\/[^/]+\/turns$/.test(new URL(r.url()).pathname))
  await page.getByRole('textbox', { name: '摘要对话输入' }).fill('虚构第一轮原文')
  await page.getByRole('button', { name: '开始摘要整理', exact: true }).click()
  await blocked
  try {
    await page.getByRole('textbox', { name: '摘要对话输入' }).fill('尚未发送的新文字不得被清空')
  } finally {
    release()
  }
  expect((await turn).status()).toBe(200)
  await expect(page.getByRole('button', { name: '生成摘要候选', exact: true })).toBeVisible()
  await expect(page.getByRole('textbox', { name: '摘要对话输入' })).toHaveValue('尚未发送的新文字不得被清空')
  await page.locator('#intake-description').fill('人工清点后保留的摘要')
  await page.getByRole('checkbox', { name: '已核对来源', exact: true }).check()
  await page.getByRole('button', { name: '应用摘要', exact: true }).click()
  await expect(page.getByRole('alertdialog', { name: '覆盖确认' })).toBeVisible()
  await expect(page.locator('#intake-description')).toHaveValue('人工清点后保留的摘要')
  await page.getByRole('alertdialog', { name: '覆盖确认' }).getByRole('button', { name: /取\s*消/ }).click()
  await expect(page.locator('#intake-description')).toHaveValue('人工清点后保留的摘要')
  await page.getByRole('button', { name: '应用摘要', exact: true }).click()
  await page.getByRole('button', { name: '确认覆盖人工修改并应用', exact: true }).click()
  await expect(page.locator('#intake-description')).toHaveValue(/虚构第一轮原文/)
  await page.locator('#intake-description').fill('人工修改后的摘要须明确解除候选来源')
  await expect(page.getByRole('button', { name: '改为人工摘要保存（不关联来源）', exact: true })).toBeVisible()
  let writes = 0
  const listener = (request: { method(): string; url(): string }) => { if (request.method() === 'PUT' && request.url().includes('/case-intakes/')) writes++ }
  page.on('request', listener)
  await page.getByRole('button', { name: '保存草稿', exact: true }).click()
  await expect(page.getByText(/来源已失效，请重新应用候选/).first()).toBeVisible()
  expect(writes).toBe(0)
  page.off('request', listener)
  await page.getByRole('button', { name: '改为人工摘要保存（不关联来源）', exact: true }).click()
  const manual = await saveResponse(page)
  expect(manual.status()).toBe(200)
  expect(manual.request().postDataJSON().source_selection).toBeUndefined()
  await expect(page.getByText('当前版本尚无已保存来源', { exact: true })).toBeVisible()
  // Even after manual unlink, history still prevents a mixed edit from becoming a full PUT.
  page.on('request', listener)
  await page.locator('#intake-title').fill('虚构的未保存标题')
  await page.getByRole('button', { name: '保存草稿', exact: true }).click()
  await expect(page.getByText(/此草稿已有来源保护，仅支持严格摘要保存/).first()).toBeVisible()
  expect(writes).toBe(0)
  await expect(page.locator('#intake-title')).toHaveValue('虚构的未保存标题')
  page.off('request', listener)

})
