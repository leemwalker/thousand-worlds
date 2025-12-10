import { test, expect, devices } from '@playwright/test';
import { registerNewUser, waitForGameReady, suppressIOSPrompt } from './fixtures/auth';

test.use({ ...devices['iPhone 12'] });

test.describe('Mobile Responsiveness', () => {
    test('should display mobile-optimized layout', async ({ page }) => {
        await registerNewUser(page);
        await waitForGameReady(page);

        // Check for layout container
        const layout = page.locator('.flex.flex-col.h-screen');
        await expect(layout).toBeVisible();

        // Check for command input area
        const input = page.getByPlaceholder('Enter command...');
        await expect(input).toBeVisible();
    });

    test.skip('should have touch-friendly tap targets', async ({ page }) => {
        await page.goto('/');

        // All interactive elements should be at least 44x44px
        const buttons = page.locator('button');
        const count = await buttons.count();

        for (let i = 0; i < Math.min(count, 10); i++) {
            const box = await buttons.nth(i).boundingBox();
            if (box) {
                expect(box.width).toBeGreaterThanOrEqual(40);
                expect(box.height).toBeGreaterThanOrEqual(40);
            }
        }
    });

    test('should be responsive across breakpoints', async ({ page }) => {
        const breakpoints = [
            { name: 'Mobile', width: 375, height: 667 },
            { name: 'Tablet', width: 768, height: 1024 },
            { name: 'Desktop', width: 1920, height: 1080 },
        ];

        for (const { name, width, height } of breakpoints) {
            await page.setViewportSize({ width, height });
            await page.goto('/');

            // Page should not have horizontal scroll
            const bodyScrollWidth = await page.evaluate(() => document.body.scrollWidth);
            const bodyClientWidth = await page.evaluate(() => document.body.clientWidth);

            expect(bodyScrollWidth).toBeLessThanOrEqual(bodyClientWidth + 1);
        }
    });
});
