import { render, fireEvent } from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';
import WorldCreationModal from './WorldCreationModal.svelte';

describe('WorldCreationModal', () => {
    it('renders open with title and form', () => {
        const { getByText, getByLabelText } = render(WorldCreationModal, { isOpen: true });

        expect(getByText('Genesis Protocol')).toBeDefined();
        // Check for sections
        expect(getByText('Physical Parameters')).toBeDefined();
        expect(getByText('Active Systems')).toBeDefined();
    });

    it('generates random name on open', async () => {
        const { getByLabelText } = render(WorldCreationModal, { isOpen: true });

        const nameInput = getByLabelText('World Designation') as HTMLInputElement;
        expect(nameInput.value).toBeTruthy();
        expect(nameInput.value.length).toBeGreaterThan(0);
    });

    it('updates parameters and submits', async () => {
        const { getByText, component } = render(WorldCreationModal, { isOpen: true });

        const completeSpy = vi.fn();
        component.$on('complete', completeSpy);

        // Click a size button (medium)
        await fireEvent.click(getByText('medium'));

        // Submit
        await fireEvent.click(getByText('Initialize Simulation'));

        expect(completeSpy).toHaveBeenCalled();
        const eventDetail = completeSpy.mock.calls[0][0].detail;

        // Check payload structure
        expect(eventDetail.name).toBeTruthy();
        expect(eventDetail.size).toBe('medium');
        expect(eventDetail.moonCount).toBeDefined();
        expect(eventDetail.sysGeology).toBe(true);
    });

    it('enforces system dependencies', async () => {
        const { getByLabelText, component } = render(WorldCreationModal, { isOpen: true });

        // Uncheck Geology
        const geologyToggle = getByLabelText('Geology') as HTMLInputElement;
        await fireEvent.click(geologyToggle);
        expect(geologyToggle.checked).toBe(false);

        // Check that dependent systems are unchecked in DOM
        const lifeToggle = getByLabelText('Life & Evolution') as HTMLInputElement;
        expect(lifeToggle.checked).toBe(false);

        const weatherToggle = getByLabelText('Weather & Climate') as HTMLInputElement;
        expect(weatherToggle.checked).toBe(false);
    });
});
