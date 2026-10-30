import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

// Dev-Server leitet /ws an den Go-Server (hkd, Port 8080) weiter,
// damit Client und Server lokal unter einer Adresse laufen.
// Unterpfad im gemeinsamen Server HanseKaiserReloaded (cmd/hkd): /kaiser/
export default defineConfig({
  base: '/kaiser/',
  // gebauter Client landet im gemeinsamen dist/, das hkd unter /kaiser/ ausliefert
  build: { outDir: '../../dist/kaiser', emptyOutDir: true },
  plugins: [svelte()],
  server: {
    port: 5173,
    proxy: {
      '/kaiser/ws': { target: 'ws://127.0.0.1:8080', ws: true },
      '/kaiser/upload': 'http://127.0.0.1:8080',
      '/kaiser/media': 'http://127.0.0.1:8080',
      '/kaiser/api': 'http://127.0.0.1:8080',
    },
  },
});
