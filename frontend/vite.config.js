import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// During `vite dev` the API and WebSocket are proxied to the Go backend so the
// browser only ever talks to one origin.
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      '/api': { target: 'http://localhost:8080', changeOrigin: true },
      '/ws': { target: 'ws://localhost:8080', ws: true },
    },
  },
});
