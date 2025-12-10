import { test, expect } from '@playwright/test';

/**
 * World Generation Age Parameter Test
 * Verifies the "Geological Age" interview question and its effects.
 */

// Helper to wait for specific message type
async function waitForGameMessage(page: any, timeout = 30000): Promise<void> {
    const messages = page.locator('[data-testid="game-output"] > div');
    const count = await messages.count();
    await expect(async () => {
        const newCount = await messages.count();
        expect(newCount).toBeGreaterThan(count);
    }).toPass({ timeout });
}

test.describe('World Generation Age Parameter', () => {

    // Testing "Old" World Age
    test('Verify World Creation with Old Geological Age', async ({ page }) => {
        const timestamp = Date.now();
        const email = `age_old_${timestamp}@example.com`;
        const username = `age_old_${timestamp}`;

        // 1. Create Account
        await page.goto('/');
        await page.locator('button').filter({ hasText: "Don't have an account?" }).click();
        await page.locator('input[type="email"]').first().fill(email);
        const usernameInput = page.getByLabel('Username');
        await usernameInput.waitFor({ state: 'visible' });
        await usernameInput.fill(username);
        await page.locator('input[type="password"]').first().fill('Password123!');
        await page.locator('button[type="submit"]', { hasText: 'Create Account' }).click();

        // 2. Start Interview
        await page.waitForSelector('input[placeholder*="command"]', { timeout: 20000 });
        const commandInput = page.locator('input[placeholder*="command"]').first();
        await commandInput.fill('tell statue I want to create a world');
        await commandInput.press('Enter');
        await expect(page.locator('[data-testid="game-output"]')).toContainText('voice resonates', { timeout: 120000 });

        // 3. Answer Questions 1-5
        const answers = [
            'A fantasy realm', // Core Concept
            'Elves and Dwarves', // Sentient Species
            'Forests and mountains', // Environment
            'High Magic', // Magic & Tech
            'Ancient war', // Conflict
        ];

        for (const answer of answers) {
            await commandInput.fill(`reply ${answer}`);
            await commandInput.press('Enter');
            await waitForGameMessage(page, 120000);
        }

        // 4. Q6: Geological Age - Select "Old"
        await commandInput.fill('reply Old with smoothed mountains');
        await commandInput.press('Enter');
        await waitForGameMessage(page, 120000);

        // 5. Finalize with Review
        // Should show summary now or ask for Name (World Name is last)

        // Wait, list is: Core, Sentient, Env, Magic, Conflict, Age, Name
        // So Age is Q6 (index 5). Name is Q7 (index 6).
        // After answering Q6, it asks Q7 (Name).

        // Answer Name
        const worldName = `OldWorld-${timestamp}`;
        await commandInput.fill(`reply ${worldName}`);
        await commandInput.press('Enter');

        // Now it shows Summary (Review)
        await expect(page.locator('[data-testid="game-output"]')).toContainText('Here is the vision for your world', { timeout: 60000 });

        // Verify Summary reflects "Old"
        await expect(page.locator('[data-testid="game-output"]')).toContainText('Geological Age');
        // Since we check the summary text, we look for our answer
        await expect(page.locator('[data-testid="game-output"]')).toContainText('Old with smoothed mountains');

        // Confirm
        await commandInput.fill('reply yes');
        await commandInput.press('Enter');

        await expect(page.locator('[data-testid="game-output"]')).toContainText('Your world is being forged', { timeout: 30000 });
    });

    // Testing "Young" World Age
    test('Verify World Creation with Young Geological Age', async ({ page }) => {
        const timestamp = Date.now();
        const email = `age_young_${timestamp}@example.com`;
        const username = `age_young_${timestamp}`;

        // 1. Create Account
        await page.goto('/');
        await page.locator('button').filter({ hasText: "Don't have an account?" }).click();
        await page.locator('input[type="email"]').first().fill(email);
        const usernameInput = page.getByLabel('Username');
        await usernameInput.waitFor({ state: 'visible' });
        await usernameInput.fill(username);
        await page.locator('input[type="password"]').first().fill('Password123!');
        await page.locator('button[type="submit"]', { hasText: 'Create Account' }).click();

        // 2. Start Interview
        await page.waitForSelector('input[placeholder*="command"]', { timeout: 20000 });
        const commandInput = page.locator('input[placeholder*="command"]').first();
        await commandInput.fill('tell statue I want to create a world');
        await commandInput.press('Enter');
        await expect(page.locator('[data-testid="game-output"]')).toContainText('voice resonates', { timeout: 120000 });

        // 3. Answer Questions 1-5
        const answers = [
            'A primal land', // Core Concept
            'Dinosaurs', // Sentient Species
            'Volcanoes and swamps', // Environment
            'No magic', // Magic & Tech
            'Survival', // Conflict
        ];

        for (const answer of answers) {
            await commandInput.fill(`reply ${answer}`);
            await commandInput.press('Enter');
            await waitForGameMessage(page, 120000);
        }

        // 4. Q6: Geological Age - Select "Young"
        await commandInput.fill('reply Young, sharp peaks everywhere');
        await commandInput.press('Enter');
        await waitForGameMessage(page, 120000);

        // Answer Name
        const worldName = `YoungWorld-${timestamp}`;
        await commandInput.fill(`reply ${worldName}`);
        await commandInput.press('Enter');

        // Verify Summary reflects "Young"
        await expect(page.locator('[data-testid="game-output"]')).toContainText('Geological Age');
        await expect(page.locator('[data-testid="game-output"]')).toContainText('Young, sharp peaks everywhere');
    });

});
