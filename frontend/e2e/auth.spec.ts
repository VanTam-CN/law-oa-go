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
    await expect(page.getByPlaceholder('账号或邮箱')).toBeVisible()
    await expect(page.getByPlaceholder('密码')).toBeVisible()
    await expect(page.getByRole('checkbox', { name: '记住我' })).not.toBeChecked()
    await expect(page.locator('button[type="submit"]')).toBeVisible()
  })

  test('登录失败应该显示错误信息', async ({ page }) => {
    await page.getByPlaceholder('账号或邮箱').fill('wronguser')
    await page.getByPlaceholder('密码').fill('wrongpass')
    await page.locator('button[type="submit"]').click()

    await expect(page.locator('.ant-message')).toContainText('账号或密码错误')
  })

  test('密码显隐控件应支持键盘操作并同步状态', async ({ page }) => {
    const passwordInput = page.getByRole('textbox', { name: '密码' })
    const toggle = page.getByRole('button', { name: '显示密码' })

    await expect(toggle).toBeVisible()
    await expect(toggle).toHaveAttribute('aria-pressed', 'false')
    await passwordInput.fill('Demo@2026')
    await toggle.focus()
    await page.keyboard.press('Space')

    await expect(passwordInput).toHaveAttribute('type', 'text')
    await expect(page.getByRole('button', { name: '隐藏密码' })).toHaveAttribute(
      'aria-pressed',
      'true',
    )

    await page.getByRole('button', { name: '隐藏密码' }).click()
    await expect(passwordInput).toHaveAttribute('type', 'password')
    await expect(page.getByRole('button', { name: '显示密码' })).toHaveAttribute(
      'aria-pressed',
      'false',
    )
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
    await page.getByPlaceholder('账号或邮箱').fill('lawyer')
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

  test('用户菜单应支持键盘打开、焦点管理和关闭', async ({ page }) => {
    await waitForAppShell(page)
    const userButton = page.getByRole('button', { name: '用户菜单：张律师' })

    await userButton.focus()
    await page.keyboard.press('Enter')
    await expect(userButton).toHaveAttribute('aria-expanded', 'true')
    await expect(page.getByRole('menu', { name: '用户菜单' })).toBeFocused()

    await page.keyboard.press('Escape')
    await expect(userButton).toHaveAttribute('aria-expanded', 'false')
    await expect(userButton).toBeFocused()

    await page.keyboard.press('Space')
    await expect(userButton).toHaveAttribute('aria-expanded', 'true')
    await page.mouse.click(600, 300)
    await expect(userButton).toHaveAttribute('aria-expanded', 'false')
    await expect(userButton).toBeFocused()
  })

  test('通知中心应支持键盘打开、关闭和焦点返回', async ({ page }) => {
    await waitForAppShell(page)
    const notificationButton = page.getByRole('button', { name: '通知中心' })

    await notificationButton.focus()
    await page.keyboard.press('Space')
    await expect(notificationButton).toHaveAttribute('aria-expanded', 'true')
    await expect(page.getByRole('menu', { name: '通知中心' })).toBeFocused()

    await page.keyboard.press('Escape')
    await expect(notificationButton).toHaveAttribute('aria-expanded', 'false')
    await expect(notificationButton).toBeFocused()
  })
})
