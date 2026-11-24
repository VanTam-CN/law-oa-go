import { defineConfig } from '@playwright/test'
import isolated from './playwright.np.config'

if (process.env.VITE_NP_MEMORY_ONLY !== 'true') throw new Error('M4 requires the memory-only pilot build')
export default defineConfig({ ...isolated, testDir: './np-m4', testMatch: '*.spec.ts' })
