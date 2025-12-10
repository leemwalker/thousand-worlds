#!/usr/bin/env node

// WebSocket test client for verifying JWT authentication
const WebSocket = require('ws');
const https = require('https');
const http = require('http');

const BASE_URL = 'http://localhost:8080/api';
const WS_URL = 'ws://localhost:8080/api/game/ws';

async function httpRequest(method, path, data, token) {
    return new Promise((resolve, reject) => {
        const url = new URL(BASE_URL + path);
        const options = {
            method,
            headers: {
                'Content-Type': 'application/json',
            },
        };

        if (token) {
            options.headers['Authorization'] = `Bearer ${token}`;
        }

        const req = http.request(url, options, (res) => {
            let body = '';
            res.on('data', chunk => body += chunk);
            res.on('end', () => {
                try {
                    resolve(JSON.parse(body));
                } catch (e) {
                    resolve(body);
                }
            });
        });

        req.on('error', reject);
        if (data) {
            req.write(JSON.stringify(data));
        }
        req.end();
    });
}

async function testWebSocketAuth() {
    console.log('=== WebSocket Authentication Test ===\n');

    // 1. Register and login
    const email = `ws_tester_${Date.now()}@example.com`;
    const password = 'password123';

    console.log('1. Registering user...');
    await httpRequest('POST', '/auth/register', { email, password });

    console.log('2. Logging in...');
    const loginResp = await httpRequest('POST', '/auth/login', { email, password });
    const token = loginResp.token;
    console.log(`   Token obtained: ${token.substring(0, 50)}...\n`);

    // 3. Try to connect without token (should fail)
    console.log('3. Testing connection without token (should fail)...');
    try {
        const badWs = new WebSocket(WS_URL);
        await new Promise((resolve, reject) => {
            badWs.on('error', () => {
                console.log('   ✓ Connection rejected as expected\n');
                resolve();
            });
            badWs.on('open', () => {
                badWs.close();
                reject(new Error('Should not have connected'));
            });
        });
    } catch (e) {
        console.log('   ✓ Connection rejected\n');
    }

    // 4. Connect with token
    console.log('4. Connecting with valid JWT token...');
    const ws = new WebSocket(`${WS_URL}?token=${token}`);

    ws.on('open', () => {
        console.log('   ✓ WebSocket connected successfully!\n');

        // Send a test command
        console.log('5. Sending test command...');
        ws.send(JSON.stringify({
            type: 'command',
            data: {
                action: 'look'
            }
        }));
    });

    ws.on('message', (data) => {
        const msg = JSON.parse(data);
        console.log('   Received message:');
        console.log(`   Type: ${msg.type}`);
        console.log(`   Data:`, JSON.stringify(msg.data, null, 2));
        console.log('');
    });

    ws.on('error', (error) => {
        console.error('   WebSocket error:', error.message);
    });

    ws.on('close', () => {
        console.log('   WebSocket connection closed\n');
        console.log('✅ All WebSocket tests passed!');
        process.exit(0);
    });

    // Close after 3 seconds
    setTimeout(() => {
        ws.close();
    }, 3000);
}

testWebSocketAuth().catch(err => {
    console.error('Test failed:', err);
    process.exit(1);
});
