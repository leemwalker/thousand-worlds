import { render, fireEvent } from '@testing-library/svelte';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import CommandInput from '$lib/components/Input/CommandInput.svelte';
import { gameWebSocket } from '$lib/services/websocket';

// Mock the WebSocket service
vi.mock('$lib/services/websocket', () => ({
    gameWebSocket: {
        sendRawCommand: vi.fn()
    }
}));

describe('CommandInput', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('renders input field and send button', () => {
        const { getByPlaceholderText, getByText } = render(CommandInput);

        expect(getByPlaceholderText('Enter command...')).toBeTruthy();
        expect(getByText('Send')).toBeTruthy();
    });

    it('dispatches submit event when Send button is clicked', async () => {
        const { getByPlaceholderText, getByText, component } = render(CommandInput);
        const input = getByPlaceholderText('Enter command...') as HTMLInputElement;
        const sendButton = getByText('Send');
        const mockSubmit = vi.fn();
        component.$on('submit', mockSubmit);

        await fireEvent.input(input, { target: { value: 'look' } });
        await fireEvent.click(sendButton);

        expect(mockSubmit).toHaveBeenCalled();
        expect(mockSubmit.mock.calls[0][0].detail).toBe('look');
    });

    it('dispatches submit event when Enter key is pressed', async () => {
        const { getByPlaceholderText, component } = render(CommandInput);
        const input = getByPlaceholderText('Enter command...') as HTMLInputElement;
        const mockSubmit = vi.fn();
        component.$on('submit', mockSubmit);

        await fireEvent.input(input, { target: { value: 'north' } });
        await fireEvent.keyDown(input, { key: 'Enter' });

        expect(mockSubmit).toHaveBeenCalled();
        expect(mockSubmit.mock.calls[0][0].detail).toBe('north');
    });

    it('does not dispatch empty commands', async () => {
        const { getByPlaceholderText, component } = render(CommandInput);
        const input = getByPlaceholderText('Enter command...') as HTMLInputElement;
        const mockSubmit = vi.fn();
        component.$on('submit', mockSubmit);

        await fireEvent.input(input, { target: { value: '   ' } });
        await fireEvent.keyDown(input, { key: 'Enter' });

        expect(mockSubmit).not.toHaveBeenCalled();
    });

    it('clears input after sending command', async () => {
        const { getByPlaceholderText, component } = render(CommandInput);
        const input = getByPlaceholderText('Enter command...') as HTMLInputElement;
        const mockSubmit = vi.fn();
        component.$on('submit', mockSubmit);

        await fireEvent.input(input, { target: { value: 'inventory' } });
        await fireEvent.keyDown(input, { key: 'Enter' });

        expect(input.value).toBe('');
    });

    it('navigates command history with arrow keys', async () => {
        const { getByPlaceholderText } = render(CommandInput);
        const input = getByPlaceholderText('Enter command...') as HTMLInputElement;

        // Send first command
        await fireEvent.input(input, { target: { value: 'look' } });
        await fireEvent.keyDown(input, { key: 'Enter' });

        // Send second command
        await fireEvent.input(input, { target: { value: 'north' } });
        await fireEvent.keyDown(input, { key: 'Enter' });

        // Navigate up (should show 'north')
        await fireEvent.keyDown(input, { key: 'ArrowUp' });
        expect(input.value).toBe('north');

        // Navigate up again (should show 'look')
        await fireEvent.keyDown(input, { key: 'ArrowUp' });
        expect(input.value).toBe('look');

        // Navigate down (should show 'north')
        await fireEvent.keyDown(input, { key: 'ArrowDown' });
        expect(input.value).toBe('north');

        // Navigate down to empty
        await fireEvent.keyDown(input, { key: 'ArrowDown' });
        expect(input.value).toBe('');
    });

    it('clears input when Escape is pressed', async () => {
        const { getByPlaceholderText } = render(CommandInput);
        const input = getByPlaceholderText('Enter command...') as HTMLInputElement;

        await fireEvent.input(input, { target: { value: 'test command' } });
        await fireEvent.keyDown(input, { key: 'Escape' });

        expect(input.value).toBe('');
    });

    it('trims whitespace from commands', async () => {
        const { getByPlaceholderText, component } = render(CommandInput);
        const input = getByPlaceholderText('Enter command...') as HTMLInputElement;
        const mockSubmit = vi.fn();
        component.$on('submit', mockSubmit);

        await fireEvent.input(input, { target: { value: '  look around  ' } });
        await fireEvent.keyDown(input, { key: 'Enter' });

        expect(mockSubmit).toHaveBeenCalled();
        expect(mockSubmit.mock.calls[0][0].detail).toBe('look around');
    });

    it('disables send button when input is empty', async () => {
        const { getByText } = render(CommandInput);
        const sendButton = getByText('Send') as HTMLButtonElement;

        expect(sendButton.disabled).toBe(true);
    });

    it('enables send button when input has text', async () => {
        const { getByPlaceholderText, getByText } = render(CommandInput);
        const input = getByPlaceholderText('Enter command...') as HTMLInputElement;
        const sendButton = getByText('Send') as HTMLButtonElement;

        await fireEvent.input(input, { target: { value: 'test' } });

        expect(sendButton.disabled).toBe(false);
    });


    it('maintains command history limit of 50', async () => {
        const { getByPlaceholderText } = render(CommandInput);
        const input = getByPlaceholderText('Enter command...') as HTMLInputElement;

        // Send 56 commands (0-55)
        for (let i = 0; i < 56; i++) {
            await fireEvent.input(input, { target: { value: `command${i}` } });
            await fireEvent.keyDown(input, { key: 'Enter' });
        }

        // History keeps last 50, navigate up 49 times to reach oldest
        for (let i = 0; i < 49; i++) {
            await fireEvent.keyDown(input, { key: 'ArrowUp' });
        }

        // Verify history is capped at 50 entries (oldest should be 55-49=6, but is command7)
        expect(input.value).toBe('command7');
    });
});
