import fs from 'node:fs'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'
import { VitePWA } from 'vite-plugin-pwa'

const isPages = process.env.GITHUB_PAGES === '1'
const base = isPages ? '/Position_of_planets/' : '/'
const useHttps = process.env.HTTPS === '1'

const httpsConfig = useHttps
  ? {
      key: fs.readFileSync('./certs/cert.key'),
      cert: fs.readFileSync('./certs/cert.crt'),
    }
  : undefined

// https://vite.dev/config/
export default defineConfig({
  base,
  plugins: [
    react(),
    VitePWA({
      registerType: 'autoUpdate',
      devOptions: {
        enabled: true,
      },
      manifest: {
        name: 'Положение планет',
        short_name: 'Планеты',
        description: 'Наблюдения за планетами Солнечной системы',
        lang: 'ru',
        start_url: base,
        scope: base,
        display: 'standalone',
        orientation: 'portrait-primary',
        background_color: '#0a0a0a',
        theme_color: '#0a0a0a',
        icons: [
          { src: 'logo192.png', sizes: '192x192', type: 'image/png' },
          { src: 'logo512.png', sizes: '512x512', type: 'image/png' },
          { src: 'maskable-512.png', sizes: '512x512', type: 'image/png', purpose: 'maskable' },
        ],
      },
      workbox: {
        globPatterns: ['**/*.{js,css,html,svg,webp,mp4}', 'media/*.png'],
        maximumFileSizeToCacheInBytes: 8 * 1024 * 1024,
      },
    }),
  ],
  server: {
    port: 3000,
    strictPort: true,
    host: true,
    allowedHosts: true,
    https: httpsConfig,
    proxy: {
      '/api': { target: 'http://localhost:8080', changeOrigin: true },
      '/minio': {
        target: 'http://localhost:9000',
        changeOrigin: true,
        rewrite: path => path.replace(/^\/minio/, ''),
      },
    },
  },
})
