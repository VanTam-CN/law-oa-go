import { expect, test } from '@playwright/test'
import { seedAuthenticatedUser } from './utils/test-helpers'

for (const stage of ['save', 'facts-confirmation', 'conflict-check'] as const) {
  test(`切换草稿后忽略旧对象迟到的 ${stage} 回执`, async ({ page }) => {
    await seedAuthenticatedUser(page, 'lawyer')
    const draft = (id: string) => ({
      id, title: 'Synthetic ' + id, client_id: 1, created_by: 1, case_type: 'commercial',
      intake_code: id, metadata: { lawyer_id: 1, business_area: '公司与并购', sub_area: '投资与融资' }, status: 'draft',
    })
    const ok = (data: unknown) => ({ json: { success: true, data } })
    await page.route('**/api/v1/case-intakes?*', route =>
      route.fulfill(ok({ items: [draft('801'), draft('802')], total: 2 })))
    let release!: () => void
    const paused = new Promise<void>(resolve => { release = resolve })
    const writes: Array<{ id: string; revision: string; title: string }> = []
    const calls: string[] = []
    await page.route(url => /^\/api\/v1\/case-intakes\/80[12](\/|$)/.test(url.pathname), async route => {
      const [, id, action = 'save'] = new URL(route.request().url()).pathname.match(/case-intakes\/(80[12])(?:\/(.+))?$/)!
      if (route.request().method() === 'GET') {
        if (action.startsWith('source-snapshots/')) {
          const revision = action.slice('source-snapshots/'.length)
          return route.fulfill(ok({ revision, source_revision: '', source: null, current: true }))
        }
        expect(action).toBe('save')
        const parties = [{ party_role: 'opposing_party', entity_name: 'Synthetic opponent', entity_type: 'LEGAL_PERSON' }]
        await route.fulfill(ok({ intake: draft(id), client: { name: 'Synthetic client' }, parties, has_source_history: false, revision: 'r' + id }))
        return
      }
      calls.push(id + '/' + action)
      if (action === 'save') {
        const body = route.request().postDataJSON()
        writes.push({ id, revision: body.expected_revision, title: body.title })
      }
      if (id === '801' && action === stage) await paused
      const data = action === 'conflict-check'
        ? { revision: 'checked801', taskId: 'A-only', status: 'COMPLETED', result: { checkId: 'A-only', hasConflict: false } }
        : { id, revision: action + id, intake_code: id }
      await route.fulfill(ok(data))
    })

    await page.goto('/case/create')
    await page.getByRole('button', { name: '打开服务端草稿' }).click()
    await page.getByRole('button', { name: 'Synthetic 801' }).click()
    const title = page.getByRole('textbox', { name: '案件名称 *' })
    const save = page.getByRole('button', { name: '保存草稿', exact: true })
    await expect(title).toHaveValue('Synthetic 801')
    if (stage === 'save') {
      await save.click()
    } else {
      await page.getByPlaceholder('输入证件号或统一社会信用代码').fill('TEST801')
      await page.getByRole('button', { name: '保存并进行利益冲突检查' }).click()
    }
    await expect.poll(() => calls).toContain('801/' + stage)
    await page.getByRole('button', { name: 'Synthetic 802' }).click()
    await page.getByRole('button', { name: '切换并恢复所选草稿' }).click()
    await expect(title).toHaveValue('Synthetic 802')
    const receipt = page.waitForResponse(res => stage === 'save' ? res.request().method() === 'PUT' : res.url().endsWith('/' + stage))
    release()
    await receipt
    await expect(title).toHaveValue('Synthetic 802')
    await title.fill('Synthetic B editing')
    await save.click()
    await expect.poll(() => writes[1]).toEqual({ id: '802', revision: 'r802', title: 'Synthetic B editing' })
    if (stage === 'facts-confirmation') expect(calls).not.toContain('801/conflict-check')
    await expect(page.locator('.batch-bottom-bar')).toContainText('未检测')
  })
}
