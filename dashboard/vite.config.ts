import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: { // Optional: configure server if needed
    port: 3001, // To avoid conflict with API gateway if run locally
    // proxy: { // Optional: proxy API requests
    //   '/api': {
    //     target: 'http://localhost:3000', // Your API gateway address
    //     changeOrigin: true,
    //     // rewrite: (path) => path.replace(/^\/api/, '')
    //   }
    // }
  }
}) 