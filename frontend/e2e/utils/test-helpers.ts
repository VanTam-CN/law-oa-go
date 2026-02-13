/**
 * E2E测试工具函数
 * Story 6.1: Playwright配置
 */

import { Page, expect } from '@playwright/test'

// 测试用户凭证
export const TEST_USERS = {
  admin: {
    username: 'admin',
    password: 'admin123',
  },
  lawyer: {
    username: 'lawyer',
    password: 'lawyer123',
  },
  assistant: {
    username: 'assistant',
    password: 'assistant123',
  },
}

// 等待页面加载完成
export async function waitForPageLoad(page: Page) {
  await page.waitForLoadState('networkidle')
  await page.waitForLoadState('domcontentloaded')
}

// 登录函数
export async function login(page: Page, username: string, password: string) {
  await page.goto('/login')
  await waitForPageLoad(page)

  // 填写登录表单
  await page.fill('input[placeholder*="用户名"], input[name="username"]', username)
  await page.fill('input[placeholder*="密码"], input[name="password"]', password)

  // 点击登录按钮
  await page.click('button:has-text("登录"), button[type="submit"]')

  // 等待跳转
  await page.waitForURL(/\/(dashboard|cases|home)/, { timeout: 10000 })
}

// 登出函数
export async function logout(page: Page) {
  // 点击用户头像或菜单
  await page.click('[data-testid="user-menu"], .ant-dropdown-trigger:has(.anticon-user)')

  // 点击登出
  await page.click('text=退出登录, text=登出, text=Logout')

  // 等待跳转到登录页
  await page.waitForURL('/login', { timeout: 5000 })
}

// 检查是否已登录
export async function isLoggedIn(page: Page): Promise<boolean> {
  try {
    // 检查是否存在登出按钮或用户菜单
    const userMenu = await page.$('[data-testid="user-menu"], .ant-dropdown-trigger')
    return userMenu !== null
  } catch {
    return false
  }
}

// 导航到指定页面
export async function navigateTo(page: Page, path: string) {
  await page.goto(path)
  await waitForPageLoad(page)
}

// 检查Toast消息
export async function checkToastMessage(page: Page, message: string) {
  const toast = page.locator('.ant-message, .ant-notification')
  await expect(toast).toContainText(message)
}

// 等待表格加载
export async function waitForTableLoad(page: Page) {
  await page.waitForSelector('.ant-table-tbody tr', { timeout: 10000 })
  await page.waitForLoadState('networkidle')
}

// 获取表格行数
export async function getTableRowCount(page: Page): Promise<number> {
  const rows = await page.$$('.ant-table-tbody tr')
  return rows.length
}

// 点击表格行操作按钮
export async function clickTableRowAction(
  page: Page,
  rowIndex: number,
  actionText: string,
) {
  const rows = await page.$$('.ant-table-tbody tr')
  if (rows[rowIndex]) {
    await rows[rowIndex].click(`button:has-text("${actionText}")`)
  }
}

// 填写表单字段
export async function fillForm(
  page: Page,
  fields: Record<string, string>,
) {
  for (const [label, value] of Object.entries(fields)) {
    const input = page.locator(
      `.ant-form-item:has-text("${label}") input, .ant-form-item:has-text("${label}") textarea`,
    )
    await input.fill(value)
  }
}

// 选择下拉选项
export async function selectOption(
  page: Page,
  label: string,
  optionText: string,
) {
  await page.click(`.ant-form-item:has-text("${label}") .ant-select-selector`)
  await page.click(`.ant-select-dropdown:visible .ant-select-item:has-text("${optionText}")`)
}

// 检查元素是否可见
export async function isElementVisible(page: Page, selector: string): Promise<boolean> {
  try {
    const element = await page.$(selector)
    if (!element) return false
    return await element.isVisible()
  } catch {
    return false
  }
}

// 等待模态框出现
export async function waitForModal(page: Page) {
  await page.waitForSelector('.ant-modal-content', { timeout: 5000 })
}

// 关闭模态框
export async function closeModal(page: Page) {
  await page.click('.ant-modal-close, button:has-text("取消")')
  await page.waitForSelector('.ant-modal-content', { state: 'hidden' })
}

// 点击确认按钮
export async function clickConfirm(page: Page, buttonText = '确定') {
  await page.click(`.ant-modal-content button:has-text("${buttonText}")`)
}

// 截图并添加到报告
export async function takeScreenshot(page: Page, name: string) {
  await page.screenshot({
    path: `playwright-report/screenshots/${name}.png`,
    fullPage: true,
  })
}

// Mock API响应
export async function mockApiResponse(
  page: Page,
  path: string,
  response: any,
  status = 200,
) {
  await page.route(`**/api/**${path}`, (route) => {
    route.fulfill({
      status,
      contentType: 'application/json',
      body: JSON.stringify(response),
    })
  })
}
