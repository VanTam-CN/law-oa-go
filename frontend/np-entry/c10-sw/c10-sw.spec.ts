import { createHash } from 'node:crypto'
import { loginAndOpenDraft } from '../../e2e/utils/np-live-draft'
import { expect, test } from '@playwright/test'

const staticBase = 'http://127.0.0.1:4173'

async function probe(pathname: string) {
  const response = await fetch(staticBase + pathname)
  return {
    status: response.status,
    contentType: response.headers.get('content-type') ?? '',
    body: await response.text(),
  }
}

test('C10-PROD clean profile: real dist loads login UI, register calls stay zero, sw.js probe 404', async ({ page }) => {
  const swProbe = await probe('/sw.js')
  expect(swProbe.status).toBe(404)
  expect(swProbe.contentType).not.toContain('javascript')

  const indexResponse = await page.request.get(staticBase + '/index.html')
  expect(indexResponse.status()).toBe(200)
  const indexHtml = await indexResponse.text()
  const assetPaths = [...new Set([...indexHtml.matchAll(/\/assets\/[A-Za-z0-9_.-]+\.js/g)].map((match) => match[0]))]
  expect(assetPaths.length).toBeGreaterThan(0)
  const bundles: Buffer[] = []
  for (const assetPath of assetPaths) {
    const assetResponse = await page.request.get(staticBase + assetPath)
    expect(assetResponse.status()).toBe(200)
    bundles.push(await assetResponse.body())
  }
  const bundleDigest = createHash('sha256').update(Buffer.concat(bundles)).digest('hex')

  await page.addInitScript(() => {
    ;(window as any).__c10RegisterCalls = 0
    const originalRegister = navigator.serviceWorker.register.bind(navigator.serviceWorker)
    Object.defineProperty(navigator.serviceWorker, 'register', {
      configurable: true,
      value: (...args: Parameters<ServiceWorkerContainer['register']>) => {
        ;(window as any).__c10RegisterCalls += 1
        return originalRegister(...args)
      },
    })
  })

  await page.goto('/index.html')
  await expect(page.getByPlaceholder('账号或邮箱')).toBeVisible()
  await expect(page.locator('[data-np-sw-hold]')).toHaveCount(0)

  expect(await page.evaluate(() => (window as any).__c10RegisterCalls)).toBe(0)
  const runtime = await page.evaluate(async () => {
    const controller = navigator.serviceWorker.controller
    const registrations = await navigator.serviceWorker.getRegistrations()
    let probeErrored: string | null = null
    try {
      await navigator.serviceWorker.register('/sw.js')
    } catch (error) {
      probeErrored = String(error)
    }
    const cacheNames = await caches.keys()
    return {
      controllerScript: controller?.scriptURL ?? null,
      registrationCount: registrations.length,
      probeErrored,
      cacheNames,
    }
  })

  const registerCalls = await page.evaluate(() => (window as any).__c10RegisterCalls as number)
  expect(registerCalls).toBe(1)

  expect(runtime.controllerScript).toBeNull()
  expect(runtime.registrationCount).toBe(0)
  expect(runtime.probeErrored).toBeTruthy()
  expect(runtime.cacheNames).toEqual([])
  console.log('C10-PROD bundle-sha256=' + bundleDigest + ' app-register=0 sw-probe=failed')
})

test('C10-OLD controlled takeover: HOLD UI, zero sensitive requests, cache evidence preserved', async ({ browser }) => {
  const context = await browser.newContext()
  const page = await context.newPage()

  const swResponse = await page.request.get(staticBase + '/prod-old-service-worker.js')
  expect(swResponse.status()).toBe(200)
  expect(swResponse.headers()['content-type']).toContain('javascript')

  const sensitiveRequests: string[] = []
  page.on('request', (request) => {
    const pathname = new URL(request.url()).pathname
    if (pathname.startsWith('/api/v1/case-intakes') || pathname.startsWith('/api/v1/case-intake-assist') || pathname.startsWith('/api/v1/cases/intake-workbench')) {
      sensitiveRequests.push(pathname)
    }
  })

  await loginAndOpenDraft(page)
  expect(sensitiveRequests.length).toBeGreaterThan(0)
  sensitiveRequests.length = 0
  await page.evaluate(async () => {
    await navigator.serviceWorker.register('/prod-old-service-worker.js', { scope: '/' })
  })
  const controller = await page.waitForFunction(() => navigator.serviceWorker.controller !== null)
  expect(controller).toBeTruthy()

  const takeover = await page.evaluate(async () => {
    await navigator.serviceWorker.ready
    const registration = await navigator.serviceWorker.getRegistration()
    const cacheNames = await caches.keys()
    return {
      controllerScript: navigator.serviceWorker.controller?.scriptURL ?? null,
      scope: registration?.scope ?? null,
      cacheNames,
    }
  })
  expect(takeover.controllerScript).toContain('/prod-old-service-worker.js')
  expect(takeover.scope).toBe(staticBase + '/')
  expect(takeover.cacheNames).toContain('np-intake-v1')

  await expect(page.locator('[data-np-sw-hold="blocked"]')).toBeVisible()
  expect(await page.evaluate(() => navigator.serviceWorker.controller !== null)).toBe(true)

  await page.reload()
  await expect(page.locator('[data-np-sw-hold="blocked"]')).toBeVisible()
  await expect(page.locator('#intake-title')).toHaveCount(0)

  const cacheStillPresent = await page.evaluate(async () => {
    const cache = await caches.open('np-intake-v1')
    const keys = await cache.keys()
    const cached = await cache.match('/api/v1/case-intakes/legacy-canary')
    return {
      cacheNames: await caches.keys(),
      keys: keys.map((request) => new URL(request.url).pathname),
      canaryBody: cached ? await cached.text() : null,
    }
  })
  expect(cacheStillPresent.cacheNames).toEqual(['np-intake-v1'])
  expect(cacheStillPresent.keys).toContain('/api/v1/case-intakes/legacy-canary')
  expect(cacheStillPresent.keys).toContain('/api/v1/cases/intake-workbench/legacy-canary')
  expect(cacheStillPresent.canaryBody).toBe('legacy-intake-body')
  expect(sensitiveRequests).toEqual([])

  await context.close()
})
