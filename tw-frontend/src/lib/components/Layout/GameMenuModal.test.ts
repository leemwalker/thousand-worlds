import { render, fireEvent } from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';
import GameMenuModal from './GameMenuModal.svelte';

describe('GameMenuModal', () => {
    it('renders closed by default', () => {
        const { queryByRole } = render(GameMenuModal, { isOpen: false });
        expect(queryByRole('dialog')).toBeNull();
    });

    it('renders open with header title', () => {
        const { getByText, getByRole } = render(GameMenuModal, { isOpen: true });
        expect(getByRole('dialog')).toBeDefined();
        expect(getByText('Menu')).toBeDefined();
    });

    it('dispatches close event on close button click', async () => {
        const { getByLabelText, component } = render(GameMenuModal, { isOpen: true });
        const closeBtn = getByLabelText('Close menu');

        const closeSpy = vi.fn();
        component.$on('close', closeSpy);

        await fireEvent.click(closeBtn);
        expect(closeSpy).toHaveBeenCalled();
    });

    it('shows tabs and switches content', async () => {
        const { getByText, queryByText } = render(GameMenuModal, { isOpen: true });

        // Default is World tab
        expect(getByText('Reset World')).toBeDefined();

        // Switch to Character
        await fireEvent.click(getByText('Character'));
        expect(getByText('Character options coming soon...')).toBeDefined();

        // Switch to Account
        await fireEvent.click(getByText('Account'));
        expect(getByText('Logout')).toBeDefined();
    });

    it('confirms world reset', async () => {
        const { getByText, component } = render(GameMenuModal, { isOpen: true });

        const resetSpy = vi.fn();
        component.$on('resetWorld', resetSpy);

        // Click Reset - shows confirmation
        await fireEvent.click(getByText('Reset World'));
        expect(getByText('Confirm Reset')).toBeDefined();

        // Confirm
        await fireEvent.click(getByText('Confirm Reset'));
        expect(resetSpy).toHaveBeenCalled();
    });
});
