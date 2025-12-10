import { test, expect } from '@playwright/test';
import { registerNewUser, waitForGameReady, sendCommand, loginUser, ensureLoggedOut } from './fixtures/auth';

test.describe('Lobby Command & Persistence', () => {
    test.setTimeout(180000); // Allow extra time for world creation flow

    test('should return to lobby and persist location on relogin', async ({ page }) => {
        // Register and authenticate
        const creds = await registerNewUser(page);
        await waitForGameReady(page);

        const gameOutput = page.locator('[data-testid="game-output"]');

        // 1. Start world creation interview
        await sendCommand(page, 'tell statue create world');
        await page.waitForTimeout(3000);

        // Answer interview questions
        const worldName = `LobbyWorld_${Date.now()}`;
        const answers = ['LobbyTest', 'Test', 'Zone', 'Magic', 'Power', worldName];
        for (const ans of answers) {
            await sendCommand(page, `reply ${ans}`);
            await page.waitForTimeout(2000);
        }

        // Wait for "forged" or world creation confirmation
        await expect(gameOutput).toContainText('forged', { timeout: 60000 });

        // 2. Enter World
        await sendCommand(page, `enter ${worldName}`);

        // Click Watcher option if modal appears
        const watcherButton = page.locator('[data-testid="entry-option-watcher"]');
        if (await watcherButton.isVisible({ timeout: 10000 }).catch(() => false)) {
            await watcherButton.click();
        }

        // Verify entry
        await expect(gameOutput).toContainText('You have entered the world', { timeout: 15000 });

        // 3. Send Lobby Command
        await sendCommand(page, 'lobby');

        // 4. Verify Return to Lobby
        await expect(gameOutput).toContainText('You return to the Grand Lobby', { timeout: 10000 });

        // 5. Verify session persistence on reload
        await page.reload();
        await waitForGameReady(page);

        // Check if we are in Lobby by looking
        await sendCommand(page, 'look');
        await expect(gameOutput).toContainText('Grand Lobby', { timeout: 10000 });

        // 6. Test logout and login persistence
        await ensureLoggedOut(page);
        await loginUser(page, creds.email, creds.password);
        await waitForGameReady(page);

        // Verify we're back in the lobby after re-login
        await sendCommand(page, 'look');
        await expect(gameOutput).toContainText('Grand Lobby', { timeout: 10000 });
    });
});
