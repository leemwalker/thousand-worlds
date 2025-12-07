import { writable } from 'svelte/store';
import { isIOS, isStandalone } from '$lib/utils/ios';

export const showIOSInstallPrompt = writable<boolean>(false);

// Check if we should show the install prompt
export function checkIOSInstallPrompt(): void {
    if (isIOS() && !isStandalone()) {
        const dismissed = localStorage.getItem('iosInstallPromptDismissed');
        const dismissedTime = dismissed ? parseInt(dismissed, 10) : 0;
        const weekAgo = Date.now() - (7 * 24 * 60 * 60 * 1000);

        // Show if never dismissed or dismissed more than a week ago
        if (!dismissed || dismissedTime < weekAgo) {
            showIOSInstallPrompt.set(true);
        }
    }
}

export function dismissIOSInstallPrompt(): void {
    showIOSInstallPrompt.set(false);
    localStorage.setItem('iosInstallPromptDismissed', String(Date.now()));
}
