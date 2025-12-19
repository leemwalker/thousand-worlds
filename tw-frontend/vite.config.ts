import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
    plugins: [
        sveltekit()
        // PWA disabled due to workbox-build ESM/CJS incompatibility in Docker
        // See: https://github.com/vite-pwa/vite-plugin-pwa/issues/655
    ],
    server: {
        host: '0.0.0.0', // Listen on all network interfaces for mobile access
        port: 5173,
        strictPort: true, // Fail if port is already in use
        proxy: {
            '/api': {
                target: 'http://localhost:8080',
                changeOrigin: true,
                ws: true,
                timeout: 120000,
                proxyTimeout: 120000
            }
        }
    }
});
