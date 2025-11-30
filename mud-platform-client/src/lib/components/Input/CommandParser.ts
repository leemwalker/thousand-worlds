export type ActionType = 'move' | 'look' | 'take' | 'drop' | 'attack' | 'talk' | 'inventory' | 'craft' | 'use' | 'unknown';
export type Direction = 'N' | 'S' | 'E' | 'W' | 'NE' | 'NW' | 'SE' | 'SW' | 'UP' | 'DOWN';

export interface ParsedCommand {
    action: ActionType;
    target?: string;
    direction?: Direction;
    quantity?: number;
    items?: string[];
    raw: string;
    confidence: number; // 0.0 to 1.0
}

export interface ParserContext {
    lastTarget?: string;
    lastNPC?: string;
    lastItem?: string;
}

export class CommandParser {
    private commandAliases: Map<string, string[]>;
    private contextMemory: ParserContext;

    constructor(context: ParserContext = {}) {
        this.contextMemory = context;
        this.commandAliases = new Map([
            ['look', ['l', 'examine', 'inspect', 'view']],
            ['move', ['go', 'walk', 'run', 'travel', 'head']],
            ['take', ['get', 'grab', 'pick', 'pickup']],
            ['drop', ['release', 'discard', 'throw']],
            ['attack', ['hit', 'fight', 'strike', 'kill']],
            ['talk', ['speak', 'chat', 'say', 'tell']],
            ['inventory', ['inv', 'i', 'items', 'bag']],
            ['craft', ['make', 'create', 'build', 'forge']],
            ['use', ['consume', 'activate', 'apply']],
        ]);
    }

    updateContext(context: Partial<ParserContext>) {
        this.contextMemory = { ...this.contextMemory, ...context };
    }

    parse(input: string): ParsedCommand {
        // Normalize input
        const normalized = input.toLowerCase().trim();

        // Check for empty input
        if (!normalized) {
            return this.createError('Empty command');
        }

        // Try exact command match first
        const exactMatch = this.tryExactMatch(normalized);
        if (exactMatch) return exactMatch;

        // Try natural language parsing
        const nlpMatch = this.parseNaturalLanguage(normalized);
        if (nlpMatch.confidence > 0.6) return nlpMatch;

        // Try fuzzy matching for typos
        const fuzzyMatch = this.tryFuzzyMatch(normalized);
        if (fuzzyMatch.confidence > 0.5) return fuzzyMatch;

        // Unknown command
        return this.createError(`Unknown command: "${input}". Type "help" for commands.`);
    }

    private createError(message: string): ParsedCommand {
        return {
            action: 'unknown',
            raw: message,
            confidence: 0.0
        };
    }

    private tryExactMatch(input: string): ParsedCommand | null {
        const words = input.split(' ');
        const command = words[0];
        const args = words.slice(1).join(' ');

        for (const [action, aliases] of this.commandAliases) {
            if (action === command || aliases.includes(command)) {
                // Special handling for move directions as single letters
                if (action === 'move' && !args) {
                    // If just "n", "s", etc.
                    const dir = this.extractDirection(command);
                    if (dir) {
                        return { action: 'move', direction: dir, raw: input, confidence: 1.0 };
                    }
                }

                return this.parseActionArgs(action as ActionType, args, input);
            }
        }

        // Check if input is just a direction
        const dir = this.extractDirection(command);
        if (dir && !args) {
            return { action: 'move', direction: dir, raw: input, confidence: 1.0 };
        }

        return null;
    }

    private parseActionArgs(action: ActionType, args: string, raw: string): ParsedCommand {
        const result: ParsedCommand = { action, raw, confidence: 1.0 };

        switch (action) {
            case 'move':
                result.direction = this.extractDirection(args);
                break;
            case 'take':
            case 'craft':
                // Use extractItemName to clean "the", "a", "an"
                // For "pick up", the command "pick" is consumed, args is "up the sword"
                // extractItemName removes "pick up" etc, but here we just have args.
                // We need a helper that just cleans articles and prepositions if they are at the start.
                result.target = this.cleanArgs(args);
                break;
            case 'drop':
            case 'use':
                result.target = this.cleanArgs(args);
                break;
            case 'attack':
                result.target = this.cleanArgs(args) || this.contextMemory.lastTarget;
                break;
            case 'talk':
                result.target = this.cleanArgs(args) || this.contextMemory.lastNPC;
                break;
            case 'look':
                result.target = this.cleanArgs(args);
                break;
        }

        return result;
    }

    private cleanArgs(args: string): string | undefined {
        if (!args) return undefined;

        // Remove common prepositions/articles that might be left in args
        // e.g. "at sword", "to merchant", "up the sword"
        let cleaned = args
            .replace(/^(at|to|with|up|down|the|a|an)\s+/i, '')
            .replace(/^(the|a|an)\s+/i, ''); // Run again for "up the" -> "the" -> ""

        return cleaned.trim() || undefined;
    }

