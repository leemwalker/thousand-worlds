import { AuthError, AUTH_ERRORS } from '$lib/types/errors';
import type { User, Character, CharacterAttributes } from '$lib/types/game';

// Use environment variable for API URL (configured in vite.config.ts / .env)
const API_URL = import.meta.env.VITE_API_URL || '/api';

export interface LoginResponse {
    token: string;
    user: User;
}

export interface CharacterCreationData {
    name: string;
    attributes: CharacterAttributes;
    background: string;
    species: string;
}

export interface SkillsResponse {
    skills: Record<string, number>; // Using number for skill level/xp
}

export class GameAPI {
    /**
     * Generic fetch wrapper with standardized error handling
     */
    private async fetchWithErrorHandling<T>(url: string, options?: RequestInit): Promise<T> {
        try {
            const response = await fetch(url, {
                ...options,
                credentials: 'include', // Always send cookies
                headers: {
                    'Content-Type': 'application/json',
                    ...options?.headers,
                },
            });

            if (!response.ok) {
                // Handle non-200 responses
                let errorData;
                try {
                    errorData = await response.json();
                } catch {
                    // unexpected non-json error
                    throw new AuthError('Server error', AUTH_ERRORS.SERVER_ERROR);
                }

                // Handle specific status codes
                if (response.status === 401) {
                    throw new AuthError('Invalid credentials', AUTH_ERRORS.INVALID_CREDENTIALS);
                } else if (response.status === 409) {
                    throw new AuthError('Conflict', AUTH_ERRORS.EMAIL_EXISTS);
                } else if (response.status >= 500) {
                    throw new AuthError('Server error', AUTH_ERRORS.SERVER_ERROR);
                }

                // Handle structured error response: { error: { code, message } }
                const errorMessage = errorData.error?.message || errorData.error || 'Request failed';
                throw new AuthError(typeof errorMessage === 'string' ? errorMessage : 'Request failed', {
                    title: 'Request Failed',
                    message: typeof errorMessage === 'string' ? errorMessage : 'An unexpected error occurred.'
                });
            }

            // For 204 No Content, return null (cast as T)
            if (response.status === 204) {
                return null as T;
            }

            return response.json();
        } catch (err) {
            if (err instanceof AuthError) throw err;
            // Wrap unknown errors
            throw new AuthError('Network error', AUTH_ERRORS.NETWORK_ERROR, err);
        }
    }

    async register(email: string, username: string, password: string): Promise<void> {
        return this.fetchWithErrorHandling<void>(`${API_URL}/auth/register`, {
            method: 'POST',
            body: JSON.stringify({ email, username, password }),
        });
    }

    async login(email: string, password: string): Promise<LoginResponse> {
        return this.fetchWithErrorHandling<LoginResponse>(`${API_URL}/auth/login`, {
            method: 'POST',
            body: JSON.stringify({ email, password }),
        });
    }

    async getMe(): Promise<User> {
        return this.fetchWithErrorHandling<User>(`${API_URL}/auth/me`);
    }

    logout(): void {
        // Fire and forget logout, but log error if it fails
        this.fetchWithErrorHandling<void>(`${API_URL}/auth/logout`, {
            method: 'POST',
        }).catch(err => console.error('Logout error:', err));
    }

    async getCharacters(): Promise<{ characters: Character[] }> {
        return this.fetchWithErrorHandling<{ characters: Character[] }>(`${API_URL}/game/characters`);
    }

    async createCharacter(data: CharacterCreationData): Promise<Character> {
        return this.fetchWithErrorHandling<Character>(`${API_URL}/game/characters`, {
            method: 'POST',
            body: JSON.stringify(data)
        });
    }

    async getSkills(characterId: string): Promise<SkillsResponse> {
        try {
            return await this.fetchWithErrorHandling<SkillsResponse>(
                `${API_URL}/game/skills?character_id=${characterId}`
            );
        } catch (err) {
            // Fallback logic preserved from original file
            if (err instanceof AuthError && (err.originalError as Response)?.status === 404) {
                console.warn("Skills endpoint not found, returning empty skills");
                // We could return empty skills here if desired, re-throwing for now
            }
            throw err;
        }
    }
}

// Singleton instance
export const gameAPI = new GameAPI();
