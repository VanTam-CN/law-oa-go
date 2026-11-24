import { defineConfig, devices } from '@playwright/test'

if (process.env.VITE_API_PROXY_TARGET !== 'http://127.0.0.1:18789') {
  throw new Error('NP tests require the isolated loopback backend on 18789')
}

export default defineConfig({
  testDir: './e2e',
  testMatch: 'np-first-slice.live.spec.ts',
  workers: 1,
  retries: 0,
  timeout: 90_000,
  expect: { timeout: 12_000 },
  reporter: [['list']],
  use: { baseURL: 'http://127.0.0.1:3003', actionTimeout: 15_000, trace: 'off', screenshot: 'off', video: 'off' },
  webServer: {
    command: 'npm run dev -- --host 127.0.0.1 --port 3003 --strictPort',
    url: 'http://127.0.0.1:3003',
    reuseExistingServer: false,
    timeout: 120_000,
  },
  preserveOutput: 'never',
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'], channel: 'chrome' } }],
})
