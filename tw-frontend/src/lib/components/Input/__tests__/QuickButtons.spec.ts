import { render, fireEvent } from '@testing-library/svelte';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import QuickButtons from '$lib/components/Input/QuickButtons.svelte';
import { gameWebSocket } from '$lib/services/websocket';

// Mock the WebSocket service
vi.mock('$lib/services/websocket', () => ({
    gameWebSocket: {
        sendRawCommand: vi.fn()
    }
}));

describe('QuickButtons', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('renders all default quick buttons', () => {
        const { getByRole } = render(QuickButtons);

        // Use getByRole with exact string matching to avoid ambiguous matches (e.g. South vs Southwest)
        expect(getByRole('button', { name: 'Look' })).toBeTruthy();
        expect(getByRole('button', { name: 'North' })).toBeTruthy();
        expect(getByRole('button', { name: 'South' })).toBeTruthy();
        expect(getByRole('button', { name: 'Inventory' })).toBeTruthy();
    });

    it('dispatches submit event when Look button is clicked', async () => {
        const { getByRole, component } = render(QuickButtons);
        const lookButton = getByRole('button', { name: 'Look' });
        const mockSubmit = vi.fn();
        component.$on('submit', mockSubmit);

        await fireEvent.click(lookButton);

        expect(mockSubmit).toHaveBeenCalled();
        expect(mockSubmit.mock.calls[0][0].detail).toBe('look');
    });

    it('dispatches submit event when North button is clicked', async () => {
        const { getByRole, component } = render(QuickButtons);
        const northButton = getByRole('button', { name: 'North' });
        const mockSubmit = vi.fn();
        component.$on('submit', mockSubmit);

        await fireEvent.click(northButton);

        expect(mockSubmit).toHaveBeenCalled();
        expect(mockSubmit.mock.calls[0][0].detail).toBe('north');
    });

    it('dispatches submit event when South button is clicked', async () => {
        const { getByRole, component } = render(QuickButtons);
        const southButton = getByRole('button', { name: 'South' });
        const mockSubmit = vi.fn();
        component.$on('submit', mockSubmit);

        await fireEvent.click(southButton);

        expect(mockSubmit).toHaveBeenCalled();
        expect(mockSubmit.mock.calls[0][0].detail).toBe('south');
    });

    it('dispatches submit event when Inventory button is clicked', async () => {
        const { getByRole, component } = render(QuickButtons);
        const inventoryButton = getByRole('button', { name: 'Inventory' });
        const mockSubmit = vi.fn();
        component.$on('submit', mockSubmit);

        await fireEvent.click(inventoryButton);

        expect(mockSubmit).toHaveBeenCalled();
        expect(mockSubmit.mock.calls[0][0].detail).toBe('inventory');
    });

    it('renders custom commands when provided', () => {
        const customCommands = [
            { label: 'Attack', command: 'attack' },
            { label: 'Defend', command: 'defend' }
        ];

        const { getByText, queryByText } = render(QuickButtons, {
            props: { commands: customCommands }
        });

        expect(getByText('Attack')).toBeTruthy();
        expect(getByText('Defend')).toBeTruthy();
        expect(queryByText('Look')).toBeNull();
        expect(queryByText('North')).toBeNull();
    });

    it('dispatches custom commands correctly', async () => {
        const customCommands = [
            { label: 'Cast Spell', command: 'cast fireball' }
        ];

        const { getByText, component } = render(QuickButtons, {
            props: { commands: customCommands }
        });

        const spellButton = getByText('Cast Spell');
        const mockSubmit = vi.fn();
        component.$on('submit', mockSubmit);

        await fireEvent.click(spellButton);

        expect(mockSubmit).toHaveBeenCalled();
        expect(mockSubmit.mock.calls[0][0].detail).toBe('cast fireball');
    });


    it('all buttons meet iOS minimum touch target size', () => {
        const { container } = render(QuickButtons);
        const buttons = container.querySelectorAll('button');

        buttons.forEach(button => {
            // Check that buttons have adequate touch target size (either w-12/48px or min-w-[44px])
            const classes = button.className;
            const hasWidthObject = classes.includes('w-12') || classes.includes('min-w-[44px]');
            const hasHeightObject = classes.includes('h-12') || classes.includes('min-h-[44px]');

            expect(hasWidthObject).toBe(true);
            expect(hasHeightObject).toBe(true);
        });
    });

    it('buttons have proper ARIA labels', () => {
        const { container } = render(QuickButtons);
        const buttons = container.querySelectorAll('button');

        buttons.forEach(button => {
            expect(button.getAttribute('aria-label')).toBeTruthy();
        });
    });
});
