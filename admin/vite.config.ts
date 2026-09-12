import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

const apiProxy = { '/api': 'http://localhost:8088' }

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5174,
    proxy: apiProxy,
  },
  preview: {
    port: 5174,
    proxy: apiProxy,
  },
})
