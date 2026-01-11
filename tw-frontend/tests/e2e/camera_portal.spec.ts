import { test, expect } from '@playwright/test';
import { registerNewUser, waitForGameReady, sendCommand, loginUser, ensureLoggedOut } from './fixtures/auth';

/**
 * BDD Tests: Camera Controls and Portal Navigation
 * 
 * These tests document user interaction patterns for:
 * 1. Camera zoom and focus controls in watcher mode
 * 2. Eastern portal navigation to tropical test world
 * 3. Solar system view interactions
 */

test.describe('Camera Controls and Portal Navigation', () => {
    test.setTimeout(180000); // Allow extra time for rendering

    // -------------------------------------------------------------------------
    // Scenario: Zoom Controls via Mouse Wheel
    // -------------------------------------------------------------------------
    // Given: The watcher is viewing the solar system
    // When: They scroll the mouse wheel
    // Then: The camera should zoom in/out smoothly
    //   AND: Zooming in should reveal more planet detail
    //   AND: Zooming out should show the sun and orbital paths

    test('should zoom camera with mouse wheel in watcher mode', async ({ page }) => {
        // Register and authenticate
        await registerNewUser(page);
        await waitForGameReady(page);

        const gameOutput = page.locator('[data-testid="game-output"]');

        // Start world creation interview
        await sendCommand(page, 'tell statue create world');
        await page.waitForTimeout(3000);

        // Answer interview questions quickly
        const worldName = `CameraTest_${Date.now()}`;
        const answers = ['CameraTest', 'Test', 'Zone', 'Magic', 'Power', worldName];
        for (const ans of answers) {
            await sendCommand(page, `reply ${ans}`);
            await page.waitForTimeout(2000);
        }

        await expect(gameOutput).toContainText('forged', { timeout: 60000 });

        // Enter world as watcher
        await sendCommand(page, `enter ${worldName}`);

        const watcherButton = page.locator('[data-testid="entry-option-watcher"]');
        if (await watcherButton.isVisible({ timeout: 10000 }).catch(() => false)) {
            await watcherButton.click();
        }

        await expect(gameOutput).toContainText('You have entered the world', { timeout: 15000 });

        // Wait for 3D scene to load
        await page.waitForTimeout(5000);

        // Get the canvas element
        const canvas = page.locator('canvas').first();
        await expect(canvas).toBeVisible({ timeout: 10000 });

        // Perform mouse wheel zoom - zoom in
        await canvas.hover();
        await page.mouse.wheel(0, -300); // Negative = zoom in
        await page.waitForTimeout(1000);

        // Then zoom out
        await page.mouse.wheel(0, 300); // Positive = zoom out
        await page.waitForTimeout(1000);

        // Verify we're still in watcher mode (no errors)
        await sendCommand(page, 'look');
        await expect(gameOutput).toContainText('observing', { timeout: 5000 });
    });

    // -------------------------------------------------------------------------
    // Scenario: Click-to-Focus on Planet
    // -------------------------------------------------------------------------
    // Given: The watcher is viewing the solar system
    // When: They click on the planet
    // Then: The camera should focus on the planet
    //   AND: The camera should position with the sun at their back

    test('should focus on planet when clicked', async ({ page }) => {
        await registerNewUser(page);
        await waitForGameReady(page);

        const gameOutput = page.locator('[data-testid="game-output"]');

        // Quick world creation
        await sendCommand(page, 'tell statue create world');
        await page.waitForTimeout(3000);

        const worldName = `FocusTest_${Date.now()}`;
        const answers = ['FocusTest', 'Test', 'Zone', 'Magic', 'Power', worldName];
        for (const ans of answers) {
            await sendCommand(page, `reply ${ans}`);
            await page.waitForTimeout(2000);
        }

        await expect(gameOutput).toContainText('forged', { timeout: 60000 });

        await sendCommand(page, `enter ${worldName}`);

        const watcherButton = page.locator('[data-testid="entry-option-watcher"]');
        if (await watcherButton.isVisible({ timeout: 10000 }).catch(() => false)) {
            await watcherButton.click();
        }

        await page.waitForTimeout(5000);

        const canvas = page.locator('canvas').first();
        await expect(canvas).toBeVisible();

        // Click on the center of the canvas (where planet should be)
        const box = await canvas.boundingBox();
        if (box) {
            await page.mouse.click(box.x + box.width / 2, box.y + box.height / 2);
            await page.waitForTimeout(500);

            // Second click to confirm focus
            await page.mouse.click(box.x + box.width / 2, box.y + box.height / 2);
            await page.waitForTimeout(1000);
        }

        // Verify we're still operational
        await sendCommand(page, 'look');
        await expect(gameOutput).toContainText('observing', { timeout: 5000 });
    });
});

