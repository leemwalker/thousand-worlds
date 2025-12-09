import { defineConfig, devices } from '@playwright/test';

/**
 * Comprehensive Playwright configuration for E2E testing
 * Tests against local dev server with multiple browsers and devices
 */
export default defineConfig({
    testDir: './tests/e2e',

    /* Maximum time one test can run */
    timeout: 60 * 1000,

    /* Run tests in files in parallel */
    fullyParallel: true,

    /* Fail the build on CI if you accidentally left test.only */
    forbidOnly: !!process.env.CI,

    /* Retry on CI only */
    retries: process.env.CI ? 2 : 0,

    /* Opt out of parallel tests on CI */
    workers: process.env.CI ? 1 : undefined,

    /* Reporter to use */
    reporter: [
        ['html'],
        ['list'],
        ['json', { outputFile: 'test-results/results.json' }]
    ],

    /* Shared settings for all projects */
    use: {
        /* Base URL to use in actions like `await page.goto('/')` */
        baseURL: 'http://localhost:5173',

        /* Collect trace when retrying the failed test */
        trace: 'on-first-retry',

        /* Screenshot on failure */
        screenshot: 'only-on-failure',

        /* Video on failure */
        video: 'retain-on-failure',

        /* Navigation wait strategy */
        actionTimeout: 20000,
        navigationTimeout: 30000,
    },

    /* Configure projects for major browsers and devices */
    projects: [
        {
            name: 'chromium-desktop',
            use: {
                ...devices['Desktop Chrome'],
                viewport: { width: 1920, height: 1080 }
            },
        },

        {
            name: 'firefox-desktop',
            use: {
                ...devices['Desktop Firefox'],
                viewport: { width: 1920, height: 1080 }
            },
        },

        {
            name: 'webkit-desktop',
            use: {
                ...devices['Desktop Safari'],
                viewport: { width: 1920, height: 1080 }
            },
        },

        /* Mobile browsers */
        {
            name: 'iphone-12',
            use: {
                ...devices['iPhone 12']
            },
        },

        {
            name: 'iphone-14-pro',
            use: {
                ...devices['iPhone 14 Pro']
            },
        },

        {
            name: 'iphone-se',
            use: {
                ...devices['iPhone SE']
            },
        },

        {
            name: 'pixel-5',
            use: {
                ...devices['Pixel 5']
            },
        },

        {
            name: 'galaxy-s9',
            use: {
                ...devices['Galaxy S9+']
            },
        },

        /* Tablet */
        {
            name: 'ipad-pro',
            use: {
                ...devices['iPad Pro']
            },
        },
    ],

    /* Run your local dev server before starting the tests */
    webServer: {
        command: 'npm run dev',
        url: 'http://localhost:5173',
        reuseExistingServer: true, // Use existing server if already running
        timeout: 120 * 1000,
    },
});
