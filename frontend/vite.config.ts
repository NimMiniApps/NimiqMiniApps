import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  server: {
    host: '0.0.0.0',
    port: 5173,
    // Dev proxy: frontend calls same-origin /api, vite forwards to the Go backend.
    // Works from LAN phones too (phone -> vite -> backend), no CORS needed.
    proxy: {
      '/api': 'http://localhost:8080',
      '/health': 'http://localhost:8080',
      '/sitemap.xml': 'http://localhost:8080',
      '/robots.txt': 'http://localhost:8080',
      '/og': 'http://localhost:8080',
    },
  },
  test: {
    // Package emits extensionless ESM imports; Vite must transform it for Node.
    server: {
      deps: {
        inline: ['@nimconnect/profile-client'],
      },
    },
  },
})
