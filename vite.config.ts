import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  root: 'http/ui',
  server: {
    proxy: {
      '/api/': {
        target: 'http://localhost:8080',
      },
      '/api/events/socket': {
        target: 'ws://localhost:8080',
        ws: true,
      },
    }
  }
})
