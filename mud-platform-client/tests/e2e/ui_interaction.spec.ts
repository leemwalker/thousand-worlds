import { test, expect } from '@playwright/test';

test.describe('UI Interaction E2E', () => {
    test.beforeEach(async ({ page }) => {
        // Navigate to game page (assumes local dev server running)
        await page.goto('http://localhost:5173/game');

        // Wait for WebSocket connection
        await page.waitForSelector('[data-testid="connection-status"].connected', {
            timeout: 5000
        });
    });

    test('CommandInput sends raw text via WebSocket', async ({ page }) => {
        // Find command input
        const input = page.locator('input[placeholder="Enter command..."]');
        await expect(input).toBeVisible();

        // Type command
        await input.fill('look');

        // Press Enter
        await input.press('Enter');

        // Input should clear after sending
        await expect(input).toHaveValue('');

        // Wait for response in output
        await page.waitForSelector('[data-testid="game-output"] .message', {
            timeout: 2000
        });
    });

    test('QuickButtons send commands on click', async ({ page }) => {
        // Find Look button
        const lookButton = page.locator('button', { hasText: 'Look' });
        await expect(lookButton).toBeVisible();

        // Click button
        await lookButton.click();

        // Wait for response
        await page.waitForSelector('[data-testid="game-output"] .message', {
            timeout: 2000
        });
    });

    test('Command history navigation with arrow keys', async ({ page }) => {
        const input = page.locator('input[placeholder="Enter command..."]');

        // Send first command
        await input.fill('look');
        await input.press('Enter');

        // Send second command
        await input.fill('north');
        await input.press('Enter');

        // Press ArrowUp to recall last command
        await input.press('ArrowUp');
        await expect(input).toHaveValue('north');

        // Press ArrowUp again for previous command
        await input.press('ArrowUp');
        await expect(input).toHaveValue('look');

        // Press ArrowDown to go forward
        await input.press('ArrowDown');
        await expect(input).toHaveValue('north');

        // Press ArrowDown to empty
        await input.press('ArrowDown');
        await expect(input).toHaveValue('');
    });

    test('Escape key clears input', async ({ page }) => {
        const input = page.locator('input[placeholder="Enter command..."]');

        await input.fill('test command');
        await input.press('Escape');

        await expect(input).toHaveValue('');
    });

    test('Send button disabled when input empty', async ({ page }) => {
        const input = page.locator('input[placeholder="Enter command..."]');
        const sendButton = page.locator('button', { hasText: 'Send' });

        // Initially disabled
        await expect(sendButton).toBeDisabled();

        // Type something
        await input.fill('look');
        await expect(sendButton).toBeEnabled();

        // Clear input
        await input.clear();
        await expect(sendButton).toBeDisabled();
    });

    test('Messages render in output area', async ({ page }) => {
        const input = page.locator('input[placeholder="Enter command..."]');
        const output = page.locator('[data-testid="game-output"]');

        // Send command
        await input.fill('look');
        await input.press('Enter');

        // Wait for message
        const message = output.locator('.message').first();
        await expect(message).toBeVisible({ timeout: 3000 });
    });

    test('Auto-scroll to bottom on new message', async ({ page }) => {
        const input = page.locator('input[placeholder="Enter command..."]');
        const output = page.locator('[data-testid="game-output"]');

        // Send multiple commands to fill output
        for (let i = 0; i < 5; i++) {
            await input.fill(`look ${i}`);
            await input.press('Enter');
            await page.waitForTimeout(200);
        }

        // Check if scrolled to bottom
        const scrollTop = await output.evaluate(el => el.scrollTop);
        const scrollHeight = await output.evaluate(el => el.scrollHeight);
        const clientHeight = await output.evaluate(el => el.clientHeight);

        expect(scrollTop + clientHeight).toBeGreaterThanOrEqual(scrollHeight - 10);
    });
});

test.describe('Performance E2E', () => {
    test.beforeEach(async ({ page }) => {
        await page.goto('http://localhost:5173/game');
        await page.waitForSelector('[data-testid="connection-status"].connected', {
            timeout: 5000
        });
    });

    test('Command response time under 100ms (excluding network)', async ({ page }) => {
        const input = page.locator('input[placeholder="Enter command..."]');

        const startTime = Date.now();
        await input.fill('look');
        await input.press('Enter');

        // Measure time for input to clear (client-side processing only)
        await expect(input).toHaveValue('');
        const endTime = Date.now();

        expect(endTime - startTime).toBeLessThan(100);
    });

    test('Handles rapid commands without lag', async ({ page }) => {
        const input = page.locator('input[placeholder="Enter command..."]');

        // Send 10 commands rapidly
        for (let i = 0; i < 10; i++) {
            await input.fill(`cmd${i}`);
            await input.press('Enter');
        }

        // All commands should be processed
        const messages = page.locator('[data-testid="game-output"] .message');
        await expect(messages).toHaveCount(10, { timeout: 10000 });
    });
});
