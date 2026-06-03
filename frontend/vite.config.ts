import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

const apiProxyTarget = process.env.VITE_API_PROXY_TARGET || 'http://localhost:8080'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 3003,
    host: true,
    strictPort: true,
    proxy: {
      '/api': {
        target: apiProxyTarget,
        changeOrigin: true,
        secure: false,
        rewrite: (path) => path,
        configure: (proxy, options) => {
          proxy.on('error', (err, req, res) => {
            console.error('[PROXY ERROR]:', err.message);
          });

          proxy.on('proxyReq', (proxyReq, req, res) => {
            console.log(`[PROXY REQ] ${req.method} ${req.url} -> ${proxyReq.protocol}//${proxyReq.host}${proxyReq.path}`);
          });

          proxy.on('proxyRes', (proxyRes, req, res) => {
            console.log(`[PROXY RES] ${proxyRes.statusCode} for ${req.url}`);
          });
        },
      },
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom'],
          antd: ['antd'],
          router: ['react-router'],
          utils: ['lodash', 'dayjs'],
        },
      },
    },
    chunkSizeWarningLimit: 1000,
    cssCodeSplit: true,
    sourcemap: false,
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: false,
        drop_debugger: false,
      },
      mangle: {
        reserved: ['message', 'console', 'error'],
      },
    },
  },
  optimizeDeps: {
    include: ['react', 'react-dom', 'antd', 'react-router'],
  },
  test: {
    globals: true,
    environment: 'jsdom',
  },
})
