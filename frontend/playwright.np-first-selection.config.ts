import { defineConfig } from '@playwright/test'
import isolated from './playwright.np.config'

export default defineConfig({ ...isolated, testDir: './np-entry', testMatch: 'first-selection.spec.ts' })
