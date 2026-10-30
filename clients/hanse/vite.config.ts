import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

// Dev-Server leitet /ws an den Go-Server (hkd, Port 8080) weiter,
// damit Client und Server lokal unter einer Adresse laufen.
// Unterpfad im gemeinsamen Server HanseKaiserReloaded (cmd/hkd): /hanse/
export default defineConfig({
  base: '/hanse/',
  // gebauter Client landet im gemeinsamen dist/, das hkd unter /hanse/ ausliefert
  build: { outDir: '../../dist/hanse', emptyOutDir: true },
  plugins: [svelte()],
  server: {
    port: 5173,
    proxy: {
      '/hanse/ws': { target: 'ws://127.0.0.1:8080', ws: true },
      '/hanse/upload': 'http://127.0.0.1:8080',
      '/hanse/media': 'http://127.0.0.1:8080',
      '/hanse/api': 'http://127.0.0.1:8080',
    },
  },
});
