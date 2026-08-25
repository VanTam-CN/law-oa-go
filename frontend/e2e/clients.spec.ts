/**
 * E2E: 客户管理入口反馈.
 */

import { test, expect } from '@playwright/test'
import { seedAuthenticatedUser, waitForAppShell, waitForPageLoad } from './utils/test-helpers'

test.describe('客户主档案', () => {
  test.beforeEach(async ({ page }) => {
    await seedAuthenticatedUser(page, 'lawyer')
    await page.goto('/client')
    await waitForPageLoad(page)
  })

  test('新增客户入口应该打开新增客户表单', async ({ page }) => {
    await waitForAppShell(page)
    await page.getByRole('button', { name: '新增客户' }).click()

    await expect(page.getByRole('dialog', { name: '新增客户' })).toBeVisible()
    await expect(page.getByText('客户类型 *')).toBeVisible()
    await expect(page.getByPlaceholder('请输入客户名称')).toBeVisible()
    await expect(page.getByText('统一社会信用代码 *')).toBeVisible()
    await expect(page.getByPlaceholder('按营业执照填写')).toBeVisible()
  })

  test('企业客户身份应按统一社会信用代码提交', async ({ page }) => {
    await waitForAppShell(page)
    await page.getByRole('button', { name: '新增客户' }).click()
    const dialog = page.getByRole('dialog', { name: '新增客户' })
    await dialog.getByPlaceholder('请输入客户名称').fill('虚构验收企业客户')
    await dialog.getByPlaceholder('按营业执照填写').fill('91310000TEST000018')

    const requestPromise = page.waitForRequest((request) =>
      request.url().includes('/api/v1/clients') && request.method() === 'POST',
    )
    await dialog.getByRole('button', { name: '保存客户' }).click()
    const request = await requestPromise
    const payload = request.postDataJSON()
    expect(payload.identity_type).toBe('SOCIAL_CREDIT_CODE')
    expect(payload.identity_number).toBe('91310000TEST000018')
    expect(payload.id_card).toBeUndefined()
  })

  test('编辑主联系人应写入独立联系人接口且不污染客户主档案', async ({ page }) => {
    await waitForAppShell(page)
    await page.getByRole('button', { name: '编辑主联系人' }).click()
    const dialog = page.getByRole('dialog', { name: '编辑主联系人' })
    await dialog.getByPlaceholder('联系人姓名').fill('虚构联系人甲')

    const requestPromise = page.waitForRequest((request) =>
      /\/api\/v1\/clients\/\d+\/primary-contact$/.test(new URL(request.url()).pathname) &&
      request.method() === 'PUT',
    )
    await dialog.getByRole('button', { name: /保\s*存/ }).click()
    const payload = (await requestPromise).postDataJSON()
    expect(payload.version).toBeGreaterThanOrEqual(0)
    expect(payload.name).toBe('虚构联系人甲')
    expect(payload.contact_person).toBeUndefined()
    expect(payload.notes).toBeUndefined()
    expect(payload.email).toBeDefined()
  })
})
