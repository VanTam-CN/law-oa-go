#!/usr/bin/env node
'use strict'
// Read-only preview auth + intake read probe (task np-pr52-preview-read).
// Contract: run with cwd=frontend and NODE_PATH=$PWD/node_modules so that
// require('playwright-core') resolves the frozen dependency tree.
// Inputs (all environment, no secrets printed):
//   NP_BROWSER_SEED      manifest JSON from compare-main-browser-db.py seed
//                        (username/target/target_title/parties/materials are read here)
//   NP_BROWSER_PASSWORD  synthetic lawyer password (>= 12 chars)
//   NP_PREVIEW_EXPECTED  JSON path {id,title,description,revision,actor_id} from a real PG query
//   NP_PREVIEW_ASSET     same-origin built asset path, e.g. /assets/index-XXXX.js,
//                        already byte-compared with the dist file by the caller
//   NP_PREVIEW_BASE      optional preview origin, default http://127.0.0.1:4173
// Output: NP_PREVIEW_* booleans/counts/hashes only. Any failure exits non-zero
// with a single NP_PREVIEW_FAIL=<REASON> marker; no response bodies, tokens or logs.

const { createHash } = require('node:crypto')
const { readFileSync } = require('node:fs')

function fail(reason) {
  console.log('NP_PREVIEW_FAIL=' + reason)
  process.exit(1)
}

function readJsonEnv(name) {
  const path = process.env[name]
  if (!path) fail('MISSING_ENV_' + name)
  try {
    return JSON.parse(readFileSync(path, 'utf8'))
  } catch (err) {
    fail('BAD_JSON_' + name)
  }
}

let playwright
try {
  playwright = require('playwright-core')
} catch (err) {
  fail('PLAYWRIGHT_CORE_UNAVAILABLE')
}

const seed = readJsonEnv('NP_BROWSER_SEED')
const expected = readJsonEnv('NP_PREVIEW_EXPECTED')
const password = process.env.NP_BROWSER_PASSWORD || ''
const base = process.env.NP_PREVIEW_BASE || 'http://127.0.0.1:4173'
const assetPath = process.env.NP_PREVIEW_ASSET || ''

if (password.length < 12) fail('PASSWORD_ENV_SHORT')
if (!seed.username || !seed.target || !seed.target_title) fail('SEED_MANIFEST_INCOMPLETE')
if (!expected || typeof expected.id !== 'string' || typeof expected.title !== 'string' ||
    typeof expected.description !== 'string' || expected.revision === undefined ||
    expected.revision === null || expected.actor_id === undefined) fail('EXPECTED_INCOMPLETE')
if (!assetPath.startsWith('/assets/')) fail('ASSET_ENV_INVALID')
if (expected.id !== seed.target) fail('EXPECTED_SEED_TARGET_MISMATCH')

const descHash = createHash('sha256').update(expected.description).digest('hex')

