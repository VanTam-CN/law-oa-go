import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

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
        target: 'http://localhost:8080',
        changeOrigin: true,
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
      '/file': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        configure: (proxy, options) => {
          proxy.on('error', (err, req, res) => {
            console.error('[PROXY ERROR]:', err.message);
          });
          
          proxy.on('proxyReq', (proxyReq, req, res) => {
            console.log(`[PROXY REQ] ${req.method} ${req.url}`);
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
          router: ['react-router-dom'],
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
        drop_console: true,
        drop_debugger: true,
      },
    },
  },
  optimizeDeps: {
    include: ['react', 'react-dom', 'antd', 'react-router-dom'],
  },
  test: {
    globals: true,
    environment: 'jsdom',
  },
})
