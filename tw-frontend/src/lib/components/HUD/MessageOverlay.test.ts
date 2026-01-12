import { render, act } from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';
import MessageOverlay from './MessageOverlay.svelte';
import { gameOutput } from '$lib/stores/ui';

describe('MessageOverlay', () => {
    it('renders messages from store', async () => {
        const { getByText } = render(MessageOverlay);

        // Mock store update
        gameOutput.set([
            { id: '1', type: 'system', text: 'Welcome to the simulation', timestamp: 0 },
            { id: '2', type: 'error', text: 'Critical Error', timestamp: 1 }
        ]);

        // Wait for Svelte reactivity
        await act();

        expect(getByText('Welcome to the simulation')).toBeDefined();
        expect(getByText('Critical Error')).toBeDefined();
    });

    it('limits visible messages', async () => {
        const { container } = render(MessageOverlay);

        const messages = Array.from({ length: 10 }, (_, i) => ({
            id: String(i),
            type: 'system',
            text: `Message ${i}`,
            timestamp: i
        }));

        gameOutput.set(messages);
        await act();

        // MAX_VISIBLE is 4 in component
        const rendered = container.querySelectorAll('.message');
        expect(rendered.length).toBe(4);
    });
});
