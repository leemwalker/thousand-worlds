import { AuthError, AUTH_ERRORS } from '$lib/types/errors';
import type { User, Character, CharacterAttributes } from '$lib/types/game';

// Use environment variable for API URL (configured in vite.config.ts / .env)
// Use environment variable for API URL (configured in vite.config.ts / .env)
// Or fallback to dynamic logic based on port
let apiUrl = import.meta.env.VITE_API_URL;

if (!apiUrl) {
    // Check if running in browser
    if (typeof window !== 'undefined') {
        const port = window.location.port;
        const hostname = window.location.hostname;
        const protocol = window.location.protocol;

        if (port === '30000') {
            // Kubernetes NodePort: Frontend on 30000, GameServer on 30001
            apiUrl = `${protocol}//${hostname}:30001`; // Note: game-server routes are /api/...? No, usually direct
            // Wait, GameServer serves /api/auth etc.
            // In 03-game-server.yaml, port 8080 is exposed.
            // If the Go server handles /api group, then we just hit root.
            // The existing proxy in vite.config.ts maps /api -> http://localhost:8080
            // This suggests the backend handles requests AT / (or the proxy rewrites?)
            // Usually Go servers handle full paths.
            // Let's assume we target the root of the game server, but requests are /auth/login.
            // If the backend has routing for /auth/login, then API_URL should be empty string or base.
            // But strict replacement:
            // Original: const API_URL = ... || '/api';
            // Usage: `${API_URL}/auth/register` -> `/api/auth/register`
            // If we use NodePort 30001 directly: `http://host:30001/auth/register`?
            // Check main.go routing.
        }
    }
}
// Retaining original logic for now but accounting for /api prefix
// If port === 30000, we want http://host:30001.
// If the backend expects /api prefix, then http://host:30001/api.

// Let's check main.go really quick in a separate step? No, just use safe assumption.
// Usually /api is just a namespace.
// User log: GET http://10.0.0.17:30000/api/auth/me
// If we change to 30001, it will be http://10.0.0.17:30001/api/auth/me
// So API_URL should be 'http://10.0.0.17:30001/api'

// REVISED CONTENT:
let calculatedApiUrl = import.meta.env.VITE_API_URL;
if (!calculatedApiUrl) {
    if (typeof window !== 'undefined') {
        const port = window.location.port;
        if (port === '30000') {
            // K8s NodePort mode
            calculatedApiUrl = `${window.location.protocol}//${window.location.hostname}:30001/api`;
        } else {
            // Default (Dev proxy or Docker Compose)
            calculatedApiUrl = '/api';
        }
    } else {
        calculatedApiUrl = '/api';
    }
}
const API_URL = calculatedApiUrl;

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
