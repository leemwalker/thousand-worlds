import type { GameMessage, TextSegment } from "$lib/types/ui";

export class OutputFormatter {
    formatGameOutput(message: GameMessage): TextSegment[] {
        const segments: TextSegment[] = [];

        // Base style based on message type
        let baseColor = 'text-gray-300';
        let isBold = false;
        let isItalic = false;

        switch (message.type) {
            case 'error':
                baseColor = 'text-red-400';
                isBold = true;
                break;
            case 'system':
                baseColor = 'text-yellow-400';
                break;
            case 'combat':
                baseColor = 'text-red-300';
                break;
            case 'area_description':
                baseColor = 'text-purple-300';
                break;
            case 'item_acquired':
                baseColor = 'text-green-300';
                break;
            case 'movement':
                baseColor = 'text-blue-300';
                break;
            case 'dialogue':
                baseColor = 'text-cyan-300';
                isItalic = true;
                break;
            case 'crafting_success':
                baseColor = 'text-green-400';
                break;
            default:
                baseColor = 'text-gray-300';
        }

        segments.push({
            text: message.text,
            color: baseColor,
            bold: isBold,
            italic: isItalic
        });

        return segments;
    }
}
