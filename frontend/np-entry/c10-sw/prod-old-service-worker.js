const CACHE_NAME = 'np-intake-v1'
const INTAKE_PREFIXES = ['/api/v1/case-intakes', '/api/v1/case-intake-assist', '/api/v1/cases/intake-workbench']

self.addEventListener('install', (event) => {
  event.waitUntil((async () => {
    const cache = await caches.open(CACHE_NAME)
    await cache.put('/api/v1/case-intakes/legacy-canary', new Response('legacy-intake-body', { status: 200 }))
    await cache.put('/api/v1/cases/intake-workbench/legacy-canary', new Response('legacy-workbench-body', { status: 200 }))
  })())
  self.skipWaiting()
})

self.addEventListener('activate', (event) => {
  event.waitUntil(self.clients.claim())
})

self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url)
  if (!INTAKE_PREFIXES.some((prefix) => url.pathname.startsWith(prefix))) return
  event.respondWith((async () => {
    const cache = await caches.open(CACHE_NAME)
    const cached = await cache.match(event.request)
    if (cached) return cached
    return fetch(event.request)
  })())
})
