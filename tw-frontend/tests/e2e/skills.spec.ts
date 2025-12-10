import { test, expect } from '@playwright/test';
import { registerNewUser, waitForGameReady, sendCommand } from './fixtures/auth';

test.describe('Player Skills UI', () => {
    test('should display skills in character sheet', async ({ page }) => {
        await registerNewUser(page);
        await waitForGameReady(page);

        const gameOutput = page.locator('[data-testid="game-output"]');

        // Create character via command
        await sendCommand(page, 'create character SkillTester Human');

        // Wait for game to load
        await expect(gameOutput).toContainText('You have entered the world', { timeout: 30000 });

        // Check skills command
        await sendCommand(page, 'skills');

        // Wait for response in game output
        await expect(gameOutput).toContainText('Skills', { timeout: 10000 });
        await expect(gameOutput).toContainText('Mining');
        await expect(gameOutput).toContainText('Level');
    });
});
