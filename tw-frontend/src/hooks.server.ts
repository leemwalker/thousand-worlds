import type { Handle } from '@sveltejs/kit';

// Backend API URL - in Docker, services communicate via container names
const BACKEND_URL = process.env.BACKEND_URL || 'http://game-server:8080';

// Headers that should not be forwarded (hop-by-hop headers)
const EXCLUDED_HEADERS = new Set([
    'connection',
    'keep-alive',
    'proxy-authenticate',
    'proxy-authorization',
    'te',
    'trailers',
    'transfer-encoding',
    'upgrade',
    'host',
]);

export const handle: Handle = async ({ event, resolve }) => {
    // Proxy /api requests to the backend
    if (event.url.pathname.startsWith('/api')) {
        const backendPath = event.url.pathname;
        const backendUrl = `${BACKEND_URL}${backendPath}${event.url.search}`;

        try {
            // Filter headers - remove hop-by-hop headers
            const headers = new Headers();
            for (const [key, value] of event.request.headers.entries()) {
                if (!EXCLUDED_HEADERS.has(key.toLowerCase())) {
                    headers.set(key, value);
                }
            }

            const response = await fetch(backendUrl, {
                method: event.request.method,
                headers,
                body: event.request.method !== 'GET' && event.request.method !== 'HEAD'
                    ? await event.request.text()
                    : undefined,
            });

            // Clone response headers, filtering out hop-by-hop headers
            const responseHeaders = new Headers();
            for (const [key, value] of response.headers.entries()) {
                if (!EXCLUDED_HEADERS.has(key.toLowerCase())) {
                    responseHeaders.set(key, value);
                }
            }

            return new Response(response.body, {
                status: response.status,
                statusText: response.statusText,
                headers: responseHeaders,
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
