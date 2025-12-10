import { test, expect } from '@playwright/test';
import {
    generateCredentials,
    registerNewUser,
    loginUser,
    ensureLoggedOut,
    suppressIOSPrompt
} from './fixtures/auth';

test.describe('Authentication Flow', () => {
    test.beforeEach(async ({ page }) => {
        await ensureLoggedOut(page);
        await suppressIOSPrompt(page);
    });

    test('should display login page', async ({ page }) => {
        await page.goto('/');
        await expect(page.getByRole('heading', { name: 'Thousand Worlds' })).toBeVisible();
        await expect(page.getByRole('heading', { name: 'Sign In' })).toBeVisible();
    });

    test('should allow user to toggle to register mode', async ({ page }) => {
        await page.goto('/');

        // Wait for page hydration
        await page.waitForLoadState('networkidle');
        await page.waitForTimeout(500);

        // Click toggle to register mode
        const toggleBtn = page.locator('button').filter({ hasText: "Don't have an account?" });
        await toggleBtn.waitFor({ state: 'visible' });
        await toggleBtn.click();

        await expect(page.getByRole('heading', { name: 'Create Account' })).toBeVisible();
        await expect(page.getByLabel('Username')).toBeVisible();
    });

    test('should successfully register new user', async ({ page }) => {
        const creds = await registerNewUser(page);

        // Verify we reached the game
        await expect(page.locator('[data-testid="game-output"]')).toBeVisible();

        // Verify no error messages
        await expect(page.locator('.bg-red-900')).not.toBeVisible();
    });

    test('should login with valid credentials', async ({ page }) => {
        // First register a user
        const creds = await registerNewUser(page);

        // Logout
        await ensureLoggedOut(page);

        // Login with same credentials  
        await loginUser(page, creds.email, creds.password);

        // Verify we reached the game
        await expect(page.locator('[data-testid="game-output"]')).toBeVisible();
    });

    test('should show error for invalid credentials', async ({ page }) => {
        await page.goto('/');
        await page.waitForLoadState('networkidle');

        const emailInput = page.locator('input[type="email"]').first();
        await emailInput.waitFor({ state: 'visible' });
        await emailInput.fill('nonexistent@example.com');

        const passwordInput = page.locator('input[type="password"]').first();
        await passwordInput.fill('WrongPassword123');

        await page.locator('button[type="submit"]').first().click();

        // Should see error message - the error div has bg-red-900/50 class
        const errorDiv = page.locator('div.border-red-500');
        await expect(errorDiv).toBeVisible({ timeout: 10000 });
    });

    test('should persist session on page reload', async ({ page }) => {
        await registerNewUser(page);

        // Reload page
        await page.reload();

        // Should still be in game (not redirected to login)
        // Give time for redirect check
        await page.waitForTimeout(2000);
        await expect(page.locator('[data-testid="game-output"]')).toBeVisible({ timeout: 15000 });
    });

    test('should redirect authenticated user from login to game', async ({ page }) => {
        // Register to establish auth
        await registerNewUser(page);

        // Try to navigate back to login page
        await page.goto('/');

        // Should redirect back to game since already authenticated
        await page.waitForURL(/\/game/, { timeout: 15000 });
        await expect(page.locator('[data-testid="game-output"]')).toBeVisible();
    });
});
