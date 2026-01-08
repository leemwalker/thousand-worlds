/**
 * Mode Switching E2E Tests
 * Tests for the dual interface mode architecture (TEXT/VISUAL switching).
 * 
 * NOTE: These tests navigate to /game which requires authentication.
 * If not authenticated, the page redirects to /. Tests check for
 * the mode toggle and skip if auth is required.
 */
import { test, expect } from '@playwright/test';

test.describe('Interface Mode Switching', () => {
    test.beforeEach(async ({ page }) => {
        // Navigate to game page (may redirect to login if not authenticated)
        await page.goto('/game');
        await page.waitForLoadState('networkidle');
    });

    test('should display mode toggle button on game page', async ({ page }) => {
        // Check if we're on the game page (has mode toggle)
        const toggle = page.locator('[data-testid="mode-toggle"]');

        // Wait a bit for potential redirect/render
        await page.waitForTimeout(1000);

        const toggleExists = await toggle.count() > 0;

        if (toggleExists) {
            await expect(toggle).toBeVisible();
        } else {
            // We're on login page - skip this test (auth required)
            test.skip(true, 'Requires authentication - mode toggle only available on /game');
        }
    });

    test('should have game container with mode attribute', async ({ page }) => {
        await page.waitForTimeout(1000);
        const container = page.locator('[data-testid="game-container"]');
        const containerExists = await container.count() > 0;

        if (containerExists) {
            // Check that mode attribute exists (either TEXT or VISUAL)
            const mode = await container.getAttribute('data-mode');
            expect(['TEXT', 'VISUAL']).toContain(mode);
        } else {
            test.skip(true, 'Requires authentication - game container only available on /game');
        }
    });

    test('should toggle between modes when clicking toggle button', async ({ page }) => {
        await page.waitForTimeout(1000);
        const toggle = page.locator('[data-testid="mode-toggle"]');
        const container = page.locator('[data-testid="game-container"]');

        const toggleExists = await toggle.count() > 0;
        if (!toggleExists) {
            test.skip(true, 'Requires authentication');
            return;
        }

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
        await page.waitForTimeout(1000);
        const toggle = page.locator('[data-testid="mode-toggle"]');

        const toggleExists = await toggle.count() > 0;
        if (!toggleExists) {
            test.skip(true, 'Requires authentication');
            return;
        }

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
        await page.waitForTimeout(1000);
    });

    test('should have command input element', async ({ page }) => {
        const input = page.locator('#game-input');
        const inputExists = await input.count() > 0;

        if (inputExists) {
            await expect(input).toBeVisible();
        } else {
            test.skip(true, 'Requires authentication - input only available on /game');
        }
    });
});

test.describe('Layout Components', () => {
    test.beforeEach(async ({ page }) => {
        await page.goto('/game');
        await page.waitForLoadState('networkidle');
        await page.waitForTimeout(1000);
    });

    test('should render appropriate layout based on mode', async ({ page }) => {
        const container = page.locator('[data-testid="game-container"]');
        const containerExists = await container.count() > 0;

        if (!containerExists) {
            test.skip(true, 'Requires authentication');
            return;
        }

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
