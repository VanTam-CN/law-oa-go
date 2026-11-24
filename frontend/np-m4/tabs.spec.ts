import { expect, test, type Page, type Route } from '@playwright/test'
import { loginAndOpenDraft, seed, uuid } from '../e2e/utils/np-live-draft'

test('C07 同源双页在登记前均读空，已提交的 A 不清除待回查的 B', async ({ context, page: a }) => {
  const b = await context.newPage()
  await loginAndOpenDraft(a)
  await loginAndOpenDraft(b)
  const prefix = `np-save-v2:${seed.actor_id}:${seed.target}:`
  const records = (p: Page) => p.evaluate(prefix =>
    Object.keys(localStorage).filter(k => k.startsWith(prefix)).sort(), prefix)
  expect(await records(a)).toEqual([])
  expect(await records(b)).toEqual([])

  // Both renderers reach native setItem after completing the application's empty-slot check.
  // Release them together; the browser still performs the real reads, writes and storage events.
  const arrivals = new Set<string>()
  let release!: () => void
  const gate = new Promise<void>(resolve => { release = resolve })
  let timer: ReturnType<typeof setTimeout> | undefined
  await context.route('**/__m4_record_barrier/*', async route => {
    arrivals.add(new URL(route.request().url()).pathname)
    if (arrivals.size === 1) timer = setTimeout(release, 12_000)
    else release()
    await gate
    await route.fulfill({ status: arrivals.size === 2 ? 204 : 408, body: '' })
  })
  const install = (page: Page, tab: string) => page.evaluate(({ prefix, tab }) => {
    const set = Storage.prototype.setItem
    Storage.prototype.setItem = function (key, value) {
      if (this === localStorage && key.startsWith(prefix)) {
        const barrier = new XMLHttpRequest()
        barrier.open('GET', '/__m4_record_barrier/' + tab, false)
        barrier.send()
        if (barrier.status !== 204) throw new Error('both tabs did not reach registration')
      }
      return set.call(this, key, value)
    }
  }, { prefix, tab })
  const intakePath = '**/api/v1/case-intakes/' + seed.target
  let acknowledge!: () => void
  const ackGate = new Promise<void>(resolve => { acknowledge = resolve })
  let keyA = '', keyB = ''
  const intercept = (tab: 'A' | 'B') => async (route: Route) => {
    const req = route.request()
    if (req.method() !== 'PUT') return route.continue()
    const key = req.headers()['idempotency-key']
    if (tab === 'B') {
      keyB = key
      return route.abort('connectionreset')
    }
    keyA = key
    const reply = await route.fetch()
    expect(reply.status()).toBe(200)
    await ackGate
    return route.fulfill({ response: reply })
  }
  await a.route(intakePath, intercept('A'))
  await b.route(intakePath, intercept('B'))
  try {
    await Promise.all([a.locator('#intake-description').fill('双页合成 A'), b.locator('#intake-description').fill('双页合成 B')])
    await Promise.all([install(a, 'A'), install(b, 'B')])
    await Promise.all([a,b].map(p => p.getByRole('button', { name: '保存草稿', exact: true }).click()))
    await expect.poll(() => [keyA, keyB].every(k => uuid.test(k))).toBe(true)
    expect(arrivals.size).toBe(2)
    expect(keyA).not.toBe(keyB)
    expect(await records(b)).toEqual([prefix + keyA, prefix + keyB].sort())
  } finally {
    release()
    clearTimeout(timer)
    acknowledge()
  }
  await expect.poll(() => records(a)).toEqual([prefix + keyB])
  expect(await records(b)).toEqual([prefix + keyB])

  b.once('dialog', dialog => dialog.accept())
  await b.reload()
  await b.getByRole('button', { name: '打开服务端草稿', exact: true }).click()
  await b.getByRole('button', { name: seed.target_title, exact: true }).click()
  const endpoint = '/write-results/' + keyB
  const unknown = b.waitForResponse(r => r.url().endsWith(endpoint))
  await b.getByRole('button', { name: '回查原保存结果' }).click()
  expect((await (await unknown).json()).data).toMatchObject({ state: 'unknown', request_id: keyB })
  expect(await records(a)).toEqual([prefix + keyB])
  const ended = b.waitForResponse(r => r.url().endsWith(endpoint + '/abandon'))
  await b.getByRole('button', { name: '结束原保存并核实结果' }).click()
  await b.getByRole('button', { name: /确\s*定|^OK$/ }).last().click()
  expect((await (await ended).json()).data).toMatchObject({ state: 'abandoned', request_id: keyB, id: seed.target })
  await expect.poll(() => records(a)).toEqual([])
})
