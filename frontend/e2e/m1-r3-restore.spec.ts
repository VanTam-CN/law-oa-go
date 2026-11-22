import { expect, test, type Page } from '@playwright/test'
import { seedAuthenticatedUser } from './utils/test-helpers'

async function openSavedDraft(page: Page, metadata: Record<string, unknown> = { aliases: ['Historic Alias'] }, aliases: string | null = null) {
  await seedAuthenticatedUser(page, 'lawyer')
  const list = '/api/v1/case-intakes'
  await page.route(url => url.pathname === list, route => route.fulfill({
    json: { success: true, data: { items: [{ id: 801, title: '打开合成草稿' }], total: 1 } },
  }))
  await page.route(url => url.pathname === list + '/801', route => route.fulfill({
    json: {
      success: true,
      data: {
        intake: { id: 801, title: 'Synthetic draft', status: 'draft', client_id: 1 },
        revision: 'fixture-rev',
        parties: [
          { party_role: 'opposing_party', entity_type: 'LEGAL_PERSON', entity_name: 'Sample opponent',
            aliases, metadata },
          { party_role: 'related_party', entity_type: 'LEGAL_PERSON', entity_name: 'Synthetic related party' },
        ],
        materials: [],
      },
    },
  }))
  await page.goto('/case/create')
  await page.getByRole('button', { name: '打开服务端草稿' }).click()
  await page.getByRole('button', { name: '打开合成草稿' }).click()
  await expect(page.getByRole('textbox', { name: '法定名称或证件姓名' })).toHaveValue('Sample opponent')
}

test('M1 R3 historical alias remains visible for lawyer review', async ({ page }) => {
  await openSavedDraft(page)
  await expect(page.locator('#intake-opponent-aliases')).toHaveValue('Historic Alias')
})

test('M1 R3 same-revision local recovery retains the related party', async ({ page }) => {
  await openSavedDraft(page)
  await expect(page.locator('.batch-related-party')).toContainText('Synthetic related party')
  await page.reload()
  await page.getByRole('button', { name: '继续未完成草稿' }).click()
  await expect(page.locator('.batch-related-party')).toContainText('Synthetic related party')
})

test('M1 R4 empty historical alias key falls through to the next legacy key', async ({ page }) => {
  const metadata = { aliases: [], client_aliases: ['Historic Alias'] }
  await openSavedDraft(page, metadata)
  await expect(page.locator('#intake-opponent-aliases')).toHaveValue('Historic Alias')
})

test('M1 R4 explicit empty column wins over historical aliases', async ({ page }) => {
  await openSavedDraft(page, { aliases: ['Historic Alias'] }, '')
  await expect(page.locator('#intake-opponent-aliases')).toHaveValue('')
})
