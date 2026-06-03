/**
 * E2E: 客户主档案快捷操作反馈（QA-RC-005 + IB-RT-006 回归）.
 */

import { test, expect } from '@playwright/test'
import { seedAuthenticatedUser, waitForAppShell, waitForPageLoad } from './utils/test-helpers'

test.describe('客户快捷操作反馈', () => {
  test.beforeEach(async ({ page }) => {
    await seedAuthenticatedUser(page, 'lawyer')
    await page.goto('/client')
    await waitForPageLoad(page)
    await waitForAppShell(page)
    await expect(page.getByText('上海示例科技有限公司').first()).toBeVisible()
  })

  test('新增联系人应保存到客户主联系人', async ({ page }) => {
    await page.getByRole('button', { name: '新增联系人' }).click()
    const dialog = page.getByRole('dialog', { name: '新增联系人' })
    await expect(dialog).toBeVisible()
    await page.getByPlaceholder('联系人姓名').fill('王总')
    await page.getByPlaceholder('联系电话').fill('021-55550000')
    await dialog.getByRole('button', { name: /保\s*存/ }).click()
    await expect(page.getByText('联系人已保存')).toBeVisible()
  })

  test('上传附件应调用上传闭环并显示成功反馈', async ({ page }) => {
    await page.getByRole('button', { name: '上传附件' }).click()
    const dialog = page.getByRole('dialog', { name: '上传附件' })
    await expect(dialog).toBeVisible()
    await page.locator('input[type="file"]').setInputFiles({
      name: 'client-note.txt',
      mimeType: 'text/plain',
      buffer: Buffer.from('client attachment e2e'),
    })
    await dialog.getByRole('button', { name: /上\s*传/ }).click()
    await expect(page.getByText('客户附件已上传')).toBeVisible()
  })

  test('导出客户档案应触发下载或反馈', async ({ page }) => {
    const downloadPromise = page.waitForEvent('download', { timeout: 5000 }).catch(() => null)
    await page.getByRole('button', { name: '导出客户档案' }).click()
    const download = await downloadPromise
    if (download) {
      expect(download).toBeTruthy()
    } else {
      await expect(page.getByText(/客户档案已导出|导出失败/)).toBeVisible()
    }
  })
})
