import { defineConfig, devices } from '@playwright/test'

const baseURL = process.env.E2E_BASE_URL || 'http://127.0.0.1:3003'
const serverURL = new URL(baseURL)
const serverPort = serverURL.port || (serverURL.protocol === 'https:' ? '443' : '80')
const webServerCommand =
  process.env.E2E_WEB_SERVER_COMMAND ||
  `npm run dev -- --host ${serverURL.hostname} --port ${serverPort}`

export default defineConfig({
  testDir: './e2e',
  timeout: 60_000,
  expect: {
    timeout: 8_000,
  },
  fullyParallel: false,
  workers: 2,
  forbidOnly: Boolean(process.env.CI),
  retries: process.env.CI ? 1 : 0,
  reporter: [
    ['list'],
    ['html', { open: 'never' }],
  ],
  use: {
    baseURL,
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  webServer:
    process.env.E2E_SKIP_WEB_SERVER === '1'
      ? undefined
      : {
          command: webServerCommand,
          url: baseURL,
          reuseExistingServer: !process.env.CI,
          timeout: 120_000,
        },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
})
