import { test, expect } from '@playwright/test';
import { registerNewUser, waitForGameReady, sendCommand } from './fixtures/auth';

test.describe('Face Command', () => {
    test('should allow character to face different directions', async ({ page }) => {
        await registerNewUser(page);
        await waitForGameReady(page);

        const gameOutput = page.locator('[data-testid="game-output"]');

        // Test "face north"
        await sendCommand(page, 'face north');
        await expect(gameOutput).toContainText('You face North', { timeout: 10000 });

        // Test "face" to check current facing
        await sendCommand(page, 'face');
        await expect(gameOutput).toContainText('You are facing North', { timeout: 10000 });

        // Test "face east"
        await sendCommand(page, 'face east');
        await expect(gameOutput).toContainText('You face East', { timeout: 10000 });

        // Test "face" again
        await sendCommand(page, 'face');
        await expect(gameOutput).toContainText('You are facing East', { timeout: 10000 });
    });
});
