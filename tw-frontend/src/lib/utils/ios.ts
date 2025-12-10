// iOS Detection and Utilities

export function isIOS(): boolean {
    return /iPad|iPhone|iPod/.test(navigator.userAgent) && !(window as any).MSStream;
}

export function isStandalone(): boolean {
    return window.matchMedia('(display-mode: standalone)').matches ||
        (window.navigator as any).standalone === true;
}

export function getIOSVersion(): number | null {
    const match = navigator.userAgent.match(/OS (\d+)_/);
    return match ? parseInt(match[1], 10) : null;
}

export function hasNotch(): boolean {
    // Check for safe area insets
    const div = document.createElement('div');
    div.style.paddingTop = 'env(safe-area-inset-top)';
    document.body.appendChild(div);
    const hasInset = getComputedStyle(div).paddingTop !== '0px';
    document.body.removeChild(div);
    return hasInset;
}

export function hasDynamicIsland(): boolean {
    // iPhone 14 Pro and newer have Dynamic Island
    const version = getIOSVersion();
    return version !== null && version >= 16 && hasNotch();
}
