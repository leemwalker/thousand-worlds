export type ActionType =
    | 'north' | 'northeast' | 'east' | 'southeast' | 'south' | 'southwest' | 'west' | 'northwest' | 'up' | 'down'
    | 'open' | 'enter' | 'look' | 'say' | 'whisper' | 'tell' | 'who'
    | 'take' | 'drop' | 'attack' | 'talk' | 'inventory' | 'craft' | 'use' | 'unknown';
export type Direction = 'N' | 'S' | 'E' | 'W' | 'NE' | 'NW' | 'SE' | 'SW' | 'UP' | 'DOWN';

export interface ParsedCommand {
    action: ActionType;
    target?: string;
    // direction?: Direction; // Deprecated
    quantity?: number;
    items?: string[];
    message?: string;
    recipient?: string;
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
            ['north', ['n']],
            ['northeast', ['ne']],
            ['east', ['e']],
            ['southeast', ['se']],
            ['south', ['s']],
            ['southwest', ['sw']],
            ['west', ['w']],
            ['northwest', ['nw']],
            ['up', ['u']],
            ['down', ['d', 'dn']],
            ['open', []],
            ['enter', ['go in', 'step through']],
            ['look', ['l', 'examine', 'inspect', 'view']],
            ['say', ['speak']],
            ['whisper', ['psst']],
            ['tell', ['message', 'msg', 'pm']],
            ['who', ['players', 'online']],
            ['take', ['get', 'grab', 'pick', 'pickup']],
            ['drop', ['release', 'discard', 'throw']],
            ['attack', ['hit', 'fight', 'strike', 'kill']],
            ['talk', ['chat']],
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
        const normalized = input.trim(); // Don't lowercase yet to preserve message case

        // Check for empty input
        if (!normalized) {
            return this.createError('Empty command');
        }

        // Check for quoted speech
        const quoted = this.parseQuotedSpeech(normalized);
        if (quoted) return quoted;

        // Lowercase for command matching
        const lowerInput = normalized.toLowerCase();

        // Try exact command match first
        const exactMatch = this.tryExactMatch(lowerInput, normalized);
        if (exactMatch) return exactMatch;

        // Try natural language parsing
        const nlpMatch = this.parseNaturalLanguage(lowerInput, normalized);
        if (nlpMatch.confidence > 0.6) return nlpMatch;

        // Try fuzzy matching for typos
        const fuzzyMatch = this.tryFuzzyMatch(lowerInput);
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

    private parseQuotedSpeech(input: string): ParsedCommand | null {
        if (input.startsWith('"') || input.startsWith("'")) {
            const message = input.substring(1).trim();
            return {
                action: 'say',
                message,
                raw: input,
                confidence: 1.0
            };
        }
        return null;
    }

    private tryExactMatch(lowerInput: string, originalInput: string): ParsedCommand | null {
        const words = lowerInput.split(' ');
        const command = words[0];
        const args = words.slice(1).join(' ');

        // Get original args to preserve case for messages
        const originalWords = originalInput.split(' ');
        const originalArgs = originalWords.slice(1).join(' ');

        for (const [action, aliases] of this.commandAliases) {
            if (action === command || aliases.includes(command)) {
                return this.parseActionArgs(action as ActionType, args, originalArgs, originalInput);
            }
        }

        return null;
    }

    private parseActionArgs(action: ActionType, args: string, originalArgs: string, raw: string): ParsedCommand {
        const result: ParsedCommand = { action, raw, confidence: 1.0 };

        switch (action) {
            case 'say':
                result.message = originalArgs;
                break;
            case 'whisper':
                // Format: whisper <recipient> <message>
                const whisperParts = originalArgs.split(' ');
                if (whisperParts.length >= 2) {
                    result.recipient = whisperParts[0];
                    result.message = whisperParts.slice(1).join(' ');
                }
                break;
            case 'tell':
                // Format: tell <recipient> <message>
                const tellParts = originalArgs.split(' ');
                if (tellParts.length >= 2) {
                    result.recipient = tellParts[0];
                    result.message = tellParts.slice(1).join(' ');
                }
                break;
            case 'open':
            case 'enter':
            case 'take':
            case 'craft':
            case 'drop':
            case 'use':
            case 'look':
                result.target = this.cleanArgs(args);
                break;
            case 'attack':
                result.target = this.cleanArgs(args) || this.contextMemory.lastTarget;
                break;
            case 'talk':
                result.target = this.cleanArgs(args) || this.contextMemory.lastNPC;
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

    private parseNaturalLanguage(input: string, originalInput: string): ParsedCommand {
        // Movement patterns: "go north", "walk east"
        if (/go|walk|move|head|travel/.test(input)) {
            for (const [action, aliases] of this.commandAliases) {
                if (['north', 'northeast', 'east', 'southeast', 'south', 'southwest', 'west', 'northwest', 'up', 'down'].includes(action)) {
                    // Check if direction is in input
                    if (input.includes(action) || aliases.some(a => input.includes(` ${a} `) || input.endsWith(` ${a}`))) {
                        return {
                            action: action as ActionType,
                            raw: originalInput,
                            confidence: 0.9,
                        };
                    }
                }
            }
        }

        // Taking items: "pick up the sword", "get iron ore"
        if (/pick up|take|get|grab/.test(input)) {
            const item = this.extractItemName(input);
            return {
                action: 'take',
                target: item,
                raw: originalInput,
                confidence: 0.85,
            };
        }

        // Dropping items
        if (/drop|throw away|discard/.test(input)) {
            const item = input.replace(/^(drop|throw away|discard)\s+/i, '').replace(/^(the|a|an)\s+/i, '').trim();
            return {
                action: 'drop',
                target: item,
                raw: originalInput,
                confidence: 0.85,
            };
        }

        // Talking
        if (/talk|speak|chat/.test(input)) {
            const target = this.extractTarget(input);
            return {
                action: 'talk',
                target: target || this.contextMemory.lastNPC,
                raw: originalInput,
                confidence: target ? 0.85 : 0.6,
            };
        }

        // Looking: "look at fountain", "examine the door"
        if (/look|examine|inspect/.test(input)) {
            const target = this.extractTarget(input);
            return {
                action: 'look',
                target,
                raw: originalInput,
                confidence: 0.8,
            };
        }

        return { action: 'unknown', raw: originalInput, confidence: 0.0 };
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
