import { defineConfig, devices } from '@playwright/test'

const staticPort = 4173

export default defineConfig({
  testDir: '.',
  testMatch: 'c10-sw.spec.ts',
  workers: 1,
  retries: 0,
  timeout: 120_000,
  expect: { timeout: 15_000 },
  reporter: [['list']],
  use: {
    baseURL: `http://127.0.0.1:${staticPort}`,
    actionTimeout: 15_000,
    trace: 'off',
    screenshot: 'off',
    video: 'off',
  },
  webServer: {
    command: 'node c10-sw-static-host.mjs',
    url: `http://127.0.0.1:${staticPort}/healthz`,
    reuseExistingServer: false,
    timeout: 30_000,
  },
  preserveOutput: 'never',
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'], channel: 'chrome' } }],
})
