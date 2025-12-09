import { test, expect } from '@playwright/test';

test.describe('Authentication Flow', () => {
    test.beforeEach(async ({ page }) => {
        await page.goto('/');
    });

    test('should display login page', async ({ page }) => {
        await expect(page.getByRole('heading', { name: 'Thousand Worlds' })).toBeVisible();
        await expect(page.getByRole('heading', { name: 'Sign In' })).toBeVisible();
    });

    test('should allow user to register', async ({ page }) => {
        // Verify we start on Sign In
        await expect(page.getByRole('heading', { name: 'Sign In' })).toBeVisible();

        // Click toggle
        const toggleBtn = page.locator('button').filter({ hasText: "Don't have an account?" });
        await toggleBtn.waitFor();

        // Wait for hydration/listeners
        await page.waitForTimeout(1000);
        await toggleBtn.click();

        // Verify switch
        await expect(page.getByRole('heading', { name: 'Create Account' })).toBeVisible();

        // Verify Username field appears
        await expect(page.getByLabel('Username')).toBeVisible();
    });

    test('should allow user to login', async ({ page }) => {
        // This test needs to be updated with actual selectors
        test.skip(true, 'Needs implementation once selectors are available');
    });

    test('should allow user to logout', async ({ page }) => {
        // This test needs to be updated with actual selectors
        test.skip(true, 'Needs implementation once selectors are available');
    });
});
