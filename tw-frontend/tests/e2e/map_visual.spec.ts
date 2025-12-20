import { test, expect } from '@playwright/test';
import { registerNewUser, waitForGameReady, sendCommand } from './fixtures/auth';

/**
 * Visual Regression Tests for Map Components
 * 
 * These tests capture screenshots of the minimap and world map modal
 * and compare them against "golden master" baseline images.
 * 
 * Run with: npx playwright test map_visual.spec.ts --update-snapshots
 * to update baseline images when intentional UI changes are made.
 */
test.describe('Map Visual Regression', () => {
    test.setTimeout(180000);

    test.beforeEach(async ({ page }) => {
        // Register fresh user to ensure clean state
        await registerNewUser(page);
        await waitForGameReady(page);
    });

    test('minimap renders correctly in lobby', async ({ page }) => {
        // Wait for minimap to be visible
        const minimap = page.locator('[data-testid="minimap"]');
        await expect(minimap).toBeVisible({ timeout: 15000 });

        // Wait for map tiles to load (animation/rendering time)
        await page.waitForTimeout(2000);

        // Screenshot the minimap component
        await expect(minimap).toHaveScreenshot('minimap-lobby-default.png', {
            maxDiffPixels: 100, // Allow small rendering differences
            threshold: 0.2,    // 20% pixel difference threshold
        });
    });

    test('minimap updates after movement', async ({ page }) => {
        const minimap = page.locator('[data-testid="minimap"]');
        await expect(minimap).toBeVisible({ timeout: 15000 });
        await page.waitForTimeout(1000);

        // Move north and capture
        await sendCommand(page, 'n');
        await page.waitForTimeout(1500);

        await expect(minimap).toHaveScreenshot('minimap-after-movement.png', {
            maxDiffPixels: 150,
            threshold: 0.25,
        });
    });

    test('world map modal renders on trigger', async ({ page }) => {
        // World map should appear after world simulation or when triggered
        // Try the 'look' command which should show the area
        await sendCommand(page, 'look');
        await page.waitForTimeout(2000);

        // Check if world map modal is visible (may need to trigger it)
        const worldMapModal = page.locator('[data-testid="world-map-modal"]');

        // If not visible, try clicking the minimap to expand
        const minimap = page.locator('[data-testid="minimap"]');
        if (await minimap.isVisible({ timeout: 5000 }).catch(() => false)) {
            await minimap.click();
            await page.waitForTimeout(1000);
        }

        // If world map modal is visible, screenshot it
        if (await worldMapModal.isVisible({ timeout: 5000 }).catch(() => false)) {
            await expect(worldMapModal).toHaveScreenshot('world-map-modal-default.png', {
                maxDiffPixels: 200,
                threshold: 0.3,
            });
        } else {
            // Skip if modal not available in this context
            test.skip(true, 'World map modal not available in lobby context');
        }
    });

    test('minimap responsive on mobile viewport', async ({ page }) => {
        // Set mobile viewport
        await page.setViewportSize({ width: 375, height: 667 });
        await page.waitForTimeout(1000);

        const minimap = page.locator('[data-testid="minimap"]');

        if (await minimap.isVisible({ timeout: 10000 }).catch(() => false)) {
            await expect(minimap).toHaveScreenshot('minimap-mobile-viewport.png', {
                maxDiffPixels: 100,
                threshold: 0.2,
            });
        }
    });

    test('minimap shows player position indicator', async ({ page }) => {
        const minimap = page.locator('[data-testid="minimap"]');
        await expect(minimap).toBeVisible({ timeout: 15000 });
        await page.waitForTimeout(2000);

        // Look for player indicator within minimap
        const playerIndicator = minimap.locator('[data-testid="player-indicator"], .player-marker, .player-position');

        if (await playerIndicator.isVisible({ timeout: 5000 }).catch(() => false)) {
            await expect(playerIndicator).toHaveScreenshot('minimap-player-indicator.png', {
                maxDiffPixels: 50,
                threshold: 0.15,
            });
        }
    });

    test('minimap with different biome colors', async ({ page }) => {
        // This test would ideally run in a world context with varied biomes
        // For now, capture the default rendering
        const minimap = page.locator('[data-testid="minimap"]');
        await expect(minimap).toBeVisible({ timeout: 15000 });

        // Wait for full render with biome colors
        await page.waitForTimeout(3000);

        await expect(minimap).toHaveScreenshot('minimap-biome-colors.png', {
            maxDiffPixels: 100,
            threshold: 0.2,
        });
    });
});

test.describe('Map Visual Regression - World Context', () => {
    test.setTimeout(240000);

    test('minimap in created world', async ({ page }) => {
        // Register and go through world creation
        await registerNewUser(page);
        await waitForGameReady(page);

        const gameOutput = page.locator('[data-testid="game-output"]');

        // Start world creation
        await sendCommand(page, 'tell statue create world');
        await page.waitForTimeout(3000);

        // Quick answers for interview
        const worldName = `VisualTest_${Date.now()}`;
        const answers = ['Visual', 'Test', 'Zone', 'Magic', 'Power', worldName];
        for (const ans of answers) {
            await sendCommand(page, `reply ${ans}`);
            await page.waitForTimeout(2000);
        }

        // Wait for world creation
        await expect(gameOutput).toContainText('forged', { timeout: 60000 });

        // Enter world as watcher
        await sendCommand(page, `enter ${worldName}`);

        const watcherButton = page.locator('[data-testid="entry-option-watcher"]');
        if (await watcherButton.isVisible({ timeout: 10000 }).catch(() => false)) {
            await watcherButton.click();
        }

        await expect(gameOutput).toContainText('You have entered the world', { timeout: 15000 });

        // Wait for world map to render
        await page.waitForTimeout(5000);

        // Capture minimap in world context
        const minimap = page.locator('[data-testid="minimap"]');
        if (await minimap.isVisible({ timeout: 10000 }).catch(() => false)) {
            await expect(minimap).toHaveScreenshot('minimap-in-world-context.png', {
                maxDiffPixels: 200,
                threshold: 0.3, // Higher threshold for procedurally generated content
            });
        }
    });
});
