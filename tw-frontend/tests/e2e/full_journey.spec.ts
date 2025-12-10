import { test, expect } from '@playwright/test';
import {
    registerNewUser,
    waitForGameReady,
    sendCommand,
    loginUser,
    ensureLoggedOut,
    type AuthCredentials
} from './fixtures/auth';

/**
 * Comprehensive End-to-End User Journey Test
 * Matches the backend mobile_user_journey_test.go pattern
 * 
 * Tests complete flow:
 * 1. User Registration & Account Creation
 * 2. User Login & Session Management
 * 3. Lobby Commands & WebSocket Communication
 * 4. Spatial Movement in World
 * 5. World Creation Interview
 * 6. World Generation & Verification
 * 7. World Entry (Watcher/Character)
 * 8. Full Game Experience
 */

test.describe('Complete MUD Platform E2E Journey', () => {
    test('Step 1-8: Complete User Journey - Registration through Gameplay', async ({ page }) => {
        let worldName = '';
        let creds: AuthCredentials;
        test.setTimeout(600000); // Allow 10 minutes for full journey with LLM

        const gameOutput = page.locator('[data-testid="game-output"]');

        // ==================== STEP 1: CREATE ACCOUNT ====================
        await test.step('Step 1: Create New User Account', async () => {
            creds = await registerNewUser(page);
            await waitForGameReady(page);
        });

        // ==================== STEP 2: VERIFY LOGIN (Logout and Login Again) ====================
        await test.step('Step 2: Logout and Login to Verify Session', async () => {
            // Logout
            const logoutButton = page.locator('button:has-text("Logout"), button:has-text("Sign Out")').first();
            if (await logoutButton.isVisible({ timeout: 2000 }).catch(() => false)) {
                await logoutButton.click();
                await page.waitForURL('/', { timeout: 5000 }).catch(() => { });
            } else {
                await ensureLoggedOut(page);
            }

            // Login with created credentials
            await loginUser(page, creds.email, creds.password);
            await waitForGameReady(page);
        });

        // ==================== STEP 3: LOBBY COMMANDS ====================
        await test.step('Step 3: Test Lobby Commands via WebSocket', async () => {
            // Test LOOK command
            await sendCommand(page, 'look');
            await page.waitForTimeout(2000);

            // Test SAY command
            await sendCommand(page, 'say Hello E2E!');
            await page.waitForTimeout(2000);

            // Test WHO command
            await sendCommand(page, 'who');
            await page.waitForTimeout(2000);
        });

        // ==================== STEP 3.1: SPATIAL MOVEMENT ====================
        await test.step('Step 3.1: Test Spatial Movement', async () => {
            const directions = ['north', 'east', 'south', 'west'];
            for (const dir of directions) {
                await sendCommand(page, dir);
                await page.waitForTimeout(1000);
            }

            // Test abbreviated directions
            await sendCommand(page, 'n');
            await page.waitForTimeout(1000);
        });

        // ==================== STEP 4: WORLD CREATION INTERVIEW ====================
        await test.step('Step 4: Complete World Creation Interview', async () => {
            // Start interview
            await sendCommand(page, 'tell statue I want to create a world');
            await expect(gameOutput).toContainText('voice resonates', { timeout: 120000 });

            // Answer interview questions
            const answers = [
                'A high-tech cyberpunk world with neon cities', // Core Concept
                'Humans and AI entities', // Sentient Species 
                'Urban sprawl with some wilderness preserves', // Environment
                'High tech, some experimental nanotech magic', // Magic & Tech
                'Corporate wars and AI rights movements', // Conflict
            ];

            for (let i = 0; i < answers.length; i++) {
                await page.waitForTimeout(15000); // Wait for LLM
                await sendCommand(page, `reply ${answers[i]}`);
            }

            // Final answer: World Name
            worldName = `E2EWorld-${Date.now()}`;
            await page.waitForTimeout(15000);
            await sendCommand(page, `reply ${worldName}`);

            // Wait for confirmation prompt
            await expect(gameOutput).toContainText('Is this correct?', { timeout: 60000 });

            // Confirm creation details
            await sendCommand(page, 'reply yes');
            await page.waitForTimeout(5000);
        });

        // ==================== STEP 5: VERIFY WORLD CREATION ====================
        await test.step('Step 5: Wait for World Generation', async () => {
            await expect(gameOutput).toContainText('Your world is being forged', { timeout: 30000 });
        });

        // ==================== STEP 6: ENTER WORLD ====================
        await test.step('Step 6: Enter Created World', async () => {
            await sendCommand(page, 'look');
            await sendCommand(page, `enter ${worldName}`);

            // Wait for entry options modal and click Watcher
            const watcherButton = page.locator('[data-testid="entry-option-watcher"]');
            await expect(watcherButton).toBeVisible({ timeout: 10000 });
            await watcherButton.click();

            // Wait for entry confirmation
            await expect(gameOutput).toContainText('You have entered the world!', { timeout: 30000 });
        });

        // ==================== STEP 8: GAMEPLAY VERIFICATION ====================
        await test.step('Step 8: Verify Full Game Experience', async () => {
            await page.waitForTimeout(2000);

            // Test inventory
            await sendCommand(page, 'inventory');

            // Test status
            await sendCommand(page, 'status');

            // Check for Character Modal
            const modal = page.locator('.character-sheet');
            await expect(modal).toBeVisible({ timeout: 5000 });

            // Check Tabs
            const statsTab = modal.locator('button', { hasText: 'Stats' });
            const skillsTab = modal.locator('button', { hasText: 'Skills' });
            await expect(statsTab).toBeVisible();
            await expect(skillsTab).toBeVisible();

            // Default: Stats active
            await expect(modal.locator('text=Strength')).toBeVisible();

            // Click Skills
            await skillsTab.click();
            await expect(modal.locator('text=Strength')).not.toBeVisible();
            await expect(modal.locator('text=No skills learned').or(modal.locator('text=Skills'))).toBeVisible();

            // Close modal
            await page.locator('[data-testid="close-character-sheet"]').click();
            await expect(modal).not.toBeVisible();

            // Test look
            await sendCommand(page, 'look');

            // Verify input still works
            const commandInput = page.locator('input[placeholder="Enter command..."]');
            await expect(commandInput).toBeEditable();
        });

        // ==================== STEP 9: VERIFY AUTO-RESUME ====================
        await test.step('Step 9: Verify Auto-Resume on Login', async () => {
            // Logout
            const logoutButton = page.locator('button', { hasText: 'Logout' });
            await logoutButton.click();
            await page.waitForURL('/', { timeout: 5000 });

            // Login again
            await loginUser(page, creds.email, creds.password);

            // Verify we auto-resume
            await expect(gameOutput).toContainText('Resuming adventure', { timeout: 15000 });
            await expect(page.locator('input[placeholder="Enter command..."]').first()).toBeVisible();

            // Verify we are in the world
            await sendCommand(page, 'look');
            await expect(gameOutput).toContainText(worldName, { timeout: 10000 });
        });
    });

    test('Returning User: Login and Resume Gameplay', async ({ page }) => {
        await test.step('Login as Returning User', async () => {
            // Create a temporary user then log them back in
            const creds = await registerNewUser(page);
            await waitForGameReady(page);

            // Navigate to login page
            await page.goto('/');

            // Login again
            await loginUser(page, creds.email, creds.password);

            // Should be back in game
            await expect(page.locator('input[placeholder="Enter command..."]').first()).toBeVisible();
        });
    });

    test('Quick Command Usage Test', async ({ page }) => {
        await test.step('Use QuickButtons for Common Commands', async () => {
            await registerNewUser(page);
            await waitForGameReady(page);

            // Test Look button if available
            const lookButton = page.locator('button:has-text("Look")').first();
            if (await lookButton.isVisible({ timeout: 2000 }).catch(() => false)) {
                await lookButton.click();
                await page.waitForTimeout(2000);

                // Test other quick buttons
                const northButton = page.locator('button:has-text("North")').first();
                if (await northButton.isVisible({ timeout: 1000 }).catch(() => false)) {
                    await northButton.click();
                    await page.waitForTimeout(2000);
                }
            }
        });
    });

    test('Command History Navigation', async ({ page }) => {
        await test.step('Test Command History with Arrow Keys', async () => {
            await registerNewUser(page);
            await waitForGameReady(page);

            const commandInput = page.locator('input[placeholder="Enter command..."]').first();

            // Send multiple commands
            await sendCommand(page, 'look');
            await page.waitForTimeout(1000);

            await sendCommand(page, 'north');
            await page.waitForTimeout(1000);

            await sendCommand(page, 'south');
            await page.waitForTimeout(1000);

            // Navigate history with arrow up
            await commandInput.press('ArrowUp');
            await expect(commandInput).toHaveValue('south');

            await commandInput.press('ArrowUp');
            await expect(commandInput).toHaveValue('north');

            await commandInput.press('ArrowUp');
            await expect(commandInput).toHaveValue('look');

            // Navigate forward
            await commandInput.press('ArrowDown');
            await expect(commandInput).toHaveValue('north');
        });
    });
});

test.describe('Performance and Stress Testing', () => {
    test('Rapid Command Execution', async ({ page }) => {
        await registerNewUser(page);
        await waitForGameReady(page);

        const commandInput = page.locator('input[placeholder="Enter command..."]').first();

        // Send 10 commands rapidly
        for (let i = 0; i < 10; i++) {
            await commandInput.fill(`test ${i}`);
            await commandInput.press('Enter');
            await page.waitForTimeout(100);
        }

        // Input should still be responsive
        await expect(commandInput).toBeEditable();
    });

    test('Page Load Performance', async ({ page }) => {
        const startTime = Date.now();
        await registerNewUser(page);
        await waitForGameReady(page);
        const loadTime = Date.now() - startTime;

        expect(loadTime).toBeLessThan(30000); // Account for auth
    });
});
