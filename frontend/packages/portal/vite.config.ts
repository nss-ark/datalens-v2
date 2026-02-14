import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
    plugins: [react()],
    resolve: {
        alias: {
            '@': path.resolve(__dirname, './src'),
            '@datalens/shared': path.resolve(__dirname, '../shared/src'),
        },
    },
    server: {
        port: 3002,
        proxy: {
            '/api': {
                target: 'http://localhost:8082',
                changeOrigin: true,
            },
        },
    },
})
