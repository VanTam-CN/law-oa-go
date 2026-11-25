import http from 'node:http'
import { createHash } from 'node:crypto'
import { appendFile, readFile } from 'node:fs/promises'
import { extname, join, normalize } from 'node:path'

const upstream = 'http://127.0.0.1:18789'
if (process.env.ROOT_URL !== upstream) throw new Error('M5 requires the isolated root on 18789')
const port = Number(process.env.NP_M5_PROXY_PORT ?? 3003)
if (![3003, 3004].includes(port)) throw new Error('M5 proxy requires browser port 3003 or probe port 3004')
const eventFile = process.env.NP_M5_PROXY_EVENT_FILE
const distRoot = new URL('../../frontend/dist/', import.meta.url).pathname
const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i
const intakePath = /^\/api\/v1\/case-intakes\/([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$/i
const mimeTypes = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'text/javascript; charset=utf-8',
  '.mjs': 'text/javascript; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
}

async function logEvent(event) {
  const line = JSON.stringify(event)
  console.log(line)
  if (eventFile) await appendFile(eventFile, line + '\n')
}

function failFast(error) {
  console.error('np-m5 proxy event persistence failed:', error instanceof Error ? error.message : error)
  process.exit(1)
}

function receiptCheck(key, authorization) {
  return new Promise(resolve => {
    const req = http.request(upstream + '/api/v1/case-intakes/write-results/' + key, {
      method: 'GET',
      headers: authorization ? { authorization } : {},
      timeout: 5000,
    }, reply => {
      let body = ''
      reply.on('data', chunk => { body += chunk })
      reply.on('end', () => {
        let data
        try { data = JSON.parse(body).data } catch { data = undefined }
        resolve({ status: reply.statusCode, data })
      })
    })
    req.on('timeout', () => req.destroy(new Error('receipt timeout')))
    req.on('error', () => resolve({ status: 0, data: undefined }))
    req.end()
  })
}

async function receiptEvent(base) {
  const receipt = await receiptCheck(base.key, base.authorization)
  const keyMatch = receipt.data?.request_id === base.key
  const idMatch = receipt.data == null ? null : receipt.data.id === base.putId
  await logEvent({
    ...base.event,
    receipt_status: receipt.status,
    receipt_key_match: keyMatch,
    receipt_id_match: idMatch,
    receipt_committed: keyMatch && idMatch === true && receipt.data?.state === 'committed',
  })
}

function forward(req, res, headers) {
  const upstreamReq = http.request(upstream + req.url, { method: req.method, headers, timeout: 15000 }, reply => {
    res.writeHead(reply.statusCode, reply.headers)
    reply.pipe(res)
  })
  upstreamReq.on('timeout', () => upstreamReq.destroy(new Error('upstream timeout')))
  upstreamReq.on('error', () => {
    if (!res.headersSent) res.writeHead(502)
    res.end()
  })
  req.pipe(upstreamReq)
}

async function afterCommitDrop(req, res, headers, key, putId) {
  const upstreamStatus = await new Promise(resolve => {
    const upstreamReq = http.request(upstream + req.url, { method: req.method, headers, timeout: 15000 }, reply => {
      reply.resume()
      reply.on('end', () => resolve(reply.statusCode))
    })
    upstreamReq.on('timeout', () => upstreamReq.destroy(new Error('upstream timeout')))
    upstreamReq.on('error', () => resolve(0))
    req.pipe(upstreamReq)
  })
  let dropped = false
  if (upstreamStatus === 200) {
    res.destroy()
    dropped = true
  } else {
    res.writeHead(upstreamStatus || 502)
    res.end()
  }
  await receiptEvent({
    key,
    putId,
    authorization: headers.authorization,
    event: {
      ts: new Date().toISOString(),
      stage: 'after-commit',
      key_sha256: createHash('sha256').update(key).digest('hex'),
      upstream_status: upstreamStatus || null,
      client_dropped: dropped,
    },
  })
}

async function beforeForwardDrop(res, headers, key, putId) {
  res.destroy()
  await receiptEvent({
    key,
    putId,
    authorization: headers.authorization,
    event: {
      ts: new Date().toISOString(),
      stage: 'before-forward',
      key_sha256: createHash('sha256').update(key).digest('hex'),
      upstream_status: null,
      client_dropped: true,
    },
  })
}

const server = http.createServer((req, res) => {
  const url = new URL(req.url ?? '/', 'http://127.0.0.1:' + port)
  if (url.pathname === '/healthz') {
    res.writeHead(200, { 'content-type': 'text/plain' }).end('ok')
    return
  }
  if (url.pathname.startsWith('/api/v1/')) {
    const drop = req.headers['x-np-drop-response']
    const key = req.headers['idempotency-key']
    const putId = intakePath.exec(url.pathname)?.[1]
    const validMarkedPut = req.method === 'PUT' && putId !== undefined &&
      (drop === 'after-commit' || drop === 'before-forward') &&
      typeof key === 'string' && uuidPattern.test(key)
    if (drop && !validMarkedPut) {
      res.writeHead(400, { 'content-type': 'application/json' }).end(JSON.stringify({ error: 'invalid fault injection' }))
      return
    }
    const headers = { ...req.headers }
    delete headers['x-np-drop-response']
    if (drop === 'before-forward') {
      beforeForwardDrop(res, headers, key, putId).catch(failFast)
    } else if (drop === 'after-commit') {
      afterCommitDrop(req, res, headers, key, putId).catch(failFast)
    } else {
      forward(req, res, headers)
    }
    return
  }
  void (async () => {
    try {
      const rel = normalize(decodeURIComponent(url.pathname)).replace(/^([/\\])+/, '')
      if (rel.includes('..')) throw new Error('traversal')
      const path = !extname(rel) ? 'index.html' : rel
      if (path.endsWith('/')) throw new Error('directory')
      const body = await readFile(join(distRoot, path))
      res.writeHead(200, { 'content-type': mimeTypes[extname(path)] ?? 'application/octet-stream' })
      res.end(body)
    } catch {
      res.writeHead(404, { 'content-type': 'text/plain' }).end('not found')
    }
  })()
})

server.headersTimeout = 10_000
server.requestTimeout = 30_000
server.keepAliveTimeout = 5_000
server.connectionsCheckingInterval = 1_000
server.listen(port, '127.0.0.1', () => {
  console.error('np-m5 proxy 127.0.0.1:' + port + ' -> ' + upstream)
})
