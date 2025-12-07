export interface UserFriendlyError {
    title: string;
    message: string;
    action?: {
        label: string;
        onClick: () => void;
    };
}

export class AuthError extends Error {
    constructor(
        message: string,
        public userFriendly: UserFriendlyError,
        public originalError?: unknown
    ) {
        super(message);
        this.name = 'AuthError';
    }
}

// Error message catalog
export const AUTH_ERRORS = {
    INVALID_CREDENTIALS: {
        title: 'Login Failed',
        message: 'Incorrect email or password. Please try again.',
    },

    NETWORK_ERROR: {
        title: 'Connection Problem',
        message: 'Unable to reach the server. Please check your internet connection and try again.',
    },

    SERVER_ERROR: {
        title: 'Server Error',
        message: 'The server is experiencing issues. Please try again in a few moments.',
    },

    EMAIL_EXISTS: {
        title: 'Account Already Exists',
        message: 'An account with this email already exists.',
    },

    WEAK_PASSWORD: {
        title: 'Weak Password',
        message: 'Password must be at least 8 characters.',
    },

    INVALID_EMAIL: {
        title: 'Invalid Email',
        message: 'Please enter a valid email address.',
    },

    TIMEOUT: {
        title: 'Request Timeout',
        message: 'The request took too long. Please try again.',
    },
} as const;
