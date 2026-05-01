import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import { resolve } from 'path';

export default defineConfig({
  plugins: [svelte()],
  resolve: {
    alias: {
      '$lib': resolve(__dirname, './src'),
      '@': resolve(__dirname, './src')
    }
  },
  build: {
    lib: {
      entry: resolve(__dirname, 'src/index.ts'),
      formats: ['es'],
      fileName: 'index'
    },
    rollupOptions: {
      external: [
        'svelte', 'svelte/internal', 'svelte/store',
        'xlsx', 'jspdf', 'jspdf-autotable',
        '@samavāya/stores', '@samavāya/core', '@samavāya/proto',
      ]
    }
  }
});