async function main() {
  const browser = await playwright.chromium.launch({ headless: true })
  try {
    const context = await browser.newContext({ baseURL: base })
    const page = await context.newPage()

    // 1. The page itself must load the exact same-origin built asset.
    const assetUrl = new URL(assetPath, base).toString()
    const assetRespP = page.waitForResponse(r => r.url() === assetUrl, { timeout: 20000 })
    await page.goto('/login', { waitUntil: 'domcontentloaded' })
    const assetResp = await assetRespP
    const assetOrigin = new URL(assetResp.url()).origin
    const pageOrigin = new URL(page.url()).origin
    const assetCT = String(assetResp.headers()['content-type'] || '')
    const assetStatusOK = assetResp.status() === 200
    const assetOriginSame = assetOrigin === pageOrigin
    const assetCTOK = /javascript|ecmascript/i.test(assetCT)
    if (!assetStatusOK || !assetOriginSame || !assetCTOK) {
      fail('ASSET_LOADED status=' + assetResp.status() + ' origin=' + (assetOriginSame ? 'same' : 'diff') + ' ct=' + assetCTOK)
    }
    console.log('NP_PREVIEW_ASSET_LOADED=true')
    console.log('NP_PREVIEW_ASSET_ORIGIN_SAME=true')

    // 2. Real native login; actor id must match the PG expectation.
    const loginP = page.waitForResponse(r => r.request().method() === 'POST' &&
      new URL(r.url()).pathname === '/api/v1/auth/login', { timeout: 20000 })
    await page.getByPlaceholder('账号或邮箱').fill(seed.username)
    await page.getByPlaceholder('密码').fill(password)
    await page.locator('button[type=submit]').click()
    const loginResp = await loginP
    let loginBody = null
    try { loginBody = await loginResp.json() } catch (err) { loginBody = null }
    if (loginResp.status() !== 200 || !loginBody || loginBody.success !== true) {
      fail('LOGIN_HTTP_' + loginResp.status())
    }
    const actorMatch = String(loginBody.data && loginBody.data.user && loginBody.data.user.id) === String(expected.actor_id)
    if (!actorMatch) fail('ACTOR_MISMATCH')
    console.log('NP_PREVIEW_LOGIN_OK=true')
    console.log('NP_PREVIEW_ACTOR_MATCH=true')

    // 3. Native flow into the intake workbench and restore the server draft.
    await page.waitForURL(/dashboard/, { timeout: 20000 })
    await page.getByRole('button', { name: /新建立案$/ }).click()
    await page.getByRole('button', { name: '打开服务端草稿', exact: true }).click()
    const detailP = page.waitForResponse(r => r.request().method() === 'GET' &&
      new URL(r.url()).pathname === '/api/v1/case-intakes/' + seed.target, { timeout: 20000 })
    await page.getByRole('button', { name: seed.target_title, exact: true }).click()
    const detailResp = await detailP
    let detail = null
    try { detail = await detailResp.json() } catch (err) { detail = null }
    if (detailResp.status() !== 200 || !detail || !detail.data || !detail.data.intake) {
      fail('INTAKE_HTTP_' + detailResp.status())
    }

    // 4. Explicit reconciliation of the same-origin API response against PG.
    const intake = detail.data.intake
    if (intake.id !== expected.id) fail('INTAKE_ID_MISMATCH')
    console.log('NP_PREVIEW_INTAKE_ID_MATCH=true')
    if (!(intake.title === expected.title && intake.title === seed.target_title)) fail('TITLE_MISMATCH')
    console.log('NP_PREVIEW_TITLE_MATCH=true')
    if (intake.description !== expected.description) fail('API_DESC_MISMATCH')
    console.log('NP_PREVIEW_DESC_SHA256=' + descHash)
    console.log('NP_PREVIEW_API_DESC_MATCH=true')
    const revisionMatch = intake.revision !== undefined && intake.revision !== null &&
      String(intake.revision) === String(expected.revision)
    if (!revisionMatch) fail('REVISION_MISMATCH')
    console.log('NP_PREVIEW_REVISION_MATCH=true')
    const parties = Array.isArray(detail.data.parties) ? detail.data.parties.length : -1
    const materials = Array.isArray(detail.data.materials) ? detail.data.materials.length : -1
    const expectedParties = Array.isArray(seed.parties) ? seed.parties.length / 2 : -1
    const expectedMaterials = Array.isArray(seed.materials) ? seed.materials.length - 1 : -1
    if (!(parties === expectedParties && parties > 0)) fail('PARTIES_COUNT_' + parties)
    if (!(materials === expectedMaterials && materials > 0)) fail('MATERIALS_COUNT_' + materials)
    console.log('NP_PREVIEW_PARTIES_COUNT=' + parties)
    console.log('NP_PREVIEW_MATERIALS_COUNT=' + materials)

    // 5. Rendered DOM must carry the same selected PG object.
    await page.waitForSelector('#intake-title', { timeout: 20000 })
    const domTitle = await page.inputValue('#intake-title')
    const domDesc = await page.inputValue('#intake-description')
    if (domTitle !== expected.title) fail('DOM_TITLE_MISMATCH')
    if (domDesc !== expected.description) fail('DOM_DESC_MISMATCH')
    console.log('NP_PREVIEW_DOM_TITLE_MATCH=true')
    console.log('NP_PREVIEW_DOM_DESC_HASH_MATCH=true')
    console.log('NP_PREVIEW_OK')
  } finally {
    await browser.close().catch(() => {})
  }
}

main().catch(err => {
  const msg = String((err && err.message) || err).replace(/\s+/g, ' ').slice(0, 300)
  fail('EXCEPTION ' + msg)
})
