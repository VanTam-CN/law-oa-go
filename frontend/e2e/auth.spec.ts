/**
 * E2E测试: 登录流程
 * Story 6.2: 关键流程E2E
 */

import { test, expect } from '@playwright/test'
import { login, logout, isLoggedIn, waitForPageLoad } from './utils/test-helpers'

test.describe('登录流程', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
    await waitForPageLoad(page)
  })

  test('应该显示登录页面', async ({ page }) => {
    // 验证登录页面元素存在
    await expect(page).toHaveURL(/login/)

    // 检查登录表单元素
    await expect(page.locator('input[placeholder*="用户名"], input[name="username"]')).toBeVisible()
    await expect(page.locator('input[placeholder*="密码"], input[name="password"]')).toBeVisible()
    await expect(page.locator('button:has-text("登录")')).toBeVisible()
  })

  test('登录失败应该显示错误信息', async ({ page }) => {
    // 使用错误的凭证
    await page.fill('input[placeholder*="用户名"], input[name="username"]', 'wronguser')
    await page.fill('input[placeholder*="密码"], input[name="password"]', 'wrongpass')
    await page.click('button:has-text("登录")')

    // 应该显示错误信息
    await expect(page.locator('.ant-message, .ant-alert')).toBeVisible({ timeout: 5000 })
  })

  test('成功登录后应该跳转到主页', async ({ page }) => {
    // 使用测试账号登录
    await login(page, 'admin', 'admin123')

    // 验证已登录
    const loggedIn = await isLoggedIn(page)
    expect(loggedIn).toBe(true)

    // 验证URL不是登录页
    await expect(page).not.toHaveURL(/login/)
  })

  test('登出后应该返回登录页', async ({ page }) => {
    // 先登录
    await login(page, 'admin', 'admin123')

    // 然后登出
    await logout(page)

    // 验证返回登录页
    await expect(page).toHaveURL(/login/)
  })

  test('未登录访问受保护页面应该重定向到登录页', async ({ page }) => {
    // 直接访问受保护页面
    await page.goto('/cases')

    // 应该被重定向到登录页
    await expect(page).toHaveURL(/login/, { timeout: 5000 })
  })
})

test.describe('登录表单验证', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
    await waitForPageLoad(page)
  })

  test('空用户名应该显示验证错误', async ({ page }) => {
    // 不填用户名，直接点击登录
    await page.fill('input[placeholder*="密码"], input[name="password"]', 'password123')
    await page.click('button:has-text("登录")')

    // 应该显示验证错误
    await expect(page.locator('.ant-form-item-explain-error')).toBeVisible()
  })

  test('空密码应该显示验证错误', async ({ page }) => {
    await page.fill('input[placeholder*="用户名"], input[name="username"]', 'admin')
    await page.click('button:has-text("登录")')

    await expect(page.locator('.ant-form-item-explain-error')).toBeVisible()
  })

  test('记住登录状态', async ({ page, context }) => {
    // 登录
    await login(page, 'admin', 'admin123')

    // 获取cookies
    const cookies = await context.cookies()
    expect(cookies.length).toBeGreaterThan(0)

    // 刷新页面
    await page.reload()

    // 应该仍然保持登录状态
    const loggedIn = await isLoggedIn(page)
    expect(loggedIn).toBe(true)
  })
})

test.describe('角色权限', () => {
  test('管理员应该能访问管理页面', async ({ page }) => {
    await login(page, 'admin', 'admin123')

    // 访问管理页面
    await page.goto('/admin')

    // 应该能访问
    await expect(page).not.toHaveURL(/login/)
  })

  test('普通律师不应该能访问管理页面', async ({ page }) => {
    await login(page, 'lawyer', 'lawyer123')

    // 尝试访问管理页面
    await page.goto('/admin')

    // 应该显示403或重定向
    const is403 = await page.locator('text=403, text=无权限').isVisible()
    const redirected = page.url().includes('/login') || page.url().includes('/dashboard')

    expect(is403 || redirected).toBe(true)
  })
})
