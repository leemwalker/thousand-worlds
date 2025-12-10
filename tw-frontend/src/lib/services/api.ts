import { AuthError, AUTH_ERRORS } from '$lib/types/errors';

const API_URL = '/api';

export interface User {
    user_id: string;
    email: string;
    created_at: string;
    last_login?: string;
    last_world_id?: string;
}

export interface LoginResponse {
    token: string;
    user: User;
}

export class GameAPI {
    // Note: Token is now stored in HttpOnly cookie by the server
    // No client-side token storage needed - cookies are sent automatically

    async register(email: string, username: string, password: string): Promise<void> {
        try {
            const response = await fetch(`${API_URL}/auth/register`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                credentials: 'include', // Important: Send cookies with request
                body: JSON.stringify({ email, username, password }),
            });

            if (!response.ok) {
                const error = await response.json();

                if (response.status === 409) { // Conflict
                    throw new AuthError('Email exists', AUTH_ERRORS.EMAIL_EXISTS);
                } else if (response.status >= 500) {
                    throw new AuthError('Server error', AUTH_ERRORS.SERVER_ERROR);
                } else {
                    // Handle structured error response: { error: { code, message } }
                    const errorMessage = error.error?.message || error.error || 'Registration failed';
                    throw new AuthError(typeof errorMessage === 'string' ? errorMessage : 'Registration failed', {
                        title: 'Registration Failed',
                        message: typeof errorMessage === 'string' ? errorMessage : 'An unexpected error occurred.'
                    });
                }
            }
        } catch (err) {
            if (err instanceof AuthError) throw err;
            throw new AuthError('Network error', AUTH_ERRORS.NETWORK_ERROR, err);
        }
    }

    async login(email: string, password: string): Promise<LoginResponse> {
        try {
            const response = await fetch(`${API_URL}/auth/login`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                credentials: 'include', // Important: Allow cookies
                body: JSON.stringify({ email, password }),
            });

            if (!response.ok) {
                const error = await response.json();

                if (response.status === 401) {
                    throw new AuthError('Invalid credentials', AUTH_ERRORS.INVALID_CREDENTIALS);
                } else if (response.status >= 500) {
                    throw new AuthError('Server error', AUTH_ERRORS.SERVER_ERROR);
                } else {
                    // Handle structured error response: { error: { code, message } }
                    const errorMessage = error.error?.message || error.error || 'Login failed';
                    throw new AuthError(typeof errorMessage === 'string' ? errorMessage : 'Login failed', {
                        title: 'Login Failed',
                        message: typeof errorMessage === 'string' ? errorMessage : 'An unexpected error occurred.'
                    });
                }
            }

            const data: LoginResponse = await response.json();
            // Token is now in HttpOnly cookie - no need to store it
            return data;
        } catch (err) {
            if (err instanceof AuthError) throw err;
            throw new AuthError('Network error', AUTH_ERRORS.NETWORK_ERROR, err);
        }
    }

    async getMe(): Promise<User> {
        const response = await fetch(`${API_URL}/auth/me`, {
            credentials: 'include', // Send cookies for authentication
        });

        if (!response.ok) {
            throw new Error('Failed to get user info');
        }

        return response.json();
    }

    logout(): void {
        // Call logout endpoint to clear server-side cookie
        fetch(`${API_URL}/auth/logout`, {
            method: 'POST',
            credentials: 'include',
        }).catch(err => console.error('Logout error:', err));
    }
    async getCharacters(): Promise<{ characters: any[] }> {
        const response = await fetch(`${API_URL}/game/characters`, {
            credentials: 'include',
        });

        if (!response.ok) {
            throw new Error('Failed to fetch characters');
        }
        return response.json();
    }

    async getSkills(characterId: string): Promise<{ skills: Record<string, any> }> {
        // Use relative URL for client-side fetch (proxied by Vite)
        const response = await fetch(`${API_URL}/game/skills?character_id=${characterId}`, {
            credentials: 'include',
        });

        if (!response.ok) {
            // Fallback for TDD until backend exists: return mock if 404
            if (response.status === 404) {
                console.warn("Skills endpoint not found, returning empty skills");
                // We could throw, but for now let's throw so the test mock handles it
                // Actually, if we mock the route in Playwright, response DO come back ok.
            }
            throw new Error('Failed to fetch skills');
        }
        return response.json();
    }
}

// Singleton instance
export const gameAPI = new GameAPI();
