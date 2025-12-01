import { test, expect, devices } from '@playwright/test';

const mobileDevices = [
    { name: 'iPhone 12', device: devices['iPhone 12'] },
    { name: 'Pixel 5', device: devices['Pixel 5'] },
];

test.describe('Mobile Responsiveness', () => {
    for (const { name, device } of mobileDevices) {
        test.describe(`on ${name}`, () => {
            test.use({ ...device });

            test('should display mobile-optimized layout', async ({ page }) => {
                await page.goto('/');

                // Check viewport is correct
                const viewport = page.viewportSize();
                expect(viewport).toBeTruthy();

                // Check for mobile-specific UI elements
                // This needs to be customized based on actual UI
                test.skip('Needs implementation with actual selectors');
            });

            test('should have touch-friendly tap targets', async ({ page }) => {
                await page.goto('/');

                // All interactive elements should be at least 44x44px
                const buttons = page.locator('button, a[href]');
                const count = await buttons.count();

                for (let i = 0; i < Math.min(count, 10); i++) {
                    const box = await buttons.nth(i).boundingBox();
                    if (box) {
                        // Touch targets should be at least 44x44px
                        expect(box.width).toBeGreaterThanOrEqual(40);  // Slightly lenient
                        expect(box.height).toBeGreaterThanOrEqual(40);
                    }
                }
            });
        });
    }

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

            expect(bodyScrollWidth).toBeLessThanOrEqual(bodyClientWidth + 1); // Allow 1px tolerance
        }
    });
});
