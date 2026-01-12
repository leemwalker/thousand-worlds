import { render, fireEvent } from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';
import WorldCreationModal from './WorldCreationModal.svelte';

describe('WorldCreationModal', () => {
    it('renders open with title', () => {
        const { getByText } = render(WorldCreationModal, { isOpen: true });
        expect(getByText('Genesis Protocol')).toBeDefined();
    });

    it('advances steps', async () => {
        const { getByText, queryByText } = render(WorldCreationModal, { isOpen: true });

        // Step 0
        expect(getByText('Initiating planetary formation sequence...')).toBeDefined();
        const initBtn = getByText('Initialize Core');

        // Advance
        await fireEvent.click(initBtn);

        // Step 1
        expect(getByText('Constructing terrain and atmosphere...')).toBeDefined();
        expect(queryByText('Initiating planetary formation sequence...')).toBeNull();
    });

    it('dispatches complete event', async () => {
        const { getByText, component } = render(WorldCreationModal, { isOpen: true });

        const completeSpy = vi.fn();
        component.$on('complete', completeSpy);

        // Step 0 -> Step 1
        await fireEvent.click(getByText('Initialize Core'));

        // Complete
        await fireEvent.click(getByText('Finalize Biosphere'));
        expect(completeSpy).toHaveBeenCalled();
    });
});
