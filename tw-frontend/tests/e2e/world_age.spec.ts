import { test, expect } from '@playwright/test';
import { registerNewUser, waitForGameReady, sendCommand, waitForMessage } from './fixtures/auth';

/**
 * World Generation Age Parameter Test
 * Verifies the "Geological Age" interview question and its effects.
 */

test.describe('World Generation Age Parameter', () => {
    test.setTimeout(300000); // 5 minutes for LLM interviews

    // Testing "Old" World Age
    test('Verify World Creation with Old Geological Age', async ({ page }) => {
        await registerNewUser(page);
        await waitForGameReady(page);

        const gameOutput = page.locator('[data-testid="game-output"]');
        const timestamp = Date.now();

        // Start Interview
        await sendCommand(page, 'tell statue I want to create a world');
        await expect(gameOutput).toContainText('voice resonates', { timeout: 120000 });

        // Answer Questions 1-5
        const answers = [
            'A fantasy realm', // Core Concept
            'Elves and Dwarves', // Sentient Species
            'Forests and mountains', // Environment
            'High Magic', // Magic & Tech
            'Ancient war', // Conflict
        ];

        for (const answer of answers) {
            await page.waitForTimeout(15000); // Wait for LLM
            await sendCommand(page, `reply ${answer}`);
        }

        // Q6: Geological Age - Select "Old"
        await page.waitForTimeout(15000);
        await sendCommand(page, 'reply Old with smoothed mountains');

        // Answer Name
        const worldName = `OldWorld-${timestamp}`;
        await page.waitForTimeout(15000);
        await sendCommand(page, `reply ${worldName}`);

        // Verify Summary reflects "Old"
        await expect(gameOutput).toContainText('Here is the vision for your world', { timeout: 60000 });
        await expect(gameOutput).toContainText('Geological Age');
        await expect(gameOutput).toContainText('Old with smoothed mountains');

        // Confirm
        await sendCommand(page, 'reply yes');

        await expect(gameOutput).toContainText('Your world is being forged', { timeout: 30000 });
    });

    // Testing "Young" World Age
    test('Verify World Creation with Young Geological Age', async ({ page }) => {
        await registerNewUser(page);
        await waitForGameReady(page);

        const gameOutput = page.locator('[data-testid="game-output"]');
        const timestamp = Date.now();

        // Start Interview
        await sendCommand(page, 'tell statue I want to create a world');
        await expect(gameOutput).toContainText('voice resonates', { timeout: 120000 });

        // Answer Questions 1-5
        const answers = [
            'A primal land', // Core Concept
            'Dinosaurs', // Sentient Species
            'Volcanoes and swamps', // Environment
            'No magic', // Magic & Tech
            'Survival', // Conflict
        ];

        for (const answer of answers) {
            await page.waitForTimeout(15000); // Wait for LLM
            await sendCommand(page, `reply ${answer}`);
        }

        // Q6: Geological Age - Select "Young"
        await page.waitForTimeout(15000);
        await sendCommand(page, 'reply Young, sharp peaks everywhere');

        // Answer Name
        const worldName = `YoungWorld-${timestamp}`;
        await page.waitForTimeout(15000);
        await sendCommand(page, `reply ${worldName}`);

        // Verify Summary reflects "Young"
        await expect(gameOutput).toContainText('Geological Age', { timeout: 60000 });
        await expect(gameOutput).toContainText('Young, sharp peaks everywhere');
    });
});
