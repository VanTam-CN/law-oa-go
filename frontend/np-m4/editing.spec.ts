import { expect, test, type Route } from '@playwright/test'
import { applyCandidate, loginAndOpenDraft, seed, uuid } from '../e2e/utils/np-live-draft'

const intakePath = '/api/v1/case-intakes/' + seed.target

test('C13：保存并退出期间编辑B保留，晚到回执不导航；重开证实服务器A、页面B', async ({ page, browser }) => {
  await loginAndOpenDraft(page)
  const submittedA = 'C13退出已提交A ' + Date.now()
  const laterB = 'C13回执后未保存B ' + Date.now()
  let release!: () => void
  const gate = new Promise<void>(resolve => { release = resolve })
  let heldResponse!: Awaited<ReturnType<Route['fetch']>>
  let heldRoute: Route | null = null
  await page.route('**' + intakePath, async (route: Route) => {
    const request = route.request()
    if (request.method() !== 'PUT') return route.continue()
    heldResponse = await route.fetch()
    heldRoute = route
    await gate
    await route.fulfill({ response: heldResponse })
    return
  })
  await page.locator('#intake-description').fill(submittedA)
  const exiting = page.getByRole('button', { name: '保存并退出', exact: true }).click()
  await expect.poll(() => Boolean(heldRoute)).toBe(true)
  await expect(page).not.toHaveURL(/\/case$/)
  await expect(page.locator('#intake-description')).toHaveValue(submittedA)
  await page.locator('#intake-description').fill(laterB)
  const receipt = page.waitForResponse(request => request.request().method() === 'PUT' &&
    new URL(request.url()).pathname === intakePath)
  release()
  expect((await receipt).status()).toBe(200)
  await exiting
  await expect(page).not.toHaveURL(/\/case$/)
  await expect(page.locator('#intake-description')).toHaveValue(laterB)
  await expect(page.getByText(/已保存.*未保存|已确认.*未保存/).first()).toBeVisible()

  const fresh = await browser.newContext()
  const reopened = await fresh.newPage()
  await loginAndOpenDraft(reopened)
  await expect(reopened.locator('#intake-description')).toHaveValue(submittedA)
  await expect(reopened.getByText('当前版本尚无已保存来源', { exact: true })).toBeVisible()
  await fresh.close()
  await expect(page.locator('#intake-description')).toHaveValue(laterB)
})

test('C13：断线保存回查期间输入B不丢；放行回执后重开证实仅A且带来源', async ({ page, browser }) => {
  await loginAndOpenDraft(page)
  const inputA = 'C13断线已提交来源A ' + Date.now()
  await applyCandidate(page, inputA)
  let droppedKey = ''
  await page.route('**' + intakePath, async route => {
    const request = route.request()
    if (request.method() === 'PUT' && !droppedKey) {
      droppedKey = request.headers()['idempotency-key']
      await route.continue({ headers: { ...request.headers(), 'X-NP-Drop-Response': 'after-commit' } })
      return
    }
    await route.continue()
  })
  await page.getByRole('button', { name: '保存草稿', exact: true }).click()
  await expect(page.getByText(/原保存结果待核实/).first()).toBeVisible()
  expect(droppedKey).toMatch(uuid)

  const inputB = 'C13回查期间人工输入B ' + Date.now()
  let releaseQuery!: () => void
  const queryGate = new Promise<void>(resolve => { releaseQuery = resolve })
  let queryArrived = false
  await page.route('**/api/v1/case-intakes/write-results/' + droppedKey, async route => {
    queryArrived = true
    await queryGate
    return route.continue()
  })
  const query = page.waitForResponse(request => request.request().method() === 'GET' &&
    new URL(request.url()).pathname === '/api/v1/case-intakes/write-results/' + droppedKey)
  await page.getByRole('button', { name: '回查原保存结果', exact: true }).click()
  await expect.poll(() => queryArrived).toBe(true)
  await page.locator('#intake-description').fill(inputB)
  await expect(page.locator('#intake-description')).toHaveValue(inputB)
  releaseQuery()
  expect((await query).status()).toBe(200)
  await expect(page.getByText(/已查到原保存结果.*未保存/).first()).toBeVisible()
  await expect(page.locator('#intake-description')).toHaveValue(inputB)
  const recoveryKey = `np-save-v2:${seed.actor_id}:${seed.target}:${droppedKey}`
  await expect.poll(() => page.evaluate(key => localStorage.getItem(key), recoveryKey)).toBeNull()

  const fresh = await browser.newContext()
  const reopened = await fresh.newPage()
  await loginAndOpenDraft(reopened)
  await expect(reopened.locator('#intake-description')).toHaveValue(new RegExp(inputA))
  const serverText = await reopened.locator('#intake-description').inputValue()
  expect(serverText).not.toContain(inputB)
  await expect(reopened.getByText('已保存摘要来源', { exact: true })).toBeVisible()
  await expect(reopened.getByText(inputA, { exact: true }).first()).toBeVisible()
  await fresh.close()
})

test('C14：重载取消不发起取数；确认后晚到响应不覆盖C也不提示恢复成功', async ({ page }) => {
  await loginAndOpenDraft(page)
  await applyCandidate(page, 'C14应用候选A，用于激活助手与来源提示 ' + Date.now())
  let heldResponse!: Awaited<ReturnType<Route['fetch']>>
  let heldRoute: Route | null = null
  let detailRequests = 0
  await page.route('**' + intakePath, async route => {
    const request = route.request()
    if (request.method() !== 'GET') return route.continue()
    detailRequests += 1
    heldResponse = await route.fetch()
    heldRoute = route
    return
  })
  const inputB = 'C14重载前人工输入B ' + Date.now()
  await page.locator('#intake-description').fill(inputB)
  await page.getByRole('button', { name: '重新加载服务端草稿', exact: true }).click()
  await expect(page.getByText('当前页面有未保存内容，切换后会用所选服务端草稿替换表单。确定切换吗？')).toBeVisible()
  await page.getByRole('dialog').getByRole('button', { name: '继续编辑当前内容' }).click()
  await expect.poll(() => detailRequests).toBe(0)
  await expect(page.locator('#intake-description')).toHaveValue(inputB)
  await expect(page.locator('#intake-title')).not.toHaveValue('')

  await page.getByRole('button', { name: '重新加载服务端草稿', exact: true }).click()
  await page.getByRole('dialog').getByRole('button', { name: '切换并恢复所选草稿' }).click()
  const inputC = inputB + '；确认后新输入C'
  await expect.poll(() => Boolean(heldRoute)).toBe(true)
  expect(detailRequests).toBe(1)
  await page.locator('#intake-description').fill(inputC)
  await heldRoute!.fulfill({ response: heldResponse })
  await expect(page.getByText(/加载期间有新编辑/).first()).toBeVisible()
  await expect(page.locator('#intake-description')).toHaveValue(inputC)
  await expect(page.locator('.ant-message')).not.toContainText('已恢复所选服务端草稿')
})
