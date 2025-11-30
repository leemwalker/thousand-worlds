import type { GameMessage, TextSegment } from '$lib/types/ui';

export class OutputFormatter {
    formatGameOutput(message: GameMessage): TextSegment[] {
        const segments: TextSegment[] = [];

        switch (message.type) {
            case 'movement':
                segments.push(
                    { text: 'You move ', color: 'text-gray-300' },
                    { text: message.direction || 'unknown', color: 'text-blue-400', bold: true },
                    { text: '.', color: 'text-gray-300' }
                );
                break;

            case 'area_description':
                segments.push(
                    { text: message.text, color: 'text-gray-100' }
                );
                break;

            case 'combat':
                segments.push(
                    { text: 'You attack ', color: 'text-red-400' },
                    {
                        text: message.targetName || 'target',
                        color: 'text-yellow-400',
                        bold: true,
                        entityID: message.targetID,
                        entityType: 'npc'
                    },
                    { text: ` for `, color: 'text-red-400' },
                    { text: (message.damage || 0).toString(), color: 'text-orange-500', bold: true },
                    { text: ' damage!', color: 'text-red-400' }
                );
                break;

            case 'dialogue':
                segments.push(
                    {
                        text: message.speakerName || 'Someone',
                        color: 'text-cyan-400',
                        bold: true,
                        entityID: message.speakerID,
                        entityType: 'npc'
                    },
                    { text: ' says: "', color: 'text-gray-300' },
                    { text: message.text, color: 'text-green-300', italic: true },
                    { text: '"', color: 'text-gray-300' }
                );
                break;

            case 'item_acquired':
                segments.push(
                    { text: 'You obtained ', color: 'text-gray-300' },
                    {
                        text: message.itemName || 'item',
                        color: this.getItemRarityColor(message.itemRarity || 'common'),
                        bold: true,
                        entityID: message.itemID,
                        entityType: 'item'
                    },
                    { text: ` (×${message.quantity || 1})`, color: 'text-gray-400' }
                );
                break;

            case 'crafting_success':
                segments.push(
                    { text: 'You successfully crafted ', color: 'text-green-400' },
                    {
                        text: message.itemName || 'item',
                        color: this.getQualityColor(message.quality || 'common'),
                        bold: true
                    },
                    { text: '!', color: 'text-green-400' }
                );
                break;

            case 'error':
                segments.push(
                    { text: message.text, color: 'text-red-500', bold: true }
                );
                break;

            case 'system':
                segments.push(
                    { text: '[System] ', color: 'text-purple-400', bold: true },
                    { text: message.text, color: 'text-gray-300' }
                );
                break;

            default:
                segments.push({ text: message.text, color: 'text-gray-300' });
        }

        return segments;
    }

    private getItemRarityColor(rarity: string): string {
        const rarityColors: Record<string, string> = {
            'common': 'text-gray-100',
            'uncommon': 'text-green-400',
            'rare': 'text-blue-400',
            'very_rare': 'text-purple-400',
            'legendary': 'text-orange-500',
        };
        return rarityColors[rarity] || 'text-gray-100';
    }

    private getQualityColor(quality: string): string {
        const qualityColors: Record<string, string> = {
            'poor': 'text-gray-400',
            'common': 'text-gray-100',
            'good': 'text-green-400',
            'excellent': 'text-blue-400',
            'masterwork': 'text-purple-500',
        };
        return qualityColors[quality] || 'text-gray-100';
    }
}
