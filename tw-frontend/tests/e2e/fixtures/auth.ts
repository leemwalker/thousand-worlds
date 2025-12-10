/**
 * Centralized Authentication Fixtures for E2E Tests
 * 
 * This module provides robust, reusable authentication helpers that all
 * Playwright E2E tests should use for consistent and reliable auth flows.
 */

import { type Page, test as base, expect } from '@playwright/test';
import { randomUUID } from 'crypto';

export interface AuthCredentials {
    email: string;
    username: string;
    password: string;
}

/**
 * Generate unique test credentials
 */
export function generateCredentials(prefix = 'e2e'): AuthCredentials {
    const uuid = randomUUID().substring(0, 8);
    const timestamp = Date.now();
    return {
        email: `${prefix}_${timestamp}_${uuid}@example.com`,
        username: `${prefix}_${uuid}`,
        password: 'SecurePass123!'
    };
}

/**
 * Suppress iOS PWA install prompt that can interfere with tests
 */
export async function suppressIOSPrompt(page: Page): Promise<void> {
    await page.addInitScript(() => {
        localStorage.setItem('iosInstallPromptDismissed', String(Date.now()));
    });
}

/**
 * Perform registration flow with robust waits and error handling
 * 
 * @param page - Playwright page object
 * @param credentials - Optional partial credentials (missing fields auto-generated)
 * @returns The credentials used for registration
 */
export async function registerNewUser(
    page: Page,
    credentials?: Partial<AuthCredentials>
): Promise<AuthCredentials> {
    const creds = {
        ...generateCredentials(),
        ...credentials
    };

    // Suppress iOS PWA install prompt
    await suppressIOSPrompt(page);

    await page.goto('/');

    // Wait for page to be fully hydrated before interacting
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(500); // Brief hydration buffer for SvelteKit

    // Switch to Register mode if needed
    const toggleButton = page.locator('button').filter({ hasText: "Don't have an account?" });

    if (await toggleButton.isVisible({ timeout: 3000 }).catch(() => false)) {
        await toggleButton.click();
        await expect(page.getByRole('heading', { name: 'Create Account' }))
            .toBeVisible({ timeout: 5000 });
    }

    // Fill registration form with explicit waits
    const emailInput = page.locator('input[type="email"]').first();
    await emailInput.waitFor({ state: 'visible', timeout: 10000 });
    await emailInput.fill(creds.email);

    const usernameInput = page.getByLabel('Username');
    await usernameInput.waitFor({ state: 'visible', timeout: 5000 });
    await usernameInput.fill(creds.username);

    const passwordInput = page.locator('input[type="password"]').first();
    await passwordInput.fill(creds.password);

    // Submit registration
    const submitButton = page.locator('button[type="submit"]', { hasText: 'Create Account' });
    await submitButton.click();

    // Wait for successful navigation to game
    await page.waitForURL(/\/game/, { timeout: 30000 });

    // Verify game interface loaded
    await page.waitForSelector('[data-testid="game-output"]', {
        state: 'visible',
        timeout: 30000
    });

    return creds;
}

/**
 * Login with existing credentials
 * 
 * @param page - Playwright page object
 * @param email - User's email
 * @param password - User's password
 */
export async function loginUser(
    page: Page,
    email: string,
    password: string
): Promise<void> {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(500); // Hydration buffer

    // Ensure we're on login form (not register)
    const signInHeading = page.getByRole('heading', { name: 'Sign In' });
    if (!await signInHeading.isVisible({ timeout: 2000 }).catch(() => false)) {
        // May need to toggle to login mode
        const toggleButton = page.locator('button').filter({ hasText: 'Already have an account?' });
        if (await toggleButton.isVisible().catch(() => false)) {
            await toggleButton.click();
            await expect(signInHeading).toBeVisible({ timeout: 5000 });
        }
    }

    const emailInput = page.locator('input[type="email"]').first();
    await emailInput.waitFor({ state: 'visible', timeout: 10000 });
    await emailInput.fill(email);

    const passwordInput = page.locator('input[type="password"]').first();
    await passwordInput.fill(password);

    const loginButton = page.locator('button[type="submit"]').first();
    await loginButton.click();

    // Wait for redirect to game
    await page.waitForURL(/\/game/, { timeout: 15000 });
    await page.waitForSelector('[data-testid="game-output"]', {
        state: 'visible',
        timeout: 30000
    });
}

/**
 * Ensure user is logged out and on landing page
 */
export async function ensureLoggedOut(page: Page): Promise<void> {
    // Clear auth state
    await page.context().clearCookies();
    await page.goto('/');
    await page.evaluate(() => {
        localStorage.clear();
        sessionStorage.clear();
    });
}

/**
 * Wait for game terminal to be fully ready for input
 */
export async function waitForGameReady(page: Page): Promise<void> {
    await page.waitForSelector('[data-testid="game-output"]', {
        state: 'visible',
        timeout: 30000
    });
    await page.waitForSelector('input[placeholder="Enter command..."]', {
        state: 'visible',
        timeout: 30000
    });
    // Brief wait for WebSocket to connect
    await page.waitForTimeout(1000);
}

/**
 * Send a game command and wait for response
 */
export async function sendCommand(page: Page, command: string): Promise<void> {
    const input = page.locator('input[placeholder="Enter command..."]');
    await input.fill(command);
    await input.press('Enter');
}

/**
 * Wait for a message containing specific text to appear in game output
 */
export async function waitForMessage(
    page: Page,
    text: string,
    timeout = 30000
): Promise<void> {
    await expect(page.locator('[data-testid="game-output"]'))
        .toContainText(text, { timeout });
}

// Extended test type with auth fixture for convenient authenticated tests
export const test = base.extend<{
    authenticatedPage: Page;
    authCredentials: AuthCredentials;
}>({
    authCredentials: async ({ }, use) => {
        const creds = generateCredentials();
        await use(creds);
    },

    authenticatedPage: async ({ page, authCredentials }, use) => {
        await registerNewUser(page, authCredentials);
        await use(page);
    }
});

export { expect };
