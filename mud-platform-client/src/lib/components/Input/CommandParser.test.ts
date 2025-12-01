import { describe, it, expect, beforeEach } from 'vitest';
import { CommandParser } from './CommandParser';

describe('CommandParser', () => {
    let parser: CommandParser;

    beforeEach(() => {
        parser = new CommandParser();
    });

    describe('Exact Matching', () => {
        it('should parse cardinal directions', () => {
            const directions = ['north', 'n', 'northeast', 'ne', 'east', 'e', 'southeast', 'se',
                'south', 's', 'southwest', 'sw', 'west', 'w', 'northwest', 'nw',
                'up', 'u', 'down', 'd'];

            directions.forEach(dir => {
                const result = parser.parse(dir);
                expect(result.action).toBeDefined();
                expect(result.confidence).toBe(1.0);
            });
        });

        it('should parse open command', () => {
            const result = parser.parse('open door');
            expect(result).toEqual({
                action: 'open',
                target: 'door',
                raw: 'open door',
                confidence: 1.0
            });
        });

        it('should parse enter command', () => {
            const result = parser.parse('enter portal');
            expect(result).toEqual({
                action: 'enter',
                target: 'portal',
                raw: 'enter portal',
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

        it('should parse who command', () => {
            const result = parser.parse('who');
            expect(result).toEqual({
                action: 'who',
                raw: 'who',
                confidence: 1.0
            });
        });
    });

    describe('Communication', () => {
        it('should parse say command', () => {
            const result = parser.parse('say hello world');
            expect(result).toEqual({
                action: 'say',
                message: 'hello world',
                raw: 'say hello world',
                confidence: 1.0
            });
        });

        it('should parse quoted speech', () => {
            const result = parser.parse('"hello world');
            expect(result).toEqual({
                action: 'say',
                message: 'hello world',
                raw: '"hello world',
                confidence: 1.0
            });
        });

        it('should parse whisper command', () => {
            const result = parser.parse('whisper bob secret message');
            expect(result).toEqual({
                action: 'whisper',
                recipient: 'bob',
                message: 'secret message',
                raw: 'whisper bob secret message',
                confidence: 1.0
            });
        });

        it('should parse tell command', () => {
            const result = parser.parse('tell alice are you there?');
            expect(result).toEqual({
                action: 'tell',
                recipient: 'alice',
                message: 'are you there?',
                raw: 'tell alice are you there?',
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
                confidence: 0.85 // Slightly lower confidence for NLP
            });
        });

        it('should parse "go north" as north command', () => {
            const result = parser.parse('go north');
            expect(result).toEqual({
                action: 'north',
                raw: 'go north',
                confidence: 0.9
            });
        });
    });

    describe('Fuzzy Matching', () => {
        it('should fix typos in commands', () => {
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
