import { defineConfig, mergeConfig } from 'vite'
import base from './vite.config'

export default mergeConfig(base, defineConfig({
  cacheDir: '/private/tmp/lawoa-trial-recovery.gwjmmq/output/vite-review-cache',
}))
