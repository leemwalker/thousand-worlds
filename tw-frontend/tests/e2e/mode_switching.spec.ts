/**
 * Mode Switching E2E Tests
 * Tests for the dual interface mode architecture (TEXT/VISUAL switching).
 */
import { test, expect } from '@playwright/test';

test.describe('Interface Mode Switching', () => {
    test.beforeEach(async ({ page }) => {
        // Wait for the app to be ready
        await page.goto('/');
        await page.waitForLoadState('networkidle');
    });

    test('should display mode toggle button', async ({ page }) => {
        const toggle = page.locator('[data-testid="mode-toggle"]');
        await expect(toggle).toBeVisible();
    });

    test('should default to VISUAL mode on desktop viewport', async ({ page }) => {
        // Set desktop viewport
        await page.setViewportSize({ width: 1920, height: 1080 });
        await page.goto('/');
        await page.waitForLoadState('networkidle');

        const container = page.locator('[data-testid="game-container"]');
        await expect(container).toHaveAttribute('data-mode', 'VISUAL');
    });

    test('should default to TEXT mode on mobile viewport', async ({ page }) => {
        // Set mobile viewport
        await page.setViewportSize({ width: 375, height: 667 });
        await page.goto('/');
        await page.waitForLoadState('networkidle');

        const container = page.locator('[data-testid="game-container"]');
        await expect(container).toHaveAttribute('data-mode', 'TEXT');
    });

    test('should toggle from VISUAL to TEXT mode when clicking toggle', async ({ page }) => {
        // Start on desktop (VISUAL mode)
        await page.setViewportSize({ width: 1920, height: 1080 });
        await page.goto('/');
        await page.waitForLoadState('networkidle');

        const container = page.locator('[data-testid="game-container"]');
        await expect(container).toHaveAttribute('data-mode', 'VISUAL');

        // Click toggle
        const toggle = page.locator('[data-testid="mode-toggle"]');
        await toggle.click();

        // Should now be TEXT mode
        await expect(container).toHaveAttribute('data-mode', 'TEXT');
    });

    test('should toggle from TEXT to VISUAL mode when clicking toggle', async ({ page }) => {
        // Start on mobile (TEXT mode)
        await page.setViewportSize({ width: 375, height: 667 });
        await page.goto('/');
        await page.waitForLoadState('networkidle');

        const container = page.locator('[data-testid="game-container"]');
        await expect(container).toHaveAttribute('data-mode', 'TEXT');

        // Click toggle
        const toggle = page.locator('[data-testid="mode-toggle"]');
        await toggle.click();

        // Should now be VISUAL mode
        await expect(container).toHaveAttribute('data-mode', 'VISUAL');
    });

    test('should persist mode preference in localStorage', async ({ page }) => {
        await page.setViewportSize({ width: 1920, height: 1080 });
        await page.goto('/');
        await page.waitForLoadState('networkidle');

        // Toggle to TEXT mode
        const toggle = page.locator('[data-testid="mode-toggle"]');
        await toggle.click();

        // Verify localStorage was updated
        const storedMode = await page.evaluate(() => {
            return localStorage.getItem('tw-interface-mode');
        });
        expect(storedMode).toBe('TEXT');

        // Reload and verify persistence
        await page.reload();
        await page.waitForLoadState('networkidle');

        const container = page.locator('[data-testid="game-container"]');
        await expect(container).toHaveAttribute('data-mode', 'TEXT');
    });
});

test.describe('Command Input in Both Modes', () => {
    test('should have command input visible in TEXT mode', async ({ page }) => {
        await page.setViewportSize({ width: 375, height: 667 });
        await page.goto('/');
        await page.waitForLoadState('networkidle');

        const input = page.locator('#game-input');
        await expect(input).toBeVisible();
    });

    test('should have command input visible in VISUAL mode', async ({ page }) => {
        await page.setViewportSize({ width: 1920, height: 1080 });
        await page.goto('/');
        await page.waitForLoadState('networkidle');

        // In VISUAL mode, command input should be in the overlay at bottom
        const input = page.locator('#game-input');
        await expect(input).toBeVisible();
    });

    test('should process movement command in both modes', async ({ page }) => {
        // Test in TEXT mode
        await page.setViewportSize({ width: 375, height: 667 });
        await page.goto('/');
        await page.waitForLoadState('networkidle');

        const input = page.locator('#game-input');
        await input.fill('go north');
        await input.press('Enter');

        // Input should be cleared after submit
        await expect(input).toHaveValue('');

        // Toggle to VISUAL mode and test again
        const toggle = page.locator('[data-testid="mode-toggle"]');
        await toggle.click();

        await input.fill('go south');
        await input.press('Enter');

        await expect(input).toHaveValue('');
    });
});

test.describe('Layout Differences Between Modes', () => {
    test('TEXT mode should have dedicated text display area', async ({ page }) => {
        await page.setViewportSize({ width: 1920, height: 1080 });
        await page.goto('/');
        await page.waitForLoadState('networkidle');

        // Toggle to TEXT mode
        const toggle = page.locator('[data-testid="mode-toggle"]');
        await toggle.click();

        // Should see MUD layout elements
        const mudLayout = page.locator('.mud-layout');
        await expect(mudLayout).toBeVisible();
    });

    test('VISUAL mode should have 3D canvas area', async ({ page }) => {
        await page.setViewportSize({ width: 1920, height: 1080 });
        await page.goto('/');
        await page.waitForLoadState('networkidle');

        // Should see simulation layout elements
        const simLayout = page.locator('.simulation-layout');
        await expect(simLayout).toBeVisible();
    });
});
