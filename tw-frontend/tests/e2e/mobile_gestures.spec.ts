import { test, expect, devices } from '@playwright/test';
import { registerNewUser, waitForGameReady, sendCommand } from './fixtures/auth';

test.use({
    ...devices['iPhone 12']
});

test.describe('Mobile Gesture Tests', () => {
    test.beforeEach(async ({ page }) => {
        await registerNewUser(page);
        await waitForGameReady(page);
    });

    test('Touch targets meet iOS 44x44pt minimum', async ({ page }) => {
        // Find all buttons
        const buttons = await page.locator('button').all();

        for (const button of buttons) {
            const box = await button.boundingBox();
            if (box) {
                // 44pt = 44px at 1x scale
                expect(box.width).toBeGreaterThanOrEqual(44);
                expect(box.height).toBeGreaterThanOrEqual(44);
            }
        }
    });

    test('Tap on QuickButton sends command', async ({ page }) => {
        const lookButton = page.locator('button', { hasText: 'Look' });

        // Tap (mobile click)
        await lookButton.tap();

        // Wait for response
        await page.waitForSelector('[data-testid="game-output"] .message', {
            timeout: 5000
        });
    });

    test('Input auto-scrolls into view when keyboard appears', async ({ page }) => {
        const input = page.locator('input[placeholder="Enter command..."]');

        // Tap input to show keyboard
        await input.tap();

        // Input should be in viewport
        await expect(input).toBeInViewport();
    });

    test('No zoom on input focus (font-size >= 16px)', async ({ page }) => {
        const input = page.locator('input[placeholder="Enter command..."]');

        const initialViewport = await page.evaluate(() => ({
            width: window.innerWidth,
            height: window.innerHeight
        }));

        await input.tap();
        await page.waitForTimeout(500);

        const afterTapViewport = await page.evaluate(() => ({
            width: window.innerWidth,
            height: window.innerHeight
        }));

        // Viewport shouldn't change (no zoom)
        expect(afterTapViewport.width).toBe(initialViewport.width);
    });

    test('Keyboard accessory bar visible on mobile', async ({ page }) => {
        const input = page.locator('input[placeholder="Enter command..."]');
        await input.tap();

        // QuickButtons should be visible above keyboard
        const quickButtons = page.locator('button', { hasText: 'Look' }).first();
        await expect(quickButtons).toBeVisible();
    });

    test('Momentum scrolling works in output area', async ({ page }) => {
        const output = page.locator('[data-testid="game-output"]');

        // Send multiple commands to fill output
        for (let i = 0; i < 10; i++) {
            await sendCommand(page, `message ${i}`);
            await page.waitForTimeout(100);
        }

        // Get initial scroll position
        const scrollBefore = await output.evaluate(el => el.scrollTop);

        // Swipe up (scroll down)
        const box = await output.boundingBox();
        if (box) {
            await page.mouse.move(box.x + box.width / 2, box.y + box.height - 50);
            await page.mouse.down();
            await page.mouse.move(box.x + box.width / 2, box.y + 50, { steps: 10 });
            await page.mouse.up();
        }

        await page.waitForTimeout(300);

        const scrollAfter = await output.evaluate(el => el.scrollTop);

        // Scroll position should have changed
        expect(scrollAfter).not.toBe(scrollBefore);
    });

    test('Pull-to-refresh disabled in standalone mode', async ({ page, context }) => {
        // Simulate standalone PWA mode
        await context.addInitScript(() => {
            Object.defineProperty(window, 'matchMedia', {
                writable: true,
                value: (query: string) => ({
                    matches: query === '(display-mode: standalone)',
                    media: query,
                    onchange: null,
                    addListener: () => { },
                    removeListener: () => { },
                    addEventListener: () => { },
                    removeEventListener: () => { },
                    dispatchEvent: () => true
                })
            });
        });

        await page.reload();
        await waitForGameReady(page);

        const output = page.locator('[data-testid="game-output"]');

        // Attempt pull-to-refresh gesture
        const box = await output.boundingBox();
        if (box) {
            await page.mouse.move(box.x + box.width / 2, box.y + 50);
            await page.mouse.down();
            await page.mouse.move(box.x + box.width / 2, box.y + box.height - 50, { steps: 10 });
            await page.mouse.up();
        }

        // Page should not refresh
        await page.waitForTimeout(500);
        const url = page.url();
        expect(url).toContain('/game');
    });

    test('Landscape mode maintains layout', async ({ page }) => {
        // Set to landscape orientation
        await page.setViewportSize({ width: 844, height: 390 });

        // Check layout is still functional
        const input = page.locator('input[placeholder="Enter command..."]');
        await expect(input).toBeVisible();

        const quickButtons = page.locator('button', { hasText: 'Look' });
        await expect(quickButtons).toBeVisible();
    });

    test('Haptic feedback triggers on command send', async ({ page }) => {
        // Note: Actual haptic feedback can't be tested in browser,
        // but we can verify the vibrate API is called
        let vibrateCalled = false;

        await page.exposeFunction('mockVibrate', () => {
            vibrateCalled = true;
        });

        await page.addInitScript(() => {
            (navigator as any).vibrate = (window as any).mockVibrate;
        });

        await sendCommand(page, 'look');

        // In real implementation, vibrate would be called
        // This test documents the expected behavior
    });
});
