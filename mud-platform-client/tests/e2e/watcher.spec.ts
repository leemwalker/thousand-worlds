import { test, expect } from '@playwright/test';

test.describe('Watcher Mode', () => {
    test.beforeEach(async ({ page }) => {
        // Unique user for each test to avoid conflicts
        const timestamp = Date.now();
        const email = `watcher_${timestamp}@example.com`;
        const username = `watcher_${timestamp}`;

        await page.goto('/');

        // Register new user
        const toggleButton = page.locator('button:has-text("Don\'t have an account? Sign up")');
        if (await toggleButton.isVisible()) {
            await toggleButton.click();
        }

        await page.locator('input[type="email"]').first().fill(email);
        await page.getByLabel('Username').fill(username);
        await page.locator('input[type="password"]').first().fill('WatcherPass123!');
        // Monitor browser console
        page.on('console', msg => console.log(`[BROWSER] ${msg.text()}`));

        await page.locator('button[type="submit"]', { hasText: 'Create Account' }).click();

        // Wait for redirect to game
        await page.waitForURL(/\/game/, { timeout: 15000 });

        // Wait to be in game
        await page.waitForSelector('input[placeholder*="command"]', { timeout: 30000 });
    });

    test('should allow entering world as watcher', async ({ page }) => {
        const commandInput = page.locator('input[placeholder*="command"]').first();

        // 1. List worlds (to ensure connection and get a valid world if possible, 
        // but for now we might rely on "enter <name>" if we know a name, 
        // or just create a world quickly first? 
        // Creating a world takes time (interview).
        // Better: Try to enter a known default world or just fail gracefully if none.
        // Or assume the test environment might have a seed world.
        // Actually, without a known world, we can't test "enter".
        // SO: We must create a world first, OR rely on a mocked world list in a different test type.
        // BUT: This is E2E against real backend.
        // Solution: Create a simple world first (or assume one from previous tests if persistence? No, isolation is better).
        // Let's create a *quick* world?
        // "tell statue create world" -> answer quickly.

        // FAST PATH: If the backend supports "enter lobby" or similar, but we are already in lobby.
        // We need another world.
        // Let's do the interview. It's the only way to be sure.

        await commandInput.fill('tell statue create world');
        await commandInput.press('Enter');
        await page.waitForTimeout(5000); // Wait for statue

        // Answer questions rapidly
        const answers = ['Cyber', 'AI', 'City', 'Tech', 'War', 'WatcherWorld'];
        for (const ans of answers) {
            await commandInput.fill(`reply ${ans}`);
            await commandInput.press('Enter');
            await page.waitForTimeout(2000); // Wait for processing
        }

        // Wait for creation "forged"
        await expect(page.locator('[data-testid="game-output"]')).toContainText('forged', { timeout: 60000 });

        // 2. Enter the world
        await commandInput.fill('enter WatcherWorld');
        await commandInput.press('Enter');

        // 3. Expect Entry Options (Watcher vs Character)
        const watcherButton = page.locator('button', { hasText: 'Watcher' });
        await watcherButton.waitFor({ state: 'visible', timeout: 10000 });

        // 4. Click Watcher button
        await watcherButton.click();

        // 5. Verify Entry
        // Should trigger "You have entered the world" or similar in game output
        const gameOutput = page.locator('[data-testid="game-output"]');
        await expect(gameOutput).toContainText('You have entered the world', { timeout: 10000 });

        // Watcher specific check: Should NOT have Inventory/Status commands or they might work but return empty?
        // Actually watcher can look.
        await commandInput.fill('look');
        await commandInput.press('Enter');
        await expect(gameOutput).toContainText('area_description'); // or actual description text if we knew it
    });
});
