import http from 'node:http'
import { readFile } from 'node:fs/promises'
import { extname, join, normalize } from 'node:path'

const port = 4173
const distRoot = new URL('../../dist/', import.meta.url).pathname

const mimeTypes = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'text/javascript; charset=utf-8',
  '.mjs': 'text/javascript; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
}

const server = http.createServer(async (req, res) => {
  const url = new URL(req.url ?? '/', `http://127.0.0.1:${port}`)
  if (url.pathname === '/healthz') {
    res.writeHead(200, { 'content-type': 'text/plain' }).end('ok')
    return
  }
  if (url.pathname.startsWith('/api/v1/')) {
    const upstream = http.request({
      hostname: '127.0.0.1', port: 18789, path: url.pathname + url.search,
      method: req.method, headers: req.headers,
    }, reply => {
      res.writeHead(reply.statusCode, reply.headers)
      reply.pipe(res)
    })
    upstream.on('error', () => res.writeHead(502).end())
    req.pipe(upstream)
    return
  }
  if (url.pathname === '/prod-old-service-worker.js') {
    const body = await readFile(new URL('./prod-old-service-worker.js', import.meta.url))
    res.writeHead(200, { 'content-type': 'text/javascript; charset=utf-8' }).end(body)
    return
  }
  try {
    const rel = normalize(decodeURIComponent(url.pathname)).replace(/^([/\\])+/, '')
    const fallbackPath = !extname(rel) ? 'index.html' : rel
    if (fallbackPath.endsWith('/')) {
      res.writeHead(404, { 'content-type': 'text/plain' }).end('not found')
      return
    }
    const body = await readFile(join(distRoot, fallbackPath))
    res.writeHead(200, { 'content-type': mimeTypes[extname(fallbackPath)] ?? 'application/octet-stream' })
    res.end(body)
  } catch {
    res.writeHead(404, { 'content-type': 'text/plain' }).end('not found')
  }
})

server.listen(port, '127.0.0.1')