    private parseNaturalLanguage(input: string): ParsedCommand {
        // Movement patterns
        if (/go|walk|move|head|travel/.test(input)) {
            const direction = this.extractDirection(input);
            if (direction) {
                return {
                    action: 'move',
                    direction,
                    raw: input,
                    confidence: 0.9,
                };
            }
        }

        // Taking items: "pick up the sword", "get iron ore"
        if (/pick up|take|get|grab/.test(input)) {
            const item = this.extractItemName(input);
            return {
                action: 'take',
                target: item,
                raw: input,
                confidence: item ? 0.85 : 0.5,
            };
        }

        // Crafting: "make iron sword", "craft health potion"
        if (/make|craft|create|forge/.test(input)) {
            const item = this.extractItemName(input);
            return {
                action: 'craft',
                target: item,
                raw: input,
                confidence: item ? 0.85 : 0.5,
            };
        }

        // Combat: "attack goblin", "fight the orc"
        if (/attack|fight|hit|kill/.test(input)) {
            const target = this.extractTarget(input);
            return {
                action: 'attack',
                target: target || this.contextMemory.lastTarget,
                raw: input,
                confidence: target ? 0.85 : 0.6,
            };
        }

        // Talking: "talk to merchant", "speak with guard"
        if (/talk|speak|chat/.test(input)) {
            const target = this.extractTarget(input);
            return {
                action: 'talk',
                target: target || this.contextMemory.lastNPC,
                raw: input,
                confidence: target ? 0.85 : 0.6,
            };
        }

        // Looking: "look at fountain", "examine the door"
        if (/look|examine|inspect/.test(input)) {
            const target = this.extractTarget(input);
            return {
                action: 'look',
                target,
                raw: input,
                confidence: 0.8,
            };
        }

        return { action: 'unknown', raw: input, confidence: 0.0 };
    }

    private extractDirection(input: string): Direction | undefined {
        const directionMap: Record<string, Direction> = {
            'north': 'N', 'n': 'N', 'up': 'UP', 'u': 'UP',
            'south': 'S', 's': 'S', 'down': 'DOWN', 'd': 'DOWN',
            'east': 'E', 'e': 'E',
            'west': 'W', 'w': 'W',
            'northeast': 'NE', 'ne': 'NE',
            'northwest': 'NW', 'nw': 'NW',
            'southeast': 'SE', 'se': 'SE',
            'southwest': 'SW', 'sw': 'SW',
        };

        // Check for exact matches or "go [direction]"
        for (const [key, direction] of Object.entries(directionMap)) {
            // Word boundary check to avoid matching "news" as "n" or "ne"
            const regex = new RegExp(`\\b${key}\\b`, 'i');
            if (regex.test(input)) {
                return direction;
            }
        }

        return undefined;
    }

    private extractItemName(input: string): string {
        // Remove command words
        let cleaned = input
            .replace(/^(pick up|take|get|grab|make|craft|create|forge)\s+/i, '')
            .replace(/^(the|a|an)\s+/i, '');

        return cleaned.trim();
    }

    private extractTarget(input: string): string | undefined {
        // Remove command words and articles
        let cleaned = input
            .replace(/^(talk|speak|chat|attack|fight|hit|look|examine)\s+/i, '')
            .replace(/^(to|with|at|the|a|an)\s+/i, '');

        return cleaned.trim() || undefined;
    }

    private tryFuzzyMatch(input: string): ParsedCommand {
        const words = input.split(' ');
        const firstWord = words[0];

        let bestMatch: string | null = null;
        let bestDistance = Infinity;

        for (const [command, aliases] of this.commandAliases) {
            const allVariants = [command, ...aliases];

            for (const variant of allVariants) {
                const distance = this.levenshteinDistance(firstWord, variant);

                if (distance < bestDistance && distance <= 2) {
                    bestDistance = distance;
                    bestMatch = command;
                }
            }
        }

        if (bestMatch) {
            // Reconstruct command with corrected first word
            const correctedInput = [bestMatch, ...words.slice(1)].join(' ');
            return this.parse(correctedInput);
        }

        return { action: 'unknown', raw: input, confidence: 0.0 };
    }

    private levenshteinDistance(str1: string, str2: string): number {
        const matrix: number[][] = [];

        for (let i = 0; i <= str2.length; i++) {
            matrix[i] = [i];
        }

        for (let j = 0; j <= str1.length; j++) {
            matrix[0][j] = j;
        }

        for (let i = 1; i <= str2.length; i++) {
            for (let j = 1; j <= str1.length; j++) {
                if (str2.charAt(i - 1) === str1.charAt(j - 1)) {
                    matrix[i][j] = matrix[i - 1][j - 1];
                } else {
                    matrix[i][j] = Math.min(
                        matrix[i - 1][j - 1] + 1,
                        matrix[i][j - 1] + 1,
                        matrix[i - 1][j] + 1
                    );
                }
            }
        }

        return matrix[str2.length][str1.length];
    }
}
