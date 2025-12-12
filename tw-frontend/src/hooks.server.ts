import type { Handle } from '@sveltejs/kit';

// Backend API URL - in Docker, services communicate via container names
const BACKEND_URL = process.env.BACKEND_URL || 'http://game-server:8080';

export const handle: Handle = async ({ event, resolve }) => {
    // Proxy /api requests to the backend
    if (event.url.pathname.startsWith('/api')) {
        const backendPath = event.url.pathname;
        const backendUrl = `${BACKEND_URL}${backendPath}${event.url.search}`;

        try {
            const response = await fetch(backendUrl, {
                method: event.request.method,
                headers: event.request.headers,
                body: event.request.method !== 'GET' && event.request.method !== 'HEAD'
                    ? await event.request.text()
                    : undefined,
                // @ts-ignore - duplex is needed for streaming but not in types
                duplex: 'half',
            });

            // Clone headers to avoid immutable header issues
            const headers = new Headers(response.headers);

            return new Response(response.body, {
                status: response.status,
                statusText: response.statusText,
                headers,
            });
        } catch (error) {
            console.error('API proxy error:', error);
            return new Response(JSON.stringify({ error: 'Unable to reach backend server' }), {
                status: 503,
                headers: { 'Content-Type': 'application/json' },
            });
        }
    }

    return resolve(event);
};
