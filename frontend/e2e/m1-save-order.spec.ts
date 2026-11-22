import { expect, test } from '@playwright/test'
import { seedAuthenticatedUser } from './utils/test-helpers'

test('切换草稿后迟到的保存回执不能改变下一次保存的对象或版本', async ({ page }) => {
  await seedAuthenticatedUser(page, 'lawyer')
  const draft = (id: string) => ({
    id, title: 'Synthetic ' + id, client_id: 1, created_by: 1, case_type: 'commercial',
    intake_code: id, metadata: { lawyer_id: 1 }, status: 'draft',
  })
  const ok = (data: unknown) => ({ json: { success: true, data } })
  await page.route('**/api/v1/case-intakes?*', route =>
    route.fulfill(ok({ items: [draft('801'), draft('802')], total: 2 })))
  let release!: () => void
  const paused = new Promise<void>(resolve => { release = resolve })
  const writes: Array<{ id: string; revision: string; title: string }> = []
  await page.route(/\/api\/v1\/case-intakes\/80[12]$/, async route => {
    const id = new URL(route.request().url()).pathname.split('/').pop()!
    if (route.request().method() === 'GET') {
      await route.fulfill(ok({ intake: draft(id), client: { name: 'Synthetic client' }, parties: [], revision: 'r' + id }))
      return
    }
    const body = route.request().postDataJSON()
    writes.push({ id, revision: body.expected_revision, title: body.title })
    if (writes.length === 1) await paused
    await route.fulfill(ok({ id, revision: 'saved' + id, intake_code: id }))
  })

  await page.goto('/case/create')
  await page.getByRole('button', { name: '打开服务端草稿' }).click()
  await page.getByRole('button', { name: 'Synthetic 801' }).click()
  const title = page.getByRole('textbox', { name: '案件名称 *' })
  const save = page.getByRole('button', { name: '保存草稿', exact: true })
  await expect(title).toHaveValue('Synthetic 801')
  await save.click()
  await expect.poll(() => writes.length).toBe(1)
  await page.getByRole('button', { name: 'Synthetic 802' }).click()
  await page.getByRole('button', { name: '切换并恢复所选草稿' }).click()
  await expect(title).toHaveValue('Synthetic 802')
  const receipt = page.waitForResponse(res => res.request().method() === 'PUT')
  release()
  await receipt
  await title.fill('Synthetic B editing')
  await save.click()
  await expect.poll(() => writes[1]).toEqual({ id: '802', revision: 'r802', title: 'Synthetic B editing' })
})
