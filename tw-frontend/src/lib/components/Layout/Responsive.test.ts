import { describe, it, expect, beforeEach } from 'vitest';
import { get } from 'svelte/store';
import { uiState, isMobile, setScreenWidth } from '$lib/stores/ui';

describe('Responsive Layout Logic', () => {
    beforeEach(() => {
        // Reset store
        setScreenWidth(1024);
    });

    it('should default to desktop mode', () => {
        const state = get(uiState);
        expect(state.layoutMode).toBe('desktop');
        expect(get(isMobile)).toBe(false);
    });

    it('should switch to mobile mode when width < 769', () => {
        setScreenWidth(768);
        const state = get(uiState);
        expect(state.layoutMode).toBe('mobile');
        expect(get(isMobile)).toBe(true);
    });

    it('should switch back to desktop mode when width >= 769', () => {
        setScreenWidth(768); // Mobile first
        expect(get(isMobile)).toBe(true);

        setScreenWidth(769); // Switch to desktop
        const state = get(uiState);
        expect(state.layoutMode).toBe('desktop');
        expect(get(isMobile)).toBe(false);
    });

    it('should update screen width in store', () => {
        setScreenWidth(1920);
        expect(get(uiState).screenWidth).toBe(1920);
    });
});
