/**
 * Global setup for Playwright tests.
 * Authenticates and saves storage state for reuse in tests.
 */
import { chromium, type FullConfig } from '@playwright/test';

const AUTH_FILE = 'playwright/.auth/user.json';

async function globalSetup(config: FullConfig) {
    const baseURL = process.env.BASE_URL || config.projects[0]?.use?.baseURL || 'http://localhost:5173';

    console.log(`[Auth Setup] Authenticating against ${baseURL}`);

    const browser = await chromium.launch();
    const page = await browser.newPage();

    try {
        // Navigate to login page
        await page.goto(`${baseURL}/`);

        // Fill login form
        await page.getByRole('textbox', { name: 'Email' }).fill('lee@walker.com');
        await page.getByRole('textbox', { name: 'Password' }).fill('M0untain');
        await page.getByRole('button', { name: 'Sign In' }).click();

        // Wait for navigation to /game
        await page.waitForURL('**/game', { timeout: 10000 });

        console.log('[Auth Setup] Login successful, saving storage state');

        // Save storage state (cookies, localStorage)
        await page.context().storageState({ path: AUTH_FILE });

    } catch (error) {
        console.error('[Auth Setup] Login failed:', error);
        // Don't fail setup - tests will skip if not authenticated
    } finally {
        await browser.close();
    }
}

export default globalSetup;
