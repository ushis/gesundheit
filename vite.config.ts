import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  root: 'http/ui',
  server: {
    proxy: {
      '/api/': {
        target: 'http://angerpi:9876',
      },
      '/api/events/socket': {
        target: 'ws://angerpi:9876',
        ws: true,
      },
    }
  }
})
