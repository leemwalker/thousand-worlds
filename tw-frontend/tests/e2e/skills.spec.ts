import { test, expect } from '@playwright/test';

test.describe('Player Skills UI', () => {
    test('should display skills in character sheet', async ({ page }) => {


        // Login first (using the existing auth flow or a shortcut if possible)
        // For TDD, let's just reuse the login pattern from auth.spec.ts or similar
        // Or assume we have a session.
        await page.goto('/');

        // Log in
        await page.getByLabel('Email').fill('skills_test@example.com');
        await page.getByLabel('Password').fill('password123');
        await page.getByRole('button', { name: 'Sign In' }).click();

        // Join world (if needed, or just check if we land in lobby/game)
        // Assuming we might need to create a character or select one.
        // For simplicity, let's mock the /api/auth/me and /api/game/join responses too if needed, 
        // but it's better to rely on real backend for auth and just mock the new feature.
        // However, if we utilize a fresh user, we need to create one.
        // Let's use the register flow to be safe, or just mock the game view load ?
        // Mocking the whole game view is hard.
        // Let's register a new user to ensure clean state.

        await page.getByRole('button', { name: "Don't have an account? Sign up" }).click();
        await page.getByLabel('Email').fill(`skills_test_${Date.now()}@example.com`);
        await page.getByLabel('Username').fill(`skills_${Date.now()}`);
        await page.getByLabel('Password').fill('Password123!');
        await page.getByRole('button', { name: 'Create Account' }).click();

        // Wait for redirect to game
        await page.waitForURL(/\/game/, { timeout: 30000 });

        // Wait for lobby to load (command input visible)
        const commandInput = page.getByPlaceholder('Enter command...');
        await commandInput.waitFor({ timeout: 30000 });

        // Create character via command
        await commandInput.fill('create character SkillTester Human');
        await commandInput.press('Enter');

        // Wait for game to load
        await expect(page.locator('[data-testid="game-output"]')).toContainText('You have entered the world', { timeout: 30000 });

        // Open Character Sheet (WE NEED TO IMPLEMENT THIS BUTTON/TRIGGER)
        // For now, let's assume there is a command "skills" or a UI button
        // If UI button doesn't exist, we use valid command

        await commandInput.fill('skills');
        await commandInput.press('Enter');

        // Wait for response in game output
        const gameOutput = page.locator('[data-testid="game-output"]');
        await expect(gameOutput).toContainText('Skills');
        await expect(gameOutput).toContainText('Mining');
        // Level might be 0 or 1 depending on defaults, checking text contains "Level"
        await expect(gameOutput).toContainText('Level');


    });
});
