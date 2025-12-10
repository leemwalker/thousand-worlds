import { test, expect } from '@playwright/test';
import { registerNewUser, waitForGameReady, sendCommand } from './fixtures/auth';

test.describe('WorldEntity System E2E Tests', () => {
    test.setTimeout(120000);

    test.beforeEach(async ({ page }) => {
        await registerNewUser(page);
        await waitForGameReady(page);
    });

    test('Statue blocks movement (collision detection)', async ({ page }) => {
        const gameOutput = page.locator('[data-testid="game-output"]');

        // Ensure we are in lobby
        await sendCommand(page, 'look');
        await expect(gameOutput).toContainText('Lobby', { timeout: 10000 });

        // Player spawns at (5,2). Statue is at (5,5). 
        // Move north toward statue
        await sendCommand(page, 'north');
        await page.waitForTimeout(500);
        await sendCommand(page, 'north');
        await page.waitForTimeout(500);

        // Now at ~(5,4), one more should hit the statue
        await sendCommand(page, 'north');

        // Should see blocked message
        await expect(gameOutput).toContainText('blocked', { timeout: 5000 });
        await expect(gameOutput).toContainText('statue', { timeout: 5000 });
    });

    test('Cannot pick up locked statue (interaction check)', async ({ page }) => {
        const gameOutput = page.locator('[data-testid="game-output"]');

        // Ensure we are in lobby
        await sendCommand(page, 'look');
        await expect(gameOutput).toContainText('Lobby', { timeout: 10000 });

        // Try to get the statue
        await sendCommand(page, 'get statue');
        await page.waitForTimeout(500);

        // Should see error message about not being able to move/pick up the statue
        await expect(gameOutput).toContainText('cannot', { timeout: 5000 });
    });

    test('Take command is alias for get', async ({ page }) => {
        const gameOutput = page.locator('[data-testid="game-output"]');

        // Ensure we are in lobby
        await sendCommand(page, 'look');
        await expect(gameOutput).toContainText('Lobby', { timeout: 10000 });

        // Try to take the statue (should use get handler)
        await sendCommand(page, 'take statue');
        await page.waitForTimeout(500);

        // Should see same error as "get statue"
        await expect(gameOutput).toContainText('cannot', { timeout: 5000 });
    });

    test('Look at statue shows description', async ({ page }) => {
        const gameOutput = page.locator('[data-testid="game-output"]');

        // Ensure we are in lobby
        await sendCommand(page, 'look');
        await expect(gameOutput).toContainText('Lobby', { timeout: 10000 });

        // Look at the statue
        await sendCommand(page, 'look statue');
        await page.waitForTimeout(500);

        // Should see statue description
        await expect(gameOutput).toContainText('marble', { timeout: 5000 });
    });

    test('Look at statue shows details when close', async ({ page }) => {
        const gameOutput = page.locator('[data-testid="game-output"]');

        // Ensure we are in lobby
        await sendCommand(page, 'look');
        await expect(gameOutput).toContainText('Lobby', { timeout: 10000 });

        // Move close to statue (spawn at 5,2, statue at 5,5)
        await sendCommand(page, 'north');
        await page.waitForTimeout(300);
        await sendCommand(page, 'north');
        await page.waitForTimeout(300);

        // Now look at the statue - should see detailed description
        await sendCommand(page, 'look statue');
        await page.waitForTimeout(500);

        // Should see the "details" field (mentions runes)
        await expect(gameOutput).toContainText('runes', { timeout: 5000 });
    });

    test('Portal frames are present and block movement', async ({ page }) => {
        const gameOutput = page.locator('[data-testid="game-output"]');

        // Ensure we are in lobby
        await sendCommand(page, 'look');
        await expect(gameOutput).toContainText('Lobby', { timeout: 10000 });

        // Move to the west wall (West Portal is at 1,5)
        // Player spawns at (5,2), so we need to go west 4 times
        await sendCommand(page, 'west');
        await page.waitForTimeout(300);
        await sendCommand(page, 'west');
        await page.waitForTimeout(300);
        await sendCommand(page, 'west');
        await page.waitForTimeout(300);
        await sendCommand(page, 'west');
        await page.waitForTimeout(300);

        // Try to move west again - should hit portal or wall
        await sendCommand(page, 'west');
        await page.waitForTimeout(500);

        // Should be blocked (either by portal or boundary)
        const outputText = await gameOutput.textContent();
        const isBlocked = outputText?.includes('blocked') || outputText?.includes('boundary');
        expect(isBlocked).toBeTruthy();
    });

    test('Can look at portal', async ({ page }) => {
        const gameOutput = page.locator('[data-testid="game-output"]');

        // Ensure we are in lobby  
        await sendCommand(page, 'look');
        await expect(gameOutput).toContainText('Lobby', { timeout: 10000 });

        // Try to look at a portal
        await sendCommand(page, 'look south portal');
        await page.waitForTimeout(500);

        // Should see portal description
        await expect(gameOutput).toContainText('portal', { timeout: 5000 });
    });
});
