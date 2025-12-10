import { test, expect } from '@playwright/test';
import { registerNewUser, waitForGameReady, sendCommand } from './fixtures/auth';

test.describe('Watcher Mode', () => {
    test.setTimeout(180000); // Allow time for world creation

    test.beforeEach(async ({ page }) => {
        await registerNewUser(page);
        await waitForGameReady(page);
    });

    test('should allow entering world as watcher', async ({ page }) => {
        const gameOutput = page.locator('[data-testid="game-output"]');
        const worldName = `WatcherWorld_${Date.now()}`;

        // 1. Create a world first (required for E2E testing against real backend)
        await sendCommand(page, 'tell statue create world');
        await page.waitForTimeout(5000);

        // Answer questions rapidly
        const answers = ['Cyber', 'AI', 'City', 'Tech', 'War', worldName];
        for (const ans of answers) {
            await sendCommand(page, `reply ${ans}`);
            await page.waitForTimeout(2000);
        }

        // Wait for creation "forged"
        await expect(gameOutput).toContainText('forged', { timeout: 60000 });

        // 2. Enter the world
        await sendCommand(page, `enter ${worldName}`);

        // 3. Expect Entry Options (Watcher vs Character)
        const watcherButton = page.locator('[data-testid="entry-option-watcher"]').or(
            page.locator('button', { hasText: 'Watcher' })
        );
        await watcherButton.waitFor({ state: 'visible', timeout: 10000 });

        // 4. Click Watcher button
        await watcherButton.click();

        // 5. Verify Entry
        await expect(gameOutput).toContainText('You have entered the world', { timeout: 10000 });

        // 6. Watcher can look
        await sendCommand(page, 'look');
        // Should see some description (don't require specific text)
        await page.waitForTimeout(2000);
    });
});
