import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
    testDir: './tests/e2e',
    timeout: 30 * 1000,
    use: {
        baseURL: 'http://localhost:5173',
        trace: 'on',
    },
    projects: [
        {
            name: 'chromium-debug',
            use: { ...devices['Desktop Chrome'] },
        },
    ],
    // INTENTIONALLY OMITTING webServer to force usage of external server
});
