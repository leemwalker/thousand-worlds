import { test, expect } from '@playwright/test';

test.describe('Movement & D-Pad E2E', () => {
    test.beforeEach(async ({ page }) => {
        // Prevent iOS install prompt from appearing
        await page.addInitScript(() => {
            localStorage.setItem('iosInstallPromptDismissed', String(Date.now()));
        });

        // Load the page
        await page.goto('/');

        // Register unique user
        const timestamp = Date.now();
        const random = Math.floor(Math.random() * 10000);
        const email = `move_test_${timestamp}_${random}@example.com`;
        const username = `move_test_${timestamp}_${random}`;

        // Just go to register page directly if possible, or handle redirect
        if (await page.getByRole('button', { name: "Don't have an account?" }).isVisible()) {
            await page.click('button:has-text("Don\'t have an account?")');
        }

        // Wait specifically for the email input to be interactive
        const emailInput = page.locator('input[type="email"]').first();
        await emailInput.waitFor({ state: 'visible', timeout: 30000 });
        await emailInput.fill(email);

        const usernameInput = page.getByLabel('Username');
        await usernameInput.fill(username);

        await page.locator('input[type="password"]').first().fill('Password123!');

        // Submit and wait for URL change
        await Promise.all([
            page.waitForURL(/\/game/, { timeout: 45000 }),
            page.locator('button[type="submit"]', { hasText: 'Create Account' }).click()
        ]);

        // Wait for Game Input
        await page.waitForSelector('input[placeholder="Enter command..."]', { state: 'attached', timeout: 45000 });

        // Ensure we are fully loaded by waiting for something on the game screen (output container)
        await page.waitForSelector('[data-testid="game-output"]', { state: 'visible', timeout: 30000 });
    });

    test('D-Pad elements exist and trigger commands', async ({ page }) => {
        // Verify D-Pad container
        const dpad = page.locator('div[role="group"][aria-label="Movement Controls"]');
        await expect(dpad).toBeVisible();

        // Verify all 8 directions + Look
        const directions = [
            { label: 'North', command: 'north' },
            { label: 'South', command: 'south' },
            { label: 'East', command: 'east' },
            { label: 'West', command: 'west' },
            { label: 'Northeast', command: 'ne' },
            { label: 'Northwest', command: 'nw' },
            { label: 'Southeast', command: 'se' },
            { label: 'Southwest', command: 'sw' },
            { label: 'Look', command: 'look' },
        ];

        for (const dir of directions) {
            const btn = page.locator(`button[aria-label="${dir.label}"]`);
            await expect(btn).toBeVisible();
            await btn.click();

            // Check output for command feedback (user message)
            // It might be like "> ne" or "You move northeast."
            // The frontend adds "> COMMAND" to the log locally before sending.
            // We check for that.
            const userMsg = page.locator('[data-testid="game-output"] .message', { hasText: `> ${dir.command}` });
            await expect(userMsg).toBeVisible({ timeout: 2000 });
        }
    });

    test('Diagonal movement works via text input', async ({ page }) => {
        const input = page.locator('input[placeholder="Enter command..."]');

        // Test "ne"
        await input.fill('ne');
        await input.press('Enter');
        await expect(page.locator('[data-testid="game-output"] .message', { hasText: '> ne' })).toBeVisible();

        // Test "southwest"
        await input.fill('southwest');
        await input.press('Enter');
        await expect(page.locator('[data-testid="game-output"] .message', { hasText: '> southwest' })).toBeVisible();
    });

    test('Face command works with diagonal directions', async ({ page }) => {
        const input = page.locator('input[placeholder="Enter command..."]');

        // Test "face ne"
        await input.fill('face ne');
        await input.press('Enter');
        await expect(page.locator('[data-testid="game-output"] .message', { hasText: '> face ne' })).toBeVisible();

        // Warning: This tests the command is sent, not necessarily the backend logic (which is unit tested).
        // But verifying no "Unknown command" error or similar would be good if we could spy on system messages.
    });
});
