import { defineConfig } from 'vitest/config';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import path from 'path';

export default defineConfig({
    plugins: [svelte({ hot: !process.env.VITEST })],
    test: {
        globals: true,
        environment: 'jsdom',
        include: ['src/**/*.{test,spec}.{js,ts}'],
        coverage: {
            provider: 'v8',
            reporter: ['text', 'json', 'html'],
            exclude: [
                'node_modules/',
                'src/**/*.spec.ts',
                'src/**/*.test.ts',
                '.svelte-kit/',
                'build/'
            ]
        }
    },
    resolve: {
        alias: {
            $lib: path.resolve('./src/lib')
        }
    }
});
