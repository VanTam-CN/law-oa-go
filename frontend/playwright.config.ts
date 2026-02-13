import { defineConfig, devices } from '@playwright/test'

/**
 * Playwright E2E测试配置
 * Story 6.1: Playwright配置
 */
export default defineConfig({
  // 测试目录
  testDir: './e2e',

  // 完全并行运行测试
  fullyParallel: true,

  // CI上失败时禁止test.only
  forbidOnly: !!process.env.CI,

  // CI上重试失败测试
  retries: process.env.CI ? 2 : 0,

  // CI上限制并行workers
  workers: process.env.CI ? 1 : undefined,

  // Reporter配置
  reporter: [
    ['html', { outputFolder: 'playwright-report' }],
    ['list'],
  ],

  // 全局测试配置
  use: {
    // 基础URL
    baseURL: 'http://localhost:5173',

    // 收集失败测试的trace
    trace: 'on-first-retry',

    // 截图
    screenshot: 'only-on-failure',

    // 视频录制
    video: 'retain-on-failure',
  },

  // 配置项目（浏览器）
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },
    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },
    // 移动端测试
    {
      name: 'Mobile Chrome',
      use: { ...devices['Pixel 5'] },
    },
    {
      name: 'Mobile Safari',
      use: { ...devices['iPhone 12'] },
    },
  ],

  // 本地开发服务器
  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:5173',
    reuseExistingServer: !process.env.CI,
    timeout: 120 * 1000,
  },
})
