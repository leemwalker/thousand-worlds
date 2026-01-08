/**
 * GameSystem Unit Tests
 * Tests for command parsing and routing logic.
 */
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { parseMovementCommand } from '../GameSystem';

describe('parseMovementCommand', () => {
    describe('single letter aliases', () => {
        it('should parse "n" as "north"', () => {
            expect(parseMovementCommand('n')).toBe('north');
        });

        it('should parse "s" as "south"', () => {
            expect(parseMovementCommand('s')).toBe('south');
        });

        it('should parse "e" as "east"', () => {
            expect(parseMovementCommand('e')).toBe('east');
        });

        it('should parse "w" as "west"', () => {
            expect(parseMovementCommand('w')).toBe('west');
        });

        it('should parse "u" as "up"', () => {
            expect(parseMovementCommand('u')).toBe('up');
        });

        it('should parse "d" as "down"', () => {
            expect(parseMovementCommand('d')).toBe('down');
        });
    });

    describe('diagonal aliases', () => {
        it('should parse "ne" as "northeast"', () => {
            expect(parseMovementCommand('ne')).toBe('northeast');
        });

        it('should parse "nw" as "northwest"', () => {
            expect(parseMovementCommand('nw')).toBe('northwest');
        });

        it('should parse "se" as "southeast"', () => {
            expect(parseMovementCommand('se')).toBe('southeast');
        });

        it('should parse "sw" as "southwest"', () => {
            expect(parseMovementCommand('sw')).toBe('southwest');
        });
    });

    describe('full direction names', () => {
        it('should parse "north" as "north"', () => {
            expect(parseMovementCommand('north')).toBe('north');
        });

        it('should parse "SOUTH" (case-insensitive) as "south"', () => {
            expect(parseMovementCommand('SOUTH')).toBe('south');
        });

        it('should parse "East" (mixed case) as "east"', () => {
            expect(parseMovementCommand('East')).toBe('east');
        });
    });

    describe('"go <direction>" format', () => {
        it('should parse "go north" as "north"', () => {
            expect(parseMovementCommand('go north')).toBe('north');
        });

        it('should parse "go n" (alias) as "north"', () => {
            expect(parseMovementCommand('go n')).toBe('north');
        });

        it('should parse "GO SOUTH" (uppercase) as "south"', () => {
            expect(parseMovementCommand('GO SOUTH')).toBe('south');
        });
    });

    describe('"move <direction>" format', () => {
        it('should parse "move north" as "north"', () => {
            expect(parseMovementCommand('move north')).toBe('north');
        });

        it('should parse "move w" (alias) as "west"', () => {
            expect(parseMovementCommand('move w')).toBe('west');
        });
    });

    describe('non-movement commands', () => {
        it('should return null for "look"', () => {
            expect(parseMovementCommand('look')).toBeNull();
        });

        it('should return null for "get sword"', () => {
            expect(parseMovementCommand('get sword')).toBeNull();
        });

        it('should return null for "say hello"', () => {
            expect(parseMovementCommand('say hello')).toBeNull();
        });

        it('should return null for empty string', () => {
            expect(parseMovementCommand('')).toBeNull();
        });

        it('should return null for whitespace only', () => {
            expect(parseMovementCommand('   ')).toBeNull();
        });
    });

    describe('edge cases', () => {
        it('should handle leading/trailing whitespace', () => {
            expect(parseMovementCommand('  north  ')).toBe('north');
        });

        it('should handle "go" with extra spaces', () => {
            expect(parseMovementCommand('go    north')).toBe('north');
        });

        it('should return null for "go" with no direction', () => {
            expect(parseMovementCommand('go')).toBeNull();
        });

        it('should return null for invalid direction after "go"', () => {
            expect(parseMovementCommand('go sideways')).toBeNull();
        });
    });
});
