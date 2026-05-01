import { defineConfig } from 'vite'
import { resolve } from 'path'
import dts from 'vite-plugin-dts'

export default defineConfig({
  plugins: [
    dts({
      insertTypesEntry: true,
      rollupTypes: true,
      exclude: ['**/*.test.ts', '**/*.spec.ts']
    })
  ],
  
  build: {
    lib: {
      entry: {
        index: resolve(__dirname, 'src/index.ts'),
        'validation/index': resolve(__dirname, 'src/validation/index.ts'),
        'formatting/index': resolve(__dirname, 'src/formatting/index.ts'),
        'date/index': resolve(__dirname, 'src/date/index.ts'),
        'file/index': resolve(__dirname, 'src/file/index.ts'),
        'string/index': resolve(__dirname, 'src/string/index.ts'),
        'number/index': resolve(__dirname, 'src/number/index.ts')
      },
      name: 'P9eUtils',
      formats: ['es']
    },
    
    rollupOptions: {
      external: [],
      
      output: {
        preserveModules: false
      }
    },
    
    minify: 'esbuild',
    sourcemap: true,
    emptyOutDir: true
  },
  
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  }
})