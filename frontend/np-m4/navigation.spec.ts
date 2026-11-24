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
  await expect(page.locator('textarea').filter({ hasText: marker })).toHaveCount(0)
})
