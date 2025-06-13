import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  base: "/dist/base",
  build: {

    rollupOptions: {
      output: {
        dir: "../../dist/base",
        entryFileNames: "ssidbde-base.js"
      }
    }
  },
  plugins: [
    react(),
    tailwindcss()
  ],
})
