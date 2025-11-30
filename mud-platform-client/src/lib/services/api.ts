const API_URL = '/api';

export interface User {
    user_id: string;
    email: string;
    created_at: string;
    last_login?: string;
}

export interface LoginResponse {
    token: string;
    user: User;
}

export class GameAPI {
    private token: string | null = null;

    setToken(token: string): void {
        this.token = token.trim();
        localStorage.setItem('auth_token', this.token);
    }

    getToken(): string | null {
        if (!this.token) {
            this.token = localStorage.getItem('auth_token');
        }
        return this.token;
    }

    clearToken(): void {
        this.token = null;
        localStorage.removeItem('auth_token');
    }

    async register(email: string, password: string): Promise<User> {
        const response = await fetch(`${API_URL}/auth/register`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email, password }),
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Registration failed');
        }

        return response.json();
    }

    async login(email: string, password: string): Promise<LoginResponse> {
        const response = await fetch(`${API_URL}/auth/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email, password }),
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Login failed');
        }

        const data: LoginResponse = await response.json();
        this.setToken(data.token);
        return data;
    }

    async getMe(): Promise<{ user_id: string }> {
        const token = this.getToken();
        if (!token) {
            throw new Error('Not authenticated');
        }

        const response = await fetch(`${API_URL}/auth/me`, {
            headers: {
                'Authorization': `Bearer ${token.trim()}`,
            },
        });

        if (!response.ok) {
            throw new Error('Failed to get user info');
        }

        return response.json();
    }

    logout(): void {
        this.clearToken();
    }
}

// Singleton instance
export const gameAPI = new GameAPI();
