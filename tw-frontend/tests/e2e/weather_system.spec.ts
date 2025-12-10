import { test, expect } from '@playwright/test';
import { registerNewUser, waitForGameReady, sendCommand } from './fixtures/auth';

test.describe('Weather System', () => {
    test('God Mode Weather Commands', async ({ page }) => {
        // Register and authenticate using centralized helper
        await registerNewUser(page);
        await waitForGameReady(page);

        const terminalInput = page.locator('input[placeholder="Enter command..."]');

        // 1. Initial State - check look
        await sendCommand(page, 'look');
        await page.waitForTimeout(1000);

        // 2. Force Rain
        await sendCommand(page, 'weather rain');
        await expect(page.locator('[data-testid="game-output"]')).toContainText('Weather changed to rain', { timeout: 10000 });

        // 3. Verify Rain in Look
        await sendCommand(page, 'look');
        await expect(page.locator('[data-testid="game-output"]')).toContainText('Rain falls steadily', { timeout: 10000 });

        // 4. Force Storm
        await sendCommand(page, 'weather storm');
        await expect(page.locator('[data-testid="game-output"]')).toContainText('Weather changed to storm', { timeout: 10000 });

        // 5. Verify Storm in Look
        await sendCommand(page, 'look');
        await expect(page.locator('[data-testid="game-output"]')).toContainText('A fierce storm rages', { timeout: 10000 });

        // 6. Force Clear
        await sendCommand(page, 'weather clear');
        await expect(page.locator('[data-testid="game-output"]')).toContainText('Weather changed to clear', { timeout: 10000 });

        // 7. Verify Clear in Look
        await sendCommand(page, 'look');
        await expect(page.locator('[data-testid="game-output"]')).toContainText('The sky is clear', { timeout: 10000 });
    });
});
