import { describe, it, expect, vi, beforeEach } from 'vitest';
import { OnboardingService } from './OnboardingService';

// Mock dependencies
const mockApi = {
    getCharacters: vi.fn(),
};

const mockWs = {
    connect: vi.fn(),
    disconnect: vi.fn(),
};

// Mock fetch
global.fetch = vi.fn();

describe('OnboardingService', () => {
    let service: OnboardingService;

    beforeEach(() => {
        service = new OnboardingService(mockApi as any, mockWs as any);
        vi.clearAllMocks();
    });

    describe('checkStatus', () => {
        it('should return lobby step if no active interview', async () => {
            (global.fetch as any).mockResolvedValue({
                ok: false,
            });

            const state = await service.checkStatus(null);
            expect(state.step).toBe('lobby');
        });

        it('should return interview step if active interview exists', async () => {
            (global.fetch as any).mockResolvedValue({
                ok: true,
                json: () => Promise.resolve({
                    status: 'in_progress',
                    session_id: '123',
                    question: 'What is your world?',
                }),
            });

            const state = await service.checkStatus(null);
            expect(state.step).toBe('interview');
            expect(state.sessionId).toBe('123');
            expect(state.currentQuestion).toBe('What is your world?');
        });

        it('should return lobby step if interview is already complete', async () => {
            (global.fetch as any).mockResolvedValue({
                ok: true,
                json: () => Promise.resolve({
                    status: 'in_progress',
                    session_id: '123',
                    question: 'The interview is already complete.',
                }),
            });

            const state = await service.checkStatus(null);
            expect(state.step).toBe('lobby');
        });
    });

    describe('startInterview', () => {
        it('should start interview successfully', async () => {
            (global.fetch as any).mockResolvedValue({
                ok: true,
                json: () => Promise.resolve({
                    session_id: 'new-session',
                    question: 'Start?',
                }),
            });

            const state = await service.startInterview();
            expect(state.step).toBe('interview');
            expect(state.sessionId).toBe('new-session');
            expect(state.currentQuestion).toBe('Start?');
        });

        it('should handle errors', async () => {
            (global.fetch as any).mockResolvedValue({
                ok: false,
                status: 500
            });

            const state = await service.startInterview();
            expect(state.error).toContain('500');
        });
    });

    describe('sendResponse', () => {
        it('should return next question', async () => {
            (global.fetch as any).mockResolvedValue({
                ok: true,
                json: () => Promise.resolve({
                    completed: false,
                    next_question: 'Next?',
                }),
            });

            const result = await service.sendResponse('123', 'My Answer', []);
            expect(result.completed).toBe(false);
            expect(result.nextQuestion).toBe('Next?');
            expect(result.newHistory).toHaveLength(2); // User + Assistant
        });

        it('should detect completion', async () => {
            (global.fetch as any).mockResolvedValue({
                ok: true,
                json: () => Promise.resolve({
                    completed: true,
                }),
            });

            const result = await service.sendResponse('123', 'Done', []);
            expect(result.completed).toBe(true);
        });
    });
});
