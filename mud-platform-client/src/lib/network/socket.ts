import { connect, type NatsConnection } from 'nats.ws';

let nc: NatsConnection | undefined;

export async function connectToBackend(): Promise<NatsConnection> {
    if (nc) return nc;

    // Connect to NATS via WebSocket
    // Assuming NATS is exposed on port 9222 for WebSockets
    // In development, this might need to be localhost:9222
    // In production, it might be behind a reverse proxy
    const serverUrl = 'ws://localhost:9222';

    try {
        console.log(`Connecting to NATS at ${serverUrl}...`);
        nc = await connect({ servers: [serverUrl] });
        console.log('Connected to NATS');

        // Handle disconnect
        nc.closed().then((err) => {
            console.log('NATS connection closed', err);
            nc = undefined;
        });

        return nc;
    } catch (err) {
        console.error('Failed to connect to NATS', err);
        throw err;
    }
}

export function getConnection(): NatsConnection | undefined {
    return nc;
}
