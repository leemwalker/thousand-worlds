import { test, expect } from '@playwright/test';
import { registerNewUser, waitForGameReady, sendCommand, waitForMessage } from './fixtures/auth';

test.describe('Ecosystem System', () => {
    test('ecosystem status shows entity count', async ({ page }) => {
        await registerNewUser(page);
        await waitForGameReady(page);

        // Status should show entity count (may have entities from other tests)
        await sendCommand(page, 'ecosystem status');
        await waitForMessage(page, 'Ecosystem Status', 10000);
    });

    test('spawn entity and view with look', async ({ page }) => {
        await registerNewUser(page);
        await waitForGameReady(page);

        // Spawn a rabbit
        await sendCommand(page, 'ecosystem spawn rabbit');
        await waitForMessage(page, 'Spawned rabbit at your location', 10000);

        // Check status shows entities (at least 1)
        await sendCommand(page, 'ecosystem status');
        await waitForMessage(page, 'entities', 10000);

        // Look should mention the rabbit
        await sendCommand(page, 'look');
        await waitForMessage(page, 'rabbit is here', 10000);

        // Look at the rabbit directly
        await sendCommand(page, 'look rabbit');
        await waitForMessage(page, 'You see a rabbit', 10000);
    });

    test('breed two entities and check lineage', async ({ page }) => {
        await registerNewUser(page);
        await waitForGameReady(page);

        // Spawn two wolves
        await sendCommand(page, 'ecosystem spawn wolf');
        await waitForMessage(page, 'Spawned wolf', 10000);

        await sendCommand(page, 'ecosystem spawn wolf');
        // Wait for second spawn confirmation
        await page.waitForTimeout(1000);

        // Get status to find IDs
        await sendCommand(page, 'ecosystem status');
        await waitForMessage(page, 'entities', 10000);

        // Extract IDs from the status output
        const output = await page.locator('[data-testid="game-output"]').textContent();
        const idMatches = output?.match(/\[wolf\] ([a-f0-9]{8})/g);

        if (!idMatches || idMatches.length < 2) {
            throw new Error('Could not find two wolf IDs in ecosystem status');
        }

        // Extract just the ID parts (8 hex chars after "[wolf] ")
        const id1 = idMatches[0].replace('[wolf] ', '');
        const id2 = idMatches[1].replace('[wolf] ', '');

        // Breed them
        await sendCommand(page, `ecosystem breed ${id1} ${id2}`);
        await waitForMessage(page, 'Bred wolf!', 10000);
        await waitForMessage(page, 'Generation 2', 10000);

        // Get the new status
        await sendCommand(page, 'ecosystem status');
        await waitForMessage(page, 'entities', 10000);

        // Extract new ID (find any wolf)
        const newOutput = await page.locator('[data-testid="game-output"]').textContent();
        const newIdMatches = newOutput?.match(/\[wolf\] ([a-f0-9]{8})/g);

        if (!newIdMatches || newIdMatches.length < 1) {
            throw new Error('Could not find wolf ID for lineage');
        }

        // Check lineage of any wolf
        const anyWolfId = newIdMatches[0].replace('[wolf] ', '');
        await sendCommand(page, `ecosystem lineage ${anyWolfId}`);
        await waitForMessage(page, 'Lineage for', 10000);
        await waitForMessage(page, 'Generation:', 10000);
    });

    test('ecosystem log shows AI decisions', async ({ page }) => {
        await registerNewUser(page);
        await waitForGameReady(page);

        // Spawn a rabbit
        await sendCommand(page, 'ecosystem spawn rabbit');
        await waitForMessage(page, 'Spawned rabbit', 10000);

        // Get the ID
        await sendCommand(page, 'ecosystem status');
        await waitForMessage(page, 'entities', 10000);

        const output = await page.locator('[data-testid="game-output"]').textContent();
        const idMatch = output?.match(/\[rabbit\] ([a-f0-9]{8})/);

        if (!idMatch) {
            throw new Error('Could not find rabbit ID');
        }

        const rabbitId = idMatch[1];

        // Wait a moment for AI to act (it should wander at least once)
        await page.waitForTimeout(1500);

        // Check logs
        await sendCommand(page, `ecosystem log ${rabbitId}`);
        await waitForMessage(page, 'Decision Logs for', 10000);
    });
});
