import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { GameAPI } from './api';
import { AuthError, AUTH_ERRORS } from '$lib/types/errors';

// Reset mocks between tests
describe('GameAPI', () => {
    let api: GameAPI;
    const fetchMock = vi.fn();

    beforeEach(() => {
        global.fetch = fetchMock;
        api = new GameAPI(); // Singleton-ish, but new instance for clean testing
        vi.unstubAllGlobals();
        vi.stubGlobal('fetch', fetchMock);
    });

    afterEach(() => {
        vi.clearAllMocks();
    });

    it('should register successfully', async () => {
        fetchMock.mockResolvedValueOnce({
            ok: true,
            status: 200,
            json: async () => ({})
        });

        await expect(api.register('test@example.com', 'user', 'pass')).resolves.not.toThrow();

        expect(fetchMock).toHaveBeenCalledWith(
            expect.stringContaining('/auth/register'),
            expect.objectContaining({
                method: 'POST',
                body: JSON.stringify({ email: 'test@example.com', username: 'user', password: 'pass' })
            })
        );
    });

    it('should handle login success', async () => {
        const mockResponse = {
            token: 'jwt-token',
            user: { user_id: '123', email: 'test@example.com' }
        };

        fetchMock.mockResolvedValueOnce({
            ok: true,
            status: 200,
            json: async () => mockResponse
        });

        const result = await api.login('test@example.com', 'pass');
        expect(result).toEqual(mockResponse);
    });

    it('should throw AuthError on 401', async () => {
        fetchMock.mockResolvedValueOnce({
            ok: false,
            status: 401,
            json: async () => ({ error: 'Unauthorized' })
        });

        await expect(api.login('test@example.com', 'wrongpass'))
            .rejects.toThrow('Invalid credentials');
    });

    it('should throw specific AuthError on 409 limit', async () => {
        fetchMock.mockResolvedValueOnce({
            ok: false,
            status: 409,
            json: async () => ({})
        });

        await expect(api.register('exists@example.com', 'u', 'p'))
            .rejects.toThrow('Conflict');
    });

    it('should handle generic server errors', async () => {
        fetchMock.mockResolvedValueOnce({
            ok: false,
            status: 500,
            json: async () => ({})
        });

        await expect(api.getMe())
            .rejects.toThrow('Server error');
    });

    it('should handle network errors', async () => {
        fetchMock.mockRejectedValueOnce(new Error('Network failure'));

        await expect(api.getMe())
            .rejects.toThrow('Network error');
    });

    it('should handle structured error responses', async () => {
        fetchMock.mockResolvedValueOnce({
            ok: false,
            status: 400,
            json: async () => ({
                error: { message: 'Custom validation error' }
            })
        });

        await expect(api.login('u', 'p'))
            .rejects.toThrow('Custom validation error');
    });
});
