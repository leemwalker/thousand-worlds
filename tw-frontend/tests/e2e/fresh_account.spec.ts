import { test, expect } from '@playwright/test';

test.describe('Account Isolation and Fresh Start', () => {
    test('New account should start at default location (West Wing)', async ({ page }) => {
        // 1. Register User A (simulating "previous account")
        const timestamp = Date.now();
        const emailA = `user_a_${timestamp}@example.com`;
        const usernameA = `user_a_${timestamp}`;

        await page.goto('/');

        // Register A
        const toggleButton = page.locator('button').filter({ hasText: "Don't have an account?" });
        if (await toggleButton.isVisible()) {
            await toggleButton.click();
        }

        await page.locator('input[type="email"]').first().fill(emailA);
        const usernameInput = page.getByLabel('Username');
        await usernameInput.waitFor({ state: 'visible', timeout: 10000 });
        await usernameInput.fill(usernameA);
        await page.locator('input[type="password"]').first().fill('Password123!');
        await page.locator('button[type="submit"]', { hasText: 'Create Account' }).click();

        // Wait for game load
        await page.waitForURL(/\/game/);
        await page.waitForSelector('[data-testid="connection-status"].connected', { timeout: 10000 });

        // Wait for initial description
        await expect(page.locator('[data-testid="game-output"]')).toContainText('West Wing');

        // 2. Logout
        await page.getByText('Logout').click();
        await page.waitForURL('/');

        // 3. Register User B (simulating "new account")
        const emailB = `user_b_${timestamp}@example.com`;
        const usernameB = `user_b_${timestamp}`;

        // Switch to Register mode again
        await expect(page.locator('button').filter({ hasText: "Don't have an account?" })).toBeVisible();
        await page.locator('button').filter({ hasText: "Don't have an account?" }).click();

        await page.locator('input[type="email"]').first().fill(emailB);
        const usernameInputB = page.getByLabel('Username');
        await usernameInputB.waitFor({ state: 'visible', timeout: 10000 });
        await usernameInputB.fill(usernameB); // Ensure field is cleared or use fill
        await page.locator('input[type="password"]').first().fill('Password123!');
        await page.locator('button[type="submit"]', { hasText: 'Create Account' }).click();

        // Wait for game load
        await page.waitForURL(/\/game/);
        await page.waitForSelector('[data-testid="connection-status"].connected', { timeout: 10000 });

        // 4. Verify User B is at default location
        // Even if User A moved (which we can't easily simulation fast enough), 
        // User B must start with the initial "West Wing" description.
        // If the session leaked, we might see User A's state or "Welcome back".
        // A fresh account sees "Welcome to Thousand Worlds".

        // Check for welcome message
        const output = page.locator('[data-testid="game-output"]');
        await expect(output).toContainText('Welcome to Thousand Worlds');
        // Note: Returning user sees "Welcome back".

        // Check for location description
        await expect(output).toContainText('West Wing');
    });
});
