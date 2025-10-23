/**
 * 改进的静态文件服务器 - 支持SPA路由
 * 为单页应用提供所有路由的fallback到index.html
 */

const http = require('http');
const fs = require('fs');
const path = require('path');
const url = require('url');

const distPath = path.join(__dirname, 'frontend/dist');

// MIME类型映射
const mimeTypes = {
  '.html': 'text/html',
  '.js': 'text/javascript',
  '.css': 'text/css',
  '.json': 'application/json',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.svg': 'image/svg+xml',
  '.ico': 'image/x-icon',
  '.woff': 'font/woff',
  '.woff2': 'font/woff2',
  '.ttf': 'font/ttf',
  '.eot': 'application/vnd.ms-fontobject',
};

const server = http.createServer((req, res) => {
  // 解析URL
  const parsedUrl = url.parse(req.url, true);
  let pathname = parsedUrl.pathname;

  // 处理根路径
  if (pathname === '/') {
    pathname = '/index.html';
  }

  // 对于SPA，所有非静态文件路径都指向index.html
  const extname = path.extname(pathname);
  const isStaticFile = extname && mimeTypes[extname];

  let filePath;
  if (isStaticFile) {
    // 静态文件，使用原始路径
    filePath = path.join(distPath, pathname);
  } else {
    // SPA路由，返回index.html
    filePath = path.join(distPath, 'index.html');
  }

  // 设置CORS头
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type, Authorization');

  // 处理OPTIONS请求
  if (req.method === 'OPTIONS') {
    res.writeHead(200);
    res.end();
    return;
  }

  const contentType = mimeTypes[extname] || 'text/html';

  // 读取文件
  fs.readFile(filePath, 'utf8', (error, content) => {
    if (error) {
      if (error.code === 'ENOENT') {
        // 文件不存在，返回404或SPA fallback
        if (isStaticFile) {
          res.writeHead(404, { 'Content-Type': 'text/plain' });
          res.end('File not found');
        } else {
          // SPA路由，返回index.html
          fs.readFile(path.join(distPath, 'index.html'), 'utf8', (fallbackError, fallbackContent) => {
            if (fallbackError) {
              res.writeHead(500, { 'Content-Type': 'text/plain' });
              res.end('Server error');
            } else {
              res.writeHead(200, { 'Content-Type': 'text/html' });
              res.end(fallbackContent);
            }
          });
        }
      } else {
        // 其他错误
        console.error('Server error:', error);
        res.writeHead(500, { 'Content-Type': 'text/plain' });
        res.end('Server error');
      }
    } else {
      // 成功读取文件
      res.writeHead(200, { 'Content-Type': contentType });
      res.end(content);
    }
  });
});

const PORT = process.env.PORT || 3003;
server.listen(PORT, () => {
  console.log(`🚀 Frontend server running on http://localhost:${PORT}`);
  console.log(`📁 Serving static files from: ${distPath}`);
  console.log(`🔄 SPA routing enabled - all routes fallback to index.html`);
});