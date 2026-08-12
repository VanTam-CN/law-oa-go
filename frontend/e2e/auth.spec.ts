/**
 * E2E: 当前登录流程与基础权限.
 */

import { test, expect } from '@playwright/test'
import {
  installApiMocks,
  isLoggedIn,
  login,
  logout,
  seedAuthenticatedUser,
  waitForAppShell,
  waitForPageLoad,
} from './utils/test-helpers'

test.describe('登录流程', () => {
  test.beforeEach(async ({ page }) => {
    await installApiMocks(page)
    await page.goto('/')
    await waitForPageLoad(page)
  })

  test('应该显示当前登录页面', async ({ page }) => {
    await expect(page).toHaveURL(/\/login$/)
    await expect(page.getByText('示例律师事务所OA登录')).toBeVisible()
    await expect(page.getByPlaceholder('账号或邮箱，如 admin / demo.admin')).toBeVisible()
    await expect(page.getByPlaceholder('密码')).toBeVisible()
    await expect(page.locator('button[type="submit"]')).toBeVisible()
    await expect(page.getByText('首次使用或无法登录？请联系律所系统管理员。')).toBeVisible()
  })

  test('登录失败应该显示错误信息', async ({ page }) => {
    await page.getByPlaceholder('账号或邮箱，如 admin / demo.admin').fill('wronguser')
    await page.getByPlaceholder('密码').fill('wrongpass')
    await page.locator('button[type="submit"]').click()

    await expect(page.locator('.ant-message')).toContainText('账号或密码错误')
    await expect(page.locator('.ant-message-notice')).toHaveCount(1)
  })

  test('服务不可用时只显示一次准确提示', async ({ page }) => {
    await page.route('**/api/v1/auth/login', async (route) => {
      await route.fulfill({ status: 500, contentType: 'text/plain', body: '' })
    })

    await page.getByPlaceholder('账号或邮箱，如 admin / demo.admin').fill('lawyer')
    await page.getByPlaceholder('密码').fill('Demo@2026')
    await page.locator('button[type="submit"]').click()

    await expect(page.locator('.ant-message')).toContainText(
      '系统服务暂不可用，请稍后重试或联系系统管理员',
    )
    await expect(page.locator('.ant-message-notice')).toHaveCount(1)
    await expect(page.locator('.ant-message')).not.toContainText('账号或密码错误')
  })

  test('密码显隐按钮支持键盘操作', async ({ page }) => {
    const password = page.locator('input[aria-label="密码"]')
    await password.fill('Demo@2026')

    const showPassword = page.getByRole('button', { name: '显示密码' })
    await showPassword.focus()
    await page.keyboard.press('Enter')
    await expect(password).toHaveAttribute('type', 'text')

    const hidePassword = page.getByRole('button', { name: '隐藏密码' })
    await hidePassword.focus()
    await page.keyboard.press('Space')
    await expect(password).toHaveAttribute('type', 'password')
  })

  test('律师账号登录后应该进入工作台', async ({ page }) => {
    await login(page, 'lawyer')
    await waitForAppShell(page)

    expect(await isLoggedIn(page)).toBe(true)
    await expect(page).toHaveURL(/\/dashboard$/)
    await expect(page.getByRole('heading', { name: '上午好，张律师' })).toBeVisible()
  })

  test('登出后应该返回登录页', async ({ page }) => {
    await login(page, 'lawyer')
    await logout(page)

    await expect(page).toHaveURL(/\/login$/)
  })

  test('未登录访问受保护页面应该重定向到登录页', async ({ page }) => {
    await page.goto('/case')
    await expect(page).toHaveURL(/\/login$/)
  })

  test('刷新后应该保持登录状态', async ({ page }) => {
    await login(page, 'lawyer')
    await page.reload()

    await waitForAppShell(page)
    await expect(page).toHaveURL(/\/dashboard$/)
  })
})

test.describe('登录表单验证', () => {
  test.beforeEach(async ({ page }) => {
    await installApiMocks(page)
    await page.goto('/login')
    await waitForPageLoad(page)
  })

  test('空账号应该显示验证错误', async ({ page }) => {
    await page.getByPlaceholder('密码').fill('Demo@2026')
    await page.locator('button[type="submit"]').click()

    await expect(page.locator('.ant-form-item-explain-error')).toContainText('请输入账号或邮箱')
  })

  test('空密码应该显示验证错误', async ({ page }) => {
    await page.getByPlaceholder('账号或邮箱，如 admin / demo.admin').fill('lawyer')
    await page.locator('button[type="submit"]').click()

    await expect(page.locator('.ant-form-item-explain-error')).toContainText('请输入密码')
  })
})

test.describe('角色权限', () => {
  test('管理员应该能访问用户管理页', async ({ page }) => {
    await seedAuthenticatedUser(page, 'admin')
    await page.goto('/user')

    await waitForAppShell(page)
    await expect(page.getByRole('heading', { name: '用户管理' })).toBeVisible()
  })

  test('普通律师不应该能访问用户管理页', async ({ page }) => {
    await seedAuthenticatedUser(page, 'lawyer')
    await page.goto('/user')

    await expect(page.getByText('无权访问')).toBeVisible()
  })
})

test.describe('顶部入口反馈', () => {
  test.beforeEach(async ({ page }) => {
    await seedAuthenticatedUser(page, 'lawyer')
    await page.goto('/dashboard')
    await waitForPageLoad(page)
  })

  test('帮助中心入口应该显示帮助反馈', async ({ page }) => {
    await waitForAppShell(page)
    await page.locator('.user-menu').click()
    await page.getByText('帮助中心').click()

    await expect(page.getByRole('dialog').getByText('帮助中心', { exact: true })).toBeVisible()
    await expect(page.getByText('完整帮助中心建设中')).toBeVisible()
  })

  test('通知为空时不应该展示全部已读动作', async ({ page }) => {
    await waitForAppShell(page)
    await page.locator('.notification-btn').click()

    await expect(page.getByText('暂无通知')).toBeVisible()
    await expect(page.getByText('全部已读')).toHaveCount(0)
  })
})
