import { expect, test } from '@playwright/test'
import { loginAndOpenDraft, seed } from '../e2e/utils/np-live-draft'

test('C03 未发送的对话留在页面，浏览器后退取消后仍可继续编辑', async ({ page }) => {
  await loginAndOpenDraft(page)
  const chat = page.getByRole('textbox', { name: '摘要对话输入' })
  const marker = '待核对的虚构口述'
  await chat.fill(marker)
  const start = page.url()
  let prompts = 0
  page.on('dialog', async dialog => {
    expect(['confirm', 'beforeunload']).toContain(dialog.type())
    prompts++
    await dialog.dismiss()
  })
  await page.evaluate(() => history.back())
  await expect.poll(() => prompts).toBe(1)
  await expect(page).toHaveURL(start)
  await expect(chat).toHaveValue(marker)
})

test('C03 当前标签的已签发令牌到期后不再展示草稿和仅内存输入', async ({ page }) => {
  await page.clock.install()
  await loginAndOpenDraft(page)
  const marker = '合成过期标签输入'
  await page.locator('#intake-description').fill(marker)
  const ms = await page.evaluate(() => {
    const payload = sessionStorage.getItem('auth_token')!.split('.')[1]
    const { exp } = JSON.parse(atob(payload.replace(/-/g, '+').replace(/_/g, '/')))
    return exp * 1000 - Date.now() + 2000
  })
  expect(ms).toBeGreaterThan(0)
  await page.clock.fastForward(ms)
  await expect(page.locator('#intake-title')).toHaveCount(0)
  await expect(page.getByText(seed.target_title, { exact: true })).toHaveCount(0)
  await expect(page.locator('#intake-description')).toHaveCount(0)
  await expect(page.getByText(/登录已到期或账号已切换/)).toBeVisible()
})

test('审核回归：同一草稿切换步骤后保留未发送口述', async ({ page }) => {
  await loginAndOpenDraft(page)
  const chat = page.getByRole('textbox', { name: '摘要对话输入' })
  await chat.fill('合成未发送口述，切步骤后仍须保留')
  const steps = page.locator('.batch-stepper button')
  await steps.nth(1).click()
  await expect(chat).not.toBeVisible()
  await steps.first().click()
  await expect(chat).toHaveValue('合成未发送口述，切步骤后仍须保留')
})

test('审核回归：取消换草稿保留待添加主体，确认切换清除上一案的临时输入', async ({ page }) => {
  await loginAndOpenDraft(page)
  const name = page.locator('#intake-related-party')
  const identity = page.getByLabel('相关方身份标识', { exact: true })
  const tag = page.locator('#intake-tag')
  await name.fill('合成未添加主体')
  await identity.fill('SYNTHETIC0000')
  await tag.fill('甲案未添加标签')
  const target = page.getByRole('button', { name: '虚构对照草稿', exact: true })
  await target.click()
  await page.getByRole('button', { name: '继续编辑当前内容' }).click()
  await expect(name).toHaveValue('合成未添加主体')
  await expect(identity).toHaveValue('SYNTHETIC0000')
  let release!: () => void
  const gate = new Promise<void>(resolve => { release = resolve })
  let gotDetail = false
  await page.route('**/api/v1/case-intakes/' + seed.control, async route => {
    const response = await route.fetch()
    gotDetail = true
    await gate
    await route.fulfill({ response })
  }, { times: 1 })
  await target.click()
  await page.getByRole('button', { name: '切换并恢复所选草稿' }).click()
  await expect.poll(() => gotDetail).toBe(true)
  await tag.fill('详情返回前新增的合成输入')
  release()
  await expect(page.getByText(/加载期间有新编辑/).first()).toBeVisible()
  await expect(tag).toHaveValue('详情返回前新增的合成输入')
  await expect(page.locator('#intake-title')).toHaveValue(seed.target_title)
  await target.click()
  await page.getByRole('button', { name: '切换并恢复所选草稿' }).click()
  await expect(page.locator('#intake-title')).toHaveValue('虚构对照草稿')
  await expect(name).toHaveValue('')
  await expect(identity).toHaveValue('')
  await expect(tag).toHaveValue('')
})

test('审核回归：空白接案只添加标签，取消菜单离开后标签仍在', async ({ page }) => {
  await loginAndOpenDraft(page)
  await page.goto('/dashboard')
  await page.getByRole('button', { name: /新建立案$/ }).click()
  await page.getByRole('button', { name: /添加标签$/ }).click()
  await expect(page.getByText('新标签1', { exact: true })).toBeVisible()
  let prompts = 0
  page.on('dialog', async dialog => { prompts++; await dialog.dismiss() })
  await page.evaluate(() => history.back())
  await expect.poll(() => prompts).toBe(1)
  await expect(page.getByText('新标签1', { exact: true })).toBeVisible()
})
