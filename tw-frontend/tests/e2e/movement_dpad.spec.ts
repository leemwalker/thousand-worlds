import { test, expect } from '@playwright/test';
import { registerNewUser, waitForGameReady, sendCommand } from './fixtures/auth';

test.describe('Movement & D-Pad E2E', () => {
    test.beforeEach(async ({ page }) => {
        await registerNewUser(page);
        await waitForGameReady(page);
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
            const userMsg = page.locator('[data-testid="game-output"] .message', { hasText: `> ${dir.command}` });
            await expect(userMsg).toBeVisible({ timeout: 5000 });
        }
    });

    test('Diagonal movement works via text input', async ({ page }) => {
        // Test "ne"
        await sendCommand(page, 'ne');
        await expect(page.locator('[data-testid="game-output"] .message', { hasText: '> ne' })).toBeVisible();

        // Test "southwest"
        await sendCommand(page, 'southwest');
        await expect(page.locator('[data-testid="game-output"] .message', { hasText: '> southwest' })).toBeVisible();
    });

    test('Face command works with diagonal directions', async ({ page }) => {
        // Test "face ne"
        await sendCommand(page, 'face ne');
        await expect(page.locator('[data-testid="game-output"] .message', { hasText: '> face ne' })).toBeVisible();

        // Warning: This tests the command is sent, not necessarily the backend logic (which is unit tested).
    });
});
