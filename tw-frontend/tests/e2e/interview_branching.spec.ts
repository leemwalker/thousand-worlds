import { test, expect } from '@playwright/test';
import { randomUUID } from 'crypto';

test.describe('Dynamic Interview Branching', () => {
    test.setTimeout(300000); // 5 minutes because LLM is slow

    // Ensure clean environment for each test
    test.beforeEach(async ({ page, context }) => {
        await context.clearCookies();
        await page.goto('/'); // Need to be on domain to clear storage
        await page.evaluate(() => {
            localStorage.clear();
            sessionStorage.clear();
        });
    });

    // Shared setup for registration
    async function registerAndLogin(page, browserName) {
        // Prevent iOS install prompt from appearing
        await page.addInitScript(() => {
            localStorage.setItem('iosInstallPromptDismissed', String(Date.now()));
        });

        const uuid = randomUUID();
        const email = `branch_${uuid}@example.com`;
        const username = `u_${uuid.substring(0, 8)}`;

        await page.goto('/');

        // Switch to Register mode if needed
        const toggleButton = page.locator('button').filter({ hasText: "Don't have an account?" });

        // Wait for it to be visible and stable
        try {
            await toggleButton.waitFor({ state: 'visible', timeout: 5000 });
            await page.waitForTimeout(1000); // Wait for hydration

            if (await toggleButton.isVisible()) {
                console.log("Found toggle button, clicking...");
                await toggleButton.click({ force: true });
                await page.getByRole('heading', { name: 'Create Account' }).waitFor({ timeout: 10000 });
            }
        } catch (e) {
            // If toggle not found or timeout, check if already on Create Account
            if (!await page.getByRole('heading', { name: 'Create Account' }).isVisible()) {
                console.log("Toggle button issue, trying to proceed anyway...");
            }
        }

        await page.locator('input[type="email"]').first().fill(email);

        // Handle username label or placeholder
        const usernameInput = page.getByLabel('Username');
        // Wait for it
        await usernameInput.waitFor({ state: 'visible', timeout: 5000 });
        await usernameInput.fill(username);

        await page.locator('input[type="password"]').first().fill('Password123!');
        await page.locator('button[type="submit"]', { hasText: 'Create Account' }).click();

        // Should redirect to game
        await page.waitForURL(/\/game/, { timeout: 30000 });

        return username;
    }

    // Helper to handle portal navigation
    async function enterWorld(page, worldName) {
        const input = page.getByPlaceholder('Enter command...');

        await input.fill(`enter ${worldName}`);
        await input.press('Enter');

        // Check if Modal is visible - we expect it to be
        await expect(page.getByRole('heading', { name: 'Choose Your Path' })).toBeVisible({ timeout: 10000 });
        await page.getByRole('button', { name: /Watcher/ }).first().click();
    }

    test('Path A: Immediate naming at Branch point', async ({ page, browserName }) => {
        test.slow();

        await registerAndLogin(page, browserName);
        const worldName = `ImmediateWorld-${Date.now()}`;

        // Wait for input to confirm game usage
        const input = page.getByPlaceholder('Enter command...');
        await input.waitFor({ state: 'visible', timeout: 30000 });

        // Start Interview by talking to Statue
        // Send trigger
        await input.fill('tell statue start');
        await input.press('Enter');

        // Wait for interview start
        await expect(page.locator('[data-testid="game-output"]')).toContainText('voice resonates', { timeout: 60000 });

        const answers = [
            "High Fantasy", // Core Concept
            "Elves and Humans", // Sentient Species
            "Forested", // Environment
            "High Magic", // Magic & Tech
            "Dark Lord", // Conflict
            "Ancient", // Geological Age
        ];

        for (const ans of answers) {
            // Wait for 'reply' instruction or generic update
            // Wait for 'reply' instruction or just generic message update
            await expect(page.locator('[data-testid="game-output"]')).toBeVisible();

            // Wait a bit for the question to fully stream in
            await page.waitForTimeout(15000);

            // Type answer
            await input.fill(`reply ${ans}`);
            await input.press('Enter');
        }

        // Now at Q7 (Branch)
        // Expect the branch prompt: "decision point... reply with 'the name is <world name>' to finish or 'continue'"
        // We wait for specific text "decision point"
        await expect(page.locator('[data-testid="game-output"]')).toContainText('decision point', { timeout: 40000 });

        // Name immediately
        await input.fill(`reply the name is ${worldName}`);
        await input.press('Enter');

        // Expect completion and world creation
        await expect(page.locator('[data-testid="game-output"]')).toContainText('world has been created', { timeout: 60000 });

        // Enter world with navigation
        await enterWorld(page, worldName);

        // Verify watcher entry message from system

        // Verify entry
        await expect(page.locator('[data-testid="game-output"]')).toContainText('enter the world as a watcher', { timeout: 30000 });
    });

    test('Path B: Continue then Global Interrupt', async ({ page, browserName }) => {
        test.slow();

        await registerAndLogin(page, browserName);
        const worldName = `InterruptWorld-${Date.now()}`;

        // Wait for input
        const input = page.getByPlaceholder('Enter command...');
        await input.waitFor({ state: 'visible', timeout: 30000 });

        // Start Interview
        await input.fill('tell statue start');
        await input.press('Enter');
        await expect(page.locator('[data-testid="game-output"]')).toContainText('voice resonates', { timeout: 60000 });

        const answers = [
            "Sci-Fi",
            "Robots",
            "Metal",
            "High Tech",
            "Virus",
            "New",
        ];

        for (const ans of answers) {
            await expect(page.locator('[data-testid="game-output"]')).toContainText('Statue', { timeout: 60000 });
            // LLM takes ~8s, allow plenty of time for response to arrive before replying again
            await page.waitForTimeout(15000);
            await input.fill(`reply ${ans}`);
            await input.press('Enter');
        }

        // At Branch
        await expect(page.locator('[data-testid="game-output"]')).toContainText('decision point', { timeout: 40000 });

        // Say continue
        await input.fill('reply continue');
        await input.press('Enter');

        // Should get Q8 (Factions)
        // Wait for response: "What factions exist?" or similar.
        await expect(page.locator('[data-testid="game-output"]')).toContainText('Statue', { timeout: 60000 });

        // Let's interrupt immediately at Q8
        await input.fill(`reply the name is ${worldName}`);
        await input.press('Enter');

        // Expect completion
        await expect(page.locator('[data-testid="game-output"]')).toContainText('world has been created', { timeout: 60000 });

        // Enter world with navigation
        await enterWorld(page, worldName);

        // Verify watcher entry message from system
        await expect(page.locator('[data-testid="game-output"]')).toContainText('enter the world as a watcher', { timeout: 30000 });
    });
});
