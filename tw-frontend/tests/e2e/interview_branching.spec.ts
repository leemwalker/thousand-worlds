import { test, expect } from '@playwright/test';
import { registerNewUser, waitForGameReady, sendCommand, ensureLoggedOut, suppressIOSPrompt } from './fixtures/auth';

test.describe('Dynamic Interview Branching', () => {
    test.setTimeout(300000); // 5 minutes because LLM is slow

    // Ensure clean environment for each test
    test.beforeEach(async ({ page, context }) => {
        await context.clearCookies();
        await page.goto('/');
        await page.evaluate(() => {
            localStorage.clear();
            sessionStorage.clear();
        });
    });

    // Helper to handle portal navigation
    async function enterWorld(page: any, worldName: string) {
        await sendCommand(page, `enter ${worldName}`);

        // Check if Modal is visible - we expect it to be
        await expect(page.getByRole('heading', { name: 'Choose Your Path' })).toBeVisible({ timeout: 10000 });
        await page.getByRole('button', { name: /Watcher/ }).first().click();
    }

    test('Path A: Immediate naming at Branch point', async ({ page }) => {
        test.slow();

        await registerNewUser(page);
        await waitForGameReady(page);

        const worldName = `ImmediateWorld-${Date.now()}`;
        const gameOutput = page.locator('[data-testid="game-output"]');

        // Start Interview by talking to Statue
        await sendCommand(page, 'tell statue start');

        // Wait for interview start
        await expect(gameOutput).toContainText('voice resonates', { timeout: 60000 });

        const answers = [
            "High Fantasy", // Core Concept
            "Elves and Humans", // Sentient Species
            "Forested", // Environment
            "High Magic", // Magic & Tech
            "Dark Lord", // Conflict
            "Ancient", // Geological Age
        ];

        for (const ans of answers) {
            // Wait for the question to fully stream in
            await page.waitForTimeout(15000);
            await sendCommand(page, `reply ${ans}`);
        }

        // Now at Q7 (Branch)
        await expect(gameOutput).toContainText('decision point', { timeout: 40000 });

        // Name immediately
        await sendCommand(page, `reply the name is ${worldName}`);

        // Expect completion and world creation
        await expect(gameOutput).toContainText('world has been created', { timeout: 60000 });

        // Enter world with navigation
        await enterWorld(page, worldName);

        // Verify entry
        await expect(gameOutput).toContainText('enter the world as a watcher', { timeout: 30000 });
    });

    test('Path B: Continue then Global Interrupt', async ({ page }) => {
        test.slow();

        await registerNewUser(page);
        await waitForGameReady(page);

        const worldName = `InterruptWorld-${Date.now()}`;
        const gameOutput = page.locator('[data-testid="game-output"]');

        // Start Interview
        await sendCommand(page, 'tell statue start');
        await expect(gameOutput).toContainText('voice resonates', { timeout: 60000 });

        const answers = [
            "Sci-Fi",
            "Robots",
            "Metal",
            "High Tech",
            "Virus",
            "New",
        ];

        for (const ans of answers) {
            await expect(gameOutput).toContainText('Statue', { timeout: 60000 });
            await page.waitForTimeout(15000);
            await sendCommand(page, `reply ${ans}`);
        }

        // At Branch
        await expect(gameOutput).toContainText('decision point', { timeout: 40000 });

        // Say continue
        await sendCommand(page, 'reply continue');

        // Should get Q8 (Factions)
        await expect(gameOutput).toContainText('Statue', { timeout: 60000 });

        // Interrupt immediately at Q8 with name
        await sendCommand(page, `reply the name is ${worldName}`);

        // Expect completion
        await expect(gameOutput).toContainText('world has been created', { timeout: 60000 });

        // Enter world with navigation
        await enterWorld(page, worldName);

        // Verify watcher entry message from system
        await expect(gameOutput).toContainText('enter the world as a watcher', { timeout: 30000 });
    });
});
