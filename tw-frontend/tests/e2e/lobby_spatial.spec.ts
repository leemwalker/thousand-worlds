import { test, expect } from '@playwright/test';
import { registerNewUser, waitForGameReady, sendCommand } from './fixtures/auth';

test.describe('Lobby Spatial Logic', () => {
    test.setTimeout(180000); // Allow time for world creation

    test.beforeEach(async ({ page }) => {
        await registerNewUser(page);
        await waitForGameReady(page);
    });

    test('Statue Collision and Portal Proximity', async ({ page }) => {
        const gameOutput = page.locator('[data-testid="game-output"]');

        // 1. Ensure we are in lobby
        await sendCommand(page, 'look');
        await expect(gameOutput).toContainText('Grand Lobby', { timeout: 10000 });

        // 2. Test movement in lobby - move north a few times toward statue
        await sendCommand(page, 'north');
        await page.waitForTimeout(500);
        await sendCommand(page, 'n');
        await page.waitForTimeout(500);

        // Try to move to statue position - should be blocked
        await sendCommand(page, 'n');

        // Expect error message about statue blocking
        await expect(gameOutput).toContainText('blocked', { timeout: 5000 });

        // 3. Test Portal Proximity
        // Create a new world for portal testing
        const worldName = `PortalTest_${Date.now()}`;
        await sendCommand(page, 'tell statue create world');
        await page.waitForTimeout(3000);

        // Quick interview answers
        const answers = ['Test', 'Test', 'Test', 'Test', 'Test', worldName];
        for (const ans of answers) {
            await sendCommand(page, `reply ${ans}`);
            await page.waitForTimeout(2000);
        }

        // Wait for world creation
        await expect(gameOutput).toContainText('forged', { timeout: 60000 });

        // Return to lobby
        await sendCommand(page, 'lobby');
        await expect(gameOutput).toContainText('Grand Lobby', { timeout: 10000 });

        // Try to enter the world - may fail if too far from portal
        await sendCommand(page, `enter ${worldName}`);

        // Wait for response
        await page.waitForTimeout(2000);

        // Get the last part of the output to check if proximity error
        const outputText = await gameOutput.textContent();

        if (outputText?.includes('too far')) {
            // Parse direction from error message and move toward portal
            const directions = ['north', 'south', 'east', 'west'];
            for (const dir of directions) {
                if (outputText.toLowerCase().includes(dir)) {
                    // Move toward the portal
                    for (let i = 0; i < 6; i++) {
                        await sendCommand(page, dir);
                        await page.waitForTimeout(300);
                    }
                    break;
                }
            }

            // Try enter again
            await sendCommand(page, `enter ${worldName}`);
        }

        // Should now see entry options or success
        const entryModal = page.locator('text=How do you want to enter?').or(
            page.locator('[data-testid="entry-option-watcher"]')
        );
        await expect(entryModal).toBeVisible({ timeout: 15000 });
    });
});
