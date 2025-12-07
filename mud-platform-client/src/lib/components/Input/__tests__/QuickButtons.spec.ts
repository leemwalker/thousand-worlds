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
        const { getByText } = render(QuickButtons);

        expect(getByText('Look')).toBeTruthy();
        expect(getByText('North')).toBeTruthy();
        expect(getByText('South')).toBeTruthy();
        expect(getByText('Inventory')).toBeTruthy();
    });

    it('sends correct command when Look button is clicked', async () => {
        const { getByText } = render(QuickButtons);
        const lookButton = getByText('Look');

        await fireEvent.click(lookButton);

        expect(gameWebSocket.sendRawCommand).toHaveBeenCalledWith('look');
    });

    it('sends correct command when North button is clicked', async () => {
        const { getByText } = render(QuickButtons);
        const northButton = getByText('North');

        await fireEvent.click(northButton);

        expect(gameWebSocket.sendRawCommand).toHaveBeenCalledWith('north');
    });

    it('sends correct command when South button is clicked', async () => {
        const { getByText } = render(QuickButtons);
        const southButton = getByText('South');

        await fireEvent.click(southButton);

        expect(gameWebSocket.sendRawCommand).toHaveBeenCalledWith('south');
    });

    it('sends correct command when Inventory button is clicked', async () => {
        const { getByText } = render(QuickButtons);
        const inventoryButton = getByText('Inventory');

        await fireEvent.click(inventoryButton);

        expect(gameWebSocket.sendRawCommand).toHaveBeenCalledWith('inventory');
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

    it('sends custom commands correctly', async () => {
        const customCommands = [
            { label: 'Cast Spell', command: 'cast fireball' }
        ];

        const { getByText } = render(QuickButtons, {
            props: { commands: customCommands }
        });

        const spellButton = getByText('Cast Spell');
        await fireEvent.click(spellButton);

        expect(gameWebSocket.sendRawCommand).toHaveBeenCalledWith('cast fireball');
    });


    it('all buttons meet iOS minimum touch target size', () => {
        const { container } = render(QuickButtons);
        const buttons = container.querySelectorAll('button');

        buttons.forEach(button => {
            // Check that buttons have the min-w-[44px] and min-h-[44px] classes
            const classes = button.className;
            expect(classes).toContain('min-w-[44px]');
            expect(classes).toContain('min-h-[44px]');
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
