import { describe, it, expect } from 'vitest';
import { OutputFormatter } from './OutputFormatter';
import type { GameMessage } from '$lib/types/ui';

describe('OutputFormatter', () => {
    const formatter = new OutputFormatter();
    const baseMessage: GameMessage = {
        id: '1',
        type: 'system',
        text: 'test',
        timestamp: new Date()
    };

    it('should format movement messages', () => {
        const msg: GameMessage = { ...baseMessage, type: 'movement', direction: 'North' };
        const segments = formatter.formatGameOutput(msg);

        expect(segments).toHaveLength(3);
        expect(segments[1].text).toBe('North');
        expect(segments[1].color).toBe('text-blue-400');
    });

    it('should format combat messages', () => {
        const msg: GameMessage = {
            ...baseMessage,
            type: 'combat',
            targetName: 'Goblin',
            damage: 10
        };
        const segments = formatter.formatGameOutput(msg);

        expect(segments[1].text).toBe('Goblin');
        expect(segments[1].color).toBe('text-yellow-400');
        expect(segments[3].text).toBe('10');
        expect(segments[3].color).toBe('text-orange-500');
    });

    it('should format item acquisition', () => {
        const msg: GameMessage = {
            ...baseMessage,
            type: 'item_acquired',
            itemName: 'Legendary Sword',
            itemRarity: 'legendary',
            quantity: 1
        };
        const segments = formatter.formatGameOutput(msg);

        expect(segments[1].text).toBe('Legendary Sword');
        expect(segments[1].color).toBe('text-orange-500'); // Legendary color
    });
});
