import { test, expect } from '@playwright/test';
import { registerNewUser, waitForGameReady, ensureLoggedOut, suppressIOSPrompt } from './fixtures/auth';

test.describe('Account Isolation and Fresh Start', () => {
    test('New account should start at default location (West Wing)', async ({ page }) => {
        // 1. Register User A (simulating "previous account")
        await suppressIOSPrompt(page);
        const credsA = await registerNewUser(page);
        await waitForGameReady(page);

        const output = page.locator('[data-testid="game-output"]');

        // Wait for initial description
        await expect(output).toContainText('West Wing', { timeout: 15000 });

        // 2. Logout
        await page.getByText('Logout').click();
        await page.waitForURL('/', { timeout: 10000 });

        // 3. Register User B (simulating "new account")
        await suppressIOSPrompt(page);
        const credsB = await registerNewUser(page);
        await waitForGameReady(page);

        // 4. Verify User B is at default location
        // A fresh account sees "Welcome to Thousand Worlds".
        await expect(output).toContainText('Welcome to Thousand Worlds', { timeout: 15000 });

        // Check for location description
        await expect(output).toContainText('West Wing', { timeout: 10000 });
    });
});
