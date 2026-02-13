import { defineConfig } from 'vite';

export default defineConfig({
    build: {
        lib: {
            entry: 'src/index.ts',
            name: 'DataLensConsent',
            // Force filename to be consent.min.js
            fileName: () => 'consent.min.js',
            formats: ['iife'],
        },
        outDir: 'dist',
        emptyOutDir: true,
        minify: 'esbuild',
    },
});
