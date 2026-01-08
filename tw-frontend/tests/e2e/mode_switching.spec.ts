/**
 * Mode Switching E2E Tests
 * Tests for the dual interface mode architecture (TEXT/VISUAL switching).
 * 
 * Authentication is handled by global-setup.ts which saves storage state.
 */
import { test, expect } from '@playwright/test';

test.describe('Interface Mode Switching', () => {
    test.beforeEach(async ({ page }) => {
        // Navigate to game page (authenticated via storage state)
        await page.goto('/game');
        await page.waitForLoadState('networkidle');
    });

    test('should display mode toggle button', async ({ page }) => {
        const toggle = page.locator('[data-testid="mode-toggle"]');
        await expect(toggle).toBeVisible();
    });

    test('should have game container with mode attribute', async ({ page }) => {
        const container = page.locator('[data-testid="game-container"]');
        await expect(container).toBeVisible();

        // Check that mode attribute exists (either TEXT or VISUAL)
        const mode = await container.getAttribute('data-mode');
        expect(['TEXT', 'VISUAL']).toContain(mode);
    });

    test('should toggle between modes when clicking toggle button', async ({ page }) => {
        const toggle = page.locator('[data-testid="mode-toggle"]');
        const container = page.locator('[data-testid="game-container"]');

        // Get initial mode
        const initialMode = await container.getAttribute('data-mode');
        expect(['TEXT', 'VISUAL']).toContain(initialMode);

        // Click toggle
        await toggle.click();
        await page.waitForTimeout(500); // Allow state to update

        // Mode should have changed
        const newMode = await container.getAttribute('data-mode');
        expect(newMode).not.toBe(initialMode);
    });

    test('should persist mode preference in localStorage', async ({ page }) => {
        const toggle = page.locator('[data-testid="mode-toggle"]');

        // Toggle to ensure a mode is set
        await toggle.click();
        await page.waitForTimeout(500);

        // Check localStorage
        const storedMode = await page.evaluate(() => {
            return localStorage.getItem('tw-interface-mode');
        });

        expect(['TEXT', 'VISUAL']).toContain(storedMode);
    });
});

test.describe('Command Input Visibility', () => {
    test.beforeEach(async ({ page }) => {
        await page.goto('/game');
        await page.waitForLoadState('networkidle');
    });

    test('should have command input element', async ({ page }) => {
        const input = page.locator('#game-input');
        await expect(input).toBeVisible();
    });
});

test.describe('Layout Components', () => {
    test.beforeEach(async ({ page }) => {
        await page.goto('/game');
        await page.waitForLoadState('networkidle');
    });

    test('should render appropriate layout based on mode', async ({ page }) => {
        const container = page.locator('[data-testid="game-container"]');
        await expect(container).toBeVisible();

        const mode = await container.getAttribute('data-mode');

        if (mode === 'TEXT') {
            // MUD layout should be present
            const mudLayout = page.locator('.mud-layout');
            await expect(mudLayout).toBeVisible();
        } else {
            // Simulation layout should be present  
            const simLayout = page.locator('.simulation-layout');
            await expect(simLayout).toBeVisible();
        }
    });
});