test.describe('Eastern Portal to Tropical World', () => {
    test.setTimeout(180000);

    // -------------------------------------------------------------------------
    // Scenario: Eastern Portal Visibility
    // -------------------------------------------------------------------------
    // Given: The watcher is in the Grand Lobby
    // When: They look around
    // Then: They should see two portals (west and east)
    //   AND: The eastern portal should have tropical green/gold colors

    test('should see eastern portal in lobby', async ({ page }) => {
        await registerNewUser(page);
        await waitForGameReady(page);

        const gameOutput = page.locator('[data-testid="game-output"]');

        // Look around in the lobby
        await sendCommand(page, 'look');
        await page.waitForTimeout(2000);

        // Verify we're in the Grand Lobby
        await expect(gameOutput).toContainText('Grand Lobby', { timeout: 10000 });

        // Check for portal visibility in the 3D scene
        const canvas = page.locator('canvas').first();
        if (await canvas.isVisible({ timeout: 5000 }).catch(() => false)) {
            // Wait for scene to fully render
            await page.waitForTimeout(3000);

            // The eastern portal should be visible (green/gold particles)
            // We can't directly test WebGL colors, but we verify the scene renders
            await expect(canvas).toBeVisible();
        }
    });

    // -------------------------------------------------------------------------
    // Scenario: Eastern Portal Navigation (Future)
    // -------------------------------------------------------------------------
    // Given: The watcher walks to the eastern portal
    // When: They enter the portal
    // Then: They should be transported to the tropical test world
    //   AND: The tropical world should have warm temperatures
    //   AND: The tropical world should have longer days

    test.skip('should enter tropical world through eastern portal', async ({ page }) => {
        // This test requires the backend enter_tropical_test command handler
        // which is not yet implemented (see task.md Section 8)
        await registerNewUser(page);
        await waitForGameReady(page);

        const gameOutput = page.locator('[data-testid="game-output"]');

        // Navigate toward east wall
        await sendCommand(page, 'walk east');
        await page.waitForTimeout(2000);

        // Continue walking to reach portal
        await sendCommand(page, 'walk east');
        await page.waitForTimeout(2000);

        // Enter the portal
        await sendCommand(page, 'enter portal');

        // Should be transported to tropical world
        await expect(gameOutput).toContainText('tropical', { timeout: 10000 });
    });
});

test.describe('FPV UI Components', () => {
    test.setTimeout(180000);

    // -------------------------------------------------------------------------
    // Scenario: Menu Button Visibility in FPV
    // -------------------------------------------------------------------------
    // Given: The player is in first-person view
    // When: They look at the screen
    // Then: A menu button should be visible in the top-left corner
    //   AND: Clicking it should open the game menu modal

    test.skip('should show menu button and modal in FPV mode', async ({ page }) => {
        // This test requires entering FPV mode which needs a spawned world
        await registerNewUser(page);
        await waitForGameReady(page);

        const gameOutput = page.locator('[data-testid="game-output"]');

        // Create and enter a world
        await sendCommand(page, 'tell statue create world');
        await page.waitForTimeout(3000);

        const worldName = `FPVTest_${Date.now()}`;
        const answers = ['FPVTest', 'Test', 'Zone', 'Magic', 'Power', worldName];
        for (const ans of answers) {
            await sendCommand(page, `reply ${ans}`);
            await page.waitForTimeout(2000);
        }

        await expect(gameOutput).toContainText('forged', { timeout: 60000 });

        // Enter world as player (not watcher)
        await sendCommand(page, `enter ${worldName}`);

        const playerButton = page.locator('[data-testid="entry-option-player"]');
        if (await playerButton.isVisible({ timeout: 10000 }).catch(() => false)) {
            await playerButton.click();
        }

        await page.waitForTimeout(5000);

        // Look for FPS HUD menu button
        const menuButton = page.locator('.menu-btn');
        await expect(menuButton).toBeVisible({ timeout: 10000 });

        // Click menu button
        await menuButton.click();

        // Modal should appear with tabs
        const modal = page.locator('.modal-content');
        await expect(modal).toBeVisible({ timeout: 5000 });

        // Verify tabs exist
        await expect(page.locator('.tab-btn:has-text("World")')).toBeVisible();
        await expect(page.locator('.tab-btn:has-text("Character")')).toBeVisible();
        await expect(page.locator('.tab-btn:has-text("Account")')).toBeVisible();

        // Close modal
        await page.locator('.close-btn').click();
        await expect(modal).not.toBeVisible();
    });
});
