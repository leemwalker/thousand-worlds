
import { test, expect } from '@playwright/test';

test.describe('Face Command', () => {
    test('should allow character to face different directions', async ({ page }) => {
        // 1. Create a random user
        const username = `faceuser${Math.floor(Math.random() * 10000)}`;
        const email = `${username}@example.com`;
        const password = 'Password123!';

        // 2. Register
        await page.goto('/');

        // Wait for connection and handling prompts
        await page.waitForLoadState('networkidle');

        // Handle iOS PWA Install Prompt if it appears (on mobile)
        const installPrompt = page.getByText('Install Thousand Worlds');
        if (await installPrompt.isVisible()) {
            await page.getByRole('button', { name: 'Got it' }).click();
        }

        // Wait for the toggle button and click it
        // Adding a small delay to ensure hydration if necessary, though waiting for visible usually suffiies
        await page.waitForTimeout(1000);
        await page.getByRole('button', { name: /Don't have an account?/ }).click();

        // Verify we switched to register mode
        await expect(page.getByRole('heading', { name: 'Create Account' })).toBeVisible();

        await page.getByLabel('Username').fill(username);
        await page.getByLabel('Email').fill(email);
        await page.getByLabel('Password').fill(password);
        await page.getByRole('button', { name: 'Create Account' }).click();

        // 3. Login (if not auto-logged in, but usually signup logs in)
        // Wait for connection or character selection
        // Assuming we land on character selection or world entry
        // If we need to create character:
        try {
            await expect(page.getByText('Create Character')).toBeVisible({ timeout: 5000 });
            await page.getByLabel('Character Name').fill(username);
            await page.getByLabel('Role').selectOption('Explorer');
            // Handle Race if needed, or defaults
            await page.getByRole('button', { name: 'Create Character' }).click();
        } catch (e) {
            // Might already be in lobby if auto-created or specific flow
        }

        // 4. Enter World / Lobby
        // Wait for "Connected to server"
        // Wait for "Connected" indicator
        await expect(page.getByText('Connected', { exact: true })).toBeVisible({ timeout: 10000 });

        // Ensure input field is available
        const input = page.locator('#game-input');
        await expect(input).toBeVisible();

        // 5. Test "face north"
        await input.fill('face north');
        await input.press('Enter');
        await expect(page.locator('#game-output')).toContainText('You face North');

        // 6. Test "face"
        await input.fill('face');
        await input.press('Enter');
        await expect(page.locator('#game-output')).toContainText('You are facing North');

        // 7. Test "face east"
        await input.fill('face east');
        await input.press('Enter');
        await expect(page.locator('#game-output')).toContainText('You face East');

        // 8. Test "face" again
        await input.fill('face');
        await input.press('Enter');
        await expect(page.locator('#game-output')).toContainText('You are facing East');
    });
});
