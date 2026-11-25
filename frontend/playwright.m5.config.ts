import { defineConfig } from '@playwright/test'
import isolated from './playwright.np.config'

if (process.env.ROOT_URL !== 'http://127.0.0.1:18789') {
  throw new Error('M5 requires the isolated root upstream on 18789')
}
if (process.env.VITE_NP_MEMORY_ONLY !== 'true') throw new Error('M5 requires the memory-only pilot build')

export default defineConfig({
  ...isolated,
  testDir: '.',
  testMatch: ['np-m5/*.spec.ts', 'np-m4/*.spec.ts', 'np-entry/first-selection.spec.ts', 'e2e/np-first-slice.live.spec.ts'],
  use: { ...isolated.use, baseURL: 'http://127.0.0.1:3003' },
  webServer: {
    command: 'node ../scripts/np-m5/static-proxy.mjs',
    url: 'http://127.0.0.1:3003/healthz',
    reuseExistingServer: false,
    timeout: 30_000,
  },
})
