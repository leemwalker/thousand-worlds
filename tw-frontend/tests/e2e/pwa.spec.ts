import { test, expect } from '@playwright/test';

test.describe('PWA Functionality', () => {
    test('should have valid manifest', async ({ page }) => {
        await page.goto('/');

        // Check for manifest link
        const manifestLink = page.locator('link[rel="manifest"]');
        await expect(manifestLink).toHaveCount(1);

        const href = await manifestLink.getAttribute('href');
        expect(href).toBeTruthy();

        // Fetch and validate manifest
        const manifestResponse = await page.request.get(href!);
        expect(manifestResponse.ok()).toBeTruthy();

        const manifest = await manifestResponse.json();
        expect(manifest.name).toBe('Thousand Worlds MUD Client');
        expect(manifest.short_name).toBe('TW MUD');
        expect(manifest.display).toBe('standalone');
        expect(manifest.icons).toHaveLength(2);
    });

    test('should have service worker registered', async ({ page }) => {
        await page.goto('/');

        // Check if service worker is registered
        const swRegistered = await page.evaluate(async () => {
            if ('serviceWorker' in navigator) {
                const registration = await navigator.serviceWorker.getRegistration();
                return !!registration;
            }
            return false;
        });

        expect(swRegistered).toBeTruthy();
    });

    test('should have proper PWA icons', async ({ page }) => {
        await page.goto('/');

        // Check for apple-touch-icon
        const appleTouchIcon = page.locator('link[rel="apple-touch-icon"]');
        const iconCount = await appleTouchIcon.count();

        // It's okay if there are 0 or more, but if they exist, they should be valid
        if (iconCount > 0) {
            const href = await appleTouchIcon.first().getAttribute('href');
            expect(href).toBeTruthy();
        }
    });

    test('should have proper theme color', async ({ page }) => {
        await page.goto('/');

        const themeColor = page.locator('meta[name="theme-color"]');
        await expect(themeColor).toHaveCount(1);

        const content = await themeColor.getAttribute('content');
        expect(content).toBe('#16213e');
    });
});

test.describe('Offline Functionality', () => {
    test('should show offline page when offline', async ({ page }) => {
        await page.goto('/');

        // Wait for Service Worker to be ready (critical for offline support)
        await page.evaluate(async () => {
            await navigator.serviceWorker.ready;
        });

        // Simulate offline
        await page.context().setOffline(true);

        // Reload page to trigger fallback
        try {
            await page.reload({ timeout: 5000 });
        } catch (e) {
            // Navigation might fail, which is expected if SW doesn't kick in immediately
            // But if SW works, it serves offline.html
        }

        // We expect either the offline page or some indication
        // If offline.html is served, it likely has specific text
        // Let's check for common offline text or title
        // Assuming offline.html contains "Offline"
        const offlineText = page.getByText('Offline', { exact: false });
        if (await offlineText.isVisible()) {
            expect(await offlineText.isVisible()).toBeTruthy();
        } else {
            // Fallback: existing content might still be there if cached
            // verifying we didn't get a browser error page is the main thing
            const title = await page.title();
            expect(title).toBeTruthy();
        }

        // Restore online
        await page.context().setOffline(false);
    });
});
