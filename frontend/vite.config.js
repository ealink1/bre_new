import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
    host: true,
    port: 5001,
    strictPort: true,
    proxy: {
      '/api': {
        target: 'https://cj-api.wsky.fun',
        changeOrigin: true
      },
      '/gold-api': {
        target: 'https://www.huilvbiao.com',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/gold-api/, '/api'),
        headers: {
          referer: 'https://www.huilvbiao.com/gold'
        }
      },
      '/silver-api': {
        target: 'https://www.huilvbiao.com',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/silver-api/, '/api'),
        headers: {
          referer: 'https://www.huilvbiao.com/silver'
        }
      }
    }
  },
  preview: {
    host: true,
    port: 5001,
    strictPort: true
  }
})
