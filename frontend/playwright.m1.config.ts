import { defineConfig } from '@playwright/test'

const baseURL = 'http://127.0.0.1:18793'

export default defineConfig({
  testDir: './e2e',
  testIgnore: '**/np-first-slice.live.spec.ts',
  workers: 1,
  timeout: 45_000,
  reporter: 'list',
  use: { baseURL, channel: 'chrome', viewport: { width: 1680, height: 1050 } },
  webServer: {
    command: 'npm run dev -- --host 127.0.0.1 --port 18793 --strictPort',
    url: baseURL,
    reuseExistingServer: false,
  },
})
