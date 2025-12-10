import { writable } from 'svelte/store';

export type HapticType = 'light' | 'medium' | 'heavy' | 'selection' | 'error' | 'success';

interface HapticPattern {
    [key: string]: number | number[];
}

const patterns: HapticPattern = {
    light: 10,
    medium: 15,
    heavy: 20,
    selection: 5,
    error: [10, 50, 10],
    success: [10, 50, 10, 50, 10]
};

// Store for haptic feedback preference
export const hapticEnabled = writable<boolean>(true);

// Initialize from localStorage
if (typeof localStorage !== 'undefined') {
    const stored = localStorage.getItem('hapticEnabled');
    if (stored !== null) {
        hapticEnabled.set(stored === 'true');
    }
}

// Save preference changes
hapticEnabled.subscribe(value => {
    if (typeof localStorage !== 'undefined') {
        localStorage.setItem('hapticEnabled', String(value));
    }
});

let currentlyEnabled = true;
hapticEnabled.subscribe(value => {
    currentlyEnabled = value;
});

export function triggerHaptic(type: HapticType): void {
    if (!currentlyEnabled) return;
    if (!('vibrate' in navigator)) return;

    const pattern = patterns[type];
    if (Array.isArray(pattern)) {
        navigator.vibrate(pattern);
    } else {
        navigator.vibrate(pattern);
    }
}

// Convenience functions
export const haptic = {
    light: () => triggerHaptic('light'),
    medium: () => triggerHaptic('medium'),
    heavy: () => triggerHaptic('heavy'),
    selection: () => triggerHaptic('selection'),
    error: () => triggerHaptic('error'),
    success: () => triggerHaptic('success')
};
