import { describe, it, expect, beforeEach } from 'vitest';
import { CommandParser } from './CommandParser';

describe('CommandParser', () => {
    let parser: CommandParser;

    beforeEach(() => {
        parser = new CommandParser();
    });

    describe('Exact Matching', () => {
        it('should parse simple movement commands', () => {
            const result = parser.parse('move north');
            expect(result).toEqual({
                action: 'move',
                direction: 'N',
                raw: 'move north',
                confidence: 1.0
            });
        });

        it('should parse shorthand movement', () => {
            const result = parser.parse('n');
            expect(result).toEqual({
                action: 'move',
                direction: 'N',
                raw: 'n',
                confidence: 1.0
            });
        });

        it('should parse look command', () => {
            const result = parser.parse('look');
            expect(result).toEqual({
                action: 'look',
                raw: 'look',
                confidence: 1.0
            });
        });

        it('should parse look at target', () => {
            const result = parser.parse('look at sword');
            expect(result).toEqual({
                action: 'look',
                target: 'sword',
                raw: 'look at sword',
                confidence: 1.0
            });
        });
    });

    describe('Natural Language Parsing', () => {
        it('should parse "pick up the sword"', () => {
            const result = parser.parse('pick up the sword');
            expect(result).toEqual({
                action: 'take',
                target: 'sword',
                raw: 'pick up the sword',
                confidence: 1.0
            });
        });

        it('should parse "talk to merchant"', () => {
            const result = parser.parse('talk to merchant');
            expect(result).toEqual({
                action: 'talk',
                target: 'merchant',
                raw: 'talk to merchant',
                confidence: 1.0
            });
        });

        it('should parse "go north"', () => {
            const result = parser.parse('go north');
            expect(result).toEqual({
                action: 'move',
                direction: 'N',
                raw: 'go north',
                confidence: 1.0
            });
        });
    });

    describe('Fuzzy Matching', () => {
        it('should fix typos in commands', () => {
            const result = parser.parse('mvoe north');
            expect(result.action).toBe('move');
            expect(result.direction).toBe('N');
        });

        it('should fix typos in look', () => {
            const result = parser.parse('loko');
            expect(result.action).toBe('look');
        });
    });

    describe('Context Awareness', () => {
        it('should remember last target for attack', () => {
            parser.updateContext({ lastTarget: 'goblin' });
            const result = parser.parse('attack');
            expect(result).toEqual({
                action: 'attack',
                target: 'goblin',
                raw: 'attack',
                confidence: 1.0
            });
        });
    });
});
